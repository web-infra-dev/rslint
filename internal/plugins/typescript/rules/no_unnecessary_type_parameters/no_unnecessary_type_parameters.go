package no_unnecessary_type_parameters

import (
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var replaceUsagesWithConstraintMessage = rule.RuleMessage{
	Id:          "replaceUsagesWithConstraint",
	Description: "Replace all usages of type parameter with its constraint.",
}

func buildSoleMessage(name, descriptor, uses string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "sole",
		Description: "Type parameter " + name + " is " + uses + " in the " + descriptor + " signature.",
		Data: map[string]string{
			"name":       name,
			"descriptor": descriptor,
			"uses":       uses,
		},
	}
}

var NoUnnecessaryTypeParametersRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-type-parameters",
	RequiresTypeInfo: true,
	Schema:           rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkFunctionNode := func(node *ast.Node) { checkNode(ctx, node, "function") }
		checkClassNode := func(node *ast.Node) { checkNode(ctx, node, "class") }

		return rule.RuleListeners{
			// Mirrors upstream's selector groups. TSConstructSignatureDeclaration
			// (interface `new <T>(): T`) is deliberately omitted: upstream's own
			// selector list excludes it too, so it is never checked.
			ast.KindArrowFunction:       checkFunctionNode,
			ast.KindFunctionDeclaration: checkFunctionNode,
			ast.KindFunctionExpression:  checkFunctionNode,
			// tsgo represents a method's value directly on the MethodDeclaration
			// node (no separate FunctionExpression/TSEmptyBodyFunctionExpression
			// wrapper the way ESTree does), so this single kind covers both.
			ast.KindMethodDeclaration: checkFunctionNode,
			ast.KindCallSignature:     checkFunctionNode,
			ast.KindConstructorType:   checkFunctionNode,
			ast.KindFunctionType:      checkFunctionNode,
			ast.KindMethodSignature:   checkFunctionNode,
			ast.KindClassDeclaration:  checkClassNode,
			ast.KindClassExpression:   checkClassNode,
		}
	},
})

func checkNode(ctx rule.RuleContext, node *ast.Node, descriptor string) {
	typeParams := node.TypeParameters()
	if len(typeParams) == 0 {
		return
	}

	// Lazily resolved: only pay for the type-checker walk when at least one
	// type parameter isn't already proven repeated by the cheap AST check.
	var counts map[*ast.Node]int

	for _, typeParamNode := range typeParams {
		if utils.IsJSDocSyntaxNode(typeParamNode) {
			continue
		}
		typeParam := typeParamNode.AsTypeParameterDeclaration()
		if typeParam == nil {
			continue
		}
		nameNode := typeParam.Name()
		if nameNode == nil || !ast.IsIdentifier(nameNode) {
			continue
		}
		sym := typeParamNode.Symbol()
		if sym == nil {
			continue
		}

		if isTypeParameterRepeatedInAST(ctx, typeParamNode, node, sym) {
			continue
		}

		if counts == nil {
			counts = countTypeParameterUsage(
				ctx.TypeChecker,
				node,
				ctx.Program() != nil && needsLegacyObjectSpreadRecovery(ctx.Program().Options()),
			)
		}
		useCount := counts[nameNode]
		if useCount == 0 || useCount > 2 {
			continue
		}

		uses := "used only once"
		if useCount == 1 {
			uses = "never used"
		}

		msg := buildSoleMessage(nameNode.AsIdentifier().Text, descriptor, uses)
		container := node
		ctx.ReportNodeWithDeferredSuggestions(typeParamNode, msg, func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message:  replaceUsagesWithConstraintMessage,
				FixesArr: buildReplaceWithConstraintFixes(ctx, container, typeParamNode, typeParam, sym),
			}}
		})
	}
}

func collectSignatureReturnTypeParameterUsages(tc *checker.Checker, sig *checker.Signature) map[*ast.Node]int {
	if sig == nil {
		return nil
	}
	normalizeMultiplicity := func(usages map[*ast.Node]int) map[*ast.Node]int {
		if returnTypeAssumesMultipleUses(tc, sig) {
			for name, count := range usages {
				if count == 1 {
					usages[name] = 2
				}
			}
		}
		return usages
	}
	declaration := sig.Declaration()
	if declaration == nil {
		return nil
	}
	if returnType := declaration.Type(); returnType != nil {
		usages := make(map[*ast.Node]int)
		collectTypeParameterUsageCounts(tc, returnType, usages, false)
		return normalizeMultiplicity(usages)
	}

	allUsages := make(map[*ast.Node]int)
	collectTypeParameterUsageCounts(tc, declaration, allUsages, false)
	nonReturnUsages := make(map[*ast.Node]int)
	if this := sig.ThisParameter(); this != nil && this.ValueDeclaration != nil {
		collectTypeParameterUsageCounts(tc, this.ValueDeclaration, nonReturnUsages, false)
	}
	for _, parameter := range sig.Parameters() {
		if parameter.ValueDeclaration != nil {
			collectTypeParameterUsageCounts(tc, parameter.ValueDeclaration, nonReturnUsages, false)
		}
	}
	for _, typeParameter := range sig.TypeParameters() {
		if symbol := typeParameter.Symbol(); symbol != nil && len(symbol.Declarations) != 0 {
			collectTypeParameterUsageCounts(tc, symbol.Declarations[0], nonReturnUsages, false)
		}
	}

	returnUsages := make(map[*ast.Node]int)
	for name, count := range allUsages {
		if returnCount := count - nonReturnUsages[name]; returnCount > 0 {
			returnUsages[name] = returnCount
		}
	}
	return normalizeMultiplicity(returnUsages)
}

type genericReturnUsage struct {
	count int
	group *ast.Node
}

// genericCallArgumentReturnUsages returns the target type parameters that an
// argument can infer and that also occur in the target's return type. Counts
// preserve repeated return positions, allowing arguments for the same target
// parameter to be merged as alternatives while distinct parameters add.
func genericCallArgumentReturnUsages(tc *checker.Checker, callNode *ast.Node, argumentIndex int) map[*ast.Node]genericReturnUsage {
	// An explicit type argument fixes the generic result independently of the
	// value argument. If it mentions the enclosing type parameter, the normal
	// return-type walk already sees that occurrence; otherwise the spread is
	// intentionally erased from the signature.
	if len(callNode.TypeArguments()) != 0 {
		return nil
	}
	// These ubiquitous library helpers preserve the argument (or its awaited
	// value) in their result. Recognizing them before asking the checker to
	// instantiate their large declaration graphs keeps the compatibility path
	// cheap on real code and avoids repeatedly materializing Promise/Array APIs.
	callee := callExpressionTarget(callNode)
	if callee != nil {
		callee = ast.SkipParentheses(callee)
	}
	if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
		propertyAccess := callee.AsPropertyAccessExpression()
		if propertyAccess.Name() != nil && ast.IsIdentifier(propertyAccess.Name()) {
			switch propertyAccess.Name().AsIdentifier().Text {
			case "resolve":
				if argumentIndex == 0 && isTypeScriptLibValueNamed(tc, propertyAccess.Expression, "Promise") {
					return map[*ast.Node]genericReturnUsage{propertyAccess.Name(): {count: 2}}
				}
				return nil
			case "map":
				if argumentIndex == 0 && tc.IsArrayType(tc.GetTypeAtLocation(propertyAccess.Expression)) &&
					isTypeScriptLibSymbol(tc, tc.GetSymbolAtLocation(propertyAccess.Name())) {
					return map[*ast.Node]genericReturnUsage{propertyAccess.Name(): {count: 2}}
				}
				return nil
			}
		}
	}
	// For ordinary local generic helpers, walking the declaration is both more
	// precise and much cheaper than resolving the instantiated parameter type.
	if callee != nil && callee.Kind == ast.KindIdentifier {
		if symbol := tc.GetSymbolAtLocation(callee); symbol != nil {
			var declaration *ast.Node
			if len(symbol.Declarations) == 1 && ast.IsFunctionLike(symbol.Declarations[0]) {
				declaration = symbol.Declarations[0]
			}
			if declaration == nil {
				goto resolveSignature
			}
			parameters := declaration.Parameters()
			parameterIndex := argumentIndex
			if parameterIndex >= len(parameters) {
				if len(parameters) == 0 || parameters[len(parameters)-1].AsParameterDeclaration().DotDotDotToken == nil {
					return nil
				}
				parameterIndex = len(parameters) - 1
			}
			if parameterType := parameters[parameterIndex].Type(); parameterType != nil {
				parameterName := directDeclaredTypeParameter(declaration, parameterType)
				if parameterName == nil {
					goto resolveSignature
				}
				returnUsages := collectSignatureReturnTypeParameterUsages(tc, tc.GetSignatureFromDeclaration(declaration))
				if count := returnUsages[parameterName]; count != 0 {
					return map[*ast.Node]genericReturnUsage{parameterName: {
						count: count,
						group: genericReturnSurfaceGroup(declaration, parameterName),
					}}
				}
				return nil
			}
		}
	}

resolveSignature:
	sig := checker.Checker_getResolvedSignature(tc, callNode, nil, checker.CheckModeNormal)
	if sig == nil {
		return nil
	}
	for sig.Target() != nil {
		sig = sig.Target()
	}
	if sig.Declaration() == nil {
		return nil
	}

	parameters := sig.Parameters()
	if len(parameters) == 0 {
		return nil
	}
	parameterIndex := argumentIndex
	if parameterIndex >= len(parameters) {
		if !sig.HasRestParameter() {
			return nil
		}
		parameterIndex = len(parameters) - 1
	}
	parameterDeclaration := parameters[parameterIndex].ValueDeclaration
	if parameterDeclaration == nil {
		return nil
	}

	returnUsages := collectSignatureReturnTypeParameterUsages(tc, sig)
	if len(returnUsages) == 0 {
		return nil
	}
	parameterUsages := make(map[*ast.Node]bool)
	collectInferableTypeParameterUsages(tc, tc.GetTypeOfSymbol(parameters[parameterIndex]), returnUsages, parameterUsages)
	result := make(map[*ast.Node]genericReturnUsage)
	for name := range parameterUsages {
		if count := returnUsages[name]; count != 0 {
			result[name] = genericReturnUsage{
				count: count,
				group: genericReturnSurfaceGroup(sig.Declaration(), name),
			}
		}
	}
	return result
}

