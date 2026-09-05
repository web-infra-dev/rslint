package prop_types

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
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
type propAlias struct {
	path           []string
	crossClassSafe bool
}
type component struct {
	node             *ast.Node
	binding          *ast.Symbol
	declared         map[string]propType
	used             []use
	declaredBlock    bool
	ignoreValidation bool
}

type componentDeclarationEvent struct {
	source         *ast.Node
	typeNode       *ast.Node
	component      *component
	declaration    propDeclaration
	propNames      []string
	validator      propType
	replaceOpacity bool
	final          bool
}

type initializerResolver func(*ast.Node) (*ast.Node, bool)

type propDeclaration struct {
	props  map[string]propType
	opaque bool
}

func concreteDeclaration(props map[string]propType) propDeclaration {
	return propDeclaration{props: props}
}

func opaqueDeclaration() propDeclaration {
	return propDeclaration{props: map[string]propType{}, opaque: true}
}

func mergeDeclaration(dst *propDeclaration, src propDeclaration) {
	mergeDeclared(dst.props, src.props)
	dst.opaque = dst.opaque || src.opaque
}

func applyDeclared(c *component, declared propDeclaration, replaceOpacity bool) {
	if c == nil {
		return
	}
	if replaceOpacity {
		c.ignoreValidation = declared.opaque
	} else if declared.opaque {
		c.ignoreValidation = true
	}
	// Complete declarations accumulate distinct top-level props, but a later
	// declaration replaces the validator for the same prop as a whole. In
	// particular, two successive shapes must not retain each other's children.
	for name, value := range declared.props {
		c.declared[name] = value
	}
}

type componentKey struct {
	name    string
	binding *ast.Symbol
	scope   *ast.Node
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

func propMap(n *ast.Node, customValidators []string, resolve initializerResolver) (map[string]propType, bool) {
	return propMapSeen(n, customValidators, map[*ast.Node]bool{}, 0, resolve)
}

func propMapSeen(n *ast.Node, customValidators []string, seen map[*ast.Node]bool, compositeDepth int, resolve initializerResolver) (map[string]propType, bool) {
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
			if k == "" && name != nil && name.Kind == ast.KindComputedPropertyName {
				k = keyName(name.AsComputedPropertyName().Expression)
			}
			if k == "" {
				continue
			}
			validator := validatorTypeSeen(value, customValidators, seen, compositeDepth, resolve)
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

func validatorType(n *ast.Node, customValidators []string, resolve initializerResolver) propType {
	return validatorTypeSeen(n, customValidators, map[*ast.Node]bool{}, 0, resolve)
}

func validatorTypeSeen(n *ast.Node, customValidators []string, seen map[*ast.Node]bool, compositeDepth int, resolve initializerResolver) propType {
	n = unwrap(n)
	if n == nil {
		return propType{any: true}
	}
	if n.Kind == ast.KindIdentifier {
		if seen[n] {
			if compositeDepth == 0 {
				// Plain alias cycles still declare the containing prop upstream.
				// Structured recursive validators fail the whole validator instead.
				return propType{open: true}
			}
			return propType{invalid: true}
		}
		seen[n] = true
		defer delete(seen, n)
		var initializer *ast.Node
		if resolve != nil {
			initializer, _ = resolve(n)
		} else {
			initializer = reactutil.ResolveIdentifierInitializer(n, nil)
		}
		if initializer != nil && initializer != n {
			return validatorTypeSeen(initializer, customValidators, seen, compositeDepth, resolve)
		}
		return propType{open: true}
	}
	if n.Kind == ast.KindPropertyAccessExpression && keyName(n.AsPropertyAccessExpression().Name()) == "isRequired" {
		return validatorTypeSeen(n.AsPropertyAccessExpression().Expression, customValidators, seen, compositeDepth, resolve)
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
			if m, ok := propMapSeen(call.Arguments.Nodes[0], customValidators, seen, compositeDepth+1, resolve); ok {
				if m["__INVALID_VALIDATOR__"].invalid {
					return propType{invalid: true}
				}
				return propType{children: m}
			}
		}
	}
	if name == "objectOf" || name == "arrayOf" {
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return propType{children: map[string]propType{"__ANY_KEY__": validatorTypeSeen(call.Arguments.Nodes[0], customValidators, seen, compositeDepth+1, resolve)}}
		}
	}
	if name == "oneOfType" && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
		argument := unwrap(call.Arguments.Nodes[0])
		if argument != nil && argument.Kind == ast.KindArrayLiteralExpression {
			var union []propType
			for _, candidate := range argument.AsArrayLiteralExpression().Elements.Nodes {
				if candidate != nil && candidate.Kind != ast.KindOmittedExpression {
					union = append(union, validatorTypeSeen(candidate, customValidators, seen, compositeDepth+1, resolve))
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

func declared(n *ast.Node, customValidators []string, propWrappers []reactutil.PropWrapperEntry, resolve initializerResolver) (propDeclaration, bool) {
	return declaredSeen(n, customValidators, propWrappers, map[*ast.Node]bool{}, resolve)
}

func declaredSeen(n *ast.Node, customValidators []string, propWrappers []reactutil.PropWrapperEntry, seen map[*ast.Node]bool, resolve initializerResolver) (propDeclaration, bool) {
	n = unwrap(n)
	if n == nil {
		// A propTypes class field without an initializer is still a declaration,
		// but it declares no runtime validators (including with skipUndeclared).
		return concreteDeclaration(map[string]propType{}), true
	}
	if n.Kind == ast.KindIdentifier {
		if seen[n] {
			return concreteDeclaration(map[string]propType{}), true
		}
		seen[n] = true
		defer delete(seen, n)
		var initializer *ast.Node
		var found bool
		if resolve != nil {
			initializer, found = resolve(n)
		} else {
			initializer = reactutil.ResolveIdentifierInitializer(n, nil)
			found = initializer != nil
		}
		if initializer != nil && initializer != n {
			return declaredSeen(initializer, customValidators, propWrappers, seen, resolve)
		}
		if found {
			return concreteDeclaration(map[string]propType{}), true
		}
		return opaqueDeclaration(), true
	}
	if n.Kind == ast.KindCallExpression && reactutil.IsPropWrapperCall(n, propWrappers) {
		call := n.AsCallExpression()
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return declaredSeen(call.Arguments.Nodes[0], customValidators, propWrappers, seen, resolve)
		}
	}
	if n.Kind == ast.KindPropertyAccessExpression {
		// External declarations are commonly accessed through a namespace, such
		// as `RcSlider.propTypes`; upstream leaves those declarations opaque.
		return opaqueDeclaration(), true
	}
	if n.Kind == ast.KindObjectLiteralExpression {
		// Runtime spreads make this declaration opaque for its own component,
		// but must not install an any-key validator that can satisfy a nested
		// component through ancestor lookup.
		props, ok := propMap(n, customValidators, resolve)
		declaration := concreteDeclaration(props)
		for _, property := range n.AsObjectLiteralExpression().Properties.Nodes {
			if property != nil && property.Kind == ast.KindSpreadAssignment {
				declaration.opaque = true
				break
			}
		}
		return declaration, ok
	}
	if n.Kind == ast.KindCallExpression {
		// An unconfigured propTypes factory call does not make the declaration
		// opaque upstream: it leaves an explicitly empty validator table.
		return concreteDeclaration(map[string]propType{}), true
	}
	// Conditional, logical, null, function, array, and other unsupported full
	// declarations are opaque. Statically transparent TS wrappers were already
	// removed by unwrap above and intentionally retain rslint's existing detail.
	return opaqueDeclaration(), true
}

func visibleTypeAlias(use *ast.Node, candidates []*ast.Node) *ast.Node {
	visible := visibleTypeAliases(use, candidates)
	if len(visible) == 0 {
		return nil
	}
	return visible[0]
}

func visibleTypeAliases(use *ast.Node, candidates []*ast.Node) []*ast.Node {
	bestDepth := -1
	var visible []*ast.Node
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		declarationScope := typeDeclarationScope(candidate)
		if declarationScope == nil {
			continue
		}
		for scope, depth := use, 0; scope != nil; scope, depth = scope.Parent, depth+1 {
			if scope.Kind != ast.KindBlock && scope.Kind != ast.KindSourceFile && scope.Kind != ast.KindModuleBlock {
				continue
			}
			if declarationScope == scope {
				if bestDepth < 0 || depth < bestDepth {
					bestDepth = depth
					visible = []*ast.Node{candidate}
				} else if depth == bestDepth {
					visible = append(visible, candidate)
				}
				break
			}
		}
	}
	return visible
}

func typeDeclarationScope(n *ast.Node) *ast.Node {
	for parent := n.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock:
			return parent
		}
	}
	return nil
}

