package no_unnecessary_type_assertion

import (
	_ "embed"
	"slices"
	"strconv"

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

func assertionExpressionType(typeChecker *checker.Checker, expression *ast.Node) *checker.Type {
	// GetTypeAtLocation regularizes literal types. Preserve freshness when
	// comparing the type inferred after removing an assertion.
	return checker.Checker_getTypeOfExpression(typeChecker, ast.GetReparsedNodeForNode(expression))
}

func assertionPreservesWidening(typeChecker *checker.Checker, original, asserted *checker.Type) bool {
	// Mutable locations and implicit returns also widen unique symbols,
	// independently of the fresh-literal metadata.
	widenedOriginal := checker.Checker_getWidenedLiteralLikeTypeForContextualType(typeChecker, original, nil)
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

// This is deliberately a one-level check. An inferred aggregate carried by a
// reference is retained when its nested member writes cannot be inspected.
func assertionAggregateMayWiden(typeChecker *checker.Checker, t *checker.Type, opaque bool) bool {
	memberMayWiden := func(member *checker.Type) bool {
		return member != nil && (typeChecker.GetWidenedLiteralType(member) != member ||
			opaque && checker.Type_objectFlags(member)&checker.ObjectFlagsContainsObjectOrArrayLiteral != 0)
	}
	if checker.Checker_isArrayOrTupleType(typeChecker, t) {
		return memberMayWiden(checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker)))
	}
	if utils.IsTypeFlagSetWithUnion(t, checker.TypeFlagsObject) {
		for _, property := range typeChecker.GetPropertiesOfType(t) {
			if memberMayWiden(checker.Checker_getTypeOfSymbol(typeChecker, property)) {
				return true
			}
		}
	}
	return false
}

func assertionMemberWriteMayWiden(ctx rule.RuleContext, value *ast.Node, original, written *checker.Type, retainedConst bool) bool {
	return written == nil || ((!retainedConst || !checker.Checker_isConstContext(ctx.TypeChecker, value)) &&
		!assertionPreservesWidening(ctx.TypeChecker, original, written))
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
			return assertionValueMayAffectInference(ctx, statement.Expression(), false, true)
		})
	}
	return assertionValueMayAffectInference(ctx, body, false, true)
}

