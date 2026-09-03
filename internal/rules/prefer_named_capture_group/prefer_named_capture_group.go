package prefer_named_capture_group

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/prefer-named-capture-group
var PreferNamedCaptureGroupRule = rule.Rule{
	Name:   "prefer-named-capture-group",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		listeners := rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
				checkRegex(ctx, node, node, pattern, flags, 1)
			},
		}
		if sourceFile := ctx.SourceFile; sourceFile != nil && sourceFile.Identifiers != nil {
			mayUseConstructor := false
			for _, name := range [...]string{"RegExp", "globalThis", "window", "self", "global"} {
				if _, ok := sourceFile.Identifiers[name]; ok {
					mayUseConstructor = true
					break
				}
			}
			if !mayUseConstructor {
				return listeners
			}
		}

		var modifiedGlobals map[string]bool
		getModifiedGlobals := func() map[string]bool {
			if modifiedGlobals == nil {
				modifiedGlobals = collectModifiedGlobalRoots(ctx)
			}
			return modifiedGlobals
		}
		var eval *utils.StaticStringEvaluator
		getEval := func() *utils.StaticStringEvaluator {
			if eval == nil {
				eval = utils.NewStaticStringEvaluatorWithoutScope()
			}
			return eval
		}
		var aliasEval *utils.StaticStringEvaluator
		getAliasEval := func() *utils.StaticStringEvaluator {
			if aliasEval == nil {
				aliasEval = utils.NewStaticStringEvaluatorWithReferenceResolver(
					ctx.TypeChecker,
					ctx.SourceFile,
					ctx.Refs,
				)
			}
			return aliasEval
		}

		checkConstructor := func(node *ast.Node, callee *ast.Node, args *ast.NodeList) {
			if args == nil || len(args.Nodes) == 0 {
				return
			}
			patternNode := utils.SkipAssertionsAndParens(args.Nodes[0])
			if !staticPatternMayContainCapture(patternNode) {
				return
			}
			calleeReferences := globalRegExpCalleeReferences(
				ctx,
				utils.SkipAssertionsAndParens(callee),
				getEval(),
				getAliasEval,
				getModifiedGlobals,
			)
			if calleeReferences == 0 {
				return
			}
			pattern, ok := staticStringValue(getEval(), patternNode)
			if !ok || pattern == "" || !strings.Contains(pattern, "(") {
				return
			}

			flags := ""
			if len(args.Nodes) >= 2 {
				flagsNode := utils.SkipAssertionsAndParens(args.Nodes[1])
				if v, ok := staticStringValue(getEval(), flagsNode); ok {
					flags = v
				}
			}

			patternSourceOffset := 1
			if patternNode.Kind == ast.KindRegularExpressionLiteral {
				patternSourceOffset = 0
			}
			for range calleeReferences {
				checkRegex(ctx, node, patternNode, pattern, flags, patternSourceOffset)
			}
		}

		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			call := node.AsCallExpression()
			checkConstructor(node, call.Expression, call.Arguments)
		}
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			checkConstructor(node, newExpr.Expression, newExpr.Arguments)
		}
		return listeners
	},
}

// staticStringValue folds a `RegExp()` argument to a string the way
// `String(value)` would: a regex literal becomes its own source text —
// delimiting slashes and flags included, so `new RegExp(/(a)/)` is read as the
// pattern `/(a)/` — and everything else goes through the static evaluator.
func staticStringValue(eval *utils.StaticStringEvaluator, node *ast.Node) (string, bool) {
	if node != nil && node.Kind == ast.KindRegularExpressionLiteral {
		return node.Text(), true
	}
	return eval.EvalToString(node)
}

// staticPatternMayContainCapture avoids callee and whole-pattern work when a
// literal argument's cooked value cannot contain a group opener. Expressions
// still go through normal static evaluation because their value is not stored
// directly on the node.
func staticPatternMayContainCapture(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindRegularExpressionLiteral, ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return strings.Contains(node.Text(), "(")
	default:
		return true
	}
}

