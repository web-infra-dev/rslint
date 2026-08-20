package lsp

import (
	"context"
	"maps"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/utils"
)

// lintProgramStore owns standalone Programs that fill gaps left by Session.
// Session-owned Programs always remain authoritative.
type lintProgramStore struct {
	server              *Server
	coverage            *lintProgramCoverage
	programs            map[string]*lintProgramState
	projectMetadata     map[string]*lintProjectMetadata
	observedRootConfigs map[string]struct{}
}

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
	overlayFS         vfs.FS
	targetPath        string
	freshOnly         bool
	overlayPrepared   bool
	usedConfig        string
	usedState         *lintProgramState
	projectMetadata   map[string]*lintProjectMetadata
	transientMetadata map[string]struct{}
}

func newLintProgramStore(server *Server) *lintProgramStore {
	return &lintProgramStore{
		server:              server,
		coverage:            newLintProgramCoverage(server),
		programs:            make(map[string]*lintProgramState),
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
) (lintProgramLoader, lintProjectMetadataLoader, func()) {
	request := &lintProgramRequest{
		store:             s,
		ctx:               ctx,
		uri:               uri,
		projectMetadata:   make(map[string]*lintProjectMetadata),
		transientMetadata: make(map[string]struct{}),
	}
	return request.load, request.loadMetadata, request.finalize
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
	var aliasesConflict bool
	r.overlayFS, aliasesConflict =
		r.store.server.currentEditorOverlayFSWithConflicts(r.uri)
	r.targetPath = uriToPath(r.uri)
	if aliasesConflict {
		r.store.Invalidate()
		r.freshOnly = true
	}
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
		if state := r.store.programs[configFileName]; state != nil && state.metadata != nil {
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
		tracker := newLintTrackingFS(r.overlayFS)
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
	configFileName string,
) (*compiler.Program, *ast.SourceFile, error) {
	r.prepareOverlay()
	configFileName = tspath.NormalizePath(configFileName)
	if r.freshOnly || !r.store.Usable() {
		return r.loadFresh(configFileName)
	}

	state := r.store.programs[configFileName]
	if state == nil {
		return r.rebuild(configFileName, r.projectMetadata[configFileName])
	}

	targetSource := state.sources.SourceFileForPath(r.targetPath)
	if content, open := r.store.server.documents[r.uri]; open {
		r.store.markStateSourceContent(state, targetSource, content)
	}
	if len(state.dirtyFiles) == 0 {
		state.tracker.Inner = r.overlayFS
		return r.result(configFileName, state)
	}

	state.tracker.Inner = r.overlayFS
	state.tracker.reset()
	program := state.program
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
			clear(state.failedLookups)
			clear(state.selectedSourceIdentities)
			break
		}
	}

	state.program = program
	state.sources = utils.NewProgramSourceLookup(program, r.store.server.fs)
	clear(state.dirtyFiles)
	added, safe := r.cover(state)
	if !safe {
		return program, state.sources.SourceFileForPath(r.targetPath), nil
	}
	if added {
		return r.rebuild(configFileName, state.metadata)
	}
	return r.result(configFileName, state)
}

func (r *lintProgramRequest) loadFresh(
	configFileName string,
) (*compiler.Program, *ast.SourceFile, error) {
	metadata, err := r.metadata(configFileName)
	if err != nil {
		return nil, nil, err
	}
	program, err := createStandaloneLintProgram(metadata, r.overlayFS)
	if err != nil {
		return nil, nil, err
	}
	return program, sourceFileForPath(program, r.targetPath, r.overlayFS), nil
}

func (r *lintProgramRequest) rebuild(
	configFileName string,
	metadata *lintProjectMetadata,
) (*compiler.Program, *ast.SourceFile, error) {
	delete(r.store.programs, configFileName)
	if metadata == nil {
		var err error
		metadata, err = r.metadata(configFileName)
		if err != nil {
			return nil, nil, err
		}
	}
	_, metadataIsTransient := r.transientMetadata[configFileName]
	if r.freshOnly || !r.store.Usable() || metadataIsTransient {
		program, err := createStandaloneLintProgram(metadata, r.overlayFS)
		if err != nil {
			return nil, nil, err
		}
		return program, sourceFileForPath(program, r.targetPath, r.overlayFS), nil
	}

	// Register every dependency discovered by the first build, then rebuild
	// once so no filesystem change can fall into the registration gap.
	for attempt := range 2 {
		tracker := newLintTrackingFS(r.overlayFS)
		program, err := createStandaloneLintProgram(metadata, tracker)
		if err != nil {
			return nil, nil, err
		}
		state := &lintProgramState{
			program:                  program,
			tracker:                  tracker,
			sources:                  utils.NewProgramSourceLookup(program, r.store.server.fs),
			dirtyFiles:               make(map[tspath.Path]struct{}),
			failedLookups:            make(map[tspath.Path]struct{}),
			selectedSourceIdentities: make(map[tspath.Path]tspath.Path),
			metadata:                 metadata,
		}
		sourceFile := state.sources.SourceFileForPath(r.targetPath)
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
		r.store.programs[configFileName] = state
		return r.result(configFileName, state)
	}
	panic("unreachable")
}