// lastReturnTypeExpression mirrors eslint-plugin-react's ast.loopNodes helper:
// scan top-level statements backwards, and when a switch is encountered recurse
// only into its final case. A non-empty trailing switch with no return in that
// case deliberately prevents falling back to earlier statements.
func lastReturnTypeExpression(statements []*ast.Node) *ast.Node {
	for i := len(statements) - 1; i >= 0; i-- {
		statement := statements[i]
		if statement == nil {
			continue
		}
		if statement.Kind == ast.KindReturnStatement {
			return statement.AsReturnStatement().Expression
		}
		if statement.Kind != ast.KindSwitchStatement {
			continue
		}
		switchStatement := statement.AsSwitchStatement()
		if switchStatement == nil || switchStatement.CaseBlock == nil {
			continue
		}
		caseBlock := switchStatement.CaseBlock.AsCaseBlock()
		if caseBlock == nil || caseBlock.Clauses == nil || len(caseBlock.Clauses.Nodes) == 0 {
			continue
		}
		lastClause := caseBlock.Clauses.Nodes[len(caseBlock.Clauses.Nodes)-1].AsCaseOrDefaultClause()
		if lastClause == nil || lastClause.Statements == nil {
			return nil
		}
		return lastReturnTypeExpression(lastClause.Statements.Nodes)
	}
	return nil
}

func returnTypeObjectDeclaration(object *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node, seen map[*ast.Node]bool, resolve initializerResolver) (propDeclaration, bool) {
	props, ok := propMap(object, nil, resolve)
	if !ok {
		return propDeclaration{}, false
	}
	result := concreteDeclaration(props)
	for _, property := range object.AsObjectLiteralExpression().Properties.Nodes {
		if property == nil || property.Kind != ast.KindSpreadAssignment {
			continue
		}
		// ReturnType keeps statically known generic spread properties, while
		// the spread still makes its owning component opaque. Keeping those two
		// facts separate prevents opacity from satisfying nested components.
		result.opaque = true
		expression := unwrap(property.AsSpreadAssignment().Expression)
		if expression == nil || expression.Kind != ast.KindCallExpression {
			continue
		}
		typeArguments := expression.AsCallExpression().TypeArguments
		if typeArguments == nil {
			continue
		}
		for _, argument := range typeArguments.Nodes {
			if declaration, ok := declaredType(argument, aliases, aliasesBySymbol, seen, resolve); ok {
				mergeDeclaration(&result, declaration)
			}
		}
	}
	return result, true
}

func returnTypeDeclaration(query *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node, seen map[*ast.Node]bool, resolve initializerResolver) (propDeclaration, bool) {
	if query == nil {
		return propDeclaration{}, false
	}
	if query.Kind == ast.KindFunctionType {
		return declaredType(query.AsFunctionTypeNode().Type.AsNode(), aliases, aliasesBySymbol, seen, resolve)
	}
	if query.Kind != ast.KindTypeQuery {
		return propDeclaration{}, false
	}
	expr := query.AsTypeQueryNode().ExprName
	if expr == nil || expr.Kind != ast.KindIdentifier {
		return propDeclaration{}, false
	}
	var initializer *ast.Node
	var found bool
	if resolve != nil {
		initializer, found = resolve(expr)
	} else {
		initializer = reactutil.ResolveIdentifierInitializer(expr, nil)
		found = initializer != nil
	}
	if initializer == nil {
		if found {
			return concreteDeclaration(map[string]propType{}), true
		}
		return propDeclaration{}, false
	}
	initializer = unwrap(initializer)
	if initializer == nil || (initializer.Kind != ast.KindArrowFunction && initializer.Kind != ast.KindFunctionExpression) {
		if found {
			return concreteDeclaration(map[string]propType{}), true
		}
		return propDeclaration{}, false
	}
	body := unwrap(initializer.Body())
	if body != nil && body.Kind == ast.KindBlock {
		block := body.AsBlock()
		if block == nil || block.Statements == nil {
			return concreteDeclaration(map[string]propType{}), true
		}
		body = unwrap(lastReturnTypeExpression(block.Statements.Nodes))
	}
	if body == nil {
		return concreteDeclaration(map[string]propType{}), true
	}
	if body.Kind == ast.KindObjectLiteralExpression {
		return returnTypeObjectDeclaration(body, aliases, aliasesBySymbol, seen, resolve)
	}
	if body.Kind == ast.KindCallExpression {
		typeArguments := body.AsCallExpression().TypeArguments
		if typeArguments == nil || len(typeArguments.Nodes) == 0 {
			return propDeclaration{}, false
		}
		declared := concreteDeclaration(map[string]propType{})
		for _, argument := range typeArguments.Nodes {
			if argumentDeclaration, ok := declaredType(argument, aliases, aliasesBySymbol, seen, resolve); ok {
				mergeDeclaration(&declared, argumentDeclaration)
			}
		}
		return declared, true
	}
	return concreteDeclaration(map[string]propType{}), true
}

