package consistent_indexed_object_style

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type Options struct {
	Mode string `json:"mode"`
}

var ConsistentIndexedObjectStyleRule = rule.Rule{
	Name: "consistent-indexed-object-style",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := &Options{Mode: "record"}
		if options != nil {
			// Handle different option formats
			switch v := options.(type) {
			case string:
				opts.Mode = v
			case []interface{}:
				// If options is passed as an array, take the first element
				if len(v) > 0 {
					if modeStr, ok := v[0].(string); ok {
						opts.Mode = modeStr
					}
				}
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
			ast.KindMappedType: func(node *ast.Node) {
				checkMappedType(ctx, node)
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
			// startPos remains unchanged
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
	// ANY reference to self (direct or nested) should NOT be converted to Record
	// This matches TypeScript-ESLint behavior
	// Examples:
	// - interface Foo { [key: string]: Foo } - Don't convert
	// - interface Foo { [key: string]: Foo[] } - Don't convert
	// - interface Foo { [key: string]: { x: Foo } } - Don't convert
	// - interface Foo1 { [key: string]: Foo2 } interface Foo2 { [key: string]: Foo1 } - Don't convert
	if interfaceName != "" {
		// Check for any reference to self in the value type (deep check)
		if containsTypeReferenceWithVisited(valueType, interfaceName, make(map[string]bool)) {
			return // Contains self-reference - don't convert
		}
		// Check for circular reference chains - handles both interface-to-interface and mixed chains
		if isPartOfUnifiedCircularChain(ctx, interfaceName) {
			return // Part of circular chain - don't convert
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
	if member.Kind != ast.KindIndexSignature && member.Kind != ast.KindMappedType {
		return
	}

	var keyType, valueType *ast.Node
	var valueText string

	if member.Kind == ast.KindIndexSignature {
		indexSig := member.AsIndexSignatureDeclaration()
		if indexSig.Parameters == nil || len(indexSig.Parameters.Nodes) == 0 {
			return
		}

		param := indexSig.Parameters.Nodes[0]
		if param.Kind != ast.KindParameter {
			return
		}

		paramDecl := param.AsParameterDeclaration()
		keyType = paramDecl.Type
		if keyType == nil {
			return
		}

		valueType = indexSig.Type
		// Handle missing value type (e.g., [k in string]; without a type)
		// In such cases, we should convert to Record<KeyType, any>
		if valueType == nil {
			valueText = "any"
		} else {
			valueText = strings.TrimSpace(ctx.SourceFile.Text()[valueType.Pos():valueType.End()])
		}
	} else if member.Kind == ast.KindMappedType {
		mappedType := member.AsMappedTypeNode()
		if mappedType.TypeParameter == nil {
			return
		}

		// For mapped types like [K in string]: T, the constraint is the key type
		keyType = mappedType.TypeParameter.AsTypeParameter().Constraint
		if keyType == nil {
			return
		}

		valueType = mappedType.Type
		// Handle missing value type (e.g., [k in string]; without a type)
		// In such cases, we should convert to Record<KeyType, any>
		if valueType == nil {
			valueText = "any"
		} else {
			valueText = strings.TrimSpace(ctx.SourceFile.Text()[valueType.Pos():valueType.End()])
		}
	} else {
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
		// Only block if this type literal would create the circular reference at the top level

		if parentName != "" {
			// Check if this type literal creates a circular dependency when converted to Record
			// Following TypeScript-ESLint logic: check if converting THIS type literal would
			// result in a circular Record type that can't be expressed
			if wouldCreateCircularRecord(node, parentName) {
				return // Would create circular dependency - don't convert
			}
		}
	}

	keyText := strings.TrimSpace(ctx.SourceFile.Text()[keyType.Pos():keyType.End()])
	// Note: valueText was already calculated above to handle missing value types

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
			// startPos remains unchanged
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

func checkMappedType(ctx rule.RuleContext, node *ast.Node) {
	mappedType := node.AsMappedTypeNode()
	if mappedType.TypeParameter == nil {
		return
	}

	// For mapped types like [K in string]: T, the constraint is the key type
	keyType := mappedType.TypeParameter.AsTypeParameter().Constraint
	if keyType == nil {
		return
	}

	valueType := mappedType.Type
	var valueText string
	// Handle missing value type (e.g., [k in string]; without a type)
	// In such cases, we should convert to Record<KeyType, any>
	if valueType == nil {
		valueText = "any"
	} else {
		valueText = strings.TrimSpace(ctx.SourceFile.Text()[valueType.Pos():valueType.End()])
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

		if parentName != "" {
			// For any mapped type that references its parent type, don't convert
			// This includes direct references and references in unions
			// Only check if valueType is not nil (for cases where value type exists)
			if valueType != nil && containsTypeReferenceWithVisited(valueType, parentName, make(map[string]bool)) {
				return // Contains self-reference - don't convert
			}

			// Check if this type alias is part of a circular chain
			if isPartOfUnifiedCircularChain(ctx, parentName) {
				return // Part of a circular chain - don't convert
			}
		}
	}

	keyText := strings.TrimSpace(ctx.SourceFile.Text()[keyType.Pos():keyType.End()])

	var recordText string
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsReadonly) {
		recordText = fmt.Sprintf("Readonly<Record<%s, %s>>", keyText, valueText)
	} else {
		recordText = fmt.Sprintf("Record<%s, %s>", keyText, valueText)
	}

	// For mapped types, we need to replace the parent type (the whole type alias)
	// because the mapped type is the direct value of the type alias
	if parentDecl != nil && parentDecl.Kind == ast.KindTypeAliasDeclaration {
		typeAlias := parentDecl.AsTypeAliasDeclaration()
		if typeAlias.Name() != nil && ast.IsIdentifier(typeAlias.Name()) {
			typeName := typeAlias.Name().AsIdentifier().Text

			// Build the full replacement for the type alias
			var genericTypes string
			if typeAlias.TypeParameters != nil && len(typeAlias.TypeParameters.Nodes) > 0 {
				var paramTexts []string
				for _, param := range typeAlias.TypeParameters.Nodes {
					paramTexts = append(paramTexts, strings.TrimSpace(ctx.SourceFile.Text()[param.Pos():param.End()]))
				}
				genericTypes = "<" + strings.Join(paramTexts, ", ") + ">"
			}

			replacement := fmt.Sprintf("type %s%s = %s;", typeName, genericTypes, recordText)

			ctx.ReportNodeWithFixes(node, rule.RuleMessage{
				Id:          "preferRecord",
				Description: "A record is preferred over an index signature.",
			}, rule.RuleFix{
				Range: core.NewTextRange(parentDecl.Pos(), parentDecl.End()),
				Text:  replacement,
			})
			return
		}
	}

	// Fallback for standalone mapped types (though this shouldn't happen normally)
	startPos := node.Pos()
	fixText := recordText
	if startPos > 0 {
		prevChar := ctx.SourceFile.Text()[startPos-1]
		if prevChar == ' ' {
			// Don't include the space in the fix range
			// startPos remains unchanged
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
	return containsTypeReferenceWithVisited(typeNode, typeName, make(map[string]bool))
}

// Check if type contains any reference to the given type name with a visited set to prevent infinite recursion
func containsTypeReferenceWithVisited(typeNode *ast.Node, typeName string, visited map[string]bool) bool {
	if typeNode == nil || typeName == "" {
		return false
	}

	switch typeNode.Kind {
	case ast.KindTypeReference:
		typeRef := typeNode.AsTypeReferenceNode()
		if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
			referencedTypeName := typeRef.TypeName.AsIdentifier().Text
			if referencedTypeName == typeName {
				return true
			}
			// For type aliases, we need to check if the referenced type alias
			// contains a reference to the target type name. This handles cases like:
			// type ExampleRoot = ExampleUnion | ExampleObject;
			// interface ExampleObject { [key: string]: ExampleRoot; }
			if !visited[referencedTypeName] {
				visited[referencedTypeName] = true
				if containsTypeReferenceInTypeAliasWithVisited(typeNode, referencedTypeName, typeName, visited) {
					return true
				}
			}
		}
		// Check type arguments
		if typeRef.TypeArguments != nil {
			for _, arg := range typeRef.TypeArguments.Nodes {
				if containsTypeReferenceWithVisited(arg, typeName, visited) {
					return true
				}
			}
		}
	case ast.KindIndexedAccessType:
		// Handle Foo[number], Foo["key"], etc.
		indexedAccess := typeNode.AsIndexedAccessTypeNode()
		if indexedAccess.ObjectType != nil && containsTypeReferenceWithVisited(indexedAccess.ObjectType, typeName, visited) {
			return true
		}
		if indexedAccess.IndexType != nil && containsTypeReferenceWithVisited(indexedAccess.IndexType, typeName, visited) {
			return true
		}
	case ast.KindUnionType:
		unionType := typeNode.AsUnionTypeNode()
		if unionType.Types != nil {
			for _, t := range unionType.Types.Nodes {
				if containsTypeReferenceWithVisited(t, typeName, visited) {
					return true
				}
			}
		}
	case ast.KindIntersectionType:
		intersectionType := typeNode.AsIntersectionTypeNode()
		if intersectionType.Types != nil {
			for _, t := range intersectionType.Types.Nodes {
				if containsTypeReferenceWithVisited(t, typeName, visited) {
					return true
				}
			}
		}
	case ast.KindArrayType:
		arrayType := typeNode.AsArrayTypeNode()
		if arrayType.ElementType != nil {
			return containsTypeReferenceWithVisited(arrayType.ElementType, typeName, visited)
		}
	case ast.KindTupleType:
		tupleType := typeNode.AsTupleTypeNode()
		if tupleType.Elements != nil {
			for _, elem := range tupleType.Elements.Nodes {
				if containsTypeReferenceWithVisited(elem, typeName, visited) {
					return true
				}
			}
		}
	case ast.KindFunctionType, ast.KindConstructorType:
		// Check return type
		if typeNode.Type() != nil {
			return containsTypeReferenceWithVisited(typeNode.Type(), typeName, visited)
		}
	case ast.KindTypeLiteral:
		// Check inside type literal members
		typeLit := typeNode.AsTypeLiteralNode()
		if typeLit.Members != nil {
			for _, member := range typeLit.Members.Nodes {
				// For property signatures, check the type
				if member.Kind == ast.KindPropertySignature {
					propSig := member.AsPropertySignatureDeclaration()
					if propSig.Type != nil && containsTypeReferenceWithVisited(propSig.Type, typeName, visited) {
						return true
					}
				}
				// For index signatures, check the value type
				if member.Kind == ast.KindIndexSignature {
					indexSig := member.AsIndexSignatureDeclaration()
					if indexSig.Type != nil && containsTypeReferenceWithVisited(indexSig.Type, typeName, visited) {
						return true
					}
				}
				// For method signatures, check return type
				if member.Kind == ast.KindMethodSignature {
					methodSig := member.AsMethodSignatureDeclaration()
					if methodSig.Type != nil && containsTypeReferenceWithVisited(methodSig.Type, typeName, visited) {
						return true
					}
				}
			}
		}
	case ast.KindConditionalType:
		// For conditional types like "Foo extends T ? string : number"
		// We need to check all parts including the check type
		conditionalType := typeNode.AsConditionalTypeNode()
		// Check the check type (the "Foo" in "Foo extends T")
		if conditionalType.CheckType != nil && containsTypeReferenceWithVisited(conditionalType.CheckType, typeName, visited) {
			return true
		}
		// Check the extends type (the "T" in "Foo extends T")
		if conditionalType.ExtendsType != nil && containsTypeReferenceWithVisited(conditionalType.ExtendsType, typeName, visited) {
			return true
		}
		// Check the true and false branches
		if conditionalType.TrueType != nil && containsTypeReferenceWithVisited(conditionalType.TrueType, typeName, visited) {
			return true
		}
		if conditionalType.FalseType != nil && containsTypeReferenceWithVisited(conditionalType.FalseType, typeName, visited) {
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

// Check if a value type is a direct self-reference (not nested in other structures)
func isDirectSelfReference(valueType *ast.Node, typeName string) bool {
	if valueType == nil || typeName == "" {
		return false
	}

	// Check if it's a direct type reference to self
	if valueType.Kind == ast.KindTypeReference {
		typeRef := valueType.AsTypeReferenceNode()
		if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
			return typeRef.TypeName.AsIdentifier().Text == typeName
		}
	}

	return false
}

// Check if a type name (interface or type alias) is part of a circular reference chain
// This handles mixed chains where interfaces and type aliases reference each other
func isPartOfUnifiedCircularChain(ctx rule.RuleContext, typeName string) bool {
	// Build a map of all types (interfaces and type aliases) and what they reference
	typeRefs := make(map[string]string)

	var checkNode ast.Visitor
	checkNode = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindInterfaceDeclaration:
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
								typeRefs[name] = refType
							}
							// Also check if the type directly contains a reference to this interface
							// This handles cases like: interface Foo { [key: string]: SomeUnion | Foo }
							if containsTypeReference(indexSig.Type, name) {
								// If it contains a self-reference, don't try to convert it
								// We mark it as referencing itself to prevent conversion
								typeRefs[name] = name
							}
						}
					}
				}
			}
		case ast.KindTypeAliasDeclaration:
			typeAlias := node.AsTypeAliasDeclaration()
			if typeAlias.Name() != nil && ast.IsIdentifier(typeAlias.Name()) {
				name := typeAlias.Name().AsIdentifier().Text
				// Check if it's a type literal with single index signature
				if typeAlias.Type != nil && typeAlias.Type.Kind == ast.KindTypeLiteral {
					typeLit := typeAlias.Type.AsTypeLiteralNode()
					if typeLit.Members != nil && len(typeLit.Members.Nodes) == 1 {
						member := typeLit.Members.Nodes[0]
						if member.Kind == ast.KindIndexSignature {
							indexSig := member.AsIndexSignatureDeclaration()
							if indexSig.Type != nil {
								// Extract the referenced type - could be direct or inside Record
								refType := extractDirectTypeReference(indexSig.Type)
								if refType != "" {
									typeRefs[name] = refType
								}
								// Also check if the type directly contains a reference to this type alias
								// This handles cases like: type Foo = { [key: string]: SomeUnion | Foo }
								if containsTypeReference(indexSig.Type, name) {
									// If it contains a self-reference, don't try to convert it
									// We mark it as referencing itself to prevent conversion
									typeRefs[name] = name
								}
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

	// Now check if there's a circular chain starting from typeName
	visited := make(map[string]bool)
	current := typeName

	for {
		if visited[current] {
			// We've seen this before - there's a cycle
			return true
		}
		visited[current] = true

		// Check what this type references
		next, exists := typeRefs[current]
		if !exists || next == "" {
			// No reference or references something else
			return false
		}

		current = next
	}
}

// Check if a type alias contains a reference to a target type
// This function looks up the type alias declaration and checks its type definition
func containsTypeReferenceInTypeAlias(sourceNode *ast.Node, typeAliasName string, targetTypeName string) bool {
	return containsTypeReferenceInTypeAliasWithVisited(sourceNode, typeAliasName, targetTypeName, make(map[string]bool))
}

// Check if a type alias contains a reference to a target type with visited tracking
func containsTypeReferenceInTypeAliasWithVisited(sourceNode *ast.Node, typeAliasName string, targetTypeName string, visited map[string]bool) bool {
	// We need access to the entire source file to look up the type alias
	// Walk up the AST to find the source file
	var sourceFile *ast.Node
	current := sourceNode
	for current != nil {
		if current.Kind == ast.KindSourceFile {
			sourceFile = current
			break
		}
		current = current.Parent
	}

	if sourceFile == nil {
		return false
	}

	// Look for the type alias declaration
	var found bool
	var checkNode ast.Visitor
	checkNode = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypeAliasDeclaration {
			typeAlias := node.AsTypeAliasDeclaration()
			if typeAlias.Name() != nil && ast.IsIdentifier(typeAlias.Name()) {
				if typeAlias.Name().AsIdentifier().Text == typeAliasName {
					// Found the type alias, check if its type contains the target type
					if typeAlias.Type != nil {
						found = containsTypeReferenceWithVisited(typeAlias.Type, targetTypeName, visited)
					}
					return true // Stop traversal
				}
			}
		}
		// Continue traversal
		node.ForEachChild(checkNode)
		return false
	}

	sourceFile.ForEachChild(checkNode)
	return found
}

// wouldCreateCircularRecord checks if converting this type literal to Record would create
// a circular reference that can't be expressed in TypeScript Record types
func wouldCreateCircularRecord(typeLiteral *ast.Node, parentTypeName string) bool {
	if typeLiteral == nil || typeLiteral.Kind != ast.KindTypeLiteral {
		return false
	}

	// Get the parent type alias declaration
	parentDecl := findParentDeclaration(typeLiteral)
	if parentDecl == nil {
		return false
	}

	// Check if this type literal is directly part of the parent type alias
	// (not nested within other type literals)
	current := typeLiteral.Parent
	isDirectChild := false

	for current != nil {
		if current == parentDecl {
			isDirectChild = true
			break
		}
		// If we encounter another type literal, we're nested
		if current.Kind == ast.KindTypeLiteral && current != typeLiteral {
			break
		}
		// Allow transparent types (union, intersection, parentheses)
		if current.Kind != ast.KindUnionType &&
			current.Kind != ast.KindIntersectionType &&
			current.Kind != ast.KindParenthesizedType {
			// Some other type - if it's not the type alias, we're nested
			if current.Kind != ast.KindTypeAliasDeclaration {
				break
			}
		}
		current = current.Parent
	}

	// Only apply circular check for direct children of the type alias
	if !isDirectChild {
		return false
	}

	// Check if this type literal deeply references the parent type
	return isDeeplyReferencingType(typeLiteral, parentTypeName, make(map[*ast.Node]bool))
}
