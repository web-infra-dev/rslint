package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/microsoft/typescript-go/shim/collections"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/trackingvfs"
)

// lintTrackingFS keeps the CompilerHost stable while swapping the upstream
// tracking wrapper after each Program generation.
type lintTrackingFS struct {
	*trackingvfs.FS
	failedLookups *collections.SyncSet[string]
}

type lintProgramLookups struct {
	seenFiles     []string
	failedLookups []string
}

func newLintTrackingFS(inner vfs.FS) *lintTrackingFS {
	fs := &lintTrackingFS{}
	fs.replace(inner)
	return fs
}

func (fs *lintTrackingFS) replace(inner vfs.FS) {
	fs.FS = &trackingvfs.FS{Inner: inner}
	fs.failedLookups = &collections.SyncSet[string]{}
}

func (fs *lintTrackingFS) reset() {
	fs.replace(fs.Inner)
}

func (fs *lintTrackingFS) drain() lintProgramLookups {
	tracker := fs.FS
	failedLookups := fs.failedLookups
	fs.reset()
	return lintProgramLookups{
		seenFiles:     tracker.SeenFiles.ToSlice(),
		failedLookups: failedLookups.ToSlice(),
	}
}

func (fs *lintTrackingFS) ReadFile(path string) (string, bool) {
	content, ok := fs.FS.ReadFile(path)
	if !ok {
		fs.failedLookups.Add(path)
	}
	return content, ok
}

func (fs *lintTrackingFS) FileExists(path string) bool {
	exists := fs.FS.FileExists(path)
	if !exists {
		fs.failedLookups.Add(path)
	}
	return exists
}

func lintProgramDirectoryWatchInput(path string) string {
	return tspath.CombinePaths(path, "__rslint_watch__")
}

func (fs *lintTrackingFS) DirectoryExists(path string) bool {
	exists := fs.Inner.DirectoryExists(path)
	if !exists {
		fs.failedLookups.Add(path)
		fs.SeenFiles.Add(lintProgramDirectoryWatchInput(path))
	}
	return exists
}

func (fs *lintTrackingFS) GetAccessibleEntries(path string) vfs.Entries {
	fs.SeenFiles.Add(lintProgramDirectoryWatchInput(path))
	return fs.Inner.GetAccessibleEntries(path)
}

func (fs *lintTrackingFS) Stat(path string) vfs.FileInfo {
	info := fs.Inner.Stat(path)
	if info != nil && info.IsDir() {
		if !tspath.IsDiskPathRoot(path) {
			fs.SeenFiles.Add(lintProgramDirectoryWatchInput(path))
		}
	} else {
		fs.SeenFiles.Add(path)
		if info == nil {
			fs.failedLookups.Add(path)
		}
	}
	return info
}

func (fs *lintTrackingFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	fs.SeenFiles.Add(lintProgramDirectoryWatchInput(root))
	fs.DirectoryExists(root)
	return fs.Inner.WalkDir(root, func(
		path string,
		entry vfs.DirEntry,
		err error,
	) error {
		if err == nil && entry != nil && entry.IsDir() {
			fs.SeenFiles.Add(lintProgramDirectoryWatchInput(path))
		} else {
			fs.SeenFiles.Add(path)
		}
		return walkFn(path, entry, err)
	})
}

func (fs *lintTrackingFS) Realpath(path string) string {
	realPath := fs.Inner.Realpath(path)
	if fs.Inner.DirectoryExists(path) {
		if !tspath.IsDiskPathRoot(path) {
			fs.SeenFiles.Add(lintProgramDirectoryWatchInput(path))
		}
		if realPath != "" && !tspath.IsDiskPathRoot(realPath) {
			fs.SeenFiles.Add(lintProgramDirectoryWatchInput(realPath))
		}
	} else {
		fs.SeenFiles.Add(path)
		if realPath != "" {
			fs.SeenFiles.Add(realPath)
		}
	}
	return realPath
}

// lintProgramCoverage maps standalone compiler reads through tsgo's own
// resolution watcher policy. It contains no Program lifecycle state.
type lintProgramCoverage struct {
	server                 *Server
	watchFiles             func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error
	coveredWatchers        map[string]struct{}
	relativePatternSupport bool
	disabled               bool
	nextWatcherID          uint64
}

