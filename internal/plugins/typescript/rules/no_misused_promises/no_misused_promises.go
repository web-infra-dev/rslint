package no_misused_promises

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_misused_promises.schema.json
var schemaJSON []byte

// Unknown expressions, spreads and assertions keep the callee-first path.
// These literal forms usually make checking the arguments cheaper instead.
func preferArgumentTypeCheck(arguments []*ast.Node) bool {
	for _, argument := range arguments {
		argument = ast.SkipParentheses(argument)
		switch argument.Kind {
		case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
			ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
			ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression,
			ast.KindObjectLiteralExpression, ast.KindVoidExpression,
			ast.KindTypeOfExpression, ast.KindDeleteExpression:
			continue
		default:
			return false
		}
	}
	return true
}

// isContextIndependentExpression reports whether node's type does not depend on
// its own contextual type and can be queried cheaply before computing that
// contextual type. Context-sensitive or compound expressions (functions, calls,
// object/array literals, template expressions, etc.) return false so callers
// check the contextual type first.
func isContextIndependentExpression(node *ast.Node) bool {
	for node != nil {
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
			ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword,
			ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindRegularExpressionLiteral,
			ast.KindIdentifier:
			return true
		case ast.KindVoidExpression:
			// tsgo's checkVoidExpression defers checking its operand via
			// checkNodeDeferred, so void does not need to recurse; typeof and delete
			// eagerly check their operands and recurse below.
			return true
		case ast.KindNonNullExpression, ast.KindPropertyAccessExpression,
			ast.KindTypeOfExpression, ast.KindDeleteExpression:
			node = node.Expression()
		case ast.KindPrefixUnaryExpression:
			node = node.AsPrefixUnaryExpression().Operand
		case ast.KindPostfixUnaryExpression:
			node = node.AsPostfixUnaryExpression().Operand
		case ast.KindElementAccessExpression:
			expr := node.AsElementAccessExpression()
			return isContextIndependentExpression(expr.Expression) &&
				isContextIndependentExpression(expr.ArgumentExpression)
		case ast.KindBinaryExpression:
			expr := node.AsBinaryExpression()
			return isContextIndependentExpression(expr.Left) &&
				isContextIndependentExpression(expr.Right)
		case ast.KindConditionalExpression:
			expr := node.AsConditionalExpression()
			return isContextIndependentExpression(expr.Condition) &&
				isContextIndependentExpression(expr.WhenTrue) &&
				isContextIndependentExpression(expr.WhenFalse)
		default:
			return false
		}
	}
	return false
}

// isGuaranteedLiteralLike reports whether node is a primitive literal, regex,
// or void expression whose type lookup avoids member/function signature
// resolution and is cheaper than resolving a typed object property's contextual
// type.
func isGuaranteedLiteralLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword,
		ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindRegularExpressionLiteral,
		ast.KindVoidExpression:
		return true
	default:
		return false
	}
}

const primitiveLiteralSlotCount = 7

// primitiveLiteralSlot maps bare (non-parenthesized) atomic literal and void
// expressions to a per-file cache slot tracking whether their apparent global
// type has call signatures. Parenthesized expressions (which may carry JSDoc
// @type casts), JSDoc nodes, and nodes inside `with` statements return -1 so
// they are always queried directly.
func primitiveLiteralSlot(node *ast.Node) int {
	if node == nil || node.Flags&(ast.NodeFlagsInWithStatement|ast.NodeFlagsJSDoc) != 0 {
		return -1
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return 0
	case ast.KindNumericLiteral:
		return 1
	case ast.KindBigIntLiteral:
		return 2
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return 3
	case ast.KindNullKeyword:
		return 4
	case ast.KindRegularExpressionLiteral:
		return 5
	case ast.KindVoidExpression:
		return 6
	default:
		return -1
	}
}

func unionTypeParts(t *checker.Type, buf *[1]*checker.Type) []*checker.Type {
	if t == nil {
		return nil
	}
	if utils.IsUnionType(t) {
		return t.Types()
	}
	buf[0] = t
	return buf[:]
}

func buildConditionalMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "conditional",
		Description: "Expected non-Promise value in a boolean conditional.",
	}
}
func buildPredicateMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "predicate",
		Description: "Expected a non-Promise value to be returned.",
	}
}
func buildSpreadMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "spread",
		Description: "Expected a non-Promise value to be spread in an object.",
	}
}
func buildVoidReturnArgumentMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnArgument",
		Description: "Promise returned in function argument where a void return was expected.",
	}
}
func buildVoidReturnAttributeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnAttribute",
		Description: "Promise-returning function provided to attribute where a void return was expected.",
	}
}
func buildVoidReturnInheritedMethodMessage(heritageTypeName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnInheritedMethod",
		Description: fmt.Sprintf("Promise-returning method provided where a void return was expected by extended/implemented type '%v'.", heritageTypeName),
	}
}
func buildVoidReturnPropertyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnProperty",
		Description: "Promise-returning function provided to property where a void return was expected.",
	}
}
func buildVoidReturnReturnValueMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnReturnValue",
		Description: "Promise-returning function provided to return value where a void return was expected.",
	}
}
func buildVoidReturnVariableMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidReturnVariable",
		Description: "Promise-returning function provided to variable where a void return was expected.",
	}
}

type NoMisusedPromisesChecksVoidReturnOptions struct {
	Arguments        bool
	Attributes       bool
	InheritedMethods bool
	Properties       bool
	Returns          bool
	Variables        bool
}

type NoMisusedPromisesOptions struct {
	ChecksConditionals   bool
	ChecksSpreads        bool
	ChecksVoidReturn     bool
	ChecksVoidReturnOpts NoMisusedPromisesChecksVoidReturnOptions
}

