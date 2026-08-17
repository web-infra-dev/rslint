package rule

// FileCache holds file-pass shared values for the duration of one source
// file's lint. Rules on one file run serially, so the cache intentionally does
// not synchronize access.
type FileCache struct {
	values                  map[any]any
	processCurrentDirectory string
}

// NewFileCache creates the cache that the linter shares with every rule on one
// file. Rules should consume it through CachedByFile rather than retaining it.
func NewFileCache() *FileCache {
	return &FileCache{}
}

// NewFileCacheWithProcessCurrentDirectory creates the per-file shared state
// used by the linter and records the process working directory once per file
// instead of copying its string into every rule context.
func NewFileCacheWithProcessCurrentDirectory(currentDirectory string) *FileCache {
	return &FileCache{processCurrentDirectory: currentDirectory}
}

// WithFileCache returns a context attached to cache. This is a linter assembly
// hook; rules should use CachedByFile to cache values derived from SourceFile.
func (ctx RuleContext) WithFileCache(cache *FileCache) RuleContext {
	ctx.fileCache = cache
	return ctx
}

// CachedByFile returns the value stored for the current file under key,
// calling build on the first request. A manually constructed RuleContext with
// no file cache calls build every time, preserving direct rule-test behavior.
//
// Callers must use an unexported comparable key type so keys from different
// packages cannot collide. Cached values must be derived from the current
// file, must not retain rule-specific settings or reporting state, and may
// rely on the linter's guarantee that listeners for one file run serially.
func CachedByFile[T any](ctx RuleContext, key any, build func() T) T {
	cache := ctx.fileCache
	if cache == nil {
		return build()
	}
	if value, ok := cache.values[key]; ok {
		result, _ := value.(T)
		return result
	}
	value := build()
	if cache.values == nil {
		cache.values = make(map[any]any)
	}
	cache.values[key] = value
	return value
}
