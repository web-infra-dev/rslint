package consistent_indexed_object_style

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConsistentIndexedObjectStyleOptions struct {
	Style string `json:"style"`
}

// ConsistentIndexedObjectStyleRule enforces consistent usage of type imports
var ConsistentIndexedObjectStyleRule = rule.CreateRule(rule.Rule{
	Name: "consistent-indexed-object-style",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentIndexedObjectStyleOptions{
		Style: "record", // default
	}

	// Parse options
	if options != nil {
		// Handle array format: ["index-signature"]
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if style, ok := optArray[0].(string); ok {
				opts.Style = style
			}
		} else if optsMap, ok := options.(map[string]interface{}); ok {
			if style, exists := optsMap["style"].(string); exists {
				opts.Style = style
			}
		} else if style, ok := options.(string); ok {
			opts.Style = style
		}
	}

	return rule.RuleListeners{
		// Check interfaces with index signatures
		ast.KindInterfaceDeclaration: func(node *ast.Node) {
			if opts.Style != "record" {
				return
			}

			if node.Kind != ast.KindInterfaceDeclaration {
				return
			}

			interfaceDecl := node.AsInterfaceDeclaration()
			if interfaceDecl == nil {
				return
			}

			// Skip if interface extends other types
			if interfaceDecl.HeritageClauses != nil && len(interfaceDecl.HeritageClauses.Nodes) > 0 {
				return
			}

			// Check if interface has exactly one member and it's an index signature
			if interfaceDecl.Members == nil || len(interfaceDecl.Members.Nodes) != 1 {
				return
			}

			member := interfaceDecl.Members.Nodes[0]
			if member.Kind != ast.KindIndexSignature {
				return
			}

			indexSig := member.AsIndexSignatureDeclaration()
			if indexSig == nil {
				return
			}

			// Check for circular references
			if isDeeplyReferencingType(interfaceDecl.Name(), indexSig.Type) {
				return
			}

			// Report violation
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "preferRecord",
				Description: "A record is preferred over an index signature.",
			})
		},

		// Check type literals with index signatures
		ast.KindTypeLiteral: func(node *ast.Node) {
			if opts.Style != "record" {
				return
			}

			if node.Kind != ast.KindTypeLiteral {
				return
			}

			typeLiteral := node.AsTypeLiteralNode()
			if typeLiteral == nil {
				return
			}

			// Check if type literal has exactly one member and it's an index signature
			if typeLiteral.Members == nil || len(typeLiteral.Members.Nodes) != 1 {
				return
			}

			member := typeLiteral.Members.Nodes[0]
			if member.Kind != ast.KindIndexSignature {
				return
			}

			indexSig := member.AsIndexSignatureDeclaration()
			if indexSig == nil {
				return
			}

			// Check for circular references - need to find the type alias name
			if isCircularTypeReference(node, indexSig.Type) {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "preferRecord",
				Description: "A record is preferred over an index signature.",
			})
		},

		// Check mapped types
		ast.KindMappedType: func(node *ast.Node) {
			if opts.Style != "record" {
				return
			}

			if node.Kind != ast.KindMappedType {
				return
			}

			mappedType := node.AsMappedTypeNode()
			if mappedType == nil {
				return
			}

			// Check if mapped type can be converted to Record
			if !canConvertMappedTypeToRecord(mappedType) {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "preferRecord",
				Description: "A record is preferred over an index signature.",
			})
		},

		// Check Record types when in index-signature mode
		ast.KindTypeReference: func(node *ast.Node) {
			if opts.Style != "index-signature" {
				return
			}

			if node.Kind != ast.KindTypeReference {
				return
			}

			typeRef := node.AsTypeReference()
			if typeRef == nil {
				return
			}

			// Check if this is a Record type reference
			if !isRecordType(typeRef) {
				return
			}

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "preferIndexSignature",
				Description: "An index signature is preferred over a record.",
			})
		},
	}
}

