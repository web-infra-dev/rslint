package prop_types

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed prop_types.schema.json
var schemaJSON []byte

type options struct {
	ignore, customValidators []string
	skipUndeclared           bool
}
type propType struct {
	children           map[string]propType
	union              []propType
	any, open, invalid bool
}
type use struct {
	node  *ast.Node
	names []string
}
type component struct {
	node          *ast.Node
	binding       *ast.Symbol
	declared      map[string]propType
	used          []use
	declaredBlock bool
	ignore        bool
	props         string
	destructured  map[string][]string
}

func parseOptions(raw []any) options {
	o := options{}
	if len(raw) == 0 {
		return o
	}
	m, _ := raw[0].(map[string]interface{})
	for _, key := range []string{"ignore", "customValidators"} {
		if a, ok := m[key].([]interface{}); ok {
			for _, v := range a {
				if s, ok := v.(string); ok {
					if key == "ignore" {
						o.ignore = append(o.ignore, s)
					} else {
						o.customValidators = append(o.customValidators, s)
					}
				}
			}
		}
	}
	o.skipUndeclared, _ = m["skipUndeclared"].(bool)
	return o
}

func keyName(n *ast.Node) string {
	if n == nil {
		return ""
	}
	if name, ok := utils.GetStaticPropertyName(n); ok {
		return name
	}
	return ""
}

func elementName(n *ast.Node) string {
	n = unwrap(n)
	if n == nil {
		return ""
	}
	switch n.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNullKeyword,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindRegularExpressionLiteral:
		return keyName(n)
	}
	return ""
}

func unwrap(n *ast.Node) *ast.Node { return reactutil.SkipExpressionWrappers(n) }

func propMap(n *ast.Node, customValidators []string) (map[string]propType, bool) {
	return propMapSeen(n, customValidators, map[*ast.Node]bool{}, true)
}

func propMapSeen(n *ast.Node, customValidators []string, seen map[*ast.Node]bool, topLevel bool) (map[string]propType, bool) {
	n = unwrap(n)
	if n == nil {
		return nil, false
	}
	if n.Kind == ast.KindObjectLiteralExpression {
		out := map[string]propType{}
		for _, p := range n.AsObjectLiteralExpression().Properties.Nodes {
			if p == nil {
				continue
			}
			if p.Kind == ast.KindSpreadAssignment {
				// A shape spread does not make every nested key valid. The
				// top-level opaque-declaration fallback is handled by declared.
				if topLevel {
					out["__ANY_KEY__"] = propType{any: true}
				}
				continue
			}
			var name, value *ast.Node
			switch p.Kind {
			case ast.KindPropertyAssignment:
				pa := p.AsPropertyAssignment()
				name, value = pa.Name(), pa.Initializer
			case ast.KindShorthandPropertyAssignment:
				name, value = p.AsShorthandPropertyAssignment().Name(), p.AsShorthandPropertyAssignment().Name()
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				name = p.Name()
				value = p
			default:
				continue
			}
			k := keyName(name)
			if k == "" {
				out["__ANY_KEY__"] = propType{any: true}
				continue
			}
			validator := validatorTypeSeen(value, customValidators, seen)
			if validator.invalid {
				out["__INVALID_VALIDATOR__"] = validator
				continue
			}
			out[k] = validator
		}
		return out, true
	}
	return nil, false
}

func validatorType(n *ast.Node, customValidators []string) propType {
	return validatorTypeSeen(n, customValidators, map[*ast.Node]bool{})
}

