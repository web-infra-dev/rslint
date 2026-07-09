package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// parallelGitignoreAndPrograms runs ReadGitignoreAsGlobs and createProgramsForConfig
// for a single rslint config (single-config and legacy JSON paths).
//
// When singleThreaded is true, both run sequentially in the calling goroutine
// — honoring the user's --singleThreaded flag (no concurrency at all).
// Otherwise the two are dispatched as parallel goroutines: they have no data
// dependency, since createProgramsForConfig only reads
// entry.LanguageOptions.ParserOptions.Project (see
// LoadTsConfigsFromRslintConfig), never entry.Ignores. Calling it before vs.
// after gitignore globs are prepended is equivalent for TS Program creation.
//
// The returned config is the gitignore-augmented config (gitignore globs
// prepended when non-empty), suitable for downstream target discovery /
// GetConfigForFile.
//
// Returns: (augmentedConfig, programs, exitCode). On non-zero exitCode,
// augmentedConfig and programs may be nil/partial — caller should propagate
// exitCode without using them.
func parallelGitignoreAndPrograms(
	rslintConfig rslintconfig.RslintConfig,
	configDir string,
	fsys vfs.FS,
	singleThreaded bool,
	seenTsConfigs map[string]struct{},
	parseCache *utils.ParseCache,
) (rslintconfig.RslintConfig, []*compiler.Program, int) {
	configIgnores := rslintconfig.ExtractConfigIgnores(rslintConfig)

	var (
		gitGlobs []string
		progs    []*compiler.Program
		exitCode int
	)
	// gitignore reading and program creation are independent
	// (createProgramsForConfig only reads parserOptions.project, never Ignores),
	// so run them on the shared WorkGroup — which honors --singleThreaded the
	// same way the lint and type-check phases do.
	wg := core.NewWorkGroup(singleThreaded)
	wg.Queue(func() {
		gitGlobs = rslintconfig.ReadGitignoreAsGlobs(configDir, fsys, configIgnores)
	})
	wg.Queue(func() {
		progs, exitCode = createProgramsForConfig(configDir, rslintConfig, singleThreaded, fsys, seenTsConfigs, parseCache)
	})
	wg.RunAndWait()

	if exitCode != 0 {
		return rslintConfig, nil, exitCode
	}
	if len(gitGlobs) > 0 {
		rslintConfig = append(
			rslintconfig.RslintConfig{{Ignores: gitGlobs}},
			rslintConfig...,
		)
	}
	return rslintConfig, progs, 0
}

// createProgramsForConfig creates tsconfig-backed TypeScript programs for a
// single config entry. Configs without a tsconfig deliberately return no
// Program here; buildProgramsWithLintTargets later creates an AST-only fallback
// from the final files-driven lint target set.
// seenTsConfigs is used for cross-config tsconfig deduplication (pass nil to skip).
// Returns the created programs or an error exit code (> 0).
func createProgramsForConfig(
	configDir string,
	entries rslintconfig.RslintConfig,
	singleThreaded bool,
	fsys vfs.FS,
	seenTsConfigs map[string]struct{},
	parseCache *utils.ParseCache,
) ([]*compiler.Program, int) {
	tsConfigs, err := rslintconfig.ResolveTsConfigPaths(entries, configDir, fsys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil, 1
	}

	var programs []*compiler.Program
	host := utils.WithParseCache(utils.CreateCompilerHost(configDir, fsys), parseCache)

	for _, tc := range tsConfigs {
		// Cross-config deduplication
		if seenTsConfigs != nil {
			normalized := tspath.NormalizePath(tc)
			if _, exists := seenTsConfigs[normalized]; exists {
				continue
			}
			seenTsConfigs[normalized] = struct{}{}
		}

		program, err := utils.CreateProgramLenient(singleThreaded, fsys, configDir, tc, host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating TS program: %v", err)
			return nil, 1
		}
		programs = append(programs, program)
	}

	return programs, 0
}

// createFallbackProgram creates a Program for selected lint targets not
// included in any existing Program. It uses minimal compiler options sufficient
// for AST parsing (no type checking).
func createFallbackProgram(
	gapFiles []string,
	singleThreaded bool,
	configDir string,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
) (*compiler.Program, int) {
	host := utils.WithParseCache(utils.CreateCompilerHost(configDir, fsys), parseCache)
	program, err := utils.CreateProgramFromOptionsLenient(singleThreaded, &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, gapFiles, host)
	if err != nil {
		// Non-fatal: gap files failing to parse should not block the entire run.
		// Log to stderr and skip.
		fmt.Fprintf(os.Stderr, "warning: failed to create program for %d file(s) outside tsconfig: %v\n", len(gapFiles), err)
		return nil, 0
	}
	return program, 0
}

