package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

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
) (rslintconfig.RslintConfig, []*compiler.Program, int) {
	configIgnores := rslintconfig.ExtractConfigIgnores(rslintConfig)

	var (
		gitGlobs []string
		progs    []*compiler.Program
		exitCode int
	)
	if singleThreaded {
		gitGlobs = rslintconfig.ReadGitignoreAsGlobs(configDir, fsys, configIgnores)
		progs, exitCode = createProgramsForConfig(configDir, rslintConfig, singleThreaded, fsys, seenTsConfigs)
	} else {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			gitGlobs = rslintconfig.ReadGitignoreAsGlobs(configDir, fsys, configIgnores)
		}()
		go func() {
			defer wg.Done()
			progs, exitCode = createProgramsForConfig(configDir, rslintConfig, singleThreaded, fsys, seenTsConfigs)
		}()
		wg.Wait()
	}

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
) ([]*compiler.Program, int) {
	tsConfigs, err := rslintconfig.ResolveTsConfigPaths(entries, configDir, fsys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil, 1
	}

	var programs []*compiler.Program
	host := utils.CreateCompilerHost(configDir, fsys)

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

// createFallbackProgram creates a Program for "gap" files — files matched by
// config `files` patterns but not included in any tsconfig. Uses minimal
// compiler options sufficient for AST parsing (no type checking).
func createFallbackProgram(
	gapFiles []string,
	singleThreaded bool,
	configDir string,
	fsys vfs.FS,
) (*compiler.Program, int) {
	host := utils.CreateCompilerHost(configDir, fsys)
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
//     from ALL diagnostics (lint rules, type-check, and the linted-file count)
//
// The returned slice is always len(programs). Entries are never nil — ignores
// must apply to every program, including the gap-file fallback program.
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

// buildTypeCheckSkipMask returns a parallel-to-Programs []bool where the entry
// at fallbackIdx is true (skip type-check) and all others are false. Returns
// nil when fallbackIdx is -1 (no fallback program), so callers don't allocate
// an all-false slice.
func buildTypeCheckSkipMask(numPrograms int, fallbackIdx int) []bool {
	if fallbackIdx < 0 {
		return nil
	}
	mask := make([]bool, numPrograms)
	if fallbackIdx < numPrograms {
		mask[fallbackIdx] = true
	}
	return mask
}
