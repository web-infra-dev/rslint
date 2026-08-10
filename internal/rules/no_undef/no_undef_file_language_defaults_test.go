package no_undef

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUndefFileLanguageDefaults(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.allow-js.json",
		t,
		&NoUndefRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     `exports; global; module; require; arguments;`,
				FileName: "file.cjs",
			},
			{
				Code:     `(() => arguments)();`,
				FileName: "arrow.cjs",
			},
			{
				Code:     `arguments;`,
				FileName: "arguments-off.cjs",
				Globals:  map[string]any{"arguments": "off"},
			},
			{
				Code:     `const require = 1; require;`,
				FileName: "local-require.cjs",
				Globals:  map[string]any{"require": "off"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `require;`,
				FileName: "plain-script.js",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     `arguments;`,
				FileName: "module-file.mjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     `require;`,
				FileName: "typed-commonjs.cts",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     `process;`,
				FileName: "node-name.cjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     `__dirname;`,
				FileName: "node-path.cjs",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
			{
				Code:     `require;`,
				FileName: "require-off.cjs",
				Globals:  map[string]any{"require": "off"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "undef", Line: 1, Column: 1}},
			},
		},
	)
}
