/**
 * @fileoverview Tests for key-spacing rule.
 * @author Brandon Mills
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/key-spacing/key-spacing._js_.test.ts
 *   packages/eslint-plugin/rules/key-spacing/key-spacing._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('key-spacing', null as never, { valid, invalid })`
 *  - The `$` unindent template tag (from eslint-vitest-rule-tester) is evaluated to
 *    its real multi-line string. `unindent` strips the common leading indentation of
 *    the non-blank lines and drops the leading/trailing blank lines.
 *  - `[ ... ].join('\n')` multi-line code/output arrays are kept verbatim (a valid
 *    expression that produces the same string).
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint resolves via tsconfig.
 *  - `type` fields (deprecated AST node type) dropped (none were present).
 *  - All key-spacing messages carry `{{computed}}`/`{{key}}` data; that data is kept
 *    verbatim so the RuleTester can resolve the rendered message via the plugin's own
 *    `meta.messages` (e.g. extraKey: "Extra space after {{computed}}key '{{key}}'.").
 *
 * No Babel/Flow cases and no external-fixture (`readFileSync`) cases exist in the
 * upstream key-spacing tests. The `._css_` / `._json_` / `._markdown_` test files
 * don't exist for this rule.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are moved to the `KNOWN GAPS` block comment at the bottom, each annotated with what
 * upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('key-spacing', null as never, {
  valid: [
    // ==== from key-spacing._js_.test.ts ====

    '({\n})',
    '({\na: b\n})',
    {
      code: '({\n})',
      options: [{ align: 'colon' }],
    },
    {
      code: '({\na: b\n})',
      options: [{ align: 'value' }],
    },
    {
      code: 'var obj = { key: value };',
      options: [{}],
    },
    {
      code: 'var obj = { [(a + b)]: value };',
      options: [{}],
    },
    {
      code: 'var foo = { a:bar };',
      options: [{
        beforeColon: false,
        afterColon: false,
      }],
    },
    {
      code: 'var foo = { a: bar };',
      options: [{
        beforeColon: false,
        afterColon: true,
      }],
    },
    {
      code: 'foo({ \'default\': function(){}});',
      options: [{
        beforeColon: false,
        afterColon: true,
      }],
    },
    {
      code: 'function foo() { return {\n    key: (foo === 4)\n}; }',
      options: [{
        beforeColon: false,
        afterColon: true,
      }],
    },
    {
      code: 'var obj = {\'key\' :42 };',
      options: [{
        beforeColon: true,
        afterColon: false,
      }],
    },
    {
      code: '({a : foo, b : bar})[\'a\'];',
      options: [{
        beforeColon: true,
        afterColon: true,
      }],
    },
    {
      code: [
        'var obj = {',
        '    \'a\'     : (42 - 12),',
        '    foobar  : \'value\',',
        '    [(expr)]: val',
        '};',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
    },
    {
      code: [
        'callExpr(arg, {',
        '    key       :val,',
        '    \'another\' :false,',
        '    [compute] :\'value\'',
        '});',
      ].join('\n'),
      options: [{
        align: 'colon',
        beforeColon: true,
        afterColon: false,
      }],
    },
    {
      code: [
        'var obj = {',
        '    a:        (42 - 12),',
        '    \'foobar\': \'value\',',
        '    bat:      function() {',
        '        return this.a;',
        '    },',
        '    baz: 42',
        '};',
      ].join('\n'),
      options: [{
        align: 'value',
      }],
    },
    {
      code: [
        'callExpr(arg, {',
        '    \'asdf\' :val,',
        '    foobar :false,',
        '    key :   value',
        '});',
      ].join('\n'),
      options: [{
        align: 'value',
        beforeColon: true,
        afterColon: false,
      }],
    },
    {
      code: [
        '({',
        '    a  : 0,',
        '    // same group',
        '    bcd: 0, /*',
        '    end of group */',
        '',
        '    // different group',
        '    e: 0,',
        '    /* group b */',
        '    f: 0',
        '})',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
    },
    {
      code: [
        'obj = { key ',
        ' : ',
        ' longName };',
      ].join('\n'),
      options: [{
        beforeColon: true,
        afterColon: true,
      }],
    },
    {
      code: [
        'obj = { key ',
        '    :longName };',
      ].join('\n'),
      options: [{
        beforeColon: true,
        afterColon: false,
        mode: 'minimum',
      }],
    },
    {
      code: 'obj = { key     :longName };',
      options: [{
        beforeColon: true,
        afterColon: false,
        mode: 'minimum',
      }],
    },
    {
      code: 'var obj = { get fn() { return 42; } };',
      options: [{}],
    },
    {
      code: '({ get fn() {} })',
      options: [{ align: 'colon' }],
    },
    {
      code: 'var obj = {foo: \'fee\', bar: \'bam\'};',
      options: [{ align: 'colon' }],
    },
    {
      code: 'var obj = {a: \'foo\', bar: \'bam\'};',
      options: [{ align: 'colon' }],
    },
    {
      code: [
        'var x = {',
        '    foo: 10',
        '  , b  : 20',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
    },
    {
      code: [
        'var x = {',
        '    foo : 10',
        '  , b   : 20',
        '};',
      ].join('\n'),
      options: [{ align: 'colon', beforeColon: true }],
    },
    {
      code: [
        'var x = {',
        '        foo: 10,',
        ' /*lol*/b  : 20',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
    },
    {
      code: [
        'var a = \'a\';',
        'var b = \'b\';',
        '',
        'export default {',
        '    a,',
        '    b',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
    },
    {
      code: [
        'var test = {',
        '    prop: 123,',
        '    a,',
        '    b',
        '};',
      ].join('\n'),
    },
    {
      code: [
        'var test = {',
        '    prop: 456,',
        '    c,',
        '    d',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
    },
    {
      code: [
        'var obj = {',
        '    foobar: 123,',
        '    prop,',
        '    baz:    456',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
    },
    {
      code: [
        'var test = {',
        '    prop: 123,',
        '    a() { }',
        '};',
      ].join('\n'),
    },
    {
      code: [
        'var test = {',
        '    prop: 123,',
        '    a() { },',
        '    b() { }',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
    },
    {
      code: [
        'var obj = {',
        '    foobar: 123,',
        '    method() { },',
        '    baz:    456',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
    },
    {
      code: [
        'var obj = {',
        '    foobar: 123,',
        '    method() {',
        '        return 42;',
        '    },',
        '    baz: 456,',
        '    10:     ',
        '    10',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
    },
    {
      code: [
        'var obj = {',
        '    foo : foo',
        '  , bar : bar',
        '  , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
    },
    {
      code: [
        'var obj = { foo : foo',
        '          , bar : bar',
        '          , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
    },
    {
      code: [
        'var obj = {',
        '    foo :  foo',
        '  , bar :  bar',
        '  , cats : cats',
        '};',
      ].join('\n'),
      options: [{
        align: 'value',
        beforeColon: true,
      }],
    },

    // https://github.com/eslint/eslint/issues/4763
    {
      code: '({a : foo, ...x, b : bar})[\'a\'];',
      options: [{
        beforeColon: true,
        afterColon: true,
      }],
    },
    {
      code: [
        'var obj = {',
        '    \'a\'     : (42 - 12),',
        '    ...x,',
        '    foobar  : \'value\',',
        '    [(expr)]: val',
        '};',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
    },
    {
      code: [
        'callExpr(arg, {',
        '    key       :val,',
        '    ...x,',
        '    ...y,',
        '    \'another\' :false,',
        '    [compute] :\'value\'',
        '});',
      ].join('\n'),
      options: [{
        align: 'colon',
        beforeColon: true,
        afterColon: false,
      }],
    },
    {
      code: [
        'var obj = {',
        '    a:        (42 - 12),',
        '    ...x,',
        '    \'foobar\': \'value\',',
        '    bat:      function() {',
        '        return this.a;',
        '    },',
        '    barfoo:',
        '    [',
        '        1',
        '    ],',
        '    baz: 42',
        '};',
      ].join('\n'),
      options: [{
        align: 'value',
      }],
    },
    {
      code: [
        '({',
        '    ...x,',
        '    a  : 0,',
        '    // same group',
        '    bcd: 0, /*',
        '    end of group */',
        '',
        '    // different group',
        '    e: 0,',
        '    ...y,',
        '    /* group b */',
        '    f: 0',
        '})',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
    },

    // https://github.com/eslint/eslint/issues/4792
    {
      code: [
        '({',
        '    a: 42,',
        '    get b() { return 42; }',
        '})',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
    },
    {
      code: [
        '({',
        '    set a(b) { b; },',
        '    c: 42',
        '})',
      ].join('\n'),
      options: [{
        align: 'value',
      }],
    },
    {
      code: [
        '({',
        '    a  : 42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def: 42',
        '})',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
    },
    {
      code: [
        '({',
        '    a  : 42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def: 42',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          afterColon: true,
          align: 'colon',
        },
      }],
    },
    {
      code: [
        '({',
        '    a   : 42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def : 42,',
        '    obj : {a: 1, b: 2, c: 3}',
        '})',
      ].join('\n'),
      options: [{
        singleLine: {
          afterColon: true,
          beforeColon: false,
        },
        multiLine: {
          afterColon: true,
          beforeColon: true,
          align: 'colon',
        },
      }],
    },
    {
      code: [
        '({',
        '    a   : 42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def : 42,',
        '    def : {a: 1, b: 2, c: 3}',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          afterColon: true,
          beforeColon: true,
          align: 'colon',
        },
        singleLine: {
          afterColon: true,
          beforeColon: false,
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    foobar: 42,',
        '    bat:    2',
        '};',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: true,
          mode: 'strict',
        },
        multiLine: {
          beforeColon: false,
          afterColon: true,
          mode: 'minimum',
        },
      }],
    },

    // https://github.com/eslint/eslint/issues/5724
    {
      code: '({...object})',
      options: [{
        align: 'colon',
      }],
    },

    // https://github.com/eslint/eslint/issues/5613

    { // if `align` is an object, but `on` is not declared, `on` defaults to `colon`
      code: [
        '({',
        '    longName: 1,',
        '    small   : 2,',
        '    f       : function() {',
        '    },',
        '    xs :3',
        '})',
      ].join('\n'),
      options: [{
        align: {
          afterColon: true,
        },
        beforeColon: true,
        afterColon: false,
      }],
    },
    {
      code: [
        '({',
        '    longName: 1,',
        '    small:    2,',
        '    f:        function() {',
        '    },',
        '    xs :3',
        '})',
      ].join('\n'),
      options: [{
        align: {
          on: 'value',
          afterColon: true,
        },
        beforeColon: true,
        afterColon: false,
      }],
    },
    {
      code: [
        '({',
        '    longName : 1,',
        '    small :    2,',
        '    xs :       3',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          align: {
            on: 'value',
            beforeColon: true,
            afterColon: true,
          },
        },
      }],
    },
    {
      code: [
        '({',
        '    longName :1,',
        '    small    :2,',
        '    xs       :3',
        '})',
      ].join('\n'),
      options: [{
        align: {
          on: 'colon',
          beforeColon: true,
          afterColon: false,
        },
      }],
    },
    {
      code: [
        '({',
        '    longName: 1,',
        '    small   : 2,',
        '    xs      :        3',
        '})',
      ].join('\n'),
      options: [{
        align: {
          on: 'colon',
          beforeColon: false,
          afterColon: true,
          mode: 'minimum',
        },
      }],
    },
    {
      code: [
        '({',
        '    longName: 1,',
        '    small   : 2,',
        '    xs      : 3',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          align: {
            on: 'colon',
            beforeColon: false,
            afterColon: true,
          },
        },
      }],
    },
    {
      code: [
        '({',
        '    func: function() {',
        '        var test = true;',
        '    },',
        '    longName : 1,',
        '    small    : 2,',
        '    xs       : 3,',
        '    func2    : function() {',
        '        var test2 = true;',
        '    },',
        '    internalGroup: {',
        '        internal : true,',
        '        ext      : false',
        '    },',
        '    func3:',
        '    function () {',
        '        var test3 = true;',
        '    }',
        '})',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: true,
        },
        multiLine: {
          beforeColon: false,
          afterColon: true,
        },
        align: {
          on: 'colon',
          beforeColon: true,
          afterColon: true,
        },
      }],
    },
    {
      code: [
        '({',
        '    func: function() {',
        '        var test = true;',
        '    },',
        '    longName: 1,',
        '    small:    2,',
        '    xs:       3,',
        '    func2:    function() {',
        '        var test2 = true;',
        '    },',
        '    final: 10',
        '})',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: true,
        },
        multiLine: {
          align: {
            on: 'value',
            beforeColon: false,
            afterColon: true,
          },
          beforeColon: false,
          afterColon: true,
        },
      }],
    },
    {
      code: [
        '({',
        '    f:function() {',
        '        var test = true;',
        '    },',
        '    stateName : \'NY\',',
        '    borough   : \'Brooklyn\',',
        '    zip       : 11201,',
        '    f2        : function() {',
        '        var test2 = true;',
        '    },',
        '    final:10',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          align: {
            on: 'colon',
            beforeColon: true,
            afterColon: true,
            mode: 'strict',
          },
          beforeColon: false,
          afterColon: false,
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    key1: 1,',
        '',
        '    key2:    2,',
        '    key3:    3,',
        '',
        '    key4: 4',
        '}',
      ].join('\n'),
      options: [{
        multiLine: {
          beforeColon: false,
          afterColon: true,
          mode: 'strict',
          align: {
            beforeColon: false,
            afterColon: true,
            on: 'colon',
            mode: 'minimum',
          },
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    key1: 1,',
        '',
        '    key2:    2,',
        '    key3:    3,',
        '',
        '    key4: 4',
        '}',
      ].join('\n'),
      options: [{
        multiLine: {
          beforeColon: false,
          afterColon: true,
          mode: 'strict',
        },
        align: {
          beforeColon: false,
          afterColon: true,
          on: 'colon',
          mode: 'minimum',
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    foo : 1, \'bar\' : 2, baz : 3, longlonglong : 4',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        multiLine: {
          beforeColon: true,
          afterColon: true,
          align: 'colon',
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    foo: 1, \'bar\': 2, baz: 3',
        '}',
      ].join('\n'),
      options: [{
        multiLine: {
          align: 'value',
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    foo: 1',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        multiLine: {
          align: 'value',
        },
      }],
    },
    {
      code: [
        'foo({',
        '    bar: 1',
        '})',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'colon',
        },
      }],
    },
    {
      code: 'var obj = { foo:1, \'bar\':2, baz:3, longlonglong:4 }',
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'colon',
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    foo         : 1,',
        '    \'bar\'       : 2, baz         : 3, longlonglong: 4',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'colon',
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    foo:          1,',
        '    \'bar\':        2, baz:          3, longlonglong: 4',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'value',
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    foo         : 1,',
        '    \'bar\'       : 2, baz         : 3,',
        '    longlonglong: 4',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'colon',
        },
      }],
    },
    {
      code: [
        'var obj = {',
        '    foo:          1,',
        '    \'bar\':        2, baz:          3,',
        '    longlonglong: 4',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'value',
        },
      }],
    },

    // https://github.com/eslint/eslint/issues/15914
    {
      code: 'var foo = {\n    "a": "bar",\n    "𐌘": "baz"\n};',
      options: [{
        align: {
          on: 'value',
        },
      }],
    },
    {
      code: 'var foo = {\n    "a": "bar",\n    "Á": "baz",\n    "o͂": "qux",\n    "m̅": "xyz",\n    "ř": "abc"\n\n};',
      options: [{
        align: {
          on: 'value',
        },
      }],
    },
    {
      code: 'var foo = {\n    "🌷": "bar", // 1 grapheme, 1 code point, 2 code units\n    "🎁": "baz", // 1 grapheme, 1 code point, 2 code units\n    "🇮🇳": "qux", // 1 grapheme, 2 code points, 4 code units\n    "🏳️‍🌈": "xyz", // 1 grapheme, 4 code points, 6 code units\n};',
      options: [{
        align: {
          on: 'value',
        },
      }],
    },
    {
      code: 'const foo = {\n    "a": "bar",\n    [𐌘]: "baz"\n};',
      options: [{
        align: {
          on: 'value',
        },
      }],
    },
    {
      code: 'const foo = {\n    "abc": "bar",\n    [ 𐌘 ]: "baz"\n};',
      options: [{
        align: {
          on: 'value',
        },
      }],
    },

    // https://github.com/eslint/eslint/issues/16490
    {
      code: 'var foo =\n{\n    id:   1,\n    code: 2,\n    [n]:  3,\n    message:\n    "some value on the next line",\n};',
      options: [{
        align: 'value',
      }],
    },
    {
      code: 'var foo =\n{\n    id   : 1,\n    code : 2,\n    message :\n    "some value on the next line",\n};',
      options: [{
        align: 'colon',
        beforeColon: true,
      }],
    },
    {
      code: '({\n    a: 1,\n    // different group\n    bcd:\n    2\n})',
      options: [{
        align: 'value',
      }],
    },
    {
      code: '({\n    foo  :  1,\n    bar  :  2,\n    foobar :\n    3\n})',
      options: [{
        align: 'value',
        beforeColon: true,
        mode: 'minimum',
      }],
    },
    {
      code: '({\n    oneLine: 1,\n    ["some key " +\n    "spanning multiple lines"]: 2\n})',
      options: [{
        align: 'value',
      }],
    },

    // https://github.com/eslint/eslint/issues/16674
    {
      code: 'a = {\n    item       : 123,\n    longerItem : (\n      1 + 1\n    ),\n};',
      options: [{
        align: {
          beforeColon: true,
          afterColon: true,
          on: 'colon',
        },
      }],
    },
    {
      code: 'a = {\n    item: 123,\n    longerItem: // a comment - not a token\n    (1 + 1),\n};',
      options: [{ align: 'value' }],
    },
    'import foo from "./foo" with { type: "json" }',
    'import "./foo" with { type: "json" }',
    'export {foo} from "./foo" with { type: "json" }',
    'export * from "./foo" with { type: "json" }',
    {
      code: 'import foo from "./foo" with\n    { type:"json", foo:"bar" }',
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        multiLine: {
          beforeColon: true,
          afterColon: true,
        },
      }],
    },
    {
      code: 'import foo from "./foo" with\n    { type : "json",\n      foo : "bar" }',
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        multiLine: {
          beforeColon: true,
          afterColon: true,
        },
      }],
    },
    {
      code: 'import foo from "./foo" with\n    {\n      type : "json",\n      foo  : "bar"\n    }',
      options: [{
        beforeColon: true,
        afterColon: true,
        align: 'colon',
      }],
    },
    {
      code: 'export {foo} from "./foo" with\n    {\n      type : "json",\n      foo  : "bar"\n    }',
      options: [{
        beforeColon: true,
        afterColon: true,
        align: 'colon',
      }],
    },
    {
      code: 'export * from "./foo" with\n    {\n      type : "json",\n      foo  : "bar"\n    }',
      options: [{
        beforeColon: true,
        afterColon: true,
        align: 'colon',
      }],
    },
    {
      code: 'import foo from "./foo"',
      options: [{ align: 'colon' }],
    },
    {
      code: 'import "./foo"',
      options: [{ align: 'colon' }],
    },
    {
      code: 'export {foo} from "./foo"',
      options: [{ align: 'colon' }],
    },
    {
      code: 'export * from "./foo"',
      options: [{ align: 'colon' }],
    },
    {
      code: 'var obj = {\n  \'a\': 42 - 12,\n  foobar : \'value\',\n  [(expr)] :val\n}',
      options: [{ ignoredNodes: ['ObjectExpression'] }],
    },
    {
      code: 'var {\n  a: b,\n  c : d,\n  e :f,\n} = obj',
      options: [{ ignoredNodes: ['ObjectPattern'] }],
    },
    {
      code: 'import foo from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      options: [{ ignoredNodes: ['ImportDeclaration'] }],
    },
    {
      code: 'export {foo} from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      options: [{ ignoredNodes: ['ExportNamedDeclaration'] }],
    },
    {
      code: 'export * from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      options: [{ ignoredNodes: ['ExportAllDeclaration'] }],
    },
    {
      code: 'var obj = {\n  \'a\': (42 - 12),\n  foobar: \'value\',\n  [(expr)]: val\n}',
      options: [{
        align: 'colon',
        ignoredNodes: ['ObjectExpression'],
      }],
    },
    {
      code: 'import foo from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      options: [{
        align: 'colon',
        ignoredNodes: ['ImportDeclaration'],
      }],
    },
    {
      code: 'export {foo} from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      options: [{
        align: 'colon',
        ignoredNodes: ['ExportNamedDeclaration'],
      }],
    },
    {
      code: 'export * from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      options: [{
        align: 'colon',
        ignoredNodes: ['ExportAllDeclaration'],
      }],
    },

    // ==== from key-spacing._ts_.test.ts ====
    // non-applicable
    {
      code: 'interface X {\n  x:\n    | number\n    | string;\n}',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  x:\n    | number\n    | string;\n}',
      options: [{}],
    },
    {
      code: 'interface X {\n  abcdef: string;\n  x:\n    | number\n    | string;\n  defgh: string;\n}',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  x:\n    | number; abcd: string;\n}',
      options: [{ align: 'value' }],
    },
    // align: value
    {
      code: 'interface X {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  "a:b": number;\n  abcde: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'let x: {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'let x: {\n  a:   number;\n  "𐌘": string;\n  [𐌘]: Date;\n  "🌷": "bar", // 2 code points\n  "🎁": "baz", // 2 code points\n  "🇮🇳": "qux", // 4 code points\n  "🏳️‍🌈": "xyz", // 6 code points\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  a: number;\n  abc: string; c: number;\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  a: number;\n  abc: string; c: number; de: boolean;\n  abcef: number;\n};',
      options: [{ align: 'colon' }],
    },
    {
      code: 'interface X {\n  a    : number;\n  abc;\n  abcef: number;\n};',
      options: [{ align: 'colon' }],
    },
    {
      code: 'interface X {\n  a?:  number;\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  a:   number;\n  // Some comment\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  a:   number;\n  // Some comment\n  // on multiple lines\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  a:   number;\n  /**\n   * Some comment\n   * on multiple lines\n   */\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  a:   number;\n  /**\n   * Doc comment\n  */\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'interface X {\n  a: number;\n\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'class X {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'class X {\n  a?:  number;\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'class X {\n  x:     number;\n  z = 1;\n  xbcef: number;\n  }',
      options: [{ align: 'value' }],
    },
    {
      code: 'class X {\n  a: number;\n\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'type X = {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'type X = {\n  a: number;\n\n  abc: string\n};',
      options: [{ align: 'value' }],
    },
    {
      code: 'type X = {\n  a :  number;\n  abc: string\n};',
      options: [{ align: 'value', mode: 'minimum' }],
    },
    {
      code: 'type X = {\n  a :  number;\n  abc: string\n};',
      options: [
        {
          align: {
            on: 'value',
            mode: 'minimum',
            beforeColon: false,
            afterColon: true,
          },
        },
      ],
    },
    {
      code: 'interface X {\n  a:    number;\n  prop: {\n    abc: number;\n    a:   number;\n  };\n  abc: string\n}',
      options: [{ align: 'value' }],
    },
    {
      code: 'class X {\n  a:    number;\n  prop: {\n    abc: number;\n    a:   number;\n  };\n  abc: string\n  x = 1;\n  d:   number;\n  z:   number = 1;\n  ef:  string;\n}',
      options: [{ align: 'value' }],
    },
    // align: colon
    {
      code: 'interface X {\n  a  : number;\n  abc: string\n};',
      options: [{ align: 'colon' }],
    },
    {
      code: 'interface X {\n  a  :number;\n  abc:string\n};',
      options: [{ align: 'colon', afterColon: false }],
    },
    {
      code: 'interface X {\n  a  :   number;\n  abc: string\n};',
      options: [{ align: 'colon', mode: 'minimum' }],
    },
    // no align
    {
      code: 'interface X {\n  a: number;\n  abc: string\n};',
      options: [{}],
    },
    {
      code: 'interface X {\n  a : number;\n  abc : string\n};',
      options: [{ beforeColon: true }],
    },
    // singleLine / multiLine
    {
      code: 'interface X {\n  a : number;\n  abc : string\n};',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
    },
    {
      code: 'interface X {\n  a :   number;\n  abc : string\n};',
      options: [
        {
          align: { on: 'value', beforeColon: true, afterColon: true },
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: { beforeColon: false, afterColon: false },
        },
      ],
    },
    {
      code: 'interface X {\n  a   : number;\n  abc : string\n};',
      options: [
        {
          align: { beforeColon: true, afterColon: true }, // defaults to 'colon'
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: { beforeColon: false, afterColon: false },
        },
      ],
    },
    {
      code: 'interface X {\n  a :   number;\n  abc : string\n};',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: { beforeColon: true, afterColon: true, align: 'value' },
        },
      ],
    },
    {
      code: 'interface X {\n  a   : number;\n  abc : string\n};',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: {
            beforeColon: true,
            afterColon: true,
            align: {
              on: 'colon',
              mode: 'strict',
              afterColon: true,
              beforeColon: true,
            },
          },
        },
      ],
    },
    {
      code: 'interface X {\n  a   : number;\n  abc : string\n};',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: {
            beforeColon: true,
            afterColon: true,
            align: {
              mode: 'strict',
              afterColon: true,
              beforeColon: true,
            },
          },
        },
      ],
    },
    {
      code: 'interface X {\n  a   : number;\n  abc : string\n};',
      options: [
        {
          beforeColon: true,
          afterColon: true,
          align: {
            on: 'colon',
            mode: 'strict',
            afterColon: true,
            beforeColon: true,
          },
        },
      ],
    },
    {
      code: 'interface X {\n  a   : number;\n  abc : string\n};',
      options: [
        {
          beforeColon: true,
          afterColon: true,
          align: {
            mode: 'strict',
            afterColon: true,
            beforeColon: true,
          },
        },
      ],
    },
    {
      code: 'interface X {\n  a  : number;\n  abc: string\n\n  xadzd : number;\n};',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: {
            beforeColon: true,
            afterColon: true,
            align: {
              on: 'colon',
              mode: 'strict',
              afterColon: true,
              beforeColon: false,
            },
          },
        },
      ],
    },
    {
      code: 'interface X {\n  a  : number;\n  abc: string\n\n  xadzd : number;\n};',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: {
            beforeColon: true,
            afterColon: true,
            mode: 'strict',
            align: {
              on: 'colon',
              afterColon: true,
              beforeColon: false,
            },
          },
        },
      ],
    },
    {
      code: 'interface X {\n  a  :    number;\n  abc: string\n\n  xadzd :    number;\n};',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: {
            beforeColon: true,
            afterColon: true,
            mode: 'minimum',
            align: {
              on: 'colon',
              afterColon: true,
              beforeColon: false,
            },
          },
        },
      ],
    },
    {
      code: 'interface X { a:number; abc:string; };',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
    },
    {
      code: 'class Foo {\n  a: (b)\n}',
    },
    {
      code: 'interface Foo {\n  a: (b)\n}',
    },
    {
      code: 'class Foo {\n  a: /** comment */ b\n}',
    },
    {
      code: 'class Foo {\n  a: (     b)\n}',
    },
    {
      code: 'class Foo { a: (b) }',
    },
    {
      code: 'class Foo {\n  a?: (string | number)\n}',
    },
    {
      code: 'type X = {\n  a :number;\n  b: string;\n  c :string;\n};',
      options: [{ ignoredNodes: ['TSTypeLiteral'] }],
    },
    {
      code: 'interface X {\n  a :number;\n  b: string;\n  c :string;\n};',
      options: [{ ignoredNodes: ['TSInterfaceBody'] }],
    },
    {
      code: 'class X {\n  a :number;\n  b: string;\n  c :string;\n};',
      options: [{ ignoredNodes: ['ClassBody'] }],
    },
    {
      code: 'type X = {\n  a: number;\n  abc: string\n};',
      options: [{ align: 'value', ignoredNodes: ['TSTypeLiteral'] }],
    },
    {
      code: 'interface X {\n  a: number;\n  abc: string\n};',
      options: [{ align: 'value', ignoredNodes: ['TSInterfaceBody'] }],
    },
    {
      code: 'class X {\n  a: number;\n  abc: string\n};',
      options: [{ align: 'value', ignoredNodes: ['ClassBody'] }],
    },
  ],
  invalid: [
    // ==== from key-spacing._js_.test.ts ====

    {
      code: 'var a ={\'key\' : value };',
      output: 'var a ={\'key\':value };',
      options: [{
        beforeColon: false,
        afterColon: false,
      }],
      errors: [
        {

          messageId: 'extraKey',
          data: { computed: '', key: 'key' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'extraValue',
          data: { computed: '', key: 'key' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'var a ={\'key\' :value };',
      output: 'var a ={\'key\': value };',
      options: [{
        beforeColon: false,
        afterColon: true,
      }],
      errors: [
        {

          messageId: 'extraKey',
          data: { computed: '', key: 'key' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'missingValue',
          data: { computed: '', key: 'key' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'var a ={\'key\'\n : \nvalue };',
      output: 'var a ={\'key\':value };',
      options: [{
        beforeColon: false,
        afterColon: false,
      }],
      errors: [
        {

          messageId: 'extraKey',
          data: { computed: '', key: 'key' },
          line: 1,
          column: 14,
          endLine: 2,
          endColumn: 2,
        },
        {
          messageId: 'extraValue',
          data: { computed: '', key: 'key' },
          line: 2,
          column: 2,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'var bat = function() { return { foo:bar, \'key\': value }; };',
      output: 'var bat = function() { return { foo:bar, \'key\':value }; };',
      options: [{
        beforeColon: false,
        afterColon: false,
      }],
      errors: [
        {
          messageId: 'extraValue',
          data: { computed: '', key: 'key' },
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'var obj = { [ (a + b) ]:value };',
      output: 'var obj = { [ (a + b) ]: value };',
      options: [{}],
      errors: [{ messageId: 'missingValue', data: { computed: 'computed ', key: 'a + b' }, line: 1, column: 25 }],
    },
    {
      code: 'fn({ foo:bar, \'key\' :value });',
      output: 'fn({ foo:bar, \'key\':value });',
      options: [{
        beforeColon: false,
        afterColon: false,
      }],
      errors: [{ messageId: 'extraKey', data: { computed: '', key: 'key' }, line: 1, column: 20, endLine: 1, endColumn: 21 }],
    },
    {
      code: 'var obj = {prop :(42)};',
      output: 'var obj = {prop : (42)};',
      options: [{
        beforeColon: true,
        afterColon: true,
      }],
      errors: [{ messageId: 'missingValue', data: { computed: '', key: 'prop' }, line: 1, column: 18 }],
    },
    {
      code: '({\'a\' : foo, b: bar() }).b();',
      output: '({\'a\' : foo, b : bar() }).b();',
      options: [{
        beforeColon: true,
        afterColon: true,
      }],
      errors: [{ messageId: 'missingKey', data: { computed: '', key: 'b' }, line: 1, column: 14 }],
    },
    {
      code: '({\'a\'  :foo(), b:  bar() }).b();',
      output: '({\'a\' : foo(), b : bar() }).b();',
      options: [{
        beforeColon: true,
        afterColon: true,
      }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'a' }, line: 1, column: 6, endLine: 1, endColumn: 8 },
        { messageId: 'missingValue', data: { computed: '', key: 'a' }, line: 1, column: 9, endLine: 1, endColumn: 12 },
        { messageId: 'missingKey', data: { computed: '', key: 'b' }, line: 1, column: 16, endLine: 1, endColumn: 17 },
        { messageId: 'extraValue', data: { computed: '', key: 'b' }, line: 1, column: 17, endLine: 1, endColumn: 20 },
      ],
    },
    {
      code: 'bar = { key:value };',
      output: 'bar = { key: value };',
      options: [{
        beforeColon: false,
        afterColon: true,
      }],
      errors: [{ messageId: 'missingValue', data: { computed: '', key: 'key' }, line: 1, column: 13 }],
    },
    {
      code: [
        'obj = {',
        '    key:   value,',
        '    foobar:fn(),',
        '    \'a\'   : (2 * 2)',
        '};',
      ].join('\n'),
      output: [
        'obj = {',
        '    key   : value,',
        '    foobar: fn(),',
        '    \'a\'   : (2 * 2)',
        '};',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'key' }, line: 2, column: 5 },
        { messageId: 'extraValue', data: { computed: '', key: 'key' }, line: 2, column: 8 },
        { messageId: 'missingValue', data: { computed: '', key: 'foobar' }, line: 3, column: 12 },
      ],
    },
    {
      code: [
        '({',
        '    \'a\' : val,',
        '    foo:fn(),',
        '    b    :[42],',
        '    c   :call()',
        '}).a();',
      ].join('\n'),
      output: [
        '({',
        '    \'a\' :val,',
        '    foo :fn(),',
        '    b   :[42],',
        '    c   :call()',
        '}).a();',
      ].join('\n'),
      options: [{
        align: 'colon',
        beforeColon: true,
        afterColon: false,
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'a' }, line: 2, column: 9 },
        { messageId: 'missingKey', data: { computed: '', key: 'foo' }, line: 3, column: 5 },
        { messageId: 'extraKey', data: { computed: '', key: 'b' }, line: 4, column: 6 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    a:    fn(),',
        '    \'b\' : 42,',
        '    foo:(bar),',
        '    bat: \'valid\',',
        '    [a] : value',
        '};',
      ].join('\n'),
      output: [
        'var obj = {',
        '    a:   fn(),',
        '    \'b\': 42,',
        '    foo: (bar),',
        '    bat: \'valid\',',
        '    [a]: value',
        '};',
      ].join('\n'),
      options: [{
        align: 'value',
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'a' }, line: 2, column: 6 },
        { messageId: 'extraKey', data: { computed: '', key: 'b' }, line: 3, column: 8 },
        { messageId: 'missingValue', data: { computed: '', key: 'foo' }, line: 4, column: 9 },
        { messageId: 'extraKey', data: { computed: 'computed ', key: 'a' }, line: 6, column: 8 },
      ],
    },
    {
      code: [
        'foo = {',
        '    a:  value,',
        '    b :  42,',
        '    foo :[\'a\'],',
        '    bar : call()',
        '};',
      ].join('\n'),
      output: [
        'foo = {',
        '    a :  value,',
        '    b :  42,',
        '    foo :[\'a\'],',
        '    bar :call()',
        '};',
      ].join('\n'),
      options: [{
        align: 'value',
        beforeColon: true,
        afterColon: false,
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'a' }, line: 2, column: 5 },
        { messageId: 'extraValue', data: { computed: '', key: 'bar' }, line: 5, column: 9 },
      ],
    },
    {
      code: [
        '({',
        '    a : 0,',
        '    bcd: 0,',
        '',
        '    e: 0,',
        '    fg:0',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    a  : 0,',
        '    bcd: 0,',
        '',
        '    e : 0,',
        '    fg: 0',
        '})',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'a' }, line: 2, column: 5 },
        { messageId: 'missingKey', data: { computed: '', key: 'e' }, line: 5, column: 5 },
        { messageId: 'missingValue', data: { computed: '', key: 'fg' }, line: 6, column: 8 },
      ],
    },
    {
      code: [
        'foo = {',
        '    key:',
        '        longValueName,',
        '    key2',
        '        :anotherLongValue',
        '};',
      ].join('\n'),
      output: [
        'foo = {',
        '    key:longValueName,',
        '    key2:anotherLongValue',
        '};',
      ].join('\n'),
      options: [{
        beforeColon: false,
        afterColon: false,
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'key' }, line: 2, column: 8 },
        { messageId: 'extraKey', data: { computed: '', key: 'key2' }, line: 4, column: 9 },
      ],
    },
    {
      code: [
        'foo = {',
        '    key1: 42,',
        '    // still the same group',
        '    key12: \'42\', /*',
        '',
        '    */',
        '    key123: \'forty two\'',
        '};',
      ].join('\n'),
      output: [
        'foo = {',
        '    key1:   42,',
        '    // still the same group',
        '    key12:  \'42\', /*',
        '',
        '    */',
        '    key123: \'forty two\'',
        '};',
      ].join('\n'),
      options: [{
        align: 'value',
      }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'key1' } },
        { messageId: 'missingValue', data: { computed: '', key: 'key12' } },
      ],
    },
    {
      code: 'foo = { key:(1+2) };',
      output: 'foo = { key: (1+2) };',
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'key' }, line: 1, column: 13 },
      ],
    },
    {
      code: 'foo = { key:( ( (1+2) ) ) };',
      output: 'foo = { key: ( ( (1+2) ) ) };',
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'key' }, line: 1, column: 13 },
      ],
    },
    {
      code: 'var obj = {a  : \'foo\', bar: \'bam\'};',
      output: 'var obj = {a: \'foo\', bar: \'bam\'};',
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'a' }, line: 1, column: 13 },
      ],
    },
    {
      code: [
        'var x = {',
        '    foo: 10',
        '  , b   : 20',
        '};',
      ].join('\n'),
      output: [
        'var x = {',
        '    foo: 10',
        '  , b  : 20',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'b' }, line: 3, column: 6 },
      ],
    },
    {
      code: [
        'var x = {',
        '        foo : 10,',
        ' /*lol*/  b : 20',
        '};',
      ].join('\n'),
      output: [
        'var x = {',
        '        foo : 10,',
        ' /*lol*/  b   : 20',
        '};',
      ].join('\n'),
      options: [{ align: 'colon', beforeColon: true }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'b' }, line: 3, column: 11 },
      ],
    },
    {
      code: [
        'obj = { key ',
        ' :     longName };',
      ].join('\n'),
      output: [
        'obj = { key ',
        ' : longName };',
      ].join('\n'),
      options: [{
        beforeColon: true,
        afterColon: true,
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'key' }, line: 2, column: 2 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foobar: 123,',
        '    prop,',
        '    baz: 456',
        '};',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foobar: 123,',
        '    prop,',
        '    baz:    456',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'baz' }, line: 4, column: 10 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foobar:  123,',
        '    prop,',
        '    baz:    456',
        '};',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foobar: 123,',
        '    prop,',
        '    baz:    456',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'foobar' }, line: 2, column: 11 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foobar: 123,',
        '    method() { },',
        '    baz: 456',
        '};',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foobar: 123,',
        '    method() { },',
        '    baz:    456',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'baz' }, line: 4, column: 10 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foobar:  123,',
        '    method() { },',
        '    baz:    456',
        '};',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foobar: 123,',
        '    method() { },',
        '    baz:    456',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'foobar' }, line: 2, column: 11 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foobar: 123,',
        '    method() {',
        '        return 42;',
        '    },',
        '    baz:    456,',
        '    10:     ',
        '    10',
        '};',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foobar: 123,',
        '    method() {',
        '        return 42;',
        '    },',
        '    baz: 456,',
        '    10:     ',
        '    10',
        '};',
      ].join('\n'),
      options: [{ align: 'value' }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'baz' }, line: 6, column: 8 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foo: foo',
        '  , cats: cats',
        '};',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foo : foo',
        '  , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' }, line: 2, column: 5 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foo : foo',
        '  , cats:  cats',
        '};',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foo : foo',
        '  , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'cats' }, line: 3, column: 9 },
      ],
    },
    {
      code: [
        'var obj = { foo: foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      output: [
        'var obj = { foo : foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' }, line: 1, column: 13 },
      ],
    },
    {
      code: [
        'var obj = { foo  : foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      output: [
        'var obj = { foo : foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'foo' }, line: 1, column: 16 },
      ],
    },
    {
      code: [
        'var obj = { foo :foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      output: [
        'var obj = { foo : foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'foo' }, line: 1, column: 18 },
      ],
    },
    {
      code: [
        'var obj = { foo :  foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      output: [
        'var obj = { foo : foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'foo' }, line: 1, column: 17 },
      ],
    },
    {
      code: [
        'var obj = { foo : foo',
        '          , cats:  cats',
        '};',
      ].join('\n'),
      output: [
        'var obj = { foo : foo',
        '          , cats: cats',
        '};',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'cats' }, line: 2, column: 17 },
      ],
    },
    // https://github.com/eslint/eslint/issues/4763
    {
      code: [
        '({',
        '    ...x,',
        '    a : 0,',
        '    // same group',
        '    bcd: 0, /*',
        '    end of group */',
        '',
        '    // different group',
        '    e: 0,',
        '    ...y,',
        '    /* group b */',
        '    f : 0',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    ...x,',
        '    a  : 0,',
        '    // same group',
        '    bcd: 0, /*',
        '    end of group */',
        '',
        '    // different group',
        '    e: 0,',
        '    ...y,',
        '    /* group b */',
        '    f: 0',
        '})',
      ].join('\n'),
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'a' }, line: 3, column: 5 },
        { messageId: 'extraKey', data: { computed: '', key: 'f' }, line: 12, column: 6 },
      ],
    },
    // https://github.com/eslint/eslint/issues/4792
    {
      code: [
        '({',
        '    a : 42,',
        '    get b() { return 42; }',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    a: 42,',
        '    get b() { return 42; }',
        '})',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'a' }, line: 2, column: 6 },
      ],
    },
    {
      code: [
        '({',
        '    set a(b) { b; },',
        '    c : 42',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    set a(b) { b; },',
        '    c: 42',
        '})',
      ].join('\n'),
      options: [{
        align: 'value',
      }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'c' }, line: 3, column: 6 },
      ],
    },
    {
      code: [
        '({',
        '    a: 42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def: 42',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    a  : 42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def: 42',
        '})',
      ].join('\n'),
      options: [{
        align: 'colon',
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'a' }, line: 2, column: 5 },
      ],
    },
    {
      code: [
        '({',
        '    a :    42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def  :  42,',
        '    def2 : {a1: 1, b1:2, c1:3}',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    a :    42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def  :  42,',
        '    def2 : {a1:1, b1:2, c1:3}',
        '})',
      ].join('\n'),
      options: [{
        singleLine: {
          afterColon: false,
          beforeColon: false,
        },
        multiLine: {
          mode: 'minimum',
          afterColon: true,
          beforeColon: true,
          align: 'value',
        },
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'a1' }, line: 6, column: 15 },
      ],
    },
    {
      code: [
        '({',
        '    a  : 42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def: 42,',
        '    de1: {a2: 1, b2 : 2, c2 : 3}',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    a  : 42,',
        '    get b() { return 42; },',
        '    set c(v) { v; },',
        '    def: 42,',
        '    de1: {a2 : 1, b2 : 2, c2 : 3}',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          afterColon: true,
          beforeColon: false,
          align: 'colon',
        },
        singleLine: {
          afterColon: true,
          beforeColon: true,
        },
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'a2' }, line: 6, column: 11 },
      ],
    },
    {
      code: [
        'obj = {',
        '   get fx() { return \'f\'; },',
        '   get gx() { return \'g\'; },',
        '   ex:e',
        '};',
      ].join('\n'),
      output: [
        'obj = {',
        '   get fx() { return \'f\'; },',
        '   get gx() { return \'g\'; },',
        '   ex: e',
        '};',
      ].join('\n'),
      options: [{
        align: 'colon',
        beforeColon: false,
        afterColon: true,
        mode: 'minimum',
      }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'ex' }, line: 4, column: 7 },
      ],
    },
    {
      code: [
        'obj = {',
        '   get fx() { return \'f\'; },',
        '   get gx() { return \'g\'; },',
        '   ex : e',
        '};',
      ].join('\n'),
      output: [
        'obj = {',
        '   get fx() { return \'f\'; },',
        '   get gx() { return \'g\'; },',
        '   ex: e',
        '};',
      ].join('\n'),
      options: [{
        align: 'colon',
        beforeColon: false,
        afterColon: true,
        mode: 'minimum',
      }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'ex' }, line: 4, column: 6 },
      ],
    },
    {
      code: [
        '({',
        '    aInv :43,',
        '    get b() { return 43; },',
        '    set c(v) { v; },',
        '    defInv: 43',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    aInv  : 43,',
        '    get b() { return 43; },',
        '    set c(v) { v; },',
        '    defInv: 43',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          afterColon: true,
          align: 'colon',
        },
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'aInv' }, line: 2, column: 5 },
        { messageId: 'missingValue', data: { computed: '', key: 'aInv' }, line: 2, column: 11 },
      ],
    },
    // https://github.com/eslint/eslint/issues/5724
    {
      code: '({ a:b, ...object, c : d })',
      output: '({ a: b, ...object, c: d })',
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'a' }, line: 1, column: 6 },
        { messageId: 'extraKey', data: { computed: '', key: 'c' }, line: 1, column: 21 },
      ],
    },
    // https://github.com/eslint/eslint/issues/5613
    {
      code: [
        '({',
        '    longName:1,',
        '    small    :2,',
        '    xs      : 3',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    longName : 1,',
        '    small    : 2,',
        '    xs       : 3',
        '})',
      ].join('\n'),
      options: [{
        align: {
          on: 'colon',
          beforeColon: true,
          afterColon: true,
          mode: 'strict',
        },
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'longName' }, line: 2, column: 5 },
        { messageId: 'missingValue', data: { computed: '', key: 'longName' }, line: 2, column: 14 },
        { messageId: 'missingValue', data: { computed: '', key: 'small' }, line: 3, column: 15 },
        { messageId: 'missingKey', data: { computed: '', key: 'xs' }, line: 4, column: 5 },
      ],
    },
    {
      code: [
        '({',
        '    func:function() {',
        '        var test = true;',
        '    },',
        '    longName: 1,',
        '    small: 2,',
        '    xs            : 3,',
        '    func2    : function() {',
        '        var test2 = true;',
        '    },',
        '    singleLine : 10',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    func: function() {',
        '        var test = true;',
        '    },',
        '    longName : 1,',
        '    small    : 2,',
        '    xs       : 3,',
        '    func2    : function() {',
        '        var test2 = true;',
        '    },',
        '    singleLine: 10',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          beforeColon: false,
          afterColon: true,
        },
        align: {
          on: 'colon',
          beforeColon: true,
          afterColon: true,
          mode: 'strict',
        },
      }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'func' }, line: 2, column: 10 },
        { messageId: 'missingKey', data: { computed: '', key: 'longName' }, line: 5, column: 5 },
        { messageId: 'missingKey', data: { computed: '', key: 'small' }, line: 6, column: 5 },
        { messageId: 'extraKey', data: { computed: '', key: 'xs' }, line: 7, column: 7 },
        { messageId: 'extraKey', data: { computed: '', key: 'singleLine' }, line: 11, column: 15 },
      ],
    },
    {
      code: [
        '({',
        '    func:function() {',
        '        var test = false;',
        '    },',
        '    longName :1,',
        '    small :2,',
        '    xs            : 3,',
        '    func2    : function() {',
        '        var test2 = true;',
        '    },',
        '    singleLine : 10',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    func: function() {',
        '        var test = false;',
        '    },',
        '    longName :1,',
        '    small    :2,',
        '    xs       :3,',
        '    func2    :function() {',
        '        var test2 = true;',
        '    },',
        '    singleLine: 10',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          beforeColon: false,
          afterColon: true,
          align: {
            on: 'colon',
            beforeColon: true,
            afterColon: false,
            mode: 'strict',
          },
        },
      }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'func' }, line: 2, column: 10 },
        { messageId: 'missingKey', data: { computed: '', key: 'small' }, line: 6, column: 5 },
        { messageId: 'extraKey', data: { computed: '', key: 'xs' }, line: 7, column: 7 },
        { messageId: 'extraValue', data: { computed: '', key: 'xs' }, line: 7, column: 19 },
        { messageId: 'extraValue', data: { computed: '', key: 'func2' }, line: 8, column: 14 },
        { messageId: 'extraKey', data: { computed: '', key: 'singleLine' }, line: 11, column: 15 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    key1: 1,',
        '',
        '    key2:    2,',
        '    key3:    3,',
        '',
        '    key4: 4',
        '}',
      ].join('\n'),
      output: [
        'var obj = {',
        '    key1: 1,',
        '',
        '    key2: 2,',
        '    key3: 3,',
        '',
        '    key4: 4',
        '}',
      ].join('\n'),
      options: [{
        multiLine: {
          beforeColon: false,
          afterColon: true,
          mode: 'strict',
          align: {
            beforeColon: false,
            afterColon: true,
            on: 'colon',
          },
        },
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'key2' }, line: 4, column: 9 },
        { messageId: 'extraValue', data: { computed: '', key: 'key3' }, line: 5, column: 9 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    key1: 1,',
        '',
        '    key2:    2,',
        '    key3:    3,',
        '',
        '    key4: 4',
        '}',
      ].join('\n'),
      output: [
        'var obj = {',
        '    key1: 1,',
        '',
        '    key2: 2,',
        '    key3: 3,',
        '',
        '    key4: 4',
        '}',
      ].join('\n'),
      options: [{
        multiLine: {
          beforeColon: false,
          afterColon: true,
          mode: 'strict',
        },
        align: {
          beforeColon: false,
          afterColon: true,
          on: 'colon',
        },
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'key2' }, line: 4, column: 9 },
        { messageId: 'extraValue', data: { computed: '', key: 'key3' }, line: 5, column: 9 },
      ],
    },
    {

      // https://github.com/eslint/eslint/issues/7603
      code: '({ foo/* comment */ : bar })',
      output: '({ foo/* comment */: bar })',
      errors: [{ messageId: 'extraKey', data: { computed: '', key: 'foo' }, line: 1, column: 20 }],
    },
    {
      code: '({ foo: /* comment */bar })',
      output: '({ foo:/* comment */bar })',
      options: [{ afterColon: false }],
      errors: [{ messageId: 'extraValue', data: { computed: '', key: 'foo' }, line: 1, column: 7 }],
    },
    {
      code: '({ foo/*comment*/:/*comment*/bar })',
      output: '({ foo/*comment*/ : /*comment*/bar })',
      options: [{ beforeColon: true, afterColon: true }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' }, line: 1, column: 7 },
        { messageId: 'missingValue', data: { computed: '', key: 'foo' }, line: 1, column: 19 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foo:1, \'bar\':2, baz:3',
        '}',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foo : 1, \'bar\' : 2, baz : 3',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        multiLine: {
          beforeColon: true,
          afterColon: true,
          align: 'colon',
        },
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' }, line: 2, column: 5 },
        { messageId: 'missingValue', data: { computed: '', key: 'foo' }, line: 2, column: 9 },
        { messageId: 'missingKey', data: { computed: '', key: 'bar' }, line: 2, column: 12 },
        { messageId: 'missingValue', data: { computed: '', key: 'bar' }, line: 2, column: 18 },
        { messageId: 'missingKey', data: { computed: '', key: 'baz' }, line: 2, column: 21 },
        { messageId: 'missingValue', data: { computed: '', key: 'baz' }, line: 2, column: 25 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foo : 1, \'bar\' : 2, baz : 3, longlonglong : 4',
        '}',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foo: 1, \'bar\': 2, baz: 3, longlonglong: 4',
        '}',
      ].join('\n'),
      options: [{
        multiLine: {
          align: 'value',
        },
      }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'foo' }, line: 2, column: 8 },
        { messageId: 'extraKey', data: { computed: '', key: 'bar' }, line: 2, column: 19 },
        { messageId: 'extraKey', data: { computed: '', key: 'baz' }, line: 2, column: 28 },
        { messageId: 'extraKey', data: { computed: '', key: 'longlonglong' }, line: 2, column: 46 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foo:1',
        '}',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foo: 1',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        multiLine: {
          align: 'value',
        },
      }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'foo' }, line: 2, column: 9 },
      ],
    },
    {
      code: [
        'foo({',
        '    bar:1',
        '})',
      ].join('\n'),
      output: [
        'foo({',
        '    bar: 1',
        '})',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'colon',
        },
      }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'bar' }, line: 2, column: 9 },
      ],
    },
    {
      code: 'var obj = { foo: 1, \'bar\': 2, baz :3, longlonglong :4 }',
      output: 'var obj = { foo:1, \'bar\':2, baz:3, longlonglong:4 }',
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'colon',
        },
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'foo' }, line: 1, column: 16 },
        { messageId: 'extraValue', data: { computed: '', key: 'bar' }, line: 1, column: 26 },
        { messageId: 'extraKey', data: { computed: '', key: 'baz' }, line: 1, column: 34 },
        { messageId: 'extraKey', data: { computed: '', key: 'longlonglong' }, line: 1, column: 51 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foo: 1,',
        '    \'bar\': 2, baz: 3, longlonglong: 4',
        '}',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foo         : 1,',
        '    \'bar\'       : 2, baz         : 3, longlonglong: 4',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'colon',
        },
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' }, line: 2, column: 5 },
        { messageId: 'missingKey', data: { computed: '', key: 'bar' }, line: 3, column: 5 },
        { messageId: 'missingKey', data: { computed: '', key: 'baz' }, line: 3, column: 15 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foo : 1,',
        '    \'bar\' : 2, baz : 3,',
        '    longlonglong: 4',
        '}',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foo         : 1,',
        '    \'bar\'       : 2, baz         : 3,',
        '    longlonglong: 4',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'colon',
        },
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' }, line: 2, column: 5 },
        { messageId: 'missingKey', data: { computed: '', key: 'bar' }, line: 3, column: 5 },
        { messageId: 'missingKey', data: { computed: '', key: 'baz' }, line: 3, column: 16 },
      ],
    },
    {
      code: [
        'var obj = {',
        '    foo: 1,',
        '    \'bar\': 2, baz: 3,',
        '    longlonglong: 4',
        '}',
      ].join('\n'),
      output: [
        'var obj = {',
        '    foo:          1,',
        '    \'bar\':        2, baz:          3,',
        '    longlonglong: 4',
        '}',
      ].join('\n'),
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        align: {
          on: 'value',
        },
      }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'foo' }, line: 2, column: 10 },
        { messageId: 'missingValue', data: { computed: '', key: 'bar' }, line: 3, column: 12 },
        { messageId: 'missingValue', data: { computed: '', key: 'baz' }, line: 3, column: 20 },
      ],
    },
    {
      code: 'const foo = {\n    "a": "bar",\n    [ 𐌘 ]: "baz"\n};',
      output: 'const foo = {\n    "a":   "bar",\n    [ 𐌘 ]: "baz"\n};',
      options: [{
        align: {
          on: 'value',
        },
      }],
      errors: [
        { messageId: 'missingValue', data: { computed: '', key: 'a' } },
      ],
    },
    {
      code: 'const foo = {\n    "a": "bar",\n    [ 𐌘 ]: "baz"\n};',
      output: 'const foo = {\n    "a"  : "bar",\n    [ 𐌘 ]: "baz"\n};',
      options: [{
        align: {
          on: 'colon',
        },
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'a' } },
      ],
    },
    {
      code: 'const foo = {\n    "a":  "bar",\n    "𐌘": "baz"\n};',
      output: 'const foo = {\n    "a": "bar",\n    "𐌘": "baz"\n};',
      options: [{
        align: {
          on: 'value',
        },
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'a' } },
      ],
    },
    {
      code: 'var foo = {\n    "🌷":     "bar", // 1 grapheme, 1 code point, 2 code units\n    "🎁":     "baz", // 1 grapheme, 1 code point, 2 code units\n    "🇮🇳":   "qux", // 1 grapheme, 2 code points, 4 code units\n    "🏳️‍🌈": "xyz", // 1 grapheme, 4 code points, 6 code units\n};',
      output: 'var foo = {\n    "🌷": "bar", // 1 grapheme, 1 code point, 2 code units\n    "🎁": "baz", // 1 grapheme, 1 code point, 2 code units\n    "🇮🇳": "qux", // 1 grapheme, 2 code points, 4 code units\n    "🏳️‍🌈": "xyz", // 1 grapheme, 4 code points, 6 code units\n};',
      options: [{
        align: {
          on: 'value',
        },
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: '🌷' } },
        { messageId: 'extraValue', data: { computed: '', key: '🎁' } },
        { messageId: 'extraValue', data: { computed: '', key: '🇮🇳' } },
      ],
    },
    // https://github.com/eslint/eslint/issues/16490
    {
      code: 'var foo =\n{\n    id:      1,\n    code:    2,\n    [n]:     3,\n    message:\n    "some value on the next line",\n};',
      output: 'var foo =\n{\n    id:   1,\n    code: 2,\n    [n]:  3,\n    message:\n    "some value on the next line",\n};',
      options: [{
        align: 'value',
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'id' } },
        { messageId: 'extraValue', data: { computed: '', key: 'code' } },
        { messageId: 'extraValue', data: { computed: 'computed ', key: 'n' } },
      ],
    },
    {
      code: 'var foo =\n{\n    id      : 1,\n    code    : 2,\n    message :\n    "some value on the next line",\n};',
      output: 'var foo =\n{\n    id   : 1,\n    code : 2,\n    message :\n    "some value on the next line",\n};',
      options: [{
        align: 'colon',
        beforeColon: true,
      }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'id' } },
        { messageId: 'extraKey', data: { computed: '', key: 'code' } },
      ],
    },
    {
      code: '({\n    a:   1,\n    // different group\n    bcd:\n    2\n})',
      output: '({\n    a: 1,\n    // different group\n    bcd:\n    2\n})',
      options: [{
        align: 'value',
      }],
      errors: [
        { messageId: 'extraValue', data: { computed: '', key: 'a' } },
      ],
    },
    {
      code: [
        '({',
        '    singleLine : 10,',
        '    newGroup :',
        '    function() {',
        '        var test3 = true;',
        '    }',
        '})',
      ].join('\n'),
      output: [
        '({',
        '    singleLine: 10,',
        '    newGroup:',
        '    function() {',
        '        var test3 = true;',
        '    }',
        '})',
      ].join('\n'),
      options: [{
        multiLine: {
          beforeColon: false,
        },
        align: {
          on: 'colon',
          beforeColon: true,
        },
      }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'singleLine' } },
        { messageId: 'extraKey', data: { computed: '', key: 'newGroup' } },
      ],
    },
    // https://github.com/eslint/eslint/issues/16674
    {
      code: 'c = {\n    item: 123,\n    longerItem: (\n      1 + 1\n    ),\n};',
      output: 'c = {\n    item      : 123,\n    longerItem: (\n      1 + 1\n    ),\n};',
      options: [{ align: 'colon' }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'item' } },
      ],
    },
    {
      code: 'import foo from "./foo" with { type  :"json" }',
      output: 'import foo from "./foo" with { type: "json" }',
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'type' } },
        { messageId: 'missingValue', data: { computed: '', key: 'type' } },
      ],
    },
    {
      code: 'import "./foo" with { type  :"json" }',
      output: 'import "./foo" with { type: "json" }',
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'type' } },
        { messageId: 'missingValue', data: { computed: '', key: 'type' } },
      ],
    },
    {
      code: 'export {foo} from "./foo" with { type  :"json" }',
      output: 'export {foo} from "./foo" with { type: "json" }',
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'type' } },
        { messageId: 'missingValue', data: { computed: '', key: 'type' } },
      ],
    },
    {
      code: 'export * from "./foo" with { type  :"json" }',
      output: 'export * from "./foo" with { type: "json" }',
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'type' } },
        { messageId: 'missingValue', data: { computed: '', key: 'type' } },
      ],
    },
    {
      code: 'import foo from "./foo" with\n    { type : "json", foo : "bar" }',
      output: 'import foo from "./foo" with\n    { type:"json", foo:"bar" }',
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        multiLine: {
          beforeColon: true,
          afterColon: true,
        },
      }],
      errors: [
        { messageId: 'extraKey', data: { computed: '', key: 'type' } },
        { messageId: 'extraValue', data: { computed: '', key: 'type' } },
        { messageId: 'extraKey', data: { computed: '', key: 'foo' } },
        { messageId: 'extraValue', data: { computed: '', key: 'foo' } },
      ],
    },
    {
      code: 'import foo from "./foo" with\n    { type:"json",\n      foo:"bar" }',
      output: 'import foo from "./foo" with\n    { type : "json",\n      foo : "bar" }',
      options: [{
        singleLine: {
          beforeColon: false,
          afterColon: false,
        },
        multiLine: {
          beforeColon: true,
          afterColon: true,
        },
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'type' } },
        { messageId: 'missingValue', data: { computed: '', key: 'type' } },
        { messageId: 'missingKey', data: { computed: '', key: 'foo' } },
        { messageId: 'missingValue', data: { computed: '', key: 'foo' } },
      ],
    },
    {
      code: 'import foo from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      output: 'import foo from "./foo" with\n    {\n      type : "json",\n      foo  : "bar"\n    }',
      options: [{
        beforeColon: true,
        afterColon: true,
        align: 'colon',
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' } },
      ],
    },
    {
      code: 'export {foo} from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      output: 'export {foo} from "./foo" with\n    {\n      type : "json",\n      foo  : "bar"\n    }',
      options: [{
        beforeColon: true,
        afterColon: true,
        align: 'colon',
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' } },
      ],
    },
    {
      code: 'const a = import("./foo", { with: {\n  type : "json",\n  foo : "bar"\n}})',
      output: 'const a = import("./foo", { with : {\n  type : "json",\n  foo  : "bar"\n}})',
      options: [{
        beforeColon: true,
        afterColon: true,
        align: 'colon',
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'with' } },
        { messageId: 'missingKey', data: { computed: '', key: 'foo' } },
      ],
    },
    {
      code: 'export * from "./foo" with\n    {\n      type : "json",\n      foo : "bar"\n    }',
      output: 'export * from "./foo" with\n    {\n      type : "json",\n      foo  : "bar"\n    }',
      options: [{
        beforeColon: true,
        afterColon: true,
        align: 'colon',
      }],
      errors: [
        { messageId: 'missingKey', data: { computed: '', key: 'foo' } },
      ],
    },

    // ==== from key-spacing._ts_.test.ts ====
    // align: value
    {
      code: 'interface X {\n  a: number;\n  abc: string\n};',
      output: 'interface X {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'interface X {\n  a: number;\n  "a:c": string\n};',
      output: 'interface X {\n  a:     number;\n  "a:c": string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'let x: {\n  a: number;\n  abc: string\n};',
      output: 'let x: {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'let x: {\n  a: number;\n  abc: string\n};',
      output: 'let x: {\n  a:   number;\n  abc: string\n};',
      options: [{ align: { on: 'value' } }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'let x: {\n  a: number;\n  "🌷": "bar", // 2 code points\n  "🎁": "baz", // 2 code points\n  "🇮🇳": "qux", // 4 code points\n  "🏳️‍🌈": "xyz", // 6 code points\n  [𐌘]: string\n  "𐌘": string\n};',
      output: 'let x: {\n  a:   number;\n  "🌷": "bar", // 2 code points\n  "🎁": "baz", // 2 code points\n  "🇮🇳": "qux", // 4 code points\n  "🏳️‍🌈": "xyz", // 6 code points\n  [𐌘]: string\n  "𐌘": string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'class X {\n  a: number;\n  abc: string\n};',
      output: 'class X {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'class X {\n  a: number;\n  abc: string\n};',
      output: 'class X {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value', mode: 'minimum' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'class X {\n  a: number;\n  b;\n  abc: string\n};',
      output: 'class X {\n  a:   number;\n  b;\n  abc: string\n};',
      options: [{ align: 'value', mode: 'minimum' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'type X = {\n  a: number;\n  abc: string\n};',
      output: 'type X = {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'interface X {\n  a:   number;\n  abc:  string\n};',
      output: 'interface X {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'class X {\n  a:   number;\n  abc:  string\n};',
      output: 'class X {\n  a:   number;\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'class X {\n  x:   number;\n  z = 1;\n  xbcef: number;\n  }',
      output: 'class X {\n  x:     number;\n  z = 1;\n  xbcef: number;\n  }',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'interface X {\n  a:   number;\n\n  abc     : string\n};',
      output: 'interface X {\n  a: number;\n\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'extraValue' }, { messageId: 'extraKey' }],
    },
    {
      code: 'class X {\n  a:   number;\n\n  abc     : string\n};',
      output: 'class X {\n  a: number;\n\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'extraValue' }, { messageId: 'extraKey' }],
    },
    {
      code: 'interface X {\n  a:   number;\n  // Some comment\n\n  // interrupted in the middle\n  abc: string\n};',
      output: 'interface X {\n  a: number;\n  // Some comment\n\n  // interrupted in the middle\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'interface X {\n  a:   number;\n  /**\n   * Multiline comment\n   */\n\n  /** interrupted in the middle */\n  abc: string\n};',
      output: 'interface X {\n  a: number;\n  /**\n   * Multiline comment\n   */\n\n  /** interrupted in the middle */\n  abc: string\n};',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'interface X {\n  a:   number;\n  prop: {\n    abc: number;\n    a:   number;\n  },\n  abc: string\n}',
      output: 'interface X {\n  a:    number;\n  prop: {\n    abc: number;\n    a:   number;\n  },\n  abc: string\n}',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'interface X {\n  a:    number;\n  prop: {\n    abc: number;\n    a:  number;\n  },\n  abc: string\n}',
      output: 'interface X {\n  a:    number;\n  prop: {\n    abc: number;\n    a:   number;\n  },\n  abc: string\n}',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'missingValue' }],
    },
    {
      code: 'interface X {\n  a:    number;\n  prop: {\n    abc: number;\n    a:   number;\n  },\n  abc:  string\n}',
      output: 'interface X {\n  a:    number;\n  prop: {\n    abc: number;\n    a:   number;\n  },\n  abc: string\n}',
      options: [{ align: 'value' }],
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'class X {\n  a:      number;\n  prop: {\n    abc: number;\n    a?: number;\n  };\n  abc: string;\n  x = 1;\n  d: number;\n  z:  number = 1;\n  ef: string;\n}',
      output: 'class X {\n  a:    number;\n  prop: {\n    abc: number;\n    a?:  number;\n  };\n  abc: string;\n  x = 1;\n  d:   number;\n  z:   number = 1;\n  ef:  string;\n}',
      options: [{ align: 'value' }],
      errors: [
        { messageId: 'extraValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
      ],
    },
    // align: colon
    {
      code: 'interface X {\n  a   : number;\n  abc: string\n};',
      output: 'interface X {\n  a  : number;\n  abc: string\n};',
      options: [{ align: 'colon' }],
      errors: [{ messageId: 'extraKey' }],
    },
    {
      code: 'interface X {\n  a   : number;\n  abc: string\n};',
      output: 'interface X {\n  a  : number;\n  abc: string\n};',
      options: [{ align: { on: 'colon' } }],
      errors: [{ messageId: 'extraKey' }],
    },
    {
      code: 'interface X {\n  a   : number;\n  abc: string\n};',
      output: 'interface X {\n  a   : number;\n  abc : string\n};',
      options: [{ align: 'colon', beforeColon: true, afterColon: true }],
      errors: [{ messageId: 'missingKey' }],
    },
    // no align
    {
      code: 'interface X {\n  [x: number]:  string;\n}',
      output: 'interface X {\n  [x: number]: string;\n}',
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'interface X {\n  [x: number]:string;\n}',
      output: 'interface X {\n  [x: number]: string;\n}',
      errors: [{ messageId: 'missingValue' }],
    },
    // singleLine / multiLine
    {
      code: 'interface X {\n  a:number;\n  abc:string\n};',
      output: 'interface X {\n  a : number;\n  abc : string\n};',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
      errors: [
        { messageId: 'missingKey' },
        { messageId: 'missingValue' },
        { messageId: 'missingKey' },
        { messageId: 'missingValue' },
      ],
    },
    {
      code: 'interface X { a : number; abc : string; };',
      output: 'interface X { a:number; abc:string; };',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: false },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
      errors: [
        { messageId: 'extraKey' },
        { messageId: 'extraValue' },
        { messageId: 'extraKey' },
        { messageId: 'extraValue' },
      ],
    },
    {
      code: 'interface X { a : number; abc : string; };',
      output: 'interface X { a: number; abc: string; };',
      options: [
        {
          singleLine: { beforeColon: false, afterColon: true },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
      errors: [{ messageId: 'extraKey' }, { messageId: 'extraKey' }],
    },
    {
      code: 'interface X { a:number; abc:string; };',
      output: 'interface X { a : number; abc : string; };',
      options: [
        {
          singleLine: { beforeColon: true, afterColon: true, mode: 'strict' },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
      errors: [
        { messageId: 'missingKey' },
        { messageId: 'missingValue' },
        { messageId: 'missingKey' },
        { messageId: 'missingValue' },
      ],
    },
    {
      code: 'interface X { a:number; abc:   string; };',
      output: 'interface X { a : number; abc :   string; };',
      options: [
        {
          singleLine: { beforeColon: true, afterColon: true, mode: 'minimum' },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
      errors: [
        { messageId: 'missingKey' },
        { messageId: 'missingValue' },
        { messageId: 'missingKey' },
      ],
    },
    {
      code: 'interface X { a : number; abc : string; };',
      output: 'interface X { a:number; abc:string; };',
      options: [
        {
          beforeColon: false,
          afterColon: false,
        },
      ],
      errors: [
        { messageId: 'extraKey' },
        { messageId: 'extraValue' },
        { messageId: 'extraKey' },
        { messageId: 'extraValue' },
      ],
    },
    {
      code: 'interface X { a:number; abc:string; };',
      output: 'interface X { a : number; abc : string; };',
      options: [
        {
          beforeColon: true,
          afterColon: true,
          mode: 'strict',
        },
      ],
      errors: [
        { messageId: 'missingKey' },
        { messageId: 'missingValue' },
        { messageId: 'missingKey' },
        { messageId: 'missingValue' },
      ],
    },
    {
      code: 'type Wacky = {\n    a: number;\n    b: string;\n    agc: number;\n    middle: Date | {\n        inner: {\n            a: boolean;\n            bc: boolean;\n            "🌷": "rose";\n        }\n        [x: number]: string;\n        abc: boolean;\n    }\n} & {\n    a: "string";\n    abc: number;\n}',
      output: 'type Wacky = {\n    a:      number;\n    b:      string;\n    agc:    number;\n    middle: Date | {\n        inner: {\n            a:   boolean;\n            bc:  boolean;\n            "🌷": "rose";\n        }\n        [x: number]: string;\n        abc:         boolean;\n    }\n} & {\n    a:   "string";\n    abc: number;\n}',
      options: [{ align: 'value' }],
      errors: [
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
      ],
    },
    {
      code: 'class Wacky {\n    a: number;\n    b?: string;\n    public z: number;\n    abc = 10;\n    private override xy: number;\n    static x = "test";\n    static abcdef: number = 1;\n    get fn(): number { return 0; };\n    inter: number;\n    get fn2(): number {\n      return 1;\n    };\n    agc: number;\n    middle: Date | {\n        inner: {\n            a: boolean;\n            bc: boolean;\n            "🌷": "rose";\n        }\n        [x: number]: string;\n        abc: boolean;\n    }\n}',
      output: 'class Wacky {\n    a:                   number;\n    b?:                  string;\n    public z:            number;\n    abc = 10;\n    private override xy: number;\n    static x = "test";\n    static abcdef:       number = 1;\n    get fn(): number { return 0; };\n    inter:               number;\n    get fn2(): number {\n      return 1;\n    };\n    agc:    number;\n    middle: Date | {\n        inner: {\n            a:   boolean;\n            bc:  boolean;\n            "🌷": "rose";\n        }\n        [x: number]: string;\n        abc:         boolean;\n    }\n}',
      options: [{ align: 'value' }],
      errors: [
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
        { messageId: 'missingValue' },
      ],
    },
    {
      code: 'class Foo {\n  a:  (b)\n}',
      output: 'class Foo {\n  a: (b)\n}',
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'interface Foo {\n  a:  (b)\n}',
      output: 'interface Foo {\n  a: (b)\n}',
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'class Foo {\n  a:  /** comment */ b\n}',
      output: 'class Foo {\n  a: /** comment */ b\n}',
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'class Foo {\n  a:    (     b)\n}',
      output: 'class Foo {\n  a: (     b)\n}',
      errors: [{ messageId: 'extraValue' }],
    },
    {
      code: 'interface X { a:(number); };',
      output: 'interface X { a : (number); };',
      options: [
        {
          singleLine: { beforeColon: true, afterColon: true, mode: 'strict' },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
      errors: [{ messageId: 'missingKey' }, { messageId: 'missingValue' }],
    },
    {
      code: 'interface X { a:/** comment */ number; };',
      output: 'interface X { a : /** comment */ number; };',
      options: [
        {
          singleLine: { beforeColon: true, afterColon: true, mode: 'strict' },
          multiLine: { beforeColon: true, afterColon: true },
        },
      ],
      errors: [{ messageId: 'missingKey' }, { messageId: 'missingValue' }],
    },
    {
      code: 'class Foo {\n  a:    (string | number)\n}',
      output: 'class Foo {\n  a: (string | number)\n}',
      errors: [{ messageId: 'extraValue' }],
    },
  ],
});

/**
 * ========================== key-spacing — KNOWN GAPS ==========================
 *
 * None. Every upstream case (all 161 valid + 121 invalid, including the
 * `$`-unindented Unicode fixtures — `𐌘`, `🌷`, `🇮🇳`, `🏳️‍🌈`, combining marks —
 * the `with { type: ... }` import attributes, the dynamic `import(...)` with
 * attributes, TS index signatures, and the full align / multiLine / singleLine
 * option matrix) runs in the green `ruleTester.run` above and matches upstream
 * exactly (diagnostic count, rendered message, and autofix output).
 *
 * ==============================================================================
 */

