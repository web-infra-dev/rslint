package prefer_at

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed prefer_at.schema.json
var schema []byte

var (
	negativeIndexMessage  = rule.RuleMessage{Id: "negative-index", Description: "Prefer `.at(…)` over `[….length - index]`."}
	indexMessage          = rule.RuleMessage{Id: "index", Description: "Prefer `.at(…)` over index access."}
	negativeCharAtMessage = rule.RuleMessage{
		Id: "string-char-at-negative", Description: "Prefer `String#at(…)` over `String#charAt(….length - index)`.",
	}
	charAtMessage    = rule.RuleMessage{Id: "string-char-at", Description: "Prefer `String#at(…)` over `String#charAt(…)`."}
	substringMessage = rule.RuleMessage{
		Id: "string-substring", Description: "Prefer `String#at(…)` over `String#substring(…)` when getting one character.",
	}
	sliceMessage = rule.RuleMessage{Id: "slice", Description: "Prefer `.at(…)` over the first element from `.slice(…)`."}
	useAt        = rule.RuleMessage{Id: "use-at", Description: "Use `.at(…)`."}
)

// PreferAtRule follows eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-at.js
var PreferAtRule = rule.Rule{
	Name:   "unicorn/prefer-at",
	Schema: rule.NewSchema(schema),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var checkAllIndexAccess bool
		var lastFunctions []string
		if len(options) != 0 {
			config, _ := options[0].(map[string]any)
			checkAllIndexAccess, _ = config["checkAllIndexAccess"].(bool)
			if names, ok := config["getLastElementFunctions"].([]any); ok {
				for _, name := range names {
					if path, ok := name.(string); ok {
						lastFunctions = append(lastFunctions, ecmascript.StringTrim(path))
					}
				}
			}
		}
		lastFunctions = append(lastFunctions, "_.last", "lodash.last", "underscore.last")

		var evaluator *utils.StaticStringEvaluator
		staticEvaluator := func() *utils.StaticStringEvaluator {
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
				evaluator.GlobalAccess = ctx.Globals.Access
			}
			return evaluator
		}
		return rule.RuleListeners{
			ast.KindElementAccessExpression: func(node *ast.Node) {
				access := node.AsElementAccessExpression()
				index := utils.ESTreeRuntimeExpression(access.ArgumentExpression)
				if index == nil || ast.NodeIsMissing(index) || isReferenceContext(node) {
					return
				}
				length := negativeIndexLength(index, access.Expression)
				message := negativeIndexMessage
				if length == nil {
					if !checkAllIndexAccess {
						return
					}
					value, _ := staticEvaluator().EvalValueIfNoSideEffects(index)
					number, ok := value.(jsnum.Number)
					if !ok || !unicornutil.IsNonNegativeSafeInteger(float64(number)) ||
						isObviouslyUnsupportedReceiver(ctx, access.Expression) {
						return
					}
					message = indexMessage
				}
				if isArguments(access.Expression) || isDOMCollection(ctx, access.Expression) {
					return
				}
				ctx.ReportNodeWithDeferredFixes(index, message, func() []rule.RuleFix {
					return indexFixes(ctx, node, length)
				})
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call, matched := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods: []string{"charAt", "substring", "slice"}, AllowOptionalCall: true, AllowOptionalMember: true,
				})
				args := node.Arguments()
				if matched && len(args) >= 1 && len(args) <= 2 && args[0].Kind != ast.KindSpreadElement &&
					(len(args) == 1 || args[1].Kind != ast.KindSpreadElement) {
					switch call.Property.Text() {
					case "charAt":
						if len(args) == 1 && !ast.IsOptionalChainRoot(node) {
							index := utils.ESTreeRuntimeExpression(args[0])
							if index == nil || ast.NodeIsMissing(index) {
								break
							}
							length := negativeIndexLength(index, call.Object)
							if length != nil || (checkAllIndexAccess && !isStaticallyNegative(staticEvaluator(), index)) {
								reportCharAt(ctx, call, index, length)
							}
						}
					case "substring":
						if index := unicornutil.SingleCharacterSubstringIndex(args); index != nil {
							reportSubstring(ctx, call, index)
						}
					case "slice":
						if !ast.IsOptionalChainRoot(node) {
							reportSlice(ctx, call)
						}
					}
				}
				reportLastFunction(ctx, node, lastFunctions)
			},
		}
	},
}

