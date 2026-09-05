package target

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

func discoverLintTargetsMultiConfigWithPreparedFiles(
	configMap map[string]rslintconfig.RslintConfig,
	scopes map[string]OwnerScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
	ownerIndex *OwnerIndex,
	explicitSet *explicitLintTargetSet,
) []File {
	if len(configMap) == 0 {
		return nil
	}

	configDirs := make([]string, 0, len(configMap))
	for configDir := range configMap {
		configDirs = append(configDirs, configDir)
	}
	sort.Strings(configDirs)

	// A config found only because a literal file bypassed its parent's global
	// ignore is not a general ownership boundary. Build automatic ownership and
	// subtree handoff decisions from the normally reachable config set; the full
	// map is still processed below so catalog-scoped literal files use their
	// explicit-only config.
	automaticConfigMap := configMapForAutomaticTargets(configMap, scopes)
	automaticOwnerIndex := mustOwnerIndexWithPathSpaces(
		automaticConfigMap,
		fsys,
		ownerIndex.PathSpaces(),
	)

	// Explicit files are assigned to their nearest config once. Passing the
	// complete list to every config makes lint-staged-style invocations
	// O(configs*files), and also asks every config to evaluate ignores for files
	// it cannot own. A non-nil empty bucket is preserved below so the explicit
	// file-only fast path still suppresses directory walking for configs that own
	// no requested files.
	assignedExplicitOwners := make(map[tspath.Path]string)
	for _, configDir := range configDirs {
		scope, ok := scopes[configDir]
		if !ok || scope.ExplicitFiles == nil {
			continue
		}
		for _, filePath := range scope.ExplicitFiles {
			key := tspath.ToPath(tspath.NormalizePath(filePath), "", true)
			if _, exists := assignedExplicitOwners[key]; !exists {
				assignedExplicitOwners[key] = configDir
			}
		}
	}
	if explicitSet == nil {
		explicitSet = newExplicitLintTargetSet([]string{}, fsys)
	}
	for _, scope := range scopes {
		explicitSet.add(scope.ExplicitFiles, false, fsys)
	}
	filesByConfig := make(map[string][]*explicitLintTarget)
	for _, explicitFile := range explicitSet.targetsForPaths(allowFiles) {
		pathID := tspath.ToPath(explicitFile.target.Path, "", true)
		if _, assigned := assignedExplicitOwners[pathID]; assigned {
			continue
		}
		owner, _ := automaticOwnerIndex.Resolve(explicitFile.target.Identity())
		if owner != "" {
			assignedExplicitOwners[pathID] = owner
			filesByConfig[owner] = appendUniqueExplicitLintTarget(
				filesByConfig[owner],
				explicitFile,
			)
		}
	}
	filesSpecifiedByConfig := make(map[string]bool, len(configDirs))
	if allowFiles != nil {
		for _, configDir := range configDirs {
			filesSpecifiedByConfig[configDir] = true
		}
	}
	for configDir, scope := range scopes {
		if scope.ExplicitFiles == nil {
			continue
		}
		// A directory can be reached both automatically and through a literal
		// target whose parent-ignore exception selected another candidate. Config
		// discovery resolves that collision to one automatic boundary; retain its
		// automatically assigned files when adding the catalog-owned literal scope.
		for _, explicitFile := range explicitSet.targetsForPaths(scope.ExplicitFiles) {
			filesByConfig[configDir] = appendUniqueExplicitLintTarget(
				filesByConfig[configDir],
				explicitFile,
			)
		}
		filesSpecifiedByConfig[configDir] = true
	}

	seen := make(map[tspath.Path]struct{})
	var allTargets []File
	for _, configDir := range configDirs {
		var configAllowFiles []*explicitLintTarget
		if filesSpecifiedByConfig[configDir] {
			configAllowFiles = filesByConfig[configDir]
			if configAllowFiles == nil {
				configAllowFiles = []*explicitLintTarget{}
			}
		}
		configAllowDirs := allowDirs
		if scopes[configDir].ExplicitOnly {
			configAllowDirs = []string{}
		}
		targets := discoverLintTargetsForConfigInMap(
			configMap,
			automaticOwnerIndex,
			assignedExplicitOwners,
			configDir,
			fsys,
			configAllowFiles,
			configAllowDirs,
			singleThreaded,
			ownerIndex.PathSpaces(),
		)
		for _, target := range targets {
			pathID := tspath.ToPath(tspath.NormalizePath(target.Path), "", true)
			if _, exists := seen[pathID]; !exists {
				seen[pathID] = struct{}{}
				allTargets = append(allTargets, target)
			}
		}
	}
	sort.Slice(allTargets, func(i, j int) bool {
		return allTargets[i].Path < allTargets[j].Path
	})
	return allTargets
}

func appendUniqueExplicitLintTarget(
	targets []*explicitLintTarget,
	target *explicitLintTarget,
) []*explicitLintTarget {
	if target == nil {
		return targets
	}
	pathID := tspath.ToPath(target.target.Path, "", true)
	for _, existing := range targets {
		if existing != nil && tspath.ToPath(existing.target.Path, "", true) == pathID {
			return targets
		}
	}
	return append(targets, target)
}

func discoverLintTargetsForConfigInMap(
	configMap map[string]rslintconfig.RslintConfig,
	ownerIndex *OwnerIndex,
	assignedExplicitOwners map[tspath.Path]string,
	configDir string,
	fsys vfs.FS,
	explicitFiles []*explicitLintTarget,
	allowDirs []string,
	singleThreaded bool,
	pathSpaces *rslintconfig.PathSpaceSnapshot,
) []File {
	cfg, ok := configMap[configDir]
	if !ok {
		return nil
	}

	stopDirs := ownerIndex.ChildOwnerDirectories(configDir)
	targets := discoverLintTargetsWithinRoot(
		cfg,
		configDir,
		configDir,
		fsys,
		explicitFiles,
		allowDirs,
		stopDirs,
		singleThreaded,
		pathSpaces,
	)
	if len(targets) == 0 {
		return targets
	}

	ownedTargets := make([]File, 0, len(targets))
	for _, target := range targets {
		targetID := tspath.ToPath(tspath.NormalizePath(target.Path), "", true)
		if assignedOwner, assigned := assignedExplicitOwners[targetID]; assigned {
			if assignedOwner == configDir {
				target.ConfigDirectory = configDir
				ownedTargets = append(ownedTargets, target)
			}
			continue
		}
		ownerDir, _ := ownerIndex.Resolve(target.Identity())
		if ownerDir == configDir {
			target.ConfigDirectory = configDir
			ownedTargets = append(ownedTargets, target)
		}
	}
	return ownedTargets
}