func declaredType(n *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node, seen map[*ast.Node]bool, resolve initializerResolver) (propDeclaration, bool) {
	if n == nil {
		return propDeclaration{}, false
	}
	if seen[n] {
		return opaqueDeclaration(), true
	}
	seen[n] = true
	defer delete(seen, n)
	var members *ast.TypeElementList
	switch n.Kind {
	case ast.KindTypeLiteral:
		members = n.AsTypeLiteralNode().Members
	case ast.KindInterfaceDeclaration:
		members = n.AsInterfaceDeclaration().Members
	case ast.KindTypeAliasDeclaration:
		return declaredType(n.AsTypeAliasDeclaration().Type, aliases, aliasesBySymbol, seen, resolve)
	case ast.KindTypeReference:
		ref := n.AsTypeReferenceNode()
		if _, propsType := importedReactGenericPropsType(n); propsType != nil {
			declared, ok := declaredType(propsType, aliases, aliasesBySymbol, seen, resolve)
			if ok {
				declared.props["children"] = propType{open: true}
			}
			return declared, ok
		}
		// Non-React qualified references are external/opaque upstream. In
		// particular, Models.Props must not fall back to a visible local Props.
		if ref.TypeName != nil && ref.TypeName.Kind == ast.KindQualifiedName {
			return opaqueDeclaration(), true
		}
		if ref.TypeName != nil && reactutil.EntityNameRightmost(ref.TypeName) != nil && reactutil.EntityNameRightmost(ref.TypeName).AsIdentifier().Text == "ReturnType" && ref.TypeArguments != nil && len(ref.TypeArguments.Nodes) == 1 {
			if declared, ok := returnTypeDeclaration(ref.TypeArguments.Nodes[0], aliases, aliasesBySymbol, seen, resolve); ok {
				return declared, true
			}
		}
		name := reactutil.EntityNameRightmost(n.AsTypeReferenceNode().TypeName)
		if name != nil {
			// TypeScript declaration merging exposes one symbol for every interface
			// declaration. Preserve all same-name interface members, as upstream does.
			visible := visibleTypeAliases(n, aliases[name.AsIdentifier().Text])
			var interfaces []*ast.Node
			for _, candidate := range visible {
				if candidate != nil && candidate.Kind == ast.KindInterfaceDeclaration {
					interfaces = append(interfaces, candidate)
				}
			}
			if len(interfaces) > 1 {
				merged := concreteDeclaration(map[string]propType{})
				for _, candidate := range interfaces {
					if members, ok := declaredType(candidate, aliases, aliasesBySymbol, seen, resolve); ok {
						mergeDeclaration(&merged, members)
					}
				}
				return merged, true
			}
			if len(visible) == 1 {
				return declaredType(visible[0], aliases, aliasesBySymbol, seen, resolve)
			}
			if symbol := name.Symbol(); symbol != nil {
				if declaration := aliasesBySymbol[symbol]; declaration != nil {
					return declaredType(declaration, aliases, aliasesBySymbol, seen, resolve)
				}
			}
			if declaration := visibleTypeAlias(n, aliases[name.AsIdentifier().Text]); declaration != nil {
				return declaredType(declaration, aliases, aliasesBySymbol, seen, resolve)
			}
		}
		// Imported or otherwise unresolved annotations are opaque upstream and
		// must not be treated as if no declaration were present.
		return opaqueDeclaration(), true
	case ast.KindIdentifier:
		if symbol := n.Symbol(); symbol != nil {
			if declaration := aliasesBySymbol[symbol]; declaration != nil {
				return declaredType(declaration, aliases, aliasesBySymbol, seen, resolve)
			}
		}
		if declaration := visibleTypeAlias(n, aliases[n.AsIdentifier().Text]); declaration != nil {
			return declaredType(declaration, aliases, aliasesBySymbol, seen, resolve)
		}
		return opaqueDeclaration(), true
	case ast.KindParenthesizedType:
		return declaredType(n.AsParenthesizedTypeNode().Type, aliases, aliasesBySymbol, seen, resolve)
	case ast.KindIntersectionType:
		declared := concreteDeclaration(map[string]propType{})
		found := false
		for _, part := range n.AsIntersectionTypeNode().Types.Nodes {
			if members, ok := declaredType(part, aliases, aliasesBySymbol, seen, resolve); ok {
				mergeDeclaration(&declared, members)
				found = true
			}
		}
		return declared, found
	case ast.KindUnionType:
		// A property present in any member of a union is a declared prop. This
		// is deliberately merged like an intersection: prop-types only decides
		// whether a read has a declaration, not which discriminated branch is
		// active at runtime.
		declared := concreteDeclaration(map[string]propType{})
		found := false
		for _, part := range n.AsUnionTypeNode().Types.Nodes {
			if members, ok := declaredType(part, aliases, aliasesBySymbol, seen, resolve); ok {
				mergeDeclaration(&declared, members)
				found = true
			}
		}
		return declared, found
	default:
		return opaqueDeclaration(), true
	}
	if members == nil {
		return concreteDeclaration(map[string]propType{}), true
	}
	declared := map[string]propType{}
	for _, member := range members.Nodes {
		if member == nil {
			continue
		}
		switch member.Kind {
		case ast.KindPropertySignature:
			property := member.AsPropertySignatureDeclaration()
			if property != nil {
				if name := keyName(property.Name()); name != "" {
					if property.Type != nil {
						declared[name] = propTypeFromType(property.Type.AsNode(), aliases, aliasesBySymbol, seen)
					} else {
						// TypeScript permits a property signature without an annotation.
						// It is still a declared prop, but its nested shape is unknown.
						declared[name] = propType{open: true}
					}
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
			return concreteDeclaration(declared), true
		}
		result := concreteDeclaration(declared)
		for _, clause := range heritageClauses.Nodes {
			if clause == nil || clause.Kind != ast.KindHeritageClause {
				continue
			}
			for _, heritage := range clause.AsHeritageClause().Types.Nodes {
				if heritage != nil && heritage.Kind == ast.KindTypeReference {
					entry := heritage.AsTypeReferenceNode()
					var base propDeclaration
					var ok bool
					if name := reactutil.EntityNameRightmost(entry.TypeName); name != nil && name.AsIdentifier().Text == "ReturnType" && entry.TypeArguments != nil && len(entry.TypeArguments.Nodes) == 1 {
						base, ok = returnTypeDeclaration(entry.TypeArguments.Nodes[0], aliases, aliasesBySymbol, seen, resolve)
					} else {
						base, ok = declaredType(entry.TypeName, aliases, aliasesBySymbol, seen, resolve)
					}
					if ok {
						mergeDeclaration(&result, base)
					} else {
						// Imported base interfaces such as React.HTMLAttributes are
						// intentionally opaque, so their additional props are valid.
						mergeDeclaration(&result, opaqueDeclaration())
					}
				}
			}
		}
		return result, true
	}
	return concreteDeclaration(declared), true
}

func propTypeFromType(n *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node, seen map[*ast.Node]bool) propType {
	// Upstream validates TypeScript props at the declared property boundary only.
	return propType{open: true}
}

func collectComponentTypeDeclaration(n *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node) {
	name := n.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}
	aliases[name.AsIdentifier().Text] = append(aliases[name.AsIdentifier().Text], n)
	if symbol := name.Symbol(); symbol != nil {
		aliasesBySymbol[symbol] = n
	}
	if symbol := n.Symbol(); symbol != nil {
		aliasesBySymbol[symbol] = n
	}
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
		if clause == nil || clause.Kind != ast.KindHeritageClause || clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
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

// extendsComponent accepts the pragma-independent `Foo.Component` form and
// JSDoc's explicit component marker, both of which eslint-plugin-react treats
// as React component classes.
func extendsComponent(class *ast.Node) bool {
	if class == nil {
		return false
	}
	for _, doc := range class.JSDoc(nil) {
		jd := doc.AsJSDoc()
		if jd == nil || jd.Tags == nil {
			continue
		}
		for _, tag := range jd.Tags.Nodes {
			if ast.IsJSDocAugmentsTag(tag) {
				aug := tag.AsJSDocAugmentsTag()
				if aug != nil && jsDocComponentName(aug.ClassName) {
					return true
				}
			}
		}
	}
	var clauses *ast.HeritageClauseList
	switch class.Kind {
	case ast.KindClassDeclaration:
		clauses = class.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		clauses = class.AsClassExpression().HeritageClauses
	}
	if clauses == nil {
		return false
	}
	for _, clause := range clauses.Nodes {
		if clause == nil || clause.Kind != ast.KindHeritageClause {
			continue
		}
		for _, typ := range clause.AsHeritageClause().Types.Nodes {
			if typ != nil && typ.Kind == ast.KindExpressionWithTypeArguments && jsDocComponentName(typ.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

func jsDocComponentName(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindExpressionWithTypeArguments {
		return jsDocComponentName(node.AsExpressionWithTypeArguments().Expression)
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		return keyName(node.AsPropertyAccessExpression().Name()) == "Component" || keyName(node.AsPropertyAccessExpression().Name()) == "PureComponent"
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text == "Component" || node.AsIdentifier().Text == "PureComponent"
	}
	return false
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

func importedReactGenericPropsType(typeNode *ast.Node) (string, *ast.Node) {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return "", nil
	}
	name := reactutil.EntityNameRightmost(typeNode.AsTypeReferenceNode().TypeName)
	if name == nil {
		return "", nil
	}
	arguments := typeNode.AsTypeReferenceNode().TypeArguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return "", nil
	}
	typeName := importedReactTypeName(name)
	// An unqualified component type is React-specific only when it is a
	// named import.  A project-local `FC<Props>` must not suppress this rule.
	if typeName == "" && typeNode.AsTypeReferenceNode().TypeName.Kind == ast.KindIdentifier {
		return "", nil
	}
	if typeName == "" {
		qualified := typeNode.AsTypeReferenceNode().TypeName
		if qualified == nil || qualified.Kind != ast.KindQualifiedName || qualified.AsQualifiedName().Left == nil || qualified.AsQualifiedName().Left.Kind != ast.KindIdentifier {
			return "", nil
		}
		qualifier := qualified.AsQualifiedName().Left.AsIdentifier().Text
		if !importedReactNamespace(name, qualifier) {
			return "", nil
		}
		typeName = name.AsIdentifier().Text
	}
	index := 0
	switch typeName {
	case "ForwardRefRenderFunction", "forwardRef":
		index = 1
	case "ComponentProps", "ComponentPropsWithRef", "ComponentPropsWithoutRef", "VFC", "VoidFunctionComponent", "PropsWithChildren", "SFC", "StatelessComponent", "FunctionComponent", "FC":
	default:
		return "", nil
	}
	if len(arguments.Nodes) <= index {
		return "", nil
	}
	return typeName, arguments.Nodes[index]
}

func reactComponentTypeArgument(typeNode *ast.Node) *ast.Node {
	typeName, propsType := importedReactGenericPropsType(typeNode)
	if !slices.Contains([]string{"FC", "FunctionComponent", "SFC", "StatelessComponent", "VFC", "VoidFunctionComponent", "ForwardRefRenderFunction"}, typeName) {
		return nil
	}
	return propsType
}

func importedReactNamespace(name *ast.Node, localName string) bool {
	if name == nil || localName == "" {
		return false
	}
	root := ast.GetSourceFileOfNode(name)
	if root == nil {
		return false
	}
	found := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil || found {
			return false
		}
		if n.Kind == ast.KindImportDeclaration {
			importDecl := n.AsImportDeclaration()
			module := importDecl.ModuleSpecifier
			if module == nil || module.Kind != ast.KindStringLiteral || module.Text() != "react" || importDecl.ImportClause == nil {
				n.ForEachChild(visit)
				return false
			}
			clause := importDecl.ImportClause.AsImportClause()
			defaultName := clause.Name()
			if defaultName != nil && defaultName.Kind == ast.KindIdentifier && defaultName.AsIdentifier().Text == localName {
				found = true
				return false
			}
			if clause.NamedBindings != nil && clause.NamedBindings.Kind == ast.KindNamespaceImport {
				ns := clause.NamedBindings.AsNamespaceImport()
				found = ns != nil && ns.Name() != nil && ns.Name().Kind == ast.KindIdentifier && ns.Name().AsIdentifier().Text == localName
			}
		}
		n.ForEachChild(visit)
		return false
	}
	root.Node.ForEachChild(visit)
	return found
}

// importedReactTypeName returns the exported name for a locally aliased React
// type (for example `import { FC as X } from "react"`).
func importedReactTypeName(name *ast.Node) string {
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	localName := name.AsIdentifier().Text
	root := ast.GetSourceFileOfNode(name)
	if root == nil {
		return ""
	}
	var imported string
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil || imported != "" {
			return false
		}
		if n.Kind == ast.KindImportSpecifier {
			specifier := n.AsImportSpecifier()
			if specifier.Name() != nil && specifier.Name().Kind == ast.KindIdentifier && specifier.Name().AsIdentifier().Text == localName {
				for parent := n.Parent; parent != nil; parent = parent.Parent {
					if parent.Kind == ast.KindImportDeclaration {
						module := parent.AsImportDeclaration().ModuleSpecifier
						if module != nil && module.Kind == ast.KindStringLiteral && module.Text() == "react" {
							imported = localName
							if specifier.PropertyName != nil && specifier.PropertyName.Kind == ast.KindIdentifier {
								imported = specifier.PropertyName.AsIdentifier().Text
							}
						}
						break
					}
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	root.Node.ForEachChild(visit)
	return imported
}

func forwardRefPropsType(n *ast.Node) (*ast.Node, bool, bool) {
	for current := n; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if parent.Kind == ast.KindCallExpression {
			call := parent.AsCallExpression()
			if call.Arguments == nil || len(call.Arguments.Nodes) == 0 || unwrap(call.Arguments.Nodes[0]) != unwrap(current) {
				return nil, false, false
			}
			callee := unwrap(call.Expression)
			isForwardRef := callee != nil && callee.Kind == ast.KindIdentifier && importedReactTypeName(callee) == "forwardRef"
			if receiver, name, ok := estreeMember(callee); ok {
				isForwardRef = name == "forwardRef" && receiver != nil && receiver.Kind == ast.KindIdentifier && receiver.AsIdentifier().Text == "React"
			}
			if !isForwardRef {
				return nil, false, false
			}
			arguments := call.TypeArguments
			if arguments == nil || len(arguments.Nodes) == 0 {
				return nil, true, false
			}
			if len(arguments.Nodes) >= 2 {
				return arguments.Nodes[1], true, true
			}
			// A single forwardRef type argument selects the ref type. Upstream
			// treats the missing props argument as an explicit empty declaration,
			// which takes precedence over an annotation on the callback parameter.
			return nil, true, true
		}
		if parent.Kind != ast.KindParenthesizedExpression && parent.Kind != ast.KindAsExpression && parent.Kind != ast.KindSatisfiesExpression && parent.Kind != ast.KindNonNullExpression && parent.Kind != ast.KindTypeAssertionExpression {
			return nil, false, false
		}
	}
	return nil, false, false
}

func componentName(n *ast.Node) string {
	var suffix []string
	if n != nil && n.Kind == ast.KindMethodDeclaration && n.Parent != nil && n.Parent.Kind == ast.KindObjectLiteralExpression {
		if name := keyName(n.Name()); name != "" {
			suffix = append(suffix, name)
		}
	}
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
		break
	}
	return reactutil.BindingIdentifierName(n)
}

func componentBinding(n *ast.Node, resolve func(*ast.Node) *ast.Symbol) *ast.Symbol {
	if n == nil {
		return nil
	}
	if n.Kind == ast.KindFunctionDeclaration || n.Kind == ast.KindClassDeclaration {
		return n.Symbol()
	}
	return componentRootBinding(n, resolve)
}

func componentRootBinding(n *ast.Node, resolve func(*ast.Node) *ast.Symbol) *ast.Symbol {
	for current := n; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression, ast.KindPropertyAssignment, ast.KindObjectLiteralExpression:
			current = parent
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				name := parent.AsVariableDeclaration().Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return utils.BindingNameSymbol(name)
				}
			}
		case ast.KindBinaryExpression:
			assignment := parent.AsBinaryExpression()
			if assignment.OperatorToken != nil && assignment.OperatorToken.Kind == ast.KindEqualsToken && assignment.Right == current {
				root, _, ok := memberNames(assignment.Left)
				if ok && root != nil && root.Kind == ast.KindIdentifier {
					if resolve != nil {
						if symbol := resolve(root); symbol != nil {
							return symbol
						}
					}
					return root.Symbol()
				}
			}
		}
		return nil
	}
	return nil
}

func componentScope(n *ast.Node) *ast.Node {
	for current := n; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock:
			return current
		}
	}
	return nil
}

