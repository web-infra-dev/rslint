package no_unsafe_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func formatSenderType(senderType *checker.Type) string {
	if utils.IsIntrinsicErrorType(senderType) {
		return "error typed"
	}
	return "`any`"
}

func buildAnyAssignmentMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "anyAssignment",
		Description: "Unsafe assignment of an " + formatSenderType(sender) + " value.",
	}
}
func buildAnyAssignmentThisMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "anyAssignmentThis",
		Description: "Unsafe assignment of an " + formatSenderType(sender) + " value. `this` is typed as `any`.\n" +
			"You can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}
func buildUnsafeArrayPatternMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArrayPattern",
		Description: "Unsafe array destructuring of an " + formatSenderType(sender) + " array value.",
	}
}
func buildUnsafeArrayPatternFromTupleMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArrayPatternFromTuple",
		Description: "Unsafe array destructuring of a tuple element with an " + formatSenderType(sender) + " value.",
	}
}
func buildUnsafeObjectPatternMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeObjectPattern",
		Description: "Unsafe object destructuring of a property with an " + formatSenderType(sender) + " value.",
	}
}
func buildUnsafeArraySpreadMessage(sender *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArraySpread",
		Description: "Unsafe spread of an " + formatSenderType(sender) + " value in an array.",
	}
}
func buildUnsafeAssignmentMessage(typeChecker *checker.Checker, sender, receiver *checker.Type) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeAssignment",
		Description: "Unsafe assignment of type `" + typeChecker.TypeToString(sender) + "` to a variable of type `" + typeChecker.TypeToString(receiver) + "`.",
	}
}

type comparisonType uint8

const (
	/** Do no assignment comparison */
	comparisonTypeNone comparisonType = iota
	/** Use the receiver's type for comparison */
	comparisonTypeBasic
	/** Use the sender's contextual type for comparison */
	comparisonTypeContextual
)

// assignmentTypeResolver compensates for a typescript-go inference difference
// that is observable only with strictNullChecks disabled. TypeScript derives a
// non-any result from the right-hand side of `any && value`, whereas
// typescript-go currently propagates `any` through the whole expression. Keep
// the fallback deliberately syntax-bound so annotated `any` values still
// produce the same diagnostics as the upstream rule.
type assignmentTypeResolver struct {
	typeChecker *checker.Checker
	enabled     bool
	state       *assignmentTypeResolverState
}

type assignmentTypeResolverState struct {
	recoveredTypes      map[*ast.Node]*checker.Type
	recovering          map[*ast.Node]bool
	functionReturnTypes map[*ast.Node]*checker.Type
	recoveringFunctions map[*ast.Node]bool
}

func newAssignmentTypeResolver(typeChecker *checker.Checker, enabled bool) assignmentTypeResolver {
	resolver := assignmentTypeResolver{
		typeChecker: typeChecker,
		enabled:     enabled,
	}
	if enabled {
		resolver.state = &assignmentTypeResolverState{}
	}
	return resolver
}

func (r assignmentTypeResolver) typeAtLocation(node *ast.Node) *checker.Type {
	rawType := r.typeChecker.GetTypeAtLocation(node)
	if !r.enabled || !utils.IsTypeAnyType(rawType) {
		return rawType
	}
	return r.recoverType(node, rawType, nil)
}

func (r assignmentTypeResolver) recoverTypeAtLocation(node *ast.Node, substitutions map[*ast.Node]*checker.Type) *checker.Type {
	return r.recoverType(node, r.typeChecker.GetTypeAtLocation(node), substitutions)
}

