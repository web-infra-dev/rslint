package rule

import "sync"

// ProgramStore is one lint run's memo for whole-Program state. A rule whose
// work is a property of the Program rather than of one file — a module graph,
// a cross-file index — builds it once here, and every later file in the run
// reuses it instead of rebuilding it per file.
//
// Keys are opaque: a rule keys by whatever configuration changes the value it
// stores, so two rule configurations that disagree about the shape of that
// value get their own entries.
type ProgramStore struct {
	mu      sync.Mutex
	entries map[any]*programEntry
}

type programEntry struct {
	once  sync.Once
	value any
}

func NewProgramStore() *ProgramStore {
	return &ProgramStore{entries: make(map[any]*programEntry)}
}

// Load returns the value memoized under key, calling build on the first
// request for it. build runs at most once per key even when rules on
// different files run concurrently; the other callers wait for it and observe
// the same value. A nil store calls build directly and memoizes nothing, so a
// context assembled without one still produces correct results.
func (store *ProgramStore) Load(key any, build func() any) any {
	if store == nil {
		return build()
	}

	store.mu.Lock()
	entry, ok := store.entries[key]
	if !ok {
		entry = &programEntry{}
		store.entries[key] = entry
	}
	store.mu.Unlock()

	entry.once.Do(func() { entry.value = build() })
	return entry.value
}
