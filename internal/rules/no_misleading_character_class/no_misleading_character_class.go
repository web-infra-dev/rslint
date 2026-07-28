// cspell:ignore FFFD Dedup

// Package no_misleading_character_class implements ESLint's
// no-misleading-character-class rule on top of the layered regex / JS string
// utilities in `internal/utils`.
//
// Architecture:
//   - utils.IterateRegexCharacterClasses + utils.ParseRegexCharacterClass
//     produce the per-class element list with byte-level source positions.
//   - utils.ParseJSStringLiteralSource / utils.ParseJSTemplateLiteralSource
//     turn a `RegExp(str, flags)` argument into a code-unit stream that the
//     same regex character-class parser can run over.
//   - This file is the rule's "detector wiring" only: it maps elements into
//     character sequences and runs the six pattern checks.
package no_misleading_character_class

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-misleading-character-class
var NoMisleadingCharacterClassRule = rule.Rule{
	Name: "no-misleading-character-class",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)
		var callTracker *regexpCallTracker
		if sourceMayUseRegExp(ctx.SourceFile) {
			callTracker = newRegExpCallTracker(ctx)
		}
		listeners := rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				handleRegexLiteral(ctx, node, opts, callTracker)
			},
		}
		if callTracker == nil {
			return listeners
		}
		eval := utils.NewStaticStringEvaluatorWithSourceFile(ctx.TypeChecker, ctx.SourceFile)
		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			call := node.AsCallExpression()
			if !callTracker.isRegExpCall(node, call.Expression) {
				return
			}
			handleRegExpConstructor(ctx, node, call.Arguments, opts, eval)
		}
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			if !callTracker.isRegExpCall(node, newExpr.Expression) {
				return
			}
			handleRegExpConstructor(ctx, node, newExpr.Arguments, opts, eval)
		}
		return listeners
	},
}

var regexpGlobalObjectNames = [...]string{
	"global",
	"globalThis",
	"self",
	"window",
}

func sourceMayUseRegExp(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	if _, ok := sourceFile.Identifiers["RegExp"]; ok {
		return true
	}
	for _, name := range regexpGlobalObjectNames {
		if _, ok := sourceFile.Identifiers[name]; ok {
			return true
		}
	}
	return false
}

type ruleOptions struct {
	allowEscape bool
}

func parseOptions(opts any) ruleOptions {
	res := ruleOptions{}
	m := utils.GetOptionsMap(opts)
	if m != nil {
		if v, ok := m["allowEscape"].(bool); ok {
			res.allowEscape = v
		}
	}
	return res
}

// ---------------------------------------------------------------------------
// One regex character as the detectors see it.
// ---------------------------------------------------------------------------

// regexChar is a single UTF-16 code unit (or combined astral code point under
// u/v) inside a character class, with its absolute source byte range.
type regexChar struct {
	value    uint32
	srcStart int
	srcEnd   int
	isUBrace bool
	raw      string // raw source text for this char, used by allowEscape; empty when unavailable
}

// foundMatch is a single detector hit.
type foundMatch struct {
	kind     string
	srcStart int
	srcEnd   int
}

// ---------------------------------------------------------------------------
// Listeners
// ---------------------------------------------------------------------------

func handleRegexLiteral(ctx rule.RuleContext, node *ast.Node, opts ruleOptions, callTracker *regexpCallTracker) {
	// Skip the regex literal if its immediate context is a RegExp(...)/new RegExp(...)
	// call with an explicit flags argument — in that case, the constructor
	// listener will handle it (using the override flags).
	if isRegexLiteralHandledByConstructor(node, callTracker) {
		return
	}
	text := node.Text()
	pattern, flags := utils.ExtractRegexPatternAndFlags(text)
	if pattern == "" && flags == "" {
		return
	}
	rxFlags := utils.ParseRegexFlags(flags)

	trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
	patternStart := trimmed.Pos() + 1 // skip leading "/"

	matches := scanPatternForMatches(pattern, rxFlags, opts, patternStart)
	if len(matches) == 0 {
		return
	}

	for _, m := range matches {
		emitMatch(ctx, m, pattern, func() []rule.RuleFix {
			return []rule.RuleFix{rule.RuleFixInsertAfter(node, "u")}
		})
	}
}

func handleRegExpConstructor(ctx rule.RuleContext, callNode *ast.Node, args *ast.NodeList, opts ruleOptions, eval *utils.StaticStringEvaluator) {
	if args == nil || len(args.Nodes) == 0 {
		return
	}

	patternNode := ast.SkipParentheses(args.Nodes[0])
	if patternNode == nil {
		return
	}

	// Resolve an identifier first-arg to its const initializer literal when a
	// TypeChecker is available. Files without type info (JS-only, no project)
	// have TypeChecker == nil; in that case we intentionally skip the
	// resolution path (same policy as other rules that rely on type info).
	// Note: identifiers resolving to *regex* literals are NOT followed — this
	// matches ESLint's `getStaticValueOrRegex`, which returns null for RegExp
	// objects so that flag-stripping patterns (e.g. `const r = /x/u;
	// new RegExp(r, "")`) aren't re-analyzed under override flags.
	if patternNode.Kind == ast.KindIdentifier && ctx.TypeChecker != nil {
		if resolved := resolveBindingInitializer(patternNode, eval); resolved != nil {
			patternNode = resolved
		}
	}

	// Pattern arg = regex literal with no flags arg → handled by the literal listener.
	if patternNode.Kind == ast.KindRegularExpressionLiteral && len(args.Nodes) == 1 {
		return
	}

	// Determine flags first — they affect how patterns are parsed (u/v collapse).
	flags, hasFlags, flagsKnown := readFlagsArg(args, eval)
	if !flagsKnown {
		return
	}

	rxFlags := utils.ParseRegexFlags(flags)

	var matches []foundMatch
	var patternForUFlagCheck string
	switch patternNode.Kind {
	case ast.KindStringLiteral:
		patternForUFlagCheck = patternNode.AsStringLiteral().Text
		matches = scanStringLiteralForMatches(ctx, patternNode, rxFlags, opts)
	case ast.KindNoSubstitutionTemplateLiteral:
		patternForUFlagCheck = patternNode.AsNoSubstitutionTemplateLiteral().Text
		matches = scanTemplateLiteralForMatches(ctx, patternNode, rxFlags, opts)
	case ast.KindRegularExpressionLiteral:
		patternForUFlagCheck, _ = utils.ExtractRegexPatternAndFlags(patternNode.Text())
		trimmed := utils.TrimNodeTextRange(ctx.SourceFile, patternNode)
		matches = scanPatternForMatches(patternForUFlagCheck, rxFlags, opts, trimmed.Pos()+1)
	default:
		// Fallback: try to evaluate to a static string. This covers
		// TemplateExpression with all-constant spans, BinaryExpression `+`
		// concatenation, `String.raw\`...\`` tagged templates, and non-const
		// identifiers with no reassignments. Positions are reported at the
		// whole pattern-node loc (matching ESLint's "no granular reports on
		// templates with expressions" / "no granular reports on identifiers"
		// behavior).
		value, ok := eval.Eval(patternNode)
		if !ok {
			return
		}
		patternForUFlagCheck = value
		matches = scanStringValueForMatches(ctx, patternNode, value, rxFlags, opts)
	}

	for _, m := range matches {
		emitMatch(ctx, m, patternForUFlagCheck, func() []rule.RuleFix {
			return buildAddUFlagFixesForCall(ctx.SourceFile, callNode, args, hasFlags)
		})
	}
}

