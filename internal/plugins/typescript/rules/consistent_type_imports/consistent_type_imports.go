package consistent_type_imports

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed consistent_type_imports.schema.json
var schemaJSON []byte

type ConsistentTypeImportsOptions struct {
	Prefer                  string `json:"prefer"`
	DisallowTypeAnnotations bool   `json:"disallowTypeAnnotations"`
	FixStyle                string `json:"fixStyle"`
}

// ConsistentTypeImportsRule enforces consistent type imports
var ConsistentTypeImportsRule = rule.CreateRule(rule.Rule{
	Name:   "consistent-type-imports",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

func parseOptions(options []any) ConsistentTypeImportsOptions {
	opts := ConsistentTypeImportsOptions{
		Prefer:                  "type-imports",
		DisallowTypeAnnotations: true,
		FixStyle:                "separate-type-imports",
	}
	if len(options) == 0 {
		return opts
	}

	optMap, _ := options[0].(map[string]interface{})
	if prefer, ok := optMap["prefer"].(string); ok {
		opts.Prefer = prefer
	}
	if disallow, ok := optMap["disallowTypeAnnotations"].(bool); ok {
		opts.DisallowTypeAnnotations = disallow
	}
	if fixStyle, ok := optMap["fixStyle"].(string); ok {
		opts.FixStyle = fixStyle
	}
	return opts
}

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)

	checkImportDeclaration := func(node *ast.Node) {
		importDecl := node.AsImportDeclaration()
		if importDecl == nil {
			return
		}

		importClauseNode := importDecl.ImportClause
		if importClauseNode == nil {
			return
		}

		importClause := importClauseNode.AsImportClause()
		if importClause == nil {
			return
		}

		// Skip if entire import is already type-only
		if importClause.IsTypeOnly() {
			// If prefer is 'no-type-imports', report error
			if opts.Prefer == "no-type-imports" {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "avoidImportType",
					Description: "Use regular imports instead of import type.",
				})
			}
			return
		}

		// For now, implement basic check: if prefer is 'type-imports',
		// we need to analyze imports to see if they're only used in type positions
		// This is a simplified version - a full implementation would require
		// tracking all references to imported symbols throughout the file

	}

	checkTSImportType := func(node *ast.Node) {
		if opts.DisallowTypeAnnotations {
			// Check if this is an import type in a type annotation position
			importType := node.AsImportTypeNode()
			if importType != nil {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noImportTypeAnnotations",
					Description: "Type imports in type annotations are not allowed.",
				})
			}
		}
	}

	return rule.RuleListeners{
		ast.KindImportDeclaration: checkImportDeclaration,
		ast.KindImportType:        checkTSImportType,
	}
}
