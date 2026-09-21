package npmsemver

import "testing"

func TestRawRange(t *testing.T) {
	for _, test := range []struct{ input, raw string }{
		{"", ""},
		{"^16 || >=20", "^16 || >=20"},
		{" \ufeff>=\u00a0 16.0.0\n\t< 20 ", ">= 16.0.0 < 20"},
		{"v16.0.0+build", "v16.0.0+build"},
	} {
		version, ok := Parse(test.input)
		if !ok || version.Raw() != test.raw {
			t.Errorf("Parse(%q).Raw() = %q, valid = %v; want %q", test.input, version.Raw(), ok, test.raw)
		}
	}
}

// Replacement-range decisions compared with npm semver 7.8.5.
func TestReplacementRanges(t *testing.T) {
	for _, tc := range []struct {
		raw      string
		valid    bool
		supports []bool
	}{
		{"<0.0.0", true, []bool{true, true, true}},
		{"<0.0.0-0", true, []bool{true, true, true}},
		{"<0.0.0-beta", true, []bool{true, true, true}},
		{"<=0.0.0", true, []bool{false, false, false}},
		{">0.0.0", true, []bool{false, false, false}},
		{"^0", true, []bool{false, false, false}},
		{"^0.0", true, []bool{false, false, false}},
		{"^0.0.0", true, []bool{false, false, false}},
		{"0.0.x", true, []bool{false, false, false}},
		{"5.9 - 5.10", true, []bool{true, false, false}},
		{"5.10 - 6", true, []bool{true, true, false}},
		{"v5.10", true, []bool{true, true, false}},
		{"v*", true, []bool{false, false, false}},
		{"=vx", true, []bool{false, false, false}},
		{">=v5.10", true, []bool{true, true, false}},
		{"~> 5.10", true, []bool{true, true, false}},
		{"^ 5.10", true, []bool{true, true, false}},
		{">=5.10.0-beta.1 <6.0.0", true, []bool{true, false, false}},
		{"5.10.0-beta.1", true, []bool{true, true, true}},
		{"5.9.0-beta.1", true, []bool{true, true, true}},
		{"5.10.0-0", true, []bool{true, true, true}},
		{">=5.10.0-0", true, []bool{true, false, false}},
		{"5.10.0 ||", true, []bool{false, false, false}},
		{"|| 5.10.0", true, []bool{false, false, false}},
		{"5.10.0 || *", true, []bool{false, false, false}},
		{"5.09", false, []bool{}},
		{"5.10.00", false, []bool{}},
		{">=5.10.0 <5.10.0", true, []bool{true, true, true}},
		{">5.10.0 <=5.10.0", true, []bool{true, true, true}},
		{"5.10.0+meta", true, []bool{true, true, false}},
		{"  >=5.10.0\t<6", true, []bool{true, true, false}},
		{">= 5.10.0", true, []bool{true, true, false}},
		{"5.10 || ^6.0.0-0", true, []bool{true, true, false}},
		{"\u00a0>=\ufeff5.10.0\u2028<6", true, []bool{true, true, false}},
		{"5.10.0 - 6.0.0 >=5.11.0", false, []bool{}},
		{">=5.10.0-1beta", true, []bool{true, false, false}},
		{"5.10.0-1beta", true, []bool{true, true, true}},
		{">=5.10.0-01beta", true, []bool{true, false, false}},
		{">=5.10.0-1-rc.2+build.1", true, []bool{true, false, false}},
		{">=5.10.0-alpha.1beta", true, []bool{true, false, false}},
		{">=5.10.0-1beta <5.10.0-2beta", true, []bool{true, false, false}},
		{">=5.10.0-2beta <5.10.0-1beta", true, []bool{true, true, true}},
		{">=5.10.0-10beta <5.10.0-2beta", true, []bool{true, false, false}},
		{">=5.10.0-2beta <5.10.0-10beta", true, []bool{true, true, true}},
		{">=5.10.0-1beta <5.10.0-alpha", true, []bool{true, false, false}},
		{">=5.10.0-9 <5.10.0-1beta", true, []bool{true, false, false}},
		{">=5.10.0-1beta <5.10.0-9", true, []bool{true, true, true}},
		{"5.10.0-1beta - 5.12.0-2beta", true, []bool{true, false, false}},
		{">=5.10.0-1beta || >=5.12.0", true, []bool{true, false, false}},
		{"5.x.1", false, []bool{}},
		{"x.1", false, []bool{}},
		{"x.x.1", false, []bool{}},
		{"5.X.0", false, []bool{}},
		{"5.*.1", false, []bool{}},
		{"^5.x.1", true, []bool{true, false, false}},
		{"~5.x.1", true, []bool{true, false, false}},
		{">=5.x.1", false, []bool{}},
		{"5.x.1 - 6", true, []bool{true, false, false}},
		{"5 - 6.x.1", true, []bool{true, false, false}},
		{"5.x.x", true, []bool{true, false, false}},
		{"*.x.x", true, []bool{false, false, false}},
		{"5.x.x-01", false, []bool{}},
		{"5.10.0-01", false, []bool{}},
		{"5.10.0-alpha..beta", false, []bool{}},
		{"5.10.0-1beta+build.01", true, []bool{true, true, true}},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			version, ok := Parse(tc.raw)
			if ok != tc.valid {
				t.Fatalf("valid = %v, want %v", ok, tc.valid)
			}
			if !ok {
				return
			}
			for i, since := range []string{"0.0.1", "5.10.0", "5.12.0"} {
				if got := version.IsAtLeast(since); got != tc.supports[i] {
					t.Errorf("IsAtLeast(%s) = %v, want %v", since, got, tc.supports[i])
				}
			}
		})
	}
}

