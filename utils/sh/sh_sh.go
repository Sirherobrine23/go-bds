package sh

import (
	"encoding/json"
	"fmt"
	"iter"
	"strings"
	"unicode"
)

type Bash struct {
	bashContent string       // Bash script
	Variables   []*BashValue // Variables set and access
}

func (pws Bash) Raw() []ShValue {
	var n []ShValue
	for _, n2 := range pws.Variables {
		n = append(n, n2)
	}
	return n
}

// Set new var to struct
func (pws *Bash) Set(name, value string) {
	pws.Variables = append(pws.Variables, &BashValue{
		Type:  VarSet,
		Start: -1, // Not contains index
		End:   -1, // Not contains index
		Name:  name,
		Value: value,
	})
}

// Get full var
func (pws Bash) GetRaw(name string) (ShValue, bool) {
	for _, varInfo := range pws.Variables {
		if varInfo.Type != VarAccess && varInfo.Type&VarAccess > 0 && varInfo.Name == name {
			return varInfo, true
		}
	}
	return nil, false
}

// Get var value
func (pws Bash) Get(name string) (string, bool) {
	for _, varInfo := range pws.Variables {
		if varInfo.Type != VarAccess && varInfo.Type&VarAccess > 0 && varInfo.Name == name {
			return varInfo.Value, true
		}
	}
	return "", false
}

func _testPrint(v any) {
	d, _ := json.MarshalIndent(v, "", "  ")
	println(string(d))
}

// Process Bash, Zsh and Sh Script
func BashScript(content string) (Sh, error) {
	bash := &Bash{bashContent: content, Variables: []*BashValue{}}
	for contentIndex := 0; contentIndex < len(content); contentIndex++ {
		switch content[contentIndex] {
		case '$':
			startAccess := contentIndex
			contentIndex++
			switch content[contentIndex] {
			case '(':
				continue
			case '{':
				endKey := strings.IndexRune(content[contentIndex:], '}')
				if endKey == -1 {
					return nil, fmt.Errorf("invalid key open and close")
				}
				contentIndex += endKey + 1
				processValue := content[startAccess+2 : contentIndex-1]
				switch strings.Count(processValue, ":") {
				case 1:
					splitDefault := strings.SplitN(processValue, ":", 2)
					_testPrint(splitDefault)
				case 2:
					splitSlice := strings.SplitN(processValue, ":", 3)
					_testPrint(splitSlice)
				default:
					valueAccess := strings.TrimSuffix(strings.Trim(processValue, "${}"), "[@]")
					bash.Variables = append(bash.Variables, &BashValue{
						Type:    VarAccess,
						Name:    valueAccess,
						Start:   startAccess,
						End:     contentIndex,
						Content: content[startAccess:contentIndex],
					})
				}
			default:
				endKey := strings.IndexFunc(content[contentIndex:], isInvalidVarName)
				if endKey == -1 {
					return nil, fmt.Errorf("invalid key open and close")
				}
				contentIndex += endKey
				valueAccess := strings.TrimSuffix(content[startAccess+1:contentIndex], "[@]")
				if valueAccess == "" {
					continue
				}
				bash.Variables = append(bash.Variables, &BashValue{
					Type:    VarAccess,
					Name:    valueAccess,
					Start:   startAccess,
					End:     contentIndex,
					Content: content[startAccess:contentIndex],
				})
			}
		case '=':
			valueSet := contentIndex
			for contentIndex--; contentIndex < len(content)-1 && isVarName(rune(content[contentIndex])); contentIndex-- {
			}
			keyName := content[contentIndex:valueSet]
			startKey := contentIndex
			contentIndex = valueSet + 1
			for ; contentIndex < len(content)-1 && unicode.IsSpace(rune(content[contentIndex])); contentIndex++ {
			}

			_ = startKey
			_ = valueSet
			_ = keyName
			switch content[contentIndex] {
			case '"':
				panic(content[contentIndex:])
			case '\'':
				panic(content[contentIndex:])
			case '(':
				panic(content[contentIndex:])
			}
		}
	}
	return bash, nil
}

type BashValue struct {
	Type               RawType // Value type
	Name, Value        string  // Name and Value string
	Start, End         int     // Content location Start and End
	SeekStart, SeekEnd int     // Object slice
	Content            string  // Original content
}

func (value BashValue) ValueType() RawType   { return value.Type }
func (value BashValue) String() string       { return value.Value }
func (value BashValue) KeyName() string      { return value.Name }
func (value BashValue) ContentIndex() [2]int { return [2]int{value.Start, value.End} }
func (value BashValue) ContentArray() iter.Seq[string] {
	if value.Type != VarSetArray {
		return func(yield func(string) bool) { yield(value.Value) }
	}

	// @(<Filed1>, <Field2>)
	// @(<Filed1> <Field2>)
	return func(yield func(string) bool) {
		copyContent := value.Value[:]
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
