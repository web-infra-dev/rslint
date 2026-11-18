import type { JSDocTagInfo, SymbolDisplayPart } from 'typescript';

/** Finite semantic protocol shared by the checker client and injected host. */
export interface Scope {
  generation: number;
  checker: number;
}
export interface NodeRef {
  file: string;
  index: number;
}
export interface FileBinding extends Scope {
  sourceFileName: string;
}
export interface TypeServiceDescriptor extends Scope {
  syntaxOnly?: boolean;
  currentDirectory: string;
  compilerOptions: Record<string, unknown>;
  sources: Array<{
    fileName: string;
    path: string;
    packages?: string[];
    external: boolean;
    defaultLibrary: boolean;
    isDeclarationFile: boolean;
  }>;
}
export interface TypeServiceProjection {
  sourceFileName: string;
  binding: FileBinding;
  tree: Uint8Array;
  origins: number[];
  reverse: Record<string, number>;
  aliases: Array<{
    index: number;
    sequence: number;
    count: number;
    start: number;
    end: number;
  }> | null;
  /** Original frame IDs shared by all aliases of a flattened sequence. */
  sequenceMembers?: Record<number, number[]>;
  /** Optional chain prefixes removed from the ordinary frame. */
  chainAliases?: Array<{
    index: number;
    origin: number;
    expression: number;
    start: number;
    end: number;
  }>;
}
export type SyntaxSource = (
  scope: Scope,
  file: string,
  metadataOnly?: boolean,
) => { fileName: string; tree?: Uint8Array } | undefined;
export interface NodeFact {
  readonly node: NodeRef;
  readonly kind: string;
  readonly flags: number;
  readonly moduleKeyword: number;
  readonly parent: NodeRef | undefined;
  readonly name: NodeRef | undefined;
  readonly text: string | undefined;
}
export interface TypeFact {
  readonly flags: readonly string[];
  readonly symbol: number;
  readonly aliasSymbol: number;
  readonly intrinsicName: string | undefined;
  readonly escapedName: string | undefined;
  readonly objectFlags: readonly string[] | undefined;
  readonly value:
    string | number | { negative: boolean; base10Value: string } | undefined;
  readonly typeOperands: readonly number[] | undefined;
  readonly target: number;
  readonly aliasTypeArguments: readonly number[] | undefined;
  readonly typeParameters: readonly number[] | undefined;
  readonly tuple: { combinedFlags: number; readonly: boolean } | undefined;
}
export interface TypeDetailFact {
  pattern?: NodeRef;
  node?: NodeRef;
  fresh: number;
  regular: number;
  base: number;
  constraint: number;
  outer?: readonly number[];
  local?: readonly number[];
  elements?: readonly number[];
  minLength?: number;
  fixedLength?: number;
  labels?: readonly (NodeRef | undefined)[];
  texts?: readonly string[];
  root?: NodeRef;
  rootCheck: number;
  rootExtends: number;
  rootAlias: number;
  distributive: boolean;
  infer?: readonly number[];
  rootOuter?: readonly number[];
  rootArguments?: readonly number[];
}
export interface SymbolFact {
  readonly name: string;
  readonly escapedName: string;
  readonly flags: number;
  readonly checkFlags: number | undefined;
  readonly hasMembers: boolean;
  readonly hasExports: boolean;
  readonly hasGlobalExports: boolean;
  readonly declarations: readonly NodeRef[] | undefined;
  readonly valueDeclaration: NodeRef | undefined;
}
export interface DiagnosticFact {
  fileName?: string;
  start: number;
  length: number;
  code: number;
  category: number;
  text: string;
  source?: string;
  reportsDeprecated?: true;
  reportsUnnecessary?: true;
  chain: DiagnosticFact[];
  related: DiagnosticFact[];
}
export interface SignatureFact {
  readonly parameters: readonly number[];
  readonly declaration: NodeRef | undefined;
  readonly typeParameters: readonly number[] | undefined;
  readonly thisParameter: number;
  readonly minArgumentCount: number;
}
export interface IndexInfoFact {
  readonly components?: readonly NodeRef[];
  readonly keyType: number;
  readonly type: number;
  readonly isReadonly: boolean;
  readonly declaration: NodeRef | undefined;
}
export interface TypePredicateFact {
  readonly kind: number;
  readonly parameterName: string | undefined;
  readonly parameterIndex: number | undefined;
  readonly type: number;
}
export interface NamedSymbolFact {
  readonly id: number;
  readonly name: string;
}
type NodeArgs = { node: NodeRef };
type IDArgs = { id: number };
type PrintArgs = IDArgs & {
  node: NodeRef | undefined;
  flags: number | undefined;
};
export interface QueryArguments {
  conditionalInstantiations: IDArgs;
  freshLiteral: IDArgs;
  evolvingElement: IDArgs;
  finalArray: IDArgs;
  simplifiedForReading: IDArgs;
  simplifiedForWriting: IDArgs;
  typeParameterConstraint: IDArgs;
  typeDetails: IDArgs;
  resolvedTrueType: IDArgs;
  resolvedFalseType: IDArgs;
  symbolExports: IDArgs;
  symbolGlobalExports: IDArgs;
  configFileParsingDiagnostics: Record<string, never>;
  optionsDiagnostics: Record<string, never>;
  globalDiagnostics: Record<string, never>;
  syntacticDiagnostics: { node?: NodeRef };
  semanticDiagnostics: { node?: NodeRef };
  declarationDiagnostics: { node?: NodeRef };
  nodeCount: Record<string, never>;
  identifierCount: Record<string, never>;
  symbolCount: Record<string, never>;
  typeCount: Record<string, never>;
  instantiationCount: Record<string, never>;
  relationCacheSizes: Record<string, never>;
  rootFileNames: Record<string, never>;
  modeForUsageLocation: NodeArgs & { file: string };
  modeForResolutionAtIndex: NodeArgs & { index: number };
  projectReferences: Record<string, never>;
  resolvedProjectReferences: Record<string, never>;
  stringLiteralType: { name: readonly number[] };
  numberLiteralType: { value: string };
  bigintLiteralType: { negative: boolean; digits: string };
  nullableType: IDArgs & { nullable: boolean; undefinable: boolean };
  privateIdentifierProperty: IDArgs & NodeArgs & { name: readonly number[] };
  moduleMember: IDArgs & { name: readonly number[] };
  validPropertyAccess: NodeArgs & { name: readonly number[] };
  parameterPropertySymbols: NodeArgs & { name: readonly number[] };
  expandoSymbol: NodeArgs & { allowDeclaration: boolean };
  indexSymbolInfos: IDArgs & { siblings?: readonly number[] };
  signatureDeclaration: PrintArgs & { kind: number };
  parameterDeclaration: PrintArgs & { kind: number };
  typeParameterDeclarations: PrintArgs & { kind: number };
  typeParameterDeclaration: PrintArgs & { kind: number };
  entityName: PrintArgs & { kind: number };
  symbolExpression: PrintArgs & { kind: number };
  symbolString: PrintArgs & { kind: number };
  predicateString: {
    kind: number;
    name?: readonly number[];
    position?: number;
    type: number;
    node?: NodeRef;
    flags?: number;
  };
  indexDeclaration: {
    key: number;
    value: number;
    readonly: boolean;
    declaration?: NodeRef;
    node?: NodeRef;
    flags?: number;
  };
  anyType: Record<string, never>;
  unknownType: Record<string, never>;
  stringType: Record<string, never>;
  numberType: Record<string, never>;
  booleanType: Record<string, never>;
  bigIntType: Record<string, never>;
  eSSymbolType: Record<string, never>;
  nonPrimitiveType: Record<string, never>;
  neverType: Record<string, never>;
  nullType: Record<string, never>;
  undefinedType: Record<string, never>;
  trueType: Record<string, never>;
  falseType: Record<string, never>;
  ambientModules: Record<string, never>;
  augmentedPropertiesOfType: IDArgs;
  rootSymbols: IDArgs;
  mergedSymbol: IDArgs;
  isArgumentsSymbol: IDArgs;
  isUndefinedSymbol: IDArgs;
  exportSpecifierLocalTargetSymbol: NodeArgs;
  propertySymbolOfDestructuringAssignment: NodeArgs;
  jsxIntrinsicTagNamesAt: NodeArgs;
  typeOfAssignmentPattern: NodeArgs;
  isOptionalParameter: NodeArgs;
  constantValue: NodeArgs;
  isImplementationOfOverload: NodeArgs;
  typeArgumentsForResolvedSignature: IDArgs;
  nodeRecord: NodeArgs;
  type: NodeArgs;
  fromTypeNode: NodeArgs;
  context: NodeArgs;
  contextualArgument: NodeArgs & { index: number };
  suggestionDiagnostics: NodeArgs;
  symbol: NodeArgs;
  signature: NodeArgs;
  resolvedSignature: NodeArgs;
  resolvedSignatureWithOptions: NodeArgs & {
    withCandidates: boolean;
    argumentCount: number | undefined;
  };
  shorthandValueSymbol: NodeArgs;
  constraint: IDArgs;
  apparent: IDArgs;
  baseLiteral: IDArgs;
  nonNullable: IDArgs;
  typeString: PrintArgs;
  typeNode: PrintArgs;
  signatureString: PrintArgs & { kind: number };
  awaited: IDArgs;
  widened: IDArgs;
  baseTypes: IDArgs;
  defaultType: IDArgs;
  isTypeAssignableTo: { source: number; target: number };
  properties: IDArgs;
  apparentProperties: IDArgs;
  typeArguments: IDArgs;
  isArray: IDArgs;
  isArrayLike: IDArgs;
  isTuple: IDArgs;
  numberIndexType: IDArgs;
  stringIndexType: IDArgs;
  indexInfo: IDArgs & { kind: number };
  indexInfos: IDArgs;
  typeRecord: IDArgs;
  mappedTypeParameter: IDArgs;
  mappedConstraintType: IDArgs;
  mappedTemplateType: IDArgs;
  mappedModifiersType: IDArgs;
  classThisType: IDArgs;
  fullyQualifiedName: IDArgs;
  symbolMembers: IDArgs;
  symbolLinkType: IDArgs;
  aliasedSymbol: IDArgs;
  immediateAliasedSymbol: IDArgs;
  isUnknownSymbol: IDArgs;
  exportsOfModule: IDArgs;
  exportSymbol: IDArgs;
  symbolsInScope: NodeArgs & { meaning: number };
  declaredType: IDArgs;
  typeOfSymbol: IDArgs;
  symbolRecord: IDArgs;
  symbolDocumentation: IDArgs & { withChecker: boolean };
  symbolJSDocTags: IDArgs & { withChecker: boolean };
  returnType: IDArgs;
  typePredicate: IDArgs;
  parameterType: IDArgs & { position: number };
  signatureRecord: IDArgs;
  signatureDocumentation: IDArgs;
  signatureJSDocTags: IDArgs;
  void: Record<string, never>;
  property: IDArgs & { name: readonly number[] };
  propertyType: IDArgs & { name: readonly number[] };
  symbolType: IDArgs & NodeArgs;
  signatures: IDArgs & { kind: number };
  resolveName: {
    name: readonly number[];
    node: NodeRef | undefined;
    kind: number;
    excludeGlobals: boolean;
  };
}
export interface QueryResults {
  conditionalInstantiations: readonly NamedSymbolFact[];
  freshLiteral: number;
  evolvingElement: number;
  finalArray: number;
  simplifiedForReading: number;
  simplifiedForWriting: number;
  typeParameterConstraint: number;
  typeDetails: TypeDetailFact;
  resolvedTrueType: number;
  resolvedFalseType: number;
  symbolExports: readonly NamedSymbolFact[];
  symbolGlobalExports: readonly NamedSymbolFact[];
  configFileParsingDiagnostics: DiagnosticFact[];
  optionsDiagnostics: DiagnosticFact[];
  globalDiagnostics: DiagnosticFact[];
  syntacticDiagnostics: DiagnosticFact[];
  semanticDiagnostics: DiagnosticFact[];
  declarationDiagnostics: DiagnosticFact[];
  nodeCount: number;
  identifierCount: number;
  symbolCount: number;
  typeCount: number;
  instantiationCount: number;
  relationCacheSizes: readonly number[];
  rootFileNames: readonly string[];
  modeForUsageLocation: number;
  modeForResolutionAtIndex: number;
  projectReferences: string;
  resolvedProjectReferences: string;
  stringLiteralType: number;
  numberLiteralType: number;
  bigintLiteralType: number;
  nullableType: number;
  privateIdentifierProperty: number;
  moduleMember: number;
  validPropertyAccess: boolean;
  parameterPropertySymbols: readonly number[];
  expandoSymbol: number;
  indexSymbolInfos: readonly IndexInfoFact[];
  signatureDeclaration: Uint8Array | undefined;
  parameterDeclaration: Uint8Array | undefined;
  typeParameterDeclarations: Uint8Array | undefined;
  typeParameterDeclaration: Uint8Array | undefined;
  entityName: Uint8Array | undefined;
  symbolExpression: Uint8Array | undefined;
  symbolString: string;
  predicateString: string;
  indexDeclaration: Uint8Array | undefined;
  anyType: number;
  unknownType: number;
  stringType: number;
  numberType: number;
  booleanType: number;
  bigIntType: number;
  eSSymbolType: number;
  nonPrimitiveType: number;
  neverType: number;
  nullType: number;
  undefinedType: number;
  trueType: number;
  falseType: number;
  ambientModules: readonly NamedSymbolFact[];
  augmentedPropertiesOfType: readonly NamedSymbolFact[];
  rootSymbols: readonly NamedSymbolFact[];
  mergedSymbol: number;
  isArgumentsSymbol: boolean;
  isUndefinedSymbol: boolean;
  exportSpecifierLocalTargetSymbol: number;
  propertySymbolOfDestructuringAssignment: number;
  jsxIntrinsicTagNamesAt: readonly NamedSymbolFact[];
  typeOfAssignmentPattern: number;
  isOptionalParameter: boolean;
  constantValue: string | number | undefined;
  isImplementationOfOverload: boolean | undefined;
  typeArgumentsForResolvedSignature: readonly number[] | undefined;
  nodeRecord: readonly NodeFact[];
  type: number;
  fromTypeNode: number;
  context: number;
  contextualArgument: number;
  suggestionDiagnostics: DiagnosticFact[];
  symbol: number;
  signature: number;
  resolvedSignature: number;
  resolvedSignatureWithOptions: readonly [number, ...number[]];
  shorthandValueSymbol: number;
  constraint: number;
  apparent: number;
  baseLiteral: number;
  nonNullable: number;
  typeString: string;
  typeNode: Uint8Array | undefined;
  signatureString: string;
  awaited: number;
  widened: number;
  baseTypes: readonly number[];
  defaultType: number;
  isTypeAssignableTo: boolean;
  properties: readonly number[];
  apparentProperties: readonly number[];
  typeArguments: readonly number[];
  isArray: boolean;
  isArrayLike: boolean;
  isTuple: boolean;
  numberIndexType: number;
  stringIndexType: number;
  indexInfo: readonly IndexInfoFact[];
  indexInfos: readonly IndexInfoFact[];
  typeRecord: TypeFact;
  mappedTypeParameter: number;
  mappedConstraintType: number;
  mappedTemplateType: number;
  mappedModifiersType: number;
  classThisType: number;
  fullyQualifiedName: string;
  symbolMembers: readonly NamedSymbolFact[];
  symbolLinkType: number;
  aliasedSymbol: number;
  immediateAliasedSymbol: number;
  isUnknownSymbol: boolean;
  exportsOfModule: readonly NamedSymbolFact[];
  exportSymbol: number;
  symbolsInScope: readonly NamedSymbolFact[];
  declaredType: number;
  typeOfSymbol: number;
  symbolRecord: SymbolFact;
  symbolDocumentation: SymbolDisplayPart[];
  symbolJSDocTags: JSDocTagInfo[];
  returnType: number;
  typePredicate: TypePredicateFact | undefined;
  parameterType: number;
  signatureRecord: SignatureFact;
  signatureDocumentation: SymbolDisplayPart[];
  signatureJSDocTags: JSDocTagInfo[];
  void: number;
  property: number;
  propertyType: number;
  symbolType: number;
  signatures: readonly number[];
  resolveName: number;
}
export type Operation = keyof QueryArguments;
export type TypeServiceQuery<K extends Operation = Operation> = {
  [P in K]: Scope & { op: P; args: QueryArguments[P] };
}[K];
export type QueryTransport = (
  query: TypeServiceQuery,
  key: string,
  cacheable: boolean,
) => Uint8Array;

