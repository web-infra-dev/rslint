package no_extraneous_class

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type NoExtraneousClassOptions struct {
	AllowConstructorOnly bool `json:"allowConstructorOnly"`
	AllowEmpty           bool `json:"allowEmpty"`
	AllowStaticOnly      bool `json:"allowStaticOnly"`
	AllowWithDecorator   bool `json:"allowWithDecorator"`
}

var NoExtraneousClassRule = rule.CreateRule(rule.Rule{
	Name: "no-extraneous-class",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoExtraneousClassOptions{
			AllowConstructorOnly: false,
			AllowEmpty:           false,
			AllowStaticOnly:      false,
			AllowWithDecorator:   false,
		}

		// Parse options with dual-format support
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
				if allowConstructorOnly, ok := optsMap["allowConstructorOnly"].(bool); ok {
					opts.AllowConstructorOnly = allowConstructorOnly
				}
				if allowEmpty, ok := optsMap["allowEmpty"].(bool); ok {
					opts.AllowEmpty = allowEmpty
				}
				if allowStaticOnly, ok := optsMap["allowStaticOnly"].(bool); ok {
					opts.AllowStaticOnly = allowStaticOnly
				}
				if allowWithDecorator, ok := optsMap["allowWithDecorator"].(bool); ok {
					opts.AllowWithDecorator = allowWithDecorator
				}
			}
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				classDecl := node.AsClassDeclaration()
				if classDecl == nil {
					return
				}

				// Get the node to report on (prefer name, fallback to class node)
				reportNode := classDecl.Name()
				if reportNode == nil {
					reportNode = node
				}

				// Skip classes that extend another class
				if classDecl.HeritageClauses != nil {
					for _, clause := range classDecl.HeritageClauses.Nodes {
						if clause.AsHeritageClause() != nil && clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
							return
						}
					}
				}

				// Check for decorators
				hasDecorators := false
				if classDecl.Modifiers() != nil {
					for _, modifier := range classDecl.Modifiers().Nodes {
						if modifier.Kind == ast.KindDecorator {
							hasDecorators = true
							break
						}
					}
				}

				if hasDecorators && opts.AllowWithDecorator {
					return
				}

				// Check class members
				hasNonStaticMember := false
				hasConstructor := false
				hasStaticMember := false
				isEmpty := true

				if classDecl.Members != nil {
					isEmpty = len(classDecl.Members.Nodes) == 0

					for _, member := range classDecl.Members.Nodes {
						// Check if it's a constructor
						if member.Kind == ast.KindConstructor {
							hasConstructor = true
							isEmpty = false

							// Check if constructor has parameter properties (public, private, protected params)
							// These act as class members
							constructor := member.AsConstructorDeclaration()
							if constructor != nil && constructor.Parameters != nil {
								for _, param := range constructor.Parameters.Nodes {
									if param.Kind == ast.KindParameter {
										paramDecl := param.AsParameterDeclaration()
										if paramDecl != nil && paramDecl.Modifiers() != nil {
											for _, mod := range paramDecl.Modifiers().Nodes {
												if mod.Kind == ast.KindPublicKeyword ||
													mod.Kind == ast.KindPrivateKeyword ||
													mod.Kind == ast.KindProtectedKeyword ||
													mod.Kind == ast.KindReadonlyKeyword {
													// This is a parameter property, counts as a non-static member
													hasNonStaticMember = true
													break
												}
											}
										}
									}
								}
							}
							continue
						}

						// Check for static members
						isStatic := false
						if member.Kind == ast.KindPropertyDeclaration {
							prop := member.AsPropertyDeclaration()
							if prop.Modifiers() != nil {
								for _, mod := range prop.Modifiers().Nodes {
									if mod.Kind == ast.KindStaticKeyword {
										isStatic = true
										break
									}
								}
							}
						} else if member.Kind == ast.KindMethodDeclaration {
							method := member.AsMethodDeclaration()
							if method.Modifiers() != nil {
								for _, mod := range method.Modifiers().Nodes {
									if mod.Kind == ast.KindStaticKeyword {
										isStatic = true
										break
									}
								}
							}
						}

						if isStatic {
							hasStaticMember = true
							isEmpty = false
						} else {
							hasNonStaticMember = true
							isEmpty = false
						}
					}
				}

				// Check for abstract class with abstract members
				isAbstract := false
				if classDecl.Modifiers() != nil {
					for _, modifier := range classDecl.Modifiers().Nodes {
						if modifier.Kind == ast.KindAbstractKeyword {
							isAbstract = true
							break
						}
					}
				}

				if isAbstract && classDecl.Members != nil {
					for _, member := range classDecl.Members.Nodes {
						if member.Modifiers() != nil {
							for _, mod := range member.Modifiers().Nodes {
								if mod.Kind == ast.KindAbstractKeyword {
									// Has abstract member, so it's a valid abstract class
									return
								}
							}
						}
					}
				}

				// Report empty class
				if isEmpty {
					if !opts.AllowEmpty {
						ctx.ReportNode(reportNode, rule.RuleMessage{
							Id:          "empty",
							Description: "Unexpected empty class.",
						})
					}
					return
				}

				// Report constructor-only class
				if hasConstructor && !hasNonStaticMember && !hasStaticMember {
					if !opts.AllowConstructorOnly {
						ctx.ReportNode(reportNode, rule.RuleMessage{
							Id:          "onlyConstructor",
							Description: "Unexpected class with only a constructor.",
						})
					}
					return
				}

				// Report static-only class
				if hasStaticMember && !hasNonStaticMember {
					if !opts.AllowStaticOnly {
						ctx.ReportNode(reportNode, rule.RuleMessage{
							Id:          "onlyStatic",
							Description: "Unexpected class with only static properties.",
						})
					}
					return
				}
			},
		}
	},
})
