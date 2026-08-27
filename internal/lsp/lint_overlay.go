package lsp

import (
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"

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

func (s *Server) currentEditorOverlayFSForTargetWithConflicts(
	uri lsproto.DocumentUri,
	target target.File,
) (vfs.FS, bool) {
	content, open := s.documents[uri]
	files, aliasesConflict := s.currentEditorOverlayFilesForFrozenTarget(uri, target, content, open)
	return newFrozenLintTargetOverlayFS(s.fs, files, target), aliasesConflict
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
	return fs.FS.Realpath(filePath)
}

func addEditorOverlayTarget(files map[string]string, target target.File, content string) {
	if target.Path != "" {
		files[tspath.NormalizePath(target.Path)] = content
	}
	if target.CanonicalPath != "" {
		files[tspath.NormalizePath(target.CanonicalPath)] = content
	}
}

func (s *Server) addEditorOverlayFile(files map[string]string, filePath string, content string) string {
	filePath = tspath.NormalizePath(filePath)
	files[filePath] = content
	caseSensitive := true
	if s.fs != nil {
		caseSensitive = s.fs.UseCaseSensitiveFileNames()
	}
	if s.fs == nil {
		return string(tspath.ToPath(filePath, "", caseSensitive))
	}
	if realPath := s.fs.Realpath(filePath); realPath != "" {
		realPath = tspath.NormalizePath(realPath)
		files[realPath] = content
		return string(tspath.ToPath(realPath, "", caseSensitive))
	}
	return string(tspath.ToPath(filePath, "", caseSensitive))
}
