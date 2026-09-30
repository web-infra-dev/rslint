package reactutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ComponentName returns the name attached to a component declaration or initializer.
func ComponentName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
		return componentPropertyName(node.Name())
	}
	if name := BindingIdentifierName(node); name != "" {
		return name
	}
	for current := node; current != nil; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			break
		}
		switch parent.Kind {
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				if name := parent.AsVariableDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
					return name.AsIdentifier().Text
				}
			}
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if bin.Right == current && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
				left := SkipExpressionWrappers(bin.Left)
				if left != nil && left.Kind == ast.KindIdentifier {
					return left.AsIdentifier().Text
				}
				if parts := PropertyAccessParts(left); len(parts) > 0 {
					return parts[len(parts)-1]
				}
			}
		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer == current {
				return componentPropertyName(parent.AsPropertyAssignment().Name())
			}
		}
	}
	return ""
}

// ClassPropsType returns the first type argument of the class extends clause.
func ClassPropsType(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	heritage := ast.GetClassExtendsHeritageElement(node)
	if heritage == nil {
		return nil
	}
	types := heritage.AsExpressionWithTypeArguments().TypeArguments
	if types == nil || len(types.Nodes) == 0 {
		return nil
	}
	return types.Nodes[0]
}

// ReactGenericArgument selects the props argument of a recognized React generic,
// resolving named, default, and namespace imports by binding identity.
func ReactGenericArgument(name *ast.Node, arguments *ast.NodeList, resolve func(*ast.Node) *ast.Symbol) *ast.Node {
	if name == nil || arguments == nil || resolve == nil {
		return nil
	}
	root, member := name, ""
	switch name.Kind {
	case ast.KindQualifiedName:
		root = name.AsQualifiedName().Left
		member = componentPropertyName(name.AsQualifiedName().Right)
	case ast.KindPropertyAccessExpression:
		root = name.AsPropertyAccessExpression().Expression
		member = componentPropertyName(name.AsPropertyAccessExpression().Name())
	}
	if root.Kind != ast.KindIdentifier {
		return nil
	}
	binding := resolve(root)
	if binding == nil {
		if member == "" || root.Text() != "React" {
			return nil
		}
		return reactGenericArgumentAtIndex(member, arguments)
	}
	imported := ""
	for _, statement := range ast.GetSourceFileOfNode(name).Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration.ModuleSpecifier == nil || declaration.ModuleSpecifier.Text() != "react" || declaration.ImportClause == nil {
			continue
		}
		clause := declaration.ImportClause.AsImportClause()
		if member != "" {
			if clause.Name() != nil && utils.BindingNameSymbol(clause.Name()) == binding {
				imported = member
			}
			if bindings := clause.NamedBindings; bindings != nil && bindings.Kind == ast.KindNamespaceImport && utils.BindingNameSymbol(bindings.Name()) == binding {
				imported = member
			}
		} else if bindings := clause.NamedBindings; bindings != nil && bindings.Kind == ast.KindNamedImports {
			for _, element := range bindings.AsNamedImports().Elements.Nodes {
				specifier := element.AsImportSpecifier()
				if utils.BindingNameSymbol(specifier.Name()) == binding {
					imported = root.Text()
					if specifier.PropertyName != nil {
						imported = specifier.PropertyName.Text()
					}
				}
			}
		}
	}
	return reactGenericArgumentAtIndex(imported, arguments)
}

func reactGenericArgumentAtIndex(imported string, arguments *ast.NodeList) *ast.Node {
	index := 0
	switch imported {
	case "forwardRef", "ForwardRefRenderFunction":
		index = 1
	case "ComponentProps", "ComponentPropsWithRef", "ComponentPropsWithoutRef", "VFC", "VoidFunctionComponent", "PropsWithChildren", "SFC", "StatelessComponent", "FunctionComponent", "FC":
	default:
		return nil
	}
	if len(arguments.Nodes) <= index {
		return nil
	}
	return arguments.Nodes[index]
}

// FunctionComponentType finds a function initializer's React type annotation
// or the props type argument of its enclosing forwardRef call.
func FunctionComponentType(node *ast.Node, resolve func(*ast.Node) *ast.Symbol) *ast.Node {
	for current := node; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			continue
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			callee := SkipExpressionWrappers(call.Expression)
			if callee != nil && call.TypeArguments != nil {
				if (callee.Kind == ast.KindIdentifier && callee.Text() == "forwardRef") ||
					(callee.Kind == ast.KindPropertyAccessExpression && componentPropertyName(callee.AsPropertyAccessExpression().Name()) == "forwardRef") {
					return ReactGenericArgument(callee, call.TypeArguments, resolve)
				}
			}
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer != current || parent.AsVariableDeclaration().Type == nil {
				return nil
			}
			typ := parent.AsVariableDeclaration().Type
			if typ.Kind == ast.KindTypeReference {
				return ReactGenericArgument(typ.AsTypeReferenceNode().TypeName, typ.AsTypeReferenceNode().TypeArguments, resolve)
			}
			return nil
		default:
			return nil
		}
	}
	return nil
}

