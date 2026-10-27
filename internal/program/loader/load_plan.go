package loader

import (
	"fmt"
	"runtime"

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
// loads may run concurrently. Returned Programs belong solely to the caller.
type ProgramLoad struct {
	TargetCount int
	Options     *core.CompilerOptions
	Load        func() (LoadResult, error)
}

// PlanLoads separates target ownership from source construction. A nil plan
// keeps the whole-binding acquisition. Compiler-only roots preserve consumers
// that require the compiler AST representation; otherwise the parser/binder
// facade handles supported project-external roots.
//
// Call on a fresh Session, before constructing any Programs. Config validation
// and root identity ranking are shared with BuildProjects. The loads retain
// immutable metadata and source snapshots, but no completed Program or AST.
func (s *Session) PlanLoads(request ProjectBuildRequest, cwd string, compilerRoots bool) ([]ProgramLoad, error) {
	if request.Scope == AllDeclared {
		return nil, nil
	}
	plan, err := s.prepareProjectPlan(request)
	if err != nil {
		return nil, err
	}
	if len(plan.specs) < 2 {
		return nil, nil
	}
	execution := newTargetedProjectExecution(s, plan, request.SingleThreaded)
	if request.Scope == ActiveOwners {
		if err := execution.validateActiveConfigs(); err != nil {
			return nil, err
		}
	}
	if plan.terminalErr != nil {
		return nil, plan.terminalErr
	}
	groups := groupTargetsByProjects(request.Targets.Files, plan.targetProjects, func(owner string) []int {
		return orderedProjectIndexesForConfig(plan, owner)
	})
	owners, err := execution.selectRootOwners(request.Targets.Files, request.Scope, groups, nil)
	if err != nil {
		return nil, err
	}
	filesByProject := make(map[int][]target.File)
	for index, file := range request.Targets.Files {
		owner := owners[index]
		filesByProject[owner] = append(filesByProject[owner], file)
	}
	projectCount := len(filesByProject)
	if _, hasRoots := filesByProject[-1]; hasRoots {
		projectCount--
	}
	if projectCount < 2 {
		return nil, nil
	}
	// Ordinary roots can require ordered source-membership fallback across
	// projects (for example a listed .js root without allowJs). Resolve that
	// coherent binding before exposing any executable units, not by retrying
	// an observation after some native or plugin rules have already run.
	// Preserve the materialized binding's declaration order, with unmatched
	// roots last, independently of target enumeration and completion timing.
	order := make([]int, 0, len(filesByProject))
	for owner := range plan.specs {
		if len(filesByProject[owner]) > 0 {
			order = append(order, owner)
		}
	}
	if len(filesByProject[-1]) > 0 {
		order = append(order, -1)
	}
	// Resolve metadata eligibility before reading source or publishing loads.
	independent := make([]bool, len(order))
	errs := make([]error, len(order))
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
				independent[i] = s.hasIndependentRoots(slot.config, filesByProject[owner], plan.specs[owner].sourceReferences)
			}
		})
	}
	work.RunAndWait()
	for i := range order {
		if errs[i] != nil {
			return nil, errs[i]
		}
		if !independent[i] {
			return nil, nil
		}
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
		if owner >= 0 {
			for _, file := range filesByProject[owner] {
				roots = append(roots, file.Path)
			}
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
	loads := make([]ProgramLoad, 0, len(order))
	for _, owner := range order {
		files := filesByProject[owner]
		if owner < 0 {
			// Keep all unmatched roots together, including cross-file binding
			// and the existing case-collision partition. Never shard their ASTs.
			loads = append(loads, ProgramLoad{
				TargetCount: len(files),
				Load: func() (LoadResult, error) {
					var err error
					binding := LoadResult{LintTargetBySourcePath: make(map[string]target.File)}
					for _, file := range files {
						storeSourceTargetMapping(binding.LintTargetBySourcePath, file.Path, file.CanonicalPath, file)
					}
					if !compilerRoots && allRootsSupportedByParser(files, s.FS().UseCaseSensitiveFileNames()) {
						err = s.appendRootPrograms(&binding, groupUnboundTargets(files, cwd, s.FS().UseCaseSensitiveFileNames()), cwd, request.SingleThreaded)
					} else {
						err = s.appendCompatibilityPrograms(&binding, files, cwd, request.SingleThreaded)
					}
					finalizeResult(&binding)
					return binding, err
				},
			})
			continue
		}
		spec := plan.specs[owner]
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
				set := ProjectSet{
					compilerPrograms: []*compiler.Program{cp},
					configOrders:     []configOrders{spec.configOrders},
					targetBinding:    &projectTargetBinding{targets: files, owners: make([]int, len(files))},
					targetProjects:   make(map[target.File][]int, len(files)),
				}
				set.programs = lintprogram.NewOwnedFromCompilers(set.compilerPrograms)
				for _, file := range files {
					set.targetProjects[file] = []int{0}
				}
				binding, unbound := s.bindTargetsToProjects(set, target.Plan{Files: files}, request.SingleThreaded)
				if len(unbound) > 0 {
					return LoadResult{}, fmt.Errorf("project root %q from %q was absent from its TypeScript Program", unbound[0].Path, spec.tsconfigPath)
				}
				finalizeResult(&binding)
				return binding, nil
			},
		})
	}
	return loads, nil
}

// Metadata suffices only for direct roots accepted by the compiler's
// extension policy. Aliases and content mappers retain the complete binding
// path, whose lookup can resolve their source identity across projects.
func (s *Session) hasIndependentRoots(parsed *tsoptions.ParsedCommandLine, files []target.File, sourceReferences bool) bool {
	if len(parsed.ContentMapperExtensions()) != 0 || (!sourceReferences && len(parsed.ProjectReferences()) != 0) {
		return false
	}
	options := parsed.CompilerOptions()
	supported := tsoptions.GetSupportedExtensionsWithJsonIfResolveJsonModule(options, tsoptions.GetSupportedExtensions(options, nil))
	roots := make(map[string]struct{}, len(parsed.FileNames()))
	identities := make(map[tspath.Path]string, len(parsed.FileNames()))
	for _, name := range parsed.FileNames() {
		identity := tspath.ToPath(name, "", s.FS().UseCaseSensitiveFileNames())
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
			name := tspath.GetCanonicalFileName(file.Path, s.FS().UseCaseSensitiveFileNames())
			for _, extensions := range supported {
				accepted = accepted || tspath.FileExtensionIsOneOf(name, extensions)
			}
			if !accepted {
				return false
			}
		}
	}
	return true
}
