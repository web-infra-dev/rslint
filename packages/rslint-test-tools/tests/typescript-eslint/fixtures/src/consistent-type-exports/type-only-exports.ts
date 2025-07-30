// This file only exports types
export type Type1 = string;
export interface Type2 {
  foo: string;
}
export type Type3 = number;

// Type-only namespace
export namespace TypeNS {
  export type X = string;
  export interface Y {
    bar: number;
  }
}
