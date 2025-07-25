package consistent_generic_constructors

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

func buildPreferConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferConstructor",
		Description: "The generic type arguments should be specified as part of the constructor type arguments.",
	}
}

func buildPreferTypeAnnotationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferTypeAnnotation", 
		Description: "The generic type arguments should be specified as part of the type annotation.",
	}
}

type lhsRhsPair struct {
	lhs *ast.Node // The left-hand side (identifier/binding pattern)  
	rhs *ast.Node // The right-hand side (initializer/value)
}

func getLHSRHS(node *ast.Node) *lhsRhsPair {
	switch node.Kind {
	case ast.KindVariableDeclaration:
		varDecl := node.AsVariableDeclaration()
		return &lhsRhsPair{
			lhs: varDecl.Name(),
			rhs: varDecl.Initializer,
		}
	case ast.KindPropertyDeclaration:
		propDecl := node.AsPropertyDeclaration()
		return &lhsRhsPair{
			lhs: node,
			rhs: propDecl.Initializer,
		}
	case ast.KindParameter:
		param := node.AsParameterDeclaration()
		paramName := param.Name()
		// Check if the parameter name is a binding pattern
		if paramName != nil && (paramName.Kind == ast.KindObjectBindingPattern || paramName.Kind == ast.KindArrayBindingPattern) {
			return &lhsRhsPair{
				lhs: paramName, // Use the binding pattern as LHS
				rhs: param.Initializer,
			}
		}
		return &lhsRhsPair{
			lhs: paramName,
			rhs: param.Initializer,
		}
	default:
		return &lhsRhsPair{lhs: nil, rhs: nil}
	}
}

func getTypeAnnotation(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindIdentifier:
		if node.Parent != nil && node.Parent.Kind == ast.KindParameter {
			param := node.Parent.AsParameterDeclaration()
			return param.Type
		}
		if node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration {
			varDecl := node.Parent.AsVariableDeclaration()
			return varDecl.Type
		}
		return nil
	case ast.KindPropertyDeclaration:
		propDecl := node.AsPropertyDeclaration()
		return propDecl.Type
	case ast.KindGetAccessor:
		accessor := node.AsGetAccessorDeclaration()
		return accessor.Type
	case ast.KindSetAccessor:
		accessor := node.AsSetAccessorDeclaration()
		if accessor.Parameters != nil && len(accessor.Parameters.Nodes) > 0 {
			param := accessor.Parameters.Nodes[0].AsParameterDeclaration()
			return param.Type
		}
		return nil
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		// For binding patterns that are parameter names, look at the parent parameter
		if node.Parent != nil && node.Parent.Kind == ast.KindParameter {
			param := node.Parent.AsParameterDeclaration()
			return param.Type
		}
		return nil
	default:
		return nil
	}
}

func isNewExpressionWithIdentifier(node *ast.Node) (*ast.Node, *ast.NodeList, bool) {
	if node == nil || node.Kind != ast.KindNewExpression {
		return nil, nil, false
	}

	newExpr := node.AsNewExpression()
	if newExpr.Expression == nil || newExpr.Expression.Kind != ast.KindIdentifier {
		return nil, nil, false
	}

	return newExpr.Expression, newExpr.TypeArguments, true
}

func isTypeReferenceWithSameName(typeNode *ast.Node, identifierName string) (*ast.NodeList, bool) {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return nil, false
	}

	typeRef := typeNode.AsTypeReference()
	if typeRef.TypeName == nil || typeRef.TypeName.Kind != ast.KindIdentifier {
		return nil, false
	}

	identifier := typeRef.TypeName.AsIdentifier()
	if identifier.Text != identifierName {
		return nil, false
	}

	return typeRef.TypeArguments, true
}

func getNodeText(ctx rule.RuleContext, node *ast.Node) string {
	if node == nil {
		return ""
	}
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
}

func getNodeListTextWithBrackets(ctx rule.RuleContext, nodeList *ast.NodeList) string {
	if nodeList == nil {
		return ""
	}
	// Find the opening and closing angle brackets using scanner
	openBracketPos := nodeList.Pos() - 1
	
	// Find closing bracket after the nodeList
	s := scanner.GetScannerForSourceFile(ctx.SourceFile, nodeList.End())
	closeBracketPos := nodeList.End()
	for s.TokenStart() < ctx.SourceFile.End() {
		if s.Token() == ast.KindGreaterThanToken {
			closeBracketPos = s.TokenEnd()
			break
		}
		if s.Token() != ast.KindWhitespaceTrivia && s.Token() != ast.KindNewLineTrivia {
			break
		}
		s.Scan()
	}
	
	textRange := core.NewTextRange(openBracketPos, closeBracketPos)
	return string(ctx.SourceFile.Text()[textRange.Pos():textRange.End()])
}

func hasParenthesesAfter(ctx rule.RuleContext, node *ast.Node) bool {
	s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.End())
	for s.TokenStart() < ctx.SourceFile.End() {
		token := s.Token()
		if token == ast.KindOpenParenToken {
			return true
		}
		if token != ast.KindWhitespaceTrivia && token != ast.KindNewLineTrivia {
			break
		}
		s.Scan()
	}
	return false
}

func getIDToAttachAnnotation(ctx rule.RuleContext, node *ast.Node, lhsName *ast.Node) *ast.Node {
	if node.Kind == ast.KindPropertyDeclaration {
		propDecl := node.AsPropertyDeclaration()
		// Check if property is computed (e.g., [key]: type)
		if propDecl.Name() != nil && propDecl.Name().Kind == ast.KindComputedPropertyName {
			// For computed properties, attach after the closing bracket
			return propDecl.Name() // Attach after the entire computed property name [key]
		}
		return propDecl.Name()
	}
	
	// For binding patterns, attach after the pattern itself
	if node.Kind == ast.KindObjectBindingPattern || node.Kind == ast.KindArrayBindingPattern {
		return lhsName
	}

	return lhsName
}

