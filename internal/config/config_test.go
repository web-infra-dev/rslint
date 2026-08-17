package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestProjectPathsUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single string",
			input:    `"tsconfig.json"`,
			expected: []string{"tsconfig.json"},
		},
		{
			name:     "array of strings",
			input:    `["tsconfig.json", "packages/*/tsconfig.json"]`,
			expected: []string{"tsconfig.json", "packages/*/tsconfig.json"},
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paths ProjectPaths
			err := json.Unmarshal([]byte(tt.input), &paths)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if len(paths) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(paths))
			}

			for i, expected := range tt.expected {
				if i >= len(paths) {
					t.Errorf("Expected %s at index %d, but paths is too short", expected, i)
					continue
				}
				if paths[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, paths[i])
				}
			}
		})
	}
}

func TestParserOptionsUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ProjectPaths
	}{
		{
			name:     "single project string",
			input:    `{"projectService": false, "project": "tsconfig.json"}`,
			expected: ProjectPaths{"tsconfig.json"},
		},
		{
			name:     "multiple project strings",
			input:    `{"projectService": false, "project": ["tsconfig.json", "packages/*/tsconfig.json"]}`,
			expected: ProjectPaths{"tsconfig.json", "packages/*/tsconfig.json"},
		},
		{
			name:     "no project field",
			input:    `{"projectService": false}`,
			expected: ProjectPaths{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts ParserOptions
			err := json.Unmarshal([]byte(tt.input), &opts)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if len(opts.Project) != len(tt.expected) {
				t.Errorf("Expected project length %d, got %d", len(tt.expected), len(opts.Project))
			}

			for i, expected := range tt.expected {
				if i >= len(opts.Project) {
					t.Errorf("Expected %s at index %d, but project is too short", expected, i)
					continue
				}
				if opts.Project[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, opts.Project[i])
				}
			}
		})
	}
}

func TestParserOptionsProjectServicePtr(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectNil     bool
		expectedValue bool
	}{
		{
			name:      "not set",
			input:     `{}`,
			expectNil: true,
		},
		{
			name:          "explicitly true",
			input:         `{"projectService": true}`,
			expectNil:     false,
			expectedValue: true,
		},
		{
			name:          "explicitly false",
			input:         `{"projectService": false}`,
			expectNil:     false,
			expectedValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts ParserOptions
			err := json.Unmarshal([]byte(tt.input), &opts)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}
			if tt.expectNil {
				if opts.ProjectService != nil {
					t.Errorf("Expected ProjectService to be nil, got %v", *opts.ProjectService)
				}
			} else {
				if opts.ProjectService == nil {
					t.Fatalf("Expected ProjectService to be non-nil")
					return
				}
				if *opts.ProjectService != tt.expectedValue {
					t.Errorf("Expected ProjectService to be %v, got %v", tt.expectedValue, *opts.ProjectService)
				}
			}
		})
	}
}

// TestRegisterAllRules_ConcurrentSafe verifies that RegisterAllRules can be
// called from multiple goroutines without panicking (concurrent map writes).
// Run with -race to detect data races: go test -race ./internal/config/...
func TestRegisterAllRules_ConcurrentSafe(t *testing.T) {
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			RegisterAllRules()
		}()
	}
	wg.Wait()

	// Verify rules were actually registered
	rules := GlobalRuleRegistry.GetAllRules()
	if len(rules) == 0 {
		t.Error("Expected rules to be registered after concurrent calls")
	}
}

