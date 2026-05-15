package class_methods_use_this

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ClassMethodsUseThisRule mirrors @typescript-eslint/class-methods-use-this,
// which extends ESLint core's class-methods-use-this with two TS-specific
// options: `ignoreOverrideMethods` and `ignoreClassesThatImplementAnInterface`.
//
// https://typescript-eslint.io/rules/class-methods-use-this
var ClassMethodsUseThisRule = rule.CreateRule(rule.Rule{
	Name: "class-methods-use-this",
	Run:  run,
})

type ignoreClassesMode int

const (
	ignoreClassesOff ignoreClassesMode = iota
	ignoreClassesAll
	ignoreClassesPublicFields
)

type ruleOptions struct {
	enforceForClassFields bool
	exceptMethods         map[string]struct{}
	ignoreClasses         ignoreClassesMode
	ignoreOverrideMethods bool
}

// parseOptions extracts the rule options. Defaults match the upstream
// `defaultOptions`: `enforceForClassFields: true`, all other flags off.
func parseOptions(raw any) ruleOptions {
	opts := ruleOptions{enforceForClassFields: true}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["enforceForClassFields"]; ok {
		if b, ok := v.(bool); ok {
			opts.enforceForClassFields = b
		}
	}
	if v, ok := m["exceptMethods"]; ok {
		if arr, ok := v.([]interface{}); ok {
			opts.exceptMethods = make(map[string]struct{}, len(arr))
			for _, it := range arr {
				if s, ok := it.(string); ok {
					opts.exceptMethods[s] = struct{}{}
				}
			}
		}
	}
	if v, ok := m["ignoreClassesThatImplementAnInterface"]; ok {
		switch t := v.(type) {
		case bool:
			if t {
				opts.ignoreClasses = ignoreClassesAll
			}
		case string:
			if t == "public-fields" {
				opts.ignoreClasses = ignoreClassesPublicFields
			}
		}
	}
	if v, ok := m["ignoreOverrideMethods"]; ok {
		if b, ok := v.(bool); ok {
			opts.ignoreOverrideMethods = b
		}
	}
	return opts
}

