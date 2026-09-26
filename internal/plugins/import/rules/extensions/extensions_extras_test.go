package extensions_test

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/extensions"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/embedfs"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func extensionsRoot(t *testing.T) embedfs.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = "/extensions"
	files := map[string]string{}
	archive := txtarfs.MustParseFile(t, "testdata/upstream.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		content, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath(root.Dir, name)] = string(content)
	}
	root.FS = utils.NewOverlayVFS(root.FS, files)
	return root
}

// Expected text is supplied independently of the implementation. Locate the
// quoted source in the small test input to assert the complete UTF-16 range.
func extensionError(code, literal, message string) rule_tester.InvalidTestCaseError {
	start := strings.Index(code, literal)
	if start < 0 {
		panic("missing expected source literal: " + literal)
	}
	prefix := code[:start]
	line := strings.Count(prefix, "\n") + 1
	column := ecmascript.StringCodeUnitCount(prefix[strings.LastIndexByte(prefix, '\n')+1:]) + 1
	return rule_tester.InvalidTestCaseError{
		MessageId: "", Message: message,
		Line: line, Column: column, EndLine: line, EndColumn: column + ecmascript.StringCodeUnitCount(literal),
	}
}

func TestExtensionsExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `import './missing';`, Options: []any{map[string]any{}}},
		{Code: `import ''; export * from ''; import(''); require('');`, Options: []any{"always"}},
		{Code: "require(`./missing.js`); import(`./missing.js`); import(`${name}.js`);", Options: []any{"never"}},
		{Code: `obj.require('./missing.js'); obj?.require('./missing.js'); obj['require']('./missing.js'); require.resolve('./missing.js'); new require('./missing.js');`, Options: []any{"never"}},
		{Code: `require('./missing.js' as string); require!('./missing.js'); (require as any)('./missing.js'); require('./missing.js', 1); require();`, Options: []any{"never"}},
		{Code: `require(...['./missing.js']); import('./missing.js' as string);`, Options: []any{"never"}},
		{Code: `import value = require('./missing.js'); type T = import('./missing.js').T; define(['./missing.js'], callback); require(['./missing.js'], callback);`, Options: []any{"never"}},
		{Code: `import type T from './missing'; export type * from './missing'; export type * as ns from './missing';`, Options: []any{"always"}},
		{Code: `export {}; export const x = 0; export default 0;`, Options: []any{"always"}},
		{Code: `import 'fs/promises'; import 'node:fs'; import 'node:test'; import 'electron/subpath';`, Options: []any{"always"}, Settings: map[string]any{"import/core-modules": []any{"electron"}}},
		{Code: `import './.hidden'; import './.hidden.ext';`, Options: []any{"always", map[string]any{"pattern": map[string]any{"": "never"}}}},
		{Code: `import './bar.json';`, Options: []any{"never"}}, // ./bar resolves to bar.js, so .json is necessary.
		{Code: `import './package.json';`, Options: []any{"never"}, Settings: map[string]any{"import/resolver": map[string]any{"node": map[string]any{"extensions": []any{}}}}},
		{Code: `import './missing.js'; import './missing';`, Options: []any{"always", map[string]any{"js": "ignorePackages", "": "ignorePackages"}}},
		{Code: `import './file.constructor';`, Options: []any{map[string]any{"constructor": "ignorePackages"}}},
	}
	var invalid []rule_tester.InvalidTestCase
	add := func(code, literal, message string, options []any, settings map[string]any) {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: options, Settings: settings, Errors: []rule_tester.InvalidTestCaseError{extensionError(code, literal, message)}})
	}
	for _, code := range []string{
		`import('./missing.js');`,
		`import(('./missing.js'), { with: { type: 'json' } });`,
		`((require))(('./missing.js'));`,
		`require?.('./missing.js');`,
		`function f(require) { require('./missing.js'); }`,
		`class C { #load() { return require('./missing.js'); } }`,
		`import type T from './missing.js';`, // Only missing extensions skip type-only declarations.
		`export type { T } from './missing.js';`,
		`export type * from './missing.js';`,
		`export * as ns from './missing.js';`,
		"const 符号 = '😀'; import './missing.js';",
		"const element = <div />;\nimport './missing.js';",
	} {
		add(code, `'./missing.js'`, `Unexpected use of file extension "js" for "./missing.js"`, nil, nil)
	}
	add(`import(/* comment */ './missing');`, `'./missing'`, `Missing file extension for "./missing"`, []any{"always"}, nil)
	add(`import { type T } from './missing';`, `'./missing'`, `Missing file extension for "./missing"`, []any{"always"}, nil)
	add(`export { type T } from './missing';`, `'./missing'`, `Missing file extension for "./missing"`, []any{"always"}, nil)
	add(`export type * as ns from './missing';`, `'./missing'`, `Missing file extension for "./missing"`, []any{"always", map[string]any{"checkTypeImports": true}}, nil)
	add(`import './.hidden';`, `'./.hidden'`, `Missing file extension for "./.hidden"`, []any{"always"}, nil)
	add(`import './.hidden.ext';`, `'./.hidden.ext'`, `Unexpected use of file extension "ext" for "./.hidden.ext"`, nil, nil)
	// The native option map does not inherit Object.prototype exemptions.
	add(`import './file.constructor';`, `'./file.constructor'`, `Unexpected use of file extension "constructor" for "./file.constructor"`, nil, nil)
	add(`import './missing.js?raw';`, `'./missing.js?raw'`, `Unexpected use of file extension "js" for "./missing.js?raw"`, nil, nil)
	add(`import './miss\u0069ng.js';`, `'./miss\u0069ng.js'`, `Unexpected use of file extension "js" for "./missing.js"`, nil, nil)
	add(`import './missing\ud800.js';`, `'./missing\ud800.js'`, "Unexpected use of file extension \"js\" for \"./missing"+ecmascript.StringFromCodeUnits([]uint16{0xd800})+".js\"", nil, nil)
	add(`import './bar';`, `'./bar'`, `Missing file extension "js" for "./bar"`, []any{"always"}, nil)
	add(`import './typescript-declare';`, `'./typescript-declare'`, `Missing file extension "ts" for "./typescript-declare"`, []any{"always"}, map[string]any{"import/resolver": "typescript"})
	add(`import './typescript-declare.ts';`, `'./typescript-declare.ts'`, `Unexpected use of file extension "ts" for "./typescript-declare.ts"`, nil, map[string]any{"import/resolver": "typescript"})
	add(`import './typescript-declare.ts?raw';`, `'./typescript-declare.ts?raw'`, `Unexpected use of file extension "ts" for "./typescript-declare.ts?raw"`, nil, map[string]any{"import/resolver": "typescript"})
	invalid = append(invalid, rule_tester.InvalidTestCase{
		Code: `require('./\
missing.js');`,
		Errors: []rule_tester.InvalidTestCaseError{{
			Message: `Unexpected use of file extension "js" for "./missing.js"`,
			Line:    1, Column: 9, EndLine: 2, EndColumn: 12,
		}},
	})
	for i := range valid {
		valid[i].FileName = "files/input.tsx"
	}
	for i := range invalid {
		invalid[i].FileName = "files/input.tsx"
	}
	rule_tester.RunRuleTester(extensionsRoot(t), "tsconfig.json", t, &extensions.ExtensionsRule, valid, invalid)
}