func TestCwdHandling(t *testing.T) {
	// Save the original working directory
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get current working directory: %v", err)
	}
	defer t.Chdir(originalCwd) // Restore after test completes

	tests := []struct {
		name         string
		filePath     string
		patterns     []string
		shouldIgnore bool
		description  string
	}{
		{
			name:         "Relative path matching",
			filePath:     "src/component.ts",
			patterns:     []string{"src/**"},
			shouldIgnore: true,
			description:  "Relative paths should match relative patterns",
		},
		{
			name:         "Absolute path to relative path matching",
			filePath:     filepath.Join(originalCwd, "src/component.ts"),
			patterns:     []string{"src/**"},
			shouldIgnore: true,
			description:  "Absolute paths should be converted to relative paths before matching",
		},
		{
			name:         "node_modules recursive matching",
			filePath:     "node_modules/package/deep/file.js",
			patterns:     []string{"node_modules/**"},
			shouldIgnore: true,
			description:  "Recursive patterns should match deeply nested files",
		},
		{
			name:         "Test file pattern matching",
			filePath:     "src/utils/helper.test.ts",
			patterns:     []string{"**/*.test.ts"},
			shouldIgnore: true,
			description:  "Global recursive patterns should match test files in any location",
		},
		{
			name:         "Non-matching file",
			filePath:     "src/component.ts",
			patterns:     []string{"dist/**", "*.log"},
			shouldIgnore: false,
			description:  "Files not matching any pattern should not be ignored",
		},
		{
			name:         "Cross-platform path handling",
			filePath:     "src\\windows\\style\\path.ts",
			patterns:     []string{"src/**"},
			shouldIgnore: true,
			description:  "Windows style paths should be handled correctly",
		},
		{
			name:         "Pattern with ./ prefix matches normalized path",
			filePath:     "src/component.ts",
			patterns:     []string{"./src/**"},
			shouldIgnore: true,
			description:  "Patterns with ./ prefix should match paths without ./ prefix",
		},
		{
			name:         "Pattern with ./ prefix matches exact file",
			filePath:     "src/rslint-test-cases.ts",
			patterns:     []string{"./src/rslint-test-cases.ts"},
			shouldIgnore: true,
			description:  "Exact file pattern with ./ prefix should match normalized path",
		},
		{
			name:         "Pattern with .. segment ignores correctly",
			filePath:     "lib/component.ts",
			patterns:     []string{"src/../lib/**"},
			shouldIgnore: true,
			description:  "Patterns with .. segments should resolve and match",
		},
		{
			name:         "Pattern with mid-path /./ ignores correctly",
			filePath:     "src/utils/helper.ts",
			patterns:     []string{"src/./utils/**"},
			shouldIgnore: true,
			description:  "Patterns with mid-path /./ should collapse and match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileIgnored(tt.filePath, ParseIgnorePatterns(tt.patterns), originalCwd)
			if result != tt.shouldIgnore {
				t.Errorf("%s: isFileIgnored(%q, %v) = %v, expected %v",
					tt.description, tt.filePath, tt.patterns, result, tt.shouldIgnore)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get working directory: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "Relative path remains unchanged",
			filePath: "src/component.ts",
			expected: "src/component.ts",
		},
		{
			name:     "Absolute path converts to relative path",
			filePath: filepath.Join(cwd, "src/component.ts"),
			expected: "src/component.ts",
		},
		{
			name:     "Current directory marker",
			filePath: "./src/component.ts",
			expected: "src/component.ts",
		},
		{
			name:     "Complex relative path",
			filePath: "src/../src/component.ts",
			expected: "src/component.ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.filePath, cwd)
			if result != tt.expected {
				t.Errorf("normalizePath(%q, %q) = %q, expected %q",
					tt.filePath, cwd, result, tt.expected)
			}
		})
	}
}

func TestNormalizePattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{name: "No-op for clean pattern", pattern: "src/**/*.ts", expected: "src/**/*.ts"},
		{name: "Strip leading ./", pattern: "./src/**/*.ts", expected: "src/**/*.ts"},
		{name: "Collapse mid-path /./", pattern: "src/./utils/*.ts", expected: "src/utils/*.ts"},
		{name: "Resolve .. segment", pattern: "src/../lib/*.ts", expected: "lib/*.ts"},
		{name: "Combined ./ prefix and .. segment", pattern: "./src/../lib/*.ts", expected: "lib/*.ts"},
		{name: "Multiple /./", pattern: "src/./utils/./deep/*.ts", expected: "src/utils/deep/*.ts"},
		{name: "Exact file with ./", pattern: "./src/file.ts", expected: "src/file.ts"},
		{name: "Plain filename unchanged", pattern: "*.ts", expected: "*.ts"},
		{name: "Double star unchanged", pattern: "**/*.ts", expected: "**/*.ts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePattern(tt.pattern)
			if result != tt.expected {
				t.Errorf("normalizePattern(%q) = %q, expected %q",
					tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestIsFileMatchedDotSlashPrefix(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get working directory: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		patterns []string
		expected bool
	}{
		{
			name:     "Pattern with ./ prefix matches relative path",
			filePath: "src/rslint-test-cases.ts",
			patterns: []string{"./src/rslint-test-cases.ts"},
			expected: true,
		},
		{
			name:     "Pattern with ./ prefix glob matches relative path",
			filePath: "src/component.ts",
			patterns: []string{"./src/**"},
			expected: true,
		},
		{
			name:     "Pattern without ./ still works",
			filePath: "src/component.ts",
			patterns: []string{"src/**"},
			expected: true,
		},
		{
			name:     "Non-matching pattern with ./ prefix",
			filePath: "lib/component.ts",
			patterns: []string{"./src/**"},
			expected: false,
		},
		{
			name:     "Pattern with mid-path /./",
			filePath: "src/utils/helper.ts",
			patterns: []string{"src/./utils/*.ts"},
			expected: true,
		},
		{
			name:     "Pattern with .. segment",
			filePath: "lib/component.ts",
			patterns: []string{"src/../lib/*.ts"},
			expected: true,
		},
		{
			name:     "Pattern with .. segment non-matching",
			filePath: "src/component.ts",
			patterns: []string{"src/../lib/*.ts"},
			expected: false,
		},
		{
			name:     "Pattern with combined ./ prefix and .. segment",
			filePath: "lib/component.ts",
			patterns: []string{"./src/../lib/*.ts"},
			expected: true,
		},
		{
			name:     "Pattern with ./ prefix and ** glob",
			filePath: "src/deep/nested/file.ts",
			patterns: []string{"./src/**/*.ts"},
			expected: true,
		},
		{
			name:     "Multiple patterns with first matching",
			filePath: "src/component.ts",
			patterns: []string{"./lib/**", "./src/**"},
			expected: true,
		},
		{
			name:     "Multiple patterns with none matching",
			filePath: "test/component.ts",
			patterns: []string{"./lib/**", "./src/**"},
			expected: false,
		},
		{
			name:     "Empty patterns list",
			filePath: "src/component.ts",
			patterns: []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFileMatched(tt.filePath, tt.patterns, cwd)
			if result != tt.expected {
				t.Errorf("isFileMatched(%q, %v, %q) = %v, expected %v",
					tt.filePath, tt.patterns, cwd, result, tt.expected)
			}
		})
	}
}

func TestDoublestarBehavior(t *testing.T) {
	// Test specific behavior of the doublestar library
	tests := []struct {
		pattern  string
		path     string
		expected bool
		name     string
	}{
		{"**/*.ts", "src/file.ts", true, "Recursive matching of TypeScript files"},
		{"**/*.ts", "src/deep/nested/file.ts", true, "Deep recursive matching"},
		{"src/**", "src/file.ts", true, "Directory recursive matching"},
		{"src/**", "src/deep/nested/file.ts", true, "Deep directory recursive matching"},
		{"*.ts", "file.ts", true, "Single-level wildcard"},
		{"*.ts", "src/file.ts", false, "Single-level wildcard doesn't match nested files"},
		{"node_modules/**", "node_modules/package/file.js", true, "node_modules recursive matching"},
		{"**/test/**", "src/test/file.js", true, "Middle recursive matching"},
		{"**/test/**", "test/file.js", true, "Beginning recursive matching"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := doublestar.Match(tt.pattern, tt.path)
			if err != nil {
				t.Errorf("doublestar.PathMatch error: %v", err)
				return
			}
			if matched != tt.expected {
				t.Errorf("doublestar.PathMatch(%q, %q) = %v, expected %v",
					tt.pattern, tt.path, matched, tt.expected)
			}
		})
	}
}

func TestExtractLanguageOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  map[string]any
		want rule.LanguageOptions
	}{
		{name: "missing uses zero-value latest"},
		{
			name: "explicit latest remains semantic latest",
			raw:  map[string]any{"ecmaVersion": "latest"},
			want: rule.LanguageOptions{},
		},
		{
			name: "edition alias 6",
			raw:  map[string]any{"ecmaVersion": float64(6)},
			want: rule.LanguageOptions{ECMAVersion: 2015},
		},
		{
			name: "edition alias 17",
			raw:  map[string]any{"ecmaVersion": 17},
			want: rule.LanguageOptions{ECMAVersion: 2026},
		},
		{
			name: "year",
			raw:  map[string]any{"ecmaVersion": float64(2020)},
			want: rule.LanguageOptions{ECMAVersion: 2020},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var input *LanguageOptions
			if tt.raw != nil {
				input = &LanguageOptions{Raw: tt.raw}
			}
			got := ExtractLanguageOptions(input)
			if got != tt.want {
				t.Fatalf("ExtractLanguageOptions() = %+v, want %+v", got, tt.want)
			}
		})
	}

	defaults := ExtractLanguageOptions(nil)
	if got := defaults.EffectiveECMAVersion(); got != rule.LatestECMAScriptVersion {
		t.Fatalf("default ecmaVersion = %d, want latest (%d)", got, rule.LatestECMAScriptVersion)
	}
}

