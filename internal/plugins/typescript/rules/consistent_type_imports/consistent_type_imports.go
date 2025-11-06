package consistent_type_imports

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConsistentTypeImportsOptions struct {
	Prefer                  string `json:"prefer"`
	DisallowTypeAnnotations bool   `json:"disallowTypeAnnotations"`
	FixStyle                string `json:"fixStyle"`
}

// ConsistentTypeImportsRule enforces consistent type imports
var ConsistentTypeImportsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-type-imports",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentTypeImportsOptions{
		Prefer:                  "type-imports",
		DisallowTypeAnnotations: true,
		FixStyle:                "separate-type-imports",
	}

	// Parse options
	if options != nil {
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if optMap, ok := optArray[0].(map[string]interface{}); ok {
				if prefer, ok := optMap["prefer"].(string); ok {
					opts.Prefer = prefer
				}
				if disallow, ok := optMap["disallowTypeAnnotations"].(bool); ok {
					opts.DisallowTypeAnnotations = disallow
				}
				if fixStyle, ok := optMap["fixStyle"].(string); ok {
					opts.FixStyle = fixStyle
				}
			}
		} else if optMap, ok := options.(map[string]interface{}); ok {
			if prefer, ok := optMap["prefer"].(string); ok {
				opts.Prefer = prefer
			}
			if disallow, ok := optMap["disallowTypeAnnotations"].(bool); ok {
				opts.DisallowTypeAnnotations = disallow
			}
			if fixStyle, ok := optMap["fixStyle"].(string); ok {
				opts.FixStyle = fixStyle
			}
		}
	}

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
		if importClause.IsTypeOnly {
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