// A call's type inside an assertion may already include return-type inferences
// supplied by that assertion. Comparing those types cannot prove that removing
// the assertion preserves the type inferred for the expression.
// Member writes also need a widening proof unless the receiving context has
// already been proved stable. retainedConst only describes an inner assertion
// that remains after the proposed fix.
func assertionValueMayAffectInference(ctx rule.RuleContext, expression *ast.Node, retainedConst, memberWrites bool) bool {
	if expression == nil {
		return false
	}
	expression = ast.SkipParentheses(expression)
	var expressionType *checker.Type
	explicitLiteral := false
	if memberWrites {
		expressionType = assertionExpressionType(ctx.TypeChecker, expression)
		contextual := checker.Checker_getContextualType(ctx.TypeChecker, expression, checker.ContextFlagsNone)
		parent := assertionWalkUpParentheses(expression).Parent
		// A declaration that explicitly fixes this literal keeps it pinned;
		// its operand still needs the inference checks below.
		explicitLiteral = contextual != nil && parent != nil && ast.IsAssertionExpression(parent) &&
			assertionHasExplicitLiteralType(ctx, parent, contextual)
		if contextual != nil && !explicitLiteral && assertionMemberWriteMayWiden(ctx, expression, expressionType, contextual, retainedConst) {
			return true
		}
	}
	switch expression.Kind {
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		// Const assertions inherit the surrounding contextual type, while
		// ordinary assertions supply their own context to the operand.
		if ast.IsConstAssertion(expression) {
			return assertionValueMayAffectInference(ctx, expression.Expression(), true, memberWrites)
		}
	case ast.KindSatisfiesExpression:
		// The constraint supplies the operand's context. The value still
		// has the operand's inferred type, including its member freshness.
		return memberWrites && assertionAggregateMayWiden(ctx.TypeChecker, expressionType, true)
	case ast.KindTemplateExpression:
		// Context can produce a template type or regular string literals;
		// type-only widening cannot recover the string otherwise produced.
		// A retained const context fixes that independently.
		return memberWrites && !explicitLiteral && !utils.IsTypeFlagSet(expressionType, checker.TypeFlagsString) &&
			(!retainedConst || !checker.Checker_isConstContext(ctx.TypeChecker, expression))
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		if !memberWrites && checker.Checker_getContextualType(ctx.TypeChecker, expression, checker.ContextFlagsNone) == nil {
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
		objectType := expressionType
		if memberWrites {
			if assertionAggregateMayWiden(ctx.TypeChecker, objectType, false) {
				return true
			}
		}
		for _, property := range expression.Properties() {
			switch property.Kind {
			case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
				var value *ast.Node
				var original *checker.Type
				if ast.IsShorthandPropertyAssignment(property) {
					value = property.Name()
					if memberWrites {
						symbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(property)
						if symbol == nil {
							return true
						}
						original = ctx.TypeChecker.GetTypeOfSymbolAtLocation(symbol, value)
					}
				} else {
					value = property.Initializer()
					if memberWrites {
						original = assertionExpressionType(ctx.TypeChecker, value)
					}
				}
				if memberWrites {
					opaque := !ast.IsObjectLiteralExpression(ast.SkipParentheses(value)) && !ast.IsArrayLiteralExpression(ast.SkipParentheses(value))
					if assertionAggregateMayWiden(ctx.TypeChecker, original, opaque) {
						return true
					}
					symbol := property.Symbol()
					if symbol == nil {
						return true
					}
					written := ctx.TypeChecker.GetPropertyOfType(objectType, symbol.Name)
					if written == nil || assertionMemberWriteMayWiden(ctx, value, original,
						checker.Checker_getTypeOfSymbol(ctx.TypeChecker, written), retainedConst) {
						return true
					}
				}
				if assertionValueMayAffectInference(ctx, value, retainedConst, memberWrites) {
					return true
				}
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				if assertionValueMayAffectInference(ctx, property, retainedConst, memberWrites) {
					return true
				}
			case ast.KindSpreadAssignment:
				if memberWrites {
					value := ast.SkipParentheses(property.Expression())
					opaque := !ast.IsObjectLiteralExpression(value) && !ast.IsArrayLiteralExpression(value)
					if assertionAggregateMayWiden(ctx.TypeChecker, assertionExpressionType(ctx.TypeChecker, value), opaque) {
						return true
					}
				}
				if assertionValueMayAffectInference(ctx, property.Expression(), false, memberWrites) {
					return true
				}
			}
		}
	case ast.KindArrayLiteralExpression:
		arrayType := expressionType
		if arrayType == nil {
			arrayType = assertionExpressionType(ctx.TypeChecker, expression)
		}
		if memberWrites && assertionAggregateMayWiden(ctx.TypeChecker, arrayType, false) {
			return true
		}
		// A mutable tuple inferred for an array literal depends on its
		// contextual type; removing that context can produce an array.
		if checker.IsTupleType(arrayType) && !arrayType.TargetTupleType().IsReadonly() {
			return true
		}
		for index, element := range expression.Elements() {
			spread := element.Kind == ast.KindSpreadElement
			if element.Kind == ast.KindSpreadElement {
				element = element.Expression()
			}
			if element.Kind == ast.KindOmittedExpression {
				continue
			}
			if memberWrites {
				value := ast.SkipParentheses(element)
				opaque := !ast.IsObjectLiteralExpression(value) && !ast.IsArrayLiteralExpression(value)
				if assertionAggregateMayWiden(ctx.TypeChecker, assertionExpressionType(ctx.TypeChecker, value), opaque) {
					return true
				}
			}
			if memberWrites && !spread && element.Kind != ast.KindOmittedExpression {
				written := checker.Checker_getIndexTypeOfType(ctx.TypeChecker, arrayType, checker.Checker_numberType(ctx.TypeChecker))
				if checker.IsTupleType(arrayType) {
					if property := ctx.TypeChecker.GetPropertyOfType(arrayType, strconv.Itoa(index)); property != nil {
						written = checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property)
					}
				}
				if assertionMemberWriteMayWiden(ctx, element,
					assertionExpressionType(ctx.TypeChecker, element), written, retainedConst) {
					return true
				}
			}
			if assertionValueMayAffectInference(ctx, element, retainedConst, memberWrites) {
				return true
			}
		}
	case ast.KindConditionalExpression:
		conditional := expression.AsConditionalExpression()
		return assertionValueMayAffectInference(ctx, conditional.WhenTrue, retainedConst, memberWrites) ||
			assertionValueMayAffectInference(ctx, conditional.WhenFalse, retainedConst, memberWrites)
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return assertionValueMayAffectInference(ctx, binary.Left, retainedConst, memberWrites) ||
				assertionValueMayAffectInference(ctx, binary.Right, retainedConst, memberWrites)
		case ast.KindAmpersandAmpersandToken, ast.KindCommaToken:
			return assertionValueMayAffectInference(ctx, binary.Right, retainedConst, memberWrites)
		}
	case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return assertionFunctionMayAffectInference(ctx, expression, true)
	case ast.KindAwaitExpression, ast.KindNonNullExpression:
		return assertionValueMayAffectInference(ctx, expression.Expression(), retainedConst, memberWrites)
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

func assertionHasFixedCallContext(ctx rule.RuleContext, node, call *ast.Node, castIsAny bool) bool {
	if genericCall := assertionGenericCall(ctx, node); genericCall != nil && genericCall != call {
		return false
	}
	calleeType := ctx.TypeChecker.GetNonNullableType(ctx.TypeChecker.GetTypeAtLocation(call.Expression()))
	if utils.IsTypeAnyType(calleeType) {
		return true
	}
	kind := checker.SignatureKindCall
	if ast.IsNewExpression(call) {
		kind = checker.SignatureKindConstruct
	}
	signatures := checker.Checker_getSignaturesOfType(ctx.TypeChecker, calleeType, kind)
	if len(signatures) == 0 {
		return false
	}
	for _, signature := range signatures {
		// Predicates and this parameters can propagate an argument's inferred
		// type even when the call's result is discarded.
		if signature.ThisParameter() != nil || ctx.TypeChecker.GetTypePredicateOfSignature(signature) != nil {
			return false
		}
		if len(signature.TypeParameters()) == 0 || call.TypeArguments() != nil {
			continue
		}
		argumentCall, _, independent := assertionCallArgument(node)
		original := assertionOriginalExpression(node)
		if castIsAny || len(signatures) != 1 || argumentCall != call || !independent ||
			!ast.IsExpressionStatement(assertionWalkUpParentheses(call).Parent) ||
			utils.IsTypeFlagSetWithUnion(ctx.TypeChecker.GetTypeAtLocation(original), checker.TypeFlagsEnumLike) ||
			(!assertionIsConceptuallyLiteral(original) && !ast.IsPropertyAssignment(assertionWalkUpParentheses(node).Parent)) {
			return false
		}
	}
	if len(signatures) > 1 {
		argumentCall, _, _ := assertionCallArgument(node)
		return argumentCall == call && !assertionIsArgumentToOverloadedFunction(ctx, node)
	}
	return true
}

func assertionHasFixedDeclarationContext(ctx rule.RuleContext, node, value, declaration *ast.Node) bool {
	if declaration.Initializer() != value || declaration.Type() == nil {
		return false
	}
	declaredType := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, declaration.Type())
	if !declaredType.IsUnion() {
		return true
	}
	property := assertionWalkUpParentheses(node).Parent
	if !ast.IsPropertyAssignment(property) || property.Parent != value || !ast.IsObjectLiteralExpression(value) {
		return false
	}
	// An assertion is not a discriminant value in the compiler. A single
	// constituent here was therefore selected by other, unchanged members.
	selected := checker.Checker_getApparentTypeOfContextualType(ctx.TypeChecker, value, checker.ContextFlagsNone)
	return selected != nil && !selected.IsUnion() && slices.Contains(declaredType.Types(), selected)
}

