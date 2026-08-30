import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('class-methods-use-this', {
  valid: [
    // ESLint v10.9.0 JavaScript-parser cases.
    'class A { constructor() {} }',
    'class A { foo() {this} }',
    "class A { foo() {this.bar = 'bar';} }",
    'class A { foo() {bar(this);} }',
    'class A extends B { foo() {super.foo();} }',
    'class A { foo() { if(true) { return this; } } }',
    'class A { static foo() {} }',
    '({ a(){} });',
    'class A { foo() { () => this; } }',
    '({ a: function () {} });',
    {
      code: 'class A { foo() {this} bar() {} }',
      options: { exceptMethods: ['bar'] },
    },
    { code: 'class A { "foo"() { } }', options: { exceptMethods: ['foo'] } },
    { code: 'class A { 42() { } }', options: { exceptMethods: ['42'] } },
    'class A { foo = function() {this} }',
    'class A { foo = () => {this} }',
    'class A { foo = () => {super.toString} }',
    'class A { static foo = function() {} }',
    'class A { static foo = () => {} }',
    { code: 'class A { #bar() {} }', options: { exceptMethods: ['#bar'] } },
    {
      code: 'class A { foo = function () {} }',
      options: { enforceForClassFields: false },
    },
    {
      code: 'class A { foo = () => {} }',
      options: { enforceForClassFields: false },
    },
    'class A { foo() { return class { [this.foo] = 1 }; } }',
    'class A { static {} }',
    // ESLint v10.9.0 TypeScript-parser cases.
    'class A { constructor() {} }',
    'class A { foo() {this} }',
    "class A { foo() {this.bar = 'bar';} }",
    'class A { foo() {bar(this);} }',
    'class A extends B { foo() {super.foo();} }',
    'class A { foo() { if(true) { return this; } } }',
    'class A { static foo() {} }',
    '({ a(){} });',
    'class A { foo() { () => this; } }',
    '({ a: function () {} });',
    {
      code: 'class A { foo() {this} bar() {} }',
      options: { exceptMethods: ['bar'] },
    },
    { code: 'class A { "foo"() { } }', options: { exceptMethods: ['foo'] } },
    { code: 'class A { 42() { } }', options: { exceptMethods: ['42'] } },
    'class A { foo = function() {this} }',
    'class A { foo = () => {this} }',
    'class A { accessor foo = function() {this} }',
    'class A { accessor foo = () => {this} }',
    'class A { accessor foo = 1; }',
    'class A { foo = () => {super.toString} }',
    'class A { static foo = function() {} }',
    'class A { static foo = () => {} }',
    'class A { static accessor foo = function() {} }',
    'class A { static accessor foo = () => {} }',
    { code: 'class A { #bar() {} }', options: { exceptMethods: ['#bar'] } },
    {
      code: 'class A { foo = function () {} }',
      options: { enforceForClassFields: false },
    },
    {
      code: 'class A { foo = () => {} }',
      options: { enforceForClassFields: false },
    },
    {
      code: 'class A { accessor foo = function () {} }',
      options: { enforceForClassFields: false },
    },
    {
      code: 'class A { accessor foo = () => {} }',
      options: { enforceForClassFields: false },
    },
    {
      code: 'class A { override foo = () => {} }',
      options: { enforceForClassFields: false },
    },
    {
      code: 'class Foo implements Bar { property = () => {} }',
      options: { enforceForClassFields: false },
    },
    'class A { foo() { return class { [this.foo] = 1 }; } }',
    'class A { static {} }',
    {
      code: 'class Foo { override method() {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { override ["method"]() {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { private override method() {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { protected override method() {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { override accessor method = () => {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { override get getter(): number {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { private override get getter(): number {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { protected override get getter(): number {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { override set setter(v: number) {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { private override set setter(v: number) {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { protected override set setter(v: number) {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo implements Bar { override method() {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'all',
      },
    },
    {
      code: 'class Foo implements Bar { private override method() {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'public-fields',
      },
    },
    {
      code: 'class Foo implements Bar { protected override method() {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'public-fields',
      },
    },
    {
      code: 'class Foo implements Bar { override get getter(): number {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'all',
      },
    },
    {
      code: 'class Foo implements Bar { private override get getter(): number {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'public-fields',
      },
    },
    {
      code: 'class Foo implements Bar { protected override get getter(): number {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'public-fields',
      },
    },
    {
      code: 'class Foo implements Bar { override set setter(v: number) {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'all',
      },
    },
    {
      code: 'class Foo implements Bar { private override set setter(v: number) {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'public-fields',
      },
    },
    {
      code: 'class Foo implements Bar { protected override set setter(v: number) {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'public-fields',
      },
    },
    {
      code: 'class Foo { override property = () => {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { private override property = () => {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo { protected override property = () => {} }',
      options: { ignoreOverrideMethods: true },
    },
    {
      code: 'class Foo implements Bar { override property = () => {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'all',
      },
    },
    {
      code: 'class Foo implements Bar { private override property = () => {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'public-fields',
      },
    },
    {
      code: 'class Foo implements Bar { protected override property = () => {} }',
      options: {
        ignoreOverrideMethods: true,
        ignoreClassesWithImplements: 'public-fields',
      },
    },
    {
      code: 'class Foo implements Bar { method() {} }',
      options: { ignoreClassesWithImplements: 'all' },
    },
    {
      code: 'class Foo implements Bar { ["method"]() {} }',
      options: { ignoreClassesWithImplements: 'all' },
    },
    {
      code: 'class Foo implements Bar { accessor method = () => {} }',
      options: { ignoreClassesWithImplements: 'all' },
    },
    {
      code: 'class Foo implements Bar { get getter() {} }',
      options: { ignoreClassesWithImplements: 'all' },
    },
    {
      code: 'class Foo implements Bar { set setter() {} }',
      options: { ignoreClassesWithImplements: 'all' },
    },
    {
      code: 'class Foo implements Bar { property = () => {} }',
      options: { ignoreClassesWithImplements: 'all' },
    },
    {
      code: 'const Foo = class implements Bar { method() {} };',
      options: { ignoreClassesWithImplements: 'all' },
    },
    {
      code: 'const Foo = class implements Bar { property = () => {} };',
      options: { ignoreClassesWithImplements: 'all' },
    },
    {
      code: 'const Foo = class implements Bar { method() {} };',
      options: { ignoreClassesWithImplements: 'public-fields' },
    },
    {
      code: 'class Foo implements Bar { ["property"] = () => {} }',
      options: { ignoreClassesWithImplements: 'public-fields' },
    },
  ],
  invalid: [
    // ESLint v10.9.0 JavaScript-parser cases.
    {
      code: 'class A { foo() {} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'class A { foo() {/**this**/} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'class A { foo() {var a = function () {this};} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'class A { foo() {var a = function () {var b = function(){this}};} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'class A { foo() {window.this} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: "class A { foo() {that.this = 'this';} }",
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'class A { foo() { () => undefined; } }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'class A { foo() {} bar() {} }',
      options: { exceptMethods: ['bar'] },
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'class A { foo() {} hasOwnProperty() {} }',
      options: { exceptMethods: ['foo'] },
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'class A { [foo]() {} }',
      options: { exceptMethods: ['foo'] },
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'class A { #foo() { } foo() {} #bar() {} }',
      options: { exceptMethods: ['#foo'] },
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 25,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: "class A { foo(){} 'bar'(){} 123(){} [`baz`](){} [a](){} [f(a)](){} get quux(){} set[a](b){} *quuux(){} }",
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 32,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 44,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 52,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 57,
          endLine: 1,
          endColumn: 63,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 68,
          endLine: 1,
          endColumn: 76,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 81,
          endLine: 1,
          endColumn: 87,
        },
        {
          messageId: 'missingThis',
          line: 1,
          column: 93,
          endLine: 1,
          endColumn: 99,
        },
      ],
    },
    {
      code: 'class A { foo = function() {} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'class A { foo = () => {} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'class A { #foo = function() {} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'class A { #foo = () => {} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class A { #foo() {} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class A { get #foo() {} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'class A { set #foo(x) {} }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'class A { foo () { return class { foo = this }; } }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class A { foo () { return function () { foo = this }; } }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'class A { foo () { return class { static { this; } } } }',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    // ESLint v10.9.0 TypeScript-parser cases.
    {
      code: '\n  class Foo {\n    method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    private method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    protected method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: '\n  class Foo {\n\taccessor method = function () {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 29,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    accessor method = () => {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 26,
          endLine: 3,
          endColumn: 28,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    private accessor method = () => {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 34,
          endLine: 3,
          endColumn: 36,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    protected accessor method = () => {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 36,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
    {
      code: '\n\tclass A {\n\t\tfoo() {\n\t\t\treturn class {\n\t\t\t\taccessor bar = this;\n\t\t\t};\n\t\t}\n\t}\n\t\t\t',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: '\n  class Derived extends Base {\n    override method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: '\n  class Derived extends Base {\n    property = () => {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: '\n  class Derived extends Base {\n    public property = () => {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      code: '\n  class Derived extends Base {\n    override property = () => {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    #method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    get getter(): number {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    private get getter(): number {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    protected get getter(): number {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    get #getter(): number {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    private set setter(b: number) {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    protected set setter(b: number) {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    set #setter(b: number) {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: '\n  function fn() {\n    this.foo = 303;\n\n    class Foo {\n      method() {}\n    }\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 6,
          column: 7,
          endLine: 6,
          endColumn: 13,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    override method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    override get getter(): number {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    override set setter(v: number) {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    override method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    override get getter(): number {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    override set setter(v: number) {}\n  }\n            ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: '\n  class Foo {\n    override property = () => {};\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    override property = () => {};\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    #method() {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    #method() {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    private method() {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    protected method() {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 21,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    get getter(): number {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    get #getter(): number {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    get #getter(): number {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    private get getter(): number {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    protected get getter(): number {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    set setter(v: number) {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    set #setter(v: number) {}\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    set #setter(v: number) {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    private set setter(v: number) {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 23,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    protected set setter(v: number) {}\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    property = () => {};\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    #property = () => {};\n  }\n        ',
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 17,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    #property = () => {};\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 17,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    private property = () => {};\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: '\n  class Foo implements Bar {\n    protected property = () => {};\n  }\n        ',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'const Foo = class implements Bar { method() {} };',
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'const Foo = class implements Bar { private method() {} };',
      options: { ignoreClassesWithImplements: 'public-fields' },
      errors: [
        {
          messageId: 'missingThis',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
  ],
});