var ConsistentGenericConstructorsRule = rule.Rule{
	Name: "consistent-generic-constructors",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		mode := "constructor" // default
		
		// Parse options - can be a string directly or in a map
		if options != nil {
			if modeStr, ok := options.(string); ok {
				mode = modeStr
			} else if optionsMap, ok := options.(map[string]interface{}); ok {
				if modeStr, ok := optionsMap["mode"].(string); ok {
					mode = modeStr
				} else if modeStr, ok := optionsMap["value"].(string); ok {
					mode = modeStr
				}
			} else if optionsSlice, ok := options.([]interface{}); ok && len(optionsSlice) > 0 {
				if modeStr, ok := optionsSlice[0].(string); ok {
					mode = modeStr
				}
			}
		}

		handleNode := func(node *ast.Node) {
			pair := getLHSRHS(node)
			if pair.rhs == nil {
				return
			}

			// Check if RHS is a new expression with identifier
			callee, rhsTypeArgs, isValidNewExpr := isNewExpressionWithIdentifier(pair.rhs)
			if !isValidNewExpr {
				return
			}

			calleeText := getNodeText(ctx, callee)
			lhsTypeAnnotation := getTypeAnnotation(pair.lhs)

			// Check if LHS type annotation matches the constructor name
			var lhsTypeArgs *ast.NodeList
			var typeMatches bool
			if lhsTypeAnnotation != nil {
				lhsTypeArgs, typeMatches = isTypeReferenceWithSameName(lhsTypeAnnotation, calleeText)
				if !typeMatches {
					return
				}
			}

			// Only process if there are generics involved
			if lhsTypeAnnotation == nil && rhsTypeArgs == nil {
				// No generics anywhere, nothing to check
				return
			}

			if mode == "type-annotation" {
				// Prefer type annotation mode
				if (lhsTypeAnnotation == nil || lhsTypeArgs == nil) && rhsTypeArgs != nil {
					// No type annotation or no type args in annotation but constructor has type args - move to type annotation
					calleeText := getNodeText(ctx, callee)
					typeArgsText := getNodeListTextWithBrackets(ctx, rhsTypeArgs)
					typeAnnotation := calleeText + typeArgsText

					idToAttach := getIDToAttachAnnotation(ctx, node, pair.lhs)

					// Find the range to remove (including angle brackets)
					openBracketPos := rhsTypeArgs.Pos() - 1
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, rhsTypeArgs.End())
					closeBracketPos := rhsTypeArgs.End()
					for s.TokenStart() < ctx.SourceFile.End() {
						if s.Token() == ast.KindGreaterThanToken {
							closeBracketPos = s.TokenEnd()
							break
						}
						if s.Token() != ast.KindWhitespaceTrivia && s.Token() != ast.KindNewLineTrivia {
							break
						}
						s.Scan()
					}

					// Determine what node to report the error on
					reportNode := node
					if node.Kind == ast.KindVariableDeclaration && node.Parent != nil {
						reportNode = node.Parent // For variable declarations, report on the statement
					}
					
					ctx.ReportNodeWithFixes(reportNode, buildPreferTypeAnnotationMessage(),
						rule.RuleFixRemoveRange(core.NewTextRange(openBracketPos, closeBracketPos)),
						rule.RuleFixInsertAfter(idToAttach, ": "+typeAnnotation),
					)
				}
			} else {
				// Prefer constructor mode (default)
				if lhsTypeAnnotation != nil && lhsTypeArgs != nil && rhsTypeArgs == nil {
					// Type annotation has type args but constructor doesn't - move to constructor
					hasParens := hasParenthesesAfter(ctx, callee)
					typeArgsText := getNodeListTextWithBrackets(ctx, lhsTypeArgs)

					// Find the colon token before the type annotation
					var fixes []rule.RuleFix
					if lhsTypeAnnotation.Parent != nil {
						s := scanner.GetScannerForSourceFile(ctx.SourceFile, lhsTypeAnnotation.Parent.Pos())
						colonStart := -1
						for s.TokenStart() < lhsTypeAnnotation.Pos() {
							if s.Token() == ast.KindColonToken {
								colonStart = s.TokenStart()
							}
							s.Scan()
						}
						
						if colonStart != -1 {
							fixes = append(fixes, rule.RuleFixReplaceRange(
								core.NewTextRange(colonStart, lhsTypeAnnotation.End()),
								"",
							))
						}
					}

					fixes = append(fixes, rule.RuleFixInsertAfter(callee, typeArgsText))

					if !hasParens {
						fixes = append(fixes, rule.RuleFixInsertAfter(callee, "()"))
					}

					// Determine what node to report the error on
					reportNode := node
					if node.Kind == ast.KindVariableDeclaration && node.Parent != nil {
						reportNode = node.Parent // For variable declarations, report on the statement
					} else if node.Kind == ast.KindParameter && pair.lhs != nil {
						reportNode = pair.lhs // For parameters, report on the parameter name/pattern
					}
					ctx.ReportNodeWithFixes(reportNode, buildPreferConstructorMessage(), fixes...)
				}
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclaration: handleNode,
			ast.KindPropertyDeclaration: handleNode,
			ast.KindParameter:           handleNode,
		}
	},
}