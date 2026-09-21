// Package htmlscan holds the byte-level scanning HTML markup needs, shared by
// the single file component splitter and the template parser.
//
// It is deliberately not a tokenizer: it answers one question at a time about
// the text at an offset, and every function returns the offset just past what
// it consumed. Both callers walk markup with their own loop because they want
// different things out of it (one wants a component's top-level blocks, the
// other wants a whole element tree), but they must agree about where a tag name
// ends and where a quoted attribute value ends, or the two would disagree about
// the same file.
//
// Case folding here is ASCII-only, because that is the fold HTML specifies for
// a tag or attribute name. It is deliberately not ecmascript.StringToLowerCase:
// that answers what JavaScript's String#toLowerCase would say about arbitrary
// text, which is a different question from what HTML says two names are.
// Attribute values are never folded, matching both HTML and Vue's exact
// comparisons against them.
package htmlscan

import "strings"

// RawTextTags are the elements whose content HTML does not parse as markup, so
// their content ends at the first matching end tag rather than a balanced one.
// A `</script>` inside a string literal therefore closes the block. That is the same
// place Vue's own parser ends it, and the reason SFC authors are told to split
// that sequence.
var RawTextTags = map[string]struct{}{
	"script":   {},
	"style":    {},
	"textarea": {},
	"title":    {},
}

// VoidTags never have content or an end tag.
var VoidTags = map[string]struct{}{
	"area": {}, "base": {}, "br": {}, "col": {}, "embed": {}, "hr": {},
	"img": {}, "input": {}, "link": {}, "meta": {}, "param": {},
	"source": {}, "track": {}, "wbr": {},
}

// IsRawText reports whether tag holds raw text rather than markup. tag must
// already be folded.
func IsRawText(tag string) bool {
	_, ok := RawTextTags[tag]
	return ok
}

// IsVoid reports whether tag never has content or an end tag. tag must already
// be folded.
func IsVoid(tag string) bool {
	_, ok := VoidTags[tag]
	return ok
}

// Attribute is one attribute of a start tag, with every range absolute in the
// text it was read from.
type Attribute struct {
	// Name is the attribute name as written, unfolded. HTML folds a name, but
	// a Vue directive argument is case-sensitive and travels in the same
	// string (`:fooBar` must keep its capital), so folding is left to the
	// caller, which knows which half it is looking at.
	Name     string
	NamePos  int
	NameEnd  int
	HasValue bool
	// Value is the attribute value with its quotes removed and no entity
	// decoding applied.
	Value    string
	ValuePos int
	ValueEnd int
}

// StartTag is a parsed start tag.
type StartTag struct {
	// End is the offset just past the closing `>`.
	End         int
	SelfClosing bool
	Attributes  []Attribute
}

// ReadStartTag consumes a start tag's attributes beginning at index, which must
// sit just past the tag name.
//
// Attribute values are parsed rather than scanned for, so a `>` inside a quoted
// value does not end the tag early, which is the difference between reading
// `<script data-x=">">` correctly and truncating it.
func ReadStartTag(text string, index int) StartTag {
	tag := StartTag{}
	for index < len(text) {
		index = SkipSpace(text, index)
		if index >= len(text) {
			break
		}
		switch text[index] {
		case '>':
			tag.End = index + 1
			return tag
		case '/':
			tag.SelfClosing = true
			index++
			continue
		}

		attribute := Attribute{NamePos: index}
		for index < len(text) && !IsSpace(text[index]) &&
			text[index] != '=' && text[index] != '>' && text[index] != '/' {
			index++
		}
		attribute.NameEnd = index
		attribute.Name = text[attribute.NamePos:index]

		if probe := SkipSpace(text, index); probe < len(text) && text[probe] == '=' {
			attribute.HasValue = true
			attribute.Value, attribute.ValuePos, attribute.ValueEnd, index =
				readAttributeValue(text, probe+1)
		}
		if attribute.Name != "" {
			tag.Attributes = append(tag.Attributes, attribute)
		}
	}
	tag.End = len(text)
	return tag
}

// readAttributeValue reads a quoted or unquoted attribute value beginning at
// index. The returned range covers the value itself, excluding any quotes.
func readAttributeValue(text string, index int) (value string, valuePos, valueEnd, next int) {
	index = SkipSpace(text, index)
	if index >= len(text) {
		return "", len(text), len(text), len(text)
	}
	if quote := text[index]; quote == '"' || quote == '\'' {
		index++
		closing := strings.IndexByte(text[index:], quote)
		if closing < 0 {
			return text[index:], index, len(text), len(text)
		}
		end := index + closing
		return text[index:end], index, end, end + 1
	}
	start := index
	for index < len(text) && !IsSpace(text[index]) && text[index] != '>' {
		index++
	}
	return text[start:index], start, index, index
}

// CommentEnd returns the offset just past the comment starting at start, or the
// end of the text when it is unterminated.
func CommentEnd(text string, start int) int {
	if closing := strings.Index(text[start:], "-->"); closing >= 0 {
		return start + closing + len("-->")
	}
	return len(text)
}

// DeclarationEnd returns the offset just past the next `>` at or after start.
func DeclarationEnd(text string, start int) int {
	if closing := strings.IndexByte(text[start:], '>'); closing >= 0 {
		return start + closing + 1
	}
	return len(text)
}

// TagNameEnd returns the offset just past the tag name starting at start, which
// equals start when no name is there.
func TagNameEnd(text string, start int) int {
	if start >= len(text) || !IsASCIILetter(text[start]) {
		return start
	}
	index := start + 1
	for index < len(text) && IsTagNameByte(text[index]) {
		index++
	}
	return index
}

// IsTagNameByte reports whether a byte may continue a tag name.
func IsTagNameByte(character byte) bool {
	return IsASCIILetter(character) ||
		(character >= '0' && character <= '9') ||
		character == '-' || character == '_' ||
		character == ':' || character == '.'
}

// IsASCIILetter reports whether a byte is an ASCII letter.
func IsASCIILetter(character byte) bool {
	return (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
}

// IsSpace reports whether a byte is HTML whitespace.
func IsSpace(character byte) bool {
	switch character {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	}
	return false
}

// SkipSpace returns the offset of the first non-whitespace byte at or after
// index.
func SkipSpace(text string, index int) int {
	for index < len(text) && IsSpace(text[index]) {
		index++
	}
	return index
}

// Fold lowercases the ASCII letters in value and leaves every other byte alone,
// which is the fold HTML specifies for a tag or attribute name.
func Fold(value string) string {
	needsFold := false
	for index := range len(value) {
		if value[index] >= 'A' && value[index] <= 'Z' {
			needsFold = true
			break
		}
	}
	if !needsFold {
		return value
	}
	folded := []byte(value)
	for index := range folded {
		if folded[index] >= 'A' && folded[index] <= 'Z' {
			folded[index] += 'a' - 'A'
		}
	}
	return string(folded)
}