func validatorTypeSeen(n *ast.Node, customValidators []string, seen map[*ast.Node]bool) propType {
	n = unwrap(n)
	if n == nil {
		return propType{any: true}
	}
	if n.Kind == ast.KindIdentifier {
		if seen[n] {
			return propType{invalid: true}
		}
		seen[n] = true
		if initializer := reactutil.ResolveIdentifierInitializer(n, nil); initializer != nil && initializer != n {
			return validatorTypeSeen(initializer, customValidators, seen)
		}
		return propType{open: true}
	}
	if n.Kind == ast.KindPropertyAccessExpression && keyName(n.AsPropertyAccessExpression().Name()) == "isRequired" {
		n = unwrap(n.AsPropertyAccessExpression().Expression)
	}
	if n.Kind != ast.KindCallExpression {
		// Primitive and broad validators (for example PropTypes.object and
		// PropTypes.string) validate any property read beneath the prop.
		return propType{open: true}
	}
	call := n.AsCallExpression()
	callee := unwrap(call.Expression)
	name := ""
	if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
		name = keyName(callee.AsPropertyAccessExpression().Name())
		if object := unwrap(callee.AsPropertyAccessExpression().Expression); object != nil && object.Kind == ast.KindIdentifier && slices.Contains(customValidators, object.AsIdentifier().Text) {
			return propType{open: true}
		}
	}
	if name == "shape" || name == "exact" {
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			if m, ok := propMapSeen(call.Arguments.Nodes[0], customValidators, seen, false); ok {
				if m["__INVALID_VALIDATOR__"].invalid {
					return propType{invalid: true}
				}
				return propType{children: m}
			}
		}
	}
	if name == "objectOf" || name == "arrayOf" {
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return propType{children: map[string]propType{"__ANY_KEY__": validatorTypeSeen(call.Arguments.Nodes[0], customValidators, seen)}}
		}
	}
	if name == "oneOfType" && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
		argument := unwrap(call.Arguments.Nodes[0])
		if argument != nil && argument.Kind == ast.KindArrayLiteralExpression {
			var union []propType
			for _, candidate := range argument.AsArrayLiteralExpression().Elements.Nodes {
				if candidate != nil && candidate.Kind != ast.KindOmittedExpression {
					union = append(union, validatorTypeSeen(candidate, customValidators, seen))
				}
			}
			if len(union) > 0 {
				return propType{union: union}
			}
		}
	}
	// Calls whose structure this port does not model (instanceOf, oneOf, or a
	// dynamic shape argument) are broad validators upstream.
	return propType{open: true}
}

func declared(n *ast.Node, customValidators []string, propWrappers []reactutil.PropWrapperEntry) (map[string]propType, bool) {
	n = unwrap(n)
	if n != nil && n.Kind == ast.KindIdentifier {
		if initializer := reactutil.ResolveIdentifierInitializer(n, nil); initializer != nil && initializer != n {
			return declared(initializer, customValidators, propWrappers)
		}
	}
	if n != nil && n.Kind == ast.KindCallExpression && reactutil.IsPropWrapperCall(n, propWrappers) {
		call := n.AsCallExpression()
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return declared(call.Arguments.Nodes[0], customValidators, propWrappers)
		}
	}
	m, ok := propMap(n, customValidators)
	return m, ok
}

func visibleTypeAlias(use *ast.Node, candidates []*ast.Node) *ast.Node {
	var best *ast.Node
	bestDepth := -1
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		for scope, depth := use, 0; scope != nil; scope, depth = scope.Parent, depth+1 {
			if scope.Kind != ast.KindBlock && scope.Kind != ast.KindSourceFile && scope.Kind != ast.KindModuleBlock {
				continue
			}
			for declarationScope := candidate.Parent; declarationScope != nil; declarationScope = declarationScope.Parent {
				if declarationScope == scope {
					if best == nil || depth < bestDepth {
						best, bestDepth = candidate, depth
					}
					break
				}
			}
		}
	}
	return best
}

func declaredType(n *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node, depth int) (map[string]propType, bool) {
	if n == nil || depth == 0 {
		return nil, false
	}
	var members *ast.TypeElementList
	switch n.Kind {
	case ast.KindTypeLiteral:
		members = n.AsTypeLiteralNode().Members
	case ast.KindInterfaceDeclaration:
		members = n.AsInterfaceDeclaration().Members
	case ast.KindTypeAliasDeclaration:
		return declaredType(n.AsTypeAliasDeclaration().Type, aliases, aliasesBySymbol, depth-1)
	case ast.KindTypeReference:
		name := reactutil.EntityNameRightmost(n.AsTypeReferenceNode().TypeName)
		if name != nil {
			if symbol := name.Symbol(); symbol != nil {
				if declaration := aliasesBySymbol[symbol]; declaration != nil {
					return declaredType(declaration, aliases, aliasesBySymbol, depth-1)
				}
			}
			if declaration := visibleTypeAlias(n, aliases[name.AsIdentifier().Text]); declaration != nil {
				return declaredType(declaration, aliases, aliasesBySymbol, depth-1)
			}
		}
		// Imported or otherwise unresolved annotations are opaque upstream and
		// must not be treated as if no declaration were present.
		return map[string]propType{"__ANY_KEY__": {any: true}}, true
	case ast.KindIdentifier:
		if symbol := n.Symbol(); symbol != nil {
			if declaration := aliasesBySymbol[symbol]; declaration != nil {
				return declaredType(declaration, aliases, aliasesBySymbol, depth-1)
			}
		}
		if declaration := visibleTypeAlias(n, aliases[n.AsIdentifier().Text]); declaration != nil {
			return declaredType(declaration, aliases, aliasesBySymbol, depth-1)
		}
		return map[string]propType{"__ANY_KEY__": {any: true}}, true
	case ast.KindParenthesizedType:
		return declaredType(n.AsParenthesizedTypeNode().Type, aliases, aliasesBySymbol, depth-1)
	case ast.KindIntersectionType:
		declared := map[string]propType{}
		found := false
		for _, part := range n.AsIntersectionTypeNode().Types.Nodes {
			if members, ok := declaredType(part, aliases, aliasesBySymbol, depth-1); ok {
				for name, prop := range members {
					declared[name] = prop
				}
				found = true
			}
		}
		return declared, found
	default:
		return nil, false
	}
	if members == nil {
		return map[string]propType{}, true
	}
	declared := map[string]propType{}
	for _, member := range members.Nodes {
		if member == nil {
			continue
		}
		switch member.Kind {
		case ast.KindPropertySignature:
			property := member.AsPropertySignatureDeclaration()
			if property != nil && property.Type != nil {
				if name := keyName(property.Name()); name != "" {
					declared[name] = propTypeFromType(property.Type.AsNode(), aliases, aliasesBySymbol, depth-1)
				}
			}
		case ast.KindMethodSignature:
			if name := keyName(member.AsMethodSignatureDeclaration().Name()); name != "" {
				declared[name] = propType{open: true}
			}
		}
	}
	if n.Kind == ast.KindInterfaceDeclaration {
		heritageClauses := n.AsInterfaceDeclaration().HeritageClauses
		if heritageClauses == nil {
			return declared, true
		}
		for _, clause := range heritageClauses.Nodes {
			if clause == nil || clause.Kind != ast.KindHeritageClause {
				continue
			}
			for _, heritage := range clause.AsHeritageClause().Types.Nodes {
				if heritage != nil && heritage.Kind == ast.KindExpressionWithTypeArguments {
					if base, ok := declaredType(heritage.AsExpressionWithTypeArguments().Expression, aliases, aliasesBySymbol, depth-1); ok {
						mergeDeclared(declared, base)
					}
				}
			}
		}
	}
	return declared, true
}

