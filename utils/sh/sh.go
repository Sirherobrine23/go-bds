// Extract variasbles from Shell scripts to Windows CMD, Powershell, Bash, Zsh a Sh shell
package sh

import (
	"iter"
	"slices"
	"strings"
	"unicode"
)

const (
	_               RawType         = iota
	Access                          // Variable is Access
	Set                             // Variable is Set
	BreakLine                       // Break line
	SetArray        = Set | iota    // Value type is array
	AcessWithObject = Access | iota // Variable is access with Object before dot.
)

var (
	_ Value = BasicSet{}

	rawString = []string{
		Access:          "access",
		Set:             "set",
		BreakLine:       "break line",
		SetArray:        "set array values",
		AcessWithObject: "access with object",
	}
)

// generic enum to ShValue Type
type RawType int

func (s RawType) IsSet() bool                  { return s == Set || s&Set > 0 }
func (s RawType) IsAccess() bool               { return s == Access || s&Access > 0 }
func (s RawType) MarshalText() ([]byte, error) { return []byte(s.String()), nil }
func (s RawType) String() string {
	if n := rawString[s]; n != "" {
		return n
	}
	return "Unkown"
}

type Value interface {
	ValueType() RawType      // Return value type
	String() string          // Return value from content
	KeyName() string         // Value key name
	Array() iter.Seq[string] // return array split value
}

type Sh iter.Seq2[string, []Value]

// Generic interface to processed file script's
type ProcessSh interface {
	Rollback()                    // Return to SavePoint caller
	SavePoint()                   // Save current point
	Back()                        // Shortcut to Add(-1)
	Add(int)                      // Add or remove line from current line
	SetKey(keyName, value string) // Set new key value
	Seq(...int) Sh                // Process line
}

func isInvalidVarName(r rune) bool { return !isVarName(r) }
func isVarName(r rune) bool {
	return r == '_' || unicode.IsDigit(r) || unicode.IsLetter(r) && (unicode.IsUpper(r) || unicode.IsLower(r))
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t':
		return true
	}
	return false
}

func splitContent(content string, start, end int) (value, startContent, endContent string) {
	value = content[start:end]
	startContent = content[:start]
	endContent = content[end:]
	return
}

func findValue(values []Value, keyName string) (Value, bool) {
	for _, v := range slices.Backward(values) {
		if v.KeyName() != keyName {
			continue
		} else if v.ValueType().IsSet() {
			return v, true
		}
	}
	return nil, false
}

type BasicSet [2]string

func (bs BasicSet) ValueType() RawType { return Set }
func (bs BasicSet) String() string     { return bs[1] }
func (bs BasicSet) KeyName() string    { return bs[0] }
func (bs BasicSet) Array() iter.Seq[string] {
	// @(<Filed1>, <Field2>)
	// @(<Filed1> <Field2>)
	return func(yield func(string) bool) {
		copyContent := bs.String()
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
					endKey = len(copyContent) - 1
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
