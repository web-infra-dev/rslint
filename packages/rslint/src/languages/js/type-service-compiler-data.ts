import type * as TypeScript from 'typescript';
import type { Operation } from './type-service-protocol.js';

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
  // These checker APIs pass a node predicate to TypeScript's getParseTreeNode.
  const kinds = [...new Set(Object.values(ts.SyntaxKind))].filter(
    (kind): kind is TypeScript.SyntaxKind => typeof kind === 'number',
  );
  const matching = (name: 'isTypeNodeKind' | 'isFunctionLikeKind') => {
    const test: (kind: TypeScript.SyntaxKind) => boolean = Reflect.get(
      ts,
      name,
    );
    return kinds.filter(test);
  };
  const k = ts.SyntaxKind;
  const functions = matching('isFunctionLikeKind');
  const originalKinds: Partial<
    Record<Operation, readonly TypeScript.SyntaxKind[]>
  > = {
    fromTypeNode: matching('isTypeNodeKind'),
    signature: functions,
    isImplementationOfOverload: functions,
    typeOfAssignmentPattern: [
      k.ArrayLiteralExpression,
      k.ObjectLiteralExpression,
    ],
    exportSpecifierLocalTargetSymbol: [k.ExportSpecifier],
    propertySymbolOfDestructuringAssignment: [k.Identifier],
    isOptionalParameter: [k.Parameter],
    parameterPropertySymbols: [k.Parameter],
    constantValue: [
      k.EnumMember,
      k.PropertyAccessExpression,
      k.ElementAccessExpression,
    ],
    validPropertyAccess: [
      k.PropertyAccessExpression,
      k.QualifiedName,
      k.ImportType,
    ],
  };
  return {
    SyntaxKind: ts.SyntaxKind,
    TypeFlags: ts.TypeFlags,
    ObjectFlags: ts.ObjectFlags,
    SymbolFlags: { TypeLiteral: ts.SymbolFlags.TypeLiteral },
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
    originalKinds,
  } as const;
}

export type TypeServiceCompilerData = ReturnType<
  typeof typeServiceCompilerData
>;
