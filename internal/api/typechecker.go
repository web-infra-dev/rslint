// Package ipc provides IPC communication between JS and Go using stdio
package ipc

import (
	"context"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
)

// TypeChecker holds the state for type checking
type TypeChecker struct {
	programs    []*compiler.Program
	checkers    map[string]*checker.Checker // filePath -> Checker
	sourceFiles map[string]*ast.SourceFile  // filePath -> SourceFile
	mu          sync.RWMutex
}

// NewTypeChecker creates a new TypeChecker from programs and source files
// Note: sourceFiles keys should be the same paths used for checker lookups
func NewTypeChecker(
	programs []*compiler.Program,
	sourceFiles map[string]*ast.SourceFile,
) *TypeChecker {
	checkers := make(map[string]*checker.Checker)

	// Build a reverse map from absolute path to relative path
	absToRelPath := make(map[string]string)
	for relPath, sf := range sourceFiles {
		absToRelPath[sf.FileName()] = relPath
	}

	// Get type checkers from programs, using the same keys as sourceFiles
	for _, program := range programs {
		c, _ := program.GetTypeChecker(context.Background())
		for _, sf := range program.GetSourceFiles() {
			// Use the relative path as key (matching sourceFiles)
			if relPath, ok := absToRelPath[sf.FileName()]; ok {
				checkers[relPath] = c
			} else {
				// Fallback to absolute path if not in our sourceFiles map
				checkers[sf.FileName()] = c
			}
		}
	}

	return &TypeChecker{
		programs:    programs,
		checkers:    checkers,
		sourceFiles: sourceFiles,
	}
}

// Helper functions

// findTokenAtPosition finds the token at the given position in a source file
func findTokenAtPosition(sourceFile *ast.SourceFile, position int) *ast.Node {
	var result *ast.Node

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Pos() <= position && position < node.End() {
			result = node
			node.ForEachChild(visit)
		}
		return false
	}

	sourceFile.AsNode().ForEachChild(visit)
	return result
}

// getSymbolFlagNames converts symbol flags to a slice of flag names
func getSymbolFlagNames(flags ast.SymbolFlags) []string {
	var names []string

	flagMap := map[ast.SymbolFlags]string{
		ast.SymbolFlagsFunctionScopedVariable: "FunctionScopedVariable",
		ast.SymbolFlagsBlockScopedVariable:    "BlockScopedVariable",
		ast.SymbolFlagsProperty:               "Property",
		ast.SymbolFlagsEnumMember:             "EnumMember",
		ast.SymbolFlagsFunction:               "Function",
		ast.SymbolFlagsClass:                  "Class",
		ast.SymbolFlagsInterface:              "Interface",
		ast.SymbolFlagsConstEnum:              "ConstEnum",
		ast.SymbolFlagsRegularEnum:            "RegularEnum",
		ast.SymbolFlagsValueModule:            "ValueModule",
		ast.SymbolFlagsNamespaceModule:        "NamespaceModule",
		ast.SymbolFlagsTypeLiteral:            "TypeLiteral",
		ast.SymbolFlagsObjectLiteral:          "ObjectLiteral",
		ast.SymbolFlagsMethod:                 "Method",
		ast.SymbolFlagsConstructor:            "Constructor",
		ast.SymbolFlagsGetAccessor:            "GetAccessor",
		ast.SymbolFlagsSetAccessor:            "SetAccessor",
		ast.SymbolFlagsSignature:              "Signature",
		ast.SymbolFlagsTypeParameter:          "TypeParameter",
		ast.SymbolFlagsTypeAlias:              "TypeAlias",
		ast.SymbolFlagsExportValue:            "ExportValue",
		ast.SymbolFlagsAlias:                  "Alias",
		ast.SymbolFlagsPrototype:              "Prototype",
		ast.SymbolFlagsExportStar:             "ExportStar",
		ast.SymbolFlagsOptional:               "Optional",
		ast.SymbolFlagsTransient:              "Transient",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		names = append(names, "None")
	}

	return names
}

