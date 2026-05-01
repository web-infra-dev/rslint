import { RuleTester } from '@typescript-eslint/rule-tester';



const ruleTester = new RuleTester();

ruleTester.run('max-params', {
  valid: [
    'function foo() {}',
    'const foo = function () {};',
    'const foo = () => {};',
    'function foo(a) {}',
    `
class Foo {
  constructor(a) {}
}
    `,
    `
class Foo {
  method(this: void, a, b, c) {}
}
    `,
    `
class Foo {
  method(this: Foo, a, b) {}
}
    `,
    {
      code: 'function foo(a, b, c, d) {}',
      options: [{ max: 4 }],
    },
    {
      code: 'function foo(a, b, c, d) {}',
      options: [{ maximum: 4 }],
    },
    {
      code: `
class Foo {
  method(this: void) {}
}
      `,
      options: [{ max: 0 }],
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      options: [{ countVoidThis: true, max: 2 }],
    },
    {
      code: `
declare function makeDate(m: number, d: number, y: number): Date;
      `,
      options: [{ max: 3 }],
    },
    {
      code: `
type sum = (a: number, b: number) => number;
      `,
      options: [{ max: 2 }],
    },
  ],
  invalid: [
    { code: 'function foo(a, b, c, d) {}', errors: [{ messageId: 'exceed' }] },
    {
      code: 'const foo = function (a, b, c, d) {};',
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'const foo = (a, b, c, d) => {};',
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'const foo = a => {};',
      errors: [{ messageId: 'exceed' }],
      options: [{ max: 0 }],
    },
    {
      code: `
class Foo {
  method(this: void, a, b, c, d) {}
}
      `,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      errors: [{ messageId: 'exceed' }],
      options: [{ countVoidThis: true, max: 1 }],
    },
    {
      code: `
class Foo {
  method(this: Foo, a, b, c) {}
}
      `,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
declare function makeDate(m: number, d: number, y: number): Date;
      `,
      errors: [{ messageId: 'exceed' }],
      options: [{ max: 1 }],
    },
    {
      code: `
type sum = (a: number, b: number) => number;
      `,
      errors: [{ messageId: 'exceed' }],
      options: [{ max: 1 }],
    },

    // ----- rslint additions: tsgo container kinds + name-resolution corners
    // (each of these locks a description / position the upstream test suite
    // never exercises because ESTree collapses these forms into
    // FunctionExpression / ArrowFunctionExpression). -----

    // Object-literal property whose value is a FunctionExpression — described
    // as "method 'foo'" (matches ESLint), not "function 'foo'".
    {
      code: 'var obj = { foo: function (a, b, c, d) {} };',
      errors: [{ messageId: 'exceed' }],
    },
    // Async object-literal property method via FunctionExpression form.
    {
      code: 'var obj = { foo: async function (a, b, c, d) {} };',
      errors: [{ messageId: 'exceed' }],
    },
    // Arrow as object-literal property value — "Arrow function 'foo'".
    {
      code: 'var obj = { foo: (a, b, c, d) => {} };',
      errors: [{ messageId: 'exceed' }],
    },
    // Class field initialized with FunctionExpression — described as
    // "Function 'foo'" (ESLint's `MethodDefinition` branch does NOT match
    // class fields, which are PropertyDefinition).
    {
      code: 'class C { foo = function (a, b, c, d) {}; }',
      errors: [{ messageId: 'exceed' }],
    },
    // declare class methods (no body) and constructors still trigger.
    {
      code: `
declare class Foo {
  method(a, b, c, d): void;
}
      `,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
declare class Foo {
  static method(a, b, c, d): void;
  constructor(a, b, c, d);
}
      `,
      errors: [{ messageId: 'exceed' }, { messageId: 'exceed' }],
    },
    // Function type as the value of a TypeAliasDeclaration — name resolved
    // from the alias.
    {
      code: 'type F = (a: number, b: number, c: number, d: number) => void;',
      errors: [{ messageId: 'exceed' }],
    },
    // Function type inside an interface PropertySignature — name resolved
    // from the property.
    {
      code: 'interface I { foo: (a, b, c, d) => void }',
      errors: [{ messageId: 'exceed' }],
    },
    // Private + static + async + generator method — modifier ordering and
    // private-identifier name preserved.
    {
      code: 'class C { static async *#m(a, b, c, d) {} }',
      errors: [{ messageId: 'exceed' }],
    },
    // Parameter properties count as parameters (TS-only, would not appear
    // in upstream eslint-core tests).
    {
      code: 'class C { constructor(public a: number, private b: string, readonly c: boolean, d: any) {} }',
      errors: [{ messageId: 'exceed' }],
    },
    // Overload signatures — each declaration counted independently.
    {
      code: `
class Foo {
  m(a, b): void;
  m(a, b, c, d): void;
  m(...args: any[]): any {}
}
      `,
      errors: [{ messageId: 'exceed' }],
    },
    // this:void + rest — stripped, rest still counts as 1.
    {
      code: 'function f(this: void, a, b, c, ...rest) {}',
      errors: [{ messageId: 'exceed' }],
    },

    // Class field with PrivateIdentifier / static — ESLint adds the
    // `private` / `static` prefix and resolves the name from the field key.
    {
      code: 'class C { #foo = (a, b, c, d) => {}; }',
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'class C { static foo = (a, b, c, d) => {}; }',
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'class C { static #foo = (a, b, c, d) => {}; }',
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'class C { #foo = function (a, b, c, d) {}; }',
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'class C { static #foo = async function (a, b, c, d) {}; }',
      errors: [{ messageId: 'exceed' }],
    },
  ],
});
