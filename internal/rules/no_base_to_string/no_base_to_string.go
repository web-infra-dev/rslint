package no_base_to_string

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func certaintyToString(certainty usefulness) string {
	switch certainty {
	case usefulnessAlways:
		return "always"
	case usefulnessNever:
		return "always"
	case usefulnessSometimes:
		return "always"
	default:
		panic("unkown certainty")
	}
}

func buildBaseArrayJoinMessage(name string, certainty usefulness) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "baseArrayJoin",
		Description: fmt.Sprintf("Using `join()` for %v %v use Object's default stringification format ('[object Object]') when stringified.", name, certaintyToString(certainty)),
	}
}
func buildBaseToStringMessage(name string, certainty usefulness) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "baseToString",
		Description: fmt.Sprintf("'%v' %v use Object's default stringification format ('[object Object]') when stringified.", name, certaintyToString(certainty)),
	}
}

type NoBaseToStringOptions struct {
	IgnoredTypeNames []string
}

type usefulness uint32

const (
	usefulnessAlways usefulness = iota
	usefulnessNever
	usefulnessSometimes
)

var NoBaseToStringRule = rule.Rule{
	Name: "no-base-to-string",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(NoBaseToStringOptions)
		if !ok {
			opts = NoBaseToStringOptions{
				IgnoredTypeNames: []string{"Error", "RegExp", "URL", "URLSearchParams"},
			}
		}

		var collectToStringCertainty func(
			t *checker.Type,
			visited []*checker.Type,
		) usefulness
		var collectJoinCertainty func(
			t *checker.Type,
			visited []*checker.Type,
		) usefulness

		checkExpression := func(node *ast.Expression, t *checker.Type) {
			// TODO(port): boolean, null, etc?
			if ast.IsLiteralExpression(node) {
				return
			}

			if t == nil {
				t = ctx.TypeChecker.GetTypeAtLocation(node)
			}

			certainty := collectToStringCertainty(
				t,
				[]*checker.Type{},
			)
			if certainty == usefulnessAlways {
				return
			}

			ctx.ReportNode(node, buildBaseToStringMessage(ctx.SourceFile.Text()[node.Pos():node.End()], certainty))
		}

		checkExpressionForArrayJoin := func(
			node *ast.Node,
			t *checker.Type,
		) {
			certainty := collectJoinCertainty(t, []*checker.Type{})

			if certainty == usefulnessAlways {
				return
			}

			ctx.ReportNode(node, buildBaseArrayJoinMessage(ctx.SourceFile.Text()[node.Pos():node.End()], certainty))
		}

		collectUnionTypeCertainty := func(
			t *checker.Type,
			collectSubTypeCertainty func(t *checker.Type) usefulness,
		) usefulness {
			certainties := utils.Map(utils.UnionTypeParts(t), collectSubTypeCertainty)

			if utils.Every(certainties, func(c usefulness) bool { return c == usefulnessNever }) {
				return usefulnessNever
			}

			if utils.Every(certainties, func(c usefulness) bool { return c == usefulnessAlways }) {
				return usefulnessAlways
			}

			return usefulnessSometimes
		}

		collectIntersectionTypeCertainty := func(
			t *checker.Type,
			collectSubTypeCertainty func(t *checker.Type) usefulness,
		) usefulness {
			if utils.Some(utils.IntersectionTypeParts(t), func(t *checker.Type) bool { return collectSubTypeCertainty(t) == usefulnessAlways }) {
				return usefulnessAlways
			}

			return usefulnessNever
		}

		collectTupleCertainty := func(
			t *checker.Type,
			visited []*checker.Type,
		) usefulness {
			typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, t)
			certainties := utils.Map(typeArgs, func(t *checker.Type) usefulness {
				return collectToStringCertainty(t, visited)
			})

			if utils.Some(certainties, func(c usefulness) bool { return c == usefulnessNever }) {
				return usefulnessNever
			}

			if utils.Some(certainties, func(c usefulness) bool { return c == usefulnessSometimes }) {
				return usefulnessSometimes
			}

			return usefulnessAlways
		}

		collectArrayCertainty := func(
			t *checker.Type,
			visited []*checker.Type,
		) usefulness {
			elemType := utils.GetNumberIndexType(ctx.TypeChecker, t)
			if elemType == nil {
				panic("array should have number index type")
			}
			return collectToStringCertainty(elemType, visited)
		}

		collectJoinCertainty = func(
			t *checker.Type,
			visited []*checker.Type,
		) usefulness {
			if utils.IsUnionType(t) {
				return collectUnionTypeCertainty(t, func(t *checker.Type) usefulness {
					return collectJoinCertainty(t, visited)
				})
			}

			if utils.IsIntersectionType(t) {
				return collectIntersectionTypeCertainty(t, func(t *checker.Type) usefulness {
					return collectJoinCertainty(t, visited)
				})
			}

			if checker.IsTupleType(t) {
				return collectTupleCertainty(t, visited)
			}

			if checker.Checker_isArrayType(ctx.TypeChecker, t) {
				return collectArrayCertainty(t, visited)
			}

			return usefulnessAlways
		}

		collectToStringCertainty = func(
			t *checker.Type,
			visited []*checker.Type,
		) usefulness {
			if slices.Contains(visited, t) {
				// don't report if this is a self referencing array or tuple type
				return usefulnessAlways
			}

			if utils.IsTypeParameter(t) {
				constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, t)
				if constraint != nil {
					return collectToStringCertainty(constraint, visited)
				}
				// unconstrained generic means `unknown`
				return usefulnessAlways
			}

			// the Boolean type definition missing toString()
			if utils.IsTypeFlagSet(t, checker.TypeFlagsBooleanLike) {
				return usefulnessAlways
			}

			if slices.Contains(opts.IgnoredTypeNames, utils.GetTypeName(ctx.TypeChecker, t)) {
				return usefulnessAlways
			}

			if utils.IsIntersectionType(t) {
				return collectIntersectionTypeCertainty(t, func(t *checker.Type) usefulness {
					return collectToStringCertainty(t, visited)
				})
			}

			if utils.IsUnionType(t) {
				return collectUnionTypeCertainty(t, func(t *checker.Type) usefulness {
					return collectToStringCertainty(t, visited)
				})
			}

			if checker.IsTupleType(t) {
				return collectTupleCertainty(t, append(visited, t))
			}

			if checker.Checker_isArrayType(ctx.TypeChecker, t) {
				return collectArrayCertainty(t, append(visited, t))
			}

			toString := checker.Checker_getPropertyOfType(ctx.TypeChecker, t, "toString")
			if toString == nil {
				toString = checker.Checker_getPropertyOfType(ctx.TypeChecker, t, "toLocaleString")
			}
			if toString == nil {
				// e.g. any/unknown
				return usefulnessAlways
			}

			declarations := toString.Declarations

			if declarations == nil || len(declarations) != 1 {
				// If there are multiple declarations, at least one of them must not be
				// the default object toString.
				//
				// This may only matter for older versions of TS
				// see https://github.com/typescript-eslint/typescript-eslint/issues/8585
				return usefulnessAlways
			}

			declaration := declarations[0]
			isBaseToString := ast.IsInterfaceDeclaration(declaration.Parent) && declaration.Parent.AsInterfaceDeclaration().Name().Text() == "Object"

			if isBaseToString {
				return usefulnessNever
			}

			return usefulnessAlways
		}

		isBuiltInStringCall := func(node *ast.CallExpression) bool {
			if ast.IsIdentifier(node.Expression) && node.Expression.AsIdentifier().Text == "String" && len(node.Arguments.Nodes) > 0 {
				tt := ctx.TypeChecker.GetTypeAtLocation(node.Expression)
				s := utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, tt, "String")
				sc := utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, tt, "StringConstructor")
				return s || sc
				// TODO(port-scopemanager)
				// const scope = context.sourceCode.getScope(node);
				// // eslint-disable-next-line @typescript-eslint/internal/prefer-ast-types-enum
				// const variable = scope.set.get('String');
				// return !variable?.defs.length;
				return true
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expr := node.AsBinaryExpression()
				if expr.OperatorToken.Kind != ast.KindPlusToken && expr.OperatorToken.Kind != ast.KindPlusEqualsToken {
					return
				}
				leftType := ctx.TypeChecker.GetTypeAtLocation(expr.Left)
				rightType := ctx.TypeChecker.GetTypeAtLocation(expr.Right)

				if utils.GetTypeName(ctx.TypeChecker, leftType) == "string" {
					checkExpression(expr.Right, rightType)
				} else if utils.GetTypeName(ctx.TypeChecker, rightType) == "string" && expr.Left.Kind != ast.KindPrivateIdentifier {
					checkExpression(expr.Left, leftType)
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {

				callExpr := node.AsCallExpression()
				if isBuiltInStringCall(callExpr) && callExpr.Arguments.Nodes[0].Kind != ast.KindSpreadElement {
					checkExpression(callExpr.Arguments.Nodes[0], nil)
					return
				}

				if ast.IsPropertyAccessExpression(callExpr.Expression) {
					memberExpr := callExpr.Expression.AsPropertyAccessExpression()
					propertyName := memberExpr.Name().Text()
					if propertyName == "join" {
						t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, memberExpr.Expression)
						checkExpressionForArrayJoin(memberExpr.Expression, t)
						return
					} else if propertyName == "toLocaleString" || propertyName == "toString" {
						checkExpression(memberExpr.Expression, nil)
						return
					}
				}
			},
			ast.KindTemplateExpression: func(node *ast.Node) {

				if ast.IsTaggedTemplateExpression(node.Parent) {
					return
				}
				for _, span := range node.AsTemplateExpression().TemplateSpans.Nodes {
					checkExpression(span.Expression(), nil)
				}
			},
		}
	},
}