// getTypeFlagNames converts type flags to a slice of flag names
func getTypeFlagNames(flags checker.TypeFlags) []string {
	var names []string

	flagMap := map[checker.TypeFlags]string{
		checker.TypeFlagsAny:             "Any",
		checker.TypeFlagsUnknown:         "Unknown",
		checker.TypeFlagsString:          "String",
		checker.TypeFlagsNumber:          "Number",
		checker.TypeFlagsBoolean:         "Boolean",
		checker.TypeFlagsEnum:            "Enum",
		checker.TypeFlagsBigInt:          "BigInt",
		checker.TypeFlagsStringLiteral:   "StringLiteral",
		checker.TypeFlagsNumberLiteral:   "NumberLiteral",
		checker.TypeFlagsBooleanLiteral:  "BooleanLiteral",
		checker.TypeFlagsEnumLiteral:     "EnumLiteral",
		checker.TypeFlagsBigIntLiteral:   "BigIntLiteral",
		checker.TypeFlagsESSymbol:        "ESSymbol",
		checker.TypeFlagsUniqueESSymbol:  "UniqueESSymbol",
		checker.TypeFlagsVoid:            "Void",
		checker.TypeFlagsUndefined:       "Undefined",
		checker.TypeFlagsNull:            "Null",
		checker.TypeFlagsNever:           "Never",
		checker.TypeFlagsTypeParameter:   "TypeParameter",
		checker.TypeFlagsObject:          "Object",
		checker.TypeFlagsUnion:           "Union",
		checker.TypeFlagsIntersection:    "Intersection",
		checker.TypeFlagsIndex:           "Index",
		checker.TypeFlagsIndexedAccess:   "IndexedAccess",
		checker.TypeFlagsConditional:     "Conditional",
		checker.TypeFlagsSubstitution:    "Substitution",
		checker.TypeFlagsNonPrimitive:    "NonPrimitive",
		checker.TypeFlagsTemplateLiteral: "TemplateLiteral",
		checker.TypeFlagsStringMapping:   "StringMapping",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		names = append(names, "None")
	}

	return names
}

// getObjectFlagNames converts object flags to a slice of flag names
func getObjectFlagNames(flags checker.ObjectFlags) []string {
	var names []string

	flagMap := map[checker.ObjectFlags]string{
		checker.ObjectFlagsClass:        "Class",
		checker.ObjectFlagsInterface:    "Interface",
		checker.ObjectFlagsReference:    "Reference",
		checker.ObjectFlagsTuple:        "Tuple",
		checker.ObjectFlagsAnonymous:    "Anonymous",
		checker.ObjectFlagsMapped:       "Mapped",
		checker.ObjectFlagsInstantiated: "Instantiated",
		checker.ObjectFlagsObjectLiteral: "ObjectLiteral",
		checker.ObjectFlagsEvolvingArray: "EvolvingArray",
		checker.ObjectFlagsReverseMapped: "ReverseMapped",
		checker.ObjectFlagsJsxAttributes: "JsxAttributes",
		checker.ObjectFlagsJSLiteral:     "JSLiteral",
		checker.ObjectFlagsFreshLiteral:  "FreshLiteral",
		checker.ObjectFlagsArrayLiteral:  "ArrayLiteral",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		names = append(names, "None")
	}

	return names
}

// getCheckFlagNames converts check flags to a slice of flag names
func getCheckFlagNames(flags ast.CheckFlags) []string {
	var names []string

	flagMap := map[ast.CheckFlags]string{
		ast.CheckFlagsInstantiated:       "Instantiated",
		ast.CheckFlagsSyntheticProperty:  "SyntheticProperty",
		ast.CheckFlagsSyntheticMethod:    "SyntheticMethod",
		ast.CheckFlagsReadonly:           "Readonly",
		ast.CheckFlagsReadPartial:        "ReadPartial",
		ast.CheckFlagsWritePartial:       "WritePartial",
		ast.CheckFlagsHasNonUniformType:  "HasNonUniformType",
		ast.CheckFlagsHasLiteralType:     "HasLiteralType",
		ast.CheckFlagsContainsPublic:     "ContainsPublic",
		ast.CheckFlagsContainsProtected:  "ContainsProtected",
		ast.CheckFlagsContainsPrivate:    "ContainsPrivate",
		ast.CheckFlagsContainsStatic:     "ContainsStatic",
		ast.CheckFlagsLate:               "Late",
		ast.CheckFlagsReverseMapped:      "ReverseMapped",
		ast.CheckFlagsOptionalParameter:  "OptionalParameter",
		ast.CheckFlagsRestParameter:      "RestParameter",
		ast.CheckFlagsDeferredType:       "DeferredType",
		ast.CheckFlagsHasNeverType:       "HasNeverType",
		ast.CheckFlagsMapped:             "Mapped",
		ast.CheckFlagsStripOptional:      "StripOptional",
		ast.CheckFlagsUnresolved:         "Unresolved",
		ast.CheckFlagsSynthetic:          "Synthetic",
		ast.CheckFlagsIsDiscriminantComputed: "IsDiscriminantComputed",
		ast.CheckFlagsIsDiscriminant:     "IsDiscriminant",
		ast.CheckFlagsPartial:            "Partial",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		names = append(names, "None")
	}

	return names
}

