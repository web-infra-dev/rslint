package loader

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
)

type projectTargetGroup struct {
	owner          string
	targetIndexes  []int
	projectIndexes []int
}

type projectIndexListID struct {
	first  *int
	length int
}

func projectIndexListIdentity(indexes []int) projectIndexListID {
	if len(indexes) == 0 {
		return projectIndexListID{}
	}
	return projectIndexListID{first: &indexes[0], length: len(indexes)}
}

func projectIndexesForTarget(
	file target.File,
	overrides map[target.File][]int,
	indexesByOwner map[string][]int,
	forOwner func(string) []int,
) []int {
	if indexes, overridden := overrides[file]; overridden {
		return indexes
	}
	indexes, cached := indexesByOwner[file.ConfigDirectory]
	if !cached {
		indexes = forOwner(file.ConfigDirectory)
		indexesByOwner[file.ConfigDirectory] = indexes
	}
	return indexes
}

func groupTargetsByProjects(
	targets []target.File,
	overrides map[target.File][]int,
	forOwner func(string) []int,
) []projectTargetGroup {
	type groupKey struct {
		owner string
		list  projectIndexListID
	}
	groups := make([]projectTargetGroup, 0)
	groupIndexes := make(map[groupKey]int)
	indexesByOwner := make(map[string][]int)
	for targetIndex, file := range targets {
		indexes := projectIndexesForTarget(file, overrides, indexesByOwner, forOwner)
		if len(indexes) == 0 {
			continue
		}
		// A path context shares one immutable candidate slice. Its storage
		// identity only batches equal requests; separately allocated lists can
		// safely remain separate groups even when their contents are equal.
		key := groupKey{owner: file.ConfigDirectory, list: projectIndexListIdentity(indexes)}
		groupIndex, found := groupIndexes[key]
		if !found {
			groupIndex = len(groups)
			groupIndexes[key] = groupIndex
			groups = append(groups, projectTargetGroup{owner: file.ConfigDirectory, projectIndexes: indexes})
		}
		groups[groupIndex].targetIndexes = append(groups[groupIndex].targetIndexes, targetIndex)
	}
	sort.Slice(groups, func(left, right int) bool {
		if groups[left].owner != groups[right].owner {
			return groups[left].owner < groups[right].owner
		}
		return groups[left].targetIndexes[0] < groups[right].targetIndexes[0]
	})
	return groups
}

type projectRootMembership struct {
	exactID   string
	pathIndex int
}

// directRootProgramOwners returns the first declared project that selects each
// target as a tsconfig root. A targeted build supplies this result directly;
// an eager build derives it in batches after all Programs are ready.
func directRootProgramOwners(
	set ProjectSet,
	targets []target.File,
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

	groups := groupTargetsByProjects(targets, set.targetProjects, func(owner string) []int {
		return orderedProgramIndexesForConfig(set, owner)
	})
	var rankedPrograms []bool
	if len(set.targetProjects) > 0 {
		rankedPrograms = make([]bool, len(set.compilerPrograms))
		rankedGroups := groups[:0]
		for _, group := range groups {
			_, overridden := set.targetProjects[targets[group.targetIndexes[0]]]
			if overridden && len(group.projectIndexes) == 1 {
				// A single candidate has no root/import ranking to
				// decide. Binding still checks its actual source and identity.
				for _, targetIndex := range group.targetIndexes {
					owners[targetIndex] = group.projectIndexes[0]
				}
				continue
			}
			rankedGroups = append(rankedGroups, group)
			for _, index := range group.projectIndexes {
				rankedPrograms[index] = true
			}
		}
		groups = rankedGroups
		if len(groups) == 0 {
			return owners
		}
	}

	rootPaths := make([]string, 0)
	rootPathIndexByID := make(map[string]int)
	membershipsByProgram := make([][]projectRootMembership, len(set.compilerPrograms))
	for programIndex, program := range set.compilerPrograms {
		if rankedPrograms != nil && !rankedPrograms[programIndex] {
			continue
		}
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

	exactOwnerPositionByTarget := make([]int, len(targets))
	for index := range exactOwnerPositionByTarget {
		exactOwnerPositionByTarget[index] = -1
	}
	for _, group := range groups {
		targetIndexes := group.targetIndexes
		orderedPrograms := group.projectIndexes
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
	for _, group := range groups {
		orderedPrograms := group.projectIndexes
		for _, targetIndex := range group.targetIndexes {
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
	for _, group := range groups {
		targetIndexes := group.targetIndexes
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

		for position, programIndex := range group.projectIndexes {
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
	target target.File,
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
	plan target.Plan,
	singleThreaded bool,
) (LoadResult, []target.File) {
	fsys := s.FS()
	binding := LoadResult{
		compilerPrograms:       append([]*compiler.Program(nil), set.compilerPrograms...),
		Programs:               append([]*lintprogram.Program(nil), set.programs...),
		TargetsByProgram:       make([][]string, len(set.compilerPrograms)),
		LintTargetBySourcePath: make(map[string]target.File),
	}

	var unbound []target.File
	programIndexesByConfig := make(map[string][]int)
	programFiles := newProgramFileIndex(set.compilerPrograms, plan.Files, fsys, singleThreaded)
	directOwners := directRootProgramOwners(set, plan.Files, fsys, singleThreaded)
	for targetIndex, target := range plan.Files {
		programIndexes := projectIndexesForTarget(target, set.targetProjects, programIndexesByConfig, func(owner string) []int {
			return orderedProgramIndexesForConfig(set, owner)
		})
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
