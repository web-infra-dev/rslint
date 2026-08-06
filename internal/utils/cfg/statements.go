package cfg

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
)

func (b *Builder[E]) statements(list *ast.NodeList) {
	if list == nil {
		return
	}
	for _, stmt := range list.Nodes {
		b.statement(stmt)
	}
}

func (b *Builder[E]) statement(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}
	b.reachedStatement(node)

	switch node.Kind {
	case ast.KindBlock:
		b.statements(node.AsBlock().Statements)

	case ast.KindVariableStatement:
		b.variableDeclarationList(node.AsVariableStatement().DeclarationList)

	case ast.KindExpressionStatement:
		b.expr(node.AsExpressionStatement().Expression)

	case ast.KindIfStatement:
		b.ifStatement(node)

	case ast.KindWhileStatement:
		b.whileStatement(node)

	case ast.KindDoStatement:
		b.doStatement(node)

	case ast.KindForStatement:
		b.forStatement(node)

	case ast.KindForInStatement, ast.KindForOfStatement:
		b.forInOfStatement(node)

	case ast.KindSwitchStatement:
		b.switchStatement(node)

	case ast.KindTryStatement:
		b.tryStatement(node)

	case ast.KindLabeledStatement:
		b.labeledStatement(node)

	case ast.KindReturnStatement:
		b.expr(node.AsReturnStatement().Expression)
		b.makeReturn()

	case ast.KindThrowStatement:
		b.expr(node.AsThrowStatement().Expression)
		b.makeThrow()

	case ast.KindBreakStatement:
		b.makeBreak(node.AsBreakStatement().Label)

	case ast.KindContinueStatement:
		b.makeContinue(node.AsContinueStatement().Label)

	case ast.KindWithStatement:
		with := node.AsWithStatement()
		b.expr(with.Expression)
		b.statement(with.Statement)

	case ast.KindEmptyStatement, ast.KindDebuggerStatement,
		ast.KindImportDeclaration, ast.KindImportEqualsDeclaration,
		ast.KindFunctionDeclaration:
		// Nothing executes here in the enclosing code path. A function
		// declaration's parameters and body are its own code path root.

	case ast.KindClassDeclaration:
		b.classLike(node)

	case ast.KindExportAssignment:
		b.expr(node.AsExportAssignment().Expression)

	case ast.KindExportDeclaration:
		// Export bindings are resolved statically; nothing is evaluated.

	default:
		node.ForEachChild(func(child *ast.Node) bool {
			b.visitUnknown(child)
			return false
		})
	}
}

func (b *Builder[E]) variableDeclarationList(list *ast.Node) {
	if list == nil {
		return
	}
	declList := list.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return
	}
	for _, decl := range declList.Declarations.Nodes {
		b.variableDeclaration(decl)
	}
}

func (b *Builder[E]) variableDeclaration(node *ast.Node) {
	if node == nil {
		return
	}
	decl := node.AsVariableDeclaration()
	if decl == nil {
		return
	}
	// A `typeof x` inside the annotation references the variable, so the
	// annotation is walked even though nothing in it runs.
	b.expr(decl.Type)
	name := decl.Name()
	if name != nil && (name.Kind == ast.KindObjectBindingPattern || name.Kind == ast.KindArrayBindingPattern) {
		// Run-time evaluation order: the initializer produces the value
		// first, then the pattern binds element by element.
		b.expr(decl.Initializer)
		b.patternBind(name)
		return
	}
	if decl.Initializer == nil {
		// `let x;` stores no value.
		b.patternReads(name)
		return
	}
	b.patternReads(name)
	b.expr(decl.Initializer)
	b.patternWrites(name)
}

func (b *Builder[E]) ifStatement(node *ast.Node) {
	stmt := node.AsIfStatement()
	b.expr(stmt.Expression)
	testEnd := b.cur
	after := b.newBlock()

	thenEntry := b.newBlock()
	b.link(testEnd, thenEntry)
	b.enter(thenEntry)
	b.statement(stmt.ThenStatement)
	b.link(b.cur, after)

	if stmt.ElseStatement != nil {
		elseEntry := b.newBlock()
		b.link(testEnd, elseEntry)
		b.enter(elseEntry)
		b.statement(stmt.ElseStatement)
		b.link(b.cur, after)
	} else {
		b.link(testEnd, after)
	}

	b.enter(after)
}

