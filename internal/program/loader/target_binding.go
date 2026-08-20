package loader

import (
	"errors"
	"fmt"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/program/projectselection"
)

func targetBindingOwners(
	set ProjectSet,
	targets []rslintconfig.DiscoveredLintTarget,
) ([]int, bool, error) {
	binding := set.targetBinding
	if binding == nil {
		return nil, false, nil
	}
	if len(binding.targets) != len(targets) || len(binding.owners) != len(targets) {
		return nil, false, errors.New("targeted project binding does not match the lint target plan")
	}
	for targetIndex, target := range targets {
		if binding.targets[targetIndex] != target {
			return nil, false, errors.New("targeted project binding does not match the lint target plan")
		}
		owner := binding.owners[targetIndex]
		if owner < projectselection.NoProject || owner >= len(set.compilerPrograms) {
			return nil, false, fmt.Errorf(
				"targeted project binding contains invalid owner %d for %q",
				owner,
				target.Path,
			)
		}
	}
	return append([]int(nil), binding.owners...), true, nil
}

type prebuiltProjectMetadata struct {
	program        *compiler.Program
	exactRoots     map[string]struct{}
	canonicalRoots map[string]struct{}
	provider       *prebuiltProjectMetadataProvider
}

type prebuiltProjectRootMembership struct {
	project int
	root    int
}

type prebuiltProjectMetadataProvider struct {
	metadata     []*prebuiltProjectMetadata
	rootPaths    []string
	memberships  []prebuiltProjectRootMembership
	programFiles *programFileIndex
	fsys         vfs.FS
	canonical    bool
}

func newPrebuiltProjectMetadataProvider(
	programs []*compiler.Program,
	programFiles *programFileIndex,
	fsys vfs.FS,
) *prebuiltProjectMetadataProvider {
	provider := &prebuiltProjectMetadataProvider{
		metadata:     make([]*prebuiltProjectMetadata, len(programs)),
		programFiles: programFiles,
		fsys:         fsys,
	}
	rootIndexByPath := make(map[string]int)
	for project, program := range programs {
		if program == nil || program.CommandLine() == nil {
			continue
		}
		metadata := &prebuiltProjectMetadata{
			program:    program,
			exactRoots: make(map[string]struct{}, len(program.CommandLine().FileNames())),
			provider:   provider,
		}
		provider.metadata[project] = metadata
		for _, rootPath := range program.CommandLine().FileNames() {
			rootPath = tspath.NormalizePath(rootPath)
			rootID := exactPathID(rootPath)
			metadata.exactRoots[rootID] = struct{}{}
			rootIndex, exists := rootIndexByPath[rootID]
			if !exists {
				rootIndex = len(provider.rootPaths)
				rootIndexByPath[rootID] = rootIndex
				provider.rootPaths = append(provider.rootPaths, rootPath)
			}
			provider.memberships = append(provider.memberships, prebuiltProjectRootMembership{
				project: project,
				root:    rootIndex,
			})
		}
	}
	return provider
}

func (provider *prebuiltProjectMetadataProvider) ensureCanonicalRoots() {
	if provider == nil || provider.canonical {
		return
	}
	provider.canonical = true
	canonicalIDs := make([]string, len(provider.rootPaths))
	provider.programFiles.initialize()
	pendingPaths := make([]string, 0, len(provider.rootPaths))
	pendingIndexes := make([]int, 0, len(provider.rootPaths))
	for index, rootPath := range provider.rootPaths {
		rootID := exactPathID(rootPath)
		if canonicalID, known := provider.programFiles.canonicalBySourcePath[rootID]; known {
			canonicalIDs[index] = canonicalID
			continue
		}
		if provider.fsys == nil {
			canonicalIDs[index] = rootID
			continue
		}
		pendingPaths = append(pendingPaths, rootPath)
		pendingIndexes = append(pendingIndexes, index)
	}
	resolvedIDs := provider.programFiles.canonicalSourcePathIDs(pendingPaths)
	for pendingIndex, canonicalID := range resolvedIDs {
		rootIndex := pendingIndexes[pendingIndex]
		canonicalIDs[rootIndex] = canonicalID
		provider.programFiles.canonicalBySourcePath[exactPathID(provider.rootPaths[rootIndex])] = canonicalID
	}
	for _, membership := range provider.memberships {
		metadata := provider.metadata[membership.project]
		if metadata.canonicalRoots == nil {
			metadata.canonicalRoots = make(map[string]struct{})
		}
		metadata.canonicalRoots[canonicalIDs[membership.root]] = struct{}{}
	}
}

