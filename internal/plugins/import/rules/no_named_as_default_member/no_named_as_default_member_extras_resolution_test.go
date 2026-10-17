package no_named_as_default_member_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_as_default_member"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
)

func TestNoNamedAsDefaultMemberResolutionFormats(t *testing.T) {
	root := defaultMemberRoot(t, "testdata/module-resolution-formats.txtar")
	for _, tc := range []struct {
		name     string
		config   string
		fileName string
		source   string
		warn     bool
	}{
		{"nested-import-cjs", "tsconfig.json", "consumer.mts", "nested", false},
		{"nested-require-esm", "tsconfig.json", "consumer.cts", "nested", true},
		{"subpath-cjs", "tsconfig.json", "consumer.mts", "nested/cjs", false},
		{"subpath-esm", "tsconfig.json", "consumer.mts", "nested/esm", true},
		{"custom-default-esm", "tsconfig.json", "consumer.mts", "custom", true},
		{"custom-development-cjs", "tsconfig.custom.json", "consumer.mts", "custom", false},
		{"default-condition-order", "tsconfig.custom.json", "consumer.mts", "default-first", false},
		{"custom-types-decoy", "tsconfig.custom-types.json", "consumer.mts", "custom-types", true},
		{"declaration-export-only", "tsconfig.json", "consumer.mts", "declaration-export-only", true},
		{"declaration-pattern-export", "tsconfig.json", "consumer.mts", "declaration-pattern/types.d.ts", true},
		{"declaration-main-only", "tsconfig.json", "consumer.mts", "declaration-main-only", true},
		{"declaration-main-cjs-subpath", "tsconfig.json", "consumer.mts", "declaration-main-only/feature.js", false},
		{"declaration-main-exports-override", "tsconfig.json", "consumer.mts", "declaration-main-exports-override", false},
		{"declaration-export-array", "tsconfig.json", "consumer.mts", "declaration-export-array", true},
		{"declaration-export-condition", "tsconfig.custom.json", "consumer.mts", "declaration-export-condition", true},
		{"declaration-import-only", "tsconfig.json", "consumer.mts", "#declaration", true},
		{"custom-versioned-types-decoy", "tsconfig.versioned-types.json", "consumer.mts", "custom-versioned-types", true},
		{"declaration-path-implementation-decoy", "tsconfig.paths.json", "consumer.mts", "mapped-types-esm", true},
		{"imports-cjs", "tsconfig.json", "consumer.mts", "#cjs", false},
		{"self-reference-esm", "tsconfig.json", "consumer.mts", "resolution-fixture", true},
		{"explicit-dcts-path", "tsconfig.paths.json", "consumer.mts", "declaration-cjs", false},
		{"explicit-unknown-declaration-path", "tsconfig.paths.json", "consumer.mts", "declaration-unknown", true},
		{"types-versions-cjs", "tsconfig.json", "consumer.mts", "versioned-cjs", false},
		{"types-versions-esm", "tsconfig.json", "consumer.mts", "versioned-esm", true},
		{"types-versions-implementation-decoy", "tsconfig.json", "consumer.mts", "versioned-decoy", true},
		{"suffix-cjs", "tsconfig.suffixes.json", "consumer.mts", "suffix-cjs", false},
		{"suffix-esm", "tsconfig.suffixes.json", "consumer.mts", "suffix-esm", true},
		{"root-dirs-cjs", "tsconfig.roots.json", "src/consumer.ts", "./cjs.js", false},
		{"root-dirs-esm", "tsconfig.roots.json", "src/consumer.ts", "./esm.js", true},
		{"nested-package-cjs", "tsconfig.json", "consumer.mts", "nested-scope/cjs/index.js", false},
		{"nested-package-esm", "tsconfig.json", "consumer.mts", "nested-scope/esm/index.js", true},
		{"extensionless-json-main", "tsconfig.json-runtime.json", "consumer.cts", "extensionless-json-main", false},
		{"extensionless-json-main-disabled-types", "tsconfig.no-json.json", "consumer.cts", "extensionless-json-main", false},
		{"javascript-cjs-with-typescript-decoy", "tsconfig.json", "consumer.mts", "js-cjs-ts-decoy", false},
		{"javascript-esm-with-typescript-decoy", "tsconfig.json", "consumer.mts", "js-esm-ts-decoy", true},
		{"query-cjs-entry", "tsconfig.json", "consumer.mts", "query-cjs", false},
		{"fragment-cjs-entry", "tsconfig.json", "consumer.mts", "fragment-cjs", false},
		{"encoded-cjs-entry", "tsconfig.json", "consumer.mts", "encoded-cjs", false},
		{"query-array-esm-entry", "tsconfig.json", "consumer.mts", "query-array-esm", true},
		{"encoded-array-esm-entry", "tsconfig.json", "consumer.mts", "encoded-array-esm", true},
		{"explicit-typescript-esm-entry", "tsconfig.json", "consumer.mts", "#explicit-ts-esm", true},
		{"explicit-typescript-cjs-entry", "tsconfig.json", "consumer.mts", "#explicit-ts-cjs", false},
		{"paths-typescript-esm", "tsconfig.implementation-paths.json", "consumer.mts", "mapped-ts-esm", true},
		{"paths-typescript-cjs", "tsconfig.implementation-paths.json", "consumer.mts", "mapped-ts-cjs", false},
		{"paths-source-typescript-substitution", "tsconfig.implementation-paths.json", "consumer.mts", "mapped-ts-source", true},
		{"suffix-package-exports-cjs", "tsconfig.suffixes.json", "consumer.mts", "suffix-export-cjs", false},
		{"main-before-suffixed-index", "tsconfig.suffixes.json", "consumer.mts", "main-before-index", false},
		{"main-fallback-suffixed-index", "tsconfig.suffixes-only.json", "consumer.mts", "missing-main-index", false},
		{"directory-main-cjs-typescript-decoy", "tsconfig.json", "consumer.mts", "directory-main-cjs", false},
		{"directory-main-esm-typescript-decoy", "tsconfig.json", "consumer.mts", "directory-main-esm", true},
		{"directory-main-suffixed-index", "tsconfig.suffixes-only.json", "consumer.mts", "directory-main-suffix", false},
		{"directory-main-empty-suffix-first", "tsconfig.suffixes-empty-first.json", "consumer.mts", "directory-main-suffix", true},
		{"directory-main-json-types-disabled", "tsconfig.no-json.json", "consumer.cts", "directory-main-json", false},
		{"directory-main-extension-before-suffix", "tsconfig.suffixes.json", "consumer.mts", "directory-main-extension-order", true},
		{"directory-main-trailing-slash", "tsconfig.json", "consumer.mts", "directory-main-trailing-slash", false},
		{"directory-main-ignores-nested-main", "tsconfig.json", "consumer.mts", "directory-main-nested-manifest", false},
		{"fallback-index-extension-before-suffix", "tsconfig.suffixes.json", "consumer.mts", "fallback-index-extension-order", true},
		{"implicit-main-ignores-extensionless-index", "tsconfig.json", "consumer.mts", "implicit-main-cjs", false},
		{"empty-main-ignores-extensionless-index", "tsconfig.json", "consumer.mts", "empty-main-cjs", false},
		{"null-main-ignores-extensionless-index", "tsconfig.json", "consumer.mts", "null-main-cjs", false},
		{"explicit-main-preserves-extensionless-index", "tsconfig.json", "consumer.mts", "explicit-main-extensionless", true},
		{"explicit-main-extensionless-commonjs", "tsconfig.json", "consumer.mts", "explicit-main-extensionless-cjs", false},
		{"implicit-main-suffixed-index", "tsconfig.suffixes-only.json", "consumer.mts", "implicit-main-suffix", false},
		{"implicit-main-empty-suffix-first", "tsconfig.suffixes-empty-first.json", "consumer.mts", "implicit-main-suffix", true},
		{"implicit-main-json-types-disabled", "tsconfig.no-json.json", "consumer.cts", "implicit-main-json", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code := "import obj from '" + tc.source + "'; obj.foo; const {foo} = obj;"
			var valid []rule_tester.ValidTestCase
			var invalid []rule_tester.InvalidTestCase
			if tc.warn {
				invalid = []rule_tester.InvalidTestCase{{Code: code, FileName: tc.fileName, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamedAsDefaultMember"}, {MessageId: "noNamedAsDefaultMember"}}}}
			} else {
				valid = []rule_tester.ValidTestCase{{Code: code, FileName: tc.fileName}}
			}
			rule_tester.RunRuleTester(root, tc.config, t, &no_named_as_default_member.NoNamedAsDefaultMemberRule, valid, invalid)
		})
	}
}

func TestNoNamedAsDefaultMemberDeclarationMainModes(t *testing.T) {
	root := defaultMemberRoot(t, "testdata/module-resolution-formats.txtar")
	// tsgo normalizes legacy Node10 to Bundler for this module setting.
	for _, config := range []string{"tsconfig.node10.json", "tsconfig.bundler.json", "tsconfig.preserve.json"} {
		t.Run(config, func(t *testing.T) {
			rule_tester.RunRuleTester(root, config, t, &no_named_as_default_member.NoNamedAsDefaultMemberRule,
				[]rule_tester.ValidTestCase{
					{Code: `import obj from 'declaration-main-only/feature.js'; obj.foo;`, FileName: "consumer.mts"},
					{Code: `import obj from 'declaration-main-exports-override'; obj.foo;`, FileName: "consumer.mts"},
				},
				[]rule_tester.InvalidTestCase{{Code: `import obj from 'declaration-main-only'; obj.foo;`, FileName: "consumer.mts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamedAsDefaultMember"}}}},
			)
		})
	}
	t.Run("declaration-main-with-module-suffixes", func(t *testing.T) {
		rule_tester.RunRuleTester(root, "tsconfig.suffixes-only.json", t, &no_named_as_default_member.NoNamedAsDefaultMemberRule, nil,
			[]rule_tester.InvalidTestCase{
				{Code: `import obj from 'declaration-main-suffix'; obj.foo;`, FileName: "consumer.mts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamedAsDefaultMember"}}},
				{Code: `import obj from 'declaration-main-native-only'; obj.foo;`, FileName: "consumer.mts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamedAsDefaultMember"}}},
			},
		)
	})
}

