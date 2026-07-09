package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/inspector"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IPCHandler implements the ipc.Handler interface
type IPCHandler struct{}

// programCache holds a cached Program instance for AST info requests
type programCache struct {
	mu              sync.RWMutex
	fileContent     string
	compilerOptions string // JSON serialized for comparison
	program         *compiler.Program
	sourceFile      *ast.SourceFile
}

// Global program cache for AST info requests
var astInfoProgramCache = &programCache{}

// HandleLint handles lint requests in IPC mode
func (h *IPCHandler) HandleLint(req api.LintRequest) (*api.LintResponse, error) {

	// Resolve the working directory WITHOUT os.Chdir: this is a long-lived,
	// reused --api process, so mutating the process-global cwd would leak
	// across requests (and race a future concurrent mode). Everything
	// downstream (resolveRequestPath / config loader / CreateCompilerHost /
	// CreateProgram) takes this directory explicitly, so a local var suffices.
	currentDirectory := req.WorkingDirectory
	if currentDirectory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error getting current directory: %w", err)
		}
		currentDirectory = cwd
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)

	// Create filesystem
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	var allowedFiles []string
	seenAllowedFiles := make(map[string]struct{})

	resolveRequestPath := func(filePath string) string {
		if tspath.PathIsAbsolute(filePath) {
			return tspath.NormalizePath(filePath)
		}
		return tspath.ResolvePath(currentDirectory, filePath)
	}

	addAllowedFile := func(filePath string) string {
		normalizedPath := resolveRequestPath(filePath)
		if _, exists := seenAllowedFiles[normalizedPath]; exists {
			return normalizedPath
		}
		seenAllowedFiles[normalizedPath] = struct{}{}
		allowedFiles = append(allowedFiles, normalizedPath)
		return normalizedPath
	}

	if req.Files != nil {
		allowedFiles = make([]string, 0, len(req.Files)+len(req.FileContents))
		for _, filePath := range req.Files {
			addAllowedFile(filePath)
		}
	}
	// Apply file contents if provided
	var fileContents map[string]string
	if len(req.FileContents) > 0 {
		if allowedFiles == nil {
			allowedFiles = make([]string, 0, len(req.FileContents))
		}
		fileContents = make(map[string]string, len(req.FileContents))
		for k, v := range req.FileContents {
			normalizedPath := addAllowedFile(k)
			fileContents[normalizedPath] = v
		}
	}

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllRules()

	// Config is the JS-resolved final config object (serialized RslintConfig).
	// --api never reads config from disk or auto-discovers — the JS side does
	// all of overrideConfig / config-file / discovery / normalize. Empty/absent
	// means "no config" (zero rules).
	var rslintConfig rslintconfig.RslintConfig
	if len(req.Config) > 0 {
		if err := json.Unmarshal(req.Config, &rslintConfig); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
		if err := rslintconfig.ValidateConfig(rslintConfig); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
	}
	configDirectory := req.ConfigDirectory
	if configDirectory == "" {
		configDirectory = currentDirectory
	}
	configDirectory = tspath.NormalizePath(configDirectory)
	if len(fileContents) > 0 {
		addEquivalentFileContentPaths(fileContents, configDirectory, currentDirectory, fs)
		fs = utils.NewOverlayVFS(fs, fileContents)
	}
	var gitignoreGlobs []string
	if len(allowedFiles) > 0 {
		gitignoreGlobs = rslintconfig.ReadGitignoreAsGlobsForFiles(configDirectory, fs, allowedFiles)
	} else {
		gitignoreGlobs = rslintconfig.ReadGitignoreAsGlobs(configDirectory, fs, rslintconfig.ExtractConfigIgnores(rslintConfig))
	}
	if len(gitignoreGlobs) > 0 {
		rslintConfig = append(rslintconfig.RslintConfig{{Ignores: gitignoreGlobs}}, rslintConfig...)
	}
	tsConfigs, err := rslintconfig.ResolveTsConfigPaths(rslintConfig, configDirectory, fs)
	if err != nil {
		return nil, fmt.Errorf("error resolving tsconfig: %w", err)
	}

	// Create compiler host with a request-scoped parse cache: the tsconfig
	// loop below builds one Program per tsconfig on this host, and shared
	// dependencies (lib/node_modules d.ts, cross-package sources) parse once.
	// The cache dies with this request — no RetainOnly sweep here, and none
	// may be added inside the tsConfigs loop (a mid-build eviction boundary
	// buys nothing per I7 and complicates reasoning).
	parseCache := utils.NewParseCache()
	host := utils.WithParseCache(utils.CreateCompilerHost(configDirectory, fs), parseCache)
	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}

	// Create programs from all tsconfig files found in rslint config
	programs := []*compiler.Program{}
	for _, configFileName := range tsConfigs {
		program, err := utils.CreateProgramLenient(false, fs, configDirectory, configFileName, host)
		if err != nil {
			return nil, fmt.Errorf("error creating TS program for %s: %w", configFileName, err)
		}
		programs = append(programs, program)
	}

	// Resolve the exact lint target set, bind targets to existing Programs, and
	// append an AST-only fallback Program for selected files absent from every
	// Program (the typical lintText/lintFiles in-memory file). Identical to the
	// CLI via the shared helper. configMap is nil: the --api path is always
	// single-config (the JS side resolves any multi-config merge into one entry
	// list).
	// The --api path never runs the type-check phase (RunLinterOptions.TypeCheck
	// stays false), so there is no per-program type-check skip mask to build.
	programs, typeInfoFiles, _, _, targetsByProgram, configPathBySourcePath := buildProgramsWithLintTargets(
		programs, nil, rslintConfig, configDirectory, nil, nil, fs, allowedFiles, nil, parseCache, false,
	)
	fileConfigResolver := newLintConfigResolver(nil, rslintConfig, configDirectory, true, typeInfoFiles, configPathBySourcePath)

	// Collect diagnostics and source files
	var diagnostics []api.Diagnostic
	var diagnosticsLock sync.Mutex
	errorsCount := 0
	warningsCount := 0
	fixableErrorsCount := 0
	fixableWarningsCount := 0
	// When Fix is requested, the original RuleDiagnostics (byte-offset fixes +
	// their SourceFile) are retained per file for the in-band fix pass below.
	var diagnosticsByFile map[string][]rule.RuleDiagnostic
	if req.Fix {
		diagnosticsByFile = make(map[string][]rule.RuleDiagnostic)
	}

	// Track source files for encoding
	sourceFiles := make(map[string]*ast.SourceFile)
	var sourceFilesLock sync.Mutex

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()
		if d.SourceFile != nil {
			sourceFilesLock.Lock()
			filePath := tspath.ConvertToRelativePath(d.FilePath, comparePathOptions)
			if sf, ok := d.SourceFile.(*ast.SourceFile); ok {
				sourceFiles[filePath] = sf
			}
			sourceFilesLock.Unlock()
		}

		diagnosticStart := d.Range.Pos()
		diagnosticEnd := d.Range.End()

		startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticStart)
		endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticEnd)

		diagnostic := api.Diagnostic{
			RuleName:  d.RuleName,
			MessageId: d.Message.Id,
			Message:   d.Message.Description,
			FilePath:  tspath.ConvertToRelativePath(d.FilePath, comparePathOptions),
			Range: api.Range{
				Start: api.Position{
					Line:   startLine + 1, // Convert to 1-based indexing
					Column: int(startColumn) + 1,
				},
				End: api.Position{
					Line:   endLine + 1,
					Column: int(endColumn) + 1,
				},
			},
			Severity: d.Severity.String(),
		}

		// Fix and suggestion ranges are flat UTF-16 offsets (ESLint's unit),
		// converted from the rule's byte offsets via byteOffsetToUTF16. This is
		// a DIFFERENT conversion than the line/column above (which counts UTF-16
		// units from the line start, not a flat file offset).
		fixText := d.SourceFile.Text()

		// Add fixes if available.
		if d.FixesPtr != nil && len(*d.FixesPtr) > 0 {
			var fixes []api.Fix
			for _, fix := range *d.FixesPtr {
				fixes = append(fixes, api.Fix{
					Text:     fix.Text,
					StartPos: byteOffsetToUTF16(fixText, fix.Range.Pos()),
					EndPos:   byteOffsetToUTF16(fixText, fix.Range.End()),
				})
			}
			diagnostic.Fixes = fixes
		}

		// Add suggestions if available — optional, user-selected fixes the
		// editor surfaces (distinct from auto-applied Fixes).
		if d.Suggestions != nil && len(*d.Suggestions) > 0 {
			suggestions := make([]api.Suggestion, 0, len(*d.Suggestions))
			for _, sug := range *d.Suggestions {
				var fixes []api.Fix
				for _, fix := range sug.FixesArr {
					fixes = append(fixes, api.Fix{
						Text:     fix.Text,
						StartPos: byteOffsetToUTF16(fixText, fix.Range.Pos()),
						EndPos:   byteOffsetToUTF16(fixText, fix.Range.End()),
					})
				}
				suggestions = append(suggestions, api.Suggestion{
					MessageId: sug.Message.Id,
					Message:   sug.Message.Description,
					Data:      sug.Message.Data,
					Fixes:     fixes,
				})
			}
			diagnostic.Suggestions = suggestions
		}

		diagnostics = append(diagnostics, diagnostic)

		// Split counts by severity (ESLint semantics): errorCount counts errors
		// only, not the total. fixable*Count counts the fixable subset.
		hasFix := d.FixesPtr != nil && len(*d.FixesPtr) > 0
		switch d.Severity {
		case rule.SeverityError:
			errorsCount++
			if hasFix {
				fixableErrorsCount++
			}
		case rule.SeverityWarning:
			warningsCount++
			if hasFix {
				fixableWarningsCount++
			}
		}

		// Retain the original diagnostic (byte-offset fixes + SourceFile) for the
		// in-band fix pass, grouped by absolute file path.
		if req.Fix && hasFix {
			diagnosticsByFile[d.FilePath] = append(diagnosticsByFile[d.FilePath], d)
		}
	}

	// Target discovery already excluded default paths, global ignores, and
	// .gitignore entries. Entry-level ignores are not target exclusions: an
	// explicit file excluded by the only matching entry is still a lint result,
	// but runs zero rules through GetActiveRulesForFile.
	shouldReportLintSyntax := func(filePath string) bool {
		return fileConfigResolver.ConfigForFile(filePath) != nil
	}
	for _, diagnostic := range collectTargetSyntacticDiagnostics(programs, targetsByProgram, nil, false, false, shouldReportLintSyntax) {
		diagnosticCollector(diagnostic)
	}

	// Run linter
	lintResult, err := linter.RunLinter(linter.RunLinterOptions{
		Programs:       programs,
		SingleThreaded: false, // Don't use single-threaded mode for IPC
		Scope:          linter.FileScope{Files: allowedFiles},
		TargetFiles:    targetsByProgram,
		// Defense-in-depth alongside the GetRulesForFile gate: RunLinter passes
		// a nil TypeChecker to rules running on files outside this set (gap /
		// fallback files), so a non-type-aware rule with optional TypeChecker
		// usage degrades gracefully. nil ⇒ no fallback ⇒ no gap distinction.
		TypeInfoFiles: typeInfoFiles,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			// Track source file for encoding
			sourceFilesLock.Lock()
			filePath := tspath.ConvertToRelativePath(sourceFile.FileName(), comparePathOptions)
			sourceFiles[filePath] = sourceFile
			sourceFilesLock.Unlock()

			// GetActiveRulesForFile applies the type-aware gate: when
			// typeInfoFiles is non-nil and this file is not in it (a gap /
			// fallback file with no type information), type-aware rules are
			// filtered out — the only guard against a type-aware rule
			// dereferencing a nil TypeChecker and crashing the process.
			// typeInfoFiles==nil ⇒ no fallback ⇒ every linted file has type
			// info ⇒ nothing to filter. Rules come solely from the resolved
			// config object (config.rules); --api has no separate rule-options
			// surface.
			//
			// enforcePlugins=true: the --api config is a resolved JS-style flat
			// config (plugins + rules), exactly like the CLI's JS/TS config path,
			// so a rule carrying a plugin prefix runs only when its plugin is
			// declared in the config's `plugins` — matching CLI and ESLint
			// semantics (a rule whose plugin is not declared is skipped).
			return fileConfigResolver.ActiveRulesForFile(sourceFile.FileName())
		},
		OnDiagnostic: diagnosticCollector,
	})
	if err != nil {
		return nil, fmt.Errorf("error running linter: %w", err)
	}

	if diagnostics == nil {
		diagnostics = []api.Diagnostic{}
	}
	// Sort diagnostics by (file, start position) only — deliberately NO
	// end/rule tie-break: ESLint and the upstream rule tests order
	// same-start diagnostics by emission order (parent reported before
	// nested child), and a file's diagnostics are all emitted by a single
	// worker, so under a STABLE sort this key is already fully
	// deterministic. Keep this comparator in sync with the CLI one in
	// cmd.go (same policy over rule.RuleDiagnostic).
	// Known pre-existing exception: a file rooted by two tsconfigs at once
	// is linted by both programs (duplicate diagnostics with
	// scheduling-dependent interleaving) — neither introduced nor fixed
	// here.
	sort.SliceStable(diagnostics, func(i, j int) bool {
		a, b := diagnostics[i], diagnostics[j]
		if a.FilePath != b.FilePath {
			return a.FilePath < b.FilePath
		}
		if a.Range.Start.Line != b.Range.Start.Line {
			return a.Range.Start.Line < b.Range.Start.Line
		}
		return a.Range.Start.Column < b.Range.Start.Column
	})

	// Apply fixes in-band when requested. ApplyRuleFixes is the same pure fixer
	// the CLI uses (cmd.go applyFixPass), but here the result stays in-memory in
	// Output — the JS side persists it via Rslint.outputFixes. Single pass over
	// each file's fixes (non-overlapping applied, overlapping left for a later
	// lint); no cross-pass re-lint cascade (P1, see design §8).
	var output map[string]string
	if req.Fix && len(diagnosticsByFile) > 0 {
		output = make(map[string]string)
		for filePath, fileDiags := range diagnosticsByFile {
			originalContent := fileDiags[0].SourceFile.Text()
			fixedContent, _, didFix := linter.ApplyRuleFixes(originalContent, fileDiags)
			if didFix {
				output[tspath.ConvertToRelativePath(filePath, comparePathOptions)] = fixedContent
			}
		}
		if len(output) == 0 {
			output = nil
		}
	}

	// The files actually linted (target discovery already excluded global
	// ignores and gitignore entries). sourceFiles was populated by
	// GetRulesForFile for every linted file under its program-canonical
	// relative path — the same path space as Diagnostic.FilePath — so the JS
	// side can seed one result per entry. Sorted for a deterministic response.
	lintedFiles := make([]string, 0, len(sourceFiles))
	for filePath := range sourceFiles {
		lintedFiles = append(lintedFiles, filePath)
	}
	sort.Strings(lintedFiles)

	// Create response
	response := &api.LintResponse{
		Diagnostics:         diagnostics,
		ErrorCount:          errorsCount,
		WarningCount:        warningsCount,
		FixableErrorCount:   fixableErrorsCount,
		FixableWarningCount: fixableWarningsCount,
		// FileCount mirrors len(LintedFiles) (unique linted files), not
		// RunLinter's per-program visit count which double-counts a file rooted
		// by two tsconfig programs — keeping it consistent with LintedFiles and
		// the ESLint per-file result shape.
		FileCount:   len(lintedFiles),
		RuleCount:   len(lintResult.ExecutedRules),
		LintedFiles: lintedFiles,
		Output:      output,
	}
	// Only include encoded source files if requested
	if req.IncludeEncodedSourceFiles {
		encodedSourceFiles := make(map[string]api.ByteArray)
		for filePath, sourceFile := range sourceFiles {
			encoded, err := api.EncodeAST(sourceFile, filePath)

			if err != nil {
				// Log error but don't fail the entire request
				fmt.Fprintf(os.Stderr, "warning: failed to encode source file %s: %v\n", filePath, err)
				continue
			}
			encodedSourceFiles[filePath] = encoded
		}
		response.EncodedSourceFiles = encodedSourceFiles
	}
	return response, nil
}

