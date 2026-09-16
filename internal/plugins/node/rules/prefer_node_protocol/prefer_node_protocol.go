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
		var bindings map[*ast.Symbol]bool
		isNodeReference := func(identifier *ast.Node) bool {
			symbol := ctx.Refs.ResolveInFile(identifier)
			if symbol == nil {
				// Includes the implicit CommonJS require binding.
				return true
			}
			if result, ok := bindings[symbol]; ok {
				return result
			}
			if bindings == nil {
				bindings = make(map[*ast.Symbol]bool)
			}
			result := isNodeBinding(ctx, identifier, symbol)
			bindings[symbol] = result
			return result
		}
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
					if callee.Text() == "require" && len(args) == 1 && call.QuestionDotToken == nil && isNodeReference(callee) {
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
					if root == nil || root.Kind != ast.KindIdentifier || root.Text() != "globalThis" || propertyName(object) != "process" || !isNodeReference(root) {
						return
					}
				} else if object.Kind != ast.KindIdentifier || object.Text() != "process" || !isNodeReference(object) {
					return
				}
				if propertyName(callee) == "getBuiltinModule" {
					check(args[0], "getBuiltinModule")
				}
			},
		}
	},
}

// The rule recognizes Node globals and direct Node imports. Unknown local
// bindings must not receive a fix that changes an arbitrary function's input.
func isNodeBinding(ctx rule.RuleContext, identifier *ast.Node, symbol *ast.Symbol) bool {
	if identifier.Text() == "process" && isBuiltinImport(ctx, identifier, "process", "default") {
		return true
	}
	if identifier.Text() != "require" && identifier.Text() != "process" || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	initializer := utils.ESTreeRuntimeExpression(declaration.Initializer())
	if initializer == nil || initializer.Kind != ast.KindCallExpression || initializer.AsCallExpression().QuestionDotToken != nil {
		return false
	}
	if identifier.Text() == "require" {
		if !isBuiltinImport(ctx, initializer.Expression(), "module", "createRequire") {
			return false
		}
	} else {
		callee := utils.ESTreeCallCallee(initializer.Expression())
		args := initializer.Arguments()
		if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "require" || len(args) != 1 {
			return false
		}
		if requireSymbol := ctx.Refs.ResolveInFile(callee); requireSymbol != nil && !isNodeBinding(ctx, callee, requireSymbol) {
			return false
		}
		source := utils.ESTreeRuntimeExpression(args[0])
		if source == nil || source.Kind != ast.KindStringLiteral || source.Text() != "process" && source.Text() != "node:process" {
			return false
		}
	}
	for _, reference := range ctx.Refs.References(symbol) {
		if utils.IsWriteReference(reference) {
			return false
		}
	}
	return true
}

func isBuiltinImport(ctx rule.RuleContext, node *ast.Node, moduleName, exportName string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	if ast.IsAccessExpression(node) {
		name, ok := utils.AccessExpressionStaticName(node)
		return ok && name == exportName && isBuiltinImport(ctx, utils.AccessExpressionObject(node), moduleName, "default")
	}
	if node.Kind != ast.KindIdentifier {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if ast.IsTypeOnlyImportDeclaration(declaration) {
		return false
	}
	switch declaration.Kind {
	case ast.KindImportClause, ast.KindNamespaceImport:
		if exportName != "default" {
			return false
		}
	case ast.KindImportSpecifier:
		name := declaration.PropertyName()
		if name == nil {
			name = declaration.Name()
		}
		if name.Text() != exportName {
			return false
		}
	default:
		return false
	}
	source := ast.GetExternalModuleName(ast.FindAncestorKind(declaration, ast.KindImportDeclaration))
	return source != nil && (source.Text() == moduleName || source.Text() == "node:"+moduleName)
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
