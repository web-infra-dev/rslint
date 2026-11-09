// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-boolean-return.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-boolean-return.md
import path from 'node:path';
import { runInNewContext } from 'node:vm';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';

const ruleName = 'prefer-boolean-return';
const message = 'Return the boolean expression directly.';

function invalidCase(
  code: string,
  output: string | null,
  filename = 'src/virtual.js',
) {
  return { code, output, filename, errors: [{ messageId: ruleName, message }] };
}

const runtimeCases = [
  [
    'function unicorn(read) { if (read(/* keep */)) { return true; } return false; }',
    'function unicorn(read) { return Boolean(read(/* keep */)); }',
    Boolean,
  ],
  [
    'function unicorn(read) { if (read(/* keep */)) { return false; } return true; }',
    'function unicorn(read) { return !read(/* keep */); }',
    (value: unknown) => !value,
  ],
  [
    'function unicorn(read) { if (read(/* keep */) > 0) { return true; } else { return false; } }',
    'function unicorn(read) { return read(/* keep */) > 0; }',
    (value: unknown) => (value as number) > 0,
  ],
] as const;

new RuleTester().run(ruleName, {} as never, {
  valid: [
    // Boolean branches and guard clauses.
    {
      code: 'function unicorn() { if (test) { return true; } doSomething(); return false; }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (test) { return a; } return b; }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (test) { return true; doSomething(); } return false; }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (test) { return; } return false; }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (a ? b : c) { return true; } return false; }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (test) { return true; } else { return true; } }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (a) { return false; } if (b) { return false; } return true; }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (a) { return false; } if (b) { return false; } if (c) { return false; } return true; }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (a) { return true; } if (b) { return true; } return false; }',
      filename: 'src/virtual.js',
    },
    {
      code: 'function unicorn() { if (a) return false; if (b) return false; return true; }',
      filename: 'src/virtual.js',
    },
    // TypeScript return expressions.
    {
      code: 'function f() {\n\tif (a) {\n\t\treturn t!;\n\t}\n\n\treturn false;\n}',
      filename: 'src/virtual.ts',
    },
    // Documentation examples.
    {
      code: 'function isValid(value) {\n\treturn value > 0;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function hasItems(items) {\n\treturn Boolean(items.length);\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'function isInvalid(value) {\n\treturn !(value > 0);\n}',
      filename: 'src/virtual.js',
    },
  ],
  invalid: [
    // Comments inside the condition.
    invalidCase(
      'function unicorn() { if (value /* keep */ > 0) { return true; } else { return false; } }',
      'function unicorn() { return value /* keep */ > 0; }',
    ),
    invalidCase(
      'function unicorn() { if (value /* keep */ > 0) { return false; } return true; }',
      'function unicorn() { return !(value /* keep */ > 0); }',
    ),
    invalidCase(
      'function unicorn() { if (test(/* keep */ value)) { return true; } return false; }',
      'function unicorn() { return Boolean(test(/* keep */ value)); }',
    ),
    // Parenthesized conditions and comments that prevent edits.
    invalidCase(
      'function unicorn() { if (((/* keep */ value))) { return true; } return false; }',
      'function unicorn() { return Boolean(((/* keep */ value))); }',
    ),
    invalidCase(
      'function unicorn() { if (((value /* keep */))) { return false; } else { return true; } }',
      'function unicorn() { return !((value /* keep */)); }',
    ),
    invalidCase(
      'function unicorn() { if (value // keep\r\n > 0) { return true; } return false; }',
      'function unicorn() { return value // keep\r\n > 0; }',
    ),
    invalidCase(
      'function unicorn() { if (value as /* keep */ boolean) { return true; } return false; }',
      'function unicorn() { return Boolean(value as /* keep */ boolean); }',
      'src/virtual.ts',
    ),
    invalidCase(
      'function run() { if (value satisfies /* keep */ boolean) { return false; } return true; }',
      'function run() { return !(value satisfies /* keep */ boolean); }',
      'src/virtual.ts',
    ),
    invalidCase(
      'function unicorn() { if (value > 0 /* keep */) { return true; } return false; }',
      null,
    ),
    invalidCase(
      'function unicorn() { if (value /* copied */ > 0) { return /* keep */ true; } return false; }',
      null,
    ),
    invalidCase(
      'function unicorn() { if (value /* copied */ > 0) { return true; } /* keep */ return false; }',
      null,
    ),
    invalidCase(
      'function unicorn() { if (value /* copied */ > 0) { return true; } return false; // keep\n}',
      null,
    ),
    invalidCase(
      'function unicorn(Boolean) { if (test(/* keep */ value)) { return true; } return false; }',
      null,
    ),
    // Conditions evaluated once.
    ...runtimeCases.map(([code, output]) => invalidCase(code, output)),
    // Boolean branches and guard clauses.
    invalidCase(
      'function unicorn() { if(test){ return true; } else{ return false; } }',
      'function unicorn() { return Boolean(test); }',
    ),
    invalidCase(
      'function unicorn() { if (test) return true; else return false; }',
      'function unicorn() { return Boolean(test); }',
    ),
    invalidCase(
      'function unicorn() { if(value > 0){ return true; } else{ return false; } }',
      'function unicorn() { return value > 0; }',
    ),
    invalidCase(
      'function unicorn() { if(items.length){ return false; } else{ return true; } }',
      'function unicorn() { return !items.length; }',
    ),
    invalidCase(
      'function unicorn(Boolean) { if(test){ return true; } else{ return false; } }',
      null,
    ),
    invalidCase(
      'function unicorn() { if(test){ return true; } else{ return false; /* comment */ } }',
      null,
    ),
    invalidCase(
      'function unicorn() { if(test){ return true; /* comment */ } else{ return false; } }',
      null,
    ),
    invalidCase(
      'function unicorn() { if (test) { return true; } return false; }',
      'function unicorn() { return Boolean(test); }',
    ),
    invalidCase(
      'function unicorn() { if (test) { return false; } return true; }',
      'function unicorn() { return !test; }',
    ),
    invalidCase(
      'function unicorn() { if (value > 0) { return true; } return false; }',
      'function unicorn() { return value > 0; }',
    ),
    invalidCase(
      'function unicorn() { if (Boolean(value)) { return true; } return false; }',
      'function unicorn() { return Boolean(value); }',
    ),
    invalidCase(
      'function unicorn() { if (a === b && c === d) { return true; } return false; }',
      'function unicorn() { return a === b && c === d; }',
    ),
    invalidCase(
      'function unicorn() { if (a && b) { return true; } return false; }',
      'function unicorn() { return Boolean(a && b); }',
    ),
    invalidCase(
      'function unicorn() { if (Array.isArray(value)) { return true; } return false; }',
      'function unicorn() { return Array.isArray(value); }',
    ),
    invalidCase(
      'function unicorn(Array) { if (Array.isArray(value)) { return true; } return false; }',
      'function unicorn(Array) { return Boolean(Array.isArray(value)); }',
    ),
    invalidCase(
      'function unicorn() { if (value > 0) { return false; } return true; }',
      'function unicorn() { return !(value > 0); }',
    ),
    invalidCase(
      'function unicorn() { if (test) return true; return false; }',
      'function unicorn() { return Boolean(test); }',
    ),
    invalidCase(
      'function unicorn() { if (items.length) { return true; } return false; }',
      'function unicorn() { return Boolean(items.length); }',
    ),
    invalidCase(
      'function unicorn() { if (items.length) { return false; } return true; }',
      'function unicorn() { return !items.length; }',
    ),
    invalidCase(
      'function unicorn() { if (object?.property) { return true; } return false; }',
      'function unicorn() { return Boolean(object?.property); }',
    ),
    invalidCase(
      'function unicorn() { if (a, b) { return true; } return false; }',
      'function unicorn() { return Boolean((a, b)); }',
    ),
    invalidCase(
      'function unicorn() { if ((a, b)) { return true; } return false; }',
      'function unicorn() { return Boolean((a, b)); }',
    ),
    invalidCase(
      'function unicorn() { if ((test)) { return true; } return false; }',
      'function unicorn() { return Boolean(test); }',
    ),
    invalidCase(
      'function unicorn() { if (value as boolean) { return true; } return false; }',
      'function unicorn() { return Boolean(value as boolean); }',
      'src/virtual.ts',
    ),
    invalidCase(
      'function unicorn(value: boolean) { if (value) { return true; } return false; }',
      'function unicorn(value: boolean) { return Boolean(value); }',
      'src/virtual.ts',
    ),
    invalidCase(
      'function unicorn() { if (value satisfies boolean) { return true; } return false; }',
      'function unicorn() { return Boolean(value satisfies boolean); }',
      'src/virtual.ts',
    ),
    invalidCase(
      'function unicorn() { if (value!) { return true; } return false; }',
      'function unicorn() { return Boolean(value!); }',
      'src/virtual.ts',
    ),
    invalidCase(
      'function unicorn() { if (<boolean>value) { return true; } return false; }',
      'function unicorn() { return Boolean(<boolean>value); }',
      'src/virtual.ts',
    ),
    invalidCase(
      'function unicorn() { if (value satisfies boolean) { return false; } return true; }',
      'function unicorn() { return !(value satisfies boolean); }',
      'src/virtual.ts',
    ),
    invalidCase(
      'function unicorn() {\n\tif (test) {\n\t\treturn true;\n\t}\n\n\t// comment\n\treturn false;\n}',
      null,
    ),
    invalidCase(
      'function unicorn() { if (test) { return true; } return false; // comment\n}',
      null,
    ),
    invalidCase(
      'function unicorn() { if (/* comment */ test) { return true; } return false; }',
      null,
    ),
    invalidCase(
      'function unicorn(Boolean) { if (test) { return true; } return false; }',
      null,
    ),
    invalidCase(
      'function unicorn(Boolean) { if (test) { return false; } return true; }',
      'function unicorn(Boolean) { return !test; }',
    ),
    invalidCase(
      'function unicorn(Boolean) { if (value > 0) { return true; } return false; }',
      'function unicorn(Boolean) { return value > 0; }',
    ),
    invalidCase(
      'function unicorn() { if (a) { return false; } if (b) { return true; } return false; }',
      'function unicorn() { if (a) { return false; } return Boolean(b); }',
    ),
    invalidCase(
      'function unicorn() {\n\tif (test({\n\t\tmultiline: true,\n\t})) {\n\t\treturn true;\n\t}\n\n\treturn false;\n}',
      'function unicorn() {\n\treturn Boolean(test({\n\t\tmultiline: true,\n\t}));\n}',
    ),
    invalidCase(
      'function unicorn() { if (!test) { return true; } return false; }',
      'function unicorn() { return !test; }',
    ),
    invalidCase(
      'function unicorn() { if (delete object.key) { return true; } return false; }',
      'function unicorn() { return delete object.key; }',
    ),
    invalidCase(
      'function unicorn() { if (typeof value) { return true; } return false; }',
      'function unicorn() { return Boolean(typeof value); }',
    ),
    invalidCase(
      'function unicorn() { if (a === b || false) { return true; } return false; }',
      'function unicorn() { return a === b || false; }',
    ),
    invalidCase(
      'function unicorn() { if ((value > 0)) { return false; } return true; }',
      'function unicorn() { return !(value > 0); }',
    ),
    invalidCase(
      'type Boolean = string; function unicorn() { if (test) { return true; } return false; }',
      'type Boolean = string; function unicorn() { return Boolean(test); }',
      'src/virtual.ts',
    ),
    // TypeScript return expressions.
    invalidCase(
      'function f() {\n\tif (a) {\n\t\treturn true as const;\n\t}\n\n\treturn false;\n}',
      'function f() {\n\treturn Boolean(a);\n}',
      'src/virtual.ts',
    ),
    invalidCase(
      'function f() {\n\tif (a) {\n\t\treturn true as boolean;\n\t}\n\n\treturn false;\n}',
      'function f() {\n\treturn Boolean(a);\n}',
      'src/virtual.ts',
    ),
    invalidCase(
      'function f() {\n\tif (a) {\n\t\treturn true satisfies boolean;\n\t}\n\n\treturn false;\n}',
      'function f() {\n\treturn Boolean(a);\n}',
      'src/virtual.ts',
    ),
    invalidCase(
      'function f() {\n\tif (a) {\n\t\treturn false as const;\n\t}\n\n\treturn true;\n}',
      'function f() {\n\treturn !a;\n}',
      'src/virtual.ts',
    ),
    invalidCase(
      'function f() {\n\tif (a) {\n\t\treturn false!;\n\t}\n\n\treturn true;\n}',
      'function f() {\n\treturn !a;\n}',
      'src/virtual.ts',
    ),
    // Documentation examples.
    invalidCase(
      'function isValid(value) {\n\tif (value > 0) {\n\t\treturn true;\n\t}\n\n\treturn false;\n}',
      'function isValid(value) {\n\treturn value > 0;\n}',
    ),
    invalidCase(
      'function hasItems(items) {\n\tif (items.length) {\n\t\treturn true;\n\t}\n\n\treturn false;\n}',
      'function hasItems(items) {\n\treturn Boolean(items.length);\n}',
    ),
    invalidCase(
      'function isInvalid(value) {\n\tif (value > 0) {\n\t\treturn false;\n\t}\n\n\treturn true;\n}',
      'function isInvalid(value) {\n\treturn !(value > 0);\n}',
    ),
  ],
});

test('comment-preserving boolean returns evaluate their condition once', async () => {
  const filename = path.resolve(import.meta.dirname, '../src/virtual.js');
  for (const [code, output, expected] of runtimeCases) {
    const fixed = await lint({
      config: [
        { plugins: ['unicorn'], rules: { [`unicorn/${ruleName}`]: 'error' } },
      ],
      fileContents: { [filename]: code },
      fix: true,
    });
    expect(fixed.fileCount).toBe(1);
    expect(fixed.ruleCount).toBe(1);
    expect(fixed.diagnostics).toEqual([]);
    const [fixedCode] = Object.values(fixed.output ?? {});
    assert.strictEqual(fixedCode, output);
    expect(fixedCode.split('/* keep */')).toHaveLength(2);
    for (const value of [0, 1, '', 'value', undefined, false, []]) {
      for (const source of [code, fixedCode]) {
        let calls = 0;
        const result = runInNewContext(`${source}\nunicorn(read);`, {
          read() {
            calls++;
            return value;
          },
        });
        expect(result).toBe(expected(value));
        expect(calls).toBe(1);
      }
    }
  }
});
