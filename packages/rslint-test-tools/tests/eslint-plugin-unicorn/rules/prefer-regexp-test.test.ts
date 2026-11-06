import { RuleTester } from '../rule-tester';

// Upstream: https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-regexp-test.js
const ruleTester = new RuleTester();

// upstream group 1
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'function foo(value: `a` | number) { if (/a/.exec(value)) {} }',
      filename: 'file.ts',
      output: 'function foo(value: `a` | number) { if (/a/.test(value)) {} }',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 49,
          suggestions: [],
        },
      ],
    },
    {
      code: 'function foo(value: `a` | RegExp) { if (/a/.exec(value)) {} }',
      filename: 'file.ts',
      output: 'function foo(value: `a` | RegExp) { if (/a/.test(value)) {} }',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 49,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream group 2
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'function foo() { if (string[0].match(/a/)) {} } const string = `abc`;',
      filename: 'case.js',
      output:
        'function foo() { if (/a/.test(string[0])) {} } const string = `abc`;',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream group 3
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    { code: 'if (object?.value.match(/./).length) {}', filename: 'case.js' },
    {
      code: 'if (object?.value.match(/./).length > 0) {}',
      filename: 'case.js',
    },
    {
      code: 'if (getObject()?.value.match(getPattern())?.length) {}',
      filename: 'case.js',
    },
    {
      code: 'if (object?.getValue().match(/./).length) {}',
      filename: 'case.js',
    },
    {
      code: 'if (object?.value.match(/./)?.length > 0) {}',
      filename: 'case.js',
    },
    { code: 'if (object?.value!.match(/./).length) {}', filename: 'file.ts' },
  ],
  invalid: [
    {
      code: 'if ((object?.value).match(/./).length) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if ((/./).test(object?.value)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (object?.regexp.exec(value).length) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (object?.regexp.test(value)) {}',
            },
          ],
        },
      ],
    },
  ],
});

// upstream group 4
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'text.search({regexp: /x/}.regexp) !== -1;',
      filename: 'case.js',
      output: '({regexp: /x/}.regexp).test(text);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
    },
    {
      code: 'text.search(({regexp: /x/}.regexp)) !== -1;',
      filename: 'case.js',
      output: '({regexp: /x/}.regexp).test((text));',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
    },
    {
      code: 'text.search({regexp: /x/}.regexp!) !== -1;',
      filename: 'file.ts',
      output: '({regexp: /x/}.regexp!).test(text);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream group 5
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: '/** @param {string} pattern */ function run(pattern) { if (text.match(pattern)) {} }',
      filename: 'file.ts',
    },
    {
      code: '/** @returns {string} */ function read() { return source; } if (text.search(read()) !== -1) {}',
      filename: 'file.ts',
    },
    {
      code: '/** @param {() => string} read */ function run(read) { if (text.match(read())) {} }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {string} regexp */ function run(regexp) { if (regexp.exec(text)) {} }',
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: '/** @param {string} text */ function run(text) { if (text.match(/x/)) {} }',
      filename: 'file.ts',
      output:
        '/** @param {string} text */ function run(text) { if (/x/.test(text)) {} }',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 54,
          endLine: 1,
          endColumn: 69,
          suggestions: [],
        },
      ],
    },
    {
      code: '/** @returns {string} */ function read() { return source; } if (read().search(/x/) !== -1) {}',
      filename: 'file.ts',
      output:
        '/** @returns {string} */ function read() { return source; } if (/x/.test(read())) {}',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 65,
          endLine: 1,
          endColumn: 83,
          suggestions: [],
        },
      ],
    },
    {
      code: '/** @param {() => string} read */ function run(read) { if (read().match(/x/g)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 78,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                '/** @param {() => string} read */ function run(read) { if (/x/g.test(read())) {} }',
            },
          ],
        },
      ],
    },
    {
      code: '/** @param {string} text */ function run(text) { return text.search(/x/) /* keep */ !== -1; }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 73,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream group 6
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: '/** @param {string} pattern */ function run(pattern) { if (text.match(pattern)) {} }',
      filename: 'file.ts',
    },
    {
      code: '/** @returns {string} */ function read() { return source; } if (text.search(read()) !== -1) {}',
      filename: 'file.ts',
    },
    {
      code: '/** @param {() => string} read */ function run(read) { if (text.match(read())) {} }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {string} regexp */ function run(regexp) { if (regexp.exec(text)) {} }',
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: '/** @param {string} text */ function run(text) { if (text.match(/x/)) {} }',
      filename: 'file.ts',
      output:
        '/** @param {string} text */ function run(text) { if (/x/.test(text)) {} }',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 54,
          endLine: 1,
          endColumn: 69,
          suggestions: [],
        },
      ],
    },
    {
      code: '/** @returns {string} */ function read() { return source; } if (read().search(/x/) !== -1) {}',
      filename: 'file.ts',
      output:
        '/** @returns {string} */ function read() { return source; } if (/x/.test(read())) {}',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 65,
          endLine: 1,
          endColumn: 83,
          suggestions: [],
        },
      ],
    },
    {
      code: '/** @param {() => string} read */ function run(read) { if (read().match(/x/g)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 78,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                '/** @param {() => string} read */ function run(read) { if (/x/g.test(read())) {} }',
            },
          ],
        },
      ],
    },
    {
      code: '/** @param {string} text */ function run(text) { return text.search(/x/) /* keep */ !== -1; }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 73,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream group 7
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: '/** @param {string} pattern */ function run(pattern) { if (text.match(pattern)) {} }',
      filename: 'file.ts',
    },
    {
      code: '/** @returns {string} */ function read() { return source; } if (text.search(read()) !== -1) {}',
      filename: 'file.ts',
    },
    {
      code: '/** @param {() => string} read */ function run(read) { if (text.match(read())) {} }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {string} regexp */ function run(regexp) { if (regexp.exec(text)) {} }',
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: '/** @param {string} text */ function run(text) { if (text.match(/x/)) {} }',
      filename: 'file.ts',
      output:
        '/** @param {string} text */ function run(text) { if (/x/.test(text)) {} }',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 54,
          endLine: 1,
          endColumn: 69,
          suggestions: [],
        },
      ],
    },
    {
      code: '/** @returns {string} */ function read() { return source; } if (read().search(/x/) !== -1) {}',
      filename: 'file.ts',
      output:
        '/** @returns {string} */ function read() { return source; } if (/x/.test(read())) {}',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 65,
          endLine: 1,
          endColumn: 83,
          suggestions: [],
        },
      ],
    },
    {
      code: '/** @param {() => string} read */ function run(read) { if (read().match(/x/g)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 78,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                '/** @param {() => string} read */ function run(read) { if (/x/g.test(read())) {} }',
            },
          ],
        },
      ],
    },
    {
      code: '/** @param {string} text */ function run(text) { return text.search(/x/) /* keep */ !== -1; }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 73,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream group 8
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: '/** @param {boolean} pattern */ function run(pattern: string) { if (text.match(pattern)) {} }',
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: '/** @param {string} pattern */ function run(pattern: RegExp) { if (text.match(pattern)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 68,
          endLine: 1,
          endColumn: 87,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                '/** @param {string} pattern */ function run(pattern: RegExp) { if (pattern.test(text)) {} }',
            },
          ],
        },
      ],
    },
  ],
});

