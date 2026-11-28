package loader

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ProgramLoad constructs and binds a complete Program, never a subset of
// its roots or dependency graph. Options and TargetCount are metadata only;
// the caller decides when to load. Load must be called once, and different
// loads may run concurrently. An unused coupled fallback candidate or a
// trailing unbound slot with no unbound targets returns an empty LoadResult{}.
// Returned Programs belong solely to the caller.
type ProgramLoad struct {
	TargetCount int
	Options     *core.CompilerOptions
	Load        func() (LoadResult, error)
}

// PlanLoads separates target ownership from source construction. A nil plan
// keeps the whole-binding acquisition. Independent configured projects stream
// as singleton loads; outside Targeted mode, any non-independent projects plus
// candidate projects reachable by their fallback lists share one lazy coupled
// binding that constructs the coupled subset together on first coupled unit
// acquisition, yields each Program in declaration order, and keeps all
// unmatched/fallback-unbound roots in one trailing group. When no project is
// independent, or when a Targeted, service-discovered, or discovery-only
// project fails metadata independence, PlanLoads returns nil for whole-binding
// acquisition. Discovery-only targets contribute configured Programs and
// service source validation. Compiler-only roots preserve consumers that
// require the compiler AST representation; otherwise the parser/binder facade
// handles supported project-external roots.
//
// Call on a fresh Session, before constructing any Programs. Config validation
// and root identity ranking are shared with BuildProjects. The loads retain
// immutable metadata and source snapshots, but no completed Program or AST.
func (s *Session) PlanLoads(request ProjectBuildRequest, cwd string, compilerRoots bool) ([]ProgramLoad, error) {
	plan, err := s.prepareProjectPlan(request)
	if err != nil {
		return nil, err
	}
	if len(plan.specs) < 2 {
		return nil, nil
	}
	execution := newTargetedProjectExecution(s, plan, request.SingleThreaded)
	if request.Scope == ActiveOwners || request.Scope == AllDeclared {
		if err := execution.validateActiveConfigs(); err != nil {
			return nil, err
		}
	}
	if plan.terminalErr != nil {
		return nil, plan.terminalErr
	}
	filesByProject, owners, err := execution.loadTargetsByProject(request)
	if err != nil {
		return nil, err
	}
	projectCount := len(filesByProject)
	if _, hasRoots := filesByProject[-1]; hasRoots {
		projectCount--
	}
	if request.Scope == AllDeclared {
		projectCount = len(plan.specs)
	}
	if projectCount < 2 {
		return nil, nil
	}
	// Ordinary roots can require ordered source-membership fallback across
	// projects (for example a listed .js root without allowJs). Partition
	// non-independent projects together with every candidate reachable by
	// their fallback lists so coherent binding resolves within that coupled
	// subset while isolated independent projects still stream. Preserve the
	// materialized binding's declaration order, with unmatched roots last,
	// independently of target enumeration and completion timing.
	order := make([]int, 0, len(filesByProject))
	inOrder := make([]bool, len(plan.specs))
	for owner := range plan.specs {
		if request.Scope == AllDeclared || len(filesByProject[owner]) > 0 {
			order = append(order, owner)
			inOrder[owner] = true
		}
	}
	if len(filesByProject[-1]) > 0 {
		order = append(order, -1)
	}
	// Resolve metadata eligibility before reading source or publishing loads.
	independent := make([]bool, len(order))
	errs := make([]error, len(order))
	var refCache planReferenceCache
	work := core.NewWorkGroup(request.SingleThreaded)
	for i, owner := range order {
		work.Queue(func() {
			if owner < 0 {
				independent[i] = true
				return
			}
			slot, err := execution.parse(owner)
			errs[i] = err
			if err == nil {
				independent[i] = s.hasIndependentRoots(slot.config, filesByProject[owner], plan.specs[owner].sourceReferences, &refCache)
			}
		})
	}
	work.RunAndWait()
	for i := range order {
		if errs[i] != nil {
			return nil, errs[i]
		}
	}
	useCaseSensitive := s.FS().UseCaseSensitiveFileNames()
	if !allRootsSupportedByParser(filesByProject[-1], useCaseSensitive) {
		return nil, nil
	}
	coupled := make([]bool, len(plan.specs))
	hasCoupled := false
	indexesByOwner := make(map[string][]int)
	for i, owner := range order {
		if owner < 0 || independent[i] {
			continue
		}
		if plan.specs[owner].sourceReferences || request.SkipLintBinding || request.Scope == Targeted ||
			!allRootsSupportedByParser(filesByProject[owner], useCaseSensitive) {
			// Service-discovered, discovery-only, Targeted-scope, and
			// parser-unsupported coupled targets retain authoritative
			// whole-binding validation before exposing any loads.
			return nil, nil
		}
		coupled[owner] = true
		hasCoupled = true
		// In broad and declared modes, a compiler-rejected root falls back
		// across every project in its target's candidate list. Pull every
		// reachable candidate into the coupled sub-plan so no independent
		// project can miss a fallback target or drift from candidate order.
		for _, file := range filesByProject[owner] {
			for _, candidate := range projectIndexesForTarget(file, plan.targetProjects, indexesByOwner, func(dir string) []int {
				return orderedProjectIndexesForConfig(plan, dir)
			}) {
				coupled[candidate] = true
			}
		}
	}
	independentProjects := 0
	for _, owner := range order {
		if owner >= 0 && !coupled[owner] {
			independentProjects++
		}
	}
	if independentProjects == 0 {
		return nil, nil
	}
	if !request.SingleThreaded {
		s.context.enableConcurrentProgramQueries()
	}
	s.context.parseCache = utils.NewWeakParseCache()
	// Preflight and Program construction consume one source snapshot. Spread
	// these independent reads across files, including roots in a large project.
	// A failed read retains the complete binding's authoritative error path.
	var roots []string
	for _, owner := range order {
		for _, file := range filesByProject[owner] {
			roots = append(roots, file.Path)
		}
	}
	workers := min(runtime.GOMAXPROCS(0), len(roots))
	if request.SingleThreaded {
		workers = 1
	}
	readable := make([]bool, workers)
	reads := core.NewWorkGroup(request.SingleThreaded)
	fsys := s.context.compilerFS()
	for worker := range workers {
		reads.Queue(func() {
			for i := worker; i < len(roots); i += workers {
				if _, ok := s.context.parseCache.ReadSourceText(fsys, roots[i]); !ok {
					return
				}
			}
			readable[worker] = true
		})
	}
	reads.RunAndWait()
	for _, ok := range readable {
		if !ok {
			return nil, nil
		}
	}
	var rootIdentities *ProgramRootIdentityProjector
	if request.IncludeProgramRootIdentities {
		rootIdentities, err = s.NewProgramRootIdentityProjector(request.Targets.Files, request.SingleThreaded)
		if err != nil {
			return nil, err
		}
	}
	var coupledGroup *coupledLoadGroup
	var coupledIndexByOrig []int
	if hasCoupled {
		coupledGroup, coupledIndexByOrig = s.newCoupledLoadGroup(
			plan, execution, request, owners, coupled, indexesByOwner,
			cwd, compilerRoots, rootIdentities,
		)
	}
	loads := make([]ProgramLoad, 0, len(order)+1)
	for owner := range plan.specs {
		if coupled[owner] {
			coupledIndex := coupledIndexByOrig[owner]
			var options *core.CompilerOptions
			if parsed := execution.slots[owner].config; parsed != nil {
				options = parsed.CompilerOptions()
			}
			loads = append(loads, ProgramLoad{
				TargetCount: len(filesByProject[owner]),
				Options:     options,
				Load: func() (LoadResult, error) {
					return coupledGroup.takeSpec(coupledIndex)
				},
			})
			continue
		}
		if !inOrder[owner] {
			continue
		}
		files := filesByProject[owner]
		spec := plan.specs[owner]
		var validationFiles []target.File
		if request.SkipLintBinding {
			validationFiles, files = files, nil
		}
		slot, err := execution.parse(owner)
		if err != nil {
			return nil, err
		}
		parsed := slot.config
		loads = append(loads, ProgramLoad{
			TargetCount: len(files),
			Options:     parsed.CompilerOptions(),
			Load: func() (LoadResult, error) {
				cp, err := s.context.createProjectProgramFromParsedConfig(request.SingleThreaded, spec.programCwd, parsed, spec.sourceReferences)
				if err != nil {
					return LoadResult{}, fmt.Errorf("create TypeScript Program from %q: %w", spec.tsconfigPath, err)
				}
				compilerPrograms := []*compiler.Program{cp}
				binding := LoadResult{Programs: lintprogram.NewOwnedFromCompilers(compilerPrograms)}
				if request.SkipLintBinding {
					for _, file := range validationFiles {
						if err := validateProjectSource(cp, s.FS(), file); err != nil {
							return LoadResult{}, err
						}
					}
				} else {
					set := ProjectSet{
						compilerPrograms: compilerPrograms,
						programs:         binding.Programs,
						configOrders:     []configOrders{spec.configOrders},
						targetBinding:    &projectTargetBinding{targets: files, owners: make([]int, len(files))},
						targetProjects:   make(map[target.File][]int, len(files)),
					}
					for _, file := range files {
						set.targetProjects[file] = []int{0}
					}
					var unbound []target.File
					binding, unbound = s.bindTargetsToProjects(set, target.Plan{Files: files}, request.SingleThreaded)
					if len(unbound) > 0 {
						return LoadResult{}, fmt.Errorf("project root %q from %q was absent from its TypeScript Program", unbound[0].Path, spec.tsconfigPath)
					}
					finalizeResult(&binding)
				}
				if rootIdentities != nil {
					binding.ProgramRootIdentities = rootIdentities.Freeze(binding.Programs)
				}
				return binding, nil
			},
		})
	}
	if hasCoupled {
		// Keep all unmatched targets and any coupled fallback-unbound targets
		// together in one trailing group, preserving target order and
		// case-collision partitioning.
		loads = append(loads, ProgramLoad{
			TargetCount: len(filesByProject[-1]),
			Load: func() (LoadResult, error) {
				return coupledGroup.takeUnbound()
			},
		})
	} else if files := filesByProject[-1]; len(files) > 0 {
		// Keep all unmatched roots together, including cross-file binding
		// and the existing case-collision partition. Never shard their ASTs.
		loads = append(loads, ProgramLoad{
			TargetCount: len(files),
			Load: func() (LoadResult, error) {
				return s.loadUnboundTargets(files, cwd, compilerRoots, request.SingleThreaded)
			},
		})
	}
	return loads, nil
}

