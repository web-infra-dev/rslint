// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
package prefer_single_call

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed prefer_single_call.schema.json
var schemaJSON []byte

var PreferSingleCallRule = rule.Rule{
	Name:   "unicorn/prefer-single-call",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var ignored []string
		if len(options) > 0 {
			object, _ := options[0].(map[string]any)
			paths, _ := object["ignore"].([]any)
			for _, path := range paths {
				if path, ok := path.(string); ok {
					ignored = append(ignored, path)
				}
			}
		}
		var evaluator *utils.StaticStringEvaluator

		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			second, ok := matchCall(node)
			if !ok || isIgnored(second, ignored) {
				return
			}
			secondStatement := utils.ESTreeParent(node)
			if secondStatement == nil || !ast.IsExpressionStatement(secondStatement) {
				return
			}
			firstStatement := previousStatement(secondStatement)
			if firstStatement == nil || !ast.IsExpressionStatement(firstStatement) {
				return
			}
			first, ok := matchCall(utils.ESTreeRuntimeExpression(firstStatement.Expression()))
			if !ok || first.method != second.method || !unicornutil.IsSameReference(first.callee, second.callee) ||
				(second.isArrayMethod() && unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, second.receiver)) {
				return
			}

			description := second.description()
			message := rule.RuleMessage{
				Id: "error/array-push", Description: "Do not call `" + description + "` multiple times.",
				Data: map[string]string{"description": description},
			}
			keepSecond := second.method == "unshift" &&
				(!unicornutil.HasOptionalChainElement(second.callee) || unicornutil.HasOptionalChainElement(first.callee))
			firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstStatement)
			secondRange := utils.TrimNodeTextRange(ctx.SourceFile, secondStatement)
			removal := core.NewTextRange(firstRange.End(), secondRange.End())
			if keepSecond {
				removal = core.NewTextRange(firstRange.Pos(), secondRange.Pos())
			}
			var checked, blocked, suggest bool
			checkEdits := func() {
				if checked {
					return
				}
				checked = true
				blocked = utils.HasCommentInSpan(ctx.Comments.All(), removal.Pos(), removal.End())
				if blocked {
					return
				}
				if evaluator == nil {
					evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					evaluator.GlobalAccess = ctx.Globals.Access
				}
				suggest = needsSuggestion(ctx, evaluator, first, second)
			}

			buildFixes := func() []rule.RuleFix {
				return mergeFixes(ctx, first.node, second.node, firstRange, secondRange, removal, keepSecond, second.method == "unshift" && !keepSecond)
			}
			ctx.ReportNodeWithDeferredFixesAndSuggestions(second.report, message,
				func() []rule.RuleFix {
					checkEdits()
					if blocked || suggest {
						return nil
					}
					return buildFixes()
				},
				func() []rule.RuleSuggestion {
					checkEdits()
					if blocked || !suggest {
						return nil
					}
					return []rule.RuleSuggestion{{
						Message:  rule.RuleMessage{Id: "suggestion", Description: "Merge with previous one.", Data: message.Data},
						FixesArr: buildFixes(),
					}}
				})
		}}
	},
}

type mergeCall struct {
	node, callee, receiver, report *ast.Node
	method                         string
}

func (call mergeCall) isArrayMethod() bool {
	return call.method == "push" || call.method == "unshift"
}

func (call mergeCall) description() string {
	if call.isArrayMethod() {
		return "Array#" + call.method + "()"
	}
	if call.method == "importScripts" {
		return "importScripts()"
	}
	return "Element#classList." + call.method + "()"
}

