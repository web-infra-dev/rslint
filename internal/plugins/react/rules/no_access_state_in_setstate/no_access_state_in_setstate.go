package no_access_state_in_setstate

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// memberPropertyName returns the property-side name of a PropertyAccessExpression
// the way ESLint sees it via `property.name`: an Identifier's text, or a
// PrivateIdentifier's text with the leading `#` stripped (per ESTree spec).
// Any other shape (e.g. ElementAccessExpression caller, null name) yields ""
// so callers can compare against a literal name without false matches.
func memberPropertyName(pa *ast.PropertyAccessExpression) string {
	if pa == nil {
		return ""
	}
	name := pa.Name()
	if name == nil {
		return ""
	}
	switch name.Kind {
	case ast.KindIdentifier:
		return name.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(name.AsPrivateIdentifier().Text, "#")
	}
	return ""
}

// isThisReceiver reports whether `expr` — with any ParenthesizedExpression
// wrappers skipped — is the `this` keyword. Mirrors ESTree's post-paren-strip
// `object.type === 'ThisExpression'` check.
func isThisReceiver(expr *ast.Node) bool {
	return expr != nil && ast.SkipParentheses(expr).Kind == ast.KindThisKeyword
}

// isSetStateCall matches `this.setState(...)` (and `this.#setState(...)`),
// replicating upstream's `callee.property.name === 'setState' &&
// callee.object.type === 'ThisExpression'`. Bracket-access
// (`this['setState']`) is excluded: upstream's `property.name` check has no
// analog for Literal keys. Parentheses around `this` are stripped to emulate
// ESTree's paren-stripping.
func isSetStateCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pa := callee.AsPropertyAccessExpression()
	if !isThisReceiver(pa.Expression) {
		return false
	}
	return memberPropertyName(pa) == "setState"
}

// isThisStateMember matches the `this.state` AST pattern ESLint detects via
// `property.name === 'state' && object.type === 'ThisExpression'`. Bracket
// access (`this['state']`) is excluded for the same reason as isSetStateCall.
func isThisStateMember(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pa := node.AsPropertyAccessExpression()
	if !isThisReceiver(pa.Expression) {
		return false
	}
	return memberPropertyName(pa) == "state"
}

// isFirstArgumentInSetStateCall mirrors upstream's helper of the same name.
// Returns true when `current` is a setState call AND the ancestor chain from
// `node` reaches `current.arguments[0]`.
func isFirstArgumentInSetStateCall(current, node *ast.Node) bool {
	if !isSetStateCall(current) {
		return false
	}
	for node != nil && node.Parent != current {
		node = node.Parent
	}
	if node == nil {
		return false
	}
	call := current.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	return call.Arguments.Nodes[0] == node
}

// containerKind captures which upstream branch a "walk stopper" corresponds to.
// Both branches populate the `methods` array, but only classMethod propagates
// transitively through the CallExpression listener — upstream's propagation
// loop is gated on `current.type === 'MethodDefinition'`, which is the ESTree
// class-body form only.
type containerKind int

const (
	containerNone      containerKind = iota
	containerClass                   // ESTree MethodDefinition (class-body method/accessor/constructor)
	containerObjectLit               // ESTree Property with FunctionExpression value (object-literal method or `foo: function(){}` form)
)

// methodContainerKind reports which "function body attached to a named key"
// shape `node` represents — the generalized form of upstream's MemberExpression
// walk stoppers. We purposefully cover every tsgo shape that maps onto ESTree's
// MethodDefinition or Property-with-FE-value so new AST forms (object-literal
// getter / setter, shorthand method, arrow in property value — see below) are
// tracked consistently instead of case-by-case.
//
// Classification:
//
//   - classMethod: function-like node whose direct parent is a ClassLike node
//     (ClassDeclaration / ClassExpression). Includes MethodDeclaration,
//     GetAccessor, SetAccessor, Constructor — all four ESTree MethodDefinition
//     kinds. PropertyDeclaration (class field, including arrow-function fields)
//     is NOT included: ESTree renders it as PropertyDefinition, which upstream's
//     walk does not stop at.
//
//   - objectLit: function-like value attached to an object-literal key.
//     Covers three tsgo shapes, all of which ESTree flattens into
//     Property-with-FE-value:
//     1. MethodDeclaration / GetAccessor / SetAccessor whose parent is
//     ObjectLiteralExpression (shorthand: `{ foo() {...} }`, `{ get foo() {...} }`).
//     2. FunctionExpression whose parent is PropertyAssignment
//     (`{ foo: function() {...} }`).
//     Arrow functions as property values (`{ foo: () => {...} }`) are
//     intentionally NOT tracked — upstream gates on `FunctionExpression`,
//     which excludes ArrowFunctionExpression.
func methodContainerKind(node *ast.Node) containerKind {
	if node == nil || node.Parent == nil {
		return containerNone
	}
	parent := node.Parent
	// Class-body member — ESTree MethodDefinition.
	if ast.IsClassLike(parent) {
		switch node.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
			return containerClass
		}
		return containerNone
	}
	// Object-literal shorthand method / accessor.
	if parent.Kind == ast.KindObjectLiteralExpression && ast.IsMethodOrAccessor(node) {
		return containerObjectLit
	}
	// Object-literal `foo: function() {...}` form.
	if parent.Kind == ast.KindPropertyAssignment && node.Kind == ast.KindFunctionExpression {
		if parent.AsPropertyAssignment().Initializer == node {
			return containerObjectLit
		}
	}
	return containerNone
}

