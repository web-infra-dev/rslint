package modules

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// AttributeState describes how precisely a module reference's import
// attributes can be understood from syntax alone.
type AttributeState uint8

const (
	AttributesNone AttributeState = iota
	AttributesStatic
	AttributesDynamic
	AttributesInvalid
)

// ImportAttribute is one statically known import attribute.
type ImportAttribute struct {
	Name  string
	Value string
}

// ImportAttributes keeps the authored attribute node together with a
// canonical, order-independent semantic representation. The value is immutable
// after collection and can therefore be shared with the SourceFile cache.
type ImportAttributes struct {
	Node    *ast.Node
	State   AttributeState
	entries []ImportAttribute
	key     string
}

// Key returns the stable semantic identity of the attribute set. An absent
// set and an explicitly empty set have the same key, while their State and Node
// still preserve the authored distinction for diagnostics and fixes.
func (attributes ImportAttributes) Key() string {
	return attributes.key
}

// Entries returns a copy of the statically known attributes in canonical
// order. Dynamic and invalid sets have no semantic entries.
func (attributes ImportAttributes) Entries() []ImportAttribute {
	return slices.Clone(attributes.entries)
}

// Value returns one unambiguous static attribute value. Duplicate names are
// invalid for interpretation even when the parser recovered an AST for them.
func (attributes ImportAttributes) Value(name string) (string, bool) {
	if attributes.State != AttributesStatic {
		return "", false
	}
	found := false
	value := ""
	for _, attribute := range attributes.entries {
		if attribute.Name != name {
			continue
		}
		if found {
			return "", false
		}
		found = true
		value = attribute.Value
	}
	return value, found
}

func staticImportAttributes(node *ast.Node) ImportAttributes {
	if node == nil {
		return ImportAttributes{State: AttributesNone}
	}
	attributes := node.AsImportAttributes()
	if attributes.Token != ast.KindWithKeyword {
		// The parser recovers `assert` into this node, but it is not `with` and
		// does not carry the same semantics.
		return unknownImportAttributes(node, AttributesInvalid)
	}
	list := attributes.Attributes
	if list == nil || len(list.Nodes) == 0 {
		return ImportAttributes{Node: node, State: AttributesStatic}
	}
	entries := make([]ImportAttribute, 0, len(list.Nodes))
	for _, attributeNode := range list.Nodes {
		if attributeNode == nil || attributeNode.Kind != ast.KindImportAttribute {
			return unknownImportAttributes(node, AttributesInvalid)
		}
		attribute := attributeNode.AsImportAttribute()
		name := attribute.Name()
		if name == nil || attribute.Value == nil || !ast.IsStringLiteralLike(attribute.Value) {
			return unknownImportAttributes(node, AttributesInvalid)
		}
		entries = append(entries, ImportAttribute{Name: name.Text(), Value: attribute.Value.Text()})
	}
	if hasDuplicateAttributeNames(entries) {
		return unknownImportAttributes(node, AttributesInvalid)
	}
	return knownImportAttributes(node, entries)
}

func dynamicImportAttributes(call *ast.CallExpression) ImportAttributes {
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return ImportAttributes{State: AttributesNone}
	}
	options := unwrapTransparentExpression(call.Arguments.Nodes[1])
	if options == nil || options.Kind != ast.KindObjectLiteralExpression {
		return unknownImportAttributes(call.Arguments.Nodes[1], AttributesDynamic)
	}
	with, lookup := staticObjectProperty(options, "with")
	switch lookup {
	case objectPropertyAbsent:
		return ImportAttributes{State: AttributesNone}
	case objectPropertyUnknown:
		return unknownImportAttributes(options, AttributesDynamic)
	}
	with = unwrapTransparentExpression(with)
	if isUndefinedValue(with) {
		return ImportAttributes{State: AttributesNone}
	}
	if with == nil || with.Kind != ast.KindObjectLiteralExpression {
		return unknownImportAttributes(with, AttributesDynamic)
	}
	properties := with.AsObjectLiteralExpression().Properties
	if properties == nil || len(properties.Nodes) == 0 {
		return ImportAttributes{Node: with, State: AttributesStatic}
	}
	entries := make([]ImportAttribute, 0, len(properties.Nodes))
	for _, propertyNode := range properties.Nodes {
		if propertyNode == nil || propertyNode.Kind != ast.KindPropertyAssignment {
			return unknownImportAttributes(with, AttributesDynamic)
		}
		if isPrototypeSetter(propertyNode) {
			// Attributes come from own enumerable properties, so a prototype
			// setter never contributes an entry.
			continue
		}
		property := propertyNode.AsPropertyAssignment()
		name, ok := staticAttributeName(property.Name())
		value := unwrapTransparentExpression(property.Initializer)
		if !ok || value == nil || !ast.IsStringLiteralLike(value) {
			return unknownImportAttributes(with, AttributesDynamic)
		}
		entries = append(entries, ImportAttribute{Name: name, Value: value.Text()})
	}
	if hasDuplicateAttributeNames(entries) {
		return unknownImportAttributes(with, AttributesDynamic)
	}
	if len(entries) == 0 {
		return ImportAttributes{Node: with, State: AttributesStatic}
	}
	return knownImportAttributes(with, entries)
}

// unwrapTransparentExpression removes parentheses and TypeScript-only wrappers
// (`as`, `satisfies`, `<T>x`, `x!`) that have no runtime effect.
func unwrapTransparentExpression(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, ast.OEKAll)
}