func setDeclaredExisting(m map[string]propType, names []string, value propType) bool {
	if len(names) == 0 {
		return true
	}
	if len(names) == 1 {
		m[names[0]] = value
		return true
	}
	p, ok := m[names[0]]
	if !ok {
		return false
	}
	if p.children == nil {
		// A broad validator already permits every nested read. eslint-plugin-react
		// currently throws while applying this assignment; keeping the broad
		// declaration is the useful, non-crashing interpretation.
		return true
	}
	if !setDeclaredExisting(p.children, names[1:], value) {
		return false
	}
	m[names[0]] = p
	return true
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

func addDestructured(c *component, pattern *ast.Node, prefix []string, aliases map[*ast.Symbol]propAlias, recordIntermediate bool) {
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
		// A dynamic computed key has no statically knowable prop name. ESLint
		// does not turn its local binding into a validation requirement.
		if be.PropertyName != nil && be.PropertyName.Kind == ast.KindComputedPropertyName {
			break
		}
		key := keyName(be.PropertyName)
		if key == "" {
			key = keyName(be.Name())
		}
		if key == "" || be.Name() == nil {
			continue
		}
		path := append(append([]string{}, prefix...), key)
		// Every direct binding is a prop use. Function parameters and the special
		// lifecycle/setState destructuring paths also use intermediate objects.
		report := be.Name()
		if be.PropertyName != nil {
			report = be.PropertyName
		}
		if be.Name().Kind != ast.KindObjectBindingPattern || recordIntermediate {
			appendUse(c, report, path)
		}
		if be.Name().Kind == ast.KindIdentifier {
			if symbol := utils.BindingNameSymbol(be.Name()); symbol != nil {
				aliases[symbol] = propAlias{path: path, crossClassSafe: len(path) > 0}
			}
		} else if be.Name().Kind == ast.KindObjectBindingPattern {
			addDestructured(c, be.Name(), path, aliases, recordIntermediate)
		}
	}
}