// containerKeyName returns the method-container's key as a string — the
// value ESLint would read via `'name' in current.key ? current.key.name :
// undefined`. Only Identifier and PrivateIdentifier populate `.name` in
// ESTree (the latter without the leading `#` per spec); StringLiteral,
// NumericLiteral, and ComputedPropertyName all yield `undefined` and never
// match a real callee name. We mirror that strictly — any non-Identifier /
// non-PrivateIdentifier key yields "" so propagation stays upstream-aligned
// even for e.g. `['nextState']() {...}` or `"nextState"() {...}`.
//
// For class-body members the key lives on the node itself (`node.Name()`);
// for object-literal property values it lives on the PropertyAssignment
// parent.
func containerKeyName(node *ast.Node, kind containerKind) string {
	switch kind {
	case containerClass:
		if node.Kind == ast.KindConstructor {
			// ESTree's constructor key is Identifier('constructor') — keep
			// parity even though no real code calls `constructor` by name.
			return "constructor"
		}
		return keyIdentifierText(node.Name())
	case containerObjectLit:
		keyHost := node
		if node.Parent.Kind == ast.KindPropertyAssignment {
			keyHost = node.Parent
		}
		return keyIdentifierText(keyHost.Name())
	}
	return ""
}

// keyIdentifierText extracts the string value ESLint would see via
// `'name' in node.key`. Returns the raw Identifier text, or a
// PrivateIdentifier's text with the leading `#` stripped. Any other key
// shape (StringLiteral, NumericLiteral, ComputedPropertyName, nil) yields
// "" — upstream's `'name' in key` evaluates to false for those.
func keyIdentifierText(n *ast.Node) string {
	if n == nil {
		return ""
	}
	switch n.Kind {
	case ast.KindIdentifier:
		return n.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return strings.TrimPrefix(n.AsPrivateIdentifier().Text, "#")
	}
	return ""
}

// isValueOrObjectPosition mirrors upstream's Identifier-listener guard:
//
//	while (current.parent.type === 'BinaryExpression') current = current.parent;
//	if (('value' in current.parent && current.parent.value === current)
//	    || ('object' in current.parent && current.parent.object === current))
//
// Mapping to tsgo:
//
//   - PropertyAssignment.Initializer  — ESTree Property.value (`{ foo: x }`)
//   - ShorthandPropertyAssignment.Name — ESTree Property where key === value
//     (`{ x }`); upstream's property.value still refers to the Identifier.
//   - PropertyAccessExpression.Expression — ESTree MemberExpression.object
//     for `x.foo`.
//   - ElementAccessExpression.Expression — ESTree MemberExpression.object
//     for `x[foo]`. The computed-key ArgumentExpression is NOT an "object"
//     position — it corresponds to ESTree's `property`, not `object`.
//
// The BinaryExpression walk is restricted to non-assignment operators because
// tsgo collapses ESTree's BinaryExpression + AssignmentExpression into one
// Kind; upstream's check only targets non-assignment binary wrappers like
// `a + b`.
func isValueOrObjectPosition(id *ast.Node) bool {
	p := walkUpBinary(id).Parent
	if p == nil {
		return false
	}
	cur := walkUpBinary(id)
	switch p.Kind {
	case ast.KindPropertyAssignment:
		return p.AsPropertyAssignment().Initializer == cur
	case ast.KindShorthandPropertyAssignment:
		return p.AsShorthandPropertyAssignment().Name() == cur
	case ast.KindPropertyAccessExpression:
		return p.AsPropertyAccessExpression().Expression == cur
	case ast.KindElementAccessExpression:
		return p.AsElementAccessExpression().Expression == cur
	}
	return false
}

// walkUpBinary returns the outermost node reached by following non-assignment
// BinaryExpression parents. Equivalent to the first loop in upstream's
// Identifier listener — used so the caller can continue walking for the
// setState-first-arg check from the post-binary-wrapper position.
func walkUpBinary(node *ast.Node) *ast.Node {
	cur := node
	for cur.Parent != nil && cur.Parent.Kind == ast.KindBinaryExpression {
		bin := cur.Parent.AsBinaryExpression()
		if bin.OperatorToken == nil || ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
			break
		}
		cur = cur.Parent
	}
	return cur
}

