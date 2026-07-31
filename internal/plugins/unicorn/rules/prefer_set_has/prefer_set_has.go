// Package prefer_set_has ports eslint-plugin-unicorn's `prefer-set-has` rule.
//
// It flags a `const` array binding that is used only for existence checks via
// `Array#includes()` and recommends declaring it as a `Set` so `Set#has()` can
// be used instead. When the binding has a rewritable TypeScript type annotation
// the fix is applied automatically; otherwise the same edit is offered as a
// manual suggestion.
package prefer_set_has

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDError      = "error"
	messageIDSuggestion = "suggestion"
)

// maxArrayLength mirrors upstream's MAX_ARRAY_LENGTH = 2 ** 32 - 1.
const maxArrayLength = float64(1<<32) - 1

//go:embed prefer_set_has.schema.json
var schemaJSON []byte

// methodsReturnsArray lists methods whose result is always an array (never a
// string), mirroring upstream's `methodsReturnsArray`.
var methodsReturnsArray = map[string]bool{
	// `Array`
	"copyWithin": true,
	"fill":       true,
	"filter":     true,
	"flat":       true,
	"flatMap":    true,
	"map":        true,
	"reverse":    true,
	"sort":       true,
	"splice":     true,
	"toReversed": true,
	"toSorted":   true,
	"toSpliced":  true,
	"with":       true,
	// `String`
	"split": true,
	// `Iterator`
	"toArray": true,
}

// methodsReturnsArrayAndString lists methods that can return an array OR a
// string, mirroring upstream's `methodsReturnsArrayAndString`.
var methodsReturnsArrayAndString = map[string]bool{
	"slice":  true,
	"concat": true,
}

type options struct {
	minimumItems int
}

func parseOptions(rawOptions []any) options {
	opts := options{minimumItems: 0}
	optsMap := utils.GetOptionsMap(rawOptions)
	if optsMap == nil {
		return opts
	}
	if value, ok := optsMap["minimumItems"]; ok {
		if number, ok := utils.CoerceIntegral(value); ok {
			opts.minimumItems = number
		}
	}
	return opts
}

func messageError(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDError,
		Description: fmt.Sprintf("`%s` should be a `Set`, and use `%s.has()` to check existence or non-existence.", name, name),
		Data:        map[string]string{"name": name},
	}
}

func messageSuggestion(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDSuggestion,
		Description: fmt.Sprintf("Switch `%s` to `Set`.", name),
		Data:        map[string]string{"name": name},
	}
}

var PreferSetHasRule = rule.Rule{
	Name:   "unicorn/prefer-set-has",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				checkIdentifier(ctx, opts, node)
			},
		}
	},
}

// checkIdentifier mirrors the body of upstream's `context.on('Identifier', …)`.
func checkIdentifier(ctx rule.RuleContext, opts options, node *ast.Node) {
	// ctx.Refs is the stand-in for ESLint's scope manager; without it we
	// cannot collect references, so degrade gracefully (JS-only runs).
	if ctx.Refs == nil {
		return
	}

	declaration := arrayVariableDeclarator(ctx, node)
	if declaration == nil {
		return
	}
	initializer := declaration.Initializer

	if opts.minimumItems > 0 {
		arraySize, ok := getKnownArraySize(ctx, initializer)
		if !ok || arraySize < opts.minimumItems {
			return
		}
	}

	// The binder attaches the symbol to the VariableDeclaration node (node's
	// parent), which is the currency of ctx.Refs.
	symbol := node.Parent.Symbol()
	if symbol == nil {
		return
	}

	// Upstream collects references via getVariableIdentifiers, which includes
	// every declaration name of the variable. When a `var` is declared more
	// than once (`var foo = ...; var foo = ...`), the other declaration's name
	// is an identifier that classifies as neither includes/length/extra, so
	// getReferenceGroups bails. ctx.Refs.References excludes declaration names,
	// so reproduce that bail explicitly: a variable with more than one
	// declaration is not a single `const` array binding.
	if len(symbol.Declarations) != 1 {
		return
	}

	references := ctx.Refs.References(symbol)
	identifiers := make([]*ast.Node, 0, len(references))
	for _, reference := range references {
		if reference != node {
			identifiers = append(identifiers, reference)
		}
	}

	groups, ok := getReferenceGroups(identifiers)
	if !ok || len(groups.includes) == 0 {
		return
	}

	if len(groups.extra) > 0 && !isKnownUniqueArrayExpression(ctx, initializer) {
		return
	}

	if len(groups.includes) == 1 && !isMultipleCall(groups.includes[0], node) {
		return
	}

	typeText, hasRewritableType := setTypeAnnotationText(ctx, declaration)

	buildFixes := func() []rule.RuleFix {
		return buildSetFixes(ctx, node, declaration, groups, typeText)
	}

	// A type annotation that can't be safely rewritten downgrades the fix to a
	// manual suggestion, matching upstream.
	if declaration.Type != nil && !hasRewritableType {
		ctx.ReportNodeWithDeferredSuggestions(node, messageError(node.AsIdentifier().Text), func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message:  messageSuggestion(node.AsIdentifier().Text),
				FixesArr: buildFixes(),
			}}
		})
		return
	}

	ctx.ReportNodeWithDeferredFixes(node, messageError(node.AsIdentifier().Text), buildFixes)
}

