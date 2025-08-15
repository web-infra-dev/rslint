package no_namespace

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// build the message for no-namespace rule
func buildNoNamespaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "moduleSyntaxIsPreferred",
		Description: "Namespace is not allowed.",
	}
}

// rule options
type NoNamespaceOptions struct {
	AllowDeclarations    *bool `json:"allowDeclarations"`
	AllowDefinitionFiles *bool `json:"allowDefinitionFiles"`
}

// default options
var defaultNoNamespaceOptions = NoNamespaceOptions{
	AllowDeclarations:    utils.Ref(false),
	AllowDefinitionFiles: utils.Ref(true),
}

// rule instance
// check if the namespace is used
var NoNamespaceRule = rule.CreateRule(rule.Rule{
	Name: "no-namespace",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := defaultNoNamespaceOptions

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
				if allowDeclarations, ok := optsMap["allowDeclarations"].(bool); ok {
					opts.AllowDeclarations = utils.Ref(allowDeclarations)
				}
				if allowDefinitionFiles, ok := optsMap["allowDefinitionFiles"].(bool); ok {
					opts.AllowDefinitionFiles = utils.Ref(allowDefinitionFiles)
				}
			}
		}

		return rule.RuleListeners{
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if moduleDecl == nil {
					return
				}

				// Check if this is a namespace declaration (keyword is KindNamespaceKeyword)
				if moduleDecl.Keyword != ast.KindNamespaceKeyword {
					return
				}

				// Check if we're in a .d.ts file and allowDefinitionFiles is true
				if opts.AllowDefinitionFiles != nil && *opts.AllowDefinitionFiles && strings.HasSuffix(ctx.SourceFile.FileName(), ".d.ts") {
					return
				}

				// Check if this is a declare namespace and allowDeclarations is true
				if opts.AllowDeclarations != nil && *opts.AllowDeclarations && utils.IncludesModifier(node, ast.KindDeclareKeyword) {
					return
				}

				// Report the namespace usage
				ctx.ReportNode(moduleDecl.Name(), buildNoNamespaceMessage())
			},
		}
	},
})
