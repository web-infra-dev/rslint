// Package inspector provides code inspection types and builders for the Playground
package inspector

// ============================================================================
// Request/Response Types
// ============================================================================

// GetAstInfoRequest represents a request for AST info at a specific position
type GetAstInfoRequest struct {
	FileContent     string         `json:"fileContent"`               // Source code content
	Position        int            `json:"position"`                  // Start position (pos)
	End             int            `json:"end,omitempty"`             // End position (optional, for exact node matching)
	Kind            int            `json:"kind,omitempty"`            // Optional: filter by node kind (when multiple nodes at same position)
	FileName        string         `json:"fileName,omitempty"`        // Target file to query (empty or /index.ts for user file, external path like lib.d.ts for external file)
	CompilerOptions map[string]any `json:"compilerOptions,omitempty"` // TypeScript compilerOptions (same as tsconfig.json)
}

// GetAstInfoResponse contains detailed AST information at a position
type GetAstInfoResponse struct {
	Node      *NodeInfo      `json:"node,omitempty"`
	Type      *TypeInfo      `json:"type,omitempty"`
	Symbol    *SymbolInfo    `json:"symbol,omitempty"`
	Signature *SignatureInfo `json:"signature,omitempty"`
	Flow      *FlowInfo      `json:"flow,omitempty"`
}

// ============================================================================
// Node Info
// ============================================================================

