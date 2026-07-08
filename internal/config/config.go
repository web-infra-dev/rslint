package config

import (
	"encoding/json"
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
	typescriptPlugin "github.com/web-infra-dev/rslint/internal/plugins/typescript"
	unicornPlugin "github.com/web-infra-dev/rslint/internal/plugins/unicorn"
	"github.com/web-infra-dev/rslint/internal/rule"
	coreRules "github.com/web-infra-dev/rslint/internal/rules"
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
}

// Settings represents shared settings accessible to rules
type Settings map[string]interface{}

// LanguageOptions contains language-specific configuration options.
type LanguageOptions struct {
	ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
	// Raw retains the full languageOptions object as authored (sourceType,
	// globals, parserOptions.ecmaFeatures, …) — fields the Go core does not
	// model but the Node eslint-plugin worker needs. Go computes the
	// per-file merged value via GetConfigForFile and forwards it on the
	// wire; it is not (de)serialized through this struct's own field tags.
	Raw map[string]any `json:"-"`
}

// UnmarshalJSON captures both the typed ParserOptions and the full raw
// object (the latter for forwarding to the eslint-plugin worker).
func (lo *LanguageOptions) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	type parserShape struct {
		ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
	}
	var ps parserShape
	if err := json.Unmarshal(data, &ps); err != nil {
		return err
	}
	lo.ParserOptions = ps.ParserOptions
	lo.Raw = raw
	return nil
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
// ProjectService uses *bool to distinguish "not set" (nil) from "explicitly false".
type ParserOptions struct {
	ProjectService *bool        `json:"projectService,omitempty"`
	Project        ProjectPaths `json:"project,omitempty"`
	// JsxPragma / JsxFragmentName mirror @typescript-eslint/parser's
	// `parserOptions.jsxPragma` / `parserOptions.jsxFragmentName` — the
	// identifier implicitly referenced by JSX elements / fragments (Preact's
	// `h`/`Fragment`, a custom factory, ...).
	JsxPragma       *string `json:"jsxPragma,omitempty"`
	JsxFragmentName *string `json:"jsxFragmentName,omitempty"`
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
		remaining := ruleArray[1:]
		if len(remaining) == 1 {
			if _, isArray := remaining[0].([]interface{}); isArray {
				// A lone option that is itself an array (e.g. ["error",
				// ["a","b"]]): keep the outer wrapper so it stays distinguishable
				// from a multi-element option list and maps to context.options ==
				// [["a","b"]]. Unwrapping would collapse it to ["a","b"],
				// indistinguishable from ["error","a","b"] — and the eslint-plugin
				// dispatch would then drop a nesting level.
				ruleConfig.Options = remaining
			} else {
				// Single non-array option: pass the value directly (string, map).
				ruleConfig.Options = remaining[0]
			}
		} else {
			// Multiple option elements: pass as array (e.g. ["both", {blockScopedFunctions: "disallow"}])
			ruleConfig.Options = remaining
		}
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

// normalizePattern cleans up a glob pattern to match paths produced by normalizePath.
// normalizePath uses tspath.NormalizePath on file paths (strips leading "./", collapses
// "/./", resolves ".."), so patterns must undergo the same transformation.
// matchGlob matches a glob pattern against a path using doublestar.
func matchGlob(pattern, path string) bool {
	m, err := doublestar.Match(pattern, path)
	return err == nil && m
}

func normalizePattern(pattern string) string {
	return tspath.NormalizePath(pattern)
}

// isDirBlockedByIgnores checks if the file's directory is blocked by a
// directory-level ignore pattern (e.g., `dir/**`). File-level patterns and
// negation patterns are excluded (by Kind) in isDirAbsolutelyBlocked. This
// aligns with ESLint v10: `dir/**` blocks directory traversal entirely, and
// `!` negation cannot undo it.
func isDirBlockedByIgnores(filePath string, patterns []IgnorePattern, cwd string) bool {
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
	return isDirAbsolutelyBlocked(dirPath, patterns)
}

// normalizePath converts file path to be relative to cwd for consistent matching
func normalizePath(filePath, cwd string) string {
	return tspath.NormalizePath(tspath.ConvertToRelativePath(filePath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: true,
		CurrentDirectory:          cwd,
	}))
}