// buildSetFixes assembles the edits shared by the autofix and suggestion paths:
// rewrite the type annotation (when possible), wrap the initializer in
// `new Set(...)`, and rename `.includes`/`.length` to `.has`/`.size`.
func buildSetFixes(
	ctx rule.RuleContext,
	node *ast.Node,
	declaration *ast.VariableDeclaration,
	groups referenceGroups,
	typeAnnotationText string,
) []rule.RuleFix {
	fixes := make([]rule.RuleFix, 0, 3+len(groups.includes)+len(groups.lengths))

	if typeAnnotationText != "" && declaration.Type != nil {
		fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, declaration.Type, typeAnnotationText))
	}

	initializer := declaration.Initializer
	fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, initializer, "new Set("))
	fixes = append(fixes, rule.RuleFixInsertAfter(initializer, ")"))

	for _, identifier := range groups.includes {
		property := propertyOfMemberAccess(ast.WalkUpParenthesizedExpressions(identifier.Parent))
		if property != nil {
			fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, property, "has"))
		}
	}

	for _, identifier := range groups.lengths {
		property := propertyOfMemberAccess(ast.WalkUpParenthesizedExpressions(identifier.Parent))
		if property != nil {
			fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, property, "size"))
		}
	}

	return fixes
}

// propertyOfMemberAccess returns the `.name` node of a non-computed property
// access, or nil.
func propertyOfMemberAccess(node *ast.Node) *ast.Node {
	if node == nil || !ast.IsPropertyAccessExpression(node) {
		return nil
	}
	return node.AsPropertyAccessExpression().Name()
}

// referenceGroups classifies the references to the candidate binding.
type referenceGroups struct {
	includes []*ast.Node
	lengths  []*ast.Node
	extra    []*ast.Node
}

// getReferenceGroups mirrors upstream's `getReferenceGroups`. It returns
// ok=false when any reference cannot be classified (upstream returns
// `undefined`, which makes the caller bail).
func getReferenceGroups(identifiers []*ast.Node) (referenceGroups, bool) {
	var groups referenceGroups
	for _, identifier := range identifiers {
		switch {
		case isIncludesCall(identifier):
			groups.includes = append(groups.includes, identifier)
		case isLengthRead(identifier):
			groups.lengths = append(groups.lengths, identifier)
			groups.extra = append(groups.extra, identifier)
		case isAllowedExtraReference(identifier):
			groups.extra = append(groups.extra, identifier)
		default:
			return referenceGroups{}, false
		}
	}
	return groups, true
}

// isIncludesCall reports whether `identifier` is the object of a
// `identifier.includes(x)` call with exactly one non-spread argument, mirroring
// upstream's `isIncludesCall` (isMethodCall defaults: non-optional, no spread).
func isIncludesCall(node *ast.Node) bool {
	member := ast.WalkUpParenthesizedExpressions(node.Parent)
	if member == nil || !ast.IsPropertyAccessExpression(member) {
		return false
	}
	call := ast.WalkUpParenthesizedExpressions(member.Parent)
	argumentsLength := 1
	match, ok := unicornutil.MatchDotMethodCall(call, unicornutil.DotMethodCallOptions{
		Method:          "includes",
		ArgumentsLength: &argumentsLength,
	})
	if !ok {
		return false
	}
	// isMethodCall defaults to allowSpreadElement:false, so `foo.includes(...x)`
	// is not a match even though it has one argument.
	if ast.IsSpreadElement(match.Call.Arguments()[0]) {
		return false
	}
	// `node.parent.parent.callee.object === node` — the receiver must be the
	// identifier itself (parentheses are transparent in ESTree).
	return ast.SkipParentheses(match.Object) == node
}

