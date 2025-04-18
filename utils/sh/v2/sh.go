// Extract variasbles from Shell scripts to Windows CMD, Powershell, Bash, Zsh a Sh shell
package sh

import "unicode"

// Generic interface to processed file script's
type Sh interface {
	Content(contentInput string) string // Process contentInput to replace variables accesses, if contains set value ignore
	Set(name, value string)             // Set variable to struct
	GetRaw(name string) (string, bool)  // Get var without processed line
	Get(name string) (string, bool)     // Get var with process variables from insider value
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
