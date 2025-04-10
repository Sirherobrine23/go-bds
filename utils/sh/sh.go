// Extract variasbles from Shell scripts to Windows CMD, Powershell, Bash, Zsh a Sh shell
package sh

import (
	"iter"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

func isInvalidVarName(r rune) bool { return !isVarName(r) }
func isVarName(r rune) bool {
	return r == '_' || unicode.IsDigit(r) || unicode.IsLetter(r) && (unicode.IsUpper(r) || unicode.IsLower(r))
}

// Variable info
type Var struct {
	Value   string
	Seek    int
	SeekEnd int
}

func (v Var) String() string {
	// Slice string safety
	if v.Seek >= 0 {
		value, seek, seekEnd := v.Value, v.Seek, v.SeekEnd
		seek = min(0, max(0, seek))
		seekEnd = min(len(value)-1, max(0, seekEnd))
		return value[seek:seekEnd]
	}
	return v.Value
}

// # Access variables:
//
// Windows Cmd:
//   - %[A-Za-z0-9_]%
//
// Windows Powershell:
//   - $[A-Za-z0-9_]
//
// bash, zsh, sh:
//   - $[A-Za-z0-9_]
//   - ${[A-Za-z0-9_]}
//   - ${[A-Za-z0-9_]:-<Value>}
//   - ${[A-Za-z0-9_]:<Seek>:<SeekEnd>}
//
// # Set variables:
//
// Windows Cmd:
//
//   - set [A-Za-z0-9_]=<Value>
//
// Windows Powershell:
//   - $[A-Za-z0-9_]=<Value>
//   - $[A-Za-z0-9_]="<Value>"
//
// bash, zsh, sh:
//   - [A-Za-z0-9_]=<Value>
//   - [A-Za-z0-9_]='<Value>'
//   - [A-Za-z0-9_]="<Value>"
type Sh struct {
	scriptContent string            // Script content
	IsCmd         bool              // Is Cmd script
	IsPowershell  bool              // Is Powershell script
	vars          map[string]*Var   // Variables set
	dValues       map[string]string // Default values if present in VariablesCall
	calls         []*Var            // Variables caller
}

func (sh Sh) Lines() []string            { return slices.Collect(sh.LinesSeq()) }
func (sh Sh) LinesSeq() iter.Seq[string] { return strings.SplitAfterSeq(sh.scriptContent, "\n") }

func (sh Sh) Seq() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for key, keyInfo := range sh.vars {
			if !yield(key, sh.replaceWithVar(keyInfo.Value, key)) {
				return
			}
		}
	}
}

// Clone [*Sh] to new [*Sh]
func (sh Sh) Clone() *Sh {
	return &Sh{
		scriptContent: sh.scriptContent,
		IsCmd:         sh.IsCmd,
		IsPowershell:  sh.IsPowershell,
		vars:          maps.Clone(sh.vars),
		dValues:       maps.Clone(sh.dValues),
		calls:         slices.Clone(sh.calls),
	}
}

func (sh *Sh) SetVar(key, value string) {
	if sh.vars[key] == nil {
		sh.vars[key] = &Var{SeekEnd: -1, Seek: -1}
	}
	sh.vars[key].Value = value
}

// Check if in script have non key name set
func (sh Sh) ContainsVar(keyName string) bool {
	return slices.ContainsFunc(sh.calls, func(v *Var) bool { return v.Value == keyName })
}

func (sh Sh) ReplaceWithVar(line string) string {
	return sh.replaceWithVar(line, "")
}

func (sh Sh) replaceWithVar(line, keyName string) string {
	return processLineCall(sh.IsPowershell, sh.IsCmd, line, func(varName, defaultValue string, seek, seekEnd int) string {
		if varName == keyName {
			return ""
		}

		switch v := sh.vars[varName]; v {
		case nil:
			return defaultValue
		default:
			return (&Var{Value: v.Value, Seek: seek, SeekEnd: seekEnd}).String()
		}
	})
}

