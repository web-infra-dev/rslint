package ecmascript

import (
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/utils/unicode17"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// LocaleComparer compares strings with the locale-sensitive semantics used by
// String.prototype.localeCompare. A comparer owns mutable collation iterators,
// so callers must not use one concurrently. Reusing it for one rule run avoids
// rebuilding the collation tables for every comparison.
//
// ECMA-402 deliberately leaves the locale data implementation-defined. This
// implementation uses the repository's bundled x/text data, filters Unicode
// extension keys to the ones Intl.Collator reads (co, kf, and kn), and carries
// the modern Nordic tailoring that x/text's old CLDR tables lack. The ASCII
// correction and tested case-first ordering match the Node 24 /
// ICU 78 oracle; other locale-data differences remain deterministic rather
// than depending on the host operating system.
type LocaleComparer struct {
	collator       *collate.Collator
	foldedCollator *collate.Collator
	upperFirst     bool
}

// NewLocaleComparer constructs a reusable locale comparer. The empty locale
// and "auto" select the deterministic root collation this port uses when the
// JavaScript call supplies no explicit locale.
func NewLocaleComparer(locale string) *LocaleComparer {
	tag := language.Und
	if locale != "" && locale != "auto" {
		tag = language.Make(locale)
	}

	base, _ := tag.Base()
	baseName := base.String()
	caseFirst := tag.TypeForKey("kf")
	cleanTag := intlCollationTag(tag)

	// x/text is pinned to CLDR 23, whose Norwegian tables put "aa" with
	// ordinary a instead of treating it like the final letter å. Danish uses
	// the same Nordic contraction and end-of-alphabet order for the ASCII
	// identifier domain, so it supplies the missing modern no/nb/nn tailoring.
	// x/text advertises no alternative collation for these locales, so any co
	// request is unsupported by this implementation and falls back to default,
	// as Intl.Collator requires.
	if baseName == "no" || baseName == "nb" || baseName == "nn" {
		cleanTag = language.Danish
		if numeric := tag.TypeForKey("kn"); numeric != "" {
			if numeric == "true" || numeric == "false" {
				cleanTag, _ = cleanTag.SetTypeForKey("kn", numeric)
			}
		}
	}

	upperFirst := caseFirst == "upper"
	if caseFirst != "upper" && caseFirst != "lower" && caseFirst != "false" {
		// These are the CLDR locales whose default sorting collation puts
		// uppercase first. Keep this locale data here rather than in callers.
		upperFirst = baseName == "da" || baseName == "mt"
	}
	comparer := &LocaleComparer{collator: collate.New(cleanTag), upperFirst: upperFirst}
	if upperFirst {
		// x/text has no kf implementation. Comparing without tertiary case
		// weights first lets us reverse only a true case tie, without undoing
		// locale contractions such as Danish "aa".
		comparer.foldedCollator = collate.New(cleanTag, collate.IgnoreCase)
	}
	return comparer
}

// Compare returns -1, 0, or 1 according to the configured locale order.
func (c *LocaleComparer) Compare(left, right string) int {
	if c.upperFirst {
		if result := c.foldedCollator.CompareString(left, right); result != 0 {
			return result
		}
		if result := compareUpperFirstCase(left, right); result != 0 {
			return result
		}
	}
	return c.collator.CompareString(left, right)
}

func compareUpperFirstCase(left, right string) int {
	leftPos, rightPos := 0, 0
	for {
		leftWeight, nextLeft, leftOK := nextCaseWeight(left, leftPos)
		rightWeight, nextRight, rightOK := nextCaseWeight(right, rightPos)
		if !leftOK || !rightOK {
			switch {
			case leftOK:
				return 1
			case rightOK:
				return -1
			default:
				return 0
			}
		}
		if leftWeight < rightWeight {
			return -1
		}
		if leftWeight > rightWeight {
			return 1
		}
		leftPos, rightPos = nextLeft, nextRight
	}
}

func nextCaseWeight(value string, pos int) (weight, next int, ok bool) {
	for pos < len(value) {
		r, size := utf8.DecodeRuneInString(value[pos:])
		pos += size
		switch {
		case unicode17.IsUppercase(r), unicode17.IsTitle(r):
			return 0, pos, true
		case unicode17.IsLowercase(r):
			return 1, pos, true
		}
	}
	return 0, pos, false
}

// intlCollationTag removes Unicode extension keys that Intl.Collator does not
// read. x/text otherwise gives keys such as ks, kc, kb, and ka meaning even
// though ECMA-402 only admits co, kf, and kn for collation.
func intlCollationTag(tag language.Tag) language.Tag {
	clean, _ := language.Compose(tag, []language.Extension{})
	for _, key := range []string{"co", "kn"} {
		if value := tag.TypeForKey(key); value != "" {
			if key == "co" && (value == "default" || value == "standard") {
				continue
			}
			if key == "kn" && value != "true" && value != "false" {
				continue
			}
			clean, _ = clean.SetTypeForKey(key, value)
		}
	}
	return clean
}
