// Package react_hooksutil holds the AST helpers shared by every
// rule in the `eslint-plugin-react-hooks` port.
//
// Both `rules-of-hooks` and `exhaustive-deps` need the same set of
// queries — "is this a hook callee?", "is this enclosing function a
// component?", "what's the display name of this function-like?",
// "is this a `React.<x>` member access?", etc. Putting them here
// keeps semantics consistent across rules and removes the second
// (and third) copy of every predicate.
//
// Naming convention follows the rest of the codebase: predicates
// start with `Is*`, getters with `Get*`. General AST helpers stay in
// `internal/utils`; this package composes them where they already exist.
package react_hooksutil

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type CompilerReactFunctionType string

const (
	CompilerReactFunctionComponent CompilerReactFunctionType = "Component"
	CompilerReactFunctionHook      CompilerReactFunctionType = "Hook"
)

// hookNameTailRegex matches the suffix part of a hook identifier:
// after the leading `use`, the next character must be uppercase Latin
// or a digit. Mirrors upstream's `/^use[A-Z0-9]/`.
var hookNameTailRegex = regexp.MustCompile(`^use[A-Z0-9]`)

// pascalCaseRegex matches identifiers whose first character is an
// uppercase Latin letter. Mirrors upstream's `/^[A-Z].*/` predicate.
var pascalCaseRegex = regexp.MustCompile(`^[A-Z]`)

// effectNameRegex mirrors upstream `exhaustive-deps`' `/Effect($|[^a-z])/g`
// — used to distinguish "effect"-style hooks (lenient about over-specified
// deps) from "memo"-style hooks (strict).
var effectNameRegex = regexp.MustCompile(`Effect($|[^a-z])`)

// IsHookName reports whether `s` follows the React hook naming
// convention: either the bare `use` (the React `use(...)` hook) or
// `useFoo` / `use1` (`use` followed by uppercase letter or digit).
func IsHookName(s string) bool {
	return s == "use" || hookNameTailRegex.MatchString(s)
}

// IsCompilerHookName reports whether `s` follows the React Compiler
// hook-like naming convention. Unlike rules-of-hooks, the compiler lint
// predicates do not treat the bare `use` identifier as a custom hook name.
func IsCompilerHookName(s string) bool {
	return hookNameTailRegex.MatchString(s)
}

// IsComponentNameStr reports whether `s` looks like a React component
// name — PascalCase. Upstream's `isComponentName` is identical.
func IsComponentNameStr(s string) bool {
	if s == "" {
		return false
	}
	return pascalCaseRegex.MatchString(s)
}

// IsEffectStyleHookName reports whether the bare hook name (no
// `React.` prefix) parses as an effect-style hook by the
// `Effect($|[^a-z])` rule. Used by exhaustive-deps to gate the
// "extra over-specified deps are okay" exception that effects get
// but memo / callback don't.
func IsEffectStyleHookName(name string) bool {
	return effectNameRegex.MatchString(name)
}

// StripReactNamespace returns the property identifier of a
// `React.foo` member expression, or the original node otherwise.
// Mirrors upstream's `getNodeWithoutReactNamespace`. Paren wrappers
// are peeled transparently — tsgo preserves them where ESTree
// flattens.
func StripReactNamespace(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindPropertyAccessExpression {
		pae := n.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pae.Expression)
		prop := pae.Name()
		if obj != nil && obj.Kind == ast.KindIdentifier &&
			obj.AsIdentifier().Text == "React" &&
			prop != nil && prop.Kind == ast.KindIdentifier {
			return prop
		}
	}
	return n
}

// IsReactCalleeNamed reports whether `node` is `<name>` (bare
// identifier) or `React.<name>` (PropertyAccessExpression with object
// `React` and property `<name>`). Mirrors upstream's
// `isReactFunction(node, functionName)`.
func IsReactCalleeNamed(node *ast.Node, name string) bool {
	n := StripReactNamespace(node)
	if n == nil || n.Kind != ast.KindIdentifier {
		return false
	}
	return n.AsIdentifier().Text == name
}