// Parse shell script to get Variables
func ProcessSh(content string) *Sh {
	sh := &Sh{
		vars:    map[string]*Var{},
		calls:   []*Var{},
		dValues: map[string]string{},
	}

	// Clean comments
	for line := range strings.Lines(strings.TrimSpace(content)) {
		trimedLine := strings.ToLower(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(trimedLine, "#"):
			continue
		case strings.HasPrefix(trimedLine, "@rem "):
			continue
		case strings.HasPrefix(trimedLine, "rem "):
			continue
		case strings.HasPrefix(trimedLine, ":: "):
			continue
		}
		if line == "" || line == "\n" || line == "\r\n" {
			continue
		}
		line = strings.ReplaceAll(strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(line, "^\n"), "^`\n"), "\\\n"), "\t", "  ")

		// cmd variables
		if !sh.IsCmd {
			start := -1
			for check := range len(line) {
				if start == -1 && line[check] == '%' {
					start = check
					continue
				} else if start != -1 && line[check] == '%' {
					keyName := line[start:check]
					if strings.ContainsFunc(keyName[1:], isInvalidVarName) {
						check++
						continue
					}
					sh.IsCmd = true
					sh.scriptContent += line
					continue
				}
			}
		}

		// Powershell variables
		if !sh.IsPowershell && !sh.IsCmd {
			for linePoint := 0; linePoint < len(line); linePoint++ {
				if line[linePoint] == '$' {
					linePoint++
					for linePoint < len(line)-1 && isVarName(rune(line[linePoint])) {
						linePoint++
					}
					if line[linePoint] != '=' {
						continue
					}
					sh.IsPowershell = true
					sh.scriptContent += line
					continue
				}
			}
		}

		// Unix Shell

		if commentIndex := strings.LastIndex(line, "#"); commentIndex > 0 {
			line = strings.TrimLeftFunc(line[:commentIndex], unicode.IsSpace)
		}
		sh.scriptContent += line
	}
	sh.scriptContent = strings.TrimSpace(sh.scriptContent)
	content = sh.scriptContent

	for line := range strings.SplitSeq(content, "\n") {
		// Variables set var
		processLineSet(sh.IsPowershell, sh.IsCmd, line, func(keyName, value string) {
			if sh.vars[keyName] == nil {
				sh.vars[keyName] = &Var{
					Value:   value,
					Seek:    0,
					SeekEnd: 0,
				}
			}
		})

		// Varaibles caller
		processLineCall(sh.IsPowershell, sh.IsCmd, line, func(varName, defaultValue string, seek, seekEnd int) string {
			if !slices.ContainsFunc(sh.calls, func(v *Var) bool { return v.Value == varName }) {
				sh.calls = append(sh.calls, &Var{
					Value:   varName,
					Seek:    seek,
					SeekEnd: seekEnd,
				})
				if defaultValue != "" {
					sh.dValues[varName] = defaultValue
				}
			}
			return ""
		})
	}

	for key := range sh.vars {
		sh.vars[key].Value = sh.replaceWithVar(sh.vars[key].Value, key)
	}

	return sh
}

func processCallLoop(isPws, isCmd bool, keyName, value string) string {
	lines := [][2]int{}
	processLineCallIndex(isPws, isCmd, value, func(startVarName, endVarName, defaultValueAt, _, _ int, varName, defaultValue string) string {
		if keyName == varName {
			if defaultValueAt != -1 {
				lines = append(lines, [2]int{startVarName, defaultValueAt})
			} else {
				lines = append(lines, [2]int{startVarName, endVarName})
			}
		}
		return ""
	})

	less := 0
	for index := range lines {
		value = value[:lines[index][0]-less] + value[lines[index][1]-less:]
		less += lines[index][1]
	}
	return value
}

