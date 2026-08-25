package target

import (
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// Plan is the stable projection of one lint-target discovery pass. Files
// retain caller-visible and canonical identities plus the config owner that
// selected them. ExplicitFileOutcomes carries the same-pass facts needed for
// caller-facing skip messages.
type Plan struct {
	Files                []File
	ExplicitFileOutcomes []ExplicitFileOutcome

	pathSpaces *rslintconfig.PathSpaceSnapshot
}

// Request keeps the two directory roles explicit. ConfigDirectory
// is the base for authored files/ignores content and the owner recorded
// on every target. ScanRoot is the filesystem root searched for a single
// invocation-wide config; when empty it defaults to ConfigDirectory. A
// directory-keyed ConfigMap owns its own walk roots and ignores ScanRoot.
type Request struct {
	ConfigMap       map[string]rslintconfig.RslintConfig
	Config          rslintconfig.RslintConfig
	ConfigDirectory string
	ScanRoot        string
	OwnerScopes     map[string]OwnerScope
	FS              vfs.FS
	Files           []string
	Directories     []string
	SingleThreaded  bool
}

// Resolve discovers the exact lint target set and rejects
// physical aliases governed by different configs. Program membership is
// deliberately not consulted at this stage.
func Resolve(request Request) (Plan, error) {
	configDirectory := tspath.NormalizePath(request.ConfigDirectory)
	scanRoot := tspath.NormalizePath(request.ScanRoot)
	if scanRoot == "" {
		scanRoot = configDirectory
	}
	plan := Plan{}
	explicitFiles := newExplicitLintTargetSet(request.Files, request.FS)
	var discoveredFiles []File
	if request.ConfigMap != nil {
		ownerIndex := NewOwnerIndex(request.ConfigMap, request.FS)
		plan.pathSpaces = ownerIndex.PathSpaces()
		discoveredFiles = discoverLintTargetsMultiConfigWithPreparedFiles(
			request.ConfigMap,
			request.OwnerScopes,
			request.FS,
			request.Files,
			request.Directories,
			request.SingleThreaded,
			ownerIndex,
			explicitFiles,
		)
	} else {
		configMap := map[string]rslintconfig.RslintConfig{configDirectory: request.Config}
		plan.pathSpaces = rslintconfig.NewPathSpaceSnapshot(configMap, request.FS)
		discoveredFiles = discoverLintTargetsWithPreparedFiles(
			request.Config,
			configDirectory,
			scanRoot,
			request.FS,
			explicitFiles.targetsForPaths(request.Files),
			request.Directories,
			nil,
			request.SingleThreaded,
			plan.pathSpaces,
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

	plan.Files = make([]File, 0, len(discoveredFiles))
	seenCanonical := make(map[string]File, len(discoveredFiles))
	for _, discovered := range discoveredFiles {
		target := normalizeLintTarget(discovered, configDirectory, request.FS)
		canonicalKey := rslintconfig.ExactPathID(target.CanonicalPath)
		if existing, exists := seenCanonical[canonicalKey]; exists {
			// ConfigDirectory is the authoritative catalog key selected during
			// discovery. Do not re-resolve either owner here: distinct lexical
			// aliases may carry different rules even when their directories share
			// a physical identity. Only a verified native-case spelling of the
			// same frozen owner is equivalent.
			if !plan.pathSpaces.SameDirectory(
				existing.ConfigDirectory,
				target.ConfigDirectory,
				useCaseSensitive,
			) {
				return Plan{}, fmt.Errorf(
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
		plan.Files = append(plan.Files, target)
	}
	return plan, nil
}

func normalizeLintTarget(target File, defaultConfigDirectory string, fsys vfs.FS) File {
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

// ActiveOwners returns stable, deduplicated owner keys for files in the plan.
func (plan Plan) ActiveOwners() []string {
	if len(plan.Files) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(plan.Files))
	owners := make([]string, 0, len(plan.Files))
	for _, file := range plan.Files {
		if _, exists := seen[file.ConfigDirectory]; exists {
			continue
		}
		seen[file.ConfigDirectory] = struct{}{}
		owners = append(owners, file.ConfigDirectory)
	}
	sort.Strings(owners)
	return owners
}

// PathSpaces returns the read-only authored path-space generation used by the
// plan. Callers use it to construct config resolvers without observing the
// filesystem again.
func (plan Plan) PathSpaces() *rslintconfig.PathSpaceSnapshot { return plan.pathSpaces }

// PreferredCallerPaths maps canonical target identity to the first lexical
// path supplied by discovery. Diagnostic deduplication uses this to retain the
// caller's spelling when a Program exposes an alias.
func (plan Plan) PreferredCallerPaths() map[string]string {
	if len(plan.Files) == 0 {
		return nil
	}
	preferred := make(map[string]string, len(plan.Files))
	for _, target := range plan.Files {
		canonicalID := rslintconfig.ExactPathID(target.CanonicalPath)
		if _, exists := preferred[canonicalID]; !exists {
			preferred[canonicalID] = target.Path
		}
	}
	return preferred
}