// scanStringValueForMatches runs detection on a resolved string that has no
// per-codeUnit source mapping (because it came from a static evaluator, not a
// literal). Matches are collapsed to at most one diagnostic per kind at the
// full `reportNode` range, matching ESLint's whole-node fallback for the
// identifier / substituted-template paths.
func scanStringValueForMatches(ctx rule.RuleContext, reportNode *ast.Node, value string, flags utils.RegexFlags, opts ruleOptions) []foundMatch {
	inner := scanPatternForMatches(value, flags, opts, 0)
	if len(inner) == 0 {
		return nil
	}
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, reportNode)
	seen := map[string]bool{}
	out := make([]foundMatch, 0, len(inner))
	for _, m := range inner {
		if seen[m.kind] {
			continue
		}
		seen[m.kind] = true
		out = append(out, foundMatch{kind: m.kind, srcStart: nodeRange.Pos(), srcEnd: nodeRange.End()})
	}
	return out
}

// isRegexLiteralHandledByConstructor reports whether `node` (a regex literal)
// is the first argument of a RegExp()/new RegExp() call that supplies an
// explicit flags argument. When true, the constructor listener owns the
// pattern (running under the override flags) and this literal listener
// defers so we don't double-report. This mirrors ESLint's `checkedPatternNodes`
// which routes inline regex-literal args through the Program handler so the
// flag-string-level autofix (inserting `u` into the flags arg) can apply.
func isRegexLiteralHandledByConstructor(node *ast.Node, callTracker *regexpCallTracker) bool {
	if callTracker == nil {
		return false
	}
	parent := node.Parent
	// Walk through parenthesized expressions upward.
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
	if !callTracker.isRegExpCall(parent, callee) {
		return false
	}
	if args == nil || len(args.Nodes) < 2 {
		return false
	}
	// Confirm node is the first argument (possibly through parens). An
	// explicit flags argument overrides the literal's own flags even when it
	// is dynamic, so the literal listener must not report against stale flags.
	if first := ast.SkipParentheses(args.Nodes[0]); first != node {
		return false
	}
	return true
}

// resolveBindingInitializer attempts to resolve an identifier to its
// initializer literal when the initializer is a string or no-substitution
// template literal AND the binding is safely constant — either `const`, or
// `let`/`var` with no write references elsewhere in the file.
//
// Returns nil if no such binding exists. Requires ctx.TypeChecker != nil.
// For let/var resolution, also requires `eval` so we can consult its
// write-reference cache.
func resolveBindingInitializer(ident *ast.Node, eval *utils.StaticStringEvaluator) *ast.Node {
	if eval == nil {
		return nil
	}
	initializer, ok := eval.ResolveIdentifierInitializer(ident)
	if !ok {
		return nil
	}
	init := utils.SkipAssertionsAndParens(initializer)
	if init == nil {
		return nil
	}
	// Note: we deliberately do NOT resolve identifiers whose initializer is a
	// regex literal. ESLint's `getStaticValueOrRegex` (from eslint-utils)
	// explicitly returns null when `getStaticValue` yields a RegExp object —
	// so `const r = /x/u; new RegExp(r, "")` is NOT analyzed under the
	// override flags by ESLint. Resolving here would cause us to over-report
	// on flag-stripping patterns (e.g., the regex literal may be safe under
	// its own `u` but "misleading" under `""`). The standalone literal
	// listener still fires on the var-decl initializer, which is enough to
	// flag genuinely broken patterns.
	switch init.Kind {
	case ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return init
	}
	return nil
}

// RegExp constructor calls are found from the global roots instead of asking
// the TypeChecker about every call expression in the file. This is the same
// direction used by ESLint's ReferenceTracker: trace reads of the global
// RegExp value (and the RegExp property of global objects) through aliases,
// destructuring, and value-preserving expressions until they reach a call or
// construct operation.
//
// Besides avoiding the flow-sensitive GetTypeAtLocation hot path, root-driven
// tracing is important for correctness: a shadowed or overwritten RegExp must
// not be treated as the global merely because its spelling matches.
type regexpTraceValue uint8

const (
	regexpTraceConstructor regexpTraceValue = iota
	regexpTraceGlobalObject
)

type regexpTraceVariable struct {
	symbol *ast.Symbol
	value  regexpTraceValue
}

type regexpTraceGlobal struct {
	name  string
	value regexpTraceValue
}

type regexpCallTracker struct {
	ctx rule.RuleContext

	// Root names are indexed eagerly. Other names are collected only when a
	// configured global is used as an assignment alias.
	identifiersByName map[string][]*ast.Node
	indexedNames      map[string]bool
	allNamesIndexed   bool

	propertyEvaluator   *utils.StaticStringEvaluator
	potentiallyShadowed map[string]bool
	disabledRoots       map[string]bool
	tracedVariables     map[regexpTraceVariable]struct{}
	tracedGlobals       map[regexpTraceGlobal]struct{}
	calls               map[*ast.Node]struct{}
}

func newRegExpCallTracker(ctx rule.RuleContext) *regexpCallTracker {
	tracker := &regexpCallTracker{
		ctx:               ctx,
		identifiersByName: make(map[string][]*ast.Node, len(regexpGlobalObjectNames)+1),
		indexedNames:      make(map[string]bool, len(regexpGlobalObjectNames)+1),
	}
	tracker.collectRootIdentifiers()
	tracker.trackGlobalRoot("RegExp", regexpTraceConstructor)
	for _, name := range regexpGlobalObjectNames {
		tracker.trackGlobalRoot(name, regexpTraceGlobalObject)
	}
	return tracker
}