/** Normal overload resolution can replace semantic results for syntax nodes. */
export function isCacheableTypeQuery(
  operation: Operation,
  signatureOptionsUsed: boolean,
): boolean {
  // The checker populates these slots during later semantic queries.
  if (
    operation === 'resolvedSignatureWithOptions' ||
    operation === 'symbolLinkType' ||
    operation === 'symbolCount' ||
    operation === 'typeCount' ||
    operation === 'instantiationCount' ||
    operation === 'relationCacheSizes' ||
    operation === 'conditionalInstantiations' ||
    operation === 'freshLiteral' ||
    operation === 'typeParameterConstraint' ||
    operation === 'finalArray' ||
    operation === 'simplifiedForReading' ||
    operation === 'simplifiedForWriting' ||
    operation === 'resolvedTrueType' ||
    operation === 'resolvedFalseType' ||
    operation === 'mappedModifiersType'
  )
    return false;
  if (!signatureOptionsUsed) return true;
  switch (operation) {
    case 'type':
    case 'fromTypeNode':
    case 'context':
    case 'contextualArgument':
    case 'symbol':
    case 'signature':
    case 'resolvedSignature':
    case 'symbolType':
      return false;
    default:
      return true;
  }
}
const operations = [
  'type',
  'fromTypeNode',
  'context',
  'symbol',
  'signature',
  'resolvedSignature',
  'constraint',
  'apparent',
  'baseLiteral',
  'nonNullable',
  'typeString',
  'properties',
  'typeArguments',
  'isArray',
  'isArrayLike',
  'isTuple',
  'numberIndexType',
  'typeRecord',
  'aliasedSymbol',
  'declaredType',
  'symbolRecord',
  'returnType',
  'signatureRecord',
  'void',
  'property',
  'symbolType',
  'signatures',
  'resolveName',
  'nodeRecord',
  'shorthandValueSymbol',
  'signatureString',
  'awaited',
  'widened',
  'baseTypes',
  'defaultType',
  'isTypeAssignableTo',
  'apparentProperties',
  'stringIndexType',
  'indexInfo',
  'indexInfos',
  'immediateAliasedSymbol',
  'isUnknownSymbol',
  'exportsOfModule',
  'exportSymbol',
  'symbolsInScope',
  'typeOfSymbol',
  'typePredicate',
  'parameterType',
  'propertyType',
  'resolvedSignatureWithOptions',
  'mappedTypeParameter',
  'mappedConstraintType',
  'mappedTemplateType',
  'symbolDocumentation',
  'symbolJSDocTags',
  'signatureDocumentation',
  'signatureJSDocTags',
  'contextualArgument',
  'suggestionDiagnostics',
  'typeNode',
  'mappedModifiersType',
  'classThisType',
  'fullyQualifiedName',
  'symbolMembers',
  'symbolLinkType',
  'anyType',
  'unknownType',
  'stringType',
  'numberType',
  'booleanType',
  'bigIntType',
  'eSSymbolType',
  'nonPrimitiveType',
  'neverType',
  'nullType',
  'undefinedType',
  'trueType',
  'falseType',
  'ambientModules',
  'augmentedPropertiesOfType',
  'rootSymbols',
  'mergedSymbol',
  'isArgumentsSymbol',
  'isUndefinedSymbol',
  'exportSpecifierLocalTargetSymbol',
  'propertySymbolOfDestructuringAssignment',
  'jsxIntrinsicTagNamesAt',
  'typeOfAssignmentPattern',
  'isOptionalParameter',
  'constantValue',
  'isImplementationOfOverload',
  'typeArgumentsForResolvedSignature',
  'stringLiteralType',
  'numberLiteralType',
  'bigintLiteralType',
  'nullableType',
  'privateIdentifierProperty',
  'moduleMember',
  'validPropertyAccess',
  'parameterPropertySymbols',
  'expandoSymbol',
  'indexSymbolInfos',
  'signatureDeclaration',
  'parameterDeclaration',
  'typeParameterDeclarations',
  'typeParameterDeclaration',
  'entityName',
  'symbolExpression',
  'symbolString',
  'predicateString',
  'indexDeclaration',
  'configFileParsingDiagnostics',
  'optionsDiagnostics',
  'globalDiagnostics',
  'syntacticDiagnostics',
  'semanticDiagnostics',
  'declarationDiagnostics',
  'nodeCount',
  'identifierCount',
  'symbolCount',
  'typeCount',
  'instantiationCount',
  'relationCacheSizes',
  'rootFileNames',
  'modeForUsageLocation',
  'modeForResolutionAtIndex',
  'projectReferences',
  'resolvedProjectReferences',
  'typeDetails',
  'resolvedTrueType',
  'resolvedFalseType',
  'symbolExports',
  'symbolGlobalExports',
  'conditionalInstantiations',
  'freshLiteral',
  'evolvingElement',
  'finalArray',
  'simplifiedForReading',
  'simplifiedForWriting',
  'typeParameterConstraint',
] as const;
const operationIDs = new Map<Operation, number>(
  operations.map((op, index) => [op, index + 1]),
);
export function typeQueryKey(
  q: TypeServiceQuery,
  files?: ReadonlyMap<string, number>,
): string {
  const nodeKey = (node: NodeRef | undefined): string =>
    node
      ? `${files?.get(node.file) ?? `${node.file.length}:${node.file}`}:${node.index}`
      : '-';
  const prefix = `${q.generation}:${q.checker}:${operationIDs.get(q.op)}:`;
  switch (q.op) {
    case 'configFileParsingDiagnostics':
    case 'optionsDiagnostics':
    case 'globalDiagnostics':
    case 'syntacticDiagnostics':
    case 'semanticDiagnostics':
    case 'declarationDiagnostics':
    case 'nodeCount':
    case 'identifierCount':
    case 'symbolCount':
    case 'typeCount':
    case 'instantiationCount':
    case 'relationCacheSizes':
    case 'rootFileNames':
    case 'modeForUsageLocation':
    case 'modeForResolutionAtIndex':
    case 'projectReferences':
    case 'resolvedProjectReferences':
      return prefix + JSON.stringify(q.args);
    case 'stringLiteralType':
    case 'numberLiteralType':
    case 'bigintLiteralType':
    case 'nullableType':
    case 'privateIdentifierProperty':
    case 'moduleMember':
    case 'validPropertyAccess':
    case 'parameterPropertySymbols':
    case 'expandoSymbol':
    case 'indexSymbolInfos':
    case 'signatureDeclaration':
    case 'parameterDeclaration':
    case 'typeParameterDeclarations':
    case 'typeParameterDeclaration':
    case 'entityName':
    case 'symbolExpression':
    case 'symbolString':
    case 'predicateString':
    case 'indexDeclaration':
      return prefix + JSON.stringify(q.args);
    case 'anyType':
    case 'unknownType':
    case 'stringType':
    case 'numberType':
    case 'booleanType':
    case 'bigIntType':
    case 'eSSymbolType':
    case 'nonPrimitiveType':
    case 'neverType':
    case 'nullType':
    case 'undefinedType':
    case 'trueType':
    case 'falseType':
    case 'ambientModules':
    case 'exportSpecifierLocalTargetSymbol':
    case 'propertySymbolOfDestructuringAssignment':
    case 'jsxIntrinsicTagNamesAt':
    case 'typeOfAssignmentPattern':
    case 'isOptionalParameter':
    case 'constantValue':
    case 'isImplementationOfOverload':
      return prefix + JSON.stringify(q.args);
    case 'nodeRecord':
    case 'type':
    case 'fromTypeNode':
    case 'context':
    case 'symbol':
    case 'signature':
    case 'resolvedSignature':
    case 'shorthandValueSymbol':
    case 'suggestionDiagnostics':
      return prefix + nodeKey(q.args.node);
    case 'contextualArgument':
      return `${prefix}${nodeKey(q.args.node)}:${q.args.index}`;
    case 'resolvedSignatureWithOptions':
      return `${prefix}${nodeKey(q.args.node)}:${q.args.withCandidates}:${q.args.argumentCount ?? '-'}`;
    case 'void':
      return prefix;
    case 'property':
    case 'propertyType':
      return `${prefix}${q.args.id}:${q.args.name.length}:${q.args.name.join(',')}`;
    case 'symbolType':
      return `${prefix}${q.args.id}:${nodeKey(q.args.node)}`;
    case 'signatures':
    case 'indexInfo':
      return `${prefix}${q.args.id}:${q.args.kind}`;
    case 'isTypeAssignableTo':
      return `${prefix}${q.args.source}:${q.args.target}`;
    case 'symbolsInScope':
      return `${prefix}${nodeKey(q.args.node)}:${q.args.meaning}`;
    case 'parameterType':
      return `${prefix}${q.args.id}:${q.args.position}`;
    case 'symbolDocumentation':
    case 'symbolJSDocTags':
      return `${prefix}${q.args.id}:${q.args.withChecker}`;
    case 'typeString':
    case 'typeNode':
      return `${prefix}${q.args.id}:${nodeKey(q.args.node)}:${q.args.flags ?? '-'}`;
    case 'signatureString':
      return `${prefix}${q.args.id}:${nodeKey(q.args.node)}:${q.args.flags ?? '-'}:${q.args.kind}`;
    case 'resolveName':
      return `${prefix}${q.args.name.length}:${q.args.name.join(',')}:${nodeKey(q.args.node)}:${q.args.kind}:${q.args.excludeGlobals}`;
    default:
      return prefix + q.args.id;
  }
}