// PropertyAccessParts splits a dotted identifier path into its names.
func PropertyAccessParts(node *ast.Node) []string {
	node = SkipExpressionWrappers(node)
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindIdentifier {
		return []string{node.AsIdentifier().Text}
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	pa := node.AsPropertyAccessExpression()
	parts := PropertyAccessParts(pa.Expression)
	if len(parts) == 0 {
		return nil
	}
	name := componentPropertyName(pa.Name())
	if name == "" {
		return nil
	}
	return append(parts, name)
}

// ComponentBinding returns the root binding owning a component initializer.
func ComponentBinding(node *ast.Node, resolve func(*ast.Node) *ast.Symbol) *ast.Symbol {
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindClassDeclaration {
		return node.Symbol()
	}
	for current := node; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
			current = parent
			continue
		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer == current && parent.Parent != nil {
				current = parent.Parent
				continue
			}
		case ast.KindObjectLiteralExpression:
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
				if root := PropertyAccessRoot(assignment.Left); root != nil && root.Kind == ast.KindIdentifier && resolve != nil {
					return resolve(root)
				}
			}
		}
		return nil
	}
	return nil
}

// ComponentTarget returns the property path owning a component initializer.
// A named function expression is addressed through its outer variable binding.
func ComponentTarget(node *ast.Node) []string {
	name := ComponentName(node)
	if name == "" {
		return nil
	}
	target := []string{name}
	memberTarget := node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor
	for current := node; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
			current = parent
			continue
		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer != current || parent.Parent == nil {
				return target
			}
			member := componentPropertyName(parent.AsPropertyAssignment().Name())
			if member == "" {
				return target
			}
			if !memberTarget {
				// A named function/class expression's own name is local to the
				// expression. External declarations address the containing object
				// property: `{ C: function Named() {} }` is `C`, not `C.Named`.
				target = []string{member}
				memberTarget = true
			} else if len(target) == 0 || target[0] != member {
				target = append([]string{member}, target...)
			}
			current = parent.Parent
			continue
		case ast.KindObjectLiteralExpression:
			current = parent
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				if objectName := parent.AsVariableDeclaration().Name(); objectName != nil && objectName.Kind == ast.KindIdentifier {
					if !memberTarget {
						return []string{objectName.Text()}
					}
					if len(target) == 1 && target[0] == objectName.AsIdentifier().Text {
						return target
					}
					return append([]string{objectName.AsIdentifier().Text}, target...)
				}
			}
		case ast.KindBinaryExpression:
			assignment := parent.AsBinaryExpression()
			if assignment.OperatorToken != nil && assignment.OperatorToken.Kind == ast.KindEqualsToken && assignment.Right == current {
				if parts := PropertyAccessParts(assignment.Left); len(parts) > 0 {
					if memberTarget {
						// The assignment owns an object container rather than the
						// component directly. Preserve the component's path inside it:
						// `ns.registry = { C: () => ... }` is `ns.registry.C`.
						return append(parts, target...)
					}
					return parts
				}
			}
		}
		return target
	}
	return target
}

// PropertyAccessRoot returns the base expression of a dotted property path.
func PropertyAccessRoot(node *ast.Node) *ast.Node {
	node = SkipExpressionWrappers(node)
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		return PropertyAccessRoot(node.AsPropertyAccessExpression().Expression)
	}
	return node
}

// DeclaredPropKey returns a prop declaration key and its diagnostic node.
// Computed identifier keys retain the syntactic name used by upstream.
func DeclaredPropKey(node *ast.Node) (string, *ast.Node) {
	if node == nil {
		return "", nil
	}
	name, ok := utils.GetStaticPropertyName(node)
	if node.Kind == ast.KindComputedPropertyName {
		expression := SkipExpressionWrappers(node.AsComputedPropertyName().Expression)
		if ok {
			return name, expression
		}
		if expression != nil && expression.Kind == ast.KindIdentifier {
			return expression.AsIdentifier().Text, expression
		}
		return "", nil
	}
	if !ok {
		return "", nil
	}
	return name, node
}

func componentPropertyName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	name, _ := utils.GetStaticPropertyName(node)
	return name
}
