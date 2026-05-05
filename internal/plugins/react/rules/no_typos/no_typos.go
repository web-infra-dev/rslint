package no_typos

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// propTypesNames is the set of valid PropTypes.* property names, mirroring
// `Object.keys(require('prop-types'))` in eslint-plugin-react's no-typos.
var propTypesNames = map[string]bool{
	"array":             true,
	"bigint":            true,
	"bool":              true,
	"func":              true,
	"number":            true,
	"object":            true,
	"string":            true,
	"symbol":            true,
	"any":               true,
	"arrayOf":           true,
	"element":           true,
	"elementType":       true,
	"instanceOf":        true,
	"node":              true,
	"objectOf":          true,
	"oneOf":             true,
	"oneOfType":         true,
	"shape":             true,
	"exact":             true,
	"checkPropTypes":    true,
	"resetWarningCache": true,
	"PropTypes":         true,
}

// staticClassProperties are the React component static-class-property names
// whose casing the rule enforces. Matching is case-insensitive.
var staticClassProperties = []string{
	"propTypes", "contextTypes", "childContextTypes", "defaultProps",
}

// lifecycleInstance lists React's instance-scope lifecycle method names (plus
// the ES5 `createReactClass` companions). Matching is case-insensitive.
var lifecycleInstance = []string{
	"getDefaultProps", "getInitialState", "getChildContext",
	"componentWillMount", "UNSAFE_componentWillMount",
	"componentDidMount",
	"componentWillReceiveProps", "UNSAFE_componentWillReceiveProps",
	"shouldComponentUpdate",
	"componentWillUpdate", "UNSAFE_componentWillUpdate",
	"getSnapshotBeforeUpdate",
	"componentDidUpdate",
	"componentDidCatch",
	"componentWillUnmount",
	"render",
}

// lifecycleStatic lists lifecycle methods that must be declared `static`.
var lifecycleStatic = []string{"getDerivedStateFromProps"}

var NoTyposRule = rule.Rule{
	Name: "react/no-typos",
	Run:  runRule,
}