func directDeclaredTypeParameter(declaration *ast.Node, typeNode *ast.Node) *ast.Node {
	if declaration == nil || typeNode == nil {
		return nil
	}
	typeNode = ast.SkipTypeParentheses(typeNode)
	switch typeNode.Kind {
	case ast.KindArrayType:
		return directDeclaredTypeParameter(declaration, typeNode.AsArrayTypeNode().ElementType)
	case ast.KindTupleType:
		elements := typeNode.AsTupleTypeNode().Elements.Nodes
		if len(elements) == 1 {
			return directDeclaredTypeParameter(declaration, elements[0])
		}
		return nil
	case ast.KindRestType, ast.KindOptionalType, ast.KindNamedTupleMember:
		return directDeclaredTypeParameter(declaration, typeNode.Type())
	case ast.KindTypeReference:
	default:
		return nil
	}
	typeName := typeNode.AsTypeReferenceNode().TypeName
	if typeName == nil || !ast.IsIdentifier(typeName) {
		return nil
	}
	name := typeName.AsIdentifier().Text
	for _, typeParameter := range declaration.TypeParameters() {
		if typeParameter.Name() != nil && ast.IsIdentifier(typeParameter.Name()) && typeParameter.Name().AsIdentifier().Text == name {
			return typeParameter.Name()
		}
	}
	return nil
}

// Type parameters that occur as sibling constituents of a naked union or
// intersection share one result surface. If two call arguments infer the same
// enclosing generic into those constituents, TypeScript collapses (T & T) or
// (T | T) to one occurrence instead of counting both target parameters.
func genericReturnSurfaceGroup(declaration *ast.Node, name *ast.Node) *ast.Node {
	if declaration == nil || name == nil {
		return name
	}
	returnType := declaration.Type()
	if returnType == nil {
		return name
	}
	returnType = ast.SkipTypeParentheses(returnType)
	if returnType.Kind != ast.KindUnionType && returnType.Kind != ast.KindIntersectionType {
		return name
	}
	var constituents []*ast.Node
	if returnType.Kind == ast.KindUnionType {
		constituents = returnType.AsUnionTypeNode().Types.Nodes
	} else {
		constituents = returnType.AsIntersectionTypeNode().Types.Nodes
	}
	for _, constituent := range constituents {
		if directDeclaredTypeParameter(declaration, constituent) == name {
			return returnType
		}
	}
	return name
}

func callExpressionTarget(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindCallExpression:
		return node.AsCallExpression().Expression
	case ast.KindNewExpression:
		return node.AsNewExpression().Expression
	case ast.KindTaggedTemplateExpression:
		return node.AsTaggedTemplateExpression().Tag
	}
	return nil
}

func isTypeScriptLibValueNamed(tc *checker.Checker, node *ast.Node, name string) bool {
	if node == nil || node.Kind != ast.KindIdentifier || node.AsIdentifier().Text != name {
		return false
	}
	symbol := tc.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = tc.GetAliasedSymbol(symbol)
	}
	return symbol != nil && symbol.Name == name && isTypeScriptLibSymbol(tc, symbol)
}

func isTypeScriptLibSymbol(tc *checker.Checker, symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = tc.GetAliasedSymbol(symbol)
	}
	for _, declaration := range symbol.Declarations {
		if sourceFile := ast.GetSourceFileOfNode(declaration); sourceFile != nil && sourceFile.IsDeclarationFile {
			base := filepath.Base(sourceFile.FileName())
			if strings.HasPrefix(base, "lib.") && strings.HasSuffix(base, ".d.ts") {
				return true
			}
		}
	}
	return false
}

func returnTypeAssumesMultipleUses(tc *checker.Checker, sig *checker.Signature) bool {
	if sig == nil {
		return false
	}
	t := tc.GetReturnTypeOfSignature(sig)
	if predicate := tc.GetTypePredicateOfSignature(sig); predicate != nil && predicate.Type() != nil {
		t = predicate.Type()
	}
	if t == nil {
		return false
	}
	if alias := t.Alias(); alias != nil && alias.TypeArguments() != nil {
		return true
	}
	if !utils.IsTypeReference(t) {
		return false
	}
	target := t.Target()
	if checker.IsTupleType(target) {
		return !target.AsTupleType().IsReadonly()
	}
	if tc.IsArrayType(target) {
		symbol := t.Symbol()
		return symbol != nil && symbol.Name == "Array"
	}
	return true
}

// collectInferableTypeParameterUsages records the generic parameters whose
// values may be inferred from an argument at a given parameter position.
// NoInfer<T> is deliberately opaque: T can occur in the target's return type
// without the argument having any influence on the inferred result.
func collectInferableTypeParameterUsages(tc *checker.Checker, root *checker.Type, wanted map[*ast.Node]int, usages map[*ast.Node]bool) {
	visited := make(map[*checker.Type]bool)
	// The caller only needs to know whether this parameter and the return type
	// share at least one inferable generic; stop at the first match instead of
	// materializing an entire standard-library type graph.
	remaining := len(wanted)
	var visit func(*checker.Type)

	visitSignature := func(sig *checker.Signature) {
		if sig == nil || remaining == 0 {
			return
		}
		if this := sig.ThisParameter(); this != nil {
			visit(tc.GetTypeOfSymbol(this))
		}
		for _, parameter := range sig.Parameters() {
			visit(tc.GetTypeOfSymbol(parameter))
		}
		visit(tc.GetReturnTypeOfSignature(sig))
	}

	visit = func(t *checker.Type) {
		if t == nil || visited[t] || remaining == 0 {
			return
		}
		visited[t] = true

		if utils.IsTypeFlagSet(t, checker.TypeFlagsSubstitution) {
			substitution := t.AsSubstitutionType()
			if constraint := substitution.SubstConstraint(); constraint != nil && utils.IsTypeFlagSet(constraint, checker.TypeFlagsUnknown) {
				return
			}
			visit(substitution.BaseType())
			return
		}
		if utils.IsTypeParameter(t) {
			if symbol := t.Symbol(); symbol != nil && len(symbol.Declarations) != 0 {
				declaration := symbol.Declarations[0]
				if ast.IsTypeParameterDeclaration(declaration) && declaration.Name() != nil && wanted[declaration.Name()] != 0 && !usages[declaration.Name()] {
					usages[declaration.Name()] = true
					remaining--
				}
			}
			return
		}
		switch {
		case utils.IsUnionType(t) || utils.IsIntersectionType(t):
			for _, constituent := range t.Types() {
				visit(constituent)
			}
		case utils.IsTypeFlagSet(t, checker.TypeFlagsIndexedAccess):
			indexed := t.AsIndexedAccessType()
			visit(indexed.ObjectType())
			visit(indexed.IndexType())
		case utils.IsTypeFlagSet(t, checker.TypeFlagsTemplateLiteral):
			for _, part := range t.AsTemplateLiteralType().Types() {
				visit(part)
			}
		case utils.IsTypeFlagSet(t, checker.TypeFlagsConditional):
			conditional := t.AsConditionalType()
			visit(conditional.CheckType())
			visit(conditional.ExtendsType())
		case utils.IsObjectType(t):
			properties := checker.Checker_getPropertiesOfType(tc, t)
			for _, property := range properties {
				visit(tc.GetTypeOfSymbol(property))
			}
			for _, info := range checker.Checker_getIndexInfosOfType(tc, t) {
				visit(info.ValueType())
			}
			if len(properties) == 0 && checker.Type_objectFlags(t)&checker.ObjectFlagsMapped != 0 {
				constraint := checker.Checker_getConstraintTypeFromMappedType(tc, t)
				nameType := checker.Checker_getNameTypeFromMappedType(tc, t)
				if constraint != nil && !utils.IsTypeFlagSet(constraint, checker.TypeFlagsNever) &&
					(nameType == nil || !utils.IsTypeFlagSet(nameType, checker.TypeFlagsNever)) {
					visit(constraint)
					visit(checker.Checker_getTemplateTypeFromMappedType(tc, t))
				}
			}
			for _, sig := range utils.GetCallSignatures(tc, t) {
				visitSignature(sig)
			}
			for _, sig := range utils.GetConstructSignatures(tc, t) {
				visitSignature(sig)
			}
		case utils.IsTypeFlagSet(t, checker.TypeFlagsIndex):
			visit(t.AsIndexType().Target())
		case utils.IsTypeFlagSet(t, checker.TypeFlagsStringMapping):
			visit(t.AsStringMappingType().Target())
		}
	}

	visit(root)
}

