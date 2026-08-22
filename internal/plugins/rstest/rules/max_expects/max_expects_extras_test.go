// TestMaxExpectsExtras locks in branches and edge shapes that the upstream
// max-expects suites do not exercise on rstest. Each case names the state
// branch, source shape, or real-user scenario it protects so refactors cannot
// silently regress rstest-specific behavior.
package max_expects_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/max_expects"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxExpectsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&max_expects.MaxExpectsRule,
		[]rule_tester.ValidTestCase{
			// ---- Real-user: Jest #1193 ----
			{
				Code: `test('static expect APIs do not count', () => {
  expect.assertions(5);
  expect(a).toBe(1);
  expect(b).toBe(2);
  expect(c).toBe(3);
  expect(d).toBe(4);
  expect(e).toBe(5);
});`,
			},
			{
				Code: `test('asymmetric matchers do not count', () => {
  expect(value).toEqual(expect.objectContaining({ ok: true }));
  expect.any(String);
  expect.anything();
  expect.not.arrayContaining([]);
});`,
				Options: max1Option,
			},

			// ---- Real-user: Jest #1205 ----
			{
				Code: `function helper() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
}

test('isolates helper from following test', () => {
  expect(true).toBeDefined();
});`,
				Options: max1Option,
			},
			{
				Code: `test('isolates following helper from test', () => {
  expect(true).toBeDefined();
});

function helper() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
}`,
				Options: max1Option,
			},

			// ---- Real-user: Jest #1516 ----
			{
				Code: `describe('wrapped callbacks stay per-test', () => {
  test('a', fakeAsync(() => {
    expect(true).toBe(true);
    expect(true).toBe(true);
    expect(true).toBe(true);
  }));

  test('b', fakeAsync(() => {
    expect(true).toBe(true);
    expect(true).toBe(true);
    expect(true).toBe(true);
  }));
});`,
				Options: max5Option,
			},

			// ---- Real-user: Vitest #891 ----
			{
				Code: `import { it as base } from '@rstest/core';
const test = base.extend({ value: 1 });

test('only expect destructured still counts as test callback', ({ expect }) => {
  expect(true).toBe(true);
});`,
				Options: max1Option,
			},
			{
				Code: `import { it as base } from '@rstest/core';
const test = base.extend({ value: 1 });

test('fixture plus expect destructured still counts as test callback', ({ value, expect }) => {
  expect(value).toBe(1);
});`,
				Options: max1Option,
			},

			// ---- Locks in upstream create() arm: registration arguments are outside the test body ----
			{
				Code: `test(expect(1).toBe(1), () => {
  expect(2).toBe(2);
});`,
				Options: max1Option,
			},
			{
				Code: `test('case', { timeout: expect(1).toBe(1) }, () => {
  expect(2).toBe(2);
});`,
				Options: max1Option,
			},
			{
				Code:    `test('case', () => {}, expect(1).toBe(1));`,
				Options: max1Option,
			},

			// ---- Locks in upstream create() arm: broken/static chains do not count ----
			{
				Code: `test('broken chains do not count', () => {
  expect(value);
  expect(value).toBe;
  expect.soft(value);
  expect(value).not;
  expect(value).toBe(1);
});`,
				Options: max1Option,
			},

			// ---- Dimension 4: chai property and multi-matcher chain count once ----
			{
				Code: `test('chai property matcher counts once', () => {
  expect(value).to.be.ok;
});`,
				Options: max1Option,
			},
			{
				Code: `test('chai chain counts once', () => {
  expect(value).to.be.a('string').and.not.be.empty;
});`,
				Options: max1Option,
			},

			// ---- Dimension 4: expect sources ----
			{
				Code: `import { expect, test } from '@rstest/core';
test('named import', () => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `import { expect as check, test } from '@rstest/core';
test('alias import', () => {
  check(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `const { expect, test } = require('@rstest/core');
test('require destructuring', () => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `import * as rstest from '@rstest/core';
rstest.test('namespace import', () => {
  rstest.expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `const rstest = require('@rstest/core');
rstest.test('whole-module require', () => {
  rstest.expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `import.meta.rstest.test('import.meta direct', () => {
  import.meta.rstest.expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `const { test, expect } = import.meta.rstest;
test('import.meta destructuring', () => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `const api = import.meta.rstest;
api.test('import.meta alias', () => {
  api.expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test('context receiver', context => {
  context.expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test('context destructuring', ({ expect }) => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test('context alias destructuring', ({ expect: check }) => {
  check(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test('browser mode element', () => {
  expect.element(locator).toBeVisible();
});`,
				Options: max1Option,
			},
			{
				Code: `import { test, expect } from '@rstest/playwright';
test('playwright profile', () => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},

			// ---- Dimension 4: registration and callback shapes ----
			{
				Code: `test('third arg timeout', () => {
  expect(1).toBe(1);
}, 1000);`,
				Options: max1Option,
			},
			{
				Code: `test('options overload', { timeout: 1 }, () => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test.each([1])('each callback', value => {
  expect(value).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test.for([1])('for callback with context', (value, context) => {
  context.expect(value).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `const eachCase = test.each([1]);
eachCase('factory alias', value => {
  expect(value).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `const forCase = test.for([1]);
forCase('factory alias', (value, context) => {
  context.expect(value).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `const testWithFixture = test.extend({ value: 1 });
testWithFixture('extended test', ({ value, expect }) => {
  expect(value).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test('named callback', callback);
function callback() {
  expect(1).toBe(1);
}`,
				Options: max1Option,
			},

			// ---- Dimension 4: function-position inheritance ----
			{
				Code: `test('call argument callback inherits outer frame', () => {
  run(() => {
    expect(1).toBe(1);
  });
});`,
				Options: max1Option,
			},
			{
				Code: `test('function declaration inside test inherits outer frame', () => {
  function helper() {
    expect(1).toBe(1);
  }
  helper();
});`,
				Options: max1Option,
			},
			// ---- Dimension 4: negative sources ----
			{Code: `import { expect } from 'vitest'; expect(1).toBe(1);`, Options: max1Option},
			{Code: `import { expect } from '@jest/globals'; expect(1).toBe(1);`, Options: max1Option},
			{Code: `import { test, expect } from '@playwright/test'; test('case', () => { expect(1).toBe(1); });`, Options: max1Option},
			{Code: `const expect = () => ({ toBe() {} }); test('case', () => { expect(1).toBe(1); });`, Options: max1Option},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Deliberate divergence: a detached helper gets its own count
			// and restores the enclosing one, where upstream resets the count to
			// zero on both entry and exit and loses the enclosing tally ----
			{
				Code: `test('restores outer count after detached helper', () => {
  expect(1).toBe(1);
  const helper = () => {
    expect(2).toBe(2);
    expect(3).toBe(3);
  };
  expect(4).toBe(4);
  expect(5).toBe(5);
});`,
				Options: max2Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(3, 2, 8, 3, 20),
				},
			},

			// ---- Real-user: Jest #1516 ----
			{
				Code: `test('a', fakeAsync(() => {
  expect(true).toBe(true);
  expect(true).toBe(true);
  expect(true).toBe(true);
}));
test('b', fakeAsync(() => {
  expect(true).toBe(true);
  expect(true).toBe(true);
  expect(true).toBe(true);
}));`,
				Options: max2Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(3, 2, 4, 3),
					exceededMaxError(3, 2, 9, 3),
				},
			},

			// ---- Dimension 4: detached function frames still enforce their own max ----
			{
				Code: `test('detached methods are isolated', () => {
  const helper = {
    check() {
      expect(1).toBe(1);
      expect(2).toBe(2);
    },
  };
  expect(3).toBe(3);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 5, 7),
				},
			},
			{
				Code: `test('constructor and accessors are isolated', () => {
  class Helper {
    constructor() {
      expect(1).toBe(1);
      expect(2).toBe(2);
    }

    get value() {
      expect(3).toBe(3);
      expect(4).toBe(4);
      return 1;
    }

    set value(next) {
      expect(next).toBe(1);
      expect(next).toBe(2);
    }
  }

  expect(5).toBe(5);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 5, 7, 24),
					exceededMaxError(2, 1, 10, 7, 24),
					exceededMaxError(2, 1, 16, 7, 27),
				},
			},

			// ---- Real-user: Vitest #891 ----
			{
				Code: `import { it as base } from '@rstest/core';
const test = base.extend({ value: 1 });

test('expect-only destructuring still counts', ({ expect }) => {
  expect(1).toBe(1);
  expect(2).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 6, 3),
				},
			},

			// ---- Dimension 4: true assertion factories ----
			{
				Code: `test('soft counts', () => {
  expect.soft(1).toBe(1);
  expect.soft(2).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 3),
				},
			},
			{
				Code: `test('poll counts', async () => {
  await expect.poll(read).toBe(1);
  await expect.poll(read).toBe(2);
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 9),
				},
			},
			{
				Code: `test('element counts', async () => {
  await expect.element(locator).toBeVisible();
  await expect.element(locator).toBeHidden();
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 9),
				},
			},

			// ---- Dimension 4: callback inheritance ----
			{
				Code: `test('call argument callback inherits current frame', () => {
  expect(1).toBe(1);
  emitter.on('event', value => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  });
});`,
				Options: max2Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(3, 2, 5, 5),
				},
			},
			{
				Code: `test('non-callback function declaration inherits current frame', () => {
  expect(1).toBe(1);
  function helper() {
    expect(2).toBe(2);
    expect(3).toBe(3);
  }
  helper();
});`,
				Options: max2Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(3, 2, 5, 5),
				},
			},

			// ---- Locks in registration fallback activation timing ----
			{
				Code: `test(expect(1).toBe(1), fakeAsync(() => {
  expect(2).toBe(2);
  expect(3).toBe(3);
}));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 3),
				},
			},
			{
				Code: `test('case', { timeout: expect(1).toBe(1) }, fakeAsync(() => {
  expect(2).toBe(2);
  expect(3).toBe(3);
}));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 3),
				},
			},
			{
				Code: `test(makeTitle(() => {
  expect('title').toBe('title');
}), () => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test('case', {
  timeout: makeTimeout(() => {
    expect('options').toBe('options');
  }),
}, () => {
  expect(1).toBe(1);
});`,
				Options: max1Option,
			},
			{
				Code: `test('wrapper second arg callback-first overload', fakeAsync(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}), 1000);`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 3),
				},
			},
			{
				Code: `test('wrapper third arg options overload', { timeout: 1000 }, fakeAsync(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 3),
				},
			},
			{
				Code: `const options = { timeout: 1000 };
test('wrapper third arg with non-literal options', options, fakeAsync(() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}));`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3),
				},
			},
			{
				Code: `type TestCallback = () => void;
test('callback behind as expression', (() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}) as TestCallback);`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3),
				},
			},
			{
				Code: `type TestCallback = () => void;
test('callback behind satisfies expression', (() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
}) satisfies TestCallback);`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 3),
				},
			},
			{
				Code: `test('callback behind non-null expression', (() => {
  expect(1).toBe(1);
  expect(2).toBe(2);
})!);`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 3),
				},
			},
		},
	)
}
