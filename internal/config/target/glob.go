package target

import (
	"context"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// ExpandGlob enumerates lexical candidates, before config ownership and ignores
// are evaluated. The returned root retains the scan provenance for discovery.
func ExpandGlob(ctx context.Context, fsys vfs.FS, cwd, pattern string) (root string, files []string, err error) {
	base, pattern := doublestar.SplitPattern(filepath.ToSlash(pattern))
	root = tspath.ResolvePath(cwd, base)
	if !doublestar.ValidatePattern(pattern) || !fsys.DirectoryExists(root) {
		return root, nil, nil
	}
	err = fsys.WalkDir(root, func(path string, entry vfs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			return vfs.SkipDir
		}
		path = tspath.NormalizePath(path)
		if entry.IsDir() {
			if rslintconfig.IsDefaultExcludedPath(path, root, fsys.UseCaseSensitiveFileNames()) {
				return vfs.SkipDir
			}
			return nil
		}
		relative := tspath.GetRelativePathFromDirectory(root, path, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
		if doublestar.MatchUnvalidated(pattern, relative) && fsys.FileExists(path) {
			files = append(files, path)
		}
		return nil
	})
	return root, files, err
}