// upstream group 9
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: '/** @param {boolean} pattern */ function run(pattern: string) { if (text.match(pattern)) {} }',
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: '/** @param {string} pattern */ function run(pattern: RegExp) { if (text.match(pattern)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 68,
          endLine: 1,
          endColumn: 87,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                '/** @param {string} pattern */ function run(pattern: RegExp) { if (pattern.test(text)) {} }',
            },
          ],
        },
      ],
    },
  ],
});

// upstream group 10
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const pattern = /a/; function run() { return"a".search(pattern) !== -1; }',
      filename: 'case.js',
      output:
        'const pattern = /a/; function run() { return pattern.test("a"); }',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 64,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const pattern = /a/; function run() { return"a".search(pattern) === -1; }',
      filename: 'case.js',
      output:
        'const pattern = /a/; function run() { return !pattern.test("a"); }',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 64,
          suggestions: [],
        },
      ],
    },
    {
      code: 'foo\n"a".search(/a/) === -1',
      filename: 'case.js',
      output: 'foo\n;!/a/.test("a")',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
          suggestions: [],
        },
      ],
    },
    {
      code: 'function run() { return"a".match(/a/) ? true : false; }',
      filename: 'case.js',
      output: 'function run() { return /a/.test("a") ? true : false; }',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
    },
    {
      code: 'for (const value of"a".match(/a/) ? [1] : [2]);',
      filename: 'case.js',
      output: 'for (const value of /a/.test("a") ? [1] : [2]);',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
    },
    {
      code: 'function run(pattern: RegExp) { return"a".search(pattern) !== -1; }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 58,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function run(pattern: RegExp) { return pattern.test("a"); }',
            },
          ],
        },
      ],
    },
  ],
});