func (tracker *regexpCallTracker) collectRootIdentifiers() {
	if tracker.ctx.SourceFile == nil {
		return
	}
	tracker.indexedNames["RegExp"] = true
	for _, name := range regexpGlobalObjectNames {
		tracker.indexedNames[name] = true
	}

	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			name := node.AsIdentifier().Text
			if !tracker.indexedNames[name] {
				node.ForEachChild(visit)
				return false
			}
			if utils.IsNonReferenceIdentifier(node) {
				if regexpBindingSymbol(node) != nil {
					if tracker.potentiallyShadowed == nil {
						tracker.potentiallyShadowed = make(map[string]bool)
					}
					tracker.potentiallyShadowed[name] = true
				}
				node.ForEachChild(visit)
				return false
			}

			isWrite := utils.IsWriteReference(node)
			if !isWrite && name != "RegExp" && !tracker.globalObjectRootCanReachRegExp(node) {
				node.ForEachChild(visit)
				return false
			}
			value := regexpTraceGlobalObject
			if name == "RegExp" {
				value = regexpTraceConstructor
			}
			if isWrite || !tracker.isDirectRootCall(node, value) {
				tracker.identifiersByName[name] = append(tracker.identifiersByName[name], node)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	tracker.ctx.SourceFile.AsNode().ForEachChild(visit)
}

func (tracker *regexpCallTracker) isDirectRootCall(node *ast.Node, value regexpTraceValue) bool {
	if value == regexpTraceGlobalObject {
		parent := node.Parent
		if parent == nil || parent.Kind != ast.KindPropertyAccessExpression ||
			parent.AsPropertyAccessExpression().Expression != node {
			return false
		}
		name := parent.AsPropertyAccessExpression().Name()
		if name == nil || name.Text() != "RegExp" {
			return false
		}
		node = parent
	}
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression == node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression == node
	}
	return false
}

func (tracker *regexpCallTracker) globalObjectRootCanReachRegExp(node *ast.Node) bool {
	for node.Parent != nil && regexpValuePassesThrough(node, node.Parent) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		name, ok := tracker.accessExpressionStaticName(parent)
		return ok && name == "RegExp"
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary != nil && binary.Right == node &&
			binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == node
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindBindingElement:
		return parent.AsBindingElement().Initializer == node
	case ast.KindShorthandPropertyAssignment:
		return parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer == node
	}
	return false
}

func (tracker *regexpCallTracker) trackGlobalRoot(name string, value regexpTraceValue) {
	if tracker.isGlobalOff(name) {
		tracker.disableRoot(name)
		return
	}
	identifiers := tracker.identifiersByName[name]
	for _, identifier := range identifiers {
		if tracker.isGlobalReference(identifier, name) && utils.IsWriteReference(identifier) {
			// ReferenceTracker drops a global root when that binding is
			// modified anywhere in the file.
			tracker.disableRoot(name)
			return
		}
	}
	for _, identifier := range identifiers {
		if tracker.isGlobalReference(identifier, name) {
			tracker.trackExpression(identifier, value)
		}
	}
}

func (tracker *regexpCallTracker) disableRoot(name string) {
	if tracker.disabledRoots == nil {
		tracker.disabledRoots = make(map[string]bool)
	}
	tracker.disabledRoots[name] = true
}

func (tracker *regexpCallTracker) isGlobalOff(name string) bool {
	declared, ok := tracker.ctx.Globals[name]
	return ok && !declared
}

func (tracker *regexpCallTracker) isGlobalReference(identifier *ast.Node, name string) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier ||
		utils.IsNonReferenceIdentifier(identifier) {
		return false
	}
	if tracker.ctx.Refs != nil {
		// RefStore uses the binder's own scope walk and is authoritative for
		// ordinary value bindings in a normal linter run.
		if tracker.ctx.Refs.Resolve(identifier) != nil {
			return false
		}
		// Namespace-only bindings are outside RefStore's value lookup. Only
		// pay for the broader syntactic scope check when this file actually
		// declared a watched root; doing it for every ordinary `window.*`
		// read would make large files quadratic.
		if !tracker.potentiallyShadowed[name] {
			return true
		}
	}
	// Keep the syntactic fallback for namespace-only bindings and direct/unit
	// callers without a Program.
	return !utils.IsShadowed(identifier, name)
}

func (tracker *regexpCallTracker) isRegExpCall(node *ast.Node, callee *ast.Node) bool {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return false
	}

	var root *ast.Node
	var rootName string
	switch callee.Kind {
	case ast.KindIdentifier:
		if callee.AsIdentifier().Text != "RegExp" {
			break
		}
		root = callee
		rootName = "RegExp"
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access.Name() == nil || access.Name().Text() != "RegExp" ||
			access.Expression == nil || access.Expression.Kind != ast.KindIdentifier {
			break
		}
		name := access.Expression.AsIdentifier().Text
		for _, globalObjectName := range regexpGlobalObjectNames {
			if name == globalObjectName {
				root = access.Expression
				rootName = name
				break
			}
		}
	}
	if root != nil && !tracker.disabledRoots[rootName] {
		if !tracker.potentiallyShadowed[rootName] || tracker.isGlobalReference(root, rootName) {
			return true
		}
	}
	_, ok := tracker.calls[node]
	return ok
}

func (tracker *regexpCallTracker) trackExpression(node *ast.Node, value regexpTraceValue) {
	if node == nil {
		return
	}

	for node.Parent != nil && regexpValuePassesThrough(node, node.Parent) {
		node = node.Parent
	}

	parent := node.Parent
	if parent == nil {
		return
	}

	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		if value != regexpTraceGlobalObject {
			return
		}
		name, ok := tracker.accessExpressionStaticName(parent)
		if ok && name == "RegExp" {
			tracker.trackExpression(parent, regexpTraceConstructor)
		}
		return
	}

	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && value == regexpTraceConstructor {
			tracker.addCall(parent)
		}

	case ast.KindNewExpression:
		if parent.AsNewExpression().Expression == node && value == regexpTraceConstructor {
			tracker.addCall(parent)
		}

	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.Right == node &&
			binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			tracker.trackAssignmentTarget(binary.Left, value)
			// An assignment expression evaluates to its assigned value.
			tracker.trackExpression(parent, value)
		}

	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		if declaration != nil && declaration.Initializer == node {
			tracker.trackAssignmentTarget(declaration.Name(), value)
		}

	case ast.KindParameter:
		parameter := parent.AsParameterDeclaration()
		if parameter != nil && parameter.Initializer == node {
			tracker.trackAssignmentTarget(parameter.Name(), value)
		}

	case ast.KindBindingElement:
		element := parent.AsBindingElement()
		if element != nil && element.Initializer == node {
			tracker.trackAssignmentTarget(element.Name(), value)
		}

	case ast.KindShorthandPropertyAssignment:
		property := parent.AsShorthandPropertyAssignment()
		if property != nil && property.ObjectAssignmentInitializer == node {
			tracker.trackAssignmentTarget(property.Name(), value)
		}
	}
}

func (tracker *regexpCallTracker) addCall(node *ast.Node) {
	if tracker.calls == nil {
		tracker.calls = make(map[*ast.Node]struct{})
	}
	tracker.calls[node] = struct{}{}
}

func (tracker *regexpCallTracker) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return "", false
		}
		if tracker.propertyEvaluator == nil {
			// ReferenceTracker evaluates property syntax without a scope:
			// literals can fold, but identifier bindings cannot.
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithSourceFile(nil, tracker.ctx.SourceFile)
		}
		return tracker.propertyEvaluator.Eval(access.ArgumentExpression)
	}
	return utils.AccessExpressionStaticName(node)
}