// IsUseEffectEventCallee reports whether `node` is the bare
// `useEffectEvent` or `React.useEffectEvent` callee.
func IsUseEffectEventCallee(node *ast.Node) bool {
	return IsReactCalleeNamed(node, "useEffectEvent")
}

// IsUseIdentifier reports whether `node` is the React `use(...)`
// callee — either bare `use` or `React.use`.
func IsUseIdentifier(node *ast.Node) bool {
	return IsReactCalleeNamed(node, "use")
}

// IsManualUseMemoCallee reports whether `node` is a direct `useMemo` or
// `React.useMemo` callee according to React Compiler's manual memoization
// input surface. It intentionally does not recognize import aliases or
// element-access forms; existing React Compiler lint ports rely on the same
// direct-call contract.
func IsManualUseMemoCallee(node *ast.Node, typeChecker *checker.Checker) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return isManualMemoIdentifier(node, "useMemo", typeChecker)
	case ast.KindPropertyAccessExpression:
		if ast.IsOptionalChain(node) {
			return false
		}
		access := node.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "useMemo" {
			return false
		}
		obj := ast.SkipParentheses(access.Expression)
		if obj == nil || obj.Kind != ast.KindIdentifier {
			return false
		}
		return isManualMemoIdentifier(obj, "React", typeChecker)
	}
	return false
}

func isManualMemoIdentifier(id *ast.Node, expected string, typeChecker *checker.Checker) bool {
	if id == nil || id.Kind != ast.KindIdentifier || id.AsIdentifier().Text != expected {
		return false
	}
	if typeChecker == nil {
		return true
	}
	sym := utils.GetReferenceSymbol(id, typeChecker)
	if sym == nil || len(sym.Declarations) == 0 {
		return true
	}
	for _, decl := range sym.Declarations {
		if isManualMemoDeclaration(decl, expected) {
			return true
		}
	}
	return false
}

func isManualMemoDeclaration(decl *ast.Node, expected string) bool {
	if decl == nil {
		return false
	}
	switch decl.Kind {
	case ast.KindImportClause, ast.KindNamespaceImport, ast.KindImportSpecifier:
		return true
	case ast.KindBindingElement:
		return !isInsideParameter(decl)
	case ast.KindVariableDeclaration:
		name := decl.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != expected {
			return false
		}
		initializer := ast.SkipParentheses(decl.AsVariableDeclaration().Initializer)
		if initializer == nil {
			return true
		}
		if expected == "useMemo" && ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return false
		}
		if expected == "React" && initializer.Kind == ast.KindObjectLiteralExpression {
			return false
		}
		return true
	case ast.KindParameter, ast.KindFunctionDeclaration:
		return false
	}
	return false
}

func isInsideParameter(node *ast.Node) bool {
	for cur := node; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindParameter:
			return true
		case ast.KindVariableDeclaration:
			return false
		}
	}
	return false
}

// IsHookCallee mirrors upstream's `isHook(node)`: the callee is
// either an Identifier whose name is a hook, or a non-computed member
// expression `Namespace.useFoo` whose object is a PascalCase
// identifier and whose property is itself a hook name.
func IsHookCallee(node *ast.Node) bool {
	if node == nil {
		return false
	}
	n := ast.SkipParentheses(node)
	switch n.Kind {
	case ast.KindIdentifier:
		return IsHookName(n.AsIdentifier().Text)
	case ast.KindPropertyAccessExpression:
		pae := n.AsPropertyAccessExpression()
		prop := pae.Name()
		if prop == nil || prop.Kind != ast.KindIdentifier {
			return false
		}
		if !IsHookName(prop.AsIdentifier().Text) {
			return false
		}
		obj := ast.SkipParentheses(pae.Expression)
		if obj == nil || obj.Kind != ast.KindIdentifier {
			return false
		}
		return IsComponentNameStr(obj.AsIdentifier().Text)
	}
	return false
}