func runRule(ctx rule.RuleContext, options any) rule.RuleListeners {
	pragma := reactutil.GetReactPragma(ctx.Settings)
	createClass := reactutil.GetReactCreateClass(ctx.Settings)

	propTypesPackageName := ""
	reactPackageName := ""

	// componentsByName records names bound to a React component anywhere in
	// the file — an ES6 class extending React.Component / PureComponent
	// (optionally via `const X = class extends ...`) or a function that
	// returns JSX. Used by the `Foo.PropTypes = {}` branch, which needs a
	// file-wide name lookup (ESLint's `utils.getRelatedComponent`).
	componentsByName := map[string]bool{}

	// Pre-walk the source file: collect imports (for propTypesPackageName /
	// reactPackageName), emit noReactBinding / noPropTypesBinding, and
	// populate componentsByName. Listeners can then rely on these.
	var pre ast.Visitor
	pre = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		switch n.Kind {
		case ast.KindImportDeclaration:
			handleImportDecl(ctx, n, &propTypesPackageName, &reactPackageName)
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if reactutil.ExtendsReactComponent(n, pragma) || hasJSDocExtendsReactComponent(n) {
				if name := reactutil.BindingIdentifierName(n); name != "" {
					componentsByName[name] = true
				}
			}
		case ast.KindFunctionDeclaration:
			if functionReturnsJSX(n) {
				if nm := n.Name(); nm != nil && nm.Kind == ast.KindIdentifier {
					componentsByName[nm.AsIdentifier().Text] = true
				}
			}
		case ast.KindVariableDeclaration:
			vd := n.AsVariableDeclaration()
			if vd == nil || vd.Initializer == nil || vd.Name() == nil || vd.Name().Kind != ast.KindIdentifier {
				break
			}
			init := ast.SkipParentheses(vd.Initializer)
			name := vd.Name().AsIdentifier().Text
			switch init.Kind {
			case ast.KindClassExpression:
				if reactutil.ExtendsReactComponent(init, pragma) {
					componentsByName[name] = true
				}
			case ast.KindFunctionExpression, ast.KindArrowFunction:
				if functionReturnsJSX(init) {
					componentsByName[name] = true
				}
			}
		}
		n.ForEachChild(pre)
		return false
	}
	ctx.SourceFile.Node.ForEachChild(pre)

	// isPropTypesPackage reports whether `expr` refers to the imported
	// PropTypes namespace — either the default/namespace binding from
	// `prop-types` (e.g. `PropTypes`), or `<reactPkg>.PropTypes` (e.g.
	// `React.PropTypes`).
	isPropTypesPackage := func(expr *ast.Node) bool {
		if propTypesPackageName == "" && reactPackageName == "" {
			return false
		}
		expr = ast.SkipParentheses(expr)
		switch expr.Kind {
		case ast.KindIdentifier:
			return propTypesPackageName != "" && expr.AsIdentifier().Text == propTypesPackageName
		case ast.KindPropertyAccessExpression:
			if reactPackageName == "" {
				return false
			}
			pa := expr.AsPropertyAccessExpression()
			obj := ast.SkipParentheses(pa.Expression)
			name := pa.Name()
			if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "PropTypes" {
				return false
			}
			return obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == reactPackageName
		}
		return false
	}

	// checkValidPropType reports a typoPropType diagnostic when `node` is an
	// Identifier whose text is not a known PropTypes property name.
	checkValidPropType := func(node *ast.Node) {
		if node == nil || node.Kind != ast.KindIdentifier {
			return
		}
		name := node.AsIdentifier().Text
		if name == "" {
			return
		}
		if propTypesNames[name] {
			return
		}
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "typoPropType",
			Description: "Typo in declared prop type: " + name,
		})
	}

	// checkValidPropTypeQualifier reports typoPropTypeChain when `node` is
	// not the literal identifier `isRequired`.
	checkValidPropTypeQualifier := func(node *ast.Node) {
		if node == nil || node.Kind != ast.KindIdentifier {
			return
		}
		name := node.AsIdentifier().Text
		if name == "isRequired" {
			return
		}
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "typoPropTypeChain",
			Description: "Typo in prop type chain qualifier: " + name,
		})
	}

	// Mutual recursion: checkValidProp may call checkValidCallExpression,
	// which in turn descends into prop values via checkValidPropObject /
	// checkValidProp.
	var checkValidProp func(node *ast.Node)
	var checkValidCallExpression func(node *ast.Node)
	var checkValidPropObject func(node *ast.Node)

	checkValidCallExpression = func(node *ast.Node) {
		if node == nil || node.Kind != ast.KindCallExpression {
			return
		}
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			return
		}
		pa := callee.AsPropertyAccessExpression()
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return
		}
		args := call.Arguments
		switch name.AsIdentifier().Text {
		case "shape":
			if args == nil || len(args.Nodes) == 0 {
				return
			}
			checkValidPropObject(args.Nodes[0])
		case "oneOfType":
			if args == nil || len(args.Nodes) == 0 {
				return
			}
			first := ast.SkipParentheses(args.Nodes[0])
			if first.Kind != ast.KindArrayLiteralExpression {
				return
			}
			for _, el := range first.AsArrayLiteralExpression().Elements.Nodes {
				checkValidProp(el)
			}
		}
	}

	checkValidProp = func(node *ast.Node) {
		if propTypesPackageName == "" && reactPackageName == "" {
			return
		}
		if node == nil {
			return
		}
		node = ast.SkipParentheses(node)
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			pa := node.AsPropertyAccessExpression()
			obj := ast.SkipParentheses(pa.Expression)
			propName := pa.Name()
			if propName == nil || propName.Kind != ast.KindIdentifier {
				return
			}
			// PropTypes.myProp.isRequired
			if obj.Kind == ast.KindPropertyAccessExpression {
				innerPa := obj.AsPropertyAccessExpression()
				innerObj := ast.SkipParentheses(innerPa.Expression)
				if isPropTypesPackage(innerObj) {
					checkValidPropType(innerPa.Name())
					checkValidPropTypeQualifier(propName)
					return
				}
			}
			// PropTypes.myProp
			if isPropTypesPackage(obj) && propName.AsIdentifier().Text != "isRequired" {
				checkValidPropType(propName)
				return
			}
			// (PropTypes.shape({...})).isRequired / .somethingElse
			if obj.Kind == ast.KindCallExpression {
				checkValidPropTypeQualifier(propName)
				checkValidCallExpression(obj)
			}
		case ast.KindCallExpression:
			checkValidCallExpression(node)
		}
	}

	checkValidPropObject = func(node *ast.Node) {
		if node == nil {
			return
		}
		node = ast.SkipParentheses(node)
		if node.Kind != ast.KindObjectLiteralExpression {
			return
		}
		for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
			value := propertyValueNode(prop)
			if value != nil {
				checkValidProp(value)
			}
		}
	}

	// reportErrorIfPropertyCasingTypo reports typoStaticClassProp /
	// typoPropDeclaration when `key` is an identifier whose lowercased text
	// equals one of the static class property names but casing differs. When
	// the property name is one of propTypes / contextTypes / childContextTypes
	// (exact casing), it also recurses into `value` via checkValidPropObject.
	reportErrorIfPropertyCasingTypo := func(value *ast.Node, key *ast.Node, isClassProperty bool) {
		if key == nil || key.Kind != ast.KindIdentifier {
			return
		}
		propertyName := key.AsIdentifier().Text
		if propertyName == "" {
			return
		}
		// Recurse into declared propTypes / contextTypes / childContextTypes.
		if propertyName == "propTypes" || propertyName == "contextTypes" || propertyName == "childContextTypes" {
			checkValidPropObject(value)
		}
		lower := strings.ToLower(propertyName)
		for _, canonical := range staticClassProperties {
			if strings.ToLower(canonical) != lower || canonical == propertyName {
				continue
			}
			messageId := "typoPropDeclaration"
			description := "Typo in property declaration"
			if isClassProperty {
				messageId = "typoStaticClassProp"
				description = "Typo in static class property declaration"
			}
			ctx.ReportNode(key, rule.RuleMessage{
				Id:          messageId,
				Description: description,
			})
			return
		}
	}

	// reportErrorIfLifecycleMethodCasingTypo reports on a class/object member
	// whose key mis-cases a lifecycle method. `isStatic` is the actual
	// static-ness of the declaration (object literal members are always
	// non-static).
	reportErrorIfLifecycleMethodCasingTypo := func(member *ast.Node, isStatic bool) {
		key := member.Name()
		if key == nil {
			return
		}
		var nodeKeyName string
		switch key.Kind {
		case ast.KindIdentifier:
			nodeKeyName = key.AsIdentifier().Text
		case ast.KindStringLiteral:
			nodeKeyName = key.AsStringLiteral().Text
		case ast.KindNoSubstitutionTemplateLiteral:
			nodeKeyName = key.AsNoSubstitutionTemplateLiteral().Text
		case ast.KindPrivateIdentifier:
			// ESLint short-circuits on PrivateName.
			return
		default:
			// Computed keys that resolve to a non-string value (numbers,
			// expressions) are excluded, matching ESLint's
			// `typeof nodeKeyName !== 'string'` guard.
			return
		}
		if nodeKeyName == "" {
			return
		}
		lowerKey := strings.ToLower(nodeKeyName)

		// staticLifecycleMethod: declared non-static but name matches a static
		// lifecycle method (case-insensitively).
		if !isStatic {
			for _, method := range lifecycleStatic {
				if strings.ToLower(method) == lowerKey {
					ctx.ReportNode(member, rule.RuleMessage{
						Id:          "staticLifecycleMethod",
						Description: "Lifecycle method should be static: " + nodeKeyName,
					})
					break
				}
			}
		}

		// typoLifecycleMethod: name case-insensitively matches a known
		// lifecycle method (instance or static) but casing differs.
		for _, method := range lifecycleInstance {
			if strings.ToLower(method) == lowerKey && method != nodeKeyName {
				ctx.ReportNode(member, rule.RuleMessage{
					Id:          "typoLifecycleMethod",
					Description: "Typo in component lifecycle method declaration: " + nodeKeyName + " should be " + method,
				})
				return
			}
		}
		for _, method := range lifecycleStatic {
			if strings.ToLower(method) == lowerKey && method != nodeKeyName {
				ctx.ReportNode(member, rule.RuleMessage{
					Id:          "typoLifecycleMethod",
					Description: "Typo in component lifecycle method declaration: " + nodeKeyName + " should be " + method,
				})
				return
			}
		}
	}

	return rule.RuleListeners{
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			// ES6 class field: `static propTypes = {...}`. Only reports when
			// declared static on a class that extends React.Component.
			if !ast.IsStatic(node) {
				return
			}
			classNode := reactutil.EnclosingClass(node)
			if classNode == nil || !reactutil.ExtendsReactComponent(classNode, pragma) {
				return
			}
			pd := node.AsPropertyDeclaration()
			reportErrorIfPropertyCasingTypo(pd.Initializer, pd.Name(), true)
		},

		ast.KindMethodDeclaration: func(node *ast.Node) {
			// ES6 class method only — object literal methods are handled
			// uniformly by the ObjectLiteralExpression listener, matching
			// ESLint's ESTree split between `MethodDefinition` (class) and
			// `Property` (object).
			parent := node.Parent
			if parent == nil {
				return
			}
			switch parent.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if !reactutil.ExtendsReactComponent(parent, pragma) {
					return
				}
				reportErrorIfLifecycleMethodCasingTypo(node, ast.IsStatic(node))
			}
		},

		// `Foo.PropTypes = {}` — handled at the assignment level so we can
		// examine both the property (LHS) and the value (RHS).
		ast.KindBinaryExpression: func(node *ast.Node) {
			bin := node.AsBinaryExpression()
			if bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken {
				return
			}
			left := ast.SkipParentheses(bin.Left)
			if left.Kind != ast.KindPropertyAccessExpression {
				return
			}
			pa := left.AsPropertyAccessExpression()
			base := ast.SkipParentheses(pa.Expression)
			if base.Kind != ast.KindIdentifier {
				return
			}
			propertyName := pa.Name()
			if propertyName == nil || propertyName.Kind != ast.KindIdentifier {
				return
			}
			lower := strings.ToLower(propertyName.AsIdentifier().Text)
			anyMatch := false
			for _, canonical := range staticClassProperties {
				if strings.ToLower(canonical) == lower {
					anyMatch = true
					break
				}
			}
			if !anyMatch {
				return
			}
			if !componentsByName[base.AsIdentifier().Text] {
				return
			}
			reportErrorIfPropertyCasingTypo(bin.Right, propertyName, true)
		},

		ast.KindObjectLiteralExpression: func(node *ast.Node) {
			// Component defined via createReactClass({...}): inspect top-level
			// properties for prop-declaration typos and lifecycle typos.
			if !reactutil.IsCreateReactClassObjectArg(node, pragma, createClass) {
				return
			}
			for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
				if prop.Kind == ast.KindSpreadAssignment {
					continue
				}
				key := prop.Name()
				value := propertyValueNode(prop)
				if key != nil {
					reportErrorIfPropertyCasingTypo(value, key, false)
				}
				reportErrorIfLifecycleMethodCasingTypo(prop, false)
			}
		},
	}
}