func (b *Builder[E]) whileStatement(node *ast.Node) {
	stmt := node.AsWhileStatement()
	test := b.newBlock()
	b.link(b.cur, test)
	b.enter(test)
	b.expr(stmt.Expression)
	testEnd := b.cur

	after := b.newBlock()
	if !isAlwaysTruthyTest(stmt.Expression) {
		b.link(testEnd, after)
	}
	body := b.newBlock()
	b.link(testEnd, body)

	b.pushJump(node, after, test, node)
	b.enter(body)
	b.statement(stmt.Statement)
	b.loop(node)
	b.link(b.cur, test)
	b.popJump()

	b.enter(after)
}

func (b *Builder[E]) doStatement(node *ast.Node) {
	stmt := node.AsDoStatement()
	body := b.newBlock()
	b.link(b.cur, body)
	test := b.newBlock()
	after := b.newBlock()

	// A `continue` restarts a `do`…`while` at its test, not at its body, so it
	// is not what starts another iteration — the test passing is.
	b.pushJump(node, after, test, nil)
	b.enter(body)
	b.statement(stmt.Statement)
	b.link(b.cur, test)
	b.popJump()

	b.enter(test)
	b.expr(stmt.Expression)
	if b.cur != nil {
		b.loop(node)
		b.link(b.cur, body)
		if !isAlwaysTruthyTest(stmt.Expression) {
			b.link(b.cur, after)
		}
	}

	b.enter(after)
}

func (b *Builder[E]) forStatement(node *ast.Node) {
	stmt := node.AsForStatement()
	if stmt.Initializer != nil {
		if stmt.Initializer.Kind == ast.KindVariableDeclarationList {
			b.variableDeclarationList(stmt.Initializer)
		} else {
			b.expr(stmt.Initializer)
		}
	}

	test := b.newBlock()
	b.link(b.cur, test)
	b.enter(test)
	if stmt.Condition != nil {
		b.expr(stmt.Condition)
	}
	testEnd := b.cur

	after := b.newBlock()
	if stmt.Condition != nil && !isAlwaysTruthyTest(stmt.Condition) {
		b.link(testEnd, after)
	}
	body := b.newBlock()
	b.link(testEnd, body)

	update := test
	if stmt.Incrementor != nil {
		update = b.newBlock()
	}

	b.pushJump(node, after, update, node)
	b.enter(body)
	b.statement(stmt.Statement)
	b.loop(node)
	b.link(b.cur, update)
	b.popJump()

	if stmt.Incrementor != nil {
		b.enter(update)
		b.expr(stmt.Incrementor)
		b.link(b.cur, test)
	}

	b.enter(after)
}

func (b *Builder[E]) forInOfStatement(node *ast.Node) {
	stmt := node.AsForInOrOfStatement()
	b.expr(stmt.Expression)

	head := b.newBlock()
	b.link(b.cur, head)
	b.enter(head)

	after := b.newBlock()
	b.link(head, after)
	body := b.newBlock()
	b.link(head, body)

	b.pushJump(node, after, head, node)
	b.enter(body)
	if stmt.Initializer != nil {
		// The loop variable takes its value from the iteration, so it binds
		// with no expression of its own; its computed keys and destructuring
		// defaults evaluate once per iteration.
		if stmt.Initializer.Kind == ast.KindVariableDeclarationList {
			declList := stmt.Initializer.AsVariableDeclarationList()
			if declList != nil && declList.Declarations != nil {
				for _, decl := range declList.Declarations.Nodes {
					b.patternBind(decl.Name())
				}
			}
		} else {
			b.patternBind(stmt.Initializer)
		}
	}
	b.statement(stmt.Statement)
	b.loop(node)
	b.link(b.cur, head)
	b.popJump()

	b.enter(after)
}