// upstream group 11
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    { code: 'const bar = !re.test(foo)', filename: 'case.js' },
    { code: 'const matches = foo.match(re) || []', filename: 'case.js' },
    { code: 'const matches = foo.match(re)', filename: 'case.js' },
    { code: 'const matches = re.exec(foo)', filename: 'case.js' },
    { code: 'string.search(regex)', filename: 'case.js' },
    { code: 'const index = string.search(regex)', filename: 'case.js' },
    { code: 'if (string.search(regex)) {}', filename: 'case.js' },
    { code: 'while (foo = re.exec(bar)) {}', filename: 'case.js' },
    { code: 'while ((foo = re.exec(bar))) {}', filename: 'case.js' },
    { code: 'if (foo.notMatch(re)) {}', filename: 'case.js' },
    { code: 'if (re.notExec(foo)) {}', filename: 'case.js' },
    { code: 'if (foo.match) {}', filename: 'case.js' },
    { code: 'if (re.exec) {}', filename: 'case.js' },
    { code: 'if (foo[match](re)) {}', filename: 'case.js' },
    { code: 'if (re[exec](foo)) {}', filename: 'case.js' },
    { code: 'if (foo["match"](re)) {}', filename: 'case.js' },
    { code: 'if (re["exec"](foo)) {}', filename: 'case.js' },
    { code: 'if (match(re)) {}', filename: 'case.js' },
    { code: 'if (exec(foo)) {}', filename: 'case.js' },
    { code: 'if (foo.match()) {}', filename: 'case.js' },
    { code: 'if (re.exec()) {}', filename: 'case.js' },
    { code: 'if (foo.search() !== -1) {}', filename: 'case.js' },
    { code: 'if (foo.match(re, another)) {}', filename: 'case.js' },
    { code: 'if (re.exec(foo, another)) {}', filename: 'case.js' },
    { code: 'if (foo.search(re, another) !== -1) {}', filename: 'case.js' },
    { code: 'if (foo.match(...[regexp])) {}', filename: 'case.js' },
    { code: 'if (re.exec(...[string])) {}', filename: 'case.js' },
    { code: 'if (foo.search(...[regexp]) !== -1) {}', filename: 'case.js' },
    { code: 'if (foo.match(1)) {}', filename: 'case.js' },
    { code: 'if (foo.match("1")) {}', filename: 'case.js' },
    { code: 'if (foo.match(null)) {}', filename: 'case.js' },
    { code: 'if (foo.match(undefined)) {}', filename: 'case.js' },
    { code: 'if (foo.match(void 0)) {}', filename: 'case.js' },
    {
      code: 'const pattern = undefined; if (foo.match(pattern)) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = void 0; if (foo.match(pattern)) {}',
      filename: 'case.js',
    },
    { code: 'if (foo.match(1n)) {}', filename: 'case.js' },
    { code: 'if (foo.match(true)) {}', filename: 'case.js' },
    { code: 'if (foo.search(undefined) !== -1) {}', filename: 'case.js' },
    { code: 'if (foo.search(void 0) !== -1) {}', filename: 'case.js' },
    { code: 'if (foo.search("1") !== -1) {}', filename: 'case.js' },
    { code: 'if (foo.search(`1`) !== -1) {}', filename: 'case.js' },
    { code: 'if (foo.search(`${pattern}`) !== -1) {}', filename: 'case.js' },
    {
      code: 'const pattern = "1"; if (foo.search(pattern) !== -1) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = `1`; if (foo.search(pattern) !== -1) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = `${value}`;\nif (foo.search(pattern) !== -1) {}',
      filename: 'case.js',
    },
    {
      code: 'type Pattern = string; function foo(pattern: Pattern) { return text.search(pattern) !== -1; }',
      filename: 'file.ts',
    },
    {
      code: 'const pattern = 1; if (foo.search(pattern) !== -1) {}',
      filename: 'case.js',
    },
    { code: 'if (foo.search(1 + 2) !== -1) {}', filename: 'case.js' },
    { code: 'if (foo.search(/re/) <= -1) {}', filename: 'case.js' },
    { code: 'if (-1 !== foo.search(/re/)) {}', filename: 'case.js' },
    { code: 'if (-1 === foo.search(/re/)) {}', filename: 'case.js' },
    { code: 'if (0 <= foo.search(/re/)) {}', filename: 'case.js' },
    { code: 'if (0 > foo.search(/re/)) {}', filename: 'case.js' },
    { code: 'if (({}).match(action)) {}', filename: 'case.js' },
    { code: 'if ((() => {}).match(action)) {}', filename: 'case.js' },
    { code: 'if (/string/.match(/regexp/)) {}', filename: 'case.js' },
    { code: 'if (/string/.search(/regexp/) !== -1) {}', filename: 'case.js' },
    {
      code: 'const actionCreator = {}; if (actionCreator.match(action)) {}',
      filename: 'case.js',
    },
    {
      code: 'const actionCreator = () => {}; if (actionCreator.match(action)) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = "1"; if (foo.match(pattern)) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = `1`; if (foo.match(pattern)) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = `${value}`;\nif (foo.match(pattern)) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = {}; if (foo.match(pattern)) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = () => {}; if (foo.match(pattern)) {}',
      filename: 'case.js',
    },
    {
      code: 'const pattern = 1 + 2; if (foo.match(pattern)) {}',
      filename: 'case.js',
    },
    {
      code: "const object = {pattern: /regexp/};\nObject.defineProperty(object, 'pattern', {get() { return 'regexp'; }});\nif (foo.match(object.pattern)) {}",
      filename: 'case.js',
    },
    { code: 'if ("regexp".exec(foo)) {}', filename: 'case.js' },
    {
      code: 'const regexp = "regexp"; if (regexp.exec(foo)) {}',
      filename: 'case.js',
    },
    {
      code: 'const regexp = {}; if (regexp.exec(foo)) {}',
      filename: 'case.js',
    },
    {
      code: 'const regexp = () => {}; if (regexp.exec(foo)) {}',
      filename: 'case.js',
    },
    { code: 'if (regexp.exec({})) {}', filename: 'case.js' },
    { code: 'if (regexp.exec(/string/)) {}', filename: 'case.js' },
    { code: 'if (regexp.exec(undefined)) {}', filename: 'case.js' },
    { code: 'if (regexp.exec(void 0)) {}', filename: 'case.js' },
    {
      code: 'const value = {}; if (regexp.exec(value)) {}',
      filename: 'case.js',
    },
    {
      code: 'const value = () => {}; if (regexp.exec(value)) {}',
      filename: 'case.js',
    },
    { code: 'if (uri.match(/unicorn/).length >= 1) {}', filename: 'case.js' },
    { code: 'if (uri.match(/unicorn/).length === 0) {}', filename: 'case.js' },
    { code: 'if (0 < uri.match(/unicorn/).length) {}', filename: 'case.js' },
    {
      code: 'const Boolean = value => value; const re = /a/; const bar = Boolean(foo.match(re))',
      filename: 'case.js',
    },
    {
      code: 'const Boolean = value => value; const isNotEmpty = Boolean(uri.match(/unicorn/).length)',
      filename: 'case.js',
    },
    {
      code: 'function unicorn(Boolean) { const re = /a/; if (Boolean(foo.match(re))) {} }',
      filename: 'case.js',
    },
    { code: 'if (uri.match?.(/unicorn/)?.length) {}', filename: 'case.js' },
    { code: 'if (!uri.match(/unicorn/).length) {}', filename: 'case.js' },
    { code: 'if (!uri.match(/unicorn/)?.length) {}', filename: 'case.js' },
    { code: 'if (!uri.match(/unicorn/).length > 0) {}', filename: 'case.js' },
    { code: 'if (!(uri.match(/unicorn/).length > 0)) {}', filename: 'case.js' },
    {
      code: 'if (!(foo || uri.match(/unicorn/).length)) {}',
      filename: 'case.js',
    },
    {
      code: 'if (!(foo || uri.match(/unicorn/).length > 0)) {}',
      filename: 'case.js',
    },
    {
      code: 'if (!Boolean(uri.match(/unicorn/).length)) {}',
      filename: 'case.js',
    },
    {
      code: 'if (!Boolean(uri.match(/unicorn/).length > 0)) {}',
      filename: 'case.js',
    },
    {
      code: 'if (!Boolean(foo || uri.match(/unicorn/).length)) {}',
      filename: 'case.js',
    },
    {
      code: 'if (foo.match(a ? (b ? "x" : null) : "y")) {}',
      filename: 'case.js',
    },
    {
      code: 'function foo(pattern: 1) { if ("string".match(pattern)) {} }',
      filename: 'file.ts',
    },
    {
      code: 'function foo(pattern: "a" | 1) { if ("string".match(pattern)) {} }',
      filename: 'file.ts',
    },
    {
      code: 'function foo(regexp: String, value: string) { if (regexp.exec(value)) {} }',
      filename: 'file.ts',
    },
    {
      code: 'function foo(regexp: (string & {brand: 1}) | number, value: string) { if (regexp.exec(value)) {} }',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value: RegExp & {brand: 1}) { if (value.match(/a/)) {} }',
      filename: 'file.ts',
    },
    {
      code: 'function foo(pattern) { if ("string".match(<number>pattern)) {} }',
      filename: 'file.ts',
    },
    {
      code: 'function foo(pattern) { if ("string".match(pattern satisfies number)) {} }',
      filename: 'file.ts',
    },
    {
      code: 'function foo(regexp: String | "a", value: string) { if (regexp.exec(value)) {} }',
      filename: 'file.ts',
    },
    { code: 'if (foo.match(a ? 1 : null)) {}', filename: 'case.js' },
  ],
  invalid: [
    {
      code: 'const re = /a/; const bar = !foo.match(re)',
      filename: 'case.js',
      output: 'const re = /a/; const bar = !re.test(foo)',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = Boolean(foo.match(re))',
      filename: 'case.js',
      output: 'const re = /a/; const bar = Boolean(re.test(foo))',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 50,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = !!Boolean(foo.match(re))',
      filename: 'case.js',
      output: 'const re = /a/; const bar = !!Boolean(re.test(foo))',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 52,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; if (foo.match(re)) {}',
      filename: 'case.js',
      output: 'const re = /a/; if (re.test(foo)) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = foo.match(re) ? 1 : 2',
      filename: 'case.js',
      output: 'const re = /a/; const bar = re.test(foo) ? 1 : 2',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; while (foo.match(re)) foo = foo.slice(1);',
      filename: 'case.js',
      output: 'const re = /a/; while (re.test(foo)) foo = foo.slice(1);',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; do {foo = foo.slice(1)} while (foo.match(re));',
      filename: 'case.js',
      output: 'const re = /a/; do {foo = foo.slice(1)} while (re.test(foo));',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 61,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; for (; foo.match(re); ) foo = foo.slice(1);',
      filename: 'case.js',
      output: 'const re = /a/; for (; re.test(foo); ) foo = foo.slice(1);',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (object.match(/regexp/)) {}',
      filename: 'case.js',
      output: 'if (/regexp/.test(object)) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const pattern = /regexp/; if (object.match(pattern)) {}',
      filename: 'case.js',
      output: 'const pattern = /regexp/; if (pattern.test(object)) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 52,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const pattern = /regexp/g; if (object.match(pattern)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const pattern = /regexp/g; if (pattern.test(object)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (uri.match(/unicorn/).length) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/unicorn/.test(uri)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (uri.match(/unicorn/).length > 0) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/unicorn/.test(uri)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (Boolean(uri.match(/unicorn/).length)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (Boolean(/unicorn/.test(uri))) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (uri.match(/unicorn/)?.length) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/unicorn/.test(uri)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (uri.match(/unicorn/)?.length > 0) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/unicorn/.test(uri)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (uri.match(/unicorn/).length /* keep */ > 0) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 25,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (\n\turi.match(/unicorn/).length ||\n\turi.match(/unicorn/).length > 0 ||\n\turi.match(/unicorn/)?.length ||\n\turi.match(/unicorn/)?.length > 0\n) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'if (\n\t/unicorn/.test(uri) ||\n\turi.match(/unicorn/).length > 0 ||\n\turi.match(/unicorn/)?.length ||\n\turi.match(/unicorn/)?.length > 0\n) {}',
            },
          ],
        },
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'if (\n\turi.match(/unicorn/).length ||\n\t/unicorn/.test(uri) ||\n\turi.match(/unicorn/)?.length ||\n\turi.match(/unicorn/)?.length > 0\n) {}',
            },
          ],
        },
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 4,
          column: 2,
          endLine: 4,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'if (\n\turi.match(/unicorn/).length ||\n\turi.match(/unicorn/).length > 0 ||\n\t/unicorn/.test(uri) ||\n\turi.match(/unicorn/)?.length > 0\n) {}',
            },
          ],
        },
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 5,
          column: 2,
          endLine: 5,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'if (\n\turi.match(/unicorn/).length ||\n\turi.match(/unicorn/).length > 0 ||\n\turi.match(/unicorn/)?.length ||\n\t/unicorn/.test(uri)\n) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = foo.search(re) !== -1',
      filename: 'case.js',
      output: 'const re = /a/; const bar = re.test(foo)',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = foo.search(re) != -1',
      filename: 'case.js',
      output: 'const re = /a/; const bar = re.test(foo)',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = foo.search(re) > -1',
      filename: 'case.js',
      output: 'const re = /a/; const bar = re.test(foo)',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = foo.search(re) >= 0',
      filename: 'case.js',
      output: 'const re = /a/; const bar = re.test(foo)',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = foo.search(re) === -1',
      filename: 'case.js',
      output: 'const re = /a/; const bar = !re.test(foo)',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = foo.search(re) == -1',
      filename: 'case.js',
      output: 'const re = /a/; const bar = !re.test(foo)',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = foo.search(re) < 0',
      filename: 'case.js',
      output: 'const re = /a/; const bar = !re.test(foo)',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (foo.search(/re/) !== -1) {}',
      filename: 'case.js',
      output: 'if (/re/.test(foo)) {}',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if ((foo).search(/re/) !== -1) {}',
      filename: 'case.js',
      output: 'if ((/re/).test(foo)) {}',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if ((foo.search(/re/)) === -1) {}',
      filename: 'case.js',
      output: 'if (!(/re/.test(foo))) {}',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (foo.search(bar.baz()) !== -1) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (bar.baz().test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.search(new RegExp("re", "g")) !== -1) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 38,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (new RegExp("re", "g").test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.search(unknown) !== -1) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (unknown.test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.search(/re/) /* keep */ !== -1) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (foo.search(/re/) /* keep */ === -1) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const regex = /weird/g;\nif (foo.search(regex) !== -1);',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const regex = /weird/g;\nif (regex.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'const regex = /weird/g;\nif (foo.search(regex) === -1);',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 22,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const regex = /weird/g;\nif (!regex.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'const re = /a/y;\nif (foo.search(re) !== -1);',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const re = /a/y;\nif (re.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = !re.exec(foo)',
      filename: 'case.js',
      output: 'const re = /a/; const bar = !re.test(foo)',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = Boolean(re.exec(foo))',
      filename: 'case.js',
      output: 'const re = /a/; const bar = Boolean(re.test(foo))',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 44,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; if (re.exec(foo)) {}',
      filename: 'case.js',
      output: 'const re = /a/; if (re.test(foo)) {}',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; const bar = re.exec(foo) ? 1 : 2',
      filename: 'case.js',
      output: 'const re = /a/; const bar = re.test(foo) ? 1 : 2',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; while (re.exec(foo)) foo = foo.slice(1);',
      filename: 'case.js',
      output: 'const re = /a/; while (re.test(foo)) foo = foo.slice(1);',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; do {foo = foo.slice(1)} while (re.exec(foo));',
      filename: 'case.js',
      output: 'const re = /a/; do {foo = foo.slice(1)} while (re.test(foo));',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 55,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; for (; re.exec(foo); ) foo = foo.slice(1);',
      filename: 'case.js',
      output: 'const re = /a/; for (; re.test(foo); ) foo = foo.slice(1);',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (/unicorn/.exec(uri).length) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/unicorn/.test(uri)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (/unicorn/.exec(uri)?.length > 0) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/unicorn/.test(uri)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'const re = /a/; if ((0, foo).match(re)) {}',
      filename: 'case.js',
      output: 'const re = /a/; if ((re).test((0, foo))) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 39,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; if ((0, foo).match((re))) {}',
      filename: 'case.js',
      output: 'const re = /a/; if ((re).test((0, foo))) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 41,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; if ((foo).match(re)) {}',
      filename: 'case.js',
      output: 'const re = /a/; if ((re).test(foo)) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /a/; if ((foo).match((re))) {}',
      filename: 'case.js',
      output: 'const re = /a/; if ((re).test((foo))) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (foo.match(/re/)) {}',
      filename: 'case.js',
      output: 'if (/re/.test(foo)) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const bar = {bar: /a/}; if (foo.match(bar.baz)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 47,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const bar = {bar: /a/}; if (bar.baz.test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(bar.baz())) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (bar.baz().test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(new RegExp("re", "g"))) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (new RegExp("re", "g").test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: "const options = {toString() { sideEffect(); return 'a'; }};\nif (getString().match(new RegExp(options))) {}",
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                "const options = {toString() { sideEffect(); return 'a'; }};\nif (new RegExp(options).test(getString())) {}",
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(new SomeRegExp())) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (new SomeRegExp().test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(new SomeRegExp)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if ((new SomeRegExp).test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(bar?.baz)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (bar?.baz.test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(bar?.baz())) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (bar?.baz().test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(bar || baz)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if ((bar || baz).test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'async function a() {\n\tif (foo.match(await bar())) {}\n}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'async function a() {\n\tif ((await bar()).test(foo)) {}\n}',
            },
          ],
        },
      ],
    },
    {
      code: 'if ((foo).match(/re/)) {}',
      filename: 'case.js',
      output: 'if ((/re/).test(foo)) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if ((foo).match(new SomeRegExp)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 32,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if ((new SomeRegExp).test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if ((foo).match(bar?.baz)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if ((bar?.baz).test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if ((foo).match(bar?.baz())) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if ((bar?.baz()).test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'const bar = false; const baz = /a/; if ((foo).match(bar || baz)) {}',
      filename: 'case.js',
      output:
        'const bar = false; const baz = /a/; if ((bar || baz).test(foo)) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 64,
          suggestions: [],
        },
      ],
    },
    {
      code: 'async function a() {\n\tif ((foo).match(await bar())) {}\n}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'async function a() {\n\tif ((await bar()).test(foo)) {}\n}',
            },
          ],
        },
      ],
    },
    {
      code: 'const re = [/a/]; if (foo.match([re][0])) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 41,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const re = [/a/]; if ([re][0].test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'async function a() {\n\tif (\n\t\t/* 1 */ foo() /* 2 */\n\t\t\t./* 3 */ match /* 4 */ (\n\t\t\t\t/* 5 */ await /* 6 */ bar() /* 7 */\n\t\t\t\t,\n\t\t\t\t/* 8 */\n\t\t\t)\n\t) {}\n}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 3,
          column: 11,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'async function a() {\n\tif (\n\t\t/* 1 */ (await /* 6 */ bar()) /* 2 */\n\t\t\t./* 3 */ test /* 4 */ (\n\t\t\t\t/* 5 */ foo() /* 7 */\n\t\t\t\t,\n\t\t\t\t/* 8 */\n\t\t\t)\n\t) {}\n}',
            },
          ],
        },
      ],
    },
    {
      code: "const regex = new RegExp('[.!?]\\s*$');\nif (foo.match(regex)) {}",
      filename: 'case.js',
      output:
        "const regex = new RegExp('[.!?]\\s*$');\nif (regex.test(foo)) {}",
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 21,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (foo.match(unknown)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 23,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (unknown.test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(/a/g));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/a/g.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(/a/y));',
      filename: 'case.js',
      output: 'if (/a/y.test(foo));',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (foo.match(/a/gy));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/a/gy.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(/a/ig));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/a/ig.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(new RegExp("a", "g")));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (new RegExp("a", "g").test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'if (/a/g.exec(foo));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/a/g.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'if (/a/y.exec(foo));',
      filename: 'case.js',
      output: 'if (/a/y.test(foo));',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 14,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (/a/gy.exec(foo));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/a/gy.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'if (/a/yi.exec(foo));',
      filename: 'case.js',
      output: 'if (/a/yi.test(foo));',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (new RegExp("a", "g").exec(foo));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (new RegExp("a", "g").test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'if (new RegExp("a", "y").exec(foo));',
      filename: 'case.js',
      output: 'if (new RegExp("a", "y").test(foo));',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const regex = /weird/g;\nif (foo.match(regex));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 21,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const regex = /weird/g;\nif (regex.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'const regex = /weird/g;\nif (regex.exec(foo));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const regex = /weird/g;\nif (regex.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'const regex = /weird/y;\nif (regex.exec(foo));',
      filename: 'case.js',
      output: 'const regex = /weird/y;\nif (regex.test(foo));',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const regex = /weird/gyi;\nif (regex.exec(foo));',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 15,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const regex = /weird/gyi;\nif (regex.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: "let re = new RegExp('foo', 'g');\nif(str.match(re));",
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: "let re = new RegExp('foo', 'g');\nif(re.test(str));",
            },
          ],
        },
      ],
    },
    {
      code: '!/a/u.exec(foo)',
      filename: 'case.js',
      output: '!/a/u.test(foo)',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 11,
          suggestions: [],
        },
      ],
    },
    {
      code: '!/a/v.exec(foo)',
      filename: 'case.js',
      output: '!/a/v.test(foo)',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 11,
          suggestions: [],
        },
      ],
    },
    {
      code: 'function foo() { if (string[0].match(/a/)) {} } const string = "abc";',
      filename: 'case.js',
      output:
        'function foo() { if (/a/.test(string[0])) {} } const string = "abc";',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (foo.match(a ? (b ? /x/ : null) : /y/)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if ((a ? (b ? /x/ : null) : /y/).test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.match(a ? /x/ : /y/)) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 29,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if ((a ? /x/ : /y/).test(foo)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(pattern: string | Foo) { if ("string".match(pattern)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 66,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern: string | Foo) { if (pattern.test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(pattern: Foo & RegExp) { if ("string".match(pattern)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 66,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern: Foo & RegExp) { if (pattern.test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(pattern: ({a: 1} & {b: 1}) | RegExp) { if ("string".match(pattern)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 80,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern: ({a: 1} & {b: 1}) | RegExp) { if (pattern.test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(pattern: Foo.Bar) { if ("string".match(pattern)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 61,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern: Foo.Bar) { if (pattern.test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(pattern) { if ("string".match(pattern as RegExp)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 62,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern) { if ((pattern as RegExp).test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(pattern) { if ("string".match(pattern as Foo)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 59,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern) { if ((pattern as Foo).test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(pattern: RegExp | undefined) { if ("string".match(pattern!)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 73,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern: RegExp | undefined) { if (pattern!.test("string")) {} }',
            },
          ],
        },
      ],
    },
  ],
});

// upstream group 12
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: 'function stringSearchPattern(pattern: string) {\n\tif ("string".search(pattern) !== -1) {}\n}\n\nfunction stringMatchPattern(pattern: string) {\n\tif ("string".match(pattern)) {}\n}\n\nfunction stringSearchAssertionPattern(pattern: string) {\n\tif ("string".search((`${pattern}` as string)) !== -1) {}\n}\n\nfunction stringMatchAssertionPattern(pattern: string) {\n\tif ("string".match((`${pattern}` as string))) {}\n}\n\nfunction actionCreatorMatcher(actionCreator: {match(action: unknown): boolean}, action: unknown) {\n\tif (actionCreator.match(action)) {}\n}\n\nfunction functionActionCreatorMatcher(actionCreator: (() => void) & {match(action: unknown): boolean}, action: unknown) {\n\tif (actionCreator.match(action)) {}\n}\n\nfunction stringExecReceiver(regexp: string, value: string) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction stringOrObjectExecReceiver(regexp: string | object, value: string) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction objectExecReceiver(regexp: {exec(value: string): unknown}, value: string) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction objectExecArgument(regexp: RegExp, value: object) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction regexpOrObjectExecArgument(regexp: RegExp, value: RegExp | object) {\n\tif (regexp.exec(value)) {}\n}\n\nfunction objectMatchReceiver(value: object, pattern: RegExp) {\n\tif (value.match(pattern)) {}\n}\n\nfunction stringOrNumberMatchPattern(pattern: string | number) {\n\tif ("string".match(pattern)) {}\n}\n\nfunction stringOrRegExpMatchPattern(pattern: string | RegExp) {\n\tif ("string".match(pattern)) {}\n}\n\nfunction stringOrUndefinedSearchPattern(pattern: string | undefined) {\n\tif ("string".search(pattern) !== -1) {}\n}\n\nfunction arrayMatchPattern(pattern: string[]) {\n\tif ("string".match(pattern)) {}\n}\n\nfunction tupleSearchPattern(pattern: [string]) {\n\tif ("string".search(pattern) !== -1) {}\n}\n\nfunction genericArrayMatchPattern(pattern: Array<string>) {\n\tif ("string".match(pattern)) {}\n}\n\nfunction genericReadonlyArraySearchPattern(pattern: ReadonlyArray<string>) {\n\tif ("string".search(pattern) !== -1) {}\n}',
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: 'function foo(pattern: unknown) { if ("string".match(pattern)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 61,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern: unknown) { if (pattern.test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(pattern: Action) { if ("string".match(pattern)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 60,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(pattern: Action) { if (pattern.test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo(regexp: RegExp, value: string) { if (regexp.exec(value)) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 58,
          endLine: 1,
          endColumn: 62,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo(regexp: RegExp, value: string) { if (regexp.test(value)) {} }',
            },
          ],
        },
      ],
    },
  ],
});

// upstream group 13
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: "type UnknownAction = {\n\ttype: string;\n};\n\ntype ActionCreatorWithMatcher<Type extends string> = (() => {type: Type}) & {\n\tmatch(action: UnknownAction): action is {type: Type};\n};\n\ndeclare function createAction<Type extends string>(type: Type): ActionCreatorWithMatcher<Type>;\n\nconst increment = createAction('counter/increment');\ndeclare const action: UnknownAction;\n\nif (increment.match(action)) {}",
      filename: 'file.ts',
    },
    {
      code: 'declare function getPattern(): string | number;\nif ("string".match(getPattern())) {}',
      filename: 'file.ts',
    },
  ],
  invalid: [
    {
      code: 'declare const pattern: any;\n\nif ("string".match(pattern)) {}',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'declare const pattern: any;\n\nif (pattern.test("string")) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'declare const pattern: unknown;\n\nif ("string".match(pattern)) {}',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'declare const pattern: unknown;\n\nif (pattern.test("string")) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'declare function getPattern(): RegExp | number; if ("string".match(getPattern())) {}',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 81,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'declare function getPattern(): RegExp | number; if (getPattern().test("string")) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'declare class Pattern extends RegExp {} declare function getPattern(): Pattern; if ("string".match(getPattern())) {}',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 85,
          endLine: 1,
          endColumn: 113,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'declare class Pattern extends RegExp {} declare function getPattern(): Pattern; if (getPattern().test("string")) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo<Type extends RegExp>(getPattern: () => Type | null) { if ("string".match(getPattern())) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 72,
          endLine: 1,
          endColumn: 100,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo<Type extends RegExp>(getPattern: () => Type | null) { if (getPattern().test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo<Type>(getPattern: () => Type) { if ("string".match(getPattern())) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 50,
          endLine: 1,
          endColumn: 78,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo<Type>(getPattern: () => Type) { if (getPattern().test("string")) {} }',
            },
          ],
        },
      ],
    },
    {
      code: 'function foo<Type>(getPattern: () => NonNullable<Type>) { if ("string".match(getPattern())) {} }',
      filename: 'file.ts',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 63,
          endLine: 1,
          endColumn: 91,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output:
                'function foo<Type>(getPattern: () => NonNullable<Type>) { if (getPattern().test("string")) {} }',
            },
          ],
        },
      ],
    },
  ],
});