// isRecordType checks if a type reference is a Record type
func isRecordType(typeRef *ast.TypeReferenceNode) bool {
	if typeRef.TypeName == nil {
		return false
	}

	if typeRef.TypeName.Kind != ast.KindIdentifier {
		return false
	}

	ident := typeRef.TypeName.AsIdentifier()
	if ident == nil {
		return false
	}

	// Check if it's "Record"
	if ident.Text != "Record" {
		return false
	}

	// Must have type arguments
	if typeRef.TypeArguments == nil || len(typeRef.TypeArguments.Nodes) < 2 {
		return false
	}

	// First type argument should be string, number, or symbol
	firstArg := typeRef.TypeArguments.Nodes[0]
	return isValidRecordKeyType(firstArg)
}

// isValidRecordKeyType checks if a type is a valid key type for Record
func isValidRecordKeyType(typeNode *ast.Node) bool {
	switch typeNode.Kind {
	case ast.KindStringKeyword, ast.KindNumberKeyword, ast.KindSymbolKeyword:
		return true
	case ast.KindLiteralType:
		return true
	case ast.KindUnionType:
		unionType := typeNode.AsUnionTypeNode()
		if unionType == nil || unionType.Types == nil {
			return false
		}
		// All union members should be valid key types
		for _, t := range unionType.Types.Nodes {
			if !isValidRecordKeyType(t) {
				return false
			}
		}
		return true
	}
	return false
}

// canConvertMappedTypeToRecord checks if a mapped type can be converted to Record
func canConvertMappedTypeToRecord(mappedType *ast.MappedTypeNode) bool {
	// Check if the type parameter constraint is string, number, or symbol
	if mappedType.TypeParameter == nil {
		return false
	}

	// TypeParameter is already a *Node, check if it's a valid type parameter
	if mappedType.TypeParameter.Kind != ast.KindTypeParameter {
		return false
	}

	typeParam := mappedType.TypeParameter.AsTypeParameter()
	if typeParam == nil {
		return false
	}

	// Check constraint
	if typeParam.Constraint == nil {
		return false
	}

	constraint := typeParam.Constraint
	switch constraint.Kind {
	case ast.KindStringKeyword, ast.KindNumberKeyword, ast.KindSymbolKeyword:
		// Valid key types
	default:
		return false
	}

	// Check if the mapped type references the type parameter in a way that prevents conversion
	if mappedType.Type != nil && isDeeplyReferencingTypeParam(typeParam.Name(), mappedType.Type) {
		return false
	}

	return true
}

// isDeeplyReferencingType checks if a type deeply references another type (circular reference)
func isDeeplyReferencingType(name *ast.Node, typeNode *ast.Node) bool {
	if name == nil || typeNode == nil {
		return false
	}

	nameIdent := name.AsIdentifier()
	if nameIdent == nil {
		return false
	}

	return checkTypeReference(nameIdent.Text, typeNode)
}

// isDeeplyReferencingTypeParam checks if a type references a type parameter
func isDeeplyReferencingTypeParam(name *ast.Node, typeNode *ast.Node) bool {
	if name == nil || typeNode == nil {
		return false
	}

	nameIdent := name.AsIdentifier()
	if nameIdent == nil {
		return false
	}

	return checkTypeReference(nameIdent.Text, typeNode)
}