func (b *Builder[E]) switchStatement(node *ast.Node) {
	stmt := node.AsSwitchStatement()
	b.expr(stmt.Expression)

	after := b.newBlock()
	var clauses []*ast.Node
	if stmt.CaseBlock != nil {
		if caseBlock := stmt.CaseBlock.AsCaseBlock(); caseBlock != nil && caseBlock.Clauses != nil {
			clauses = caseBlock.Clauses.Nodes
		}
	}
	if len(clauses) == 0 {
		b.link(b.cur, after)
		b.enter(after)
		return
	}

	bodies := make([]*Block[E], len(clauses))
	for i := range clauses {
		bodies[i] = b.newBlock()
	}

	defaultIndex := -1
	testCur := b.cur
	for i, clause := range clauses {
		if clause.Kind == ast.KindDefaultClause {
			defaultIndex = i
			continue
		}
		if testCur == nil {
			break
		}
		b.cur = testCur
		b.expr(clause.AsCaseOrDefaultClause().Expression)
		b.link(b.cur, bodies[i])
		if b.cur == nil {
			testCur = nil
			continue
		}
		nextTest := b.newBlock()
		b.link(b.cur, nextTest)
		testCur = nextTest
	}
	if testCur != nil {
		if defaultIndex >= 0 {
			b.link(testCur, bodies[defaultIndex])
		} else {
			b.link(testCur, after)
		}
	}

	b.pushJump(node, after, nil, nil)
	var fallthroughFrom *Block[E]
	for i, clause := range clauses {
		b.link(fallthroughFrom, bodies[i])
		b.enter(bodies[i])
		b.statements(clause.AsCaseOrDefaultClause().Statements)
		fallthroughFrom = b.cur
	}
	b.link(fallthroughFrom, after)
	b.popJump()

	b.enter(after)
}

func (b *Builder[E]) tryStatement(node *ast.Node) {
	stmt := node.AsTryStatement()
	hasCatch := stmt.CatchClause != nil
	hasFinally := stmt.FinallyBlock != nil

	frame := &tryFrame[E]{position: posTry, hasFinally: hasFinally}
	if hasCatch {
		frame.catchEntry = b.newBlock()
	}
	if hasFinally {
		frame.finallyEntry = b.newBlock()
	}
	b.tryStack = append(b.tryStack, frame)

	b.statement(stmt.TryBlock)

	var normalEnds []*Block[E]
	if hasCatch {
		if b.cur != nil {
			// ESLint routes the end of a `try` block into the `catch` block as
			// well as past the statement.
			b.link(b.cur, frame.catchEntry)
			normalEnds = append(normalEnds, b.cur)
		}
		frame.position = posCatch
		frame.thrownForked = false
		b.enter(frame.catchEntry)
		catchClause := stmt.CatchClause.AsCatchClause()
		if catchClause.VariableDeclaration != nil {
			b.patternBind(catchClause.VariableDeclaration.Name())
		}
		b.statement(catchClause.Block)
		if b.cur != nil {
			normalEnds = append(normalEnds, b.cur)
		}
	} else if b.cur != nil {
		normalEnds = append(normalEnds, b.cur)
	}

	b.tryStack = b.tryStack[:len(b.tryStack)-1]

	after := b.newBlock()
	if !hasFinally {
		for _, end := range normalEnds {
			b.link(end, after)
		}
		b.enter(after)
		return
	}

	snapshot := b.snapshotForks()
	if len(normalEnds) > 0 {
		normalEntry := b.newBlock()
		for _, end := range normalEnds {
			b.link(end, normalEntry)
		}
		b.enter(normalEntry)
		b.statement(stmt.FinallyBlock)
		b.link(b.cur, after)
	}

	if frame.finallyEntry.hasIncoming {
		// The `finally` block also runs on the path that leaves the statement
		// abruptly, so it is laid out a second time. Rewinding the enclosing
		// frames lets that copy fork at the same node the first copy did.
		b.restoreForks(snapshot)
		b.enter(frame.finallyEntry)
		b.statement(stmt.FinallyBlock)
		if b.cur != nil {
			// A `return` inside this `try`/`catch` still needs to run every
			// enclosing `finally` on its way out, not just this one — so this
			// second copy's own abrupt exit must keep propagating outward.
			if frame.returnedAny {
				if i := b.returnFrame(); i >= 0 {
					outer := b.tryStack[i]
					outer.returnedAny = true
					b.link(b.cur, outer.finallyEntry)
				}
			}
			if frame.thrownAny {
				if i := b.throwFrame(); i >= 0 {
					outer := b.tryStack[i]
					outer.thrownAny = true
					b.link(b.cur, throwTarget(outer))
				}
			}
		}
	}

	b.enter(after)
}