// Expected containment from node-semver 7.8.5, resolved by eslint-plugin-n v18.3.0.
func TestNodeProtocolSubsetRanges(t *testing.T) {
	for _, test := range []struct {
		text     string
		esm, cjs bool
	}{
		// Empty alternatives must not make union order change the answer.
		// npm subset() incorrectly rejects the first ordering.
		{">=16 || >20 <16", true, true},
		{">20 <16 || >=16", true, true},
		{">20 <16 || >=16 || >20 <16", true, true},
		{">=10 || >20 <16", false, false},
		{">20 <16 || >=10", false, false},
		{"16.0.0 >=16.0.0-rc.1", true, true},
		{"", false, false},
		{"*", false, false},
		{" ", false, false},
		{"16", true, true},
		{"16.x", true, true},
		{">= 16", true, true},
		{"> 15", true, true},
		{"^16", true, true},
		{"~16", true, true},
		{"~> 16", true, true},
		{"v16.0.0", true, true},
		{"= v16.0.0", true, true},
		{"12.20.0", true, false},
		{"12.19.1", false, false},
		{"^12.20.0", true, false},
		{"~12.20", true, false},
		{"^12.20.0 || >=14.13.1", true, false},
		{"12.20.0 - 12.22", true, false},
		{"^12.20.0 || ^14.18.0 || >=16", true, false},
		{"12 || 14 || 16", false, false},
		{"^12.20.0 || 13", false, false},
		{">=12.20.0", false, false},
		{"13.14.0", false, false},
		{"14.13.0", false, false},
		{"14.13.1", true, false},
		{"14.17.6", true, false},
		{"14.18.0", true, true},
		{"15.14.0", true, false},
		{"16.0.0", true, true},
		{"^14.13.1", true, false},
		{"^14.18.0", true, true},
		{"^14.18.0 || >=16.0.0", true, true},
		{">=14.18.0", true, false},
		{">=14.18.0 <15", true, true},
		{">=16.0.0 <16.0.0", true, true},
		{"<0.0.0-0", false, false},
		{"<0.0.0", false, false},
		{">20 <16", true, true},
		{"16.0.0 17.0.0", true, true},
		{">=16.0.0 <=16.0.0", true, true},
		{">16.0.0 <16.0.1", true, true},
		{"^16.0.0-rc.1", false, false},
		{">=16.0.0 <17.0.0-rc.1", false, false},
		{">=16.0.0 <17.0.0-0", true, true},
		{"16.0.0-rc.1", false, false},
		{"16.0.0-rc.1 <17", true, true},
		{"16.0.0-rc.1 >14", true, true},
		{"16.0.0-rc.1 >=16", true, true},
		{"16.0.0-rc.1 >16.0.0-beta", false, false},
		{">=16.0.0-rc.1 <=16.0.0-rc.1", false, false},
		{">=16.0.0 || <0.0.0-0", true, true},
		{"<0.0.0-0 || >=16.0.0", true, true},
		{"<0.0.0-0 || <0.0.0-0", false, false},
		{">=16.0.0 || <0.0.0", false, false},
		{"<0.0.0 || >=16.0.0", false, false},
		{"16.0.0+build", true, true},
		{">=16.0.0-beta >=16.0.0", true, true},
		{"\ufeff>=\u00a016\u2028", true, true},
		{">=16 ||", false, false},
		{"|| >=16", false, false},
		{"<*", false, false},
		{"^0.0.0", false, false},
		{"~0.1", false, false},
	} {
		t.Run(test.text, func(t *testing.T) {
			r, ok := Parse(test.text)
			if !ok {
				t.Fatal("valid range rejected")
			}
			if got := r.IsSubsetOf("^12.20.0 || >=14.13.1"); got != test.esm {
				t.Errorf("ESM: got %v, want %v", got, test.esm)
			}
			if got := r.IsSubsetOf("^14.18.0 || >=16.0.0"); got != test.cjs {
				t.Errorf("CJS: got %v, want %v", got, test.cjs)
			}
		})
	}
}

