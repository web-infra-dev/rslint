package consistent_indexed_object_style

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

type ConsistentIndexedObjectStyleOptions struct {
	Style string `json:"style"`
}

func buildPreferIndexSignatureMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferIndexSignature",
		Description: "An index signature is preferred over a record.",
	}
}

func buildPreferIndexSignatureSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferIndexSignatureSuggestion",
		Description: "Change into an index signature instead of a record.",
	}
}

func buildPreferRecordMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferRecord",
		Description: "A record is preferred over an index signature.",
	}
}

// Helper to get node text
func getNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	if node == nil {
		return ""
	}
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	return string(sourceFile.Text()[nodeRange.Pos():nodeRange.End()])
}

// Check if the type name is "Record"
func isRecordType(typeRef *ast.TypeReferenceNode) bool {
	if typeRef.TypeName == nil {
		return false
	}
	if !ast.IsIdentifier(typeRef.TypeName) {
		return false
	}
	identifier := typeRef.TypeName.AsIdentifier()
	return identifier.Text == "Record"
}

// Check if index param is simple (string, number, symbol keywords)
func isSimpleIndexParam(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringKeyword, ast.KindNumberKeyword, ast.KindSymbolKeyword:
		return true
	default:
		return false
	}
}

// Find parent type alias declaration for circular reference checking
func findParentDeclaration(node *ast.Node) *ast.TypeAliasDeclaration {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindTypeAliasDeclaration {
			return current.AsTypeAliasDeclaration()
		}
	}
	return nil
}

// Extract all referenced type names from various node types
func extractAllReferencedTypeNames(node *ast.Node) []string {
	names := []string{}
	extractReferencedTypeNamesHelper(node, &names, false)
	return names
}

// Helper to recursively extract type names
// skipConditionalTest indicates whether to skip the check type in conditional types
func extractReferencedTypeNamesHelper(node *ast.Node, names *[]string, skipConditionalTest bool) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindTypeReference:
		typeRef := node.AsTypeReferenceNode()
		if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
			typeName := typeRef.TypeName.AsIdentifier().Text
			// Check if it's a Record type with type arguments
			if typeName == "Record" && typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) >= 2 {
				// Extract types from Record's value type (second argument)
				extractReferencedTypeNamesHelper(typeRef.TypeArguments.Nodes[1], names, skipConditionalTest)
			} else if typeName != "Record" {
				*names = append(*names, typeName)
				// Also check type arguments for non-Record generic types
				if typeRef.TypeArguments != nil {
					for _, arg := range typeRef.TypeArguments.Nodes {
						extractReferencedTypeNamesHelper(arg, names, skipConditionalTest)
					}
				}
			}
		} else if typeRef.TypeName != nil && ast.IsQualifiedName(typeRef.TypeName) {
			// For qualified names like A.Foo, we don't treat them as references to Foo
			// since they are in a different namespace
			// We could extract the full qualified name if needed in the future
		}
	case ast.KindUnionType:
		// For union types, check each type
		unionType := node.AsUnionTypeNode()
		for _, t := range unionType.Types.Nodes {
			extractReferencedTypeNamesHelper(t, names, skipConditionalTest)
		}
	case ast.KindIntersectionType:
		// For intersection types, check each type
		intersectionType := node.AsIntersectionTypeNode()
		for _, t := range intersectionType.Types.Nodes {
			extractReferencedTypeNamesHelper(t, names, skipConditionalTest)
		}
	case ast.KindArrayType:
		// For array types like Foo[], check the element type
		arrayType := node.AsArrayTypeNode()
		extractReferencedTypeNamesHelper(arrayType.ElementType, names, skipConditionalTest)
	case ast.KindParenthesizedType:
		// For parenthesized types, check the inner type
		parenType := node.AsParenthesizedTypeNode()
		extractReferencedTypeNamesHelper(parenType.Type, names, skipConditionalTest)
	case ast.KindConditionalType:
		// For conditional types, only check the true and false types if skipConditionalTest is true
		// Otherwise check all parts (for general type analysis)
		conditionalType := node.AsConditionalTypeNode()
		if !skipConditionalTest {
			// Check the check type and extends type for general analysis
			if conditionalType.CheckType != nil {
				extractReferencedTypeNamesHelper(conditionalType.CheckType, names, skipConditionalTest)
			}
			if conditionalType.ExtendsType != nil {
				extractReferencedTypeNamesHelper(conditionalType.ExtendsType, names, skipConditionalTest)
			}
		}
		// Always check the result types
		if conditionalType.TrueType != nil {
			extractReferencedTypeNamesHelper(conditionalType.TrueType, names, skipConditionalTest)
		}
		if conditionalType.FalseType != nil {
			extractReferencedTypeNamesHelper(conditionalType.FalseType, names, skipConditionalTest)
		}
	case ast.KindTypeLiteral:
		// For type literals, check members
		typeLiteral := node.AsTypeLiteralNode()
		for _, member := range typeLiteral.Members.Nodes {
			if member.Kind == ast.KindIndexSignature {
				indexSig := member.AsIndexSignatureDeclaration()
				if indexSig.Type != nil {
					extractReferencedTypeNamesHelper(indexSig.Type, names, skipConditionalTest)
				}
			}
		}
	}
}

