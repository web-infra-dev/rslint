package enforce_node_protocol_usage

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/semver"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

//go:embed enforce_node_protocol_usage.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/enforce-node-protocol-usage.js
var EnforceNodeProtocolUsageRule = rule.Rule{
	Name:   "import/enforce-node-protocol-usage",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var mode string
		if len(options) > 0 {
			mode, _ = options[0].(string)
		}
		var version semver.Version
		var versionErr error
		versionChecked := false
		check := func(source *ast.Node) {
			source = utils.ESTreeRuntimeExpression(source)
			if source == nil || source.Kind != ast.KindStringLiteral {
				return
			}
			if !versionChecked {
				version, versionErr = import_utils.NodeVersion(ctx.Settings)
				versionChecked = true
			}
			if versionErr != nil {
				ctx.ReportNode(source, rule.RuleMessage{Description: versionErr.Error()})
				return
			}
			name := source.Text()
			prefixed := strings.HasPrefix(name, "node:")
			replacement, removed := "node:", 0
			if mode == "never" {
				if !prefixed {
					return
				}
				name = strings.TrimPrefix(name, "node:")
				replacement, removed = "", 5
			} else if mode != "always" || prefixed {
				return
			}
			if !modules.IsNodeBuiltinAtVersion(name, version) {
				return
			}
			if mode == "always" && !modules.IsNodeBuiltinAtVersion("node:"+name, version) {
				return
			}
			preferred, other := "node:"+name, name
			if mode == "never" {
				preferred, other = other, preferred
			}
			ctx.ReportNodeWithDeferredFixes(source, rule.RuleMessage{
				Description: "Prefer `" + preferred + "` over `" + other + "`.",
				Data:        map[string]string{"moduleName": name},
			}, func() []rule.RuleFix {
				start := utils.TrimNodeTextRange(ctx.SourceFile, source).Pos() + 1
				// An escaped prefix is not five raw characters. Replace the literal
				// contents with the known builtin name instead of corrupting it.
				if mode == "never" && !strings.HasPrefix(ctx.SourceFile.Text()[start:source.End()-1], "node:") {
					return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(start, source.End()-1), name)}
				}
				return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(start, start+removed), replacement)}
			})
		}
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				check(node.ModuleSpecifier())
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				// ExportAllDeclaration (including export * as ns) has no upstream listener.
				clause := node.AsExportDeclaration().ExportClause
				if clause != nil && clause.Kind == ast.KindNamedExports {
					check(node.ModuleSpecifier())
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				args := node.Arguments()
				if len(args) == 0 {
					return
				}
				// Unlike tsgo's IsImportCall, upstream does not match import.defer().
				if node.Expression().Kind == ast.KindImportKeyword {
					check(args[0])
					return
				}
				call := node.AsCallExpression()
				callee := utils.ESTreeCallCallee(call.Expression)
				if len(args) == 1 && call.QuestionDotToken == nil && callee != nil && callee.Kind == ast.KindIdentifier && callee.Text() == "require" {
					check(args[0])
				}
			},
		}
	},
}
