package discovery

// collectExplicitGitSources walks only the requested directory entries.
// The config path has no role in Git boundary selection.
func (coordinator *discoveryCoordinator) collectExplicitGitSources(state *configLoadState) error {
	coordinator.collectExactGitFiles()
	ownerDirectory, ownerPath := "", ""
	if state != nil && state.failure == nil {
		ownerDirectory = state.candidate.directory
		ownerPath = state.candidate.path
	}
	roots := make([]discoveryWalkNode, 0, len(coordinator.targetDirectories))
	for _, directory := range coordinator.targetDirectories {
		if isDefaultDiscoveryExcluded(directory.LexicalPath, coordinator.request.CWD, coordinator.fs.UseCaseSensitiveFileNames()) {
			coordinator.walkStats.directoriesPruned++
			continue
		}
		roots = append(roots, discoveryWalkNode{
			directory:          directory.LexicalPath,
			canonicalDirectory: directory.CanonicalPath,
			ownerDir:           ownerDirectory,
			ownerPath:          ownerPath,
			configReachable:    true,
			gitCursor:          coordinator.git.Ancestors(directory.CanonicalPath),
		})
	}
	return coordinator.walkDirectories(roots)
}