// NodeInfo contains detailed information about an AST node
type NodeInfo struct {
	Id        uint64   `json:"id,omitempty"`        // Node ID
	Kind      int      `json:"kind"`                // AST node kind number
	KindName  string   `json:"kindName"`            // Human-readable kind name
	Pos       int      `json:"pos"`                 // Start position (in FileName if external, otherwise in current file)
	End       int      `json:"end"`                 // End position
	Flags     int      `json:"flags"`               // Node flags
	FlagNames []string `json:"flagNames,omitempty"` // Human-readable flag names
	Text      string   `json:"text,omitempty"`      // Text content for identifiers/literals
	FileName  string   `json:"fileName,omitempty"`  // Source file path (only set for external files like lib.d.ts)

	// Common node properties (shallow - only kind/pos/end for nested nodes)
	Parent        *NodeInfo `json:"parent,omitempty"`        // Parent node
	Name          *NodeInfo `json:"name,omitempty"`          // Name node (for declarations)
	Expression    *NodeInfo `json:"expression,omitempty"`    // Expression node
	Left          *NodeInfo `json:"left,omitempty"`          // Left operand (BinaryExpression)
	Right         *NodeInfo `json:"right,omitempty"`         // Right operand (BinaryExpression)
	OperatorToken *NodeInfo `json:"operatorToken,omitempty"` // Operator token (BinaryExpression)
	Operand       *NodeInfo `json:"operand,omitempty"`       // Operand (UnaryExpression)
	Condition     *NodeInfo `json:"condition,omitempty"`     // Condition (ConditionalExpression, IfStatement)
	WhenTrue      *NodeInfo `json:"whenTrue,omitempty"`      // True branch (ConditionalExpression)
	WhenFalse     *NodeInfo `json:"whenFalse,omitempty"`     // False branch (ConditionalExpression)
	ThenStatement *NodeInfo `json:"thenStatement,omitempty"` // Then branch (IfStatement)
	ElseStatement *NodeInfo `json:"elseStatement,omitempty"` // Else branch (IfStatement)
	Body          *NodeInfo `json:"body,omitempty"`          // Body node (functions, loops, etc.)
	Initializer   *NodeInfo `json:"initializer,omitempty"`   // Initializer (VariableDeclaration, etc.)
	TypeNode      *NodeInfo `json:"type,omitempty"`          // Type annotation node

	// Additional node properties for class/interface/function declarations
	Members         []*NodeInfo `json:"members,omitempty"`         // Members (ClassDeclaration, InterfaceDeclaration)
	HeritageClauses []*NodeInfo `json:"heritageClauses,omitempty"` // Heritage clauses (extends, implements)
	TypeParameters  []*NodeInfo `json:"typeParameters,omitempty"`  // Type parameters (generic declarations)
	Parameters      []*NodeInfo `json:"parameters,omitempty"`      // Parameters (FunctionDeclaration, MethodDeclaration)
	Modifiers       []*NodeInfo `json:"modifiers,omitempty"`       // Modifiers (public, private, static, etc.)
	Arguments       []*NodeInfo `json:"arguments,omitempty"`       // Arguments (CallExpression)
	Statements      []*NodeInfo `json:"statements,omitempty"`      // Statements (Block, SourceFile)
	Properties      []*NodeInfo `json:"properties,omitempty"`      // Properties (ObjectLiteralExpression)
	Elements        []*NodeInfo `json:"elements,omitempty"`        // Elements (ArrayLiteralExpression)

	// Variable/Declaration properties
	DeclarationList *NodeInfo   `json:"declarationList,omitempty"` // Declaration list (VariableStatement)
	Declarations    []*NodeInfo `json:"declarations,omitempty"`    // Declarations (VariableDeclarationList)

	// Import/Export properties
	ImportClause    *NodeInfo `json:"importClause,omitempty"`    // Import clause (ImportDeclaration)
	ModuleSpecifier *NodeInfo `json:"moduleSpecifier,omitempty"` // Module specifier (ImportDeclaration, ExportDeclaration)
	NamedBindings   *NodeInfo `json:"namedBindings,omitempty"`   // Named bindings (ImportClause)
	ExportClause    *NodeInfo `json:"exportClause,omitempty"`    // Export clause (ExportDeclaration)

	// Loop/Control flow properties
	Incrementor *NodeInfo `json:"incrementor,omitempty"` // Incrementor (ForStatement)
	Statement   *NodeInfo `json:"statement,omitempty"`   // Statement (loop body)

	// Switch statement properties
	CaseBlock *NodeInfo   `json:"caseBlock,omitempty"` // Case block (SwitchStatement)
	Clauses   []*NodeInfo `json:"clauses,omitempty"`   // Clauses (CaseBlock)

	// Try/Catch properties
	TryBlock            *NodeInfo `json:"tryBlock,omitempty"`            // Try block
	CatchClause         *NodeInfo `json:"catchClause,omitempty"`         // Catch clause
	FinallyBlock        *NodeInfo `json:"finallyBlock,omitempty"`        // Finally block
	VariableDeclaration *NodeInfo `json:"variableDeclaration,omitempty"` // Variable declaration (CatchClause)
	Block               *NodeInfo `json:"block,omitempty"`               // Block (CatchClause)

	// Property access properties
	ArgumentExpression *NodeInfo `json:"argumentExpression,omitempty"` // Argument expression (ElementAccessExpression)

	// Shorthand property assignment properties
	EqualsToken                 *NodeInfo `json:"equalsToken,omitempty"`                 // Equals token (ShorthandPropertyAssignment)
	ObjectAssignmentInitializer *NodeInfo `json:"objectAssignmentInitializer,omitempty"` // Object assignment initializer

	// Token properties (for optional/rest/generator/arrow/etc.)
	QuestionToken          *NodeInfo `json:"questionToken,omitempty"`          // Optional marker token (?)
	DotDotDotToken         *NodeInfo `json:"dotDotDotToken,omitempty"`         // Rest parameter token (...)
	ExclamationToken       *NodeInfo `json:"exclamationToken,omitempty"`       // Definite assignment token (!)
	AsteriskToken          *NodeInfo `json:"asteriskToken,omitempty"`          // Generator token (*)
	EqualsGreaterThanToken *NodeInfo `json:"equalsGreaterThanToken,omitempty"` // Arrow function token (=>)
	QuestionDotToken       *NodeInfo `json:"questionDotToken,omitempty"`       // Optional chaining token (?.)

	// Type-related node properties
	TypeArguments []*NodeInfo `json:"typeArguments,omitempty"` // Type arguments (CallExpression, NewExpression)
	Constraint    *NodeInfo   `json:"constraint,omitempty"`    // Type constraint (TypeParameterDeclaration)
	DefaultType   *NodeInfo   `json:"defaultType,omitempty"`   // Default type (TypeParameterDeclaration)

	// Template literal properties
	Head          *NodeInfo   `json:"head,omitempty"`          // Template head
	TemplateSpans []*NodeInfo `json:"templateSpans,omitempty"` // Template spans
	Literal       *NodeInfo   `json:"literal,omitempty"`       // Literal (TemplateSpan)
	Tag           *NodeInfo   `json:"tag,omitempty"`           // Tag (TaggedTemplateExpression)
	Template      *NodeInfo   `json:"template,omitempty"`      // Template (TaggedTemplateExpression)

	// Locals - symbols declared in this node's scope (shallow)
	Locals []*SymbolInfo `json:"locals,omitempty"`

	// SourceFile-specific properties
	EndOfFileToken    *NodeInfo   `json:"endOfFileToken,omitempty"`    // End of file token (SourceFile)
	Imports           []*NodeInfo `json:"imports,omitempty"`           // Import declarations (SourceFile)
	IsDeclarationFile bool        `json:"isDeclarationFile,omitempty"` // Whether this is a .d.ts file (SourceFile)
	ScriptKind        int         `json:"scriptKind,omitempty"`        // Script kind: 1=JS, 2=JSX, 3=TS, 4=TSX, 5=External, 6=JSON, 7=Deferred (SourceFile)
	IdentifierCount   int         `json:"identifierCount,omitempty"`   // Number of identifiers in file (SourceFile)
	SymbolCount       int         `json:"symbolCount,omitempty"`       // Number of symbols in file (SourceFile)
	NodeCount         int         `json:"nodeCount,omitempty"`         // Number of nodes in file (SourceFile)

	// List metadata for array properties (Pos, End, HasTrailingComma)
	// Key is the property name e.g. "Parameters", "Arguments", "Members"
	ListMetas map[string]*NodeListMeta `json:"listMetas,omitempty"`
}

