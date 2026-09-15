// TestJsxCurlyNewlineUpstream ports eslint-plugin-react v7.37.5's complete
// jsx-curly-newline suite, plus the rule documentation's examples.
package jsx_curly_newline

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func errorAt(id, message string, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{MessageId: id, Message: message, Line: line, Column: column, EndLine: line, EndColumn: column + 1}
}

func TestJsxCurlyNewlineUpstream(t *testing.T) {
	consistentOption := []any{"consistent"}
	neverOption := []any{"never"}
	requireMultiline := []any{map[string]any{"singleline": "consistent", "multiline": "require"}}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxCurlyNewlineRule, []rule_tester.ValidTestCase{
		{Code: `<div>{foo}</div>`, Tsx: true, Options: consistentOption},
		{Code: `
        <div>
          {
            foo
          }
        </div>
      `, Tsx: true, Options: consistentOption},
		{Code: `
        <div>
          { foo &&
            foo.bar }
        </div>
      `, Tsx: true, Options: consistentOption},
		{Code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `, Tsx: true, Options: consistentOption},
		{Code: `
        <div foo={
          bar
        } />
      `, Tsx: true, Options: consistentOption},
		{Code: `<div>{foo}</div>`, Tsx: true, Options: requireMultiline},
		{Code: `<div foo={bar} />`, Tsx: true, Options: requireMultiline},
		{Code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `, Tsx: true, Options: requireMultiline},
		{Code: `
        <div>
          {
            foo
          }
        </div>
      `, Tsx: true, Options: requireMultiline},
		{Code: `<div>{foo}</div>`, Tsx: true, Options: neverOption},
		{Code: `<div foo={bar} />`, Tsx: true, Options: neverOption},
		{Code: `
        <div>
          { foo &&
            foo.bar }
        </div>
      `, Tsx: true, Options: neverOption},
		// Documentation examples.
		{Code: `<div>{ foo }</div>`, Tsx: true},
		{Code: "<div>{\n  foo\n}</div>", Tsx: true},
		{Code: "<div>{ foo &&\n  foo.bar }</div>", Tsx: true, Options: neverOption},
	}, []rule_tester.InvalidTestCase{
		{Code: "\n        <div>\n          { foo \n}\n        </div>\n      ", Tsx: true, Output: []string{`
        <div>
          { foo}
        </div>
      `}, Options: consistentOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBefore"}}},
		{Code: "\n        <div>\n          { foo &&\n            foo.bar \n}\n        </div>\n      ", Tsx: true, Output: []string{`
        <div>
          { foo &&
            foo.bar}
        </div>
      `}, Options: consistentOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBefore"}}},
		{Code: `
        <div>
          { foo &&
            bar
          }
        </div>
      `, Tsx: true, Output: []string{`
        <div>
          { foo &&
            bar}
        </div>
      `}, Options: consistentOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBefore"}}},
		{Code: `<div>{foo
}</div>`, Tsx: true, Output: []string{`<div>{foo}</div>`}, Options: requireMultiline, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBefore"}}},
		{Code: `<div>{
foo}</div>`, Tsx: true, Output: []string{`<div>{
foo
}</div>`}, Options: requireMultiline, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedBefore"}}},
		{Code: `
        <div>
          { foo &&
            bar }
        </div>
      `, Tsx: true, Output: []string{"\n        <div>\n          {\n foo &&\n            bar \n}\n        </div>\n      "}, Options: requireMultiline, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedAfter"}, {MessageId: "expectedBefore"}}},
		{Code: `
        <div style={foo &&
          foo.bar
        } />
      `, Tsx: true, Output: []string{`
        <div style={
foo &&
          foo.bar
        } />
      `}, Options: requireMultiline, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectedAfter"}}},
		{Code: `
        <div>
          {
foo
}
        </div>
      `, Tsx: true, Output: []string{`
        <div>
          {foo}
        </div>
      `}, Options: neverOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedAfter"}, {MessageId: "unexpectedBefore"}}},
		{Code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `, Tsx: true, Output: []string{`
        <div>
          {foo &&
            foo.bar}
        </div>
      `}, Options: neverOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedAfter"}, {MessageId: "unexpectedBefore"}}},
		{Code: `
        <div>
          { foo &&
            foo.bar
          }
        </div>
      `, Tsx: true, Output: []string{`
        <div>
          { foo &&
            foo.bar}
        </div>
      `}, Options: neverOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBefore"}}},
		{Code: `
        <div>
          { /* not fixed due to comment */
            foo }
        </div>
      `, Tsx: true, Options: neverOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedAfter"}}},
		{Code: `
        <div>
          { foo
            /* not fixed due to comment */}
        </div>
      `, Tsx: true, Options: neverOption, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedBefore"}}},
		// Documentation examples.
		{Code: "<div>{ foo\n}</div>", Tsx: true, Output: []string{"<div>{ foo}</div>"}, Errors: []rule_tester.InvalidTestCaseError{errorAt("unexpectedBefore", "Unexpected newline before '}'.", 2, 1)}},
		{Code: "<div>{\n  foo }</div>", Tsx: true, Output: []string{"<div>{\n  foo \n}</div>"}, Errors: []rule_tester.InvalidTestCaseError{errorAt("expectedBefore", "Expected newline before '}'.", 2, 7)}},
	})
}
