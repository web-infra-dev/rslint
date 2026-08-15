package config

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// LintTargetPlan is the stable, config-owned projection of one lint target
// discovery pass. Targets retain both the caller-visible lexical path and the
// canonical filesystem identity established by discovery, together with the
// config owner that selected them.
type LintTargetPlan struct {
	Targets []DiscoveredLintTarget
}

// ResolveLintTargetPlan discovers the exact lint target set and rejects
// physical aliases governed by different configs. Program membership is
// deliberately not consulted at this stage.
func ResolveLintTargetPlan(
	configMap map[string]RslintConfig,
	config RslintConfig,
	currentDirectory string,
	configTargetScopes map[string]LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) (LintTargetPlan, error) {
	var discoveredTargets []DiscoveredLintTarget
	if configMap != nil {
		discoveredTargets = DiscoverLintTargetsMultiConfig(
			configMap,
			configTargetScopes,
			fsys,
			allowFiles,
			allowDirs,
			singleThreaded,
		)
	} else {
		discoveredTargets = DiscoverLintTargets(
			config,
			currentDirectory,
			fsys,
			allowFiles,
			allowDirs,
			singleThreaded,
		)
	}

	plan := LintTargetPlan{Targets: make([]DiscoveredLintTarget, 0, len(discoveredTargets))}
	seenCanonical := make(map[string]DiscoveredLintTarget, len(discoveredTargets))
	for _, discovered := range discoveredTargets {
		target := normalizeLintTarget(discovered, currentDirectory, fsys)
		canonicalKey := exactPathID(target.CanonicalPath)
		if existing, exists := seenCanonical[canonicalKey]; exists {
			if canonicalPathID(existing.ConfigDirectory, fsys) != canonicalPathID(target.ConfigDirectory, fsys) {
				return LintTargetPlan{}, fmt.Errorf(
					"lint target aliases %q and %q resolve to the same file but are governed by different configs %q and %q",
					existing.Path,
					target.Path,
					existing.ConfigDirectory,
					target.ConfigDirectory,
				)
			}
			continue
		}
		seenCanonical[canonicalKey] = target
		plan.Targets = append(plan.Targets, target)
	}
	return plan, nil
}

func normalizeLintTarget(target DiscoveredLintTarget, defaultConfigDirectory string, fsys vfs.FS) DiscoveredLintTarget {
	target.Path = tspath.NormalizePath(target.Path)
	if target.CanonicalPath == "" {
		target.CanonicalPath = target.Path
		if fsys != nil {
			if realPath := fsys.Realpath(target.CanonicalPath); realPath != "" {
				target.CanonicalPath = realPath
			}
		}
	}
	target.CanonicalPath = tspath.NormalizePath(target.CanonicalPath)
	if target.ConfigDirectory == "" {
		target.ConfigDirectory = defaultConfigDirectory
	}
	target.ConfigDirectory = tspath.NormalizePath(target.ConfigDirectory)
	return target
}

// ActiveConfigs returns only configs that own at least one target in the plan.
func (plan LintTargetPlan) ActiveConfigs(configMap map[string]RslintConfig) map[string]RslintConfig {
	if len(configMap) == 0 || len(plan.Targets) == 0 {
		return nil
	}
	active := make(map[string]RslintConfig)
	for _, target := range plan.Targets {
		if entries, ok := configMap[target.ConfigDirectory]; ok {
			active[target.ConfigDirectory] = entries
		}
	}
	return active
}

// PreferredCallerPaths maps canonical target identity to the first lexical
// path supplied by discovery. Diagnostic deduplication uses this to retain the
// caller's spelling when a Program exposes an alias.
func (plan LintTargetPlan) PreferredCallerPaths() map[string]string {
	if len(plan.Targets) == 0 {
		return nil
	}
	preferred := make(map[string]string, len(plan.Targets))
	for _, target := range plan.Targets {
		canonicalID := exactPathID(target.CanonicalPath)
		if _, exists := preferred[canonicalID]; !exists {
			preferred[canonicalID] = target.Path
		}
	}
	return preferred
}

// MatchPath returns the target path used for files/ignores matching. Program
// source aliases are deliberately excluded: Program membership must not change
// the lint configuration selected for a target.
func (target DiscoveredLintTarget) MatchPath(fsys vfs.FS) string {
	matchPath, _ := ResolveConfigPathSpaceWithCanonical(
		target.Path,
		target.CanonicalPath,
		target.ConfigDirectory,
		fsys,
	)
	return matchPath
}

func exactPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

func canonicalPathID(filePath string, fsys vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return exactPathID(filePath)
}
