package sh

import (
	"iter"
	"strings"
	"unicode"
)

var (
	_ ProcessSh = &CmdProcess{}
	_ Value     = CmdValue{}
)

// Cmd creates a new CMD script processor.
func Cmd(content string) ProcessSh {
	return CmdWithValues(content, nil)
}

// CmdWithValues creates a new CMD script processor with initial values.
// It preprocesses the script to handle line continuations (^).
func CmdWithValues(scriptFile string, previusValues []Value) ProcessSh {
	// Preprocess to handle line continuations (^)
	rawLines := strings.Split(strings.ReplaceAll(scriptFile, "\r\n", "\n"), "\n")
	processedLines := []string{}
	i := 0
	for i < len(rawLines) {
		line := rawLines[i]
		// Trim trailing spaces before checking for ^
		trimmedLine := strings.TrimRightFunc(line, unicode.IsSpace)
		if strings.HasSuffix(trimmedLine, "^") {
			// Remove ^ and potentially merge with next line
			prefix := trimmedLine[:len(trimmedLine)-1]
			i++
			if i < len(rawLines) {
				// Append next line directly without adding a separator
				rawLines[i] = prefix + rawLines[i]
				continue // Stay on the potentially modified next line index
			} else {
				// ^ was on the last line, just keep the prefix
				processedLines = append(processedLines, prefix)
				// break // No more lines
			}
		} else {
			processedLines = append(processedLines, line) // Keep original line if no ^
			i++
		}
	}

	// Normalize variable names in initial values to uppercase
	normalizedValues := make([]Value, len(previusValues))
	for i, v := range previusValues {
		normalizedValues[i] = &CmdValue{
			Type:          v.ValueType(),
			Name:          strings.ToUpper(v.KeyName()), // Normalize name
			Value:         v.String(),
			OriginalValue: v.String(), // Assuming original value doesn't need separate tracking here
		}
	}

	return &CmdProcess{
		CurrentLine:   0,
		previusValues: normalizedValues,
		scriptFile:    processedLines, // Use preprocessed lines
	}
}

// cmdPoint stores the state for rollback.
type cmdPoint struct {
	currentLine   int
	previusValues []Value
}

// CmdProcess holds the state for processing a CMD script.
type CmdProcess struct {
	CurrentLine int

	scriptFile    []string
	previusValues []Value

	savePoint []*cmdPoint // Stores saved states for Roolback
}

// Back moves the current line pointer one step back.
func (cmd *CmdProcess) Back() { cmd.Add(-1) }

// Add moves the current line pointer by 'dir' steps.
func (cmd *CmdProcess) Add(dir int) {
	targetLine := cmd.CurrentLine - dir // Consistent with BashProcess direction
	if targetLine < 0 {
		targetLine = 0
	}
	if targetLine >= len(cmd.scriptFile) {
		// Allow setting index just past the end? Or clamp to last line?
		// Clamping to last valid index seems safer.
		if len(cmd.scriptFile) > 0 {
			targetLine = len(cmd.scriptFile) - 1
		} else {
			targetLine = 0 // No lines, index is 0
		}
	}
	cmd.CurrentLine = targetLine
}

// Rollback restores the processor state to the last SavePoint.
func (cmd *CmdProcess) Rollback() {
	if len(cmd.savePoint) > 0 {
		last := cmd.savePoint[len(cmd.savePoint)-1]
		cmd.savePoint = cmd.savePoint[:len(cmd.savePoint)-1] // Pop the last save point
		cmd.CurrentLine = last.currentLine
		cmd.previusValues = last.previusValues // Restore values state
	}
}

// SavePoint saves the current processor state.
func (cmd *CmdProcess) SavePoint() {
	// Create a copy of the previusValues slice to prevent modifications
	// in the main processor affecting the saved state.
	valuesCopy := make([]Value, len(cmd.previusValues))
	copy(valuesCopy, cmd.previusValues)

	cmd.savePoint = append(cmd.savePoint, &cmdPoint{
		currentLine:   cmd.CurrentLine,
		previusValues: valuesCopy,
	})
}

// SetKey adds or updates a variable in the processor's state.
// Normalizes the key name to uppercase.
func (cmd *CmdProcess) SetKey(keyName, value string) {
	normalizedKey := strings.ToUpper(keyName)
	// Process the value for substitutions before setting
	processedValue := processCmdValueSubstitutions(value, cmd.previusValues)

	// Check if key already exists to update it, otherwise append.
	// For simplicity here, we just append. findValueCmd will find the latest.
	cmd.previusValues = append(cmd.previusValues, &CmdValue{
		Type:          Set,
		Name:          normalizedKey,    // Store normalized name
		Value:         processedValue, // Store processed value
		OriginalValue: value,          // Store original value as provided
	})
}

