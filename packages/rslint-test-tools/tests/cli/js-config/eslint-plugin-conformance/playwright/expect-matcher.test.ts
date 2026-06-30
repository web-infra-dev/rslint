/**
 * Conformance: eslint-plugin-playwright (expect-matcher) mounted in rslint via `plugins`
 * must report identically to ESLint v10. playwright rules are AST pattern
 * matches over test/expect/page call shapes (no type info), so rslint
 * reproduces ESLint byte-for-byte. Representative triggers from the upstream
 * test suite (v2.10.4): matcher preferences, await-ness, restricted/valid expect usage.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'missing-playwright-await',
    code: "test('test', async () => { expect(page).toBeChecked() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'missing-playwright-await',
    code: "test('test', async () => { expect(page).not.toBeEnabled() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'missing-playwright-await',
    code: `test('test', async () => { expect(async () => {
          expect(await page.evaluate(() => window.foo)).toBe('bar')
        }).toPass() })`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'missing-playwright-await',
    code: "test('test', async () => { expect(page).toBeCustomThing(false) })",
    options: [{ customMatchers: ['toBeCustomThing'] }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-expect',
    code: `
        test('foo', () => {
          something && expect(something).toHaveBeenCalled();
        })
      `,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-expect',
    code: `
        test('foo', async () => {
          await test.step('bar', async () => {
            something && expect(something).toHaveBeenCalled();
          })
        })
      `,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-expect',
    code: `
        test('foo', () => {
          a || b && expect(something).toHaveBeenCalled();
        })
      `,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-expect',
    code: `
        test('foo', () => {
          (a || b) && expect(something).toHaveBeenCalled();
        })
      `,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-matchers',
    code: 'expect(a).toBe(b)',
    options: [{ toBe: null }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-matchers',
    code: 'expect.soft(a).toBe(b)',
    options: [{ toBe: null }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-matchers',
    code: 'expect["poll"](() => a)["toBe"](b)',
    options: [{ toBe: null }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-matchers',
    code: 'expect(a).not.toBe()',
    options: [{ not: null }],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-not',
    code: 'expect(locator).not.toBeVisible()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-not',
    code: 'expect(locator).not.toBeVisible({ visible: true })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-not',
    code: 'expect(locator).toBeVisible({ visible: true })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-not',
    code: 'expect(locator).toBeVisible({ visible: false })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-comparison-matcher',
    code: 'expect(value > 1).toBe(true);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-comparison-matcher',
    code: 'expect(value < 1).toBe(true);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-comparison-matcher',
    code: 'expect(value >= 1).toBe(true);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-comparison-matcher',
    code: 'expect(value <= 1).toBe(true);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-equality-matcher',
    code: 'expect(value === 1).toBe(true);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-equality-matcher',
    code: 'expect(value !== 1).toBe(true);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-equality-matcher',
    code: 'expect(value === 1).toBe(false);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-strict-equal',
    code: 'expect(something).toEqual(somethingElse);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-strict-equal',
    code: 'expect(something)["toEqual"](somethingElse);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-strict-equal',
    code: `const custom = test.extend({});
custom("foo", () => {
  expect(something).toEqual(somethingElse);
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-strict-equal',
    code: `import { expect as assuming } from '@playwright/test';
assuming(something).toEqual(somethingElse);`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-be',
    code: 'expect(value).toEqual("my string");',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-be',
    code: 'expect(value).toStrictEqual("my string");',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-be',
    code: 'expect(value).toStrictEqual(1);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-be',
    code: 'expect(value).toStrictEqual(-1);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-contain',
    code: 'expect(a.includes(b)).toEqual(true);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-contain',
    code: 'expect(a.includes(b,),).toEqual(true,);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-contain',
    code: "expect(a['includes'](b)).toEqual(true);",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-contain',
    code: "expect(a['includes'](b)).toEqual(false);",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-count',
    code: 'expect(await files.count()).toBe(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-count',
    code: 'expect(await files.count()).not.toBe(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-count',
    code: 'expect.soft(await files["count"]()).not.toBe(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-count',
    code: 'expect(await files["count"]()).not["toBe"](1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-length',
    code: 'expect(files.length).toBe(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-length',
    code: 'expect(files.length).not.toBe(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-length',
    code: 'expect.soft(files["length"]).not.toBe(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-length',
    code: 'expect(files["length"]).not["toBe"](1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-web-first-assertions',
    code: 'test(\'test\', async () => { expect(await page.locator(".tweet").isVisible()).toBe(true) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-web-first-assertions',
    code: 'test(\'test\', async () => { expect(page.locator(".tweet").isVisible()).toBe(true) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-web-first-assertions',
    code: 'test(\'test\', async () => { expect(await page.locator(".tweet").isVisible()).toBe(false) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-web-first-assertions',
    code: `test('test', async () => { 
        const unrelatedAssignment = 'unrelated'
        const isTweetVisible = await page.locator(".tweet").isVisible()
        expect(isTweetVisible).toBe(true)
       })`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-soft-assertions',
    code: 'expect(page).toHaveTitle("baz")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-soft-assertions',
    code: 'expect(page.locator("foo")).toHaveText("bar")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-soft-assertions',
    code: 'await expect(page.locator("foo")).toHaveText("bar")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-soft-assertions',
    code: `const custom = test.extend({});
custom("foo", () => {
  expect(page).toHaveTitle("baz");
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-pass-timeout',
    code: 'await expect(async () => { await foo() }).toPass();',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-pass-timeout',
    code: 'await expect(async () => { await foo() }).toPass({});',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-pass-timeout',
    code: 'await expect(async () => { await foo() }).toPass({ intervals: [1000] });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-pass-timeout',
    code: 'await expect.soft(async () => { await foo() }).toPass();',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-throw-message',
    code: "expect(() => { throw new Error('a'); }).toThrow();",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-throw-message',
    code: "expect(() => { throw new Error('a'); }).toThrowError();",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-throw-message',
    code: `test('empty rejects.toThrow', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow();
  await expect(throwErrorAsync()).rejects.toThrowError();
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-throw-message',
    code: `const custom = test.extend({});
custom("test", () => { expect(() => { throw new Error('a'); }).toThrow(); });`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect',
    code: 'expect(foo)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect',
    code: 'softExpect(foo)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect',
    code: 'expect(foo).not',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect',
    code: 'expect.soft(foo)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect-in-promise',
    code: `const myFn = () => {
  Promise.resolve().then(() => {
    expect(true).toBe(false);
  });
};

test('foo', () => {
  somePromise.then(() => {
    expect(someThing).toEqual(true);
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect-in-promise',
    code: `test('foo', () => {
  somePromise.then(() => {
    expect(someThing).toEqual(true);
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect-in-promise',
    code: `test('foo', () => {
  somePromise.finally(() => {
    expect(someThing).toEqual(true);
  });
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect-in-promise',
    code: `
       test('foo', () => {
         somePromise['then'](() => {
           expect(someThing).toEqual(true);
         });
       });
      `,
  },
];

const CLEAN: DiffCase[] = [
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'missing-playwright-await',
    code: "test('test', async () => { await expect(page).toBeEditable })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'missing-playwright-await',
    code: 'test(\'test\', async () => { await expect(page).toEqualTitle("text") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'missing-playwright-await',
    code: 'test(\'test\', async () => { await expect(page).not.toHaveText("text") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-expect',
    code: `test('foo', () => {
  expect(1).toBe(2);
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-expect',
    code: `test('foo', () => {
  expect(!true).toBe(false);
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-conditional-expect',
    code: `test('foo', () => {
  const expected = arr.map((x) => {
    if (typeof x === 'string') {
      return expect.arrayContaining([x])
    } else {
      return b
    }
  })

  expect([1, 2, 3]).toEqual(expected);
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-matchers',
    code: 'expect(a)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-matchers',
    code: 'expect(a).toBe()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-matchers',
    code: 'expect(a).not.toContain()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-not',
    code: 'expect(locator).toBeVisible()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-not',
    code: 'expect(locator).toBeHidden()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-not',
    code: 'expect(locator).toBeEnabled()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-comparison-matcher',
    code: 'expect(value).toBeGreaterThan(1);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-comparison-matcher',
    code: 'expect(value).toBeLessThanOrEqual(1);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-equality-matcher',
    code: 'expect(value).toBe(1);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-equality-matcher',
    code: 'expect(value).not.toBe(1);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-strict-equal',
    code: 'expect(something).toStrictEqual(somethingElse);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-strict-equal',
    code: "a().toEqual('b')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-strict-equal',
    code: 'expect(a);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-be',
    code: 'expect(null).toBeNull();',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-be',
    code: 'expect(null).not.toBeNull();',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-be',
    code: 'expect(null).toBe(1);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-contain',
    code: 'expect().toBe(false);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-contain',
    code: 'expect(a).toContain(b);',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-contain',
    code: "expect(a.name).toBe('b');",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-count',
    code: 'await expect(files).toHaveCount(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-count',
    code: 'await expect(files).toHaveCount(foo)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-count',
    code: 'await expect(files.length).toBe(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-length',
    code: 'expect(files).toHaveLength(1)',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-length',
    code: "expect(files.name).toBe('file')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-to-have-length',
    code: "expect(files['name']).toBe('file')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-web-first-assertions',
    code: 'test(\'test\', async () => { await expect(page.locator(".tweet")).toBeVisible() })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-web-first-assertions',
    code: "test('test', async () => { await expect(bar).toBeEnabled() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-web-first-assertions',
    code: 'test(\'test\', async () => { await expect.soft(page.locator(".tweet")).toBeDisabled() })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-soft-assertions',
    code: 'expect.soft(page).toHaveTitle("baz")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-soft-assertions',
    code: 'expect.soft(page.locator("foo")).toHaveText("bar")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-soft-assertions',
    code: 'expect["soft"](foo).toBe("bar")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-pass-timeout',
    code: 'await expect(async () => { await foo() }).toPass({ timeout: 5000 });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-pass-timeout',
    code: 'await expect(async () => { await foo() }).toPass({ timeout: 5000, intervals: [1000] });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-pass-timeout',
    code: 'await expect.soft(async () => { await foo() }).toPass({ timeout: 5000 });',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-throw-message',
    code: "expect(() => { throw new Error('a'); }).toThrow('a');",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-throw-message',
    code: "expect(() => { throw new Error('a'); }).toThrowError('a');",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'require-to-throw-message',
    code: `test('string', async () => {
  const throwErrorAsync = async () => { throw new Error('a') };
  await expect(throwErrorAsync()).rejects.toThrow('a');
  await expect(throwErrorAsync()).rejects.toThrowError('a');
})`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect',
    code: 'expectPayButtonToBeEnabled()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect',
    code: 'expect("something").toBe("else")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect',
    code: 'expect.anything()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect-in-promise',
    code: "test('something', () => Promise.resolve().then(() => expect(1).toBe(2)));",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect-in-promise',
    code: 'Promise.resolve().then(() => expect(1).toBe(2))',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'valid-expect-in-promise',
    code: 'const x = Promise.resolve().then(() => expect(1).toBe(2))',
  },
];

runConformanceSuite('eslint-plugin-playwright', CASES, CLEAN);
