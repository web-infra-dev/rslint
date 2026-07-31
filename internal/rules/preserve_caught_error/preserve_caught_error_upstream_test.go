package preserve_caught_error

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreserveCaughtErrorUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/preserve-caught-error.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in the preserve_caught_error_extras_test.go file.
func TestPreserveCaughtErrorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreserveCaughtErrorRule,
		[]rule_tester.ValidTestCase{
			{Code: `try {
        throw new Error("Original error");
    } catch (error) {
        throw new Error("Failed to perform error prone operations", { cause: error });
    }`},
			{Code: `try {
		doSomething();
	} catch (error) {
		throw new Error("Something failed", { 'cause': error });
	}`},
			{Code: `try {
		doSomething();
	} catch (error) {
		throw new Error("Something failed", { "cause": error });
	}`},
			{Code: `try {
		doSomething();
	} catch (error) {
		throw new Error("Something failed", { ['cause']: error });
	}`},
			{Code: `try {
		doSomething();
	} catch (error) {
		throw new Error("Something failed", { ["cause"]: error });
	}`},
			{Code: "try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error(\"Something failed\", { [`cause`]: error });\n\t}"},
			{Code: `try {
        doSomething();
    } catch (e) {
        console.error(e);
    }`},
			{Code: `try {
        doSomething();
    } catch (err) {
        throw new Error("Failed", { cause: err, extra: 42 });
    }`},
			{Code: `try {
        doSomething();
    } catch (error) {
        switch (error.code) {
            case "A":
                throw new Error("Type A", { cause: error });
            case "B":
                throw new Error("Type B", { cause: error });
            default:
                throw new Error("Other", { cause: error });
        }
    }`},
			{Code: `try {
		// ...
	} catch (err) {
		const opts = { cause: err }
		throw new Error("msg", { ...opts });
	}
	`},
			{Code: `try {
	} catch (error) {
		foo = {
			bar() {
				throw new Error();
			}
		};
	}`},
			{Code: `try {
				doSomething();
			} catch (error) {
				const args = [];
				throw new Error(...args);
		}`},
			{Code: `import { Error } from "./my-custom-error.js";
			try {
				doSomething();
			} catch (error) {
				throw Error("Failed to perform error prone operations");
			}`},
			{Code: `try {
		doSomething();
	} catch {
		throw new Error("Something went wrong");
	}`, Options: map[string]interface{}{"requireCatchParameter": false}},
			{Code: `try {
			doSomething();
		} catch (error) {
			throw new Error("Something failed", { cause: anotherError, cause: error });
		}`},
			{Code: `try {
			doSomething();
		} catch (error) {
			throw new Error("Something failed", { "cause": anotherError, "cause": error });
		}`},
			{Code: `try {
			doSomething();
		} catch (error) {
			throw new Error("Something failed", { cause: anotherError, "cause": error });
		}`},
			{Code: `
			const errors = { AppError: class AppError extends Error {} };
			try {
				doSomething();
			} catch (error) {
				throw new errors.AppError("Something failed", { cause: error });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}}},
			{Code: `
			const lib = { APIError: class APIError extends Error {} };
			try {
				doSomething();
			} catch (error) {
				throw new lib.APIError("Something failed", { cause: error });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{"APIError"}}},
			{Code: `
			class CustomApplicationError extends Error {}
			try {
				doSomething();
			} catch (err) {
				throw new CustomApplicationError("Cause not provided", { cause: err });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{"CustomApplicationError"}}},
			{Code: `
			class APIError extends Error {}
			try {
				doSomething();
			} catch (error) {
				throw new APIError("API failed", { cause: error });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{"APIError"}}},
			{Code: `
			class CustomError extends Error {}
			try {
				doSomething();
			} catch (err) {
				throw new CustomError("No cause");
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{}}},
			{Code: `
			try {
				doSomething();
			} catch (err) {
				throw new MyError("failed", someData, { cause: err });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "MyError", "argumentPosition": 3}}}},
			{Code: `
			try {
				doSomething();
			} catch (err) {
				throw new SimpleError({ cause: err });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "SimpleError", "argumentPosition": 1}}}},
			{Code: `
			try {
				doSomething();
			} catch (err) {
				throw new AppError("Message", {}, { cause: err });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AppError", "argumentPosition": 2}, map[string]interface{}{"name": "AppError", "argumentPosition": 3}}}},
			{Code: `
			try {
			    doSomething();
			} catch (err) {
			    throw new context.Error("failed", someData, { cause: err });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "Error", "argumentPosition": 3}}}},
			{Code: `
			try {
			    doSomething();
			} catch (err) {
			    throw new Error("failed", { cause: err });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "Error", "argumentPosition": 3}}}},
			{Code: `import { AggregateError } from "some-module";
			try {
				doSomething();
			} catch (err) {
				throw new AggregateError({ cause: err });
			}`, Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AggregateError", "argumentPosition": 1}}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `try {
            doSomething();
        } catch (err) {
            throw new Error("Something failed");
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 13, EndLine: 4, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (err) {
            throw new Error("Something failed", { cause: err });
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (err) {
            const unrelated = new Error("other");
            throw new Error("Something failed", { cause: unrelated });
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 5, Column: 58, EndLine: 5, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (err) {
            const unrelated = new Error("other");
            throw new Error("Something failed", { cause: err });
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (err) {
            const unrelated = new Error("other");
            throw new Error("Something failed", { "cause": unrelated });
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 5, Column: 60, EndLine: 5, EndColumn: 69, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (err) {
            const unrelated = new Error("other");
            throw new Error("Something failed", { "cause": err });
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (err) {
            const e = err;
            throw new Error("Failed", { cause: e });
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 5, Column: 48, EndLine: 5, EndColumn: 49, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (err) {
            const e = err;
            throw new Error("Failed", { cause: err });
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (error) {
            throw new Error("Failed", { cause: error.message });
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 4, Column: 48, EndLine: 4, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (error) {
            throw new Error("Failed", { cause: error });
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (error) {
            if (shouldThrow) {
                while (true) {
                    if (Math.random() > 0.5) {
                        throw new Error("Failed without cause");
                    }
                }
            }
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 7, Column: 25, EndLine: 7, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (error) {
            if (shouldThrow) {
                while (true) {
                    if (Math.random() > 0.5) {
                        throw new Error("Failed without cause", { cause: error });
                    }
                }
            }
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (error) {
            switch (error.code) {
                case "A":
                    throw new Error("Type A");
                case "B":
                    throw new Error("Type B", { cause: error });
                default:
                    throw new Error("Other", { cause: error });
            }
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 6, Column: 21, EndLine: 6, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (error) {
            switch (error.code) {
                case "A":
                    throw new Error("Type A", { cause: error });
                case "B":
                    throw new Error("Type B", { cause: error });
                default:
                    throw new Error("Other", { cause: error });
            }
        }`}}}},
			},
			{
				Code:   "try {\n            doSomething();\n        } catch (error) {\n            throw new Error(`The certificate key \"${chalk.yellow(keyFile)}\" is invalid.\n${err.message}`);\n        }",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 13, EndLine: 5, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: "try {\n            doSomething();\n        } catch (error) {\n            throw new Error(`The certificate key \"${chalk.yellow(keyFile)}\" is invalid.\n${err.message}`, { cause: error });\n        }"}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (error) {
            const errorMessage = "Operation failed";
            throw new Error(errorMessage);
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 13, EndLine: 5, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (error) {
            const errorMessage = "Operation failed";
            throw new Error(errorMessage, { cause: error });
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (error) {
            const errorMessage = "Operation failed";
            throw new Error(errorMessage, { existingOption: true, complexOption: { moreOptions: {} } });
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 13, EndLine: 5, EndColumn: 105, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (error) {
            const errorMessage = "Operation failed";
            throw new Error(errorMessage, { existingOption: true, complexOption: { moreOptions: {} }, cause: error });
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (err) {
            if (err.code === "A") {
                throw new Error("Type A");
            }
            throw new TypeError("Fallback error");
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 17, EndLine: 5, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (err) {
            if (err.code === "A") {
                throw new Error("Type A", { cause: err });
            }
            throw new TypeError("Fallback error");
        }`}}}, {MessageId: "missingCause", Line: 7, Column: 13, EndLine: 7, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (err) {
            if (err.code === "A") {
                throw new Error("Type A");
            }
            throw new TypeError("Fallback error", { cause: err });
        }`}}}},
			},
			{
				Code: `try {
            doSomething();
        } catch (err) {
            throw Error("Something failed");
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 13, EndLine: 4, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (err) {
            throw Error("Something failed", { cause: err });
        }`}}}},
			},
			{
				Code: `try {
        } catch (err) {
            my_label:
            throw new Error("Failed without cause");
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 13, EndLine: 4, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
        } catch (err) {
            my_label:
            throw new Error("Failed without cause", { cause: err });
        }`}}}},
			},
			{
				Code: `try {
        } catch (err) {
            {
                throw new Error("Something went wrong");
            }
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 17, EndLine: 4, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
        } catch (err) {
            {
                throw new Error("Something went wrong", { cause: err });
            }
        }`}}}},
			},
			{
				Code: `try {
        } catch (err) {
            {
                throw new Error();
            }
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 17, EndLine: 4, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
        } catch (err) {
            {
                throw new Error("", { cause: err });
            }
        }`}}}},
			},
			{
				Code: `try {
        } catch (err) {
            {
                throw new AggregateError([], "Lorem ipsum");
            }
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 17, EndLine: 4, EndColumn: 61, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
        } catch (err) {
            {
                throw new AggregateError([], "Lorem ipsum", { cause: err });
            }
        }`}}}},
			},
			{
				Code: `try {
        } catch (err) {
            {
                throw new AggregateError();
            }
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 17, EndLine: 4, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
        } catch (err) {
            {
                throw new AggregateError([], "", { cause: err });
            }
        }`}}}},
			},
			{
				Code: `try {
        } catch (err) {
            {
                throw new AggregateError([]);
            }
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 17, EndLine: 4, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
        } catch (err) {
            {
                throw new AggregateError([], "", { cause: err });
            }
        }`}}}},
			},
			{
				Code: `try {
			doSomething();
		} catch {
			throw new Error("Something went wrong");
		}`,
				Options: map[string]interface{}{"requireCatchParameter": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCatchErrorParam", Line: 4, Column: 4, EndLine: 4, EndColumn: 44}},
			},
			{
				Code: `try {
            doSomething();
        } catch (err) {
            throw new Error("Something failed", { cause });
        }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 4, Column: 51, EndLine: 4, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
            doSomething();
        } catch (err) {
            throw new Error("Something failed", { cause: err });
        }`}}}},
			},
			{
				Code: `try {
				doSomething();
			} catch ({ message }) {
				throw new Error(message);
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "partiallyLostError", Line: 3, Column: 6, EndLine: 5, EndColumn: 5}},
			},
			{
				Code: `try {
				doSomethingElse();
			} catch ({ ...error }) {
				throw new Error(error.message);
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "partiallyLostError", Line: 3, Column: 6, EndLine: 5, EndColumn: 5}},
			},
			{
				Code: `try {
				doSomething();
			} catch (error) {
				if (whatever) {
					const error = anotherError;
					throw new Error("Something went wrong", { cause: error });
				}
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "caughtErrorShadowed", Line: 6, Column: 6, EndLine: 6, EndColumn: 64}},
			},
			{
				Code: `try {
				doSomething();
			} catch (error) {
				throw new Error(
					"Something went wrong" // some comments
				);
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 5, EndLine: 6, EndColumn: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
				doSomething();
			} catch (error) {
				throw new Error(
					"Something went wrong", { cause: error } // some comments
				);
			}`}}}},
			},
			{
				Code: `try {
				doSomething();
			} catch (err) {
				throw new Error("Something failed", {});
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 4, Column: 5, EndLine: 4, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
				doSomething();
			} catch (err) {
				throw new Error("Something failed", {cause: err});
			}`}}}},
			},
			{
				Code: `try {
			doSomething();
		} catch (error) {
			const cause = "desc";
			throw new Error("Something failed", { [cause]: "Some error" });
		}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 4, EndLine: 5, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
			doSomething();
		} catch (error) {
			const cause = "desc";
			throw new Error("Something failed", { [cause]: "Some error", cause: error });
		}`}}}},
			},
			{
				Code: `try {
			doSomething();
			} catch (error) {
			throw new Error("Something failed", { cause() { /* do something */ }  });
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 4, Column: 47, EndLine: 4, EndColumn: 72, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {
			doSomething();
			} catch (error) {
			throw new Error("Something failed", { cause: error  });
			}`}}}},
			},
			{
				Code: `try {} catch (error) {
				throw new Error("Something failed", { cause: error, cause: anotherError });
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 2, Column: 64, EndLine: 2, EndColumn: 76}},
			},
			{
				Code: `try {} catch (error) {
				throw new Error("Something failed", { cause: error, "cause": anotherError });
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 2, Column: 66, EndLine: 2, EndColumn: 78}},
			},
			{
				Code: `try {} catch (error) {
				throw new Error("Something failed", { get cause() { } });
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 2, Column: 52, EndLine: 2, EndColumn: 58, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (error) {
				throw new Error("Something failed", { cause: error });
			}`}}}},
			},
			{
				Code: `try {} catch (error) {
				throw new Error("Something failed", { set cause(value) { } });
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 2, Column: 52, EndLine: 2, EndColumn: 63, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (error) {
				throw new Error("Something failed", { cause: error });
			}`}}}},
			},
			{
				Code: `try {} catch (error) {
				throw new Error("Something failed", {
					get cause() { return error; },
					set cause(value) { error = value; },
				});
			}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 4, Column: 15, EndLine: 4, EndColumn: 41}},
			},
			{
				Code:   `try { doSomething(); } catch (err) { throw new Error(("Something failed")); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 38, EndLine: 1, EndColumn: 76, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try { doSomething(); } catch (err) { throw new Error(("Something failed"), { cause: err }); }`}}}},
			},
			{
				Code:   `try { doSomething(); } catch (err) { throw new Error(("Something failed"),); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 38, EndLine: 1, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try { doSomething(); } catch (err) { throw new Error(("Something failed"), { cause: err },); }`}}}},
			},
			{
				Code:   `try { doSomething(); } catch (err) { throw new AggregateError((errors)); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 38, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try { doSomething(); } catch (err) { throw new AggregateError((errors), "", { cause: err }); }`}}}},
			},
			{
				Code:   `try { doSomething(); } catch (err) { throw new AggregateError(errors, ("message")); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 38, EndLine: 1, EndColumn: 84, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try { doSomething(); } catch (err) { throw new AggregateError(errors, ("message"), { cause: err }); }`}}}},
			},
			{
				Code:    `try { doSomething(); } catch (err) { throw new CustomError((foo)); }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"CustomError"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 38, EndLine: 1, EndColumn: 67, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try { doSomething(); } catch (err) { throw new CustomError((foo), { cause: err }); }`}}}},
			},
			{
				Code: `
			class CustomApplicationError extends Error {}
			try {
				doSomething();
			} catch (err) {
				throw new CustomApplicationError("Cause not provided");
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"CustomApplicationError"}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 6, Column: 5, EndLine: 6, EndColumn: 60, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			class CustomApplicationError extends Error {}
			try {
				doSomething();
			} catch (err) {
				throw new CustomApplicationError("Cause not provided", { cause: err });
			}`}}}},
			},
			{
				Code: `
			class APIError extends Error {}
			try {
				doSomething();
			} catch (error) {
				throw new APIError("API failed", { cause: wrong });
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"APIError"}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 6, Column: 47, EndLine: 6, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			class APIError extends Error {}
			try {
				doSomething();
			} catch (error) {
				throw new APIError("API failed", { cause: error });
			}`}}}},
			},
			{
				Code: `
			const errors = { AppError: class AppError extends Error {} };
			try {
				doSomething();
			} catch (error) {
				throw new errors.AppError("Something failed");
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"AppError"}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 6, Column: 5, EndLine: 6, EndColumn: 51, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			const errors = { AppError: class AppError extends Error {} };
			try {
				doSomething();
			} catch (error) {
				throw new errors.AppError("Something failed", { cause: error });
			}`}}}},
			},
			{
				Code: `
			const lib = { APIError: class APIError extends Error {} };
			try {
				doSomething();
			} catch (error) {
				throw new lib.APIError("Something failed", { cause: wrong });
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"APIError"}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 6, Column: 57, EndLine: 6, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			const lib = { APIError: class APIError extends Error {} };
			try {
				doSomething();
			} catch (error) {
				throw new lib.APIError("Something failed", { cause: error });
			}`}}}},
			},
			{
				Code: `
			try {
				doSomething();
			} catch (err) {
				throw new MyError("failed", someData);
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "MyError", "argumentPosition": 3}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 5, EndLine: 5, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			try {
				doSomething();
			} catch (err) {
				throw new MyError("failed", someData, { cause: err });
			}`}}}},
			},
			{
				Code: `
			try {
				doSomething();
			} catch (err) {
				throw new MyError("failed", someData, { cause: wrong });
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "MyError", "argumentPosition": 3}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 5, Column: 52, EndLine: 5, EndColumn: 57, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			try {
				doSomething();
			} catch (err) {
				throw new MyError("failed", someData, { cause: err });
			}`}}}},
			},
			{
				Code: `
			try {
				doSomething();
			} catch (err) {
				throw new SimpleError();
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "SimpleError", "argumentPosition": 1}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 5, EndLine: 5, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			try {
				doSomething();
			} catch (err) {
				throw new SimpleError({ cause: err });
			}`}}}},
			},
			{
				Code: `
			try {
				doSomething();
			} catch (err) {
				throw new MyError();
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{"MyError"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 5, EndLine: 5, EndColumn: 25}},
			},
			{
				Code: `
			try {
			    doSomething();
			} catch (err) {
			    throw new context.Error("failed", someData);
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "Error", "argumentPosition": 3}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 8, EndLine: 5, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			try {
			    doSomething();
			} catch (err) {
			    throw new context.Error("failed", someData, { cause: err });
			}`}}}},
			},
			{
				Code: `
			try {
			    doSomething();
			} catch (err) {
			    throw new Error("failed");
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "Error", "argumentPosition": 3}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 8, EndLine: 5, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `
			try {
			    doSomething();
			} catch (err) {
			    throw new Error("failed", { cause: err });
			}`}}}},
			},
			{
				Code: `import { AggregateError } from "some-module";
			try {
				doSomething();
			} catch (err) {
				throw new AggregateError();
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AggregateError", "argumentPosition": 1}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 5, Column: 5, EndLine: 5, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `import { AggregateError } from "some-module";
			try {
				doSomething();
			} catch (err) {
				throw new AggregateError({ cause: err });
			}`}}}},
			},
			{
				Code: `import { AggregateError } from "some-module";
			try {
				doSomething();
			} catch (err) {
				throw new AggregateError({ cause: wrong });
			}`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "AggregateError", "argumentPosition": 1}}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "incorrectCause", Line: 5, Column: 39, EndLine: 5, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `import { AggregateError } from "some-module";
			try {
				doSomething();
			} catch (err) {
				throw new AggregateError({ cause: err });
			}`}}}},
			},
			{
				Code:   `try {} catch (err) { throw new Error; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new Error("", { cause: err }); }`}}}},
			},
			{
				Code:   `try {} catch (err) { throw new AggregateError; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new AggregateError([], "", { cause: err }); }`}}}},
			},
			{
				Code:    `try {} catch (err) { throw new CustomError; }`,
				Options: map[string]interface{}{"errorClassNames": []interface{}{map[string]interface{}{"name": "CustomError", "argumentPosition": 1}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new CustomError({ cause: err }); }`}}}},
			},
			{
				Code:   `try {} catch (err) { throw new (Error); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCause", Line: 1, Column: 22, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "includeCause", Output: `try {} catch (err) { throw new (Error)("", { cause: err }); }`}}}},
			},
		},
	)
}
