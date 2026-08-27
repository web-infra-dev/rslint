package lsp

import (
	"net/url"
	"unicode"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
)

// documentLintSnapshot freezes every config-sensitive fact for one editor
// operation. Owner selection, project selection, Go rules, plugin rules,
// and fixes must all consume this same target/config pair; none may rediscover
// the target identity from the filesystem mid-operation.
type documentLintSnapshot struct {
	target                target.File
	config                config.RslintConfig
	resolvedConfig        config.ResolvedFileConfig
	pathSpaces            *config.PathSpaceSnapshot
	ruleCatalog           *rule.Catalog
	configResolved        bool
	typeScriptConfigPaths []string
	usesJavaScriptConfig  bool
	pluginGeneration      string
	unavailable           bool
}

func (s *Server) currentRuleCatalog() *rule.Catalog {
	if s == nil || s.ruleCatalog == nil {
		panic("LSP rule catalog is not initialized")
	}
	return s.ruleCatalog
}

func resolveDocumentLintSnapshotConfig(
	snapshot documentLintSnapshot,
	fs vfs.FS,
) documentLintSnapshot {
	if snapshot.ruleCatalog == nil {
		panic("document lint snapshot requires a rule catalog")
	}
	if snapshot.pathSpaces == nil {
		panic("document lint snapshot requires a path-space snapshot")
	}
	resolver, err := config.NewFileConfigResolverWithPathSpaces(
		snapshot.config,
		snapshot.target.ConfigDirectory,
		fs,
		snapshot.pathSpaces,
		snapshot.ruleCatalog,
		snapshot.usesJavaScriptConfig,
	)
	if err != nil {
		panic(err)
	}
	snapshot.resolvedConfig = resolver.ResolveTarget(snapshot.target.Identity())
	snapshot.configResolved = true
	return snapshot
}
func isLintableScriptFile(uri lsproto.DocumentUri) bool {
	return config.IsSupportedLintFile(uriToPath(uri))
}

func uriToPath(uri lsproto.DocumentUri) string {
	// Convert file:// URI to file path using net/url for proper percent-decoding.
	// Handles spaces (%20), CJK characters, and other encoded chars in paths.
	// file:///home/user       → /home/user  (Unix)
	// file:///C:/Users        → C:/Users    (Windows — strip the leading slash)
	// file:///path%20name/f   → /path name/f
	uriStr := string(uri)
	if uriStr == "" {
		return ""
	}
	u, err := url.ParseRequestURI(uriStr)
	if err != nil {
		return uriStr // fallback: return as-is for non-URI strings
	}
	p := u.Path
	if u.Host != "" {
		return "//" + u.Host + p
	}
	// Windows drive letter: /C:/... → C:/...
	if len(p) >= 3 && p[0] == '/' && unicode.IsLetter(rune(p[1])) && p[2] == ':' {
		return p[1:]
	}
	return p
}
func lspFilesystemPathID(filePath string, fs vfs.FS) string {
	filePath = tspath.NormalizePath(filePath)
	caseSensitive := true
	if fs != nil {
		caseSensitive = fs.UseCaseSensitiveFileNames()
		if realPath := fs.Realpath(filePath); realPath != "" {
			filePath = tspath.NormalizePath(realPath)
		}
	}
	return lspLexicalPathID(filePath, caseSensitive)
}

func lspTargetIdentity(filePath string, fs vfs.FS) target.File {
	return target.File{
		PathIdentity: target.FreezeFileIdentity(filePath, fs),
	}
}

func lspConfigTarget(filePath string, configDirectory string, fs vfs.FS) target.File {
	target := lspTargetIdentity(filePath, fs)
	target.ConfigDirectory = tspath.NormalizePath(configDirectory)
	return target
}

func isDefaultExcludedLintPath(filePath string, cwd string, fs vfs.FS) bool {
	useCaseSensitive := true
	if fs != nil {
		useCaseSensitive = fs.UseCaseSensitiveFileNames()
	}
	return config.IsDefaultExcludedPath(filePath, cwd, useCaseSensitive)
}

type documentConfigSelection struct {
	entries       config.RslintConfig
	resolved      config.ResolvedFileConfig
	pathSpaces    *config.PathSpaceSnapshot
	ruleCatalog   *rule.Catalog
	directory     string
	configKey     string
	usesJSConfig  bool
	configMissing bool
}

