package no_cycle

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_cycle.schema.json
var schemaJSON []byte

const unlimitedDepth = int(^uint(0) >> 1)

type ruleOptions struct {
	maxDepth                           int
	ignoreExternal                     bool
	allowUnsafeDynamicCyclicDependency bool
	syntax                             rule.ModuleSyntax
}

type routeStep struct {
	value string
	line  int
}

// queuedModule is one breadth-first search entry. It keeps a parent link into
// the queue instead of a copy of the route so far, and names the reference it
// arrived through, so a route is materialized only for a path that closes a
// cycle.
type queuedModule struct {
	node    int32
	parent  int32
	viaNode int32
	viaRef  int32
	depth   int32
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
		maxDepth: unlimitedDepth,
		syntax: rule.ModuleSyntax{
			ESModule: true,
		},
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
	opts.syntax.CommonJS, _ = optsMap["commonjs"].(bool)
	opts.syntax.AMD, _ = optsMap["amd"].(bool)
	if esmodule, ok := optsMap["esmodule"].(bool); ok {
		opts.syntax.ESModule = esmodule
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
	if ctx.SourceFile == nil || ctx.Program == nil {
		return
	}

	myPath := ctx.SourceFile.FileName()
	if myPath == "" || myPath == "<text>" {
		return
	}

	graph := moduleGraphFor(ctx, opts)
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
	stayInGroup := !opts.allowUnsafeDynamicCyclicDependency
	traversed := make(map[int32]bool)
	for r := range node.refs {
		if !graph.isSearchable(opts, self, r) {
			continue
		}
		// A search confined to the group only ever rules out references whose
		// target is in it, so the ones outside can be skipped outright. The
		// unconfined search shares its traversal set with every reference, and
		// skipping any of them would change what the later ones see.
		if stayInGroup && graph.group[node.edge[r]] != graph.group[self] {
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

// isSearchable reports whether reference r of file self is one the rule
// searches from: an edge this configuration keeps, pointing at another file.
// Direct self imports belong to import/no-self-import, and a dynamic import
// is not a cycle when the configuration says so.
func (graph *moduleGraph) isSearchable(opts ruleOptions, self int32, r int) bool {
	node := &graph.nodes[self]
	target := node.edge[r]
	if target < 0 || target == self {
		return false
	}
	return !opts.allowUnsafeDynamicCyclicDependency || !node.refs[r].Dynamic()
}

// hasCyclicCandidate reports whether any reference of self could be reported.
// A reported reference has a route from its target back to self, and self
// imports the target, so the two are mutually reachable and share a group.
func (graph *moduleGraph) hasCyclicCandidate(opts ruleOptions, self int32) bool {
	node := &graph.nodes[self]
	for r := range node.refs {
		if graph.isSearchable(opts, self, r) && graph.group[node.edge[r]] == graph.group[self] {
			return true
		}
	}
	return false
}

// detectCycle walks breadth-first from start looking for a way back to self
// and returns the route it arrived by. traversed is shared across one file's
// references, so a target an earlier reference already ruled out is not
// searched again.
//
// The walk stays inside self's strongly connected group: every file on a route
// back to self both reaches self and is reached from it, which is what that
// group means, so a file outside it can neither lie on a route nor lead to
// one. Withholding dynamic edges breaks that correspondence, because the
// group numbers still describe the edges the search no longer follows, so
// such a configuration searches the whole graph instead.
func (graph *moduleGraph) detectCycle(opts ruleOptions, self int32, traversed map[int32]bool, start int32) ([]routeStep, bool) {
	group := graph.group[self]
	stayInGroup := !opts.allowUnsafeDynamicCyclicDependency

	queue := []queuedModule{{node: start, parent: -1, viaNode: -1, viaRef: -1}}
	for head := 0; head < len(queue); head++ {
		next := queue[head]
		if traversed[next.node] {
			continue
		}
		traversed[next.node] = true

		for r, target := range graph.nodes[next.node].expand {
			if target < 0 || traversed[target] {
				continue
			}
			if target == self {
				return graph.routeTo(queue, head), true
			}
			if int(next.depth)+1 >= opts.maxDepth {
				continue
			}
			if stayInGroup && graph.group[target] != group {
				continue
			}
			queue = append(queue, queuedModule{
				node:    target,
				parent:  int32(head),
				viaNode: next.node,
				viaRef:  int32(r),
				depth:   next.depth + 1,
			})
		}
	}

	return nil, false
}

// routeTo follows the parent links back from a queue entry, materializing the
// specifier and source line of every reference the search passed through.
func (graph *moduleGraph) routeTo(queue []queuedModule, index int) []routeStep {
	depth := int(queue[index].depth)
	route := make([]routeStep, depth)
	for routeIndex := depth - 1; routeIndex >= 0; routeIndex-- {
		next := queue[index]
		edge := &graph.nodes[next.viaNode].refs[next.viaRef]
		route[routeIndex] = routeStep{
			value: edge.Text(),
			line:  sourceLine(edge.From, edge.Specifier),
		}
		index = int(next.parent)
	}
	return route
}

func sourceLine(sourceFile *ast.SourceFile, source *ast.Node) int {
	if sourceFile == nil || source == nil {
		return 1
	}
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, source.Pos())
	return line + 1
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