func propTypeFromType(n *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node, depth int) propType {
	if children, ok := declaredType(n, aliases, aliasesBySymbol, depth); ok {
		return propType{children: children}
	}
	return propType{open: true}
}

func classPropsType(class *ast.Node) *ast.Node {
	if class == nil {
		return nil
	}
	var clauses *ast.HeritageClauseList
	switch class.Kind {
	case ast.KindClassDeclaration:
		clauses = class.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		clauses = class.AsClassExpression().HeritageClauses
	default:
		return nil
	}
	if clauses == nil {
		return nil
	}
	for _, clause := range clauses.Nodes {
		if clause == nil || clause.Kind != ast.KindHeritageClause {
			continue
		}
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typ := range types.Nodes {
			if typ == nil || typ.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			arguments := typ.AsExpressionWithTypeArguments().TypeArguments
			if arguments != nil && len(arguments.Nodes) > 0 {
				return arguments.Nodes[0]
			}
		}
	}
	return nil
}

func componentVariableType(n *ast.Node) *ast.Node {
	for current := n; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
			current = parent
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current && parent.AsVariableDeclaration().Type != nil {
				return parent.AsVariableDeclaration().Type.AsNode()
			}
		}
		return nil
	}
	return nil
}

func reactComponentTypeArgument(typeNode *ast.Node) *ast.Node {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return nil
	}
	name := reactutil.EntityNameRightmost(typeNode.AsTypeReferenceNode().TypeName)
	if name == nil {
		return nil
	}
	arguments := typeNode.AsTypeReferenceNode().TypeArguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return nil
	}
	if name.AsIdentifier().Text == "ForwardRefRenderFunction" && len(arguments.Nodes) >= 2 {
		return arguments.Nodes[1]
	}
	if !slices.Contains([]string{"FC", "FunctionComponent", "SFC", "StatelessComponent", "ComponentType", "VFC", "VoidFunctionComponent"}, name.AsIdentifier().Text) {
		return nil
	}
	return arguments.Nodes[0]
}

func forwardRefPropsType(n *ast.Node) *ast.Node {
	for current := n; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if parent.Kind == ast.KindCallExpression {
			call := parent.AsCallExpression()
			if call.Arguments == nil || len(call.Arguments.Nodes) == 0 || unwrap(call.Arguments.Nodes[0]) != unwrap(current) {
				return nil
			}
			callee := unwrap(call.Expression)
			isForwardRef := callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "forwardRef"
			if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
				isForwardRef = keyName(callee.AsPropertyAccessExpression().Name()) == "forwardRef"
			}
			if !isForwardRef {
				return nil
			}
			arguments := call.TypeArguments
			if arguments != nil && len(arguments.Nodes) >= 2 {
				return arguments.Nodes[1]
			}
			return nil
		}
		if parent.Kind != ast.KindParenthesizedExpression && parent.Kind != ast.KindAsExpression && parent.Kind != ast.KindSatisfiesExpression && parent.Kind != ast.KindNonNullExpression && parent.Kind != ast.KindTypeAssertionExpression {
			return nil
		}
	}
	return nil
}

