package reactutil

import (
	"slices"

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
	return reactGenericArgumentAtIndex(reactImportedName(name, resolve, false), arguments)
}

// ReactGenericArgumentWithAmbient is ReactGenericArgument with support for the
// configured default React namespace supplied by an external ambient declaration.
func ReactGenericArgumentWithAmbient(name *ast.Node, arguments *ast.NodeList, resolve func(*ast.Node) *ast.Symbol) *ast.Node {
	if name == nil || arguments == nil || resolve == nil {
		return nil
	}
	return reactGenericArgumentAtIndex(reactImportedName(name, resolve, true), arguments)
}

// reactImportedName returns the exported React name referenced by an imported
// identifier or namespace/default member expression.
func reactImportedName(name *ast.Node, resolve func(*ast.Node) *ast.Symbol, allowAmbient bool) string {
	if name == nil || resolve == nil {
		return ""
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
		return ""
	}
	binding := resolve(root)
	sourceFile := ast.GetSourceFileOfNode(name)
	if member != "" && root.Text() == DefaultReactPragma &&
		(binding == nil || allowAmbient && sourceFile != nil && !utils.IsSymbolDeclaredInFile(binding, sourceFile)) {
		return member
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
			if clause.Name() != nil && (utils.BindingNameSymbol(clause.Name()) == binding || binding == nil && clause.Name().Text() == root.Text()) {
				imported = member
			}
			if bindings := clause.NamedBindings; bindings != nil && bindings.Kind == ast.KindNamespaceImport &&
				(utils.BindingNameSymbol(bindings.Name()) == binding || binding == nil && bindings.Name().Text() == root.Text()) {
				imported = member
			}
		} else if bindings := clause.NamedBindings; bindings != nil && bindings.Kind == ast.KindNamedImports {
			for _, element := range bindings.AsNamedImports().Elements.Nodes {
				specifier := element.AsImportSpecifier()
				if utils.BindingNameSymbol(specifier.Name()) == binding || binding == nil && specifier.Name().Text() == root.Text() {
					imported = root.Text()
					if specifier.PropertyName != nil {
						imported = specifier.PropertyName.Text()
					}
				}
			}
		}
	}
	return imported
}

func reactGenericArgumentAtIndex(imported string, arguments *ast.NodeList) *ast.Node {
	if arguments == nil {
		return nil
	}
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

// FunctionComponentTypeDetails reports the props type and whether a function
// is the first argument of React.forwardRef. A forwardRef call with type
// arguments but no props argument is distinguished from an untyped call.
func FunctionComponentTypeDetails(node *ast.Node, resolve func(*ast.Node) *ast.Symbol) (propsType *ast.Node, isForwardRef, hasTypeArguments bool) {
	return functionComponentTypeDetails(node, DefaultReactPragma, resolve, false)
}

// FunctionComponentTypeDetailsWithPragma opts into configured pragma support
// without changing the upstream-compatible contract used by existing rules.
func FunctionComponentTypeDetailsWithPragma(node *ast.Node, pragma string, resolve func(*ast.Node) *ast.Symbol) (propsType *ast.Node, isForwardRef, hasTypeArguments bool) {
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	return functionComponentTypeDetails(node, pragma, resolve, true)
}

func functionComponentTypeDetails(node *ast.Node, pragma string, resolve func(*ast.Node) *ast.Symbol, honorConfiguredPragma bool) (propsType *ast.Node, isForwardRef, hasTypeArguments bool) {
	for current := node; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			continue
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			if call.Arguments == nil || len(call.Arguments.Nodes) == 0 || SkipExpressionWrappers(call.Arguments.Nodes[0]) != SkipExpressionWrappers(current) {
				return nil, false, false
			}
			callee := SkipExpressionWrappers(call.Expression)
			matchesForwardRef := callee != nil && callee.Kind == ast.KindIdentifier &&
				reactImportedName(callee, resolve, false) == "forwardRef"
			if receiver, name, ok := PropMemberIdentifier(callee); ok &&
				name == "forwardRef" && receiver != nil && receiver.Kind == ast.KindIdentifier {
				receiverBinding := resolve(receiver)
				sourceFile := ast.GetSourceFileOfNode(receiver)
				unshadowedPragma := receiverBinding == nil ||
					sourceFile != nil && !utils.IsSymbolDeclaredInFile(receiverBinding, sourceFile)
				matchesForwardRef = receiver.Text() == DefaultReactPragma ||
					honorConfiguredPragma && (reactImportedName(callee, resolve, false) == "forwardRef" ||
						receiver.Text() == pragma && unshadowedPragma)
			}
			if !matchesForwardRef {
				return nil, false, false
			}
			if call.TypeArguments == nil || len(call.TypeArguments.Nodes) == 0 {
				return nil, true, false
			}
			if len(call.TypeArguments.Nodes) >= 2 {
				return call.TypeArguments.Nodes[1], true, true
			}
			return nil, true, true
		default:
			return nil, false, false
		}
	}
	return nil, false, false
}

// ComponentTypeAnnotation returns the authored variable annotation around a
// component initializer. It keeps the outer React generic intact.
func ComponentTypeAnnotation(node *ast.Node) *ast.Node {
	for current := node; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
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

// ReactFunctionComponentArgument returns the props argument of a recognized
// React function-component annotation.
func ReactFunctionComponentArgument(typeNode *ast.Node, resolve func(*ast.Node) *ast.Symbol) *ast.Node {
	return reactFunctionComponentArgument(typeNode, resolve, false)
}

// ReactFunctionComponentArgumentWithAmbient opts into ambient React namespace
// recognition for rules that intentionally improve on upstream behavior.
func ReactFunctionComponentArgumentWithAmbient(typeNode *ast.Node, resolve func(*ast.Node) *ast.Symbol) *ast.Node {
	return reactFunctionComponentArgument(typeNode, resolve, true)
}

func reactFunctionComponentArgument(typeNode *ast.Node, resolve func(*ast.Node) *ast.Symbol, allowAmbient bool) *ast.Node {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return nil
	}
	reference := typeNode.AsTypeReferenceNode()
	imported := reactImportedName(reference.TypeName, resolve, allowAmbient)
	if !slices.Contains([]string{"FC", "FunctionComponent", "SFC", "StatelessComponent", "VFC", "VoidFunctionComponent", "ForwardRefRenderFunction"}, imported) {
		return nil
	}
	return reactGenericArgumentAtIndex(imported, reference.TypeArguments)
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
			} else {
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
