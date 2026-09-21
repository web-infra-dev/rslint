package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnusedVarsVue locks in how a Vue component's bindings are judged. A
// component with a script setup block exposes every top-level value binding to
// its template, which the rule cannot see, so those bindings count as used. A
// type emits nothing a template could reach, so it is still judged normally.
func TestNoUnusedVarsVue(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     "<script setup lang=\"ts\">\nimport MyThing from './MyThing.vue';\nimport { ref } from 'vue';\nconst count = ref(0);\nfunction handler(): void {}\n</script>\n\n<template>\n  <MyThing :n=\"count\" @go=\"handler\" />\n</template>\n",
				FileName: "template-uses-setup-bindings.vue",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "<script setup lang=\"ts\">\ninterface Shape { a: number }\n</script>\n",
				FileName: "unused-interface.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2}},
			},
			{
				Code:     "<script lang=\"ts\">\nconst unused = 1;\nexport default {};\n</script>\n",
				FileName: "options-api.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 2}},
			},
			{
				Code:     "<script setup lang=\"ts\">\nfunction outer(): void {\n  const inner = 1;\n}\n</script>\n",
				FileName: "nested-binding.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 3}},
			},
		},
	)
}
