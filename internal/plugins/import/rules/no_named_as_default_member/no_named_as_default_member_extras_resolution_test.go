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
