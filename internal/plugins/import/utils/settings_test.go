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
