package no_restricted_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoRestrictedImportsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedImportsRule, []rule_tester.ValidTestCase{
		// Basic valid cases
		{Code: "import foo from 'foo';"},
		{Code: "import foo = require('foo');"},
		{Code: "import 'foo';"},
		{Code: "export { foo } from 'foo';"},
		{Code: "export * from 'foo';"},
		
		// Empty options
		{Code: "import foo from 'foo';", Options: []interface{}{}},
		{Code: "import foo from 'foo';", Options: map[string]interface{}{"paths": []interface{}{}}},
		{Code: "import foo from 'foo';", Options: map[string]interface{}{"patterns": []interface{}{}}},
		{Code: "import foo from 'foo';", Options: map[string]interface{}{"paths": []interface{}{}, "patterns": []interface{}{}}},
		
		// Valid imports with restrictions on other modules
		{
			Code:    "import foo from 'foo';",
			Options: []interface{}{"import1", "import2"},
		},
		{
			Code:    "import foo = require('foo');",
			Options: []interface{}{"import1", "import2"},
		},
		{
			Code:    "export { foo } from 'foo';",
			Options: []interface{}{"import1", "import2"},
		},
		{
			Code:    "import foo from 'foo';",
			Options: map[string]interface{}{"paths": []interface{}{"import1", "import2"}},
		},
		{
			Code:    "export { foo } from 'foo';",
			Options: map[string]interface{}{"paths": []interface{}{"import1", "import2"}},
		},
		{
			Code:    "import 'foo';",
			Options: []interface{}{"import1", "import2"},
		},
		
		// Pattern-based restrictions
		{
			Code: "import foo from 'foo';",
			Options: map[string]interface{}{
				"paths":    []interface{}{"import1", "import2"},
				"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
			},
		},
		{
			Code: "export { foo } from 'foo';",
			Options: map[string]interface{}{
				"paths":    []interface{}{"import1", "import2"},
				"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
			},
		},
		
		// Custom message paths
		{
			Code: "import foo from 'foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
					map[string]interface{}{
						"name":    "import-baz",
						"message": "Please use import-quux instead.",
					},
				},
			},
		},
		{
			Code: "export { foo } from 'foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
					map[string]interface{}{
						"name":    "import-baz",
						"message": "Please use import-quux instead.",
					},
				},
			},
		},
		
		// Import names restrictions
		{
			Code: "import foo from 'foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "import-foo",
						"importNames": []interface{}{"Bar"},
						"message":     "Please use Bar from /import-bar/baz/ instead.",
					},
				},
			},
		},
		{
			Code: "export { foo } from 'foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "import-foo",
						"importNames": []interface{}{"Bar"},
						"message":     "Please use Bar from /import-bar/baz/ instead.",
					},
				},
			},
		},
		
		// Pattern groups with messages
		{
			Code: "import foo from 'foo';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import1/private/*"},
						"message": "usage of import1 private modules not allowed.",
					},
					map[string]interface{}{
						"group":   []interface{}{"import2/*", "!import2/good"},
						"message": "import2 is deprecated, except the modules in import2/good.",
					},
				},
			},
		},
		{
			Code: "export { foo } from 'foo';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import1/private/*"},
						"message": "usage of import1 private modules not allowed.",
					},
					map[string]interface{}{
						"group":   []interface{}{"import2/*", "!import2/good"},
						"message": "import2 is deprecated, except the modules in import2/good.",
					},
				},
			},
		},
		
		// Type imports with allowTypeImports
		{
			Code: "import type foo from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: "import type _ = require('import-foo');",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: "import type { Bar } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"importNames":      []interface{}{"Bar"},
						"message":          "Please use Bar from /import-bar/baz/ instead.",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: "export type { Bar } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"importNames":      []interface{}{"Bar"},
						"message":          "Please use Bar from /import-bar/baz/ instead.",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: "import type foo from 'import1/private/bar';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":            []interface{}{"import1/private/*"},
						"message":          "usage of import1 private modules not allowed.",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: "export type { foo } from 'import1/private/bar';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":            []interface{}{"import1/private/*"},
						"message":          "usage of import1 private modules not allowed.",
						"allowTypeImports": true,
					},
				},
			},
		},
		
		// Mixed type imports
		{
			Code: "import { type Bar } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"importNames":      []interface{}{"Bar"},
						"message":          "Please use Bar from /import-bar/baz/ instead.",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: "export { type Bar } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"importNames":      []interface{}{"Bar"},
						"message":          "Please use Bar from /import-bar/baz/ instead.",
						"allowTypeImports": true,
					},
				},
			},
		},
		
		// Regex patterns
		{
			Code: "import type { foo } from 'import1/private/bar';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"regex":            "import1/.*",
						"message":          "usage of import1 private modules not allowed.",
						"allowTypeImports": true,
					},
				},
			},
		},
		{
			Code: "import { foo } from 'import1/private';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"regex":            "import1/[A-Z]+",
						"message":          "usage of import1 private modules not allowed.",
						"caseSensitive":    true,
						"allowTypeImports": true,
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		// Basic invalid cases
		{
			Code:    "import foo from 'import1';",
			Options: []interface{}{"import1", "import2"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    "import foo = require('import1');",
			Options: []interface{}{"import1", "import2"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    "export { foo } from 'import1';",
			Options: []interface{}{"import1", "import2"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    "import foo from 'import1';",
			Options: map[string]interface{}{"paths": []interface{}{"import1", "import2"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    "export { foo } from 'import1';",
			Options: map[string]interface{}{"paths": []interface{}{"import1", "import2"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Pattern-based errors
		{
			Code: "import foo from 'import1/private/foo';",
			Options: map[string]interface{}{
				"paths":    []interface{}{"import1", "import2"},
				"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { foo } from 'import1/private/foo';",
			Options: map[string]interface{}{
				"paths":    []interface{}{"import1", "import2"},
				"patterns": []interface{}{"import1/private/*", "import2/*", "!import2/good"},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patterns",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Custom message errors
		{
			Code: "import foo from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
					map[string]interface{}{
						"name":    "import-baz",
						"message": "Please use import-quux instead.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { foo } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":    "import-foo",
						"message": "Please use import-bar instead.",
					},
					map[string]interface{}{
						"name":    "import-baz",
						"message": "Please use import-quux instead.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Import names errors
		{
			Code: "import { Bar } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "import-foo",
						"importNames": []interface{}{"Bar"},
						"message":     "Please use Bar from /import-bar/baz/ instead.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importNameWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { Bar } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":        "import-foo",
						"importNames": []interface{}{"Bar"},
						"message":     "Please use Bar from /import-bar/baz/ instead.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importNameWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Pattern group errors
		{
			Code: "import foo from 'import1/private/foo';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import1/private/*"},
						"message": "usage of import1 private modules not allowed.",
					},
					map[string]interface{}{
						"group":   []interface{}{"import2/*", "!import2/good"},
						"message": "import2 is deprecated, except the modules in import2/good.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patternWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { foo } from 'import1/private/foo';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":   []interface{}{"import1/private/*"},
						"message": "usage of import1 private modules not allowed.",
					},
					map[string]interface{}{
						"group":   []interface{}{"import2/*", "!import2/good"},
						"message": "import2 is deprecated, except the modules in import2/good.",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patternWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Side-effect imports
		{
			Code: "import 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name": "import-foo",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "import 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Type imports not allowed
		{
			Code: "import foo from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "import foo = require('import-foo');",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"message":          "Please use import-bar instead.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "pathWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "import { Bar } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"importNames":      []interface{}{"Bar"},
						"message":          "Please use Bar from /import-bar/baz/ instead.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importNameWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { Bar } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"importNames":      []interface{}{"Bar"},
						"message":          "Please use Bar from /import-bar/baz/ instead.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importNameWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "import foo from 'import1/private/bar';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":            []interface{}{"import1/private/*"},
						"message":          "usage of import1 private modules not allowed.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patternWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { foo } from 'import1/private/bar';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"group":            []interface{}{"import1/private/*"},
						"message":          "usage of import1 private modules not allowed.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patternWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Export all
		{
			Code:    "export * from 'import1';",
			Options: []interface{}{"import1"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "path",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Regex patterns
		{
			Code: "export { foo } from 'import1/private/bar';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"regex":            "import1/.*",
						"message":          "usage of import1 private modules not allowed.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patternWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "import { foo } from 'import1/private-package';",
			Options: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"regex":            "import1/private-[a-z]*",
						"message":          "usage of import1 private modules not allowed.",
						"caseSensitive":    true,
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "patternWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		
		// Mixed type imports - both Bar and Baz restricted
		{
			Code: "import { Bar, type Baz } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"importNames":      []interface{}{"Bar", "Baz"},
						"message":          "Please use Bar and Baz from /import-bar/baz/ instead.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importNameWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "export { Bar, type Baz } from 'import-foo';",
			Options: map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"name":             "import-foo",
						"importNames":      []interface{}{"Bar", "Baz"},
						"message":          "Please use Bar and Baz from /import-bar/baz/ instead.",
						"allowTypeImports": true,
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "importNameWithCustomMessage",
					Line:      1,
					Column:    1,
				},
			},
		},
	})
}