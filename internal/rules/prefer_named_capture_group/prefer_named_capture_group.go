package prefer_named_capture_group

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/prefer-named-capture-group
var PreferNamedCaptureGroupRule = rule.Rule{
	Name:   "prefer-named-capture-group",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		modifiedGlobals := collectModifiedGlobalRoots(ctx)
		var eval *utils.StaticStringEvaluator
		getEval := func() *utils.StaticStringEvaluator {
			if eval == nil {
				eval = utils.NewStaticStringEvaluatorWithoutScope()
			}
			return eval
		}

		checkConstructor := func(node *ast.Node, callee *ast.Node, args *ast.NodeList) {
			calleeReferences := globalRegExpCalleeReferences(ctx, utils.SkipAssertionsAndParens(callee), getEval(), modifiedGlobals)
			if calleeReferences == 0 {
				return
			}
			if args == nil || len(args.Nodes) == 0 {
				return
			}
			patternNode := utils.SkipAssertionsAndParens(args.Nodes[0])
			pattern, ok := staticStringValue(getEval(), patternNode)
			if !ok || pattern == "" {
				return
			}

			flags := ""
			if len(args.Nodes) >= 2 {
				flagsNode := utils.SkipAssertionsAndParens(args.Nodes[1])
				if v, ok := staticStringValue(getEval(), flagsNode); ok {
					flags = v
				}
			}

			for range calleeReferences {
				checkRegex(ctx, node, patternNode, pattern, flags)
			}
		}

		return rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
				checkRegex(ctx, node, node, pattern, flags)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				checkConstructor(node, call.Expression, call.Arguments)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				checkConstructor(node, newExpr.Expression, newExpr.Arguments)
			},
		}
	},
}

// staticStringValue folds a `RegExp()` argument to a string the way
// `String(value)` would: a regex literal becomes its own source text —
// delimiting slashes and flags included, so `new RegExp(/(a)/)` is read as the
// pattern `/(a)/` — and everything else goes through the static evaluator.
func staticStringValue(eval *utils.StaticStringEvaluator, node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindRegularExpressionLiteral {
		return node.Text(), true
	}
	return eval.EvalToString(node)
}

// checkRegex parses pattern (as it would be read under flags) and reports
// every capturing group that isn't named. reportNode anchors the diagnostic
// range (the literal, or the whole call/new expression for a constructor
// call); patternSourceNode is the AST node the pattern text came from, used
// to decide whether — and where — a fix-safe suggestion can be offered.
func checkRegex(ctx rule.RuleContext, reportNode *ast.Node, patternSourceNode *ast.Node, pattern string, flags string) {
	rxFlags := utils.ParseRegexFlags(flags)
	groups, ok := utils.RegexCapturingGroups(pattern, rxFlags)
	if !ok {
		return
	}

	for _, group := range groups {
		if group.Name != "" {
			continue
		}
		raw := pattern[group.Start:group.End]
		groupStart := group.Start

		msg := rule.RuleMessage{
			Id:          "required",
			Description: fmt.Sprintf("Capture group '%s' should be converted to a named or non-capturing group.", raw),
			Data:        map[string]string{"group": raw},
		}

		ctx.ReportNodeWithDeferredSuggestions(reportNode, msg, func() []rule.RuleSuggestion {
			return buildSuggestions(ctx, patternSourceNode, pattern, groupStart)
		})
	}
}

var existingTempNameRe = regexp.MustCompile(`temp(\d+)`)

// buildSuggestions offers two fixer suggestions — name the group, or make it
// non-capturing — but only when pattern maps 1:1 onto patternSourceNode's raw
// source text, so the computed insertion offset is safe to use directly.
// String concatenation and substituted template literals never satisfy that,
// and get no suggestions (matching ESLint's suggestIfPossible).
func buildSuggestions(ctx rule.RuleContext, patternSourceNode *ast.Node, pattern string, groupStart int) []rule.RuleSuggestion {
	if patternSourceNode == nil {
		return nil
	}

	rawText := utils.TrimmedNodeText(ctx.SourceFile, patternSourceNode)
	switch patternSourceNode.Kind {
	case ast.KindRegularExpressionLiteral:
		// A regex literal carries its pattern verbatim in its own source.
	case ast.KindStringLiteral:
		if strings.Contains(rawText, "\\") {
			return nil
		}
	case ast.KindNoSubstitutionTemplateLiteral:
		if len(rawText) < 2 || rawText[1:len(rawText)-1] != pattern {
			return nil
		}
	default:
		return nil
	}

	nodeStart := utils.TrimNodeTextRange(ctx.SourceFile, patternSourceNode).Pos()
	insertAt := nodeStart + groupStart + 2

	nextTemp := 1
	for _, match := range existingTempNameRe.FindAllStringSubmatch(pattern, -1) {
		if n, err := strconv.Atoi(match[1]); err == nil && n+1 > nextTemp {
			nextTemp = n + 1
		}
	}

	insertRange := core.NewTextRange(insertAt, insertAt)
	return []rule.RuleSuggestion{
		{
			Message: rule.RuleMessage{Id: "addGroupName", Description: "Add name to capture group."},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(insertRange, fmt.Sprintf("?<temp%d>", nextTemp)),
			},
		},
		{
			Message: rule.RuleMessage{Id: "addNonCapture", Description: "Convert group to non-capturing."},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(insertRange, "?:"),
			},
		},
	}
}

