/**
 * Type definitions for RemoteTypeChecker
 * These types mirror the Go-side types in internal/api/checker_session.go
 */

/**
 * NodeLocation identifies a node in a source file using structured parameters.
 * This replaces the magic string format "filePath.pos.kind".
 */
export interface NodeLocation {
  /** Relative file path (e.g., "index.ts") */
  filePath: string;
  /** Start position of the node */
  pos: number;
  /** SyntaxKind of the node */
  kind: number;
}

/**
 * Detailed type information for NodeTypeInfo response.
 *
 * Note: Field names use PascalCase to match Go-side JSON output directly,
 * enabling frontend display of field names as-is. This differs from the
 * lightweight *Info types above which use camelCase.
 */

/**
 * Type information with full details
 */
export interface TypeDetails {
  Id: number;
  Flags: number;
  FlagNames: string[];
  ObjectFlags: number;
  ObjectFlagNames: string[];
  Symbol?: number;
  Target?: number;
  Types?: number[];
  TypeString: string;
  IntrinsicName?: string;
  Value?: string | number | boolean;
  TypeParameters?: number[];
  FixedLength?: number;
  ElementInfos?: Array<{ Flags: number }>;
  Properties?: number[];
  CallSignatures?: string[];
  ConstructSignatures?: string[];
}

/**
 * Symbol information with full details
 */
export interface SymbolDetails {
  Id: number;
  Flags: number;
  FlagNames: string[];
  CheckFlags: number;
  CheckFlagNames: string[];
  Name: string;
  SymbolString: string;
  Declarations?: NodeLocation[];
  ValueDeclaration?: NodeLocation;
  Members?: Record<string, number>;
  Exports?: Record<string, number>;
  Parent?: number;
}

/**
 * Signature information with full details
 * Note: Signature has no internal ID in typescript-go
 */
export interface SignatureDetails {
  SignatureString: string;
  TypeParameters?: number[];
  Parameters: Array<{
    Name: string;
    SymbolId: number;
  }>;
  ThisParameter?: {
    Name: string;
    SymbolId: number;
  };
  HasRestParameter: boolean;
  ReturnType?: number;
  Declaration?: NodeLocation;
}

/**
 * FlowNode information with full details
 * Note: FlowNode has no internal ID in typescript-go
 */
export interface FlowNodeDetails {
  Flags: number;
  FlagNames: string[];
  Node?: NodeLocation;
  Antecedent?: FlowNodeDetails;
  Antecedents?: FlowNodeDetails[];
}

/**
 * Response from getNodeType (lazy loading)
 */
export interface NodeTypeResponse {
  Type?: TypeDetails;
  ContextualType?: TypeDetails;
  /** Related types collected during traversal (for reference lookup by ID) */
  RelatedTypes?: Record<number, TypeDetails>;
  /** Related symbols collected during traversal (for reference lookup by ID) */
  RelatedSymbols?: Record<number, SymbolDetails>;
}

/**
 * Response from getNodeSymbol (lazy loading)
 */
export interface NodeSymbolResponse {
  Symbol?: SymbolDetails;
  /** Related types collected during traversal (for reference lookup by ID) */
  RelatedTypes?: Record<number, TypeDetails>;
  /** Related symbols collected during traversal (for reference lookup by ID) */
  RelatedSymbols?: Record<number, SymbolDetails>;
}

/**
 * Response from getNodeSignature (lazy loading)
 */
export interface NodeSignatureResponse {
  Signature?: SignatureDetails;
  /** Related types collected during traversal (for reference lookup by ID) */
  RelatedTypes?: Record<number, TypeDetails>;
  /** Related symbols collected during traversal (for reference lookup by ID) */
  RelatedSymbols?: Record<number, SymbolDetails>;
}

/**
 * Response from getNodeFlowNode (lazy loading)
 */
export interface NodeFlowNodeResponse {
  FlowNode?: FlowNodeDetails;
}

/**
 * Response from getNodeInfo (lazy loading)
 */
export interface NodeInfoResponse {
  Kind: number;
  KindName: string;
  Flags: number;
  FlagNames: string[];
  ModifierFlags: number;
  ModifierFlagNames: string[];
  Pos: number;
  End: number;
}
