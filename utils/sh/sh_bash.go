package sh

import (
	"iter"
	"strconv"
	"strings"
	"unicode"
)

var (
	_ ProcessSh = &BashProcess{}
	_ Value     = BashValue{}
)

// Proces Bash style of scripts
func Bash(content string) ProcessSh {
	return BashWithValues(content, nil)
}

// Proces Bash style of scripts
func BashWithValues(scriptFile string, previusValues []Value) ProcessSh {
	scriptFile = strings.ReplaceAll(scriptFile, "\\\n", "")
	return &BashProcess{
		CurrentLine:   0,
		previusValues: previusValues,
		scriptFile:    strings.Split(scriptFile, "\n"),
	}
}

type bashPoint struct {
	currentLine   int
	previusValues []Value
}

type BashProcess struct {
	CurrentLine int

	scriptFile    []string
	previusValues []Value

	savePoint []*bashPoint
}

func (bash *BashProcess) Back() { bash.Add(-1) }
func (bash *BashProcess) Add(dir int) {
	bash.CurrentLine = max(min(len(bash.scriptFile)-1, bash.CurrentLine-dir), 0)
}

func (bash *BashProcess) Rollback() {
	if len(bash.savePoint) > 0 {
		last := bash.savePoint[len(bash.savePoint)-1]
		bash.savePoint = bash.savePoint[:len(bash.savePoint)-1]
		bash.CurrentLine = last.currentLine
		bash.previusValues = last.previusValues
	}
}

func (bash *BashProcess) SavePoint() {
	bash.savePoint = append(bash.savePoint, &bashPoint{
		currentLine:   bash.CurrentLine,
		previusValues: bash.previusValues[:],
	})
}

func (bash *BashProcess) SetKey(keyName, value string) {
	bash.previusValues = append(bash.previusValues, &BashValue{
		Type:          Set,
		Name:          keyName,
		Value:         value,
		OriginalValue: value,
	})
}

