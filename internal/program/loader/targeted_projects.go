package loader

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
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

	rootFilesOnce sync.Once
	rootFiles     *lintprogram.RootFileIndex

	buildOnce sync.Once
	program   *compiler.Program
	buildErr  error

	lookupOnce sync.Once
	lookupMu   sync.Mutex
	lookup     *utils.ProgramSourceLookup
}

type targetedProjectExecution struct {
	session        *Session
	plan           projectPlan
	singleThreaded bool
	slots          []targetedProjectSlot
}

type targetedProjectBuildQueue struct {
	execution *targetedProjectExecution
	parallel  bool
	workersN  int
	jobs      chan int
	workers   sync.WaitGroup
	mu        sync.Mutex
	enqueued  []bool
	errs      []error
}

func newTargetedProjectBuildQueue(execution *targetedProjectExecution) *targetedProjectBuildQueue {
	queue := &targetedProjectBuildQueue{
		execution: execution,
		enqueued:  make([]bool, len(execution.plan.specs)),
		errs:      make([]error, len(execution.plan.specs)),
	}
	workerCount := min(runtime.GOMAXPROCS(0), len(execution.plan.specs))
	queue.parallel = !execution.singleThreaded && workerCount > 1
	queue.workersN = workerCount
	return queue
}

func (queue *targetedProjectBuildQueue) enqueue(index int) error {
	queue.mu.Lock()
	if queue.enqueued[index] {
		queue.mu.Unlock()
		return nil
	}
	queue.enqueued[index] = true
	if queue.parallel && queue.jobs == nil {
		queue.execution.session.context.enableConcurrentProgramQueries()
		queue.jobs = make(chan int, len(queue.execution.plan.specs))
		queue.workers.Add(queue.workersN)
		for range queue.workersN {
			go func() {
				defer queue.workers.Done()
				for index := range queue.jobs {
					queue.errs[index] = queue.execution.build(index)
				}
			}()
		}
	}
	jobs := queue.jobs
	queue.mu.Unlock()
	if !queue.parallel {
		queue.errs[index] = queue.execution.build(index)
		return nil
	}
	jobs <- index
	return nil
}

func (queue *targetedProjectBuildQueue) wait() {
	if queue.parallel && queue.jobs != nil {
		close(queue.jobs)
		queue.workers.Wait()
	}
}

func newTargetedProjectExecution(
	session *Session,
	plan projectPlan,
	singleThreaded bool,
) *targetedProjectExecution {
	return &targetedProjectExecution{
		session:        session,
		plan:           plan,
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

// rootFileIndex constructs membership indexes only when queried. Config
// validation and program construction do not need this derived lookup.
func (execution *targetedProjectExecution) rootFileIndex(index int) (*lintprogram.RootFileIndex, error) {
	slot, err := execution.parse(index)
	if err != nil {
		return nil, err
	}
	slot.rootFilesOnce.Do(func() {
		slot.rootFiles = lintprogram.NewRootFileIndex(slot.config.FileNames(), execution.session.FS())
	})
	return slot.rootFiles, nil
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
		slot.lookup = utils.NewProgramSourceLookup(slot.program, execution.session.FS())
	})
	slot.lookupMu.Lock()
	defer slot.lookupMu.Unlock()
	return slot.lookup.SourceFileForTarget(target.Path, target.CanonicalPath) != nil
}

func (execution *targetedProjectExecution) parseConcurrent(indexes []int) {
	if len(indexes) == 0 {
		return
	}
	workerCount := min(runtime.GOMAXPROCS(0), len(indexes))
	if execution.singleThreaded || workerCount <= 1 {
		for _, index := range indexes {
			_, _ = execution.parse(index)
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
				// Speculation must not make an otherwise unreachable malformed
				// config observable. The ordered consumer below reports an error
				// only if ownership resolution actually reaches this slot.
				_, _ = execution.parse(index)
			}
		}()
	}
	for _, index := range indexes {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
}

