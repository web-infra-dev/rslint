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
		var asymmetricMatchers *asymmetricMatcherAnalysis
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
					!isNonInvokingMatcher(matcher.Name) {
					return false
				}
				if usesCustomEqualityTesters(matcher.Name) && matcher.Entry.Call != nil {
					if asymmetricMatchers == nil {
						asymmetricMatchers = newAsymmetricMatcherAnalysis(ctx, analysis)
					}
					if asymmetricMatchers.matcherMayInvokeReceived(matcher.Name, matcher.Entry.Call) {
						return false
					}
				}
			}
			return true
		})
	},
}

type asymmetricMatcherAnalysis struct {
	ctx                rule.RuleContext
	expect             *rstestUtils.RstestCallAnalysis
	stableInitializers *utils.StaticStringEvaluator
	callResults        map[*ast.Node]bool
	visitingValues     map[*ast.Node]bool
	visitingTypes      map[*checker.Type]bool
	bindingMutations   map[*ast.Symbol]bindingMutation
	checkedBindings    map[*ast.Symbol]bool
}

type bindingMutation struct {
	changed             bool
	mayIntroduceMatcher bool
}

func newAsymmetricMatcherAnalysis(ctx rule.RuleContext, expect *rstestUtils.RstestCallAnalysis) *asymmetricMatcherAnalysis {
	return &asymmetricMatcherAnalysis{
		ctx:                ctx,
		expect:             expect,
		stableInitializers: utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs),
		callResults:        map[*ast.Node]bool{},
		visitingValues:     map[*ast.Node]bool{},
		visitingTypes:      map[*checker.Type]bool{},
		bindingMutations:   map[*ast.Symbol]bindingMutation{},
		checkedBindings:    map[*ast.Symbol]bool{},
	}
}

func (analysis *asymmetricMatcherAnalysis) matcherMayInvokeReceived(name string, call *ast.Node) bool {
	if result, ok := analysis.callResults[call]; ok {
		return result
	}
	result := false
	arguments := call.Arguments()
	if name == "toBeOneOf" {
		if len(arguments) != 0 {
			result = analysis.collectionMayContainInvokingMatcher(arguments[0])
		}
	} else {
		for _, argument := range arguments {
			if analysis.valueMayInvokeReceived(argument) {
				result = true
				break
			}
		}
	}
	analysis.callResults[call] = result
	return result
}

// valueMayInvokeReceived examines only values compared directly with the
// received method. Ordinary equality stops on a Function/Object type mismatch,
// so recursively scanning properties or elements here would report asymmetric
// matchers that the runtime never reaches.
func (analysis *asymmetricMatcherAnalysis) valueMayInvokeReceived(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || analysis.visitingValues[node] {
		return false
	}
	analysis.visitingValues[node] = true
	defer delete(analysis.visitingValues, node)

	if node.Kind == ast.KindCallExpression {
		parsed := analysis.expect.ParseExpectCall(node)
		if parsed != nil && parsed.Entry == rstestUtils.RstestExpectEntryStatic {
			if analysis.expect.IsExpectMatcherOverridden(parsed.Matcher) {
				return true
			}
			if !rstestUtils.RSTEST_ASYMMETRIC_MATCHERS[parsed.Matcher] {
				return false
			}
			switch parsed.Matcher {
			case "toSatisfy", "schemaMatching":
				return true
			case "toBeOneOf":
				arguments := node.Arguments()
				return len(arguments) != 0 && analysis.collectionMayContainInvokingMatcher(arguments[0])
			default:
				return false
			}
		}
	}
	t := analysis.ctx.TypeChecker.GetTypeAtLocation(node)
	if !utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown) &&
		analysis.typeHasCallableAsymmetricMatcher(t, node) {
		return true
	}
	if initializer, resolved, mutation := analysis.localInitializer(node); mutation.mayIntroduceMatcher {
		return true
	} else if mutation.changed {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown) ||
			analysis.typeHasCallableAsymmetricMatcher(t, node)
	} else if resolved {
		return analysis.valueMayInvokeReceived(initializer)
	}
	switch node.Kind {
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return analysis.valueMayInvokeReceived(conditional.WhenTrue) ||
			analysis.valueMayInvokeReceived(conditional.WhenFalse)
	}
	return analysis.typeHasCallableAsymmetricMatcher(t, node)
}