// Extract the type name from various node types (handles TypeReference, UnionType, etc.)
func extractReferencedTypeName(node *ast.Node) string {
	names := extractAllReferencedTypeNames(node)
	if len(names) > 0 {
		return names[0]
	}
	return ""
}

// Check if a type node contains conditional types
func containsConditionalType(node *ast.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindConditionalType:
		return true
	case ast.KindUnionType:
		// Check if any type in the union contains conditional types
		unionType := node.AsUnionTypeNode()
		for _, t := range unionType.Types.Nodes {
			if containsConditionalType(t) {
				return true
			}
		}
	case ast.KindIntersectionType:
		// Check if any type in the intersection contains conditional types
		intersectionType := node.AsIntersectionTypeNode()
		for _, t := range intersectionType.Types.Nodes {
			if containsConditionalType(t) {
				return true
			}
		}
	case ast.KindArrayType:
		// Check the element type
		arrayType := node.AsArrayTypeNode()
		return containsConditionalType(arrayType.ElementType)
	case ast.KindParenthesizedType:
		// Check the inner type
		parenType := node.AsParenthesizedTypeNode()
		return containsConditionalType(parenType.Type)
	case ast.KindTupleType:
		// Check tuple elements
		tupleType := node.AsTupleTypeNode()
		for _, elem := range tupleType.Elements.Nodes {
			if containsConditionalType(elem) {
				return true
			}
		}
	case ast.KindTypeLiteral:
		// Check members
		typeLiteral := node.AsTypeLiteralNode()
		for _, member := range typeLiteral.Members.Nodes {
			if member.Kind == ast.KindIndexSignature {
				indexSig := member.AsIndexSignatureDeclaration()
				if indexSig.Type != nil && containsConditionalType(indexSig.Type) {
					return true
				}
			}
		}
	case ast.KindFunctionType:
		// Check return type of function
		funcType := node.AsFunctionTypeNode()
		if funcType.Type != nil {
			return containsConditionalType(funcType.Type)
		}
	}

	return false
}

