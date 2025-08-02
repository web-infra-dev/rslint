package init_declarations

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type InitDeclarationsOptions struct {
	Mode              string `json:"mode"`
	IgnoreForLoopInit bool   `json:"ignoreForLoopInit"`
}

var InitDeclarationsRule = rule.Rule{
	Name: "init-declarations",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Default options - default to "always" mode like ESLint
		opts := InitDeclarationsOptions{
			Mode:              "always",
			IgnoreForLoopInit: false,
		}

		// Parse options - handle both array format [mode, options] and direct object format
		if options != nil {
			var optsMap map[string]interface{}
			var mode string

			// Handle array format: ["always", {"ignoreForLoopInit": true}] or ["never"]
			if optsArray, ok := options.([]interface{}); ok {
				if len(optsArray) > 0 {
					if modeStr, ok := optsArray[0].(string); ok {
						mode = modeStr
					}
				}
				if len(optsArray) > 1 {
					if optObj, ok := optsArray[1].(map[string]interface{}); ok {
						optsMap = optObj
					}
				}
			} else if optObj, ok := options.(map[string]interface{}); ok {
				// Handle direct object format: {"ignoreForLoopInit": true}
				optsMap = optObj
			} else if modeStr, ok := options.(string); ok {
				// Handle string format: "always" or "never"
				mode = modeStr
			}

			// Apply mode if provided
			if mode != "" {
				opts.Mode = mode
			}

			// Apply object options if provided
			if optsMap != nil {
				if ignoreForLoopInit, ok := optsMap["ignoreForLoopInit"].(bool); ok {
					opts.IgnoreForLoopInit = ignoreForLoopInit
				}
			}
		}

		// Helper function to check if a variable declaration is in a declare namespace
		isAncestorNamespaceDeclared := func(node *ast.Node) bool {
			ancestor := node.Parent
			for ancestor != nil {
				if ancestor.Kind == ast.KindModuleDeclaration {
					// Check if it has declare modifier
					if ast.HasSyntacticModifier(ancestor, ast.ModifierFlagsAmbient) {
						return true
					}
				}
				ancestor = ancestor.Parent
			}
			return false
		}

		// Helper function to check if a variable declaration is in a for loop init
		isInForLoopInit := func(node *ast.Node) bool {
			// node could be either a VariableDeclarationList or its parent node
			varDeclList := node

			// If node is not a VariableDeclarationList, check if it contains one
			if node.Kind != ast.KindVariableDeclarationList {
				// For VariableStatement, the declaration list is a child
				if node.Kind == ast.KindVariableStatement {
					varStmt := node.AsVariableStatement()
					if varStmt.DeclarationList != nil {
						varDeclList = varStmt.DeclarationList
					}
				} else {
					return false
				}
			}

			// Check if the parent is a for loop and this is the initializer
			parent := varDeclList.Parent
			if parent != nil {
				switch parent.Kind {
				case ast.KindForStatement:
					forStmt := parent.AsForStatement()
					return forStmt.Initializer == varDeclList
				case ast.KindForInStatement:
					forInStmt := parent.AsForInOrOfStatement()
					return forInStmt.Initializer == varDeclList
				case ast.KindForOfStatement:
					forOfStmt := parent.AsForInOrOfStatement()
					return forOfStmt.Initializer == varDeclList
				}
			}

			return false
		}

		// Helper function to check if we're in a for-in or for-of loop (which are valid without initializers)
		isInForInOrOfLoop := func(parentNode *ast.Node) bool {
			if parentNode == nil {
				return false
			}

			switch parentNode.Kind {
			case ast.KindForInStatement, ast.KindForOfStatement:
				return true
			}

			// Check if the parent of parentNode is for-in/for-of (for VariableDeclarationList case)
			if parentNode.Parent != nil {
				switch parentNode.Parent.Kind {
				case ast.KindForInStatement, ast.KindForOfStatement:
					return true
				}
			}
			return false
		}

		// Helper function to get report location for identifier only
		getReportLoc := func(node *ast.Node) core.TextRange {
			// Get identifier name for proper range
			declarator := node.AsVariableDeclaration()
			if declarator.Name().Kind == ast.KindIdentifier {
				identifier := declarator.Name()
				// Report just the identifier part
				return utils.TrimNodeTextRange(ctx.SourceFile, identifier)
			}
			// For non-identifier patterns, use default range
			return utils.TrimNodeTextRange(ctx.SourceFile, node)
		}

		// Shared function to handle variable declaration lists
		handleVarDeclList := func(varDeclList *ast.VariableDeclarationList, parentNode *ast.Node) {
			// Skip if ignoreForLoopInit is true and this is in a for loop
			// This applies to both "always" and "never" modes
			isForLoopInit := isInForLoopInit(parentNode)

			// Debug info
			// if varDeclList.Parent != nil && varDeclList.Parent.Kind == ast.KindForStatement {
			//	fmt.Printf("DEBUG: In for loop, ignoreForLoopInit=%v, isForLoopInit=%v, mode=%s\n", opts.IgnoreForLoopInit, isForLoopInit, opts.Mode)
			// }

			if opts.IgnoreForLoopInit && isForLoopInit {
				return
			}

			// Skip ambient declarations (declare keyword or in declare namespace)
			if ast.HasSyntacticModifier(parentNode, ast.ModifierFlagsAmbient) {
				return
			}
			if isAncestorNamespaceDeclared(parentNode) {
				return
			}

			isConst := varDeclList.Flags&ast.NodeFlagsConst != 0

			// Check each variable declarator
			for _, decl := range varDeclList.Declarations.Nodes {
				declarator := decl.AsVariableDeclaration()
				hasInit := declarator.Initializer != nil

				// Get identifier name for error message
				var idName string
				if declarator.Name().Kind == ast.KindIdentifier {
					idName = declarator.Name().AsIdentifier().Text
				} else {
					// For destructuring patterns, we skip for now
					// The base ESLint rule only reports on identifiers
					continue
				}

				if opts.Mode == "always" && !hasInit {
					// const declarations are allowed without initialization in ambient contexts
					// (declare statements, declare namespaces, .d.ts files)
					if isConst {
						// Check if we're in an ambient context
						if ast.HasSyntacticModifier(parentNode, ast.ModifierFlagsAmbient) || isAncestorNamespaceDeclared(parentNode) {
							continue
						}
						// In non-ambient contexts, const without initializer should be reported
					}

					// For-in and for-of loop variables don't need initializers in "always" mode
					// But only if they are the actual loop variable, not variables in other statements
					if isInForInOrOfLoop(parentNode) {
						// Check if this variable is actually the loop variable
						parent := parentNode.Parent
						if parent != nil {
							switch parent.Kind {
							case ast.KindForInStatement:
								forInStmt := parent.AsForInOrOfStatement()
								if forInStmt.Initializer == parentNode {
									continue
								}
							case ast.KindForOfStatement:
								forOfStmt := parent.AsForInOrOfStatement()
								if forOfStmt.Initializer == parentNode {
									continue
								}
							}
						}
					}

					ctx.ReportRange(getReportLoc(decl), rule.RuleMessage{
						Id:          "initialized",
						Description: fmt.Sprintf("Variable '%s' should be initialized at declaration.", idName),
					})
				} else if opts.Mode == "never" {
					// const declarations MUST be initialized by language spec
					// so we don't report them in "never" mode
					if isConst {
						continue
					}

					// In "never" mode, report variables with explicit initializers
					shouldReport := hasInit

					// Also report for-in/for-of loop variables (they are effectively initialized by the loop)
					// BUT only if they are not for-in/for-of variables themselves
					if !hasInit && isInForInOrOfLoop(parentNode) {
						// Check if this variable is actually the loop variable
						parent := parentNode.Parent
						if parent != nil {
							switch parent.Kind {
							case ast.KindForInStatement:
								forInStmt := parent.AsForInOrOfStatement()
								if forInStmt.Initializer == parentNode {
									shouldReport = true
								}
							case ast.KindForOfStatement:
								forOfStmt := parent.AsForInOrOfStatement()
								if forOfStmt.Initializer == parentNode {
									shouldReport = true
								}
							}
						}
					}

					if shouldReport {
						// Report the entire declarator including initialization
						ctx.ReportNode(decl, rule.RuleMessage{
							Id:          "notInitialized",
							Description: fmt.Sprintf("Variable '%s' should not be initialized.", idName),
						})
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindVariableStatement: func(node *ast.Node) {
				varStmt := node.AsVariableStatement()
				if varStmt.DeclarationList == nil {
					return
				}

				varDeclList := varStmt.DeclarationList.AsVariableDeclarationList()
				handleVarDeclList(varDeclList, node)
			},

			// Handle variable declarations in for loops that are not wrapped in VariableStatement
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				// Only process if this is not already handled by VariableStatement
				// Check if parent is a for loop (not a VariableStatement)
				if node.Parent != nil {
					switch node.Parent.Kind {
					case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
						varDeclList := node.AsVariableDeclarationList()
						// Pass the VariableDeclarationList node, not its parent
						handleVarDeclList(varDeclList, node)
					}
				}
			},
		}
	},
}
