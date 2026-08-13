package ecmascript

import (
	"slices"
	"testing"
)

// Every expectation here is what `new RegExp("^"+x+"$", flags).test(y)` answers
// in JavaScript, which is Canonicalize comparing the two characters. The two
// readings are tested together because the whole point of the flag is that they
// disagree.
func TestCanonicalize(t *testing.T) {
	tests := []struct {
		name  string
		left  rune
		right rune
		plain bool // same under `i`
		uFlag bool // same under `iu`
	}{
		{name: "ascii", left: 'a', right: 'A', plain: true, uFlag: true},
		// A final sigma, a small sigma and a capital sigma are one character
		// under either reading.
		{name: "final sigma", left: 0x03C2, right: 0x03A3, plain: true, uFlag: true},
		{name: "small sigma", left: 0x03C3, right: 0x03A3, plain: true, uFlag: true},
		{name: "both sigmas", left: 0x03C2, right: 0x03C3, plain: true, uFlag: true},
		// Without `u` a character outside ASCII never canonicalizes into
		// ASCII. Folding has no such rule, so these join up under `u`.
		{name: "kelvin sign", left: 'k', right: 0x212A, plain: false, uFlag: true},
		{name: "long s", left: 's', right: 0x017F, plain: false, uFlag: true},
		// Neither character is ASCII, so the rule above never applies and only
		// the choice of mapping decides.
		{name: "angstrom sign", left: 0x00E5, right: 0x212B, plain: false, uFlag: true},
		{name: "ohm sign", left: 0x03C9, right: 0x2126, plain: false, uFlag: true},
		// The uppercase of an eszett is two characters, so it stays put; the
		// capital eszett folds onto the small one.
		{name: "eszett", left: 0x00DF, right: 0x1E9E, plain: false, uFlag: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, mode := range []struct {
				unicodeMode bool
				want        bool
				flags       string
			}{{false, test.plain, "i"}, {true, test.uFlag, "iu"}} {
				got := Canonicalize(test.left, mode.unicodeMode) == Canonicalize(test.right, mode.unicodeMode)
				if got != mode.want {
					t.Errorf("/%s/ Canonicalize(%U)==Canonicalize(%U) is %v, want %v (got %U and %U)",
						mode.flags, test.left, test.right, got, mode.want,
						Canonicalize(test.left, mode.unicodeMode), Canonicalize(test.right, mode.unicodeMode))
				}
			}
		})
	}
}

func TestCaseEquivalents(t *testing.T) {
	tests := []struct {
		name        string
		r           rune
		unicodeMode bool
		want        []rune
	}{
		{name: "ascii letter", r: 'k', want: []rune{'K', 'k'}},
		{name: "sigma", r: 0x03A3, want: []rune{0x03A3, 0x03C2, 0x03C3}},
		// Alone in its group, so there is nothing to widen to.
		{name: "kelvin sign", r: 0x212A, want: nil},
		{name: "long s", r: 0x017F, want: nil},
		{name: "digit", r: '4', want: nil},
		// Under `u` the same two characters pull `k` and `s` into their group.
		{name: "ascii letter under u", r: 'k', unicodeMode: true, want: []rune{'K', 'k', 0x212A}},
		{name: "kelvin sign under u", r: 0x212A, unicodeMode: true, want: []rune{'K', 'k', 0x212A}},
		{name: "long s under u", r: 0x017F, unicodeMode: true, want: []rune{'S', 's', 0x017F}},
		{name: "digit under u", r: '4', unicodeMode: true, want: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CaseEquivalents(test.r, test.unicodeMode)
			if !slices.Equal(got, test.want) {
				t.Errorf("CaseEquivalents(%U, %v) = %U, want %U", test.r, test.unicodeMode, got, test.want)
			}
		})
	}
}

// TestCaseEquivalenceGroupsAgree covers the two views of the same table: every
// member of every group has to look the group back up, and canonicalize onto
// the same character as the rest of it.
func TestCaseEquivalenceGroupsAgree(t *testing.T) {
	for _, unicodeMode := range []bool{false, true} {
		groups := CaseEquivalenceGroups(unicodeMode)
		if len(groups) == 0 {
			t.Fatalf("CaseEquivalenceGroups(%v) is empty", unicodeMode)
		}
		for _, members := range groups {
			if len(members) < 2 {
				t.Errorf("group %U has nothing to widen to", members)
			}
			canonical := Canonicalize(members[0], unicodeMode)
			for _, member := range members {
				if Canonicalize(member, unicodeMode) != canonical {
					t.Errorf("group %U holds %U, which canonicalizes to %U not %U",
						members, member, Canonicalize(member, unicodeMode), canonical)
				}
				if !slices.Equal(CaseEquivalents(member, unicodeMode), members) {
					t.Errorf("CaseEquivalents(%U, %v) = %U, want the group %U",
						member, unicodeMode, CaseEquivalents(member, unicodeMode), members)
				}
			}
		}
	}
}
