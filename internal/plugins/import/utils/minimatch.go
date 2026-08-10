package utils

import (
	"regexp"
	"strings"
)

// MinimatchOptions is the subset of minimatch options that changes pure
// string matching. Filesystem-walking options do not apply to import
// specifiers and are intentionally absent.
type MinimatchOptions struct {
	NoNegate   bool
	NoComment  bool
	NoCase     bool
	MatchBase  bool
	NoGlobStar bool
	NoExt      bool
	NoBrace    bool
	Dot        bool
}

// MatchMinimatch matches an import specifier using minimatch-compatible glob,
// brace, and extglob syntax.
func MatchMinimatch(name string, pattern string, options MinimatchOptions) bool {
	if !options.NoComment && strings.HasPrefix(pattern, "#") {
		return false
	}
	negated := false
	if !options.NoNegate {
		for strings.HasPrefix(pattern, "!") {
			negated = !negated
			pattern = pattern[1:]
		}
	}
	if options.MatchBase && !strings.Contains(pattern, "/") {
		if slash := strings.LastIndexByte(name, '/'); slash >= 0 {
			name = name[slash+1:]
		}
	}

	patterns := []string{pattern}
	if !options.NoBrace {
		patterns = expandBracePatterns(pattern)
	}
	matched := false
	for _, candidate := range patterns {
		if matchMinimatchPath(name, candidate, options) {
			matched = true
			break
		}
	}
	if negated {
		return !matched
	}
	return matched
}

func matchMinimatchPath(name string, pattern string, options MinimatchOptions) bool {
	names := strings.Split(name, "/")
	patterns := strings.Split(pattern, "/")
	if !options.NoGlobStar {
		patterns = collapseConsecutiveGlobstars(patterns)
	}
	type state struct{ pattern, name int }
	memo := map[state]bool{}
	seen := map[state]bool{}
	var match func(int, int) bool
	match = func(patternIndex, nameIndex int) bool {
		key := state{pattern: patternIndex, name: nameIndex}
		if seen[key] {
			return memo[key]
		}
		seen[key] = true

		if patternIndex == len(patterns) {
			memo[key] = nameIndex == len(names)
			return memo[key]
		}
		part := patterns[patternIndex]
		if part == "**" && !options.NoGlobStar {
			// A trailing `/**` still requires the slash to exist (`foo/**`
			// does not match `foo`), while a leading/bare globstar can match
			// zero segments.
			canMatchZero := patternIndex == 0 || patternIndex+1 < len(patterns) ||
				nameIndex > patternIndex
			if canMatchZero && match(patternIndex+1, nameIndex) {
				memo[key] = true
				return true
			}
			if nameIndex < len(names) && globstarMayConsume(names[nameIndex], options.Dot) && match(patternIndex, nameIndex+1) {
				memo[key] = true
				return true
			}
			return false
		}
		if nameIndex >= len(names) || !matchMinimatchSegment(names[nameIndex], part, options) {
			return false
		}
		memo[key] = match(patternIndex+1, nameIndex+1)
		return memo[key]
	}
	return match(0, 0)
}

func collapseConsecutiveGlobstars(patterns []string) []string {
	result := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		if pattern == "**" && len(result) > 0 && result[len(result)-1] == "**" {
			continue
		}
		result = append(result, pattern)
	}
	return result
}

func globstarMayConsume(segment string, dot bool) bool {
	if segment == "." || segment == ".." {
		return false
	}
	return dot || segment == "" || !strings.HasPrefix(segment, ".")
}

func matchMinimatchSegment(name string, pattern string, options MinimatchOptions) bool {
	if strings.HasPrefix(name, ".") && !patternExplicitlyStartsWithDot(pattern) {
		if name == "." || name == ".." || !options.Dot {
			return false
		}
	}
	builder := globRegexBuilder{options: options}
	if options.NoCase {
		builder.expression.WriteString("(?i)")
	}
	builder.expression.WriteByte('^')
	builder.writeFragment([]rune(pattern))
	builder.expression.WriteByte('$')
	re, err := regexp.Compile(builder.expression.String())
	if err != nil {
		return false
	}
	matches := re.FindStringSubmatch(name)
	matched := matches != nil
	if matched {
		for _, negative := range builder.negatives {
			if negative.group >= len(matches) {
				continue
			}
			candidate := matches[negative.group]
			for _, alternative := range negative.alternatives {
				alternativeOptions := options
				alternativeOptions.NoNegate = true
				alternativeOptions.NoComment = true
				alternativeOptions.MatchBase = false
				if MatchMinimatch(candidate, alternative, alternativeOptions) {
					matched = false
					break
				}
			}
			if !matched {
				break
			}
		}
	}
	return matched
}

