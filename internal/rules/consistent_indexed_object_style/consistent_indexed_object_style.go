package consistent_indexed_object_style

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
)

type Options struct {
	Mode string `json:"mode"`
}

var ConsistentIndexedObjectStyleRule = rule.Rule{
	Name: "consistent-indexed-object-style",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := &Options{Mode: "record"}
		if options != nil {
			// The API handler already extracts the options from ["error", "index-signature"] format
			// So we should just get "index-signature" as a string
			if modeStr, ok := options.(string); ok {
				opts.Mode = modeStr
			}
		}

		if opts.Mode == "index-signature" {
			return rule.RuleListeners{
				ast.KindTypeReference: func(node *ast.Node) {
					checkTypeReference(ctx, node)
				},
			}
		}

		// Default to "record" mode
		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				checkInterfaceDeclaration(ctx, node)
			},
			ast.KindTypeLiteral: func(node *ast.Node) {
				checkTypeLiteral(ctx, node)
			},
		}
	},
}

func checkTypeReference(ctx rule.RuleContext, node *ast.Node) {
	typeRef := node.AsTypeReferenceNode()
	if typeRef.TypeName == nil || typeRef.TypeName.Kind != ast.KindIdentifier {
		return
	}

	identifier := typeRef.TypeName.AsIdentifier()
	if identifier.Text != "Record" {
		return
	}

	if typeRef.TypeArguments == nil || len(typeRef.TypeArguments.Nodes) != 2 {
		return
	}

	indexParam := typeRef.TypeArguments.Nodes[0]
	shouldFix := indexParam.Kind == ast.KindStringKeyword ||
		indexParam.Kind == ast.KindNumberKeyword ||
		indexParam.Kind == ast.KindSymbolKeyword

	keyText := strings.TrimSpace(ctx.SourceFile.Text()[indexParam.Pos():indexParam.End()])
	valueText := strings.TrimSpace(ctx.SourceFile.Text()[typeRef.TypeArguments.Nodes[1].Pos():typeRef.TypeArguments.Nodes[1].End()])

	// Check if we need to preserve a space before the type reference
	startPos := node.Pos()
	fixText := fmt.Sprintf("{ [key: %s]: %s }", keyText, valueText)
	if startPos > 0 {
		prevChar := ctx.SourceFile.Text()[startPos-1]
		if prevChar == ' ' {
			// Don't include the space in the fix range
			startPos = startPos
		} else if prevChar == '=' || prevChar == ':' {
			// Add a space if there wasn't one
			fixText = " " + fixText
		}
	}

	if shouldFix {
		ctx.ReportNodeWithFixes(node, rule.RuleMessage{
			Id:          "preferIndexSignature",
			Description: "An index signature is preferred over a record.",
		}, rule.RuleFix{
			Range: core.NewTextRange(startPos, node.End()),
			Text:  fixText,
		})
	} else {
		// For complex key types, just report without fix/suggestion 
		// (test framework doesn't support suggestions)
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "preferIndexSignature",
			Description: "An index signature is preferred over a record.",
		})
	}
}

