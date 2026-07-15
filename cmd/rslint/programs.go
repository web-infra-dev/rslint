package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/compilerpath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/hostpath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// programConfigOrders maps normalized config-directory identities to the
// declaration order of one Program's tsconfig in each config. A shared path
// has one Program instance but may have multiple associations.
type programConfigOrders map[string]int

// lintProgramSet is the unique set of real tsconfig-backed Programs used by a
// lint run. ConfigOrders is parallel to Programs. Synthetic fallback Programs
// are appended only while binding a lint target plan and have no config entry.
type lintProgramSet struct {
	Programs     []*compiler.Program
	ConfigOrders []programConfigOrders
}

func compilerPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

func compilerCanRepresentHostPath(filePath string) bool {
	return compilerpath.CanRepresent(filePath)
}

func exactHostPathID(filePath string) string {
	return hostpath.NormalizeForRoot(filePath, filePath)
}

func authoritativeHostPath(filePath string, fsys vfs.FS) string {
	// Use the path's own root as the governing syntax so synthetic Windows
	// paths in cross-platform tests retain drive semantics on POSIX hosts.
	// Authored relative values never reach this physical-identity boundary.
	filePath = hostpath.NormalizeForRoot(filePath, filePath)
	// Compiler-owned virtual paths (for example bundled:///libs/lib.es5.d.ts)
	// are not host filesystem operands. osvfs.Realpath intentionally requires
	// an absolute host path and panics for these schemes.
	if fsys != nil && hostpath.IsAbsoluteForRoot(filePath, filePath) {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			return hostpath.NormalizeForRoot(realPath, realPath)
		}
	}
	return filePath
}

func canonicalHostPathID(filePath string, fsys vfs.FS) string {
	return exactHostPathID(authoritativeHostPath(filePath, fsys))
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
	normalizedSource := compilerPathID(sourcePath)
	mapping[normalizedSource] = value
	if canonicalSourcePath != "" && compilerCanRepresentHostPath(canonicalSourcePath) {
		mapping[compilerPathID(canonicalSourcePath)] = value
	}
}

// createProgramSetForConfigs builds each normalized tsconfig path once while
// retaining every config that declared it. Config roots are processed in a
// stable order; target binding later uses the per-config project order rather
// than map iteration or Program construction order.
func createProgramSetForConfigs(
	configMap map[string]rslintconfig.RslintConfig,
	singleThreaded bool,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
) (lintProgramSet, error) {
	if len(configMap) == 0 {
		return lintProgramSet{}, nil
	}

	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	set := lintProgramSet{}
	programByTsconfig := make(map[string]int)
	for _, configDir := range configDirs {
		entries := configMap[configDir]
		normalizedConfigDir := tspath.NormalizePath(configDir)
		configDirID := exactHostPathID(configDir)
		tsConfigs, err := rslintconfig.ResolveTsConfigPaths(entries, normalizedConfigDir, fsys)
		if err != nil {
			return lintProgramSet{}, fmt.Errorf("resolve tsconfigs for %q: %w", configDir, err)
		}

		for order, tsconfigPath := range tsConfigs {
			tsconfigPath = tspath.NormalizePath(tsconfigPath)
			tsconfigID := compilerPathID(tsconfigPath)
			if programIndex, ok := programByTsconfig[tsconfigID]; ok {
				if _, alreadyAssociated := set.ConfigOrders[programIndex][configDirID]; !alreadyAssociated {
					set.ConfigOrders[programIndex][configDirID] = order
				}
				continue
			}

			// Relative paths in a tsconfig are resolved from the declared path,
			// including when that path is a file symlink. This matches tsc/tsgo;
			// realpath is only a source-identity fallback during target binding.
			programCwd := tspath.GetDirectoryPath(tsconfigPath)
			host := utils.WithParseCache(utils.CreateCompilerHost(programCwd, fsys), parseCache)
			program, err := utils.CreateProgramLenient(singleThreaded, fsys, programCwd, tsconfigPath, host)
			if err != nil {
				return lintProgramSet{}, fmt.Errorf("create TypeScript Program from %q: %w", tsconfigPath, err)
			}

			programByTsconfig[tsconfigID] = len(set.Programs)
			set.Programs = append(set.Programs, program)
			set.ConfigOrders = append(set.ConfigOrders, programConfigOrders{configDirID: order})
		}
	}

	return set, nil
}

