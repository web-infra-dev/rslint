package no_invalid_this

import (
	_ "embed"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_invalid_this.schema.json
var schemaJSON []byte

// NoInvalidThisRule flags `this` keywords in contexts where the value of
// `this` is `undefined` — i.e. wherever the default `this` binding applies
// under strict mode.
//
// NOTE: Unlike ESLint, rslint does not expose `languageOptions.parserOptions`
// (`ecmaFeatures.globalReturn`, `ecmaFeatures.impliedStrict`). ESLint's
// `ecmaFeatures.globalReturn` escape hatch therefore has no rslint equivalent.
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

// EngineOptions configures BuildListeners with the policy point
// that genuinely differs between `no-invalid-this` and
// `@typescript-eslint/no-invalid-this` (see BuildListeners doc comment).
// Everything else — the parent-chain walker, JSDoc/this-param recognition,
// computed-key deferral, method-decorator handling — is identical between
// the two rules and lives here as the single shared implementation.
type EngineOptions struct {
	CapIsConstructor bool
	// FieldFrameScopedToValue decides how much of a class field the field's
	// always-valid frame covers. ESLint core opens it as a code path rooted
	// at the initializer *value* (`onCodePathStart` for the
	// class-field-initializer path, plus `AccessorProperty > *.value`), so
	// the positions visited before the value — decorators, a computed key,
	// and the type annotation — see the enclosing scope instead: true.
	// typescript-eslint's wrapper instead pushes on `PropertyDefinition` /
	// `AccessorProperty` entry, so the field's frame covers the whole
	// member and swallows the report in both positions: false. Method-likes
	// are unaffected: under both rules their decorators resolve to the
	// enclosing scope and their computed keys defer the push.
	FieldFrameScopedToValue bool
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

// CoreEngineOptions returns the ESLint-core policy for the effective language
// options on ctx. The TypeScript wrapper reuses this policy and changes only
// the field-frame boundary that its upstream wrapper adds.
func CoreEngineOptions(ctx rule.RuleContext, opts Options) EngineOptions {
	sourceType := ctx.LanguageOptions.SourceType
	if sourceType == "" {
		// Production contexts are normalized before rule execution. Keep the
		// zero-value context useful to direct unit harnesses by falling back to
		// the parsed source shape, matching this rule's historical behavior.
		if ast.IsExternalModule(ctx.SourceFile) {
			sourceType = "module"
		} else {
			sourceType = "script"
		}
	}
	ecmaVersion := ctx.LanguageOptions.EffectiveECMAVersion()
	return EngineOptions{
		CapIsConstructor: opts.CapIsConstructor,
		TopLevelValid:    sourceType != "module",
		IsStrict: func(fn *ast.Node, sf *ast.SourceFile) bool {
			return isStrictFunction(fn, sf, ecmaVersion, sourceType == "module")
		},
		FieldFrameScopedToValue: true,
	}
}

// BuildListeners is the shared `this`-validity walker behind both
// `no-invalid-this` (this package) and `@typescript-eslint/no-invalid-this`
// (internal/plugins/typescript/rules/no_invalid_this), which wraps this
// function directly rather than duplicating it. The rules disagree only on
// how much of a class field the field's own frame covers (see
// `FieldFrameScopedToValue`).
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

	// enterPropertyDeclaration defers the physical push past a computed key
	// exactly like `enterMethodLike` — but only under
	// `FieldFrameScopedToValue`.
	// Core ESLint gives a class field's *value* position its own
	// implicit-function code path (`PropertyDefinition#value` — see
	// `isDefaultThisBinding`'s doc comment) but does NOT extend that
	// treatment to the key: a computed key (`[this.foo]`) is evaluated at
	// class-definition time in the *enclosing* scope, not the field's own
	// context, so it must see whatever frame was already on the stack. A
	// type annotation is visited after the key; enclosingFrameSkip therefore
	// looks past this physical frame until the initializer begins. The
	// wrapper's `PropertyDefinition()` / `AccessorProperty()` listeners push
	// on entry unconditionally, so there the key is covered too.
	enterPropertyDeclaration := func(node *ast.Node) {
		if eo.FieldFrameScopedToValue && hasComputedKey(node) {
			return
		}
		push(true)
	}

	reportThis := func(node *ast.Node) {
		// A decorator's expression is evaluated at class-definition
		// time, in the scope surrounding the class — so `this` inside
		// one belongs to the enclosing frame, not the decorated
		// member's. `enclosingFrameSkip` counts how many frames on the
		// stack sit between `this` and the frame that governs it.
		idx := len(stack) - 1 - enclosingFrameSkip(node, eo.FieldFrameScopedToValue)
		var valid bool
		if idx < 0 {
			valid = eo.TopLevelValid
		} else {
			valid = stack[idx]
		}
		if !valid {
			ctx.ReportNode(node, msg)
		}
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
				ast.KindGetAccessor, ast.KindSetAccessor:
				push(true)
			case ast.KindPropertyDeclaration:
				// Already pushed on entry unless the field's frame is
				// scoped to the initializer value.
				if eo.FieldFrameScopedToValue {
					push(true)
				}
			}
		},

		// Class static block: own `this` context bound to the class — VALID.
		ast.KindClassStaticBlockDeclaration:                      func(*ast.Node) { push(true) },
		rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): func(*ast.Node) { pop() },

		// Arrow function: lexical `this`. Intentionally NOT registered so
		// the enclosing frame governs.

		ast.KindThisKeyword: reportThis,
		ast.KindThisType: func(node *ast.Node) {
			// typescript-estree represents the operand in `typeof this` as
			// ThisExpression, while tsgo uses ThisType. Ordinary `this` types
			// remain TSThisType upstream and must not be treated as expressions.
			if isTypeQueryRoot(node) {
				reportThis(node)
			}
		},
		ast.KindIdentifier: func(node *ast.Node) {
			// In current tsgo parser shapes, the same authored operand may be
			// represented as an Identifier beneath TypeQuery or at the left root
			// of its qualified entity name.
			if node.AsIdentifier().Text == "this" && isTypeQueryRoot(node) {
				reportThis(node)
			}
		},
	}
}

