package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func fallbackCompilerOptions() *core.CompilerOptions {
	return &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}
}

// programConfigOrders maps normalized config-directory identities to the
// declaration order of one Program's tsconfig in each config. A shared path
// has one Program instance but may have multiple associations.
type programConfigOrders map[string]int

// compilerProgramSet is the unique set of ts-go Programs used while assembling
// a lint run. ConfigOrders is parallel to CompilerPrograms. Compatibility
// Programs are appended only while binding targets and have no config entry.
type compilerProgramSet struct {
	CompilerPrograms []*compiler.Program
	ConfigOrders     []programConfigOrders
}

// programBuildSpec is one stable slot in the Program registry. Planning and
// construction are deliberately separate: config/project order and shared
// tsconfig associations are resolved serially, while independent Program
// construction may fill the already-ordered slots concurrently.
type programBuildSpec struct {
	tsconfigPath string
	programCwd   string
	configOrders programConfigOrders
}

type programBuildPlan struct {
	specs       []programBuildSpec
	terminalErr error
}

func exactFilesystemPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

func authoritativeFilesystemPath(filePath string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
	}
	return filePath
}

func canonicalFilesystemPathID(filePath string, fsys vfs.FS) string {
	return exactFilesystemPathID(authoritativeFilesystemPath(filePath, fsys))
}

// configPathForLintTarget returns the target path used for files/ignores
// matching. Program source names are deliberately excluded: adding or removing
// a file from a TypeScript Program must not change its lint configuration.
func configPathForLintTarget(target resolvedLintTarget, fsys vfs.FS) string {
	matchPath, _ := rslintconfig.ResolveConfigPathSpaceWithCanonical(
		target.Path,
		target.CanonicalPath,
		target.OwnerConfigDir,
		fsys,
	)
	return matchPath
}

func storeSourcePathMapping(mapping map[string]string, sourcePath string, canonicalSourcePath string, value string) {
	if mapping == nil {
		return
	}
	normalizedSource := tspath.NormalizePath(sourcePath)
	mapping[normalizedSource] = value
	if canonicalSourcePath != "" {
		mapping[exactFilesystemPathID(canonicalSourcePath)] = value
	}
}

// buildProgramPlan resolves each normalized tsconfig path once while retaining
// every config that declared it. It stops at the first resolution failure but
// retains earlier work: the serial implementation built those earlier Programs
// before surfacing that failure, so executeProgramBuildPlan must preserve the
// same error precedence.
func buildProgramPlan(
	configMap map[string]rslintconfig.RslintConfig,
	fsys vfs.FS,
) programBuildPlan {
	if len(configMap) == 0 {
		return programBuildPlan{}
	}

	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	plan := programBuildPlan{}
	programByTsconfig := make(map[string]int)
	for _, configDir := range configDirs {
		entries := configMap[configDir]
		normalizedConfigDir := tspath.NormalizePath(configDir)
		configDirID := exactFilesystemPathID(normalizedConfigDir)
		tsConfigs, err := rslintconfig.ResolveTsConfigPaths(entries, normalizedConfigDir, fsys)
		if err != nil {
			plan.terminalErr = fmt.Errorf("resolve tsconfigs for %q: %w", configDir, err)
			return plan
		}

		for order, tsconfigPath := range tsConfigs {
			tsconfigPath = tspath.NormalizePath(tsconfigPath)
			tsconfigID := exactFilesystemPathID(tsconfigPath)
			if programIndex, ok := programByTsconfig[tsconfigID]; ok {
				if _, alreadyAssociated := plan.specs[programIndex].configOrders[configDirID]; !alreadyAssociated {
					plan.specs[programIndex].configOrders[configDirID] = order
				}
				continue
			}

			// Relative paths in a tsconfig are resolved from the declared path,
			// including when that path is a file symlink. This matches tsc/tsgo;
			// realpath is only a source-identity fallback during target binding.
			programByTsconfig[tsconfigID] = len(plan.specs)
			plan.specs = append(plan.specs, programBuildSpec{
				tsconfigPath: tsconfigPath,
				programCwd:   tspath.GetDirectoryPath(tsconfigPath),
				configOrders: programConfigOrders{configDirID: order},
			})
		}
	}

	return plan
}