func createProgramSetForConfig(
	configDir string,
	entries rslintconfig.RslintConfig,
	singleThreaded bool,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
) (lintProgramSet, error) {
	return createProgramSetForConfigs(
		map[string]rslintconfig.RslintConfig{configDir: entries},
		singleThreaded,
		fsys,
		parseCache,
	)
}

// parallelGitignoreAndPrograms reads gitignore state and builds the Program
// registry for an invocation-wide config (explicit JS/TS and JSON/JSONC).
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
	fsys vfs.FS,
	targetFiles []string,
	singleThreaded bool,
	parseCache *utils.ParseCache,
) (rslintconfig.RslintConfig, lintProgramSet, error) {
	var (
		configWithIgnores rslintconfig.RslintConfig
		programs          lintProgramSet
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
		programs, programErr = createProgramSetForConfig(configDir, rslintConfig, singleThreaded, fsys, parseCache)
	})
	wg.RunAndWait()

	if programErr != nil {
		return rslintConfig, lintProgramSet{}, programErr
	}
	return configWithIgnores, programs, nil
}

// createFallbackProgram creates a Program for selected lint targets not
// included in any existing Program. It uses minimal compiler options sufficient
// for AST parsing (no type checking).
type fallbackProgram struct {
	program            *compiler.Program
	sourcePathByTarget map[string]string
}

func createFallbackProgram(
	gapFiles []string,
	singleThreaded bool,
	configDir string,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
) (fallbackProgram, error) {
	compilerFiles := make([]string, 0, len(gapFiles))
	sourcePathByTarget := make(map[string]string, len(gapFiles))
	virtualFiles := make(map[string]string)
	reserved := make(map[string]struct{}, len(gapFiles)*2)
	for _, filePath := range gapFiles {
		reserved[compilerPathID(filePath)] = struct{}{}
	}
	for _, filePath := range gapFiles {
		sourcePath := filePath
		if !compilerCanRepresentHostPath(filePath) {
			contents, ok := fsys.ReadFile(filePath)
			if !ok {
				return fallbackProgram{}, fmt.Errorf("read fallback lint target %q", filePath)
			}
			sourcePath = compilerpath.Alias(filePath, fsys, reserved)
			virtualFiles[sourcePath] = contents
		}
		compilerFiles = append(compilerFiles, sourcePath)
		sourcePathByTarget[exactHostPathID(filePath)] = sourcePath
	}
	compilerFS := fsys
	if len(virtualFiles) > 0 {
		compilerFS = utils.NewOverlayVFS(fsys, virtualFiles)
	}
	host := utils.WithParseCache(utils.CreateCompilerHost(configDir, compilerFS), parseCache)
	program, err := utils.CreateProgramFromOptionsLenient(singleThreaded, &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, compilerFiles, host)
	if err != nil {
		return fallbackProgram{}, fmt.Errorf("create fallback Program for %d lint target(s): %w", len(gapFiles), err)
	}
	return fallbackProgram{program: program, sourcePathByTarget: sourcePathByTarget}, nil
}

type resolvedLintTarget struct {
	Path           string
	CanonicalPath  string
	OwnerConfigDir string
	MergedConfig   *rslintconfig.MergedConfig
}

type lintTargetPlan struct {
	Targets []resolvedLintTarget
}