// globalRegExpCalleeReferences returns how many global RegExp references flow
// to callee through the ReferenceTracker pass-through expressions. A logical
// or conditional expression can contain more than one tracked root, and ESLint
// reports once for each one. A write to a global root disables every use of
// that root, including source-only runs where RefStore cannot resolve lib.d.ts
// globals.
func globalRegExpCalleeReferences(ctx rule.RuleContext, callee *ast.Node, eval *utils.StaticStringEvaluator, modifiedGlobals map[string]bool) int {
	if callee == nil {
		return 0
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		if callee.AsIdentifier().Text == "RegExp" && isUnmodifiedGlobal(ctx, callee, "RegExp", modifiedGlobals) {
			return 1
		}
		return 0
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return 0
		}
		if access.Name().AsIdentifier().Text != "RegExp" {
			return 0
		}
		return knownGlobalObjectReferences(ctx, access.Expression, modifiedGlobals)
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return 0
		}
		value, ok := eval.EvalToString(utils.SkipAssertionsAndParens(access.ArgumentExpression))
		if !ok || value != "RegExp" {
			return 0
		}
		return knownGlobalObjectReferences(ctx, access.Expression, modifiedGlobals)
	case ast.KindBinaryExpression:
		binary := callee.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return 0
		}
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken:
			return globalRegExpCalleeReferences(ctx, utils.SkipAssertionsAndParens(binary.Right), eval, modifiedGlobals)
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			return globalRegExpCalleeReferences(ctx, utils.SkipAssertionsAndParens(binary.Left), eval, modifiedGlobals) +
				globalRegExpCalleeReferences(ctx, utils.SkipAssertionsAndParens(binary.Right), eval, modifiedGlobals)
		}
	case ast.KindConditionalExpression:
		conditional := callee.AsConditionalExpression()
		if conditional != nil {
			return globalRegExpCalleeReferences(ctx, utils.SkipAssertionsAndParens(conditional.WhenTrue), eval, modifiedGlobals) +
				globalRegExpCalleeReferences(ctx, utils.SkipAssertionsAndParens(conditional.WhenFalse), eval, modifiedGlobals)
		}
	}

	return 0
}

func knownGlobalObjectReferences(ctx rule.RuleContext, node *ast.Node, modifiedGlobals map[string]bool) int {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return 0
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			switch binary.OperatorToken.Kind {
			case ast.KindCommaToken:
				return knownGlobalObjectReferences(ctx, binary.Right, modifiedGlobals)
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
				return knownGlobalObjectReferences(ctx, binary.Left, modifiedGlobals) + knownGlobalObjectReferences(ctx, binary.Right, modifiedGlobals)
			}
		}
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		if conditional != nil {
			return knownGlobalObjectReferences(ctx, conditional.WhenTrue, modifiedGlobals) + knownGlobalObjectReferences(ctx, conditional.WhenFalse, modifiedGlobals)
		}
	}
	if node.Kind != ast.KindIdentifier {
		return 0
	}
	name := node.AsIdentifier().Text
	switch name {
	case "globalThis", "window", "self", "global":
		if isUnmodifiedGlobal(ctx, node, name, modifiedGlobals) {
			return 1
		}
	default:
		return 0
	}
	return 0
}

func isUnmodifiedGlobal(ctx rule.RuleContext, node *ast.Node, name string, modifiedGlobals map[string]bool) bool {
	return !utils.IsShadowed(node, name) && ctx.Globals.Access(name).IsDeclared() && !modifiedGlobals[name]
}

func collectModifiedGlobalRoots(ctx rule.RuleContext) map[string]bool {
	modified := make(map[string]bool)
	if ctx.SourceFile == nil {
		return modified
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			name := node.AsIdentifier().Text
			switch name {
			case "RegExp", "globalThis", "window", "self", "global":
				if !utils.IsNonReferenceIdentifier(node) && utils.IsWriteReference(node) && !utils.IsShadowed(node, name) {
					modified[name] = true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	ctx.SourceFile.AsNode().ForEachChild(visit)
	return modified
}