func patternExplicitlyStartsWithDot(pattern string) bool {
	if strings.HasPrefix(pattern, ".") || strings.HasPrefix(pattern, `\.`) {
		return true
	}
	return strings.HasPrefix(pattern, "[.]")
}

func expandBracePatterns(pattern string) []string {
	runes := []rune(pattern)
	for index, char := range runes {
		if char != '{' {
			continue
		}
		end, ok := matchingDelimiter(runes, index+1, '{', '}')
		if !ok {
			return []string{pattern}
		}
		alternatives := splitGlobAlternatives(runes[index+1:end], ',')
		if len(alternatives) <= 1 {
			continue
		}
		prefix := string(runes[:index])
		suffix := string(runes[end+1:])
		var result []string
		for _, alternative := range alternatives {
			result = append(result, expandBracePatterns(prefix+string(alternative)+suffix)...)
		}
		return result
	}
	return []string{pattern}
}

type negativeExtglob struct {
	group        int
	alternatives []string
}

type globRegexBuilder struct {
	options    MinimatchOptions
	expression strings.Builder
	groups     int
	negatives  []negativeExtglob
}

func (builder *globRegexBuilder) openGroup() int {
	builder.groups++
	builder.expression.WriteByte('(')
	return builder.groups
}

func (builder *globRegexBuilder) writeFragment(fragment []rune) {
	for index := 0; index < len(fragment); {
		char := fragment[index]
		if !builder.options.NoExt && index+1 < len(fragment) && fragment[index+1] == '(' && strings.ContainsRune("?*+@!", char) {
			if end, ok := matchingDelimiter(fragment, index+2, '(', ')'); ok {
				alternatives := splitGlobAlternatives(fragment[index+2:end], '|')
				if char == '!' {
					group := builder.openGroup()
					builder.expression.WriteString("[^/]*")
					builder.expression.WriteByte(')')
					raw := make([]string, len(alternatives))
					for i, alternative := range alternatives {
						raw[i] = string(alternative)
					}
					builder.negatives = append(builder.negatives, negativeExtglob{group: group, alternatives: raw})
				} else {
					builder.openGroup()
					for i, alternative := range alternatives {
						if i > 0 {
							builder.expression.WriteByte('|')
						}
						builder.writeFragment(alternative)
					}
					builder.expression.WriteByte(')')
					switch char {
					case '?', '*', '+':
						builder.expression.WriteRune(char)
					}
				}
				index = end + 1
				continue
			}
		}
		if char == '{' && !builder.options.NoBrace {
			if end, ok := matchingDelimiter(fragment, index+1, '{', '}'); ok {
				alternatives := splitGlobAlternatives(fragment[index+1:end], ',')
				if len(alternatives) > 1 {
					builder.openGroup()
					for i, alternative := range alternatives {
						if i > 0 {
							builder.expression.WriteByte('|')
						}
						builder.writeFragment(alternative)
					}
					builder.expression.WriteByte(')')
					index = end + 1
					continue
				}
			}
		}

		switch char {
		case '*':
			for index < len(fragment) && fragment[index] == '*' {
				index++
			}
			builder.expression.WriteString("[^/]*")
		case '?':
			builder.expression.WriteString("[^/]")
			index++
		case '[':
			end := index + 1
			for end < len(fragment) && fragment[end] != ']' {
				end++
			}
			if end >= len(fragment) || end == index+1 {
				builder.expression.WriteString(`\[`)
				index++
				continue
			}
			body := string(fragment[index+1 : end])
			if strings.HasPrefix(body, "!") {
				body = "^" + body[1:]
			}
			builder.expression.WriteByte('[')
			builder.expression.WriteString(body)
			builder.expression.WriteByte(']')
			index = end + 1
		case '\\':
			if index+1 < len(fragment) {
				builder.expression.WriteString(regexp.QuoteMeta(string(fragment[index+1])))
				index += 2
			} else {
				builder.expression.WriteString(`\\`)
				index++
			}
		default:
			builder.expression.WriteString(regexp.QuoteMeta(string(char)))
			index++
		}
	}
}

func matchingDelimiter(fragment []rune, start int, open rune, closing rune) (int, bool) {
	depth := 1
	for index := start; index < len(fragment); index++ {
		switch fragment[index] {
		case open:
			depth++
		case closing:
			depth--
			if depth == 0 {
				return index, true
			}
		}
	}
	return -1, false
}

func splitGlobAlternatives(fragment []rune, separator rune) [][]rune {
	var alternatives [][]rune
	start := 0
	parenDepth := 0
	braceDepth := 0
	for index, char := range fragment {
		switch char {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case '{':
			braceDepth++
		case '}':
			braceDepth--
		default:
			if char == separator && parenDepth == 0 && braceDepth == 0 {
				alternatives = append(alternatives, fragment[start:index])
				start = index + 1
			}
		}
	}
	return append(alternatives, fragment[start:])
}
