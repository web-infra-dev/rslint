package lsp

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (s *Server) currentEditorOverlayFSForTarget(
	uri lsproto.DocumentUri,
	target target.File,
) vfs.FS {
	content, open := s.documents[uri]
	files, _ := s.currentEditorOverlayFilesForFrozenTarget(uri, target, content, open)
	return newFrozenLintTargetOverlayFS(s.fs, files, target)
}

// currentEditorOverlayFilesForFrozenTarget resolves every other open document
// normally, but inserts the selected document only under the identity frozen
// by documentLintSnapshot. Re-resolving the selected URI here could mix two
// symlink generations between config/project selection and Program creation.
func (s *Server) currentEditorOverlayFilesForFrozenTarget(
	uri lsproto.DocumentUri,
	target target.File,
	targetContent string,
	includeTarget bool,
) (map[string]string, bool) {
	files := make(map[string]string, len(s.documents)*2)
	contentByPhysicalPath := make(map[string]string, len(s.documents))
	aliasesConflict := false
	add := func(filePath string, content string) {
		physicalPath := s.addEditorOverlayFile(files, filePath, content)
		if previous, exists := contentByPhysicalPath[physicalPath]; exists && previous != content {
			aliasesConflict = true
		}
		contentByPhysicalPath[physicalPath] = content
	}
	for documentURI, documentContent := range s.documents {
		if documentURI == uri {
			continue
		}
		add(uriToPath(documentURI), documentContent)
	}
	if includeTarget {
		physicalPath := frozenLintTargetPhysicalPathID(target, s.fs)
		if previous, exists := contentByPhysicalPath[physicalPath]; exists && previous != targetContent {
			aliasesConflict = true
		}
		contentByPhysicalPath[physicalPath] = targetContent
		addEditorOverlayTarget(files, target, targetContent)
	}
	return files, aliasesConflict
}

func frozenLintTargetPhysicalPathID(target target.File, fs vfs.FS) string {
	filePath := target.CanonicalPath
	if filePath == "" {
		filePath = target.Path
	}
	caseSensitive := true
	if fs != nil {
		caseSensitive = fs.UseCaseSensitiveFileNames()
	}
	return lspLexicalPathID(filePath, caseSensitive)
}

type frozenLintTargetOverlayFS struct {
	vfs.FS
	overlayFiles    map[string]string
	lexicalPathID   string
	canonicalPathID string
	canonicalPath   string
}

func newFrozenLintTargetOverlayFS(baseFS vfs.FS, files map[string]string, target target.File) vfs.FS {
	caseSensitive := true
	if baseFS != nil {
		caseSensitive = baseFS.UseCaseSensitiveFileNames()
	}
	canonicalPath := target.CanonicalPath
	if canonicalPath == "" {
		canonicalPath = target.Path
	}
	canonicalPath = tspath.NormalizePath(canonicalPath)
	return &frozenLintTargetOverlayFS{
		FS:              utils.NewOverlayVFS(baseFS, files),
		overlayFiles:    files,
		lexicalPathID:   lspLexicalPathID(target.Path, caseSensitive),
		canonicalPathID: lspLexicalPathID(canonicalPath, caseSensitive),
		canonicalPath:   canonicalPath,
	}
}

func (fs *frozenLintTargetOverlayFS) Realpath(filePath string) string {
	caseSensitive := fs.UseCaseSensitiveFileNames()
	pathID := lspLexicalPathID(filePath, caseSensitive)
	if pathID == fs.lexicalPathID || pathID == fs.canonicalPathID {
		return fs.canonicalPath
	}
	return config.FreezeFileIdentity(filePath, fs.FS).CanonicalPath
}

// A project can name an unsaved file through a different directory alias than
// its editor URI. Resolve that alias's existing ancestry before consulting the
// overlay; the missing suffix cannot be resolved by the disk filesystem alone.
func (fs *frozenLintTargetOverlayFS) ReadFile(filePath string) (string, bool) {
	if content, ok := fs.FS.ReadFile(filePath); ok {
		return content, true
	}
	if canonical, overlay := fs.overlayPath(filePath); overlay {
		return fs.FS.ReadFile(canonical)
	}
	return "", false
}

