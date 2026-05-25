import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';


import { getFixturesRootDir } from '../RuleTester';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

/**
 * Quote a string in "double quotes" because it’s painful
 * with a double-quoted string literal
 */
function q(str: string): string {
  return `"${str}"`;
}

ruleTester.run('dot-notation', {
  valid: [
    //  baseRule
    'a.b;',
    'a.b.c;',
    "a['12'];",
    'a[b];',
    'a[0];',
    { code: 'a.b.c;', options: [{ allowKeywords: false }] },
    { code: 'a.arguments;', options: [{ allowKeywords: false }] },
    { code: 'a.let;', options: [{ allowKeywords: false }] },
    { code: 'a.yield;', options: [{ allowKeywords: false }] },
    { code: 'a.eval;', options: [{ allowKeywords: false }] },
    { code: 'a[0];', options: [{ allowKeywords: false }] },
    { code: "a['while'];", options: [{ allowKeywords: false }] },
    { code: "a['true'];", options: [{ allowKeywords: false }] },
    { code: "a['null'];", options: [{ allowKeywords: false }] },
    { code: 'a[true];', options: [{ allowKeywords: false }] },
    { code: 'a[null];', options: [{ allowKeywords: false }] },
    { code: 'a.true;', options: [{ allowKeywords: true }] },
    { code: 'a.null;', options: [{ allowKeywords: true }] },
    {
      code: "a['snake_case'];",
      options: [{ allowPattern: '^[a-z]+(_[a-z]+)+$' }],
    },
    {
      code: "a['lots_of_snake_case'];",
      options: [{ allowPattern: '^[a-z]+(_[a-z]+)+$' }],
    },
    {
      code: 'a[`time${range}`];',
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    {
      code: 'a[`while`];',
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      options: [{ allowKeywords: false }],
    },
    {
      code: 'a[`time range`];',
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
    },
    'a.true;',
    'a.null;',
    'a[undefined];',
    'a[void 0];',
    'a[b()];',
    {
      code: 'a[/(?<zero>0)/];',
      languageOptions: { parserOptions: { ecmaVersion: 2018 } },
    },

    {
      code: `
class X {
  private priv_prop = 123;
}

const x = new X();
x['priv_prop'] = 123;
      `,
      options: [{ allowPrivateClassPropertyAccess: true }],
    },

    {
      code: `
class X {
  protected protected_prop = 123;
}

const x = new X();
x['protected_prop'] = 123;
      `,
      options: [{ allowProtectedClassPropertyAccess: true }],
    },
    {
      code: `
class X {
  prop: string;
  [key: string]: number;
}

const x = new X();
x['hello'] = 3;
      `,
      options: [{ allowIndexSignaturePropertyAccess: true }],
    },
    {
      code: `
interface Nested {
  property: string;
  [key: string]: number | string;
}

class Dingus {
  nested: Nested;
}

let dingus: Dingus | undefined;

dingus?.nested.property;
dingus?.nested['hello'];
      `,
      languageOptions: { parserOptions: { ecmaVersion: 2020 } },
      options: [{ allowIndexSignaturePropertyAccess: true }],
    },
    {
      code: `
class X {
  private priv_prop = 123;
}

let x: X | undefined;
console.log(x?.['priv_prop']);
      `,
      options: [{ allowPrivateClassPropertyAccess: true }],
    },
    {
      code: `
class X {
  protected priv_prop = 123;
}

let x: X | undefined;
console.log(x?.['priv_prop']);
      `,
      options: [{ allowProtectedClassPropertyAccess: true }],
    },
    {
      code: `
type Foo = {
  bar: boolean;
  [key: \`key_\${string}\`]: number;
};
declare const foo: Foo;
foo['key_baz'];
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noPropertyAccessFromIndexSignature.json',
          projectService: false,
          tsconfigRootDir: rootPath,
        },
      },
    },
    {
      code: `
type Key = Lowercase<string>;
type Foo = {
  BAR: boolean;
  [key: Lowercase<string>]: number;
};
declare const foo: Foo;
foo['bar'];
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noPropertyAccessFromIndexSignature.json',
          projectService: false,
          tsconfigRootDir: rootPath,
        },
      },
    },
    {
      code: `
type ExtraKey = \`extra\${string}\`;

type Foo = {
  foo: string;
  [extraKey: ExtraKey]: number;
};

function f<T extends Foo>(x: T) {
  x['extraKey'];
}
      `,
      languageOptions: {
        parserOptions: {
          project: './tsconfig.noPropertyAccessFromIndexSignature.json',
          projectService: false,
          tsconfigRootDir: rootPath,
        },
      },
    },

    // --- Non-identifier string keys (ESLint's ASCII-only regex filters out) ---
    "a['with-dash'];",
    "a['has space'];",
    "a[''];",
    "a['12valid'];",
    "a['\\n'];",
    "a['it\\'s'];",
    { code: "a['$ok'];", options: [{ allowPattern: '^\\$' }] },

    // --- Non-ASCII identifier-looking strings are not valid identifiers ---
    "a['ñ'];",
    "a['中文'];",
    "a['café'];",

    // --- Numeric / bigint literal keys are not string-literal-like ---
    'a[0x1];',
    'a[42];',
    'a[42n];',

    // --- allowPattern may cover null/true/false literal keys ---
    {
      code: "a['null'];",
      options: [{ allowPattern: '^(null|true|false)$' }],
    },
    {
      code: 'a[null];',
      options: [{ allowPattern: '^null$' }],
    },

    // --- Private-identifier (#name) class field access is not a regular Identifier ---
    `
class X {
  #priv = 1;
  foo() { return this.#priv; }
}
    `,
    {
      code: `
class X {
  #priv = 1;
  foo() { return this.#priv; }
}
      `,
      options: [{ allowKeywords: false }],
    },

    // --- readonly / nullable-valued index signatures still count as string-like ---
    {
      code: `
type RO = { readonly [k: string]: number };
declare const m: RO;
m['foo'];
      `,
      options: [{ allowIndexSignaturePropertyAccess: true }],
    },
    {
      code: `
type M = { [k: string]: number | undefined };
declare const m: M;
m['foo'];
      `,
      options: [{ allowIndexSignaturePropertyAccess: true }],
    },

    // --- Generic constraint carries the index signature ---
    {
      code: `
function f<T extends { a: number; [k: string]: number }>(x: T) {
  x['b'];
}
      `,
      options: [{ allowIndexSignaturePropertyAccess: true }],
    },

    // --- `this` access on a class with an index signature ---
    {
      code: `
class X {
  [k: string]: number;
  foo() { return this['bar']; }
}
      `,
      options: [{ allowIndexSignaturePropertyAccess: true }],
    },

    // --- Decorator argument listener fires normally; allowPattern filters it ---
    {
      code: `
declare function d(v: unknown): ClassDecorator;
@d((undefined as any)['_ok_'])
class C {}
      `,
      options: [{ allowPattern: '^_' }],
    },

    // --- `as Record<string, T>` cast exposes the index signature ---
    {
      code: `
declare const x: unknown;
(x as Record<string, number>)['foo'];
      `,
      options: [{ allowIndexSignaturePropertyAccess: true }],
    },
  ],
  invalid: [
    {
      code: `
class X {
  private priv_prop = 123;
}

const x = new X();
x['priv_prop'] = 123;
      `,
      errors: [{ messageId: 'useDot' }],
      options: [{ allowPrivateClassPropertyAccess: false }],
      output: `
class X {
  private priv_prop = 123;
}

const x = new X();
x.priv_prop = 123;
      `,
    },
    {
      code: `
class X {
  public pub_prop = 123;
}

const x = new X();
x['pub_prop'] = 123;
      `,
      errors: [{ messageId: 'useDot' }],
      output: `
class X {
  public pub_prop = 123;
}

const x = new X();
x.pub_prop = 123;
      `,
    },
    //  baseRule

    // {
    //     code: 'a.true;',
    //     output: "a['true'];",
    //     options: [{ allowKeywords: false }],
    //     errors: [{ messageId: "useBrackets", data: { key: "true" } }],
    // },
    {
      code: "a['true'];",
      errors: [{ data: { key: q('true') }, messageId: 'useDot' }],
      output: 'a.true;',
    },
    {
      code: "a['time'];",
      errors: [{ data: { key: '"time"' }, messageId: 'useDot' }],
      languageOptions: { parserOptions: { ecmaVersion: 6 } },
      output: 'a.time;',
    },
    {
      code: 'a[null];',
      errors: [{ data: { key: 'null' }, messageId: 'useDot' }],
      output: 'a.null;',
    },
    {
      code: 'a[true];',
      errors: [{ data: { key: 'true' }, messageId: 'useDot' }],
      output: 'a.true;',
    },
    {
      code: 'a[false];',
      errors: [{ data: { key: 'false' }, messageId: 'useDot' }],
      output: 'a.false;',
    },
    {
      code: "a['b'];",
      errors: [{ data: { key: q('b') }, messageId: 'useDot' }],
      output: 'a.b;',
    },
    {
      code: "a.b['c'];",
      errors: [{ data: { key: q('c') }, messageId: 'useDot' }],
      output: 'a.b.c;',
    },
    {
      code: "a['_dangle'];",
      errors: [{ data: { key: q('_dangle') }, messageId: 'useDot' }],
      options: [{ allowPattern: '^[a-z]+(_[a-z]+)+$' }],
      output: 'a._dangle;',
    },
    {
      code: "a['SHOUT_CASE'];",
      errors: [{ data: { key: q('SHOUT_CASE') }, messageId: 'useDot' }],
      options: [{ allowPattern: '^[a-z]+(_[a-z]+)+$' }],
      output: 'a.SHOUT_CASE;',
    },
    {
      code: noFormat`
a
  ['SHOUT_CASE'];
      `,
      errors: [
        {
          column: 4,
          data: { key: q('SHOUT_CASE') },
          line: 3,
          messageId: 'useDot',
        },
      ],
      output: `
a
  .SHOUT_CASE;
      `,
    },
    {
      code:
        'getResource()\n' +
        '    .then(function(){})\n' +
        '    ["catch"](function(){})\n' +
        '    .then(function(){})\n' +
        '    ["catch"](function(){});',
      errors: [
        {
          column: 6,
          data: { key: q('catch') },
          line: 3,
          messageId: 'useDot',
        },
        {
          column: 6,
          data: { key: q('catch') },
          line: 5,
          messageId: 'useDot',
        },
      ],
      output:
        'getResource()\n' +
        '    .then(function(){})\n' +
        '    .catch(function(){})\n' +
        '    .then(function(){})\n' +
        '    .catch(function(){});',
    },
    {
      code: noFormat`
foo
  .while;
      `,
      errors: [{ data: { key: 'while' }, messageId: 'useBrackets' }],
      options: [{ allowKeywords: false }],
      output: `
foo
  ["while"];
      `,
    },
    {
      code: "foo[/* comment */ 'bar'];",
      errors: [{ data: { key: q('bar') }, messageId: 'useDot' }],
      output: null, // Not fixed due to comment
    },
    {
      code: "foo['bar' /* comment */];",
      errors: [{ data: { key: q('bar') }, messageId: 'useDot' }],
      output: null, // Not fixed due to comment
    },
    {
      code: "foo['bar'];",
      errors: [{ data: { key: q('bar') }, messageId: 'useDot' }],
      output: 'foo.bar;',
    },
    {
      code: 'foo./* comment */ while;',
      errors: [{ data: { key: 'while' }, messageId: 'useBrackets' }],
      options: [{ allowKeywords: false }],
      output: null, // Not fixed due to comment
    },
    {
      code: 'foo[null];',
      errors: [{ data: { key: 'null' }, messageId: 'useDot' }],
      output: 'foo.null;',
    },
    {
      code: "foo['bar'] instanceof baz;",
      errors: [{ data: { key: q('bar') }, messageId: 'useDot' }],
      output: 'foo.bar instanceof baz;',
    },
    {
      code: 'let.if();',
      errors: [{ data: { key: 'if' }, messageId: 'useBrackets' }],
      options: [{ allowKeywords: false }],
      output: null, // `let["if"]()` is a syntax error because `let[` indicates a destructuring variable declaration
    },
    {
      code: `
class X {
  protected protected_prop = 123;
}

const x = new X();
x['protected_prop'] = 123;
      `,
      errors: [{ messageId: 'useDot' }],
      options: [{ allowProtectedClassPropertyAccess: false }],
      output: `
class X {
  protected protected_prop = 123;
}

const x = new X();
x.protected_prop = 123;
      `,
    },
    {
      code: `
class X {
  prop: string;
  [key: string]: number;
}

const x = new X();
x['prop'] = 'hello';
      `,
      errors: [{ messageId: 'useDot' }],
      options: [{ allowIndexSignaturePropertyAccess: true }],
      output: `
class X {
  prop: string;
  [key: string]: number;
}

const x = new X();
x.prop = 'hello';
      `,
    },
    {
      code: `
type Foo = {
  bar: boolean;
  [key: \`key_\${string}\`]: number;
};
foo['key_baz'];
      `,
      errors: [{ messageId: 'useDot' }],
      output: `
type Foo = {
  bar: boolean;
  [key: \`key_\${string}\`]: number;
};
foo.key_baz;
      `,
    },
    {
      code: `
type ExtraKey = \`extra\${string}\`;

type Foo = {
  foo: string;
  [extraKey: ExtraKey]: number;
};

function f<T extends Foo>(x: T) {
  x['extraKey'];
}
      `,
      errors: [{ messageId: 'useDot' }],
      output: `
type ExtraKey = \`extra\${string}\`;

type Foo = {
  foo: string;
  [extraKey: ExtraKey]: number;
};

function f<T extends Foo>(x: T) {
  x.extraKey;
}
      `,
    },
    // Number-only index signature is NOT string-like — must still report.
    {
      code: `
type NumMap = { [k: number]: number };
declare const m: NumMap;
m['foo'];
      `,
      errors: [{ messageId: 'useDot' }],
      options: [{ allowIndexSignaturePropertyAccess: true }],
      output: `
type NumMap = { [k: number]: number };
declare const m: NumMap;
m.foo;
      `,
    },
    // Concrete named property still reported even if an index signature exists.
    {
      code: `
interface WithIdx { bar: number; [k: string]: number }
declare const m: WithIdx;
m['bar'];
      `,
      errors: [{ messageId: 'useDot' }],
      options: [{ allowIndexSignaturePropertyAccess: true }],
      output: `
interface WithIdx { bar: number; [k: string]: number }
declare const m: WithIdx;
m.bar;
      `,
    },

    // --- Unicode escape in string literal — `.Text` returns the unescaped
    // value, which happens to be a plain identifier; autofix produces `.b`. ---
    {
      code: "a['\\u0062'];",
      errors: [{ messageId: 'useDot' }],
      output: 'a.b;',
    },

    // --- allowKeywords:false + dot access to null/true/false keyword literal ---
    {
      code: 'a.null;',
      errors: [{ data: { key: 'null' }, messageId: 'useBrackets' }],
      options: [{ allowKeywords: false }],
      output: 'a["null"];',
    },
    {
      code: 'a.true;',
      errors: [{ data: { key: 'true' }, messageId: 'useBrackets' }],
      options: [{ allowKeywords: false }],
      output: 'a["true"];',
    },
    {
      code: 'a.false;',
      errors: [{ data: { key: 'false' }, messageId: 'useBrackets' }],
      options: [{ allowKeywords: false }],
      output: 'a["false"];',
    },

    // --- Assignment targets: simple / compound / update expressions ---
    {
      code: "a['foo'] = 1;",
      errors: [{ data: { key: '"foo"' }, messageId: 'useDot' }],
      output: 'a.foo = 1;',
    },
    {
      code: "a['foo'] += 1;",
      errors: [{ data: { key: '"foo"' }, messageId: 'useDot' }],
      output: 'a.foo += 1;',
    },
    {
      code: "a['foo']++;",
      errors: [{ data: { key: '"foo"' }, messageId: 'useDot' }],
      output: 'a.foo++;',
    },

    // --- `any`-typed bracket access still reported even with allowIndexSig ---
    {
      code: `
declare const x: any;
x['foo'];
      `,
      errors: [{ messageId: 'useDot' }],
      options: [{ allowIndexSignaturePropertyAccess: true }],
      output: `
declare const x: any;
x.foo;
      `,
    },

    // --- `this['member']` on concrete class member ---
    {
      code: `
class X {
  a = 1;
  foo() { return this['a']; }
}
      `,
      errors: [{ messageId: 'useDot' }],
      output: `
class X {
  a = 1;
  foo() { return this.a; }
}
      `,
    },

    // --- `super['member']` access ---
    {
      code: `
class B { foo = 1; }
class X extends B {
  bar() { return super['foo']; }
}
      `,
      errors: [{ messageId: 'useDot' }],
      output: `
class B { foo = 1; }
class X extends B {
  bar() { return super.foo; }
}
      `,
    },

    // --- JSX expression container — listener fires inside JSX ---
    {
      code: `
declare const p: { foo: string };
const el = <div>{p['foo']}</div>;
      `,
      errors: [{ messageId: 'useDot' }],
      languageOptions: {
        parserOptions: { ecmaFeatures: { jsx: true } },
      },
      output: `
declare const p: { foo: string };
const el = <div>{p.foo}</div>;
      `,
    },

    // --- Namespace / module declaration nesting ---
    {
      code: `
namespace N {
  declare const m: { foo: number };
  m['foo'];
}
      `,
      errors: [{ messageId: 'useDot' }],
      output: `
namespace N {
  declare const m: { foo: number };
  m.foo;
}
      `,
    },
  ],
});