// NodeListMeta contains metadata about a NodeList
type NodeListMeta struct {
	Pos              int  `json:"pos"`
	End              int  `json:"end"`
	HasTrailingComma bool `json:"hasTrailingComma"`
}

// ============================================================================
// Type Info
// ============================================================================

// TypeInfo contains detailed information about a TypeScript type
// For nested types, only basic info is included; use the declaration position to fetch full details
type TypeInfo struct {
	Id              uint32   `json:"id,omitempty"`              // Type ID
	Flags           uint32   `json:"flags"`                     // TypeFlags
	FlagNames       []string `json:"flagNames,omitempty"`       // Human-readable flag names
	ObjectFlags     uint32   `json:"objectFlags,omitempty"`     // ObjectFlags for object types
	ObjectFlagNames []string `json:"objectFlagNames,omitempty"` // Human-readable object flag names
	IntrinsicName   string   `json:"intrinsicName,omitempty"`   // Name for intrinsic types
	TypeString      string   `json:"typeString"`                // String representation of the type

	// Position for on-demand fetching (from symbol's declaration)
	Pos      int    `json:"pos,omitempty"`      // Position to fetch full type info
	FileName string `json:"fileName,omitempty"` // Source file path (only set for external files)

	// Literal type properties (only present for literal types)
	Value       any       `json:"value,omitempty"`       // Literal value (string, number, bigint, boolean)
	FreshType   *TypeInfo `json:"freshType,omitempty"`   // Fresh literal type (shallow)
	RegularType *TypeInfo `json:"regularType,omitempty"` // Regular literal type (shallow)

	// Nested objects (shallow - only basic info included)
	Symbol              *SymbolInfo      `json:"symbol,omitempty"`              // Associated symbol (shallow)
	AliasSymbol         *SymbolInfo      `json:"aliasSymbol,omitempty"`         // Type alias symbol (shallow)
	TypeArguments       []*TypeInfo      `json:"typeArguments,omitempty"`       // Generic type arguments (shallow)
	BaseTypes           []*TypeInfo      `json:"baseTypes,omitempty"`           // Base types for class/interface (shallow)
	Properties          []*SymbolInfo    `json:"properties,omitempty"`          // Type properties (shallow)
	CallSignatures      []*SignatureInfo `json:"callSignatures,omitempty"`      // Call signatures
	ConstructSignatures []*SignatureInfo `json:"constructSignatures,omitempty"` // Construct signatures
	IndexInfos          []*IndexInfoType `json:"indexInfos,omitempty"`          // Index signatures
	Types               []*TypeInfo      `json:"types,omitempty"`               // Union/Intersection type members (shallow)
	Constraint          *TypeInfo        `json:"constraint,omitempty"`          // Type parameter constraint (shallow)
	Default             *TypeInfo        `json:"default,omitempty"`             // Type parameter default (shallow)
	Target              *TypeInfo        `json:"target,omitempty"`              // Target type for type references (shallow)
}

// ============================================================================
// Index Info
// ============================================================================

// IndexInfoType contains information about an index signature
type IndexInfoType struct {
	KeyType    *TypeInfo `json:"keyType"`              // Key type (string, number, symbol)
	ValueType  *TypeInfo `json:"valueType"`            // Value type
	IsReadonly bool      `json:"isReadonly,omitempty"` // Whether the index signature is readonly
}

// ============================================================================
// Symbol Info
// ============================================================================

