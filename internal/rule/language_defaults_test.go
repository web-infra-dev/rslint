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
	}{
		{fileName: "/repo/file.js", nonGlobalTopLevelScope: true},
		{fileName: "/repo/file.mjs", nonGlobalTopLevelScope: true},
		{fileName: "/repo/file.jsx"},
		{fileName: "/repo/file.ts"},
		{fileName: "/repo/file.cts"},
		{fileName: "/repo/file.CJS"},
		{fileName: "/repo/file.cjs.js", nonGlobalTopLevelScope: true},
		{fileName: "/repo/dir.cjs/file.js", nonGlobalTopLevelScope: true},
		{fileName: "/repo/file.cjs", commonJS: true, nonGlobalTopLevelScope: true},
		{fileName: ".cjs", commonJS: true, nonGlobalTopLevelScope: true},
		{fileName: `C:\repo\file.cjs`, commonJS: true, nonGlobalTopLevelScope: true},
	}

	for _, test := range tests {
		t.Run(test.fileName, func(t *testing.T) {
			t.Parallel()
			globalsInit, refsInit := ResolveLanguageDefaults(test.fileName)

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

			if got := refsInit.defines("arguments"); got != test.commonJS {
				t.Errorf("refsInit.defines(arguments) = %v, want %v", got, test.commonJS)
			}
			if got := refsInit.nonGlobalTopLevelScope; got != test.nonGlobalTopLevelScope {
				t.Errorf("nonGlobalTopLevelScope = %v, want %v", got, test.nonGlobalTopLevelScope)
			}
			for _, name := range []string{"exports", "global", "module", "require", "process"} {
				if refsInit.defines(name) {
					t.Errorf("refsInit unexpectedly defines %q", name)
				}
			}
		})
	}
}
