package no_unused_class_component_methods

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var lifecycleMethods = map[string]bool{
	"constructor":                      true,
	"componentDidCatch":                true,
	"componentDidMount":                true,
	"componentDidUpdate":               true,
	"componentWillMount":               true,
	"componentWillReceiveProps":        true,
	"componentWillUnmount":             true,
	"componentWillUpdate":              true,
	"getChildContext":                  true,
	"getSnapshotBeforeUpdate":          true,
	"render":                           true,
	"shouldComponentUpdate":            true,
	"UNSAFE_componentWillMount":        true,
	"UNSAFE_componentWillReceiveProps": true,
	"UNSAFE_componentWillUpdate":       true,
}

var es6Lifecycle = map[string]bool{
	"state": true,
}

var es5Lifecycle = map[string]bool{
	"getInitialState": true,
	"getDefaultProps": true,
	"mixins":          true,
}

// propertyDef records a property definition — the node to report on and the
// extracted static name.
type propertyDef struct {
	node *ast.Node
	name string
}

// classInfo tracks a single React component's declared properties and usages.
type classInfo struct {
	properties     []*propertyDef
	usedProperties map[string]bool
	isClass        bool
	className      string
}

func newClassInfo(isClass bool, className string) *classInfo {
	return &classInfo{
		usedProperties: make(map[string]bool),
		isClass:        isClass,
		className:      className,
	}
}

// skipAssertionsAndParens strips parentheses and TS assertion wrappers,
// matching ESLint's `uncast` which peels TypeCastExpression wrappers.
func skipAssertionsAndParens(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
}

// isThisExpression reports whether `node` (after unwrapping assertions /
// parens) is a bare `this`.
func isThisExpression(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return skipAssertionsAndParens(node).Kind == ast.KindThisKeyword
}

// resolveLiteralKey extracts the static name from a property-key / member-name
// node, AND the inner node that should be used as the diagnostic position.
// Mirrors upstream's `isKeyLiteralLike` which accepts:
//   - Identifier when computed === false
//   - String / numeric / boolean / null / regex literals (computed or not)
//   - TemplateLiteral with no expressions (computed or not)
//
// Name extraction is delegated to utils.GetStaticPropertyName; this helper's
// extra job is to surface the reportable inner node (tsgo wraps computed keys
// in `ComputedPropertyName`; we unwrap so the diagnostic range sits on the
// literal inside the brackets, matching ESTree-oriented `node.key` positions).
// PrivateIdentifier (`#foo`) and dynamic computed keys (`[foo]`) are rejected
// for parity with upstream.
func resolveLiteralKey(nameNode *ast.Node) (string, *ast.Node) {
	if nameNode == nil {
		return "", nil
	}
	name, ok := utils.GetStaticPropertyName(nameNode)
	if !ok {
		return "", nil
	}
	if nameNode.Kind == ast.KindComputedPropertyName {
		return name, skipAssertionsAndParens(nameNode.AsComputedPropertyName().Expression)
	}
	return name, nameNode
}

// resolveLiteralAccessKey extracts the name from an `X[key]` access key.
// Upstream's `isKeyLiteralLike(node, node.property)` (with `computed === true`)
// accepts string / numeric / template-no-expr / boolean / null / regex
// literals — the same set utils.GetStaticPropertyName handles when wrapped in
// a ComputedPropertyName. We forward to it by synthesizing the wrapper pattern
// through the same Kind dispatch.
func resolveLiteralAccessKey(keyExpr *ast.Node) (string, *ast.Node) {
	if keyExpr == nil {
		return "", nil
	}
	keyExpr = skipAssertionsAndParens(keyExpr)
	switch keyExpr.Kind {
	case ast.KindStringLiteral:
		return keyExpr.AsStringLiteral().Text, keyExpr
	case ast.KindNumericLiteral:
		return utils.NormalizeNumericLiteral(keyExpr.AsNumericLiteral().Text), keyExpr
	case ast.KindNoSubstitutionTemplateLiteral:
		return keyExpr.AsNoSubstitutionTemplateLiteral().Text, keyExpr
	case ast.KindTrueKeyword:
		return "true", keyExpr
	case ast.KindFalseKeyword:
		return "false", keyExpr
	case ast.KindNullKeyword:
		return "null", keyExpr
	}
	return "", nil
}

// addProperty records a property definition (its reportable key node and name).
func (ci *classInfo) addProperty(node *ast.Node, name string) {
	if node == nil || name == "" {
		return
	}
	ci.properties = append(ci.properties, &propertyDef{node: node, name: name})
}

// markUsed records a name as used.
func (ci *classInfo) markUsed(name string) {
	if name == "" {
		return
	}
	ci.usedProperties[name] = true
}