func isStaticallyNegative(evaluator *utils.StaticStringEvaluator, index *ast.Node) bool {
	value, known := evaluator.EvalValueIfNoSideEffects(index)
	if !known {
		return false
	}
	if number, ok := value.(jsnum.Number); ok {
		return number < 0
	}
	// String, bigint, and ordinary aggregate values can be compared with zero
	// through their numeric string. Other primitives cannot be negative.
	text, ok := evaluator.EvalToString(index)
	if !ok {
		return false
	}
	number, _ := ecmascript.StringToNumber(text)
	return number < 0
}

func reportCharAt(ctx rule.RuleContext, call unicornutil.DotMethodCall, index, length *ast.Node) {
	message := charAtMessage
	if length != nil {
		message = negativeCharAtMessage
	}
	ctx.ReportNodeWithDeferredSuggestions(index, message, func() []rule.RuleSuggestion {
		fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call.Property, "at")}
		if length != nil {
			fix, ok := removeLength(ctx, length)
			if !ok {
				return nil
			}
			fixes = append(fixes, fix)
		}
		return []rule.RuleSuggestion{{Message: useAt, FixesArr: fixes}}
	})
}

func reportSubstring(ctx rule.RuleContext, call unicornutil.DotMethodCall, index *ast.Node) {
	ctx.ReportNodeWithDeferredSuggestions(call.Call, substringMessage, func() []rule.RuleSuggestion {
		args := call.Call.Arguments()
		start := utils.TrimNodeTextRange(ctx.SourceFile, args[0]).Pos()
		end := args[1].End()
		if utils.HasCommentInSpan(ctx.Comments.All(), start, end) {
			return nil
		}
		return []rule.RuleSuggestion{{Message: useAt, FixesArr: []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, call.Property, "at"),
			rule.RuleFixReplaceRange(core.NewTextRange(start, end), utils.TrimmedNodeText(ctx.SourceFile, index)),
		}}}
	})
}

func reportSlice(ctx rule.RuleContext, call unicornutil.DotMethodCall) {
	args := call.Call.Arguments()
	start, ok := negativeIntegerLiteral(args[0])
	if !ok {
		return
	}
	if len(args) == 1 {
		if start != -1 {
			return
		}
	} else if end, ok := negativeIntegerLiteral(args[1]); !ok || end != start+1 {
		return
	}

	member := utils.ESTreeParent(call.Call)
	if member == nil || !ast.IsAccessExpression(member) || utils.ESTreeRuntimeExpression(member.Expression()) != call.Call ||
		ast.IsOptionalChainRoot(member) || (ast.IsOptionalChain(call.Call) && !ast.IsOptionalChain(member)) {
		return
	}
	var result *ast.Node
	var method unicornutil.DotMethodCall
	if ast.IsElementAccessExpression(member) {
		index, ok := unicornutil.SafeIntegerLiteral(member.AsElementAccessExpression().ArgumentExpression)
		if !ok || index != 0 || isReferenceContext(member) {
			return
		}
		result = member
	} else {
		zero := 0
		method, ok = unicornutil.MatchDotMethodCall(utils.ESTreeParent(member), unicornutil.DotMethodCallOptions{
			Methods: []string{"pop", "shift"}, ArgumentsLength: &zero,
		})
		if !ok || method.Callee != member {
			return
		}
		result = method.Call
	}
	if isArguments(call.Object) || isDOMCollection(ctx, call.Object) {
		return
	}
	ctx.ReportNodeWithDeferredFixes(call.Property, sliceMessage, func() []rule.RuleFix {
		return sliceFixes(ctx, call, result, method)
	})
}

func reportLastFunction(ctx rule.RuleContext, node *ast.Node, paths []string) {
	args := node.Arguments()
	if len(args) != 1 || args[0].Kind == ast.KindSpreadElement || ast.IsOptionalChainRoot(node) {
		return
	}
	callee := utils.ESTreeCallCallee(node.Expression())
	for _, path := range paths {
		if !unicornutil.NodeMatchesPath(callee, path) {
			continue
		}
		receiver := args[0]
		if isArguments(receiver) || isDOMCollection(ctx, receiver) {
			return
		}
		message := rule.RuleMessage{
			Id:          "get-last-function",
			Description: "Prefer `.at(-1)` over `" + path + "(…)` to get the last element.",
			Data:        map[string]string{"description": path},
		}
		ctx.ReportNodeWithDeferredFixes(callee, message, func() []rule.RuleFix {
			return lastFunctionFixes(ctx, node, receiver)
		})
		return
	}
}