// objectBindingPatternPropertyName extracts the "property-side name" of a
// BindingElement — matching ESLint's `'name' in property.key && property.key.name`
// lookup. Only Identifier keys (explicit or shorthand) participate; computed
// keys, string-literal keys, and rest elements yield "".
func objectBindingPatternPropertyName(be *ast.Node) string {
	if be == nil || be.Kind != ast.KindBindingElement {
		return ""
	}
	bindingElem := be.AsBindingElement()
	if bindingElem.DotDotDotToken != nil {
		return ""
	}
	if bindingElem.PropertyName != nil {
		if bindingElem.PropertyName.Kind == ast.KindIdentifier {
			return bindingElem.PropertyName.AsIdentifier().Text
		}
		return ""
	}
	// Shorthand: `{ state }` — the local Name doubles as the property key.
	n := bindingElem.Name()
	if n != nil && n.Kind == ast.KindIdentifier {
		return n.AsIdentifier().Text
	}
	return ""
}

// findEnclosingClassMethod returns the nearest ancestor that is an ESTree
// MethodDefinition (class-body member). Used by the CallExpression listener's
// propagation loop.
func findEnclosingClassMethod(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node, func(n *ast.Node) bool {
		return methodContainerKind(n) == containerClass
	})
}

var NoAccessStateInSetstateRule = rule.Rule{
	Name: "react/no-access-state-in-setstate",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		isClassComponent := func(node *ast.Node) bool {
			return reactutil.GetEnclosingReactComponent(node, pragma, createClass) != nil
		}

		report := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "useCallback",
				Description: "Use callback in setState when referencing the previous state.",
			})
		}

		type methodEntry struct {
			methodName string
			node       *ast.Node
		}
		// varEntry stores one tracked binding.
		//
		//   - symbol: non-nil when the binding was resolved through the type
		//     checker. Use-site matching then compares symbol identity, which
		//     correctly distinguishes `let`/`const` bindings declared in
		//     sibling blocks of the same function (ESLint's block-scope
		//     semantics — a strict upstream match).
		//   - scope / variableName: fallback match used when either the
		//     binding or the use site couldn't be resolved to a symbol
		//     (e.g. plain-JS files with no TypeChecker available). Coarser
		//     than block-scope resolution; preserves the previous function-
		//     granular behavior rather than silently under-reporting.
		type varEntry struct {
			node         *ast.Node
			symbol       *ast.Symbol
			scope        *ast.Node
			variableName string
		}
		// methods accumulates function containers whose body reads `this.state`
		// (directly, or transitively via another tracked method). vars
		// accumulates local bindings initialized with / destructured from
		// `this.state` / `this`. Both persist across the whole file walk,
		// matching upstream's closure variables.
		var methods []methodEntry
		var vars []varEntry

		// symbolAt returns the symbol at `node` when the TypeChecker is
		// available, or nil otherwise. Centralized so every call site
		// consistently nil-guards on `ctx.TypeChecker`.
		symbolAt := func(node *ast.Node) *ast.Symbol {
			if ctx.TypeChecker == nil || node == nil {
				return nil
			}
			return ctx.TypeChecker.GetSymbolAtLocation(node)
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if !isThisStateMember(node) {
					return
				}
				if !isClassComponent(node) {
					return
				}
				for current := node; current != nil; current = current.Parent {
					if isFirstArgumentInSetStateCall(current, node) {
						report(node)
						return
					}
					if kind := methodContainerKind(current); kind != containerNone {
						methods = append(methods, methodEntry{
							methodName: containerKeyName(current, kind),
							node:       node,
						})
						return
					}
					if current.Kind == ast.KindVariableDeclaration {
						declName := current.AsVariableDeclaration().Name()
						vars = append(vars, varEntry{
							node:         node,
							symbol:       symbolAt(declName),
							scope:        utils.FindEnclosingScope(node),
							variableName: identifierOrEmpty(declName),
						})
						return
					}
				}
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if !isClassComponent(node) {
					return
				}
				call := node.AsCallExpression()
				// Upstream's `'name' in node.callee` admits both Identifier and
				// PrivateIdentifier callees (the latter's `.name` excludes the
				// leading `#` per ESTree spec). MemberExpression callees are
				// excluded — propagation only triggers through bare-identifier
				// calls. We mirror this and only compute a name for Identifier
				// / PrivateIdentifier callees.
				calleeName := ""
				switch call.Expression.Kind {
				case ast.KindIdentifier:
					calleeName = call.Expression.AsIdentifier().Text
				case ast.KindPrivateIdentifier:
					calleeName = strings.TrimPrefix(call.Expression.AsPrivateIdentifier().Text, "#")
				}

				// Propagate tracked methods up one class-body frame: if this
				// call invokes a tracked method by bare identifier, find the
				// enclosing MethodDefinition and register it too so later
				// `this.setState(outerMethod())` reports.
				if calleeName != "" {
					// Snapshot length so appends during iteration don't feed
					// back into the same loop — mirrors upstream's single
					// forEach pass.
					n := len(methods)
					for i := range n {
						method := methods[i]
						if method.methodName != calleeName {
							continue
						}
						if enclosing := findEnclosingClassMethod(node.Parent); enclosing != nil {
							methods = append(methods, methodEntry{
								methodName: containerKeyName(enclosing, containerClass),
								node:       method.node,
							})
						}
					}
				}

				// If this call sits inside the first argument of a setState
				// call, report once for every tracked method whose name
				// matches the callee — each tracked method stores the
				// original `this.state` node as the report position, per
				// upstream.
				for current := node.Parent; current != nil; current = current.Parent {
					if !isFirstArgumentInSetStateCall(current, node) {
						continue
					}
					if calleeName != "" {
						for _, method := range methods {
							if method.methodName == calleeName {
								report(method.node)
							}
						}
					}
					return
				}
			},

			ast.KindIdentifier: func(node *ast.Node) {
				if !isValueOrObjectPosition(node) {
					return
				}
				// Walk up through non-assignment BinaryExpression wrappers
				// first — upstream's inner `while` — so the outer walk starts
				// from the same position ESLint sees.
				start := walkUpBinary(node)
				useName := node.AsIdentifier().Text
				// Prefer symbol-identity matching (aligns with ESLint's
				// block-scope manager); fall back to function-scope + name
				// when either side lacks a symbol (no TypeChecker, or the
				// use site doesn't resolve — e.g. undeclared identifier).
				useSymbol := utils.GetReferenceSymbol(node, ctx.TypeChecker)
				var useScope *ast.Node // lazily computed for fallback
				for current := start; current != nil; current = current.Parent {
					if !isFirstArgumentInSetStateCall(current, node) {
						continue
					}
					for _, v := range vars {
						var matched bool
						if v.symbol != nil && useSymbol != nil {
							matched = v.symbol == useSymbol
						} else {
							if useScope == nil {
								useScope = utils.FindEnclosingScope(node)
							}
							matched = v.scope == useScope && v.variableName == useName
						}
						if matched {
							report(v.node)
						}
					}
					// Upstream does NOT break here — it keeps walking up,
					// potentially matching a second, outer setState call.
				}
			},

			ast.KindObjectBindingPattern: func(node *ast.Node) {
				// Upstream's `'init' in node.parent && node.parent.init &&
				// node.parent.init.type === 'ThisExpression'`. Only
				// VariableDeclarator parents carry an `.init`; Parameter and
				// nested BindingElement parents do not, so they are skipped.
				// Parentheses around `this` are skipped to emulate ESTree's
				// paren-stripping.
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindVariableDeclaration {
					return
				}
				init := parent.AsVariableDeclaration().Initializer
				if init == nil || !isThisReceiver(init) {
					return
				}
				scope := utils.FindEnclosingScope(node)
				pat := node.AsBindingPattern()
				if pat == nil || pat.Elements == nil {
					return
				}
				for _, elem := range pat.Elements.Nodes {
					if elem.Kind != ast.KindBindingElement {
						continue
					}
					if objectBindingPatternPropertyName(elem) != "state" {
						continue
					}
					be := elem.AsBindingElement()
					// Report node mirrors upstream's `property.key`: the
					// property-side identifier — PropertyName when renaming,
					// otherwise the Name (which doubles as key in shorthand).
					keyNode := be.PropertyName
					if keyNode == nil {
						keyNode = be.Name()
					}
					// Only capture a symbol for shorthand destructuring.
					// Renamed form `{state: aliased}` deliberately goes through
					// the name-based fallback: upstream's variableName='state'
					// will never match the user-typed 'aliased' at the use
					// site, so binding the local 'aliased' symbol here would
					// divergently "fix" a latent upstream quirk. Keeping the
					// renamed path symbol-less preserves upstream parity.
					var sym *ast.Symbol
					if be.PropertyName == nil {
						sym = symbolAt(be.Name())
					}
					vars = append(vars, varEntry{
						node:         keyNode,
						symbol:       sym,
						scope:        scope,
						variableName: "state",
					})
				}
			},
		}
	},
}

// identifierOrEmpty returns the text of `n` when it is an Identifier, "" otherwise.
// Used for VariableDeclaration bindings where only plain-Identifier declarators
// participate in the rule's variable-tracking (destructuring patterns are
// handled by the ObjectBindingPattern listener).
func identifierOrEmpty(n *ast.Node) string {
	if n == nil || n.Kind != ast.KindIdentifier {
		return ""
	}
	return n.AsIdentifier().Text
}