// checkRegex parses pattern (as it would be read under flags) and reports
// every capturing group that isn't named. reportNode anchors the diagnostic
// range (the literal, or the whole call/new expression for a constructor
// call); patternSourceNode is the AST node the pattern text came from, used
// to decide whether — and where — a fix-safe suggestion can be offered.
func checkRegex(ctx rule.RuleContext, reportNode *ast.Node, patternSourceNode *ast.Node, pattern string, flags string, patternSourceOffset int) {
	if !strings.Contains(pattern, "(") {
		return
	}
	rxFlags := utils.ParseRegexFlags(flags)
	groups, ok := utils.RegexCapturingGroups(pattern, rxFlags)
	if !ok {
		return
	}
	var suggestionPlan *regexSuggestionPlan
	suggestionPlanBuilt := false
	getSuggestionPlan := func() *regexSuggestionPlan {
		if !suggestionPlanBuilt {
			suggestionPlan = buildRegexSuggestionPlan(ctx, patternSourceNode, pattern, rxFlags, groups, patternSourceOffset)
			suggestionPlanBuilt = true
		}
		return suggestionPlan
	}

	for _, group := range groups {
		if group.Name != "" {
			continue
		}
		raw := pattern[group.Start:group.End]
		groupStart := group.Start

		msg := rule.RuleMessage{
			Id:          "required",
			Description: fmt.Sprintf("Capture group '%s' should be converted to a named or non-capturing group.", raw),
			Data:        map[string]string{"group": raw},
		}

		ctx.ReportNodeWithDeferredSuggestions(reportNode, msg, func() []rule.RuleSuggestion {
			return buildSuggestions(getSuggestionPlan(), groupStart)
		})
	}
}

type regexSuggestionPlan struct {
	nodeStart           int
	patternSourceOffset int
	tempName            string
	canAddName          bool
	canMakeNonCapturing bool
}

func buildRegexSuggestionPlan(
	ctx rule.RuleContext,
	patternSourceNode *ast.Node,
	pattern string,
	flags utils.RegexFlags,
	groups []utils.RegexCapturingGroup,
	patternSourceOffset int,
) *regexSuggestionPlan {
	nodeStart, ok := regexPatternSourceStart(ctx, patternSourceNode, pattern, patternSourceOffset)
	if !ok {
		return nil
	}

	firstUnnamedStart := -1
	for _, group := range groups {
		if group.Name == "" {
			firstUnnamedStart = group.Start
			break
		}
	}
	if firstUnnamedStart < 0 {
		return nil
	}

	tempName := nextTempName(pattern, groups)
	// Adding the same fresh name changes neither the capture count nor any
	// existing name. In legacy mode, the resulting activation of named-reference
	// syntax is pattern-global, so the edit's validity is independent of which
	// unnamed group receives it. Likewise, converting any one unnamed group to
	// non-capturing reduces the total capture count by exactly one; numeric
	// reference validity depends on that count, not on which group was changed.
	// Validate one representative of each edit kind here instead of parsing the
	// pattern again for every diagnostic.
	validateLiteralSource := patternSourceOffset == 0 && patternSourceNode.Kind == ast.KindRegularExpressionLiteral
	return &regexSuggestionPlan{
		nodeStart:           nodeStart,
		patternSourceOffset: patternSourceOffset,
		tempName:            tempName,
		canAddName: isValidRegexSuggestion(
			pattern, firstUnnamedStart, "?<"+tempName+">", flags, validateLiteralSource,
		),
		canMakeNonCapturing: isValidRegexSuggestion(
			pattern, firstUnnamedStart, "?:", flags, validateLiteralSource,
		),
	}
}

func isValidRegexSuggestion(
	pattern string,
	groupStart int,
	prefix string,
	flags utils.RegexFlags,
	validateLiteralSource bool,
) bool {
	modified := insertRegexGroupPrefix(pattern, groupStart, prefix)
	if !utils.IsValidRegexPattern(modified, flags) {
		return false
	}
	if !validateLiteralSource {
		return true
	}
	// A regex literal passed to RegExp is visited in two semantic contexts.
	// The constructor check sees the literal's complete source as its static
	// pattern, while the parser still has to accept the edited literal itself.
	// A suggestion must be valid in both contexts.
	literalPattern, literalFlags := utils.ExtractRegexPatternAndFlags(modified)
	return utils.IsValidRegexPattern(literalPattern, utils.ParseRegexFlags(literalFlags))
}

