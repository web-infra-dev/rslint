package no_cycle

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// The dependency graph this rule reads, and every question asked of it, live
// in graph.go. What is left here is the rule itself: its options, the per-file
// entry point, and the message it reports.

//go:embed no_cycle.schema.json
var schemaJSON []byte

const unlimitedDepth = int(^uint(0) >> 1)

type ruleOptions struct {
	maxDepth                           int
	ignoreExternal                     bool
	allowUnsafeDynamicCyclicDependency bool
	referenceKinds                     program.ModuleReferenceKinds
}

// NoCycleRule forbids dependency paths that resolve back to the linted module.
//
// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/no-cycle.js
var NoCycleRule = rule.Rule{
	Name:   "import/no-cycle",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		checkSourceFile(ctx, opts)
		return rule.RuleListeners{}
	},
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{
		maxDepth:       unlimitedDepth,
		referenceKinds: program.ESModuleReferences,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})

	if maxDepth, ok := parseMaxDepth(optsMap["maxDepth"]); ok {
		opts.maxDepth = maxDepth
	}
	opts.ignoreExternal, _ = optsMap["ignoreExternal"].(bool)
	opts.allowUnsafeDynamicCyclicDependency, _ = optsMap["allowUnsafeDynamicCyclicDependency"].(bool)
	if commonJS, _ := optsMap["commonjs"].(bool); commonJS {
		opts.referenceKinds |= program.CommonJSReferences
	}
	if amd, _ := optsMap["amd"].(bool); amd {
		opts.referenceKinds |= program.AMDReferences
	}
	if esmodule, ok := optsMap["esmodule"].(bool); ok {
		if esmodule {
			opts.referenceKinds |= program.ESModuleReferences
		} else {
			opts.referenceKinds &^= program.ESModuleReferences
		}
	}

	return opts
}

func parseMaxDepth(raw any) (int, bool) {
	switch value := raw.(type) {
	case nil:
		return 0, false
	case int:
		return normalizeDepth(value)
	case int64:
		return normalizeDepth(int(value))
	case float64:
		if math.IsInf(value, 1) {
			return unlimitedDepth, true
		}
		if math.Trunc(value) != value {
			return 0, false
		}
		return normalizeDepth(int(value))
	case float32:
		if math.IsInf(float64(value), 1) {
			return unlimitedDepth, true
		}
		if math.Trunc(float64(value)) != float64(value) {
			return 0, false
		}
		return normalizeDepth(int(value))
	case json.Number:
		if i, err := value.Int64(); err == nil {
			return normalizeDepth(int(i))
		}
		if f, err := value.Float64(); err == nil {
			return normalizeDepth(int(f))
		}
	case string:
		if value == "∞" {
			return unlimitedDepth, true
		}
	}
	return 0, false
}

func normalizeDepth(depth int) (int, bool) {
	if depth < 1 {
		return 0, false
	}
	return depth, true
}

func checkSourceFile(ctx rule.RuleContext, opts ruleOptions) {
	if ctx.SourceFile == nil || !ctx.Program().IsValid() {
		return
	}

	myPath := ctx.SourceFile.FileName()
	if myPath == "" || myPath == "<text>" {
		return
	}

	// Every report sits on a reference this file wrote, so a file that wrote
	// none has no answer to look up. Asking that of the file alone, before the
	// graph, is what keeps an import-free file off the build: the graph spans
	// the whole effective source set, and the editor discards it on every keystroke, so
	// there is no run to amortize it over on that path.
	sourceGraph := ctx.Program().ModuleGraph()
	if len(sourceGraph.References(ctx.SourceFile, opts.referenceKinds)) == 0 {
		return
	}

	graph := moduleGraphFor(ctx, sourceGraph, opts)
	if graph == nil {
		return
	}
	self, ok := graph.index[ctx.SourceFile]
	if !ok {
		return
	}

	// A reference is only reportable when its target reaches this file again,
	// which puts the two in one strongly connected group. Files holding no
	// such reference — nearly all of them in a healthy project — are answered
	// from the group numbers alone, without walking the graph.
	if !graph.hasCyclicCandidate(opts, self) {
		return
	}

	node := &graph.nodes[self]
	traversed := make(map[int32]bool)
	for r := range node.refs {
		if !graph.isReportCandidate(opts, self, r) {
			continue
		}
		route, found := graph.detectCycle(opts, self, traversed, node.edge[r])
		if !found {
			continue
		}

		reportNode := node.refs[r].Declaration
		if reportNode == nil {
			reportNode = node.refs[r].Specifier
		}
		ctx.ReportNode(reportNode, messageCycle(route))
	}
}

func messageCycle(route []routeStep) rule.RuleMessage {
	description := "Dependency cycle detected."
	if len(route) > 0 {
		parts := make([]string, 0, len(route))
		for _, step := range route {
			parts = append(parts, fmt.Sprintf("%s:%d", step.value, step.line))
		}
		description = "Dependency cycle via " + strings.Join(parts, "=>")
	}
	return rule.RuleMessage{
		Id:          "cycle",
		Description: description,
	}
}
