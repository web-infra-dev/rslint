package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/module"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
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
) (lintProgramSet, error) {
	if len(plan.specs) == 0 {
		if plan.terminalErr != nil {
			return lintProgramSet{}, plan.terminalErr
		}
		return lintProgramSet{}, nil
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
			return lintProgramSet{}, fmt.Errorf(
				"create TypeScript Program from %q: %w",
				plan.specs[index].tsconfigPath,
				err,
			)
		}
	}
	if plan.terminalErr != nil {
		return lintProgramSet{}, plan.terminalErr
	}

	configOrders := make([]programConfigOrders, len(plan.specs))
	for index := range plan.specs {
		configOrders[index] = plan.specs[index].configOrders
	}
	return lintProgramSet{Programs: programs, ConfigOrders: configOrders}, nil
}

// createProgramSetForConfigs builds each planned Program into its stable
// registry slot. At most min(GOMAXPROCS, Program count) Programs are built
// concurrently; --singleThreaded retains the original serial, fail-fast path
// exactly.
func createProgramSetForConfigs(
	configMap map[string]rslintconfig.RslintConfig,
	singleThreaded bool,
	buildContext *utils.ProgramBuildContext,
) (lintProgramSet, error) {
	plan := buildProgramPlan(configMap, buildContext.FS())
	return executeProgramBuildPlan(plan, singleThreaded, buildContext)
}

func createProgramSetForConfig(
	configDir string,
	entries rslintconfig.RslintConfig,
	singleThreaded bool,
	buildContext *utils.ProgramBuildContext,
) (lintProgramSet, error) {
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
) (rslintconfig.RslintConfig, lintProgramSet, error) {
	fsys := buildContext.FS()
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
		programs, programErr = createProgramSetForConfig(configDir, rslintConfig, singleThreaded, buildContext)
	})
	wg.RunAndWait()

	if programErr != nil {
		return rslintConfig, lintProgramSet{}, programErr
	}
	return configWithIgnores, programs, nil
}

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

// createFallbackProgram creates a Program for selected lint targets not
// included in any existing Program. It uses minimal compiler options sufficient
// for AST parsing (no type checking).
func createFallbackProgram(
	gapFiles []string,
	singleThreaded bool,
	configDir string,
	buildContext *utils.ProgramBuildContext,
) (*compiler.Program, error) {
	program, err := buildContext.CreateProgramFromOptionsLenient(
		singleThreaded,
		configDir,
		fallbackCompilerOptions(),
		gapFiles,
	)
	if err != nil {
		return nil, fmt.Errorf("create fallback Program for %d lint target(s): %w", len(gapFiles), err)
	}
	return program, nil
}

type resolvedLintTarget struct {
	Path           string
	CanonicalPath  string
	OwnerConfigDir string
}

type lintTargetPlan struct {
	Targets []resolvedLintTarget
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
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	configTargetScopes map[string]rslintconfig.LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) (lintTargetPlan, error) {
	type targetWithOwner struct {
		path          string
		canonicalPath string
		owner         string
	}
	var targetFiles []targetWithOwner
	if configMap != nil {
		for _, target := range discoverLintFilesMultiConfig(configMap, configTargetScopes, fsys, allowFiles, allowDirs, singleThreaded) {
			targetFiles = append(targetFiles, targetWithOwner{
				path:          target.Path,
				canonicalPath: target.CanonicalPath,
				owner:         target.ConfigDirectory,
			})
		}
	} else {
		for _, target := range rslintconfig.DiscoverLintTargets(rslintConfig, currentDirectory, fsys, allowFiles, allowDirs, singleThreaded) {
			targetFiles = append(targetFiles, targetWithOwner{
				path:          target.Path,
				canonicalPath: target.CanonicalPath,
				owner:         currentDirectory,
			})
		}
	}

	plan := lintTargetPlan{Targets: make([]resolvedLintTarget, 0, len(targetFiles))}
	seenCanonical := make(map[string]resolvedLintTarget, len(targetFiles))
	for _, discovered := range targetFiles {
		targetPath := discovered.path
		ownerConfigDir := discovered.owner
		canonicalPath := tspath.NormalizePath(targetPath)
		if discovered.canonicalPath != "" {
			canonicalPath = tspath.NormalizePath(discovered.canonicalPath)
		}
		if discovered.canonicalPath == "" && fsys != nil {
			if realPath := fsys.Realpath(canonicalPath); realPath != "" {
				canonicalPath = tspath.NormalizePath(realPath)
			}
		}
		canonicalKey := exactFilesystemPathID(canonicalPath)
		target := resolvedLintTarget{
			Path:           tspath.NormalizePath(targetPath),
			CanonicalPath:  canonicalPath,
			OwnerConfigDir: tspath.NormalizePath(ownerConfigDir),
		}
		if existing, exists := seenCanonical[canonicalKey]; exists {
			if canonicalFilesystemPathID(existing.OwnerConfigDir, fsys) != canonicalFilesystemPathID(target.OwnerConfigDir, fsys) {
				return lintTargetPlan{}, fmt.Errorf(
					"lint target aliases %q and %q resolve to the same file but are governed by different configs %q and %q",
					existing.Path,
					target.Path,
					existing.OwnerConfigDir,
					target.OwnerConfigDir,
				)
			}
			continue
		}
		seenCanonical[canonicalKey] = target
		plan.Targets = append(plan.Targets, target)
	}
	return plan, nil
}