// IsCompilerHookCallee is the React Compiler variant of IsHookCallee:
// it accepts `useFoo` / `Namespace.useFoo`, but not the bare `use`.
func IsCompilerHookCallee(node *ast.Node) bool {
	if node == nil {
		return false
	}
	n := ast.SkipParentheses(node)
	switch n.Kind {
	case ast.KindIdentifier:
		return IsCompilerHookName(n.AsIdentifier().Text)
	case ast.KindPropertyAccessExpression:
		pae := n.AsPropertyAccessExpression()
		prop := pae.Name()
		if prop == nil || prop.Kind != ast.KindIdentifier || !IsCompilerHookName(prop.AsIdentifier().Text) {
			return false
		}
		obj := ast.SkipParentheses(pae.Expression)
		return obj != nil && obj.Kind == ast.KindIdentifier && IsComponentNameStr(obj.AsIdentifier().Text)
	}
	return false
}

// IsCompilerFunctionKind reports whether `node` is one of the function
// syntaxes that React Compiler's Factories diagnostic traverses:
// FunctionDeclaration, FunctionExpression, or ArrowFunctionExpression.
func IsCompilerFunctionKind(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.IsFunctionDeclaration(node) || ast.IsFunctionExpressionOrArrowFunction(node)
}

// AccessChainRootIdentifier returns the identifier at the root of an access
// chain like `value.x.y` or `(value as T)["x"]`.
func AccessChainRootIdentifier(node *ast.Node) *ast.Node {
	node = utils.SkipAssertionsAndParens(node)
	for node != nil && ast.IsAccessExpression(node) {
		node = utils.SkipAssertionsAndParens(utils.AccessExpressionObject(node))
	}
	if node != nil && node.Kind == ast.KindIdentifier {
		return node
	}
	return nil
}

// ImportSpecifierImportedName returns the module-exported name for an import
// specifier. For `import {foo as bar}`, this is `foo`; for `import {bar}`,
// this is `bar`.
func ImportSpecifierImportedName(spec *ast.ImportSpecifier) string {
	if spec == nil {
		return ""
	}
	name := spec.Name()
	if spec.PropertyName != nil {
		name = spec.PropertyName
	}
	return moduleExportNameText(name)
}

func moduleExportNameText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return node.Text()
	}
	return ""
}

// AssignmentTargetIdentifier is a binding written by an assignment target.
// Identifier points at the actual identifier node, while Node is the broader
// target to report when destructuring defaults need the full assignment node.
type AssignmentTargetIdentifier struct {
	Identifier *ast.Node
	Node       *ast.Node
	Name       string
}

// CollectAssignmentTargetIdentifiers returns every identifier written by an
// assignment target, including nested array/object destructuring targets and
// default values in destructuring patterns. It only peels parentheses, matching
// the upstream globals rule's assignment-target handling.
func CollectAssignmentTargetIdentifiers(node *ast.Node) []AssignmentTargetIdentifier {
	var targets []AssignmentTargetIdentifier
	collectAssignmentTargetIdentifiersInto(node, &targets, false)
	return targets
}

// CollectAssignmentTargetIdentifiersThroughAssertions is the same assignment
// target collector, but it also peels TS assertion wrappers before matching the
// target shape.
func CollectAssignmentTargetIdentifiersThroughAssertions(node *ast.Node) []AssignmentTargetIdentifier {
	var targets []AssignmentTargetIdentifier
	collectAssignmentTargetIdentifiersInto(node, &targets, true)
	return targets
}