func parseOptions(options []any) NoMisusedPromisesOptions {
	opts := NoMisusedPromisesOptions{
		ChecksConditionals: true,
		ChecksSpreads:      true,
		ChecksVoidReturn:   true,
		ChecksVoidReturnOpts: NoMisusedPromisesChecksVoidReturnOptions{
			Arguments:        true,
			Attributes:       true,
			InheritedMethods: true,
			Properties:       true,
			Returns:          true,
			Variables:        true,
		},
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if value, ok := optsMap["checksConditionals"].(bool); ok {
		opts.ChecksConditionals = value
	}
	if value, ok := optsMap["checksSpreads"].(bool); ok {
		opts.ChecksSpreads = value
	}
	// `checksVoidReturn` is either a boolean enabling or disabling every
	// sub-check, or an object overriding them individually.
	switch value := optsMap["checksVoidReturn"].(type) {
	case bool:
		opts.ChecksVoidReturn = value
	case map[string]any:
		if v, ok := value["arguments"].(bool); ok {
			opts.ChecksVoidReturnOpts.Arguments = v
		}
		if v, ok := value["attributes"].(bool); ok {
			opts.ChecksVoidReturnOpts.Attributes = v
		}
		if v, ok := value["inheritedMethods"].(bool); ok {
			opts.ChecksVoidReturnOpts.InheritedMethods = v
		}
		if v, ok := value["properties"].(bool); ok {
			opts.ChecksVoidReturnOpts.Properties = v
		}
		if v, ok := value["returns"].(bool); ok {
			opts.ChecksVoidReturnOpts.Returns = v
		}
		if v, ok := value["variables"].(bool); ok {
			opts.ChecksVoidReturnOpts.Variables = v
		}
	}
	return opts
}

var NoMisusedPromisesRule = rule.CreateRule(rule.Rule{
	Name:             "no-misused-promises",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		// Intrinsic flags whose reduced apparent type never carries call signatures
		// (even under strictNullChecks: false, getApparentType(unknown) returns
		// emptyObjectType, which has no call signatures).
		const nonCallableTypeFlags = checker.TypeFlagsVoid | checker.TypeFlagsUndefined | checker.TypeFlagsNull | checker.TypeFlagsNever | checker.TypeFlagsAny | checker.TypeFlagsUnknown

		anySignatureIsThenableType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			if t.Flags()&nonCallableTypeFlags != 0 {
				return false
			}
			return utils.Some(utils.GetCallSignatures(ctx.TypeChecker, t), func(sig *checker.Signature) bool {
				return utils.IsThenableType(ctx.TypeChecker, node, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig))
			})
		}

		// Per-file cache for bare primitive/void literals whose apparent global
		// type has zero call signatures.
		var primitiveLiteralNonCallable [primitiveLiteralSlotCount]bool

		returnsThenable := func(node *ast.Node) bool {
			slot := primitiveLiteralSlot(node)
			if slot >= 0 && primitiveLiteralNonCallable[slot] {
				return false
			}
			t := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(node))
			var singlePart [1]*checker.Type
			hasCallable := false
			for _, part := range unionTypeParts(t, &singlePart) {
				if part.Flags()&nonCallableTypeFlags != 0 {
					continue
				}
				sigs := utils.GetCallSignatures(ctx.TypeChecker, part)
				if len(sigs) == 0 {
					continue
				}
				hasCallable = true
				for _, sig := range sigs {
					if utils.IsThenableType(ctx.TypeChecker, node, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig)) {
						return true
					}
				}
			}
			if slot >= 0 && !hasCallable {
				primitiveLiteralNonCallable[slot] = true
			}
			return false
		}

		checkArrayPredicates := func(node *ast.Node) {
			parent := node.Parent

			if !ast.IsCallExpression(parent) {
				return
			}

			expr := parent.AsCallExpression()
			// Original selector matches `CallExpression > MemberExpression.callee`,
			// so only fire when the access expression IS the callee. Without this
			// guard, the listener also fires for member-access arguments and would
			// double-report on the callback.
			if expr.Expression != node {
				return
			}

			arguments := expr.Arguments.Nodes
			if len(arguments) == 0 {
				return
			}

			callback := arguments[0]

			if utils.IsArrayMethodCallWithPredicate(ctx.TypeChecker, expr) && returnsThenable(callback) {
				ctx.ReportNode(callback, buildPredicateMessage())
			}
		}

		isFunctionParam := func(
			param *ast.Symbol,
			node *ast.Node,
		) bool {
			t := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeOfSymbolAtLocation(param, node))
			if t == nil {
				return false
			}
			return utils.Some(utils.UnionTypeParts(t), func(t *checker.Type) bool {
				return len(utils.GetCallSignatures(ctx.TypeChecker, t)) != 0
			})
		}

		// Variation on the thenable check which requires all forms of the type (read:
		// alternates in a union) to be thenable. Otherwise, you might be trying to
		// check if something is defined or undefined and get caught because one of the
		// branches is thenable.
		isAlwaysThenable := func(node *ast.Node) bool {
			t := ctx.TypeChecker.GetTypeAtLocation(node)

			for _, subType := range utils.UnionTypeParts(checker.Checker_getApparentType(ctx.TypeChecker, t)) {
				thenProp := checker.Checker_getPropertyOfType(ctx.TypeChecker, subType, "then")

				// If one of the alternates has no then property, it is not thenable in all
				// cases.
				if thenProp == nil {
					return false
				}

				// We walk through each variation of the then property. Since we know it
				// exists at this point, we just need at least one of the alternates to
				// be of the right form to consider it thenable.
				thenType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(thenProp, node)
				hasThenableSignature := false
				for _, subType := range utils.UnionTypeParts(thenType) {
					for _, signature := range utils.GetCallSignatures(ctx.TypeChecker, subType) {
						params := checker.Signature_parameters(signature)
						if len(params) != 0 && isFunctionParam(params[0], node) {
							hasThenableSignature = true
						}
					}

					// We only need to find one variant of the then property that has a
					// function signature for it to be thenable.
					if hasThenableSignature {
						break
					}
				}

				// If no flavors of the then property are thenable, we don't consider the
				// overall type to be thenable
				if !hasThenableSignature {
					return false
				}
			}

			// If all variants are considered thenable (i.e. haven't returned false), we
			// consider the overall type thenable
			return true
		}

		checkedNodes := map[*ast.Node](struct{}){}

		var checkConditional func(
			node *ast.Expression,
			isTestExpr bool,
		)
		checkConditional = func(
			node *ast.Expression,
			isTestExpr bool,
		) {
			if node == nil || ast.IsAssignmentExpression(node, false) {
				return
			}
			// prevent checking the same node multiple times
			if _, ok := checkedNodes[node]; ok {
				return
			}
			checkedNodes[node] = struct{}{}

			node = ast.SkipParentheses(node)

			if ast.IsBinaryExpression(node) && ast.IsLogicalExpression(node) {
				expr := node.AsBinaryExpression()
				// ignore the left operand for nullish coalescing expressions not in a context of a test expression
				if expr.OperatorToken.Kind != ast.KindQuestionQuestionToken || isTestExpr {
					checkConditional(expr.Left, isTestExpr)
				}
				// we ignore the right operand when not in a context of a test expression
				if isTestExpr {
					checkConditional(expr.Right, isTestExpr)
				}
				return
			}

			if isAlwaysThenable(node) {
				ctx.ReportNode(node, buildConditionalMessage())
			}
		}

		getMemberIfExists := func(
			t *checker.Type,
			memberName string,
		) *ast.Symbol {
			// tsgo's SymbolTable uses raw member names (internal symbols are
			// distinguished by an invalid-UTF8 prefix), so unlike upstream
			// typescript-eslint we don't need an escapeLeadingUnderscores step.
			symbol := checker.Type_symbol(t)
			if symbol != nil {
				if member, ok := symbol.Members[memberName]; ok {
					return member
				}
			}

			return checker.Checker_getPropertyOfType(ctx.TypeChecker, t, memberName)
		}

		isVoidReturningFunctionType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			if t == nil || t.Flags()&nonCallableTypeFlags != 0 {
				return false
			}
			hadVoidReturn := false
			var singlePart [1]*checker.Type
			for _, part := range unionTypeParts(t, &singlePart) {
				if part.Flags()&nonCallableTypeFlags != 0 {
					continue
				}
				for _, sig := range utils.GetCallSignatures(ctx.TypeChecker, part) {
					returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig)
					// If a certain positional argument accepts both thenable and void returns,
					// a promise-returning function is valid
					if utils.IsThenableType(ctx.TypeChecker, node, returnType) {
						return false
					}

					hadVoidReturn = hadVoidReturn || utils.IsTypeFlagSet(returnType, checker.TypeFlagsVoid)
				}
			}
			return hadVoidReturn
		}

		/**
		 * Checks `heritageType` for a member named `memberName` that returns void; reports the
		 * 'voidReturnInheritedMethod' message if found.
		 * @param nodeMember Node member that returns a Promise
		 * @param heritageType Heritage type to check against
		 * @param memberName Name of the member to check for
		 */
		checkHeritageTypeForMemberReturningVoid := func(
			nodeMember *ast.Node,
			heritageType *checker.Type,
			memberName string,
		) {
			heritageMember := getMemberIfExists(heritageType, memberName)
			if heritageMember == nil {
				return
			}
			memberType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(
				heritageMember,
				nodeMember,
			)
			if !isVoidReturningFunctionType(nodeMember, memberType) {
				return
			}
			ctx.ReportNode(nodeMember, buildVoidReturnInheritedMethodMessage(ctx.TypeChecker.TypeToString(heritageType)))
		}

		// isVoidContextualThenable checks whether valueNode returns a thenable in a
		// position whose contextual type expects a void-returning function. For
		// context-independent expressions, checking returnsThenable first avoids
		// contextual-type resolution while keeping TypeChecker authority (including
		// global primitive prototype augmentation).
		isVoidContextualThenable := func(
			contextNode *ast.Node,
			valueNode *ast.Node,
		) bool {
			if isContextIndependentExpression(valueNode) {
				if !returnsThenable(valueNode) {
					return false
				}
				contextualType := checker.Checker_getContextualType(ctx.TypeChecker, contextNode, checker.ContextFlagsNone)
				return contextualType != nil && isVoidReturningFunctionType(contextNode, contextualType)
			}
			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, contextNode, checker.ContextFlagsNone)
			return contextualType != nil && isVoidReturningFunctionType(contextNode, contextualType) && returnsThenable(valueNode)
		}

		checkJSXAttribute := func(node *ast.JsxAttribute) {
			if node.Initializer == nil || node.Initializer.Kind != ast.KindJsxExpression {
				return
			}
			expression := node.Initializer.AsJsxExpression().Expression
			if expression == nil {
				return
			}
			if isVoidContextualThenable(node.Initializer, expression) {
				ctx.ReportNode(node.Initializer, buildVoidReturnAttributeMessage())
			}
		}

		checkSpread := func(node *ast.Node) {
			if utils.IsThenableType(ctx.TypeChecker, node.Expression(), nil) {
				ctx.ReportNode(node.Expression(), buildSpreadMessage())
			}
		}

		isThenableReturningFunctionType := func(
			node *ast.Node,
			t *checker.Type,
		) bool {
			var singlePart [1]*checker.Type
			for _, part := range unionTypeParts(t, &singlePart) {
				if anySignatureIsThenableType(node, part) {
					return true
				}
			}
			return false
		}

		// Get the positions of arguments which are void functions (and not also
		// thenable functions). These are the candidates for the void-return check at
		// the current call site.
		// If the function parameters end with a 'rest' parameter, then we consider
		// the array type parameter (e.g. '...args:Array<SomeType>') when determining
		// if trailing arguments are candidates.
		// Precondition: len(args) > 0 (callers filter out zero-argument calls and
		// argument-less `new` expressions beforehand).
		voidFunctionArguments := func(
			node *ast.Expression,
			args []*ast.Node,
		) []int {
			argCount := len(args)

			const (
				argVisited uint8 = 1 << iota
				argVoid
				argThenable
			)
			var (
				argStatesBuf [8]uint8
				lastTypesBuf [8]*checker.Type
				argStates    []uint8
				lastTypes    []*checker.Type
			)
			if argCount <= len(argStatesBuf) {
				argStates = argStatesBuf[:argCount]
				lastTypes = lastTypesBuf[:argCount]
			} else {
				argStates = make([]uint8, argCount)
				lastTypes = make([]*checker.Type, argCount)
			}
			callee := node.Expression()

			checkArgType := func(t *checker.Type, index int) {
				if argStates[index]&argThenable != 0 {
					return
				}
				argStates[index] |= argVisited
				if t == nil || t == lastTypes[index] {
					return
				}
				lastTypes[index] = t
				if isThenableReturningFunctionType(callee, t) {
					argStates[index] |= argThenable
				} else if isVoidReturningFunctionType(callee, t) {
					argStates[index] |= argVoid
				}
			}

			t := ctx.TypeChecker.GetTypeAtLocation(callee)

			// We can't use checker.getResolvedSignature because it prefers an early '() => void' over a later '() => Promise<void>'
			// See https://github.com/microsoft/TypeScript/issues/48077

			isCall := ast.IsCallExpression(node)
			for _, subType := range utils.UnionTypeParts(t) {
				// Standard function calls and `new` have two different types of signatures
				var signatures []*checker.Signature
				if isCall {
					signatures = utils.GetCallSignatures(ctx.TypeChecker, subType)
				} else {
					signatures = utils.GetConstructSignatures(ctx.TypeChecker, subType)
				}
				for _, signature := range signatures {
					for index, parameter := range checker.Signature_parameters(signature) {
						if index >= argCount {
							break
						}
						decl := parameter.ValueDeclaration
						// If this is a array 'rest' parameter, check all of the argument indices
						// from the current argument to the end.
						if decl != nil && utils.IsRestParameterDeclaration(decl) {
							hasUnresolved := false
							for i := index; i < argCount; i++ {
								if argStates[i]&argThenable == 0 {
									hasUnresolved = true
									break
								}
							}
							if !hasUnresolved {
								break
							}
							paramType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(parameter, callee)
							if checker.Checker_isArrayType(ctx.TypeChecker, paramType) {
								// Unwrap 'Array<MaybeVoidFunction>' to 'MaybeVoidFunction',
								// so that we'll handle it in the same way as a non-rest
								// 'param: MaybeVoidFunction'
								typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, paramType)
								if len(typeArgs) == 0 {
									break
								}
								elemType := typeArgs[0]
								elemFlag := argVisited
								if elemType != nil {
									if isThenableReturningFunctionType(callee, elemType) {
										elemFlag |= argThenable
									} else if isVoidReturningFunctionType(callee, elemType) {
										elemFlag |= argVoid
									}
								}
								for i := index; i < argCount; i++ {
									if argStates[i]&argThenable != 0 {
										continue
									}
									argStates[i] |= elemFlag
									lastTypes[i] = elemType
								}
							} else if checker.IsTupleType(paramType) {
								// Check each type in the tuple - for example, [boolean, () => void] would
								// add the index of the second tuple parameter to 'voidReturnIndices'
								typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, paramType)
								for i := index; i < argCount && i-index < len(typeArgs); i++ {
									checkArgType(typeArgs[i-index], i)
								}
							}
							break
						}
						if argStates[index]&argThenable != 0 {
							continue
						}
						paramType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(parameter, callee)
						checkArgType(paramType, index)
					}
				}
			}

			// Query each visited, non-thenable argument's contextual type at most
			// once (e.g. to pick up generic instantiations) instead of per overload.
			var voidReturnIndices []int
			for index := range argCount {
				if argStates[index]&argVisited == 0 || argStates[index]&argThenable != 0 {
					continue
				}
				contextualType := checker.Checker_getContextualTypeForArgumentAtIndex(ctx.TypeChecker, node, index)
				checkArgType(contextualType, index)
				if argStates[index]&argVoid != 0 && argStates[index]&argThenable == 0 {
					voidReturnIndices = append(voidReturnIndices, index)
				}
			}

			return voidReturnIndices
		}

		// isPromiseFinallyMethod mirrors typescript-eslint's `parseFinallyCall`
		// gate: skip `<Promise>.finally(callback)` call sites because returning
		// a Promise from a `finally` handler is a legitimate pattern.
		isPromiseFinallyMethod := func(node *ast.Node) bool {
			if !ast.IsCallExpression(node) {
				return false
			}
			callee := node.AsCallExpression().Expression
			if !ast.IsAccessExpression(callee) {
				return false
			}
			methodName, ok := checker.Checker_getAccessedPropertyName(ctx.TypeChecker, callee)
			if !ok || methodName != "finally" {
				return false
			}
			objectType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())
			return utils.IsPromiseLike(ctx.Program(), ctx.TypeChecker, objectType)
		}

		checkArguments := func(
			node *ast.Expression,
		) {
			args := node.Arguments()
			if len(args) == 0 {
				return
			}
			if preferArgumentTypeCheck(args) {
				// Literal forms are usually non-callable, but global primitive
				// augmentation can add signatures. The checker remains authoritative.
				possible := false
				for _, argument := range args {
					if returnsThenable(argument) {
						possible = true
						break
					}
				}
				if !possible {
					return
				}
			}
			if isPromiseFinallyMethod(node) {
				return
			}

			voidArgs := voidFunctionArguments(node, args)
			for _, index := range voidArgs {
				argument := args[index]
				if returnsThenable(argument) {
					ctx.ReportNode(argument, buildVoidReturnArgumentMessage())
				}
			}
		}

		checkClassLikeOrInterfaceNode := func(
			node *ast.Node,
		) {
			heritageClauses := utils.GetHeritageClauses(node)
			if heritageClauses == nil || len(heritageClauses.Nodes) == 0 {
				return
			}

			heritageTypes := utils.Flatten(utils.Map(heritageClauses.Nodes, func(h *ast.Node) []*checker.Type {
				return utils.Map(h.AsHeritageClause().Types.Nodes, func(n *ast.Node) *checker.Type {
					return ctx.TypeChecker.GetTypeAtLocation(n)
				})
			}))

			for _, nodeMember := range node.Members() {
				if nodeMember.Name() == nil {
					// Call/construct/index signatures don't have names. TS allows call signatures to mismatch,
					// and construct signatures can't be async.
					// TODO - Once we're able to use `checker.isTypeAssignableTo` (v8), we can check an index
					// signature here against its compatible index signatures in `heritageTypes`
					continue
				}
				if !ast.IsIdentifier(nodeMember.Name()) && !ast.IsPrivateIdentifier(nodeMember.Name()) && !ast.IsStringLiteral(nodeMember.Name()) && !ast.IsNumericLiteral(nodeMember.Name()) && !ast.IsBigIntLiteral(nodeMember.Name()) {
					continue
				}
				memberName := nodeMember.Name().Text()
				if ast.IsStatic(nodeMember) {
					continue
				}
				if !returnsThenable(nodeMember) {
					continue
				}
				for _, heritageType := range heritageTypes {
					checkHeritageTypeForMemberReturningVoid(
						nodeMember,
						heritageType,
						memberName,
					)
				}
			}
		}

		// Active enclosing ArrayLiteralExpressions and multi-property
		// ObjectLiteralExpressions whose contextual-type presence has already
		// been determined, maintained in strict O(1) amortized LIFO order by
		// lazily expiring exited ancestors (`End() <= curPos`) without
		// registering per-array or per-object exit listeners.
		type contextualCacheEntry struct {
			node  *ast.Node
			isNil bool
		}
		expireContextStack := func(stack *[]contextualCacheEntry, curPos int) {
			s := *stack
			for len(s) > 0 && s[len(s)-1].node.End() <= curPos {
				s[len(s)-1] = contextualCacheEntry{}
				s = s[:len(s)-1]
			}
			*stack = s
		}

		var arrayContextStack []contextualCacheEntry
		lookupArrayContext := func(arr *ast.Node, curPos int) (isNil bool, known bool) {
			if arr == nil {
				return false, false
			}
			expireContextStack(&arrayContextStack, curPos)
			if n := len(arrayContextStack); n > 0 && arrayContextStack[n-1].node == arr {
				return arrayContextStack[n-1].isNil, true
			}
			return false, false
		}
		markArrayContext := func(arr *ast.Node, isNil bool, curPos int) {
			if arr == nil || len(arr.Elements()) <= 1 {
				return
			}
			expireContextStack(&arrayContextStack, curPos)
			if n := len(arrayContextStack); n > 0 && arrayContextStack[n-1].node == arr {
				return
			}
			arrayContextStack = append(arrayContextStack, contextualCacheEntry{node: arr, isNil: isNil})
		}

		var objectContextStack []contextualCacheEntry
		lookupObjectContext := func(obj *ast.Node, curPos int) (isNil bool, known bool) {
			if obj == nil {
				return false, false
			}
			expireContextStack(&objectContextStack, curPos)
			if n := len(objectContextStack); n > 0 && objectContextStack[n-1].node == obj {
				return objectContextStack[n-1].isNil, true
			}
			return false, false
		}
		markObjectContext := func(obj *ast.Node, isNil bool, curPos int) {
			if obj == nil || len(obj.Properties()) <= 1 {
				return
			}
			expireContextStack(&objectContextStack, curPos)
			if n := len(objectContextStack); n > 0 && objectContextStack[n-1].node == obj {
				return
			}
			objectContextStack = append(objectContextStack, contextualCacheEntry{node: obj, isNil: isNil})
		}

		// enclosingContextParent walks past transparent wrappers (parentheses
		// and `as const` / `<const>` assertions) that forward their parent's
		// contextual type unchanged. In JS files, tsgo already reparses
		// `/** @type {T} */ (...)` and `/** @satisfies {T} */ (...)` into
		// synthetic AsExpression / SatisfiesExpression nodes inside the parens,
		// so explicit JSDoc casts stop naturally before reaching the parens.
		enclosingContextParent := func(node *ast.Node) *ast.Node {
			parent := node.Parent
			for parent != nil {
				if parent.Kind == ast.KindParenthesizedExpression {
					parent = parent.Parent
					continue
				}
				if (parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindTypeAssertionExpression) && ast.IsConstAssertion(parent) {
					parent = parent.Parent
					continue
				}
				break
			}
			return parent
		}

		// hasLikelyTypedContext is a purely syntactic hint used only to pick
		// query order (property-first vs object-first) so typed objects and
		// arrays avoid an extra Checker_getContextualType(obj) call while
		// untyped multi-property objects avoid querying twice on a nil result.
		hasLikelyTypedContext := func(obj *ast.Node, curPos int) bool {
			node := obj
			for depth := 0; depth < 4 && node != nil; depth++ {
				parent := enclosingContextParent(node)
				if parent == nil {
					return false
				}
				if parent.Kind == ast.KindArrayLiteralExpression {
					if isNil, known := lookupArrayContext(parent, curPos); known {
						return !isNil
					}
					parent = enclosingContextParent(parent)
					if parent == nil {
						return false
					}
				}
				switch parent.Kind {
				case ast.KindVariableDeclaration, ast.KindParameter, ast.KindPropertyDeclaration:
					return parent.Type() != nil
				case ast.KindTypeAssertionExpression, ast.KindAsExpression:
					return !ast.IsConstAssertion(parent)
				case ast.KindSatisfiesExpression, ast.KindCallExpression, ast.KindNewExpression, ast.KindJsxExpression:
					return true
				case ast.KindPropertyAssignment:
					if parent.Type() != nil {
						return true
					}
					if isNil, known := lookupObjectContext(parent.Parent, curPos); known {
						return !isNil
					}
					return false
				case ast.KindArrowFunction:
					if parent.Type() != nil {
						return true
					}
					node = parent
					continue
				case ast.KindReturnStatement:
					fn := parent.Parent
					for fn != nil && !ast.IsFunctionLike(fn) {
						if fn.Kind == ast.KindClassStaticBlockDeclaration {
							return false
						}
						fn = fn.Parent
					}
					if fn == nil {
						return false
					}
					if fn.Type() != nil {
						return true
					}
					if fn.Kind == ast.KindArrowFunction || fn.Kind == ast.KindFunctionExpression {
						node = fn
						continue
					}
					return false
				}
				return false
			}
			return false
		}

		var lastObjectLiteral *ast.Node
		var lastObjectLiteralContext *checker.Type
		getObjectLiteralContext := func(obj *ast.Node, curPos int) *checker.Type {
			if obj == lastObjectLiteral {
				return lastObjectLiteralContext
			}
			parent := enclosingContextParent(obj)
			if parent != nil {
				switch parent.Kind {
				case ast.KindArrayLiteralExpression:
					if isNil, known := lookupArrayContext(parent, curPos); known && isNil {
						lastObjectLiteral = obj
						lastObjectLiteralContext = nil
						return nil
					}
				case ast.KindPropertyAssignment:
					if parent.Type() == nil {
						if isNil, known := lookupObjectContext(parent.Parent, curPos); known && isNil {
							lastObjectLiteral = obj
							lastObjectLiteralContext = nil
							return nil
						}
					}
				}
			}
			t := checker.Checker_getContextualType(ctx.TypeChecker, obj, checker.ContextFlagsNone)
			lastObjectLiteral = obj
			lastObjectLiteralContext = t
			if parent != nil && parent.Kind == ast.KindArrayLiteralExpression && len(parent.Elements()) > 1 {
				// Only cache the array as untyped when the array itself has no
				// contextual type (e.g. not a tuple or numeric-keyed object whose
				// element at this index has no contextual type).
				if _, known := lookupArrayContext(parent, curPos); !known {
					if t == nil {
						arrCtx := checker.Checker_getContextualType(ctx.TypeChecker, parent, checker.ContextFlagsNone)
						markArrayContext(parent, arrCtx == nil, curPos)
					} else {
						markArrayContext(parent, false, curPos)
					}
				}
			}
			return t
		}

		// Mirrors tsgo's getContextualTypeForObjectLiteralElement: after a
		// property without its own JSDoc @type (node.Type() == nil) observes a
		// nil contextual type, inspect its enclosing multi-property object or
		// multi-element array so subsequent siblings/elements skip redundant
		// queries.
		maybeMarkNilContext := func(node *ast.Node, checkRemainingSiblings bool, curPos int) {
			if node.Type() != nil || node.Parent == nil || node.Parent.Kind != ast.KindObjectLiteralExpression {
				return
			}
			obj := node.Parent
			props := obj.Properties()
			_, objKnown := lookupObjectContext(obj, curPos)
			hasMultiProps := !objKnown && len(props) > 1 && (!checkRemainingSiblings || props[len(props)-1] != node)
			parent := enclosingContextParent(obj)
			hasMultiArray := parent != nil && parent.Kind == ast.KindArrayLiteralExpression && len(parent.Elements()) > 1
			if hasMultiArray {
				if _, arrKnown := lookupArrayContext(parent, curPos); arrKnown {
					hasMultiArray = false
				}
			}
			if !hasMultiProps && !hasMultiArray {
				return
			}
			objCtx := getObjectLiteralContext(obj, curPos)
			if hasMultiProps {
				markObjectContext(obj, objCtx == nil, curPos)
			}
		}

		checkProperty := func(node *ast.Node) {
			curPos := node.Pos()
			if ast.IsPropertyAssignment(node) {
				property := node.AsPropertyAssignment()
				if isGuaranteedLiteralLike(property.Initializer) {
					if node.Type() == nil && node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
						if isNil, known := lookupObjectContext(node.Parent, curPos); known && isNil {
							return
						}
					}
					if slot := primitiveLiteralSlot(property.Initializer); slot < 0 || !primitiveLiteralNonCallable[slot] {
						// On a cold primitive literal inside an unannotated property,
						// check the enclosing object's contextual type first so
						// untyped literal tables avoid initializing global primitive
						// interface call signatures on a fresh checker.
						if node.Type() == nil && node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
							objCtx := getObjectLiteralContext(node.Parent, curPos)
							markObjectContext(node.Parent, objCtx == nil, curPos)
							if objCtx == nil {
								return
							}
						}
					}
					if !returnsThenable(property.Initializer) {
						return
					}
					contextualType := checker.Checker_getContextualType(ctx.TypeChecker, property.Initializer, checker.ContextFlagsNone)
					if contextualType == nil {
						return
					}
					if isVoidReturningFunctionType(property.Initializer, contextualType) {
						ctx.ReportNode(property.Initializer, buildVoidReturnPropertyMessage())
					}
					return
				}
				if node.Type() == nil && node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
					obj := node.Parent
					if isNil, known := lookupObjectContext(obj, curPos); known {
						if isNil {
							return
						}
					} else {
						if parent := enclosingContextParent(obj); parent != nil && parent.Kind == ast.KindArrayLiteralExpression {
							if isNil, arrKnown := lookupArrayContext(parent, curPos); arrKnown && isNil {
								markObjectContext(obj, true, curPos)
								return
							}
						}
						props := obj.Properties()
						if len(props) > 1 && props[len(props)-1] != node && !hasLikelyTypedContext(obj, curPos) {
							objCtx := getObjectLiteralContext(obj, curPos)
							markObjectContext(obj, objCtx == nil, curPos)
							if objCtx == nil {
								return
							}
						}
					}
				}
				contextualType := checker.Checker_getContextualType(ctx.TypeChecker, property.Initializer, checker.ContextFlagsNone)
				if contextualType == nil {
					maybeMarkNilContext(node, true, curPos)
					return
				}
				if node.Type() == nil && node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
					markObjectContext(node.Parent, false, curPos)
				}
				if isVoidReturningFunctionType(property.Initializer, contextualType) &&
					returnsThenable(property.Initializer) {
					if ast.IsFunctionLike(property.Initializer) {
						returnType := property.Initializer.Type()
						if returnType != nil {
							ctx.ReportNode(returnType, buildVoidReturnPropertyMessage())
						} else {
							ctx.ReportRange(
								utils.GetFunctionHeadLoc(ctx.SourceFile, property.Initializer),
								buildVoidReturnPropertyMessage(),
							)
						}
					} else {
						ctx.ReportNode(property.Initializer, buildVoidReturnPropertyMessage())
					}
				}
			} else if ast.IsShorthandPropertyAssignment(node) {
				if node.Type() == nil && node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
					obj := node.Parent
					if isNil, known := lookupObjectContext(obj, curPos); known {
						if isNil {
							return
						}
					} else {
						if parent := enclosingContextParent(obj); parent != nil && parent.Kind == ast.KindArrayLiteralExpression {
							if isNil, arrKnown := lookupArrayContext(parent, curPos); arrKnown && isNil {
								markObjectContext(obj, true, curPos)
								return
							}
						}
						props := obj.Properties()
						if len(props) > 1 && props[len(props)-1] != node && !hasLikelyTypedContext(obj, curPos) {
							objCtx := getObjectLiteralContext(obj, curPos)
							markObjectContext(obj, objCtx == nil, curPos)
							if objCtx == nil {
								return
							}
						}
					}
				}
				// Keep contextual-type-first here: GetTypeAtLocation(node.Name()) enters
				// checkShorthandPropertyAssignment -> checkExpressionForMutableLocation,
				// which always queries getContextualType(node.Name()) internally.
				contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node.Name(), checker.ContextFlagsNone)
				if contextualType == nil {
					maybeMarkNilContext(node, true, curPos)
					return
				}
				if node.Type() == nil && node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
					markObjectContext(node.Parent, false, curPos)
				}
				if isVoidReturningFunctionType(node.Name(), contextualType) &&
					returnsThenable(node.Name()) {
					ctx.ReportNode(node.Name(), buildVoidReturnPropertyMessage())
				}
			} else if ast.IsMethodDeclaration(node) {
				if ast.IsComputedPropertyName(node.Name()) {
					return
				}
				obj := node.Parent

				// Below condition isn't satisfied unless something goes wrong,
				// but is needed for type checking.
				// 'node' does not include class method declaration so 'obj' is
				// always an object literal expression, but after converting 'node'
				// to TypeScript AST, its type includes MethodDeclaration which
				// does include the case of class method declaration.
				if !ast.IsObjectLiteralExpression(obj) {
					return
				}

				if isNil, known := lookupObjectContext(obj, curPos); known && isNil {
					return
				}
				if !returnsThenable(node) {
					return
				}
				objType := getObjectLiteralContext(obj, curPos)
				markObjectContext(obj, objType == nil, curPos)
				if objType == nil {
					return
				}
				// Mirror upstream's `unionConstituents(objType).map(...).find(p => p)`:
				// for object literals assigned to `T | undefined` (or any union),
				// the member must be looked up on each constituent individually
				// because TS's getPropertyOfType on the union returns nothing when
				// the property is missing from any branch.
				memberName := node.Name().Text()
				var propertySymbol *ast.Symbol
				var singlePart [1]*checker.Type
				for _, part := range unionTypeParts(objType, &singlePart) {
					if sym := checker.Checker_getPropertyOfType(ctx.TypeChecker, part, memberName); sym != nil {
						propertySymbol = sym
						break
					}
				}
				if propertySymbol == nil {
					return
				}

				contextualType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(
					propertySymbol,
					node.Name(),
				)

				if isVoidReturningFunctionType(node.Name(), contextualType) {
					if node.Type() != nil {
						ctx.ReportNode(node.Type(), buildVoidReturnPropertyMessage())
					} else {
						ctx.ReportRange(
							utils.GetFunctionHeadLoc(ctx.SourceFile, node),
							buildVoidReturnPropertyMessage(),
						)
					}
				}
			}
		}

		/**
		 * A syntactic check to see if an annotated type is maybe a function type.
		 * This is a perf optimization to help avoid requesting types where possible
		 */
		isPossiblyFunctionType := func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindConditionalType,
				ast.KindConstructorType,
				ast.KindFunctionType,
				ast.KindImportType,
				ast.KindIndexedAccessType,
				ast.KindInferType,
				ast.KindIntersectionType,
				ast.KindQualifiedName,
				ast.KindThisType,
				ast.KindTypeOperator,
				ast.KindTypeQuery,
				ast.KindTypeReference,
				ast.KindUnionType:
				return true

			case ast.KindTypeLiteral:
				return utils.Some(node.AsTypeLiteralNode().Members.Nodes, func(member *ast.Node) bool {
					return member.Kind == ast.KindCallSignature || member.Kind == ast.KindConstructSignature
				})

			case ast.KindAbstractKeyword,
				ast.KindAnyKeyword,
				ast.KindArrayType,
				ast.KindAsyncKeyword,
				ast.KindBigIntKeyword,
				ast.KindBooleanKeyword,
				ast.KindDeclareKeyword,
				ast.KindExportKeyword,
				ast.KindIntrinsicKeyword,
				ast.KindLiteralType,
				ast.KindMappedType,
				ast.KindNamedTupleMember,
				ast.KindNeverKeyword,
				ast.KindNullKeyword,
				ast.KindNumberKeyword,
				ast.KindObjectKeyword,
				ast.KindOptionalType,
				ast.KindPrivateKeyword,
				ast.KindProtectedKeyword,
				ast.KindPublicKeyword,
				ast.KindReadonlyKeyword,
				ast.KindRestType,
				ast.KindStaticKeyword,
				ast.KindStringKeyword,
				ast.KindSymbolKeyword,
				ast.KindTemplateLiteralType,
				ast.KindTupleType,
				ast.KindTypePredicate,
				ast.KindUndefinedKeyword,
				ast.KindUnknownKeyword,
				ast.KindVoidKeyword:
				return false
			}
			return false
		}

		checkReturnStatement := func(node *ast.ReturnStatement) {
			if node.Expression == nil {
				return
			}

			// syntactically ignore some known-good cases to avoid touching type info
			crossedLiteralOrStaticBlock := false
			current := node.Parent
			for current != nil && !ast.IsFunctionLike(current) {
				switch current.Kind {
				case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindClassStaticBlockDeclaration:
					crossedLiteralOrStaticBlock = true
				}
				current = current.Parent
			}
			functionNode := current

			if functionNode == nil {
				return
			}

			if functionNode.Type() != nil && !isPossiblyFunctionType(functionNode.Type()) {
				return
			}

			isUnannotatedObjectMethod := !crossedLiteralOrStaticBlock &&
				ast.IsObjectLiteralMethod(functionNode) &&
				functionNode.Type() == nil &&
				functionNode.FunctionLikeData().FullSignature == nil
			if isUnannotatedObjectMethod {
				if isNil, known := lookupObjectContext(functionNode.Parent, node.Pos()); known && isNil {
					return
				}
			}

			if functionNode.Type() != nil || isGuaranteedLiteralLike(node.Expression) {
				if isVoidContextualThenable(node.Expression, node.Expression) {
					ctx.ReportNode(node.Expression, buildVoidReturnReturnValueMessage())
				}
				return
			}

			// Keep unannotated functions contextual-type-first for non-literal
			// returns: getContextualReturnType returns nil immediately when no
			// contextual type applies, avoiding value type lookup on non-literals.
			contextualType := checker.Checker_getContextualType(ctx.TypeChecker, node.Expression, checker.ContextFlagsNone)
			if contextualType == nil {
				if isUnannotatedObjectMethod {
					maybeMarkNilContext(functionNode, false, node.Pos())
				}
				return
			}
			if isVoidReturningFunctionType(
				node.Expression,
				contextualType,
			) && returnsThenable(node.Expression) {
				ctx.ReportNode(node.Expression, buildVoidReturnReturnValueMessage())
			}
		}

		checkAssignment := func(node *ast.BinaryExpression) {
			varType := ctx.TypeChecker.GetTypeAtLocation(node.Left)
			if !isVoidReturningFunctionType(node.Left, varType) {
				return
			}

			if returnsThenable(node.Right) {
				ctx.ReportNode(node.Right, buildVoidReturnVariableMessage())
			}
		}

		checkVariableDeclaration := func(node *ast.VariableDeclaration) {
			if node.Initializer == nil ||
				node.Type == nil {
				return
			}

			// syntactically ignore some known-good cases to avoid touching type info
			if !isPossiblyFunctionType(node.Type) {
				return
			}

			varType := ctx.TypeChecker.GetTypeAtLocation(node.Name())
			if !isVoidReturningFunctionType(node.Initializer, varType) {
				return
			}

			if returnsThenable(node.Initializer) {
				ctx.ReportNode(node.Initializer, buildVoidReturnVariableMessage())
			}
		}

		listeners := rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if opts.ChecksConditionals {
					checkConditional(node, false)
				}

				if opts.ChecksVoidReturn && opts.ChecksVoidReturnOpts.Variables && ast.IsAssignmentExpression(node, false) {
					checkAssignment(node.AsBinaryExpression())
				}
			},
		}
		if opts.ChecksConditionals {
			listeners[ast.KindPropertyAccessExpression] = checkArrayPredicates
			listeners[ast.KindElementAccessExpression] = checkArrayPredicates

			listeners[ast.KindPrefixUnaryExpression] = func(node *ast.Node) {
				expr := node.AsPrefixUnaryExpression()
				if expr.Operator == ast.KindExclamationToken {
					checkConditional(expr.Operand, true)
				}
			}

			listeners[ast.KindConditionalExpression] = func(node *ast.Node) { checkConditional(node.AsConditionalExpression().Condition, true) }
			listeners[ast.KindForStatement] = func(node *ast.Node) { checkConditional(node.AsForStatement().Condition, true) }
			listeners[ast.KindDoStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
			listeners[ast.KindWhileStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
			listeners[ast.KindIfStatement] = func(node *ast.Node) { checkConditional(node.Expression(), true) }
		}

		if opts.ChecksVoidReturn {
			if opts.ChecksVoidReturnOpts.Arguments {
				listeners[ast.KindCallExpression] = checkArguments
				listeners[ast.KindNewExpression] = checkArguments
			}
			if opts.ChecksVoidReturnOpts.Attributes {
				listeners[ast.KindJsxAttribute] = func(node *ast.Node) { checkJSXAttribute(node.AsJsxAttribute()) }
			}
			if opts.ChecksVoidReturnOpts.InheritedMethods {
				listeners[ast.KindClassDeclaration] = checkClassLikeOrInterfaceNode
				listeners[ast.KindClassExpression] = checkClassLikeOrInterfaceNode
				listeners[ast.KindInterfaceDeclaration] = checkClassLikeOrInterfaceNode
			}
			if opts.ChecksVoidReturnOpts.Properties {
				listeners[ast.KindPropertyAssignment] = checkProperty
				listeners[ast.KindMethodDeclaration] = checkProperty
				listeners[ast.KindShorthandPropertyAssignment] = checkProperty
			}
			if opts.ChecksVoidReturnOpts.Returns {
				listeners[ast.KindReturnStatement] = func(node *ast.Node) { checkReturnStatement(node.AsReturnStatement()) }
			}
			if opts.ChecksVoidReturnOpts.Variables {
				listeners[ast.KindVariableDeclaration] = func(node *ast.Node) { checkVariableDeclaration(node.AsVariableDeclaration()) }
			}

		}
		if opts.ChecksSpreads {
			listeners[ast.KindSpreadElement] = checkSpread
			listeners[ast.KindSpreadAssignment] = checkSpread
		}

		return listeners

	},
})
