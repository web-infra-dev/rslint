/**
 * @fileoverview Tests for curly-newline rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/curly-newline/curly-newline.test.ts
 *
 * The upstream file does NOT use literal `valid`/`invalid` arrays. It builds them
 * imperatively via two helpers and then calls `run({ name, rule, valid, invalid })`
 * once. Both helpers are evaluated to their final cases here:
 *
 *  - `test(options, code)`                      -> a `valid` case `{ code, options }`.
 *  - `test(options, code, output, ...errors)`   -> an `invalid` case
 *      `{ code, output, options, errors }` (output may be `null` = source unchanged).
 *  - `specializationTest(spec, code, output?, openCol?, closeCol?)` wraps `code` in
 *      `{\n<code>\n}`, sets options `[{ [spec]: 'always' }]`, and emits:
 *        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
 *        // only when an `output` arg was passed:
 *        { line: 2, column: openCol,  messageId: 'expectedLinebreakAfterOpeningBrace' },
 *        { line: 2, column: closeCol, messageId: 'expectedLinebreakBeforeClosingBrace' },
 *        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' }
 *      with `output` = `{<output ?? code>}`.
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('curly-newline', null as never, { valid, invalid })`.
 *  - The `name` / `rule` fields and all imports of upstream test infra are dropped.
 *  - `options` of `undefined` (upstream default) becomes `options: []`.
 *  - The four messageIds carry static templates (no `{{data}}`), so each asserts an
 *    exact message; line/column/endLine/endColumn are asserted only where upstream
 *    pinned them.
 *
 * There are no `$`/unindent template tags, no Babel/Flow cases, no `skipBabel`
 * blocks, no spread error helpers beyond the two functions above, and no
 * suggestions in this rule's tests. The `._css_` / `._json_` / `._markdown_`
 * test files don't exist for this rule.
 *
 * Every upstream case is ported into the green set: all fixtures (including the
 * `with` statement and the TS `module _{}` namespace block) parse and behave
 * identically under rslint/ts-go, verified against the live CLI. There are no
 * KNOWN GAPS for this rule.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('curly-newline', null as never, {
  valid: [
    // default ------------------------------------------------------------
    { code: '{}', options: [] },
    { code: '{\n}', options: [] },
    { code: '{void {foo}}', options: [] },
    { code: '{\nvoid {foo}\n}', options: [] },

    // "always" -----------------------------------------------------------
    { code: '{\n}', options: ['always'] },
    { code: '{\nvoid {foo}\n}', options: ['always'] },
    { code: '{\nvoid {foo};void {foo}\n}', options: ['always'] },
    { code: '{\nvoid {foo}\nvoid {foo}\n}', options: ['always'] },

    // "never" ------------------------------------------------------------
    { code: '{}', options: ['never'] },
    { code: '{void {foo}}', options: ['never'] },
    { code: '{void {foo};void {foo}}', options: ['never'] },
    { code: '{void {foo}\nvoid {foo}}', options: ['never'] },
    { code: '{void {\nfoo\n}}', options: ['never'] },

    // "multiline" --------------------------------------------------------
    { code: '{}', options: [{ multiline: true }] },
    { code: '{void {}}', options: [{ multiline: true }] },
    { code: '{void {}; void {}}', options: [{ multiline: true }] },
    { code: '{\nvoid {}\nvoid {}\n}', options: [{ multiline: true }] },
    { code: '{\nvoid {\nfoo\n}\n}', options: [{ multiline: true }] },
    { code: '{\n// comment\nvoid {}\n}', options: [{ multiline: true }] },
    { code: '{ // comment\nvoid {}\n}', options: [{ multiline: true }] },

    // "minElements" ------------------------------------------------------
    { code: '{}', options: [{ minElements: 2 }] },
    { code: '{void 0}', options: [{ minElements: 2 }] },
    { code: '{\nvoid 0;void 0\n}', options: [{ minElements: 2 }] },

    // "multiline" and "minElements" --------------------------------------
    { code: '{}', options: [{ multiline: true, minElements: 2 }] },
    { code: '{void 0}', options: [{ multiline: true, minElements: 2 }] },
    { code: '{\nvoid 0; void 0\n}', options: [{ multiline: true, minElements: 2 }] },
    { code: '{\nvoid 0\nvoid 0\n}', options: [{ multiline: true, minElements: 2 }] },
    { code: '{\nvoid {\nfoo\n}\n}', options: [{ multiline: true, minElements: 2 }] },

    // "consistent" -------------------------------------------------------
    { code: '{\nvoid 0\n}', options: [{ consistent: true }] },
    { code: '{\nvoid 0\nvoid 0\n}', options: [{ consistent: true }] },
    { code: '{void {foo}}', options: [{ consistent: true }] },
    { code: '{\nvoid {foo}\n}', options: [{ consistent: true }] },
    { code: '{\nvoid {\nfoo\n}\n}', options: [{ consistent: true }] },
    { code: '{void 0;void 0}', options: [{ consistent: true }] },

    // "consistent" and "minElements" -------------------------------------
    { code: '{void 0}', options: [{ multiline: true, consistent: true, minElements: 2 }] },
    { code: '{\nvoid 0\n}', options: [{ multiline: true, consistent: true, minElements: 2 }] },
    { code: '{\nvoid 0;void 0\n}', options: [{ multiline: true, consistent: true, minElements: 2 }] },
  ],
  invalid: [
    // default ------------------------------------------------------------
    {
      code: '{void {foo}\n}',
      output: '{void {foo}}',
      options: [],
      errors: [{ line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' }],
    },
    {
      code: '{\nvoid {foo}}',
      output: '{void {foo}}',
      options: [],
      errors: [{ line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' }],
    },

    // "always" -----------------------------------------------------------
    {
      code: '{}',
      output: '{\n}',
      options: ['always'],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 1, column: 2, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {foo}}',
      output: '{\nvoid {foo}\n}',
      options: ['always'],
      errors: [
        { line: 1, column: 1, endLine: 1, endColumn: 2, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 1, column: 12, endLine: 1, endColumn: 13, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {foo};void {foo}}',
      output: '{\nvoid {foo};void {foo}\n}',
      options: ['always'],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 1, column: 23, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {foo};\n  void {foo}}',
      output: '{\nvoid {foo};\n  void {foo}\n}',
      options: ['always'],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 13, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {\n}}',
      output: '{\nvoid {\n}\n}',
      options: ['always'],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 2, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{ void {foo} }',
      output: '{\n void {foo} \n}',
      options: ['always'],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 1, column: 14, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },

    // "never" ------------------------------------------------------------
    {
      code: '{\n}',
      output: '{}',
      options: ['never'],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {foo}}',
      output: '{void {foo}}',
      options: ['never'],
      errors: [{ line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' }],
    },
    {
      code: '{void {foo}\n}',
      output: '{void {foo}}',
      options: ['never'],
      errors: [{ line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' }],
    },
    {
      code: '{\nvoid {foo}\n}',
      output: '{void {foo}}',
      options: ['never'],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {foo}\nvoid {foo}\n}',
      output: '{void {foo}\nvoid {foo}}',
      options: ['never'],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 4, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {\nfoo\n}\n}',
      output: '{void {\nfoo\n}}',
      options: ['never'],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 5, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },

    // "multiline" --------------------------------------------------------
    {
      code: '{\n}',
      output: '{}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\n/*comment*/\n}',
      output: null,
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{// comment\n}',
      output: null,
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {}\n}',
      output: '{void {}}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {} // comment\n}',
      output: '{void {} // comment\n}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {};void {}\n}',
      output: '{void {};void {}}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {};void {} // comment\n}',
      output: '{void {};void {} // comment\n}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {}\n  void {}}',
      output: '{\nvoid {}\n  void {}\n}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {\n}}',
      output: '{\nvoid {\n}\n}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 2, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {} // comment\n  void {}}',
      output: '{\nvoid {} // comment\n  void {}\n}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {\nfoo\n}}',
      output: '{\nvoid {\nfoo\n}\n}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 2, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void { // comment\nfoo\n}}',
      output: '{\nvoid { // comment\nfoo\n}\n}',
      options: [{ multiline: true }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 2, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {} /* comment */\nvoid {}\n}',
      output: '{\nvoid {} /* comment */\nvoid {}\n}',
      options: [{ multiline: true }],
      errors: [{ line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' }],
    },
    {
      code: '{/* comment */void {}\nvoid {}\n}',
      output: null,
      options: [{ multiline: true }],
      errors: [{ line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' }],
    },
    {
      code: '{\n/* comment */\nvoid {}}',
      output: '{\n/* comment */\nvoid {}\n}',
      options: [{ multiline: true }],
      errors: [{ line: 3, column: 8, messageId: 'expectedLinebreakBeforeClosingBrace' }],
    },

    // "minElements" ------------------------------------------------------
    {
      code: '{\n}',
      output: '{}',
      options: [{ minElements: 2 }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {}\n}',
      output: '{void {}}',
      options: [{ minElements: 2 }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {};void {}}',
      output: '{\nvoid {};void {}\n}',
      options: [{ minElements: 2 }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 1, column: 17, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },

    // "multiline" and "minElements" --------------------------------------
    {
      code: '{\n}',
      output: '{}',
      options: [{ multiline: true, minElements: 2 }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {}\n}',
      output: '{void {}}',
      options: [{ multiline: true, minElements: 2 }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {};void {}}',
      output: '{\nvoid {};void {}\n}',
      options: [{ multiline: true, minElements: 2 }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 1, column: 17, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {};\n  void {}}',
      output: '{\nvoid {};\n  void {}\n}',
      options: [{ multiline: true, minElements: 2 }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {\n}}',
      output: '{\nvoid {\n}\n}',
      options: [{ multiline: true, minElements: 2 }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 2, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },

    // "consistent" -------------------------------------------------------
    {
      code: '{void {}\n}',
      output: '{void {}}',
      options: [{ consistent: true }],
      errors: [{ line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' }],
    },
    {
      code: '{\nvoid {}}',
      output: '{void {}}',
      options: [{ consistent: true }],
      errors: [{ line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' }],
    },
    {
      code: '{void {};void {}\n}',
      output: '{void {};void {}}',
      options: [{ consistent: true }],
      errors: [{ line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' }],
    },
    {
      code: '{\nvoid {};void {}}',
      output: '{void {};void {}}',
      options: [{ consistent: true }],
      errors: [{ line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' }],
    },
    {
      code: '{\nvoid {}\nvoid {}}',
      output: '{void {}\nvoid {}}',
      options: [{ consistent: true }],
      errors: [{ line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' }],
    },

    // "multiline" and "consistent" -------------------------------------
    {
      code: '{void {}\n}',
      output: '{void {}}',
      options: [{ multiline: true, consistent: true }],
      errors: [{ line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' }],
    },
    {
      code: '{\nvoid {}}',
      output: '{void {}}',
      options: [{ multiline: true, consistent: true }],
      errors: [{ line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' }],
    },
    {
      code: '{void {};void {}\n}',
      output: '{void {};void {}}',
      options: [{ multiline: true, consistent: true }],
      errors: [{ line: 2, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' }],
    },
    {
      code: '{\nvoid {};void {}}',
      output: '{void {};void {}}',
      options: [{ multiline: true, consistent: true }],
      errors: [{ line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' }],
    },
    {
      code: '{void {}\n  void {}}',
      output: '{\nvoid {}\n  void {}\n}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{void {\n}}',
      output: '{\nvoid {\n}\n}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { line: 1, column: 1, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 2, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    {
      code: '{\nvoid {}\nvoid {}}',
      output: '{\nvoid {}\nvoid {}\n}',
      options: [{ multiline: true, consistent: true }],
      errors: [{ line: 3, column: 8, messageId: 'expectedLinebreakBeforeClosingBrace' }],
    },

    // specializations ----------------------------------------------------
    {
      code: 'for(;;){{}}',
      output: 'for(;;){{\n}}',
      options: [{ BlockStatement: 'always' }],
      errors: [
        { line: 1, column: 9, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 1, column: 10, messageId: 'expectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('IfStatementConsequent', `if(true){}`, `if(true){\n}`, 9, 10)
    {
      code: '{\nif(true){}\n}',
      output: '{if(true){\n}}',
      options: [{ IfStatementConsequent: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 9, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('IfStatementAlternative', `if(true){}else{}`, `if(true){}else{\n}`, 15, 16)
    {
      code: '{\nif(true){}else{}\n}',
      output: '{if(true){}else{\n}}',
      options: [{ IfStatementAlternative: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 15, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 16, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('DoWhileStatement', `do{}while(true)`, `do{\n}while(true)`, 3, 4)
    {
      code: '{\ndo{}while(true)\n}',
      output: '{do{\n}while(true)}',
      options: [{ DoWhileStatement: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 3, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 4, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('ForInStatement', `for(const {} in {}){}`, `for(const {} in {}){\n}`, 20, 21)
    {
      code: '{\nfor(const {} in {}){}\n}',
      output: '{for(const {} in {}){\n}}',
      options: [{ ForInStatement: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 20, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 21, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('ForOfStatement', `for(const {} of {}){}`, `for(const {} of {}){\n}`, 20, 21)
    {
      code: '{\nfor(const {} of {}){}\n}',
      output: '{for(const {} of {}){\n}}',
      options: [{ ForOfStatement: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 20, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 21, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('ForStatement', `for(;;){}`, `for(;;){\n}`, 8, 9)
    {
      code: '{\nfor(;;){}\n}',
      output: '{for(;;){\n}}',
      options: [{ ForStatement: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 8, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 9, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('WhileStatement', `while({}){}`, `while({}){\n}`, 10, 11)
    {
      code: '{\nwhile({}){}\n}',
      output: '{while({}){\n}}',
      options: [{ WhileStatement: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 11, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('SwitchStatement', `switch({}){}`, `switch({}){\n}`, 11, 12)
    {
      code: '{\nswitch({}){}\n}',
      output: '{switch({}){\n}}',
      options: [{ SwitchStatement: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 11, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 12, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('SwitchStatement', `switch({}){case {}:}`, `switch({}){\ncase {}:\n}`, 11, 20)
    {
      code: '{\nswitch({}){case {}:}\n}',
      output: '{switch({}){\ncase {}:\n}}',
      options: [{ SwitchStatement: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 11, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 20, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('SwitchCase', `switch({}){case {}:}`)  — no output arg
    {
      code: '{\nswitch({}){case {}:}\n}',
      output: '{switch({}){case {}:}}',
      options: [{ SwitchCase: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('SwitchCase', `switch({}){case {}: {}break}`)  — no output arg
    {
      code: '{\nswitch({}){case {}: {}break}\n}',
      output: '{switch({}){case {}: {}break}}',
      options: [{ SwitchCase: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('SwitchCase', `switch({}){case {}: {}}`, `switch({}){case {}: {\n}}`, 21, 22)
    {
      code: '{\nswitch({}){case {}: {}}\n}',
      output: '{switch({}){case {}: {\n}}}',
      options: [{ SwitchCase: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 21, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 22, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('TryStatementBlock', `try{}finally{}`, `try{\n}finally{}`, 4, 5)
    {
      code: '{\ntry{}finally{}\n}',
      output: '{try{\n}finally{}}',
      options: [{ TryStatementBlock: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 4, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 5, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('TryStatementHandler', `try{}catch(_){}`, `try{}catch(_){\n}`, 14, 15)
    {
      code: '{\ntry{}catch(_){}\n}',
      output: '{try{}catch(_){\n}}',
      options: [{ TryStatementHandler: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 14, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 15, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('TryStatementFinalizer', `try{}finally{}`, `try{}finally{\n}`, 13, 14)
    {
      code: '{\ntry{}finally{}\n}',
      output: '{try{}finally{\n}}',
      options: [{ TryStatementFinalizer: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 13, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 14, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('ArrowFunctionExpression', `(() => {})`, `(() => {\n})`, 8, 9)
    {
      code: '{\n(() => {})\n}',
      output: '{(() => {\n})}',
      options: [{ ArrowFunctionExpression: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 8, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 9, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('FunctionDeclaration', `function _() {}`, `function _() {\n}`, 14, 15)
    {
      code: '{\nfunction _() {}\n}',
      output: '{function _() {\n}}',
      options: [{ FunctionDeclaration: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 14, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 15, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('FunctionExpression', `(function() {})`, `(function() {\n})`, 13, 14)
    {
      code: '{\n(function() {})\n}',
      output: '{(function() {\n})}',
      options: [{ FunctionExpression: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 13, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 14, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('Property', `void {_: function(){}}`)  — no output arg
    {
      code: '{\nvoid {_: function(){}}\n}',
      output: '{void {_: function(){}}}',
      options: [{ Property: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('Property', `void {_(){}}`, `void {_(){\n}}`, 10, 11)
    {
      code: '{\nvoid {_(){}}\n}',
      output: '{void {_(){\n}}}',
      options: [{ Property: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 11, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('ClassBody', `(class {})`, `(class {\n})`, 8, 9)
    {
      code: '{\n(class {})\n}',
      output: '{(class {\n})}',
      options: [{ ClassBody: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 8, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 9, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('StaticBlock', `(class {static{}})`, `(class {static{\n}})`, 15, 16)
    {
      code: '{\n(class {static{}})\n}',
      output: '{(class {static{\n}})}',
      options: [{ StaticBlock: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 15, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 16, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('WithStatement', `with({}){}`, `with({}){\n}`, 9, 10)
    {
      code: '{\nwith({}){}\n}',
      output: '{with({}){\n}}',
      options: [{ WithStatement: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 9, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
    // specializationTest('TSModuleBlock', `module _{}`, `module _{\n}`, 9, 10)
    {
      code: '{\nmodule _{}\n}',
      output: '{module _{\n}}',
      options: [{ TSModuleBlock: 'always' }],
      errors: [
        { line: 1, column: 1, messageId: 'unexpectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 9, messageId: 'expectedLinebreakAfterOpeningBrace' },
        { line: 2, column: 10, messageId: 'expectedLinebreakBeforeClosingBrace' },
        { line: 3, column: 1, messageId: 'unexpectedLinebreakBeforeClosingBrace' },
      ],
    },
  ],
});