func addEquivalentFileContentPaths(fileContents map[string]string, configDirectory string, currentDirectory string, fs vfs.FS) {
	if len(fileContents) == 0 || fs == nil {
		return
	}

	type fileContentEntry struct {
		path    string
		content string
	}
	entries := make([]fileContentEntry, 0, len(fileContents))
	for filePath, content := range fileContents {
		entries = append(entries, fileContentEntry{path: filePath, content: content})
	}

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          currentDirectory,
		UseCaseSensitiveFileNames: fs.UseCaseSensitiveFileNames(),
	}
	addAlias := func(alias string, content string) {
		if alias == "" {
			return
		}
		if _, exists := fileContents[alias]; exists {
			return
		}
		fileContents[alias] = content
	}
	addDirectoryAlias := func(fromDir string, toDir string, filePath string, content string) {
		if fromDir == "" || toDir == "" || !tspath.ContainsPath(fromDir, filePath, comparePathOptions) {
			return
		}
		relativePath := tspath.GetRelativePathFromDirectory(fromDir, filePath, comparePathOptions)
		if relativePath == "" {
			return
		}
		addAlias(tspath.ResolvePath(toDir, relativePath), content)
	}

	realConfigDirectory := fs.Realpath(configDirectory)
	for _, entry := range entries {
		if realPath := fs.Realpath(entry.path); realPath != "" && realPath != entry.path {
			addAlias(realPath, entry.content)
		}
		if realConfigDirectory != "" && tspath.ComparePaths(configDirectory, realConfigDirectory, comparePathOptions) != 0 {
			addDirectoryAlias(configDirectory, realConfigDirectory, entry.path, entry.content)
			addDirectoryAlias(realConfigDirectory, configDirectory, entry.path, entry.content)
		}
	}
}