func TestExtensionsOptionTruthiness(t *testing.T) {
	const packageImport = `import 'pkg/subpath';`
	const typeImport = `import type T from './missing';`
	valid := []rule_tester.ValidTestCase{
		{Code: typeImport, Options: []any{"always", map[string]any{"checkTypeImports": false}}},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: packageImport, Options: []any{"always", map[string]any{"ignorePackages": false}}, Errors: []rule_tester.InvalidTestCaseError{
			extensionError(packageImport, `'pkg/subpath'`, `Missing file extension for "pkg/subpath"`),
		}},
	}
	// The legacy extension-map schema admits each mode string for these flags.
	// Even "never" is truthy; keep boolean true/false as controls.
	for _, value := range []any{true, "always", "ignorePackages", "never"} {
		valid = append(valid, rule_tester.ValidTestCase{Code: packageImport, Options: []any{"always", map[string]any{"ignorePackages": value}}})
		for _, code := range []string{typeImport, `export type { T } from './missing';`} {
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: []any{"always", map[string]any{"checkTypeImports": value}}, Errors: []rule_tester.InvalidTestCaseError{
				extensionError(code, `'./missing'`, `Missing file extension for "./missing"`),
			}})
		}
	}
	rule_tester.RunRuleTester(extensionsRoot(t), "tsconfig.json", t, &extensions.ExtensionsRule, valid, invalid)
}