// Check if a type node directly references the given type name (not as part of a qualified name)
func isDirectReferenceToType(node *ast.Node, typeName string) bool {
	if node == nil || typeName == "" {
		return false
	}

	switch node.Kind {
	case ast.KindTypeReference:
		typeRef := node.AsTypeReferenceNode()
		if typeRef.TypeName != nil && ast.IsIdentifier(typeRef.TypeName) {
			// Direct identifier reference (e.g., Foo)
			return typeRef.TypeName.AsIdentifier().Text == typeName
		} else if typeRef.TypeName != nil && ast.IsQualifiedName(typeRef.TypeName) {
			// Qualified name (e.g., A.Foo) - not a direct reference to Foo
			return false
		}
	case ast.KindUnionType:
		// Check if any type in the union is a direct reference
		unionType := node.AsUnionTypeNode()
		for _, t := range unionType.Types.Nodes {
			if isDirectReferenceToType(t, typeName) {
				return true
			}
		}
	case ast.KindIntersectionType:
		// Check if any type in the intersection is a direct reference
		intersectionType := node.AsIntersectionTypeNode()
		for _, t := range intersectionType.Types.Nodes {
			if isDirectReferenceToType(t, typeName) {
				return true
			}
		}
	case ast.KindArrayType:
		// Check the element type
		arrayType := node.AsArrayTypeNode()
		return isDirectReferenceToType(arrayType.ElementType, typeName)
	case ast.KindParenthesizedType:
		// Check the inner type
		parenType := node.AsParenthesizedTypeNode()
		return isDirectReferenceToType(parenType.Type, typeName)
	case ast.KindTypeLiteral:
		// Check members
		typeLiteral := node.AsTypeLiteralNode()
		for _, member := range typeLiteral.Members.Nodes {
			if member.Kind == ast.KindIndexSignature {
				indexSig := member.AsIndexSignatureDeclaration()
				if indexSig.Type != nil && isDirectReferenceToType(indexSig.Type, typeName) {
					return true
				}
			}
		}
	case ast.KindFunctionType:
		// Check return type of function
		funcType := node.AsFunctionTypeNode()
		if funcType.Type != nil {
			return isDirectReferenceToType(funcType.Type, typeName)
		}
	case ast.KindTupleType:
		// Check tuple elements
		tupleType := node.AsTupleTypeNode()
		for _, elem := range tupleType.Elements.Nodes {
			if isDirectReferenceToType(elem, typeName) {
				return true
			}
		}
	case ast.KindConditionalType:
		// For conditional types like "Foo extends T ? string : number"
		// we should check the true and false types, but NOT the check type or extends type
		// because conditional types are not actual structural recursion
		conditionalType := node.AsConditionalTypeNode()
		if conditionalType.TrueType != nil && isDirectReferenceToType(conditionalType.TrueType, typeName) {
			return true
		}
		if conditionalType.FalseType != nil && isDirectReferenceToType(conditionalType.FalseType, typeName) {
			return true
		}
		return false
	}

	return false
}

// Build a map of all interfaces and type aliases that have a single index signature
func buildInterfaceGraph(sourceFile *ast.SourceFile) map[string][]string {
	interfaceGraph := make(map[string][]string)

	sourceFile.ForEachChild(func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindInterfaceDeclaration:
			interfaceDecl := node.AsInterfaceDeclaration()
			if name := interfaceDecl.Name(); name != nil && name.Kind == ast.KindIdentifier {
				declName := name.AsIdentifier().Text

				// Check if it has a single index signature
				if len(interfaceDecl.Members.Nodes) == 1 {
					member := interfaceDecl.Members.Nodes[0]
					if member.Kind == ast.KindIndexSignature {
						indexSig := member.AsIndexSignatureDeclaration()
						if indexSig.Type != nil {
							// Extract all referenced type names, skipping conditional test types to avoid false positives
							names := []string{}
							extractReferencedTypeNamesHelper(indexSig.Type, &names, true)
							if len(names) > 0 {
								interfaceGraph[declName] = names
							}
						}
					}
				}
			}
		case ast.KindTypeAliasDeclaration:
			typeAlias := node.AsTypeAliasDeclaration()
			if name := typeAlias.Name(); name != nil && name.Kind == ast.KindIdentifier {
				declName := name.AsIdentifier().Text
				
				// Check if the type is a type literal with a single index signature
				if typeAlias.Type != nil && typeAlias.Type.Kind == ast.KindTypeLiteral {
					typeLiteral := typeAlias.Type.AsTypeLiteralNode()
					if len(typeLiteral.Members.Nodes) == 1 {
						member := typeLiteral.Members.Nodes[0]
						if member.Kind == ast.KindIndexSignature {
							indexSig := member.AsIndexSignatureDeclaration()
							if indexSig.Type != nil {
								// Extract all referenced type names, skipping conditional test types to avoid false positives
								names := []string{}
								extractReferencedTypeNamesHelper(indexSig.Type, &names, true)
								if len(names) > 0 {
									interfaceGraph[declName] = names
								}
							}
						}
					}
				} else if typeAlias.Type != nil && typeAlias.Type.Kind == ast.KindMappedType {
					// Track mapped types as well
					mappedType := typeAlias.Type.AsMappedTypeNode()
					if mappedType.Type != nil {
						names := []string{}
						extractReferencedTypeNamesHelper(mappedType.Type, &names, true)
						if len(names) > 0 {
							interfaceGraph[declName] = names
						}
					}
				} else if typeAlias.Type != nil {
					// Also track type aliases that reference other types (for circular dependency detection)
					names := []string{}
					extractReferencedTypeNamesHelper(typeAlias.Type, &names, true)
					if len(names) > 0 {
						interfaceGraph[declName] = names
					}
				}
			}
		}
		return false // Continue traversing
	})

	return interfaceGraph
}

