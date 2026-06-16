/**
 * @fileoverview Tests for quote-props rule.
 * @author Mathias Bynens <http://mathiasbynens.be/>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/quote-props/quote-props._js_.test.ts
 *   packages/eslint-plugin/rules/quote-props/quote-props._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('quote-props', null as never, { valid, invalid })`
 *  - `parserOptions` (ecmaVersion) dropped — rslint resolves via tsconfig (target
 *    `esnext`), so every ES2018/2020/2021/2022 feature in these fixtures
 *    (object spread, BigInt keys, numeric separators) parses natively.
 *  - No `type` / `parser` / `$` fields appear in the upstream quote-props tests.
 *
 * No Babel/Flow cases and no external-fixture (`readFileSync`) cases exist in the
 * upstream quote-props tests. The `._css_` / `._json_` / `._markdown_` test files
 * don't exist for this rule. Each file has exactly one `run()` block (no trailing
 * skipBabel block). The rule has no suggestions.
 *
 * KNOWN GAPS: none. Every fixture (BigInt key `1n`, numeric separators `1_0` /
 * `0b1_000` / `1_2.3_4e0_2`, trailing-dot `5.`, hex/exp `0x123` / `1e2`, import /
 * export `with { ... }` attributes) was verified through the rslint CLI to parse
 * cleanly under ts-go and to produce the upstream-expected diagnostics + fixes, so
 * nothing is isolated. The `quote-props` rule key/value reads come from the AST
 * (espree-tokenized), independent of sloppy/strict mode, so no sourceType gap.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('quote-props', null as never, {
  valid: [
    // ---- from quote-props._js_.test.ts ----
    '({ \'0\': 0 })',
    '({ \'a\': 0 })',
    '({ "a": 0 })',
    '({ \'null\': 0 })',
    '({ \'true\': 0 })',
    '({ \'a-b\': 0 })',
    '({ \'if\': 0 })',
    '({ \'@\': 0 })',

    { code: '({ \'a\': 0, b(){} })' },
    { code: '({ [x]: 0 });' },
    { code: '({ x });' },
    { code: '({ a: 0, b(){} })', options: ['as-needed'] },
    { code: '({ a: 0, [x]: 1 })', options: ['as-needed'] },
    { code: '({ a: 0, x })', options: ['as-needed'] },
    { code: '({ \'@\': 0, [x]: 1 })', options: ['as-needed'] },
    { code: '({ \'@\': 0, x })', options: ['as-needed'] },
    { code: '({ a: 0, b: 0 })', options: ['as-needed'] },
    { code: '({ a: 0, 0: 0 })', options: ['as-needed'] },
    { code: '({ a: 0, true: 0 })', options: ['as-needed'] },
    { code: '({ a: 0, null: 0 })', options: ['as-needed'] },
    { code: '({ a: 0, if: 0 })', options: ['as-needed'] },
    { code: '({ a: 0, while: 0 })', options: ['as-needed'] },
    { code: '({ a: 0, volatile: 0 })', options: ['as-needed'] },
    { code: '({ a: 0, \'-b\': 0 })', options: ['as-needed'] },
    { code: '({ a: 0, \'@\': 0 })', options: ['as-needed'] },
    { code: '({ a: 0, \'0x0\': 0 })', options: ['as-needed'] },
    { code: '({ \' 0\': 0, \'0x0\': 0 })', options: ['as-needed'] },
    { code: '({ \'0 \': 0 })', options: ['as-needed'] },
    { code: '({ \'hey//meh\': 0 })', options: ['as-needed'] },
    { code: '({ \'hey/*meh\': 0 })', options: ['as-needed'] },
    { code: '({ \'hey/*meh*/\': 0 })', options: ['as-needed'] },
    { code: '({ \'a\': 0, \'-b\': 0 })', options: ['consistent'] },
    { code: '({ \'true\': 0, \'b\': 0 })', options: ['consistent'] },
    { code: '({ null: 0, a: 0 })', options: ['consistent'] },
    { code: '({ a: 0, b: 0 })', options: ['consistent'] },
    { code: '({ \'a\': 1, [x]: 0 });', options: ['consistent'] },
    { code: '({ \'a\': 1, x });', options: ['consistent'] },
    { code: '({ a: 0, b: 0 })', options: ['consistent-as-needed'] },
    { code: '({ a: 0, null: 0 })', options: ['consistent-as-needed'] },
    { code: '({ \'a\': 0, \'-b\': 0 })', options: ['consistent-as-needed'] },
    { code: '({ \'@\': 0, \'B\': 0 })', options: ['consistent-as-needed'] },
    { code: '({ \'while\': 0, \'B\': 0 })', options: ['consistent-as-needed', { keywords: true }] },
    { code: '({ \'@\': 0, \'B\': 0 })', options: ['consistent-as-needed', { keywords: true }] },
    { code: '({ \'@\': 1, [x]: 0 });', options: ['consistent-as-needed'] },
    { code: '({ \'@\': 1, x });', options: ['consistent-as-needed'] },
    { code: '({ a: 1, [x]: 0 });', options: ['consistent-as-needed'] },
    { code: '({ a: 1, x });', options: ['consistent-as-needed'] },
    { code: '({ a: 0, \'if\': 0 })', options: ['as-needed', { keywords: true }] },
    { code: '({ a: 0, \'while\': 0 })', options: ['as-needed', { keywords: true }] },
    { code: '({ a: 0, \'volatile\': 0 })', options: ['as-needed', { keywords: true }] },
    { code: '({\'unnecessary\': 1, \'if\': 0})', options: ['as-needed', { keywords: true, unnecessary: false }] },
    { code: '({\'1\': 1})', options: ['as-needed', { numbers: true }] },
    { code: '({1: 1, x: 2})', options: ['consistent', { numbers: true }] },
    { code: '({1: 1, x: 2})', options: ['consistent-as-needed', { numbers: true }] },
    { code: '({ ...x })', options: ['as-needed'] },
    { code: '({ ...x })', options: ['consistent'] },
    { code: '({ ...x })', options: ['consistent-as-needed'] },
    { code: '({ 1n: 1 })', options: ['as-needed'] },
    { code: '({ 1n: 1 })', options: ['as-needed', { numbers: false }] },
    { code: '({ 1n: 1 })', options: ['consistent'] },
    { code: '({ 1n: 1 })', options: ['consistent-as-needed'] },
    { code: '({ \'99999999999999999\': 1 })', options: ['as-needed'] },
    { code: '({ \'1n\': 1 })', options: ['as-needed'] },
    { code: '({ 1_0: 1 })', options: ['as-needed'] },
    { code: '({ 1_0: 1 })', options: ['as-needed', { numbers: false }] },
    { code: '({ \'1_0\': 1 })', options: ['as-needed'] },
    { code: '({ \'1_0\': 1 })', options: ['as-needed', { numbers: false }] },
    { code: '({ \'1_0\': 1 })', options: ['as-needed', { numbers: true }] },
    { code: '({ 1_0: 1, 1: 1 })', options: ['consistent-as-needed'] },
    'import "./foo" with { "a": "foo" }',
    { code: 'import "./foo" with { "a": "foo" }', options: ['always'] },
    { code: 'import "./foo" with { a: "foo" }', options: ['as-needed'] },
    { code: 'import "./foo" with { "a": "foo" }', options: ['consistent'] },
    { code: 'import "./foo" with { a: "foo" }', options: ['consistent-as-needed'] },
    'import "./foo" with { ":": "foo" }',
    { code: 'import "./foo" with { ":": "foo" }', options: ['always'] },
    { code: 'import "./foo" with { ":": "foo" }', options: ['as-needed'] },
    { code: 'import "./foo" with { ":": "foo" }', options: ['consistent'] },
    { code: 'import "./foo" with { ":": "foo" }', options: ['consistent-as-needed'] },
    'import "./foo" with { "a": "foo", "b": "foo", "c": "foo" }',
    { code: 'import "./foo" with { "a": "foo", "b": "foo", "c": "foo" }', options: ['always'] },
    { code: 'import "./foo" with { a: "foo", b: "foo", c: "foo" }', options: ['as-needed'] },
    { code: 'import "./foo" with { "a": "foo", "b": "foo", "c": "foo" }', options: ['consistent'] },
    { code: 'import "./foo" with { a: "foo", b: "foo", c: "foo" }', options: ['consistent-as-needed'] },
    'import "./foo" with { "a": "foo", ":": "foo", "c": "foo" }',
    { code: 'import "./foo" with { "a": "foo", ":": "foo", "c": "foo" }', options: ['always'] },
    { code: 'import "./foo" with { a: "foo", ":": "foo", c: "foo" }', options: ['as-needed'] },
    { code: 'import "./foo" with { "a": "foo", ":": "foo", "c": "foo" }', options: ['consistent'] },
    { code: 'import "./foo" with { "a": "foo", ":": "foo", "c": "foo" }', options: ['consistent-as-needed'] },
    'export {foo} from "./foo" with { "a": "foo" }',
    'export * from "./foo" with { "a": "foo" }',

    // ---- from quote-props._ts_.test.ts ----
    'type x = { "a": 1, b(): void, "c"(): void }',
    'interface x { "a": 1, b(): void, "c"(): void }',
    'enum x { "a" }',

    { code: 'type x = { a: 1, "b-b": 1 }', options: ['as-needed'] },
    { code: 'interface x { a: 1, "b-b": 1 }', options: ['as-needed'] },
    { code: 'enum x { a = 1, "b-b" = 2 }', options: ['as-needed'] },

    { code: 'type x = { "a": 1, "b-b": 1 }', options: ['consistent-as-needed'] },
    { code: 'interface x { "a": 1, "b-b": 1 }', options: ['consistent-as-needed'] },
    { code: 'enum x { "a" = 1, "b-b" = 2 }', options: ['consistent-as-needed'] },
  ],
  invalid: [
    // ---- from quote-props._js_.test.ts ----
    {
      code: '({ a: 0 })',
      output: '({ "a": 0 })',
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: 'a' },
      }],
    }, {
      code: '({ 0: \'0\' })',
      output: '({ "0": \'0\' })',
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: '0' },
      }],
    }, {
      code: '({ \'a\': 0 })',
      output: '({ a: 0 })',
      options: ['as-needed'],
      errors: [{
        messageId: 'unnecessarilyQuotedProperty',
        data: { property: 'a' },
      }],
    }, {
      code: '({ \'null\': 0 })',
      output: '({ null: 0 })',
      options: ['as-needed'],
      errors: [{
        messageId: 'unnecessarilyQuotedProperty',
        data: { property: 'null' },
      }],
    }, {
      code: '({ \'true\': 0 })',
      output: '({ true: 0 })',
      options: ['as-needed'],
      errors: [{
        messageId: 'unnecessarilyQuotedProperty',
        data: { property: 'true' },
      }],
    }, {
      code: '({ \'0\': 0 })',
      output: '({ 0: 0 })',
      options: ['as-needed'],
      errors: [{
        messageId: 'unnecessarilyQuotedProperty',
        data: { property: '0' },
      }],
    }, {
      code: '({ \'-a\': 0, b: 0 })',
      output: '({ \'-a\': 0, "b": 0 })',
      options: ['consistent'],
      errors: [{
        messageId: 'inconsistentlyQuotedProperty',
        data: { key: 'b' },
      }],
    }, {
      code: '({ a: 0, \'b\': 0 })',
      output: '({ "a": 0, \'b\': 0 })',
      options: ['consistent'],
      errors: [{
        messageId: 'inconsistentlyQuotedProperty',
        data: { key: 'a' },
      }],
    }, {
      code: '({ \'-a\': 0, b: 0 })',
      output: '({ \'-a\': 0, "b": 0 })',
      options: ['consistent-as-needed'],
      errors: [{
        messageId: 'inconsistentlyQuotedProperty',
        data: { key: 'b' },
      }],
    }, {
      code: '({ \'a\': 0, \'b\': 0 })',
      output: '({ a: 0, b: 0 })',
      options: ['consistent-as-needed'],
      errors: [
        { messageId: 'redundantQuoting' },
        { messageId: 'redundantQuoting' },
      ],
    }, {
      code: '({ \'a\': 0, [x]: 0 })',
      output: '({ a: 0, [x]: 0 })',
      options: ['consistent-as-needed'],
      errors: [
        { messageId: 'redundantQuoting' },
      ],
    }, {
      code: '({ \'a\': 0, x })',
      output: '({ a: 0, x })',
      options: ['consistent-as-needed'],
      errors: [{
        messageId: 'redundantQuoting',
      }],
    }, {
      code: '({ \'true\': 0, \'null\': 0 })',
      output: '({ true: 0, null: 0 })',
      options: ['consistent-as-needed'],
      errors: [
        { messageId: 'redundantQuoting' },
        { messageId: 'redundantQuoting' },
      ],
    }, {
      code: '({ true: 0, \'null\': 0 })',
      output: '({ "true": 0, \'null\': 0 })',
      options: ['consistent'],
      errors: [{
        messageId: 'inconsistentlyQuotedProperty',
        data: { key: 'true' },
      }],
    }, {
      code: '({ \'a\': 0, \'b\': 0 })',
      output: '({ a: 0, b: 0 })',
      options: ['consistent-as-needed', { keywords: true }],
      errors: [
        { messageId: 'redundantQuoting' },
        { messageId: 'redundantQuoting' },
      ],
    }, {
      code: '({ while: 0, b: 0 })',
      output: '({ "while": 0, "b": 0 })',
      options: ['consistent-as-needed', { keywords: true }],
      errors: [
        {
          messageId: 'requireQuotesDueToReservedWord',
          data: { property: 'while' },
        },
        {
          messageId: 'requireQuotesDueToReservedWord',
          data: { property: 'while' },
        },
      ],
    }, {
      code: '({ while: 0, \'b\': 0 })',
      output: '({ "while": 0, \'b\': 0 })',
      options: ['consistent-as-needed', { keywords: true }],
      errors: [{
        messageId: 'requireQuotesDueToReservedWord',
        data: { property: 'while' },

      }],
    }, {
      code: '({ foo: 0, \'bar\': 0 })',
      output: '({ foo: 0, bar: 0 })',
      options: ['consistent-as-needed', { keywords: true }],
      errors: [
        { messageId: 'redundantQuoting' },
      ],
    }, {
      code:
          '({\n'
          + '  /* a */ \'prop1\' /* b */ : /* c */ value1 /* d */ ,\n'
          + '  /* e */ prop2 /* f */ : /* g */ value2 /* h */,\n'
          + '  /* i */ "prop3" /* j */ : /* k */ value3 /* l */\n'
          + '})',
      output:
          '({\n'
          + '  /* a */ \'prop1\' /* b */ : /* c */ value1 /* d */ ,\n'
          + '  /* e */ "prop2" /* f */ : /* g */ value2 /* h */,\n'
          + '  /* i */ "prop3" /* j */ : /* k */ value3 /* l */\n'
          + '})',
      options: ['consistent'],
      errors: [{
        messageId: 'inconsistentlyQuotedProperty',
        data: { key: 'prop2' },
      }],
    }, {
      code:
          '({\n'
          + '  /* a */ "foo" /* b */ : /* c */ value1 /* d */ ,\n'
          + '  /* e */ "bar" /* f */ : /* g */ value2 /* h */,\n'
          + '  /* i */ "baz" /* j */ : /* k */ value3 /* l */\n'
          + '})',
      output:
          '({\n'
          + '  /* a */ foo /* b */ : /* c */ value1 /* d */ ,\n'
          + '  /* e */ bar /* f */ : /* g */ value2 /* h */,\n'
          + '  /* i */ baz /* j */ : /* k */ value3 /* l */\n'
          + '})',
      options: ['consistent-as-needed'],
      errors: [
        { messageId: 'redundantQuoting' },
        { messageId: 'redundantQuoting' },
        { messageId: 'redundantQuoting' },
      ],
    }, {
      code: '({\'if\': 0})',
      output: '({if: 0})',
      options: ['as-needed'],
      errors: [{
        messageId: 'unnecessarilyQuotedProperty',
        data: { property: 'if' },
      }],
    }, {
      code: '({\'synchronized\': 0})',
      output: '({synchronized: 0})',
      options: ['as-needed'],
      errors: [{
        messageId: 'unnecessarilyQuotedProperty',
        data: { property: 'synchronized' },
      }],
    }, {
      code: '({while: 0})',
      output: '({"while": 0})',
      options: ['as-needed', { keywords: true }],
      errors: [{
        messageId: 'unquotedReservedProperty',
        data: { property: 'while' },
      }],
    }, {
      code: '({\'unnecessary\': 1, if: 0})',
      output: '({\'unnecessary\': 1, "if": 0})',
      options: ['as-needed', { keywords: true, unnecessary: false }],
      errors: [{
        messageId: 'unquotedReservedProperty',
        data: { property: 'if' },
      }],
    }, {
      code: '({1: 1})',
      output: '({"1": 1})',
      options: ['as-needed', { numbers: true }],
      errors: [{
        messageId: 'unquotedNumericProperty',
        data: { property: '1' },
      }],
    }, {
      code: '({1: 1})',
      output: '({"1": 1})',
      options: ['always', { numbers: false }],
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: '1' },
      }],
    }, {
      code: '({0x123: 1})',
      output: '({"291": 1})', // 0x123 === 291
      options: ['always'],
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: '291' },

      }],
    }, {
      code: '({1e2: 1})',
      output: '({"100": 1})',
      options: ['always', { numbers: false }],
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: '100' },

      }],
    }, {
      code: '({5.: 1})',
      output: '({"5": 1})',
      options: ['always', { numbers: false }],
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: '5' },

      }],
    }, {
      code: '({ 1n: 1 })',
      output: '({ "1": 1 })',
      options: ['always'],
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: '1' },
      }],
    }, {
      code: '({ 1n: 1 })',
      output: '({ "1": 1 })',
      options: ['as-needed', { numbers: true }],
      errors: [{
        messageId: 'unquotedNumericProperty',
        data: { property: '1' },
      }],
    }, {
      code: '({ 1_0: 1 })',
      output: '({ "10": 1 })',
      options: ['as-needed', { numbers: true }],
      errors: [{
        messageId: 'unquotedNumericProperty',
        data: { property: '10' },
      }],
    }, {
      code: '({ 1_2.3_4e0_2: 1 })',
      output: '({ "1234": 1 })',
      options: ['always'],
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: '1234' },
      }],
    }, {
      code: '({ 0b1_000: 1 })',
      output: '({ "8": 1 })',
      options: ['always'],
      errors: [{
        messageId: 'unquotedPropertyFound',
        data: { property: '8' },
      }],
    }, {
      code: '({ 1_000: a, \'1_000\': b })',
      output: '({ "1000": a, \'1_000\': b })',
      options: ['consistent-as-needed'],
      errors: [{
        messageId: 'inconsistentlyQuotedProperty',
        data: { key: '1000' },
      }],
    }, {
      code: 'import "./foo" with { a: "foo" }',
      output: 'import "./foo" with { "a": "foo" }',
      errors: [{ messageId: 'unquotedPropertyFound', data: { property: 'a' } }],
    }, {
      code: 'import "./foo" with { a: "foo", "b": "foo", "c": "foo" }',
      output: 'import "./foo" with { "a": "foo", "b": "foo", "c": "foo" }',
      errors: [{ messageId: 'unquotedPropertyFound', data: { property: 'a' } }],
    }, {
      code: 'import "./foo" with { a: "foo", ":": "foo", "c": "foo" }',
      output: 'import "./foo" with { "a": "foo", ":": "foo", "c": "foo" }',
      errors: [{ messageId: 'unquotedPropertyFound', data: { property: 'a' } }],
    }, {
      code: 'import "./foo" with { a: "foo" }',
      output: 'import "./foo" with { "a": "foo" }',
      options: ['always'],
      errors: [{ messageId: 'unquotedPropertyFound', data: { property: 'a' } }],
    }, {
      code: 'import "./foo" with { a: "foo", "b": "foo", "c": "foo" }',
      output: 'import "./foo" with { "a": "foo", "b": "foo", "c": "foo" }',
      options: ['always'],
      errors: [{ messageId: 'unquotedPropertyFound', data: { property: 'a' } }],
    }, {
      code: 'import "./foo" with { a: "foo", ":": "foo", "c": "foo" }',
      output: 'import "./foo" with { "a": "foo", ":": "foo", "c": "foo" }',
      options: ['always'],
      errors: [{ messageId: 'unquotedPropertyFound', data: { property: 'a' } }],
    }, {
      code: 'import "./foo" with { "a": "foo" }',
      output: 'import "./foo" with { a: "foo" }',
      options: ['as-needed'],
      errors: [{ messageId: 'unnecessarilyQuotedProperty', data: { property: 'a' } }],
    }, {
      code: 'import "./foo" with { a: "foo", "b": "foo", "c": "foo" }',
      output: 'import "./foo" with { a: "foo", b: "foo", c: "foo" }',
      options: ['as-needed'],
      errors: [
        { messageId: 'unnecessarilyQuotedProperty', data: { property: 'b' } },
        { messageId: 'unnecessarilyQuotedProperty', data: { property: 'c' } },
      ],
    }, {
      code: 'import "./foo" with { a: "foo", ":": "foo", "c": "foo" }',
      output: 'import "./foo" with { a: "foo", ":": "foo", c: "foo" }',
      options: ['as-needed'],
      errors: [{ messageId: 'unnecessarilyQuotedProperty', data: { property: 'c' } }],
    }, {
      code: 'import "./foo" with { a: "foo", "b": "foo", "c": "foo" }',
      output: 'import "./foo" with { "a": "foo", "b": "foo", "c": "foo" }',
      options: ['consistent'],
      errors: [{ messageId: 'inconsistentlyQuotedProperty', data: { key: 'a' } }],
    }, {
      code: 'import "./foo" with { a: "foo", ":": "foo", "c": "foo" }',
      output: 'import "./foo" with { "a": "foo", ":": "foo", "c": "foo" }',
      options: ['consistent'],
      errors: [{ messageId: 'inconsistentlyQuotedProperty', data: { key: 'a' } }],
    }, {
      code: 'import "./foo" with { "a": "foo" }',
      output: 'import "./foo" with { a: "foo" }',
      options: ['consistent-as-needed'],
      errors: [{ messageId: 'redundantQuoting' }],
    }, {
      code: 'import "./foo" with { a: "foo", "b": "foo", "c": "foo" }',
      output: 'import "./foo" with { a: "foo", b: "foo", c: "foo" }',
      options: ['consistent-as-needed'],
      errors: [
        { messageId: 'redundantQuoting', column: 33 },
        { messageId: 'redundantQuoting', column: 45 },
      ],
    }, {
      code: 'import "./foo" with { a: "foo", ":": "foo", c: "foo" }',
      output: 'import "./foo" with { "a": "foo", ":": "foo", "c": "foo" }',
      options: ['consistent-as-needed'],
      errors: [
        { messageId: 'inconsistentlyQuotedProperty', data: { key: 'a' } },
        { messageId: 'inconsistentlyQuotedProperty', data: { key: 'c' } },
      ],
    }, {
      code: 'export {foo} from "./foo" with { a: "foo" }',
      output: 'export {foo} from "./foo" with { "a": "foo" }',
      errors: [
        { messageId: 'unquotedPropertyFound', data: { property: 'a' } },
      ],
    }, {
      code: 'export * from "./foo" with { a: "foo" }',
      output: 'export * from "./foo" with { "a": "foo" }',
      errors: [
        { messageId: 'unquotedPropertyFound', data: { property: 'a' } },
      ],
    },

    // ---- from quote-props._ts_.test.ts ----
    {
      code: 'type x = { a: 1 }',
      output: 'type x = { "a": 1 }',
      errors: [{ messageId: 'unquotedPropertyFound' }],
    },
    {
      code: 'interface x { a: 1 }',
      output: 'interface x { "a": 1 }',
      errors: [{ messageId: 'unquotedPropertyFound' }],
    },
    {
      code: 'enum x { a = 1 }',
      output: 'enum x { "a" = 1 }',
      errors: [{ messageId: 'unquotedPropertyFound' }],
    },

    {
      code: 'type x = { "a": 1 }',
      output: 'type x = { a: 1 }',
      options: ['as-needed'],
      errors: [{ messageId: 'unnecessarilyQuotedProperty' }],
    },
    {
      code: 'interface x { "a": 1, "b-b": 1 }',
      output: 'interface x { a: 1, "b-b": 1 }',
      options: ['as-needed'],
      errors: [{ messageId: 'unnecessarilyQuotedProperty' }],
    },
    {
      code: 'enum x { "a" = 1, "b-b" = 2 }',
      output: 'enum x { a = 1, "b-b" = 2 }',
      options: ['as-needed'],
      errors: [{ messageId: 'unnecessarilyQuotedProperty' }],
    },

    {
      code: 'type x = { a: 1, "b-b": 1 }',
      output: 'type x = { "a": 1, "b-b": 1 }',
      options: ['consistent-as-needed'],
      errors: [{ messageId: 'inconsistentlyQuotedProperty' }],
    },
    {
      code: 'interface x { a: 1, "b-b": 1 }',
      output: 'interface x { "a": 1, "b-b": 1 }',
      options: ['consistent-as-needed'],
      errors: [{ messageId: 'inconsistentlyQuotedProperty' }],
    },
    {
      code: 'enum x { a = 1, "b-b" = 2 }',
      output: 'enum x { "a" = 1, "b-b" = 2 }',
      options: ['consistent-as-needed'],
      errors: [{ messageId: 'inconsistentlyQuotedProperty' }],
    },
  ],
});
