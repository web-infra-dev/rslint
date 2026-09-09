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
	// ServiceRootDirectory is the resolved absolute root for an enabled
	// project service. An empty value means service discovery is disabled.
	ServiceRootDirectory   string
	DefaultProjectDisabled bool
	ProjectDisabled        bool
	// TSConfigRootDirOverride rebases ordinary project declarations only when
	// an explicit root survives configuration merging. Omission and null leave
	// their authored bases intact; they do not copy the service default here.
	TSConfigRootDirOverride string
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

// ResolveProjectPolicy projects the same effective config used for rules and
// plugins. defaultRootDirectory is the loaded config module's directory, or
// the invocation cwd when no config file exists. It is independent of an
// entry's matching base and is resolved here before service consumers run.
func ResolveProjectPolicy(resolved ResolvedFileConfig, defaultRootDirectory string) (ProjectPolicy, error) {
	merged := resolved.MergedConfig
	if merged == nil || merged.LanguageOptions == nil || merged.LanguageOptions.ParserOptions == nil {
		return ProjectPolicy{}, nil
	}
	options := merged.LanguageOptions.ParserOptions
	policy := ProjectPolicy{}
	serviceEnabled := options.ProjectService != nil && *options.ProjectService
	if options.ProjectService != nil {
		policy.DefaultProjectDisabled = !serviceEnabled
	}
	if options.TsconfigRootDir != nil {
		root := *options.TsconfigRootDir
		absolute := filepath.IsAbs(root)
		if runtime.GOOS == "windows" {
			absolute = absolute && len(root) >= 2 && root[1] == ':' &&
				((root[0] >= 'a' && root[0] <= 'z') || (root[0] >= 'A' && root[0] <= 'Z'))
		}
		if !absolute {
			return ProjectPolicy{}, fmt.Errorf("parserOptions.tsconfigRootDir must be an absolute path: %q", root)
		}
		policy.TSConfigRootDirOverride = tspath.NormalizePath(filepath.Clean(root))
	} else if options.rootDirInvalid != nil {
		return ProjectPolicy{}, errors.New("parserOptions.tsconfigRootDir must be an absolute path string")
	}
	if serviceEnabled {
		policy.ServiceRootDirectory = policy.TSConfigRootDirOverride
		if policy.ServiceRootDirectory == "" {
			if defaultRootDirectory == "" {
				return ProjectPolicy{}, errors.New("project service requires its configuration's default root directory")
			}
			policy.ServiceRootDirectory = tspath.NormalizePath(defaultRootDirectory)
		}
	}
	if serviceEnabled && (options.Project != nil || options.projectAutomatic) {
		return ProjectPolicy{}, errors.New("enabling parserOptions.project does nothing when projectService is enabled; remove project or set projectService to false")
	}
	if options.projectAutomatic {
		return ProjectPolicy{}, errors.New("parserOptions.project: true is not supported; use projectService: true for automatic discovery")
	}
	policy.ProjectDisabled = options.ProjectDisabled && !serviceEnabled
	return policy, nil
}
