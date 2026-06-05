package prefer_to_have_length

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func buildErrorUseToHaveLengthMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToHaveLength",
		Description: "Use `toHaveLength()` instead",
	}
}

func checkIsEqualityMethod(members []string) bool {
	for _, member := range members {
		if jestUtils.EQUALITY_METHOD_NAMES[member] {
			return true
		}
	}
	return false
}

func unwrapLengthAccessProperty(arg *ast.Node) *ast.Node {
	if arg == nil {
		return nil
	}
	arg = ast.SkipParentheses(arg)
	switch arg.Kind {
	case ast.KindElementAccessExpression:
		if ast.IsOptionalChain(arg) {
			return nil
		}
		el := arg.AsElementAccessExpression()
		if !jestUtils.IsNamedMember(ast.SkipParentheses(el.ArgumentExpression), "length") {
			return nil
		}
		return el.Expression
	case ast.KindPropertyAccessExpression:
		if ast.IsOptionalChain(arg) {
			return nil
		}
		pa := arg.AsPropertyAccessExpression()
		if !jestUtils.IsNamedMember(pa.Name(), "length") {
			return nil
		}
		return pa.Expression
	default:
		return nil
	}
}

var PreferToHaveLengthRule = rule.Rule{
	Name: "jest/prefer-to-have-length",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}

				if jestFnCall.Kind != jestUtils.JestFnTypeExpect {
					return
				}

				members := jestFnCall.Members
				if len(members) == 0 || !checkIsEqualityMethod(members) {
					return
				}

				expectCall := jestFnCall.Head.Local.Node.Parent
				if expectCall == nil || expectCall.Kind != ast.KindCallExpression {
					return
				}

				args := expectCall.Arguments()
				if len(args) == 0 {
					return
				}

				inner := unwrapLengthAccessProperty(args[0])
				if inner == nil {
					return
				}

				matcherCall := node.AsCallExpression()
				if matcherCall == nil {
					return
				}

				argList := matcherCall.Arguments
				if argList == nil || len(argList.Nodes) != 1 {
					return
				}

				beforeMatcher := jestUtils.ReceiverBeforeInvocation(node)
				if beforeMatcher == nil {
					return
				}

				sourceFile := ast.GetSourceFileOfNode(node)
				if sourceFile == nil {
					return
				}

				fileText := sourceFile.Text()
				innerText := rslintUtils.TrimmedNodeText(sourceFile, inner)
				expectedText := fileText[argList.Pos():argList.End()]
				middleText := fileText[expectCall.End():beforeMatcher.End()]

				newText := fmt.Sprintf("expect(%s)%s.toHaveLength(%s)", innerText, middleText, expectedText)
				trimmedExpectCall := rslintUtils.TrimNodeTextRange(sourceFile, expectCall)
				fixRange := core.NewTextRange(trimmedExpectCall.Pos(), node.End())
				reportNode := node

				if n := len(jestFnCall.MemberEntries); n > 0 {
					if entry := jestFnCall.MemberEntries[n-1].Node; entry != nil {
						reportNode = entry
					}
				}

				ctx.ReportNodeWithFixes(
					reportNode,
					buildErrorUseToHaveLengthMessage(),
					rule.RuleFixReplaceRange(fixRange, newText),
				)
			},
		}
	},
}