func preferredCallerTargetPaths(plan lintTargetPlan) map[string]string {
	if len(plan.Targets) == 0 {
		return nil
	}
	preferred := make(map[string]string, len(plan.Targets))
	for _, target := range plan.Targets {
		canonicalID := exactFilesystemPathID(target.CanonicalPath)
		if _, exists := preferred[canonicalID]; !exists {
			preferred[canonicalID] = target.Path
		}
	}
	return preferred
}

type lintTargetBinding struct {
	Programs                   []*compiler.Program
	TypeInfoFiles              map[string]struct{}
	GapFiles                   []string
	TargetsByProgram           [][]string
	TargetPathBySourcePath     map[string]string
	ConfigPathBySourcePath     map[string]string
	OwnerConfigDirBySourcePath map[string]string
	standaloneGapGroups        [][]resolvedLintTarget
}

func exactProgramSourceFile(program *compiler.Program, targetPath string) *ast.SourceFile {
	if program == nil || targetPath == "" {
		return nil
	}
	targetPath = tspath.NormalizePath(targetPath)
	sourceFile := program.GetSourceFile(targetPath)
	if sourceFile == nil || exactFilesystemPathID(sourceFile.FileName()) != exactFilesystemPathID(targetPath) {
		return nil
	}
	return sourceFile
}

// programFileIndex joins lint targets to Program sources by exact physical
// path. It is scoped to one binding pass and builds the governing config's
// Program indexes only after one of them misses an exact lexical lookup.
// Canonical target identities established during discovery are reused before
// consulting the filesystem, and unknown source paths are resolved at most once
// across all Programs in the pass.
type programFileIndex struct {
	programs              []*compiler.Program
	targets               []resolvedLintTarget
	fsys                  vfs.FS
	singleThreaded        bool
	initialized           bool
	builtByProgram        []bool
	sourcesByProgram      []map[string]*ast.SourceFile
	targetCanonicalIDs    map[string]struct{}
	canonicalBySourcePath map[string]string
	directoryIdentities   map[string]programDirectoryIdentity
}

func newProgramFileIndex(
	programs []*compiler.Program,
	targets []resolvedLintTarget,
	fsys vfs.FS,
	singleThreaded bool,
) *programFileIndex {
	return &programFileIndex{
		programs:       programs,
		targets:        targets,
		fsys:           fsys,
		singleThreaded: singleThreaded,
	}
}

func (index *programFileIndex) initialize() {
	if index.initialized {
		return
	}
	index.initialized = true
	index.builtByProgram = make([]bool, len(index.programs))
	index.sourcesByProgram = make([]map[string]*ast.SourceFile, len(index.programs))
	index.targetCanonicalIDs = make(map[string]struct{}, len(index.targets))
	index.canonicalBySourcePath = make(map[string]string, len(index.targets)*2)
	// Seed physical identities first so a lexical hint cannot overwrite a path
	// that is itself the canonical identity of another target.
	for _, target := range index.targets {
		if target.CanonicalPath == "" {
			continue
		}
		canonicalID := exactFilesystemPathID(target.CanonicalPath)
		index.targetCanonicalIDs[canonicalID] = struct{}{}
		index.canonicalBySourcePath[canonicalID] = canonicalID
	}
	for _, target := range index.targets {
		if target.Path == "" || target.CanonicalPath == "" {
			continue
		}
		sourcePathID := exactFilesystemPathID(target.Path)
		if _, isCanonicalTarget := index.targetCanonicalIDs[sourcePathID]; !isCanonicalTarget {
			index.canonicalBySourcePath[sourcePathID] = exactFilesystemPathID(target.CanonicalPath)
		}
	}
	index.targets = nil
}

