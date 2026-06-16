/**
 * @fileoverview Tests for dot-location rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/dot-location/dot-location._js_.test.ts
 *   packages/eslint-plugin/rules/dot-location/dot-location._jsx_.test.ts
 *   packages/eslint-plugin/rules/dot-location/dot-location._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('dot-location', null as never, { valid, invalid })`
 *  - All fixtures are plain single-/double-quoted strings or backtick templates
 *    that are NOT the `$` unindent tag, so they are kept byte-for-byte.
 *  - `parserOptions` (ecmaFeatures.jsx / sourceType / ecmaVersion) dropped — rslint
 *    resolves via tsconfig; the RuleTester picks a `.tsx` fixture when JSX is present.
 *  - `type` fields (deprecated AST node type) dropped. No `data` is attached to any
 *    error: both messageIds (`expectedDotAfterObject` / `expectedDotBeforeProperty`)
 *    render static text with no `{{}}` placeholders.
 *
 * No Babel/Flow cases, no `suggestions`, and no external-fixture (`readFileSync`)
 * cases exist in the upstream dot-location tests. The `._css_` / `._json_` /
 * `._markdown_` / `._unknown_` test files don't exist (or are non-JS/TS) for this
 * rule and are excluded.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are moved to the `dot-location — KNOWN GAPS` block comment at the bottom, each
 * annotated with what upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('dot-location', null as never, {
  valid: [
    // ---- from dot-location._js_.test.ts ----
    'obj.prop',
    'obj.\nprop',
    'obj. \nprop',
    'obj.\n prop',
    '(obj).\nprop',
    'obj\n[\'prop\']',
    'obj[\'prop\']',
    {
      code: 'obj.\nprop',
      options: ['object'],
    },
    {
      code: 'obj\n.prop',
      options: ['property'],
    },
    {
      code: '(obj)\n.prop',
      options: ['property'],
    },
    {
      code: 'obj . prop',
      options: ['object'],
    },
    {
      code: 'obj /* a */ . prop',
      options: ['object'],
    },
    {
      code: 'obj . \nprop',
      options: ['object'],
    },
    {
      code: 'obj . prop',
      options: ['property'],
    },
    {
      code: 'obj . /* a */ prop',
      options: ['property'],
    },
    {
      code: 'obj\n. prop',
      options: ['property'],
    },
    {
      code: 'f(a\n).prop',
      options: ['object'],
    },
    {
      code: '`\n`.prop',
      options: ['object'],
    },
    {
      code: 'obj[prop]',
      options: ['object'],
    },
    {
      code: 'obj\n[prop]',
      options: ['object'],
    },
    {
      code: 'obj[\nprop]',
      options: ['object'],
    },
    {
      code: 'obj\n[\nprop\n]',
      options: ['object'],
    },
    {
      code: 'obj[prop]',
      options: ['property'],
    },
    {
      code: 'obj\n[prop]',
      options: ['property'],
    },
    {
      code: 'obj[\nprop]',
      options: ['property'],
    },
    {
      code: 'obj\n[\nprop\n]',
      options: ['property'],
    },

    // https://github.com/eslint/eslint/issues/11868 (also in invalid)
    {
      code: '(obj).prop',
      options: ['object'],
    },
    {
      code: '(obj).\nprop',
      options: ['object'],
    },
    {
      code: '(obj\n).\nprop',
      options: ['object'],
    },
    {
      code: '(\nobj\n).\nprop',
      options: ['object'],
    },
    {
      code: '((obj\n)).\nprop',
      options: ['object'],
    },
    {
      code: '(f(a)\n).\nprop',
      options: ['object'],
    },
    {
      code: '((obj\n)\n).\nprop',
      options: ['object'],
    },
    {
      code: '(\na &&\nb()\n).toString()',
      options: ['object'],
    },

    // Optional chaining
    {
      code: 'obj?.prop',
      options: ['object'],
    },
    {
      code: 'obj?.[key]',
      options: ['object'],
    },
    {
      code: 'obj?.\nprop',
      options: ['object'],
    },
    {
      code: 'obj\n?.[key]',
      options: ['object'],
    },
    {
      code: 'obj?.\n[key]',
      options: ['object'],
    },
    {
      code: 'obj?.[\nkey]',
      options: ['object'],
    },
    {
      code: 'obj?.prop',
      options: ['property'],
    },
    {
      code: 'obj?.[key]',
      options: ['property'],
    },
    {
      code: 'obj\n?.prop',
      options: ['property'],
    },
    {
      code: 'obj\n?.[key]',
      options: ['property'],
    },
    {
      code: 'obj?.\n[key]',
      options: ['property'],
    },
    {
      code: 'obj?.[\nkey]',
      options: ['property'],
    },

    // Private properties
    {
      code: 'class C { #a; foo() { this.\n#a; } }',
      options: ['object'],
    },
    {
      code: 'class C { #a; foo() { this\n.#a; } }',
      options: ['property'],
    },

    // MetaProperty
    'import.meta',

    // ---- from dot-location._jsx_.test.ts ----
    {
      code: '<Form.\nInput />',
      options: ['object'],
    },
    {
      code: '<Form\n.Input />',
      options: ['property'],
    },

    // ---- from dot-location._ts_.test.ts ----
    // TSImportType
    'type Foo = import(\'foo\')',
    {
      code: 'type Foo = import(\'foo\').\nProp',
      options: ['object'],
    },
    {
      code: 'type Foo = import(\'foo\')\n.Prop',
      options: ['property'],
    },

    // TSQualifiedName
    {
      code: 'type Foo = Obj.\nProp',
      options: ['object'],
    },
    {
      code: 'type Foo = Obj\n.Prop',
      options: ['property'],
    },
  ],

  invalid: [
    // ---- from dot-location._js_.test.ts ----
    {
      code: 'obj\n.property',
      output: 'obj.\nproperty',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1, endLine: 2, endColumn: 2 }],
    },
    {
      code: 'obj.\nproperty',
      output: 'obj\n.property',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty', line: 1, column: 4, endLine: 1, endColumn: 5 }],
    },
    {
      code: '(obj).\nproperty',
      output: '(obj)\n.property',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty', line: 1, column: 6 }],
    },
    {
      code: '5\n.toExponential()',
      output: '5 .\ntoExponential()',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }],
    },
    {
      code: '-5\n.toExponential()',
      output: '-5 .\ntoExponential()',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }],
    },
    // NOTE: the three legacy-numeric-literal cases `01`, `08`, `0190` are in KNOWN
    // GAPS below — upstream runs them with `parserOptions.sourceType: 'script'`
    // (sloppy mode); under rslint's ts-go parser (strict/module ES) a legacy octal
    // literal `01` is TS1121 and `08`/`0190` are TS1489, so the fixture is a SYNTAX
    // ERROR and produces 0 dot-location diagnostics.
    {
      code: '5_000\n.toExponential()',
      output: '5_000 .\ntoExponential()',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }],
    },
    {
      code: '5_000_00\n.toExponential()',
      output: '5_000_00 .\ntoExponential()',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }],
    },
    {
      code: '5.000_000\n.toExponential()',
      output: '5.000_000.\ntoExponential()',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }],
    },
    {
      code: '0b1010_1010\n.toExponential()',
      output: '0b1010_1010.\ntoExponential()',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }],
    },
    {
      code: 'foo /* a */ . /* b */ \n /* c */ bar',
      output: 'foo /* a */  /* b */ \n /* c */ .bar',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty', line: 1, column: 13 }],
    },
    {
      code: 'foo /* a */ \n /* b */ . /* c */ bar',
      output: 'foo. /* a */ \n /* b */  /* c */ bar',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 10 }],
    },
    {
      code: 'f(a\n)\n.prop',
      output: 'f(a\n).\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 3, column: 1 }],
    },
    {
      code: '`\n`\n.prop',
      output: '`\n`.\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 3, column: 1 }],
    },

    // https://github.com/eslint/eslint/issues/11868 (also in valid)
    {
      code: '(a\n)\n.prop',
      output: '(a\n).\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 3, column: 1 }],
    },
    {
      code: '(a\n)\n.\nprop',
      output: '(a\n).\n\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 3, column: 1 }],
    },
    {
      code: '(f(a)\n)\n.prop',
      output: '(f(a)\n).\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 3, column: 1 }],
    },
    {
      code: '(f(a\n)\n)\n.prop',
      output: '(f(a\n)\n).\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 4, column: 1 }],
    },
    {
      code: '((obj\n))\n.prop',
      output: '((obj\n)).\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 3, column: 1 }],
    },
    {
      code: '((obj\n)\n)\n.prop',
      output: '((obj\n)\n).\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 4, column: 1 }],
    },
    {
      code: '(a\n) /* a */ \n.prop',
      output: '(a\n). /* a */ \nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 3, column: 1 }],
    },
    {
      code: '(a\n)\n/* a */\n.prop',
      output: '(a\n).\n/* a */\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 4, column: 1 }],
    },
    {
      code: '(a\n)\n/* a */.prop',
      output: '(a\n).\n/* a */prop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 3, column: 8 }],
    },
    {
      code: '(5)\n.toExponential()',
      output: '(5).\ntoExponential()',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }],
    },

    // Optional chaining
    {
      code: 'obj\n?.prop',
      output: 'obj?.\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: '10\n?.prop',
      output: '10?.\nprop',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: 'obj?.\nprop',
      output: 'obj\n?.prop',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },

    // Private properties
    {
      code: 'class C { #a; foo() { this\n.#a; } }',
      output: 'class C { #a; foo() { this.\n#a; } }',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: 'class C { #a; foo() { this.\n#a; } }',
      output: 'class C { #a; foo() { this\n.#a; } }',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },

    // ---- from dot-location._jsx_.test.ts ----
    // Upstream pins ONLY code+output here (eslint-vitest-rule-tester tolerates an
    // errors-less invalid case); kept output-only — RuleTester verifies the fix
    // plus >=1 diagnostic, no invented positions.
    {
      code: '<Form\n.Input />',
      output: '<Form.\nInput />',
      options: ['object'],
    },
    {
      code: '<Form.\nInput />',
      output: '<Form\n.Input />',
      options: ['property'],
    },

    // ---- from dot-location._ts_.test.ts ----
    // TSImportType
    {
      code: 'type Foo = import(\'foo\')\n.Prop',
      output: 'type Foo = import(\'foo\').\nProp',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: 'type Foo = import(\'foo\').\nProp',
      output: 'type Foo = import(\'foo\')\n.Prop',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },

    // TSQualifiedName
    {
      code: 'type Foo = Obj\n.Prop',
      output: 'type Foo = Obj.\nProp',
      options: ['object'],
      errors: [{ messageId: 'expectedDotAfterObject' }],
    },
    {
      code: 'type Foo = Obj.\nProp',
      output: 'type Foo = Obj\n.Prop',
      options: ['property'],
      errors: [{ messageId: 'expectedDotBeforeProperty' }],
    },
  ],
});

