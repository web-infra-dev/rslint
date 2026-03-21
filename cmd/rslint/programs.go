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
	loader := rslintconfig.NewConfigLoader(fsys, configDir)
	tsConfigs, err := loader.LoadTsConfigsFromRslintConfig(entries, configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil, 1
	}

	// Auto-detect tsconfig.json if none specified in config
	if len(tsConfigs) == 0 {
		defaultTsConfig := tspath.ResolvePath(configDir, "tsconfig.json")
		if fsys.FileExists(defaultTsConfig) {
			tsConfigs = []string{defaultTsConfig}
		}
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
		excludes := []string{"node_modules"}
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