func (index *programFileIndex) sourceFile(
	programIndexes []int,
	programIndex int,
	canonicalTarget string,
) *ast.SourceFile {
	if index == nil || index.fsys == nil || canonicalTarget == "" ||
		programIndex < 0 || programIndex >= len(index.programs) {
		return nil
	}
	index.initialize()
	if !index.builtByProgram[programIndex] {
		index.buildPrograms(programIndexes)
	}
	return index.sourcesByProgram[programIndex][exactFilesystemPathID(canonicalTarget)]
}

type programSourceMembership struct {
	programIndex int
	sourceIndex  int
	canonicalID  string
	sourceFile   *ast.SourceFile
}

type programDirectoryIdentity struct {
	canonicalPath string
	entries       vfs.Entries
}

func regularFileNameFromEntries(entries vfs.Entries, fileName string, useCaseSensitive bool) (string, bool) {
	if entries.Symlinks == nil {
		return "", false
	}
	canonicalFileName := tspath.GetCanonicalFileName(fileName, useCaseSensitive)
	match := ""
	for _, entryName := range entries.Files {
		if tspath.GetCanonicalFileName(entryName, useCaseSensitive) != canonicalFileName {
			continue
		}
		if match != "" && match != entryName {
			return "", false
		}
		match = entryName
	}
	if match == "" {
		return "", false
	}
	if _, isSymlink := entries.Symlinks[match]; isSymlink {
		return "", false
	}
	// GetAccessibleEntries supplies the filesystem's actual casing. Once its
	// complete metadata proves this entry is regular, realpath(file) is exactly
	// realpath(parent) joined with this name.
	return match, true
}

func (index *programFileIndex) canonicalSourcePathIDs(sourcePaths []string) []string {
	canonicalIDs := make([]string, len(sourcePaths))
	if len(sourcePaths) == 0 {
		return canonicalIDs
	}
	if index.directoryIdentities == nil {
		index.directoryIdentities = make(map[string]programDirectoryIdentity)
	}

	sourceIndexesByDirectoryID := make(map[string][]int)
	directoryPathByID := make(map[string]string)
	for i, sourcePath := range sourcePaths {
		directoryPath := tspath.GetDirectoryPath(sourcePath)
		directoryID := exactFilesystemPathID(directoryPath)
		directoryPathByID[directoryID] = directoryPath
		sourceIndexesByDirectoryID[directoryID] = append(sourceIndexesByDirectoryID[directoryID], i)
	}

	directoryIDs := make([]string, 0, len(sourceIndexesByDirectoryID))
	for directoryID := range sourceIndexesByDirectoryID {
		directoryIDs = append(directoryIDs, directoryID)
	}
	sort.Strings(directoryIDs)
	pendingIdentities := make([]programDirectoryIdentity, len(directoryIDs))
	hasPendingIdentity := make([]bool, len(directoryIDs))
	useCaseSensitive := index.fsys.UseCaseSensitiveFileNames()
	work := core.NewWorkGroup(index.singleThreaded)
	queueSource := func(sourceIndex int, directory programDirectoryIdentity) {
		work.Queue(func() {
			sourcePath := sourcePaths[sourceIndex]
			fileName, regular := regularFileNameFromEntries(
				directory.entries,
				tspath.GetBaseFileName(sourcePath),
				useCaseSensitive,
			)
			if regular {
				canonicalIDs[sourceIndex] = exactFilesystemPathID(
					tspath.CombinePaths(directory.canonicalPath, fileName),
				)
			} else {
				canonicalIDs[sourceIndex] = canonicalFilesystemPathID(sourcePath, index.fsys)
			}
		})
	}
	for directoryIndex, directoryID := range directoryIDs {
		sourceIndexes := sourceIndexesByDirectoryID[directoryID]
		directory, cached := index.directoryIdentities[directoryID]
		if len(sourceIndexes) == 1 && !cached {
			sourceIndex := sourceIndexes[0]
			work.Queue(func() {
				canonicalIDs[sourceIndex] = canonicalFilesystemPathID(sourcePaths[sourceIndex], index.fsys)
			})
			continue
		}
		if cached {
			for _, sourceIndex := range sourceIndexes {
				queueSource(sourceIndex, directory)
			}
		} else {
			work.Queue(func() {
				directoryPath := directoryPathByID[directoryID]
				directory = programDirectoryIdentity{
					canonicalPath: authoritativeFilesystemPath(directoryPath, index.fsys),
					entries:       index.fsys.GetAccessibleEntries(directoryPath),
				}
				pendingIdentities[directoryIndex] = directory
				hasPendingIdentity[directoryIndex] = true
				// Queue per-file work before the directory task returns. A
				// WorkGroup permits nested Queue calls until RunAndWait has
				// returned, retaining file-level parallelism without a second
				// directory-to-file barrier.
				for _, sourceIndex := range sourceIndexes {
					queueSource(sourceIndex, directory)
				}
			})
		}
	}
	work.RunAndWait()
	for i, resolved := range hasPendingIdentity {
		if resolved {
			index.directoryIdentities[directoryIDs[i]] = pendingIdentities[i]
		}
	}
	return canonicalIDs
}