// handleImportDecl inspects an ImportDeclaration: tracks propTypesPackageName
// and reactPackageName for prop-type typo detection, and reports
// noPropTypesBinding / noReactBinding when the import has no value binding
// (e.g. `import 'prop-types'`).
func handleImportDecl(ctx rule.RuleContext, node *ast.Node, propTypesPackageName, reactPackageName *string) {
	decl := node.AsImportDeclaration()
	if decl == nil || decl.ModuleSpecifier == nil || decl.ModuleSpecifier.Kind != ast.KindStringLiteral {
		return
	}
	moduleName := decl.ModuleSpecifier.AsStringLiteral().Text
	if moduleName != "prop-types" && moduleName != "react" {
		return
	}
	specifiers := importSpecifierLocalNames(decl)
	if moduleName == "prop-types" {
		if len(specifiers) == 0 {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noPropTypesBinding",
				Description: "`'prop-types'` imported without a local `PropTypes` binding.",
			})
			return
		}
		// ESLint reads `node.specifiers[0].local.name`, so we mirror that —
		// the first specifier's local name becomes the PropTypes binding.
		*propTypesPackageName = specifiers[0]
		return
	}
	// react
	if len(specifiers) == 0 {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "noReactBinding",
			Description: "`'react'` imported without a local `React` binding.",
		})
		return
	}
	*reactPackageName = specifiers[0]
	// If any named import is `PropTypes`, treat its local name as the
	// PropTypes binding (overrides any previous state — matches upstream).
	if propTypesLocal := findPropTypesNamedImport(decl); propTypesLocal != "" {
		*propTypesPackageName = propTypesLocal
	}
}

