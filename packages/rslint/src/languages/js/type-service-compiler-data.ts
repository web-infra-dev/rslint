import type * as TypeScript from 'typescript';

/** Build-time data shared by the worker and the parser-services tests. */
export function typeServiceCompilerData(ts: typeof TypeScript) {
  // The compiler exports this predicate without a public declaration.
  const canHaveJSDoc: (node: { kind: TypeScript.SyntaxKind }) => boolean =
    Reflect.get(ts, 'canHaveJSDoc');
  const jsDocKinds: Record<string, boolean> = {};
  for (const [name, kind] of Object.entries(ts.SyntaxKind)) {
    if (typeof kind === 'number' && canHaveJSDoc({ kind }))
      jsDocKinds[name] = true;
  }
  // The JS factory initializes jsDoc on these two additional nodes. EOF and
  // semicolon class elements acquire the property only when a comment exists.
  jsDocKinds.ModuleBlock = jsDocKinds.MissingDeclaration = true;
  delete jsDocKinds.EndOfFileToken;
  delete jsDocKinds.SemicolonClassElement;
  return {
    SyntaxKind: ts.SyntaxKind,
    TypeFlags: ts.TypeFlags,
    ObjectFlags: ts.ObjectFlags,
    SymbolFlags: { TypeLiteral: ts.SymbolFlags.TypeLiteral },
    NodeFlags: ts.NodeFlags,
    JsxEmit: ts.JsxEmit,
    ModuleDetectionKind: ts.ModuleDetectionKind,
    ScriptTarget: ts.ScriptTarget,
    jsDocKinds,
  } as const;
}

export type TypeServiceCompilerData = ReturnType<
  typeof typeServiceCompilerData
>;