func TestInvalidRanges(t *testing.T) {
	for _, text := range []string{"invalid", ">=", "16 | 20", "16, 20", "\u008510", "10 - 12\u0085", "<=04294967296", "<=4294967296.invalid", ">=4294967296 || invalid", ">>4294967296", "<=9007199254740992"} {
		if _, ok := Parse(text); ok {
			t.Errorf("accepted invalid range %q", text)
		}
	}
}

// Decisions independently checked against npm semver 7.8.5.
func TestExtremeRanges(t *testing.T) {
	for _, test := range []struct {
		text                      string
		valid, esm, cjs, supports bool
	}{
		{"<=4294967296", true, false, false, false},
		{">=12 <4294967296", true, false, false, true},
		{">4294967295", true, true, true, true},
		{">=16 <4294967296", true, true, true, true},
		{"^0.0.4294967295", true, false, false, false},
		{"~12.4294967295.0", true, true, false, true},
		{"12 - 4294967295", true, false, false, true},
		{"^16 || <=4294967296", true, false, false, false},
		{"<=9007199254740991", false, false, false, false},
		{"9007199254740991.0.0", true, true, true, true},
		{">=9007199254740991.0.0", true, true, true, true},
		{"16.0.0+4294967296", true, true, true, true},
		{"16.0.0-4294967296", true, false, false, true},
		{"14.13.4294967296", true, true, false, true},
		{"14.18.4294967295", true, true, true, true},
		{"2147483647 - 2147483648", true, true, true, true},
		{">2147483647 <2147483649", true, true, true, true},
		{">=4294967296.0.0 <4294967296.0.0", true, true, true, true},
		{"4294967296.0.0 >=4294967296.0.0", true, true, true, true},
		{"4294967296.0.0 <4294967296.0.0", true, true, true, true},
		{"^4294967295.1.2 || 12.20.0", true, true, false, true},
		{"^12.x.9007199254740992", true, false, false, true},
		{"~12.x.9999999999999999999999999999", true, false, false, true},
		{"12.x.9007199254740992 - 16.0.0", true, false, false, true},
		{"16.0.0 - 16.x.9007199254740992", true, true, true, true},
		{"^*.9007199254740992.0", true, false, false, false},
		{"~*.9007199254740992.0", true, false, false, false},
		{"^0.4294967295.4294967296", true, false, false, false},
		{"~1.2147483647.4294967296", true, false, false, false},
		{"^4294967296.4294967295.2147483648", true, true, true, true},
		{">=4294967296.4294967295.2147483648", true, true, true, true},
		{"4294967295.4294967295.4294967295 - 4294967296.4294967296.4294967296", true, true, true, true},
	} {
		t.Run(test.text, func(t *testing.T) {
			version, ok := Parse(test.text)
			if ok != test.valid {
				t.Fatalf("valid = %v, want %v", ok, test.valid)
			}
			if !ok {
				return
			}
			if got := version.IsSubsetOf("^12.20.0 || >=14.13.1"); got != test.esm {
				t.Errorf("ESM = %v, want %v", got, test.esm)
			}
			if got := version.IsSubsetOf("^14.18.0 || >=16.0.0"); got != test.cjs {
				t.Errorf("CJS = %v, want %v", got, test.cjs)
			}
			if got := version.IsAtLeast("5.10.0"); got != test.supports {
				t.Errorf("replacement support = %v, want %v", got, test.supports)
			}
		})
	}
}

// npm 7.8.5 expands implicit upper bounds with -0. Keep an explicit
// prerelease lower bound in another term from changing wildcard expansion.
func TestGeneratedBounds(t *testing.T) {
	for _, test := range []struct {
		text, domain string
		want         bool
	}{
		{"^1.2.3", ">=1.2.3-beta <2", true},
		{"^1.2.3-beta.1", ">=1.2.3-beta <2", true},
		{"~1.2.3", ">=1.2.3-beta <1.3", true},
		{"1.2.3 - 1.2", ">=1.2.3-beta <1.3", true},
		{"1.2.3 - 2", "<3.0.0-0", true},
		{"^1.2.3 >2.0.0-alpha", ">=3", true},
		{"~1.2.3 >1.3.0-alpha", ">=3", true},
		{">=2.0.0-alpha <2.0.0", "<2.0.0-0", false},
		{"1.x >=1.0.0-0", ">=1.0.0", true},
		// Preserve existing corrections to npm subset(): stable zero-major
		// ranges fit *, and a singleton prerelease fits a range admitting it.
		{"^0", "*", true},
		{"~0.0", "*", true},
		{"1.2.3-beta", ">=1.2.3-beta <2", true},
	} {
		t.Run(test.text+" in "+test.domain, func(t *testing.T) {
			r, ok := Parse(test.text)
			if !ok {
				t.Fatal("valid range rejected")
			}
			if got := r.IsSubsetOf(test.domain); got != test.want {
				t.Fatalf("IsSubsetOf(%q) = %v, want %v", test.domain, got, test.want)
			}
		})
	}
}
