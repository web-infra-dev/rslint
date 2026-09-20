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
)

type configOrders map[string]int

// ProjectScope separates target-driven lint from program-wide type checking.
// Ordinary CLI and API calls share the same LintTargets contract.
type ProjectScope uint8

const (
	AllDeclared ProjectScope = iota
	LintTargets
)

type ProjectBuildRequest struct {
	// Configs retains the program-wide declaration range for AllDeclared.
	// Ordinary lint candidates come exclusively from Policies.
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

type projectPolicyID struct {
	options       rslintconfig.ProjectPolicy
	baseDirectory string
	firstPattern  *string
	patternCount  int
}

func projectPolicyIdentity(policy rslintconfig.ProjectPolicy) projectPolicyID {
	key := projectPolicyID{options: policy}
	key.options.ExplicitProject = nil
	if declaration := policy.ExplicitProject; declaration != nil {
		key.baseDirectory = declaration.BaseDirectory
		key.patternCount = len(declaration.Patterns)
		if key.patternCount > 0 {
			key.firstPattern = &declaration.Patterns[0]
		}
	}
	return key
}

func buildProjectPlan(request ProjectBuildRequest, fsys vfs.FS) projectPlan {
	plan := projectPlan{}
	if request.Scope == AllDeclared {
		plan = buildDeclaredProjectPlan(request, fsys)
		if plan.terminalErr != nil {
			return plan
		}
	}
	if len(request.Targets.Files) == 0 {
		return plan
	}
	// The complete type-check set and a lint target's candidates are distinct.
	// Reuse an existing explicit project when present, but always publish an
	// override (including empty) so unrelated targets cannot borrow its types.
	programByTsconfig := make(map[string]int, len(plan.specs))
	for index, spec := range plan.specs {
		programByTsconfig[exactPathID(spec.tsconfigPath)] = index
	}
	plan.targetProjects = make(map[target.File][]int, len(request.Targets.Files))
	// Different rule matches can wrap the same immutable authored project
	// patterns in separate declarations. Reuse their expansion and candidate
	// slice within this request, while retaining every path/policy context.
	indexesByPolicy := make(map[projectPolicyID][]int)
	for _, file := range request.Targets.Files {
		policy, resolved := request.Policies[file]
		if !resolved {
			plan.terminalErr = fmt.Errorf("missing effective project policy for %q", file.Path)
			return plan
		}
		key := projectPolicyIdentity(policy)
		indexes, cached := indexesByPolicy[key]
		if !cached {
			paths, err := rslintconfig.ResolveProjectPaths(policy, fsys)
			if err != nil {
				plan.terminalErr = fmt.Errorf("resolve tsconfigs for %q: %w", file.Path, err)
				return plan
			}
			for _, path := range paths {
				pathID := exactPathID(path)
				index, exists := programByTsconfig[pathID]
				if !exists {
					index = len(plan.specs)
					programByTsconfig[pathID] = index
					plan.specs = append(plan.specs, projectSpec{
						tsconfigPath: path, programCwd: tspath.GetDirectoryPath(path),
					})
				}
				indexes = append(indexes, index)
			}
			indexesByPolicy[key] = indexes
		}
		plan.targetProjects[file] = indexes
	}
	return plan
}

// buildDeclaredProjectPlan preserves the program-wide checking range. Raw
// declarations and its historical default project are intentionally independent
// of the effective per-file candidates added by buildProjectPlan.
func buildDeclaredProjectPlan(request ProjectBuildRequest, fsys vfs.FS) projectPlan {
	configMap := request.Configs
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
	addPaths := func(owner string, paths []string) {
		ownerID := exactPathID(owner)
		for order, path := range paths {
			pathID := exactPathID(path)
			index, exists := programByTsconfig[pathID]
			if !exists {
				index = len(plan.specs)
				programByTsconfig[pathID] = index
				plan.specs = append(plan.specs, projectSpec{
					tsconfigPath: path, programCwd: tspath.GetDirectoryPath(path), configOrders: configOrders{},
				})
			}
			if _, associated := plan.specs[index].configOrders[ownerID]; !associated {
				plan.specs[index].configOrders[ownerID] = order
			}
		}
	}
	type pathContext struct {
		root            string
		defaultDisabled bool
	}
	targetsByOwner := make(map[string][]target.File)
	for _, file := range request.Targets.Files {
		targetsByOwner[file.ConfigDirectory] = append(targetsByOwner[file.ConfigDirectory], file)
	}
	for _, configDir := range configDirs {
		contexts := make(map[pathContext]struct{})
		for _, file := range targetsByOwner[configDir] {
			policy := request.Policies[file]
			key := pathContext{root: policy.TSConfigRootDirOverride, defaultDisabled: policy.DefaultProjectDisabled}
			if policy.ServiceRootDirectory != "" || policy.ProjectDisabled {
				// Type checking retains raw explicit declarations at this target's
				// root, but a service/clear target never requests an implicit project.
				key.defaultDisabled = true
			}
			contexts[key] = struct{}{}
		}
		if len(contexts) == 0 {
			contexts[pathContext{}] = struct{}{}
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
			addPaths(configDir, paths)
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
// then applies lint or type-check scope. LoadCLI/LoadAPI bind
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
	if request.Scope == AllDeclared {
		set, err = s.executeProjectPlan(plan, request.SingleThreaded)
	} else {
		set, err = s.executeTargetProjectPlan(plan, request)
	}
	if err != nil {
		return ProjectSet{}, err
	}
	if request.Scope == AllDeclared {
		// Target-driven lint already validates every service source. Full
		// project checking needs the same check before publishing selections, even
		// when type-check-only will never enter the lint binding phase.
		programFiles := newProgramFileIndex(set.compilerPrograms, request.Targets.Files, s.FS(), request.SingleThreaded)
		for _, file := range request.Targets.Files {
			if request.Policies[file].ServiceRootDirectory == "" {
				continue
			}
			indexes := set.targetProjects[file]
			if len(indexes) == 0 {
				continue
			}
			program := set.compilerPrograms[indexes[0]]
			if programFiles.sourceFileForTarget(indexes, indexes[0], file) == nil {
				return ProjectSet{}, fmt.Errorf("project root %q from %q was absent from its TypeScript Program", file.Path, program.CommandLine().ConfigName())
			}
		}
	}
	return set, nil
}
