import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-template', {
  valid: [
    `'use strict';`,
    `var foo = 'foo' + '\\0';`,
    `var foo = 'bar';`,
    `var foo = 'bar' + 'baz';`,
    `var foo = foo + +'100';`,
    'var foo = `bar`;',
    'var foo = `hello, ${name}!`;',
    // https://github.com/eslint/eslint/issues/3507
    'var foo = `foo` + `bar` + "hoge";',
    'var foo = `foo` +\n    `bar` +\n    "hoge";',
  ],
  invalid: [
    {
      code: `var foo = 'hello, ' + name + '!';`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = bar + 'baz';`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: 'var foo = bar + `baz`;',
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = +100 + 'yen';`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = 'bar' + baz;`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = '\uffe5' + (n * 1000) + '-'`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = 'aaa' + aaa; var bar = 'bbb' + bbb;`,
      errors: [
        { messageId: 'unexpectedStringConcatenation' },
        { messageId: 'unexpectedStringConcatenation' },
      ],
    },
    {
      code: `var string = (number + 1) + 'px';`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = 'bar' + baz + 'qux';`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = '0 backslashes: \${bar}' + baz;`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = '1 backslash: \\\${bar}' + baz;`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = '2 backslashes: \\\\\${bar}' + baz;`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = '3 backslashes: \\\\\\\${bar}' + baz;`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "var foo = bar + 'this is a backtick: `' + baz;",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "var foo = bar + 'this is a backtick preceded by a backslash: \\`' + baz;",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "var foo = bar + 'this is a backtick preceded by two backslashes: \\\\`' + baz;",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: 'var foo = bar + `${baz}foo`;',
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code:
        "var foo = 'favorites: ' + favorites.map(f => {\n" +
        '    return f.name;\n' +
        "}) + ';';",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = bar + baz + 'qux';`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code:
        "var foo = 'favorites: ' +\n" +
        '    favorites.map(f => {\n' +
        '        return f.name;\n' +
        '    }) +\n' +
        "';';",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = /* a */ 'bar' /* b */ + /* c */ baz /* d */ + 'qux' /* e */ ;`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = bar + ('baz') + 'qux' + (boop);`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + 'unescapes an escaped single quote in a single-quoted string: \\''`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + "unescapes an escaped double quote in a double-quoted string: \\""`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + 'does not unescape an escaped double quote in a single-quoted string: \\"'`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + "does not unescape an escaped single quote in a double-quoted string: \\'"`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      // "\\x27" === "'"
      code: `foo + 'handles unicode escapes correctly: \\x27'`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + '\\\\033'`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + '\\0'`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    // https://github.com/eslint/eslint/issues/15083
    {
      code:
        `"default-src 'self' https://*.google.com;"\n` +
        `            + "frame-ancestors 'none';"\n` +
        `            + "report-to " + foo + ";"`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo + 'c' + 'd'`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b + c' + foo + 'd' + 'e'`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo + ('c' + 'd')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo + ('a' + 'b')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo + ('c' + 'd') + ('e' + 'f')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + ('a' + 'b') + ('c' + 'd')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + foo + ('b' + 'c') + ('d' + bar + 'e')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + ('b' + 'c') + ('d' + bar + 'e')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo + ('c' + 'd' + 'e')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo + ('c' + bar + 'd')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo + ('c' + bar + ('d' + 'e') + 'f')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + 'b' + foo + ('c' + bar + 'e') + 'f' + test`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + foo + ('b' + bar + 'c') + ('d' + test)`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + foo + ('b' + 'c') + ('d' + bar)`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `foo + ('a' + bar + 'b') + 'c' + test`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "'a' + '`b`' + c",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "'a' + '`b` + `c`' + d",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "'a' + b + ('`c`' + '`d`')",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "'`a`' + b + ('`c`' + '`d`')",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "foo + ('`a`' + bar + '`b`') + '`c`' + test",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + ('b' + 'c') + d`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "'a' + ('`b`' + '`c`') + d",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `a + ('b' + 'c') + d`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `a + ('b' + 'c') + (d + 'e')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "a + ('`b`' + '`c`') + d",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: "a + ('`b` + `c`' + '`d`') + e",
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + ('b' + 'c' + 'd') + e`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'a' + ('b' + 'c' + 'd' + (e + 'f') + 'g' +'h' + 'i') + j`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `a + (('b' + 'c') + 'd')`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `(a + 'b') + ('c' + 'd') + e`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `var foo = "Hello " + "world " + "another " + test`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `'Hello ' + '"world" ' + test`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
    {
      code: `"Hello " + "'world' " + test`,
      errors: [{ messageId: 'unexpectedStringConcatenation' }],
    },
  ],
});
