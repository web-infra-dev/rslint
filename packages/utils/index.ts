export const AST_NODE_TYPES: any = {};
export enum AST_TOKEN_TYPES {
  Identifier,
}
export const TSESTree: any = {};
export type TSESTree = any;
export namespace TSESTree {
  export type Program = any;
  export type Node = any;
}
export namespace TSESLint {
  export class Linter {
    constructor(options: { configType: string }) {}
    defineParser(name: string, parser: any) {}
    verifyAndFix(code: string, options: any, options2: any): any {}
  }
}