type stackEntry struct {
	classNode *ast.Node // ClassDeclaration / ClassExpression — nil when no member
	member    *ast.Node // MethodDeclaration / GetAccessor / SetAccessor / Constructor / PropertyDeclaration — nil when anonymous
	parent    *stackEntry
	usesThis  bool
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := parseOptions(options)
	var stack *stackEntry

	// pushMember pushes a class-member context whose `member` is the given
	// node. Mirrors upstream's pushContext(member) when the member's parent
	// is a ClassBody (i.e., a real class member, not an object literal /
	// type-literal method).
	pushMember := func(member *ast.Node) {
		parent := member.Parent
		if parent != nil && ast.IsClassLike(parent) {
			stack = &stackEntry{classNode: parent, member: member, parent: stack}
			return
		}
		stack = &stackEntry{parent: stack}
	}

	// pushAnonymous pushes a context with no member. Used for nested
	// function-likes that are not class members (e.g. inside a method body),
	// for class static blocks, and for the value-visit slot of a class field
	// (mirroring upstream's `PropertyDefinition > *.key:exit → pushContext()`).
	pushAnonymous := func() {
		stack = &stackEntry{parent: stack}
	}

	popContext := func() *stackEntry {
		old := stack
		if stack != nil {
			stack = stack.parent
		}
		return old
	}

	// classImplementsInterface reports whether the class node has an
	// `implements` heritage clause. Mirrors upstream's
	// `stackContext.class.implements.length > 0`.
	classImplementsInterface := func(classNode *ast.Node) bool {
		if classNode == nil {
			return false
		}
		hc := utils.GetHeritageClauses(classNode)
		if hc == nil {
			return false
		}
		for _, clause := range hc.Nodes {
			if clause == nil {
				continue
			}
			hcNode := clause.AsHeritageClause()
			if hcNode == nil {
				continue
			}
			if hcNode.Token == ast.KindImplementsKeyword && hcNode.Types != nil && len(hcNode.Types.Nodes) > 0 {
				return true
			}
		}
		return false
	}

	// isPublicField mirrors upstream's `isPublicField`: true when the
	// member has no `private` / `protected` accessibility modifier.
	//
	// PrivateIdentifier (`#x`) keys cannot syntactically carry an
	// accessibility modifier in TypeScript, so `isPublicField` always
	// returns `true` for them. Under `ignoreClassesThatImplementAnInterface:
	// 'public-fields'` this means `#`-keyed members of an `implements`-class
	// are *skipped* (treated as public). This is upstream's intentional
	// behaviour — the rule keys off the modifier, not the `#` prefix.
	isPublicField := func(member *ast.Node) bool {
		flags := member.ModifierFlags()
		return flags&(ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected) == 0
	}

	// isComputedKey reports whether the member's property-key is a
	// ComputedPropertyName. Upstream uses `node.computed`; tsgo encodes it
	// via the name node's kind.
	isComputedKey := func(member *ast.Node) bool {
		n := ast.GetNameOfDeclaration(member)
		return n != nil && n.Kind == ast.KindComputedPropertyName
	}

	// memberKey returns the canonical key used to match against
	// `exceptMethods`. Mirrors upstream's `(hashIfNeeded) + getStaticMemberAccessValue(node)`:
	// PrivateIdentifier text already carries the `#` prefix in tsgo, so no
	// additional prefixing is needed. Returns ("", false) when the key is
	// not statically resolvable (handled by callers as "no match").
	memberKey := func(member *ast.Node) (string, bool) {
		n := ast.GetNameOfDeclaration(member)
		if n == nil {
			return "", false
		}
		if n.Kind == ast.KindPrivateIdentifier {
			return n.AsPrivateIdentifier().Text, true
		}
		return utils.GetStaticPropertyName(n)
	}

	// isIncludedInstanceMethod mirrors upstream's predicate of the same
	// name. Order matches upstream so the early-out cases short-circuit
	// before the exceptMethods set lookup.
	isIncludedInstanceMethod := func(member *ast.Node) bool {
		if member == nil {
			return false
		}
		// static members and constructors are exempt.
		if ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			return false
		}
		if member.Kind == ast.KindConstructor {
			return false
		}
		// Class fields (regular + auto-accessor) only participate when
		// `enforceForClassFields` is on (default true). Both shapes land
		// on KindPropertyDeclaration in tsgo — the `accessor` keyword is
		// modeled as `ModifierFlagsAccessor` on a PropertyDeclaration.
		if member.Kind == ast.KindPropertyDeclaration && !opts.enforceForClassFields {
			return false
		}
		// Computed keys: always included, regardless of `exceptMethods`
		// (upstream's `if (node.computed || exceptMethods.size === 0) return true`).
		if isComputedKey(member) {
			return true
		}
		if len(opts.exceptMethods) == 0 {
			return true
		}
		name, ok := memberKey(member)
		if !ok {
			return true
		}
		_, found := opts.exceptMethods[name]
		return !found
	}

	// classFieldOfFunctionLike returns the surrounding PropertyDeclaration
	// when `node` is the initializer of a class field, walking through any
	// ParenthesizedExpression wrappers. ESTree elides parentheses, so
	// upstream's `PropertyDefinition > ArrowFunctionExpression.value` selector
	// matches whether or not the arrow is paren-wrapped; tsgo preserves the
	// parens, so we recover the same shape via `ast.WalkUpParenthesizedExpressions`.
	classFieldOfFunctionLike := func(node *ast.Node) *ast.Node {
		parent := node.Parent
		if parent == nil {
			return nil
		}
		// WalkUpParenthesizedExpressions advances past ParenthesizedExpression
		// ancestors; passes through unchanged when `parent` is not itself a
		// paren (covering the common, non-wrapped case).
		parent = ast.WalkUpParenthesizedExpressions(parent)
		if parent != nil && parent.Kind == ast.KindPropertyDeclaration {
			return parent
		}
		return nil
	}

	// exitFunction pops the current stack frame and, if it represents a
	// reportable class member that did not use `this`/`super`, emits the
	// diagnostic. Mirrors upstream's `exitFunction`.
	exitFunction := func(node *ast.Node) {
		frame := popContext()
		if frame == nil || frame.member == nil || frame.usesThis {
			return
		}
		if opts.ignoreOverrideMethods && frame.member.ModifierFlags()&ast.ModifierFlagsOverride != 0 {
			return
		}
		if opts.ignoreClasses != ignoreClassesOff && classImplementsInterface(frame.classNode) {
			switch opts.ignoreClasses {
			case ignoreClassesAll:
				return
			case ignoreClassesPublicFields:
				if isPublicField(frame.member) {
					return
				}
			}
		}
		if !isIncludedInstanceMethod(frame.member) {
			return
		}

		// Class-field arrows / function expressions are classified as
		// "method" by ESLint v9's getFunctionNameWithKind (parent.value === node
		// && parent.type === PropertyDefinition/AccessorProperty branch).
		// rslint's shared helpers retain the function-kind tokens
		// ("arrow function" / "function") and key off the immediate parent
		// for head-loc; both need rewriting for class-field initializers,
		// including those wrapped in one or more ParenthesizedExpressions
		// (tsgo preserves what ESTree elides).
		var name string
		var loc core.TextRange
		if field := classFieldOfFunctionLike(node); field != nil &&
			(node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression) {
			name = classFieldFunctionDisplayName(field, node)
			if node.Parent != nil && node.Parent.Kind == ast.KindPropertyDeclaration {
				// Direct child of PropertyDeclaration — the shared helper
				// already handles this shape correctly.
				loc = utils.GetFunctionHeadLoc(ctx.SourceFile, node)
			} else {
				// Paren-wrapped: reconstruct the upstream head loc as
				// "<field-start>...<function's own open paren>".
				loc = classFieldHeadLocAcrossParens(ctx.SourceFile, field, node)
			}
		} else {
			name = utils.GetFunctionNameWithKind(node)
			loc = utils.GetFunctionHeadLoc(ctx.SourceFile, node)
		}
		ctx.ReportRange(
			loc,
			rule.RuleMessage{
				Id:          "missingThis",
				Description: fmt.Sprintf("Expected 'this' to be used by class %s.", name),
			},
		)
	}

	// enterClassLikeMember handles direct class-body function-likes:
	// MethodDeclaration, GetAccessor, SetAccessor, Constructor. Pushes a
	// member context only when the parent is a class (object-literal /
	// type-literal members fall into the anonymous bucket).
	//
	// Bodyless members (`abstract foo(): void`, overload signatures,
	// ambient declarations) have Body() == nil; upstream's ESTree
	// representation routes those through TSAbstractMethodDefinition /
	// TSDeclareMethod nodes so the FunctionExpression listener never sees
	// them, and tsgo collapses them onto the same kind. Match upstream by
	// pushing an anonymous frame so the matching exit pop is balanced but
	// never reports.
	//
	// Computed-key members defer the push until ComputedPropertyName:exit
	// so `this` inside `[this.foo]() {}` attributes to the enclosing scope,
	// not the method itself. Mirrors upstream's effective traversal order:
	// in ESTree the key visits BEFORE pushContext(member) because pushContext
	// is invoked from FunctionExpression entry, after the MethodDefinition's
	// key has already been visited.
	//
	// The computed-key deferral also applies to *bodyless* members:
	// `abstract [this.foo](): void` must let `this` in the computed key flow
	// to the enclosing scope, not be eaten by the bodyless anonymous frame.
	// The matching ComputedPropertyName:exit branch handles bodyless and
	// bodied identically — pushing anonymous when Body() == nil, member when
	// non-nil — keeping the stack balanced against the unconditional pop on
	// the member's exit listener.
	enterClassLikeMember := func(node *ast.Node) {
		if name := ast.GetNameOfDeclaration(node); name != nil && name.Kind == ast.KindComputedPropertyName {
			// Defer push to ComputedPropertyName:exit.
			return
		}
		if node.Body() == nil {
			pushAnonymous()
			return
		}
		pushMember(node)
	}

	// enterFreestandingFunction handles FunctionExpression and ArrowFunction
	// occurrences. Per upstream:
	//   - FunctionExpression: enterFunction unconditionally — anonymous push
	//     unless parent is a class field with the function as its initializer.
	//   - ArrowFunction: NO listener unless the arrow is a class-field
	//     initializer. ESLint's selectors
	//     `PropertyDefinition > ArrowFunctionExpression.value` and
	//     `AccessorProperty > ArrowFunctionExpression.value` only match
	//     those shapes; arrows nested inside method bodies inherit the
	//     enclosing `this` instead of getting their own frame.
	enterFreestandingFunction := func(node *ast.Node) {
		field := classFieldOfFunctionLike(node)
		if node.Kind == ast.KindArrowFunction {
			if field == nil {
				// Arrow inside a method body / variable initializer /
				// argument: inherits enclosing `this`, no frame push.
				return
			}
			pushMember(field)
			return
		}
		// FunctionExpression.
		if field != nil {
			pushMember(field)
			return
		}
		pushAnonymous()
	}

	exitFreestandingFunction := func(node *ast.Node) {
		if node.Kind == ast.KindArrowFunction {
			if classFieldOfFunctionLike(node) == nil {
				return
			}
		}
		exitFunction(node)
	}

	markUsesThis := func(*ast.Node) {
		if stack != nil {
			stack.usesThis = true
		}
	}

	return rule.RuleListeners{
		// Function declarations always carry their own `this` context but
		// are never reportable members — push anonymous, pop on exit.
		ast.KindFunctionDeclaration:                      func(*ast.Node) { pushAnonymous() },
		rule.ListenerOnExit(ast.KindFunctionDeclaration): func(*ast.Node) { popContext() },

		ast.KindFunctionExpression:                      enterFreestandingFunction,
		rule.ListenerOnExit(ast.KindFunctionExpression): exitFreestandingFunction,

		ast.KindArrowFunction:                      enterFreestandingFunction,
		rule.ListenerOnExit(ast.KindArrowFunction): exitFreestandingFunction,

		ast.KindMethodDeclaration:                      enterClassLikeMember,
		rule.ListenerOnExit(ast.KindMethodDeclaration): exitFunction,
		ast.KindGetAccessor:                            enterClassLikeMember,
		rule.ListenerOnExit(ast.KindGetAccessor):       exitFunction,
		ast.KindSetAccessor:                            enterClassLikeMember,
		rule.ListenerOnExit(ast.KindSetAccessor):       exitFunction,
		ast.KindConstructor:                            enterClassLikeMember,
		rule.ListenerOnExit(ast.KindConstructor):       exitFunction,

		// Class field key/value scope split. Upstream:
		//   `PropertyDefinition > *.key:exit` → pushContext()
		//   `PropertyDefinition:exit`         → popContext()
		// This anonymous frame catches `this`/`super` that appear in the
		// field's value position (e.g. `class C { x = this.y }`) without
		// charging the enclosing method's frame. For computed keys
		// (`[this.expr]`), the push must happen AFTER the key is visited so
		// `this` inside the key flows to the enclosing scope — that's why
		// non-computed keys push on enter, but computed keys defer the push
		// to ComputedPropertyName:exit.
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			if isComputedKey(node) {
				return
			}
			pushAnonymous()
		},
		rule.ListenerOnExit(ast.KindPropertyDeclaration): func(*ast.Node) { popContext() },
		rule.ListenerOnExit(ast.KindComputedPropertyName): func(node *ast.Node) {
			parent := node.Parent
			if parent == nil {
				return
			}
			switch parent.Kind {
			case ast.KindPropertyDeclaration:
				pushAnonymous()
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				// Deferred push for computed-key class members. Bodyless
				// members (abstract / overload signatures) get an anonymous
				// frame so the matching exit pop stays balanced without
				// reporting; bodied members get a real member frame.
				if parent.Body() == nil {
					pushAnonymous()
				} else {
					pushMember(parent)
				}
			}
		},

		// Static blocks have their own `this`; isolate them.
		ast.KindClassStaticBlockDeclaration:                      func(*ast.Node) { pushAnonymous() },
		rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): func(*ast.Node) { popContext() },

		ast.KindThisKeyword:  markUsesThis,
		ast.KindSuperKeyword: markUsesThis,
	}
}

