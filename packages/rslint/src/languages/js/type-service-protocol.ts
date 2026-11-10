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
export interface SymbolFact {
  readonly name: string;
  readonly escapedName: string;
  readonly flags: number;
  readonly declarations: readonly NodeRef[] | undefined;
  readonly valueDeclaration: NodeRef | undefined;
}
export interface SignatureFact {
  readonly parameters: readonly number[];
  readonly declaration: NodeRef | undefined;
  readonly typeParameters: readonly number[] | undefined;
  readonly thisParameter: number;
  readonly minArgumentCount: number;
}
export interface IndexInfoFact {
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
  nodeRecord: NodeArgs;
  type: NodeArgs;
  fromTypeNode: NodeArgs;
  context: NodeArgs;
  symbol: NodeArgs;
  signature: NodeArgs;
  resolvedSignature: NodeArgs;
  shorthandValueSymbol: NodeArgs;
  constraint: IDArgs;
  apparent: IDArgs;
  baseLiteral: IDArgs;
  nonNullable: IDArgs;
  typeString: PrintArgs;
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
  aliasedSymbol: IDArgs;
  immediateAliasedSymbol: IDArgs;
  isUnknownSymbol: IDArgs;
  exportsOfModule: IDArgs;
  exportSymbol: IDArgs;
  symbolsInScope: NodeArgs & { meaning: number };
  declaredType: IDArgs;
  typeOfSymbol: IDArgs;
  symbolRecord: IDArgs;
  returnType: IDArgs;
  typePredicate: IDArgs;
  parameterType: IDArgs & { position: number };
  signatureRecord: IDArgs;
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
  nodeRecord: readonly NodeFact[];
  type: number;
  fromTypeNode: number;
  context: number;
  symbol: number;
  signature: number;
  resolvedSignature: number;
  shorthandValueSymbol: number;
  constraint: number;
  apparent: number;
  baseLiteral: number;
  nonNullable: number;
  typeString: string;
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
  aliasedSymbol: number;
  immediateAliasedSymbol: number;
  isUnknownSymbol: boolean;
  exportsOfModule: readonly NamedSymbolFact[];
  exportSymbol: number;
  symbolsInScope: readonly NamedSymbolFact[];
  declaredType: number;
  typeOfSymbol: number;
  symbolRecord: SymbolFact;
  returnType: number;
  typePredicate: TypePredicateFact | undefined;
  parameterType: number;
  signatureRecord: SignatureFact;
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
) => Uint8Array;
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
    case 'nodeRecord':
    case 'type':
    case 'fromTypeNode':
    case 'context':
    case 'symbol':
    case 'signature':
    case 'resolvedSignature':
    case 'shorthandValueSymbol':
      return prefix + nodeKey(q.args.node);
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
    case 'typeString':
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
  type: 1,
  fromTypeNode: 1,
  context: 1,
  symbol: 1,
  signature: 1,
  resolvedSignature: 1,
  shorthandValueSymbol: 1,
  constraint: 1,
  apparent: 1,
  baseLiteral: 1,
  nonNullable: 1,
  typeString: 3,
  signatureString: 3,
  awaited: 1,
  widened: 1,
  baseTypes: 4,
  defaultType: 1,
  isTypeAssignableTo: 2,
  properties: 4,
  apparentProperties: 4,
  typeArguments: 4,
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
  returnType: 1,
  typePredicate: 10,
  parameterType: 1,
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
