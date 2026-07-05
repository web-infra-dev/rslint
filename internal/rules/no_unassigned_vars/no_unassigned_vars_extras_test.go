package no_unassigned_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnassignedVarsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / real-user shape it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
func TestNoUnassignedVarsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnassignedVarsRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream shouldSkip arm 1: initializer means the variable is assigned.
			{Code: `let assigned = undefined; log(assigned);`},
			// Locks in upstream shouldSkip arm 2: non-Identifier binding patterns are ignored.
			{Code: `let [item] = items; log(item);`},
			// Locks in upstream shouldSkip arm 3: const declarations are ignored.
			{Code: `const value = undefined; log(value);`},
			// Locks in upstream reference loop arm 1: any write reference suppresses the report.
			{Code: `let value; value = compute(); log(value);`},
			// Locks in upstream reference loop arm 1: a later write suppresses an earlier read.
			{Code: `let value; log(value); value = compute();`},
			// Locks in upstream reference loop arm 1: writes inside nested functions still count as writes.
			{Code: `let value; function init() { value = compute(); } log(value);`},
			// Locks in upstream reference loop arm 1: update expressions are writes.
			{Code: `let value; value++; log(value);`},
			// Locks in upstream reference loop arm 1: compound assignments are writes.
			{Code: `let value; value += 1; log(value);`},
			// Locks in upstream reference loop arm 1: logical assignments are writes.
			{Code: `let value; value ||= compute(); log(value);`},
			// Locks in upstream no-read arm: no overlap with no-unused-vars.
			{Code: `let unread;`},
			// Locks in upstream declaration.declare arm: ambient variables are declarations, not unassigned locals.
			{Code: `declare let ambient: string; log(ambient);`},
			// Locks in upstream insideDeclareModule arm: nested namespace contents stay ambient.
			{Code: `declare namespace Lib { namespace Nested { let value: string; export { value }; } }`},
			// Locks in type-only reference filtering: type queries do not read the runtime variable.
			{Code: `let value; type Alias = typeof value;`},
			// Locks in type-only export filtering: export type specifiers are not runtime reads.
			{Code: `let value; export type { value };`},
			// Locks in re-export filtering: re-export specifiers do not read same-named locals.
			{Code: `let value; export { value } from "mod";`},

			// ---- Dimension 4: declaration/container forms ----
			{Code: `for (let item of items) { log(item); }`},
			// ---- Dimension 4: declaration/container forms ----
			{Code: `for (var key in object) { log(key); }`},
			// ---- Dimension 4: assignment target forms ----
			{Code: `let value; [value] = items; log(value);`},
			// ---- Dimension 4: assignment target forms ----
			{Code: `let value; ({ value } = object); log(value);`},
			// ---- Dimension 4: assignment target forms ----
			{Code: `let value; ({ nested: { value } } = object); log(value);`},
			// ---- Dimension 4: assignment target forms ----
			{Code: `let rest; ({ ...rest } = object); log(rest);`},
			// ---- Dimension 4: assignment target forms ----
			{Code: `let tail; [, ...tail] = items; log(tail);`},
			// ---- Dimension 4: assignment target forms ----
			{Code: `let value; ({ value = fallback } = object); log(value);`},
			// ---- Dimension 4: assignment target forms ----
			{Code: `let value; for (value of items) { } log(value);`},
			// ---- Dimension 4: assignment target forms ----
			{Code: `let value; for (value in object) { } log(value);`},
			// ---- Dimension 4: parenthesized assignment targets ----
			{Code: `let value; ((value)) = compute(); log(value);`},
			// ---- Dimension 4: nesting/traversal boundary ----
			{Code: `let outer; function f() { let outer = 1; return outer; } outer = 2; log(outer);`},
			// ---- Dimension 4: nesting/traversal boundary ----
			{Code: `let status;
const f = function status() {
  log(status);
};`},
			// ---- Dimension 4: nesting/traversal boundary ----
			{Code: `let status;
function f() {
  if (condition) {
    var status = "ready";
  }
  log(status);
}`},
			// ---- Real-user: callback ref writes before later reads ----
			{Code: `let input; const ref = (element: HTMLInputElement) => { input = element; }; mount(ref); log(input);`},
			// ---- Real-user: ambient global augmentation remains declaration-only ----
			{Code: `declare global { namespace App { let value: string; export { value }; } }`},
			// N/A: autofix boundaries do not apply; no-unassigned-vars has no autofix.
			// N/A: overload, abstract, and declare members do not create var/let declarators.
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream report arm: read with zero writes reports exact message text.
			{
				Code: `let value; log(value);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// Locks in upstream reference loop arm 2: declaration-only references are not reads.
			{
				Code: `let value; export { value };`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// Locks in export local alias arm: exported local specifiers are runtime reads.
			{
				Code: `let value; export { value as alias };`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// Locks in reference identity: a catch binding shadows instead of writing the outer variable.
			{
				Code: `let error; try {} catch (error) {} log(error);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("error", 1, 5, 1, 10),
				},
			},
			// Locks in reference identity: block-scoped shadowing does not write the outer variable.
			{
				Code: `let value; { let value = 1; log(value); } log(value);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// Locks in reference identity: a hoisted var in a nested function shadows the outer variable.
			{
				Code: `let status;
function f() {
  if (condition) {
    var status;
  }
  log(status);
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("status", 4, 9, 4, 15),
				},
			},
			// ---- Dimension 4: parenthesized read wrappers ----
			{
				Code: `let value; log(((value)));`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// ---- Dimension 4: TS type-expression wrappers ----
			{
				Code: `let value: string | undefined; log((value as any)!.trim());`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 30),
				},
			},
			// ---- Dimension 4: optional chain reads ----
			{
				Code: `let config; config?.enabled;`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("config", 1, 5, 1, 11),
				},
			},
			// ---- Dimension 4: optional chain reads ----
			{
				Code: `let fn; fn?.();`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("fn", 1, 5, 1, 7),
				},
			},
			// ---- Dimension 4: element access reads ----
			{
				Code: `let obj; obj["x"] = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("obj", 1, 5, 1, 8),
				},
			},
			// ---- Dimension 4: computed property key reads ----
			{
				Code: `let key; const obj = { [key]: 1 };`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("key", 1, 5, 1, 8),
				},
			},
			// ---- Dimension 4: computed class keys ----
			{
				Code: `let key; class C { [key]() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("key", 1, 5, 1, 8),
				},
			},
			// ---- Dimension 4: shorthand property reads ----
			{
				Code: `let obj; consume({ obj });`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("obj", 1, 5, 1, 8),
				},
			},
			// ---- Dimension 4: multi-line position in a function container ----
			{
				Code: `function load() {
	let options: { debug?: boolean };
	return options?.debug;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("options", 2, 6, 2, 34),
				},
			},
			// ---- Dimension 4: nesting/traversal boundary ----
			{
				Code: `let outer; function f() { return () => outer; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("outer", 1, 5, 1, 10),
				},
			},
			// ---- Dimension 4: class static block reads ----
			{
				Code: `let value; class C { static { value; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// ---- Dimension 4: class heritage reads ----
			{
				Code: `let Base; class Derived extends Base {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("Base", 1, 5, 1, 9),
				},
			},
			// ---- Dimension 4: class field initializer reads ----
			{
				Code: `let value; class C { field = value; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// ---- Dimension 4: enum initializer reads ----
			{
				Code: `let value; enum E { A = value }`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// ---- Dimension 4: tagged template reads ----
			{
				Code: "let tag; tag`value`;",
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("tag", 1, 5, 1, 8),
				},
			},
			// ---- Dimension 4: export default expression reads ----
			{
				Code: `let value; export default value;`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// ---- Dimension 4: TS satisfies expression reads ----
			{
				Code: `let value; value satisfies string;`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 1, 5, 1, 10),
				},
			},
			// ---- Dimension 4: TSX expression reads ----
			{
				Code: `declare namespace JSX { interface IntrinsicElements { div: any } }
let value;
const element = <div data-value={value} />;`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("value", 2, 5, 2, 10),
				},
			},
			// ---- Real-user: eslint/eslint#12497 property writes read the receiver but do not assign it ----
			{
				Code: `var obj; obj.x = 1; obj.y = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("obj", 1, 5, 1, 8),
				},
			},
			// ---- Real-user: eslint/eslint#12497 assigned siblings do not hide an unassigned variable ----
			{
				Code: `let a, b, c; if (something) { a = 0; b = false; } else { a = 1; b = true; } f(a, b, c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("c", 1, 11, 1, 12),
				},
			},
			// ---- Real-user: eslint/eslint#20169 SolidJS refs are reads unless the variable is written ----
			{
				Code: `declare function onMount(cb: () => void): void;
declare namespace JSX { interface IntrinsicElements { input: any } }
export function Input() {
	let inputRef: HTMLInputElement | undefined;
	onMount(() => inputRef?.select());
	return <input ref={inputRef} />;
}`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					unassignedError("inputRef", 4, 6, 4, 44),
				},
			},
		},
	)
}