// collectLegacyNullableSpreadUsages identifies the narrow inference delta
// between TypeScript 5.9 and tsgo: with omitted strictness options, spreading
// an optional T preserves object-like constituents (including an implicitly
// unconstrained T); tsgo's strict default erases them. Primitive/unknown
// constraints and non-nullable operands agree already and are not recovered.
func collectLegacyNullableSpreadUsages(tc *checker.Checker, expression *ast.Node) map[*ast.Node]int {
	if expression == nil {
		return nil
	}
	expression = ast.SkipParentheses(expression)
	if expression.Kind == ast.KindBinaryExpression && expression.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
		binary := expression.AsBinaryExpression()
		leftUsages := collectLegacyNullableSpreadUsages(tc, binary.Left)
		if len(leftUsages) == 0 {
			return nil
		}
		rightType := tc.GetTypeAtLocation(binary.Right)
		for name := range leftUsages {
			if !typeContainsTypeParameterDeclaration(rightType, name, make(map[*checker.Type]bool)) {
				delete(leftUsages, name)
			}
		}
		return leftUsages
	}
	t := tc.GetTypeAtLocation(expression)
	if t == nil || !utils.IsUnionType(t) {
		return nil
	}
	hasNullish := false
	for _, constituent := range t.Types() {
		if utils.IsTypeFlagSet(constituent, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			hasNullish = true
			break
		}
	}
	if !hasNullish {
		return nil
	}

	var usages map[*ast.Node]int
	for _, constituent := range t.Types() {
		collectLegacySpreadTypeParameterUsages(tc, constituent, false, false, &usages)
	}
	return usages
}

func collectLegacySpreadTypeParameterUsages(tc *checker.Checker, t *checker.Type, objectGuaranteed bool, requireObjectGuarantee bool, usages *map[*ast.Node]int) {
	if t == nil || utils.IsTypeFlagSet(t, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
		return
	}
	if utils.IsTypeParameter(t) {
		symbol := t.Symbol()
		if symbol == nil || len(symbol.Declarations) == 0 || !ast.IsTypeParameterDeclaration(symbol.Declarations[0]) {
			return
		}
		declaration := symbol.Declarations[0].AsTypeParameterDeclaration()
		if declaration.Name() == nil {
			return
		}
		if requireObjectGuarantee && !objectGuaranteed {
			return
		}
		if declaration.Constraint != nil && !isLegacyObjectLikeConstraint(tc.GetConstraintOfTypeParameter(t)) && !objectGuaranteed {
			return
		}
		if *usages == nil {
			*usages = make(map[*ast.Node]int)
		}
		(*usages)[declaration.Name()]++
		return
	}
	if !utils.IsIntersectionType(t) {
		return
	}
	intersectionGuaranteesObject := objectGuaranteed
	if !intersectionGuaranteesObject {
		for _, constituent := range t.Types() {
			if !utils.IsTypeParameter(constituent) && isLegacyObjectLikeConstraint(constituent) {
				intersectionGuaranteesObject = true
				break
			}
		}
	}
	for _, constituent := range t.Types() {
		collectLegacySpreadTypeParameterUsages(tc, constituent, intersectionGuaranteesObject, true, usages)
	}
}

func isLegacyObjectLikeConstraint(t *checker.Type) bool {
	if t == nil {
		return false
	}
	if utils.IsObjectType(t) || utils.IsTypeFlagSet(t, checker.TypeFlagsNonPrimitive) {
		return true
	}
	// `T extends U` is retained by legacy object-spread inference even when U
	// is itself unconstrained. This differs from explicit `unknown`/`any` and
	// primitive constraints, which erase T from the spread result.
	if utils.IsTypeParameter(t) {
		return true
	}
	if utils.IsIntersectionType(t) {
		foundObjectLike := false
		for _, constituent := range t.Types() {
			if !isLegacyObjectLikeConstraint(constituent) {
				return false
			}
			foundObjectLike = true
		}
		return foundObjectLike
	}
	if !utils.IsUnionType(t) {
		return false
	}
	foundObjectLike := false
	for _, constituent := range t.Types() {
		if utils.IsTypeFlagSet(constituent, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}
		if !isLegacyObjectLikeConstraint(constituent) {
			return false
		}
		foundObjectLike = true
	}
	return foundObjectLike
}

func typeContainsTypeParameterDeclaration(t *checker.Type, name *ast.Node, visited map[*checker.Type]bool) bool {
	if t == nil || visited[t] {
		return false
	}
	visited[t] = true
	if utils.IsTypeParameter(t) {
		symbol := t.Symbol()
		return symbol != nil && len(symbol.Declarations) != 0 && ast.IsTypeParameterDeclaration(symbol.Declarations[0]) && symbol.Declarations[0].Name() == name
	}
	if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
		for _, constituent := range t.Types() {
			if typeContainsTypeParameterDeclaration(constituent, name, visited) {
				return true
			}
		}
	}
	return false
}

func isObjectSpreadProperty(property *ast.Symbol) bool {
	if checker.GetDeclarationModifierFlagsFromSymbol(property)&(ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected) != 0 {
		return false
	}
	isClassElement := false
	for _, declaration := range property.Declarations {
		if ast.IsPrivateIdentifierClassElementDeclaration(declaration) {
			return false
		}
		if declaration.Parent != nil && ast.IsClassLike(declaration.Parent) {
			isClassElement = true
		}
	}
	return property.Flags&(ast.SymbolFlagsMethod|ast.SymbolFlagsGetAccessor|ast.SymbolFlagsSetAccessor) == 0 || !isClassElement
}

func computedPropertyAssumesMultipleUses(tc *checker.Checker, property *ast.Node) bool {
	name := property.Name()
	if name == nil || name.Kind != ast.KindComputedPropertyName || property.Parent == nil {
		return false
	}
	for _, info := range checker.Checker_getIndexInfosOfType(tc, tc.GetTypeAtLocation(property.Parent)) {
		if keyType := info.KeyType(); keyType != nil && utils.IsTypeFlagSet(keyType, checker.TypeFlagsString|checker.TypeFlagsNumber) {
			return true
		}
	}
	return false
}

// collectInferredObjectSpreadUsages follows only expression positions that
// contribute to an inferred return (including yielded values and generic
// identity-like calls), then records type parameters carried by object-spread
// operands. It deliberately does not traverse arbitrary subexpressions: an
// object spread passed to a consumer or hidden behind an explicit annotation
// does not become part of the function signature.
func collectInferredObjectSpreadUsages(tc *checker.Checker, node *ast.Node) map[*ast.Node]int {
	if node != nil && ast.IsFunctionLike(node) && node.Type() == nil {
		if body := node.Body(); body != nil && body.Kind != ast.KindBlock {
			if usages, handled := collectSimpleLegacyObjectSpreadUsages(tc, body); handled {
				return usages
			}
		}
	}
	return collectInferredObjectSpreadUsagesWorker(tc, node, nil)
}

func collectInferredObjectSpreadExpressionUsages(tc *checker.Checker, expression *ast.Node) map[*ast.Node]int {
	if usages, handled := collectSimpleLegacyObjectSpreadUsages(tc, expression); handled {
		return usages
	}
	return collectInferredObjectSpreadUsagesWorker(tc, nil, expression)
}

// collectSimpleLegacyObjectSpreadUsages handles the dominant compatibility
// shape without constructing the general return-flow walker: an inferred
// expression-bodied function (or default initializer) whose result is one
// object literal with one direct spread. The general path remains responsible
// for nested spreads, branches, calls, selections, and destructuring.
func collectSimpleLegacyObjectSpreadUsages(tc *checker.Checker, expression *ast.Node) (map[*ast.Node]int, bool) {
	for expression != nil {
		expression = ast.SkipParentheses(expression)
		switch expression.Kind {
		case ast.KindNonNullExpression, ast.KindSatisfiesExpression,
			ast.KindAwaitExpression, ast.KindPartiallyEmittedExpression:
			expression = expression.Expression()
			continue
		case ast.KindAsExpression, ast.KindTypeAssertionExpression:
			if ast.IsConstTypeReference(expression.Type()) {
				expression = expression.Expression()
				continue
			}
			return nil, true
		}
		break
	}
	if expression == nil || expression.Kind != ast.KindObjectLiteralExpression {
		return nil, false
	}

	var spreadExpression *ast.Node
	for _, property := range expression.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind == ast.KindSpreadAssignment {
			if spreadExpression != nil {
				return nil, false
			}
			spreadExpression = property.AsSpreadAssignment().Expression
			continue
		}
		if containsObjectSpreadSyntax(property) {
			return nil, false
		}
	}
	if spreadExpression == nil {
		return nil, true
	}
	if usages := collectLegacyNullableSpreadUsages(tc, spreadExpression); len(usages) != 0 {
		return usages, true
	}
	if !containsObjectSpreadSyntax(spreadExpression) {
		return nil, true
	}
	return nil, false
}

func containsObjectSpreadSyntax(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindSpreadAssignment {
		return true
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if containsObjectSpreadSyntax(child) {
			found = true
			return true
		}
		return false
	})
	return found
}

func objectRestSelectionKey(excluded map[string]bool) string {
	names := make([]string, 0, len(excluded))
	for name := range excluded {
		names = append(names, name)
	}
	slices.Sort(names)
	return "\x00object-rest:" + strings.Join(names, "\x00")
}