func (index *programFileIndex) buildPrograms(programIndexes []int) {
	sourceIndexByPath := make(map[string]int)
	var sourcePathIDs []string
	var sourcePaths []string
	var memberships []programSourceMembership

	for _, programIndex := range programIndexes {
		if programIndex < 0 || programIndex >= len(index.programs) ||
			index.builtByProgram[programIndex] {
			continue
		}
		index.builtByProgram[programIndex] = true

		program := index.programs[programIndex]
		if program == nil {
			continue
		}
		for _, sourceFile := range program.GetSourceFiles() {
			sourcePath := tspath.NormalizePath(sourceFile.FileName())
			sourcePathID := exactFilesystemPathID(sourcePath)
			canonicalID, known := index.canonicalBySourcePath[sourcePathID]
			sourceIndex := -1
			if !known {
				var pending bool
				sourceIndex, pending = sourceIndexByPath[sourcePathID]
				if !pending {
					sourceIndex = len(sourcePaths)
					sourceIndexByPath[sourcePathID] = sourceIndex
					sourcePathIDs = append(sourcePathIDs, sourcePathID)
					sourcePaths = append(sourcePaths, sourcePath)
				}
			}
			memberships = append(memberships, programSourceMembership{
				programIndex: programIndex,
				sourceIndex:  sourceIndex,
				canonicalID:  canonicalID,
				sourceFile:   sourceFile,
			})
		}
	}

	canonicalIDs := index.canonicalSourcePathIDs(sourcePaths)
	for i, sourcePathID := range sourcePathIDs {
		index.canonicalBySourcePath[sourcePathID] = canonicalIDs[i]
	}

	for _, membership := range memberships {
		canonicalID := membership.canonicalID
		if membership.sourceIndex >= 0 {
			canonicalID = canonicalIDs[membership.sourceIndex]
		}
		if _, isTarget := index.targetCanonicalIDs[canonicalID]; !isTarget {
			continue
		}
		sources := index.sourcesByProgram[membership.programIndex]
		if sources == nil {
			sources = make(map[string]*ast.SourceFile)
			index.sourcesByProgram[membership.programIndex] = sources
		}
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
	keysByGroup := make([]map[tspath.Path]struct{}, 0, 1)
	for _, gap := range gaps {
		key := tspath.ToPath(gap.Path, currentDirectory, useCaseSensitive)
		groupIndex := -1
		for i, keys := range keysByGroup {
			if _, exists := keys[key]; !exists {
				groupIndex = i
				break
			}
		}
		if groupIndex == -1 {
			groupIndex = len(groups)
			groups = append(groups, nil)
			keysByGroup = append(keysByGroup, make(map[tspath.Path]struct{}))
		}
		groups[groupIndex] = append(groups[groupIndex], gap)
		keysByGroup[groupIndex][key] = struct{}{}
	}
	return groups
}

func orderedProgramIndexesForConfig(set lintProgramSet, configDir string) []int {
	configDirID := exactFilesystemPathID(configDir)
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

// bindLintTargetsToRealPrograms resolves every stable target against the real
// Programs declared by its governing config. Gap mappings are published before
// a fallback policy is chosen so config resolution sees the same lexical,
// canonical, and owner identities in either path.
func bindLintTargetsToRealPrograms(
	set lintProgramSet,
	plan lintTargetPlan,
	buildContext *utils.ProgramBuildContext,
	singleThreaded bool,
) (lintTargetBinding, []resolvedLintTarget) {
	fsys := buildContext.FS()
	binding := lintTargetBinding{
		Programs:                   append([]*compiler.Program(nil), set.Programs...),
		TargetsByProgram:           make([][]string, len(set.Programs)),
		TypeInfoFiles:              make(map[string]struct{}),
		TargetPathBySourcePath:     make(map[string]string),
		ConfigPathBySourcePath:     make(map[string]string),
		OwnerConfigDirBySourcePath: make(map[string]string),
	}

	var gaps []resolvedLintTarget
	programIndexesByConfig := make(map[string][]int)
	programFiles := newProgramFileIndex(set.Programs, plan.Targets, fsys, singleThreaded)
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
				sourceFile = programFiles.sourceFile(programIndexes, programIndex, target.CanonicalPath)
			}
			if sourceFile == nil {
				continue
			}
			sourcePath := sourceFile.FileName()
			binding.TargetsByProgram[programIndex] = append(binding.TargetsByProgram[programIndex], sourcePath)
			storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, sourcePath, target.CanonicalPath, target.OwnerConfigDir)
			storeSourcePathMapping(binding.ConfigPathBySourcePath, sourcePath, target.CanonicalPath, configPathForLintTarget(target, fsys))
			binding.TypeInfoFiles[sourcePath] = struct{}{}
			binding.TypeInfoFiles[target.Path] = struct{}{}
			binding.TypeInfoFiles[target.CanonicalPath] = struct{}{}
			if tspath.NormalizePath(sourcePath) != target.Path {
				storeSourcePathMapping(binding.TargetPathBySourcePath, sourcePath, target.CanonicalPath, target.Path)
			}
			bound = true
			break
		}
		if !bound {
			gaps = append(gaps, target)
			binding.GapFiles = append(binding.GapFiles, target.Path)
			storeSourcePathMapping(
				binding.OwnerConfigDirBySourcePath,
				target.Path,
				target.CanonicalPath,
				target.OwnerConfigDir,
			)
			storeSourcePathMapping(
				binding.ConfigPathBySourcePath,
				target.Path,
				target.CanonicalPath,
				configPathForLintTarget(target, fsys),
			)
		}
	}
	return binding, gaps
}