func (b *Builder[E]) labeledStatement(node *ast.Node) {
	stmt := node.AsLabeledStatement()
	if isBreakableStatement(stmt.Statement) {
		// The label belongs to the loop or switch itself; pushJump reads it
		// back off the parent chain.
		b.statement(stmt.Statement)
		return
	}
	after := b.newBlock()
	b.jumps = append(b.jumps, jumpTarget[E]{labels: []string{stmt.Label.Text()}, breakTo: after})
	b.statement(stmt.Statement)
	b.popJump()
	b.link(b.cur, after)
	b.enter(after)
}

func (b *Builder[E]) pushJump(node *ast.Node, breakTo, continueTo *Block[E], loop *ast.Node) {
	b.jumps = append(b.jumps, jumpTarget[E]{labels: labelsOf(node), breakTo: breakTo, continueTo: continueTo, loop: loop, breakable: true})
}

func (b *Builder[E]) popJump() {
	b.jumps = b.jumps[:len(b.jumps)-1]
}

func (b *Builder[E]) makeBreak(label *ast.Node) {
	if b.cur == nil {
		return
	}
	name := ""
	if label != nil {
		name = label.Text()
	}
	for i := len(b.jumps) - 1; i >= 0; i-- {
		target := b.jumps[i]
		if name == "" {
			if !target.breakable {
				continue
			}
		} else if !slices.Contains(target.labels, name) {
			continue
		}
		b.link(b.cur, target.breakTo)
		break
	}
	b.cur = nil
}

func (b *Builder[E]) makeContinue(label *ast.Node) {
	if b.cur == nil {
		return
	}
	name := ""
	if label != nil {
		name = label.Text()
	}
	for i := len(b.jumps) - 1; i >= 0; i-- {
		target := b.jumps[i]
		if target.continueTo == nil {
			continue
		}
		if name == "" || slices.Contains(target.labels, name) {
			b.loop(target.loop)
			b.link(b.cur, target.continueTo)
			break
		}
	}
	b.cur = nil
}

func (b *Builder[E]) makeReturn() {
	if b.cur == nil {
		return
	}
	if i := b.returnFrame(); i >= 0 {
		frame := b.tryStack[i]
		frame.returnedAny = true
		b.link(b.cur, frame.finallyEntry)
	}
	b.cur = nil
}

func (b *Builder[E]) makeThrow() {
	if b.cur == nil {
		return
	}
	if i := b.throwFrame(); i >= 0 {
		frame := b.tryStack[i]
		frame.thrownAny = true
		b.link(b.cur, throwTarget(frame))
	}
	b.cur = nil
}

// makeYield mirrors ESLint's handling of a suspended generator: the caller may
// resume it with `.return()` or `.throw()`, so both leaving paths are recorded
// before the normal continuation resumes in a fresh block.
func (b *Builder[E]) makeYield() {
	if b.cur == nil {
		return
	}
	if i := b.returnFrame(); i >= 0 {
		frame := b.tryStack[i]
		frame.returnedAny = true
		b.link(b.cur, frame.finallyEntry)
	}
	if i := b.throwFrame(); i >= 0 {
		frame := b.tryStack[i]
		frame.thrownAny = true
		b.link(b.cur, throwTarget(frame))
	}
	next := b.newBlock()
	b.link(b.cur, next)
	b.cur = next
}
