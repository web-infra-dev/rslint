package no_unnecessary_type_assertion

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_unnecessary_type_assertion.schema.json
var schemaJSON []byte

const nullableTypeFlags = checker.TypeFlagsAny |
	checker.TypeFlagsUnknown |
	checker.TypeFlagsNull |
	checker.TypeFlagsUndefined |
	checker.TypeFlagsVoid

func getUnionTypeFlags(t *checker.Type) checker.TypeFlags {
	var flags checker.TypeFlags
	for _, part := range utils.UnionTypeParts(t) {
		flags |= checker.Type_flags(part)
	}
	return flags
}

func buildContextuallyUnnecessaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "contextuallyUnnecessary",
		Description: "This assertion is unnecessary since the receiver accepts the original type of the expression.",
	}
}
func buildUnnecessaryAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryAssertion",
		Description: "This assertion is unnecessary since it does not change the type of the expression.",
	}
}

type NoUnnecessaryTypeAssertionOptions struct {
	// TODO(port): maybe typeOrValueSpecifier?
	TypesToIgnore []string
	// Whether to check const assertions on literal values
	// When true, reports cases like `const foo = 'bar' as const` where the assertion is unnecessary
	CheckLiteralConstAssertions bool
}

func parseOptions(options []any) NoUnnecessaryTypeAssertionOptions {
	opts := NoUnnecessaryTypeAssertionOptions{
		TypesToIgnore: []string{},
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if raw, ok := optsMap["typesToIgnore"].([]any); ok {
		opts.TypesToIgnore = utils.ToStringSlice(raw)
	}
	if value, ok := optsMap["checkLiteralConstAssertions"].(bool); ok {
		opts.CheckLiteralConstAssertions = value
	}
	return opts
}

func assertionArgumentIndex(callNode, assertion *ast.Node) (int, bool) {
	if callNode == nil || (!ast.IsCallExpression(callNode) && !ast.IsNewExpression(callNode)) {
		return 0, false
	}
	for index, argument := range callNode.Arguments() {
		if ast.SkipParentheses(argument) == assertion {
			return index, true
		}
	}
	return 0, false
}

func assertionWalkUpParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Parent != nil && ast.IsParenthesizedExpression(node.Parent) && node.Parent.Expression() == node {
		node = node.Parent
	}
	return node
}

func assertionIsStartOfArrowFunctionBody(sourceFile *ast.SourceFile, node *ast.Node) bool {
	nodeStart := utils.TrimNodeTextRange(sourceFile, node).Pos()
	for current := node; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if ast.IsParenthesizedExpression(parent) {
			return false
		}
		if ast.IsArrowFunction(parent) {
			return parent.AsArrowFunction().Body == current
		}
		if !ast.IsExpression(parent) || utils.TrimNodeTextRange(sourceFile, parent).Pos() != nodeStart {
			return false
		}
	}
	return false
}

func assertionTypeArguments(typeChecker *checker.Checker, t *checker.Type) []*checker.Type {
	if alias := checker.Type_alias(t); alias != nil && len(alias.TypeArguments()) > 0 {
		return alias.TypeArguments()
	}
	if utils.IsTypeReference(t) {
		return checker.Checker_getTypeArguments(typeChecker, t)
	}
	return nil
}

func assertionTypeContains(typeChecker *checker.Checker, root *checker.Type, predicate func(*checker.Type) bool) bool {
	if root == nil {
		return false
	}
	seen := make(map[*checker.Type]bool)
	stack := []*checker.Type{root}
	for len(stack) > 0 {
		t := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if t == nil || seen[t] {
			continue
		}
		seen[t] = true
		if predicate(t) {
			return true
		}
		if t.IsUnion() || t.IsIntersection() {
			stack = append(stack, t.Types()...)
			continue
		}
		stack = append(stack, assertionTypeArguments(typeChecker, t)...)
		for _, signature := range utils.GetCallSignatures(typeChecker, t) {
			stack = append(stack, checker.Checker_getReturnTypeOfSignature(typeChecker, signature))
			for _, parameter := range checker.Signature_parameters(signature) {
				stack = append(stack, checker.Checker_getTypeOfSymbol(typeChecker, parameter))
			}
		}
	}
	return false
}

type assertionSignatureIdentity struct {
	declaration *ast.Node
	signature   *checker.Signature
}

func assertionGetSignatureIdentity(signature *checker.Signature) assertionSignatureIdentity {
	if declaration := signature.Declaration(); declaration != nil {
		return assertionSignatureIdentity{declaration: declaration}
	}
	for signature.Target() != nil && signature.Target() != signature {
		signature = signature.Target()
		if declaration := signature.Declaration(); declaration != nil {
			return assertionSignatureIdentity{declaration: declaration}
		}
	}
	return assertionSignatureIdentity{signature: signature}
}

type assertionContainsAnyFrameKind uint8

const (
	assertionContainsAnyFrameType assertionContainsAnyFrameKind = iota
	assertionContainsAnyFrameSignature
	assertionContainsAnyFrameSignatureExit
)

type assertionContainsAnyFrame struct {
	kind      assertionContainsAnyFrameKind
	typeValue *checker.Type
	signature *checker.Signature
	identity  assertionSignatureIdentity
}

const assertionContainsAnySignatureLimit = 10_000

