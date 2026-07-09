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
