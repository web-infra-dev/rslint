package regexp

import "strings"

// ReplaceFirst replaces the first match using JavaScript replacement-string
// substitutions: $$, $&, $`, $', $1 through $99, and $<name>. Numeric captures
// follow their opening parentheses, including named groups. An unmatched
// capture substitutes the empty string; an unknown numeric capture stays
// literal. The g flag does not request additional replacements from this method.
// Matching uses the same engine semantics as TestOrError. A timeout returns the
// unchanged input and an error, so callers can skip a transformation safely.
func (r *RegExp) ReplaceFirst(input, replacement string) (string, error) {
	if r == nil || r.re == nil {
		return input, nil
	}
	runes := []rune(input)
	match, err := r.re.FindRunesMatch(runes)
	if err != nil || match == nil {
		return input, err
	}
	start, end := match.Index, match.Index+match.Length
	if len(runes) != len(input) {
		start = runeByteIndex(input, match.Index)
		end = start + runeByteIndex(input[start:], match.Length)
	}
	prefix, suffix := input[:start], input[end:]
	var result strings.Builder
	result.Grow(len(input) + len(replacement))
	result.WriteString(prefix)
	for index := 0; index < len(replacement); index++ {
		if replacement[index] != '$' || index+1 == len(replacement) {
			result.WriteByte(replacement[index])
			continue
		}
		next := replacement[index+1]
		switch next {
		case '$':
			result.WriteByte('$')
		case '&':
			result.WriteString(input[start:end])
		case '`':
			result.WriteString(prefix)
		case '\'':
			result.WriteString(suffix)
		case '<':
			end := strings.IndexByte(replacement[index+2:], '>')
			if !r.captures.named || end < 0 {
				result.WriteByte('$')
				continue
			}
			if number, ok := r.captures.names[replacement[index+2:index+2+end]]; ok {
				result.WriteString(match.GroupByNumber(r.captures.number(number)).String())
			}
			index += end + 1
		default:
			if next < '0' || next > '9' {
				result.WriteByte('$')
				continue
			}
			number := int(next - '0')
			if index+2 < len(replacement) && replacement[index+2] >= '0' && replacement[index+2] <= '9' {
				two := number*10 + int(replacement[index+2]-'0')
				if two > 0 && two <= r.captures.count {
					number = two
					index++
				}
			}
			if number == 0 || number > r.captures.count {
				result.WriteByte('$')
				continue
			}
			result.WriteString(match.GroupByNumber(r.captures.number(number)).String())
		}
		index++
	}
	result.WriteString(suffix)
	return result.String(), nil
}

func runeByteIndex(s string, count int) int {
	for index := range s {
		if count == 0 {
			return index
		}
		count--
	}
	return len(s)
}
