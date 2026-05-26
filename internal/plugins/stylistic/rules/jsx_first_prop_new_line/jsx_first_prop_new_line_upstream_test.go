// TestJsxFirstPropNewLineUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/jsx-first-prop-new-line/
// jsx-first-prop-new-line.test.ts 1:1. Upstream asserts only messageId on its
// invalid cases; the Line/Column here point at the first prop (the report
// node), computed from the exact source each case carries, and Message asserts
// the exact diagnostic text. rslint-specific lock-in cases live in
// jsx_first_prop_new_line_extras_test.go.
package jsx_first_prop_new_line

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	msgPropOnNewLine  = "Property should be placed on a new line"
	msgPropOnSameLine = "Property should be placed on the same line as the component declaration"
)

func TestJsxFirstPropNewLineUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxFirstPropNewLineRule, []rule_tester.ValidTestCase{
		// ---- 'never' ----
		{Code: `<Foo />`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<Foo prop="bar" />`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<Foo {...this.props} />`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<Foo a a a />`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `
        <Foo a
          b
        />
      `, Tsx: true, Options: []interface{}{"never"}},
		// ---- 'multiline' ----
		{Code: `<Foo />`, Tsx: true, Options: []interface{}{"multiline"}},
		{Code: `<Foo prop="one" />`, Tsx: true, Options: []interface{}{"multiline"}},
		{Code: `<Foo {...this.props} />`, Tsx: true, Options: []interface{}{"multiline"}},
		{Code: `<Foo a a a />`, Tsx: true, Options: []interface{}{"multiline"}},
		{Code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `, Tsx: true, Options: []interface{}{"multiline"}},
		{Code: `
        <Foo
          {...this.props}
          propTwo="two"
        />
      `, Tsx: true, Options: []interface{}{"multiline"}},
		// ---- 'multiline-multiprop' (default) ----
		{Code: `
        <Foo bar />
      `, Tsx: true, Options: []interface{}{"multiline-multiprop"}},
		{Code: `
        <Foo bar baz />
      `, Tsx: true, Options: []interface{}{"multiline-multiprop"}},
		{Code: `
        <Foo prop={{
        }} />
      `, Tsx: true, Options: []interface{}{"multiline-multiprop"}},
		{Code: `
        <Foo
          foo={{
          }}
          bar
        />
      `, Tsx: true, Options: []interface{}{"multiline-multiprop"}},
		// ---- 'always' ----
		{Code: `<Foo />`, Tsx: true, Options: []interface{}{"always"}},
		{Code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `, Tsx: true, Options: []interface{}{"always"}},
		{Code: `
        <Foo
          {...this.props}
          propTwo="two"
        />
      `, Tsx: true, Options: []interface{}{"always"}},
		// ---- 'multiprop' ----
		{Code: `
        <Foo />
      `, Tsx: true, Options: []interface{}{"multiprop"}},
		{Code: `
        <Foo bar />
      `, Tsx: true, Options: []interface{}{"multiprop"}},
		{Code: `
        <Foo {...this.props} />
      `, Tsx: true, Options: []interface{}{"multiprop"}},
	}, []rule_tester.InvalidTestCase{
		// ---- 'always': single-line tag, first prop on same line ----
		{
			Code: `
        <Foo propOne="one" propTwo="two" />
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output: []string{`
        <Foo
propOne="one" propTwo="two" />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 2, Column: 14},
			},
		},
		// ---- 'always': multiline tag, first prop on same line ----
		{
			Code: `
        <Foo propOne="one"
          propTwo="two"
        />
      `,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output: []string{`
        <Foo
propOne="one"
          propTwo="two"
        />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 2, Column: 14},
			},
		},
		// ---- 'never': first prop on new line ----
		{
			Code: `
        <Foo
          propOne="one"
          propTwo="two"
        />
      `,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output: []string{`
        <Foo propOne="one"
          propTwo="two"
        />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnSameLine", Message: msgPropOnSameLine, Line: 3, Column: 11},
			},
		},
		// ---- 'multiline': prop on same line of a multiline tag ----
		{
			Code: `
        <Foo prop={{
        }} />
      `,
			Tsx:     true,
			Options: []interface{}{"multiline"},
			Output: []string{`
        <Foo
prop={{
        }} />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 2, Column: 14},
			},
		},
		// ---- 'multiline-multiprop': multiline, multiple props, first on same line ----
		{
			Code: `
        <Foo bar={{
        }} baz />
      `,
			Tsx:     true,
			Options: []interface{}{"multiline-multiprop"},
			Output: []string{`
        <Foo
bar={{
        }} baz />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 2, Column: 14},
			},
		},
		// ---- 'multiprop': multiple props on same line ----
		{
			Code: `
      <Foo propOne="one" propTwo="two" />
      `,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output: []string{`
      <Foo
propOne="one" propTwo="two" />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 2, Column: 12},
			},
		},
		// ---- 'multiprop': single prop on new line in a multiline tag → same line ----
		{
			Code: `
      <Foo
bar />
      `,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output: []string{`
      <Foo bar />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnSameLine", Message: msgPropOnSameLine, Line: 3, Column: 1},
			},
		},
		// ---- 'multiprop': single spread prop on new line → same line ----
		{
			Code: `
      <Foo
{...this.props} />
      `,
			Tsx:     true,
			Options: []interface{}{"multiprop"},
			Output: []string{`
      <Foo {...this.props} />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnSameLine", Message: msgPropOnSameLine, Line: 3, Column: 1},
			},
		},
		// ---- 'multiline': TypeScript generic component (typeArguments fix anchor) ----
		{
			Code: `
        <DataTable<Items> fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `,
			Tsx:     true,
			Options: []interface{}{"multiline"},
			Output: []string{`
        <DataTable<Items>
fullscreen keyField="id" items={items}
          activeSortableColumn={sorting}
          onSortClick={handleSortedClick}
          rowActions={[
          ]}
        />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propOnNewLine", Message: msgPropOnNewLine, Line: 2, Column: 27},
			},
		},
	})
}
