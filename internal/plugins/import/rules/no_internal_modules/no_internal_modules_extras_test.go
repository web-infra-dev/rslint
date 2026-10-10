// cspell:ignore nternal
package no_internal_modules_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_internal_modules"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInternalModulesExtras(t *testing.T) {
	runInternalTests(t, []rule_tester.ValidTestCase{
		// Empty forbid.
		{Code: "import \"./plugin2/internal\"", Options: []any{map[string]any{"forbid": []any{}}}},
		// unresolved reaching is exempt by default.
		{Code: "import \"./missing/internal\""},
		// only resolved path allows reaching.
		{Code: "import \"./plugin2/internal\"", Options: []any{map[string]any{"allow": []any{"**/internal-modules/plugins/plugin2/internal.js"}}}},
		// absolute-style raw pattern.
		{Code: "import \"./plugin2/internal\"", Options: []any{map[string]any{"allow": []any{"/plugin2/internal"}}}},
		// dot and parent segment normalization.
		{Code: "import \"./plugin2/../plugin2//./internal\"", Options: []any{map[string]any{"allow": []any{"plugin2/internal"}}}},
		// dotfile wildcard excludes a dot.
		{Code: "import \"./plugin2/.private\"", Options: []any{map[string]any{"forbid": []any{"plugin2/*"}}}},
		// builtin and absolute types excluded.
		{Code: "import \"fs/promises\"; import \"node:fs\"; import \"/tmp/private/module\";", Options: []any{map[string]any{"forbid": []any{"**"}}}},
		// configured core-module subpath excluded.
		{Code: "import \"virtual-core/private\"", Options: []any{map[string]any{"forbid": []any{"**"}}}, Settings: map[string]any{"import/core-modules": []any{"virtual-core"}}},
		// unknown unresolved name excluded.
		{Code: "import \"?private/module\"", Options: []any{map[string]any{"forbid": []any{"**"}}}},
		// commonjs non-literal and member and extra arguments and amd ignored.
		{Code: "require(`./plugin2/internal`); require(name); obj.require(\"./plugin2/internal\"); require(\"./plugin2/internal\", true); define([\"./plugin2/internal\"],()=>{});"},
		// dynamic templates and expressions ignored.
		{Code: "import(`./plugin2/internal`); import(name);"},
		// import-equals and TS assertions ignored by moduleVisitor.
		{Code: "import api = require(\"./plugin2/internal\"); require(\"./plugin2/internal\" as string); (require as any)(\"./plugin2/internal\");", FileName: "internal-modules/plugins/plugin.ts"}}, []rule_tester.InvalidTestCase{
		// Explicit default options.
		{Code: "import \"./plugin2/internal\"", Options: []any{map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		// empty allow.
		{Code: "import \"./plugin2/internal\"", Options: []any{map[string]any{"allow": []any{}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		// unresolved reaching is still forbidden lexically.
		{Code: "import \"./missing/internal\"", Options: []any{map[string]any{"forbid": []any{"missing/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./missing/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		// only resolved path forbids reaching.
		{Code: "import \"./plugin2/internal\"", Options: []any{map[string]any{"forbid": []any{"**/internal-modules/plugins/plugin2/internal.js"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		// normalize backslash separators before glob matching.
		{Code: "import \".\\\\plugin2\\\\internal\"", Options: []any{map[string]any{"forbid": []any{"plugin2/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \".\\plugin2\\internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 30}}},
		// negated pattern.
		{Code: "import \"./plugin2/internal\"", Options: []any{map[string]any{"forbid": []any{"!plugin2/app/*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		// brace and extglob patterns.
		{Code: "import \"./plugin2/internal\"", Options: []any{map[string]any{"forbid": []any{"plugin{1,2}/@(internal|other)"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		// dotfile explicit pattern.
		{Code: "import \"./plugin2/.private\"", Options: []any{map[string]any{"forbid": []any{"plugin2/.*"}}}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/.private\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		// internal regex takes precedence over builtin type.
		{Code: "import \"fs/promises\"", Options: []any{map[string]any{"forbid": []any{"**"}}}, Settings: map[string]any{"import/internal-regex": "^fs/"}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"fs/promises\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
		// internal regex makes unknown name eligible.
		{Code: "import \"?private/module\"", Options: []any{map[string]any{"forbid": []any{"**"}}}, Settings: map[string]any{"import/internal-regex": "^\\?"}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"?private/module\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
		// raw glob checks precede resolution success.
		{Code: "import \"@/api/service\"", Options: []any{map[string]any{"forbid": []any{"**/api/*"}}}, Settings: map[string]any{"import/internal-regex": "^@/"}, Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"@/api/service\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 23}}},
		// commonjs literal and optional and shadowed calls.
		{Code: "require(\"./plugin2/internal\"); (require)((\"./plugin2/internal\")); require?.(\"./plugin2/internal\"); function f(require) {require(\"./plugin2/internal\");}", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 9, EndLine: 1, EndColumn: 29}, {Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 43, EndLine: 1, EndColumn: 63}, {Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 77, EndLine: 1, EndColumn: 97}, {Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 129, EndLine: 1, EndColumn: 149}}},
		// dynamic import literals and options.
		{Code: "import(\"./plugin2/internal\"); import((\"./plugin2/internal\"), {with:{type:\"json\"}});", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 28}, {Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 39, EndLine: 1, EndColumn: 59}}},
		// type-only import and export and namespace export.
		{Code: "import type {Value} from \"./plugin2/internal\"; export type {Value} from \"./plugin2/internal\"; export * as api from \"./plugin2/internal\";", FileName: "internal-modules/plugins/plugin.ts", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 26, EndLine: 1, EndColumn: 46}, {Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 73, EndLine: 1, EndColumn: 93}, {Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 116, EndLine: 1, EndColumn: 136}}},
		// non-ASCII text before literal uses utf16 columns.
		{Code: "const emoji = \"😀\"; import(\"./plugin2/internal\");", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 28, EndLine: 1, EndColumn: 48}}},
		// cooked literal value and raw diagnostic range.
		{Code: "import \"./plugin2/\\u0069nternal\";", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 1, EndColumn: 33}}},
		// multiline literal range.
		{Code: "import \"./plugin2/in\\\nternal\";", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 8, EndLine: 2, EndColumn: 8}}},
		// jsx expression contains module reference.
		{Code: "const view = <div>{import(\"./plugin2/internal\")}</div>;", FileName: "internal-modules/plugins/plugin.tsx", Errors: []rule_tester.InvalidTestCaseError{{Message: "Reaching to \"./plugin2/internal\" is not allowed.", Line: 1, Column: 27, EndLine: 1, EndColumn: 47}}}})
}

func TestNoInternalModulesSchema(t *testing.T) {
	for _, options := range [][]any{nil, {map[string]any{}}, {map[string]any{"allow": []any{"x", "x"}}}, {map[string]any{"forbid": []any{}}}} {
		if err := no_internal_modules.NoInternalModulesRule.Schema.Validate(options); err != nil {
			t.Errorf("valid options %v: %v", options, err)
		}
	}
	for _, options := range [][]any{{true}, {nil}, {map[string]any{"allow": true}}, {map[string]any{"forbid": []any{1}}}, {map[string]any{"allow": []any{}, "forbid": []any{}}}, {map[string]any{"extra": true}}, {map[string]any{}, map[string]any{}}} {
		if err := no_internal_modules.NoInternalModulesRule.Schema.Validate(options); err == nil {
			t.Errorf("accepted invalid options %v", options)
		}
	}
}

func TestNoInternalModulesResolverAndInvalidPatterns(t *testing.T) {
	typescript := map[string]any{"import/resolver": "typescript"}
	runInternalTests(t, []rule_tester.ValidTestCase{
		// Scope-like segments do not count toward the default depth check.
		{Code: `import "@alias/internal";`, Settings: typescript},
		{Code: `import "alias/internal";`},
		{Code: `import "alias/internal";`, Settings: typescript, Options: map[string]any{"allow": []any{"**/plugin2/internal.js"}}},
		{Code: `import "./plugin2/internal";`, Options: map[string]any{"forbid": []any{"", "#comment"}}},
		{Code: `/** @type {import("./plugin2/internal").Value} */ let value;`},
	}, []rule_tester.InvalidTestCase{
		{Code: `import "alias/internal";`, Settings: typescript, Errors: []rule_tester.InvalidTestCaseError{{Message: `Reaching to "alias/internal" is not allowed.`, Line: 1, Column: 8, EndLine: 1, EndColumn: 24}}},
		{Code: `export { default as getUser } from '../actions/getUser';`, FileName: "my-project/entry.js", Options: map[string]any{"forbid": []any{"**/actions/*"}}, Errors: []rule_tester.InvalidTestCaseError{{Message: `Reaching to "../actions/getUser" is not allowed.`, Line: 1, Column: 36, EndLine: 1, EndColumn: 56}}},
		{Code: `import "./plugin2/internal";`, Options: map[string]any{"allow": []any{"", "#comment"}}, Errors: []rule_tester.InvalidTestCaseError{{Message: `Reaching to "./plugin2/internal" is not allowed.`, Line: 1, Column: 8, EndLine: 1, EndColumn: 28}}},
		{Code: "import './plugin2/internal';\nimport './plugin2/app/index';", Settings: map[string]any{"import/resolver": "missing-resolver"}, Errors: []rule_tester.InvalidTestCaseError{{Message: `Resolve error: unable to load resolver "missing-resolver".`, Line: 1, Column: 1, EndLine: 1, EndColumn: 1}}},
	})
}