func (execution *targetedProjectExecution) predictedProjectPosition(
	orderedProjectIndexes []int,
	target target.File,
) int {
	useCaseSensitive := true
	if fsys := execution.session.FS(); fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	options := tspath.ComparePathsOptions{UseCaseSensitiveFileNames: useCaseSensitive}
	bestPosition := -1
	bestDirectoryLength := -1
	for position, projectIndex := range orderedProjectIndexes {
		directory := execution.plan.specs[projectIndex].programCwd
		if !tspath.ContainsPath(directory, target.Path, options) {
			continue
		}
		if len(directory) > bestDirectoryLength {
			bestPosition = position
			bestDirectoryLength = len(directory)
		}
	}
	return bestPosition
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

// executeTargetProjectPlan selects parsed roots over each target's ordered
// candidates. Unmatched targets never construct projects to probe imports.
// All contexts share one execution and one slot per declared tsconfig.
func (s *Session) executeTargetProjectPlan(
	plan projectPlan,
	request ProjectBuildRequest,
) (ProjectSet, error) {
	targetPlan := request.Targets
	singleThreaded := request.SingleThreaded
	execution := newTargetedProjectExecution(s, plan, singleThreaded)
	if request.Scope == ActiveOwners {
		// Broad lint still validates every active declaration, in plan order,
		// before reporting a later path-resolution failure. Parsing metadata
		// does not require loading any project's sources or imports.
		indexes := make([]int, len(plan.specs))
		for index := range indexes {
			indexes[index] = index
		}
		execution.parseConcurrent(indexes)
		for index := range indexes {
			if err := execution.slots[index].parseErr; err != nil {
				return ProjectSet{}, fmt.Errorf("create TypeScript Program from %q: %w", plan.specs[index].tsconfigPath, err)
			}
		}
	}
	if plan.terminalErr != nil {
		return ProjectSet{}, plan.terminalErr
	}
	if len(plan.specs) == 0 {
		return ProjectSet{targetProjects: plan.targetProjects}, nil
	}

	directBuilds := newTargetedProjectBuildQueue(execution)
	directProjectByTarget := make([]int, len(targetPlan.Files))
	for index := range directProjectByTarget {
		directProjectByTarget[index] = -1
	}
	groups := groupTargetsByProjects(targetPlan.Files, plan.targetProjects, func(owner string) []int {
		return orderedProjectIndexesForConfig(plan, owner)
	})
	if request.Scope == ActiveOwners {
		roots := make([][]string, len(plan.specs))
		for index := range roots {
			roots[index] = execution.slots[index].config.FileNames()
		}
		// The eager binder's exact identity ranking is also valid before
		// construction. Copy the groups because ranking filters its input slice.
		directProjectByTarget = directRootOwners(roots, targetPlan.Files, plan.targetProjects,
			append([]projectTargetGroup(nil), groups...), s.FS(), singleThreaded)
	}

	err := runTargetProjectTasks(groups, singleThreaded, func(group projectTargetGroup) error {
		// Service discovery already selected one owner from parsed roots.
		// Reuse that result in every construction scope instead of making a
		// second ownership decision with the ordinary batch identity policy.
		if len(group.projectIndexes) == 1 && plan.specs[group.projectIndexes[0]].sourceReferences {
			projectIndex := group.projectIndexes[0]
			for _, targetIndex := range group.targetIndexes {
				directProjectByTarget[targetIndex] = projectIndex
			}
			return directBuilds.enqueue(projectIndex)
		}
		if request.Scope == ActiveOwners {
			for _, targetIndex := range group.targetIndexes {
				if index := directProjectByTarget[targetIndex]; index >= 0 {
					if err := directBuilds.enqueue(index); err != nil {
						return err
					}
				}
			}
			return nil
		}
		targetIndexes := group.targetIndexes
		unresolved := len(targetIndexes)
		orderedProjectIndexes := group.projectIndexes
		scanProject := func(projectIndex int) error {
			rootFiles, err := execution.rootFileIndex(projectIndex)
			if err != nil {
				return err
			}
			selected := false
			for _, targetIndex := range targetIndexes {
				if directProjectByTarget[targetIndex] >= 0 {
					continue
				}
				target := targetPlan.Files[targetIndex]
				if rootFiles.Contains(target.Path, target.CanonicalPath) {
					directProjectByTarget[targetIndex] = projectIndex
					unresolved--
					selected = true
				}
			}
			if selected {
				if err := directBuilds.enqueue(projectIndex); err != nil {
					return err
				}
			}
			return nil
		}

		nextPosition := 0
		if !singleThreaded {
			predictedTargetsByProject := make(map[int][]int)
			maxPredictedPosition := -1
			for _, targetIndex := range targetIndexes {
				position := execution.predictedProjectPosition(
					orderedProjectIndexes,
					targetPlan.Files[targetIndex],
				)
				if position < 0 {
					continue
				}
				projectIndex := orderedProjectIndexes[position]
				predictedTargetsByProject[projectIndex] = append(
					predictedTargetsByProject[projectIndex],
					targetIndex,
				)
				maxPredictedPosition = max(maxPredictedPosition, position)
			}

			if maxPredictedPosition >= 0 {
				predictedProjects := make([]int, 0, len(predictedTargetsByProject))
				for _, projectIndex := range orderedProjectIndexes[:maxPredictedPosition+1] {
					if _, predicted := predictedTargetsByProject[projectIndex]; predicted {
						predictedProjects = append(predictedProjects, projectIndex)
					}
				}
				execution.parseConcurrent(predictedProjects)
				for _, projectIndex := range predictedProjects {
					rootFiles, parseErr := execution.rootFileIndex(projectIndex)
					if parseErr != nil {
						continue
					}
					for _, targetIndex := range predictedTargetsByProject[projectIndex] {
						target := targetPlan.Files[targetIndex]
						if rootFiles.Contains(target.Path, target.CanonicalPath) {
							if err := directBuilds.enqueue(projectIndex); err != nil {
								return err
							}
							break
						}
					}
				}

				// The nearest containing tsconfig is only a latency hint. Parsing
				// its declaration-order prefix concurrently proves whether an
				// earlier config owns the target; results are still committed in
				// order and no speculative Program can win by finishing first.
				execution.parseConcurrent(orderedProjectIndexes[:maxPredictedPosition+1])
				for nextPosition <= maxPredictedPosition && unresolved > 0 {
					if err := scanProject(orderedProjectIndexes[nextPosition]); err != nil {
						return err
					}
					nextPosition++
				}
			}
		}

		for nextPosition < len(orderedProjectIndexes) && unresolved > 0 {
			if err := scanProject(orderedProjectIndexes[nextPosition]); err != nil {
				return err
			}
			nextPosition++
		}
		return nil
	})
	directBuilds.wait()
	if err != nil {
		return ProjectSet{}, err
	}
	validatedDirectBuilds := make(map[int]struct{})
	for _, projectIndex := range directProjectByTarget {
		if projectIndex < 0 {
			continue
		}
		if _, validated := validatedDirectBuilds[projectIndex]; validated {
			continue
		}
		if err := execution.build(projectIndex); err != nil {
			return ProjectSet{}, err
		}
		validatedDirectBuilds[projectIndex] = struct{}{}
	}

	var rejectedRoots map[int]bool
	for targetIndex, projectIndex := range directProjectByTarget {
		if projectIndex >= 0 && !execution.containsTarget(projectIndex, targetPlan.Files[targetIndex]) {
			if request.Scope == ActiveOwners && !plan.specs[projectIndex].sourceReferences {
				// A listed root rejected by the compiler keeps the broad mode's
				// compatibility fallback. Retain its root owner so binding can
				// distinguish this case from a genuinely unmatched target.
				if rejectedRoots == nil {
					rejectedRoots = make(map[int]bool)
				}
				rejectedRoots[targetIndex] = true
				continue
			}
			return ProjectSet{}, fmt.Errorf(
				"project root %q from %q was absent from its TypeScript Program",
				targetPlan.Files[targetIndex].Path,
				plan.specs[projectIndex].tsconfigPath,
			)
		}
	}
	keep := make([]bool, len(plan.specs))
	for _, projectIndex := range directProjectByTarget {
		if projectIndex >= 0 {
			keep[projectIndex] = true
		}
	}
	if len(rejectedRoots) > 0 {
		// Only matched-but-rejected roots retain ordered source compatibility.
		// Unmatched targets do not expand this construction range.
		fallbackBuilds := newTargetedProjectBuildQueue(execution)
		for _, group := range groups {
			for _, targetIndex := range group.targetIndexes {
				if !rejectedRoots[targetIndex] {
					continue
				}
				for _, projectIndex := range group.projectIndexes {
					if err := fallbackBuilds.enqueue(projectIndex); err != nil {
						fallbackBuilds.wait()
						return ProjectSet{}, err
					}
				}
				break
			}
		}
		fallbackBuilds.wait()
		for index := range execution.slots {
			if err := execution.slots[index].buildErr; err != nil {
				return ProjectSet{}, fmt.Errorf("create TypeScript Program from %q: %w", plan.specs[index].tsconfigPath, err)
			}
			keep[index] = execution.slots[index].program != nil
		}
	}

	return execution.projectSet(keep, directProjectByTarget, targetPlan.Files), nil
}