// getFlowFlagNames converts flow flags to a slice of flag names
func getFlowFlagNames(flags ast.FlowFlags) []string {
	var names []string

	flagMap := map[ast.FlowFlags]string{
		ast.FlowFlagsUnreachable:    "Unreachable",
		ast.FlowFlagsStart:          "Start",
		ast.FlowFlagsBranchLabel:    "BranchLabel",
		ast.FlowFlagsLoopLabel:      "LoopLabel",
		ast.FlowFlagsAssignment:     "Assignment",
		ast.FlowFlagsTrueCondition:  "TrueCondition",
		ast.FlowFlagsFalseCondition: "FalseCondition",
		ast.FlowFlagsSwitchClause:   "SwitchClause",
		ast.FlowFlagsArrayMutation:  "ArrayMutation",
		ast.FlowFlagsCall:           "Call",
		ast.FlowFlagsReduceLabel:    "ReduceLabel",
		ast.FlowFlagsReferenced:     "Referenced",
		ast.FlowFlagsShared:         "Shared",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		names = append(names, "None")
	}

	return names
}

// formatNodeLocation creates a NodeLocation from file path and node
func formatNodeLocation(filePath string, node *ast.Node) NodeLocation {
	return NodeLocation{
		FilePath: filePath,
		Pos:      node.Pos(),
		Kind:     int(node.Kind),
	}
}

// TypeInfoCollector recursively collects type information while tracking visited objects
type TypeInfoCollector struct {
	checker        *checker.Checker
	filePath       string
	types          map[uint32]TypeDetails
	symbols        map[uint64]SymbolDetails
	visitedTypes   map[uint32]bool
	visitedSymbols map[uint64]bool
	// Use pointer-based maps for cycle detection (no IDs exposed)
	visitedFlowNodes map[*ast.FlowNode]*FlowNodeDetails
}

// newTypeInfoCollector creates a new TypeInfoCollector
func newTypeInfoCollector(c *checker.Checker, filePath string) *TypeInfoCollector {
	return &TypeInfoCollector{
		checker:          c,
		filePath:         filePath,
		types:            make(map[uint32]TypeDetails),
		symbols:          make(map[uint64]SymbolDetails),
		visitedTypes:     make(map[uint32]bool),
		visitedSymbols:   make(map[uint64]bool),
		visitedFlowNodes: make(map[*ast.FlowNode]*FlowNodeDetails),
	}
}