// resolveDiscoveredLintTargetPlan binds Go discovery's exact lexical targets
// to physical identities without re-running config matching or ownership.
// Physical identity is only a Program-membership hint: ESLint preserves two
// distinct lexical targets even when they resolve to the same file, so this
// stage must not coalesce or reject aliases.
func resolveDiscoveredLintTargetPlan(
	targets []discovery.DiscoveredTarget,
	fsys vfs.FS,
) (lintTargetPlan, error) {
	plan := lintTargetPlan{Targets: make([]resolvedLintTarget, 0, len(targets))}
	for _, discovered := range targets {
		if discovered.MergedConfig == nil {
			return lintTargetPlan{}, fmt.Errorf("config discovery invariant: target %q has no exact merged config", discovered.Path)
		}
		canonicalPath := hostpath.Normalize(discovered.Path)
		if fsys != nil {
			if realPath := fsys.Realpath(canonicalPath); realPath != "" {
				canonicalPath = hostpath.Normalize(realPath)
			}
		}
		plan.Targets = append(plan.Targets, resolvedLintTarget{
			Path:           hostpath.Normalize(discovered.Path),
			CanonicalPath:  canonicalPath,
			OwnerConfigDir: hostpath.Normalize(discovered.ConfigDirectory),
			MergedConfig:   discovered.MergedConfig,
		})
	}
	return plan, nil
}

func configsForLintTargetPlan(
	configMap map[string]rslintconfig.RslintConfig,
	plan lintTargetPlan,
) map[string]rslintconfig.RslintConfig {
	if len(configMap) == 0 || len(plan.Targets) == 0 {
		return nil
	}
	active := make(map[string]rslintconfig.RslintConfig)
	for _, target := range plan.Targets {
		if entries, ok := configMap[target.OwnerConfigDir]; ok {
			active[target.OwnerConfigDir] = entries
		}
	}
	return active
}

func resolveLintTargetPlan(
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) (lintTargetPlan, error) {
	targetFiles := rslintconfig.DiscoverLintTargets(
		rslintConfig,
		currentDirectory,
		fsys,
		allowFiles,
		allowDirs,
		singleThreaded,
	)
	plan := lintTargetPlan{Targets: make([]resolvedLintTarget, 0, len(targetFiles))}
	for _, discovered := range targetFiles {
		targetPath := discovered.Path
		canonicalPath := hostpath.Normalize(targetPath)
		if discovered.CanonicalPath != "" {
			canonicalPath = hostpath.Normalize(discovered.CanonicalPath)
		}
		if discovered.CanonicalPath == "" && fsys != nil {
			if realPath := fsys.Realpath(canonicalPath); realPath != "" {
				canonicalPath = hostpath.Normalize(realPath)
			}
		}
		plan.Targets = append(plan.Targets, resolvedLintTarget{
			Path:           hostpath.Normalize(targetPath),
			CanonicalPath:  canonicalPath,
			OwnerConfigDir: hostpath.Normalize(currentDirectory),
		})
	}
	return plan, nil
}

func preferredCallerTargetPaths(plan lintTargetPlan) map[string]string {
	if len(plan.Targets) == 0 {
		return nil
	}
	preferred := make(map[string]string, len(plan.Targets))
	for _, target := range plan.Targets {
		canonicalID := exactHostPathID(target.CanonicalPath)
		if _, exists := preferred[canonicalID]; !exists {
			preferred[canonicalID] = target.Path
		}
	}
	return preferred
}

type lintTargetBinding struct {
	Programs              []*compiler.Program
	Views                 []lintTargetView
	SkipTypeCheckPrograms []bool
	GapFiles              []string
}

// lintTargetView is the phase-1 identity of one set of lexical targets. Views
// may share a project-backed Program, but never share config/path mappings.
// This lets two symlink aliases use distinct nearest configs without cloning
// the Program or losing its TypeChecker.
type lintTargetView struct {
	ProgramIndex               int
	TargetFiles                []string
	TypeInfoFiles              map[string]struct{}
	TargetPathBySourcePath     map[string]string
	ConfigPathBySourcePath     map[string]string
	OwnerConfigDirBySourcePath map[string]string
	MergedConfigBySourcePath   map[string]*rslintconfig.MergedConfig
}