func TestNoNamedAsDefaultMemberSymlinkFormats(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/module-resolution-formats.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "symlink"))
	nodeModules := tspath.ResolvePath(directory, "node_modules")
	if err := os.MkdirAll(nodeModules, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"esm", "cjs", "declaration"} {
		if err := os.Symlink(tspath.ResolvePath(directory, "packages/"+name), tspath.ResolvePath(nodeModules, name)); err != nil {
			if runtime.GOOS == "windows" {
				t.Skipf("symlink creation is unavailable: %v", err)
			}
			t.Fatal(err)
		}
	}
	root := rule_tester.Root{Dir: directory, FS: bundled.WrapFS(osvfs.FS())}
	for _, config := range []string{"tsconfig.json", "tsconfig.preserve.json"} {
		t.Run(config, func(t *testing.T) {
			rule_tester.RunRuleTester(root, config, t, &no_named_as_default_member.NoNamedAsDefaultMemberRule,
				[]rule_tester.ValidTestCase{{Code: `import obj from 'cjs'; obj.foo; const {foo} = obj;`, FileName: "consumer.mts"}},
				[]rule_tester.InvalidTestCase{
					{Code: `import obj from 'esm'; obj.foo; const {foo} = obj;`, FileName: "consumer.mts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamedAsDefaultMember"}, {MessageId: "noNamedAsDefaultMember"}}},
					{Code: `import obj from 'declaration'; obj.foo;`, FileName: "consumer.mts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noNamedAsDefaultMember"}}},
				},
			)
		})
	}
}
