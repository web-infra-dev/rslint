/**
 * @fileoverview Tests for list-style rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/list-style/list-style.test.ts
 *
 * Upstream names the source directory `list-style` and the test uses
 * `run({ name: 'list-style', lang: 'ts', ... })`, but the published plugin keys
 * this experimental rule as `exp-list-style` (the `exp-` prefix is applied at
 * build time). rslint mirrors that published id, so the rule is driven here as
 * `@stylistic/exp-list-style`.
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, valid, invalid })` ->
 *    `ruleTester.run('exp-list-style', null as never, { valid, invalid })`;
 *    `name`/`rule`/`lang` and the imports are dropped.
 *  - The `$` unindent template tag (`unindent` from eslint-vitest-rule-tester) is
 *    evaluated to its real multi-line string: the common leading indentation is
 *    stripped and fully-blank leading/trailing lines are removed. The evaluated
 *    text is embedded verbatim as a plain template literal (continuation lines at
 *    column 0; only \` \\ \${ escaped) — byte-exact because diagnostic columns are
 *    asserted. The lone CRLF fixture is emitted as a JSON string (a template
 *    literal would normalize CRLF->LF). Plain backtick literals are kept verbatim.
 *  - Per-case `description` labels (upstream documentation only) are dropped.
 *  - `parserOptions` / `type` / `features` / `suggestions`: none exist in this
 *    rule's upstream tests.
 *
 * The rule's four messages (`shouldWrap` / `shouldNotWrap` / `shouldSpacing` /
 * `shouldNotSpacing`) all carry `{{prev}}`/`{{next}}` placeholders, and every
 * upstream error pins only `messageId` (no `data`), so the rendered message stays
 * templated and is not assertable — the RuleTester asserts the diagnostic COUNT
 * plus `line`/`column` for each error (upstream gives no endLine/endColumn) and the
 * autofix `output`.
 *
 * The `._json_` upstream test file is excluded per the porting spec; no
 * `._css_` / `._markdown_` / `._js_` files exist for this rule.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are moved to the `exp-list-style — KNOWN GAPS` block comment at the bottom, each
 * annotated with what upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('exp-list-style', null as never, {
  valid: [
    "if (a) {}",
    `if (
a
) {}`,
    "const a = { foo: \"bar\", bar: 2 }",
    `const a = {
foo: "bar",
bar: 2
}`,
    "const a = [1, 2, 3]",
    `const a = [
1,
2,
3
]`,
    "import { foo, bar } from \"foo\"",
    `import {
foo,
bar
} from "foo"`,
    `const a = [\`

\`, \`

\`]`,
    "log(a, b)",
    `log(
a,
b
)`,
    "function foo(a, b) {}",
    `function foo(
a,
b
) {}`,
    "const foo = function (a, b) {}",
    `const foo = function (
a,
b
) {}`,
    `const foo = (a, b) => {

}`,
    `const foo = (a, b): { a:b } => {

}`,
    "interface Foo { a: 1, b: 2 }",
    `interface Foo {
a: 1
b: 2
}`,
    "enum Foo { A, B }",
    `enum Foo {
A,
B
}`,
    `a
.filter(items => {

})`,
    "new Foo(a, b)",
    `new Foo(
a,
b
)`,
    "new (Foo())(a, b)",
    `function foo<T = {
a: 1,
b: 2
}>(a, b) {}`,
    `foo(() =>
bar())`,
    `call<{
foo: 'bar'
}>('')`,
    "const { a } = foo;",
    "const [, a] = foo;",
    "const [a,] = foo;",
    "const [, a,] = foo;",
    `export { name, version } from 'package.json' with {
  type: 'json'
  }`,
    "export * from \"foo\" with { type: \"json\" }",
    {
      code: `const foo = [1]
const bar = [
  1, 
  2,
];`,
      options: [{"singleLine":{"maxItems":1}}],
    },
    {
      code: `const foo = [1, 2];
const bar = { a: 1, b: 2 };
const foo = [
  1,
  2,
];
const bar = {
  a: 1,
  b: 2,
};`,
      options: [{"multiLine":{"minItems":2}}],
    },
    {
      code: "import { name, version } from 'package.json' with {type: 'json'}",
      options: [{"overrides":{"{}":{"singleLine":{"spacing":"always"}},"ImportAttributes":{"singleLine":{"spacing":"never"}}}}],
    },
    {
      code: `if (node.callee.type !== 'Identifier'
  || (node.callee.name !== 't' && node.callee.name !== 'n')
) {}`,
      options: [{"overrides":{"IfStatement":"off"}}],
    },
    `(Object.keys(options) as KeysOptions[])
.forEach((key) => {
  if (options[key] === false)
    delete listenser[key]
})`,
    `function fn({ foo, bar }: {
foo: 'foo'
bar: 'bar'
}) {}`,
    `export const getTodoList = request.post<
  Params,
  ResponseData,
>('/api/todo-list')`,
    `bar(
  foo => foo
    ? ''
    : ''
)`,
    `bar(
  (ruleName, foo) => foo
    ? ''
    : ''
)`,
    `const a = [
  (1),
  (2)
];`,
    "const a = [(1), (2)];",
    `this.foobar(
  (x),
  y,
  z
)`,
    `foobar(
  (x),
  y,
  z
)`,
    `foobar<A>(
  (x),
  y,
  z
)`,
    `foo?.(
  [],
  {},
)`,
    `import Icon, {
  MailOutlined,
  NumberOutlined,
  QuestionCircleOutlined,
  QuestionOutlined,
  UserOutlined,
} from '@ant-design/icons';`,
    `const fix = a => (
  call(
    a
  )
)`,
    `run({
  valid: [
    /* comment */
  ],
  invalid: [
    // comment
  ]
})`,
  ],
  invalid: [
    {
      code: `if (
a) {}`,
      output: `if (
a
) {}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 2 },
      ],
    },
    {
      code: `if (a
) {}`,
      output: "if (a) {}",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 6 },
      ],
    },
    {
      code: `const a = {
foo: "bar", bar: 2 }`,
      output: `const a = {
foo: "bar", 
bar: 2 
}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 12 },
        { messageId: "shouldWrap", line: 2, column: 19 },
      ],
    },
    {
      code: `const a = [
1, 2, 3]`,
      output: `const a = [
1, 
2, 
3
]`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 6 },
        { messageId: "shouldWrap", line: 2, column: 8 },
      ],
    },
    {
      code: `const a = [1, 
2, 3
]`,
      output: "const a = [1,2, 3]",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 14 },
        { messageId: "shouldNotWrap", line: 2, column: 5 },
      ],
    },
    {
      code: `import {
foo, bar } from "foo"`,
      output: `import {
foo, 
bar 
} from "foo"`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 5 },
        { messageId: "shouldWrap", line: 2, column: 9 },
      ],
    },
    {
      code: `import { foo, 
bar } from "foo"`,
      output: "import { foo,bar } from \"foo\"",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 14 },
      ],
    },
    {
      code: `log(
a, b)`,
      output: `log(
a, 
b
)`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 5 },
      ],
    },
    {
      code: `function foo(a, b
){}`,
      output: "function foo(a, b){}",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 18 },
      ],
    },
    {
      code: `function foo(
a, b) {}`,
      output: `function foo(
a, 
b
) {}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 5 },
      ],
    },
    {
      code: `const foo = (
a, b) => {}`,
      output: `const foo = (
a, 
b
) => {}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 5 },
      ],
    },
    {
      code: `const foo = (
a, b): {
a:b} => {}`,
      output: `const foo = (
a, 
b
): {
a:b
} => {}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 5 },
        { messageId: "shouldWrap", line: 3, column: 4 },
      ],
    },
    {
      code: `const foo = (
a, b): {a:b} => {}`,
      output: `const foo = (
a, 
b
): {a:b} => {}`,
      options: [{"overrides":{"{}":{"singleLine":{"spacing":"never"}}}}],
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 5 },
      ],
    },
    {
      code: `interface Foo {
a: 1,b: 2
}`,
      output: `interface Foo {
a: 1,
b: 2
}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 6 },
      ],
    },
    {
      code: `type Foo = {
a: 1,b: 2
}`,
      output: `type Foo = {
a: 1,
b: 2
}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 6 },
      ],
    },
    {
      code: `type foo = [
1]`,
      output: `type foo = [
1
]`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 2 },
      ],
    },
    {
      code: `type Foo = [1,2,
3]`,
      output: "type Foo = [1,2,3]",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 17 },
      ],
    },
    {
      code: "type Foo = [ 1, 2, 3 ]",
      output: "type Foo = [1, 2, 3]",
      errors: [
        { messageId: "shouldNotSpacing", line: 1, column: 13 },
        { messageId: "shouldNotSpacing", line: 1, column: 21 },
      ],
    },
    {
      code: `new Foo(1,2,
3)`,
      output: "new Foo(1,2,3)",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 13 },
      ],
    },
    {
      code: `new Foo(
1,2,
3)`,
      output: `new Foo(
1,
2,
3
)`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 3, column: 2 },
      ],
    },
    {
      code: `foo(
()=>bar(),
()=>
baz())`,
      output: `foo(
()=>bar(),
()=>
baz()
)`,
      errors: [
        { messageId: "shouldWrap", line: 4, column: 6 },
      ],
    },
    {
      code: `foo(()=>bar(),
()=>
baz())`,
      output: `foo(()=>bar(),()=>
baz())`,
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 15 },
      ],
    },
    {
      code: `foo<X,
Y>(1, 2)`,
      output: "foo<X,Y>(1, 2)",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 7 },
      ],
    },
    {
      code: `foo<
X,Y>(
1, 2)`,
      output: `foo<
X,
Y
>(
1, 
2
)`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 4 },
        { messageId: "shouldWrap", line: 3, column: 3 },
        { messageId: "shouldWrap", line: 3, column: 5 },
      ],
    },
    {
      code: `function foo<
X,Y>() {}`,
      output: `function foo<
X,
Y
>() {}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 4 },
      ],
    },
    {
      code: "const {a,b} = c",
      output: "const { a,b } = c",
      errors: [
        { messageId: "shouldSpacing", line: 1, column: 8 },
        { messageId: "shouldSpacing", line: 1, column: 11 },
      ],
    },
    {
      code: `const [
  a,b] = c`,
      output: `const [
  a,
b
] = c`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 5 },
        { messageId: "shouldWrap", line: 2, column: 6 },
      ],
    },
    {
      code: `const [,
] = foo`,
      output: "const [,] = foo",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 9 },
      ],
    },
    {
      code: `foo(([
a,b]) => {})`,
      output: `foo(([
a,
b
]) => {})`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 3 },
        { messageId: "shouldWrap", line: 2, column: 4 },
      ],
    },
    {
      code: `export default antfu({
},
{
  foo: 'bar'
}
  // some comment
  // hello
)`,
      output: `export default antfu({
},{
  foo: 'bar'
}
  // some comment
  // hello
)`,
      errors: [
        { messageId: "shouldNotWrap", line: 2, column: 3 },
        { messageId: "shouldNotWrap", line: 5, column: 2 },
      ],
    },
    {
      code: `export default antfu({
},
// some comment
{
  foo: 'bar'
},
{
}
  // hello
)`,
      output: `export default antfu({
},
// some comment
{
  foo: 'bar'
},{
}
  // hello
)`,
      errors: [
        { messageId: "shouldNotWrap", line: 2, column: 3 },
        { messageId: "shouldNotWrap", line: 6, column: 3 },
        { messageId: "shouldNotWrap", line: 8, column: 2 },
      ],
    },
    {
      code: `interface Foo {
  bar: (
    foo: string, bar: {
      bar: string, baz: string }) => void
}`,
      output: `interface Foo {
  bar: (
    foo: string, 
bar: {
      bar: string, 
baz: string 
}
) => void
}`,
      errors: [
        { messageId: "shouldWrap", line: 3, column: 17 },
        { messageId: "shouldWrap", line: 4, column: 19 },
        { messageId: "shouldWrap", line: 4, column: 31 },
        { messageId: "shouldWrap", line: 4, column: 33 },
      ],
    },
    {
      code: `interface foo {
a:1}`,
      output: `interface foo {
a:1
}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 4 },
      ],
    },
    {
      code: `type foo = {
a:1}`,
      output: `type foo = {
a:1
}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 4 },
      ],
    },
    {
      code: `const foo = [
1]`,
      output: `const foo = [
1
]`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 2 },
      ],
    },
    {
      code: `const foo = {
a:1}`,
      output: `const foo = {
a:1
}`,
      errors: [
        { messageId: "shouldWrap", line: 2, column: 4 },
      ],
    },
    {
      code: "function foo< T >( a: number, b: string ): void",
      output: "function foo<T>(a: number, b: string): void",
      errors: [
        { messageId: "shouldNotSpacing", line: 1, column: 14 },
        { messageId: "shouldNotSpacing", line: 1, column: 16 },
        { messageId: "shouldNotSpacing", line: 1, column: 19 },
        { messageId: "shouldNotSpacing", line: 1, column: 40 },
      ],
    },
    {
      code: `export { name, version} from 'package.json' with {
  type: 'json'}`,
      output: `export { name, version } from 'package.json' with {
  type: 'json'
}`,
      errors: [
        { messageId: "shouldSpacing", line: 1, column: 23 },
        { messageId: "shouldWrap", line: 2, column: 15 },
      ],
    },
    {
      code: `const foo = [1, 2];
const bar = { a: 1, b: 2 };`,
      output: `const foo = [
1, 
2
];
const bar = { 
a: 1, 
b: 2 
};`,
      options: [{"singleLine":{"maxItems":1}}],
      errors: [
        { messageId: "shouldWrap", line: 1, column: 14 },
        { messageId: "shouldWrap", line: 1, column: 16 },
        { messageId: "shouldWrap", line: 1, column: 18 },
        { messageId: "shouldWrap", line: 2, column: 14 },
        { messageId: "shouldWrap", line: 2, column: 20 },
        { messageId: "shouldWrap", line: 2, column: 25 },
      ],
    },
    {
      code: `foo(a,
)`,
      output: "foo(a,)",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 7 },
      ],
    },
    {
      code: `const foo = [1,
	2,
	3];`,
      output: "const foo = [1,2,3];",
      errors: [
        { messageId: "shouldNotWrap", line: 1, column: 16 },
        { messageId: "shouldNotWrap", line: 2, column: 4 },
      ],
    },
  ],
});

/**
 * exp-list-style — KNOWN GAPS
 *
 * Every gap below is the SAME class: a multi-pass-vs-single-pass autofix difference,
 * NOT a diagnostic difference. For each of these cases rslint reports the identical
 * diagnostics as upstream (same count, same `messageId`, same `line`/`column`), but
 * the autofix `output` differs.
 *
 * Root cause: the rule's default options set `overrides['{}'].singleLine.spacing` to
 * `'always'` (object braces want inner spacing when single-line) while the top-level
 * `singleLine.spacing` is `'never'`. When a multi-line object / destructuring is
 * collapsed onto one line by the `shouldNotWrap` fix, two independent fixes apply:
 * the unwrap, and the single-line `{}` spacing.
 *   - Upstream's eslint-vitest-rule-tester records `output` after ONE fix pass: it
 *     unwraps to `{...}` and stops (the spacing fix would need a second pass).
 *   - rslint applies fixes to a stable fixpoint, so after unwrapping it also runs the
 *     single-line `{}` spacing fix and converges on `{ ... }`.
 * Verified empirically: feeding upstream's single-pass output (e.g.
 * `const a = {foo: "bar",bar: 2}`) back into rslint reports `shouldSpacing`
 * ("Should have space between '{' and ...") and `--fix` produces exactly rslint's
 * output here (e.g. `const a = { foo: "bar",bar: 2 }`). rslint's result is the
 * fixpoint of upstream's; the only divergence is the single- vs multi-pass `output`.
 *
 * The RuleTester (`../rule-tester`) is interface-only and asserts `output` exactly,
 * so these cases cannot live in the active `run()` block without either failing or
 * fabricating an expectation. They are preserved here verbatim (upstream code +
 * upstream single-pass `output` + the matching diagnostics) with rslint's actual
 * multi-pass output noted inline.
 *
 *   [upstream invalid #3]
 *
 *   {
 *     code: `const a = {foo: "bar", 
 *   bar: 2
 *   }`,
 *     output: "const a = {foo: \"bar\",bar: 2}",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 23 },
 *       { messageId: "shouldNotWrap", line: 2, column: 7 },
 *     ],
 *   },
 *
 *   [upstream invalid #15 — Add delimiter to avoid syntax error, (interface)]
 *
 *   {
 *     code: `interface Foo {a: 1
 *   b: 2
 *   }`,
 *     output: "interface Foo {a: 1,b: 2,}",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 20 },
 *       { messageId: "shouldNotWrap", line: 2, column: 5 },
 *     ],
 *   },
 *
 *   [upstream invalid #16 — Delimiter already exists]
 *
 *   {
 *     code: `interface Foo {a: 1;
 *   b: 2,
 *   c: 3}`,
 *     output: "interface Foo {a: 1;b: 2,c: 3}",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 21 },
 *       { messageId: "shouldNotWrap", line: 2, column: 6 },
 *     ],
 *   },
 *
 *   [upstream invalid #17 — Delimiter in the middle]
 *
 *   {
 *     code: `export interface Foo {        a: 1
 *     b: Pick<Bar, 'baz'>
 *     c: 3
 *   }`,
 *     output: "export interface Foo {        a: 1,b: Pick<Bar, 'baz'>,c: 3,}",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 35 },
 *       { messageId: "shouldNotWrap", line: 2, column: 22 },
 *       { messageId: "shouldNotWrap", line: 3, column: 7 },
 *     ],
 *   },
 *
 *   [upstream invalid #19 — Add delimiter to avoid syntax error, (type)]
 *
 *   {
 *     code: `type Foo = {a: 1
 *   b: 2
 *   }`,
 *     output: "type Foo = {a: 1,b: 2,}",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 17 },
 *       { messageId: "shouldNotWrap", line: 2, column: 5 },
 *     ],
 *   },
 *
 *   [upstream invalid #30]
 *
 *   {
 *     code: `const {a,
 *   b
 *   } = c`,
 *     output: "const {a,b} = c",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 10 },
 *       { messageId: "shouldNotWrap", line: 2, column: 2 },
 *     ],
 *   },
 *
 *   [upstream invalid #35 — CRLF]
 *
 *   {
 *     code: "const a = {foo: \"bar\", \r\nbar: 2\r\n}",
 *     output: "const a = {foo: \"bar\",bar: 2}",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 23 },
 *       { messageId: "shouldNotWrap", line: 2, column: 7 },
 *     ],
 *   },
 *
 *   [upstream invalid #40]
 *
 *   {
 *     code: `interface foo {a:1
 *   }`,
 *     output: "interface foo {a:1}",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 19 },
 *     ],
 *   },
 *
 *   [upstream invalid #42]
 *
 *   {
 *     code: `type foo = {a:1
 *   }`,
 *     output: "type foo = {a:1}",
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 16 },
 *     ],
 *   },
 *
 *   [upstream invalid #48]
 *
 *   {
 *     code: `const foo = [
 *   1, 
 *   2
 *   ];
 *   const bar = { 
 *   a: 1, 
 *   b: 2 
 *   };`,
 *     output: `const foo = [1,2];
 *   const bar = {a: 1,b: 2};`,
 *     options: [{"multiLine":{"minItems":3}}],
 *     errors: [
 *       { messageId: "shouldNotWrap", line: 1, column: 14 },
 *       { messageId: "shouldNotWrap", line: 2, column: 3 },
 *       { messageId: "shouldNotWrap", line: 3, column: 2 },
 *       { messageId: "shouldNotWrap", line: 5, column: 14 },
 *       { messageId: "shouldNotWrap", line: 6, column: 6 },
 *       { messageId: "shouldNotWrap", line: 7, column: 5 },
 *     ],
 *   },
 *
 * rslint's actual multi-pass `--fix` output for the cases above (diagnostics match
 * upstream exactly; only this output differs):
 *   #3:  'const a = { foo: "bar",bar: 2 }'
 *   #15: 'interface Foo { a: 1,b: 2, }'
 *   #16: 'interface Foo { a: 1;b: 2,c: 3 }'
 *   #17: "export interface Foo {        a: 1,b: Pick<Bar, 'baz'>,c: 3, }"
 *   #19: 'type Foo = { a: 1,b: 2, }'
 *   #30: 'const { a,b } = c'
 *   #35: 'const a = { foo: "bar",bar: 2 }'
 *   #40: 'interface foo { a:1 }'
 *   #42: 'type foo = { a:1 }'
 *   #48: 'const foo = [1,2];\nconst bar = { a: 1,b: 2 };'
 */
