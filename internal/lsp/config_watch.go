package lsp

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net/url"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/project"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/rule"
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

func gitPolicyFileWatchers(directories []config.DirectoryIdentity, sourceDirectories, configFiles []string, relativePatterns bool) []*lsproto.FileSystemWatcher {
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
	for _, path := range configFiles {
		add(tspath.GetDirectoryPath(path), tspath.GetBaseFileName(path))
	}
	return watchers
}

func (s *Server) updateGitPolicyWatchers(ctx context.Context, roots []config.DirectoryIdentity, currentDirectories []string, currentConfigs map[string]string) error {
	configs := make(map[string]string, len(currentConfigs)+len(s.gitConfigWatchFiles))
	maps.Copy(configs, s.gitConfigWatchFiles)
	maps.Copy(configs, currentConfigs)
	if !s.watchEnabled || s.outgoingQueue == nil {
		s.gitConfigWatchFiles = configs
		return nil
	}
	directories := slices.Concat(currentDirectories, s.committedGitWatchDirectories)
	slices.Sort(directories)
	directories = slices.Compact(directories)
	configFiles := slices.Sorted(maps.Values(configs))
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
	for id := range configs {
		scopeKeys = append(scopeKeys, "config:"+id)
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
		if err := s.WatchFiles(ctx, id, gitPolicyFileWatchers(roots, directories, configFiles, relativePatterns)); err != nil {
			return err
		}
		s.gitPolicyWatcherID, s.gitPolicyWatcherKey = id, key
	}
	s.gitConfigWatchFiles = configs
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
	if params == nil || len(params.Changes) == 0 {
		return nil
	}
	s.documentPathNames = nil
	if s.lintPrograms != nil &&
		s.lintPrograms.DidChangeWatchedFiles(params.Changes) {
		s.invalidateOpenDocumentDiagnostics()
		_ = s.RefreshDiagnostics(ctx)
	}

	needsTypeInfoRebuild := false
	needsIgnoreRefresh := false
	needsAncestorJSConfigRefresh := false
	for _, change := range params.Changes {
		uri := string(change.Uri)
		path := uriToPath(change.Uri)
		if isTsConfigURI(uri) {
			needsTypeInfoRebuild = true
		}
		if isGitignoreURI(uri) || isGitMarkerURI(uri) || s.isWatchedGitConfig(path) {
			needsIgnoreRefresh = true
		}
		if isStrictAncestorAutoJSConfigPath(path, s.cwd, s.fs) {
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
		// Discard resident project contents and parsed configuration metadata.
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

func (s *Server) isWatchedGitConfig(path string) bool {
	if len(s.gitConfigWatchFiles) == 0 {
		return false
	}
	if s.gitConfigWatchFiles[config.ExactPathID(path)] != "" {
		return true
	}
	name := tspath.GetBaseFileName(path)
	if !strings.EqualFold(name, "config") && !strings.EqualFold(name, "config.worktree") {
		return false
	}
	// A client may report an alias of the physical repository directory. Use
	// the parent as well so deletion of a previously present file still matches.
	file := config.FreezeFileIdentity(path, s.fs)
	return s.gitConfigWatchFiles[config.ExactPathID(file.CanonicalPath)] != "" ||
		s.gitConfigWatchFiles[config.ExactPathID(tspath.CombinePaths(file.CanonicalParentPath, name))] != ""
}

func isStrictAncestorAutoJSConfigPath(filePath string, cwd string, fsys vfs.FS) bool {
	if filePath == "" || cwd == "" || fsys == nil {
		return false
	}
	caseSensitive := fsys.UseCaseSensitiveFileNames()
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
		tspath.StartsWithDirectory(workspace, directory, caseSensitive)
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
