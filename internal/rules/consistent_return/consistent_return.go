package consistent_return

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

func buildMissingReturnMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturn",
		Description: fmt.Sprintf("Expected to return a value at the end of %s.", name),
	}
}

func buildMissingReturnValueMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingReturnValue",
		Description: fmt.Sprintf("%s expected a return value.", name),
	}
}

func buildUnexpectedReturnValueMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedReturnValue",
		Description: fmt.Sprintf("%s expected no return value.", name),
	}
}

type functionInfo struct {
	node            *ast.Node
	hasReturn       bool
	hasReturnValue  bool
	hasNoReturnValue bool
	isAsync         bool
	functionName    string
}

var ConsistentReturnRule = rule.Rule{
	Name: "consistent-return",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		treatUndefinedAsUnspecified := false
		
		// Parse options with dual-format support
		if options != nil {
			if optionsMap, ok := options.(map[string]interface{}); ok {
				if val, exists := optionsMap["treatUndefinedAsUnspecified"]; exists {
					if boolVal, ok := val.(bool); ok {
						treatUndefinedAsUnspecified = boolVal
					} else if floatVal, ok := val.(float64); ok {
						treatUndefinedAsUnspecified = floatVal != 0
					}
				}
			}
		}

		functionStack := []*functionInfo{}

		getCurrentFunction := func() *functionInfo {
			if len(functionStack) == 0 {
				return nil
			}
			return functionStack[len(functionStack)-1]
		}

		getNodeName := func(nameNode *ast.Node) string {
			if nameNode == nil {
				return ""
			}
			switch nameNode.Kind {
			case ast.KindIdentifier:
				return nameNode.AsIdentifier().Text
			case ast.KindStringLiteral:
				text := nameNode.AsStringLiteral().Text
				if len(text) >= 2 && ((text[0] == '"' && text[len(text)-1] == '"') || (text[0] == '\'' && text[len(text)-1] == '\'')) {
					return text[1 : len(text)-1]
				}
				return text
			case ast.KindNumericLiteral:
				textRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
			default:
				textRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
				return ctx.SourceFile.Text()[textRange.Pos():textRange.End()]
			}
		}

		getFunctionName := func(node *ast.Node) string {
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				funcDecl := node.AsFunctionDeclaration()
				if funcDecl.Name() != nil {
					name := getNodeName(funcDecl.Name())
					if utils.IncludesModifier(funcDecl, ast.KindAsyncKeyword) {
						return fmt.Sprintf("Async function '%s'", name)
					}
					return fmt.Sprintf("Function '%s'", name)
				}
				if utils.IncludesModifier(funcDecl, ast.KindAsyncKeyword) {
					return "Async function"
				}
				return "Function"
			case ast.KindFunctionExpression:
				funcExpr := node.AsFunctionExpression()
				if funcExpr.Name() != nil {
					name := getNodeName(funcExpr.Name())
					if utils.IncludesModifier(funcExpr, ast.KindAsyncKeyword) {
						return fmt.Sprintf("Async function '%s'", name)
					}
					return fmt.Sprintf("Function '%s'", name)
				}
				if utils.IncludesModifier(funcExpr, ast.KindAsyncKeyword) {
					return "Async function"
				}
				return "Function"
			case ast.KindArrowFunction:
				if utils.IncludesModifier(node, ast.KindAsyncKeyword) {
					return "Async arrow function"
				}
				return "Arrow function"
			case ast.KindMethodDeclaration:
				methodDecl := node.AsMethodDeclaration()
				name := getNodeName(methodDecl.Name())
				if utils.IncludesModifier(methodDecl, ast.KindAsyncKeyword) {
					return fmt.Sprintf("Async method '%s'", name)
				}
				return fmt.Sprintf("Method '%s'", name)
			case ast.KindGetAccessor:
				getAccessor := node.AsGetAccessorDeclaration()
				name := getNodeName(getAccessor.Name())
				return fmt.Sprintf("Getter '%s'", name)
			case ast.KindSetAccessor:
				setAccessor := node.AsSetAccessorDeclaration()
				name := getNodeName(setAccessor.Name())
				return fmt.Sprintf("Setter '%s'", name)
			}
			return "Function"
		}

		var isPromiseVoid func(node *ast.Node, t *checker.Type) bool
		isPromiseVoid = func(node *ast.Node, t *checker.Type) bool {
			if !utils.IsThenableType(ctx.TypeChecker, node, t) {
				return false
			}
			
			// Check if it's a type reference (Promise<T>, etc.)
			if !utils.IsObjectType(t) {
				return false
			}
			
			// Get type arguments
			typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, t)
			if len(typeArgs) == 0 {
				return false
			}
			
			awaitedType := typeArgs[0]
			if utils.IsTypeFlagSet(awaitedType, checker.TypeFlagsVoid) {
				return true
			}
			
			// Recursively check nested Promise types
			return isPromiseVoid(node, awaitedType)
		}

		// Check if a Promise type resolves to a union that includes void
		var isPromiseUnionWithVoid func(funcNode *ast.Node, t *checker.Type) bool
		isPromiseUnionWithVoid = func(funcNode *ast.Node, t *checker.Type) bool {
			var checkPromiseUnion func(t *checker.Type) bool
			checkPromiseUnion = func(t *checker.Type) bool {
				if utils.IsThenableType(ctx.TypeChecker, funcNode, t) && utils.IsObjectType(t) {
					typeArgs := checker.Checker_getTypeArguments(ctx.TypeChecker, t)
					if len(typeArgs) > 0 {
						awaitedType := typeArgs[0]
						// If the awaited type is also a Promise, recurse
						if utils.IsThenableType(ctx.TypeChecker, funcNode, awaitedType) {
							return checkPromiseUnion(awaitedType)
						}
						// Check if the awaited type is a union that includes void
						if utils.IsUnionType(awaitedType) {
							for _, unionMember := range awaitedType.Types() {
								if utils.IsTypeFlagSet(unionMember, checker.TypeFlagsVoid) {
									return true
								}
								// Only treat undefined as void in very specific nested Promise cases
								// where we have Promise<Promise<void | undefined>>
								if utils.IsTypeFlagSet(unionMember, checker.TypeFlagsUndefined) {
									// Check if there's also a void in the union
									hasVoid := false
									for _, otherMember := range awaitedType.Types() {
										if utils.IsTypeFlagSet(otherMember, checker.TypeFlagsVoid) {
											hasVoid = true
											break
										}
									}
									if hasVoid {
										return true
									}
								}
							}
						}
					}
				}
				return false
			}
			return checkPromiseUnion(t)
		}

		// Determine the return policy for a function:
		// 0 = strict consistency required
		// 1 = empty returns allowed (void/Promise<void> functions)
		// 2 = mixed returns allowed (union types with void)
		getReturnPolicy := func(funcNode *ast.Node) int {
			t := ctx.TypeChecker.GetTypeAtLocation(funcNode)
			signatures := utils.GetCallSignatures(ctx.TypeChecker, t)
			
			for _, signature := range signatures {
				returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
				
				// Check if function is async
				isAsync := utils.IncludesModifier(funcNode, ast.KindAsyncKeyword)
				
				if isAsync {
					// For async functions, check the Promise resolution deeply
					if isPromiseVoid(funcNode, returnType) {
						return 1 // Empty returns allowed
					}
					// Check if Promise resolves to a union that includes void
					if isPromiseUnionWithVoid(funcNode, returnType) {
						return 2 // Mixed returns allowed
					}
				} else {
					// For sync functions
					if utils.IsTypeFlagSet(returnType, checker.TypeFlagsVoid) {
						return 1 // Empty returns allowed
					}
					// Check if return type is a union that includes void
					if utils.IsUnionType(returnType) {
						for _, unionMember := range returnType.Types() {
							if utils.IsTypeFlagSet(unionMember, checker.TypeFlagsVoid) {
								return 2 // Mixed returns allowed
							}
						}
					}
				}
			}
			
			return 0 // Strict consistency required
		}

		// Check if the return type allows both void and non-void returns (e.g., number | void)
		// isReturnUnionWithVoid := func(funcNode *ast.Node) bool {
		//	t := ctx.TypeChecker.GetTypeAtLocation(funcNode)
		//	signatures := utils.GetCallSignatures(ctx.TypeChecker, t)
		//	
		//	for _, signature := range signatures {
		//		returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
		//		
		//		// For sync functions, check if return type is a union that includes void
		//		if utils.IsUnionType(returnType) {
		//			for _, unionMember := range returnType.Types() {
		//				if utils.IsTypeFlagSet(unionMember, checker.TypeFlagsVoid) {
		//					return true
		//				}
		//			}
		//		}
		//	}
		//	
		//	return false
		// }

		enterFunction := func(node *ast.Node) {
			isAsync := utils.IncludesModifier(node, ast.KindAsyncKeyword)

			functionStack = append(functionStack, &functionInfo{
				node:             node,
				hasReturn:        false,
				hasReturnValue:   false,
				hasNoReturnValue: false,
				isAsync:          isAsync,
				functionName:     getFunctionName(node),
			})
		}

		exitFunction := func(node *ast.Node) {
			if len(functionStack) == 0 {
				return
			}

			funcInfo := getCurrentFunction()
			functionStack = functionStack[:len(functionStack)-1]

			if funcInfo.hasReturn {
				if funcInfo.hasReturnValue && funcInfo.hasNoReturnValue {
					// Inconsistent returns - this will be reported by individual return statements
				}
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindArrowFunction:                            enterFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):       exitFunction,
			ast.KindMethodDeclaration:                        enterFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
			ast.KindGetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
			ast.KindSetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,

			ast.KindReturnStatement: func(node *ast.Node) {
				funcInfo := getCurrentFunction()
				if funcInfo == nil {
					return
				}

				returnStmt := node.AsReturnStatement()
				hasArgument := returnStmt.Expression != nil

				returnPolicy := getReturnPolicy(funcInfo.node)

				// If function allows empty returns (void) and this is an empty return, allow it
				if !hasArgument && returnPolicy >= 1 {
					return
				}

				// If function allows mixed returns (union with void), allow any combination
				if returnPolicy >= 2 {
					return
				}

				// Handle treatUndefinedAsUnspecified option
				if treatUndefinedAsUnspecified && hasArgument {
					returnType := ctx.TypeChecker.GetTypeAtLocation(returnStmt.Expression)
					if utils.IsTypeFlagSet(returnType, checker.TypeFlagsUndefined) {
						hasArgument = false
					}
				}

				funcInfo.hasReturn = true

				if hasArgument {
					if funcInfo.hasNoReturnValue {
						// This return has a value but previous returns didn't
						// Report error on the return expression (the value that shouldn't be there)
						ctx.ReportNode(returnStmt.Expression, buildUnexpectedReturnValueMessage(funcInfo.functionName))
					}
					funcInfo.hasReturnValue = true
				} else {
					if funcInfo.hasReturnValue {
						// This return has no value but previous returns did
						// Report error on the return statement
						ctx.ReportNode(node, buildMissingReturnValueMessage(funcInfo.functionName))
					}
					funcInfo.hasNoReturnValue = true
				}
			},
		}
	},
}