package loader

import (
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type configOrders map[string]int

// ProjectScope preserves the construction choices made by CLI/API callers.
type ProjectScope uint8

const (
	AllDeclared ProjectScope = iota
	ActiveOwners
	Targeted
)

type ProjectBuildRequest struct {
	Configs        map[string]rslintconfig.RslintConfig
	Targets        target.Plan
	Policies       map[target.File]rslintconfig.ProjectPolicy
	Scope          ProjectScope
	SingleThreaded bool
}

// ProjectSet is the stable, deduplicated set of configured project generations
// built for one load pass. Its compiler backing and config associations are
// private assembly details; consumers use Programs after loading or binding.
type ProjectSet struct {
	compilerPrograms []*compiler.Program
	programs         []*lintprogram.Program
	configOrders     []configOrders
	targetBinding    *projectTargetBinding
	// Missing entries use ordinary owner declarations. Present entries are an
	// ordered candidate list; an empty list explicitly leaves a target unbound.
	targetProjects map[target.File][]int
}

// Programs returns the configured rslint Programs in stable project-plan order.
// The slice is read-only and remains owned by the ProjectSet.
func (projects ProjectSet) Programs() []*lintprogram.Program {
	return projects.programs
}

func (projects ProjectSet) Len() int {
	return len(projects.programs)
}

type projectSpec struct {
	tsconfigPath string
	programCwd   string
	configOrders configOrders
	// Discovery can supply parsed roots before construction. Service projects
	// use source references; ordinary declarations keep their existing mode.
	parsed           *tsoptions.ParsedCommandLine
	sourceReferences bool
}

type projectPlan struct {
	specs       []projectSpec
	terminalErr error
	// These indexes refer to specs until execution compacts the retained set.
	targetProjects map[target.File][]int
}

func exactPathID(filePath string) string {
	return rslintconfig.ExactPathID(filePath)
}

func buildProjectPlan(request ProjectBuildRequest, fsys vfs.FS) projectPlan {
	configMap := request.Configs
	if request.Scope != AllDeclared {
		configMap = configsForActiveOwners(configMap, request.Targets)
	}
	if len(configMap) == 0 {
		return projectPlan{}
	}

	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	plan := projectPlan{}
	programByTsconfig := make(map[string]int)
	addPaths := func(owner string, paths []string, ordinary bool) []int {
		var indexes []int
		if !ordinary {
			indexes = make([]int, 0, len(paths))
		}
		ownerID := exactPathID(owner)
		for order, path := range paths {
			path = tspath.NormalizePath(path)
			pathID := exactPathID(path)
			index, exists := programByTsconfig[pathID]
			if !exists {
				index = len(plan.specs)
				programByTsconfig[pathID] = index
				plan.specs = append(plan.specs, projectSpec{
					tsconfigPath: path, programCwd: tspath.GetDirectoryPath(path), configOrders: configOrders{},
				})
			}
			if ordinary {
				if _, associated := plan.specs[index].configOrders[ownerID]; !associated {
					plan.specs[index].configOrders[ownerID] = order
				}
			}
			if !ordinary {
				indexes = append(indexes, index)
			}
		}
		return indexes
	}
	// No new project options need no per-target preparation. In particular,
	// AllDeclared keeps its original path when no target discovery was needed.
	if len(request.Policies) == 0 {
		for _, owner := range configDirs {
			paths, err := rslintconfig.ResolveTsConfigPaths(configMap[owner], owner, fsys)
			if err != nil {
				plan.terminalErr = fmt.Errorf("resolve tsconfigs for %q: %w", owner, err)
				return plan
			}
			addPaths(owner, paths, true)
		}
		return plan
	}

	type pathContext struct {
		root            string
		defaultDisabled bool
	}
	targetsByOwner := make(map[string][]target.File)
	for _, file := range request.Targets.Files {
		targetsByOwner[file.ConfigDirectory] = append(targetsByOwner[file.ConfigDirectory], file)
	}
	plan.targetProjects = make(map[target.File][]int, len(request.Policies))
	for _, configDir := range configDirs {
		contexts := make(map[pathContext][]target.File)
		for _, file := range targetsByOwner[configDir] {
			policy := request.Policies[file]
			key := pathContext{root: policy.TSConfigRootDirOverride, defaultDisabled: policy.DefaultProjectDisabled}
			if policy.ServiceRootDirectory != "" || policy.ProjectDisabled {
				plan.targetProjects[file] = nil
				if request.Scope != AllDeclared {
					continue
				}
				// Type checking retains raw explicit declarations at this target's
				// root, but a service/clear target never requests an implicit project.
				key.defaultDisabled = true
				if _, exists := contexts[key]; !exists {
					contexts[key] = nil
				}
				continue
			}
			contexts[key] = append(contexts[key], file)
		}
		if len(targetsByOwner[configDir]) == 0 && request.Scope == AllDeclared {
			contexts[pathContext{}] = nil
		}
		keys := make([]pathContext, 0, len(contexts))
		for key := range contexts {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(left, right int) bool {
			if keys[left].root != keys[right].root {
				return keys[left].root < keys[right].root
			}
			return !keys[left].defaultDisabled && keys[right].defaultDisabled
		})
		for _, key := range keys {
			paths, err := rslintconfig.ResolveTsConfigPathsWithPolicy(configMap[configDir], configDir, fsys, rslintconfig.ProjectPolicy{
				TSConfigRootDirOverride: key.root, DefaultProjectDisabled: key.defaultDisabled,
			})
			if err != nil {
				plan.terminalErr = fmt.Errorf("resolve tsconfigs for %q: %w", configDir, err)
				return plan
			}
			ordinary := key == (pathContext{})
			indexes := addPaths(configDir, paths, ordinary)
			if !ordinary {
				for _, file := range contexts[key] {
					plan.targetProjects[file] = indexes
				}
			}
		}
	}
	return plan
}

func (s *Session) executeProjectPlan(plan projectPlan, singleThreaded bool) (ProjectSet, error) {
	if err := s.validate(); err != nil {
		return ProjectSet{}, err
	}
	if len(plan.specs) == 0 {
		if plan.terminalErr != nil {
			return ProjectSet{}, plan.terminalErr
		}
		return ProjectSet{targetProjects: plan.targetProjects}, nil
	}
	compilerPrograms := make([]*compiler.Program, len(plan.specs))
	errs := make([]error, len(plan.specs))
	workerCount := min(runtime.GOMAXPROCS(0), len(plan.specs))
	parallel := !singleThreaded && workerCount > 1
	if parallel {
		s.context.enableConcurrentProgramQueries()
	}
	build := func(index int) {
		spec := plan.specs[index]
		if spec.parsed != nil {
			compilerPrograms[index], errs[index] = s.context.createProjectProgramFromParsedConfig(
				singleThreaded, spec.programCwd, spec.parsed, spec.sourceReferences,
			)
			return
		}
		compilerPrograms[index], errs[index] = s.context.createProjectProgram(
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
			return ProjectSet{}, fmt.Errorf(
				"create TypeScript Program from %q: %w",
				plan.specs[index].tsconfigPath,
				err,
			)
		}
	}
	if plan.terminalErr != nil {
		return ProjectSet{}, plan.terminalErr
	}

	orders := make([]configOrders, len(plan.specs))
	for index := range plan.specs {
		orders[index] = plan.specs[index].configOrders
	}
	return ProjectSet{
		compilerPrograms: compilerPrograms,
		programs:         lintprogram.NewFromCompilers(compilerPrograms),
		configOrders:     orders,
		targetProjects:   plan.targetProjects,
	}, nil
}

// BuildProjects prepares explicit and discovered configs in one project plan,
// then applies the caller's existing construction scope. LoadCLI/LoadAPI bind
// targets to that plan once; discovery never creates a separate Program set.
func (s *Session) BuildProjects(request ProjectBuildRequest) (ProjectSet, error) {
	if err := s.validate(); err != nil {
		return ProjectSet{}, err
	}
	plan := buildProjectPlan(request, s.FS())
	if plan.terminalErr == nil {
		if err := s.discoverServiceProjects(&plan, request.Targets, request.Policies); err != nil {
			return ProjectSet{}, err
		}
	}
	var set ProjectSet
	var err error
	if request.Scope == Targeted {
		set, err = s.executeTargetProjectPlan(plan, request.Targets, request.SingleThreaded)
	} else {
		set, err = s.executeProjectPlan(plan, request.SingleThreaded)
	}
	if err != nil {
		return ProjectSet{}, err
	}
	if request.Scope != Targeted {
		// Focused execution already validates every selected direct root. Eager
		// modes need the same check before publishing service selections, even
		// when type-check-only will never enter the lint binding phase.
		for _, file := range request.Targets.Files {
			if request.Policies[file].ServiceRootDirectory == "" {
				continue
			}
			indexes := set.targetProjects[file]
			if len(indexes) == 0 {
				continue
			}
			program := set.compilerPrograms[indexes[0]]
			if utils.NewProgramSourceLookup(program, s.FS()).SourceFileForTarget(file.Path, file.CanonicalPath) == nil {
				return ProjectSet{}, fmt.Errorf("project root %q from %q was absent from its TypeScript Program", file.Path, program.CommandLine().ConfigName())
			}
		}
	}
	return set, nil
}