func storeMergedConfigMapping(
	mapping map[string]*rslintconfig.MergedConfig,
	sourcePath string,
	canonicalSourcePath string,
	merged *rslintconfig.MergedConfig,
) {
	if mapping == nil || merged == nil {
		return
	}
	mapping[compilerPathID(sourcePath)] = merged
	if canonicalSourcePath != "" && compilerCanRepresentHostPath(canonicalSourcePath) {
		mapping[compilerPathID(canonicalSourcePath)] = merged
	}
}

func storeViewSourcePathMapping(
	view *lintTargetView,
	sourcePath string,
	canonicalSourcePath string,
	targetPath string,
	configPath string,
	ownerConfigDir string,
	merged *rslintconfig.MergedConfig,
) {
	if view.TargetPathBySourcePath == nil {
		view.TargetPathBySourcePath = make(map[string]string)
		view.ConfigPathBySourcePath = make(map[string]string)
		view.OwnerConfigDirBySourcePath = make(map[string]string)
		view.MergedConfigBySourcePath = make(map[string]*rslintconfig.MergedConfig)
	}
	storeSourcePathMapping(view.TargetPathBySourcePath, sourcePath, canonicalSourcePath, targetPath)
	storeSourcePathMapping(view.ConfigPathBySourcePath, sourcePath, canonicalSourcePath, configPath)
	storeSourcePathMapping(view.OwnerConfigDirBySourcePath, sourcePath, canonicalSourcePath, ownerConfigDir)
	storeMergedConfigMapping(view.MergedConfigBySourcePath, sourcePath, canonicalSourcePath, merged)
}

func exactProgramSourceFile(program *compiler.Program, targetPath string) *ast.SourceFile {
	if program == nil || targetPath == "" {
		return nil
	}
	if !compilerCanRepresentHostPath(targetPath) {
		return nil
	}
	targetPath = tspath.NormalizePath(targetPath)
	sourceFile := program.GetSourceFile(targetPath)
	if sourceFile == nil || compilerPathID(sourceFile.FileName()) != compilerPathID(targetPath) {
		return nil
	}
	return sourceFile
}

// programFileIndex joins lint targets to Program sources by exact physical
// path. It resolves each unique source path once and retains Program membership
// so project declaration order remains authoritative during lookup.
type programFileIndex struct {
	programs         []*compiler.Program
	fsys             vfs.FS
	singleThreaded   bool
	built            bool
	sourcesByProgram []map[string]*ast.SourceFile
}

func newProgramFileIndex(programs []*compiler.Program, fsys vfs.FS, singleThreaded bool) *programFileIndex {
	return &programFileIndex{
		programs:       programs,
		fsys:           fsys,
		singleThreaded: singleThreaded,
	}
}

func (index *programFileIndex) sourceFile(programIndex int, canonicalTarget string) *ast.SourceFile {
	if index == nil || index.fsys == nil || canonicalTarget == "" ||
		programIndex < 0 || programIndex >= len(index.programs) {
		return nil
	}
	if !compilerCanRepresentHostPath(canonicalTarget) {
		return nil
	}
	if !index.built {
		index.build()
	}
	return index.sourcesByProgram[programIndex][compilerPathID(canonicalTarget)]
}

type programSourceMembership struct {
	programIndex int
	sourceIndex  int
	sourceFile   *ast.SourceFile
}

