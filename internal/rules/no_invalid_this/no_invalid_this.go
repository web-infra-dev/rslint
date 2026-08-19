package no_invalid_this

import (
	_ "embed"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_invalid_this.schema.json
var schemaJSON []byte

// NoInvalidThisRule flags `this` keywords in contexts where the value of
// `this` is `undefined` — i.e. wherever the default `this` binding applies
// under strict mode.
//
// NOTE: Unlike ESLint, rslint does not expose `languageOptions.parserOptions`
// (`ecmaFeatures.globalReturn`, `ecmaFeatures.impliedStrict`) or an
// `ecmaVersion` setting. Consequences:
//   - Top-level `this` validity is derived solely from
//     `ast.IsExternalModule` (presence of import/export): scripts are always
//     valid, ES modules are always invalid. ESLint's `ecmaFeatures.globalReturn`
//     escape hatch (Node.js CommonJS-style top-level `return`) has no rslint
//     equivalent and is therefore unreachable.
//   - Every "use strict" directive is honored regardless of ECMAScript
//     version. ESLint only applies directive-based strict mode from ES5
//     onward (`ecmaVersion: 3` code with a body-level "use strict" stays
//     non-strict); rslint has no such version gate.
//
// https://eslint.org/docs/latest/rules/no-invalid-this
var NoInvalidThisRule = rule.Rule{
	Name:   "no-invalid-this",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
}

// Options is the parsed option set — shared with
// `@typescript-eslint/no-invalid-this`, which reuses ParseOptions directly.
type Options struct {
	// CapIsConstructor: when true (default), a capitalized-name function is
	// treated as an ES5 constructor and its `this` is considered valid.
	CapIsConstructor bool
}

// ParseOptions parses the rule's single object option. Exported so
// `@typescript-eslint/no-invalid-this` — which accepts the identical option
// shape — can reuse it instead of re-implementing the same map access.
func ParseOptions(options []any) Options {
	opts := Options{CapIsConstructor: true}
	if len(options) == 0 {
		return opts
	}
	m, _ := options[0].(map[string]interface{})
	if v, ok := m["capIsConstructor"]; ok {
		if b, ok := v.(bool); ok {
			opts.CapIsConstructor = b
		}
	}
	return opts
}

// EngineOptions configures BuildListeners with the three policy points
// that genuinely differ between `no-invalid-this` and
// `@typescript-eslint/no-invalid-this` (see BuildListeners doc comment).
// Everything else — the parent-chain walker, JSDoc/this-param recognition,
// computed-key deferral, method-decorator handling — is identical between
// the two rules and lives here as the single shared implementation.
type EngineOptions struct {
	CapIsConstructor bool
	// FieldDecoratorUsesEnclosingScope decides which frame `this` inside a
	// decorator attached to a class field belongs to. ESLint core opens the
	// field's frame as a code path rooted at the initializer *value*
	// (`onCodePathStart` for the class-field-initializer path, plus
	// `AccessorProperty > *.value`), so a decorator — visited before the
	// value — sees the enclosing scope: true. typescript-eslint's wrapper
	// instead pushes on `PropertyDefinition` / `AccessorProperty` entry,
	// which covers the decorator too and makes the field's always-valid
	// frame swallow the report: false. Method-like decorators resolve to
	// the enclosing scope under both rules.
	FieldDecoratorUsesEnclosingScope bool
	// TopLevelValid is the validity of `this` when no function/class-member
	// frame is on the stack (i.e. at the top level of the file).
	TopLevelValid bool
	// IsStrict decides whether a FunctionDeclaration/FunctionExpression
	// frame requires the full default-binding computation (true) or is
	// unconditionally valid (false — the sloppy-mode default-this-is-the-
	// global-object outcome). Class members, fields, and static blocks are
	// unaffected by this policy: they are always valid regardless of
	// strict mode (see the `enterMethodLike` / `enterPropertyDeclaration`
	// doc comments below), matching both rules identically.
	IsStrict func(fn *ast.Node, sf *ast.SourceFile) bool
}

// BuildListeners is the shared `this`-validity walker behind both
// `no-invalid-this` (this package) and `@typescript-eslint/no-invalid-this`
// (internal/plugins/typescript/rules/no_invalid_this), which wraps this
// function directly rather than duplicating it. The two rules differ in
// exactly two respects, captured by EngineOptions:
//   - `no-invalid-this` gates function frames on real strict-mode detection
//     and derives top-level validity from `ast.IsExternalModule`.
//   - `@typescript-eslint/no-invalid-this` assumes `sourceType: "module"`
//     (typescript-eslint's RuleTester default), so every function frame is
//     always strict and top-level `this` is always invalid.
//   - The two also disagree on which frame a class field's decorator sees
//     (see `FieldDecoratorUsesEnclosingScope`).
//
// Everything else — the AST shapes recognized, the order branches are
// checked in, computed-key and decorator handling — is identical, since
// ESLint core's own algorithm already natively recognizes the TypeScript
// `this` parameter and `accessor` class fields that typescript-eslint's
// wrapper was originally written to add on top of an older core rule.
func BuildListeners(ctx rule.RuleContext, eo EngineOptions) rule.RuleListeners {
	sf := ctx.SourceFile

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
		push(computeFunctionValid(node, sf, ctx.Comments.All(), eo.CapIsConstructor, eo.IsStrict))
	}

	// enterMethodLike defers push past the computed key (if any). Applies
	// to Method/Constructor/Get/Set. `this` inside a method/constructor/
	// accessor always refers to the instance (or the class itself, for
	// statics) regardless of strict mode, so no strict-mode gate is needed
	// here — only the ES2015-class-is-always-strict outcome, applied
	// unconditionally.
	enterMethodLike := func(node *ast.Node) {
		if hasComputedKey(node) {
			return
		}
		push(true)
	}

	// enterPropertyDeclaration defers the push past a computed key, exactly
	// like `enterMethodLike`. Core ESLint gives a class field's *value*
	// position its own implicit-function code path (`PropertyDefinition#value`
	// — see `isDefaultThisBinding`'s doc comment) but does NOT extend that
	// treatment to the key: a computed key (`[this.foo]`) is evaluated at
	// class-definition time in the *enclosing* scope, not the field's own
	// context, so it must see whatever frame was already on the stack.
	enterPropertyDeclaration := func(node *ast.Node) {
		if hasComputedKey(node) {
			return
		}
		push(true)
	}

	return rule.RuleListeners{
		// Non-arrow function-like containers — push a frame whose validity
		// depends on strict mode, parameter shape, JSDoc, name, and
		// surrounding context.
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
		// entry verbatim with upstream's `PropertyDefinition()` /
		// `AccessorProperty()` listeners.
		ast.KindPropertyDeclaration:                      enterPropertyDeclaration,
		rule.ListenerOnExit(ast.KindPropertyDeclaration): func(*ast.Node) { pop() },

		// Deferred push for computed-key method-likes and property
		// declarations. The matching pop happens unconditionally on the
		// member's own exit listener, so the stack stays balanced
		// regardless of whether the push happened on enter (non-computed)
		// or here (computed).
		rule.ListenerOnExit(ast.KindComputedPropertyName): func(node *ast.Node) {
			parent := node.Parent
			if parent == nil {
				return
			}
			switch parent.Kind {
			case ast.KindMethodDeclaration, ast.KindConstructor,
				ast.KindGetAccessor, ast.KindSetAccessor,
				ast.KindPropertyDeclaration:
				push(true)
			}
		},

		// Class static block: own `this` context bound to the class — VALID.
		ast.KindClassStaticBlockDeclaration:                      func(*ast.Node) { push(true) },
		rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): func(*ast.Node) { pop() },

		// Arrow function: lexical `this`. Intentionally NOT registered so
		// the enclosing frame governs.

		ast.KindThisKeyword: func(node *ast.Node) {
			// A decorator's expression is evaluated at class-definition
			// time, in the scope surrounding the class — so `this` inside
			// one belongs to the enclosing frame, not the decorated
			// member's. `decoratorFrameSkip` counts how many frames on the
			// stack sit between `this` and the frame that governs it.
			idx := len(stack) - 1 - decoratorFrameSkip(node, eo.FieldDecoratorUsesEnclosingScope)
			var valid bool
			if idx < 0 {
				valid = eo.TopLevelValid
			} else {
				valid = stack[idx]
			}
			if !valid {
				ctx.ReportNode(node, msg)
			}
		},
	}
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := ParseOptions(options)
	return BuildListeners(ctx, EngineOptions{
		CapIsConstructor: opts.CapIsConstructor,
		// Top-level `this` refers to the global object in scripts (always
		// valid) and is `undefined` in ES modules (always invalid) —
		// independent of strict mode.
		TopLevelValid:                    !ast.IsExternalModule(ctx.SourceFile),
		IsStrict:                         isStrictFunction,
		FieldDecoratorUsesEnclosingScope: true,
	})
}

