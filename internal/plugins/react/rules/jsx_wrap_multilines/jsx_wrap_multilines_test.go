package jsx_wrap_multilines

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxWrapMultilinesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxWrapMultilinesRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div />`,
			Tsx:  true,
		},
		{
			Code: `var x = (<div>
  <span />
</div>)`,
			Tsx: true,
		},
		{
			Code: `function foo() {
  return (<div>
    <span />
  </div>)
}`,
			Tsx: true,
		},
		{
			Code: `var x = <div />`,
			Tsx:  true,
		},
		{
			// Arrow body with parens is valid
			Code: `var f = () => (<div>
  <span />
</div>)`,
			Tsx: true,
		},
		{
			// Logical expression: JSX on left side should NOT be checked (ESLint only checks right)
			Code: `var x = <div>
  <span />
</div> && flag`,
			Tsx:     true,
			Options: map[string]interface{}{"logical": "parens"},
		},
		{
			// Logical expression: wrapped right side is valid
			Code: `var x = flag && (<div>
  <span />
</div>)`,
			Tsx:     true,
			Options: map[string]interface{}{"logical": "parens"},
		},
		{
			// Condition in declaration: when condition is "ignore" (default), JSX in ternary branches should be checked with "declaration" setting
			Code: `var x = (true
  ? (<div>
      <span />
    </div>)
  : null)`,
			Tsx: true,
		},
		{
			// "never" mode: unwrapped multiline JSX in declaration is valid
			Code: `var x = <div>
  <span />