func TestValidateConfigECMAVersion(t *testing.T) {
	t.Parallel()

	for _, value := range []any{
		"latest", 3, 5, 6, 17, 2015, 2026, float64(2020), float32(2021),
	} {
		cfg := RslintConfig{{LanguageOptions: &LanguageOptions{Raw: map[string]any{"ecmaVersion": value}}}}
		if err := ValidateConfig(cfg); err != nil {
			t.Errorf("valid ecmaVersion %#v rejected: %v", value, err)
		}
	}

	for _, value := range []any{
		nil, "2020", "LATEST", 4, 18, 2014, 2027, 6.5, true, []any{2020},
	} {
		cfg := RslintConfig{{LanguageOptions: &LanguageOptions{Raw: map[string]any{"ecmaVersion": value}}}}
		err := ValidateConfig(cfg)
		if err == nil {
			t.Errorf("invalid ecmaVersion %#v was accepted", value)
			continue
		}
		if !strings.Contains(err.Error(), "key \"languageOptions.ecmaVersion\"") {
			t.Errorf("error for %#v does not identify ecmaVersion: %v", value, err)
		}
	}
}

func TestRuleRegistryPropagatesLanguageOptions(t *testing.T) {
	t.Parallel()

	registry := NewRuleRegistry()
	registry.Register("probe", rule.Rule{Name: "probe"})
	configured, _ := registry.GetEnabledRules(RslintConfig{{
		LanguageOptions: &LanguageOptions{Raw: map[string]any{
			"ecmaVersion": float64(16),
		}},
		Rules: Rules{"probe": "error"},
	}}, "file.js", "", false)
	if len(configured) != 1 {
		t.Fatalf("configured rules = %d, want 1", len(configured))
	}
	want := rule.LanguageOptions{ECMAVersion: 2025}
	if got := configured[0].Environment.LanguageOptions; got != want {
		t.Fatalf("configured language options = %+v, want %+v", got, want)
	}
}

func TestMergedLanguageOptionsCanRestoreLatest(t *testing.T) {
	t.Parallel()

	merged := mergeLanguageOptions(
		&LanguageOptions{Raw: map[string]any{
			"ecmaVersion": float64(5),
		}},
		&LanguageOptions{Raw: map[string]any{
			"ecmaVersion": "latest",
		}},
	)
	got := ExtractLanguageOptions(merged)
	if got.ECMAVersion != 0 || got.EffectiveECMAVersion() != rule.LatestECMAScriptVersion {
		t.Fatalf("merged ecmaVersion = %+v, want semantic latest", got)
	}
}

// ExtractGlobals resolves every alias ESLint accepts to one of its three access
// levels. Booleans follow the `globals` npm package: true is writable, false is
// readonly.
func TestExtractGlobals(t *testing.T) {
	langOpts := &LanguageOptions{
		Raw: map[string]any{
			"globals": map[string]any{
				"boolTrue":       true,
				"stringTrue":     "true",
				"writable":       "writable",
				"writeable":      "writeable",
				"boolFalse":      false,
				"stringFalse":    "false",
				"readonly":       "readonly",
				"readable":       "readable",
				"nullReadonly":   nil,
				"stringDisabled": "off",
				"invalid":        "nonsense",
			},
		},
	}

	globals := ExtractGlobals(langOpts)

	cases := map[string]utils.GlobalAccess{
		"boolTrue":       utils.GlobalAccessWritable,
		"stringTrue":     utils.GlobalAccessWritable,
		"writable":       utils.GlobalAccessWritable,
		"writeable":      utils.GlobalAccessWritable,
		"boolFalse":      utils.GlobalAccessReadonly,
		"stringFalse":    utils.GlobalAccessReadonly,
		"readonly":       utils.GlobalAccessReadonly,
		"readable":       utils.GlobalAccessReadonly,
		"nullReadonly":   utils.GlobalAccessReadonly,
		"stringDisabled": utils.GlobalAccessOff,
		"invalid":        utils.GlobalAccessUnset,
	}
	for name, want := range cases {
		if got := globals[name]; got != want {
			t.Errorf("ExtractGlobals()[%q] = %v, want %v", name, got, want)
		}
	}
}

