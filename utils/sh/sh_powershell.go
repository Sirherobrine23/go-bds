package sh

import (
	"iter"
	"strings"
	"unicode"
)

// Ensure PowershellProcess implements ProcessSh and ShValue implements Value
var (
	_ ProcessSh = &PowershellProcess{}
	// Assuming ShValue replaces BashValue for clarity,
	// or BashValue is kept for compatibility as in the original code.
	// Let's define ShValue for demonstration, but use BashValue if required by external code.
	_ Value = &ShValue{}
)

// Constants for readability
const (
	charDollar      = '$'
	charSingleQuote = '\''
	charDoubleQuote = '"'
	charBacktick    = '`'
	charEquals      = '='
	charDot         = '.'
	charOpenParen   = '('
	charCloseParen  = ')'
	charAt          = '@'
	charSemicolon   = ';'
	charColon       = ':' // For scope, e.g., $env:PATH
)

// ShValue implements the Value interface for PowerShell variables.
// Using this instead of BashValue for better naming convention within this file.
// If BashValue must be used due to external constraints, replace ShValue occurrences with BashValue.
type ShValue struct {
	Type          RawType
	Name          string
	Value         string // Processed value (e.g., after substitution)
	OriginalValue string // Raw value string as found in the script
}

func (sv *ShValue) ValueType() RawType      { return sv.Type }
func (sv *ShValue) String() string          { return sv.Value }
func (sv *ShValue) KeyName() string         { return sv.Name }
func (sv *ShValue) Array() iter.Seq[string] { return splitArrayValue(sv.Value) } // Delegate splitting logic

// Powershell creates a new PowerShell script processor.
func Powershell(content string) ProcessSh {
	return PowershellWithValues(content, nil)
}

// PowershellWithValues creates a new PowerShell script processor with initial values.
func PowershellWithValues(content string, previousValues []Value) ProcessSh {
	// Remove line continuation characters (`^` is not standard PS, usually it's backtick `)
	// Assuming the input might incorrectly use ^` instead of just ` at line end.
	// Standard PowerShell line continuation is ` at the very end of the line.
	// Let's remove backtick followed by newline for robustness.
	content = strings.ReplaceAll(content, "`\n", "") // More standard PS line continuation
	content = strings.ReplaceAll(content, "^\n", "") // Keep original behavior too
	return &PowershellProcess{
		currentLine:    0,
		previousValues: previousValues,
		scriptLines:    strings.Split(content, "\n"),
		savePoints:     []*shState{}, // Initialize slice
	}
}

// PowershellProcess holds the state for parsing a PowerShell script.
type PowershellProcess struct {
	currentLine int
	scriptLines []string

	// Use a copy-on-write approach or careful management for previousValues
	// to ensure SavePoint/Rollback works correctly with shared slice backings.
	// Appending generally handles this okay, but be mindful.
	previousValues []Value
	savePoints     []*shState // Renamed from bashPoint
}

// shState represents a saved state for Rollback.
type shState struct {
	currentLine    int
	previousValues []Value
}

// Back moves the current line pointer one step back.
func (pws *PowershellProcess) Back() {
	pws.SetLine(pws.currentLine - 1)
}

// Add moves the current line pointer by 'delta' steps.
// Deprecated: Prefer SetLine or Next/Back for clarity. Kept for compatibility.
func (pws *PowershellProcess) Add(delta int) {
	pws.SetLine(pws.currentLine + delta)
}

// SetLine sets the current line pointer to a specific line number, clamped within bounds.
func (pws *PowershellProcess) SetLine(newLine int) {
	maxLine := 0
	if len(pws.scriptLines) > 0 {
		maxLine = len(pws.scriptLines) - 1
	}
	// Clamp using max/min is less readable than direct checks
	if newLine < 0 {
		pws.currentLine = 0
	} else if newLine > maxLine {
		pws.currentLine = maxLine
	} else {
		pws.currentLine = newLine
	}
}

// Next moves to the next line, returning false if already at the end.
func (pws *PowershellProcess) Next() bool {
	if pws.currentLine >= len(pws.scriptLines)-1 {
		return false
	}
	pws.currentLine++
	return true
}

