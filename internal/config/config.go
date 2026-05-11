package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/typescript-go/shim/tspath"
	importPlugin "github.com/web-infra-dev/rslint/internal/plugins/import"
	jestPlugin "github.com/web-infra-dev/rslint/internal/plugins/jest"
	jsxA11yPlugin "github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y"
	promisePlugin "github.com/web-infra-dev/rslint/internal/plugins/promise"
	reactPlugin "github.com/web-infra-dev/rslint/internal/plugins/react"
	reactHooksPlugin "github.com/web-infra-dev/rslint/internal/plugins/react_hooks"
	stylisticPlugin "github.com/web-infra-dev/rslint/internal/plugins/stylistic"
	typescriptPlugin "github.com/web-infra-dev/rslint/internal/plugins/typescript"
	unicornPlugin "github.com/web-infra-dev/rslint/internal/plugins/unicorn"
	coreRules "github.com/web-infra-dev/rslint/internal/rules"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// RslintConfig represents the top-level configuration array
type RslintConfig []ConfigEntry

// ConfigEntry represents a single configuration entry in the config array
type ConfigEntry struct {
	Files           []string         `json:"files,omitempty"`
	Ignores         []string         `json:"ignores,omitempty"`
	LanguageOptions *LanguageOptions `json:"languageOptions,omitempty"`
	Rules           Rules            `json:"rules"`
	Plugins         []string         `json:"plugins,omitempty"`
	Settings        Settings         `json:"settings,omitempty"`
	// EslintPlugins carries user-supplied ESLint-compat plugin entries.
	// Orthogonal to Plugins (which gates native rules by name only).
	// Each entry's Prefix becomes the rule namespace ("uc/no-null"); the
	// runner imports the plugin from ResolvedPath at lint time.
	// Populated Node-side from the JS config's `eslintPlugins: { ns: pluginObj }`.
	EslintPlugins []EslintPluginEntry `json:"eslintPlugins,omitempty"`
}

// Settings represents shared settings accessible to rules
type Settings map[string]interface{}

// LanguageOptions contains language-specific configuration options.
//
// Architecture (typed-native + opaque-compat):
//
//   - `ParserOptions.Project` / `ParserOptions.ProjectService` are typed
//     because Go internals read them directly (ts-go Program creation,
//     tsconfig discovery).
//   - Every OTHER `languageOptions.*` field — `globals`, `parser`, future
//     ESLint flat-config additions — flows through `Compat` opaquely.
//     Go never introspects this map; the field-merge semantics fall out
//     of generic JSON deep-merge (object keys later-win at each path);
//     the worker decodes the fields it consumes (globals → scope-manager,
//     parserOptions.ecmaVersion → oxc-parser, etc.).
//
// This decouples the wire / runner contract from Go-side knowledge of
// individual ESLint compat fields. New flat-config fields land in the
// runner alongside the TS type signature; Go needs ZERO changes.
type LanguageOptions struct {
	// Typed: native consumers read this struct directly.
	ParserOptions *ParserOptions `json:"-"`

	// Opaque: all non-`parserOptions` top-level fields the user wrote.
	// Generic deep-merged across config entries; forwarded to the worker
	// as part of the wire payload. Never introspected by Go.
	Compat map[string]any `json:"-"`
}

// UnmarshalJSON splits the incoming user JSON into the typed
// `ParserOptions` (when present) and the rest into `Compat`. Custom
// because the standard struct decoder would silently drop any field
// outside the typed set — which is the whole opacity-loss problem
// we're avoiding here.
func (lo *LanguageOptions) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if po, ok := raw["parserOptions"]; ok {
		delete(raw, "parserOptions")
		lo.ParserOptions = &ParserOptions{}
		if err := json.Unmarshal(po, lo.ParserOptions); err != nil {
			return err
		}
	}
	if len(raw) > 0 {
		lo.Compat = make(map[string]any, len(raw))
		for k, v := range raw {
			var val any
			if err := json.Unmarshal(v, &val); err != nil {
				return err
			}
			lo.Compat[k] = val
		}
	}
	return nil
}

