// TestNoImportTypeSideEffectsExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
package no_import_type_side_effects

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImportTypeSideEffectsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoImportTypeSideEffectsRule, []rule_tester.ValidTestCase{
		// ---- Locks in upstream selector arm "importKind == type": top-level `import type` is excluded by the selector ----
		{Code: `import type { A, B } from 'mod';`},
		{Code: `import type { A as AA } from 'mod';`},
		// ---- Locks in upstream `node.specifiers.length === 0` early-return: pure side-effect import ----
		{Code: `import 'mod';`},
		// ---- Dimension 4: Graceful degradation — empty named imports `{}` (zero specifiers) must not crash or report ----
		// In tsgo this lands here with NamedImports.Elements.Nodes == [], distinct from the no-clause path above.
		{Code: `import {} from 'mod';`},
		// ---- Locks in upstream specifier-walk arm "not ImportSpecifier": default specifier ----
		{Code: `import T from 'mod';`},
		// ---- Locks in upstream specifier-walk arm "not ImportSpecifier": namespace specifier ----
		{Code: `import * as T from 'mod';`},
		// ---- Locks in upstream specifier-walk arm "not ImportSpecifier": default + named (default is not ImportSpecifier) ----
		{Code: `import T, { type U } from 'mod';`},
		{Code: `import T, { U } from 'mod';`},
		// ---- Locks in upstream specifier-walk arm "importKind != type": at least one value specifier ----
		{Code: `import { A } from 'mod';`},
		{Code: `import { A, B } from 'mod';`},
		{Code: `import { type A, B } from 'mod';`},
		{Code: `import { A, type B } from 'mod';`},
		{Code: `import { type A, B, type C } from 'mod';`},
		// ---- Dimension 4: Access / key forms — string-literal `imported` (ES2022 `{ "x" as y }`) without `type` ----
		{Code: `import { "x" as y } from 'mod';`},
		// ---- Dimension 4: Quote-style variance — double-quoted module specifier (parser-normalized but source-text preserved) ----
		{Code: `import { type A, B } from "mod";`},
		// ---- Dimension 4: Declaration / container forms — N/A (rule targets ImportDeclaration only; no function/class wrappers apply) ----
		// ---- Dimension 4: Nesting / traversal boundaries — N/A (imports are top-level statements; no same-kind nesting) ----
		// ---- Dimension 4: Optional chain / non-null / `as` / `satisfies` wrappers — N/A (rule inspects no expression-position child nodes) ----
		// ---- Real-user: mixed type/value default+named where the named portion happens to be only type — still has a value (default), so leaving the import is correct ----
		{Code: `import React, { type FC } from 'react';`},
		// ---- Real-user: side-effect polyfill imports must remain untouched ----
		{Code: `import 'core-js/stable';`},
		// ---- Real-user: `import type` already top-level with multi-specifier — common after running this rule's fix ----
		{Code: `import type { ComponentProps, ReactNode } from 'react';`},
		// ---- Dimension 4 / Nesting boundary: ImportDeclaration nested inside `declare module` block — listener walks tsgo's full tree ----
		// Should NOT fire because the inner import has a mixed specifier (value `Z`).
		{Code: `declare module 'foo' { import { type X, Z } from 'bar'; }`},
		// ---- Dimension 4 / Nesting boundary: nested ImportDeclaration with a top-level `type` qualifier — selector skip still applies inside `declare module` ----
		{Code: `declare module 'foo' { import type { X } from 'bar'; }`},
		// ---- Locks in listener arm "kind != ImportDeclaration": dynamic `import(...)` is a CallExpression, never fires the listener ----
		{Code: `async function load() { return import('mod'); }`},
		// ---- Locks in listener arm "kind != ImportDeclaration": `import X = require('mod')` is a KindImportEqualsDeclaration ----
		{Code: `import X = require('mod');`},
		// ---- Locks in listener arm "kind != ImportDeclaration": `export { type X } from 'mod'` is a KindExportDeclaration ----
		{Code: `export { type X } from 'mod';`},
		{Code: `export type { X } from 'mod';`},
		{Code: `export * from 'mod';`},
		// ---- Dimension 4: import attributes (`with { type: 'json' }`) with a mixed specifier — Attributes child must not be treated as a specifier ----
		{Code: `import { type Schema, parse } from 'mod' with { type: 'json' };`},
		// ---- Dimension 4: `default as Foo` without `type` — `default` is the PropertyName and the specifier carries no IsTypeOnly flag, must remain valid ----
		{Code: `import { default as Foo } from 'mod';`},
		// ---- Idempotency: every fix output (single, multi, alias, multi-line) must itself be a valid input on a re-run ----
		// These exact strings come from the corresponding invalid cases' `Output` and lock the fix point under re-evaluation.
		{Code: `import type { A } from 'mod';`},
		{Code: `import type { A, B } from 'mod';`},
		{Code: `import type { A as AA, B as BB } from 'mod';`},
		{Code: `import type {A} from 'mod';`},
		{Code: `import type {   A   } from 'mod';`},
		{Code: `import type {
  A,
  B,
} from 'mod';`},
		// ---- Real-user: multiple imports in a single file, none type-only — listener fires N times but no diagnostic ----
		{Code: `import { A } from 'm1';
import { B } from 'm2';
import { type C, D } from 'm3';`},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: Access / key forms — string-literal `imported` (`{ type "x" as y }`) ----
		// PropertyName is a StringLiteral; fix must compute `importedStart` from PropertyName, not from name(y).
		{
			Code:   `import { type "x" as y } from 'mod';`,
			Output: []string{`import type { "x" as y } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): no whitespace inside braces — `{type X}` ----
		// Verifies the `type` token range computation doesn't depend on surrounding whitespace.
		{
			Code:   `import {type A} from 'mod';`,
			Output: []string{`import type {A} from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): extra inner whitespace — `{ type   X   }` ----
		// The fix removes from `type` start to imported start; the trailing spaces survive.
		{
			Code:   `import {   type   A   } from 'mod';`,
			Output: []string{`import type {   A   } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): multi-line named imports with trailing comma ----
		{
			Code: `import {
  type A,
  type B,
} from 'mod';`,
			Output: []string{`import type {
  A,
  B,
} from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): comment between `type` and the imported identifier ----
		// Upstream removes [type.range[0], imported.range[0]] which sweeps the comment with it.
		// Locks in matching behavior on the tsgo side.
		{
			Code:   `import { type /* keep? */ A } from 'mod';`,
			Output: []string{`import type { A } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Locks in fix Step 2: insert ` type` after the `import` keyword token, not at node start ----
		// Single specifier with `as` alias.
		{
			Code:   `import { type Foo as Bar } from 'mod';`,
			Output: []string{`import type { Foo as Bar } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Real-user: many specifiers all marked inline `type` — common after a refactor that adds `type` per-line ----
		{
			Code:   `import { type A, type B, type C, type D } from 'mod';`,
			Output: []string{`import type { A, B, C, D } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): trailing comma after the last type specifier ----
		{
			Code:   `import { type A, type B, } from 'mod';`,
			Output: []string{`import type { A, B, } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): trailing comment after the imported identifier — comment must survive (fix only spans `type ... A`) ----
		{
			Code:   `import { type A /* keep */ } from 'mod';`,
			Output: []string{`import type { A /* keep */ } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 4 / Nesting boundary: ImportDeclaration nested inside `declare module` block — fix still emits at the inner declaration's `import` keyword ----
		{
			Code:   `declare module 'foo' { import { type X } from 'bar'; }`,
			Output: []string{`declare module 'foo' { import type { X } from 'bar'; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 24},
			},
		},
		// ---- Dimension 4: `default` keyword as PropertyName — must be allowed on the imported-start computation ----
		// PropertyName is an Identifier whose Text is "default"; `default as Foo` is the only legal way to import the default export by name.
		{
			Code:   `import { type default as Foo } from 'mod';`,
			Output: []string{`import type { default as Foo } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 4: Unicode identifier in imported name — multi-byte position arithmetic in fix range ----
		// If the rule used byte-length math (instead of `imported.Pos()`) the fix would land mid-rune; lock that down.
		{
			Code:   `import { type Λ } from 'mod';`,
			Output: []string{`import type { Λ } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): zero whitespace anywhere — `import{type A}from'mod';` ----
		// The `import` token is immediately followed by `{`; the insertion of ` type` must land between them, not inside the brace block.
		{
			Code:   `import{type A}from'mod';`,
			Output: []string{`import type{A}from'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): newline between `type` and the imported identifier ----
		// SkipTrivia must walk past the `\n` and trailing spaces; the fix collapses `type\n  ` into the empty string,
		// then ` type` is inserted after the `import` keyword.
		{
			Code: `import { type
  A } from 'mod';`,
			Output: []string{`import type { A } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 3 (Autofix boundaries): line comment between `type` and the imported identifier ----
		// `//` runs to the next `\n`; SkipTrivia must consume the comment AND the newline so the imported start lands on `A`.
		{
			Code: `import { type // explainer
  A } from 'mod';`,
			Output: []string{`import type { A } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Dimension 4: import attributes (`with { type: 'json' }`) — Attributes child must not interfere with the fix ----
		// The `import` keyword is at position 0; the Attributes block sits AFTER the module specifier and is structurally separate.
		{
			Code:   `import { type Schema } from 'mod' with { type: 'json' };`,
			Output: []string{`import type { Schema } from 'mod' with { type: 'json' };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Real-user: two adjacent type-only imports — each must produce an independent diagnostic at its own line/column with its own fix ----
		// Locks in that the listener is per-node, not per-file, and that fixes don't pollute one another.
		{
			Code: `import { type A } from 'm1';
import { type B } from 'm2';`,
			Output: []string{`import type { A } from 'm1';
import type { B } from 'm2';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
				{MessageId: "useTopLevelQualifier", Line: 2, Column: 1, EndLine: 2, EndColumn: 29},
			},
		},
		// ---- Real-user: mixed run of imports where only some are type-only — only those fire, the rest stay untouched ----
		{
			Code: `import { A } from 'm1';
import { type B } from 'm2';
import { type C, D } from 'm3';
import { type E } from 'm4';`,
			Output: []string{`import { A } from 'm1';
import type { B } from 'm2';
import { type C, D } from 'm3';
import type { E } from 'm4';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 2, Column: 1},
				{MessageId: "useTopLevelQualifier", Line: 4, Column: 1},
			},
		},
		// ---- Real-user: JSX file (.tsx) — same diagnostic + fix; ensures the rule isn't accidentally gated on file kind ----
		{
			Code:   `import { type FC } from 'react';`,
			Tsx:    true,
			Output: []string{`import type { FC } from 'react';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Real-user: mixed `as` and plain inline type specifiers ----
		{
			Code:   `import { type A, type B as BB } from 'mod';`,
			Output: []string{`import type { A, B as BB } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useTopLevelQualifier", Line: 1, Column: 1},
			},
		},
		// ---- Message text lock-in: every reported diagnostic must carry the exact upstream message string ----
		{
			Code:   `import { type Only } from 'mod';`,
			Output: []string{`import type { Only } from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useTopLevelQualifier",
					Message:   "TypeScript will only remove the inline type specifiers which will leave behind a side effect import at runtime. Convert this to a top-level type qualifier to properly remove the entire import.",
					Line:      1,
					Column:    1,
				},
			},
		},
		// ---- Position lock-in: report range covers the whole ImportDeclaration, EndLine/EndColumn included ----
		{
			Code: `
import {
  type A,
} from 'mod';`,
			Output: []string{`
import type {
  A,
} from 'mod';`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useTopLevelQualifier",
					Line:      2,
					Column:    1,
					EndLine:   4,
					EndColumn: 14,
				},
			},
		},
	})
}
