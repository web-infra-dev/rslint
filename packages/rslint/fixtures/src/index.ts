let a: any = 10;
a.b = 10;

let b: any = 200;
b.c = 200;

// Test adjacent-overload-signatures
interface TestInterface {
  foo(x: string): void;
  bar(): void;
  foo(x: number): void; // This should trigger adjacent-overload-signatures
}

// Test array-type
let arr1: Array<string> = []; // This should trigger array-type (prefer string[])
let arr2: ReadonlyArray<number> = []; // This should trigger array-type (prefer readonly number[])

// Test class-literal-property-style
class TestClass {
  readonly prop1 = 'literal'; // This should trigger class-literal-property-style (prefer getter)
}
