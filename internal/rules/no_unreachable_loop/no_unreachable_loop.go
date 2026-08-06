package no_unreachable_loop

import (
	_ "embed"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

//go:embed no_unreachable_loop.schema.json
var schemaJSON []byte

var invalidLoop = rule.RuleMessage{
	Id:          "invalid",
	Description: "Invalid loop. Its body allows only one iteration.",
}

// https://eslint.org/docs/latest/rules/no-unreachable-loop
var NoUnreachableLoopRule = rule.Rule{
	Name:   "no-unreachable-loop",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		if opts.ignored == allLoopKinds {
			// Match ESLint's empty-selector fast path when every loop kind is
			// ignored: the rule has nothing to collect or report.
			return nil
		}

		collector := loopRootCollector{options: opts}
		collect := collector.collectLoop

		var listeners rule.RuleListeners
		if opts.ignored == 0 {
			listeners = rule.RuleListeners{
				ast.KindWhileStatement: collect,
				ast.KindDoStatement:    collect,
				ast.KindForStatement:   collect,
				ast.KindForInStatement: collect,
				ast.KindForOfStatement: collect,
			}
		} else {
			listeners = make(rule.RuleListeners, len(loopKinds)+1)
			for _, loop := range loopKinds {
				if opts.ignored&loop.bit == 0 {
					listeners[loop.kind] = collect
				}
			}
		}

		statements := ctx.SourceFile.Statements
		if statements == nil || len(statements.Nodes) == 0 {
			return listeners
		}
		lastTopLevelNode := statements.Nodes[len(statements.Nodes)-1]
		listeners[rule.ListenerOnExit(lastTopLevelNode.Kind)] = func(node *ast.Node) {
			if node == lastTopLevelNode {
				reportLoops(&ctx, collector.rootOrder, opts)
			}
		}
		return listeners
	},
}

// loopRootCollector deduplicates code path roots supplied by loop listeners.
type loopRootCollector struct {
	options   ruleOptions
	rootOrder []*ast.Node
	seen      map[*ast.Node]struct{}
}

func (collector *loopRootCollector) collectLoop(node *ast.Node) {
	if collector.options.ignores(node.Kind) {
		return
	}
	root := cfg.RootOf(node)
	if root == nil {
		return
	}
	if collector.seen == nil {
		if len(collector.rootOrder) == 0 {
			collector.rootOrder = append(collector.rootOrder, root)
			return
		}
		if collector.rootOrder[0] == root {
			return
		}
		collector.seen = map[*ast.Node]struct{}{collector.rootOrder[0]: {}}
	}
	if _, exists := collector.seen[root]; exists {
		return
	}
	collector.seen[root] = struct{}{}
	collector.rootOrder = append(collector.rootOrder, root)
}