func (r *lintProgramRequest) result(
	configFileName string,
	state *lintProgramState,
) (*compiler.Program, *ast.SourceFile, error) {
	sourceFile := state.sources.SourceFileForPath(r.targetPath)
	if sourceFile != nil {
		state.rememberSelectedSource(r.targetPath, r.store.server.fs)
		r.usedConfig = configFileName
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
	if !r.store.Usable() || r.usedState == nil {
		return
	}
	added, safe := r.cover(r.usedState)
	if !safe || added {
		// Rules and the checker can perform lazy reads. New watcher coverage is
		// active now, but this Program predates it.
		delete(r.store.programs, r.usedConfig)
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
	content, ok := s.server.fs.ReadFile(uriToPath(uri))
	if !ok {
		s.Invalidate()
		return
	}
	s.markContent(uriToPath(uri), content, false)
}

// DidChangeWatchedFiles returns whether resident state was discarded. The
// caller uses that signal to refresh diagnostics even when Session does not
// own the custom project that registered the watcher.
func (s *lintProgramStore) DidChangeWatchedFiles(
	changes []*lsproto.FileEvent,
) bool {
	discarded := false
	for _, change := range changes {
		if s.isOpenSourceOverlayWatchChange(change) {
			uri := change.Uri
			discarded = s.markContent(
				uriToPath(uri),
				s.server.documents[uri],
				false,
			) || discarded
			continue
		}
		return s.Invalidate() || discarded
	}
	return discarded
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
	clear(s.projectMetadata)
	clear(s.observedRootConfigs)
	return hadState
}

func (s *lintProgramStore) markContent(
	fileName string,
	content string,
	invalidatePossibleNewSource bool,
) (discarded bool) {
	if tspath.HasJSONFileExtension(fileName) {
		return s.Invalidate()
	}
	if invalidatePossibleNewSource {
		for configFileName, metadata := range s.projectMetadata {
			if metadata != nil && metadata.commandLine != nil &&
				metadata.commandLine.PossiblyMatchesFileName(fileName) {
				delete(s.projectMetadata, configFileName)
				discarded = true
			}
		}
	}
	lexicalID := lintProgramLexicalPathID(fileName, s.server.fs)
	var physicalID tspath.Path
	physicalPathID := func() tspath.Path {
		if physicalID == "" {
			physicalID = tspath.Path(lspFilesystemPathID(fileName, s.server.fs))
		}
		return physicalID
	}
	for configFileName, state := range s.programs {
		previousIdentity, identityKnown :=
			state.selectedSourceIdentities[lexicalID]
		if identityKnown && previousIdentity != physicalPathID() {
			delete(s.programs, configFileName)
			discarded = true
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
				discarded = true
				continue
			}
			s.markStateSourceContent(state, sourceFile, content)
			continue
		}
		if invalidatePossibleNewSource &&
			(state.program.CommandLine().PossiblyMatchesFileName(fileName) ||
				state.failedLookupMatches(fileName, s.server.fs)) {
			delete(s.programs, configFileName)
			discarded = true
		}
	}
	return discarded
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

func (s *lintProgramState) rememberSelectedSource(fileName string, fs vfs.FS) {
	s.selectedSourceIdentities[lintProgramLexicalPathID(fileName, fs)] =
		tspath.Path(lspFilesystemPathID(fileName, fs))
}

func lintProgramLexicalPathID(fileName string, fs vfs.FS) tspath.Path {
	caseSensitive := true
	if fs != nil {
		caseSensitive = fs.UseCaseSensitiveFileNames()
	}
	return tspath.Path(lspLexicalPathID(fileName, caseSensitive))
}

// TypeScript config declarations are exact lexical identities. Realpath and
// filesystem-wide case folding cannot merge two authored declarations because
// their relative include/extends path spaces may differ.
func lintProjectDeclarationPathID(fileName string) tspath.Path {
	return tspath.ToPath(tspath.NormalizePath(fileName), "", true)
}

func sortedLintProgramPaths(values map[tspath.Path]struct{}) []tspath.Path {
	return slices.Sorted(maps.Keys(values))
}
