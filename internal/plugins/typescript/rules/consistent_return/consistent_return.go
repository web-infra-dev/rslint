package consistent_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ConsistentReturnOptions struct {
	TreatUndefinedAsUnspecified bool `json:"treatUndefinedAsUnspecified"`
}

// ConsistentReturnRule enforces consistent return statements
var ConsistentReturnRule = rule.CreateRule(rule.Rule{
	Name: "consistent-return",
	Run:  run,
})

// functionInfo tracks information about a function's return statements
type functionInfo struct {
	node                *ast.Node
	hasReturnWithValue  bool
	hasReturnWithoutValue bool
	isVoidOrPromiseVoid bool
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentReturnOptions{
		TreatUndefinedAsUnspecified: false,
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if optsMap, ok := optArray[0].(map[string]interface{}); ok {
				if v, exists := optsMap["treatUndefinedAsUnspecified"].(bool); exists {
					opts.TreatUndefinedAsUnspecified = v
				}
			}
		} else if optsMap, ok := options.(map[string]interface{}); ok {
			if v, exists := optsMap["treatUndefinedAsUnspecified"].(bool); exists {
				opts.TreatUndefinedAsUnspecified = v
			}
		}
	}

	// Stack to track nested functions
	functionStack := make([]*functionInfo, 0)

	// Helper to get current function
	getCurrentFunction := func() *functionInfo {
		if len(functionStack) > 0 {
			return functionStack[len(functionStack)-1]
		}
		return nil
	}

	// Helper to check if type is Promise<void>
	var isPromiseVoid func(typeChecker *checker.Checker, node *ast.Node, typeToCheck *checker.Type) bool
	isPromiseVoid = func(typeChecker *checker.Checker, node *ast.Node, typeToCheck *checker.Type) bool {
		if typeToCheck == nil {
			return false
		}

		// Check if it's a thenable type
		if !utils.IsThenableType(typeChecker, node, typeToCheck) {
			return false
		}

		// Check if it's an object type (Promise<T>)
		if utils.IsObjectType(typeToCheck) {
			typeArgs := checker.Checker_getTypeArguments(typeChecker, typeToCheck)
			if typeArgs != nil && len(typeArgs) > 0 {
				awaitedType := typeArgs[0]
				if utils.IsIntrinsicVoidType(awaitedType) {
					return true
				}
				// Recursively check for nested Promise<void>
				return isPromiseVoid(typeChecker, node, awaitedType)
			}
		}

		return false
	}

	// Helper to check if a function returns void, undefined, or Promise<void>
	isReturnVoidOrPromiseVoid := func(node *ast.Node) bool {
		if ctx.TypeChecker == nil {
			return false
		}

		// Get the type of the function
		funcType := ctx.TypeChecker.GetTypeAtLocation(node)
		if funcType == nil {
			return false
		}

		// Get call signatures
		callSignatures := utils.GetCallSignatures(ctx.TypeChecker, funcType)
		if len(callSignatures) == 0 {
			return false
		}

		for _, sig := range callSignatures {
			returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig)
			if returnType == nil {
				continue
			}

			// Check if return type is void
			if utils.IsIntrinsicVoidType(returnType) {
				return true
			}

			// Check if return type is undefined (for explicit : undefined type annotation)
			if utils.IsTypeFlagSet(returnType, checker.TypeFlagsUndefined) && !utils.IsTypeFlagSet(returnType, checker.TypeFlagsUnion) {
				// Only return type that is purely undefined (not a union)
				return true
			}

			// Check if return type is void | undefined
			if utils.IsTypeFlagSet(returnType, checker.TypeFlagsUnion) {
				// Check if it's a union that contains both void and undefined
				unionParts := utils.UnionTypeParts(returnType)
				if unionParts != nil {
					hasVoid := false
					hasUndefined := false
					for _, t := range unionParts {
						if t == nil {
							continue
						}
						if utils.IsIntrinsicVoidType(t) {
							hasVoid = true
						}
						if utils.IsTypeFlagSet(t, checker.TypeFlagsUndefined) {
							hasUndefined = true
						}
					}
					if hasVoid && hasUndefined {
						return true
					}
				}
			}

			// Check if it's an async function returning Promise<void>
			if node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindMethodDeclaration {
				isAsync := ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)

				if isAsync && isPromiseVoid(ctx.TypeChecker, node, returnType) {
					return true
				}
			}
		}

		return false
	}

	// Helper to check if return type is undefined
	isUndefinedType := func(node *ast.Node) bool {
		if ctx.TypeChecker == nil || node == nil {
			return false
		}

		typeAtLocation := ctx.TypeChecker.GetTypeAtLocation(node)
		if typeAtLocation == nil {
			return false
		}

		return utils.IsTypeFlagSet(typeAtLocation, checker.TypeFlagsUndefined)
	}

	// Helper to get function name for better error messages
	getFunctionName := func(node *ast.Node) string {
		switch node.Kind {
		case ast.KindFunctionDeclaration:
			fn := node.AsFunctionDeclaration()
			if fn != nil && fn.Name() != nil && fn.Name().Kind == ast.KindIdentifier {
				ident := fn.Name().AsIdentifier()
				if ident != nil {
					return "function '" + ident.Text + "'"
				}
			}
			return "function"
		case ast.KindConstructor:
			return "constructor"
		case ast.KindMethodDeclaration:
			method := node.AsMethodDeclaration()
			if method != nil && method.Name() != nil {
				name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
				return "method '" + name + "'"
			}
			return "method"
		case ast.KindGetAccessor:
			accessor := node.AsGetAccessorDeclaration()
			if accessor != nil && accessor.Name() != nil {
				name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
				return "getter '" + name + "'"
			}
			return "getter"
		case ast.KindFunctionExpression:
			parent := node.Parent
			if parent != nil {
				switch parent.Kind {
				case ast.KindMethodDeclaration:
					method := parent.AsMethodDeclaration()
					if method != nil && method.Name() != nil {
						name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
						return "method '" + name + "'"
					}
				case ast.KindPropertyDeclaration:
					prop := parent.AsPropertyDeclaration()
					if prop != nil && prop.Name() != nil {
						name, _ := utils.GetNameFromMember(ctx.SourceFile, prop.Name())
						if name != "" {
							return "function '" + name + "'"
						}
					}
				case ast.KindPropertyAssignment:
					prop := parent.AsPropertyAssignment()
					if prop != nil && prop.Name() != nil {
						name, _ := utils.GetNameFromMember(ctx.SourceFile, prop.Name())
						if name != "" {
							return "function '" + name + "'"
						}
					}
				case ast.KindVariableDeclaration:
					decl := parent.AsVariableDeclaration()
					if decl != nil && decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
						ident := decl.Name().AsIdentifier()
						if ident != nil {
							return "function '" + ident.Text + "'"
						}
					}
				}
			}
			return "function"
		case ast.KindArrowFunction:
			parent := node.Parent
			if parent != nil && parent.Kind == ast.KindVariableDeclaration {
				decl := parent.AsVariableDeclaration()
				if decl != nil && decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
					ident := decl.Name().AsIdentifier()
					if ident != nil {
						return "arrow function '" + ident.Text + "'"
					}
				}
			}
			return "arrow function"
		default:
			return "function"
		}
	}

	enterFunction := func(node *ast.Node) {
		info := &functionInfo{
			node:                  node,
			hasReturnWithValue:    false,
			hasReturnWithoutValue: false,
			isVoidOrPromiseVoid:   isReturnVoidOrPromiseVoid(node),
		}
		functionStack = append(functionStack, info)
	}

	exitFunction := func(node *ast.Node) {
		if len(functionStack) == 0 {
			return
		}

		info := functionStack[len(functionStack)-1]
		functionStack = functionStack[:len(functionStack)-1]

		// Check for inconsistent returns
		if info.hasReturnWithValue && info.hasReturnWithoutValue {
			// Report error on the function
			funcName := getFunctionName(node)

			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingReturnValue",
				Description: funcName + " has inconsistent return statements. Either all return statements should return a value, or none should.",
			})
		}
	}

	return rule.RuleListeners{
		ast.KindFunctionDeclaration:                      enterFunction,
		rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,

		ast.KindFunctionExpression:                      enterFunction,
		rule.ListenerOnExit(ast.KindFunctionExpression): exitFunction,

		ast.KindArrowFunction:                      enterFunction,
		rule.ListenerOnExit(ast.KindArrowFunction): exitFunction,

		ast.KindMethodDeclaration:                      enterFunction,
		rule.ListenerOnExit(ast.KindMethodDeclaration): exitFunction,

		ast.KindGetAccessor:                      enterFunction,
		rule.ListenerOnExit(ast.KindGetAccessor): exitFunction,

		ast.KindReturnStatement: func(node *ast.Node) {
			funcInfo := getCurrentFunction()
			if funcInfo == nil {
				return
			}

			returnExpr := node.Expression()

			// If no return value and function returns void/Promise<void>, it's ok
			if returnExpr == nil && funcInfo.isVoidOrPromiseVoid {
				return
			}

			// Check if we're treating undefined as unspecified
			if opts.TreatUndefinedAsUnspecified && returnExpr != nil {
				if isUndefinedType(returnExpr) {
					// Treat this as a return without value
					funcInfo.hasReturnWithoutValue = true
					return
				}
			}

			// Track whether this return has a value
			if returnExpr != nil {
				funcInfo.hasReturnWithValue = true
			} else {
				funcInfo.hasReturnWithoutValue = true
			}
		},
	}
}
