package no_useless_backreference

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-useless-backreference
var NoUselessBackreferenceRule = rule.Rule{
	Name: "no-useless-backreference",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		mayUseRegExp := sourceMayUseRegExp(ctx)
		var calleeCache *regExpCalleeCache
		listeners := rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				if mayUseRegExp && isRegexLiteralHandledByConstructor(ctx, node, calleeCache) {
					return
				}
				pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
				if pattern == "" && flags == "" {
					return
				}
				rxFlags := utils.ParseRegexFlags(flags)
				checkRegex(ctx, node, pattern, rxFlags)
			},
		}
		if !mayUseRegExp {
			return listeners
		}
		calleeCache = &regExpCalleeCache{}

		var eval *utils.StaticStringEvaluator
		getEval := func() *utils.StaticStringEvaluator {
			if eval == nil {
				eval = utils.NewStaticStringEvaluatorWithSourceFile(ctx.TypeChecker, ctx.SourceFile)
			}
			return eval
		}
		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			call := node.AsCallExpression()
			handleRegExpConstructor(ctx, node, call.Expression, call.Arguments, getEval, calleeCache)
		}
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			handleRegExpConstructor(ctx, node, newExpr.Expression, newExpr.Arguments, getEval, calleeCache)
		}
		return listeners
	},
}

func sourceMayUseRegExp(ctx rule.RuleContext) bool {
	// With type information, an alias can arrive through an import or ambient
	// declaration without any RegExp-related identifier in this source file.
	// Keep the broad listeners in that case so those aliases remain observable.
	if ctx.TypeChecker != nil && ctx.Program != nil {
		return true
	}

	// Without type information only the syntactically recognized global
	// constructor forms can match, so files lacking all of their names can
	// safely avoid broad call/new listeners.
	sourceFile := ctx.SourceFile
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	for _, name := range [...]string{
		"RegExp",
		"globalThis",
		"window",
		"self",
		"global",
	} {
		if _, ok := sourceFile.Identifiers[name]; ok {
			return true
		}
	}
	return false
}

func handleRegExpConstructor(
	ctx rule.RuleContext,
	callNode *ast.Node,
	callee *ast.Node,
	args *ast.NodeList,
	getEval func() *utils.StaticStringEvaluator,
	calleeCache *regExpCalleeCache,
) {
	if args == nil || len(args.Nodes) == 0 {
		return
	}
	patternNode := ast.SkipParentheses(args.Nodes[0])
	if patternNode == nil {
		return
	}

	flags := ""
	var pattern string
	patternReady := true
	switch patternNode.Kind {
	case ast.KindRegularExpressionLiteral:
		pattern, flags = utils.ExtractRegexPatternAndFlags(patternNode.Text())
	case ast.KindStringLiteral:
		pattern = patternNode.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		pattern = patternNode.AsNoSubstitutionTemplateLiteral().Text
	default:
		if !mayEvaluateToRegexPattern(patternNode) {
			return
		}
		patternReady = false
	}

	// Most calls with literal arguments are unrelated to RegExp. Reject them
	// before asking the checker for the callee's flow-sensitive type.
	if patternReady && !mayContainBackreference(pattern) {
		return
	}

	callee = ast.SkipParentheses(callee)
	if !isBuiltinRegExpCallee(ctx, callee, calleeCache) {
		return
	}

	if !patternReady {
		var patternOk bool
		pattern, patternOk = getEval().Eval(patternNode)
		if !patternOk {
			return
		}
	}

	if len(args.Nodes) >= 2 {
		flagsNode := ast.SkipParentheses(args.Nodes[1])
		if flagsNode != nil {
			if v, ok := literalStringValue(flagsNode); ok {
				flags = v
			} else if v, ok := getEval().Eval(flagsNode); ok {
				flags = v
			} else {
				flags = ""
			}
		}
	}

	rxFlags := utils.ParseRegexFlags(flags)
	checkRegex(ctx, callNode, pattern, rxFlags)
}

// mayEvaluateToRegexPattern is a conservative negative syntactic filter for
// StaticStringEvaluator.Eval. Only expression forms whose values are
// intrinsically non-string are rejected. Unknown and future syntax stays on the
// normal checker/evaluator path so extending StaticStringEvaluator cannot turn
// this optimization into a false negative.
func mayEvaluateToRegexPattern(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression,
		ast.KindPrefixUnaryExpression,
		ast.KindPostfixUnaryExpression,
		ast.KindDeleteExpression,
		ast.KindVoidExpression:
		return false
	default:
		return true
	}
}

