/**
 * @fileoverview Tests for no-whitespace-before-property rule.
 * @author Kai Cataldo
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/no-whitespace-before-property/no-whitespace-before-property._js_.test.ts
 *   packages/eslint-plugin/rules/no-whitespace-before-property/no-whitespace-before-property._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('no-whitespace-before-property', null as never, { valid, invalid })`
 *  - The single `messageId: 'unexpectedWhitespace'` carries `data.propName`; the
 *    RuleTester renders `Unexpected whitespace before property {{propName}}.` from
 *    the plugin's own meta + data.
 *  - `parserOptions` (ecmaVersion 2020/2021 for optional-chaining / numeric
 *    separators, `sourceType`) dropped — rslint parses every fixture at esnext
 *    module semantics via the generated tsconfig.
 *  - No `$` (unindent) cases, no Babel/Flow cases, no output-only invalid cases,
 *    and no suggestions exist for this rule.
 *
 * KNOWN GAPS: 3 invalid cases (see the block comment after `ruleTester.run`).
 * They feed ts-go a leading-zero / legacy-octal numeric literal, which is a
 * TypeScript SYNTAX ERROR under rslint's strict/module ES parser — upstream runs
 * them with `parserOptions.sourceType: 'script'` (sloppy mode) where they parse.
 * The rule logic is not at fault; the source itself is unparseable.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-whitespace-before-property', null as never, {
  valid: [
    // ---- from no-whitespace-before-property._js_.test.ts ----
    "foo.bar",
    "foo.bar()",
    "foo[bar]",
    "foo['bar']",
    "foo[0]",
    "foo[ bar ]",
    "foo[ 'bar' ]",
    "foo[ 0 ]",
    "foo\n.bar",
    "foo.\nbar",
    "foo\n.bar()",
    "foo.\nbar()",
    "foo\n[bar]",
    "foo\n['bar']",
    "foo\n[0]",
    "foo\n[ bar ]",
    "foo.\n bar",
    "foo\n. bar",
    "foo.\n bar()",
    "foo\n. bar()",
    "foo\n [bar]",
    "foo\n ['bar']",
    "foo\n [0]",
    "foo\n [ bar ]",
    "foo.\n\tbar",
    "foo\n.\tbar",
    "foo.\n\tbar()",
    "foo\n.\tbar()",
    "foo\n\t[bar]",
    "foo\n\t['bar']",
    "foo\n\t[0]",
    "foo\n\t[ bar ]",
    "foo.bar.baz",
    "foo\n.bar\n.baz",
    "foo.\nbar.\nbaz",
    "foo.bar().baz()",
    "foo\n.bar()\n.baz()",
    "foo.\nbar().\nbaz()",
    "foo\n.bar\n[baz]",
    "foo\n.bar\n['baz']",
    "foo\n.bar\n[0]",
    "foo\n.bar\n[ baz ]",
    "foo\n .bar\n .baz",
    "foo.\n bar.\n baz",
    "foo\n .bar()\n .baz()",
    "foo.\n bar().\n baz()",
    "foo\n .bar\n [baz]",
    "foo\n .bar\n ['baz']",
    "foo\n .bar\n [0]",
    "foo\n .bar\n [ baz ]",
    "foo\n\t.bar\n\t.baz",
    "foo.\n\tbar.\n\tbaz",
    "foo\n\t.bar()\n\t.baz()",
    "foo.\n\tbar().\n\tbaz()",
    "foo\n\t.bar\n\t[baz]",
    "foo\n\t.bar\n\t['baz']",
    "foo\n\t.bar\n\t[0]",
    "foo\n\t.bar\n\t[ baz ]",
    "foo['bar' + baz]",
    "foo[ 'bar' + baz ]",
    "(foo + bar).baz",
    "( foo + bar ).baz",
    "(foo ? bar : baz).qux",
    "( foo ? bar : baz ).qux",
    "(foo ? bar : baz)[qux]",
    "( foo ? bar : baz )[qux]",
    "( foo ? bar : baz )[0].qux",
    "foo.bar[('baz')]",
    "foo.bar[ ('baz') ]",
    "foo[[bar]]",
    "foo[ [ bar ] ]",
    "foo[['bar']]",
    "foo[ [ 'bar' ] ]",
    "foo[(('baz'))]",
    "foo[ (('baz'))]",
    "foo[0][[('baz')]]",
    "foo[bar.baz('qux')]",
    "foo[(bar.baz() + 0) + qux]",
    "foo['bar ' + 1 + ' baz']",
    "5['toExponential']()",
    {
      code: "obj?.prop",
    },
    {
      code: "( obj )?.prop",
    },
    {
      code: "obj\n  ?.prop",
    },
    {
      code: "obj?.\n  prop",
    },
    {
      code: "obj?.[key]",
    },
    {
      code: "( obj )?.[ key ]",
    },
    {
      code: "obj\n  ?.[key]",
    },
    {
      code: "obj?.\n  [key]",
    },
    {
      code: "obj\n  ?.\n  [key]",
    },

    // ---- from no-whitespace-before-property._ts_.test.ts ----
    "type Foo = import(A)",
    "type Foo = A['B']",
    "type Foo = A.B",
    "type Foo = import(A).B",
    "type Test = ( typeof arr )[ number ];",
  ],

  invalid: [
    // ---- from no-whitespace-before-property._js_.test.ts ----
    {
      code: "foo. bar",
      output: "foo.bar",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo .bar",
      output: "foo.bar",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo [bar]",
      output: "foo[bar]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo [0]",
      output: "foo[0]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "0" } }
      ],
    },
    {
      code: "foo ['bar']",
      output: "foo['bar']",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'bar'" } }
      ],
    },
    {
      code: "foo. bar. baz",
      output: "foo.bar.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo .bar. baz",
      output: "foo.bar.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo [bar] [baz]",
      output: "foo[bar][baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo [bar][baz]",
      output: "foo[bar][baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo[bar] [baz]",
      output: "foo[bar][baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo.bar [baz]",
      output: "foo.bar[baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo. bar[baz]",
      output: "foo.bar[baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo[bar]. baz",
      output: "foo[bar].baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo[ bar ] [ baz ]",
      output: "foo[ bar ][ baz ]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo [ 0 ][ baz ]",
      output: "foo[ 0 ][ baz ]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "0" } }
      ],
    },
    {
      code: "foo[ 0 ] [ 'baz' ]",
      output: "foo[ 0 ][ 'baz' ]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'baz'" } }
      ],
    },
    {
      code: "foo\t.bar",
      output: "foo.bar",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo.\tbar",
      output: "foo.bar",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t.bar()",
      output: "foo.bar()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo.\tbar()",
      output: "foo.bar()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t[bar]",
      output: "foo[bar]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t[0]",
      output: "foo[0]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "0" } }
      ],
    },
    {
      code: "foo\t['bar']",
      output: "foo['bar']",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'bar'" } }
      ],
    },
    {
      code: "foo.\tbar.\tbaz",
      output: "foo.bar.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t.bar.\tbaz",
      output: "foo.bar.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo.\tbar().\tbaz()",
      output: "foo.bar().baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t.bar().\tbaz()",
      output: "foo.bar().baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t[bar]\t[baz]",
      output: "foo[bar][baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t[bar][baz]",
      output: "foo[bar][baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo[bar]\t[baz]",
      output: "foo[bar][baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo.bar\t[baz]",
      output: "foo.bar[baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo.\tbar[baz]",
      output: "foo.bar[baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo[bar].\tbaz",
      output: "foo[bar].baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo [bar]\n .baz",
      output: "foo[bar]\n .baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo. bar\n .baz",
      output: "foo.bar\n .baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo .bar\n.baz",
      output: "foo.bar\n.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo.\n bar. baz",
      output: "foo.\n bar.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo.\nbar . baz",
      output: "foo.\nbar.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo. bar()\n .baz()",
      output: "foo.bar()\n .baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo .bar()\n.baz()",
      output: "foo.bar()\n.baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo.\n bar(). baz()",
      output: "foo.\n bar().baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo.\nbar() . baz()",
      output: "foo.\nbar().baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo\t[bar]\n\t.baz",
      output: "foo[bar]\n\t.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo.\tbar\n\t.baz",
      output: "foo.bar\n\t.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t.bar\n.baz",
      output: "foo.bar\n.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo.\n\tbar.\tbaz",
      output: "foo.\n\tbar.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo.\nbar\t.\tbaz",
      output: "foo.\nbar.baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo.\tbar()\n\t.baz()",
      output: "foo.bar()\n\t.baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo\t.bar()\n.baz()",
      output: "foo.bar()\n.baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo.\n\tbar().\tbaz()",
      output: "foo.\n\tbar().baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo.\nbar()\t.\tbaz()",
      output: "foo.\nbar().baz()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo ['bar' + baz]",
      output: "foo['bar' + baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'bar' + baz" } }
      ],
    },
    {
      code: "(foo + bar) .baz",
      output: "(foo + bar).baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "(foo + bar). baz",
      output: "(foo + bar).baz",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "(foo + bar) [baz]",
      output: "(foo + bar)[baz]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "(foo ? bar : baz) .qux",
      output: "(foo ? bar : baz).qux",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "qux" } }
      ],
    },
    {
      code: "(foo ? bar : baz). qux",
      output: "(foo ? bar : baz).qux",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "qux" } }
      ],
    },
    {
      code: "(foo ? bar : baz) [qux]",
      output: "(foo ? bar : baz)[qux]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "qux" } }
      ],
    },
    {
      code: "( foo ? bar : baz ) [0].qux",
      output: "( foo ? bar : baz )[0].qux",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "0" } }
      ],
    },
    {
      code: "( foo ? bar : baz )[0] .qux",
      output: "( foo ? bar : baz )[0].qux",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "qux" } }
      ],
    },
    {
      code: "( foo ? bar : baz )[0]. qux",
      output: "( foo ? bar : baz )[0].qux",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "qux" } }
      ],
    },
    {
      code: "( foo ? bar : baz ) [0]. qux",
      output: "( foo ? bar : baz )[0].qux",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "qux" } },
        { messageId: "unexpectedWhitespace", data: { propName: "0" } }
      ],
    },
    {
      code: "foo.bar [('baz')]",
      output: "foo.bar[('baz')]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'baz'" } }
      ],
    },
    {
      code: "foo .bar[('baz')]",
      output: "foo.bar[('baz')]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo .bar [('baz')]",
      output: "foo.bar[('baz')]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'baz'" } },
        { messageId: "unexpectedWhitespace", data: { propName: "bar" } }
      ],
    },
    {
      code: "foo [(('baz'))]",
      output: "foo[(('baz'))]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'baz'" } }
      ],
    },
    {
      code: "foo [[baz]]",
      output: "foo[[baz]]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "[baz]" } }
      ],
    },
    {
      code: "foo [ [ baz ] ]",
      output: "foo[ [ baz ] ]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "[ baz ]" } }
      ],
    },
    {
      code: "foo [['baz']]",
      output: "foo[['baz']]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "['baz']" } }
      ],
    },
    {
      code: "foo [ [ 'baz' ] ]",
      output: "foo[ [ 'baz' ] ]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "[ 'baz' ]" } }
      ],
    },
    {
      code: "foo[0] [[('baz')]]",
      output: "foo[0][[('baz')]]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "[('baz')]" } }
      ],
    },
    {
      code: "foo [0][[('baz')]]",
      output: "foo[0][[('baz')]]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "0" } }
      ],
    },
    {
      code: "foo [0] [[('baz')]]",
      output: "foo[0][[('baz')]]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "[('baz')]" } },
        { messageId: "unexpectedWhitespace", data: { propName: "0" } }
      ],
    },
    {
      code: "foo [bar.baz('qux')]",
      output: "foo[bar.baz('qux')]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar.baz('qux')" } }
      ],
    },
    {
      code: "foo[bar .baz('qux')]",
      output: "foo[bar.baz('qux')]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo [bar . baz('qux')]",
      output: "foo[bar.baz('qux')]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "bar . baz('qux')" } },
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo [(bar.baz() + 0) + qux]",
      output: "foo[(bar.baz() + 0) + qux]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "(bar.baz() + 0) + qux" } }
      ],
    },
    {
      code: "foo[(bar. baz() + 0) + qux]",
      output: "foo[(bar.baz() + 0) + qux]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo [(bar. baz() + 0) + qux]",
      output: "foo[(bar.baz() + 0) + qux]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "(bar. baz() + 0) + qux" } },
        { messageId: "unexpectedWhitespace", data: { propName: "baz" } }
      ],
    },
    {
      code: "foo ['bar ' + 1 + ' baz']",
      output: "foo['bar ' + 1 + ' baz']",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'bar ' + 1 + ' baz'" } }
      ],
    },
    {
      code: "5 .toExponential()",
      output: null,
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5       .toExponential()",
      output: null,
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5_000       .toExponential()",
      output: null,
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5_000_00       .toExponential()",
      output: null,
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5. .toExponential()",
      output: "5..toExponential()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5.0 .toExponential()",
      output: "5.0.toExponential()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5.0_0 .toExponential()",
      output: "5.0_0.toExponential()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "0x5 .toExponential()",
      output: "0x5.toExponential()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "0x56_78 .toExponential()",
      output: "0x56_78.toExponential()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5e0 .toExponential()",
      output: "5e0.toExponential()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5e-0 .toExponential()",
      output: "5e-0.toExponential()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "toExponential" } }
      ],
    },
    {
      code: "5 ['toExponential']()",
      output: "5['toExponential']()",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "'toExponential'" } }
      ],
    },
    {
      code: "obj?. prop",
      output: "obj?.prop",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "prop" } }
      ],
    },
    {
      code: "obj ?.prop",
      output: "obj?.prop",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "prop" } }
      ],
    },
    {
      code: "obj?. [key]",
      output: "obj?.[key]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "key" } }
      ],
    },
    {
      code: "obj ?.[key]",
      output: "obj?.[key]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "key" } }
      ],
    },
    {
      code: "5 ?. prop",
      output: "5?.prop",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "prop" } }
      ],
    },
    {
      code: "5 ?. [key]",
      output: "5?.[key]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "key" } }
      ],
    },
    {
      code: "obj/* comment */?. prop",
      output: null,
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "prop" } }
      ],
    },
    {
      code: "obj ?./* comment */prop",
      output: null,
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "prop" } }
      ],
    },

    // ---- from no-whitespace-before-property._ts_.test.ts ----
    {
      code: "type Foo = A [B]",
      output: "type Foo = A[B]",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "B" } }
      ],
    },
    {
      code: "type Foo = A .B",
      output: "type Foo = A.B",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "B" } }
      ],
    },
    {
      code: "type Foo = import(A) .B",
      output: "type Foo = import(A).B",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "B" } }
      ],
    },
    {
      code: "type Test = ( typeof arr ) [ number ];",
      output: "type Test = ( typeof arr )[ number ];",
      errors: [
        { messageId: "unexpectedWhitespace", data: { propName: "number" } }
      ],
    },
  ],
});

