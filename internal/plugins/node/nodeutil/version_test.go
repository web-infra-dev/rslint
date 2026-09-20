package nodeutil

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/npmsemver"
)

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
			want, ok := npmsemver.Parse(test.expected)
			if !ok {
				t.Fatal("invalid expected range")
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("configured range differs from %q", test.expected)
			}
		})
	}
}
