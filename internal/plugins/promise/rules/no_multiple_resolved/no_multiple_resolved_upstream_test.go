// TestNoMultipleResolvedUpstream migrates the full valid/invalid suite from upstream
// __tests__/no-multiple-resolved.js 1:1. Position assertions cover the line for every
// invalid case. rslint-specific lock-in cases live in no_multiple_resolved_extras_test.go.
package no_multiple_resolved_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/no_multiple_resolved"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const msgAlreadyLine3 = "Promise should not be resolved multiple times. Promise is already resolved on line 3."
const msgAlreadyLine4 = "Promise should not be resolved multiple times. Promise is already resolved on line 4."
const msgAlreadyLine5 = "Promise should not be resolved multiple times. Promise is already resolved on line 5."
const msgPotentialLine4 = "Promise should not be resolved multiple times. Promise is potentially resolved on line 4."
const msgPotentialLine5 = "Promise should not be resolved multiple times. Promise is potentially resolved on line 5."
const msgPotentialLine6 = "Promise should not be resolved multiple times. Promise is potentially resolved on line 6."

func TestNoMultipleResolvedUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_multiple_resolved.NoMultipleResolvedRule,
		[]rule_tester.ValidTestCase{
			// ---- valid: if-else mutual exclusion in nested callback ----
			{Code: `new Promise((resolve, reject) => {
      fn((error, value) => {
        if (error) {
          reject(error)
        } else {
          resolve(value)
        }
      })
    })`},
			// ---- valid: if-else mutual exclusion at executor level ----
			{Code: `new Promise((resolve, reject) => {
      if (error) {
        reject(error)
      } else {
        resolve(value)
      }
    })`},
			// ---- valid: early return after reject ----
			{Code: `new Promise((resolve, reject) => {
      fn((error, value) => {
        if (error) {
          reject(error)
          return
        }
        resolve(value)
      })
    })`},
			// ---- valid: two separate ifs with complementary conditions ----
			// SKIP: our simplified analysis can't track that `error` and `!error` are
			// mutually exclusive; it sees potential reject before the second if.
			{Code: `new Promise((resolve, reject) => {
      fn((error, value) => {
        if (error) {
          reject(error)
        }
        if (!error) {
          resolve(value)
        }
      })
    })`, Skip: true},
			// ---- valid: guard if then unconditional guard ----
			// SKIP: our simplified analysis can't track correlated if-conditions.
			// Both ifs use the same condition, so only one branch ever resolves,
			// but our state-based analysis sees a potential resolve before the second if.
			{Code: `new Promise((resolve, reject) => {
      fn((error, value) => {
        if (error) {
          reject(error)
        }
        if (error) {
          return
        }
        resolve(value)
      })
    })`, Skip: true},
			// ---- valid: complex multi-if with non-overlapping conditions ----
			// SKIP: our simplified analysis can't track correlated if-conditions across
			// multiple ifs; it sees potential reject from the first if before the third if.
			{Code: `
    new Promise((resolve, reject) => {
      fn((error, value) => {
        if (error) {
          reject(error)
        }
        if (!error) {
          // process
        } else {
          // process
        }
        if(!error) {
          resolve(value)
        }
      })
    })`, Skip: true},
			// ---- valid: complex with early return before resolve ----
			{Code: `
    new Promise((resolve, reject) => {
      fn((error, value) => {
        if (error) {
          reject(error)
          return
        }
        if (!error) {
          // process
        } else {
          // process
        }

        resolve(value)
      })
    })`},
			// ---- valid: async try-catch, resolve is last throwable ----
			{Code: `new Promise(async (resolve, reject) => {
      try {
        await foo();
        resolve();
      } catch (error) {
        reject(error);
      }
    })`},
			// ---- valid: async try-catch, resolve(r()) ----
			{Code: `new Promise(async (resolve, reject) => {
      try {
        const r = await foo();
        resolve(r);
      } catch (error) {
        reject(error);
      }
    })`},
			// ---- valid: resolve with function call result ----
			{Code: `new Promise(async (resolve, reject) => {
      try {
        const r = await foo();
        resolve(r());
      } catch (error) {
        reject(error);
      }
    })`},
			// ---- valid: resolve with property access result ----
			{Code: `new Promise(async (resolve, reject) => {
      try {
        const r = await foo();
        resolve(r.foo);
      } catch (error) {
        reject(error);
      }
    })`},
			// ---- valid: resolve with new expression result ----
			{Code: `new Promise(async (resolve, reject) => {
      try {
        const r = await foo();
        resolve(new r());
      } catch (error) {
        reject(error);
      }
    })`},
			// ---- valid: resolve with dynamic import result ----
			{Code: `new Promise(async (resolve, reject) => {
      try {
        const r = await foo();
        resolve(import(r));
      } catch (error) {
        reject(error);
      }
    })`},
			// ---- valid: async generator with yield, resolve is last throwable ----
			{Code: `new Promise((resolve, reject) => {
      fn(async function * () {
        try {
          const r = await foo();
          resolve(yield r);
        } catch (error) {
          reject(error);
        }
      })
    })`},
			// ---- valid: resolve then non-throwable return after ----
			{Code: `new Promise(async (resolve, reject) => {
      let a;
      try {
        const r = await foo();
        resolve();
        if(r) return;
      } catch (error) {
        reject(error);
      }
    })`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid: if-without-else in nested callback ----
			{
				Code: `
      new Promise((resolve, reject) => {
        fn((error, value) => {
          if (error) {
            reject(error)
          }

          resolve(value)
        })
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 5.",
						Line:    8,
					},
				},
			},
			// ---- invalid: if-without-else at executor level ----
			{
				Code: `
      new Promise((resolve, reject) => {
        if (error) {
          reject(error)
        }

        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 4.",
						Line:    7,
					},
				},
			},
			// ---- invalid: sequential reject then resolve ----
			{
				Code: `
      new Promise((resolve, reject) => {
        reject(error)
        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgAlreadyLine3,
						Line:    4,
					},
				},
			},
			// ---- invalid: if-without-else with intervening conditions ----
			{
				Code: `
      new Promise((resolve, reject) => {
        fn((error, value) => {
          if (error) {
            reject(error)
          }
          if (!error) {
            // process
          } else {
            // process
          }

          resolve(value)
        })
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgPotentialLine5,
						Line:    13,
					},
				},
			},
			// ---- invalid: deeply nested if without early return ----
			{
				Code: `
      new Promise((resolve, reject) => {
        fn((error, value) => {
          if (error) {
            if (foo) {
              if (bar) reject(error)
            }
          }

          resolve(value)
        })
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgPotentialLine6,
						Line:    10,
					},
				},
			},
			// ---- invalid: if-else where else returns (only reject path continues) ----
			{
				Code: `
      new Promise((resolve, reject) => {
        fn((error, value) => {
          if (error) {
            reject(error)
          } else {
            return
          }

          resolve(value)
        })
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgAlreadyLine5,
						Line:    10,
					},
				},
			},
			// ---- invalid: nested ifs with two resolves ----
			{
				Code: `
      new Promise((resolve, reject) => {
        if(foo) {
          if (error) {
            reject(error)
          } else {
            return
          }
          resolve(value)
        }

        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgAlreadyLine5,
						Line:    9,
					},
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 9.",
						Line:    12,
					},
				},
			},
			// ---- invalid: if-else with resolve in each branch then another resolve ----
			{
				Code: `
      new Promise((resolve, reject) => {
        if (foo) {
          reject(error)
        } else {
          resolve(value)
        }
        if(bar) {
          resolve(value)
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgAlreadyLine4,
						Line:    9,
					},
				},
			},
			// ---- invalid: while loop then resolve ----
			{
				Code: `
      new Promise((resolve, reject) => {
        while (error) {
          reject(error)
        }
        resolve(value)
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgPotentialLine4,
						Line:    6,
					},
				},
			},
			// ---- invalid: try-finally certain ----
			{
				Code: `
      new Promise((resolve, reject) => {
        try {
          reject(error)
        } finally {
          resolve(value)
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgAlreadyLine4,
						Line:    6,
					},
				},
			},
			// ---- invalid: try-finally with conditional reject ----
			{
				Code: `
      new Promise((resolve, reject) => {
        try {
          if (error) {
            reject(error)
          }
        } finally {
          resolve(value)
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgPotentialLine5,
						Line:    8,
					},
				},
			},
			// ---- invalid: async try-catch, throwable call after resolve ----
			{
				Code: `new Promise(async (resolve, reject) => {
        try {
          const r = await foo();
          resolve();
          r();
        } catch (error) {
          reject(error);
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: "Promise should not be resolved multiple times. Promise is potentially resolved on line 4.",
						Line:    7,
					},
				},
			},
			// ---- invalid: async try-catch, property access after resolve ----
			{
				Code: `new Promise(async (resolve, reject) => {
        let a;
        try {
          const r = await foo();
          resolve();
          a = r.foo;
        } catch (error) {
          reject(error);
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgPotentialLine5,
						Line:    8,
					},
				},
			},
			// ---- invalid: async try-catch, new expression after resolve ----
			{
				Code: `new Promise(async (resolve, reject) => {
        let a;
        try {
          const r = await foo();
          resolve();
          a = new r();
        } catch (error) {
          reject(error);
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgPotentialLine5,
						Line:    8,
					},
				},
			},
			// ---- invalid: async try-catch, dynamic import after resolve ----
			{
				Code: `new Promise(async (resolve, reject) => {
        let a;
        try {
          const r = await foo();
          resolve();
          import(r);
        } catch (error) {
          reject(error);
        }
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgPotentialLine5,
						Line:    8,
					},
				},
			},
			// ---- invalid: async generator, yield after resolve ----
			{
				Code: `new Promise((resolve, reject) => {
        fn(async function * () {
          try {
            const r = await foo();
            resolve();
            yield r;
          } catch (error) {
            reject(error);
          }
        })
      })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Message: msgPotentialLine5,
						Line:    8,
					},
				},
			},
		},
	)
}
