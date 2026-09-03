package utils

// RegexCapturingGroup is one capturing group found by RegexCapturingGroups:
// either a plain `(...)` group or a named `(?<name>...)` group. Start is the
// byte offset of the group's opening `(` within the pattern; End is one past
// the matching `)`.
type RegexCapturingGroup struct {
	Start int
	End   int
	Name  string // "" for an unnamed group
}

// RegexCapturingGroups validates an ECMAScript regex pattern and returns every
// capturing group in source order. Syntax validation belongs to
// IsValidRegexPattern; after that succeeds, this function only locates escapes,
// character-class boundaries, and matching group delimiters. Keeping metadata
// extraction separate avoids maintaining a second regular-expression parser.
func RegexCapturingGroups(pattern string, flags RegexFlags) (groups []RegexCapturingGroup, ok bool) {
	if !IsValidRegexPattern(pattern, flags) {
		return nil, false
	}

	groupStack := make([]int, 0, 8)
	for i := 0; i < len(pattern); {
		switch pattern[i] {
		case '\\':
			step, valid := SkipPatternEscape(pattern, i, flags)
			if !valid {
				return nil, false
			}
			i += step
		case '[':
			end, valid := ClassEnd(pattern, i, flags)
			if !valid {
				return nil, false
			}
			i = end
		case '(':
			name, captures, valid := regexCapturingGroupAt(pattern, i)
			if !valid {
				return nil, false
			}
			groupIndex := -1
			if captures {
				groups = append(groups, RegexCapturingGroup{Start: i, Name: name})
				groupIndex = len(groups) - 1
			}
			groupStack = append(groupStack, groupIndex)
			i++
		case ')':
			if len(groupStack) == 0 {
				return nil, false
			}
			last := len(groupStack) - 1
			if groupIndex := groupStack[last]; groupIndex >= 0 {
				groups[groupIndex].End = i + 1
			}
			groupStack = groupStack[:last]
			i++
		default:
			i++
		}
	}
	if len(groupStack) != 0 {
		return nil, false
	}
	return groups, true
}

// regexCapturingGroupAt classifies a validated group opener. readAngleName
// keeps the raw spelling of a named capture while validation remains solely
// responsible for checking its IdentifierName grammar.
func regexCapturingGroupAt(pattern string, start int) (name string, captures bool, ok bool) {
	if start+1 >= len(pattern) || pattern[start+1] != '?' {
		return "", true, true
	}
	if start+2 >= len(pattern) || pattern[start+2] != '<' {
		return "", false, true
	}
	if start+3 < len(pattern) && (pattern[start+3] == '=' || pattern[start+3] == '!') {
		return "", false, true
	}
	name, _, ok = readAngleName(pattern, start+2)
	return name, true, ok
}

// These ASCII helpers are shared with the character-class boundary scanner.
func isASCIILetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isRegexDigit(c byte) bool { return c >= '0' && c <= '9' }
