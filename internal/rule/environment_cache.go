package rule

type environmentCacheKey struct {
	environment *RuleEnvironment
	key         any
}

// CachedByEnvironment shares immutable configuration-derived values across
// files using the same environment and Program generation. Keys must include
// other construction inputs, such as cwd, and use a consumer-owned named type.
// Values must not retain a Program, RuleContext, or file-specific state.
//
// Manually assembled contexts without a bound environment call build directly,
// allowing consumers to retain their content-based compatibility cache.
func CachedByEnvironment[T any](ctx RuleContext, key any, build func() T) T {
	if ctx.fileCache == nil || ctx.fileCache.environment == nil {
		return build()
	}
	return CachedByProgram(ctx, environmentCacheKey{ctx.fileCache.environment, key}, build)
}
