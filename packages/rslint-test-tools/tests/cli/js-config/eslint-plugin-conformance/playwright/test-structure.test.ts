/**
 * Conformance: eslint-plugin-playwright (test-structure) mounted in rslint via `plugins`
 * must report identically to ESLint v10. playwright rules are AST pattern
 * matches over test/expect/page call shapes (no type info), so rslint
 * reproduces ESLint byte-for-byte. Representative triggers from the upstream
 * test suite (v2.10.4): describe/test/hook organization, titles, skip/focus/slow markers.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'consistent-spacing-between-blocks',
    code: `const someText = 'abc';
test.afterAll(() => {
});
test.describe('someText', () => {
  const something = 'abc';
  // A comment
  test.afterAll(() => {
    // stuff
  });
  test.afterAll(() => {
    // other stuff
  });
});

test.describe('someText', () => {
  const something = 'abc';
  test.afterAll(() => {
    // stuff
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'consistent-spacing-between-blocks',
    code: `test('does something', () => {});
const someVariable = 'value';`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'consistent-spacing-between-blocks',
    code: `test('does something', () => {});
function helperFunction() {
  return true;
}`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'consistent-spacing-between-blocks',
    code: `test('does something', () => {});
// A comment after test
const x = 1;`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'expect-expect',
    code: 'test("should fail", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'expect-expect',
    code: 'test.skip("should fail", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'expect-expect',
    code: `test('should fail', async ({ page }) => {
  await assertCustomCondition(page)
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'expect-expect',
    code: `const custom = test.extend({});
custom("should fail", () => {});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-expects',
    code: `test('should not pass', function () {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-expects',
    code: `test('should not pass', async () => {
  await test.step('part 1', async () => {
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
  })

  await test.step('part 1', async () => {
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
  })
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-expects',
    code: `test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-expects',
    code: `describe('test', () => {
  test('should not pass', () => {
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-nested-describe',
    code: `test.describe.only('foo', function() {
  test.describe('bar', function() {
    test.describe('baz', function() {
      test.describe('qux', function() {
        test.describe('quux', function() {
          test.describe.only('quuz', function() { });
        });
      });
    });
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-nested-describe',
    code: `test.describe.serial.only('foo', function() {
  test.describe('bar', function() {
    test.describe('baz', function() {
      test.describe('qux', function() {
        test.describe('quux', function() {
          test.describe('quuz', function() { });
        });
      });
    });
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-nested-describe',
    code: `test.describe('qux', () => {
  test('should get something', () => {
    expect(getSomething()).toBe('Something');
  });
});`,
    options: [{ max: 0 }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-nested-describe',
    code: `test.describe('foo', () => {
  test.describe('bar', () => {
    test.describe('baz', () => {
      test("test1", () => {
        expect(true).toBe(true);
      });
      test("test2", () => {
        expect(true).toBe(true);
      });
    });
  });
});`,
    options: [{ max: 2 }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-commented-out-tests',
    code: '// test.describe("foo", function () {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-commented-out-tests',
    code: '// describe["skip"]("foo", function () {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-commented-out-tests',
    code: '// describe[\'skip\']("foo", function () {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-commented-out-tests',
    code: '// test("foo", function () {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-in-test',
    code: 'test("foo", () => { if (true) { expect(1).toBe(1); } });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-in-test',
    code: `test.describe("foo", () => {
  test("bar", () => {
    if (someCondition()) {
      expect(1).toBe(1);
    } else {
      expect(2).toBe(2);
    }
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-in-test',
    code: `describe('foo', () => {
  test('bar', () => {
    if ('bar') {}
  })
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-in-test',
    code: `describe('foo', () => {
  test('bar', () => {
    if ('bar') {}
  })
  test('baz', () => {
    if ('qux') {}
    if ('quux') {}
  })
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-hooks',
    code: `test.describe("foo", () => {
  test.beforeEach(() => {}),
  test.beforeEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-hooks',
    code: `describe.skip("foo", () => {
  test.beforeEach(() => {}),
  test.beforeAll(() => {}),
  test.beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-hooks',
    code: `describe.skip("foo", () => {
  test.afterEach(() => {}),
  test.afterEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-hooks',
    code: `describe.skip("foo", () => {
  test.afterAll(() => {}),
  test.afterAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-slow',
    code: `test('should do something', async ({ page }) => {
  test.slow();
  await doSomething();
  test.slow();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-slow',
    code: `test('should do something', async ({ page }) => {
  await test.step('complete first form', async () => {
    test.slow();
    await fillForm();
  });

  await test.step('complete other form', async () => {
    test.slow();
    await page.reload();
    await fillForm();
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-slow',
    code: `test('should do something', async ({ page }) => {
  test.slow();
  test.slow();
  test.slow();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-slow',
    code: `test.slow();

test('should do something', async ({ page }) => {
  test.slow();
  await doSomething();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-focused-test',
    code: 'test.describe.only("skip this describe", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-focused-test',
    code: 'test.describe["only"]("skip this describe", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-focused-test',
    code: 'test["describe"][`only`]("skip this describe", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-focused-test',
    code: 'test.describe.parallel.only("skip this describe", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-hooks',
    code: 'test.beforeAll(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-hooks',
    code: 'test.beforeEach(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-hooks',
    code: 'test.afterAll(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-hooks',
    code: 'test.afterEach(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nested-step',
    code: `test('foo', async () => {
  await test.step("step1", async () => {
    await test.step("nested step1", async () => {
      await expect(true).toBe(true);
    });
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nested-step',
    code: `test('foo', async () => {
  await test.step("step1", async () => {
    await test.step("nested step1", async () => {
      await expect(true).toBe(true);
    });
    await test.step("nested step1", async () => {
      await expect(true).toBe(true);
    });
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nested-step',
    code: `test('foo', async () => {
  await test.step("step1", async () => {
    await test.step("nested step1", async () => {
      await expect(true).toBe(true);
    });
    await test.step.skip("nested step2", async () => {
      await expect(true).toBe(true);
    });
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nested-step',
    code: `import { test as custom } from '@playwright/test';
custom('foo', async () => {
  await custom.step("step1", async () => {
    await custom.step("nested step1", async () => {
      await expect(true).toBe(true);
    });
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-skipped-test',
    code: 'test.skip("skip this test", async ({ page }) => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-skipped-test',
    code: 'test["skip"]("skip this test", async ({ page }) => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-skipped-test',
    code: 'test[`skip`]("skip this test", async ({ page }) => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-skipped-test',
    code: 'test.skip("a test", { tag: ["@fast", "@login"] }, () => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-slowed-test',
    code: 'test.slow("slow this test", async ({ page }) => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-slowed-test',
    code: 'test["slow"]("slow this test", async ({ page }) => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-slowed-test',
    code: 'test[`slow`]("slow this test", async ({ page }) => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-slowed-test',
    code: 'test.slow("a test", { tag: ["@fast", "@login"] }, () => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-standalone-expect',
    code: 'test.describe("a test", () => { expect(1).toBe(1); });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-standalone-expect',
    code: 'test.describe("a test", () => { expect.soft(1).toBe(1); });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-standalone-expect',
    code: 'test.describe("a test", () => { expect.poll(() => 1).toBe(1); });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-standalone-expect',
    code: "(() => {})('testing', () => expect(true).toBe(false))",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-in-order',
    code: `const withDatabase = () => {
  test.afterAll(() => {
    removeMyDatabase();
  });
  test.beforeAll(() => {
    createMyDatabase();
  });
};`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-in-order',
    code: `test.afterAll(() => {
  removeMyDatabase();
});
test.beforeAll(() => {
  createMyDatabase();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-in-order',
    code: `test.afterAll(() => {});
test.beforeAll(() => {});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-in-order',
    code: `test.afterEach(() => {});
test.beforeEach(() => {});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-on-top',
    code: `test.describe('foo', () => {
  test.beforeEach(() => {});
  test('bar', () => {
    someFn();
  });

  test.beforeAll(() => {});
  test('bar', () => {
    someFn();
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-on-top',
    code: `test.describe('foo', () => {
  test.beforeEach(() => {});
  test.only('bar', () => {
    someFn();
  });

  test.beforeAll(() => {});
  test.only('bar', () => {
    someFn();
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-on-top',
    code: `import { test as custom } from '@playwright/test';
custom.describe('foo', () => {
  custom.beforeEach(() => {});
  custom('bar', () => {
    someFn();
  });

  custom.beforeAll(() => {});
  custom('bar', () => {
    someFn();
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-on-top',
    code: `test.describe('foo', () => {
  test('bar', () => {});
  test.afterEach(() => {});
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-lowercase-title',
    code: "test('Foo',  () => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-lowercase-title',
    code: 'test(`Foo bar`,  () => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-lowercase-title',
    code: "test.skip('Foo Bar',  () => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-lowercase-title',
    code: 'test.skip(`Foo`,  () => {})',
  },
  { pkg: 'eslint-plugin-playwright', rule: 'require-hook', code: 'setup();' },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-hook',
    code: `test.describe('some tests', () => {
  setup();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-hook',
    code: `test.describe('some tests', { tag: '@slow' }, () => {
  setup();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-hook',
    code: `let { setup } = require('./test-utils');

test.describe('some tests', () => {
  setup();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-tags',
    code: "test('my test', async ({ page }) => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-tags',
    code: "test('my test', { timeout: 5000 }, async ({ page }) => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-tags',
    code: "test.skip('my test', async ({ page }) => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-tags',
    code: "test.fixme('my test', async ({ page }) => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-top-level-describe',
    code: 'test.beforeAll(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-top-level-describe',
    code: 'test.beforeEach(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-top-level-describe',
    code: 'test.afterAll(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-top-level-describe',
    code: 'test.afterEach(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-describe-callback',
    code: 'test.describe("foo")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-describe-callback',
    code: 'test.describe("foo", { tag: ["@slow"] });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-describe-callback',
    code: 'test.describe("foo", "foo2")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-describe-callback',
    code: 'test.describe("foo", foo2)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-test-tags',
    code: "test('my test', { tag: 'e2e' }, async ({ page }) => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-test-tags',
    code: 'test(`my test`, { tag: `e2e` }, async ({ page }) => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-test-tags',
    code: 'test(`my test`, { tag: `@skip` }, async ({ page }) => {})',
    options: [{ disallowedTags: ['@skip', '@todo'] }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-test-tags',
    code: 'test(`my test`, { tag: `@e2e` }, async ({ page }) => {})',
    options: [{ allowedTags: ['@regression', '@smoke'] }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-title',
    code: 'test("the correct way to properly handle all things", () => {});',
    options: [{ disallowedWords: ['correct', 'properly', 'all'] }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-title',
    code: 'test.describe("the correct way to do things", function () {})',
    options: [{ disallowedWords: ['correct'] }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-title',
    code: 'test("has ALL the things", () => {})',
    options: [{ disallowedWords: ['all'] }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-title',
    code: "test.describe('Very Descriptive Title Goes Here', function () {})",
    options: [{ disallowedWords: ['descriptive'] }],
  },
];

const CLEAN: DiffCase[] = [
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'consistent-spacing-between-blocks',
    code: `const a = 'value';
doSomething('does something', () => {});
const b = 'value';
testing('does something', () => {});
testing.beforeEach(() => {});
helloWorld();
class C{}
function helloWorld() {}
let d = 'value';
if (e) {
  doSomething('does something', () => {});
}`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'consistent-spacing-between-blocks',
    code: `test('does something', () => {});

const someVariable = 'value';`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'consistent-spacing-between-blocks',
    code: `test('does something', () => {});

function helperFunction() {
  return true;
}`,
  },
  { pkg: 'eslint-plugin-playwright', rule: 'expect-expect', code: 'foo();' },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'expect-expect',
    code: '["bar"]();',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'expect-expect',
    code: 'testing("will test something eventually", () => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-expects',
    code: "test('should pass')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-expects',
    code: "test('should pass', () => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-expects',
    code: "test.skip('should pass', () => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-nested-describe',
    code: 'test.describe("describe tests", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-nested-describe',
    code: 'test.describe.only("describe focus tests", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'max-nested-describe',
    code: 'test.describe.serial.only("describe serial focus tests", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-commented-out-tests',
    code: '// foo("bar", function () {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-commented-out-tests',
    code: 'test.describe("foo", function () {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-commented-out-tests',
    code: 'test("foo", function () {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-in-test',
    code: 'test("foo", () => { expect(1).toBe(1); });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-in-test',
    code: 'test.fixme("some broken test", () => { expect(1).toBe(1); });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-in-test',
    code: 'test.describe("foo", () => { if(true) { test("bar", () => { expect(true).toBe(true); }); } });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-hooks',
    code: `test.describe("foo", () => {
  test.beforeEach(() => {})
  test("bar", () => {
    someFn();
  })
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-hooks',
    code: `test.beforeEach(() => {})
test("bar", () => {
  someFn();
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-hooks',
    code: `test.describe("foo", () => {
  test.beforeAll(() => {}),
  test.beforeEach(() => {})
  test.afterEach(() => {})
  test.afterAll(() => {})

  test("bar", () => {
    someFn();
  })
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-slow',
    code: `test('should do something', async ({ page }) => {
  test.slow();
  await doSomething();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-slow',
    code: `test('test one', async ({ page }) => {
  test.slow();
  await doSomething();
});

test('test two', async ({ page }) => {
  test.slow();
  await doSomething();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-duplicate-slow',
    code: `test('should do something', async ({ browserName }) => {
  test.slow(browserName === 'firefox', 'Slow on Firefox');
  await doSomething();
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-focused-test',
    code: 'test.describe("describe tests", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-focused-test',
    code: 'test.describe.skip("describe tests", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-focused-test',
    code: 'test("one", async ({code: page }) => {});',
  },
  { pkg: 'eslint-plugin-playwright', rule: 'no-hooks', code: 'test("foo")' },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-hooks',
    code: 'test.describe("foo", () => { test("bar") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-hooks',
    code: 'test("foo", () => { expect(subject.beforeEach()).toBe(true) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nested-step',
    code: 'await test.step("step1", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nested-step',
    code: 'await test.step("step1", async () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nested-step',
    code: 'await test.step.skip("step1", async () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-skipped-test',
    code: 'test("a test", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-skipped-test',
    code: 'test("a test", { tag: "@fast" }, () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-skipped-test',
    code: 'test("a test", { tag: ["@fast", "@report"] }, () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-slowed-test',
    code: 'test("a test", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-slowed-test',
    code: 'test("a test", { tag: "@fast" }, () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-slowed-test',
    code: 'test("a test", { tag: ["@fast", "@report"] }, () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-standalone-expect',
    code: 'expect.any(String)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-standalone-expect',
    code: 'expect.extend({})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-standalone-expect',
    code: 'test.describe("a test", () => { test("an it", () => {expect(1).toBe(1); }); });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-in-order',
    code: 'test.beforeAll(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-in-order',
    code: 'test.beforeEach(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-in-order',
    code: 'test.afterEach(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-on-top',
    code: `test.describe('foo', () => {
  test.beforeEach(() => {});
  someSetupFn();
  test.afterEach(() => {});

  test('bar', () => {
    someFn();
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-on-top',
    code: `test.describe('foo', () => {
  someSetupFn();
  test.beforeEach(() => {});
  test.afterEach(() => {});

  test('bar', () => {
    someFn();
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-hooks-on-top',
    code: `it.describe('foo', () => {
  it.beforeEach(() => {});
  someSetupFn();
  it.afterEach(() => {});

  it('bar', () => {
    someFn();
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-lowercase-title',
    code: 'randomFunction()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-lowercase-title',
    code: 'foo.bar()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-lowercase-title',
    code: "test('foo Bar', () => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-hook',
    code: 'test.use({ locale: "en-US" })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-hook',
    code: 'test("some test", async ({ page }) => { })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-hook',
    code: 'test("some test", { tag: "@slow" }, async ({ page }) => { })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-tags',
    code: "test('my test', { tag: '@e2e' }, async ({ page }) => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-tags',
    code: 'test(`@e2e my test`, async ({ page }) => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-tags',
    code: "test('my test', { tag: ['@e2e', '@login'] }, async ({ page }) => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-top-level-describe',
    code: 'foo()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-top-level-describe',
    code: 'test.info()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-top-level-describe',
    code: 'test.use({ locale: "en-US" })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-describe-callback',
    code: 'describe(() => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-describe-callback',
    code: 'describe.configure({ timeout: 600_000 })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-describe-callback',
    code: 'describe("foo", function() {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-test-tags',
    code: "test('my test', { tag: '@e2e' }, async ({ page }) => {})",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-test-tags',
    code: 'test(`my test`, { tag: `@e2e` }, async ({ page }) => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-test-tags',
    code: 'test(`my test`, { tag: [`@e2e`, `@login`] }, async ({ page }) => {})',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-title',
    code: 'test.describe("the correct way to properly handle all the things", () => {});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-title',
    code: 'test.describe.configure({ mode: "parallel" })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-title',
    code: 'test.describe(() => {})',
  },
];

runConformanceSuite('eslint-plugin-playwright', CASES, CLEAN);