func (r assignmentTypeResolver) recoverType(node *ast.Node, rawType *checker.Type, substitutions map[*ast.Node]*checker.Type) (result *checker.Type) {
	node = ast.SkipParentheses(node)
	if substitutedType, ok := r.substitutedIdentifierType(node, substitutions); ok {
		return substitutedType
	}
	if !utils.IsTypeAnyType(rawType) {
		return rawType
	}

	var identifierInitializer *ast.Node
	switch {
	case ast.IsBinaryExpression(node):
		if node.AsBinaryExpression().OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
			return rawType
		}
	case ast.IsIdentifier(node):
		symbol := r.typeChecker.GetSymbolAtLocation(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return rawType
		}
		declaration := symbol.Declarations[0]
		if !ast.IsVariableDeclaration(declaration) || declaration.Type() != nil {
			return rawType
		}
		identifierInitializer = declaration.Initializer()
		if identifierInitializer == nil {
			return rawType
		}
	case ast.IsPropertyAccessExpression(node), ast.IsConditionalExpression(node), ast.IsCallExpression(node):
		// These forms can expose a non-any type through their children.
	default:
		return rawType
	}

	cacheable := substitutions == nil
	if cacheable && r.state.recoveredTypes != nil {
		if recoveredType, ok := r.state.recoveredTypes[node]; ok {
			return recoveredType
		}
	}
	if r.state.recovering != nil && r.state.recovering[node] {
		return rawType
	}
	if r.state.recovering == nil {
		r.state.recovering = make(map[*ast.Node]bool)
	}
	r.state.recovering[node] = true
	defer delete(r.state.recovering, node)

	result = rawType
	if cacheable {
		if r.state.recoveredTypes == nil {
			r.state.recoveredTypes = make(map[*ast.Node]*checker.Type)
		}
		defer func() {
			r.state.recoveredTypes[node] = result
		}()
	}

	switch {
	case ast.IsBinaryExpression(node):
		binary := node.AsBinaryExpression()
		rightType := r.recoverTypeAtLocation(binary.Right, substitutions)
		if !utils.IsTypeAnyType(rightType) {
			return rightType
		}
	case ast.IsIdentifier(node):
		initializerType := r.recoverTypeAtLocation(identifierInitializer, substitutions)
		if !utils.IsTypeAnyType(initializerType) {
			return initializerType
		}
	case ast.IsPropertyAccessExpression(node):
		objectType := r.recoverTypeAtLocation(node.Expression(), substitutions)
		if utils.IsTypeAnyType(objectType) {
			break
		}
		propertyName, ok := checker.Checker_getAccessedPropertyName(r.typeChecker, node)
		if !ok {
			break
		}
		property := checker.Checker_getPropertyOfType(r.typeChecker, objectType, propertyName)
		if property == nil {
			break
		}
		propertyType := r.typeChecker.GetTypeOfSymbolAtLocation(property, node)
		if !utils.IsTypeAnyType(propertyType) {
			return propertyType
		}
	case ast.IsConditionalExpression(node):
		conditional := node.AsConditionalExpression()
		return r.mergeRecoveredTypes(rawType, []*checker.Type{
			r.recoverTypeAtLocation(conditional.WhenTrue, substitutions),
			r.recoverTypeAtLocation(conditional.WhenFalse, substitutions),
		})
	case ast.IsCallExpression(node):
		if returnType := r.recoverFunctionReturnType(node, rawType, substitutions); !utils.IsTypeAnyType(returnType) {
			return returnType
		}
	}

	return rawType
}

func (r assignmentTypeResolver) substitutedIdentifierType(node *ast.Node, substitutions map[*ast.Node]*checker.Type) (*checker.Type, bool) {
	if substitutions == nil || !ast.IsIdentifier(node) {
		return nil, false
	}
	symbol := r.typeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return nil, false
	}
	for _, declaration := range symbol.Declarations {
		if substitutedType, ok := substitutions[declaration]; ok {
			return substitutedType, true
		}
	}
	return nil, false
}

