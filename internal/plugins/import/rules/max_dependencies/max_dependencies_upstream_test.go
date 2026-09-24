package max_dependencies_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/max_dependencies"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All cases from https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/max-dependencies.js
// and https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/max-dependencies.md.
// Flow and TypeScript type imports share the same tsgo representation.
func TestMaxDependenciesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &max_dependencies.MaxDependenciesRule,
		[]rule_tester.ValidTestCase{
			// max-dependencies.
			{
				Code: "import \"./foo.js\"",
			},
			{
				Code:    "import \"./foo.js\"; import \"./bar.js\";",
				Options: []any{map[string]any{"max": 2}},
			},
			{
				Code:    "import \"./foo.js\"; import \"./bar.js\"; const a = require(\"./foo.js\"); const b = require(\"./bar.js\");",
				Options: []any{map[string]any{"max": 2}},
			},
			{
				Code: "import {x, y, z} from \"./foo\"",
			},
			// max-dependencies (typescript).
			{
				Code:    "import type { x } from './foo'; import { y } from './bar';",
				Options: []any{map[string]any{"max": 1, "ignoreTypeImports": true}},
			},
			// Documentation examples.
			{
				Code:    "import a from './a'; // 1\nconst anotherA = require('./a'); // still 1\nimport {x, y, z} from './foo'; // 2",
				Options: []any{map[string]any{"max": 2}},
			},
			{
				Code:    "import a from './a';\nimport b from './b';\nimport type c from './c'; // Doesn't count against max",
				Options: []any{map[string]any{"max": 2, "ignoreTypeImports": true}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// max-dependencies.
			{
				Code:    "import { x } from './foo'; import { y } from './foo'; import {z} from './bar';",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 71, EndLine: 1, EndColumn: 78}},
			},
			{
				Code:    "import { x } from './foo'; import { y } from './bar'; import { z } from './baz';",
				Options: []any{map[string]any{"max": 2}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (2) exceeded.", Line: 1, Column: 73, EndLine: 1, EndColumn: 80}},
			},
			{
				Code:    "import { x } from './foo'; require(\"./bar\"); import { z } from './baz';",
				Options: []any{map[string]any{"max": 2}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (2) exceeded.", Line: 1, Column: 64, EndLine: 1, EndColumn: 71}},
			},
			{
				Code:    "import { x } from './foo'; import { z } from './foo'; require(\"./bar\"); const path = require(\"path\");",
				Options: []any{map[string]any{"max": 2}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (2) exceeded.", Line: 1, Column: 94, EndLine: 1, EndColumn: 100}},
			},
			{
				Code:    "import type { x } from './foo'; import type { y } from './bar'",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 56, EndLine: 1, EndColumn: 63}},
			},
			{
				Code:    "import type { x } from './foo'; import type { y } from './bar'; import type { z } from './baz'",
				Options: []any{map[string]any{"max": 2, "ignoreTypeImports": false}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (2) exceeded.", Line: 1, Column: 88, EndLine: 1, EndColumn: 95}},
			},
			// max-dependencies (typescript).
			{
				Code:    "import type { x } from './foo'; import type { y } from './bar'",
				Options: []any{map[string]any{"max": 1}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (1) exceeded.", Line: 1, Column: 56, EndLine: 1, EndColumn: 63}},
			},
			{
				Code:    "import type { x } from './foo'; import type { y } from './bar'; import type { z } from './baz'",
				Options: []any{map[string]any{"max": 2, "ignoreTypeImports": false}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (2) exceeded.", Line: 1, Column: 88, EndLine: 1, EndColumn: 95}},
			},
			// Documentation examples.
			{
				Code:    "import a from './a'; // 1\nconst b = require('./b'); // 2\nimport c from './c'; // 3 - exceeds max!",
				Options: []any{map[string]any{"max": 2}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (2) exceeded.", Line: 3, Column: 15, EndLine: 3, EndColumn: 20}},
			},
			{
				Code:    "import a from './a';\nimport b from './b';\nimport c from './c';",
				Options: []any{map[string]any{"max": 2, "ignoreTypeImports": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "Maximum number of dependencies (2) exceeded.", Line: 3, Column: 15, EndLine: 3, EndColumn: 20}},
			},
		},
	)
}