func literalStringValue(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	default:
		return "", false
	}
}

// isRegexLiteralHandledByConstructor returns true when this regex literal is
// the first arg of a `RegExp(literal, flags)` call — in that case the
// constructor listener owns it (using the override flags).
func isRegexLiteralHandledByConstructor(ctx rule.RuleContext, node *ast.Node, calleeCache *regExpCalleeCache) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	var callee *ast.Node
	var args *ast.NodeList
	switch parent.Kind {
	case ast.KindCallExpression:
		c := parent.AsCallExpression()
		callee = c.Expression
		args = c.Arguments
	case ast.KindNewExpression:
		n := parent.AsNewExpression()
		callee = n.Expression
		args = n.Arguments
	default:
		return false
	}
	if !isBuiltinRegExpCallee(ctx, ast.SkipParentheses(callee), calleeCache) {
		return false
	}
	if args == nil || len(args.Nodes) == 0 {
		return false
	}
	if first := ast.SkipParentheses(args.Nodes[0]); first != node {
		return false
	}
	return true
}

type regExpCalleeCache struct {
	// Cache the pure builtin-type predicate, not identifier symbols: the
	// checker may give one const/import/global symbol different flow-narrowed
	// types at different call sites.
	types map[*checker.Type]bool
}

func isBuiltinRegExpCallee(ctx rule.RuleContext, callee *ast.Node, calleeCache *regExpCalleeCache) bool {
	if callee == nil {
		return false
	}
	// Config `off` un-declares `RegExp`: the constructor path goes dead and
	// regex-literal arguments fall back to the plain literal listener, matching
	// ESLint's ReferenceTracker walk over the global scope.
	if declared, ok := ctx.Globals["RegExp"]; ok && !declared {
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		name := callee.AsIdentifier().Text
		if name == "RegExp" {
			// Direct `RegExp` reference — must not be shadowed.
			if utils.IsShadowed(callee, "RegExp") {
				return false
			}
			if ctx.TypeChecker != nil {
				sym := ctx.TypeChecker.GetSymbolAtLocation(callee)
				if sym == nil {
					return false
				}
				return !utils.IsSymbolDeclaredInFile(sym, ctx.SourceFile)
			}
			return true
		}
		// Identifier alias such as `const r = RegExp; new r(...)` — only the
		// type check can recognize this. No syntactic fallback to avoid
		// over-matching arbitrary identifiers when type info is unavailable.
		if ctx.TypeChecker != nil && ctx.Program != nil {
			t := ctx.TypeChecker.GetTypeAtLocation(callee)
			if t == nil {
				return false
			}
			if cached, ok := calleeCache.types[t]; ok {
				return cached
			}
			result := utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, t, "RegExpConstructor")
			if calleeCache.types == nil {
				calleeCache.types = make(map[*checker.Type]bool)
			}
			calleeCache.types[t] = result
			return result
		}
		return false
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		pae := callee.AsPropertyAccessExpression()
		if pae.Name() != nil && pae.Name().Kind == ast.KindIdentifier && pae.Name().AsIdentifier().Text == "RegExp" {
			if pae.Expression != nil && pae.Expression.Kind == ast.KindIdentifier {
				name := pae.Expression.AsIdentifier().Text
				return name == "globalThis" || name == "window" || name == "self" || name == "global"
			}
		}
	}
	return false
}

// checkRegex parses the pattern and reports useless backreferences. `node` is
// the AST node receiving the diagnostic (the regex literal or RegExp call).
func checkRegex(ctx rule.RuleContext, node *ast.Node, pattern string, flags utils.RegexFlags) {
	if !mayContainBackreference(pattern) {
		return
	}
	_, brefs, ok := parsePattern(pattern, flags)
	if !ok {
		return
	}

	var reportRange core.TextRange
	hasReportRange := false
	for _, bref := range brefs {
		first, problemCount, hasProblem := analyzeBackref(bref)
		if !hasProblem {
			continue
		}
		if !hasReportRange {
			reportRange = utils.TrimNodeTextRange(ctx.SourceFile, node)
			hasReportRange = true
		}
		reportBackref(ctx, reportRange, bref, first, problemCount)
	}
}

