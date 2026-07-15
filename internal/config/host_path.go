package config

import (
	"github.com/web-infra-dev/rslint/internal/hostpath"
)

// Native POSIX paths must preserve backslash as a filename byte. TypeScript's
// path helpers intentionally treat it as a separator on every platform, so
// paths passed to the host filesystem always use filepath on POSIX.
//
// A few config helpers are also exercised with synthetic Windows roots in
// platform-independent tests. Those helpers choose their path implementation
// from the governing absolute root, never from an authored relative value. In
// particular, a POSIX config with basePath "C:/scope" must resolve that value
// below the config directory just like node:path.posix.resolve does.
func pathUsesNativePOSIXSemantics(governingRoot string) bool {
	return hostpath.UsesNativePOSIX(governingRoot)
}

func NormalizeHostPath(value string) string {
	return hostpath.Normalize(value)
}

func HostDirectoryPath(value string) string {
	return hostpath.Directory(value)
}

func HostBaseFileName(value string) string {
	return hostpath.Base(value)
}

func HostPathIsAbsolute(value string) bool {
	return hostpath.IsAbsolute(value)
}

func ResolveHostPath(basePath string, value string) string {
	return hostpath.Resolve(basePath, value)
}

func normalizePathForRoot(governingRoot string, value string) string {
	return hostpath.NormalizeForRoot(governingRoot, value)
}

func directoryPathForRoot(governingRoot string, value string) string {
	return hostpath.DirectoryForRoot(governingRoot, value)
}

func resolvePathForRoot(governingRoot string, basePath string, value string) string {
	return hostpath.ResolveForRoot(governingRoot, basePath, value)
}

func pathIsAbsoluteForRoot(governingRoot string, value string) bool {
	return hostpath.IsAbsoluteForRoot(governingRoot, value)
}

func pathsEqualForRoot(governingRoot string, left string, right string, caseSensitive bool) bool {
	return hostpath.EqualForRoot(governingRoot, left, right, caseSensitive)
}