// HandleGetAstInfo handles get AST info requests in IPC mode
func (h *IPCHandler) HandleGetAstInfo(req api.GetAstInfoRequest) (*api.GetAstInfoResponse, error) {
	// Fixed user file name for program creation
	const userFileName = "/index.ts"

	// Serialize compiler options for comparison
	compilerOptionsJSON := "{}"
	if req.CompilerOptions != nil {
		jsonBytes, err := json.Marshal(req.CompilerOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal compiler options: %w", err)
		}
		compilerOptionsJSON = string(jsonBytes)
	}

	// Check if we can use cached program
	program, userSourceFile := getCachedProgram(req.FileContent, compilerOptionsJSON)
	if program == nil || userSourceFile == nil {
		// Cache miss - create new program
		var err error
		program, userSourceFile, err = createAndCacheProgram(userFileName, req.FileContent, compilerOptionsJSON, req.CompilerOptions)
		if err != nil {
			return nil, err
		}
	}

	// Get type checker
	typeChecker, done := program.GetTypeChecker(context.Background())
	defer done()

	// Determine which source file to query
	// If FileName is set to an external file, query that file (e.g., lib.d.ts)
	// Otherwise, query the user's source file
	var targetSourceFile *ast.SourceFile
	if req.FileName != "" && req.FileName != userFileName {
		targetSourceFile = program.GetSourceFile(req.FileName)
		if targetSourceFile == nil {
			return &api.GetAstInfoResponse{}, nil
		}
	} else {
		targetSourceFile = userSourceFile
	}

	isExternalFile := targetSourceFile != userSourceFile

	// Build the response
	// Use userSourceFile as the "current" file for the builder
	// This determines which files are considered "external" (fileName will be set for nodes not in userSourceFile)
	builder := api.NewAstInfoBuilder(typeChecker, userSourceFile)
	response := &api.GetAstInfoResponse{}

	// Special case: if requesting SourceFile by kind, build it directly without Node conversion
	if req.Kind > 0 && ast.Kind(req.Kind) == ast.KindSourceFile {
		response.Node = builder.BuildSourceFileNodeInfo(targetSourceFile)
		// SourceFile doesn't have type/symbol/signature/flow, so return early
		return response, nil
	}

	// Find the node at the specified position (with optional end for exact matching)
	node := inspector.FindNodeAtPosition(targetSourceFile, req.Position, req.End, req.Kind)
	if node == nil {
		return &api.GetAstInfoResponse{}, nil
	}

	// Build node info
	response.Node = builder.BuildNodeInfo(node)

	// Build type info
	t := inspector.GetTypeAtNode(typeChecker, node)
	if t != nil {
		response.Type = builder.BuildTypeInfo(t)
	}

	// Build symbol info
	// First try to get symbol directly from node
	symbol := typeChecker.GetSymbolAtLocation(node)
	// If no symbol at node, try to get it from the type
	if symbol == nil && t != nil {
		symbol = t.Symbol()
	}
	if symbol != nil {
		response.Symbol = builder.BuildSymbolInfo(symbol)
	}

	// Build signature info
	sig := inspector.GetSignatureOfNode(typeChecker, node)
	if sig != nil {
		response.Signature = builder.BuildSignatureInfo(sig)
	}

	// Build flow info (only for nodes in user's source file)
	if !isExternalFile {
		flowNode := inspector.GetFlowNodeOfNode(node)
		if flowNode != nil {
			response.Flow = builder.BuildFlowInfo(flowNode)
		}
	}

	return response, nil
}

