package no_namespace

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var noNamespaceMessage = rule.RuleMessage{
	Id:          "moduleSyntaxIsPreferred",
	Description: "ES2015 module syntax is preferred over namespaces.",
}

type NoNamespaceOptions struct {
	AllowDeclarations    bool `json:"allowDeclarations"`
	AllowDefinitionFiles bool `json:"allowDefinitionFiles"`
}

var defaultNoNamespaceOptions = NoNamespaceOptions{
	AllowDeclarations:    false,
	AllowDefinitionFiles: true,
}

func parseNoNamespaceOptions(options []any) NoNamespaceOptions {
	opts := defaultNoNamespaceOptions
	if len(options) == 0 {
		return opts
	}

	optsMap, ok := options[0].(map[string]interface{})
	if !ok {
		return opts
	}
	if allowDeclarations, ok := optsMap["allowDeclarations"].(bool); ok {
		opts.AllowDeclarations = allowDeclarations
	}
	if allowDefinitionFiles, ok := optsMap["allowDefinitionFiles"].(bool); ok {
		opts.AllowDefinitionFiles = allowDefinitionFiles
	}
	return opts
}

// isDeclaredNamespace mirrors typescript-eslint's recursive `declare` check.
// In ordinary source files the parser propagates the ambient context flag to
// every descendant, which makes the common path constant-time. Declaration
// files are implicitly ambient, so when allowDefinitionFiles is disabled we
// must instead look for an explicit `declare` modifier on the namespace or one
// of its containing module declarations.
func isDeclaredNamespace(node *ast.Node, isDefinitionFile bool) bool {
	if !isDefinitionFile && node.Flags&ast.NodeFlagsAmbient != 0 {
		return true
	}
	for current := node; current != nil; current = current.Parent {
		if current.Kind == ast.KindModuleDeclaration &&
			ast.HasSyntacticModifier(current, ast.ModifierFlagsAmbient) {
			return true
		}
	}
	return false
}

var NoNamespaceRule = rule.CreateRule(rule.Rule{
	Name: "no-namespace",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseNoNamespaceOptions(options)
		if opts.AllowDefinitionFiles && ctx.SourceFile.IsDeclarationFile {
			return nil
		}

		return rule.RuleListeners{
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				name := moduleDecl.Name()
				if name == nil ||
					name.Kind == ast.KindStringLiteral ||
					ast.IsGlobalScopeAugmentation(node) {
					return
				}

				// TypeScript represents `namespace A.B {}` as nested module
				// declarations. TSESTree exposes it as one declaration, so only
				// the outer node corresponds to an upstream listener event.
				if node.Parent != nil && node.Parent.Kind == ast.KindModuleDeclaration {
					return
				}

				if opts.AllowDeclarations && isDeclaredNamespace(node, ctx.SourceFile.IsDeclarationFile) {
					return
				}

				// ReportRange avoids ReportNode's scanner allocation while
				// preserving its trivia-trimmed, whole-declaration range.
				ctx.ReportRange(
					node.Loc.WithPos(scanner.SkipTrivia(ctx.SourceFile.Text(), node.Pos())),
					noNamespaceMessage,
				)
			},
		}
	},
})
