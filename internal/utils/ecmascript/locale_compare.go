package ecmascript

//go:generate node ../../../scripts/gen-locale-collation.mjs

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// LocaleComparer compares strings with a deterministic approximation of
// String.prototype.localeCompare. A comparer owns reusable buffers, so callers
// must not use one concurrently.
//
// ECMA-402 leaves collation data implementation-defined, and the order rule
// fixes need is a stable total order rather than a copy of the host's ICU. The
// comparer therefore carries only what source identifiers use:
//
//   - Printable ASCII follows the ICU root collation, and the locale
//     rules that change printable-ASCII order (contractions such as
//     Czech "ch" or Danish "aa", reordered letters, and upper-first case
//     order). The tables are generated from Node.js's ICU and verified against
//     Intl.Collator, so this domain matches Node exactly.
//   - Control characters are ignorable, as in ICU.
//   - Other text is compared in NFD, where a combining mark is a secondary
//     difference on its base letter, so "é" sorts with "e" as in ICU. Any
//     other character sorts after printable ASCII by its simple lowercase
//     mapping, with case as a tertiary difference. This is stable but is not
//     ICU's order for compatibility characters such as "Ⅰ" or for locale
//     rules for non-ASCII letters.
//
// Locales with no table, or whose ASCII tailoring the tables cannot express,
// use the root order. The Unicode extension keys Intl.Collator reads (co, kf,
// and kn) are honored; all others are ignored.
type LocaleComparer struct {
	table       *collationTable
	upperFirst  bool
	numeric     bool
	starts      [128]bool
	left, right []collationElement
}

type collationTable struct {
	// singles holds four bytes for each of U+0020 through U+007E: primary,
	// secondary, tertiary with lowercase first, and tertiary with uppercase
	// first.
	singles      string
	contractions []collationContraction
}

type collationContraction struct {
	text    string
	element string
}

type collationLocale struct {
	table      int
	upperFirst bool
}

// collationElement holds one collation element's weights. A zero weight is
// ignorable at that level, so table ranks are stored shifted by one.
type collationElement struct {
	primary, secondary, tertiary uint32
}

const (
	// Primary weights of characters outside the tables start above every
	// table weight, which fits in a byte.
	collationOtherPrimary = 0x100
	// Secondary weights of combining marks start above every table weight.
	collationMarkSecondary = 0x100
)

// NewLocaleComparer constructs a reusable locale comparer. The empty locale
// and "auto" select the root collation rather than the host locale, so results
// do not depend on the machine running the linter.
func NewLocaleComparer(locale string) *LocaleComparer {
	comparer := &LocaleComparer{table: &collationTables[0]}
	if locale != "" && locale != "auto" {
		tag := language.Make(locale)
		keys := unicodeExtensionKeys(tag)
		base, _ := tag.Base()
		script, _ := tag.Script()
		key := base.String() + "-" + script.String()
		// Intl.Collator falls back to the default collation when the requested
		// one is not available.
		entry, ok := collationLocales[key]
		if co := keys["co"]; co != "" {
			if collation, found := collationLocales[key+"-u-co-"+co]; found {
				entry, ok = collation, true
			}
		}
		if ok {
			comparer.table = &collationTables[entry.table]
			comparer.upperFirst = entry.upperFirst
		}
		switch keys["kf"] {
		case "upper":
			comparer.upperFirst = true
		case "lower", "false":
			comparer.upperFirst = false
		}
		comparer.numeric = keys["kn"] == "true"
	}
	for _, contraction := range comparer.table.contractions {
		comparer.starts[contraction.text[0]] = true
	}
	return comparer
}

// unicodeExtensionKeys returns the keys of tag's -u- extension. A key without
// a type has the type "true", as in BCP 47.
func unicodeExtensionKeys(tag language.Tag) map[string]string {
	extension, ok := tag.Extension('u')
	if !ok {
		return nil
	}
	keys := map[string]string{}
	key := ""
	var types []string
	flush := func() {
		if key == "" {
			return
		}
		if len(types) == 0 {
			keys[key] = "true"
		} else {
			keys[key] = strings.Join(types, "-")
		}
	}
	for _, subtag := range strings.Split(extension.String(), "-")[1:] {
		if len(subtag) == 2 {
			flush()
			key, types = subtag, nil
		} else if key != "" {
			types = append(types, subtag)
		}
	}
	flush()
	return keys
}

