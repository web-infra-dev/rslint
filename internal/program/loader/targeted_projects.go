package loader

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type projectTargetBinding struct {
	targets []target.File
	owners  []int
}

func (c *buildContext) createProjectProgramFromParsedConfig(
	singleThreaded bool,
	cwd string,
	config *tsoptions.ParsedCommandLine,
	sourceReferences bool,
) (*compiler.Program, error) {
	if sourceReferences {
		return utils.CreateProgramFromParsedConfigLenientWithProjectReferences(
			singleThreaded, config, c.newCompilerHostWithCache(cwd),
		)
	}
	return utils.CreateProgramFromParsedConfigLenient(
		singleThreaded,
		config,
		c.newCompilerHostWithCache(cwd),
	)
}

// selectedProjectSet retains only projects that provide a selected source.
// Binding reuses the selector's decision instead of ranking candidates again.
func selectedProjectSet(plan projectPlan, targets []target.File, selected []ProjectSourceSelection) ProjectSet {
	binding := &projectTargetBinding{
		targets: append([]target.File(nil), targets...),
		owners:  make([]int, len(targets)),
	}
	set := ProjectSet{targetBinding: binding}
	programs := make([]*compiler.Program, len(plan.specs))
	for _, result := range selected {
		if result.CandidateIndex >= 0 {
			programs[result.CandidateIndex] = result.Program
		}
	}
	projectSetIndexByPlanIndex := make([]int, len(plan.specs))
	for index, program := range programs {
		projectSetIndexByPlanIndex[index] = -1
		if program == nil {
			continue
		}
		set.compilerPrograms = append(set.compilerPrograms, program)
		set.programs = append(set.programs, lintprogram.NewFromCompiler(program))
		set.configOrders = append(set.configOrders, plan.specs[index].configOrders)
		projectSetIndexByPlanIndex[index] = len(set.compilerPrograms) - 1
	}
	for targetIndex, result := range selected {
		binding.owners[targetIndex] = -1
		if result.CandidateIndex >= 0 {
			binding.owners[targetIndex] = projectSetIndexByPlanIndex[result.CandidateIndex]
		}
	}
	if plan.targetProjects != nil {
		set.targetProjects = make(map[target.File][]int, len(plan.targetProjects))
		remapped := make(map[projectIndexListID][]int)
		for file, candidates := range plan.targetProjects {
			key := projectIndexListIdentity(candidates)
			indexes, exists := remapped[key]
			if !exists {
				for _, index := range candidates {
					if retained := projectSetIndexByPlanIndex[index]; retained >= 0 {
						indexes = append(indexes, retained)
					}
				}
				remapped[key] = indexes
			}
			set.targetProjects[file] = indexes
		}
	}
	return set
}

func orderedProjectIndexesForConfig(plan projectPlan, configDir string) []int {
	configDirID := exactPathID(configDir)
	indexes := make([]int, 0, len(plan.specs))
	for index := range plan.specs {
		if _, ok := plan.specs[index].configOrders[configDirID]; ok {
			indexes = append(indexes, index)
		}
	}
	sort.SliceStable(indexes, func(left, right int) bool {
		leftOrder := plan.specs[indexes[left]].configOrders[configDirID]
		rightOrder := plan.specs[indexes[right]].configOrders[configDirID]
		if leftOrder != rightOrder {
			return leftOrder < rightOrder
		}
		return indexes[left] < indexes[right]
	})
	return indexes
}

// executeTargetProjectPlan adapts invocation-scoped construction to the same
// target selection used by editor sessions. It never adds lint targets.
func (s *Session) executeTargetProjectPlan(plan projectPlan, request ProjectBuildRequest) (ProjectSet, error) {
	if plan.terminalErr != nil {
		return ProjectSet{}, plan.terminalErr
	}
	if len(plan.specs) == 0 {
		return ProjectSet{targetProjects: plan.targetProjects}, nil
	}
	candidates := make([]ProjectCandidate, len(plan.specs))
	for index, spec := range plan.specs {
		candidates[index] = ProjectCandidate{
			ConfigPath:       spec.tsconfigPath,
			SourceReferences: spec.sourceReferences,
		}
	}
	candidateIndexes := make([][]int, len(request.Targets.Files))
	indexesByOwner := make(map[string][]int)
	for index, file := range request.Targets.Files {
		candidateIndexes[index] = projectIndexesForTarget(file, plan.targetProjects, indexesByOwner, func(owner string) []int {
			return orderedProjectIndexesForConfig(plan, owner)
		})
	}
	if !request.SingleThreaded {
		s.context.enableConcurrentProgramQueries()
	}
	selected, err := SelectProjectSources(ProjectSelectionRequest{
		Targets:          request.Targets.Files,
		CandidateIndexes: candidateIndexes,
		Candidates:       candidates,
		FS:               s.FS(),
		SingleThreaded:   request.SingleThreaded,
		Metadata: func(index int) (*tsoptions.ParsedCommandLine, error) {
			spec := plan.specs[index]
			if spec.parsed != nil {
				return spec.parsed, nil
			}
			return s.context.parseConfig(spec.programCwd, spec.tsconfigPath)
		},
		Program: func(index int, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
			spec := plan.specs[index]
			return s.context.createProjectProgramFromParsedConfig(request.SingleThreaded, spec.programCwd, parsed, spec.sourceReferences)
		},
	})
	if err != nil {
		return ProjectSet{}, err
	}
	return selectedProjectSet(plan, request.Targets.Files, selected), nil
}
