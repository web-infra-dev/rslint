// AST Info types - matching the Go backend types

export interface NodeListMeta {
  pos: number;
  end: number;
  hasTrailingComma: boolean;
}

export interface NodeInfo {
  id?: number;
  kind: number;
  kindName: string;
  pos: number;
  end: number;
  flags: number;
  flagNames?: string[];
  text?: string;
  fileName?: string; // External file path (only set for nodes from lib.d.ts etc.)
  // Common node properties (shallow - only kind/pos/end for nested nodes)
  parent?: NodeInfo;
  name?: NodeInfo;
  expression?: NodeInfo;
  left?: NodeInfo;
  right?: NodeInfo;
  operatorToken?: NodeInfo;
  operand?: NodeInfo;
  condition?: NodeInfo;
  whenTrue?: NodeInfo;
  whenFalse?: NodeInfo;
  thenStatement?: NodeInfo;
  elseStatement?: NodeInfo;
  body?: NodeInfo;
  initializer?: NodeInfo;
  type?: NodeInfo;
  // Additional node properties for class/interface/function declarations
  members?: NodeInfo[];
  heritageClauses?: NodeInfo[];
  typeParameters?: NodeInfo[];
  parameters?: NodeInfo[];
  modifiers?: NodeInfo[];
  arguments?: NodeInfo[];
  statements?: NodeInfo[];
  properties?: NodeInfo[];
  elements?: NodeInfo[];

  // Variable/Declaration properties
  declarationList?: NodeInfo;
  declarations?: NodeInfo[];

  // Import/Export properties
  importClause?: NodeInfo;
  moduleSpecifier?: NodeInfo;
  namedBindings?: NodeInfo;
  exportClause?: NodeInfo;

  // Loop/Control flow properties
  incrementor?: NodeInfo;
  statement?: NodeInfo;

  // Switch statement properties
  caseBlock?: NodeInfo;
  clauses?: NodeInfo[];

  // Try/Catch properties
  tryBlock?: NodeInfo;
  catchClause?: NodeInfo;
  finallyBlock?: NodeInfo;
  variableDeclaration?: NodeInfo;
  block?: NodeInfo;

  // Property access properties
  argumentExpression?: NodeInfo;

  // Shorthand property assignment properties
  equalsToken?: NodeInfo;
  objectAssignmentInitializer?: NodeInfo;

  // Template literal properties
  head?: NodeInfo;
  templateSpans?: NodeInfo[];
  literal?: NodeInfo;
  tag?: NodeInfo;
  template?: NodeInfo;

  // Token properties (for optional/rest/generator/arrow/etc.)
  questionToken?: NodeInfo;
  dotDotDotToken?: NodeInfo;
  exclamationToken?: NodeInfo;
  asteriskToken?: NodeInfo;
  equalsGreaterThanToken?: NodeInfo;
  questionDotToken?: NodeInfo;

  // Type-related node properties
  typeArguments?: NodeInfo[];
  constraint?: NodeInfo;
  defaultType?: NodeInfo;

  // Locals - symbols declared in this node's scope
  locals?: SymbolInfo[];

  // SourceFile-specific properties
  endOfFileToken?: NodeInfo;
  imports?: NodeInfo[];
  isDeclarationFile?: boolean;
  scriptKind?: number; // 1=JS, 2=JSX, 3=TS, 4=TSX, 5=External, 6=JSON, 7=Deferred
  identifierCount?: number;
  symbolCount?: number;
  nodeCount?: number;

  // List metadata for array properties (Pos, End, HasTrailingComma)
  // Key is the property name e.g. "Parameters", "Arguments", "Members"
  listMetas?: Record<string, NodeListMeta>;
}

export interface IndexInfo {
  keyType: TypeInfo;
  valueType: TypeInfo;
  isReadonly?: boolean;
}

export interface TypeInfo {
  id?: number;
  flags: number;
  flagNames?: string[];
  objectFlags?: number;
  objectFlagNames?: string[];
  intrinsicName?: string;
  typeString: string;
  pos?: number; // Position for on-demand fetching
  fileName?: string; // External file path (only set for types from lib.d.ts etc.)
  // Literal type properties (only present for literal types)
  value?: unknown; // Literal value (string, number, bigint, boolean)
  freshType?: TypeInfo; // Fresh literal type (shallow)
  regularType?: TypeInfo; // Regular literal type (shallow)
  // Nested objects
  symbol?: SymbolInfo;
  aliasSymbol?: SymbolInfo;
  typeArguments?: TypeInfo[];
  baseTypes?: TypeInfo[];
  properties?: SymbolInfo[];
  callSignatures?: SignatureInfo[];
  constructSignatures?: SignatureInfo[];
  indexInfos?: IndexInfo[];
  types?: TypeInfo[];
  constraint?: TypeInfo;
  default?: TypeInfo;
  target?: TypeInfo; // Target type for type references
}

export interface SymbolInfo {
  id?: number;
  name: string;
  escapedName?: string;
  flags: number;
  flagNames?: string[];
  checkFlags?: number;
  checkFlagNames?: string[];
  pos?: number; // Position for on-demand fetching
  fileName?: string; // External file path (only set for symbols from lib.d.ts etc.)
  declarations?: NodeInfo[]; // Declarations as nodes (for lazy loading)
  valueDeclaration?: NodeInfo; // Value declaration as node (for lazy loading)
  members?: SymbolInfo[];
  exports?: SymbolInfo[];
}

export interface SignatureInfo {
  flags: number;
  flagNames?: string[];
  minArgumentCount: number;
  pos?: number; // Position for on-demand fetching
  fileName?: string; // External file path (only set for signatures from lib.d.ts etc.)
  // Parameters and thisParameter use Symbol data (shallow)
  parameters?: SymbolInfo[];
  thisParameter?: SymbolInfo;
  // Type parameters, return type, and type predicate use Type data (shallow)
  typeParameters?: TypeInfo[];
  returnType?: TypeInfo;
  typePredicate?: TypePredicateInfo;
  declaration?: NodeInfo; // Source declaration (shallow, supports lazy loading)
}

export interface TypeParamInfo {
  name: string;
  constraint?: TypeInfo;
  default?: TypeInfo;
}

export interface TypePredicateInfo {
  kind: number;
  kindName: string;
  parameterName?: string;
  parameterIndex?: number;
  type?: TypeInfo;
}

export interface FlowInfo {
  flags: number;
  flagNames?: string[];
  node?: NodeInfo;
  antecedent?: FlowInfo;
  antecedents?: FlowInfo[];
  graph?: FlowGraph; // Complete flow graph (only on top-level)
}

// Flow graph for visualization
export interface FlowGraph {
  nodes: FlowGraphNode[];
  edges: FlowEdge[];
}

export interface FlowGraphNode {
  id: number;
  flags: number;
  flagNames?: string[];
  nodePos?: number;
  nodeEnd?: number;
  nodeKindName?: string;
  nodeText?: string; // Text content for identifiers/literals
}

export interface FlowEdge {
  from: number; // Antecedent node ID
  to: number; // Current node ID
}

export interface GetAstInfoResponse {
  node?: NodeInfo;
  type?: TypeInfo;
  symbol?: SymbolInfo;
  signature?: SignatureInfo;
  flow?: FlowInfo;
}

export interface GetAstInfoRequest {
  fileContent: string;
  position: number;
  end?: number; // Optional: end position for exact node matching
  kind?: number; // Optional: filter by node kind (when multiple nodes at same position)
  fileName?: string; // Target file to query (empty for user file, external path like lib.d.ts for external file)
  compilerOptions?: Record<string, unknown>;
}
