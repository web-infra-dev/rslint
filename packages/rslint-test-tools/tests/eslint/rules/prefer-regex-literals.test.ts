import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-regex-literals', {
  valid: [
    '/abc/',
    '/abc/g',
    'new RegExp(pattern)',
    "new RegExp('\\\\p{Emoji_Presentation}\\\\P{Script_Extensions=Latin}' + '', `ug`)",
    "new RegExp('\\\\cA' + '')",
    "RegExp(pattern, 'g')",
    "new RegExp(f('a'))",
    "RegExp(prefix + 'a')",
    "new RegExp('a' + suffix)",
    'RegExp(`a` + suffix);',
    'new RegExp(String.raw`a` + suffix);',
    "RegExp('a', flags)",
    "const flags = 'gu';RegExp('a', flags)",
    "RegExp('a', 'g' + flags)",
    'new RegExp(String.raw`a`, flags);',
    'RegExp(`${prefix}abc`)',
    'new RegExp(`a${b}c`);',
    "new RegExp(`a${''}c`);",
    'new RegExp(String.raw`a${b}c`);',
    "new RegExp(String.raw`a${''}c`);",
    "new RegExp('a' + 'b')",
    'RegExp(1)',
    "new RegExp('(\\\\p{Emoji_Presentation})\\\\1' + '', `ug`)",
    "RegExp(String.raw`\\78\\126` + '\\\\5934', '' + `g` + '')",
    "func(new RegExp(String.raw`a${''}c\\d`, 'u'),new RegExp(String.raw`a${''}c\\d`, 'u'))",
    'new RegExp(\'\\\\[\' + "b\\\\]")',
    {
      code: 'new RegExp(/a/, flags);',
      options: { disallowRedundantWrapping: true },
    },
    {
      code: 'new RegExp(/a/, `u${flags}`);',
      options: { disallowRedundantWrapping: true },
    },
    { code: 'new RegExp(/a/);', options: {} },
    {
      code: 'new RegExp(/a/);',
      options: { disallowRedundantWrapping: false },
    },
    'new RegExp;',
    'new RegExp();',
    'RegExp();',
    "new RegExp('a', 'g', 'b');",
    "RegExp('a', 'g', 'b');",
    'new RegExp(`a`, `g`, `b`);',
    'RegExp(`a`, `g`, `b`);',
    'new RegExp(String.raw`a`, String.raw`g`, String.raw`b`);',
    'RegExp(String.raw`a`, String.raw`g`, String.raw`b`);',
    {
      code: "new RegExp(/a/, 'u', 'foo');",
      options: { disallowRedundantWrapping: true },
    },
    'new RegExp(String`a`);',
    'RegExp(raw`a`);',
    'new RegExp(f(String.raw)`a`);',
    'RegExp(string.raw`a`);',
    'new RegExp(String.Raw`a`);',
    'new RegExp(String[raw]`a`);',
    'RegExp(String.raw.foo`a`);',
    'new RegExp(String.foo.raw`a`);',
    'RegExp(foo.String.raw`a`);',
    'new RegExp(String.raw);',
    'let String; new RegExp(String.raw`a`);',
    'function foo() { var String; new RegExp(String.raw`a`); }',
    'function foo(String) { RegExp(String.raw`a`); }',
    'if (foo) { const String = bar; RegExp(String.raw`a`); }',
    { code: '/* globals String:off */ new RegExp(String.raw`a`);', skip: true },
    { code: "RegExp('a', String.raw`g`);", skip: true },
    "new Regexp('abc');",
    'Regexp(`a`);',
    'new Regexp(String.raw`a`);',
    "let RegExp; new RegExp('a');",
    "function foo() { var RegExp; RegExp('a', 'g'); }",
    'function foo(RegExp) { new RegExp(String.raw`a`); }',
    "if (foo) { const RegExp = bar; RegExp('a'); }",
    { code: "/* globals RegExp:off */ new RegExp('a');", skip: true },
    { code: "RegExp('a');", skip: true },
    { code: "new globalThis.RegExp('a');", skip: true },
    { code: "new globalThis.RegExp('a');", skip: true },
    { code: "new globalThis.RegExp('a');", skip: true },
    "class C { #RegExp; foo() { globalThis.#RegExp('a'); } }",
    "new RegExp('[[A--B]]' + a, 'v')",
  ],
  invalid: [
    { code: "new RegExp('abc');", errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: "RegExp('abc');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "new RegExp('abc', 'g');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('abc', 'g');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: 'new RegExp(`abc`);', errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: 'RegExp(`abc`);', errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: 'new RegExp(`abc`, `g`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'RegExp(`abc`, `g`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp(String.raw`abc`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp(String.raw`abc\nabc`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp(String.raw`\tabc\nabc`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'RegExp(String.raw`abc`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp(String.raw`abc`, String.raw`g`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'RegExp(String.raw`abc`, String.raw`g`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp(String['raw']`a`);",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "new RegExp('');", errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: "RegExp('', '');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: 'new RegExp(String.raw``);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('a', `g`);",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "RegExp(`a`, 'g');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "RegExp(String.raw`a`, 'g');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp(String.raw`\\d`, `g`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp(String.raw`\\\\d`, `g`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp(String['raw']`\\\\d`, `g`);",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp(String["raw"]`\\\\d`, `g`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('a', String.raw`g`);",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new globalThis.RegExp('a');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "globalThis.RegExp('a');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp(/a/);',
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExp', line: 1, column: 1 }],
    },
    {
      code: "new RegExp(/a/, 'u');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: "new RegExp(/a/g, '');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: "new RegExp(/a/g, 'g');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: "new RegExp(/a/ig, 'g');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: "new RegExp(/a/g, 'ig');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: "new RegExp(/a/i, 'g');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: "new RegExp(/a/i, 'i');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: 'new RegExp(/a/, `u`);',
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: 'new RegExp(/a/, `gi`);',
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: "new RegExp('a');",
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRegExp', line: 1, column: 1 }],
    },
    {
      code: 'new RegExp(/a/, String.raw`u`);',
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: 'new RegExp(/a/ /* comment */);',
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExp', line: 1, column: 1 }],
    },
    {
      code: "new RegExp(/a/, 'd');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 1, column: 1 },
      ],
    },
    {
      code: '(a)\nnew RegExp(/b/);',
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExp', line: 2, column: 1 }],
    },
    {
      code: "(a)\nnew RegExp(/b/, 'g');",
      options: { disallowRedundantWrapping: true },
      errors: [
        { messageId: 'unexpectedRedundantRegExpWithFlags', line: 2, column: 1 },
      ],
    },
    {
      code: 'a/RegExp(/foo/);',
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExp', line: 1, column: 3 }],
    },
    {
      code: 'RegExp(/foo/)in a;',
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExp', line: 1, column: 1 }],
    },
    {
      code: 'new RegExp((String?.raw)`a`);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "new RegExp('+');", errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: "new RegExp('*');", errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: "RegExp('+');", errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: "RegExp('*');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "new RegExp('+', 'g');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('*', 'g');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "RegExp('+', 'g');", errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: "RegExp('*', 'g');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "RegExp('abc', 'u');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('abc', 'd');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('abc', 'd');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\\\\\\\', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "RegExp('\\n', '');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "RegExp('\\n\\n', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "RegExp('\\t', '');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "RegExp('\\t\\t', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\r\\n', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\u1234', 'g')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\u{1234}', 'g')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\u{11111}', 'g')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "RegExp('\\v', '');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "RegExp('\\v\\v', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "RegExp('\\f', '');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "RegExp('\\f\\f', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\\\b', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\\\b\\\\b', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('\\\\B\\\\b', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\\\w', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new globalThis.RegExp('\\\\W', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\\\s', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('\\\\S', '')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "globalThis.RegExp('\\\\d', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "globalThis.RegExp('\\\\D', '')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "globalThis.RegExp('\\\\\\\\\\\\D', '')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('\\\\D\\\\D', '')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new globalThis.RegExp('\\\\0\\\\0', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('\\\\0\\\\0', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('\\0\\0', 'g');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\\\0\\\\0\\\\0', '')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp('\\\\78\\\\126\\\\5934', '')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new window['RegExp']('\\\\x56\\\\x78\\\\x45', '');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "a in(RegExp('abc'))",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'x = y\n            RegExp("foo").test(x) ? bar() : baz()',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "func(new RegExp(String.raw`\\w{1, 2`, 'u'),new RegExp(String.raw`\\w{1, 2`, 'u'))",
      errors: [
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
      ],
    },
    {
      code: 'x = y;\n            RegExp("foo").test(x) ? bar() : baz()',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'typeof RegExp("foo")',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "RegExp(\"foo\") instanceof RegExp(String.raw`blahblah`, 'g') ? typeof new RegExp('(\\\\p{Emoji_Presentation})\\\\1', `ug`) : false",
      errors: [
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
      ],
    },
    {
      code: '[   new RegExp(`someregular`)]',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "const totallyValidatesEmails = new RegExp(\"\\\\S+@(\\\\S+\\\\.)+\\\\S+\")\n            if (typeof totallyValidatesEmails === 'object') {\n                runSomethingThatExists(Regexp('stuff'))\n            }",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "!new RegExp('^Hey, ', 'u') && new RegExp('jk$') && ~new RegExp('^Sup, ') || new RegExp('hi') + new RegExp('person') === -new RegExp('hi again') ? 5 * new RegExp('abc') : 'notregbutstring'",
      errors: [
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
      ],
    },
    {
      code: '#!/usr/bin/sh\n            RegExp("foo")',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'async function abc(){await new RegExp("foo")}',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'function* abc(){yield new RegExp("foo")}',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'function* abc(){yield* new RegExp("foo")}',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "console.log({ ...new RegExp('a') })",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "delete RegExp('a');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: "void RegExp('a');", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: 'new RegExp("\\\\S+@(\\\\S+\\\\.)+\\\\S+")**RegExp(\'a\')',
      errors: [
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
      ],
    },
    {
      code: 'new RegExp("\\\\S+@(\\\\S+\\\\.)+\\\\S+")%RegExp(\'a\')',
      errors: [
        { messageId: 'unexpectedRegExp' },
        { messageId: 'unexpectedRegExp' },
      ],
    },
    { code: "a in RegExp('abc')", errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: "\n            /abc/ == new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ === new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ != new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ !== new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ > new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ < new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ >= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ <= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ << new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ >> new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ >>> new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ ^ new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ & new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            /abc/ | new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            null ?? new RegExp('blah')\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc *= new RegExp('blah')\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            console.log({a: new RegExp('sup')})\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            console.log(() => {new RegExp('sup')})\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            function abc() {new RegExp('sup')}\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            function abc() {return new RegExp('sup')}\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc <<= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc >>= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc >>>= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc ^= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc &= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc |= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc ??= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc &&= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc ||= new RegExp('cba');\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc **= new RegExp('blah')\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc /= new RegExp('blah')\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc += new RegExp('blah')\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc -= new RegExp('blah')\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            abc %= new RegExp('blah')\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            () => new RegExp('blah')\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'a/RegExp("foo")in b',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'a/RegExp("foo")instanceof b',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'do RegExp("foo")\nwhile (true);',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "for(let i;i<5;i++) { break\nnew RegExp('search')}",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "for(let i;i<5;i++) { continue\nnew RegExp('search')}",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "\n            switch (value) {\n                case \"possibility\":\n                    console.log('possibility matched')\n                case RegExp('myReg').toString():\n                    console.log('matches a regexp\\' toString value')\n                    break;\n            }\n            ",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "throw new RegExp('abcdefg') // fail with a regular expression",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "for (value of new RegExp('something being searched')) { console.log(value) }",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "(async function(){for await (value of new RegExp('something being searched')) { console.log(value) }})()",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "for (value in new RegExp('something being searched')) { console.log(value) }",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "if (condition1 && condition2) new RegExp('avalue').test(str);",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "debugger\nnew RegExp('myReg')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    { code: 'RegExp("\\\\\\n")', errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: 'RegExp("\\\\\\t")', errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: 'RegExp("\\\\\\f")', errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: 'RegExp("\\\\\\v")', errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: 'RegExp("\\\\\\r")', errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: 'new RegExp("\t")', errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: 'new RegExp("/")', errors: [{ messageId: 'unexpectedRegExp' }] },
    { code: 'new RegExp("\\.")', errors: [{ messageId: 'unexpectedRegExp' }] },
    {
      code: 'new RegExp("\\\\.")',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp("\\\\\\n\\\\\\n")',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp("\\\\\\n\\\\\\f\\\\\\n")',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: 'new RegExp("\\u000A\\u000A");',
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('mysafereg' /* comment explaining its safety */)",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('[[A--B]]', 'v')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('[[A--B]]', 'v')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('[[A&&&]]', 'v')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('a', 'uv')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp(/a/, 'v')",
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExpWithFlags' }],
    },
    {
      code: "new RegExp(/a/, 'v')",
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExpWithFlags' }],
    },
    {
      code: "new RegExp(/a/g, 'v')",
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExpWithFlags' }],
    },
    {
      code: "new RegExp(/[[A--B]]/v, 'g')",
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExpWithFlags' }],
    },
    {
      code: "new RegExp(/a/u, 'v')",
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExpWithFlags' }],
    },
    {
      code: "new RegExp(/a/v, 'u')",
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExpWithFlags' }],
    },
    {
      code: "new RegExp(/[[A--B]]/v, 'u')",
      options: { disallowRedundantWrapping: true },
      errors: [{ messageId: 'unexpectedRedundantRegExpWithFlags' }],
    },
    {
      code: "new RegExp('(?i:foo)bar')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "new RegExp('(?i:foo)bar')",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
    {
      code: "var regex = new RegExp('foo', 'u');",
      errors: [{ messageId: 'unexpectedRegExp' }],
    },
  ],
});
