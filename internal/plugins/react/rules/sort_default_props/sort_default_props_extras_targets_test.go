// TestSortDefaultPropsExtrasTargets covers review regressions beyond the upstream suite.
package sort_default_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortDefaultPropsExtrasTargets(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortDefaultPropsRule, []rule_tester.ValidTestCase{
		// ---- Review regression: Sorted static member C["defaultProps"] ----
		{Code: `C["defaultProps"]={a:0,z:0};`, Tsx: true},
		// ---- Review regression: Sorted static member C[`defaultProps`] ----
		{Code: "C[`defaultProps`]={a:0,z:0};", Tsx: true},
		// ---- Review regression: Sorted static member C[("defaultProps")] ----
		{Code: `C[("defaultProps")]={a:0,z:0};`, Tsx: true},
		// ---- Review regression: Sorted static member C["getDefaultProps"] ----
		{Code: `C["getDefaultProps"]={a:0,z:0};`, Tsx: true},
		// ---- Review regression: Computed identifier name is not its value ----
		{Code: `const defaultProps="other"; C[defaultProps]={z:0,a:0};`, Tsx: true},
		// ---- Review regression: Computed class identifier name is not its value ----
		{Code: `const defaultProps="other"; class C {[defaultProps]={z:0,a:0};}`, Tsx: true},
		// ---- Review regression: Unknown computed name stays unknown ----
		{Code: `const prop=getName(); C[prop]={z:0,a:0};`, Tsx: true},
		// ---- Review regression: Constant identifier keys are not evaluated ----
		{Code: `const prop="defaultProps"; C[prop]={z:0,a:0};`, Tsx: true},
		// ---- Review regression: Comparison is not a defaults declaration ----
		{Code: `const equal=C.defaultProps === {z:0,a:0};`, Tsx: true},
		// ---- Review regression: Conditional style object is not defaults ----
		{Code: `const style=C.defaultProps && {zIndex:1,background:"red"};`, Tsx: true},
		// ---- Review regression: For-in enumeration source is not defaults ----
		{Code: `for(C.defaultProps in {z:0,a:0}){}`, Tsx: true},
		// ---- Review regression: For-of iteration source is not defaults ----
		{Code: `for(C.defaultProps of {z:0,a:0}){}`, Tsx: true},
		// ---- Review regression: Arithmetic assignment cannot store RHS object ----
		{Code: `C.defaultProps += {z:0,a:0}; C.defaultProps *= {z:0,a:0};`, Tsx: true},
		// ---- Review regression: Optional reads do not declare defaults ----
		{Code: `C?.defaultProps || {z:0,a:0};`, Tsx: true},
		// ---- Review regression: Private fields do not name public defaults ----
		{Code: `class C {#defaultProps={z:0,a:0}; m(){this.#defaultProps={z:0,a:0};}}`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Review regression: Static member C["defaultProps"] ----
		{Code: `C["defaultProps"]={z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 24, EndLine: 1, EndColumn: 27}}},
		// ---- Review regression: Static member C[`defaultProps`] ----
		{Code: "C[`defaultProps`]={z:0,a:0};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 24, EndLine: 1, EndColumn: 27}}},
		// ---- Review regression: Static member C[("defaultProps")] ----
		{Code: `C[("defaultProps")]={z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 26, EndLine: 1, EndColumn: 29}}},
		// ---- Review regression: Static member C["getDefaultProps"] ----
		{Code: `C["getDefaultProps"]={z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 27, EndLine: 1, EndColumn: 30}}},
		// ---- Review regression: Static class field "defaultProps" ----
		{Code: `class C {static "defaultProps"={z:0,a:0};}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 37, EndLine: 1, EndColumn: 40}}},
		// ---- Review regression: Static class field ["defaultProps"] ----
		{Code: `class C {static ["defaultProps"]={z:0,a:0};}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 39, EndLine: 1, EndColumn: 42}}},
		// ---- Review regression: Static class field [`defaultProps`] ----
		{Code: "class C {static [`defaultProps`]={z:0,a:0};}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 39, EndLine: 1, EndColumn: 42}}},
		// ---- Review regression: Static class field [("getDefaultProps")] ----
		{Code: `class C {static [("getDefaultProps")]={z:0,a:0};}`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 44, EndLine: 1, EndColumn: 47}}},
		// ---- Review regression: Object assignment = ----
		{Code: `C.defaultProps = {z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 23, EndLine: 1, EndColumn: 26}}},
		// ---- Review regression: Object assignment ||= ----
		{Code: `C.defaultProps ||= {z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 25, EndLine: 1, EndColumn: 28}}},
		// ---- Review regression: Object assignment &&= ----
		{Code: `C.defaultProps &&= {z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 25, EndLine: 1, EndColumn: 28}}},
		// ---- Review regression: Object assignment ??= ----
		{Code: `C.defaultProps ??= {z:0,a:0};`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 25, EndLine: 1, EndColumn: 28}}},
		// ---- Review regression: Assignment pattern really assigns defaults ----
		{Code: `({x:C.defaultProps={z:0,a:0}}=source);`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "propsNotSorted", Message: "Default prop types declarations should be sorted alphabetically", Line: 1, Column: 25, EndLine: 1, EndColumn: 28}}},
	})
}
