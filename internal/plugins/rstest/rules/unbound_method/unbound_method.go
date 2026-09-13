package unbound_method

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	unboundMethod "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/unbound_method"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var UnboundMethodRule = rule.Rule{
	Name:             "rstest/unbound-method",
	Schema:           unboundMethod.UnboundMethodRule.Schema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var analysis *rstestUtils.RstestCallAnalysis
		stableRoots := map[*ast.Symbol]bool{}
		return unboundMethod.CreateListeners(ctx, options, func(node *ast.Node) bool {
			argument := node
			for argument.Parent != nil && utils.SkipAssertionsAndParens(argument.Parent) == node {
				argument = argument.Parent
			}
			call := argument.Parent
			if call == nil || !ast.IsCallExpression(call) || len(call.Arguments()) == 0 || call.Arguments()[0] != argument {
				return false
			}
			callee := utils.SkipAssertionsAndParens(call.Expression())
			root := callee
			for ast.IsAccessExpression(root) {
				root = utils.SkipAssertionsAndParens(root.Expression())
			}
			if ast.IsIdentifier(root) {
				symbol := ctx.Refs.Resolve(root)
				stable, checked := stableRoots[symbol]
				if !checked {
					stable = true
					for _, reference := range ctx.Refs.References(symbol) {
						for reference.Parent != nil && (utils.SkipAssertionsAndParens(reference.Parent) == reference ||
							(ast.IsAccessExpression(reference.Parent) && reference.Parent.Expression() == reference)) {
							reference = reference.Parent
						}
						if utils.IsWriteReference(reference) {
							stable = false
							break
						}
					}
					stableRoots[symbol] = stable
				}
				if !stable {
					return false
				}
			}
			if member, ok := utils.AccessExpressionStaticName(callee); ok && member == "mocked" && rstestUtils.IsUtilitiesObject(ctx, callee.Expression()) {
				return true
			}
			if analysis == nil {
				analysis = rstestUtils.GetRstestCallAnalysis(ctx)
			}
			parsed := analysis.ParseExpectCall(rstestUtils.FindTopMostCallExpression(call))
			if parsed == nil || parsed.Head != call || parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				(parsed.Entry != rstestUtils.RstestExpectEntryCall && parsed.Entry != rstestUtils.RstestExpectEntrySoft) || len(parsed.Matchers) == 0 {
				return false
			}
			for _, member := range parsed.MemberEntries {
				accessor := member.Node
				for accessor.Parent != nil && ast.IsParenthesizedExpression(accessor.Parent) {
					accessor = accessor.Parent
				}
				if accessor.Parent != nil && ast.IsElementAccessExpression(accessor.Parent) {
					if _, ok := utils.AccessExpressionStaticName(accessor.Parent); !ok {
						return false
					}
				}
			}
			for _, modifier := range parsed.Modifiers {
				if modifier != "not" {
					return false
				}
			}
			for _, matcher := range parsed.Matchers {
				if !isNonInvokingMatcher(matcher.Name) {
					return false
				}
			}
			return true
		})
	},
}

// Unknown/custom matchers may call the subject, so a throwing-matcher denylist is insufficient.
func isNonInvokingMatcher(name string) bool {
	if rstestUtils.RSTEST_SNAPSHOT_MATCHERS[name] {
		return name != "toThrowErrorMatchingSnapshot" && name != "toThrowErrorMatchingInlineSnapshot"
	}
	switch name {
	case "toBe", "toEqual", "toStrictEqual", "toMatchObject", "toMatch", "toContain", "toContainEqual", "toBeOneOf",
		"toBeTruthy", "toBeFalsy", "toBeNaN", "toBeUndefined", "toBeNull", "toBeNullable", "toBeDefined",
		"toBeTypeOf", "toBeInstanceOf", "toHaveLength", "toHaveProperty", "toBeCloseTo",
		"toBeGreaterThan", "toBeGreaterThanOrEqual", "toBeLessThan", "toBeLessThanOrEqual",
		"toHaveBeenCalled", "toBeCalled", "toHaveBeenCalledTimes", "toBeCalledTimes", "toHaveBeenCalledOnce",
		"toHaveBeenCalledWith", "toBeCalledWith", "toHaveBeenCalledExactlyOnceWith",
		"toHaveBeenNthCalledWith", "nthCalledWith", "toHaveBeenLastCalledWith", "lastCalledWith",
		"toHaveBeenCalledBefore", "toHaveBeenCalledAfter", "toHaveReturned", "toReturn",
		"toHaveReturnedTimes", "toReturnTimes", "toHaveReturnedWith", "toReturnWith",
		"toHaveLastReturnedWith", "lastReturnedWith", "toHaveNthReturnedWith", "nthReturnedWith",
		"toHaveResolved", "toHaveResolvedTimes", "toHaveResolvedWith", "toHaveLastResolvedWith", "toHaveNthResolvedWith",
		"equal", "equals", "eq", "eql", "eqls", "a", "an", "instanceof", "instanceOf", // cspell:ignore eqls
		"include", "contain", "contains", "includes", "above", "gt", "greaterThan",
		"least", "gte", "greaterThanOrEqual", "below", "lt", "lessThan", "most", "lte", "lessThanOrEqual", "within",
		"property", "ownProperty", "haveOwnProperty", "ownPropertyDescriptor", "haveOwnPropertyDescriptor",
		"length", "lengthOf", "match", "matches", "string", "keys", "key", "respondTo", "respondsTo",
		"closeTo", "approximately", "members", "oneOf", "containSubset",
		"ok", "true", "false", "null", "undefined", "NaN", "exist", "exists", "empty",
		"arguments", "Arguments", "numeric", "callable", "iterable", "extensible", "sealed", "frozen", "finite",
		"called", "calledOnce", "calledTwice", "calledThrice", "calledWith", "callCount",
		"calledOnceWith", "calledBefore", "calledAfter", "returned", "returnedWith", "returnedTimes":
		return true
	default:
		return false
	}
}
