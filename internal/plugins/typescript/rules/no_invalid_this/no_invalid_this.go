package no_invalid_this

import (
	"regexp"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// NoInvalidThisRule mirrors @typescript-eslint/no-invalid-this, which wraps
// ESLint core's no-invalid-this with two TypeScript-specific recognitions:
//   - A function whose signature declares a `this` parameter
//     (`function foo(this: T)`) has its `this` validity short-circuited to
//     `true`, matching the upstream `thisIsValidStack` push.
//   - Class field initializers (regular and `accessor`-modified) are
//     short-circuited to `true`, since their implicit-function context
//     binds `this` to the class instance.
//
// All other validity decisions delegate to the same parent-walker logic the
// upstream ESLint rule uses (`isDefaultThisBinding`), with `capIsConstructor`
// honored at every uppercase-name branch (VariableDeclaration init,
// AssignmentExpression target, ParameterDeclaration default,
// BindingElement default).
//
// https://typescript-eslint.io/rules/no-invalid-this
var NoInvalidThisRule = rule.CreateRule(rule.Rule{
	Name: "no-invalid-this",
	Run:  run,
})

type ruleOptions struct {
	// capIsConstructor: when true (default), a capitalized-name function is
	// treated as an ES5 constructor and its `this` is considered valid.
	capIsConstructor bool
}

func parseOptions(raw any) ruleOptions {
	opts := ruleOptions{capIsConstructor: true}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["capIsConstructor"]; ok {
		if b, ok := v.(bool); ok {
			opts.capIsConstructor = b
		}
	}
	return opts
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)
	sf := ctx.SourceFile

	// Top-level `this` validity. typescript-eslint's wrapper defaults to
	// `parserOptions.sourceType: 'module'`, which makes top-level `this`
	// always invalid. rslint does not expose `sourceType` /
	// `parserOptions.ecmaFeatures.globalReturn`, so we adopt the same
	// default and treat top-level `this` as invalid — a framework-layer
	// consequence of rslint not surfacing parser options, applied
	// uniformly across rules.
	topLevelValid := false

	// Stack of `this`-validity flags, one per non-arrow function-like /
	// class-member container currently on the visitor's path. Arrow functions
	// inherit the surrounding frame and therefore do NOT push (lexical `this`).
	var stack []bool

	push := func(valid bool) {
		stack = append(stack, valid)
	}
	pop := func() {
		if n := len(stack); n > 0 {
			stack = stack[:n-1]
		}
	}

	msg := rule.RuleMessage{
		Id:          "unexpectedThis",
		Description: "Unexpected 'this'.",
	}

	pushFunction := func(node *ast.Node) {
		push(computeFunctionValid(node, sf, opts.capIsConstructor))
	}

	// hasComputedKey reports whether a class member uses a `[expr]` computed
	// key. For Method / Constructor / Get/SetAccessor the wrapper's
	// FunctionExpression push fires on the FE child of MethodDefinition —
	// which in ESTree is visited AFTER the computed key. We mirror this by
	// deferring the push to `ComputedPropertyName:exit`.
	hasComputedKey := func(node *ast.Node) bool {
		name := ast.GetNameOfDeclaration(node)
		return name != nil && name.Kind == ast.KindComputedPropertyName
	}

	// enterMethodLike defers push past the computed key (if any). Applies
	// to Method/Constructor/Get/Set whose ESTree counterpart's wrapper push
	// happens on the FunctionExpression value, AFTER the key visit.
	enterMethodLike := func(node *ast.Node) {
		if hasComputedKey(node) {
			return
		}
		push(true)
	}

	// enterPropertyDeclaration always pushes on entry. The upstream wrapper's
	// `PropertyDefinition` / `AccessorProperty` listeners push on entry too,
	// which intentionally — or unintentionally — masks `this` in computed
	// keys of class fields. We reproduce that behavior verbatim so output
	// matches `@typescript-eslint/no-invalid-this` byte for byte.
	enterPropertyDeclaration := func(*ast.Node) {
		push(true)
	}

	return rule.RuleListeners{
		// Non-arrow function-like containers — push a frame whose validity
		// depends on parameter shape, JSDoc, name, and surrounding context.
		ast.KindFunctionDeclaration:                      pushFunction,
		rule.ListenerOnExit(ast.KindFunctionDeclaration): func(*ast.Node) { pop() },
		ast.KindFunctionExpression:                       pushFunction,
		rule.ListenerOnExit(ast.KindFunctionExpression):  func(*ast.Node) { pop() },

		// Class members (and equivalent object-literal accessors): `this`
		// always refers to the class instance / static class object — VALID.
		// Computed-key members defer the push to ComputedPropertyName:exit
		// to mirror the upstream wrapper, whose `FunctionExpression`
		// listener fires on the method's FE value (visited AFTER the key).
		ast.KindMethodDeclaration:                      enterMethodLike,
		rule.ListenerOnExit(ast.KindMethodDeclaration): func(*ast.Node) { pop() },
		ast.KindConstructor:                            enterMethodLike,
		rule.ListenerOnExit(ast.KindConstructor):       func(*ast.Node) { pop() },
		ast.KindGetAccessor:                            enterMethodLike,
		rule.ListenerOnExit(ast.KindGetAccessor):       func(*ast.Node) { pop() },
		ast.KindSetAccessor:                            enterMethodLike,
		rule.ListenerOnExit(ast.KindSetAccessor):       func(*ast.Node) { pop() },

		// Class field (regular + `accessor` auto-accessor — tsgo collapses
		// ESTree's PropertyDefinition / AccessorProperty onto this kind,
		// distinguishing the latter via `ModifierFlagsAccessor`). Push on
		// entry verbatim with upstream wrapper's `PropertyDefinition()` /
		// `AccessorProperty()` listeners — including the wrapper's behavior
		// of masking `this` in computed keys and decorators.
		ast.KindPropertyDeclaration:                      enterPropertyDeclaration,
		rule.ListenerOnExit(ast.KindPropertyDeclaration): func(*ast.Node) { pop() },

		// Deferred push for computed-key method-likes. The matching pop
		// happens unconditionally on the member's own exit listener, so
		// the stack stays balanced regardless of whether the push happened
		// on enter (non-computed) or here (computed). PropertyDeclaration
		// is intentionally excluded — it pushes on entry.
		rule.ListenerOnExit(ast.KindComputedPropertyName): func(node *ast.Node) {
			parent := node.Parent
			if parent == nil {
				return
			}
			switch parent.Kind {
			case ast.KindMethodDeclaration, ast.KindConstructor,
				ast.KindGetAccessor, ast.KindSetAccessor:
				push(true)
			}
		},

		// Class static block: own `this` context bound to the class — VALID.
		ast.KindClassStaticBlockDeclaration:                      func(*ast.Node) { push(true) },
		rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): func(*ast.Node) { pop() },

		// Arrow function: lexical `this`. Intentionally NOT registered so
		// the enclosing frame governs.

		ast.KindThisKeyword: func(node *ast.Node) {
			// Decorators on Method / Constructor / Get/Set members are
			// visited BEFORE the wrapper's FunctionExpression push fires
			// in ESTree (the wrapper hooks FE entry, not MethodDefinition
			// entry). tsgo collapses MethodDefinition's FE child into the
			// member node itself, so our `enterMethodLike` push happens
			// one visitor tick too early to reproduce that timing. To
			// compensate, peek one frame deeper when `this` appears inside
			// such a decorator. PropertyDeclaration is intentionally NOT
			// in this list: upstream's `PropertyDefinition` / `AccessorProperty`
			// wrapper listeners DO push on entry, so decorators on fields
			// see the field's frame (matching our default behavior).
			skip := 0
			if isInsideDecoratorOfMethodLike(node) {
				skip = 1
			}
			idx := len(stack) - 1 - skip
			var valid bool
			if idx < 0 {
				valid = topLevelValid
			} else {
				valid = stack[idx]
			}
			if !valid {
				ctx.ReportNode(node, msg)
			}
		},
	}
}

