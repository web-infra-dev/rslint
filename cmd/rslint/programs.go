package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/vfsmatch"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
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

// createProgramsForConfig creates TypeScript programs for a single config entry.
// It handles tsconfig extraction, auto-detection, deduplication, and the no-tsconfig fallback.
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

	if len(tsConfigs) > 0 {
		for _, tc := range tsConfigs {
			// Cross-config deduplication
			if seenTsConfigs != nil {
				normalized := tspath.NormalizePath(tc)
				if _, exists := seenTsConfigs[normalized]; exists {
					continue
				}
				seenTsConfigs[normalized] = struct{}{}
			}

			program, err := utils.CreateProgram(singleThreaded, fsys, configDir, tc, host)
			if err != nil {
				w := bufio.NewWriter(os.Stderr)
				if !reportSyntacticErrors(err, w, tspath.ComparePathsOptions{
					CurrentDirectory:          configDir,
					UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
				}) {
					fmt.Fprintf(os.Stderr, "error creating TS program: %v", err)
				}
				return nil, 1
			}
			programs = append(programs, program)
		}
	} else {
		// No tsconfig fallback: scan directory for pure JS projects
		sourceExts := rslintconfig.DefaultLintFileExtensions
		excludes := utils.DefaultExcludeDirNames
		includes := []string{"**/*"}
		rootFiles := vfsmatch.ReadDirectory(fsys, configDir, configDir, sourceExts, excludes, includes, vfsmatch.UnlimitedDepth)
		if len(rootFiles) > 0 {
			program, err := utils.CreateProgramFromOptions(singleThreaded, &core.CompilerOptions{AllowJs: core.TSTrue}, rootFiles, host)
			if err != nil {
				w := bufio.NewWriter(os.Stderr)
				if !reportSyntacticErrors(err, w, tspath.ComparePathsOptions{
					CurrentDirectory:          configDir,
					UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
				}) {
					fmt.Fprintf(os.Stderr, "error creating program: %v", err)
				}
				return nil, 1
			}
			programs = append(programs, program)
		}
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
		Target:  core.ScriptTargetESNext,
		Module:  core.ModuleKindESNext,
		Jsx:     core.JsxEmitPreserve,
		AllowJs: core.TSTrue,
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
// The appended fallback Program (like the no-tsconfig directory-scan Program)
// carries synthesized CompilerOptions with no ConfigFilePath, which is how
// buildTypeCheckSkipMask later recognizes and excludes it from the CLI
// --type-check phase — no index needs to be threaded out of here.
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
) ([]*compiler.Program, map[string]struct{}, []string, []string, [][]string) {
	var typeInfoFiles map[string]struct{}
	var capturedGapFiles []string

	programFiles := utils.CollectProgramFiles(programs, fs, singleThreaded)

	var targetFiles []string
	if configMap != nil {
		targetFiles = discoverLintFilesMultiConfig(configMap, configTargetFiles, fs, allowFiles, allowDirs, singleThreaded)
	} else {
		targetFiles = rslintconfig.DiscoverLintFiles(rslintConfig, currentDirectory, fs, allowFiles, allowDirs, singleThreaded)
	}

	gapFiles := make([]string, 0, len(targetFiles))
	for _, target := range targetFiles {
		if _, inProgram := programFiles[target]; !inProgram {
			gapFiles = append(gapFiles, target)
		}
	}

	if len(gapFiles) > 0 {
		// Build type-info set from existing (tsconfig) Programs BEFORE
		// appending the fallback, so fallback files are NOT in this set.
		typeInfoFiles = utils.CollectProgramFiles(programs, fs, singleThreaded)
		capturedGapFiles = gapFiles

		fallback, _ := createFallbackProgram(gapFiles, singleThreaded, currentDirectory, fs, parseCache)
		if fallback != nil {
			programs = append(programs, fallback)
		}
	}

	targetsByProgram := assignLintTargetsToPrograms(programs, configMap, programConfigDirs, targetFiles, fs)

	return programs, typeInfoFiles, capturedGapFiles, targetFiles, targetsByProgram
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
	for configDir, cfg := range configMap {
		configAllowFiles := allowFiles
		configAllowDirs := allowDirs
		if targetFiles, scoped := configTargetFiles[configDir]; scoped {
			configAllowFiles = targetFiles
			configAllowDirs = nil
		}

		targets := rslintconfig.DiscoverLintFiles(cfg, configDir, fs, configAllowFiles, configAllowDirs, singleThreaded)
		for _, f := range targets {
			ownerDir, _ := rslintconfig.FindNearestConfig(f, configMap)
			if ownerDir != configDir {
				continue
			}
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

// assignLintTargetsToPrograms binds resolved lint targets to exactly one
// Program. It accepts imported non-root SourceFiles; ownership/dedup is driven
// by the target plan, not by CommandLine().FileNames().
func assignLintTargetsToPrograms(
	programs []*compiler.Program,
	configMap map[string]rslintconfig.RslintConfig,
	programConfigDirs []string,
	targetFiles []string,
	fs vfs.FS,
) [][]string {
	targetsByProgram := make([][]string, len(programs))
	if len(programs) == 0 || len(targetFiles) == 0 {
		return targetsByProgram
	}

	byPath := make(map[string][]programFileCandidate)
	for i, prog := range programs {
		for _, sf := range prog.GetSourceFiles() {
			name := sf.FileName()
			candidate := programFileCandidate{programIndex: i, fileName: name}
			byPath[name] = append(byPath[name], candidate)
			if fs != nil {
				if real := fs.Realpath(name); real != "" && real != name {
					byPath[real] = append(byPath[real], candidate)
				}
			}
		}
	}

	seenProgramFile := make([]map[string]struct{}, len(programs))
	for _, target := range targetFiles {
		candidates := byPath[target]
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
	}
	for i := range targetsByProgram {
		sort.Strings(targetsByProgram[i])
	}
	return targetsByProgram
}

// buildFileFilters returns per-program file filters for config `ignores`.
// Target ownership/deduplication is handled by DiscoverLintFilesMultiConfig and
// assignLintTargetsToPrograms before RunLinter receives TargetFiles.
//
// These filters are consumed by RunLinter only in Phase 1 (lint). Phase 2
// (type-check) does NOT consult them — type-check mirrors `tsgo --noEmit`
// over the full tsconfig-determined program. See linter.go RunLinter doc
// and website/docs/en/guide/type-checking.md for the contract.
//
// The returned slice is always len(programs). Entries are never nil — ignore
// semantics must apply to every program in Phase 1, including the gap-file
// fallback program.
//
// singleConfig / singleConfigDir are used when configMap is nil (single-config
// mode). When configMap is non-nil, the per-file nearest config is looked up.
func buildFileFilters(
	programs []*compiler.Program,
	configMap map[string]rslintconfig.RslintConfig,
	programConfigDirs []string,
	singleConfig rslintconfig.RslintConfig,
	singleConfigDir string,
) []func(string) bool {
	filters := make([]func(string) bool, len(programs))
	for i := range programs {
		filters[i] = func(fileName string) bool {
			// Ignore check: resolve the config that governs this file and
			// consult its global `ignores` patterns.
			var cfg rslintconfig.RslintConfig
			var cwd string
			if configMap != nil {
				cwd, cfg = rslintconfig.FindNearestConfig(fileName, configMap)
			} else {
				cfg = singleConfig
				cwd = singleConfigDir
			}
			if cfg != nil && cfg.IsFileIgnored(fileName, cwd) {
				return false
			}
			return true
		}
	}
	return filters
}

// toFileFilters converts the legacy `[]func(string) bool` per-program filter
// slice (used by buildFileFilters) into linter.FileFilter typed entries. The
// underlying functions are identical; this is purely a type adapter.
func toFileFilters(in []func(string) bool) []linter.FileFilter {
	if in == nil {
		return nil
	}
	out := make([]linter.FileFilter, len(in))
	for i, f := range in {
		out[i] = f
	}
	return out
}

// buildTypeCheckSkipMask returns a parallel-to-programs []bool marking which
// programs must be excluded from the type-check phase. A program is skipped
// when it was NOT built from a real tsconfig on disk — i.e. its CompilerOptions
// carry no ConfigFilePath. That covers both synthetic-options program kinds:
//
//   - the no-tsconfig directory-scan Program (createProgramsForConfig's else
//     branch), built with default options for a config directory that has no
//     tsconfig; and
//   - the gap-file fallback Program (createFallbackProgram).
//
// Their options are synthesized, not the user's tsconfig, so semantic
// diagnostics there are unreliable and must not be surfaced. Concretely, the
// scan program sets neither Target nor Lib, so its default lib can lack modern
// globals like Symbol.asyncDispose and emit spurious diagnostics against
// typings pulled in from node_modules. This mirrors the type-checking boundary:
// only Programs backed by parserOptions.project or the auto-detected
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
