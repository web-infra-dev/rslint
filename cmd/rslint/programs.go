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
	"github.com/web-infra-dev/rslint/internal/utils"
)

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
		rootFiles := vfs.ReadDirectory(fsys, configDir, configDir, sourceExts, excludes, includes, nil)
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