func (r assignmentTypeResolver) recoverFunctionReturnType(callNode *ast.Node, rawType *checker.Type, substitutions map[*ast.Node]*checker.Type) (result *checker.Type) {
	callee := callNode.Expression()
	callee = ast.SkipParentheses(callee)
	var functionNode *ast.Node
	if ast.IsArrowFunction(callee) || ast.IsFunctionExpression(callee) {
		functionNode = callee
	} else if ast.IsIdentifier(callee) {
		symbol := r.typeChecker.GetSymbolAtLocation(callee)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return rawType
		}
		declaration := symbol.Declarations[0]
		if ast.IsFunctionDeclaration(declaration) {
			functionNode = declaration
		} else if ast.IsVariableDeclaration(declaration) {
			initializer := declaration.Initializer()
			if initializer != nil {
				initializer = ast.SkipParentheses(initializer)
				if ast.IsArrowFunction(initializer) || ast.IsFunctionExpression(initializer) {
					functionNode = initializer
				}
			}
		}
	}
	if functionNode == nil || functionNode.Type() != nil {
		return rawType
	}
	if r.state.recoveringFunctions != nil && r.state.recoveringFunctions[functionNode] {
		return rawType
	}

	callSubstitutions := substitutions
	if functionNode.TypeParameters() != nil {
		callSubstitutions = make(map[*ast.Node]*checker.Type, len(substitutions)+len(functionNode.Parameters()))
		for declaration, substitutedType := range substitutions {
			callSubstitutions[declaration] = substitutedType
		}
		resolvedSignature := checker.Checker_getResolvedSignature(r.typeChecker, callNode, nil, checker.CheckModeNormal)
		if resolvedSignature != nil {
			instantiatedParameters := checker.Signature_parameters(resolvedSignature)
			parameterIndex := 0
			for _, parameter := range functionNode.Parameters() {
				parameterName := parameter.Name()
				if ast.IsIdentifier(parameterName) && parameterName.Text() == "this" {
					continue
				}
				if parameterIndex >= len(instantiatedParameters) {
					break
				}
				parameterType := r.typeChecker.GetTypeAtLocation(parameterName)
				if r.typeContainsTypeParameter(parameterType, 0) {
					callSubstitutions[parameter] = r.typeChecker.GetTypeOfSymbolAtLocation(instantiatedParameters[parameterIndex], callNode)
				}
				parameterIndex++
			}
		}
	}

	cacheable := callSubstitutions == nil
	if cacheable && r.state.functionReturnTypes != nil {
		if returnType, ok := r.state.functionReturnTypes[functionNode]; ok {
			return returnType
		}
	}
	if r.state.recoveringFunctions == nil {
		r.state.recoveringFunctions = make(map[*ast.Node]bool)
	}
	r.state.recoveringFunctions[functionNode] = true
	defer delete(r.state.recoveringFunctions, functionNode)
	result = rawType
	if cacheable {
		if r.state.functionReturnTypes == nil {
			r.state.functionReturnTypes = make(map[*ast.Node]*checker.Type)
		}
		defer func() {
			r.state.functionReturnTypes[functionNode] = result
		}()
	}

	body := functionNode.Body()
	if body == nil {
		return rawType
	}
	if !ast.IsBlock(body) {
		returnType := r.recoverTypeAtLocation(body, callSubstitutions)
		if r.typeContainsTypeParameter(returnType, 0) {
			return rawType
		}
		return returnType
	}

	var returnTypes []*checker.Type
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node != body && ast.IsFunctionLike(node) {
			return false
		}
		if ast.IsReturnStatement(node) {
			expression := node.AsReturnStatement().Expression
			if expression != nil {
				returnTypes = append(returnTypes, r.recoverTypeAtLocation(expression, callSubstitutions))
			}
			return false
		}
		return node.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	if len(returnTypes) == 0 {
		return rawType
	}
	return r.mergeRecoveredTypes(rawType, returnTypes)
}

