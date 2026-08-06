package minimatch

import (
	"strconv"
	"strings"
)

// Placeholders standing in for the characters a pattern escaped with a
// backslash, so brace parsing never treats them as syntax. They carry a NUL
// byte because a glob pattern cannot.
const (
	escSlash  = "\x00SLASH\x00"
	escOpen   = "\x00OPEN\x00"
	escClose  = "\x00CLOSE\x00"
	escComma  = "\x00COMMA\x00"
	escPeriod = "\x00PERIOD\x00"
)

// BraceExpand is a port of the npm package brace-expansion: it expands the
// brace sets of a pattern into the list of patterns they stand for.
//
//	a{b,c}d       -> abd acd
//	a{b,}c        -> abc ac
//	a{0..3}d      -> a0d a1d a2d a3d
//	a{b,c{d,e}f}g -> abg acdfg acefg
//
// Invalid sets are left alone: `a{2..}b` and `a{b}c` expand to themselves.
func BraceExpand(str string) []string {
	if str == "" {
		return []string{}
	}

	// Bash 4.3 preserves the first two bytes of a pattern starting with `{}`,
	// but only at the top level: `{},a}b` expands to nothing while `a{},b}c`
	// expands to `a}c` and `abc`.
	if strings.HasPrefix(str, "{}") {
		str = `\{\}` + str[2:]
	}

	expanded := expandBraces(escapeBraces(str), true)
	result := make([]string, 0, len(expanded))
	for _, item := range expanded {
		result = append(result, unescapeBraces(item))
	}
	return result
}

var escapeReplacer = strings.NewReplacer(
	`\\`, escSlash,
	`\{`, escOpen,
	`\}`, escClose,
	`\,`, escComma,
	`\.`, escPeriod,
)

var unescapeReplacer = strings.NewReplacer(
	escSlash, `\`,
	escOpen, `{`,
	escClose, `}`,
	escComma, `,`,
	escPeriod, `.`,
)

func escapeBraces(str string) string   { return escapeReplacer.Replace(str) }
func unescapeBraces(str string) string { return unescapeReplacer.Replace(str) }

// parseCommaParts splits on commas, keeping a nested brace section together as
// a single member the way `{a,{b,c},d}` needs.
func parseCommaParts(str string) []string {
	if str == "" {
		return []string{""}
	}

	match, ok := findBalanced("{", "}", str)
	if !ok {
		return strings.Split(str, ",")
	}

	parts := strings.Split(match.pre, ",")
	parts[len(parts)-1] += "{" + match.body + "}"

	postParts := parseCommaParts(match.post)
	if match.post != "" {
		parts[len(parts)-1] += postParts[0]
		parts = append(parts, postParts[1:]...)
	}
	return parts
}

func expandBraces(str string, isTop bool) []string {
	match, ok := findBalanced("{", "}", str)
	if !ok || strings.HasSuffix(match.pre, "$") {
		return []string{str}
	}

	alphaSequence := isAlphaSequence(match.body)
	sequence := alphaSequence || isNumericSequence(match.body)
	options := strings.Contains(match.body, ",")

	if !sequence && !options {
		// `{a},b}` — the brace set really starts after the first `}`.
		if hasCommaBeforeBrace(match.post) {
			return expandBraces(match.pre+"{"+match.body+escClose+match.post, false)
		}
		return []string{str}
	}

	var n []string
	if sequence {
		n = strings.Split(match.body, "..")
	} else {
		n = parseCommaParts(match.body)
		if len(n) == 1 {
			// x{{a,b}}y ==> x{a}y x{b}y
			nested := expandBraces(n[0], false)
			n = make([]string, 0, len(nested))
			for _, item := range nested {
				n = append(n, "{"+item+"}")
			}

			if len(n) == 1 {
				post := []string{""}
				if match.post != "" {
					post = expandBraces(match.post, false)
				}
				values := make([]string, 0, len(post))
				for _, item := range post {
					values = append(values, match.pre+n[0]+item)
				}
				return values
			}
		}
	}

	// At this point n holds the members, and we know it is not a comma set
	// with a single entry. The prefix needs no expansion: it is guaranteed to
	// be free of brace sets.
	post := []string{""}
	if match.post != "" {
		post = expandBraces(match.post, false)
	}

	var members []string
	if sequence {
		members = expandSequence(n, alphaSequence)
	} else {
		for _, item := range n {
			members = append(members, expandBraces(item, false)...)
		}
	}

	expansions := []string{}
	for _, member := range members {
		for _, postItem := range post {
			expansion := match.pre + member + postItem
			if !isTop || sequence || expansion != "" {
				expansions = append(expansions, expansion)
			}
		}
	}
	return expansions
}

// expandSequence expands the `{a..z}` and `{0..9..2}` forms, whose members n
// holds already split on `..`.
func expandSequence(n []string, alpha bool) []string {
	x := sequenceValue(n[0])
	y := sequenceValue(n[1])
	width := max(len(n[0]), len(n[1]))

	incr := 1
	if len(n) == 3 {
		incr = sequenceValue(n[2])
		if incr < 0 {
			incr = -incr
		}
	}
	if incr == 0 {
		return nil
	}

	test := func(i int) bool { return i <= y }
	if y < x {
		incr = -incr
		test = func(i int) bool { return i >= y }
	}

	padded := false
	for _, item := range n {
		padded = padded || isPadded(item)
	}

	members := []string{}
	for i := x; test(i); i += incr {
		var member string
		if alpha {
			member = string(rune(i))
			if member == `\` {
				member = ""
			}
		} else {
			member = strconv.Itoa(i)
			if padded {
				if need := width - len(member); need > 0 {
					if i < 0 {
						member = "-" + strings.Repeat("0", need) + member[1:]
					} else {
						member = strings.Repeat("0", need) + member
					}
				}
			}
		}
		members = append(members, member)
	}
	return members
}

// sequenceValue reads a sequence endpoint as a number, falling back to the
// character code of an alphabetic endpoint.
func sequenceValue(str string) int {
	if value, err := strconv.Atoi(str); err == nil {
		return value
	}
	return int(str[0])
}

// isPadded reports whether a numeric endpoint asks for zero-padded members,
// which it does by carrying a leading zero of its own.
func isPadded(el string) bool {
	el = strings.TrimPrefix(el, "-")
	return len(el) >= 2 && el[0] == '0' && isDigit(el[1])
}

func isNumericSequence(body string) bool {
	parts := strings.Split(body, "..")
	if len(parts) != 2 && len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if !isInteger(part) {
			return false
		}
	}
	return true
}

func isAlphaSequence(body string) bool {
	parts := strings.Split(body, "..")
	if len(parts) != 2 && len(parts) != 3 {
		return false
	}
	if len(parts[0]) != 1 || !isAlpha(parts[0][0]) {
		return false
	}
	if len(parts[1]) != 1 || !isAlpha(parts[1][0]) {
		return false
	}
	return len(parts) == 2 || isInteger(parts[2])
}

// hasCommaBeforeBrace reports whether post carries a comma that a later `}`
// closes, which means the brace set continues past the `}` already consumed.
func hasCommaBeforeBrace(post string) bool {
	comma := strings.IndexByte(post, ',')
	return comma >= 0 && strings.IndexByte(post[comma:], '}') >= 0
}

func isInteger(str string) bool {
	str = strings.TrimPrefix(str, "-")
	if str == "" {
		return false
	}
	for i := range len(str) {
		if !isDigit(str[i]) {
			return false
		}
	}
	return true
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isAlpha(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
