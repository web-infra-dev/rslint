import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const error = (name: string) => ({
  messageId: 'error',
  message: `\`${name}\` should be a \`Set\`, and use \`${name}.has()\` to check existence or non-existence.`,
});

const js = (code: string) => ({ code, filename: 'file.js' });

ruleTester.run('prefer-set-has', null as never, {
  valid: [
    js(
      'const foo = new Set([1, 2, 3]);\nfunction unicorn() {\n\treturn foo.has(1);\n}',
    ),
    // Only called once
    js('const foo = [1, 2, 3];\nconst isExists = foo.includes(1);'),
    js('const foo = [1, 2, 3];\n(() => {})(foo.includes(1));'),
    // Not `VariableDeclarator`
    js('foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1);\n}'),
    js('const exists = foo.includes(1);'),
    js('const exists = [1, 2, 3].includes(1);'),
    // Didn't call `includes()`
    js('const foo = [1, 2, 3];'),
    // Not `foo.includes()`
    js(
      'const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes;\n}',
    ),
    js(
      'const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn bar.includes(foo);\n}',
    ),
    js(
      'const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.indexOf(1) !== -1;\n}',
    ),
    // Unsupported extra references
    js(
      'const foo = [1, 2, 3];\nfunction unicorn() {\n\tfoo.includes(1);\n\tfoo.length = 1;\n}',
    ),
    // Duplicates / NaN / -0
    js(
      'const foo = [1, 1, 2];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    js(
      'const foo = [0, -0];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    // Unknown uniqueness
    js(
      'const value = 1;\nconst foo = [value, value];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    // Unsupported initializer shapes
    js(
      'const foo = [1, , 2];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    js(
      'const foo = [...bar];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    // forEach with non-arrow / multi-param callback
    js(
      'const foo = [1, 2, 3];\nfoo.forEach((element, index) => {\n\tconsole.log(element, index);\n});\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    js(
      'const foo = [1, 2, 3];\nfoo.forEach(callback);\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    // `.length` writes
    js(
      'const foo = [1, 2, 3];\ndelete foo.length;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    js(
      'const foo = [1, 2, 3];\nfoo.length++;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    ),
    // Declared more than once
    js(
      'var foo = [1, 2, 3];\nvar foo = [4, 5, 6];\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
    ),
    js('const foo = bar;\nfunction unicorn() {\n\treturn foo.includes(1);\n}'),
    // Extra / spread arguments
    js(
      'const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1, 1);\n}',
    ),
    js(
      'const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(...[1]);\n}',
    ),
    // Optional
    js(
      'const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo?.includes(1);\n}',
    ),
    js(
      'const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes?.(1);\n}',
    ),
    // Different scope
    js(
      'function unicorn() {\n\tconst foo = [1, 2, 3];\n}\nfunction unicorn2() {\n\treturn foo.includes(1);\n}',
    ),
    // `export`
    js(
      'export const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
    ),
    // `Array.from` / `Array.of` — not Array
    js(
      'const foo = NotArray.of(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
    ),
    js(
      "const foo = Array['of'](1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}",
    ),
    // Not listed method
    js(
      'const foo = bar.notListed();\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
    ),
    // string slice/concat false positive
    js(
      "const text = 'abc'.slice();\ntext.includes('ab') || text.includes('bc');",
    ),
    // minimumItems below threshold
    {
      ...js(
        'const foo = [1, 2, 3, 4];\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      options: [{ minimumItems: 5 }],
    },
    {
      ...js(
        'const foo = Array(4);\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      options: [{ minimumItems: 5 }],
    },
    // TS `.length` write targets
    {
      code: 'const foo = [1, 2, 3];\n(foo.length as number) = 1;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
    },
    // A `for` initializer runs once, so a single `includes` in the loop body
    // is still a single call — the walk stops at the `ForStatement` root.
    js('for (const foo = [1, 2]; ;) {\n\tfoo.includes(1);\n}'),
    // Folded elements that collide are duplicates, so the array is not
    // known-unique and the extra `forEach` reference blocks the report.
    js(
      "const foo = ['ab', 'a' + 'b'];\nfoo.forEach(x => log(x));\nfunction f() { return foo.includes('ab'); }",
    ),
  ],
  invalid: [
    {
      ...js(
        'const foo = [1, 2, 3];\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    // Called multiple times
    {
      ...js(
        'const foo = [1, 2, 3];\nconst isExists = foo.includes(1);\nconst isExists2 = foo.includes(2);',
      ),
      errors: [error('foo')],
    },
    // Loops
    {
      ...js(
        'const foo = [1, 2, 3];\nfor (const a of b) {\n\tfoo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = [1, 2, 3];\nfor (let i = 0; i < n; i++) {\n\tfoo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js('const foo = [1, 2, 3];\nwhile (a)  {\n\tfoo.includes(1);\n}'),
      errors: [error('foo')],
    },
    {
      ...js('const foo = [1, 2, 3];\ndo {\n\tfoo.includes(1);\n} while (a)'),
      errors: [error('foo')],
    },
    // Function boundaries
    {
      ...js(
        'const foo = [1, 2, 3];\nfunction * unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js('const foo = [1, 2, 3];\nconst unicorn = () => foo.includes(1);'),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = [1, 2, 3];\nclass A {\n\tb() {\n\t\treturn foo.includes(1);\n\t}\n}',
      ),
      errors: [error('foo')],
    },
    // Multiple references
    {
      ...js(
        'const foo = [1, 2, 3];\nfunction unicorn() {\n\tconst exists = foo.includes(1);\n\tfunction isExists(find) {\n\t\treturn foo.includes(find);\n\t}\n}',
      ),
      errors: [error('foo')],
    },
    // Extra references with a known-unique literal
    {
      ...js(
        'const foo = [1, 2, 3];\nfor (const element of foo) {\n\tconsole.log(element);\n}\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = [1, 2, 3];\nconst length = foo.length;\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = [1, 2, 3];\nfoo.forEach(element => {\n\tconsole.log(element);\n});\n\nfunction unicorn(value) {\n\treturn foo.includes(value);\n}',
      ),
      errors: [error('foo')],
    },
    // Array constructors / statics
    {
      ...js(
        'const foo = Array(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = new Array(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = Array.from({length: 1}, (_, index) => index);\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = Array.of(1, 2);\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    // Array-returning methods
    {
      ...js(
        'const foo = bar.map();\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = [1, 2, 3].slice();\nfoo.includes(1) || foo.includes(2);',
      ),
      errors: [error('foo')],
    },
    // minimumItems at/above threshold
    {
      ...js(
        'const foo = [1, 2, 3, 4, 5];\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      options: [{ minimumItems: 5 }],
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = Array(5);\nfunction unicorn() {\n\treturn foo.includes(1);\n}',
      ),
      options: [{ minimumItems: 5 }],
      errors: [error('foo')],
    },
    // TypeScript type-annotation autofix / suggestion
    {
      code: "const a: string[] = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
      errors: [error('a')],
    },
    {
      code: "const a: [string, string] = ['foo', 'bar']\n\nfor (let i = 0; i < 3; i++) {\n\tif (a.includes(someString)) {\n\t\tconsole.log(123)\n\t}\n}",
      errors: [error('a')],
    },
    // `for` initializer declaration — upstream's structural gate is the
    // declaration node, which also sits at `ForStatement.init`.
    {
      ...js(
        'for (const foo = [1, 2]; ;) {\n\tfoo.includes(1);\n\tfoo.includes(2);\n}',
      ),
      errors: [error('foo')],
    },
    // `Array.from({length})` shorthand property drives the size check.
    {
      ...js(
        'const length = 3;\nconst foo = Array.from({length});\nfoo.includes(1);\nfoo.includes(2);',
      ),
      options: [{ minimumItems: 2 }],
      errors: [error('foo')],
    },
    // A comment inside `Array<…>` stays inside the type-argument span, so the
    // annotation is still rewritable and this autofixes rather than degrading
    // to a suggestion.
    {
      code: "const foo: Array</* c */ string> = ['a', 'b'];\nfoo.includes('a');\nfoo.includes('b');",
      errors: [error('foo')],
    },
    // Statically folded elements: string concatenation, template literals and
    // well-known numeric globals are all comparable, so the extra `forEach`
    // reference does not block the report.
    {
      ...js(
        "const foo = ['a' + 'b', 'c'];\nfoo.forEach(x => log(x));\nfunction f() { return foo.includes('c'); }",
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        "const foo = [`a${1}`, 'b'];\nfoo.forEach(x => log(x));\nfunction f() { return foo.includes('b'); }",
      ),
      errors: [error('foo')],
    },
    {
      ...js(
        'const foo = [Number.NaN, 1];\nfoo.forEach(x => log(x));\nfunction f() { return foo.includes(1); }',
      ),
      errors: [error('foo')],
    },
  ],
});