func collectInferredObjectSpreadUsagesWorker(tc *checker.Checker, node *ast.Node, rootExpression *ast.Node) map[*ast.Node]int {
	if rootExpression == nil {
		if node == nil || node.Type() != nil || !ast.IsFunctionLike(node) || node.Body() == nil {
			return nil
		}
		if node.SubtreeFacts()&ast.SubtreeContainsESObjectRestOrSpread == 0 {
			return nil
		}
	}

	var usages map[*ast.Node]int
	var activeDeclarations map[*ast.Node]bool
	var collectExpression func(*ast.Node)
	var collectFunction func(*ast.Node)
	var collectProperty func(*ast.Node)
	var collectDeclaration func(*ast.Node)
	var collectSelectedProperty func(*ast.Node, string)
	var collectObjectRest func(*ast.Node, map[string]bool)
	var collectArrayRest func(*ast.Node, int)
	var collectNestedBindingSource func(*ast.Node)
	var collectResolvedFunction func(*ast.Node)
	var activeSelections map[*ast.Node]map[string]bool
	activeFunctions := make(map[*ast.Node]bool)
	cloneUsages := func(source map[*ast.Node]int) map[*ast.Node]int {
		if len(source) == 0 {
			return nil
		}
		cloned := make(map[*ast.Node]int, len(source))
		for name, count := range source {
			cloned[name] = count
		}
		return cloned
	}
	collectAlternativeActions := func(actions []func(), alternativesRemainDistinct bool) {
		if alternativesRemainDistinct {
			for _, action := range actions {
				action()
			}
			return
		}
		before := cloneUsages(usages)
		maxDeltas := make(map[*ast.Node]int)
		for _, action := range actions {
			usages = cloneUsages(before)
			action()
			for name, count := range usages {
				if delta := count - before[name]; delta > maxDeltas[name] {
					maxDeltas[name] = delta
				}
			}
		}
		usages = cloneUsages(before)
		for name, delta := range maxDeltas {
			if usages == nil {
				usages = make(map[*ast.Node]int)
			}
			usages[name] += delta
		}
	}
	collectAlternatives := func(expressions []*ast.Node, alternativesRemainDistinct bool) {
		actions := make([]func(), 0, len(expressions))
		for _, expression := range expressions {
			actions = append(actions, func() { collectExpression(expression) })
		}
		collectAlternativeActions(actions, alternativesRemainDistinct)
	}
	type callArgumentGroup struct {
		returnCount int
		arguments   []*ast.Node
	}
	collectCallArguments := func(callNode *ast.Node, arguments []*ast.Node, parameterOffset int) {
		groups := make(map[*ast.Node]*callArgumentGroup)
		for index, argument := range arguments {
			for name, returnUsage := range genericCallArgumentReturnUsages(tc, callNode, index+parameterOffset) {
				groupKey := returnUsage.group
				if groupKey == nil {
					groupKey = name
				}
				group := groups[groupKey]
				if group == nil {
					group = &callArgumentGroup{returnCount: returnUsage.count}
					groups[groupKey] = group
				} else if returnUsage.count > group.returnCount {
					group.returnCount = returnUsage.count
				}
				group.arguments = append(group.arguments, argument)
			}
		}
		for _, group := range groups {
			before := cloneUsages(usages)
			collectAlternatives(group.arguments, false)
			if group.returnCount <= 1 {
				continue
			}
			for name, count := range usages {
				if count-before[name] == 1 {
					usages[name]++
				}
			}
		}
	}

	recordSpreadUsages := func(expression *ast.Node, spreadUsages map[*ast.Node]int) {
		var before map[*ast.Node]int
		if len(spreadUsages) != 0 {
			before = make(map[*ast.Node]int, len(spreadUsages))
		}
		for name, count := range spreadUsages {
			before[name] = usages[name]
			if usages == nil {
				usages = make(map[*ast.Node]int)
			}
			usages[name] += count
		}
		// A spread operand can itself be an object assembled from another
		// generic spread, or a local initialized from one. For parameters whose
		// resolved surface already exposes a generic, retain only the outer
		// occurrence so a nested construction isn't counted twice.
		if len(spreadUsages) == 0 || expression.SubtreeFacts()&ast.SubtreeContainsESObjectRestOrSpread != 0 {
			collectExpression(expression)
			for name, count := range spreadUsages {
				usages[name] = before[name] + count
			}
		}
	}
	recordSpread := func(expression *ast.Node) {
		recordSpreadUsages(expression, collectLegacyNullableSpreadUsages(tc, expression))
	}

	collectProperty = func(property *ast.Node) {
		if property == nil {
			return
		}
		switch property.Kind {
		case ast.KindSpreadAssignment:
			recordSpread(property.AsSpreadAssignment().Expression)
		case ast.KindPropertyAssignment:
			initializer := property.AsPropertyAssignment().Initializer
			if !computedPropertyAssumesMultipleUses(tc, property) {
				collectExpression(initializer)
				return
			}
			before := cloneUsages(usages)
			collectExpression(initializer)
			for name, count := range usages {
				if count-before[name] == 1 {
					usages[name]++
				}
			}
		case ast.KindShorthandPropertyAssignment:
			if symbol := tc.GetShorthandAssignmentValueSymbol(property); symbol != nil {
				collectDeclaration(symbol.ValueDeclaration)
			}
			collectExpression(property.AsShorthandPropertyAssignment().ObjectAssignmentInitializer)
		case ast.KindMethodDeclaration, ast.KindGetAccessor:
			collectFunction(property)
		case ast.KindPropertyDeclaration:
			collectExpression(property.Initializer())
		}
	}

	collectObjectRest = func(base *ast.Node, excluded map[string]bool) {
		if base == nil {
			return
		}
		base = ast.SkipParentheses(base)
		switch base.Kind {
		case ast.KindNonNullExpression, ast.KindSatisfiesExpression:
			collectObjectRest(base.Expression(), excluded)
			return
		case ast.KindAsExpression, ast.KindTypeAssertionExpression:
			if ast.IsConstTypeReference(base.Type()) {
				collectObjectRest(base.Expression(), excluded)
			} else {
				recordSpread(base)
			}
			return
		case ast.KindIdentifier:
			if directUsages := collectLegacyNullableSpreadUsages(tc, base); len(directUsages) != 0 {
				recordSpreadUsages(base, directUsages)
				return
			}
			if symbol := tc.GetSymbolAtLocation(base); symbol != nil && symbol.ValueDeclaration != nil {
				declaration := symbol.ValueDeclaration
				selectionKey := objectRestSelectionKey(excluded)
				seenNames := activeSelections[declaration]
				if seenNames == nil {
					seenNames = make(map[string]bool)
					if activeSelections == nil {
						activeSelections = make(map[*ast.Node]map[string]bool)
					}
					activeSelections[declaration] = seenNames
				}
				if seenNames[selectionKey] {
					return
				}
				seenNames[selectionKey] = true
				collectObjectRest(declaration.Initializer(), excluded)
				delete(seenNames, selectionKey)
				return
			}
		case ast.KindObjectLiteralExpression:
			seen := excluded
			var surfaceSpreadUsages map[*ast.Node]int
			properties := base.AsObjectLiteralExpression().Properties.Nodes
			for index := len(properties) - 1; index >= 0; index-- {
				property := properties[index]
				if property.Kind == ast.KindSpreadAssignment {
					spread := ast.SkipParentheses(property.AsSpreadAssignment().Expression)
					directUsages := collectLegacyNullableSpreadUsages(tc, spread)
					var before map[*ast.Node]int
					if len(directUsages) != 0 {
						before = make(map[*ast.Node]int, len(directUsages))
						for name := range directUsages {
							before[name] = usages[name]
						}
					}
					if len(directUsages) != 0 {
						recordSpreadUsages(spread, directUsages)
					} else {
						if seen == nil {
							seen = make(map[string]bool)
						}
						collectObjectRest(spread, seen)
					}
					for name, count := range directUsages {
						previous := surfaceSpreadUsages[name]
						allowed := count - previous
						if allowed < 0 {
							allowed = 0
						}
						if usages[name]-before[name] > allowed {
							usages[name] = before[name] + allowed
						}
						if count > previous {
							if surfaceSpreadUsages == nil {
								surfaceSpreadUsages = make(map[*ast.Node]int)
							}
							surfaceSpreadUsages[name] = count
						}
					}
					continue
				}
				nameNode := property.Name()
				if nameNode == nil {
					continue
				}
				name := ast.GetPropertyNameForPropertyNameNode(nameNode)
				if name == "" {
					collectProperty(property)
					continue
				}
				if seen[name] {
					continue
				}
				if seen == nil {
					seen = make(map[string]bool)
				}
				seen[name] = true
				collectProperty(property)
			}
			return
		}

		// A non-literal spread source can still carry the enclosing generic
		// through its residual surface after named properties are removed. Its
		// known own properties also overwrite earlier literal properties.
		if spreadType := tc.GetTypeAtLocation(base); spreadType != nil {
			if excluded == nil {
				excluded = make(map[string]bool)
			}
			for _, spreadProperty := range checker.Checker_getPropertiesOfType(tc, spreadType) {
				if isObjectSpreadProperty(spreadProperty) {
					excluded[spreadProperty.Name] = true
				}
			}
		}
		recordSpread(base)
	}

	collectArrayRest = func(base *ast.Node, start int) {
		if base == nil {
			return
		}
		base = ast.SkipParentheses(base)
		if base.Kind == ast.KindIdentifier {
			if symbol := tc.GetSymbolAtLocation(base); symbol != nil && symbol.ValueDeclaration != nil {
				declaration := symbol.ValueDeclaration
				selectionKey := "\x00array-rest:" + strconv.Itoa(start)
				seenNames := activeSelections[declaration]
				if seenNames == nil {
					seenNames = make(map[string]bool)
					if activeSelections == nil {
						activeSelections = make(map[*ast.Node]map[string]bool)
					}
					activeSelections[declaration] = seenNames
				}
				if seenNames[selectionKey] {
					return
				}
				seenNames[selectionKey] = true
				collectArrayRest(declaration.Initializer(), start)
				delete(seenNames, selectionKey)
			}
			return
		}
		if base.Kind != ast.KindArrayLiteralExpression {
			collectExpression(base)
			return
		}
		for index, element := range base.AsArrayLiteralExpression().Elements.Nodes {
			if index < start || element == nil || element.Kind == ast.KindOmittedExpression {
				continue
			}
			if element.Kind == ast.KindSpreadElement {
				collectExpression(element.AsSpreadElement().Expression)
			} else {
				collectExpression(element)
			}
		}
	}

	collectNestedBindingSource = func(bindingNode *ast.Node) {
		var selectors []string
		current := bindingNode
		var base *ast.Node
		for current != nil && current.Kind == ast.KindBindingElement {
			binding := current.AsBindingElement()
			if binding.DotDotDotToken != nil || current.Parent == nil {
				return
			}
			switch current.Parent.Kind {
			case ast.KindObjectBindingPattern:
				name := binding.PropertyName
				if name == nil {
					name = binding.Name()
				}
				if name == nil {
					return
				}
				selectors = append(selectors, ast.GetPropertyNameForPropertyNameNode(name))
			case ast.KindArrayBindingPattern:
				index := slices.Index(current.Parent.AsBindingPattern().Elements.Nodes, current)
				if index < 0 {
					return
				}
				selectors = append(selectors, strconv.Itoa(index))
			default:
				return
			}

			owner := current.Parent.Parent
			if owner == nil {
				return
			}
			if owner.Kind == ast.KindVariableDeclaration {
				if owner.Type() != nil {
					return
				}
				base = owner.Initializer()
				break
			}
			current = owner
		}
		if base == nil {
			return
		}

		var resolveSelectedInitializer func(*ast.Node, string) *ast.Node
		resolving := make(map[*ast.Node]bool)
		resolveSelectedInitializer = func(expression *ast.Node, name string) *ast.Node {
			if expression == nil {
				return nil
			}
			expression = ast.SkipParentheses(expression)
			switch expression.Kind {
			case ast.KindNonNullExpression, ast.KindSatisfiesExpression:
				return resolveSelectedInitializer(expression.Expression(), name)
			case ast.KindAsExpression, ast.KindTypeAssertionExpression:
				if ast.IsConstTypeReference(expression.Type()) {
					return resolveSelectedInitializer(expression.Expression(), name)
				}
				return nil
			case ast.KindIdentifier:
				if symbol := tc.GetSymbolAtLocation(expression); symbol != nil && symbol.ValueDeclaration != nil {
					declaration := symbol.ValueDeclaration
					if resolving[declaration] {
						return nil
					}
					resolving[declaration] = true
					return resolveSelectedInitializer(declaration.Initializer(), name)
				}
			case ast.KindObjectLiteralExpression:
				properties := expression.AsObjectLiteralExpression().Properties.Nodes
				for index := len(properties) - 1; index >= 0; index-- {
					property := properties[index]
					if property.Kind == ast.KindSpreadAssignment {
						spread := property.AsSpreadAssignment().Expression
						spreadType := tc.GetTypeAtLocation(spread)
						if spreadType != nil && checker.Checker_getPropertyOfType(tc, spreadType, name) != nil {
							return resolveSelectedInitializer(spread, name)
						}
						continue
					}
					propertyName := property.Name()
					if propertyName == nil || ast.GetPropertyNameForPropertyNameNode(propertyName) != name {
						continue
					}
					switch property.Kind {
					case ast.KindPropertyAssignment:
						return property.AsPropertyAssignment().Initializer
					case ast.KindShorthandPropertyAssignment:
						return property.Name()
					}
					return nil
				}
			case ast.KindArrayLiteralExpression:
				index, err := strconv.Atoi(name)
				if err == nil && index >= 0 {
					elements := expression.AsArrayLiteralExpression().Elements.Nodes
					if index < len(elements) && elements[index] != nil && elements[index].Kind != ast.KindOmittedExpression && elements[index].Kind != ast.KindSpreadElement {
						return elements[index]
					}
				}
			}
			return nil
		}

		for index := len(selectors) - 1; index >= 0; index-- {
			base = resolveSelectedInitializer(base, selectors[index])
			if base == nil {
				return
			}
		}
		collectExpression(base)
	}

	collectDeclaration = func(declaration *ast.Node) {
		if declaration == nil || activeDeclarations[declaration] {
			return
		}
		if activeDeclarations == nil {
			activeDeclarations = make(map[*ast.Node]bool)
		}
		activeDeclarations[declaration] = true
		defer delete(activeDeclarations, declaration)
		if ast.IsFunctionLike(declaration) {
			collectFunction(declaration)
			return
		}
		switch declaration.Kind {
		case ast.KindVariableDeclaration, ast.KindParameter,
			ast.KindPropertyDeclaration, ast.KindPropertySignature:
			if declaration.Type() != nil {
				return
			}
			collectExpression(declaration.Initializer())
		case ast.KindPropertyAssignment, ast.KindEnumMember:
			collectExpression(declaration.Initializer())
		case ast.KindBindingElement:
			binding := declaration.AsBindingElement()
			if root := ast.GetRootDeclaration(declaration); root != nil &&
				(root.Kind == ast.KindVariableDeclaration || root.Kind == ast.KindParameter) && root.Type() != nil {
				return
			}
			if declaration.Parent != nil && declaration.Parent.Kind == ast.KindObjectBindingPattern {
				root := declaration.Parent.Parent
				if binding.DotDotDotToken != nil && root != nil && root.Kind == ast.KindVariableDeclaration {
					excluded := make(map[string]bool)
					for _, sibling := range declaration.Parent.AsBindingPattern().Elements.Nodes {
						if sibling == declaration || sibling == nil || sibling.Kind != ast.KindBindingElement {
							continue
						}
						siblingBinding := sibling.AsBindingElement()
						name := siblingBinding.PropertyName
						if name == nil {
							name = siblingBinding.Name()
						}
						if name != nil {
							excluded[ast.GetPropertyNameForPropertyNameNode(name)] = true
						}
					}
					collectObjectRest(root.Initializer(), excluded)
					collectExpression(binding.Initializer)
					return
				}
				name := binding.PropertyName
				if name == nil {
					name = binding.Name()
				}
				if name != nil && root != nil && root.Kind == ast.KindVariableDeclaration {
					collectSelectedProperty(root.Initializer(), ast.GetPropertyNameForPropertyNameNode(name))
				} else if root != nil && root.Kind == ast.KindBindingElement {
					collectNestedBindingSource(declaration)
				}
				collectExpression(binding.Initializer)
			} else if declaration.Parent != nil && declaration.Parent.Kind == ast.KindArrayBindingPattern {
				root := declaration.Parent.Parent
				if root != nil && root.Kind == ast.KindVariableDeclaration {
					index := slices.Index(declaration.Parent.AsBindingPattern().Elements.Nodes, declaration)
					if binding.DotDotDotToken != nil {
						collectArrayRest(root.Initializer(), index)
					} else if index >= 0 {
						collectSelectedProperty(root.Initializer(), strconv.Itoa(index))
					}
				} else if root != nil && root.Kind == ast.KindBindingElement {
					collectNestedBindingSource(declaration)
				}
				collectExpression(binding.Initializer)
			} else {
				collectExpression(binding.Initializer)
			}
		case ast.KindShorthandPropertyAssignment:
			collectProperty(declaration)
		}
	}

	collectSelectedProperty = func(base *ast.Node, name string) {
		if base == nil || name == "" {
			return
		}
		base = ast.SkipParentheses(base)
		switch base.Kind {
		case ast.KindNonNullExpression, ast.KindSatisfiesExpression:
			collectSelectedProperty(base.Expression(), name)
			return
		case ast.KindAsExpression, ast.KindTypeAssertionExpression:
			if ast.IsConstTypeReference(base.Type()) {
				collectSelectedProperty(base.Expression(), name)
			}
			return
		case ast.KindConditionalExpression:
			conditional := base.AsConditionalExpression()
			collectAlternativeActions([]func(){
				func() { collectSelectedProperty(conditional.WhenTrue, name) },
				func() { collectSelectedProperty(conditional.WhenFalse, name) },
			}, utils.IsUnionType(tc.GetTypeAtLocation(base)))
			return
		case ast.KindBinaryExpression:
			binary := base.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				collectAlternativeActions([]func(){
					func() { collectSelectedProperty(binary.Left, name) },
					func() { collectSelectedProperty(binary.Right, name) },
				}, utils.IsUnionType(tc.GetTypeAtLocation(base)))
			case ast.KindCommaToken, ast.KindEqualsToken:
				collectSelectedProperty(binary.Right, name)
			}
			return
		case ast.KindIdentifier:
			symbol := tc.GetSymbolAtLocation(base)
			if symbol != nil && symbol.ValueDeclaration != nil {
				declaration := symbol.ValueDeclaration
				seenNames := activeSelections[declaration]
				if seenNames == nil {
					seenNames = make(map[string]bool)
					if activeSelections == nil {
						activeSelections = make(map[*ast.Node]map[string]bool)
					}
					activeSelections[declaration] = seenNames
				}
				if seenNames[name] {
					return
				}
				seenNames[name] = true
				switch declaration.Kind {
				case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement,
					ast.KindPropertyDeclaration, ast.KindPropertyAssignment:
					collectSelectedProperty(declaration.Initializer(), name)
				}
				delete(seenNames, name)
			}
			return
		case ast.KindObjectLiteralExpression:
			properties := base.AsObjectLiteralExpression().Properties.Nodes
			for index := len(properties) - 1; index >= 0; index-- {
				property := properties[index]
				if property.Kind == ast.KindSpreadAssignment {
					spread := property.AsSpreadAssignment().Expression
					spreadType := tc.GetTypeAtLocation(spread)
					if spreadType != nil && checker.Checker_getPropertyOfType(tc, spreadType, name) != nil {
						collectSelectedProperty(spread, name)
						return
					}
					continue
				}
				propertyName := property.Name()
				if propertyName != nil && ast.GetPropertyNameForPropertyNameNode(propertyName) == name {
					collectProperty(property)
					return
				}
			}
			return
		case ast.KindArrayLiteralExpression:
			index, err := strconv.Atoi(name)
			if err != nil || index < 0 {
				return
			}
			position := 0
			elements := base.AsArrayLiteralExpression().Elements.Nodes
			for elementIndex, element := range elements {
				if element == nil || element.Kind == ast.KindOmittedExpression {
					if position == index {
						return
					}
					position++
					continue
				}
				if element.Kind == ast.KindSpreadElement {
					// Without evaluating tuple length, the selected slot may come
					// from either side of this spread. Follow the remaining return
					// candidates conservatively.
					collectExpression(element.AsSpreadElement().Expression)
					for _, remaining := range elements[elementIndex+1:] {
						collectExpression(remaining)
					}
					return
				}
				if position == index {
					collectExpression(element)
					return
				}
				position++
			}
			return
		}

		if objectType := tc.GetTypeAtLocation(base); objectType != nil {
			if property := checker.Checker_getPropertyOfType(tc, objectType, name); property != nil {
				collectProperty(property.ValueDeclaration)
			}
		}
	}

	collectFunction = func(function *ast.Node) {
		if function == nil || !ast.IsFunctionLike(function) || function.Type() != nil || activeFunctions[function] {
			return
		}
		activeFunctions[function] = true
		defer delete(activeFunctions, function)
		assumeMultipleReturnUses := returnTypeAssumesMultipleUses(tc, tc.GetSignatureFromDeclaration(function))
		collectReturnedExpressions := func(expressions []*ast.Node) {
			before := cloneUsages(usages)
			returnType := tc.GetReturnTypeOfSignature(tc.GetSignatureFromDeclaration(function))
			collectAlternatives(expressions, utils.IsUnionType(returnType))
			if !assumeMultipleReturnUses {
				return
			}
			for name, count := range usages {
				if count-before[name] == 1 {
					usages[name]++
				}
			}
		}
		if function != node {
			for _, parameter := range function.Parameters() {
				if parameter.Type() == nil {
					collectExpression(parameter.Initializer())
				}
			}
		}
		functionBody := function.Body()
		if functionBody == nil {
			return
		}
		if functionBody.Kind != ast.KindBlock {
			collectReturnedExpressions([]*ast.Node{functionBody})
			return
		}
		var returnExpressions []*ast.Node
		ast.ForEachReturnStatement(functionBody, func(statement *ast.Node) bool {
			returnExpressions = append(returnExpressions, statement.Expression())
			return false
		})
		collectReturnedExpressions(returnExpressions)
		if ast.GetFunctionFlags(function)&ast.FunctionFlagsGenerator != 0 {
			var yieldExpressions []*ast.Node
			var visitYield func(*ast.Node) bool
			visitYield = func(current *ast.Node) bool {
				if current != functionBody && ast.IsFunctionLike(current) {
					return false
				}
				if current.Kind == ast.KindYieldExpression {
					yieldExpressions = append(yieldExpressions, current.AsYieldExpression().Expression)
					return false
				}
				return current.ForEachChild(visitYield)
			}
			functionBody.ForEachChild(visitYield)
			collectReturnedExpressions(yieldExpressions)
		}
	}

	collectResolvedFunction = func(callNode *ast.Node) {
		sig := checker.Checker_getResolvedSignature(tc, callNode, nil, checker.CheckModeNormal)
		if sig != nil {
			collectFunction(sig.Declaration())
		}
	}

	collectExpression = func(expression *ast.Node) {
		if expression == nil {
			return
		}
		switch expression.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression,
			ast.KindSatisfiesExpression, ast.KindAwaitExpression,
			ast.KindPartiallyEmittedExpression:
			collectExpression(expression.Expression())

		case ast.KindAsExpression, ast.KindTypeAssertionExpression:
			if ast.IsConstTypeReference(expression.Type()) {
				collectExpression(expression.Expression())
			}

		case ast.KindConditionalExpression:
			conditional := expression.AsConditionalExpression()
			collectAlternatives([]*ast.Node{conditional.WhenTrue, conditional.WhenFalse}, utils.IsUnionType(tc.GetTypeAtLocation(expression)))

		case ast.KindBinaryExpression:
			binary := expression.AsBinaryExpression()
			switch binary.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				collectAlternatives([]*ast.Node{binary.Left, binary.Right}, utils.IsUnionType(tc.GetTypeAtLocation(expression)))
			case ast.KindCommaToken, ast.KindEqualsToken:
				collectExpression(binary.Right)
			}

		case ast.KindObjectLiteralExpression:
			collectObjectRest(expression, nil)

		case ast.KindArrayLiteralExpression:
			for _, element := range expression.AsArrayLiteralExpression().Elements.Nodes {
				if element.Kind == ast.KindSpreadElement {
					collectExpression(element.AsSpreadElement().Expression)
				} else {
					collectExpression(element)
				}
			}

		case ast.KindArrowFunction, ast.KindFunctionExpression:
			collectFunction(expression)

		case ast.KindClassExpression:
			if members := expression.MemberList(); members != nil {
				for _, member := range members.Nodes {
					collectProperty(member)
				}
			}

		case ast.KindCallExpression:
			collectCallArguments(expression, expression.AsCallExpression().Arguments.Nodes, 0)
			collectResolvedFunction(expression)

		case ast.KindNewExpression:
			collectCallArguments(expression, expression.AsNewExpression().Arguments.Nodes, 0)

		case ast.KindTaggedTemplateExpression:
			tagged := expression.AsTaggedTemplateExpression()
			if tagged.Template.Kind == ast.KindTemplateExpression {
				spans := tagged.Template.AsTemplateExpression().TemplateSpans.Nodes
				arguments := make([]*ast.Node, 0, len(spans))
				for _, span := range spans {
					arguments = append(arguments, span.AsTemplateSpan().Expression)
				}
				collectCallArguments(expression, arguments, 1)
			}
			collectResolvedFunction(expression)

		case ast.KindSpreadElement:
			collectExpression(expression.AsSpreadElement().Expression)

		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			name := ast.GetElementOrPropertyAccessName(expression)
			if name != nil {
				collectSelectedProperty(expression.Expression(), ast.GetPropertyNameForPropertyNameNode(name))
			} else {
				collectExpression(expression.Expression())
			}

		case ast.KindIdentifier:
			symbol := tc.GetSymbolAtLocation(expression)
			if symbol == nil {
				return
			}
			collectDeclaration(symbol.ValueDeclaration)
		}
	}

	if rootExpression != nil {
		collectExpression(rootExpression)
	} else {
		collectFunction(node)
	}
	return usages
}