// coupledLoadGroup lazily executes the multi-project build and target-binding
// path for the coupled subset of a partially streamed plan. Each coupled
// Program is moved out on its own declaration-order ProgramLoad so it can be
// released as soon as its unit finishes, and any unbound targets are parsed on
// demand in the trailing unbound load.
type coupledLoadGroup struct {
	session        *Session
	plan           projectPlan
	targets        []target.File
	scope          ProjectScope
	singleThreaded bool
	cwd            string
	compilerRoots  bool
	rootIdentities *ProgramRootIdentityProjector

	once    sync.Once
	err     error
	mu      sync.Mutex
	bySpec  []LoadResult
	unbound []target.File
}

func (s *Session) newCoupledLoadGroup(
	plan projectPlan,
	execution *targetedProjectExecution,
	request ProjectBuildRequest,
	owners []int,
	coupled []bool,
	indexesByOwner map[string][]int,
	cwd string,
	compilerRoots bool,
	rootIdentities *ProgramRootIdentityProjector,
) (*coupledLoadGroup, []int) {
	coupledIndexByOrig := make([]int, len(plan.specs))
	for i := range coupledIndexByOrig {
		coupledIndexByOrig[i] = -1
	}
	coupledPlan := projectPlan{}
	for orig, isCoupled := range coupled {
		if !isCoupled {
			continue
		}
		coupledIndexByOrig[orig] = len(coupledPlan.specs)
		spec := plan.specs[orig]
		if parsed := execution.slots[orig].config; parsed != nil {
			spec.parsed = parsed
		}
		coupledPlan.specs = append(coupledPlan.specs, spec)
	}
	var coupledFiles []target.File
	coupledTargetProjects := make(map[target.File][]int)
	remappedLists := make(map[projectIndexListID][]int)
	for index, file := range request.Targets.Files {
		owner := owners[index]
		if owner >= 0 && !coupled[owner] {
			continue
		}
		coupledFiles = append(coupledFiles, file)
		candidates := projectIndexesForTarget(file, plan.targetProjects, indexesByOwner, func(dir string) []int {
			return orderedProjectIndexesForConfig(plan, dir)
		})
		key := projectIndexListIdentity(candidates)
		remapped, cached := remappedLists[key]
		if !cached {
			for _, orig := range candidates {
				if coupledIndex := coupledIndexByOrig[orig]; coupledIndex >= 0 {
					remapped = append(remapped, coupledIndex)
				}
			}
			remappedLists[key] = remapped
		}
		coupledTargetProjects[file] = remapped
	}
	if len(coupledFiles) > 0 {
		coupledPlan.targetProjects = coupledTargetProjects
	}
	return &coupledLoadGroup{
		session:        s,
		plan:           coupledPlan,
		targets:        coupledFiles,
		scope:          request.Scope,
		singleThreaded: request.SingleThreaded,
		cwd:            cwd,
		compilerRoots:  compilerRoots,
		rootIdentities: rootIdentities,
	}, coupledIndexByOrig
}

