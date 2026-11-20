package no_unused_modules

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Cases mirror eslint-plugin-import v2.32.0:
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-unused-modules.js
//
// Upstream keeps one module-level usage index across all of these test runs
// and updates it from each linted file, so many cases depend on the text an
// earlier case linted into another file. rslint builds the index from one
// source generation instead; each group passes the earlier file text it
// depends on as an overlay. Upstream asserts messages only; positions below
// are TSESTree's.

var (
	missingExportsOptions = map[string]any{"missingExports": true}

	unusedExportsOptions = map[string]any{
		"unusedExports": true,
		"src":           []any{fixturePath("./no-unused-modules/**/*.js")},
		"ignoreExports": []any{fixturePath("./no-unused-modules/*ignored*.js")},
	}

	unusedExportsTypescriptOptions = map[string]any{
		"unusedExports": true,
		"src":           []any{fixturePath("./no-unused-modules/typescript")},
	}

	unusedExportsTypescriptIgnoreUnusedTypesOptions = map[string]any{
		"unusedExports":           true,
		"ignoreUnusedTypeExports": true,
		"src":                     []any{fixturePath("./no-unused-modules/typescript")},
	}

	unusedExportsJsxOptions = map[string]any{
		"unusedExports": true,
		"src":           []any{fixturePath("./no-unused-modules/jsx")},
	}

	// config/typescript.js
	typescriptSettings = map[string]any{
		"import/extensions":              []any{".ts", ".cts", ".mts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"},
		"import/external-module-folders": []any{"node_modules", "node_modules/@types"},
		"import/parsers":                 map[string]any{"@typescript-eslint/parser": []any{".ts", ".cts", ".mts", ".tsx"}},
		"import/resolver": map[string]any{"node": map[string]any{
			"extensions": []any{".ts", ".cts", ".mts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"},
		}},
	}

	// config/react.js
	jsxSettings = map[string]any{"import/extensions": []any{".js", ".jsx"}}
)

// file0Code is file-0.js as the "tests for exports" invalid case lints it.
// Later upstream runs see the index it leaves behind: unlike file-0.js on
// disk, it no longer imports `q` from file-q.js.
const file0Code = `
        import eslint from 'eslint'
        import fileA from './file-a'
        import { b } from './file-b'
        import { c1, c2 } from './file-c'
        import { d } from './file-d'
        import { e } from './file-e'
        import { e2 } from './file-e'
        import { h2 } from './file-h'
        import * as l from './file-l'
        export * from './file-n'
        export { default, o0, o3 } from './file-o'
        export { p } from './file-p'
        import s from './file-s'
      `