// computeFunctionValid produces the `this`-validity flag pushed when a
// FunctionDeclaration / FunctionExpression frame is entered.
//
//  1. Not in strict mode → VALID unconditionally: the default `this`
//     binding in sloppy-mode code is the global object, never `undefined`.
//     This rule "applies only in strict mode" (ESLint docs).
//  2. `this` parameter on the signature → VALID (TypeScript syntax; ESLint
//     core recognizes this natively as of the version this port targets).
//  3. `@this` JSDoc tag attached to the function (or its statement context)
//     → VALID.
//  4. `capIsConstructor: true` AND the function carries an uppercase own name
//     (`function Foo()` / `var x = function Bar()`) → VALID (ES5 constructor).
//  5. Otherwise, walk the parent chain via `isDefaultThisBinding` — VALID
//     iff the surrounding context binds `this` explicitly (method assignment,
//     `.call`/`.apply`/`.bind`, `Reflect.apply`, array-method `thisArg`, …).
func computeFunctionValid(node *ast.Node, sf *ast.SourceFile, comments []*ast.CommentRange, capIsConstructor bool, isStrict func(*ast.Node, *ast.SourceFile) bool) bool {
	if !isStrict(node, sf) {
		return true
	}
	if hasThisParameter(node) {
		return true
	}
	if hasJSDocThisTag(node, comments, sf.Text()) {
		return true
	}
	if capIsConstructor && utils.IsES5Constructor(node) {
		return true
	}
	return !isDefaultThisBinding(node, capIsConstructor)
}

