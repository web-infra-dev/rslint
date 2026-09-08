package config

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// ProjectPolicy is the effective project configuration for one lint target.
// Project paths are absolute (and may contain globs); discovery keeps the
// target's lexical path. Resolving this policy never resolves lint rules.
type ProjectPolicy struct {
	ProjectService  bool
	TsconfigRootDir string
	Project         ProjectPaths
	ProjectDisabled bool
}

// NeedsProjectPolicy distinguishes per-file project settings from the legacy
// config-wide explicit-project path. The latter retains its declaration order.
func NeedsProjectPolicy(config RslintConfig) bool {
	for _, entry := range config {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}
		options := entry.LanguageOptions.ParserOptions
		if options.ProjectService != nil || options.TsconfigRootDir != "" || options.rootDirSet || options.ProjectDisabled || options.projectAutomatic {
			return true
		}
	}
	return false
}

type ProjectPolicyResolver struct {
	config          RslintConfig
	configDirectory string
	matcher         *configTargetResolver
}

func NewProjectPolicyResolver(config RslintConfig, configDirectory string, fsys vfs.FS) *ProjectPolicyResolver {
	return &ProjectPolicyResolver{config, configDirectory, newConfigTargetResolver(config, configDirectory, fsys)}
}

// NewProjectPolicyResolverWithPathSpaces shares the exact config-match
// generation already frozen by target discovery or an editor snapshot.
func NewProjectPolicyResolverWithPathSpaces(
	config RslintConfig,
	configDirectory string,
	fsys vfs.FS,
	pathSpaces *PathSpaceSnapshot,
) (*ProjectPolicyResolver, error) {
	matcher, err := NewTargetMatcherWithPathSpaces(config, configDirectory, fsys, pathSpaces)
	if err != nil {
		return nil, err
	}
	return &ProjectPolicyResolver{config, configDirectory, matcher.resolver}, nil
}

// CanLoadDeclaredProjectsWithoutTargets preserves program-wide checking for
// legacy explicit projects, including earlier entries whose service setting is later
// disabled for every target. Scoped overrides still require target matching.
func (resolver *ProjectPolicyResolver) CanLoadDeclaredProjectsWithoutTargets() bool {
	if !NeedsProjectPolicy(resolver.config) {
		return true
	}
	hasProjects := false
	mayEnableService := false
	for index, entry := range resolver.config {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}
		options := entry.LanguageOptions.ParserOptions
		// Do not fold project declarations: their legacy union cannot restore
		// projects removed by an explicit clear, even after a later override.
		if options.TsconfigRootDir != "" || options.rootDirSet || options.ProjectDisabled || options.projectAutomatic || options.Project != nil && len(options.Project) == 0 {
			return false
		}
		hasProjects = hasProjects || options.Project != nil
		if options.ProjectService == nil {
			continue
		}
		if *options.ProjectService {
			mayEnableService = true
			continue
		}
		prepared := resolver.matcher.entries[index]
		// These are precisely the entry predicates used by resolveTarget.
		// With none present, false applies in every authored path space. The
		// original entries remain intact for project-path origin resolution.
		if !hasFileSelectors(entry) && len(prepared.ignorePatterns) == 0 && prepared.configArrayBaseIndex < 0 {
			mayEnableService = false
		}
	}
	return hasProjects && !mayEnableService
}

func (resolver *ProjectPolicyResolver) Resolve(target PathIdentity) (ProjectPolicy, error) {
	decision := resolver.matcher.resolveTarget(target)
	if !decision.selected || decision.globallyIgnored {
		return ProjectPolicy{ProjectDisabled: true}, nil
	}
	var merged *LanguageOptions
	defaultRoot := resolver.configDirectory
	projectBase := resolver.configDirectory
	for index, entry := range resolver.config {
		if !decision.key.contains(index) || entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}
		options := entry.LanguageOptions.ParserOptions
		origin := configEntryPathOrigin(entry, resolver.configDirectory)
		if options.ProjectService != nil {
			defaultRoot = origin.directory
			if origin.basePathScoped {
				defaultRoot = origin.configArrayBase
			}
		}
		if options.Project != nil {
			projectBase = origin.directory
		}
		// Only parser options belong to this phase. Raw plugin language options
		// and rules are merged later by the file-config resolver.
		merged = mergeLanguageOptions(merged, &LanguageOptions{ParserOptions: options})
	}
	policy := ProjectPolicy{TsconfigRootDir: tspath.NormalizePath(defaultRoot)}
	if merged == nil || merged.ParserOptions == nil {
		return policy, nil
	}
	options := merged.ParserOptions
	policy.ProjectService = options.ProjectService != nil && *options.ProjectService
	policy.ProjectDisabled = options.ProjectDisabled
	if options.TsconfigRootDir != "" {
		if !tspath.IsRootedDiskPath(options.TsconfigRootDir) {
			return ProjectPolicy{}, fmt.Errorf("parserOptions.tsconfigRootDir must be an absolute path: %q", options.TsconfigRootDir)
		}
		policy.TsconfigRootDir = tspath.NormalizePath(options.TsconfigRootDir)
		projectBase = policy.TsconfigRootDir
	}
	if policy.ProjectService && (options.Project != nil || options.projectAutomatic) {
		return ProjectPolicy{}, fmt.Errorf("%s: enabling parserOptions.project does nothing when projectService is enabled; remove project or set projectService to false", target.Path)
	}
	if options.projectAutomatic {
		return ProjectPolicy{}, fmt.Errorf("%s: parserOptions.project: true is not supported; use projectService: true for automatic discovery", target.Path)
	}
	if options.Project != nil {
		policy.Project = make(ProjectPaths, len(options.Project))
		for index, path := range options.Project {
			policy.Project[index] = tspath.ResolvePath(projectBase, path)
		}
	} else if options.ProjectService != nil && !policy.ProjectService {
		policy.ProjectDisabled = true
	}
	return policy, nil
}
