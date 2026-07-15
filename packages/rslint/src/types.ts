/**
 * Shared types for rslint IPC protocol across all environments
 */
import type {
  ActivateConfigsRequest,
  LoadConfigsRequest,
} from './config/config-discovery-protocol.js';

export interface Position {
  line: number;
  column: number;
}

export interface Range {
  start: Position;
  end: Position;
}

export interface Fix {
  text: string;
  startPos: number;
  endPos: number;
}

// An optional, user-selected fix (ESLint's suggestions): surfaced for the
// editor/user to pick, NOT applied by `fix: true`. `data` carries the
// messageId's placeholder values (ESLint v10 suggestion.data).
export interface Suggestion {
  messageId: string;
  message: string;
  data?: Record<string, string>;
  fixes?: Fix[];
}

export interface Diagnostic {
  ruleName: string;
  message: string;
  messageId: string;
  filePath: string;
  range: Range;
  severity?: string;
  fixes?: Fix[];
  suggestions?: Suggestion[];
}

export interface LintResponse {
  diagnostics: Diagnostic[];
  // errorCount / warningCount are split by severity (ESLint semantics):
  // errorCount counts only error-severity diagnostics, not the total.
  errorCount: number;
  warningCount: number;
  // fixableErrorCount / fixableWarningCount count the auto-fixable subset.
  fixableErrorCount: number;
  fixableWarningCount: number;
  fileCount: number;
  ruleCount: number;
  // Files actually linted (config `ignores` excluded), each a requested target
  // path relative to configDirectory — same path space as Diagnostic.filePath.
  // Present for lintFiles so the Rslint class seeds one result per linted file.
  lintedFiles?: string[];
  output?: Record<string, string>; // Per-file fixed source, present when fix:true applied a fix
  encodedSourceFiles?: Record<string, string>; // Binary encoded source files as base64-encoded strings
}

export interface LintOptions {
  files?: string[];
  // Optional physical paths parallel to files. High-level Node APIs provide
  // these after target planning so Go does not repeat realpath resolution.
  canonicalFiles?: string[];
  // A caller-supplied, already normalized config (normalizeConfig output:
  // plain objects, plugins as string[], no live functions). High-level Node
  // APIs normally use configDiscovery instead; overrideConfigFile:true uses
  // this self-contained path. Empty/absent = no config.
  //
  // Listing a plugin does NOT auto-enable its rules: every rule must be named
  // explicitly under an entry's `rules` (or pulled in via a preset that does
  // so), matching ESLint flat-config semantics.
  config?: Record<string, unknown>[];
  /**
   * Ask native Go to discover and route JS/TS configs for this lint request.
   * `config` and `configDiscovery` are mutually exclusive. Browser/WASM
   * backends intentionally do not advertise this host-filesystem capability.
   */
  configDiscovery?: {
    mode: 'auto' | 'explicit';
    explicitConfigPath?: string;
    /** Static glob roots; Go visits only branches leading to the supplied files. */
    directories?: string[];
    /** Parallel to `files`; true only for caller-literal file targets. */
    explicitFiles?: boolean[];
    /** Normalized API override entries appended to every selected config. */
    overrideConfig?: Record<string, unknown>[];
  };
  // Community ESLint-plugin rules available to this request. Go registers
  // request-scoped placeholder rules from this metadata, then asks the Node
  // peer to execute them through a reverse `pluginLint` request.
  eslintPlugins?: Array<{ prefix: string; ruleNames: string[] }>;
  // Anchor dir for resolving the config's relative files/ignores/project.
  configDirectory?: string;
  // Opaque routing key for community-plugin workers. High-level APIs set this
  // when config path rebasing makes it differ from configDirectory.
  pluginConfigDirectory?: string;
  workingDirectory?: string;
  fileContents?: Record<string, string>; // Map of file paths to their contents for VFS
  includeEncodedSourceFiles?: boolean; // Whether to include encoded source files in response
  // Apply rule auto-fixes in-band (ESLint's `fix: true`); the fixed source per
  // file is returned in LintResponse.output and is NOT written to disk. Rules
  // and languageOptions live in the config entries — there is no separate
  // ruleOptions / languageOptions override surface.
  fix?: boolean;
}

export interface RSlintOptions {
  rslintPath?: string;
  workingDirectory?: string;
}

export interface PendingMessage {
  resolve: (data: any) => void;
  reject: (error: Error) => void;
}