func assertionTypeContainsAny(typeChecker *checker.Checker, root *checker.Type) bool {
	if root == nil {
		return false
	}
	seen := make(map[*checker.Type]bool)
	activeSignatures := make(map[assertionSignatureIdentity]bool)
	stack := []assertionContainsAnyFrame{{
		kind:      assertionContainsAnyFrameType,
		typeValue: root,
	}}
	expandedSignatures := 0
	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch frame.kind {
		case assertionContainsAnyFrameSignatureExit:
			delete(activeSignatures, frame.identity)
		case assertionContainsAnyFrameSignature:
			identity := assertionGetSignatureIdentity(frame.signature)
			if activeSignatures[identity] {
				// Recursive generic signatures can keep producing fresh types. A
				// partial search cannot prove that a later instantiation excludes
				// any, so preserve the assertion instead of offering an unsafe fix.
				return true
			}
			expandedSignatures++
			if expandedSignatures > assertionContainsAnySignatureLimit {
				// Source-declared recursion is caught by the stable declaration
				// identity above. This is a final guard for synthetic signatures
				// that have neither a declaration nor a reusable target.
				return true
			}
			activeSignatures[identity] = true
			stack = append(stack, assertionContainsAnyFrame{
				kind:     assertionContainsAnyFrameSignatureExit,
				identity: identity,
			})
			parameters := checker.Signature_parameters(frame.signature)
			for index := len(parameters) - 1; index >= 0; index-- {
				stack = append(stack, assertionContainsAnyFrame{
					kind:      assertionContainsAnyFrameType,
					typeValue: checker.Checker_getTypeOfSymbol(typeChecker, parameters[index]),
				})
			}
			stack = append(stack, assertionContainsAnyFrame{
				kind:      assertionContainsAnyFrameType,
				typeValue: checker.Checker_getReturnTypeOfSignature(typeChecker, frame.signature),
			})
		case assertionContainsAnyFrameType:
			t := frame.typeValue
			if t == nil || seen[t] {
				continue
			}
			seen[t] = true
			if utils.IsTypeAnyType(t) {
				return true
			}
			if t.IsUnion() || t.IsIntersection() {
				types := t.Types()
				for index := len(types) - 1; index >= 0; index-- {
					stack = append(stack, assertionContainsAnyFrame{
						kind:      assertionContainsAnyFrameType,
						typeValue: types[index],
					})
				}
				continue
			}
			signatures := utils.GetCallSignatures(typeChecker, t)
			for index := len(signatures) - 1; index >= 0; index-- {
				stack = append(stack, assertionContainsAnyFrame{
					kind:      assertionContainsAnyFrameSignature,
					signature: signatures[index],
				})
			}
			typeArguments := assertionTypeArguments(typeChecker, t)
			for index := len(typeArguments) - 1; index >= 0; index-- {
				stack = append(stack, assertionContainsAnyFrame{
					kind:      assertionContainsAnyFrameType,
					typeValue: typeArguments[index],
				})
			}
		}
	}
	return false
}

func assertionTypeContainsTypeVariable(typeChecker *checker.Checker, root *checker.Type) bool {
	return assertionTypeContains(typeChecker, root, func(t *checker.Type) bool {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsTypeVariable|checker.TypeFlagsIndex)
	})
}

func assertionIsTypeLiteral(t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsLiteral|checker.TypeFlagsBooleanLiteral)
}

func assertionHasIndexSignature(typeChecker *checker.Checker, t *checker.Type) bool {
	for _, part := range utils.UnionTypeParts(t) {
		if len(checker.Checker_getIndexInfosOfType(typeChecker, part)) > 0 {
			return true
		}
	}
	return false
}

func assertionHasSameTypeArguments(typeChecker *checker.Checker, uncast, cast *checker.Type) bool {
	uncastArguments := assertionTypeArguments(typeChecker, uncast)
	castArguments := assertionTypeArguments(typeChecker, cast)
	if len(uncastArguments) != len(castArguments) {
		return false
	}
	for index, uncastArgument := range uncastArguments {
		if uncastArgument != castArguments[index] {
			return false
		}
	}
	return true
}

func assertionHasGenericSignature(typeChecker *checker.Checker, t *checker.Type, kind checker.SignatureKind) bool {
	for _, signature := range checker.Checker_getSignaturesOfType(typeChecker, t, kind) {
		if len(signature.TypeParameters()) > 0 {
			return true
		}
	}
	return false
}

func assertionGenericCall(ctx rule.RuleContext, node *ast.Node) *ast.Node {
	seenFunction := false
	for child := assertionWalkUpParentheses(node); child.Parent != nil; child = child.Parent {
		current := child.Parent
		switch current.Kind {
		case ast.KindFunctionDeclaration:
			return nil
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			// ESTree wraps method parameters and bodies in a function expression.
			// Computed names and decorators remain in the enclosing context.
			if child == current.Body() || ast.IsParameterDeclaration(child) {
				return nil
			}
		case ast.KindFunctionExpression, ast.KindArrowFunction:
			body := current.Body()
			if (body != nil && ast.IsBlock(body)) || seenFunction {
				return nil
			}
			seenFunction = true
		case ast.KindCallExpression, ast.KindNewExpression:
			if current.TypeArguments() != nil {
				continue
			}
			calleeType := ctx.TypeChecker.GetNonNullableType(ctx.TypeChecker.GetTypeAtLocation(current.Expression()))
			kind := checker.SignatureKindCall
			if ast.IsNewExpression(current) {
				kind = checker.SignatureKindConstruct
			}
			if assertionHasGenericSignature(ctx.TypeChecker, calleeType, kind) {
				return current
			}
		}
	}
	return nil
}

func assertionIsConceptuallyLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNumericLiteral,
		ast.KindStringLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindArrayLiteralExpression,
		ast.KindObjectLiteralExpression,
		ast.KindTemplateExpression,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindClassExpression,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindJsxElement,
		ast.KindJsxSelfClosingElement,
		ast.KindJsxFragment:
		return true
	}
	return false
}

func assertionOriginalExpression(node *ast.Node) *ast.Node {
	expression := ast.SkipParentheses(node.Expression())
	for expression != nil && (expression.Kind == ast.KindAsExpression || expression.Kind == ast.KindTypeAssertionExpression) {
		expression = ast.SkipParentheses(expression.Expression())
	}
	return expression
}

func assertionPreservesWidening(typeChecker *checker.Checker, original, asserted *checker.Type) bool {
	widenedOriginal := typeChecker.GetWidenedLiteralType(original)
	widenedAsserted := typeChecker.GetWidenedLiteralType(asserted)
	return (widenedOriginal == original && widenedAsserted == asserted) ||
		checker.Checker_isTypeIdenticalTo(typeChecker, widenedOriginal, widenedAsserted)
}

func assertionHasExplicitLiteralType(ctx rule.RuleContext, node *ast.Node, asserted *checker.Type) bool {
	if !assertionIsTypeLiteral(asserted) {
		return false
	}
	semanticNode := assertionWalkUpParentheses(node)
	parent := semanticNode.Parent
	if parent == nil || (!ast.IsVariableDeclaration(parent) && !ast.IsPropertyDeclaration(parent)) ||
		parent.Type() == nil || parent.Initializer() != semanticNode {
		return false
	}
	return checker.Checker_isTypeIdenticalTo(ctx.TypeChecker,
		checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, parent.Type()), asserted)
}

