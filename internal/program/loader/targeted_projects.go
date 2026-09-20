package loader

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"

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

type targetedProjectSlot struct {
	parseOnce sync.Once
	config    *tsoptions.ParsedCommandLine
	parseErr  error

	buildOnce sync.Once
	program   *compiler.Program
	buildErr  error

	lookupOnce sync.Once
	lookupMu   sync.Mutex
	lookup     *programFileIndex
}

type targetedProjectExecution struct {
	session        *Session
	plan           projectPlan
	singleThreaded bool
	slots          []targetedProjectSlot
	targets        []target.File
}

func newTargetedProjectExecution(
	session *Session,
	plan projectPlan,
	targets []target.File,
	singleThreaded bool,
) *targetedProjectExecution {
	return &targetedProjectExecution{
		session:        session,
		plan:           plan,
		targets:        targets,
		singleThreaded: singleThreaded,
		slots:          make([]targetedProjectSlot, len(plan.specs)),
	}
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

func (execution *targetedProjectExecution) parse(index int) (*targetedProjectSlot, error) {
	slot := &execution.slots[index]
	spec := execution.plan.specs[index]
	slot.parseOnce.Do(func() {
		slot.config = spec.parsed
		if slot.config == nil {
			slot.config, slot.parseErr = execution.session.context.parseConfig(
				spec.programCwd,
				spec.tsconfigPath,
			)
		}
		if slot.parseErr == nil && slot.config == nil {
			slot.parseErr = errors.New("no parsed config returned")
		}
	})
	if slot.parseErr != nil {
		return nil, fmt.Errorf("parse TypeScript config %q: %w", spec.tsconfigPath, slot.parseErr)
	}
	return slot, nil
}

func (execution *targetedProjectExecution) build(index int) error {
	slot := &execution.slots[index]
	spec := execution.plan.specs[index]
	slot.buildOnce.Do(func() {
		parsed, err := execution.parse(index)
		if err != nil {
			slot.buildErr = err
			return
		}
		slot.program, slot.buildErr = execution.session.context.createProjectProgramFromParsedConfig(
			execution.singleThreaded,
			spec.programCwd,
			parsed.config,
			spec.sourceReferences,
		)
	})
	if slot.buildErr != nil {
		return fmt.Errorf("create TypeScript Program from %q: %w", spec.tsconfigPath, slot.buildErr)
	}
	return nil
}

func (execution *targetedProjectExecution) containsTarget(
	index int,
	target target.File,
) bool {
	slot := &execution.slots[index]
	if slot.program == nil {
		return false
	}
	slot.lookupOnce.Do(func() {
		slot.lookup = newProgramFileIndex(
			[]*compiler.Program{slot.program}, execution.targets, execution.session.FS(), execution.singleThreaded,
		)
	})
	slot.lookupMu.Lock()
	defer slot.lookupMu.Unlock()
	return slot.lookup.sourceFileForTarget([]int{0}, 0, target) != nil
}

func (execution *targetedProjectExecution) supportsTarget(index int, file target.File) (bool, error) {
	parsed, err := execution.parse(index)
	if err != nil {
		return false, err
	}
	// Service construction enables non-TS roots on a private options copy.
	return execution.plan.specs[index].sourceReferences || projectSupportsTarget(
		parsed.config.CompilerOptions(), file, execution.session.FS().UseCaseSensitiveFileNames(),
	), nil
}

// forEachProject bounds preparation work without letting completion order
// decide ownership or errors. Callers inspect the slots in their stable order.
func (execution *targetedProjectExecution) forEachProject(indexes []int, task func(int)) {
	workerCount := min(runtime.GOMAXPROCS(0), len(indexes))
	if execution.singleThreaded || workerCount <= 1 {
		for _, index := range indexes {
			task(index)
		}
		return
	}

	execution.session.context.enableConcurrentProgramQueries()
	jobs := make(chan int, workerCount)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				task(index)
			}
		}()
	}
	for _, index := range indexes {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
}

