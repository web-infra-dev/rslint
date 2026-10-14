package no_relative_packages_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	target "github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_relative_packages"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func packageRoot(t *testing.T) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-relative-packages")
	archive := txtarfs.MustParseFile(t, "testdata/modules.txtar")
	names, err := archive.FileNames("")
	if err != nil || len(names) == 0 {
		t.Fatalf("empty or invalid fixtures: %v", err)
	}
	files := map[string]string{}
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

func packageCase(file, code, literal, imported, proposed string, options any) rule_tester.InvalidTestCase {
	start := strings.Index(code, literal)
	if start < 0 {
		panic("missing literal")
	}
	position := func(offset int) (int, int) {
		prefix := code[:offset]
		return strings.Count(prefix, "\n") + 1, len(utf16.Encode([]rune(prefix[strings.LastIndex(prefix, "\n")+1:]))) + 1
	}
	line, column := position(start)
	endLine, endColumn := position(start + len(literal))
	return rule_tester.InvalidTestCase{FileName: file, Code: code, Options: options,
		Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: fmt.Sprintf("Relative import from another package is not allowed. Use `%s` instead of `%s`", proposed, imported), Line: line, Column: column, EndLine: endLine, EndColumn: endColumn}},
		Output: []string{strings.Replace(code, literal, fmt.Sprintf("%q", proposed), 1)},
	}
}

func runPackages(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	rule_tester.RunRuleTester(packageRoot(t), "tsconfig.json", t, &target.NoRelativePackagesRule, valid, invalid)
}

// All seven valid and four invalid upstream cases, without skips.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-relative-packages.js
func TestNoRelativePackagesUpstream(t *testing.T) {
	runPackages(t, []rule_tester.ValidTestCase{{FileName: `tests/files/package/index.js`, Code: `import foo from "./index.js"`},
		{FileName: `tests/files/package/index.js`, Code: `import bar from "../bar"`},
		{FileName: `tests/files/package-named/index.js`, Code: `import {foo} from "a"`},
		{FileName: `tests/files/package/index.js`, Code: `const bar = require("../bar.js")`},
		{FileName: `tests/files/package/index.js`, Code: `const bar = require("../not/a/file/path.js")`},
		{FileName: `tests/files/package/index.js`, Code: `import "package"`},
		{FileName: `tests/files/package/index.js`, Code: `require("../bar.js")`}}, []rule_tester.InvalidTestCase{packageCase(`tests/files/bar.js`, `import foo from "./package-named"`, `"./package-named"`, `./package-named`, `package-named`, nil),
		packageCase(`tests/files/package/index.js`, `import foo from "../package-named"`, `"../package-named"`, `../package-named`, `package-named`, nil),
		packageCase(`tests/files/package/index.js`, `import foo from "../package-scoped"`, `"../package-scoped"`, `../package-scoped`, `@scope/package-named`, nil),
		packageCase(`tests/files/package-named/index.js`, `import bar from "../bar"`, `"../bar"`, `../bar`, `eslint-plugin-import/tests/files/bar`, nil)})
}

// The docs place foo at packages/foo.js in a comment, but their tree puts it
// at packages/foo/index.js; use the tree so sibling and parent imports resolve.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-relative-packages.md
func TestNoRelativePackagesDocs(t *testing.T) {
	runPackages(t, []rule_tester.ValidTestCase{{FileName: `my-project/packages/foo/index.js`, Code: `import bar from 'bar';`},
		{FileName: `my-project/entry.js`, Code: `import bar from 'bar';`}}, []rule_tester.InvalidTestCase{packageCase(`my-project/packages/foo/index.js`, `import bar from '../bar';`, `'../bar'`, `../bar`, `bar`, nil),
		packageCase(`my-project/packages/foo/index.js`, `import entry from '../../entry.js';`, `'../../entry.js'`, `../../entry.js`, `my-project/entry.js`, nil),
		packageCase(`my-project/entry.js`, `import bar from './packages/bar';`, `'./packages/bar'`, `./packages/bar`, `bar`, nil)})
}
