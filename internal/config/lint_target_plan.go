package config

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// LintTargetPlan is the stable, config-owned projection of one lint target
// discovery pass. Targets retain both the caller-visible lexical path and the
// canonical filesystem identity established by discovery, together with the
// config owner that selected them. ExplicitFileOutcomes carries the same-pass
// existence and ignore decisions needed for caller-facing skip messages.
type LintTargetPlan struct {
	Targets              []DiscoveredLintTarget
	ExplicitFileOutcomes []ExplicitFileOutcome

	frozenConfigBases map[string]configTargetBase
}

// LintTargetPlanRequest keeps the two directory roles explicit. ConfigDirectory
// is the base for authored files/ignores content and the owner recorded
// on every target. ScanRoot is the filesystem root searched for a single
// invocation-wide config; when empty it defaults to ConfigDirectory. A
// directory-keyed ConfigMap owns its own walk roots and ignores ScanRoot.
type LintTargetPlanRequest struct {
	ConfigMap       map[string]RslintConfig
	Config          RslintConfig
	ConfigDirectory string
	ScanRoot        string
	ConfigScopes    map[string]LintDiscoveryScope
	FS              vfs.FS
	Files           []string
	Directories     []string
	SingleThreaded  bool
}

// ResolveLintTargetPlan discovers the exact lint target set and rejects
// physical aliases governed by different configs. Program membership is
// deliberately not consulted at this stage.
func ResolveLintTargetPlan(request LintTargetPlanRequest) (LintTargetPlan, error) {
	configDirectory := tspath.NormalizePath(request.ConfigDirectory)
	scanRoot := tspath.NormalizePath(request.ScanRoot)
	if scanRoot == "" {
		scanRoot = configDirectory
	}
	plan := LintTargetPlan{}
	explicitFiles := newExplicitLintTargetSet(request.Files, request.FS)
	var discoveredTargets []DiscoveredLintTarget
	if request.ConfigMap != nil {
		ownerResolver := NewConfigOwnerResolver(request.ConfigMap, request.FS)
		plan.frozenConfigBases = ownerResolver.frozenBases
		discoveredTargets = discoverLintTargetsMultiConfigWithPreparedFiles(
			request.ConfigMap,
			request.ConfigScopes,
			request.FS,
			request.Files,
			request.Directories,
			request.SingleThreaded,
			ownerResolver,
			explicitFiles,
		)
	} else {
		configMap := map[string]RslintConfig{configDirectory: request.Config}
		frozenBases := freezeConfigTargetBases(configMap, request.FS)
		plan.frozenConfigBases = frozenBases
		discoveredTargets = discoverLintTargetsWithPreparedFiles(
			request.Config,
			configDirectory,
			scanRoot,
			request.FS,
			explicitFiles.targetsForPaths(request.Files),
			request.Directories,
			nil,
			request.SingleThreaded,
			frozenBases,
		)
	}
	useCaseSensitive := request.FS == nil || request.FS.UseCaseSensitiveFileNames()
	if explicitFiles != nil {
		for _, explicitFile := range explicitFiles.byPath {
			if !explicitFile.evaluated {
				explicitFile.selectedBy(
					configDirectory,
					scanRoot,
					useCaseSensitive,
					nil,
				)
			}
		}
		plan.ExplicitFileOutcomes = explicitFiles.outcomes()
	}

	plan.Targets = make([]DiscoveredLintTarget, 0, len(discoveredTargets))
	seenCanonical := make(map[string]DiscoveredLintTarget, len(discoveredTargets))
	for _, discovered := range discoveredTargets {
		target := normalizeLintTarget(discovered, configDirectory, request.FS)
		canonicalKey := exactPathID(target.CanonicalPath)
		if existing, exists := seenCanonical[canonicalKey]; exists {
			// ConfigDirectory is the authoritative catalog key selected during
			// discovery. Do not re-resolve either owner here: distinct lexical
			// aliases may carry different rules even when their directories share
			// a physical identity. Only a verified native-case spelling of the
			// same frozen owner is equivalent.
			if !sameFrozenConfigOwner(
				existing.ConfigDirectory,
				target.ConfigDirectory,
				useCaseSensitive,
				plan.frozenConfigBases,
			) {
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

// NewFileConfigResolver evaluates the supplied execution config using the
// authored path spaces frozen by target discovery. The config value remains a
// caller concern because CLI --rule and API overrides may intentionally be
// applied after the target set is selected.
func (plan LintTargetPlan) NewFileConfigResolver(
	config RslintConfig,
	configDirectory string,
	fsys vfs.FS,
	enforcePlugins bool,
) *FileConfigResolver {
	if len(plan.frozenConfigBases) == 0 {
		return NewFileConfigResolverWithFS(
			config,
			configDirectory,
			fsys,
			enforcePlugins,
		)
	}
	return &FileConfigResolver{
		config:         config,
		enforcePlugins: enforcePlugins,
		targetResolver: newConfigTargetResolverWithBases(
			config,
			configDirectory,
			fsys,
			plan.frozenConfigBases,
		),
	}
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
	if target.CanonicalParentPath == "" {
		target.CanonicalParentPath = tspath.GetDirectoryPath(target.Path)
		if fsys != nil {
			if realPath := fsys.Realpath(target.CanonicalParentPath); realPath != "" {
				target.CanonicalParentPath = realPath
			}
		}
	}
	target.CanonicalParentPath = tspath.NormalizePath(target.CanonicalParentPath)
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

func exactPathID(filePath string) string {
	return string(tspath.ToPath(tspath.NormalizePath(filePath), "", true))
}

func sameFrozenConfigOwner(
	left string,
	right string,
	useCaseSensitive bool,
	frozenBases map[string]configTargetBase,
) bool {
	left = tspath.NormalizePath(left)
	right = tspath.NormalizePath(right)
	if exactPathID(left) == exactPathID(right) {
		return true
	}
	if useCaseSensitive || !pathsEqual(left, right, false) {
		return false
	}
	leftBase, leftFrozen := frozenBases[exactPathID(left)]
	rightBase, rightFrozen := frozenBases[exactPathID(right)]
	return leftFrozen && rightFrozen &&
		exactPathID(leftBase.physicalDirectory) == exactPathID(rightBase.physicalDirectory)
}
