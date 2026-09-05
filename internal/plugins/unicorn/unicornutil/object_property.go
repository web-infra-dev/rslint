package unicornutil

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// ObjectDataProperty returns the name and value of an object-literal data
// property. ESTree represents both `key: value` and shorthand `key` entries as
// Property nodes with kind "init"; tsgo gives them distinct node kinds.
// Methods, accessors, and spreads are not data properties.
func ObjectDataProperty(property *ast.Node) (name *ast.Node, value *ast.Node, ok bool) {
	if property == nil {
		return nil, nil, false
	}

	switch property.Kind {
	case ast.KindPropertyAssignment:
		assignment := property.AsPropertyAssignment()
		name = assignment.Name()
		value = assignment.Initializer
	case ast.KindShorthandPropertyAssignment:
		shorthand := property.AsShorthandPropertyAssignment()
		name = shorthand.Name()
		value = name
	default:
		return nil, nil, false
	}

	if name == nil || value == nil {
		return nil, nil, false
	}
	return name, value, true
}
