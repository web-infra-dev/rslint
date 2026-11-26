import type { JSDocTagInfo, SymbolDisplayPart } from 'typescript';

/** Finite semantic protocol shared by the checker client and injected host. */
export interface Scope {
  readonly generation: number;
}

export interface SyntaxTokenRef {
  readonly kind: number;
  readonly pos: number;
  readonly end: number;
}

export interface NodeRef {
  readonly file: string;
  readonly index: number;
  readonly token?: SyntaxTokenRef;
}

export interface GeneratedSyntax {
  readonly tree: Uint8Array;
  readonly origins?: readonly (NodeRef | undefined)[];
  readonly emitFlags?: readonly number[];
}

export interface FileBinding extends Scope {
  readonly sourceFileName: string;
}

export interface SourceDescriptor {
  readonly fileName: string;
  readonly path: string;
  readonly packages?: readonly string[];
  readonly impliedNodeFormat?: number;
  readonly checkJsDirective?: {
    readonly pos: number;
    readonly end: number;
    readonly enabled: boolean;
  };
  readonly external: boolean;
  readonly defaultLibrary: boolean;
  readonly isDeclarationFile: boolean;
}

export interface TypeServiceDescriptor extends Scope {
  readonly syntaxOnly?: boolean;
  readonly currentDirectory: string;
  readonly compilerOptions: Readonly<Record<string, unknown>>;
  readonly sources: readonly SourceDescriptor[];
  readonly diagnosticSources?: readonly SourceDescriptor[];
}

export interface SequenceAlias {
  readonly index: number;
  readonly sequence: number;
  readonly count: number;
  readonly start: number;
  readonly end: number;
}

export interface ChainAlias {
  readonly index: number;
  readonly origin: number;
  readonly expression: number;
  readonly start: number;
  readonly end: number;
}

export interface TypeServiceProjection {
  readonly sourceFileName: string;
  readonly binding: FileBinding;
  readonly tree: Uint8Array;
  readonly origins: readonly number[];
  readonly firstTokens?: Readonly<Record<number, number>>;
  readonly reverse: Readonly<Record<string, number>>;
  readonly aliases: readonly SequenceAlias[] | null;
  /** Original frame IDs shared by all aliases of a flattened sequence. */
  readonly sequenceMembers?: Readonly<Record<number, readonly number[]>>;
  /** Optional chain prefixes removed from the ordinary frame. */
  readonly chainAliases?: readonly ChainAlias[];
}

export type SyntaxSource = (
  scope: Scope,
  file: string,
  metadataOnly?: boolean,
) =>
  | {
      readonly fileName: string;
      readonly path?: string;
      readonly tree?: Uint8Array;
    }
  | undefined;

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
    | string
    | number
    | { readonly negative: boolean; readonly base10Value: string }
    | undefined;
  readonly typeOperands: readonly number[] | undefined;
  readonly target: number;
  readonly aliasTypeArguments: readonly number[] | undefined;
  readonly typeParameters: readonly number[] | undefined;
  readonly tuple:
    { readonly combinedFlags: number; readonly readonly: boolean } | undefined;
}

export interface TypeDetailFact {
  readonly outer?: readonly number[];
  readonly local?: readonly number[];
  readonly elements?: readonly number[];
  readonly minLength?: number;
  readonly fixedLength?: number;
  readonly labels?: readonly (NodeRef | undefined)[];
  readonly texts?: readonly string[];
  readonly root?: NodeRef;
  readonly rootCheck: number;
  readonly rootExtends: number;
  readonly rootAlias: number;
  readonly distributive: boolean;
  readonly infer?: readonly number[];
  readonly rootOuter?: readonly number[];
  readonly rootArguments?: readonly number[];
}

export interface SymbolFact {
  readonly name: string;
  readonly escapedName: string;
  readonly flags: number;
  readonly checkFlags: number | undefined;
  readonly declarations: readonly NodeRef[] | undefined;
  readonly valueDeclaration: NodeRef | undefined;
}

export interface DiagnosticFact {
  readonly fileName?: string;
  readonly filePath?: string;
  readonly start: number;
  readonly length: number;
  readonly code: number;
  readonly category: number;
  readonly text: string;
  readonly source?: string;
  readonly reportsDeprecated?: true;
  readonly reportsUnnecessary?: true;
  readonly chain: readonly DiagnosticFact[];
  readonly related: readonly DiagnosticFact[];
}