func addThisPropsDestructured(c *component, pattern *ast.Node, aliases map[*ast.Symbol]propAlias) bool {
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
			addDestructured(c, be.Name(), nil, aliases, true)
		case ast.KindIdentifier:
			if symbol := utils.BindingNameSymbol(be.Name()); symbol != nil {
				aliases[symbol] = propAlias{}
			}
		}
		return true
	}
	return false
}

func componentUsesClassProps(c *component) bool {
	if c == nil || c.node == nil {
		return false
	}
	return c.node.Kind == ast.KindClassDeclaration || c.node.Kind == ast.KindClassExpression || c.node.Kind == ast.KindObjectLiteralExpression
}

func commonPropsName(name string) bool {
	return name == "props" || name == "nextProps" || name == "prevProps"
}

// estreeKeyName mirrors ESTree's `node.key.name`: identifier keys (including
// a computed identifier) have a name, while string/numeric literal keys do not.
func estreeKeyName(name *ast.Node) string {
	if name == nil {
		return ""
	}
	if name.Kind == ast.KindComputedPropertyName {
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	return reactutil.EsTreeName(name)
}

func estreeMember(n *ast.Node) (*ast.Node, string, bool) {
	n = unwrap(n)
	if n == nil {
		return nil, "", false
	}
	switch n.Kind {
	case ast.KindPropertyAccessExpression:
		member := n.AsPropertyAccessExpression()
		return unwrap(member.Expression), reactutil.EsTreeName(member.Name()), true
	case ast.KindElementAccessExpression:
		member := n.AsElementAccessExpression()
		return unwrap(member.Expression), reactutil.EsTreeName(ast.SkipParentheses(member.ArgumentExpression)), true
	}
	return nil, "", false
}

func lifecycleFunctionName(fn *ast.Node) string {
	if fn == nil {
		return ""
	}
	if fn.Kind == ast.KindConstructor {
		return "constructor"
	}
	if fn.Kind == ast.KindMethodDeclaration || fn.Kind == ast.KindGetAccessor || fn.Kind == ast.KindSetAccessor {
		return estreeKeyName(fn.Name())
	}
	if fn.Kind != ast.KindArrowFunction && fn.Kind != ast.KindFunctionExpression {
		return ""
	}
	parent := reactutil.SkipExpressionWrappersUp(fn)
	if parent == nil {
		return ""
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
		return estreeKeyName(parent.Name())
	}
	return ""
}

func isLifecycleFunction(fn *ast.Node, checkAsyncSafe bool) bool {
	if fn != nil && fn.Kind == ast.KindConstructor {
		return true
	}
	switch lifecycleFunctionName(fn) {
	case "componentWillReceiveProps", "shouldComponentUpdate", "componentWillUpdate", "componentDidUpdate":
		return true
	case "getDerivedStateFromProps", "getSnapshotBeforeUpdate", "UNSAFE_componentWillReceiveProps", "UNSAFE_componentWillUpdate":
		return checkAsyncSafe
	}
	return false
}

func isInsideLifecycle(ident *ast.Node, c *component, checkAsyncSafe bool) bool {
	for current := ident.Parent; current != nil && current != c.node; current = current.Parent {
		if ast.IsFunctionLike(current) && isLifecycleFunction(current, checkAsyncSafe) {
			return true
		}
	}
	return false
}

func estreeFirstMemberName(root *ast.Node) string {
	member := reactutil.SkipExpressionWrappersUp(root)
	receiver, name, ok := estreeMember(member)
	if ok && receiver == unwrap(root) {
		return name
	}
	return ""
}

func propsPath(root *ast.Node, names []string, c *component, aliases map[*ast.Symbol]propAlias, resolve func(*ast.Node) *ast.Symbol, checkAsyncSafe bool) ([]string, bool) {
	root = unwrap(root)
	if root == nil || c == nil {
		return nil, false
	}
	if crossesNonReactClass(root, c) {
		if root.Kind == ast.KindIdentifier {
			if len(aliases) > 0 && resolve != nil {
				if alias, ok := aliases[resolve(root)]; ok && alias.crossClassSafe {
					return append(append([]string{}, alias.path...), names...), true
				}
			}
			if root.AsIdentifier().Text == "props" && len(names) > 0 {
				return names, true
			}
		}
		return nil, false
	}
	if root.Kind == ast.KindThisKeyword {
		if !componentUsesClassProps(c) || len(names) == 0 || estreeFirstMemberName(root) != "props" {
			return nil, false
		}
		return names[1:], true
	}
	if root.Kind != ast.KindIdentifier {
		return nil, false
	}
	if len(aliases) > 0 && resolve != nil {
		if alias, ok := aliases[resolve(root)]; ok {
			return append(append([]string{}, alias.path...), names...), true
		}
	}
	name := root.AsIdentifier().Text
	if !componentUsesClassProps(c) {
		if name == "props" {
			return names, true
		}
		return nil, false
	}
	if commonPropsName(name) && isInsideLifecycle(root, c, checkAsyncSafe) {
		return names, true
	}
	if isSetStatePropsParameter(root, c) {
		return names, true
	}
	return nil, false
}

func isSetStatePropsParameter(ident *ast.Node, c *component) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier || c == nil {
		return false
	}
	for current := ident.Parent; current != nil && current != c.node; current = current.Parent {
		if !ast.IsFunctionLike(current) || !isSetStateCallback(current) {
			continue
		}
		params := reactutil.FunctionParameters(current)
		if len(params) >= 2 && params[1] != nil && params[1].Kind == ast.KindParameter && functionParameterName(params[1]) == ident.AsIdentifier().Text {
			return true
		}
	}
	return false
}

