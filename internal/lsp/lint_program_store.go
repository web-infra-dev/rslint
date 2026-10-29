package lsp

import (
	"context"
	"maps"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// lintProgramStore owns the configured Programs selected by open documents.
// Explicit and discovered projects share the same editor construction semantics.
type lintProgramStore struct {
	server              *Server
	documentPrograms    map[lsproto.DocumentUri]lintProgramKey
	parseCache          *utils.ParseCache
	coverage            *lintProgramCoverage
	programs            map[lintProgramKey]*lintProgramState
	projectMetadata     map[string]*lintProjectMetadata
	observedRootConfigs map[string]struct{}
}

type lintProgramKey = tspath.Path

type lintProgramState struct {
	program    *compiler.Program
	tracker    *lintTrackingFS
	sources    *utils.ProgramSourceLookup
	dirtyFiles map[tspath.Path]struct{}
	// Failed lookups close the gap between a file appearing and the client's
	// corresponding watcher notification.
	failedLookups            map[tspath.Path]struct{}
	selectedSourceIdentities map[tspath.Path]tspath.Path
	metadata                 *lintProjectMetadata
}

type lintProgramRequest struct {
	store             *lintProgramStore
	ctx               context.Context
	uri               lsproto.DocumentUri
	overlayFiles      map[string]string
	overlayFS         vfs.FS
	discoveryFS       *lintTrackingFS
	target            target.File
	freshOnly         bool
	overlayPrepared   bool
	usedKey           lintProgramKey
	usedState         *lintProgramState
	projectMetadata   map[string]*lintProjectMetadata
	transientMetadata map[string]struct{}
}

func newLintProgramStore(server *Server) *lintProgramStore {
	return &lintProgramStore{
		server:              server,
		documentPrograms:    make(map[lsproto.DocumentUri]lintProgramKey),
		parseCache:          utils.NewParseCache(),
		coverage:            newLintProgramCoverage(server),
		programs:            make(map[lintProgramKey]*lintProgramState),
		projectMetadata:     make(map[string]*lintProjectMetadata),
		observedRootConfigs: make(map[string]struct{}),
	}
}

func (s *lintProgramStore) Usable() bool {
	return s != nil && !s.coverage.disabled
}

func (s *lintProgramStore) Request(
	ctx context.Context,
	uri lsproto.DocumentUri,
	target target.File,
) (lintProgramLoader, lintProjectMetadataLoader, func()) {
	request := s.request(ctx, uri, target)
	return request.load, request.loadMetadata, request.finalize
}

func (s *lintProgramStore) request(
	ctx context.Context,
	uri lsproto.DocumentUri,
	target target.File,
) *lintProgramRequest {
	target.Path = tspath.NormalizePath(target.Path)
	if target.CanonicalPath != "" {
		target.CanonicalPath = tspath.NormalizePath(target.CanonicalPath)
	}
	if target.CanonicalParentPath != "" {
		target.CanonicalParentPath = tspath.NormalizePath(target.CanonicalParentPath)
	}
	return &lintProgramRequest{
		store:             s,
		ctx:               ctx,
		uri:               uri,
		target:            target,
		projectMetadata:   make(map[string]*lintProjectMetadata),
		transientMetadata: make(map[string]struct{}),
	}
}

func (r *lintProgramRequest) key(configPath string) lintProgramKey {
	return lintProgramLexicalPathID(configPath, r.store.server.fs)
}

func (r *lintProgramRequest) createProgram(metadata *lintProjectMetadata, fsys vfs.FS) (*compiler.Program, error) {
	commandLine := singleCheckerLintCommandLine(metadata.commandLine)
	host := utils.WithASTCache(utils.CreateCompilerHost(tspath.GetDirectoryPath(metadata.configPath), fsys), r.store.parseCache)
	return utils.CreateProgramFromParsedConfigLenientWithProjectReferences(false, commandLine,
		newLintProjectReferenceHost(host, r.projectMetadata))
}

func (r *lintProgramRequest) loadMetadata(
	configFileName string,
) (*lintProjectMetadata, bool, error) {
	metadata, err := r.metadata(configFileName)
	return metadata, err == nil && metadata != nil, err
}

func (r *lintProgramRequest) prepareOverlay() {
	if r.overlayPrepared {
		return
	}
	r.overlayPrepared = true
	server := r.store.server
	content, open := server.documents[r.uri]
	files, aliasesConflict := server.currentEditorOverlayFilesForFrozenTarget(r.uri, r.target, content, open)
	r.overlayFiles = files
	r.overlayFS = newFrozenLintTargetOverlayFS(server.fs, files, r.target)
	r.discoveryFS = newLintTrackingFS(r.overlayFS)
	if aliasesConflict {
		r.store.Invalidate()
		r.freshOnly = true
	}
}

// freshFS shares repeated disk and projected-overlay queries within one
// observation. A new view after watcher registration or an editor event keeps
// cached misses and symlink identities out of the next observation.
func (r *lintProgramRequest) freshFS() vfs.FS {
	disk := cachedvfs.From(r.store.server.fs)
	return cachedvfs.From(newFrozenLintTargetOverlayFS(disk, r.overlayFiles, r.target))
}

func (r *lintProgramRequest) metadata(
	configFileName string,
) (*lintProjectMetadata, error) {
	r.prepareOverlay()
	configFileName = tspath.NormalizePath(configFileName)
	if metadata := r.projectMetadata[configFileName]; metadata != nil {
		return metadata, nil
	}
	if !r.freshOnly && r.store.Usable() {
		if state := r.store.programs[r.key(configFileName)]; state != nil && state.metadata != nil {
			r.projectMetadata[configFileName] = state.metadata
			return state.metadata, nil
		}
		if metadata := r.store.projectMetadata[configFileName]; metadata != nil {
			r.projectMetadata[configFileName] = metadata
			return metadata, nil
		}
	}

	r.store.observedRootConfigs[configFileName] = struct{}{}
	metadata, retain, err := r.parseProjectMetadata(configFileName)
	if err != nil {
		return nil, err
	}
	r.projectMetadata[configFileName] = metadata
	if retain {
		r.store.projectMetadata[configFileName] = metadata
	} else {
		r.transientMetadata[configFileName] = struct{}{}
	}
	return metadata, nil
}

func (r *lintProgramRequest) parseProjectMetadata(
	configFileName string,
) (*lintProjectMetadata, bool, error) {
	if r.freshOnly || !r.store.Usable() {
		metadata, err := parseStandaloneLintProject(
			configFileName,
			r.overlayFS,
			r.store.server.fs,
		)
		return metadata, false, err
	}

	for attempt := range 2 {
		tracker := newLintTrackingFS(r.freshFS())
		metadata, err := parseStandaloneLintProject(
			configFileName,
			tracker,
			r.store.server.fs,
		)
		if err != nil {
			return nil, false, err
		}
		lookups := tracker.drain()
		added, safe, _ := r.store.coverage.ensure(
			r.ctx,
			lookups.seenFiles,
			tspath.GetDirectoryPath(configFileName),
		)
		if !safe {
			r.store.coverage.disabled = true
			r.store.Invalidate()
			r.freshOnly = true
			return metadata, false, nil
		}
		if added && attempt == 0 {
			continue
		}
		if added {
			// The second parse expanded watcher coverage again. Serve this
			// request, but do not retain metadata from the uncovered interval.
			return metadata, false, nil
		}
		return metadata, true, nil
	}
	panic("unreachable")
}

func (r *lintProgramRequest) load(
	metadata *lintProjectMetadata,
) (*compiler.Program, *ast.SourceFile, error) {
	r.prepareOverlay()
	configFileName := metadata.configPath
	if r.freshOnly || !r.store.Usable() {
		return r.loadFresh(metadata)
	}

	state := r.store.programs[r.key(configFileName)]
	if state == nil || !r.matchesMetadata(state, metadata) {
		return r.rebuild(configFileName, metadata)
	}

	targetSource := state.sources.SourceFileForTarget(
		r.target.Path,
		r.target.CanonicalPath,
	)
	if content, open := r.store.server.documents[r.uri]; open {
		r.store.markStateSourceContent(state, targetSource, content)
	}
	// Cache filesystem queries only within this editor observation. The stable
	// host must see a fresh view on every request, including incremental updates.
	state.tracker.Inner = r.freshFS()
	if len(state.dirtyFiles) == 0 {
		return r.result(configFileName, state)
	}

	state.tracker.reset()
	program := state.program
	allGraphsReused := true
	for _, dirtyPath := range sortedLintProgramPaths(state.dirtyFiles) {
		oldSourceFile := program.GetSourceFileByPath(dirtyPath)
		if oldSourceFile == nil {
			return r.rebuild(configFileName, state.metadata)
		}
		updated, newSourceFile, graphReused :=
			program.UpdateProgram(dirtyPath, program.Host(), nil)
		if updated == nil {
			return r.rebuild(configFileName, state.metadata)
		}
		updated.BindSourceFiles()
		if graphReused &&
			!lintProgramGraphReusable(program, updated, oldSourceFile, newSourceFile) {
			return r.rebuild(configFileName, state.metadata)
		}
		program = updated
		if !graphReused {
			allGraphsReused = false
			clear(state.failedLookups)
			clear(state.selectedSourceIdentities)
			break
		}
	}

	state.program = program
	if allGraphsReused {
		state.sources.UpdateProgramForReusedGraph(program, r.overlayFS)
	} else {
		state.sources = utils.NewProgramSourceLookup(program, r.overlayFS)
	}
	clear(state.dirtyFiles)
	added, safe := r.cover(state)
	if !safe {
		return program, state.sources.SourceFileForTarget(
			r.target.Path,
			r.target.CanonicalPath,
		), nil
	}
	if added {
		return r.rebuild(configFileName, state.metadata)
	}
	return r.result(configFileName, state)
}

// matchesMetadata checks only snapshots this request has already observed.
// Watchers still own invalidation; a newly observed reference may not be
// discarded merely because the resident root snapshot is unchanged.
func (r *lintProgramRequest) matchesMetadata(state *lintProgramState, root *lintProjectMetadata) bool {
	if state.metadata != root || len(r.transientMetadata) != 0 {
		return false
	}
	rootPath := lintProgramLexicalPathID(root.configPath, r.store.server.fs)
	for _, metadata := range r.projectMetadata {
		path := lintProgramLexicalPathID(metadata.configPath, r.store.server.fs)
		if path == rootPath {
			continue
		}
		if built, referenced := state.program.GetResolvedProjectReferenceFor(path); referenced &&
			!lintProjectSnapshotsEqual(metadata.commandLine, built, r.store.server.fs) {
			return false
		}
	}
	return true
}

func (r *lintProgramRequest) loadFresh(
	metadata *lintProjectMetadata,
) (*compiler.Program, *ast.SourceFile, error) {
	program, err := r.createProgram(metadata, r.freshFS())
	if err != nil {
		return nil, nil, err
	}
	return program, sourceFileForTarget(program, r.target, r.overlayFS), nil
}

func (r *lintProgramRequest) rebuild(
	configFileName string,
	metadata *lintProjectMetadata,
) (*compiler.Program, *ast.SourceFile, error) {
	delete(r.store.programs, r.key(configFileName))
	if metadata == nil {
		var err error
		metadata, err = r.metadata(configFileName)
		if err != nil {
			return nil, nil, err
		}
	}
	// A selected reference may also predate stable watcher coverage. The host
	// can retain any successful metadata from this request, so none of those
	// snapshots may cross the request boundary while coverage is unsettled.
	if r.freshOnly || !r.store.Usable() || len(r.transientMetadata) != 0 {
		program, err := r.createProgram(metadata, r.freshFS())
		if err != nil {
			return nil, nil, err
		}
		return program, sourceFileForTarget(program, r.target, r.overlayFS), nil
	}

	r.store.coverage.prepareProgram(r.ctx, tspath.GetDirectoryPath(configFileName))

	// Register any dependencies outside the prepared coverage, then rebuild
	// once so no filesystem change can fall into the registration gap.
	for attempt := range 2 {
		// Each retry starts after watcher registration with a fresh filesystem
		// view, so cached misses cannot hide a change in the registration gap.
		tracker := newLintTrackingFS(r.freshFS())
		program, err := r.createProgram(metadata, tracker)
		if err != nil {
			return nil, nil, err
		}
		state := &lintProgramState{
			program:                  program,
			tracker:                  tracker,
			sources:                  utils.NewProgramSourceLookup(program, r.overlayFS),
			dirtyFiles:               make(map[tspath.Path]struct{}),
			failedLookups:            make(map[tspath.Path]struct{}),
			selectedSourceIdentities: make(map[tspath.Path]tspath.Path),
			metadata:                 metadata,
		}
		sourceFile := state.sources.SourceFileForTarget(
			r.target.Path,
			r.target.CanonicalPath,
		)
		if sourceFile == nil {
			// A fallback probe that did not contain this target must not make an
			// unrelated Program resident. Previously selected resident Programs
			// are handled by load and remain available to their documents.
			return program, nil, nil
		}
		added, safe := r.cover(state)
		if !safe {
			return program, sourceFile, nil
		}
		if added && attempt == 0 {
			continue
		}
		if added {
			// Lazy compiler reads expanded coverage again. Serve this Program
			// once, but do not retain it across the uncovered interval.
			return program, sourceFile, nil
		}
		delete(r.store.projectMetadata, configFileName)
		r.store.programs[r.key(configFileName)] = state
		return r.result(configFileName, state)
	}
	panic("unreachable")
}

func (r *lintProgramRequest) result(
	configFileName string,
	state *lintProgramState,
) (*compiler.Program, *ast.SourceFile, error) {
	sourceFile := state.sources.SourceFileForTarget(
		r.target.Path,
		r.target.CanonicalPath,
	)
	if sourceFile != nil {
		state.rememberSelectedTarget(r.target, r.store.server.fs)
		r.usedKey = r.key(configFileName)
		r.usedState = state
	}
	return state.program, sourceFile, nil
}

func (r *lintProgramRequest) cover(state *lintProgramState) (added bool, safe bool) {
	lookups := state.tracker.drain()
	state.recordFailedLookups(lookups, r.store.server.fs)
	added, safe, err := r.store.coverage.ensure(
		r.ctx,
		lookups.seenFiles,
		state.program.GetCurrentDirectory(),
	)
	if err != nil || !safe {
		r.store.coverage.disabled = true
		r.store.Invalidate()
		return false, false
	}
	return added, true
}

func (r *lintProgramRequest) finalize() {
	defer r.store.prune()
	delete(r.store.documentPrograms, r.uri)
	if !r.store.Usable() {
		return
	}
	if r.discoveryFS != nil {
		lookups := r.discoveryFS.drain()
		added, safe, err := r.store.coverage.ensure(r.ctx, lookups.seenFiles, r.store.server.cwd)
		if err != nil || !safe {
			r.store.coverage.disabled = true
			r.store.Invalidate()
			return
		}
		if added {
			// Re-select after registering missing-config probes, closing the interval
			// in which a new config could appear before its watcher became active.
			r.store.Invalidate()
			_ = r.store.server.RefreshDiagnostics(r.ctx)
			return
		}
	}
	if r.usedState == nil {
		return
	}
	added, safe := r.cover(r.usedState)
	if !safe || added {
		delete(r.store.programs, r.usedKey)
		return
	}
	if _, open := r.store.server.documents[r.uri]; open && r.store.programs[r.usedKey] == r.usedState {
		r.store.documentPrograms[r.uri] = r.usedKey
	}
}

func (s *Server) releaseDocumentProgram(uri lsproto.DocumentUri) {
	if s.lintPrograms != nil {
		delete(s.lintPrograms.documentPrograms, uri)
		s.lintPrograms.prune()
	}
}

// prune bounds both Program and parsed-source residency to document ownership.
// Requests already holding a generation keep their own Go references.
func (s *lintProgramStore) prune() {
	live := make(map[lintProgramKey]bool, len(s.documentPrograms))
	for _, key := range s.documentPrograms {
		live[key] = true
	}
	programs := make([]*compiler.Program, 0, len(s.programs))
	for key, state := range s.programs {
		if !live[key] {
			delete(s.programs, key)
		} else {
			programs = append(programs, state.program)
		}
	}
	s.parseCache.RetainOnly(programs)
	if len(s.documentPrograms) == 0 {
		clear(s.projectMetadata)
		clear(s.observedRootConfigs)
	}
}

func (s *lintProgramStore) DidOpen(
	uri lsproto.DocumentUri,
	content string,
	existedOnDisk bool,
) {
	if !existedOnDisk {
		s.Invalidate()
		return
	}
	s.markContent(uriToPath(uri), content, true)
}

func (s *lintProgramStore) DidChange(uri lsproto.DocumentUri, content string) {
	s.markContent(uriToPath(uri), content, false)
}

func (s *lintProgramStore) DidSave(uri lsproto.DocumentUri, open bool) {
	if !open {
		s.Invalidate()
		return
	}
	s.markContent(uriToPath(uri), s.server.documents[uri], false)
}

func (s *lintProgramStore) DidClose(uri lsproto.DocumentUri) {
	delete(s.documentPrograms, uri)
	defer s.prune()
	content, ok := s.server.fs.ReadFile(uriToPath(uri))
	if !ok {
		s.Invalidate()
		return
	}
	s.markContent(uriToPath(uri), content, false)
}

// DidChangeWatchedFiles returns whether diagnostics need to be refreshed. A
// custom project path may previously have failed to resolve, leaving no
// resident Program even though a filesystem change alters its lint policy.
func (s *lintProgramStore) DidChangeWatchedFiles(
	changes []*lsproto.FileEvent,
) bool {
	needsRefresh := false
	for _, change := range changes {
		if s.isOpenSourceOverlayWatchChange(change) {
			uri := change.Uri
			s.markContent(
				uriToPath(uri),
				s.server.documents[uri],
				false,
			)
			// A dependent document can still display older diagnostics even
			// when its Program has already incorporated the saved buffer.
			needsRefresh = true
			continue
		}
		invalidated := s.Invalidate()
		return invalidated || needsRefresh || len(s.server.documents) > 0
	}
	return needsRefresh
}

func (s *lintProgramStore) isOpenSourceOverlayWatchChange(
	change *lsproto.FileEvent,
) bool {
	if change == nil || change.Type != lsproto.FileChangeTypeChanged {
		return false
	}
	uri := change.Uri
	_, open := s.server.documents[uri]
	return open && isLintableScriptFile(uri)
}

func (s *lintProgramStore) Invalidate() bool {
	hadState := len(s.programs) != 0 ||
		len(s.projectMetadata) != 0 ||
		len(s.observedRootConfigs) != 0
	clear(s.programs)
	clear(s.documentPrograms)
	s.parseCache.RetainOnly(nil)
	clear(s.projectMetadata)
	clear(s.observedRootConfigs)
	return hadState
}

func (s *lintProgramStore) markContent(
	fileName string,
	content string,
	invalidatePossibleNewSource bool,
) {
	if tspath.HasJSONFileExtension(fileName) {
		s.Invalidate()
		return
	}
	if invalidatePossibleNewSource {
		for configFileName, metadata := range s.projectMetadata {
			if metadata != nil && metadata.commandLine != nil &&
				metadata.commandLine.PossiblyMatchesFileName(fileName) {
				delete(s.projectMetadata, configFileName)
			}
		}
	}
	lexicalID := lintProgramLexicalPathID(fileName, s.server.fs)
	var physicalID tspath.Path
	physicalPathID := func() tspath.Path {
		if physicalID == "" {
			// An unsaved dependency can live beneath a directory alias even
			// when its missing suffix cannot be resolved by disk Realpath.
			identity := config.FreezeFileIdentity(fileName, s.server.fs)
			physicalID = lintProgramLexicalPathID(identity.CanonicalPath, s.server.fs)
		}
		return physicalID
	}
	for configFileName, state := range s.programs {
		previousIdentity, identityKnown :=
			state.selectedSourceIdentities[lexicalID]
		if identityKnown && previousIdentity != physicalPathID() {
			delete(s.programs, configFileName)
			continue
		}
		var sourceFile *ast.SourceFile
		if invalidatePossibleNewSource {
			sourceFile = s.sourceFileForOpen(state, fileName)
		} else {
			sourceFile = state.sources.SourceFileForPath(fileName)
		}
		if sourceFile != nil {
			if !identityKnown && lexicalID != physicalPathID() {
				delete(s.programs, configFileName)
				continue
			}
			s.markStateSourceContent(state, sourceFile, content)
			continue
		}
		if invalidatePossibleNewSource &&
			(state.program.CommandLine().PossiblyMatchesFileName(fileName) ||
				state.failedLookupMatches(fileName, s.server.fs)) {
			delete(s.programs, configFileName)
		}
	}
}

// sourceFileForOpen avoids a canonical source scan for every unrelated
// Program when the editor opens a new target. Change and close events use the
// complete shared lookup because they normally address an existing source.
func (s *lintProgramStore) sourceFileForOpen(
	state *lintProgramState,
	fileName string,
) *ast.SourceFile {
	fileName = tspath.NormalizePath(fileName)
	if sourceFile := state.sources.SourceFileForCandidate(fileName, ""); sourceFile != nil {
		return sourceFile
	}
	if s.server.fs != nil {
		if realPath := s.server.fs.Realpath(fileName); realPath != "" &&
			tspath.NormalizePath(realPath) != fileName {
			realPath = tspath.NormalizePath(realPath)
			return state.sources.SourceFileForCandidate(realPath, realPath)
		}
	}
	return nil
}

func (s *lintProgramStore) markStateSourceContent(
	state *lintProgramState,
	sourceFile *ast.SourceFile,
	content string,
) {
	if sourceFile == nil {
		return
	}
	if sourceFile.Text() == content {
		delete(state.dirtyFiles, sourceFile.Path())
	} else {
		state.dirtyFiles[sourceFile.Path()] = struct{}{}
	}
}

func (s *lintProgramState) recordFailedLookups(
	lookups lintProgramLookups,
	fs vfs.FS,
) {
	caseSensitive := true
	if fs != nil {
		caseSensitive = fs.UseCaseSensitiveFileNames()
	}
	currentDirectory := s.program.GetCurrentDirectory()
	for _, fileName := range lookups.failedLookups {
		if fileName != "" {
			s.failedLookups[tspath.ToPath(
				fileName,
				currentDirectory,
				caseSensitive,
			)] = struct{}{}
		}
	}
}

func (s *lintProgramState) failedLookupMatches(fileName string, fs vfs.FS) bool {
	caseSensitive := true
	if fs != nil {
		caseSensitive = fs.UseCaseSensitiveFileNames()
	}
	currentDirectory := s.program.GetCurrentDirectory()
	if s.failedLookupPathMatches(tspath.ToPath(
		fileName,
		currentDirectory,
		caseSensitive,
	)) {
		return true
	}
	if fs == nil {
		return false
	}
	realPath := fs.Realpath(fileName)
	if realPath == "" {
		return false
	}
	return s.failedLookupPathMatches(tspath.ToPath(
		realPath,
		currentDirectory,
		caseSensitive,
	))
}

func (s *lintProgramState) failedLookupPathMatches(path tspath.Path) bool {
	if _, ok := s.failedLookups[path]; ok {
		return true
	}
	for directory := path.GetDirectoryPath(); directory != ""; {
		if _, ok := s.failedLookups[directory]; ok {
			return true
		}
		parent := directory.GetDirectoryPath()
		if parent == directory {
			break
		}
		directory = parent
	}
	return false
}

func (s *lintProgramState) rememberSelectedTarget(
	target target.File,
	fs vfs.FS,
) {
	canonicalPath := target.CanonicalPath
	if canonicalPath == "" {
		canonicalPath = target.Path
	}
	caseSensitive := true
	if fs != nil {
		caseSensitive = fs.UseCaseSensitiveFileNames()
	}
	s.selectedSourceIdentities[lintProgramLexicalPathID(target.Path, fs)] =
		tspath.Path(lspLexicalPathID(canonicalPath, caseSensitive))
}

func lintProgramLexicalPathID(fileName string, fs vfs.FS) tspath.Path {
	caseSensitive := true
	if fs != nil {
		caseSensitive = fs.UseCaseSensitiveFileNames()
	}
	return tspath.Path(lspLexicalPathID(fileName, caseSensitive))
}

func sortedLintProgramPaths(values map[tspath.Path]struct{}) []tspath.Path {
	return slices.Sorted(maps.Keys(values))
}