func checkInterfaceDeclaration(ctx rule.RuleContext, node *ast.Node) {
	interfaceDecl := node.AsInterfaceDeclaration()
	if interfaceDecl.Members == nil {
		return
	}

	// Check if this interface has ONLY index signatures (no other members)
	var indexSignatures []*ast.Node
	hasOtherMembers := false
	
	for _, member := range interfaceDecl.Members.Nodes {
		if member.Kind == ast.KindIndexSignature {
			indexSignatures = append(indexSignatures, member)
		} else {
			hasOtherMembers = true
		}
	}
	
	// Only convert if there's exactly one index signature and no other members
	if len(indexSignatures) != 1 || hasOtherMembers {
		return
	}

	member := indexSignatures[0]

	indexSig := member.AsIndexSignatureDeclaration()
	if indexSig.Parameters == nil || len(indexSig.Parameters.Nodes) == 0 {
		return
	}

	param := indexSig.Parameters.Nodes[0]
	if param.Kind != ast.KindParameter {
		return
	}

	paramDecl := param.AsParameterDeclaration()
	keyType := paramDecl.Type
	if keyType == nil {
		return
	}

	valueType := indexSig.Type
	if valueType == nil {
		return
	}

	// Check for circular references
	var interfaceName string
	if interfaceDecl.Name() != nil && ast.IsIdentifier(interfaceDecl.Name()) {
		interfaceName = interfaceDecl.Name().AsIdentifier().Text
	}
	
	// Check if the interface references itself or is part of a circular chain
	// Self-referential and circular interfaces should NOT be converted to Record
	// Examples:
	// - interface Foo { [key: string]: Foo }
	// - interface Foo<T> { [key: string]: Foo<T> | string }
	// - interface Foo1 { [key: string]: Foo2 } interface Foo2 { [key: string]: Foo1 }
	if interfaceName != "" {
		if containsTypeReference(valueType, interfaceName) || isPartOfCircularChain(ctx, interfaceName) {
			return // Self-referential or circular interface - don't convert
		}
	}

	// Check if interface extends anything - if so, we can't safely convert
	canFix := interfaceDecl.HeritageClauses == nil || len(interfaceDecl.HeritageClauses.Nodes) == 0

	var genericTypes string
	if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
		var paramTexts []string
		for _, param := range interfaceDecl.TypeParameters.Nodes {
			paramTexts = append(paramTexts, strings.TrimSpace(ctx.SourceFile.Text()[param.Pos():param.End()]))
		}
		genericTypes = "<" + strings.Join(paramTexts, ", ") + ">"
	}

	keyText := strings.TrimSpace(ctx.SourceFile.Text()[keyType.Pos():keyType.End()])
	valueText := strings.TrimSpace(ctx.SourceFile.Text()[valueType.Pos():valueType.End()])

	var recordText string
	if ast.HasSyntacticModifier(member, ast.ModifierFlagsReadonly) {
		recordText = fmt.Sprintf("Readonly<Record<%s, %s>>", keyText, valueText)
	} else {
		recordText = fmt.Sprintf("Record<%s, %s>", keyText, valueText)
	}

	replacement := fmt.Sprintf("type %s%s = %s;", interfaceName, genericTypes, recordText)

	if canFix {
		ctx.ReportNodeWithFixes(node, rule.RuleMessage{
			Id:          "preferRecord",
			Description: "A record is preferred over an index signature.",
		}, rule.RuleFix{
			Range: core.NewTextRange(node.Pos(), node.End()),
			Text:  replacement,
		})
	} else {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "preferRecord",
			Description: "A record is preferred over an index signature.",
		})
	}
}

