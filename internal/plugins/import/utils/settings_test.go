package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNodeVersion(t *testing.T) {
	for _, tc := range []struct {
		settings map[string]any
		want     string
	}{
		{nil, "22.0.0"},
		{map[string]any{"import/node-version": "16.0.0"}, "16.0.0"},
		{map[string]any{"import/node-version": "014.018.000"}, "14.18.0"},
		{map[string]any{"import/node-version": "0.0.0"}, "0.0.0"},
		{map[string]any{"node": map[string]any{"version": "16.0.0"}}, "22.0.0"},
	} {
		got, err := import_utils.NodeVersion(tc.settings)
		if err != nil || got.String() != tc.want {
			t.Fatalf("settings %v: got %v, %v; want %s", tc.settings, got, err, tc.want)
		}
	}
	for _, invalid := range []any{nil, false, 16, "", "16", "16.0", "v16.0.0", "^16.0.0", "16.0.0-beta.1", "16.0.0+build", "16.0.0\n", "4294967296.0.0"} {
		_, err := import_utils.NodeVersion(map[string]any{"import/node-version": invalid})
		if err == nil || err.Error() != "`import/node-version` setting must be a string in the format \"10.23.45\" (a semver version, with no leading zero)" {
			t.Fatalf("setting %v: got error %v", invalid, err)
		}
	}
}

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
		{specifier: "_http_agent", want: true},
		{specifier: "node:_stream_readable/extra", want: true},
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
		{name: "hoisted custom folder", settings: map[string]interface{}{"import/external-module-folders": []string{"vendor/"}}, packagePath: "/repo/packages/app", resolvedPath: "/repo/vendor/pkg/index.js", caseSensitive: true, want: true},
		{name: "hoisted folder prefix is not external", packagePath: "/repo/packages/app", resolvedPath: "/repo/node_modules-extra/pkg/index.js", caseSensitive: true},
		{name: "absolute external folder excludes siblings", settings: map[string]interface{}{"import/external-module-folders": []string{"/dependencies"}}, packagePath: "/repo/app", resolvedPath: "/repo/shared/index.js", caseSensitive: true},
		{name: "target outside package", packagePath: "/repo/packages/app", resolvedPath: "/repo/packages/shared/index.js", caseSensitive: true},
		{name: "sibling package prefix is outside", packagePath: "/repo/app", resolvedPath: "/repo/application/index.js", caseSensitive: true},
		{name: "package root itself is internal", packagePath: "/repo/app", resolvedPath: "/repo/app", caseSensitive: true},
		{name: "ordinary source inside package", packagePath: "/repo/app", resolvedPath: "/repo/app/src/index.js", caseSensitive: true},
		{name: "relative segments are normalized", packagePath: "/repo/app", resolvedPath: "/repo/app/src/../node_modules/pkg/index.js", caseSensitive: true, want: true},
		{name: "absolute external folder", settings: map[string]interface{}{"import/external-module-folders": []string{"/dependencies"}}, packagePath: "/repo", resolvedPath: "/dependencies/pkg/index.js", caseSensitive: true, want: true},
		{name: "custom folder exact root", settings: map[string]interface{}{"import/external-module-folders": []string{"vendor"}}, packagePath: "/repo/app", resolvedPath: "/repo/app/vendor", caseSensitive: true, want: true},
		{name: "custom folder sibling prefix", settings: map[string]interface{}{"import/external-module-folders": []string{"vendor"}}, packagePath: "/repo/app", resolvedPath: "/repo/app/vendor-extra/pkg/index.js", caseSensitive: true},
		{name: "empty folder denotes package root", settings: map[string]interface{}{"import/external-module-folders": []string{""}}, packagePath: "/repo", resolvedPath: "/repo/src/local.js", caseSensitive: true, want: true},
		{name: "explicit empty folders keep package target internal", settings: map[string]interface{}{"import/external-module-folders": []string{}}, packagePath: "/repo", resolvedPath: "/repo/node_modules/pkg/index.js", caseSensitive: true},
		{name: "explicit empty folders keep outside target internal", settings: map[string]interface{}{"import/external-module-folders": []string{}}, packagePath: "/repo/app", resolvedPath: "/repo/shared/index.js", caseSensitive: true},
		{name: "case insensitive host", packagePath: "/REPO/APP", resolvedPath: "/repo/app/NODE_MODULES/pkg/index.js", want: true},
		{name: "case sensitive host treats casing mismatch as outside", packagePath: "/REPO/APP", resolvedPath: "/repo/app/src/index.js", caseSensitive: true},
		{name: "Windows drive roots are case insensitive", packagePath: "C:/repo/app", resolvedPath: "c:/repo/app/node_modules/pkg/index.js", caseSensitive: true, want: true},
		{name: "target on another Windows drive is internal", packagePath: "C:/repo/app", resolvedPath: "D:/deps/pkg/index.js", caseSensitive: false},
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

