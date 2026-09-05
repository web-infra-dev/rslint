package lsp

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
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

func gitignoreFileWatchers(cwd string, relativePatternSupport bool) []*lsproto.FileSystemWatcher {
	workspaceRoot := filepath.Clean(cwd)
	watchers := []*lsproto.FileSystemWatcher{
		fileSystemWatcher(workspaceRoot, "**/.gitignore", relativePatternSupport),
	}
	// Automatic discovery may select a config above the workspace. Exact
	// watchers on the strict lexical ancestors cover every possible source from
	// that config directory down to cwd without recursively watching siblings.
	for current := filepath.Dir(workspaceRoot); current != workspaceRoot; {
		watchers = append(watchers, fileSystemWatcher(current, ".gitignore", relativePatternSupport))
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return watchers
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
		if isGitignoreURI(uri) {
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
		// tsconfig changed — rebuild tsConfigPaths so type-aware rule filtering
		// stays in sync. Session already handles the project state update and
		// triggers RefreshDiagnostics for relinting.
		if err := s.rebuildTsConfigPaths(); err != nil {
			log.Printf("[rslint] Failed to rebuild tsconfig paths: %v", err)
		}
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

// resolveTsConfigPaths resolves parserOptions.project from a config while
// preserving each declared path. TypeScript resolves relative includes from
// that lexical location, so a symlinked tsconfig is not interchangeable with
// its physical target.
func (s *Server) resolveTsConfigPaths(cfg config.RslintConfig, cwd string) ([]string, error) {
	return resolveTsConfigPathsWithFS(cfg, cwd, s.fs)
}

// rebuildTsConfigPaths resolves parserOptions.project from the current config.
// Called when a tsconfig changes so that type-aware rule filtering stays in
// sync. Config transactions resolve the same declarations while preparing
// their snapshot.
//
// We resolve per-config directory into tsConfigPathsByConfig.
// A config whose parserOptions.project is empty and has no auto-detected
// tsconfig resolves to nil. Files governed by that config have no type info,
// without affecting files governed by other configs. A nested template or
// fixture config without a tsconfig must not change sibling config behavior.
func (s *Server) rebuildTsConfigPaths() error {
	var byConfig map[string][]string
	if len(s.jsConfigs) > 0 {
		byConfig = make(map[string][]string, len(s.jsConfigs))
		for dir, entries := range s.jsConfigs {
			paths, err := s.resolveTsConfigPaths(entries, dir)
			if err != nil {
				return fmt.Errorf("resolve tsconfig paths for %q: %w", dir, err)
			}
			byConfig[dir] = paths
		}
	}

	s.tsConfigPathsByConfig = byConfig
	s.invalidateLintProjectCaches()
	return nil
}
