// Extract variasbles from Shell scripts to Windows CMD, Powershell, Bash, Zsh a Sh shell
package sh

import (
	"fmt"
	"iter"
	"unicode"
)

const (
	_                  RawType            = iota
	VarAccess                             // Variable is Access
	VarSet                                // Variable is Set
	VarBreakLine                          // Break line
	VarAcessWithObject = VarAccess | iota // Variable is access with Object before dot.
	VarSetArray        = VarSet | iota    // Value type is array
)

var rawString = []string{
	VarAccess:          "access",
	VarSet:             "set",
	VarBreakLine:       "break line",
	VarAcessWithObject: "access with object",
	VarSetArray:        "set array values",
}

// generic enum to ShValue Type
type RawType int

type ShValue interface {
	fmt.Stringer                    // Return value from content
	ValueType() RawType             // Return value type
	KeyName() string                // Value key name
	ContentIndex() [2]int           // return index of content value, [start, end]
	ContentArray() iter.Seq[string] // return array split value
}

// Generic interface to processed file script's
type Sh interface {
	Set(name, value string)             // Set variable to struct
	GetRaw(name string) (ShValue, bool) // Get var without processed line
	Get(name string) (string, bool)     // Get var with process variables from insider value
	Raw() []ShValue                     // Get all values types
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

// Process contentInput to replace variables accesses, if contains set value ignore
func Content(sh Sh, contentInput string) string {
	return contentInput
}

func ContainsAccessVar(input Sh, varName string) bool {
	for _, info := range input.Raw() {
		if info.ValueType() == VarAccess || info.ValueType()&VarAccess > 0 {
			if info.KeyName() == varName {
				return true
			}
		}
	}
	return false
}

func Must(n Sh, _ error) Sh {
	return n
}

func Seq(s Sh) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, value := range s.Raw() {
			if !yield(value.KeyName(), value.String()) {
				return
			}
		}
	}
}