// buildProgramsWithLintTargets resolves the exact lint target set, appends an
// AST-only fallback Program for selected files absent from every existing
// Program, and returns the per-program target plan consumed by RunLinter.
//
// The appended fallback Program carries synthesized CompilerOptions with no
// ConfigFilePath, which is how buildTypeCheckSkipMask later recognizes and
// excludes it from the CLI --type-check phase — no index needs to be threaded
// out of here.
//
// The type-info set is captured from the tsconfig Programs BEFORE the fallback
// is appended, so fallback files are deliberately absent from it. That absence
// is exactly the signal GetActiveRulesForFile keys off to filter type-aware
// rules off files with no type information — the only guard against a
// type-aware rule dereferencing a nil TypeChecker (which crashes the whole
// process). Both the CLI (executeLintPipeline) and the --api path (HandleLint)
// call this, so target discovery, fallback binding, and type-aware gating are
// identical by construction rather than by two parallel implementations kept
// in sync.
//
// configMap is non-nil only in the multi-config (stdin) path; the --api and
// single-config paths pass nil and resolve against rslintConfig +
// currentDirectory.
func buildProgramsWithLintTargets(
	programs []*compiler.Program,
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	programConfigDirs []string,
	configTargetFiles map[string][]string,
	fs vfs.FS,
	allowFiles []string,
	allowDirs []string,
	parseCache *utils.ParseCache,
	singleThreaded bool,
) ([]*compiler.Program, map[string]struct{}, []string, []string, [][]string, map[string]string) {
	var typeInfoFiles map[string]struct{}
	var capturedGapFiles []string

	programIndex := buildProgramFileIndex(programs, fs, singleThreaded)

	var targetFiles []string
	if configMap != nil {
		targetFiles = discoverLintFilesMultiConfig(configMap, configTargetFiles, fs, allowFiles, allowDirs, singleThreaded)
	} else {
		targetFiles = rslintconfig.DiscoverLintFiles(rslintConfig, currentDirectory, fs, allowFiles, allowDirs, singleThreaded)
	}

	gapFiles := make([]string, 0, len(targetFiles))
	for _, target := range targetFiles {
		if len(programIndex.candidatesFor(target, fs)) == 0 {
			gapFiles = append(gapFiles, target)
		}
	}

	if len(gapFiles) > 0 {
		// Build type-info set from existing (tsconfig) Programs BEFORE
		// appending the fallback, so fallback files are NOT in this set.
		typeInfoFiles = cloneStringSet(programIndex.files)
		capturedGapFiles = gapFiles

		fallback, _ := createFallbackProgram(gapFiles, singleThreaded, currentDirectory, fs, parseCache)
		if fallback != nil {
			addProgramToFileIndex(&programIndex, fallback, len(programs), fs, singleThreaded)
			programs = append(programs, fallback)
		}
	}

	targetsByProgram, configPathBySourcePath := assignLintTargetsToPrograms(programs, configMap, programConfigDirs, targetFiles, programIndex, fs)

	return programs, typeInfoFiles, capturedGapFiles, targetFiles, targetsByProgram, configPathBySourcePath
}