// isAssignmentTarget reports whether `access` is the LHS of an assignment.
// Mirrors upstream's `node.parent.type === 'AssignmentExpression' && node.parent.left === node`.
// In tsgo, all assignments (`=`, `+=`, `-=`, …) collapse into BinaryExpression
// with an assignment operator — we gate on that.
//
// Parenthesized LHS: ESTree flattens parens, so upstream sees `(this.foo) = 1`
// as `AssignmentExpression(left: MemberExpression)` directly. tsgo preserves
// the ParenthesizedExpression wrapper, so we must walk up through any number
// of paren wrappers before checking the assignment shape — otherwise we'd
// mis-classify `(this.foo) = 1` (a definition) as a use.
func isAssignmentTarget(access *ast.Node) bool {
	cur := access
	parent := cur.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		cur = parent
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := parent.AsBinaryExpression()
	if bin.OperatorToken == nil || !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
		return false
	}
	return bin.Left == cur
}

// walkExpressions recursively visits every descendant of `node`, invoking the
// per-rule handlers for PropertyAccess, ElementAccess, and VariableDeclaration.
// The walk stops at nested class boundaries to avoid cross-component
// interference (a nested React class would be visited separately by the
// top-level ClassDeclaration listener).
func (ci *classInfo) walkExpressions(node *ast.Node) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return
	case ast.KindPropertyAccessExpression:
		ci.handlePropertyAccess(node)
	case ast.KindElementAccessExpression:
		ci.handleElementAccess(node)
	case ast.KindVariableDeclaration:
		ci.handleVariableDeclaration(node)
	}
	node.ForEachChild(func(child *ast.Node) bool {
		ci.walkExpressions(child)
		return false
	})
}

// handlePropertyAccess processes `this.X` / `(this).X` expressions.
func (ci *classInfo) handlePropertyAccess(node *ast.Node) {
	pa := node.AsPropertyAccessExpression()
	if !isThisExpression(pa.Expression) {
		return
	}
	nameNode := pa.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return
	}
	name := nameNode.AsIdentifier().Text
	if isAssignmentTarget(node) {
		ci.addProperty(nameNode, name)
	} else {
		ci.markUsed(name)
	}
}

// handleElementAccess processes `this['X']` / `this[0]` / `this[`X`]` accesses
// whose key is a literal-like expression. Dynamic keys are ignored (upstream's
// `isKeyLiteralLike` check on the key rejects them).
func (ci *classInfo) handleElementAccess(node *ast.Node) {
	ea := node.AsElementAccessExpression()
	if !isThisExpression(ea.Expression) {
		return
	}
	name, inner := resolveLiteralAccessKey(ea.ArgumentExpression)
	if inner == nil {
		return
	}
	if isAssignmentTarget(node) {
		ci.addProperty(inner, name)
	} else {
		ci.markUsed(name)
	}
}

// handleVariableDeclaration processes `const { foo, bar: baz } = this`
// patterns, marking `foo` and `bar` as used. Mirrors upstream's
// VariableDeclarator handler. In tsgo, VariableDeclaration is the nearest
// analog of ESTree's VariableDeclarator.
func (ci *classInfo) handleVariableDeclaration(node *ast.Node) {
	vd := node.AsVariableDeclaration()
	if vd.Initializer == nil {
		return
	}
	if !isThisExpression(vd.Initializer) {
		return
	}
	nameNode := vd.Name()
	if nameNode == nil || nameNode.Kind != ast.KindObjectBindingPattern {
		return
	}
	bp := nameNode.AsBindingPattern()
	if bp.Elements == nil {
		return
	}
	for _, elem := range bp.Elements.Nodes {
		if elem.Kind != ast.KindBindingElement {
			continue
		}
		be := elem.AsBindingElement()
		// Rest element (`...rest`) has no key — skip. Upstream's filter requires
		// `prop.type === 'Property'` which excludes RestElement.
		if be.DotDotDotToken != nil {
			continue
		}
		// Explicit key: `{ foo: local }` — use the property name.
		if be.PropertyName != nil {
			name, _ := resolveLiteralKey(be.PropertyName)
			ci.markUsed(name)
			continue
		}
		// Shorthand: `{ foo }` — the binding name doubles as the key.
		local := be.Name()
		if local != nil && local.Kind == ast.KindIdentifier {
			ci.markUsed(local.AsIdentifier().Text)
		}
	}
}