func (metadata *prebuiltProjectMetadata) DirectRoot(target projectselection.Target) bool {
	if metadata == nil {
		return false
	}
	if _, exists := metadata.exactRoots[exactPathID(target.Path)]; exists {
		return true
	}
	metadata.provider.ensureCanonicalRoots()
	canonicalPath := target.CanonicalPath
	if canonicalPath == "" {
		canonicalPath = authoritativePath(target.Path, metadata.provider.fsys)
	}
	_, exists := metadata.canonicalRoots[exactPathID(canonicalPath)]
	return exists
}

func (metadata *prebuiltProjectMetadata) Supports(target projectselection.Target) bool {
	return metadata != nil && metadata.program != nil && metadata.program.CommandLine() != nil &&
		lintprogram.CompilerOptionsSupportFileName(
			metadata.program.CommandLine().CompilerOptions(),
			target.Path,
		)
}

func selectedTarget(
	target rslintconfig.DiscoveredLintTarget,
) projectselection.Target {
	return projectselection.Target{
		Path:          target.Path,
		CanonicalPath: target.CanonicalPath,
	}
}

// selectPrebuiltProjectOwners applies the same ownership state machine used by
// focused CLI/API requests and LSP. Eager construction changes only the
// provider: every configured Program is already available before selection.
func selectPrebuiltProjectOwners(
	set ProjectSet,
	targets []rslintconfig.DiscoveredLintTarget,
	programFiles *programFileIndex,
	fsys vfs.FS,
) ([]int, error) {
	owners := make([]int, len(targets))
	for index := range owners {
		owners[index] = projectselection.NoProject
	}
	if len(targets) == 0 || len(set.compilerPrograms) == 0 {
		return owners, nil
	}

	metadataProvider := newPrebuiltProjectMetadataProvider(
		set.compilerPrograms,
		programFiles,
		fsys,
	)
	targetIndexesByConfig := make(map[string][]int)
	configDirs := make([]string, 0)
	for targetIndex, target := range targets {
		if _, exists := targetIndexesByConfig[target.ConfigDirectory]; !exists {
			configDirs = append(configDirs, target.ConfigDirectory)
		}
		targetIndexesByConfig[target.ConfigDirectory] = append(
			targetIndexesByConfig[target.ConfigDirectory],
			targetIndex,
		)
	}

	for _, configDir := range configDirs {
		targetIndexes := targetIndexesByConfig[configDir]
		projectIndexes := orderedProgramIndexesForConfig(set, configDir)
		selectionTargets := make([]projectselection.Target, len(targetIndexes))
		for index, targetIndex := range targetIndexes {
			selectionTargets[index] = selectedTarget(targets[targetIndex])
		}
		bindings, err := projectselection.Resolve(
			projectselection.Plan{
				Targets:  selectionTargets,
				Projects: projectIndexes,
			},
			func(project int) (projectselection.Metadata, bool, error) {
				if project < 0 || project >= len(metadataProvider.metadata) {
					return nil, false, nil
				}
				metadata := metadataProvider.metadata[project]
				return metadata, metadata != nil, nil
			},
			func(project int) (bool, error) {
				return project >= 0 && project < len(set.compilerPrograms) &&
					set.compilerPrograms[project] != nil, nil
			},
			func(project int, target projectselection.Target) bool {
				if project < 0 || project >= len(set.compilerPrograms) {
					return false
				}
				if exactProgramSourceFile(set.compilerPrograms[project], target.Path) != nil {
					return true
				}
				return programFiles.sourceFile(
					projectIndexes,
					project,
					target.CanonicalPath,
				) != nil
			},
		)
		if err != nil {
			return nil, err
		}
		for index, binding := range bindings {
			owners[targetIndexes[index]] = binding.Project
		}
	}
	return owners, nil
}

