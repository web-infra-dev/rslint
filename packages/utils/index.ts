// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const AST_NODE_TYPES: any = {};
export enum AST_TOKEN_TYPES {
  Identifier,
}
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const TSESTree: any = {};
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type TSESTree = any;

// eslint-disable-next-line @typescript-eslint/no-namespace
export namespace TSESTree {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  export type Program = any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  export type Node = any;
}

// eslint-disable-next-line @typescript-eslint/no-namespace
export namespace TSESLint {
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  export class Linter {
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    constructor(options: { configType: string }) {}
    // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-explicit-any
    defineParser(name: string, parser: any) {}
    // eslint-disable-next-line @typescript-eslint/no-empty-function, @typescript-eslint/no-explicit-any
    verifyAndFix(code: string, options: any, options2: any): any {}
  }
}
