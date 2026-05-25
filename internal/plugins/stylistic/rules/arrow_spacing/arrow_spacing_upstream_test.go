// TestArrowSpacingUpstream migrates the full valid/invalid suite from upstream
// packages/eslint-plugin/rules/arrow-spacing/arrow-spacing.test.ts 1:1.
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in arrow_spacing_extras_test.go.
package arrow_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/arrow_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func optsBA(before, after bool) []any {
	return []any{map[string]any{"before": before, "after": after}}
}
func optsA(after bool) []any { return []any{map[string]any{"after": after}} }
func optsEmpty() []any       { return []any{map[string]any{}} }

func TestArrowSpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&arrow_spacing.ArrowSpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- { before: true, after: true } ----
			{Code: `a => a`, Options: optsBA(true, true)},
			{Code: `() => {}`, Options: optsBA(true, true)},
			{Code: `(a) => {}`, Options: optsBA(true, true)},

			// ---- { before: false, after: true } ----
			{Code: `a=> a`, Options: optsBA(false, true)},
			{Code: `()=> {}`, Options: optsBA(false, true)},
			{Code: `(a)=> {}`, Options: optsBA(false, true)},

			// ---- { before: true, after: false } ----
			{Code: `a =>a`, Options: optsBA(true, false)},
			{Code: `() =>{}`, Options: optsBA(true, false)},
			{Code: `(a) =>{}`, Options: optsBA(true, false)},

			// ---- { before: false, after: false } ----
			{Code: `a=>a`, Options: optsBA(false, false)},
			{Code: `()=>{}`, Options: optsBA(false, false)},
			{Code: `(a)=>{}`, Options: optsBA(false, false)},

			// ---- empty options object — defaults to { before: true, after: true } ----
			{Code: `a => a`, Options: optsEmpty()},
			{Code: `() => {}`, Options: optsEmpty()},
			{Code: `(a) => {}`, Options: optsEmpty()},

			// ---- no options at all — defaults to { before: true, after: true } ----
			{Code: "(a) =>\n{}"},
			{Code: "(a) =>\r\n{}"},
			{Code: "(a) =>\n    0"},

			// ---- TSFunctionType ----
			{Code: `type Foo = () => void`},
			{Code: `type Foo = ()=>void`, Options: optsBA(false, false)},

			// ---- TSConstructorType ----
			{Code: `type T = new () => P`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- { before: true, after: true } ----
			{
				Code:    `a=>a`,
				Output:  []string{`a => a`},
				Options: optsBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 4},
				},
			},
			{
				Code:    `()=>{}`,
				Output:  []string{`() => {}`},
				Options: optsBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 2},
					{MessageId: "expectedAfter", Line: 1, Column: 5},
				},
			},
			{
				Code:    `(a)=>{}`,
				Output:  []string{`(a) => {}`},
				Options: optsBA(true, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 3},
					{MessageId: "expectedAfter", Line: 1, Column: 6},
				},
			},

			// ---- { before: true, after: false } ----
			{
				Code:    `a=> a`,
				Output:  []string{`a =>a`},
				Options: optsBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 1},
					{MessageId: "unexpectedAfter", Line: 1, Column: 5},
				},
			},
			{
				Code:    `()=> {}`,
				Output:  []string{`() =>{}`},
				Options: optsBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 2},
					{MessageId: "unexpectedAfter", Line: 1, Column: 6},
				},
			},
			{
				Code:    `(a)=> {}`,
				Output:  []string{`(a) =>{}`},
				Options: optsBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 3},
					{MessageId: "unexpectedAfter", Line: 1, Column: 7},
				},
			},
			{
				Code:    `a=>  a`,
				Output:  []string{`a =>a`},
				Options: optsBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 1},
					{MessageId: "unexpectedAfter", Line: 1, Column: 6},
				},
			},
			{
				Code:    `()=>  {}`,
				Output:  []string{`() =>{}`},
				Options: optsBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 2},
					{MessageId: "unexpectedAfter", Line: 1, Column: 7},
				},
			},
			{
				Code:    `(a)=>  {}`,
				Output:  []string{`(a) =>{}`},
				Options: optsBA(true, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 3},
					{MessageId: "unexpectedAfter", Line: 1, Column: 8},
				},
			},

			// ---- { before: false, after: true } ----
			{
				Code:    `a =>a`,
				Output:  []string{`a=> a`},
				Options: optsBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 5},
				},
			},
			{
				Code:    `() =>{}`,
				Output:  []string{`()=> {}`},
				Options: optsBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 2},
					{MessageId: "expectedAfter", Line: 1, Column: 6},
				},
			},
			{
				Code:    `(a) =>{}`,
				Output:  []string{`(a)=> {}`},
				Options: optsBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 3},
					{MessageId: "expectedAfter", Line: 1, Column: 7},
				},
			},
			{
				Code:    `a  =>a`,
				Output:  []string{`a=> a`},
				Options: optsBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
					{MessageId: "expectedAfter", Line: 1, Column: 6},
				},
			},
			{
				Code:    `()  =>{}`,
				Output:  []string{`()=> {}`},
				Options: optsBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 2},
					{MessageId: "expectedAfter", Line: 1, Column: 7},
				},
			},
			{
				Code:    `(a)  =>{}`,
				Output:  []string{`(a)=> {}`},
				Options: optsBA(false, true),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 3},
					{MessageId: "expectedAfter", Line: 1, Column: 8},
				},
			},

			// ---- { before: false, after: false } ----
			{
				Code:    `a => a`,
				Output:  []string{`a=>a`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
					{MessageId: "unexpectedAfter", Line: 1, Column: 6},
				},
			},
			{
				Code:    `() => {}`,
				Output:  []string{`()=>{}`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 2},
					{MessageId: "unexpectedAfter", Line: 1, Column: 7},
				},
			},
			{
				Code:    `(a) => {}`,
				Output:  []string{`(a)=>{}`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 3},
					{MessageId: "unexpectedAfter", Line: 1, Column: 8},
				},
			},
			{
				Code:    `a  =>  a`,
				Output:  []string{`a=>a`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 1},
					{MessageId: "unexpectedAfter", Line: 1, Column: 8},
				},
			},
			{
				Code:    `()  =>  {}`,
				Output:  []string{`()=>{}`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 2},
					{MessageId: "unexpectedAfter", Line: 1, Column: 9},
				},
			},
			{
				Code:    `(a)  =>  {}`,
				Output:  []string{`(a)=>{}`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 3},
					{MessageId: "unexpectedAfter", Line: 1, Column: 10},
				},
			},
			{
				Code:    "(a)  =>\n{}",
				Output:  []string{"(a)  =>{}"},
				Options: optsA(false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedAfter", Line: 2, Column: 1},
				},
			},

			// ---- TSFunctionType ----
			{
				Code:   `type Foo = ()=>void`,
				Output: []string{`type Foo = () => void`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 13},
					{MessageId: "expectedAfter", Line: 1, Column: 16},
				},
			},
			{
				Code:    "type Foo = () =>\nvoid",
				Output:  []string{`type Foo = ()=>void`},
				Options: optsBA(false, false),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedBefore", Line: 1, Column: 13},
					{MessageId: "unexpectedAfter", Line: 2, Column: 1},
				},
			},

			// ---- TSConstructorType ----
			{
				Code:   `type T = new ()=>P`,
				Output: []string{`type T = new () => P`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 15},
					{MessageId: "expectedAfter", Line: 1, Column: 18},
				},
			},

			// ---- eslint/eslint#7079: nested arrow inside default-value initializer ----
			{
				Code:   `(a = ()=>0)=>1`,
				Output: []string{`(a = () => 0) => 1`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 7},
					{MessageId: "expectedAfter", Line: 1, Column: 10},
					{MessageId: "expectedBefore", Line: 1, Column: 11},
					{MessageId: "expectedAfter", Line: 1, Column: 14},
				},
			},
			{
				Code:   `(a = ()=>0)=>(1)`,
				Output: []string{`(a = () => 0) => (1)`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectedBefore", Line: 1, Column: 7},
					{MessageId: "expectedAfter", Line: 1, Column: 10},
					{MessageId: "expectedBefore", Line: 1, Column: 11},
					{MessageId: "expectedAfter", Line: 1, Column: 14},
				},
			},
		},
	)
}
