package loader

import (
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

// LoadResult is the complete Program input for one lint generation. It carries
// only the unified Program sequence, lint projection, and source/target path
// mappings needed by integrations; compiler and parser assembly details remain
// private to the loader. Its slices and maps are immutable after LoadCLI or
// LoadAPI returns.
type LoadResult struct {
	compilerPrograms       []*compiler.Program
	Programs               []*lintprogram.Program
	TargetsByProgram       [][]string
	LintTargetBySourcePath map[string]rslintconfig.DiscoveredLintTarget
}

func sourceOnlyCompilerOptions() *core.CompilerOptions {
	return &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}
}

func authoritativePath(filePath string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
	}
	return filePath
}

func canonicalPathID(filePath string, fsys vfs.FS) string {
	return exactPathID(authoritativePath(filePath, fsys))
}

func storeSourceTargetMapping(
	mapping map[string]rslintconfig.DiscoveredLintTarget,
	sourcePath string,
	canonicalSourcePath string,
	target rslintconfig.DiscoveredLintTarget,
) {
	if mapping == nil {
		return
	}
	normalizedSource := tspath.NormalizePath(sourcePath)
	mapping[normalizedSource] = target
	if canonicalSourcePath != "" {
		mapping[exactPathID(canonicalSourcePath)] = target
	}
}

