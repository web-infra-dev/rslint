// TestSortImportsUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/sort-imports.js 1:1. Position assertions cover every
// invalid case. rslint-specific lock-in cases live in sort_imports_extras_test.go.
package sort_imports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func declarationError(id, message string, line, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: id, Message: message, Line: line, Column: 1, EndLine: line, EndColumn: endColumn}
}

func memberError(name string, line, column, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "sortMembersAlphabetically",
		Message:   "Member '" + name + "' of the import declaration should be sorted alphabetically.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

func TestSortImportsUpstream(t *testing.T) {
	ignoreCase := []any{map[string]any{"ignoreCase": true}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SortImportsRule,
		[]rule_tester.ValidTestCase{
			{Code: "import a from 'foo.js';\nimport b from 'bar.js';\nimport c from 'baz.js';\n"},
			{Code: "import * as B from 'foo.js';\nimport A from 'bar.js';"},
			{Code: "import * as B from 'foo.js';\nimport {a, b} from 'bar.js';"},
			{Code: "import {b, c} from 'bar.js';\nimport A from 'foo.js';"},
			{Code: "import A from 'bar.js';\nimport {b, c} from 'foo.js';", Options: []any{map[string]any{"memberSyntaxSortOrder": []any{"single", "multiple", "none", "all"}}}},
			{Code: "import {a, b} from 'bar.js';\nimport {c, d} from 'foo.js';"},
			{Code: "import A from 'foo.js';\nimport B from 'bar.js';"},
			{Code: "import A from 'foo.js';\nimport a from 'bar.js';"},
			{Code: "import a, * as b from 'foo.js';\nimport c from 'bar.js';"},
			{Code: "import 'foo.js';\n import a from 'bar.js';"},
			{Code: "import B from 'foo.js';\nimport a from 'bar.js';"},
			{Code: "import a from 'foo.js';\nimport B from 'bar.js';", Options: ignoreCase},
			{Code: "import {a, b, c, d} from 'foo.js';"},
			{Code: "import a from 'foo.js';\nimport B from 'bar.js';", Options: []any{map[string]any{"ignoreDeclarationSort": true}}},
			{Code: "import {b, A, C, d} from 'foo.js';", Options: []any{map[string]any{"ignoreMemberSort": true}}},
			{Code: "import {B, a, C, d} from 'foo.js';", Options: []any{map[string]any{"ignoreMemberSort": true}}},
			{Code: "import {a, B, c, D} from 'foo.js';", Options: ignoreCase},
			{Code: "import a, * as b from 'foo.js';"},
			{Code: "import * as a from 'foo.js';\n\nimport b from 'bar.js';"},
			{Code: "import * as bar from 'bar.js';\nimport * as foo from 'foo.js';"},
			// ---- eslint/eslint#5130 ----
			{Code: "import 'foo';\nimport bar from 'bar';", Options: ignoreCase},
			// ---- eslint/eslint#5305 ----
			{Code: "import React, {Component} from 'react';"},
			// ---- allowSeparatedGroups ----
			{Code: "import b from 'b';\n\nimport a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
			{Code: "import a from 'a';\n\nimport 'b';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
			{Code: "import { b } from 'b';\n\n\nimport { a } from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
			{Code: "import b from 'b';\n// comment\nimport a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
			{Code: "import b from 'b';\nfoo();\nimport a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
			{Code: "import { b } from 'b';/*\n comment \n*/import { a } from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
			{Code: "import b from\n'b';\n\nimport\n a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
			{Code: "import c from 'c';\n\nimport a from 'a';\nimport b from 'b';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
			{Code: "import c from 'c';\n\nimport b from 'b';\n\nimport a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "import a from 'foo.js';\nimport A from 'bar.js';", Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 24)}},
			{Code: "import b from 'foo.js';\nimport a from 'bar.js';", Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 24)}},
			{Code: "import {b, c} from 'foo.js';\nimport {a, d} from 'bar.js';", Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 29)}},
			{Code: "import * as foo from 'foo.js';\nimport * as bar from 'bar.js';", Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 31)}},
			{Code: "import a from 'foo.js';\nimport {b, c} from 'bar.js';", Errors: []rule_tester.InvalidTestCaseError{declarationError("unexpectedSyntaxOrder", "Expected 'multiple' syntax before 'single' syntax.", 2, 29)}},
			{Code: "import a from 'foo.js';\nimport * as b from 'bar.js';", Errors: []rule_tester.InvalidTestCaseError{declarationError("unexpectedSyntaxOrder", "Expected 'all' syntax before 'single' syntax.", 2, 29)}},
			{Code: "import a from 'foo.js';\nimport 'bar.js';", Errors: []rule_tester.InvalidTestCaseError{declarationError("unexpectedSyntaxOrder", "Expected 'none' syntax before 'single' syntax.", 2, 17)}},
			{Code: "import b from 'bar.js';\nimport * as a from 'foo.js';", Options: []any{map[string]any{"memberSyntaxSortOrder": []any{"all", "single", "multiple", "none"}}}, Errors: []rule_tester.InvalidTestCaseError{declarationError("unexpectedSyntaxOrder", "Expected 'all' syntax before 'single' syntax.", 2, 29)}},
			{Code: "import {b, a, d, c} from 'foo.js';", Output: []string{"import {a, b, c, d} from 'foo.js';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 12, 13)}},
			{Code: "import {b, a, d, c} from 'foo.js';\nimport {e, f, g, h} from 'bar.js';", Output: []string{"import {a, b, c, d} from 'foo.js';\nimport {e, f, g, h} from 'bar.js';"}, Options: []any{map[string]any{"ignoreDeclarationSort": true}}, Errors: []rule_tester.InvalidTestCaseError{memberError("a", 1, 12, 13)}},
			{Code: "import {a, B, c, D} from 'foo.js';", Output: []string{"import {B, D, a, c} from 'foo.js';"}, Errors: []rule_tester.InvalidTestCaseError{memberError("B", 1, 12, 13)}},
			{Code: "import {zzzzz, /* comment */ aaaaa} from 'foo.js';", Errors: []rule_tester.InvalidTestCaseError{memberError("aaaaa", 1, 30, 35)}},
			{Code: "import {zzzzz /* comment */, aaaaa} from 'foo.js';", Errors: []rule_tester.InvalidTestCaseError{memberError("aaaaa", 1, 30, 35)}},
			{Code: "import {/* comment */ zzzzz, aaaaa} from 'foo.js';", Errors: []rule_tester.InvalidTestCaseError{memberError("aaaaa", 1, 30, 35)}},
			{Code: "import {zzzzz, aaaaa /* comment */} from 'foo.js';", Errors: []rule_tester.InvalidTestCaseError{memberError("aaaaa", 1, 16, 21)}},
			{
				Code:   "\n              import {\n                boop,\n                foo,\n                zoo,\n                baz as qux,\n                bar,\n                beep\n              } from 'foo.js';\n            ",
				Output: []string{"\n              import {\n                bar,\n                beep,\n                boop,\n                foo,\n                baz as qux,\n                zoo\n              } from 'foo.js';\n            "},
				Errors: []rule_tester.InvalidTestCaseError{memberError("qux", 6, 17, 27)},
			},
			// ---- allowSeparatedGroups ----
			{Code: "import b from 'b';\nimport a from 'a';", Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 19)}},
			{Code: "import b from 'b';\n\nimport a from 'a';", Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 3, 19)}},
			{Code: "import b from 'b';\nimport a from 'a';", Options: []any{map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 19)}},
			{Code: "import b from 'b';\nimport a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 19)}},
			{Code: "import b from 'b';import a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortImportsAlphabetically", Message: "Imports should be sorted alphabetically.", Line: 1, Column: 19, EndLine: 1, EndColumn: 37}}},
			{Code: "import b from 'b'; /* comment */ import a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortImportsAlphabetically", Line: 1, Column: 34}}},
			{Code: "import b from 'b'; // comment\nimport a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 2, 19)}},
			{Code: "import b from 'b'; // comment 1\n/* comment 2 */import a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortImportsAlphabetically", Line: 2, Column: 16}}},
			{Code: "import { b } from 'b'; /* comment line 1 \n comment line 2 */ import { a } from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortImportsAlphabetically", Line: 2, Column: 20}}},
			{Code: "import b\nfrom 'b'; import a\nfrom 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortImportsAlphabetically", Line: 2, Column: 11}}},
			{Code: "import { b } from \n'b'; /* comment */ import\n { a } from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortImportsAlphabetically", Line: 2, Column: 20}}},
			{Code: "import { b } from \n'b';\nimport\n { a } from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortImportsAlphabetically", Line: 3, Column: 1}}},
			{Code: "import c from 'c';\n\nimport b from 'b';\nimport a from 'a';", Options: []any{map[string]any{"allowSeparatedGroups": true}}, Errors: []rule_tester.InvalidTestCaseError{declarationError("sortImportsAlphabetically", "Imports should be sorted alphabetically.", 4, 19)}},
			{Code: "import b from 'b';\n\nimport { c, a } from 'c';", Output: []string{"import b from 'b';\n\nimport { a, c } from 'c';"}, Options: []any{map[string]any{"allowSeparatedGroups": true}}, Errors: []rule_tester.InvalidTestCaseError{memberError("a", 3, 13, 14)}},
		},
	)
}
