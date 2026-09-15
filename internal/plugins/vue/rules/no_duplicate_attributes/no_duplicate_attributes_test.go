package no_duplicate_attributes

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/vue/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func noCoexist(class, style bool) map[string]any {
	return map[string]any{"allowCoexistClass": class, "allowCoexistStyle": style}
}

func TestNoDuplicateAttributesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoDuplicateAttributesRule,
		[]rule_tester.ValidTestCase{
			{
				Code:     "<template><div foo=\"a\" bar=\"b\"/></template>\n",
				FileName: "valid-distinct.vue",
			},
			{
				Code:     "<template><div :foo=\"a\" :bar=\"b\"/></template>\n",
				FileName: "valid-distinct-bindings.vue",
			},
			// The same name on two different elements is not a duplicate.
			{
				Code:     "<template><div foo=\"a\"/><span foo=\"b\"/></template>\n",
				FileName: "valid-different-elements.vue",
			},
			// Vue merges a static class with a bound one, so this is the idiom
			// rather than a mistake.
			{
				Code:     "<template><div class=\"a\" :class=\"b\"/></template>\n",
				FileName: "valid-class-coexists.vue",
			},
			{
				Code:     "<template><div style=\"color:red\" :style=\"s\"/></template>\n",
				FileName: "valid-style-coexists.vue",
			},
			// A v-bind with no argument spreads an object whose keys are only
			// known at run time, so it can collide with nothing.
			{
				Code:     "<template><div v-bind=\"attrs\" foo=\"a\"/></template>\n",
				FileName: "valid-object-spread.vue",
			},
			{
				Code:     "<template><div v-bind=\"a\" v-bind=\"b\"/></template>\n",
				FileName: "valid-two-spreads.vue",
			},
			// A dynamic argument names nothing statically.
			{
				Code:     "<template><div :[key]=\"a\" :[other]=\"b\"/></template>\n",
				FileName: "valid-dynamic-arguments.vue",
			},
			// A directive that is not v-bind is not an attribute of the
			// rendered element.
			{
				Code:     "<template><div v-if=\"a\" v-else-if=\"b\"/></template>\n",
				FileName: "valid-structural-directives.vue",
			},
			// Distinct events are distinct names.
			{
				Code:     "<template><div @click=\"a\" @focus=\"b\"/></template>\n",
				FileName: "valid-distinct-events.vue",
			},
			// Two handlers for one event are not a duplicate: Vue invokes
			// both, and upstream compares names only for v-bind.
			{
				Code:     "<template><div @click=\"a\" @click=\"b\"/></template>\n",
				FileName: "valid-two-handlers.vue",
			},
			// No template block at all.
			{
				Code:     "<script>\nexport default {};\n</script>\n",
				FileName: "valid-no-template.vue",
			},
			{
				Code:     "const value = 1;\nexport default value;\n",
				FileName: "valid-not-a-component.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "<template><div foo=\"a\" foo=\"b\"/></template>\n",
				FileName: "invalid-plain.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1}},
			},
			// The two spellings name the same attribute.
			{
				Code:     "<template><div foo=\"a\" :foo=\"b\"/></template>\n",
				FileName: "invalid-plain-and-bound.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1}},
			},
			{
				Code:     "<template><div :foo=\"a\" v-bind:foo=\"b\"/></template>\n",
				FileName: "invalid-shorthand-and-long.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1}},
			},
			// Coexistence is allowed across the two halves, never within one.
			{
				Code:     "<template><div class=\"a\" class=\"b\"/></template>\n",
				FileName: "invalid-two-static-classes.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1}},
			},
			{
				Code:     "<template><div :class=\"a\" :class=\"b\"/></template>\n",
				FileName: "invalid-two-bound-classes.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1}},
			},
			// Three of a kind reports the second and the third.
			{
				Code:     "<template><div foo=\"a\" foo=\"b\" foo=\"c\"/></template>\n",
				FileName: "invalid-three.vue",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: messageID, Line: 1},
					{MessageId: messageID, Line: 1},
				},
			},
			// A nested element is checked too, and reported on its own line.
			{
				Code:     "<template>\n  <ul>\n    <li foo=\"a\" foo=\"b\"/>\n  </ul>\n</template>\n",
				FileName: "invalid-nested.vue",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 3}},
			},
			// Turning the allowance off makes the idiom a duplicate.
			{
				Code:     "<template><div class=\"a\" :class=\"b\"/></template>\n",
				FileName: "invalid-class-not-allowed.vue",
				Options:  noCoexist(false, true),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1}},
			},
			{
				Code:     "<template><div style=\"color:red\" :style=\"s\"/></template>\n",
				FileName: "invalid-style-not-allowed.vue",
				Options:  noCoexist(true, false),
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: messageID, Line: 1}},
			},
		},
	)
}
