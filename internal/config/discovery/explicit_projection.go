package discovery

import (
	"github.com/microsoft/typescript-go/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
)

// projectExplicitConfigGitignore freezes the invocation-scoped Git projection
// after the exact config has loaded. Config selection is already complete:
// full and directory-target projections use the catalog's parallel frontier
// with one fixed owner, while exact-file-only targets reuse the source-chain
// path used by automatic literals. Neither path discovers config candidates.
func (coordinator *discoveryCoordinator) projectExplicitConfigGitignore(state *configLoadState) error {
	if state == nil || state.failure != nil {
		return nil
	}
	files := coordinator.normalizedFiles()
	var filePaths []string
	if coordinator.request.Files != nil {
		filePaths = make([]string, 0, len(files))
		for _, file := range files {
			filePaths = append(filePaths, file.Path)
		}
	}
	var directoryPaths []string
	if coordinator.request.Directories != nil {
		directoryPaths = coordinator.normalizedDirectoryRoots()
	}
	scopes := rslintconfig.PlanGitignoreCollectionScopes(
		coordinator.request.CWD,
		coordinator.fs,
		filePaths,
		directoryPaths,
	)
	var walkRoots []discoveryWalkNode
	for _, scope := range scopes {
		coordinator.git.recordScope(state.candidate.directory, scope.Root)
		if len(scope.Directories) == 0 && scope.Files != nil {
			seeds := make([]*discoverySeed, 0, len(scope.Files))
			for _, file := range scope.Files {
				seeds = append(seeds, &discoverySeed{
					path: rslintconfig.ResolveGitignoreCollectionPath(
						file,
						"",
						scope.Root,
						coordinator.fs,
					),
					ownerDir:      state.candidate.directory,
					ownerPath:     state.candidate.path,
					gitignoreRoot: scope.Root,
				})
			}
			coordinator.git.collectExactTargets(seeds)
			continue
		}
		var targets *discoveryTargetTrie
		if scope.Files != nil || scope.Directories != nil {
			var hasTargets bool
			targets, hasTargets = coordinator.explicitProjectionTargets(scope)
			if !hasTargets {
				continue
			}
		}
		canonicalRoot := coordinator.fs.Realpath(scope.Root)
		if discoveryPathsEqual(scope.Root, canonicalRoot, coordinator.fs.UseCaseSensitiveFileNames()) {
			canonicalRoot = ""
		}
		walkRoots = append(walkRoots, discoveryWalkNode{
			directory:          scope.Root,
			canonicalDirectory: canonicalRoot,
			ownerDir:           state.candidate.directory,
			ownerPath:          state.candidate.path,
			gitDirectory:       scope.Root,
			gitCursor:          gitignore.NewCursor(scope.Root, coordinator.fs.UseCaseSensitiveFileNames()),
			gitActive:          true,
			targets:            targets,
		})
	}
	if len(walkRoots) == 0 {
		return coordinator.ctx.Err()
	}
	return coordinator.walkDirectories(walkRoots)
}

func (coordinator *discoveryCoordinator) explicitProjectionTargets(scope rslintconfig.GitignoreCollectionScope) (*discoveryTargetTrie, bool) {
	root := &discoveryTargetTrie{}
	hasTargets := false
	for _, file := range scope.Files {
		collectionPath := rslintconfig.ResolveGitignoreCollectionPath(
			file,
			"",
			scope.Root,
			coordinator.fs,
		)
		hasTargets = addDiscoveryTarget(
			root,
			scope.Root,
			tspath.GetDirectoryPath(collectionPath),
			false,
			coordinator.fs.UseCaseSensitiveFileNames(),
		) || hasTargets
	}
	for _, directory := range scope.Directories {
		collectionPath := rslintconfig.ResolveGitignoreCollectionDirectory(
			directory,
			scope.Root,
			coordinator.fs,
		)
		hasTargets = addDiscoveryTarget(
			root,
			scope.Root,
			collectionPath,
			true,
			coordinator.fs.UseCaseSensitiveFileNames(),
		) || hasTargets
	}
	return root, hasTargets
}
