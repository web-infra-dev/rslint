package loader

import (
	"errors"
	"fmt"
	"runtime"
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
	targets []rslintconfig.PlannedLintTarget
	// owners is complete: direct-root owner, import-fallback owner, or -1.
	owners []int
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
		lintprogram.CompilerOptionsSupportFileName(metadata.config.CompilerOptions(), target.Path)
}

type targetedProjectSlot struct {
	parseOnce sync.Once
	config    *tsoptions.ParsedCommandLine
	rootFiles *lintprogram.RootFileIndex
	parseErr  error

	buildOnce sync.Once
	buildDone chan struct{}
	program   *compiler.Program
	buildErr  error

	lookupOnce sync.Once
	lookupMu   sync.Mutex
	lookup     *utils.ProgramSourceLookup
}

type targetedProjectExecution struct {
	session         *Session
	plan            projectPlan
	singleThreaded  bool
	slots           []targetedProjectSlot
	identities      *lintprogram.PathIdentityResolver
	prefetchedFiles *programFileIndex
}

func (execution *targetedProjectExecution) preparePrefetchedEvidence(
	targets []projectselection.Target,
	candidateIndexes []int,
) {
	programs := make([]*compiler.Program, len(execution.slots))
	discoveredTargets := make([]rslintconfig.DiscoveredLintTarget, len(targets))
	for targetIndex, target := range targets {
		discoveredTargets[targetIndex] = rslintconfig.DiscoveredLintTarget{
			Path:          target.Path,
			CanonicalPath: target.CanonicalPath,
		}
	}
	for project := range execution.slots {
		programs[project] = execution.slots[project].program
	}
	execution.prefetchedFiles = newProgramFileIndexWithResolver(
		programs,
		discoveredTargets,
		execution.session.FS(),
		execution.identities,
	)
	rootFilesByCandidate := make([][]string, len(candidateIndexes))
	for candidate, project := range candidateIndexes {
		if config := execution.slots[project].config; config != nil {
			rootFilesByCandidate[candidate] = config.FileNames()
		}
	}
	rootIndexes := lintprogram.NewRootFileIndexes(
		rootFilesByCandidate,
		execution.prefetchedFiles.identities,
	)
	for candidate, project := range candidateIndexes {
		execution.slots[project].rootFiles = rootIndexes[candidate]
	}
}

func newTargetedProjectExecution(
	session *Session,
	plan projectPlan,
	targets []projectselection.Target,
	singleThreaded bool,
) *targetedProjectExecution {
	known := make([]lintprogram.PathIdentity, len(targets))
	for index, target := range targets {
		known[index] = lintprogram.PathIdentity{
			Path:          target.Path,
			CanonicalPath: target.CanonicalPath,
		}
	}
	execution := &targetedProjectExecution{
		session:        session,
		plan:           plan,
		singleThreaded: singleThreaded,
		slots:          make([]targetedProjectSlot, len(plan.specs)),
		identities: lintprogram.NewPathIdentityResolver(
			session.FS(),
			singleThreaded,
			known,
		),
	}
	for index := range execution.slots {
		execution.slots[index].buildDone = make(chan struct{})
	}
	return execution
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
	if index < 0 || index >= len(execution.slots) {
		return nil, fmt.Errorf("invalid project index %d", index)
	}
	slot := &execution.slots[index]
	spec := execution.plan.specs[index]
	slot.parseOnce.Do(func() {
		_, slot.config, slot.parseErr = execution.session.context.parseConfig(spec.programCwd, spec.tsconfigPath)
		if slot.parseErr == nil && slot.config == nil {
			slot.parseErr = errors.New("no parsed config returned")
		}
		if slot.parseErr == nil {
			slot.rootFiles = lintprogram.NewRootFileIndexWithResolver(
				slot.config.FileNames(),
				execution.identities,
			)
		}
	})
	if slot.parseErr != nil {
		return nil, fmt.Errorf("parse TypeScript config %q: %w", spec.tsconfigPath, slot.parseErr)
	}
	return slot, nil
}

