package rule

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestResolveLanguageDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fileName               string
		commonJS               bool
		nonGlobalTopLevelScope bool
		sourceType             string
	}{
		{fileName: "/repo/file.js", nonGlobalTopLevelScope: true, sourceType: "module"},
		{fileName: "/repo/file.mjs", nonGlobalTopLevelScope: true, sourceType: "module"},
		{fileName: "/repo/file.jsx"},
		{fileName: "/repo/file.ts"},
		{fileName: "/repo/file.cts"},
		{fileName: "/repo/file.CJS"},
		{fileName: "/repo/file.cjs.js", nonGlobalTopLevelScope: true, sourceType: "module"},
		{fileName: "/repo/dir.cjs/file.js", nonGlobalTopLevelScope: true, sourceType: "module"},
		{fileName: "/repo/file.cjs", commonJS: true, nonGlobalTopLevelScope: true, sourceType: "commonjs"},
		{fileName: ".cjs", commonJS: true, nonGlobalTopLevelScope: true, sourceType: "commonjs"},
		{fileName: `C:\repo\file.cjs`, commonJS: true, nonGlobalTopLevelScope: true, sourceType: "commonjs"},
	}

	for _, test := range tests {
		t.Run(test.fileName, func(t *testing.T) {
			t.Parallel()
			globalsInit, refsInit, languageOptions := ResolveLanguageDefaults(test.fileName, LanguageOptions{})
			if got := languageOptions.SourceType; got != test.sourceType {
				t.Errorf("sourceType = %q, want %q", got, test.sourceType)
			}

			wantAccess := map[string]utils.GlobalAccess{}
			if test.commonJS {
				wantAccess = map[string]utils.GlobalAccess{
					"exports": utils.GlobalAccessWritable,
					"global":  utils.GlobalAccessReadonly,
					"module":  utils.GlobalAccessReadonly,
					"require": utils.GlobalAccessReadonly,
				}
			}

			for _, name := range []string{
				"exports", "global", "module", "require",
				"arguments", "__dirname", "__filename", "process", "Buffer", "console",
			} {
				if got, want := globalsInit.access(name), wantAccess[name]; got != want {
					t.Errorf("globalsInit.access(%q) = %s, want %s", name, got, want)
				}
			}

			if got := refsInit.hasImplicitWrapperBinding("arguments"); got != test.commonJS {
				t.Errorf("refsInit.hasImplicitWrapperBinding(arguments) = %v, want %v", got, test.commonJS)
			}
			if got := refsInit.nonGlobalTopLevelScope; got != test.nonGlobalTopLevelScope {
				t.Errorf("nonGlobalTopLevelScope = %v, want %v", got, test.nonGlobalTopLevelScope)
			}
			for _, name := range []string{"exports", "global", "module", "require", "process"} {
				if refsInit.hasImplicitWrapperBinding(name) {
					t.Errorf("refsInit unexpectedly defines %q", name)
				}
			}
		})
	}

	t.Run("authored sourceType selects inits independently of extension", func(t *testing.T) {
		tests := []struct {
			fileName               string
			authored               string
			wantSourceType         string
			commonJS               bool
			nonGlobalTopLevelScope bool
		}{
			{fileName: "file.js", authored: "commonjs", wantSourceType: "commonjs", commonJS: true, nonGlobalTopLevelScope: true},
			{fileName: "file.js", authored: "script", wantSourceType: "script"},
			{fileName: "file.cjs", authored: "module", wantSourceType: "module", nonGlobalTopLevelScope: true},
			{fileName: "file.cjs", authored: "script", wantSourceType: "script"},
			{fileName: "file.ts", authored: "commonjs", wantSourceType: "commonjs", commonJS: true, nonGlobalTopLevelScope: true},
			{fileName: "file.ts", authored: "module", wantSourceType: "module", nonGlobalTopLevelScope: true},
			{fileName: "file.tsx", authored: "script", wantSourceType: "script"},
			{fileName: "file.jsx", authored: "commonjs", wantSourceType: "commonjs", commonJS: true, nonGlobalTopLevelScope: true},
		}
		for _, test := range tests {
			t.Run(test.fileName+"/"+test.authored, func(t *testing.T) {
				t.Parallel()
				globalsInit, refsInit, languageOptions := ResolveLanguageDefaults(test.fileName, LanguageOptions{SourceType: test.authored})
				if got := languageOptions.SourceType; got != test.wantSourceType {
					t.Errorf("sourceType = %q, want %q", got, test.wantSourceType)
				}
				if got := languageOptions.EffectiveSourceType(); got != test.wantSourceType {
					t.Errorf("EffectiveSourceType() = %q, want %q", got, test.wantSourceType)
				}
				wantAccess := utils.GlobalAccessUnset
				if test.commonJS {
					wantAccess = utils.GlobalAccessReadonly
				}
				if got := globalsInit.access("require"); got != wantAccess {
					t.Errorf("globalsInit.access(require) = %s, want %s", got, wantAccess)
				}
				if got := refsInit.hasImplicitWrapperBinding("arguments"); got != test.commonJS {
					t.Errorf("hasImplicitWrapperBinding(arguments) = %v, want %v", got, test.commonJS)
				}
				if got := refsInit.nonGlobalTopLevelScope; got != test.nonGlobalTopLevelScope {
					t.Errorf("nonGlobalTopLevelScope = %v, want %v", got, test.nonGlobalTopLevelScope)
				}
			})
		}
	})
}