// computeFunctionValid produces the `this`-validity flag pushed when a
// FunctionDeclaration / FunctionExpression frame is entered. Order matches
// upstream's combined wrapper+base logic:
//  1. `this` parameter on the signature → VALID (typescript-eslint wrapper).
//  2. `@this` JSDoc tag attached to the function (or its statement context)
//     → VALID.
//  3. `capIsConstructor: true` AND the function carries an uppercase own name
//     (`function Foo()` / `var x = function Bar()`) → VALID (ES5 constructor).
//  4. Otherwise, walk the parent chain via `isDefaultThisBinding` — VALID
//     iff the surrounding context binds `this` explicitly (method assignment,
//     `.call`/`.apply`/`.bind`, `Reflect.apply`, array-method `thisArg`, …).
func computeFunctionValid(node *ast.Node, sf *ast.SourceFile, capIsConstructor bool) bool {
	if hasThisParameter(node) {
		return true
	}
	if hasJSDocThisTag(node, sf) {
		return true
	}
	if capIsConstructor && isES5Constructor(node) {
		return true
	}
	return !isDefaultThisBinding(node, capIsConstructor)
}

// hasThisParameter reports whether the function's parameter list begins with
// (or contains) a `this: T` parameter. Mirrors the upstream wrapper's
// `params.some(param => param.type === Identifier && param.name === 'this')`.
// tsgo encodes the `this` parameter as a ParameterDeclaration whose `name`
// is an Identifier with text `"this"`, which `ast.IsThisParameter` resolves.
func hasThisParameter(node *ast.Node) bool {
	for _, p := range node.Parameters() {
		if ast.IsThisParameter(p) {
			return true
		}
	}
	return false
}