// processES6Class walks an ES6 class body collecting property definitions and
// usages. Static members are skipped entirely (upstream's `inStatic` guard).
func processES6Class(ci *classInfo, classNode *ast.Node) {
	members := classNode.Members()
	if members == nil {
		return
	}
	for _, member := range members {
		if ast.IsStatic(member) {
			continue
		}
		switch member.Kind {
		case ast.KindPropertyDeclaration:
			pd := member.AsPropertyDeclaration()
			if name, keyNode := resolveLiteralKey(pd.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
			if pd.Initializer != nil {
				ci.walkExpressions(pd.Initializer)
			}
		case ast.KindMethodDeclaration:
			md := member.AsMethodDeclaration()
			if name, keyNode := resolveLiteralKey(md.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
			if md.Body != nil {
				ci.walkExpressions(md.Body)
			}
		case ast.KindGetAccessor:
			ga := member.AsGetAccessorDeclaration()
			if name, keyNode := resolveLiteralKey(ga.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
			if ga.Body != nil {
				ci.walkExpressions(ga.Body)
			}
		case ast.KindSetAccessor:
			sa := member.AsSetAccessorDeclaration()
			if name, keyNode := resolveLiteralKey(sa.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
			if sa.Body != nil {
				ci.walkExpressions(sa.Body)
			}
		case ast.KindConstructor:
			// The constructor itself has no user-provided key to add (upstream
			// adds `constructor` then filters it via LIFECYCLE_METHODS — the
			// net effect is nothing). We only need to walk the body for
			// `this.X = …` definitions and `this.X` usages.
			cd := member.AsConstructorDeclaration()
			if cd.Body != nil {
				ci.walkExpressions(cd.Body)
			}
		}
	}
}

// processES5Object walks a createReactClass object literal collecting property
// definitions and usages.
func processES5Object(ci *classInfo, obj *ast.Node) {
	ole := obj.AsObjectLiteralExpression()
	if ole.Properties == nil {
		return
	}
	for _, prop := range ole.Properties.Nodes {
		switch prop.Kind {
		case ast.KindPropertyAssignment:
			pa := prop.AsPropertyAssignment()
			if name, keyNode := resolveLiteralKey(pa.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
			if pa.Initializer != nil {
				ci.walkExpressions(pa.Initializer)
			}
		case ast.KindShorthandPropertyAssignment:
			spa := prop.AsShorthandPropertyAssignment()
			if name, keyNode := resolveLiteralKey(spa.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
		case ast.KindMethodDeclaration:
			md := prop.AsMethodDeclaration()
			if name, keyNode := resolveLiteralKey(md.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
			if md.Body != nil {
				ci.walkExpressions(md.Body)
			}
		case ast.KindGetAccessor:
			ga := prop.AsGetAccessorDeclaration()
			if name, keyNode := resolveLiteralKey(ga.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
			if ga.Body != nil {
				ci.walkExpressions(ga.Body)
			}
		case ast.KindSetAccessor:
			sa := prop.AsSetAccessorDeclaration()
			if name, keyNode := resolveLiteralKey(sa.Name()); keyNode != nil {
				ci.addProperty(keyNode, name)
			}
			if sa.Body != nil {
				ci.walkExpressions(sa.Body)
			}
		}
	}
}

// reportUnused emits a diagnostic for each declared property that was never
// referenced and is not a recognized lifecycle method / reserved name.
func reportUnused(ctx rule.RuleContext, ci *classInfo) {
	for _, pd := range ci.properties {
		if ci.usedProperties[pd.name] {
			continue
		}
		if lifecycleMethods[pd.name] {
			continue
		}
		if ci.isClass {
			if es6Lifecycle[pd.name] {
				continue
			}
		} else {
			if es5Lifecycle[pd.name] {
				continue
			}
		}
		if ci.className != "" {
			ctx.ReportNode(pd.node, rule.RuleMessage{
				Id:          "unusedWithClass",
				Description: fmt.Sprintf("Unused method or property %q of class %q", pd.name, ci.className),
			})
		} else {
			ctx.ReportNode(pd.node, rule.RuleMessage{
				Id:          "unused",
				Description: fmt.Sprintf("Unused method or property %q", pd.name),
			})
		}
	}
}

// getClassName returns the class's identifier text, or "" for an anonymous
// class expression.
func getClassName(classNode *ast.Node) string {
	name := classNode.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

var NoUnusedClassComponentMethodsRule = rule.Rule{
	Name: "react/no-unused-class-component-methods",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		runOnClass := func(node *ast.Node) {
			if !reactutil.ExtendsReactComponent(node, pragma) {
				return
			}
			ci := newClassInfo(true, getClassName(node))
			processES6Class(ci, node)
			reportUnused(ctx, ci)
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: runOnClass,
			ast.KindClassExpression:  runOnClass,
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateClassCall(call, pragma, createClass) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				arg := ast.SkipParentheses(call.Arguments.Nodes[0])
				if arg.Kind != ast.KindObjectLiteralExpression {
					return
				}
				ci := newClassInfo(false, "")
				processES5Object(ci, arg)
				reportUnused(ctx, ci)
			},
		}
	},
}