// Check if there's a circular dependency starting from the given interface
func hasCircularDependency(graph map[string][]string, startInterface string, visited map[string]bool, path map[string]bool) bool {
	if path[startInterface] {
		return true // Found a cycle
	}

	if visited[startInterface] {
		return false // Already checked this node in a different path
	}

	visited[startInterface] = true
	path[startInterface] = true

	if nextInterfaces, exists := graph[startInterface]; exists {
		for _, nextInterface := range nextInterfaces {
			if hasCircularDependency(graph, nextInterface, visited, path) {
				return true
			}
		}
	}

	delete(path, startInterface) // Remove from current path
	return false
}

// Check if interface or type alias has circular reference by checking if there are mutual references in the file
func hasCircularReference(sourceFile *ast.SourceFile, typeName string) bool {
	if typeName == "" {
		return false
	}

	// Build the interface/type alias dependency graph
	graph := buildInterfaceGraph(sourceFile)

	// Build complete graph before checking
	// fmt.Printf("Graph for %s: %v\n", typeName, graph)

	// Check if this type is part of any circular dependency
	visited := make(map[string]bool)
	path := make(map[string]bool)

	// Start from the type we're checking and see if we can reach back to it
	if hasCircularDependency(graph, typeName, visited, path) {
		// Found circular dependency
		return true
	}

	return false
}