func (index *programFileIndex) build() {
	index.built = true
	index.sourcesByProgram = make([]map[string]*ast.SourceFile, len(index.programs))

	sourceIndexByPath := make(map[string]int)
	var sourcePaths []string
	var memberships []programSourceMembership
	for programIndex, program := range index.programs {
		if program == nil {
			continue
		}
		for _, sourceFile := range program.GetSourceFiles() {
			sourcePath := tspath.NormalizePath(sourceFile.FileName())
			sourcePathID := compilerPathID(sourcePath)
			sourceIndex, exists := sourceIndexByPath[sourcePathID]
			if !exists {
				sourceIndex = len(sourcePaths)
				sourceIndexByPath[sourcePathID] = sourceIndex
				sourcePaths = append(sourcePaths, sourcePath)
			}
			memberships = append(memberships, programSourceMembership{
				programIndex: programIndex,
				sourceIndex:  sourceIndex,
				sourceFile:   sourceFile,
			})
		}
	}

	canonicalIDs := make([]string, len(sourcePaths))
	work := core.NewWorkGroup(index.singleThreaded)
	for i := range sourcePaths {
		work.Queue(func() {
			canonicalIDs[i] = compilerPathID(authoritativeHostPath(sourcePaths[i], index.fsys))
		})
	}
	work.RunAndWait()

	for _, membership := range memberships {
		sources := index.sourcesByProgram[membership.programIndex]
		if sources == nil {
			sources = make(map[string]*ast.SourceFile)
			index.sourcesByProgram[membership.programIndex] = sources
		}
		canonicalID := canonicalIDs[membership.sourceIndex]
		existing := sources[canonicalID]
		if existing == nil || membership.sourceFile.FileName() < existing.FileName() {
			sources[canonicalID] = membership.sourceFile
		}
	}
}

func groupFallbackTargets(
	gaps []resolvedLintTarget,
	currentDirectory string,
	useCaseSensitive bool,
) [][]resolvedLintTarget {
	if len(gaps) == 0 {
		return nil
	}

	groups := make([][]resolvedLintTarget, 0, 1)
	keysByGroup := make([]map[string]struct{}, 0, 1)
	canonicalKeysByGroup := make([]map[string]struct{}, 0, 1)
	for _, gap := range gaps {
		key := exactHostPathID(gap.Path)
		canonicalKey := exactHostPathID(gap.CanonicalPath)
		if !useCaseSensitive {
			key = strings.ToLower(key)
			canonicalKey = strings.ToLower(canonicalKey)
		}
		groupIndex := -1
		for i, keys := range keysByGroup {
			if _, exists := keys[key]; exists {
				continue
			}
			if _, exists := canonicalKeysByGroup[i][canonicalKey]; exists {
				continue
			}
			groupIndex = i
			break
		}
		if groupIndex == -1 {
			groupIndex = len(groups)
			groups = append(groups, nil)
			keysByGroup = append(keysByGroup, make(map[string]struct{}))
			canonicalKeysByGroup = append(canonicalKeysByGroup, make(map[string]struct{}))
		}
		groups[groupIndex] = append(groups[groupIndex], gap)
		keysByGroup[groupIndex][key] = struct{}{}
		canonicalKeysByGroup[groupIndex][canonicalKey] = struct{}{}
	}
	return groups
}

func orderedProgramIndexesForConfig(set lintProgramSet, configDir string) []int {
	configDirID := exactHostPathID(configDir)
	indexes := make([]int, 0, len(set.Programs))
	for i := range set.Programs {
		if i < len(set.ConfigOrders) {
			if _, ok := set.ConfigOrders[i][configDirID]; ok {
				indexes = append(indexes, i)
			}
		}
	}
	sort.SliceStable(indexes, func(i, j int) bool {
		left := set.ConfigOrders[indexes[i]][configDirID]
		right := set.ConfigOrders[indexes[j]][configDirID]
		if left != right {
			return left < right
		}
		return indexes[i] < indexes[j]
	})
	return indexes
}

