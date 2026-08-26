package lsp

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type lintPassResult struct {
	Diagnostics     []rule.RuleDiagnostic
	HasSyntaxErrors bool
}

// LintResponse represents a lint response from Go to JS
type LintResponse struct {
	Diagnostics []lsproto.Diagnostic `json:"diagnostics"`
	ErrorCount  int                  `json:"errorCount"`
	FileCount   int                  `json:"fileCount"`
	RuleCount   int                  `json:"ruleCount"`
}

func runLintWithSession(uri lsproto.DocumentUri, session *project.Session, ctx context.Context, rslintConfig config.RslintConfig, cwd string, enforcePlugins bool, tsConfigPaths []string, fs vfs.FS) ([]rule.RuleDiagnostic, error) {
	target := lspConfigTarget(uriToPath(uri), cwd, fs)
	snapshot := resolveDocumentLintSnapshotConfig(documentLintSnapshot{
		target: target,
		config: rslintConfig,
		pathSpaces: config.NewPathSpaceSnapshot(
			map[string]config.RslintConfig{target.ConfigDirectory: rslintConfig},
			fs,
		),
		ruleCatalog:           rules.All(),
		typeScriptConfigPaths: tsConfigPaths,
		usesJavaScriptConfig:  enforcePlugins,
	}, fs)
	result, err := runLintWithProgramLoader(
		uri,
		session,
		ctx,
		snapshot,
		cwd,
		fs,
		lintProjectLoaders{},
		nil,
	)
	return result.Diagnostics, err
}

// runLintWithProgramLoader resolves one document against two distinct
// directories: configCwd is the config's own path space, which a nested JS
// config moves to its own directory, while processCwd is the server's working
// directory that rules see through RuleContext.ProcessCurrentDirectory.
func runLintWithProgramLoader(
	uri lsproto.DocumentUri,
	session *project.Session,
	ctx context.Context,
	snapshot documentLintSnapshot,
	processCwd string,
	fs vfs.FS,
	loaders lintProjectLoaders,
	sessionRoots *lintSessionProjectRootCache,
) (lintPassResult, error) {
	if !snapshot.configResolved {
		snapshot = resolveDocumentLintSnapshotConfig(snapshot, fs)
	}
	target := snapshot.target
	if isDefaultExcludedLintPath(target.Path, processCwd, fs) {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}, nil
	}

	// Files excluded by the config's `ignores` patterns produce no diagnostics,
	// matching CLI behavior. Return early before spinning up the language service.
	if snapshot.resolvedConfig.GloballyIgnored {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}, nil
	}

	program, sourceFile, hasTypeInfo, err := selectLintProgram(
		uri,
		target,
		session,
		ctx,
		snapshot.typeScriptConfigPaths,
		fs,
		loaders,
		sessionRoots,
	)
	if err != nil {
		return lintPassResult{}, err
	}
	return lintSingleFile(
		program,
		sourceFile,
		target,
		processCwd,
		hasTypeInfo,
		snapshot.resolvedConfig.EnabledRules,
		rule.EditDemandAll,
		ctx,
	), nil
}

// rulesSkippedInEditors names the rules the language server never runs. A rule
// belongs here when what it checks is a property of the file's bytes rather
// than of its text: an editor's document holds text that has already been
// decoded, so such a property is neither visible in the document nor reachable
// by a text edit.
var rulesSkippedInEditors = map[string]bool{
	// unicode-bom checks for a leading byte order mark. An editor decodes the
	// mark into the document's encoding — VS Code shows it in the status bar as
	// "UTF-8 with BOM" — so the document text never carries it and no text edit
	// adds or removes it. The file on disk is the only remaining witness, and an
	// unsaved buffer may already disagree with it. `rslint --fix`, which rewrites
	// the file itself, is where the rule applies.
	"unicode-bom": true,
}

