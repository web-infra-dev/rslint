package prefer_to_contain

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// Message Builder

func buildUseToContainErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToContain",
		Description: "Use toContain() instead",
	}
}

func unwrapTypeAssertions(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		default:
			return node
		}
	}
	return node
}

func getIncludesCalleeName(callee *ast.Node) (receiver *ast.Node, ok bool) {
	if callee == nil {
		return nil, false
	}
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		prop := callee.AsPropertyAccessExpression()
		name := prop.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "includes" {
			return nil, false
		}

		return prop.Expression, true
	case ast.KindElementAccessExpression:
		el := callee.AsElementAccessExpression()
		arg := ast.SkipParentheses(el.ArgumentExpression)
		if arg == nil {
			return nil, false
		}

		switch arg.Kind {
		case ast.KindStringLiteral:
			if arg.AsStringLiteral().Text != "includes" {
				return nil, false
			}
		case ast.KindNoSubstitutionTemplateLiteral:
			if arg.AsNoSubstitutionTemplateLiteral().Text != "includes" {
				return nil, false
			}
		default:
			return nil, false
		}
		return el.Expression, true
	}

	return nil, false
}

func receiverBeforeInvocation(matcherCall *ast.Node) *ast.Node {
	if matcherCall == nil || matcherCall.Kind != ast.KindCallExpression {
		return nil
	}

	expr := matcherCall.AsCallExpression().Expression
	switch expr.Kind {
	case ast.KindPropertyAccessExpression:
		return expr.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return expr.AsElementAccessExpression().Expression
	}

	return nil
}

var PreferToContainRule = rule.Rule{
	Name: "jest/prefer-to-contain",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := jestUtils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != jestUtils.JestFnTypeExpect || len(jestFnCall.Members) == 0 {
					return
				}

				lastMember := jestFnCall.Members[len(jestFnCall.Members)-1]
				if !jestUtils.EQUALITY_METHOD_NAMES[lastMember] {
					return
				}

				expectCall := jestFnCall.Head.Local.Node.Parent
				if expectCall == nil || expectCall.Kind != ast.KindCallExpression {
					return
				}

				expectArgs := expectCall.Arguments()
				if len(expectArgs) == 0 {
					return
				}

				includesCall := ast.SkipParentheses(expectArgs[0])
				if includesCall == nil || includesCall.Kind != ast.KindCallExpression {
					return
				}

				includesExpr := includesCall.AsCallExpression()
				receiver, ok := getIncludesCalleeName(includesExpr.Expression)
				if !ok || includesExpr.Arguments == nil || len(includesExpr.Arguments.Nodes) != 1 {
					return
				}

				includesArg := includesExpr.Arguments.Nodes[0]
				if includesArg.Kind == ast.KindSpreadElement {
					return
				}

				matcherCall := node.AsCallExpression()
				if matcherCall.Arguments == nil || len(matcherCall.Arguments.Nodes) != 1 {
					return
				}

				matcherArg := matcherCall.Arguments.Nodes[0]
				if matcherArg.Kind == ast.KindSpreadElement {
					return
				}

				unwrapped := unwrapTypeAssertions(matcherArg)
				if unwrapped == nil {
					return
				}

				var isTrueLiteral bool
				switch unwrapped.Kind {
				case ast.KindTrueKeyword:
					isTrueLiteral = true
				case ast.KindFalseKeyword:
				default:
					return
				}

				hasNotModifier := slices.Contains(jestFnCall.Modifiers, "not")
				shouldNegate := isTrueLiteral == hasNotModifier
				if receiverBeforeInvocation(node) == nil {
					return
				}

				sourceFile := ctx.SourceFile
				receiverText := rslintUtils.TrimmedNodeText(sourceFile, receiver)
				includesArgText := rslintUtils.TrimmedNodeText(sourceFile, includesArg)

				// Preserve non-`not` modifiers (e.g. `resolves`, `rejects`) so the
				// asynchronous semantics of the original matcher chain are kept.
				var chainParts []string
				for _, modifier := range jestFnCall.Modifiers {
					if modifier == "not" {
						continue
					}
					chainParts = append(chainParts, modifier)
				}
				if shouldNegate {
					chainParts = append(chainParts, "not")
				}
				chainParts = append(chainParts, "toContain")
				chainReplacement := "." + strings.Join(chainParts, ".")

				reportNode := node
				if n := len(jestFnCall.MemberEntries); n > 0 {
					if entry := jestFnCall.MemberEntries[n-1].Node; entry != nil {
						reportNode = entry
					}
				}

				ctx.ReportNodeWithFixes(
					reportNode,
					buildUseToContainErrorMessage(),
					// Replace the `.includes(...)` argument of `expect(...)` with the
					// receiver, preserving the surrounding `expect(<callee>, ...)` tokens
					// (including any trailing comma after the first argument).
					rule.RuleFixReplace(sourceFile, expectArgs[0], receiverText),
					// Normalize the matcher chain (e.g. `.not.toEqual`, `['toEqual']`,
					// `['not'].toEqual`) into `.toContain` or `.not.toContain`.
					rule.RuleFixReplaceRange(
						core.NewTextRange(expectCall.End(), matcherCall.Expression.End()),
						chainReplacement,
					),
					// Replace the matcher's boolean argument with the original
					// `.includes` argument, preserving trailing commas/whitespace.
					rule.RuleFixReplace(sourceFile, matcherArg, includesArgText),
				)
			},
		}
	},
}
