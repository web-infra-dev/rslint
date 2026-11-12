package no_var_requires

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed no_var_requires.schema.json
var schemaJSON []byte

type Options struct {
	Allow []string `json:"allow"`
}

func parseOptions(options []any) *Options {
	opts := &Options{}
	if len(options) == 0 {
		return opts
	}
	optMap, _ := options[0].(map[string]interface{})
	if allowSlice, ok := optMap["allow"].([]interface{}); ok {
		for _, item := range allowSlice {
			if str, ok := item.(string); ok {
				opts.Allow = append(opts.Allow, str)
			}
		}
	}
	return opts
}

var NoVarRequiresRule = rule.CreateRule(rule.Rule{
	Name:   "no-var-requires",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		// Compile allow patterns into regexes
		var allowPatterns []*esregexp.RegExp
		for _, pattern := range opts.Allow {
			if re, err := esregexp.Compile(pattern, "u"); err == nil {
				allowPatterns = append(allowPatterns, re)
			}
		}

		isImportPathAllowed := func(importPath string) bool {
			for _, pattern := range allowPatterns {
				if pattern.TestOrTimeout(importPath) {
					return true
				}
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				callee := ast.SkipParentheses(callExpr.Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "require" {
					return
				}

				// Only literal strings and templates without substitutions are
				// static here. Their AST text is already decoded, including an
				// empty string; parentheses are transparent in the upstream AST.
				if args := callExpr.Arguments; args != nil && len(args.Nodes) > 0 {
					arg := ast.SkipParentheses(args.Nodes[0])
					if value, ok := utils.GetStaticStringLiteralValue(arg); ok && isImportPathAllowed(value) {
						return
					}
				}

				// ESTree omits parentheses. Match the immediate semantic parent,
				// rather than treating every descendant of a variable as an import.
				parent := node.Parent
				for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
					parent = parent.Parent
				}
				if parent == nil {
					return
				}
				switch parent.Kind {
				case ast.KindCallExpression, ast.KindPropertyAccessExpression,
					ast.KindElementAccessExpression, ast.KindNewExpression,
					ast.KindAsExpression, ast.KindTypeAssertionExpression,
					ast.KindVariableDeclaration:
					// findVariable in typescript-eslint counts identifiers in
					// both type and value spaces. Keep lookup lexical to this
					// file so ambient globals from other files never hide reports.
					symbol := ctx.Refs.ResolveInFileWithMeaning(callee,
						ast.SymbolFlagsValue|ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
					if symbol != nil && len(symbol.Declarations) > 0 {
						return
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Description: "Require statement not part of import statement.",
						Id:          "noVarReqs",
					})
				}
			},
		}
	},
})