func regexPatternSourceStart(
	ctx rule.RuleContext,
	patternSourceNode *ast.Node,
	pattern string,
	patternSourceOffset int,
) (int, bool) {
	if patternSourceNode == nil {
		return 0, false
	}

	rawText := utils.TrimmedNodeText(ctx.SourceFile, patternSourceNode)
	switch patternSourceNode.Kind {
	case ast.KindRegularExpressionLiteral:
		switch patternSourceOffset {
		case 0:
			if rawText != pattern {
				return 0, false
			}
		case 1:
			literalPattern, _ := utils.ExtractRegexPatternAndFlags(rawText)
			if literalPattern != pattern {
				return 0, false
			}
		default:
			return 0, false
		}
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		if patternSourceOffset != 1 || len(rawText) < 2 || strings.Contains(rawText, "\\") || rawText[1:len(rawText)-1] != pattern {
			return 0, false
		}
	default:
		return 0, false
	}

	return utils.TrimNodeTextRange(ctx.SourceFile, patternSourceNode).Pos(), true
}

func nextTempName(pattern string, groups []utils.RegexCapturingGroup) string {
	nextTemp := 1
	for start := 0; start < len(pattern); {
		rel := strings.Index(pattern[start:], "temp")
		if rel < 0 {
			break
		}
		digitStart := start + rel + len("temp")
		digitEnd := digitStart
		for digitEnd < len(pattern) && pattern[digitEnd] >= '0' && pattern[digitEnd] <= '9' {
			digitEnd++
		}
		if digitEnd > digitStart {
			updateNextTempIndex(pattern[digitStart:digitEnd], &nextTemp)
		}
		start = digitStart
	}
	var occupiedNames map[string]struct{}
	for _, group := range groups {
		if group.Name == "" {
			continue
		}
		name, ok := utils.NormalizeRegexCaptureName(group.Name)
		if !ok {
			continue
		}
		if occupiedNames == nil {
			occupiedNames = make(map[string]struct{})
		}
		occupiedNames[name] = struct{}{}
		if !strings.HasPrefix(name, "temp") || len(name) == len("temp") {
			continue
		}
		digits := name[len("temp"):]
		allDigits := true
		for i := range len(digits) {
			if digits[i] < '0' || digits[i] > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			updateNextTempIndex(digits, &nextTemp)
		}
	}

	candidate := "temp" + strconv.Itoa(nextTemp)
	if _, occupied := occupiedNames[candidate]; !occupied {
		return candidate
	}
	// Atoi cannot advance beyond MaxInt. If that boundary name is already in
	// use, choose a low free suffix instead. There are fewer occupied capture
	// names than positive integers, so this loop terminates after at most the
	// number of named groups plus one iterations.
	for fallback := 1; ; fallback++ {
		candidate = "temp" + strconv.Itoa(fallback)
		if _, occupied := occupiedNames[candidate]; !occupied {
			return candidate
		}
	}
}

func updateNextTempIndex(digits string, nextTemp *int) {
	if n, err := strconv.Atoi(digits); err == nil && n < int(^uint(0)>>1) && n+1 > *nextTemp {
		*nextTemp = n + 1
	}
}

func insertRegexGroupPrefix(pattern string, groupStart int, prefix string) string {
	var result strings.Builder
	result.Grow(len(pattern) + len(prefix))
	result.WriteString(pattern[:groupStart+1])
	result.WriteString(prefix)
	result.WriteString(pattern[groupStart+1:])
	return result.String()
}

// buildSuggestions offers each edit kind that the pattern-scoped plan proved
// syntactically safe. A nil plan means the evaluated pattern did not map 1:1
// onto authored source, as with concatenation or an escaped string literal.
func buildSuggestions(plan *regexSuggestionPlan, groupStart int) []rule.RuleSuggestion {
	if plan == nil {
		return nil
	}
	insertAt := plan.nodeStart + plan.patternSourceOffset + groupStart + 1
	insertRange := core.NewTextRange(insertAt, insertAt)
	suggestions := make([]rule.RuleSuggestion, 0, 2)
	if plan.canAddName {
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{Id: "addGroupName", Description: "Add name to capture group."},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(insertRange, "?<"+plan.tempName+">"),
			},
		})
	}
	if plan.canMakeNonCapturing {
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{Id: "addNonCapture", Description: "Convert group to non-capturing."},
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(insertRange, "?:"),
			},
		})
	}
	return suggestions
}