func (g *coupledLoadGroup) ensureBuilt() error {
	g.once.Do(func() {
		plan := g.plan
		targets := g.targets
		defer func() {
			g.plan = projectPlan{}
			g.targets = nil
			g.rootIdentities = nil
		}()
		var set ProjectSet
		if g.scope == AllDeclared {
			set, g.err = g.session.executeProjectPlan(plan, g.singleThreaded)
		} else {
			set, g.err = g.session.executeTargetProjectPlan(plan, ProjectBuildRequest{
				Targets:        target.Plan{Files: targets},
				Scope:          g.scope,
				SingleThreaded: g.singleThreaded,
			})
		}
		if g.err != nil {
			return
		}
		binding, unbound, mappingsByProgram := g.session.bindTargetsToProjectsWithPerProgramMappings(
			set,
			target.Plan{Files: targets},
			g.singleThreaded,
			true,
		)
		g.unbound = unbound
		finalizeResult(&binding)
		g.bySpec = make([]LoadResult, len(plan.specs))
		var rootIDsByProgram [][]string
		if g.rootIdentities != nil {
			rootIDsByProgram = g.rootIdentities.FreezePerProgram(binding.Programs)
		}
		for setIndex, prog := range binding.Programs {
			specIndex := setIndex
			if setIndex < len(set.planIndexes) {
				specIndex = set.planIndexes[setIndex]
			}
			boundTargets := binding.TargetsByProgram[setIndex]
			var rootIDs []string
			if setIndex < len(rootIDsByProgram) {
				rootIDs = rootIDsByProgram[setIndex]
			}
			g.bySpec[specIndex] = LoadResult{
				Programs:               []*lintprogram.Program{prog},
				TargetsByProgram:       [][]string{boundTargets},
				LintTargetBySourcePath: mappingsByProgram[setIndex],
				ProgramRootIdentities:  rootIDs,
			}
		}
	})
	return g.err
}