func processLineSet(isPws, isCmd bool, line string, fn func(keyName, value string)) {
	if isCmd {
		for {
			if setStart, keySet := strings.Index(strings.ToLower(line), "set "), strings.Index(line, "="); setStart >= 0 && (keySet == -1 || keySet > setStart) {
				if keySet == -1 {
					keySet = len(line) - 1
				}

				keyName := line[setStart+4 : keySet]
				if strings.ContainsFunc(keyName, isInvalidVarName) {
					line = line[keySet:]
					continue
				}

				// Remove set from line
				line = strings.TrimRightFunc(line[keySet+1:], unicode.IsSpace)
				if line == "" {
					fn(keyName, line)
					continue
				}

				endKey := -1
				switch line[0] {
				case '\'':
					line = line[1:]
					if endKey = strings.IndexRune(line[1:], '\''); endKey != -1 {
						endKey++
					}
				case '"':
					line = line[1:]
					if endKey = strings.IndexRune(line[1:], '"'); endKey != -1 {
						endKey++
					}
				}

				// Check if is negative value
				if endKey == -1 {
					endKey = len(line)
				}

				fn(keyName, processCallLoop(isPws, isCmd, keyName, line[:endKey]))
				line = line[endKey:]
				continue
			}
			break
		}
		return
	} else if isPws {
		for linePoint := 0; linePoint < len(line); linePoint++ {
			if line[linePoint] == '$' {
				linePoint++
				varNameStart := linePoint
				for linePoint < len(line)-1 && (line[linePoint] == ' ' || isVarName(rune(line[linePoint]))) {
					linePoint++
				}
				if line[linePoint] != '=' {
					continue
				}
				keyName := strings.TrimSpace(line[varNameStart:linePoint])
				if keyName == "" || strings.ContainsFunc(keyName, isInvalidVarName) {
					continue
				}
				line = strings.TrimLeftFunc(line[linePoint+1:], unicode.IsSpace)
				switch line[0] {
				case '@':
					endArray := strings.IndexRune(line, ')')
					if endArray == -1 {
						endArray = len(line) - 1
					}
					fn(keyName, processCallLoop(isPws, isCmd, keyName, line[2:endArray]))
					line = line[endArray+1:]
				case '\'':
					line = line[1:]
					endArray := strings.IndexRune(line, '\'')
					if endArray == -1 {
						endArray = len(line) - 1
					}
					fn(keyName, processCallLoop(isPws, isCmd, keyName, line[:endArray]))
					line = line[endArray+1:]
				case '"':
					line = line[1:]
					endArray := strings.IndexRune(line, '"')
					if endArray == -1 {
						endArray = len(line) - 1
					} else if endArray > 0 && line[endArray-1] == '`' {
						endArray += strings.IndexRune(line[endArray+1:], '`')
						endArray += strings.IndexRune(line[endArray+1:], '"') + 1
					}
					fn(keyName, processCallLoop(isPws, isCmd, keyName, strings.Trim(line[:endArray], "`\"")))
					line = line[endArray+1:]
				default:
					endString := strings.IndexFunc(line, unicode.IsSpace)
					if endString == -1 {
						endString = len(line) - 1
					}
					fn(keyName, processCallLoop(isPws, isCmd, keyName, line[:endString]))
					line = line[endString+1:]
				}
			}
		}
		return
	}

	for linePoint := 0; linePoint < len(line); linePoint++ {
		if line[linePoint] == '=' {
			varNameStart := linePoint // Invert point to get varName
			linePoint--
			for ; linePoint == 0 || isVarName(rune(line[linePoint])); linePoint-- {
				if linePoint == 0 {
					break
				}
			}
			keyName := strings.TrimSpace(line[linePoint:varNameStart])
			if keyName == "" || strings.ContainsFunc(keyName, isInvalidVarName) {
				linePoint = varNameStart + 1
				continue
			}
			if line = strings.TrimLeftFunc(line[varNameStart+1:], unicode.IsSpace); line == "" {
				fn(keyName, line)
				return
			}
			switch line[0] {
			case '(':
				endArray := strings.IndexRune(line, ')')
				if endArray == -1 {
					endArray = len(line) - 1
				}
				fn(keyName, processCallLoop(isPws, isCmd, keyName, line[1:endArray]))
				line = line[endArray+1:]
			case '\'':
				line = line[1:]
				endArray := strings.IndexRune(line, '\'')
				if endArray == -1 {
					endArray = len(line) - 1
				}
				fn(keyName, processCallLoop(isPws, isCmd, keyName, line[:endArray]))
				line = line[endArray+1:]
			case '"':
				line = line[1:]
				endArray := strings.IndexRune(line, '"')
				if endArray == -1 {
					endArray = len(line) - 1
				}

				fn(keyName, processCallLoop(isPws, isCmd, keyName, strings.Trim(line[:endArray], `"`)))
				line = line[endArray+1:]
			default:
				endString := strings.IndexFunc(line, unicode.IsSpace)
				if endString == -1 {
					endString = len(line) - 1
				}
				fn(keyName, processCallLoop(isPws, isCmd, keyName, line[:endString]))
				line = line[endString+1:]
			}
		}
	}
}