func newLintProgramCoverage(server *Server) *lintProgramCoverage {
	relativePatternSupport := false
	if server.initializeParams != nil {
		capabilities := server.initializeParams.Capabilities
		if capabilities != nil &&
			capabilities.Workspace != nil &&
			capabilities.Workspace.DidChangeWatchedFiles != nil {
			relativePatternSupport = ptrIsTrue(
				capabilities.Workspace.DidChangeWatchedFiles.RelativePatternSupport,
			)
		}
	}
	return &lintProgramCoverage{
		server:                 server,
		watchFiles:             server.WatchFiles,
		coveredWatchers:        make(map[string]struct{}),
		relativePatternSupport: relativePatternSupport,
	}
}

func (c *lintProgramCoverage) ensure(
	ctx context.Context,
	seenFiles []string,
	currentDirectory string,
) (added bool, safe bool, err error) {
	if len(seenFiles) == 0 {
		return false, true, nil
	}

	inputs := &collections.SyncSet[tspath.Path]{}
	caseSensitive := c.server.fs.UseCaseSensitiveFileNames()
	currentDirectoryPath := tspath.ToPath(
		currentDirectory,
		currentDirectory,
		caseSensitive,
	)
	watchKind := lsproto.WatchKindCreate |
		lsproto.WatchKindChange |
		lsproto.WatchKindDelete
	exactWatchers := make([]*lsproto.FileSystemWatcher, 0)
	for _, fileName := range seenFiles {
		path := tspath.ToPath(fileName, currentDirectory, caseSensitive)
		if tspath.IsDiskPathRoot(string(path)) {
			continue
		}
		// Module detection probes package.json through every ancestor. Keep
		// those probes exact so an ancestor at the volume root does not force a
		// recursive root watcher.
		if tspath.GetBaseFileName(string(path)) == "package.json" &&
			path.GetDirectoryPath().ContainsPath(currentDirectoryPath) {
			watcher := fileSystemWatcher(
				string(path.GetDirectoryPath()),
				"package.json",
				c.relativePatternSupport,
			)
			watcher.Kind = &watchKind
			exactWatchers = append(exactWatchers, watcher)
			continue
		}
		inputs.Add(path)
	}

	mapPatterns := project.CreateResolutionLookupGlobMapper(
		c.server.cwd,
		c.server.defaultLibraryPath,
		currentDirectory,
		caseSensitive,
	)
	watchers := project.NewWatchedFiles(
		"rslint standalone programs",
		watchKind,
		c.relativePatternSupport,
		mapPatterns,
	).Clone(inputs).Watchers()
	if len(watchers.IgnoredPaths) != 0 {
		return false, false, nil
	}

	candidates := append(exactWatchers, watchers.WorkspaceWatchers...)
	candidates = append(candidates, watchers.OutsideWorkspaceWatchers...)
	newWatchers := make([]*lsproto.FileSystemWatcher, 0, len(candidates))
	newKeys := make([]string, 0, len(candidates))
	for _, watcher := range candidates {
		key := lintProgramWatcherKey(watcher)
		if _, covered := c.coveredWatchers[key]; covered {
			continue
		}
		newWatchers = append(newWatchers, watcher)
		newKeys = append(newKeys, key)
	}
	if len(newWatchers) == 0 {
		return false, true, nil
	}

	c.nextWatcherID++
	watcherID := project.WatcherID(fmt.Sprintf(
		"rslint-standalone-program-%d",
		c.nextWatcherID,
	))
	watchCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := c.watchFiles(watchCtx, watcherID, newWatchers); err != nil {
		return false, false, err
	}
	for _, key := range newKeys {
		c.coveredWatchers[key] = struct{}{}
	}
	return true, true, nil
}

func lintProgramWatcherKey(watcher *lsproto.FileSystemWatcher) string {
	kind := lsproto.WatchKindCreate | lsproto.WatchKindChange | lsproto.WatchKindDelete
	if watcher.Kind != nil {
		kind = *watcher.Kind
	}
	return fmt.Sprintf("%d:%s", kind, project.FileSystemWatcherGlobString(watcher))
}
