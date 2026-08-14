package main

import (
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type lintTargetBinding struct {
	CompilerPrograms           []*compiler.Program
	Programs                   []*lintprogram.Program
	TargetsByProgram           [][]string
	TargetPathBySourcePath     map[string]string
	ConfigPathBySourcePath     map[string]string
	OwnerConfigDirBySourcePath map[string]string
	GapGroups                  [][]resolvedLintTarget
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

func orderedProgramIndexesForConfig(set compilerProgramSet, configDir string) []int {
	configDirID := exactFilesystemPathID(configDir)
	indexes := make([]int, 0, len(set.CompilerPrograms))
	for i := range set.CompilerPrograms {
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

func bindLintTargetsToRealPrograms(
	set compilerProgramSet,
	plan lintTargetPlan,
	buildContext *utils.ProgramBuildContext,
	singleThreaded bool,
) (lintTargetBinding, []resolvedLintTarget) {
	fsys := buildContext.FS()
	binding := lintTargetBinding{
		CompilerPrograms:           append([]*compiler.Program(nil), set.CompilerPrograms...),
		Programs:                   lintprogram.NewFromCompilers(set.CompilerPrograms),
		TargetsByProgram:           make([][]string, len(set.CompilerPrograms)),
		TargetPathBySourcePath:     make(map[string]string),
		ConfigPathBySourcePath:     make(map[string]string),
		OwnerConfigDirBySourcePath: make(map[string]string),
	}

	var gaps []resolvedLintTarget
	programIndexesByConfig := make(map[string][]int)
	programFiles := newProgramFileIndex(set.CompilerPrograms, plan.Targets, fsys, singleThreaded)
	for _, target := range plan.Targets {
		programIndexes, cached := programIndexesByConfig[target.OwnerConfigDir]
		if !cached {
			programIndexes = orderedProgramIndexesForConfig(set, target.OwnerConfigDir)
			programIndexesByConfig[target.OwnerConfigDir] = programIndexes
		}
		bound := false
		for _, programIndex := range programIndexes {
			sourceFile := exactProgramSourceFile(set.CompilerPrograms[programIndex], target.Path)
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
			if tspath.NormalizePath(sourcePath) != target.Path {
				storeSourcePathMapping(binding.TargetPathBySourcePath, sourcePath, target.CanonicalPath, target.Path)
			}
			bound = true
			break
		}
		if !bound {
			gaps = append(gaps, target)
			storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, target.Path, target.CanonicalPath, target.OwnerConfigDir)
			storeSourcePathMapping(binding.ConfigPathBySourcePath, target.Path, target.CanonicalPath, configPathForLintTarget(target, fsys))
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
		fallbackIndex := len(binding.CompilerPrograms)
		binding.CompilerPrograms = append(binding.CompilerPrograms, fallback)
		fallbackProgram, err := lintprogram.NewFromBoundSources(fallback, fallback.SourceFiles())
		if err != nil {
			return fmt.Errorf("adapt fallback Program: %w", err)
		}
		binding.Programs = append(binding.Programs, fallbackProgram)
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
				storeSourcePathMapping(binding.ConfigPathBySourcePath, sourcePath, gap.CanonicalPath, binding.ConfigPathBySourcePath[tspath.NormalizePath(gap.Path)])
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
	for i := range binding.GapGroups {
		sort.Slice(binding.GapGroups[i], func(left, right int) bool {
			return binding.GapGroups[i][left].Path < binding.GapGroups[i][right].Path
		})
	}
	if len(binding.TargetPathBySourcePath) == 0 {
		binding.TargetPathBySourcePath = nil
	}
	if len(binding.GapGroups) == 0 {
		binding.GapGroups = nil
	}
}

// bindLintTargetPlan retains the Program-backed API contract.
func bindLintTargetPlan(
	set compilerProgramSet,
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

func allGapRootsSupportedByRootParser(gaps []resolvedLintTarget, useCaseSensitive bool) bool {
	options := fallbackCompilerOptions()
	supportedExtensions := tsoptions.GetSupportedExtensionsWithJsonIfResolveJsonModule(options, tspath.AllSupportedExtensions)
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

// bindCLILintTargetPlan represents every supported gap as parser/binder input
// to an rslint Program. Unsupported roots retain the legacy ts-go admission
// and error behavior.
func bindCLILintTargetPlan(
	set compilerProgramSet,
	plan lintTargetPlan,
	currentDirectory string,
	buildContext *utils.ProgramBuildContext,
	singleThreaded bool,
) (lintTargetBinding, error) {
	binding, gaps := bindLintTargetsToRealPrograms(set, plan, buildContext, singleThreaded)
	if len(gaps) == 0 {
		finalizeLintTargetBinding(&binding)
		return binding, nil
	}
	useCaseSensitive := true
	if fsys := buildContext.FS(); fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	if !allGapRootsSupportedByRootParser(gaps, useCaseSensitive) {
		if err := appendFallbackPrograms(&binding, gaps, currentDirectory, buildContext, singleThreaded); err != nil {
			return lintTargetBinding{}, err
		}
	} else {
		binding.GapGroups = groupFallbackTargets(gaps, currentDirectory, useCaseSensitive)
	}
	finalizeLintTargetBinding(&binding)
	return binding, nil
}