// upstream group 14
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const re = /a/v; !re.exec(foo)',
      filename: 'case.js',
      output: 'const re = /a/v; !re.test(foo)',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = new RegExp("a", "v"); !re.exec(foo)',
      filename: 'case.js',
      output: 'const re = new RegExp("a", "v"); !re.test(foo)',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream group 15
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'const re = /a/y; if (foo.search(re) !== -1);',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const re = /a/y; if (re.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'const re = new RegExp("a", "y"); if (foo.search(re) !== -1);',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 52,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'const re = new RegExp("a", "y"); if (re.test(foo));',
            },
          ],
        },
      ],
    },
    {
      code: 'if (foo.search(new RegExp("a", "y")) !== -1);',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 37,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (new RegExp("a", "y").test(foo));',
            },
          ],
        },
      ],
    },
  ],
});

// upstream group 16
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'a\nfoo.search(/b/) >= 0;',
      filename: 'case.js',
      output: 'a\n;/b/.test(foo);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
          suggestions: [],
        },
      ],
    },
    {
      code: 'foo()\nfoo.search(/b/) > -1;',
      filename: 'case.js',
      output: 'foo()\n;/b/.test(foo);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (q) foo()\nfoo.match(/b/) ? 1 : 2;',
      filename: 'case.js',
      output: 'if (q) foo()\n;/b/.test(foo) ? 1 : 2;',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const re = /b/;\nfoo.search(re) >= 0;',
      filename: 'case.js',
      output: 'const re = /b/;\nre.test(foo);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 15,
          suggestions: [],
        },
      ],
    },
  ],
});

