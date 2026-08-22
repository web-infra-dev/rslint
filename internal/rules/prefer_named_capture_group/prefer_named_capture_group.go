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
		var eval *utils.StaticStringEvaluator
		getEval := func() *utils.StaticStringEvaluator {
			if eval == nil {
				eval = utils.NewStaticStringEvaluatorWithoutScope()
			}
			return eval
		}

		checkConstructor := func(node *ast.Node, callee *ast.Node, args *ast.NodeList) {
			if !isGlobalRegExpCallee(ctx, utils.SkipAssertionsAndParens(callee)) {
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

			checkRegex(ctx, node, patternNode, pattern, flags)
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

// isGlobalRegExpCallee reports whether callee is a reference to the global
// `RegExp` constructor: the bare identifier, or a property/element access on
// one of the global-object aliases ESLint's ReferenceTracker recognizes
// (`globalThis`, `window`, `self`, `global`). Each name must resolve to the
// real global — not a local declaration or an `off`-disabled global.
func isGlobalRegExpCallee(ctx rule.RuleContext, callee *ast.Node) bool {
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		if callee.AsIdentifier().Text != "RegExp" || utils.IsShadowed(callee, "RegExp") {
			return false
		}
		return ctx.Globals.Access("RegExp").IsDeclared()
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return false
		}
		if access.Name().AsIdentifier().Text != "RegExp" {
			return false
		}
		return isKnownGlobalObject(ctx, access.Expression)
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return false
		}
		value, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(access.ArgumentExpression))
		if !ok || value != "RegExp" {
			return false
		}
		return isKnownGlobalObject(ctx, access.Expression)
	}

	return false
}

func isKnownGlobalObject(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	name := node.AsIdentifier().Text
	switch name {
	case "globalThis", "window", "self", "global":
		return !utils.IsShadowed(node, name) && ctx.Globals.Access(name).IsDeclared()
	default:
		return false
	}
}