func processLineCall(isPws, isCmd bool, line string, fn func(varName, defaultValue string, valueSeek, valueSeekEnd int) string) string {
	return processLineCallIndex(isPws, isCmd, line, func(startVarName, endVarName, defaultValueAt, valueSeek, valueSeekEnd int, varName, defaultValue string) string {
		return fn(varName, defaultValue, valueSeek, valueSeekEnd)
	})
}

func processLineCallIndex(isPws, isCmd bool, line string, fn func(startVarName, endVarName, defaultValueAt, valueSeek, valueSeekEnd int, varName, defaultValue string) string) string {
	varNameStart := -1
	if isCmd {
		for linePoint := 0; linePoint < len(line); linePoint++ {
			if line[linePoint] == '%' {
				if varNameStart == -1 {
					varNameStart = linePoint
					continue
				} else {
					keyName := line[varNameStart+1 : linePoint]
					if keyName == "" || strings.ContainsFunc(keyName, isInvalidVarName) {
						varNameStart = linePoint
						continue
					}

					valueOf := fn(varNameStart+1, linePoint, -1, -1, -1, keyName, "")
					if valueOf != "" {
						line = line[:varNameStart] + valueOf + line[linePoint+1:]
						linePoint = varNameStart - 1
						varNameStart = -1
					}
				}
			}
		}
		return line
	}
	for linePoint := 0; linePoint < len(line); linePoint++ {
		if line[linePoint] == '$' {
			varNameStart := linePoint
			linePoint++
			if !isPws && line[linePoint] == '{' {
				varNameStart += 2
				linePoint = varNameStart
				for linePoint < len(line)-1 && line[linePoint] != '}' {
					linePoint++
				}

				keyName, defaultValue, splitVar := strings.TrimSpace(line[varNameStart:linePoint]), "", -1
				if keyName == "" || strings.ContainsFunc(keyName, isInvalidVarName) {
					if splitVar = strings.IndexRune(keyName, ':'); splitVar > 0 {
						defaultValue = keyName[splitVar+1:]
						keyName = keyName[:splitVar]
					}
					if !(strings.Contains(keyName, "[") && strings.Contains(keyName, "]")) && (keyName == "" || strings.ContainsFunc(keyName, isInvalidVarName)) {
						varNameStart = linePoint
						continue
					}
				}

				switch strings.Count(defaultValue, ":") {
				case 1:
					seeks := strings.SplitN(defaultValue, ":", 2)
					seek, _ := strconv.Atoi(seeks[0])
					seekEnd, _ := strconv.Atoi(seeks[1])
					valueOf := fn(varNameStart, linePoint+1, splitVar, seek, seekEnd, keyName, "")
					if valueOf != "" {
						line = line[:varNameStart-2] + valueOf + line[linePoint+1:]
						linePoint = varNameStart - 2
					}
				default:
					if len(defaultValue) >= 1 && defaultValue[0] == '-' {
						defaultValue = defaultValue[1:]
					}
					valueOf := fn(varNameStart, linePoint+1, splitVar, -1, -1, keyName, defaultValue)
					if valueOf != "" {
						line = line[:varNameStart-2] + valueOf + line[linePoint+1:]
						linePoint = varNameStart - 2
					}
				}
				continue
			}
			for linePoint < len(line)-1 && isVarName(rune(line[linePoint])) {
				linePoint++
			}
			if line[linePoint] == '=' {
				continue
			}
			keyName := strings.TrimSpace(line[varNameStart+1 : linePoint])
			if keyName == "" || strings.ContainsFunc(keyName, isInvalidVarName) {
				continue
			}
			valueOf := fn(varNameStart, linePoint, -1, -1, -1, keyName, "")
			if valueOf != "" {
				line = line[:varNameStart] + valueOf + line[linePoint:]
				linePoint = varNameStart
			}
		}
	}
	return line
}
