package constructor_super

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Each bit represents paths with an actual successful super() call, or without
// one. Zero means unreachable. A value return is a completion, not a call.
type paths uint8

const (
	uncalled paths = 1
	called   paths = 2
)

type jump struct {
	target     *ast.Node
	continuing bool
}
type flow struct {
	normal, returned, returnedValue, thrown paths
	jumps                                   map[jump]paths
}

func (f *flow) merge(g flow) {
	f.normal |= g.normal
	f.returned |= g.returned
	f.returnedValue |= g.returnedValue
	f.thrown |= g.thrown
	for key, s := range g.jumps {
		f.addJump(key, s)
	}
}
func (f *flow) addJump(key jump, s paths) {
	if s == 0 {
		return
	}
	if f.jumps == nil {
		f.jumps = make(map[jump]paths)
	}
	f.jumps[key] |= s
}
func (f *flow) takeJump(key jump) paths { s := f.jumps[key]; delete(f.jumps, key); return s }
func (f flow) clone() flow {
	out := f
	out.jumps = nil
	for k, s := range f.jumps {
		out.addJump(k, s)
	}
	return out
}
func (f flow) entries() paths {
	s := f.normal | f.returned | f.returnedValue | f.thrown
	for _, v := range f.jumps {
		s |= v
	}
	return s
}

type statementKey struct {
	node  *ast.Node
	state paths
}
type statementResult struct {
	output     flow
	exceptions paths
}
type analyzer struct {
	possible, instanceElements bool
	reports                    map[*ast.Node]rule.RuleMessage
	statementsCache            map[statementKey]statementResult
	truth                      *utils.StaticStringEvaluator
	exceptions                 *paths
}

func (a *analyzer) mayThrow(s paths) {
	if a.exceptions != nil {
		*a.exceptions |= s
	}
}
func (a *analyzer) capture(n *ast.Node, s paths) (flow, paths) {
	saved := a.exceptions
	exceptions := paths(0)
	a.exceptions = &exceptions
	result := a.statement(n, s)
	a.exceptions = saved
	return result, exceptions
}
func (a *analyzer) captureBinding(n *ast.Node, s paths) (paths, paths) {
	saved := a.exceptions
	exceptions := paths(0)
	a.exceptions = &exceptions
	s = a.binding(n, s, true)
	a.exceptions = saved
	return s, exceptions
}
func (a *analyzer) reportCall(n *ast.Node, s paths) {
	var message rule.RuleMessage
	if s&called != 0 {
		// A call initially reported as invalid is not reconsidered as its
		// own loop duplicate under the pinned diagnostic precedence.
		if a.reports[n].Id == "badSuper" {
			return
		}
		message = buildDuplicate()
	} else {
		if a.possible || a.reports[n].Id == "duplicate" {
			return
		}
		message = buildBadSuper()
	}
	if a.reports == nil {
		a.reports = make(map[*ast.Node]rule.RuleMessage)
	}
	a.reports[n] = message
}

// Leaves transfer a scalar normal state without heap allocations. Compound
// nodes memoize their three possible inputs, bounding nested-loop analysis.
// Cached jump sets are copied so consuming a completion cannot change a memo.
func (a *analyzer) statement(n *ast.Node, s paths) flow {
	if n == nil || s == 0 {
		return flow{normal: s}
	}
	// A constructor's outer body is entered once and has no pending caller
	// completion. Its nested loops/finally blocks keep their own memoization.
	if n.Kind == ast.KindBlock && n.Parent != nil && n.Parent.Kind == ast.KindConstructor {
		return a.statementUncached(n, s)
	}
	switch n.Kind {
	case ast.KindBlock, ast.KindIfStatement, ast.KindSwitchStatement, ast.KindTryStatement,
		ast.KindLabeledStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement:
	default:
		return a.statementUncached(n, s)
	}
	key := statementKey{n, s}
	if result, ok := a.statementsCache[key]; ok {
		a.mayThrow(result.exceptions)
		return result.output.clone()
	}
	saved := a.exceptions
	exceptions := paths(0)
	a.exceptions = &exceptions
	out := a.statementUncached(n, s)
	a.exceptions = saved
	a.mayThrow(exceptions)
	if a.statementsCache == nil {
		a.statementsCache = make(map[statementKey]statementResult)
	}
	a.statementsCache[key] = statementResult{out.clone(), exceptions}
	return out
}
