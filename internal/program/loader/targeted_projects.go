package loader

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type projectTargetBinding struct {
	targets []rslintconfig.DiscoveredLintTarget
	// owners is complete: direct-root owner, import-fallback owner, or -1.
	owners []int
}

type targetedProjectSlot struct {
	parseOnce sync.Once
	config    *tsoptions.ParsedCommandLine
	rootFiles *lintprogram.RootFileIndex
	parseErr  error

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
) (*compiler.Program, error) {
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
		_, slot.config, slot.parseErr = execution.session.context.parseConfig(
			spec.programCwd,
			spec.tsconfigPath,
		)
		if slot.parseErr == nil && slot.config == nil {
			slot.parseErr = errors.New("no parsed config returned")
		}
		if slot.parseErr == nil {
			slot.rootFiles = lintprogram.NewRootFileIndex(
				slot.config.FileNames(),
				execution.session.FS(),
			)
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
		)
	})
	if slot.buildErr != nil {
		return fmt.Errorf("create TypeScript Program from %q: %w", spec.tsconfigPath, slot.buildErr)
	}
	return nil
}

func (execution *targetedProjectExecution) containsTarget(
	index int,
	target rslintconfig.DiscoveredLintTarget,
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
	return slot.lookup.SourceFileForPath(target.Path) != nil
}