func appendFallbackPrograms(
	binding *lintTargetBinding,
	gaps []resolvedLintTarget,
	currentDirectory string,
	buildContext *utils.ProgramBuildContext,
	singleThreaded bool,
) error {
	if len(gaps) == 0 {
		return nil
	}
	fsys := buildContext.FS()
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	for _, fallbackTargets := range groupFallbackTargets(gaps, currentDirectory, useCaseSensitive) {
		fallbackFiles := make([]string, 0, len(fallbackTargets))
		for _, gap := range fallbackTargets {
			fallbackFiles = append(fallbackFiles, gap.Path)
		}
		fallback, err := createFallbackProgram(fallbackFiles, singleThreaded, currentDirectory, buildContext)
		if err != nil {
			return err
		}
		if fallback == nil {
			return fmt.Errorf("create fallback Program for %d lint target(s): no Program returned", len(fallbackTargets))
		}
		fallbackIndex := len(binding.Programs)
		binding.Programs = append(binding.Programs, fallback)
		binding.TargetsByProgram = append(binding.TargetsByProgram, nil)
		for _, gap := range fallbackTargets {
			sourceFile := exactProgramSourceFile(fallback, gap.Path)
			if sourceFile == nil {
				return fmt.Errorf("fallback Program did not contain lint target %q", gap.Path)
			}
			sourcePath := sourceFile.FileName()
			binding.TargetsByProgram[fallbackIndex] = append(binding.TargetsByProgram[fallbackIndex], sourcePath)
			if tspath.NormalizePath(sourcePath) != gap.Path {
				storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, sourcePath, gap.CanonicalPath, gap.OwnerConfigDir)
				storeSourcePathMapping(
					binding.ConfigPathBySourcePath,
					sourcePath,
					gap.CanonicalPath,
					binding.ConfigPathBySourcePath[tspath.NormalizePath(gap.Path)],
				)
				storeSourcePathMapping(binding.TargetPathBySourcePath, sourcePath, gap.CanonicalPath, gap.Path)
			}
		}
	}
	return nil
}

func finalizeLintTargetBinding(binding *lintTargetBinding) {
	for i := range binding.TargetsByProgram {
		sort.Strings(binding.TargetsByProgram[i])
	}
	if len(binding.GapFiles) == 0 {
		binding.TypeInfoFiles = nil
	}
	if len(binding.TargetPathBySourcePath) == 0 {
		binding.TargetPathBySourcePath = nil
	}
}

