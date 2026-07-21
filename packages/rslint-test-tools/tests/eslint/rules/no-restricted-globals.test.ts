import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const customMessage = 'Use bar instead.';

ruleTester.run('no-restricted-globals', {
  valid: [
    'foo',
    { code: 'foo', options: ['bar'] as any },
    { code: 'var foo = 1;', options: ['foo'] as any },
    { code: 'event', options: ['bar'] as any },
    { code: "import foo from 'bar';", options: ['foo'] as any },
    { code: 'function foo() {}', options: ['foo'] as any },
    { code: 'function fn() { var foo; }', options: ['foo'] as any },
    { code: 'foo.bar', options: ['bar'] as any },
    {
      code: 'foo',
      options: [{ name: 'bar', message: 'Use baz instead.' }] as any,
    },
    { code: 'foo', options: [{ globals: ['bar'] }] as any },
    { code: 'const foo = 1', options: [{ globals: ['foo'] }] as any },
    { code: 'event', options: [{ globals: ['bar'] }] as any },
    {
      code: "import foo from 'bar';",
      options: [{ globals: ['foo'] }] as any,
    },
    { code: 'function foo() {}', options: [{ globals: ['foo'] }] as any },
    {
      code: 'function fn() { let foo; }',
      options: [{ globals: ['foo'] }] as any,
    },
    { code: 'foo.bar', options: [{ globals: ['bar'] }] as any },
    {
      code: 'foo',
      options: [
        { globals: [{ name: 'bar', message: 'Use baz instead.' }] },
      ] as any,
    },
    // checkGlobalObject defaults to false: plain global-object access is never checked
    { code: 'window.foo()', options: [{ globals: ['foo'] }] as any },
    { code: 'self.foo()', options: [{ globals: ['foo'] }] as any },
    { code: 'globalThis.foo()', options: [{ globals: ['foo'] }] as any },
    {
      code: 'myGlobal.foo()',
      options: [{ globals: ['foo'], globalObjects: ['myGlobal'] }] as any,
    },
    // checkGlobalObject: the restricted name must be the final property
    {
      code: 'foo.window.bar()',
      options: [{ globals: ['bar'], checkGlobalObject: true }] as any,
    },
    {
      code: 'foo.self.bar()',
      options: [{ globals: ['bar'], checkGlobalObject: true }] as any,
    },
    {
      code: 'foo.globalThis.bar()',
      options: [{ globals: ['bar'], checkGlobalObject: true }] as any,
    },
    {
      code: 'foo.myGlobal.bar()',
      options: [
        {
          globals: ['bar'],
          checkGlobalObject: true,
          globalObjects: ['myGlobal'],
        },
      ] as any,
    },
    // checkGlobalObject: a local shadowing declaration suppresses the report
    {
      code: 'let window; window.foo()',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
    },
    {
      code: 'let self; self.foo()',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
    },
    {
      code: 'let globalThis; globalThis.foo()',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
    },
    {
      code: 'let myGlobal; myGlobal.foo()',
      options: [
        {
          globals: ['foo'],
          checkGlobalObject: true,
          globalObjects: ['myGlobal'],
        },
      ] as any,
    },
    // TypeScript: type positions are never flagged
    { code: 'const foo: number = 1;', options: ['foo'] as any },
    { code: 'function foo(): void {}', options: ['foo'] as any },
    { code: 'function fn(): void { let foo; }', options: ['foo'] as any },
    {
      code: 'type Handler = (event: string) => any',
      options: ['event'] as any,
    },
    { code: 'let value: Test', options: ['Test'] as any },
    { code: 'class Derived implements Test {}', options: ['Test'] as any },
    { code: 'interface Derived extends Test {}', options: ['Test'] as any },
    { code: 'let value: NS.Test', options: ['NS'] as any },
    { code: 'let value: typeof Test', options: ['Test'] as any },
    { code: 'let value: Type<Test>', options: ['Type', 'Test'] as any },
  ],
  invalid: [
    {
      code: 'foo',
      options: ['foo'] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'function fn() { foo; }',
      options: ['foo'] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'event',
      options: ['foo', 'event'] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'foo()',
      options: ['foo'] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'foo.bar()',
      options: ['foo'] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'foo',
      options: [{ name: 'foo' }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'foo',
      options: [{ name: 'foo', message: customMessage }] as any,
      errors: [{ messageId: 'customMessage' }],
    },
    {
      code: "var foo = obj => hasOwnProperty(obj, 'name');",
      options: ['hasOwnProperty'] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'foo',
      options: [{ globals: ['foo'] }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'foo',
      options: [{ globals: [{ name: 'foo' }] }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'foo',
      options: [{ globals: [{ name: 'foo', message: customMessage }] }] as any,
      errors: [{ messageId: 'customMessage' }],
    },
    // checkGlobalObject: dot / bracket / optional-chain access
    {
      code: 'window.foo()',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'self.foo()',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'window.window.foo()',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'globalThis.foo()',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'myGlobal.foo()',
      options: [
        {
          globals: ['foo'],
          checkGlobalObject: true,
          globalObjects: ['myGlobal'],
        },
      ] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'window["foo"]',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'window?.foo()',
      options: [{ globals: ['foo'], checkGlobalObject: true }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    {
      code: 'window.foo(); myGlobal.foo()',
      options: [
        {
          globals: ['foo'],
          checkGlobalObject: true,
          globalObjects: ['myGlobal'],
        },
      ] as any,
      errors: [
        { messageId: 'defaultMessage' },
        { messageId: 'defaultMessage' },
      ],
    },
    {
      code: 'function onClick(event) { console.log(event); console.log(window.event); }',
      options: [{ globals: ['event'], checkGlobalObject: true }] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
    // TypeScript: value reference reported, type annotation is not
    {
      code: 'const x: Promise<any> = Promise.resolve();',
      options: ['Promise'] as any,
      errors: [{ messageId: 'defaultMessage' }],
    },
  ],
});