func (g *coupledLoadGroup) takeSpec(specIndex int) (LoadResult, error) {
	if err := g.ensureBuilt(); err != nil {
		return LoadResult{}, err
	}
	g.mu.Lock()
	var result LoadResult
	if specIndex < len(g.bySpec) {
		result = g.bySpec[specIndex]
		g.bySpec[specIndex] = LoadResult{}
	}
	g.mu.Unlock()
	return result, nil
}

func (g *coupledLoadGroup) takeUnbound() (LoadResult, error) {
	if err := g.ensureBuilt(); err != nil {
		return LoadResult{}, err
	}
	g.mu.Lock()
	files := g.unbound
	g.unbound = nil
	g.mu.Unlock()
	if len(files) == 0 {
		return LoadResult{}, nil
	}
	return g.session.loadUnboundTargets(files, g.cwd, g.compilerRoots, g.singleThreaded)
}

func (s *Session) loadUnboundTargets(
	files []target.File,
	cwd string,
	compilerRoots bool,
	singleThreaded bool,
) (LoadResult, error) {
	var err error
	binding := LoadResult{LintTargetBySourcePath: make(map[string]target.File)}
	for _, file := range files {
		storeSourceTargetMapping(binding.LintTargetBySourcePath, file.Path, file.CanonicalPath, file)
	}
	useCaseSensitive := true
	if fsys := s.FS(); fsys != nil {
		useCaseSensitive = fsys.UseCaseSensitiveFileNames()
	}
	if !compilerRoots && allRootsSupportedByParser(files, useCaseSensitive) {
		err = s.appendRootPrograms(&binding, groupUnboundTargets(files, cwd, useCaseSensitive), cwd, singleThreaded)
	} else {
		err = s.appendCompatibilityPrograms(&binding, files, cwd, singleThreaded)
	}
	finalizeResult(&binding)
	return binding, err
}

