package no_useless_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/utils"
)

// This file builds a control-flow graph for one code path root (a source file,
// a function, a class static block, or a class field initializer) and answers
// "can the value written here still be read?" over it.
//
// The graph mirrors the shape ESLint's code path analysis produces, because the
// rule's observable behaviour depends on it:
//
//   - the end of a `try` block flows into its `catch` block as well as past the
//     statement;
//   - the first node inside a `try` block (or inside the `catch` block of a
//     `try` that has a `finally`) that could throw forks to the handler, so an
//     assignment later in the block is bypassed on the throwing path;
//   - a `finally` block is laid out twice, once for normal completion and once
//     for the path that leaves the `try` statement through `return`, `throw`,
//     or a suspended `yield`; `break` and `continue` do not take that path;
//   - `&&`, `||`, `??`, `?:`, optional chains, and destructuring defaults fork
//     around the operand they may skip.
//
// Reads and writes are recorded as events in evaluation order, so liveness is a
// plain backward dataflow over the graph.

type eventKind uint8

const (
	eventRead eventKind = iota
	eventWrite
)

type event struct {
	kind   eventKind
	sym    *ast.Symbol
	assign *assignment
}

type block struct {
	events      []event
	successors  []*block
	hasIncoming bool

	// Per-symbol dataflow scratch, reset by liveWrites for each symbol.
	use     bool
	def     bool
	liveIn  bool
	liveOut bool
}

// writeSite is one place in the graph where an assignment appears. An
// assignment inside a `finally` block appears twice — once per copy of the
// block — and only counts as unused when every site is dead.
type writeSite struct {
	blk   *block
	index int
}

// assignment is one reportable write: the identifier ESLint reports on plus the
// variable it targets.
type assignment struct {
	sym        *ast.Symbol
	identifier *ast.Node
	sites      []writeSite
	// silent marks an assignment that still overwrites the variable but is
	// never reported, because it sits in a `try` block.
	silent bool
}

type tryPosition uint8

const (
	posTry tryPosition = iota
	posCatch
)

// tryFrame tracks one enclosing `try` statement while its blocks are built.
type tryFrame struct {
	position     tryPosition
	hasFinally   bool
	catchEntry   *block
	finallyEntry *block
	thrownForked bool
	thrownAny    bool
	returnedAny  bool
}

// jumpTarget is one `break`/`continue` destination. continueTo is nil for
// targets that only `break` can reach (switch statements, labelled blocks).
type jumpTarget struct {
	label      string
	breakTo    *block
	continueTo *block
}

type builder struct {
	// assignByIdent maps a write target identifier to the assignment it
	// belongs to; identifiers absent from the map target untracked variables.
	assignByIdent map[*ast.Node]*assignment
	// readNodes maps a reference identifier that reads a tracked variable to
	// that variable's symbol.
	readNodes map[*ast.Node]*ast.Symbol

	blocks []*block
	cur    *block

	jumps      []jumpTarget
	tryStack   []*tryFrame
	chainJoins []*block
}

func (b *builder) newBlock() *block {
	blk := &block{}
	b.blocks = append(b.blocks, blk)
	return blk
}

func (b *builder) link(from, to *block) {
	if from == nil || to == nil {
		return
	}
	from.successors = append(from.successors, to)
	to.hasIncoming = true
}

// enter makes blk the current block when something can reach it, and marks the
// path unreachable otherwise.
func (b *builder) enter(blk *block) {
	if blk.hasIncoming {
		b.cur = blk
	} else {
		b.cur = nil
	}
}

func (b *builder) emitRead(sym *ast.Symbol) {
	if b.cur == nil {
		return
	}
	b.cur.events = append(b.cur.events, event{kind: eventRead, sym: sym})
}

func (b *builder) emitWrite(a *assignment) {
	if b.cur == nil {
		return
	}
	a.sites = append(a.sites, writeSite{blk: b.cur, index: len(b.cur.events)})
	b.cur.events = append(b.cur.events, event{kind: eventWrite, sym: a.sym, assign: a})
}