func executeProgramBuildPlan(
	plan programBuildPlan,
	singleThreaded bool,
	buildContext *utils.ProgramBuildContext,
) (compilerProgramSet, error) {
	if len(plan.specs) == 0 {
		if plan.terminalErr != nil {
			return compilerProgramSet{}, plan.terminalErr
		}
		return compilerProgramSet{}, nil
	}

	programs := make([]*compiler.Program, len(plan.specs))
	errs := make([]error, len(plan.specs))

	workerCount := min(runtime.GOMAXPROCS(0), len(plan.specs))
	parallel := !singleThreaded && workerCount > 1
	if parallel {
		buildContext.EnableConcurrentProgramQueries()
	}
	build := func(index int) {
		spec := plan.specs[index]
		programs[index], errs[index] = buildContext.CreateProgramLenient(
			singleThreaded,
			spec.programCwd,
			spec.tsconfigPath,
		)
	}

	if !parallel {
		for index := range plan.specs {
			build(index)
			if errs[index] != nil {
				break
			}
		}
	} else {
		jobs := make(chan int, workerCount)
		var workers sync.WaitGroup
		workers.Add(workerCount)
		for range workerCount {
			go func() {
				defer workers.Done()
				for index := range jobs {
					build(index)
				}
			}()
		}
		for index := range plan.specs {
			jobs <- index
		}
		close(jobs)
		workers.Wait()
	}

	for index, err := range errs {
		if err != nil {
			return compilerProgramSet{}, fmt.Errorf(
				"create TypeScript Program from %q: %w",
				plan.specs[index].tsconfigPath,
				err,
			)
		}
	}
	if plan.terminalErr != nil {
		return compilerProgramSet{}, plan.terminalErr
	}

	configOrders := make([]programConfigOrders, len(plan.specs))
	for index := range plan.specs {
		configOrders[index] = plan.specs[index].configOrders
	}
	return compilerProgramSet{CompilerPrograms: programs, ConfigOrders: configOrders}, nil
}

// createProgramSetForConfigs builds each planned Program into its stable
// registry slot. At most min(GOMAXPROCS, Program count) Programs are built
// concurrently; --singleThreaded retains the original serial, fail-fast path
// exactly.
func createProgramSetForConfigs(
	configMap map[string]rslintconfig.RslintConfig,
	singleThreaded bool,
	buildContext *utils.ProgramBuildContext,
) (compilerProgramSet, error) {
	plan := buildProgramPlan(configMap, buildContext.FS())
	return executeProgramBuildPlan(plan, singleThreaded, buildContext)
}

func createProgramSetForConfig(
	configDir string,
	entries rslintconfig.RslintConfig,
	singleThreaded bool,
	buildContext *utils.ProgramBuildContext,
) (compilerProgramSet, error) {
	return createProgramSetForConfigs(
		map[string]rslintconfig.RslintConfig{configDir: entries},
		singleThreaded,
		buildContext,
	)
}

// parallelGitignoreAndPrograms reads gitignore state and builds the Program
// registry for an invocation-wide JSON/JSONC config. Staged JS/TS catalogs
// already contain their frozen Git projection before the lint pipeline starts.
//
// When singleThreaded is true, both run sequentially in the calling goroutine
// — honoring the user's --singleThreaded flag (no concurrency at all).
// Otherwise the two are dispatched as parallel goroutines: they have no data
// dependency, since Program creation only reads
// entry.LanguageOptions.ParserOptions.Project (see
// LoadTsConfigsFromRslintConfig), never entry.Ignores. Calling it before vs.
// after .gitignore patterns are injected is equivalent for TS Program creation.
//
// The returned config carries the collected .gitignore patterns used by
// downstream target admission. File-only calls can supply an exact target set.
func parallelGitignoreAndPrograms(
	rslintConfig rslintconfig.RslintConfig,
	configDir string,
	targetFiles []string,
	singleThreaded bool,
	buildContext *utils.ProgramBuildContext,
) (rslintconfig.RslintConfig, compilerProgramSet, error) {
	fsys := buildContext.FS()
	var (
		configWithIgnores rslintconfig.RslintConfig
		programs          compilerProgramSet
		programErr        error
	)
	// .gitignore collection and program creation are independent
	// (Program creation only reads parserOptions.project, never Ignores),
	// so run them on the shared WorkGroup — which honors --singleThreaded the
	// same way the lint and type-check phases do.
	wg := core.NewWorkGroup(singleThreaded)
	wg.Queue(func() {
		configWithIgnores = rslintconfig.ConfigWithGitignore(rslintConfig, configDir, fsys, targetFiles)
	})
	wg.Queue(func() {
		programs, programErr = createProgramSetForConfig(configDir, rslintConfig, singleThreaded, buildContext)
	})
	wg.RunAndWait()

	if programErr != nil {
		return rslintConfig, compilerProgramSet{}, programErr
	}
	return configWithIgnores, programs, nil
}

// createFallbackProgram creates a Program for selected lint targets not
// included in any existing Program. It uses minimal compiler options sufficient
// for AST parsing (no type checking).
func createFallbackProgram(
	gapFiles []string,
	singleThreaded bool,
	configDir string,
	buildContext *utils.ProgramBuildContext,
) (*compiler.Program, error) {
	program, err := buildContext.CreateProgramFromOptionsLenient(singleThreaded, configDir, fallbackCompilerOptions(), gapFiles)
	if err != nil {
		return nil, fmt.Errorf("create fallback Program for %d lint target(s): %w", len(gapFiles), err)
	}
	return program, nil
}
