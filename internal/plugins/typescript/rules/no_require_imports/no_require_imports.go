package no_require_imports

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// getStaticStringValue extracts static string value from literal or template
func getStaticStringValue(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindTemplateExpression:
		// Only handle simple template literals without expressions
		te := node.AsTemplateExpression()
		if te.TemplateSpans == nil || len(te.TemplateSpans.Nodes) == 0 {
			return te.Head.Text(), true
		}
	case ast.KindNoSubstitutionTemplateLiteral:
		// Handle simple template literals `string`
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	}
	return "", false
}

var noRequireImportsMessage = rule.RuleMessage{
	Id:          "noRequireImports",
	Description: "A `require()` style import is forbidden.",
}

var NoRequireImportsRule = rule.CreateRule(rule.Rule{
	Name: "no-require-imports",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		optsMap := utils.GetOptionsMap(options)
		allowAsImport, _ := optsMap["allowAsImport"].(bool)
		var allowPatterns []*regexp.Regexp
		if allow, ok := optsMap["allow"].([]interface{}); ok {
			for _, rawPattern := range allow {
				pattern, ok := rawPattern.(string)
				if !ok {
					continue
				}
				if compiled, err := regexp.Compile(pattern); err == nil {
					allowPatterns = append(allowPatterns, compiled)
				}
			}
		}

		isImportPathAllowed := func(importPath string) bool {
			for _, pattern := range allowPatterns {
				if pattern.MatchString(importPath) {
					return true
				}
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				requireIdentifier := callExpr.Expression
				if !ast.IsIdentifier(requireIdentifier) || requireIdentifier.AsIdentifier().Text != "require" {
					return
				}

				// Check if first argument matches allowed patterns
				if len(callExpr.Arguments.Nodes) > 0 {
					if argValue, ok := getStaticStringValue(callExpr.Arguments.Nodes[0]); ok && isImportPathAllowed(argValue) {
						return
					}
				}

				// Ignore user-land bindings named require. Resolve follows the
				// binder's lexical scope chain, while the declaration-in-file check
				// keeps ambient/config globals classified as the CommonJS require.
				if symbol := ctx.Refs.Resolve(requireIdentifier); utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile) {
					return
				}

				ctx.ReportNode(node, noRequireImportsMessage)
			},

			ast.KindExternalModuleReference: func(node *ast.Node) {
				extModRef := node.AsExternalModuleReference()

				// Check if expression matches allowed patterns
				if argValue, ok := getStaticStringValue(extModRef.Expression); ok && isImportPathAllowed(argValue) {
					return
				}

				// Check if allowAsImport is true and parent is TSImportEqualsDeclaration
				if allowAsImport && node.Parent != nil &&
					node.Parent.Kind == ast.KindImportEqualsDeclaration {
					return
				}

				ctx.ReportNode(node, noRequireImportsMessage)
			},
		}
	},
})