// ---------------------------------------------------------------------------
// throw / return routing
// ---------------------------------------------------------------------------

// throwFrame returns the innermost `try` frame an exception raised at the
// current position lands in, or -1 when it leaves the code path entirely.
func (b *builder) throwFrame() int {
	for i := len(b.tryStack) - 1; i >= 0; i-- {
		f := b.tryStack[i]
		if f.position == posTry || (f.hasFinally && f.position == posCatch) {
			return i
		}
	}
	return -1
}

func throwTarget(f *tryFrame) *block {
	if f.position == posTry && f.catchEntry != nil {
		return f.catchEntry
	}
	return f.finallyEntry
}

// returnFrame returns the innermost `try` frame whose `finally` block runs
// before a `return` leaves the code path, or -1 when none does.
func (b *builder) returnFrame() int {
	for i := len(b.tryStack) - 1; i >= 0; i-- {
		if b.tryStack[i].hasFinally {
			return i
		}
	}
	return -1
}

// firstThrowableFork forks to the enclosing handler the first time a node that
// could throw is completed inside a `try` block (or inside the `catch` block of
// a `try` that has a `finally`). Later throwable nodes in the same block reuse
// that first fork, matching ESLint's approximation.
func (b *builder) firstThrowableFork() {
	if b.cur == nil {
		return
	}
	i := b.throwFrame()
	if i < 0 {
		return
	}
	f := b.tryStack[i]
	if f.thrownForked {
		return
	}
	f.thrownForked = true
	f.thrownAny = true
	b.link(b.cur, throwTarget(f))
	next := b.newBlock()
	b.link(b.cur, next)
	b.cur = next
}

// snapshotForks records the "already forked" state of every enclosing `try` so
// the second copy of a `finally` block forks at the same place the first did.
func (b *builder) snapshotForks() []bool {
	snapshot := make([]bool, len(b.tryStack))
	for i, f := range b.tryStack {
		snapshot[i] = f.thrownForked
	}
	return snapshot
}

func (b *builder) restoreForks(snapshot []bool) {
	for i, forked := range snapshot {
		if i < len(b.tryStack) {
			b.tryStack[i].thrownForked = forked
		}
	}
}

// ---------------------------------------------------------------------------
// statements
// ---------------------------------------------------------------------------

func (b *builder) statements(list *ast.NodeList) {
	if list == nil {
		return
	}
	for _, stmt := range list.Nodes {
		b.statement(stmt)
	}
}

func (b *builder) statement(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}

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

