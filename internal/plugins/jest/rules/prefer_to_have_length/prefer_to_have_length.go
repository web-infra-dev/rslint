package prefer_to_have_length

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_to_have_length"
)

func parseLengthAccessor(arg *ast.Node) (*ast.Node, testFramework.MemberEntry, bool) {
	if arg == nil {
		return nil, testFramework.MemberEntry{}, false
	}
	arg = ast.SkipParentheses(arg)
	if arg == nil || ast.IsOptionalChain(arg) {
		return nil, testFramework.MemberEntry{}, false
	}

	var receiver, name *ast.Node
	switch arg.Kind {
	case ast.KindElementAccessExpression:
		element := arg.AsElementAccessExpression()
		receiver, name = element.Expression, ast.SkipParentheses(element.ArgumentExpression)
	case ast.KindPropertyAccessExpression:
		property := arg.AsPropertyAccessExpression()
		receiver, name = property.Expression, property.Name()
	default:
		return nil, testFramework.MemberEntry{}, false
	}
	if !jestUtils.IsNamedMember(name, "length") || receiver == nil {
		return nil, testFramework.MemberEntry{}, false
	}
	return receiver, testFramework.MemberEntry{Name: "length", Node: name}, true
}

func buildFixes(ctx rule.RuleContext, match shared.Match) []rule.RuleFix {
	parsed := match.Expect
	if !parsed.CanFix || len(parsed.MatcherCall.Arguments()) != 1 {
		return nil
	}
	beforeMatcher := jestUtils.ReceiverBeforeInvocation(parsed.MatcherCall)
	if beforeMatcher == nil {
		return nil
	}

	argList := parsed.MatcherCall.AsCallExpression().Arguments
	if argList == nil {
		return nil
	}
	fileText := ctx.SourceFile.Text()
	innerText := utils.TrimmedNodeText(ctx.SourceFile, match.Receiver)
	expectedText := fileText[argList.Pos():argList.End()]
	middleText := fileText[parsed.HeadCall.End():beforeMatcher.End()]
	newText := fmt.Sprintf("expect(%s)%s.toHaveLength(%s)", innerText, middleText, expectedText)
	trimmedExpectCall := utils.TrimNodeTextRange(ctx.SourceFile, parsed.HeadCall)
	return []rule.RuleFix{rule.RuleFixReplaceRange(
		core.NewTextRange(trimmedExpectCall.Pos(), parsed.MatcherCall.End()),
		newText,
	)}
}

var PreferToHaveLengthRule = shared.NewRule(shared.Config{
	Name:                "jest/prefer-to-have-length",
	ParseLengthAccessor: parseLengthAccessor,
	BuildFixes:          buildFixes,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		return shared.Runtime{Parse: func(node *ast.Node) []*shared.ExpectCall {
			parsed := jestUtils.ParseJestFnCall(node, ctx)
			if parsed == nil || parsed.Kind != jestUtils.JestFnTypeExpect || parsed.MatcherEntry == nil ||
				len(node.Arguments()) != 1 {
				return nil
			}
			headCall := parsed.Head.Local.Node.Parent
			if headCall == nil || headCall.Kind != ast.KindCallExpression || jestUtils.ReceiverBeforeInvocation(node) == nil {
				return nil
			}
			return []*shared.ExpectCall{{
				HeadCall:     headCall,
				MatcherCall:  node,
				MatcherEntry: *parsed.MatcherEntry,
				Matcher:      parsed.Matcher,
				CanFix:       true,
			}}
		}}
	},
})
