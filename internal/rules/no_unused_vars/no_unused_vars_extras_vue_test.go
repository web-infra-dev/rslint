package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnusedVarsExtrasVue locks in how a Vue component's bindings are judged.
// A component with a script setup block exposes every top-level binding to its
// template, which the rule cannot see, so those bindings count as used.
func TestNoUnusedVarsExtrasVue(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// Every binding here is used only by the template.
			{
				Code:     "<script setup>\nimport MyThing from './MyThing.vue';\nimport { ref } from 'vue';\nconst count = ref(0);\nfunction handler() {}\n</script>\n\n<template>\n  <MyThing :n=\"count\" @go=\"handler\" />\n</template>\n",
				FileName: "template-uses-setup-bindings.vue",
			},
			// compileScript also exposes a plain script's bindings when a setup
			// block sits beside it.
			{
				Code:     "<script>\nconst shared = 1;\n</script>\n<script setup>\nconst local = 2;\n</script>\n<template>{{ shared }} {{ local }}</template>\n",
				FileName: "plain-script-beside-setup.vue",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Without a setup block nothing is exposed implicitly.
			{
				Code:     "<script>\nconst unused = 1;\nexport default {};\n</script>\n",
				FileName: "options-api.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "removeVar", Output: "<script>\n\nexport default {};\n</script>\n"}}}},
			},
			// Only top-level bindings reach the template.
			{
				Code:     "<script setup>\nfunction outer() {\n  const inner = 1;\n}\n</script>\n",
				FileName: "nested-binding.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "removeVar", Output: "<script setup>\nfunction outer() {\n  \n}\n</script>\n"}}}},
			},
		},
	)
}
