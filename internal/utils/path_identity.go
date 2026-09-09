package utils

import "github.com/microsoft/TypeScript/tsc/shim/tspath"

// NormalizeAbsoluteDrive gives absolute Windows paths a consistent drive
// spelling without changing directory names, file names, or path separators.
// Callers remain responsible for resolving or normalizing the rest of the path.
func NormalizeAbsoluteDrive(filePath string) string {
	// SplitVolumePath also accepts drive-relative paths. Keep those spellings
	// intact, and avoid rebuilding paths whose volume is already canonical.
	volume, path, ok := tspath.SplitVolumePath(filePath)
	if ok && tspath.GetRootLength(filePath) > len(volume) && filePath[:len(volume)] != volume {
		return volume + path
	}
	return filePath
}