func (execution *targetedProjectExecution) loadTargetsByProject(request ProjectBuildRequest) (map[int][]target.File, []int, error) {
	plan := execution.plan
	files := request.Targets.Files
	filesByProject := make(map[int][]target.File)
	if request.SkipLintBinding {
		// Discovery has already selected the service owners. The complete
		// Program set only needs to validate their selected sources.
		owners := make([]int, len(files))
		for index := range owners {
			owners[index] = -1
		}
		for index, file := range files {
			indexes := plan.targetProjects[file]
			if request.Policies[file].ServiceRootDirectory != "" && len(indexes) > 0 {
				owner := indexes[0]
				owners[index] = owner
				filesByProject[owner] = append(filesByProject[owner], file)
			}
		}
		return filesByProject, owners, nil
	}
	groups := groupTargetsByProjects(files, plan.targetProjects, func(owner string) []int {
		return orderedProjectIndexesForConfig(plan, owner)
	})
	var owners []int
	if request.Scope == AllDeclared {
		// Use the complete build's batch identity ranking and preserve
		// authoritative service selections, including missing sources.
		owners = execution.activeRootOwners(files, groups)
	} else {
		var err error
		owners, err = execution.selectRootOwners(files, request.Scope, groups, nil)
		if err != nil {
			return nil, nil, err
		}
	}
	for index, file := range files {
		owner := owners[index]
		filesByProject[owner] = append(filesByProject[owner], file)
	}
	return filesByProject, owners, nil
}

const maxPlanReferenceCacheEntries = 8

type planReferenceEntry struct {
	name  string
	child *tsoptions.ParsedCommandLine
	err   error
}

// planReferenceCache bounds retained referenced tsconfig graphs within a single
// PlanLoads call so workspaces with many one-use composite libraries do not
// spike peak memory while still reusing hot shared references.
type planReferenceCache struct {
	mu  sync.Mutex
	lru []planReferenceEntry // most recently used at index 0
}