func mayContainBackreference(pattern string) bool {
	// Patterns without a numeric or named backreference do not need an AST.
	// Advance over escape pairs so escaped backslashes (`\\1`) do not look
	// like references.
	for i := 0; i+1 < len(pattern); {
		if pattern[i] != '\\' {
			i++
			continue
		}
		next := pattern[i+1]
		if next == 'k' || next >= '1' && next <= '9' {
			return true
		}
		i += 2
	}
	return false
}

type problem struct {
	messageId string
	group     *rxNode
}

func analyzeBackref(bref *rxNode) (first problem, problemCount int, hasProblem bool) {
	groups := bref.resolved
	if len(groups) == 0 {
		return problem{}, 0, false
	}

	var firstSameDisjunction problem
	sameDisjunctionCount := 0
	for _, group := range groups {
		current, ok := classifyPair(bref, group)
		if !ok {
			// If any pair has no problem, the backreference can match.
			return problem{}, 0, false
		}
		if problemCount == 0 {
			first = current
		}
		problemCount++
		if current.messageId != "disjunctive" {
			if sameDisjunctionCount == 0 {
				firstSameDisjunction = current
			}
			sameDisjunctionCount++
		}
	}
	if sameDisjunctionCount > 0 {
		return firstSameDisjunction, sameDisjunctionCount, true
	}
	return first, problemCount, true
}

func reportBackref(ctx rule.RuleContext, reportRange core.TextRange, bref *rxNode, first problem, problemCount int) {
	otherGroups := ""
	otherCount := problemCount - 1
	switch {
	case otherCount == 1:
		otherGroups = " and another group"
	case otherCount > 1:
		otherGroups = fmt.Sprintf(" and other %d groups", otherCount)
	}

	desc := descriptionFor(first.messageId, bref.raw, first.group.raw, otherGroups)
	ctx.ReportRange(reportRange, rule.RuleMessage{
		Id:          first.messageId,
		Description: desc,
	})
}

func classifyPair(bref *rxNode, group *rxNode) (problem, bool) {
	for current := bref; current != nil; current = current.parent {
		if current != group {
			continue
		}
		// Group is bref's ancestor → bref is nested within the group, which
		// hasn't matched yet when bref starts to match.
		return problem{messageId: "nested", group: group}, true
	}

	brefDepth := rxNodeDepth(bref)
	groupDepth := rxNodeDepth(group)
	brefAncestor := bref
	groupAncestor := group
	for brefDepth > groupDepth {
		brefAncestor = brefAncestor.parent
		brefDepth--
	}
	for groupDepth > brefDepth {
		groupAncestor = groupAncestor.parent
		groupDepth--
	}
	for brefAncestor != groupAncestor {
		brefAncestor = brefAncestor.parent
		groupAncestor = groupAncestor.parent
	}
	lowestCommonAncestor := groupAncestor

	var lowestCommonLookaround *rxNode
	for current := lowestCommonAncestor; current != nil; current = current.parent {
		if isLookaround(current) {
			lowestCommonLookaround = current
			break
		}
	}
	matchingBackward := lowestCommonLookaround != nil && isLookbehind(lowestCommonLookaround)

	groupBranch := group
	for groupBranch.parent != lowestCommonAncestor {
		groupBranch = groupBranch.parent
	}
	if groupBranch.kind == nkAlternative {
		return problem{messageId: "disjunctive", group: group}, true
	}
	if !matchingBackward && bref.end <= group.start {
		return problem{messageId: "forward", group: group}, true
	}
	if matchingBackward && group.end <= bref.start {
		return problem{messageId: "backward", group: group}, true
	}
	for current := group; current != lowestCommonAncestor; current = current.parent {
		if isNegativeLookaround(current) {
			return problem{messageId: "intoNegativeLookaround", group: group}, true
		}
	}
	return problem{}, false
}

func rxNodeDepth(node *rxNode) int {
	depth := 0
	for current := node; current != nil; current = current.parent {
		depth++
	}
	return depth
}

func descriptionFor(messageId, bref, group, otherGroups string) string {
	switch messageId {
	case "nested":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s from within that group.", bref, group, otherGroups)
	case "forward":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s which appears later in the pattern.", bref, group, otherGroups)
	case "backward":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s which appears before in the same lookbehind.", bref, group, otherGroups)
	case "disjunctive":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s which is in another alternative.", bref, group, otherGroups)
	case "intoNegativeLookaround":
		return fmt.Sprintf("Backreference '%s' will be ignored. It references group '%s'%s which is in a negative lookaround.", bref, group, otherGroups)
	}
	return ""
}
