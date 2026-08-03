package no_unreachable_loop

import (
	_ "embed"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/cfg"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		opts := parseOptions(rule.LegacyUnwrapOptions(_options))

		// Only the code paths that contain a loop worth checking need a graph.
		var rootOrder []*ast.Node
		seen := make(map[*ast.Node]bool)
		collect := func(node *ast.Node) {
			if slices.Contains(opts.ignore, node.Kind) {
				return
			}
			root := cfg.RootOf(node)
			if root == nil || seen[root] {
				return
			}
			seen[root] = true
			rootOrder = append(rootOrder, root)
		}

		listeners := rule.RuleListeners{
			ast.KindWhileStatement: collect,
			ast.KindDoStatement:    collect,
			ast.KindForStatement:   collect,
			ast.KindForInStatement: collect,
			ast.KindForOfStatement: collect,
		}
		statements := ctx.SourceFile.Statements
		if statements == nil || len(statements.Nodes) == 0 {
			return listeners
		}
		lastTopLevelNode := statements.Nodes[len(statements.Nodes)-1]
		listeners[rule.ListenerOnExit(lastTopLevelNode.Kind)] = func(node *ast.Node) {
			if node == lastTopLevelNode {
				reportLoops(&ctx, rootOrder, opts)
			}
		}
		return listeners
	},
}

// reportLoops lays out each collected code path and reports the loops nothing
// ever re-enters. ESLint collects across the whole file and reports at the end,
// so the reports are ordered by position rather than by code path.
func reportLoops(ctx *rule.RuleContext, rootOrder []*ast.Node, opts options) {
	var reports []*ast.Node
	for _, root := range rootOrder {
		reports = append(reports, singleIterationLoops(root, opts)...)
	}
	slices.SortFunc(reports, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
	for _, loop := range reports {
		ctx.ReportNode(loop, invalidLoop)
	}
}

// singleIterationLoops returns the loops in root whose body allows only one
// iteration: control reaches the loop, but never flows back into it for a
// second one. A loop the code path never reaches is left alone — whether its
// body could repeat says nothing useful about code that does not run.
//
// NOTE: Unlike ESLint, a loop wrapped in more than one label repeats when a
// `continue` names any of them. ESLint attaches only the innermost label to the
// loop, so `outer: inner: while (foo) { continue outer; }` looks to it like a
// loop nothing re-enters; internal/cfg keeps the edge, so the loop is accepted.
func singleIterationLoops(root *ast.Node, opts options) []*ast.Node {
	var candidates []*ast.Node
	repeats := make(map[*ast.Node]bool)
	cfg.Build(root, cfg.Hooks[struct{}]{
		Statement: func(_ *cfg.Builder[struct{}], node *ast.Node) {
			if !isLoop(node) || slices.Contains(opts.ignore, node.Kind) {
				return
			}
			if _, known := repeats[node]; !known {
				// A loop inside a `finally` block is laid out twice; the
				// candidate list holds it once.
				candidates = append(candidates, node)
				repeats[node] = false
			}
		},
		Loop: func(_ *cfg.Builder[struct{}], loop *ast.Node) {
			repeats[loop] = true
		},
	})
	return slices.DeleteFunc(candidates, func(loop *ast.Node) bool { return repeats[loop] })
}

func isLoop(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}

// loopKinds maps the names the `ignore` option uses — ESLint's ESTree node
// types — onto the syntax kinds they correspond to.
var loopKinds = map[string]ast.Kind{
	"WhileStatement":   ast.KindWhileStatement,
	"DoWhileStatement": ast.KindDoStatement,
	"ForStatement":     ast.KindForStatement,
	"ForInStatement":   ast.KindForInStatement,
	"ForOfStatement":   ast.KindForOfStatement,
}

type options struct {
	ignore []ast.Kind
}

func parseOptions(raw any) options {
	opts := options{}
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	add := func(name string) {
		if kind, ok := loopKinds[name]; ok {
			opts.ignore = append(opts.ignore, kind)
		}
	}
	switch ignore := optsMap["ignore"].(type) {
	case []interface{}:
		for _, item := range ignore {
			if name, ok := item.(string); ok {
				add(name)
			}
		}
	case []string:
		for _, name := range ignore {
			add(name)
		}
	}
	return opts
}
