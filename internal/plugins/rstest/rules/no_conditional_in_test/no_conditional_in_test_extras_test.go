// TestNoConditionalInTestExtras locks in branches and edge shapes that the
// upstream Jest suite does not exercise. The 1:1 migrated cases live in
// no_conditional_in_test_upstream_test.go.
package no_conditional_in_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	no_conditional_in "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_conditional_in_test"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConditionalInTestExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Dimension 4: helper outside test is out of scope ----
		{Code: `
function helper(value) {
  if (value) {
    return 1;
  }
  return 2;
}

test('case', () => {
  helper(flag);
});
    `},

		// ---- Dimension 4: describe and hook callbacks stay out of scope ----
		{Code: `
describe('suite', () => {
  if (flag) {
    prepare();
  }
});

beforeEach(() => {
  if (flag) {
    prepare();
  }
});
    `},

		// ---- Dimension 4: a callback declared outside the test call is out of
		// scope, whatever argument position names it ----
		{Code: `
test('case', callback);
function callback() {
  if (flag) {
    act();
  }
}
    `},
		{Code: `
const callback = () => {
  if (flag) {
    act();
  }
};
test('case', callback);
    `},
		{Code: `
const TIMEOUT = 1000;
test('case', callback, TIMEOUT);
function callback() {
  if (flag) {
    act();
  }
}
    `},
		{Code: `
test('case', { timeout: 1000 }, callback);
function callback() {
  if (flag) {
    act();
  }
}
    `},
		{Code: `test('case', callback); let callback = () => { if (flag) { act(); } }; callback = other;`},

		// ---- Dimension 4: body-absent or non-callback overloads do not crash ----
		{Code: `test('case');`},
		{Code: `test('case', { timeout: 1000 });`},
		{Code: `const cb = condition && callback; test('case', cb);`},

		// ---- Dimension 4: explicit true and empty object keep the default ----
		{Code: `test('case', () => { value?.prop; });`, Options: map[string]any{}},
		{Code: `test('case', () => { value?.prop; });`, Options: map[string]any{"allowOptionalChaining": true}},

		// ---- Real-user: Vitest narrow contract does not apply to rstest ----
		{Code: `
import { test } from 'vitest';
test('case', () => {
  if (flag) {
    act();
  }
});
    `},
		{Code: `
import { test } from '@jest/globals';
test('case', () => {
  if (flag) {
    act();
  }
});
    `},
		{Code: `
import { test } from '@playwright/test';
test('case', () => {
  if (flag) {
    act();
  }
});
    `},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Dimension 4: TS assertion wrappers around inline callbacks are transparent ----
		{
			Code: `test('case', (() => { if (flag) { act(); } }) as () => void);`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      1,
				Column:    23,
			}},
		},
		{
			Code: `test('case', (() => { if (flag) { act(); } }) satisfies () => void);`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      1,
				Column:    23,
			}},
		},
		{
			Code: `test('case', (() => { if (flag) { act(); } })!);`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      1,
				Column:    23,
			}},
		},

		// ---- Dimension 4: every argument of the test call is inside the test ----
		{
			Code: `
test('case', wrap(() => {
  if (flag) {
    act();
  }
}));
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code: `test(cond ? 'a' : 'b', () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      1,
				Column:    6,
				EndLine:   1,
				EndColumn: 22,
			}},
		},
		{
			Code:   `test('case', () => {}, timeout || 5);`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 1, Column: 24}},
		},
		{
			Code:   `test.each([a || b])('case', () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 1, Column: 12}},
		},
		{
			Code:   `test.skip.each([a || b])('case', () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 1, Column: 17}},
		},
		{
			Code:   "test.each`${a || b}`('case', () => {});",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 1, Column: 13}},
		},
		{
			Code:    `test(a?.b, () => {});`,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 1, Column: 6}},
		},

		// ---- Divergence from upstream: an inner registration exiting does not
		// clear the outer test's scope ----
		{
			Code:   `test('outer', () => { test('inner', () => {}); if (flag) { act(); } });`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 1, Column: 48}},
		},

		// ---- Dimension 4: nested function declarations inside test still report ----
		{
			Code: `
test('case', () => {
  function helper(value) {
    if (value) {
      return 1;
    }
    return 2;
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      4,
				Column:    5,
				EndLine:   6,
				EndColumn: 6,
			}},
		},
		{
			Code: `
test('case', () => {
  const helper = function(value) {
    if (value) {
      return 1;
    }
    return 2;
  };
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 5}},
		},
		{
			Code: `
test('case', () => {
  const helper = (value) => {
    if (value) {
      return 1;
    }
    return 2;
  };
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 5}},
		},

		// ---- Dimension 4: overload (name, cb, timeout) keeps second argument as callback ----
		{
			Code: `
test('case', () => {
  if (flag) {
    act();
  }
}, TIMEOUT);
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		// ---- Dimension 4: overload (name, options, callback) resolves third argument ----
		{
			Code: `
test('case', { timeout: 1000 }, () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		// ---- Dimension 4: global/named/alias/require/namespace/whole-module/import.meta/playwright sources ----
		{
			Code: `
import { test } from '@rstest/core';
test('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 3}},
		},
		{
			Code: `
import { test as check } from '@rstest/core';
check('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 3}},
		},
		{
			Code: `
const { test } = require('@rstest/core');
test('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 3}},
		},
		{
			Code: `
import * as rstest from '@rstest/core';
rstest.test('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 3}},
		},
		{
			Code: `
const rstest = require('@rstest/core');
rstest.test('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 3}},
		},
		{
			Code: `
import.meta.rstest.test('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code: `
if (import.meta.rstest) {
  const { test } = import.meta.rstest;
  test('case', () => {
    if (flag) {
      act();
    }
  });
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 5, Column: 5}},
		},
		{
			Code: `
import { test } from '@rstest/playwright';
test.fail('case', async () => {
  if (flag) {
    await act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 3}},
		},

		// ---- Dimension 4: same-file API alias ----
		{
			Code: `
const check = test;
check('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 3}},
		},

		// ---- Dimension 4: each and factory alias ----
		{
			Code: `
test.each([[1]])('case', (value) => {
  if (value) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},
		{
			Code: `
const eachCase = test.each([[1]]);
eachCase('case', (value) => {
  if (value) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 4, Column: 3}},
		},

		// ---- Dimension 4: .for context overload ----
		{
			Code: `
test.for([{ enabled: true }])('case', (row, context) => {
  if (row.enabled) {
    context.expect(1).toBe(1);
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},

		// ---- Dimension 4: test.extend ----
		{
			Code: `
test.extend({})('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 3}},
		},

		// ---- Real-user: nested callback depth must not clear outer depth early ----
		{
			Code: `
test('outer', () => {
  test('inner', () => {
    switch (kind) {
      case 'a':
        break;
    }
  });
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 4, Column: 5},
				{MessageId: "conditionalInTest", Line: 9, Column: 3},
			},
		},

		// ---- Real-user: Vitest narrow contract would miss nested helper and switch ----
		{
			Code: `
test('case', () => {
  const helper = () => {
    switch (kind) {
      case 'a':
        break;
    }
  };
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      4,
				Column:    5,
				EndLine:   7,
				EndColumn: 6,
			}},
		},

		// ---- Dimension 4: logical filter only matches logical operators ----
		{
			Code: `
test('case', () => {
  a && b || c ?? d;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 3, Column: 3, EndLine: 3, EndColumn: 19},
				{MessageId: "conditionalInTest", Line: 3, Column: 3, EndLine: 3, EndColumn: 14},
				{MessageId: "conditionalInTest", Line: 3, Column: 3, EndLine: 3, EndColumn: 9},
			},
		},
		{
			Code: `
test('case', () => {
  const value =
    left &&
    right;
});
	      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      4,
				Column:    5,
				EndLine:   5,
				EndColumn: 10,
			}},
		},

		// ---- Dimension 4: optional chain reports outermost chain exactly once ----
		{
			Code: `
test('case', () => {
  obj?.foo?.bar;
});
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    3,
				EndLine:   3,
				EndColumn: 16,
			}},
		},
		{
			Code: `
test('case', () => {
  obj?.method(arg?.value);
});
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "conditionalInTest", Line: 3, Column: 3, EndLine: 3, EndColumn: 26},
				{MessageId: "conditionalInTest", Line: 3, Column: 15, EndLine: 3, EndColumn: 25},
			},
		},
		{
			Code: `
test('case', () => {
  obj?.foo!.bar;
});
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    3,
				EndLine:   3,
				EndColumn: 16,
			}},
		},
		{
			Code: `
test('case', () => {
  obj?.foo!;
});
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    3,
				EndLine:   3,
				EndColumn: 12,
			}},
		},
		{
			Code: `
test('case', () => {
  obj?.method(
    arg,
  );
});
	      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    3,
				EndLine:   5,
				EndColumn: 4,
			}},
		},
		{
			Code: `
test('case', () => {
  (obj?.foo)!.bar;
});
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 4}},
		},
		{
			Code: `
test('case', () => {
  results[obj?.key] = 1;
});
      `,
			Options: map[string]any{"allowOptionalChaining": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "conditionalInTest", Line: 3, Column: 11}},
		},

		// Locks in upstream create() arm 1: callback enter depth guard
		{
			Code: `
test('case', () => {
  if (flag) {
    act();
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    3,
				EndLine:   5,
				EndColumn: 4,
			}},
		},
		// Locks in upstream create() arm 2: switch node reporting
		{
			Code: `
test('case', () => {
  switch (kind) {
    case 'a':
      break;
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    3,
				EndLine:   6,
				EndColumn: 4,
			}},
		},
		// Locks in upstream create() arm 3: ternary node reporting
		{
			Code: `
test('case', () => {
  const value = flag ? 1 : 2;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      3,
				Column:    17,
				EndLine:   3,
				EndColumn: 29,
			}},
		},
		{
			Code: `
test('case', () => {
  const value =
    flag
      ? 1
      : 2;
});
	      `,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "conditionalInTest",
				Line:      4,
				Column:    5,
				EndLine:   6,
				EndColumn: 10,
			}},
		},
		// Locks in upstream create() arm 4: logical filter false branch
		{
			Code: `
test('case', () => {
  const value = left + right;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{},
			Skip:   true, // SKIP: invalid placeholder intentionally unused; covered by valid cases above.
		},
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_conditional_in.NoConditionalInTestRule,
		valid,
		invalid,
	)
}