// MergedConfig is the final computed configuration for a single file
type MergedConfig struct {
	Rules           map[string]*RuleConfig
	Settings        Settings
	LanguageOptions *LanguageOptions
	Plugins         map[string]struct{}
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

		// 3. Entry-level ignores. Parsed per entry; entry.Ignores is usually
		// empty (ESLint configs put ignores in a dedicated global-ignore entry),
		// so ParseIgnorePatterns returns nil and this is free in the common case.
		if isFileIgnored(filePath, ParseIgnorePatterns(entry.Ignores), cwd) {
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

// mergeLanguageOptions deep-merges two LanguageOptions, with override taking precedence
func mergeLanguageOptions(base, override *LanguageOptions) *LanguageOptions {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}
	merged := *base
	if override.ParserOptions != nil {
		if merged.ParserOptions == nil {
			merged.ParserOptions = override.ParserOptions
		} else {
			po := *merged.ParserOptions
			if override.ParserOptions.ProjectService != nil {
				po.ProjectService = override.ParserOptions.ProjectService
			}
			if len(override.ParserOptions.Project) > 0 {
				po.Project = override.ParserOptions.Project
			}
			if override.ParserOptions.JsxPragma != nil {
				po.JsxPragma = override.ParserOptions.JsxPragma
			}
			if override.ParserOptions.JsxFragmentName != nil {
				po.JsxFragmentName = override.ParserOptions.JsxFragmentName
			}
			merged.ParserOptions = &po
		}
	}
	// Deep-merge the raw languageOptions map (override wins per key; nested
	// objects like `parserOptions` are merged key-by-key rather than replaced
	// wholesale). merged is a shallow copy of base, so build a fresh map
	// rather than mutating base.Raw.
	if len(override.Raw) > 0 {
		merged.Raw = deepMergeRawMaps(base.Raw, override.Raw)
	}
	return &merged
}

// deepMergeRawMaps recursively merges override into base (override wins on
// conflicting scalar/array keys); when both sides have a `map[string]any`
// at the same key, they are merged recursively instead of override
// replacing the whole nested object. Matches ESLint flat-config's
// languageOptions merge semantics closely enough for parserOptions/globals
// nesting without pulling in a general-purpose deep-merge dependency.
func deepMergeRawMaps(base, override map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		if baseVal, ok := merged[k]; ok {
			if baseMap, ok := baseVal.(map[string]any); ok {
				if overrideMap, ok := v.(map[string]any); ok {
					merged[k] = deepMergeRawMaps(baseMap, overrideMap)
					continue
				}
			}
		}
		merged[k] = v
	}
	return merged
}

// ResolveJsxPragmaOptions scans entries for a `languageOptions.parserOptions`
// jsxPragma / jsxFragmentName, in entry order, later entries overriding
// earlier ones (matching GetConfigForFile's merge order). It exists for the
// Programs that synthesize CompilerOptions for a batch of files instead of
// resolving a single file's merged config — the no-tsconfig directory scan
// and the gap-file fallback Program (cmd/rslint/programs.go).
func ResolveJsxPragmaOptions(entries RslintConfig) (jsxFactory, jsxFragmentFactory string) {
	for _, entry := range entries {
		if entry.LanguageOptions == nil || entry.LanguageOptions.ParserOptions == nil {
			continue
		}
		po := entry.LanguageOptions.ParserOptions
		if po.JsxPragma != nil && *po.JsxPragma != "" {
			jsxFactory = *po.JsxPragma
		}
		if po.JsxFragmentName != nil && *po.JsxFragmentName != "" {
			jsxFragmentFactory = *po.JsxFragmentName
		}
	}
	return jsxFactory, jsxFragmentFactory
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