// isUndefinedValue reports whether an expression evaluates to `undefined`.
func isUndefinedValue(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindVoidExpression || node.Kind == ast.KindIdentifier && node.Text() == "undefined")
}

// isNonObjectPrototypeValue reports whether `__proto__: node` leaves the
// prototype unchanged. The runtime ignores primitives and undefined, and null
// adds no inherited properties.
func isNonObjectPrototypeValue(node *ast.Node) bool {
	if isUndefinedValue(node) || ast.IsStringLiteralLike(node) {
		return true
	}
	switch node.Kind {
	case ast.KindNullKeyword, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	}
	return false
}

// isPrototypeSetter reports whether a property is a non-computed, non-shorthand
// `__proto__: value`, which sets the prototype instead of creating a property.
func isPrototypeSetter(propertyNode *ast.Node) bool {
	if propertyNode == nil || propertyNode.Kind != ast.KindPropertyAssignment {
		return false
	}
	name := propertyNode.AsPropertyAssignment().Name()
	if name == nil || (name.Kind != ast.KindIdentifier && !ast.IsStringLiteralLike(name)) {
		return false
	}
	return name.Text() == "__proto__"
}

func hasDuplicateAttributeNames(entries []ImportAttribute) bool {
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if _, exists := seen[entry.Name]; exists {
			return true
		}
		seen[entry.Name] = struct{}{}
	}
	return false
}

type objectPropertyLookup uint8

const (
	objectPropertyAbsent objectPropertyLookup = iota
	objectPropertyFound
	objectPropertyUnknown
)

func staticObjectProperty(object *ast.Node, wanted string) (*ast.Node, objectPropertyLookup) {
	properties := object.AsObjectLiteralExpression().Properties
	if properties == nil {
		return nil, objectPropertyAbsent
	}
	var value *ast.Node
	found := false
	inheritedUnknown := false
	for _, propertyNode := range properties.Nodes {
		if propertyNode == nil || propertyNode.Kind == ast.KindSpreadAssignment {
			return nil, objectPropertyUnknown
		}
		if isPrototypeSetter(propertyNode) {
			// Only an object prototype can supply an inherited `with`; a
			// non-object value leaves the prototype unchanged.
			initializer := unwrapTransparentExpression(propertyNode.AsPropertyAssignment().Initializer)
			if initializer == nil {
				return nil, objectPropertyUnknown
			}
			if !isNonObjectPrototypeValue(initializer) {
				inheritedUnknown = true
			}
			continue
		}
		var nameNode *ast.Node
		var propertyValue *ast.Node
		switch propertyNode.Kind {
		case ast.KindPropertyAssignment:
			property := propertyNode.AsPropertyAssignment()
			nameNode = property.Name()
			propertyValue = property.Initializer
		case ast.KindShorthandPropertyAssignment:
			nameNode = propertyNode.AsShorthandPropertyAssignment().Name()
			propertyValue = nameNode
		default:
			nameNode = propertyNode.Name()
			propertyValue = propertyNode
		}
		name, ok := staticAttributeName(nameNode)
		if !ok {
			return nil, objectPropertyUnknown
		}
		if name != wanted {
			continue
		}
		if found {
			return nil, objectPropertyUnknown
		}
		found = true
		value = propertyValue
	}
	if found {
		// An own property shadows anything inherited, whatever its position.
		return value, objectPropertyFound
	}
	if inheritedUnknown {
		return nil, objectPropertyUnknown
	}
	return nil, objectPropertyAbsent
}

func staticAttributeName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindComputedPropertyName {
		expression := ast.SkipParentheses(node.AsComputedPropertyName().Expression)
		if expression == nil || !ast.IsStringLiteralLike(expression) && expression.Kind != ast.KindNumericLiteral {
			return "", false
		}
		return expression.Text(), true
	}
	if node.Kind != ast.KindIdentifier && !ast.IsStringLiteralLike(node) && node.Kind != ast.KindNumericLiteral {
		return "", false
	}
	return node.Text(), true
}

func knownImportAttributes(node *ast.Node, entries []ImportAttribute) ImportAttributes {
	slices.SortFunc(entries, func(left, right ImportAttribute) int {
		if order := strings.Compare(left.Name, right.Name); order != 0 {
			return order
		}
		return strings.Compare(left.Value, right.Value)
	})
	return ImportAttributes{
		Node:    node,
		State:   AttributesStatic,
		entries: entries,
		key:     encodeImportAttributes(entries),
	}
}

func unknownImportAttributes(node *ast.Node, state AttributeState) ImportAttributes {
	key := "?" + strconv.Itoa(int(state))
	if node != nil {
		// Unknown forms are never safe to merge. Their source range keeps each
		// recovered or dynamic request distinct without retaining source text.
		key += ":" + strconv.Itoa(node.Pos()) + ":" + strconv.Itoa(node.End())
	}
	return ImportAttributes{Node: node, State: state, key: key}
}

func encodeImportAttributes(entries []ImportAttribute) string {
	var result strings.Builder
	for _, entry := range entries {
		result.WriteString(strconv.Itoa(len(entry.Name)))
		result.WriteByte(':')
		result.WriteString(entry.Name)
		result.WriteString(strconv.Itoa(len(entry.Value)))
		result.WriteByte(':')
		result.WriteString(entry.Value)
	}
	return result.String()
}
