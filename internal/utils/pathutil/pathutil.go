// Package pathutil adapts rslint's string paths to the compiler's rooted paths.
package pathutil

import "github.com/microsoft/TypeScript/tsc/shim/tspath"

// ComparePathsOptions describes string paths that may be relative to a caller's
// working directory. Compiler paths and canonical keys use tspath directly.
type ComparePathsOptions struct {
	CurrentDirectory          string
	UseCaseSensitiveFileNames bool
}

func CaseSensitivity(caseSensitive bool) tspath.CaseSensitivity {
	if caseSensitive {
		return tspath.CaseSensitive
	}
	return tspath.CaseInsensitive
}

func ToPath(fileName, basePath string, caseSensitive bool) tspath.PathKey {
	return CaseSensitivity(caseSensitive).PathKey(tspath.ToRootedPath(fileName, Directory(basePath)))
}

func Directory(path string) tspath.RootedDirectoryPath {
	if path == "" {
		return ""
	}
	return tspath.RootedDirectoryPathFromAbsolute(path)
}

func File(path string) tspath.RootedFilePath {
	if path == "" {
		return ""
	}
	return tspath.RootedFilePathFromAbsolute(path)
}

func Path(path string) tspath.RootedPath {
	if path == "" {
		return ""
	}
	return tspath.RootedPathFromAbsolute(path)
}

func ComparePaths(a, b string, options ComparePathsOptions) int {
	return tspath.ComparePathsRelativeTo(a, b, Directory(options.CurrentDirectory), CaseSensitivity(options.UseCaseSensitiveFileNames))
}

func ContainsPath(parent, child string, options ComparePathsOptions) bool {
	if options.CurrentDirectory != "" {
		parent = tspath.ResolvePath(options.CurrentDirectory, parent)
		child = tspath.ResolvePath(options.CurrentDirectory, child)
	}
	return tspath.ContainsPath(parent, child, CaseSensitivity(options.UseCaseSensitiveFileNames))
}

func ConvertToRelativePath(path string, options ComparePathsOptions) string {
	return tspath.ConvertToRelativePath(path, Directory(options.CurrentDirectory), CaseSensitivity(options.UseCaseSensitiveFileNames))
}

func GetRelativePathFromDirectory(from, to string, options ComparePathsOptions) string {
	return tspath.ResolveRelativePathFromDirectory(from, to, Directory(options.CurrentDirectory), CaseSensitivity(options.UseCaseSensitiveFileNames))
}

func GetNormalizedAbsolutePath(path, cwd string) string {
	return tspath.GetNormalizedAbsolutePath(path, Directory(cwd))
}

func StartsWithDirectory(path, directory string, caseSensitive bool) bool {
	return tspath.StartsWithDirectory(path, directory, CaseSensitivity(caseSensitive))
}

func GetAnyExtensionFromPath(path string, extensions []string, ignoreCase bool) string {
	return tspath.GetAnyExtensionFromPath(path, extensions, CaseSensitivity(!ignoreCase))
}

func GetCanonicalFileName(path string, caseSensitive bool) string {
	return CaseSensitivity(caseSensitive).Canonicalize(path)
}

func Files(paths []string) []tspath.RootedFilePath {
	return ResolveFiles(paths, "")
}

func ResolveFiles(paths []string, directory tspath.RootedDirectoryPath) []tspath.RootedFilePath {
	if paths == nil {
		return nil
	}
	result := make([]tspath.RootedFilePath, len(paths))
	for i, path := range paths {
		result[i] = tspath.ToRootedFilePath(path, directory)
	}
	return result
}

func Strings[T ~string](paths []T) []string {
	if paths == nil {
		return nil
	}
	result := make([]string, len(paths))
	for i, path := range paths {
		result[i] = string(path)
	}
	return result
}

func IsDiskPathRoot(path string) bool {
	rootLength := tspath.GetEncodedRootLength(path)
	return rootLength > 0 && rootLength == len(path)
}

func RemoveTrailingDirectorySeparators(path string) string {
	for tspath.HasTrailingDirectorySeparator(path) {
		path = tspath.RemoveTrailingDirectorySeparator(path)
	}
	return path
}

func Directories(paths []string) []tspath.RootedDirectoryPath {
	if paths == nil {
		return nil
	}
	result := make([]tspath.RootedDirectoryPath, len(paths))
	for i, path := range paths {
		result[i] = Directory(path)
	}
	return result
}
