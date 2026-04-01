package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsImports(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// import: used
		{Code: `import type { Foo } from "./foo"; const bar: Foo = {} as any; console.log(bar);`},
		// namespace import: used
		{Code: `import * as path from "path"; console.log(path.join("a", "b"));`},
		// import equals: used
		{Code: `import path = require("path"); console.log(path.join("a", "b"));`},
		// side-effect import: no binding, not affected
		{Code: `import "path";`},
		// import then re-export: used
		{Code: `import { join } from "path"; export { join };`},
		// re-export with rename: used
		{Code: `import { join } from "path"; export { join as myJoin };`},
		// namespace import re-exported
		{Code: `import * as path from "path"; export { path };`},
		// import used via export default
		{Code: `import { join } from "path"; export default join;`},
		// direct re-export (no local binding, rule doesn't apply)
		{Code: `export { join } from "path";`},
		// import used in multiple places
		{Code: `import { join, resolve } from "path"; console.log(join("a"), resolve("b"));`},
		// default + named import, both used
		{Code: `import Def, { join } from "path"; console.log(Def, join("a"));`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// --- import: unused (with suggestion fixers) ---
		{
			Code: `import { Foo } from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// namespace import: unused
		{
			Code: `import * as ns from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// import equals: unused
		{
			Code: `import path = require("path");`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// type import unused → remove entire line
		{
			Code: `import type { Foo } from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// default import unused → remove entire line
		{
			Code: `import Foo from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// alias import unused — reported at the local name (r), not the original name
		{
			Code: `import { resolve as r } from "path";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},

		// --- enableAutofixRemoval.imports = true: fix applied as autofix ---
		{
			Code:    `import { Foo } from "./foo";`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 10}},
			Output:  []string{``},
		},
		// enableAutofixRemoval.imports = false (explicit): fix applied as suggestion
		{
			Code:    `import { Foo } from "./foo";`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": false}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// autofix mode: namespace import
		{
			Code:    `import * as ns from "./foo";`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 13}},
			Output:  []string{``},
		},
		// autofix mode: import equals
		{
			Code:    `import path = require("path");`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 8}},
			Output:  []string{``},
		},

		// --- import fix: partial removal ---
		// first specifier unused, second used → remove first + trailing comma
		{
			Code: `import { resolve, join } from "path"; console.log(join("a", "b"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join } from "path"; console.log(join("a", "b"));`,
				}},
			}},
		},
		// second specifier unused, first used → remove leading comma + second
		{
			Code: `import { join, resolve } from "path"; console.log(join("a", "b"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join } from "path"; console.log(join("a", "b"));`,
				}},
			}},
		},
		// partial removal with autofix mode
		{
			Code:    `import { join, resolve } from "path"; console.log(join("a", "b"));`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 16}},
			Output:  []string{`import { join } from "path"; console.log(join("a", "b"));`},
		},
		// middle specifier unused (three specifiers)
		{
			Code: `import { join, resolve, basename } from "path"; console.log(join("a", "b"), basename("c"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join, basename } from "path"; console.log(join("a", "b"), basename("c"));`,
				}},
			}},
		},
		// alias import: partial removal (alias unused, another specifier used)
		{
			Code: `import { join, resolve as r } from "path"; console.log(join("a", "b"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join } from "path"; console.log(join("a", "b"));`,
				}},
			}},
		},
		// partial re-export: join re-exported (used), resolve not → resolve reported
		{
			Code: `import { join, resolve } from "path"; export { join };`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join } from "path"; export { join };`,
				}},
			}},
		},
		// multiline import: partial removal
		{
			Code: `import {
  join,
  resolve
} from "path";
console.log(join("a", "b"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 3, Column: 3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output: `import {
  join
} from "path";
console.log(join("a", "b"));`,
				}},
			}},
		},
		// three specifiers, two unused → each gets its own suggestion
		{
			Code: `import { join, resolve, basename } from "path"; console.log(resolve("/"));`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { resolve, basename } from "path"; console.log(resolve("/"));`,
					}},
				},
				{MessageId: "unusedVar", Line: 1, Column: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { join, resolve } from "path"; console.log(resolve("/"));`,
					}},
				},
			},
		},
		// four specifiers, three unused
		{
			Code: `import { a, b, c, d } from "./foo"; console.log(b);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { b, c, d } from "./foo"; console.log(b);`,
					}},
				},
				{MessageId: "unusedVar", Line: 1, Column: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { a, b, d } from "./foo"; console.log(b);`,
					}},
				},
				{MessageId: "unusedVar", Line: 1, Column: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { a, b, c } from "./foo"; console.log(b);`,
					}},
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}