// classFieldFunctionDisplayName produces ESLint's diagnostic name for a
// class-field arrow / function-expression initializer (`class C { foo = () => {} }`,
// including paren-wrapped variants `class C { foo = (() => {}); }`).
// Upstream's `getFunctionNameWithKind` classifies these as `method`, optionally
// prefixed by `static` / `private` / `async` / `generator` and suffixed with
// the property key in single quotes. rslint's shared `GetFunctionNameWithKind`
// retains the function-kind tokens ("arrow function" / "function") because
// that's what other consumers expect; reproducing upstream's relabeling here
// keeps the public message text aligned without disturbing the shared helper.
func classFieldFunctionDisplayName(field, node *ast.Node) string {
	var tokens []string
	flags := field.ModifierFlags()
	if flags&ast.ModifierFlagsStatic != 0 {
		tokens = append(tokens, "static")
	}
	keyNode := field.Name()
	if keyNode != nil && keyNode.Kind == ast.KindPrivateIdentifier {
		tokens = append(tokens, "private")
	}
	fnFlags := ast.GetFunctionFlags(node)
	if fnFlags&ast.FunctionFlagsAsync != 0 {
		tokens = append(tokens, "async")
	}
	if fnFlags&ast.FunctionFlagsGenerator != 0 {
		tokens = append(tokens, "generator")
	}
	tokens = append(tokens, "method")
	if name := propertyDisplayName(keyNode); name != "" {
		tokens = append(tokens, fmt.Sprintf("'%s'", name))
	}
	return strings.Join(tokens, " ")
}