func (b *builder) variableDeclarationList(list *ast.Node) {
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

func (b *builder) variableDeclaration(node *ast.Node) {
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
	if decl.Initializer == nil {
		// `let x;` writes nothing this rule can report on, but a binding
		// pattern can still contain evaluated expressions.
		b.patternReads(decl.Name())
		return
	}
	b.patternReads(decl.Name())
	b.expr(decl.Initializer)
	b.patternWrites(decl.Name())
}

func (b *builder) ifStatement(node *ast.Node) {
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

func (b *builder) whileStatement(node *ast.Node) {
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

	b.pushJump(node, after, test)
	b.enter(body)
	b.statement(stmt.Statement)
	b.link(b.cur, test)
	b.popJump()

	b.enter(after)
}

func (b *builder) doStatement(node *ast.Node) {
	stmt := node.AsDoStatement()
	body := b.newBlock()
	b.link(b.cur, body)
	test := b.newBlock()
	after := b.newBlock()

	b.pushJump(node, after, test)
	b.enter(body)
	b.statement(stmt.Statement)
	b.link(b.cur, test)
	b.popJump()

	b.enter(test)
	b.expr(stmt.Expression)
	if b.cur != nil {
		b.link(b.cur, body)
		if !isAlwaysTruthyTest(stmt.Expression) {
			b.link(b.cur, after)
		}
	}

	b.enter(after)
}

func (b *builder) forStatement(node *ast.Node) {
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

	b.pushJump(node, after, update)
	b.enter(body)
	b.statement(stmt.Statement)
	b.link(b.cur, update)
	b.popJump()

	if stmt.Incrementor != nil {
		b.enter(update)
		b.expr(stmt.Incrementor)
		b.link(b.cur, test)
	}

	b.enter(after)
}

func (b *builder) forInOfStatement(node *ast.Node) {
	stmt := node.AsForInOrOfStatement()
	b.expr(stmt.Expression)

	head := b.newBlock()
	b.link(b.cur, head)
	b.enter(head)

	after := b.newBlock()
	b.link(head, after)
	body := b.newBlock()
	b.link(head, body)

	b.pushJump(node, after, head)
	b.enter(body)
	if stmt.Initializer != nil {
		if stmt.Initializer.Kind == ast.KindVariableDeclarationList {
			// The loop variable is bound, not assigned: ESLint only tracks
			// declarators that carry an initializer.
			declList := stmt.Initializer.AsVariableDeclarationList()
			if declList != nil && declList.Declarations != nil {
				for _, decl := range declList.Declarations.Nodes {
					b.patternReads(decl.Name())
				}
			}
		} else {
			// `for (x of xs)` / `for ([a] of xs)` — a write reference, but not
			// an assignment this rule reports on.
			b.patternReads(stmt.Initializer)
		}
	}
	b.statement(stmt.Statement)
	b.link(b.cur, head)
	b.popJump()

	b.enter(after)
}

func (b *builder) switchStatement(node *ast.Node) {
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

	bodies := make([]*block, len(clauses))
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

	b.pushJump(node, after, nil)
	var fallthroughFrom *block
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

func (b *builder) tryStatement(node *ast.Node) {
	stmt := node.AsTryStatement()
	hasCatch := stmt.CatchClause != nil
	hasFinally := stmt.FinallyBlock != nil

	frame := &tryFrame{position: posTry, hasFinally: hasFinally}
	if hasCatch {
		frame.catchEntry = b.newBlock()
	}
	if hasFinally {
		frame.finallyEntry = b.newBlock()
	}
	b.tryStack = append(b.tryStack, frame)

	b.statement(stmt.TryBlock)

	var normalEnds []*block
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
			b.patternReads(catchClause.VariableDeclaration.Name())
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
		if b.cur != nil && frame.thrownAny {
			if i := b.throwFrame(); i >= 0 {
				outer := b.tryStack[i]
				outer.thrownAny = true
				b.link(b.cur, throwTarget(outer))
			}
		}
	}

	b.enter(after)
}

func (b *builder) labeledStatement(node *ast.Node) {
	stmt := node.AsLabeledStatement()
	if isBreakableStatement(stmt.Statement) {
		// The label belongs to the loop or switch itself; pushJump reads it
		// back off the parent chain.
		b.statement(stmt.Statement)
		return
	}
	after := b.newBlock()
	b.jumps = append(b.jumps, jumpTarget{label: stmt.Label.Text(), breakTo: after})
	b.statement(stmt.Statement)
	b.popJump()
	b.link(b.cur, after)
	b.enter(after)
}

func (b *builder) pushJump(node *ast.Node, breakTo, continueTo *block) {
	b.jumps = append(b.jumps, jumpTarget{label: labelOf(node), breakTo: breakTo, continueTo: continueTo})
}

func (b *builder) popJump() {
	b.jumps = b.jumps[:len(b.jumps)-1]
}

func (b *builder) makeBreak(label *ast.Node) {
	if b.cur == nil {
		return
	}
	name := ""
	if label != nil {
		name = label.Text()
	}
	for i := len(b.jumps) - 1; i >= 0; i-- {
		target := b.jumps[i]
		if name == "" || target.label == name {
			b.link(b.cur, target.breakTo)
			break
		}
	}
	b.cur = nil
}

func (b *builder) makeContinue(label *ast.Node) {
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
		if name == "" || target.label == name {
			b.link(b.cur, target.continueTo)
			break
		}
	}
	b.cur = nil
}

func (b *builder) makeReturn() {
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

func (b *builder) makeThrow() {
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
func (b *builder) makeYield() {
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

// ---------------------------------------------------------------------------
// expressions
// ---------------------------------------------------------------------------

func (b *builder) expr(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}

	switch node.Kind {
	case ast.KindIdentifier:
		if sym, ok := b.readNodes[node]; ok {
			b.emitRead(sym)
		}
		if isThrowableIdentifier(node) {
			b.firstThrowableFork()
		}

	case ast.KindParenthesizedExpression:
		b.expr(node.AsParenthesizedExpression().Expression)

	case ast.KindBinaryExpression:
		b.binaryExpression(node)

	case ast.KindConditionalExpression:
		b.conditionalExpression(node)

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			b.updateExpression(unary.Operand)
			return
		}
		b.expr(unary.Operand)

	case ast.KindPostfixUnaryExpression:
		b.updateExpression(node.AsPostfixUnaryExpression().Operand)

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
		ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		b.accessOrCall(node)

	case ast.KindYieldExpression:
		b.expr(node.AsYieldExpression().Expression)
		b.makeYield()

	case ast.KindClassExpression:
		b.classLike(node)

	default:
		if isCodePathRoot(node) {
			b.nestedFunction(node)
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			b.visitUnknown(child)
			return false
		})
	}
}

// visitUnknown routes a child of a node the builder has no dedicated handler
// for. Statements and expressions are handled by their own walkers so control
// flow inside them is still modelled.
func (b *builder) visitUnknown(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}
	if ast.IsStatement(node) {
		b.statement(node)
		return
	}
	b.expr(node)
}

