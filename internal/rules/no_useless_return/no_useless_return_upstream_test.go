package no_useless_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUselessReturnUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-useless-return.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in the
// no_useless_return_extras_test.go file.
func TestNoUselessReturnUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessReturnRule,
		[]rule_tester.ValidTestCase{
			{Code: `function foo() { return 5; }`},
			{Code: `function foo() { return null; }`},
			{Code: `function foo() { return doSomething(); }`},
			{Code: `
          function foo() {
            if (bar) {
              doSomething();
              return;
            } else {
              doSomethingElse();
            }
            qux();
          }
        `},
			{Code: `
          function foo() {
            switch (bar) {
              case 1:
                doSomething();
                return;
              default:
                doSomethingElse();
            }
          }
        `},
			{Code: `
          function foo() {
            switch (bar) {
              default:
                doSomething();
                return;
              case 1:
                doSomethingElse();
            }
          }
        `},
			{Code: `
          function foo() {
            switch (bar) {
              case 1:
                if (a) {
                  doSomething();
                  return;
                } else {
                  doSomething();
                  return;
                }
              default:
                doSomethingElse();
            }
          }
        `},
			{Code: `
          function foo() {
            for (var foo = 0; foo < 10; foo++) {
              return;
            }
          }
        `},
			{Code: `
          function foo() {
            for (var foo in bar) {
              return;
            }
          }
        `},
			{Code: `
          function foo() {
            try {
              return 5;
            } finally {
              return; // This is allowed because it can override the returned value of 5
            }
          }
        `},
			{Code: `
          function foo() {
            try {
              bar();
              return;
            } catch (err) {}
            baz();
          }
        `},
			{Code: `
          function foo() {
              if (something) {
                  try {
                      bar();
                      return;
                  } catch (err) {}
              }
              baz();
          }
        `},
			{Code: `
          function foo() {
            return;
            doSomething();
          }
        `},
			{Code: `
              function foo() {
                for (var foo of bar) return;
              }
            `},
			{Code: `() => { if (foo) return; bar(); }`},
			{Code: `() => 5`},
			{Code: `() => { return; doSomething(); }`},
			{Code: `if (foo) { return; } doSomething();`},

			// https://github.com/eslint/eslint/issues/7477
			{Code: `
          function foo() {
            if (bar) return;
            return baz;
          }
        `},
			{Code: `
          function foo() {
            if (bar) {
              return;
            }
            return baz;
          }
        `},
			{Code: `
          function foo() {
            if (bar) baz();
            else return;
            return 5;
          }
        `},

			// https://github.com/eslint/eslint/issues/7583
			{Code: `
          function foo() {
            return;
            while (foo) return;
            foo;
          }
        `},

			// https://github.com/eslint/eslint/issues/7855
			{Code: `
          try {
            throw new Error('foo');
            while (false);
          } catch (err) {}
        `},

			// https://github.com/eslint/eslint/issues/11647
			{Code: `
          function foo(arg) {
            throw new Error("Debugging...");
            if (!arg) {
              return;
            }
            console.log(arg);
          }
        `},

			// https://github.com/eslint/eslint/pull/16996#discussion_r1138622844
			{Code: `
        function foo() {
          try {
              bar();
              return;
          } finally {
              baz();
          }
          qux();
        }
        `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `function foo() { return; }`,
				Output: []string{`function foo() {  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Message: "Unnecessary return statement.", Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:   `function foo() { doSomething(); return; }`,
				Output: []string{`function foo() { doSomething();  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 33, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:   `function foo() { if (condition) { bar(); return; } else { baz(); } }`,
				Output: []string{`function foo() { if (condition) { bar();  } else { baz(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 42, EndLine: 1, EndColumn: 49},
				},
			},
			{
				// Removing the return would leave the `if` without a body, so
				// there is no fix.
				Code: `function foo() { if (foo) return; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 27, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: `function foo() { bar(); return/**/; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: `function foo() { bar(); return//
; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 25, EndLine: 2, EndColumn: 2},
				},
			},
			{
				Code:   `foo(); return;`,
				Output: []string{`foo(); `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   `if (foo) { bar(); return; } else { baz(); }`,
				Output: []string{`if (foo) { bar();  } else { baz(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code: `
              function foo() {
                if (foo) {
                  return;
                }
                return;
              }
            `,
				// Each fix covers the whole function, so the second one waits
				// for the pass after the first.
				Output: []string{`
              function foo() {
                if (foo) {
                  
                }
                return;
              }
            `, `
              function foo() {
                if (foo) {
                  
                }
                
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 4, Column: 19, EndLine: 4, EndColumn: 26},
					{MessageId: "unnecessaryReturn", Line: 6, Column: 17, EndLine: 6, EndColumn: 24},
				},
			},
			{
				Code: `
              function foo() {
                switch (bar) {
                  case 1:
                    doSomething();
                  default:
                    doSomethingElse();
                    return;
                }
              }
            `,
				Output: []string{`
              function foo() {
                switch (bar) {
                  case 1:
                    doSomething();
                  default:
                    doSomethingElse();
                    
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 8, Column: 21, EndLine: 8, EndColumn: 28},
				},
			},
			{
				Code: `
              function foo() {
                switch (bar) {
                  default:
                    doSomething();
                  case 1:
                    doSomething();
                    return;
                }
              }
            `,
				Output: []string{`
              function foo() {
                switch (bar) {
                  default:
                    doSomething();
                  case 1:
                    doSomething();
                    
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 8, Column: 21, EndLine: 8, EndColumn: 28},
				},
			},
			{
				Code: `
              function foo() {
                switch (bar) {
                  case 1:
                    if (a) {
                      doSomething();
                      return;
                    }
                    break;
                  default:
                    doSomethingElse();
                }
              }
            `,
				Output: []string{`
              function foo() {
                switch (bar) {
                  case 1:
                    if (a) {
                      doSomething();
                      
                    }
                    break;
                  default:
                    doSomethingElse();
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 7, Column: 23, EndLine: 7, EndColumn: 30},
				},
			},
			{
				Code: `
              function foo() {
                switch (bar) {
                  case 1:
                    if (a) {
                      doSomething();
                      return;
                    } else {
                      doSomething();
                    }
                    break;
                  default:
                    doSomethingElse();
                }
              }
            `,
				Output: []string{`
              function foo() {
                switch (bar) {
                  case 1:
                    if (a) {
                      doSomething();
                      
                    } else {
                      doSomething();
                    }
                    break;
                  default:
                    doSomethingElse();
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 7, Column: 23, EndLine: 7, EndColumn: 30},
				},
			},
			{
				Code: `
              function foo() {
                switch (bar) {
                  case 1:
                    if (a) {
                      doSomething();
                      return;
                    }
                  default:
                }
              }
            `,
				Output: []string{`
              function foo() {
                switch (bar) {
                  case 1:
                    if (a) {
                      doSomething();
                      
                    }
                  default:
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 7, Column: 23, EndLine: 7, EndColumn: 30},
				},
			},
			{
				Code: `
              function foo() {
                try {} catch (err) { return; }
              }
            `,
				Output: []string{`
              function foo() {
                try {} catch (err) {  }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 3, Column: 38, EndLine: 3, EndColumn: 45},
				},
			},
			{
				Code: `
              function foo() {
                try {
                  foo();
                  return;
                } catch (err) {
                  return 5;
                }
              }
            `,
				Output: []string{`
              function foo() {
                try {
                  foo();
                  
                } catch (err) {
                  return 5;
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 5, Column: 19, EndLine: 5, EndColumn: 26},
				},
			},
			{
				Code: `
              function foo() {
                  if (something) {
                      try {
                          bar();
                          return;
                      } catch (err) {}
                  }
              }
            `,
				Output: []string{`
              function foo() {
                  if (something) {
                      try {
                          bar();
                          
                      } catch (err) {}
                  }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 6, Column: 27, EndLine: 6, EndColumn: 34},
				},
			},
			{
				Code: `
              function foo() {
                try {
                  return;
                } catch (err) {
                  foo();
                }
              }
            `,
				Output: []string{`
              function foo() {
                try {
                  
                } catch (err) {
                  foo();
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 4, Column: 19, EndLine: 4, EndColumn: 26},
				},
			},
			{
				Code: `
              function foo() {
                  try {
                      return;
                  } finally {
                      bar();
                  }
              }
            `,
				Output: []string{`
              function foo() {
                  try {
                      
                  } finally {
                      bar();
                  }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 4, Column: 23, EndLine: 4, EndColumn: 30},
				},
			},
			{
				Code: `
              function foo() {
                try {
                  bar();
                } catch (e) {
                  try {
                    baz();
                    return;
                  } catch (e) {
                    qux();
                  }
                }
              }
            `,
				Output: []string{`
              function foo() {
                try {
                  bar();
                } catch (e) {
                  try {
                    baz();
                    
                  } catch (e) {
                    qux();
                  }
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 8, Column: 21, EndLine: 8, EndColumn: 28},
				},
			},
			{
				Code: `
              function foo() {
                try {} finally {}
                return;
              }
            `,
				Output: []string{`
              function foo() {
                try {} finally {}
                
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 4, Column: 17, EndLine: 4, EndColumn: 24},
				},
			},
			{
				Code: `
              function foo() {
                try {
                  return 5;
                } finally {
                  function bar() {
                    return;
                  }
                }
              }
            `,
				Output: []string{`
              function foo() {
                try {
                  return 5;
                } finally {
                  function bar() {
                    
                  }
                }
              }
            `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 7, Column: 21, EndLine: 7, EndColumn: 28},
				},
			},
			{
				Code:   `() => { return; }`,
				Output: []string{`() => {  }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 9, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:   `function foo() { return; return; }`,
				Output: []string{`function foo() {  return; }`, `function foo() {   }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryReturn", Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
				},
			},
		},
	)
}