func (execution *targetedProjectExecution) metadata(index int) (*targetedProjectMetadata, error) {
	slot, err := execution.parse(index)
	if err != nil {
		return nil, err
	}
	return &targetedProjectMetadata{
		config:    slot.config,
		rootFiles: slot.rootFiles,
	}, nil
}

func (execution *targetedProjectExecution) build(index int) error {
	slot := &execution.slots[index]
	spec := execution.plan.specs[index]
	slot.buildOnce.Do(func() {
		defer close(slot.buildDone)
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
	return execution.buildError(index)
}

func (execution *targetedProjectExecution) buildError(index int) error {
	slot := &execution.slots[index]
	if slot.buildErr != nil {
		return fmt.Errorf(
			"create TypeScript Program from %q: %w",
			execution.plan.specs[index].tsconfigPath,
			slot.buildErr,
		)
	}
	return nil
}

func (execution *targetedProjectExecution) containsTarget(index int, target projectselection.Target) bool {
	if files := execution.prefetchedFiles; files != nil {
		if index < 0 || index >= len(files.programs) {
			return false
		}
		if exactProgramSourceFile(files.programs[index], target.Path) != nil {
			return true
		}
		canonicalPath := files.canonicalPath(target.Path, target.CanonicalPath)
		return files.sourceFile(target.Projects, index, canonicalPath) != nil
	}
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
	workerCount := min(runtime.GOMAXPROCS(0), len(execution.plan.specs))
	return &targetedProjectBuildQueue{
		execution: execution,
		parallel:  !execution.singleThreaded && workerCount > 1,
		workersN:  workerCount,
		enqueued:  make([]bool, len(execution.plan.specs)),
	}
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
				for project := range queue.jobs {
					_ = queue.execution.build(project)
				}
			}()
		}
	}
	jobs := queue.jobs
	queue.mu.Unlock()
	if queue.parallel {
		jobs <- index
		return
	}
	_ = queue.execution.build(index)
}

func (queue *targetedProjectBuildQueue) await(index int) error {
	queue.enqueue(index)
	<-queue.execution.slots[index].buildDone
	return queue.execution.buildError(index)
}

func (queue *targetedProjectBuildQueue) awaitCompletion(index int) {
	queue.enqueue(index)
	<-queue.execution.slots[index].buildDone
}

func (queue *targetedProjectBuildQueue) wait() {
	if queue.parallel && queue.jobs != nil {
		close(queue.jobs)
		queue.workers.Wait()
	}
}

// scheduleDirectHint proves that one proximity candidate directly contains
// every request target. In parallel demand mode it overlaps that Program with
// the ordered metadata frontier; for a directory batch the same proof avoids
// broader candidate prefetch. The selector still scans every earlier candidate
// and alone commits ownership or observes errors.
func (execution *targetedProjectExecution) scheduleDirectHint(
	targets []projectselection.Target,
	builds *targetedProjectBuildQueue,
) bool {
	if builds == nil || len(targets) == 0 {
		return false
	}
	options := tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true}
	hintedProject := -1
	for _, target := range targets {
		project := -1
		bestDirectoryLength := -1
		for _, candidate := range target.Projects {
			if candidate < 0 || candidate >= len(execution.plan.specs) {
				continue
			}
			directory := execution.plan.specs[candidate].programCwd
			if !tspath.ContainsPath(directory, target.Path, options) ||
				len(directory) <= bestDirectoryLength {
				continue
			}
			project = candidate
			bestDirectoryLength = len(directory)
		}
		if project < 0 || hintedProject >= 0 && hintedProject != project {
			return false
		}
		hintedProject = project
	}
	if hintedProject < 0 {
		return false
	}
	slot, err := execution.parse(hintedProject)
	if err != nil || slot.rootFiles == nil {
		return false
	}
	for _, target := range targets {
		if !slot.rootFiles.Contains(target.Path, target.CanonicalPath) {
			return false
		}
	}
	if builds.parallel {
		builds.enqueue(hintedProject)
	}
	return true
}