func (tracker *regexpCallTracker) staticPropertyName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindComputedPropertyName {
		computed := node.AsComputedPropertyName()
		if computed == nil || computed.Expression == nil {
			return "", false
		}
		if tracker.propertyEvaluator == nil {
			tracker.propertyEvaluator = utils.NewStaticStringEvaluatorWithSourceFile(nil, tracker.ctx.SourceFile)
		}
		return tracker.propertyEvaluator.Eval(computed.Expression)
	}
	return utils.GetStaticPropertyName(node)
}

func regexpValuePassesThrough(node *ast.Node, parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindParenthesizedExpression:
		return parent.AsParenthesizedExpression().Expression == node
	case ast.KindAsExpression:
		return parent.AsAsExpression().Expression == node
	case ast.KindSatisfiesExpression:
		return parent.AsSatisfiesExpression().Expression == node
	case ast.KindTypeAssertionExpression:
		return parent.AsTypeAssertion().Expression == node
	case ast.KindNonNullExpression:
		return parent.AsNonNullExpression().Expression == node
	case ast.KindPartiallyEmittedExpression:
		return parent.Expression() == node
	case ast.KindExpressionWithTypeArguments:
		return parent.AsExpressionWithTypeArguments().Expression == node
	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		return conditional.WhenTrue == node || conditional.WhenFalse == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			return binary.Left == node || binary.Right == node
		case ast.KindCommaToken:
			return binary.Right == node
		}
	}
	return false
}

func (tracker *regexpCallTracker) trackAssignmentTarget(node *ast.Node, value regexpTraceValue) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindIdentifier:
		tracker.trackIdentifierVariable(node, value)

	case ast.KindObjectBindingPattern:
		if value != regexpTraceGlobalObject {
			return
		}
		pattern := node.AsBindingPattern()
		if pattern == nil || pattern.Elements == nil {
			return
		}
		for _, elementNode := range pattern.Elements.Nodes {
			element := elementNode.AsBindingElement()
			if element == nil || element.DotDotDotToken != nil || element.Name() == nil {
				continue
			}
			propertyName := element.PropertyName
			if propertyName == nil {
				propertyName = element.Name()
			}
			name, ok := tracker.staticPropertyName(propertyName)
			if ok && name == "RegExp" {
				tracker.trackAssignmentTarget(element.Name(), regexpTraceConstructor)
			}
		}

	case ast.KindObjectLiteralExpression:
		if value != regexpTraceGlobalObject {
			return
		}
		for _, propertyNode := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch propertyNode.Kind {
			case ast.KindPropertyAssignment:
				property := propertyNode.AsPropertyAssignment()
				name, ok := tracker.staticPropertyName(property.Name())
				if ok && name == "RegExp" {
					tracker.trackAssignmentTarget(property.Initializer, regexpTraceConstructor)
				}
			case ast.KindShorthandPropertyAssignment:
				property := propertyNode.AsShorthandPropertyAssignment()
				name, ok := tracker.staticPropertyName(property.Name())
				if ok && name == "RegExp" {
					tracker.trackAssignmentTarget(property.Name(), regexpTraceConstructor)
				}
			}
		}

	case ast.KindBinaryExpression:
		// Assignment-pattern defaults retain the value on their left side.
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil &&
			binary.OperatorToken.Kind == ast.KindEqualsToken {
			tracker.trackAssignmentTarget(binary.Left, value)
		}
	}
}

func (tracker *regexpCallTracker) trackIdentifierVariable(identifier *ast.Node, value regexpTraceValue) {
	if symbol := regexpBindingSymbol(identifier); symbol != nil {
		tracker.trackVariable(symbol, value)
		return
	}
	if tracker.ctx.Refs != nil {
		if symbol := tracker.ctx.Refs.Resolve(identifier); symbol != nil {
			tracker.trackVariable(symbol, value)
			return
		}
	}

	name := identifier.AsIdentifier().Text
	if declared, ok := tracker.ctx.Globals[name]; ok && declared {
		// The initial index contains only built-in roots. Indexing this
		// configured-global name also records namespace-only declarations
		// before deciding whether the assignment target is truly global.
		tracker.identifiersForName(name)
		if tracker.isGlobalReference(identifier, name) {
			tracker.trackConfiguredGlobal(name, value)
		}
	}
}

func regexpBindingSymbol(identifier *ast.Node) *ast.Symbol {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.Parent == nil {
		return nil
	}
	declaration := identifier.Parent
	if declaration.Name() != identifier {
		return nil
	}
	symbol := declaration.Symbol()
	if symbol == nil {
		return nil
	}
	const flags = ast.SymbolFlagsVariable |
		ast.SymbolFlagsFunction |
		ast.SymbolFlagsClass |
		ast.SymbolFlagsEnum |
		ast.SymbolFlagsModule |
		ast.SymbolFlagsAlias
	if symbol.Flags&flags == 0 {
		return nil
	}
	return symbol
}

func (tracker *regexpCallTracker) trackVariable(symbol *ast.Symbol, value regexpTraceValue) {
	if tracker.ctx.Refs == nil || symbol == nil {
		return
	}
	key := regexpTraceVariable{symbol: symbol, value: value}
	if _, ok := tracker.tracedVariables[key]; ok {
		return
	}
	if tracker.tracedVariables == nil {
		tracker.tracedVariables = make(map[regexpTraceVariable]struct{})
	}
	tracker.tracedVariables[key] = struct{}{}

	for _, reference := range tracker.ctx.Refs.References(symbol) {
		if ast.IsWriteOnlyAccess(reference) {
			continue
		}
		tracker.trackExpression(reference, value)
	}
}

func (tracker *regexpCallTracker) trackConfiguredGlobal(name string, value regexpTraceValue) {
	key := regexpTraceGlobal{name: name, value: value}
	if _, ok := tracker.tracedGlobals[key]; ok {
		return
	}
	if tracker.tracedGlobals == nil {
		tracker.tracedGlobals = make(map[regexpTraceGlobal]struct{})
	}
	tracker.tracedGlobals[key] = struct{}{}

	for _, identifier := range tracker.identifiersForName(name) {
		if ast.IsWriteOnlyAccess(identifier) || !tracker.isGlobalReference(identifier, name) {
			continue
		}
		tracker.trackExpression(identifier, value)
	}
}

