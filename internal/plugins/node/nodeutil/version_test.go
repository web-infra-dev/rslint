package nodeutil

import "testing"

// Replacement-range decisions compared with npm semver 7.8.5.
func TestNodeVersionReplacementRanges(t *testing.T) {
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
			version, ok := parseNodeVersion(tc.raw)
			if ok != tc.valid {
				t.Fatalf("valid = %v, want %v", ok, tc.valid)
			}
			if !ok {
				return
			}
			for i, since := range []string{"0.0.1", "5.10.0", "5.12.0"} {
				if got := version.Supports(since); got != tc.supports[i] {
					t.Errorf("Supports(%s) = %v, want %v", since, got, tc.supports[i])
				}
			}
		})
	}
}