// needsLegacyObjectSpreadRecovery identifies the one default mismatch behind
// this compatibility path: TypeScript 5.9 defaults omitted `strict` and
// `strictNullChecks` to false, while current tsgo defaults them to true.
func needsLegacyObjectSpreadRecovery(options *core.CompilerOptions) bool {
	return options != nil &&
		options.Strict == core.TSUnknown &&
		options.StrictNullChecks == core.TSUnknown
}

// collectLegacyObjectSpreadSignatureUsages is kept separate from the normal
// checker walk so projects with explicit strictness retain the rule's cheap
// allocation profile. The compatibility analysis is only needed when both
// strictness settings were omitted.
func collectLegacyObjectSpreadSignatureUsages(tc *checker.Checker, node *ast.Node, usages map[*ast.Node]int) {
	collectDeclaration := func(declaration *ast.Node) {
		if !ast.IsFunctionLike(declaration) {
			return
		}
		sig := tc.GetSignatureFromDeclaration(declaration)
		if sig == nil {
			return
		}
		for _, parameter := range sig.Parameters() {
			if parameter.ValueDeclaration == nil || parameter.ValueDeclaration.Type() != nil {
				continue
			}
			for name, count := range collectInferredObjectSpreadExpressionUsages(tc, parameter.ValueDeclaration.Initializer()) {
				usages[name] += count
			}
		}
		for name, count := range collectInferredObjectSpreadUsages(tc, sig.Declaration()) {
			usages[name] += count
		}
	}

	if ast.IsClassLike(node) {
		if members := node.MemberList(); members != nil {
			for _, member := range members.Nodes {
				collectDeclaration(member)
				if member.Kind == ast.KindPropertyDeclaration || member.Kind == ast.KindPropertySignature {
					if initializer := member.Initializer(); ast.IsFunctionLike(initializer) {
						collectDeclaration(initializer)
					}
				}
			}
		}
		return
	}
	collectDeclaration(node)
}