func buildProjectPathPlan(paths ...[]string) (projectPlan, [][]int) {
	plan := projectPlan{}
	indexesByList := make([][]int, len(paths))
	projectByPath := make(map[string]int)
	for listIndex, projectPaths := range paths {
		indexes := make([]int, 0, len(projectPaths))
		for _, projectPath := range projectPaths {
			projectPath = tspath.NormalizePath(projectPath)
			projectID := exactPathID(projectPath)
			projectIndex, exists := projectByPath[projectID]
			if !exists {
				projectIndex = len(plan.specs)
				projectByPath[projectID] = projectIndex
				plan.specs = append(plan.specs, projectSpec{
					tsconfigPath: projectPath,
					programCwd:   tspath.GetDirectoryPath(projectPath),
				})
			}
			indexes = append(indexes, projectIndex)
		}
		indexesByList[listIndex] = indexes
	}
	return plan, indexesByList
}

func selectionError(err error, plan projectPlan) error {
	var absent *projectselection.DirectRootAbsentError
	if errors.As(err, &absent) {
		return fmt.Errorf(
			"project root %q from %q was absent from its TypeScript Program",
			absent.Target.Path,
			plan.specs[absent.Project].tsconfigPath,
		)
	}
	var unavailable *projectselection.DirectProjectUnavailableError
	if errors.As(err, &unavailable) {
		return fmt.Errorf(
			"project root %q from %q was absent from its TypeScript Program",
			unavailable.Target.Path,
			plan.specs[unavailable.Project].tsconfigPath,
		)
	}
	return err
}

// SelectProjects binds one target-effective project plan. catalogProjects is
// the independent program-wide type-check role. prefetchCandidates may start
// every lint candidate early, but it cannot widen a target's candidates,
// publish an error, or choose an owner.
func (s *Session) SelectProjects(
	lintPlan rslintconfig.LintProjectPlan,
	catalogProjects []string,
	prefetchCandidates bool,
	singleThreaded bool,
) (ProjectSet, error) {
	if err := s.validate(); err != nil {
		return ProjectSet{}, err
	}
	catalogRequested := catalogProjects != nil
	pathLists := make([][]string, 0, len(lintPlan.Targets)+1)
	pathLists = append(pathLists, catalogProjects)
	for _, target := range lintPlan.Targets {
		pathLists = append(pathLists, target.ProjectPaths)
	}
	plan, indexesByList := buildProjectPathPlan(pathLists...)
	catalogIndexes := indexesByList[0]
	selectionTargets := make([]projectselection.Target, len(lintPlan.Targets))
	for targetIndex, target := range lintPlan.Targets {
		selectionTargets[targetIndex] = projectselection.Target{
			Path:          target.Target.Path,
			CanonicalPath: target.Target.CanonicalPath,
			Projects:      indexesByList[targetIndex+1],
		}
	}

	if len(plan.specs) == 0 {
		return projectSetWithoutProjects(lintPlan, catalogRequested), nil
	}
	execution := newTargetedProjectExecution(s, plan, selectionTargets, singleThreaded)
	builds := newTargetedProjectBuildQueue(execution)
	defer builds.wait()

	// Catalog construction is an explicit program-wide consumer. Build it in
	// parallel, then observe errors in stable catalog order before lint project
	// selection, matching the established type-check boundary.
	for _, project := range catalogIndexes {
		builds.enqueue(project)
	}
	for _, project := range catalogIndexes {
		if err := builds.await(project); err != nil {
			return ProjectSet{}, err
		}
	}
	completeDirectHint := false
	if builds.parallel || prefetchCandidates {
		completeDirectHint = execution.scheduleDirectHint(selectionTargets, builds)
	}
	if prefetchCandidates && !completeDirectHint {
		candidateIndexes := make([]int, 0, len(plan.specs))
		candidateSeen := make([]bool, len(plan.specs))
		for _, target := range selectionTargets {
			for _, project := range target.Projects {
				builds.enqueue(project)
				if !candidateSeen[project] {
					candidateSeen[project] = true
					candidateIndexes = append(candidateIndexes, project)
				}
			}
		}
		// Prefetch is a provider scheduling policy: complete the bounded build
		// phase without observing its errors, then expose batched root/source
		// evidence to the same selector used by demand-driven requests.
		for _, project := range candidateIndexes {
			builds.awaitCompletion(project)
		}
		execution.preparePrefetchedEvidence(selectionTargets, candidateIndexes)
	}

	loadMetadata := func(project int) (projectselection.Metadata, bool, error) {
		metadata, err := execution.metadata(project)
		if err != nil {
			return nil, false, err
		}
		return metadata, metadata != nil, nil
	}
	loadProject := func(project int) (bool, error) {
		if err := builds.await(project); err != nil {
			return false, err
		}
		return execution.slots[project].program != nil, nil
	}
	contains := func(project int, target projectselection.Target) bool {
		return execution.containsTarget(project, target)
	}
	selection, err := projectselection.ResolveDirect(
		projectselection.Plan{Targets: selectionTargets},
		loadMetadata,
		builds.enqueue,
	)
	if err != nil {
		return ProjectSet{}, selectionError(err, plan)
	}
	bindings, err := selection.Complete(loadMetadata, loadProject, contains)
	if err != nil {
		return ProjectSet{}, selectionError(err, plan)
	}
	return execution.projectSet(lintPlan, bindings, catalogIndexes, catalogRequested), nil
}