func (analysis *asymmetricMatcherAnalysis) collectionMayContainInvokingMatcher(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || analysis.visitingValues[node] {
		return false
	}
	analysis.visitingValues[node] = true
	defer delete(analysis.visitingValues, node)

	switch node.Kind {
	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindSpreadElement {
				if analysis.collectionMayContainInvokingMatcher(element.AsSpreadElement().Expression) {
					return true
				}
			} else if analysis.sampleMayInvokeReceived(element) {
				return true
			}
		}
		return false
	case ast.KindNewExpression:
		if utils.IsBuiltinSymbolLike(analysis.ctx.Program(), analysis.ctx.TypeChecker, analysis.ctx.TypeChecker.GetTypeAtLocation(node), "Set") {
			arguments := node.Arguments()
			return len(arguments) != 0 && analysis.collectionMayContainInvokingMatcher(arguments[0])
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return analysis.collectionMayContainInvokingMatcher(conditional.WhenTrue) ||
			analysis.collectionMayContainInvokingMatcher(conditional.WhenFalse)
	}
	if initializer, resolved, mutation := analysis.localInitializer(node); mutation.mayIntroduceMatcher {
		return true
	} else if mutation.changed {
		return analysis.collectionTypeMayContainInvokingMatcher(analysis.ctx.TypeChecker.GetTypeAtLocation(node), node)
	} else if resolved {
		return analysis.collectionMayContainInvokingMatcher(initializer)
	}
	return analysis.collectionTypeMayContainInvokingMatcher(analysis.ctx.TypeChecker.GetTypeAtLocation(node), node)
}

func (analysis *asymmetricMatcherAnalysis) sampleMayInvokeReceived(node *ast.Node) bool {
	if analysis.valueMayInvokeReceived(node) {
		return true
	}
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindCallExpression {
		parsed := analysis.expect.ParseExpectCall(node)
		if parsed != nil && parsed.Entry == rstestUtils.RstestExpectEntryStatic &&
			rstestUtils.RSTEST_ASYMMETRIC_MATCHERS[parsed.Matcher] &&
			!analysis.expect.IsExpectMatcherOverridden(parsed.Matcher) {
			return false
		}
	}
	if initializer, resolved, mutation := analysis.localInitializer(node); mutation.mayIntroduceMatcher {
		return true
	} else if mutation.changed {
		t := analysis.ctx.TypeChecker.GetTypeAtLocation(node)
		return utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown) ||
			analysis.typeHasCallableAsymmetricMatcher(t, node)
	} else if resolved {
		return analysis.sampleMayInvokeReceived(initializer)
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		return analysis.sampleMayInvokeReceived(conditional.WhenTrue) ||
			analysis.sampleMayInvokeReceived(conditional.WhenFalse)
	}
	t := analysis.ctx.TypeChecker.GetTypeAtLocation(node)
	return utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown)
}

func (analysis *asymmetricMatcherAnalysis) typeHasCallableAsymmetricMatcher(t *checker.Type, location *ast.Node) bool {
	if t == nil {
		return false
	}
	for _, part := range utils.UnionTypeParts(checker.Checker_getApparentType(analysis.ctx.TypeChecker, t)) {
		member := checker.Checker_getPropertyOfType(analysis.ctx.TypeChecker, part, "asymmetricMatch")
		if member != nil {
			memberType := analysis.ctx.TypeChecker.GetTypeOfSymbolAtLocation(member, location)
			if len(utils.GetCallSignatures(analysis.ctx.TypeChecker, memberType)) != 0 {
				return true
			}
		}
	}
	return false
}

func (analysis *asymmetricMatcherAnalysis) collectionTypeMayContainInvokingMatcher(t *checker.Type, location *ast.Node) bool {
	if t == nil || analysis.visitingTypes[t] {
		return false
	}
	analysis.visitingTypes[t] = true
	defer delete(analysis.visitingTypes, t)

	for _, part := range utils.UnionTypeParts(checker.Checker_getApparentType(analysis.ctx.TypeChecker, t)) {
		if elementType := utils.GetNumberIndexType(analysis.ctx.TypeChecker, part); elementType != nil {
			if utils.IsTypeFlagSet(elementType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) ||
				analysis.typeHasCallableAsymmetricMatcher(elementType, location) {
				return true
			}
		}
		if !utils.IsBuiltinSymbolLike(analysis.ctx.Program(), analysis.ctx.TypeChecker, part, "Set") {
			continue
		}
		for _, elementType := range checker.Checker_getTypeArguments(analysis.ctx.TypeChecker, part) {
			if utils.IsTypeFlagSet(elementType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) ||
				analysis.typeHasCallableAsymmetricMatcher(elementType, location) {
				return true
			}
		}
	}
	return false
}

