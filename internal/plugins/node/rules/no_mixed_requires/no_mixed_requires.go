package no_mixed_requires

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_mixed_requires.schema.json
var schemaJSON []byte

var noMixRequire = rule.RuleMessage{
	Id:          "noMixRequire",
	Description: "Do not mix 'require' and other declarations.",
}

var noMixGrouping = rule.RuleMessage{
	Id:          "noMixCoreModuleFileComputed",
	Description: "Do not mix core, module, file and computed requires.",
}

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-mixed-requires.js
var NoMixedRequiresRule = rule.Rule{
	Name:   "node/no-mixed-requires",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		grouping, allowCall := false, false
		if len(options) > 0 {
			switch option := options[0].(type) {
			case bool:
				grouping = option
			case map[string]any:
				grouping, _ = option["grouping"].(bool)
				allowCall, _ = option["allowCall"].(bool)
			}
		}

		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				declarations := node.AsVariableDeclarationList().Declarations.Nodes
				if len(declarations) < 2 {
					return
				}
				hasRequire, hasOther, mixedGroups := false, false, false
				firstGroup := ""
				for _, declaration := range declarations {
					initializer := declaration.AsVariableDeclaration().Initializer
					if !isRequireDeclaration(initializer, allowCall) {
						hasOther = true
					} else {
						hasRequire = true
						if grouping {
							group := moduleType(initializer)
							if firstGroup == "" {
								firstGroup = group
							} else if firstGroup != group {
								mixedGroups = true
							}
						}
					}
					// A mixed declaration takes precedence over any grouping error.
					if hasRequire && hasOther {
						break
					}
				}

				message := noMixGrouping
				if hasRequire && hasOther {
					message = noMixRequire
				} else if !mixedGroups {
					return
				}
				// ESTree includes a statement's semicolon, but excludes export modifiers
				// and the separator after a for-loop initializer.
				textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				if node.Parent.Kind == ast.KindVariableStatement {
					textRange = textRange.WithEnd(node.Parent.End())
					if modifiers := node.Parent.Modifiers(); modifiers != nil {
						for _, modifier := range modifiers.Nodes {
							if modifier.Kind == ast.KindDeclareKeyword {
								textRange = textRange.WithPos(utils.TrimNodeTextRange(ctx.SourceFile, modifier).Pos())
								break
							}
						}
					}
				}
				ctx.ReportRange(textRange, message)
			},
		}
	},
}

func isRequireDeclaration(node *ast.Node, allowCall bool) bool {
	for node = utils.ESTreeRuntimeExpression(node); node != nil && !ast.IsOptionalChain(node); node = utils.ESTreeRuntimeExpression(node) {
		switch node.Kind {
		case ast.KindCallExpression:
			callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
			if callee != nil && callee.Kind == ast.KindIdentifier && callee.Text() == "require" {
				return true
			}
			if !allowCall || callee == nil || callee.Kind != ast.KindCallExpression {
				return false
			}
			node = callee
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			node, _ = utils.MemberExpressionParts(node)
		default:
			return false
		}
	}
	return false
}

// The caller has already classified the initializer as a require declaration.
func moduleType(node *ast.Node) string {
	for node = utils.ESTreeRuntimeExpression(node); ; node = utils.ESTreeRuntimeExpression(node) {
		if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
			node, _ = utils.MemberExpressionParts(node)
			continue
		}
		// With allowCall, upstream classifies the outer call's first argument.
		arguments := node.AsCallExpression().Arguments.Nodes
		if len(arguments) == 0 {
			return "computed"
		}
		argument := utils.ESTreeRuntimeExpression(arguments[0])
		if argument.Kind != ast.KindStringLiteral {
			return "computed"
		}
		name := argument.Text()
		if isCoreModule(name) {
			return "core"
		}
		if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
			return "file"
		}
		return "module"
	}
}

// Upstream freezes this list at Node 13.8.0. Modern builtin resolution would
// change grouping for node: specifiers and newer modules, so keep it rule-local.
func isCoreModule(name string) bool {
	switch name {
	case "_http_agent", "_http_client", "_http_common", "_http_incoming", "_http_outgoing", "_http_server",
		"_stream_duplex", "_stream_passthrough", "_stream_readable", "_stream_transform", "_stream_wrap", "_stream_writable",
		"_tls_common", "_tls_wrap", "assert", "async_hooks", "buffer", "child_process", "cluster", "console",
		"constants", "crypto", "dgram", "dns", "domain", "events", "fs", "http", "http2", "https", "inspector",
		"module", "net", "os", "path", "perf_hooks", "process", "punycode", "querystring", "readline", "repl",
		"stream", "string_decoder", "sys", "timers", "tls", "trace_events", "tty", "url", "util", "v8", "vm",
		"worker_threads", "zlib":
		return true
	}
	return false
}