// isMultipleCall mirrors upstream's `isMultipleCall`. Walking up from the
// `includes` call towards the declaration's container, it reports whether a
// loop/function boundary is crossed (so the single call may run many times).
//
// Upstream's `root = node.parent.parent.parent` is the block/program
// containing the declaration. In ESTree that is Identifier → VariableDeclarator
// → VariableDeclaration → container; tsgo inserts a VariableDeclarationList and
// VariableStatement between the declaration and its container, so the same
// container sits one hop deeper.
func isMultipleCall(includesIdentifier *ast.Node, node *ast.Node) bool {
	root := ancestorAt(node, 4)

	// Upstream starts the walk at the `.includes()` CallExpression
	// (`identifier.parent.parent`). Parentheses are transparent in ESTree.
	member := ast.WalkUpParenthesizedExpressions(includesIdentifier.Parent)
	parent := ast.WalkUpParenthesizedExpressions(member.Parent)
	for parent != nil && parent != root {
		if isMultipleCallBoundary(parent) {
			return true
		}
		parent = parent.Parent
	}
	return false
}

// isMultipleCallBoundary reports whether crossing `node` means the enclosed
// `includes` call may execute more than once: a loop, or a function-like
// container. The function-like set is shared via utils.IsFunctionLikeContainer,
// which already accounts for tsgo modeling object/class methods, accessors, and
// constructors as their own kinds rather than wrapping a FunctionExpression.
func isMultipleCallBoundary(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindForOfStatement, ast.KindForStatement, ast.KindForInStatement,
		ast.KindWhileStatement, ast.KindDoStatement:
		return true
	}
	return utils.IsFunctionLikeContainer(node)
}

// ancestorAt returns the ancestor reached by following `.Parent` `depth` times.
func ancestorAt(node *ast.Node, depth int) *ast.Node {
	current := node
	for i := 0; i < depth && current != nil; i++ {
		current = current.Parent
	}
	return current
}

// isLengthRead mirrors upstream's `isLengthRead`: a non-computed `.length`
// member read that is neither an assignment target nor a call callee.
// Parentheses are transparent in ESTree, so unwrap them (both around the
// receiver and around the member) to match upstream's flattened view — e.g.
// `(foo).length`.
func isLengthRead(identifier *ast.Node) bool {
	parent := ast.WalkUpParenthesizedExpressions(identifier.Parent)
	if parent == nil || !ast.IsPropertyAccessExpression(parent) {
		return false
	}
	access := parent.AsPropertyAccessExpression()
	if ast.SkipParentheses(access.Expression) != identifier || ast.IsOptionalChainRoot(parent) {
		return false
	}
	property := access.Name()
	if property == nil || !ast.IsIdentifier(property) || property.AsIdentifier().Text != "length" {
		return false
	}

	if isAssignmentTarget(parent) {
		return false
	}

	// `grandParent.type === 'CallExpression' && grandParent.callee === parent`
	grandParent := ast.WalkUpParenthesizedExpressions(parent.Parent)
	if grandParent != nil && ast.IsCallExpression(grandParent) {
		if ast.SkipParentheses(grandParent.AsCallExpression().Expression) == parent {
			return false
		}
	}
	return true
}

// isAssignmentTarget mirrors upstream's `isAssignmentTarget`, skipping TS
// expression wrappers before testing for a write / delete / for-in / for-of
// target.
func isAssignmentTarget(node *ast.Node) bool {
	target := node
	for isTypeScriptExpressionWrapper(target.Parent, target) {
		target = target.Parent
	}
	if utils.IsWriteReference(target) {
		return true
	}
	// Upstream's isLeftHandSide also treats `delete target` as a left-hand
	// side; IsWriteReference does not cover delete.
	if parent := target.Parent; parent != nil && parent.Kind == ast.KindDeleteExpression {
		if parent.AsDeleteExpression().Expression == target {
			return true
		}
	}
	return isForInOrForOfTarget(target)
}

// isTypeScriptExpressionWrapper reports whether `parent` is a TS assertion
// wrapper (`as` / `<T>` / `!` / `satisfies`) whose expression is `child`. It
// keys off ast.OEKAssertions — the same assertion set the static-value
// evaluator unwraps via SkipOuterExpressions — so both directions stay
// consistent (previously `satisfies` was handled by the evaluator but not
// here). isAssignmentTarget walks parents upward, hence the manual child check
// instead of the downward SkipOuterExpressions helper.
func isTypeScriptExpressionWrapper(parent *ast.Node, child *ast.Node) bool {
	if parent == nil || !ast.IsOuterExpression(parent, ast.OEKAssertions) {
		return false
	}
	return parent.Expression() == child
}