func runTargetProjectTasks(
	groups []projectTargetGroup,
	singleThreaded bool,
	task func(group projectTargetGroup) error,
) error {
	if len(groups) == 0 {
		return nil
	}
	workerCount := min(runtime.GOMAXPROCS(0), len(groups))
	if singleThreaded || workerCount <= 1 {
		for _, group := range groups {
			if err := task(group); err != nil {
				return err
			}
		}
		return nil
	}

	errs := make([]error, len(groups))
	jobs := make(chan int, workerCount)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				errs[index] = task(groups[index])
			}
		}()
	}
	for index := range groups {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (execution *targetedProjectExecution) projectSet(
	keep []bool,
	directProjectByTarget []int,
	targets []target.File,
) ProjectSet {
	binding := &projectTargetBinding{
		targets: append([]target.File(nil), targets...),
		owners:  make([]int, len(targets)),
	}
	set := ProjectSet{
		targetBinding: binding,
	}
	for index := range binding.owners {
		binding.owners[index] = -1
	}
	projectSetIndexByPlanIndex := make([]int, len(execution.plan.specs))
	for index := range projectSetIndexByPlanIndex {
		projectSetIndexByPlanIndex[index] = -1
	}
	for index := range execution.plan.specs {
		if index >= len(keep) || !keep[index] {
			continue
		}
		slot := &execution.slots[index]
		if slot.program == nil {
			continue
		}
		set.compilerPrograms = append(set.compilerPrograms, slot.program)
		set.programs = append(set.programs, lintprogram.NewFromCompiler(slot.program))
		set.configOrders = append(set.configOrders, execution.plan.specs[index].configOrders)
		projectSetIndexByPlanIndex[index] = len(set.compilerPrograms) - 1
	}
	for targetIndex, projectIndex := range directProjectByTarget {
		if projectIndex < 0 || targetIndex >= len(targets) {
			continue
		}
		setIndex := projectSetIndexByPlanIndex[projectIndex]
		if setIndex < 0 {
			continue
		}
		binding.owners[targetIndex] = setIndex
	}
	if execution.plan.targetProjects != nil {
		set.targetProjects = make(map[target.File][]int, len(execution.plan.targetProjects))
		remapped := make(map[projectIndexListID][]int)
		for file, candidates := range execution.plan.targetProjects {
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

// parseRootCandidates stops each candidate list after every target has an
// exact root. Earlier physical aliases can still outrank that root; the shared
// root-ranking pass resolves them after these declaration prefixes are ready.
// Targets without an exact root require the complete list.
func (execution *targetedProjectExecution) parseRootCandidates(groups []projectTargetGroup) error {
	if !execution.singleThreaded && len(groups) > 1 {
		execution.session.context.enableConcurrentProgramQueries()
	}
	return runTargetProjectTasks(groups, execution.singleThreaded, func(group projectTargetGroup) error {
		pending := make(map[string]struct{}, len(group.targetIndexes))
		for _, targetIndex := range group.targetIndexes {
			pending[exactPathID(execution.targets[targetIndex].Path)] = struct{}{}
		}
		for _, projectIndex := range group.projectIndexes {
			if len(pending) == 0 {
				break
			}
			parsed, err := execution.parse(projectIndex)
			if err != nil {
				return err
			}
			for _, root := range parsed.config.FileNames() {
				delete(pending, exactPathID(root))
			}
		}
		return nil
	})
}

// executeTargetProjectPlan uses one ownership and validation policy for every
// ordinary lint request. Invocation spelling never changes project selection.
func (s *Session) executeTargetProjectPlan(
	plan projectPlan,
	request ProjectBuildRequest,
) (ProjectSet, error) {
	targetPlan := request.Targets
	singleThreaded := request.SingleThreaded
	execution := newTargetedProjectExecution(s, plan, targetPlan.Files, singleThreaded)

	// Validate every effective candidate before constructing source graphs.
	// Includes are expanded only when root selection reaches the candidate.
	// Stable plan order also preserves the precedence of an earlier metadata
	// failure over a later owner's project-path resolution failure.
	indexes := make([]int, len(plan.specs))
	for index := range indexes {
		indexes[index] = index
	}
	validationErrors := make([]error, len(indexes))
	execution.forEachProject(indexes, func(index int) {
		spec := plan.specs[index]
		if spec.parsed != nil {
			return
		}
		if s.context.metadataFS == nil {
			// The diagnostic cache escape hatch must not add a second read of
			// a mutable VFS. Keep the parsed result instead of probing first.
			_, _ = execution.parse(index)
			validationErrors[index] = execution.slots[index].parseErr
			return
		}
		validationErrors[index] = s.context.validateConfigRead(spec.programCwd, spec.tsconfigPath)
	})
	for index := range indexes {
		if err := validationErrors[index]; err != nil {
			return ProjectSet{}, fmt.Errorf("create TypeScript Program from %q: %w", plan.specs[index].tsconfigPath, err)
		}
	}
	if plan.terminalErr != nil {
		return ProjectSet{}, plan.terminalErr
	}
	if len(plan.specs) == 0 {
		return ProjectSet{targetProjects: plan.targetProjects}, nil
	}

	groups := groupTargetsByProjects(targetPlan.Files, plan.targetProjects, func(owner string) []int {
		return orderedProjectIndexesForConfig(plan, owner)
	})
	if err := execution.parseRootCandidates(groups); err != nil {
		return ProjectSet{}, err
	}
	roots := make([][]string, len(plan.specs))
	for index := range roots {
		if config := execution.slots[index].config; config != nil {
			roots[index] = config.FileNames()
		}
	}
	// Root ranking filters its input groups; source fallback still needs all
	// targets, including ones whose metadata root the compiler cannot admit.
	directProjectByTarget := directRootOwners(roots, targetPlan.Files, plan.targetProjects,
		append([]projectTargetGroup(nil), groups...), s.FS(), singleThreaded)

	directIndexes := make([]int, 0, len(plan.specs))
	seenDirect := make([]bool, len(plan.specs))
	for targetIndex, index := range directProjectByTarget {
		if index < 0 {
			continue
		}
		supported, err := execution.supportsTarget(index, targetPlan.Files[targetIndex])
		if err != nil {
			return ProjectSet{}, err
		}
		if !supported {
			// Keep metadata-root priority: a rejected first root proceeds to
			// ordered source fallback, without constructing its source graph.
			directProjectByTarget[targetIndex] = -1
			continue
		}
		if !seenDirect[index] {
			seenDirect[index] = true
			directIndexes = append(directIndexes, index)
		}
	}
	execution.forEachProject(directIndexes, func(index int) { _ = execution.build(index) })
	for _, index := range directIndexes {
		if err := execution.build(index); err != nil {
			return ProjectSet{}, err
		}
	}

	keep := make([]bool, len(plan.specs))
	for targetIndex, projectIndex := range directProjectByTarget {
		if projectIndex < 0 {
			continue
		}
		if !execution.containsTarget(projectIndex, targetPlan.Files[targetIndex]) {
			if plan.specs[projectIndex].sourceReferences {
				return ProjectSet{}, fmt.Errorf(
					"project root %q from %q was absent from its TypeScript Program",
					targetPlan.Files[targetIndex].Path,
					plan.specs[projectIndex].tsconfigPath,
				)
			}
			// Explicit projects can list roots that their compiler options do
			// not admit. Preserve ordered source fallback before creating gaps.
			directProjectByTarget[targetIndex] = -1
			continue
		}
		keep[projectIndex] = true
	}

	// Different candidate groups can proceed concurrently. Within a group,
	// stop once all remaining targets have an actual source, so an import-only
	// target does not construct every later candidate's dependency graph.
	// Unsupported targets cannot require construction or borrow a Program built
	// for another target; containsTarget applies the same per-file eligibility.
	if !singleThreaded && len(groups) > 1 {
		s.context.enableConcurrentProgramQueries()
	}
	var keepMu sync.Mutex
	err := runTargetProjectTasks(groups, singleThreaded, func(group projectTargetGroup) error {
		pending := make([]int, 0, len(group.targetIndexes))
		for _, targetIndex := range group.targetIndexes {
			if directProjectByTarget[targetIndex] < 0 {
				pending = append(pending, targetIndex)
			}
		}
		for _, projectIndex := range group.projectIndexes {
			if len(pending) == 0 {
				break
			}
			supported := false
			for _, targetIndex := range pending {
				var err error
				supported, err = execution.supportsTarget(projectIndex, targetPlan.Files[targetIndex])
				if err != nil {
					return err
				}
				if supported {
					break
				}
			}
			if !supported {
				continue
			}
			if err := execution.build(projectIndex); err != nil {
				return err
			}
			selected := false
			unresolved := pending[:0]
			for _, targetIndex := range pending {
				if execution.containsTarget(projectIndex, targetPlan.Files[targetIndex]) {
					selected = true
					continue
				}
				// Service discovery promises an actual source. A case-folded
				// metadata match can miss direct ranking; do not turn it into a
				// silent gap by dropping this unbound service Program.
				if plan.specs[projectIndex].sourceReferences {
					return fmt.Errorf(
						"project root %q from %q was absent from its TypeScript Program",
						targetPlan.Files[targetIndex].Path,
						plan.specs[projectIndex].tsconfigPath,
					)
				}
				unresolved = append(unresolved, targetIndex)
			}
			pending = unresolved
			if selected {
				keepMu.Lock()
				keep[projectIndex] = true
				keepMu.Unlock()
			}
		}
		return nil
	})
	if err != nil {
		return ProjectSet{}, err
	}

	return execution.projectSet(keep, directProjectByTarget, targetPlan.Files), nil
}
