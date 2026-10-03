// TestSortDefaultPropsExtras locks in upstream branches and native AST edge
// shapes not exercised by sort_default_props_upstream_test.go.
// N/A: autofix boundaries and edit demand; this rule supplies no edits.
// N/A: overload function bodies are not inspected by this rule.
package sort_default_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortDefaultPropsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortDefaultPropsRule, []rule_tester.ValidTestCase{
		// ---- Locks in checkSorted: empty/singleton/equal ----
		{Code: `C.defaultProps={}; C.getDefaultProps={a:0}; C.defaultProps={a:0,a:1};`, Tsx: true},
		// ---- Locks in options: true ----
		{Code: `C.defaultProps={a:0,Z:0};`, Tsx: true, Options: map[string]any{"ignoreCase": true}},
		// ---- Locks in checkNode: missing and unsupported values ----
		{Code: `class C { defaultProps; } C.defaultProps; C.defaultProps=null; C.defaultProps=[]; C.defaultProps=make(); C.defaultProps=()=>({z:0,a:0}); C.defaultProps=missing;`, Tsx: true},
		// ---- Locks in checkNode: one initializer only ----
		{Code: `const a={z:0,a:0}; const b=a; C.defaultProps=b;`, Tsx: true},
		// ---- Locks in findVariableByName: imports/functions/parameters ----
		{Code: `import d from "defaults"; C.defaultProps=d; function f(d){C.defaultProps=d;} C.defaultProps=f;`, Tsx: true},
		// ---- Locks in findVariableByName: first uninitialized declaration ----
		{Code: `var d; var d={z:0,a:0}; C.defaultProps=d;`, Tsx: true},
		// ---- Dimension 4: TypeScript wrappers remain opaque ----
		{Code: `C.defaultProps=({z:0,a:0} as any); C.defaultProps=({z:0,a:0} satisfies T); C.defaultProps=defaults!;`, Tsx: true},
		// ---- Dimension 4: optional chain expression boundary ----
		{Code: `C?.defaultProps || {z:0,a:0}; C?.x.defaultProps ?? {z:0,a:0};`, Tsx: true},
		// ---- Dimension 4: literal and dynamic targets ignored ----
		{Code: "C[\"defaultProps\"]={z:0,a:0}; C[`defaultProps`]={z:0,a:0}; C[0]={z:0,a:0}; C[Symbol.iterator]={z:0,a:0}; class D { \"defaultProps\"={z:0,a:0}; [\"defaultProps\"]={z:0,a:0}; }", Tsx: true},
		// ---- Dimension 4: methods and function bodies are not declarations ----
		{Code: `const C=createReactClass({getDefaultProps(){return {z:0,a:0};}}); class D {get defaultProps(){return {z:0,a:0};} getDefaultProps(){return {z:0,a:0};}}`, Tsx: true},
		// ---- Dimension 4: rest binding ----
		{Code: `const {d,...rest}={d:0}; C.defaultProps=rest;`, Tsx: true},
		// ---- Dimension 4: empty/bodyless containers ----
		{Code: `class C {} function f(){} const {}={}; declare class D { defaultProps: object; } abstract class E {abstract getDefaultProps():object;}`, Tsx: true},
		// ---- Real-user: #2178 spread precedence separates sorted groups ----
		{Code: `C.defaultProps={a:"a",c:"c",...foo,b:"b",d:"d"};`, Tsx: true},
		// ---- Locks in MemberExpression: nested assignment and comma ignored ----
		{Code: `C.defaultProps.x={z:0,a:0}; (C.defaultProps,{z:0,a:0});`, Tsx: true},
		// ---- Review: lexical lookup excludes unrelated child scopes ----
		// Upstream reports the unrelated inner object; this identifier is unresolved.
		{Code: `function f(){const defaults={z:0,a:0};} C.defaultProps=defaults;`, Tsx: true},
		// ---- Review: a child scope cannot hide the actual outer defaults ----
		{Code: `const defaults={a:0,z:0}; function f(){function g(){const defaults={z:0,a:0};} C.defaultProps=defaults;}`, Tsx: true},
		// ---- Review: sibling blocks cannot supply a variable initializer ----
		{Code: `{const defaults={z:0,a:0};} {C.defaultProps=defaults;}`, Tsx: true},
		// ---- Review: a parameter binding shadows an outer initializer ----
		{Code: `const defaults={z:0,a:0}; function f(defaults){C.defaultProps=defaults;}`, Tsx: true},
		// ---- Dimension 4: TypeScript auto-accessors are not class fields ----
		{Code: `class C { accessor defaultProps={z:0,a:0}; }`, Tsx: true},
		// ---- Review: destructured variables do not refer to their container ----
		// Upstream incorrectly checks the entire destructuring initializer.
		{Code: `const {d}={z:0,a:0}; C.defaultProps=d;`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Locks in checkSorted: retain maximum after inversions ----
		{Code: `C.defaultProps={z:0,a:0,b:0,y:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 25, EndLine: 1, EndColumn: 28},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 29, EndLine: 1, EndColumn: 32},
		}},
		// ---- Locks in options: omitted/empty/false ----
		{Code: `C.defaultProps={a:0,Z:0};`, Tsx: true, Options: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24},
		}},
		// ---- Locks in options: explicit false ----
		{Code: `C.defaultProps={a:0,Z:0};`, Tsx: true, Options: map[string]any{"ignoreCase": false}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24},
		}},
		// ---- Locks in findVariableByName: first initialized declaration ----
		{Code: `var d={z:0,a:0}; var d={a:0,z:0}; C.defaultProps=d;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 12, EndLine: 1, EndColumn: 15},
		}},
		// ---- Locks in findVariableByName: declaration after use ----
		{Code: `C.defaultProps=d; const d={z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 32, EndLine: 1, EndColumn: 35},
		}},
		// ---- Locks in findVariableByName: reassignment ignored ----
		{Code: `let d={z:0,a:0}; d={a:0,z:0}; C.defaultProps=d;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 12, EndLine: 1, EndColumn: 15},
		}},
		// ---- Dimension 4: parenthesized value and assignment target ----
		{Code: `((C).defaultProps)=(({z:0,a:0}));`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 27, EndLine: 1, EndColumn: 30},
		}},
		// ---- Dimension 4: parenthesized variable initializer ----
		{Code: `const d=(({z:0,a:0})); C.defaultProps=((d));`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 16, EndLine: 1, EndColumn: 19},
		}},
		// ---- Dimension 4: TypeScript receiver wrappers ----
		{Code: `(C as any).defaultProps={z:0,a:0}; C!.defaultProps={z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 30, EndLine: 1, EndColumn: 33},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 57, EndLine: 1, EndColumn: 60},
		}},
		// ---- Dimension 4: computed identifier target ----
		{Code: `C[(defaultProps)]={z:0,a:0}; class C { static [(getDefaultProps)]={z:0,a:0}; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 72, EndLine: 1, EndColumn: 75},
		}},
		// ---- Dimension 4: private field name ----
		{Code: `class C { #defaultProps={z:0,a:0}; f(){this.#defaultProps={z:0,a:0};} }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 30, EndLine: 1, EndColumn: 33},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 64, EndLine: 1, EndColumn: 67},
		}},
		// ---- Dimension 4: raw string/numeric/computed keys ----
		{Code: `C.defaultProps={2:0,10:0,0x0:0,"z":0,"a":0,[z]:0,[(a)]:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 26, EndLine: 1, EndColumn: 31},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 32, EndLine: 1, EndColumn: 37},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 38, EndLine: 1, EndColumn: 43},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 50, EndLine: 1, EndColumn: 57},
		}},
		// ---- Dimension 4: raw escapes and computed expression text ----
		{Code: `C.defaultProps={a:0,\u0062:0}; C.defaultProps={[z /*comment*/ + x]:0, [a + x]:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 29},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 71, EndLine: 1, EndColumn: 80},
		}},
		// ---- Dimension 4: shorthand/accessor/method keys ----
		{Code: `C.defaultProps={z,a,get b(){},set c(v){},async d(){},*e(){},async *f(){}};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 30},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 31, EndLine: 1, EndColumn: 41},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 42, EndLine: 1, EndColumn: 53},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 54, EndLine: 1, EndColumn: 60},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 61, EndLine: 1, EndColumn: 73},
		}},
		// ---- Dimension 4: class expression and instance fields ----
		{Code: `const C=class {defaultProps={z:0,a:0};}; class D { static getDefaultProps={z:0,a:0}; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 34, EndLine: 1, EndColumn: 37},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 80, EndLine: 1, EndColumn: 83},
		}},
		// ---- Dimension 4: same-kind nesting ----
		// The Go harness observes listener order; the public API sorts by location.
		{Code: `class C {static defaultProps={z:class D {defaultProps={z:0,a:0};},a:0};}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 67, EndLine: 1, EndColumn: 70},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 60, EndLine: 1, EndColumn: 63},
		}},
		// ---- Dimension 4: scope shadowing across arrow/method/static block ----
		{Code: `const d={a:0,z:0}; const f=()=>{const d={z:0,a:0}; C.defaultProps=d;}; class D {static {const d={z:0,a:0}; C.defaultProps=d;} m(d){C.defaultProps=d;}} C.defaultProps=d;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 46, EndLine: 1, EndColumn: 49},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 102, EndLine: 1, EndColumn: 105},
		}},
		// ---- Dimension 4: consecutive/leading/trailing spreads ----
		{Code: `C.defaultProps={...x,...y,z:0,a:0,...q,b:0,a:0,...x};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 31, EndLine: 1, EndColumn: 34},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 44, EndLine: 1, EndColumn: 47},
		}},
		// ---- Dimension 4: BOM/CRLF/comments/non-BMP locations ----
		{Code: "\ufeffC.defaultProps = {\r\n  \"😀\": 0,\r\n /* retained */ \"a\": 0\r\n};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 3, Column: 17, EndLine: 3, EndColumn: 23},
		}},
		// ---- Unicode: UTF-16 relational order ----
		{Code: `C.defaultProps={"Ｚ":0,"𐐀":0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 23, EndLine: 1, EndColumn: 29},
		}},
		// ---- Unicode: lower-case expansion ----
		{Code: `C.defaultProps={İ:0,i:0};`, Tsx: true, Options: map[string]any{"ignoreCase": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24},
		}},
		// ---- Unicode: contextual final sigma ----
		{Code: `C.defaultProps={ΟΣ:0,ΟΡ:0};`, Tsx: true, Options: map[string]any{"ignoreCase": true}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
		}},
		// ---- Real-user: #2178 sorting still applies after a spread ----
		{Code: `C.defaultProps={a:"a",...foo,c:"c",b:"b",d:"d"};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 36, EndLine: 1, EndColumn: 41},
		}},
		// ---- Real-user: #2347 no fix offered for out-of-order defaults ----
		{Code: `C.defaultProps={z:loadZ(),a:loadA()};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 27, EndLine: 1, EndColumn: 36},
		}},
		// ---- Locks in MemberExpression: getDefaultProps assignment ----
		{Code: `C.getDefaultProps={z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
		}},
		// ---- Locks in MemberExpression: comparison and logical parents ----
		{Code: `C.defaultProps === {z:0,a:0}; C.defaultProps || {z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 25, EndLine: 1, EndColumn: 28},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 54, EndLine: 1, EndColumn: 57},
		}},
	})
}