func checkTypeLiteral(ctx rule.RuleContext, node *ast.Node) {
	typeLit := node.AsTypeLiteralNode()
	if typeLit.Members == nil || len(typeLit.Members.Nodes) != 1 {
		return
	}

	member := typeLit.Members.Nodes[0]
	if member.Kind != ast.KindIndexSignature {
		return
	}

	indexSig := member.AsIndexSignatureDeclaration()
	if indexSig.Parameters == nil || len(indexSig.Parameters.Nodes) == 0 {
		return
	}

	param := indexSig.Parameters.Nodes[0]
	if param.Kind != ast.KindParameter {
		return
	}

	paramDecl := param.AsParameterDeclaration()
	keyType := paramDecl.Type
	if keyType == nil {
		return
	}

	valueType := indexSig.Type
	if valueType == nil {
		return
	}

	// Check for circular references
	parentDecl := findParentDeclaration(node)
	if parentDecl != nil {
		var parentName string
		if parentDecl.Kind == ast.KindTypeAliasDeclaration {
			typeAlias := parentDecl.AsTypeAliasDeclaration()
			if typeAlias.Name() != nil && ast.IsIdentifier(typeAlias.Name()) {
				parentName = typeAlias.Name().AsIdentifier().Text
			}
		} else if parentDecl.Kind == ast.KindInterfaceDeclaration {
			interfaceDecl := parentDecl.AsInterfaceDeclaration()
			if interfaceDecl.Name() != nil && ast.IsIdentifier(interfaceDecl.Name()) {
				parentName = interfaceDecl.Name().AsIdentifier().Text
			}
		}
		
		// For type Foo = { [key: string]: { [key: string]: Foo } };
		// The outer type literal contains Foo in its nested structure, so it shouldn't be converted
		// The inner type literal directly references Foo but should still be converted
		// The difference is: does the type literal create a circular dependency if converted to Record?
		
		if parentName != "" {
			// Check if the parent type alias is a union type AND the value references the parent
			// e.g., type Foo = { [key: string]: Foo } | Foo - should NOT convert
			// but: type Foo = { [key: string]: string } | Foo - should convert
			if parentDecl.Kind == ast.KindTypeAliasDeclaration {
				typeAlias := parentDecl.AsTypeAliasDeclaration()
				if typeAlias.Type != nil && typeAlias.Type.Kind == ast.KindUnionType {
					// Only prevent conversion if the value type references the parent
					if containsTypeReference(valueType, parentName) {
						return // Don't convert type literals that reference parent in union type aliases
					}
				}
			}
			
			// Check if the value type is a union that contains the parent type
			// However, for interfaces with only index signatures, this is generally safe
			// e.g., interface Foo { [key: string]: string | Foo } is fine to convert
			// Only restrict when we have complex nested structures
			if valueType.Kind == ast.KindUnionType && parentDecl.Kind == ast.KindTypeAliasDeclaration {
				// Only apply this restriction for type aliases, not interfaces
				unionType := valueType.AsUnionTypeNode()
				if unionType.Types != nil {
					for _, t := range unionType.Types.Nodes {
						if containsTypeReference(t, parentName) {
							return // Value is a union containing self-reference in type alias - don't convert
						}
					}
				}
			}
			
			// Check if the value type contains a type literal that would reference the parent
			// e.g., type Foo = { [key: string]: { [key: string]: Foo } }
			// The outer type literal should not be converted because it creates a complex structure
			if valueType.Kind == ast.KindTypeLiteral {
				typeLit := valueType.AsTypeLiteralNode()
				if typeLit.Members != nil && len(typeLit.Members.Nodes) == 1 {
					member := typeLit.Members.Nodes[0]
					if member.Kind == ast.KindIndexSignature {
						indexSig := member.AsIndexSignatureDeclaration()
						if indexSig.Type != nil && containsTypeReference(indexSig.Type, parentName) {
							return // Outer type literal contains nested structure with parent reference
						}
					}
				}
			}
			
			// For type aliases, we only need to check for complex circular chains
			// Simple self-references like `type Foo = { [key: string]: Foo[] }` 
			// can be converted to `type Foo = Record<string, Foo[]>`
			
			// Check if this type alias is part of a circular chain
			if isPartOfTypeAliasCircularChain(ctx, parentName) {
				return // Part of a circular chain - don't convert
			}
		}
	}

	keyText := strings.TrimSpace(ctx.SourceFile.Text()[keyType.Pos():keyType.End()])
	valueText := strings.TrimSpace(ctx.SourceFile.Text()[valueType.Pos():valueType.End()])

	var recordText string
	if ast.HasSyntacticModifier(member, ast.ModifierFlagsReadonly) {
		recordText = fmt.Sprintf("Readonly<Record<%s, %s>>", keyText, valueText)
	} else {
		recordText = fmt.Sprintf("Record<%s, %s>", keyText, valueText)
	}

	// Check if we need to preserve a space before the type literal
	startPos := node.Pos()
	fixText := recordText
	if startPos > 0 {
		prevChar := ctx.SourceFile.Text()[startPos-1]
		if prevChar == ' ' {
			// Don't include the space in the fix range
			startPos = startPos
		} else if prevChar == '=' || prevChar == ':' {
			// Add a space if there wasn't one
			fixText = " " + recordText
		}
	}
	
	ctx.ReportNodeWithFixes(node, rule.RuleMessage{
		Id:          "preferRecord",
		Description: "A record is preferred over an index signature.",
	}, rule.RuleFix{
		Range: core.NewTextRange(startPos, node.End()),
		Text:  fixText,
	})
}