func bindTargetToProgram(
	binding *LoadResult,
	set ProjectSet,
	programFiles *programFileIndex,
	programIndexes []int,
	programIndex int,
	target rslintconfig.DiscoveredLintTarget,
	fsys vfs.FS,
) bool {
	if programIndex < 0 || programIndex >= len(set.compilerPrograms) {
		return false
	}
	sourceFile := exactProgramSourceFile(set.compilerPrograms[programIndex], target.Path)
	if sourceFile == nil {
		sourceFile = programFiles.sourceFile(programIndexes, programIndex, target.CanonicalPath)
	}
	if sourceFile == nil {
		return false
	}
	sourcePath := sourceFile.FileName()
	binding.TargetsByProgram[programIndex] = append(binding.TargetsByProgram[programIndex], sourcePath)
	storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, sourcePath, target.CanonicalPath, target.ConfigDirectory)
	storeSourcePathMapping(binding.ConfigPathBySourcePath, sourcePath, target.CanonicalPath, target.MatchPath(fsys))
	if tspath.NormalizePath(sourcePath) != target.Path {
		storeSourcePathMapping(binding.TargetPathBySourcePath, sourcePath, target.CanonicalPath, target.Path)
	}
	return true
}

func (s *Session) bindTargetsToProjects(
	set ProjectSet,
	plan rslintconfig.LintTargetPlan,
	singleThreaded bool,
) (LoadResult, []rslintconfig.DiscoveredLintTarget, error) {
	fsys := s.FS()
	binding := LoadResult{
		compilerPrograms:           append([]*compiler.Program(nil), set.compilerPrograms...),
		Programs:                   append([]*lintprogram.Program(nil), set.programs...),
		TargetsByProgram:           make([][]string, len(set.compilerPrograms)),
		TargetPathBySourcePath:     make(map[string]string),
		ConfigPathBySourcePath:     make(map[string]string),
		OwnerConfigDirBySourcePath: make(map[string]string),
	}

	programFiles := newProgramFileIndex(set.compilerPrograms, plan.Targets, fsys, singleThreaded)
	owners, completeTargetBinding, err := targetBindingOwners(set, plan.Targets)
	if err != nil {
		return LoadResult{}, nil, err
	}
	if !completeTargetBinding {
		owners, err = selectPrebuiltProjectOwners(set, plan.Targets, programFiles, fsys)
		if err != nil {
			return LoadResult{}, nil, err
		}
	}

	var unbound []rslintconfig.DiscoveredLintTarget
	programIndexesByConfig := make(map[string][]int)
	for targetIndex, target := range plan.Targets {
		programIndexes, cached := programIndexesByConfig[target.ConfigDirectory]
		if !cached {
			programIndexes = orderedProgramIndexesForConfig(set, target.ConfigDirectory)
			programIndexesByConfig[target.ConfigDirectory] = programIndexes
		}
		owner := owners[targetIndex]
		if bindTargetToProgram(
			&binding,
			set,
			programFiles,
			programIndexes,
			owner,
			target,
			fsys,
		) {
			continue
		}
		if owner >= 0 {
			return LoadResult{}, nil, fmt.Errorf(
				"selected project %d did not contain lint target %q during binding",
				owner,
				target.Path,
			)
		}

		unbound = append(unbound, target)
		storeSourcePathMapping(binding.OwnerConfigDirBySourcePath, target.Path, target.CanonicalPath, target.ConfigDirectory)
		storeSourcePathMapping(binding.ConfigPathBySourcePath, target.Path, target.CanonicalPath, target.MatchPath(fsys))
	}
	return binding, unbound, nil
}