func assertionReturnMayAffectInference(ctx rule.RuleContext, expression *ast.Node) bool {
	if assertionMayAffectInference(ctx, expression) {
		return true
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil || ast.IsFunctionLike(node) || ast.IsAssertionExpression(node) ||
			ast.IsSatisfiesExpression(node) || ast.IsPartOfTypeNode(node) {
			return false
		}
		if ast.IsExpressionNode(node) && !ast.IsDeclarationName(node) {
			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone)
			if contextualType != nil && !assertionPreservesWidening(ctx.TypeChecker,
				checker.Checker_getTypeOfExpression(ctx.TypeChecker, ast.GetReparsedNodeForNode(node)), contextualType) {
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	return visit(expression)
}

func assertionFunctionMayAffectInference(ctx rule.RuleContext, expression *ast.Node, contextualParameters bool) bool {
	if contextualParameters && ast.HasContextSensitiveParameters(expression) {
		return true
	}
	if expression.Type() != nil {
		return false
	}
	// Both yielded values and the value of a yield expression (TNext)
	// can be inferred from the assertion's contextual generator type.
	if ast.GetFunctionFlags(expression)&ast.FunctionFlagsGenerator != 0 {
		return true
	}
	body := expression.Body()
	if body != nil && ast.IsBlock(body) {
		return ast.ForEachReturnStatement(body, func(statement *ast.Node) bool {
			return assertionReturnMayAffectInference(ctx, statement.Expression())
		})
	}
	return assertionReturnMayAffectInference(ctx, body)
}

// A call's type inside an assertion may already include return-type inferences
// supplied by that assertion. Comparing those types cannot prove that removing
// the assertion preserves the type inferred for the expression.
func assertionMayAffectInference(ctx rule.RuleContext, expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		// Const assertions inherit the surrounding contextual type, while
		// ordinary assertions supply their own context to the operand.
		if ast.IsConstAssertion(expression) {
			return assertionMayAffectInference(ctx, expression.Expression())
		}
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		if checker.Checker_getContextualType(ctx.TypeChecker, expression, checker.ContextFlagsNone) == nil {
			return false
		}
		var callee *ast.Node
		kind := checker.SignatureKindCall
		if expression.Kind == ast.KindTaggedTemplateExpression {
			callee = expression.AsTaggedTemplateExpression().Tag
		} else {
			callee = expression.Expression()
			if expression.Kind == ast.KindNewExpression {
				kind = checker.SignatureKindConstruct
			}
		}
		callee = ast.SkipParentheses(callee)
		// An IIFE's parameters are inferred from its arguments, while its
		// return expression can still depend on the assertion's context.
		if expression.Kind == ast.KindCallExpression && ast.IsFunctionExpressionOrArrowFunction(callee) &&
			assertionFunctionMayAffectInference(ctx, callee, false) {
			return true
		}
		// Explicit type arguments, including trailing defaults, fix this
		// call's type parameters, but not inference inside an IIFE's body.
		if len(expression.TypeArguments()) != 0 {
			return false
		}
		calleeType := ctx.TypeChecker.GetNonNullableType(ctx.TypeChecker.GetTypeAtLocation(callee))
		// Removing the contextual type can change overload selection, so the
		// currently resolved signature alone cannot establish safety.
		for _, signature := range checker.Checker_getSignaturesOfType(ctx.TypeChecker, calleeType, kind) {
			if len(signature.TypeParameters()) != 0 &&
				checker.Checker_couldContainTypeVariables(ctx.TypeChecker)(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)) {
				return true
			}
		}
	case ast.KindObjectLiteralExpression:
		for _, property := range expression.Properties() {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				if assertionMayAffectInference(ctx, property.Initializer()) {
					return true
				}
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				if assertionMayAffectInference(ctx, property) {
					return true
				}
			case ast.KindSpreadAssignment:
				if assertionMayAffectInference(ctx, property.Expression()) {
					return true
				}
			}
		}
	case ast.KindArrayLiteralExpression:
		arrayType := ctx.TypeChecker.GetTypeAtLocation(expression)
		// A mutable tuple inferred for an array literal depends on its
		// contextual type; removing that context can produce an array.
		if checker.IsTupleType(arrayType) && !arrayType.TargetTupleType().IsReadonly() {
			return true
		}
		for _, element := range expression.Elements() {
			if element.Kind == ast.KindSpreadElement {
				element = element.Expression()
			}
			if assertionMayAffectInference(ctx, element) {
				return true
			}
		}
	case ast.KindConditionalExpression:
		conditional := expression.AsConditionalExpression()
		return assertionMayAffectInference(ctx, conditional.WhenTrue) || assertionMayAffectInference(ctx, conditional.WhenFalse)
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return assertionMayAffectInference(ctx, binary.Left) || assertionMayAffectInference(ctx, binary.Right)
		case ast.KindAmpersandAmpersandToken, ast.KindCommaToken:
			return assertionMayAffectInference(ctx, binary.Right)
		}
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return assertionFunctionMayAffectInference(ctx, expression, true)
	case ast.KindAwaitExpression, ast.KindNonNullExpression:
		return assertionMayAffectInference(ctx, expression.Expression())
	}
	return false
}

func assertionIsInDestructuringDeclaration(node *ast.Node) bool {
	semanticNode := assertionWalkUpParentheses(node)
	parent := semanticNode.Parent
	if !ast.IsVariableDeclaration(parent) || parent.AsVariableDeclaration().Initializer != semanticNode {
		return false
	}
	name := parent.Name()
	return name != nil && (name.Kind == ast.KindObjectBindingPattern || name.Kind == ast.KindArrayBindingPattern)
}

func assertionIsPropertyInProblematicContext(ctx rule.RuleContext, node *ast.Node) bool {
	semanticNode := assertionWalkUpParentheses(node)
	property := semanticNode.Parent
	if !ast.IsPropertyAssignment(property) || property.AsPropertyAssignment().Initializer != semanticNode {
		return false
	}
	objectExpression := property.Parent
	if !ast.IsObjectLiteralExpression(objectExpression) {
		return false
	}
	objectContext := checker.Checker_getContextualType(ctx.TypeChecker, objectExpression, checker.ContextFlagsNone)
	if objectContext != nil && objectContext.IsUnion() {
		propertyContext := checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone)
		if propertyContext == nil {
			return true
		}
		propertyContext = ctx.TypeChecker.GetNonNullableType(propertyContext)
		if propertyContext.IsUnion() {
			return true
		}
		uncastType := ctx.TypeChecker.GetTypeAtLocation(node.Expression())
		return !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, uncastType, propertyContext)
	}
	objectParent := assertionWalkUpParentheses(objectExpression).Parent
	if objectParent == nil {
		return false
	}
	if objectParent.Kind == ast.KindSatisfiesExpression {
		return true
	}
	if !ast.IsCallExpression(objectParent) {
		return false
	}
	return assertionWalkUpParentheses(objectParent).Parent != nil &&
		assertionWalkUpParentheses(objectParent).Parent.Kind == ast.KindSatisfiesExpression
}

