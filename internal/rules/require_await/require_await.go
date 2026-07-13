package require_await

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildMissingAwaitMessage(node *ast.Node) rule.RuleMessage {
	name := utils.UpperCaseFirstASCII(utils.GetFunctionNameWithKindCore(node))
	return rule.RuleMessage{
		Id:          "missingAwait",
		Description: name + " has no 'await' expression.",
		Data:        map[string]string{"name": name},
	}
}

func buildRemoveAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeAsync",
		Description: "Remove 'async'.",
	}
}

type scopeFrame struct {
	node     *ast.Node
	hasAwait bool
}

type asyncKeywordInfo struct {
	tokenRange  core.TextRange
	removeRange core.TextRange
}

func findAsyncKeyword(sourceFile *ast.SourceFile, node *ast.Node) (asyncKeywordInfo, bool) {
	mods := node.Modifiers()
	if mods == nil {
		return asyncKeywordInfo{}, false
	}
	for _, mod := range mods.Nodes {
		if mod == nil || mod.Kind != ast.KindAsyncKeyword {
			continue
		}
		tokenRange := utils.TrimNodeTextRange(sourceFile, mod)
		removeEnd := utils.SkipLeadingWhitespace(sourceFile.Text(), tokenRange.End(), len(sourceFile.Text()))
		return asyncKeywordInfo{
			tokenRange:  tokenRange,
			removeRange: core.NewTextRange(tokenRange.Pos(), removeEnd),
		}, true
	}
	return asyncKeywordInfo{}, false
}

func shouldReplaceAsyncWithSemicolon(sourceFile *ast.SourceFile, node *ast.Node, asyncInfo asyncKeywordInfo) bool {
	nextToken, ok := utils.TokenAtOrAfter(sourceFile, asyncInfo.tokenRange.End())
	if !ok {
		return false
	}

	if nextToken.Kind == ast.KindOpenParenToken &&
		utils.IsStartOfExpressionStatement(sourceFile, node) &&
		utils.NeedsPrecedingSemicolon(sourceFile, node) {
		return true
	}

	if node.Kind != ast.KindMethodDeclaration {
		return false
	}
	return utils.NeedsClassMemberLeadingSemicolon(
		sourceFile,
		ast.GetContainingClass(node),
		node,
		nextToken,
		utils.ClassMemberLeadingSemicolonOptions{},
	)
}

func reportMissingAwait(ctx rule.RuleContext, node *ast.Node) {
	asyncInfo, ok := findAsyncKeyword(ctx.SourceFile, node)
	if !ok {
		ctx.ReportRange(utils.GetFunctionHeadLoc(ctx.SourceFile, node), buildMissingAwaitMessage(node))
		return
	}

	replacement := ""
	if shouldReplaceAsyncWithSemicolon(ctx.SourceFile, node, asyncInfo) {
		replacement = ";"
	}

	ctx.ReportRangeWithSuggestions(
		utils.GetFunctionHeadLoc(ctx.SourceFile, node),
		buildMissingAwaitMessage(node),
		rule.RuleSuggestion{
			Message:  buildRemoveAsyncMessage(),
			FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(asyncInfo.removeRange, replacement)},
		},
	)
}

// https://eslint.org/docs/latest/rules/require-await
var RequireAwaitRule = rule.Rule{
	Name: "require-await",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		stack := make([]scopeFrame, 0, 8)

		enterFunction := func(node *ast.Node) {
			stack = append(stack, scopeFrame{node: node})
		}

		exitFunction := func(node *ast.Node) {
			n := len(stack)
			if n == 0 {
				return
			}
			top := stack[n-1]
			stack = stack[:n-1]

			flags := ast.GetFunctionFlags(node)
			if flags&ast.FunctionFlagsAsync == 0 ||
				flags&ast.FunctionFlagsGenerator != 0 ||
				top.hasAwait ||
				!utils.HasNonEmptyFunctionBody(node) {
				return
			}

			reportMissingAwait(ctx, node)
		}

		markAwaitAt := func(node *ast.Node) {
			for i := len(stack) - 1; i >= 0; i-- {
				bp, be, ok := utils.BodyLikeRange(stack[i].node)
				if !ok {
					continue
				}
				if node.Pos() >= bp && node.End() <= be {
					stack[i].hasAwait = true
					return
				}
			}
		}

		listeners := rule.RuleListeners{
			ast.KindAwaitExpression: markAwaitAt,
			ast.KindForOfStatement: func(node *ast.Node) {
				if node.AsForInOrOfStatement().AwaitModifier != nil {
					markAwaitAt(node)
				}
			},
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if ast.IsVarAwaitUsing(node) {
					markAwaitAt(node)
				}
			},
		}

		for _, kind := range []ast.Kind{
			ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindMethodDeclaration,
			ast.KindArrowFunction,
		} {
			listeners[kind] = enterFunction
			listeners[rule.ListenerOnExit(kind)] = exitFunction
		}

		return listeners
	},
}