// isTypeParameterRepeatedInAST is the cheap syntactic shortcut: if the type
// parameter is textually referenced at least twice within its own declaring
// signature (excluding its own declaration and anything past the function
// body / end of signature), it's provably not "sole" and the type-checker
// walk can be skipped entirely.
func isTypeParameterRepeatedInAST(ctx rule.RuleContext, typeParamNode *ast.Node, container *ast.Node, sym *ast.Symbol) bool {
	cutoff := declaringSignatureEnd(container)
	total := 0
	for _, ref := range ctx.Refs.References(sym) {
		// References inside the type parameter's own declaration (e.g. a
		// self-referential constraint like `T extends Foo<T>`) don't count.
		if ref.Pos() < typeParamNode.End() && ref.End() > typeParamNode.Pos() {
			continue
		}
		// References outside the declaring signature (i.e. only reachable
		// from the function body) don't count either.
		if ref.Pos() > cutoff {
			continue
		}
		if isDirectGenericTypeArgumentUsage(ref) {
			return true
		}
		total++
		if total >= 2 {
			return true
		}
	}
	return false
}

// declaringSignatureEnd returns the position past which a reference no
// longer belongs to node's own declaring signature: the start of its body
// (function-like with an implementation, or a class's member list), the end
// of an explicit return-type annotation when there's no body, or unbounded
// when neither exists.
func declaringSignatureEnd(node *ast.Node) int {
	if body := node.Body(); body != nil {
		return body.Pos()
	}
	if ast.IsClassLike(node) {
		if members := node.MemberList(); members != nil {
			return members.Pos()
		}
	}
	if returnType := node.Type(); returnType != nil {
		return returnType.End()
	}
	return math.MaxInt
}