// importSpecifierLocalNames returns the local names of all value-binding
// specifiers in `decl`, in declaration order. The default binding (if any)
// comes first, followed by namespace / named specifiers. Mirrors ESLint's
// `node.specifiers` traversal order on ESTree.
func importSpecifierLocalNames(decl *ast.ImportDeclaration) []string {
	if decl.ImportClause == nil {
		return nil
	}
	clause := decl.ImportClause.AsImportClause()
	if clause == nil {
		return nil
	}
	var names []string
	if clause.Name() != nil && clause.Name().Kind == ast.KindIdentifier {
		names = append(names, clause.Name().AsIdentifier().Text)
	}
	if clause.NamedBindings != nil {
		switch clause.NamedBindings.Kind {
		case ast.KindNamespaceImport:
			ns := clause.NamedBindings.AsNamespaceImport()
			if ns != nil && ns.Name() != nil && ns.Name().Kind == ast.KindIdentifier {
				names = append(names, ns.Name().AsIdentifier().Text)
			}
		case ast.KindNamedImports:
			named := clause.NamedBindings.AsNamedImports()
			if named != nil && named.Elements != nil {
				for _, elem := range named.Elements.Nodes {
					spec := elem.AsImportSpecifier()
					if spec == nil || spec.Name() == nil || spec.Name().Kind != ast.KindIdentifier {
						continue
					}
					names = append(names, spec.Name().AsIdentifier().Text)
				}
			}
		}
	}
	return names
}