// MarshalJSON inverts UnmarshalJSON: emits a flat object that mirrors
// what the user originally wrote, so downstream JSON consumers (wire,
// `config_init` formatting, debug printing) see a single canonical shape.
func (lo LanguageOptions) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, len(lo.Compat)+1)
	for k, v := range lo.Compat {
		out[k] = v
	}
	if lo.ParserOptions != nil {
		out["parserOptions"] = lo.ParserOptions
	}
	return json.Marshal(out)
}

// ToCompatWire serializes LanguageOptions into the flat shape the
// runner consumes — EXCLUDING the native-only fields
// (`parserOptions.project`, `parserOptions.projectService`) which the
// worker has no business seeing. The result is the exact `map[string]any`
// that lands in `linter.CompatLintFile.LanguageOptions`.
//
// Returns nil when nothing compat-relevant is present, so JSON
// omitempty on the wire field actually drops it.
func (lo *LanguageOptions) ToCompatWire() map[string]any {
	if lo == nil {
		return nil
	}
	out := make(map[string]any, len(lo.Compat)+1)
	for k, v := range lo.Compat {
		out[k] = v
	}
	if lo.ParserOptions != nil && len(lo.ParserOptions.Compat) > 0 {
		nested := make(map[string]any, len(lo.ParserOptions.Compat))
		for k, v := range lo.ParserOptions.Compat {
			nested[k] = v
		}
		out["parserOptions"] = nested
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ProjectPaths represents project paths that can be either a single string or an array of strings
type ProjectPaths []string

// UnmarshalJSON implements custom JSON unmarshaling to support both string and string[] formats
func (p *ProjectPaths) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var singlePath string
	if err := json.Unmarshal(data, &singlePath); err == nil {
		*p = []string{singlePath}
		return nil
	}

	// If that fails, try to unmarshal as array of strings
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return err
	}
	*p = paths
	return nil
}

// ParserOptions contains parser-specific configuration.
//
// Typed fields (`Project` / `ProjectService`) drive ts-go's Program
// creation and tsconfig discovery — Go reads them directly.
//
// `Compat` captures every other parserOptions field the user wrote
// (`ecmaVersion`, `sourceType`, `ecmaFeatures`, `parser`, future flat-
// config additions). Generic deep-merged across config entries; opaque
// to Go. The worker decodes the fields it consumes when shaping the
// LintTask request.
type ParserOptions struct {
	ProjectService *bool        `json:"-"`
	Project        ProjectPaths `json:"-"`

	// Opaque: all non-`project` / non-`projectService` parserOptions
	// fields. Forwarded to the runner; never read by Go.
	Compat map[string]any `json:"-"`
}

// UnmarshalJSON extracts the typed fields (`project`, `projectService`)
// from the user's parserOptions blob and stashes everything else in
// `Compat`. Same rationale as `LanguageOptions.UnmarshalJSON`: a plain
// struct decode would silently drop unknown fields.
func (po *ParserOptions) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["projectService"]; ok {
		delete(raw, "projectService")
		var b bool
		if err := json.Unmarshal(v, &b); err != nil {
			return fmt.Errorf("parserOptions.projectService: %w", err)
		}
		po.ProjectService = &b
	}
	if v, ok := raw["project"]; ok {
		delete(raw, "project")
		if err := json.Unmarshal(v, &po.Project); err != nil {
			return fmt.Errorf("parserOptions.project: %w", err)
		}
	}
	if len(raw) > 0 {
		po.Compat = make(map[string]any, len(raw))
		for k, v := range raw {
			var val any
			if err := json.Unmarshal(v, &val); err != nil {
				return err
			}
			po.Compat[k] = val
		}
	}
	return nil
}