// isES5Constructor mirrors ESLint's `isES5Constructor`: a function with an
// own name whose first character is an uppercase letter is treated as an
// ES5 constructor under `capIsConstructor: true`. Anonymous functions
// (no own name) fall through.
func isES5Constructor(node *ast.Node) bool {
	var name *ast.Node
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		name = node.AsFunctionDeclaration().Name()
	case ast.KindFunctionExpression:
		name = node.AsFunctionExpression().Name()
	default:
		return false
	}
	if name == nil || !ast.IsIdentifier(name) {
		return false
	}
	return startsWithUpperCase(name.AsIdentifier().Text)
}

// hasOwnFunctionName reports whether the function has an `id`/name in
// ESTree terms (used to gate ES5-constructor recognition for the uppercase-
// variable / uppercase-assignment-target branches: a *named* function
// expression keeps its own name as the binding, so the uppercase-of-the-
// surrounding-target heuristic does not apply).
func hasOwnFunctionName(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Name() != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Name() != nil
	}
	return false
}

func startsWithUpperCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// ESLint uses `s[0] !== s[0].toLocaleLowerCase()` which captures any
	// Unicode uppercase letter. Restrict to ASCII for now — the tsgo
	// fixtures all use ASCII names and the typescript-eslint suite never
	// exercises the Unicode branch.
	c := s[0]
	return c >= 'A' && c <= 'Z'
}

// thisTagPattern is ESLint's exact pattern: `^[\s*]*@this` applied to a
// comment's *value*, i.e. with `/*` / `*/` / `//` markers stripped.
var thisTagPattern = regexp.MustCompile(`(?m)^[\s*]*@this\b`)

