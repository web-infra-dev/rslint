package class_methods_use_this

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed class_methods_use_this.schema.json
var schemaJSON []byte

// ClassMethodsUseThisRule enforces that class instance methods use this or
// super. ESLint core also exposes TypeScript-aware options when the configured
// parser produces override modifiers and implements clauses.
// https://eslint.org/docs/latest/rules/class-methods-use-this
var ClassMethodsUseThisRule = rule.Rule{
	Name:   "class-methods-use-this",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
}

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
func parseOptions(raw []any) ruleOptions {
	opts := ruleOptions{enforceForClassFields: true}
	if len(raw) == 0 {
		return opts
	}
	m, _ := raw[0].(map[string]interface{})

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
	if v, ok := m["ignoreClassesWithImplements"].(string); ok {
		switch v {
		case "all":
			opts.ignoreClasses = ignoreClassesAll
		case "public-fields":
			opts.ignoreClasses = ignoreClassesPublicFields
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

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)
	var stack *stackEntry
	headLocator := utils.NewFunctionHeadRangeLocator(ctx.SourceFile)

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

	lastDecorator := func(node *ast.Node) *ast.Node {
		if node == nil || node.Modifiers() == nil {
			return nil
		}
		var last *ast.Node
		for _, modifier := range node.Modifiers().Nodes {
			if modifier.Kind == ast.KindDecorator {
				last = modifier
			}
		}
		return last
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
			if clause == nil || utils.IsJSDocSyntaxNode(clause) {
				continue
			}
			hcNode := clause.AsHeritageClause()
			if hcNode == nil {
				continue
			}
			if hcNode.Token == ast.KindImplementsKeyword && hcNode.Types != nil {
				for _, heritageType := range hcNode.Types.Nodes {
					if !utils.IsJSDocSyntaxNode(heritageType) {
						return true
					}
				}
			}
		}
		return false
	}

	// isPublicField mirrors ESLint core's public-fields arm: private names and
	// members with private/protected accessibility remain enforced.
	isPublicField := func(member *ast.Node) bool {
		name := ast.GetNameOfDeclaration(member)
		if name != nil && name.Kind == ast.KindPrivateIdentifier {
			return false
		}
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

	isAbstractProperty := func(member *ast.Node) bool {
		return member.Kind == ast.KindPropertyDeclaration &&
			ast.HasSyntacticModifier(member, ast.ModifierFlagsAbstract)
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
	// when `node` is the initializer of a class field, walking through wrappers
	// that ESTree elides. In addition to parentheses, tsgo inserts assertion
	// wrappers for JSDoc @type and @satisfies casts in JavaScript, while ESTree
	// retains those casts only as comments. Therefore upstream's
	// `PropertyDefinition > ArrowFunctionExpression.value` selector
	// still sees the function directly under the field in all of these forms.
	classFieldOfFunctionLike := func(node *ast.Node) *ast.Node {
		parent := utils.ESTreeParent(node)
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
		if opts.ignoreOverrideMethods && utils.ESTreeModifierFlags(frame.member)&ast.ModifierFlagsOverride != 0 {
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

		name := utils.GetFunctionNameWithKindCore(node)
		loc := headLocator.Range(node)
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
	// leaving the enclosing frame active for the complete bodyless node.
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
	// The matching ComputedPropertyName:exit branch pushes a frame only for a
	// bodied member; bodyless members never push or pop a frame.
	pushClassLikeMember := func(node *ast.Node) {
		if node.Body() == nil {
			return
		}
		pushMember(node)
	}

	exitClassLikeMember := func(node *ast.Node) {
		if node.Body() != nil {
			exitFunction(node)
		}
	}

	enterClassLikeMember := func(node *ast.Node) {
		if name := ast.GetNameOfDeclaration(node); name != nil && name.Kind == ast.KindComputedPropertyName {
			// Defer push to ComputedPropertyName:exit.
			return
		}
		if lastDecorator(node) != nil {
			// ESTree enters the member's FunctionExpression only after visiting
			// decorators. Defer the frame until the final decorator exits.
			return
		}
		pushClassLikeMember(node)
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

	markTypeQueryThis := func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text != "this" {
			return
		}

		for node.Parent != nil {
			parent := node.Parent
			switch parent.Kind {
			case ast.KindTypeQuery:
				markUsesThis(node)
				return
			case ast.KindQualifiedName:
				if parent.AsQualifiedName().Left != node {
					return
				}
			case ast.KindPropertyAccessExpression:
				if parent.AsPropertyAccessExpression().Expression != node {
					return
				}
			default:
				return
			}
			node = parent
		}
	}

	return rule.RuleListeners{
		// Function declarations with bodies carry their own `this` context but
		// are never reportable members. Bodyless TypeScript declarations map to
		// TSDeclareFunction upstream and therefore leave the enclosing frame active.
		ast.KindFunctionDeclaration: func(node *ast.Node) {
			if node.Body() != nil {
				pushAnonymous()
			}
		},
		rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) {
			if node.Body() != nil {
				popContext()
			}
		},

		ast.KindFunctionExpression:                      enterFreestandingFunction,
		rule.ListenerOnExit(ast.KindFunctionExpression): exitFreestandingFunction,

		ast.KindArrowFunction:                      enterFreestandingFunction,
		rule.ListenerOnExit(ast.KindArrowFunction): exitFreestandingFunction,

		ast.KindMethodDeclaration:                      enterClassLikeMember,
		rule.ListenerOnExit(ast.KindMethodDeclaration): exitClassLikeMember,
		ast.KindGetAccessor:                            enterClassLikeMember,
		rule.ListenerOnExit(ast.KindGetAccessor):       exitClassLikeMember,
		ast.KindSetAccessor:                            enterClassLikeMember,
		rule.ListenerOnExit(ast.KindSetAccessor):       exitClassLikeMember,
		ast.KindConstructor:                            enterClassLikeMember,
		rule.ListenerOnExit(ast.KindConstructor):       exitClassLikeMember,

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
			if isAbstractProperty(node) || isComputedKey(node) || lastDecorator(node) != nil {
				return
			}
			pushAnonymous()
		},
		rule.ListenerOnExit(ast.KindPropertyDeclaration): func(node *ast.Node) {
			if !isAbstractProperty(node) {
				popContext()
			}
		},
		rule.ListenerOnExit(ast.KindComputedPropertyName): func(node *ast.Node) {
			parent := node.Parent
			if parent == nil {
				return
			}
			switch parent.Kind {
			case ast.KindPropertyDeclaration:
				if !isAbstractProperty(parent) {
					pushAnonymous()
				}
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				// Deferred push for computed-key class members. Bodyless
				// members never enter the upstream function listener.
				if parent.Body() != nil {
					pushMember(parent)
				}
			}
		},
		rule.ListenerOnExit(ast.KindDecorator): func(node *ast.Node) {
			parent := node.Parent
			if parent == nil || lastDecorator(parent) != node || isComputedKey(parent) {
				return
			}
			switch parent.Kind {
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
				pushClassLikeMember(parent)
			case ast.KindPropertyDeclaration:
				if !isAbstractProperty(parent) {
					pushAnonymous()
				}
			}
		},

		// Static blocks have their own `this`; isolate them.
		ast.KindClassStaticBlockDeclaration:                      func(*ast.Node) { pushAnonymous() },
		rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): func(*ast.Node) { popContext() },

		ast.KindThisKeyword:  markUsesThis,
		ast.KindSuperKeyword: markUsesThis,
		ast.KindThisType:     markTypeQueryThis,
		ast.KindIdentifier:   markTypeQueryThis,
	}
}
