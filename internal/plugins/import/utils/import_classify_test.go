package utils_test

import (
	"testing"

	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
)

func TestBaseModuleMatchesImportType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{name: "package/subpath", want: "package"},
		{name: "@scope/package/subpath", want: "@scope/package"},
		{name: "@scope", want: "@scope/undefined"},
		{name: "@scope/", want: "@scope/"},
		{name: "node:fs", want: "node:fs"},
	}
	for _, test := range tests {
		if got := import_utils.BaseModule(test.name); got != test.want {
			t.Errorf("BaseModule(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestMalformedScopedCoreModuleUsesJavaScriptUndefined(t *testing.T) {
	t.Parallel()

	if !import_utils.IsBuiltinModule("@scope", map[string]interface{}{
		"import/core-modules": []interface{}{"@scope/undefined"},
	}, "") {
		t.Fatal("malformed scoped name did not use importType.js's base module")
	}
	if import_utils.IsBuiltinModule("@scope", map[string]interface{}{
		"import/core-modules": []interface{}{"@scope"},
	}, "") {
		t.Fatal("malformed scoped name was classified using the unnormalized name")
	}
}
