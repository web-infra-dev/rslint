// TestNoReturnAwaitUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/no-return-await.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in the
// no_return_await_extras_test.go file.
package no_return_await

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoReturnAwaitUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoReturnAwaitRule,
		[]rule_tester.ValidTestCase{
			{Code: `
async function foo() {
	await bar(); return;
}
`},
			{Code: `
async function foo() {
	const x = await bar(); return x;
}
`},
			{Code: `
async () => { return bar(); }
`},
			{Code: `
async () => bar()
`},
			{Code: `
async function foo() {
if (a) {
		if (b) {
			return bar();
		}
	}
}
`},
			{Code: `
async () => {
if (a) {
		if (b) {
			return bar();
		}
	}
}
`},
			{Code: `
async function foo() {
	return (await bar() && a);
}
`},
			{Code: `
async function foo() {
	return (await bar() || a);
}
`},
			{Code: `
async function foo() {
	return (a && await baz() && b);
}
`},
			{Code: `
async function foo() {
	return (await bar(), a);
}
`},
			{Code: `
async function foo() {
	return (await baz(), await bar(), a);
}
`},
			{Code: `
async function foo() {
	return (a, b, (await bar(), c));
}
`},
			{Code: `
async function foo() {
	return (await bar() ? a : b);
}
`},
			{Code: `
async function foo() {
	return ((a && await bar()) ? b : c);
}
`},
			{Code: `
async function foo() {
	return (baz() ? (await bar(), a) : b);
}
`},
			{Code: `
async function foo() {
	return (baz() ? (await bar() && a) : b);
}
`},
			{Code: `
async function foo() {
	return (baz() ? a : (await bar(), b));
}
`},
			{Code: `
async function foo() {
	return (baz() ? a : (await bar() && b));
}
`},
			{Code: `
async () => (await bar(), a)
`},
			{Code: `
async () => (await bar() && a)
`},
			{Code: `
async () => (await bar() || a)
`},
			{Code: `
async () => (a && await bar() && b)
`},
			{Code: `
async () => (await baz(), await bar(), a)
`},
			{Code: `
async () => (a, b, (await bar(), c))
`},
			{Code: `
async () => (await bar() ? a : b)
`},
			{Code: `
async () => ((a && await bar()) ? b : c)
`},
			{Code: `
async () => (baz() ? (await bar(), a) : b)
`},
			{Code: `
async () => (baz() ? (await bar() && a) : b)
`},
			{Code: `
async () => (baz() ? a : (await bar(), b))
`},
			{Code: `
async () => (baz() ? a : (await bar() && b))
`},
			{Code: `
          async function foo() {
            try {
              return await bar();
            } catch (e) {
              baz();
            }
          }
        `},
			{Code: `
          async function foo() {
            try {
              return await bar();
            } finally {
              baz();
            }
          }
        `},
			{Code: `
          async function foo() {
            try {}
            catch (e) {
              return await bar();
            } finally {
              baz();
            }
          }
        `},
			{Code: `
          async function foo() {
            try {
              try {}
              finally {
                return await bar();
              }
            } finally {
              baz();
            }
          }
        `},
			{Code: `
          async function foo() {
            try {
              try {}
              catch (e) {
                return await bar();
              }
            } finally {
              baz();
            }
          }
        `},
			{Code: `
          async function foo() {
            try {
              return (a, await bar());
            } catch (e) {
              baz();
            }
          }
        `},
			{Code: `
          async function foo() {
            try {
              return (qux() ? await bar() : b);
            } catch (e) {
              baz();
            }
          }
        `},
			{Code: `
          async function foo() {
            try {
              return (a && await bar());
            } catch (e) {
              baz();
            }
          }
        `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
async function foo() {
	return await bar();
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    9,
						EndLine:   3,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return bar();
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return await(bar());
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    9,
						EndLine:   3,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (bar());
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (a, await bar());
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    13,
						EndLine:   3,
						EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (a, bar());
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (a, b, await bar());
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    16,
						EndLine:   3,
						EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (a, b, bar());
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (a && await bar());
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    15,
						EndLine:   3,
						EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (a && bar());
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (a && b && await bar());
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    20,
						EndLine:   3,
						EndColumn: 31,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (a && b && bar());
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (a || await bar());
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    15,
						EndLine:   3,
						EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (a || bar());
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (a, b, (c, d, await bar()));
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    23,
						EndLine:   3,
						EndColumn: 34,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (a, b, (c, d, bar()));
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (a, b, (c && await bar()));
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    22,
						EndLine:   3,
						EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (a, b, (c && bar()));
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (await baz(), b, await bar());
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    26,
						EndLine:   3,
						EndColumn: 37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (await baz(), b, bar());
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (baz() ? await bar() : b);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    18,
						EndLine:   3,
						EndColumn: 29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (baz() ? bar() : b);
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (baz() ? a : await bar());
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    22,
						EndLine:   3,
						EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (baz() ? a : bar());
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (baz() ? (a, await bar()) : b);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    22,
						EndLine:   3,
						EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (baz() ? (a, bar()) : b);
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (baz() ? a : (b, await bar()));
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    26,
						EndLine:   3,
						EndColumn: 37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (baz() ? a : (b, bar()));
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (baz() ? (a && await bar()) : b);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    24,
						EndLine:   3,
						EndColumn: 35,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (baz() ? (a && bar()) : b);
}
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
	return (baz() ? a : (b && await bar()));
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    28,
						EndLine:   3,
						EndColumn: 39,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
	return (baz() ? a : (b && bar()));
}
`},
						},
					},
				},
			},
			{
				Code: `
async () => { return await bar(); }
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async () => { return bar(); }
`},
						},
					},
				},
			},
			{
				Code: `
async () => await bar()
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      2,
						Column:    13,
						EndLine:   2,
						EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async () => bar()
`},
						},
					},
				},
			},
			{
				Code: `
async () => (a, b, await bar())
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      2,
						Column:    20,
						EndLine:   2,
						EndColumn: 31,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async () => (a, b, bar())
`},
						},
					},
				},
			},
			{
				Code: `
async () => (a && await bar())
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      2,
						Column:    19,
						EndLine:   2,
						EndColumn: 30,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async () => (a && bar())
`},
						},
					},
				},
			},
			{
				Code: `
async () => (baz() ? await bar() : b)
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      2,
						Column:    22,
						EndLine:   2,
						EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async () => (baz() ? bar() : b)
`},
						},
					},
				},
			},
			{
				Code: `
async () => (baz() ? a : (b, await bar()))
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      2,
						Column:    30,
						EndLine:   2,
						EndColumn: 41,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async () => (baz() ? a : (b, bar()))
`},
						},
					},
				},
			},
			{
				Code: `
async () => (baz() ? a : (b && await bar()))
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      2,
						Column:    32,
						EndLine:   2,
						EndColumn: 43,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async () => (baz() ? a : (b && bar()))
`},
						},
					},
				},
			},
			{
				Code: `
async function foo() {
if (a) {
		if (b) {
			return await bar();
		}
	}
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      5,
						Column:    11,
						EndLine:   5,
						EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async function foo() {
if (a) {
		if (b) {
			return bar();
		}
	}
}
`},
						},
					},
				},
			},
			{
				Code: `
async () => {
if (a) {
		if (b) {
			return await bar();
		}
	}
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      5,
						Column:    11,
						EndLine:   5,
						EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
async () => {
if (a) {
		if (b) {
			return bar();
		}
	}
}
`},
						},
					},
				},
			},
			{
				Code: `
              async function foo() {
                try {}
                finally {
                  return await bar();
                }
              }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      5,
						Column:    26,
						EndLine:   5,
						EndColumn: 37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
              async function foo() {
                try {}
                finally {
                  return bar();
                }
              }
            `},
						},
					},
				},
			},
			{
				Code: `
              async function foo() {
                try {}
                catch (e) {
                  return await bar();
                }
              }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      5,
						Column:    26,
						EndLine:   5,
						EndColumn: 37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
              async function foo() {
                try {}
                catch (e) {
                  return bar();
                }
              }
            `},
						},
					},
				},
			},
			{
				Code: `
              try {
                async function foo() {
                  return await bar();
                }
              } catch (e) {}
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      4,
						Column:    26,
						EndLine:   4,
						EndColumn: 37,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
              try {
                async function foo() {
                  return bar();
                }
              } catch (e) {}
            `},
						},
					},
				},
			},
			{
				Code: `
              try {
                async () => await bar();
              } catch (e) {}
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    29,
						EndLine:   3,
						EndColumn: 40,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
              try {
                async () => bar();
              } catch (e) {}
            `},
						},
					},
				},
			},
			{
				Code: `
              async function foo() {
                try {}
                catch (e) {
                  try {}
                  catch (e) {
                    return await bar();
                  }
                }
              }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      7,
						Column:    28,
						EndLine:   7,
						EndColumn: 39,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
              async function foo() {
                try {}
                catch (e) {
                  try {}
                  catch (e) {
                    return bar();
                  }
                }
              }
            `},
						},
					},
				},
			},
			{
				Code: `
              async function foo() {
                return await new Promise(resolve => {
                  resolve(5);
                });
              }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    24,
						EndLine:   5,
						EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
              async function foo() {
                return new Promise(resolve => {
                  resolve(5);
                });
              }
            `},
						},
					},
				},
			},
			{
				Code: `
              async () => {
                return await (
                  foo()
                )
              };
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    24,
						EndLine:   5,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "removeAwait", Output: `
              async () => {
                return (
                  foo()
                )
              };
            `},
						},
					},
				},
			},
			{
				Code: `
              async function foo() {
                return await // Test
                  5;
              }
            `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redundantUseOfAwait",
						Line:      3,
						Column:    24,
						EndLine:   4,
						EndColumn: 20,
					},
				},
			},
		},
	)
}
