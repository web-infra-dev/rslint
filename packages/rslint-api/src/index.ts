// Single source of truth: every export resolves into the typescript-go
// submodule we pin via .gitmodules. Avoid `@typescript/native-preview/ast`
// for type/value re-exports — that npm package's `gitHead` and our
// submodule HEAD can drift, leaving the bundle with two parallel `Node`
// definitions and risking SyntaxKind enum-value skew. The npm package is
// still a devDep, but only as the source of the `tsgo` binary used at
// build time; nothing it ships ends up in the published dist of @rslint/api.
//
// api-extractor (rslib's d.ts bundler) does not follow Node subpath imports
// declared in the submodule package's own package.json. Without help it
// leaks unresolvable `#enums/*` literals into the published `.d.ts` — fixed
// by a `paths` entry in tsconfig.build.json (rslib's `dts.alias` is a no-op
// for `bundle: true`, so we have to go through tsconfig).
export { RemoteSourceFile } from '../../../typescript-go/_packages/native-preview/src/api/node/node.ts';
export { SyntaxKind } from '../../../typescript-go/_packages/native-preview/src/ast/index.ts';
export type { Node } from '../../../typescript-go/_packages/native-preview/src/ast/index.ts';