func (b *builder) binaryExpression(node *ast.Node) {
	binary := node.AsBinaryExpression()
	operator := binary.OperatorToken.Kind

	switch {
	case operator == ast.KindAmpersandAmpersandToken ||
		operator == ast.KindBarBarToken ||
		operator == ast.KindQuestionQuestionToken:
		b.expr(binary.Left)
		join := b.newBlock()
		b.link(b.cur, join)
		right := b.newBlock()
		b.link(b.cur, right)
		b.enter(right)
		b.expr(binary.Right)
		b.link(b.cur, join)
		b.enter(join)

	case ast.IsLogicalOrCoalescingAssignmentOperator(operator):
		// `a ||= b` reads a, then conditionally evaluates b and writes a.
		b.patternReads(binary.Left)
		b.expr(binary.Left)
		join := b.newBlock()
		b.link(b.cur, join)
		right := b.newBlock()
		b.link(b.cur, right)
		b.enter(right)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)
		b.link(b.cur, join)
		b.enter(join)

	case operator == ast.KindEqualsToken:
		b.patternReads(binary.Left)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)

	case ast.IsAssignmentOperator(operator):
		// Compound assignment reads the target before writing it.
		b.patternReads(binary.Left)
		b.expr(binary.Left)
		b.expr(binary.Right)
		b.patternWrites(binary.Left)

	default:
		b.expr(binary.Left)
		b.expr(binary.Right)
	}
}

func (b *builder) conditionalExpression(node *ast.Node) {
	cond := node.AsConditionalExpression()
	b.expr(cond.Condition)
	testEnd := b.cur
	join := b.newBlock()

	whenTrue := b.newBlock()
	b.link(testEnd, whenTrue)
	b.enter(whenTrue)
	b.expr(cond.WhenTrue)
	b.link(b.cur, join)

	whenFalse := b.newBlock()
	b.link(testEnd, whenFalse)
	b.enter(whenFalse)
	b.expr(cond.WhenFalse)
	b.link(b.cur, join)

	b.enter(join)
}

func (b *builder) updateExpression(operand *ast.Node) {
	b.patternReads(operand)
	b.expr(operand)
	b.patternWrites(operand)
}

