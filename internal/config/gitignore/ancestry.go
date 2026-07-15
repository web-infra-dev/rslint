package gitignore

import (
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/vfs"
)

// AncestrySource is the immutable contribution of one directory to an exact
// target's Git-ignore ancestry. Globs are already rooted at the collector's
// base directory. A blocked source inherits its parent's rules but must not be
// read, nor may any descendant source be read.
type AncestrySource struct {
	Directory string
	Parent    string
	Globs     []string
	Blocked   bool
}

type ancestrySourceState struct {
	ready      chan struct{}
	source     AncestrySource
	pruneRules []gitignorePruneRule
}

// AncestryCollector snapshots each .gitignore source at most once while
// retaining Git's parent-prune and descendant-symlink rules. Independent
// sibling sources load concurrently after their shared parent state is ready.
type AncestryCollector struct {
	root          string
	fs            vfs.FS
	caseSensitive bool

	mu      sync.Mutex
	sources map[string]*ancestrySourceState
}

func NewAncestryCollector(root string, fsys vfs.FS) *AncestryCollector {
	root = normalizeGlobPath(root)
	caseSensitive := true
	if fsys != nil {
		caseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	return &AncestryCollector{
		root:          root,
		fs:            fsys,
		caseSensitive: caseSensitive,
		sources:       make(map[string]*ancestrySourceState),
	}
}

// Source returns the one-directory contribution for directory. The returned
// value and its Globs slice are immutable and may be shared by concurrent
// evaluator queries.
func (collector *AncestryCollector) Source(directory string) AncestrySource {
	if collector == nil || collector.fs == nil {
		return AncestrySource{Blocked: true}
	}
	directory = normalizeGlobPath(directory)
	if _, within := relativeDir(collector.root, directory, collector.caseSensitive); !within {
		return AncestrySource{Directory: directory, Blocked: true}
	}
	state := collector.sourceState(directory)
	<-state.ready
	return state.source
}

func (collector *AncestryCollector) sourceState(directory string) *ancestrySourceState {
	identity := directory
	if !collector.caseSensitive {
		identity = strings.ToLower(identity)
	}

	collector.mu.Lock()
	if existing := collector.sources[identity]; existing != nil {
		collector.mu.Unlock()
		return existing
	}
	state := &ancestrySourceState{ready: make(chan struct{})}
	collector.sources[identity] = state
	collector.mu.Unlock()

	state.source.Directory = directory
	relative, _ := relativeDir(collector.root, directory, collector.caseSensitive)
	if relative != "" {
		parent := parentDir(directory)
		state.source.Parent = parent
		parentState := collector.sourceState(parent)
		<-parentState.ready
		state.pruneRules = parentState.pruneRules
		state.source.Blocked = parentState.source.Blocked
		if !state.source.Blocked {
			state.source.Blocked = isDescendantSymlinkDir(collector.root, directory, collector.fs) ||
				isDirIgnoredByPruneRules(directory, parentState.pruneRules, collector.caseSensitive)
		}
	}

	if !state.source.Blocked {
		if content, ok := collector.fs.ReadFile(joinHostPath(directory, ".gitignore")); ok {
			state.source.Globs = convertGitignoreToGlobs(content, relative)
			if rule, ok := newGitignorePruneRule(directory, content); ok {
				// Parent state is immutable after ready closes. Copy-on-append keeps
				// siblings from sharing a writable spare-capacity backing array.
				state.pruneRules = append(append([]gitignorePruneRule(nil), state.pruneRules...), rule)
			}
		}
	}
	close(state.ready)
	return state
}
