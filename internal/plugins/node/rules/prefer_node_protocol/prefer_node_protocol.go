package prefer_node_protocol

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed prefer_node_protocol.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/prefer-node-protocol.js
var PreferNodeProtocolRule = rule.Rule{
	Name:   "node/prefer-node-protocol",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var option map[string]any
		if len(options) > 0 {
			option, _ = options[0].(map[string]any)
		}
		var versionChecked, esmEnabled, cjsEnabled bool
		check := func(source *ast.Node, style string) {
			source = utils.ESTreeRuntimeExpression(source)
			if source == nil || source.Kind != ast.KindStringLiteral || !shouldPrefix(source.Text()) {
				return
			}
			// getBuiltinModule itself implies support. Other forms only need
			// version configuration once there is a builtin to check.
			if style != "getBuiltinModule" {
				if !versionChecked {
					version := nodeutil.ConfiguredNodeVersion(ctx, option)
					esmEnabled = version.IsSubsetOf("^12.20.0 || >=14.13.1")
					cjsEnabled = esmEnabled && version.IsSubsetOf("^14.18.0 || >=16.0.0")
					versionChecked = true
				}
				if !esmEnabled || style == "require" && !cjsEnabled {
					return
				}
			}
			name := source.Text()
			ctx.ReportNodeWithDeferredFixes(source, rule.RuleMessage{
				Id:          "preferNodeProtocol",
				Description: "Prefer `node:" + name + "` over `" + name + "`.",
				Data:        map[string]string{"moduleName": name},
			}, func() []rule.RuleFix {
				position := utils.TrimNodeTextRange(ctx.SourceFile, source).Pos() + 1
				return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(position, position), "node:")}
			})
		}
		checkModule := func(node *ast.Node) {
			check(ast.GetExternalModuleName(node), "import")
		}
		var propertyEvaluator *utils.StaticStringEvaluator
		propertyName := func(node *ast.Node) string {
			if name, ok := utils.AccessExpressionStaticName(node); ok {
				return name
			}
			if node.Kind == ast.KindElementAccessExpression {
				if propertyEvaluator == nil {
					propertyEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
				}
				name, _ := propertyEvaluator.EvalToString(node.AsElementAccessExpression().ArgumentExpression)
				return name
			}
			return ""
		}
		return rule.RuleListeners{
			ast.KindImportDeclaration: checkModule,
			ast.KindExportDeclaration: checkModule,
			ast.KindCallExpression: func(node *ast.Node) {
				args := node.Arguments()
				if len(args) == 0 {
					return
				}
				if ast.IsImportCall(node) {
					check(args[0], "import")
					return
				}
				call := node.AsCallExpression()
				callee := utils.ESTreeCallCallee(call.Expression)
				if callee == nil {
					return
				}
				if callee.Kind == ast.KindIdentifier {
					if callee.Text() == "require" && len(args) == 1 && call.QuestionDotToken == nil {
						check(args[0], "require")
					}
					return
				}
				if !ast.IsAccessExpression(callee) {
					return
				}
				object := utils.ESTreeCallCallee(utils.AccessExpressionObject(callee))
				if object == nil {
					return
				}
				if ast.IsAccessExpression(object) {
					root := utils.ESTreeCallCallee(utils.AccessExpressionObject(object))
					if root == nil || root.Kind != ast.KindIdentifier || root.Text() != "globalThis" || propertyName(object) != "process" {
						return
					}
				} else if object.Kind != ast.KindIdentifier || object.Text() != "process" {
					return
				}
				if propertyName(callee) == "getBuiltinModule" {
					check(args[0], "getBuiltinModule")
				}
			},
		}
	},
}

func shouldPrefix(name string) bool {
	// eslint-plugin-n's NodeBuiltinModules has no node: counterpart for these
	// legacy entries. Keep this rule's policy separate from resolution, which
	// correctly accepts them, and reuse tsgo's remaining builtin names.
	switch name {
	case "constants", "domain", "punycode", "sys":
		return false
	}
	return core.UnprefixedNodeCoreModules[name]
}
