package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	importPlugin "github.com/web-infra-dev/rslint/internal/plugins/import"
	jestPlugin "github.com/web-infra-dev/rslint/internal/plugins/jest"
	jsxA11yPlugin "github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y"
	promisePlugin "github.com/web-infra-dev/rslint/internal/plugins/promise"
	reactPlugin "github.com/web-infra-dev/rslint/internal/plugins/react"
	reactHooksPlugin "github.com/web-infra-dev/rslint/internal/plugins/react_hooks"
	rstestPlugin "github.com/web-infra-dev/rslint/internal/plugins/rstest"
	typescriptPlugin "github.com/web-infra-dev/rslint/internal/plugins/typescript"
	unicornPlugin "github.com/web-infra-dev/rslint/internal/plugins/unicorn"
	"github.com/web-infra-dev/rslint/internal/rule"
	coreRules "github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// RslintConfig represents the top-level configuration array
type RslintConfig []ConfigEntry

// ConfigEntry represents a single configuration entry in the config array
type ConfigEntry struct {
	Name string `json:"name,omitempty"`
	// Files retains the established Go construction API for top-level OR
	// patterns. FilePatternGroups stores nested arrays, each of which is an AND
	// group. JSON encoding combines both fields back into one mixed `files`
	// array.
	Files             []string         `json:"files,omitempty"`
	FilePatternGroups [][]string       `json:"-"`
	Ignores           []string         `json:"ignores,omitempty"`
	LanguageOptions   *LanguageOptions `json:"languageOptions,omitempty"`
	// Omit an absent rules map when marshaling. Emitting an invented
	// `"rules": null` changes flat-config object-shape semantics when the JSON
	// is decoded again: an ignores-only global entry becomes an entry-local
	// ignore. Authored `rules: null` is still preserved as non-global on decode.
	Rules    Rules    `json:"rules,omitempty"`
	Plugins  []string `json:"plugins,omitempty"`
	Settings Settings `json:"settings,omitempty"`

	// collectedGitignore marks the process-local synthetic entry prepended by
	// ConfigWithGitignore and retains its once-compiled directory-node
	// matchers. Keeping the uncommon metadata behind one pointer preserves the
	// original ConfigEntry footprint for every ordinary authored config entry.
	collectedGitignore *collectedGitignoreMetadata
}

type collectedGitignoreMetadata struct {
	ignores []IgnorePattern
}

func (entry ConfigEntry) MarshalJSON() ([]byte, error) {
	type configEntryAlias ConfigEntry
	encoded, err := json.Marshal(configEntryAlias(entry))
	if err != nil {
		return nil, err
	}
	if len(entry.FilePatternGroups) == 0 && entry.Rules == nil && entry.Plugins == nil && entry.Settings == nil {
		return encoded, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		return nil, err
	}
	if entry.Rules != nil {
		rulesJSON, err := json.Marshal(entry.Rules)
		if err != nil {
			return nil, err
		}
		object["rules"] = rulesJSON
	}
	// Empty maps/slices are meaningful at the flat-config boundary: their key
	// presence keeps an ignores-bearing object entry-local. encoding/json's
	// omitempty drops them, so restore the authored shape explicitly. Settings{}
	// also serves as the decoder's sentinel for a currently unsupported key or
	// an authored null field whose presence made the object non-global.
	if entry.Plugins != nil {
		pluginsJSON, err := json.Marshal(entry.Plugins)
		if err != nil {
			return nil, err
		}
		object["plugins"] = pluginsJSON
	}
	if entry.Settings != nil {
		settingsJSON, err := json.Marshal(entry.Settings)
		if err != nil {
			return nil, err
		}
		object["settings"] = settingsJSON
	}
	if len(entry.FilePatternGroups) > 0 {
		selectors := make([]any, 0, len(entry.Files)+len(entry.FilePatternGroups))
		for _, pattern := range entry.Files {
			selectors = append(selectors, pattern)
		}
		for _, group := range entry.FilePatternGroups {
			selectors = append(selectors, group)
		}
		filesJSON, err := json.Marshal(selectors)
		if err != nil {
			return nil, err
		}
		object["files"] = filesJSON
	}
	return json.Marshal(object)
}

