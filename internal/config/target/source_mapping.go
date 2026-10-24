package target

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// LookupSourceTarget maps a Program source path back to the selected lint
// target represented by that source. Exact lexical identity wins; canonical
// filesystem identity is only a fallback for Program path aliases.
func LookupSourceTarget(
	targetsBySourcePath map[string]File,
	sourcePath string,
	fsys vfs.FS,
) (File, bool) {
	if len(targetsBySourcePath) == 0 {
		return File{}, false
	}
	normalizedSourcePath := tspath.NormalizePath(sourcePath)
	if lintTarget, ok := targetsBySourcePath[rslintconfig.ExactPathID(normalizedSourcePath)]; ok {
		return lintTarget, true
	}
	canonicalID := rslintconfig.ExactPathID(canonicalPathOrSelf(normalizedSourcePath, fsys))
	lintTarget, ok := targetsBySourcePath[canonicalID]
	return lintTarget, ok
}

func canonicalPathOrSelf(path string, fsys vfs.FS) string {
	if fsys != nil {
		if physical := fsys.Realpath(pathutil.Path(path)); physical != "" {
			return tspath.NormalizePath(physical.AsString())
		}
	}
	return tspath.NormalizePath(path)
}