func TestExtractGlobals_NoLanguageOptions(t *testing.T) {
	if got := ExtractGlobals(nil); got != nil {
		t.Errorf("ExtractGlobals(nil) = %v, want nil", got)
	}
	if got := ExtractGlobals(&LanguageOptions{}); got != nil {
		t.Errorf("ExtractGlobals(empty) = %v, want nil", got)
	}
}

func TestMergeLanguageOptions_MergesGlobalsByName(t *testing.T) {
	base := &LanguageOptions{Raw: map[string]any{
		"globals": map[string]any{
			"baseOnly": "readonly",
			"shared":   "writable",
		},
	}}
	override := &LanguageOptions{Raw: map[string]any{
		"globals": map[string]any{
			"overrideOnly": "readonly",
			"shared":       "off",
		},
	}}

	merged := mergeLanguageOptions(base, override)
	merged = mergeLanguageOptions(merged, &LanguageOptions{Raw: map[string]any{
		"globals": map[string]any{},
	}})

	rawGlobals, ok := merged.Raw["globals"].(map[string]any)
	if !ok {
		t.Fatalf("merged globals has type %T, want map[string]any", merged.Raw["globals"])
	}
	if got := rawGlobals["baseOnly"]; got != "readonly" {
		t.Errorf("baseOnly = %v, want readonly", got)
	}
	if got := rawGlobals["overrideOnly"]; got != "readonly" {
		t.Errorf("overrideOnly = %v, want readonly", got)
	}
	if got := rawGlobals["shared"]; got != "off" {
		t.Errorf("shared = %v, want off", got)
	}
	if got := ExtractGlobals(merged)["shared"]; got != utils.GlobalAccessOff {
		t.Errorf("later off value should undeclare shared, got %v", got)
	}
}

// TestParseArrayRuleConfig_OptionShapes pins how an array-style rule config's
// post-severity args map to RuleConfig.Options: always the raw remaining
// slice, with no bare-value collapsing. This also covers the load-bearing
// case of a lone option that is itself an array (["error", ["a","b"]]): it
// naturally stays wrapped as [["a","b"]], distinguishable from a two-element
// option list (["error","a","b"]) which becomes ["a","b"].
func TestParseArrayRuleConfig_OptionShapes(t *testing.T) {
	tests := []struct {
		name      string
		in        []interface{}
		wantLevel string
		want      []interface{}
	}{
		{
			name:      "no options",
			in:        []interface{}{"error"},
			wantLevel: "error",
			want:      nil,
		},
		{
			name:      "numeric zero is off",
			in:        []interface{}{0},
			wantLevel: "off",
			want:      nil,
		},
		{
			name:      "numeric one is warn",
			in:        []interface{}{float64(1)},
			wantLevel: "warn",
			want:      nil,
		},
		{
			name:      "numeric two is error",
			in:        []interface{}{uint8(2)},
			wantLevel: "error",
			want:      nil,
		},
		{
			name:      "single string option stays wrapped",
			in:        []interface{}{"error", "both"},
			wantLevel: "error",
			want:      []interface{}{"both"},
		},
		{
			name:      "single object option stays wrapped",
			in:        []interface{}{"error", map[string]interface{}{"k": float64(1)}},
			wantLevel: "error",
			want:      []interface{}{map[string]interface{}{"k": float64(1)}},
		},
		{
			name:      "single array option keeps its wrapper",
			in:        []interface{}{"error", []interface{}{"a", "b"}},
			wantLevel: "error",
			want:      []interface{}{[]interface{}{"a", "b"}},
		},
		{
			name:      "multiple options pass through as the args list",
			in:        []interface{}{"error", "a", "b"},
			wantLevel: "error",
			want:      []interface{}{"a", "b"},
		},
		{
			name:      "multiple options including an object",
			in:        []interface{}{"error", "both", map[string]interface{}{"k": float64(1)}},
			wantLevel: "error",
			want:      []interface{}{"both", map[string]interface{}{"k": float64(1)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := parseArrayRuleConfig(tt.in)
			if rc == nil {
				t.Fatal("parseArrayRuleConfig returned nil")
				return
			}
			if rc.Level != tt.wantLevel {
				t.Errorf("Level = %q, want %q", rc.Level, tt.wantLevel)
			}
			if !reflect.DeepEqual(rc.Options, tt.want) {
				t.Errorf("Options = %#v, want %#v", rc.Options, tt.want)
			}
		})
	}
}