func assertionIsAssignmentInNonStatementContext(node *ast.Node) bool {
	semanticNode := assertionWalkUpParentheses(node)
	parent := semanticNode.Parent
	if !ast.IsBinaryExpression(parent) {
		return false
	}
	binary := parent.AsBinaryExpression()
	if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Right != semanticNode {
		return false
	}
	assignment := assertionWalkUpParentheses(parent)
	return assignment.Parent == nil || !ast.IsExpressionStatement(assignment.Parent)
}

func assertionIsRightHandSideOfLogicalAssignment(node *ast.Node) bool {
	semanticNode := assertionWalkUpParentheses(node)
	parent := semanticNode.Parent
	if !ast.IsBinaryExpression(parent) {
		return false
	}
	binary := parent.AsBinaryExpression()
	return binary.Right == semanticNode && binary.OperatorToken != nil &&
		ast.IsLogicalOrCoalescingAssignmentOperator(binary.OperatorToken.Kind)
}

func assertionContextualType(ctx rule.RuleContext, node *ast.Node) *checker.Type {
	if contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone); contextualType != nil {
		return contextualType
	}
	semanticNode := assertionWalkUpParentheses(node)
	parent := semanticNode.Parent
	if ast.IsParameterDeclaration(parent) || ast.IsPropertyDeclaration(parent) {
		if contextualType := utils.GetContextualType(ctx.TypeChecker, node); contextualType != nil {
			return contextualType
		}
	}
	return nil
}

func assertionIsArgumentToOverloadedFunction(ctx rule.RuleContext, node *ast.Node) bool {
	parent, argumentIndex, _ := assertionCallArgument(node)
	if parent == nil {
		return false
	}
	calleeType := ctx.TypeChecker.GetNonNullableType(ctx.TypeChecker.GetTypeAtLocation(parent.Expression()))
	kind := checker.SignatureKindCall
	if ast.IsNewExpression(parent) {
		kind = checker.SignatureKindConstruct
	}
	signatures := checker.Checker_getSignaturesOfType(ctx.TypeChecker, calleeType, kind)
	if len(signatures) <= 1 {
		return false
	}
	parameterTypes := make([]*checker.Type, 0, len(signatures))
	for _, signature := range signatures {
		parameters := checker.Signature_parameters(signature)
		if argumentIndex >= len(parameters) {
			return true
		}
		parameter := parameters[argumentIndex]
		parameterType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameter)
		if parameter.ValueDeclaration != nil && ast.IsParameterDeclaration(parameter.ValueDeclaration) &&
			parameter.ValueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
			if typeArguments := assertionTypeArguments(ctx.TypeChecker, parameterType); len(typeArguments) > 0 {
				parameterType = typeArguments[0]
			}
		}
		parameterTypes = append(parameterTypes, parameterType)
	}
	first := parameterTypes[0]
	allSame := true
	for _, parameterType := range parameterTypes[1:] {
		if !checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, parameterType, first) {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}
	// Selecting a different overload is harmless only when this is the
	// sole argument and all non-generic candidates accept it and expose
	// the same result. Type predicates can affect later control flow.
	if len(parent.Arguments()) != 1 || ast.SkipParentheses(parent.Arguments()[0]) != node {
		return true
	}
	uncastType := ctx.TypeChecker.GetTypeAtLocation(assertionOriginalExpression(node))
	returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signatures[0])
	for index, signature := range signatures {
		if len(signature.TypeParameters()) != 0 || signature.ThisParameter() != nil ||
			ctx.TypeChecker.GetTypePredicateOfSignature(signature) != nil ||
			!checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, returnType, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)) ||
			!checker.Checker_isTypeAssignableTo(ctx.TypeChecker, uncastType, parameterTypes[index]) {
			return true
		}
	}
	return false
}

func assertionGenericsMismatch(ctx rule.RuleContext, uncastType, contextualType *checker.Type) bool {
	for _, property := range ctx.TypeChecker.GetPropertiesOfType(contextualType) {
		propertyType := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property)
		if !assertionHasGenericSignature(ctx.TypeChecker, propertyType, checker.SignatureKindCall) {
			continue
		}
		uncastProperty := ctx.TypeChecker.GetPropertyOfType(uncastType, property.Name)
		if uncastProperty == nil || !assertionHasGenericSignature(
			ctx.TypeChecker,
			checker.Checker_getTypeOfSymbol(ctx.TypeChecker, uncastProperty),
			checker.SignatureKindCall,
		) {
			return true
		}
	}
	return false
}

func assertionIsEmptyObjectType(typeChecker *checker.Checker, t *checker.Type) bool {
	return utils.IsTypeFlagSet(t, checker.TypeFlagsNonPrimitive) ||
		(len(typeChecker.GetPropertiesOfType(t)) == 0 &&
			len(utils.GetCallSignatures(typeChecker, t)) == 0 &&
			len(utils.GetConstructSignatures(typeChecker, t)) == 0 &&
			typeChecker.GetStringIndexType(t) == nil &&
			typeChecker.GetNumberIndexType(t) == nil)
}

func assertionHasPhantomTypeArguments(typeChecker *checker.Checker, t *checker.Type) bool {
	return assertionIsEmptyObjectType(typeChecker, t) && len(assertionTypeArguments(typeChecker, t)) > 0
}

