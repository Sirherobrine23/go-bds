package sh

import (
	"iter"
	"strings"
	"unicode"
)

var (
	_ ShValue = &CmdValue{}
	_ Sh      = &CommandPrompt{}
)

type CmdValue struct {
	Type       RawType // Value type
	Start, End int     // Content start and end
	Name       string  // Key name
	Value      string  // Set value
	Content    string  // Full content type
}

func (value CmdValue) ValueType() RawType   { return value.Type }
func (value CmdValue) String() string       { return value.Value }
func (value CmdValue) KeyName() string      { return value.Name }
func (value CmdValue) ContentIndex() [2]int { return [2]int{value.Start, value.End} }
func (value CmdValue) ContentArray() iter.Seq[string] {
	return func(yield func(string) bool) { yield(value.Value) }
}

// Set variables in bat:
//   - set [A-Za-z0-9_]=<Value>
//   - set [A-Za-z0-9_]="<Value>"
//   - set [A-Za-z0-9_]='<Value>'
//
// Access variables in bat:
//   - %[A-Za-z0-9_]%
type CommandPrompt struct {
	cmdContent string      // Windows bat script
	Variables  []*CmdValue // Variables set and access
}

func (cmd CommandPrompt) Raw() []ShValue {
	var n []ShValue
	for _, n2 := range cmd.Variables {
		n = append(n, n2)
	}
	return n
}

// Set new var to struct
func (cmd *CommandPrompt) Set(name, value string) {
	cmd.Variables = append(cmd.Variables, &CmdValue{
		Type:  VarSet,
		Start: -1, // Not contains index
		End:   -1, // Not contains index
		Name:  name,
		Value: value,
	})
}

// Get full var
func (cmd CommandPrompt) GetRaw(name string) (ShValue, bool) {
	for _, varInfo := range cmd.Variables {
		if varInfo.Type != VarAccess && varInfo.Type&VarAccess > 0 && varInfo.Name == name {
			return varInfo, true
		}
	}
	return nil, false
}

// Get var value
func (cmd CommandPrompt) Get(name string) (string, bool) {
	for _, varInfo := range cmd.Variables {
		if varInfo.Type != VarAccess && varInfo.Type&VarAccess > 0 && varInfo.Name == name {
			return varInfo.Value, true
		}
	}
	return "", false
}

// Parse command prompt/bat script if valid return new [*CommandPrompt]
func CommandPromptScript(content string) (Sh, error) {
	cmd := &CommandPrompt{cmdContent: content, Variables: []*CmdValue{}}
	for contentIndex := 0; contentIndex < len(content); contentIndex++ {
		switch content[contentIndex] {
		case '%':
			startVar := contentIndex
			for contentIndex += 1; contentIndex < len(content)-1 && isVarName(rune(content[contentIndex])); contentIndex++ {
			}
			if content[contentIndex] != '%' {
				contentIndex = startVar + 1
				continue
			}
			keyName := content[startVar+1 : contentIndex]
			if keyName == "" {
				contentIndex = startVar + 1
				continue
			}
			// println(keyName)
			cmd.Variables = append(cmd.Variables, &CmdValue{
				Type:    VarAccess,
				Name:    keyName,
				Start:   startVar,
				End:     contentIndex + 1,
				Content: content[startVar : contentIndex+1],
			})
		case 's':
			setStart, fistStart := contentIndex, contentIndex
			indexSet := strings.IndexFunc(content[contentIndex:], unicode.IsSpace)
			if indexSet == -1 {
				continue
			}
			contentIndex += indexSet
			if content[setStart:contentIndex] != "set" {
				continue
			}
			setStart = contentIndex + 1
			setKeyIndex := strings.IndexRune(content[contentIndex:], '=')
			if setKeyIndex == -1 {
				continue
			}
			contentIndex += setKeyIndex
			setKeyName := content[setStart:contentIndex]
			contentIndex++
			buffStart := contentIndex
			// get value
			switch content[contentIndex] {
			case '"':
				endLine := strings.IndexRune(content[contentIndex+1:], '"')
				if endLine == -1 {
					continue
				}
				contentIndex += endLine + 1

				cmd.Variables = append(cmd.Variables, &CmdValue{
					Type:    VarSet,
					Name:    setKeyName,
					Start:   fistStart,
					End:     contentIndex,
					Value:   content[buffStart+1 : contentIndex],
					Content: content[fistStart : contentIndex+1],
				})
			default:
				endLine := strings.IndexRune(content[contentIndex:], '\n')
				if endLine == -1 {
					if endLine = strings.IndexFunc(content[contentIndex:], unicode.IsSpace); endLine == -1 {
						endLine = len(content) - 1
					}
				}
				contentIndex += endLine
				valueContent := content[buffStart:contentIndex]
				if commentIndex := strings.IndexRune(valueContent, '#'); commentIndex > 0 {
					valueContent = valueContent[:commentIndex]
				}

				cmd.Variables = append(cmd.Variables, &CmdValue{
					Type:    VarSet,
					Name:    setKeyName,
					Start:   fistStart,
					End:     contentIndex,
					Content: content[fistStart:contentIndex],
					Value:   valueContent,
				})
			}
		}
	}
	return cmd, nil
}
