package no_deprecated_api

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const defaultNodeVersion = ">=16.0.0"

type options struct {
	version           string
	ignoreModuleItems map[string]bool
	ignoreGlobalItems map[string]bool
}

func parseOptions(o any) options {
	opts := options{
		ignoreModuleItems: map[string]bool{},
		ignoreGlobalItems: map[string]bool{},
	}
	m := utils.GetOptionsMap(o)
	if m == nil {
		return opts
	}
	if v, ok := m["version"].(string); ok {
		opts.version = v
	}
	opts.ignoreModuleItems = toStringSet(m["ignoreModuleItems"])
	opts.ignoreGlobalItems = toStringSet(m["ignoreGlobalItems"])
	return opts
}

func toStringSet(v any) map[string]bool {
	out := map[string]bool{}
	if arr, ok := v.([]interface{}); ok {
		for _, e := range arr {
			if s, ok := e.(string); ok {
				out[s] = true
			}
		}
	}
	return out
}

// getConfiguredNodeVersion mirrors upstream's util: options.version, then
// settings.n.version / settings.node.version, then the >=16.0.0 default.
//
// NOTE: Unlike ESLint, rslint does not read `engines.node` / `devEngines` from
// package.json (no file-system lookup in the rule layer); configure the target
// version via the `version` option or `settings.n.version` instead. Documented
// in no_deprecated_api.md.
func getConfiguredNodeVersion(ctx rule.RuleContext, opts options) semverRange {
	if opts.version != "" {
		if r, ok := parseSemverRange(opts.version); ok {
			return r
		}
	}
	if r, ok := versionFromSettings(ctx.Settings, "n"); ok {
		return r
	}
	if r, ok := versionFromSettings(ctx.Settings, "node"); ok {
		return r
	}
	r, _ := parseSemverRange(defaultNodeVersion)
	return r
}

func versionFromSettings(settings map[string]interface{}, key string) (semverRange, bool) {
	if settings == nil {
		return semverRange{}, false
	}
	sub, ok := settings[key].(map[string]interface{})
	if !ok {
		return semverRange{}, false
	}
	v, ok := sub["version"].(string)
	if !ok || v == "" {
		return semverRange{}, false
	}
	return parseSemverRange(v)
}

// toReplaceMessage builds the " . Use X instead" suffix, filtering array-form
// alternatives to those already available on the configured node version.
func toReplaceMessage(info *deprecatedInfo, version semverRange) string {
	var message string
	if info.replacedList != nil {
		var names []string
		for _, entry := range info.replacedList {
			r, ok := parseSemverRange("<" + entry.supported)
			if !ok {
				continue
			}
			if !version.intersects(r) {
				names = append(names, entry.name)
			}
		}
		message = strings.Join(names, " or ")
	} else {
		message = info.replacedText
	}
	if message == "" {
		return ""
	}
	return ". Use " + message + " instead"
}

func reportItem(ctx rule.RuleContext, version semverRange, node *ast.Node, name string, info *deprecatedInfo) {
	if info.removed != "" {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "removed",
			Description: fmt.Sprintf("%s was deprecated since v%s, and removed in v%s.", name, info.since, info.removed),
			Data:        map[string]string{"name": name, "version": info.since, "removed": info.removed},
		})
		return
	}
	replace := toReplaceMessage(info, version)
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "deprecated",
		Description: fmt.Sprintf("%s was deprecated since v%s%s.", name, info.since, replace),
		Data:        map[string]string{"name": name, "version": info.since, "removed": "", "replace": replace},
	})
}

// iterateProcessGetBuiltinModuleReferences handles
// `process.getBuiltinModule('X').Y` chains, mirroring upstream's util of the
// same name: resolve the module from the string argument, then track property
// references off the call expression.
func (t *referenceTracker) iterateProcessGetBuiltinModuleReferences(tm map[string]*traceMap) {
	for _, procId := range t.collectGlobalRefs("process") {
		callNode := getBuiltinModuleCall(procId)
		if callNode == nil {
			continue
		}
		key := moduleNameArg(callNode)
		if key == "" {
			continue
		}
		nextMap, ok := tm[key]
		if !ok {
			continue
		}
		path := []string{key}
		if nextMap.read != nil {
			t.emit(callNode, path, refRead, nextMap.read)
		}
		t.iteratePropertyReferences(callNode, path, nextMap)
	}
}

// getBuiltinModuleCall returns the `process.getBuiltinModule(...)` CallExpression
// for which `procId` is the `process` reference, or nil.
func getBuiltinModuleCall(procId *ast.Node) *ast.Node {
	node := procId
	for isPassThrough(node) {
		node = node.Parent
	}
	pa := node.Parent
	if pa == nil || pa.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	paExpr := pa.AsPropertyAccessExpression()
	if paExpr.Expression != node {
		return nil
	}
	name := paExpr.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "getBuiltinModule" {
		return nil
	}
	return callFromCallee(pa)
}

var NoDeprecatedApiRule = rule.Rule{
	Name: "n/no-deprecated-api",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		version := getConfiguredNodeVersion(ctx, opts)

		// no-deprecated-api needs whole-file analysis (upstream runs on
		// Program:exit). rslint doesn't visit the SourceFile root node, but
		// ctx (SourceFile + ReportNode) is fully bound when Run is called, so
		// the analysis runs here directly and no listeners are returned.
		tracker := newReferenceTracker(ctx)

		globalResults := tracker.capture(func() {
			tracker.iterateGlobalReferences(globals)
		})
		moduleResults := tracker.capture(func() {
			tracker.iterateCjsReferences(modules)
			tracker.iterateProcessGetBuiltinModuleReferences(modules)
			tracker.iterateEsmReferences(modules)
		})

		// Collect both passes, then report in source order. `globals` / `modules`
		// are Go maps (random iteration order), so reporting as we iterate would
		// make multi-finding files non-deterministic. Sorting by position also
		// matches ESLint, which orders diagnostics by location before output.
		type pendingReport struct {
			node *ast.Node
			name string
			info *deprecatedInfo
		}
		var reports []pendingReport
		for _, r := range globalResults {
			name := toName(r.typ, r.path)
			if !opts.ignoreGlobalItems[name] {
				reports = append(reports, pendingReport{node: r.node, name: "'" + name + "'", info: r.info})
			}
		}
		for _, r := range moduleResults {
			name := toName(r.typ, r.path)
			suffix := ""
			if len(r.path) == 1 {
				suffix = " module"
			}
			if !opts.ignoreModuleItems[name] {
				reports = append(reports, pendingReport{node: r.node, name: "'" + name + "'" + suffix, info: r.info})
			}
		}
		sort.SliceStable(reports, func(i, j int) bool {
			return reports[i].node.Pos() < reports[j].node.Pos()
		})
		for _, p := range reports {
			reportItem(ctx, version, p.node, p.name, p.info)
		}

		return rule.RuleListeners{}
	},
}