func TestExtensionsQueriesAndPaths(t *testing.T) {
	// Query stripping follows JavaScript's non-dotAll regexp. A query with a
	// line terminator is part of the extension unless a later '?' can match.
	var invalid []rule_tester.InvalidTestCase
	for _, item := range []struct {
		literal, decoded, extension string
	}{
		{`'./missing.js?raw\nnext'`, "./missing.js?raw\nnext", "js?raw\nnext"},
		{`'./missing.js?raw\n'`, "./missing.js?raw\n", "js?raw\n"},
		{`'./missing.js?raw\r'`, "./missing.js?raw\r", "js?raw\r"},
		{`'./missing.js?raw\r\n'`, "./missing.js?raw\r\n", "js?raw\r\n"},
		{`'./missing.js?raw\u2028'`, "./missing.js?raw\u2028", "js?raw\u2028"},
		{`'./missing.js?raw\u2029'`, "./missing.js?raw\u2029", "js?raw\u2029"},
		{`'./missing.js?raw\n?next'`, "./missing.js?raw\n?next", "js?raw\n"},
		{`'./missing\n.js?raw'`, "./missing\n.js?raw", "js"},
		{`'./missing.js??raw'`, "./missing.js??raw", "js"},
		{`'./missing.js#hash'`, "./missing.js#hash", "js#hash"},
		{`'./missing/..hidden'`, "./missing/..hidden", "hidden"},
	} {
		code := "import " + item.literal + ";"
		message := "Unexpected use of file extension \"" + item.extension + "\" for \"" + item.decoded + "\""
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{extensionError(code, item.literal, message)}})
	}
	for _, item := range []struct{ path, message string }{
		{"./missing.", `Missing file extension for "./missing."`},
		{"./missing.js/", `Missing file extension "js" for "./missing.js/"`},
		{"./missing.js//", `Missing file extension "js" for "./missing.js//"`},
		{"./missing.js/.", `Missing file extension for "./missing.js/."`},
		// The parent segment resolves to the fixture's index.js.
		{"./missing.js/..", `Missing file extension "js" for "./missing.js/.."`},
	} {
		literal := "'" + item.path + "'"
		code := "import " + literal + ";"
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{extensionError(code, literal, item.message)}})
	}
	rule_tester.RunRuleTester(extensionsRoot(t), "tsconfig.json", t, &extensions.ExtensionsRule, nil, invalid)
}

// cspell:ignore nobrace noext nonegate nonull
func TestExtensionsPathGroups(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, item := range []struct {
		name, pattern  string
		patternOptions map[string]any
		match          bool
	}{
		{"./missing.js", "./**/*.js", nil, true},
		{"./missing.js", "./**/*.json", nil, false},
		{"./missing.js?raw", "./**/*.js?raw", nil, true},
		{"./.hidden.js", "./**/*.js", nil, false},
		{"./.hidden.js", "./**/*.js", map[string]any{"dot": true}, true},
		{"./deep/missing.js", "*.js", map[string]any{"matchBase": true}, true},
		{"./MISSING.JS", "./**/*.js", map[string]any{"nocase": true}, true},
		{"./MISSING.JS", "./**/*.js", map[string]any{"nocase": "yes"}, true},
		{"./MISSING.JS", "./**/*.js", map[string]any{"nocase": 0}, false},
		{"./missing.js", "./{missing,other}.js", nil, true},
		{"./missing.js", "./{missing,other}.js", map[string]any{"nobrace": true}, false},
		{"./missing.js", "./@(missing|other).js", nil, true},
		{"./missing.js", "./@(missing|other).js", map[string]any{"noext": true}, false},
		{"./deep/missing.js", "./**/*.js", map[string]any{"noglobstar": true}, true},
		{"./deep/deeper/missing.js", "./**/*.js", map[string]any{"noglobstar": true}, false},
		{"./missing.js", "!./other.js", nil, true},
		{"./missing.js", "!./other.js", map[string]any{"nonegate": true}, false},
		{"./missing.js", "!./missing.js", map[string]any{"flipNegate": true}, true},
		{"./missing.js", "./missing.js/deeper", map[string]any{"partial": true}, true},
		{"./missing.js", "./other.js", map[string]any{"nonull": true}, false},
		{"#alias/missing.js", "#alias/*.js", nil, true},
		{"#alias/missing.js", "#alias/*.js", map[string]any{}, false},
		{"#alias/missing.js", "#alias/*.js", map[string]any{"nocomment": true}, true},
	} {
		group := map[string]any{"pattern": item.pattern, "action": "ignore"}
		if item.patternOptions != nil {
			group["patternOptions"] = item.patternOptions
		}
		opts := []any{"never", map[string]any{"ignorePackages": false, "pathGroupOverrides": []any{group}}}
		code := "import '" + item.name + "';"
		if item.match {
			valid = append(valid, rule_tester.ValidTestCase{Code: code, Options: opts})
		} else {
			extension := "js"
			if strings.HasSuffix(item.name, ".JS") {
				extension = "JS"
			}
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: opts, Errors: []rule_tester.InvalidTestCaseError{extensionError(code, "'"+item.name+"'", "Unexpected use of file extension \""+extension+"\" for \""+item.name+"\"")}})
		}
	}
	for _, name := range []string{"fs", "@scope/pkg", "root.js", "a/subpath", "@scope/pkg/subpath"} {
		code := "import '" + name + "';"
		mode := "always"
		message := "Missing file extension for \"" + name + "\""
		if name == "root.js" {
			mode = "never"
			message = `Unexpected use of file extension "js" for "root.js"`
		}
		opts := []any{mode, map[string]any{"ignorePackages": true, "pathGroupOverrides": []any{map[string]any{"pattern": "**", "action": "enforce"}, map[string]any{"pattern": "**", "action": "ignore"}}}}
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: opts, Errors: []rule_tester.InvalidTestCaseError{extensionError(code, "'"+name+"'", message)}})
	}
	valid = append(valid,
		rule_tester.ValidTestCase{Code: `import 'root.js';`, Options: []any{"never", map[string]any{"pattern": map[string]any{}, "pathGroupOverrides": []any{map[string]any{"pattern": "**", "action": "ignore"}, map[string]any{"pattern": "**", "action": "enforce"}}}}},
		// An unresolved fs.js must not be equated with the builtin fs result.
		rule_tester.ValidTestCase{Code: `import 'fs.js';`, Options: []any{"never", map[string]any{"ignorePackages": false, "pathGroupOverrides": []any{map[string]any{"pattern": "**", "action": "enforce"}}}}},
	)
	code := `import './missing.js';`
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: []any{"never", map[string]any{"pathGroupOverrides": []any{map[string]any{"pattern": "**", "action": "ignore"}}}}, Errors: []rule_tester.InvalidTestCaseError{extensionError(code, `'./missing.js'`, `Unexpected use of file extension "js" for "./missing.js"`)}})
	rule_tester.RunRuleTester(extensionsRoot(t), "tsconfig.json", t, &extensions.ExtensionsRule, valid, invalid)
}

