package order_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestOrderMinimatchCompatibility(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&order.OrderRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    "import react from 'react';\nimport scope from 'scope';",
				Options: minimatchPathGroupOptions("scope/package", map[string]any{}),
			},
		},
		[]rule_tester.InvalidTestCase{
			// Empty module specifiers and patterns are both legal strings;
			// minimatch 3 treats them as an exact match.
			{
				Code:    "import react from 'react';\nimport empty from '';",
				Options: minimatchPathGroupOptions("", nil),
				Output:  []string{"import empty from '';\nimport react from 'react';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// import-js 2.32 uses minimatch 3.1.5, whose brace expansion includes
			// numeric/letter ranges in addition to comma alternatives.
			{
				Code:    "import react from 'react';\nimport ranged from 'pkg/2';",
				Options: minimatchPathGroupOptions("pkg/{1..3}", nil),
				Output:  []string{"import ranged from 'pkg/2';\nimport react from 'react';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// partial accepts a module specifier that is a matching prefix of the
			// configured pattern.
			{
				Code:    "import react from 'react';\nimport scope from 'scope';",
				Options: minimatchPathGroupOptions("scope/package", map[string]any{"partial": true}),
				Output:  []string{"import scope from 'scope';\nimport react from 'react';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
			// flipNegate makes the non-negated portion of a leading-! pattern the
			// positive match, matching minimatch's option semantics.
			{
				Code:    "import other from 'other';\nimport react from 'react';",
				Options: minimatchPathGroupOptions("!react", map[string]any{"flipNegate": true}),
				Output:  []string{"import react from 'react';\nimport other from 'other';\n"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "order", Line: 2, Column: 1}},
			},
		},
	)
}

func minimatchPathGroupOptions(pattern string, patternOptions map[string]any) map[string]any {
	pathGroup := map[string]any{
		"pattern": pattern,
		"group":   "internal",
	}
	if patternOptions != nil {
		pathGroup["patternOptions"] = patternOptions
	}
	return map[string]any{
		"groups":                        []any{"internal", "external"},
		"pathGroups":                    []any{pathGroup},
		"pathGroupsExcludedImportTypes": []any{},
	}
}