func (bash *BashProcess) Seq(limit ...int) Sh {
	limitRead := -1
	return func(yield func(string, []Value) bool) {
		if len(limit) >= 1 {
			bash.SavePoint()
			defer bash.Rollback()
			limitRead = limit[0]
			if len(limit) > 1 {
				bash.Add(limit[1])
			}
		}

		linesProcessed := 0
		for bash.CurrentLine < len(bash.scriptFile) {
			if limitRead != -1 {
				if linesProcessed >= limitRead {
					return
				}
				linesProcessed++
			}

			line := bash.scriptFile[bash.CurrentLine]
			bash.CurrentLine++
			lineValues := []Value{}
			skipDoble, skipSingle := false, false
			tempLineIndex := 0
			for tempLineIndex < len(line) {
				currentChar := line[tempLineIndex]
				originalIndex := tempLineIndex

				if currentChar == '\'' && !skipDoble {
					skipSingle = !skipSingle
					tempLineIndex++
					continue
				}
				if currentChar == '"' && !skipSingle {
					skipDoble = !skipDoble
					tempLineIndex++
					continue
				}

				if skipSingle {
					tempLineIndex++
					continue
				}

				switch currentChar {
				case '$':
					startKey := tempLineIndex
					tempLineIndex++
					if tempLineIndex >= len(line) {
						break
					}

					var varName string
					var shInfo *BashValue
					var substitutionMade bool
					var substitutedValue string
					var endIndex int

					switch line[tempLineIndex] {
					case '{':
						braceIndex := tempLineIndex
						tempLineIndex++
						endBrace := strings.IndexRune(line[tempLineIndex:], '}')
						if endBrace == -1 {
							tempLineIndex = startKey + 1
							continue
						}
						endIndex = tempLineIndex + endBrace + 1
						fullVarSyntax := line[braceIndex+1 : endIndex-1]

						if colonPos := strings.Index(fullVarSyntax, ":-"); colonPos != -1 {
							varName = fullVarSyntax[:colonPos]
							defaultValue := fullVarSyntax[colonPos+2:]
							shInfo = &BashValue{Type: Access, Name: varName, OriginalValue: defaultValue}
							if v, ok := findValue(bash.previusValues, varName); ok && v.String() != "" {
								substitutedValue = v.String()
								shInfo.Value = substitutedValue
								substitutionMade = true
							} else {
								substitutedValue = defaultValue
								shInfo.Value = substitutedValue
								substitutionMade = true
							}
						} else if colonPos := strings.Index(fullVarSyntax, ":"); colonPos != -1 && len(fullVarSyntax) > colonPos+1 && unicode.IsDigit(rune(fullVarSyntax[colonPos+1])) {
							varName = fullVarSyntax[:colonPos]
							parts := strings.SplitN(fullVarSyntax[colonPos+1:], ":", 2)
							seek, errSeek := strconv.Atoi(parts[0])
							seekEnd := -1
							if len(parts) == 2 {
								var errSeekEnd error
								seekEnd, errSeekEnd = strconv.Atoi(parts[1])
								if errSeekEnd != nil {
									tempLineIndex = endIndex
									continue
								}
							}
							if errSeek != nil {
								tempLineIndex = endIndex
								continue
							}

							shInfo = &BashValue{Type: Access, Name: varName, Seek: seek, SeekEnd: seekEnd}
							if v, ok := findValue(bash.previusValues, varName); ok {
								baseValue := v.String()
								if seek < 0 {
									seek = len(baseValue) + seek
								}
								if seek < 0 {
									seek = 0
								}
								if seek > len(baseValue) {
									seek = len(baseValue)
								}
								endPos := len(baseValue)
								if seekEnd != -1 {
									endPos = seek + seekEnd
								}
								if endPos > len(baseValue) {
									endPos = len(baseValue)
								}

								if seek >= endPos {
									substitutedValue = ""
								} else {
									substitutedValue = baseValue[seek:endPos]
								}

								shInfo.Value = substitutedValue
								substitutionMade = true
							} else {
								substitutedValue = ""
								shInfo.Value = ""
								substitutionMade = true
							}
						} else {
							varName = fullVarSyntax
							shInfo = &BashValue{Type: Access, Name: varName}
							if v, ok := findValue(bash.previusValues, varName); ok {
								substitutedValue = v.String()
								shInfo.Value = substitutedValue
								substitutionMade = true
							} else {
								substitutedValue = ""
								shInfo.Value = ""
								substitutionMade = true
							}
						}
					case '(':
						tempLineIndex = startKey + 1
						continue
					default:
						scanIndex := tempLineIndex
						varEnd := strings.IndexFunc(line[scanIndex:], isInvalidVarName)
						if varEnd == -1 {
							varEnd = len(line[scanIndex:])
						}
						if varEnd == 0 {
							tempLineIndex = startKey + 1
							continue
						}
						varName = line[scanIndex : scanIndex+varEnd]
						endIndex = scanIndex + varEnd
						shInfo = &BashValue{Type: Access, Name: varName}
						if v, ok := findValue(bash.previusValues, varName); ok {
							substitutedValue = v.String()
							shInfo.Value = substitutedValue
							substitutionMade = true
						} else {
							substitutedValue = ""
							shInfo.Value = ""
							substitutionMade = true
						}
					}

					if shInfo != nil {
						if substitutionMade {
							startContent := line[:startKey]
							endContent := line[endIndex:]
							newLen := len(substitutedValue)
							line = startContent + substitutedValue + endContent
							tempLineIndex = startKey + newLen
						} else {
							startContent := line[:startKey]
							endContent := line[endIndex:]
							line = startContent + endContent
							tempLineIndex = startKey
						}
						bash.previusValues = append(bash.previusValues, shInfo)
						lineValues = append(lineValues, shInfo)
					}
					continue
				case '=':
					if skipDoble || skipSingle || strings.HasPrefix(strings.TrimSpace(line), "for") {
						tempLineIndex++
						continue
					}
					nameEnd := tempLineIndex
					nameStart := nameEnd
					for nameStart > 0 {
						if !isVarName(rune(line[nameStart-1])) {
							if isSpace(rune(line[nameStart-1])) {
								prefixCheckStart := nameStart - 1
								for prefixCheckStart > 0 && isSpace(rune(line[prefixCheckStart-1])) {
									prefixCheckStart--
								}
								if strings.HasSuffix(line[:prefixCheckStart+1], "export") {
									nameStart = strings.LastIndex(line[:nameEnd], "export") + len("export")
									for nameStart < nameEnd && isSpace(rune(line[nameStart])) {
										nameStart++
									}
									break
								}
							}
							break
						}
						nameStart--
					}

					keyName := strings.TrimSpace(line[nameStart:nameEnd])
					if keyName == "" || strings.ContainsFunc(keyName, isInvalidVarName) {
						tempLineIndex++
						continue
					}

					valueStart := nameEnd + 1
					for valueStart < len(line) && isSpace(rune(line[valueStart])) {
						valueStart++
					}
					if valueStart >= len(line) {
						shInfo := &BashValue{Type: Set, Name: keyName, Value: "", OriginalValue: ""}
						bash.previusValues = append(bash.previusValues, shInfo)
						lineValues = append(lineValues, shInfo)
						tempLineIndex = valueStart
						continue
					}

					var valueEnd int
					var assignedValue string
					var originalAssignedValue string
					valueType := Set
					valChar := line[valueStart]
					if valChar == '"' {
						valueStartIndex := valueStart + 1
						endQuote := -1
						searchIdx := valueStartIndex
						for {
							idx := strings.IndexRune(line[searchIdx:], '"')
							if idx == -1 {
								break
							}
							if idx > 0 && line[searchIdx+idx-1] == '\\' {
								searchIdx += idx + 1
							} else {
								endQuote = searchIdx + idx
								break
							}
						}

						if endQuote == -1 {
							tempLineIndex++
							continue
						}
						valueEnd = endQuote + 1
						originalAssignedValue = line[valueStartIndex:endQuote]
						processedValue := originalAssignedValue
						subProcessor := BashWithValues(originalAssignedValue, bash.previusValues)
						subSeq := subProcessor.Seq(1)
						for subLine := range subSeq {
							processedValue = subLine
							break
						}
						assignedValue = processedValue
					} else if valChar == '\'' {
						valueStartIndex := valueStart + 1
						endQuote := strings.IndexRune(line[valueStartIndex:], '\'')
						if endQuote == -1 {
							tempLineIndex++
							continue
						}
						valueEnd = valueStartIndex + endQuote + 1
						originalAssignedValue = line[valueStartIndex : valueEnd-1]
						assignedValue = originalAssignedValue
					} else if valChar == '(' && valueStart > 0 && line[valueStart-1] == '=' {
						valueStartIndex := valueStart + 1
						endParen := strings.IndexRune(line[valueStartIndex:], ')')
						if endParen == -1 {
							tempLineIndex++
							continue
						}
						valueEnd = valueStartIndex + endParen + 1
						originalAssignedValue = line[valueStartIndex : valueEnd-1]
						assignedValue = originalAssignedValue
						valueType = SetArray
					} else {
						valueEnd = valueStart
						inCmdSubst := false
						parenLevel := 0
						for valueEnd < len(line) {
							c := line[valueEnd]
							if c == '$' && valueEnd+1 < len(line) && line[valueEnd+1] == '(' {
								inCmdSubst = true
								parenLevel++
								valueEnd += 2
								continue
							}
							if inCmdSubst {
								if c == '(' {
									parenLevel++
								}
								if c == ')' {
									parenLevel--
								}
								if parenLevel == 0 {
									inCmdSubst = false
								}
								valueEnd++
								continue
							}
							if isSpace(rune(c)) || c == ';' || c == '&' || c == '|' {
								break
							}
							valueEnd++
						}
						originalAssignedValue = line[valueStart:valueEnd]
						processedValue := originalAssignedValue
						subProcessor := BashWithValues(originalAssignedValue, bash.previusValues)
						subSeq := subProcessor.Seq(1)
						for subLine := range subSeq {
							processedValue = subLine
							break
						}
						assignedValue = processedValue
					}
					shInfo := &BashValue{
						Type:          valueType,
						Name:          keyName,
						Value:         assignedValue,
						OriginalValue: originalAssignedValue,
					}
					bash.previusValues = append(bash.previusValues, shInfo)
					lineValues = append(lineValues, shInfo)
					tempLineIndex = valueEnd
					continue
				default:
					tempLineIndex++
				}
				if tempLineIndex == originalIndex {
					tempLineIndex++
				}
			}
			if !yield(line, lineValues) {
				return
			}
		}
	}
}

type BashValue struct {
	Type          RawType // Value type
	Name, Value   string  // Process name and value
	OriginalValue string  // Orignal value string
	Seek, SeekEnd int     // Value slice if string
}

func (shValue BashValue) ValueType() RawType { return shValue.Type }
func (shValue BashValue) KeyName() string    { return shValue.Name }
func (shValue BashValue) String() string     { return shValue.Value }
func (shValue BashValue) Array() iter.Seq[string] {
	if shValue.Type != SetArray {
		return func(yield func(string) bool) { yield(shValue.Value) }
	}

	// @(<Filed1>, <Field2>)
	// @(<Filed1> <Field2>)
	return func(yield func(string) bool) {
		copyContent := shValue.Value[:]
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
