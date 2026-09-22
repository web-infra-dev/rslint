package node_builtins

import (
	"cmp"
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/npmsemver"
)

//go:embed node_builtins.schema.json
var schemaJSON []byte

type builtinNode struct{ parent, text uint16 }
type builtinFeature struct {
	node, supported, experimental uint16
	kind                          uint8
}
type builtinAlias struct{ parent, text, target uint16 }
type supportInfo struct {
	supported, experimental string
	order                   int
}
type builtinAPI struct {
	read, call, construct *supportInfo
	properties            map[string]*builtinAPI
}

func newBuiltinAPIs() (global, module, importMeta *builtinAPI) {
	apis := make([]*builtinAPI, len(builtinNodes))
	for i := range apis {
		apis[i] = &builtinAPI{}
	}
	for i := 3; i < len(builtinNodes); i++ {
		node := builtinNodes[i]
		parent := apis[node.parent]
		if parent.properties == nil {
			parent.properties = make(map[string]*builtinAPI)
		}
		parent.properties[builtinTexts[node.text]] = apis[i]
	}
	for order, feature := range builtinFeatures {
		info := &supportInfo{builtinVersions[feature.supported], builtinVersions[feature.experimental], order}
		api := apis[feature.node]
		switch feature.kind {
		case 0:
			api.read = info
		case 1:
			api.call = info
		case 2:
			api.construct = info
		}
	}
	for _, alias := range builtinAliases {
		parent := apis[alias.parent]
		if parent.properties == nil {
			parent.properties = make(map[string]*builtinAPI)
		}
		parent.properties[builtinTexts[alias.text]] = apis[alias.target]
	}
	return apis[0], apis[1], apis[2]
}

var globalAPIs, moduleAPIs, importMetaAPIs = newBuiltinAPIs()

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unsupported-features/node-builtins.js
var NodeBuiltinsRule = rule.Rule{
	Name:   "node/no-unsupported-features/node-builtins",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		version := nodeutil.ConfiguredNodeVersion(ctx, opts)
		allowExperimental, _ := opts["allowExperimental"].(bool)
		ignores, _ := opts["ignores"].([]any)
		type diagnostic struct {
			node    *ast.Node
			message rule.RuleMessage
			order   int
		}
		var diagnostics []diagnostic
		type availability struct{ id, label string }
		checked := map[*supportInfo]availability{}
		var traceFor func(*builtinAPI, string) *nodeutil.ReferenceTrace
		children := func(api *builtinAPI, path string) map[string]*nodeutil.ReferenceTrace {
			if len(api.properties) == 0 {
				return nil
			}
			properties := make(map[string]*nodeutil.ReferenceTrace, len(api.properties))
			for name, child := range api.properties {
				childPath := name
				if path != "" {
					childPath = path + "." + name
				}
				properties[name] = traceFor(child, childPath)
			}
			return properties
		}
		report := func(info *supportInfo, path string, node *ast.Node) {
			name := strings.TrimPrefix(path, "node:")
			if info == nil || utils.IsJSDocSyntaxNode(node) || slices.Contains(ignores, any(name)) {
				return
			}
			result, ok := checked[info]
			if !ok {
				if allowExperimental && info.experimental != "" {
					if !supports(version, info.experimental) {
						result = availability{"not-experimental-till", versionLabel(info.experimental)}
					}
				} else if info.supported == "" {
					result.id = "not-supported-yet"
				} else if !supports(version, info.supported) {
					result = availability{"not-supported-till", versionLabel(info.supported)}
				}
				checked[info] = result
			}
			if result.id == "" {
				return
			}
			message := "The '" + name + "' is still an experimental feature"
			switch result.id {
			case "not-experimental-till":
				message = "The '" + name + "' is not an experimental feature until Node.js " + result.label + "."
			case "not-supported-till":
				message += " and is not supported until Node.js " + result.label + "."
			}
			message += " The configured version range is '" + version.Raw() + "'."
			diagnostics = append(diagnostics, diagnostic{node, rule.RuleMessage{Id: result.id, Description: message}, info.order})
		}
		traceFor = func(api *builtinAPI, path string) *nodeutil.ReferenceTrace {
			trace := &nodeutil.ReferenceTrace{}
			trace.Read = func(node *ast.Node) {
				// Expand only visited properties, including recursive namespaces
				// such as module.Module, without eagerly expanding their cycles.
				if trace.Properties == nil {
					trace.Properties = children(api, path)
				}
				report(api.read, path, node)
			}
			if api.call != nil {
				trace.Call = func(node *ast.Node) { report(api.call, path, node) }
			}
			if api.construct != nil {
				trace.Construct = func(node *ast.Node) { report(api.construct, path, node) }
			}
			return trace
		}
		tracker := nodeutil.NewReferenceTracker(ctx)
		var importMetaTrace *nodeutil.ReferenceTrace
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindMetaProperty): func(node *ast.Node) {
				if ast.IsImportMeta(node) {
					if importMetaTrace == nil {
						importMetaTrace = &nodeutil.ReferenceTrace{Properties: children(importMetaAPIs, "import.meta")}
					}
					tracker.TrackExpression(node, importMetaTrace)
				}
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				modules := children(moduleAPIs, "")
				// The upstream table explicitly lists supported spellings. Do not
				// synthesize node: aliases for legacy modules absent from that table.
				for name := range moduleAPIs.properties {
					if !strings.HasPrefix(name, "node:") && modules["node:"+name] == nil {
						modules["node:"+name] = &nodeutil.ReferenceTrace{}
					}
				}
				tracker.TrackModules(modules)
				tracker.TrackGlobals(children(globalAPIs, ""))
				slices.SortStableFunc(diagnostics, func(a, b diagnostic) int {
					if a.node == b.node && a.node.Kind == ast.KindExportDeclaration {
						return cmp.Compare(a.order, b.order)
					}
					return cmp.Compare(a.node.Pos(), b.node.Pos())
				})
				for _, diagnostic := range diagnostics {
					ctx.ReportNode(diagnostic.node, diagnostic.message)
				}
			},
		}
	},
}

// Releases in the pinned table are sorted newest first by the generator.
func supports(version npmsemver.Range, releases string) bool {
	versions := strings.Split(releases, ",")
	return version.IsSubsetOf("^" + strings.Join(versions, " || ^") + " || >=" + versions[0])
}

func versionLabel(releases string) string {
	latest, backports, hasBackports := strings.Cut(releases, ",")
	if !hasBackports {
		return latest
	}
	return latest + " (backported: ^" + strings.ReplaceAll(backports, ",", ", ^") + ")"
}
