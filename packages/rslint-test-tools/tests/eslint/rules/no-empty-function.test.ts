import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

const ruleTester = new RuleTester();

interface Seed {
  code: string;
  allow: string[];
}

function lineColumnForOffset(code: string, offset: number) {
  let line = 1;
  let column = 1;
  for (let i = 0; i < offset; i++) {
    if (code[i] === '\n') {
      line++;
      column = 1;
    } else {
      column++;
    }
  }
  return { line, column };
}

function invalidFromSeed(seed: Seed): InvalidTestCase {
  const offset = seed.code.lastIndexOf('{}');
  const { line, column } = lineColumnForOffset(seed.code, offset);
  return {
    code: seed.code,
    errors: [{ messageId: 'unexpected', line, column }],
  };
}

function validFromSeed(seed: Seed): ValidTestCase[] {
  return [
    seed.code.replace('{}', '{ bar(); }'),
    seed.code.replace('{}', '{ /* empty */ }'),
    seed.code.replace('{}', '{\n    // empty\n}'),
    ...seed.allow.map((allow) => ({
      code: `${seed.code} // allow: ${allow}`,
      options: { allow: [allow] },
    })),
  ];
}

const upstreamSeeds: Seed[] = [
  { code: 'function foo() {}', allow: ['functions'] },
  { code: 'var foo = function() {};', allow: ['functions'] },
  { code: 'var obj = {foo: function() {}};', allow: ['functions'] },
  { code: 'var foo = () => {};', allow: ['arrowFunctions'] },
  { code: 'function* foo() {}', allow: ['generatorFunctions'] },
  { code: 'var foo = function*() {};', allow: ['generatorFunctions'] },
  { code: 'var obj = {foo: function*() {}};', allow: ['generatorFunctions'] },
  { code: 'var obj = {foo() {}};', allow: ['methods'] },
  { code: 'class A {foo() {}}', allow: ['methods'] },
  { code: 'class A {static foo() {}}', allow: ['methods'] },
  { code: 'var A = class {foo() {}};', allow: ['methods'] },
  { code: 'var A = class {static foo() {}};', allow: ['methods'] },
  { code: 'var obj = {*foo() {}};', allow: ['generatorMethods'] },
  { code: 'class A {*foo() {}}', allow: ['generatorMethods'] },
  { code: 'class A {static *foo() {}}', allow: ['generatorMethods'] },
  { code: 'var A = class {*foo() {}};', allow: ['generatorMethods'] },
  { code: 'var A = class {static *foo() {}};', allow: ['generatorMethods'] },
  { code: 'var obj = {get foo() {}};', allow: ['getters'] },
  { code: 'class A {get foo() {}}', allow: ['getters'] },
  { code: 'class A {static get foo() {}}', allow: ['getters'] },
  { code: 'var A = class {get foo() {}};', allow: ['getters'] },
  { code: 'var A = class {static get foo() {}};', allow: ['getters'] },
  { code: 'var obj = {set foo(value) {}};', allow: ['setters'] },
  { code: 'class A {set foo(value) {}}', allow: ['setters'] },
  { code: 'class A {static set foo(value) {}}', allow: ['setters'] },
  { code: 'var A = class {set foo(value) {}};', allow: ['setters'] },
  { code: 'var A = class {static set foo(value) {}};', allow: ['setters'] },
  { code: 'class A {constructor() {}}', allow: ['constructors'] },
  { code: 'var A = class {constructor() {}};', allow: ['constructors'] },
  { code: 'const foo = { async method() {} }', allow: ['asyncMethods'] },
  { code: 'async function a(){}', allow: ['asyncFunctions'] },
  { code: 'const foo = async function () {}', allow: ['asyncFunctions'] },
  { code: 'class Foo { async bar() {} }', allow: ['asyncMethods'] },
  { code: 'const foo = async () => {};', allow: ['arrowFunctions'] },
  { code: 'function foo() {}', allow: ['functions'] },
  { code: 'const foo = function(param: string) {};', allow: ['functions'] },
  {
    code: 'const obj = {foo: function(param: string) {}};',
    allow: ['functions'],
  },
  { code: 'const foo = (param: string) => {};', allow: ['arrowFunctions'] },
  { code: 'function* foo(param: string) {}', allow: ['generatorFunctions'] },
  {
    code: 'const foo = function*(param: string) {};',
    allow: ['generatorFunctions'],
  },
  {
    code: 'const obj = {foo: function*(param: string) {}};',
    allow: ['generatorFunctions'],
  },
  { code: 'const obj = {foo(param: string) {}};', allow: ['methods'] },
  { code: 'class A { foo(param: string) {} }', allow: ['methods'] },
  { code: 'class A { private foo() {} }', allow: ['methods'] },
  { code: 'class A { protected foo() {} }', allow: ['methods'] },
  { code: 'class A { static foo(param: string) {} }', allow: ['methods'] },
  { code: 'class A { private static foo() {} }', allow: ['methods'] },
  { code: 'class A { protected static foo() {} }', allow: ['methods'] },
  { code: 'const A = class {foo(param: string) {}};', allow: ['methods'] },
  {
    code: 'const A = class {static foo(param: string) {}};',
    allow: ['methods'],
  },
  { code: 'const A = class {private static foo() {}};', allow: ['methods'] },
  { code: 'const A = class {protected static foo() {}};', allow: ['methods'] },
  {
    code: 'class B { @decorator() foo() {} }',
    allow: ['methods', 'decoratedFunctions'],
  },
  {
    code: 'const B = class { @decorator() foo() {} }',
    allow: ['methods', 'decoratedFunctions'],
  },
  {
    code: 'class B extends C { override foo() {} }',
    allow: ['methods', 'overrideMethods'],
  },
  {
    code: 'class B extends C { @decorator() override foo() {} }',
    allow: ['methods', 'decoratedFunctions', 'overrideMethods'],
  },
  {
    code: 'const obj = {*foo(param: string) {}};',
    allow: ['generatorMethods'],
  },
  { code: 'class A { *foo(param: string) {} }', allow: ['generatorMethods'] },
  {
    code: 'class A {static *foo(param: string) {}}',
    allow: ['generatorMethods'],
  },
  { code: 'class A {private static *foo() {}}', allow: ['generatorMethods'] },
  { code: 'class A {protected static *foo() {}}', allow: ['generatorMethods'] },
  {
    code: 'const A = class {*foo(param: string) {}};',
    allow: ['generatorMethods'],
  },
  {
    code: 'const A = class {static *foo(param: string) {}};',
    allow: ['generatorMethods'],
  },
  { code: 'const obj = {get foo(): string {}};', allow: ['getters'] },
  { code: 'class A {get foo(): string {}}', allow: ['getters'] },
  { code: 'class A {static get foo(): string {}}', allow: ['getters'] },
  { code: 'const A = class {get foo(): string {}};', allow: ['getters'] },
  {
    code: 'const A = class {static get foo(): string {}};',
    allow: ['getters'],
  },
  {
    code: 'class A {@decorator() get foo(): string {}}',
    allow: ['getters', 'decoratedFunctions'],
  },
  {
    code: 'class A {@decorator() static get foo(): string {}}',
    allow: ['getters', 'decoratedFunctions'],
  },
  {
    code: 'const A = class {@decorator() get foo(): string {}};',
    allow: ['getters', 'decoratedFunctions'],
  },
  {
    code: 'const A = class {@decorator() static get foo(): string {}};',
    allow: ['getters', 'decoratedFunctions'],
  },
  {
    code: 'class A extends B {override get foo(): string {}}',
    allow: ['getters', 'overrideMethods'],
  },
  {
    code: 'class A extends B {static override get foo(): string {}}',
    allow: ['getters', 'overrideMethods'],
  },
  {
    code: 'const A = class extends B {override get foo(): string {}};',
    allow: ['getters', 'overrideMethods'],
  },
  {
    code: 'const A = class extends B {static override get foo(): string {}};',
    allow: ['getters', 'overrideMethods'],
  },
  { code: 'const obj = {set foo(value: string) {}};', allow: ['setters'] },
  { code: 'class A {set foo(value: string) {}}', allow: ['setters'] },
  { code: 'class A {static set foo(value: string) {}}', allow: ['setters'] },
  { code: 'const A = class {set foo(value: string) {}};', allow: ['setters'] },
  {
    code: 'const A = class {static set foo(value: string) {}};',
    allow: ['setters'],
  },
  {
    code: 'class A {@decorator() set foo(value: string) {}}',
    allow: ['setters', 'decoratedFunctions'],
  },
  {
    code: 'class A {@decorator() static set foo(value: string) {}}',
    allow: ['setters', 'decoratedFunctions'],
  },
  {
    code: 'const A = class {@decorator() set foo(value: string) {}};',
    allow: ['setters', 'decoratedFunctions'],
  },
  {
    code: 'const A = class {@decorator() static set foo(value: string) {}};',
    allow: ['setters', 'decoratedFunctions'],
  },
  {
    code: 'class A extends B {override set foo(value: string) {}}',
    allow: ['setters', 'overrideMethods'],
  },
  {
    code: 'class A extends B {static override set foo(value: string) {}}',
    allow: ['setters', 'overrideMethods'],
  },
  {
    code: 'const A = class extends B {override set foo(value: string) {}};',
    allow: ['setters', 'overrideMethods'],
  },
  {
    code: 'const A = class extends B {static override set foo(value: string) {}};',
    allow: ['setters', 'overrideMethods'],
  },
  {
    code: 'class A { constructor(param: string) {} }',
    allow: ['constructors'],
  },
  {
    code: 'class B { private constructor() {} }',
    allow: ['constructors', 'privateConstructors'],
  },
  {
    code: 'class B { protected constructor() {} }',
    allow: ['constructors', 'protectedConstructors'],
  },
  {
    code: 'const A = class {constructor(param: string) {}};',
    allow: ['constructors'],
  },
  {
    code: 'const B = class { private constructor() {} }',
    allow: ['constructors', 'privateConstructors'],
  },
  {
    code: 'const B = class { protected constructor() {} }',
    allow: ['constructors', 'protectedConstructors'],
  },
  {
    code: 'const foo = { async method(param: string) {} }',
    allow: ['asyncMethods'],
  },
  { code: 'async function a(param: string){}', allow: ['asyncFunctions'] },
  {
    code: 'const foo = async function(param: string) {}',
    allow: ['asyncFunctions'],
  },
  { code: 'class A { async foo(param: string) {} }', allow: ['asyncMethods'] },
  {
    code: 'class A { @decorator() async foo(param: string) {} }',
    allow: ['asyncMethods', 'decoratedFunctions'],
  },
  {
    code: 'class A extends B { override async foo(param: string) {} }',
    allow: ['asyncMethods', 'overrideMethods'],
  },
  {
    code: 'const foo = async (): Promise<void> => {};',
    allow: ['arrowFunctions'],
  },
  {
    code: 'class A { private constructor() {} }',
    allow: ['privateConstructors', 'constructors'],
  },
  {
    code: 'const A = class { private constructor() {} };',
    allow: ['privateConstructors', 'constructors'],
  },
  {
    code: 'class A { protected constructor() {} }',
    allow: ['protectedConstructors', 'constructors'],
  },
  {
    code: 'const A = class { protected constructor() {} };',
    allow: ['protectedConstructors', 'constructors'],
  },
  {
    code: 'class A { @decorator() foo() {} }',
    allow: ['decoratedFunctions', 'methods'],
  },
  {
    code: 'const A = class { @decorator() foo() {} }',
    allow: ['decoratedFunctions', 'methods'],
  },
  {
    code: 'class B {@decorator() get foo(): string {}}',
    allow: ['decoratedFunctions', 'getters'],
  },
  {
    code: 'class B {@decorator() static get foo(): string {}}',
    allow: ['decoratedFunctions', 'getters'],
  },
  {
    code: 'const B = class {@decorator() get foo(): string {}};',
    allow: ['decoratedFunctions', 'getters'],
  },
  {
    code: 'const B = class {@decorator() static get foo(): string {}};',
    allow: ['decoratedFunctions', 'getters'],
  },
  {
    code: 'class B {@decorator() set foo(value: string) {}}',
    allow: ['decoratedFunctions', 'setters'],
  },
  {
    code: 'class B {@decorator() static set foo(value: string) {}}',
    allow: ['decoratedFunctions', 'setters'],
  },
  {
    code: 'const B = class {@decorator() set foo(value: string) {}};',
    allow: ['decoratedFunctions', 'setters'],
  },
  {
    code: 'const B = class {@decorator() static set foo(value: string) {}};',
    allow: ['decoratedFunctions', 'setters'],
  },
  {
    code: 'class B { @decorator() async foo(param: string) {} }',
    allow: ['decoratedFunctions', 'asyncMethods'],
  },
  {
    code: 'class A extends B { @decorator() override foo() {} }',
    allow: ['decoratedFunctions', 'methods', 'overrideMethods'],
  },
  {
    code: 'class B extends C {override get foo(): string {}}',
    allow: ['overrideMethods', 'getters'],
  },
  {
    code: 'class B extends C {static override get foo(): string {}}',
    allow: ['overrideMethods', 'getters'],
  },
  {
    code: 'const B = class extends C {override get foo(): string {}};',
    allow: ['overrideMethods', 'getters'],
  },
  {
    code: 'const B = class extends C {static override get foo(): string {}};',
    allow: ['overrideMethods', 'getters'],
  },
  {
    code: 'class B extends C {override set foo(value: string) {}}',
    allow: ['overrideMethods', 'setters'],
  },
  {
    code: 'class B extends C {static override set foo(value: string) {}}',
    allow: ['overrideMethods', 'setters'],
  },
  {
    code: 'const B = class extends C {override set foo(value: string) {}};',
    allow: ['overrideMethods', 'setters'],
  },
  {
    code: 'const B = class extends C {static override set foo(value: string) {}};',
    allow: ['overrideMethods', 'setters'],
  },
  {
    code: 'class B extends C { override async foo(param: string) {} }',
    allow: ['overrideMethods', 'asyncMethods'],
  },
  {
    code: 'class C extends D { @decorator() override foo() {} }',
    allow: ['overrideMethods', 'methods', 'decoratedFunctions'],
  },
  {
    code: 'class A extends B { override foo() {} }',
    allow: ['overrideMethods', 'methods'],
  },
];

const valid: ValidTestCase[] = [
  'var foo = () => 0;',
  'class A { constructor(public param: string) {} }',
  'class A { constructor(private param: string) {} }',
  'class A { constructor(protected param: string) {} }',
  'class A { constructor(readonly param: string) {} }',
  ...upstreamSeeds.flatMap(validFromSeed),
];

const invalid: InvalidTestCase[] = [
  {
    code: 'function foo() {}',
    errors: [{ messageId: 'unexpected', line: 1, column: 16 }],
  },
  {
    code: 'var foo = function () {\n}',
    errors: [{ messageId: 'unexpected', line: 1, column: 23 }],
  },
  {
    code: 'var foo = () => {\n\n  }',
    errors: [{ messageId: 'unexpected', line: 1, column: 17 }],
  },
  {
    code: 'var obj = {\n\tfoo() {\n\t}\n}',
    errors: [{ messageId: 'unexpected', line: 2, column: 8 }],
  },
  {
    code: 'class A { foo() { } }',
    errors: [{ messageId: 'unexpected', line: 1, column: 17 }],
  },
  ...upstreamSeeds.map(invalidFromSeed),
];

ruleTester.run('no-empty-function', { valid, invalid });
