package lsp

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/project"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

const gitignoreWatcherID project.WatcherID = "rslint-gitignore-policy"
const ancestorJSConfigWatcherID project.WatcherID = "rslint-ancestor-js-config"

// gitPolicyWatchDirectories includes the physical directory chains used by
// collection, plus the lexical workspace ancestry used by config discovery.
// Watching missing markers is necessary: creating .git changes the boundary.
func gitPolicyWatchDirectories(directories []config.DirectoryIdentity, files []config.PathIdentity) []string {
	seen := make(map[string]struct{})
	addAncestors := func(directory string) {
		for directory != "" {
			directory = tspath.NormalizePath(directory)
			seen[directory] = struct{}{}
			if strings.HasPrefix(directory, "//") && strings.Count(strings.Trim(directory, "/"), "/") <= 1 {
				break
			}
			parent := tspath.GetDirectoryPath(directory)
			if parent == directory {
				break
			}
			directory = parent
		}
	}
	for _, directory := range directories {
		addAncestors(directory.LexicalPath)
		addAncestors(directory.CanonicalPath)
	}
	for _, file := range files {
		addAncestors(lspGitDirectory(file))
	}
	result := make([]string, 0, len(seen))
	for directory := range seen {
		result = append(result, directory)
	}
	sort.Strings(result)
	return result
}

func gitPolicyFileWatchers(directories []config.DirectoryIdentity, sourceDirectories []string, relativePatterns bool) []*lsproto.FileSystemWatcher {
	seen := make(map[string]struct{})
	var watchers []*lsproto.FileSystemWatcher
	add := func(directory, pattern string) {
		key := directory + "/" + pattern
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		watchers = append(watchers, fileSystemWatcher(directory, pattern, relativePatterns))
	}
	for _, directory := range directories {
		for _, path := range []string{directory.LexicalPath, directory.CanonicalPath} {
			add(path, "**/.gitignore")
			add(path, "**/.git")
		}
	}
	for _, directory := range sourceDirectories {
		add(directory, ".gitignore")
		add(directory, ".git")
	}
	return watchers
}

func (s *Server) updateGitPolicyWatchers(ctx context.Context, roots []config.DirectoryIdentity, current []string) error {
	if !s.watchEnabled || s.outgoingQueue == nil {
		return nil
	}
	directorySet := make(map[string]struct{}, len(current)+len(s.committedGitWatchDirectories))
	for _, directory := range append(append([]string(nil), current...), s.committedGitWatchDirectories...) {
		directorySet[directory] = struct{}{}
	}
	directories := make([]string, 0, len(directorySet))
	for directory := range directorySet {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	// Exact source chains and recursive roots are different subscriptions. An
	// alias can move between already observed directories without changing the
	// exact-directory union, while its recursive watch must still be replaced.
	scopeKeys := make([]string, 0, len(directories)+2*len(roots))
	for _, directory := range directories {
		scopeKeys = append(scopeKeys, "exact:"+directory)
	}
	for _, root := range roots {
		scopeKeys = append(scopeKeys, "recursive:"+root.LexicalPath, "recursive:"+root.CanonicalPath)
	}
	sort.Strings(scopeKeys)
	key := strings.Join(scopeKeys, "\x00")
	if key != s.gitPolicyWatcherKey {
		relativePatterns := false
		if s.initializeParams != nil && s.initializeParams.Capabilities != nil && s.initializeParams.Capabilities.Workspace != nil && s.initializeParams.Capabilities.Workspace.DidChangeWatchedFiles != nil {
			relativePatterns = ptrIsTrue(s.initializeParams.Capabilities.Workspace.DidChangeWatchedFiles.RelativePatternSupport)
		}
		s.gitPolicyWatcherGeneration++
		id := project.WatcherID(fmt.Sprintf("%s-%d", gitignoreWatcherID, s.gitPolicyWatcherGeneration))
		if err := s.WatchFiles(ctx, id, gitPolicyFileWatchers(roots, directories, relativePatterns)); err != nil {
			// A failed or canceled response may still have installed this ID.
			// Keep it in the existing registration set for later cleanup.
			s.watchers.Add(id)
			return err
		}
		s.gitPolicyWatcherID, s.gitPolicyWatcherKey = id, key
	}
	// The current range is installed. Old registrations can cause redundant
	// notifications, but their cleanup cannot invalidate this configuration.
	for id := range s.watchers.Keys() {
		if id == s.gitPolicyWatcherID || !strings.HasPrefix(string(id), string(gitignoreWatcherID)+"-") {
			continue
		}
		if err := s.UnwatchFiles(ctx, id); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("[rslint] Failed to remove old Git policy watcher %s: %v", id, err)
		}
	}
	return ctx.Err()
}

