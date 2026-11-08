// Keep the native syntax reader aligned with the Go compiler's pinned source.
// This private entry is bundled into core; it is not a public package surface.
export {
  RemoteSourceFile,
  RemoteNode,
} from '../../../typescript-go/packages/typescript/src/api/node/node.ts';
export {
  SyntaxKind,
  NodeFlags,
} from '../../../typescript-go/packages/typescript/src/ast/index.ts';
export { Wtf8Decoder } from '../../../typescript-go/packages/typescript/src/api/node/wtf8.ts';