// globalRegExpCalleeReferences returns how many global RegExp references flow
// to callee through the ReferenceTracker pass-through expressions. A logical
// or conditional expression can contain more than one tracked root, and ESLint
// reports once for each one. A write to a global root disables every use of
// that root, including source-only runs where RefStore cannot resolve lib.d.ts
// globals.

const maxRegExpReferenceCount = 128

func globalRegExpCalleeReferences(
	ctx rule.RuleContext,
	callee *ast.Node,
	eval *utils.StaticStringEvaluator,
	getAliasEval func() *utils.StaticStringEvaluator,
	getModifiedGlobals func() map[string]bool,
) int {
	return globalRegExpCalleeReferencesInner(
		ctx,
		callee,
		eval,
		getAliasEval,
		getModifiedGlobals,
		nil,
	)
}

func globalRegExpCalleeReferencesInner(
	ctx rule.RuleContext,
	callee *ast.Node,
	eval *utils.StaticStringEvaluator,
	getAliasEval func() *utils.StaticStringEvaluator,
	getModifiedGlobals func() map[string]bool,
	activeAliases map[*ast.Node]bool,
) int {
	if callee == nil {
		return 0
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		if callee.AsIdentifier().Text == "RegExp" && isUnmodifiedGlobal(ctx, callee, "RegExp", getModifiedGlobals) {
			return 1
		}
		initializer, ok := getAliasEval().ResolveIdentifierInitializer(callee)
		if !ok {
			return 0
		}
		initializer = utils.SkipAssertionsAndParens(initializer)
		if initializer == nil || activeAliases[initializer] {
			return 0
		}
		if activeAliases == nil {
			activeAliases = make(map[*ast.Node]bool)
		}
		activeAliases[initializer] = true
		defer delete(activeAliases, initializer)
		return globalRegExpCalleeReferencesInner(
			ctx,
			initializer,
			eval,
			getAliasEval,
			getModifiedGlobals,
			activeAliases,
		)
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return 0
		}
		if access.Name().AsIdentifier().Text != "RegExp" {
			return 0
		}
		return knownGlobalObjectReferences(ctx, access.Expression, getAliasEval, getModifiedGlobals, activeAliases)
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return 0
		}
		value, ok := eval.EvalToString(utils.SkipAssertionsAndParens(access.ArgumentExpression))
		if !ok || value != "RegExp" {
			return 0
		}
		return knownGlobalObjectReferences(ctx, access.Expression, getAliasEval, getModifiedGlobals, activeAliases)
	case ast.KindBinaryExpression:
		binary := callee.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return 0
		}
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken:
			return globalRegExpCalleeReferencesInner(ctx, utils.SkipAssertionsAndParens(binary.Right), eval, getAliasEval, getModifiedGlobals, activeAliases)
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			left := globalRegExpCalleeReferencesInner(ctx, utils.SkipAssertionsAndParens(binary.Left), eval, getAliasEval, getModifiedGlobals, activeAliases)
			if left >= maxRegExpReferenceCount {
				return maxRegExpReferenceCount
			}
			return addRegExpReferenceCounts(left, globalRegExpCalleeReferencesInner(ctx, utils.SkipAssertionsAndParens(binary.Right), eval, getAliasEval, getModifiedGlobals, activeAliases))
		default:
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				return globalRegExpCalleeReferencesInner(ctx, utils.SkipAssertionsAndParens(binary.Right), eval, getAliasEval, getModifiedGlobals, activeAliases)
			}
		}
	case ast.KindConditionalExpression:
		conditional := callee.AsConditionalExpression()
		if conditional != nil {
			whenTrue := globalRegExpCalleeReferencesInner(ctx, utils.SkipAssertionsAndParens(conditional.WhenTrue), eval, getAliasEval, getModifiedGlobals, activeAliases)
			if whenTrue >= maxRegExpReferenceCount {
				return maxRegExpReferenceCount
			}
			return addRegExpReferenceCounts(whenTrue, globalRegExpCalleeReferencesInner(ctx, utils.SkipAssertionsAndParens(conditional.WhenFalse), eval, getAliasEval, getModifiedGlobals, activeAliases))
		}
	}

	return 0
}

