package config

import (
	"sort"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"gotest.tools/v3/assert"
)

// A representative mix of authored and gitignore-converted patterns keeps
// target discovery and per-file config resolution on the same matcher path.
func TestRealWorldDiscoverGapFilesMatchesConfigOracle(t *testing.T) {
	layout := []string{
		"packages/core/src/index.ts",
		"packages/core/dist/bundle.ts",
		"packages/core/node_modules/dep/i.ts",
		"target/build/a.ts",
		"tests/rspack-test/configCases/pkg/node_modules/d.ts",
		"tests/rspack-test/configCases/c.ts",
		"scripts/build.ts",
		"npm/darwin-arm64/index.ts",
		"npm/win32-x64-msvc/index.ts",
		"npm/util.ts",
		"src/app/main.ts",
		"src/util.tsx",
	}
	configDir, paths := setupDiscoveryFixture(t, layout)
	config := RslintConfig{
		{Ignores: []string{
			"**/tests/**",
			"**/dist/**/*",
			"**/node_modules/**/*",
			"!tests/rspack-test/*/**/node_modules",
			"**/target/**/*",
			"npm/**/*.node",
			"npm/*",
			"!npm/darwin-arm64/**/*",
		}},
		{Files: []string{"**/*.ts", "**/*.tsx"}, Rules: Rules{"test-rule": "error"}},
	}

	oracle := make([]string, 0, len(paths))
	for _, absolutePath := range paths {
		if config.GetConfigForFile(absolutePath, configDir) != nil {
			oracle = append(oracle, absolutePath)
		}
	}
	sort.Strings(oracle)

	got := DiscoverGapFiles(config, configDir, osvfs.FS(), map[string]struct{}{}, nil, nil, false)
	sort.Strings(got)
	assert.DeepEqual(t, got, oracle)

	want := []string{
		paths["packages/core/src/index.ts"],
		paths["scripts/build.ts"],
		paths["src/app/main.ts"],
		paths["src/util.tsx"],
	}
	sort.Strings(want)
	assert.DeepEqual(t, got, want)
}
