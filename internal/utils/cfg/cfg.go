// Package cfg builds a control-flow graph for one code path root — a source
// file, a function, a class static block, or a class field initializer.
//
// The graph mirrors the shape ESLint's code path analysis produces, because the
// observable behaviour of the rules built on it depends on it:
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
//     around the operand they may skip;
//   - the code after a `return`, `throw`, `break`, or `continue` is laid out in
//     blocks nothing reaches rather than dropped, keeping the edge that leads
//     into it.
//
// The graph carries no meaning of its own beyond that shape. A consumer names
// an event type and supplies Hooks; the builder calls them where a reference is
// read and where one is written, in evaluation order, and the consumer records
// whatever it needs into the current block. Analysis is then a plain dataflow
// over Graph.Blocks.
package cfg

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// Block is one basic block: the events its consumer recorded, in evaluation
// order, and the blocks led to from it.
//
// Successors spans the whole graph, unreachable blocks included, so a walk that
// only wants the code that runs skips the blocks whose Reachable is false.
type Block[E any] struct {
	Events     []E
	Successors []*Block[E]

	// Reachable reports whether control can arrive here. The code after an
	// abrupt exit is laid out in unreachable blocks rather than dropped, and a
	// consumer's hooks run in them, so a rule can ask what would have run had
	// the exit not been there.
	//
	// It is settled when the block is entered, from the edges that lead into
	// it by then. Every block is entered after the edges that could make it
	// reachable exist: a block only ever gains an edge afterwards from code
	// laid out inside it, which is unreachable whenever the block itself is.
	Reachable bool

	index        int32
	hasIncoming  bool
	final        bool
	thrown       bool
	cycleBarrier []bool
	successors   [2]*Block[E]
}

const blockChunkSize = 8

type blockChunk[E any] [blockChunkSize]Block[E]

// Index reports the block's position in Graph.Blocks, so a consumer can hold
// per-block analysis state in a slice rather than a map.
func (blk *Block[E]) Index() int {
	return int(blk.index)
}

// Graph is the control-flow graph of one code path root, unreachable blocks
// included. Blocks[0] is the entry block; the rest are in construction order,
// which no analysis should depend on.
//
// FinalBlocks are the reachable blocks where the root completes normally,
// returns, or is suspended by a yield. ThrownBlocks are the reachable blocks
// where an uncaught throw leaves the root. A block can be in both sets when a
// yield gives the caller both ways to terminate the generator. Keeping the two
// sets separate lets consumers mirror analyses, such as ESLint code paths,
// that deliberately exclude thrown exits. Different abrupt completions can
// also share one block after a finally clause, placing that block in both sets.
//
// EndReachable reports whether control can run off the end of the root — the
// question ESLint answers with `isAnySegmentReachable(codePath.currentSegments)`
// at the root's `:exit`. It is false when every path out of the root returns,
// throws, or loops forever.
type Graph[E any] struct {
	Blocks       []*Block[E]
	FinalBlocks  []*Block[E]
	ThrownBlocks []*Block[E]
	EndReachable bool
}

// Hooks are the points where a consumer turns a position in the walk into an
// event. Every hook may be nil, and every one is called from every position the
// walk reaches, whether or not control can arrive there — a consumer that has
// no use for unreachable code checks Builder.Current for it.
//
// Expression runs before each expression is evaluated. Read runs where an
// identifier is evaluated as a read, Write where its value is stored; a target
// that is read before it is written — a compound assignment, an update
// expression, a destructuring element — sees both, in that order.
//
// Statement runs as each statement is reached, before it is laid out. Loop runs
// where control flows back into a loop for another iteration of it.
type Hooks[E any] struct {
	Expression func(b *Builder[E], node *ast.Node)
	Read       func(b *Builder[E], node *ast.Node)
	Write      func(b *Builder[E], node *ast.Node)
	Statement  func(b *Builder[E], node *ast.Node)
	Loop       func(b *Builder[E], loop *ast.Node)
}

// tryPosition is where in a `try` statement construction currently is.
type tryPosition uint8

const (
	posTry tryPosition = iota
	posCatch
)

// tryFrame tracks one enclosing `try` statement while its blocks are built.
type tryFrame[E any] struct {
	position     tryPosition
	hasFinally   bool
	catchEntry   *Block[E]
	finallyEntry *Block[E]
	thrownForked bool
	thrownAny    bool
	implicit     []*Block[E]
	returnedAny  bool
}