export type ScalarOperation = Exclude<
  Operation,
  'typeRecord' | 'symbolRecord' | 'signatureRecord' | 'nodeRecord'
>;
export const SCALAR_KINDS: Record<ScalarOperation, number> = {
  conditionalInstantiations: 11,
  freshLiteral: 1,
  evolvingElement: 1,
  finalArray: 1,
  simplifiedForReading: 1,
  simplifiedForWriting: 1,
  typeParameterConstraint: 1,
  typeDetails: 19,
  resolvedTrueType: 1,
  resolvedFalseType: 1,
  symbolExports: 11,
  symbolGlobalExports: 11,
  configFileParsingDiagnostics: 15,
  optionsDiagnostics: 15,
  globalDiagnostics: 15,
  syntacticDiagnostics: 15,
  semanticDiagnostics: 15,
  declarationDiagnostics: 15,
  nodeCount: 1,
  identifierCount: 1,
  symbolCount: 1,
  typeCount: 1,
  instantiationCount: 1,
  relationCacheSizes: 4,
  rootFileNames: 18,
  modeForUsageLocation: 1,
  modeForResolutionAtIndex: 1,
  projectReferences: 3,
  resolvedProjectReferences: 3,
  stringLiteralType: 1,
  numberLiteralType: 1,
  bigintLiteralType: 1,
  nullableType: 1,
  privateIdentifierProperty: 1,
  moduleMember: 1,
  validPropertyAccess: 2,
  parameterPropertySymbols: 4,
  expandoSymbol: 1,
  indexSymbolInfos: 9,
  signatureDeclaration: 14,
  parameterDeclaration: 14,
  typeParameterDeclarations: 14,
  typeParameterDeclaration: 14,
  entityName: 14,
  symbolExpression: 14,
  symbolString: 3,
  predicateString: 3,
  indexDeclaration: 14,
  anyType: 1,
  unknownType: 1,
  stringType: 1,
  numberType: 1,
  booleanType: 1,
  bigIntType: 1,
  eSSymbolType: 1,
  nonPrimitiveType: 1,
  neverType: 1,
  nullType: 1,
  undefinedType: 1,
  trueType: 1,
  falseType: 1,
  ambientModules: 11,
  augmentedPropertiesOfType: 11,
  rootSymbols: 11,
  mergedSymbol: 1,
  isArgumentsSymbol: 2,
  isUndefinedSymbol: 2,
  exportSpecifierLocalTargetSymbol: 1,
  propertySymbolOfDestructuringAssignment: 1,
  jsxIntrinsicTagNamesAt: 11,
  typeOfAssignmentPattern: 1,
  isOptionalParameter: 2,
  constantValue: 16,
  isImplementationOfOverload: 2,
  typeArgumentsForResolvedSignature: 17,
  type: 1,
  fromTypeNode: 1,
  context: 1,
  contextualArgument: 1,
  suggestionDiagnostics: 15,
  symbol: 1,
  signature: 1,
  resolvedSignature: 1,
  resolvedSignatureWithOptions: 4,
  shorthandValueSymbol: 1,
  constraint: 1,
  apparent: 1,
  baseLiteral: 1,
  nonNullable: 1,
  typeString: 3,
  typeNode: 14,
  signatureString: 3,
  awaited: 1,
  widened: 1,
  baseTypes: 4,
  defaultType: 1,
  isTypeAssignableTo: 2,
  properties: 4,
  apparentProperties: 4,
  typeArguments: 4,
  mappedTypeParameter: 1,
  mappedConstraintType: 1,
  mappedTemplateType: 1,
  mappedModifiersType: 1,
  classThisType: 1,
  fullyQualifiedName: 3,
  symbolMembers: 11,
  symbolLinkType: 1,
  isArray: 2,
  isArrayLike: 2,
  isTuple: 2,
  numberIndexType: 1,
  stringIndexType: 1,
  indexInfo: 9,
  indexInfos: 9,
  aliasedSymbol: 1,
  immediateAliasedSymbol: 1,
  isUnknownSymbol: 2,
  exportsOfModule: 11,
  exportSymbol: 1,
  symbolsInScope: 11,
  declaredType: 1,
  typeOfSymbol: 1,
  symbolDocumentation: 12,
  symbolJSDocTags: 13,
  returnType: 1,
  typePredicate: 10,
  parameterType: 1,
  signatureDocumentation: 12,
  signatureJSDocTags: 13,
  void: 1,
  property: 1,
  propertyType: 1,
  symbolType: 1,
  signatures: 4,
  resolveName: 1,
};

export type BatchTransport = (
  queries: readonly TypeServiceQuery[],
  keys: readonly string[],
) => readonly Uint8Array[];

export type ServiceRequest<K extends Operation = Operation> = {
  [P in K]: { op: P; args: QueryArguments[P] };
}[K];
export type ScalarRequest = ServiceRequest<ScalarOperation>;

export function nameToUnits(name: string): number[] {
  const units: number[] = [];
  for (let i = 0; i < name.length; i++) units.push(name.charCodeAt(i));
  return units;
}