func (entry *ConfigEntry) UnmarshalJSON(data []byte) error {
	wrapped := make([]byte, 0, len(data)+2)
	wrapped = append(wrapped, '[')
	wrapped = append(wrapped, data...)
	wrapped = append(wrapped, ']')
	var config RslintConfig
	if err := config.UnmarshalJSON(wrapped); err != nil {
		return err
	}
	if len(config) != 1 {
		return errors.New("expected one config entry")
	}
	*entry = config[0]
	return nil
}

// Settings represents shared settings accessible to rules
type Settings map[string]interface{}

// UnmarshalJSON rejects invalid explicit `files` values at the config boundary.
// An omitted `files` field remains valid and means "use rslint's default
// lintable extensions"; explicit null/empty arrays are invalid.
func (config *RslintConfig) UnmarshalJSON(data []byte) error {
	if strings.TrimSpace(string(data)) == "null" {
		*config = nil
		return nil
	}

	var rawEntries []json.RawMessage
	if err := json.Unmarshal(data, &rawEntries); err != nil {
		return err
	}

	type configEntryAlias ConfigEntry
	entries := make(RslintConfig, 0, len(rawEntries))
	for index, rawEntry := range rawEntries {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rawEntry, &raw); err != nil {
			return err
		}

		var decoded configEntryAlias
		entryForDecode := rawEntry
		if rawFiles, ok := raw["files"]; ok {
			files, groups, err := decodeFilesSelectors(rawFiles, index)
			if err != nil {
				return err
			}
			decoded.Files = files
			decoded.FilePatternGroups = groups
			withoutFiles := make(map[string]json.RawMessage, len(raw)-1)
			for key, value := range raw {
				if key != "files" {
					withoutFiles[key] = value
				}
			}
			entryForDecode, err = json.Marshal(withoutFiles)
			if err != nil {
				return err
			}
		}
		if err := json.Unmarshal(entryForDecode, &decoded); err != nil {
			return err
		}
		if err := validateConfigRules(decoded.Rules); err != nil {
			return fmt.Errorf("config entry at index %d: %w", index, err)
		}
		// Global-ignore semantics depend on object shape, not on whether a
		// present field decodes to a non-nil Go value. Preserve the non-global
		// shape of entries such as {ignores, rules: null} or entries carrying a
		// currently unsupported field. An empty Settings map is behaviorally
		// neutral but remains non-nil for isGlobalIgnoreEntry.
		hasNonGlobalKey := false
		for key := range raw {
			if key != "ignores" && key != "name" {
				hasNonGlobalKey = true
				break
			}
		}
		if hasNonGlobalKey && decoded.Files == nil && decoded.FilePatternGroups == nil && decoded.Rules == nil && decoded.Plugins == nil && decoded.Settings == nil && decoded.LanguageOptions == nil {
			decoded.Settings = Settings{}
		}
		entries = append(entries, ConfigEntry(decoded))
	}

	*config = entries
	return nil
}

func decodeFilesSelectors(raw json.RawMessage, entryIndex int) ([]string, [][]string, error) {
	var selectors []json.RawMessage
	if err := json.Unmarshal(raw, &selectors); err != nil || len(selectors) == 0 {
		if err != nil {
			return nil, nil, fmt.Errorf("config entry at index %d: key \"files\": expected value to be a non-empty array: %w", entryIndex, err)
		}
		return nil, nil, fmt.Errorf("config entry at index %d: key \"files\": expected value to be a non-empty array", entryIndex)
	}

	files := make([]string, 0, len(selectors))
	var groups [][]string
	for selectorIndex, selector := range selectors {
		var value any
		if err := json.Unmarshal(selector, &value); err != nil {
			return nil, nil, err
		}
		switch value := value.(type) {
		case string:
			files = append(files, value)
		case []any:
			group := make([]string, 0, len(value))
			for _, item := range value {
				pattern, ok := item.(string)
				if !ok {
					return nil, nil, fmt.Errorf(
						"config entry at index %d: key \"files\": item at index %d must be a string or an array of strings",
						entryIndex,
						selectorIndex,
					)
				}
				group = append(group, pattern)
			}
			groups = append(groups, group)
		default:
			return nil, nil, fmt.Errorf(
				"config entry at index %d: key \"files\": item at index %d must be a string or an array of strings",
				entryIndex,
				selectorIndex,
			)
		}
	}
	return files, groups, nil
}

