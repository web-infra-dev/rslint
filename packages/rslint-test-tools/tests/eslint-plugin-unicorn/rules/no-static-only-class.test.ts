import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code });

ruleTester.run('no-static-only-class', null as never, {
  valid: [
    // Empty class
    valid('class A {}'),
    valid('const A = class {}'),
    // `superClass`
    valid('class A extends B { static a() {}; }'),
    valid('const A = class extends B { static a() {}; }'),
    // Not static
    valid('class A { a() {} }'),
    valid('class A { constructor() {} }'),
    valid('class A { get a() {} }'),
    valid('class A { set a(value) {} }'),
    // `private` (PrivateIdentifier on key)
    valid('class A3 { static #a() {}; }'),
    valid('class A3 { static #a = 1; }'),
    valid('const A3 = class { static #a() {}; }'),
    valid('const A3 = class { static #a = 1; }'),
    // Static block
    valid('class A2 { static {}; }'),
    // TS-only modifiers exclude the member from the "static only" check
    valid('class A { static public a = 1; }'),
    valid('class A { static private a = 1; }'),
    valid('class A { static readonly a = 1; }'),
    valid('class A { static declare a = 1; }'),
  ],
  invalid: [
    { code: 'class A { static a() {}; }', errors: 1 },
    { code: 'class A { static a() {} }', errors: 1 },
    { code: 'const A = class A { static a() {}; }', errors: 1 },
    { code: 'const A = class { static a() {}; }', errors: 1 },
    { code: 'class A { static constructor() {}; }', errors: 1 },
    { code: 'export default class A { static a() {}; }', errors: 1 },
    { code: 'export default class { static a() {}; }', errors: 1 },
    { code: 'export class A { static a() {}; }', errors: 1 },
    {
      code:
        'function a() {\n' +
        '\treturn class\n' +
        '\t{\n' +
        '\t\tstatic a() {}\n' +
        '\t}\n' +
        '}',
      errors: 1,
    },
    {
      code:
        'function a() {\n' +
        '\treturn class /* comment */\n' +
        '\t{\n' +
        '\t\tstatic a() {}\n' +
        '\t}\n' +
        '}',
      errors: 1,
    },
    {
      code:
        'function a() {\n' +
        '\treturn class // comment\n' +
        '\t{\n' +
        '\t\tstatic a() {}\n' +
        '\t}\n' +
        '}',
      errors: 1,
    },
    // Breaking edge cases
    {
      code: 'class A {static a(){}}\nclass B extends A {}',
      errors: 1,
    },
    {
      code: 'class A {static a(){}}\nconsole.log(typeof A)',
      errors: 1,
    },
    {
      code: 'class A {static a(){}}\nconst a = new A;',
      errors: 1,
    },

    // TS-specific
    {
      code:
        'class A {\n' +
        '\tstatic a\n' +
        '\tstatic b = 1\n' +
        '\tstatic [c] = 2\n' +
        '\tstatic [d]\n' +
        '\tstatic e() {}\n' +
        '\tstatic [f]() {}\n' +
        '}',
      errors: 1,
    },
    {
      code: 'class A {\n\tstatic a = 1;\n\tstatic b = this.a;\n}',
      errors: 1,
    },
    { code: 'class A {static [this.a] = 1}', errors: 1 },
    { code: 'declare class A { static a = 1; }', errors: 1 },
    { code: 'abstract class A { static a = 1; }', errors: 1 },
    { code: 'class A implements B { static a = 1; }', errors: 1 },
    {
      code:
        'class NotebookKernelProviderAssociationRegistry {\n' +
        '\tstatic extensionIds: (string | null)[] = [];\n' +
        '\tstatic extensionDescriptions: string[] = [];\n' +
        '}',
      errors: 1,
    },
  ],
});