/**
 * ==================== no-whitespace-before-property — KNOWN GAPS ====================
 *
 * Ported verbatim from upstream but NOT run through the green `ruleTester.run`
 * above: each feeds ts-go a numeric literal that is a SYNTAX ERROR under rslint's
 * ts-go parser (strict / module ES semantics for `.ts` files). Upstream parses
 * these with `parserOptions.sourceType: 'script'` (sloppy mode), where a leading
 * zero is a legacy octal / non-octal-decimal literal and parses fine. Under ts-go
 * the source aborts with a `TSxxxx` diagnostic and produces ZERO
 * `@stylistic/no-whitespace-before-property` diagnostics — and because the rslint
 * CLI aborts JSONL for the whole batch on any syntax error, such a fixture would
 * zero out every other case in the same run, which is exactly why they live
 * outside the green set. The rule logic is not at fault; the input is unparseable.
 *
 * ---- invalid (upstream expects 1 `unexpectedWhitespace` diagnostic + the fix) ----
 *
 *   { code: '08      .toExponential()', output: null, parserOptions: { sourceType: 'script' },
 *     errors: [{ messageId: 'unexpectedWhitespace', data: { propName: 'toExponential' } }] }
 *   { code: '0192    .toExponential()', output: null, parserOptions: { sourceType: 'script' },
 *     errors: [{ messageId: 'unexpectedWhitespace', data: { propName: 'toExponential' } }] }
 *   { code: '05 .toExponential()',      output: '05.toExponential()', parserOptions: { sourceType: 'script' },
 *     errors: [{ messageId: 'unexpectedWhitespace', data: { propName: 'toExponential' } }] }
 *
 *   rslint: TypeScript(TS1489) "Decimals with leading zeros are not allowed." for
 *   the `08` / `0192` forms, and TypeScript(TS1121) "Octal literals are not
 *   allowed. Use the syntax '0o5'." for the `05` form -> 0 rule diagnostics each.
 *
 * ====================================================================================
 */