// jumpTarget is one `break`/`continue` destination. continueTo is nil for
// targets that only `break` can reach (switch statements, labelled blocks).
// breakable mirrors ESLint's BreakContext#breakable: it is true for loops and
// switch statements, and false for a labelled statement that wraps neither, so
// an unlabelled `break` skips past it to the enclosing loop/switch. labels
// holds every label the target answers to — a loop wrapped as
// `outer: inner: while (…)` accepts both names. loop names the loop a
// `continue` starts another iteration of, and is nil where continuing does not
// re-enter the loop's first node: a `do`…`while` continues into its test, which
// may still end the loop.
type jumpTarget[E any] struct {
	labels     []string
	breakTo    *Block[E]
	continueTo *Block[E]
	loop       *ast.Node
	breakable  bool
	broken     bool
}

// Builder lays out a graph. Consumers only see it through their Hooks, where
// Current and Emit are the operations that matter.
type Builder[E any] struct {
	hooks Hooks[E]

	blocks      []*Block[E]
	cur         *Block[E]
	finals      []*Block[E]
	thrown      []*Block[E]
	blockChunks []*blockChunk[E]

	jumps      []jumpTarget[E]
	tryStack   []*tryFrame[E]
	chainJoins []*Block[E]
}

// Build lays out the control-flow graph of one code path root.
func Build[E any](root *ast.Node, hooks Hooks[E]) *Graph[E] {
	b := &Builder[E]{hooks: hooks, blocks: make([]*Block[E], 0, blockChunkSize)}
	b.cur = b.newBlock()
	b.cur.hasIncoming = true
	b.cur.Reachable = true

	switch root.Kind {
	case ast.KindSourceFile:
		b.statements(root.AsSourceFile().Statements)

	case ast.KindClassStaticBlockDeclaration:
		b.statement(root.AsClassStaticBlockDeclaration().Body)

	case ast.KindPropertyDeclaration:
		b.expr(root.AsPropertyDeclaration().Initializer)

	default:
		for _, parameter := range root.Parameters() {
			b.parameter(parameter)
		}
		// A `typeof x` inside the signature's type annotations references the
		// variable even though nothing in them runs.
		b.expr(root.Type())
		if body := root.Body(); body != nil {
			if body.Kind == ast.KindBlock {
				b.statement(body)
			} else {
				b.expr(body)
			}
		}
	}

	endReachable := b.cur.Reachable
	if endReachable {
		b.markFinal(b.cur)
	}
	return &Graph[E]{
		Blocks:       b.blocks,
		FinalBlocks:  b.finals,
		ThrownBlocks: b.thrown,
		EndReachable: endReachable,
	}
}

// parameter models a parameter binding: its default value is skipped when an
// argument was passed.
func (b *Builder[E]) parameter(node *ast.Node) {
	parameter := node.AsParameterDeclaration()
	if parameter == nil {
		return
	}
	b.expr(parameter.Type)
	b.bindWithDefault(parameter.Name(), parameter.Initializer)
}

// ---------------------------------------------------------------------------
// blocks
// ---------------------------------------------------------------------------

// Current returns the block being filled. Its Reachable field says whether the
// position control is at is one control can arrive at.
func (b *Builder[E]) Current() *Block[E] {
	return b.cur
}

// Emit records one event in the current block and returns where it landed, so a
// consumer can refer back to it.
func (b *Builder[E]) Emit(e E) (*Block[E], int) {
	index := len(b.cur.Events)
	b.cur.Events = append(b.cur.Events, e)
	return b.cur, index
}

func (b *Builder[E]) newBlock() *Block[E] {
	index := len(b.blocks)
	chunkIndex := index / blockChunkSize
	if chunkIndex == len(b.blockChunks) {
		b.blockChunks = append(b.blockChunks, new(blockChunk[E]))
	}
	blk := &b.blockChunks[chunkIndex][index%blockChunkSize]
	blk.index = int32(index)
	b.blocks = append(b.blocks, blk)
	return blk
}

func (b *Builder[E]) markFinal(blk *Block[E]) {
	if blk == nil || !blk.Reachable || blk.final {
		return
	}
	blk.final = true
	b.finals = append(b.finals, blk)
}

func (b *Builder[E]) markThrown(blk *Block[E]) {
	if blk == nil || !blk.Reachable || blk.thrown {
		return
	}
	blk.thrown = true
	b.thrown = append(b.thrown, blk)
}

func (b *Builder[E]) link(from, to *Block[E]) {
	b.linkWithCycleBarrier(from, to, false)
}

func (b *Builder[E]) linkWithCycleBarrier(from, to *Block[E], barrier bool) int {
	if from == nil || to == nil {
		return -1
	}
	index := len(from.Successors)
	b.appendSuccessor(from, to)
	b.setCycleBarrier(from, index, barrier)
	if from.Reachable {
		to.hasIncoming = true
	}
	return index
}