// ancestorJSConfigFileWatchers covers the strict lexical ancestors that Go's
// automatic config discovery searches before walking the workspace. The
// extension already owns a workspace-scoped RelativePattern watcher, so the
// workspace itself is deliberately excluded to avoid duplicate refreshes.
// Register every filename separately: creating a higher-priority sibling in an
// ancestor directory can change the selected config even when another config
// basename already exists there.
func ancestorJSConfigFileWatchers(cwd string, relativePatternSupport bool) []*lsproto.FileSystemWatcher {
	workspaceRoot := filepath.Clean(cwd)
	watchers := make([]*lsproto.FileSystemWatcher, 0)
	for current := filepath.Dir(workspaceRoot); current != workspaceRoot; {
		for _, configName := range discovery.AutoJSConfigFileNames {
			watchers = append(watchers, fileSystemWatcher(current, configName, relativePatternSupport))
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return watchers
}

func fileSystemWatcher(baseDir string, pattern string, relativePatternSupport bool) *lsproto.FileSystemWatcher {
	if relativePatternSupport {
		uri := fileURIFromPath(baseDir)
		return &lsproto.FileSystemWatcher{
			GlobPattern: lsproto.PatternOrRelativePattern{
				RelativePattern: &lsproto.RelativePattern{
					BaseUri: lsproto.WorkspaceFolderOrURI{URI: &uri},
					Pattern: pattern,
				},
			},
		}
	}
	absolute := filepath.ToSlash(filepath.Join(baseDir, pattern))
	return &lsproto.FileSystemWatcher{
		GlobPattern: lsproto.PatternOrRelativePattern{Pattern: &absolute},
	}
}

func fileURIFromPath(filePath string) lsproto.URI {
	uriPath := filepath.ToSlash(filePath)
	if len(uriPath) >= 2 && uriPath[1] == ':' {
		uriPath = "/" + uriPath
	}
	return lsproto.URI((&url.URL{Scheme: "file", Path: uriPath}).String())
}

func (s *Server) invalidateLintProjectCaches() {
	if s.lintPrograms != nil {
		s.lintPrograms.Invalidate()
	}
	if s.lintSessionRoots != nil {
		s.lintSessionRoots.Invalidate()
	}
}

func validateRuleOptionsForConfig(
	entries config.RslintConfig,
	configDirectory string,
	catalog *rule.Catalog,
) (config.RslintConfig, error) {
	normalized, optionsErrs := config.ValidateRuleOptions(entries, catalog)
	if len(optionsErrs) == 0 {
		return normalized, nil
	}
	msgs := make([]string, len(optionsErrs))
	for i, optionsErr := range optionsErrs {
		msgs[i] = optionsErr.Error()
	}
	return nil, fmt.Errorf("invalid rule options for %q:\n%s", configDirectory, strings.Join(msgs, "\n"))
}

// handleDidChangeWatchedFiles handles file change notifications from the client.
func (s *Server) handleDidChangeWatchedFiles(ctx context.Context, params *lsproto.DidChangeWatchedFilesParams) error {
	if params == nil {
		return nil
	}
	if s.lintPrograms != nil &&
		s.lintPrograms.DidChangeWatchedFiles(params.Changes) {
		s.invalidateOpenDocumentDiagnostics()
		_ = s.RefreshDiagnostics(ctx)
	}

	// Preserve Session's original watched-file input. It owns configured
	// project identity and may need a disk event while an overlay is open.
	if s.session != nil {
		s.session.DidChangeWatchedFiles(ctx, params.Changes)
	}
	needsTypeInfoRebuild := false
	needsIgnoreRefresh := false
	needsAncestorJSConfigRefresh := false
	for _, change := range params.Changes {
		uri := string(change.Uri)
		if isTsConfigURI(uri) {
			needsTypeInfoRebuild = true
		}
		if isGitignoreURI(uri) || isGitMarkerURI(uri) {
			needsIgnoreRefresh = true
		}
		if isStrictAncestorAutoJSConfigPath(uriToPath(change.Uri), s.cwd, s.fs) {
			needsAncestorJSConfigRefresh = true
		}
	}
	needsAutomaticAncestorRefresh := needsAncestorJSConfigRefresh && s.configRefreshConfigPath == ""
	if (needsIgnoreRefresh || needsAutomaticAncestorRefresh) && s.configDiscoveryActive {
		// didChangeWatchedFiles and configRefresh are both blocking methods, so
		// this direct call stays on the server's serialized dispatch loop and
		// cannot race an extension-initiated transaction. The workspace fallback
		// is part of the candidate snapshot, so a later activation failure keeps
		// the complete last-good generation live.
		reason := "gitignore-change"
		if needsAutomaticAncestorRefresh {
			reason = "config-change"
		}
		_, err := s.refreshConfig(ctx)
		if err == nil {
			return nil
		}
		// Discovery/activation failure preserves the complete last-good
		// generation, including its .gitignore view. Recompute diagnostics from
		// that committed view after an ignore event so invalidated editor
		// results are republished without leaking the rejected filesystem state. A
		// config-only failure has no independently live state to invalidate.
		log.Printf("[rslint] Failed to refresh config catalog after watched %s: %v", reason, err)
		if needsIgnoreRefresh {
			s.invalidateOpenDocumentDiagnostics()
			return s.RefreshDiagnostics(ctx)
		}
		return nil
	}
	if needsTypeInfoRebuild {
		// Discard project contents, including projects outside the Session.
		s.invalidateLintProjectCaches()
		s.invalidateOpenDocumentDiagnostics()
		return s.RefreshDiagnostics(ctx)
	}
	if needsIgnoreRefresh {
		s.invalidateOpenDocumentDiagnostics()
		return s.RefreshDiagnostics(ctx)
	}

	return nil
}

func isGitignoreURI(uri string) bool {
	idx := strings.LastIndex(uri, "/")
	return idx >= 0 && strings.EqualFold(uri[idx+1:], ".gitignore")
}

func isGitMarkerURI(uri string) bool {
	idx := strings.LastIndex(uri, "/")
	return idx >= 0 && strings.EqualFold(uri[idx+1:], ".git")
}

func isStrictAncestorAutoJSConfigPath(filePath string, cwd string, fsys vfs.FS) bool {
	if filePath == "" || cwd == "" || fsys == nil {
		return false
	}
	caseSensitive := fsys.CaseSensitivity().IsCaseSensitive()
	baseName := tspath.GetBaseFileName(tspath.NormalizePath(filePath))
	isAutoConfig := false
	for _, configName := range discovery.AutoJSConfigFileNames {
		if pathStringsEqual(baseName, configName, caseSensitive) {
			isAutoConfig = true
			break
		}
	}
	if !isAutoConfig {
		return false
	}
	directory := tspath.GetDirectoryPath(tspath.NormalizePath(filePath))
	workspace := tspath.NormalizePath(cwd)
	return !pathStringsEqual(directory, workspace, caseSensitive) &&
		pathutil.StartsWithDirectory(workspace, directory, caseSensitive)
}

func pathStringsEqual(left string, right string, caseSensitive bool) bool {
	if caseSensitive {
		return left == right
	}
	return strings.EqualFold(left, right)
}

func isTsConfigURI(uri string) bool {
	idx := strings.LastIndex(uri, "/")
	if idx < 0 {
		return false
	}
	name := uri[idx+1:]
	return (strings.HasPrefix(name, "tsconfig") || strings.HasPrefix(name, "jsconfig")) &&
		strings.HasSuffix(name, ".json")
}