// findPropTypesNamedImport returns the local binding name of a `PropTypes`
// named import (`import { PropTypes as X } from 'react'`), or "" if no such
// specifier exists.
func findPropTypesNamedImport(decl *ast.ImportDeclaration) string {
	if decl.ImportClause == nil {
		return ""
	}
	clause := decl.ImportClause.AsImportClause()
	if clause == nil || clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return ""
	}
	named := clause.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil {
		return ""
	}
	for _, elem := range named.Elements.Nodes {
		spec := elem.AsImportSpecifier()
		if spec == nil {
			continue
		}
		// `imported` is PropertyName (the name in the module's namespace).
		// Under tsgo, the "propertyName" field captures `imported` when
		// `import { X as Y }`; otherwise `name` captures the imported name.
		imported := spec.PropertyName
		local := spec.Name()
		if imported == nil {
			imported = local
		}
		if imported == nil || imported.Kind != ast.KindIdentifier {
			continue
		}
		if imported.AsIdentifier().Text != "PropTypes" {
			continue
		}
		if local != nil && local.Kind == ast.KindIdentifier {
			return local.AsIdentifier().Text
		}
	}
	return ""
}

// propertyValueNode extracts the value expression of an object-literal
// property. Returns nil for shorthand / spread / method shorthand, whose
// "value" is not a separate expression in tsgo.
func propertyValueNode(prop *ast.Node) *ast.Node {
	if prop == nil {
		return nil
	}
	switch prop.Kind {
	case ast.KindPropertyAssignment:
		return prop.AsPropertyAssignment().Initializer
	case ast.KindShorthandPropertyAssignment:
		// In ESTree, value === key for shorthand. Upstream still calls
		// `reportErrorIfPropertyCasingTypo(property.value, property.key, ...)`;
		// `checkValidPropObject` early-returns when `value.type !== 'ObjectExpression'`,
		// so shorthand is effectively a no-op there. For static-class-property
		// typo detection the value is unused, so returning nil is fine.
		return nil
	}
	return nil
}