var ConsistentIndexedObjectStyleRule = rule.Rule{
	Name: "consistent-indexed-object-style",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := ConsistentIndexedObjectStyleOptions{
			Style: "record", // default
		}

		if options != nil {
			// The options can be passed in different formats:
			// 1. Just the style: ["index-signature"] or "index-signature"
			// 2. With severity: ["error", "index-signature"]
			// We need to extract the actual option value
			
			switch v := options.(type) {
			case []interface{}:
				// Array format - find the style option
				for _, opt := range v {
					if styleStr, ok := opt.(string); ok {
						if styleStr == "record" || styleStr == "index-signature" {
							opts.Style = styleStr
							break
						}
					}
				}
			case string:
				// Direct string format
				if v == "record" || v == "index-signature" {
					opts.Style = v
				}
			}
		}

		mode := opts.Style

		// Helper function to check members for index signatures
		checkMembers := func(members []*ast.Node, node *ast.Node, parentId *ast.Identifier, prefix, postfix string, safeFix bool) {
			if len(members) != 1 {
				return
			}

			if parentId != nil {
				// Processing members for parent
			}

			member := members[0]
			if member.Kind != ast.KindIndexSignature {
				return
			}

			// Check for circular references early to avoid infinite loops

			indexSig := member.AsIndexSignatureDeclaration()
			if len(indexSig.Parameters.Nodes) == 0 {
				return
			}

			parameter := indexSig.Parameters.Nodes[0]
			if parameter.Kind != ast.KindParameter {
				return
			}

			param := parameter.AsParameterDeclaration()
			if param.Type == nil {
				return
			}

			keyType := param.Type
			if keyType == nil {
				return
			}

			valueType := indexSig.Type
			if valueType == nil {
				return
			}



			// Check for circular references
			if parentId != nil {
				parentName := parentId.Text
				


				// Check if the value type directly references the parent type (self-reference)
				// This handles cases like:
				// - interface Foo { [key: string]: Foo; } (self-reference)
				// - interface Foo { [key: string]: Foo | string; } (union with self-reference)
				// - interface Foo { [key: string]: Foo[]; } (array of self-reference)
				// But NOT cases like:
				// - interface Foo { [key: string]: A.Foo; } (qualified name, different type)
				// - interface Foo { [key: string]: Foo extends T ? string : number; } (conditional type)
				
				// Check if the value type contains conditional types - if so, skip conversion
				if containsConditionalType(valueType) {
					return
				}

				directRef := isDirectReferenceToType(valueType, parentName)
				if directRef {
					// Found direct self-reference, skip conversion
					return
				}

				// For interfaces and type aliases, also check for mutual circular references (A -> B -> A)
				// But skip this check for nested type literals to allow flagging inner index signatures
				if node.Kind == ast.KindInterfaceDeclaration {
					if hasCircularReference(ctx.SourceFile, parentName) {
						return
					}
				}

			}

			// Create the Record type replacement
			keyText := getNodeText(ctx.SourceFile, keyType)
			valueText := getNodeText(ctx.SourceFile, valueType)

			var record string
			// Check if the index signature has readonly modifier
			isReadonly := false
			if indexSig.Modifiers() != nil {
				for _, modifier := range indexSig.Modifiers().Nodes {
					if modifier.Kind == ast.KindReadonlyKeyword {
						isReadonly = true
						break
					}
				}
			}

			if isReadonly {
				record = fmt.Sprintf("Readonly<Record<%s, %s>>", keyText, valueText)
			} else {
				record = fmt.Sprintf("Record<%s, %s>", keyText, valueText)
			}

			replacement := prefix + record + postfix

			if safeFix {
				ctx.ReportNodeWithFixes(node, buildPreferRecordMessage(),
					rule.RuleFixReplace(ctx.SourceFile, node, replacement))
			} else {
				ctx.ReportNode(node, buildPreferRecordMessage())
			}
		}

		listeners := rule.RuleListeners{}

		// Handle index-signature mode
		if mode == "index-signature" {
			listeners[ast.KindTypeReference] = func(node *ast.Node) {
				typeRef := node.AsTypeReferenceNode()

				if !isRecordType(typeRef) {
					return
				}

				if typeRef.TypeArguments == nil || len(typeRef.TypeArguments.Nodes) != 2 {
					return
				}

				params := typeRef.TypeArguments.Nodes
				indexParam := params[0]

				shouldFix := isSimpleIndexParam(indexParam)

				if shouldFix {
					keyText := getNodeText(ctx.SourceFile, params[0])
					typeText := getNodeText(ctx.SourceFile, params[1])
					replacement := fmt.Sprintf("{ [key: %s]: %s }", keyText, typeText)

					ctx.ReportNodeWithFixes(node, buildPreferIndexSignatureMessage(),
						rule.RuleFixReplace(ctx.SourceFile, node, replacement))
				} else {
					keyText := getNodeText(ctx.SourceFile, params[0])
					typeText := getNodeText(ctx.SourceFile, params[1])
					replacement := fmt.Sprintf("{ [key: %s]: %s }", keyText, typeText)

					ctx.ReportNodeWithSuggestions(node, buildPreferIndexSignatureMessage(),
						rule.RuleSuggestion{
							Message: buildPreferIndexSignatureSuggestionMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, node, replacement),
							},
						})
				}
			}
		} else if mode == "record" {
			// Handle record mode
			listeners[ast.KindInterfaceDeclaration] = func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()

				var genericTypes string
				if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
					var typeParams []string
					for _, param := range interfaceDecl.TypeParameters.Nodes {
						typeParams = append(typeParams, getNodeText(ctx.SourceFile, param))
					}
					genericTypes = fmt.Sprintf("<%s>", strings.Join(typeParams, ", "))
				}

				var interfaceName string
				var nameId *ast.Identifier
				if name := interfaceDecl.Name(); name != nil && name.Kind == ast.KindIdentifier {
					nameId = name.AsIdentifier()
					interfaceName = nameId.Text
				}
				// Process interface members

				prefix := fmt.Sprintf("type %s%s = ", interfaceName, genericTypes)
				postfix := ";"

				// Check if interface extends other types - if so, don't auto-fix
				safeFix := true
				if interfaceDecl.HeritageClauses != nil {
					for _, clause := range interfaceDecl.HeritageClauses.Nodes {
						if clause.Kind == ast.KindHeritageClause {
							heritageClause := clause.AsHeritageClause()
							if len(heritageClause.Types.Nodes) > 0 {
								safeFix = false
								break
							}
						}
					}
				}

				checkMembers(interfaceDecl.Members.Nodes, node, nameId, prefix, postfix, safeFix)
			}

			listeners[ast.KindMappedType] = func(node *ast.Node) {
				mappedType := node.AsMappedTypeNode()

				if mappedType.TypeParameter == nil {
					return
				}
				typeParam := mappedType.TypeParameter.AsTypeParameter()
				constraint := typeParam.Constraint
				if constraint == nil {
					return
				}

				// If the key is used to compute the value, we can't convert to a Record
				if mappedType.Type != nil {
					var keyName string
					if name := typeParam.Name(); name != nil && name.Kind == ast.KindIdentifier {
						keyName = name.AsIdentifier().Text
					}
					valueText := getNodeText(ctx.SourceFile, mappedType.Type)
					if strings.Contains(valueText, keyName) {
						return
					}
				}

				// Skip keyof operators that are not parenthesized
				if constraint.Kind == ast.KindTypeOperator {
					typeOp := constraint.AsTypeOperatorNode()
					if typeOp.Operator == ast.KindKeyOfKeyword {
						return
					}
				}

				// Check for circular references in mapped types
				parentDecl := findParentDeclaration(node)
				if parentDecl != nil && mappedType.Type != nil {
					if parentName := parentDecl.Name(); parentName != nil && parentName.Kind == ast.KindIdentifier {
						parentNameText := parentName.AsIdentifier().Text
						valueText := getNodeText(ctx.SourceFile, mappedType.Type)
						// Skip if the mapped type references its parent type (circular)
						if strings.Contains(valueText, parentNameText) {
							return
						}
						
						// Also check for indirect circular references
						// For example: type Bar<K> = { [k in K]: Foo }; type Foo = Bar;
						if hasCircularReference(ctx.SourceFile, parentNameText) {
							return
						}
					}
				}

				// Check if we can provide a fix
				canFix := true
				if mappedType.ReadonlyToken != nil {
					readonlyText := getNodeText(ctx.SourceFile, mappedType.ReadonlyToken)
					if strings.Contains(readonlyText, "-") {
						canFix = false
					}
				}

				if !canFix {
					ctx.ReportNode(node, buildPreferRecordMessage())
					return
				}

				keyType := getNodeText(ctx.SourceFile, constraint)
				var valueType string
				if mappedType.Type != nil {
					valueType = getNodeText(ctx.SourceFile, mappedType.Type)
				} else {
					valueType = "any"
				}

				recordText := fmt.Sprintf("Record<%s, %s>", keyType, valueType)

				// Handle optional modifiers
				if mappedType.QuestionToken != nil {
					questionText := getNodeText(ctx.SourceFile, mappedType.QuestionToken)
					if questionText == "+?" || questionText == "?" {
						recordText = fmt.Sprintf("Partial<%s>", recordText)
					} else if questionText == "-?" {
						recordText = fmt.Sprintf("Required<%s>", recordText)
					}
				}

				// Handle readonly modifiers
				if mappedType.ReadonlyToken != nil {
					readonlyText := getNodeText(ctx.SourceFile, mappedType.ReadonlyToken)
					if readonlyText == "+readonly" || readonlyText == "readonly" {
						recordText = fmt.Sprintf("Readonly<%s>", recordText)
					}
				}

				ctx.ReportNodeWithFixes(node, buildPreferRecordMessage(),
					rule.RuleFixReplace(ctx.SourceFile, node, recordText))
			}

			listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
				typeLiteral := node.AsTypeLiteralNode()
				parent := findParentDeclaration(node)
				var parentId *ast.Identifier
				if parent != nil {
					if name := parent.Name(); name != nil && name.Kind == ast.KindIdentifier {
						parentId = name.AsIdentifier()
						// Debug: print parent name
						// fmt.Printf("TypeLiteral parent: %s\n", parentId.Text)
					}
				}
				checkMembers(typeLiteral.Members.Nodes, node, parentId, "", "", true)
			}
		}

		return listeners
	},
}