// ValidateConfig checks config invariants for configs constructed in Go. JSON
// config ingress rejects explicit null/empty `files` during unmarshaling.
func ValidateConfig(config RslintConfig) error {
	for index, entry := range config {
		if (entry.Files != nil || entry.FilePatternGroups != nil) && len(entry.Files) == 0 && len(entry.FilePatternGroups) == 0 {
			return fmt.Errorf("config entry at index %d: key \"files\": expected value to be a non-empty array", index)
		}
		if err := validateLanguageOptions(entry.LanguageOptions); err != nil {
			return fmt.Errorf("config entry at index %d: %w", index, err)
		}
		if err := validateConfigRules(entry.Rules); err != nil {
			return fmt.Errorf("config entry at index %d: %w", index, err)
		}
	}
	return nil
}

func validateLanguageOptions(languageOptions *LanguageOptions) error {
	if err := validateConfigGlobals(languageOptions); err != nil {
		return err
	}
	if languageOptions == nil || languageOptions.Raw == nil {
		return nil
	}

	if value, present := languageOptions.Raw["ecmaVersion"]; present {
		if _, ok := normalizeConfigECMAVersion(value); !ok {
			return fmt.Errorf(
				"key \"languageOptions.ecmaVersion\": invalid value %v; expected \"latest\", 3, 5, an edition from 6 through %d, or a year from 2015 through %d",
				value,
				rule.LatestECMAScriptVersion-2009,
				rule.LatestECMAScriptVersion,
			)
		}
	}
	if value, present := languageOptions.Raw["sourceType"]; present {
		sourceType, ok := value.(string)
		if !ok || (sourceType != "module" && sourceType != "script" && sourceType != "commonjs") {
			return fmt.Errorf(
				"key \"languageOptions.sourceType\": invalid value %v; expected \"module\", \"script\", or \"commonjs\"",
				value,
			)
		}
	}

	return nil
}

func normalizeConfigECMAVersion(value any) (int, bool) {
	if text, ok := value.(string); ok {
		if text == "latest" {
			// Keep `latest` semantic instead of freezing the config to this
			// release's numeric edition. rule.LanguageOptions uses zero for latest.
			return 0, true
		}
		return 0, false
	}
	version, ok := utils.CoerceIntegral(value)
	if !ok {
		return 0, false
	}
	return rule.NormalizeECMAScriptVersion(version)
}

func validateConfigGlobals(languageOptions *LanguageOptions) error {
	if languageOptions == nil || languageOptions.Raw == nil {
		return nil
	}
	value, present := languageOptions.Raw["globals"]
	if !present {
		return nil
	}
	globals, ok := value.(map[string]any)
	if !ok {
		return errors.New("key \"languageOptions.globals\": expected an object")
	}
	for name, access := range globals {
		if name != strings.TrimSpace(name) {
			return fmt.Errorf("key \"languageOptions.globals\": global %q has leading or trailing whitespace", name)
		}
		if _, valid := utils.NormalizeGlobalAccess(access); !valid {
			return fmt.Errorf(
				"key \"languageOptions.globals\": global %q has invalid access %v; expected a boolean, null, \"true\", \"false\", \"readonly\", \"readable\", \"writable\", \"writeable\", or \"off\"",
				name,
				access,
			)
		}
	}
	return nil
}

func validateConfigRules(rules Rules) error {
	for name, value := range rules {
		if _, _, err := parseRuleConfigValue(value); err != nil {
			return fmt.Errorf("key \"rules\": rule %q: %w", name, err)
		}
	}
	return nil
}