export interface IpcMessage {
  id: number;
  kind: string;
  data: any;
}

/** Handler for a positive-id request frame sent by the Go peer. */
export type InboundRequestHandler = (message: IpcMessage) => unknown;

/** Reverse-request handlers that are scoped to one outer lint request. */
export interface LintInboundHandlers {
  pluginLint?: (request: unknown) => unknown;
  loadConfigs?: (request: LoadConfigsRequest) => unknown;
  activateConfigs?: (request: ActivateConfigsRequest) => unknown;
}

// Service interface that all implementations must follow
export interface RslintServiceInterface {
  sendMessage(kind: string, data: any): Promise<any>;
  /** Optional for one-way backends such as the current browser worker. */
  setInboundHandler?(handler: InboundRequestHandler | null): void;
  terminate(): void;
}

// ============================================================================
// AST Info Types - Used by Playground to display detailed AST information
// ============================================================================

/**
 * Request for AST info at a specific position
 */
export interface GetAstInfoRequest {
  fileContent: string;
  position: number;
  end?: number; // End position (optional, for exact node matching)
  kind?: number; // Optional: filter by node kind (when multiple nodes at same position)
  depth?: number; // Max recursion depth (default: 2)
  fileName?: string; // Target file to query (empty for user file, external path like lib.d.ts for external file)
  compilerOptions?: Record<string, unknown>; // TypeScript compilerOptions (same format as tsconfig.json)
}

/**
 * Response containing detailed AST information
 */
export interface GetAstInfoResponse {
  node?: NodeInfo;
  type?: TypeInfo;
  symbol?: SymbolInfo;
  signature?: SignatureInfo;
  flow?: FlowInfo;
}

/**
 * NodeList metadata (Pos, End, HasTrailingComma)
 */
export interface NodeListMeta {
  pos: number;
  end: number;
  hasTrailingComma: boolean;
}

/**
 * Detailed information about an AST node
 */
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

/**
 * Detailed information about a TypeScript type
 */
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
  types?: TypeInfo[]; // Union/Intersection type members
  constraint?: TypeInfo;
  default?: TypeInfo;
  target?: TypeInfo; // Target type for type references
}

/**
 * Index signature information
 */
export interface IndexInfo {
  keyType: TypeInfo;
  valueType: TypeInfo;
  isReadonly: boolean;
}

/**
 * Detailed information about a TypeScript symbol
 */
export interface SymbolInfo {
  id?: number;
  name: string;
  escapedName?: string;
  flags: number;
  flagNames?: string[];
  checkFlags?: number;
  checkFlagNames?: string[];
  pos?: number;
  fileName?: string; // External file path (only set for symbols from lib.d.ts etc.)
  declarations?: NodeInfo[];
  valueDeclaration?: NodeInfo;
  members?: SymbolInfo[];
  exports?: SymbolInfo[];
}

/**
 * Detailed information about a function/method signature
 */
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

/**
 * Function parameter information (deprecated - use SymbolInfo instead)
 */
export interface ParameterInfo {
  name: string;
  type?: TypeInfo;
  optional: boolean;
  rest: boolean;
}

/**
 * Generic type parameter information
 */
export interface TypeParamInfo {
  name: string;
  constraint?: TypeInfo;
  default?: TypeInfo;
}

/**
 * Type predicate information for type guards
 */
export interface TypePredicateInfo {
  kind: number;
  kindName: string;
  parameterName?: string;
  parameterIndex?: number;
  type?: TypeInfo;
}

/**
 * Control flow analysis information
 */
export interface FlowInfo {
  flags: number;
  flagNames?: string[];
  nodeKind?: number;
  nodeKindName?: string;
  nodePos?: number;
  nodeEnd?: number;
  antecedent?: FlowInfo;
  antecedents?: FlowInfo[];
  graph?: FlowGraph; // Complete flow graph (only on top-level)
}

/**
 * Flow graph for visualization
 */
export interface FlowGraph {
  nodes: FlowGraphNode[];
  edges: FlowEdge[];
}

/**
 * Node in the flow graph
 */
export interface FlowGraphNode {
  id: number;
  flags: number;
  flagNames?: string[];
  nodePos?: number;
  nodeEnd?: number;
  nodeKindName?: string;
}

/**
 * Edge in the flow graph (from antecedent to current)
 */
export interface FlowEdge {
  from: number; // Antecedent node ID
  to: number; // Current node ID
}
