package no_else_return

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoElseReturnUpstream migrates the full valid/invalid suite from upstream
// ESLint tests/lib/rules/no-else-return.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the no_else_return_extras_test.go file.
func TestNoElseReturnUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoElseReturnRule,
		[]rule_tester.ValidTestCase{
			{Code: "function foo() { if (true) { if (false) { return x; } } else { return y; } }"},
			{Code: "function foo() { if (true) { return x; } return y; }"},
			{Code: "function foo() { if (true) { for (;;) { return x; } } else { return y; } }"},
			{Code: "function foo() { var x = true; if (x) { return x; } else if (x === false) { return false; } }"},
			{Code: "function foo() { if (true) notAReturn(); else return y; }"},
			{Code: "function foo() {if (x) { notAReturn(); } else if (y) { return true; } else { notAReturn(); } }"},
			{Code: "function foo() {if (x) { return true; } else if (y) { notAReturn() } else { notAReturn(); } }"},
			{Code: "if (0) { if (0) {} else {} } else {}"},
			{Code: `
            function foo() {
                if (foo)
                    if (bar) return;
                    else baz;
                else qux;
            }
        `},
			{Code: `
            function foo() {
                while (foo)
                    if (bar) return;
                    else baz;
            }
        `},
			{
				Code:    "function foo19() { if (true) { return x; } else if (false) { return y; } }",
				Options: map[string]interface{}{"allowElseIf": true},
			},
			{
				Code:    "function foo20() {if (x) { return true; } else if (y) { notAReturn() } else { notAReturn(); } }",
				Options: map[string]interface{}{"allowElseIf": true},
			},
			{
				Code:    "function foo21() { var x = true; if (x) { return x; } else if (x === false) { return false; } }",
				Options: map[string]interface{}{"allowElseIf": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "function foo1() { if (true) { return x; } else { return y; } }",
				Output: []string{"function foo1() { if (true) { return x; }  return y;  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo1() { if (true) { return x; } else { return y; } }", "{ return y; }"),
				},
			},
			{
				Code:   "function foo2() { if (true) { var x = bar; return x; } else { var y = baz; return y; } }",
				Output: []string{"function foo2() { if (true) { var x = bar; return x; }  var y = baz; return y;  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo2() { if (true) { var x = bar; return x; } else { var y = baz; return y; } }", "{ var y = baz; return y; }"),
				},
			},
			{
				Code:   "function foo3() { if (true) return x; else return y; }",
				Output: []string{"function foo3() { if (true) return x; return y; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo3() { if (true) return x; else return y; }", "return y;"),
				},
			},
			{
				Code: "function foo4() { if (true) { if (false) return x; else return y; } else { return z; } }",
				Output: []string{
					"function foo4() { if (true) { if (false) return x; return y; } else { return z; } }",
					"function foo4() { if (true) { if (false) return x; return y; }  return z;  }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo4() { if (true) { if (false) return x; else return y; } else { return z; } }", "return y;"),
					unexpectedAt("function foo4() { if (true) { if (false) return x; else return y; } else { return z; } }", "{ return z; }"),
				},
			},
			{
				Code:   "function foo5() { if (true) { if (false) { if (true) return x; else { w = y; } } else { w = x; } } else { return z; } }",
				Output: []string{"function foo5() { if (true) { if (false) { if (true) return x;  w = y;  } else { w = x; } } else { return z; } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo5() { if (true) { if (false) { if (true) return x; else { w = y; } } else { w = x; } } else { return z; } }", "{ w = y; }"),
				},
			},
			{
				Code:   "function foo6() { if (true) { if (false) { if (true) return x; else return y; } } else { return z; } }",
				Output: []string{"function foo6() { if (true) { if (false) { if (true) return x; return y; } } else { return z; } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo6() { if (true) { if (false) { if (true) return x; else return y; } } else { return z; } }", "return y;"),
				},
			},
			{
				Code: "function foo7() { if (true) { if (false) { if (true) return x; else return y; } return w; } else { return z; } }",
				Output: []string{
					"function foo7() { if (true) { if (false) { if (true) return x; return y; } return w; } else { return z; } }",
					"function foo7() { if (true) { if (false) { if (true) return x; return y; } return w; }  return z;  }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo7() { if (true) { if (false) { if (true) return x; else return y; } return w; } else { return z; } }", "return y;"),
					unexpectedAt("function foo7() { if (true) { if (false) { if (true) return x; else return y; } return w; } else { return z; } }", "{ return z; }"),
				},
			},
			{
				Code: "function foo8() { if (true) { if (false) { if (true) return x; else return y; } else { w = x; } } else { return z; } }",
				Output: []string{
					"function foo8() { if (true) { if (false) { if (true) return x; return y; } else { w = x; } } else { return z; } }",
					"function foo8() { if (true) { if (false) { if (true) return x; return y; }  w = x;  } else { return z; } }",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo8() { if (true) { if (false) { if (true) return x; else return y; } else { w = x; } } else { return z; } }", "return y;"),
					unexpectedAt("function foo8() { if (true) { if (false) { if (true) return x; else return y; } else { w = x; } } else { return z; } }", "{ w = x; }"),
				},
			},
			{
				Code:   "function foo9() {if (x) { return true; } else if (y) { return true; } else { notAReturn(); } }",
				Output: []string{"function foo9() {if (x) { return true; } else if (y) { return true; }  notAReturn();  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo9() {if (x) { return true; } else if (y) { return true; } else { notAReturn(); } }", "{ notAReturn(); }"),
				},
			},
			{
				Code: "function foo9a() {if (x) { return true; } else if (y) { return true; } else { notAReturn(); } }",
				Output: []string{
					"function foo9a() {if (x) { return true; } if (y) { return true; } else { notAReturn(); } }",
					"function foo9a() {if (x) { return true; } if (y) { return true; }  notAReturn();  }",
				},
				Options: map[string]interface{}{"allowElseIf": false},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo9a() {if (x) { return true; } else if (y) { return true; } else { notAReturn(); } }", "if (y) { return true; } else { notAReturn(); }"),
				},
			},
			{
				Code:    "function foo9b() {if (x) { return true; } if (y) { return true; } else { notAReturn(); } }",
				Output:  []string{"function foo9b() {if (x) { return true; } if (y) { return true; }  notAReturn();  }"},
				Options: map[string]interface{}{"allowElseIf": false},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo9b() {if (x) { return true; } if (y) { return true; } else { notAReturn(); } }", "{ notAReturn(); }"),
				},
			},
			{
				Code:   "function foo10() { if (foo) return bar; else (foo).bar(); }",
				Output: []string{"function foo10() { if (foo) return bar; (foo).bar(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo10() { if (foo) return bar; else (foo).bar(); }", "(foo).bar();"),
				},
			},
			{
				Code: "function foo11() { if (foo) return bar \nelse { [1, 2, 3].map(foo) } }",
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo11() { if (foo) return bar \nelse { [1, 2, 3].map(foo) } }", "{ [1, 2, 3].map(foo) }"),
				},
			},
			{
				Code: "function foo12() { if (foo) return bar \nelse { baz() } \n[1, 2, 3].map(foo) }",
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo12() { if (foo) return bar \nelse { baz() } \n[1, 2, 3].map(foo) }", "{ baz() }"),
				},
			},
			{
				Code:   "function foo13() { if (foo) return bar; \nelse { [1, 2, 3].map(foo) } }",
				Output: []string{"function foo13() { if (foo) return bar; \n [1, 2, 3].map(foo)  }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo13() { if (foo) return bar; \nelse { [1, 2, 3].map(foo) } }", "{ [1, 2, 3].map(foo) }"),
				},
			},
			{
				Code:   "function foo14() { if (foo) return bar \nelse { baz(); } \n[1, 2, 3].map(foo) }",
				Output: []string{"function foo14() { if (foo) return bar \n baz();  \n[1, 2, 3].map(foo) }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo14() { if (foo) return bar \nelse { baz(); } \n[1, 2, 3].map(foo) }", "{ baz(); }"),
				},
			},
			{
				Code: "function foo15() { if (foo) return bar; else { baz() } qaz() }",
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo15() { if (foo) return bar; else { baz() } qaz() }", "{ baz() }"),
				},
			},
			{
				Code: "function foo16() { if (foo) return bar \nelse { baz() } qaz() }",
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo16() { if (foo) return bar \nelse { baz() } qaz() }", "{ baz() }"),
				},
			},
			{
				Code:   "function foo17() { if (foo) return bar \nelse { baz() } \nqaz() }",
				Output: []string{"function foo17() { if (foo) return bar \n baz()  \nqaz() }"},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo17() { if (foo) return bar \nelse { baz() } \nqaz() }", "{ baz() }"),
				},
			},
			{
				Code: "function foo18() { if (foo) return function() {} \nelse [1, 2, 3].map(bar) }",
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo18() { if (foo) return function() {} \nelse [1, 2, 3].map(bar) }", "[1, 2, 3].map(bar)"),
				},
			},
			{
				Code:    "function foo19() { if (true) { return x; } else if (false) { return y; } }",
				Output:  []string{"function foo19() { if (true) { return x; } if (false) { return y; } }"},
				Options: map[string]interface{}{"allowElseIf": false},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo19() { if (true) { return x; } else if (false) { return y; } }", "if (false) { return y; }"),
				},
			},
			{
				Code:    "function foo20() {if (x) { return true; } else if (y) { notAReturn() } else { notAReturn(); } }",
				Output:  []string{"function foo20() {if (x) { return true; } if (y) { notAReturn() } else { notAReturn(); } }"},
				Options: map[string]interface{}{"allowElseIf": false},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo20() {if (x) { return true; } else if (y) { notAReturn() } else { notAReturn(); } }", "if (y) { notAReturn() } else { notAReturn(); }"),
				},
			},
			{
				Code:    "function foo21() { var x = true; if (x) { return x; } else if (x === false) { return false; } }",
				Output:  []string{"function foo21() { var x = true; if (x) { return x; } if (x === false) { return false; } }"},
				Options: map[string]interface{}{"allowElseIf": false},
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo21() { var x = true; if (x) { return x; } else if (x === false) { return false; } }", "if (x === false) { return false; }"),
				},
			},

			// https://github.com/eslint/eslint/issues/11069
			upstreamInvalid("function foo() { var a; if (bar) { return true; } else { var a; } }", "{ var a; }", "function foo() { var a; if (bar) { return true; }  var a;  }", nil),
			upstreamInvalid("function foo() { if (bar) { var a; if (baz) { return true; } else { var a; } } }", "{ var a; }", "function foo() { if (bar) { var a; if (baz) { return true; }  var a;  } }", nil),
			upstreamInvalid("function foo() { var a; if (bar) { return true; } else { var a; } }", "{ var a; }", "function foo() { var a; if (bar) { return true; }  var a;  }", nil),
			upstreamInvalid("function foo() { if (bar) { var a; if (baz) { return true; } else { var a; } } }", "{ var a; }", "function foo() { if (bar) { var a; if (baz) { return true; }  var a;  } }", nil),
			upstreamInvalidNoFix("function foo() { let a; if (bar) { return true; } else { let a; } }", "{ let a; }"),
			upstreamInvalidNoFix("class foo { bar() { let a; if (baz) { return true; } else { let a; } } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { let a; if (baz) { return true; } else { let a; } } }", "{ let a; }"),
			upstreamInvalid("function foo() {let a; if (bar) { if (baz) { return true; } else { let a; } } }", "{ let a; }", "function foo() {let a; if (bar) { if (baz) { return true; }  let a;  } }", nil),
			upstreamInvalidNoFix("function foo() { const a = 1; if (bar) { return true; } else { let a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { const a = 1; if (baz) { return true; } else { let a; } } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { let a; if (bar) { return true; } else { const a = 1 } }", "{ const a = 1 }"),
			upstreamInvalidNoFix("function foo() { if (bar) { let a; if (baz) { return true; } else { const a = 1; } } }", "{ const a = 1; }"),
			upstreamInvalidNoFix("function foo() { class a {}; if (bar) { return true; } else { const a = 1; } }", "{ const a = 1; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { class a {}; if (baz) { return true; } else { const a = 1; } } }", "{ const a = 1; }"),
			upstreamInvalidNoFix("function foo() { const a = 1; if (bar) { return true; } else { class a {} } }", "{ class a {} }"),
			upstreamInvalidNoFix("function foo() { if (bar) { const a = 1; if (baz) { return true; } else { class a {} } } }", "{ class a {} }"),
			upstreamInvalidNoFix("function foo() { var a; if (bar) { return true; } else { let a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { var a; return true; } else { let a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else { let a; }  while (baz) { var a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo(a) { if (bar) { return true; } else { let a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo(a = 1) { if (bar) { return true; } else { let a; } }", "{ let a; }"),
			{
				Code: "function foo(a, b = a) { if (bar) { return true; } else { let a; }  if (bar) { return true; } else { let b; }}",
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedAt("function foo(a, b = a) { if (bar) { return true; } else { let a; }  if (bar) { return true; } else { let b; }}", "{ let a; }"),
					unexpectedAt("function foo(a, b = a) { if (bar) { return true; } else { let a; }  if (bar) { return true; } else { let b; }}", "{ let b; }"),
				},
			},
			upstreamInvalidNoFix("function foo(...args) { if (bar) { return true; } else { let args; } }", "{ let args; }"),
			upstreamInvalidNoFix("function foo() { try {} catch (a) { if (bar) { return true; } else { let a; } } }", "{ let a; }"),
			upstreamInvalid("function foo() { try {} catch (a) { if (bar) { if (baz) { return true; } else { let a; } } } }", "{ let a; }", "function foo() { try {} catch (a) { if (bar) { if (baz) { return true; }  let a;  } } }", nil),
			upstreamInvalidNoFix("function foo() { try {} catch ({bar, a = 1}) { if (baz) { return true; } else { let a; } } }", "{ let a; }"),
			upstreamInvalid("function foo() { if (bar) { return true; } else { let arguments; } }", "{ let arguments; }", "function foo() { if (bar) { return true; }  let arguments;  }", nil),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else { let arguments; } return arguments[0]; }", "{ let arguments; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else { let arguments; } if (baz) { return arguments[0]; } }", "{ let arguments; }"),
			upstreamInvalid("function foo() { if (bar) { if (baz) { return true; } else { let arguments; } } }", "{ let arguments; }", "function foo() { if (bar) { if (baz) { return true; }  let arguments;  } }", nil),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else { let a; } a; }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else { let a; } if (baz) { a; } }", "{ let a; }"),
			upstreamInvalid("function foo() { if (bar) { if (baz) { return true; } else { let a; } } a; }", "{ let a; }", "function foo() { if (bar) { if (baz) { return true; }  let a;  } a; }", nil),
			upstreamInvalidNoFix("function foo() { if (bar) { if (baz) { return true; } else { let a; } a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { if (baz) { return true; } else { let a; } if (quux) { a; } } }", "{ let a; }"),
			upstreamInvalidNoFix("function a() { if (foo) { return true; } else { let a; } a(); }", "{ let a; }"),
			upstreamInvalidNoFix("function a() { if (a) { return true; } else { let a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function a() { if (foo) { return a; } else { let a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else { let a; } function baz() { a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { if (baz) { return true; } else { let a; } (() => a) } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else { let a; } var a; }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { if (baz) { return true; } else { let a; } var a; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { if (baz) { return true; } else { let a; } var { a } = {}; } }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (bar) { if (baz) { return true; } else { let a; } if (quux) { var a; } } }", "{ let a; }"),
			upstreamInvalid("function foo() { if (bar) { if (baz) { return true; } else { let a; } } if (quux) { var a; } }", "{ let a; }", "function foo() { if (bar) { if (baz) { return true; }  let a;  } if (quux) { var a; } }", nil),
			upstreamInvalid("function foo() { if (quux) { var a; } if (bar) { if (baz) { return true; } else { let a; } } }", "{ let a; }", "function foo() { if (quux) { var a; } if (bar) { if (baz) { return true; }  let a;  } }", nil),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else { let a; } function a(){} }", "{ let a; }"),
			upstreamInvalidNoFix("function foo() { if (baz) { if (bar) { return true; } else { let a; } function a(){} } }", "{ let a; }"),
			upstreamInvalid("function foo() { if (bar) { if (baz) { return true; } else { let a; } } if (quux) { function a(){}  } }", "{ let a; }", "function foo() { if (bar) { if (baz) { return true; }  let a;  } if (quux) { function a(){}  } }", nil),
			upstreamInvalid("function foo() { if (bar) { if (baz) { return true; } else { let a; } } function a(){} }", "{ let a; }", "function foo() { if (bar) { if (baz) { return true; }  let a;  } function a(){} }", nil),
			upstreamInvalidNoFix("function foo() { let a; if (bar) { return true; } else { function a(){} } }", "{ function a(){} }"),
			upstreamInvalidNoFix("function foo() { var a; if (bar) { return true; } else { function a(){} } }", "{ function a(){} }"),
			upstreamInvalidNoFix("function foo() { if (bar) { return true; } else function baz() {} };", "function baz() {}"),
			upstreamInvalid("if (foo) { return true; } else { let a; }", "{ let a; }", "if (foo) { return true; }  let a; ", nil),
			upstreamInvalidNoFix("let a; if (foo) { return true; } else { let a; }", "{ let a; }"),
		},
	)
}