// Compare returns -1, 0, or 1 according to the configured locale order.
func (c *LocaleComparer) Compare(left, right string) int {
	if left == right {
		return 0
	}
	c.left = c.appendElements(c.left[:0], left)
	c.right = c.appendElements(c.right[:0], right)
	for level := range 3 {
		if result := compareCollationLevel(c.left, c.right, level); result != 0 {
			return result
		}
	}
	return 0
}

// compareCollationLevel compares the non-ignorable weights of one level.
func compareCollationLevel(left, right []collationElement, level int) int {
	i, j := 0, 0
	for {
		for i < len(left) && left[i].weight(level) == 0 {
			i++
		}
		for j < len(right) && right[j].weight(level) == 0 {
			j++
		}
		switch {
		case i == len(left) && j == len(right):
			return 0
		case i == len(left):
			return -1
		case j == len(right):
			return 1
		}
		if a, b := left[i].weight(level), right[j].weight(level); a != b {
			return compareOrdered(a, b)
		}
		i++
		j++
	}
}

func (e collationElement) weight(level int) uint32 {
	switch level {
	case 0:
		return e.primary
	case 1:
		return e.secondary
	default:
		return e.tertiary
	}
}

func compareOrdered(left, right uint32) int {
	if left < right {
		return -1
	}
	return 1
}

func (c *LocaleComparer) appendElements(out []collationElement, value string) []collationElement {
	for i := range len(value) {
		if value[i] >= utf8.RuneSelf {
			value = norm.NFD.String(value)
			break
		}
	}
	for i := 0; i < len(value); {
		b := value[i]
		if b >= utf8.RuneSelf {
			r, size := utf8.DecodeRuneInString(value[i:])
			out = append(out, otherCollationElement(r, c.upperFirst))
			i += size
			continue
		}
		if b < 0x20 || b == 0x7f {
			i++
			continue
		}
		if c.numeric && isASCIIDigit(b) {
			out, i = c.appendNumber(out, value, i)
			continue
		}
		if c.starts[b] {
			if element, size := c.matchContraction(value[i:]); size > 0 {
				out = append(out, c.element(element))
				i += size
				continue
			}
		}
		out = append(out, c.single(b))
		i++
	}
	return out
}

// appendNumber collates a run of digits by numeric value, as ICU does for
// kn=true: leading zeros are ignored, and a longer number sorts later.
func (c *LocaleComparer) appendNumber(out []collationElement, value string, start int) ([]collationElement, int) {
	end := start
	for end < len(value) && isASCIIDigit(value[end]) {
		end++
	}
	digits := strings.TrimLeft(value[start:end], "0")
	if digits == "" {
		digits = "0"
	}
	out = append(out, c.single('0'), collationElement{primary: uint32(len(digits)), secondary: 1, tertiary: 1})
	for i := range len(digits) {
		out = append(out, c.single(digits[i]))
	}
	return out, end
}

func (c *LocaleComparer) matchContraction(value string) (element string, size int) {
	for _, contraction := range c.table.contractions {
		if len(contraction.text) > size && strings.HasPrefix(value, contraction.text) {
			element, size = contraction.element, len(contraction.text)
		}
	}
	return element, size
}

func (c *LocaleComparer) single(b byte) collationElement {
	offset := int(b-0x20) * 4
	return c.element(c.table.singles[offset : offset+4])
}

func (c *LocaleComparer) element(data string) collationElement {
	tertiary := data[2]
	if c.upperFirst {
		tertiary = data[3]
	}
	return collationElement{primary: uint32(data[0]) + 1, secondary: uint32(data[1]) + 1, tertiary: uint32(tertiary) + 1}
}

func otherCollationElement(r rune, upperFirst bool) collationElement {
	if unicode.Is(unicode.Mn, r) {
		return collationElement{secondary: collationMarkSecondary + uint32(r), tertiary: 1}
	}
	// Uppercase and Lowercase include non-letter characters such as Roman
	// numerals and circled letters, beyond the Lu/Ll categories.
	var tertiary uint32
	switch {
	case unicode.In(r, unicode.Lu, unicode.Other_Uppercase):
		tertiary = 2
	case unicode.Is(unicode.Lt, r):
		tertiary = 1
	}
	if upperFirst {
		tertiary = 2 - tertiary
	}
	return collationElement{primary: collationOtherPrimary + uint32(unicode.ToLower(r)), secondary: 1, tertiary: tertiary + 1}
}

func isASCIIDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
