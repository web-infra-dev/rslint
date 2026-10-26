package discovery

import rslintconfig "github.com/web-infra-dev/rslint/internal/config"

// collectExplicitGitSources walks only the requested directory entries.
// The config path has no role in Git boundary selection.
func (coordinator *discoveryCoordinator) collectExplicitGitSources(ownerDirectory, ownerPath string, matcher rslintconfig.GlobalIgnoreMatcher, targetsByRoot map[string]*discoveryTargetTrie) error {
	coordinator.collectExactGitFiles()
	roots := make([]discoveryWalkNode, 0, len(coordinator.targetDirectories))
	for _, directory := range coordinator.targetDirectories {
		if targetsByRoot != nil && targetsByRoot[directory.LexicalPath] == nil {
			continue
		}
		if isDefaultDiscoveryExcluded(directory.LexicalPath, coordinator.request.CWD, coordinator.fs.CaseSensitivity().IsCaseSensitive()) {
			coordinator.walkStats.directoriesPruned++
			continue
		}
		roots = append(roots, discoveryWalkNode{
			directory:          directory.LexicalPath,
			canonicalDirectory: directory.CanonicalPath,
			ownerDir:           ownerDirectory,
			ownerPath:          ownerPath,
			ignoreMatcher:      matcher,
			configReachable:    true,
			gitCursor:          coordinator.git.Ancestors(directory.CanonicalPath),
			targets:            targetsByRoot[directory.LexicalPath],
		})
	}
	return coordinator.walkDirectories(roots, false)
}