func (c *planReferenceCache) load(
	name string,
	parse func(string) (*tsoptions.ParsedCommandLine, error),
) (*tsoptions.ParsedCommandLine, error) {
	c.mu.Lock()
	for i := range c.lru {
		if c.lru[i].name == name {
			hit := c.lru[i]
			if i > 0 {
				copy(c.lru[1:i+1], c.lru[:i])
				c.lru[0] = hit
			}
			c.mu.Unlock()
			return hit.child, hit.err
		}
	}
	c.mu.Unlock()

	child, err := parse(name)

	c.mu.Lock()
	item := planReferenceEntry{name: name, child: child, err: err}
	found := -1
	for i := range c.lru {
		if c.lru[i].name == name {
			found = i
			break
		}
	}
	switch {
	case found >= 0:
		if found > 0 {
			copy(c.lru[1:found+1], c.lru[:found])
		}
		c.lru[0] = item
	case len(c.lru) < maxPlanReferenceCacheEntries:
		c.lru = append(c.lru, planReferenceEntry{})
		copy(c.lru[1:], c.lru[:len(c.lru)-1])
		c.lru[0] = item
	default:
		// Overwrite the evicted tail slot so GC can reclaim the full
		// ParsedCommandLine and its declaration maps once no active visit holds it.
		copy(c.lru[1:], c.lru[:len(c.lru)-1])
		c.lru[0] = item
	}
	c.mu.Unlock()
	return child, err
}

// Metadata suffices only for direct roots accepted by the compiler's
// extension policy. Aliases and content mappers retain the complete binding
// path, whose lookup can resolve their source identity across projects.
func (s *Session) hasIndependentRoots(parsed *tsoptions.ParsedCommandLine, files []target.File, sourceReferences bool, refCache *planReferenceCache) bool {
	if len(files) == 0 {
		// With no sources to bind or validate, the Program can be constructed
		// independently with its complete roots, references and options.
		return true
	}
	if len(parsed.ContentMapperExtensions()) != 0 {
		return false
	}
	options := parsed.CompilerOptions()
	supported := tsoptions.GetSupportedExtensionsWithJsonIfResolveJsonModule(options, tsoptions.GetSupportedExtensions(options, nil))
	useCaseSensitive := s.FS().UseCaseSensitiveFileNames()
	roots := make(map[string]struct{}, len(parsed.FileNames()))
	identities := make(map[tspath.Path]string, len(parsed.FileNames()))
	for _, name := range parsed.FileNames() {
		identity := tspath.ToPath(name, "", useCaseSensitive)
		if previous, exists := identities[identity]; exists && previous != name {
			return false
		}
		identities[identity] = name
		roots[name] = struct{}{}
	}
	for _, file := range files {
		if _, direct := roots[file.Path]; !direct {
			return false
		}
		if !sourceReferences && !options.AllowNonTsExtensions.IsTrue() {
			accepted := false
			name := tspath.GetCanonicalFileName(file.Path, useCaseSensitive)
			for _, extensions := range supported {
				accepted = accepted || tspath.FileExtensionIsOneOf(name, extensions)
			}
			if !accepted {
				return false
			}
		}
	}
	useSourceRedirect := sourceReferences && !options.DisableSourceOfProjectReferenceRedirect.IsTrue()
	if len(parsed.ProjectReferences()) != 0 {
		if !useSourceRedirect {
			selected := make(map[tspath.Path]struct{}, len(files))
			for _, file := range files {
				selected[tspath.ToPath(file.Path, "", useCaseSensitive)] = struct{}{}
			}
			return s.referencesLeaveRootsUnchanged(parsed, selected, false, refCache)
		}
		preserveSymlinks := options.PreserveSymlinks.IsTrue()
		var selected map[tspath.Path]struct{}
		for _, file := range files {
			isDecl := tspath.IsDeclarationFileName(tspath.GetCanonicalFileName(file.Path, useCaseSensitive))
			isNodeModulesSymlink := preserveSymlinks && strings.Contains(file.Path, "/node_modules/")
			if !isDecl && !isNodeModulesSymlink {
				continue
			}
			if selected == nil {
				selected = make(map[tspath.Path]struct{})
			}
			selected[tspath.ToPath(file.Path, "", useCaseSensitive)] = struct{}{}
			if isNodeModulesSymlink {
				if file.CanonicalPath != "" {
					selected[tspath.ToPath(file.CanonicalPath, "", useCaseSensitive)] = struct{}{}
				}
				if resolved := s.FS().Realpath(file.Path); resolved != "" {
					selected[tspath.ToPath(resolved, "", useCaseSensitive)] = struct{}{}
				}
			}
		}
		if len(selected) != 0 {
			return s.referencesLeaveRootsUnchanged(parsed, selected, true, refCache)
		}
	}
	return true
}