// bindLintTargetPlan binds every stable target to a Program from its governing
// config. API callers retain this conservative path. The CLI may defer the
// fallback decision until it knows whether any gap has an executable rule.
// Calling either path for each fix pass recomputes gap status from the current
// import graph instead of retaining an initial gap classification.
func bindLintTargetPlan(
	set lintProgramSet,
	plan lintTargetPlan,
	currentDirectory string,
	buildContext *utils.ProgramBuildContext,
	singleThreaded bool,
) (lintTargetBinding, error) {
	binding, gaps := bindLintTargetsToRealPrograms(set, plan, buildContext, singleThreaded)
	if err := appendFallbackPrograms(&binding, gaps, currentDirectory, buildContext, singleThreaded); err != nil {
		return lintTargetBinding{}, err
	}
	finalizeLintTargetBinding(&binding)
	return binding, nil
}

func anyGapHasActiveRules(
	gaps []resolvedLintTarget,
	resolver *lintConfigResolver,
	singleThreaded bool,
) bool {
	if len(gaps) == 0 || resolver == nil {
		return false
	}
	hasRules := func(gap resolvedLintTarget) bool {
		return len(resolver.ActiveRulesForFile(gap.Path)) > 0
	}
	workerCount := min(runtime.GOMAXPROCS(0), len(gaps))
	if singleThreaded || workerCount < 2 {
		for _, gap := range gaps {
			if hasRules(gap) {
				return true
			}
		}
		return false
	}

	var found atomic.Bool
	chunkSize := (len(gaps) + workerCount - 1) / workerCount
	work := core.NewWorkGroup(false)
	for worker := range workerCount {
		start := worker * chunkSize
		end := min(start+chunkSize, len(gaps))
		if start >= end {
			continue
		}
		chunk := gaps[start:end]
		work.Queue(func() {
			for _, gap := range chunk {
				if found.Load() {
					return
				}
				if hasRules(gap) {
					found.Store(true)
					return
				}
			}
		})
	}
	work.RunAndWait()
	return found.Load()
}

func allGapRootsSupportedByFallback(gaps []resolvedLintTarget, useCaseSensitive bool) bool {
	options := fallbackCompilerOptions()
	supportedExtensions := tsoptions.GetSupportedExtensionsWithJsonIfResolveJsonModule(
		options,
		tspath.AllSupportedExtensions,
	)
	for _, gap := range gaps {
		if !tspath.HasExtension(gap.Path) {
			return false
		}
		fileName := tspath.GetCanonicalFileName(gap.Path, useCaseSensitive)
		supported := false
		for _, extensions := range supportedExtensions {
			if tspath.FileExtensionIsOneOf(fileName, extensions) {
				supported = true
				break
			}
		}
		if !supported {
			return false
		}
	}
	return true
}

// bindCLILintTargetPlan keeps the API's Program-backed contract separate from
// the CLI syntax-only fast path. A single active rule on any gap retains the
// complete legacy fallback set so cross-file and Program-dependent rules see
// exactly the same roots as before.
func bindCLILintTargetPlan(
	set lintProgramSet,
	plan lintTargetPlan,
	currentDirectory string,
	buildContext *utils.ProgramBuildContext,
	singleThreaded bool,
	resolverOptions lintConfigResolverOptions,
) (lintTargetBinding, *lintConfigResolver, error) {
	binding, gaps := bindLintTargetsToRealPrograms(set, plan, buildContext, singleThreaded)
	resolverOptions.TypeInfoFiles = binding.TypeInfoFiles
	if len(gaps) == 0 {
		// Preserve the no-gap invariant: nil means every bound target has type
		// information and avoids a redundant per-file membership lookup.
		resolverOptions.TypeInfoFiles = nil
	}
	resolverOptions.ConfigPathBySourcePath = binding.ConfigPathBySourcePath
	resolverOptions.OwnerConfigDirBySourcePath = binding.OwnerConfigDirBySourcePath
	resolverOptions.SourceMappingsCanonical = true
	resolverOptions.FS = buildContext.FS()
	resolver := newLintConfigResolver(resolverOptions)
	if len(gaps) == 0 {
		finalizeLintTargetBinding(&binding)
		return binding, resolver, nil
	}

	useCaseSensitive := true
	if fsys := buildContext.FS(); fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	// Direct parsing is valid only for roots the fallback Program itself would
	// admit. Unsupported explicit targets must retain the legacy failure path.
	if anyGapHasActiveRules(gaps, resolver, singleThreaded) || !allGapRootsSupportedByFallback(gaps, useCaseSensitive) {
		if err := appendFallbackPrograms(&binding, gaps, currentDirectory, buildContext, singleThreaded); err != nil {
			return lintTargetBinding{}, nil, err
		}
	} else {
		binding.standaloneGapGroups = groupFallbackTargets(gaps, currentDirectory, useCaseSensitive)
	}
	finalizeLintTargetBinding(&binding)
	return binding, resolver, nil
}