func isSetStateCallback(fn *ast.Node) bool {
	call := reactutil.SkipExpressionWrappersUp(fn)
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	invocation := call.AsCallExpression()
	if invocation.Arguments == nil || len(invocation.Arguments.Nodes) == 0 || unwrap(invocation.Arguments.Nodes[0]) != fn {
		return false
	}
	_, name, ok := estreeMember(invocation.Expression)
	return ok && name == "setState"
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

func getterReturn(body *ast.Node) (*ast.Node, bool) {
	if body == nil || body.Kind != ast.KindBlock {
		return nil, false
	}
	stmts := body.AsBlock().Statements
	if stmts == nil {
		return nil, false
	}
	for i := len(stmts.Nodes) - 1; i >= 0; i-- {
		if stmt := stmts.Nodes[i]; stmt != nil && stmt.Kind == ast.KindReturnStatement {
			return stmt.AsReturnStatement().Expression, true
		}
	}
	return nil, false
}

func componentFor(node *ast.Node, byNode map[*ast.Node]*component) *component {
	for current := node; current != nil; current = current.Parent {
		if c := byNode[current]; c != nil {
			return c
		}
	}
	return nil
}

func crossesNonReactClass(node *ast.Node, c *component) bool {
	if c == nil || c.node == nil || (c.node.Kind != ast.KindClassDeclaration && c.node.Kind != ast.KindClassExpression) {
		return false
	}
	enclosing := reactutil.EnclosingClass(node)
	return enclosing != nil && enclosing != c.node
}

func parenthesizedCreateClassComponent(node *ast.Node, byNode map[*ast.Node]*component) *component {
	for current := node; current != nil; current = current.Parent {
		if !ast.IsFunctionLike(current) {
			continue
		}
		parent := current.Parent
		parenthesized := false
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parenthesized = true
			parent = parent.Parent
		}
		if !parenthesized || parent == nil || parent.Kind != ast.KindPropertyAssignment {
			continue
		}
		if c := byNode[parent.Parent]; c != nil && c.node.Kind == ast.KindObjectLiteralExpression {
			return c
		}
	}
	return nil
}

func usageComponent(node *ast.Node, byNode map[*ast.Node]*component, pragma, createClass string) *component {
	// ES6 ownership stops at the first enclosing class. A non-React class
	// therefore blocks an outer React class, while a registered class (including
	// JSDoc and pragma-independent forms) owns the use immediately.
	blockedByClass := false
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind != ast.KindClassDeclaration && current.Kind != ast.KindClassExpression {
			continue
		}
		if c := byNode[current]; c != nil {
			return c
		}
		blockedByClass = true
		break
	}
	// The shared helper preserves createReactClass's scope walk, which may cross
	// a non-React class after the ES6 lookup above has stopped.
	if owner := reactutil.GetParentReactComponentScopeBased(node, pragma, createClass); owner != nil {
		if c := byNode[owner]; c != nil {
			return c
		}
	}
	if c := parenthesizedCreateClassComponent(node, byNode); c != nil {
		return c
	}
	// Functional components are the final arm of upstream's owner lookup.
	// Reusing the components already classified during collection avoids
	// repeating JSX-return analysis for every member access.
	for current := node.Parent; current != nil; current = current.Parent {
		if c := byNode[current]; c != nil && !componentUsesClassProps(c) {
			return c
		}
	}
	// Upstream's free component lookup is reached here only after the ES6
	// lookup stopped at a non-React class. propsPath narrows that fallback to a
	// direct textual `props.x` use.
	if blockedByClass {
		return componentFor(node.Parent, byNode)
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
	return memberNamesWithMode(n, false)
}

func propTypesAssignmentNames(n *ast.Node) (*ast.Node, []string, bool) {
	return memberNamesWithMode(n, true)
}

func finishMemberNames(names []string, estreePropertyNames bool) []string {
	slices.Reverse(names)
	if estreePropertyNames {
		return names
	}
	for i, name := range names {
		if !strings.HasPrefix(name, "__NUMERIC_PROP__:") {
			continue
		}
		if i == 0 {
			names[i] = strings.TrimPrefix(name, "__NUMERIC_PROP__:")
		} else {
			names[i] = "__COMPUTED_PROP__"
		}
	}
	return names
}

func memberNamesWithMode(n *ast.Node, estreePropertyNames bool) (*ast.Node, []string, bool) {
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
			names = append(names, k)
			if report == nil {
				report = pa.Name()
			}
			cur = unwrap(pa.Expression)
		case ast.KindElementAccessExpression:
			ea := cur.AsElementAccessExpression()
			argument := unwrap(ea.ArgumentExpression)
			k := ""
			if estreePropertyNames {
				argument = ast.SkipParentheses(ea.ArgumentExpression)
				if argument == nil || argument.Kind != ast.KindIdentifier {
					return nil, nil, false
				}
				k = argument.AsIdentifier().Text
			} else {
				k = elementName(ea.ArgumentExpression)
				if argument != nil && argument.Kind == ast.KindNumericLiteral {
					k = "__NUMERIC_PROP__:" + k
				}
				if k == "" {
					if argument != nil && argument.Kind == ast.KindBigIntLiteral {
						return nil, nil, false
					}
					k = "__COMPUTED_PROP__"
				}
			}
			names = append(names, k)
			if report == nil {
				report = ea.ArgumentExpression
			}
			cur = unwrap(ea.Expression)
		default:
			return cur, finishMemberNames(names, estreePropertyNames), true
		}
	}
	return report, finishMemberNames(names, estreePropertyNames), false
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

func isDeclared(m map[string]propType, names []string) bool {
	if len(names) == 0 {
		return true
	}
	p, ok := m[names[0]]
	if !ok {
		p, ok = m["__ANY_KEY__"]
	}
	if !ok {
		return names[0] == "__COMPUTED_PROP__"
	}
	return propTypeDeclares(p, names[1:])
}