// Project references can redirect a listed root in either reference mode:
// without source redirection (ordinary configs, or projectService configs with
// disableSourceOfProjectReferenceRedirect), a referenced source root redirects
// to a .d.ts output; with source redirection enabled, a referenced .d.ts output
// (or a /node_modules/ symlink to one under preserveSymlinks) redirects back to
// its .ts source. Their presence alone does not couple unrelated lint bindings:
// only roots present in the recursive reference metadata can be redirected.
// This conservative membership check never constructs a Program or changes its
// reference mode. Unknown metadata retains the complete binding path and its
// authoritative validation/error order.
func (s *Session) referencesLeaveRootsUnchanged(
	parsed *tsoptions.ParsedCommandLine,
	selected map[tspath.Path]struct{},
	useSourceRedirect bool,
	refCache *planReferenceCache,
) bool {
	useCaseSensitive := s.FS().UseCaseSensitiveFileNames()
	root := tspath.ToPath(parsed.ConfigName(), "", useCaseSensitive)
	seen := make(map[tspath.Path]struct{})
	parseRef := func(name string) (*tsoptions.ParsedCommandLine, error) {
		return s.context.parseConfig(tspath.GetDirectoryPath(name), name)
	}
	loadRef := func(name string) (*tsoptions.ParsedCommandLine, error) {
		if refCache == nil {
			return parseRef(name)
		}
		return refCache.load(name, parseRef)
	}
	var visit func(*tsoptions.ParsedCommandLine) bool
	visit = func(parent *tsoptions.ParsedCommandLine) bool {
		for _, name := range parent.ResolvedProjectReferencePaths() {
			path := tspath.ToPath(name, "", useCaseSensitive)
			// Conservatively fall back on a reference cycle back to the root
			// config (ordinary construction can reparse the root with a
			// distinct config identity).
			if path == root {
				return false
			}
			if _, visited := seen[path]; visited {
				continue
			}
			seen[path] = struct{}{}
			child, err := loadRef(name)
			if err != nil || child == nil || len(child.ContentMapperExtensions()) != 0 {
				return false
			}
			if useSourceRedirect {
				child.ParseInputOutputNames()
				outputToRef := child.OutputDtsToProjectReference()
				if len(outputToRef) > 0 {
					if len(selected) <= len(outputToRef) {
						for selectedPath := range selected {
							if outputToRef[selectedPath] != nil {
								return false
							}
						}
					} else {
						for outputPath := range outputToRef {
							if _, overlaps := selected[outputPath]; overlaps {
								return false
							}
						}
					}
				}
			} else {
				childDir := child.GetCurrentDirectory()
				parsedOutputNames := false
				for _, fileName := range child.FileNames() {
					childPath := tspath.ToPath(fileName, childDir, useCaseSensitive)
					if _, overlaps := selected[childPath]; !overlaps {
						continue
					}
					if tspath.IsDeclarationFileName(fileName) || tspath.FileExtensionIs(fileName, tspath.ExtensionJson) {
						continue
					}
					if !parsedOutputNames {
						child.ParseInputOutputNames()
						parsedOutputNames = true
					}
					if info := child.SourceToProjectReference()[childPath]; info != nil && info.OutputDts != "" {
						return false
					}
				}
			}
			if !visit(child) {
				return false
			}
		}
		return true
	}
	return visit(parsed)
}