func (fs *frozenLintTargetOverlayFS) FileExists(filePath string) bool {
	if fs.FS.FileExists(filePath) {
		return true
	}
	_, overlay := fs.overlayPath(filePath)
	return overlay
}

func (fs *frozenLintTargetOverlayFS) Stat(filePath string) vfs.FileInfo {
	if info := fs.FS.Stat(filePath); info != nil {
		return info
	}
	if canonical, overlay := fs.overlayPath(filePath); overlay {
		return fs.FS.Stat(canonical)
	}
	return nil
}

func (fs *frozenLintTargetOverlayFS) DirectoryExists(directory string) bool {
	if fs.FS.DirectoryExists(directory) {
		return true
	}
	canonical := fs.Realpath(directory)
	for filePath := range fs.overlayFiles {
		if relative, within := config.RelativePathWithinConfigRoot(filePath, canonical, fs.UseCaseSensitiveFileNames()); within && relative != "" {
			return true
		}
	}
	return false
}

func (fs *frozenLintTargetOverlayFS) GetAccessibleEntries(directory string) vfs.Entries {
	entries := fs.FS.GetAccessibleEntries(directory)
	canonical := fs.Realpath(directory)
	if canonical == tspath.NormalizePath(directory) {
		return entries
	}
	entries.Files = slices.Clone(entries.Files)
	entries.Directories = slices.Clone(entries.Directories)
	// Project only editor entries. Keep the selected lexical buffer reachable
	// if its alias moves, and preserve the underlying directory's disk view.
	for filePath := range fs.overlayFiles {
		if relative, within := config.RelativePathWithinConfigRoot(filePath, canonical, fs.UseCaseSensitiveFileNames()); within && relative != "" {
			name, _, nested := strings.Cut(relative, "/")
			if nested {
				if !slices.Contains(entries.Directories, name) {
					entries.Directories = append(entries.Directories, name)
				}
			} else if !slices.Contains(entries.Files, name) {
				entries.Files = append(entries.Files, name)
			}
		}
	}
	slices.Sort(entries.Files)
	slices.Sort(entries.Directories)
	return entries
}

func (fs *frozenLintTargetOverlayFS) SourceHasBOM(filePath string) bool {
	if canonical, overlay := fs.overlayPath(filePath); overlay {
		return utils.SourceHasBOM(fs.FS, canonical)
	}
	return utils.SourceHasBOM(fs.FS, filePath)
}

func (fs *frozenLintTargetOverlayFS) overlayPath(filePath string) (string, bool) {
	canonical := fs.Realpath(filePath)
	_, overlay := fs.overlayFiles[utils.NormalizeAbsoluteDrive(tspath.NormalizePath(canonical))]
	return canonical, overlay
}

func addEditorOverlayTarget(files map[string]string, target target.File, content string) {
	if target.Path != "" {
		files[utils.NormalizeAbsoluteDrive(tspath.NormalizePath(target.Path))] = content
	}
	if target.CanonicalPath != "" {
		files[utils.NormalizeAbsoluteDrive(tspath.NormalizePath(target.CanonicalPath))] = content
	}
}

func (s *Server) addEditorOverlayFile(files map[string]string, filePath string, content string) string {
	filePath = tspath.NormalizePath(filePath)
	// Use one overlay key for drive aliases, so the selected target's later
	// write replaces any competing editor content for its physical file.
	files[utils.NormalizeAbsoluteDrive(filePath)] = content
	caseSensitive := true
	if s.fs != nil {
		caseSensitive = s.fs.UseCaseSensitiveFileNames()
	}
	if s.fs == nil {
		return string(tspath.ToPath(filePath, "", caseSensitive))
	}
	if realPath := config.FreezeFileIdentity(filePath, s.fs).CanonicalPath; realPath != "" {
		files[utils.NormalizeAbsoluteDrive(realPath)] = content
		return string(tspath.ToPath(realPath, "", caseSensitive))
	}
	return string(tspath.ToPath(filePath, "", caseSensitive))
}
