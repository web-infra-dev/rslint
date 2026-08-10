package utils_test

import (
	"testing"

	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
)

func TestIsExternalModulePath(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := import_utils.IsExternalModulePath(tt.settings, tt.specifier, tt.resolvedPath)
			if got != tt.want {
				t.Fatalf("IsExternalModulePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsExternalModulePathFromPackage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		settings      map[string]interface{}
		packagePath   string
		resolvedPath  string
		caseSensitive bool
		want          bool
	}{
		{name: "default folder inside package", packagePath: "/repo", resolvedPath: "/repo/node_modules/pkg/index.js", caseSensitive: true, want: true},
		{name: "similar segment is internal", packagePath: "/repo", resolvedPath: "/repo/src/node_modules-ish/pkg.ts", caseSensitive: true, want: false},
		{name: "relative custom folder", settings: map[string]interface{}{"import/external-module-folders": []interface{}{"vendor"}}, packagePath: "/repo", resolvedPath: "/repo/vendor/pkg/index.js", caseSensitive: true, want: true},
		{name: "hoisted default folder", packagePath: "/repo/packages/app", resolvedPath: "/repo/node_modules/pkg/index.js", caseSensitive: true, want: true},
		{name: "outside without module segment", packagePath: "/repo/packages/app", resolvedPath: "/repo/shared/pkg/index.js", caseSensitive: true, want: false},
		{name: "absolute custom folder", settings: map[string]interface{}{"import/external-module-folders": []string{"/dependencies"}}, packagePath: "/repo", resolvedPath: "/dependencies/pkg/index.js", caseSensitive: true, want: true},
		{name: "absolute folder prefix is not enough", settings: map[string]interface{}{"import/external-module-folders": []string{"/dependencies"}}, packagePath: "/repo", resolvedPath: "/dependencies-extra/pkg/index.js", caseSensitive: true, want: false},
		{name: "case insensitive host", packagePath: "/REPO", resolvedPath: "/repo/NODE_MODULES/pkg/index.js", caseSensitive: false, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := import_utils.IsExternalModulePathFromPackage(test.settings, test.packagePath, test.resolvedPath, test.caseSensitive)
			if got != test.want {
				t.Fatalf("IsExternalModulePathFromPackage() = %v, want %v", got, test.want)
			}
		})
	}
}
