package prefer_to_be

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_to_be"
)

func message(kind shared.Kind) rule.RuleMessage {
	switch kind {
	case shared.KindToBe:
		return rule.RuleMessage{Id: "useToBe", Description: "Use `toBe` instead"}
	case shared.KindNull:
		return rule.RuleMessage{Id: "useToBeNull", Description: "Use `toBeNull()` instead"}
	case shared.KindNaN:
		return rule.RuleMessage{Id: "useToBeNaN", Description: "Use `toBeNaN()` instead"}
	case shared.KindUndefined:
		return rule.RuleMessage{Id: "useToBeUndefined", Description: "Use `toBeUndefined()` instead"}
	case shared.KindDefined:
		return rule.RuleMessage{Id: "useToBeDefined", Description: "Use `toBeDefined()` instead"}
	default:
		return rule.RuleMessage{}
	}
}

func matcherName(kind shared.Kind) string {
	switch kind {
	case shared.KindToBe:
		return "toBe"
	case shared.KindNull:
		return "toBeNull"
	case shared.KindNaN:
		return "toBeNaN"
	case shared.KindUndefined:
		return "toBeUndefined"
	case shared.KindDefined:
		return "toBeDefined"
	default:
		return ""
	}
}

func buildFixes(ctx rule.RuleContext, match shared.Match) []rule.RuleFix {
	replacement := matcherName(match.Kind)
	if replacement == "" {
		return nil
	}

	nameRange, nameText, ok := testFramework.AccessorReplacement(
		ctx.SourceFile,
		match.Expect.MatcherEntry.Node,
		replacement,
	)
	if !ok {
		return nil
	}
	fixes := []rule.RuleFix{rule.RuleFixReplaceRange(nameRange, nameText)}

	if match.Kind != shared.KindToBe && len(match.Expect.MatcherCall.Arguments()) > 0 {
		argumentsRange, ok := testFramework.CallArgumentListRange(
			ctx.SourceFile,
			match.Expect.MatcherCall,
		)
		if !ok || internalUtils.HasCommentInSpan(
			ctx.Comments.All(),
			argumentsRange.Pos(),
			argumentsRange.End(),
		) {
			return nil
		}
		fixes = append(fixes, rule.RuleFixRemoveRange(argumentsRange))
	}

	if match.NotEntry != nil {
		ranges, ok := testFramework.RemoveAccessorEntryRanges(
			ctx.SourceFile,
			ctx.Comments.All(),
			match.NotEntry,
		)
		if !ok {
			return nil
		}
		for _, textRange := range ranges {
			fixes = append(fixes, rule.RuleFixRemoveRange(textRange))
		}
	}
	return fixes
}

func isUnshadowedSpecialIdentifier(ctx rule.RuleContext, node *ast.Node, _ string) bool {
	if ctx.Refs == nil {
		return true
	}
	symbol := ctx.Refs.Resolve(node)
	return symbol == nil || !internalUtils.IsRuntimeValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
}

var PreferToBeRule = shared.NewRule(shared.Config{
	Name:    "rstest/prefer-to-be",
	Message: message,
	// NOTE: Unlike @vitest/eslint-plugin, locally shadowed `undefined` and
	// `NaN` are values chosen by the author, not the built-ins named by the
	// dedicated matchers.
	IsSpecialIdentifier: isUnshadowedSpecialIdentifier,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
			parsed := analysis.ParseExpectCall(node)
			if parsed == nil ||
				parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				parsed.Head == nil ||
				parsed.Entry == rstestUtils.RstestExpectEntryElement ||
				parsed.MatcherEntry == nil ||
				len(parsed.Matchers) == 0 ||
				parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall ||
				testFramework.IsComputedIdentifierAccessor(parsed.MatcherEntry.Node) {
				return nil
			}

			matcherCall := testFramework.InvokedAccessorCall(parsed.MatcherEntry)
			if matcherCall == nil {
				return nil
			}
			return &shared.ExpectCall{
				Matcher:      parsed.Matcher,
				MatcherEntry: *parsed.MatcherEntry,
				MatcherCall:  matcherCall,
				Modifiers:    parsed.ModifierEntries,
			}
		}}
	},
	// NOTE: Unlike @vitest/eslint-plugin, the shared decision records `.not`
	// for equality checks against undefined too. buildFixes therefore turns
	// `not.toEqual(undefined)` into the equivalent `toBeDefined()`.
	// AllowFractionalNumbers stays false: Rstest's Vitest matcher layer
	// recommends toBeCloseTo for both positive and negative decimals.
	BuildFixes: buildFixes,
})
