package prefer_to_be

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message Builders

func buildUseToBeErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBe",
		Description: "Use `toBe` when expecting primitive literals",
	}
}

func buildUseToBeUndefinedErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeUndefined",
		Description: "Use `toBeUndefined` instead",
	}
}

func buildUseToBeDefinedErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeDefined",
		Description: "Use `toBeDefined` instead",
	}
}

func buildUseToBeNullErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeNull",
		Description: "Use `toBeNull` instead",
	}
}

func buildUseToBeNaNErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useToBeNaN",
		Description: "Use `toBeNaN` instead",
	}
}

type preferKind uint8

const (
	preferKindToBe preferKind = iota
	preferKindNull
	preferKindNaN
	preferKindUndefined
	preferKindDefined
)

// firstMatcherArgument returns expect(...).matcher(arg0)’s first argument
// after peeling type assertions; nil if the matcher call has no arguments.
func firstMatcherArgument(matcherCall *ast.Node) *ast.Node {
	call := matcherCall.AsCallExpression()
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return nil
	}
	return utils.UnwrapBasicTypeAssertions(call.Arguments.Nodes[0])
}

func isNullLiteral(arg *ast.Node) bool {
	return arg != nil && arg.Kind == ast.KindNullKeyword
}

func isIdentifier(arg *ast.Node, name string) bool {
	return arg != nil && arg.Kind == ast.KindIdentifier && arg.AsIdentifier().Text == name
}

// shouldUseToBe reports whether arg is a primitive literal for which `toBe`
// is preferred over `toEqual` / `toStrictEqual`. arg must already be
// unwrapped (see firstMatcherArgument).
func shouldUseToBe(arg *ast.Node) bool {
	if arg == nil {
		return false
	}
	if arg.Kind == ast.KindPrefixUnaryExpression {
		unary := arg.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken {
			arg = utils.UnwrapBasicTypeAssertions(unary.Operand)
		}
	}
	if arg == nil {
		return false
	}
	switch arg.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression:
		return true
	default:
		return false
	}
}

func appendRemoveNotModifierFix(
	ctx rule.RuleContext,
	fixes []rule.RuleFix,
	jestFnCall *utils.ParsedJestFnCall,
) ([]rule.RuleFix, bool) {
	for _, modEntry := range jestFnCall.ModifierEntries {
		if modEntry.Name != "not" || modEntry.Node == nil {
			continue
		}

		removeFixes, ok := utils.RemoveMemberAccessorFixes(ctx, &modEntry)
		if !ok {
			return nil, false
		}
		return append(fixes, removeFixes...), true
	}
	return nil, false
}

// reportPreferToBe emits a diagnostic and an autofix replacing the matcher
// chain with the canonical equivalent.
//
// stripNot removes a leading `.not` / `["not"]` segment when true (e.g.
// `not.toBe(undefined)` → `toBeDefined()`).
func reportPreferToBe(ctx rule.RuleContext, kind preferKind, jestFnCall *utils.ParsedJestFnCall, node *ast.Node, stripNot bool) {
	n := len(jestFnCall.MemberEntries)
	if n == 0 {
		return
	}

	var newMatcher string
	var msg rule.RuleMessage
	dropArgs := true
	switch kind {
	case preferKindToBe:
		newMatcher = "toBe"
		msg = buildUseToBeErrorMessage()
		dropArgs = false
	case preferKindNull:
		newMatcher = "toBeNull"
		msg = buildUseToBeNullErrorMessage()
	case preferKindNaN:
		newMatcher = "toBeNaN"
		msg = buildUseToBeNaNErrorMessage()
	case preferKindUndefined:
		newMatcher = "toBeUndefined"
		msg = buildUseToBeUndefinedErrorMessage()
	case preferKindDefined:
		newMatcher = "toBeDefined"
		msg = buildUseToBeDefinedErrorMessage()
	default:
		return
	}

	matcherEntry := jestFnCall.MemberEntries[n-1]
	reportNode := matcherEntry.Node
	if reportNode == nil {
		reportNode = node
	}
	reportWithoutFix := func() {
		ctx.ReportNode(reportNode, msg)
	}

	matcherCall := node.AsCallExpression()
	if matcherCall == nil {
		reportWithoutFix()
		return
	}
	matcherExpr := matcherCall.Expression
	if matcherExpr == nil {
		reportWithoutFix()
		return
	}

	// At most: rename matcher, replace (...)/[], strip `.not`.
	fixes := make([]rule.RuleFix, 0, 3)
	renameFix, ok := utils.ReplaceMemberNameFix(ctx, &matcherEntry, newMatcher)
	if !ok {
		reportWithoutFix()
		return
	}
	fixes = append(fixes, renameFix)

	if dropArgs {
		callFix, ok := utils.ReplaceCallSuffixFix(ctx.SourceFile, node, "()")
		if !ok {
			reportWithoutFix()
			return
		}
		fixes = append(fixes, callFix)
	}

	if stripNot {
		var ok bool
		fixes, ok = appendRemoveNotModifierFix(ctx, fixes, jestFnCall)
		if !ok {
			reportWithoutFix()
			return
		}
	}

	ctx.ReportNodeWithFixes(reportNode, msg, fixes...)
}

var PreferToBeRule = rule.Rule{
	Name:   "jest/prefer-to-be",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeExpect {
					return
				}

				members := jestFnCall.Members
				if len(members) == 0 {
					return
				}

				matcherName := members[len(members)-1]
				hasNot := slices.Contains(jestFnCall.Modifiers, "not")

				if hasNot {
					switch matcherName {
					case "toBeUndefined":
						reportPreferToBe(ctx, preferKindDefined, jestFnCall, node, true)
						return
					case "toBeDefined":
						reportPreferToBe(ctx, preferKindUndefined, jestFnCall, node, true)
						return
					}
				}

				if !utils.EQUALITY_METHOD_NAMES[matcherName] {
					return
				}

				firstArg := firstMatcherArgument(node)
				if firstArg == nil {
					return
				}

				switch {
				case isNullLiteral(firstArg):
					reportPreferToBe(ctx, preferKindNull, jestFnCall, node, false)
				case isIdentifier(firstArg, "undefined"):
					if hasNot {
						reportPreferToBe(ctx, preferKindDefined, jestFnCall, node, true)
					} else {
						reportPreferToBe(ctx, preferKindUndefined, jestFnCall, node, false)
					}
				case isIdentifier(firstArg, "NaN"):
					reportPreferToBe(ctx, preferKindNaN, jestFnCall, node, false)
				case matcherName != "toBe" && shouldUseToBe(firstArg):
					reportPreferToBe(ctx, preferKindToBe, jestFnCall, node, false)
				}
			},
		}
	},
}