// findValueCmd searches for a variable case-insensitively (by checking uppercase).
// It returns the latest set value for the key.
func findValueCmd(values []Value, keyName string) (Value, bool) {
	upperKey := strings.ToUpper(keyName)
	// Iterate backwards to find the latest definition
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		// Assumes names in previusValues are already normalized (uppercase)
		if v.KeyName() == upperKey && v.ValueType().IsSet() {
			return v, true
		}
	}
	return nil, false
}

// processCmdValueSubstitutions takes a string value and performs %VAR% substitutions.
// It returns the processed string.
func processCmdValueSubstitutions(value string, currentValues []Value) string {
	processedLine := value // Start with the original value
	i := 0
	for i < len(processedLine) {
		if processedLine[i] == '%' {
			startPercent := i
			i++ // Move past first %
			endPercent := strings.IndexRune(processedLine[i:], '%')
			if endPercent != -1 {
				// Found potential %VAR%
				varName := processedLine[i : i+endPercent]
				if varName != "" { // Ensure non-empty name like %%
					normalizedKey := strings.ToUpper(varName)
					substitutedValue := "" // Default if not found
					if v, ok := findValueCmd(currentValues, normalizedKey); ok {
						substitutedValue = v.String()
					} else {
						// Variable not found - replace with empty string
						substitutedValue = ""
					}

					// Perform substitution
					startContent := processedLine[:startPercent]
					endContent := processedLine[i+endPercent+1:] // Content after second %
					processedLine = startContent + substitutedValue + endContent

					// Adjust index to the point *after* the substitution
					i = startPercent + len(substitutedValue)
					continue // Continue scanning from the new position
				}
				// else: empty name %%, treat literally? Move past second %
				i += endPercent + 1
			} else {
				// No closing %, treat first % literally
				// Index `i` is already past the first %
				continue
			}
		} else {
			i++ // Move to next character
		}
	}
	return processedLine
}