// isDirectGenericTypeArgumentUsage reports whether identNode — a bare
// reference to the type parameter used as a type (`T`) — sits directly as a
// type argument of an outer generic type reference, e.g. the T in `Map<T,
// V>` or `Foo<T>`. Such a usage is treated as an automatic "used multiple
// times", mirroring upstream, unless the outer reference is `Array<T>` /
// `ReadonlyArray<T>`, which defer to the more precise type-checker phase.
func isDirectGenericTypeArgumentUsage(identNode *ast.Node) bool {
	immediateParent := identNode.Parent
	if immediateParent == nil || immediateParent.Kind != ast.KindTypeReference {
		return false
	}
	// ESLint's AST has no parenthesized-type node, so in `Foo<(T)>` the bare
	// reference is what its type-argument list holds. Peel tsgo's wrappers to
	// reach the node that actually sits in the list.
	argument := immediateParent
	for argument.Parent != nil && argument.Parent.Kind == ast.KindParenthesizedType {
		argument = argument.Parent
	}
	grandparent := skipUnionIntersectionUpward(argument.Parent)
	if grandparent == nil || !canHaveTypeArgumentsList(grandparent.Kind) {
		return false
	}
	if !slices.Contains(grandparent.TypeArguments(), argument) {
		return false
	}
	// The Array<T>/ReadonlyArray<T> exclusion only makes sense for an actual
	// type reference (e.g. `Foo<T>`), not for a type argument attached to a
	// call/new/tagged-template/JSX expression (e.g. `foo<T>()`).
	if grandparent.Kind == ast.KindTypeReference {
		outerName := grandparent.AsTypeReferenceNode().TypeName
		if outerName != nil && outerName.Kind == ast.KindIdentifier {
			switch outerName.AsIdentifier().Text {
			case "Array", "ReadonlyArray":
				return false
			}
		}
	}
	return true
}

// canHaveTypeArgumentsList reports whether kind is a node that can directly
// own a `<...>` type-argument list — matching every construct ESLint
// represents via a TSTypeParameterInstantiation wrapper: generic type
// references, and generic call/new/tagged-template/JSX expressions.
func canHaveTypeArgumentsList(kind ast.Kind) bool {
	switch kind {
	case ast.KindTypeReference, ast.KindExpressionWithTypeArguments,
		ast.KindImportType, ast.KindTypeQuery,
		ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression,
		ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		return true
	}
	return false
}

func skipUnionIntersectionUpward(node *ast.Node) *ast.Node {
	for node != nil && (node.Kind == ast.KindUnionType || node.Kind == ast.KindIntersectionType) {
		node = node.Parent
	}
	return node
}

// countTypeParameterUsage resolves, for every type parameter reachable from
// node, how many times the type checker sees it used across the signature
// (including inferred return types). For a class, every member contributes
// (with generic type-argument usages always treated as "multiple"); for a
// function-like node, only its own signature contributes.
func countTypeParameterUsage(tc *checker.Checker, node *ast.Node, recoverLegacyObjectSpreads bool) map[*ast.Node]int {
	counts := make(map[*ast.Node]int)

	if ast.IsClassLike(node) {
		for _, typeParamNode := range node.TypeParameters() {
			collectTypeParameterUsageCounts(tc, typeParamNode, counts, true)
		}
		if members := node.MemberList(); members != nil {
			for _, member := range members.Nodes {
				collectTypeParameterUsageCounts(tc, member, counts, true)
			}
		}
	} else {
		collectTypeParameterUsageCounts(tc, node, counts, false)
	}
	if recoverLegacyObjectSpreads {
		collectLegacyObjectSpreadSignatureUsages(tc, node, counts)
	}

	return counts
}

// collectTypeParameterUsageCounts walks the resolved type graph rooted at
// node, incrementing foundIdentifierUsages for every type-parameter
// declaration identifier it encounters. Each type parameter always
// contributes a baseline "self" count of 1 (from being enumerated as one of
// its declaring signature's own type parameters, see the
// `signature.TypeParameters()` walk in visitSignature and the class
// self-visit loop in countTypeParameterUsage); every further occurrence adds
// 1 (or 2, when assumeMultipleUses applies) on top of that baseline. A final
// count of 1 or 2 is "sole"; 3 or more means it's genuinely used more than
// once.
//
// fromClass marks members visited on behalf of a class: type-reference type
// arguments always count as multiple use there, since a class's type
// parameters are shared across every member.
func collectTypeParameterUsageCounts(tc *checker.Checker, node *ast.Node, foundIdentifierUsages map[*ast.Node]int, fromClass bool) {
	collectTypeParameterUsageCountsWorker(tc, node, foundIdentifierUsages, fromClass)
}

func collectTypeParameterUsageCountsWorker(tc *checker.Checker, node *ast.Node, foundIdentifierUsages map[*ast.Node]int, fromClass bool) {
	visitedProperties := make(map[*checker.Type]bool)
	visitedConstraints := make(map[*ast.Node]bool)
	visitedDefault := false
	typeUsages := make(map[*checker.Type]int)
	functionLikeType := false

	incrementIdentifierCount := func(id *ast.Node, assumeMultipleUses bool) {
		value := 1
		if assumeMultipleUses {
			value = 2
		}
		foundIdentifierUsages[id] += value
	}

	// incrementTypeUsages guards against unbounded recursion on recursive
	// generic types (e.g. `type T = { [P in keyof T]: T }`): once a type has
	// been seen more than 9 times, every referenced type parameter has
	// already been counted enough to qualify as used, so visiting stops.
	incrementTypeUsages := func(t *checker.Type) int {
		typeUsages[t]++
		return typeUsages[t]
	}

	var visitType func(t *checker.Type, assumeMultipleUses bool, isReturnType bool)

	visitSymbolsListOnce := func(owner *checker.Type, symbols []*ast.Symbol, assumeMultipleUses bool) {
		if visitedProperties[owner] {
			return
		}
		visitedProperties[owner] = true
		for _, sym := range symbols {
			visitType(tc.GetTypeOfSymbol(sym), assumeMultipleUses, false)
		}
	}

	visitTypesList := func(types []*checker.Type, assumeMultipleUses bool) {
		for _, t := range types {
			visitType(t, assumeMultipleUses, false)
		}
	}

	visitSignature := func(sig *checker.Signature) {
		if sig == nil {
			return
		}
		if this := sig.ThisParameter(); this != nil {
			visitType(tc.GetTypeOfSymbol(this), false, false)
		}
		for _, param := range sig.Parameters() {
			visitType(tc.GetTypeOfSymbol(param), false, false)
		}
		for _, typeParam := range sig.TypeParameters() {
			visitType(typeParam, false, false)
		}
		returnType := tc.GetReturnTypeOfSignature(sig)
		if predicate := tc.GetTypePredicateOfSignature(sig); predicate != nil && predicate.Type() != nil {
			returnType = predicate.Type()
		}
		visitType(returnType, false, true)
	}

	visitType = func(t *checker.Type, assumeMultipleUses bool, isReturnType bool) {
		if t == nil || incrementTypeUsages(t) > 9 {
			return
		}

		switch {
		case utils.IsTypeParameter(t):
			sym := t.Symbol()
			if sym == nil || len(sym.Declarations) == 0 {
				return
			}
			// A polymorphic `this` type also carries TypeFlagsTypeParameter, but
			// its symbol is the declaring class/interface, so its first
			// declaration is not a type parameter. Such a type contributes no
			// count and has no constraint or default to recurse into.
			decl := sym.Declarations[0]
			if !ast.IsTypeParameterDeclaration(decl) {
				return
			}
			declTypeParam := decl.AsTypeParameterDeclaration()
			nameNode := declTypeParam.Name()
			if nameNode != nil {
				incrementIdentifierCount(nameNode, assumeMultipleUses)
			}
			// Resolve the constraint/default through the checker on t itself
			// (not by re-resolving the raw declaration node) so a mapped
			// type's own bound type parameter (`K` in `{ [K in T]: ... }`)
			// picks up whatever substitution applies to this specific
			// instantiation, e.g. when reached through an inferred return
			// type of a call to another generic function.
			if declTypeParam.Constraint != nil && !visitedConstraints[declTypeParam.Constraint] {
				visitedConstraints[declTypeParam.Constraint] = true
				visitType(tc.GetConstraintOfTypeParameter(t), false, false)
			}
			if declTypeParam.DefaultType != nil && !visitedDefault {
				visitedDefault = true
				visitType(tc.GetDefaultFromTypeParameter(t), false, false)
			}

		case t.Alias() != nil && t.Alias().TypeArguments() != nil:
			// A generic type-alias reference like `Exclude<T, null>`. We don't
			// descend into the alias's own definition, so it's safest to assume
			// every argument is used multiple times.
			visitTypesList(t.Alias().TypeArguments(), true)

		case utils.IsUnionType(t) || utils.IsIntersectionType(t):
			visitTypesList(t.Types(), assumeMultipleUses)

		case utils.IsTypeFlagSet(t, checker.TypeFlagsIndexedAccess):
			iat := t.AsIndexedAccessType()
			visitType(iat.ObjectType(), assumeMultipleUses, false)
			visitType(iat.IndexType(), assumeMultipleUses, false)

		case utils.IsTypeReference(t):
			target := t.Target()
			for _, typeArgument := range checker.Checker_getTypeArguments(tc, t) {
				thisAssumeMultipleUses := fromClass || assumeMultipleUses
				if !thisAssumeMultipleUses {
					switch {
					case checker.IsTupleType(target):
						thisAssumeMultipleUses = isReturnType && !target.AsTupleType().IsReadonly()
					case tc.IsArrayType(target):
						sym := t.Symbol()
						thisAssumeMultipleUses = isReturnType && sym != nil && sym.Name == "Array"
					default:
						thisAssumeMultipleUses = true
					}
				}
				visitType(typeArgument, thisAssumeMultipleUses, isReturnType)
			}

		case utils.IsTypeFlagSet(t, checker.TypeFlagsTemplateLiteral):
			for _, subType := range t.AsTemplateLiteralType().Types() {
				visitType(subType, assumeMultipleUses, false)
			}

		case utils.IsTypeFlagSet(t, checker.TypeFlagsConditional):
			ct := t.AsConditionalType()
			visitType(ct.CheckType(), assumeMultipleUses, false)
			visitType(ct.ExtendsType(), assumeMultipleUses, false)

		case utils.IsObjectType(t):
			properties := checker.Checker_getPropertiesOfType(tc, t)
			visitSymbolsListOnce(t, properties, false)

			if checker.Type_objectFlags(t)&checker.ObjectFlagsMapped != 0 {
				visitType(checker.Checker_getTypeParameterFromMappedType(tc, t), false, false)
				if len(properties) == 0 {
					template := checker.Checker_getTemplateTypeFromMappedType(tc, t)
					if template == nil {
						template = checker.Checker_getConstraintTypeFromMappedType(tc, t)
					}
					visitType(template, false, false)
				}
				// A key remapped through `as` is a signature position like any
				// other. A mapped type's name type is internal to the checker,
				// so upstream — which only has the public ts.Type surface —
				// never descends here and reads a type parameter spelled only
				// in an `as` clause as unused.
				visitType(checker.Checker_getNameTypeFromMappedType(tc, t), false, false)
			}

			// Every index signature instantiates its value type once per key,
			// so all of them count as multiple uses. The public ts.Type
			// surface only offers getStringIndexType/getNumberIndexType, so
			// upstream sees no use at all in a symbol-keyed or pattern-keyed
			// signature.
			for _, info := range checker.Checker_getIndexInfosOfType(tc, t) {
				visitType(info.ValueType(), true, false)
			}

			for _, sig := range utils.GetCallSignatures(tc, t) {
				functionLikeType = true
				visitSignature(sig)
			}
			for _, sig := range utils.GetConstructSignatures(tc, t) {
				functionLikeType = true
				visitSignature(sig)
			}

		case utils.IsTypeFlagSet(t, checker.TypeFlagsIndex):
			// `keyof T`.
			visitType(t.AsIndexType().Target(), assumeMultipleUses, false)

		case utils.IsTypeFlagSet(t, checker.TypeFlagsStringMapping):
			// `Uppercase<T>` / `Lowercase<T>` / `Capitalize<T>` / `Uncapitalize<T>`.
			visitType(t.AsStringMappingType().Target(), assumeMultipleUses, false)
		}
	}

	if node.Kind == ast.KindCallSignature || node.Kind == ast.KindConstructor {
		functionLikeType = true
		visitSignature(tc.GetSignatureFromDeclaration(node))
	}
	if !functionLikeType {
		visitType(tc.GetTypeAtLocation(node), false, false)
	}
}

