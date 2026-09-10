package prefer_to_be

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_to_be"
)

func message(kind shared.Kind) rule.RuleMessage {
	switch kind {
	case shared.KindToBe:
		return rule.RuleMessage{Id: "useToBe", Description: "Use `toBe` when expecting primitive literals"}
	case shared.KindNull:
		return rule.RuleMessage{Id: "useToBeNull", Description: "Use `toBeNull` instead"}
	case shared.KindNaN:
		return rule.RuleMessage{Id: "useToBeNaN", Description: "Use `toBeNaN` instead"}
	case shared.KindUndefined:
		return rule.RuleMessage{Id: "useToBeUndefined", Description: "Use `toBeUndefined` instead"}
	case shared.KindDefined:
		return rule.RuleMessage{Id: "useToBeDefined", Description: "Use `toBeDefined` instead"}
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

	// NOTE: replaceAccessorFixer keeps an identifier accessor bare, which turns
	// `expect(x)[toEqual](1)` into a reference to an undeclared `toBe`. A
	// computed key is quoted instead so the fixed code still runs.
	renameFix, ok := jestUtils.ReplaceMemberNameFix(
		ctx,
		&match.Expect.MatcherEntry,
		replacement,
	)
	if !ok {
		return nil
	}
	fixes := []rule.RuleFix{renameFix}

	arguments := match.Expect.MatcherCall.Arguments()
	if match.Kind != shared.KindToBe {
		// NOTE: the dedicated matchers declare no type parameters, so type
		// arguments written for toBe/toEqual have to go with the value
		// arguments. Upstream leaves them behind and produces TS2558.
		if typeArgumentsRange, hasTypeArguments := testFramework.CallTypeArgumentListRange(
			ctx.SourceFile,
			match.Expect.MatcherCall,
		); hasTypeArguments {
			fixes = append(fixes, rule.RuleFixRemoveRange(typeArgumentsRange))
		}
	}
	if match.Kind != shared.KindToBe && len(arguments) > 0 {
		argumentsRange, ok := testFramework.CallArgumentListRange(
			ctx.SourceFile,
			match.Expect.MatcherCall,
		)
		if !ok {
			return nil
		}
		// Match removeExtraArgumentsFixer: keep comments before the first
		// argument, but remove the arguments, trailing comma and later trivia.
		argumentsRange = core.NewTextRange(
			internalUtils.TrimNodeTextRange(ctx.SourceFile, arguments[0]).Pos(),
			argumentsRange.End(),
		)
		fixes = append(fixes, rule.RuleFixRemoveRange(argumentsRange))
	}

	if match.NotEntry != nil {
		removeFixes, ok := jestUtils.RemoveMemberAccessorFixes(ctx, match.NotEntry)
		if !ok {
			return nil
		}
		fixes = append(fixes, removeFixes...)
	}
	return fixes
}

func hasParenthesizedOptionalChainBoundary(entries []jestUtils.ParsedJestFnMemberEntry) bool {
	for index := range entries {
		_, accessor := testFramework.AccessorReceiverAndParent(&entries[index])
		if accessor == nil || !ast.IsOptionalChain(accessor) {
			continue
		}
		if parent := accessor.Parent; parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			return true
		}
	}
	return false
}

var PreferToBeRule = shared.NewRule(shared.Config{
	Name:                   "jest/prefer-to-be",
	AllowFractionalNumbers: true,
	Message:                message,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{Parse: func(node *ast.Node) *shared.ExpectCall {
			parsed := jestUtils.ParseJestFnCall(node, ctx)
			if parsed == nil ||
				parsed.Kind != jestUtils.JestFnTypeExpect ||
				parsed.MatcherEntry == nil ||
				hasParenthesizedOptionalChainBoundary(parsed.MemberEntries) {
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
	BuildFixes: buildFixes,
})
