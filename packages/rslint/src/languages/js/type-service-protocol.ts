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
) => { fileName: string; tree: Uint8Array } | undefined;
export interface TypeFact {
  readonly flags: readonly string[];
  readonly symbol: number;
  readonly aliasSymbol: number;
  readonly intrinsicName: string | undefined;
  readonly value:
    string | number | { negative: boolean; base10Value: string } | undefined;
  readonly types: readonly number[] | undefined;
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
}
type NodeArgs = { node: NodeRef };
type IDArgs = { id: number };
export interface QueryArguments {
  type: NodeArgs;
  fromTypeNode: NodeArgs;
  context: NodeArgs;
  symbol: NodeArgs;
  signature: NodeArgs;
  resolvedSignature: NodeArgs;
  constraint: IDArgs;
  apparent: IDArgs;
  baseLiteral: IDArgs;
  nonNullable: IDArgs;
  typeString: IDArgs;
  properties: IDArgs;
  typeArguments: IDArgs;
  isArray: IDArgs;
  isTuple: IDArgs;
  typeRecord: IDArgs;
  aliasedSymbol: IDArgs;
  declaredType: IDArgs;
  symbolRecord: IDArgs;
  returnType: IDArgs;
  signatureRecord: IDArgs;
  void: Record<string, never>;
  property: IDArgs & { name: readonly number[] };
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
  type: number;
  fromTypeNode: number;
  context: number;
  symbol: number;
  signature: number;
  resolvedSignature: number;
  constraint: number;
  apparent: number;
  baseLiteral: number;
  nonNullable: number;
  typeString: string;
  properties: readonly number[];
  typeArguments: readonly number[];
  isArray: boolean;
  isTuple: boolean;
  typeRecord: TypeFact;
  aliasedSymbol: number;
  declaredType: number;
  symbolRecord: SymbolFact;
  returnType: number;
  signatureRecord: SignatureFact;
  void: number;
  property: number;
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
const nodeKey = (node: NodeRef | undefined): string =>
  node ? `${node.file.length}:${node.file}:${node.index}` : '-';
export function typeQueryKey(q: TypeServiceQuery): string {
  const prefix = `${q.generation}:${q.checker}:${q.op}:`;
  switch (q.op) {
    case 'type':
    case 'fromTypeNode':
    case 'context':
    case 'symbol':
    case 'signature':
    case 'resolvedSignature':
      return prefix + nodeKey(q.args.node);
    case 'void':
      return prefix;
    case 'property':
      return `${prefix}${q.args.id}:${q.args.name.length}:${q.args.name.join(',')}`;
    case 'symbolType':
      return `${prefix}${q.args.id}:${nodeKey(q.args.node)}`;
    case 'signatures':
      return `${prefix}${q.args.id}:${q.args.kind}`;
    case 'resolveName':
      return `${prefix}${q.args.name.length}:${q.args.name.join(',')}:${nodeKey(q.args.node)}:${q.args.kind}:${q.args.excludeGlobals}`;
    default:
      return prefix + q.args.id;
  }
}

export type ScalarOperation = Exclude<
  Operation,
  'typeRecord' | 'symbolRecord' | 'signatureRecord'
>;
export const SCALAR_KINDS: Record<ScalarOperation, number> = {
  type: 1,
  fromTypeNode: 1,
  context: 1,
  symbol: 1,
  signature: 1,
  resolvedSignature: 1,
  constraint: 1,
  apparent: 1,
  baseLiteral: 1,
  nonNullable: 1,
  typeString: 3,
  properties: 4,
  typeArguments: 4,
  isArray: 2,
  isTuple: 2,
  aliasedSymbol: 1,
  declaredType: 1,
  returnType: 1,
  void: 1,
  property: 1,
  symbolType: 1,
  signatures: 4,
  resolveName: 1,
};

export type ServiceRequest<K extends Operation = Operation> = {
  [P in K]: { op: P; args: QueryArguments[P] };
}[K];
export type ScalarRequest = ServiceRequest<ScalarOperation>;

export function nameToUnits(name: string): number[] {
  const units: number[] = [];
  for (let i = 0; i < name.length; i++) units.push(name.charCodeAt(i));
  return units;
}