// hasJSDocThisTag mirrors `astUtils.hasJSDocThisTag`. Two sources are checked
// per ESLint:
//  1. The function's own JSDoc comment — either attached directly, or, when
//     the function is the value of a transparent expression context (return
//     statement, variable initializer, call argument, …), the JSDoc attached
//     to the enclosing statement. eslint-utils's `getJSDocComment` walks up
//     through such transparent ancestors; we replicate that walk.
//  2. The function's leading non-JSDoc comments (`getCommentsBefore`) — covers
//     the callback-with-inline-tag shape `foo(/* @this */ function(){})`,
//     where the comment sits between the call's `(` and the function and
//     therefore lives in the function's leading trivia.
//
// Together these match every case the upstream test suite exercises:
//   - `/** @this */ function foo()` (own JSDoc)
//   - `function foo() { /** @this */ return function bar() {} }` (parent ReturnStatement JSDoc)
//   - `foo(/* @this */ function(){})` (leading comment between `(` and `function`)
//
// Out of scope (and correctly NOT matched): `/** @this */ foo(function(){})`
// — the JSDoc here belongs to the enclosing CallExpression, not the function
// argument; ESLint's `getJSDocComment` stops walking at CallExpression
// parents, so we do too.
func hasJSDocThisTag(fn *ast.Node, sf *ast.SourceFile) bool {
	text := sf.Text()
	if hasThisTagInLeadingComments(fn, text) {
		return true
	}
	// Walk up through transparent statement-context parents. Stop at the
	// first parent whose comments we should *not* attribute to the function
	// (e.g. CallExpression — upstream's `getJSDocComment` excludes it).
	current := fn
	for {
		parent := current.Parent
		// Skip parens and TS expression wrappers — eslint-utils's
		// `getJSDocComment` keeps walking past them when looking for a
		// JSDoc anchor, so we do the same.
		for parent != nil && ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions) {
			current = parent
			parent = current.Parent
		}
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindReturnStatement, ast.KindExpressionStatement:
			return hasThisTagInLeadingComments(parent, text)
		case ast.KindVariableDeclaration:
			// Walk through VariableDeclarationList → VariableStatement so the
			// JSDoc on the statement itself is checked. tsgo splits what
			// ESTree models as a flat VariableDeclaration into three nested
			// nodes; the user-visible JSDoc anchor is the outermost.
			grand := parent.Parent
			if grand != nil && grand.Kind == ast.KindVariableDeclarationList {
				vs := grand.Parent
				if vs != nil && vs.Kind == ast.KindVariableStatement {
					return hasThisTagInLeadingComments(vs, text)
				}
			}
			return hasThisTagInLeadingComments(parent, text)
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
			ast.KindPropertyDeclaration, ast.KindBindingElement, ast.KindParameter:
			return hasThisTagInLeadingComments(parent, text)
		case ast.KindBinaryExpression:
			// Assignment / logical / conditional — defer to the enclosing
			// statement. Continue walking.
			current = parent
			continue
		case ast.KindConditionalExpression:
			current = parent
			continue
		default:
			return false
		}
	}
}

// hasThisTagInLeadingComments scans the leading-comment ranges at `node.Pos()`
// for `@this`. `node.Pos()` in tsgo lands BEFORE the leading trivia, so the
// scanner returns every comment that visually precedes the node within the
// current scope. The comment range as tsgo reports it includes the `/*` /
// `*/` / `//` markers; ESLint's comment objects expose only the value, so
// we strip the markers before applying the regex.
func hasThisTagInLeadingComments(node *ast.Node, text string) bool {
	if node == nil {
		return false
	}
	nodeFactory := &ast.NodeFactory{}
	for c := range scanner.GetLeadingCommentRanges(nodeFactory, text, node.Pos()) {
		raw := text[c.Pos():c.End()]
		value := stripCommentMarkers(raw, c.Kind)
		if thisTagPattern.MatchString(value) {
			return true
		}
	}
	return false
}

// stripCommentMarkers removes `/*`/`*/` from block comments and `//` from
// line comments, matching ESLint's `comment.value` representation.
func stripCommentMarkers(raw string, kind ast.Kind) string {
	switch kind {
	case ast.KindSingleLineCommentTrivia:
		return strings.TrimPrefix(raw, "//")
	case ast.KindMultiLineCommentTrivia:
		v := strings.TrimPrefix(raw, "/*")
		v = strings.TrimSuffix(v, "*/")
		return v
	}
	return raw
}

