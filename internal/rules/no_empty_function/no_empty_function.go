package no_empty_function

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

type NoEmptyFunctionOptions struct {
	Allow []string `json:"allow"`
}

var NoEmptyFunctionRule = rule.Rule{
	Name: "no-empty-function",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoEmptyFunctionOptions{
			Allow: []string{},
		}
		if options != nil {
			if optsSlice, ok := options.([]interface{}); ok && len(optsSlice) > 0 {
				if optsMap, ok := optsSlice[0].(map[string]interface{}); ok {
					if allow, ok := optsMap["allow"].([]interface{}); ok {
						for _, a := range allow {
							if str, ok := a.(string); ok {
								opts.Allow = append(opts.Allow, str)
							}
						}
					}
				}
			}
		}

		// Helper to check if a type is allowed
		isAllowed := func(allowType string) bool {
			for _, a := range opts.Allow {
				if a == allowType {
					return true
				}
			}
			return false
		}

		// Check if the function body is empty
		isBodyEmpty := func(node *ast.Node) bool {
			if node.Kind == ast.KindFunctionDeclaration {
				fn := node.AsFunctionDeclaration()
				return fn.Body != nil && len(fn.Body.Statements()) == 0
			} else if node.Kind == ast.KindFunctionExpression {
				fn := node.AsFunctionExpression()
				return fn.Body != nil && len(fn.Body.Statements()) == 0
			} else if node.Kind == ast.KindArrowFunction {
				fn := node.AsArrowFunction()
				// Arrow functions can have expression bodies (no block)
				if fn.Body == nil {
					return false
				}
				if fn.Body.Kind != ast.KindBlock {
					return false // Expression body, not empty
				}
				block := fn.Body.AsBlock()
				return len(block.Statements.Nodes) == 0
			}
			return false
		}

		// Check if function has parameter properties (TypeScript constructor feature)
		hasParameterProperties := func(node *ast.Node) bool {
			var params []*ast.Node
			if node.Kind == ast.KindFunctionDeclaration {
				if node.AsFunctionDeclaration().Parameters != nil {
					params = node.AsFunctionDeclaration().Parameters.Nodes
				}
			} else if node.Kind == ast.KindFunctionExpression {
				if node.AsFunctionExpression().Parameters != nil {
					params = node.AsFunctionExpression().Parameters.Nodes
				}
			} else if node.Kind == ast.KindArrowFunction {
				if node.AsArrowFunction().Parameters != nil {
					params = node.AsArrowFunction().Parameters.Nodes
				}
			}

			for _, param := range params {
				if param.Kind == ast.KindParameter {
					// Check if parameter has modifiers (public/private/protected/readonly)
					if ast.GetCombinedModifierFlags(param)&(ast.ModifierFlagsPublic|ast.ModifierFlagsPrivate|ast.ModifierFlagsProtected|ast.ModifierFlagsReadonly) != 0 {
						return true
					}
				}
			}
			return false
		}

		// Get the function name for error message
		getFunctionName := func(node *ast.Node) string {
			if node.Kind == ast.KindFunctionDeclaration {
				fn := node.AsFunctionDeclaration()
				if fn.Name() != nil && fn.Name().Kind == ast.KindIdentifier {
					return "function '" + fn.Name().AsIdentifier().Text + "'"
				}
				return "function"
			} else if node.Kind == ast.KindFunctionExpression {
				parent := node.Parent
				if parent != nil {
					if parent.Kind == ast.KindMethodDeclaration {
						method := parent.AsMethodDeclaration()
						if method.Name() != nil {
							name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
							if method.Kind == ast.KindConstructor {
								return "constructor"
							}
							if method.Kind == ast.KindGetAccessor {
								return "getter '" + name + "'"
							}
							if method.Kind == ast.KindSetAccessor {
								return "setter '" + name + "'"
							}
							return "method '" + name + "'"
						}
					} else if parent.Kind == ast.KindPropertyDeclaration || parent.Kind == ast.KindPropertyAssignment {
						// Check for variable declaration or property assignment
						var name string
						if parent.Kind == ast.KindPropertyDeclaration {
							prop := parent.AsPropertyDeclaration()
							if prop.Name() != nil {
								name, _ = utils.GetNameFromMember(ctx.SourceFile, prop.Name())
							}
						} else if parent.Kind == ast.KindPropertyAssignment {
							prop := parent.AsPropertyAssignment()
							if prop.Name() != nil {
								name, _ = utils.GetNameFromMember(ctx.SourceFile, prop.Name())
							}
						}
						if name != "" {
							return "function '" + name + "'"
						}
					} else if parent.Kind == ast.KindVariableDeclaration {
						decl := parent.AsVariableDeclaration()
						if decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
							return "function '" + decl.Name().AsIdentifier().Text + "'"
						}
					}
				}
				return "function"
			} else if node.Kind == ast.KindArrowFunction {
				parent := node.Parent
				if parent != nil && parent.Kind == ast.KindVariableDeclaration {
					decl := parent.AsVariableDeclaration()
					if decl.Name() != nil && decl.Name().Kind == ast.KindIdentifier {
						return "arrow function '" + decl.Name().AsIdentifier().Text + "'"
					}
				}
				return "arrow function"
			}
			return "function"
		}

		// Main check function for all function types
		checkFunction := func(node *ast.Node) {
			if !isBodyEmpty(node) {
				return
			}

			parent := node.Parent
			isAsync := false
			isGenerator := false

			// Detect async and generator functions
			if node.Kind == ast.KindFunctionDeclaration {
				fn := node.AsFunctionDeclaration()
				isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
				isGenerator = fn.AsteriskToken != nil
			} else if node.Kind == ast.KindFunctionExpression {
				fn := node.AsFunctionExpression()
				isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
				isGenerator = fn.AsteriskToken != nil
			} else if node.Kind == ast.KindArrowFunction {
				isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
			}

			// Check for various allowed types
			if parent != nil && parent.Kind == ast.KindMethodDeclaration {
				method := parent.AsMethodDeclaration()

				// Constructor checks
				if method.Kind == ast.KindConstructor {
					if isAllowed("constructors") {
						return
					}

					// Check accessibility modifiers
					hasPrivate := ast.HasSyntacticModifier(parent, ast.ModifierFlagsPrivate)
					hasProtected := ast.HasSyntacticModifier(parent, ast.ModifierFlagsProtected)

					if hasPrivate && isAllowed("private-constructors") {
						return
					}
					if hasProtected && isAllowed("protected-constructors") {
						return
					}

					// Constructors with parameter properties are allowed
					if hasParameterProperties(node) {
						return
					}
				}

				// Getter/Setter checks
				if method.Kind == ast.KindGetAccessor && isAllowed("getters") {
					return
				}
				if method.Kind == ast.KindSetAccessor && isAllowed("setters") {
					return
				}

				// Decorated function check
				if ast.GetCombinedModifierFlags(parent)&ast.ModifierFlagsDecorator != 0 && isAllowed("decoratedFunctions") {
					return
				}

				// Override method check
				if ast.HasSyntacticModifier(parent, ast.ModifierFlagsOverride) && isAllowed("overrideMethods") {
					return
				}

				// Regular method checks
				if method.Kind == ast.KindMethodSignature || method.Kind == ast.KindMethodDeclaration {
					if isAsync && isAllowed("asyncMethods") {
						return
					}
					if isGenerator && isAllowed("generatorMethods") {
						return
					}
					if isAllowed("methods") {
						return
					}
				}
			} else {
				// Not in a method, check function types
				if node.Kind == ast.KindArrowFunction && isAllowed("arrowFunctions") {
					return
				}
				if isAsync && isAllowed("asyncFunctions") {
					return
				}
				if isGenerator && isAllowed("generatorFunctions") {
					return
				}
				if isAllowed("functions") {
					return
				}
			}

			// Report the error
			funcName := getFunctionName(node)
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "unexpected",
				Description: "Unexpected empty " + funcName + ".",
			})
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkFunction,
			ast.KindFunctionExpression:  checkFunction,
			ast.KindArrowFunction:       checkFunction,
		}
	},
}