export interface SignatureFact {
  readonly parameters: readonly number[];
  readonly declaration: NodeRef | undefined;
  readonly typeParameters: readonly number[] | undefined;
  readonly thisParameter: number;
  readonly minArgumentCount: number;
}

export interface IndexInfoFact {
  readonly id: number;
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

type EmptyArgs = Readonly<Record<string, never>>;
type NodeArgs = { readonly node: NodeRef };
type IDArgs = { readonly id: number };
type PrintArgs = IDArgs & {
  readonly node: NodeRef | undefined;
  readonly flags: number | undefined;
};

export interface QueryArguments {
  // Program queries
  rootFileNames: EmptyArgs;
  configFileParsingDiagnostics: { readonly node?: NodeRef };
  optionsDiagnostics: { readonly node?: NodeRef };
  globalDiagnostics: { readonly node?: NodeRef };
  syntacticDiagnostics: { readonly node?: NodeRef };
  semanticDiagnostics: { readonly node?: NodeRef };
  declarationDiagnostics: { readonly node?: NodeRef };
  // Syntax summaries
  nodeRecord: NodeArgs;
  // Intrinsic & literal types & ambient modules
  errorType: EmptyArgs;
  anyType: EmptyArgs;
  unknownType: EmptyArgs;
  stringType: EmptyArgs;
  numberType: EmptyArgs;
  booleanType: EmptyArgs;
  bigIntType: EmptyArgs;
  eSSymbolType: EmptyArgs;
  nonPrimitiveType: EmptyArgs;
  neverType: EmptyArgs;
  nullType: EmptyArgs;
  undefinedType: EmptyArgs;
  trueType: EmptyArgs;
  falseType: EmptyArgs;
  void: EmptyArgs;
  stringLiteralType: { readonly value: readonly number[] };
  numberLiteralType: {
    readonly value: number;
    readonly nonFinite?: 'NaN' | 'Infinity' | '-Infinity';
  };
  bigIntLiteralType: {
    readonly negative: boolean;
    readonly base10Value: string;
  };
  nullableType: IDArgs & { readonly flags: number };
  ambientModules: EmptyArgs;
  // Node -> semantic queries
  type: NodeArgs;
  fromTypeNode: NodeArgs;
  context: NodeArgs;
  contextualArgument: NodeArgs & { readonly index: number };
  suggestionDiagnostics: NodeArgs;
  symbol: NodeArgs;
  declarationSymbol: NodeArgs;
  signature: NodeArgs;
  resolvedSignature: NodeArgs & {
    readonly withCandidates?: boolean;
    readonly argumentCount?: number;
  };
  shorthandValueSymbol: NodeArgs;
  exportSpecifierLocalTargetSymbol: NodeArgs;
  propertySymbolOfDestructuringAssignment: NodeArgs;
  jsxIntrinsicTagNamesAt: NodeArgs;
  constantValue: NodeArgs;
  isOptionalParameter: NodeArgs;
  isImplementationOfOverload: NodeArgs;
  validPropertyAccess: NodeArgs & { readonly name: readonly number[] };
  parameterPropertySymbols: NodeArgs & { readonly name: readonly number[] };
  symbolsInScope: NodeArgs & { readonly meaning: number };
  resolveName: {
    readonly name: readonly number[];
    readonly node: NodeRef | undefined;
    readonly kind: number;
    readonly excludeGlobals: boolean;
  };
  // Type queries
  typeRecord: IDArgs;
  typeDetails: IDArgs;
  declaredIndexInfos: IDArgs;
  mappedFields: IDArgs;
  classThisType: IDArgs;
  constraint: IDArgs;
  apparent: IDArgs;
  baseLiteral: IDArgs;
  nonNullable: IDArgs;
  awaited: IDArgs;
  widened: IDArgs;
  baseTypes: IDArgs;
  defaultType: IDArgs;
  isTypeAssignableTo: { readonly source: number; readonly target: number };
  properties: IDArgs;
  augmentedPropertiesOfType: IDArgs;
  property: IDArgs & { readonly name: readonly number[] };
  propertyType: IDArgs & { readonly name: readonly number[] };
  typeArguments: IDArgs;
  isArray: IDArgs;
  isArrayLike: IDArgs;
  isTuple: IDArgs;
  numberIndexType: IDArgs;
  stringIndexType: IDArgs;
  indexInfo: IDArgs & { readonly kind: number };
  indexInfos: IDArgs;
  signatures: IDArgs & { readonly kind: number };
  typeString: PrintArgs;
  typeNode: PrintArgs;
  typePredicateString: {
    readonly kind: number;
    readonly parameterName?: readonly number[];
    readonly parameterIndex?: number;
    readonly type: number;
    readonly node: NodeRef | undefined;
    readonly flags: number | undefined;
  };
  // Symbol queries
  symbolRecord: IDArgs;
  symbolMembers: IDArgs;
  symbolExports: IDArgs;
  symbolDocumentation: IDArgs & { readonly withChecker: boolean };
  symbolJSDocTags: IDArgs & { readonly withChecker: boolean };
  fullyQualifiedName: IDArgs;
  aliasedSymbol: IDArgs;
  immediateAliasedSymbol: IDArgs;
  isUnknownSymbol: IDArgs;
  exportsOfModule: IDArgs;
  exportSymbol: IDArgs;
  declaredType: IDArgs;
  typeOfSymbol: IDArgs;
  symbolType: IDArgs & NodeArgs;
  rootSymbols: IDArgs;
  mergedSymbol: IDArgs;
  isArgumentsSymbol: IDArgs;
  isUndefinedSymbol: IDArgs;
  moduleMember: IDArgs & { readonly name: readonly number[] };
  symbolString: PrintArgs & { readonly kind: number };
  // Signature queries
  signatureRecord: IDArgs;
  returnType: IDArgs;
  typePredicate: IDArgs;
  parameterType: IDArgs & { readonly position: number };
  signatureDocumentation: IDArgs;
  signatureJSDocTags: IDArgs;
  signatureString: PrintArgs & { readonly kind: number };
}

export interface QueryResults {
  rootFileNames: readonly string[];
  configFileParsingDiagnostics: readonly DiagnosticFact[];
  optionsDiagnostics: readonly DiagnosticFact[];
  globalDiagnostics: readonly DiagnosticFact[];
  syntacticDiagnostics: readonly DiagnosticFact[];
  semanticDiagnostics: readonly DiagnosticFact[];
  declarationDiagnostics: readonly DiagnosticFact[];
  nodeRecord: readonly NodeFact[];
  errorType: number;
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
  void: number;
  stringLiteralType: number;
  numberLiteralType: number;
  bigIntLiteralType: number;
  nullableType: number;
  ambientModules: readonly NamedSymbolFact[];
  type: number;
  fromTypeNode: number;
  context: number;
  contextualArgument: number;
  suggestionDiagnostics: readonly DiagnosticFact[];
  symbol: number;
  declarationSymbol: number;
  signature: number;
  resolvedSignature: readonly number[];
  shorthandValueSymbol: number;
  exportSpecifierLocalTargetSymbol: number;
  propertySymbolOfDestructuringAssignment: number;
  jsxIntrinsicTagNamesAt: readonly NamedSymbolFact[];
  constantValue: string | number | undefined;
  isOptionalParameter: boolean;
  isImplementationOfOverload: boolean;
  validPropertyAccess: boolean;
  parameterPropertySymbols: readonly number[];
  symbolsInScope: readonly NamedSymbolFact[];
  resolveName: number;
  typeRecord: TypeFact;
  typeDetails: TypeDetailFact;
  declaredIndexInfos: readonly IndexInfoFact[] | undefined;
  mappedFields: readonly number[];
  classThisType: number;
  constraint: number;
  apparent: number;
  baseLiteral: number;
  nonNullable: number;
  awaited: number;
  widened: number;
  baseTypes: readonly number[];
  defaultType: number;
  isTypeAssignableTo: boolean;
  properties: readonly NamedSymbolFact[];
  augmentedPropertiesOfType: readonly NamedSymbolFact[];
  property: number;
  propertyType: number;
  typeArguments: readonly number[];
  isArray: boolean;
  isArrayLike: boolean;
  isTuple: boolean;
  numberIndexType: number;
  stringIndexType: number;
  indexInfo: readonly IndexInfoFact[];
  indexInfos: readonly IndexInfoFact[];
  signatures: readonly number[];
  typeString: string;
  typeNode: GeneratedSyntax | undefined;
  typePredicateString: string;
  symbolRecord: SymbolFact;
  symbolMembers: readonly NamedSymbolFact[] | undefined;
  symbolExports: readonly NamedSymbolFact[] | undefined;
  symbolDocumentation: readonly SymbolDisplayPart[];
  symbolJSDocTags: readonly JSDocTagInfo[];
  fullyQualifiedName: string;
  aliasedSymbol: number;
  immediateAliasedSymbol: number;
  isUnknownSymbol: boolean;
  exportsOfModule: readonly NamedSymbolFact[];
  exportSymbol: number;
  declaredType: number;
  typeOfSymbol: number;
  symbolType: number;
  rootSymbols: readonly NamedSymbolFact[];
  mergedSymbol: number;
  isArgumentsSymbol: boolean;
  isUndefinedSymbol: boolean;
  moduleMember: number;
  symbolString: string;
  signatureRecord: SignatureFact;
  returnType: number;
  typePredicate: TypePredicateFact | undefined;
  parameterType: number;
  signatureDocumentation: readonly SymbolDisplayPart[];
  signatureJSDocTags: readonly JSDocTagInfo[];
  signatureString: string;
}

export type Operation = keyof QueryArguments;
export type TypeServiceQuery<K extends Operation = Operation> = {
  [P in K]: Scope & {
    readonly op: P;
    readonly args: QueryArguments[P];
    readonly skipRecords?: boolean;
  };
}[K];
export type QueryTransport = (
  query: TypeServiceQuery,
  key: string,
  cacheable: boolean,
) => Uint8Array;

const NON_CACHEABLE_OPS = new Set<Operation>([
  'configFileParsingDiagnostics',
  'optionsDiagnostics',
  'globalDiagnostics',
  'semanticDiagnostics',
  'declarationDiagnostics',
  'suggestionDiagnostics',
  'declaredIndexInfos',
  'mappedFields',
  'symbolMembers',
  'symbolExports',
]);

/**
 * Only genuinely immutable compiler facts may be cached across queries.
 * Queries backed by lazily mutated checker tables (e.g. symbol tables,
 * unresolved mapped/declared member slots, diagnostics) stay uncached.
 */
export function isCacheableTypeQuery(query: TypeServiceQuery): boolean {
  if (query.op === 'nodeRecord' || query.op === 'syntacticDiagnostics') {
    return true;
  }
  if (NON_CACHEABLE_OPS.has(query.op)) {
    return false;
  }
  return !('node' in query.args && query.args.node !== undefined);
}

function formatNodeKey(node: NodeRef | undefined): string {
  if (!node) return '-';
  const base = `${node.file.length}:${node.file}:${node.index}`;
  return node.token
    ? `${base}:${node.token.kind}:${node.token.pos}:${node.token.end}`
    : base;
}

/**
 * Operations for which Go's `directRecords` bundles supplemental shallow
 * records in the wire reply. Only these operations change wire representation
 * when `skipRecords: true` is set, so only their transport cache keys need to
 * distinguish compact from full replies.
 */
export const DIRECT_RECORD_OPS: ReadonlySet<Operation> = new Set<Operation>([
  // typeRecord
  'type',
  'fromTypeNode',
  'context',
  'constraint',
  'apparent',
  'baseLiteral',
  'nonNullable',
  'errorType',
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
  'stringLiteralType',
  'numberLiteralType',
  'bigIntLiteralType',
  'nullableType',
  'void',
  'declaredType',
  'symbolType',
  'returnType',
  'numberIndexType',
  'typeArguments',
  'stringIndexType',
  'awaited',
  'widened',
  'baseTypes',
  'defaultType',
  'typeOfSymbol',
  'indexInfo',
  'indexInfos',
  'declaredIndexInfos',
  'typePredicate',
  'parameterType',
  'propertyType',
  'mappedFields',
  'classThisType',
  'contextualArgument',
  // symbolRecord
  'symbol',
  'declarationSymbol',
  'aliasedSymbol',
  'property',
  'properties',
  'resolveName',
  'exportSymbol',
  'immediateAliasedSymbol',
  'shorthandValueSymbol',
  'exportSpecifierLocalTargetSymbol',
  'propertySymbolOfDestructuringAssignment',
  'mergedSymbol',
  'moduleMember',
  'parameterPropertySymbols',
  // signatureRecord
  'signature',
  'resolvedSignature',
  'signatures',
]);

export function typeQueryKey(q: TypeServiceQuery): string {
  const compact = q.skipRecords && DIRECT_RECORD_OPS.has(q.op);
  const prefix = `${q.generation}:${operationIDs.get(q.op) ?? 0}${compact ? 's' : ''}:`;
  switch (q.op) {
    case 'rootFileNames':
    case 'errorType':
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
    case 'void':
    case 'ambientModules':
      return prefix;
    case 'stringLiteralType':
      return `${prefix}${q.args.value.length}:${q.args.value.join(',')}`;
    case 'numberLiteralType':
      return `${prefix}${q.args.nonFinite ?? (Object.is(q.args.value, -0) ? '0' : String(q.args.value))}`;
    case 'bigIntLiteralType':
      return `${prefix}${q.args.negative && q.args.base10Value !== '0' ? '-' : ''}${q.args.base10Value}`;
    case 'nullableType':
      return `${prefix}${q.args.id}:${q.args.flags}`;
    case 'configFileParsingDiagnostics':
    case 'optionsDiagnostics':
    case 'globalDiagnostics':
    case 'syntacticDiagnostics':
    case 'semanticDiagnostics':
    case 'declarationDiagnostics':
      return prefix + formatNodeKey(q.args.node);
    case 'nodeRecord':
    case 'type':
    case 'fromTypeNode':
    case 'context':
    case 'symbol':
    case 'declarationSymbol':
    case 'signature':
    case 'shorthandValueSymbol':
    case 'suggestionDiagnostics':
    case 'exportSpecifierLocalTargetSymbol':
    case 'propertySymbolOfDestructuringAssignment':
    case 'jsxIntrinsicTagNamesAt':
    case 'constantValue':
    case 'isOptionalParameter':
    case 'isImplementationOfOverload':
      return prefix + formatNodeKey(q.args.node);
    case 'resolvedSignature':
      return `${prefix}${formatNodeKey(q.args.node)}:${q.args.withCandidates ? 1 : 0}:${q.args.argumentCount ?? '-'}`;
    case 'contextualArgument':
      return `${prefix}${formatNodeKey(q.args.node)}:${q.args.index}`;
    case 'validPropertyAccess':
    case 'parameterPropertySymbols':
      return `${prefix}${formatNodeKey(q.args.node)}:${q.args.name.length}:${q.args.name.join(',')}`;
    case 'property':
    case 'propertyType':
    case 'moduleMember':
      return `${prefix}${q.args.id}:${q.args.name.length}:${q.args.name.join(',')}`;
    case 'symbolType':
      return `${prefix}${q.args.id}:${formatNodeKey(q.args.node)}`;
    case 'signatures':
    case 'indexInfo':
      return `${prefix}${q.args.id}:${q.args.kind}`;
    case 'isTypeAssignableTo':
      return `${prefix}${q.args.source}:${q.args.target}`;
    case 'symbolsInScope':
      return `${prefix}${formatNodeKey(q.args.node)}:${q.args.meaning}`;
    case 'parameterType':
      return `${prefix}${q.args.id}:${q.args.position}`;
    case 'symbolDocumentation':
    case 'symbolJSDocTags':
      return `${prefix}${q.args.id}:${q.args.withChecker}`;
    case 'typeString':
    case 'typeNode':
      return `${prefix}${q.args.id}:${formatNodeKey(q.args.node)}:${q.args.flags ?? '-'}`;
    case 'typePredicateString':
      return `${prefix}${q.args.kind}:${q.args.parameterName ? `${q.args.parameterName.length}:${q.args.parameterName.join(',')}` : '-'}:${q.args.parameterIndex ?? '-'}:${q.args.type}:${formatNodeKey(q.args.node)}:${q.args.flags ?? '-'}`;
    case 'signatureString':
    case 'symbolString':
      return `${prefix}${q.args.id}:${formatNodeKey(q.args.node)}:${q.args.flags ?? '-'}:${q.args.kind}`;
    case 'resolveName':
      return `${prefix}${q.args.name.length}:${q.args.name.join(',')}:${formatNodeKey(q.args.node)}:${q.args.kind}:${q.args.excludeGlobals}`;
    case 'typeRecord':
    case 'typeDetails':
    case 'declaredIndexInfos':
    case 'mappedFields':
    case 'classThisType':
    case 'constraint':
    case 'apparent':
    case 'baseLiteral':
    case 'nonNullable':
    case 'awaited':
    case 'widened':
    case 'baseTypes':
    case 'defaultType':
    case 'properties':
    case 'augmentedPropertiesOfType':
    case 'typeArguments':
    case 'isArray':
    case 'isArrayLike':
    case 'isTuple':
    case 'numberIndexType':
    case 'stringIndexType':
    case 'indexInfos':
    case 'symbolRecord':
    case 'symbolMembers':
    case 'symbolExports':
    case 'fullyQualifiedName':
    case 'aliasedSymbol':
    case 'immediateAliasedSymbol':
    case 'isUnknownSymbol':
    case 'exportsOfModule':
    case 'exportSymbol':
    case 'declaredType':
    case 'typeOfSymbol':
    case 'rootSymbols':
    case 'mergedSymbol':
    case 'isArgumentsSymbol':
    case 'isUndefinedSymbol':
    case 'signatureRecord':
    case 'returnType':
    case 'typePredicate':
    case 'signatureDocumentation':
    case 'signatureJSDocTags':
      return `${prefix}${q.args.id}`;
  }
}

export type ScalarOperation = Exclude<
  Operation,
  'typeRecord' | 'symbolRecord' | 'signatureRecord' | 'nodeRecord'
>;

export const SCALAR_KINDS: Readonly<Record<ScalarOperation, number>> = {
  rootFileNames: 17,
  configFileParsingDiagnostics: 15,
  optionsDiagnostics: 15,
  globalDiagnostics: 15,
  syntacticDiagnostics: 15,
  semanticDiagnostics: 15,
  declarationDiagnostics: 15,
  errorType: 1,
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
  void: 1,
  stringLiteralType: 1,
  numberLiteralType: 1,
  bigIntLiteralType: 1,
  nullableType: 1,
  ambientModules: 11,
  type: 1,
  fromTypeNode: 1,
  context: 1,
  contextualArgument: 1,
  suggestionDiagnostics: 15,
  symbol: 1,
  declarationSymbol: 1,
  signature: 1,
  resolvedSignature: 4,
  shorthandValueSymbol: 1,
  exportSpecifierLocalTargetSymbol: 1,
  propertySymbolOfDestructuringAssignment: 1,
  jsxIntrinsicTagNamesAt: 11,
  constantValue: 16,
  isOptionalParameter: 2,
  isImplementationOfOverload: 2,
  validPropertyAccess: 2,
  parameterPropertySymbols: 4,
  symbolsInScope: 11,
  resolveName: 1,
  typeDetails: 18,
  declaredIndexInfos: 20,
  mappedFields: 4,
  classThisType: 1,
  constraint: 1,
  apparent: 1,
  baseLiteral: 1,
  nonNullable: 1,
  awaited: 1,
  widened: 1,
  baseTypes: 4,
  defaultType: 1,
  isTypeAssignableTo: 2,
  properties: 11,
  augmentedPropertiesOfType: 11,
  property: 1,
  propertyType: 1,
  typeArguments: 4,
  isArray: 2,
  isArrayLike: 2,
  isTuple: 2,
  numberIndexType: 1,
  stringIndexType: 1,
  indexInfo: 9,
  indexInfos: 9,
  signatures: 4,
  typeString: 3,
  typeNode: 14,
  typePredicateString: 3,
  symbolMembers: 19,
  symbolExports: 19,
  symbolDocumentation: 12,
  symbolJSDocTags: 13,
  fullyQualifiedName: 3,
  aliasedSymbol: 1,
  immediateAliasedSymbol: 1,
  isUnknownSymbol: 2,
  exportsOfModule: 11,
  exportSymbol: 1,
  declaredType: 1,
  typeOfSymbol: 1,
  symbolType: 1,
  rootSymbols: 11,
  mergedSymbol: 1,
  isArgumentsSymbol: 2,
  isUndefinedSymbol: 2,
  moduleMember: 1,
  symbolString: 3,
  returnType: 1,
  typePredicate: 10,
  parameterType: 1,
  signatureDocumentation: 12,
  signatureJSDocTags: 13,
  signatureString: 3,
};

const operationIDs = new Map(
  [
    ...Object.keys(SCALAR_KINDS),
    'typeRecord',
    'symbolRecord',
    'signatureRecord',
    'nodeRecord',
  ].map((op, index) => [op, index + 1] as const),
);

export type ServiceRequest<K extends Operation = Operation> = {
  [P in K]: { readonly op: P; readonly args: QueryArguments[P] };
}[K];
export type ScalarRequest = ServiceRequest<ScalarOperation>;

export function nameToUnits(name: string): number[] {
  const units: number[] = [];
  for (let i = 0; i < name.length; i++) units.push(name.charCodeAt(i));
  return units;
}