// accessOrCall walks a member access or call and forks around each `?.` link so
// the rest of the chain can be skipped.
func (b *builder) accessOrCall(node *ast.Node) {
	outermost := ast.IsOptionalChain(node) && ast.IsOutermostOptionalChain(node)
	if outermost {
		b.chainJoins = append(b.chainJoins, b.newBlock())
	}

	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		b.expr(node.AsPropertyAccessExpression().Expression)
		b.forkOptionalChain(node)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		b.expr(access.Expression)
		b.forkOptionalChain(node)
		b.expr(access.ArgumentExpression)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		b.expr(call.Expression)
		b.forkOptionalChain(node)
		if call.Arguments != nil {
			for _, arg := range call.Arguments.Nodes {
				b.expr(arg)
			}
		}
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		b.expr(newExpr.Expression)
		if newExpr.Arguments != nil {
			for _, arg := range newExpr.Arguments.Nodes {
				b.expr(arg)
			}
		}
	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		b.expr(tagged.Tag)
		b.expr(tagged.Template)
	}

	b.firstThrowableFork()

	if outermost {
		join := b.chainJoins[len(b.chainJoins)-1]
		b.chainJoins = b.chainJoins[:len(b.chainJoins)-1]
		b.link(b.cur, join)
		b.enter(join)
	}
}

func (b *builder) forkOptionalChain(node *ast.Node) {
	if b.cur == nil || len(b.chainJoins) == 0 || !ast.IsOptionalChainRoot(node) {
		return
	}
	b.link(b.cur, b.chainJoins[len(b.chainJoins)-1])
	next := b.newBlock()
	b.link(b.cur, next)
	b.cur = next
}

// nestedFunction keeps evaluating the parts of a nested function that belong to
// the enclosing code path — a computed member name and decorators — while its
// parameters and body form their own code path root.
func (b *builder) nestedFunction(node *ast.Node) {
	b.memberHeader(node)
}

func (b *builder) classLike(node *ast.Node) {
	class := node.ClassLikeData()
	if class == nil {
		return
	}
	b.decorators(node)
	if class.HeritageClauses != nil {
		for _, clause := range class.HeritageClauses.Nodes {
			heritage := clause.AsHeritageClause()
			if heritage == nil || heritage.Types == nil {
				continue
			}
			for _, typeNode := range heritage.Types.Nodes {
				b.expr(typeNode)
			}
		}
	}
	if class.Members != nil {
		for _, member := range class.Members.Nodes {
			b.memberHeader(member)
		}
	}
}

// memberHeader evaluates the parts of a class or object member that belong to
// the enclosing code path: its decorators, its computed key, and its type
// annotation. The member's parameters and body form their own code path.
func (b *builder) memberHeader(node *ast.Node) {
	b.decorators(node)
	if ast.IsFunctionLikeDeclaration(node) {
		for _, parameter := range node.Parameters() {
			b.decorators(parameter)
		}
	}
	name := node.Name()
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		b.expr(name.AsComputedPropertyName().Expression)
	}
	if node.Kind == ast.KindPropertyDeclaration {
		b.expr(node.AsPropertyDeclaration().Type)
	}
}

func (b *builder) decorators(node *ast.Node) {
	for _, decorator := range node.Decorators() {
		b.expr(decorator.AsDecorator().Expression)
	}
}

// ---------------------------------------------------------------------------
// assignment patterns
// ---------------------------------------------------------------------------