func (b *Builder[E]) appendSuccessor(from, to *Block[E]) {
	index := len(from.Successors)
	if index < len(from.successors) {
		from.successors[index] = to
		from.Successors = from.successors[:index+1]
		return
	}
	from.Successors = append(from.Successors, to)
}

func (b *Builder[E]) setCycleBarrier(from *Block[E], index int, barrier bool) {
	if from == nil || index < 0 {
		return
	}
	if from.cycleBarrier == nil {
		if !barrier {
			return
		}
		from.cycleBarrier = make([]bool, len(from.Successors))
	} else {
		for len(from.cycleBarrier) < len(from.Successors) {
			from.cycleBarrier = append(from.cycleBarrier, false)
		}
	}
	from.cycleBarrier[index] = barrier
}

// enter makes blk the current block and settles its reachability, so every edge
// that could make it reachable has to be linked before it is entered.
func (b *Builder[E]) enter(blk *Block[E]) {
	blk.Reachable = blk.hasIncoming
	b.cur = blk
}

// makeUnreachable ends the path at the current block and carries the walk on in
// a fresh block nothing reaches. The edge into it is recorded, but — unlike
// every other edge — it does not make its target something control arrives at,
// which is what leaving through it meant in the first place. That is why the
// block is never entered: its reachability is settled here, as false.
func (b *Builder[E]) makeUnreachable() {
	next := b.newBlock()
	b.appendSuccessor(b.cur, next)
	if b.cur.cycleBarrier != nil {
		b.cur.cycleBarrier = append(b.cur.cycleBarrier, false)
	}
	b.cur = next
}

// enterDisconnected makes blk current when the position the walk stands at is
// one control arrives at, rather than when an edge into blk already is. It is
// how a block whose only edge in is laid out later still gets laid out here,
// which is the order ESLint traverses one in.
func (b *Builder[E]) enterDisconnected(blk *Block[E]) {
	blk.hasIncoming = blk.hasIncoming || b.cur.Reachable
	b.enter(blk)
}

func (b *Builder[E]) read(node *ast.Node) {
	if b.hooks.Read != nil {
		b.hooks.Read(b, node)
	}
}

func (b *Builder[E]) write(node *ast.Node) {
	if b.hooks.Write != nil {
		b.hooks.Write(b, node)
	}
}

func (b *Builder[E]) reachedStatement(node *ast.Node) {
	if b.hooks.Statement != nil {
		b.hooks.Statement(b, node)
	}
}

func (b *Builder[E]) loop(node *ast.Node) {
	if b.hooks.Loop != nil && node != nil {
		b.hooks.Loop(b, node)
	}
}

// ---------------------------------------------------------------------------
// throw / return routing
// ---------------------------------------------------------------------------

// throwFrame returns the innermost `try` frame an exception raised at the
// current position lands in, or -1 when it leaves the code path entirely.
func (b *Builder[E]) throwFrame() int {
	for i := len(b.tryStack) - 1; i >= 0; i-- {
		f := b.tryStack[i]
		if f.position == posTry || (f.hasFinally && f.position == posCatch) {
			return i
		}
	}
	return -1
}

func throwTarget[E any](f *tryFrame[E]) *Block[E] {
	if f.position == posTry && f.catchEntry != nil {
		return f.catchEntry
	}
	return f.finallyEntry
}

// returnFrame returns the innermost `try` frame whose `finally` block runs
// before a `return` leaves the code path, or -1 when none does.
func (b *Builder[E]) returnFrame() int {
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
func (b *Builder[E]) firstThrowableFork() {
	if !b.cur.Reachable {
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
	if f.position == posCatch && f.hasFinally {
		// ESLint merges an implicit exception from a catch body into the
		// catch's ordinary finally continuation. An explicit throw still uses
		// the abrupt layout through makeThrow.
		f.implicit = append(f.implicit, b.cur)
		next := b.newBlock()
		b.link(b.cur, next)
		b.enter(next)
		return
	}
	f.thrownAny = true
	b.link(b.cur, throwTarget(f))
	next := b.newBlock()
	b.link(b.cur, next)
	b.enter(next)
}

// snapshotForks records the "already forked" state of every enclosing `try` so
// the second copy of a `finally` block forks at the same place the first did.
func (b *Builder[E]) snapshotForks() []bool {
	snapshot := make([]bool, len(b.tryStack))
	for i, f := range b.tryStack {
		snapshot[i] = f.thrownForked
	}
	return snapshot
}

func (b *Builder[E]) restoreForks(snapshot []bool) {
	for i, forked := range snapshot {
		if i < len(b.tryStack) {
			b.tryStack[i].thrownForked = forked
		}
	}
}
