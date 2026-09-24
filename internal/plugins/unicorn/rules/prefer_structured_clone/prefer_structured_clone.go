// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package prefer_structured_clone

import (
	_ "embed"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed prefer_structured_clone.schema.json
var schemaJSON []byte

const (
	messageIDError      = "prefer-structured-clone/error"
	messageIDSuggestion = "prefer-structured-clone/suggestion"
)

var suggestionMessage = rule.RuleMessage{
	Id:          messageIDSuggestion,
	Description: "Switch to `structuredClone(…)`.",
}

func errorMessage(description string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDError,
		Description: "Prefer `structuredClone(…)` over `" + description + "` to create a deep clone.",
	}
}

var PreferStructuredCloneRule = rule.Rule{
	Name:   "unicorn/prefer-structured-clone",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		functions := configuredFunctions(rawOptions)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if outer, inner, ok := matchJSONClone(node); ok {
					reportJSONClone(ctx, node, outer, inner)
					return
				}
				reportConfiguredClone(ctx, node, functions)
			},
		}
	},
}

func configuredFunctions(rawOptions []any) []string {
	functions := []string{"_.cloneDeep", "lodash.cloneDeep"}
	if len(rawOptions) == 0 {
		return functions
	}
	options, _ := rawOptions[0].(map[string]any)
	return append(utils.ToStringSlice(options["functions"]), functions...)
}

func matchJSONClone(node *ast.Node) (unicornutil.DotMethodCall, unicornutil.DotMethodCall, bool) {
	oneArgument := 1
	outer, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method:              "parse",
		ArgumentsLength:     &oneArgument,
		RejectSpreadElement: true,
	})
	if !ok || !unicornutil.NodeMatchesPath(outer.Object, "JSON") {
		return unicornutil.DotMethodCall{}, unicornutil.DotMethodCall{}, false
	}

	// The outer matcher already guarantees exactly one non-spread argument.
	innerNode := utils.ESTreeRuntimeExpression(node.Arguments()[0])
	inner, ok := unicornutil.MatchDotMethodCall(innerNode, unicornutil.DotMethodCallOptions{
		Method:              "stringify",
		ArgumentsLength:     &oneArgument,
		RejectSpreadElement: true,
	})
	if !ok || !unicornutil.NodeMatchesPath(inner.Object, "JSON") {
		return unicornutil.DotMethodCall{}, unicornutil.DotMethodCall{}, false
	}

	return outer, inner, true
}

func reportJSONClone(ctx rule.RuleContext, outerNode *ast.Node, outer, inner unicornutil.DotMethodCall) {
	start := utils.TrimNodeTextRange(ctx.SourceFile, outerNode).Pos()
	end := utils.TrimNodeTextRange(ctx.SourceFile, inner.Callee).End()
	ctx.ReportRangeWithDeferredSuggestions(
		core.NewTextRange(start, end),
		errorMessage("JSON.parse(JSON.stringify(…))"),
		func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message:  suggestionMessage,
				FixesArr: jsonCloneSuggestionFixes(ctx.SourceFile, outer, inner),
			}}
		},
	)
}

func jsonCloneSuggestionFixes(
	sourceFile *ast.SourceFile,
	outer unicornutil.DotMethodCall,
	inner unicornutil.DotMethodCall,
) []rule.RuleFix {
	opening, closing, trailingComma, hasTrailingComma := callSyntax(sourceFile, inner.Call)

	fixes := []rule.RuleFix{
		rule.RuleFixReplace(sourceFile, outer.Callee, "structuredClone"),
	}
	fixes = append(fixes, removeTransparentCalleeFixes(sourceFile, inner.RawCallee, inner.Callee)...)
	if typeArguments, ok := callTypeArgumentsRange(sourceFile, inner.Call); ok {
		fixes = append(fixes, rule.RuleFixRemoveRange(typeArguments))
	}
	fixes = append(fixes, rule.RuleFixRemoveRange(opening))
	if hasTrailingComma {
		fixes = append(fixes, rule.RuleFixRemoveRange(trailingComma))
	}
	fixes = append(fixes, rule.RuleFixRemoveRange(closing))
	return fixes
}

