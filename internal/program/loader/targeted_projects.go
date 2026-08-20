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
	"github.com/web-infra-dev/rslint/internal/program/projectselection"
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
	metadata  *targetedProjectMetadata
	parseErr  error

	buildOnce sync.Once
	program   *compiler.Program
	buildErr  error

	lookupOnce sync.Once
	lookupMu   sync.Mutex
	lookup     *utils.ProgramSourceLookup
}

type targetedProjectMetadata struct {
	config    *tsoptions.ParsedCommandLine
	rootFiles *lintprogram.RootFileIndex
}

func (metadata *targetedProjectMetadata) DirectRoot(target projectselection.Target) bool {
	return metadata != nil && metadata.rootFiles != nil &&
		metadata.rootFiles.Contains(target.Path, target.CanonicalPath)
}

func (metadata *targetedProjectMetadata) Supports(target projectselection.Target) bool {
	return metadata != nil && metadata.config != nil &&
		lintprogram.CompilerOptionsSupportFileName(
			metadata.config.CompilerOptions(),
			target.Path,
		)
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
}

func newTargetedProjectBuildQueue(execution *targetedProjectExecution) *targetedProjectBuildQueue {
	queue := &targetedProjectBuildQueue{
		execution: execution,
		enqueued:  make([]bool, len(execution.plan.specs)),
	}
	workerCount := min(runtime.GOMAXPROCS(0), len(execution.plan.specs))
	queue.parallel = !execution.singleThreaded && workerCount > 1
	queue.workersN = workerCount
	return queue
}

func (queue *targetedProjectBuildQueue) enqueue(index int) {
	queue.mu.Lock()
	if queue.enqueued[index] {
		queue.mu.Unlock()
		return
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
					_ = queue.execution.build(index)
				}
			}()
		}
	}
	jobs := queue.jobs
	queue.mu.Unlock()
	if !queue.parallel {
		_ = queue.execution.build(index)
		return
	}
	jobs <- index
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
			slot.metadata = &targetedProjectMetadata{
				config: slot.config,
				rootFiles: lintprogram.NewRootFileIndex(
					slot.config.FileNames(),
					execution.session.FS(),
				),
			}
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
	targetPath string,
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
	return slot.lookup.SourceFileForPath(targetPath) != nil
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
	configIndexByDirectory := make(map[string]int, len(configDirs))
	for index, configDir := range configDirs {
		configIndexByDirectory[configDir] = index
	}
	selections := make([]*projectselection.Selection, len(configDirs))
	directTargets := make([]projectselection.DirectTarget, len(targetPlan.Targets))
	loadMetadata := func(projectIndex int) (projectselection.Metadata, bool, error) {
		parsed, parseErr := execution.parse(projectIndex)
		if parseErr != nil {
			return nil, false, parseErr
		}
		return parsed.metadata, parsed.metadata != nil, nil
	}
	loadProject := func(projectIndex int) (bool, error) {
		if buildErr := execution.build(projectIndex); buildErr != nil {
			return false, buildErr
		}
		return true, nil
	}
	containsTarget := func(projectIndex int, target projectselection.Target) bool {
		return execution.containsTarget(projectIndex, target.Path)
	}
	selectionError := func(selectionErr error) error {
		var absent *projectselection.DirectRootAbsentError
		if errors.As(selectionErr, &absent) {
			return fmt.Errorf(
				"project root %q from %q was absent from its TypeScript Program",
				absent.Target.Path,
				plan.specs[absent.Project].tsconfigPath,
			)
		}
		var unavailable *projectselection.DirectProjectUnavailableError
		if errors.As(selectionErr, &unavailable) {
			return fmt.Errorf(
				"project root %q from %q was absent from its TypeScript Program",
				unavailable.Target.Path,
				plan.specs[unavailable.Project].tsconfigPath,
			)
		}
		return selectionErr
	}

	err := runTargetConfigTasks(configDirs, singleThreaded, func(configDir string) error {
		targetIndexes := targetIndexesByConfig[configDir]
		orderedProjectIndexes := orderedProjectIndexesForConfig(plan, configDir)

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
						if parsed.metadata.DirectRoot(selectedTarget(targetPlan.Targets[targetIndex])) {
							directBuilds.enqueue(projectIndex)
							break
						}
					}
				}

				// The nearest containing tsconfig is only a latency hint. Parsing
				// its declaration-order prefix concurrently proves whether an
				// earlier config owns the target; results are still committed in
				// order and no speculative Program can win by finishing first.
				execution.parseConcurrent(orderedProjectIndexes[:maxPredictedPosition+1])
			}
		}

		targets := make([]projectselection.Target, len(targetIndexes))
		for index, targetIndex := range targetIndexes {
			targets[index] = selectedTarget(targetPlan.Targets[targetIndex])
		}
		selection, selectionErr := projectselection.ResolveDirect(
			projectselection.Plan{
				Targets:  targets,
				Projects: orderedProjectIndexes,
			},
			loadMetadata,
			directBuilds.enqueue,
		)
		if selectionErr != nil {
			return selectionError(selectionErr)
		}
		selections[configIndexByDirectory[configDir]] = selection
		for localIndex, targetIndex := range targetIndexes {
			directTargets[targetIndex] = projectselection.DirectTarget{
				Selection: selection,
				Index:     localIndex,
			}
		}
		return nil
	})
	directBuilds.wait()
	if err != nil {
		return ProjectSet{}, err
	}
	if selectionErr := projectselection.ValidateDirectTargets(
		directTargets,
		loadProject,
		containsTarget,
	); selectionErr != nil {
		return ProjectSet{}, selectionError(selectionErr)
	}

	// Direct ownership has been decided for every target before this fallback
	// starts. A project built for another target cannot steal a direct target
	// merely because it imports that file.
	if !singleThreaded && len(configDirs) > 1 {
		s.context.enableConcurrentProgramQueries()
	}
	ownerProjectByTarget := make([]int, len(targetPlan.Targets))
	for index := range ownerProjectByTarget {
		ownerProjectByTarget[index] = projectselection.NoProject
	}
	err = runTargetConfigTasks(configDirs, singleThreaded, func(configDir string) error {
		selection := selections[configIndexByDirectory[configDir]]
		bindings, selectionErr := selection.Complete(loadProject, containsTarget)
		if selectionErr != nil {
			return selectionError(selectionErr)
		}
		targetIndexes := targetIndexesByConfig[configDir]
		for targetIndex, binding := range bindings {
			ownerProjectByTarget[targetIndexes[targetIndex]] = binding.Project
		}
		return nil
	})
	if err != nil {
		return ProjectSet{}, err
	}

	keep := make([]bool, len(plan.specs))
	for _, projectIndex := range ownerProjectByTarget {
		if projectIndex >= 0 {
			keep[projectIndex] = true
		}
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
