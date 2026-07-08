package main

import (
	"bufio"
	"fmt"
	"os"

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
// prepended when non-empty), suitable for downstream DiscoverGapFiles /
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
		sourceExts := []string{".ts", ".tsx", ".js", ".jsx", ".mts", ".mjs"}
		excludes := utils.DefaultExcludeDirNames
		includes := []string{"**/*"}
		rootFiles := vfsmatch.ReadDirectory(fsys, configDir, configDir, sourceExts, excludes, includes, vfsmatch.UnlimitedDepth)
		if len(rootFiles) > 0 {
			jsxFactory, jsxFragmentFactory := rslintconfig.ResolveJsxPragmaOptions(entries)
			program, err := utils.CreateProgramFromOptions(singleThreaded, &core.CompilerOptions{
				AllowJs:            core.TSTrue,
				JsxFactory:         jsxFactory,
				JsxFragmentFactory: jsxFragmentFactory,
			}, rootFiles, host)
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

// createFallbackProgram creates a Program for "gap" files — files matched by
// config `files` patterns but not included in any tsconfig. Uses minimal
// compiler options sufficient for AST parsing (no type checking).
// Empty jsxFactory / jsxFragmentFactory leave TypeScript's own "React" /
// "Fragment" defaults in place.
func createFallbackProgram(
	gapFiles []string,
	singleThreaded bool,
	configDir string,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
	jsxFactory string,
	jsxFragmentFactory string,
) (*compiler.Program, int) {
	host := utils.WithParseCache(utils.CreateCompilerHost(configDir, fsys), parseCache)
	program, err := utils.CreateProgramFromOptionsLenient(singleThreaded, &core.CompilerOptions{
		Target:             core.ScriptTargetESNext,
		Module:             core.ModuleKindESNext,
		Jsx:                core.JsxEmitPreserve,
		AllowJs:            core.TSTrue,
		JsxFactory:         jsxFactory,
		JsxFragmentFactory: jsxFragmentFactory,
	}, gapFiles, host)
	if err != nil {
		// Non-fatal: gap files failing to parse should not block the entire run.
		// Log to stderr and skip.
		fmt.Fprintf(os.Stderr, "warning: failed to create program for %d file(s) outside tsconfig: %v\n", len(gapFiles), err)
		return nil, 0
	}
	return program, 0
}

// buildProgramsWithGapFallback discovers "gap" files — files matched by config
// `files` patterns (or explicitly requested and matched by config) but absent
// from every tsconfig Program — and appends one AST-only fallback Program for
// them. It returns the (possibly extended) program slice, the type-info file
// set, and the gap files retained for --fix rebuilds.
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
// call this, so their gap / fallback / gate behavior is identical by
// construction rather than by two parallel implementations kept in sync.
//
// configMap is non-nil only in the multi-config (stdin) path; the --api and
// single-config paths pass nil and resolve against rslintConfig +
// currentDirectory.
func buildProgramsWithGapFallback(
	programs []*compiler.Program,
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	fs vfs.FS,
	allowFiles []string,
	allowDirs []string,
	parseCache *utils.ParseCache,
	singleThreaded bool,
) ([]*compiler.Program, map[string]struct{}, []string) {
	var typeInfoFiles map[string]struct{}
	var capturedGapFiles []string

	programFiles := utils.CollectProgramFiles(programs, fs, singleThreaded)

	var gapFiles []string
	if configMap != nil {
		gapFiles = rslintconfig.DiscoverGapFilesMultiConfig(configMap, fs, programFiles, allowFiles, allowDirs, singleThreaded)
	} else {
		gapFiles = rslintconfig.DiscoverGapFiles(rslintConfig, currentDirectory, fs, programFiles, allowFiles, allowDirs, singleThreaded)
	}

	// Explicit file args bypass config `files` patterns (ESLint behavior):
	// if a file is explicitly requested, lint it even if no config entry has a
	// matching `files` pattern — as long as the config would assign it rules.
	if gapFiles != nil && len(allowFiles) > 0 {
		gapSet := make(map[string]struct{}, len(gapFiles))
		for _, f := range gapFiles {
			gapSet[f] = struct{}{}
		}
		for _, f := range allowFiles {
			nf := tspath.NormalizePath(f)
			if _, inProgram := programFiles[nf]; inProgram {
				continue
			}
			if _, alreadyGap := gapSet[nf]; alreadyGap {
				continue
			}
			// Check if config would assign any rules to this file.
			var merged *rslintconfig.MergedConfig
			if configMap != nil {
				cfgDir, cfg := rslintconfig.FindNearestConfig(nf, configMap)
				if cfg != nil {
					merged = cfg.GetConfigForFile(nf, cfgDir)
				}
			} else {
				merged = rslintConfig.GetConfigForFile(nf, currentDirectory)
			}
			if merged != nil {
				gapFiles = append(gapFiles, nf)
			}
		}
	}

	if gapFiles != nil {
		// Build type-info set from existing (tsconfig) Programs BEFORE
		// appending the fallback, so fallback files are NOT in this set.
		typeInfoFiles = utils.CollectProgramFiles(programs, fs, singleThreaded)
		capturedGapFiles = gapFiles

		if len(gapFiles) > 0 {
			jsxFactory, jsxFragmentFactory := resolveGapJsxPragma(configMap, rslintConfig)
			fallback, _ := createFallbackProgram(gapFiles, singleThreaded, currentDirectory, fs, parseCache, jsxFactory, jsxFragmentFactory)
			if fallback != nil {
				programs = append(programs, fallback)
			}
		}
	}

	return programs, typeInfoFiles, capturedGapFiles
}

// resolveGapJsxPragma resolves the jsxPragma / jsxFragmentName to seed the
// gap-file fallback Program's CompilerOptions with. In multi-config mode
// (configMap non-nil) the gap-file batch can span multiple config
// directories, so this unions across all of them.
func resolveGapJsxPragma(configMap map[string]rslintconfig.RslintConfig, rslintConfig rslintconfig.RslintConfig) (jsxFactory, jsxFragmentFactory string) {
	if configMap != nil {
		for _, entries := range configMap {
			f, ff := rslintconfig.ResolveJsxPragmaOptions(entries)
			if f != "" {
				jsxFactory = f
			}
			if ff != "" {
				jsxFragmentFactory = ff
			}
		}
		return jsxFactory, jsxFragmentFactory
	}
	return rslintconfig.ResolveJsxPragmaOptions(rslintConfig)
}

// buildFileOwnerMap determines which config directory "owns" each file across
// all programs. Ownership is based on nearest config lookup (deepest matching
// configDirectory), aligning with ESLint v10's per-file config resolution.
// This ensures each file is linted exactly once by the program belonging to
// its nearest config.
func buildFileOwnerMap(programs []*compiler.Program, configMap map[string]rslintconfig.RslintConfig) map[string]string {
	fileOwner := make(map[string]string)
	for _, prog := range programs {
		for _, sf := range prog.GetSourceFiles() {
			fn := sf.FileName()
			if _, ok := fileOwner[fn]; !ok {
				nearestDir, _ := rslintconfig.FindNearestConfig(fn, configMap)
				fileOwner[fn] = nearestDir
			}
		}
	}
	return fileOwner
}

// buildFileFilters returns per-program file filters combining two concerns:
//   - multi-config ownership: a file is linted by the program belonging to its
//     nearest config (only active when len(configMap) > 1)
//   - config `ignores`: files matching the user's ignore patterns are excluded
//     from lint rules and the linted-file count
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
	var fileOwner map[string]string
	if len(configMap) > 1 {
		fileOwner = buildFileOwnerMap(programs, configMap)
	}

	filters := make([]func(string) bool, len(programs))
	for i := range programs {
		var ownerDir string
		if fileOwner != nil && i < len(programConfigDirs) {
			ownerDir = programConfigDirs[i]
		}
		filters[i] = func(fileName string) bool {
			// Ownership check: only when we have multiple configs AND this
			// program is anchored to a configDir (gap fallback has "").
			if fileOwner != nil && ownerDir != "" {
				if owner, ok := fileOwner[fileName]; ok && owner != ownerDir {
					return false
				}
			}
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
