package no_unsafe_argument

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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
	Index         int
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
	var paramTypes []*checker.Type
	var restT restType
	hasRestType := false

	for i, param := range parameters {
		t := typeChecker.GetTypeOfSymbolAtLocation(param, node)

		if len(param.Declarations) != 0 {
			decl := param.Declarations[0]
			if utils.IsRestParameterDeclaration(decl) {
				// is a rest param
				if checker.Checker_isArrayType(typeChecker, t) {
					restT = restType{
						Type:  checker.Checker_getTypeArguments(typeChecker, t)[0],
						Index: i,
						Kind:  restTypeKindArray,
					}
				} else if checker.IsTupleType(t) {
					restT = restType{
						Index:         i,
						Kind:          restTypeKindTuple,
						TypeArguments: checker.Checker_getTypeArguments(typeChecker, t),
					}
				} else {
					restT = restType{
						Type:  t,
						Index: i,
						Kind:  restTypeKindOther,
					}
				}
				hasRestType = true
				break
			}
		}

		if paramTypes == nil {
			paramTypes = make([]*checker.Type, 0, len(parameters))
		}
		paramTypes = append(paramTypes, t)
	}

	return functionSignature{
		paramTypes:  paramTypes,
		restType:    restT,
		hasRestType: hasRestType,
	}, true
}

type functionSignature struct {
	hasConsumedArguments bool
	parameterTypeIndex   int

	paramTypes  []*checker.Type
	restType    restType
	hasRestType bool
}

func (s *functionSignature) consumeRemainingArguments() {
	s.hasConsumedArguments = true
}

func (s *functionSignature) getNextParameterType() *checker.Type {
	index := s.parameterTypeIndex
	s.parameterTypeIndex += 1

	if index >= len(s.paramTypes) || s.hasConsumedArguments {
		if !s.hasRestType {
			return nil
		}

		switch s.restType.Kind {
		case restTypeKindTuple:
			typeArguments := s.restType.TypeArguments
			if len(typeArguments) == 0 {
				return nil
			}
			if s.hasConsumedArguments {
				// all types consumed by a rest - just assume it's the last type
				// there is one edge case where this is wrong, but we ignore it because
				// it's rare and really complicated to handle
				// eg: function foo(...a: [number, ...string[], number])
				return typeArguments[len(typeArguments)-1]
			}

			typeIndex := index - s.restType.Index
			if typeIndex >= len(typeArguments) {
				return typeArguments[len(typeArguments)-1]
			}

			return typeArguments[typeIndex]
		case restTypeKindArray, restTypeKindOther:
			return s.restType.Type
		}
	}
	return s.paramTypes[index]
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

			// ignore any-typed calls as these are caught by no-unsafe-call
			if utils.IsTypeAnyType(ctx.TypeChecker.GetTypeAtLocation(callee)) {
				return
			}

			signature, ok := newFunctionSignature(ctx.TypeChecker, node)
			if !ok {
				panic("Expected to a signature resolved")
			}

			if ast.IsTaggedTemplateExpression(node) {
				// Consumes the first parameter (TemplateStringsArray) of the function called with TaggedTemplateExpression.
				signature.getNextParameterType()
			}

			for _, argument := range args {
				switch argument.Kind {
				// spreads consume
				case ast.KindSpreadElement:
					spreadArgType := ctx.TypeChecker.GetTypeAtLocation(argument.Expression())

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
							parameterType := signature.getNextParameterType()
							if parameterType == nil {
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
							signature.consumeRemainingArguments()
						}

					} else
					//nolint:staticcheck // FIXME: todo
					{
						// something that's iterable
						// handling this will be pretty complex - so we ignore it for now
						// TODO - handle generic iterable case
					}

				default:
					parameterType := signature.getNextParameterType()
					if parameterType == nil {
						continue
					}

					reportNode := ast.SkipParentheses(argument)
					argumentType := ctx.TypeChecker.GetTypeAtLocation(reportNode)
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