// Rollback restores the parser state to the last SavePoint.
func (pws *PowershellProcess) Rollback() {
	if len(pws.savePoints) > 0 {
		last := pws.savePoints[len(pws.savePoints)-1]
		pws.savePoints = pws.savePoints[:len(pws.savePoints)-1]
		pws.currentLine = last.currentLine
		// Restore the exact slice (including capacity potentially)
		pws.previousValues = last.previousValues
	}
}

// SavePoint saves the current parser state.
func (pws *PowershellProcess) SavePoint() {
	// Create a *new* slice with the same elements for Rollback safety.
	// This prevents modifications after SavePoint affecting the saved state.
	valuesCopy := make([]Value, len(pws.previousValues))
	copy(valuesCopy, pws.previousValues)

	pws.savePoints = append(pws.savePoints, &shState{
		currentLine:    pws.currentLine,
		previousValues: valuesCopy,
	})
}

// SetKey adds a new variable definition to the known values.
func (pws *PowershellProcess) SetKey(keyName, value string) {
	// Use ShValue or BashValue as appropriate
	newValue := &ShValue{ // Or &BashValue{...}
		Type:          Set,
		Name:          keyName,
		Value:         value,
		OriginalValue: value, // Original is same as value when set programmatically
	}
	pws.addValue(newValue)
}