// bindLintTargetPlan binds every stable target to a Program from its governing
// config. Calling this for each fix pass recomputes gap status from the current
// import graph instead of retaining an initial gap classification.
func bindLintTargetPlan(
	set lintProgramSet,
	plan lintTargetPlan,
	currentDirectory string,
	fsys vfs.FS,
	parseCache *utils.ParseCache,
	singleThreaded bool,
) (lintTargetBinding, error) {
	binding := lintTargetBinding{
		Programs: append([]*compiler.Program(nil), set.Programs...),
		Views:    make([]lintTargetView, len(set.Programs)),
	}
	for i := range binding.Views {
		binding.Views[i].ProgramIndex = i
	}

	var gaps []resolvedLintTarget
	programIndexesByConfig := make(map[string][]int)
	programFiles := newProgramFileIndex(set.Programs, fsys, singleThreaded)
	usedSourcesByProgram := make([]map[string]struct{}, len(set.Programs))
	for _, target := range plan.Targets {
		programIndexes, cached := programIndexesByConfig[target.OwnerConfigDir]
		if !cached {
			programIndexes = orderedProgramIndexesForConfig(set, target.OwnerConfigDir)
			programIndexesByConfig[target.OwnerConfigDir] = programIndexes
		}
		bound := false
		for _, programIndex := range programIndexes {
			sourceFile := exactProgramSourceFile(set.Programs[programIndex], target.Path)
			if sourceFile == nil {
				sourceFile = programFiles.sourceFile(programIndex, target.CanonicalPath)
			}
			if sourceFile == nil {
				continue
			}
			sourcePath := sourceFile.FileName()
			viewIndex := programIndex
			sourceID := compilerPathID(sourcePath)
			if usedSourcesByProgram[programIndex] == nil {
				usedSourcesByProgram[programIndex] = make(map[string]struct{})
			}
			if _, occupied := usedSourcesByProgram[programIndex][sourceID]; occupied {
				// One Program exposes one SourceFile identity, while ESLint keeps
				// every lexical input/config identity. Reuse the project-backed
				// Program as a separate lint-only view instead of degrading the
				// alias to a no-type-info fallback Program.
				viewIndex = len(binding.Views)
				binding.Views = append(binding.Views, lintTargetView{ProgramIndex: programIndex})
			} else {
				usedSourcesByProgram[programIndex][sourceID] = struct{}{}
			}
			binding.Views[viewIndex].TargetFiles = append(binding.Views[viewIndex].TargetFiles, sourcePath)
			storeViewSourcePathMapping(
				&binding.Views[viewIndex],
				sourcePath,
				target.CanonicalPath,
				target.Path,
				configPathForLintTarget(target, fsys),
				target.OwnerConfigDir,
				target.MergedConfig,
			)
			bound = true
			break
		}
		if !bound {
			gaps = append(gaps, target)
			binding.GapFiles = append(binding.GapFiles, target.Path)
		}
	}

	if len(gaps) > 0 {
		useCaseSensitive := true
		if fsys != nil {
			useCaseSensitive = fsys.UseCaseSensitiveFileNames()
		}
		for _, fallbackTargets := range groupFallbackTargets(gaps, currentDirectory, useCaseSensitive) {
			fallbackFiles := make([]string, 0, len(fallbackTargets))
			for _, gap := range fallbackTargets {
				fallbackFiles = append(fallbackFiles, gap.Path)
			}
			fallbackResult, err := createFallbackProgram(fallbackFiles, singleThreaded, currentDirectory, fsys, parseCache)
			if err != nil {
				return lintTargetBinding{}, err
			}
			fallback := fallbackResult.program
			if fallback == nil {
				return lintTargetBinding{}, fmt.Errorf("create fallback Program for %d lint target(s): no Program returned", len(fallbackTargets))
			}
			fallbackIndex := len(binding.Programs)
			binding.Programs = append(binding.Programs, fallback)
			viewIndex := len(binding.Views)
			// An explicitly empty set means no target in this synthesized
			// Program has reliable project type information.
			binding.Views = append(binding.Views, lintTargetView{
				ProgramIndex:  fallbackIndex,
				TypeInfoFiles: map[string]struct{}{},
			})
			for _, gap := range fallbackTargets {
				sourcePath := fallbackResult.sourcePathByTarget[exactHostPathID(gap.Path)]
				sourceFile := exactProgramSourceFile(fallback, sourcePath)
				if sourceFile == nil {
					return lintTargetBinding{}, fmt.Errorf("fallback Program did not contain lint target %q", gap.Path)
				}
				sourcePath = sourceFile.FileName()
				binding.Views[viewIndex].TargetFiles = append(binding.Views[viewIndex].TargetFiles, sourcePath)
				storeViewSourcePathMapping(
					&binding.Views[viewIndex],
					sourcePath,
					gap.CanonicalPath,
					gap.Path,
					configPathForLintTarget(gap, fsys),
					gap.OwnerConfigDir,
					gap.MergedConfig,
				)
			}
		}
	}

	for i := range binding.Views {
		sort.Strings(binding.Views[i].TargetFiles)
	}
	binding.SkipTypeCheckPrograms = buildTypeCheckSkipMask(binding.Programs)
	return binding, nil
}

