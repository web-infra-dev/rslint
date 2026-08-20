package config

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// ProjectDeclarationState preserves the authored parserOptions.project shape.
// Unspecified and ExplicitEmpty intentionally retain the current rslint
// behavior in this change: both probe the governing config directory's default
// tsconfig.json. Keeping them distinct prevents the config pipeline from
// erasing information needed by a future explicitly versioned product change.
type ProjectDeclarationState uint8

const (
	ProjectUnspecified ProjectDeclarationState = iota
	ProjectExplicitEmpty
	ProjectExplicit
)

// ProjectState reports how parserOptions.project was supplied by the final
// matched config entry shape.
func (plan *EffectiveFileConfig) ProjectState() ProjectDeclarationState {
	if plan == nil || plan.projectEntry < 0 {
		return ProjectUnspecified
	}
	projects := plan.resolver.config[plan.projectEntry].LanguageOptions.ParserOptions.Project
	if len(projects) == 0 {
		return ProjectExplicitEmpty
	}
	return ProjectExplicit
}

type projectOwnerPlan struct {
	directory       string
	resolver        *FileConfigResolver
	pathsByEntry    map[int][]string
	defaultProjects []string
	catalogProjects []string
}

// ProjectPathResolver is one immutable config-generation view. Construction
// validates every parserOptions.project declaration in each supplied owner,
// preserving the existing config-acceptance boundary. ResolveLintProjectPlan
// then projects only the declaration effective for each target.
type ProjectPathResolver struct {
	fs              vfs.FS
	owners          map[string]*projectOwnerPlan
	ownerByPhysical map[string]*projectOwnerPlan
	orderedOwners   []*projectOwnerPlan
}

// NewProjectPathResolver validates and snapshots either a multi-owner config
// map or one invocation-wide config. A non-nil configMap selects multi-owner
// mode; otherwise config belongs to currentDirectory.
func NewProjectPathResolver(
	configMap map[string]RslintConfig,
	config RslintConfig,
	currentDirectory string,
	fsys vfs.FS,
	enforcePlugins bool,
) (*ProjectPathResolver, error) {
	resolver := &ProjectPathResolver{
		fs:              fsys,
		owners:          make(map[string]*projectOwnerPlan),
		ownerByPhysical: make(map[string]*projectOwnerPlan),
	}
	if configMap == nil {
		owner, err := newProjectOwnerPlan(config, currentDirectory, fsys, enforcePlugins)
		if err != nil {
			return nil, fmt.Errorf("resolve tsconfigs for %q: %w", currentDirectory, err)
		}
		resolver.addOwner(owner, fsys)
		return resolver, nil
	}

	normalizedConfigs := make(map[string]RslintConfig, len(configMap))
	directories := make([]string, 0, len(configMap))
	for directory, entries := range configMap {
		directory = tspath.NormalizePath(directory)
		if _, exists := normalizedConfigs[directory]; !exists {
			directories = append(directories, directory)
		}
		normalizedConfigs[directory] = entries
	}
	sort.Strings(directories)
	for _, directory := range directories {
		owner, err := newProjectOwnerPlan(normalizedConfigs[directory], directory, fsys, enforcePlugins)
		if err != nil {
			return nil, fmt.Errorf("resolve tsconfigs for %q: %w", directory, err)
		}
		resolver.addOwner(owner, fsys)
	}
	return resolver, nil
}

func (resolver *ProjectPathResolver) addOwner(owner *projectOwnerPlan, fsys vfs.FS) {
	resolver.owners[exactPathID(owner.directory)] = owner
	physicalID := canonicalPathID(owner.directory, fsys)
	if existing, ok := resolver.ownerByPhysical[physicalID]; !ok || existing == owner {
		resolver.ownerByPhysical[physicalID] = owner
	} else {
		// Discovery rejects ambiguous physical owners before this stage. Retain a
		// nil sentinel defensively so a future direct caller cannot pick one.
		resolver.ownerByPhysical[physicalID] = nil
	}
	resolver.orderedOwners = append(resolver.orderedOwners, owner)
}

func newProjectOwnerPlan(
	config RslintConfig,
	directory string,
	fsys vfs.FS,
	enforcePlugins bool,
) (*projectOwnerPlan, error) {
	directory = tspath.NormalizePath(directory)
	_, matchDirectory := ResolveConfigPathSpace(directory, directory, fsys)
	fileResolver := newFileConfigResolver(config, matchDirectory, enforcePlugins)
	owner := &projectOwnerPlan{
		directory:    directory,
		resolver:     fileResolver,
		pathsByEntry: make(map[int][]string),
	}
	loader := NewConfigLoader(fsys, directory)
	resolutionByValue := make(map[string][]string)
	seenCatalog := make(map[string]struct{})
	for entryIndex, entry := range config {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil ||
			entry.LanguageOptions.ParserOptions.Project == nil {
			continue
		}
		projects := entry.LanguageOptions.ParserOptions.Project
		key := projectPathsKey(projects)
		paths, cached := resolutionByValue[key]
		if !cached {
			if fsys == nil {
				paths = nil
				resolutionByValue[key] = paths
				owner.pathsByEntry[entryIndex] = paths
				continue
			}
			var err error
			paths, err = loader.resolveProjectPaths(projects, directory)
			if err != nil {
				return nil, err
			}
			paths = immutableStrings(paths)
			resolutionByValue[key] = paths
		}
		owner.pathsByEntry[entryIndex] = paths
		for _, path := range paths {
			pathID := exactPathID(path)
			if _, seen := seenCatalog[pathID]; seen {
				continue
			}
			seenCatalog[pathID] = struct{}{}
			owner.catalogProjects = append(owner.catalogProjects, path)
		}
	}

	defaultPath := tspath.ResolvePath(directory, "tsconfig.json")
	if fsys != nil && fsys.FileExists(defaultPath) {
		owner.defaultProjects = []string{tspath.NormalizePath(defaultPath)}
	}
	// Preserve the existing owner-wide type-check behavior: use the default
	// only when no authored declaration resolved to a project for this owner.
	if len(owner.catalogProjects) == 0 {
		owner.catalogProjects = owner.defaultProjects
	}
	owner.defaultProjects = immutableStrings(owner.defaultProjects)
	owner.catalogProjects = immutableStrings(owner.catalogProjects)
	return owner, nil
}