func className(n *ast.Node) string { return reactutil.BindingIdentifierName(n) }

func componentName(n *ast.Node) string {
	if name := className(n); name != "" {
		return name
	}
	var suffix []string
	for current := n; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindNonNullExpression,
			ast.KindTypeAssertionExpression,
			ast.KindCallExpression:
			current = parent
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				name := parent.AsVariableDeclaration().Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return strings.Join(append([]string{name.AsIdentifier().Text}, suffix...), ".")
				}
			}
		case ast.KindBinaryExpression:
			assignment := parent.AsBinaryExpression()
			if assignment.OperatorToken != nil && assignment.OperatorToken.Kind == ast.KindEqualsToken && assignment.Right == current {
				if name := componentReference(assignment.Left); name != "" {
					return strings.Join(append([]string{name}, suffix...), ".")
				}
			}
		case ast.KindPropertyAssignment:
			assignment := parent.AsPropertyAssignment()
			if assignment.Initializer == current {
				if name := keyName(assignment.Name()); name != "" {
					suffix = append([]string{name}, suffix...)
					current = parent
					continue
				}
			}
		case ast.KindObjectLiteralExpression:
			current = parent
			continue
		}
		return ""
	}
	return ""
}

func componentBinding(n *ast.Node) *ast.Symbol {
	if n == nil {
		return nil
	}
	if n.Kind == ast.KindFunctionDeclaration || n.Kind == ast.KindClassDeclaration {
		return n.Symbol()
	}
	for current := n; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
			current = parent
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				return parent.Symbol()
			}
			return nil
		default:
			return nil
		}
	}
	return nil
}

func setDeclared(m map[string]propType, names []string, value propType) {
	if len(names) == 0 {
		return
	}
	if len(names) == 1 {
		m[names[0]] = value
		return
	}
	p := m[names[0]]
	if p.children == nil {
		p.children = map[string]propType{}
	}
	setDeclared(p.children, names[1:], value)
	m[names[0]] = p
}

func mergeDeclared(dst, src map[string]propType) {
	for name, value := range src {
		if old, ok := dst[name]; ok && old.children != nil && value.children != nil {
			mergeDeclared(old.children, value.children)
			value.children = old.children
		}
		dst[name] = value
	}
}

func addDestructured(c *component, pattern *ast.Node, prefix []string) {
	if c == nil || pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return
	}
	for _, e := range pattern.AsBindingPattern().Elements.Nodes {
		if e == nil || e.Kind != ast.KindBindingElement {
			continue
		}
		be := e.AsBindingElement()
		if be.DotDotDotToken != nil {
			continue
		}
		key := keyName(be.PropertyName)
		if key == "" {
			key = keyName(be.Name())
		}
		if key == "" || be.Name() == nil {
			continue
		}
		path := append(append([]string{}, prefix...), key)
		// eslint-plugin-react records a destructuring binding as a props use even
		// when its local name is never read afterwards. Keep the authored property
		// node so the diagnostic uses the same terminal range as an ESTree pattern.
		report := be.Name()
		if be.PropertyName != nil {
			report = be.PropertyName
		}
		appendUse(c, report, path)
		if be.Name().Kind == ast.KindIdentifier {
			c.destructured[be.Name().AsIdentifier().Text] = path
		} else if be.Name().Kind == ast.KindObjectBindingPattern {
			addDestructured(c, be.Name(), path)
		}
	}
}

func addThisPropsDestructured(c *component, pattern *ast.Node) bool {
	if c == nil || pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return false
	}
	for _, e := range pattern.AsBindingPattern().Elements.Nodes {
		if e == nil || e.Kind != ast.KindBindingElement {
			continue
		}
		be := e.AsBindingElement()
		key := keyName(be.PropertyName)
		if key == "" {
			key = keyName(be.Name())
		}
		if key != "props" || be.Name() == nil {
			continue
		}
		switch be.Name().Kind {
		case ast.KindObjectBindingPattern:
			addDestructured(c, be.Name(), nil)
		case ast.KindIdentifier:
			c.destructured[be.Name().AsIdentifier().Text] = nil
		}
		return true
	}
	return false
}

func propsPath(root *ast.Node, names []string, c *component) ([]string, bool) {
	root = unwrap(root)
	if root == nil || c == nil {
		return nil, false
	}
	if root.Kind == ast.KindThisKeyword {
		if len(names) > 0 && names[0] == "props" {
			return names[1:], true
		}
		return nil, false
	}
	if root.Kind != ast.KindIdentifier {
		return nil, false
	}
	if root.AsIdentifier().Text == c.props || isClassPropsParameter(root, c) {
		return names, true
	}
	if prefix, ok := c.destructured[root.AsIdentifier().Text]; ok {
		return append(append([]string{}, prefix...), names...), true
	}
	return nil, false
}