// reportLoops lays out each collected code path and reports the loops nothing
// ever re-enters. ESLint collects across the whole file and reports at the end,
// so the reports are ordered by position rather than by code path.
func reportLoops(ctx *rule.RuleContext, rootOrder []*ast.Node, opts ruleOptions) {
	var reports []*ast.Node
	for _, root := range rootOrder {
		rootReports := singleIterationLoops(root, opts)
		if reports == nil {
			reports = rootReports
		} else {
			reports = append(reports, rootReports...)
		}
	}
	if len(reports) > 1 {
		slices.SortFunc(reports, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
	}
	for _, loop := range reports {
		ctx.ReportNode(loop, invalidLoop)
	}
}

// singleIterationLoops returns the loops in root whose body allows only one
// iteration: control reaches the loop, but never flows back into it for a
// second one. Unreachable blocks are skipped — whether a loop that never runs
// could repeat says nothing useful, and neither does a `continue` that never
// runs.
//
// NOTE: A loop wrapped in more than one label repeats when a `continue` names
// any of them. ESLint 10.8.0 reports the double-labelled `while` and `do` forms
// when the outer label is targeted, although it accepts the analogous `for`
// forms and three-label cases. internal/utils/cfg consistently keeps the edge.
//
// NOTE: Unlike ESLint, a `continue` in a `finally` block repeats the enclosing
// loop even when the `try` it belongs to can only be left abruptly. That
// `continue` overrides the pending `return` or `throw`, so `for (;;) { try {
// return; } finally { continue; } }` really does start a second iteration.
// ESLint 10.8.0 still reports some of these shapes depending on the surrounding
// code path; internal/utils/cfg keeps the edge consistently.
func singleIterationLoops(root *ast.Node, opts ruleOptions) []*ast.Node {
	if opts.ignored == 0 {
		return singleIterationLoopsDefault(root)
	}

	var candidates []*ast.Node
	repeats := make(map[*ast.Node]bool)
	cfg.Build(root, cfg.Hooks[struct{}]{
		Statement: func(b *cfg.Builder[struct{}], node *ast.Node) {
			if !b.Current().Reachable {
				return
			}
			bit := loopBitForKind(node.Kind)
			if bit == 0 || opts.ignored&bit != 0 {
				return
			}
			if _, known := repeats[node]; !known {
				// A loop inside a `finally` block is laid out twice; the
				// candidate list holds it once.
				candidates = append(candidates, node)
				repeats[node] = false
			}
		},
		Loop: func(b *cfg.Builder[struct{}], loop *ast.Node) {
			// Ignored loops are not candidates, so retaining their repeat edge
			// is harmless and avoids another option lookup on this hot hook.
			if !b.Current().Reachable {
				return
			}
			repeats[loop] = true
		},
	})
	return slices.DeleteFunc(candidates, func(loop *ast.Node) bool { return repeats[loop] })
}

func singleIterationLoopsDefault(root *ast.Node) []*ast.Node {
	var candidates []*ast.Node
	repeats := make(map[*ast.Node]bool)
	cfg.Build(root, cfg.Hooks[struct{}]{
		Statement: func(b *cfg.Builder[struct{}], node *ast.Node) {
			if !b.Current().Reachable || !isLoop(node) {
				return
			}
			if _, known := repeats[node]; !known {
				// A loop inside a `finally` block is laid out twice; the
				// candidate list holds it once.
				candidates = append(candidates, node)
				repeats[node] = false
			}
		},
		Loop: func(b *cfg.Builder[struct{}], loop *ast.Node) {
			if !b.Current().Reachable {
				return
			}
			repeats[loop] = true
		},
	})
	return slices.DeleteFunc(candidates, func(loop *ast.Node) bool { return repeats[loop] })
}

// isLoop is used for every reachable statement in the default CFG hot path.
// Keep the direct kind switch here; it is cheaper than the general tsgo helper
// in this callback.
func isLoop(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement:
		return true
	default:
		return false
	}
}

type loopKindSet uint8

const (
	loopKindWhile loopKindSet = 1 << iota
	loopKindDoWhile
	loopKindFor
	loopKindForIn
	loopKindForOf
	allLoopKinds = loopKindWhile | loopKindDoWhile | loopKindFor | loopKindForIn | loopKindForOf
)

// loopKinds maps the names the `ignore` option uses — ESLint's ESTree node
// types — onto the syntax kinds they correspond to. The bitset keeps option
// checks allocation-free because a rule's options are parsed once per file.
var loopKinds = [...]struct {
	name string
	kind ast.Kind
	bit  loopKindSet
}{
	{name: "WhileStatement", kind: ast.KindWhileStatement, bit: loopKindWhile},
	{name: "DoWhileStatement", kind: ast.KindDoStatement, bit: loopKindDoWhile},
	{name: "ForStatement", kind: ast.KindForStatement, bit: loopKindFor},
	{name: "ForInStatement", kind: ast.KindForInStatement, bit: loopKindForIn},
	{name: "ForOfStatement", kind: ast.KindForOfStatement, bit: loopKindForOf},
}

type ruleOptions struct {
	ignored loopKindSet
}

func (opts ruleOptions) ignores(kind ast.Kind) bool {
	if opts.ignored == 0 {
		return false
	}
	return opts.ignored&loopBitForKind(kind) != 0
}

func loopBitForKind(kind ast.Kind) loopKindSet {
	switch kind {
	case ast.KindWhileStatement:
		return loopKindWhile
	case ast.KindDoStatement:
		return loopKindDoWhile
	case ast.KindForStatement:
		return loopKindFor
	case ast.KindForInStatement:
		return loopKindForIn
	case ast.KindForOfStatement:
		return loopKindForOf
	default:
		return 0
	}
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	ignore, _ := optsMap["ignore"].([]interface{})
	for _, item := range ignore {
		name, _ := item.(string)
		for _, loop := range loopKinds {
			if loop.name == name {
				opts.ignored |= loop.bit
				break
			}
		}
	}
	return opts
}