func (analysis *asymmetricMatcherAnalysis) localInitializer(node *ast.Node) (initializer *ast.Node, resolved bool, mutation bindingMutation) {
	initializer, ok := analysis.stableInitializers.ResolveIdentifierInitializer(node)
	if !ok {
		return nil, false, bindingMutation{}
	}
	return initializer, true, analysis.bindingMutation(node)
}

func (analysis *asymmetricMatcherAnalysis) bindingMutation(node *ast.Node) bindingMutation {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || !ast.IsIdentifier(node) {
		return bindingMutation{}
	}
	symbol := analysis.ctx.Refs.Resolve(node)
	return analysis.bindingMutationForSymbol(symbol)
}

func (analysis *asymmetricMatcherAnalysis) bindingMutationForSymbol(symbol *ast.Symbol) bindingMutation {
	if symbol == nil {
		return bindingMutation{changed: true, mayIntroduceMatcher: true}
	}
	if analysis.checkedBindings[symbol] {
		return analysis.bindingMutations[symbol]
	}
	analysis.checkedBindings[symbol] = true
	for _, reference := range analysis.ctx.Refs.References(symbol) {
		outer := reference
		var access *ast.Node
		for outer.Parent != nil {
			parent := outer.Parent
			if ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions) && parent.Expression() == outer {
				outer = parent
				continue
			}
			if ast.IsAccessExpression(parent) && parent.Expression() == outer {
				access = parent
				outer = parent
				continue
			}
			break
		}
		if utils.IsWriteReference(outer) {
			mutation := bindingMutation{changed: true}
			if access == nil {
				mutation.mayIntroduceMatcher = true
			} else if name, ok := utils.AccessExpressionStaticName(access); !ok || name == "asymmetricMatch" {
				mutation.mayIntroduceMatcher = true
			}
			analysis.bindingMutations[symbol] = mutation
			if mutation.mayIntroduceMatcher {
				return mutation
			}
			continue
		}
		if outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression {
			analysis.bindingMutations[symbol] = bindingMutation{changed: true}
			continue
		}
		if access == nil {
			if alias, aliasName := analysis.aliasInitializedFrom(outer); alias != nil && alias != symbol {
				aliasMutation := analysis.bindingMutationForSymbol(alias)
				if aliasMutation.changed {
					aliasType := analysis.ctx.TypeChecker.GetTypeAtLocation(aliasName)
					aliasMutation.mayIntroduceMatcher = aliasMutation.mayIntroduceMatcher ||
						utils.IsTypeFlagSet(aliasType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) ||
						analysis.typeHasCallableAsymmetricMatcher(aliasType, outer) ||
						analysis.collectionTypeMayContainInvokingMatcher(aliasType, outer)
					// A mutation through an alias invalidates the original initializer.
					// The caller will fall back to the original binding's type unless
					// the alias mutation itself is inherently opaque.
					analysis.bindingMutations[symbol] = aliasMutation
				}
				if aliasMutation.mayIntroduceMatcher {
					return aliasMutation
				}
			}
			continue
		}
		if outer.Parent == nil || outer.Parent.Kind != ast.KindCallExpression || outer.Parent.Expression() != outer {
			continue
		}
		if name, ok := utils.AccessExpressionStaticName(access); !ok || collectionMutatingMethod(name) {
			mutation := bindingMutation{changed: true, mayIntroduceMatcher: !ok}
			analysis.bindingMutations[symbol] = mutation
			if mutation.mayIntroduceMatcher {
				return mutation
			}
		}
	}
	return analysis.bindingMutations[symbol]
}

func (analysis *asymmetricMatcherAnalysis) aliasInitializedFrom(node *ast.Node) (*ast.Symbol, *ast.Node) {
	for node != nil && node.Parent != nil && ast.IsOuterExpression(node.Parent, ast.OEKParentheses|ast.OEKAssertions) && node.Parent.Expression() == node {
		node = node.Parent
	}
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindVariableDeclaration {
		return nil, nil
	}
	declaration := node.Parent.AsVariableDeclaration()
	if declaration.Initializer != node || !ast.IsIdentifier(declaration.Name()) {
		return nil, nil
	}
	return utils.GetVariableDeclarationSymbol(node.Parent, analysis.ctx.TypeChecker), declaration.Name()
}

func collectionMutatingMethod(name string) bool {
	switch name {
	case "add", "clear", "copyWithin", "delete", "fill", "pop", "push", "reverse", "shift", "sort", "splice", "unshift":
		return true
	default:
		return false
	}
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
