package no_lonely_if

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var NoLonelyIfRule = rule.Rule{
	Name:   "unicorn/no-lonely-if",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindIfStatement: func(node *ast.Node) {
				if node.AsIfStatement().ElseStatement != nil {
					return
				}
				body := node
				if parent := node.Parent; parent != nil && parent.Kind == ast.KindBlock {
					if len(parent.AsBlock().Statements.Nodes) != 1 {
						return
					}
					body = parent
				}
				outer := body.Parent
				if outer == nil || outer.Kind != ast.KindIfStatement ||
					outer.AsIfStatement().ElseStatement != nil || outer.AsIfStatement().ThenStatement != body {
					return
				}
				ctx.ReportNodeWithDeferredFixes(node, rule.RuleMessage{
					Id:          "no-lonely-if",
					Description: "Unexpected `if` as the only statement in a `if` block without `else`.",
				}, func() []rule.RuleFix {
					return mergeIfStatements(ctx.SourceFile, outer, node)
				})
			},
		}
	},
}

type ifText struct {
	prefix, test string
	conditionEnd int
	body         core.TextRange
}

func getIfText(sourceFile *ast.SourceFile, node *ast.Node) (ifText, bool) {
	text := sourceFile.Text()
	start := utils.TrimNodeTextRange(sourceFile, node).Pos()
	body := utils.TrimNodeTextRange(sourceFile, node.AsIfStatement().ThenStatement)
	open, hasOpen := utils.TokenAtOrAfter(sourceFile, start+2)
	expression := node.AsIfStatement().Expression
	closing := scanner.SkipTrivia(text, expression.End())
	if !hasOpen || open.Kind != ast.KindOpenParenToken || closing >= len(text) || text[closing] != ')' ||
		open.End > closing || scanner.SkipTrivia(text, closing+1) != body.Pos() {
		return ifText{}, false
	}
	// Avoid offering edits for parser-recovered, unterminated blocks.
	if node.AsIfStatement().ThenStatement.Kind == ast.KindBlock &&
		(body.End() <= body.Pos()+1 || text[body.End()-1] != '}') {
		return ifText{}, false
	}
	test := text[open.End:closing]
	if ast.GetExpressionPrecedence(expression) < ast.OperatorPrecedenceLogicalAND {
		test = "(" + test + ")"
	}
	return ifText{prefix: text[start+2 : open.Start], test: test, conditionEnd: closing + 1, body: body}, true
}

func mergeIfStatements(sourceFile *ast.SourceFile, outerNode, innerNode *ast.Node) []rule.RuleFix {
	outer, outerOK := getIfText(sourceFile, outerNode)
	inner, innerOK := getIfText(sourceFile, innerNode)
	if !outerOK || !innerOK {
		return nil
	}
	text := sourceFile.Text()
	innerRange := utils.TrimNodeTextRange(sourceFile, innerNode)
	var conditionGap, leading, trailing string
	if outerNode.AsIfStatement().ThenStatement.Kind == ast.KindBlock {
		if innerNode.End() > outer.body.End()-1 {
			return nil
		}
		conditionGap = preserveGap(text[outer.conditionEnd:outer.body.Pos()])
		leading = preserveGap(text[outer.body.Pos()+1 : innerRange.Pos()])
		trailing = preserveGap(text[innerNode.End() : outer.body.End()-1])
	} else {
		leading = preserveGap(text[outer.conditionEnd:innerRange.Pos()])
	}

	var conditionSuffix string
	if strings.HasPrefix(strings.TrimLeftFunc(conditionGap, ecmascript.IsWhiteSpaceOrLineTerminator), "//") {
		// Line-scoped directives must stay on the merged condition's line.
		conditionSuffix = conditionGap
	} else {
		leading = joinSourceText(conditionGap, leading)
	}
	leading = strings.TrimLeftFunc(leading, ecmascript.IsWhiteSpaceOrLineTerminator)

	replacementRange := utils.TrimNodeTextRange(sourceFile, outerNode)
	consequent := innerNode.AsIfStatement().ThenStatement
	consequentText := text[inner.conditionEnd:inner.body.Pos()]
	if consequent.Kind == ast.KindBlock {
		consequentText += text[inner.body.Pos():inner.body.End()-1] + trailing + "}"
	} else {
		consequentText += text[inner.body.Pos():inner.body.End()] + trailing
		next, hasNext := utils.TokenAtOrAfter(sourceFile, outerNode.End())
		if hasNext && inner.body.End() > inner.body.Pos() && text[inner.body.End()-1] != ';' {
			last, hasLast := utils.TokenBeforePosition(sourceFile, consequent.End())
			if hasLast && unicornutil.NeedsSemicolonAfter(sourceFile, last, next.Text, utils.IsSameLine(sourceFile, outerNode.End(), next.Start)) {
				replacementRange = replacementRange.WithEnd(next.Start)
				gap := text[outerNode.End():next.Start]
				consequentText += ";"
				if !ecmascript.IsBlank(gap) || strings.Contains(gap, "\n") {
					consequentText += gap
				}
			}
		}
	}
	// One contiguous edit keeps nested merges and other rules' fixes from
	// applying only part of the rewrite (upstream regression #2915).
	return []rule.RuleFix{rule.RuleFixReplaceRange(replacementRange,
		leading+"if"+outer.prefix+"("+outer.test+" && "+joinSourceText(preserveGap(inner.prefix), inner.test)+")"+conditionSuffix+consequentText)}
}

func preserveGap(text string) string {
	if ecmascript.IsBlank(text) {
		return ""
	}
	return text
}

func joinSourceText(left, right string) string {
	last, _ := utf8.DecodeLastRuneInString(left)
	first, _ := utf8.DecodeRuneInString(right)
	if ecmascript.IsWhiteSpaceOrLineTerminator(last) && ecmascript.IsWhiteSpaceOrLineTerminator(first) {
		right = strings.TrimLeftFunc(right, ecmascript.IsWhiteSpaceOrLineTerminator)
	}
	return left + right
}
