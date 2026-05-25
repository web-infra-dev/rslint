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
// start with `Is*`, getters with `Get*`. The package never imports
// `internal/utils/` — the rules consume `internal/utils/` separately
// and pass values through.
package react_hooksutil

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
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

// IsFunctionLikeContainer reports whether `node` is one of the
// function-like kinds the upstream rules treat as a "code path"
// boundary.
func IsFunctionLikeContainer(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
	}
	return false
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
		if fn.Parent != nil && fn.Parent.Kind == ast.KindObjectLiteralExpression {
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

// IsForwardRefOrMemoCallback reports whether `fn` is the immediate
// argument of a CallExpression whose callee is `<name>` or
// `React.<name>`.
func IsForwardRefOrMemoCallback(fn *ast.Node, name string) bool {
	if fn == nil || fn.Parent == nil {
		return false
	}
	p := fn.Parent
	if p.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(p.AsCallExpression().Expression)
	return IsReactCalleeNamed(callee, name)
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
		return p.Kind == ast.KindClassDeclaration || p.Kind == ast.KindClassExpression
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		if p.Kind == ast.KindPropertyDeclaration {
			gp := p.Parent
			if gp != nil && (gp.Kind == ast.KindClassDeclaration || gp.Kind == ast.KindClassExpression) {
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