// checkTypeReference recursively checks if a type references a given identifier
func checkTypeReference(targetName string, typeNode *ast.Node) bool {
	if typeNode == nil {
		return false
	}

	switch typeNode.Kind {
	case ast.KindTypeReference:
		typeRef := typeNode.AsTypeReference()
		if typeRef != nil && typeRef.TypeName != nil {
			if typeRef.TypeName.Kind == ast.KindIdentifier {
				ident := typeRef.TypeName.AsIdentifier()
				if ident != nil && ident.Text == targetName {
					return true
				}
			}
			// Check type arguments
			if typeRef.TypeArguments != nil {
				for _, arg := range typeRef.TypeArguments.Nodes {
					if checkTypeReference(targetName, arg) {
						return true
					}
				}
			}
		}

	case ast.KindArrayType:
		arrayType := typeNode.AsArrayTypeNode()
		if arrayType != nil && arrayType.ElementType != nil {
			return checkTypeReference(targetName, arrayType.ElementType)
		}

	case ast.KindUnionType:
		unionType := typeNode.AsUnionTypeNode()
		if unionType != nil && unionType.Types != nil {
			for _, t := range unionType.Types.Nodes {
				if checkTypeReference(targetName, t) {
					return true
				}
			}
		}

	case ast.KindIntersectionType:
		intersectionType := typeNode.AsIntersectionTypeNode()
		if intersectionType != nil && intersectionType.Types != nil {
			for _, t := range intersectionType.Types.Nodes {
				if checkTypeReference(targetName, t) {
					return true
				}
			}
		}

	case ast.KindParenthesizedType:
		parenType := typeNode.AsParenthesizedTypeNode()
		if parenType != nil && parenType.Type != nil {
			return checkTypeReference(targetName, parenType.Type)
		}

	case ast.KindTupleType:
		tupleType := typeNode.AsTupleTypeNode()
		if tupleType != nil && tupleType.Elements != nil {
			for _, elem := range tupleType.Elements.Nodes {
				if checkTypeReference(targetName, elem) {
					return true
				}
			}
		}

	case ast.KindTypeLiteral:
		typeLiteral := typeNode.AsTypeLiteralNode()
		if typeLiteral != nil && typeLiteral.Members != nil {
			for _, member := range typeLiteral.Members.Nodes {
				if checkMemberReference(targetName, member) {
					return true
				}
			}
		}

	case ast.KindFunctionType, ast.KindConstructorType:
		// For function/constructor types, check parameters and return type
		if typeNode.Kind == ast.KindFunctionType {
			funcType := typeNode.AsFunctionTypeNode()
			if funcType != nil {
				if funcType.Type != nil && checkTypeReference(targetName, funcType.Type) {
					return true
				}
				if funcType.Parameters != nil {
					for _, param := range funcType.Parameters.Nodes {
						paramDecl := param.AsParameterDeclaration()
						if paramDecl != nil && paramDecl.Type != nil {
							if checkTypeReference(targetName, paramDecl.Type) {
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}

// checkMemberReference checks if a type member references a given identifier
func checkMemberReference(targetName string, member *ast.Node) bool {
	if member == nil {
		return false
	}

	switch member.Kind {
	case ast.KindPropertySignature:
		propSig := member.AsPropertySignatureDeclaration()
		if propSig != nil && propSig.Type != nil {
			return checkTypeReference(targetName, propSig.Type)
		}

	case ast.KindMethodSignature:
		methodSig := member.AsMethodSignatureDeclaration()
		if methodSig != nil && methodSig.Type != nil {
			return checkTypeReference(targetName, methodSig.Type)
		}

	case ast.KindIndexSignature:
		indexSig := member.AsIndexSignatureDeclaration()
		if indexSig != nil && indexSig.Type != nil {
			return checkTypeReference(targetName, indexSig.Type)
		}
	}

	return false
}

// isCircularTypeReference checks if a type literal in a type alias has a circular reference
func isCircularTypeReference(typeLiteralNode *ast.Node, valueType *ast.Node) bool {
	if typeLiteralNode == nil || valueType == nil {
		return false
	}

	// Walk up the AST to find the type alias declaration
	parent := typeLiteralNode.Parent
	for parent != nil {
		if parent.Kind == ast.KindTypeAliasDeclaration {
			typeAlias := parent.AsTypeAliasDeclaration()
			if typeAlias != nil && typeAlias.Name() != nil {
				// Check if the value type references the type alias name
				return isDeeplyReferencingType(typeAlias.Name(), valueType)
			}
			break
		}
		parent = parent.Parent
	}

	return false
}
