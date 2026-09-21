package es_builtins

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
	"github.com/web-infra-dev/rslint/internal/utils/referencetracker"
)

//go:embed es_builtins.schema.json
var schemaJSON []byte

// The upstream table is constant. Parse each support range once rather than
// repeating semver parsing for every file, including files with no API reads.
var supportRanges = func() map[string]npmsemver.Range {
	ranges := map[string]npmsemver.Range{}
	for _, builtin := range builtins {
		if _, exists := ranges[builtin.since]; exists {
			continue
		}
		parsed, ok := npmsemver.Parse(">=" + builtin.since)
		if !ok {
			panic("invalid ES builtin support version: " + builtin.since)
		}
		ranges[builtin.since] = parsed
	}
	return ranges
}()

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unsupported-features/es-builtins.js
var ESBuiltinsRule = rule.Rule{
	Name:   "node/no-unsupported-features/es-builtins",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		ignores, _ := opts["ignores"].([]any)
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			version := nodeutil.ConfiguredNodeVersion(ctx, opts)
			globals := map[string]*referencetracker.Trace{}
			// Every entry in the pinned ES table has one stable support version.
			// Reuse decisions for entries introduced in the same release.
			supported := map[string]bool{}
			type diagnostic struct {
				node    *ast.Node
				message rule.RuleMessage
			}
			var diagnostics []diagnostic
			for _, builtin := range builtins {
				if slices.Contains(ignores, any(builtin.name)) {
					continue
				}
				available, checked := supported[builtin.since]
				if !checked {
					available = version.IsSubsetOfRange(supportRanges[builtin.since])
					supported[builtin.since] = available
				}
				if available {
					continue
				}
				properties := globals
				var trace *referencetracker.Trace
				for _, part := range strings.Split(builtin.name, ".") {
					trace = properties[part]
					if trace == nil {
						trace = &referencetracker.Trace{Properties: map[string]*referencetracker.Trace{}}
						properties[part] = trace
					}
					properties = trace.Properties
				}
				trace.Read = func(node *ast.Node) {
					// JSDoc type nodes are comments to the upstream JavaScript parser.
					if utils.IsJSDocSyntaxNode(node) {
						return
					}
					diagnostics = append(diagnostics, diagnostic{node, rule.RuleMessage{
						Id: "not-supported-till",
						Description: "The '" + builtin.name + "' is still an experimental feature and is not supported until Node.js " +
							builtin.since + ". The configured version range is '" + version.Raw() + "'.",
						Data: map[string]string{"name": builtin.name, "supported": builtin.since, "version": version.Raw()},
					}})
				}
			}
			if len(globals) == 0 {
				return
			}
			referencetracker.New(ctx).TrackGlobals(globals)
			slices.SortStableFunc(diagnostics, func(a, b diagnostic) int {
				return cmp.Or(cmp.Compare(a.node.Pos(), b.node.Pos()), cmp.Compare(a.node.End(), b.node.End()))
			})
			for _, diagnostic := range diagnostics {
				ctx.ReportNode(diagnostic.node, diagnostic.message)
			}
		}}
	},
}