func matchCall(node *ast.Node) (mergeCall, bool) {
	if node == nil || !ast.IsCallExpression(node) {
		return mergeCall{}, false
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	if callee != nil && ast.IsIdentifier(callee) && callee.Text() == "importScripts" {
		return mergeCall{node: node, callee: callee, report: callee, method: "importScripts"}, true
	}
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods: []string{"push", "unshift", "add", "remove"}, AllowOptionalMember: true,
	})
	if !ok {
		return mergeCall{}, false
	}
	method := call.Property.Text()
	if method == "add" || method == "remove" {
		object := utils.ESTreeRuntimeExpression(call.Object)
		if ast.IsOptionalChainRoot(call.Callee) || !ast.IsPropertyAccessExpression(object) ||
			(call.Object != object && ast.IsOptionalChain(object)) ||
			!ast.IsIdentifier(object.Name()) || object.Name().Text() != "classList" {
			return mergeCall{}, false
		}
	}
	return mergeCall{node: node, callee: call.Callee, receiver: call.Object, report: call.Property, method: method}, true
}

func isIgnored(call mergeCall, ignored []string) bool {
	if call.isArrayMethod() && !ast.IsOptionalChainRoot(call.callee) {
		for _, receiver := range []string{"stream", "this", "this.stream", "process.stdin", "process.stdout", "process.stderr"} {
			if unicornutil.NodeMatchesPath(call.receiver, receiver) {
				return true
			}
		}
	}
	for _, path := range ignored {
		if unicornutil.NodeMatchesPath(call.callee, path) {
			return true
		}
	}
	return false
}

func previousStatement(statement *ast.Node) *ast.Node {
	parent := statement.Parent
	if parent == nil || !parent.CanHaveStatements() {
		return nil
	}
	statements := parent.Statements()
	if index := ast.IndexOfNode(statements, statement); index > 0 {
		return statements[index-1]
	}
	return nil
}

func needsSuggestion(ctx rule.RuleContext, evaluator *utils.StaticStringEvaluator, first, second mergeCall) bool {
	if second.isArrayMethod() {
		if !unicornutil.IsArray(ctx, second.receiver) {
			return true
		}
		// Both sets of arguments move before the first mutation. A read such as
		// array.length, a spread, or an unknown value must remain a suggestion.
		for _, call := range []*ast.Node{first.node, second.node} {
			for _, argument := range call.Arguments() {
				if _, known := evaluator.EvalValueIfNoSideEffects(argument); !known {
					return true
				}
			}
		}
		return false
	}
	for _, argument := range second.node.Arguments() {
		if evaluator.HasSideEffect(argument, false) {
			return true
		}
	}
	return false
}

func mergeFixes(ctx rule.RuleContext, first, second *ast.Node, firstRange, secondRange, removal core.TextRange, keepSecond, prepend bool) []rule.RuleFix {
	target, source := first, second
	if keepSecond {
		target, source = second, first
	}
	text := ctx.SourceFile.Text()
	var fixes []rule.RuleFix
	if len(source.Arguments()) > 0 {
		start, end := source.AsCallExpression().Arguments.Pos(), source.End()-1
		if prepend && source.AsCallExpression().Arguments.HasTrailingComma() {
			comma, _ := utils.TokenAtOrAfter(ctx.SourceFile, source.Arguments()[len(source.Arguments())-1].End())
			end = comma.Start
		}
		arguments := text[start:end]
		position := target.End() - 1
		if prepend {
			position = target.AsCallExpression().Arguments.Pos()
			if len(target.Arguments()) > 0 {
				arguments += ", "
			}
		} else if target.AsCallExpression().Arguments.HasTrailingComma() {
			position = target.AsCallExpression().Arguments.End()
			arguments = " " + arguments
		} else if len(target.Arguments()) > 0 {
			arguments = ", " + arguments
		}
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(position, position), arguments))
	}
	replacement := ""
	if !keepSecond && text[firstRange.End()-1] != ';' && text[secondRange.End()-1] == ';' {
		replacement = ";"
	}
	if keepSecond {
		previous, ok := utils.TokenBeforePosition(ctx.SourceFile, firstRange.Pos())
		if ok && unicornutil.NeedsSemicolonAfter(ctx.SourceFile, previous, text[secondRange.Pos():secondRange.Pos()+1], false) {
			replacement = ";"
		}
	}
	return append(fixes, rule.RuleFixReplaceRange(removal, replacement))
}
