// cspell:ignore imsi
package utils

import (
	"reflect"
	"testing"
)

// TestRegexCapturingGroups_Rejects covers patterns the host engine rejects, so
// RegexCapturingGroups has to report ok=false rather than hand back groups its
// caller would then flag inside a regex that never compiles. Expectations come
// from regexpp, the parser ESLint runs these patterns through.
func TestRegexCapturingGroups_Rejects(t *testing.T) {
	u := RegexFlags{Unicode: true}
	v := RegexFlags{UnicodeSets: true}

	cases := []struct {
		pattern string
		flags   RegexFlags
	}{
		// u/v mode has no identity-escape fallback: an escape either spells one
		// of the listed forms or the pattern is invalid.
		{`\c(a)`, u},
		{`\c(a)`, v},
		{`\q{a}(b)`, u},
		{`\q{a}(b)`, v},
		{`\a(b)`, u},
		{`\-(b)`, u},
		{`\x4(a)`, u},
		{`\xZZ(a)`, u},
		{`\u(a)`, u},
		{`\u{}(a)`, u},
		{`\uZZZZ(a)`, u},
		{`\p(a)`, u},
		{`\k(a)`, u},
		{`\01(a)`, u},

		// Backreferences have to resolve under u/v.
		{`\8(a)`, u},
		{`\2(a)`, u},
		{`\k<n>(a)`, u},
		{`\k<n>(?<m>a)`, v},

		// `\q` is a class-only string disjunction, and a v-only one at that.
		{`[\q{a}](b)`, u},
		{`[\q](b)`, u},
		{`[\q](b)`, v},

		// Class-only `\c` forms, and `\c` with nothing quantifiable after it.
		{`[\c](a)`, u},
		{`[\c0](a)`, u},

		// Modifier-group headers the engine rejects.
		{`(?ii:(a))`, RegexFlags{}},
		{`(?i-i:(a))`, RegexFlags{}},
		{`(?ims-ims:(a))`, RegexFlags{}},
		{`(?imsi:(a))`, RegexFlags{}},
		{`(?-:(a))`, RegexFlags{}},
		{`(?u:(a))`, RegexFlags{}},

		// A quantifier with no operand, or a second one on the same atom.
		{`{2}(a)`, RegexFlags{}},
		{`{2,}(a)`, RegexFlags{}},
		{`{1,3}?(a)`, RegexFlags{}},
		{`(a){1}{2}`, RegexFlags{}},
		{`*(a)`, RegexFlags{}},

		// Assertions can't be quantified. Annex B exempts lookahead; u/v doesn't.
		{`^*(a)`, RegexFlags{}},
		{`$*(a)`, RegexFlags{}},
		{`\b*(a)`, RegexFlags{}},
		{`\B*(a)`, RegexFlags{}},
		{`(?<=a)*(b)`, RegexFlags{}},
		{`(?<!a)*(b)`, RegexFlags{}},
		{`(?=a)*(b)`, u},
		{`(?!a)*(b)`, u},
		{`(?=a){2}(b)`, v},

		// Group names have to be an IdentifierName.
		{`(?<1n>a)`, RegexFlags{}},
		{`(?<n->a)`, RegexFlags{}},
		{`(?<n a>b)`, RegexFlags{}},
		{`(?<>a)`, RegexFlags{}},
		{`(?<n\>a)`, RegexFlags{}},

		// Closers with no opener are literals in Annex B only.
		{`](a)`, u},
		{`}(a)`, u},
		{`](a)`, v},

		// Unbalanced or unterminated.
		{`(a`, RegexFlags{}},
		{`(a))`, RegexFlags{}},
		{`[a(b)`, RegexFlags{}},
		{`(a)\`, RegexFlags{}},
	}

	for _, tc := range cases {
		if groups, ok := RegexCapturingGroups(tc.pattern, tc.flags); ok {
			t.Errorf("RegexCapturingGroups(%q, %+v) = %v, true; want ok=false", tc.pattern, tc.flags, groups)
		}
	}
}

// TestRegexCapturingGroups_Accepts is the other half of the same comparison:
// patterns regexpp accepts, with the groups it finds. Offsets are byte offsets,
// where regexpp counts UTF-16 code units.
func TestRegexCapturingGroups_Accepts(t *testing.T) {
	u := RegexFlags{Unicode: true}
	v := RegexFlags{UnicodeSets: true}

	cases := []struct {
		pattern string
		flags   RegexFlags
		want    []RegexCapturingGroup
	}{
		// Annex B reads `\c` with no control letter as a literal backslash, so
		// the `(` right after it still opens a group.
		{`\c(a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 2, End: 5}}},
		{`\c0(a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 3, End: 6}}},
		{`\cA(b)`, RegexFlags{}, []RegexCapturingGroup{{Start: 3, End: 6}}},
		{`\cA(b)`, u, []RegexCapturingGroup{{Start: 3, End: 6}}},
		{`[\c](a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 4, End: 7}}},
		{`[\c0](a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 5, End: 8}}},
		{`[\c]](a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 5, End: 8}}},

		// Outside u/v, an unrecognized escape is an identity escape.
		{`\q{a}(b)`, RegexFlags{}, []RegexCapturingGroup{{Start: 5, End: 8}}},
		{`\a(b)`, RegexFlags{}, []RegexCapturingGroup{{Start: 2, End: 5}}},
		{`\8(a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 2, End: 5}}},
		{`\k<n>(a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 5, End: 8}}},

		// Resolvable backreferences under u/v.
		{`(a)\1`, u, []RegexCapturingGroup{{Start: 0, End: 3}}},
		{`(?<n>a)\k<n>`, u, []RegexCapturingGroup{{Start: 0, End: 7, Name: "n"}}},
		{`(?<n>a)\k<n>`, v, []RegexCapturingGroup{{Start: 0, End: 7, Name: "n"}}},

		// The escapes u/v does spell out.
		{`\p{L}(a)`, u, []RegexCapturingGroup{{Start: 5, End: 8}}},
		{`\u{41}(a)`, u, []RegexCapturingGroup{{Start: 6, End: 9}}},
		{`\x41(a)`, u, []RegexCapturingGroup{{Start: 4, End: 7}}},
		{`\0(a)`, u, []RegexCapturingGroup{{Start: 2, End: 5}}},
		{`\/(a)`, u, []RegexCapturingGroup{{Start: 2, End: 5}}},
		{`[\q{ab}](c)`, v, []RegexCapturingGroup{{Start: 8, End: 11}}},

		// Well-formed modifier-group headers, including an empty set on either
		// side of the `-`.
		{`(?ms-i:(a))`, RegexFlags{}, []RegexCapturingGroup{{Start: 7, End: 10}}},
		{`(?i-:(a))`, RegexFlags{}, []RegexCapturingGroup{{Start: 5, End: 8}}},
		{`(?-i:(a))`, RegexFlags{}, []RegexCapturingGroup{{Start: 5, End: 8}}},
		{`(?i:(a))`, RegexFlags{}, []RegexCapturingGroup{{Start: 4, End: 7}}},

		// Annex B keeps lookahead quantifiable and stray closers literal.
		{`(?=a)*(b)`, RegexFlags{}, []RegexCapturingGroup{{Start: 6, End: 9}}},
		{`](a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 1, End: 4}}},
		{`}(a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 1, End: 4}}},
		{`a{(b)`, RegexFlags{}, []RegexCapturingGroup{{Start: 2, End: 5}}},

		// Group names, including the non-ASCII and `\u`-escaped forms. Name
		// carries the raw source text, so an escape stays unresolved in it.
		{`(?<$_>a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 0, End: 8, Name: "$_"}}},
		{`(?<é>a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 0, End: 8, Name: "é"}}},
		{`(?<n\u0041>a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 0, End: 13, Name: `n\u0041`}}},

		// Non-ASCII characters advance by their UTF-8 width.
		{`\é(a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 3, End: 6}}},
		{`\😀(a)`, RegexFlags{}, []RegexCapturingGroup{{Start: 5, End: 8}}},

		// Nesting, alternation and the non-capturing forms.
		{`(a)(?:b)(?<n>c)`, RegexFlags{}, []RegexCapturingGroup{
			{Start: 0, End: 3},
			{Start: 8, End: 15, Name: "n"},
		}},
		{`((a))`, RegexFlags{}, []RegexCapturingGroup{{Start: 0, End: 5}, {Start: 1, End: 4}}},
		{`(?<n>a)|(?<n>b)`, RegexFlags{}, []RegexCapturingGroup{
			{Start: 0, End: 7, Name: "n"},
			{Start: 8, End: 15, Name: "n"},
		}},
		{`(?=(a))(?!(b))`, RegexFlags{}, []RegexCapturingGroup{{Start: 3, End: 6}, {Start: 10, End: 13}}},
		{`(?<=(a))(?<!(b))`, RegexFlags{}, []RegexCapturingGroup{{Start: 4, End: 7}, {Start: 12, End: 15}}},
	}

	for _, tc := range cases {
		got, ok := RegexCapturingGroups(tc.pattern, tc.flags)
		if !ok {
			t.Errorf("RegexCapturingGroups(%q, %+v) = _, false; want ok=true", tc.pattern, tc.flags)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("RegexCapturingGroups(%q, %+v) = %v; want %v", tc.pattern, tc.flags, got, tc.want)
		}
	}
}