func exactProgramSourceFile(program *compiler.Program, targetPath string) *ast.SourceFile {
	if program == nil || targetPath == "" {
		return nil
	}
	targetPath = tspath.NormalizePath(targetPath)
	sourceFile := program.GetSourceFile(targetPath)
	if sourceFile == nil || exactPathID(sourceFile.FileName()) != exactPathID(targetPath) {
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
	targets               []rslintconfig.DiscoveredLintTarget
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
	targets []rslintconfig.DiscoveredLintTarget,
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
		canonicalID := exactPathID(target.CanonicalPath)
		index.targetCanonicalIDs[canonicalID] = struct{}{}
		index.canonicalBySourcePath[canonicalID] = canonicalID
	}
	for _, target := range index.targets {
		if target.Path == "" || target.CanonicalPath == "" {
			continue
		}
		sourcePathID := exactPathID(target.Path)
		if _, isCanonicalTarget := index.targetCanonicalIDs[sourcePathID]; !isCanonicalTarget {
			index.canonicalBySourcePath[sourcePathID] = exactPathID(target.CanonicalPath)
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
	return index.sourcesByProgram[programIndex][exactPathID(canonicalTarget)]
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
		directoryID := exactPathID(directoryPath)
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
				canonicalIDs[sourceIndex] = exactPathID(
					tspath.CombinePaths(directory.canonicalPath, fileName),
				)
			} else {
				canonicalIDs[sourceIndex] = canonicalPathID(sourcePath, index.fsys)
			}
		})
	}
	for directoryIndex, directoryID := range directoryIDs {
		sourceIndexes := sourceIndexesByDirectoryID[directoryID]
		directory, cached := index.directoryIdentities[directoryID]
		if len(sourceIndexes) == 1 && !cached {
			sourceIndex := sourceIndexes[0]
			work.Queue(func() {
				canonicalIDs[sourceIndex] = canonicalPathID(sourcePaths[sourceIndex], index.fsys)
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
					canonicalPath: authoritativePath(directoryPath, index.fsys),
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
			sourcePathID := exactPathID(sourcePath)
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

func groupUnboundTargets(
	targets []rslintconfig.DiscoveredLintTarget,
	currentDirectory string,
	useCaseSensitive bool,
) [][]rslintconfig.DiscoveredLintTarget {
	if len(targets) == 0 {
		return nil
	}

	groups := make([][]rslintconfig.DiscoveredLintTarget, 0, 1)
	keysByGroup := make([]map[tspath.Path]struct{}, 0, 1)
	for _, target := range targets {
		key := tspath.ToPath(target.Path, currentDirectory, useCaseSensitive)
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
		groups[groupIndex] = append(groups[groupIndex], target)
		keysByGroup[groupIndex][key] = struct{}{}
	}
	return groups
}

func orderedProgramIndexesForConfig(set ProjectSet, configDir string) []int {
	configDirID := exactPathID(configDir)
	indexes := make([]int, 0, len(set.compilerPrograms))
	for i := range set.compilerPrograms {
		if i < len(set.configOrders) {
			if _, ok := set.configOrders[i][configDirID]; ok {
				indexes = append(indexes, i)
			}
		}
	}
	sort.SliceStable(indexes, func(i, j int) bool {
		left := set.configOrders[indexes[i]][configDirID]
		right := set.configOrders[indexes[j]][configDirID]
		if left != right {
			return left < right
		}
		return indexes[i] < indexes[j]
	})
	return indexes
}

func (s *Session) appendCompatibilityPrograms(
	binding *LoadResult,
	targets []rslintconfig.DiscoveredLintTarget,
	currentDirectory string,
	singleThreaded bool,
) error {
	if len(targets) == 0 {
		return nil
	}
	fsys := s.FS()
	useCaseSensitive := true
	if fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	for _, group := range groupUnboundTargets(targets, currentDirectory, useCaseSensitive) {
		rootFileNames := make([]string, 0, len(group))
		for _, target := range group {
			rootFileNames = append(rootFileNames, target.Path)
		}
		compilerProgram, err := s.context.createCompatibilityProgram(
			singleThreaded,
			currentDirectory,
			sourceOnlyCompilerOptions(),
			rootFileNames,
		)
		if err != nil {
			return fmt.Errorf("create Program for %d lint target(s): %w", len(rootFileNames), err)
		}
		if compilerProgram == nil {
			return fmt.Errorf("create Program for %d lint target(s): no Program returned", len(group))
		}
		programIndex := len(binding.compilerPrograms)
		binding.compilerPrograms = append(binding.compilerPrograms, compilerProgram)
		sourceProgram, err := lintprogram.NewFromBoundSources(compilerProgram, compilerProgram.SourceFiles())
		if err != nil {
			return fmt.Errorf("adapt source-only Program: %w", err)
		}
		binding.Programs = append(binding.Programs, sourceProgram)
		binding.TargetsByProgram = append(binding.TargetsByProgram, nil)
		for _, target := range group {
			sourceFile := exactProgramSourceFile(compilerProgram, target.Path)
			if sourceFile == nil {
				return fmt.Errorf("program did not contain lint target %q", target.Path)
			}
			sourcePath := sourceFile.FileName()
			binding.TargetsByProgram[programIndex] = append(binding.TargetsByProgram[programIndex], sourcePath)
			storeSourceTargetMapping(binding.LintTargetBySourcePath, sourcePath, target.CanonicalPath, target)
		}
	}
	return nil
}

func finalizeResult(binding *LoadResult) {
	for i := range binding.TargetsByProgram {
		sort.Strings(binding.TargetsByProgram[i])
	}
	if len(binding.LintTargetBySourcePath) == 0 {
		binding.LintTargetBySourcePath = nil
	}
}

// LoadAPI preserves the API's compatibility Program admission behavior while
// returning only unified Programs to the caller.
func (s *Session) LoadAPI(
	set ProjectSet,
	plan rslintconfig.LintTargetPlan,
	currentDirectory string,
	singleThreaded bool,
) (LoadResult, error) {
	if err := s.validate(); err != nil {
		return LoadResult{}, err
	}
	binding, unbound := s.bindTargetsToProjects(set, plan, singleThreaded)
	if err := s.appendCompatibilityPrograms(&binding, unbound, currentDirectory, singleThreaded); err != nil {
		return LoadResult{}, err
	}
	finalizeResult(&binding)
	return binding, nil
}

func allRootsSupportedByParser(targets []rslintconfig.DiscoveredLintTarget, useCaseSensitive bool) bool {
	options := sourceOnlyCompilerOptions()
	supportedExtensions := tsoptions.GetSupportedExtensionsWithJsonIfResolveJsonModule(options, tspath.AllSupportedExtensions)
	for _, target := range targets {
		if !tspath.HasExtension(target.Path) {
			return false
		}
		fileName := tspath.GetCanonicalFileName(target.Path, useCaseSensitive)
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

func (s *Session) appendRootPrograms(
	binding *LoadResult,
	groups [][]rslintconfig.DiscoveredLintTarget,
	currentDirectory string,
	singleThreaded bool,
) error {
	for _, group := range groups {
		sort.Slice(group, func(left, right int) bool {
			return group[left].Path < group[right].Path
		})
		rootFileNames := make([]string, len(group))
		for index, target := range group {
			rootFileNames[index] = target.Path
		}
		rootProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
			RootFileNames:   rootFileNames,
			Host:            s.context.newTransientCompilerHost(currentDirectory),
			CompilerOptions: sourceOnlyCompilerOptions(),
			SingleThreaded:  singleThreaded,
		})
		if err != nil {
			return err
		}
		binding.Programs = append(binding.Programs, rootProgram)
		files := rootProgram.SourceFiles()
		targets := make([]string, len(files))
		for index, file := range files {
			targets[index] = file.FileName()
		}
		binding.TargetsByProgram = append(binding.TargetsByProgram, targets)
	}
	return nil
}

// LoadCLI binds targets into one backend-agnostic Program sequence. Supported
// project-external roots use the native parser/binder facade; roots that the
// facade cannot admit retain the compatibility compiler behavior internally.
func (s *Session) LoadCLI(
	set ProjectSet,
	plan rslintconfig.LintTargetPlan,
	currentDirectory string,
	singleThreaded bool,
) (LoadResult, error) {
	if err := s.validate(); err != nil {
		return LoadResult{}, err
	}
	binding, unbound := s.bindTargetsToProjects(set, plan, singleThreaded)
	useCaseSensitive := true
	if fsys := s.FS(); fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	if !allRootsSupportedByParser(unbound, useCaseSensitive) {
		if err := s.appendCompatibilityPrograms(&binding, unbound, currentDirectory, singleThreaded); err != nil {
			return LoadResult{}, err
		}
	} else {
		groups := groupUnboundTargets(unbound, currentDirectory, useCaseSensitive)
		// Project ASTs that were parsed while building broad tsconfigs but retained
		// by no current compiler Program are evicted before root parsing. The root
		// backend shares source snapshots, not bound compiler AST objects.
		s.retainCompilerPrograms(binding.compilerPrograms)
		if err := s.appendRootPrograms(&binding, groups, currentDirectory, singleThreaded); err != nil {
			return LoadResult{}, err
		}
		finalizeResult(&binding)
		return binding, nil
	}
	s.retainCompilerPrograms(binding.compilerPrograms)
	finalizeResult(&binding)
	return binding, nil
}