// upstream group 17
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: 'switch (foo.match(/b/)) { case null: break; }',
      filename: 'case.js',
    },
    {
      code: 'switch (foo.search(/b/)) { case -1: break; }',
      filename: 'case.js',
    },
    {
      code: 'switch (foo.match(/b/)) { case foo.match(/b/): break; }',
      filename: 'case.js',
    },
  ],
  invalid: [
    {
      code: 'switch (foo.match(/b/) ? 1 : 2) { case 1: break; }',
      filename: 'case.js',
      output: 'switch (/b/.test(foo) ? 1 : 2) { case 1: break; }',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
    },
  ],
});

// Upstream runtime tests
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [
    {
      code: 'const object = null; let calls = 0; function getPattern() {calls++; return /./;} [Boolean(object?.value.match(getPattern()).length), calls];',
      filename: 'case.js',
    },
    {
      code: 'const object = undefined; let calls = 0; function getPattern() {calls++; return /./;} [Boolean(object?.value.match(getPattern()).length), calls];',
      filename: 'case.js',
    },
    {
      code: 'const object = {value: "x"}; let calls = 0; function getPattern() {calls++; return /./;} [Boolean(object?.value.match(getPattern()).length), calls];',
      filename: 'case.js',
    },
  ],
  invalid: [
    {
      code: 'a;\nfoo.search(/b/) >= 0;',
      filename: 'case.js',
      output: 'a;\n/b/.test(foo);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
          suggestions: [],
        },
      ],
    },
    {
      code: '  foo.search(/b/) >= 0;',
      filename: 'case.js',
      output: '  /b/.test(foo);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
    },
    {
      code: 'a\r\nfoo.search(/b/) >= 0;',
      filename: 'case.js',
      output: 'a\r\n;/b/.test(foo);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const a = 1; foo.search(/b/) >= 0;',
      filename: 'case.js',
      output: 'const a = 1; /b/.test(foo);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (q) foo();\nfoo.search(/b/) >= 0;',
      filename: 'case.js',
      output: 'if (q) foo();\n/b/.test(foo);',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 16,
          suggestions: [],
        },
      ],
    },
  ],
});

