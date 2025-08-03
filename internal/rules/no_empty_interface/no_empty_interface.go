package no_empty_interface

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type NoEmptyInterfaceOptions struct {
	AllowSingleExtends bool `json:"allowSingleExtends"`
}

var NoEmptyInterfaceRule = rule.Rule{
	Name: "no-empty-interface",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoEmptyInterfaceOptions{
			AllowSingleExtends: false,
		}
		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowSingleExtends, ok := optsMap["allowSingleExtends"].(bool); ok {
					opts.AllowSingleExtends = allowSingleExtends
				}
			}
		}

		return rule.RuleListeners{
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				interfaceDecl := node.AsInterfaceDeclaration()

				// Check if interface has members
				if interfaceDecl.Members != nil && len(interfaceDecl.Members.Nodes) > 0 {
					// interface contains members --> Nothing to report
					return
				}

				// Count extended interfaces
				extendCount := 0
				var extendClause *ast.HeritageClause
				if interfaceDecl.HeritageClauses != nil {
					for _, clause := range interfaceDecl.HeritageClauses.Nodes {
						heritageClause := clause.AsHeritageClause()
						if heritageClause.Token == ast.KindExtendsKeyword {
							extendClause = heritageClause
							extendCount = len(heritageClause.Types.Nodes)
							break
						}
					}
				}

				// Report empty interface with no extends
				if extendCount == 0 {
					ctx.ReportNode(interfaceDecl.Name(), rule.RuleMessage{
						Description: "An empty interface is equivalent to `{}`.",
					})
					return
				}

				// Report empty interface with single extend if not allowed
				if extendCount == 1 && !opts.AllowSingleExtends {
					// Check for merged class declaration
					mergedWithClassDeclaration := false
					if ctx.TypeChecker != nil {
						symbol := ctx.TypeChecker.GetSymbolAtLocation(interfaceDecl.Name())
						if symbol != nil {
							// Check if this symbol has a class declaration
							for _, decl := range symbol.Declarations {
								if decl.Kind == ast.KindClassDeclaration {
									mergedWithClassDeclaration = true
									break
								}
							}
						}
					}

					// Check if in ambient declaration (.d.ts file)
					isInAmbientDeclaration := false
					if strings.HasSuffix(ctx.SourceFile.FileName(), ".d.ts") {
						// Check if we're inside a declared module
						parent := node.Parent
						for parent != nil {
							if parent.Kind == ast.KindModuleDeclaration {
								moduleDecl := parent.AsModuleDeclaration()
								modifiers := moduleDecl.Modifiers()
								if modifiers != nil {
									for _, modifier := range modifiers.Nodes {
										if modifier.Kind == ast.KindDeclareKeyword {
											isInAmbientDeclaration = true
											break
										}
									}
								}
							}
							if isInAmbientDeclaration {
								break
							}
							parent = parent.Parent
						}
					}

					// Build the replacement text
					// Extract interface name
					nameRange := utils.TrimNodeTextRange(ctx.SourceFile, interfaceDecl.Name())
					nameText := string(ctx.SourceFile.Text()[nameRange.Pos():nameRange.End()])

					// Extract type parameters if present
					var typeParamsText string
					if interfaceDecl.TypeParameters != nil && len(interfaceDecl.TypeParameters.Nodes) > 0 {
						// Create text range for the entire type parameters list
						firstParam := interfaceDecl.TypeParameters.Nodes[0]
						lastParam := interfaceDecl.TypeParameters.Nodes[len(interfaceDecl.TypeParameters.Nodes)-1]
						firstRange := utils.TrimNodeTextRange(ctx.SourceFile, firstParam)
						lastRange := utils.TrimNodeTextRange(ctx.SourceFile, lastParam)
						typeParamsRange := firstRange.WithEnd(lastRange.End())
						// Include the angle brackets
						typeParamsRange = typeParamsRange.WithPos(typeParamsRange.Pos() - 1).WithEnd(typeParamsRange.End() + 1)
						typeParamsText = string(ctx.SourceFile.Text()[typeParamsRange.Pos():typeParamsRange.End()])
					}

					extendedTypeRange := utils.TrimNodeTextRange(ctx.SourceFile, extendClause.Types.Nodes[0])
					extendedTypeText := string(ctx.SourceFile.Text()[extendedTypeRange.Pos():extendedTypeRange.End()])

					replacement := fmt.Sprintf("type %s%s = %s", nameText, typeParamsText, extendedTypeText)

					message := rule.RuleMessage{
						Description: "An interface declaring no members is equivalent to its supertype.",
					}

					// Determine the appropriate reporting method
					if isInAmbientDeclaration || mergedWithClassDeclaration {
						// Just report without fix or suggestion for ambient declarations or merged class declarations
						ctx.ReportNode(interfaceDecl.Name(), message)
					} else {
						// Use auto-fix for non-ambient, non-merged cases
						ctx.ReportNodeWithFixes(interfaceDecl.Name(), message,
							rule.RuleFixReplace(ctx.SourceFile, node, replacement))
					}
				}
			},
		}
	},
}
