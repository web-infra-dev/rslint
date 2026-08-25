package target

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

func normalizeStopWalkDirs(configDir string, stopDirs []string, useCaseSensitive bool) []string {
	if len(stopDirs) == 0 {
		return nil
	}

	compareOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          configDir,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	seen := make(map[string]struct{}, len(stopDirs))
	normalized := make([]string, 0, len(stopDirs))
	for _, rawDir := range stopDirs {
		dir := tspath.NormalizePath(rawDir)
		if pathsEqual(dir, configDir, useCaseSensitive) ||
			!tspath.StartsWithDirectory(dir, configDir, useCaseSensitive) {
			continue
		}
		rel := tspath.NormalizePath(tspath.GetRelativePathFromDirectory(configDir, dir, compareOptions))
		if rel == "" || rel == "." {
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
	useCaseSensitive bool,
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
		if root != "." {
			if isStoppedWalkPath(root, stopWalkDirs, useCaseSensitive) ||
				(canPrune != nil && canPrune(root)) {
				continue
			}
		}
		filtered = append(filtered, root)
	}
	return filtered
}

func isStoppedWalkPath(walkPath string, stopWalkDirs []string, useCaseSensitive bool) bool {
	if len(stopWalkDirs) == 0 {
		return false
	}
	walkPath = tspath.NormalizePath(walkPath)
	if walkPath == "" || walkPath == "." {
		return false
	}
	for _, stopDir := range stopWalkDirs {
		if pathsEqual(walkPath, stopDir, useCaseSensitive) ||
			tspath.StartsWithDirectory(walkPath, stopDir, useCaseSensitive) {
			return true
		}
	}
	return false
}

type resolvedAllowedDirectory = rslintconfig.DirectoryIdentity

type requestedPathProjection struct {
	Path          string
	CanonicalPath string
}

func resolveAllowedDirectories(allowDirs []string, fsys vfs.FS) []resolvedAllowedDirectory {
	return rslintconfig.ResolveDirectoryIdentities(allowDirs, fsys)
}

// projectPathThroughRequestedDirectories returns every caller-visible spelling
// of one walked physical path. Multiple requested directory aliases may point
// at the same subtree, and each spelling remains a legitimate config target;
// callers perform selection for every projection and canonical-deduplicate
// only after selection.
func projectPathThroughRequestedDirectories(
	filePath string,
	matchPath string,
	allowDirs []resolvedAllowedDirectory,
	useCaseSensitive bool,
) []requestedPathProjection {
	projections := make([]requestedPathProjection, 0, len(allowDirs))
	seen := make(map[string]struct{}, len(allowDirs))
	appendProjection := func(path string, canonicalPath string) {
		path = tspath.NormalizePath(path)
		if canonicalPath != "" {
			canonicalPath = tspath.NormalizePath(canonicalPath)
		}
		key := rslintconfig.ExactPathID(path)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		projections = append(projections, requestedPathProjection{
			Path:          path,
			CanonicalPath: canonicalPath,
		})
	}
	for _, dir := range allowDirs {
		if _, within := rslintconfig.RelativePathWithinConfigRoot(filePath, dir.LexicalPath, useCaseSensitive); within {
			relative, exactWithin := rslintconfig.RelativePathWithinConfigRoot(
				filePath,
				dir.LexicalPath,
				true,
			)
			if !exactWithin {
				// A case-insensitive lexical resemblance is authoritative only
				// when the frozen physical spelling verifies the same root. This
				// keeps native casing aliases while rejecting a distinct tree that
				// differs only by case on a case-sensitive filesystem facade.
				physicalRelative, verified := rslintconfig.RelativePathWithinConfigRoot(
					filePath,
					dir.CanonicalPath,
					true,
				)
				if !verified && matchPath != filePath {
					physicalRelative, verified = rslintconfig.RelativePathWithinConfigRoot(
						matchPath,
						dir.CanonicalPath,
						true,
					)
				}
				if !verified {
					continue
				}
				relative = physicalRelative
			}
			canonicalPath := ""
			if dir.LexicalPath != dir.CanonicalPath {
				canonicalPath = tspath.ResolvePath(dir.CanonicalPath, relative)
			}
			appendProjection(tspath.ResolvePath(dir.LexicalPath, relative), canonicalPath)
		}
		if dir.LexicalPath == dir.CanonicalPath {
			continue
		}
		relative, within := rslintconfig.RelativePathWithinConfigRoot(filePath, dir.CanonicalPath, true)
		if !within && matchPath != filePath {
			relative, within = rslintconfig.RelativePathWithinConfigRoot(matchPath, dir.CanonicalPath, true)
		}
		if within {
			appendProjection(
				tspath.ResolvePath(dir.LexicalPath, relative),
				tspath.ResolvePath(dir.CanonicalPath, relative),
			)
		}
	}
	sort.Slice(projections, func(i, j int) bool {
		return projections[i].Path < projections[j].Path
	})
	return projections
}

func discoverWalkRoots(configDir string, allowDirs []string, fsys vfs.FS) []string {
	if allowDirs == nil {
		return []string{"."}
	}
	if len(allowDirs) == 0 {
		return nil
	}

	configDir = tspath.NormalizePath(configDir)
	canonicalConfigDir := configDir
	if fsys != nil {
		if realPath := fsys.Realpath(configDir); realPath != "" {
			canonicalConfigDir = tspath.NormalizePath(realPath)
		}
	}
	seen := make(map[string]struct{}, len(allowDirs))
	roots := make([]string, 0, len(allowDirs))
	addRoot := func(root string) {
		if root == "" {
			root = "."
		}
		root = tspath.NormalizePath(root)
		if root == "." {
			roots = []string{"."}
			seen = map[string]struct{}{".": {}}
			return
		}
		for _, existing := range roots {
			if existing == "." {
				return
			}
			if pathsEqual(existing, root, true) ||
				tspath.StartsWithDirectory(root, existing, true) {
				return
			}
		}
		filtered := roots[:0]
		seen = make(map[string]struct{}, len(allowDirs))
		for _, existing := range roots {
			if pathsEqual(existing, root, true) ||
				tspath.StartsWithDirectory(existing, root, true) {
				continue
			}
			seen[existing] = struct{}{}
			filtered = append(filtered, existing)
		}
		roots = filtered
		if _, ok := seen[root]; ok {
			return
		}
		seen[root] = struct{}{}
		roots = append(roots, root)
	}

	for _, rawDir := range allowDirs {
		dir := tspath.NormalizePath(rawDir)
		if pathsEqual(dir, configDir, true) ||
			tspath.StartsWithDirectory(configDir, dir, true) {
			return []string{"."}
		}
		if relative, within := rslintconfig.RelativePathWithinConfigRoot(dir, configDir, true); within {
			addRoot(relative)
			continue
		}
		if fsys == nil {
			continue
		}
		realDir := fsys.Realpath(dir)
		if realDir == "" {
			continue
		}
		canonicalDir := tspath.NormalizePath(realDir)
		if pathsEqual(canonicalDir, canonicalConfigDir, true) ||
			tspath.StartsWithDirectory(canonicalConfigDir, canonicalDir, true) {
			return []string{"."}
		}
		if relative, within := rslintconfig.RelativePathWithinConfigRoot(canonicalDir, canonicalConfigDir, true); within {
			addRoot(relative)
		}
	}

	sort.Strings(roots)
	return roots
}

func pathsEqual(a, b string, useCaseSensitive bool) bool {
	return rslintconfig.PathsEqual(a, b, useCaseSensitive)
}
