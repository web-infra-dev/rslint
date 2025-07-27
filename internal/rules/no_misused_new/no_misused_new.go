package no_misused_new

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
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
	if parent == nil {
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
				
				// Check if the method name is 'new'
				methodName, _ := utils.GetNameFromMember(ctx.SourceFile, &methodDecl.Node)
				if methodName != "new" {
					return
				}

				// Check if it's in a class body
				if node.Parent == nil || node.Parent.Kind != ast.KindBlock {
					return
				}

				classBody := node.Parent
				if classBody.Parent == nil {
					return
				}

				// Check if the method has an empty body (TSEmptyBodyFunctionExpression)
				if methodDecl.Body != nil {
					// Method has a body, so it's OK
					return
				}

				// Check if the return type matches the parent class
				var returnType *ast.Node
				if methodDecl.Type != nil {
					returnType = methodDecl.Type
				}

				if isMatchingParentType(classBody.Parent, returnType) {
					ctx.ReportNode(node, buildErrorMessageClassMessage())
				}
			},

			// Check for interface constructor signatures
			ast.KindConstructSignature: func(node *ast.Node) {
				constructSig := node.AsConstructSignatureDeclaration()

				// Check if it's in an interface body
				if node.Parent == nil || node.Parent.Kind != ast.KindBlock {
					return
				}

				interfaceBody := node.Parent
				if interfaceBody.Parent == nil || interfaceBody.Parent.Kind != ast.KindInterfaceDeclaration {
					return
				}

				// Check if the return type matches the parent interface
				var returnType *ast.Node
				if constructSig.Type != nil {
					returnType = constructSig.Type
				}

				if isMatchingParentType(interfaceBody.Parent, returnType) {
					ctx.ReportNode(node, buildErrorMessageInterfaceMessage())
				}
			},

			// Check for interface method signatures named 'constructor'
			ast.KindMethodSignature: func(node *ast.Node) {
				methodSig := node.AsMethodSignatureDeclaration()

				// Check if the method name is 'constructor'
				methodName, _ := utils.GetNameFromMember(ctx.SourceFile, &methodSig.Node)
				if methodName != "constructor" {
					return
				}

				// Report error for any method signature named 'constructor' in interfaces
				ctx.ReportNode(node, buildErrorMessageInterfaceMessage())
			},
		}
	},
}