// SymbolInfo contains detailed information about a TypeScript symbol
// For nested symbols, only basic info is included; use the declaration position to fetch full details
type SymbolInfo struct {
	Id             uint64   `json:"id,omitempty"`             // Symbol ID
	Name           string   `json:"name"`                     // Symbol name (formatted for display)
	EscapedName    string   `json:"escapedName,omitempty"`    // Internal escaped name (formatted, only if different from Name)
	Flags          uint32   `json:"flags"`                    // SymbolFlags
	FlagNames      []string `json:"flagNames,omitempty"`      // Human-readable flag names
	CheckFlags     uint32   `json:"checkFlags"`               // CheckFlags (for transient symbols)
	CheckFlagNames []string `json:"checkFlagNames,omitempty"` // Human-readable check flag names

	// Position for on-demand fetching (from valueDeclaration or declarations[0])
	Pos      int    `json:"pos,omitempty"`      // Position to fetch full symbol info
	FileName string `json:"fileName,omitempty"` // Source file path (only set for external files)

	// Declaration nodes (shallow NodeInfo for lazy loading)
	Declarations     []*NodeInfo `json:"declarations,omitempty"`     // All declarations (as nodes)
	ValueDeclaration *NodeInfo   `json:"valueDeclaration,omitempty"` // Primary value declaration (as node)

	// Nested objects (shallow - only basic info included)
	Members []*SymbolInfo `json:"members,omitempty"` // Member symbols (shallow)
	Exports []*SymbolInfo `json:"exports,omitempty"` // Exported symbols (shallow)
}


// ============================================================================
// Signature Info
// ============================================================================

// SignatureInfo contains detailed information about a function/method signature
// For nested types, only basic info is included; use the declaration position to fetch full details
type SignatureInfo struct {
	Flags            uint32   `json:"flags"`                   // SignatureFlags
	FlagNames        []string `json:"flagNames,omitempty"`     // Human-readable flag names
	MinArgumentCount int      `json:"minArgumentCount"`        // Minimum required arguments

	// Position for on-demand fetching (from declaration)
	Pos      int    `json:"pos,omitempty"`      // Position to fetch full signature info
	FileName string `json:"fileName,omitempty"` // Source file path (only set for external files)

	// Parameters and thisParameter use Symbol data (shallow)
	Parameters    []*SymbolInfo `json:"parameters,omitempty"`    // Parameter symbols (shallow)
	ThisParameter *SymbolInfo   `json:"thisParameter,omitempty"` // Explicit 'this' parameter symbol (shallow)

	// Type parameters, return type, and type predicate use Type data (shallow)
	TypeParameters []*TypeInfo        `json:"typeParameters,omitempty"` // Generic type parameters (shallow)
	ReturnType     *TypeInfo          `json:"returnType,omitempty"`     // Return type (shallow)
	TypePredicate  *TypePredicateInfo `json:"typePredicate,omitempty"`  // Type predicate for type guards

	// Declaration node (shallow)
	Declaration *NodeInfo `json:"declaration,omitempty"` // Source declaration (shallow)
}

// TypePredicateInfo contains information about a type predicate
type TypePredicateInfo struct {
	Kind           int       `json:"kind"`                    // TypePredicateKind numeric value
	KindName       string    `json:"kindName"`                // Human-readable kind name
	ParameterName  string    `json:"parameterName,omitempty"`
	ParameterIndex int       `json:"parameterIndex"`          // 0 is valid (first param), don't use omitempty
	Type           *TypeInfo `json:"type,omitempty"`
}

// ============================================================================
// Flow Info
// ============================================================================

// FlowInfo contains control flow analysis information
type FlowInfo struct {
	Flags       uint32      `json:"flags"`                 // FlowFlags
	FlagNames   []string    `json:"flagNames,omitempty"`   // Human-readable flag names
	Node        *NodeInfo   `json:"node,omitempty"`        // Associated AST node
	Antecedent  *FlowInfo   `json:"antecedent,omitempty"`  // Single antecedent
	Antecedents []*FlowInfo `json:"antecedents,omitempty"` // Multiple antecedents for labels
	Graph       *FlowGraph  `json:"graph,omitempty"`       // Complete flow graph (only on top-level)
}

// FlowGraph contains the complete flow graph for visualization
type FlowGraph struct {
	Nodes []*FlowGraphNode `json:"nodes"`
	Edges []*FlowEdge      `json:"edges"`
}

// FlowGraphNode represents a node in the flow graph
type FlowGraphNode struct {
	Id           uint64   `json:"id"`
	Flags        uint32   `json:"flags"`
	FlagNames    []string `json:"flagNames,omitempty"`
	NodePos      int      `json:"nodePos,omitempty"`
	NodeEnd      int      `json:"nodeEnd,omitempty"`
	NodeKindName string   `json:"nodeKindName,omitempty"`
	NodeText     string   `json:"nodeText,omitempty"` // Text content for identifiers/literals
}

// FlowEdge represents an edge in the flow graph (from antecedent to current)
type FlowEdge struct {
	From uint64 `json:"from"` // Antecedent node ID
	To   uint64 `json:"to"`   // Current node ID
}
