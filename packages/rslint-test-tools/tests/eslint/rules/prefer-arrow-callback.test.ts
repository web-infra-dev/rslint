import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const errors = [{ messageId: 'preferArrowCallback' }];

ruleTester.run('prefer-arrow-callback', {
  valid: [
    'foo(a => a);',
    'foo(function*() {});',
    'foo(function() { this; });',
    {
      code: 'foo(function bar() {});',
      options: [{ allowNamedFunctions: true }] as any,
    },
    'foo(function() { (() => this); });',
    'foo(function() { this; }.bind(obj));',
    'foo(function() { this; }.call(this));',
    'foo(a => { (function() {}); });',
    'var foo = function foo() {};',
    '(function foo() {})();',
    'foo(function bar() { bar; });',
    'foo(function bar() { arguments; });',
    'foo(function bar() { arguments; }.bind(this));',
    'foo(function bar() { new.target; });',
    'foo(function bar() { new.target; }.bind(this));',
    'foo(function bar() { this; }.bind(this, somethingElse));',
    'foo((function() {}).bind.bar)',
    'foo((function() { this.bar(); }).bind(obj).bind(this))',

    // TypeScript upstream valid cases
    'foo((a:string) => a);',
    {
      code: 'foo(function bar(a:string) {});',
      options: [{ allowNamedFunctions: true }] as any,
    },
    "test('clean', function (this: any) { this.foo = 'Cleaned!';});",
    "obj.test('clean', function (foo) { this.foo = 'Cleaned!'; });",
  ],
  invalid: [
    {
      code: 'foo(function bar() {});',
      errors,
    },
    {
      code: 'foo(function() {});',
      options: [{ allowNamedFunctions: true }] as any,
      errors,
    },
    {
      code: 'foo(function bar() {});',
      options: [{ allowNamedFunctions: false }] as any,
      errors,
    },
    {
      code: 'foo(function() {});',
      errors,
    },
    {
      code: 'foo(nativeCb || function() {});',
      errors,
    },
    {
      code: 'foo(bar ? function() {} : function() {});',
      errors: [errors[0], errors[0]],
    },
    {
      code: 'foo(function() { (function() { this; }); });',
      errors,
    },
    {
      code: 'foo(function() { this; }.bind(this));',
      errors,
    },
    {
      code: 'foo(bar || function() { this; }.bind(this));',
      errors,
    },
    {
      code: 'foo(function() { (() => this); }.bind(this));',
      errors,
    },
    {
      code: 'foo(function bar(a) { a; });',
      errors,
    },
    {
      code: 'foo(function(a) { a; });',
      errors,
    },
    {
      code: 'foo(function(arguments) { arguments; });',
      errors,
    },
    {
      code: 'foo(function() { this; });',
      options: [{ allowUnboundThis: false }] as any,
      errors,
    },
    {
      code: 'foo(function() { (() => this); });',
      options: [{ allowUnboundThis: false }] as any,
      errors,
    },
    {
      code: 'qux(function(foo, bar, baz) { return foo * 2; })',
      errors,
    },
    {
      code: 'qux(function(foo, bar, baz) { return foo * bar; }.bind(this))',
      errors,
    },
    {
      code: 'qux(function(foo, bar, baz) { return foo * this.qux; }.bind(this))',
      errors,
    },
    {
      code: 'foo(function() {}.bind(this, somethingElse))',
      errors,
    },
    {
      code: "qux(function(foo = 1, [bar = 2] = [], {qux: baz = 3} = {foo: 'bar'}) { return foo + bar; });",
      errors,
    },
    {
      code: 'qux(function(baz, baz) { })',
      errors,
    },
    {
      code: 'qux(function( /* no params */ ) { })',
      errors,
    },
    {
      code: 'qux(function( /* a */ foo /* b */ , /* c */ bar /* d */ , /* e */ baz /* f */ ) { return foo; })',
      errors,
    },
    {
      code: 'qux(async function (foo = 1, bar = 2, baz = 3) { return baz; })',
      errors,
    },
    {
      code: 'qux(async function (foo = 1, bar = 2, baz = 3) { return this; }.bind(this))',
      errors,
    },
    {
      code: 'foo(async function /*\n*/ () { return 1; });',
      errors,
    },
    {
      code: 'foo(async function // c\n () { return 1; });',
      errors,
    },
    {
      code: 'foo(async function\n () { return 1; });',
      errors,
    },
    {
      code: 'foo(async function /* c */ () { return 1; });',
      errors,
    },
    {
      code: 'foo((bar || function() {}).bind(this))',
      errors,
    },
    {
      code: 'foo(function() {}.bind(this).bind(obj))',
      errors,
    },

    // Optional chaining
    {
      code: 'foo?.(function() {});',
      errors,
    },
    {
      code: 'foo?.(function() { return this; }.bind(this));',
      errors,
    },
    {
      code: 'foo(function() { return this; }?.bind(this));',
      errors,
    },
    {
      code: 'foo((function() { return this; }?.bind)(this));',
      errors,
    },

    // https://github.com/eslint/eslint/issues/16718
    {
      code: `
            test(
                function ()
                { }
            );
            `,
      errors,
    },
    {
      code: `
            test(
                function (
                    ...args
                ) /* Lorem ipsum
                dolor sit amet. */ {
                    return args;
                }
            );
            `,
      errors,
    },

    // TypeScript upstream invalid cases
    {
      code: 'foo(function(a:string) {});',
      options: [{ allowNamedFunctions: true }] as any,
      errors,
    },
    {
      code: 'foo(function bar(a:string) { a; });',
      errors,
    },
    {
      code: 'foo(function(a:any) { a; });',
      errors,
    },
    {
      code: 'foo(function(arguments:any) { arguments; });',
      errors,
    },
    {
      code: 'foo(function(a:string) { this; });',
      options: [{ allowUnboundThis: false }] as any,
      errors,
    },
    {
      code: 'qux(function(foo:string, bar:number, baz:string) { return foo * 2; })',
      errors,
    },
    {
      code: 'qux(function(foo:number, bar:number, baz:number) { return foo * bar; }.bind(this))',
      errors,
    },
    {
      code: 'qux(function(foo:any, bar:any, baz:any) { return foo * this.qux; }.bind(this))',
      errors,
    },
    {
      code: 'qux(function(baz:string, baz:string) { })',
      errors,
    },
    {
      code: 'qux(function( /* a */ foo:string /* b */ , /* c */ bar:string /* d */ , /* e */ baz:string /* f */ ) { return foo; })',
      errors,
    },
    {
      code: 'qux(async function (foo:number = 1, bar:number = 2, baz:number = 3) { return baz; })',
      errors,
    },
    {
      code: 'qux(async function (foo:number = 1, bar:number = 2, baz:number = 3) { return this; }.bind(this))',
      errors,
    },
    {
      code: "foo(function():string { return 'foo' });",
      errors,
    },
    {
      code: "test('foo', function (this: any) {});",
      errors,
    },
  ],
});
