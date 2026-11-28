package no_unsafe_argument

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnsafeArgumentMessage(sender string, receiver string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArgument",
		Description: "Unsafe argument of type " + sender + " assigned to a parameter of type " + receiver + ".",
	}
}
func buildUnsafeArraySpreadMessage(sender string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeArraySpread",
		Description: "Unsafe spread of an " + sender + " array type.",
	}
}
func buildUnsafeSpreadMessage(sender string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeSpread",
		Description: "Unsafe spread of an " + sender + " type.",
	}
}
func buildUnsafeTupleSpreadMessage(sender string, receiver string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeTupleSpread",
		Description: "Unsafe spread of a tuple type. The argument is " + sender + " and is assigned to a parameter of type " + receiver + ".",
	}
}

type restTypeKind uint32

const (
	restTypeKindArray restTypeKind = iota
	restTypeKindTuple
	restTypeKindOther
)

type restType struct {
	Kind          restTypeKind
	Type          *checker.Type
	TypeArguments []*checker.Type
}

func newFunctionSignature(
	typeChecker *checker.Checker,
	node *ast.Node,
) (functionSignature, bool) {
	signature := checker.Checker_getResolvedSignature(typeChecker, node, nil, checker.CheckModeNormal)
	if signature == nil {
		return functionSignature{}, false
	}

	parameters := checker.Signature_parameters(signature)
	restIndex := -1
	for i, param := range parameters {
		if len(param.Declarations) != 0 && utils.IsRestParameterDeclaration(param.Declarations[0]) {
			restIndex = i
			break
		}
	}

	return functionSignature{
		parameters: parameters,
		restIndex:  restIndex,
	}, true
}

type functionSignature struct {
	parameters   []*ast.Symbol
	restIndex    int
	restType     restType
	restResolved bool
}

func (s *functionSignature) resolveRestType(typeChecker *checker.Checker, node *ast.Node) {
	if s.restResolved {
		return
	}
	s.restResolved = true

	param := s.parameters[s.restIndex]
	t := typeChecker.GetTypeOfSymbolAtLocation(param, node)
	constrainedType := checker.Checker_getBaseConstraintOfType(typeChecker, t)
	if constrainedType == nil {
		constrainedType = t
	}

	if checker.IsTupleType(constrainedType) {
		s.restType = restType{
			Kind:          restTypeKindTuple,
			TypeArguments: checker.Checker_getTypeArguments(typeChecker, constrainedType),
		}
	} else if elementType := utils.GetNumberIndexType(typeChecker, constrainedType); elementType != nil {
		s.restType = restType{
			Type: elementType,
			Kind: restTypeKindArray,
		}
	} else {
		s.restType = restType{
			Type: constrainedType,
			Kind: restTypeKindOther,
		}
	}
}

func (s *functionSignature) getParameterType(
	typeChecker *checker.Checker,
	node *ast.Node,
	index int,
	hasConsumedArguments bool,
) *checker.Type {
	regularCount := len(s.parameters)
	if s.restIndex >= 0 {
		regularCount = s.restIndex
	}

	if index >= regularCount || hasConsumedArguments {
		if s.restIndex < 0 {
			return nil
		}
		s.resolveRestType(typeChecker, node)

		switch s.restType.Kind {
		case restTypeKindTuple:
			typeArguments := s.restType.TypeArguments
			if len(typeArguments) == 0 {
				return nil
			}
			if hasConsumedArguments {
				// all types consumed by a rest - just assume it's the last type
				// there is one edge case where this is wrong, but we ignore it because
				// it's rare and really complicated to handle
				// eg: function foo(...a: [number, ...string[], number])
				return typeArguments[len(typeArguments)-1]
			}

			typeIndex := index - s.restIndex
			if typeIndex >= len(typeArguments) {
				return typeArguments[len(typeArguments)-1]
			}

			return typeArguments[typeIndex]
		case restTypeKindArray, restTypeKindOther:
			return s.restType.Type
		}
		return nil
	}

	return typeChecker.GetTypeOfSymbolAtLocation(s.parameters[index], node)
}