// rulesServedToEditors drops the rules the language server never runs. The
// input is a cached slice shared across files, so filtering builds a new one
// and an unaffected configuration keeps the original.
func rulesServedToEditors(rules []rule.ConfiguredRule) []rule.ConfiguredRule {
	skipped := func(r rule.ConfiguredRule) bool { return rulesSkippedInEditors[r.Name] }
	if !slices.ContainsFunc(rules, skipped) {
		return rules
	}
	served := make([]rule.ConfiguredRule, 0, len(rules))
	for _, configured := range rules {
		if !skipped(configured) {
			served = append(served, configured)
		}
	}
	return served
}

func lintSingleFile(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	target target.File,
	processCwd string,
	hasTypeInfo bool,
	enabledRules []rule.ConfiguredRule,
	editDemand rule.EditDemand,
	ctx context.Context,
) lintPassResult {
	if sourceFile == nil {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}
	}
	sourceProgram := lintprogram.NewFromCompiler(program)
	if syntacticDiagnostics := sourceProgram.SyntacticDiagnostics(ctx, sourceFile); len(syntacticDiagnostics) > 0 {
		diagnostics := make([]rule.RuleDiagnostic, 0, len(syntacticDiagnostics))
		for _, diagnostic := range syntacticDiagnostics {
			diagnostics = append(diagnostics, rule.RuleDiagnostic{
				RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
				SourceFile:   sourceFile,
				FilePath:     sourceFile.FileName(),
				Range:        diagnostic.Loc(),
				Message:      rule.RuleMessage{Description: diagnostic.String()},
				Severity:     rule.SeverityError,
				Origin:       rule.DiagnosticOriginTypeScript,
				PreFormatted: true,
			})
		}
		return lintPassResult{Diagnostics: diagnostics, HasSyntaxErrors: true}
	}

	// Collect diagnostics
	var diagnostics []rule.RuleDiagnostic
	var diagnosticsLock sync.Mutex

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()
		diagnostics = append(diagnostics, d)
	}

	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     sourceProgram,
		File:        sourceFile.FileName(),
		Cwd:         processCwd,
		HasTypeInfo: hasTypeInfo,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return rulesServedToEditors(enabledRules)
		},
		Consumer: rule.DiagnosticConsumer{
			Demand: editDemand,
			Report: diagnosticCollector,
		},
	})

	if diagnostics == nil {
		diagnostics = []rule.RuleDiagnostic{}
	}
	return lintPassResult{Diagnostics: diagnostics}
}

func selectLintProgram(
	uri lsproto.DocumentUri,
	target target.File,
	session *project.Session,
	ctx context.Context,
	tsConfigPaths []string,
	fs vfs.FS,
	fallbackLoaders lintProjectLoaders,
	sessionRoots *lintSessionProjectRootCache,
) (*compiler.Program, *ast.SourceFile, bool, error) {
	// Flush pending document changes and collect every already-loaded project
	// containing the file. The default language service remains the
	// non-project-backed fallback when none of the config's declared projects
	// contains the file.
	_, languageService, loadedProjects, err := session.GetLanguageServiceAndProjectsForFile(ctx, uri)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to get language service: %w", err)
	}
	program := languageService.GetProgram()

	// Type information follows parserOptions.project declaration order, not the
	// TypeScript session's default-project heuristic. Direct config roots outrank
	// files present only through imports. Custom config names that the main
	// project service has not loaded are supplied by rslint-owned
	// session-external ts-go Programs.
	type loadedLintProject struct {
		program     *compiler.Program
		commandLine *tsoptions.ParsedCommandLine
	}
	loadedByConfig := make(map[tspath.Path]loadedLintProject, len(loadedProjects))
	for _, candidate := range loadedProjects {
		if candidate == nil || candidate.GetProgram() == nil {
			continue
		}
		candidateProgram := candidate.GetProgram()
		configPath := string(candidate.Id())
		if configPath == "" {
			continue
		}
		commandLine := candidateProgram.CommandLine()
		if sessionProject, ok := candidate.(*project.Project); ok && sessionProject.CommandLine != nil {
			// The Program command line may include automatic type-acquisition
			// roots. Project.CommandLine is the authored tsconfig root set.
			commandLine = sessionProject.CommandLine
		}
		loadedByConfig[lintProgramLexicalPathID(configPath, fs)] = loadedLintProject{
			program:     candidateProgram,
			commandLine: commandLine,
		}
	}
	loaders := lintProjectLoaders{
		metadata: func(tsConfigPath string) (*lintProjectMetadata, bool, error) {
			if loadedProject, ok := loadedByConfig[lintProgramLexicalPathID(tsConfigPath, fs)]; ok {
				metadata := sessionRoots.metadata(tsConfigPath, loadedProject.commandLine, fs)
				return metadata, metadata != nil, nil
			}
			if fallbackLoaders.metadata == nil {
				return nil, false, nil
			}
			return fallbackLoaders.metadata(tsConfigPath)
		},
		program: func(tsConfigPath string) (*compiler.Program, *ast.SourceFile, error) {
			if loadedProject, ok := loadedByConfig[lintProgramLexicalPathID(tsConfigPath, fs)]; ok {
				return loadedProject.program, sourceFileForTarget(loadedProject.program, target, fs), nil
			}
			if fallbackLoaders.program == nil {
				return nil, nil, nil
			}
			return fallbackLoaders.program(tsConfigPath)
		},
	}
	selected, found, err := selectConfiguredLintProject(tsConfigPaths, target, loaders)
	if err != nil {
		return nil, nil, false, err
	}
	if found {
		return selected.program, selected.sourceFile, true, nil
	}
	return program, sourceFileForTarget(program, target, fs), false, nil
}

