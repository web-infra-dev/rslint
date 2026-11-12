package no_base_to_string

import (
	_ "embed"
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_base_to_string.schema.json
var schemaJSON []byte

func certaintyToString(certainty usefulness) string {
	switch certainty {
	case usefulnessAlways:
		return "always"
	case usefulnessNever:
		return "will"
	case usefulnessSometimes:
		return "may"
	default:
		panic("unknown certainty")
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

func parseOptions(options []any) NoBaseToStringOptions {
	opts := NoBaseToStringOptions{
		IgnoredTypeNames: []string{"Error", "RegExp", "URL", "URLSearchParams"},
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if raw, ok := optsMap["ignoredTypeNames"].([]any); ok {
		opts.IgnoredTypeNames = utils.ToStringSlice(raw)
	}
	return opts
}

type usefulness uint32

const (
	usefulnessAlways usefulness = iota
	usefulnessNever
	usefulnessSometimes
)

var NoBaseToStringRule = rule.CreateRule(rule.Rule{
	Name:             "no-base-to-string",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		var collectToStringCertainty func(
			t *checker.Type,
			visited []*checker.Type,
		) usefulness
		var collectJoinCertainty func(
			t *checker.Type,
			visited []*checker.Type,
		) usefulness

		checkExpression := func(node *ast.Expression, t *checker.Type) {
			// ESTree omits ParenthesizedExpression nodes and leading trivia.
			node = ast.SkipParentheses(node)
			if utils.IsESTreeLiteralKind(node.Kind) {
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

			ctx.ReportNode(node, buildBaseToStringMessage(utils.TrimmedNodeText(ctx.SourceFile, node), certainty))
		}

		checkExpressionForArrayJoin := func(
			node *ast.Node,
			t *checker.Type,
		) {
			node = ast.SkipParentheses(node)
			certainty := collectJoinCertainty(t, []*checker.Type{})

			if certainty == usefulnessAlways {
				return
			}

			ctx.ReportNode(node, buildBaseArrayJoinMessage(utils.TrimmedNodeText(ctx.SourceFile, node), certainty))
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
			callee := ast.SkipParentheses(node.Expression)
			if ast.IsIdentifier(callee) && callee.Text() == "String" && len(node.Arguments.Nodes) > 0 {
				// A script's local String declaration can merge with the lib symbol.
				// Check the file's lexical binding before consulting its inferred type.
				if ctx.Refs.ResolveInFile(callee) != nil {
					return false
				}
				tt := ctx.TypeChecker.GetTypeAtLocation(callee)
				s := utils.IsBuiltinSymbolLike(ctx.Program(), ctx.TypeChecker, tt, "String")
				sc := utils.IsBuiltinSymbolLike(ctx.Program(), ctx.TypeChecker, tt, "StringConstructor")
				return s || sc
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

				callee := ast.SkipParentheses(callExpr.Expression)
				var object, property *ast.Node
				if ast.IsPropertyAccessExpression(callee) {
					memberExpr := callee.AsPropertyAccessExpression()
					object, property = memberExpr.Expression, memberExpr.Name()
				} else if ast.IsElementAccessExpression(callee) {
					memberExpr := callee.AsElementAccessExpression()
					object, property = memberExpr.Expression, ast.SkipParentheses(memberExpr.ArgumentExpression)
				}
				// Upstream selects Identifier properties, including computed identifiers,
				// but not string literals or private names.
				if property != nil && ast.IsIdentifier(property) {
					switch property.Text() {
					case "join":
						t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, object)
						checkExpressionForArrayJoin(object, t)
						return
					case "toLocaleString", "toString":
						checkExpression(object, nil)
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
})