func collectAssignmentTargetIdentifiersInto(node *ast.Node, targets *[]AssignmentTargetIdentifier, throughAssertions bool) {
	node = skipAssignmentTargetWrappers(node, throughAssertions)
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		appendAssignmentTargetIdentifier(targets, node, node)
	case ast.KindObjectLiteralExpression:
		obj := node.AsObjectLiteralExpression()
		if obj == nil || obj.Properties == nil {
			return
		}
		for _, prop := range obj.Properties.Nodes {
			switch prop.Kind {
			case ast.KindShorthandPropertyAssignment:
				shorthand := prop.AsShorthandPropertyAssignment()
				name := prop.Name()
				if shorthand != nil && shorthand.ObjectAssignmentInitializer != nil {
					appendAssignmentTargetIdentifier(targets, name, prop)
					continue
				}
				collectAssignmentTargetIdentifiersInto(name, targets, throughAssertions)
			case ast.KindPropertyAssignment:
				assignment := prop.AsPropertyAssignment()
				if assignment != nil {
					collectAssignmentTargetIdentifiersInto(assignment.Initializer, targets, throughAssertions)
				}
			case ast.KindSpreadAssignment:
				spread := prop.AsSpreadAssignment()
				if spread != nil {
					collectAssignmentTargetIdentifiersInto(spread.Expression, targets, throughAssertions)
				}
			}
		}
	case ast.KindArrayLiteralExpression:
		array := node.AsArrayLiteralExpression()
		if array == nil || array.Elements == nil {
			return
		}
		for _, elem := range array.Elements.Nodes {
			collectAssignmentTargetIdentifiersInto(elem, targets, throughAssertions)
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
			return
		}
		left := skipAssignmentTargetWrappers(binary.Left, throughAssertions)
		if left != nil && left.Kind == ast.KindIdentifier {
			appendAssignmentTargetIdentifier(targets, left, node)
			return
		}
		collectAssignmentTargetIdentifiersInto(left, targets, throughAssertions)
	case ast.KindSpreadElement:
		spread := node.AsSpreadElement()
		if spread != nil {
			collectAssignmentTargetIdentifiersInto(spread.Expression, targets, throughAssertions)
		}
	}
}

func skipAssignmentTargetWrappers(node *ast.Node, throughAssertions bool) *ast.Node {
	if throughAssertions {
		return utils.SkipAssertionsAndParens(node)
	}
	return ast.SkipParentheses(node)
}

func appendAssignmentTargetIdentifier(targets *[]AssignmentTargetIdentifier, id, reportNode *ast.Node) {
	if id == nil || id.Kind != ast.KindIdentifier {
		return
	}
	name := id.AsIdentifier().Text
	if name == "" {
		return
	}
	if reportNode == nil {
		reportNode = id
	}
	*targets = append(*targets, AssignmentTargetIdentifier{
		Identifier: id,
		Node:       reportNode,
		Name:       name,
	})
}

// ContainsNode reports whether `descendant` is inside `ancestor` in the same
// source file.
func ContainsNode(ancestor, descendant *ast.Node) bool {
	if ancestor == nil || descendant == nil {
		return false
	}
	if descendant.Pos() < ancestor.Pos() || descendant.End() > ancestor.End() {
		return false
	}
	return ast.GetSourceFileOfNode(ancestor) == ast.GetSourceFileOfNode(descendant)
}

// GetCompilerReactFunctionType mirrors the React Compiler classifier used by
// the Factories diagnostic: a PascalCase function is a component only when it
// directly creates JSX or calls a hook, has component-like parameters, and does
// not return obviously non-ReactNode values; a hook-like function must directly
// create JSX or call a hook; memo/forwardRef callbacks are components.
func GetCompilerReactFunctionType(fn *ast.Node) CompilerReactFunctionType {
	name := GetFunctionName(fn)
	if name != nil && name.Kind == ast.KindIdentifier && IsComponentNameStr(name.AsIdentifier().Text) {
		if CallsHooksOrCreatesJsx(fn) &&
			IsValidCompilerComponentParams(fn) &&
			!ReturnsCompilerNonNode(fn) {
			return CompilerReactFunctionComponent
		}
		return ""
	}
	if name != nil && IsCompilerHookCallee(name) {
		if CallsHooksOrCreatesJsx(fn) {
			return CompilerReactFunctionHook
		}
		return ""
	}
	if ast.IsFunctionExpressionOrArrowFunction(fn) {
		if IsForwardRefOrMemoCallback(fn, "forwardRef") || IsForwardRefOrMemoCallback(fn, "memo") {
			if CallsHooksOrCreatesJsx(fn) {
				return CompilerReactFunctionComponent
			}
		}
	}
	return ""
}

