package sh

import (
	"fmt"
	"iter"
	"strings"
	"unicode"
)

func isPowershellVarName(r rune) bool {
	return r == ':' || isVarName(r)
}

var (
	_ Sh      = &Powershell{}
	_ ShValue = &PsValue{}
)

type PsValue struct {
	Type       RawType // Value type
	Start, End int     // Content start and end
	Name       string  // Key name
	Value      string  // Set value
	Content    string  // Full content type
}

func (psValue PsValue) ValueType() RawType   { return psValue.Type }
func (psValue PsValue) String() string       { return psValue.Value }
func (psValue PsValue) KeyName() string      { return psValue.Name }
func (psValue PsValue) ContentIndex() [2]int { return [2]int{psValue.Start, psValue.End} }
func (psValue PsValue) ContentArray() iter.Seq[string] {
	if psValue.Type == VarSetArray {
		// @(<Filed1>, <Field2>)
		// @(<Filed1> <Field2>)
		return func(yield func(string) bool) {
			copyContent := psValue.Value[:]
			for copyContent != "" {
				switch copyContent[0] {
				case '"':
					endKey := strings.IndexRune(copyContent[1:], '"')
					if endKey == -1 {
						return
					}
					endKey++
					if !yield(copyContent[1:endKey]) {
						return
					}
					copyContent = copyContent[endKey+1:]
				case '\'':
					endKey := strings.IndexRune(copyContent[1:], '\'')
					if endKey == -1 {
						return
					}
					endKey++
					if !yield(copyContent[1:endKey]) {
						return
					}
					copyContent = copyContent[endKey+1:]
				default:
					if unicode.IsSpace(rune(copyContent[0])) || copyContent[0] == ',' {
						copyContent = copyContent[1:]
						continue
					}
					endKey := strings.IndexFunc(copyContent, unicode.IsSpace)
					if endKey == -1 {
						return
					}
					endKey++
					if !yield(copyContent[:endKey]) {
						return
					}
					copyContent = copyContent[endKey+1:]
				}
			}
		}
	}
	return func(yield func(string) bool) { yield(psValue.Value) }
}

// Powershell variables is equal or same to Bash script variables  with min to comptaible
//
// Set variables in Powershell:
//   - $[A-Za-z0-9_]=<Value>
//   - $[A-Za-z0-9_]=@(<Value 1>, <Value 2>)
//   - $[A-Za-z0-9_]=@(<Value 1> <Value 2>)
//   - $[A-Za-z0-9_]="<Value>"
//   - $[A-Za-z0-9_]='<Value>'
//
// Access variables in Powershell:
//   - $[A-Za-z0-9_]
//   - $[A-Za-z0-9_](.<Object Name[A-Za-z0-9_]>)?
type Powershell struct {
	psContent string     // Powershell script
	Variables []*PsValue // Variables set and access
}

func (pws Powershell) Raw() []ShValue {
	var n []ShValue
	for _, n2 := range pws.Variables {
		n = append(n, n2)
	}
	return n
}

// Set new var to struct
func (pws *Powershell) Set(name, value string) {
	pws.Variables = append(pws.Variables, &PsValue{
		Type:  VarSet,
		Start: -1, // Not contains index
		End:   -1, // Not contains index
		Name:  name,
		Value: value,
	})
}

// Get full var
func (pws Powershell) GetRaw(name string) (ShValue, bool) {
	for _, varInfo := range pws.Variables {
		if varInfo.Type != VarAccess && varInfo.Type&VarAccess > 0 && varInfo.Name == name {
			return varInfo, true
		}
	}
	return nil, false
}

// Get var value
func (pws Powershell) Get(name string) (string, bool) {
	for _, varInfo := range pws.Variables {
		if varInfo.Type != VarAccess && varInfo.Type&VarAccess > 0 && varInfo.Name == name {
			return varInfo.Value, true
		}
	}
	return "", false
}

