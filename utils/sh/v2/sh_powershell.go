package sh

import (
	"fmt"
	"strings"
)

func isPowershellVarName(r rune) bool {
	return r == ':' || isVarName(r)
}
func powershellObject(r rune) bool {
	return r == '.' || isVarName(r)
}

var _ Sh = &Powershell{}

const (
	_                 PsType            = iota
	PsAccess                            // Variable is Access
	PsSet                               // Variable is Set
	PsBreakLine                         // Break line
	PsAcessWithObject = PsAccess | iota // Variable is access with Object before dot.
	PsSetArray        = PsSet | iota    // Value type is array
)

type PsType int

type PsTypeValue struct {
	Type       PsType // Value type
	Start, End int    // Content start and end
	Name       string // Key name
	Value      string // Set value
	Content    string // Full content type
}

// Powershell variables is equal or same to Bash script variables  with min to comptaible
//
// Set variables in Powershell:
//   - $[A-Za-z0-9_]=<Value>
//   - $[A-Za-z0-9_]="<Value>"
//   - $[A-Za-z0-9_]='<Value>'
//
// Access variables in Powershell:
//   - $[A-Za-z0-9_]
//   - $[A-Za-z0-9_](.<Object Name[A-Za-z0-9_]>)?
type Powershell struct {
	psContent string         // Powershell script
	Variables []*PsTypeValue // Variables set and access
}

func (pws Powershell) Content(contentInput string) string {
	return contentInput
}

// Set new var to struct
func (pws *Powershell) Set(name, value string) {
	pws.Variables = append(pws.Variables, &PsTypeValue{
		Type:  PsSet,
		Start: -1, // Not contains index
		End:   -1, // Not contains index
		Name:  name,
		Value: value,
	})
}

// Get full var
func (pws Powershell) GetRaw(name string) (string, bool) {
	for _, varInfo := range pws.Variables {
		if varInfo.Type != PsAccess && varInfo.Type&PsAccess > 0 && varInfo.Name == name {
			return pws.psContent[varInfo.Start:varInfo.End], true
		}
	}
	return "", false
}

// Get var value
func (pws Powershell) Get(name string) (string, bool) {
	for _, varInfo := range pws.Variables {
		if varInfo.Type != PsAccess && varInfo.Type&PsAccess > 0 && varInfo.Name == name {
			return varInfo.Value, true
		}
	}
	return "", false
}

// Parse powershell script if valid return new [*Powershell]
func PowershellScript(content string) (Sh, error) {
	ps1 := &Powershell{psContent: content, Variables: []*PsTypeValue{}}

	for contentIndex := 0; contentIndex < len(content); contentIndex++ {
		switch content[contentIndex] {
		case '\'':
			contentIndex += strings.IndexRune(content[contentIndex+1:], '\'')+1
		case '"':
			contentIndex += strings.IndexRune(content[contentIndex+1:], '"')+1
		case '$':
			startVar := contentIndex
			for contentIndex += 1; contentIndex < len(content)-1 && isPowershellVarName(rune(content[contentIndex])); contentIndex++ {
			}

			for contentIndex < len(content)-1 && isSpace(rune(content[contentIndex])) {
				contentIndex++
			}

			switch content[contentIndex] {
			case '=':
				fullKeyName := strings.TrimSpace(content[startVar+1 : contentIndex])
				for contentIndex += 1; contentIndex < len(content)-1 && isSpace(rune(content[contentIndex])); contentIndex++ {
				}
				valueStart := contentIndex

				psTypeSet := PsSet
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
					psTypeSet = PsSetArray
					endIndex := strings.IndexRune(content[contentIndex:], ')')
					if endIndex == -1 {
						return nil, fmt.Errorf("cannot get final of string set")
					}
					contentIndex += endIndex
				default:
					endIndex := strings.IndexFunc(content[contentIndex:], isSpace)
					if endIndex == -1 {
						return nil, fmt.Errorf("cannot get final of string set")
					}
					contentIndex += endIndex
				}

				ps1.Variables = append(ps1.Variables, &PsTypeValue{
					Type:    psTypeSet,
					Start:   startVar,
					End:     contentIndex,
					Name:    fullKeyName,
					Value:   strings.Trim(content[valueStart:contentIndex], "@()\"`'"),
					Content: content[startVar : contentIndex+1],
				})
				contentIndex++
			case '.':
				endIndex := strings.IndexFunc(content[contentIndex:], powershellObject)
				if endIndex == -1 {
					return nil, fmt.Errorf("cannot get final of string set")
				}
				contentIndex += endIndex
				panic(fmt.Sprintf("line: %q", content[startVar:contentIndex]))
			default:
				// Var access
				ps1.Variables = append(ps1.Variables, &PsTypeValue{
					Type:    PsAccess,
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