// LanguageOptions contains language-specific configuration options.
type LanguageOptions struct {
	ParserOptions *ParserOptions `json:"parserOptions,omitempty"`
	// Raw retains the full languageOptions object as authored (sourceType,
	// ecmaVersion, globals, parserOptions.ecmaFeatures, …). Go normalizes
	// ecmaVersion for native globals and forwards the full value to the Node
	// eslint-plugin worker. It is not (de)serialized through this struct's own
	// field tags.
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

// MarshalJSON preserves the open languageOptions object captured in Raw while
// folding any typed ParserOptions updates back into its parserOptions member.
// Without this counterpart to UnmarshalJSON, a config JSON round trip would
// silently discard ecmaVersion, globals, sourceType, and unknown parser keys.
func (lo LanguageOptions) MarshalJSON() ([]byte, error) {
	raw := deepMergeConfigObjects(nil, lo.Raw)
	if lo.ParserOptions != nil {
		encodedParserOptions, err := json.Marshal(lo.ParserOptions)
		if err != nil {
			return nil, err
		}
		var typedParserOptions map[string]any
		if err := json.Unmarshal(encodedParserOptions, &typedParserOptions); err != nil {
			return nil, err
		}
		baseParserOptions, _ := configObject(raw["parserOptions"])
		raw["parserOptions"] = deepMergeConfigObjects(baseParserOptions, typedParserOptions)
	}
	return json.Marshal(raw)
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
	Level   string        `json:"level,omitempty"`   // "error", "warn", "off"
	Options []interface{} `json:"options,omitempty"` // ESLint's context.options array (post-severity elements)
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
func (rc *RuleConfig) GetOptions() []interface{} {
	if rc == nil {
		return nil
	}
	return rc.Options
}

// SetOptions sets the rule options
func (rc *RuleConfig) SetOptions(options []interface{}) {
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
		RulePrefix:  "rstest",
		DeclNames:   []string{"rstest"},
		getAllRules: func() []rule.Rule { return rstestPlugin.GetAllRules() },
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
// - [0] -> disabled rule
// - ["error"] -> enabled rule with error severity
// - [2] -> enabled rule with error severity
// - ["warn"] -> enabled rule with warning severity
// - [1] -> enabled rule with warning severity
// - ["error", {...options}] -> enabled rule with error severity and options
// - ["error", "both"] -> enabled rule with string option (e.g. no-inner-declarations)
// - ["error", "both", {...options}] -> enabled rule with string + object options
func parseArrayRuleConfig(ruleArray []interface{}) *RuleConfig {
	ruleConfig, _, err := parseRuleConfigValue(ruleArray)
	if err != nil {
		return nil
	}
	return ruleConfig
}

func parseRuleConfigValue(value any) (*RuleConfig, bool, error) {
	if ruleArray, ok := value.([]interface{}); ok {
		if len(ruleArray) == 0 {
			return nil, false, errors.New("rule configuration array must contain a severity")
		}
		level, err := normalizeRuleSeverity(ruleArray[0])
		if err != nil {
			return nil, false, fmt.Errorf("invalid array severity at index 0: %w", err)
		}
		ruleConfig := &RuleConfig{Level: level}
		if len(ruleArray) > 1 {
			// Keep ESLint's context.options array form without collapsing a
			// single positional option to a bare value.
			ruleConfig.Options = ruleArray[1:]
		}
		return ruleConfig, len(ruleArray) > 1, nil
	}

	level, err := normalizeRuleSeverity(value)
	if err != nil {
		return nil, false, err
	}
	return &RuleConfig{Level: level}, false, nil
}

func normalizeRuleSeverity(value any) (string, error) {
	if value == nil {
		return "", invalidRuleSeverity(value)
	}
	if number, ok := value.(json.Number); ok {
		numeric, err := number.Float64()
		if err != nil {
			return "", invalidRuleSeverity(value)
		}
		return normalizeNumericRuleSeverity(numeric, value)
	}

	severity := reflect.ValueOf(value)
	switch severity.Kind() {
	case reflect.String:
		level := severity.String()
		switch level {
		case "off", "warn", "error":
			return level, nil
		default:
			return "", invalidRuleSeverity(value)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return normalizeNumericRuleSeverity(float64(severity.Int()), value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return normalizeNumericRuleSeverity(float64(severity.Uint()), value)
	case reflect.Float32, reflect.Float64:
		return normalizeNumericRuleSeverity(severity.Float(), value)
	}

	return "", invalidRuleSeverity(value)
}

func normalizeNumericRuleSeverity(value float64, original any) (string, error) {
	switch value {
	case 0:
		return "off", nil
	case 1:
		return "warn", nil
	case 2:
		return "error", nil
	default:
		return "", invalidRuleSeverity(original)
	}
}

func invalidRuleSeverity(value any) error {
	return fmt.Errorf(
		"invalid severity %v (%T); expected \"off\", \"warn\", \"error\", 0, 1, or 2",
		value,
		value,
	)
}

var registerOnce sync.Once

func RegisterAllRules() {
	registerOnce.Do(func() {
		registerAllTypeScriptEslintPluginRules()
		registerAllImportPluginRules()
		registerAllReactPluginRules()
		registerAllReactHooksPluginRules()
		registerAllJestPluginRules()
		registerAllRstestPluginRules()
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

func registerAllRstestPluginRules() {
	for _, rule := range rstestPlugin.GetAllRules() {
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
func normalizePattern(pattern string) string {
	return tspath.NormalizePath(pattern)
}

// isDirBlockedByIgnores checks if the file's directory is blocked by a
// directory-level ignore pattern (e.g., `dir/**`). File-level patterns and
// negation patterns are excluded (by Kind) in isDirAbsolutelyBlocked. This
// aligns with ESLint v10: `dir/**` blocks directory traversal entirely, and
// `!` negation cannot undo it.
func isDirBlockedByIgnores(filePath string, patterns []IgnorePattern, cwd string) bool {
	dirPath, ok := ignoreDirectoryPath(filePath, cwd)
	return ok && isDirAbsolutelyBlocked(dirPath, patterns)
}

func ignoreDirectoryPath(filePath string, cwd string) (string, bool) {
	var dirPath string
	if cwd != "" {
		dirPath = normalizePath(tspath.GetDirectoryPath(filePath), cwd)
	} else {
		dirPath = tspath.GetDirectoryPath(filePath)
	}
	dirPath = strings.ReplaceAll(dirPath, "\\", "/")
	dirPath = strings.TrimSuffix(dirPath, "/")
	if dirPath == "" || dirPath == "." {
		return "", false
	}
	return dirPath, true
}

// directoryBlockMatcher caches path-only directory decisions for one immutable
// ignore-pattern set and cwd. The exact lexical parent is the key: no case,
// separator, realpath, or filesystem equivalence is introduced by the cache.
type directoryBlockMatcher struct {
	patterns []IgnorePattern
	cwd      string
	results  publishOnceCache[string, bool]
}

func newDirectoryBlockMatcher(patterns []IgnorePattern, cwd string) *directoryBlockMatcher {
	if len(patterns) == 0 {
		return nil
	}
	return &directoryBlockMatcher{patterns: patterns, cwd: cwd}
}

func (matcher *directoryBlockMatcher) blocksFileDirectory(filePath string) bool {
	if matcher == nil {
		return false
	}
	lexicalDirectory := tspath.GetDirectoryPath(filePath)
	return matcher.results.getOrInit(lexicalDirectory, func() bool {
		dirPath, ok := ignoreDirectoryPath(filePath, matcher.cwd)
		return ok && isDirAbsolutelyBlocked(dirPath, matcher.patterns)
	})
}

// normalizePath converts file path to be relative to cwd for consistent matching
func normalizePath(filePath, cwd string) string {
	return normalizePathWithCaseSensitivity(filePath, cwd, true)
}

func normalizePathWithCaseSensitivity(filePath, cwd string, useCaseSensitive bool) string {
	return tspath.NormalizePath(tspath.ConvertToRelativePath(filePath, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: useCaseSensitive,
		CurrentDirectory:          cwd,
	}))
}

// fileMatchPath keeps the path representations shared by config-entry
// selectors and file-level ignores for one resolution. It initializes lazily
// so directory-blocked files and entries without path matchers do no extra
// normalization work.
type fileMatchPath struct {
	original   string
	cwd        string
	normalized string
	unix       string
	ready      bool
}

func newFileMatchPath(filePath string, cwd string) fileMatchPath {
	return fileMatchPath{original: filePath, cwd: cwd}
}

func (path *fileMatchPath) normalizedPaths() (string, string) {
	if !path.ready {
		path.normalized = path.original
		if path.cwd != "" {
			path.normalized = normalizePath(path.original, path.cwd)
		}
		path.unix = strings.ReplaceAll(path.normalized, "\\", "/")
		path.ready = true
	}
	return path.normalized, path.unix
}

// MergedConfig is the final computed configuration for a single file
type MergedConfig struct {
	Rules           map[string]*RuleConfig
	Settings        Settings
	LanguageOptions *LanguageOptions
	Plugins         map[string]struct{}
}

func extractConfigIgnores(config RslintConfig) []IgnorePattern {
	var ignores []IgnorePattern
	for _, entry := range config {
		if !isGlobalIgnoreEntry(entry) {
			continue
		}
		var entryIgnores []IgnorePattern
		if entry.collectedGitignore != nil {
			entryIgnores = entry.collectedGitignore.ignores
		} else {
			entryIgnores = ParseIgnorePatterns(entry.Ignores)
		}
		if len(entryIgnores) == 0 {
			continue
		}
		if len(ignores) == 0 {
			// Reuse the first immutable group. Limiting capacity guarantees a
			// later append cannot mutate its config-owned backing array.
			ignores = entryIgnores[:len(entryIgnores):len(entryIgnores)]
		} else {
			ignores = append(ignores, entryIgnores...)
		}
	}
	return ignores
}

// IsFileIgnored reports whether filePath is excluded by the config's global
// `ignores` patterns. It is distinct from GetConfigForFile returning nil,
// which also covers "no entry matched this file" — callers that need ESLint's
// "ignores removes the file from lint target discovery" semantics should use
// this method. Program-wide type-check diagnostics are intentionally governed
// by tsconfig membership instead.
func (config RslintConfig) IsFileIgnored(filePath string, cwd string) bool {
	patterns := extractConfigIgnores(config)
	if len(patterns) == 0 {
		return false
	}
	return isDirBlockedByIgnores(filePath, patterns, cwd) ||
		isFileIgnored(filePath, patterns, cwd)
}

// GetConfigForFile computes the merged configuration for a selected file
// following ESLint flat config semantics. It returns nil when the file is
// globally ignored, outside the config's implicit/explicit selector union, or
// no entry contributes configuration.
//
// Global ignore evaluation happens in two phases:
//  1. Directory-level (isDirBlockedByIgnores): patterns like dir/** block entire directories.
//     Negation (!) cannot override directory-level blocking.
//  2. File-level (isFileIgnored): sequential evaluation with ! negation support for re-inclusion.
//
// After global ignore check, entries are merged in order if their files match and ignores don't.
// cwd is the directory the config lives in; file paths are resolved relative to it.
func (config RslintConfig) GetConfigForFile(filePath string, cwd string) *MergedConfig {
	// Collect all global ignore patterns and evaluate once. This allows `!`
	// negation patterns in separate entries to work correctly, aligned with
	// ESLint v10 which merges all global ignores before evaluating. Callers
	// resolving many files against the same config (FileConfigResolver)
	// should precompute this once via extractConfigIgnores and call
	// getConfigForFileWithIgnores directly instead of re-parsing every
	// pattern string on every call.
	return config.getConfigForFileWithIgnores(filePath, cwd, extractConfigIgnores(config))
}

// getConfigForFileWithIgnores is GetConfigForFile with the global ignore
// patterns supplied by the caller, so repeated calls against the same config
// (one per lint target) don't re-parse the same ignore pattern strings.
func (config RslintConfig) getConfigForFileWithIgnores(filePath string, cwd string, globalIgnorePatterns []IgnorePattern) *MergedConfig {
	key, matched := config.matchConfigEntries(filePath, cwd, globalIgnorePatterns, nil, nil)
	if !matched {
		return nil
	}
	return config.mergeConfigEntries(key)
}

// isGlobalIgnoreEntry returns true if the entry has only ignores and an
// optional name. Empty config objects are still present and make ignores local
// to the entry, matching ESLint flat config semantics.
func isGlobalIgnoreEntry(entry ConfigEntry) bool {
	return entry.Files == nil &&
		entry.FilePatternGroups == nil &&
		entry.Rules == nil &&
		entry.Plugins == nil &&
		entry.Settings == nil &&
		entry.LanguageOptions == nil &&
		len(entry.Ignores) > 0
}

func hasFileSelectors(entry ConfigEntry) bool {
	return len(entry.Files) > 0 || len(entry.FilePatternGroups) > 0
}

func isFileMatchedByConfigEntry(filePath string, entry ConfigEntry, cwd string) bool {
	matchPath := newFileMatchPath(filePath, cwd)
	return matchPath.matchesConfigEntry(entry)
}

func (path *fileMatchPath) matchesConfigEntry(entry ConfigEntry) bool {
	if path.matchesAny(entry.Files) {
		return true
	}
	for _, group := range entry.FilePatternGroups {
		// ESLint treats an empty nested selector as a vacuously true AND group.
		matched := true
		for _, pattern := range group {
			if !path.matchesSingle(pattern) {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func deepMergeConfigObjects(base map[string]any, override map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(override))
	for key, value := range base {
		merged[key] = cloneConfigValue(value)
	}
	for key, value := range override {
		baseObject, baseIsObject := configObject(base[key])
		overrideObject, overrideIsObject := configObject(value)
		if baseIsObject && overrideIsObject {
			merged[key] = deepMergeConfigObjects(baseObject, overrideObject)
			continue
		}
		merged[key] = cloneConfigValue(value)
	}
	return merged
}

func configObject(value any) (map[string]any, bool) {
	switch object := value.(type) {
	case map[string]any:
		return object, true
	case Settings:
		return map[string]any(object), true
	default:
		return nil, false
	}
}

func cloneConfigValue(value any) any {
	if object, ok := configObject(value); ok {
		return deepMergeConfigObjects(nil, object)
	}
	switch values := value.(type) {
	case []any:
		cloned := make([]any, len(values))
		for index, item := range values {
			cloned[index] = cloneConfigValue(item)
		}
		return cloned
	case []string:
		return append([]string(nil), values...)
	default:
		return value
	}
}

// isFileMatched checks if a file matches any of the given glob patterns
func isFileMatched(filePath string, patterns []string, cwd string) bool {
	matchPath := newFileMatchPath(filePath, cwd)
	return matchPath.matchesAny(patterns)
}

func (path *fileMatchPath) matchesAny(patterns []string) bool {
	for _, pattern := range patterns {
		if path.matchesSingle(pattern) {
			return true
		}
	}
	return false
}

func (path *fileMatchPath) matchesSingle(pattern string) bool {
	negated := false
	for strings.HasPrefix(pattern, "!") {
		negated = !negated
		pattern = strings.TrimPrefix(pattern, "!")
	}
	matched := path.matchesPositive(pattern)
	if negated {
		return !matched
	}
	return matched
}

func (path *fileMatchPath) matchesPositive(pattern string) bool {
	normalizedPath, unixPath := path.normalizedPaths()

	normalizedPattern := normalizePattern(pattern)

	if utils.MatchGlob(normalizedPattern, normalizedPath) {
		return true
	}
	if normalizedPath != path.original {
		if utils.MatchGlob(normalizedPattern, path.original) {
			return true
		}
	}
	if unixPath != normalizedPath {
		if utils.MatchGlob(normalizedPattern, unixPath) {
			return true
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
			if override.ParserOptions.Project != nil {
				po.Project = override.ParserOptions.Project
			}
			merged.ParserOptions = &po
		}
	}
	merged.Raw = deepMergeConfigObjects(base.Raw, override.Raw)
	return &merged
}

// ExtractGlobals reads the effective `languageOptions.globals` for a merged
// config and normalizes every authored alias to its access level.
//
// Values ValidateConfig rejects are dropped here rather than guessed at, so a
// config that skipped validation behaves as if it never mentioned the name.
func ExtractGlobals(langOpts *LanguageOptions) map[string]utils.GlobalAccess {
	if langOpts == nil || langOpts.Raw == nil {
		return nil
	}
	raw, ok := langOpts.Raw["globals"].(map[string]any)
	if !ok {
		return nil
	}
	globals := make(map[string]utils.GlobalAccess, len(raw))
	for name, value := range raw {
		if access, ok := utils.NormalizeGlobalAccess(value); ok {
			globals[name] = access
		}
	}
	return globals
}

// ExtractLanguageOptions normalizes the effective per-file language options
// for native rules. A missing sourceType keeps ESLint flat config's
// extension-sensitive default; ECMAVersion keeps zero as the moving latest
// edition.
func ExtractLanguageOptions(langOpts *LanguageOptions) rule.LanguageOptions {
	result := rule.LanguageOptions{SourceType: rule.SourceTypeDefault}
	if langOpts == nil || langOpts.Raw == nil {
		return result
	}
	if value, present := langOpts.Raw["ecmaVersion"]; present {
		if version, ok := normalizeConfigECMAVersion(value); ok {
			result.ECMAVersion = version
		}
	}
	if sourceType, ok := langOpts.Raw["sourceType"].(string); ok {
		switch sourceType {
		case "module", "script", "commonjs":
			result.SourceType = sourceType
		}
	}
	return result
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
