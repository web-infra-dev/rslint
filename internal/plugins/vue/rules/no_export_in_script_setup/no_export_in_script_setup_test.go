package no_export_in_script_setup

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/vue/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExportInScriptSetupRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExportInScriptSetupRule,
		[]rule_tester.ValidTestCase{
			// A plain <script> block is an ordinary module.
			{
				Code:     "<script>\nexport default { name: 'App' };\n</script>\n",
				FileName: "valid-plain-script.vue",
			},
			// Nothing exported.
			{
				Code:     "<script setup>\nconst value = 1;\n</script>\n",
				FileName: "valid-no-export.vue",
			},
			// The exports live in the plain block; the setup block beside it is
			// clean. This is the case a rule that only looked at the parsed
			// text would get wrong.
			{
				Code: "<script>\nexport default { name: 'App' };\nexport const shared = 1;\n</script>\n" +
					"<script setup>\nconst local = 2;\n</script>\n",
				FileName: "valid-exports-in-plain-block.vue",
			},
			// A type-only export emits nothing, so the Vue compiler never sees
			// it.
			{
				Code:     "<script setup lang=\"ts\">\nexport type Value = number;\n</script>\n",
				FileName: "valid-type-alias.vue",
			},
			{
				Code:     "<script setup lang=\"ts\">\ninterface Value { a: number }\nexport type { Value };\n</script>\n",
				FileName: "valid-export-type-clause.vue",
			},
			{
				Code:     "<script setup lang=\"ts\">\ninterface Value { a: number }\nexport { type Value };\n</script>\n",
				FileName: "valid-type-only-specifiers.vue",
			},
			{
				Code:     "<script setup lang=\"ts\">\nexport interface Value { a: number }\n</script>\n",
				FileName: "valid-exported-interface.vue",
			},
			// An import is not an export.
			{
				Code:     "<script setup>\nimport { ref } from 'vue';\nconst value = ref(1);\n</script>\n",
				FileName: "valid-import.vue",
			},
			// A file that is not a component has no setup block to speak of.
			{
				Code:     "export default { name: 'App' };\n",
				FileName: "valid-not-a-component.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "<script setup>\nexport default { name: 'App' };\n</script>\n",
				FileName: "invalid-default.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 2}},
			},
			{
				Code:     "<script setup>\nconst value = 1;\nexport { value };\n</script>\n",
				FileName: "invalid-named-clause.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 3}},
			},
			{
				Code:     "<script setup>\nexport * from './other';\n</script>\n",
				FileName: "invalid-export-all.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 2}},
			},
			// The shape upstream sees as ExportNamedDeclaration with a
			// declaration, and TypeScript sees as a declaration carrying an
			// export modifier.
			{
				Code:     "<script setup>\nexport const value = 1;\n</script>\n",
				FileName: "invalid-export-const.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 2}},
			},
			{
				Code:     "<script setup>\nexport function handler() {}\n</script>\n",
				FileName: "invalid-export-function.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 2}},
			},
			{
				Code:     "<script setup>\nexport class Thing {}\n</script>\n",
				FileName: "invalid-export-class.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 2}},
			},
			// An enum does emit a runtime value, unlike an interface.
			{
				Code:     "<script setup lang=\"ts\">\nexport enum Kind { A }\n</script>\n",
				FileName: "invalid-export-enum.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 2}},
			},
			// Only the setup block's export is reported when both blocks are
			// present and both export.
			{
				Code: "<script>\nexport const shared = 1;\n</script>\n" +
					"<script setup>\nexport const local = 2;\n</script>\n",
				FileName: "invalid-only-setup-block.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 5}},
			},
			{
				Code:     "<script setup>\nexport const a = 1;\nexport const b = 2;\n</script>\n",
				FileName: "invalid-two-exports.vue",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 2},
					{MessageId: messageID, Line: 3},
				},
			},
			// A mixed clause still carries a runtime export.
			{
				Code:     "<script setup lang=\"ts\">\nconst value = 1;\ninterface Shape { a: number }\nexport { value, type Shape };\n</script>\n",
				FileName: "invalid-mixed-specifiers.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 4}},
			},
		},
	)
}
