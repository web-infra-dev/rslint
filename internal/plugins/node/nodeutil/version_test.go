package nodeutil

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

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

func TestConfiguredNodeVersion(t *testing.T) {
	for _, test := range []struct {
		name, metadata, expected string
		options, settings        map[string]any
	}{
		{name: "fallback", metadata: `{}`, expected: ">=16.0.0"},
		{name: "extreme option keeps precedence", metadata: `{"engines":{"node":"^20"}}`, options: map[string]any{"version": "<=4294967296"}, settings: map[string]any{"node": map[string]any{"version": "^20"}}, expected: "<=4294967296"},
		{name: "extreme setting keeps precedence", metadata: `{"engines":{"node":"^20"}}`, settings: map[string]any{"node": map[string]any{"version": "~12.4294967295.0"}}, expected: "~12.4294967295.0"},
		{name: "extreme engine keeps precedence", metadata: `{"engines":{"node":"<=4294967296"},"devEngines":{"runtime":{"name":"node","version":"^20"}}}`, expected: "<=4294967296"},
		{name: "extreme dev engine keeps precedence", metadata: `{"devEngines":{"runtime":{"name":"node","version":"^0.0.4294967295"}}}`, expected: "^0.0.4294967295"},
		{name: "engines", metadata: `{"engines":{"node":"^12.20.0"}}`, expected: "^12.20.0"},
		{name: "empty engine range", metadata: `{"engines":{"node":""}}`, expected: "*"},
		{name: "dev engine", metadata: `{"devEngines":{"runtime":{"name":"node","version":"^14.18.0"}}}`, expected: "^14.18.0"},
		{name: "dev engine array", metadata: `{"devEngines":{"runtime":[null,42,{"name":"bun","version":"1"},{"name":"node"},{"name":"node","version":"^20"}]}}`, expected: "^20"},
		{name: "engine precedence", metadata: `{"engines":{"node":"^12.20.0"},"devEngines":{"runtime":{"name":"node","version":"^20"}}}`, expected: "^12.20.0"},
		{name: "invalid engine", metadata: `{"engines":{"node":"invalid"},"devEngines":{"runtime":{"name":"node","version":"^20"}}}`, expected: "^20"},
		{name: "non-string engine", metadata: `{"engines":{"node":12},"devEngines":{"runtime":{"name":"node","version":"^20"}}}`, expected: "^20"},
		{name: "first node runtime", metadata: `{"devEngines":{"runtime":[{"name":"node","version":"invalid"},{"name":"node","version":"^12"}]}}`, expected: ">=16.0.0"},
		{name: "invalid metadata shapes", metadata: `{"engines":12,"devEngines":[]}`, expected: ">=16.0.0"},
		{name: "non-node runtime", metadata: `{"devEngines":{"runtime":{"name":"bun","version":"1"}}}`, expected: ">=16.0.0"},
		{name: "rule option", metadata: `{"engines":{"node":"^12"}}`, options: map[string]any{"version": "^20"}, settings: map[string]any{"node": map[string]any{"version": "^14"}}, expected: "^20"},
		{name: "node setting", metadata: `{"engines":{"node":"^12"}}`, settings: map[string]any{"node": map[string]any{"version": "^20"}}, expected: "^20"},
		{name: "n setting precedence", metadata: `{}`, settings: map[string]any{"n": map[string]any{"version": "^12"}, "node": map[string]any{"version": "^20"}}, expected: "^12"},
		{name: "invalid settings fall through", metadata: `{"engines":{"node":"^12"}}`, options: map[string]any{"version": "invalid"}, settings: map[string]any{"n": map[string]any{"version": "invalid"}, "node": map[string]any{"version": ""}}, expected: "^12"},
		{name: "numeric setting", metadata: `{}`, settings: map[string]any{"node": map[string]any{"version": float64(12)}}, expected: "12"},
		{name: "falsy setting", metadata: `{"engines":{"node":"^12"}}`, settings: map[string]any{"n": map[string]any{"version": false}, "node": map[string]any{"version": float64(0)}}, expected: "^12"},
		{name: "array setting", metadata: `{}`, settings: map[string]any{"node": map[string]any{"version": []any{}}}, expected: "*"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := packageRoot(t)
			root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
				tspath.ResolvePath(root.Dir, "package.json"): test.metadata,
			})
			program, file, err := rule_tester.NewProgramHelper(root).CreateTestProgram(`import "fs";`, "nested/input.js", "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			ctx := (rule.RuleContext{SourceFile: file, Settings: test.settings}).WithProgram(lintprogram.NewFromCompiler(program))
			got := ConfiguredNodeVersion(ctx, test.options)
			want, ok := parseNodeVersion(test.expected)
			if !ok {
				t.Fatal("invalid expected range")
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("configured range differs from %q", test.expected)
			}
		})
	}
}

// Expected containment from node-semver 7.8.5, resolved by eslint-plugin-n v18.3.0.
func TestNodeVersionSubsetRanges(t *testing.T) {
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
			r, ok := parseNodeVersion(test.text)
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

func TestInvalidNodeVersionRanges(t *testing.T) {
	for _, text := range []string{"invalid", ">=", "16 | 20", "16, 20", "\u008510", "10 - 12\u0085", "<=04294967296", "<=4294967296.invalid", ">=4294967296 || invalid", ">>4294967296", "<=9007199254740992"} {
		if _, ok := parseNodeVersion(text); ok {
			t.Errorf("accepted invalid range %q", text)
		}
	}
}

func TestExtremeNodeVersionRanges(t *testing.T) {
	for _, text := range []string{
		"<=4294967296", ">=12 <4294967296", ">4294967295", "^0.0.4294967295",
		"~12.4294967295.0", "12 - 4294967295", "^16 || <=4294967296", "<=9007199254740991",
	} {
		t.Run(text, func(t *testing.T) {
			version, ok := parseNodeVersion(text)
			if !ok || !version.uncertain {
				t.Fatal("expected a valid range with uncertain feature support")
			}
			if version.Supports("5.10.0") || version.IsSubsetOf("^12.20.0 || >=14.13.1") {
				t.Fatal("an extreme range must not imply feature support")
			}
		})
	}
	for _, text := range []string{"16.0.0+4294967296", "16.0.0-4294967296", "4294967294.0.0"} {
		if version, ok := parseNodeVersion(text); !ok || version.uncertain {
			t.Errorf("ordinary range or large metadata treated as uncertain: %s", text)
		}
	}
}
