package hostpath

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/microsoft/typescript-go/shim/tspath"
)

// Native helpers are for actual host filesystem operands. On POSIX a
// backslash is data, not a separator.
func Normalize(value string) string {
	if value == "" {
		return ""
	}
	if runtime.GOOS != "windows" {
		return filepath.Clean(value)
	}
	return tspath.NormalizePath(value)
}

func Directory(value string) string {
	if runtime.GOOS != "windows" {
		directory := filepath.Dir(value)
		if directory == "." && !strings.Contains(value, "/") {
			return ""
		}
		return directory
	}
	return tspath.GetDirectoryPath(value)
}

func Base(value string) string {
	if runtime.GOOS != "windows" {
		return filepath.Base(value)
	}
	return tspath.GetBaseFileName(value)
}

func IsAbsolute(value string) bool {
	if runtime.GOOS != "windows" {
		return filepath.IsAbs(value)
	}
	return tspath.GetRootLength(value) > 0
}

func Resolve(basePath string, value string) string {
	if runtime.GOOS != "windows" {
		if filepath.IsAbs(value) {
			return filepath.Clean(value)
		}
		return filepath.Join(basePath, value)
	}
	return tspath.ResolvePath(basePath, value)
}

// Root-aware helpers also support synthetic Windows roots in cross-platform
// config tests. The governing root, never an authored relative value, selects
// the path implementation.
func UsesNativePOSIX(governingRoot string) bool {
	return runtime.GOOS != "windows" && !isSyntheticWindowsRoot(governingRoot)
}

func isSyntheticWindowsRoot(value string) bool {
	if runtime.GOOS == "windows" {
		return true
	}
	if len(value) >= 3 && value[1] == ':' && (value[2] == '/' || value[2] == '\\') {
		return true
	}
	return strings.HasPrefix(value, "//")
}

func NormalizeForRoot(governingRoot string, value string) string {
	if value == "" {
		return ""
	}
	if UsesNativePOSIX(governingRoot) {
		return filepath.Clean(value)
	}
	return tspath.NormalizePath(value)
}

func DirectoryForRoot(governingRoot string, value string) string {
	if UsesNativePOSIX(governingRoot) {
		directory := filepath.Dir(value)
		if directory == "." && !strings.Contains(value, "/") {
			return ""
		}
		return directory
	}
	return tspath.GetDirectoryPath(value)
}

func BaseForRoot(governingRoot string, value string) string {
	if UsesNativePOSIX(governingRoot) {
		return filepath.Base(value)
	}
	return tspath.GetBaseFileName(value)
}

func ResolveForRoot(governingRoot string, basePath string, value string) string {
	if UsesNativePOSIX(governingRoot) {
		if filepath.IsAbs(value) {
			return filepath.Clean(value)
		}
		return filepath.Join(basePath, value)
	}
	return tspath.ResolvePath(basePath, value)
}

func IsAbsoluteForRoot(governingRoot string, value string) bool {
	if UsesNativePOSIX(governingRoot) {
		return filepath.IsAbs(value)
	}
	return tspath.GetRootLength(value) > 0
}

func EqualForRoot(governingRoot string, left string, right string, caseSensitive bool) bool {
	left = NormalizeForRoot(governingRoot, left)
	right = NormalizeForRoot(governingRoot, right)
	if caseSensitive {
		return left == right
	}
	return strings.EqualFold(left, right)
}

// Identity returns a stable host-path map key without changing path spelling
// rules. POSIX backslashes remain filename data; Windows separators are
// normalized through tspath.
func Identity(value string, currentDirectory string, caseSensitive bool) string {
	if value == "" {
		return ""
	}
	if !IsAbsoluteForRoot(currentDirectory, value) {
		value = ResolveForRoot(currentDirectory, currentDirectory, value)
	}
	value = NormalizeForRoot(currentDirectory, value)
	if !caseSensitive {
		value = strings.ToLower(value)
	}
	return value
}

// ConvertToRelativePath mirrors tspath.ConvertToRelativePath for host paths.
// It is intentionally lexical: symlink aliases and POSIX backslashes survive
// user-facing diagnostics, API results, and warnings.
func ConvertToRelativePath(value string, currentDirectory string, caseSensitive bool) string {
	if !IsAbsoluteForRoot(currentDirectory, value) {
		return value
	}
	if !UsesNativePOSIX(currentDirectory) {
		return tspath.ConvertToRelativePath(value, tspath.ComparePathsOptions{
			CurrentDirectory:          currentDirectory,
			UseCaseSensitiveFileNames: caseSensitive,
		})
	}
	relative, err := filepath.Rel(filepath.Clean(currentDirectory), filepath.Clean(value))
	if err != nil {
		return filepath.Clean(value)
	}
	return filepath.ToSlash(relative)
}

// RelativeWithin returns the target spelling relative to root only when it is
// contained by root. Case-insensitive POSIX filesystems compare components
// without losing the caller's lexical spelling.
func RelativeWithin(target string, root string, caseSensitive bool) (string, bool) {
	if UsesNativePOSIX(root) {
		target = filepath.Clean(target)
		root = filepath.Clean(root)
		if caseSensitive {
			relative, err := filepath.Rel(root, target)
			if err != nil {
				return "", false
			}
			if relative == "." {
				return "", true
			}
			if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				return "", false
			}
			return filepath.ToSlash(relative), true
		}
		rootParts, rootAbsolute := splitPOSIXPath(root)
		targetParts, targetAbsolute := splitPOSIXPath(target)
		if rootAbsolute != targetAbsolute || len(rootParts) > len(targetParts) {
			return "", false
		}
		for index := range rootParts {
			if !strings.EqualFold(rootParts[index], targetParts[index]) {
				return "", false
			}
		}
		return strings.Join(targetParts[len(rootParts):], "/"), true
	}

	options := tspath.ComparePathsOptions{
		CurrentDirectory:          root,
		UseCaseSensitiveFileNames: caseSensitive,
	}
	if tspath.ComparePaths(target, root, options) == 0 {
		return "", true
	}
	if !tspath.StartsWithDirectory(target, root, caseSensitive) {
		return "", false
	}
	return tspath.GetRelativePathFromDirectory(root, target, options), true
}

func splitPOSIXPath(value string) ([]string, bool) {
	absolute := strings.HasPrefix(value, "/")
	trimmed := strings.TrimPrefix(value, "/")
	if trimmed == "" || trimmed == "." {
		return nil, absolute
	}
	return strings.Split(trimmed, "/"), absolute
}