func TestModuleSettingsIsExternalModuleInFolder(t *testing.T) {
	t.Parallel()
	root := tspath.NormalizePath(t.TempDir())
	packagePath := tspath.ResolvePath(root, "packages/app")
	for _, directory := range []string{"packages/app", "node_modules/linked", "node_modules/@scope/linked", "vendor/custom", "absolute/installed"} {
		if err := os.MkdirAll(filepath.FromSlash(tspath.ResolvePath(root, directory)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.FromSlash(tspath.ResolvePath(root, "node_modules/single.js")), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	newProgram := func() *program.Program {
		t.Helper()
		p, err := program.NewFromRoots(program.RootOptions{
			Host:            utils.CreateCompilerHost(packagePath, osvfs.FS()),
			CompilerOptions: &core.CompilerOptions{AllowJs: core.TSTrue, NoLib: core.TSTrue},
			RootFileNames:   []string{tspath.ResolvePath(root, "node_modules/single.js")},
			SingleThreaded:  true,
		})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	p := newProgram()
	for _, tc := range []struct {
		name      string
		folders   any
		specifier string
		want      bool
	}{
		{name: "hoisted package", specifier: "linked/subpath", want: true},
		{name: "scoped package", specifier: "@scope/linked/subpath", want: true},
		{name: "missing package", specifier: "missing"},
		{name: "no prefix match", specifier: "link"},
		{name: "single file", specifier: "single.js", want: true},
		{name: "no implicit extension", specifier: "single"},
		{name: "disabled folders", folders: []string{}, specifier: "linked"},
		{name: "custom folder", folders: []string{"vendor/"}, specifier: "custom/subpath", want: true},
		{name: "absolute folder", folders: []string{tspath.ResolvePath(root, "absolute")}, specifier: "installed", want: true},
		{name: "absolute folder miss", folders: []string{tspath.ResolvePath(root, "absolute")}, specifier: "linked"},
		{name: "empty folder walks ancestors", folders: []string{""}, specifier: "vendor/custom", want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			settings := map[string]any{}
			if tc.folders != nil {
				settings["import/external-module-folders"] = tc.folders
			}
			compiled := import_utils.CompileModuleSettings(settings)
			if got := compiled.IsExternalModuleInFolder(p, packagePath, tc.specifier); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
	compiled := import_utils.CompileModuleSettings(nil)
	if compiled.IsExternalModuleInFolder(p, packagePath, "later/subpath") {
		t.Fatal("missing package was found")
	}
	if err := os.MkdirAll(filepath.FromSlash(tspath.ResolvePath(root, "node_modules/later")), 0o755); err != nil {
		t.Fatal(err)
	}
	if compiled.IsExternalModuleInFolder(p, packagePath, "later/another-subpath") {
		t.Fatal("one generation did not share its package lookup")
	}
	if !compiled.IsExternalModuleInFolder(newProgram(), packagePath, "later/subpath") {
		t.Fatal("new generation reused a stale package lookup")
	}
}
