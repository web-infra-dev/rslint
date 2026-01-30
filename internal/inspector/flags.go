package inspector

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// ============================================================================
// Flag name helpers - Convert flag values to human-readable strings
// ============================================================================

// GetNodeFlagNames converts NodeFlags to a slice of flag names with package prefix
func GetNodeFlagNames(flags ast.NodeFlags) []string {
	var names []string

	// Individual bit flags only (not combined flags)
	flagMap := map[ast.NodeFlags]string{
		ast.NodeFlagsLet:                             "ast.NodeFlagsLet",
		ast.NodeFlagsConst:                           "ast.NodeFlagsConst",
		ast.NodeFlagsUsing:                           "ast.NodeFlagsUsing",
		ast.NodeFlagsReparsed:                        "ast.NodeFlagsReparsed",
		ast.NodeFlagsSynthesized:                     "ast.NodeFlagsSynthesized",
		ast.NodeFlagsOptionalChain:                   "ast.NodeFlagsOptionalChain",
		ast.NodeFlagsExportContext:                   "ast.NodeFlagsExportContext",
		ast.NodeFlagsContainsThis:                    "ast.NodeFlagsContainsThis",
		ast.NodeFlagsHasImplicitReturn:               "ast.NodeFlagsHasImplicitReturn",
		ast.NodeFlagsHasExplicitReturn:               "ast.NodeFlagsHasExplicitReturn",
		ast.NodeFlagsDisallowInContext:               "ast.NodeFlagsDisallowInContext",
		ast.NodeFlagsYieldContext:                    "ast.NodeFlagsYieldContext",
		ast.NodeFlagsDecoratorContext:                "ast.NodeFlagsDecoratorContext",
		ast.NodeFlagsAwaitContext:                    "ast.NodeFlagsAwaitContext",
		ast.NodeFlagsDisallowConditionalTypesContext: "ast.NodeFlagsDisallowConditionalTypesContext",
		ast.NodeFlagsThisNodeHasError:                "ast.NodeFlagsThisNodeHasError",
		ast.NodeFlagsJavaScriptFile:                  "ast.NodeFlagsJavaScriptFile",
		ast.NodeFlagsThisNodeOrAnySubNodesHasError:   "ast.NodeFlagsThisNodeOrAnySubNodesHasError",
		ast.NodeFlagsHasAggregatedChildData:          "ast.NodeFlagsHasAggregatedChildData",
		ast.NodeFlagsPossiblyContainsDynamicImport:   "ast.NodeFlagsPossiblyContainsDynamicImport",
		ast.NodeFlagsPossiblyContainsImportMeta:      "ast.NodeFlagsPossiblyContainsImportMeta",
		ast.NodeFlagsHasJSDoc:                        "ast.NodeFlagsHasJSDoc",
		ast.NodeFlagsJSDoc:                           "ast.NodeFlagsJSDoc",
		ast.NodeFlagsAmbient:                         "ast.NodeFlagsAmbient",
		ast.NodeFlagsInWithStatement:                 "ast.NodeFlagsInWithStatement",
		ast.NodeFlagsJsonFile:                        "ast.NodeFlagsJsonFile",
		ast.NodeFlagsDeprecated:                      "ast.NodeFlagsDeprecated",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	return names
}

// GetTypeFlagNames converts TypeFlags to a slice of flag names with package prefix
func GetTypeFlagNames(flags checker.TypeFlags) []string {
	var names []string

	// Individual bit flags only (not combined flags)
	flagMap := map[checker.TypeFlags]string{
		checker.TypeFlagsAny:             "checker.TypeFlagsAny",
		checker.TypeFlagsUnknown:         "checker.TypeFlagsUnknown",
		checker.TypeFlagsUndefined:       "checker.TypeFlagsUndefined",
		checker.TypeFlagsNull:            "checker.TypeFlagsNull",
		checker.TypeFlagsVoid:            "checker.TypeFlagsVoid",
		checker.TypeFlagsString:          "checker.TypeFlagsString",
		checker.TypeFlagsNumber:          "checker.TypeFlagsNumber",
		checker.TypeFlagsBigInt:          "checker.TypeFlagsBigInt",
		checker.TypeFlagsBoolean:         "checker.TypeFlagsBoolean",
		checker.TypeFlagsESSymbol:        "checker.TypeFlagsESSymbol",
		checker.TypeFlagsStringLiteral:   "checker.TypeFlagsStringLiteral",
		checker.TypeFlagsNumberLiteral:   "checker.TypeFlagsNumberLiteral",
		checker.TypeFlagsBigIntLiteral:   "checker.TypeFlagsBigIntLiteral",
		checker.TypeFlagsBooleanLiteral:  "checker.TypeFlagsBooleanLiteral",
		checker.TypeFlagsUniqueESSymbol:  "checker.TypeFlagsUniqueESSymbol",
		checker.TypeFlagsEnumLiteral:     "checker.TypeFlagsEnumLiteral",
		checker.TypeFlagsEnum:            "checker.TypeFlagsEnum",
		checker.TypeFlagsNonPrimitive:    "checker.TypeFlagsNonPrimitive",
		checker.TypeFlagsNever:           "checker.TypeFlagsNever",
		checker.TypeFlagsTypeParameter:   "checker.TypeFlagsTypeParameter",
		checker.TypeFlagsObject:          "checker.TypeFlagsObject",
		checker.TypeFlagsIndex:           "checker.TypeFlagsIndex",
		checker.TypeFlagsTemplateLiteral: "checker.TypeFlagsTemplateLiteral",
		checker.TypeFlagsStringMapping:   "checker.TypeFlagsStringMapping",
		checker.TypeFlagsSubstitution:    "checker.TypeFlagsSubstitution",
		checker.TypeFlagsIndexedAccess:   "checker.TypeFlagsIndexedAccess",
		checker.TypeFlagsConditional:     "checker.TypeFlagsConditional",
		checker.TypeFlagsUnion:           "checker.TypeFlagsUnion",
		checker.TypeFlagsIntersection:    "checker.TypeFlagsIntersection",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	return names
}

// GetObjectFlagNames converts ObjectFlags to a slice of flag names with package prefix
func GetObjectFlagNames(flags checker.ObjectFlags) []string {
	var names []string

	// Individual bit flags only (not combined flags)
	flagMap := map[checker.ObjectFlags]string{
		checker.ObjectFlagsClass:                                      "checker.ObjectFlagsClass",
		checker.ObjectFlagsInterface:                                  "checker.ObjectFlagsInterface",
		checker.ObjectFlagsReference:                                  "checker.ObjectFlagsReference",
		checker.ObjectFlagsTuple:                                      "checker.ObjectFlagsTuple",
		checker.ObjectFlagsAnonymous:                                  "checker.ObjectFlagsAnonymous",
		checker.ObjectFlagsMapped:                                     "checker.ObjectFlagsMapped",
		checker.ObjectFlagsInstantiated:                               "checker.ObjectFlagsInstantiated",
		checker.ObjectFlagsObjectLiteral:                              "checker.ObjectFlagsObjectLiteral",
		checker.ObjectFlagsEvolvingArray:                              "checker.ObjectFlagsEvolvingArray",
		checker.ObjectFlagsObjectLiteralPatternWithComputedProperties: "checker.ObjectFlagsObjectLiteralPatternWithComputedProperties",
		checker.ObjectFlagsReverseMapped:                              "checker.ObjectFlagsReverseMapped",
		checker.ObjectFlagsJsxAttributes:                              "checker.ObjectFlagsJsxAttributes",
		checker.ObjectFlagsJSLiteral:                                  "checker.ObjectFlagsJSLiteral",
		checker.ObjectFlagsFreshLiteral:                               "checker.ObjectFlagsFreshLiteral",
		checker.ObjectFlagsArrayLiteral:                               "checker.ObjectFlagsArrayLiteral",
		checker.ObjectFlagsPrimitiveUnion:                             "checker.ObjectFlagsPrimitiveUnion",
		checker.ObjectFlagsContainsWideningType:                       "checker.ObjectFlagsContainsWideningType",
		checker.ObjectFlagsContainsObjectOrArrayLiteral:               "checker.ObjectFlagsContainsObjectOrArrayLiteral",
		checker.ObjectFlagsNonInferrableType:                          "checker.ObjectFlagsNonInferrableType",
		checker.ObjectFlagsCouldContainTypeVariablesComputed:          "checker.ObjectFlagsCouldContainTypeVariablesComputed",
		checker.ObjectFlagsCouldContainTypeVariables:                  "checker.ObjectFlagsCouldContainTypeVariables",
		checker.ObjectFlagsMembersResolved:                            "checker.ObjectFlagsMembersResolved",
		checker.ObjectFlagsContainsSpread:                             "checker.ObjectFlagsContainsSpread",
		checker.ObjectFlagsObjectRestType:                             "checker.ObjectFlagsObjectRestType",
		checker.ObjectFlagsInstantiationExpressionType:                "checker.ObjectFlagsInstantiationExpressionType",
		checker.ObjectFlagsSingleSignatureType:                        "checker.ObjectFlagsSingleSignatureType",
		checker.ObjectFlagsIsClassInstanceClone:                       "checker.ObjectFlagsIsClassInstanceClone",
		checker.ObjectFlagsIdenticalBaseTypeCalculated:                "checker.ObjectFlagsIdenticalBaseTypeCalculated",
		checker.ObjectFlagsIdenticalBaseTypeExists:                    "checker.ObjectFlagsIdenticalBaseTypeExists",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	return names
}

// GetSymbolFlagNames converts SymbolFlags to a slice of flag names with package prefix
func GetSymbolFlagNames(flags ast.SymbolFlags) []string {
	var names []string

	// Individual bit flags only (not combined flags)
	flagMap := map[ast.SymbolFlags]string{
		ast.SymbolFlagsFunctionScopedVariable: "ast.SymbolFlagsFunctionScopedVariable",
		ast.SymbolFlagsBlockScopedVariable:    "ast.SymbolFlagsBlockScopedVariable",
		ast.SymbolFlagsProperty:               "ast.SymbolFlagsProperty",
		ast.SymbolFlagsEnumMember:             "ast.SymbolFlagsEnumMember",
		ast.SymbolFlagsFunction:               "ast.SymbolFlagsFunction",
		ast.SymbolFlagsClass:                  "ast.SymbolFlagsClass",
		ast.SymbolFlagsInterface:              "ast.SymbolFlagsInterface",
		ast.SymbolFlagsConstEnum:              "ast.SymbolFlagsConstEnum",
		ast.SymbolFlagsRegularEnum:            "ast.SymbolFlagsRegularEnum",
		ast.SymbolFlagsValueModule:            "ast.SymbolFlagsValueModule",
		ast.SymbolFlagsNamespaceModule:        "ast.SymbolFlagsNamespaceModule",
		ast.SymbolFlagsTypeLiteral:            "ast.SymbolFlagsTypeLiteral",
		ast.SymbolFlagsObjectLiteral:          "ast.SymbolFlagsObjectLiteral",
		ast.SymbolFlagsMethod:                 "ast.SymbolFlagsMethod",
		ast.SymbolFlagsConstructor:            "ast.SymbolFlagsConstructor",
		ast.SymbolFlagsGetAccessor:            "ast.SymbolFlagsGetAccessor",
		ast.SymbolFlagsSetAccessor:            "ast.SymbolFlagsSetAccessor",
		ast.SymbolFlagsSignature:              "ast.SymbolFlagsSignature",
		ast.SymbolFlagsTypeParameter:          "ast.SymbolFlagsTypeParameter",
		ast.SymbolFlagsTypeAlias:              "ast.SymbolFlagsTypeAlias",
		ast.SymbolFlagsExportValue:            "ast.SymbolFlagsExportValue",
		ast.SymbolFlagsAlias:                  "ast.SymbolFlagsAlias",
		ast.SymbolFlagsPrototype:              "ast.SymbolFlagsPrototype",
		ast.SymbolFlagsExportStar:             "ast.SymbolFlagsExportStar",
		ast.SymbolFlagsOptional:               "ast.SymbolFlagsOptional",
		ast.SymbolFlagsTransient:              "ast.SymbolFlagsTransient",
		ast.SymbolFlagsAssignment:             "ast.SymbolFlagsAssignment",
		ast.SymbolFlagsModuleExports:          "ast.SymbolFlagsModuleExports",
		ast.SymbolFlagsConstEnumOnlyModule:    "ast.SymbolFlagsConstEnumOnlyModule",
		ast.SymbolFlagsReplaceableByMethod:    "ast.SymbolFlagsReplaceableByMethod",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	return names
}

// GetCheckFlagNames converts CheckFlags to a slice of flag names with package prefix
func GetCheckFlagNames(flags ast.CheckFlags) []string {
	var names []string

	// Individual bit flags only (not combined flags)
	flagMap := map[ast.CheckFlags]string{
		ast.CheckFlagsInstantiated:           "ast.CheckFlagsInstantiated",
		ast.CheckFlagsSyntheticProperty:      "ast.CheckFlagsSyntheticProperty",
		ast.CheckFlagsSyntheticMethod:        "ast.CheckFlagsSyntheticMethod",
		ast.CheckFlagsReadonly:               "ast.CheckFlagsReadonly",
		ast.CheckFlagsReadPartial:            "ast.CheckFlagsReadPartial",
		ast.CheckFlagsWritePartial:           "ast.CheckFlagsWritePartial",
		ast.CheckFlagsHasNonUniformType:      "ast.CheckFlagsHasNonUniformType",
		ast.CheckFlagsHasLiteralType:         "ast.CheckFlagsHasLiteralType",
		ast.CheckFlagsContainsPublic:         "ast.CheckFlagsContainsPublic",
		ast.CheckFlagsContainsProtected:      "ast.CheckFlagsContainsProtected",
		ast.CheckFlagsContainsPrivate:        "ast.CheckFlagsContainsPrivate",
		ast.CheckFlagsContainsStatic:         "ast.CheckFlagsContainsStatic",
		ast.CheckFlagsLate:                   "ast.CheckFlagsLate",
		ast.CheckFlagsReverseMapped:          "ast.CheckFlagsReverseMapped",
		ast.CheckFlagsOptionalParameter:      "ast.CheckFlagsOptionalParameter",
		ast.CheckFlagsRestParameter:          "ast.CheckFlagsRestParameter",
		ast.CheckFlagsDeferredType:           "ast.CheckFlagsDeferredType",
		ast.CheckFlagsHasNeverType:           "ast.CheckFlagsHasNeverType",
		ast.CheckFlagsMapped:                 "ast.CheckFlagsMapped",
		ast.CheckFlagsStripOptional:          "ast.CheckFlagsStripOptional",
		ast.CheckFlagsUnresolved:             "ast.CheckFlagsUnresolved",
		ast.CheckFlagsIsDiscriminantComputed: "ast.CheckFlagsIsDiscriminantComputed",
		ast.CheckFlagsIsDiscriminant:         "ast.CheckFlagsIsDiscriminant",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	return names
}

// GetFlowFlagNames converts FlowFlags to a slice of flag names with package prefix
// Note: Referenced and Shared are internal bookkeeping flags, not semantic flow types, so we exclude them
func GetFlowFlagNames(flags ast.FlowFlags) []string {
	var names []string

	// Individual bit flags only (not combined flags)
	// Excludes Referenced and Shared as they are internal state flags
	flagMap := map[ast.FlowFlags]string{
		ast.FlowFlagsUnreachable:    "ast.FlowFlagsUnreachable",
		ast.FlowFlagsStart:          "ast.FlowFlagsStart",
		ast.FlowFlagsBranchLabel:    "ast.FlowFlagsBranchLabel",
		ast.FlowFlagsLoopLabel:      "ast.FlowFlagsLoopLabel",
		ast.FlowFlagsAssignment:     "ast.FlowFlagsAssignment",
		ast.FlowFlagsTrueCondition:  "ast.FlowFlagsTrueCondition",
		ast.FlowFlagsFalseCondition: "ast.FlowFlagsFalseCondition",
		ast.FlowFlagsSwitchClause:   "ast.FlowFlagsSwitchClause",
		ast.FlowFlagsArrayMutation:  "ast.FlowFlagsArrayMutation",
		ast.FlowFlagsCall:           "ast.FlowFlagsCall",
		ast.FlowFlagsReduceLabel:    "ast.FlowFlagsReduceLabel",
		// Note: FlowFlagsReferenced and FlowFlagsShared are excluded as they are internal
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	return names
}

// GetSignatureFlagNames converts SignatureFlags to a slice of flag names with package prefix
func GetSignatureFlagNames(flags checker.SignatureFlags) []string {
	var names []string

	// Individual bit flags only (not combined flags)
	flagMap := map[checker.SignatureFlags]string{
		checker.SignatureFlagsHasRestParameter:                    "checker.SignatureFlagsHasRestParameter",
		checker.SignatureFlagsHasLiteralTypes:                     "checker.SignatureFlagsHasLiteralTypes",
		checker.SignatureFlagsConstruct:                           "checker.SignatureFlagsConstruct",
		checker.SignatureFlagsAbstract:                            "checker.SignatureFlagsAbstract",
		checker.SignatureFlagsIsInnerCallChain:                    "checker.SignatureFlagsIsInnerCallChain",
		checker.SignatureFlagsIsOuterCallChain:                    "checker.SignatureFlagsIsOuterCallChain",
		checker.SignatureFlagsIsUntypedSignatureInJSFile:          "checker.SignatureFlagsIsUntypedSignatureInJSFile",
		checker.SignatureFlagsIsNonInferrable:                     "checker.SignatureFlagsIsNonInferrable",
		checker.SignatureFlagsIsSignatureCandidateForOverloadFailure: "checker.SignatureFlagsIsSignatureCandidateForOverloadFailure",
	}

	for flag, name := range flagMap {
		if flags&flag != 0 {
			names = append(names, name)
		}
	}

	return names
}

// GetNodeText extracts text content from simple nodes (identifiers and literals)
// For complex expressions, use scanner.GetSourceTextOfNodeFromSourceFile instead
func GetNodeText(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		if id := node.AsIdentifier(); id != nil {
			return id.Text
		}
	case ast.KindStringLiteral:
		if str := node.AsStringLiteral(); str != nil {
			return str.Text
		}
	case ast.KindNumericLiteral:
		if num := node.AsNumericLiteral(); num != nil {
			return num.Text
		}
	case ast.KindBigIntLiteral:
		if big := node.AsBigIntLiteral(); big != nil {
			return big.Text
		}
	case ast.KindNoSubstitutionTemplateLiteral:
		if tmpl := node.AsNoSubstitutionTemplateLiteral(); tmpl != nil {
			return tmpl.Text
		}
	}

	return ""
}
