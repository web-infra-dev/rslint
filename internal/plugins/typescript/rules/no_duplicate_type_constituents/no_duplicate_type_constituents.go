package no_duplicate_type_constituents

import (
	_ "embed"
	"fmt"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
		syntax := newTypeSyntaxIndex(ctx.SourceFile)
		// A fix must not remove the last member of a surviving group. Reserve
		// removals only inside fix builders, after inline-disable suppression
		// and edit-demand checks, so suppressed diagnostics do not consume them.
		remaining := make(map[*ast.Node]int)
		reserved := make(map[*ast.Node]struct{})
		comments := ctx.Comments
		if comments == nil {
			comments = rule.NewCommentStore(ctx.SourceFile)
		}

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
			previous *ast.Node,
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
			parent := outer.Parent
			var types []*ast.Node
			if parent.Kind == ast.KindUnionType {
				types = parent.AsUnionTypeNode().Types.Nodes
			} else {
				types = parent.AsIntersectionTypeNode().Types.Nodes
			}
			fixNode := constituentNode
			// A preceding pass may leave a one-member group with a leading
			// separator. Its anonymous member need not share a checker type
			// with the earlier occurrence. Remove the redundant group through
			// its enclosing same-kind group, but never remove the previous
			// occurrence that establishes the duplication.
			for len(types) == 1 {
				if previous != nil && parent.Pos() <= previous.Pos() && previous.End() <= parent.End() {
					break
				}
				group := parent
				for group.Parent.Kind == ast.KindParenthesizedType {
					group = group.Parent
				}
				if group.Parent.Kind != parent.Kind {
					break
				}
				fixNode, outer, parent = parent, group, group.Parent
				if parent.Kind == ast.KindUnionType {
					types = parent.AsUnionTypeNode().Types.Nodes
				} else {
					types = parent.AsIntersectionTypeNode().Types.Nodes
				}
			}
			ctx.ReportRangeWithDeferredFixes(textRange, message, func() []rule.RuleFix {
				// A singleton wrapper can contain comments between its leading
				// separator and its member. Keep those comments by deleting
				// wrapper tokens separately from the actual type body.
				body := fixNode
				for {
					body = ast.SkipTypeParentheses(body)
					if body.Kind == ast.KindUnionType && len(body.AsUnionTypeNode().Types.Nodes) == 1 {
						body = body.AsUnionTypeNode().Types.Nodes[0]
					} else if body.Kind == ast.KindIntersectionType && len(body.AsIntersectionTypeNode().Types.Nodes) == 1 {
						body = body.AsIntersectionTypeNode().Types.Nodes[0]
					} else {
						break
					}
				}
				bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
				// Directives inside the type body itself cannot be discarded
				// safely. Keep the diagnostic without a fix in that case.
				allComments := comments.All()
				first := sort.Search(len(allComments), func(i int) bool { return allComments[i].Pos() >= bodyRange.Pos() })
				for _, comment := range allComments[first:] {
					if comment.Pos() >= bodyRange.End() {
						break
					}
					if utils.IsDirectiveComment(comment.Kind, ecmascript.StringTrim(utils.CommentValue(ctx.SourceFile.Text(), comment))) {
						return nil
					}
				}
				if _, alreadyReserved := reserved[outer]; !alreadyReserved {
					count, seen := remaining[parent]
					if !seen {
						count = len(types)
					}
					if count <= 1 {
						return nil
					}
					remaining[parent] = count - 1
					reserved[outer] = struct{}{}
				}
				index := 0
				for types[index] != outer {
					index++
				}
				// Remove tokens separately to preserve whitespace and comments.
				fixes := []rule.RuleFix{rule.RuleFixRemoveRange(bodyRange)}
				scanTokens := func(start, end int) {
					scanner := scanner.GetScannerForSourceFile(ctx.SourceFile, start)
					for ; scanner.TokenStart() < end && scanner.Token() != ast.KindEndOfFile; scanner.Scan() {
						fixes = append(fixes, rule.RuleFixRemoveRange(scanner.TokenRange()))
					}
				}
				scanTokens(outer.Pos(), bodyRange.Pos())
				scanTokens(body.End(), outer.End())
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
			cachedTypeMap map[*checker.Type][]*ast.Node,
			uniqueConstituents map[uint64][]*ast.Node,
			forEachNodeType func(t *checker.Type, node *ast.Node),
		)
		checkDuplicateRecursively = func(
			unionOrIntersection unionOrIntersection,
			constituentNode *ast.Node,
			cachedTypeMap map[*checker.Type][]*ast.Node,
			uniqueConstituents map[uint64][]*ast.Node,
			forEachNodeType func(t *checker.Type, node *ast.Node),
		) {
			constituentNode = ast.SkipTypeParentheses(constituentNode)
			reportDuplicate := func(previous *ast.Node) {
				previousRange := utils.TrimNodeTextRange(ctx.SourceFile, previous)
				report(buildDuplicateMessage(unionOrIntersection, ctx.SourceFile.Text()[previousRange.Pos():previousRange.End()]), constituentNode, previous)
			}
			fingerprint := syntax.hash(constituentNode)
			if previous := syntax.find(uniqueConstituents[fingerprint], constituentNode); previous != nil {
				reportDuplicate(previous)
				return
			}
			t := ctx.TypeChecker.GetTypeAtLocation(constituentNode)
			if utils.IsIntrinsicErrorType(t) {
				return
			}

			for _, previous := range cachedTypeMap[t] {
				// A union such as (string | never) has the same checker type
				// as its string child. The containing group is not another
				// occurrence: removing its child would change the type.
				if previous.Pos() <= constituentNode.Pos() && constituentNode.End() <= previous.End() {
					continue
				}
				reportDuplicate(previous)
				return
			}

			if forEachNodeType != nil {
				forEachNodeType(t, constituentNode)
			}
			cachedTypeMap[t] = append(cachedTypeMap[t], constituentNode)
			uniqueConstituents[fingerprint] = append(uniqueConstituents[fingerprint], constituentNode)

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
			cachedTypeMap := map[*checker.Type][]*ast.Node{}
			uniqueConstituents := make(map[uint64][]*ast.Node)

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
					uniqueConstituents,
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
						report(buildUnnecessaryMessage(), constituentNode, nil)
					}
				})
			}
		}

		return ruleListeners
	},
})
