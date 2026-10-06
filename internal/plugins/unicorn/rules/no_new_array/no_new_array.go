package no_new_array

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoNewArrayRule = rule.Rule{
	Name:   "unicorn/no-new-array",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				construction := node.AsNewExpression()
				callee := utils.ESTreeRuntimeExpression(construction.Expression)
				if callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "Array" || len(node.Arguments()) != 1 {
					return
				}

				argument := utils.ESTreeRuntimeExpression(node.Arguments()[0])
				message := rule.RuleMessage{
					Id:          "error",
					Description: "`new Array()` is unclear in intent; use an array literal or `Array.from()`.",
				}
				nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				buildFix := func() []rule.RuleFix {
					// Unlike upstream, preserve the explicitly declared element type.
					if construction.TypeArguments != nil && len(construction.TypeArguments.Nodes) > 0 {
						return nil
					}
					// Replacing the whole constructor must not discard comments.
					if utils.HasCommentInSpan(ctx.Comments.All(), nodeRange.Pos(), nodeRange.End()) {
						return nil
					}
					if argument.Kind != ast.KindSpreadElement {
						if evaluator == nil {
							evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
							evaluator.GlobalAccess = ctx.Globals.Access
						}
						value, known := evaluator.EvalControlFlowValue(argument)
						_, number := value.(jsnum.Number)
						// A number creates holes; an unknown value may be a number.
						// Built-in objects, functions, and symbols unsupported by the
						// shared evaluator remain report-only, including their typeof.
						if !known || number {
							return nil
						}
					}

					// A same-named custom constructor can have unrelated behavior.
					if !ctx.Globals.Access("Array").IsDeclared() || !unicornutil.IsGlobalReference(ctx, callee) {
						return nil
					}
					text := utils.TrimmedNodeText(ctx.SourceFile, argument)
					if ast.IsParenthesizedExpression(node.Arguments()[0]) {
						text = "(" + text + ")"
					}
					text = "[" + text + "]"
					if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, text) {
						text = ";" + text
					}
					return []rule.RuleFix{rule.RuleFixReplaceRange(nodeRange, text)}
				}

				if argument.Kind == ast.KindSpreadElement {
					ctx.ReportRangeWithDeferredSuggestions(nodeRange, message, func() []rule.RuleSuggestion {
						fixes := buildFix()
						if len(fixes) == 0 {
							return nil
						}
						return []rule.RuleSuggestion{{
							Message:  rule.RuleMessage{Id: "spread", Description: "Spread the argument."},
							FixesArr: fixes,
						}}
					})
					return
				}
				ctx.ReportRangeWithDeferredFixes(nodeRange, message, buildFix)
			},
		}
	},
}