// patternReads emits the reads an assignment target performs before anything is
// written: the object and computed key of a member target, and the key
// expressions of a destructuring pattern.
func (b *builder) patternReads(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		// A pure target reads nothing, but ESLint still counts naming the
		// variable as a point where the enclosing `try` block can throw — so
		// the handler sees the value the target had before the assignment.
		if isThrowableIdentifier(node) {
			b.firstThrowableFork()
		}

	case ast.KindOmittedExpression:
		// An array hole writes nothing.

	case ast.KindParenthesizedExpression:
		b.patternReads(node.AsParenthesizedExpression().Expression)

	case ast.KindNonNullExpression:
		b.patternReads(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		b.patternReads(node.AsAsExpression().Expression)

	case ast.KindSatisfiesExpression:
		b.patternReads(node.AsSatisfiesExpression().Expression)

	case ast.KindTypeAssertionExpression:
		b.patternReads(node.AsTypeAssertion().Expression)

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		pattern := node.AsBindingPattern()
		if pattern.Elements == nil {
			return
		}
		for _, element := range pattern.Elements.Nodes {
			binding := element.AsBindingElement()
			if binding == nil {
				continue
			}
			if binding.PropertyName != nil && binding.PropertyName.Kind == ast.KindComputedPropertyName {
				b.expr(binding.PropertyName.AsComputedPropertyName().Expression)
			}
			b.patternReads(binding.Name())
		}

	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				assignment := property.AsPropertyAssignment()
				if name := assignment.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					b.expr(name.AsComputedPropertyName().Expression)
				}
				b.patternReads(assignment.Initializer)
			case ast.KindShorthandPropertyAssignment:
				// The name is the target; its default is evaluated later.
			case ast.KindSpreadAssignment:
				b.patternReads(property.AsSpreadAssignment().Expression)
			}
		}

	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			b.patternReads(element)
		}

	case ast.KindSpreadElement:
		b.patternReads(node.AsSpreadElement().Expression)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			// A default inside a destructuring target; the default expression
			// is evaluated during patternWrites.
			b.patternReads(binary.Left)
			return
		}
		b.expr(node)

	default:
		// A member target such as `obj.x` or `obj[k]` evaluates its receiver.
		b.expr(node)
	}
}

// patternWrites emits the writes an assignment target performs, in source
// order, forking around each default value the target may skip.
func (b *builder) patternWrites(node *ast.Node) {
	if node == nil || b.cur == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		if a, ok := b.assignByIdent[node]; ok {
			b.emitWrite(a)
		}

	case ast.KindOmittedExpression:

	case ast.KindParenthesizedExpression:
		b.patternWrites(node.AsParenthesizedExpression().Expression)

	case ast.KindNonNullExpression:
		b.patternWrites(node.AsNonNullExpression().Expression)

	case ast.KindAsExpression:
		b.patternWrites(node.AsAsExpression().Expression)

	case ast.KindSatisfiesExpression:
		b.patternWrites(node.AsSatisfiesExpression().Expression)

	case ast.KindTypeAssertionExpression:
		b.patternWrites(node.AsTypeAssertion().Expression)

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		pattern := node.AsBindingPattern()
		if pattern.Elements == nil {
			return
		}
		for _, element := range pattern.Elements.Nodes {
			binding := element.AsBindingElement()
			if binding == nil {
				continue
			}
			b.withDefault(binding.Initializer, binding.Name())
		}

	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				b.patternWrites(property.AsPropertyAssignment().Initializer)
			case ast.KindShorthandPropertyAssignment:
				shorthand := property.AsShorthandPropertyAssignment()
				b.withDefault(shorthand.ObjectAssignmentInitializer, shorthand.Name())
			case ast.KindSpreadAssignment:
				b.patternWrites(property.AsSpreadAssignment().Expression)
			}
		}

	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			b.patternWrites(element)
		}

	case ast.KindSpreadElement:
		b.patternWrites(node.AsSpreadElement().Expression)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			b.withDefault(binary.Right, binary.Left)
		}
	}
}

// withDefault models `target = fallback` inside a destructuring pattern: the
// fallback is only evaluated when the incoming value is undefined.
func (b *builder) withDefault(fallback *ast.Node, target *ast.Node) {
	if fallback == nil {
		b.patternWrites(target)
		return
	}
	if b.cur == nil {
		return
	}
	join := b.newBlock()
	bypass := b.newBlock()
	b.link(b.cur, bypass)
	withFallback := b.newBlock()
	b.link(b.cur, withFallback)

	b.enter(withFallback)
	b.expr(fallback)
	b.patternWrites(target)
	b.link(b.cur, join)

	b.enter(bypass)
	b.patternWrites(target)
	b.link(b.cur, join)

	b.enter(join)
}

// ---------------------------------------------------------------------------
// liveness
// ---------------------------------------------------------------------------