// getCachedProgram returns the cached program if it matches the current request
func getCachedProgram(fileContent, compilerOptionsJSON string) (*compiler.Program, *ast.SourceFile) {
	astInfoProgramCache.mu.RLock()
	defer astInfoProgramCache.mu.RUnlock()

	if astInfoProgramCache.program == nil {
		return nil, nil
	}

	// Check if cache is valid (only fileContent and compilerOptions matter)
	if astInfoProgramCache.fileContent == fileContent &&
		astInfoProgramCache.compilerOptions == compilerOptionsJSON {
		return astInfoProgramCache.program, astInfoProgramCache.sourceFile
	}

	return nil, nil
}

// createAndCacheProgram creates a new program and caches it
func createAndCacheProgram(fileName, fileContent, compilerOptionsJSON string, compilerOptions map[string]any) (*compiler.Program, *ast.SourceFile, error) {
	// Create a virtual filesystem with the provided file content
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	fileContents := map[string]string{
		fileName: fileContent,
	}
	fs = utils.NewOverlayVFS(fs, fileContents)

	// Build tsconfig from request options or use defaults
	tsconfigContent := buildTsConfigContent(fileName, compilerOptions)
	tsconfigPath := "/tsconfig.json"
	fs = utils.NewOverlayVFS(fs, map[string]string{
		tsconfigPath: tsconfigContent,
	})

	// Create compiler host and program
	host := utils.CreateCompilerHost("/", fs)
	program, err := utils.CreateProgram(false, fs, "/", tsconfigPath, host)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create program: %w", err)
	}

	// Get the source file
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		return nil, nil, errors.New("failed to get source file")
	}

	// Update cache
	astInfoProgramCache.mu.Lock()
	astInfoProgramCache.fileContent = fileContent
	astInfoProgramCache.compilerOptions = compilerOptionsJSON
	astInfoProgramCache.program = program
	astInfoProgramCache.sourceFile = sourceFile
	astInfoProgramCache.mu.Unlock()

	return program, sourceFile, nil
}