func (execution *targetedProjectExecution) supportsTarget(
	index int,
	target rslintconfig.DiscoveredLintTarget,
) (bool, error) {
	parsed, err := execution.parse(index)
	if err != nil {
		return false, err
	}
	return lintprogram.CompilerOptionsSupportFileName(
		parsed.config.CompilerOptions(),
		target.Path,
	), nil
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
	target rslintconfig.DiscoveredLintTarget,
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

func runTargetConfigTasks(
	configDirs []string,
	singleThreaded bool,
	task func(configDir string) error,
) error {
	if len(configDirs) == 0 {
		return nil
	}
	workerCount := min(runtime.GOMAXPROCS(0), len(configDirs))
	if singleThreaded || workerCount <= 1 {
		for _, configDir := range configDirs {
			if err := task(configDir); err != nil {
				return err
			}
		}
		return nil
	}

	errs := make([]error, len(configDirs))
	jobs := make(chan int, workerCount)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				errs[index] = task(configDirs[index])
			}
		}()
	}
	for index := range configDirs {
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
	ownerProjectByTarget []int,
	targets []rslintconfig.DiscoveredLintTarget,
) ProjectSet {
	binding := &projectTargetBinding{
		targets: append([]rslintconfig.DiscoveredLintTarget(nil), targets...),
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
	for targetIndex, projectIndex := range ownerProjectByTarget {
		if projectIndex < 0 || targetIndex >= len(targets) {
			continue
		}
		setIndex := projectSetIndexByPlanIndex[projectIndex]
		if setIndex < 0 {
			continue
		}
		binding.owners[targetIndex] = setIndex
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

// BuildTargetProjects materializes only configured projects needed to decide
// ownership for the supplied lint targets. TypeScript config roots have first
// priority; targets outside every root retain the historical declaration-order
// fallback to projects that contain them through module resolution.
func (s *Session) BuildTargetProjects(
	configs map[string]rslintconfig.RslintConfig,
	targetPlan rslintconfig.LintTargetPlan,
	singleThreaded bool,
) (ProjectSet, error) {
	if err := s.validate(); err != nil {
		return ProjectSet{}, err
	}
	if len(configs) == 0 || len(targetPlan.Targets) == 0 {
		return ProjectSet{}, nil
	}
	plan := buildProjectPlan(configs, s.FS())
	if plan.terminalErr != nil {
		return ProjectSet{}, plan.terminalErr
	}
	if len(plan.specs) == 0 {
		return ProjectSet{}, nil
	}

	execution := newTargetedProjectExecution(s, plan, singleThreaded)
	directBuilds := newTargetedProjectBuildQueue(execution)
	directProjectByTarget := make([]int, len(targetPlan.Targets))
	for index := range directProjectByTarget {
		directProjectByTarget[index] = -1
	}
	targetIndexesByConfig := make(map[string][]int)
	for targetIndex, target := range targetPlan.Targets {
		targetIndexesByConfig[target.ConfigDirectory] = append(
			targetIndexesByConfig[target.ConfigDirectory],
			targetIndex,
		)
	}
	configDirs := make([]string, 0, len(targetIndexesByConfig))
	for configDir := range targetIndexesByConfig {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	err := runTargetConfigTasks(configDirs, singleThreaded, func(configDir string) error {
		targetIndexes := targetIndexesByConfig[configDir]
		unresolved := len(targetIndexes)
		orderedProjectIndexes := orderedProjectIndexesForConfig(plan, configDir)
		scanProject := func(projectIndex int) error {
			parsed, err := execution.parse(projectIndex)
			if err != nil {
				return err
			}
			selected := false
			for _, targetIndex := range targetIndexes {
				if directProjectByTarget[targetIndex] >= 0 {
					continue
				}
				target := targetPlan.Targets[targetIndex]
				if parsed.rootFiles.Contains(target.Path, target.CanonicalPath) {
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
					targetPlan.Targets[targetIndex],
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
					parsed, parseErr := execution.parse(projectIndex)
					if parseErr != nil {
						continue
					}
					for _, targetIndex := range predictedTargetsByProject[projectIndex] {
						target := targetPlan.Targets[targetIndex]
						if parsed.rootFiles.Contains(target.Path, target.CanonicalPath) {
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

	keep := make([]bool, len(plan.specs))
	for _, projectIndex := range directProjectByTarget {
		if projectIndex >= 0 {
			keep[projectIndex] = true
		}
	}
	for targetIndex, projectIndex := range directProjectByTarget {
		if projectIndex >= 0 && !execution.containsTarget(projectIndex, targetPlan.Targets[targetIndex]) {
			return ProjectSet{}, fmt.Errorf(
				"project root %q from %q was absent from its TypeScript Program",
				targetPlan.Targets[targetIndex].Path,
				plan.specs[projectIndex].tsconfigPath,
			)
		}
	}

	// Direct ownership has been decided for every target before this fallback
	// starts. A project built for another target cannot steal a direct target
	// merely because it imports that file.
	if !singleThreaded && len(configDirs) > 1 {
		s.context.enableConcurrentProgramQueries()
	}
	var keepMu sync.Mutex
	ownerProjectByTarget := append([]int(nil), directProjectByTarget...)
	err = runTargetConfigTasks(configDirs, singleThreaded, func(configDir string) error {
		pending := make(map[int]struct{})
		for _, targetIndex := range targetIndexesByConfig[configDir] {
			if directProjectByTarget[targetIndex] < 0 {
				pending[targetIndex] = struct{}{}
			}
		}
		orderedProjectIndexes := orderedProjectIndexesForConfig(plan, configDir)
		fallbackProjectIndexes := make([]int, 0, len(orderedProjectIndexes))
		for _, projectIndex := range orderedProjectIndexes {
			for targetIndex := range pending {
				supported, supportErr := execution.supportsTarget(
					projectIndex,
					targetPlan.Targets[targetIndex],
				)
				if supportErr != nil {
					return supportErr
				}
				if supported {
					fallbackProjectIndexes = append(fallbackProjectIndexes, projectIndex)
					break
				}
			}
		}
		for _, projectIndex := range fallbackProjectIndexes {
			if len(pending) == 0 {
				break
			}
			if err := execution.build(projectIndex); err != nil {
				return err
			}
			selected := false
			for targetIndex := range pending {
				if execution.containsTarget(projectIndex, targetPlan.Targets[targetIndex]) {
					delete(pending, targetIndex)
					ownerProjectByTarget[targetIndex] = projectIndex
					selected = true
				}
			}
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

	return execution.projectSet(keep, ownerProjectByTarget, targetPlan.Targets), nil
}

func (s *Session) BuildTargetProject(
	configDirectory string,
	config rslintconfig.RslintConfig,
	targetPlan rslintconfig.LintTargetPlan,
	singleThreaded bool,
) (ProjectSet, error) {
	return s.BuildTargetProjects(
		map[string]rslintconfig.RslintConfig{configDirectory: config},
		targetPlan,
		singleThreaded,
	)
}
