package utils_test

import (
	"testing"

	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
)

func TestModuleSettingsIsExternalPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		settings     map[string]interface{}
		specifier    string
		resolvedPath string
		want         bool
	}{
		{
			name:         "resolved node_modules path",
			specifier:    "external-package",
			resolvedPath: "/repo/node_modules/external-package/index.d.ts",
			want:         true,
		},
		{
			name:      "unresolved bare specifier",
			specifier: "external-package",
			want:      true,
		},
		{
			name:      "unresolved relative specifier",
			specifier: "./local",
			want:      false,
		},
		{
			name:      "unresolved absolute path",
			specifier: "/repo/src/local.ts",
			want:      false,
		},
		{
			name:         "ts path alias resolved inside project",
			specifier:    "@cycles/alias-b",
			resolvedPath: "/repo/src/no-cycle/alias-b.ts",
			want:         false,
		},
		{
			name:         "custom external module folder",
			settings:     map[string]interface{}{"import/external-module-folders": []interface{}{"vendor"}},
			specifier:    "@vendor/pkg",
			resolvedPath: "/repo/vendor/pkg/index.ts",
			want:         true,
		},
		{
			name:         "explicit empty folder list disables the default",
			settings:     map[string]interface{}{"import/external-module-folders": []interface{}{}},
			specifier:    "external-package",
			resolvedPath: "/repo/node_modules/external-package/index.d.ts",
			want:         false,
		},
		{
			name:         "empty folder classifies every resolved target",
			settings:     map[string]interface{}{"import/external-module-folders": []interface{}{""}},
			specifier:    "@app/local",
			resolvedPath: "/repo/src/local.ts",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := import_utils.CompileModuleSettings(tt.settings).IsExternalPath(tt.specifier, tt.resolvedPath)
			if got != tt.want {
				t.Fatalf("IsExternalPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleSettingsIsIgnoredPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings map[string]interface{}
		fileName string
		want     bool
	}{
		{
			name:     "array of interface strings matches as regexp",
			settings: map[string]interface{}{"import/ignore": []interface{}{"ignored-missing-default"}},
			fileName: "/repo/ignored-missing-default.ts",
			want:     true,
		},
		{
			name:     "array of strings matches as regexp",
			settings: map[string]interface{}{"import/ignore": []string{`\.css$`}},
			fileName: "/repo/styles.css",
			want:     true,
		},
		{
			name:     "non-string entries and invalid regexps are ignored",
			settings: map[string]interface{}{"import/ignore": []interface{}{123, "["}},
			fileName: "/repo/ignored-missing-default.ts",
			want:     false,
		},
		{
			name:     "missing setting does not ignore",
			settings: map[string]interface{}{},
			fileName: "/repo/ignored-missing-default.ts",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := import_utils.CompileModuleSettings(tc.settings).IsIgnoredPath(tc.fileName)
			if got != tc.want {
				t.Fatalf("IsIgnoredPath() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestModuleSettingsIsInternalSpecifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		settings  map[string]interface{}
		specifier string
		want      bool
	}{
		{
			name:      "matching regexp",
			settings:  map[string]interface{}{"import/internal-regex": `^@app(?:/|$)`},
			specifier: "@app/components/button",
			want:      true,
		},
		{
			name:      "non-matching regexp",
			settings:  map[string]interface{}{"import/internal-regex": `^@app(?:/|$)`},
			specifier: "@application/button",
		},
		{
			name:      "invalid regexp is ignored",
			settings:  map[string]interface{}{"import/internal-regex": `[`},
			specifier: "anything",
		},
		{
			name:      "missing regexp",
			specifier: "@app/button",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compiled := import_utils.CompileModuleSettings(test.settings)
			if got := compiled.IsInternalSpecifier(test.specifier); got != test.want {
				t.Fatalf("IsInternalSpecifier(%q) = %v, want %v", test.specifier, got, test.want)
			}
		})
	}
}

func TestModuleSettingsIsCoreModuleSpecifier(t *testing.T) {
	t.Parallel()

	compiled := import_utils.CompileModuleSettings(map[string]interface{}{
		"import/core-modules": []interface{}{"virtual", "@scope/pkg", "..", "@broken/undefined", 42},
	})
	tests := []struct {
		specifier string
		want      bool
	}{
		{specifier: "fs/promises", want: true},
		{specifier: "node:sqlite/database", want: true},
		{specifier: "virtual/subpath", want: true},
		{specifier: "@scope/pkg/subpath", want: true},
		{specifier: "../missing", want: true},
		{specifier: "@broken", want: true},
		{specifier: "external-package"},
		{specifier: ""},
	}

	for _, test := range tests {
		if got := compiled.IsCoreModuleSpecifier(test.specifier); got != test.want {
			t.Errorf("IsCoreModuleSpecifier(%q) = %v, want %v", test.specifier, got, test.want)
		}
	}
}

func TestIsNodeBuiltinSpecifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		specifier string
		want      bool
	}{
		{specifier: "buffer", want: true},
		{specifier: "node:buffer", want: true},
		{specifier: "fs/promises", want: true},
		{specifier: "node:sqlite", want: true},
		{specifier: "buffer/"},
		{specifier: "fs/not-a-builtin"},
		{specifier: "node:sqlite/database"},
		{specifier: ""},
	}

	for _, test := range tests {
		if got := import_utils.IsNodeBuiltinSpecifier(test.specifier); got != test.want {
			t.Errorf("IsNodeBuiltinSpecifier(%q) = %v, want %v", test.specifier, got, test.want)
		}
	}
}

func TestIsScopedModuleSpecifier(t *testing.T) {
	t.Parallel()

	loneSurrogate := string([]byte{0xED, 0xA0, 0x80})
	tests := []struct {
		specifier string
		want      bool
	}{
		{specifier: "@scope/pkg", want: true},
		{specifier: "@scope", want: true},
		{specifier: "@a/pkg", want: true},
		{specifier: "@😀", want: true},
		{specifier: "@a"},
		{specifier: "@é"},
		{specifier: "@a/"},
		{specifier: "@a//pkg"},
		{specifier: "package"},
		{specifier: "@" + loneSurrogate},
		{specifier: "@" + loneSurrogate + loneSurrogate, want: true},
		{specifier: "@" + loneSurrogate + "/pkg", want: true},
	}

	for _, test := range tests {
		if got := import_utils.IsScopedModuleSpecifier(test.specifier); got != test.want {
			t.Errorf("IsScopedModuleSpecifier(%q) = %v, want %v", test.specifier, got, test.want)
		}
	}
}

func TestModuleSettingsIsExternalPathFromPackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		settings      map[string]interface{}
		packagePath   string
		resolvedPath  string
		caseSensitive bool
		want          bool
	}{
		{name: "dependency inside package", packagePath: "/repo/app", resolvedPath: "/repo/app/node_modules/pkg/index.js", caseSensitive: true, want: true},
		{name: "hoisted dependency", packagePath: "/repo/packages/app", resolvedPath: "/repo/node_modules/pkg/index.js", caseSensitive: true, want: true},
		{name: "target outside package", packagePath: "/repo/packages/app", resolvedPath: "/repo/packages/shared/index.js", caseSensitive: true, want: true},
		{name: "sibling package prefix is outside", packagePath: "/repo/app", resolvedPath: "/repo/application/index.js", caseSensitive: true, want: true},
		{name: "package root itself is internal", packagePath: "/repo/app", resolvedPath: "/repo/app", caseSensitive: true},
		{name: "ordinary source inside package", packagePath: "/repo/app", resolvedPath: "/repo/app/src/index.js", caseSensitive: true},
		{name: "relative segments are normalized", packagePath: "/repo/app", resolvedPath: "/repo/app/src/../node_modules/pkg/index.js", caseSensitive: true, want: true},
		{name: "absolute external folder", settings: map[string]interface{}{"import/external-module-folders": []string{"/dependencies"}}, packagePath: "/repo", resolvedPath: "/dependencies/pkg/index.js", caseSensitive: true, want: true},
		{name: "custom folder exact root", settings: map[string]interface{}{"import/external-module-folders": []string{"vendor"}}, packagePath: "/repo/app", resolvedPath: "/repo/app/vendor", caseSensitive: true, want: true},
		{name: "custom folder sibling prefix", settings: map[string]interface{}{"import/external-module-folders": []string{"vendor"}}, packagePath: "/repo/app", resolvedPath: "/repo/app/vendor-extra/pkg/index.js", caseSensitive: true},
		{name: "empty folder denotes package root", settings: map[string]interface{}{"import/external-module-folders": []string{""}}, packagePath: "/repo", resolvedPath: "/repo/src/local.js", caseSensitive: true, want: true},
		{name: "explicit empty folders keep package target internal", settings: map[string]interface{}{"import/external-module-folders": []string{}}, packagePath: "/repo", resolvedPath: "/repo/node_modules/pkg/index.js", caseSensitive: true},
		{name: "explicit empty folders do not change outside target", settings: map[string]interface{}{"import/external-module-folders": []string{}}, packagePath: "/repo/app", resolvedPath: "/repo/shared/index.js", caseSensitive: true, want: true},
		{name: "case insensitive host", packagePath: "/REPO/APP", resolvedPath: "/repo/app/NODE_MODULES/pkg/index.js", want: true},
		{name: "case sensitive host treats casing mismatch as outside", packagePath: "/REPO/APP", resolvedPath: "/repo/app/src/index.js", caseSensitive: true, want: true},
		{name: "Windows drive roots are case insensitive", packagePath: "C:/repo/app", resolvedPath: "c:/repo/app/node_modules/pkg/index.js", caseSensitive: true, want: true},
		{name: "target on another Windows drive is outside", packagePath: "C:/repo/app", resolvedPath: "D:/deps/pkg/index.js", caseSensitive: false, want: true},
		{name: "empty resolved path is never external", packagePath: "/repo/app", caseSensitive: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compiled := import_utils.CompileModuleSettings(test.settings)
			got := compiled.IsExternalPathFromPackage(test.packagePath, test.resolvedPath, test.caseSensitive)
			if got != test.want {
				t.Fatalf("IsExternalPathFromPackage() = %v, want %v", got, test.want)
			}
		})
	}
}
