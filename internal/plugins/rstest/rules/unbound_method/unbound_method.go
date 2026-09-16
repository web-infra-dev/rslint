package unbound_method

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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
		asymmetricMatcherArguments := map[*ast.Node]bool{}
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
						for reference.Parent != nil && (utils.SkipAssertionsAndParens(reference.Parent) == utils.SkipAssertionsAndParens(reference) ||
							(ast.IsAccessExpression(reference.Parent) && reference.Parent.Expression() == reference)) {
							reference = reference.Parent
						}
						if utils.IsWriteReference(reference) || (reference.Parent != nil && reference.Parent.Kind == ast.KindDeleteExpression) {
							stable = false
							break
						}
					}
					stableRoots[symbol] = stable
				}
				if !stable {
					return false
				}
			} else {
				if analysis == nil {
					analysis = rstestUtils.GetRstestCallAnalysis(ctx)
				}
				if analysis.HasImportMetaRstestWrites() {
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
			// Modifiers only contains entries before the first matcher, while Chai
			// permits promise modifiers later in the chain. `rejects` invokes a
			// function subject regardless of where it appears.
			for _, member := range parsed.Members {
				if member == "rejects" {
					return false
				}
			}
			for _, matcher := range parsed.Matchers {
				if analysis.IsExpectMatcherOverridden(matcher.Name) ||
					(analysis.HasCustomEqualityTesters() && usesCustomEqualityTesters(matcher.Name)) ||
					(usesCustomEqualityTesters(matcher.Name) && matcher.Entry.Call != nil &&
						matcherArgumentsContainAsymmetricMatcher(ctx, analysis, matcher.Entry.Call, asymmetricMatcherArguments)) ||
					!isNonInvokingMatcher(matcher.Name) {
					return false
				}
			}
			return true
		})
	},
}

func matcherArgumentsContainAsymmetricMatcher(
	ctx rule.RuleContext,
	analysis *rstestUtils.RstestCallAnalysis,
	call *ast.Node,
	cache map[*ast.Node]bool,
) bool {
	if result, ok := cache[call]; ok {
		return result
	}
	result := false
	for _, argument := range call.Arguments() {
		if expressionContainsAsymmetricMatcher(ctx, analysis, argument) {
			result = true
			break
		}
	}
	cache[call] = result
	return result
}

// expressionContainsAsymmetricMatcher follows only values traversed by an
// equality comparison. It deliberately does not walk into function bodies or
// call arguments, which are not themselves compared.
func expressionContainsAsymmetricMatcher(ctx rule.RuleContext, analysis *rstestUtils.RstestCallAnalysis, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || isSafeBuiltinAsymmetricMatcher(analysis, node) {
		return false
	}
	t := ctx.TypeChecker.GetTypeAtLocation(node)
	for _, part := range utils.UnionTypeParts(checker.Checker_getApparentType(ctx.TypeChecker, t)) {
		member := checker.Checker_getPropertyOfType(ctx.TypeChecker, part, "asymmetricMatch")
		if member != nil {
			memberType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(member, node)
			if len(utils.GetCallSignatures(ctx.TypeChecker, memberType)) != 0 {
				return true
			}
		}
	}
	switch node.Kind {
	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			if expressionContainsAsymmetricMatcher(ctx, analysis, element) {
				return true
			}
		}
	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			var value *ast.Node
			switch property.Kind {
			case ast.KindPropertyAssignment:
				value = property.AsPropertyAssignment().Initializer
			case ast.KindShorthandPropertyAssignment:
				value = property.Name()
			case ast.KindSpreadAssignment:
				value = property.AsSpreadAssignment().Expression
			}
			if expressionContainsAsymmetricMatcher(ctx, analysis, value) {
				return true
			}
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return expressionContainsAsymmetricMatcher(ctx, analysis, conditional.WhenTrue) ||
			expressionContainsAsymmetricMatcher(ctx, analysis, conditional.WhenFalse)
	}
	return false
}

func isSafeBuiltinAsymmetricMatcher(analysis *rstestUtils.RstestCallAnalysis, node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	parsed := analysis.ParseExpectCall(node)
	return parsed != nil && parsed.Entry == rstestUtils.RstestExpectEntryStatic &&
		rstestUtils.RSTEST_ASYMMETRIC_MATCHERS[parsed.Matcher] &&
		parsed.Matcher != "toSatisfy" && parsed.Matcher != "schemaMatching"
}

func usesCustomEqualityTesters(name string) bool {
	switch name {
	case "toBe", "toEqual", "toStrictEqual", "toMatchObject", "toBeOneOf":
		return true
	default:
		return false
	}
}

// Unknown/custom matchers may call the subject, so a throwing-matcher denylist is insufficient.
func isNonInvokingMatcher(name string) bool {
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