func TestExtensionsPackageSettings(t *testing.T) {
	root := extensionsRoot(t)
	archive := txtarfs.MustParseFile(t, "testdata/packages.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
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
	opts := []any{"always", map[string]any{"ignorePackages": true}}
	typescript := map[string]any{"import/resolver": "typescript"}
	valid := []rule_tester.ValidTestCase{
		{Code: `import 'a/subpath';`, Options: opts},
		{Code: `import '@scope/pkg/subpath';`, Options: opts, Settings: map[string]any{"import/internal-regex": "^@scope"}},
		{Code: `import 'outside/dep';`, Options: opts, Settings: typescript},
		{Code: `import 'local/dep';`, Options: opts, Settings: map[string]any{"import/resolver": "typescript", "import/external-module-folders": []any{"src"}}},
		{Code: `import 'local/dep';`, Options: opts, Settings: map[string]any{"import/resolver": "typescript", "import/external-module-folders": []any{""}}},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, item := range []struct {
		code, literal, message string
		settings               map[string]any
	}{
		{`import 'a/subpath';`, `'a/subpath'`, `Missing file extension for "a/subpath"`, map[string]any{"import/internal-regex": "^a/"}},
		{`import 'local/dep';`, `'local/dep'`, `Missing file extension "ts" for "local/dep"`, typescript},
		{`import '@/dep';`, `'@/dep'`, `Missing file extension "ts" for "@/dep"`, typescript},
		{`import 'local/dep';`, `'local/dep'`, `Missing file extension "ts" for "local/dep"`, map[string]any{"import/resolver": "typescript", "import/external-module-folders": []any{}}},
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: item.code, Options: opts, Settings: item.settings, Errors: []rule_tester.InvalidTestCaseError{extensionError(item.code, item.literal, item.message)}})
	}
	for i := range valid {
		valid[i].FileName = "app/input.ts"
	}
	for i := range invalid {
		invalid[i].FileName = "app/input.ts"
	}
	rule_tester.RunRuleTester(root, "packages.tsconfig.json", t, &extensions.ExtensionsRule, valid, invalid)
}

func TestExtensionsResolverError(t *testing.T) {
	code := `import './one.js'; import './two.js';`
	second := extensionError(code, `'./two.js'`, `Unexpected use of file extension "js" for "./two.js"`)
	rule_tester.RunRuleTester(extensionsRoot(t), "tsconfig.json", t, &extensions.ExtensionsRule, nil, []rule_tester.InvalidTestCase{
		{Code: code, Settings: map[string]any{"import/resolver": "webpack"}, Errors: []rule_tester.InvalidTestCaseError{
			{Message: `Resolve error: unable to load resolver "webpack".`, Line: 1, Column: 1, EndLine: 1, EndColumn: 1},
			extensionError(code, `'./one.js'`, `Unexpected use of file extension "js" for "./one.js"`),
			second,
		}},
	})
}
