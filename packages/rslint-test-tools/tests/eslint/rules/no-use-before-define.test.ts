import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors Layer 1 only — the upstream valid/invalid semantic set. tsgo edge
// shapes, upstream-branch lock-ins, and issue-tracker shapes live in the Go
// suite (internal/rules/no_use_before_define/*_extras_test.go).
ruleTester.run('no-use-before-define', {
  valid: [
    'unresolved',
    'Array',
    'function foo () { arguments; }',
    'var a=10; alert(a);',
    'function b(a) { alert(a); }',
    'Object.hasOwnProperty.call(a);',
    'a(); try { throw new Error() } catch (a) {}',
    'class A {} new A();',
    'var a = 0, b = a;',
    'var {a = 0, b = a} = {};',
    'var [a = 0, b = a] = {};',
    'function foo() { foo(); }',
    'var foo = function() { foo(); };',
    'var a; for (a in a) {}',
    'var a; for (a of a) {}',
    'let a; class C { static { a; } }',
    'class C { static { let a; a; } }',

    // Block-level bindings
    '"use strict"; a(); { function a() {} }',
    'switch (foo) { case 1:  { a(); } default: { let a; }}',
    'a(); { let a = function () {}; }',

    // "nofunc" / object style options
    {
      code: 'a(); function a() { alert(arguments); }',
      options: ['nofunc'] as any,
    },
    {
      code: 'a(); function a() { alert(arguments); }',
      options: [{ functions: false }] as any,
    },
    {
      code: '"use strict"; { a(); function a() {} }',
      options: [{ functions: false }] as any,
    },
    {
      code: 'function foo() { new A(); } class A {};',
      options: [{ classes: false }] as any,
    },

    // "variables" option
    {
      code: 'function foo() { bar; } var bar;',
      options: [{ variables: false }] as any,
    },
    {
      code: 'var foo = () => bar; var bar;',
      options: [{ variables: false }] as any,
    },
    {
      code: 'class C { static { () => foo; let foo; } }',
      options: [{ variables: false }] as any,
    },

    // Class definition evaluation — not TDZ errors
    'class C extends (class { method() { C; } }) {}',
    'class C extends (class { field = C; }) {}',
    'class C { [() => C](){} }',
    'class C { method() { C; } }',
    'class C { field = C; }',
    'class C { static field = C; }',
    'class C { field = class extends C {}; }',
    'class C { static field = class { [C]; }; }',
    'const C = class { static field = class { field = C; }; };',
    'class C { static { C; } }',
    'class C { static { class D extends C {} } }',
    'class C { static { () => C; } }',
    'const C = class C { static { C.x; } }',
    {
      code: 'class C { field = a; } let a;',
      options: [{ variables: false }] as any,
    },
    {
      code: 'class C { field = D; } class D {}',
      options: [{ classes: false }] as any,
    },

    // "allowNamedExports" option
    {
      code: 'export { a }; const a = 1;',
      options: [{ allowNamedExports: true }] as any,
    },
    {
      code: 'export { a as b }; const a = 1;',
      options: [{ allowNamedExports: true }] as any,
    },
    {
      code: 'export { f }; function f() {}',
      options: [{ allowNamedExports: true }] as any,
    },
    {
      code: 'export { C }; class C {}',
      options: [{ allowNamedExports: true }] as any,
    },

    // TypeScript syntax
    'type foo = 1;\nconst x: foo = 1;',
    'type foo = 1;\ntype bar = foo;',
    'interface Foo {}\nconst x: Foo = {};',
    'declare function a();',
    'declare class a {\n  foo();\n}',
    'const updatedAt = data?.updatedAt;',
    'var a = { b: 5 };\nalert(a?.b);',
    'interface Foo {\n  bar: string;\n}\nconst bar = "blah";',
    'type T = (value: unknown) => value is Id;',
    'namespace A.X.Y {}\n\nimport Z = A.X.Y;\n\nconst X = 23;',

    // "typedefs" option
    {
      code: 'var x: Foo = 2;\ntype Foo = string | number;',
      options: [{ typedefs: false }] as any,
    },
    {
      code: 'var x: Foo = {};\ninterface Foo {}',
      options: [{ typedefs: false, ignoreTypeReferences: false }] as any,
    },

    // "ignoreTypeReferences" option
    {
      code: 'interface Bar {\n  type: typeof Foo;\n}\n\nconst Foo = 2;',
      options: [{ ignoreTypeReferences: true }] as any,
    },
    {
      code: 'interface Bar {\n  type: typeof Foo.Bar.Baz;\n}\n\nconst Foo = { Bar: { Baz: 1 } };',
      options: [{ ignoreTypeReferences: true }] as any,
    },

    // "enums" option
    {
      code: 'function foo(): Foo {\n  return Foo.FOO;\n}\n\nenum Foo {\n  FOO,\n}',
      options: [{ enums: false }] as any,
    },
    {
      code: 'let foo: Foo;\n\nenum Foo {\n  FOO,\n}',
      options: [{ enums: false }] as any,
    },

    // Decorators run after the class binding is initialized
    '@Directive({ providers: [{ useExisting: CidrIpPatternDirective }] })\nexport class CidrIpPatternDirective {}',

    // JSX
    { code: 'const App = () => <div/>; <App />;', filename: 'src/virtual.tsx' },
    { code: 'let Foo, Bar; <Foo><Bar /></Foo>;', filename: 'src/virtual.tsx' },
    {
      code: 'function App() { return <div/> } <App />;',
      filename: 'src/virtual.tsx',
    },
    {
      code: '<App />; function App() { return <div/> }',
      options: [{ functions: false }] as any,
      filename: 'src/virtual.tsx',
    },
  ],
  invalid: [
    {
      code: 'a++; var a=19;',
      errors: [
        {
          messageId: 'usedBeforeDefined',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'a(); var a=function() {};',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 1 }],
    },
    {
      code: 'alert(a[1]); var a=[1,3];',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 7 }],
    },
    {
      code: 'a(); function a() { alert(b); var b=10; a(); }',
      errors: [
        { messageId: 'usedBeforeDefined', line: 1, column: 1 },
        { messageId: 'usedBeforeDefined', line: 1, column: 27 },
      ],
    },
    {
      code: 'a(); var a=function() {};',
      options: ['nofunc'] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 1 }],
    },
    {
      code: '(() => { alert(a); var a = 42; })();',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 16 }],
    },
    {
      code: '(() => a())(); function a() { }',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 8 }],
    },
    {
      code: 'a(); try { throw new Error() } catch (foo) {var a;}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 1 }],
    },
    {
      code: 'var f = () => a; var a;',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 15 }],
    },
    {
      code: 'new A(); class A {};',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 5 }],
    },
    {
      code: 'function foo() { new A(); } class A {};',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 22 }],
    },
    {
      code: 'new A(); var A = class {};',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 5 }],
    },

    // Block-level bindings
    {
      code: 'a++; { var a; }',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 1 }],
    },
    {
      code: '"use strict"; { a(); function a() {} }',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 17 }],
    },
    {
      code: '{a; let a = 1}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 2 }],
    },
    {
      code: 'switch (foo) { case 1: a();\n default: \n let a;}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 24 }],
    },
    {
      code: 'if (true) { function foo() { a; } let a;}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 30 }],
    },

    // object style options
    {
      code: 'a(); var a=function() {};',
      options: [{ functions: false, classes: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 1 }],
    },
    {
      code: 'new A(); class A {};',
      options: [{ functions: false, classes: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 5 }],
    },

    // invalid initializers
    {
      code: 'var a = a;',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 9 }],
    },
    {
      code: 'const a = foo(a);',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 15 }],
    },
    {
      code: 'function foo(a = a) {}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 18 }],
    },
    {
      code: 'var {a = a} = [];',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 10 }],
    },
    {
      code: 'var {b = a, a} = {};',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 10 }],
    },
    {
      code: 'var {a = 0} = a;',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 15 }],
    },
    {
      code: 'for (var a in a) {}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 15 }],
    },
    {
      code: 'for (var a of a) {}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 15 }],
    },

    // "variables" option
    {
      code: 'function foo() { bar; var bar = 1; } var bar;',
      options: [{ variables: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 18 }],
    },
    {
      code: 'foo; var foo;',
      options: [{ variables: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 1 }],
    },

    // https://github.com/eslint/eslint/issues/10227
    {
      code: 'for (let x = x;;); let x = 0',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 14 }],
    },
    {
      code: 'for (let x in xs); let xs = []',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 15 }],
    },
    {
      code: "try {} catch ({message = x}) {} let x = ''",
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 26 }],
    },
    {
      code: 'with (obj) x; let x = {}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 12 }],
    },

    // Class definition evaluation — TDZ errors
    {
      code: 'class C extends C {}',
      options: [{ classes: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 17 }],
    },
    {
      code: 'const C = class extends C {};',
      options: [{ variables: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 25 }],
    },
    {
      code: 'class C { [C](){} }',
      options: [{ classes: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 12 }],
    },
    {
      code: 'const C = class { static field = C; };',
      options: [{ variables: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 34 }],
    },
    {
      code: 'class C { static field = a; } let a;',
      options: [{ variables: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 26 }],
    },
    {
      code: 'class C { static { a; } } let a;',
      options: [{ variables: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 20 }],
    },

    // "allowNamedExports" option
    {
      code: 'export { a }; const a = 1;',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 10 }],
    },
    {
      code: 'export { a as b }; const a = 1;',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 10 }],
    },
    {
      code: 'export { a, b }; let a, b;',
      errors: [
        { messageId: 'usedBeforeDefined', line: 1, column: 10 },
        { messageId: 'usedBeforeDefined', line: 1, column: 13 },
      ],
    },
    {
      code: 'export const foo = a; const a = 1;',
      options: [{ allowNamedExports: true }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 20 }],
    },

    // TypeScript-specific options
    {
      code: 'interface Bar {\n  type: typeof Foo;\n}\n\nconst Foo = 2;',
      options: [{ ignoreTypeReferences: false }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 2, column: 16 }],
    },
    {
      code: 'let var1: StringOrNumber;\n\ntype StringOrNumber = string | number;',
      options: [{ ignoreTypeReferences: false, typedefs: true }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 11 }],
    },
    {
      code: 'const foo = Foo.Foo;\n\nenum Foo {\n  FOO,\n}',
      options: [{ enums: true }] as any,
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 13 }],
    },
    {
      code: 'export { Foo };\n\nenum Foo {\n  BAR,\n}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 10 }],
    },
    {
      code: 'export { Foo };\n\nnamespace Foo {\n  export let bar = () => 1;\n}',
      errors: [{ messageId: 'usedBeforeDefined', line: 1, column: 10 }],
    },
    {
      code: '@decorator\nclass C {\n  static x = "foo";\n  [C.x]() {}\n}',
      errors: [{ messageId: 'usedBeforeDefined', line: 4, column: 4 }],
    },

    // JSX
    {
      code: '<App />; const App = () => <div />;',
      filename: 'src/virtual.tsx',
      errors: [
        {
          messageId: 'usedBeforeDefined',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: '<Foo.Bar />; const Foo = { Bar: () => <div/> };',
      filename: 'src/virtual.tsx',
      errors: [
        {
          messageId: 'usedBeforeDefined',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
  ],
});
