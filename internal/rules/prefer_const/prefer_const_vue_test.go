package prefer_const

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferConstVue locks in why a script setup `let` is left alone: the
// template may assign to it, so the `const` this rule would suggest, and the
// autofix that writes it, would break the component.
func TestPreferConstVue(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferConstRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     "<script setup>\nlet count = 0;\n</script>\n<template><button @click=\"count++\">{{ count }}</button></template>\n",
				FileName: "template-assigns-let.vue",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Without a setup block nothing is exposed, and the fix lands in the
			// component at the right offset.
			{
				Code:     "<script>\nlet x = 1;\nconsole.log(x);\nexport default {};\n</script>\n",
				FileName: "options-api.vue",
				Output:   []string{"<script>\nconst x = 1;\nconsole.log(x);\nexport default {};\n</script>\n"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "useConst", Line: 2}},
			},
			// Only top-level bindings reach the template.
			{
				Code:     "<script setup>\nfunction f() {\n  let y = 1;\n  return y;\n}\n</script>\n",
				FileName: "nested-let.vue",
				Output:   []string{"<script setup>\nfunction f() {\n  const y = 1;\n  return y;\n}\n</script>\n"},
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "useConst", Line: 3}},
			},
		},
	)
}
