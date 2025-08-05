package no_empty_function

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
			var optsMap map[string]interface{}
			if optsArray, ok := options.([]interface{}); ok && len(optsArray) > 0 {
				if opts, ok := optsArray[0].(map[string]interface{}); ok {
					optsMap = opts
				}
			} else if opts, ok := options.(map[string]interface{}); ok {
				optsMap = opts
			}

			if optsMap != nil {
				if allow, ok := optsMap["allow"].([]interface{}); ok {
					for _, a := range allow {
						if str, ok := a.(string); ok {
							opts.Allow = append(opts.Allow, str)
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
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				fn := node.AsFunctionDeclaration()
				if fn == nil || fn.Body == nil {
					return false
				}
				return len(fn.Body.Statements()) == 0
			case ast.KindFunctionExpression:
				fn := node.AsFunctionExpression()
				if fn == nil || fn.Body == nil {
					return false
				}
				return len(fn.Body.Statements()) == 0
			case ast.KindArrowFunction:
				fn := node.AsArrowFunction()
				if fn == nil || fn.Body == nil {
					return false
				}
				// Arrow functions can have expression bodies (no block)
				if fn.Body.Kind != ast.KindBlock {
					return false // Expression body, not empty
				}
				block := fn.Body.AsBlock()
				if block == nil || block.Statements == nil {
					return false
				}
				return len(block.Statements.Nodes) == 0
			case ast.KindConstructor:
				constructor := node.AsConstructorDeclaration()
				if constructor == nil || constructor.Body == nil {
					return false
				}
				return len(constructor.Body.Statements()) == 0
			case ast.KindMethodDeclaration:
				method := node.AsMethodDeclaration()
				if method == nil || method.Body == nil {
					return false
				}
				return len(method.Body.Statements()) == 0
			case ast.KindGetAccessor:
				accessor := node.AsGetAccessorDeclaration()
				if accessor == nil || accessor.Body == nil {
					return false
				}
				return len(accessor.Body.Statements()) == 0
			case ast.KindSetAccessor:
				accessor := node.AsSetAccessorDeclaration()
				if accessor == nil || accessor.Body == nil {
					return false
				}
				return len(accessor.Body.Statements()) == 0
			default:
				return false
			}
		}

		// Check if function has parameter properties (TypeScript constructor feature)
		hasParameterProperties := func(node *ast.Node) bool {
			var params []*ast.Node
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				fn := node.AsFunctionDeclaration()
				if fn != nil && fn.Parameters != nil {
					params = fn.Parameters.Nodes
				}
			case ast.KindFunctionExpression:
				fn := node.AsFunctionExpression()
				if fn != nil && fn.Parameters != nil {
					params = fn.Parameters.Nodes
				}
			case ast.KindArrowFunction:
				fn := node.AsArrowFunction()
				if fn != nil && fn.Parameters != nil {
					params = fn.Parameters.Nodes
				}
			case ast.KindConstructor:
				constructor := node.AsConstructorDeclaration()
				if constructor != nil && constructor.Parameters != nil {
					params = constructor.Parameters.Nodes
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

		// Get the opening brace position of a function body
		getOpenBracePosition := func(node *ast.Node) (core.TextRange, bool) {
			var body *ast.Node
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				fn := node.AsFunctionDeclaration()
				if fn != nil {
					body = fn.Body
				}
			case ast.KindFunctionExpression:
				fn := node.AsFunctionExpression()
				if fn != nil {
					body = fn.Body
				}
			case ast.KindArrowFunction:
				fn := node.AsArrowFunction()
				if fn != nil && fn.Body != nil && fn.Body.Kind == ast.KindBlock {
					body = fn.Body
				}
			case ast.KindConstructor:
				constructor := node.AsConstructorDeclaration()
				if constructor != nil {
					body = constructor.Body
				}
			case ast.KindMethodDeclaration:
				method := node.AsMethodDeclaration()
				if method != nil {
					body = method.Body
				}
			case ast.KindGetAccessor:
				accessor := node.AsGetAccessorDeclaration()
				if accessor != nil {
					body = accessor.Body
				}
			case ast.KindSetAccessor:
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil {
					body = accessor.Body
				}
			}

			if body == nil {
				return core.TextRange{}, false
			}

			// Find the opening brace by searching for '{' character from node start to body end
			sourceText := ctx.SourceFile.Text()
			nodeStart := node.Pos()
			bodyStart := body.Pos()

			// Search for the opening brace between node start and body start
			for i := nodeStart; i <= bodyStart && i < len(sourceText); i++ {
				if sourceText[i] == '{' {
					return core.TextRange{}.WithPos(i).WithEnd(i + 1), true
				}
			}

			// Fallback: use the body's start position
			return core.TextRange{}.WithPos(bodyStart).WithEnd(bodyStart + 1), true
		}

		// Get the function name for error message
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
			case ast.KindSetAccessor:
				accessor := node.AsSetAccessorDeclaration()
				if accessor != nil && accessor.Name() != nil {
					name, _ := utils.GetNameFromMember(ctx.SourceFile, accessor.Name())
					return "setter '" + name + "'"
				}
				return "setter"
			case ast.KindFunctionExpression:
				parent := node.Parent
				if parent != nil {
					switch parent.Kind {
					case ast.KindMethodDeclaration:
						method := parent.AsMethodDeclaration()
						if method != nil && method.Name() != nil {
							name, _ := utils.GetNameFromMember(ctx.SourceFile, method.Name())
							if method.Kind == ast.KindGetAccessor {
								return "getter '" + name + "'"
							}
							if method.Kind == ast.KindSetAccessor {
								return "setter '" + name + "'"
							}
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
			} else if node.Kind == ast.KindArrowFunction {
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
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				fn := node.AsFunctionDeclaration()
				if fn != nil {
					isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
					isGenerator = fn.AsteriskToken != nil
				}
			case ast.KindFunctionExpression:
				fn := node.AsFunctionExpression()
				if fn != nil {
					isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
					isGenerator = fn.AsteriskToken != nil
				}
			case ast.KindArrowFunction:
				isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
			case ast.KindMethodDeclaration:
				method := node.AsMethodDeclaration()
				if method != nil {
					isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
					isGenerator = method.AsteriskToken != nil
				}
			case ast.KindGetAccessor:
				isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
			case ast.KindSetAccessor:
				isAsync = ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
			case ast.KindConstructor:
				// Check accessibility modifiers for constructors
				hasPrivate := ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate)
				hasProtected := ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected)

				if isAllowed("constructors") {
					return
				}
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

			// Check for arrow functions first (before parent checks)
			if node.Kind == ast.KindArrowFunction && isAllowed("arrowFunctions") {
				return
			}

			// Check for async/generator functions early
			if isAsync && isAllowed("asyncFunctions") {
				return
			}
			if isGenerator && isAllowed("generatorFunctions") {
				return
			}
			if node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindFunctionExpression {
				if isAllowed("functions") {
					return
				}
			}

			// Check for method declarations directly
			if node.Kind == ast.KindMethodDeclaration {
				// Decorated function check
				if ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsDecorator != 0 && isAllowed("decoratedFunctions") {
					return
				}

				// Override method check
				if ast.HasSyntacticModifier(node, ast.ModifierFlagsOverride) && isAllowed("overrideMethods") {
					return
				}

				// Regular method checks
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

			// Check for accessor declarations directly
			if node.Kind == ast.KindGetAccessor && isAllowed("getters") {
				return
			}
			if node.Kind == ast.KindSetAccessor && isAllowed("setters") {
				return
			}

			// Check for various allowed types (parent-based logic for function expressions)
			if parent != nil && parent.Kind == ast.KindMethodDeclaration {
				method := parent.AsMethodDeclaration()
				if method != nil {

					// Constructor checks - not needed here since we handle KindConstructor directly above

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

			// Report the error at the opening brace position
			funcName := getFunctionName(node)
			if braceRange, found := getOpenBracePosition(node); found {
				ctx.ReportRange(braceRange, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Unexpected empty " + funcName + ".",
				})
			} else {
				// Fallback to reporting on the entire node
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Unexpected empty " + funcName + ".",
				})
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkFunction,
			ast.KindFunctionExpression:  checkFunction,
			ast.KindArrowFunction:       checkFunction,
			ast.KindConstructor:         checkFunction,
			ast.KindMethodDeclaration:   checkFunction,
			ast.KindGetAccessor:         checkFunction,
			ast.KindSetAccessor:         checkFunction,
		}
	},
}