func propTypeDeclares(p propType, remaining []string) bool {
	if p.invalid {
		return false
	}
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
		checkAsyncSafe := !reactutil.ReactVersionLessThan(ctx.Settings, 16, 3, 0)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		propWrappers := reactutil.GetPropWrapperFunctions(ctx.Settings)
		resolveSymbol := func(identifier *ast.Node) *ast.Symbol {
			if identifier == nil || identifier.Kind != ast.KindIdentifier {
				return nil
			}
			if ctx.Refs != nil {
				return ctx.Refs.Resolve(identifier)
			}
			return utils.GetReferenceSymbol(identifier, ctx.TypeChecker)
		}
		resolveInitializer := func(identifier *ast.Node) (*ast.Node, bool) {
			if symbol := resolveSymbol(identifier); symbol != nil {
				declaration := symbol.ValueDeclaration
				if declaration == nil && len(symbol.Declarations) > 0 {
					declaration = symbol.Declarations[0]
				}
				if declaration != nil && declaration.Kind == ast.KindVariableDeclaration {
					return declaration.AsVariableDeclaration().Initializer, true
				}
			}
			initializer := reactutil.ResolveIdentifierInitializer(identifier, ctx.TypeChecker)
			return initializer, initializer != nil
		}
		typeAliases := map[string][]*ast.Node{}
		typeAliasesBySymbol := map[*ast.Symbol]*ast.Node{}
		var comps []*component
		byNode := map[*ast.Node]*component{}
		byComponentKey := map[componentKey]*component{}
		aliases := map[*ast.Symbol]propAlias{}
		var declarationEvents []componentDeclarationEvent
		queueDeclaration := func(source *ast.Node, c *component, declaration propDeclaration, replaceOpacity, final bool) {
			if source != nil && c != nil {
				declarationEvents = append(declarationEvents, componentDeclarationEvent{
					source: source, component: c, declaration: declaration, replaceOpacity: replaceOpacity, final: final,
				})
			}
		}
		queueRuntimeDeclaration := func(source *ast.Node, c *component, expression *ast.Node) {
			declaration, ok := declared(expression, o.customValidators, propWrappers, resolveInitializer)
			if !ok {
				declaration = opaqueDeclaration()
			}
			queueDeclaration(source, c, declaration, false, false)
		}
		registerComponent := func(c *component, name string) {
			if name == "" {
				return
			}
			if c.binding != nil {
				byComponentKey[componentKey{name: name, binding: c.binding}] = c
			} else {
				byComponentKey[componentKey{name: name, scope: componentScope(c.node)}] = c
			}
		}
		lookupComponent := func(root *ast.Node, name string) *component {
			if root == nil || root.Kind != ast.KindIdentifier {
				return nil
			}
			if symbol := resolveSymbol(root); symbol != nil {
				return byComponentKey[componentKey{name: name, binding: symbol}]
			}
			return byComponentKey[componentKey{name: name, scope: componentScope(root)}]
		}
		var propTypeAssignments []*ast.Node
		queuePropTypeAssignment := func(assignment *ast.Node) {
			binary := assignment.AsBinaryExpression()
			root, names, ok := propTypesAssignmentNames(binary.Left)
			if !ok || root == nil || root.Kind != ast.KindIdentifier {
				return
			}
			propTypesIndex := slices.Index(names, "propTypes")
			if propTypesIndex < 0 {
				return
			}
			componentParts := append([]string{root.AsIdentifier().Text}, names[:propTypesIndex]...)
			componentName := strings.Join(componentParts, ".")
			c := lookupComponent(root, componentName)
			if c == nil {
				return
			}
			propNames := names[propTypesIndex+1:]
			if len(propNames) == 0 {
				queueRuntimeDeclaration(assignment, c, binary.Right)
			} else {
				declarationEvents = append(declarationEvents, componentDeclarationEvent{
					source: assignment, component: c, propNames: propNames,
					validator: validatorType(binary.Right, o.customValidators, resolveInitializer),
				})
			}
		}
		newComponent := func(node *ast.Node) *component {
			c := &component{node: node, declared: map[string]propType{}}
			comps = append(comps, c)
			byNode[node] = c
			c.binding = componentBinding(node, resolveSymbol)
			if name := componentName(node); name != "" {
				registerComponent(c, name)
			}
			if node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindClassExpression {
				if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
					if symbol := utils.BindingNameSymbol(name); symbol != nil {
						innerName := name.AsIdentifier().Text
						byComponentKey[componentKey{name: innerName, binding: symbol}] = c
						// Keep the historical text fallback for an unresolved reference to
						// an expression's inner name, but constrain it to the declaration
						// scope so unrelated blocks cannot claim the declaration.
						byComponentKey[componentKey{name: innerName, scope: componentScope(node)}] = c
					}
				}
			}
			return c
		}
		findUsageComponent := func(node *ast.Node) *component {
			if len(comps) == 0 {
				return nil
			}
			return usageComponent(node, byNode, pragma, createClass)
		}
		var walk ast.Visitor
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration:
				collectComponentTypeDeclaration(n, typeAliases, typeAliasesBySymbol)
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if reactutil.ExtendsReactComponent(n, pragma) || extendsComponent(n) {
					c := newComponent(n)
					propsType := classPropsType(n)
					if propsType != nil {
						declarationEvents = append(declarationEvents, componentDeclarationEvent{
							source: propsType, typeNode: propsType, component: c,
							final: n.Kind == ast.KindClassExpression,
						})
					}
				}
			case ast.KindObjectLiteralExpression:
				if reactutil.IsCreateReactClassObjectArg(n, pragma, createClass) {
					newComponent(n)
				}
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
				isAccessor := n.Kind == ast.KindGetAccessor || n.Kind == ast.KindSetAccessor
				if !isAccessor && !isGeneratorFunction(n) && reactutil.IsStatelessReactComponentWithWrappers(n, pragma, ctx.TypeChecker, wrappers) {
					c := newComponent(n)
					ps := n.Parameters()
					if len(ps) > 0 && ps[0].Kind == ast.KindParameter {
						parameter := ps[0].AsParameterDeclaration()
						name := parameter.Name()
						if name != nil && name.Kind == ast.KindObjectBindingPattern {
							addDestructured(findUsageComponent(name), name, nil, aliases, true)
						}
						propsType, isForwardRef, typedForwardRef := forwardRefPropsType(n)
						if typedForwardRef {
							if propsType == nil {
								queueDeclaration(n, c, concreteDeclaration(map[string]propType{}), true, false)
							} else {
								declarationEvents = append(declarationEvents, componentDeclarationEvent{
									source: n, typeNode: propsType, component: c,
									declaration: propDeclaration{opaque: true},
								})
							}
						} else if parameter.Type != nil {
							propsType := parameter.Type.AsNode()
							declarationEvents = append(declarationEvents, componentDeclarationEvent{
								source: propsType, typeNode: propsType, component: c,
							})
						} else if !isForwardRef {
							variableType := componentVariableType(n)
							if reactComponentTypeArgument(variableType) != nil {
								declarationEvents = append(declarationEvents, componentDeclarationEvent{
									source: variableType, typeNode: variableType, component: c,
								})
							}
						}
					}
				}
				if c := findUsageComponent(n); c != nil {
					params := reactutil.FunctionParameters(n)
					if len(params) > 0 && params[0] != nil && params[0].Kind == ast.KindParameter {
						if pattern := params[0].AsParameterDeclaration().Name(); pattern != nil && pattern.Kind == ast.KindObjectBindingPattern && isLifecycleFunction(n, checkAsyncSafe) {
							addDestructured(c, pattern, nil, aliases, true)
						}
					}
					// The second parameter of the first callback passed to any
					// `.setState` member is the current props object upstream.
					if len(params) >= 2 && params[1] != nil && params[1].Kind == ast.KindParameter && isSetStateCallback(n) {
						if pattern := params[1].AsParameterDeclaration().Name(); pattern != nil && pattern.Kind == ast.KindObjectBindingPattern {
							addDestructured(c, pattern, nil, aliases, true)
						}
					}
				}
				if n.Kind == ast.KindGetAccessor {
					ga := n.AsGetAccessorDeclaration()
					if ast.HasSyntacticModifier(n, ast.ModifierFlagsStatic) && keyName(ga.Name()) == "propTypes" {
						if c := componentFor(n.Parent, byNode); c != nil {
							if expression, found := getterReturn(ga.Body); found {
								queueRuntimeDeclaration(n, c, expression)
							}
						}
					}
				}
			case ast.KindPropertyAssignment:
				pa := n.AsPropertyAssignment()
				if keyName(pa.Name()) == "propTypes" {
					if c := componentFor(n.Parent, byNode); c != nil {
						queueRuntimeDeclaration(n, c, pa.Initializer)
					}
				}
			case ast.KindPropertyDeclaration:
				pd := n.AsPropertyDeclaration()
				if keyName(pd.Name()) == "props" && !ast.HasSyntacticModifier(n, ast.ModifierFlagsStatic) {
					if c := componentFor(n.Parent, byNode); c != nil && pd.Type != nil {
						propsType := pd.Type.AsNode()
						declarationEvents = append(declarationEvents, componentDeclarationEvent{
							source: propsType, typeNode: propsType, component: c,
						})
					}
				}
				if keyName(pd.Name()) == "propTypes" {
					if c := componentFor(n.Parent, byNode); c != nil {
						queueRuntimeDeclaration(n, c, pd.Initializer)
					}
				}
			case ast.KindVariableDeclaration:
				vd := n.AsVariableDeclaration()
				if vd.Name() != nil {
					if root, names, ok := memberNames(vd.Initializer); ok && root != nil {
						if c := findUsageComponent(n); c != nil {
							if root.Kind == ast.KindThisKeyword && len(names) == 0 && vd.Name().Kind == ast.KindObjectBindingPattern {
								addThisPropsDestructured(c, vd.Name(), aliases)
							}
							handledLifecycleDestructuring := false
							if crossesNonReactClass(n, c) && len(names) == 0 && vd.Name().Kind == ast.KindObjectBindingPattern && root.Kind == ast.KindIdentifier && commonPropsName(root.AsIdentifier().Text) && isInsideLifecycle(root, c, checkAsyncSafe) {
								addDestructured(c, vd.Name(), nil, aliases, true)
								handledLifecycleDestructuring = true
							}
							if path, ok := propsPath(root, names, c, aliases, resolveSymbol, checkAsyncSafe); ok && !handledLifecycleDestructuring {
								switch vd.Name().Kind {
								case ast.KindObjectBindingPattern:
									addDestructured(c, vd.Name(), path, aliases, true)
								case ast.KindIdentifier:
									if len(path) > 0 || componentUsesClassProps(c) {
										if symbol := utils.BindingNameSymbol(vd.Name()); symbol != nil {
											crossClassSafe := len(names) > 0
											aliases[symbol] = propAlias{path: path, crossClassSafe: crossClassSafe}
										}
									}
								}
							}
						}
					}
				}
			case ast.KindBinaryExpression:
				b := n.AsBinaryExpression()
				left := unwrap(b.Left)
				if b.OperatorToken != nil && b.OperatorToken.Kind == ast.KindEqualsToken && left != nil && (left.Kind == ast.KindPropertyAccessExpression || left.Kind == ast.KindElementAccessExpression) {
					root, names, ok := propTypesAssignmentNames(b.Left)
					if ok && root != nil && root.Kind == ast.KindIdentifier {
						propTypesIndex := slices.Index(names, "propTypes")
						if propTypesIndex < 0 {
							break
						}
						propTypeAssignments = append(propTypeAssignments, n)
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
				c := findUsageComponent(n)
				if path, ok := propsPath(root, names, c, aliases, resolveSymbol, checkAsyncSafe); ok {
					// The dynamic element itself is not a prop use; a surrounding
					// static member still retains the full path for declaration-time
					// computed-key handling.
					if n.Kind == ast.KindElementAccessExpression && len(path) > 0 && path[len(path)-1] == "__COMPUTED_PROP__" {
						break
					}
					appendUse(c, memberReportNode(n), path)
				}
			}
			n.ForEachChild(walk)
			return false
		}
		ctx.SourceFile.Node.ForEachChild(walk)
		// Resolve component types only after the existing walk has collected all
		// declarations. This preserves support for declarations after components
		// without a second full-file traversal.
		resolvedEvents := declarationEvents[:0]
		for _, event := range declarationEvents {
			if event.typeNode != nil {
				declaration, ok := declaredType(event.typeNode, typeAliases, typeAliasesBySymbol, map[*ast.Node]bool{}, resolveInitializer)
				if !ok {
					if !event.declaration.opaque {
						continue
					}
					declaration = opaqueDeclaration()
				}
				event.declaration = declaration
				event.typeNode = nil
				event.replaceOpacity = true
			}
			resolvedEvents = append(resolvedEvents, event)
		}
		declarationEvents = resolvedEvents
		// Components and their bindings are now complete, so every assignment can
		// be associated exactly once, including assignments before a declaration.
		for _, assignmentNode := range propTypeAssignments {
			queuePropTypeAssignment(assignmentNode)
		}
		slices.SortStableFunc(declarationEvents, func(a, b componentDeclarationEvent) int {
			if a.final != b.final {
				if a.final {
					return 1
				}
				return -1
			}
			return a.source.Pos() - b.source.Pos()
		})
		for _, event := range declarationEvents {
			if len(event.propNames) > 0 {
				if !setDeclaredExisting(event.component.declared, event.propNames, event.validator) {
					// A nested propTypes assignment would itself fail at runtime when
					// its parent validator has not been declared. Upstream abandons
					// validation for the component rather than inventing that path.
					event.component.ignoreValidation = true
				}
			} else {
				applyDeclared(event.component, event.declaration, event.replaceOpacity)
			}
			event.component.declaredBlock = true
		}
		isDeclaredInComponentChain := func(c *component, names []string) bool {
			for current := c.node; current != nil; current = current.Parent {
				if owner := byNode[current]; owner != nil && isDeclared(owner.declared, names) {
					return true
				}
			}
			return false
		}
		for _, c := range comps {
			if len(c.used) == 0 || c.ignoreValidation || !c.declaredBlock && o.skipUndeclared {
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
				if !isDeclaredInComponentChain(c, u.names) {
					displayName := reportName(u.names)
					ctx.ReportNode(u.node, rule.RuleMessage{Id: "missingPropType", Description: fmt.Sprintf("'%s' is missing in props validation", displayName), Data: map[string]string{"name": displayName}})
				}
			}
		}
		return rule.RuleListeners{}
	},
}
