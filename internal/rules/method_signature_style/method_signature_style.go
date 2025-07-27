package method_signature_style

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

func getNodeText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
}

type MethodSignatureStyleOptions struct {
	Mode string `json:"mode"`
}

func getMethodKey(ctx rule.RuleContext, node *ast.Node) string {
	var name string
	var optional bool
	
	if node.Kind == ast.KindPropertySignature {
		propSig := node.AsPropertySignatureDeclaration()
		if propSig.Name() != nil {
			name = getNodeText(ctx, propSig.Name())
			optional = propSig.PostfixToken != nil && propSig.PostfixToken.Kind == ast.KindQuestionToken
		}
	} else if node.Kind == ast.KindMethodSignature {
		methodSig := node.AsMethodSignatureDeclaration()
		if methodSig.Name() != nil {
			name = getNodeText(ctx, methodSig.Name())
			optional = methodSig.PostfixToken != nil && methodSig.PostfixToken.Kind == ast.KindQuestionToken
		}
	}
	
	if optional {
		name = fmt.Sprintf("%s?", name)
	}
	
	return name
}

func getMethodParams(ctx rule.RuleContext, node *ast.Node) string {
	var params []*ast.Node
	var typeParams *ast.NodeList
	
	if node.Kind == ast.KindMethodSignature {
		methodSig := node.AsMethodSignatureDeclaration()
		params = methodSig.Parameters.Nodes
		typeParams = methodSig.TypeParameters
	} else if node.Kind == ast.KindFunctionType {
		funcType := node.AsFunctionTypeNode()
		params = funcType.Parameters.Nodes
		typeParams = funcType.TypeParameters
	}
	
	paramsStr := "()"
	if len(params) > 0 {
		// Find opening and closing parentheses
		s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Pos())
		openParenPos := -1
		closeParenPos := -1
		
		// Find opening paren before first parameter
		for s.TokenStart() < params[0].Pos() {
			if s.Token() == ast.KindOpenParenToken {
				openParenPos = s.TokenStart()
			}
			s.Scan()
		}
		
		// Find closing paren after last parameter
		lastParam := params[len(params)-1]
		s = scanner.GetScannerForSourceFile(ctx.SourceFile, lastParam.End())
		for s.TokenStart() < node.End() {
			if s.Token() == ast.KindCloseParenToken {
				closeParenPos = s.TokenEnd()
				break
			}
			s.Scan()
		}
		
		if openParenPos != -1 && closeParenPos != -1 {
			paramsStr = string(ctx.SourceFile.Text()[openParenPos:closeParenPos])
		}
	}
	
	if typeParams != nil && len(typeParams.Nodes) > 0 {
		typeParamsText := string(ctx.SourceFile.Text()[typeParams.Pos():typeParams.End()])
		// Extract just the type parameters part
		start := strings.Index(typeParamsText, "<")
		end := strings.Index(typeParamsText, ">")
		if start != -1 && end != -1 {
			paramsStr = typeParamsText[start:end+1] + paramsStr
		}
	}
	
	return paramsStr
}

func getMethodReturnType(ctx rule.RuleContext, node *ast.Node) string {
	var typeNode *ast.Node
	
	if node.Kind == ast.KindMethodSignature {
		methodSig := node.AsMethodSignatureDeclaration()
		typeNode = methodSig.Type
	} else if node.Kind == ast.KindFunctionType {
		funcType := node.AsFunctionTypeNode()
		typeNode = funcType.Type
	}
	
	if typeNode == nil {
		return "any"
	}
	
	typeRange := utils.TrimNodeTextRange(ctx.SourceFile, typeNode)
	return string(ctx.SourceFile.Text()[typeRange.Pos():typeRange.End()])
}

func getDelimiter(ctx rule.RuleContext, node *ast.Node) string {
	// Find the last token of the node
	s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.End()-1)
	s.Scan()
	
	if s.Token() == ast.KindSemicolonToken || s.Token() == ast.KindCommaToken {
		return string(ctx.SourceFile.Text()[s.TokenStart():s.TokenEnd()])
	}
	
	return ""
}

func isNodeParentModuleDeclaration(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		if parent.Kind == ast.KindModuleDeclaration {
			return true
		}
		if parent.Kind == ast.KindSourceFile {
			return false
		}
		parent = parent.Parent
	}
	return false
}