</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"declaration": "never"},
		},
		{
			// "parens-new-line" mode: parens on separate lines is valid
			Code: `var x = (
<div>
  <span />
</div>
)`,
			Tsx:     true,
			Options: map[string]interface{}{"declaration": "parens-new-line"},
		},
		{
			// "prop" context: wrapped JSX prop is valid
			Code: `var x = <Foo bar={(<div>
  <span />
</div>)} />`,
			Tsx:     true,
			Options: map[string]interface{}{"prop": "parens", "declaration": "ignore"},
		},
		// --- Fragment context ---
		{
			// Fragment with parens is valid
			Code: `var x = (<>
  <span />
</>)`,
			Tsx: true,
		},
		{
			// Fragment in return with parens is valid
			Code: `function foo() {
  return (<>
    <span />
  </>)
}`,
			Tsx: true,
		},
		// --- Assignment context ---
		{
			Code: `var x; x = <div />`,
			Tsx:  true,
		},
		{
			Code: `var x; x = (<div>
  <span />
</div>)`,
			Tsx: true,
		},
		{
			// Assignment without parens with "ignore" option
			Code: `var x; x = <div>
  <span />
</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"assignment": "ignore"},
		},
		// --- Condition context ---
		{
			// Condition "ignore" (default): no error for unwrapped condition
			Code: `var x = true ? <div>
  <span />
</div> : null`,
			Tsx:     true,
			Options: map[string]interface{}{"declaration": "ignore"},
		},
		{
			// Condition "parens": wrapped is valid
			Code: `var x = true ? (<div>
  <span />
</div>) : null`,
			Tsx:     true,
			Options: map[string]interface{}{"condition": "parens", "declaration": "ignore"},
		},
		{
			// Condition "parens-new-line": valid
			Code: `var x = true ? (
<div>
  <span />
</div>
) : null`,
			Tsx:     true,
			Options: map[string]interface{}{"condition": "parens-new-line", "declaration": "ignore"},
		},
		// --- Logical context ---
		{
			// Single-line logical: no wrapping needed
			Code: `var x = flag && <div />`,
			Tsx:     true,
			Options: map[string]interface{}{"logical": "parens"},
		},
		{
			// Nullish coalescing: wrapped right side is valid
			Code: `var x = flag ?? (<div>
  <span />
</div>)`,
			Tsx:     true,
			Options: map[string]interface{}{"logical": "parens"},
		},
		// --- Return single line ---
		{
			Code: `function foo() { return <div /> }`,
			Tsx:  true,
		},
		// --- All never ---
		{
			Code: `function foo() {
  return <div>
    <span />
  </div>
}`,
			Tsx: true,
			Options: map[string]interface{}{
				"declaration": "never",
				"assignment":  "never",
				"return":      "never",
				"arrow":       "never",
				"condition":   "never",
				"logical":     "never",
				"prop":        "never",
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			// Fragment without parens in declaration should be reported
			Code: `var x = <>
  <span />
</>`,
			Tsx: true,
			Output: []string{`var x = (<>
  <span />
</>)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			// Fragment without parens in return should be reported
			Code: `function foo() {
  return <>
    <span />
  </>
}`,
			Tsx: true,
			Output: []string{`function foo() {
  return (<>
    <span />
  </>)
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      2,
					Column:    10,
				},
			},
		},
		{
			// Condition in declaration: unwrapped multiline JSX in ternary should be reported
			Code: `var x = (true
  ? <div>
      <span />
    </div>
  : null)`,
			Tsx: true,
			Output: []string{`var x = (true
  ? (<div>
      <span />
    </div>)
  : null)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      2,
					Column:    5,
				},
			},
		},
		{
			// Arrow body without parens should be reported
			Code: `var f = () => <div>
  <span />
</div>`,
			Tsx: true,
			Output: []string{`var f = () => (<div>
  <span />
</div>)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    15,
				},
			},
		},
		{
			// Logical expression: unwrapped right side should be reported
			Code: `var x = flag && <div>
  <span />
</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"logical": "parens"},
			Output: []string{`var x = flag && (<div>
  <span />
</div>)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    17,
				},
			},
		},
		{
			Code: `var x = <div>
  <span />
</div>`,
			Tsx: true,
			Output: []string{`var x = (<div>
  <span />
</div>)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			Code: `function foo() {
  return <div>
    <span />
  </div>
}`,
			Tsx: true,
			Output: []string{`function foo() {
  return (<div>
    <span />
  </div>)
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      2,
					Column:    10,
				},
			},
		},
		{
			// "never" mode: wrapped multiline JSX should be reported
			Code: `var x = (<div>
  <span />
</div>)`,
			Tsx:     true,
			Options: map[string]interface{}{"declaration": "never"},
			Output: []string{`var x = <div>
  <span />
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "extraParens",
					Line:      1,
					Column:    10,
				},
			},
		},
		{
			// "parens-new-line" mode: parens on same line should be reported
			Code: `var x = (<div>
  <span />
</div>)`,
			Tsx:     true,
			Options: map[string]interface{}{"declaration": "parens-new-line"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "parensOnNewLines",
					Line:      1,
					Column:    10,
				},
			},
		},
		{
			// "parens-new-line" mode: no parens at all should be reported
			Code: `var x = <div>
  <span />
</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"declaration": "parens-new-line"},
			Output: []string{`var x = (
<div>
  <span />
</div>
)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			// "prop" context: unwrapped JSX prop should be reported
			Code: `var x = <Foo bar={<div>
  <span />
</div>} />`,
			Tsx:     true,
			Options: map[string]interface{}{"prop": "parens", "declaration": "ignore"},
			Output: []string{`var x = <Foo bar={(<div>
  <span />
</div>)} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    19,
				},
			},
		},
		// --- Assignment context ---
		{
			// Assignment without parens should be reported
			Code: `var x; x = <div>
  <span />
</div>`,
			Tsx: true,
			Output: []string{`var x; x = (<div>
  <span />
</div>)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			// Assignment "never": wrapped should be reported
			Code: `var x; x = (<div>
  <span />
</div>)`,
			Tsx:     true,
			Options: map[string]interface{}{"assignment": "never"},
			Output: []string{`var x; x = <div>
  <span />
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "extraParens",
					Line:      1,
					Column:    13,
				},
			},
		},
		// --- Condition context ---
		{
			// Condition "parens": unwrapped should be reported
			Code: `var x = true ? <div>
  <span />
</div> : null`,
			Tsx:     true,
			Options: map[string]interface{}{"condition": "parens", "declaration": "ignore"},
			Output: []string{`var x = true ? (<div>
  <span />
</div>) : null`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    16,
				},
			},
		},
		// --- Return "never": wrapped should be reported ---
		{
			Code: `function foo() {
  return (<div>
    <span />
  </div>)
}`,
			Tsx:     true,
			Options: map[string]interface{}{"return": "never"},
			Output: []string{`function foo() {
  return <div>
    <span />
  </div>
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "extraParens",
					Line:      2,
					Column:    11,
				},
			},
		},
		// --- Arrow "never": wrapped should be reported ---
		{
			Code: `var f = () => (<div>
  <span />
</div>)`,
			Tsx:     true,
			Options: map[string]interface{}{"arrow": "never"},
			Output: []string{`var f = () => <div>
  <span />
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "extraParens",
					Line:      1,
					Column:    16,
				},
			},
		},
		// --- Return "parens-new-line" ---
		{
			// Return without any parens: missingParens
			Code: `function foo() {
  return <div>
    <span />
  </div>
}`,
			Tsx:     true,
			Options: map[string]interface{}{"return": "parens-new-line"},
			Output: []string{`function foo() {
  return (
  <div>
    <span />
  </div>
  )
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      2,
					Column:    10,
				},
			},
		},
		{
			// Return with parens on same line: parensOnNewLines
			Code: `function foo() {
  return (<div>
    <span />
  </div>)
}`,
			Tsx:     true,
			Options: map[string]interface{}{"return": "parens-new-line"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "parensOnNewLines",
					Line:      2,
					Column:    11,
				},
			},
		},
		// --- Arrow "parens-new-line" ---
		{
			Code: `var f = () => <div>
  <span />
</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"arrow": "parens-new-line"},
			Output: []string{`var f = () => (
<div>
  <span />
</div>
)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    15,
				},
			},
		},
		// --- Nullish coalescing: unwrapped right side ---
		{
			Code: `var x = flag ?? <div>
  <span />
</div>`,
			Tsx:     true,
			Options: map[string]interface{}{"logical": "parens"},
			Output: []string{`var x = flag ?? (<div>
  <span />
</div>)`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingParens",
					Line:      1,
					Column:    17,
				},
			},
		},
		// --- Logical "never": wrapped should be reported ---
		{
			Code: `var x = flag && (<div>
  <span />
</div>)`,
			Tsx:     true,
			Options: map[string]interface{}{"logical": "never"},
			Output: []string{`var x = flag && <div>
  <span />
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "extraParens",
					Line:      1,
					Column:    18,
				},
			},
		},
		// --- Prop "never": wrapped should be reported ---
		{
			Code: `var x = <Foo bar={(<div>
  <span />
</div>)} />`,
			Tsx:     true,
			Options: map[string]interface{}{"prop": "never", "declaration": "ignore"},
			Output: []string{`var x = <Foo bar={<div>
  <span />
</div>} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "extraParens",
					Line:      1,
					Column:    20,
				},
			},
		},
	})
}
