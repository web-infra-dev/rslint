/**
 * Conformance: eslint-plugin-cypress (v6.4.1) mounted in rslint via `plugins`
 * must report identically to ESLint v10. All 13 rules covered; each case
 * reproduces ESLint v10 byte-for-byte (the rules are AST pattern matches over
 * Cypress `cy.*` command chains, no type info). Triggers from the upstream
 * test suite.
 */
import { runConformanceSuite } from './conformance.js';
import type { DiffCase } from './harness.js';

const CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'assertion-before-screenshot',
    code: 'cy.screenshot()',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'assertion-before-screenshot',
    code: 'cy.visit("somepage"); cy.screenshot();',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'assertion-before-screenshot',
    code: 'cy.custom(); cy.screenshot()',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'assertion-before-screenshot',
    code: 'cy.get(".some-element").click(); cy.screenshot()',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-and',
    code: "cy.and('be.visible')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-and',
    code: "cy.get('elem').and('have.text', 'blah')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-and',
    code: "cy.get('foo').and('be.visible')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-and',
    code: "cy.get('foo').find('.bar').and('have.class', 'active')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-assigning-return-values',
    code: 'let a = cy.get("foo")',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-assigning-return-values',
    code: 'const a = cy.get("foo")',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-assigning-return-values',
    code: 'var a = cy.get("foo")',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-assigning-return-values',
    code: 'let a = cy.contains("foo")',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-before',
    code: "before('a test case', async () => { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-before',
    code: "beforeEach('a test case', async () => { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-before',
    code: "before('a test case', async function () { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-before',
    code: "beforeEach('a test case', async function () { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-tests',
    code: "it('a test case', async () => { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-tests',
    code: "test('a test case', async () => { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-tests',
    code: "it('a test case', async function () { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-tests',
    code: "test('a test case', async function () { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-chained-get',
    code: "cy.get('div').get('div')",
  },
  { pkg: 'eslint-plugin-cypress', rule: 'no-debug', code: 'cy.debug()' },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-debug',
    code: 'cy.debug({ log: false })',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-debug',
    code: "cy.get('button').debug()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-debug',
    code: "cy.get('a').should('have.attr', 'href').and('match', /dashboard/).debug()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-force',
    code: "cy.get('button').click({force: true})",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-force',
    code: "cy.get('button').dblclick({force: true})",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-force',
    code: "cy.get('input').type('somth', {force: true})",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-force',
    code: "cy.get('div').find('.foo').type('somth', {force: true})",
  },
  { pkg: 'eslint-plugin-cypress', rule: 'no-pause', code: 'cy.pause()' },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-pause',
    code: 'cy.pause({ log: false })',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-pause',
    code: "cy.get('button').pause()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-pause',
    code: "cy.get('a').should('have.attr', 'href').and('match', /dashboard/).pause()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-unnecessary-waiting',
    code: 'cy.wait(0)',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-unnecessary-waiting',
    code: 'cy.wait(100)',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-unnecessary-waiting',
    code: 'cy.wait(5000)',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-unnecessary-waiting',
    code: 'const someNumber=500; cy.wait(someNumber)',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-xpath',
    code: 'cy.xpath(\'//div[@class="container"]/p[1]\').click()',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-xpath',
    code: "cy.xpath('//p[1]').should('exist')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'require-data-selectors',
    code: "cy.get('[daedta-cy=submit]').click()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'require-data-selectors',
    code: "cy.get('[d-cy=submit]')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'require-data-selectors',
    code: 'cy.get(".btn-large").click()',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'require-data-selectors',
    code: 'cy.get(".btn-.large").click()',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'unsafe-to-chain-command',
    code: 'cy.get("new-todo").type("todo A{enter}").should("have.class", "active");',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'unsafe-to-chain-command',
    code: 'cy.get("new-todo").type("todo A{enter}").type("todo B{enter}");',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'unsafe-to-chain-command',
    code: 'cy.get("new-todo").focus().should("have.class", "active");',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'unsafe-to-chain-command',
    code: 'cy.get("new-todo").customType("todo A{enter}").customClick();',
    options: [{ methods: ['customType', 'customClick'] }],
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'assertion-before-screenshot',
    code: 'cy.get(".some-element"); cy.screenshot();',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'assertion-before-screenshot',
    code: 'cy.get(".some-element").should("exist").screenshot();',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'assertion-before-screenshot',
    code: 'cy.get(".some-element").should("exist").screenshot().click()',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-and',
    code: "cy.get('elem').should('have.text', 'blah')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-and',
    code: "cy.get('foo').should('be.visible')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-and',
    code: "cy.get('foo').should('be.visible').should('have.text', 'bar')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-assigning-return-values',
    code: 'var foo = true;',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-assigning-return-values',
    code: 'let foo = true;',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-assigning-return-values',
    code: 'const foo = true;',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-before',
    code: "before('a before case', () => { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-before',
    code: "before('a before case', async () => { await somethingAsync(); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-before',
    code: 'async function nonTestFn () { return await somethingAsync(); }',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-tests',
    code: "it('a test case', () => { cy.get('.someClass'); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-tests',
    code: "it('a test case', async () => { await somethingAsync(); })",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-async-tests',
    code: 'async function nonTestFn () { return await somethingAsync(); }',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-chained-get',
    code: "cy.get('div')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-chained-get',
    code: "cy.get('.div').find().get()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-chained-get',
    code: "cy.get('input').should('be.disabled')",
  },
  { pkg: 'eslint-plugin-cypress', rule: 'no-debug', code: 'debug()' },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-debug',
    code: "cy.get('button').dblclick()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-force',
    code: "cy.get('button').click()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-force',
    code: "cy.get('button').click({multiple: true})",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-force',
    code: "cy.get('button').dblclick()",
  },
  { pkg: 'eslint-plugin-cypress', rule: 'no-pause', code: 'pause()' },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-pause',
    code: "cy.get('button').dblclick()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-unnecessary-waiting',
    code: 'foo.wait(10)',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-unnecessary-waiting',
    code: 'cy.wait("@someRequest")',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-unnecessary-waiting',
    code: 'cy.wait("@someRequest", { log: false })',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'no-xpath',
    code: 'cy.get("button").click({force: true})',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'require-data-selectors',
    code: "cy.get('[data-cy=submit]').click()",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'require-data-selectors',
    code: "cy.get('[data-QA=submit]')",
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'require-data-selectors',
    code: 'cy.clock(5000)',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'unsafe-to-chain-command',
    code: 'cy.focused().should("be.visible");',
  },
  {
    pkg: 'eslint-plugin-cypress',
    rule: 'unsafe-to-chain-command',
    code: 'cy.submitBtn().click();',
  },
];

runConformanceSuite('eslint-plugin-cypress', CASES, CLEAN_CASES);
