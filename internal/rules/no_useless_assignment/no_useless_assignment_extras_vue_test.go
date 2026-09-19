package no_useless_assignment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUselessAssignmentExtrasVue locks in that a store read only by a Vue
// template is not useless: the template reads script setup bindings at render
// time, which no path through the script shows.
func TestNoUselessAssignmentExtrasVue(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessAssignmentRule,
		[]rule_tester.ValidTestCase{
			// The last store follows the script's only read of message, so the
			// template is the only thing that reads it.
			{
				Code:     "<script setup>\nimport { ref } from 'vue';\nconst title = ref('x');\nlet message = 'a';\nconsole.log(message, title.value);\nmessage = 'b';\n</script>\n<template>{{ message }}</template>\n",
				FileName: "template-reads-last-store.vue",
			},
		},
		[]rule_tester.InvalidTestCase{
			// A binding the template cannot reach is still tracked.
			{
				Code:     "<script setup>\nfunction f() {\n  let v = 1;\n  v = 2;\n  return v;\n}\n</script>\n",
				FileName: "nested-store.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "unnecessaryAssignment", Line: 3}},
			},
		},
	)
}