func (tracker *regexpCallTracker) identifiersForName(name string) []*ast.Node {
	if tracker.indexedNames[name] || tracker.allNamesIndexed || tracker.ctx.SourceFile == nil {
		return tracker.identifiersByName[name]
	}

	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			identifierName := node.AsIdentifier().Text
			if utils.IsNonReferenceIdentifier(node) {
				if regexpBindingSymbol(node) != nil {
					if tracker.potentiallyShadowed == nil {
						tracker.potentiallyShadowed = make(map[string]bool)
					}
					tracker.potentiallyShadowed[identifierName] = true
				}
			} else if !tracker.indexedNames[identifierName] {
				tracker.identifiersByName[identifierName] = append(tracker.identifiersByName[identifierName], node)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	tracker.ctx.SourceFile.AsNode().ForEachChild(visit)
	tracker.allNamesIndexed = true
	return tracker.identifiersByName[name]
}

func readFlagsArg(args *ast.NodeList, eval *utils.StaticStringEvaluator) (flags string, hasFlags bool, known bool) {
	if len(args.Nodes) < 2 {
		return "", false, true
	}
	node := ast.SkipParentheses(args.Nodes[1])
	if node == nil {
		return "", true, false
	}
	// Any expression that statically evaluates to a string — this covers
	// `"u"`, `` `u` ``, `"u" + ""`, `String.raw\`u\``, and `const F = "u"; ...F`.
	if value, ok := eval.Eval(node); ok {
		return value, true, true
	}
	return "", true, false
}

// ---------------------------------------------------------------------------
// Scanning entry points
// ---------------------------------------------------------------------------

// Sentinel values used in the regexChar stream to mark positions where
// sequences should break. They occupy the out-of-range uint32 space so real
// code units / code points never collide with them.
const (
	sentinelBreaker       uint32 = 0xFFFFFF00 // CharacterSet (\d, \p{}, \q{}, nested class, set op)
	sentinelRangeBoundary uint32 = 0xFFFFFFFF // split point between min/max of a CharacterClassRange
)

// scanPatternForMatches walks `pattern` (as raw regex source, e.g. the body
// of a regex literal between the `/` delimiters), finds each character class,
// parses its elements, and runs the detectors. `srcOffset` is added to each
// element's byte position to translate into absolute source positions.
func scanPatternForMatches(pattern string, flags utils.RegexFlags, opts ruleOptions, srcOffset int) []foundMatch {
	var matches []foundMatch
	utils.IterateRegexCharacterClasses(pattern, flags, func(start, end int) {
		els, _, ok := utils.ParseRegexCharacterClassWithEnd(pattern, start, end, flags)
		if !ok {
			return
		}
		matches = append(matches, runDetectorsOnElements(els, flags, opts, srcOffset, pattern)...)
	})
	return matches
}

// scanStringLiteralForMatches handles `RegExp(stringLiteral, ...)`. The
// string's resolved value is the regex pattern; positions are mapped through
// the per-code-unit table so reports point to the source.
func scanStringLiteralForMatches(ctx rule.RuleContext, node *ast.Node, flags utils.RegexFlags, opts ruleOptions) []foundMatch {
	nodeText := nodeRawText(ctx.SourceFile, node)
	units := utils.ParseJSStringLiteralSource(nodeText)
	if units == nil {
		return nil
	}
	return scanLiteralUnitsForMatches(units, nodeText, flags, opts, utils.TrimNodeTextRange(ctx.SourceFile, node).Pos())
}

func scanTemplateLiteralForMatches(ctx rule.RuleContext, node *ast.Node, flags utils.RegexFlags, opts ruleOptions) []foundMatch {
	nodeText := nodeRawText(ctx.SourceFile, node)
	units := utils.ParseJSTemplateLiteralSource(nodeText)
	if units == nil {
		return nil
	}
	return scanLiteralUnitsForMatches(units, nodeText, flags, opts, utils.TrimNodeTextRange(ctx.SourceFile, node).Pos())
}

// scanLiteralUnitsForMatches turns a code-unit stream into a Go string we can
// hand to the existing pattern scanner, then translates element positions
// (which are byte offsets within the resolved string) back through the unit
// table into absolute source positions.
//
// Important: the resolved string we feed the regex parser is built from the
// units' values — we encode each unit as UTF-8 (substituting 0xFFFD for lone
// surrogate values, which simplifies byte handling without losing structural
// information since the regex parser only needs `[`, `]`, `\`, etc. to
// recognize structure).
func scanLiteralUnitsForMatches(units []utils.StringCodeUnit, nodeText string, flags utils.RegexFlags, opts ruleOptions, nodeStart int) []foundMatch {
	// Build resolved-bytes buffer + byte→unit index map.
	var sb strings.Builder
	byteToUnit := make([]int, 0, len(units)*2)
	for i, u := range units {
		// Encode using UTF-8; surrogate halves can't be encoded as legal UTF-8
		// so we substitute the replacement character. The regex scanner cares
		// about `[` `]` `\` `^` `-` `&` and digits — none of those are
		// surrogates — so this substitution is safe for structural parsing.
		writeStart := sb.Len()
		switch {
		case u.Value <= 0x7F:
			sb.WriteByte(byte(u.Value))
		case u.Value <= 0xFFFF && (u.Value < 0xD800 || u.Value > 0xDFFF):
			sb.WriteRune(rune(u.Value))
		default:
			// Lone surrogate — use replacement character (3 bytes).
			sb.WriteRune('\uFFFD')
		}
		writeEnd := sb.Len()
		// Record the unit index for each resolved byte produced.
		for b := writeStart; b < writeEnd; b++ {
			_ = b
			byteToUnit = append(byteToUnit, i)
		}
	}
	resolved := sb.String()

	var matches []foundMatch
	utils.IterateRegexCharacterClasses(resolved, flags, func(start, end int) {
		els, _, ok := utils.ParseRegexCharacterClassWithEnd(resolved, start, end, flags)
		if !ok {
			return
		}
		matches = append(matches, runDetectorsOnLiteralElements(els, units, byteToUnit, nodeText, flags, opts, nodeStart)...)
	})
	return matches
}

// ---------------------------------------------------------------------------
// Element → detector pipeline
// ---------------------------------------------------------------------------

// runDetectorsOnElements produces detector matches from char-class elements
// where the elements' Start/End are byte offsets in the regex pattern source
// (with `srcOffset` to apply for absolute position).
//
// Under non-u/v mode, raw astral characters in the pattern are split into
// surrogate-pair regexChars (one element → two units) so the detectors see
// what the JS regex engine sees.
func runDetectorsOnElements(els []utils.RegexCharElement, flags utils.RegexFlags, opts ruleOptions, srcOffset int, pattern string) []foundMatch {
	transformed := make([]regexChar, 0, transformedElementCapacity(els, !flags.UV()))
	for _, e := range els {
		if e.Kind == utils.RegexCharBreaker {
			transformed = append(transformed, regexChar{
				value: sentinelBreaker, srcStart: srcOffset + e.Start, srcEnd: srcOffset + e.End,
			})
			continue
		}
		raw := ""
		if e.Start >= 0 && e.End <= len(pattern) {
			raw = pattern[e.Start:e.End]
		}
		if e.Kind == utils.RegexCharRange {
			transformed = append(transformed,
				makeRegexChar(e.Value, e.IsUBrace, srcOffset+e.Start, srcOffset+e.End, raw),
				rangeBoundaryMarker(),
				makeRegexChar(e.Max, e.MaxIsUBrace, srcOffset+e.Start, srcOffset+e.End, raw),
			)
		} else {
			// Under non-u/v mode, an astral code point (raw or via `\` + astral
			// identity escape) is seen by the JS regex engine as a surrogate
			// pair. Split so detectors see both halves. `\uHHHH\uHHHH` does not
			// reach here as an astral single — layer 2 collapses it only under
			// u/v mode, so under non-u it arrives as two separate BMP elements.
			splitAstral := !flags.UV() && e.Value > 0xFFFF
			if splitAstral {
				hi, lo := splitSurrogatePair(e.Value)
				transformed = append(transformed,
					regexChar{value: hi, srcStart: srcOffset + e.Start, srcEnd: srcOffset + e.End, raw: raw},
					regexChar{value: lo, srcStart: srcOffset + e.Start, srcEnd: srcOffset + e.End, raw: raw},
				)
			} else {
				transformed = append(transformed,
					regexChar{value: e.Value, srcStart: srcOffset + e.Start, srcEnd: srcOffset + e.End, isUBrace: e.IsUBrace, raw: raw},
				)
			}
		}
	}
	return runDetectorsOnChars(transformed, flags, opts)
}

func makeRegexChar(value uint32, isUBrace bool, srcStart, srcEnd int, raw string) regexChar {
	return regexChar{
		value: value, srcStart: srcStart, srcEnd: srcEnd, isUBrace: isUBrace, raw: raw,
	}
}

func splitSurrogatePair(cp uint32) (hi, lo uint32) {
	cp -= 0x10000
	return 0xD800 + (cp >> 10), 0xDC00 + (cp & 0x3FF)
}

// runDetectorsOnLiteralElements is the variant for string-literal patterns.
// The element values come from the resolved-bytes parser, where lone
// surrogates were encoded as U+FFFD; we recover the original surrogate
// values via the units table. We also recover the raw source text for each
// element (so allowEscape can compare it to the cooked value).
func runDetectorsOnLiteralElements(els []utils.RegexCharElement, units []utils.StringCodeUnit, byteToUnit []int, nodeText string, flags utils.RegexFlags, opts ruleOptions, nodeStart int) []foundMatch {
	transformed := make([]regexChar, 0, transformedElementCapacity(els, false))
	for _, e := range els {
		if e.Kind == utils.RegexCharBreaker {
			s, en := litElementSourceSpan(e, units, byteToUnit, nodeStart)
			transformed = append(transformed, regexChar{value: sentinelBreaker, srcStart: s, srcEnd: en})
			continue
		}
		// Each element spans 1+ resolved bytes. If it spans exactly one
		// codeUnit (the typical case for raw chars and lone surrogates), we
		// recover the unit's value directly. If it spans multiple units (an
		// escape parsed from string content like `\x41`, which would only
		// happen if the user double-escaped), use e.Value.
		startUnitIdx, endUnitIdx := elementUnitRange(e, byteToUnit)
		minVal, maxVal := e.Value, e.Max
		if e.Kind == utils.RegexCharSingle && startUnitIdx >= 0 && endUnitIdx >= 0 {
			switch {
			case startUnitIdx == endUnitIdx:
				// Single resolved code unit — its value is the source of truth.
				minVal = units[startUnitIdx].Value
			case isSurrogateValue(units[endUnitIdx].Value):
				// Multi-byte escape whose trailing unit is a lone surrogate
				// — i.e., `\<high>` or `\<low>` identity escape in the
				// resolved string. The parser reported U+FFFD (our placeholder)
				// as the value; the correct one is the unit's original value.
				minVal = units[endUnitIdx].Value
			}
		}
		s, en := litElementSourceSpan(e, units, byteToUnit, nodeStart)
		raw := ""
		if startUnitIdx >= 0 && endUnitIdx >= 0 && startUnitIdx < len(units) && endUnitIdx < len(units) {
			raw = nodeText[units[startUnitIdx].Start:units[endUnitIdx].End]
		}
		if e.Kind == utils.RegexCharRange {
			transformed = append(transformed,
				regexChar{value: minVal, srcStart: s, srcEnd: en, isUBrace: e.IsUBrace, raw: raw},
				rangeBoundaryMarker(),
				regexChar{value: maxVal, srcStart: s, srcEnd: en, isUBrace: e.MaxIsUBrace, raw: raw},
			)
		} else {
			transformed = append(transformed,
				regexChar{value: minVal, srcStart: s, srcEnd: en, isUBrace: e.IsUBrace, raw: raw},
			)
		}
	}
	if flags.UV() {
		transformed = collapseSurrogatePairs(transformed)
	}
	return runDetectorsOnChars(transformed, flags, opts)
}

func transformedElementCapacity(els []utils.RegexCharElement, splitAstral bool) int {
	capacity := len(els)
	for _, element := range els {
		if element.Kind == utils.RegexCharRange {
			capacity += 2 // min, boundary, max replace one input element
		} else if splitAstral && element.Kind == utils.RegexCharSingle && element.Value > 0xFFFF {
			capacity++ // one astral input expands to a surrogate pair
		}
	}
	return capacity
}

// collapseSurrogatePairs combines two consecutive non-`\u{}` surrogate-pair
// regexChars into one astral entry, matching what regexpp does under the
// u/v flag.
func collapseSurrogatePairs(chars []regexChar) []regexChar {
	out := chars[:0]
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		if i+1 < len(chars) {
			n := chars[i+1]
			if c.value >= 0xD800 && c.value <= 0xDBFF &&
				n.value >= 0xDC00 && n.value <= 0xDFFF &&
				!c.isUBrace && !n.isUBrace {
				cp := 0x10000 + (c.value-0xD800)*0x400 + (n.value - 0xDC00)
				out = append(out, regexChar{
					value: cp, srcStart: c.srcStart, srcEnd: n.srcEnd,
					raw: c.raw + n.raw,
				})
				i++
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func litElementSourceSpan(e utils.RegexCharElement, units []utils.StringCodeUnit, byteToUnit []int, nodeStart int) (int, int) {
	if e.Start >= len(byteToUnit) || e.End-1 >= len(byteToUnit) || e.End-1 < 0 {
		return nodeStart, nodeStart
	}
	return nodeStart + units[byteToUnit[e.Start]].Start, nodeStart + units[byteToUnit[e.End-1]].End
}

// elementUnitRange returns the [startUnit, endUnit] inclusive index pair the
// element spans within the codeUnits table. Returns (-1, -1) on out-of-range.
func elementUnitRange(e utils.RegexCharElement, byteToUnit []int) (int, int) {
	if e.Start < 0 || e.End <= 0 || e.End-1 >= len(byteToUnit) {
		return -1, -1
	}
	return byteToUnit[e.Start], byteToUnit[e.End-1]
}

// rangeBoundaryMarker returns a sentinel regexChar that splits sequences but
// otherwise is invisible to detectors.
func rangeBoundaryMarker() regexChar {
	return regexChar{value: sentinelRangeBoundary}
}

// runDetectorsOnChars splits a flat character stream at breaker and
// range-boundary sentinels without materializing per-sequence slices or one
// heap pointer per character.
func runDetectorsOnChars(chars []regexChar, flags utils.RegexFlags, opts ruleOptions) []foundMatch {
	var matches []foundMatch
	sequenceStart := 0
	for i := 0; i <= len(chars); i++ {
		atEnd := i == len(chars)
		if !atEnd {
			switch chars[i].value {
			case sentinelBreaker, sentinelRangeBoundary:
			default:
				continue
			}
		}
		if sequenceStart < i {
			matches = appendDetectorMatches(matches, chars[sequenceStart:i], flags, opts.allowEscape)
		}
		sequenceStart = i + 1
	}
	return matches
}

// ---------------------------------------------------------------------------
// Detectors
// ---------------------------------------------------------------------------

func appendDetectorMatches(matches []foundMatch, chars []regexChar, flags utils.RegexFlags, allowEscape bool) []foundMatch {
	uvMode := flags.UV()

	// 1. surrogatePairWithoutUFlag (non-uv only)
	if !uvMode {
		for i := 1; i < len(chars); i++ {
			prev, cur := &chars[i-1], &chars[i]
			if !detectorCharActive(prev, allowEscape) || !detectorCharActive(cur, allowEscape) {
				continue
			}
			if isSurrogatePair(prev.value, cur.value) && !prev.isUBrace && !cur.isUBrace {
				matches = append(matches, foundMatch{kind: "surrogatePairWithoutUFlag", srcStart: prev.srcStart, srcEnd: cur.srcEnd})
			}
		}
	}
	// 2. surrogatePair (uv only, with at least one \u{...})
	if uvMode {
		for i := 1; i < len(chars); i++ {
			prev, cur := &chars[i-1], &chars[i]
			if !detectorCharActive(prev, allowEscape) || !detectorCharActive(cur, allowEscape) {
				continue
			}
			if isSurrogatePair(prev.value, cur.value) && (prev.isUBrace || cur.isUBrace) {
				matches = append(matches, foundMatch{kind: "surrogatePair", srcStart: prev.srcStart, srcEnd: cur.srcEnd})
			}
		}
	}
	// 3. combiningClass — combining char preceded by a non-combining char.
	//    Use unfiltered for the previous-char check (allowEscape semantics).
	for i := 1; i < len(chars); i++ {
		prev, cur := &chars[i-1], &chars[i]
		if !detectorCharActive(cur, allowEscape) {
			continue
		}
		if isCombiningCharacter(cur.value) && !isCombiningCharacter(prev.value) {
			matches = append(matches, foundMatch{kind: "combiningClass", srcStart: prev.srcStart, srcEnd: cur.srcEnd})
		}
	}
	// 4. emojiModifier
	for i := 1; i < len(chars); i++ {
		prev, cur := &chars[i-1], &chars[i]
		if !detectorCharActive(prev, allowEscape) || !detectorCharActive(cur, allowEscape) {
			continue
		}
		if isEmojiModifier(cur.value) && !isEmojiModifier(prev.value) {
			matches = append(matches, foundMatch{kind: "emojiModifier", srcStart: prev.srcStart, srcEnd: cur.srcEnd})
		}
	}
	// 5. regionalIndicatorSymbol — both adjacent chars are RIS.
	for i := 1; i < len(chars); i++ {
		prev, cur := &chars[i-1], &chars[i]
		if !detectorCharActive(prev, allowEscape) || !detectorCharActive(cur, allowEscape) {
			continue
		}
		if isRegionalIndicatorSymbol(cur.value) && isRegionalIndicatorSymbol(prev.value) {
			matches = append(matches, foundMatch{kind: "regionalIndicatorSymbol", srcStart: prev.srcStart, srcEnd: cur.srcEnd})
		}
	}
	// 6. zwj — character sequence joined by U+200D, possibly chained.
	seqStart, seqEnd := -1, -1
	for i := 1; i < len(chars)-1; i++ {
		prev, cur, next := &chars[i-1], &chars[i], &chars[i+1]
		if !detectorCharActive(prev, allowEscape) ||
			!detectorCharActive(cur, allowEscape) ||
			!detectorCharActive(next, allowEscape) {
			continue
		}
		if cur.value == 0x200D && prev.value != 0x200D && next.value != 0x200D {
			if seqStart >= 0 && seqEnd == prev.srcEnd {
				seqEnd = next.srcEnd
			} else {
				if seqStart >= 0 {
					matches = append(matches, foundMatch{kind: "zwj", srcStart: seqStart, srcEnd: seqEnd})
				}
				seqStart = prev.srcStart
				seqEnd = next.srcEnd
			}
		}
	}
	if seqStart >= 0 {
		matches = append(matches, foundMatch{kind: "zwj", srcStart: seqStart, srcEnd: seqEnd})
	}
	return matches
}

func detectorCharActive(char *regexChar, allowEscape bool) bool {
	return !allowEscape || !isAcceptableEscape(char)
}

// ---------------------------------------------------------------------------
// Predicates
// ---------------------------------------------------------------------------

func isSurrogatePair(hi, lo uint32) bool {
	return hi >= 0xD800 && hi <= 0xDBFF && lo >= 0xDC00 && lo <= 0xDFFF
}
func isSurrogateValue(v uint32) bool { return v >= 0xD800 && v <= 0xDFFF }
func isCombiningCharacter(cp uint32) bool {
	return cp <= 0x10FFFF && unicode.Is(unicode.M, rune(cp))
}
func isEmojiModifier(cp uint32) bool { return cp >= 0x1F3FB && cp <= 0x1F3FF }
func isRegionalIndicatorSymbol(cp uint32) bool {
	return cp >= 0x1F1E6 && cp <= 0x1F1FF
}

// isAcceptableEscape mirrors ESLint's checkForAcceptableEscape: the source
// form starts with `\` and ends with a character whose semantic UTF-16 code
// unit does NOT match the resolved value. (This is how `allowEscape` filters
// out the "escaped" half of a pair.)
//
// Two subtleties on top of a straight last-rune comparison:
//
//   - When the char's value is a lone surrogate, ESLint works in JS UTF-16 and
//     its `.source` already reflects a single UTF-16 code unit (either the
//     lone surrogate for identity-escape-of-raw-astral, or the hex sequence
//     for `\uHHHH`). An identity-escape source form (`\` + one rune) means
//     the "last code unit" IS the surrogate itself, i.e., the element is
//     NOT acceptable. Hex-escape source forms (`\uHHHH`, `\u{H}`) end in a
//     hex digit or `}`, which is never equal to a surrogate value, so they
//     remain acceptable.
func isAcceptableEscape(c *regexChar) bool {
	if c == nil || c.raw == "" {
		return false
	}
	if !strings.HasPrefix(c.raw, "\\") {
		return false
	}
	// Special case: lone-surrogate value written as an identity escape
	// (`\<rune>`). The source representation IS that surrogate in JS UTF-16
	// terms, so it's not an alternative form.
	if isSurrogateValue(c.value) && isIdentityEscapeForm(c.raw) {
		return false
	}
	lastRune, _ := utf8.DecodeLastRuneInString(c.raw)
	if lastRune == utf8.RuneError {
		return false
	}
	return uint32(lastRune) != c.value
}

// isIdentityEscapeForm reports whether `raw` is of the form `\<one-rune>`
// (a single backslash followed by exactly one UTF-8-encoded character).
func isIdentityEscapeForm(raw string) bool {
	if len(raw) < 2 || raw[0] != '\\' {
		return false
	}
	rest := raw[1:]
	_, w := utf8.DecodeRuneInString(rest)
	return w == len(rest)
}

// ---------------------------------------------------------------------------
// Reporting
// ---------------------------------------------------------------------------

func emitMatch(ctx rule.RuleContext, m foundMatch, pattern string, makeFixes func() []rule.RuleFix) {
	msg := rule.RuleMessage{Id: m.kind, Description: messageDescriptionFor(m.kind)}
	r := core.NewTextRange(m.srcStart, m.srcEnd)
	if m.kind != "surrogatePairWithoutUFlag" {
		ctx.ReportRange(r, msg)
		return
	}
	ctx.ReportRangeWithDeferredSuggestions(r, msg, func() []rule.RuleSuggestion {
		if !patternValidWithUFlag(pattern) {
			return nil
		}
		fixes := makeFixes()
		if fixes == nil {
			return nil
		}
		return []rule.RuleSuggestion{{
			Message:  rule.RuleMessage{Id: "suggestUnicodeFlag", Description: "Add unicode 'u' flag to regex."},
			FixesArr: fixes,
		}}
	})
}

func messageDescriptionFor(kind string) string {
	switch kind {
	case "surrogatePairWithoutUFlag":
		return "Unexpected surrogate pair in character class. Use 'u' flag."
	case "surrogatePair":
		return "Unexpected surrogate pair in character class."
	case "combiningClass":
		return "Unexpected combined character in character class."
	case "emojiModifier":
		return "Unexpected modified Emoji in character class."
	case "regionalIndicatorSymbol":
		return "Unexpected national flag in character class."
	case "zwj":
		return "Unexpected joined character sequence in character class."
	}
	return ""
}

// ---------------------------------------------------------------------------
// patternValidWithUFlag — heuristic
// ---------------------------------------------------------------------------

// patternValidWithUFlag reports whether adding the `u` flag would keep the
// pattern syntactically valid. Used to decide whether the
// `suggestUnicodeFlag` fix is safe to offer.
//
// We combine two checks — the pattern is valid iff BOTH accept it:
//
//   - regexp2 compile with ECMAScript+Unicode: catches structural errors
//     (unbalanced brackets, bad quantifiers, truncated hex escapes) that
//     would reject the pattern under any engine.
//   - A narrow ES-u-mode identity-escape check: ECMAScript under `u` rejects
//     identity escapes on letters/digits (`\a`, `\z`, `\9`, etc.), which
//     regexp2 accepts because its .NET lineage is laxer there.
//
// This combination catches all the cases in ESLint's test suite that should
// suppress the suggestion without special-casing.
func patternValidWithUFlag(pattern string) bool {
	if _, err := regexp2.Compile(pattern, regexp2.ECMAScript|regexp2.Unicode); err != nil {
		return false
	}
	return !hasInvalidIdentityEscapeForUFlag(pattern)
}

// hasInvalidIdentityEscapeForUFlag reports whether the pattern contains an
// escape that is illegal under the `u` flag.
//
// Under u, the only valid escapes are the recognized ones (\d, \w, \s, \b,
// \B, \n, \t, …), hex forms (\xHH, \uHHHH, \u{H}), control escapes (\cX),
// backreferences (\k<name>, \0..\9), property classes (\p{}, \P{}), and
// identity escapes whose target is a SyntaxCharacter (`^ $ \ . * + ? ( )
// [ ] { } |`) or `/`. Everything else is a syntax error. This function is
// conservative: when in doubt it reports true (i.e. suppresses the fix),
// which preserves safety of the auto-suggested `u` flag.
func hasInvalidIdentityEscapeForUFlag(pattern string) bool {
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		if c != '\\' || i+1 >= len(pattern) {
			i++
			continue
		}
		next := pattern[i+1]
		switch next {
		// Recognized character escapes (single-char).
		case 'b', 'B', 'd', 'D', 'f', 'n', 'r', 's', 'S', 't', 'v', 'w', 'W',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'k', 'p', 'P', 'q':
			i += 2
		// Identity escapes of syntax characters (plus `/`) — the only
		// non-escape-letter identities legal under u.
		case '^', '$', '\\', '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|', '/':
			i += 2
		case 'c':
			i += 3
		case 'x':
			if i+3 < len(pattern) && utils.IsHexDigit(pattern[i+2]) && utils.IsHexDigit(pattern[i+3]) {
				i += 4
			} else {
				return true
			}
		case 'u':
			if i+2 < len(pattern) && pattern[i+2] == '{' {
				closeRel := strings.IndexByte(pattern[i+3:], '}')
				if closeRel < 0 {
					return true
				}
				hex := pattern[i+3 : i+3+closeRel]
				if !utils.AllHexDigits(hex) {
					return true
				}
				i += 3 + closeRel + 1
			} else if i+5 < len(pattern) && utils.AllHexDigits(pattern[i+2:i+6]) {
				i += 6
			} else {
				return true
			}
		default:
			// Any other char after `\` is an invalid identity escape under u:
			// non-syntax ASCII letters (`\a`, `\z`), non-ASCII chars like
			// `\👍`, etc. Conservative: treat as invalid.
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Suggestion fix builders
// ---------------------------------------------------------------------------

func buildAddUFlagFixesForCall(sf *ast.SourceFile, _ *ast.Node, args *ast.NodeList, hasFlags bool) []rule.RuleFix {
	if !hasFlags {
		// Insert `, "u"` immediately after the last argument. This is safe
		// whether or not a trailing comma follows — ES2017+ allows trailing
		// commas in argument lists, so `Fn(a,)` → `Fn(a, "u",)` stays valid,
		// and `Fn(a)` → `Fn(a, "u")` is the canonical non-trailing form.
		lastArg := args.Nodes[len(args.Nodes)-1]
		lastArgEnd := utils.TrimNodeTextRange(sf, lastArg).End()
		return []rule.RuleFix{{
			Range: core.NewTextRange(lastArgEnd, lastArgEnd),
			Text:  `, "u"`,
		}}
	}

	flagsNode := ast.SkipParentheses(args.Nodes[1])
	if flagsNode == nil {
		return nil
	}
	if flagsNode.Kind != ast.KindStringLiteral && flagsNode.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return nil
	}
	flagsRange := utils.TrimNodeTextRange(sf, flagsNode)
	insertAt := flagsRange.End() - 1
	return []rule.RuleFix{{
		Range: core.NewTextRange(insertAt, insertAt),
		Text:  "u",
	}}
}

func nodeRawText(sf *ast.SourceFile, n *ast.Node) string {
	r := utils.TrimNodeTextRange(sf, n)
	return sf.Text()[r.Pos():r.End()]
}
