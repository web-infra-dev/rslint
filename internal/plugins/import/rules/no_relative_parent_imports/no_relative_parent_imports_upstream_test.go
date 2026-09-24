package no_relative_parent_imports_test

import (
	"fmt"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_relative_parent_imports"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const upstreamFile = "internal-modules/plugins/plugin2/index.js"

func parentRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-relative-parent-imports")
	archive := txtarfs.MustParseFile(t, "testdata/modules.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("fixture archive is empty")
	}
	files := make(map[string]string, len(names))
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(data)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	return root
}

func parentError(fileName, dependency string, line, column, endLine, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "",
		Message: fmt.Sprintf("Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `%s` to same directory as `%s` or consider making `%s` a package.",
			fileName, dependency, dependency),
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}}
}

func runParentTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	for i := range valid {
		if valid[i].FileName == "" {
			valid[i].FileName = upstreamFile
		}
	}
	for i := range invalid {
		if invalid[i].FileName == "" {
			invalid[i].FileName = upstreamFile
		}
	}
	rule_tester.RunRuleTester(parentRoot(t), "tsconfig.json", t, &no_relative_parent_imports.NoRelativeParentImportsRule, valid, invalid)
}

// All 12 valid and 6 invalid cases, with Babel's standard syntax parsed natively.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-relative-parent-imports.js
func TestNoRelativeParentImportsUpstream(t *testing.T) {
	runParentTests(t, []rule_tester.ValidTestCase{
		{Code: `import foo from "./internal.js"`},
		{Code: `import foo from "./app/index.js"`},
		{Code: `import foo from "package"`},
		{Code: `require("./internal.js")`, Options: map[string]any{"commonjs": true}},
		{Code: `require("./app/index.js")`, Options: map[string]any{"commonjs": true}},
		{Code: `require("package")`, Options: map[string]any{"commonjs": true}},
		{Code: `import("./internal.js")`},
		{Code: `import("./app/index.js")`},
		{Code: `import(".")`},
		{Code: `import("path")`},
		{Code: `import("package")`},
		{Code: `import("@scope/package")`},
	}, []rule_tester.InvalidTestCase{
		{Code: `import foo from "../plugin.js"`, Errors: parentError("index.js", "../plugin.js", 1, 17, 1, 31)},
		{Code: `require("../plugin.js")`, Options: map[string]any{"commonjs": true}, Errors: parentError("index.js", "../plugin.js", 1, 9, 1, 23)},
		{Code: `import("../plugin.js")`, Errors: parentError("index.js", "../plugin.js", 1, 8, 1, 22)},
		{Code: `import foo from "./../plugin.js"`, Errors: parentError("index.js", "./../plugin.js", 1, 17, 1, 33)},
		{Code: `import foo from "../../api/service"`, Errors: parentError("index.js", "../../api/service", 1, 17, 1, 36)},
		{Code: `import("../../api/service")`, Errors: parentError("index.js", "../../api/service", 1, 8, 1, 27)},
	})
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-relative-parent-imports.md
func TestNoRelativeParentImportsUpstreamDocs(t *testing.T) {
	runParentTests(t, []rule_tester.ValidTestCase{
		{FileName: "add.js", Code: `export default function (numbers) { return numbers.reduce((sum, n) => sum + n, 0); }`},
		{FileName: "numbers/three.js", Code: `export default function three(add) { return add([1, 2]); }`},
		{FileName: "use.js", Code: "import add from './add';\nimport three from './numbers/three';\nconsole.log(three(add));"},
		{FileName: "numbers/three.js", Code: "import add from 'add';\nexport default function three() { return add([1,2]); }"},
		{FileName: "main.js", Code: "import foo from 'foo';\nimport a from './lib/a';"},
		{FileName: "lib/a.js", Code: `import b from './b';`},
		// Both relocation strategies from the prose examples.
		{FileName: "three.js", Code: `import add from './add';`},
		{FileName: "three.js", Code: `import add from './math/add';`},
	}, []rule_tester.InvalidTestCase{
		{FileName: "numbers/three.js", Code: "import add from '../add';\nexport default function three() { return add([1, 2]); }", Errors: parentError("three.js", "../add", 1, 17, 1, 25)},
		{FileName: "lib/a.js", Code: `import bar from '../main';`, Errors: parentError("a.js", "../main", 1, 17, 1, 26)},
	})
}