// classFieldHeadLocAcrossParens reconstructs upstream's `getFunctionHeadLoc`
// output for a paren-wrapped class-field initializer:
//
//	class C { foo = (() => {}); }   // upstream: "foo = " head ending at inner '('
//
// Upstream's range runs from the PropertyDefinition's start (after decorators)
// to the function's own open paren — except for parenless single-parameter
// arrows, where it ends at whatever token immediately precedes the parameter
// identifier (the outer wrapper '(' when paren-wrapped). rslint's shared
// `GetFunctionHeadLoc` inspects only `node.Parent`, so when the immediate
// parent is `ParenthesizedExpression` the existing helper falls through to
// the default arrow case ("just the `=>` token") and we lose the field
// context. This local helper mirrors ESLint's `getOpeningParenOfParams`
// branches exactly:
//
//   - Arrow with `params.length === 1`: peek the first arrow-owned token.
//     If it's '(' (parens-form like `(x) => …`), end at that '('.
//     Otherwise (parenless `x => …`), end at the trimmed start of the
//     immediate ParenthesizedExpression parent (the wrapping '(').
//   - Arrow with 0 or 2+ params: scan for the first '(' between the
//     arrow's start and its body. Falls back to the `=>` token when no
//     '(' appears (shouldn't happen for these param counts; defensive).
//   - FunctionExpression: scan for the first '(' between the function's
//     start and its body — this is always the parameter list's '(',
//     because the outer wrapper '(' sits before `node.Pos()`.
//
// The field start follows `nodeStartSkippingDecorators` semantics (which the
// shared `GetFunctionHeadLoc` uses for the non-paren-wrapped case): skip past
// any leading `@decorator` tokens so the head range matches ESLint's
// `PropertyDefinition.loc.start` (decorators sit outside that range in
// ESTree, but are part of `field.Pos()` in tsgo).
func classFieldHeadLocAcrossParens(sf *ast.SourceFile, field, node *ast.Node) core.TextRange {
	start := fieldStartAfterDecorators(sf, field)

	endLimit := node.End()
	if body := node.Body(); body != nil {
		endLimit = body.Pos()
	}

	if node.Kind == ast.KindArrowFunction {
		af := node.AsArrowFunction()
		params := af.Parameters
		if params != nil && len(params.Nodes) == 1 {
			// ESLint's special path for single-parameter arrows.
			firstToken := scanner.GetScannerForSourceFile(sf, node.Pos())
			if firstToken.Token() == ast.KindOpenParenToken {
				return core.NewTextRange(start, firstToken.TokenStart())
			}
			// Parenless `x => …`: token immediately before the param is the
			// outer wrapper `(`. Use the immediate ParenthesizedExpression
			// parent's trimmed Pos() (which is the position of that '(').
			if node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
				return core.NewTextRange(start, utils.TrimNodeTextRange(sf, node.Parent).Pos())
			}
			// Defensive fallback — this helper is only invoked when paren-wrapped,
			// so the branch above should always succeed.
			return core.NewTextRange(start, utils.TrimNodeTextRange(sf, params.Nodes[0]).Pos())
		}
		if pos := firstOpenParenPos(sf, node.Pos(), endLimit); pos >= 0 {
			return core.NewTextRange(start, pos)
		}
		// 0 / 2+ params with no '(' before the body is impossible in valid
		// TS; keep an `=>`-positioned fallback to stay total.
		return core.NewTextRange(start, af.EqualsGreaterThanToken.Pos())
	}

	// FunctionExpression: outer paren is before node.Pos(); first '(' from
	// `function` keyword to body is always the parameter list's '('.
	if pos := firstOpenParenPos(sf, node.Pos(), endLimit); pos >= 0 {
		return core.NewTextRange(start, pos)
	}
	return core.NewTextRange(start, endLimit)
}

