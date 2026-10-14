package no_extraneous_dependencies_test

import (
	"fmt"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	. "github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_extraneous_dependencies"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func missing(name string, line, column, endLine, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{MessageId: "", Message: fmt.Sprintf("'%s' should be listed in the project's dependencies. Run 'npm i -S %s' to add it", name, name), Line: line, Column: column, EndLine: endLine, EndColumn: endColumn}}
}
func TestExtras(t *testing.T) {
	root := dependencyRoot(t)
	files := map[string]string{}
	for name, text := range map[string]string{
		"broken/package.json":                 `{{`,
		"aliases/package.json":                `{"dependencies":{"alias":"*"}}`,
		"empty/package.json":                  `{}`,
		"one/package.json":                    `{"bundledDependencies":["@generated/foo","@generated/bar"]}`,
		"two/package.json":                    `{"bundleDependencies":{"react":"*"}}`,
		"node_modules/alias/package.json":     `{"name":"react","main":"index.js"}`,
		"node_modules/alias/index.js":         `module.exports={};`,
		"node_modules/alias/sub/package.json": `{"main":"index.js"}`,
		"node_modules/alias/sub/index.js":     `module.exports={};`,
	} {
		files[tspath.ResolvePath(root.Dir, name)] = text
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoExtraneousDependenciesRule, []rule_tester.ValidTestCase{
		{Code: `import("fs"); require(6); require.resolve("not-a-dependency"); require(` + "`not-a-dependency`" + `);`},
		{Code: `import "missing";`, Options: map[string]any{"includeInternal": true}},
		{Code: `import { type A, type B } from "not-a-dependency";`},
		{Code: `import "eslint";`, FileName: "tests/files/example.ts", Options: map[string]any{"packageDir": ""}},
		{Code: `import "eslint";`, FileName: "tests/files/example.ts", Options: map[string]any{"packageDir": []any{}}},
		{Code: `import "alias";`, Options: map[string]any{"packageDir": "aliases"}},
		{Code: `import "@generated/foo";`, Options: map[string]any{"packageDir": []any{"broken", "tests/files"}}},
		{Code: `import "@generated/bar";`, Options: map[string]any{"packageDir": []any{"one", "two"}}},
		{Code: `import "alias";`, Options: map[string]any{"packageDir": "tests/files/monorepo/packages/nested-package"}},
		{Code: `import "alias/sub";`, Options: map[string]any{"packageDir": "tests/files/monorepo/packages/nested-package"}},
		{Code: `import "eslint";`, Options: map[string]any{"devDependencies": false, "peerDependencies": true, "packageDir": "tests/files"}},
	}, []rule_tester.InvalidTestCase{
		{Code: `import("not-a-dependency")`, Errors: missing("not-a-dependency", 1, 1, 1, 27)},
		{Code: `(require)("not-a-dependency")`, Errors: missing("not-a-dependency", 1, 1, 1, 30)},
		{Code: "// 🙂\nimport {\n  value\n} from 'not-a-dependency';", Errors: missing("not-a-dependency", 2, 1, 4, 27)},
		{Code: `import "@generated/foo";`, Options: map[string]any{"packageDir": []any{"one", "two"}}, Errors: missing("@generated/foo", 1, 1, 1, 25)},
		{Code: `import "alias";`, Options: map[string]any{"packageDir": "empty"}, Errors: missing("react", 1, 1, 1, 16)},
		{Code: `import "not-a-dependency";`, Options: map[string]any{"includeInternal": true}, Settings: map[string]any{"import/internal-regex": "^not-a-dependency"}, Errors: missing("not-a-dependency", 1, 1, 1, 27)},
		{Code: `import {type A, B} from "not-a-dependency";`, Errors: missing("not-a-dependency", 1, 1, 1, 44)},
		{Code: `import type A from "not-a-dependency";`, Options: map[string]any{"includeTypes": true}, Errors: missing("not-a-dependency", 1, 1, 1, 39)},
		{Code: `import "eslint";`, Options: map[string]any{"packageDir": "tests/files", "devDependencies": []any{}, "peerDependencies": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "", Message: "'eslint' should be listed in the project's dependencies, not devDependencies.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}},
	})
}

func TestDependencyCategoryOptions(t *testing.T) {
	for _, field := range []string{"devDependencies", "optionalDependencies", "peerDependencies", "bundledDependencies"} {
		t.Run(field, func(t *testing.T) {
			root := dependencyRoot(t)
			manifest := fmt.Sprintf(`{"%s":{"react":"*"}}`, field)
			root.FS = utils.NewOverlayVFS(root.FS, map[string]string{tspath.ResolvePath(root.Dir, "category/package.json"): manifest})
			code := `import "react";`
			options := func(value any) map[string]any { return map[string]any{"packageDir": "category", field: value} }
			message := "'react' should be listed in the project's dependencies. Run 'npm i -S react' to add it"
			if field == "devDependencies" || field == "optionalDependencies" {
				message = fmt.Sprintf("'react' should be listed in the project's dependencies, not %s.", field)
			}
			errors := []rule_tester.InvalidTestCaseError{{MessageId: "", Message: message, Line: 1, Column: 1, EndLine: 1, EndColumn: 16}}
			rule_tester.RunRuleTester(root, "tsconfig.json", t, &NoExtraneousDependenciesRule, []rule_tester.ValidTestCase{
				{Code: code, Options: options(true)},
				{Code: code, FileName: "category/example.test.ts", Options: options([]any{"**/*.{test,spec}.ts"})},
			}, []rule_tester.InvalidTestCase{
				{Code: code, Options: options(false), Errors: errors},
				{Code: code, FileName: "category/example.ts", Options: options([]any{"**/*.test.ts"}), Errors: errors},
			})
		})
	}
}
