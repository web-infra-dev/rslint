import type * as TypeScript from 'typescript';

/** Build-time enum tables shared by the worker and parser-services tests. */
export function typeServiceCompilerData(ts: Readonly<typeof TypeScript>) {
  // TypeScript exports canHaveJSDoc without a public type declaration.
  const canHaveJSDoc: unknown = Reflect.get(ts, 'canHaveJSDoc');
  const isJSDocHost =
    typeof canHaveJSDoc === 'function'
      ? (node: { readonly kind: TypeScript.SyntaxKind }) =>
          Boolean(Reflect.apply(canHaveJSDoc, ts, [node]))
      : () => false;
  const jsDocKinds: Record<string, boolean> = {};
  for (const [name, kind] of Object.entries(ts.SyntaxKind)) {
    if (typeof kind === 'number' && isJSDocHost({ kind })) {
      jsDocKinds[name] = true;
    }
  }
  // TypeScript's JS AST initializes jsDoc on ModuleBlock and MissingDeclaration,
  // while EndOfFileToken and SemicolonClassElement carry it only when commented.
  jsDocKinds.ModuleBlock = true;
  jsDocKinds.MissingDeclaration = true;
  delete jsDocKinds.EndOfFileToken;
  delete jsDocKinds.SemicolonClassElement;

  return {
    SyntaxKind: ts.SyntaxKind,
    TypeFlags: ts.TypeFlags,
    ObjectFlags: ts.ObjectFlags,
    NodeFlags: ts.NodeFlags,
    JsxEmit: ts.JsxEmit,
    ModuleDetectionKind: ts.ModuleDetectionKind,
    ScriptTarget: ts.ScriptTarget,
    builderEmitFlags: Array.from(
      { length: 8 },
      (_, mask) =>
        (mask & 1 ? ts.EmitFlags.SingleLine : 0) |
        (mask & 2 ? ts.EmitFlags.NoIndentation : 0) |
        (mask & 4 ? ts.EmitFlags.NoAsciiEscaping : 0),
    ),
    jsDocKinds,
  } as const;
}

export type TypeServiceCompilerData = ReturnType<
  typeof typeServiceCompilerData
>;