// isForInOrForOfTarget reports whether `node` is the left-hand target of a
// for-in / for-of statement.
func isForInOrForOfTarget(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind != ast.KindForInStatement && parent.Kind != ast.KindForOfStatement {
		return false
	}
	return parent.AsForInOrOfStatement().Initializer == node
}

// isAllowedExtraReference mirrors upstream's `isAllowedExtraReference`.
func isAllowedExtraReference(identifier *ast.Node) bool {
	return isIterableUse(identifier) ||
		isArrayOrArgumentSpread(identifier) ||
		isAllowedForEachCall(identifier)
}

// isIterableUse reports whether the identifier is the iterable of a for-of
// statement (`for (const x of identifier)`). Parentheses around the iterable
// (`for (const x of (foo))`) are transparent in ESTree, so unwrap them.
func isIterableUse(identifier *ast.Node) bool {
	parent := ast.WalkUpParenthesizedExpressions(identifier.Parent)
	if parent == nil || parent.Kind != ast.KindForOfStatement {
		return false
	}
	return ast.SkipParentheses(parent.AsForInOrOfStatement().Expression) == identifier
}

// isArrayOrArgumentSpread reports whether the identifier is spread into an
// array literal, call, or new expression (`[...x]`, `f(...x)`, `new F(...x)`).
// Parentheses around the spread argument (`[...(foo)]`) are transparent in
// ESTree, so unwrap them.
func isArrayOrArgumentSpread(identifier *ast.Node) bool {
	parent := ast.WalkUpParenthesizedExpressions(identifier.Parent)
	if parent == nil || !ast.IsSpreadElement(parent) {
		return false
	}
	if ast.SkipParentheses(parent.AsSpreadElement().Expression) != identifier {
		return false
	}
	grandParent := parent.Parent
	if grandParent == nil {
		return false
	}
	switch grandParent.Kind {
	case ast.KindArrayLiteralExpression, ast.KindCallExpression, ast.KindNewExpression:
		return true
	}
	return false
}

// isAllowedForEachCall reports whether the identifier is the receiver of a
// `identifier.forEach(arrow)` call whose callback is a one-parameter arrow
// function, mirroring upstream's `isAllowedForEachCall`. Parentheses around the
// receiver (`(foo).forEach(...)`) are transparent in ESTree, so walk up through
// them (like isIncludesCall) instead of assuming a fixed ancestor depth.
func isAllowedForEachCall(identifier *ast.Node) bool {
	member := ast.WalkUpParenthesizedExpressions(identifier.Parent)
	if member == nil || !ast.IsPropertyAccessExpression(member) {
		return false
	}
	callExpression := ast.WalkUpParenthesizedExpressions(member.Parent)
	argumentsLength := 1
	match, ok := unicornutil.MatchDotMethodCall(callExpression, unicornutil.DotMethodCallOptions{
		Method:          "forEach",
		ArgumentsLength: &argumentsLength,
	})
	if !ok || ast.SkipParentheses(match.Object) != identifier {
		return false
	}
	return isOneParameterArrowFunction(callExpression.Arguments()[0])
}

// isOneParameterArrowFunction reports whether the node is an arrow function
// with exactly one non-rest parameter, mirroring upstream.
func isOneParameterArrowFunction(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindArrowFunction {
		return false
	}
	params := node.Parameters()
	if len(params) != 1 {
		return false
	}
	return params[0].AsParameterDeclaration().DotDotDotToken == nil
}

// arrayVariableDeclarator mirrors upstream's
// `isArrayVariableDeclaratorIdentifier`: the identifier must be the binding
// name of a non-exported `const foo = <array-source>` declaration. It returns
// the declaration when matched, else nil.
func arrayVariableDeclarator(ctx rule.RuleContext, node *ast.Node) *ast.VariableDeclaration {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return nil
	}
	declaration := parent.AsVariableDeclaration()
	if declaration.Name() != node || declaration.Initializer == nil {
		return nil
	}

	declarationList := parent.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	// Upstream only checks `VariableDeclaration` → `VariableDeclaration` (the
	// declarator) → `VariableDeclaration` (statement); it does NOT require
	// `const` here (the const check happens for referenced initializers). But
	// the identifier must resolve through `findVariable`, and non-const shapes
	// are excluded downstream via reference analysis. Match upstream's exact
	// structural gate: parent list's parent is the statement.
	statement := declarationList.Parent
	if statement == nil || statement.Kind != ast.KindVariableStatement {
		return nil
	}

	// Exclude `export const foo = [];`
	if ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
		return nil
	}

	if !isArrayMethodCall(ctx, declaration.Initializer, map[*ast.Symbol]bool{}) {
		return nil
	}
	return declaration
}