func projectSetWithoutProjects(
	lintPlan rslintconfig.LintProjectPlan,
	catalogRequested bool,
) ProjectSet {
	binding := &projectTargetBinding{
		targets: append([]rslintconfig.PlannedLintTarget(nil), lintPlan.Targets...),
		owners:  make([]int, len(lintPlan.Targets)),
	}
	for index := range binding.owners {
		binding.owners[index] = projectselection.NoProject
	}
	set := ProjectSet{targetBinding: binding}
	if catalogRequested {
		set.typeCheckPrograms = make([]*lintprogram.Program, 0)
	}
	return set
}

func (execution *targetedProjectExecution) projectSet(
	lintPlan rslintconfig.LintProjectPlan,
	bindings []projectselection.Binding,
	catalogIndexes []int,
	catalogRequested bool,
) ProjectSet {
	keep := make([]bool, len(execution.plan.specs))
	for _, project := range catalogIndexes {
		keep[project] = true
	}
	for _, binding := range bindings {
		if binding.Project >= 0 {
			keep[binding.Project] = true
		}
	}
	set := ProjectSet{targetBinding: &projectTargetBinding{
		targets: append([]rslintconfig.PlannedLintTarget(nil), lintPlan.Targets...),
		owners:  make([]int, len(bindings)),
	}}
	if catalogRequested {
		set.typeCheckPrograms = make([]*lintprogram.Program, 0, len(catalogIndexes))
	}
	for index := range set.targetBinding.owners {
		set.targetBinding.owners[index] = projectselection.NoProject
	}
	setIndexByProject := make([]int, len(execution.plan.specs))
	for index := range setIndexByProject {
		setIndexByProject[index] = -1
	}
	for project, retained := range keep {
		if !retained || execution.slots[project].program == nil {
			continue
		}
		setIndexByProject[project] = len(set.compilerPrograms)
		set.compilerPrograms = append(set.compilerPrograms, execution.slots[project].program)
		set.programs = append(set.programs, lintprogram.NewFromCompiler(execution.slots[project].program))
	}
	for targetIndex, binding := range bindings {
		if binding.Project >= 0 {
			set.targetBinding.owners[targetIndex] = setIndexByProject[binding.Project]
		}
	}
	seenTypeCheck := make(map[int]struct{}, len(catalogIndexes))
	for _, project := range catalogIndexes {
		setIndex := setIndexByProject[project]
		if setIndex < 0 {
			continue
		}
		if _, exists := seenTypeCheck[setIndex]; exists {
			continue
		}
		seenTypeCheck[setIndex] = struct{}{}
		set.typeCheckPrograms = append(set.typeCheckPrograms, set.programs[setIndex])
	}
	return set
}
