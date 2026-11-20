package target

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// Child config boundaries use the lexical directory entries admitted by
// discovery. The host's default casing may differ from the volume being walked.
func normalizeStopWalkDirs(configDir string, stopDirs []string) []string {
	if len(stopDirs) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(stopDirs))
	normalized := make([]string, 0, len(stopDirs))
	for _, dir := range stopDirs {
		rel, within := rslintconfig.RelativePathWithinConfigRoot(dir, configDir, true)
		if !within || rel == "" {
			continue
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		seen[rel] = struct{}{}
		normalized = append(normalized, rel)
	}
	sort.Strings(normalized)
	return normalized
}

func filterInitialWalkRoots(
	roots []string,
	isDefaultExcluded func(string) bool,
	canPrune func(string) bool,
	stopWalkDirs []string,
) []string {
	if len(roots) == 0 {
		return roots
	}

	filtered := make([]string, 0, len(roots))
	for _, root := range roots {
		root = tspath.NormalizePath(root)
		if root == "" {
			root = "."
		}
		// A root represented as "." may still be nested below a default-excluded
		// directory in the config's path space when ConfigDirectory and ScanRoot
		// differ (for example an external config invoked from node_modules/pkg).
		if isDefaultExcluded != nil && isDefaultExcluded(root) {
			continue
		}
		if isStoppedWalkPath(root, stopWalkDirs) ||
			(canPrune != nil && canPrune(root)) {
			continue
		}
		filtered = append(filtered, root)
	}
	return filtered
}

func isStoppedWalkPath(walkPath string, stopWalkDirs []string) bool {
	if len(stopWalkDirs) == 0 {
		return false
	}
	walkPath = tspath.NormalizePath(walkPath)
	if walkPath == "" || walkPath == "." {
		return false
	}
	for _, stopDir := range stopWalkDirs {
		if walkPath == stopDir || tspath.StartsWithDirectory(walkPath, stopDir, true) {
			return true
		}
	}
	return false
}

type resolvedAllowedDirectory = rslintconfig.DirectoryIdentity

// projectPathThroughRequestedDirectories returns every caller-visible spelling
// of one walked physical path. Multiple requested directory aliases may point
// at the same subtree, and each spelling remains a legitimate config target;
// callers perform selection for every projection and canonical-deduplicate
// only after selection.
func projectPathThroughRequestedDirectories(
	filePath string,
	physicalPath string,
	allowDirs []resolvedAllowedDirectory,
) []string {
	projections := make([]string, 0, len(allowDirs))
	seen := make(map[string]struct{}, len(allowDirs))
	appendProjection := func(path string) {
		path = tspath.NormalizePath(path)
		key := rslintconfig.ExactPathID(path)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		projections = append(projections, path)
	}
	for _, dir := range allowDirs {
		if relative, within := rslintconfig.RelativePathWithinConfigRoot(filePath, dir.LexicalPath, true); within {
			appendProjection(tspath.ResolvePath(dir.LexicalPath, relative))
		}
		if dir.LexicalPath == dir.CanonicalPath && filePath == physicalPath {
			continue
		}
		relative, within := rslintconfig.RelativePathWithinConfigRoot(filePath, dir.CanonicalPath, true)
		if !within && physicalPath != filePath {
			relative, within = rslintconfig.RelativePathWithinConfigRoot(physicalPath, dir.CanonicalPath, true)
		}
		if within {
			appendProjection(tspath.ResolvePath(dir.LexicalPath, relative))
		}
	}
	sort.Strings(projections)
	return projections
}

// frozenWalkRoot uses request identities first and frozen authored bases for
// owner handoffs. Target planning never resolves a requested root again.
func frozenWalkRoot(path string, directories []rslintconfig.DirectoryIdentity, spaces *rslintconfig.PathSpaceSnapshot) rslintconfig.DirectoryIdentity {
	for _, directory := range directories {
		if rslintconfig.ExactPathID(directory.LexicalPath) == rslintconfig.ExactPathID(path) {
			return directory
		}
	}
	physical, ok := spaces.PhysicalDirectory(path)
	if !ok {
		physical = path
	}
	return rslintconfig.DirectoryIdentity{LexicalPath: path, CanonicalPath: physical}
}

func discoverWalkRoots(root rslintconfig.DirectoryIdentity, allowDirs []rslintconfig.DirectoryIdentity) []string {
	if allowDirs == nil {
		return []string{"."}
	}
	seen := make(map[string]struct{}, len(allowDirs))
	var roots []string
	add := func(relative string) {
		if relative == "" {
			relative = "."
		}
		key := rslintconfig.ExactPathID(relative)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			roots = append(roots, relative)
		}
	}
	for _, directory := range allowDirs {
		if _, contains := rslintconfig.RelativePathWithinConfigRoot(root.LexicalPath, directory.LexicalPath, true); contains {
			add(".")
		} else if relative, within := rslintconfig.RelativePathWithinConfigRoot(directory.LexicalPath, root.LexicalPath, true); within {
			add(relative)
		} else if _, contains := rslintconfig.RelativePathWithinConfigRoot(root.CanonicalPath, directory.CanonicalPath, true); contains {
			add(".")
		} else if relative, within := rslintconfig.RelativePathWithinConfigRoot(directory.CanonicalPath, root.CanonicalPath, true); within {
			add(relative)
		}
	}
	// Each explicit directory remains an independent entry: its ancestors may
	// be pruned even though this directory begins a different Git repository.
	sort.Strings(roots)
	return roots
}

func pathsEqual(a, b string, useCaseSensitive bool) bool {
	return rslintconfig.PathsEqual(a, b, useCaseSensitive)
}