// collectType recursively collects Type information
func (c *TypeInfoCollector) collectType(t *checker.Type) uint32 {
	if t == nil {
		return 0
	}

	typeId := uint32(t.Id())

	// Check if already visited to prevent cycles
	if c.visitedTypes[typeId] {
		return typeId
	}
	c.visitedTypes[typeId] = true

	flags := t.Flags()
	objectFlags := t.ObjectFlags()

	details := TypeDetails{
		Id:              typeId,
		Flags:           uint32(flags),
		FlagNames:       getTypeFlagNames(flags),
		ObjectFlags:     uint32(objectFlags),
		ObjectFlagNames: getObjectFlagNames(objectFlags),
		TypeString:      c.checker.TypeToString(t),
	}

	// Collect Symbol
	symbol := t.Symbol()
	if symbol != nil {
		symbolId := c.collectSymbol(symbol)
		details.Symbol = &symbolId
	}

	// Collect Target (for TypeReference types)
	if objectFlags&checker.ObjectFlagsReference != 0 {
		objType := t.AsObjectType()
		if objType != nil && objType.Target() != nil {
			targetId := c.collectType(objType.Target())
			details.Target = &targetId
		}
	}

	// Collect Types (for Union/Intersection)
	if flags&checker.TypeFlagsUnionOrIntersection != 0 {
		types := t.Types()
		if types != nil {
			typeIds := make([]uint32, 0, len(types))
			for _, subType := range types {
				typeIds = append(typeIds, c.collectType(subType))
			}
			details.Types = typeIds
		}
	}

	// Collect IntrinsicName (for IntrinsicType)
	if flags&checker.TypeFlagsIntrinsic != 0 {
		// Check for intrinsic types (any, unknown, string, number, boolean, void, undefined, null, never, etc.)
		intrinsicType := t.AsIntrinsicType()
		if intrinsicType != nil {
			details.IntrinsicName = intrinsicType.IntrinsicName()
		}
	}

	// Collect Value (for LiteralType)
	if flags&checker.TypeFlagsLiteral != 0 {
		literalType := t.AsLiteralType()
		if literalType != nil {
			details.Value = literalType.Value()
		}
	}

	// Collect TypeParameters (for InterfaceType)
	if objectFlags&checker.ObjectFlagsClassOrInterface != 0 {
		interfaceType := t.AsInterfaceType()
		if interfaceType != nil {
			typeParams := interfaceType.TypeParameters()
			if len(typeParams) > 0 {
				paramIds := make([]uint32, 0, len(typeParams))
				for _, tp := range typeParams {
					paramIds = append(paramIds, c.collectType(tp))
				}
				details.TypeParameters = paramIds
			}
		}
	}

	// Collect FixedLength and ElementInfos (for TupleType)
	if checker.IsTupleType(t) {
		tupleType := t.AsTupleType()
		if tupleType != nil {
			fixedLen := tupleType.FixedLength()
			details.FixedLength = &fixedLen
			elementInfos := tupleType.ElementInfos()
			if len(elementInfos) > 0 {
				elemDetails := make([]ElementInfoDetail, 0, len(elementInfos))
				for _, ei := range elementInfos {
					elemDetails = append(elemDetails, ElementInfoDetail{
						Flags: uint32(ei.TupleElementFlags()),
					})
				}
				details.ElementInfos = elemDetails
			}
		}
	}

	// Collect Properties, CallSignatures, ConstructSignatures (for StructuredType)
	if flags&checker.TypeFlagsStructuredType != 0 {
		structuredType := t.AsStructuredType()
		if structuredType != nil {
			// Properties
			props := structuredType.Properties()
			if len(props) > 0 {
				propIds := make([]uint64, 0, len(props))
				for _, prop := range props {
					propIds = append(propIds, c.collectSymbol(prop))
				}
				details.Properties = propIds
			}

			// CallSignatures - store SignatureString directly (no IDs for Signature)
			callSigs := structuredType.CallSignatures()
			if len(callSigs) > 0 {
				sigStrings := make([]string, 0, len(callSigs))
				for _, sig := range callSigs {
					sigStrings = append(sigStrings, c.checker.SignatureToStringEx(sig, nil, 0))
				}
				details.CallSignatures = sigStrings
			}

			// ConstructSignatures - store SignatureString directly (no IDs for Signature)
			constructSigs := structuredType.ConstructSignatures()
			if len(constructSigs) > 0 {
				sigStrings := make([]string, 0, len(constructSigs))
				for _, sig := range constructSigs {
					sigStrings = append(sigStrings, c.checker.SignatureToStringEx(sig, nil, 0))
				}
				details.ConstructSignatures = sigStrings
			}
		}
	}

	c.types[typeId] = details
	return typeId
}

// collectSymbol collects Symbol information
func (c *TypeInfoCollector) collectSymbol(symbol *ast.Symbol) uint64 {
	if symbol == nil {
		return 0
	}

	symbolId := uint64(ast.GetSymbolId(symbol))

	// Check if already visited to prevent cycles
	if c.visitedSymbols[symbolId] {
		return symbolId
	}
	c.visitedSymbols[symbolId] = true

	details := SymbolDetails{
		Id:             symbolId,
		Flags:          uint32(symbol.Flags),
		FlagNames:      getSymbolFlagNames(symbol.Flags),
		CheckFlags:     uint32(symbol.CheckFlags),
		CheckFlagNames: getCheckFlagNames(symbol.CheckFlags),
		Name:           symbol.Name,
		SymbolString:   c.checker.SymbolToString(symbol),
	}

	// Collect Declarations
	if len(symbol.Declarations) > 0 {
		decls := make([]NodeLocation, 0, len(symbol.Declarations))
		for _, decl := range symbol.Declarations {
			decls = append(decls, formatNodeLocation(c.filePath, decl))
		}
		details.Declarations = decls
	}

	// ValueDeclaration
	if symbol.ValueDeclaration != nil {
		loc := formatNodeLocation(c.filePath, symbol.ValueDeclaration)
		details.ValueDeclaration = &loc
	}

	// Members (recursively collect)
	if len(symbol.Members) > 0 {
		members := make(map[string]uint64)
		for name, member := range symbol.Members {
			members[name] = c.collectSymbol(member)
		}
		details.Members = members
	}

	// Exports (recursively collect)
	if len(symbol.Exports) > 0 {
		exports := make(map[string]uint64)
		for name, export := range symbol.Exports {
			exports[name] = c.collectSymbol(export)
		}
		details.Exports = exports
	}

	// Parent (ID only, don't recursively collect to avoid large trees)
	if symbol.Parent != nil {
		parentId := uint64(ast.GetSymbolId(symbol.Parent))
		details.Parent = &parentId
	}

	c.symbols[symbolId] = details
	return symbolId
}