func sourceFileForPath(program *compiler.Program, filename string, fs vfs.FS) *ast.SourceFile {
	return utils.NewProgramSourceLookup(program, fs).SourceFileForPath(filename)
}

func sourceFileForTarget(
	program *compiler.Program,
	target target.File,
	fs vfs.FS,
) *ast.SourceFile {
	return utils.NewProgramSourceLookup(program, fs).
		SourceFileForTarget(target.Path, target.CanonicalPath)
}

func (s *Server) currentEditorOverlayFSForTarget(
	uri lsproto.DocumentUri,
	target target.File,
) vfs.FS {
	content, open := s.documents[uri]
	files, _ := s.currentEditorOverlayFilesForFrozenTarget(
		uri,
		target,
		content,
		open,
	)
	return newFrozenLintTargetOverlayFS(s.fs, files, target)
}

func (s *Server) currentEditorOverlayFSForTargetWithConflicts(
	uri lsproto.DocumentUri,
	target target.File,
) (vfs.FS, bool) {
	content, open := s.documents[uri]
	files, aliasesConflict := s.currentEditorOverlayFilesForFrozenTarget(
		uri,
		target,
		content,
		open,
	)
	return newFrozenLintTargetOverlayFS(s.fs, files, target), aliasesConflict
}

func (s *Server) currentEditorOverlayFilesForTarget(
	uri lsproto.DocumentUri,
	target target.File,
	content string,
) map[string]string {
	files, _ := s.currentEditorOverlayFilesForFrozenTarget(
		uri,
		target,
		content,
		true,
	)
	return files
}

// currentEditorOverlayFilesForFrozenTarget resolves every other open document
// normally, but inserts the selected document only under the identity frozen
// by documentLintSnapshot. Re-resolving the selected URI here could mix two
// symlink generations between config/project selection and Program creation.
func (s *Server) currentEditorOverlayFilesForFrozenTarget(
	uri lsproto.DocumentUri,
	target target.File,
	targetContent string,
	includeTarget bool,
) (map[string]string, bool) {
	files := make(map[string]string, len(s.documents)*2)
	contentByPhysicalPath := make(map[string]string, len(s.documents))
	aliasesConflict := false
	add := func(filePath string, content string) {
		physicalPath := s.addEditorOverlayFile(files, filePath, content)
		if previous, exists := contentByPhysicalPath[physicalPath]; exists &&
			previous != content {
			aliasesConflict = true
		}
		contentByPhysicalPath[physicalPath] = content
	}
	for documentURI, documentContent := range s.documents {
		if documentURI == uri {
			continue
		}
		add(uriToPath(documentURI), documentContent)
	}
	if includeTarget {
		physicalPath := frozenLintTargetPhysicalPathID(target, s.fs)
		if previous, exists := contentByPhysicalPath[physicalPath]; exists &&
			previous != targetContent {
			aliasesConflict = true
		}
		contentByPhysicalPath[physicalPath] = targetContent
		addEditorOverlayTarget(files, target, targetContent)
	}
	return files, aliasesConflict
}