// Contextual assignability is only a removal proof at a type boundary. In
// particular, a contextual callback return type can still participate in the
// inference of its caller, including through a block-bodied return statement.
func assertionHasStableContext(ctx rule.RuleContext, node *ast.Node, castIsAny bool) bool {
	original := assertionOriginalExpression(node)
	if ast.IsTemplateExpression(original) ||
		(!castIsAny && ast.IsArrayLiteralExpression(original)) ||
		assertionIsInDestructuringDeclaration(node) || assertionIsPropertyInProblematicContext(ctx, node) {
		return false
	}
	for current := assertionWalkUpParentheses(node); current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindPropertyAssignment:
			if parent.Initializer() != current || ast.IsComputedPropertyName(parent.Name()) {
				return false
			}
		case ast.KindObjectLiteralExpression:
			contextual := checker.Checker_getContextualType(ctx.TypeChecker, parent, checker.ContextFlagsNone)
			if contextual != nil && contextual.IsUnion() {
				selected := checker.Checker_getApparentTypeOfContextualType(ctx.TypeChecker, parent, checker.ContextFlagsNone)
				if selected == nil || selected.IsUnion() {
					return false
				}
			}
		case ast.KindArrayLiteralExpression:
		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().Condition == current {
				return false
			}
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == current {
				// Assignment flow only preserves a non-union declared type. The
				// assignment's own value must not feed another inference.
				if !ast.IsExpressionStatement(assertionWalkUpParentheses(parent).Parent) {
					return false
				}
				symbol := ctx.TypeChecker.GetSymbolAtLocation(binary.Left)
				if symbol == nil {
					return false
				}
				target := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, symbol)
				return !target.IsUnion() && !utils.IsTypeAnyType(target)
			}
			if castIsAny || !ast.IsLogicalOrCoalescingBinaryExpression(parent) {
				return false
			}
			// A left operand also controls flow in the right operand. A literal
			// has no references whose narrowed types could change.
			if binary.Left == current && !ast.IsLiteralExpression(binary.Right) &&
				binary.Right.Kind != ast.KindNullKeyword && binary.Right.Kind != ast.KindTrueKeyword && binary.Right.Kind != ast.KindFalseKeyword {
				return false
			}
		case ast.KindVariableDeclaration, ast.KindParameter, ast.KindPropertyDeclaration:
			return assertionHasFixedDeclarationContext(ctx, node, current, parent)
		case ast.KindArrowFunction:
			if parent.Body() != current {
				return false
			}
			if parent.Type() != nil {
				return true
			}
			if len(parent.Parameters()) != 0 {
				return false
			}
		case ast.KindReturnStatement:
			function := ast.GetContainingFunction(parent)
			return function != nil && function.Type() != nil
		case ast.KindCallExpression, ast.KindNewExpression:
			for _, argument := range parent.Arguments() {
				if argument == current {
					return assertionHasFixedCallContext(ctx, node, parent, castIsAny)
				}
			}
			return false
		case ast.KindJsxExpression:
			attribute := parent.Parent
			if !ast.IsJsxAttribute(attribute) || attribute.Initializer() != parent {
				return false
			}
			opening := attribute.Parent.Parent
			var tag *ast.Node
			var attributes *ast.Node
			if ast.IsJsxOpeningElement(opening) {
				tag = opening.AsJsxOpeningElement().TagName
				attributes = opening.AsJsxOpeningElement().Attributes
			} else if ast.IsJsxSelfClosingElement(opening) {
				tag = opening.AsJsxSelfClosingElement().TagName
				attributes = opening.AsJsxSelfClosingElement().Attributes
			}
			if !ast.IsIdentifier(tag) || !scanner.IsIntrinsicJsxName(tag.Text()) {
				return false
			}
			props := checker.Checker_getContextualType(ctx.TypeChecker, attributes, checker.ContextFlagsNone)
			return props != nil && !props.IsUnion()
		default:
			return false
		}
		current = assertionWalkUpParentheses(parent)
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
		needsParenthesesAfterRemoval := func(assertion, operand *ast.Node) bool {
			first := getTokenRange(operand.Pos())
			switch sourceText[first.Pos():first.End()] {
			case "{":
				return utils.IsStartOfExpressionStatement(ctx.SourceFile, assertion) ||
					assertionIsStartOfArrowFunctionBody(ctx.SourceFile, assertion)
			case "function", "class":
				return utils.IsStartOfExpressionStatement(ctx.SourceFile, assertion)
			}
			return false
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
			return assertionExpressionType(ctx.TypeChecker, expression)
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
			canRemove := func(originalExpression *ast.Node, originalType *checker.Type, memberWrites bool) bool {
				return !assertionValueMayAffectInference(ctx, originalExpression, false, memberWrites) &&
					(assertionPreservesWidening(ctx.TypeChecker, originalType, castType) ||
						checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone) != nil)
			}
			reportAssertion := func(msg rule.RuleMessage) {
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
					if needsParenthesesAfterRemoval(node, expression) {
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
				if !ast.IsAssertionExpression(innerExpression) || ast.IsConstAssertion(innerExpression) {
					return
				}
				intermediateType := ctx.TypeChecker.GetTypeAtLocation(innerExpression)
				if !utils.IsTypeFlagSet(intermediateType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) || utils.IsIntrinsicErrorType(intermediateType) {
					return
				}
				// Only consume this bridge. Earlier assertions, especially as const,
				// must remain in the expression used by both the proof and the fix.
				retainedExpression := innerExpression.Expression()
				retainedType := assertionExpressionType(ctx.TypeChecker, retainedExpression)
				msg := rule.RuleMessage{}
				if isTypeUnchanged(node, retainedExpression, retainedType, castType) && !utils.IsTypeAnyType(castType) &&
					canRemove(retainedExpression, retainedType, true) {
					msg = buildUnnecessaryAssertionMessage()
				} else if contextualType != nil && !assertionTypeContainsAny(ctx.TypeChecker, retainedType) &&
					!assertionTypeContainsAny(ctx.TypeChecker, contextualType) {
					if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, retainedType, contextualType) &&
						canRemove(retainedExpression, retainedType, false) {
						msg = buildContextuallyUnnecessaryMessage()
					}
				}
				if msg.Id == "" {
					return
				}
				ctx.ReportRangeWithDeferredFixes(getTokenRange(node.Pos()).WithEnd(node.End()), msg, func() []rule.RuleFix {
					retainedRange := utils.TrimNodeTextRange(ctx.SourceFile, retainedExpression)
					replacement := sourceText[retainedRange.Pos():retainedRange.End()]
					if needsParenthesesAfterRemoval(node, retainedExpression) {
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
				if canRemove(expression, uncastType, true) {
					reportAssertion(buildUnnecessaryAssertionMessage())
					return
				}
			}

			semanticNode := assertionWalkUpParentheses(node)
			parent := semanticNode.Parent
			castIsAny := utils.IsTypeFlagSet(castType, checker.TypeFlagsAny)
			var contextualType *checker.Type
			if assertionHasStableContext(ctx, node, castIsAny) {
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
					!isNullishLiteralToUnion && canRemove(expression, uncastType, false) {
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