// collectSignature collects Signature information
// Note: Signature has no internal ID in typescript-go, returns details directly
func (c *TypeInfoCollector) collectSignature(sig *checker.Signature) *SignatureDetails {
	if sig == nil {
		return nil
	}

	details := &SignatureDetails{
		SignatureString:  c.checker.SignatureToStringEx(sig, nil, 0),
		HasRestParameter: sig.HasRestParameter(),
	}

	// TypeParameters
	typeParams := sig.TypeParameters()
	if len(typeParams) > 0 {
		paramIds := make([]uint32, 0, len(typeParams))
		for _, tp := range typeParams {
			paramIds = append(paramIds, c.collectType(tp))
		}
		details.TypeParameters = paramIds
	}

	// Parameters
	params := checker.Signature_parameters(sig)
	if len(params) > 0 {
		paramDetails := make([]ParameterDetail, 0, len(params))
		for _, param := range params {
			paramDetails = append(paramDetails, ParameterDetail{
				Name:     param.Name,
				SymbolId: c.collectSymbol(param),
			})
		}
		details.Parameters = paramDetails
	}

	// ThisParameter
	thisParam := sig.ThisParameter()
	if thisParam != nil {
		details.ThisParameter = &ParameterDetail{
			Name:     thisParam.Name,
			SymbolId: c.collectSymbol(thisParam),
		}
	}

	// ReturnType
	returnType := c.checker.GetReturnTypeOfSignature(sig)
	if returnType != nil {
		returnTypeId := c.collectType(returnType)
		details.ReturnType = &returnTypeId
	}

	// Declaration
	decl := checker.Signature_declaration(sig)
	if decl != nil {
		loc := formatNodeLocation(c.filePath, decl)
		details.Declaration = &loc
	}

	return details
}

// collectFlowNode collects FlowNode information
// Note: FlowNode has no internal ID in typescript-go, returns details directly
// Uses pointer-based map for cycle detection
func (c *TypeInfoCollector) collectFlowNode(flow *ast.FlowNode) *FlowNodeDetails {
	if flow == nil {
		return nil
	}

	// Check if already visited to prevent cycles
	if existing, exists := c.visitedFlowNodes[flow]; exists {
		return existing
	}

	// Create details and register immediately for cycle detection
	details := &FlowNodeDetails{
		Flags:     uint32(flow.Flags),
		FlagNames: getFlowFlagNames(flow.Flags),
	}
	c.visitedFlowNodes[flow] = details

	// Node
	if flow.Node != nil {
		loc := formatNodeLocation(c.filePath, flow.Node)
		details.Node = &loc
	}

	// Antecedent (single)
	if flow.Antecedent != nil {
		details.Antecedent = c.collectFlowNode(flow.Antecedent)
	}

	// Antecedents (list)
	if flow.Antecedents != nil {
		var antecedents []*FlowNodeDetails
		for list := flow.Antecedents; list != nil; list = list.Next {
			if list.Flow != nil {
				antecedents = append(antecedents, c.collectFlowNode(list.Flow))
			}
		}
		if len(antecedents) > 0 {
			details.Antecedents = antecedents
		}
	}

	return details
}

// findNodeAtPosition finds the node at position with matching kind
func findNodeAtPosition(sourceFile *ast.SourceFile, position int, kind int) *ast.Node {
	var result *ast.Node

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Pos() <= position && position < node.End() {
			// If kind matches and position matches, this is our target
			if node.Pos() == position && int(node.Kind) == kind {
				result = node
				return true // Stop searching
			}
			// Otherwise continue searching children
			node.ForEachChild(visit)
		}
		return false
	}

	sourceFile.AsNode().ForEachChild(visit)
	return result
}