func upstreamInvalid(code, reportSnippet, output string, options any) rule_tester.InvalidTestCase {
	tc := rule_tester.InvalidTestCase{
		Code: code,
		Output: []string{
			output,
		},
		Errors: []rule_tester.InvalidTestCaseError{
			unexpectedAt(code, reportSnippet),
		},
	}
	if options != nil {
		tc.Options = options
	}
	return tc
}

func upstreamInvalidNoFix(code, reportSnippet string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			unexpectedAt(code, reportSnippet),
		},
	}
}

func unexpectedAt(code, snippet string) rule_tester.InvalidTestCaseError {
	start := strings.Index(code, snippet)
	if start < 0 {
		panic("test snippet not found: " + snippet)
	}
	end := start + len(snippet)
	line, column := lineColumn(code, start)
	endLine, endColumn := lineColumn(code, end)
	return rule_tester.InvalidTestCaseError{
		MessageId:   "unexpected",
		Message:     "Unnecessary 'else' after 'return'.",
		Line:        line,
		Column:      column,
		EndLine:     endLine,
		EndColumn:   endColumn,
		Suggestions: nil,
	}
}

func lineColumn(code string, offset int) (int, int) {
	line := 1
	column := 1
	for i := 0; i < offset && i < len(code); i++ {
		if code[i] == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}
