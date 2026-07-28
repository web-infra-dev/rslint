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

// lintProgramStore owns only the standalone Programs that fill gaps left by
// Session. Session-owned projects always remain authoritative.
type lintProgramStore struct {
	server   *Server
	coverage *lintProgramCoverage
	programs map[string]*lintProgramState
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
}

type lintProgramRequest struct {
	store           *lintProgramStore
	ctx             context.Context
	uri             lsproto.DocumentUri
	overlayFS       vfs.FS
	targetPath      string
	freshOnly       bool
	overlayPrepared bool
	usedConfig      string
	usedState       *lintProgramState
}

func newLintProgramStore(server *Server) *lintProgramStore {
	return &lintProgramStore{
		server:   server,
		coverage: newLintProgramCoverage(server),
		programs: make(map[string]*lintProgramState),
	}
}

func (s *lintProgramStore) Usable() bool {
	return s != nil && !s.coverage.disabled
}

func (s *lintProgramStore) Request(
	ctx context.Context,
	uri lsproto.DocumentUri,
) (lintProgramLoader, func()) {
	request := &lintProgramRequest{
		store: s,
		ctx:   ctx,
		uri:   uri,
	}
	load := func(configFileName string) (*compiler.Program, *ast.SourceFile, error) {
		if !request.overlayPrepared {
			request.overlayPrepared = true
			var aliasesConflict bool
			request.overlayFS, aliasesConflict =
				s.server.currentEditorOverlayFSWithConflicts(uri)
			request.targetPath = uriToPath(uri)
			if aliasesConflict {
				s.Invalidate()
				request.freshOnly = true
			}
		}
		return request.load(configFileName)
	}
	return load, request.finalize
}

func (r *lintProgramRequest) load(
	configFileName string,
) (*compiler.Program, *ast.SourceFile, error) {
	if r.freshOnly || !r.store.Usable() {
		return r.loadFresh(configFileName)
	}

	state := r.store.programs[configFileName]
	if state == nil {
		return r.rebuild(configFileName)
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
			return r.rebuild(configFileName)
		}
		updated, newSourceFile, graphReused :=
			program.UpdateProgram(dirtyPath, program.Host(), nil)
		if updated == nil {
			return r.rebuild(configFileName)
		}
		updated.BindSourceFiles()
		if graphReused &&
			!lintProgramGraphReusable(program, updated, oldSourceFile, newSourceFile) {
			return r.rebuild(configFileName)
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
		return r.rebuild(configFileName)
	}
	return r.result(configFileName, state)
}

func (r *lintProgramRequest) loadFresh(
	configFileName string,
) (*compiler.Program, *ast.SourceFile, error) {
	program, err := createStandaloneLintProgram(configFileName, r.overlayFS)
	if err != nil {
		return nil, nil, err
	}
	return program, sourceFileForPath(program, r.targetPath, r.overlayFS), nil
}

func (r *lintProgramRequest) rebuild(
	configFileName string,
) (*compiler.Program, *ast.SourceFile, error) {
	delete(r.store.programs, configFileName)

	// Register every dependency discovered by the first build, then rebuild
	// once so no filesystem change can fall into the registration gap.
	for attempt := range 2 {
		tracker := newLintTrackingFS(r.overlayFS)
		program, err := createStandaloneLintProgram(configFileName, tracker)
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
		}
		added, safe := r.cover(state)
		if !safe {
			return program, state.sources.SourceFileForPath(r.targetPath), nil
		}
		if added && attempt == 0 {
			continue
		}
		if added {
			// Lazy compiler reads expanded coverage again. Serve this Program
			// once, but do not retain it across the uncovered interval.
			return program, state.sources.SourceFileForPath(r.targetPath), nil
		}
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
	hadPrograms := len(s.programs) != 0
	clear(s.programs)
	return hadPrograms
}

func (s *lintProgramStore) markContent(
	fileName string,
	content string,
	invalidatePossibleNewSource bool,
) (discarded bool) {
	if tspath.HasJSONFileExtension(fileName) {
		return s.Invalidate()
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

func sortedLintProgramPaths(values map[tspath.Path]struct{}) []tspath.Path {
	return slices.Sorted(maps.Keys(values))
}
