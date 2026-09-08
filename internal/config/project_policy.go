package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// ProjectPolicy projects project-service options from an already resolved
// target config. Ordinary project declarations remain owned by the raw config.
// The zero value preserves ordinary project loading and its default fallback.
type ProjectPolicy struct {
	ProjectService         bool
	DefaultProjectDisabled bool
	ProjectDisabled        bool
	// TsconfigRootDir is an explicit absolute override. An empty value leaves
	// service discovery at the config owner and project paths at their authored bases.
	TsconfigRootDir string
}

// HasProjectOptions identifies configs that need per-target project policy.
// Ordinary project strings and arrays continue through the existing loader.
func HasProjectOptions(entries RslintConfig) bool {
	for _, entry := range entries {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}
		options := entry.LanguageOptions.ParserOptions
		if options.ProjectService != nil || options.TsconfigRootDir != nil || options.rootDirSet ||
			options.ProjectDisabled || options.projectAutomatic {
			return true
		}
	}
	return false
}

// ResolveProjectPolicy is a pure projection of the same effective config used
// for rules and plugins. Matching and option merging have already happened.
func ResolveProjectPolicy(resolved ResolvedFileConfig) (ProjectPolicy, error) {
	merged := resolved.MergedConfig
	if merged == nil || merged.LanguageOptions == nil || merged.LanguageOptions.ParserOptions == nil {
		return ProjectPolicy{}, nil
	}
	options := merged.LanguageOptions.ParserOptions
	policy := ProjectPolicy{}
	if options.ProjectService != nil {
		policy.ProjectService = *options.ProjectService
		policy.DefaultProjectDisabled = !policy.ProjectService
	}
	if options.TsconfigRootDir != nil {
		root := *options.TsconfigRootDir
		absolute := filepath.IsAbs(root)
		if runtime.GOOS == "windows" {
			absolute = absolute && len(root) >= 2 && root[1] == ':' &&
				((root[0] >= 'a' && root[0] <= 'z') || (root[0] >= 'A' && root[0] <= 'Z'))
		}
		if !absolute {
			return ProjectPolicy{}, fmt.Errorf("parserOptions.tsconfigRootDir must be an absolute path: %q", *options.TsconfigRootDir)
		}
		policy.TsconfigRootDir = tspath.NormalizePath(filepath.Clean(root))
	}
	if policy.ProjectService && (options.Project != nil || options.projectAutomatic) {
		return ProjectPolicy{}, errors.New("enabling parserOptions.project does nothing when projectService is enabled; remove project or set projectService to false")
	}
	if options.projectAutomatic {
		return ProjectPolicy{}, errors.New("parserOptions.project: true is not supported; use projectService: true for automatic discovery")
	}
	policy.ProjectDisabled = options.ProjectDisabled && !policy.ProjectService
	return policy, nil
}
