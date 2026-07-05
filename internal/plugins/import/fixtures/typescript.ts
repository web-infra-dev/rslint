export type MyType = string;

export enum MyEnum {
  Foo,
  Bar,
  Baz,
}

export function getFoo(): MyType {
  return 'foo';
}

export namespace MyNamespace {
  export function namespaceFunction() {}
}
