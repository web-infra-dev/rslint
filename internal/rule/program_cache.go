package rule

import "github.com/web-infra-dev/rslint/internal/program"

// CachedByProgram shares immutable derived state across every rule context in
// one source generation. Cache lifetime and backend ownership are encapsulated
// by Program; rules only provide a configuration-complete key and builder.
func CachedByProgram[T any](ctx RuleContext, key any, build func() T) T {
	return program.Cached(ctx.Program(), key, build)
}