// isDefaultThisBinding mirrors `astUtils.isDefaultThisBinding` from
// ESLint core — the parent-chain walk that decides whether a function's
// `this` is bound by its surrounding context. Returns true when the
// surrounding context does NOT bind `this` (default binding → global /
// undefined → invalid).
//
// Only `KindParenthesizedExpression` is skipped (ESTree elides parens,
// so this preserves byte-for-byte parity). TypeScript expression wrappers
// (`as`, `satisfies`, `!`) are intentionally NOT skipped: upstream's
// walker has no `TSAsExpression` / `TSSatisfiesExpression` /
// `TSNonNullExpression` case, so it falls through to `default: return
// true` (default binding). Treating them opaquely matches that behavior.
//
// Method / Constructor / Accessor parents never appear when walking from
// a Function*Expression in tsgo: tsgo collapses ESTree's
// `MethodDefinition.value` into the method node itself, so the function
// and its container are the same node and we push for them via the
// dedicated KindMethodDeclaration / KindConstructor / KindGet/SetAccessor
// listeners. The corresponding ESTree branch is therefore unreachable
// here.
func isDefaultThisBinding(node *ast.Node, capIsConstructor bool) bool {
	isAnonymous := !hasOwnFunctionName(node)
	current := node
	for {
		parent := current.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			parent = current.Parent
		}
		if parent == nil {
			return true
		}

		switch parent.Kind {
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			opKind := bin.OperatorToken.Kind
			switch opKind {
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
				// Logical / nullish — transparent (ESLint case "LogicalExpression").
				current = parent
				continue
			case ast.KindEqualsToken:
				// AssignmentExpression: function is the right-hand value.
				if bin.Right != current {
					return true
				}
				left := ast.SkipParentheses(bin.Left)
				if left == nil {
					return true
				}
				if ast.IsPropertyAccessExpression(left) || ast.IsElementAccessExpression(left) {
					// obj.foo = function(){} / obj['foo'] = function(){}
					return false
				}
				if capIsConstructor && isAnonymous && ast.IsIdentifier(left) &&
					startsWithUpperCase(left.AsIdentifier().Text) {
					// Foo = function(){} — assignment to an uppercase variable
					// (anonymous function) is treated as an ES5 constructor.
					return false
				}
				return true
			default:
				return true
			}

		case ast.KindConditionalExpression:
			current = parent
			continue

		case ast.KindReturnStatement:
			// `return function(){}` is transparent ONLY when the surrounding
			// function is invoked immediately (IIFE). Otherwise the returned
			// function escapes to an unknown caller — default binding.
			fn := ast.GetContainingFunction(parent)
			if fn == nil || !isCalleeParenOnly(fn) {
				return true
			}
			// ESLint advances `currentNode = func.parent`, which lands on the
			// IIFE CallExpression (parens are elided in ESTree). In tsgo the
			// function may be wrapped in `ParenthesizedExpression` / TS
			// expression assertions, so walk through those to reach the
			// CallExpression itself.
			callExpr := findEnclosingCall(fn)
			if callExpr == nil {
				return true
			}
			current = callExpr
			continue

		case ast.KindArrowFunction:
			// `(() => function(){})()` — arrow concise body that's itself
			// IIFE'd. Same logic as ReturnStatement: only transparent when
			// the arrow is immediately called.
			af := parent.AsArrowFunction()
			if af.Body != current || !isCalleeParenOnly(parent) {
				return true
			}
			callExpr := findEnclosingCall(parent)
			if callExpr == nil {
				return true
			}
			current = callExpr
			continue

		case ast.KindPropertyAssignment:
			pa := parent.AsPropertyAssignment()
			if pa.Initializer != current {
				return true
			}
			return false

		case ast.KindShorthandPropertyAssignment:
			// `{Foo = function(){}}` destructuring shorthand with default.
			// ObjectAssignmentInitializer plays the same role as
			// AssignmentPattern's right operand in ESTree.
			spa := parent.AsShorthandPropertyAssignment()
			if spa.ObjectAssignmentInitializer != current {
				return true
			}
			name := spa.Name()
			if capIsConstructor && isAnonymous && name != nil && ast.IsIdentifier(name) &&
				startsWithUpperCase(name.AsIdentifier().Text) {
				return false
			}
			return true

		case ast.KindPropertyDeclaration:
			pd := parent.AsPropertyDeclaration()
			if pd.Initializer != current {
				return true
			}
			// ESLint's walker has explicit cases for `Property` /
			// `PropertyDefinition` / `MethodDefinition` but NOT for
			// `AccessorProperty`. tsgo collapses both ESTree kinds onto
			// KindPropertyDeclaration; we re-introduce the distinction here
			// by checking `ModifierFlagsAccessor`. For auto-accessors a
			// function-expression initializer falls through to the walker's
			// default branch (default-bound), matching how upstream's
			// baseRule walker treats `AccessorProperty` parents.
			if parent.ModifierFlags()&ast.ModifierFlagsAccessor != 0 {
				return true
			}
			return false

		case ast.KindVariableDeclaration:
			vd := parent.AsVariableDeclaration()
			if vd.Initializer != current {
				return true
			}
			if capIsConstructor && isAnonymous {
				name := vd.Name()
				if name != nil && ast.IsIdentifier(name) &&
					startsWithUpperCase(name.AsIdentifier().Text) {
					return false
				}
			}
			return true

		case ast.KindParameter:
			pd := parent.AsParameterDeclaration()
			if pd.Initializer != current {
				return true
			}
			if capIsConstructor && isAnonymous {
				name := pd.Name()
				if name != nil && ast.IsIdentifier(name) &&
					startsWithUpperCase(name.AsIdentifier().Text) {
					return false
				}
			}
			return true

		case ast.KindBindingElement:
			be := parent.AsBindingElement()
			if be.Initializer != current {
				return true
			}
			if capIsConstructor && isAnonymous {
				name := be.Name()
				if name != nil && ast.IsIdentifier(name) &&
					startsWithUpperCase(name.AsIdentifier().Text) {
					return false
				}
			}
			return true

		case ast.KindPropertyAccessExpression:
			// `(function(){}).call(obj)` / `.bind(obj)` / `.apply(obj)`.
			pae := parent.AsPropertyAccessExpression()
			if pae.Expression != current {
				return true
			}
			name := pae.Name()
			if name == nil || !ast.IsIdentifier(name) || !isCallApplyBind(name.AsIdentifier().Text) {
				return true
			}
			return !invokesAsCalleeWithNonNullFirstArg(parent)

		case ast.KindElementAccessExpression:
			// `(function(){})['call'](obj)` / `['bind']` / `['apply']`.
			eae := parent.AsElementAccessExpression()
			if eae.Expression != current {
				return true
			}
			methodName, ok := utils.GetStaticExpressionValue(ast.SkipParentheses(eae.ArgumentExpression))
			if !ok || !isCallApplyBind(methodName) {
				return true
			}
			return !invokesAsCalleeWithNonNullFirstArg(parent)

		case ast.KindCallExpression:
			// Function passed as an argument to a known thisArg-accepting
			// callable: `Reflect.apply(fn, ctx, args)`,
			// `Array.from(iter, fn, ctx)`, `arr.forEach(fn, ctx)`, ….
			call := parent.AsCallExpression()
			if call.Arguments == nil {
				return true
			}
			args := call.Arguments.Nodes
			callee := call.Expression

			if utils.IsSpecificMemberAccess(callee, "Reflect", "apply") {
				if len(args) != 3 || args[0] != current {
					return true
				}
				return utils.IsNullOrUndefined(args[1])
			}
			if utils.IsSpecificMemberAccess(callee, "Array", "from") ||
				utils.IsSpecificMemberAccess(callee, "Array", "fromAsync") {
				if len(args) != 3 || args[1] != current {
					return true
				}
				return utils.IsNullOrUndefined(args[2])
			}
			if isMethodWhichHasThisArg(callee) {
				if len(args) != 2 || args[0] != current {
					return true
				}
				return utils.IsNullOrUndefined(args[1])
			}
			return true

		default:
			return true
		}
	}
}

