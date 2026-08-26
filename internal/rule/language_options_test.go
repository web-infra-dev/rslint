package rule

import "testing"

func TestNormalizeECMAScriptVersion(t *testing.T) {
	t.Parallel()

	valid := map[int]int{3: 3, 5: 5}
	latestEditionAlias := LatestECMAScriptVersion - 2009
	for edition := 6; edition <= latestEditionAlias; edition++ {
		valid[edition] = edition + 2009
	}
	for year := 2015; year <= LatestECMAScriptVersion; year++ {
		valid[year] = year
	}
	for input, want := range valid {
		got, ok := NormalizeECMAScriptVersion(input)
		if !ok || got != want {
			t.Errorf("NormalizeECMAScriptVersion(%d) = (%d, %v), want (%d, true)", input, got, ok, want)
		}
	}
	for _, input := range []int{-1, 0, 1, 4, latestEditionAlias + 1, 2014, LatestECMAScriptVersion + 1} {
		if got, ok := NormalizeECMAScriptVersion(input); ok {
			t.Errorf("NormalizeECMAScriptVersion(%d) = (%d, true), want invalid", input, got)
		}
	}
}

func TestLanguageOptionsDefaultECMAScriptVersion(t *testing.T) {
	t.Parallel()

	var defaults LanguageOptions
	if got := defaults.EffectiveECMAVersion(); got != LatestECMAScriptVersion {
		t.Fatalf("zero-value ecmaVersion = %d, want latest %d", got, LatestECMAScriptVersion)
	}
}

func TestLanguageOptionsEffectiveSourceType(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		options LanguageOptions
		want    string
	}{
		{want: "module"},
		{options: LanguageOptions{SourceType: "module"}, want: "module"},
		{options: LanguageOptions{SourceType: "script"}, want: "script"},
		{options: LanguageOptions{SourceType: "commonjs"}, want: "commonjs"},
	} {
		if got := test.options.EffectiveSourceType(); got != test.want {
			t.Errorf("EffectiveSourceType() = %q, want %q", got, test.want)
		}
	}
}
