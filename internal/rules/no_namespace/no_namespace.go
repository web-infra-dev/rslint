package no_namespace

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/rslint/internal/rule"
)

type NoNamespaceOptions struct {
	AllowDeclarations    bool `json:"allowDeclarations"`
	AllowDefinitionFiles bool `json:"allowDefinitionFiles"`
}

// isDeclaration checks if the node or any of its ancestors has the declare modifier
func isDeclaration(node *ast.Node) bool {
	if node.Kind == ast.KindModuleDeclaration {
		moduleDecl := node.AsModuleDeclaration()
		if moduleDecl.Modifiers() != nil {
			for _, modifier := range moduleDecl.Modifiers().Nodes {
				if modifier.Kind == ast.KindDeclareKeyword {
					return true
				}
			}
		}
	}

	if node.Parent != nil {
		return isDeclaration(node.Parent)
	}

	return false
}

var NoNamespaceRule = rule.Rule{
	Name: "no-namespace",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := NoNamespaceOptions{
			AllowDeclarations:    false,
			AllowDefinitionFiles: true,
		}
		if options != nil {
			if optsMap, ok := options.(map[string]interface{}); ok {
				if allowDeclarations, ok := optsMap["allowDeclarations"].(bool); ok {
					opts.AllowDeclarations = allowDeclarations
				}
				if allowDefinitionFiles, ok := optsMap["allowDefinitionFiles"].(bool); ok {
					opts.AllowDefinitionFiles = allowDefinitionFiles
				}
			}
		}

		return rule.RuleListeners{
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()

				// Skip global declarations
				if ast.IsGlobalScopeAugmentation(node) {
					return
				}

				// Skip module declarations with string literal names (like "module 'foo' {}")
				if moduleDecl.Name() != nil && ast.IsStringLiteral(moduleDecl.Name()) {
					return
				}

				// Skip if parent is also a module declaration (nested namespaces)
				if node.Parent != nil && node.Parent.Kind == ast.KindModuleDeclaration {
					return
				}

				// Check if allowed by options
				if opts.AllowDefinitionFiles && strings.HasSuffix(ctx.SourceFile.FileName(), ".d.ts") {
					return
				}

				if opts.AllowDeclarations && isDeclaration(node) {
					return
				}

				// Report the error
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "moduleSyntaxIsPreferred",
					Description: "ES2015 module syntax is preferred over namespaces.",
				})
			},
		}
	},
}