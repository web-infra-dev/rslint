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

// Check if interface has circular reference by checking if there are mutual references in the file
func hasCircularReference(sourceFile *ast.SourceFile, interfaceName string) bool {
	if interfaceName == "" {
		return false
	}
	
	// Build a map of interface name -> value type
	interfaceTypes := make(map[string]string)
	
	sourceFile.ForEachChild(func(node *ast.Node) bool {
		if node.Kind == ast.KindInterfaceDeclaration {
			interfaceDecl := node.AsInterfaceDeclaration()
			if name := interfaceDecl.Name(); name != nil && name.Kind == ast.KindIdentifier {
				declName := name.AsIdentifier().Text
				
				// Check if it has a single index signature
				if len(interfaceDecl.Members.Nodes) == 1 {
					member := interfaceDecl.Members.Nodes[0]
					if member.Kind == ast.KindIndexSignature {
						indexSig := member.AsIndexSignatureDeclaration()
						if indexSig.Type != nil {
							valueText := getNodeText(sourceFile, indexSig.Type)
							interfaceTypes[declName] = valueText
						}
					}
				}
			}
		}
		return false // Continue traversing
	})
	
	// Check for circular reference: A -> B and B -> A
	if valueType, exists := interfaceTypes[interfaceName]; exists {
		if referencedType, hasRef := interfaceTypes[valueType]; hasRef {
			if referencedType == interfaceName {
				return true
			}
		}
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
			if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
				if styleStr, ok := optsArray[0].(string); ok {
					opts.Style = styleStr
				}
			}
		}

		mode := opts.Style

		// Helper function to check members for index signatures
		checkMembers := func(members []*ast.Node, node *ast.Node, parentId *ast.Identifier, prefix, postfix string, safeFix bool) {
			if len(members) != 1 {
				return
			}
			
			member := members[0]
			if member.Kind != ast.KindIndexSignature {
				return
			}
			
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
			
			// Skip conversion if the value type is itself a convertible type literal
			// This ensures we convert innermost type literals first
			if valueType.Kind == ast.KindTypeLiteral {
				typeLiteral := valueType.AsTypeLiteralNode()
				if len(typeLiteral.Members.Nodes) == 1 && typeLiteral.Members.Nodes[0].Kind == ast.KindIndexSignature {
					// The value type is a convertible type literal, so skip this one
					return
				}
			}
			
			// Check for circular references
			if parentId != nil {
				parentName := parentId.Text
				valueText := getNodeText(ctx.SourceFile, valueType)
				
				// For interfaces, check for mutual circular references
				if node.Parent != nil && node.Parent.Kind == ast.KindInterfaceDeclaration {
					if hasCircularReference(ctx.SourceFile, parentName) {
						return
					}
				}
				
				// For type aliases with direct self-reference, skip conversion
				if valueText == parentName {
					// Check if this is a top-level type alias
					parentDecl := findParentDeclaration(node)
					if parentDecl != nil {
						return
					}
				}
				
				// Skip if the value already contains "Record" and references the parent type
				// This prevents creating deeply nested Record types like Record<string, Record<string, Foo>>
				if strings.Contains(valueText, "Record") && strings.Contains(valueText, parentName) {
					return
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
		}

		// Handle record mode
		if mode == "record" {
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
				if name := interfaceDecl.Name(); name != nil && name.Kind == ast.KindIdentifier {
					interfaceName = name.AsIdentifier().Text
				}
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
				
				var nameId *ast.Identifier
				if name := interfaceDecl.Name(); name != nil && name.Kind == ast.KindIdentifier {
					nameId = name.AsIdentifier()
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
					}
				}
				checkMembers(typeLiteral.Members.Nodes, node, parentId, "", "", true)
			}
		}

		return listeners
	},
}