// Documentation examples
ruleTester.run('prefer-regexp-test', {} as never, {
  valid: [{ code: 'if (/unicorn/.test(string)) {}', filename: 'case.js' }],
  invalid: [
    {
      code: 'if (string.match(/unicorn/)) {}',
      filename: 'case.js',
      output: 'if (/unicorn/.test(string)) {}',
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (/unicorn/.exec(string)) {}',
      filename: 'case.js',
      output: 'if (/unicorn/.test(string)) {}',
      errors: [
        {
          messageId: 'regexp-exec',
          message: 'Prefer `.test(…)` over `.exec(…)`.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
    },
    {
      code: 'if (string.match(/unicorn/).length > 0) {}',
      filename: 'case.js',
      output: null,
      errors: [
        {
          messageId: 'string-match',
          message: 'Prefer `RegExp#test(…)` over `String#match(…)`.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 28,
          suggestions: [
            {
              messageId: 'suggestion',
              desc: 'Switch to `RegExp#test(…)`.',
              output: 'if (/unicorn/.test(string)) {}',
            },
          ],
        },
      ],
    },
    {
      code: 'if (string.search(/unicorn/) !== -1) {}',
      filename: 'case.js',
      output: 'if (/unicorn/.test(string)) {}',
      errors: [
        {
          messageId: 'string-search',
          message:
            'Prefer `RegExp#test(…)` over `String#search(…)` when checking for existence.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
    },
  ],
});

// Native parsing does not support Vue/Svelte templates. Retain the upstream inputs as explicit skips.
test.skip('Vue template: <template><div v-if="/re/.exec(string)"></div></template>', () => {});
test.skip('Vue template: <template><div v-if="\'string\'.match(/re/)"></div></template>', () => {});
test.skip('Vue template: <template><div v-if="\'string\'.search(/re/) !== -1"></div></template>', () => {});
test.skip('Vue template: <script>if(/re/.exec(string));</script>', () => {});
test.skip('Vue template: <template><button @click="setup()\nvalue.search(/x/) !== -1"></button></template>', () => {});
test.skip('Vue template: <template><button @click="setup()\r\nvalue.search(/x/) !== -1"></button></template>', () => {});
test.skip('Vue template: <template><button @click="setup()\nvalue.match(/x/) ? found() : missing()"></button></template>', () => {});
test.skip('Vue template: <template><button @click="if (q) run();\nvalue.search(/x/) !== -1"></button></template>', () => {});
test.skip('Vue template: <template><div>{{ (() => { setup()\nvalue.search(/x/) !== -1 })() }}</div></template>', () => {});
test.skip('Vue template: <template><button @click="() => value.search(/x/) !== -1"></button></template>', () => {});
test.skip('Vue template: <template><button @click="setup(); if (/x/.exec(value)) run()"></button></template>', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{value.search(/x/) !== -1}', () => {});
test.skip('Svelte template: <script>const value = "x";</script><div>{value.match(/x/) ? yes : no}</div>', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{#if value}{value.search(/x/) !== -1}{/if}', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{ value.search(/x/) !== -1}', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{/* keep */ value.search(/x/) !== -1}', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{@html value.search(/x/) !== -1}', () => {});
test.skip('Svelte template: <script>const value = "x";</script><div style:color={value.search(/x/) !== -1}></div>', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{(value.search(/x/) !== -1)}', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{value.search(/x/) === -1}', () => {});
test.skip('Svelte template: <script>const value = "x";</script><div title={value.search(/x/) !== -1}></div>', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{#if value.search(/x/) !== -1}yes{/if}', () => {});
test.skip('Svelte template: <script>const value = "x";</script><button onclick={() => value.search(/x/) !== -1}>test</button>', () => {});
test.skip('Svelte template: <script>const value = "x";</script>{value.search(/* keep */ /x/) !== -1}', () => {});
test.skip('Vue template: <template><div v-if="value.search(/x/) /* keep */ !== -1"></div></template>', () => {});
test.skip('Vue template: <template><div v-if="value.search(/x/) === /* keep */ -1"></div></template>', () => {});
test.skip('Vue template: <template><div v-if="value.match(/x/) /* keep */ .length"></div></template>', () => {});
test.skip('Vue template: <template><div v-if="value.match(/x/).length /* keep */ > 0"></div></template>', () => {});
test.skip('Vue documentation template: <template><div v-if="/unicorn/.test(string)">Vue</div></template>', () => {});
test.skip('Vue documentation template: <template><div v-if="/unicorn/.exec(string)">Vue</div></template>', () => {});