func projectPathsKey(paths ProjectPaths) string {
	var key strings.Builder
	for _, path := range paths {
		fmt.Fprintf(&key, "%d:%s;", len(path), path)
	}
	return key.String()
}

func immutableStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := append([]string(nil), values...)
	return cloned[:len(cloned):len(cloned)]
}

func (resolver *ProjectPathResolver) owner(directory string) *projectOwnerPlan {
	if resolver == nil {
		return nil
	}
	if owner := resolver.owners[exactPathID(directory)]; owner != nil {
		return owner
	}
	return resolver.ownerByPhysical[canonicalPathID(directory, resolver.fs)]
}

func (owner *projectOwnerPlan) projectsFor(plan *EffectiveFileConfig) ([]string, error) {
	if plan == nil {
		return nil, nil
	}
	if plan.resolver != owner.resolver {
		return nil, fmt.Errorf("effective config does not belong to project owner %q", owner.directory)
	}
	if plan.projectEntry < 0 {
		return owner.defaultProjects, nil
	}
	projects := owner.resolver.config[plan.projectEntry].LanguageOptions.ParserOptions.Project
	if len(projects) == 0 {
		return owner.defaultProjects, nil
	}
	return owner.pathsByEntry[plan.projectEntry], nil
}

// PlannedLintTarget freezes the config and ordered project declarations that
// apply to one discovered target. All fields are immutable after planning.
type PlannedLintTarget struct {
	Target       DiscoveredLintTarget
	MatchPath    string
	Effective    *EffectiveFileConfig
	ProjectPaths []string
}

// LintProjectPlan is the request-wide semantic input shared by CLI, API, and
// LSP project selection. Candidate lists remain target-specific; consumers may
// share their backing storage but must never merge their declaration order.
type LintProjectPlan struct {
	Targets []PlannedLintTarget
}

// ResolveLintProjectPlan resolves effective configs and project candidates for
// exactly the already-discovered lint targets.
func (resolver *ProjectPathResolver) ResolveLintProjectPlan(
	targetPlan LintTargetPlan,
	singleThreaded bool,
) (LintProjectPlan, error) {
	plan := LintProjectPlan{Targets: make([]PlannedLintTarget, len(targetPlan.Targets))}
	workerCount := min(runtime.GOMAXPROCS(0), len(targetPlan.Targets))
	if singleThreaded || workerCount <= 1 {
		for index := range targetPlan.Targets {
			planned, err := resolver.ResolveLintTarget(targetPlan.Targets[index])
			if err != nil {
				return LintProjectPlan{}, err
			}
			plan.Targets[index] = planned
		}
		return plan, nil
	}

	errs := make([]error, len(targetPlan.Targets))
	jobs := make(chan int, workerCount)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				plan.Targets[index], errs[index] = resolver.ResolveLintTarget(targetPlan.Targets[index])
			}
		}()
	}
	for index := range targetPlan.Targets {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	for _, err := range errs {
		if err != nil {
			return LintProjectPlan{}, err
		}
	}
	return plan, nil
}

// ResolveLintTarget resolves one already-owned target without rediscovering
// files or config ownership. LSP uses this against a committed config
// generation; batch callers normally use ResolveLintProjectPlan.
func (resolver *ProjectPathResolver) ResolveLintTarget(target DiscoveredLintTarget) (PlannedLintTarget, error) {
	target = normalizeLintTarget(target, target.ConfigDirectory, resolver.fs)
	owner := resolver.owner(target.ConfigDirectory)
	if owner == nil {
		return PlannedLintTarget{}, fmt.Errorf("no project owner for lint target %q", target.Path)
	}
	matchPath, _ := ResolveConfigPathSpaceWithCanonical(
		target.Path,
		target.CanonicalPath,
		target.ConfigDirectory,
		resolver.fs,
	)
	effective := owner.resolver.PlanForFile(matchPath)
	projects, err := owner.projectsFor(effective)
	if err != nil {
		return PlannedLintTarget{}, err
	}
	return PlannedLintTarget{
		Target:       target,
		MatchPath:    matchPath,
		Effective:    effective,
		ProjectPaths: projects,
	}, nil
}

// CatalogProjectPaths returns the full type-check project catalog, applying
// default-tsconfig probing independently for each config owner.
func (resolver *ProjectPathResolver) CatalogProjectPaths() []string {
	if resolver == nil {
		return nil
	}
	var projects []string
	seen := make(map[string]struct{})
	for _, owner := range resolver.orderedOwners {
		for _, project := range owner.catalogProjects {
			projectID := exactPathID(project)
			if _, exists := seen[projectID]; exists {
				continue
			}
			seen[projectID] = struct{}{}
			projects = append(projects, project)
		}
	}
	return projects
}
