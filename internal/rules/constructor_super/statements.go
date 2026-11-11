package constructor_super

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (a *analyzer) statements(nodes []*ast.Node, s paths) flow {
	out := flow{normal: s}
	for _, n := range nodes {
		if out.normal == 0 {
			break
		}
		s = out.normal
		out.normal = 0
		out.merge(a.statement(n, s))
	}
	return out
}
func (a *analyzer) statementUncached(n *ast.Node, s paths) flow {
	switch n.Kind {
	case ast.KindBlock:
		return a.statements(n.Statements(), s)
	case ast.KindIfStatement:
		i := n.AsIfStatement()
		yes, no := a.condition(i.Expression, s, false)
		out := a.statement(i.ThenStatement, yes)
		out.merge(a.statement(i.ElseStatement, no))
		return out
	case ast.KindReturnStatement:
		if n.Expression() != nil {
			return flow{returnedValue: a.expression(n.Expression(), s)}
		}
		return flow{returned: s}
	case ast.KindThrowStatement:
		return flow{thrown: a.expression(n.Expression(), s)}
	case ast.KindBreakStatement, ast.KindContinueStatement:
		out := flow{}
		out.addJump(resolveJump(n), s)
		return out
	case ast.KindLabeledStatement:
		l := n.AsLabeledStatement()
		out := a.statement(l.Statement, s)
		out.normal |= out.takeJump(jump{target: labelTarget(l.Statement)})
		return out
	case ast.KindWhileStatement:
		w := n.AsWhileStatement()
		return a.loop(n, w.Statement, w.Expression, nil, s)
	case ast.KindDoStatement:
		d := n.AsDoStatement()
		return a.loop(n, d.Statement, d.Expression, nil, s)
	case ast.KindForStatement:
		f := n.AsForStatement()
		s = a.expression(f.Initializer, s)
		return a.loop(n, f.Statement, f.Condition, f.Incrementor, s)
	case ast.KindForInStatement, ast.KindForOfStatement:
		f := n.AsForInOrOfStatement()
		s = a.expression(f.Expression, s)
		// Iterator/enumeration setup can invoke user code before the first binding.
		a.mayThrow(s)
		return a.loop(n, f.Statement, nil, nil, s)
	case ast.KindSwitchStatement:
		return a.switchStatement(n, s)
	case ast.KindTryStatement:
		return a.tryStatement(n, s)
	default:
		return flow{normal: a.expression(n, s)}
	}
}
func labelTarget(n *ast.Node) *ast.Node {
	for n != nil && n.Kind == ast.KindLabeledStatement {
		n = n.AsLabeledStatement().Statement
	}
	return n
}
func isLoop(n *ast.Node) bool {
	return n != nil && (n.Kind == ast.KindWhileStatement || n.Kind == ast.KindDoStatement || n.Kind == ast.KindForStatement || n.Kind == ast.KindForInStatement || n.Kind == ast.KindForOfStatement)
}
func resolveJump(n *ast.Node) jump {
	continuing := n.Kind == ast.KindContinueStatement
	label := n.Label()
	for p := n.Parent; p != nil; p = p.Parent {
		if utils.IsFunctionLikeContainer(p) {
			break
		}
		if label != nil {
			if p.Kind == ast.KindLabeledStatement && p.AsLabeledStatement().Label.Text() == label.Text() {
				target := labelTarget(p.AsLabeledStatement().Statement)
				if !continuing || isLoop(target) {
					return jump{target, continuing}
				}
				break
			}
		} else if isLoop(p) || !continuing && p.Kind == ast.KindSwitchStatement {
			return jump{p, continuing}
		}
	}
	return jump{continuing: continuing}
}
func (a *analyzer) loop(n, body, test, update *ast.Node, s paths) flow {
	out := flow{}
	head := s
	do := n.Kind == ast.KindDoStatement
	enumerating := n.Kind == ast.KindForInStatement || n.Kind == ast.KindForOfStatement
	for {
		entry := head
		if !do {
			if enumerating {
				out.normal |= entry
			} else {
				yes, no := a.condition(test, entry, true)
				out.normal |= no
				entry = yes
			}
		}
		if entry == 0 {
			break
		}
		if enumerating {
			a.mayThrow(entry)
			init := n.AsForInOrOfStatement().Initializer
			if init != nil && init.Kind == ast.KindVariableDeclarationList {
				for _, decl := range init.AsVariableDeclarationList().Declarations.Nodes {
					entry = a.binding(decl.Name(), entry, true)
				}
			} else {
				entry = a.binding(init, entry, false)
			}
		}
		f := a.statement(body, entry)
		back := f.normal | f.takeJump(jump{n, true})
		f.normal = 0
		out.normal |= f.takeJump(jump{n, false})
		out.merge(f)
		back = a.expression(update, back)
		if do {
			yes, no := a.condition(test, back, true)
			out.normal |= no
			back = yes
		}
		next := head | back
		if next == head {
			break
		}
		head = next
	}
	return out
}
func (a *analyzer) switchStatement(n *ast.Node, s paths) flow {
	sw := n.AsSwitchStatement()
	s = a.expression(sw.Expression, s)
	if sw.CaseBlock == nil {
		return flow{normal: s}
	}
	clauses := sw.CaseBlock.AsCaseBlock().Clauses.Nodes
	entries := make([]paths, len(clauses))
	defaultIndex := -1
	// Selection and execution have different orders when default is not last.
	for i, c := range clauses {
		if c.Kind == ast.KindDefaultClause {
			defaultIndex = i
			continue
		}
		s = a.expression(c.AsCaseOrDefaultClause().Expression, s)
		entries[i] = s
	}
	out := flow{}
	if defaultIndex >= 0 {
		entries[defaultIndex] = s
	} else {
		out.normal = s
	}
	fall := paths(0)
	for i, c := range clauses {
		f := a.statements(c.AsCaseOrDefaultClause().Statements.Nodes, entries[i]|fall)
		fall = f.normal
		f.normal = 0
		out.normal |= f.takeJump(jump{target: n})
		out.merge(f)
	}
	out.normal |= fall
	return out
}
func (a *analyzer) tryStatement(n *ast.Node, s paths) flow {
	t := n.AsTryStatement()
	out, implicit := a.capture(t.TryBlock, s)
	thrown := out.thrown | implicit
	out.thrown = 0
	if t.CatchClause != nil {
		c := t.CatchClause.AsCatchClause()
		entry, bindThrow := thrown, paths(0)
		if c.VariableDeclaration != nil {
			entry, bindThrow = a.captureBinding(c.VariableDeclaration.Name(), entry)
		}
		caught, catchThrow := a.capture(c.Block, entry)
		caught.thrown |= bindThrow | catchThrow
		out.merge(caught)
	} else {
		out.thrown = thrown
	}
	if t.FinallyBlock == nil {
		return out
	}
	// Diagnose once against all entries, then preserve each pending completion
	// only when finally completes normally. Its own completion replaces it.
	a.capture(t.FinallyBlock, out.entries())
	result, finThrow := a.capture(t.FinallyBlock, out.normal)
	result.thrown |= finThrow
	apply := func(state paths, resume func(paths)) {
		if state == 0 {
			return
		}
		fin, throws := a.capture(t.FinallyBlock, state)
		resume(fin.normal)
		fin.normal = 0
		fin.thrown |= throws
		result.merge(fin)
	}
	apply(out.returned, func(s paths) { result.returned |= s })
	apply(out.returnedValue, func(s paths) { result.returnedValue |= s })
	apply(out.thrown, func(s paths) { result.thrown |= s })
	for key, state := range out.jumps {
		apply(state, func(s paths) { result.addJump(key, s) })
	}
	return result
}
