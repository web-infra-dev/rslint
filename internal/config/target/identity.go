package target

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// FreezeFileIdentity observes the caller-visible file together with its
// physical file and parent paths. The returned value contains no owner or
// config state. Reading the parent before and after the file retries one
// concurrent directory-alias move instead of combining two known generations.
func FreezeFileIdentity(filePath string, fsys vfs.FS) rslintconfig.PathIdentity {
	filePath = tspath.NormalizePath(filePath)
	parentPath := tspath.GetDirectoryPath(filePath)
	canonicalPath := filePath
	canonicalParentPath := parentPath
	if fsys == nil {
		return rslintconfig.PathIdentity{
			Path:                filePath,
			CanonicalPath:       canonicalPath,
			CanonicalParentPath: canonicalParentPath,
		}
	}
	for range 2 {
		parentBefore := canonicalPathOrSelf(parentPath, fsys)
		canonicalPath = canonicalPathOrSelf(filePath, fsys)
		parentAfter := canonicalPathOrSelf(parentPath, fsys)
		canonicalParentPath = parentAfter
		if parentBefore == parentAfter {
			break
		}
	}
	return rslintconfig.PathIdentity{
		Path:                filePath,
		CanonicalPath:       canonicalPath,
		CanonicalParentPath: canonicalParentPath,
	}
}

func canonicalPathOrSelf(filePath string, fsys vfs.FS) string {
	if fsys != nil {
		if realPath := fsys.Realpath(filePath); realPath != "" {
			return tspath.NormalizePath(realPath)
		}
	}
	return tspath.NormalizePath(filePath)
}