// buildTsConfigContent creates a tsconfig.json content string from compiler options
func buildTsConfigContent(fileName string, compilerOptions map[string]any) string {
	// Default compiler options
	opts := map[string]any{
		"target":           "ESNext",
		"module":           "ESNext",
		"strict":           true,
		"strictNullChecks": true,
	}

	// Merge with provided options (provided options override defaults)
	for k, v := range compilerOptions {
		opts[k] = v
	}

	// Serialize compiler options to JSON
	optsJSON, err := json.Marshal(opts)
	if err != nil {
		// Fallback to minimal config on error
		return fmt.Sprintf(`{"compilerOptions":{"target":"ESNext","module":"ESNext","strict":true},"files":["%s"]}`, fileName)
	}

	return fmt.Sprintf(`{"compilerOptions":%s,"files":["%s"]}`, string(optsJSON), fileName)
}

// byteOffsetToUTF16 converts a byte offset within text to a flat UTF-16 code
// unit offset — the unit ESLint uses for fix / suggestion ranges. This is a
// DIFFERENT conversion than line/column (scanner.GetECMALineAndUTF16CharacterOfPosition,
// which counts UTF-16 units from a line start); fix ranges are flat offsets
// from the start of the file.
func byteOffsetToUTF16(text string, byteOffset int) int {
	if byteOffset <= 0 {
		return 0
	}
	if byteOffset >= len(text) {
		return int(core.UTF16Len(text))
	}
	return int(core.UTF16Len(text[:byteOffset]))
}

// runAPI runs the linter in IPC mode
func runAPI() int {
	handler := &IPCHandler{}
	service := api.NewService(os.Stdin, os.Stdout, handler)

	if err := service.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error in IPC mode: %v\n", err)
		return 1
	}
	return 0
}
