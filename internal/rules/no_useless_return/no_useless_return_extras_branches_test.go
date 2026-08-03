package no_useless_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUselessReturnExtrasBranches gives every reachable arm of the upstream
// rule source a minimum input, including the arms upstream's own suite never
// reaches, so a future refactor cannot flip one of them silently. Each case
// names the arm it locks in. Every verdict below was taken from ESLint itself.
func TestNoUselessReturnExtrasBranches(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessReturnRule,
		[]rule_tester.ValidTestCase{
			// ---- Locks in upstream ReturnStatement() arm 1: an argument marks the earlier returns used. ----
			{Code: `function f() { if (c) { return; } return 5; }`},
			// ---- Locks in upstream ReturnStatement() arm 2: isInLoop. ----
			{Code: `function f() { while (c) { return; } }`},
			// ---- Locks in upstream ReturnStatement() arm 2: isInLoop. ----
			{Code: `function f() { do { return; } while (c); }`},
			// ---- Locks in upstream ReturnStatement() arm 2: isInLoop. ----
			{Code: `function f() { for (;;) { return; } }`},
			// ---- Locks in upstream ReturnStatement() arm 2: isInLoop. ----
			{Code: `function f() { for (const k in o) { return; } }`},
			// ---- Locks in upstream ReturnStatement() arm 2: isInLoop. ----
			{Code: `function f() { for (const k of o) { return; } }`},
			// ---- Locks in upstream ReturnStatement() arm 2: isInLoop. ----
			{Code: `function f() { for (;;) { if (c) { return; } } }`},
			// ---- isInLoop crosses a class static block, so a loop outside the class still counts. ----
			{Code: `function f() { for (;;) { class K { static { return; } } } }`},
			// ---- Locks in upstream ReturnStatement() arm 3: isInFinally. ----
			{Code: `function f() { try { g(); } finally { return; } }`},
			// ---- Locks in upstream ReturnStatement() arm 3: isInFinally. ----
			{Code: `function f() { try { g(); } finally { if (c) { return; } } }`},
			// ---- Locks in upstream ReturnStatement() arm 4: a return nothing reaches is left alone. ----
			{Code: `function f() { throw e; return; }`},
			// ---- Locks in upstream ReturnStatement() arm 4: a return nothing reaches is left alone. ----
			{Code: `function f() { while (true) { g(); } return; }`},
			// ---- Locks in upstream ReturnStatement() arm 4: a return nothing reaches is left alone. ----
			{Code: `function f() { if (c) return 1; else return 2; return; }`},
			// ---- Locks in upstream ReturnStatement() arm 4: a return nothing reaches is left alone. ----
			{Code: `function f() { return; return; post(); }`},
			// ---- Locks in upstream markReturnStatementsOnCurrentSegmentsAsUsed exclusions: a function declaration is hoisted, so it does not use the return. ----
			{Code: `function f() { return; function h() {} post(); }`},
			// ---- A block does nothing on its own, but the statements in it do. ----
			{Code: `function f() { return; { post(); } }`},
			// ---- A break control reaches carries the return to what the break leaves. ----
			{Code: `function f() { switch (s) { case 1: if (c) { return; } break; case 2: two(); } post(); }`},
			// ---- A break nothing reaches goes to the next segment merely. ----
			{Code: `function f() { switch (s) { case 1: return; break; case 2: two(); } }`},
			// ---- A break nothing reaches goes to the next segment merely. ----
			{Code: `function f() { l: { return; break l; } post(); }`},
			// ---- A break nothing reaches goes to the next segment merely. ----
			{Code: `function f() { switch (s) { case 1: { return; } break; case 2: two(); } }`},
			// ---- A break nothing reaches goes to the next segment merely. ----
			{Code: `function f() { switch (s) { case 1: l: { return; } break; case 2: two(); } }`},
			// ---- A break nothing reaches goes to the next segment merely. ----
			{Code: `function f() { switch (s) { case 1: try { return; } catch (e) {} break; case 2: two(); } }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; post(); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; ;; }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; debugger; }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; var v = 1; }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; class C {} }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; if (c) post(); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; while (c) post(); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; do post(); while (c); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; for (;;) post(); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; for (const k in o) post(); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; for (const k of o) post(); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; switch (s) {} }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; throw e; }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; try {} catch (e) {} }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; l: post(); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; with (o) post(); }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { return; return 5; }`},
			// ---- Every statement kind ESLint does list clears the return. ----
			{Code: `function f() { for (;;) { if (c) { return; } continue; } }`},
			// ---- A namespace body is walked through, so a statement in it does clear the return. ----
			{Code: `if (c) { return; }
namespace N { post(); }`},
			// ---- A namespace body is walked through, so a statement in it does clear the return. ----
			{Code: `if (c) { return; }
namespace N { namespace M { post(); } }`},
			// ---- A namespace body is walked through, so a statement in it does clear the return. ----
			{Code: `if (c) { return; }
namespace N.M { post(); }`},
			// ---- A namespace body is walked through, so a statement in it does clear the return. ----
			{Code: `if (c) { return; }
declare module 'x' { export const z: number; }`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export function h() {}`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export interface I {}`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export type T = 1;`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export enum E {}`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export namespace N {}`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export import q = require('y');`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export default 1;`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export {};`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
export * from 'x';`},
			// ---- An `export` modifier puts the declaration back into an ESTree export node, which does clear it. ----
			{Code: `if (c) { return; }
import 'x';`},
			// ---- Locks in upstream markReturnStatementsOnSegmentAsUsed arm 1: a return inside a `try` block survives everything in the `catch` clause and the `finally` block. ----
			{Code: `function f() { try { return; } finally { throw e; } post(); }`},
			// ---- Past the `try` block the shield is gone: what follows the statement clears the return. ----
			{Code: `function f() { try { return; } catch (e) { post(); } tail(); }`},
			// ---- Past the `try` block the shield is gone: what follows the statement clears the return. ----
			{Code: `function f() { try { if (c) { return; } } catch (e) { post(); } tail(); }`},
			// ---- A return in the `catch` clause is not shielded from the `finally` block. ----
			{Code: `function f() { try {} catch (e) { return; } finally { post(); } }`},
			// ---- Locks in upstream getUselessReturns: a segment reached only through the branch that returned still carries the return. ----
			{Code: `function f() { if (c) { return; } else { alt(); } tail(); }`},
			// ---- Locks in upstream getUselessReturns: a segment reached only through the branch that returned still carries the return. ----
			{Code: `function f() { if (c) { return; } if (d) { post(); } }`},
			// ---- A switch clause falls through into the one after it. ----
			{Code: `function f() { switch (s) { case 1: return; default: post(); } }`},
			// ---- Locks in upstream makeBreak with no matching context: a break with nothing to leave still ends the path. ----
			{Code: `function f() { return; break; tail(); }`},
			// ---- A break out of a `try` still runs the `finally` on its way, unless the return is inside the `try` block it belongs to. ----
			{Code: `function f() { try {} catch (e) { if (c) { return; } break; } finally { g(); } }`},
			// ---- A break out of a `try` still runs the `finally` on its way, unless the return is inside the `try` block it belongs to. ----
			{Code: `function f() { l1: { try {} catch (e) { if (c) { return; } break l1; } finally { g(); } } }`},
			// ---- A break out of a `try` still runs the `finally` on its way, unless the return is inside the `try` block it belongs to. ----
			{Code: `function f() { l1: { if (c) { return; } break l1; } tail(); }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in upstream ReturnStatement() arm 1: an argument marks the earlier returns used. ----
			{
				Code:   `function f() { if (c) { return; } return; }`,
				Output: []string{`function f() { if (c) {  } return; }`, `function f() { if (c) {  }  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
				},
			},
			// ---- isInFinally stops at a function, so a return in a function inside the finally is still reported. ----
			{
				Code:   `function f() { try { g(); } finally { function inner() { return; } } }`,
				Output: []string{`function f() { try { g(); } finally { function inner() {  } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 58, EndLine: 1, EndColumn: 65},
				},
			},
			// ---- Locks in upstream ReturnStatement() arm 4: a return nothing reaches is left alone. ----
			{
				Code:   `function f() { return; return; }`,
				Output: []string{`function f() {  return; }`, `function f() {   }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnCurrentSegmentsAsUsed exclusions: a function declaration is hoisted, so it does not use the return. ----
			{
				Code:   `function f() { return; function h() {} }`,
				Output: []string{`function f() {  function h() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnCurrentSegmentsAsUsed exclusions: a function declaration is hoisted, so it does not use the return. ----
			{
				Code:   `function f() { return; function h() {} function i() {} }`,
				Output: []string{`function f() {  function h() {} function i() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- A block does nothing on its own, but the statements in it do. ----
			{
				Code:   `function f() { return; { function h() {} } }`,
				Output: []string{`function f() {  { function h() {} } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- A break control reaches carries the return to what the break leaves. ----
			{
				Code:   `function f() { switch (s) { case 1: if (c) { return; } break; case 2: two(); } }`,
				Output: []string{`function f() { switch (s) { case 1: if (c) {  } break; case 2: two(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 46, EndLine: 1, EndColumn: 53},
				},
			},
			// ---- A break nothing reaches goes to the next segment merely. ----
			{
				Code: `function f() { switch (s) { case 1: { if (c) return; } break; case 2: two(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 46, EndLine: 1, EndColumn: 53},
				},
			},
			// ---- A break nothing reaches goes to the next segment merely. ----
			{
				Code:   `function f() { switch (s) { case 1: l: { if (c) break l; return; } break; case 2: two(); } }`,
				Output: []string{`function f() { switch (s) { case 1: l: { if (c) break l;  } break; case 2: two(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 58, EndLine: 1, EndColumn: 65},
				},
			},
			// ---- A break nothing reaches goes to the next segment merely. ----
			{
				Code: `function f() { switch (s) { case 1: try { if (c) return; } catch (e) {} break; case 2: two(); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 50, EndLine: 1, EndColumn: 57},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code:   `function f() { return; interface I {} }`,
				Output: []string{`function f() {  interface I {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code:   `function f() { return; type T = 1; }`,
				Output: []string{`function f() {  type T = 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code:   `function f() { return; enum E {} }`,
				Output: []string{`function f() {  enum E {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code:   `function f() { return; declare function q(): void; }`,
				Output: []string{`function f() {  declare function q(): void; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code: `if (c) { return; }
import q = require('y');`,
				Output: []string{`if (c) {  }
import q = require('y');`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code: `if (c) { return; }
export = 1;`,
				Output: []string{`if (c) {  }
export = 1;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code: `if (c) { return; }
declare global {}`,
				Output: []string{`if (c) {  }
declare global {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code: `if (c) { return; }
namespace N {}`,
				Output: []string{`if (c) {  }
namespace N {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code: `if (c) { return; }
namespace N { type T = 1; }`,
				Output: []string{`if (c) {  }
namespace N { type T = 1; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- A TypeScript-only declaration is not one of the ESTree types ESLint lists, so it does not clear the return. ----
			{
				Code: `if (c) { return; }
declare module 'x' {}`,
				Output: []string{`if (c) {  }
declare module 'x' {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnSegmentAsUsed arm 1: a return inside a `try` block survives everything in the `catch` clause and the `finally` block. ----
			{
				Code:   `function f() { try { return; } catch (e) { post(); } }`,
				Output: []string{`function f() { try {  } catch (e) { post(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnSegmentAsUsed arm 1: a return inside a `try` block survives everything in the `catch` clause and the `finally` block. ----
			{
				Code:   `function f() { try { return; } finally { post(); } }`,
				Output: []string{`function f() { try {  } finally { post(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnSegmentAsUsed arm 1: a return inside a `try` block survives everything in the `catch` clause and the `finally` block. ----
			{
				Code:   `function f() { try { return; } catch (e) { return 5; } }`,
				Output: []string{`function f() { try {  } catch (e) { return 5; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnSegmentAsUsed arm 1: a return inside a `try` block survives everything in the `catch` clause and the `finally` block. ----
			{
				Code:   `function f() { try { return; } catch (e) { throw e; } }`,
				Output: []string{`function f() { try {  } catch (e) { throw e; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnSegmentAsUsed arm 1: a return inside a `try` block survives everything in the `catch` clause and the `finally` block. ----
			{
				Code:   `function f() { try { if (c) { return; } } catch (e) { post(); } }`,
				Output: []string{`function f() { try { if (c) {  } } catch (e) { post(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 31, EndLine: 1, EndColumn: 38},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnSegmentAsUsed arm 1: a return inside a `try` block survives everything in the `catch` clause and the `finally` block. ----
			{
				Code:   `function f() { try { return; } catch (e) { try { post(); } catch (e2) {} } }`,
				Output: []string{`function f() { try {  } catch (e) { try { post(); } catch (e2) {} } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Locks in upstream markReturnStatementsOnSegmentAsUsed arm 1: a return inside a `try` block survives everything in the `catch` clause and the `finally` block. ----
			{
				Code:   `function f() { try { try { return; } finally { post(); } } catch (e) {} }`,
				Output: []string{`function f() { try { try {  } finally { post(); } } catch (e) {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 28, EndLine: 1, EndColumn: 35},
				},
			},
			// ---- A return in the `catch` clause is not shielded from the `finally` block. ----
			{
				Code:   `function f() { try {} catch (e) { return; } finally {} }`,
				Output: []string{`function f() { try {} catch (e) {  } finally {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
				},
			},
			// ---- A return in the `catch` clause is not shielded from the `finally` block. ----
			{
				Code:   `function f() { try {} catch (e) { return; } finally { return; post(); } }`,
				Output: []string{`function f() { try {} catch (e) {  } finally { return; post(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
				},
			},
			// ---- A return in the `catch` clause is not shielded from the `finally` block. ----
			{
				Code:   `function f() { try {} catch (e) { return; } }`,
				Output: []string{`function f() { try {} catch (e) {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
				},
			},
			// ---- Locks in upstream getUselessReturns: a segment reached only through the branch that returned still carries the return. ----
			{
				Code:   `function f() { if (c) { return; } else { alt(); } }`,
				Output: []string{`function f() { if (c) {  } else { alt(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- Locks in upstream getUselessReturns: a segment reached only through the branch that returned still carries the return. ----
			{
				Code:   `function f() { if (c) alt(); else { return; } }`,
				Output: []string{`function f() { if (c) alt(); else {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
				},
			},
			// ---- A switch clause falls through into the one after it. ----
			{
				Code:   `function f() { switch (s) { case 1: return; case 2: } }`,
				Output: []string{`function f() { switch (s) { case 1:  case 2: } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
				},
			},
			// ---- A switch clause falls through into the one after it. ----
			{
				Code:   `function f() { switch (s) { default: return; } }`,
				Output: []string{`function f() { switch (s) { default:  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 38, EndLine: 1, EndColumn: 45},
				},
			},
			// ---- Locks in upstream isRemovable: only a statement in a statement list can be dropped. ----
			{
				Code: `function f() { if (c) return; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			// ---- Locks in upstream isRemovable: only a statement in a statement list can be dropped. ----
			{
				Code: `function f() { l: return; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
				},
			},
			// ---- Locks in upstream isRemovable: only a statement in a statement list can be dropped. ----
			{
				Code: `function f() { with (o) return; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- Locks in upstream isRemovable: only a statement in a statement list can be dropped. ----
			{
				Code:   `function f() { switch (s) { case 1: return; } }`,
				Output: []string{`function f() { switch (s) { case 1:  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
				},
			},
			// ---- Locks in upstream isRemovable: only a statement in a statement list can be dropped. ----
			{
				Code:   `class K { static { return; } }`,
				Output: []string{`class K { static {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 20, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- Locks in upstream fix(): a comment inside the return blocks the fix. ----
			{
				Code: `function f() { g(); return /* keep */; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 39},
				},
			},
			// ---- Locks in upstream fix(): a comment inside the return blocks the fix. ----
			{
				Code:   `function f() { g(); return; /* after */ }`,
				Output: []string{`function f() { g();  /* after */ }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
				},
			},
			// ---- Locks in upstream FixTracker.retainEnclosingFunction: returns in separate functions are fixed together, ones in the same function are not. ----
			{
				Code:   `function a() { return; } function b() { return; }`,
				Output: []string{`function a() {  } function b() {  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
				},
			},
			// ---- Locks in upstream FixTracker.retainEnclosingFunction: returns in separate functions are fixed together, ones in the same function are not. ----
			{
				Code:   `class K { m() { return; } n() { return; } }`,
				Output: []string{`class K { m() {  } n() {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 17, EndLine: 1, EndColumn: 24},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 33, EndLine: 1, EndColumn: 40},
				},
			},
			// ---- Locks in upstream FixTracker.retainEnclosingFunction: returns in separate functions are fixed together, ones in the same function are not. ----
			{
				Code: `class K { static { return; } }
class L { static { return; } }`,
				Output: []string{`class K { static {  } }
class L { static { return; } }`, `class K { static {  } }
class L { static {  } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 20, EndLine: 1, EndColumn: 27},
					{MessageId: "unnecessaryReturn", Line: 2, Column: 20, EndLine: 2, EndColumn: 27},
				},
			},
			// ---- Locks in upstream FixTracker.retainEnclosingFunction: returns in separate functions are fixed together, ones in the same function are not. ----
			{
				Code:   `function f() { if (c) { return; } return; }`,
				Output: []string{`function f() { if (c) {  } return; }`, `function f() { if (c) {  }  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
					{MessageId: "unnecessaryReturn", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
				},
			},
			// ---- Locks in upstream makeBreak with no matching context: a break with nothing to leave still ends the path. ----
			{
				Code:   `function f() { if (c) { return; } break; tail(); }`,
				Output: []string{`function f() { if (c) {  } break; tail(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- Locks in upstream makeBreak with no matching context: a break with nothing to leave still ends the path. ----
			{
				Code:   `function f() { if (c) { return; } break l9; tail(); }`,
				Output: []string{`function f() { if (c) {  } break l9; tail(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- A break out of a `try` still runs the `finally` on its way, unless the return is inside the `try` block it belongs to. ----
			{
				Code:   `function f() { try { if (c) { return; } break; } finally { g(); } tail(); }`,
				Output: []string{`function f() { try { if (c) {  } break; } finally { g(); } tail(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 31, EndLine: 1, EndColumn: 38},
				},
			},
			// ---- A break out of a `try` still runs the `finally` on its way, unless the return is inside the `try` block it belongs to. ----
			{
				Code:   `function f() { l1: { try { if (c) { return; } break l1; } finally { g(); } } }`,
				Output: []string{`function f() { l1: { try { if (c) {  } break l1; } finally { g(); } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
				},
			},
			// ---- A break out of a `try` still runs the `finally` on its way, unless the return is inside the `try` block it belongs to. ----
			{
				Code:   `function f() { try {} catch (e) { if (c) { return; } break; } finally { return; } tail(); }`,
				Output: []string{`function f() { try {} catch (e) { if (c) {  } break; } finally { return; } tail(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
				},
			},
			// ---- A break out of a `try` still runs the `finally` on its way, unless the return is inside the `try` block it belongs to. ----
			{
				Code:   `function f() { l1: { if (c) { return; } break l1; } }`,
				Output: []string{`function f() { l1: { if (c) {  } break l1; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 31, EndLine: 1, EndColumn: 38},
				},
			},
			// ---- A return with no semicolon, and one the parser terminates for it. ----
			{
				Code:   `function f() { g(); return }`,
				Output: []string{`function f() { g();  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 21, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- A return with no semicolon, and one the parser terminates for it. ----
			{
				Code: `function f() {
  g()
  return
}`,
				Output: []string{`function f() {
  g()
  
}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
				},
			},
		},
	)
}