func discoverLintFilesMultiConfig(
	configMap map[string]rslintconfig.RslintConfig,
	configTargetFiles map[string][]string,
	fs vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []string {
	if len(configTargetFiles) == 0 {
		return rslintconfig.DiscoverLintFilesMultiConfig(configMap, fs, allowFiles, allowDirs, singleThreaded)
	}

	seen := make(map[string]struct{})
	var allTargets []string
	for configDir := range configMap {
		configAllowFiles := allowFiles
		configAllowDirs := allowDirs
		if targetFiles, scoped := configTargetFiles[configDir]; scoped {
			configAllowFiles = targetFiles
			configAllowDirs = nil
		}

		targets := rslintconfig.DiscoverLintFilesForConfigInMap(configMap, configDir, fs, configAllowFiles, configAllowDirs, singleThreaded)
		for _, f := range targets {
			if _, exists := seen[f]; !exists {
				seen[f] = struct{}{}
				allTargets = append(allTargets, f)
			}
		}
	}
	sort.Strings(allTargets)
	return allTargets
}

type programFileCandidate struct {
	programIndex int
	fileName     string
}

type programFileIndex struct {
	files  map[string]struct{}
	byPath map[string][]programFileCandidate
}

func (index programFileIndex) candidatesFor(path string, fs vfs.FS) []programFileCandidate {
	if len(index.byPath) == 0 {
		return nil
	}
	if candidates := index.byPath[path]; len(candidates) > 0 {
		return candidates
	}
	if fs == nil {
		return nil
	}
	realPath := fs.Realpath(path)
	if realPath == "" || realPath == path {
		return nil
	}
	return index.byPath[realPath]
}

type indexedProgram struct {
	program      *compiler.Program
	programIndex int
}

func buildProgramFileIndex(programs []*compiler.Program, fs vfs.FS, singleThreaded bool) programFileIndex {
	index := programFileIndex{
		files:  make(map[string]struct{}),
		byPath: make(map[string][]programFileCandidate),
	}
	indexedPrograms := make([]indexedProgram, 0, len(programs))
	for i, program := range programs {
		indexedPrograms = append(indexedPrograms, indexedProgram{
			program:      program,
			programIndex: i,
		})
	}
	addProgramsToFileIndex(&index, indexedPrograms, fs, singleThreaded)
	return index
}

func addProgramToFileIndex(index *programFileIndex, program *compiler.Program, programIndex int, fs vfs.FS, singleThreaded bool) {
	addProgramsToFileIndex(index, []indexedProgram{{
		program:      program,
		programIndex: programIndex,
	}}, fs, singleThreaded)
}

func addProgramsToFileIndex(index *programFileIndex, programs []indexedProgram, fs vfs.FS, singleThreaded bool) {
	if len(programs) == 0 {
		return
	}

	candidatesByName := make(map[string][]programFileCandidate)
	names := make([]string, 0)
	for _, entry := range programs {
		if entry.program == nil {
			continue
		}
		for _, sf := range entry.program.GetSourceFiles() {
			name := sf.FileName()
			candidate := programFileCandidate{programIndex: entry.programIndex, fileName: name}
			index.files[name] = struct{}{}
			index.byPath[name] = append(index.byPath[name], candidate)
			if _, seen := candidatesByName[name]; !seen {
				names = append(names, name)
			}
			candidatesByName[name] = append(candidatesByName[name], candidate)
		}
	}
	if fs == nil || len(candidatesByName) == 0 {
		return
	}

	resolved := make([]string, len(names))
	wg := core.NewWorkGroup(singleThreaded)
	for i := range names {
		wg.Queue(func() {
			if realPath := fs.Realpath(names[i]); realPath != "" && realPath != names[i] {
				resolved[i] = realPath
			}
		})
	}
	wg.RunAndWait()

	for i, realPath := range resolved {
		if realPath == "" {
			continue
		}
		index.files[realPath] = struct{}{}
		index.byPath[realPath] = append(index.byPath[realPath], candidatesByName[names[i]]...)
	}
}

func cloneStringSet(set map[string]struct{}) map[string]struct{} {
	if set == nil {
		return nil
	}
	cloned := make(map[string]struct{}, len(set))
	for key := range set {
		cloned[key] = struct{}{}
	}
	return cloned
}

// assignLintTargetsToPrograms binds resolved lint targets to exactly one
// Program. It accepts imported non-root SourceFiles; ownership/dedup is driven
// by the target plan, not by CommandLine().FileNames().
func assignLintTargetsToPrograms(
	programs []*compiler.Program,
	configMap map[string]rslintconfig.RslintConfig,
	programConfigDirs []string,
	targetFiles []string,
	programIndex programFileIndex,
	fs vfs.FS,
) ([][]string, map[string]string) {
	targetsByProgram := make([][]string, len(programs))
	if len(programs) == 0 || len(targetFiles) == 0 {
		return targetsByProgram, nil
	}

	seenProgramFile := make([]map[string]struct{}, len(programs))
	configPathBySourcePath := make(map[string]string)
	for _, target := range targetFiles {
		candidates := programIndex.candidatesFor(target, fs)
		if len(candidates) == 0 {
			continue
		}
		chosen := candidates[0]
		if configMap != nil && len(programConfigDirs) > 0 {
			ownerDir, _ := rslintconfig.FindNearestConfig(target, configMap)
			for _, candidate := range candidates {
				if candidate.programIndex < len(programConfigDirs) && programConfigDirs[candidate.programIndex] == ownerDir {
					chosen = candidate
					break
				}
			}
		}
		if seenProgramFile[chosen.programIndex] == nil {
			seenProgramFile[chosen.programIndex] = make(map[string]struct{})
		}
		if _, seen := seenProgramFile[chosen.programIndex][chosen.fileName]; seen {
			continue
		}
		seenProgramFile[chosen.programIndex][chosen.fileName] = struct{}{}
		targetsByProgram[chosen.programIndex] = append(targetsByProgram[chosen.programIndex], chosen.fileName)
		if chosen.fileName != target {
			configPathBySourcePath[chosen.fileName] = target
		}
	}
	for i := range targetsByProgram {
		sort.Strings(targetsByProgram[i])
	}
	if len(configPathBySourcePath) == 0 {
		configPathBySourcePath = nil
	}
	return targetsByProgram, configPathBySourcePath
}

func buildTypeInfoFilesForPrograms(programs []*compiler.Program, skipTypeCheck []bool, fs vfs.FS, singleThreaded bool) map[string]struct{} {
	if len(skipTypeCheck) == 0 {
		return nil
	}

	index := programFileIndex{
		files:  make(map[string]struct{}),
		byPath: make(map[string][]programFileCandidate),
	}
	indexedPrograms := make([]indexedProgram, 0, len(programs))
	for i, program := range programs {
		if i < len(skipTypeCheck) && skipTypeCheck[i] {
			continue
		}
		indexedPrograms = append(indexedPrograms, indexedProgram{
			program:      program,
			programIndex: i,
		})
	}
	addProgramsToFileIndex(&index, indexedPrograms, fs, singleThreaded)
	return index.files
}

// buildTypeCheckSkipMask returns a parallel-to-programs []bool marking which
// programs must be excluded from the type-check phase. A program is skipped
// when it was NOT built from a real tsconfig on disk — i.e. its CompilerOptions
// carry no ConfigFilePath. That covers the AST-only fallback Program used for
// selected files absent from every tsconfig-backed Program, including projects
// with no tsconfig at all.
//
// Their options are synthesized, not the user's tsconfig, so semantic
// diagnostics there are unreliable and must not be surfaced. This mirrors the
// type-checking boundary: only Programs backed by parserOptions.project or the auto-detected
// tsconfig.json participate in program-level --type-check diagnostics. Programs
// built via utils.CreateProgram -> GetParsedCommandLineOfConfigFile carry a
// non-empty ConfigFilePath.
//
// Returns nil when every program is tsconfig-backed, so callers don't allocate
// an all-false slice. Deriving from the programs keeps the CLI initial build
// and the --fix rebuild path consistent by construction.
func buildTypeCheckSkipMask(programs []*compiler.Program) []bool {
	var mask []bool
	for i, prog := range programs {
		opts := prog.Options()
		if opts == nil || opts.ConfigFilePath == "" {
			if mask == nil {
				mask = make([]bool, len(programs))
			}
			mask[i] = true
		}
	}
	return mask
}

type syntacticDiagnosticKey struct {
	path string
	code int32
	pos  int
	end  int
}

func collectTargetSyntacticDiagnostics(
	programs []*compiler.Program,
	targetsByProgram [][]string,
	skipTypeCheck []bool,
	typeCheck bool,
	typeCheckOnly bool,
	shouldReportForFile func(string) bool,
) []rule.RuleDiagnostic {
	if len(programs) == 0 || len(targetsByProgram) == 0 {
		return nil
	}

	seen := make(map[syntacticDiagnosticKey]struct{})
	var diagnostics []rule.RuleDiagnostic
	for i, program := range programs {
		// When --type-check runs, tsconfig-backed Programs surface syntactic
		// diagnostics through the type-check phase. Synthetic Programs are
		// skipped there, so plain lint still needs to report their target
		// syntax errors. In --type-check-only, lint-phase diagnostics stay off.
		coveredByTypeCheck := typeCheck && (i >= len(skipTypeCheck) || !skipTypeCheck[i])
		if coveredByTypeCheck || typeCheckOnly {
			continue
		}
		if i >= len(targetsByProgram) || len(targetsByProgram[i]) == 0 {
			continue
		}

		targets := make(map[string]struct{}, len(targetsByProgram[i]))
		for _, target := range targetsByProgram[i] {
			targets[target] = struct{}{}
		}

		ctx := context.Background()
		for target := range targets {
			if shouldReportForFile != nil && !shouldReportForFile(target) {
				continue
			}
			file := program.GetSourceFile(target)
			if file == nil {
				continue
			}
			for _, diagnostic := range program.GetSyntacticDiagnostics(ctx, file) {
				loc := diagnostic.Loc()
				key := syntacticDiagnosticKey{
					path: file.FileName(),
					code: diagnostic.Code(),
					pos:  loc.Pos(),
					end:  loc.End(),
				}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				diagnostics = append(diagnostics, rule.RuleDiagnostic{
					RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
					SourceFile:   file,
					FilePath:     file.FileName(),
					Range:        loc,
					Message:      rule.RuleMessage{Description: diagnostic.String()},
					Severity:     rule.SeverityError,
					PreFormatted: true,
				})
			}
		}
	}
	return diagnostics
}
