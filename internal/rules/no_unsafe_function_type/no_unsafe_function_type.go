package no_unsafe_function_type

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

func buildBannedFunctionTypeMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "bannedFunctionType",
		Description: "The `Function` type accepts any function-like value.\nPrefer explicitly defining any function parameters and return type.",
	}
}

// isReferenceToGlobalFunction checks if the given identifier node references the global Function type
func isReferenceToGlobalFunction(ctx rule.RuleContext, node *ast.Node) bool {
	if !ast.IsIdentifier(node) || node.AsIdentifier().Text != "Function" {
		return false
	}

	// Get the symbol for the identifier
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}

	// Check if this symbol is from the default library (lib.*.d.ts)
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		
		sourceFile := ast.GetSourceFileOfNode(declaration)
		if sourceFile == nil {
			continue
		}
		
		// If any declaration is NOT from the default library, this is user-defined
		if !utils.IsSourceFileDefaultLibrary(ctx.Program, sourceFile) {
			return false
		}
	}
	
	// If we have declarations and they're all from the default library, this is the global Function
	return len(symbol.Declarations) > 0
}


var NoUnsafeFunctionTypeRule = rule.Rule{
	Name: "no-unsafe-function-type",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkBannedTypes := func(node *ast.Node) {
			if isReferenceToGlobalFunction(ctx, node) {
				ctx.ReportNode(node, buildBannedFunctionTypeMessage())
			}
		}

		return rule.RuleListeners{
			// Check type references like: let value: Function;
			ast.KindTypeReference: func(node *ast.Node) {
				typeRef := node.AsTypeReferenceNode()
				checkBannedTypes(typeRef.TypeName)
			},

			// Check class implements clauses like: class Foo implements Function {}
			ast.KindHeritageClause: func(node *ast.Node) {
				heritageClause := node.AsHeritageClause()
				
				// Only check implements and extends clauses
				if heritageClause.Token != ast.KindImplementsKeyword && heritageClause.Token != ast.KindExtendsKeyword {
					return
				}

				// Check if this is a class implements or interface extends
				parent := node.Parent
				if parent == nil {
					return
				}

				isClassImplements := ast.IsClassDeclaration(parent) && heritageClause.Token == ast.KindImplementsKeyword
				isInterfaceExtends := ast.IsInterfaceDeclaration(parent) && heritageClause.Token == ast.KindExtendsKeyword

				if !isClassImplements && !isInterfaceExtends {
					return
				}

				// Check each type in the heritage clause
				for _, heritageType := range heritageClause.Types.Nodes {
					if heritageType.AsExpressionWithTypeArguments().Expression != nil {
						checkBannedTypes(heritageType.AsExpressionWithTypeArguments().Expression)
					}
				}
			},
		}
	},
}