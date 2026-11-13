package regexp

import (
	"fmt"
	"testing"
)

func TestValidate(t *testing.T) {
	for _, tc := range []struct{ pattern, flags, reason string }{
		{`\u{0}*`, "", "Nothing to repeat"},
		{`[`, "", "Unterminated character class"},
		{`[[]`, "v", "Unterminated character class"},
		{`(?ii:foo)`, "", "Duplicated flag 'i'"},
		{`(?<k>a)(?<k>b)`, "", "Duplicate capture group name"},
		{`(?ims-:foo)`, "", ""},
		{`[\p{Basic_Emoji}--\q{a|bc|def}]`, "v", ""},
	} {
		t.Run(tc.pattern+tc.flags, func(t *testing.T) {
			err := Validate(tc.pattern, tc.flags)
			want := ""
			if tc.reason != "" {
				want = "Invalid regular expression: /" + tc.pattern + "/" + tc.flags + ": " + tc.reason
			}
			got := ""
			if err != nil {
				got = err.Error()
			}
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

// Expectations from @eslint-community/regexpp as resolved by ESLint v10.14.0.
// Covers Annex B, Unicode sets, names, escapes, quantifiers, and error priority.
func TestValidateDifferential(t *testing.T) {
	for _, tc := range []struct {
		pattern string
		reasons [3]string
	}{
		{"", [3]string{"", "", ""}},
		{"a", [3]string{"", "", ""}},
		{"/", [3]string{"", "", ""}},
		{"\n", [3]string{"", "", ""}},
		{"\r", [3]string{"", "", ""}},
		{"\\\n", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\\r", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\", [3]string{"\\ at end of pattern", "Invalid escape", "Invalid escape"}},
		{"[", [3]string{"Unterminated character class", "Unterminated character class", "Invalid character in character class"}},
		{"[]", [3]string{"", "", ""}},
		{"[^]", [3]string{"", "", ""}},
		{"[[]", [3]string{"", "", "Unterminated character class"}},
		{"[[a]]", [3]string{"", "Lone quantifier brackets", ""}},
		{"[a", [3]string{"Unterminated character class", "Unterminated character class", "Unterminated character class"}},
		{"[a\\", [3]string{"Invalid character in character class", "Invalid escape", "Invalid character in character class"}},
		{"[\\]", [3]string{"Unterminated character class", "Unterminated character class", "Unterminated character class"}},
		{"[a/]", [3]string{"", "", "Invalid character in character class"}},
		{"[z-a]", [3]string{"Range out of order in character class", "Range out of order in character class", "Range out of order in character class"}},
		{"[\\d-a]", [3]string{"", "Invalid character class", "Invalid character in character class"}},
		{"[a-\\d]", [3]string{"", "Invalid character class", "Invalid character in character class"}},
		{"[a&&b]", [3]string{"", "", ""}},
		{"[a--b]", [3]string{"Range out of order in character class", "Range out of order in character class", ""}},
		{"[a&&&b]", [3]string{"", "", "Invalid character in character class"}},
		{"[a!!]", [3]string{"", "", "Invalid character in character class"}},
		{"[^\\q{ab}]", [3]string{"", "Invalid escape", "Negated character class may contain strings"}},
		{"[\\q{a|bc}]", [3]string{"", "Invalid escape", ""}},
		{"[\\p{Basic_Emoji}]", [3]string{"", "Invalid property name", ""}},
		{"[\\P{Basic_Emoji}]", [3]string{"", "Invalid property name", "Invalid property name"}},
		{"\\q{a}", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\p{Letter}", [3]string{"", "", ""}},
		{"\\p{Script=Greek}", [3]string{"", "", ""}},
		{"\\p{bogus}", [3]string{"", "Invalid property name", "Invalid property name"}},
		{"\\p", [3]string{"", "Invalid property name", "Invalid property name"}},
		{"\\P{ASCII}", [3]string{"", "", ""}},
		{"\\a", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\b", [3]string{"", "", ""}},
		{"\\c", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\c1", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\x", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\x0", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\x41", [3]string{"", "", ""}},
		{"\\u", [3]string{"", "Invalid unicode escape", "Invalid unicode escape"}},
		{"\\u0", [3]string{"", "Invalid unicode escape", "Invalid unicode escape"}},
		{"\\u{0}", [3]string{"", "", ""}},
		{"\\u{0}*", [3]string{"Nothing to repeat", "", ""}},
		{"\\u{110000}", [3]string{"", "Invalid unicode escape", "Invalid unicode escape"}},
		{"\\1", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\2(a)", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"(a)\\1", [3]string{"", "", ""}},
		{"\\8", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\0", [3]string{"", "", ""}},
		{"\\00", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\01", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\k", [3]string{"", "Invalid named reference", "Invalid named reference"}},
		{"\\k<a>", [3]string{"", "Invalid named capture referenced", "Invalid named capture referenced"}},
		{"(?<a>a)\\k<a>", [3]string{"", "", ""}},
		{"(?<a>a)\\k<b>", [3]string{"Invalid named capture referenced", "Invalid named capture referenced", "Invalid named capture referenced"}},
		{"(?<a>a)(?<a>b)", [3]string{"Duplicate capture group name", "Duplicate capture group name", "Duplicate capture group name"}},
		{"(?<a>a)|(?<a>b)", [3]string{"", "", ""}},
		{"(?<1>a)", [3]string{"Invalid capture group name", "Invalid capture group name", "Invalid capture group name"}},
		{"(?<\\u{1d49c}>a)", [3]string{"", "", ""}},
		{"(?<\\ud835\\udc9c>a)", [3]string{"", "", ""}},
		{"(?<a>a)|(?<a>b)(?<a>c)", [3]string{"Duplicate capture group name", "Duplicate capture group name", "Duplicate capture group name"}},
		{"(?", [3]string{"Invalid group", "Invalid group", "Invalid group"}},
		{"(?x:a)", [3]string{"Invalid group", "Invalid group", "Invalid group"}},
		{"(?i:a)", [3]string{"", "", ""}},
		{"(?i-i:a)", [3]string{"Duplicated flag 'i'", "Duplicated flag 'i'", "Duplicated flag 'i'"}},
		{"(?-:a)", [3]string{"Invalid empty flags", "Invalid empty flags", "Invalid empty flags"}},
		{"(?i-:a)", [3]string{"", "", ""}},
		{"(?g:a)", [3]string{"Invalid group", "Invalid group", "Invalid group"}},
		{"(?i", [3]string{"Invalid group", "Invalid group", "Invalid group"}},
		{"(?<", [3]string{"Invalid capture group name", "Invalid capture group name", "Invalid capture group name"}},
		{"(", [3]string{"Unterminated group", "Unterminated group", "Unterminated group"}},
		{"a)", [3]string{"Unmatched ')'", "Unmatched ')'", "Unmatched ')'"}},
		{"[a](", [3]string{"Unterminated group", "Unterminated group", "Unterminated group"}},
		{"*", [3]string{"Nothing to repeat", "Nothing to repeat", "Nothing to repeat"}},
		{"^*", [3]string{"Nothing to repeat", "Nothing to repeat", "Nothing to repeat"}},
		{"(?=a)*", [3]string{"", "Nothing to repeat", "Nothing to repeat"}},
		{"(?<=a)*", [3]string{"Nothing to repeat", "Nothing to repeat", "Nothing to repeat"}},
		{"a**", [3]string{"Nothing to repeat", "Nothing to repeat", "Nothing to repeat"}},
		{"a{1,0}", [3]string{"numbers out of order in {} quantifier", "numbers out of order in {} quantifier", "numbers out of order in {} quantifier"}},
		{"a{", [3]string{"", "Incomplete quantifier", "Incomplete quantifier"}},
		{"a{1", [3]string{"", "Incomplete quantifier", "Incomplete quantifier"}},
		{"a{1,", [3]string{"", "Incomplete quantifier", "Incomplete quantifier"}},
		{"a{,1}", [3]string{"", "Incomplete quantifier", "Incomplete quantifier"}},
		{"{", [3]string{"", "Lone quantifier brackets", "Lone quantifier brackets"}},
		{"}", [3]string{"", "Lone quantifier brackets", "Lone quantifier brackets"}},
		{"a{2}*", [3]string{"Nothing to repeat", "Nothing to repeat", "Nothing to repeat"}},
		{"a{999999999999999999,2}", [3]string{"numbers out of order in {} quantifier", "numbers out of order in {} quantifier", "numbers out of order in {} quantifier"}},
		{"(?:a)[b-c]", [3]string{"", "", ""}},
		{"[[a]/]", [3]string{"", "Lone quantifier brackets", "Invalid character in character class"}},
		{"[[a]--[b]]", [3]string{"", "Lone quantifier brackets", ""}},
		{"[a&&[b]]", [3]string{"", "Lone quantifier brackets", ""}},
		{"[^]", [3]string{"", "", ""}},
		{"[^", [3]string{"Unterminated character class", "Unterminated character class", "Invalid character in character class"}},
		{"[[", [3]string{"Unterminated character class", "Unterminated character class", "Invalid character in character class"}},
		{"[a--]", [3]string{"Range out of order in character class", "Range out of order in character class", "Invalid character in character class"}},
		{"[a&&]", [3]string{"", "", "Invalid character in character class"}},
		{"[a--b&&c]", [3]string{"Range out of order in character class", "Range out of order in character class", "Invalid character in character class"}},
		{"[a--\\q{bc}]", [3]string{"Range out of order in character class", "Range out of order in character class", ""}},
		{"[\\q{}]", [3]string{"", "Invalid escape", ""}},
		{"[\\q{a}--b]", [3]string{"Range out of order in character class", "Invalid escape", ""}},
		{"(\\", [3]string{"Unterminated group", "Invalid escape", "Invalid escape"}},
		{"(?i:(", [3]string{"Unterminated group", "Unterminated group", "Unterminated group"}},
		{"(?<a>", [3]string{"Unterminated group", "Unterminated group", "Unterminated group"}},
		{"(?<a", [3]string{"Invalid capture group name", "Invalid capture group name", "Invalid capture group name"}},
		{"(?<1>a)", [3]string{"Invalid capture group name", "Invalid capture group name", "Invalid capture group name"}},
		{"(?<a>a)\\k", [3]string{"Invalid named reference", "Invalid named reference", "Invalid named reference"}},
		{"(?<a>a)\\k<>", [3]string{"Invalid capture group name", "Invalid capture group name", "Invalid capture group name"}},
		{"(?<a>a)\\k<1>", [3]string{"Invalid capture group name", "Invalid capture group name", "Invalid capture group name"}},
		{"(?<a>a)\\k<b>)", [3]string{"Unmatched ')'", "Unmatched ')'", "Unmatched ')'"}},
		{"[\\1-\\2]", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"[\\8-a]", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"(a)[\\1]", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"(a)\\2", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"(a)\\11", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"(a)\\111", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"(a)\\777", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"[\\c1-a]", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"[\\a-z]", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"[\\x-a]", [3]string{"Range out of order in character class", "Invalid escape", "Invalid escape"}},
		{"[\\u{1}-a]", [3]string{"Range out of order in character class", "", ""}},
		{"[\\\n-a]", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"a\\\n*", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\p{ASCII_Hex_Digit}", [3]string{"", "", ""}},
		{"\\p{scx=Latin}", [3]string{"", "", ""}},
		{"\\p{General_Category=Letter}", [3]string{"", "", ""}},
		{"\\p{Script_Extensions=Latin}", [3]string{"", "", ""}},
		{"\\P{RGI_Emoji}", [3]string{"", "Invalid property name", "Invalid property name"}},
		{"[\\P{RGI_Emoji}]", [3]string{"", "Invalid property name", "Invalid property name"}},
		{"(?<𝒜>.)", [3]string{"", "", ""}},
		{"(?<a>a)|(?<a>b)", [3]string{"", "", ""}},
		{"(?:(?<a>x))(?<a>y)", [3]string{"Duplicate capture group name", "Duplicate capture group name", "Duplicate capture group name"}},
		{"(?i-m:a)", [3]string{"", "", ""}},
		{"(?z:a)", [3]string{"Invalid group", "Invalid group", "Invalid group"}},
		{"(?:)*", [3]string{"", "", ""}},
		{"(?<=a){2}", [3]string{"Nothing to repeat", "Nothing to repeat", "Nothing to repeat"}},
		{"(?=a){2}", [3]string{"", "Nothing to repeat", "Nothing to repeat"}},
		{"a{9999999999999999999999999999}", [3]string{"", "", ""}},
		{"a{2,9999999999999999999999999999}", [3]string{"", "", ""}},
		{"a*", [3]string{"", "", ""}},
		{"a{0}", [3]string{"", "", ""}},
		{"a{1,0}", [3]string{"numbers out of order in {} quantifier", "numbers out of order in {} quantifier", "numbers out of order in {} quantifier"}},
		{"a[", [3]string{"Unterminated character class", "Unterminated character class", "Invalid character in character class"}},
		{"a\\", [3]string{"\\ at end of pattern", "Invalid escape", "Invalid escape"}},
		{"a)", [3]string{"Unmatched ')'", "Unmatched ')'", "Unmatched ')'"}},
		{"\\u{0}*", [3]string{"Nothing to repeat", "", ""}},
		{"\\u{0}{0}", [3]string{"Nothing to repeat", "", ""}},
		{"\\u{0}{1,0}", [3]string{"Nothing to repeat", "numbers out of order in {} quantifier", "numbers out of order in {} quantifier"}},
		{"\\u{0}[", [3]string{"Unterminated character class", "Unterminated character class", "Invalid character in character class"}},
		{"\\u{0}\\", [3]string{"\\ at end of pattern", "Invalid escape", "Invalid escape"}},
		{"\\u{0})", [3]string{"Unmatched ')'", "Unmatched ')'", "Unmatched ')'"}},
		{"\\c*", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\c{0}", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\c{1,0}", [3]string{"numbers out of order in {} quantifier", "Invalid escape", "Invalid escape"}},
		{"\\c[", [3]string{"Unterminated character class", "Invalid escape", "Invalid escape"}},
		{"\\c\\", [3]string{"\\ at end of pattern", "Invalid escape", "Invalid escape"}},
		{"\\c)", [3]string{"Unmatched ')'", "Invalid escape", "Invalid escape"}},
		{"\\p{Letter}*", [3]string{"", "", ""}},
		{"\\p{Letter}{0}", [3]string{"", "", ""}},
		{"\\p{Letter}{1,0}", [3]string{"numbers out of order in {} quantifier", "numbers out of order in {} quantifier", "numbers out of order in {} quantifier"}},
		{"\\p{Letter}[", [3]string{"Unterminated character class", "Unterminated character class", "Invalid character in character class"}},
		{"\\p{Letter}\\", [3]string{"\\ at end of pattern", "Invalid escape", "Invalid escape"}},
		{"\\p{Letter})", [3]string{"Unmatched ')'", "Unmatched ')'", "Unmatched ')'"}},
		{"\\1*", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\1{0}", [3]string{"", "Invalid escape", "Invalid escape"}},
		{"\\1{1,0}", [3]string{"numbers out of order in {} quantifier", "Invalid escape", "Invalid escape"}},
		{"\\1[", [3]string{"Unterminated character class", "Invalid escape", "Invalid escape"}},
		{"\\1\\", [3]string{"\\ at end of pattern", "Invalid escape", "Invalid escape"}},
		{"\\1)", [3]string{"Unmatched ')'", "Invalid escape", "Invalid escape"}},
		{"(?=a)*", [3]string{"", "Nothing to repeat", "Nothing to repeat"}},
		{"(?=a){0}", [3]string{"", "Nothing to repeat", "Nothing to repeat"}},
		{"(?=a){1,0}", [3]string{"numbers out of order in {} quantifier", "Nothing to repeat", "Nothing to repeat"}},
		{"(?=a)[", [3]string{"Unterminated character class", "Unterminated character class", "Invalid character in character class"}},
		{"(?=a)\\", [3]string{"\\ at end of pattern", "Invalid escape", "Invalid escape"}},
		{"(?=a))", [3]string{"Unmatched ')'", "Unmatched ')'", "Unmatched ')'"}},
		{"(?i:a)*", [3]string{"", "", ""}},
		{"(?i:a){0}", [3]string{"", "", ""}},
		{"(?i:a){1,0}", [3]string{"numbers out of order in {} quantifier", "numbers out of order in {} quantifier", "numbers out of order in {} quantifier"}},
		{"(?i:a)[", [3]string{"Unterminated character class", "Unterminated character class", "Invalid character in character class"}},
		{"(?i:a)\\", [3]string{"\\ at end of pattern", "Invalid escape", "Invalid escape"}},
		{"(?i:a))", [3]string{"Unmatched ')'", "Unmatched ')'", "Unmatched ')'"}},
		{"[a--b]*", [3]string{"Range out of order in character class", "Range out of order in character class", ""}},
		{"[a--b]{0}", [3]string{"Range out of order in character class", "Range out of order in character class", ""}},
		{"[a--b]{1,0}", [3]string{"Range out of order in character class", "Range out of order in character class", "numbers out of order in {} quantifier"}},
		{"[a--b][", [3]string{"Range out of order in character class", "Range out of order in character class", "Invalid character in character class"}},
		{"[a--b]\\", [3]string{"Range out of order in character class", "Range out of order in character class", "Invalid escape"}},
		{"[a--b])", [3]string{"Range out of order in character class", "Range out of order in character class", "Unmatched ')'"}},
	} {
		for i, flags := range []string{"", "u", "v"} {
			t.Run(fmt.Sprintf("%q/%s", tc.pattern, flags), func(t *testing.T) {
				err := Validate(tc.pattern, flags)
				want := ""
				if tc.reasons[i] != "" {
					want = fmt.Sprintf("Invalid regular expression: /%s/%s: %s", tc.pattern, flags, tc.reasons[i])
				}
				got := ""
				if err != nil {
					got = err.Error()
				}
				if got != want {
					t.Errorf("got %q, want %q", got, want)
				}
			})
		}
	}
}
