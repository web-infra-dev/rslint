package ecmascript

import (
	"slices"
	"testing"
)

// Every expectation here is what `new RegExp("^"+x+"$", "i").test(y)` answers
// in JavaScript, which is Canonicalize comparing the two characters.
func TestCanonicalize(t *testing.T) {
	tests := []struct {
		name  string
		left  rune
		right rune
		same  bool
	}{
		{name: "ascii", left: 'a', right: 'A', same: true},
		// A final sigma, a small sigma and a capital sigma are one character.
		{name: "final sigma", left: 0x03C2, right: 0x03A3, same: true},
		{name: "small sigma", left: 0x03C3, right: 0x03A3, same: true},
		{name: "both sigmas", left: 0x03C2, right: 0x03C3, same: true},
		// A character outside ASCII never canonicalizes into ASCII, which is
		// where Unicode case folding gives the other answer.
		{name: "kelvin sign", left: 'k', right: 0x212A, same: false},
		{name: "long s", left: 's', right: 0x017F, same: false},
		// The uppercase of an eszett is two characters, so it stays put.
		{name: "eszett", left: 0x00DF, right: 0x1E9E, same: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Canonicalize(test.left) == Canonicalize(test.right)
			if got != test.same {
				t.Errorf("Canonicalize(%U)==Canonicalize(%U) is %v, want %v (got %U and %U)",
					test.left, test.right, got, test.same,
					Canonicalize(test.left), Canonicalize(test.right))
			}
		})
	}
}

func TestCaseEquivalents(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want []rune
	}{
		{name: "ascii letter", r: 'k', want: []rune{'K', 'k'}},
		{name: "sigma", r: 0x03A3, want: []rune{0x03A3, 0x03C2, 0x03C3}},
		// Alone in its group, so there is nothing to widen to.
		{name: "kelvin sign", r: 0x212A, want: nil},
		{name: "long s", r: 0x017F, want: nil},
		{name: "digit", r: '4', want: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CaseEquivalents(test.r)
			if !slices.Equal(got, test.want) {
				t.Errorf("CaseEquivalents(%U) = %U, want %U", test.r, got, test.want)
			}
		})
	}
}

// TestCaseEquivalenceGroupsAgree covers the two views of the same table: every
// member of every group has to look the group back up, and canonicalize onto
// the same character as the rest of it.
func TestCaseEquivalenceGroupsAgree(t *testing.T) {
	groups := CaseEquivalenceGroups()
	if len(groups) == 0 {
		t.Fatal("CaseEquivalenceGroups() is empty")
	}
	for _, members := range groups {
		if len(members) < 2 {
			t.Errorf("group %U has nothing to widen to", members)
		}
		canonical := Canonicalize(members[0])
		for _, member := range members {
			if Canonicalize(member) != canonical {
				t.Errorf("group %U holds %U, which canonicalizes to %U not %U",
					members, member, Canonicalize(member), canonical)
			}
			if !slices.Equal(CaseEquivalents(member), members) {
				t.Errorf("CaseEquivalents(%U) = %U, want the group %U", member, CaseEquivalents(member), members)
			}
		}
	}
}
