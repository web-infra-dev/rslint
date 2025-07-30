// Type declarations for test modules used in consistent-type-imports tests

declare module 'foo' {
  export default class Foo {
    constructor();
    fn(): { Foo: any };
  }
  export interface A {}
  export interface B {}
  export interface C {}
  export type Type = any;
  export class Bar {}
  export const V: any;
}

declare module 'bar' {
  export default class Bar {
    constructor();
  }
  export interface A {}
  export interface B {}
}

declare module './classA' {
  export class ClassA {}
}

// Add other test modules as needed