// calleeMayShortCircuitQuickType reports whether TypeChecker.GetTypeAtLocation
// can resolve callee via tsgo's getQuickTypeOfExpression without checking inner
// sub-expressions (assertions, await-wrapped assertions/calls/new, or single
// non-generic calls/new). Resolving the call signature ensures inner callee
// expressions are checked in source order with a non-nil currentNode.
func calleeMayShortCircuitQuickType(callee *ast.Expression) bool {
	expr := ast.SkipParentheses(callee)
	for ast.IsAwaitExpression(expr) {
		expr = ast.SkipParentheses(expr.Expression())
	}
	return ast.IsAssertionExpression(expr) ||
		ast.IsCallExpression(expr) ||
		ast.IsNewExpression(expr)
}

var NoUnsafeArgumentRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-argument",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		describeType := func(t *checker.Type) string {
			if utils.IsIntrinsicErrorType(t) {
				return "error typed"
			}

			return "`" + ctx.TypeChecker.TypeToString(t) + "`"
		}

		describeTypeForSpread := func(t *checker.Type) string {
			if checker.Checker_isArrayType(ctx.TypeChecker, t) && utils.IsIntrinsicErrorType(checker.Checker_getTypeArguments(ctx.TypeChecker, t)[0]) {
				return "error"
			}

			return describeType(t)
		}

		describeTypeForTuple := func(t *checker.Type) string {
			if utils.IsIntrinsicErrorType(t) {
				return "error typed"
			}

			return "of type `" + ctx.TypeChecker.TypeToString(t) + "`"
		}

		checkUnsafeArguments := func(
			args []*ast.Node,
			callee *ast.Expression,
			node *ast.Node,
		) {
			if len(args) == 0 {
				return
			}

			// ignore any-typed calls as these are caught by no-unsafe-call;
			// always resolve the callee before skipping literal arguments so
			// chained generic callees (e.g. `api<T>().on("ready")`) preserve
			// TypeScript's checker evaluation order, and return before
			// resolving call signatures on `any` callees so arguments inside
			// `any` calls are not checked prematurely.
			calleeType := ctx.TypeChecker.GetTypeAtLocation(callee)
			if utils.IsTypeAnyType(calleeType) {
				return
			}

			var signature functionSignature
			signatureLoaded := false

			ensureSignature := func() {
				if !signatureLoaded {
					var ok bool
					signature, ok = newFunctionSignature(ctx.TypeChecker, node)
					if !ok {
						panic("Expected to a signature resolved")
					}
					signatureLoaded = true
				}
			}

			argIndex := 0
			hasConsumedArguments := false
			if ast.IsTaggedTemplateExpression(node) {
				// Consumes the first parameter (TemplateStringsArray) of the function called with TaggedTemplateExpression.
				argIndex = 1
			}

			getParameterType := func(index int, consumed bool) *checker.Type {
				ensureSignature()
				return signature.getParameterType(ctx.TypeChecker, node, index, consumed)
			}

			// Resolve signatures before skipping literal arguments when the call
			// has explicit type arguments or a non-`any` callee that can
			// short-circuit via tsgo's getQuickTypeOfExpression (assertions,
			// calls, new, await) so TypeScript evaluates them in source order.
			if len(node.TypeArguments()) > 0 || calleeMayShortCircuitQuickType(callee) {
				ensureSignature()
			}

			for _, argument := range args {
				switch argument.Kind {
				// spreads consume
				case ast.KindSpreadElement:
					// Resolve the call signature before querying argument types so tsgo
					// checks assertions/calls inside arguments with a non-nil currentNode
					// instead of emitting file-less global diagnostics via getQuickTypeOfExpression.
					ensureSignature()
					spreadArgType := ctx.TypeChecker.GetTypeAtLocation(argument.Expression())

					// Match upstream: check any, any[] and concrete tuples only. Other
					// iterable spreads are ignored without advancing the parameter cursor.
					if utils.IsTypeAnyType(spreadArgType) {
						// foo(...any)
						ctx.ReportNode(argument, buildUnsafeSpreadMessage(describeType(spreadArgType)))
					} else if utils.IsTypeAnyArrayType(spreadArgType, ctx.TypeChecker) {
						// foo(...any[])

						// TODO - we could break down the spread and compare the array type against each argument
						ctx.ReportNode(argument, buildUnsafeArraySpreadMessage(describeTypeForSpread(spreadArgType)))
					} else if checker.IsTupleType(spreadArgType) {
						// foo(...[tuple1, tuple2])
						spreadTypeArguments := checker.Checker_getTypeArguments(ctx.TypeChecker, spreadArgType)
						for _, tupleType := range spreadTypeArguments {
							index := argIndex
							argIndex += 1
							parameterType := getParameterType(index, hasConsumedArguments)
							if parameterType == nil || utils.IsTypeFlagSet(parameterType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
								continue
							}
							if !utils.IsTypeAnyType(tupleType) && (!utils.IsTypeReference(tupleType) || !utils.IsTypeReference(parameterType) || tupleType.Target() != parameterType.Target()) {
								continue
							}
							_, _, unsafe := utils.IsUnsafeAssignment(
								tupleType,
								parameterType,
								ctx.TypeChecker,
								// we can't pass the individual tuple members in here as this will most likely be a spread variable
								// not a spread array
								nil,
							)
							if unsafe {
								ctx.ReportNode(argument, buildUnsafeTupleSpreadMessage(describeTypeForTuple(tupleType), describeType(parameterType)))
							}
						}
						if checker.TupleType_combinedFlags(spreadArgType.Target().AsTupleType())&checker.ElementFlagsVariable != 0 {
							// the last element was a rest - so all remaining defined arguments can be considered "consumed"
							// all remaining arguments should be compared against the rest type (if one exists)
							hasConsumedArguments = true
						}
					}

				default:
					index := argIndex
					argIndex += 1

					reportNode := ast.SkipParentheses(argument)
					if utils.IsGuaranteedSafeLiteral(reportNode) {
						continue
					}
					// Skip excess arguments before querying their type, and query
					// bound argument types before skipping `any`/`unknown`
					// parameters so failed earlier arguments or arity errors do
					// not leave generic payloads unchecked before later calls.
					parameterType := getParameterType(index, hasConsumedArguments)
					if parameterType == nil {
						continue
					}
					argumentType := ctx.TypeChecker.GetTypeAtLocation(reportNode)
					if utils.IsTypeFlagSet(parameterType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
						continue
					}

					if !utils.IsTypeAnyType(argumentType) && (!utils.IsTypeReference(argumentType) || !utils.IsTypeReference(parameterType) || argumentType.Target() != parameterType.Target()) {
						continue
					}

					_, _, unsafe := utils.IsUnsafeAssignment(
						argumentType,
						parameterType,
						ctx.TypeChecker,
						reportNode,
					)
					if unsafe {
						ctx.ReportNode(reportNode, buildUnsafeArgumentMessage(describeType(argumentType), describeType(parameterType)))
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				checkUnsafeArguments(node.Arguments(), node.Expression(), node)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				checkUnsafeArguments(node.Arguments(), node.Expression(), node)
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				expr := node.AsTaggedTemplateExpression()
				template := expr.Template
				if ast.IsTemplateExpression(template) {
					checkUnsafeArguments(utils.Map(template.AsTemplateExpression().TemplateSpans.Nodes, func(span *ast.Node) *ast.Node {
						return span.Expression()
					}), expr.Tag, node)
				}
			},
		}
	},
})
