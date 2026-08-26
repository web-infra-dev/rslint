package default_case

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDefaultCaseRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&DefaultCaseRule,
		[]rule_tester.ValidTestCase{
			{Code: `switch (a) { case 1: break; default: break; }`},
			{Code: `switch (a) { case 1: break; case 2: default: break; }`},
			{Code: "switch (a) {\n  case 1:\n    break;\n  // no default\n}"},
			{Code: `switch (a) { case 1: break; /* no default */ }`},
			{Code: "switch (a) {\n  case 1:\n    break;\n  // No Default\n}"},
			{Code: `switch (a) {}`},
			{
				Code:    "switch (a) {\n  case 1:\n    break;\n  // skip default\n}",
				Options: []any{map[string]any{"commentPattern": `^skip\sdefault`}},
			},
			{
				Code:    `switch (a) { case 1: break; /* skip default */ }`,
				Options: []any{map[string]any{"commentPattern": `^skip\sdefault`}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// A configured commentPattern is compiled without the `i` flag, so
			// it matches case-sensitively, unlike the default pattern.
			{
				Code:    "switch (a) {\n  case 1:\n    break;\n  // SKIP DEFAULT\n}",
				Options: []any{map[string]any{"commentPattern": `^skip\sdefault`}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingDefaultCase", Line: 1, Column: 1},
				},
			},
			// A configured commentPattern replaces the default one.
			{
				Code:    "switch (a) {\n  case 1:\n    break;\n  // no default\n}",
				Options: []any{map[string]any{"commentPattern": `^skip\sdefault`}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingDefaultCase", Line: 1, Column: 1},
				},
			},
			{
				Code: `switch (a) { case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingDefaultCase", Line: 1, Column: 1},
				},
			},
			{
				Code: `switch (a) { case 1: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingDefaultCase", Line: 1, Column: 1},
				},
			},
			// Only the last comment is checked — "no default" followed by another comment should still report
			{
				Code: "switch (a) {\n  case 1:\n    break;\n  // no default\n  // actually we should add one later\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingDefaultCase", Line: 1, Column: 1},
				},
			},
		},
	)
}

// TestDefaultCaseCommentPatternSchema locks in the fail-fast behavior for an
// invalid commentPattern. Upstream constructs `new RegExp(commentPattern, "u")`
// while loading the rule, so an invalid pattern like `"("` throws before any
// linting happens. rslint's equivalent surface is config validation: the schema
// marks commentPattern with `format: "regex"`, so validation rejects the config
// up front instead of silently linting with the default pattern.
func TestDefaultCaseCommentPatternSchema(t *testing.T) {
	invalid := []any{map[string]any{"commentPattern": "("}}
	if err := DefaultCaseRule.Schema.Validate(invalid); err == nil {
		t.Error("expected an invalid commentPattern regex to fail schema validation")
	}
	// Lookbehind is JS-legal but RE2-illegal; it must still validate, proving
	// the schema checks patterns with the ECMAScript engine, not Go's regexp.
	valid := []any{map[string]any{"commentPattern": "(?<=a)b"}}
	if err := DefaultCaseRule.Schema.Validate(valid); err != nil {
		t.Errorf("expected a valid commentPattern regex to pass schema validation, got: %v", err)
	}
}
