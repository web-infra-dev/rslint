package rule

import (
	"sync"

	"github.com/web-infra-dev/rslint/internal/program"
)

// GenerationScope is shared by every rule context of one native lint pass over
// one immutable source generation: all of its Programs, files and rules. It
// lets a rule derive a cross-Program answer once, such as which exports of a
// project are imported anywhere, instead of rebuilding it per Program or file.
//
// The linter creates a scope only for passes whose source universe is frozen
// for the whole request (CLI and API, including each autofix round). Editor
// passes lint a changing document set and receive no scope, so rules must treat
// a missing scope as "project analysis unavailable" rather than infer an answer
// from the current Program alone.
//
// Values live exactly as long as the pass, so they may reference its Programs
// and source files. They must be immutable after their builder returns, because
// rules on every Program and file read them concurrently.
type GenerationScope struct {
	programs       []*program.Program
	singleThreaded bool

	mu     sync.Mutex
	values map[any]*generationValue
}

type generationValue struct {
	once  sync.Once
	value any
}

// NewGenerationScope creates the scope of one lint pass. programs is the
// generation's complete ordered Program sequence.
func NewGenerationScope(programs []*program.Program, singleThreaded bool) *GenerationScope {
	return &GenerationScope{
		programs:       append([]*program.Program(nil), programs...),
		singleThreaded: singleThreaded,
		values:         make(map[any]*generationValue),
	}
}

// Programs returns the generation's Programs in their stable order. The slice
// is read-only.
func (scope *GenerationScope) Programs() []*program.Program {
	if scope == nil {
		return nil
	}
	return scope.programs
}

// SingleThreaded reports whether the pass asked not to start parallel work.
func (scope *GenerationScope) SingleThreaded() bool {
	return scope != nil && scope.singleThreaded
}

// GenerationScope returns the pass-wide scope, or nil when the pass does not
// provide project analysis.
func (ctx *RuleContext) GenerationScope() *GenerationScope {
	if ctx == nil || ctx.fileCache == nil {
		return nil
	}
	return ctx.fileCache.generation
}

// CachedByGeneration builds the value for key once per lint pass, even when
// rules on several Programs request it concurrently. ok is false, and build is
// not called, when the pass has no generation scope. Keys must be comparable,
// use a type owned by the consumer package, and include every configuration
// input that changes the result.
func CachedByGeneration[T any](ctx RuleContext, key any, build func(*GenerationScope) T) (value T, ok bool) {
	scope := ctx.GenerationScope()
	if scope == nil {
		return value, false
	}
	scope.mu.Lock()
	entry := scope.values[key]
	if entry == nil {
		entry = &generationValue{}
		scope.values[key] = entry
	}
	scope.mu.Unlock()
	entry.once.Do(func() { entry.value = build(scope) })
	value, _ = entry.value.(T)
	return value, true
}
