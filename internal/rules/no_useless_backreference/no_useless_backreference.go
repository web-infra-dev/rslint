package no_useless_backreference

import (
	"fmt"
	"strings"

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
		// Every pattern this rule can report contains a backslash. Constructor
		// patterns folded by StaticStringEvaluator also get that character from
		// a literal in this source file, so files without one need no listeners.
		if ctx.SourceFile != nil && !strings.Contains(ctx.SourceFile.Text(), `\`) {
			return nil
		}

		mayUseRegExp := sourceMayUseRegExp(ctx)
		var calleeCache *regExpCalleeCache
		listeners := rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
				if pattern == "" && flags == "" {
					return
				}
				if !mayContainBackreference(pattern) {
					return
				}
				if mayUseRegExp && isRegexLiteralHandledByConstructor(ctx, node, calleeCache) {
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
	if args == nil || len(args.Nodes) == 0 {
		return false
	}
	if first := ast.SkipParentheses(args.Nodes[0]); first != node {
		return false
	}
	return isBuiltinRegExpCallee(ctx, ast.SkipParentheses(callee), calleeCache)
}

type regExpCalleeCache struct {
	// Cache the pure builtin-type predicate, not identifier symbols: the
	// checker may give one const/import/global symbol different flow-narrowed
	// types at different call sites.
	types       map[*checker.Type]bool
	aliases     map[regExpAliasKey]regExpAliasDependency
	resolving   map[regExpAliasKey]bool
	globalRoots map[string]bool
}

type regExpAliasKind uint8

const (
	regExpAliasConstructor regExpAliasKind = iota
	regExpAliasGlobalObject
)

type regExpAliasKey struct {
	symbol *ast.Symbol
	kind   regExpAliasKind
}

// An alias may be reachable from one or more effective-global roots and also
// have unrelated value sources. ReferenceTracker follows every read of an
// alias reached from a root, even if another assignment writes an unknown
// value to that alias. Keep those facts independently instead of requiring all
// possible origins to be globals.
type regExpAliasDependency struct {
	hasRoot        bool
	rootAvailable  bool
	hasIndependent bool
	incomplete     bool
}

func isBuiltinRegExpCallee(ctx rule.RuleContext, callee *ast.Node, calleeCache *regExpCalleeCache) bool {
	if callee == nil {
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		name := callee.AsIdentifier().Text
		if name == "RegExp" {
			// A direct reference starts from the `RegExp` variable in ESLint's
			// effective global scope. Config/inline `off` removes that binding,
			// while a declaration in this file shadows it even when TypeScript's
			// checker merges the declaration with lib.d.ts.
			if ctx.Refs != nil {
				if ctx.Refs.ResolveInFile(callee) == nil {
					return isRegExpGlobalReference(ctx, callee) &&
						isAvailableRegExpGlobalRoot(ctx, callee, "RegExp", calleeCache)
				}
				// A same-named local can itself be an alias, for example
				// `const { RegExp } = globalThis`. Let the type/provenance path
				// below distinguish that from an ordinary shadowing declaration.
			} else {
				// NameResolver does not surface every TypeScript value declaration
				// here (notably an identifier-named namespace), so retain the
				// syntactic check as the binder-only fallback.
				return isAvailableRegExpGlobalRoot(ctx, callee, "RegExp", calleeCache)
			}
		}
		// Effective-global provenance is authoritative for aliases rooted at
		// RegExp or a global object. This path deliberately does not require
		// TypeScript to know a configured host object such as `window`.
		dependency := classifyRegExpAliasIdentifier(ctx, callee, regExpAliasConstructor, calleeCache)
		if dependency.rootAvailable {
			if !dependency.hasIndependent {
				return true
			}
			// Preserve rslint's flow-sensitive TypeScript precision when the
			// alias also receives unrelated values. If type information is
			// missing or indeterminate, effective-global reachability remains
			// authoritative, matching ReferenceTracker's behavior.
			isConstructor, conclusive := classifyBuiltinRegExpConstructorType(ctx, callee, calleeCache)
			if conclusive {
				return isConstructor
			}
			return true
		}
		if dependency.hasRoot && !dependency.hasIndependent {
			// Every recognized source is an effective-global root that is
			// unavailable for this file. Do not let TypeScript's default libs
			// reintroduce a binding removed by languageOptions.
			return false
		}
		if name == "RegExp" {
			// An unrooted same-named local is an ordinary shadow.
			return false
		}

		// A local/imported/ambient alias whose value is independent of the
		// effective globals is an rslint TypeScript extension. Only the type
		// check can recognize it; without type info, do not guess from its name.
		isConstructor, _ := classifyBuiltinRegExpConstructorType(ctx, callee, calleeCache)
		return isConstructor
	}

	// `globalThis.RegExp` and host equivalents start from the global-object
	// variable, not from the bare `RegExp` variable. Turning `RegExp` off does
	// not remove that object's property; conversely, a host root such as
	// `window` exists only when the effective globals declare it.
	if ast.IsAccessExpression(callee) {
		property, ok := utils.AccessExpressionStaticName(callee)
		if !ok || property != "RegExp" {
			return false
		}
		object := utils.SkipAssertionsAndParens(utils.AccessExpressionObject(callee))
		if object == nil || object.Kind != ast.KindIdentifier {
			return false
		}
		name := object.AsIdentifier().Text
		switch name {
		case "globalThis", "window", "self", "global":
			if isRegExpGlobalReference(ctx, object) &&
				isAvailableRegExpGlobalRoot(ctx, object, name, calleeCache) {
				return true
			}
		}
		dependency := classifyRegExpAliasIdentifier(ctx, object, regExpAliasGlobalObject, calleeCache)
		return dependency.rootAvailable
	}
	return false
}

func classifyBuiltinRegExpConstructorType(
	ctx rule.RuleContext,
	callee *ast.Node,
	cache *regExpCalleeCache,
) (isConstructor bool, conclusive bool) {
	if ctx.TypeChecker == nil || ctx.Program == nil || callee == nil {
		return false, false
	}
	t := ctx.TypeChecker.GetTypeAtLocation(callee)
	if t == nil {
		return false, false
	}
	for _, part := range utils.UnionTypeParts(t) {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) ||
			utils.IsIntrinsicErrorType(part) {
			return false, false
		}
	}
	if result, ok := cache.types[t]; ok {
		return result, true
	}
	result := utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, t, "RegExpConstructor")
	if cache.types == nil {
		cache.types = make(map[*checker.Type]bool)
	}
	cache.types[t] = result
	return result, true
}

func classifyRegExpAliasIdentifier(
	ctx rule.RuleContext,
	identifier *ast.Node,
	kind regExpAliasKind,
	cache *regExpCalleeCache,
) regExpAliasDependency {
	if ctx.Refs == nil || identifier == nil || identifier.Kind != ast.KindIdentifier {
		return regExpAliasDependency{}
	}
	symbol := ctx.Refs.ResolveInFile(identifier)
	if symbol == nil {
		return regExpAliasDependency{}
	}
	return classifyRegExpAliasSymbol(ctx, symbol, kind, cache)
}

func classifyRegExpAliasSymbol(
	ctx rule.RuleContext,
	symbol *ast.Symbol,
	kind regExpAliasKind,
	cache *regExpCalleeCache,
) regExpAliasDependency {
	if symbol == nil {
		return regExpAliasDependency{}
	}
	key := regExpAliasKey{symbol: symbol, kind: kind}
	if dependency, ok := cache.aliases[key]; ok {
		return dependency
	}
	if cache.resolving[key] {
		// A cycle contributes no value source of its own. Mark the result as
		// incomplete so it is not memoized before another entry point has had a
		// chance to reach a real source in the cycle.
		return regExpAliasDependency{incomplete: true}
	}
	if cache.resolving == nil {
		cache.resolving = make(map[regExpAliasKey]bool)
	}
	cache.resolving[key] = true
	defer delete(cache.resolving, key)

	dependencies := make([]regExpAliasDependency, 0, len(symbol.Declarations)+1)
	for _, declaration := range symbol.Declarations {
		dependencies = appendAliasDeclarationDependencies(dependencies, ctx, declaration, kind, cache)
	}
	if ctx.Refs != nil {
		for _, reference := range ctx.Refs.References(symbol) {
			dependency, ok := classifyRegExpWriteDependency(ctx, reference, kind, cache)
			if ok {
				dependencies = append(dependencies, dependency)
			}
		}
	}

	dependency := combineRegExpAliasDependencies(dependencies...)
	if len(dependencies) == 0 {
		// Imports, ambient declarations, and uninitialized locals have no
		// syntactic source to trace. The TypeChecker may still recognize a
		// RegExpConstructor type for non-shadowing aliases.
		dependency.hasIndependent = true
	}
	if !dependency.incomplete {
		if cache.aliases == nil {
			cache.aliases = make(map[regExpAliasKey]regExpAliasDependency)
		}
		cache.aliases[key] = dependency
	}
	return dependency
}

func appendAliasDeclarationDependencies(
	dependencies []regExpAliasDependency,
	ctx rule.RuleContext,
	declaration *ast.Node,
	kind regExpAliasKind,
	cache *regExpCalleeCache,
) []regExpAliasDependency {
	if declaration == nil {
		return dependencies
	}
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		variable := declaration.AsVariableDeclaration()
		if variable != nil && variable.Name() != nil && variable.Name().Kind == ast.KindIdentifier && variable.Initializer != nil {
			return append(dependencies, classifyRegExpAliasExpression(ctx, variable.Initializer, kind, cache))
		}
	case ast.KindParameter:
		parameter := declaration.AsParameterDeclaration()
		if parameter != nil && parameter.Name() != nil && parameter.Name().Kind == ast.KindIdentifier && parameter.Initializer != nil {
			return append(dependencies, classifyRegExpAliasExpression(ctx, parameter.Initializer, kind, cache))
		}
	case ast.KindBindingElement:
		element := declaration.AsBindingElement()
		if element == nil {
			return dependencies
		}
		if kind == regExpAliasConstructor {
			if source := directObjectBindingInitializer(declaration, "RegExp"); source != nil {
				dependencies = append(dependencies, classifyRegExpAliasExpression(ctx, source, regExpAliasGlobalObject, cache))
			}
		}
		if element.Initializer != nil {
			dependencies = append(dependencies, classifyRegExpAliasExpression(ctx, element.Initializer, kind, cache))
		}
	}
	return dependencies
}

func directObjectBindingInitializer(bindingElement *ast.Node, propertyName string) *ast.Node {
	if bindingElement == nil || bindingElement.Parent == nil ||
		bindingElement.Parent.Kind != ast.KindObjectBindingPattern {
		return nil
	}
	element := bindingElement.AsBindingElement()
	if element == nil || element.DotDotDotToken != nil {
		return nil
	}
	property := element.PropertyName
	if property == nil {
		property = element.Name()
	}
	name, ok := utils.GetStaticPropertyName(property)
	if !ok || name != propertyName {
		return nil
	}
	pattern := bindingElement.Parent
	container := pattern.Parent
	if container == nil || container.Name() != pattern {
		return nil
	}
	switch container.Kind {
	case ast.KindVariableDeclaration:
		return container.AsVariableDeclaration().Initializer
	case ast.KindParameter:
		return container.AsParameterDeclaration().Initializer
	}
	return nil
}

func classifyRegExpWriteDependency(
	ctx rule.RuleContext,
	reference *ast.Node,
	kind regExpAliasKind,
	cache *regExpCalleeCache,
) (regExpAliasDependency, bool) {
	if !utils.IsWriteReference(reference) {
		return regExpAliasDependency{}, false
	}
	current := reference
	for current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindTypeAssertionExpression,
			ast.KindNonNullExpression:
			current = parent
			continue
		}
		break
	}
	if current.Parent != nil && current.Parent.Kind == ast.KindBinaryExpression {
		binary := current.Parent.AsBinaryExpression()
		if binary != nil && binary.Left == current && binary.OperatorToken != nil &&
			ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			return classifyRegExpAliasExpression(ctx, binary.Right, kind, cache), true
		}
	}

	if kind != regExpAliasConstructor || reference.Parent == nil {
		return regExpAliasDependency{}, false
	}
	property := reference.Parent
	var propertyNode *ast.Node
	switch property.Kind {
	case ast.KindPropertyAssignment:
		if property.AsPropertyAssignment().Initializer == reference {
			propertyNode = property.AsPropertyAssignment().Name()
		}
	case ast.KindShorthandPropertyAssignment:
		if property.AsShorthandPropertyAssignment().Name() == reference {
			propertyNode = reference
		}
	}
	if propertyNode == nil || property.Parent == nil || property.Parent.Kind != ast.KindObjectLiteralExpression {
		return regExpAliasDependency{}, false
	}
	name, ok := utils.GetStaticPropertyName(propertyNode)
	if !ok || name != "RegExp" {
		return regExpAliasDependency{}, false
	}
	left := property.Parent
	for left.Parent != nil && left.Parent.Kind == ast.KindParenthesizedExpression {
		left = left.Parent
	}
	if left.Parent == nil || left.Parent.Kind != ast.KindBinaryExpression {
		return regExpAliasDependency{}, false
	}
	binary := left.Parent.AsBinaryExpression()
	if binary == nil || binary.Left != left || binary.OperatorToken == nil ||
		binary.OperatorToken.Kind != ast.KindEqualsToken {
		return regExpAliasDependency{}, false
	}
	return classifyRegExpAliasExpression(ctx, binary.Right, regExpAliasGlobalObject, cache), true
}

func classifyRegExpAliasExpression(
	ctx rule.RuleContext,
	expression *ast.Node,
	kind regExpAliasKind,
	cache *regExpCalleeCache,
) regExpAliasDependency {
	expression = utils.SkipAssertionsAndParens(expression)
	if expression == nil {
		return regExpAliasDependency{}
	}
	if expression.Kind == ast.KindIdentifier {
		name := expression.AsIdentifier().Text
		if symbol := ctx.Refs.ResolveInFile(expression); symbol != nil {
			return classifyRegExpAliasSymbol(ctx, symbol, kind, cache)
		}
		if !isRegExpGlobalReference(ctx, expression) {
			return regExpAliasDependency{hasIndependent: true}
		}
		switch kind {
		case regExpAliasConstructor:
			if name == "RegExp" {
				return regExpAliasDependency{hasRoot: true, rootAvailable: isAvailableRegExpGlobalRoot(ctx, expression, name, cache)}
			}
		case regExpAliasGlobalObject:
			switch name {
			case "globalThis", "window", "self", "global":
				return regExpAliasDependency{hasRoot: true, rootAvailable: isAvailableRegExpGlobalRoot(ctx, expression, name, cache)}
			}
		}
		return regExpAliasDependency{hasIndependent: true}
	}

	if kind == regExpAliasConstructor && ast.IsAccessExpression(expression) {
		property, ok := utils.AccessExpressionStaticName(expression)
		if ok && property == "RegExp" {
			return classifyRegExpAliasExpression(ctx, utils.AccessExpressionObject(expression), regExpAliasGlobalObject, cache)
		}
		return regExpAliasDependency{hasIndependent: true}
	}

	switch expression.Kind {
	case ast.KindConditionalExpression:
		conditional := expression.AsConditionalExpression()
		return combineRegExpAliasDependencies(
			classifyRegExpAliasExpression(ctx, conditional.WhenTrue, kind, cache),
			classifyRegExpAliasExpression(ctx, conditional.WhenFalse, kind, cache),
		)
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return regExpAliasDependency{hasIndependent: true}
		}
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			return combineRegExpAliasDependencies(
				classifyRegExpAliasExpression(ctx, binary.Left, kind, cache),
				classifyRegExpAliasExpression(ctx, binary.Right, kind, cache),
			)
		case ast.KindCommaToken, ast.KindEqualsToken:
			return classifyRegExpAliasExpression(ctx, binary.Right, kind, cache)
		}
	}
	return regExpAliasDependency{hasIndependent: true}
}

func combineRegExpAliasDependencies(dependencies ...regExpAliasDependency) regExpAliasDependency {
	var combined regExpAliasDependency
	for _, dependency := range dependencies {
		combined.hasRoot = combined.hasRoot || dependency.hasRoot
		combined.rootAvailable = combined.rootAvailable || dependency.rootAvailable
		combined.hasIndependent = combined.hasIndependent || dependency.hasIndependent
		combined.incomplete = combined.incomplete || dependency.incomplete
	}
	return combined
}

func isAvailableRegExpGlobalRoot(
	ctx rule.RuleContext,
	identifier *ast.Node,
	name string,
	cache *regExpCalleeCache,
) bool {
	if !ctx.Globals.Access(name).IsDeclared() || identifier == nil ||
		identifier.Kind != ast.KindIdentifier {
		return false
	}
	if ctx.Refs != nil && ctx.Refs.ResolveInFile(identifier) != nil {
		return false
	}
	if !isRegExpGlobalReference(ctx, identifier) {
		return false
	}
	if available, ok := cache.globalRoots[name]; ok {
		return available
	}

	// ReferenceTracker drops an entire root when that global binding is
	// written anywhere in the file, including aliases captured before the
	// write. Scan once per root name and cache both the true and false result.
	available := true
	if ctx.SourceFile != nil {
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if !available {
				return true
			}
			if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == name &&
				!utils.IsNonReferenceIdentifier(node) && utils.IsWriteReference(node) {
				if isRegExpGlobalReference(ctx, node) {
					available = false
					return true
				}
			}
			node.ForEachChild(visit)
			return false
		}
		ctx.SourceFile.AsNode().ForEachChild(visit)
	}
	if cache.globalRoots == nil {
		cache.globalRoots = make(map[string]bool)
	}
	cache.globalRoots[name] = available
	return available
}

func isRegExpGlobalReference(ctx rule.RuleContext, identifier *ast.Node) bool {
	if ctx.Refs != nil {
		if symbol := ctx.Refs.Resolve(identifier); symbol != nil {
			return !utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
		}
	}
	return !utils.IsShadowed(identifier, identifier.AsIdentifier().Text)
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
