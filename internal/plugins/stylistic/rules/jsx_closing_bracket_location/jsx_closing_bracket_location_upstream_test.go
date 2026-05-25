// TestJsxClosingBracketLocationUpstream migrates the full valid/invalid suite
// from upstream packages/eslint-plugin/rules/jsx-closing-bracket-location/
// jsx-closing-bracket-location.test.ts 1:1. Position assertions cover
// line/column for every invalid case that upstream itself asserts them on.
// rslint-specific lock-in cases live in
// jsx_closing_bracket_location_extras_test.go.
package jsx_closing_bracket_location

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxClosingBracketLocationUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxClosingBracketLocationRule, []rule_tester.ValidTestCase{
		// ---- default tag-aligned, simple shapes ----
		{Code: "\n        <App />\n      ", Tsx: true},
		{Code: "\n        <App\n          // comment\n        />\n      ", Tsx: true},
		{Code: "\n        <App /** comment */ />\n      ", Tsx: true},
		{Code: "\n        <App foo />\n      ", Tsx: true},
		{Code: "\n        <App\n          foo\n        />\n      ", Tsx: true},
		{Code: "\n        <App\n          foo\n          // comment\n        />\n      ", Tsx: true},
		{Code: "\n        <App\n          {...foo}\n        />\n      ", Tsx: true},
		{Code: "\n        <App\n          {...foo}\n          // comment\n        />\n      ", Tsx: true},
		// ---- string / object option equivalence for single-line forms ----
		{Code: "\n        <App foo />\n      ", Tsx: true, Options: map[string]interface{}{"location": "after-props"}},
		{Code: "\n        <App foo />\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		{Code: "\n        <App foo />\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		// ---- string options on multi-line shapes ----
		{Code: "\n        <App\n          foo />\n      ", Tsx: true, Options: "after-props"},
		{Code: "\n        <App\n          foo\n          />\n      ", Tsx: true, Options: "props-aligned"},
		// ---- object options on multi-line shapes ----
		{Code: "\n        <App\n          foo />\n      ", Tsx: true, Options: map[string]interface{}{"location": "after-props"}},
		{Code: "\n        <App\n          foo\n        />\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		{Code: "\n        <App\n          foo\n        />\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		{Code: "\n        <App\n          foo\n          />\n      ", Tsx: true, Options: map[string]interface{}{"location": "props-aligned"}},
		// ---- non-self-closing equivalents ----
		{Code: "\n        <App foo></App>\n      ", Tsx: true},
		{Code: "\n        <App\n          foo\n        ></App>\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		{Code: "\n        <App\n          foo\n        ></App>\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		{Code: "\n        <App\n          foo\n          ></App>\n      ", Tsx: true, Options: map[string]interface{}{"location": "props-aligned"}},
		// ---- multi-line attribute value ----
		{Code: "\n        <App\n          foo={function() {\n            console.log('bar');\n          }} />\n      ", Tsx: true, Options: map[string]interface{}{"location": "after-props"}},
		{Code: "\n        <App\n          foo={function() {\n            console.log('bar');\n          }}\n          />\n      ", Tsx: true, Options: map[string]interface{}{"location": "props-aligned"}},
		{Code: "\n        <App\n          foo={function() {\n            console.log('bar');\n          }}\n        />\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		{Code: "\n        <App\n          foo={function() {\n            console.log('bar');\n          }}\n        />\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		// ---- attribute on opening line, closing same-line collapses to after-props ----
		{Code: "\n        <App foo={function() {\n          console.log('bar');\n        }}/>\n      ", Tsx: true, Options: map[string]interface{}{"location": "after-props"}},
		{Code: "\n         <App foo={function() {\n                console.log('bar');\n              }}\n              />\n      ", Tsx: true, Options: map[string]interface{}{"location": "props-aligned"}},
		{Code: "\n        <App foo={function() {\n          console.log('bar');\n        }}\n        />\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		{Code: "\n        <App foo={function() {\n          console.log('bar');\n        }}\n        />\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		// ---- selfClosing / nonEmpty per-form configuration ----
		{Code: "\n        <Provider store>\n          <App\n            foo />\n        </Provider>\n      ", Tsx: true, Options: map[string]interface{}{"selfClosing": "after-props"}},
		{Code: "\n        <Provider\n          store\n        >\n          <App\n            foo />\n        </Provider>\n      ", Tsx: true, Options: map[string]interface{}{"selfClosing": "after-props"}},
		{Code: "\n        <Provider\n          store>\n          <App\n            foo\n          />\n        </Provider>\n      ", Tsx: true, Options: map[string]interface{}{"nonEmpty": "after-props"}},
		{Code: "\n        <Provider store>\n          <App\n            foo\n            />\n        </Provider>\n      ", Tsx: true, Options: map[string]interface{}{"selfClosing": "props-aligned"}},
		{Code: "\n        <Provider\n          store\n          >\n          <App\n            foo\n          />\n        </Provider>\n      ", Tsx: true, Options: map[string]interface{}{"nonEmpty": "props-aligned"}},
		// ---- larger realistic shapes (function return, var assignment, etc.) ----
		{Code: "\n        var x = function() {\n          return <App\n            foo\n                 >\n              bar\n                 </App>\n        }\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		{Code: "\n        var x = function() {\n          return <App\n            foo\n                 />\n        }\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		{Code: "\n        var x = <App\n          foo\n                />\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		{Code: "\n        var x = function() {\n          return <App\n            foo={function() {\n              console.log('bar');\n            }}\n          />\n        }\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		{Code: "\n        var x = <App\n          foo={function() {\n            console.log('bar');\n          }}\n        />\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		{Code: "\n        <Provider\n          store\n        >\n          <App\n            foo={function() {\n              console.log('bar');\n            }}\n          />\n        </Provider>\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		{Code: "\n        <Provider\n          store\n        >\n          {baz && <App\n            foo={function() {\n              console.log('bar');\n            }}\n          />}\n        </Provider>\n      ", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		// ---- nonEmpty:false / selfClosing:false disable per-form ----
		{Code: "\n        <App>\n          <Foo\n            bar\n          >\n          </Foo>\n          <Foo\n            bar />\n        </App>\n      ", Tsx: true, Options: map[string]interface{}{"nonEmpty": false, "selfClosing": "after-props"}},
		{Code: "\n        <App>\n          <Foo\n            bar>\n          </Foo>\n          <Foo\n            bar\n          />\n        </App>\n      ", Tsx: true, Options: map[string]interface{}{"nonEmpty": "after-props", "selfClosing": false}},
		// ---- multi-line attribute with array literal closing on its own line ----
		{Code: "\n        <div className={[\n          \"some\",\n          \"stuff\",\n          2 ]}\n        >\n          Some text\n        </div>\n      ", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
	}, []rule_tester.InvalidTestCase{
		// ---- 1. after-tag (no props), fix collapses to single-line ----
		{
			Code:   "\n        <App\n        />\n      ",
			Tsx:    true,
			Output: []string{"\n        <App />\n      "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the opening tag",
			}},
		},
		// ---- 2. after-props default, prop on opening line, `/>` on next ----
		{
			Code:   "\n        <App foo\n        />\n      ",
			Tsx:    true,
			Output: []string{"\n        <App foo/>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		// ---- 3. non-self-closing, after-props default ----
		{
			Code:   "\n        <App foo\n        ></App>\n      ",
			Tsx:    true,
			Output: []string{"\n        <App foo></App>\n      "},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		// ---- 4-6. multi-line prop, `/>` on same line as prop ----
		{
			Code:    "\n        <App\n          foo />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n          />\n      "},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the last prop (expected column 11 on the next line)",
				Line:      3, Column: 15,
			}},
		},
		{
			Code:    "\n        <App\n          foo />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n        />\n      "},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 9 on the next line)",
				Line:      3, Column: 15,
			}},
		},
		{
			Code:    "\n        <App\n          foo />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n        />\n      "},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9 on the next line)",
				Line:      3, Column: 15,
			}},
		},
		// ---- 7-11. multi-line prop, after-props collapses; props/tag/line aligned move bracket ----
		{
			Code:    "\n        <App\n          foo\n        />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo/>\n      "},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		{
			Code:    "\n        <App\n          foo\n        />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n          />\n      "},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the last prop (expected column 11)",
				Line:      4, Column: 9,
			}},
		},
		{
			Code:    "\n        <App\n          foo\n          />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo/>\n      "},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		{
			Code:    "\n        <App\n          foo\n          />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n        />\n      "},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 9)",
				Line:      4, Column: 11,
			}},
		},
		{
			Code:    "\n        <App\n          foo\n          />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n        />\n      "},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9)",
				Line:      4, Column: 11,
			}},
		},
		// ---- 12-16. non-self-closing variants ----
		{
			Code:    "\n        <App\n          foo\n        ></App>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo></App>\n      "},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		{
			Code:    "\n        <App\n          foo\n        ></App>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n          ></App>\n      "},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the last prop (expected column 11)",
				Line:      4, Column: 9,
			}},
		},
		{
			Code:    "\n        <App\n          foo\n          ></App>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo></App>\n      "},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		{
			Code:    "\n        <App\n          foo\n          ></App>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n        ></App>\n      "},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 9)",
				Line:      4, Column: 11,
			}},
		},
		{
			Code:    "\n        <App\n          foo\n          ></App>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <App\n          foo\n        ></App>\n      "},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9)",
				Line:      4, Column: 11,
			}},
		},
		// ---- 17. nested Provider/App, selfClosing config ----
		{
			Code:    "\n        <Provider\n          store>\n          <App\n            foo\n            />\n        </Provider>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <Provider\n          store\n        >\n          <App\n            foo\n            />\n        </Provider>\n      "},
			Options: map[string]interface{}{"selfClosing": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 9 on the next line)",
				Line:      3, Column: 16,
			}},
		},
		// ---- 18-20. very-far-right closing bracket ----
		{
			Code:    "\n        const Button = function(props) {\n          return (\n            <Button\n              size={size}\n              onClick={onClick}\n                                            >\n              Button Text\n            </Button>\n          );\n        };\n      ",
			Tsx:     true,
			Output:  []string{"\n        const Button = function(props) {\n          return (\n            <Button\n              size={size}\n              onClick={onClick}\n              >\n              Button Text\n            </Button>\n          );\n        };\n      "},
			Options: "props-aligned",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the last prop (expected column 15)",
				Line:      7, Column: 45,
			}},
		},
		{
			Code:    "\n        const Button = function(props) {\n          return (\n            <Button\n              size={size}\n              onClick={onClick}\n                                            >\n              Button Text\n            </Button>\n          );\n        };\n      ",
			Tsx:     true,
			Output:  []string{"\n        const Button = function(props) {\n          return (\n            <Button\n              size={size}\n              onClick={onClick}\n            >\n              Button Text\n            </Button>\n          );\n        };\n      "},
			Options: "tag-aligned",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 13)",
				Line:      7, Column: 45,
			}},
		},
		{
			Code:    "\n        const Button = function(props) {\n          return (\n            <Button\n              size={size}\n              onClick={onClick}\n                                            >\n              Button Text\n            </Button>\n          );\n        };\n      ",
			Tsx:     true,
			Output:  []string{"\n        const Button = function(props) {\n          return (\n            <Button\n              size={size}\n              onClick={onClick}\n            >\n              Button Text\n            </Button>\n          );\n        };\n      "},
			Options: "line-aligned",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 13)",
				Line:      7, Column: 45,
			}},
		},
		// ---- 21-23. nonEmpty / selfClosing per-form ----
		{
			Code:    "\n        <Provider\n          store\n          >\n          <App\n            foo\n            />\n        </Provider>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <Provider\n          store\n          >\n          <App\n            foo\n          />\n        </Provider>\n      "},
			Options: map[string]interface{}{"nonEmpty": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 11)",
				Line:      7, Column: 13,
			}},
		},
		{
			Code:    "\n        <Provider\n          store>\n          <App\n            foo />\n        </Provider>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <Provider\n          store\n        >\n          <App\n            foo />\n        </Provider>\n      "},
			Options: map[string]interface{}{"selfClosing": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 9 on the next line)",
				Line:      3, Column: 16,
			}},
		},
		{
			Code:    "\n        <Provider\n          store>\n          <App\n            foo\n            />\n        </Provider>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <Provider\n          store>\n          <App\n            foo\n          />\n        </Provider>\n      "},
			Options: map[string]interface{}{"nonEmpty": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 11)",
				Line:      6, Column: 13,
			}},
		},
		// ---- 24-25. function return, line-aligned ----
		{
			Code:    "\n        var x = function() {\n          return <App\n            foo\n                />\n        }\n      ",
			Tsx:     true,
			Output:  []string{"\n        var x = function() {\n          return <App\n            foo\n          />\n        }\n      "},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 11)",
				Line:      5, Column: 17,
			}},
		},
		{
			Code:    "\n        var x = <App\n          foo\n                />\n      ",
			Tsx:     true,
			Output:  []string{"\n        var x = <App\n          foo\n        />\n      "},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9)",
				Line:      4, Column: 17,
			}},
		},
		// ---- 26-27. paren-wrapped JSX, spread / nested element ----
		{
			Code:    "\n        var x = (\n          <div\n            className=\"MyComponent\"\n            {...props} />\n        )\n      ",
			Tsx:     true,
			Output:  []string{"\n        var x = (\n          <div\n            className=\"MyComponent\"\n            {...props}\n          />\n        )\n      "},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 11 on the next line)",
				Line:      5, Column: 24,
			}},
		},
		{
			Code:    "\n        var x = (\n          <Something\n            content={<Foo />} />\n        )\n      ",
			Tsx:     true,
			Output:  []string{"\n        var x = (\n          <Something\n            content={<Foo />}\n          />\n        )\n      "},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 11 on the next line)",
				Line:      4, Column: 31,
			}},
		},
		// ---- 28. after-tag on Something — collapse to single line ----
		{
			Code:    "\n        var x = (\n          <Something\n            />\n        )\n      ",
			Tsx:     true,
			Output:  []string{"\n        var x = (\n          <Something />\n        )\n      "},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the opening tag",
			}},
		},
		// ---- 29. div className=[...] tag-aligned ----
		{
			Code:    "\n        <div className={[\n          \"some\",\n          \"stuff\",\n          2 ]}>\n          Some text\n        </div>\n      ",
			Tsx:     true,
			Output:  []string{"\n        <div className={[\n          \"some\",\n          \"stuff\",\n          2 ]}\n        >\n          Some text\n        </div>\n      "},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 9 on the next line)",
				Line:      5, Column: 15,
			}},
		},
		// ---- 30-32. tab-indented variants ----
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo />\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the last prop (expected column 6 on the next line)",
				Line:      3, Column: 10,
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo />\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 5 on the next line)",
				Line:      3, Column: 10,
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo />\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5 on the next line)",
				Line:      3, Column: 10,
			}},
		},
		// ---- 33-36. tab-indented existing-newline variants ----
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the last prop (expected column 6)",
				Line:      4, Column: 5,
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 5)",
				Line:      4, Column: 6,
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5)",
				Line:      4, Column: 6,
			}},
		},
		// ---- 37-40. tab-indented non-self-closing ----
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t></App>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo></App>\n\t\t\t"},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t></App>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t></App>\n\t\t\t"},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the last prop (expected column 6)",
				Line:      4, Column: 5,
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t></App>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo></App>\n\t\t\t"},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t></App>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t></App>\n\t\t\t"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 5)",
				Line:      4, Column: 6,
			}},
		},
		{
			Code:    "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t></App>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t></App>\n\t\t\t"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5)",
				Line:      4, Column: 6,
			}},
		},
		// ---- 41. nested Provider/App with selfClosing (tab) ----
		{
			Code:    "\n\t\t\t\t<Provider\n\t\t\t\t\tstore>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<Provider\n\t\t\t\t\tstore\n\t\t\t\t>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t"},
			Options: map[string]interface{}{"selfClosing": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 5 on the next line)",
				Line:      3, Column: 11,
			}},
		},
		// ---- 42-44. very-far-right (tab) ----
		{
			Code:    "\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t"},
			Options: "props-aligned",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the last prop (expected column 8)",
				Line:      7, Column: 23,
			}},
		},
		{
			Code:    "\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t"},
			Options: "tag-aligned",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 7)",
				Line:      7, Column: 23,
			}},
		},
		{
			Code:    "\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t"},
			Options: "line-aligned",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 7)",
				Line:      7, Column: 23,
			}},
		},
		// ---- 45. tab Provider/App nonEmpty: props-aligned ----
		{
			Code:    "\n\t\t\t\t<Provider\n\t\t\t\t\tstore\n\t\t\t\t\t>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<Provider\n\t\t\t\t\tstore\n\t\t\t\t\t>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t"},
			Options: map[string]interface{}{"nonEmpty": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 6)",
				Line:      7, Column: 7,
			}},
		},
		// ---- 46-47. tab selfClosing / nonEmpty after-props ----
		{
			Code:    "\n\t\t\t\t<Provider\n\t\t\t\t\tstore>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo />\n\t\t\t\t</Provider>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<Provider\n\t\t\t\t\tstore\n\t\t\t\t>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo />\n\t\t\t\t</Provider>\n\t\t\t"},
			Options: map[string]interface{}{"selfClosing": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 5 on the next line)",
				Line:      3, Column: 11,
			}},
		},
		{
			Code:    "\n\t\t\t\t<Provider\n\t\t\t\t\tstore>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<Provider\n\t\t\t\t\tstore>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t"},
			Options: map[string]interface{}{"nonEmpty": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 6)",
				Line:      6, Column: 7,
			}},
		},
		// ---- 48-50. line-aligned cases (var x = function / var x = <App / paren-wrapped) ----
		{
			Code:    "\n\t\t\t\tvar x = function() {\n\t\t\t\t\treturn <App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t\t\t/>\n\t\t\t\t}\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\tvar x = function() {\n\t\t\t\t\treturn <App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t\t}\n\t\t\t"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 6)",
				Line:      5, Column: 9,
			}},
		},
		{
			Code:    "\n\t\t\t\tvar x = <App\n\t\t\t\t\tfoo\n\t\t\t\t\t\t\t\t/>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\tvar x = <App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5)",
				Line:      4, Column: 9,
			}},
		},
		{
			Code:    "\n\t\t\t\tvar x = (\n\t\t\t\t\t<div\n\t\t\t\t\t\tclassName=\"MyComponent\"\n\t\t\t\t\t\t{...props} />\n\t\t\t\t)\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\tvar x = (\n\t\t\t\t\t<div\n\t\t\t\t\t\tclassName=\"MyComponent\"\n\t\t\t\t\t\t{...props}\n\t\t\t\t\t/>\n\t\t\t\t)\n\t\t\t"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 6 on the next line)",
				Line:      5, Column: 18,
			}},
		},
		// ---- 51-53. var x = ( <Something content={<Foo />} /> ) / <Something /> ----
		{
			Code:    "\n\t\t\t\tvar x = (\n\t\t\t\t\t<Something\n\t\t\t\t\t\tcontent={<Foo />} />\n\t\t\t\t)\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\tvar x = (\n\t\t\t\t\t<Something\n\t\t\t\t\t\tcontent={<Foo />}\n\t\t\t\t\t/>\n\t\t\t\t)\n\t\t\t"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 6 on the next line)",
				Line:      4, Column: 25,
			}},
		},
		{
			Code:    "\n\t\t\t\tvar x = (\n\t\t\t\t\t<Something\n\t\t\t\t\t\t/>\n\t\t\t\t)\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\tvar x = (\n\t\t\t\t\t<Something />\n\t\t\t\t)\n\t\t\t"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the opening tag",
			}},
		},
		// ---- 54. tab div className=[...] tag-aligned ----
		{
			Code:    "\n\t\t\t\t<div className={[\n\t\t\t\t\t\"some\",\n\t\t\t\t\t\"stuff\",\n\t\t\t\t\t2 ]}>\n\t\t\t\t\tSome text\n\t\t\t\t</div>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<div className={[\n\t\t\t\t\t\"some\",\n\t\t\t\t\t\"stuff\",\n\t\t\t\t\t2 ]}\n\t\t\t\t>\n\t\t\t\t\tSome text\n\t\t\t\t</div>\n\t\t\t"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 5 on the next line)",
				Line:      5, Column: 10,
			}},
		},
		// ---- 55. Mixed-tab edge case (opening 6 tabs, prop 7 tabs, closing 5 tabs) ----
		{
			Code:    "\n\t\t\t\t\t\t\t<div\n\t\t\t\t\t\t\t\tclassName={styles}\n\t\t\t\t\t >\n\t\t\t\t\t\t\t\t{props}\n\t\t\t\t\t\t\t</div>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t\t\t\t<div\n\t\t\t\t\t\t\t\tclassName={styles}\n\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\t\t{props}\n\t\t\t\t\t\t\t</div>\n\t\t\t"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 8)",
				Line:      4, Column: 7,
			}},
		},
		// ---- 56. space variant: div, tag-aligned ----
		{
			Code:    "\n          <div\n            className={styles}\n            >\n            {props}\n          </div>\n      ",
			Tsx:     true,
			Output:  []string{"\n          <div\n            className={styles}\n          >\n            {props}\n          </div>\n      "},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 11)",
				Line:      4, Column: 13,
			}},
		},
		// ---- 57. space variant: App tag-aligned ----
		{
			Code:    "\n          <App\n            foo\n            />\n      ",
			Tsx:     true,
			Output:  []string{"\n          <App\n            foo\n          />\n      "},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 11)",
				Line:      4, Column: 13,
			}},
		},
		// ---- 58. mixed-tab opening/prop/closing (6/7/5) ----
		{
			Code:    "\n\t\t\t\t\t\t<App\n\t\t\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t\t\t<App\n\t\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 7)",
				Line:      4, Column: 6,
			}},
		},
		// ---- 59. trailing line-comment promotes after-tag/after-props → line-aligned (tag-aligned mode) ----
		// Upstream: location=tag-aligned mismatches → no upgrade (still tag-aligned).
		{
			Code:    "\n        <input\n          // comment\n          type=\"text\"\n          // comment\n          />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <input\n          // comment\n          type=\"text\"\n          // comment\n        />\n      "},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 9)",
				Line:      6, Column: 11,
			}},
		},
		// ---- 60. trailing block-comment, tag-aligned mode ----
		{
			Code:    "\n        <input\n          // comment\n          type=\"text\"\n          /**\n           * \n           * comment\n           * \n           */\n          />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <input\n          // comment\n          type=\"text\"\n          /**\n           * \n           * comment\n           * \n           */\n        />\n      "},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 9)",
				Line:      10, Column: 11,
			}},
		},
		// ---- 61. after-props with NO trailing comment between last attr and `/>` → stays after-props ----
		// The blank line after `type="text"` is whitespace only, not a comment, so
		// the after-props upgrade-to-line-aligned does NOT fire.
		{
			Code:    "\n        <input\n          // comment\n          type=\"text\"\n\n        />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <input\n          // comment\n          type=\"text\"/>\n      "},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
				Line:      6, Column: 9,
			}},
		},
		// ---- 62. after-props upgraded to line-aligned by trailing comment AFTER last attr ----
		{
			Code:    "\n        <input\n          // comment\n          type=\"text\"\n          // comment\n          />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <input\n          // comment\n          type=\"text\"\n          // comment\n        />\n      "},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9)",
				Line:      6, Column: 11,
			}},
		},
		// ---- 63. zero-prop variant: comment between tag-name and `/>` upgrades after-tag→line-aligned ----
		{
			Code:    "\n        <input\n          // comment\n          // comment\n          />\n      ",
			Tsx:     true,
			Output:  []string{"\n        <input\n          // comment\n          // comment\n        />\n      "},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9)",
				Line:      5, Column: 11,
			}},
		},
		// ---- 64. tab-indented after-props upgraded to line-aligned ----
		{
			Code:    "\n\t\t\t\t<a\n\t\t\t\t\thref=\"javascript:;\"\n\t\t\t\t\t// comment\n\t\t\t\t\t// comment\n\t\t\t\t\t>\n\t\t\t\t\ttext\n\t\t\t\t</a>\n      ",
			Tsx:     true,
			Output:  []string{"\n\t\t\t\t<a\n\t\t\t\t\thref=\"javascript:;\"\n\t\t\t\t\t// comment\n\t\t\t\t\t// comment\n\t\t\t\t>\n\t\t\t\t\ttext\n\t\t\t\t</a>\n      "},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5)",
				Line:      6, Column: 6,
			}},
		},
	})
}
