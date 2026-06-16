/**
 * @fileoverview Tests for indent-binary-ops rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/indent-binary-ops/indent-binary-ops.test.ts
 *
 * Transformations applied per the porting spec:
 *   - `run<RuleOptions, MessageIds>({ name, rule, recursive, ... })` ->
 *     `ruleTester.run('indent-binary-ops', null as never, { valid, invalid })`.
 *   - The `$` (unindent) template tag is evaluated to its real string; each
 *     multi-line fixture is emitted as a `\n`-escaped literal so the exact
 *     (indentation-sensitive) bytes are preserved and cannot be reflowed.
 *   - `name` / `rule` / `recursive` / the `#test` + rule imports dropped.
 *
 * Every upstream invalid case pins ONLY `code` + `output` (no `errors`): the
 * upstream eslint-vitest-rule-tester verifies the autofix, not a diagnostic set.
 * They are ported `errors`-less — the RuleTester asserts the `--fix` output and a
 * sanity check that the case genuinely reports (>=1 diagnostic); positions are
 * never invented. There are no Babel/Flow, suggestion, or external-fixture cases,
 * and no `._css_` / `._json_` / `._markdown_` files exist for this rule.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('indent-binary-ops', null as never, {
  valid: [
    "type a = {\n  [K in keyof T]: T[K] extends Date\n    ? Date | string\n    : T[K] extends (Date | null)\n      ? Date | string | null\n      : T[K];\n}",
    "type Foo =\n  | A\n  | B",
    "if (\n  this.level >= this.max ||\n  this.level <= this.min\n) {\n  this.overflow = true;\n}",
    "const woof = computed(() => keys.value.filter(\n  ({ type }) => type === 'bark' ||\n    type === 'pooque' ||\n    type === 'srenque'\n  ));",
    "if (a\n  && b\n  && c\n  && (d\n    || e\n    || f\n  )\n) {\n  foo()\n}",
    "type Foo = Pick<Bar,\n  Baz\n  | Qux,\n>;",
    "type Foo = [Bar,\n  Baz\n  | Qux,\n];",
    "type Foo = { x: Foo,\n  y: Baz\n    | Quz\n};",
    "type Foo = Pick<Bar\n  | Baz,\n  Baz\n  | Qux,\n>;",
    "type Foo = [Bar\n  | Baz,\n  Baz\n  | Qux,\n];",
    "type Foo = { x: Foo\n  | Baz,\n  y: Baz\n    | Quz\n};",
    "const a = 1\n  + 2\n  + 3;",
    "a = 1\n  + 2\n  + 3;",
    "const a = 1 +\n  2 +\n  3;",
    "a = 1 +\n  2 +\n  3;",
    "this.a = this.b\n  || c\n  || d;",
    "{ aaaaa &&\n  bbbbb &&\n  ccccc }",
    "{\n  aaaaa &&\n  bbbbb &&\n  ccccc\n}",
    "if (condition1 &&\n  condition2 &&\n  condition3\n) {\n  a &&\n  b() &&\n  c()\n}",
  ],
  invalid: [
    {
      code: "if (\n  a && (\n    a.b ||\n      a.c\n  ) &&\n    a.d\n) {}",
      output: "if (\n  a && (\n    a.b ||\n    a.c\n  ) &&\n  a.d\n) {}",
    },
    {
      code: "const a =\n  x +\n    y * z",
      output: "const a =\n  x +\n  y * z",
    },
    {
      code: "if (\n  aaaaaa >\nbbbbb\n) {}",
      output: "if (\n  aaaaaa >\n  bbbbb\n) {}",
    },
    {
      code: "function foo() {\n  if (a\n  || b\n      || c || d\n        || (d && b)\n  ) {\n    foo()\n  }\n}",
      output: "function foo() {\n  if (a\n    || b\n    || c || d\n    || (d && b)\n  ) {\n    foo()\n  }\n}",
    },
    {
      code: "type Foo = A | B\n  | C | D\n    | E",
      output: "type Foo = A | B\n  | C | D\n  | E",
    },
    {
      code: "type Foo =\n| A | C\n  | B",
      output: "type Foo =\n  | A | C\n  | B",
    },
    {
      code: "type Foo =\n| A | C\n  | B",
      output: "type Foo =\n  | A | C\n  | B",
    },
    {
      code: "type T =\n& A\n  & (B\n  | A\n  | D)",
      output: "type T =\n  & A\n  & (B\n    | A\n    | D)",
    },
    {
      code: "type T =\na\n| b\n  | c",
      output: "type T =\n  a\n  | b\n  | c",
    },
    {
      code: "function TSPropertySignatureToProperty(\n  node:\n  | TSESTree.TSEnumMember\n    | TSESTree.TSPropertySignature\n  | TSESTree.TypeElement,\n  type:\n  | AST_NODE_TYPES.Property\n    | AST_NODE_TYPES.PropertyDefinition = AST_NODE_TYPES.Property,\n): TSESTree.Node | null {}",
      output: "function TSPropertySignatureToProperty(\n  node:\n    | TSESTree.TSEnumMember\n    | TSESTree.TSPropertySignature\n    | TSESTree.TypeElement,\n  type:\n    | AST_NODE_TYPES.Property\n    | AST_NODE_TYPES.PropertyDefinition = AST_NODE_TYPES.Property,\n): TSESTree.Node | null {}",
    },
    {
      code: "type Foo = Merge<\n    A\n  & B\n    & C\n>",
      output: "type Foo = Merge<\n  A\n  & B\n  & C\n>",
    },
    {
      code: "if (\n  typeof woof === 'string' &&\n  typeof woof === 'string' &&\n    typeof woof === 'string' &&\n  isNaN(null) &&\n    isNaN(NaN)\n) {\n  return;\n}",
      output: "if (\n  typeof woof === 'string' &&\n  typeof woof === 'string' &&\n  typeof woof === 'string' &&\n  isNaN(null) &&\n  isNaN(NaN)\n) {\n  return;\n}",
    },
    {
      code: "const a = () => b\n|| c\n\nconst a = (\n  p,\n) => (b\n|| c)\n\nconst a = b\n+ c;\nconst a = {\n  p: b\n  + c,\n};",
      output: "const a = () => b\n  || c\n\nconst a = (\n  p,\n) => (b\n  || c)\n\nconst a = b\n  + c;\nconst a = {\n  p: b\n    + c,\n};",
    },
    {
      code: "const a = (\n  (b\n      && c)\n    || (d\n  && e)\n)",
      output: "const a = (\n  (b\n    && c)\n  || (d\n    && e)\n)",
    },
    {
      code: "const a = (\n  (b\n      && c)\n    || (d\n  && e)\n)",
      output: "const a = (\n  (b\n    && c)\n  || (d\n    && e)\n)",
    },
    {
      code: "{\n  const a = false\n  || (a && b)\n  || (c && d)\n  || (e && f)\n  || (g && h)\n}",
      output: "{\n  const a = false\n    || (a && b)\n    || (c && d)\n    || (e && f)\n    || (g && h)\n}",
    },
    {
      code: "type Type =\n  | ({\n    type: 'a';\n  } & A)\n  | ({\n    type: 'b';\n    } & B)\n  | ({\n    type: 'c';\n  } & {\n    c: string;\n  });",
      output: "type Type =\n  | ({\n    type: 'a';\n  } & A)\n  | ({\n    type: 'b';\n    } & B)\n    | ({\n    type: 'c';\n  } & {\n    c: string;\n  });",
    },
    {
      code: "type Foo = Pick<Bar,\nBaz\n    | Qux,\n>;",
      output: "type Foo = Pick<Bar,\n  Baz\n  | Qux,\n>;",
    },
    {
      code: "type Foo = [Bar,\nBaz\n      | Qux,\n];",
      output: "type Foo = [Bar,\n  Baz\n  | Qux,\n];",
    },
    {
      code: "type Foo = { x: Foo,\n  y: Baz\n  | Quz\n};",
      output: "type Foo = { x: Foo,\n  y: Baz\n    | Quz\n};",
    },
    {
      code: "type Foo = Pick<Bar\n| Baz,\nBaz\n    | Qux,\n>;",
      output: "type Foo = Pick<Bar\n  | Baz,\n  Baz\n  | Qux,\n>;",
    },
    {
      code: "type Foo = [Bar\n| Baz,\nBaz\n      | Qux,\n];",
      output: "type Foo = [Bar\n  | Baz,\n  Baz\n  | Qux,\n];",
    },
    {
      code: "type Foo = { x: Foo\n| Baz,\n  y: Baz\n  | Quz\n};",
      output: "type Foo = { x: Foo\n  | Baz,\n  y: Baz\n    | Quz\n};",
    },
    {
      code: "const a = 1\n+ 2\n    + 3;",
      output: "const a = 1\n  + 2\n  + 3;",
    },
    {
      code: "a = 1\n- 2\n    - 3;",
      output: "a = 1\n  - 2\n  - 3;",
    },
    {
      code: "const a = 1 *\n2 *\n    3;",
      output: "const a = 1 *\n  2 *\n  3;",
    },
    {
      code: "a = 1 /\n2 /\n    3;",
      output: "a = 1 /\n  2 /\n  3;",
    },
    {
      code: "this.a = this.b\n|| 2\n    || 3;",
      output: "this.a = this.b\n  || 2\n  || 3;",
    },
    {
      code: "{ aaaaa &&\n      bbbbb &&\n    ccccc }",
      output: "{ aaaaa &&\n  bbbbb &&\n  ccccc }",
    },
  ],
});
