package no_misused_new

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildErrorMessageClassMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorMessageClass",
		Description: "Class cannot have method named `new`.",
	}
}

func buildErrorMessageInterfaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorMessageInterface",
		Description: "Interfaces cannot be constructed, only classes.",
	}
}

// getTypeReferenceName extracts the name from various type nodes
func getTypeReferenceName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindTypeReference:
		typeRef := node.AsTypeReferenceNode()
		return getTypeReferenceName(typeRef.TypeName)
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	default:
		return ""
	}
}

// isMatchingParentType checks if the return type matches the parent class/interface name
func isMatchingParentType(parent *ast.Node, returnType *ast.Node) bool {
	if parent == nil || returnType == nil {
		return false
	}

	var parentName string
	switch parent.Kind {
	case ast.KindClassDeclaration:
		classDecl := parent.AsClassDeclaration()
		if classDecl.Name() != nil {
			parentName = classDecl.Name().AsIdentifier().Text
		}
	case ast.KindClassExpression:
		classExpr := parent.AsClassExpression()
		if classExpr.Name() != nil {
			parentName = classExpr.Name().AsIdentifier().Text
		}
	case ast.KindInterfaceDeclaration:
		interfaceDecl := parent.AsInterfaceDeclaration()
		parentName = interfaceDecl.Name().AsIdentifier().Text
	default:
		return false
	}

	if parentName == "" {
		return false
	}

	returnTypeName := getTypeReferenceName(returnType)
	return returnTypeName == parentName
}

var NoMisusedNewRule = rule.Rule{
	Name: "no-misused-new",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{

			// Check for class methods named 'new'
			ast.KindMethodDeclaration: func(node *ast.Node) {
				methodDecl := node.AsMethodDeclaration()

				// Check if the method name is 'new' - use direct approach like method signatures
				if methodDecl.Name() == nil || methodDecl.Name().Kind != ast.KindIdentifier {
					return
				}
				methodName := methodDecl.Name().AsIdentifier().Text
				if methodName != "new" {
					return
				}

				// Check if the method has a body - if it does, it's OK
				if methodDecl.Body != nil {
					return
				}

				// Find the parent class/interface and check return type matching
				parent := node.Parent
				for parent != nil {
					if parent.Kind == ast.KindClassDeclaration {
						// For class declarations, only flag if return type matches class name
						var returnType *ast.Node
						if methodDecl.Type != nil {
							returnType = methodDecl.Type
						}

						if isMatchingParentType(parent, returnType) {
							ctx.ReportNode(node, buildErrorMessageClassMessage())
						}
						return
					} else if parent.Kind == ast.KindClassExpression {
						// Class expressions are generally OK unless they have a name that matches return type
						// Based on the test cases, class expressions with "new(): X;" should be allowed
						// Only flag if it's a named class expression and the return type matches
						classExpr := parent.AsClassExpression()
						if classExpr.Name() != nil {
							var returnType *ast.Node
							if methodDecl.Type != nil {
								returnType = methodDecl.Type
							}

							if isMatchingParentType(parent, returnType) {
								ctx.ReportNode(node, buildErrorMessageClassMessage())
							}
						}
						return
					}
					parent = parent.Parent
				}
			},

			// Check for interface constructor signatures (new (): Type)
			ast.KindConstructSignature: func(node *ast.Node) {
				constructSig := node.AsConstructSignatureDeclaration()

				// Find the parent interface
				parent := node.Parent
				for parent != nil {
					if parent.Kind == ast.KindInterfaceDeclaration {
						// Check if the return type matches the parent interface
						var returnType *ast.Node
						if constructSig.Type != nil {
							returnType = constructSig.Type
						}

						if isMatchingParentType(parent, returnType) {
							ctx.ReportNode(node, buildErrorMessageInterfaceMessage())
						}
						return
					} else if parent.Kind == ast.KindTypeLiteral {
						// 'new' in type literal is OK - we don't know the type name
						return
					}
					parent = parent.Parent
				}
			},

			// Check for interface/type method signatures named 'constructor' or class method signatures named 'new'
			ast.KindMethodSignature: func(node *ast.Node) {
				methodSig := node.AsMethodSignatureDeclaration()

				// Get the name directly from the method signature's name property
				if methodSig.Name() != nil && methodSig.Name().Kind == ast.KindIdentifier {
					methodName := methodSig.Name().AsIdentifier().Text

					if methodName == "constructor" {
						// Always flag any method signature named 'constructor' in interfaces/types
						ctx.ReportNode(node, buildErrorMessageInterfaceMessage())
					} else if methodName == "new" {
						// For method signatures named 'new', only flag in class declarations
						// (not class expressions) and only if return type matches class name
						parent := node.Parent
						for parent != nil {
							if parent.Kind == ast.KindClassDeclaration {
								// Check if the return type matches the parent class name
								var returnType *ast.Node
								if methodSig.Type != nil {
									returnType = methodSig.Type
								}

								if isMatchingParentType(parent, returnType) {
									ctx.ReportNode(node, buildErrorMessageClassMessage())
								}
								return
							}
							parent = parent.Parent
						}
					}
				}
			},

			// Also check property signatures that might be 'constructor'
			ast.KindPropertySignature: func(node *ast.Node) {
				propSig := node.AsPropertySignatureDeclaration()

				// Get the name directly from the property signature's name property
				if propSig.Name() != nil && propSig.Name().Kind == ast.KindIdentifier {
					methodName := propSig.Name().AsIdentifier().Text
					if methodName == "constructor" {
						// Always flag any property signature named 'constructor' in interfaces
						ctx.ReportNode(node, buildErrorMessageInterfaceMessage())
					}
				}
			},
		}
	},
}