func (s *Server) selectDocumentConfig(lintFile target.File) documentConfigSelection {
	jsRuleCatalog := s.currentRuleCatalog()
	evaluationFS := s.fs
	jsOwnerIndex := s.jsConfigOwnerIndex
	jsFileConfigResolvers := s.jsFileConfigResolvers
	if !s.configSnapshotIncludesGitignore {
		// The target identity was frozen before entering this function. Bootstrap
		// config evaluation now gets its own short-lived filesystem generation so
		// owner selection, Git collection, and authored path bases cannot observe
		// different config-directory aliases within one document operation.
		if s.fs != nil {
			evaluationFS = newConfigSnapshotFS(bundled.WrapFS(cachedvfs.From(s.fs)))
		}
		if len(s.jsConfigs) > 0 {
			jsOwnerIndex = target.NewOwnerIndex(s.jsConfigs, evaluationFS)
		}
	}
	evaluateKnownConfig := func(
		entries config.RslintConfig,
		configDirectory string,
		ownerIndex *target.OwnerIndex,
		fileConfigResolver *config.FileConfigResolver,
		catalog *rule.Catalog,
		enforcePlugins bool,
	) (config.RslintConfig, config.ResolvedFileConfig, *config.PathSpaceSnapshot, bool) {
		configDirectory = tspath.NormalizePath(configDirectory)
		if !s.configSnapshotIncludesGitignore {
			entries = config.ConfigWithGitignoreForExactTarget(
				entries,
				configDirectory,
				evaluationFS,
				lintFile.Identity(),
			)
			ownerIndex = nil
			fileConfigResolver = nil
		}
		if ownerIndex == nil {
			ownerIndex = target.NewOwnerIndex(
				map[string]config.RslintConfig{configDirectory: entries},
				evaluationFS,
			)
		}
		if _, ok := ownerIndex.PathSpaces().PhysicalDirectory(configDirectory); !ok {
			return nil, config.ResolvedFileConfig{}, ownerIndex.PathSpaces(), false
		}
		if fileConfigResolver == nil {
			if s.configSnapshotIncludesGitignore {
				panic("committed config snapshot is missing its file config resolver")
			}
			var err error
			fileConfigResolver, err = config.NewFileConfigResolverWithPathSpaces(
				entries,
				configDirectory,
				evaluationFS,
				ownerIndex.PathSpaces(),
				catalog,
				enforcePlugins,
			)
			if err != nil {
				panic(err)
			}
		}
		return entries, fileConfigResolver.ResolveTarget(lintFile.Identity()), ownerIndex.PathSpaces(), true
	}

	if len(s.jsConfigs) > 0 {
		if s.configRefreshConfigPath != "" {
			configKey := tspath.GetDirectoryPath(
				tspath.NormalizePath(s.configRefreshConfigPath),
			)
			if entries, active := s.jsConfigs[configKey]; active {
				entries, resolved, pathSpaces, ok := evaluateKnownConfig(
					entries,
					configKey,
					jsOwnerIndex,
					jsFileConfigResolvers[configKey],
					jsRuleCatalog,
					true,
				)
				return documentConfigSelection{
					entries:       entries,
					resolved:      resolved,
					pathSpaces:    pathSpaces,
					ruleCatalog:   jsRuleCatalog,
					directory:     configKey,
					configKey:     configKey,
					usesJSConfig:  true,
					configMissing: !ok,
				}
			}
		} else if jsOwnerIndex != nil {
			configKey, owned := jsOwnerIndex.Resolve(lintFile.Identity())
			if owned {
				entries := s.jsConfigs[configKey]
				entries, resolved, pathSpaces, ok := evaluateKnownConfig(
					entries,
					configKey,
					jsOwnerIndex,
					jsFileConfigResolvers[configKey],
					jsRuleCatalog,
					true,
				)
				return documentConfigSelection{
					entries:       entries,
					resolved:      resolved,
					pathSpaces:    pathSpaces,
					ruleCatalog:   jsRuleCatalog,
					directory:     configKey,
					configKey:     configKey,
					usesJSConfig:  true,
					configMissing: !ok,
				}
			}
		}
	}

	configDirectory := tspath.NormalizePath(s.cwd)
	goRuleCatalog := rules.All()
	entries, resolved, pathSpaces, ok := evaluateKnownConfig(
		s.jsonConfig,
		configDirectory,
		s.jsonConfigOwnerIndex,
		s.jsonFileConfigResolver,
		goRuleCatalog,
		false,
	)
	return documentConfigSelection{
		entries:       entries,
		resolved:      resolved,
		pathSpaces:    pathSpaces,
		ruleCatalog:   goRuleCatalog,
		directory:     configDirectory,
		configMissing: !ok,
	}
}