// getNodeAndChecker is a helper that finds the node and checker for a given NodeLocation
func (tc *TypeChecker) getNodeAndChecker(loc NodeLocation) (*ast.Node, *checker.Checker, string) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	sourceFile, ok := tc.sourceFiles[loc.FilePath]
	if !ok {
		return nil, nil, loc.FilePath
	}

	c, ok := tc.checkers[loc.FilePath]
	if !ok {
		return nil, nil, loc.FilePath
	}

	// Find node at position matching the kind
	node := findNodeAtPosition(sourceFile, loc.Pos, loc.Kind)
	if node == nil {
		// Fallback to token at position
		node = findTokenAtPosition(sourceFile, loc.Pos)
	}

	return node, c, loc.FilePath
}

// GetNodeType returns type information for a node (lazy loading)
func (tc *TypeChecker) GetNodeType(loc NodeLocation) *NodeTypeResponse {
	node, c, filePath := tc.getNodeAndChecker(loc)
	if node == nil || c == nil {
		return nil
	}

	response := &NodeTypeResponse{}
	collector := newTypeInfoCollector(c, filePath)

	// Get Type
	t := c.GetTypeAtLocation(node)
	if t != nil {
		collector.collectType(t)
		if details, ok := collector.types[uint32(t.Id())]; ok {
			response.Type = &details
		}
	}

	// Get ContextualType (only for expressions)
	if ast.IsExpression(node) {
		contextualType := checker.Checker_getContextualType(c, node, 0)
		if contextualType != nil {
			collector.collectType(contextualType)
			if details, ok := collector.types[uint32(contextualType.Id())]; ok {
				response.ContextualType = &details
			}
		}
	}

	// Include all collected related types and symbols for reference lookup
	if len(collector.types) > 0 {
		response.RelatedTypes = collector.types
	}
	if len(collector.symbols) > 0 {
		response.RelatedSymbols = collector.symbols
	}

	return response
}

// GetNodeSymbol returns symbol information for a node (lazy loading)
func (tc *TypeChecker) GetNodeSymbol(loc NodeLocation) *NodeSymbolResponse {
	node, c, filePath := tc.getNodeAndChecker(loc)
	if node == nil || c == nil {
		return nil
	}

	response := &NodeSymbolResponse{}
	collector := newTypeInfoCollector(c, filePath)

	// Get Symbol
	symbol := c.GetSymbolAtLocation(node)
	if symbol != nil {
		symbolId := collector.collectSymbol(symbol)
		if details, ok := collector.symbols[symbolId]; ok {
			response.Symbol = &details
		}
	}

	// Include all collected related types and symbols for reference lookup
	if len(collector.types) > 0 {
		response.RelatedTypes = collector.types
	}
	if len(collector.symbols) > 0 {
		response.RelatedSymbols = collector.symbols
	}

	return response
}

// GetNodeSignature returns signature information for a node (lazy loading)
func (tc *TypeChecker) GetNodeSignature(loc NodeLocation) *NodeSignatureResponse {
	node, c, filePath := tc.getNodeAndChecker(loc)
	if node == nil || c == nil {
		return nil
	}

	response := &NodeSignatureResponse{}
	collector := newTypeInfoCollector(c, filePath)

	// Get Signature (only for call-like expressions)
	if ast.IsCallLikeExpression(node) {
		sig := c.GetResolvedSignature(node)
		if sig != nil {
			response.Signature = collector.collectSignature(sig)
		}
	}

	// Include all collected related types and symbols for reference lookup
	if len(collector.types) > 0 {
		response.RelatedTypes = collector.types
	}
	if len(collector.symbols) > 0 {
		response.RelatedSymbols = collector.symbols
	}

	return response
}

// GetNodeFlowNode returns flow node information for a node (lazy loading)
func (tc *TypeChecker) GetNodeFlowNode(loc NodeLocation) *NodeFlowNodeResponse {
	node, c, filePath := tc.getNodeAndChecker(loc)
	if node == nil || c == nil {
		return nil
	}

	response := &NodeFlowNodeResponse{}

	// Get FlowNode
	flowNodeData := node.FlowNodeData()
	if flowNodeData != nil && flowNodeData.FlowNode != nil {
		collector := newTypeInfoCollector(c, filePath)
		response.FlowNode = collector.collectFlowNode(flowNodeData.FlowNode)
	}

	return response
}

