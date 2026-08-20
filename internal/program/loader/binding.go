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
// only the unified Program sequence, lint projection, and caller-path mapping
// needed by integrations; compiler and parser assembly details remain
// private to the loader. Its slices and maps are immutable after LoadCLI or
// LoadAPI returns.
type LoadResult struct {
	compilerPrograms        []*compiler.Program
	Programs                []*lintprogram.Program
	TypeCheckPrograms       []*lintprogram.Program
	TargetsByProgram        [][]string
	TargetPathBySourcePath  map[string]string
	targetIndexBySourcePath map[string]int
}

// TargetIndexForSourcePath returns the immutable lint-plan target represented
// by a Program source path. Rules and plugin dispatch use this index instead of
// re-running config ownership or files matching after Program binding.
func (result LoadResult) TargetIndexForSourcePath(sourcePath string) (int, bool) {
	index, ok := result.targetIndexBySourcePath[exactPathID(sourcePath)]
	return index, ok
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

func storeSourcePathMapping(mapping map[string]string, sourcePath string, canonicalSourcePath string, value string) {
	if mapping == nil {
		return
	}
	normalizedSource := tspath.NormalizePath(sourcePath)
	mapping[normalizedSource] = value
	if canonicalSourcePath != "" {
		mapping[exactPathID(canonicalSourcePath)] = value
	}
}

func storeTargetIndexMapping(mapping map[string]int, sourcePath string, canonicalSourcePath string, index int) {
	if mapping == nil {
		return
	}
	mapping[exactPathID(sourcePath)] = index
	if canonicalSourcePath != "" {
		mapping[exactPathID(canonicalSourcePath)] = index
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
	programs           []*compiler.Program
	targets            []rslintconfig.DiscoveredLintTarget
	fsys               vfs.FS
	identities         *lintprogram.PathIdentityResolver
	initialized        bool
	builtByProgram     []bool
	sourcesByProgram   []map[string]*ast.SourceFile
	targetCanonicalIDs map[string]struct{}
}

func newProgramFileIndex(
	programs []*compiler.Program,
	targets []rslintconfig.DiscoveredLintTarget,
	fsys vfs.FS,
	singleThreaded bool,
) *programFileIndex {
	known := make([]lintprogram.PathIdentity, len(targets))
	for index, target := range targets {
		known[index] = lintprogram.PathIdentity{
			Path:          target.Path,
			CanonicalPath: target.CanonicalPath,
		}
	}
	return newProgramFileIndexWithResolver(
		programs,
		targets,
		fsys,
		lintprogram.NewPathIdentityResolver(fsys, singleThreaded, known),
	)
}

func newProgramFileIndexWithResolver(
	programs []*compiler.Program,
	targets []rslintconfig.DiscoveredLintTarget,
	fsys vfs.FS,
	identities *lintprogram.PathIdentityResolver,
) *programFileIndex {
	return &programFileIndex{
		programs:   programs,
		targets:    targets,
		fsys:       fsys,
		identities: identities,
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
	var unresolvedPaths []string
	for _, target := range index.targets {
		if target.CanonicalPath != "" {
			index.targetCanonicalIDs[exactPathID(target.CanonicalPath)] = struct{}{}
		} else if target.Path != "" {
			unresolvedPaths = append(unresolvedPaths, target.Path)
		}
	}
	for _, canonicalID := range index.identities.CanonicalPathIDs(unresolvedPaths) {
		if canonicalID != "" {
			index.targetCanonicalIDs[canonicalID] = struct{}{}
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

func (index *programFileIndex) canonicalPath(path string, canonicalPath string) string {
	if canonicalPath != "" || index == nil || path == "" {
		return canonicalPath
	}
	canonicalIDs := index.identities.CanonicalPathIDs([]string{path})
	if len(canonicalIDs) == 0 {
		return ""
	}
	return canonicalIDs[0]
}

type programSourceMembership struct {
	programIndex int
	sourceIndex  int
	sourceFile   *ast.SourceFile
}

func (index *programFileIndex) buildPrograms(programIndexes []int) {
	sourceIndexByPath := make(map[string]int)
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

	canonicalIDs := index.identities.CanonicalPathIDs(sourcePaths)

	for _, membership := range memberships {
		canonicalID := canonicalIDs[membership.sourceIndex]
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
	targets []plannedTargetRef,
	currentDirectory string,
	useCaseSensitive bool,
) [][]plannedTargetRef {
	if len(targets) == 0 {
		return nil
	}

	groups := make([][]plannedTargetRef, 0, 1)
	keysByGroup := make([]map[tspath.Path]struct{}, 0, 1)
	for _, target := range targets {
		key := tspath.ToPath(target.target.Target.Path, currentDirectory, useCaseSensitive)
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

func (s *Session) appendCompatibilityPrograms(
	binding *LoadResult,
	targets []plannedTargetRef,
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
			rootFileNames = append(rootFileNames, target.target.Target.Path)
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
		for _, targetRef := range group {
			target := targetRef.target
			sourceFile := exactProgramSourceFile(compilerProgram, target.Target.Path)
			if sourceFile == nil {
				return fmt.Errorf("program did not contain lint target %q", target.Target.Path)
			}
			sourcePath := sourceFile.FileName()
			binding.TargetsByProgram[programIndex] = append(binding.TargetsByProgram[programIndex], sourcePath)
			storeTargetIndexMapping(
				binding.targetIndexBySourcePath,
				sourcePath,
				target.Target.CanonicalPath,
				targetRef.index,
			)
			if tspath.NormalizePath(sourcePath) != target.Target.Path {
				storeSourcePathMapping(
					binding.TargetPathBySourcePath,
					sourcePath,
					target.Target.CanonicalPath,
					target.Target.Path,
				)
			}
		}
	}
	return nil
}

func finalizeResult(binding *LoadResult) {
	for i := range binding.TargetsByProgram {
		sort.Strings(binding.TargetsByProgram[i])
	}
	if len(binding.TargetPathBySourcePath) == 0 {
		binding.TargetPathBySourcePath = nil
	}
}

// LoadAPI preserves the API's compatibility Program admission behavior while
// returning only unified Programs to the caller.
func (s *Session) LoadAPI(
	set ProjectSet,
	plan rslintconfig.LintProjectPlan,
	currentDirectory string,
	singleThreaded bool,
) (LoadResult, error) {
	if err := s.validate(); err != nil {
		return LoadResult{}, err
	}
	binding, unbound, err := s.bindTargetsToProjects(set, plan, singleThreaded)
	if err != nil {
		return LoadResult{}, err
	}
	if err := s.appendCompatibilityPrograms(&binding, unbound, currentDirectory, singleThreaded); err != nil {
		return LoadResult{}, err
	}
	finalizeResult(&binding)
	return binding, nil
}

func allRootsSupportedByParser(targets []plannedTargetRef, useCaseSensitive bool) bool {
	options := sourceOnlyCompilerOptions()
	supportedExtensions := tsoptions.GetSupportedExtensionsWithJsonIfResolveJsonModule(options, tspath.AllSupportedExtensions)
	for _, targetRef := range targets {
		target := targetRef.target.Target
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
	groups [][]plannedTargetRef,
	currentDirectory string,
	singleThreaded bool,
) error {
	for _, group := range groups {
		sort.Slice(group, func(left, right int) bool {
			return group[left].target.Target.Path < group[right].target.Target.Path
		})
		rootFileNames := make([]string, len(group))
		for index, target := range group {
			rootFileNames[index] = target.target.Target.Path
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
			storeTargetIndexMapping(
				binding.targetIndexBySourcePath,
				file.FileName(),
				group[index].target.Target.CanonicalPath,
				group[index].index,
			)
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
	plan rslintconfig.LintProjectPlan,
	currentDirectory string,
	singleThreaded bool,
) (LoadResult, error) {
	if err := s.validate(); err != nil {
		return LoadResult{}, err
	}
	binding, unbound, err := s.bindTargetsToProjects(set, plan, singleThreaded)
	if err != nil {
		return LoadResult{}, err
	}
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
