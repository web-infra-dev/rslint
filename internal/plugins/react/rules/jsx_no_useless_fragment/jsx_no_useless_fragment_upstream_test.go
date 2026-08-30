package jsx_no_useless_fragment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestJsxNoUselessFragmentUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/jsx-no-useless-fragment.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in jsx_no_useless_fragment_extras_test.go.
func TestJsxNoUselessFragmentUpstream(t *testing.T) {
	allowExpressions := []interface{}{map[string]interface{}{"allowExpressions": true}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoUselessFragmentRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `<><Foo /><Bar /></>`, Tsx: true},
		{Code: `<>foo<div /></>`, Tsx: true},
		{Code: `<> <div /></>`, Tsx: true},
		{Code: `<>{"moo"} </>`, Tsx: true},
		{Code: `<NotFragment />`, Tsx: true},
		{Code: `<React.NotFragment />`, Tsx: true},
		{Code: `<NotReact.Fragment />`, Tsx: true},
		{Code: `<Foo><><div /><div /></></Foo>`, Tsx: true},
		{Code: `<div p={<>{"a"}{"b"}</>} />`, Tsx: true},
		{Code: `<Fragment key={item.id}>{item.value}</Fragment>`, Tsx: true},
		{Code: `<Fooo content={<>eeee ee eeeeeee eeeeeeee</>} />`, Tsx: true},
		{Code: `<>{foos.map(foo => foo)}</>`, Tsx: true},
		{Code: `<>{moo}</>`, Tsx: true, Options: allowExpressions},
		{
			Code: `
        <>
          {moo}
        </>
      `,
			Tsx:     true,
			Options: allowExpressions,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `<></>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code: `<>{}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code:   `<p>moo<>foo</></p>`,
			Tsx:    true,
			Output: []string{`<p>moofoo</p>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 7},
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 7},
			},
		},
		{
			Code: `<>{meow}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code:   `<p><>{meow}</></p>`,
			Tsx:    true,
			Output: []string{`<p>{meow}</p>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 4},
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 4},
			},
		},
		{
			Code:   `<><div/></>`,
			Tsx:    true,
			Output: []string{`<div/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code: `
        <>
          <div/>
        </>
      `,
			Tsx: true,
			Output: []string{`
        <div/>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 2, Column: 9},
			},
		},
		{
			Code: `<Fragment />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code: `
        <React.Fragment>
          <Foo />
        </React.Fragment>
      `,
			Tsx: true,
			Output: []string{`
        <Foo />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 2, Column: 9},
			},
		},
		{
			Code: `
        <SomeReact.SomeFragment>
          {foo}
        </SomeReact.SomeFragment>
      `,
			Tsx: true,
			Settings: map[string]interface{}{
				"react": map[string]interface{}{
					"pragma":   "SomeReact",
					"fragment": "SomeFragment",
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 2, Column: 9},
			},
		},
		{
			// Not safe to fix this case because `Eeee` might require child be ReactElement
			Code: `<Eeee><>foo</></Eeee>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 7},
			},
		},
		{
			Code:   `<div><>foo</></div>`,
			Tsx:    true,
			Output: []string{`<div>foo</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 6},
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 6},
			},
		},
		{
			Code:   `<div><>{"a"}{"b"}</></div>`,
			Tsx:    true,
			Output: []string{`<div>{"a"}{"b"}</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 6},
			},
		},
		{
			// SKIP: upstream repeats the previous case for the legacy
			// typescript-eslint parser, which reported no autofix there.
			// rslint has a single parser, so the case above already covers it.
			Skip: true,
			Code: `<div><>{"a"}{"b"}</></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 6},
			},
		},
		{
			Code: `
        <section>
          <Eeee />
          <Eeee />
          <>{"a"}{"b"}</>
        </section>`,
			Tsx: true,
			Output: []string{`
        <section>
          <Eeee />
          <Eeee />
          {"a"}{"b"}
        </section>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 5, Column: 11},
			},
		},
		{
			Code:   `<div><Fragment>{"a"}{"b"}</Fragment></div>`,
			Tsx:    true,
			Output: []string{`<div>{"a"}{"b"}</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 6},
			},
		},
		{
			// whitespace tricky case
			Code: `
        <section>
          git<>
            <b>hub</b>.
          </>

          git<> <b>hub</b></>
        </section>`,
			Tsx: true,
			Output: []string{`
        <section>
          git<b>hub</b>.

          git <b>hub</b>
        </section>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 3, Column: 14},
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 7, Column: 14},
			},
		},
		{
			Code:   `<div>a <>{""}{""}</> a</div>`,
			Tsx:    true,
			Output: []string{`<div>a {""}{""} a</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 8},
			},
		},
		{
			Code: `
        const Comp = () => (
          <html>
            <React.Fragment />
          </html>
        );
      `,
			Tsx: true,
			Output: []string{`
        const Comp = () => (
          <html>
            
          </html>
        );
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 4, Column: 13},
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 4, Column: 13},
			},
		},
		{
			// Ensure allowExpressions still catches expected violations
			Code:    `<><Foo>{moo}</Foo></>`,
			Tsx:     true,
			Options: allowExpressions,
			Output:  []string{`<Foo>{moo}</Foo>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
	})
}
