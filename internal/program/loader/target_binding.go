package loader

import (
	"errors"
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/program/projectselection"
)

type plannedTargetRef struct {
	index  int
	target rslintconfig.PlannedLintTarget
}

func plannedTargetsEqual(left, right []rslintconfig.PlannedLintTarget) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Target != right[index].Target ||
			left[index].MatchPath != right[index].MatchPath ||
			left[index].Effective != right[index].Effective ||
			!slices.Equal(left[index].ProjectPaths, right[index].ProjectPaths) {
			return false
		}
	}
	return true
}

func targetBindingOwners(set ProjectSet, plan rslintconfig.LintProjectPlan) ([]int, error) {
	binding := set.targetBinding
	if binding == nil || !plannedTargetsEqual(binding.targets, plan.Targets) ||
		len(binding.owners) != len(plan.Targets) {
		return nil, errors.New("project binding does not match the lint project plan")
	}
	for targetIndex, target := range plan.Targets {
		owner := binding.owners[targetIndex]
		if owner < projectselection.NoProject || owner >= len(set.compilerPrograms) {
			return nil, fmt.Errorf(
				"project binding contains invalid owner %d for %q",
				owner,
				target.Target.Path,
			)
		}
	}
	return append([]int(nil), binding.owners...), nil
}

func bindTargetToProgram(
	binding *LoadResult,
	set ProjectSet,
	programFiles *programFileIndex,
	programIndex int,
	targetIndex int,
	target rslintconfig.PlannedLintTarget,
) bool {
	if programIndex < 0 || programIndex >= len(set.compilerPrograms) {
		return false
	}
	sourceFile := exactProgramSourceFile(set.compilerPrograms[programIndex], target.Target.Path)
	if sourceFile == nil {
		canonicalPath := programFiles.canonicalPath(target.Target.Path, target.Target.CanonicalPath)
		sourceFile = programFiles.sourceFile([]int{programIndex}, programIndex, canonicalPath)
	}
	if sourceFile == nil {
		return false
	}
	sourcePath := sourceFile.FileName()
	binding.TargetsByProgram[programIndex] = append(binding.TargetsByProgram[programIndex], sourcePath)
	storeTargetIndexMapping(binding.targetIndexBySourcePath, sourcePath, target.Target.CanonicalPath, targetIndex)
	if tspath.NormalizePath(sourcePath) != target.Target.Path {
		storeSourcePathMapping(
			binding.TargetPathBySourcePath,
			sourcePath,
			target.Target.CanonicalPath,
			target.Target.Path,
		)
	}
	return true
}

func (s *Session) bindTargetsToProjects(
	set ProjectSet,
	plan rslintconfig.LintProjectPlan,
	singleThreaded bool,
) (LoadResult, []plannedTargetRef, error) {
	discoveredTargets := make([]rslintconfig.DiscoveredLintTarget, len(plan.Targets))
	for index, target := range plan.Targets {
		discoveredTargets[index] = target.Target
	}
	var typeCheckPrograms []*lintprogram.Program
	if set.typeCheckPrograms != nil {
		typeCheckPrograms = append([]*lintprogram.Program{}, set.typeCheckPrograms...)
	}
	binding := LoadResult{
		compilerPrograms:        append([]*compiler.Program(nil), set.compilerPrograms...),
		Programs:                append([]*lintprogram.Program(nil), set.programs...),
		TypeCheckPrograms:       typeCheckPrograms,
		TargetsByProgram:        make([][]string, len(set.compilerPrograms)),
		TargetPathBySourcePath:  make(map[string]string),
		targetIndexBySourcePath: make(map[string]int),
	}
	owners, err := targetBindingOwners(set, plan)
	if err != nil {
		return LoadResult{}, nil, err
	}
	var programFiles *programFileIndex
	if set.pathIdentities != nil {
		programFiles = newProgramFileIndexWithResolver(
			set.compilerPrograms,
			discoveredTargets,
			s.FS(),
			set.pathIdentities,
		)
	} else {
		programFiles = newProgramFileIndex(
			set.compilerPrograms,
			discoveredTargets,
			s.FS(),
			singleThreaded,
		)
	}
	var unbound []plannedTargetRef
	for targetIndex, target := range plan.Targets {
		owner := owners[targetIndex]
		if bindTargetToProgram(
			&binding,
			set,
			programFiles,
			owner,
			targetIndex,
			target,
		) {
			continue
		}
		if owner >= 0 {
			return LoadResult{}, nil, fmt.Errorf(
				"selected project %d did not contain lint target %q during binding",
				owner,
				target.Target.Path,
			)
		}
		unbound = append(unbound, plannedTargetRef{index: targetIndex, target: target})
	}
	return binding, unbound, nil
}