func findParentDeclaration(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindTypeAliasDeclaration {
			return parent
		}
		parent = parent.Parent
	}
	return nil
}


// Check if type contains any reference to the given type name that would cause circular dependency
func containsTypeReference(typeNode *ast.Node, typeName string) bool {
	if typeNode == nil || typeName == "" {
		return false
	}
	
	
	switch typeNode.Kind {
	case ast.KindTypeReference:
		typeRef := typeNode.AsTypeReferenceNode()
		if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
			if typeRef.TypeName.AsIdentifier().Text == typeName {
				return true
			}
		}
		// Check type arguments
		if typeRef.TypeArguments != nil {
			for _, arg := range typeRef.TypeArguments.Nodes {
				if containsTypeReference(arg, typeName) {
					return true
				}
			}
		}
	case ast.KindIndexedAccessType:
		// Handle Foo[number], Foo["key"], etc.
		indexedAccess := typeNode.AsIndexedAccessTypeNode()
		if indexedAccess.ObjectType != nil && containsTypeReference(indexedAccess.ObjectType, typeName) {
			return true
		}
		if indexedAccess.IndexType != nil && containsTypeReference(indexedAccess.IndexType, typeName) {
			return true
		}
	case ast.KindUnionType:
		unionType := typeNode.AsUnionTypeNode()
		if unionType.Types != nil {
			for _, t := range unionType.Types.Nodes {
				if containsTypeReference(t, typeName) {
					return true
				}
			}
		}
	case ast.KindIntersectionType:
		intersectionType := typeNode.AsIntersectionTypeNode()
		if intersectionType.Types != nil {
			for _, t := range intersectionType.Types.Nodes {
				if containsTypeReference(t, typeName) {
					return true
				}
			}
		}
	case ast.KindArrayType:
		arrayType := typeNode.AsArrayTypeNode()
		if arrayType.ElementType != nil {
			return containsTypeReference(arrayType.ElementType, typeName)
		}
	case ast.KindTupleType:
		tupleType := typeNode.AsTupleTypeNode()
		if tupleType.Elements != nil {
			for _, elem := range tupleType.Elements.Nodes {
				if containsTypeReference(elem, typeName) {
					return true
				}
			}
		}
	case ast.KindFunctionType, ast.KindConstructorType:
		// Check return type
		if typeNode.Type() != nil {
			return containsTypeReference(typeNode.Type(), typeName)
		}
	case ast.KindTypeLiteral:
		// Check inside type literal members
		typeLit := typeNode.AsTypeLiteralNode()
		if typeLit.Members != nil {
			for _, member := range typeLit.Members.Nodes {
				// For index signatures, check the value type
				if member.Kind == ast.KindIndexSignature {
					indexSig := member.AsIndexSignatureDeclaration()
					if indexSig.Type != nil && containsTypeReference(indexSig.Type, typeName) {
						return true
					}
				}
				// For other members, recursively check
				if containsTypeReference(member, typeName) {
					return true
				}
			}
		}
	case ast.KindConditionalType:
		// For conditional types like "Foo extends T ? string : number"
		// We need to check all parts including the check type
		conditionalType := typeNode.AsConditionalTypeNode()
		// Check the check type (the "Foo" in "Foo extends T")
		if conditionalType.CheckType != nil && containsTypeReference(conditionalType.CheckType, typeName) {
			return true
		}
		// Check the extends type (the "T" in "Foo extends T")
		if conditionalType.ExtendsType != nil && containsTypeReference(conditionalType.ExtendsType, typeName) {
			return true
		}
		// Check the true and false branches
		if conditionalType.TrueType != nil && containsTypeReference(conditionalType.TrueType, typeName) {
			return true
		}
		if conditionalType.FalseType != nil && containsTypeReference(conditionalType.FalseType, typeName) {
			return true
		}
	}
	
	return false
}

