package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/collections"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/project"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/trackingvfs"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
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
		seenFiles:     pathutil.Strings(tracker.SeenFiles.ToSlice()),
		failedLookups: failedLookups.ToSlice(),
	}
}

func (fs *lintTrackingFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	content, ok := fs.FS.ReadFile(path)
	if !ok {
		fs.failedLookups.Add(path.AsString())
	}
	return content, ok
}

func (fs *lintTrackingFS) FileExists(path tspath.RootedFilePath) bool {
	exists := fs.FS.FileExists(path)
	if !exists {
		fs.failedLookups.Add(path.AsString())
	}
	return exists
}

func lintProgramDirectoryWatchInput(path string) string {
	return tspath.CombinePaths(path, "__rslint_watch__")
}

func (fs *lintTrackingFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	exists := fs.Inner.DirectoryExists(path)
	if !exists {
		fs.failedLookups.Add(path.AsString())
		fs.SeenFiles.Add(pathutil.Path(lintProgramDirectoryWatchInput(path.AsString())))
	}
	return exists
}

func (fs *lintTrackingFS) GetAccessibleEntries(path tspath.RootedDirectoryPath) vfs.Entries {
	fs.SeenFiles.Add(pathutil.Path(lintProgramDirectoryWatchInput(path.AsString())))
	return fs.Inner.GetAccessibleEntries(path)
}

func (fs *lintTrackingFS) Stat(path tspath.RootedPath) vfs.FileInfo {
	info := fs.Inner.Stat(path)
	if info != nil && info.IsDir() {
		if !pathutil.IsDiskPathRoot(path.AsString()) {
			fs.SeenFiles.Add(pathutil.Path(lintProgramDirectoryWatchInput(path.AsString())))
		}
	} else {
		fs.SeenFiles.Add(path)
		if info == nil {
			fs.failedLookups.Add(path.AsString())
		}
	}
	return info
}

func (fs *lintTrackingFS) Realpath(path tspath.RootedPath) tspath.RootedPath {
	realPath := fs.Inner.Realpath(path)
	if fs.Inner.DirectoryExists(tspath.RootedDirectoryPathFromPath(path)) {
		if !pathutil.IsDiskPathRoot(path.AsString()) {
			fs.SeenFiles.Add(pathutil.Path(lintProgramDirectoryWatchInput(path.AsString())))
		}
		if realPath != "" && !pathutil.IsDiskPathRoot(realPath.AsString()) {
			fs.SeenFiles.Add(pathutil.Path(lintProgramDirectoryWatchInput(realPath.AsString())))
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

	inputs := &collections.SyncMap[tspath.PathKey, tspath.RootedFilePath]{}
	caseSensitive := c.server.fs.CaseSensitivity().IsCaseSensitive()
	currentDirectoryPath := pathutil.ToPath(
		currentDirectory,
		currentDirectory,
		caseSensitive,
	)
	watchKind := lsproto.WatchKindCreate |
		lsproto.WatchKindChange |
		lsproto.WatchKindDelete
	exactWatchers := make([]*lsproto.FileSystemWatcher, 0)
	for _, fileName := range seenFiles {
		path := pathutil.ToPath(fileName, currentDirectory, caseSensitive)
		if pathutil.IsDiskPathRoot(string(path)) {
			continue
		}
		// Module detection probes package.json through every ancestor. Keep
		// those probes exact so an ancestor at the volume root does not force a
		// recursive root watcher.
		if tspath.GetBaseFileName(string(path)) == "package.json" &&
			tspath.ContainsPath(path.Parent().AsString(), currentDirectoryPath.AsString(), tspath.CaseSensitive) {
			watcher := fileSystemWatcher(
				string(path.Parent()),
				"package.json",
				c.relativePatternSupport,
			)
			watcher.Kind = &watchKind
			exactWatchers = append(exactWatchers, watcher)
			continue
		}
		inputs.Store(path, tspath.ToRootedFilePath(fileName, pathutil.Directory(currentDirectory)))
	}

	mapPatterns := project.CreateResolutionLookupGlobMapper(
		pathutil.Directory(c.server.cwd),
		pathutil.Directory(c.server.defaultLibraryPath),
		pathutil.Directory(currentDirectory),
		pathutil.CaseSensitivity(caseSensitive),
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