// Parse powershell script if valid return new [*Powershell]
func PowershellScript(content string) (Sh, error) {
	ps1 := &Powershell{psContent: content, Variables: []*PsValue{}}

	for contentIndex := 0; contentIndex < len(content); contentIndex++ {
		switch content[contentIndex] {
		case '\'':
			contentIndex += strings.IndexRune(content[contentIndex+1:], '\'') + 1
		case '"':
			contentIndex += strings.IndexRune(content[contentIndex+1:], '"') + 1
		case '$':
			startVar := contentIndex
			for contentIndex += 1; contentIndex < len(content)-1 && isPowershellVarName(rune(content[contentIndex])); contentIndex++ {
			}

			for contentIndex < len(content)-1 && isSpace(rune(content[contentIndex])) {
				contentIndex++
			}

			switch content[contentIndex] {
			case '(':
			case '=':
				fullKeyName := strings.TrimSpace(content[startVar+1 : contentIndex])
				for contentIndex += 1; contentIndex < len(content)-1 && isSpace(rune(content[contentIndex])); contentIndex++ {
				}
				valueStart := contentIndex

				psTypeSet := VarSet
				switch content[contentIndex] {
				case '\'':
					contentIndex++
					endIndex := strings.IndexRune(content[contentIndex:], '\'')
					if endIndex == -1 {
						return nil, fmt.Errorf("cannot get final of string set")
					}
					contentIndex += endIndex
				case '"':
					contentIndex++
					endIndex := strings.IndexRune(content[contentIndex:], '"')
					if endIndex == -1 {
						return nil, fmt.Errorf("cannot get final of string set")
					}

					if content[contentIndex+endIndex-1] == '`' {
						endIndex += strings.IndexRune(content[contentIndex+endIndex+1:], '`')
						endIndex += strings.IndexRune(content[contentIndex+endIndex+1:], '"') + 1
					}
					contentIndex += endIndex
				case '@':
					contentIndex += 2
					endIndex := strings.IndexRune(content[contentIndex:], ')')
					if endIndex == -1 {
						return nil, fmt.Errorf("cannot get final of string set")
					}
					contentIndex += endIndex

					ps1.Variables = append(ps1.Variables, &PsValue{
						Type:    VarSetArray,
						Start:   startVar,
						End:     contentIndex,
						Name:    fullKeyName,
						Value:   strings.Trim(content[valueStart:contentIndex], "@()"),
						Content: content[startVar : contentIndex+1],
					})

					continue
				default:
					endIndex := strings.IndexFunc(content[contentIndex:], isSpace)
					if endIndex == -1 {
						return nil, fmt.Errorf("cannot get final of string set")
					}
					contentIndex += endIndex
				}

				ps1.Variables = append(ps1.Variables, &PsValue{
					Type:    psTypeSet,
					Start:   startVar,
					End:     contentIndex,
					Name:    fullKeyName,
					Value:   strings.Trim(content[valueStart:contentIndex], "@()\"`'"),
					Content: content[startVar : contentIndex+1],
				})
				contentIndex++
			case '.':
				endIndex := strings.IndexFunc(content[contentIndex:], func(r rune) bool { return !(r == '.' || isVarName(r)) })
				if endIndex == -1 {
					return nil, fmt.Errorf("cannot get final of string set")
				}
				contentIndex += endIndex
				if content[contentIndex] == '(' {
					endIndex := strings.IndexRune(content[contentIndex+1:], ')')
					if endIndex == -1 {
						return nil, fmt.Errorf("cannot get final of string set")
					}
					contentIndex += endIndex + 1
				}
				// Var access with object
				ps1.Variables = append(ps1.Variables, &PsValue{
					Type:    VarAcessWithObject,
					Start:   startVar,
					End:     contentIndex,
					Name:    strings.TrimSpace(content[startVar+1 : contentIndex]),
					Content: content[startVar:contentIndex],
				})
			default:
				// Var access
				ps1.Variables = append(ps1.Variables, &PsValue{
					Type:    VarAccess,
					Start:   startVar,
					End:     contentIndex,
					Name:    strings.TrimSpace(content[startVar+1 : contentIndex]),
					Content: content[startVar:contentIndex],
				})
			}
		}
	}
	return ps1, nil
}