// functionReturnsJSX reports whether the given function-like node contains a
// JSX element / fragment in a position that a `return <...>` would produce —
// either as the arrow-expression body or inside a return statement in the
// block body. Nested functions are not descended into.
func functionReturnsJSX(fn *ast.Node) bool {
	if fn == nil {
		return false
	}
	var body *ast.Node
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		body = fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		body = fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		af := fn.AsArrowFunction()
		if af.Body == nil {
			return false
		}
		if af.Body.Kind != ast.KindBlock {
			// Implicit return: check whether the expression contains a JSX
			// element / fragment transitively.
			return containsJSX(af.Body)
		}
		body = af.Body
	default:
		return false
	}
	if body == nil {
		return false
	}
	found := false
	var walk ast.Visitor
	walk = func(n *ast.Node) bool {
		if found || n == nil {
			return found
		}
		switch n.Kind {
		case ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindArrowFunction:
			return false
		case ast.KindReturnStatement:
			rs := n.AsReturnStatement()
			if rs.Expression != nil && containsJSX(rs.Expression) {
				found = true
				return true
			}
		}
		n.ForEachChild(walk)
		return found
	}
	walk(body)
	return found
}

// containsJSX reports whether `node` contains a JSX element / fragment /
// self-closing element anywhere in its subtree, not crossing nested
// function-like boundaries.
func containsJSX(node *ast.Node) bool {
	if node == nil {
		return false
	}
	found := false
	var walk ast.Visitor
	walk = func(n *ast.Node) bool {
		if found || n == nil {
			return found
		}
		switch n.Kind {
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
			found = true
			return true
		case ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindArrowFunction:
			return false
		}
		n.ForEachChild(walk)
		return found
	}
	walk(node)
	return found
}

// hasJSDocExtendsReactComponent reports whether a class has a JSDoc
// `@extends React.Component` / `@augments React.Component` tag.
//
// Mirrors eslint-plugin-react's componentUtil which, on a class with a
// `@extends` JSDoc tag, treats it as a React component even when the
// syntactic `extends` clause does not name React.Component. Only the
// simple `.Component` / `.PureComponent` qualifier (with or without the
// `React.` prefix) is recognized — matches the regex ESLint uses.
func hasJSDocExtendsReactComponent(classNode *ast.Node) bool {
	if classNode == nil {
		return false
	}
	jsDocs := classNode.JSDoc(nil) // nil file → walks to the enclosing SourceFile
	for _, doc := range jsDocs {
		jd := doc.AsJSDoc()
		if jd == nil || jd.Tags == nil {
			continue
		}
		for _, tag := range jd.Tags.Nodes {
			if !ast.IsJSDocAugmentsTag(tag) {
				continue
			}
			aug := tag.AsJSDocAugmentsTag()
			if aug == nil || aug.ClassName == nil {
				continue
			}
			if jsDocTypeRefMatchesComponent(aug.ClassName) {
				return true
			}
		}
	}
	return false
}

// jsDocTypeRefMatchesComponent recognizes `Component` / `PureComponent` and
// `<any>.Component` / `<any>.PureComponent` shapes inside a JSDoc @extends /
// @augments tag's type reference.
func jsDocTypeRefMatchesComponent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		t := node.AsIdentifier().Text
		return t == "Component" || t == "PureComponent"
	case ast.KindExpressionWithTypeArguments:
		return jsDocTypeRefMatchesComponent(node.AsExpressionWithTypeArguments().Expression)
	case ast.KindPropertyAccessExpression:
		pa := node.AsPropertyAccessExpression()
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		t := name.AsIdentifier().Text
		return t == "Component" || t == "PureComponent"
	}
	return false
}
