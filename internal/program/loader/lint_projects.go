package loader

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/program/projectservice"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ProjectLoadScope controls how broadly legacy declarations are constructed.
// Automatic project discovery always starts from the selected target paths.
type ProjectLoadScope uint8

const (
	ProjectScopeTarget ProjectLoadScope = iota
	ProjectScopeOwner
	ProjectScopeAll
)

// BuildLintProjects selects projects using the effective policy of each target.
// Legacy configs retain config-wide declaration order and broad-load behavior;
// projectService discovers ownership from each target's lexical location.
func (s *Session) BuildLintProjects(
	configs map[string]rslintconfig.RslintConfig,
	plan target.Plan,
	cwd string,
	singleThreaded bool,
	scope ProjectLoadScope,
) (ProjectSet, error) {
	if err := s.validate(); err != nil {
		return ProjectSet{}, err
	}
	usesPolicy := false
	for _, config := range configs {
		usesPolicy = usesPolicy || rslintconfig.NeedsProjectPolicy(config)
	}
	if !usesPolicy {
		if scope == ProjectScopeAll {
			return s.BuildProjects(configs, singleThreaded)
		}
		if scope == ProjectScopeOwner {
			return s.BuildProjectsForTargetOwners(configs, plan, singleThreaded)
		}
		return s.BuildTargetProjects(configs, plan, singleThreaded)
	}

	type legacyGroup struct {
		config rslintconfig.RslintConfig
		owner  string
		files  []target.File
	}
	groups := make([]legacyGroup, 0)
	groupIndexes := make(map[string]int)
	resolvers := make(map[string]*rslintconfig.ProjectPolicyResolver)
	resolverForOwner := func(owner string) (*rslintconfig.ProjectPolicyResolver, error) {
		if resolver := resolvers[owner]; resolver != nil {
			return resolver, nil
		}
		resolver, err := rslintconfig.NewProjectPolicyResolverWithPathSpaces(configs[owner], owner, cwd, s.FS(), plan.PathSpaces())
		if err != nil {
			return nil, err
		}
		resolvers[owner] = resolver
		return resolver, nil
	}
	if scope == ProjectScopeAll {
		// Program-wide checking preserves inactive legacy owners as well.
		// Policy-mode owners need real targets to resolve their flat config.
		owners := make([]string, 0, len(configs))
		for owner := range configs {
			owners = append(owners, owner)
		}
		sort.Strings(owners)
		for _, owner := range owners {
			canLoadDeclarations := !rslintconfig.NeedsProjectPolicy(configs[owner])
			if !canLoadDeclarations {
				resolver, err := resolverForOwner(owner)
				if err != nil {
					return ProjectSet{}, err
				}
				canLoadDeclarations = resolver.CanLoadDeclaredProjectsWithoutTargets()
			}
			if canLoadDeclarations {
				groupIndexes[owner] = len(groups)
				groups = append(groups, legacyGroup{config: configs[owner], owner: owner})
			}
		}
	}
	policies := make([]rslintconfig.ProjectPolicy, len(plan.Files))
	for index, file := range plan.Files {
		config := configs[file.ConfigDirectory]
		groupKey := file.ConfigDirectory
		if rslintconfig.NeedsProjectPolicy(config) {
			resolver, err := resolverForOwner(file.ConfigDirectory)
			if err != nil {
				return ProjectSet{}, err
			}
			policy, err := resolver.Resolve(rslintconfig.PathIdentity{
				Path: file.Path, CanonicalPath: file.CanonicalPath, CanonicalParentPath: file.CanonicalParentPath,
			})
			if err != nil {
				return ProjectSet{}, err
			}
			policies[index] = policy
			if policy.ProjectService || policy.ProjectDisabled {
				continue
			}
			// Matching and authored-path resolution belong to config. Legacy
			// project construction receives only the effective declarations.
			config = rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{
				ParserOptions: &rslintconfig.ParserOptions{Project: policy.Project},
			}}}
			encoded, err := json.Marshal(policy.Project)
			if err != nil {
				return ProjectSet{}, err
			}
			groupKey += "\x00" + string(encoded)
		}
		groupIndex, found := groupIndexes[groupKey]
		if !found {
			groupIndex = len(groups)
			groupIndexes[groupKey] = groupIndex
			groups = append(groups, legacyGroup{config: config, owner: file.ConfigDirectory})
		}
		groups[groupIndex].files = append(groups[groupIndex].files, file)
	}

	set := ProjectSet{targetBinding: &projectTargetBinding{
		targets:  append([]target.File(nil), plan.Files...),
		owners:   make([]int, len(plan.Files)),
		complete: true,
	}}
	for index := range set.targetBinding.owners {
		set.targetBinding.owners[index] = -1
	}
	selector := projectservice.New(projectservice.Host{
		FS: s.FS(),
		ParseConfig: func(path string) (*tsoptions.ParsedCommandLine, error) {
			_, parsed, err := s.context.parseConfig(tspath.GetDirectoryPath(path), path)
			return parsed, err
		},
		CreateProgram: func(path string, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
			return utils.CreateProgramFromParsedConfigLenientWithProjectReferences(
				singleThreaded, parsed, s.context.newCompilerHostWithCache(tspath.GetDirectoryPath(path)),
			)
		},
	})
	programIndexes := make(map[*compiler.Program]int)
	for index, file := range plan.Files {
		policy := policies[index]
		if !policy.ProjectService {
			continue
		}
		selected, err := selector.Select(file.Path, policy.TsconfigRootDir)
		if err != nil {
			return ProjectSet{}, err
		}
		if selected.Program == nil {
			continue // The loader's existing gap path handles this unbound target.
		}
		programIndex, found := programIndexes[selected.Program]
		if !found {
			programIndex = len(set.compilerPrograms)
			programIndexes[selected.Program] = programIndex
			set.compilerPrograms = append(set.compilerPrograms, selected.Program)
			set.configOrders = append(set.configOrders, configOrders{})
		}
		set.configOrders[programIndex][exactPathID(file.ConfigDirectory)] = programIndex
		set.targetBinding.owners[index] = programIndex
	}

	targetIndexes := make(map[target.File]int, len(plan.Files))
	for index, file := range plan.Files {
		targetIndexes[file] = index
	}
	// Share existing construction slots across explicit policy groups. Their
	// declaration order and target scope remain independent; source-reference
	// Programs from the service selector never enter this cache.
	explicitSession := &Session{context: s.context, projectSlots: make(map[string]*targetedProjectSlot)}
	for _, group := range groups {
		groupPlan := plan
		groupPlan.Files = group.files
		groupConfigs := map[string]rslintconfig.RslintConfig{group.owner: group.config}
		var projects ProjectSet
		var err error
		if scope != ProjectScopeTarget {
			projects, err = explicitSession.BuildProjects(groupConfigs, singleThreaded)
		} else {
			projects, err = explicitSession.BuildTargetProjects(groupConfigs, groupPlan, singleThreaded)
		}
		if err != nil {
			return ProjectSet{}, fmt.Errorf("load explicit projects for %q: %w", group.owner, err)
		}
		binding, _ := s.bindTargetsToProjects(projects, groupPlan, singleThreaded)
		indexes := make([]int, len(projects.compilerPrograms))
		for index, program := range projects.compilerPrograms {
			programIndex, found := programIndexes[program]
			if !found {
				programIndex = len(set.compilerPrograms)
				programIndexes[program] = programIndex
				set.compilerPrograms = append(set.compilerPrograms, program)
				set.configOrders = append(set.configOrders, projects.configOrders[index])
			} else {
				for owner, order := range projects.configOrders[index] {
					if previous, found := set.configOrders[programIndex][owner]; !found || order < previous {
						set.configOrders[programIndex][owner] = order
					}
				}
			}
			indexes[index] = programIndex
		}
		for programIndex, sources := range binding.TargetsByProgram {
			for _, source := range sources {
				file := binding.LintTargetBySourcePath[source]
				set.targetBinding.owners[targetIndexes[file]] = indexes[programIndex]
			}
		}
	}
	set.programs = lintprogram.NewFromCompilers(set.compilerPrograms)
	return set, nil
}
