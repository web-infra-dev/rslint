package no_named_as_default_member_test

import (
	"fmt"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_as_default_member"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func defaultMemberRoot(t *testing.T, archiveName string) rule_tester.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = tspath.ResolvePath(root.Dir, "no-named-as-default-member-rule")
	archive := txtarfs.MustParseFile(t, archiveName)
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
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
func defaultMemberError(object, property, source string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noNamedAsDefaultMember",
		Message:   fmt.Sprintf("Caution: `%s` also has a named export `%s`. Check if you meant to write `import {%s} from '%s'` instead.", object, property, property, source),
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from eslint-plugin-import v2.32.0 tests/src/rules/no-named-as-default-member.js
// and docs/rules/no-named-as-default-member.md, including expanded SYNTAX_CASES.
// Upstream uses literal messages without message IDs; positions were checked against that version.
func TestNoNamedAsDefaultMemberUpstream(t *testing.T) {
	rule_tester.RunRuleTester(defaultMemberRoot(t, "testdata/upstream.txtar"), "tsconfig.json", t, &no_named_as_default_member.NoNamedAsDefaultMemberRule,
		[]rule_tester.ValidTestCase{
			// upstream
			{Code: `import bar, {foo} from "./bar";`},
			{Code: `import bar from "./bar"; const baz = bar.baz`},
			{Code: `import {foo} from "./bar"; const baz = foo.baz;`},
			{Code: `import * as named from "./named-exports"; const a = named.a`},
			{Code: `import foo from "./default-export-default-property"; const a = foo.default`},
			// Arbitrary module namespace identifier names
			{Code: `import bar, { foo } from "./export-default-string-and-named"`},
			// SYNTAX_CASES
			{Code: `for (let { foo, bar } of baz) {}`},
			{Code: `for (let [ foo, bar ] of baz) {}`},
			{Code: `const { x, y } = bar`},
			{Code: `const { x, y, ...z } = bar`},
			{Code: `let x; export { x }`},
			{Code: `let x; export { x as y }`},
			{Code: `export const x = null`},
			{Code: `export var x = null`},
			{Code: `export let x = null`},
			{Code: `export default x`},
			{Code: `export default class x {}`},
			{Code: `import json from "./data.json"`, Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}}},
			{Code: `import foo from "./foobar.json";`, Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}}},
			{Code: `import foo from "./foobar";`, Settings: map[string]interface{}{"import/extensions": []interface{}{".js"}}},
			{Code: `import { foo } from "./issue-370-commonjs-namespace/bar"`, Settings: map[string]interface{}{"import/ignore": []interface{}{"foo"}}},
			{Code: `export * from "./issue-370-commonjs-namespace/bar"`, Settings: map[string]interface{}{"import/ignore": []interface{}{"foo"}}},
			{Code: `import * as a from "./commonjs-namespace/a"; a.b`},
			{Code: `import { foo } from "./ignore.invalid.extension"`},
			// Documentation module and valid consumer
			{Code: `export default 'foo'; export const bar = 'baz';`},
			{Code: `import foo, {bar} from './foo.js';`},
		},
		[]rule_tester.InvalidTestCase{
			// upstream
			{Code: `import bar from "./bar"; const foo = bar.foo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("bar", "foo", "./bar", 1, 38, 1, 45)}},
			{Code: `import bar from "./bar"; bar.foo();`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("bar", "foo", "./bar", 1, 26, 1, 33)}},
			{Code: `import bar from "./bar"; const {foo} = bar;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("bar", "foo", "./bar", 1, 33, 1, 36)}},
			{Code: `import bar from "./bar"; const {foo: foo2, baz} = bar;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("bar", "foo", "./bar", 1, 33, 1, 36)}},
			// Arbitrary module namespace identifier names
			{Code: `import bar from "./export-default-string-and-named"; const foo = bar.foo;`, Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("bar", "foo", "./export-default-string-and-named", 1, 66, 1, 73)}},
			// Documentation member access
			{Code: "import foo from './foo.js';\nconst bar = foo.bar;", Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("foo", "bar", "./foo.js", 2, 13, 2, 20)}},
			// Documentation destructuring
			{Code: "import foo from './foo.js';\nconst {bar} = foo;", Errors: []rule_tester.InvalidTestCaseError{defaultMemberError("foo", "bar", "./foo.js", 2, 8, 2, 11)}},
		},
	)
}
