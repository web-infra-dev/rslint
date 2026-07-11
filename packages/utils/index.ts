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
  export class Linter {
    constructor(options: { configType: string }) {
      void options;
    }
    defineParser(name: string, parser: any) {
      void name;
      void parser;
    }
    verifyAndFix(code: string, options: any, options2: any): any {
      void code;
      void options;
      void options2;
    }
  }
}