type fallbackSourceParser struct {
	host     compiler.CompilerHost
	options  *core.CompilerOptions
	resolver *module.Resolver
}

func newFallbackSourceParser(
	currentDirectory string,
	buildContext *utils.ProgramBuildContext,
) *fallbackSourceParser {
	options := fallbackCompilerOptions()
	host := buildContext.NewTransientCompilerHost(currentDirectory)
	return &fallbackSourceParser{
		host:     host,
		options:  options,
		resolver: module.NewResolver(host, options, "", ""),
	}
}

func (p *fallbackSourceParser) sourceFileMetaData(fileName string) ast.SourceFileMetaData {
	packageJsonScope := p.resolver.GetPackageScopeForPath(tspath.GetDirectoryPath(fileName))
	moduleResolutionKind := p.options.GetModuleResolutionKind()

	var packageJsonType, packageJsonDirectory string
	if packageJsonScope.Exists() {
		packageJsonDirectory = packageJsonScope.PackageDirectory
		if value, ok := packageJsonScope.Contents.Type.GetValue(); ok {
			hasExplicitFormat := tspath.FileExtensionIsOneOf(fileName, []string{
				tspath.ExtensionMts,
				tspath.ExtensionCts,
				tspath.ExtensionMjs,
				tspath.ExtensionCjs,
			})
			usesNodeFormat := core.ModuleResolutionKindNode16 <= moduleResolutionKind &&
				moduleResolutionKind <= core.ModuleResolutionKindNodeNext
			if (!hasExplicitFormat && usesNodeFormat) || strings.Contains(fileName, "/node_modules/") {
				packageJsonType = value
			}
		}
	}

	return ast.SourceFileMetaData{
		PackageJsonType:      packageJsonType,
		PackageJsonDirectory: packageJsonDirectory,
		ImpliedNodeFormat:    ast.GetImpliedNodeFormatForFile(fileName, packageJsonType),
	}
}

func (p *fallbackSourceParser) parse(target resolvedLintTarget) *ast.SourceFile {
	fileName := tspath.GetNormalizedAbsolutePath(target.Path, p.host.GetCurrentDirectory())
	metadata := p.sourceFileMetaData(fileName)
	return p.host.GetSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path: tspath.ToPath(
			fileName,
			p.host.GetCurrentDirectory(),
			p.host.FS().UseCaseSensitiveFileNames(),
		),
		ExternalModuleIndicatorOptions: ast.GetExternalModuleIndicatorOptions(fileName, p.options, metadata),
	})
}

func (p *fallbackSourceParser) syntacticDiagnostics(file *ast.SourceFile) []*ast.Diagnostic {
	diagnostics := append([]*ast.Diagnostic(nil), file.Diagnostics()...)
	diagnostics = append(diagnostics, file.JSDiagnostics()...)
	if ast.IsSourceFileJS(file) && !ast.IsCheckJSEnabledForFile(file, p.options) {
		diagnostics = append(diagnostics, compiler.GetAdditionalJSSyntacticDiagnostics(file, p.options)...)
	}
	return compiler.SortAndDeduplicateDiagnostics(diagnostics)
}

type standaloneGapParseResult struct {
	fileName    string
	path        tspath.Path
	diagnostics []rule.RuleDiagnostic
	err         error
}

type standaloneGapRef struct {
	groupIndex  int
	targetIndex int
}