// fieldStartAfterDecorators mirrors the private `nodeStartSkippingDecorators`
// helper used by `GetFunctionHeadLoc` for the non-paren-wrapped case: walk
// the PropertyDeclaration's modifiers, find the last `@decorator`, and
// return the position of the next token after it. Without this, an `@dec`
// field reports the head as starting at the `@`, while ESLint reports
// starting after the decorators.
func fieldStartAfterDecorators(sf *ast.SourceFile, field *ast.Node) int {
	fallback := utils.TrimNodeTextRange(sf, field).Pos()
	mods := field.Modifiers()
	if mods == nil || len(mods.Nodes) == 0 {
		return fallback
	}
	var lastDecoratorEnd int
	for _, mod := range mods.Nodes {
		if mod.Kind == ast.KindDecorator && mod.End() > lastDecoratorEnd {
			lastDecoratorEnd = mod.End()
		}
	}
	if lastDecoratorEnd == 0 {
		return fallback
	}
	return scanner.GetRangeOfTokenAtPosition(sf, lastDecoratorEnd).Pos()
}

// firstOpenParenPos scans for the first `(` token in [start, end). Returns
// -1 if none. Bounded so it never reads through the function body.
func firstOpenParenPos(sf *ast.SourceFile, start, end int) int {
	s := scanner.GetScannerForSourceFile(sf, start)
	for s.TokenStart() < end {
		if s.Token() == ast.KindOpenParenToken {
			return s.TokenStart()
		}
		s.Scan()
	}
	return -1
}

// propertyDisplayName resolves a property-name node to the string ESLint
// would emit inside single quotes (or empty when the key is not statically
// resolvable). Mirrors `GetStaticPropertyName` but also retains the leading
// `#` of a PrivateIdentifier.
func propertyDisplayName(name *ast.Node) string {
	if name == nil {
		return ""
	}
	if name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text
	}
	if name.Kind == ast.KindPrivateIdentifier {
		return name.AsPrivateIdentifier().Text
	}
	if s, ok := utils.GetStaticPropertyName(name); ok {
		return s
	}
	return ""
}