func isCallApplyBind(name string) bool {
	return name == "call" || name == "apply" || name == "bind"
}

// isInsideDecoratorOfMethodLike reports whether `thisNode` is positioned
// inside a `@decorator(...)` expression attached to a method-like class
// member (`KindMethodDeclaration` / `KindConstructor` / `KindGetAccessor` /
// `KindSetAccessor`). Decorators on these members run at class-evaluation
// time, so their `this` resolves to the enclosing scope, NOT the member's
// own implicit `this`. PropertyDeclaration is excluded — upstream's
// wrapper pushes for `PropertyDefinition` / `AccessorProperty` on entry,
// so decorators on fields stay in the field's frame to mirror that.
//
// Arrow functions are walked past transparently because their `this` is
// lexical — `@deco(() => this)` resolves `this` to whatever scope the
// arrow is defined in (the decorator's scope, which is outside the
// member). Non-arrow function-likes and class-member boundaries return
// false: a non-arrow function inside a decorator has its OWN `this`
// (default-bound), independent of the decorator's host.
func isInsideDecoratorOfMethodLike(thisNode *ast.Node) bool {
	current := thisNode.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindDecorator:
			parent := current.Parent
			if parent == nil {
				return false
			}
			switch parent.Kind {
			case ast.KindMethodDeclaration, ast.KindConstructor,
				ast.KindGetAccessor, ast.KindSetAccessor:
				return true
			}
			return false
		case ast.KindArrowFunction:
			// Lexical `this` — keep walking up.
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindPropertyDeclaration,
			ast.KindClassStaticBlockDeclaration:
			return false
		}
		current = current.Parent
	}
	return false
}

