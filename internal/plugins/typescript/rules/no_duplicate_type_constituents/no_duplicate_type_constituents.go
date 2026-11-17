package no_duplicate_type_constituents

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_duplicate_type_constituents.schema.json
var schemaJSON []byte

func buildDuplicateMessage(unionOrIntersection unionOrIntersection, previous string) rule.RuleMessage {
	var msg string
	if unionOrIntersection == unionOrIntersection_Intersection {
		msg = "Intersection"
	} else {
		msg = "Union"
	}
	return rule.RuleMessage{
		Id:          "duplicate",
		Description: fmt.Sprintf("%v type constituent is duplicated with %v.", msg, previous),
	}
}
func buildUnnecessaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessary",
		Description: "Explicit undefined is unnecessary on an optional parameter.",
	}
}

type unionOrIntersection uint8

const (
	unionOrIntersection_Union unionOrIntersection = iota
	unionOrIntersection_Intersection
)

type NoDuplicateTypeConstituentsOptions struct {
	IgnoreIntersections bool
	IgnoreUnions        bool
}

func parseOptions(options []any) NoDuplicateTypeConstituentsOptions {
	opts := NoDuplicateTypeConstituentsOptions{
		IgnoreIntersections: false,
		IgnoreUnions:        false,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if value, ok := optsMap["ignoreIntersections"].(bool); ok {
		opts.IgnoreIntersections = value
	}
	if value, ok := optsMap["ignoreUnions"].(bool); ok {
		opts.IgnoreUnions = value
	}
	return opts
}

var NoDuplicateTypeConstituentsRule = rule.CreateRule(rule.Rule{
	Name:             "no-duplicate-type-constituents",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		unwindedParentType := func(node *ast.Node, kind ast.Kind) *ast.Node {
			for {
				node = node.Parent
				if node == nil {
					return nil
				}

				if node.Kind == kind {
					return node
				}

				if node.Kind != ast.KindParenthesizedType {
					return nil
				}
			}
		}

		report := func(
			message rule.RuleMessage,
			constituentNode *ast.Node,
		) {
			constituentNode = ast.SkipTypeParentheses(constituentNode)
			outer := constituentNode
			for outer.Parent.Kind == ast.KindParenthesizedType {
				outer = outer.Parent
			}
			textRange := utils.TrimNodeTextRange(ctx.SourceFile, constituentNode)
			// ESTree omits wrapping parentheses. Upstream starts at the inner
			// node, but includes the closing parentheses in the diagnostic.
			textRange = core.NewTextRange(textRange.Pos(), outer.End())
			ctx.ReportRangeWithDeferredFixes(textRange, message, func() []rule.RuleFix {
				parent := outer.Parent
				var types []*ast.Node
				if parent.Kind == ast.KindUnionType {
					types = parent.AsUnionTypeNode().Types.Nodes
				} else {
					types = parent.AsIntersectionTypeNode().Types.Nodes
				}
				index := 0
				for types[index] != outer {
					index++
				}
				// Remove tokens separately to preserve whitespace and comments.
				fixes := []rule.RuleFix{rule.RuleFixRemoveRange(utils.TrimNodeTextRange(ctx.SourceFile, constituentNode))}
				scanTokens := func(start, end int) {
					scanner := scanner.GetScannerForSourceFile(ctx.SourceFile, start)
					for ; scanner.TokenStart() < end && scanner.Token() != ast.KindEndOfFile; scanner.Scan() {
						fixes = append(fixes, rule.RuleFixRemoveRange(scanner.TokenRange()))
					}
				}
				scanTokens(outer.Pos(), utils.TrimNodeTextRange(ctx.SourceFile, constituentNode).Pos())
				scanTokens(constituentNode.End(), outer.End())
				if index > 0 {
					scanTokens(types[index-1].End(), utils.TrimNodeTextRange(ctx.SourceFile, outer).Pos())
				} else {
					// A leading | or & belongs to the union/intersection too.
					beforeCount := len(fixes)
					scanTokens(parent.Pos(), utils.TrimNodeTextRange(ctx.SourceFile, outer).Pos())
					if len(fixes) == beforeCount {
						scanTokens(outer.End(), utils.TrimNodeTextRange(ctx.SourceFile, types[1]).Pos())
					}
				}
				return fixes
			})
		}

		var checkDuplicateRecursively func(
			unionOrIntersection unionOrIntersection,
			constituentNode *ast.Node,
			cachedTypeMap map[*checker.Type]*ast.Node,
			uniqueConstituents *[]*ast.Node,
			forEachNodeType func(t *checker.Type, node *ast.Node),
		)
		checkDuplicateRecursively = func(
			unionOrIntersection unionOrIntersection,
			constituentNode *ast.Node,
			cachedTypeMap map[*checker.Type]*ast.Node,
			uniqueConstituents *[]*ast.Node,
			forEachNodeType func(t *checker.Type, node *ast.Node),
		) {
			constituentNode = ast.SkipTypeParentheses(constituentNode)
			reportDuplicate := func(previous *ast.Node) {
				previousRange := utils.TrimNodeTextRange(ctx.SourceFile, previous)
				report(buildDuplicateMessage(unionOrIntersection, ctx.SourceFile.Text()[previousRange.Pos():previousRange.End()]), constituentNode)
			}
			for _, previous := range *uniqueConstituents {
				if sameTypeSyntax(ctx.SourceFile, previous, constituentNode) {
					reportDuplicate(previous)
					return
				}
			}
			t := ctx.TypeChecker.GetTypeAtLocation(constituentNode)
			if utils.IsIntrinsicErrorType(t) {
				return
			}

			duplicatedPreviousNode, duplicatedPrevious := cachedTypeMap[t]

			if duplicatedPrevious {
				reportDuplicate(duplicatedPreviousNode)
				return
			}

			if forEachNodeType != nil {
				forEachNodeType(t, constituentNode)
			}
			cachedTypeMap[t] = constituentNode
			*uniqueConstituents = append(*uniqueConstituents, constituentNode)

			var types []*ast.Node
			withoutParens := ast.SkipTypeParentheses(constituentNode)
			if unionOrIntersection == unionOrIntersection_Union && withoutParens.Kind == ast.KindUnionType {
				types = withoutParens.AsUnionTypeNode().Types.Nodes
			} else if unionOrIntersection == unionOrIntersection_Intersection && withoutParens.Kind == ast.KindIntersectionType {
				types = withoutParens.AsIntersectionTypeNode().Types.Nodes
			} else {
				return
			}

			for _, constituent := range types {
				checkDuplicateRecursively(unionOrIntersection, constituent, cachedTypeMap, uniqueConstituents, forEachNodeType)
			}
		}

		checkDuplicate := func(
			node *ast.Node,
			forEachNodeType func(
				constituentNodeType *checker.Type,
				constituentNode *ast.Node,
			),
		) {
			cachedTypeMap := map[*checker.Type]*ast.Node{}
			uniqueConstituents := []*ast.Node{}

			var unionOrIntersection unionOrIntersection
			var types []*ast.Node
			switch node.Kind {
			case ast.KindIntersectionType:
				unionOrIntersection = unionOrIntersection_Intersection
				types = node.AsIntersectionTypeNode().Types.Nodes
			case ast.KindUnionType:
				unionOrIntersection = unionOrIntersection_Union
				types = node.AsUnionTypeNode().Types.Nodes
			default:
				panic(fmt.Sprintf("expected union or intersection, got %v", node.Kind))
			}

			for _, t := range types {
				checkDuplicateRecursively(
					unionOrIntersection,
					t,
					cachedTypeMap,
					&uniqueConstituents,
					forEachNodeType,
				)
			}
		}

		ruleListeners := make(rule.RuleListeners, 2)

		if !opts.IgnoreIntersections {
			ruleListeners[ast.KindIntersectionType] = func(node *ast.Node) {
				if unwindedParentType(node, ast.KindIntersectionType) != nil {
					return
				}
				checkDuplicate(node, nil)
			}
		}

		if !opts.IgnoreUnions {
			ruleListeners[ast.KindUnionType] = func(node *ast.Node) {
				if unwindedParentType(node, ast.KindUnionType) != nil {
					return
				}

				checkDuplicate(node, func(constituentNodeType *checker.Type, constituentNode *ast.Node) {
					outer := node
					for outer.Parent.Kind == ast.KindParenthesizedType {
						outer = outer.Parent
					}
					if !ast.IsParameterDeclaration(outer.Parent) {
						return
					}
					param := outer.Parent.AsParameterDeclaration()
					if param.QuestionToken == nil || !ast.IsIdentifier(param.Name()) || ast.IsParameterPropertyDeclaration(outer.Parent, outer.Parent.Parent) {
						return
					}
					if utils.IsTypeFlagSet(constituentNodeType, checker.TypeFlagsUndefined) {
						report(buildUnnecessaryMessage(), constituentNode)
					}
				})
			}
		}

		return ruleListeners
	},
})

