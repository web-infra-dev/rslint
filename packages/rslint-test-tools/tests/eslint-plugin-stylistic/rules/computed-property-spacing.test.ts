/**
 * @fileoverview Disallows or enforces spaces inside computed properties.
 * @author Jamund Ferguson
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0, merging both run() blocks:
 *   packages/eslint-plugin/rules/computed-property-spacing/computed-property-spacing._js_.test.ts
 *   packages/eslint-plugin/rules/computed-property-spacing/computed-property-spacing._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, valid, invalid })` -> `ruleTester.run('computed-property-spacing', null as never, { valid, invalid })`.
 *    The `._js_` block (lang: 'js') and the `._ts_` block (default lang: 'ts') are
 *    merged into the single valid/invalid pair below (js cases first, ts cases last).
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *    Upstream `[...].join('\n')` array-literal sources are likewise flattened to a
 *    single `'...\n...'` string.
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint resolves via
 *    tsconfig and is always `esnext`. The rule's options (`'always'`/`'never'` plus
 *    `enforceForClassMembers`) do not depend on ecmaVersion normalization.
 *  - No `type` AST fields, no `filename`, no `settings`, no `suggestions` exist in
 *    the upstream tests.
 *
 * The rule's `meta.messages` interpolate `{{tokenValue}}` from each error's `data`,
 * so the RuleTester renders e.g. `data: { tokenValue: '[' }` against:
 *   unexpectedSpaceBefore: 'There should be no space before \'{{tokenValue}}\'.'
 *   unexpectedSpaceAfter:  'There should be no space after \'{{tokenValue}}\'.'
 *   missingSpaceBefore:    'A space is required before \'{{tokenValue}}\'.'
 *   missingSpaceAfter:     'A space is required after \'{{tokenValue}}\'.'
 * A few upstream errors omit `data` (pinning only messageId + columns); those are
 * ported as-is — the RuleTester then checks columns only, not the rendered message.
 *
 * Every upstream invalid case pins an explicit `errors` array (no output-only cases).
 *
 * No rslint<->upstream gap surfaced for this rule: every valid case reports zero
 * diagnostics and every invalid case matches upstream's diagnostic count, rendered
 * message, pinned positions, and single-pass autofix output (including the `accessor`
 * and `TSIndexedAccessType` cases from the `._ts_` block). There is therefore no
 * `KNOWN GAPS` block below. (Had any case diverged — a parser-level syntax error, a
 * multi-pass fix difference, etc. — it would be moved here verbatim, annotated with
 * upstream-expected vs. rslint-produced, never deleted or altered to force green.)
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('computed-property-spacing', null as never, {
  valid: [
    'obj[foo]',
    'obj[\'foo\']',
    {
      code: 'var x = {[b]: a}',
    },
    {
      code: 'obj[ foo ]',
      options: ['always'],
    },
    {
      code: 'obj[\nfoo\n]',
      options: ['always'],
    },
    {
      code: 'obj[ \'foo\' ]',
      options: ['always'],
    },
    {
      code: 'obj[ \'foo\' + \'bar\' ]',
      options: ['always'],
    },
    {
      code: 'obj[ obj2[ foo ] ]',
      options: ['always'],
    },
    {
      code: 'obj.map(function(item) { return [\n1,\n2,\n3,\n4\n]; })',
      options: ['always'],
    },
    {
      code: 'obj[ \'map\' ](function(item) { return [\n1,\n2,\n3,\n4\n]; })',
      options: ['always'],
    },
    {
      code: 'obj[ \'for\' + \'Each\' ](function(item) { return [\n1,\n2,\n3,\n4\n]; })',
      options: ['always'],
    },
    {
      code: 'var foo = obj[ 1 ]',
      options: ['always'],
    },
    {
      code: 'var foo = obj[ \'foo\' ];',
      options: ['always'],
    },
    {
      code: 'var foo = obj[ [1, 1] ];',
      options: ['always'],
    },
    {
      code: 'var x = {[ "a" ]: a}',
      options: ['always'],
    },
    {
      code: 'var y = {[ x ]: a}',
      options: ['always'],
    },
    {
      code: 'var x = {[ "a" ]() {}}',
      options: ['always'],
    },
    {
      code: 'var y = {[ x ]() {}}',
      options: ['always'],
    },
    {
      code: 'var foo = {};',
      options: ['always'],
    },
    {
      code: 'var foo = [];',
      options: ['always'],
    },
    {
      code: 'obj[foo]',
      options: ['never'],
    },
    {
      code: 'obj[\'foo\']',
      options: ['never'],
    },
    {
      code: 'obj[\'foo\' + \'bar\']',
      options: ['never'],
    },
    {
      code: 'obj[\'foo\'+\'bar\']',
      options: ['never'],
    },
    {
      code: 'obj[obj2[foo]]',
      options: ['never'],
    },
    {
      code: 'obj.map(function(item) { return [\n1,\n2,\n3,\n4\n]; })',
      options: ['never'],
    },
    {
      code: 'obj[\'map\'](function(item) { return [\n1,\n2,\n3,\n4\n]; })',
      options: ['never'],
    },
    {
      code: 'obj[\'for\' + \'Each\'](function(item) { return [\n1,\n2,\n3,\n4\n]; })',
      options: ['never'],
    },
    {
      code: 'obj[\nfoo]',
      options: ['never'],
    },
    {
      code: 'obj[foo\n]',
      options: ['never'],
    },
    {
      code: 'var foo = obj[1]',
      options: ['never'],
    },
    {
      code: 'var foo = obj[\'foo\'];',
      options: ['never'],
    },
    {
      code: 'var foo = obj[[ 1, 1 ]];',
      options: ['never'],
    },
    {
      code: 'var x = {["a"]: a}',
      options: ['never'],
    },
    {
      code: 'var y = {[x]: a}',
      options: ['never'],
    },
    {
      code: 'var x = {["a"]() {}}',
      options: ['never'],
    },
    {
      code: 'var y = {[x]() {}}',
      options: ['never'],
    },
    {
      code: 'var foo = {};',
      options: ['never'],
    },
    {
      code: 'var foo = [];',
      options: ['never'],
    },
    {
      code: 'class A { [ a ](){} }',
      options: ['never', { enforceForClassMembers: false }],
    },
    {
      code: 'A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }',
      options: ['never', { enforceForClassMembers: false }],
    },
    {
      code: 'A = class { [a](){} }',
      options: ['always', { enforceForClassMembers: false }],
    },
    {
      code: 'class A { [a](){} get [b](){} set [b](foo){} static [c](){} static get [d](){} static set [d](bar){} }',
      options: ['always', { enforceForClassMembers: false }],
    },
    {
      code: 'class A { [ a ]; }',
      options: ['never', { enforceForClassMembers: false }],
    },
    {
      code: 'class A { [a]; }',
      options: ['always', { enforceForClassMembers: false }],
    },
    {
      code: 'A = class { [a](){} }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'class A { [a] ( ) { } }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class { [ \n a \n ](){} }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'class A { [a](){} get [b](){} set [b](foo){} static [c](){} static get [d](){} static set [d](bar){} }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'class A { [ a ](){} }',
      options: ['always', { enforceForClassMembers: true }],
    },
    {
      code: 'class A { [ a ](){}[ b ](){} }',
      options: ['always', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class { [\na\n](){} }',
      options: ['always', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }',
      options: ['always', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class { [a]; static [a]; [a] = 0; static [a] = 0; }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class { [ a ]; static [ a ]; [ a ] = 0; static [ a ] = 0; }',
      options: ['always', { enforceForClassMembers: true }],
    },
    {
      code: 'class A { a ( ) { } get b(){} set b ( foo ){} static c (){} static get d() {} static set d( bar ) {} }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class {a(){}get b(){}set b(foo){}static c(){}static get d(){}static set d(bar){} }',
      options: ['always', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class { foo; #a; static #b; #c = 0; static #d = 0; }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class { foo; #a; static #b; #c = 0; static #d = 0; }',
      options: ['always', { enforceForClassMembers: true }],
    },
    {
      code: 'const foo = {\n  [ (a) ]: 1\n}',
      options: ['always'],
    },
    {
      code: 'const foo = {\n  [ ( a ) ]: 1\n}',
      options: ['always'],
    },
    {
      code: 'const foo = {\n  [( a )]: 1\n}',
      options: ['never'],
    },
    {
      code: 'const foo = {\n  [ /**/ a /**/ ]: 1\n}',
      options: ['always'],
    },
    {
      code: 'const foo = {\n  [/**/ a /**/]: 1\n}',
      options: ['never'],
    },
    {
      code: 'const foo = {\n  [ a[ b ] ]: 1\n}',
      options: ['always'],
    },
    {
      code: 'const foo = {\n  [a[b]]: 1\n}',
      options: ['never'],
    },
    {
      code: 'const foo = {\n  [ a[ /**/ b ]/**/ ]: 1\n}',
      options: ['always'],
    },
    {
      code: 'const foo = {\n  [/**/a[b /**/] /**/]: 1\n}',
      options: ['never'],
    },
    {
      code: 'const { [a]: someProp } = obj;',
      options: ['never'],
    },
    {
      code: '({ [a]: someProp } = obj);',
      options: ['never'],
    },
    {
      code: 'const { [ a ]: someProp } = obj;',
      options: ['always'],
    },
    {
      code: '({ [ a ]: someProp } = obj);',
      options: ['always'],
    },
    {
      code: 'obj = { foo: bar }',
    },
    {
      code: 'class A { accessor [ b ]; }',
      options: ['never', { enforceForClassMembers: false }],
    },
    {
      code: 'class A { accessor [b]; }',
      options: ['always', { enforceForClassMembers: false }],
    },
    {
      code: 'A = class { accessor [b] = 1 }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'class A { accessor [b] = 1 }',
      options: ['never', { enforceForClassMembers: true }],
    },
    {
      code: 'A = class {\n  accessor [\n    b\n  ] = 1\n}',
      options: ['never', { enforceForClassMembers: true }],
    },
    'type Foo = A[B]',
  ],

  invalid: [
    {
      code: 'var foo = obj[ 1];',
      output: 'var foo = obj[ 1 ];',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var foo = obj[1 ];',
      output: 'var foo = obj[ 1 ];',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'var foo = obj[ 1];',
      output: 'var foo = obj[1];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'var foo = obj[1 ];',
      output: 'var foo = obj[1];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'obj[ foo ]',
      output: 'obj[foo]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'obj[foo ]',
      output: 'obj[foo]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'obj[ foo]',
      output: 'obj[foo]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'var foo = obj[1]',
      output: 'var foo = obj[ 1 ]',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'obj[    foo]',
      output: 'obj[foo]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'obj[  foo  ]',
      output: 'obj[foo]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 7,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'obj[   foo ]',
      output: 'obj[foo]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'obj[ foo + \n  bar   ]',
      output: 'obj[foo + \n  bar]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'obj[\n foo  ]',
      output: 'obj[\n foo]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'var x = {[a]: b}',
      output: 'var x = {[ a ]: b}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var x = {[a ]: b}',
      output: 'var x = {[ a ]: b}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var x = {[ a]: b}',
      output: 'var x = {[ a ]: b}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var x = {[ a ]: b}',
      output: 'var x = {[a]: b}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var x = {[a ]: b}',
      output: 'var x = {[a]: b}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var x = {[ a]: b}',
      output: 'var x = {[a]: b}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var x = {[ a\n]: b}',
      output: 'var x = {[a\n]: b}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'class A { [ a ](){} }',
      output: 'class A { [a](){} }',
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class A { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }',
      output: 'class A { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 29,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 41,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 43,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 61,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 62,
          endLine: 1,
          endColumn: 63,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 81,
          endLine: 1,
          endColumn: 82,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 83,
          endLine: 1,
          endColumn: 84,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 102,
          endLine: 1,
          endColumn: 103,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 104,
          endLine: 1,
          endColumn: 105,
        },
      ],
    },
    {
      code: 'A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }',
      output: 'A = class { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }',
      options: ['never', {  }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 29,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 43,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 45,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 62,
          endLine: 1,
          endColumn: 63,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 64,
          endLine: 1,
          endColumn: 65,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 83,
          endLine: 1,
          endColumn: 84,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 85,
          endLine: 1,
          endColumn: 86,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 104,
          endLine: 1,
          endColumn: 105,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 106,
          endLine: 1,
          endColumn: 107,
        },
      ],
    },
    {
      code: 'A = class { [a](){} }',
      output: 'A = class { [ a ](){} }',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'A = class { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }',
      output: 'A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 38,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 40,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 56,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 58,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 74,
          endLine: 1,
          endColumn: 75,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 76,
          endLine: 1,
          endColumn: 77,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 93,
          endLine: 1,
          endColumn: 94,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 95,
          endLine: 1,
          endColumn: 96,
        },
      ],
    },
    {
      code: 'class A { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }',
      output: 'class A { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }',
      options: ['always', {  }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 36,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 38,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 54,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 56,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 72,
          endLine: 1,
          endColumn: 73,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 74,
          endLine: 1,
          endColumn: 75,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 91,
          endLine: 1,
          endColumn: 92,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 93,
          endLine: 1,
          endColumn: 94,
        },
      ],
    },
    {
      code: 'class A { [ a](){} }',
      output: 'class A { [a](){} }',
      options: ['never', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'A = class { [a](){} b(){} static [c ](){} static [d](){}}',
      output: 'A = class { [a](){} b(){} static [c](){} static [d](){}}',
      options: ['never', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'class A { get [a ](){} set [ a](foo){} get b(){} static set b(bar){} static get [ a](){} static set [a ](baz){} }',
      output: 'class A { get [a](){} set [a](foo){} get b(){} static set b(bar){} static get [a](){} static set [a](baz){} }',
      options: ['never', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 82,
          endLine: 1,
          endColumn: 83,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 103,
          endLine: 1,
          endColumn: 104,
        },
      ],
    },
    {
      code: 'A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }',
      output: 'A = class { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }',
      options: ['never', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 29,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 43,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 45,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 62,
          endLine: 1,
          endColumn: 63,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 64,
          endLine: 1,
          endColumn: 65,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 83,
          endLine: 1,
          endColumn: 84,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 85,
          endLine: 1,
          endColumn: 86,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 104,
          endLine: 1,
          endColumn: 105,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 106,
          endLine: 1,
          endColumn: 107,
        },
      ],
    },
    {
      code: 'class A { [ a]; [b ]; [ c ]; [ a] = 0; [b ] = 0; [ c ] = 0; }',
      output: 'class A { [a]; [b]; [c]; [a] = 0; [b] = 0; [c] = 0; }',
      options: ['never', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          column: 12,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          column: 19,
          endColumn: 20,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          column: 24,
          endColumn: 25,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          column: 26,
          endColumn: 27,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          column: 31,
          endColumn: 32,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          column: 42,
          endColumn: 43,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          column: 51,
          endColumn: 52,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          column: 53,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'class A { [ a](){} }',
      output: 'class A { [ a ](){} }',
      options: ['always', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'A = class { [ a ](){} b(){} static [c ](){} static [ d ](){}}',
      output: 'A = class { [ a ](){} b(){} static [ c ](){} static [ d ](){}}',
      options: ['always', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'class A { get [a ](){} set [ a](foo){} get b(){} static set b(bar){} static get [ a](){} static set [a ](baz){} }',
      output: 'class A { get [ a ](){} set [ a ](foo){} get b(){} static set b(bar){} static get [ a ](){} static set [ a ](baz){} }',
      options: ['always', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 84,
          endLine: 1,
          endColumn: 85,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 101,
          endLine: 1,
          endColumn: 102,
        },
      ],
    },
    {
      code: 'A = class { [a](){} get [b](){} set [c](foo){} static [d](){} static get [e](){} static set [f](bar){} }',
      output: 'A = class { [ a ](){} get [ b ](){} set [ c ](foo){} static [ d ](){} static get [ e ](){} static set [ f ](bar){} }',
      options: ['always', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 38,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 40,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 56,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 58,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 74,
          endLine: 1,
          endColumn: 75,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 76,
          endLine: 1,
          endColumn: 77,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 1,
          column: 93,
          endLine: 1,
          endColumn: 94,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 1,
          column: 95,
          endLine: 1,
          endColumn: 96,
        },
      ],
    },
    {
      code: 'class A { [ a]; [b ]; [c]; [ a] = 0; [b ] = 0; [c] = 0; }',
      output: 'class A { [ a ]; [ b ]; [ c ]; [ a ] = 0; [ b ] = 0; [ c ] = 0; }',
      options: ['always', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'missingSpaceBefore',
          column: 14,
          endColumn: 15,
        },
        {
          messageId: 'missingSpaceAfter',
          column: 17,
          endColumn: 18,
        },
        {
          messageId: 'missingSpaceAfter',
          column: 23,
          endColumn: 24,
        },
        {
          messageId: 'missingSpaceBefore',
          column: 25,
          endColumn: 26,
        },
        {
          messageId: 'missingSpaceBefore',
          column: 31,
          endColumn: 32,
        },
        {
          messageId: 'missingSpaceAfter',
          column: 38,
          endColumn: 39,
        },
        {
          messageId: 'missingSpaceAfter',
          column: 48,
          endColumn: 49,
        },
        {
          messageId: 'missingSpaceBefore',
          column: 50,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'const foo = {\n  [(a)]: 1\n}',
      output: 'const foo = {\n  [ (a) ]: 1\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 4,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'const foo = {\n  [( a )]: 1\n}',
      output: 'const foo = {\n  [ ( a ) ]: 1\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 4,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'const foo = {\n  [ ( a ) ]: 1\n}',
      output: 'const foo = {\n  [( a )]: 1\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 5,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'const foo = {\n  [/**/ a /**/]: 1\n}',
      output: 'const foo = {\n  [ /**/ a /**/ ]: 1\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 4,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'const foo = {\n  [ /**/ a /**/ ]: 1\n}',
      output: 'const foo = {\n  [/**/ a /**/]: 1\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 5,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'const foo = {\n  [a[b]]: 1\n}',
      output: 'const foo = {\n  [ a[ b ] ]: 1\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 4,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 6,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 8,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'const foo = {\n  [ a[ b ] ]: 1\n}',
      output: 'const foo = {\n  [a[b]]: 1\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 5,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 8,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 10,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'const foo = {\n  [a[/**/ b ]/**/]: 1\n}',
      output: 'const foo = {\n  [ a[ /**/ b ]/**/ ]: 1\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 4,
        },
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 6,
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'const foo = {\n  [ /**/a[ b /**/ ] /**/]: 1\n}',
      output: 'const foo = {\n  [/**/a[b /**/] /**/]: 1\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 5,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 12,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
          line: 2,
          column: 18,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'obj?.[1];',
      output: 'obj?.[ 1 ];',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
    {
      code: 'obj?.[ 1 ];',
      output: 'obj?.[1];',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
    {
      code: 'const { [ a]: someProp } = obj;',
      output: 'const { [a]: someProp } = obj;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
        },
      ],
    },
    {
      code: 'const { [a ]: someProp } = obj;',
      output: 'const { [a]: someProp } = obj;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
    {
      code: 'const { [ a ]: someProp } = obj;',
      output: 'const { [a]: someProp } = obj;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
    {
      code: '({ [ a ]: someProp } = obj);',
      output: '({ [a]: someProp } = obj);',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
    {
      code: 'const { [a]: someProp } = obj;',
      output: 'const { [ a ]: someProp } = obj;',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
    {
      code: '({ [a]: someProp } = obj);',
      output: '({ [ a ]: someProp } = obj);',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
    {
      code: 'class A { accessor [ a ] = 0 }',
      output: 'class A { accessor [a] = 0 }',
      options: ['never', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          column: 21,
          endColumn: 22,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          column: 23,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'class A { accessor [a] = 0 }',
      output: 'class A { accessor [ a ] = 0 }',
      options: ['always', { enforceForClassMembers: true }],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          column: 20,
          endColumn: 21,
        },
        {
          messageId: 'missingSpaceBefore',
          column: 22,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'type Foo = A[ B ]',
      output: 'type Foo = A[B]',
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { tokenValue: '[' },
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
    {
      code: 'type Foo = A[B]',
      output: 'type Foo = A[ B ]',
      options: ['always'],
      errors: [
        {
          messageId: 'missingSpaceAfter',
          data: { tokenValue: '[' },
        },
        {
          messageId: 'missingSpaceBefore',
          data: { tokenValue: ']' },
        },
      ],
    },
  ],
});