// documentLintSnapshot resolves one immutable target identity, then derives
// its owner, config, and TypeScript projects from that same identity. This is
// the only production entry from an editor URI into lint configuration.
func (s *Server) documentLintSnapshot(uri lsproto.DocumentUri) documentLintSnapshot {
	target := lspTargetIdentity(uriToPath(uri), s.fs)
	selection := s.selectDocumentConfig(target)
	target.ConfigDirectory = selection.directory
	typeScriptConfigPaths := s.tsConfigPaths
	if selection.usesJSConfig {
		typeScriptConfigPaths = s.tsConfigPathsByConfig[selection.configKey]
	}
	_, unavailable := s.jsUnavailableConfigs[selection.configKey]
	return documentLintSnapshot{
		target:                target,
		config:                selection.entries,
		resolvedConfig:        selection.resolved,
		pathSpaces:            selection.pathSpaces,
		ruleCatalog:           selection.ruleCatalog,
		configResolved:        !selection.configMissing,
		typeScriptConfigPaths: typeScriptConfigPaths,
		usesJavaScriptConfig:  selection.usesJSConfig,
		pluginGeneration:      s.eslintPluginConfigGeneration,
		unavailable:           selection.usesJSConfig && unavailable,
	}
}

// getConfigForURI is retained for package-level helpers and tests. Production
// lint operations use documentLintSnapshot so config and project ownership
// cannot be resolved in separate filesystem observations.
func (s *Server) getConfigForURI(uri lsproto.DocumentUri) (config.RslintConfig, string, bool) {
	target := lspTargetIdentity(uriToPath(uri), s.fs)
	selection := s.selectDocumentConfig(target)
	return selection.entries, selection.directory, selection.usesJSConfig
}

func (s *Server) getLintConfigForURI(uri lsproto.DocumentUri) (config.RslintConfig, string, bool) {
	snapshot := s.documentLintSnapshot(uri)
	return snapshot.config, snapshot.target.ConfigDirectory, snapshot.usesJavaScriptConfig
}

func (s *Server) isUnavailableConfigForURI(uri lsproto.DocumentUri) bool {
	target := lspTargetIdentity(uriToPath(uri), s.fs)
	configKey, ok := s.jsConfigKeyForTarget(target)
	if !ok {
		return false
	}
	_, unavailable := s.jsUnavailableConfigs[configKey]
	return unavailable
}

// nearestJSConfigKey returns the nearest supplied JS/TS config directory for
// uri. Matching uses normalized filesystem paths instead of URI identity;
// lexical ownership is tried before a realpath fallback.
func (s *Server) nearestJSConfigKey(uri lsproto.DocumentUri) (string, bool) {
	return s.jsConfigKeyForTarget(lspTargetIdentity(uriToPath(uri), s.fs))
}

func (s *Server) jsConfigKeyForTarget(target target.File) (string, bool) {
	if len(s.jsConfigs) == 0 {
		return "", false
	}
	// An explicitly selected config governs every editor target, while its
	// directory remains the base for files/ignores matching. Ancestry
	// models automatic config discovery, not this invocation-wide scope.
	if s.configRefreshConfigPath != "" {
		configDir := tspath.GetDirectoryPath(tspath.NormalizePath(s.configRefreshConfigPath))
		_, active := s.jsConfigs[configDir]
		return configDir, active
	}
	if s.jsConfigOwnerIndex == nil {
		return "", false
	}
	configDir, _ := s.jsConfigOwnerIndex.Resolve(target.Identity())
	if configDir == "" {
		return "", false
	}
	_, active := s.jsConfigs[configDir]
	return configDir, active
}

// tsConfigPathsForURI returns parserOptions.project paths from the config owner
// selected by getConfigForURI. A nested config with no tsconfig therefore does
// not affect type-info decisions for sibling configs.
//
// A nil return means the governing config has no resolved tsconfig, so callers
// must disable type-aware rules for this file.
func (s *Server) tsConfigPathsForURI(uri lsproto.DocumentUri) []string {
	target := lspTargetIdentity(uriToPath(uri), s.fs)
	if configKey, ok := s.jsConfigKeyForTarget(target); ok {
		return s.tsConfigPathsByConfig[configKey]
	}
	return s.tsConfigPaths
}