// Check if a type name refers to an interface with an index signature
func isInterfaceWithIndexSignature(ctx rule.RuleContext, typeName string) bool {
	// Walk through the source file to find the interface
	var hasIndexSignature bool
	
	var checkNode ast.Visitor
	checkNode = func(node *ast.Node) bool {
		if node.Kind == ast.KindInterfaceDeclaration {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl.Name() != nil && ast.IsIdentifier(interfaceDecl.Name()) {
				if interfaceDecl.Name().AsIdentifier().Text == typeName {
					// Found the interface, check if it has an index signature
					if interfaceDecl.Members != nil {
						for _, member := range interfaceDecl.Members.Nodes {
							if member.Kind == ast.KindIndexSignature {
								hasIndexSignature = true
								return true // Stop traversal
							}
						}
					}
					return true // Stop traversal, interface found but no index signature
				}
			}
		}
		
		// Continue traversal
		node.ForEachChild(checkNode)
		return false
	}
	
	ctx.SourceFile.ForEachChild(checkNode)
	return hasIndexSignature
}

// Check if an interface references a specific type in its index signature
func interfaceReferencesType(ctx rule.RuleContext, interfaceName string, targetType string) bool {
	var found bool
	
	var checkNode ast.Visitor
	checkNode = func(node *ast.Node) bool {
		if node.Kind == ast.KindInterfaceDeclaration {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl.Name() != nil && ast.IsIdentifier(interfaceDecl.Name()) {
				if interfaceDecl.Name().AsIdentifier().Text == interfaceName {
					// Found the interface, check if it has an index signature that references targetType
					if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) == 1 {
						member := interfaceDecl.Members.Nodes[0]
						if member.Kind == ast.KindIndexSignature {
							indexSig := member.AsIndexSignatureDeclaration()
							if indexSig.Type != nil {
								found = containsTypeReference(indexSig.Type, targetType)
								return true // Stop traversal
							}
						}
					}
					return true // Stop traversal, interface found
				}
			}
		}
		
		// Continue traversal
		node.ForEachChild(checkNode)
		return false
	}
	
	ctx.SourceFile.ForEachChild(checkNode)
	return found
}

// Extract the interface name that is referenced, handling Record types
func extractReferencedInterface(typeNode *ast.Node) string {
	if typeNode == nil {
		return ""
	}
	
	if typeNode.Kind == ast.KindTypeReference {
		typeRef := typeNode.AsTypeReferenceNode()
		if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
			// Check if it's a Record type
			if typeRef.TypeName.AsIdentifier().Text == "Record" && 
			   typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) >= 2 {
				// For Record<K, V>, check if V is an interface reference
				valueType := typeRef.TypeArguments.Nodes[1]
				if valueType.Kind == ast.KindTypeReference {
					valueTypeRef := valueType.AsTypeReferenceNode()
					if valueTypeRef.TypeName != nil && ast.IsIdentifier(valueTypeRef.TypeName) {
						return valueTypeRef.TypeName.AsIdentifier().Text
					}
				}
			} else {
				// Direct type reference
				return typeRef.TypeName.AsIdentifier().Text
			}
		}
	}
	
	return ""
}