var MethodSignatureStyleRule = rule.Rule{
	Name: "method-signature-style",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := MethodSignatureStyleOptions{
			Mode: "property", // default
		}
		
		if options != nil {
			switch v := options.(type) {
			case string:
				opts.Mode = v
			case []interface{}:
				if len(v) > 0 {
					if str, ok := v[0].(string); ok {
						opts.Mode = str
					}
				}
			}
		}
		
		listeners := rule.RuleListeners{}
		
		if opts.Mode == "property" {
			listeners[ast.KindMethodSignature] = func(node *ast.Node) {
				methodNode := node.AsMethodSignatureDeclaration()
			if methodNode.Name() == nil {
				return
			}
				
				// Skip getters and setters
			nameText := ""
			if methodNode.Name() != nil {
				if methodNode.Name().Kind == ast.KindIdentifier {
					nameText = methodNode.Name().AsIdentifier().Text
				} else if methodNode.Name().Kind == ast.KindStringLiteral {
					nameText = methodNode.Name().AsStringLiteral().Text
				}
			}
			if nameText == "get" || nameText == "set" {
				return
			}
				
				parent := node.Parent
				var members []*ast.Node
				
				if parent.Kind == ast.KindInterfaceDeclaration {
				interfaceDecl := parent.AsInterfaceDeclaration()
				members = interfaceDecl.Members.Nodes
			} else if parent.Kind == ast.KindTypeLiteral {
				typeLit := parent.AsTypeLiteralNode()
				members = typeLit.Members.Nodes
			}
				
				// Find duplicate method signatures
				key := getMethodKey(ctx, node)
				var duplicates []*ast.Node
				
				for _, member := range members {
					if member == node || member.Kind != ast.KindMethodSignature {
						continue
					}
					if getMethodKey(ctx, member) == key {
						duplicates = append(duplicates, member)
					}
				}
				
				isParentModule := isNodeParentModuleDeclaration(node)
				
				if len(duplicates) > 0 {
					if isParentModule {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "errorMethod",
							Description: "Shorthand method signature is forbidden. Use a function property instead.",
						})
					} else {
						// Combine all overloads into intersection type
						allMethods := append([]*ast.Node{node}, duplicates...)
						
						// Sort by position
						for i := 0; i < len(allMethods); i++ {
							for j := i + 1; j < len(allMethods); j++ {
								if allMethods[i].Pos() > allMethods[j].Pos() {
									allMethods[i], allMethods[j] = allMethods[j], allMethods[i]
								}
							}
						}
						
						var typeStrings []string
						for _, method := range allMethods {
							params := getMethodParams(ctx, method)
							returnType := getMethodReturnType(ctx, method)
							typeStrings = append(typeStrings, fmt.Sprintf("(%s => %s)", params, returnType))
						}
						
						typeString := strings.Join(typeStrings, " & ")
						delimiter := getDelimiter(ctx, node)
						
						fixes := []rule.RuleFix{
							rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("%s: %s%s", key, typeString, delimiter)),
						}
						
						// Remove duplicate methods
						for _, dup := range duplicates {
							// Find any whitespace/comments between this node and the next
							nextNode := findNextMember(members, dup)
							if nextNode != nil {
								fixes = append(fixes, rule.RuleFixReplaceRange(
							core.NewTextRange(dup.Pos(), nextNode.Pos()),
							"",
						))
							} else {
								fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, dup, ""))
							}
						}
						
						ctx.ReportNodeWithFixes(node, rule.RuleMessage{
							Id:          "errorMethod",
							Description: "Shorthand method signature is forbidden. Use a function property instead.",
						}, fixes...)
					}
					return
				}
				
				// Single method signature
				if isParentModule {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "errorMethod",
						Description: "Shorthand method signature is forbidden. Use a function property instead.",
					})
				} else {
					key := getMethodKey(ctx, node)
					params := getMethodParams(ctx, node)
					returnType := getMethodReturnType(ctx, node)
					delimiter := getDelimiter(ctx, node)
					
					ctx.ReportNodeWithFixes(node, rule.RuleMessage{
						Id:          "errorMethod",
						Description: "Shorthand method signature is forbidden. Use a function property instead.",
					}, rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("%s: %s => %s%s", key, params, returnType, delimiter)))
				}
			}
		}
		
		if opts.Mode == "method" {
			listeners[ast.KindPropertySignature] = func(node *ast.Node) {
				propNode := node.AsPropertySignatureDeclaration()
				if propNode.Type == nil || propNode.Type.Kind != ast.KindTypeReference {
					return
				}
				
				typeRef := propNode.Type.AsTypeReference()
				if typeRef.TypeName == nil || typeRef.TypeName.Kind != ast.KindFunctionType {
					// Check if the type reference points to a function type
					if propNode.Type.Kind == ast.KindFunctionType {
						funcType := propNode.Type.AsFunctionTypeNode()
						
						key := getMethodKey(ctx, node)
				params := getMethodParams(ctx, funcType.AsNode())
				returnType := getMethodReturnType(ctx, funcType.AsNode())
						delimiter := getDelimiter(ctx, node)
						
						ctx.ReportNodeWithFixes(node, rule.RuleMessage{
							Id:          "errorProperty",
							Description: "Function property signature is forbidden. Use a method shorthand instead.",
						}, rule.RuleFixReplace(ctx.SourceFile, node, fmt.Sprintf("%s%s: %s%s", key, params, returnType, delimiter)))
					}
				}
			}
		}
		
		return listeners
	},
}

func findNextMember(members []*ast.Node, current *ast.Node) *ast.Node {
	found := false
	for _, member := range members {
		if found {
			return member
		}
		if member == current {
			found = true
		}
	}
	return nil
}