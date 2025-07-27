// Type exports
export type Type1 = string;
export interface Type2 {
  foo: string;
}

// Value exports
export const value1 = 'value1';
export const value2 = 'value2';

// Mixed namespace (has both types and values)
export namespace ValueNS {
  export const x = 1;
  export type Y = string;
}