// Compare the syntax before consulting type identity, just as upstream does.
// Raw literal spelling is part of ESTree's Literal nodes; punctuation between
// members and redundant parentheses are not.
func sameTypeSyntax(sf *ast.SourceFile, left, right *ast.Node) bool {
	left = ast.SkipTypeParentheses(ast.SkipParentheses(left))
	right = ast.SkipTypeParentheses(ast.SkipParentheses(right))
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindRegularExpressionLiteral,
		ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
		a, b := utils.TrimNodeTextRange(sf, left), utils.TrimNodeTextRange(sf, right)
		return sf.Text()[a.Pos():a.End()] == sf.Text()[b.Pos():b.End()]
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		return left.Text() == right.Text()
	case ast.KindTypeParameter:
		// Constraints and defaults occupy different ESTree fields even when
		// their child types have identical syntax.
		a, b := left.AsTypeParameterDeclaration(), right.AsTypeParameterDeclaration()
		if (a.Constraint == nil) != (b.Constraint == nil) || (a.DefaultType == nil) != (b.DefaultType == nil) {
			return false
		}
	case ast.KindImportType:
		if left.AsImportTypeNode().IsTypeOf != right.AsImportTypeNode().IsTypeOf {
			return false
		}
	case ast.KindTypeOperator:
		if left.AsTypeOperatorNode().Operator != right.AsTypeOperatorNode().Operator {
			return false
		}
	case ast.KindPrefixUnaryExpression:
		if left.AsPrefixUnaryExpression().Operator != right.AsPrefixUnaryExpression().Operator {
			return false
		}
	}
	var a, b []*ast.Node
	left.ForEachChild(func(node *ast.Node) bool { a = append(a, node); return false })
	right.ForEachChild(func(node *ast.Node) bool { b = append(b, node); return false })
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !sameTypeSyntax(sf, a[i], b[i]) {
			return false
		}
	}
	return true
}