// deadWrites reports, per assignment, whether every site of that assignment is
// dead — no read of the variable is reachable before the value is overwritten.
func deadWrites(blocks []*block, assignments []*assignment) map[*assignment]bool {
	bySymbol := make(map[*ast.Symbol][]*assignment)
	for _, a := range assignments {
		bySymbol[a.sym] = append(bySymbol[a.sym], a)
	}

	dead := make(map[*assignment]bool, len(assignments))
	for sym, symAssignments := range bySymbol {
		computeLiveness(blocks, sym)
		for _, a := range symAssignments {
			if len(a.sites) == 0 {
				continue
			}
			isDead := true
			for _, site := range a.sites {
				if liveAfter(sym, site) {
					isDead = false
					break
				}
			}
			dead[a] = isDead
		}
	}
	return dead
}

// computeLiveness runs a backward liveness dataflow for one variable.
func computeLiveness(blocks []*block, sym *ast.Symbol) {
	for _, blk := range blocks {
		blk.use = false
		blk.def = false
		blk.liveIn = false
		blk.liveOut = false
		for _, e := range blk.events {
			if e.sym != sym {
				continue
			}
			if e.kind == eventRead {
				if !blk.def {
					blk.use = true
				}
			} else if !blk.use {
				blk.def = true
			}
			if blk.use || blk.def {
				break
			}
		}
		blk.liveIn = blk.use
	}

	for changed := true; changed; {
		changed = false
		for i := len(blocks) - 1; i >= 0; i-- {
			blk := blocks[i]
			liveOut := false
			for _, successor := range blk.successors {
				if successor.liveIn {
					liveOut = true
					break
				}
			}
			liveIn := blk.use || (!blk.def && liveOut)
			if liveOut != blk.liveOut || liveIn != blk.liveIn {
				blk.liveOut = liveOut
				blk.liveIn = liveIn
				changed = true
			}
		}
	}
}

// liveAfter reports whether the value written at site can still be read.
func liveAfter(sym *ast.Symbol, site writeSite) bool {
	for _, e := range site.blk.events[site.index+1:] {
		if e.sym != sym {
			continue
		}
		return e.kind == eventRead
	}
	return site.blk.liveOut
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func isBreakableStatement(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindSwitchStatement, ast.KindWhileStatement, ast.KindDoStatement,
		ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}

func labelOf(node *ast.Node) string {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindLabeledStatement {
		return ""
	}
	labeled := parent.AsLabeledStatement()
	if labeled.Statement != node {
		return ""
	}
	return labeled.Label.Text()
}

// isAlwaysTruthyTest mirrors ESLint's `getBooleanValueIfSimpleConstant`: only a
// literal test can mark the loop infinite, so the code after it is unreachable.
func isAlwaysTruthyTest(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindRegularExpressionLiteral:
		return true
	case ast.KindNumericLiteral:
		return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text) != "0"
	case ast.KindBigIntLiteral:
		text := node.AsBigIntLiteral().Text
		return text != "0n" && text != "0"
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text != ""
	}
	return false
}

// isThrowableIdentifier reports whether completing this identifier is one of
// the points ESLint treats as able to throw inside a `try` block. Names that
// only declare or label something are excluded, as are JSX names, which ESLint
// sees as a separate node type.
func isThrowableIdentifier(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if ast.IsJsxTagName(node) {
		return false
	}
	switch parent.Kind {
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement,
		ast.KindArrayBindingPattern, ast.KindImportSpecifier, ast.KindExportSpecifier,
		ast.KindImportClause, ast.KindNamespaceImport, ast.KindNamespaceExport,
		ast.KindJsxAttribute, ast.KindJsxNamespacedName:
		return false
	case ast.KindBindingElement:
		binding := parent.AsBindingElement()
		if binding.DotDotDotToken != nil || binding.PropertyName == node {
			return false
		}
		return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectBindingPattern
	case ast.KindCatchClause:
		return false
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindClassDeclaration, ast.KindClassExpression, ast.KindVariableDeclaration,
		ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindEnumDeclaration,
		ast.KindModuleDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return parent.Name() != node
	}
	return true
}