// isCalleeParenOnly mirrors ESLint's `isCallee` — `node.parent` (after
// stripping ParenthesizedExpression wrappers, which ESTree elides) must
// be a CallExpression or NewExpression with `node` as the callee. TS
// expression wrappers (`as` / `satisfies` / `!`) are intentionally NOT
// stripped to stay byte-for-byte aligned with upstream's walker, which
// has no case for them and would never treat their parent as a call.
func isCalleeParenOnly(node *ast.Node) bool {
	current := node
	parent := current.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		current = parent
		parent = current.Parent
	}
	if parent == nil {
		return false
	}
	if ast.IsCallExpression(parent) && parent.AsCallExpression().Expression == current {
		return true
	}
	if parent.Kind == ast.KindNewExpression && parent.AsNewExpression().Expression == current {
		return true
	}
	return false
}

// findEnclosingCall walks up from a function-like that has just been
// established as the callee of a CallExpression (via `isCalleeParenOnly`)
// until it reaches that CallExpression. The intermediate hops are
// ParenthesizedExpression wrappers that tsgo preserves and ESTree elides.
// Returns nil if no enclosing CallExpression is found.
func findEnclosingCall(node *ast.Node) *ast.Node {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || !ast.IsCallExpression(parent) {
		return nil
	}
	return parent
}

// invokesAsCalleeWithNonNullFirstArg reports whether `memberAccess` (a
// PropertyAccess or ElementAccess that resolves to `.call`/`.apply`/`.bind`)
// is the callee of a CallExpression whose first argument is a real
// (non-null/undefined) value. Mirrors ESLint's `MemberExpression` branch of
// `isDefaultThisBinding`, including the maybeCalleeNode-via-ChainExpression
// dance that tsgo doesn't need (no ChainExpression wrapper).
func invokesAsCalleeWithNonNullFirstArg(memberAccess *ast.Node) bool {
	callParent := memberAccess.Parent
	for callParent != nil && callParent.Kind == ast.KindParenthesizedExpression {
		callParent = callParent.Parent
	}
	if callParent == nil || !ast.IsCallExpression(callParent) {
		return false
	}
	call := callParent.AsCallExpression()
	if ast.SkipParentheses(call.Expression) != memberAccess {
		return false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) < 1 {
		return false
	}
	return !utils.IsNullOrUndefined(call.Arguments.Nodes[0])
}

// arrayMethodsWithThisArg enumerates the standard Array.prototype methods
// whose second argument is a `thisArg` (matches ESLint's
// `arrayMethodWithThisArgPattern` /^(?:every|filter|find(?:Last)?(?:Index)?|flatMap|forEach|map|some)$/).
var arrayMethodsWithThisArg = []string{
	"every", "filter", "find", "findIndex", "findLast", "findLastIndex",
	"flatMap", "forEach", "map", "some",
}

// isMethodWhichHasThisArg reports whether the callee is a member access
// `<anything>.<name>` where `<name>` is one of the array methods that
// accept a `thisArg` second parameter. Mirrors ESLint's
// `isSpecificMemberAccess(node, null, arrayMethodWithThisArgPattern)`.
func isMethodWhichHasThisArg(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name == nil || !ast.IsIdentifier(name) {
			return false
		}
		return slices.Contains(arrayMethodsWithThisArg, name.AsIdentifier().Text)
	case ast.KindElementAccessExpression:
		argText, ok := utils.GetStaticExpressionValue(
			ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression),
		)
		if !ok {
			return false
		}
		return slices.Contains(arrayMethodsWithThisArg, argText)
	}
	return false
}
