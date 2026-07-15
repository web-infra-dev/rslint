package no_unused_labels

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type labelScope struct {
	label string
	used  bool
	upper *labelScope
}

func buildUnusedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unused",
		Description: fmt.Sprintf("'%s:' is defined but never used.", name),
		Data:        map[string]string{"name": name},
	}
}

func isFunctionBodyBlock(node *ast.Node) bool {
	return node != nil &&
		node.Kind == ast.KindBlock &&
		node.Parent != nil &&
		ast.IsFunctionLike(node.Parent)
}

// A removable label before a string-like expression can create a directive
// prologue entry after this or another fix pass, so ESLint intentionally skips
// the fix in program and function-body containers.
func isPotentialDirectiveLabel(node *ast.Node) bool {
	ls := node.AsLabeledStatement()
	if ls == nil || ls.Statement == nil || ls.Statement.Kind != ast.KindExpressionStatement {
		return false
	}

	ancestor := ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		return n.Kind != ast.KindLabeledStatement
	})
	if ancestor == nil || (ancestor.Kind != ast.KindSourceFile && !isFunctionBodyBlock(ancestor)) {
		return false
	}

	expr := ast.SkipParentheses(ls.Statement.AsExpressionStatement().Expression)
	return ast.IsStringLiteralLike(expr)
}

func labelSeparatorRange(ctx rule.RuleContext, labelEnd int, bodyStart int) (core.TextRange, bool) {
	s := scanner.GetScannerForSourceFile(ctx.SourceFile, labelEnd)
	for s.Token() != ast.KindEndOfFile && s.TokenStart() < bodyStart {
		if s.Token() == ast.KindColonToken {
			return core.NewTextRange(s.TokenStart(), s.TokenEnd()), true
		}
		s.Scan()
	}
	return core.TextRange{}, false
}

func hasCommentBetweenLabelAndBody(ctx rule.RuleContext, node *ast.Node) bool {
	ls := node.AsLabeledStatement()
	labelToken := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, ls.Label.Pos())
	bodyStart := scanner.SkipTrivia(ctx.SourceFile.Text(), ls.Statement.Pos())
	colon, ok := labelSeparatorRange(ctx, labelToken.End(), bodyStart)
	if !ok {
		return utils.HasCommentInSpan(ctx.Comments, labelToken.End(), bodyStart)
	}
	return utils.HasCommentInSpan(ctx.Comments, labelToken.End(), colon.Pos()) ||
		utils.HasCommentInSpan(ctx.Comments, colon.End(), bodyStart)
}

func isFixable(ctx rule.RuleContext, node *ast.Node) bool {
	ls := node.AsLabeledStatement()
	if ls == nil || ls.Label == nil || ls.Statement == nil {
		return false
	}

	if hasCommentBetweenLabelAndBody(ctx, node) {
		return false
	}

	return !isPotentialDirectiveLabel(node)
}

// https://eslint.org/docs/latest/rules/no-unused-labels
var NoUnusedLabelsRule = rule.Rule{
	Name: "no-unused-labels",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var scopeInfo *labelScope

		enterLabeledScope := func(node *ast.Node) {
			ls := node.AsLabeledStatement()
			scopeInfo = &labelScope{
				label: ls.Label.Text(),
				used:  false,
				upper: scopeInfo,
			}
		}

		exitLabeledScope := func(node *ast.Node) {
			if scopeInfo == nil {
				return
			}

			current := scopeInfo
			if !current.used {
				msg := buildUnusedMessage(current.label)
				if isFixable(ctx, node) {
					ls := node.AsLabeledStatement()
					start := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
					end := scanner.SkipTrivia(ctx.SourceFile.Text(), ls.Statement.Pos())
					ctx.ReportNodeWithFixes(
						ls.Label,
						msg,
						rule.RuleFixRemoveRange(core.NewTextRange(start, end)),
					)
				} else {
					ctx.ReportNode(node.AsLabeledStatement().Label, msg)
				}
			}

			scopeInfo = current.upper
		}

		markAsUsed := func(labelNode *ast.Node) {
			if labelNode == nil {
				return
			}

			label := labelNode.Text()
			for info := scopeInfo; info != nil; info = info.upper {
				if info.label == label {
					info.used = true
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindLabeledStatement:                      enterLabeledScope,
			rule.ListenerOnExit(ast.KindLabeledStatement): exitLabeledScope,
			ast.KindBreakStatement: func(node *ast.Node) {
				markAsUsed(node.AsBreakStatement().Label)
			},
			ast.KindContinueStatement: func(node *ast.Node) {
				markAsUsed(node.AsContinueStatement().Label)
			},
		}
	},
}
