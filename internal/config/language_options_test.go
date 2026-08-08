package config

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

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
	if got := configured[0].LanguageOptions; got != want {
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