func knownGlobalObjectReferences(
	ctx rule.RuleContext,
	node *ast.Node,
	getAliasEval func() *utils.StaticStringEvaluator,
	getModifiedGlobals func() map[string]bool,
	activeAliases map[*ast.Node]bool,
) int {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return 0
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			switch binary.OperatorToken.Kind {
			case ast.KindCommaToken:
				return knownGlobalObjectReferences(ctx, binary.Right, getAliasEval, getModifiedGlobals, activeAliases)
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
				left := knownGlobalObjectReferences(ctx, binary.Left, getAliasEval, getModifiedGlobals, activeAliases)
				if left >= maxRegExpReferenceCount {
					return maxRegExpReferenceCount
				}
				return addRegExpReferenceCounts(left, knownGlobalObjectReferences(ctx, binary.Right, getAliasEval, getModifiedGlobals, activeAliases))
			default:
				if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
					return knownGlobalObjectReferences(ctx, binary.Right, getAliasEval, getModifiedGlobals, activeAliases)
				}
			}
		}
	}
	if node.Kind == ast.KindConditionalExpression {
		conditional := node.AsConditionalExpression()
		if conditional != nil {
			whenTrue := knownGlobalObjectReferences(ctx, conditional.WhenTrue, getAliasEval, getModifiedGlobals, activeAliases)
			if whenTrue >= maxRegExpReferenceCount {
				return maxRegExpReferenceCount
			}
			return addRegExpReferenceCounts(whenTrue, knownGlobalObjectReferences(ctx, conditional.WhenFalse, getAliasEval, getModifiedGlobals, activeAliases))
		}
	}
	if node.Kind != ast.KindIdentifier {
		return 0
	}
	name := node.AsIdentifier().Text
	switch name {
	case "globalThis", "window", "self", "global":
		if isUnmodifiedGlobal(ctx, node, name, getModifiedGlobals) {
			return 1
		}
	}
	initializer, ok := getAliasEval().ResolveIdentifierInitializer(node)
	if !ok {
		return 0
	}
	initializer = utils.SkipAssertionsAndParens(initializer)
	if initializer == nil || activeAliases[initializer] {
		return 0
	}
	if activeAliases == nil {
		activeAliases = make(map[*ast.Node]bool)
	}
	activeAliases[initializer] = true
	defer delete(activeAliases, initializer)
	return knownGlobalObjectReferences(ctx, initializer, getAliasEval, getModifiedGlobals, activeAliases)
}

func addRegExpReferenceCounts(left, right int) int {
	if right >= maxRegExpReferenceCount-left {
		return maxRegExpReferenceCount
	}
	return left + right
}

func isUnmodifiedGlobal(ctx rule.RuleContext, node *ast.Node, name string, getModifiedGlobals func() map[string]bool) bool {
	if !ctx.Globals.Access(name).IsDeclared() || getModifiedGlobals()[name] {
		return false
	}
	if ctx.Refs != nil {
		return ctx.Refs.IsGlobalReference(node)
	}
	return !utils.IsShadowed(node, name)
}

func collectModifiedGlobalRoots(ctx rule.RuleContext) map[string]bool {
	modified := make(map[string]bool)
	if ctx.SourceFile == nil {
		return modified
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			name := node.AsIdentifier().Text
			switch name {
			case "RegExp", "globalThis", "window", "self", "global":
				if !utils.IsNonReferenceIdentifier(node) && utils.IsWriteReference(node) {
					isGlobal := !utils.IsShadowed(node, name)
					if ctx.Refs != nil {
						isGlobal = ctx.Refs.IsGlobalReference(node)
					}
					if isGlobal {
						modified[name] = true
					}
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	ctx.SourceFile.AsNode().ForEachChild(visit)
	return modified
}
