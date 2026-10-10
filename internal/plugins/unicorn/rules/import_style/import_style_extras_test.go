package import_style_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	import_style "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/import_style"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func extraOptions(module string, styles map[string]any) []any {
	return []any{map[string]any{
		"extendDefaultStyles": false,
		"checkExportFrom":     true,
		"styles":              map[string]any{module: styles},
	}}
}

func extraError(message, id string, line, column, endLine, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: id, Message: message,
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{},
	}}
}

func TestImportStyleExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// An empty named-import list does not erase an existing default binding.
		{Code: `import path, {} from 'node:path';`, FileName: "file.js"},
		// Parentheses are transparent in ESTree for the callee and module expression.
		{Code: `(require)(("free"));`, FileName: "file.js", Options: extraOptions("free", map[string]any{"unassigned": true})},
		// Authored TypeScript wrappers are visible to TSESTree and therefore are not direct require calls.
		{Code: `(require as any)('named');`, FileName: "file.ts", Tsx: false, Options: extraOptions("named", map[string]any{"unassigned": true})},
		// Optional calls do not match upstream's unassigned require shape.
		{Code: `require?.('named');`, FileName: "file.js", Options: extraOptions("named", map[string]any{"unassigned": false, "named": true})},
		// The assigned-require matcher also rejects optional calls.
		{Code: `const value = require?.('named');`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true})},
		// Stable local aliases are evaluated with scope/reference semantics.
		{Code: `const moduleName = 'named'; const {x} = require(moduleName);`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true})},
		// A reassigned alias is not constant and must not produce a false positive.
		{Code: `let moduleName = 'named'; moduleName = other; require(moduleName);`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true})},
		// Configured globals feed constant evaluation; disabling String prevents String.raw folding.
		{Code: "require(String.raw`util`);", FileName: "file.js", Globals: map[string]any{"String": "off"}},
		// A shadowed require is still matched by upstream's syntax-only helper.
		{Code: `function f(require) { const {x} = require('named'); }`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true})},
		// Computed literal binding keys are literals in ESTree, so they contribute no style.
		{Code: `const {['x']: x} = require('named');`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true})},
		// Plain export-all remains namespace style.
		{Code: `export * from 'module';`, FileName: "file.js", Options: extraOptions("module", map[string]any{"namespace": true})},
		// Espree keeps export namespace-from as ExportAllDeclaration, so it is namespace style too.
		{Code: `export * as ns from 'module';`, FileName: "file.js", Options: extraOptions("module", map[string]any{"namespace": true})},
		// TypeScript import-equals and import types have no upstream listener.
		{Code: `import value = require('named'); type T = import('named');`, FileName: "file.ts", Tsx: false, Options: extraOptions("named", map[string]any{"unassigned": true})},
	}

	invalid := []rule_tester.InvalidTestCase{
		// Complete range and multiline reporting for each reporting shape.
		{Code: `import value from 'named';`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true}), Output: []string{}, Errors: extraError("Use named import for module `named`.", "importStyle", 1, 1, 1, 27)},
		// Columns are UTF-16 code units: the astral prefix occupies two columns.
		{Code: `"😀"; import value from 'named';`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true}), Output: []string{}, Errors: extraError("Use named import for module `named`.", "importStyle", 1, 7, 1, 33)},
		{Code: "async () => {\n\tconst value = await import('named');\n}", FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true}), Errors: extraError("Use named import for module `named`.", "importStyle", 2, 8, 2, 37)},
		{Code: `import('named');`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true}), Errors: extraError("Use named import for module `named`.", "importStyle", 1, 1, 1, 16)},
		{Code: `const value = require('named');`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true}), Errors: extraError("Use named import for module `named`.", "importStyle", 1, 7, 1, 31)},
		{Code: `require('named');`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true}), Errors: extraError("Use named import for module `named`.", "importStyle", 1, 1, 1, 17)},
		// Computed identifier keys remain identifiers in ESTree and are named-style imports.
		{Code: `const {[key]: value} = require('chalk');`, FileName: "file.js", Errors: extraError("Use default import for module `chalk`.", "importStyle", 1, 7, 1, 40)},
		{Code: `async function f(){ const {[key]: value} = await import('chalk'); }`, FileName: "file.js", Errors: extraError("Use default import for module `chalk`.", "importStyle", 1, 27, 1, 65)},
		{Code: "require(String.raw`util`);", FileName: "file.js", Output: []string{}, Errors: extraError("Use named import for module `util`.", "importStyle", 1, 1, 1, 26)},
		{Code: `export * from 'named';`, FileName: "file.js", Options: extraOptions("named", map[string]any{"named": true}), Errors: extraError("Use named import for module `named`.", "importStyle", 1, 1, 1, 23)},

		// TypeScript type-only export-from still has an ESTree ExportNamedDeclaration and exported specifier.
		{Code: `export type {T} from 'module';`, FileName: "file.ts", Tsx: false, Options: extraOptions("module", map[string]any{"namespace": true}), Errors: extraError("Use namespace import for module `module`.", "importStyle", 1, 1, 1, 31)},
		// Type arguments do not wrap the ESTree callee, so generic require calls retain normal assignment semantics.
		{Code: `const value = require<string>('named');`, FileName: "file.ts", Tsx: false, Options: extraOptions("named", map[string]any{"unassigned": true}), Errors: extraError("Use unassigned import for module `named`.", "importStyle", 1, 7, 1, 39)},

		// The schema permits arbitrary style names. They remain allowed values and prevent banned classification.
		{Code: `import value from 'custom';`, FileName: "file.js", Options: extraOptions("custom", map[string]any{"unassigned": false, "default": false, "namespace": false, "named": false, "custom": true}), Errors: extraError("Use custom import for module `custom`.", "importStyle", 1, 1, 1, 28)},
		// Option maps cannot preserve JS insertion order. Rslint uses the documented deterministic order.
		{Code: `import 'ordered';`, FileName: "file.js", Options: extraOptions("ordered", map[string]any{"default": true, "named": true}), Errors: extraError("Use named or default import for module `ordered`.", "importStyle", 1, 1, 1, 18)},
		// Custom names follow known style names and are sorted for deterministic diagnostics.
		{Code: `import 'custom-order';`, FileName: "file.js", Options: extraOptions("custom-order", map[string]any{"zeta": true, "named": true, "alpha": true}), Errors: extraError("Use named, alpha, or zeta import for module `custom-order`.", "importStyle", 1, 1, 1, 23)},

		// Both message variants report the entire relevant node and never attach edits.
		{Code: `import value from 'banned';`, FileName: "file.js", Options: extraOptions("banned", map[string]any{"unassigned": false, "default": false, "namespace": false, "named": false}), Output: []string{}, Errors: extraError("All import styles are disabled for module `banned`. Use the `no-restricted-imports` rule to disallow a module.", "importStyleBanned", 1, 1, 1, 28)},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &import_style.ImportStyleRule, valid, invalid)
}