// isVarNameChar checks if a rune is valid for a simple PowerShell variable name (excluding scope/properties).
func isVarNameChar(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isScopeSeparator checks for the scope resolution operator.
func isScopeSeparator(r rune) bool {
	return r == charColon
}

// Seq processes the script lines sequentially, yielding each processed line and any variables found.
// limit: Optional. limit[0] = max lines to read (-1 for all). limit[1] = lines to skip initially.
func (pws *PowershellProcess) Seq(limit ...int) Sh {
	maxLines := -1
	skipLines := 0

	if len(limit) >= 1 {
		maxLines = limit[0]
		if len(limit) > 1 {
			skipLines = limit[1]
		}
		// Apply skipLines by advancing the pointer
		pws.SetLine(pws.currentLine + skipLines)

		// Use SavePoint/Rollback only if limits are temporary for this Seq call
		// If Seq is meant to permanently advance state, don't use SavePoint/Rollback here.
		// Assuming temporary based on original code:
		pws.SavePoint()
		defer pws.Rollback()
	}

	return func(yield func(line string, values []Value) bool) {
		linesRead := 0
		for pws.currentLine < len(pws.scriptLines) { // Iterate while within bounds
			if maxLines != -1 && linesRead >= maxLines {
				break // Reached line limit
			}

			currentScriptLine := pws.scriptLines[pws.currentLine]
			processedLine, lineValues := pws.parseLine(currentScriptLine)

			// Advance line pointer *before* yielding
			currentLineIndex := pws.currentLine
			pws.currentLine++ // Move to next line for the *next* iteration or after yield returns false
			linesRead++

			if !yield(processedLine, lineValues) {
				// If yield returns false, restore the line pointer to the line that was just yielded
				pws.currentLine = currentLineIndex
				return // Stop iteration
			}
		}
	}
}

// parseLine processes a single line of PowerShell script.
// It finds variable accesses and assignments.
// Returns the (potentially modified) line and a slice of values found.
func (pws *PowershellProcess) parseLine(line string) (string, []Value) {
	var lineValues []Value
	var processedLine strings.Builder // Use builder for efficient string construction if modifications occur
	processedLine.Grow(len(line))     // Pre-allocate approximate size

	lastIndex := 0 // Tracks the end of the last processed segment

	for i := 0; i < len(line); {
		char := rune(line[i])

		switch char {
		case charSingleQuote:
			// Write content before the quote
			processedLine.WriteString(line[lastIndex:i])
			end := strings.IndexRune(line[i+1:], charSingleQuote)
			if end == -1 { // Unterminated string
				processedLine.WriteString(line[i:]) // Write rest of line
				lastIndex = len(line)
				i = len(line) // Stop processing
			} else {
				end += i + 1                               // Adjust index relative to start of line
				processedLine.WriteString(line[i : end+1]) // Write the quoted string
				i = end + 1
				lastIndex = i
			}

		case charDoubleQuote:
			// Write content before the quote
			processedLine.WriteString(line[lastIndex:i])
			end := findClosingDoubleQuote(line, i+1)
			if end == -1 { // Unterminated string
				processedLine.WriteString(line[i:]) // Write rest of line
				lastIndex = len(line)
				i = len(line) // Stop processing
			} else {
				// Potentially process variables *inside* double quotes later if needed.
				// For now, treat as literal block like single quotes for simplicity.
				processedLine.WriteString(line[i : end+1]) // Write the quoted string
				i = end + 1
				lastIndex = i
			}

		case charDollar:
			// Write content before the '$'
			processedLine.WriteString(line[lastIndex:i])

			// Attempt to parse variable access or assignment
			value, length := pws.parseVariableConstruct(line, i)
			if value != nil {
				// Add found value
				lineValues = append(lineValues, value)
				pws.addValue(value) // Add to global state

				// Handle potential line modification based on access
				if value.ValueType().IsAccess() {
					if resolvedVal, found := findValue(pws.previousValues, value.KeyName()); found {
						// Substitute accessed variable with its value in the processed line
						// Note: This changes the line content, affecting subsequent parsing on this line.
						processedLine.WriteString(resolvedVal.String())
					} else {
						// Variable accessed but not found, write nothing or handle as error?
						// Original code wrote nothing. Let's keep that.
					}
				} else {
					// If it was an assignment, write the original construct? Or nothing?
					// Original code seemed to remove the assignment part from the line.
					// Let's write nothing for assignments here, assuming the goal is
					// extraction, not perfect line reconstruction post-assignment removal.
				}

				i += length // Move index past the parsed construct
				lastIndex = i
			} else {
				// '$' not followed by a valid variable construct, treat as literal '$'
				processedLine.WriteRune(charDollar)
				i++ // Move past the '$'
				lastIndex = i
			}

		default:
			// Regular character, advance index
			i++
		}
	}

	// Append any remaining part of the line
	processedLine.WriteString(line[lastIndex:])

	return processedLine.String(), lineValues
}

// findClosingDoubleQuote handles escaped quotes (`"`") within double-quoted strings.
func findClosingDoubleQuote(line string, start int) int {
	for i := start; i < len(line); i++ {
		switch line[i] {
		case charBacktick:
			// Skip the escaped character (e.g., `"` or `` `)
			if i+1 < len(line) {
				i++ // Skip next character
			}
		case charDoubleQuote:
			return i // Found the closing quote
		}
	}
	return -1 // Closing quote not found
}

// parseVariableConstruct parses a potential variable access or assignment starting at startIndex.
// Returns the parsed Value and the length of the construct in the original string.
func (pws *PowershellProcess) parseVariableConstruct(line string, startIndex int) (Value, int) {
	if startIndex+1 >= len(line) {
		return nil, 0 // '$' at end of line
	}

	// 1. Parse Variable Name (including scope if present)
	nameStart := startIndex + 1
	nameEnd := nameStart
	hasScope := false
	for nameEnd < len(line) {
		r := rune(line[nameEnd])
		if isVarNameChar(r) {
			nameEnd++
		} else if isScopeSeparator(r) && !hasScope && nameEnd > nameStart { // Allow one scope separator
			hasScope = true
			nameEnd++
		} else {
			break // End of variable name
		}
	}

	// Check if a valid name was parsed
	varName := line[nameStart:nameEnd]
	if varName == "" || (hasScope && nameEnd > 0 && line[nameEnd-1] == charColon) { // Empty name or ends with ':'
		return nil, 0
	}

	// 2. Check for property access or assignment
	currentIndex := nameEnd
	originalAccessPath := line[startIndex:nameEnd] // Initial path is just the variable

	// Handle property access (.$prop or .Method())
	for currentIndex < len(line) && line[currentIndex] == charDot {
		propStart := currentIndex + 1
		propEnd := propStart
		for propEnd < len(line) && isVarNameChar(rune(line[propEnd])) {
			propEnd++
		}
		if propEnd == propStart { // No property name after '.'
			break
		}
		// Check for method call parentheses
		methodCallEnd := propEnd
		if propEnd < len(line) && line[propEnd] == charOpenParen {
			// Find closing parenthesis - simple matching for now
			parenDepth := 1
			methodCallEnd = propEnd + 1
			for methodCallEnd < len(line) {
				if line[methodCallEnd] == charOpenParen {
					parenDepth++
				} else if line[methodCallEnd] == charCloseParen {
					parenDepth--
					if parenDepth == 0 {
						methodCallEnd++ // Include the closing parenthesis
						break
					}
				}
				methodCallEnd++
			}
			if parenDepth != 0 {
				methodCallEnd = propEnd // Unterminated method call, treat as property access up to '('
			}
		}
		currentIndex = methodCallEnd
		originalAccessPath = line[startIndex:currentIndex] // Update full path
	}

	// 3. Check for assignment operator (=)
	assignIndex := currentIndex
	for assignIndex < len(line) && isSpace(rune(line[assignIndex])) {
		assignIndex++ // Skip whitespace
	}

	if assignIndex < len(line) && line[assignIndex] == charEquals {
		// --- Assignment Found ---
		valueStartIndex := assignIndex + 1
		for valueStartIndex < len(line) && isSpace(rune(line[valueStartIndex])) {
			valueStartIndex++ // Skip whitespace after '='
		}

		valueEndIndex, valueStr, valueType := pws.parseAssignmentValue(line, valueStartIndex)
		if valueEndIndex == -1 { // Failed to parse value
			return nil, 0
		}

		// Perform substitution on the value string if it contains variables
		processedValueStr := pws.substituteVariables(valueStr)

		// Create the appropriate Value type
		val := &ShValue{ // Or &BashValue{...}
			Type:          valueType, // Set or SetArray
			Name:          varName,   // Base variable name being assigned
			Value:         processedValueStr,
			OriginalValue: valueStr, // Raw value string from script
		}
		return val, valueEndIndex - startIndex // Length includes variable, '=', and value

	} else {
		// --- Access Found ---
		accessType := Access
		if currentIndex > nameEnd { // Properties were accessed
			accessType = AcessWithObject
		}

		val := &ShValue{ // Or &BashValue{...}
			Type:          accessType,
			Name:          varName,            // Base variable name being accessed
			OriginalValue: originalAccessPath, // Full access path ($var or $var.prop)
			// Value field will be populated by findValue if found during line processing substitution
		}
		return val, currentIndex - startIndex // Length of the access path
	}
}

// parseAssignmentValue parses the value part of an assignment.
// Returns the end index of the value, the raw value string, and its type (Set or SetArray).
func (pws *PowershellProcess) parseAssignmentValue(line string, valueStartIndex int) (int, string, RawType) {
	if valueStartIndex >= len(line) {
		return valueStartIndex, "", Set // Assigning empty string effectively
	}

	switch line[valueStartIndex] {
	case charSingleQuote:
		end := strings.IndexRune(line[valueStartIndex+1:], charSingleQuote)
		if end == -1 {
			return -1, "", 0 // Unterminated string
		}
		end += valueStartIndex + 1 // Adjust index
		value := line[valueStartIndex+1 : end]
		return end + 1, value, Set

	case charDoubleQuote:
		end := findClosingDoubleQuote(line, valueStartIndex+1)
		if end == -1 {
			return -1, "", 0 // Unterminated string
		}
		value := line[valueStartIndex+1 : end] // Keep escapes for now, maybe process later
		// Remove PowerShell escape characters (like `"`) within the string if needed
		// value = strings.ReplaceAll(value, "`\"", "\"")
		// value = strings.ReplaceAll(value, "``", "`")
		return end + 1, value, Set

	case charAt: // Potential array @(...) or hashtable @{...}
		if valueStartIndex+1 < len(line) {
			if line[valueStartIndex+1] == charOpenParen { // Array @(...)
				end := findMatchingParen(line, valueStartIndex+1)
				if end == -1 {
					return -1, "", 0 // Unterminated array
				}
				value := line[valueStartIndex+2 : end] // Content inside @(...)
				return end + 1, strings.TrimSpace(value), SetArray
			}
			// TODO: Add support for hashtables @{...} if needed
		}
		// Fall through if not a recognized @ construct

	default: // Simple value (word, number, or potentially another variable)
		// Read until whitespace or semicolon (end of statement)
		end := valueStartIndex
		for end < len(line) && !isSpace(rune(line[end])) && line[end] != charSemicolon {
			// Need to handle cases like $var=$anotherVar.Property - this simple loop isn't enough.
			// Let's find the end based on space or semicolon for now.
			end++
		}
		value := line[valueStartIndex:end]
		return end, value, Set // Treat as a single value initially
	}

	// If it falls through, treat as a single word value
	end := valueStartIndex
	for end < len(line) && !isSpace(rune(line[end])) && line[end] != charSemicolon {
		end++
	}
	value := line[valueStartIndex:end]
	return end, value, Set
}

// findMatchingParen finds the matching closing parenthesis for an opening one at startIndex.
// Handles nested parentheses. Returns -1 if not found.
func findMatchingParen(line string, startIndex int) int {
	if startIndex >= len(line) || line[startIndex] != charOpenParen {
		return -1
	}
	depth := 1
	for i := startIndex + 1; i < len(line); i++ {
		switch line[i] {
		case charOpenParen:
			depth++
		case charCloseParen:
			depth--
			if depth == 0 {
				return i // Found matching parenthesis
			}
		}
	}
	return -1 // Not found
}

// substituteVariables replaces $var constructs within a string with their known values.
// This is a simplified version, doesn't handle complex expressions or nested quotes perfectly.
func (pws *PowershellProcess) substituteVariables(valueStr string) string {
	// Basic substitution: find $var, look up, replace.
	// This won't handle `${var}` or complex scenarios well without a more robust parser.
	var result strings.Builder
	result.Grow(len(valueStr))
	lastIndex := 0
	for i := 0; i < len(valueStr); {
		if valueStr[i] == charDollar {
			result.WriteString(valueStr[lastIndex:i]) // Write text before $

			nameStart := i + 1
			nameEnd := nameStart
			hasScope := false
			for nameEnd < len(valueStr) {
				r := rune(valueStr[nameEnd])
				if isVarNameChar(r) {
					nameEnd++
				} else if isScopeSeparator(r) && !hasScope && nameEnd > nameStart {
					hasScope = true
					nameEnd++
				} else {
					break
				}
			}
			varName := valueStr[nameStart:nameEnd]

			if varName != "" && !(hasScope && valueStr[nameEnd-1] == charColon) {
				if val, ok := findValue(pws.previousValues, varName); ok {
					result.WriteString(val.String()) // Substitute with found value
				} else {
					// Variable not found, write nothing or original '$var'?
					// Let's write nothing, consistent with access logic.
				}
				i = nameEnd // Move past the variable name
				lastIndex = i
			} else {
				// Not a valid variable name after '$', write '$' literally
				result.WriteByte(charDollar)
				i++
				lastIndex = i
			}
		} else {
			i++ // Move to next character
		}
	}
	result.WriteString(valueStr[lastIndex:]) // Write remaining part
	return result.String()
}

// addValue appends a value to the internal list, ensuring uniqueness if needed (optional).
// Currently just appends. Could be extended to update existing keys.
func (pws *PowershellProcess) addValue(v Value) {
	// Simple append. If updates are needed, find existing key first.
	pws.previousValues = append(pws.previousValues, v)
}

// Helper to split comma/space separated array values, handling quotes.
func splitArrayValue(value string) iter.Seq[string] {
	return func(yield func(string) bool) {
		remaining := strings.TrimSpace(value)
		for len(remaining) > 0 {
			var item string
			var advance int

			r := rune(remaining[0])
			if isSpace(r) || r == ',' {
				remaining = strings.TrimLeftFunc(remaining, func(r rune) bool { return isSpace(r) || r == ',' })
				continue
			}

			switch r {
			case charSingleQuote:
				end := strings.IndexRune(remaining[1:], charSingleQuote)
				if end == -1 {
					// Unterminated, yield rest? Or error? Yield rest for robustness.
					item = remaining[1:]
					advance = len(remaining)
				} else {
					item = remaining[1 : end+1]
					advance = end + 2 // Past the closing quote
				}
			case charDoubleQuote:
				end := findClosingDoubleQuote(remaining, 1)
				if end == -1 {
					// Unterminated, yield rest?
					item = remaining[1:]
					advance = len(remaining)
				} else {
					item = remaining[1:end] // Content inside quotes
					// Potentially unescape `"` etc. here if needed
					advance = end + 1 // Past the closing quote
				}
			default: // Unquoted item
				end := strings.IndexFunc(remaining, func(r rune) bool { return isSpace(r) || r == ',' })
				if end == -1 {
					item = remaining
					advance = len(remaining)
				} else {
					item = remaining[:end]
					advance = end
				}
			}

			if !yield(item) {
				return
			}
			if advance >= len(remaining) {
				break
			}
			remaining = remaining[advance:]
			remaining = strings.TrimLeftFunc(remaining, func(r rune) bool { return isSpace(r) || r == ',' })
		}
	}
}
