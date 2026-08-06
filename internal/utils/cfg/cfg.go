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
//     around the operand they may skip.
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
// order, and the blocks control can reach from it.
type Block[E any] struct {
	Events     []E
	Successors []*Block[E]

	index       int
	hasIncoming bool
}

// Index reports the block's position in Graph.Blocks, so a consumer can hold
// per-block analysis state in a slice rather than a map.
func (blk *Block[E]) Index() int {
	return blk.index
}

// Graph is the control-flow graph of one code path root. Blocks[0] is the entry
// block; the rest are in construction order, which no analysis should depend on.
type Graph[E any] struct {
	Blocks []*Block[E]
}

// Hooks are the points where a consumer turns a position in the walk into an
// event. Every hook may be nil, and every one is called only from a position
// control can reach, so Builder.Emit always lands somewhere.
//
// Read runs where node is evaluated as a read, Write where its value is stored;
// a target that is read before it is written — a compound assignment, an update
// expression, a destructuring element — sees both, in that order.
//
// Statement runs as each statement is reached, before it is laid out. Loop runs
// where control flows back into a loop for another iteration of it.
type Hooks[E any] struct {
	Read      func(b *Builder[E], node *ast.Node)
	Write     func(b *Builder[E], node *ast.Node)
	Statement func(b *Builder[E], node *ast.Node)
	Loop      func(b *Builder[E], loop *ast.Node)
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
}

// Builder lays out a graph. Consumers only see it through their Hooks, where
// Current and Emit are the operations that matter.
type Builder[E any] struct {
	hooks Hooks[E]

	blocks []*Block[E]
	cur    *Block[E]

	jumps      []jumpTarget[E]
	tryStack   []*tryFrame[E]
	chainJoins []*Block[E]
}

// Build lays out the control-flow graph of one code path root.
func Build[E any](root *ast.Node, hooks Hooks[E]) *Graph[E] {
	b := &Builder[E]{hooks: hooks}
	b.cur = b.newBlock()
	b.cur.hasIncoming = true

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

	return &Graph[E]{Blocks: b.blocks}
}

// parameter models a parameter binding: its default value is skipped when an
// argument was passed.
func (b *Builder[E]) parameter(node *ast.Node) {
	if b.cur == nil {
		return
	}
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

// Current returns the block being filled, or nil when the position is
// unreachable.
func (b *Builder[E]) Current() *Block[E] {
	return b.cur
}

// Emit records one event in the current block and returns where it landed, so a
// consumer can refer back to it. The block is nil when the position is
// unreachable and nothing was recorded.
func (b *Builder[E]) Emit(e E) (*Block[E], int) {
	if b.cur == nil {
		return nil, 0
	}
	index := len(b.cur.Events)
	b.cur.Events = append(b.cur.Events, e)
	return b.cur, index
}

func (b *Builder[E]) newBlock() *Block[E] {
	blk := &Block[E]{index: len(b.blocks)}
	b.blocks = append(b.blocks, blk)
	return blk
}

func (b *Builder[E]) link(from, to *Block[E]) {
	if from == nil || to == nil {
		return
	}
	from.Successors = append(from.Successors, to)
	to.hasIncoming = true
}

// enter makes blk the current block when something can reach it, and marks the
// path unreachable otherwise.
func (b *Builder[E]) enter(blk *Block[E]) {
	if blk.hasIncoming {
		b.cur = blk
	} else {
		b.cur = nil
	}
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
	if b.hooks.Loop != nil && b.cur != nil && node != nil {
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