func assertionHasPhantomTypeArgumentMismatch(ctx rule.RuleContext, node *ast.Node, uncastType, contextualType *checker.Type) bool {
	if assertionGenericCall(ctx, node) == nil ||
		(!assertionHasPhantomTypeArguments(ctx.TypeChecker, uncastType) &&
			!assertionHasPhantomTypeArguments(ctx.TypeChecker, contextualType)) {
		return false
	}
	uncastArguments := assertionTypeArguments(ctx.TypeChecker, uncastType)
	contextualArguments := assertionTypeArguments(ctx.TypeChecker, contextualType)
	if len(uncastArguments) != len(contextualArguments) {
		return true
	}
	for index, uncastArgument := range uncastArguments {
		if uncastArgument != contextualArguments[index] {
			return true
		}
	}
	return false
}

// Find the nearest argument that can receive contextual types through these
// containers. The final result also identifies a sole inference path, without
// sibling properties, elements, arguments, or callback parameters.
func assertionCallArgument(node *ast.Node) (*ast.Node, int, bool) {
	independent := true
	for current := node; ; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			return nil, 0, false
		}
		switch parent.Kind {
		case ast.KindCallExpression, ast.KindNewExpression:
			for index, argument := range parent.Arguments() {
				if current == argument {
					return parent, index, independent && len(parent.Arguments()) == 1
				}
			}
			return nil, 0, false
		case ast.KindParenthesizedExpression:
		case ast.KindPropertyAssignment:
			if parent.Initializer() != current || ast.IsComputedPropertyName(parent.Name()) {
				return nil, 0, false
			}
		case ast.KindObjectLiteralExpression:
			independent = independent && len(parent.Properties()) == 1
		case ast.KindArrayLiteralExpression:
			independent = independent && len(parent.Elements()) == 1
		case ast.KindArrowFunction:
			if parent.Body() != current {
				return nil, 0, false
			}
			independent = independent && len(parent.Parameters()) == 0
		default:
			return nil, 0, false
		}
	}
}

func shouldSkipAssertionContextualTypeFallback(ctx rule.RuleContext, node *ast.Node, castIsAny bool) bool {
	semanticNode := assertionWalkUpParentheses(node)
	parent := semanticNode.Parent
	if parent == nil {
		return true
	}
	if castIsAny {
		return ast.IsLogicalOrCoalescingBinaryExpression(parent) || assertionGenericCall(ctx, node) != nil
	}
	if ast.IsTemplateExpression(ast.SkipParentheses(node.Expression())) {
		return true
	}
	switch parent.Kind {
	case ast.KindAsExpression,
		ast.KindTypeAssertionExpression,
		ast.KindSpreadElement,
		ast.KindSpreadAssignment,
		ast.KindSatisfiesExpression:
		return true
	}
	if ast.IsArrayLiteralExpression(ast.SkipParentheses(node.Expression())) ||
		assertionIsInDestructuringDeclaration(node) ||
		assertionIsPropertyInProblematicContext(ctx, node) ||
		assertionIsAssignmentInNonStatementContext(node) ||
		assertionIsRightHandSideOfLogicalAssignment(node) ||
		assertionIsArgumentToOverloadedFunction(ctx, node) {
		return true
	}
	if genericCall := assertionGenericCall(ctx, node); genericCall != nil {
		original := assertionOriginalExpression(node)
		argumentCall, _, independent := assertionCallArgument(node)
		if utils.IsTypeFlagSetWithUnion(ctx.TypeChecker.GetTypeAtLocation(original), checker.TypeFlagsEnumLike) ||
			!ast.IsExpressionStatement(assertionWalkUpParentheses(genericCall).Parent) ||
			argumentCall != genericCall || !independent {
			return true
		}
		return !assertionIsConceptuallyLiteral(original) && !ast.IsPropertyAssignment(parent)
	}
	return false
}

var NoUnnecessaryTypeAssertionRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-type-assertion",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		sourceText := ctx.SourceFile.Text()
		// Match TrimNodeTextRange's scanner setup while sharing the scanner
		// across diagnostic ranges and synchronously materialized fixes.
		var tokenScanner *scanner.Scanner
		getTokenRange := func(pos int) core.TextRange {
			if tokenScanner == nil {
				tokenScanner = scanner.NewScanner()
			} else {
				tokenScanner.Reset()
			}
			tokenScanner.SetText(sourceText)
			tokenScanner.SetLanguageVariant(ctx.SourceFile.LanguageVariant)
			tokenScanner.ResetPos(pos)
			tokenScanner.Scan()
			return tokenScanner.TokenRange()
		}

		compilerOptions := ctx.Program().Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		/**
		 * Returns true if there's a chance the variable has been used before a value has been assigned to it
		 */
		isPossiblyUsedBeforeAssigned := func(
			node *ast.Node,
			declaration *ast.Declaration,
			constrainedType *checker.Type,
		) bool {
			if declaration == nil {
				// don't know what the declaration is for some reason, so just assume the worst
				return true
			}
			// non-strict mode doesn't care about used before assigned errors
			if !isStrictNullChecks {
				return false
			}
			// ignore class properties as they are compile time guarded
			// also ignore function arguments as they can't be used before defined
			if !ast.IsVariableDeclaration(declaration) {
				return false
			}

			decl := declaration.AsVariableDeclaration()

			// For var declarations, we need to check whether the node
			// is actually in a descendant of its declaration or not. If not,
			// it may be used before defined.

			// eg
			// if (Math.random() < 0.5) {
			//     var x: number  = 2;
			// } else {
			//     x!.toFixed();
			// }
			if ast.IsVariableDeclarationList(declaration.Parent) &&
				// var
				declaration.Parent.Flags == ast.NodeFlagsNone {
				// If they are not in the same file it will not exist.
				// This situation must not occur using before defined.
				declaratorScope := ast.GetEnclosingBlockScopeContainer(declaration)
				scope := ast.GetEnclosingBlockScopeContainer(node)

				parentScope := declaratorScope
				for {
					parentScope = ast.GetEnclosingBlockScopeContainer(parentScope)
					if parentScope == nil {
						break
					}
					if parentScope == scope {
						return true
					}
				}
			}

			if
			// is it `const x: number`
			decl.Initializer == nil &&
				decl.ExclamationToken == nil &&
				decl.Type != nil {
				// check if the defined variable type has changed since assignment
				declarationType := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, declaration.Type())
				if constrainedType == nil {
					constrainedType = utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
				}
				if declarationType == constrainedType &&
					// `declare`s are never narrowed, so never skip them
					(!ast.IsVariableDeclarationList(declaration.Parent) || !ast.IsVariableStatement(declaration.Parent.Parent) || !utils.IncludesModifier(declaration.Parent.Parent.AsVariableStatement(), ast.KindDeclareKeyword)) {
					// possibly used before assigned, so just skip it
					// better to false negative and skip it, than false positive and fix to compile erroring code
					//
					// no better way to figure this out right now
					// https://github.com/Microsoft/TypeScript/issues/31124
					return true
				}
			}

			return false
		}

		type identifierInfo struct {
			declaration     *ast.Declaration
			constrainedType *checker.Type
			typeFlags       checker.TypeFlags
		}
		identifierInfoCache := make(map[*ast.Symbol]identifierInfo)
		getIdentifierInfo := func(node *ast.Node) (identifierInfo, bool) {
			symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
			if symbol == nil {
				return identifierInfo{}, false
			}
			if info, ok := identifierInfoCache[symbol]; ok {
				return info, true
			}

			info := identifierInfo{}
			if len(symbol.Declarations) > 0 {
				info.declaration = symbol.Declarations[0]
			}
			declaredType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(symbol, nil)
			if declaredType != nil {
				info.constrainedType = declaredType
				if constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, declaredType); constraint != nil {
					info.constrainedType = constraint
				}
				info.typeFlags = getUnionTypeFlags(info.constrainedType)
			}
			identifierInfoCache[symbol] = info
			return info, true
		}

		isConstAssertion := func(node *ast.Node) bool {
			if !ast.IsTypeReferenceNode(node) {
				return false
			}
			typeName := node.AsTypeReferenceNode().TypeName
			return ast.IsIdentifier(typeName) && typeName.Text() == "const"
		}

		isImplicitlyNarrowedLiteralDeclaration := func(node *ast.Node) bool {
			expression := node.Expression()
			/**
			 * Even on `const` variable declarations, template literals with expressions can sometimes be widened without a type assertion.
			 * @see https://github.com/typescript-eslint/typescript-eslint/issues/8737
			 */
			if ast.IsTemplateExpression(expression) {
				return false
			}

			semanticNode := assertionWalkUpParentheses(node)
			parent := semanticNode.Parent
			return (ast.IsVariableDeclaration(parent) && ast.IsVariableDeclarationList(parent.Parent) && parent.Parent.Flags&ast.NodeFlagsConst != 0) ||
				(ast.IsPropertyDeclaration(parent) && parent.ModifierFlags()&ast.ModifierFlagsReadonly != 0)

		}

		areUnionPartsEquivalentIgnoringUndefined := func(uncast, cast *checker.Type) bool {
			uncastParts := utils.Set[*checker.Type]{}
			for _, part := range utils.UnionTypeParts(uncast) {
				if !utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					uncastParts.Add(part)
				}
			}
			castPartsCount := 0
			for _, part := range utils.UnionTypeParts(cast) {
				if !utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					if !uncastParts.Has(part) {
						return false
					}
					castPartsCount++
				}
			}
			return uncastParts.Len() == castPartsCount
		}

		isTypeUnchanged := func(node, expression *ast.Node, uncast, cast *checker.Type) bool {
			if assertionHasExplicitLiteralType(ctx, node, cast) {
				return checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, uncast, cast)
			}
			// Fresh literals can be identical to an asserted regular type
			// while widening differently in later assignments or arrays.
			if !assertionPreservesWidening(ctx.TypeChecker, uncast, cast) {
				return false
			}
			if uncast == cast {
				return true
			}

			if node.Type().Kind == ast.KindIntersectionType && assertionTypeContainsTypeVariable(ctx.TypeChecker, cast) {
				return false
			}

			if getUnionTypeFlags(uncast)&checker.TypeFlagsUndefined != 0 &&
				getUnionTypeFlags(cast)&checker.TypeFlagsUndefined != 0 &&
				compilerOptions.ExactOptionalPropertyTypes.IsTrue() {
				return areUnionPartsEquivalentIgnoringUndefined(uncast, cast)
			}

			if (utils.IsTypeFlagSetWithUnion(uncast, checker.TypeFlagsNonPrimitive) && !utils.IsTypeFlagSetWithUnion(cast, checker.TypeFlagsNonPrimitive)) ||
				assertionHasIndexSignature(ctx.TypeChecker, uncast) != assertionHasIndexSignature(ctx.TypeChecker, cast) ||
				assertionTypeContainsAny(ctx.TypeChecker, uncast) ||
				assertionTypeContainsAny(ctx.TypeChecker, cast) ||
				(assertionTypeContainsTypeVariable(ctx.TypeChecker, cast) && !assertionTypeContainsTypeVariable(ctx.TypeChecker, uncast)) {
				return false
			}

			expression = ast.SkipParentheses(expression)
			if assertionIsConceptuallyLiteral(expression) {
				shouldPreserveAssertion := !ast.IsObjectLiteralExpression(expression) || len(expression.Properties()) == 0
				if !shouldPreserveAssertion {
					for _, property := range ctx.TypeChecker.GetPropertiesOfType(cast) {
						if assertionIsTypeLiteral(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property)) {
							shouldPreserveAssertion = true
							break
						}
					}
				}
				if shouldPreserveAssertion {
					return false
				}
			}

			if cast.IsIntersection() && !uncast.IsIntersection() {
				castParts := cast.Types()
				if utils.IsTypeParameter(uncast) && len(castParts) == 2 {
					var otherPart *checker.Type
					if castParts[0] == uncast {
						otherPart = castParts[1]
					} else if castParts[1] == uncast {
						otherPart = castParts[0]
					}
					if otherPart != nil && assertionIsEmptyObjectType(ctx.TypeChecker, otherPart) &&
						!assertionTypeContainsTypeVariable(ctx.TypeChecker, otherPart) {
						constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, uncast)
						if constraint != nil && !utils.IsNullableType(constraint) {
							return true
						}
					}
				}
				return false
			}

			return assertionHasSameTypeArguments(ctx.TypeChecker, uncast, cast) &&
				checker.Checker_isTypeIdenticalTo(ctx.TypeChecker, uncast, cast)
		}

		getUncastType := func(node *ast.Node) *checker.Type {
			expression := node.Expression()
			if ast.IsCallExpression(expression) {
				callee := ast.SkipParentheses(expression.Expression())
				if ast.IsArrowFunction(callee) || ast.IsFunctionExpression(callee) {
					functionType := ctx.TypeChecker.GetTypeAtLocation(callee)
					signatures := utils.GetCallSignatures(ctx.TypeChecker, functionType)
					if len(signatures) > 0 {
						returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signatures[0])
						if callee.Type() == nil && utils.IsTypeFlagSet(returnType, checker.TypeFlagsUndefined) {
							return ctx.TypeChecker.GetVoidType()
						}
						return returnType
					}
				}
			}
			// GetTypeAtLocation regularizes literal types. Keep the expression's
			// actual freshness so later widening is part of the comparison.
			return checker.Checker_getTypeOfExpression(ctx.TypeChecker, ast.GetReparsedNodeForNode(expression))
		}

		checkTypeAssertion := func(node *ast.Node) {
			typeNode := node.Type()
			if len(opts.TypesToIgnore) > 0 {
				ignoredType := typeNode
				// ESTree omits parentheses around the type annotation's text.
				if ignoredType.Kind == ast.KindParenthesizedType {
					ignoredType = ast.SkipTypeParentheses(ignoredType)
				}
				if slices.Contains(opts.TypesToIgnore, utils.TrimmedNodeText(ctx.SourceFile, ignoredType)) {
					return
				}
			}

			castType := ctx.TypeChecker.GetTypeAtLocation(node)
			castTypeIsLiteral := assertionIsTypeLiteral(castType)
			typeAnnotationIsConstAssertion := isConstAssertion(typeNode)
			if !opts.CheckLiteralConstAssertions && castTypeIsLiteral && typeAnnotationIsConstAssertion {
				return
			}

			expression := node.Expression()
			uncastType := getUncastType(node)
			reportAssertion := func(msg rule.RuleMessage) {
				if assertionMayAffectInference(ctx, expression) ||
					(!assertionPreservesWidening(ctx.TypeChecker, uncastType, castType) &&
						checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone) == nil) {
					return
				}
				ctx.ReportRangeWithDeferredFixes(getTokenRange(node.Pos()).WithEnd(node.End()), msg, func() []rule.RuleFix {
					if node.Kind == ast.KindAsExpression {
						asKeywordRange := getTokenRange(expression.End())
						startPos := ecmascript.SkipTrailingWhitespace(sourceText, expression.End(), asKeywordRange.Pos())

						fixRange := asKeywordRange.WithPos(startPos).WithEnd(typeNode.End())
						return []rule.RuleFix{rule.RuleFixRemoveRange(fixRange)}
					}

					openingAngleBracket := getTokenRange(node.Pos())
					fixRange := openingAngleBracket.WithEnd(expression.Pos())
					fixes := []rule.RuleFix{rule.RuleFixRemoveRange(fixRange)}
					firstOperandRange := getTokenRange(expression.Pos())
					firstOperandText := sourceText[firstOperandRange.Pos():firstOperandRange.End()]
					breaksExpressionStatement := (firstOperandText == "{" || firstOperandText == "function" || firstOperandText == "class") &&
						utils.IsStartOfExpressionStatement(ctx.SourceFile, node)
					breaksArrowFunctionBody := firstOperandText == "{" && assertionIsStartOfArrowFunctionBody(ctx.SourceFile, node)
					if breaksExpressionStatement || breaksArrowFunctionBody {
						fixes = append(fixes,
							rule.RuleFixInsertBefore(ctx.SourceFile, node, "("),
							rule.RuleFixInsertAfter(node, ")"),
						)
					}
					return fixes
				})
			}
			reportDoubleAssertionIfUnnecessary := func(contextualType *checker.Type) {
				innerExpression := ast.SkipParentheses(expression)
				if innerExpression == nil || (innerExpression.Kind != ast.KindAsExpression && innerExpression.Kind != ast.KindTypeAssertionExpression) {
					return
				}
				originalExpression := assertionOriginalExpression(node)
				originalType := checker.Checker_getTypeOfExpression(ctx.TypeChecker, ast.GetReparsedNodeForNode(originalExpression))
				msg := rule.RuleMessage{}
				if isTypeUnchanged(node, innerExpression, originalType, castType) && !utils.IsTypeAnyType(castType) {
					msg = buildUnnecessaryAssertionMessage()
				} else if contextualType != nil {
					intermediateType := ctx.TypeChecker.GetTypeAtLocation(innerExpression)
					if utils.IsTypeFlagSet(intermediateType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) &&
						checker.Checker_isTypeAssignableTo(ctx.TypeChecker, originalType, contextualType) {
						msg = buildContextuallyUnnecessaryMessage()
					}
				}
				if msg.Id == "" || assertionMayAffectInference(ctx, originalExpression) ||
					(!assertionPreservesWidening(ctx.TypeChecker, originalType, castType) &&
						checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone) == nil) {
					return
				}
				ctx.ReportRangeWithDeferredFixes(getTokenRange(node.Pos()).WithEnd(node.End()), msg, func() []rule.RuleFix {
					originalRange := utils.TrimNodeTextRange(ctx.SourceFile, originalExpression)
					replacement := sourceText[originalRange.Pos():originalRange.End()]
					semanticNode := assertionWalkUpParentheses(node)
					if originalExpression.Kind == ast.KindObjectLiteralExpression &&
						semanticNode.Parent != nil && ast.IsArrowFunction(semanticNode.Parent) &&
						semanticNode.Parent.AsArrowFunction().Body == semanticNode {
						replacement = "(" + replacement + ")"
					}
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)}
				})
			}

			typeIsUnchanged := isTypeUnchanged(node, expression, uncastType, castType)
			wouldSameTypeBeInferred := !typeAnnotationIsConstAssertion
			if castTypeIsLiteral {
				wouldSameTypeBeInferred = assertionHasExplicitLiteralType(ctx, node, castType) || isImplicitlyNarrowedLiteralDeclaration(node)
			}
			if typeIsUnchanged && wouldSameTypeBeInferred {
				reportAssertion(buildUnnecessaryAssertionMessage())
				return
			}

			semanticNode := assertionWalkUpParentheses(node)
			parent := semanticNode.Parent
			parentSkipsAnyFallback := parent != nil && (parent.Kind == ast.KindAsExpression ||
				parent.Kind == ast.KindTypeAssertionExpression ||
				parent.Kind == ast.KindSpreadElement ||
				parent.Kind == ast.KindSpreadAssignment ||
				parent.Kind == ast.KindSatisfiesExpression)
			castIsAny := utils.IsTypeFlagSet(castType, checker.TypeFlagsAny) && !parentSkipsAnyFallback
			var contextualType *checker.Type
			if !shouldSkipAssertionContextualTypeFallback(ctx, node, castIsAny) {
				contextualType = assertionContextualType(ctx, node)
			}
			if contextualType != nil {
				contextualTypeIsAny := utils.IsTypeAnyType(contextualType)
				_, isCallArgument := assertionArgumentIndex(parent, node)
				anyInvolvedInContextualCheck := !assertionTypeContainsAny(ctx.TypeChecker, contextualType)
				if contextualTypeIsAny {
					anyInvolvedInContextualCheck = isCallArgument && !assertionTypeContainsAny(ctx.TypeChecker, castType)
				}

				originalExpression := ast.SkipParentheses(expression)
				isNullishLiteralToUnion := castType.IsUnion() && (originalExpression.Kind == ast.KindNullKeyword ||
					(ast.IsIdentifier(originalExpression) && originalExpression.Text() == "undefined"))
				if !typeAnnotationIsConstAssertion &&
					!assertionTypeContainsAny(ctx.TypeChecker, uncastType) &&
					anyInvolvedInContextualCheck &&
					!assertionHasPhantomTypeArgumentMismatch(ctx, node, uncastType, contextualType) &&
					(castIsAny || !assertionGenericsMismatch(ctx, uncastType, contextualType)) &&
					(contextualTypeIsAny || checker.Checker_isTypeAssignableTo(ctx.TypeChecker, uncastType, contextualType)) &&
					!isNullishLiteralToUnion {
					reportAssertion(buildContextuallyUnnecessaryMessage())
					return
				}
			}
			reportDoubleAssertionIfUnnecessary(contextualType)
		}

		return rule.RuleListeners{
			ast.KindAsExpression:            checkTypeAssertion,
			ast.KindTypeAssertionExpression: checkTypeAssertion,

			ast.KindNonNullExpression: func(node *ast.Node) {
				expression := node.Expression()

				buildRemoveExclamationFix := func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixRemoveRange(getTokenRange(expression.End()))}
				}

				if ast.IsAssignmentExpression(node.Parent, true) {
					if node.Parent.AsBinaryExpression().Left == node {
						ctx.ReportRangeWithDeferredFixes(getTokenRange(node.Pos()).WithEnd(node.End()), buildContextuallyUnnecessaryMessage(), buildRemoveExclamationFix)
					}
					// for all other = assignments we ignore non-null checks
					// this is because non-null assertions can change the type-flow of the code
					// so whilst they might be unnecessary for the assignment - they are necessary
					// for following code
					return
				}

				var (
					expressionIdentifierInfo identifierInfo
					hasIdentifierInfo        bool
				)
				if ast.IsIdentifier(expression) {
					expressionIdentifierInfo, hasIdentifierInfo = getIdentifierInfo(expression)
					// The checker assumes an identifier below a non-null expression is
					// initialized, and flow analysis can only preserve or narrow its
					// declared constrained type. A declared non-nullish type therefore
					// cannot become nullish at this location.
					if hasIdentifierInfo &&
						expressionIdentifierInfo.declaration != nil &&
						expressionIdentifierInfo.constrainedType != nil &&
						expressionIdentifierInfo.typeFlags&nullableTypeFlags == 0 {
						if isPossiblyUsedBeforeAssigned(expression, expressionIdentifierInfo.declaration, nil) {
							return
						}
						ctx.ReportRangeWithDeferredFixes(getTokenRange(node.Pos()).WithEnd(node.End()), buildUnnecessaryAssertionMessage(), buildRemoveExclamationFix)
						return
					}
				}

				constrainedType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)
				actualType := ctx.TypeChecker.GetTypeAtLocation(expression)
				constrainedTypeIsNullable := utils.IsNullableType(constrainedType)
				actualTypeIsNullable := utils.IsNullableType(actualType)

				if !constrainedTypeIsNullable && !actualTypeIsNullable {
					if ast.IsIdentifier(expression) {
						declaration := expressionIdentifierInfo.declaration
						if !hasIdentifierInfo {
							declaration = utils.GetDeclaration(ctx.TypeChecker, expression)
						}
						if isPossiblyUsedBeforeAssigned(expression, declaration, constrainedType) {
							return
						}
					}
					ctx.ReportRangeWithDeferredFixes(getTokenRange(node.Pos()).WithEnd(node.End()), buildUnnecessaryAssertionMessage(), buildRemoveExclamationFix)
				} else {
					// we know it's a nullable type
					// so figure out if the variable is used in a place that accepts nullable types
					if constrainedType != actualType {
						return
					}
					contextualType := utils.GetContextualType(ctx.TypeChecker, node)
					if contextualType != nil {
						tFlags := getUnionTypeFlags(constrainedType)
						contextualFlags := getUnionTypeFlags(contextualType)

						if tFlags&checker.TypeFlagsUnknown != 0 && contextualFlags&checker.TypeFlagsUnknown == 0 {
							return
						}

						// in strict mode you can't assign null to undefined, so we have to make sure that
						// the two types share a nullable type
						typeIncludesUndefined := tFlags&checker.TypeFlagsUndefined != 0
						typeIncludesNull := tFlags&checker.TypeFlagsNull != 0
						typeIncludesVoid := tFlags&checker.TypeFlagsVoid != 0

						contextualTypeIncludesUndefined := contextualFlags&checker.TypeFlagsUndefined != 0
						contextualTypeIncludesNull := contextualFlags&checker.TypeFlagsNull != 0
						contextualTypeIncludesVoid := contextualFlags&checker.TypeFlagsVoid != 0

						// make sure that the parent accepts the same types
						// i.e. assigning `string | null | undefined` to `string | undefined` is invalid
						isValidUndefined := !typeIncludesUndefined || contextualTypeIncludesUndefined
						isValidNull := !typeIncludesNull || contextualTypeIncludesNull
						isValidVoid := !typeIncludesVoid || contextualTypeIncludesVoid

						if isValidUndefined && isValidNull && isValidVoid {
							ctx.ReportRangeWithDeferredFixes(getTokenRange(node.Pos()).WithEnd(node.End()), buildContextuallyUnnecessaryMessage(), buildRemoveExclamationFix)
						}
					}
				}
			},
		}
	},
})
