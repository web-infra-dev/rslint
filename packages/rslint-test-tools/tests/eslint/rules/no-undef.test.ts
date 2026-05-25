import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-undef', {
  valid: [
    // === Variable / function / class declarations ===
    'var a = 1; a;',
    'let b = 2; b;',
    'const c = 3; c;',
    'function f() { } f();',
    'class MyClass {} new MyClass();',
    'var f = function() {}; f();',
    'var f = () => {}; f();',
    'var a; a = 1;',

    // === Parameters ===
    'function f(x: number) { return x; }',
    'function foo(a: number, b: string) { return a + b; }',

    // === typeof (default: no report) ===
    'typeof a',
    'typeof a === "string"',
    'typeof (a)',
    'typeof ((a))',

    // === Property access / object literal keys ===
    'var obj = { x: 1 }; obj.x;',
    'var obj = { key: 1 };',

    // === Labels ===
    'loop: for (var i = 0; i < 10; i++) { break loop; }',

    // === Built-in globals via lib ===
    'var p = new Promise<void>((resolve) => resolve());',

    // === Type-only positions: type annotations ===
    'type MyType = string; var x: MyType;',
    'interface MyInterface { x: number; } var y: MyInterface;',
    'function f(x: string): string { return x; }',

    // === Type-only positions: generic type arguments ===
    'function identity<T>(val: T): T { return val; }',
    'function constrained<T extends object>(val: T) { return val; }',

    // === Type-only positions: as / satisfies (type part is type-only) ===
    'var x = 1; var y = x as any;',
    'var x = { a: 1 } satisfies Record<string, number>;',

    // === Type-only positions: typeof in type position (TypeQuery) ===
    'var x = 1; type T = typeof x;',

    // === Type-only positions: mapped / conditional / indexed types ===
    'interface I { a: number; } type Mapped = { [K in keyof I]: string };',
    'type IsStr<T> = T extends string ? true : false;',

    // === Type-only positions: interface extends (type-only) ===
    'interface Base { x: number; } interface Derived extends Base { y: number; }',

    // === Type-only positions: class implements (type-only) ===
    'interface I { x: number; } class C implements I { x = 1; }',

    // === Value positions: class extends (value!) ===
    'class Base {} class Child extends Base {}',
    'class A {} function f() { class B extends A {} return B; }',

    // === Shorthand property with declared variable ===
    'var x = 1; var obj = { x };',

    // === Destructuring ===
    'var { a, b } = { a: 1, b: 2 }; a; b;',
    'var [x, y] = [1, 2]; x; y;',

    // === For loop variables ===
    'for (var i = 0; i < 10; i++) { i; }',
    'for (let i = 0; i < 10; i++) { i; }',

    // === Catch clause variable ===
    'try {} catch (e) { e; }',

    // === Class members ===
    'class Foo { bar() {} }; new Foo().bar();',
    'class Foo { bar = 1; baz() { return this.bar; } }',
    'var x = 1; x++;',

    // === Enum ===
    'enum Direction { Up, Down }; Direction.Up;',

    // === /*global*/ comments ===
    '/*global myVar*/ myVar = 1;',
    '/*global a, b*/ a = 1; b = 2;',
    '/*global myVar:writable*/ myVar = 1;',

    // === Namespace ===
    'namespace MyNS { export var x = 1; } MyNS.x;',

    // === Ambient declaration ===
    'declare var declaredAmbient: number; declaredAmbient;',

    // === Type-only positions: union / intersection ===
    'type U = string | number;',
    'type I2 = string & { tag: string };',

    // === Type-only positions: tuple / array / function type ===
    'type Tup = [string, number];',
    'type Arr = string[];',
    'type Fn = (x: string) => number;',

    // === Type-only positions: type predicate ===
    "function isStr(x: any): x is string { return typeof x === 'string'; }",

    // === Type-only positions: nested generics ===
    'type Nested = Map<string, Array<Promise<number>>>;',

    // === Type-only positions: index signature ===
    'interface Indexed { [key: string]: number; }',

    // === Type-only positions: keyof ===
    'interface KI { a: number; } type Keys = keyof KI;',

    // === Type-only positions: infer ===
    'type Unpack<T> = T extends Array<infer U> ? U : T;',

    // === Type-only positions: template literal type ===
    'type EventName = `on${string}`;',

    // === Generic type argument in class extends ===
    'class GenBase<T> { value!: T; } class GenChild extends GenBase<string> {}',

    // === Both extends + implements ===
    'class BothBase {} interface IC { c(): void; } class Both extends BothBase implements IC { c() {} }',

    // === Class expression extends (declared) ===
    'class ExprBase {} var CE = class extends ExprBase {};',

    // === Multiple implements ===
    'interface IA { a(): void; } interface IB { b(): void; } class Multi implements IA, IB { a() {} b() {} }',

    // === import.meta ===
    'var url = import.meta.url;',

    // === as const ===
    'var x = [1, 2, 3] as const;',

    // === Assertion function return type (type-only) ===
    'function assertDefined(x: any): asserts x is string { if (!x) throw new Error(); }',

    // === Conditional type with infer ===
    'type RetType<T> = T extends (...args: any[]) => infer R ? R : never;',

    // === Index access type (type-only) ===
    "interface Obj { key: string; } type Val = Obj['key'];",

    // === Destructuring rename (property key not reported) ===
    'var obj = { a: 1 }; var { a: renamed } = obj; renamed;',

    // === Nested destructuring rename ===
    'var obj = { a: { b: 1 } } as any; var { a: { b: renamed } } = obj; renamed;',

    // === Declared default parameter ===
    'var defaultVal = 1; function f(x = defaultVal) { return x; }',

    // === Declared class property initializer ===
    'var propInit = 42; class C { x = propInit; }',

    // === Rest element in destructuring ===
    'var [first, ...rest] = [1, 2, 3]; first; rest;',
    'var obj = { a: 1, b: 2 }; var { a, ...others } = obj; a; others;',

    // === Built-in globals ===
    'var g = globalThis;',
    'var u = undefined;',
    'var n = NaN;',
    'var inf = Infinity;',

    // === arguments inside function ===
    'function f() { return arguments; }',

    // === this in class method ===
    'class C { x = 1; m() { return this.x; } }',

    // === Chained property access ===
    'var obj = { a: { b: { c: 1 } } }; obj.a.b.c;',

    // === Export declared ===
    'var exportedVar = 1; export { exportedVar };',
    'var localVar = 2; export { localVar as renamedExport };',

    // === delete on declared ===
    'var obj = { prop: 1 } as any; delete obj.prop;',

    // === Computed destructuring key with declared ===
    "var key = 'a'; var { [key]: val } = { a: 1 } as any; val;",

    // === Element access with declared ===
    'var arr = [1, 2, 3]; var idx = 1; var v = arr[idx];',

    // === Getter / setter with declared ===
    'var getterVal = 1; var obj = { get x() { return getterVal; } };',

    // === Private field with declared ===
    'var initVal = 10; class C { #x = initVal; m() { return this.#x; } }',

    // === Parenthesized declared ===
    'var declared = 1; var v = (declared);',

    // === keyof typeof (type-only) ===
    'var someObj = { a: 1 }; type Keys = keyof typeof someObj;',

    // === Mapped type with as clause (type-only) ===
    'interface Src { a: number; b: string; } type OnlyStrings = { [K in keyof Src as Src[K] extends string ? K : never]: Src[K] };',

    // === Multiple interface extends (type-only) ===
    'interface A { a: number; } interface B { b: string; } interface AB extends A, B {}',

    // === Optional parameter type (type-only) ===
    'type MyType = string; function f(x?: MyType) { return x; }',

    // === Class static block with declared ===
    'var staticInit = 42; class C { static { var v = staticInit; } }',
  ],
  invalid: [
    // === Basic undeclared references ===
    {
      code: 'a = 1;',
      errors: [{ messageId: 'undef' }],
    },
    {
      code: 'var a = b;',
      errors: [{ messageId: 'undef' }],
    },
    {
      code: 'undeclaredFunc();',
      errors: [{ messageId: 'undef' }],
    },

    // === typeof with checkTypeof: true ===
    {
      code: "typeof anUndefinedVar === 'string'",
      options: { typeof: true },
      errors: [{ messageId: 'undef' }],
    },

    // === Multiple undeclared variables ===
    {
      code: 'var x = foo + bar;',
      errors: [{ messageId: 'undef' }, { messageId: 'undef' }],
    },

    // === Nested scope ===
    {
      code: 'function foo() { return undeclaredVar123; }',
      errors: [{ messageId: 'undef' }],
    },

    // === Various expression positions ===
    {
      code: 'var x = unknownFunc123();',
      errors: [{ messageId: 'undef' }],
    },
    {
      code: 'if (unknownCondition123) {}',
      errors: [{ messageId: 'undef' }],
    },

    // === /*global*/ mismatch ===
    {
      code: '/*global otherVar*/ unknownVar123 = 1;',
      errors: [{ messageId: 'undef' }],
    },

    // === Shorthand property with undeclared variable ===
    {
      code: 'var obj = { undeclaredShorthand123 };',
      errors: [{ messageId: 'undef' }],
    },

    // === Class extends with undeclared base ===
    {
      code: 'class Child extends undeclaredBase123 {}',
      errors: [{ messageId: 'undef' }],
    },

    // === Template literal ===
    {
      code: 'var s = `${undeclaredTpl123}`;',
      errors: [{ messageId: 'undef' }],
    },

    // === Array literal ===
    {
      code: 'var arr = [undeclaredArr123];',
      errors: [{ messageId: 'undef' }],
    },

    // === Destructuring default value ===
    {
      code: 'var { d = undeclaredDefault123 } = {};',
      errors: [{ messageId: 'undef' }],
    },

    // === Ternary condition ===
    {
      code: 'var x = undeclaredTernary123 ? 1 : 0;',
      errors: [{ messageId: 'undef' }],
    },

    // === Optional chaining on undeclared ===
    {
      code: 'var x = undeclaredOptional123?.prop;',
      errors: [{ messageId: 'undef' }],
    },

    // === Computed property key ===
    {
      code: 'var obj = { [undeclaredComputed123]: 1 };',
      errors: [{ messageId: 'undef' }],
    },

    // === Spread element ===
    {
      code: 'var arr = [...undeclaredSpread123];',
      errors: [{ messageId: 'undef' }],
    },

    // === As expression value side ===
    {
      code: 'var x = undeclaredAsVal123 as any;',
      errors: [{ messageId: 'undef' }],
    },

    // === Nested class extends with undeclared base ===
    {
      code: 'function f() { class Inner extends undeclaredNested123 {} }',
      errors: [{ messageId: 'undef' }],
    },

    // === Undeclared in enum value ===
    {
      code: 'enum E { A = undeclaredEnumVal123 }',
      errors: [{ messageId: 'undef' }],
    },

    // === Function argument ===
    {
      code: 'var fn = (x: any) => x; fn(undeclaredArg123);',
      errors: [{ messageId: 'undef' }],
    },

    // === Nested shorthand property ===
    {
      code: 'var obj = { nested: { undeclaredNestedShort123 } };',
      errors: [{ messageId: 'undef' }],
    },

    // === Class expression extends undeclared ===
    {
      code: 'var CE = class extends undefClassExpr123 {};',
      errors: [{ messageId: 'undef' }],
    },

    // === new undeclared ===
    {
      code: 'new undefNew123();',
      errors: [{ messageId: 'undef' }],
    },

    // === Arrow function body ===
    {
      code: 'var f1 = () => undefArrow123;',
      errors: [{ messageId: 'undef' }],
    },

    // === for-of iterable ===
    {
      code: 'for (var x1 of undefForOf123) {}',
      errors: [{ messageId: 'undef' }],
    },

    // === for-in object ===
    {
      code: 'for (var x2 in undefForIn123) {}',
      errors: [{ messageId: 'undef' }],
    },

    // === throw ===
    {
      code: 'function throwIt() { throw undefThrow123; }',
      errors: [{ messageId: 'undef' }],
    },

    // === Logical AND ===
    {
      code: 'var v1 = true && undefAnd123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Nullish coalescing ===
    {
      code: 'var v2 = null ?? undefNullish123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Unary NOT ===
    {
      code: 'var v3 = !undefNot123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Unary negation ===
    {
      code: 'var v4 = -undefNeg123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Tagged template ===
    {
      code: 'undefTag123`hello`;',
      errors: [{ messageId: 'undef' }],
    },

    // === void ===
    {
      code: 'void undefVoid123;',
      errors: [{ messageId: 'undef' }],
    },

    // === instanceof ===
    {
      code: 'var v5 = ({}) instanceof undefInst123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Deeply nested arrow ===
    {
      code: 'var f2 = () => () => undefDeep123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Generator yield ===
    {
      code: 'function* gen() { yield undefYield123; }',
      errors: [{ messageId: 'undef' }],
    },

    // === Async await ===
    {
      code: 'async function af() { await undefAwait123; }',
      errors: [{ messageId: 'undef' }],
    },

    // === satisfies value side ===
    {
      code: 'var v6 = undefSatisfies123 satisfies any;',
      errors: [{ messageId: 'undef' }],
    },

    // === Multiple shorthand undeclared ===
    {
      code: 'var obj1 = { undefShortA123, undefShortB123 };',
      errors: [{ messageId: 'undef' }, { messageId: 'undef' }],
    },

    // === Mixed declared/undeclared shorthand ===
    {
      code: 'var declared = 1; var obj2 = { declared, undefShortMix123 };',
      errors: [{ messageId: 'undef' }],
    },

    // === Shorthand in class method ===
    {
      code: 'class CS { m() { return { undefMethShort123 }; } }',
      errors: [{ messageId: 'undef' }],
    },

    // === Nested in if-else ===
    {
      code: 'if (true) { var v7 = undefIfElse123; }',
      errors: [{ messageId: 'undef' }],
    },

    // === Switch case ===
    {
      code: 'switch (1) { case 1: var v8 = undefSwitch123; break; }',
      errors: [{ messageId: 'undef' }],
    },

    // === Logical OR assignment ===
    {
      code: 'var v9: any; v9 ||= undefLogical123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Export default undeclared ===
    {
      code: 'export default undefExportDefault123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Computed method name ===
    {
      code: 'class C1 { [undefComputedMethod123]() {} }',
      errors: [{ messageId: 'undef' }],
    },

    // === Non-null assertion ===
    {
      code: 'var v1 = undefNonNull123!;',
      errors: [{ messageId: 'undef' }],
    },

    // === Angle-bracket type assertion ===
    {
      code: 'var v2 = <any>undefAngleBracket123;',
      errors: [{ messageId: 'undef' }],
    },

    // === for-of assignment target (not declaration) ===
    {
      code: 'for (undefForOfTarget123 of [1, 2]) {}',
      errors: [{ messageId: 'undef' }],
    },

    // === Destructuring assignment (not declaration) ===
    {
      code: 'var arr: number[]; [undefAssignTarget123] = [1];',
      errors: [{ messageId: 'undef' }],
    },

    // === Default parameter value ===
    {
      code: 'function f1(x = undefDefaultParam123) {}',
      errors: [{ messageId: 'undef' }],
    },

    // === Class property initializer ===
    {
      code: 'class C2 { x = undefClassProp123; }',
      errors: [{ messageId: 'undef' }],
    },

    // === Class static property ===
    {
      code: 'class C3 { static x = undefStaticProp123; }',
      errors: [{ messageId: 'undef' }],
    },

    // === Object spread ===
    {
      code: 'var obj1 = { ...undefObjSpread123 };',
      errors: [{ messageId: 'undef' }],
    },

    // === Ternary both branches ===
    {
      code: 'var v1 = true ? undefTernaryA123 : undefTernaryB123;',
      errors: [{ messageId: 'undef' }, { messageId: 'undef' }],
    },

    // === Nested destructuring default (only default value, not property key) ===
    {
      code: 'var { a: { b = undefNestedDefault123 } = {} as any } = {} as any;',
      errors: [{ messageId: 'undef' }],
    },

    // === Comma expression ===
    {
      code: 'var v2 = (1, undefComma123);',
      errors: [{ messageId: 'undef' }],
    },

    // === Exponentiation ===
    {
      code: 'var v3 = undefExponent123 ** 2;',
      errors: [{ messageId: 'undef' }],
    },

    // === Assignment operator ===
    {
      code: 'var v4: any; v4 += undefPlusAssign123;',
      errors: [{ messageId: 'undef' }],
    },

    // === Nullish assignment ===
    {
      code: 'var v5: any; v5 ??= undefNullishAssign123;',
      errors: [{ messageId: 'undef' }],
    },

    // === in operator ===
    {
      code: "var v6 = 'key' in undefInOperator123;",
      errors: [{ messageId: 'undef' }],
    },

    // === delete on undeclared ===
    {
      code: 'delete undefDelete123.prop;',
      errors: [{ messageId: 'undef' }],
    },

    // === IIFE with undeclared ===
    {
      code: '(function() { return undefIIFE123; })();',
      errors: [{ messageId: 'undef' }],
    },

    // === Arrow IIFE ===
    {
      code: '(() => undefArrowIIFE123)();',
      errors: [{ messageId: 'undef' }],
    },

    // === Deep nested destructuring default ===
    {
      code: 'var { x: { y: { z = undefDeepDefault123 } = {} as any } = {} as any } = {} as any;',
      errors: [{ messageId: 'undef' }],
    },

    // === Undeclared in template literal conditional ===
    {
      code: "var v7 = `${undefInTemplate123 ? 'a' : 'b'}`;",
      errors: [{ messageId: 'undef' }],
    },

    // === Logical OR ===
    {
      code: 'var v8 = false || undefLogicalOr123;',
      errors: [{ messageId: 'undef' }],
    },

    // === typeof with parentheses + checkTypeof: true ===
    {
      code: 'typeof (anUndefinedVar)',
      options: { typeof: true },
      errors: [{ messageId: 'undef' }],
    },

    // === Computed destructuring key ===
    {
      code: 'var { [undefComputedKey123]: val } = {} as any;',
      errors: [{ messageId: 'undef' }],
    },

    // === Element access expression ===
    {
      code: 'var obj: any = {}; var v = obj[undefElementAccess123];',
      errors: [{ messageId: 'undef' }],
    },

    // === Class static block ===
    {
      code: 'class C1 { static { var v = undefStaticBlock123; } }',
      errors: [{ messageId: 'undef' }],
    },

    // === Getter body ===
    {
      code: 'var obj = { get x() { return undefGetter123; } };',
      errors: [{ messageId: 'undef' }],
    },

    // === Setter body ===
    {
      code: 'var obj = { set x(v: any) { var t = undefSetter123; } };',
      errors: [{ messageId: 'undef' }],
    },

    // === Private field initializer ===
    {
      code: 'class C2 { #x = undefPrivateField123; m() { return this.#x; } }',
      errors: [{ messageId: 'undef' }],
    },

    // === Parenthesized undeclared ===
    {
      code: 'var v = (undefParenthesized123);',
      errors: [{ messageId: 'undef' }],
    },

    // === Double non-null ===
    {
      code: 'var v = undefDoubleNonNull123!!;',
      errors: [{ messageId: 'undef' }],
    },

    // === Nested as/satisfies ===
    {
      code: 'var v = (undefNestedAs123 as any) satisfies any;',
      errors: [{ messageId: 'undef' }],
    },

    // === Optional call on undeclared ===
    {
      code: 'undefOptionalCall123?.();',
      errors: [{ messageId: 'undef' }],
    },
  ],
});