func collectStandaloneGapSyntacticDiagnostics(
	groups [][]resolvedLintTarget,
	currentDirectory string,
	buildContext *utils.ProgramBuildContext,
	singleThreaded bool,
) ([]rule.RuleDiagnostic, int32, error) {
	if len(groups) == 0 {
		return nil, 0, nil
	}
	parsers := make([]*fallbackSourceParser, len(groups))
	results := make([][]standaloneGapParseResult, len(groups))
	var refs []standaloneGapRef
	for groupIndex, group := range groups {
		// A resolver's package.json cache uses the host's case-sensitivity for
		// path keys. Keep collision groups isolated exactly as fallback Programs
		// were, so distinct exact-case roots can never share package metadata.
		parsers[groupIndex] = newFallbackSourceParser(currentDirectory, buildContext)
		results[groupIndex] = make([]standaloneGapParseResult, len(group))
		for targetIndex := range group {
			refs = append(refs, standaloneGapRef{groupIndex: groupIndex, targetIndex: targetIndex})
		}
	}

	parse := func(ref standaloneGapRef) {
		target := groups[ref.groupIndex][ref.targetIndex]
		parser := parsers[ref.groupIndex]
		file := parser.parse(target)
		if file == nil {
			results[ref.groupIndex][ref.targetIndex].err = fmt.Errorf(
				"fallback Program did not contain lint target %q",
				target.Path,
			)
			return
		}

		diagnostics := parser.syntacticDiagnostics(file)
		converted := make([]rule.RuleDiagnostic, len(diagnostics))
		for i, diagnostic := range diagnostics {
			converted[i] = typeScriptRuleDiagnostic(file, diagnostic)
		}
		results[ref.groupIndex][ref.targetIndex] = standaloneGapParseResult{
			fileName:    file.FileName(),
			path:        file.Path(),
			diagnostics: converted,
		}
	}

	workerCount := min(runtime.GOMAXPROCS(0), len(refs))
	if singleThreaded || workerCount < 2 {
		for _, ref := range refs {
			parse(ref)
		}
	} else {
		chunkSize := (len(refs) + workerCount - 1) / workerCount
		work := core.NewWorkGroup(false)
		for worker := range workerCount {
			start := worker * chunkSize
			end := min(start+chunkSize, len(refs))
			if start >= end {
				continue
			}
			chunk := refs[start:end]
			work.Queue(func() {
				for _, ref := range chunk {
					parse(ref)
				}
			})
		}
		work.RunAndWait()
	}

	var diagnostics []rule.RuleDiagnostic
	var lintedFileCount int32
	for _, group := range results {
		seen := make(map[string]struct{}, len(group))
		for _, result := range group {
			if result.err != nil {
				return nil, 0, result.err
			}
			diagnostics = append(diagnostics, result.diagnostics...)
			if _, exists := seen[result.fileName]; exists {
				continue
			}
			seen[result.fileName] = struct{}{}
			if !utils.IsExcludedLintPath(string(result.path), utils.ExcludePaths) {
				lintedFileCount++
			}
		}
	}
	return diagnostics, lintedFileCount, nil
}

func discoverLintFilesMultiConfig(
	configMap map[string]rslintconfig.RslintConfig,
	configTargetScopes map[string]rslintconfig.LintDiscoveryScope,
	fs vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) []rslintconfig.DiscoveredLintTarget {
	return rslintconfig.DiscoverLintTargetsMultiConfig(configMap, configTargetScopes, fs, allowFiles, allowDirs, singleThreaded)
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

func typeScriptRuleDiagnostic(file *ast.SourceFile, diagnostic *ast.Diagnostic) rule.RuleDiagnostic {
	return rule.RuleDiagnostic{
		RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
		SourceFile:   file,
		FilePath:     file.FileName(),
		Range:        diagnostic.Loc(),
		Message:      rule.RuleMessage{Description: diagnostic.String()},
		Severity:     rule.SeverityError,
		Origin:       rule.DiagnosticOriginTypeScript,
		PreFormatted: true,
	}
}

func collectTargetSyntacticDiagnostics(
	programs []*compiler.Program,
	targetsByProgram [][]string,
	skipTypeCheck []bool,
	typeCheck bool,
	typeCheckOnly bool,
) ([]rule.RuleDiagnostic, map[string]struct{}) {
	syntaxErrorFiles := make(map[string]struct{})
	if len(programs) == 0 || len(targetsByProgram) == 0 {
		return nil, syntaxErrorFiles
	}

	seen := make(map[syntacticDiagnosticKey]struct{})
	var diagnostics []rule.RuleDiagnostic
	for i, program := range programs {
		// When --type-check runs, tsconfig-backed Programs surface syntactic
		// diagnostics through the type-check phase. We still inspect every target
		// here so the lint-rule phase can skip malformed files, matching ESLint.
		coveredByTypeCheck := typeCheck && (i >= len(skipTypeCheck) || !skipTypeCheck[i])
		if i >= len(targetsByProgram) || len(targetsByProgram[i]) == 0 {
			continue
		}
		ctx := context.Background()
		for _, target := range targetsByProgram[i] {
			file := program.GetSourceFile(target)
			if file == nil {
				continue
			}
			for _, diagnostic := range program.GetSyntacticDiagnostics(ctx, file) {
				syntaxErrorFiles[file.FileName()] = struct{}{}
				if coveredByTypeCheck || typeCheckOnly {
					continue
				}
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
				diagnostics = append(diagnostics, typeScriptRuleDiagnostic(file, diagnostic))
			}
		}
	}
	return diagnostics, syntaxErrorFiles
}