// Seq processes lines sequentially, handling variable access and setting.
func (cmd *CmdProcess) Seq(limit ...int) Sh {
	limitRead := -1 // Default: read all lines
	return func(yield func(string, []Value) bool) {
		if len(limit) >= 1 {
			cmd.SavePoint()       // Save state before potentially limiting/moving
			defer cmd.Rollback()  // Ensure state is restored after limited sequence
			limitRead = limit[0]   // Number of lines to read
			if len(limit) > 1 {
				cmd.Add(limit[1]) // Move current line (relative jump)
			}
		}

		linesProcessed := 0
		for cmd.CurrentLine < len(cmd.scriptFile) {
			// Check limit if applicable
			if limitRead != -1 {
				if linesProcessed >= limitRead {
					return // Limit reached
				}
				linesProcessed++
			}

			line := cmd.scriptFile[cmd.CurrentLine]
			cmd.CurrentLine++ // Move to next line for the *next* iteration

			lineValues := []Value{} // Values found or set on this line

			// --- Skip Comment Lines ---
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "REM") || strings.HasPrefix(trimmedLine, "::") {
				// Treat comments as yielding the original line but finding no values
				if !yield(line, lineValues) {
					return // Stop if consumer requests
				}
				continue // Move to the next line
			}

			// --- Process SET command ---
			// Use case-insensitive check for "SET "
			if len(trimmedLine) > 4 && strings.EqualFold(trimmedLine[:4], "SET ") {
				setArgs := strings.TrimSpace(trimmedLine[4:])

				var keyName string
				var rawValue string // Value exactly as parsed after =
				valueType := Set

				eqIndex := strings.IndexRune(setArgs, '=')

				// Handle SET "VAR=value" syntax
				if strings.HasPrefix(setArgs, "\"") && eqIndex > 0 && strings.HasSuffix(setArgs, "\"") {
					innerArgs := setArgs[1 : len(setArgs)-1]
					eqIndex = strings.IndexRune(innerArgs, '=')
					if eqIndex != -1 {
						keyName = innerArgs[:eqIndex]
						rawValue = innerArgs[eqIndex+1:]
						// Quotes were around the assignment, not part of value
					} else { // Malformed SET "VAR value"
						keyName = innerArgs // Treat whole thing as key? Or skip? Skip for now.
						rawValue = ""
					}
				} else if eqIndex != -1 {
					// Handle SET VAR=value or SET VAR="quoted value"
					keyName = setArgs[:eqIndex]
					rawValue = setArgs[eqIndex+1:]
					// Check if value is quoted, typically quotes become part of value in CMD
					// No modification here, store raw value including quotes if present
				} else {
					// Handle SET VAR (displays variable, not setting) or SET (displays all)
					// We are only parsing SET VAR=..., so skip lines without '='
					if !yield(line, lineValues) { return } // Yield original line, no values set/accessed
					continue
				}

				// Validate keyName (basic check) - CMD vars usually don't contain '='
				if keyName != "" && !strings.ContainsRune(keyName, '=') {
					normalizedKey := strings.ToUpper(keyName)

					// --- Process value for substitutions ---
					processedValue := processCmdValueSubstitutions(rawValue, cmd.previusValues)
					// --- End process value ---

					shInfo := &CmdValue{
						Type:          valueType,
						Name:          normalizedKey,    // Store normalized name
						Value:         processedValue, // Store processed value
						OriginalValue: rawValue,       // Store original raw value
					}
					cmd.previusValues = append(cmd.previusValues, shInfo)
					lineValues = append(lineValues, shInfo)
					// Don't modify the line itself, just record the SET operation
				}
				// SET command processed, continue to variable expansion on the *same line*
			}

			// --- Process Variable Access (%VAR%) on the rest of the line ---
			// Note: This re-processes the line *after* a SET command might have been handled.
			// This ensures variables set on the same line can be used later on that line,
			// and also expands any other variables present.
			processedLine := line // Work on a copy for this phase if needed, but modifying `line` is current pattern
			i := 0
			for i < len(processedLine) {
				if processedLine[i] == '%' {
					startPercent := i
					i++ // Move past first %
					endPercent := strings.IndexRune(processedLine[i:], '%')
					if endPercent != -1 {
						// Found potential %VAR%
						varName := processedLine[i : i+endPercent]
						if varName != "" { // Ensure non-empty name like %%
							normalizedKey := strings.ToUpper(varName)

							// Create Access record regardless of found value
							shInfo := &CmdValue{Type: Access, Name: normalizedKey, OriginalValue: "%" + varName + "%"}

							substitutedValue := "" // Default if not found
							// found := false
							if v, ok := findValueCmd(cmd.previusValues, normalizedKey); ok {
								substitutedValue = v.String()
								shInfo.Value = substitutedValue // Record the found value
								// found = true
							} else {
								// Variable not found - replace with empty string
								shInfo.Value = "" // Record that it resolved to empty
								substitutedValue = ""
							}

							// Perform substitution in the main line being processed for yielding
							startContent := processedLine[:startPercent]
							endContent := processedLine[i+endPercent+1:] // Content after second %
							processedLine = startContent + substitutedValue + endContent

							// Add access record
							// Avoid adding duplicate access records if SET already added one implicitly via processCmdValueSubstitutions
							// For simplicity, we might add duplicates now, or add logic to check lineValues
							cmd.previusValues = append(cmd.previusValues, shInfo)
							lineValues = append(lineValues, shInfo)

							// Adjust index to the point *after* the substitution
							i = startPercent + len(substitutedValue)
							continue // Continue scanning from the new position
						}
						// else: empty name %%, treat literally? Move past second %
						i += endPercent + 1

					} else {
						// No closing %, treat first % literally
						// Index `i` is already past the first %
						continue
					}
				} else {
					i++ // Move to next character
				}
			} // End variable access loop

			// Yield the processed line and the values found/set on it
			if !yield(processedLine, lineValues) {
				return // Consumer requested stop
			}
		} // End line processing loop
	}
}

// CmdValue holds information about a variable access or set operation in CMD.
type CmdValue struct {
	Type          RawType // Set or Access
	Name          string  // Variable name (normalized to uppercase)
	Value         string  // Processed value (substituted or set)
	OriginalValue string  // Original syntax (%VAR%) or original set value string
	// CMD doesn't have native arrays like Bash/PS, so Seek/SeekEnd are omitted
}

// ValueType returns the type of operation (Set or Access).
func (cv CmdValue) ValueType() RawType { return cv.Type }

// KeyName returns the name of the variable (normalized).
func (cv CmdValue) KeyName() string { return cv.Name }

// String returns the processed value.
func (cv CmdValue) String() string { return cv.Value }

// Array for CMD variables simply yields the single value.
func (cv CmdValue) Array() iter.Seq[string] {
	return func(yield func(string) bool) {
		yield(cv.Value) // Yield the single value as a one-element sequence
	}
}