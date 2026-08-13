// Package isglob ports the npm packages is-glob 4.0.3 and the is-extglob
// 2.1.1 it pulls in: the question "was this string written as a glob pattern,
// or is it a plain path?".
//
// An ESLint plugin asks it to decide how to read a rule option — a value that
// is a glob gets matched, and a value that is not gets compared. Answering it
// differently from upstream moves the line between the two, so the port
// follows the original character by character rather than reaching for
// something that merely looks equivalent.
//
// Only the strict mode is covered, which is the default and the one every
// caller in this repository wants.
package isglob

import "strings"

// Is reports whether str was written as a glob pattern.
func Is(str string) bool {
	if str == "" {
		return false
	}
	if IsExtglob(str) {
		return true
	}
	return strictCheck(str)
}

func strictCheck(str string) bool {
	if str[0] == '!' {
		return true
	}

	index := 0
	pipeIndex := -2
	closeSquareIndex := -2
	closeCurlyIndex := -2
	closeParenIndex := -2
	backSlashIndex := -2

	for index < len(str) {
		if str[index] == '*' {
			return true
		}

		if charAt(str, index+1) == '?' && strings.IndexByte("].+)", str[index]) >= 0 {
			return true
		}

		if closeSquareIndex != -1 && str[index] == '[' && charAt(str, index+1) != ']' {
			if closeSquareIndex < index {
				closeSquareIndex = indexOfFrom(str, ']', index)
			}
			if closeSquareIndex > index {
				if backSlashIndex == -1 || backSlashIndex > closeSquareIndex {
					return true
				}
				backSlashIndex = indexOfFrom(str, '\\', index)
				if backSlashIndex == -1 || backSlashIndex > closeSquareIndex {
					return true
				}
			}
		}

		if closeCurlyIndex != -1 && str[index] == '{' && charAt(str, index+1) != '}' {
			closeCurlyIndex = indexOfFrom(str, '}', index)
			if closeCurlyIndex > index {
				backSlashIndex = indexOfFrom(str, '\\', index)
				if backSlashIndex == -1 || backSlashIndex > closeCurlyIndex {
					return true
				}
			}
		}

		if closeParenIndex != -1 && str[index] == '(' && charAt(str, index+1) == '?' &&
			strings.IndexByte(":!=", charAt(str, index+2)) >= 0 && charAt(str, index+3) != ')' {
			closeParenIndex = indexOfFrom(str, ')', index)
			if closeParenIndex > index {
				backSlashIndex = indexOfFrom(str, '\\', index)
				if backSlashIndex == -1 || backSlashIndex > closeParenIndex {
					return true
				}
			}
		}

		if pipeIndex != -1 && str[index] == '(' && charAt(str, index+1) != '|' {
			if pipeIndex < index {
				pipeIndex = indexOfFrom(str, '|', index)
			}
			if pipeIndex != -1 && charAt(str, pipeIndex+1) != ')' {
				closeParenIndex = indexOfFrom(str, ')', pipeIndex)
				if closeParenIndex > pipeIndex {
					backSlashIndex = indexOfFrom(str, '\\', pipeIndex)
					if backSlashIndex == -1 || backSlashIndex > closeParenIndex {
						return true
					}
				}
			}
		}

		if str[index] == '\\' {
			open := charAt(str, index+1)
			index += 2
			if closer, ok := closingChar(open); ok {
				if n := indexOfFrom(str, closer, index); n != -1 {
					index = n + 1
				}
			}
			if charAt(str, index) == '!' {
				return true
			}
		} else {
			index++
		}
	}

	return false
}

// charAt returns 0 for out-of-range positions, standing in for the `undefined`
// JavaScript reads past the end of a string.
func charAt(str string, index int) byte {
	if index < 0 || index >= len(str) {
		return 0
	}
	return str[index]
}

func indexOfFrom(str string, char byte, from int) int {
	if from < 0 {
		from = 0
	}
	if from > len(str) {
		return -1
	}
	offset := strings.IndexByte(str[from:], char)
	if offset == -1 {
		return -1
	}
	return from + offset
}

func closingChar(open byte) (byte, bool) {
	switch open {
	case '{':
		return '}', true
	case '(':
		return ')', true
	case '[':
		return ']', true
	}
	return 0, false
}
