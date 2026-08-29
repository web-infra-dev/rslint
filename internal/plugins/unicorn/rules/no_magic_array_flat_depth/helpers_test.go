// Test helpers shared by the upstream and extras test files. Kept in a
// separate _test.go file so each suite can be compiled and run independently
// while still agreeing on a single source of truth for the depth-token
// position arithmetic and the InvalidTestCase builder.
package no_magic_array_flat_depth_test

import (
	"strings"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID     = "no-magic-array-flat-depth"
	messageString = "Magic number as depth is not allowed."
)

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.mjs"}
}

// depthInvalid reports the depth-argument position (line/column 1-based) of
// the only numeric-literal argument in `code` and asserts the error fires on
// the depth token.
func depthInvalid(code string) rule_tester.InvalidTestCase {
	offset, length := findDepthLiteral(code)
	if offset < 0 {
		panic("no numeric literal found in no-magic-array-flat-depth test: " + code)
	}

	line, column := lineColumnForOffset(code, offset)
	endLine, endColumn := lineColumnForOffset(code, offset+length)

	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mjs",
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID,
			Message:   messageString,
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
		}},
	}
}

// findDepthLiteral returns the byte offset and length of the numeric
// argument passed to the only `.flat(` call in `code`. The upstream test
// cases always use a single literal depth (no expressions), so a simple scan
// for the first numeric-literal-shaped token after the last `.flat` call
// is enough.
func findDepthLiteral(code string) (int, int) {
	flatIndex := strings.LastIndex(code, ".flat")
	if flatIndex < 0 {
		return -1, 0
	}
	openOffset := strings.IndexByte(code[flatIndex+len(".flat"):], '(')
	if openOffset < 0 {
		return -1, 0
	}
	after := flatIndex + len(".flat") + openOffset + 1
	// Skip whitespace and parentheses around the depth literal.
	i := after
	for i < len(code) && (code[i] == ' ' || code[i] == '\t' || code[i] == '\n' || code[i] == '(') {
		i++
	}
	if i >= len(code) {
		return -1, 0
	}
	start := i
	// A numeric literal must start with a digit (or a digit-prefix `0`).
	if !isDigit(code[i]) {
		return -1, 0
	}
	// Detect 0x/0b/0o/0X/0B/0O prefix; consume the prefix and switch the
	// character set for the rest of the literal.
	hexDigits := false
	binDigits := false
	octDigits := false
	if code[i] == '0' && i+1 < len(code) {
		switch code[i+1] {
		case 'x', 'X':
			i += 2
			hexDigits = true
		case 'b', 'B':
			i += 2
			binDigits = true
		case 'o', 'O':
			i += 2
			octDigits = true
		}
	}
	// Consume integer part digits (and `_` separators).
	for i < len(code) {
		c := code[i]
		if c == '_' {
			i++
			continue
		}
		if isDigit(c) {
			i++
			continue
		}
		if hexDigits && isHexLetter(c) {
			i++
			continue
		}
		if binDigits && (c == '0' || c == '1') {
			i++
			continue
		}
		if octDigits && c >= '0' && c <= '7' {
			i++
			continue
		}
		break
	}
	// Decimal fraction and exponent (e.g. `1.5`, `1e2`, `1.5e-3`). These only
	// apply to plain decimal literals, not hex/binary/octal.
	if !hexDigits && !binDigits && !octDigits {
		if i < len(code) && code[i] == '.' && i+1 < len(code) && isDigit(code[i+1]) {
			i++
			for i < len(code) {
				c := code[i]
				if c == '_' || isDigit(c) {
					i++
					continue
				}
				break
			}
		}
		if i < len(code) && (code[i] == 'e' || code[i] == 'E') {
			i++
			if i < len(code) && (code[i] == '+' || code[i] == '-') {
				i++
			}
			for i < len(code) {
				c := code[i]
				if c == '_' || isDigit(c) {
					i++
					continue
				}
				break
			}
		}
	}
	if i == start {
		return -1, 0
	}
	return start, i - start
}

func isDigit(b byte) bool     { return b >= '0' && b <= '9' }
func isHexLetter(b byte) bool { return (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F') }

func lineColumnForOffset(code string, offset int) (int, int) {
	line := 1
	column := 1
	for index := 0; index < offset; {
		r, size := utf8.DecodeRuneInString(code[index:])
		if r == '\n' {
			line++
			column = 1
		} else if r > 0xFFFF {
			column += 2
		} else {
			column++
		}
		index += size
	}
	return line, column
}