// buildReplaceWithConstraintFixes builds the suggestion that replaces every
// use of the type parameter with its constraint (or "unknown" when it has
// none) and removes the type parameter from the declaration. Unlike
// detection, which only looks at the declaring signature, this replaces
// every reference reachable from the type parameter's declaration,
// including uses inside a function body.
func buildReplaceWithConstraintFixes(ctx rule.RuleContext, container *ast.Node, typeParamNode *ast.Node, typeParam *ast.TypeParameterDeclaration, sym *ast.Symbol) []rule.RuleFix {
	sourceFile := ctx.SourceFile

	constraintText := "unknown"
	var unwrappedConstraint *ast.Node
	if typeParam.Constraint != nil {
		unwrappedConstraint = ast.SkipTypeParentheses(typeParam.Constraint)
		if unwrappedConstraint.Kind != ast.KindAnyKeyword {
			constraintText = utils.TrimmedNodeText(sourceFile, unwrappedConstraint)
		}
	}
	fixes := make([]rule.RuleFix, 0, len(ctx.Refs.References(sym))+1)
	for _, refNode := range ctx.Refs.References(sym) {
		text := constraintText
		if constraintNeedsParentheses(refNode, unwrappedConstraint) {
			text = "(" + constraintText + ")"
		}
		fixes = append(fixes, rule.RuleFixReplace(sourceFile, refNode, text))
	}

	fixes = append(fixes, removeTypeParameterFix(sourceFile, container, typeParamNode))
	return fixes
}

// Binding levels of the TypeScript type grammar, loosest first. A constraint
// written at a level looser than its substitution site accepts has to be
// parenthesized or the surrounding operators re-associate: replacing T in
// `T[]` with `readonly string[]` must yield `(readonly string[])[]`, since
// `readonly string[][]` is an array of `string[]` instead.
const (
	precLoose        = iota // conditional, function and constructor types
	precUnion               // `A | B`
	precIntersection        // `A & B`
	precTypeOperator        // `keyof A`, `readonly A`, `unique symbol`
	precPostfix             // `A[]`, `A[B]`, and every primary type
)

func typePrecedence(node *ast.Node) int {
	switch node.Kind {
	case ast.KindConditionalType, ast.KindFunctionType, ast.KindConstructorType:
		return precLoose
	case ast.KindUnionType:
		return precUnion
	case ast.KindIntersectionType:
		return precIntersection
	case ast.KindTypeOperator:
		return precTypeOperator
	}
	return precPostfix
}

// constraintNeedsParentheses reports whether constraint, written in place of
// a reference to the type parameter, has to be wrapped to keep its meaning.
func constraintNeedsParentheses(refNode *ast.Node, constraint *ast.Node) bool {
	if constraint == nil {
		return false
	}
	reference := refNode.Parent
	if reference == nil {
		return false
	}
	site := reference.Parent
	if site == nil || site.Kind == ast.KindParenthesizedType {
		// The position already carries parentheses of its own.
		return false
	}
	if typePrecedence(constraint) < requiredPrecedence(site, reference) {
		return true
	}
	// Upstream wraps a union, intersection or conditional constraint in these
	// positions whether or not the grammar asks for it. Keep those parentheses
	// so the suggestion reads the same as ESLint's.
	switch constraint.Kind {
	case ast.KindUnionType, ast.KindIntersectionType, ast.KindConditionalType:
		switch site.Kind {
		case ast.KindArrayType, ast.KindIndexedAccessType, ast.KindIntersectionType, ast.KindUnionType:
			return true
		}
	}
	return false
}

// requiredPrecedence returns the tightest binding a type must have to sit at
// child's position within site without parentheses.
func requiredPrecedence(site *ast.Node, child *ast.Node) int {
	switch site.Kind {
	case ast.KindArrayType:
		return precPostfix
	case ast.KindIndexedAccessType:
		// Only the object half is a postfix operand; the index sits inside
		// brackets, which accept any type.
		if site.AsIndexedAccessTypeNode().ObjectType == child {
			return precPostfix
		}
	case ast.KindTypeOperator, ast.KindIntersectionType:
		return precTypeOperator
	case ast.KindUnionType:
		return precIntersection
	case ast.KindConditionalType:
		// The check and extends halves are parsed without conditional and
		// function types; the branches take a full type.
		conditional := site.AsConditionalTypeNode()
		if conditional.CheckType == child || conditional.ExtendsType == child {
			return precUnion
		}
	}
	return precLoose
}

// removeTypeParameterFix removes typeParamNode from container's `<...>`
// clause: the whole clause when it's the only type parameter, otherwise
// typeParamNode plus its separating comma.
func removeTypeParameterFix(sourceFile *ast.SourceFile, container *ast.Node, typeParamNode *ast.Node) rule.RuleFix {
	typeParams := container.TypeParameters()
	paramRange := utils.TrimNodeTextRange(sourceFile, typeParamNode)

	if len(typeParams) == 1 {
		list := container.TypeParameterList()
		start, end, ok := utils.RangeEnclosingDelimiters(sourceFile.Text(), list.Pos(), list.End(), '<', '>')
		if !ok {
			return rule.RuleFixRemoveRange(paramRange)
		}
		return rule.RuleFixRemoveRange(core.NewTextRange(start, end))
	}

	index := slices.Index(typeParams, typeParamNode)
	if index == 0 {
		commaEnd := findTokenEnd(sourceFile, paramRange.End(), ast.KindCommaToken)
		nextParamStart := utils.TrimNodeTextRange(sourceFile, typeParams[1]).Pos()
		nextStart := ecmascript.SkipLeadingWhitespace(sourceFile.Text(), commaEnd, nextParamStart)
		return rule.RuleFixRemoveRange(core.NewTextRange(paramRange.Pos(), nextStart))
	}

	commaStart := findTokenStart(sourceFile, typeParams[index-1].End(), ast.KindCommaToken)
	return rule.RuleFixRemoveRange(core.NewTextRange(commaStart, paramRange.End()))
}

func findTokenStart(sourceFile *ast.SourceFile, from int, kind ast.Kind) int {
	s := scanner.GetScannerForSourceFile(sourceFile, from)
	for s.Token() != kind && s.Token() != ast.KindEndOfFile {
		s.Scan()
	}
	return s.TokenStart()
}

func findTokenEnd(sourceFile *ast.SourceFile, from int, kind ast.Kind) int {
	s := scanner.GetScannerForSourceFile(sourceFile, from)
	for s.Token() != kind && s.Token() != ast.KindEndOfFile {
		s.Scan()
	}
	return s.TokenEnd()
}
