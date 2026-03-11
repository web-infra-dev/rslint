import { RuleTester } from '@typescript-eslint/rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-invalid-void-type', {
  valid: [
    // === allowInGenericTypeArguments: false ===
    {
      code: 'type Generic<T> = [T];',
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      // https://github.com/typescript-eslint/typescript-eslint/issues/1946
      code: `
function foo(): void | never {
  throw new Error('Test');
}
      `,
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: 'type voidNeverUnion = void | never;',
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: 'type neverVoidUnion = never | void;',
      options: [{ allowInGenericTypeArguments: false }],
    },

    // === allowInGenericTypeArguments: true (default) ===
    'function func(): void {}',
    'type NormalType = () => void;',
    // Callable signatures in interfaces
    'interface Callable { (...args: string[]): void; }',
    'interface Constructable { new (...args: string[]): void; }',
    'let normalArrow = (): void => {};',
    'let voidPromise: Promise<void> = new Promise<void>(() => {});',
    'let voidMap: Map<string, void> = new Map<string, void>();',
    `
      function returnsVoidPromiseDirectly(): Promise<void> {
        return Promise.resolve();
      }
    `,
    'async function returnsVoidPromiseAsync(): Promise<void> {}',
    'type UnionType = string | number;',
    'type GenericVoid = Generic<void>;',
    'type Generic<T> = [T];',
    'type voidPromiseUnion = void | Promise<void>;',
    'type promiseNeverUnion = Promise<void> | never;',
    'const arrowGeneric1 = <T = void,>(arg: T) => {};',
    'declare function functionDeclaration1<T = void>(arg: T): void;',
    'type FunctionType = (x: number) => void;',
    'interface Foo { method(): void; }',
    // Callable signatures should be valid even with allowInGenericTypeArguments: false
    {
      code: 'interface Callable { (...args: string[]): void; }',
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: 'interface Constructable { new (...args: string[]): void; }',
      options: [{ allowInGenericTypeArguments: false }],
    },
    'interface GenericCallable { <T>(arg: T): void; }',

    // === void in heritage clause type arguments (extends/implements) ===
    `
interface IObservable<T, U> {}
export interface IObservableSignal<TChange> extends IObservable<void, TChange> {}
    `,
    `
interface Base<T> {}
class Foo implements Base<void> {}
    `,
    // Multiple extends with void
    `
interface A<T> {}
interface B<T> {}
interface C extends A<void>, B<void> {}
    `,
    // Nested generics in extends
    `
interface Wrapper<T> {}
interface Foo extends Wrapper<Promise<void>> {}
    `,
    // Multiple type arguments including void
    `
interface Multi<A, B, C> {}
interface Foo extends Multi<string, void, number> {}
    `,
    // Class implements multiple interfaces with void
    `
interface A<T> {}
interface B<T> {}
class Foo implements A<void>, B<void> {}
    `,
    // Deep nesting: extends with generic of generic of void
    `
interface Deep<T> {}
interface Foo extends Deep<Map<string, void>> {}
    `,

    // === type alias with void as generic type argument ===
    'type Foo = Promise<void>;',
    'type Foo = Map<string, void>;',
    'type Foo = Promise<void> & Map<string, void>;',
    'type Foo = Array<void>;',
    'type Foo = Readonly<Promise<void>>;',

    // === allowInGenericTypeArguments: whitelist ===
    'type Allowed<T> = [T];',
    'type Banned<T> = [T];',
    {
      code: 'type AllowedVoid = Allowed<void>;',
      options: [{ allowInGenericTypeArguments: ['Allowed'] }],
    },
    {
      code: 'type voidPromiseUnion = void | Promise<void>;',
      options: [{ allowInGenericTypeArguments: ['Promise'] }],
    },
    {
      code: 'type promiseVoidUnion = Promise<void> | void;',
      options: [{ allowInGenericTypeArguments: ['Promise'] }],
    },
    {
      code: 'type promiseNeverUnion = Promise<void> | never;',
      options: [{ allowInGenericTypeArguments: ['Promise'] }],
    },
    {
      code: 'type voidPromiseNeverUnion = void | Promise<void> | never;',
      options: [{ allowInGenericTypeArguments: ['Promise'] }],
    },
    // Heritage clause with whitelist - in whitelist
    {
      code: `
interface Base<T> {}
interface Foo extends Base<void> {}
      `,
      options: [{ allowInGenericTypeArguments: ['Base'] }],
    },
    {
      code: `
interface A<T> {}
interface B<T> {}
interface Foo extends A<void>, B<void> {}
      `,
      options: [{ allowInGenericTypeArguments: ['A', 'B'] }],
    },

    // === allowAsThisParameter: true ===
    {
      code: 'function f(this: void) {}',
      options: [{ allowAsThisParameter: true }],
    },
    {
      code: `
class Test {
  public static helper(this: void) {}
  method(this: void) {}
}
      `,
      options: [{ allowAsThisParameter: true }],
    },

    // === void operator expressions (not type) ===
    'let ughThisThing = void 0;',
    'function takeThing(thing: undefined) {}',
    'takeThing(void 0);',

    // === accessor property with non-void type ===
    `
class ClassName {
  accessor propName: number;
}
    `,

    // === Function overloads ===
    `
function f(): void;
function f(x: string): string;
function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}
    `,
    `
class SomeClass {
  f(): void;
  f(x: string): string;
  f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
class SomeClass {
  ['f'](): void;
  ['f'](x: string): string;
  ['f'](x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
class SomeClass {
  [Symbol.iterator](): void;
  [Symbol.iterator](x: string): string;
  [Symbol.iterator](x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
class SomeClass {
  'f'(): void;
  'f'(x: string): string;
  'f'(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
class SomeClass {
  1(): void;
  1(x: string): string;
  1(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
const staticSymbol = Symbol.for('static symbol');

class SomeClass {
  [staticSymbol](): void;
  [staticSymbol](x: string): string;
  [staticSymbol](x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
declare module foo {
  function f(): void;
  function f(x: string): string;
  function f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
{
  function f(): void;
  function f(x: string): string;
  function f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
function f(): Promise<void>;
function f(x: string): Promise<string>;
async function f(x?: string): Promise<void | string> {
  if (x !== undefined) { return x; }
}
    `,
    `
class SomeClass {
  f(): Promise<void>;
  f(x: string): Promise<string>;
  async f(x?: string): Promise<void | string> {
    if (x !== undefined) { return x; }
  }
}
    `,
    `
function f(): void;

const a = 5;

function f(x: string): string;
function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}
    `,
    `
export default function (): void;
export default function (x: string): string;
export default function (x?: string): string | void {
  if (x !== undefined) { return x; }
}
    `,
    `
export function f(): void;
export function f(x: string): string;
export function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}
    `,
    `
export {};

export function f(): void;
export function f(x: string): string;
export function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}
    `,

    // === Dotted generic names with whitelist ===
    {
      code: 'type AllowedVoid = Ex.Mx.Tx<void>;',
      options: [{ allowInGenericTypeArguments: ['Ex.Mx.Tx'] }],
    },
    {
      code: 'type AllowedVoid = Ex . Mx . Tx<void>;',
      options: [{ allowInGenericTypeArguments: ['Ex.Mx.Tx'] }],
    },
    {
      code: 'type AllowedVoid = Ex.Mx.Tx<void>;',
      options: [{ allowInGenericTypeArguments: ['Ex . Mx . Tx'] }],
    },

    // === void in function parameter type within whitelist ===
    {
      code: `
async function foo(bar: () => void | Promise<void>) {
  await bar();
}
      `,
      options: [{ allowInGenericTypeArguments: ['Promise'] }],
    },
  ],
  invalid: [
    // === allowInGenericTypeArguments: false ===
    {
      code: 'type GenericVoid = Generic<void>;',
      errors: [
        {
          column: 28,
          line: 1,
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: 'function takeVoid(thing: void) {}',
      errors: [
        {
          column: 26,
          line: 1,
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: 'let voidPromise: Promise<void> = new Promise<void>(() => {});',
      errors: [
        {
          column: 26,
          line: 1,
          messageId: 'invalidVoidNotReturn',
        },
        {
          column: 46,
          line: 1,
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: 'let voidMap: Map<string, void> = new Map<string, void>();',
      errors: [
        {
          column: 26,
          line: 1,
          messageId: 'invalidVoidNotReturn',
        },
        {
          column: 50,
          line: 1,
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: 'type invalidVoidUnion = void | number;',
      errors: [
        {
          column: 25,
          line: 1,
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: `
interface Base<T> {}
interface Foo extends Base<void> {}`,
      errors: [
        {
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    // Multiple extends with void, allowInGenericTypeArguments: false
    {
      code: `
interface A<T> {}
interface B<T> {}
interface C extends A<void>, B<void> {}`,
      errors: [
        {
          messageId: 'invalidVoidNotReturn',
        },
        {
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    // Class implements with void, allowInGenericTypeArguments: false
    {
      code: `
interface A<T> {}
class Foo implements A<void> {}`,
      errors: [
        {
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    // Type alias with void generic, allowInGenericTypeArguments: false
    {
      code: 'type Foo = Promise<void>;',
      errors: [
        {
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },
    {
      code: 'type Foo = Map<string, void>;',
      errors: [
        {
          messageId: 'invalidVoidNotReturn',
        },
      ],
      options: [{ allowInGenericTypeArguments: false }],
    },

    // === allowInGenericTypeArguments: true (default) ===
    {
      code: 'function takeVoid(thing: void) {}',
      errors: [
        {
          column: 26,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'const arrowGeneric = <T extends void>(arg: T) => {};',
      errors: [
        {
          column: 33,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'const arrowGeneric2 = <T extends void = void>(arg: T) => {};',
      errors: [
        {
          column: 34,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'function functionGeneric<T extends void>(arg: T) {}',
      errors: [
        {
          column: 36,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'function functionGeneric2<T extends void = void>(arg: T) {}',
      errors: [
        {
          column: 37,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'declare function functionDeclaration<T extends void>(arg: T): void;',
      errors: [
        {
          column: 48,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'declare function functionDeclaration2<T extends void = void>(arg: T): void;',
      errors: [
        {
          column: 49,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'functionGeneric<void>(undefined);',
      errors: [
        {
          column: 17,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'let letVoid: void;',
      errors: [
        {
          column: 14,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: `
        type VoidType = void;
      `,
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'type UnionType2 = string | number | void;',
      errors: [
        {
          column: 37,
          line: 1,
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },
    {
      code: 'type IntersectionType = string & number & void;',
      errors: [
        {
          column: 43,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: `
        interface Interface {
          lambda: () => void;
          voidProp: void;
        }
      `,
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: `
        class ClassName {
          private readonly propName: void;
        }
      `,
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'type invalidVoidUnion = void | Map<string, number>;',
      errors: [
        {
          column: 25,
          line: 1,
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },
    {
      code: 'type invalidVoidUnion = void | Map;',
      errors: [
        {
          column: 25,
          line: 1,
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },
    // Void in callable signature parameter position should be invalid
    {
      code: 'interface Foo { (arg: void): string; }',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === allowInGenericTypeArguments: whitelist ===
    {
      code: 'type BannedVoid = Banned<void>;',
      errors: [
        {
          column: 26,
          line: 1,
          messageId: 'invalidVoidForGeneric',
        },
      ],
      options: [{ allowInGenericTypeArguments: ['Allowed'] }],
    },
    {
      code: 'function takeVoid(thing: void) {}',
      errors: [
        {
          column: 26,
          line: 1,
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
      options: [{ allowInGenericTypeArguments: ['Allowed'] }],
    },
    // Heritage clause with whitelist - not in whitelist
    {
      code: `
interface Base<T> {}
interface Foo extends Base<void> {}`,
      errors: [
        {
          messageId: 'invalidVoidForGeneric',
        },
      ],
      options: [{ allowInGenericTypeArguments: ['Allowed'] }],
    },
    // Heritage clause with whitelist - multiple extends, one allowed one not
    {
      code: `
interface Allowed<T> {}
interface Banned<T> {}
interface Foo extends Allowed<void>, Banned<void> {}`,
      errors: [
        {
          messageId: 'invalidVoidForGeneric',
        },
      ],
      options: [{ allowInGenericTypeArguments: ['Allowed'] }],
    },

    // === allowAsThisParameter: true ===
    {
      code: 'type alias = void;',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrThisParamOrGeneric',
        },
      ],
      options: [
        { allowAsThisParameter: true, allowInGenericTypeArguments: true },
      ],
    },
    {
      code: 'type alias = void;',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrThisParam',
        },
      ],
      options: [
        { allowAsThisParameter: true, allowInGenericTypeArguments: false },
      ],
    },
    {
      code: 'type alias = Array<void>;',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrThisParam',
        },
      ],
      options: [
        { allowAsThisParameter: true, allowInGenericTypeArguments: false },
      ],
    },

    // === void in array types ===
    {
      code: 'declare function voidArray(args: void[]): void[];',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === void in type assertions ===
    {
      code: 'let value = undefined as void;',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'let value = <void>undefined;',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === void in rest parameter ===
    {
      code: 'function takesThings(...things: void[]): void {}',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === keyof void ===
    {
      code: 'type KeyofVoid = keyof void;',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === accessor property with void type ===
    {
      code: `
class ClassName {
  accessor propName: void;
}`,
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === void type alias with class usage ===
    {
      code: `
type VoidType = void;
class OtherClassName {
  private propName: VoidType;
}`,
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === nested union with void ===
    {
      code: 'type UnionType3 = string | ((number & any) | (string | void));',
      errors: [
        {
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },

    // === declared function return union ===
    {
      code: 'declare function test(): number | void;',
      errors: [
        {
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },

    // === void in union inside type parameter constraint ===
    {
      code: 'declare function test<T extends number | void>(): T;',
      errors: [
        {
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },

    // === mapped type ===
    {
      code: `
type MappedType<T> = {
  [K in keyof T]: void;
};`,
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === conditional type ===
    {
      code: `
type ConditionalType<T> = {
  [K in keyof T]: T[K] extends string ? void : string;
};`,
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === readonly void array ===
    {
      code: 'type ManyVoid = readonly void[];',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },
    {
      code: 'function foo(arr: readonly void[]) {}',
      errors: [
        {
          messageId: 'invalidVoidNotReturnOrGeneric',
        },
      ],
    },

    // === Class method without overloads - invalid ===
    {
      code: `
class SomeClass {
  f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}`,
      errors: [
        {
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },

    // === Export default function without overloads - invalid ===
    {
      code: 'export default function (x?: string): string | void {}',
      errors: [
        {
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },

    // === Export function without overloads - invalid ===
    {
      code: 'export function f(x?: string): string | void {}',
      errors: [
        {
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },

    // === Overload where intermediate signature has void union - invalid ===
    {
      code: `
function f(): void;
function f(x: string): string | void;
function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}`,
      errors: [
        {
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },

    // === Class method overload where intermediate signature has void union - invalid ===
    {
      code: `
class SomeClass {
  f(): void;
  f(x: string): string | void;
  f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}`,
      errors: [
        {
          messageId: 'invalidVoidUnionConstituent',
        },
      ],
    },

    // === Whitelist invalid with dotted name ===
    {
      code: 'type BannedVoid = Ex.Mx.Tx<void>;',
      errors: [
        {
          messageId: 'invalidVoidForGeneric',
        },
      ],
      options: [{ allowInGenericTypeArguments: ['Tx'] }],
    },
  ],
});
