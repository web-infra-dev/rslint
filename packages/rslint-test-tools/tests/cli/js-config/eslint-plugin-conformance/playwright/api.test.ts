/**
 * Conformance: eslint-plugin-playwright (api) mounted in rslint via `plugins`
 * must report identically to ESLint v10. playwright rules are AST pattern
 * matches over test/expect/page call shapes (no type info), so rslint
 * reproduces ESLint byte-for-byte. Representative triggers from the upstream
 * test suite (v2.10.4): locators, page actions, wait-for patterns.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-element-handle',
    code: 'test(\'test\', async () => { const handle = await page.$("text=Submit"); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-element-handle',
    code: 'test(\'test\', async () => { const handle = await this.page.$("text=Submit"); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-element-handle',
    code: 'test(\'test\', async () => { const handle = await page["$$"]("text=Submit"); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-element-handle',
    code: 'test(\'test\', async () => { const handle = await page[`$$`]("text=Submit"); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-eval',
    code: 'test(\'test\', async () => { const searchValue = await page.$eval("#search", el => el.value); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-eval',
    code: 'test(\'test\', async () => { const searchValue = await this.page.$eval("#search", el => el.value); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-eval',
    code: 'test(\'test\', async () => { const searchValue = await page["$eval"]("#search", el => el.value); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-eval',
    code: 'test(\'test\', async () => { const searchValue = await page[`$eval`]("#search", el => el.value); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-force-option',
    code: 'test(\'test\', async () => { await page.locator("check").check({ force: true }) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-force-option',
    code: 'test(\'test\', async () => { await page.locator("check").uncheck({ ["force"]: true }) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-force-option',
    code: 'test(\'test\', async () => { await page.locator("button").click({ [`force`]: true }) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-force-option',
    code: `test('test', async () => { 
        const button = page["locator"]("button")
        await button.click({ force: true })
       })`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-get-by-title',
    code: 'test(\'test\', async () => { await page.getByTitle("lorem ipsum") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-get-by-title',
    code: 'test(\'test\', async () => { await this.page.getByTitle("lorem ipsum") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-get-by-title',
    code: 'test(\'test\', async () => { await page.locator("div").getByTitle("lorem ipsum") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-networkidle',
    code: 'page.waitForLoadState("networkidle")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-networkidle',
    code: 'page.waitForURL(url, { waitUntil: "networkidle" })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-networkidle',
    code: 'page["waitForURL"](url, { waitUntil: "networkidle" })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-networkidle',
    code: 'page[`waitForURL`](url, { waitUntil: "networkidle" })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nth-methods',
    code: 'page.locator("button").first()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nth-methods',
    code: 'frame.locator("button").first()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nth-methods',
    code: 'foo.locator("button").first()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nth-methods',
    code: 'foo.first()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-page-pause',
    code: "test('test', async () => { await page.pause() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-page-pause',
    code: "test('test', async () => { await this.page.pause() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-page-pause',
    code: 'test(\'test\', async () => { await page["pause"]() })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-page-pause',
    code: "test('test', async () => { await page[`pause`]() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-raw-locators',
    code: "test('test', async () => { await page.locator() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-raw-locators',
    code: "test('test', async () => { const locator = await page.locator() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-raw-locators',
    code: "test('test', async () => { let locator = await page.locator() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-raw-locators',
    code: "test('test', async () => { await this.page.locator() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-locators',
    code: 'test(\'test\', async () => { await page.getByTestId("button") })',
    options: [['getByTestId']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-locators',
    code: 'test(\'test\', async () => { await this.page.getByTestId("button") })',
    options: [['getByTestId']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-locators',
    code: 'test(\'test\', async () => { await page["getByTestId"]("button") })',
    options: [['getByTestId']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-locators',
    code: 'test(\'test\', async () => { await page[`getByTestId`]("button") })',
    options: [['getByTestId']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-roles',
    code: 'test(\'test\', async () => { await page.getByRole("progressbar") })',
    options: [['progressbar']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-roles',
    code: 'test(\'test\', async () => { await this.page.getByRole("progressbar") })',
    options: [['progressbar']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-roles',
    code: 'test(\'test\', async () => { await page["getByRole"]("progressbar") })',
    options: [['progressbar']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-roles',
    code: 'test(\'test\', async () => { await page[`getByRole`]("progressbar") })',
    options: [['progressbar']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unsafe-references',
    code: `const x = 10
const result = await page.evaluate(() => {
  return Promise.resolve(x);
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unsafe-references',
    code: `const x = 10
const result = await page.addInitScript(() => {
  return Promise.resolve(x);
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unsafe-references',
    code: `const x = 10
const result = await page.evaluate(function () {
  return Promise.resolve(x);
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unsafe-references',
    code: `const x = 10
const result = await page.addInitScript(function () {
  return Promise.resolve(x);
});`,
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unused-locators',
    code: "page.getByRole('button', { name: 'Sign in' })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unused-locators',
    code: "await page.getByTestId('my-test-id')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unused-locators',
    code: "page.locator('.btn')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unused-locators',
    code: "await page.locator('.btn')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-await',
    code: 'await page.locator(".my-element")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-await',
    code: 'await frame.locator(".my-element")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-await',
    code: 'await foo.locator(".my-element")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-await',
    code: 'await foo.first()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-navigation',
    code: 'const navigationPromise = page.waitForNavigation();',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-navigation',
    code: 'async function fn() { await page.waitForNavigation() }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-navigation',
    code: 'async function fn() { await this.page.waitForNavigation() }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-navigation',
    code: 'async function fn() { await page["waitForNavigation"]() }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-selector',
    code: 'async function fn() { await page.waitForSelector(".bar") }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-selector',
    code: 'async function fn() { await this.page.waitForSelector(".bar") }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-selector',
    code: 'async function fn() { await page["waitForSelector"](".bar") }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-selector',
    code: 'async function fn() { await page[`waitForSelector`](".bar") }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-timeout',
    code: 'async function fn() { await page.waitForTimeout(1000) }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-timeout',
    code: 'async function fn() { await this.page.waitForTimeout(1000) }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-timeout',
    code: 'async function fn() { await page["waitForTimeout"](1000) }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-timeout',
    code: 'async function fn() { await page[`waitForTimeout`](1000) }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-locator',
    code: "test('test', async () => { await page.fill('input[type=\"password\"]', 'password') })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-locator',
    code: "test('test', async () => { await page.dblclick('xpath=//button') })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-locator',
    code: "page.click('xpath=//button')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-locator',
    code: "test('test', async () => { await page.frame('frame-name').click('css=button') })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-native-locators',
    code: 'page.locator(\'[aria-label="View more"]\')',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-native-locators',
    code: "page.locator('[aria-label=Edit]')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-native-locators',
    code: 'page.locator(\'[role="button"]\')',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-native-locators',
    code: "page.locator('[role=button]')",
  },
];

const CLEAN: DiffCase[] = [
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-element-handle',
    code: 'test(\'test\', async () => { page.locator("a") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-element-handle',
    code: 'test(\'test\', async () => { this.page.locator("a") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-element-handle',
    code: 'test(\'test\', async () => { await page.locator("a").click(); })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-eval',
    code: 'test(\'test\', async () => { await page.locator(".tweet").evaluate(node => node.innerText) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-eval',
    code: 'test(\'test\', async () => { await this.page.locator(".tweet").evaluate(node => node.innerText) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-eval',
    code: 'test(\'test\', async () => { await page.locator(".tweet")["evaluate"](node => node.innerText) })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-force-option',
    code: "test('test', async () => { await page.locator('check').check() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-force-option',
    code: "test('test', async () => { await page.locator('check').uncheck() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-force-option',
    code: "test('test', async () => { await page.locator('button').click() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-get-by-title',
    code: 'test(\'test\', async () => { await page.locator("[title=lorem ipsum]") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-get-by-title',
    code: 'test(\'test\', async () => { await page.getByRole("button") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-networkidle',
    code: 'foo("networkidle")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-networkidle',
    code: 'foo(url, { waitUntil: "networkidle" })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-networkidle',
    code: 'foo.bar("networkidle")',
  },
  { pkg: 'eslint-plugin-playwright', rule: 'no-nth-methods', code: 'page' },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nth-methods',
    code: 'page.locator("button")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-nth-methods',
    code: 'frame.locator("button")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-page-pause',
    code: "test('test', async () => { await page.click() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-page-pause',
    code: "test('test', async () => { await this.page.click() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-page-pause',
    code: 'test(\'test\', async () => { await page["hover"]() })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-raw-locators',
    code: "test('test', async () => { await page.click() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-raw-locators',
    code: "test('test', async () => { await this.page.click() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-raw-locators',
    code: 'test(\'test\', async () => { await page["hover"]() })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-locators',
    code: 'test(\'test\', async () => { await page.getByTestId("button") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-locators',
    code: 'test(\'test\', async () => { await page.getByTitle("tooltip") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-locators',
    code: 'test(\'test\', async () => { await page.getByRole("button") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-roles',
    code: 'test(\'test\', async () => { await page.getByRole("button") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-roles',
    code: 'test(\'test\', async () => { await page.getByRole("progressbar") })',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-restricted-roles',
    code: 'test(\'test\', async () => { await page.getByRole("heading") })',
    options: [['progressbar']],
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unsafe-references',
    code: 'page.pause()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unsafe-references',
    code: 'page.evaluate()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unsafe-references',
    code: 'page.evaluate("1 + 2")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unused-locators',
    code: "await page.getByRole('button', { name: 'Sign in' }).all()",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unused-locators',
    code: "const btn = page.getByLabel('Sign in')",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-unused-locators',
    code: "const btn = page.getByPlaceholder('User Name').first()",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-await',
    code: 'await foo()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-await',
    code: 'await foo(".my-element")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-useless-await',
    code: 'await foo.bar()',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-navigation',
    code: 'function waitForNavigation() {}',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-navigation',
    code: 'async function fn() { await waitForNavigation(); }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-navigation',
    code: 'async function fn() { await this.foo.waitForNavigation(); }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-selector',
    code: 'function waitForSelector() {}',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-selector',
    code: 'async function fn() { await waitForSelector("#foo"); }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-selector',
    code: 'async function fn() { await this.foo.waitForSelector("#foo"); }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-timeout',
    code: 'function waitForTimeout() {}',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-timeout',
    code: 'async function fn() { await waitForTimeout(4343); }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'no-wait-for-timeout',
    code: 'async function fn() { await this.foo.waitForTimeout(4343); }',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-locator',
    code: 'const locator = page.locator(\'input[type="password"]\')',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-locator',
    code: "test('test', async () => { await page.locator('input[type=\"password\"]').fill('password') })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-locator',
    code: "test('test', async () => { await page.locator('xpath=//button').dblclick() })",
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-native-locators',
    code: 'page.getByLabel("View more")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-native-locators',
    code: 'page.getByRole("button")',
  },
  {
    pkg: 'eslint-plugin-playwright',
    rule: 'prefer-native-locators',
    code: 'page.getByRole("button", {name: "Open"})',
  },
];

runConformanceSuite('eslint-plugin-playwright', CASES, CLEAN);