// MarshalJSON emits the flat shape that mirrors what the user wrote —
// typed fields ride alongside Compat. Used for debug printing and any
// internal callers that re-serialize the struct (config_init's
// formatter, etc.); the WIRE shape goes through ToCompatWire above and
// intentionally excludes the typed native fields.
func (po ParserOptions) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, len(po.Compat)+2)
	for k, v := range po.Compat {
		out[k] = v
	}
	if po.ProjectService != nil {
		out["projectService"] = *po.ProjectService
	}
	if len(po.Project) > 0 {
		out["project"] = po.Project
	}
	return json.Marshal(out)
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// Rules represents the rules configuration
// This can be extended to include specific rule configurations
type Rules map[string]interface{}

// RuleConfig represents individual rule configuration
type RuleConfig struct {
	Level   string      `json:"level,omitempty"`   // "error", "warn", "off"
	Options interface{} `json:"options,omitempty"` // Rule-specific options (string, map, array, etc.)
}

// IsEnabled returns true if the rule is enabled (not "off")
func (rc *RuleConfig) IsEnabled() bool {
	if rc == nil {
		return false
	}
	return rc.Level != "off" && rc.Level != ""
}

// GetLevel returns the rule level, defaulting to "error" if not specified
func (rc *RuleConfig) GetLevel() string {
	if rc == nil || rc.Level == "" {
		return "error"
	}
	return rc.Level
}

// GetOptions returns the rule options, ensuring we return a usable value
func (rc *RuleConfig) GetOptions() interface{} {
	if rc == nil || rc.Options == nil {
		return nil
	}
	return rc.Options
}

// SetOptions sets the rule options
func (rc *RuleConfig) SetOptions(options interface{}) {
	if rc != nil {
		rc.Options = options
	}
}

// GetSeverity returns the diagnostic severity for this rule configuration
func (rc *RuleConfig) GetSeverity() rule.DiagnosticSeverity {
	if rc == nil {
		return rule.SeverityError
	}
	return rule.ParseSeverity(rc.Level)
}

// PluginInfo defines a known plugin with its rule prefix and all accepted declaration names.
type PluginInfo struct {
	RulePrefix  string   // Rule name prefix, e.g. "import"
	DeclNames   []string // All accepted declaration names, e.g. ["eslint-plugin-import", "import"]
	getAllRules func() []rule.Rule
}