// buildTypeCheckSkipMask returns a parallel-to-programs []bool marking which
// programs must be excluded from the type-check phase. A program is skipped
// when it was NOT built from a real tsconfig on disk — i.e. its CompilerOptions
// carry no ConfigFilePath. That covers the non-project-backed fallback Program used for
// selected files absent from every tsconfig-backed Program, including projects
// with no tsconfig at all.
//
// Their options are synthesized, not the user's tsconfig, so semantic
// diagnostics there are unreliable and must not be surfaced. This mirrors the
// type-checking boundary: only Programs backed by parserOptions.project or the
// auto-detected tsconfig.json participate in program-level --type-check
// diagnostics. Programs built via utils.CreateProgram ->
// GetParsedCommandLineOfConfigFile carry a non-empty ConfigFilePath.
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

func collectTargetViewSyntacticDiagnostics(
	binding lintTargetBinding,
	typeCheck bool,
	typeCheckOnly bool,
) ([]rule.RuleDiagnostic, map[string]struct{}) {
	syntaxErrorFiles := make(map[string]struct{})
	seen := make(map[syntacticDiagnosticKey]struct{})
	var diagnostics []rule.RuleDiagnostic
	for _, view := range binding.Views {
		if view.ProgramIndex < 0 || view.ProgramIndex >= len(binding.Programs) {
			continue
		}
		program := binding.Programs[view.ProgramIndex]
		coveredByTypeCheck := typeCheck && (view.ProgramIndex >= len(binding.SkipTypeCheckPrograms) || !binding.SkipTypeCheckPrograms[view.ProgramIndex])
		ctx := context.Background()
		for _, target := range view.TargetFiles {
			file := program.GetSourceFile(target)
			if file == nil {
				continue
			}
			for _, diagnostic := range program.GetSyntacticDiagnostics(ctx, file) {
				syntaxErrorFiles[file.FileName()] = struct{}{}
				if coveredByTypeCheck || typeCheckOnly {
					continue
				}
				filePath := file.FileName()
				if targetPath := view.TargetPathBySourcePath[compilerPathID(filePath)]; targetPath != "" {
					filePath = targetPath
				}
				loc := diagnostic.Loc()
				key := syntacticDiagnosticKey{path: filePath, code: diagnostic.Code(), pos: loc.Pos(), end: loc.End()}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				diagnostics = append(diagnostics, rule.RuleDiagnostic{
					RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
					SourceFile:   file,
					FilePath:     filePath,
					Range:        loc,
					Message:      rule.RuleMessage{Description: diagnostic.String()},
					Severity:     rule.SeverityError,
					Origin:       rule.DiagnosticOriginTypeScript,
					PreFormatted: true,
				})
			}
		}
	}
	return diagnostics, syntaxErrorFiles
}
