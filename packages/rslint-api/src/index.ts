// The checked-in compatibility layer preserves the existing binary AST and
// declaration contract. Its fixed source pin, licenses, and per-file hashes are
// recorded in ast-compat/UPSTREAM.json; normal builds do not read an upstream
// compiler checkout or execute a TypeScript Go compiler.
export { RemoteSourceFile } from './ast-compat/api/node/node.ts';
export { SyntaxKind } from './ast-compat/ast/index.ts';
export type { Node } from './ast-compat/ast/index.ts';