func (r assignmentTypeResolver) mergeRecoveredTypes(rawType *checker.Type, types []*checker.Type) *checker.Type {
	var representative *checker.Type
	for _, currentType := range types {
		if utils.IsTypeAnyType(currentType) || r.typeContainsTypeParameter(currentType, 0) {
			return rawType
		}
		if utils.IsTypeFlagSet(currentType, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}
		if representative == nil {
			representative = currentType
			continue
		}
		if _, _, unsafe := utils.IsUnsafeAssignment(currentType, representative, r.typeChecker, nil); unsafe {
			representative = currentType
		}
	}
	if representative == nil {
		return rawType
	}
	return representative
}

func (r assignmentTypeResolver) typeContainsTypeParameter(t *checker.Type, depth int) bool {
	if t == nil {
		return false
	}
	// Type arguments and union/intersection constituents are normally shallow.
	// Treat a pathological recursive type as unresolved instead of risking an
	// unsafe recovery or spending unbounded time walking it.
	if depth >= 32 {
		return true
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsTypeParameter) {
		return true
	}
	for _, part := range utils.UnionTypeParts(t) {
		if part != t && r.typeContainsTypeParameter(part, depth+1) {
			return true
		}
	}
	for _, part := range utils.IntersectionTypeParts(t) {
		if part != t && r.typeContainsTypeParameter(part, depth+1) {
			return true
		}
	}
	if checker.IsNonDeferredTypeReference(t) {
		for _, typeArgument := range checker.Checker_getTypeArguments(r.typeChecker, t) {
			if r.typeContainsTypeParameter(typeArgument, depth+1) {
				return true
			}
		}
	}
	return false
}

// objectLiteralAnyValuesAreRecoverable identifies object literals whose only
// unsafe property values are direct expressions covered by recoverType. This
// lets the generic-aware assignment comparison avoid reporting the stale any
// embedded in the checker's object type without hiding an explicit any value.
func (r assignmentTypeResolver) objectLiteralAnyValuesAreRecoverable(node *ast.Node) bool {
	if !r.enabled {
		return false
	}
	node = ast.SkipParentheses(node)
	if !ast.IsObjectLiteralExpression(node) {
		return false
	}

	foundAny := false
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		var value *ast.Node
		switch {
		case ast.IsPropertyAssignment(property):
			value = property.Initializer()
		case ast.IsShorthandPropertyAssignment(property):
			value = property.Name()
		case ast.IsSpreadAssignment(property):
			value = property.Expression()
		default:
			continue
		}
		valueType := r.typeChecker.GetTypeAtLocation(value)
		if !utils.IsTypeAnyType(valueType) {
			continue
		}
		foundAny = true
		if utils.IsTypeAnyType(r.typeAtLocation(value)) {
			return false
		}
	}
	return foundAny
}

var NoUnsafeAssignmentRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-assignment",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		compilerOptions := ctx.Program.Options()
		isNoImplicitThis := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.NoImplicitThis,
		)
		typeResolver := newAssignmentTypeResolver(
			ctx.TypeChecker,
			!utils.IsStrictCompilerOptionEnabled(
				compilerOptions,
				compilerOptions.StrictNullChecks,
			),
		)

		var checkArrayDestructure func(
			receiverNode *ast.Node,
			senderType *checker.Type,
			senderNode *ast.Node,
		) bool
		var checkObjectDestructure func(
			receiverNode *ast.Node,
			senderType *checker.Type,
			senderNode *ast.Node,
		) bool

		// returns true if the assignment reported
		checkObjectDestructure = func(
			receiverNode *ast.Node,
			senderType *checker.Type,
			senderNode *ast.Node,
		) bool {
			checkObjectProperty := func(propertyKey *ast.Node, propertyValue *ast.Node) bool {
				var key string
				if !ast.IsComputedPropertyName(propertyKey) {
					key = propertyKey.Text()
				} else if ast.IsLiteralExpression(propertyKey.Expression()) {
					key = propertyKey.Expression().Text()
				} else {
					// can't figure out the name, so skip it
					return false
				}

				property := checker.Checker_getPropertyOfType(ctx.TypeChecker, senderType, key)
				if property == nil {
					return false
				}
				senderType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(property, senderNode)

				// check for the any type first so we can handle {x: {y: z}} = {x: any}
				if utils.IsTypeAnyType(senderType) {
					ctx.ReportNode(propertyValue, buildUnsafeObjectPatternMessage(senderType))
					return true
				} else if ast.IsArrayBindingPattern(propertyValue) || ast.IsArrayLiteralExpression(propertyValue) {
					return checkArrayDestructure(
						propertyValue,
						senderType,
						senderNode,
					)
				} else if ast.IsObjectBindingPattern(propertyValue) || ast.IsObjectLiteralExpression(propertyValue) {
					return checkObjectDestructure(
						propertyValue,
						senderType,
						senderNode,
					)
				}
				return false
			}

			didReport := false
			if ast.IsObjectLiteralExpression(receiverNode) {
				for _, receiverProperty := range receiverNode.AsObjectLiteralExpression().Properties.Nodes {
					if ast.IsSpreadAssignment(receiverProperty) {
						// don't bother checking rest
						continue
					}

					if (ast.IsPropertyAssignment(receiverProperty) && checkObjectProperty(receiverProperty.Name(), receiverProperty.Initializer())) || (ast.IsShorthandPropertyAssignment(receiverProperty) && checkObjectProperty(receiverProperty.Name(), receiverProperty.Name())) {
						didReport = true
					}
				}
			} else if ast.IsObjectBindingPattern(receiverNode) {
				for _, receiverProperty := range receiverNode.AsBindingPattern().Elements.Nodes {
					property := receiverProperty.AsBindingElement()
					if property.DotDotDotToken != nil {
						// don't bother checking rest
						continue
					}

					propertyKey := property.PropertyName
					if propertyKey == nil {
						propertyKey = property.Name()
					}

					if checkObjectProperty(propertyKey, property.Name()) {
						didReport = true
					}
				}
			}

			return didReport
		}

		// returns true if the assignment reported
		checkObjectDestructureHelper := func(
			receiverNode *ast.Node,
			senderNode *ast.Node,
		) bool {
			if !ast.IsObjectBindingPattern(receiverNode) && !ast.IsObjectLiteralExpression(receiverNode) {
				return false
			}

			senderType := typeResolver.typeAtLocation(senderNode)

			return checkObjectDestructure(receiverNode, senderType, senderNode)
		}

		// returns true if the assignment reported
		checkArrayDestructure = func(
			receiverNode *ast.Node,
			senderType *checker.Type,
			senderNode *ast.Node,
		) bool {
			// any array
			// const [x] = ([] as any[]);
			if utils.IsTypeAnyArrayType(senderType, ctx.TypeChecker) {
				ctx.ReportNode(receiverNode, buildUnsafeArrayPatternMessage(senderType))
				return false
			}

			if !checker.IsTupleType(senderType) {
				return true
			}

			tupleElements := checker.Checker_getTypeArguments(ctx.TypeChecker, senderType)

			checkArrayElement := func(receiverElement *ast.Node, receiverIndex int) bool {
				if receiverElement == nil {
					return false
				}
				if receiverIndex >= len(tupleElements) {
					return false
				}
				senderType := tupleElements[receiverIndex]

				// check for the any type first so we can handle [[[x]]] = [any]
				if utils.IsTypeAnyType(senderType) {
					ctx.ReportNode(receiverElement, buildUnsafeArrayPatternFromTupleMessage(senderType))
					return true
				} else if ast.IsArrayBindingPattern(receiverElement) || ast.IsArrayLiteralExpression(receiverElement) {
					return checkArrayDestructure(
						receiverElement,
						senderType,
						senderNode,
					)
				} else if ast.IsObjectBindingPattern(receiverElement) || ast.IsObjectLiteralExpression(receiverElement) {
					return checkObjectDestructure(
						receiverElement,
						senderType,
						senderNode,
					)
				}

				return false
			}

			// tuple with any
			// const [x] = [1 as any];
			didReport := false
			if ast.IsArrayLiteralExpression(receiverNode) {
				for receiverIndex, receiverElement := range receiverNode.AsArrayLiteralExpression().Elements.Nodes {
					if ast.IsSpreadElement(receiverElement) {
						// don't handle rests as they're not a 1:1 assignment
						continue
					}

					if checkArrayElement(receiverElement, receiverIndex) {
						didReport = true
					}
				}
			} else if ast.IsArrayBindingPattern(receiverNode) {
				for receiverIndex, receiverElement := range receiverNode.AsBindingPattern().Elements.Nodes {
					elem := receiverElement.AsBindingElement()
					if elem.DotDotDotToken != nil {
						// don't handle rests as they're not a 1:1 assignment
						continue
					}

					if checkArrayElement(receiverElement.Name(), receiverIndex) {
						// TODO(port): in original rule didReport was reassigned every time. isn't it a bug?
						didReport = true
					}
				}
			}

			return didReport
		}

		// returns true if the assignment reported
		checkArrayDestructureHelper := func(
			receiverNode *ast.Node,
			senderNode *ast.Node,
		) bool {
			if !ast.IsArrayBindingPattern(receiverNode) && !ast.IsArrayLiteralExpression(receiverNode) {
				return false
			}

			senderType := typeResolver.typeAtLocation(senderNode)

			return checkArrayDestructure(receiverNode, senderType, senderNode)
		}

		// returns true if the assignment reported
		checkAssignment := func(
			receiverNode *ast.Node,
			senderNode *ast.Node,
			reportingNode *ast.Node,
			compType comparisonType,
		) bool {
			senderType := typeResolver.typeAtLocation(senderNode)

			if utils.IsTypeAnyType(senderType) {
				// handle cases when we assign any ==> unknown.
				var receiverType *checker.Type
				if compType == comparisonTypeContextual {
					receiverType = utils.GetContextualType(ctx.TypeChecker, receiverNode)
				}
				if receiverType == nil {
					receiverType = ctx.TypeChecker.GetTypeAtLocation(receiverNode)
				}
				if utils.IsTypeUnknownType(receiverType) {
					return false
				}

				if !isNoImplicitThis {
					// `var foo = this`
					thisExpression := utils.GetThisExpression(senderNode)
					if thisExpression != nil && utils.IsTypeAnyType(utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)) {
						ctx.ReportNode(reportingNode, buildAnyAssignmentThisMessage(senderType))
						return true
					}
				}

				ctx.ReportNode(reportingNode, buildAnyAssignmentMessage(senderType))
				return true
			}

			if compType == comparisonTypeNone {
				return false
			}
			var receiverType *checker.Type
			if compType == comparisonTypeContextual {
				receiverType = utils.GetContextualType(ctx.TypeChecker, receiverNode)
			}
			if receiverType == nil {
				receiverType = ctx.TypeChecker.GetTypeAtLocation(receiverNode)
			}

			receiver, sender, unsafe := utils.IsUnsafeAssignment(
				senderType,
				receiverType,
				ctx.TypeChecker,
				senderNode,
			)
			if !unsafe {
				return false
			}
			if typeResolver.objectLiteralAnyValuesAreRecoverable(senderNode) {
				return false
			}

			ctx.ReportNode(reportingNode, buildUnsafeAssignmentMessage(ctx.TypeChecker, sender, receiver))
			return true
		}

		getComparisonType := func(
			nodeWithTypeAnnotation *ast.Node,
		) comparisonType {
			if nodeWithTypeAnnotation.Type() != nil {
				// if there's a type annotation, we can do a comparison
				return comparisonTypeBasic
			}
			// no type annotation means the variable's type will just be inferred, thus equal
			return comparisonTypeNone
		}

		checkAssignmentFull := func(id *ast.Node, init *ast.Node, node *ast.Node) {
			if id == nil || init == nil {
				return
			}
			didReport := checkAssignment(
				id,
				init,
				node,
				// the variable already has some form of a type to compare against
				comparisonTypeBasic,
			)

			if !didReport {
				didReport = checkArrayDestructureHelper(id, init)
			}
			if !didReport {
				checkObjectDestructureHelper(id, init)
			}
		}

		return rule.RuleListeners{
			// ESTree PropertyDefinition, AccessorProperty
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				initializer := node.Initializer()
				if initializer == nil {
					return
				}
				checkAssignment(node.Name(), initializer, node, getComparisonType(node))
			},

			// ESTree AssignmentExpression, AssignmentPattern
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node, true) {
					return
				}

				expr := node.AsBinaryExpression()
				checkAssignmentFull(expr.Left, expr.Right, node)
			},

			// ESTree AssignmentPattern
			ast.KindBindingElement: func(node *ast.Node) {
				checkAssignmentFull(node.Name(), node.Initializer(), node)
			},
			// ESTree AssignmentPattern
			ast.KindParameter: func(node *ast.Node) {
				checkAssignmentFull(node.Name(), node.Initializer(), node)
			},
			// ESTree AssignmentPattern
			ast.KindShorthandPropertyAssignment: func(node *ast.Node) {
				assignment := node.AsShorthandPropertyAssignment()
				checkAssignmentFull(assignment.Name(), assignment.ObjectAssignmentInitializer, node)
			},

			ast.KindVariableDeclaration: func(node *ast.Node) {
				init := node.Initializer()
				if init == nil {
					return
				}

				id := node.Name()
				didReport := checkAssignment(
					id,
					init,
					node,
					getComparisonType(node),
				)

				if !didReport {
					didReport = checkArrayDestructureHelper(id, init)
				}
				if !didReport {
					checkObjectDestructureHelper(id, init)
				}
			},

			// object pattern props are checked via assignments
			rule.ListenerOnNotAllowPattern(ast.KindObjectLiteralExpression): func(node *ast.Node) {
				for _, node := range node.AsObjectLiteralExpression().Properties.Nodes {
					var init *ast.Node
					if ast.IsPropertyAssignment(node) {
						init = node.Initializer()
					} else if ast.IsShorthandPropertyAssignment(node) {
						assignment := node.AsShorthandPropertyAssignment()
						if assignment.ObjectAssignmentInitializer != nil {
							// Assignment patterns are handled by the shorthand listener.
							continue
						}
						init = assignment.Name()
					} else {
						continue
					}

					if init == nil {
						continue
					}
					init = ast.SkipParentheses(init)

					checkAssignment(node.Name(), init, node, comparisonTypeContextual)
				}
			},

			rule.ListenerOnNotAllowPattern(ast.KindArrayLiteralExpression): func(node *ast.Node) {
				for _, node := range node.AsArrayLiteralExpression().Elements.Nodes {
					if !ast.IsSpreadElement(node) {
						continue
					}

					restType := typeResolver.typeAtLocation(node.Expression())
					if utils.IsTypeAnyType(restType) || utils.IsTypeAnyArrayType(restType, ctx.TypeChecker) {
						ctx.ReportNode(node, buildUnsafeArraySpreadMessage(restType))
					}
				}
			},

			ast.KindJsxAttribute: func(node *ast.Node) {
				init := node.Initializer()
				if init == nil || init.Kind != ast.KindJsxExpression {
					return
				}

				expr := init.AsJsxExpression().Expression
				if expr == nil {
					return
				}

				checkAssignment(node.Name(), expr, expr, comparisonTypeContextual)
			},
		}
	},
})
