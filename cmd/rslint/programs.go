package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// createProgramsForConfig creates TypeScript programs for a single config entry.
// It handles tsconfig extraction, auto-detection, and deduplication.
// seenTsConfigs is used for cross-config tsconfig deduplication (pass nil to skip).
// Returns the created programs, whether a tsconfig was found for this config
// (see DiscoverGapFiles — this gates whether gap-file discovery is a no-op),
// and an error exit code (> 0).
func createProgramsForConfig(
	configDir string,
	entries rslintconfig.RslintConfig,
	singleThreaded bool,
	fsys vfs.FS,
	seenTsConfigs map[string]struct{},
	parseCache *utils.ParseCache,
) ([]*compiler.Program, bool, int) {
	tsConfigs, err := rslintconfig.ResolveTsConfigPaths(entries, configDir, fsys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil, false, 1
	}

	if len(tsConfigs) == 0 {
		// No tsconfig for this config directory: buildProgramsWithGapFallback's
		// DiscoverGapFiles call (with hasTsConfig=false) is responsible for
		// walking configDir and building the fallback Program — there is no
		// separate "no tsconfig" scan here anymore.
		return nil, false, 0
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

		program, err := utils.CreateProgram(singleThreaded, fsys, configDir, tc, host)
		if err != nil {
			w := bufio.NewWriter(os.Stderr)
			if !reportSyntacticErrors(err, w, tspath.ComparePathsOptions{
				CurrentDirectory:          configDir,
				UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
			}) {
				fmt.Fprintf(os.Stderr, "error creating TS program: %v", err)
			}
			return nil, true, 1
		}
		programs = append(programs, program)
	}

	return programs, true, 0
}

// createFallbackProgram creates a Program for "gap" files — files matched by
// config `files` patterns but not included in any tsconfig, or every source
// file in a project with no tsconfig at all. Uses minimal compiler options
// sufficient for AST parsing (no type checking).
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

// buildProgramsWithGapFallback discovers "gap" files — files matched by config
// `files` patterns (or explicitly requested and matched by config) but absent
// from every tsconfig Program, or every source file in a tsconfig-less
// project — and appends one AST-only fallback Program for them. It returns
// the (possibly extended) program slice, the type-info file set, the gap
// files retained for --fix rebuilds, and the rslintConfig/configMap updated
// with any .gitignore-derived global ignores discovered while walking.
//
// The appended fallback Program carries synthesized CompilerOptions with no
// ConfigFilePath, which is how buildTypeCheckSkipMask later recognizes and
// excludes it from the CLI --type-check phase — no index needs to be
// threaded out of here.
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
// currentDirectory. hasTsConfig / hasTsConfigByDir mirror that split: exactly
// one of them is consulted, matching which of rslintConfig/configMap is live.
//
// Callers MUST use the returned rslintConfig/configMap (not their pre-call
// values) for anything downstream — buildFileFilters, RunLinter, and any
// --fix rebuild — so that files excluded by a discovered .gitignore rule
// (even ones already inside a tsconfig Program) are correctly filtered out
// of lint results. The .gitignore-derived entry is prepended (not appended)
// to preserve existing negation precedence: gitignore rules first, user
// config `ignores` after, so a user's `!` can override a gitignore rule.
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
	hasTsConfig bool,
	hasTsConfigByDir map[string]bool,
) (
	resultPrograms []*compiler.Program,
	typeInfoFiles map[string]struct{},
	capturedGapFiles []string,
	updatedRslintConfig rslintconfig.RslintConfig,
	updatedConfigMap map[string]rslintconfig.RslintConfig,
) {
	programFiles := utils.CollectProgramFiles(programs, fs, singleThreaded)

	var gapFiles []string
	if configMap != nil {
		var gitignoreGlobsByDir map[string][]string
		gapFiles, gitignoreGlobsByDir = rslintconfig.DiscoverGapFilesMultiConfig(
			configMap, fs, programFiles, allowFiles, allowDirs, singleThreaded, hasTsConfigByDir,
		)
		if len(gitignoreGlobsByDir) > 0 {
			updatedConfigMap = make(map[string]rslintconfig.RslintConfig, len(configMap))
			for dir, cfg := range configMap {
				updatedConfigMap[dir] = cfg
			}
			for dir, globs := range gitignoreGlobsByDir {
				updatedConfigMap[dir] = append(
					rslintconfig.RslintConfig{{Ignores: globs}},
					updatedConfigMap[dir]...,
				)
			}
		} else {
			updatedConfigMap = configMap
		}
		updatedRslintConfig = rslintConfig
	} else {
		var gitignoreGlobs []string
		gapFiles, gitignoreGlobs = rslintconfig.DiscoverGapFiles(
			rslintConfig, currentDirectory, fs, programFiles, allowFiles, allowDirs, singleThreaded, hasTsConfig,
		)
		if len(gitignoreGlobs) > 0 {
			updatedRslintConfig = append(
				rslintconfig.RslintConfig{{Ignores: gitignoreGlobs}},
				rslintConfig...,
			)
		} else {
			updatedRslintConfig = rslintConfig
		}
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
			if updatedConfigMap != nil {
				cfgDir, cfg := rslintconfig.FindNearestConfig(nf, updatedConfigMap)
				if cfg != nil {
					merged = cfg.GetConfigForFile(nf, cfgDir)
				}
			} else {
				merged = updatedRslintConfig.GetConfigForFile(nf, currentDirectory)
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
			fallback, _ := createFallbackProgram(gapFiles, singleThreaded, currentDirectory, fs, parseCache)
			if fallback != nil {
				programs = append(programs, fallback)
			}
		}
	}

	return programs, typeInfoFiles, capturedGapFiles, updatedRslintConfig, updatedConfigMap
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
// carry no ConfigFilePath. That is the gap-file fallback Program
// (createFallbackProgram): its options are synthesized, not the user's
// tsconfig, so semantic diagnostics there are unreliable and must not be
// surfaced. Concretely, it sets neither Target nor Lib beyond ESNext defaults,
// so it can lack modern globals or emit spurious diagnostics against typings
// pulled in from node_modules. This mirrors the type-checking boundary: only
// Programs backed by parserOptions.project or the auto-detected tsconfig.json
// participate in program-level --type-check diagnostics. Programs built via
// utils.CreateProgram -> GetParsedCommandLineOfConfigFile carry a non-empty
// ConfigFilePath.
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