func removeTransparentCalleeFixes(sourceFile *ast.SourceFile, rawCallee, callee *ast.Node) []rule.RuleFix {
	fixes := []rule.RuleFix{rule.RuleFixRemoveRange(utils.TrimNodeTextRange(sourceFile, callee))}
	for current := rawCallee; current != nil && current != callee; {
		switch current.Kind {
		case ast.KindParenthesizedExpression:
			tokens := utils.TokensOfNode(sourceFile, current)
			if len(tokens) >= 2 && tokens[0].Kind == ast.KindOpenParenToken &&
				tokens[len(tokens)-1].Kind == ast.KindCloseParenToken {
				fixes = append(fixes,
					rule.RuleFixRemoveRange(tokens[0].Range()),
					rule.RuleFixRemoveRange(tokens[len(tokens)-1].Range()),
				)
			}
			current = current.AsParenthesizedExpression().Expression
		case ast.KindAsExpression, ast.KindSatisfiesExpression:
			// MatchDotMethodCall only makes these wrappers transparent when they
			// are parser-synthesized JSDoc casts, so the inner runtime expression
			// is guaranteed to exist here.
			current = utils.JSDocTypeCastExpression(current)
		}
	}
	return fixes
}

func callTypeArgumentsRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, bool) {
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil || call.TypeArguments == nil ||
		len(call.TypeArguments.Nodes) == 0 {
		return core.TextRange{}, false
	}
	// matchJSONClone already rejected optional calls and this is a parsed
	// CallExpression with an explicit type-argument list. The scanner ranges
	// therefore point at the authored angle brackets.
	openAngle := scanner.GetRangeOfTokenAtPosition(sourceFile, call.Expression.End())
	closeAngle := scanner.GetRangeOfTokenAtPosition(sourceFile, call.TypeArguments.End())
	return core.NewTextRange(openAngle.Pos(), closeAngle.End()), true
}

// callSyntax relies only on invariants established by MatchDotMethodCall: node is
// a parsed CallExpression with a complete argument list. Walking backward from
// the call's final ')' finds its matching '(' without caring about parentheses
// inside the callee or argument.
func callSyntax(sourceFile *ast.SourceFile, node *ast.Node) (
	opening core.TextRange,
	closing core.TextRange,
	trailingComma core.TextRange,
	hasTrailingComma bool,
) {
	tokens := utils.TokensOfNode(sourceFile, node)
	closingIndex := len(tokens) - 1
	closing = tokens[closingIndex].Range()

	depth := 0
	for index := closingIndex; index >= 0; index-- {
		switch tokens[index].Kind {
		case ast.KindCloseParenToken:
			depth++
		case ast.KindOpenParenToken:
			depth--
			if depth == 0 {
				opening = tokens[index].Range()
				index = -1
			}
		}
	}

	if closingIndex > 0 && tokens[closingIndex-1].Kind == ast.KindCommaToken {
		trailingComma = tokens[closingIndex-1].Range()
		hasTrailingComma = true
	}
	return opening, closing, trailingComma, hasTrailingComma
}

func reportConfiguredClone(ctx rule.RuleContext, node *ast.Node, functions []string) {
	call := node.AsCallExpression()
	if call == nil || ast.IsOptionalChainRoot(node) {
		return
	}
	arguments := node.Arguments()
	if len(arguments) != 1 || arguments[0] == nil || arguments[0].Kind == ast.KindSpreadElement {
		return
	}

	rawCallee := call.Expression
	callee := utils.ESTreeRuntimeExpression(rawCallee)
	if callee == nil || unicornutil.HasOptionalChainElement(rawCallee) {
		return
	}

	matched := ""
	for _, function := range functions {
		if unicornutil.NodeMatchesPath(rawCallee, function) {
			matched = ecmascript.StringTrim(function)
			break
		}
	}
	if matched == "" {
		return
	}

	ctx.ReportRangeWithDeferredSuggestions(
		utils.TrimNodeTextRange(ctx.SourceFile, callee),
		errorMessage(matched+"(…)"),
		func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message: suggestionMessage,
				FixesArr: []rule.RuleFix{
					rule.RuleFixReplace(ctx.SourceFile, callee, "structuredClone"),
				},
			}}
		},
	)
}