func isClassPropsParameter(ident *ast.Node, c *component) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier || c == nil || (c.node.Kind != ast.KindClassDeclaration && c.node.Kind != ast.KindClassExpression) {
		return false
	}
	for current := ident.Parent; current != nil && current != c.node; current = current.Parent {
		switch current.Kind {
		case ast.KindConstructor:
			return functionFirstParameterName(current) == ident.AsIdentifier().Text
		case ast.KindMethodDeclaration:
			name := keyName(current.Name())
			switch name {
			case "componentWillReceiveProps", "shouldComponentUpdate", "componentWillUpdate", "componentDidUpdate",
				"getDerivedStateFromProps", "getSnapshotBeforeUpdate", "UNSAFE_componentWillReceiveProps", "UNSAFE_componentWillUpdate":
				return functionFirstParameterName(current) == ident.AsIdentifier().Text
			default:
				return false
			}
		case ast.KindArrowFunction, ast.KindFunctionExpression:
			return isSetStatePropsParameter(current, ident)
		case ast.KindFunctionDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			return false
		}
	}
	return false
}

func isSetStatePropsParameter(fn, ident *ast.Node) bool {
	if fn == nil || ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	params := reactutil.FunctionParameters(fn)
	if len(params) < 2 || params[1] == nil || params[1].Kind != ast.KindParameter || functionParameterName(params[1]) != ident.AsIdentifier().Text {
		return false
	}
	call := reactutil.SkipExpressionWrappersUp(fn)
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	invocation := call.AsCallExpression()
	if invocation.Arguments == nil || len(invocation.Arguments.Nodes) == 0 || unwrap(invocation.Arguments.Nodes[0]) != fn {
		return false
	}
	callee := unwrap(invocation.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression || keyName(callee.AsPropertyAccessExpression().Name()) != "setState" {
		return false
	}
	receiver := unwrap(callee.AsPropertyAccessExpression().Expression)
	return receiver != nil && receiver.Kind == ast.KindThisKeyword
}

func functionFirstParameterName(fn *ast.Node) string {
	params := reactutil.FunctionParameters(fn)
	if len(params) == 0 || params[0] == nil || params[0].Kind != ast.KindParameter {
		return ""
	}
	return functionParameterName(params[0])
}

func functionParameterName(param *ast.Node) string {
	name := param.AsParameterDeclaration().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

func isGeneratorFunction(n *ast.Node) bool {
	if n == nil {
		return false
	}
	switch n.Kind {
	case ast.KindFunctionDeclaration:
		return n.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return n.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return n.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

func getterReturn(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	stmts := body.AsBlock().Statements
	if stmts == nil {
		return nil
	}
	for i := len(stmts.Nodes) - 1; i >= 0; i-- {
		if stmt := stmts.Nodes[i]; stmt != nil && stmt.Kind == ast.KindReturnStatement {
			return stmt.AsReturnStatement().Expression
		}
	}
	return nil
}

func componentFor(node *ast.Node, comps []*component) *component {
	var best *component
	for p := node; p != nil; p = p.Parent {
		for _, c := range comps {
			if c.node == p {
				best = c
				break
			}
		}
		if best != nil {
			return best
		}
	}
	return nil
}

func appendUse(c *component, node *ast.Node, names []string) {
	if c != nil && len(names) > 0 {
		c.used = append(c.used, use{node: node, names: names})
	}
}

// memberReportNode matches the ESTree MemberExpression report range used by
// eslint-plugin-react: diagnostics attach to the terminal property, rather
// than to the complete member chain.
func memberReportNode(n *ast.Node) *ast.Node {
	n = unwrap(n)
	if n == nil {
		return n
	}
	switch n.Kind {
	case ast.KindPropertyAccessExpression:
		return n.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		return n.AsElementAccessExpression().ArgumentExpression
	default:
		return n
	}
}

func memberNames(n *ast.Node) (*ast.Node, []string, bool) {
	var names []string
	cur := unwrap(n)
	var report *ast.Node
	for cur != nil {
		switch cur.Kind {
		case ast.KindPropertyAccessExpression:
			pa := cur.AsPropertyAccessExpression()
			k := keyName(pa.Name())
			if k == "" {
				return nil, nil, false
			}
			names = append([]string{k}, names...)
			report = pa.Name()
			cur = unwrap(pa.Expression)
		case ast.KindElementAccessExpression:
			ea := cur.AsElementAccessExpression()
			k := elementName(ea.ArgumentExpression)
			if k == "" {
				if argument := unwrap(ea.ArgumentExpression); argument != nil && argument.Kind == ast.KindBigIntLiteral {
					return nil, nil, false
				}
				k = "__COMPUTED_PROP__"
			}
			names = append([]string{k}, names...)
			report = ea.ArgumentExpression
			cur = unwrap(ea.Expression)
		default:
			return cur, names, true
		}
	}
	return report, names, false
}

func componentReference(n *ast.Node) string {
	root, names, ok := memberNames(n)
	if !ok || root == nil || root.Kind != ast.KindIdentifier {
		return ""
	}
	parts := append([]string{root.AsIdentifier().Text}, names...)
	return strings.Join(parts, ".")
}

func isAssignmentTarget(n *ast.Node) bool {
	for n != nil && n.Parent != nil {
		parent := n.Parent
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression == n {
				n = parent
				continue
			}
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression == n {
				n = parent
				continue
			}
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			n = parent
			continue
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left == n
		}
		return false
	}
	return false
}

func rootMatches(n *ast.Node, c *component) bool {
	n = unwrap(n)
	if n == nil {
		return false
	}
	if n.Kind == ast.KindThisKeyword {
		return true
	}
	if n.Kind != ast.KindIdentifier {
		return false
	}
	if n.AsIdentifier().Text == c.props || isClassPropsParameter(n, c) {
		return true
	}
	_, ok := c.destructured[n.AsIdentifier().Text]
	return ok
}

func isBindingName(n *ast.Node) bool {
	if n == nil || n.Parent == nil {
		return false
	}
	switch n.Parent.Kind {
	case ast.KindBindingElement:
		return n.Parent.AsBindingElement().Name() == n
	case ast.KindVariableDeclaration:
		return n.Parent.AsVariableDeclaration().Name() == n
	}
	return false
}

func isDeclared(m map[string]propType, names []string) bool {
	if len(names) == 0 {
		return true
	}
	p, ok := m[names[0]]
	if !ok {
		p, ok = m["__ANY_KEY__"]
	}
	if !ok {
		return false
	}
	return propTypeDeclares(p, names[1:])
}

func propTypeDeclares(p propType, remaining []string) bool {
	if len(remaining) == 0 || p.any || p.open {
		return true
	}
	for _, candidate := range p.union {
		if propTypeDeclares(candidate, remaining) {
			return true
		}
	}
	if p.children == nil {
		return false
	}
	return isDeclared(p.children, remaining)
}

func reportName(names []string) string {
	return strings.ReplaceAll(strings.Join(names, "."), ".__COMPUTED_PROP__", "[]")
}

func isObjectPrototypeProperty(name string) bool {
	switch name {
	case "__defineGetter__", "__defineSetter__", "hasOwnProperty", "__lookupGetter__", "__lookupSetter__",
		"isPrototypeOf", "propertyIsEnumerable", "toString", "valueOf", "__proto__", "toLocaleString", "constructor":
		return true
	}
	return false
}

var PropTypesRule = rule.Rule{
	Name: "react/prop-types", Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, raw []any) rule.RuleListeners {
		o := parseOptions(raw)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		propWrappers := reactutil.GetPropWrapperFunctions(ctx.Settings)
		typeAliases := map[string][]*ast.Node{}
		typeAliasesBySymbol := map[*ast.Symbol]*ast.Node{}
		var collectTypeAliases ast.Visitor
		collectTypeAliases = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case ast.KindTypeAliasDeclaration:
				if name := n.AsTypeAliasDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
					typeAliases[name.AsIdentifier().Text] = append(typeAliases[name.AsIdentifier().Text], n)
					if symbol := name.Symbol(); symbol != nil {
						typeAliasesBySymbol[symbol] = n
					}
					if symbol := n.Symbol(); symbol != nil {
						typeAliasesBySymbol[symbol] = n
					}
				}
			case ast.KindInterfaceDeclaration:
				if name := n.AsInterfaceDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
					typeAliases[name.AsIdentifier().Text] = append(typeAliases[name.AsIdentifier().Text], n)
					if symbol := name.Symbol(); symbol != nil {
						typeAliasesBySymbol[symbol] = n
					}
					if symbol := n.Symbol(); symbol != nil {
						typeAliasesBySymbol[symbol] = n
					}
				}
			}
			n.ForEachChild(collectTypeAliases)
			return false
		}
		ctx.SourceFile.Node.ForEachChild(collectTypeAliases)
		var comps []*component
		byName := map[string]*component{}
		byBinding := map[*ast.Symbol]*component{}
		var propTypeAssignments []*ast.Node
		var walk ast.Visitor
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if reactutil.ExtendsReactComponent(n, pragma) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
					c.binding = componentBinding(n)
					if c.binding != nil {
						byBinding[c.binding] = c
					}
					if declared, ok := declaredType(classPropsType(n), typeAliases, typeAliasesBySymbol, 8); ok {
						c.declared = declared
						c.declaredBlock = true
					}
					if name := componentName(n); name != "" {
						byName[name] = c
					}
				}
			case ast.KindObjectLiteralExpression:
				if reactutil.IsCreateReactClassObjectArg(n, pragma, createClass) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
					c.binding = componentBinding(n)
					if c.binding != nil {
						byBinding[c.binding] = c
					}
					if name := componentName(n); name != "" {
						byName[name] = c
					}
				}
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
				if !isGeneratorFunction(n) && reactutil.IsStatelessReactComponentWithWrappers(n, pragma, ctx.TypeChecker, wrappers) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
					c.binding = componentBinding(n)
					if c.binding != nil {
						byBinding[c.binding] = c
					}
					if name := componentName(n); name != "" {
						byName[name] = c
					}
					ps := n.Parameters()
					if len(ps) > 0 && ps[0].Kind == ast.KindParameter {
						parameter := ps[0].AsParameterDeclaration()
						name := parameter.Name()
						if name != nil && name.Kind == ast.KindIdentifier {
							c.props = name.AsIdentifier().Text
						} else if name != nil && name.Kind == ast.KindObjectBindingPattern {
							addDestructured(c, name, nil)
						}
						if parameter.Type != nil {
							if declared, ok := declaredType(parameter.Type.AsNode(), typeAliases, typeAliasesBySymbol, 8); ok {
								c.declared = declared
								c.declaredBlock = true
							}
						} else if declared, ok := declaredType(reactComponentTypeArgument(componentVariableType(n)), typeAliases, typeAliasesBySymbol, 8); ok {
							c.declared = declared
							c.declaredBlock = true
						} else if declared, ok := declaredType(forwardRefPropsType(n), typeAliases, typeAliasesBySymbol, 8); ok {
							c.declared = declared
							c.declaredBlock = true
						}
					}
				}
				if n.Kind == ast.KindMethodDeclaration {
					if c := componentFor(n.Parent, comps); c != nil {
						name := keyName(n.AsMethodDeclaration().Name())
						switch name {
						case "componentWillReceiveProps", "shouldComponentUpdate", "componentWillUpdate", "componentDidUpdate", "getDerivedStateFromProps", "getSnapshotBeforeUpdate", "UNSAFE_componentWillReceiveProps", "UNSAFE_componentWillUpdate":
							params := reactutil.FunctionParameters(n)
							if len(params) > 0 && params[0] != nil && params[0].Kind == ast.KindParameter {
								if pattern := params[0].AsParameterDeclaration().Name(); pattern != nil && pattern.Kind == ast.KindObjectBindingPattern {
									addDestructured(c, pattern, nil)
								}
							}
						}
					}
				}
			case ast.KindPropertyAssignment:
				pa := n.AsPropertyAssignment()
				if keyName(pa.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						if m, ok := declared(pa.Initializer, o.customValidators, propWrappers); ok {
							c.declared = m
							c.declaredBlock = true
						} else {
							c.declaredBlock = true
							c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
						}
					}
				}
			case ast.KindPropertyDeclaration:
				pd := n.AsPropertyDeclaration()
				if keyName(pd.Name()) == "props" && !ast.HasSyntacticModifier(n, ast.ModifierFlagsStatic) {
					if c := componentFor(n.Parent, comps); c != nil && pd.Type != nil {
						if declared, ok := declaredType(pd.Type.AsNode(), typeAliases, typeAliasesBySymbol, 8); ok {
							c.declared = declared
							c.declaredBlock = true
						}
					}
				}
				if keyName(pd.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						if m, ok := declared(pd.Initializer, o.customValidators, propWrappers); ok {
							c.declared = m
							c.declaredBlock = true
						} else {
							c.declaredBlock = true
							c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
						}
					}
				}
			case ast.KindGetAccessor:
				ga := n.AsGetAccessorDeclaration()
				if ast.HasSyntacticModifier(n, ast.ModifierFlagsStatic) && keyName(ga.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						if m, ok := declared(getterReturn(ga.Body), o.customValidators, propWrappers); ok {
							c.declared = m
							c.declaredBlock = true
						}
					}
				}
			case ast.KindVariableDeclaration:
				vd := n.AsVariableDeclaration()
				if vd.Name() != nil {
					if root, names, ok := memberNames(vd.Initializer); ok && root != nil {
						if c := componentFor(n, comps); c != nil {
							if root.Kind == ast.KindThisKeyword && len(names) == 0 && vd.Name().Kind == ast.KindObjectBindingPattern {
								addThisPropsDestructured(c, vd.Name())
							}
							if path, ok := propsPath(root, names, c); ok {
								switch vd.Name().Kind {
								case ast.KindObjectBindingPattern:
									addDestructured(c, vd.Name(), path)
								case ast.KindIdentifier:
									c.destructured[vd.Name().AsIdentifier().Text] = path
								}
							}
						}
					}
				}
			case ast.KindBinaryExpression:
				b := n.AsBinaryExpression()
				if b.OperatorToken != nil && b.OperatorToken.Kind == ast.KindEqualsToken && b.Left != nil && b.Left.Kind == ast.KindPropertyAccessExpression {
					root, names, ok := memberNames(b.Left)
					if ok && root != nil && root.Kind == ast.KindIdentifier {
						propTypesIndex := slices.Index(names, "propTypes")
						if propTypesIndex < 0 {
							break
						}
						propTypeAssignments = append(propTypeAssignments, n)
						componentParts := append([]string{root.AsIdentifier().Text}, names[:propTypesIndex]...)
						c := byName[strings.Join(componentParts, ".")]
						if propTypesIndex == 0 {
							symbol := root.Symbol()
							if symbol == nil && ctx.TypeChecker != nil {
								symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
							}
							if bySymbol := byBinding[symbol]; bySymbol != nil {
								c = bySymbol
							}
						}
						if c != nil {
							propNames := names[propTypesIndex+1:]
							if len(propNames) == 0 {
								if m, ok := declared(b.Right, o.customValidators, propWrappers); ok {
									mergeDeclared(c.declared, m)
								} else {
									c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
								}
							} else {
								setDeclared(c.declared, propNames, validatorType(b.Right, o.customValidators))
							}
							c.declaredBlock = true
						}
					}
				}
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				if isAssignmentTarget(n) {
					break
				}
				root, names, ok := memberNames(n)
				if !ok || root == nil {
					break
				}
				if len(names) == 0 {
					break
				}
				// A nested component may capture its enclosing component's props.
				// Attribute a use to every containing component whose own props
				// binding matches the member root, rather than only the nearest one.
				for _, c := range comps {
					contains := false
					for parent := n; parent != nil; parent = parent.Parent {
						if parent == c.node {
							contains = true
							break
						}
					}
					if contains && rootMatches(root, c) {
						if path, ok := propsPath(root, names, c); ok {
							appendUse(c, memberReportNode(n), path)
						}
					}
				}
			}
			n.ForEachChild(walk)
			return false
		}
		ctx.SourceFile.Node.ForEachChild(walk)
		// A component declaration can be hoisted below its propTypes assignment.
		// Replay assignments after collection so declaration order does not decide
		// whether the component is known.
		for _, assignmentNode := range propTypeAssignments {
			assignment := assignmentNode.AsBinaryExpression()
			root, names, ok := memberNames(assignment.Left)
			if !ok || root == nil || root.Kind != ast.KindIdentifier {
				continue
			}
			propTypesIndex := slices.Index(names, "propTypes")
			if propTypesIndex < 0 {
				continue
			}
			componentParts := append([]string{root.AsIdentifier().Text}, names[:propTypesIndex]...)
			c := byName[strings.Join(componentParts, ".")]
			if propTypesIndex == 0 {
				symbol := root.Symbol()
				if symbol == nil && ctx.TypeChecker != nil {
					symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
				}
				if bySymbol := byBinding[symbol]; bySymbol != nil {
					c = bySymbol
				}
			}
			if c == nil {
				continue
			}
			propNames := names[propTypesIndex+1:]
			if len(propNames) == 0 {
				if m, ok := declared(assignment.Right, o.customValidators, propWrappers); ok {
					mergeDeclared(c.declared, m)
				} else {
					c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
				}
			} else {
				setDeclared(c.declared, propNames, validatorType(assignment.Right, o.customValidators))
			}
			c.declaredBlock = true
		}
		for _, c := range comps {
			if c.ignore || len(c.used) == 0 || !c.declaredBlock && o.skipUndeclared {
				continue
			}
			slices.SortFunc(c.used, func(a, b use) int {
				if position := a.node.Pos() - b.node.Pos(); position != 0 {
					return position
				}
				return len(a.names) - len(b.names)
			})
			for _, u := range c.used {
				name := u.names[0]
				if isObjectPrototypeProperty(u.names[len(u.names)-1]) || slices.Contains(o.ignore, name) {
					continue
				}
				if !isDeclared(c.declared, u.names) {
					ctx.ReportNode(u.node, rule.RuleMessage{Id: "missingPropType", Description: fmt.Sprintf("'%s' is missing in props validation", reportName(u.names)), Data: map[string]string{"name": reportName(u.names)}})
				}
			}
		}
		return rule.RuleListeners{}
	},
}