func frozenLintTargetPhysicalPathID(
	target target.File,
	fs vfs.FS,
) string {
	filePath := target.CanonicalPath
	if filePath == "" {
		filePath = target.Path
	}
	caseSensitive := true
	if fs != nil {
		caseSensitive = fs.UseCaseSensitiveFileNames()
	}
	return lspLexicalPathID(filePath, caseSensitive)
}

type frozenLintTargetOverlayFS struct {
	vfs.FS
	lexicalPathID   string
	canonicalPathID string
	canonicalPath   string
}

func newFrozenLintTargetOverlayFS(
	baseFS vfs.FS,
	files map[string]string,
	target target.File,
) vfs.FS {
	caseSensitive := true
	if baseFS != nil {
		caseSensitive = baseFS.UseCaseSensitiveFileNames()
	}
	canonicalPath := target.CanonicalPath
	if canonicalPath == "" {
		canonicalPath = target.Path
	}
	canonicalPath = tspath.NormalizePath(canonicalPath)
	return &frozenLintTargetOverlayFS{
		FS:              utils.NewOverlayVFS(baseFS, files),
		lexicalPathID:   lspLexicalPathID(target.Path, caseSensitive),
		canonicalPathID: lspLexicalPathID(canonicalPath, caseSensitive),
		canonicalPath:   canonicalPath,
	}
}

func (fs *frozenLintTargetOverlayFS) Realpath(filePath string) string {
	caseSensitive := fs.UseCaseSensitiveFileNames()
	pathID := lspLexicalPathID(filePath, caseSensitive)
	if pathID == fs.lexicalPathID || pathID == fs.canonicalPathID {
		return fs.canonicalPath
	}
	return fs.FS.Realpath(filePath)
}

func addEditorOverlayTarget(
	files map[string]string,
	target target.File,
	content string,
) {
	if target.Path != "" {
		files[tspath.NormalizePath(target.Path)] = content
	}
	if target.CanonicalPath != "" {
		files[tspath.NormalizePath(target.CanonicalPath)] = content
	}
}

func (s *Server) addEditorOverlayFile(
	files map[string]string,
	filePath string,
	content string,
) string {
	filePath = tspath.NormalizePath(filePath)
	files[filePath] = content
	caseSensitive := true
	if s.fs != nil {
		caseSensitive = s.fs.UseCaseSensitiveFileNames()
	}
	if s.fs == nil {
		return string(tspath.ToPath(filePath, "", caseSensitive))
	}
	if realPath := s.fs.Realpath(filePath); realPath != "" {
		realPath = tspath.NormalizePath(realPath)
		files[realPath] = content
		return string(tspath.ToPath(realPath, "", caseSensitive))
	}
	return string(tspath.ToPath(filePath, "", caseSensitive))
}

func createStandaloneFallbackProgram(filename string, cwd string, fs vfs.FS) (*compiler.Program, error) {
	host := utils.CreateCompilerHost(cwd, fs)
	return utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, []string{filename}, host)
}

