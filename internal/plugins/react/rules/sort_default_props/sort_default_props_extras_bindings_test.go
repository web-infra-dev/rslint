// TestSortDefaultPropsExtrasBindings covers review regressions beyond the upstream suite.
package sort_default_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortDefaultPropsExtrasBindings(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortDefaultPropsRule, []rule_tester.ValidTestCase{
		// ---- Review regression: Computed method: ignore unrelated body defaults ----
		{Code: `const defaults={a:0,z:0}; const o={[C.defaultProps=defaults](){const defaults={z:0,a:0};}};`, Tsx: true},
		// ---- Review regression: Computed get: ignore unrelated body defaults ----
		{Code: `const defaults={a:0,z:0}; const o={get [C.defaultProps=defaults](){const defaults={z:0,a:0};return null;}};`, Tsx: true},
		// ---- Review regression: Computed set: ignore unrelated body defaults ----
		{Code: `const defaults={a:0,z:0}; const o={set [C.defaultProps=defaults](value){const defaults={z:0,a:0};}};`, Tsx: true},
		// ---- Review regression: Computed async: ignore unrelated body defaults ----
		{Code: `const defaults={a:0,z:0}; const o={async [C.defaultProps=defaults](){const defaults={z:0,a:0};}};`, Tsx: true},
		// ---- Review regression: Computed generator: ignore unrelated body defaults ----
		{Code: `const defaults={a:0,z:0}; const o={*[C.defaultProps=defaults](){const defaults={z:0,a:0};}};`, Tsx: true},
		// ---- Review regression: Computed class: ignore unrelated body defaults ----
		{Code: `const defaults={a:0,z:0}; class D {[C.defaultProps=defaults](){const defaults={z:0,a:0};}}`, Tsx: true},
		// ---- Review regression: Script function before var is not a variable definition ----
		{Code: `function defaults(){} var defaults={z:0,a:0}; C.defaultProps=defaults;`, Tsx: true, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
		// ---- Review regression: Type-only declaration does not supply a defaults object ----
		{Code: `type Defaults={z:number,a:number}; C.defaultProps=Defaults;`, Tsx: true},
		// ---- Review regression: Parameter shadows outer variable ----
		{Code: `const defaults={z:0,a:0}; function f(defaults){C.defaultProps=defaults;}`, Tsx: true},
		// ---- Review regression: Named function self binding shadows outer variable ----
		{Code: `const defaults={z:0,a:0}; const f=function defaults(){C.defaultProps=defaults;};`, Tsx: true},
		// ---- Review regression: Catch binding shadows outer variable ----
		{Code: `const defaults={z:0,a:0}; try{}catch(defaults){C.defaultProps=defaults;}`, Tsx: true},
		// ---- Review regression: Destructuring initializer is not the bound defaults object ----
		{Code: `const {defaults={z:0,a:0}}=props; C.defaultProps=defaults;`, Tsx: true},
		// ---- Review regression: No alias recursion ----
		{Code: `const a=b; const b=a; C.defaultProps=a;`, Tsx: true},
		// ---- Review regression: unrelated namespace exports are not lexical defaults ----
		{Code: `namespace A {export const defaults={z:0,a:0};} namespace B {C.defaultProps=defaults;}`, Tsx: true},
		// ---- Review regression: private namespace variables do not cross declarations ----
		{Code: `namespace A {const defaults={z:0,a:0};} namespace A {C.defaultProps=defaults;}`, Tsx: true},
		// ---- Review regression: unresolved computed key cannot see a same-namespace method body ----
		{Code: `namespace A {const o={[C.defaultProps=defaults](){const defaults={z:0,a:0};}};}`, Tsx: true},
		// ---- Review regression: namespace exports take precedence over outer values ----
		{Code: `const defaults={z:0,a:0}; namespace A {export const defaults={a:0,z:0};} namespace A {C.defaultProps=defaults;}`, Tsx: true},
		// ---- Review regression: parameters still shadow merged namespace exports ----
		{Code: `namespace A {export const defaults={z:0,a:0};} namespace A {function f(defaults){C.defaultProps=defaults;}}`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Review regression: Computed method: resolve outer defaults ----
		{Code: `const defaults={z:0,a:0}; const o={[C.defaultProps=defaults](){const defaults={a:0,z:0};}};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Computed get: resolve outer defaults ----
		{Code: `const defaults={z:0,a:0}; const o={get [C.defaultProps=defaults](){const defaults={a:0,z:0};return null;}};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Computed set: resolve outer defaults ----
		{Code: `const defaults={z:0,a:0}; const o={set [C.defaultProps=defaults](value){const defaults={a:0,z:0};}};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Computed async: resolve outer defaults ----
		{Code: `const defaults={z:0,a:0}; const o={async [C.defaultProps=defaults](){const defaults={a:0,z:0};}};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Computed generator: resolve outer defaults ----
		{Code: `const defaults={z:0,a:0}; const o={*[C.defaultProps=defaults](){const defaults={a:0,z:0};}};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Computed class: resolve outer defaults ----
		{Code: `const defaults={z:0,a:0}; class D {[C.defaultProps=defaults](){const defaults={a:0,z:0};}}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Parameter default resolves outside function body ----
		{Code: `const defaults={z:0,a:0}; function f(x=(C.defaultProps=defaults)){const defaults={a:0,z:0};}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Type alias before value does not hide initializer ----
		{Code: `type Defaults={z:number,a:number}; const Defaults={z:0,a:0}; C.defaultProps=Defaults;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 56, EndLine: 1, EndColumn: 59}}},
		// ---- Review regression: Value before type alias gives the same result ----
		{Code: `const Defaults={z:0,a:0}; type Defaults={z:number,a:number}; C.defaultProps=Defaults;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Interface before value does not hide initializer ----
		{Code: `interface Defaults {z:number;a:number;} const Defaults={z:0,a:0}; C.defaultProps=Defaults;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 61, EndLine: 1, EndColumn: 64}}},
		// ---- Review regression: Inner type-only name does not shadow outer value ----
		{Code: `const defaults={z:0,a:0}; function f(){type defaults={}; C.defaultProps=defaults;}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: Script var before function retains first definition ----
		{Code: `var defaults={z:0,a:0}; function defaults(){} C.defaultProps=defaults;`, Tsx: true, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 19, EndLine: 1, EndColumn: 22}}},
		// ---- Review regression: Later declaration remains visible ----
		{Code: `C.defaultProps=defaults; const defaults={z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 46, EndLine: 1, EndColumn: 49}}},
		// ---- Review regression: Class field and member reference same defaults independently ----
		{Code: `const defaults={z:0,a:0}; class C {static defaultProps=defaults;} C.defaultProps=defaults;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24},
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 21, EndLine: 1, EndColumn: 24}}},
		// ---- Review regression: reopened namespace exports retain defaults initializers ----
		{Code: `namespace A {export const defaults={z:0,a:0};} namespace A {C.defaultProps=defaults;}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 41, EndLine: 1, EndColumn: 44},
		}},
		// ---- Review regression: merged namespace lookup still respects computed method scope ----
		{Code: `namespace A {export const defaults={z:0,a:0};} namespace A {const o={[C.defaultProps=defaults](){const defaults={a:0,z:0};}};}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 41, EndLine: 1, EndColumn: 44},
		}},
		// ---- Review regression: merged exports cannot be hidden by an outer sorted object ----
		{Code: `const defaults={a:0,z:0}; namespace A {export const defaults={z:0,b:0};} namespace A {C.defaultProps=defaults;}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 67, EndLine: 1, EndColumn: 70},
		}},
	})
}