// isTypeQueryRoot reports whether node is the root entity of a TypeQuery,
// either directly (`typeof this`) or through the left side of one or more
// QualifiedNames (`typeof this.x.y`). A right-side name is merely a property
// segment and must not be treated as the queried `this` expression.
func isTypeQueryRoot(node *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil && current.Parent.Kind == ast.KindQualifiedName {
		qualified := current.Parent.AsQualifiedName()
		if qualified == nil || qualified.Left != current {
			return false
		}
		current = current.Parent
	}
	return current != nil && current.Parent != nil && current.Parent.Kind == ast.KindTypeQuery
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := ParseOptions(options)
	return BuildListeners(ctx, CoreEngineOptions(ctx, opts))
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
	if hasJSDocThisTag(node, comments, sf) {
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
// "use strict", or a class body), OR the function declares its own "use
// strict" directive. Class decorators are evaluated outside their decorated
// class body, so that class boundary is skipped while walking outward; member
// decorators remain inside the class's strict scope.
func isStrictFunction(fn *ast.Node, sf *ast.SourceFile, ecmaVersion int, isModule bool) bool {
	// Modules, classes, and TypeScript namespaces impose strict mode
	// independently of directive prologues and therefore remain strict even
	// when the configured language version is ES3.
	if isModule {
		return true
	}
	skipDecoratedClassBody := false
	for current := fn.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindDecorator {
			host := current.Parent
			if host != nil && ast.IsClassLike(host) {
				skipDecoratedClassBody = true
			}
		}
		if current.Kind == ast.KindModuleDeclaration || current.Kind == ast.KindEnumDeclaration {
			return true
		}
		if ast.IsClassLike(current) {
			if skipDecoratedClassBody {
				skipDecoratedClassBody = false
				continue
			}
			return true
		}
		if (ecmaVersion != 3 || !sf.IsJS()) && ast.IsFunctionLike(current) && !skipDecoratedClassBody {
			body := current.Body()
			if body != nil && body.Kind == ast.KindBlock && utils.HasUseStrictDirective(body) {
				return true
			}
		}
	}

	// Strict-mode directives were introduced in ES5. In an ES3 JavaScript
	// source, only the syntax-imposed contexts handled above can make the
	// function strict. typescript-eslint keeps directives active for TypeScript
	// sources even when languageOptions.ecmaVersion is 3.
	if ecmaVersion == 3 && sf.IsJS() {
		return false
	}
	if utils.HasUseStrictDirective(sf.AsNode()) {
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

// matchesThisTagPattern implements ESLint's `/^[\s*]*@this/mu` against a
// comment value. The whitespace predicate deliberately uses ECMAScript's
// `\s` semantics, which include U+00A0, U+FEFF, and every Space_Separator.
func matchesThisTagPattern(value string) bool {
	for lineStart := 0; lineStart <= len(value); {
		pos := lineStart
		for pos < len(value) {
			r, size := utf8.DecodeRuneInString(value[pos:])
			if r != '*' && !ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				break
			}
			pos += size
		}
		if strings.HasPrefix(value[pos:], "@this") {
			return true
		}

		for lineStart < len(value) {
			r, size := utf8.DecodeRuneInString(value[lineStart:])
			lineStart += size
			if ecmascript.IsLineTerminator(r) {
				break
			}
		}
		if lineStart == len(value) {
			break
		}
	}
	return false
}

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
func hasJSDocThisTag(fn *ast.Node, comments []*ast.CommentRange, sf *ast.SourceFile) bool {
	if hasThisTagInLeadingComments(fn, comments, sf) {
		return true
	}
	if fn.Kind == ast.KindFunctionDeclaration && hasThisTagAfterExportModifiers(fn, comments, sf) {
		return true
	}
	if fn.Kind != ast.KindFunctionExpression {
		return false
	}

	// ESTree discards parentheses, so find the first semantic parent before
	// applying its direct-call exception. TypeScript assertion wrappers remain
	// visible in typescript-estree and must participate in the ancestor walk.
	current := fn
	parent := current.Parent
	for parent != nil && ast.IsOuterExpression(parent, ast.OEKParentheses) {
		current = parent
		parent = current.Parent
	}
	if parent == nil || parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindNewExpression {
		return false
	}

	// ESLint keeps climbing while the intermediate node has no leading
	// comments and is not a function, method, or property boundary. This is
	// intentionally broader than a hand-picked expression list: arrays,
	// unary expressions, nested calls, and other containers are transparent.
	for parent != nil && parent.Kind != ast.KindSourceFile {
		if isJSDocLookupBoundary(parent) {
			// getJSDocComment checks the reached boundary itself unless it is a
			// FunctionDeclaration (or Program, excluded by the loop condition).
			if parent.Kind == ast.KindFunctionDeclaration {
				return false
			}
			_, hasTag := ancestorJSDocSummary(parent, comments, sf)
			return hasTag
		}
		hasComments, hasTag := ancestorJSDocSummary(parent, comments, sf)
		if hasComments {
			return hasTag
		}
		current = parent
		parent = current.Parent
	}
	return false
}

// hasThisTagAfterExportModifiers handles exported declarations, whose ts-go
// range begins at `export` while ESTree's declaration starts after the export
// modifiers. Comments in that gap belong to the function upstream.
func hasThisTagAfterExportModifiers(fn *ast.Node, comments []*ast.CommentRange, sf *ast.SourceFile) bool {
	s := scanner.GetScannerForSourceFile(sf, fn.Pos())
	exportModifierEnd := -1
	for s.Token() != ast.KindEndOfFile && s.TokenStart() < fn.End() {
		switch s.Token() {
		case ast.KindExportKeyword, ast.KindDefaultKeyword:
			exportModifierEnd = s.TokenEnd()
		default:
			if exportModifierEnd < 0 {
				return false
			}
			declarationStart := s.TokenStart()
			text := sf.Text()
			pos := exportModifierEnd
			idx := sort.Search(len(comments), func(i int) bool { return comments[i].Pos() >= pos })
			for i := idx; i < len(comments) && comments[i].Pos() < declarationStart; i++ {
				comment := comments[i]
				if !ecmascript.IsBlank(text[pos:comment.Pos()]) {
					return false
				}
				if matchesThisTagPattern(stripCommentMarkers(text[comment.Pos():comment.End()], comment.Kind)) {
					return true
				}
				pos = comment.End()
			}
			return false
		}
		s.Scan()
	}
	return false
}

// ancestorJSDocSummary mirrors findJSDocComment after getJSDocComment has
// climbed to a declaration or statement. Any leading comment stops the climb,
// but only the closest adjacent `/** ... */` block can carry the JSDoc tag.
func ancestorJSDocSummary(node *ast.Node, comments []*ast.CommentRange, sf *ast.SourceFile) (hasComments, hasTag bool) {
	if node == nil {
		return false, false
	}
	text := sf.Text()
	tokenStart := estreeNodeStart(node, sf)
	idx := sort.Search(len(comments), func(i int) bool { return comments[i].Pos() >= tokenStart })
	if idx == 0 {
		return false, false
	}
	closest := comments[idx-1]
	if closest.End() > tokenStart || !ecmascript.IsBlank(text[closest.End():tokenStart]) {
		return false, false
	}
	if closest.Kind != ast.KindMultiLineCommentTrivia {
		return true, false
	}
	raw := text[closest.Pos():closest.End()]
	if !strings.HasPrefix(raw, "/**") {
		return true, false
	}
	lineMap := sf.ECMALineMap()
	if scanner.ComputeLineOfPosition(lineMap, tokenStart)-scanner.ComputeLineOfPosition(lineMap, closest.End()) > 1 {
		return true, false
	}
	return true, matchesThisTagPattern(stripCommentMarkers(raw, closest.Kind))
}

func isJSDocLookupBoundary(node *ast.Node) bool {
	if ast.IsFunctionLike(node) {
		return true
	}
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindPropertyDeclaration:
		return true
	case ast.KindBindingElement:
		// ESTree object-pattern properties are JSDoc lookup boundaries, but
		// array-pattern AssignmentPattern elements are transparent.
		return node.Parent != nil && node.Parent.Kind == ast.KindObjectBindingPattern
	}
	return false
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
func hasThisTagInLeadingComments(node *ast.Node, comments []*ast.CommentRange, sf *ast.SourceFile) bool {
	_, hasTag := leadingCommentSummary(node, comments, sf)
	return hasTag
}

func leadingCommentSummary(node *ast.Node, comments []*ast.CommentRange, sf *ast.SourceFile) (hasComments, hasTag bool) {
	if node == nil {
		return false, false
	}
	text := sf.Text()
	nodeStart := estreeNodeStart(node, sf)
	pos := node.Pos()
	idx := sort.Search(len(comments), func(i int) bool { return comments[i].Pos() >= pos })
	for i := idx; i < len(comments); i++ {
		c := comments[i]
		if c.Pos() >= nodeStart {
			break
		}
		if !ecmascript.IsBlank(text[pos:c.Pos()]) {
			break
		}
		hasComments = true
		value := stripCommentMarkers(text[c.Pos():c.End()], c.Kind)
		if matchesThisTagPattern(value) {
			hasTag = true
		}
		pos = c.End()
	}
	if !ecmascript.IsBlank(text[pos:nodeStart]) {
		return hasComments, false
	}
	return hasComments, hasTag
}

// estreeNodeStart returns the first token retained in the corresponding ESTree
// node's range. Parentheses are absent from ESTree, so a comment separated from
// the semantic node by `(` is not a leading comment of that node itself.
func estreeNodeStart(node *ast.Node, sf *ast.SourceFile) int {
	if node.Kind == ast.KindArrowFunction {
		// A parenthesized arrow's opening `(` belongs to its ESTree range as
		// the parameter-list token; it is not an erased grouping wrapper.
		return scanner.SkipTrivia(sf.Text(), node.Pos())
	}
	end := node.End()
	s := scanner.GetScannerForSourceFile(sf, node.Pos())
	for s.Token() != ast.KindEndOfFile && s.TokenStart() < end {
		if s.Token() != ast.KindOpenParenToken {
			return s.TokenStart()
		}
		s.Scan()
	}
	return scanner.SkipTrivia(sf.Text(), node.Pos())
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
			// Both regular fields and auto-accessors give their initializer
			// value an always-valid field frame. tsgo collapses both ESTree
			// node kinds onto PropertyDeclaration, so they share this branch.
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
			if isArrayFromMethod(callee) ||
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

// isArrayFromMethod mirrors ESLint's `isArrayFromMethod`: the receiver must
// be an identifier whose name ends in "Array", which includes Array and all
// TypedArray constructors, and the statically-known property must be "from".
func isArrayFromMethod(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil || !ast.IsIdentifier(name) || name.AsIdentifier().Text != "from" {
			return false
		}
		receiver = access.Expression
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		name, ok := utils.GetStaticExpressionValue(ast.SkipParentheses(access.ArgumentExpression))
		if !ok || name != "from" {
			return false
		}
		receiver = access.Expression
	default:
		return false
	}

	receiver = ast.SkipParentheses(receiver)
	return receiver != nil && ast.IsIdentifier(receiver) &&
		strings.HasSuffix(receiver.AsIdentifier().Text, "Array")
}

// enclosingFrameSkip reports how many frames the `this` at `thisNode` must
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
// A core class field's physical frame is already on the stack while tsgo
// visits a non-computed field's type annotation (and after a computed key),
// even though upstream's field-value frame is not logically active there.
// That position costs one skip and remains transparent. A wrapper field takes
// part in either kind of hop only when `FieldFrameScopedToValue` says so;
// otherwise its whole-member frame governs and the walk stops there.
func enclosingFrameSkip(thisNode *ast.Node, fieldFrameScopedToValue bool) int {
	skip := 0
	prev, current := thisNode, thisNode.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindDecorator:
			member := current.Parent
			if member != nil && ast.IsClassLike(member) {
				// A class decorator owns no validity frame. Keep walking so a
				// surrounding member decorator can select its enclosing frame.
				prev, current = member, member.Parent
				continue
			}
			if member == nil || !isDecoratedClassMember(member) {
				return skip
			}
			if member.Kind == ast.KindPropertyDeclaration && !fieldFrameScopedToValue {
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
		case ast.KindPropertyDeclaration:
			if !fieldFrameScopedToValue {
				return skip
			}
			property := current.AsPropertyDeclaration()
			if property != nil && property.Type == prev {
				skip++
				break
			}
			fallthrough
		case ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
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