/**
 * ============================ dot-location — KNOWN GAPS ============================
 *
 * The cases below are ported verbatim from upstream but are NOT run through the
 * green `ruleTester.run` above: each is a *parser-level* incompatibility. rslint
 * parses every fixture with the ts-go parser, which enforces strict / module ES
 * semantics for `.ts` files. Upstream pins `parserOptions.sourceType: 'script'`
 * (sloppy mode) on these, where a legacy/non-decimal-prefixed numeric literal is
 * legal. Under ts-go the source itself is a SYNTAX ERROR, so rslint emits a
 * TypeScript diagnostic and produces ZERO `@stylistic/dot-location` diagnostics
 * for that file (and, because the rslint CLI aborts JSONL for the whole batch on a
 * syntax error, such a fixture would zero out every other case in the same run —
 * which is exactly why they must live outside the green set). The rule logic is
 * not at fault; the input is unparseable. Expected upstream behaviour preserved:
 *
 * ---- invalid (upstream expects 1 `expectedDotAfterObject` diagnostic + the fix) ----
 *
 *   { code: `01\n.toExponential()`,   output: `01.\ntoExponential()`,   options: ['object'], errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }], parserOptions: { sourceType: 'script' } }
 *   { code: `08\n.toExponential()`,   output: `08 .\ntoExponential()`,  options: ['object'], errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }], parserOptions: { sourceType: 'script' } }
 *   { code: `0190\n.toExponential()`, output: `0190 .\ntoExponential()`, options: ['object'], errors: [{ messageId: 'expectedDotAfterObject', line: 2, column: 1 }], parserOptions: { sourceType: 'script' } }
 *
 *   rslint: TypeScript(TS1121) "Octal literals are not allowed. Use the syntax
 *   '0o1'." for `01`, and TypeScript(TS1489) "Decimals with leading zeros are not
 *   allowed." for `08` / `0190` -> 0 dot-location diagnostics each.
 *
 * ==================================================================================
 */
