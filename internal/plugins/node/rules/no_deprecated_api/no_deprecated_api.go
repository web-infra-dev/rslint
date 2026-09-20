package no_deprecated_api

import (
	"cmp"
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/npmsemver"
)

//go:embed no_deprecated_api.schema.json
var schemaJSON []byte

type replacement struct{ name, supported string }
type deprecatedAPI struct {
	path, kind, since, removed, replacement string
	alternatives                            []replacement
	global                                  bool
}

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-deprecated-api.js
var NoDeprecatedAPIRule = rule.Rule{
	Name:   "node/no-deprecated-api",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			var version npmsemver.Range
			versionReady := false
			globals, modules := map[string]*nodeutil.ReferenceTrace{}, map[string]*nodeutil.ReferenceTrace{}
			type diagnostic struct {
				node    *ast.Node
				message rule.RuleMessage
				order   int
			}
			var diagnostics []diagnostic
			for order, api := range deprecatedAPIs {
				name := api.path
				switch api.kind {
				case "call":
					name += "()"
				case "construct":
					name = "new " + name + "()"
				}
				ignored := opts["ignoreModuleItems"]
				traces := modules
				if api.global {
					ignored, traces = opts["ignoreGlobalItems"], globals
				}
				if items, ok := ignored.([]any); ok && slices.Contains(items, any(name)) {
					continue
				}
				var trace *nodeutil.ReferenceTrace
				for _, part := range strings.Split(api.path, ".") {
					trace = traces[part]
					if trace == nil {
						trace = &nodeutil.ReferenceTrace{Properties: map[string]*nodeutil.ReferenceTrace{}}
						traces[part] = trace
					}
					traces = trace.Properties
				}
				var message rule.RuleMessage
				report := func(node *ast.Node) {
					if message.Id == "" {
						if len(api.alternatives) > 0 && !versionReady {
							version = nodeutil.ConfiguredNodeVersion(ctx, opts)
							versionReady = true
						}
						message = api.message(name, version)
					}
					diagnostics = append(diagnostics, diagnostic{node, message, order})
				}
				switch api.kind {
				case "read":
					trace.Read = report
				case "call":
					trace.Call = report
				case "construct":
					trace.Construct = report
				}
			}
			tracker := nodeutil.NewReferenceTracker(ctx)
			tracker.TrackGlobals(globals)
			tracker.TrackModules(modules)
			slices.SortStableFunc(diagnostics, func(a, b diagnostic) int {
				if a.node == b.node && a.node.Kind == ast.KindExportDeclaration {
					// Export-all reports share a range. Preserve the upstream API table's
					// order instead of the shared tracker's map enumeration order.
					return cmp.Compare(a.order, b.order)
				}
				return cmp.Compare(a.node.Pos(), b.node.Pos())
			})
			for _, diagnostic := range diagnostics {
				ctx.ReportNode(diagnostic.node, diagnostic.message)
			}
		}}
	},
}

func (api deprecatedAPI) message(name string, version npmsemver.Range) rule.RuleMessage {
	label := "'" + name + "'"
	if !api.global && !strings.Contains(api.path, ".") {
		label += " module"
	}
	if api.removed != "" {
		return rule.RuleMessage{Id: "removed", Description: label + " was deprecated since v" + api.since + ", and removed in v" + api.removed + "."}
	}
	replace := api.replacement
	if len(api.alternatives) > 0 {
		var available []string
		for _, alternative := range api.alternatives {
			if version.IsAtLeast(alternative.supported) {
				available = append(available, alternative.name)
			}
		}
		replace = strings.Join(available, " or ")
	}
	if replace != "" {
		replace = ". Use " + replace + " instead"
	}
	return rule.RuleMessage{Id: "deprecated", Description: label + " was deprecated since v" + api.since + replace + "."}
}
