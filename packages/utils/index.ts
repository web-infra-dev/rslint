export const AST_NODE_TYPES: any = {};
export enum AST_TOKEN_TYPES {
  Identifier,
}
export const TSESTree: any = {};
export type TSESTree = any;

// eslint-disable-next-line @typescript-eslint/no-namespace
export namespace TSESTree {
  export type Program = any;
  export type Node = any;
}

// eslint-disable-next-line @typescript-eslint/no-namespace
export namespace TSESLint {
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  export class Linter {
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    constructor(options: { configType: string }) {}
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    defineParser(name: string, parser: any) {}
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    verifyAndFix(code: string, options: any, options2: any): any {}
  }
}