func (s *Server) runConfiguredLint(
	uri lsproto.DocumentUri,
	ctx context.Context,
	snapshot documentLintSnapshot,
) (lintPassResult, error) {
	request := newStandaloneLintProjectRequest(
		snapshot.target,
		func() vfs.FS { return s.currentEditorOverlayFSForTarget(uri, snapshot.target) },
	)
	loaders := request.loaders()
	finalize := func() {}
	if s.lintPrograms != nil && s.lintPrograms.Usable() {
		loadProgram, loadMetadata, requestFinalize := s.lintPrograms.Request(
			ctx,
			uri,
			snapshot.target,
		)
		loaders = lintProjectLoaders{
			program:  loadProgram,
			metadata: loadMetadata,
		}
		finalize = requestFinalize
	}
	defer finalize()
	return runLintWithProgramLoader(
		uri,
		s.session,
		ctx,
		snapshot,
		s.cwd,
		s.fs,
		loaders,
		s.lintSessionRoots,
	)
}

// runConfiguredLintForContent lints a speculative fix pass against an
// isolated overlay. It never mutates the TypeScript Session's open document,
// so cancelling or declining a code action cannot change later LSP results.
func (s *Server) runConfiguredLintForContent(
	uri lsproto.DocumentUri,
	ctx context.Context,
	content string,
	rslintConfig config.RslintConfig,
	cwd string,
	enforcePlugins bool,
	tsConfigPaths []string,
) (lintPassResult, error) {
	ruleCatalog := rules.All()
	if enforcePlugins {
		ruleCatalog = s.currentRuleCatalog()
	}
	target := lspConfigTarget(uriToPath(uri), cwd, s.fs)
	snapshot := resolveDocumentLintSnapshotConfig(documentLintSnapshot{
		target: target,
		config: rslintConfig,
		pathSpaces: config.NewPathSpaceSnapshot(
			map[string]config.RslintConfig{target.ConfigDirectory: rslintConfig},
			s.fs,
		),
		ruleCatalog:           ruleCatalog,
		typeScriptConfigPaths: tsConfigPaths,
		usesJavaScriptConfig:  enforcePlugins,
	}, s.fs)
	return s.runConfiguredLintForContentWithSnapshot(uri, ctx, content, snapshot)
}

func (s *Server) runConfiguredLintForContentWithSnapshot(
	uri lsproto.DocumentUri,
	ctx context.Context,
	content string,
	snapshot documentLintSnapshot,
) (lintPassResult, error) {
	if !snapshot.configResolved {
		snapshot = resolveDocumentLintSnapshotConfig(snapshot, s.fs)
	}
	target := snapshot.target
	cwd := target.ConfigDirectory
	if isDefaultExcludedLintPath(target.Path, s.cwd, s.fs) {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}, nil
	}
	if snapshot.resolvedConfig.GloballyIgnored {
		return lintPassResult{Diagnostics: []rule.RuleDiagnostic{}}, nil
	}

	// Apply the speculative target last so it wins over every open URI alias
	// that resolves to the same physical file, without resolving the frozen
	// target through the filesystem again.
	files := s.currentEditorOverlayFilesForTarget(uri, target, content)
	overlayFS := newFrozenLintTargetOverlayFS(s.fs, files, target)

	request := newStandaloneLintProjectRequestWithFS(target, overlayFS)
	selected, found, err := selectConfiguredLintProject(
		snapshot.typeScriptConfigPaths,
		target,
		request.loaders(),
	)
	if err != nil {
		return lintPassResult{}, err
	}
	if found {
		return lintSingleFile(
			selected.program,
			selected.sourceFile,
			target,
			s.cwd,
			true,
			snapshot.resolvedConfig.EnabledRules,
			rule.EditDemandAutofix,
			ctx,
		), nil
	}

	program, err := createStandaloneFallbackProgram(target.Path, cwd, overlayFS)
	if err != nil {
		return lintPassResult{}, fmt.Errorf("create fallback lint program: %w", err)
	}
	return lintSingleFile(
		program,
		sourceFileForTarget(program, target, overlayFS),
		target,
		s.cwd,
		false,
		snapshot.resolvedConfig.EnabledRules,
		rule.EditDemandAutofix,
		ctx,
	), nil
}