// CallsHooksOrCreatesJsx reports whether `fn` directly creates JSX or directly
// calls a compiler hook. Nested function bodies are skipped, matching React
// Compiler's traversal boundary.
func CallsHooksOrCreatesJsx(fn *ast.Node) bool {
	found := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		if node != fn && IsCompilerFunctionKind(node) {
			return
		}
		if ast.IsJsxElement(node) || ast.IsJsxSelfClosingElement(node) || ast.IsJsxFragment(node) {
			found = true
			return
		}
		if ast.IsCallExpression(node) {
			if IsCompilerHookCallee(node.AsCallExpression().Expression) {
				found = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(fn)
	return found
}

// IsValidCompilerComponentParams mirrors React Compiler's
// isValidComponentParams helper.
func IsValidCompilerComponentParams(fn *ast.Node) bool {
	params := fn.Parameters()
	if len(params) == 0 {
		return true
	}
	if len(params) > 2 || !isValidCompilerPropsAnnotation(params[0]) {
		return false
	}
	if len(params) == 1 {
		return !utils.IsRestParameterDeclaration(params[0])
	}
	secondName := params[1].AsParameterDeclaration().Name()
	if secondName == nil || secondName.Kind != ast.KindIdentifier {
		return false
	}
	name := secondName.AsIdentifier().Text
	return strings.Contains(name, "ref") || strings.Contains(name, "Ref")
}

func isValidCompilerPropsAnnotation(param *ast.Node) bool {
	if param == nil || !ast.IsParameterDeclaration(param) {
		return false
	}
	typ := ast.GetTypeAnnotationNode(param)
	if typ == nil {
		return true
	}
	switch typ.Kind {
	case ast.KindArrayType, ast.KindBigIntKeyword, ast.KindBooleanKeyword,
		ast.KindConstructorType, ast.KindFunctionType, ast.KindLiteralType,
		ast.KindNeverKeyword, ast.KindNumberKeyword, ast.KindStringKeyword,
		ast.KindSymbolKeyword, ast.KindTupleType:
		return false
	}
	return true
}

// ReturnsCompilerNonNode mirrors React Compiler's returnsNonNode helper.
func ReturnsCompilerNonNode(fn *ast.Node) bool {
	returnsNonNode := false
	if ast.IsArrowFunction(fn) {
		body := fn.AsArrowFunction().Body
		if body != nil && body.Kind != ast.KindBlock {
			returnsNonNode = IsCompilerNonNode(body)
		}
	}

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != fn && IsCompilerFunctionKind(node) {
			return
		}
		if isCompilerObjectMethod(node) {
			return
		}
		if node.Kind == ast.KindReturnStatement {
			returnsNonNode = IsCompilerNonNode(node.AsReturnStatement().Expression)
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(fn)
	return returnsNonNode
}

func isCompilerObjectMethod(node *ast.Node) bool {
	return node != nil && ast.IsObjectLiteralMethod(node)
}

// IsCompilerNonNode mirrors React Compiler's isNonNode helper for returns
// whose value is definitely not a React node.
func IsCompilerNonNode(node *ast.Node) bool {
	if node == nil {
		return true
	}
	n := ast.SkipParentheses(node)
	if ast.IsObjectLiteralExpression(n) ||
		ast.IsFunctionExpressionOrArrowFunction(n) ||
		ast.IsClassExpression(n) ||
		ast.IsNewExpression(n) {
		return true
	}
	return ast.IsBigIntLiteral(n)
}

// IsFunctionLikeContainer keeps the react-hooks shared API while delegating to
// the repository-wide function-scope boundary helper.
func IsFunctionLikeContainer(node *ast.Node) bool {
	return utils.IsFunctionLikeContainer(node)
}

// FindEnclosingFunction walks up from `node` and returns the nearest
// function-like ancestor, or nil when `node` is at top level.
func FindEnclosingFunction(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	p := node.Parent
	for p != nil {
		if IsFunctionLikeContainer(p) {
			return p
		}
		p = p.Parent
	}
	return nil
}

// GetFunctionBody returns the body of a function-like node — Block
// for normal functions, the BlockOrExpression body for arrow
// functions, or nil for abstract/declared signatures.
func GetFunctionBody(fn *ast.Node) *ast.Node {
	if fn == nil {
		return nil
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		return fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		return fn.AsArrowFunction().Body
	case ast.KindMethodDeclaration:
		return fn.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		return fn.AsGetAccessorDeclaration().Body
	case ast.KindSetAccessor:
		return fn.AsSetAccessorDeclaration().Body
	case ast.KindConstructor:
		return fn.AsConstructorDeclaration().Body
	}
	return nil
}

// HasAsyncModifier reports whether the function-like node carries
// the `async` modifier.
func HasAsyncModifier(fn *ast.Node) bool {
	if fn == nil {
		return false
	}
	return ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync)
}

// GetFunctionName mirrors upstream's `getFunctionName(node)`. Returns:
//   - For named FunctionDeclaration / FunctionExpression: the `id` Identifier.
//   - For MethodDeclaration / GetAccessor / SetAccessor inside an
//     ObjectLiteralExpression: the method's own name.
//   - For ArrowFunction / anonymous FunctionExpression: the assignment
//     target — VariableDeclaration name, BinaryExpression LHS,
//     PropertyAssignment name, BindingElement name, or
//     ShorthandPropertyAssignment name.
//   - nil otherwise.
//
// MethodDeclaration / GetAccessor / SetAccessor inside a class body
// intentionally returns nil so the class-member branch can take over.
func GetFunctionName(fn *ast.Node) *ast.Node {
	if fn == nil {
		return nil
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		return fn.Name()
	case ast.KindFunctionExpression:
		if name := fn.Name(); name != nil {
			return name
		}
		// fall through to anonymous handling
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if fn.Parent != nil && ast.IsObjectLiteralExpression(fn.Parent) {
			return fn.Name()
		}
		return nil
	case ast.KindArrowFunction:
		// fall through
	default:
		return nil
	}
	// Walk past wrapping ParenthesizedExpression layers so shapes like
	// `const foo = (() => {})` are recognized: the function-like itself
	// is the parens' inner expression, but the meaningful "parent kind"
	// for name lookup is the VariableDeclaration ABOVE the parens.
	// `child` tracks the node we're checking for `parent.X === child`
	// equality (must be the wrapper after each peel, not the original
	// fn).
	child := fn
	p := fn.Parent
	for p != nil && p.Kind == ast.KindParenthesizedExpression {
		child = p
		p = p.Parent
	}
	if p == nil {
		return nil
	}
	switch p.Kind {
	case ast.KindVariableDeclaration:
		vd := p.AsVariableDeclaration()
		if vd.Initializer == child {
			return p.Name()
		}
	case ast.KindBinaryExpression:
		be := p.AsBinaryExpression()
		if be.OperatorToken != nil && be.OperatorToken.Kind == ast.KindEqualsToken && be.Right == child {
			return be.Left
		}
	case ast.KindPropertyAssignment:
		pa := p.AsPropertyAssignment()
		if pa.Initializer == child {
			return p.Name()
		}
	case ast.KindBindingElement:
		be := p.AsBindingElement()
		if be.Initializer == child {
			return p.Name()
		}
	case ast.KindShorthandPropertyAssignment:
		spa := p.AsShorthandPropertyAssignment()
		if spa.ObjectAssignmentInitializer == child {
			return p.Name()
		}
	}
	return nil
}

// GetReactCallbackCall returns the CallExpression when `fn` is the immediate
// argument of a call whose callee is `<name>` or `React.<name>`.
// Parenthesized callback expressions are transparent: `memo((() => null))` is
// the same callback shape as `memo(() => null)`.
func GetReactCallbackCall(fn *ast.Node, name string) *ast.Node {
	if fn == nil {
		return nil
	}
	child := fn
	p := fn.Parent
	for p != nil && p.Kind == ast.KindParenthesizedExpression {
		child = p
		p = p.Parent
	}
	if p == nil || p.Kind != ast.KindCallExpression {
		return nil
	}
	call := p.AsCallExpression()
	if call.Arguments == nil {
		return nil
	}
	isArg := false
	for _, arg := range call.Arguments.Nodes {
		if arg == child {
			isArg = true
			break
		}
	}
	if !isArg {
		return nil
	}
	callee := ast.SkipParentheses(call.Expression)
	if !IsReactCalleeNamed(callee, name) {
		return nil
	}
	return p
}

// GetForwardRefOrMemoCallbackCall is kept for existing callers that only check
// React's `memo` / `forwardRef` callback shapes.
func GetForwardRefOrMemoCallbackCall(fn *ast.Node, name string) *ast.Node {
	return GetReactCallbackCall(fn, name)
}

// IsForwardRefOrMemoCallback reports whether `fn` is the immediate
// argument of a CallExpression whose callee is `<name>` or
// `React.<name>`.
func IsForwardRefOrMemoCallback(fn *ast.Node, name string) bool {
	return GetForwardRefOrMemoCallbackCall(fn, name) != nil
}

// IsClassMember reports whether `fn` is a member of a class — either
// a MethodDeclaration / GetAccessor / SetAccessor / Constructor whose
// direct parent is a class, or an ArrowFunction / FunctionExpression
// that initializes a class PropertyDeclaration (class-field arrow).
func IsClassMember(fn *ast.Node) bool {
	if fn == nil || fn.Parent == nil {
		return false
	}
	p := fn.Parent
	switch fn.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return ast.IsClassLike(p)
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		if p.Kind == ast.KindPropertyDeclaration {
			gp := p.Parent
			if gp != nil && ast.IsClassLike(gp) {
				return true
			}
		}
	}
	return false
}

// IsComponentOrHookFn reports whether the function-like itself is a
// React component or hook (named appropriately, or an anonymous arg
// to forwardRef / memo). Mirrors upstream's
// `isDirectlyInsideComponentOrHook` predicate at the function level.
func IsComponentOrHookFn(fn *ast.Node) bool {
	name := GetFunctionName(fn)
	if name != nil {
		switch name.Kind {
		case ast.KindIdentifier:
			text := name.AsIdentifier().Text
			return IsComponentNameStr(text) || IsHookName(text)
		case ast.KindPropertyAccessExpression:
			return IsHookCallee(name)
		}
		return false
	}
	return IsForwardRefOrMemoCallback(fn, "forwardRef") || IsForwardRefOrMemoCallback(fn, "memo")
}

// IsInsideComponentOrHook walks up from `node` and returns true once
// any ancestor function-like classifies as a component or hook.
func IsInsideComponentOrHook(node *ast.Node) bool {
	cur := node
	for cur != nil {
		if IsFunctionLikeContainer(cur) && IsComponentOrHookFn(cur) {
			return true
		}
		cur = cur.Parent
	}
	return false
}

// AdditionalHooksFromSettings reads
// `settings['react-hooks'].<key>` (typically `additionalHooks` for
// exhaustive-deps, or `additionalEffectHooks` for rules-of-hooks)
// and compiles it as a regex. Returns nil when the setting is absent
// or the pattern fails to compile — mirroring upstream's lenient
// behavior of silently ignoring malformed regex strings.
func AdditionalHooksFromSettings(settings map[string]interface{}, key string) *regexp.Regexp {
	if settings == nil {
		return nil
	}
	raw, ok := settings["react-hooks"]
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	pattern, ok := m[key].(string)
	if !ok || pattern == "" {
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re
}