// KnownPlugins is the single source of truth for all supported plugins.
var KnownPlugins = []PluginInfo{
	{
		RulePrefix:  "@typescript-eslint",
		DeclNames:   []string{"@typescript-eslint"},
		getAllRules: func() []rule.Rule { return typescriptPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "import",
		DeclNames:   []string{"eslint-plugin-import", "import"},
		getAllRules: func() []rule.Rule { return importPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "jest",
		DeclNames:   []string{"eslint-plugin-jest", "jest"},
		getAllRules: func() []rule.Rule { return jestPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "jsx-a11y",
		DeclNames:   []string{"eslint-plugin-jsx-a11y", "jsx-a11y"},
		getAllRules: func() []rule.Rule { return jsxA11yPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "promise",
		DeclNames:   []string{"eslint-plugin-promise", "promise"},
		getAllRules: func() []rule.Rule { return promisePlugin.GetAllRules() },
	},
	{
		RulePrefix:  "react",
		DeclNames:   []string{"react"},
		getAllRules: func() []rule.Rule { return reactPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "react-hooks",
		DeclNames:   []string{"eslint-plugin-react-hooks", "react-hooks"},
		getAllRules: func() []rule.Rule { return reactHooksPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "@stylistic",
		DeclNames:   []string{"@stylistic", "@stylistic/eslint-plugin"},
		getAllRules: func() []rule.Rule { return stylisticPlugin.GetAllRules() },
	},
	{
		RulePrefix:  "unicorn",
		DeclNames:   []string{"eslint-plugin-unicorn", "unicorn"},
		getAllRules: func() []rule.Rule { return unicornPlugin.GetAllRules() },
	},
}

// pluginByDeclName is a lookup table built from KnownPlugins: declaration name → *PluginInfo.
var pluginByDeclName map[string]*PluginInfo

func init() {
	pluginByDeclName = make(map[string]*PluginInfo)
	for i := range KnownPlugins {
		for _, name := range KnownPlugins[i].DeclNames {
			pluginByDeclName[name] = &KnownPlugins[i]
		}
	}
}

// NormalizePluginName converts a plugin declaration name to its rule prefix form.
// Looks up KnownPlugins; returns the input unchanged if not found.
func NormalizePluginName(pluginName string) string {
	if info, ok := pluginByDeclName[pluginName]; ok {
		return info.RulePrefix
	}
	return pluginName
}

// parseArrayRuleConfig parses array-style rule configuration like ["error", {...options}]
// Supports ESLint-compatible formats:
// - ["off"] -> disabled rule
// - ["error"] -> enabled rule with error severity
// - ["warn"] -> enabled rule with warning severity
// - ["error", {...options}] -> enabled rule with error severity and options
// - ["error", "both"] -> enabled rule with string option (e.g. no-inner-declarations)
// - ["error", "both", {...options}] -> enabled rule with string + object options
func parseArrayRuleConfig(ruleArray []interface{}) *RuleConfig {
	if len(ruleArray) == 0 {
		return nil
	}

	// First element should always be the severity level
	level, ok := ruleArray[0].(string)
	if !ok {
		return nil
	}

	ruleConfig := &RuleConfig{Level: level}

	// Remaining elements are rule options — pass them through to the rule's
	// option parser which knows how to interpret its own format.
	if len(ruleArray) > 1 {
		// Store the FULL positional options array — ESLint's
		// `context.options` shape. Do NOT unwrap a single element here:
		// a single ARRAY-valued option (`['error', ['a', 'b']]`) would be
		// mangled into the options list, so the compat layer's
		// context.options came out as `['a','b']` instead of `[['a','b']]`.
		// The native path unwraps a lone option itself (rule_registry's
		// nativeRuleOptions) to preserve its legacy single-value shape.
		ruleConfig.Options = ruleArray[1:]
	}

	return ruleConfig
}

var registerOnce sync.Once

func RegisterAllRules() {
	registerOnce.Do(func() {
		registerAllTypeScriptEslintPluginRules()
		registerAllImportPluginRules()
		registerAllReactPluginRules()
		registerAllReactHooksPluginRules()
		registerAllJestPluginRules()
		registerAllJsxA11yPluginRules()
		registerAllPromisePluginRules()
		registerAllStylisticPluginRules()
		registerAllUnicornPluginRules()
		registerAllCoreEslintRules()
	})
}

func registerAllReactPluginRules() {
	for _, rule := range reactPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllReactHooksPluginRules() {
	for _, rule := range reactHooksPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllJestPluginRules() {
	for _, rule := range jestPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllJsxA11yPluginRules() {
	for _, rule := range jsxA11yPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllPromisePluginRules() {
	for _, rule := range promisePlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllStylisticPluginRules() {
	for _, rule := range stylisticPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllUnicornPluginRules() {
	for _, rule := range unicornPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllTypeScriptEslintPluginRules() {
	for _, rule := range typescriptPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllImportPluginRules() {
	for _, rule := range importPlugin.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

func registerAllCoreEslintRules() {
	for _, rule := range coreRules.GetAllRules() {
		GlobalRuleRegistry.Register(rule.Name, rule)
	}
}

// isFileIgnored checks if a file is matched by ignore patterns, evaluated sequentially.
// Later patterns override earlier ones; a `!` prefix negates (re-includes) a previously
// ignored file. This aligns with ESLint v10's ignore semantics.
//
// For directory-level blocking (dir/** prevents traversal entirely), use isDirPathBlocked.
func isFileIgnored(filePath string, ignorePatterns []string, cwd string) bool {
	if cwd == "" {
		return isFileIgnoredSimple(filePath, ignorePatterns)
	}

	// Normalize the file path relative to cwd
	normalizedPath := normalizePath(filePath, cwd)
	unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")

	// Evaluate patterns sequentially. Later patterns override earlier ones.
	// A `!` prefix negates (re-includes) a previously ignored file.
	// This aligns with ESLint v10's ignore semantics.
	ignored := false
	for _, pattern := range ignorePatterns {
		negated := false
		if strings.HasPrefix(pattern, "!") {
			negated = true
			pattern = pattern[1:]
		}

		normalizedPattern := normalizePattern(pattern)

		// Match against the relative path only. Do NOT fall back to the
		// absolute filePath — patterns with **/ prefix (e.g., **/tmp/**/*)
		// would incorrectly match system directory names in the absolute path
		// (e.g., /tmp/ on Linux/macOS).
		matched := matchGlob(normalizedPattern, normalizedPath)
		// Windows path separator fallback.
		if !matched && unixPath != normalizedPath {
			matched = matchGlob(normalizedPattern, unixPath)
		}

		if matched {
			ignored = !negated
		}
	}
	return ignored
}

// normalizePattern cleans up a glob pattern to match paths produced by normalizePath.
// normalizePath uses tspath.NormalizePath on file paths (strips leading "./", collapses
// "/./", resolves ".."), so patterns must undergo the same transformation.
// matchGlob matches a glob pattern against a path using doublestar.
func matchGlob(pattern, path string) bool {
	m, err := doublestar.Match(pattern, path)
	return err == nil && m
}

// isFileLevelPattern returns true if the pattern only matches files (not directories).
// File-level patterns end with /**/* or /* (but not /**).
// These do NOT block directory traversal in ESLint v10's isDirectoryIgnored.
func isFileLevelPattern(pattern string) bool {
	return strings.HasSuffix(pattern, "/**/*") ||
		(strings.HasSuffix(pattern, "/*") && !strings.HasSuffix(pattern, "/**"))
}

func normalizePattern(pattern string) string {
	return tspath.NormalizePath(pattern)
}

// isDirBlockedByIgnores checks if the file's directory is blocked by a
// directory-level ignore pattern (e.g., `dir/**`). File-level patterns
// (`dir/**/*`, `dir/*`) and negation patterns are skipped.
// This aligns with ESLint v10: `dir/**` blocks directory traversal entirely,
// and `!` negation cannot undo it.
func isDirBlockedByIgnores(filePath string, ignorePatterns []string, cwd string) bool {
	var dirPath string
	if cwd != "" {
		dirPath = normalizePath(tspath.GetDirectoryPath(filePath), cwd)
	} else {
		dirPath = tspath.GetDirectoryPath(filePath)
	}
	dirPath = strings.ReplaceAll(dirPath, "\\", "/")
	dirPath = strings.TrimSuffix(dirPath, "/")
	if dirPath == "" || dirPath == "." {
		return false
	}
	return isDirPathBlocked(dirPath, ignorePatterns)
}

// isDirPathBlocked checks if a directory path is blocked by any directory-level ignore
// pattern. Shared between GetConfigForFile and DiscoverGapFiles.
//
// A directory is blocked if a pattern matches the path itself or any parent segment.
// For example, pattern "dir1/**" blocks "dir1", "dir1/sub", and "dir1/sub/deep".
// File-level patterns (ending with /**/* or /*) and negation (!) patterns are skipped —
// directory blocking is absolute and cannot be negated.
func isDirPathBlocked(dirPath string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		if pattern == "" || strings.HasPrefix(pattern, "!") {
			continue
		}
		if isFileLevelPattern(pattern) {
			continue
		}

		normalizedPattern := normalizePattern(pattern)

		if matchGlob(normalizedPattern, dirPath) || matchGlob(normalizedPattern, dirPath+"/x") {
			return true
		}
		segments := strings.Split(dirPath, "/")
		for i := 1; i < len(segments); i++ {
			partial := strings.Join(segments[:i], "/")
			if matchGlob(normalizedPattern, partial) || matchGlob(normalizedPattern, partial+"/x") {
				return true
			}
		}
	}
	return false
}

// normalizePath converts file path to be relative to cwd for consistent matching
func normalizePath(filePath, cwd string) string {
	return tspath.NormalizePath(tspath.ConvertToRelativePath(filePath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: true,
		CurrentDirectory:          cwd,
	}))
}

// isFileIgnoredSimple provides fallback matching when cwd is unavailable
func isFileIgnoredSimple(filePath string, ignorePatterns []string) bool {
	ignored := false
	for _, pattern := range ignorePatterns {
		negated := false
		if strings.HasPrefix(pattern, "!") {
			negated = true
			pattern = pattern[1:]
		}
		normalizedPattern := normalizePattern(pattern)
		if matched, err := doublestar.Match(normalizedPattern, filePath); err == nil && matched {
			ignored = !negated
		}
	}
	return ignored
}

// MergedConfig is the final computed configuration for a single file
type MergedConfig struct {
	Rules           map[string]*RuleConfig
	Settings        Settings
	LanguageOptions *LanguageOptions
	Plugins         map[string]struct{}
	// EslintPlugins is the per-file union of every matching ConfigEntry's
	// EslintPlugins, in the order matching entries appeared. The Node side
	// has already done any plugin-level merging / validation before
	// sending entries over IPC, so Go neither coalesces nor deduplicates
	// here — the same prefix may appear twice when two configs share a
	// prefix at different resolvedPaths, which is intentional for
	// monorepo multi-version setups (per-file dispatch picks the right
	// plugin instance using the file's ConfigKey).
	// Nil when no matching entry contributed eslintPlugins.
	EslintPlugins []EslintPluginEntry
}

// IsFileIgnored reports whether filePath is excluded by the config's global
// `ignores` patterns. It is distinct from GetConfigForFile returning nil,
// which also covers "no entry matched this file" — callers that need ESLint's
// "ignores hides the file from the linter entirely" semantics (including
// type-check diagnostics and file counts) should use this method.
func (config RslintConfig) IsFileIgnored(filePath string, cwd string) bool {
	patterns := ExtractConfigIgnores(config)
	if len(patterns) == 0 {
		return false
	}
	return isDirBlockedByIgnores(filePath, patterns, cwd) ||
		isFileIgnored(filePath, patterns, cwd)
}

// GetConfigForFile computes the merged configuration for a file following ESLint flat config semantics.
// Returns nil if the file is globally ignored or no entry matches (should not be linted).
//
// Global ignore evaluation happens in two phases:
//  1. Directory-level (isDirBlockedByIgnores): patterns like dir/** block entire directories.
//     Negation (!) cannot override directory-level blocking.
//  2. File-level (isFileIgnored): sequential evaluation with ! negation support for re-inclusion.
//
// After global ignore check, entries are merged in order if their files match and ignores don't.
// cwd is the directory the config lives in; file paths are resolved relative to it.
func (config RslintConfig) GetConfigForFile(filePath string, cwd string) *MergedConfig {
	merged := &MergedConfig{
		Rules:   make(map[string]*RuleConfig),
		Plugins: make(map[string]struct{}),
	}

	// 1. Collect all global ignore patterns and evaluate once.
	// This allows `!` negation patterns in separate entries to work correctly,
	// aligned with ESLint v10 which merges all global ignores before evaluating.
	globalIgnorePatterns := ExtractConfigIgnores(config)
	if len(globalIgnorePatterns) > 0 {
		// Phase 1: directory-level check. Patterns like `dir/**` block the
		// directory entirely — `!` negation cannot undo this. Aligned with
		// ESLint v10's isDirectoryIgnored behavior.
		if isDirBlockedByIgnores(filePath, globalIgnorePatterns, cwd) {
			return nil
		}
		// Phase 2: file-level check with sequential `!` negation support.
		if isFileIgnored(filePath, globalIgnorePatterns, cwd) {
			return nil
		}
	}

	// Track whether any non-global entry matched this file
	entryMatched := false

	for _, entry := range config {
		if isGlobalIgnoreEntry(entry) {
			continue
		}

		// 2. files matching
		if len(entry.Files) > 0 && !isFileMatched(filePath, entry.Files, cwd) {
			continue
		}

		// 3. Entry-level ignores
		if isFileIgnored(filePath, entry.Ignores, cwd) {
			continue
		}

		entryMatched = true

		// 4. Rules: shallow merge, later entries override earlier ones
		for ruleName, ruleValue := range entry.Rules {
			switch v := ruleValue.(type) {
			case string:
				merged.Rules[ruleName] = &RuleConfig{Level: v}
			case []interface{}:
				if rc := parseArrayRuleConfig(v); rc != nil {
					merged.Rules[ruleName] = rc
				}
			case map[string]interface{}:
				ruleConfig := &RuleConfig{}
				if level, ok := v["level"].(string); ok {
					ruleConfig.Level = level
				}
				if options, ok := v["options"].(map[string]interface{}); ok {
					ruleConfig.Options = options
				}
				merged.Rules[ruleName] = ruleConfig
			}
		}

		// 5. Plugins: union from all matching entries (normalized to rule prefix form)
		for _, plugin := range entry.Plugins {
			merged.Plugins[NormalizePluginName(plugin)] = struct{}{}
		}

		// 6. Settings: shallow merge
		if entry.Settings != nil {
			if merged.Settings == nil {
				merged.Settings = make(Settings)
			}
			for k, v := range entry.Settings {
				merged.Settings[k] = v
			}
		}

		// 7. LanguageOptions: deep merge
		merged.LanguageOptions = mergeLanguageOptions(merged.LanguageOptions, entry.LanguageOptions)

		// 8. EslintPlugins: append across matching entries. The Node side
		//    coalesces per-prefix rule names across configs in
		//    `cli.ts::runViaEngine` before sending entries over IPC, so
		//    Go just unions them here. Same prefix appearing twice is OK
		//    — the placeholder rule registry deduplicates by full rule
		//    name, and the runner uses ConfigKey, not prefix, for dispatch.
		if len(entry.EslintPlugins) > 0 {
			merged.EslintPlugins = append(merged.EslintPlugins, entry.EslintPlugins...)
		}
	}

	// No entry matched this file — do not lint it
	if !entryMatched {
		return nil
	}

	return merged
}

// isGlobalIgnoreEntry returns true if the entry is a global ignore entry
// (has only ignores, no other fields).
func isGlobalIgnoreEntry(entry ConfigEntry) bool {
	return len(entry.Files) == 0 &&
		len(entry.Rules) == 0 &&
		len(entry.Plugins) == 0 &&
		len(entry.EslintPlugins) == 0 &&
		entry.Settings == nil &&
		entry.LanguageOptions == nil &&
		len(entry.Ignores) > 0
}

// isFileMatched checks if a file matches any of the given glob patterns
func isFileMatched(filePath string, patterns []string, cwd string) bool {
	var normalizedPath string
	if cwd != "" {
		normalizedPath = normalizePath(filePath, cwd)
	} else {
		normalizedPath = filePath
	}

	for _, pattern := range patterns {
		normalizedPattern := normalizePattern(pattern)

		if matched, err := doublestar.Match(normalizedPattern, normalizedPath); err == nil && matched {
			return true
		}
		if normalizedPath != filePath {
			if matched, err := doublestar.Match(normalizedPattern, filePath); err == nil && matched {
				return true
			}
		}
		unixPath := strings.ReplaceAll(normalizedPath, "\\", "/")
		if unixPath != normalizedPath {
			if matched, err := doublestar.Match(normalizedPattern, unixPath); err == nil && matched {
				return true
			}
		}
	}
	return false
}

// mergeLanguageOptions merges two LanguageOptions, with `override`
// winning on conflict. The merge has two halves:
//
//   - Typed native fields (`Project`, `ProjectService`): replaced
//     per-field. `override`'s non-empty value wins; empty
//     value leaves base intact (matches v9 flat-config semantics where
//     later entries only override what they explicitly declare).
//   - Opaque compat blob (`Compat`): generic JSON deep-merge —
//     objects merge their keys recursively, scalars / arrays / type
//     mismatches use later-wins replace. This matches ESLint's own
//     `languageOptions` merge for every field family currently in the
//     flat-config spec (`globals`, `parserOptions.ecmaVersion`,
//     `parserOptions.ecmaFeatures.*`, etc.) without Go having to know
//     each field's individual semantics.
//
// The Compat side is the load-bearing piece: new ESLint compat fields
// (parser, allowImportExportEverywhere, ...) flow through without ANY
// Go change because they land in the same opaque map.
func mergeLanguageOptions(base, override *LanguageOptions) *LanguageOptions {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}
	merged := &LanguageOptions{
		ParserOptions: mergeParserOptions(base.ParserOptions, override.ParserOptions),
		Compat:        deepMergeMap(base.Compat, override.Compat),
	}
	return merged
}

// mergeParserOptions does the same typed+opaque split for the nested
// `parserOptions` block. Typed fields (`Project`, `ProjectService`)
// are per-field overrides; everything else flows through `Compat` via
// generic deep-merge.
func mergeParserOptions(base, override *ParserOptions) *ParserOptions {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}
	merged := &ParserOptions{
		ProjectService: base.ProjectService,
		Project:        base.Project,
		Compat:         deepMergeMap(base.Compat, override.Compat),
	}
	if override.ProjectService != nil {
		merged.ProjectService = override.ProjectService
	}
	if len(override.Project) > 0 {
		merged.Project = override.Project
	}
	return merged
}

// deepMergeMap is a generic JSON-shaped deep-merge: object values at
// the same path get merged recursively; everything else (scalars,
// arrays, type mismatches) uses later-wins replace. The input maps are
// never mutated — the result is a freshly allocated copy.
//
// This is the canonical merge for opaque ESLint-compat config blobs.
// The semantics happen to match ESLint v9's own
// `lib/config/flat-config-helpers.js` mergeOption logic for every
// `languageOptions.*` field family in the current flat-config spec:
//
//   - `globals` (map) — object keys merge, later-wins per-key.
//   - `parserOptions.ecmaVersion` (scalar) — same key, later-wins.
//   - `parserOptions.sourceType` (scalar) — same.
//   - `parserOptions.ecmaFeatures` (map) — keys merge, later-wins.
//   - `parserOptions.parser` (object|string) — later-wins replace.
//
// Future ESLint additions that introduce array semantics here would
// need special-cased rules; we'd carve a typed exception OR replicate
// the field via TS-side merge before sending to Go. For now, the spec
// has no such case.
func deepMergeMap(base, override map[string]any) map[string]any {
	if override == nil {
		return cloneAnyMap(base)
	}
	if base == nil {
		return cloneAnyMap(override)
	}
	out := make(map[string]any, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, ov := range override {
		if bv, exists := out[k]; exists {
			if bm, ok1 := bv.(map[string]any); ok1 {
				if om, ok2 := ov.(map[string]any); ok2 {
					out[k] = deepMergeMap(bm, om)
					continue
				}
			}
		}
		out[k] = ov
	}
	return out
}

// cloneAnyMap returns a shallow copy of m (never mutating the input).
// Nested objects keep their original references — deepMergeMap reads
// them but never mutates, so sharing is safe.
func cloneAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// RulePluginPrefix extracts the plugin prefix from a rule name.
// "@typescript-eslint/no-explicit-any" → "@typescript-eslint"
// "import/no-unresolved" → "import"
// "no-debugger" → "" (core rule)
func RulePluginPrefix(ruleName string) string {
	lastSlash := strings.LastIndex(ruleName, "/")
	if lastSlash < 0 {
		return ""
	}
	return ruleName[:lastSlash]
}

// GetCoreRules returns core ESLint rules (those without a "/" prefix in their registered name).
func GetCoreRules() []rule.Rule {
	return coreRules.GetAllRules()
}

// InitDefaultConfig, createDefaultConfig, migrateJSONConfig and related helpers
// are in config_init.go.