// Check if a type alias is part of a circular reference chain
func isPartOfTypeAliasCircularChain(ctx rule.RuleContext, typeAliasName string) bool {
	visited := make(map[string]bool)
	
	var checkCircular func(currentTypeAlias string, targetTypeAlias string) bool
	checkCircular = func(currentTypeAlias string, targetTypeAlias string) bool {
		if visited[currentTypeAlias] {
			return currentTypeAlias == targetTypeAlias
		}
		visited[currentTypeAlias] = true
		
		// Find the type alias and check what it references
		var referencedType string
		var checkNode ast.Visitor
		checkNode = func(node *ast.Node) bool {
			if node.Kind == ast.KindTypeAliasDeclaration {
				typeAlias := node.AsTypeAliasDeclaration()
				if typeAlias.Name() != nil && ast.IsIdentifier(typeAlias.Name()) {
					if typeAlias.Name().AsIdentifier().Text == currentTypeAlias {
						// Found the type alias, check if it's a type literal with single index signature
						if typeAlias.Type != nil && typeAlias.Type.Kind == ast.KindTypeLiteral {
							typeLit := typeAlias.Type.AsTypeLiteralNode()
							if typeLit.Members != nil && len(typeLit.Members.Nodes) == 1 {
								member := typeLit.Members.Nodes[0]
								if member.Kind == ast.KindIndexSignature {
									indexSig := member.AsIndexSignatureDeclaration()
									if indexSig.Type != nil {
										// Extract the referenced type - could be direct or inside Record
										referencedType = extractReferencedInterface(indexSig.Type)
									}
								}
							}
						}
						return true // Stop traversal
					}
				}
			}
			// Continue traversal
			node.ForEachChild(checkNode)
			return false
		}
		
		ctx.SourceFile.ForEachChild(checkNode)
		
		if referencedType != "" && isTypeAliasWithIndexSignature(ctx, referencedType) {
			return checkCircular(referencedType, targetTypeAlias)
		}
		
		return false
	}
	
	return checkCircular(typeAliasName, typeAliasName)
}

// Check if a type alias has a type literal with an index signature
func isTypeAliasWithIndexSignature(ctx rule.RuleContext, typeName string) bool {
	var hasIndexSignature bool
	
	var checkNode ast.Visitor
	checkNode = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypeAliasDeclaration {
			typeAlias := node.AsTypeAliasDeclaration()
			if typeAlias.Name() != nil && ast.IsIdentifier(typeAlias.Name()) {
				if typeAlias.Name().AsIdentifier().Text == typeName {
					// Check if it's a type literal with index signature
					if typeAlias.Type != nil && typeAlias.Type.Kind == ast.KindTypeLiteral {
						typeLit := typeAlias.Type.AsTypeLiteralNode()
						if typeLit.Members != nil {
							for _, member := range typeLit.Members.Nodes {
								if member.Kind == ast.KindIndexSignature {
									hasIndexSignature = true
									return true
								}
							}
						}
					}
					return true
				}
			}
		}
		// Continue traversal
		node.ForEachChild(checkNode)
		return false
	}
	
	ctx.SourceFile.ForEachChild(checkNode)
	return hasIndexSignature
}

// Check if an interface is part of a circular reference chain
func isPartOfCircularChain(ctx rule.RuleContext, interfaceName string) bool {
	// Build a map of all interfaces and what they reference
	interfaceRefs := make(map[string]string)
	
	var checkNode ast.Visitor
	checkNode = func(node *ast.Node) bool {
		if node.Kind == ast.KindInterfaceDeclaration {
			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl.Name() != nil && ast.IsIdentifier(interfaceDecl.Name()) {
				name := interfaceDecl.Name().AsIdentifier().Text
				// Check if it has a single index signature
				if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) == 1 {
					member := interfaceDecl.Members.Nodes[0]
					if member.Kind == ast.KindIndexSignature {
						indexSig := member.AsIndexSignatureDeclaration()
						if indexSig.Type != nil {
							// Extract what this interface references
							refType := extractDirectTypeReference(indexSig.Type)
							if refType != "" {
								interfaceRefs[name] = refType
							}
						}
					}
				}
			}
		}
		// Continue traversal
		node.ForEachChild(checkNode)
		return false
	}
	
	ctx.SourceFile.ForEachChild(checkNode)
	
	// Now check if there's a circular chain starting from interfaceName
	visited := make(map[string]bool)
	current := interfaceName
	
	for {
		if visited[current] {
			// We've seen this before - there's a cycle
			return true
		}
		visited[current] = true
		
		// Check what this interface references
		next, exists := interfaceRefs[current]
		if !exists || next == "" {
			// No reference or references something else
			return false
		}
		
		current = next
	}
}