func TestNoUnusedModulesUpstream(t *testing.T) {
	file0 := fixturePath("./no-unused-modules/file-0.js")

	// tests for missing exports
	runGroup(t, "missing exports", nil, []rule_tester.ValidTestCase{
		{Code: "export default function noOptions() {}", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "export default () => 1", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "export const a = 1", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "const a = 1; export { a }", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "function a() { return true }; export { a }", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "const a = 1; const b = 2; export { a, b }", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "const a = 1; export default a", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "export class Foo {}", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "export const [foobar] = [];", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "export const [foobar] = foobarFactory();", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: `
        export default function NewComponent () {
          return 'I am new component'
        }
      `, FileName: fixturePath("missing.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: missingExportsOptions, Code: "const a = 1", FileName: fixturePath("missing.js"), Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 12)}},
		// A comment-only Program has an empty range after its comments.
		{Options: missingExportsOptions, Code: "/* const a = 1 */", FileName: fixturePath("missing.js"), Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 18, 1, 18)}},
	})

	// tests for exports
	runGroup(t, "exports", nil, []rule_tester.ValidTestCase{
		// Upstream parses these with babel-eslint; the syntax is plain ESM.
		{Options: unusedExportsOptions, Code: `import { o2 } from "./file-o";export default () => 12`, FileName: fixturePath("./no-unused-modules/file-a.js")},
		{Options: unusedExportsOptions, Code: "export const b = 2", FileName: fixturePath("./no-unused-modules/file-b.js")},
		{Options: unusedExportsOptions, Code: "const c1 = 3; function c2() { return 3 }; export { c1, c2 }", FileName: fixturePath("./no-unused-modules/file-c.js")},
		{Options: unusedExportsOptions, Code: "export function d() { return 4 }", FileName: fixturePath("./no-unused-modules/file-d.js")},
		{Options: unusedExportsOptions, Code: "export class q { q0() {} }", FileName: fixturePath("./no-unused-modules/file-q.js")},
		{Options: unusedExportsOptions, Code: "const e0 = 5; export { e0 as e }", FileName: fixturePath("./no-unused-modules/file-e.js")},
		{Options: unusedExportsOptions, Code: "const l0 = 5; const l = 10; export { l0 as l1, l }; export default () => {}", FileName: fixturePath("./no-unused-modules/file-l.js")},
		{Options: unusedExportsOptions, Code: "const o0 = 0; const o1 = 1; export { o0, o1 as o2 }; export default () => {}", FileName: fixturePath("./no-unused-modules/file-o.js")},
		{Options: unusedExportsOptions, Code: `
        export const [o0, o2] = createLoadingAndErrorSelectors(
          AUTH_USER
        );
      `, FileName: fixturePath("./no-unused-modules/file-o.js")},
	}, []rule_tester.InvalidTestCase{
		{
			Options: unusedExportsOptions, Code: file0Code, FileName: file0,
			Errors: []rule_tester.InvalidTestCaseError{
				unusedRange("default", 12, 18, 12, 25),
				unusedRange("o0", 12, 27, 12, 29),
				unusedRange("o3", 12, 31, 12, 33),
				unusedRange("p", 13, 18, 13, 19),
			},
		},
		{
			Options: unusedExportsOptions, Code: `const n0 = 'n0'; const n1 = 42; export { n0, n1 }; export default () => {}`,
			FileName: fixturePath("./no-unused-modules/file-n.js"),
			Errors:   []rule_tester.InvalidTestCaseError{unusedRange("default", 1, 52, 1, 75)},
		},
	})

	// test for unused exports
	afterFile0 := map[string]string{file0: file0Code}
	runGroup(t, "unused exports", afterFile0, nil, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export default () => 13", FileName: fixturePath("./no-unused-modules/file-f.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("default", 1, 1, 1, 24)}},
		{Options: unusedExportsOptions, Code: "export const g = 2", FileName: fixturePath("./no-unused-modules/file-g.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("g", 1, 1, 1, 19)}},
		{Options: unusedExportsOptions, Code: "const h1 = 3; function h2() { return 3 }; const h3 = true; export { h1, h2, h3 }", FileName: fixturePath("./no-unused-modules/file-h.js"), Errors: []rule_tester.InvalidTestCaseError{unusedError("h1", 1, 69)}},
		{Options: unusedExportsOptions, Code: "const i1 = 3; function i2() { return 3 }; export { i1, i2 }", FileName: fixturePath("./no-unused-modules/file-i.js"), Errors: []rule_tester.InvalidTestCaseError{unusedError("i1", 1, 52), unusedError("i2", 1, 56)}},
		{Options: unusedExportsOptions, Code: "export function j() { return 4 }", FileName: fixturePath("./no-unused-modules/file-j.js"), Errors: []rule_tester.InvalidTestCaseError{unusedError("j", 1, 1)}},
		{Options: unusedExportsOptions, Code: "export class q { q0() {} }", FileName: fixturePath("./no-unused-modules/file-q.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("q", 1, 1, 1, 27)}},
		{Options: unusedExportsOptions, Code: "const k0 = 5; export { k0 as k }", FileName: fixturePath("./no-unused-modules/file-k.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("k", 1, 24, 1, 31)}},
	})

	// dynamic imports
	runGroup(t, "dynamic imports", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `
            export const a = 10
            export const b = 20
            export const c = 30
            const d = 40
            export default d
            `, FileName: fixturePath("./no-unused-modules/exports-for-dynamic-js.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: `
        export const a = 10
        export const b = 20
        export const c = 30
        const d = 40
        export default d
        `, FileName: fixturePath("./no-unused-modules/exports-for-dynamic-js-2.js"), Errors: []rule_tester.InvalidTestCaseError{
			unusedError("a", 2, 9), unusedError("b", 3, 9), unusedError("c", 4, 9), unusedError("default", 6, 9),
		}},
	})
	runGroup(t, "dynamic imports typescript", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `
            export const ts_a = 10
            export const ts_b = 20
            export const ts_c = 30
            const ts_d = 40
            export default ts_d
            `, FileName: fixturePath("./no-unused-modules/typescript/exports-for-dynamic-ts.ts")},
		{
			Code:     "\n        import App from './App';\n      ",
			FileName: fixturePath("./unused-modules-reexport-crash/src/index.tsx"),
			Settings: typescriptSettings,
			Options:  map[string]any{"unusedExports": true, "ignoreExports": []any{"**/magic/**"}},
		},
	}, nil)

	// test for export from
	runGroup(t, "export from", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `export { default } from './file-o'`, FileName: fixturePath("./no-unused-modules/file-s.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export { k } from '" + fixturePath("./no-unused-modules/file-k.js") + "'", FileName: fixturePath("./no-unused-modules/file-j.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("k", 1, 10, 1, 11)}},
	})
	runGroup(t, "export from usage", map[string]string{
		fixturePath("./no-unused-modules/file-j.js"): "export { k } from '" + fixturePath("./no-unused-modules/file-k.js") + "'",
	}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: "const k0 = 5; export { k0 as k }", FileName: fixturePath("./no-unused-modules/file-k.js")},
	}, nil)

	// test for ignored files
	runGroup(t, "ignored files", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: "export default () => 14", FileName: fixturePath("./no-unused-modules/file-ignored-a.js")},
		{Options: unusedExportsOptions, Code: "export const b = 2", FileName: fixturePath("./no-unused-modules/file-ignored-b.js")},
		{Options: unusedExportsOptions, Code: "const c1 = 3; function c2() { return 3 }; export { c1, c2 }", FileName: fixturePath("./no-unused-modules/file-ignored-c.js")},
		{Options: unusedExportsOptions, Code: "export function d() { return 4 }", FileName: fixturePath("./no-unused-modules/file-ignored-d.js")},
		{Options: unusedExportsOptions, Code: "const f = 5; export { f as e }", FileName: fixturePath("./no-unused-modules/file-ignored-e.js")},
		{Options: unusedExportsOptions, Code: "const l0 = 5; const l = 10; export { l0 as l1, l }; export default () => {}", FileName: fixturePath("./no-unused-modules/file-ignored-l.js")},
	}, nil)

	importer := func(code string) map[string]string { return map[string]string{file0: code} }
	fileF := fixturePath("./no-unused-modules/file-f.js")
	fileG := fixturePath("./no-unused-modules/file-g.js")

	// add named import for file with default export
	namedF := "import { f } from '" + fileF + "'"
	runGroup(t, "add named import for file with default export", importer(namedF), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: namedF, FileName: file0},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export default () => 15", FileName: fileF, Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 1)}},
	})

	// add default import for file with default export
	defaultF := "import f from '" + fileF + "'"
	runGroup(t, "add default import for file with default export", importer(defaultF), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: defaultF, FileName: file0},
		{Options: unusedExportsOptions, Code: "export default () => 16", FileName: fileF},
	}, nil)

	// add default import for file with named export
	defaultG := "import g from '" + fileG + "';import {h} from '" + fixturePath("./no-unused-modules/file-gg.js") + "'"
	runGroup(t, "add default import for file with named export", importer(defaultG), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: defaultG, FileName: file0},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export const g = 2", FileName: fileG, Errors: []rule_tester.InvalidTestCaseError{unusedError("g", 1, 1)}},
	})

	// add named import for file with named export
	namedG := "import { g } from '" + fileG + "'; import eslint from 'eslint'"
	runGroup(t, "add named import for file with named export", importer(namedG), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: namedG, FileName: file0},
		{Options: unusedExportsOptions, Code: "export const g = 2", FileName: fileG},
	}, nil)

	// add different named import for file with named export
	namedC := "import { c } from '" + fixturePath("./no-unused-modules/file-b.js") + "'"
	runGroup(t, "add different named import for file with named export", importer(namedC), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: namedC, FileName: file0},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export const b = 2", FileName: fixturePath("./no-unused-modules/file-b.js"), Errors: []rule_tester.InvalidTestCaseError{unusedError("b", 1, 1)}},
	})

	// add renamed named import for file with named export
	renamedG := "import { g as g1 } from '" + fileG + "'; import eslint from 'eslint'"
	runGroup(t, "add renamed named import for file with named export", importer(renamedG), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: renamedG, FileName: file0},
		{Options: unusedExportsOptions, Code: "export const g = 2", FileName: fileG},
	}, nil)

	// add different renamed named import for file with named export
	renamedG1 := "import { g1 as g } from '" + fileG + "'"
	runGroup(t, "add different renamed named import for file with named export", importer(renamedG1), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: renamedG1, FileName: file0},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export const g = 2", FileName: fileG, Errors: []rule_tester.InvalidTestCaseError{unusedError("g", 1, 1)}},
	})

	// remove default import for file with default export
	namedA := "import { a1, a2 } from '" + fixturePath("./no-unused-modules/file-a.js") + "'"
	runGroup(t, "remove default import for file with default export", importer(namedA), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: namedA, FileName: file0},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export default () => 17", FileName: fixturePath("./no-unused-modules/file-a.js"), Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 1)}},
	})

	fileM := fixturePath("./no-unused-modules/file-m.js")
	fileMCode := "const m0 = 5; const m = 10; export { m0 as m1, m }; export default () => {}"
	allMUnused := []rule_tester.InvalidTestCaseError{unusedError("m1", 1, 38), unusedError("m", 1, 48), unusedError("default", 1, 53)}

	// add namespace import for file with unused exports
	runGroup(t, "add namespace import for file with unused exports", nil, nil, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: fileMCode, FileName: fileM, Errors: allMUnused},
	})
	namespaceM := "import * as m from '" + fileM + "'; import unknown from 'unknown-module'"
	runGroup(t, "namespace import", importer(namespaceM), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: namespaceM, FileName: file0},
		{Options: unusedExportsOptions, Code: fileMCode, FileName: fileM},
	}, nil)

	// remove all exports
	commentedM := "/* import * as m from '" + fileM + "' */"
	runGroup(t, "remove all exports", importer(commentedM), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: commentedM, FileName: file0},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: fileMCode, FileName: fileM, Errors: allMUnused},
	})

	starM := "export * from '" + fileM + "';"
	runGroup(t, "export star", importer(starM), []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: starM, FileName: file0},
	}, nil)
	runGroup(t, "export star usage", importer(starM), nil, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: fileMCode, FileName: fileM, Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 53)}},
	})

	namedReexportM := "export { m1, m} from '" + fileM + "';"
	runGroup(t, "named re-export", importer(namedReexportM), nil, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: namedReexportM, FileName: file0, Errors: []rule_tester.InvalidTestCaseError{unusedError("m1", 1, 10), unusedError("m", 1, 14)}},
		{Options: unusedExportsOptions, Code: fileMCode, FileName: fileM, Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 53)}},
	})

	// The upstream TODO valid case (`export { default, m1 } from` used by
	// nobody) is commented out upstream as well.
	defaultReexportM := "export { default, m1 } from '" + fileM + "';"
	runGroup(t, "default re-export", importer(defaultReexportM), nil, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: defaultReexportM, FileName: file0, Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 10), unusedError("m1", 1, 19)}},
		{Options: unusedExportsOptions, Code: fileMCode, FileName: fileM, Errors: []rule_tester.InvalidTestCaseError{unusedError("m", 1, 48)}},
	})

	// Test that import and export in the same file both counts as usage
	runGroup(t, "import and export in the same file", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `export const a = 5;export const b = 't1'`, FileName: fixturePath("./no-unused-modules/import-export-1.js")},
	}, nil)

	// renameDefault
	runGroup(t, "renameDefault", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `export { default as Component } from "./Component"`, FileName: fixturePath("./no-unused-modules/renameDefault/components.js")},
		{Options: unusedExportsOptions, Code: "export default function Component() {}", FileName: fixturePath("./no-unused-modules/renameDefault/Component.js")},
	}, nil)
	runGroup(t, "renameDefault-2", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `export { default as ComponentA } from "./ComponentA";export { default as ComponentB } from "./ComponentB";`, FileName: fixturePath("./no-unused-modules/renameDefault-2/components.js")},
		{Options: unusedExportsOptions, Code: "export default function ComponentA() {};", FileName: fixturePath("./no-unused-modules/renameDefault-2/ComponentA.js")},
	}, nil)

	// test behavior for new file: upstream writes file-added-*.js to disk
	// before these runs and removes them afterwards.
	added0 := fixturePath("./no-unused-modules/file-added-0.js")
	added1 := fixturePath("./no-unused-modules/file-added-1.js")
	namespaceMInAdded := "import * as m from '" + fileM + "'"
	runGroup(t, "add import in newly created file", map[string]string{added0: namespaceMInAdded}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: namespaceMInAdded, FileName: added0},
		{Options: unusedExportsOptions, Code: fileMCode, FileName: fileM},
	}, nil)
	runGroup(t, "add export for newly created file", nil, nil, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export default () => {2}", FileName: added0, Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 1)}},
	})
	defaultAdded0 := "import def from '" + added0 + "'"
	runGroup(t, "default import of newly created file", map[string]string{file0: defaultAdded0, added0: ""}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: defaultAdded0, FileName: file0},
		{Options: unusedExportsOptions, Code: "export default () => {}", FileName: added0},
	}, nil)

	// export * only considers named imports. default imports still need to be reported
	starAdded0 := "export * from '" + added0 + "'"
	runGroup(t, "export star of newly created file", map[string]string{file0: starAdded0, added0: ""}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: starAdded0, FileName: file0},
		// Test export * from 'external-compiled-library'
		{Options: unusedExportsOptions, Code: `export * from 'external-compiled-library'`, FileName: fixturePath("./no-unused-modules/file-r.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: `export const z = 'z';export default () => {}`, FileName: added0, Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 22)}},
	})
	runGroup(t, "named export of newly created file under export star", map[string]string{file0: starAdded0}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `export const a = 2`, FileName: added0},
	}, nil)

	// remove export *. all exports need to be reported
	namedAdded0 := "export { a } from '" + added0 + "'"
	runGroup(t, "remove export star", map[string]string{file0: namedAdded0, added0: "export const a = 2"}, nil, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: namedAdded0, FileName: file0, Errors: []rule_tester.InvalidTestCaseError{unusedError("a", 1, 10)}},
		{Options: unusedExportsOptions, Code: `export const z = 'z';export default () => {}`, FileName: added0, Errors: []rule_tester.InvalidTestCaseError{unusedError("z", 1, 1), unusedError("default", 1, 22)}},
	})

	starAdded1 := "export * from '" + added1 + "'"
	runGroup(t, "export star of second newly created file", map[string]string{file0: starAdded1, added1: ""}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: starAdded1, FileName: file0},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: `export const z = 'z';export default () => {}`, FileName: added1, Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 22)}},
	})

	added2 := fixturePath("./no-unused-modules/file-added-2.js")
	defaultAdded2 := "import added from '" + added2 + "'"
	runGroup(t, "default import between newly created files", map[string]string{added1: defaultAdded2, added2: ""}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: defaultAdded2, FileName: added1},
		{Options: unusedExportsOptions, Code: `export default () => {}`, FileName: added2},
	}, nil)

	added3 := fixturePath("./no-unused-modules/file-added-3.js")
	namedAdded3 := "import { added } from '" + added3 + "'"
	runGroup(t, "named import between newly created files", map[string]string{added1: namedAdded3, added3: ""}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: namedAdded3, FileName: added1},
		{Options: unusedExportsOptions, Code: `export const added = () => {}`, FileName: added3},
	}, nil)

	// test behavior for destructured exports
	runGroup(t, "destructured exports", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: "import { destructured } from '" + fixturePath("./no-unused-modules/file-destructured-1.js") + "'", FileName: fixturePath("./no-unused-modules/file-destructured-2.js")},
		{Options: unusedExportsOptions, Code: `export const { destructured } = {};`, FileName: fixturePath("./no-unused-modules/file-destructured-1.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: `export const { destructured2 } = {};`, FileName: fixturePath("./no-unused-modules/file-destructured-1.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("destructured2", 1, 1, 1, 37)}},
	})

	added4 := fixturePath("./no-unused-modules/file-added-4.js.js")
	namespaceAdded4 := "import * as added from '" + added4 + "'"
	runGroup(t, "namespace import of file with a double extension", map[string]string{added1: namespaceAdded4, added4: ""}, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: namespaceAdded4, FileName: added1},
		{Options: unusedExportsOptions, Code: `export const added = () => {}; export default () => {}`, FileName: added4},
	}, nil)

	// do not report missing export for ignored file
	runGroup(t, "missing export for ignored file", nil, []rule_tester.ValidTestCase{
		{Options: map[string]any{
			"src":            []any{fixturePath("./no-unused-modules/**/*.js")},
			"ignoreExports":  []any{fixturePath("./no-unused-modules/*ignored*.js")},
			"missingExports": true,
		}, Code: "export const test = true", FileName: fixturePath("./no-unused-modules/file-ignored-a.js")},
	}, nil)

	// lint file not available in `src`
	runGroup(t, "file outside src", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `export const jsxFoo = 'foo'; export const jsxBar = 'bar'`, FileName: fixturePath("./jsx/named.jsx")},
	}, nil)

	// do not report unused export for files mentioned in package.json
	runGroup(t, "package.json entry points", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `export const bin = "bin"`, FileName: fixturePath("./no-unused-modules/bin.js")},
		{Options: unusedExportsOptions, Code: `export const binObject = "binObject"`, FileName: fixturePath("./no-unused-modules/binObject/index.js")},
		{Options: unusedExportsOptions, Code: `export const browser = "browser"`, FileName: fixturePath("./no-unused-modules/browser.js")},
		{Options: unusedExportsOptions, Code: `export const browserObject = "browserObject"`, FileName: fixturePath("./no-unused-modules/browserObject/index.js")},
		{Options: unusedExportsOptions, Code: `export const main = "main"`, FileName: fixturePath("./no-unused-modules/main/index.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: `export const privatePkg = "privatePkg"`, FileName: fixturePath("./no-unused-modules/privatePkg/index.js"), Errors: []rule_tester.InvalidTestCaseError{unusedError("privatePkg", 1, 1)}},
	})

	// Avoid errors if re-export all from umd compiled library
	runGroup(t, "re-export all from umd compiled library", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: "export * from '" + fixturePath("./no-unused-modules/bin.js") + "'", FileName: fixturePath("./no-unused-modules/main/index.js")},
	}, nil)

	// TypeScript: upstream repeats this block for every TypeScript parser.
	typescriptFile := func(name string) string { return fixturePath("./no-unused-modules/typescript/" + name) }
	runGroup(t, "typescript", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `
          import {b} from './file-ts-b';
          import {c} from './file-ts-c';
          import {d} from './file-ts-d';
          import {e} from './file-ts-e';

          const a = b + 1 + e.f;
          const a2: c = {};
          const a3: d = {};
          `, FileName: typescriptFile("file-ts-a.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export const b = 2;`, FileName: typescriptFile("file-ts-b.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export interface c {};`, FileName: typescriptFile("file-ts-c.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export type d = {};`, FileName: typescriptFile("file-ts-d.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export enum e { f };`, FileName: typescriptFile("file-ts-e.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `
          import type {b} from './file-ts-b-used-as-type';
          import type {c} from './file-ts-c-used-as-type';
          import type {d} from './file-ts-d-used-as-type';
          import type {e} from './file-ts-e-used-as-type';

          const a: typeof b = 2;
          const a2: c = {};
          const a3: d = {};
          const a4: typeof e = undefined;
          `, FileName: typescriptFile("file-ts-a-import-type.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export const b = 2;`, FileName: typescriptFile("file-ts-b-used-as-type.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export interface c {};`, FileName: typescriptFile("file-ts-c-used-as-type.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export type d = {};`, FileName: typescriptFile("file-ts-d-used-as-type.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export enum e { f };`, FileName: typescriptFile("file-ts-e-used-as-type.ts")},
		// Should also be valid when the exporting files are linted before the importing ones
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export interface g {}`, FileName: typescriptFile("file-ts-g.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `import {g} from './file-ts-g';`, FileName: typescriptFile("file-ts-f.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export interface g {}; /* used-as-type */`, FileName: typescriptFile("file-ts-g-used-as-type.ts")},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `import type {g} from './file-ts-g';`, FileName: typescriptFile("file-ts-f-import-type.ts")},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export const b = 2;`, FileName: typescriptFile("file-ts-b-unused.ts"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("b", 1, 1, 1, 20)}},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export interface c {};`, FileName: typescriptFile("file-ts-c-unused.ts"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("c", 1, 1, 1, 22)}},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export type d = {};`, FileName: typescriptFile("file-ts-d-unused.ts"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("d", 1, 1, 1, 20)}},
		{Options: unusedExportsTypescriptOptions, Settings: typescriptSettings, Code: `export enum e { f };`, FileName: typescriptFile("file-ts-e-unused.ts"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("e", 1, 1, 1, 20)}},
	})

	// ignoreUnusedTypeExports
	runGroup(t, "ignoreUnusedTypeExports", nil, []rule_tester.ValidTestCase{
		// unused vars should not report
		{Options: unusedExportsTypescriptIgnoreUnusedTypesOptions, Settings: typescriptSettings, Code: `export interface c {};`, FileName: typescriptFile("file-ts-c-unused.ts")},
		{Options: unusedExportsTypescriptIgnoreUnusedTypesOptions, Settings: typescriptSettings, Code: `export type d = {};`, FileName: typescriptFile("file-ts-d-unused.ts")},
		{Options: unusedExportsTypescriptIgnoreUnusedTypesOptions, Settings: typescriptSettings, Code: `export enum e { f };`, FileName: typescriptFile("file-ts-e-unused.ts")},
		// used vars should not report
		{Options: unusedExportsTypescriptIgnoreUnusedTypesOptions, Settings: typescriptSettings, Code: `export interface c {};`, FileName: typescriptFile("file-ts-c-used-as-type.ts")},
		{Options: unusedExportsTypescriptIgnoreUnusedTypesOptions, Settings: typescriptSettings, Code: `export type d = {};`, FileName: typescriptFile("file-ts-d-used-as-type.ts")},
		{Options: unusedExportsTypescriptIgnoreUnusedTypesOptions, Settings: typescriptSettings, Code: `export enum e { f };`, FileName: typescriptFile("file-ts-e-used-as-type.ts")},
	}, nil)

	// correctly work with JSX only files
	jsxA := fixturePath("./no-unused-modules/jsx/file-jsx-a.jsx")
	runGroup(t, "jsx only files", map[string]string{jsxA: `import a from "file-jsx-a";`}, []rule_tester.ValidTestCase{
		{Options: unusedExportsJsxOptions, Settings: jsxSettings, Code: `import a from "file-jsx-a";`, FileName: jsxA},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsJsxOptions, Settings: jsxSettings, Code: `export const b = 2;`, FileName: fixturePath("./no-unused-modules/jsx/file-jsx-b.jsx"), Errors: []rule_tester.InvalidTestCaseError{unusedError("b", 1, 1)}},
	})

	// ignore flow types: Flow annotations need babel-eslint. rslint parses
	// `.js` files with the TypeScript parser, where `export type`/`export
	// interface` are TypeScript declarations rather than Flow types.
	runGroup(t, "flow types", nil, []rule_tester.ValidTestCase{
		{Skip: true, Options: unusedExportsOptions, Code: `import { type FooType, type FooInterface } from "./flow-2";`, FileName: fixturePath("./no-unused-modules/flow/flow-0.js")},
		{Skip: true, Options: unusedExportsOptions, Code: `// @flow strict
               export type FooType = string;
               export interface FooInterface {};
               `, FileName: fixturePath("./no-unused-modules/flow/flow-2.js")},
		{Skip: true, Options: unusedExportsOptions, Code: `import type { FooType, FooInterface } from "./flow-4";`, FileName: fixturePath("./no-unused-modules/flow/flow-3.js")},
		{Skip: true, Options: unusedExportsOptions, Code: `// @flow strict
               export type FooType = string;
               export interface FooInterface {};
               `, FileName: fixturePath("./no-unused-modules/flow/flow-4.js")},
		{Skip: true, Options: unusedExportsOptions, Code: `// @flow strict
               export type Bar = number;
               export interface BarInterface {};
               `, FileName: fixturePath("./no-unused-modules/flow/flow-1.js")},
	}, nil)

	// support (nested) destructuring assignment
	runGroup(t, "nested destructuring", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `import {a, b} from "./destructuring-b";`, FileName: fixturePath("./no-unused-modules/destructuring-a.js")},
		{Options: unusedExportsOptions, Code: `const obj = {a: 1, dummy: {b: 2}}; export const {a, dummy: {b}} = obj;`, FileName: fixturePath("./no-unused-modules/destructuring-b.js")},
	}, nil)

	// support ES2022 Arbitrary module namespace identifier names
	runGroup(t, "arbitrary module namespace identifier names", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `import { "foo" as foo } from "./arbitrary-module-namespace-identifier-name-a"`, FileName: fixturePath("./no-unused-modules/arbitrary-module-namespace-identifier-name-b.js")},
		{Options: unusedExportsOptions, Code: "const foo = 333;\nexport { foo as \"foo\" }", FileName: fixturePath("./no-unused-modules/arbitrary-module-namespace-identifier-name-a.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "const foo = 333\nexport { foo as \"foo\" }", FileName: fixturePath("./no-unused-modules/arbitrary-module-namespace-identifier-name-c.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("foo", 2, 10, 2, 22)}},
	})

	// parser ignores prefixes like BOM and hashbang
	prefixChild := fixturePath("./no-unused-modules/prefix-child.js")
	for _, parent := range []struct{ name, file, code string }{
		{"bom, hashbang", "prefix-parent-bom.js", "\uFEFF#!/usr/bin/env node\nimport {foo} from './prefix-child.js';\n"},
		{"no bom, hashbang", "prefix-parent-hashbang.js", "#!/usr/bin/env node\nimport {foo} from './prefix-child.js';\n"},
		{"bom, no hashbang", "prefix-parent-bomhashbang.js", "\uFEFF#!/usr/bin/env node\nimport {foo} from './prefix-child.js';\n"},
		{"no bom, no hashbang", "prefix-parent.js", "import {foo} from './prefix-child.js';\n"},
	} {
		runGroup(t, "prefix "+parent.name, nil, []rule_tester.ValidTestCase{
			{Options: unusedExportsOptions, Code: "export const foo = 1;\n", FileName: prefixChild},
			{Options: unusedExportsOptions, Code: parent.code, FileName: fixturePath("./no-unused-modules/" + parent.file)},
		}, nil)
	}

	// supports flat eslint: the configuration format does not change rslint.
	runGroup(t, "flat config", nil, []rule_tester.ValidTestCase{
		{Options: unusedExportsOptions, Code: `import { o2 } from "./file-o"; export default () => 12`, FileName: fixturePath("./no-unused-modules/file-a.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: unusedExportsOptions, Code: "export default () => 13", FileName: fileF, Errors: []rule_tester.InvalidTestCaseError{unusedError("default", 1, 1)}},
	})

	// Documentation examples. The docs import bare names such as 'file-a';
	// they are written as relative imports here so they resolve.
	docs := func(name string) string { return fixturePath("./docs/" + name) }
	docsOptions := map[string]any{"unusedExports": true, "src": []any{fixturePath("./docs")}}
	docFiles := map[string]string{
		docs("file-f.js"): "import { e } from './file-a'\nimport { f } from './file-b'\nimport * as fileC from  './file-c'\nexport { default, i0 } from './file-d' // both will be reported\n\nexport const j = 99 // will be reported\n",
		docs("file-d.js"): "export const i0 = 9 // will not be reported\nexport const i1 = 9 // will be reported\nexport default () => {} // will not be reported\n",
		docs("file-c.js"): "export const h = 8 // will not be reported\nexport default () => {} // will be reported, as export * only considers named exports and ignores default exports\n",
		docs("file-b.js"): "import two, { b, c, doAnything } from './file-a'\n\nexport const f = 6 // will not be reported\n",
		docs("file-a.js"): "const b = 2\nconst c = 3\nconst d = 4\n\nexport const a = 1 // will be reported\n\nexport { b, c } // will not be reported\n\nexport { d as e } // will not be reported\n\nexport function doAnything() {\n  // some code\n}  // will not be reported\n\nexport default 5 // will not be reported\n",
	}
	runGroup(t, "documentation unused exports", docFiles, []rule_tester.ValidTestCase{
		{Options: docsOptions, Code: docFiles[docs("file-b.js")], FileName: docs("file-b.js")},
		// The docs say file-c's default is reported, but file-f imports
		// file-c as a namespace, which uses every export; upstream does not
		// report it either.
		{Options: docsOptions, Code: docFiles[docs("file-c.js")], FileName: docs("file-c.js")},
	}, []rule_tester.InvalidTestCase{
		{Options: docsOptions, Code: docFiles[docs("file-f.js")], FileName: docs("file-f.js"), Errors: []rule_tester.InvalidTestCaseError{
			unusedRange("default", 4, 10, 4, 17), unusedRange("i0", 4, 19, 4, 21), unusedRange("j", 6, 1, 6, 20),
		}},
		{Options: docsOptions, Code: docFiles[docs("file-d.js")], FileName: docs("file-d.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("i1", 2, 1, 2, 20)}},
		{Options: docsOptions, Code: docFiles[docs("file-a.js")], FileName: docs("file-a.js"), Errors: []rule_tester.InvalidTestCaseError{unusedRange("a", 5, 1, 5, 19)}},
	})
	runGroup(t, "documentation ignoreUnusedTypeExports", nil, []rule_tester.ValidTestCase{
		// The docs write `export interface Foo = {}`, which is not valid
		// TypeScript, and declare Foo three times.
		{Options: map[string]any{"unusedExports": true, "ignoreUnusedTypeExports": true, "src": []any{fixturePath("./docs")}}, Code: "export type Foo = {}; // will not be reported\nexport interface Bar {}; // will not be reported\nexport enum Baz {}; // will not be reported", FileName: docs("types.ts")},
	}, nil)
	runGroup(t, "documentation missing exports", nil, []rule_tester.ValidTestCase{
		{Options: missingExportsOptions, Code: "export default function () { /*...*/ }", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "export const foo = function () { /*...*/ }", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "const foo = 1, bar = 2; export { foo, bar }", FileName: fixturePath("missing.js")},
		{Options: missingExportsOptions, Code: "const foo = 1; export { foo as bar }", FileName: fixturePath("missing.js")},
	}, []rule_tester.InvalidTestCase{
		// The docs write `const class MyClass`, which does not parse.
		{Options: missingExportsOptions, Code: "class MyClass { /*...*/ }\n\nfunction makeClass() { return new MyClass(...arguments) }", FileName: fixturePath("missing.js"), Errors: []rule_tester.InvalidTestCaseError{noExportsError(1, 1, 1, 26)}},
	})

	// Upstream's "with eslint 9+" test checks the error ESLint 9 raises
	// without an .eslintrc file. rslint enumerates `src` itself and needs no
	// legacy configuration file, so there is no equivalent case.
}