// isStrictFunction reports whether `fn`'s own body scope is strict-mode —
// mirrors ESLint's `sourceCode.getScope(node).isStrict` for a function node.
// True when an enclosing scope is already strict (ES module, ancestor
// "use strict", or a class body — see utils.IsInStrictMode), OR the function
// declares its own "use strict" directive.
func isStrictFunction(fn *ast.Node, sf *ast.SourceFile) bool {
	if utils.IsInStrictMode(fn, sf) {
		return true
	}
	body := fn.Body()
	return body != nil && body.Kind == ast.KindBlock && utils.HasUseStrictDirective(body)
}

// hasThisParameter reports whether the function's parameter list begins with
// (or contains) a `this: T` parameter. tsgo encodes the `this` parameter as
// a ParameterDeclaration whose `name` is an Identifier with text `"this"`,
// which `ast.IsThisParameter` resolves.
func hasThisParameter(node *ast.Node) bool {
	for _, p := range node.Parameters() {
		if ast.IsThisParameter(p) {
			return true
		}
	}
	return false
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
func hasJSDocThisTag(fn *ast.Node, comments []*ast.CommentRange, text string) bool {
	if hasThisTagInLeadingComments(fn, comments, text) {
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
			return hasThisTagInLeadingComments(parent, comments, text)
		case ast.KindVariableDeclaration:
			// Walk through VariableDeclarationList → VariableStatement so the
			// JSDoc on the statement itself is checked. tsgo splits what
			// ESTree models as a flat VariableDeclaration into three nested
			// nodes; the user-visible JSDoc anchor is the outermost.
			grand := parent.Parent
			if grand != nil && grand.Kind == ast.KindVariableDeclarationList {
				vs := grand.Parent
				if vs != nil && vs.Kind == ast.KindVariableStatement {
					return hasThisTagInLeadingComments(vs, comments, text)
				}
			}
			return hasThisTagInLeadingComments(parent, comments, text)
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
			ast.KindPropertyDeclaration, ast.KindBindingElement, ast.KindParameter:
			return hasThisTagInLeadingComments(parent, comments, text)
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

// hasThisTagInLeadingComments reports whether any comment in `node`'s
// leading trivia matches `@this`. `node.Pos()` sits right after the
// previous real token — i.e. at the start of the trivia region owned by
// `node` — so any comments belonging to `node` lie in
// `[node.Pos(), realTokenStart)`, found by scanning *forward* from
// `node.Pos()` and stopping at the first gap that contains non-whitespace
// (real code, meaning the scan has left `node`'s trivia).
//
// `comments` is `ctx.Comments.All()`: the file's full, source-ordered,
// deduplicated comment list. Using it (rather than
// `scanner.GetLeadingCommentRanges` directly) matters here:
// that scanner API only starts collecting a "leading" comment run after a
// line break (or at file position 0), so a same-line comment like
// `foo(/* @this */ function(){})` — attached as *trailing* trivia of the
// preceding `(` token, not *leading* trivia of the function by that
// definition — would be silently missed. `ctx.Comments.All()` already
// merges both leading- and trailing-collected ranges (see
// `CommentStore.All`), so a plain forward scan over it sees every case
// uniformly.
func hasThisTagInLeadingComments(node *ast.Node, comments []*ast.CommentRange, text string) bool {
	if node == nil {
		return false
	}
	pos := node.Pos()
	idx := sort.Search(len(comments), func(i int) bool { return comments[i].Pos() >= pos })
	for i := idx; i < len(comments); i++ {
		c := comments[i]
		if strings.TrimSpace(text[pos:c.Pos()]) != "" {
			break
		}
		value := stripCommentMarkers(text[c.Pos():c.End()], c.Kind)
		if thisTagPattern.MatchString(value) {
			return true
		}
		pos = c.End()
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
			switch {
			case opKind == ast.KindBarBarToken || opKind == ast.KindAmpersandAmpersandToken || opKind == ast.KindQuestionQuestionToken:
				// Logical / nullish — transparent (ESLint case "LogicalExpression").
				current = parent
				continue
			case ast.IsAssignmentOperator(opKind):
				// AssignmentExpression (`=`, `+=`, `&&=`, `||=`, `??=`, …):
				// upstream's walker treats every assignment operator alike —
				// function is the right-hand value.
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
					utils.StartsWithUpperCase(left.AsIdentifier().Text) {
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
				utils.StartsWithUpperCase(name.AsIdentifier().Text) {
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
					utils.StartsWithUpperCase(name.AsIdentifier().Text) {
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
					utils.StartsWithUpperCase(name.AsIdentifier().Text) {
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
					utils.StartsWithUpperCase(name.AsIdentifier().Text) {
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

// decoratorFrameSkip reports how many frames the `this` at `thisNode` must
// look past on the validity stack before reaching the frame that governs
// it. It walks the ancestor chain outward from `thisNode` until it meets
// the construct whose frame is on top of the stack at visit time.
//
// A `@decorator(...)` expression is evaluated at class-evaluation time, in
// the scope surrounding the class, so `this` inside one resolves to the
// enclosing scope rather than to the decorated member. Our listeners push
// the member's frame on entry, which happens before tsgo visits the
// modifier list, so each such decorator hop costs one skip and the walk
// resumes outside the member — a decorator nested inside another
// decorator (`@dec(class I { @dec2(this) m(){} })`) therefore skips both.
//
// Constructs that have NOT pushed a frame at decorator-visit time are
// walked past transparently and cost no skip:
//   - Arrow functions, whose `this` is lexical.
//   - A member reached through its own computed key: the push is deferred
//     to `ComputedPropertyName:exit`, which runs after the key is visited.
//
// Everything else — a non-arrow function-like, a static block, or a member
// reached through its body or initializer — owns the stack top, so the
// walk stops there with whatever has been counted so far.
//
// Field decorators hop only when `FieldDecoratorUsesEnclosingScope` says
// so; otherwise the field's own frame governs and the walk stops.
func decoratorFrameSkip(thisNode *ast.Node, fieldDecoratorUsesEnclosingScope bool) int {
	skip := 0
	prev, current := thisNode, thisNode.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindDecorator:
			member := current.Parent
			if member == nil || !isDecoratedClassMember(member) {
				return skip
			}
			if member.Kind == ast.KindPropertyDeclaration && !fieldDecoratorUsesEnclosingScope {
				return skip
			}
			if !hasComputedKey(member) {
				skip++
			}
			// Resume outside the decorated member: further decorator hops
			// (a nested class inside this decorator) add their own skips.
			prev, current = member, member.Parent
			continue
		case ast.KindArrowFunction:
			// Lexical `this` — keep walking up.
		case ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindPropertyDeclaration:
			if !hasComputedKey(current) || prev != ast.GetNameOfDeclaration(current) {
				return skip
			}
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindClassStaticBlockDeclaration:
			return skip
		}
		prev, current = current, current.Parent
	}
	return skip
}

// isDecoratedClassMember reports whether `node` is one of the class
// members this rule pushes a frame for. A decorator attached to anything
// else (a class declaration, a parameter) leaves the stack top alone.
func isDecoratedClassMember(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindPropertyDeclaration:
		return true
	}
	return false
}

// hasComputedKey is exposed as a package-level helper so the
// `KindThisKeyword` listener can re-query it without re-binding through
// the `run` closure.
func hasComputedKey(node *ast.Node) bool {
	name := ast.GetNameOfDeclaration(node)
	return name != nil && name.Kind == ast.KindComputedPropertyName
}

// isCalleeParenOnly mirrors ESLint's `isCallee` exactly — `node.parent`
// (after stripping ParenthesizedExpression wrappers, which ESTree elides)
// must be a CallExpression with `node` as the callee. NewExpression is
// NOT accepted: `new fn()` returns the new instance, not what `fn`
// returns, so it doesn't propagate `this` through the walker's IIFE
// arms. TS expression wrappers (`as` / `satisfies` / `!`) are also
// intentionally NOT stripped — upstream's walker has no case for them.
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
	return ast.IsCallExpression(parent) && parent.AsCallExpression().Expression == current
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
