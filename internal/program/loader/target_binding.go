package loader

import (
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

type projectRootMembership struct {
	exactID   string
	pathIndex int
}

// directRootProgramOwners returns the first declared project that selects each
// target as a tsconfig root. A targeted build supplies this result directly;
// an eager build derives it in batches after all Programs are ready.
func directRootProgramOwners(
	set ProjectSet,
	targets []rslintconfig.DiscoveredLintTarget,
	fsys vfs.FS,
	singleThreaded bool,
) []int {
	owners := make([]int, len(targets))
	for index := range owners {
		owners[index] = -1
	}
	if binding := set.targetBinding; binding != nil &&
		len(binding.targets) == len(targets) &&
		len(binding.owners) == len(targets) {
		matchesPlan := true
		for targetIndex, target := range targets {
			if binding.targets[targetIndex] != target {
				matchesPlan = false
				break
			}
		}
		if matchesPlan {
			copy(owners, binding.owners)
			return owners
		}
	}

	rootPaths := make([]string, 0)
	rootPathIndexByID := make(map[string]int)
	membershipsByProgram := make([][]projectRootMembership, len(set.compilerPrograms))
	for programIndex, program := range set.compilerPrograms {
		if program == nil || program.CommandLine() == nil {
			continue
		}
		for _, rootFileName := range program.CommandLine().FileNames() {
			rootFileName = tspath.NormalizePath(rootFileName)
			exactID := exactPathID(rootFileName)
			rootPathIndex, ok := rootPathIndexByID[exactID]
			if !ok {
				rootPathIndex = len(rootPaths)
				rootPathIndexByID[exactID] = rootPathIndex
				rootPaths = append(rootPaths, rootFileName)
			}
			membershipsByProgram[programIndex] = append(
				membershipsByProgram[programIndex],
				projectRootMembership{exactID: exactID, pathIndex: rootPathIndex},
			)
		}
	}

	targetIndexesByConfig := make(map[string][]int)
	for targetIndex, target := range targets {
		targetIndexesByConfig[target.ConfigDirectory] = append(
			targetIndexesByConfig[target.ConfigDirectory],
			targetIndex,
		)
	}
	orderedProgramsByConfig := make(map[string][]int, len(targetIndexesByConfig))
	exactOwnerPositionByTarget := make([]int, len(targets))
	for index := range exactOwnerPositionByTarget {
		exactOwnerPositionByTarget[index] = -1
	}
	for configDirectory, targetIndexes := range targetIndexesByConfig {
		orderedPrograms := orderedProgramIndexesForConfig(set, configDirectory)
		orderedProgramsByConfig[configDirectory] = orderedPrograms
		targetsByExactID := make(map[string][]int, len(targetIndexes))
		for _, targetIndex := range targetIndexes {
			exactID := exactPathID(targets[targetIndex].Path)
			targetsByExactID[exactID] = append(targetsByExactID[exactID], targetIndex)
		}

		unresolved := len(targetIndexes)
		for position, programIndex := range orderedPrograms {
			for _, membership := range membershipsByProgram[programIndex] {
				for _, targetIndex := range targetsByExactID[membership.exactID] {
					if owners[targetIndex] >= 0 {
						continue
					}
					owners[targetIndex] = programIndex
					exactOwnerPositionByTarget[targetIndex] = position
					unresolved--
				}
			}
			if unresolved == 0 {
				break
			}
		}
	}

	// Only an alias in a project before the exact winner can change ownership.
	// Resolve those root identities once per path, in directory batches.
	canonicalLimitByTarget := make([]int, len(targets))
	needsCanonicalRoots := make([]bool, len(set.compilerPrograms))
	for configDirectory, targetIndexes := range targetIndexesByConfig {
		orderedPrograms := orderedProgramsByConfig[configDirectory]
		for _, targetIndex := range targetIndexes {
			limit := exactOwnerPositionByTarget[targetIndex]
			if limit < 0 {
				limit = len(orderedPrograms)
			}
			canonicalLimitByTarget[targetIndex] = limit
			for position := range limit {
				needsCanonicalRoots[orderedPrograms[position]] = true
			}
		}
	}

	identityResolver := newProgramFileIndex(nil, targets, fsys, singleThreaded)
	identityResolver.initialize()
	canonicalRootPaths := make([]string, 0)
	canonicalRootPathIndexes := make([]int, 0)
	seenRootPathIndexes := make(map[int]struct{})
	canonicalIDsByRootPathIndex := make(map[int]string)
	for programIndex, needed := range needsCanonicalRoots {
		if !needed {
			continue
		}
		for _, membership := range membershipsByProgram[programIndex] {
			if canonicalID, known := identityResolver.canonicalBySourcePath[membership.exactID]; known {
				canonicalIDsByRootPathIndex[membership.pathIndex] = canonicalID
				continue
			}
			if _, seen := seenRootPathIndexes[membership.pathIndex]; seen {
				continue
			}
			seenRootPathIndexes[membership.pathIndex] = struct{}{}
			canonicalRootPathIndexes = append(canonicalRootPathIndexes, membership.pathIndex)
			canonicalRootPaths = append(canonicalRootPaths, rootPaths[membership.pathIndex])
		}
	}
	if fsys == nil {
		for index, rootPath := range canonicalRootPaths {
			canonicalIDsByRootPathIndex[canonicalRootPathIndexes[index]] = exactPathID(rootPath)
		}
	} else {
		canonicalIDs := identityResolver.canonicalSourcePathIDs(canonicalRootPaths)
		for index, canonicalID := range canonicalIDs {
			canonicalIDsByRootPathIndex[canonicalRootPathIndexes[index]] = canonicalID
		}
	}

	canonicalOwnerFound := make([]bool, len(targets))
	for configDirectory, targetIndexes := range targetIndexesByConfig {
		targetsByCanonicalID := make(map[string][]int, len(targetIndexes))
		for _, targetIndex := range targetIndexes {
			target := targets[targetIndex]
			canonicalPath := target.CanonicalPath
			if canonicalPath == "" {
				canonicalPath = target.Path
			}
			canonicalID := exactPathID(canonicalPath)
			targetsByCanonicalID[canonicalID] = append(
				targetsByCanonicalID[canonicalID],
				targetIndex,
			)
		}

		for position, programIndex := range orderedProgramsByConfig[configDirectory] {
			if !needsCanonicalRoots[programIndex] {
				continue
			}
			for _, membership := range membershipsByProgram[programIndex] {
				canonicalID := canonicalIDsByRootPathIndex[membership.pathIndex]
				for _, targetIndex := range targetsByCanonicalID[canonicalID] {
					if canonicalOwnerFound[targetIndex] ||
						position >= canonicalLimitByTarget[targetIndex] {
						continue
					}
					owners[targetIndex] = programIndex
					canonicalOwnerFound[targetIndex] = true
				}
			}
		}
	}
	return owners
}

func bindTargetToProgram(
	binding *LoadResult,
	set ProjectSet,
	programFiles *programFileIndex,
	programIndexes []int,
	programIndex int,
	target rslintconfig.DiscoveredLintTarget,
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
	storeSourceTargetMapping(binding.LintTargetBySourcePath, sourcePath, target.CanonicalPath, target)
	return true
}

func (s *Session) bindTargetsToProjects(
	set ProjectSet,
	plan rslintconfig.LintTargetPlan,
	singleThreaded bool,
) (LoadResult, []rslintconfig.DiscoveredLintTarget) {
	fsys := s.FS()
	binding := LoadResult{
		compilerPrograms:       append([]*compiler.Program(nil), set.compilerPrograms...),
		Programs:               append([]*lintprogram.Program(nil), set.programs...),
		TargetsByProgram:       make([][]string, len(set.compilerPrograms)),
		LintTargetBySourcePath: make(map[string]rslintconfig.DiscoveredLintTarget),
	}

	var unbound []rslintconfig.DiscoveredLintTarget
	programIndexesByConfig := make(map[string][]int)
	programFiles := newProgramFileIndex(set.compilerPrograms, plan.Targets, fsys, singleThreaded)
	directOwners := directRootProgramOwners(set, plan.Targets, fsys, singleThreaded)
	for targetIndex, target := range plan.Targets {
		programIndexes, cached := programIndexesByConfig[target.ConfigDirectory]
		if !cached {
			programIndexes = orderedProgramIndexesForConfig(set, target.ConfigDirectory)
			programIndexesByConfig[target.ConfigDirectory] = programIndexes
		}
		if bindTargetToProgram(
			&binding,
			set,
			programFiles,
			programIndexes,
			directOwners[targetIndex],
			target,
		) {
			continue
		}

		bound := false
		for _, programIndex := range programIndexes {
			if bindTargetToProgram(
				&binding,
				set,
				programFiles,
				programIndexes,
				programIndex,
				target,
			) {
				bound = true
				break
			}
		}
		if !bound {
			unbound = append(unbound, target)
			storeSourceTargetMapping(binding.LintTargetBySourcePath, target.Path, target.CanonicalPath, target)
		}
	}
	return binding, unbound
}