// GetNodeInfo returns basic node information (Kind, Flags, ModifierFlags, Pos, End)
func (tc *TypeChecker) GetNodeInfo(loc NodeLocation) *NodeInfoResponse {
	node, _, _ := tc.getNodeAndChecker(loc)
	if node == nil {
		return nil
	}

	return &NodeInfoResponse{
		Kind:              int(node.Kind),
		KindName:          node.Kind.String(),
		Flags:             uint32(node.Flags),
		FlagNames:         getNodeFlagNames(node.Flags),
		ModifierFlags:     uint32(node.ModifierFlags()),
		ModifierFlagNames: getModifierFlagNames(node.ModifierFlags()),
		Pos:               node.Pos(),
		End:               node.End(),
	}
}

// getNodeFlagNames converts node flags to a slice of flag names
func getNodeFlagNames(flags ast.NodeFlags) []string {
	var names []string

	flagMap := map[ast.NodeFlags]string{
		ast.NodeFlagsLet:                             "Let",
		ast.NodeFlagsConst:                           "Const",
		ast.NodeFlagsUsing:                           "Using",
		ast.NodeFlagsReparsed:                        "Reparsed",
		ast.NodeFlagsSynthesized:                     "Synthesized",
		ast.NodeFlagsOptionalChain:                   "OptionalChain",
		ast.NodeFlagsExportContext:                   "ExportContext",
		ast.NodeFlagsContainsThis:                    "ContainsThis",
		ast.NodeFlagsHasImplicitReturn:               "HasImplicitReturn",
		ast.NodeFlagsHasExplicitReturn:               "HasExplicitReturn",
		ast.NodeFlagsDisallowInContext:               "DisallowInContext",
		ast.NodeFlagsYieldContext:                    "YieldContext",
		ast.NodeFlagsDecoratorContext:                "DecoratorContext",
		ast.NodeFlagsAwaitContext:                    "AwaitContext",
		ast.NodeFlagsDisallowConditionalTypesContext: "DisallowConditionalTypesContext",
		ast.NodeFlagsThisNodeHasError:                "ThisNodeHasError",
		ast.NodeFlagsJavaScriptFile:                  "JavaScriptFile",
		ast.NodeFlagsThisNodeOrAnySubNodesHasError:   "ThisNodeOrAnySubNodesHasError",
		ast.NodeFlagsHasAggregatedChildData:          "HasAggregatedChildData",
		ast.NodeFlagsPossiblyContainsDynamicImport:   "PossiblyContainsDynamicImport",
		ast.NodeFlagsPossiblyContainsImportMeta:      "PossiblyContainsImportMeta",
		ast.NodeFlagsHasJSDoc:                        "HasJSDoc",
		ast.NodeFlagsJSDoc:                           "JSDoc",
		ast.NodeFlagsAmbient:                         "Ambient",
		ast.NodeFlagsInWithStatement:                 "InWithStatement",
		ast.NodeFlagsJsonFile:                        "JsonFile",
		ast.NodeFlagsDeprecated:                      "Deprecated",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		names = append(names, "None")
	}

	return names
}

// getModifierFlagNames converts modifier flags to a slice of flag names
func getModifierFlagNames(flags ast.ModifierFlags) []string {
	var names []string

	flagMap := map[ast.ModifierFlags]string{
		ast.ModifierFlagsPublic:     "Public",
		ast.ModifierFlagsPrivate:    "Private",
		ast.ModifierFlagsProtected:  "Protected",
		ast.ModifierFlagsReadonly:   "Readonly",
		ast.ModifierFlagsOverride:   "Override",
		ast.ModifierFlagsExport:     "Export",
		ast.ModifierFlagsAbstract:   "Abstract",
		ast.ModifierFlagsAmbient:    "Ambient",
		ast.ModifierFlagsStatic:     "Static",
		ast.ModifierFlagsAccessor:   "Accessor",
		ast.ModifierFlagsAsync:      "Async",
		ast.ModifierFlagsDefault:    "Default",
		ast.ModifierFlagsConst:      "Const",
		ast.ModifierFlagsIn:         "In",
		ast.ModifierFlagsOut:        "Out",
		ast.ModifierFlagsDecorator:  "Decorator",
		ast.ModifierFlagsDeprecated: "Deprecated",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		names = append(names, "None")
	}

	return names
}