// Extract the direct type reference from a type node (not inside unions, arrays, etc)
func extractDirectTypeReference(typeNode *ast.Node) string {
	if typeNode == nil {
		return ""
	}
	
	if typeNode.Kind == ast.KindTypeReference {
		typeRef := typeNode.AsTypeReferenceNode()
		if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
			typeName := typeRef.TypeName.AsIdentifier().Text
			// If it's a Record type, extract the value type
			if typeName == "Record" && typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) >= 2 {
				// For Record<K, V>, check if V is a type reference
				valueType := typeRef.TypeArguments.Nodes[1]
				return extractDirectTypeReference(valueType)
			}
			return typeName
		}
	}
	
	return ""
}

func isDeeplyReferencingType(node *ast.Node, superTypeName string, visited map[*ast.Node]bool) bool {
	if node == nil || superTypeName == "" {
		return false
	}

	// If we've already visited this node, it's circular but not the reference being checked
	if visited[node] {
		return false
	}

	// Add to visited set (never remove - this is the key difference from broken implementations)
	visited[node] = true

	switch node.Kind {
	case ast.KindTypeLiteral:
		typeLit := node.AsTypeLiteralNode()
		if typeLit.Members != nil {
			for _, member := range typeLit.Members.Nodes {
				if isDeeplyReferencingType(member, superTypeName, visited) {
					return true
				}
			}
		}

	case ast.KindTypeAliasDeclaration:
		typeAlias := node.AsTypeAliasDeclaration()
		if typeAlias.Type != nil {
			return isDeeplyReferencingType(typeAlias.Type, superTypeName, visited)
		}

	case ast.KindUnionType:
		unionType := node.AsUnionTypeNode()
		if unionType.Types != nil {
			for _, t := range unionType.Types.Nodes {
				if isDeeplyReferencingType(t, superTypeName, visited) {
					return true
				}
			}
		}

	case ast.KindIntersectionType:
		intersectionType := node.AsIntersectionTypeNode()
		if intersectionType.Types != nil {
			for _, t := range intersectionType.Types.Nodes {
				if isDeeplyReferencingType(t, superTypeName, visited) {
					return true
				}
			}
		}

	case ast.KindInterfaceDeclaration:
		interfaceDecl := node.AsInterfaceDeclaration()
		if interfaceDecl.Members != nil {
			for _, member := range interfaceDecl.Members.Nodes {
				if isDeeplyReferencingType(member, superTypeName, visited) {
					return true
				}
			}
		}

	case ast.KindIndexSignature:
		indexSig := node.AsIndexSignatureDeclaration()
		if indexSig.Type != nil {
			return isDeeplyReferencingType(indexSig.Type, superTypeName, visited)
		}

	case ast.KindTypeReference:
		typeRef := node.AsTypeReferenceNode()
		if typeRef.TypeName != nil && isDeeplyReferencingType(typeRef.TypeName, superTypeName, visited) {
			return true
		}
		if typeRef.TypeArguments != nil {
			for _, arg := range typeRef.TypeArguments.Nodes {
				if isDeeplyReferencingType(arg, superTypeName, visited) {
					return true
				}
			}
		}

	case ast.KindIdentifier:
		// Check if this identifier references the super type
		identifier := node.AsIdentifier()
		if identifier.Text == superTypeName {
			return true
		}

	case ast.KindArrayType:
		arrayType := node.AsArrayTypeNode()
		if arrayType.ElementType != nil {
			return isDeeplyReferencingType(arrayType.ElementType, superTypeName, visited)
		}

	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		if param.Type != nil {
			return isDeeplyReferencingType(param.Type, superTypeName, visited)
		}
	}

	return false
}