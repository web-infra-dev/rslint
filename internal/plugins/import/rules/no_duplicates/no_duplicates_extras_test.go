package no_duplicates_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_duplicates"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/embedfs"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

func TestNoDuplicatesRuntimePaths(t *testing.T) {
	root := noDuplicatesRuntimeRoot(t)

	for _, extensions := range [][2]string{{"js", "ts"}, {"jsx", "tsx"}, {"mjs", "mts"}, {"cjs", "cts"}} {
		for _, attributes := range []string{"", " with { type: 'text' }"} {
			for _, reverse := range []bool{false, true} {
				first, second := extensions[0], extensions[1]
				if reverse {
					first, second = second, first
				}
				code := fmt.Sprintf("import { value as first } from './coexisting/text.%s'%s;\nimport { value as second } from './coexisting/text.%s'%s;\n", first, attributes, second, attributes)
				t.Run(first+"-"+second+attributes, func(t *testing.T) {
					assertNoDuplicatesRuntimeResult(t, root, code, nil, 0, code)
				})
			}
		}
	}

	for _, test := range []struct {
		name, code, output string
		options            []any
		diagnostics        int
	}{
		{
			name: "source phase and evaluation phase stay separate",
			code: "import source compiled from './coexisting/text.js';\nimport { value } from './coexisting/text.js';\n",
		},
		{
			name: "side effect and source phase stay separate",
			code: "import './coexisting/text.js';\nimport source compiled from './coexisting/text.js';\n",
		},
		{
			name:        "duplicate source imports preserve the phase",
			code:        "import source compiled from './coexisting/text.js';\nimport source compiled from './coexisting/text.js';\n",
			output:      "import source compiled from './coexisting/text.js';\n",
			diagnostics: 2,
		},
		{
			name:        "normalized existing file",
			code:        "import { value as first } from './coexisting/text.js';\nimport { value as second } from './coexisting/sub/../text.js';\n",
			output:      "import { value as first , value as second } from './coexisting/text.js';\n",
			diagnostics: 2,
		},
		{
			name:        "duplicate js beside ts",
			code:        "import { value as first } from './coexisting/text.js';\nimport { value as second } from './coexisting/text.js';\nimport { value as third } from './coexisting/text.ts';\n",
			output:      "import { value as first , value as second } from './coexisting/text.js';\nimport { value as third } from './coexisting/text.ts';\n",
			diagnostics: 2,
		},
		{
			name:        "duplicate ts beside js",
			code:        "import { value as first } from './coexisting/text.ts';\nimport { value as second } from './coexisting/text.ts';\nimport { value as third } from './coexisting/text.js';\n",
			output:      "import { value as first , value as second } from './coexisting/text.ts';\nimport { value as third } from './coexisting/text.js';\n",
			diagnostics: 2,
		},
		{
			name: "different attributes on existing file",
			code: "import { value as first } from './coexisting/text.js';\nimport { value as second } from './coexisting/text.js' with { type: 'text' };\n",
		},
		{
			name:        "query strings ignored",
			code:        "import { value as first } from './coexisting/text.js?one';\nimport { value as second } from './coexisting/sub/../text.js?two';\n",
			output:      "import { value as first , value as second } from './coexisting/text.js?one';\n",
			diagnostics: 2,
		},
		{
			name:    "query strings distinguished",
			code:    "import { value as first } from './coexisting/text.js?one';\nimport { value as second } from './coexisting/text.js?two';\n",
			options: []any{map[string]any{"considerQueryString": true}},
		},
		{
			name:        "identical query strings normalized",
			code:        "import { value as first } from './coexisting/text.js?raw';\nimport { value as second } from './coexisting/sub/../text.js?raw';\n",
			output:      "import { value as first , value as second } from './coexisting/text.js?raw';\n",
			options:     []any{map[string]any{"considerQueryString": true}},
			diagnostics: 2,
		},
		{
			name:    "identical query strings on distinct files",
			code:    "import { value as first } from './coexisting/text.js?raw';\nimport { value as second } from './coexisting/text.ts?raw';\n",
			options: []any{map[string]any{"considerQueryString": true}},
		},
		{
			name: "ambiguous alias and explicit ts",
			code: "import { value as first } from '@fixtures/coexisting/text.js';\nimport { value as second } from './coexisting/text.ts';\n",
		},
		{
			name: "ambiguous extensionless alias and explicit ts",
			code: "import { value as first } from '@fixtures/coexisting/text';\nimport { value as second } from './coexisting/text.ts';\n",
		},
		{
			name: "ambiguous alias and explicit js remain separate",
			code: "import { value as first } from '@fixtures/coexisting/text.js';\nimport { value as second } from './coexisting/text.js';\n",
		},
		{
			name: "path alias shadows installed package",
			code: "import { value as first } from 'shadowed-package';\nimport { value as second } from 'shadowed-package/index.js';\n",
		},
		{
			name:        "shadowing alias still matches explicit target",
			code:        "import { value as first } from 'shadowed-package';\nimport { value as second } from './substitution/text.ts';\n",
			output:      "import { value as first , value as second } from 'shadowed-package';\n",
			diagnostics: 2,
		},
		{
			name: "wildcard alias shadows installed package",
			code: "import { value as first } from 'shadowed-wildcard/text.js';\nimport { value as second } from './node_modules/shadowed-wildcard/text.js';\n",
		},
		{
			name: "package imports with coexisting extensions",
			code: "import { value as first } from 'extension-pair/text.js';\nimport { value as second } from 'extension-pair/text.ts';\n",
		},
		{
			name: "package entry points sharing declarations",
			code: "import { value as first } from 'shared-types/first';\nimport { value as second } from 'shared-types/second';\n",
		},
		{
			name:        "package entry points sharing runtime file",
			code:        "import { value as first } from 'shared-types/first';\nimport { value as second } from 'shared-types/alias';\n",
			output:      "import { value as first , value as second } from 'shared-types/first';\n",
			diagnostics: 2,
		},
		{
			name:        "package query matches request without query",
			code:        "import { value as first } from 'shared-types/first';\nimport { value as second } from 'shared-types/first?raw';\n",
			output:      "import { value as first , value as second } from 'shared-types/first';\n",
			diagnostics: 2,
		},
		{
			name: "shared runtime with distinct type declarations",
			code: "import type { First } from 'shared-runtime/first';\nimport type { Second } from 'shared-runtime/second';\n",
		},
		{
			name:    "shared runtime with distinct type declarations and prefer inline",
			code:    "import type { First } from 'shared-runtime/first';\nimport type { Second } from 'shared-runtime/second';\n",
			options: []any{map[string]any{"prefer-inline": true}},
		},
		{
			name:    "queried shared runtime with distinct type declarations",
			code:    "import type { First } from 'shared-runtime/first?raw';\nimport type { Second } from 'shared-runtime/second?raw';\n",
			options: []any{map[string]any{"prefer-inline": true}},
		},
		{
			name: "shared runtime with distinct inline type declarations",
			code: "import { type First } from 'shared-runtime/first';\nimport { type Second } from 'shared-runtime/second';\n",
		},
		{
			name:    "shared runtime with distinct inline types and prefer inline",
			code:    "import { type First } from 'shared-runtime/first';\nimport { type Second } from 'shared-runtime/second';\n",
			options: []any{map[string]any{"prefer-inline": true}},
		},
		{
			name: "shared runtime with distinct value declarations",
			code: "import { value as first } from 'shared-runtime/first';\nimport { value as second } from 'shared-runtime/second';\n",
		},
		{
			name:        "prefer inline with matching runtime and declarations",
			code:        "import type { First } from 'shared-runtime/first';\nimport { value } from 'shared-runtime/first';\n",
			output:      "import { type First , value } from 'shared-runtime/first';\n",
			options:     []any{map[string]any{"prefer-inline": true}},
			diagnostics: 2,
		},
		{
			name: "posix backslash is not a proven separator",
			code: "import { value as first } from './coexisting\\\\text.js';\nimport { value as second } from './coexisting/text.js';\n",
		},
		{
			name:        "unambiguous path alias",
			code:        "import { value as first } from '@fixtures/substitution/text';\nimport { value as second } from './substitution/text.ts';\n",
			output:      "import { value as first , value as second } from '@fixtures/substitution/text';\n",
			diagnostics: 2,
		},
		{
			name: "type and value imports stay separate",
			code: "import type { Value } from './substitution/text.ts';\nimport { value } from './substitution/text.ts';\n",
		},
		{
			name:        "prefer inline still merges type imports",
			code:        "import type { Value } from './substitution/text.ts';\nimport { value } from './substitution/text.ts';\n",
			output:      "import { type Value , value } from './substitution/text.ts';\n",
			options:     []any{map[string]any{"prefer-inline": true}},
			diagnostics: 2,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.output == "" {
				test.output = test.code
			}
			assertNoDuplicatesRuntimeResult(t, root, test.code, test.options, test.diagnostics, test.output)
		})
	}

	for _, extensions := range [][2]string{{"js", "ts"}, {"jsx", "tsx"}, {"mjs", "mts"}, {"cjs", "cts"}} {
		t.Run("substitution-"+extensions[0], func(t *testing.T) {
			code := fmt.Sprintf("import { value as first } from './substitution/text.%s';\nimport { value as second } from './substitution/text.%s';\n", extensions[0], extensions[1])
			output := fmt.Sprintf("import { value as first , value as second } from './substitution/text.%s';\n", extensions[0])
			assertNoDuplicatesRuntimeResult(t, root, code, nil, 2, output)
		})
	}
}

func TestNoDuplicatesPackageURLs(t *testing.T) {
	root := noDuplicatesRuntimeRoot(t)
	for _, test := range []struct {
		name, first, second string
		diagnostics         int
	}{
		{"distinct queries", "query-one", "query-two", 0},
		{"distinct fragments", "fragment-one", "fragment-two", 0},
		{"query versus plain", "query-one", "plain", 0},
		{"fragment versus plain", "fragment-one", "plain", 0},
		{"empty query versus plain", "empty-query", "plain", 0},
		{"empty fragment versus plain", "empty-fragment", "plain", 0},
		{"identical queries", "query-one", "query-alias", 2},
		{"identical fragments", "fragment-one", "fragment-alias", 2},
	} {
		for _, considerQueryString := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/considerQueryString=%v", test.name, considerQueryString), func(t *testing.T) {
				code := fmt.Sprintf("import { x } from 'url-targets/%s';\nimport { y } from 'url-targets/%s';\n", test.first, test.second)
				output := code
				if test.diagnostics != 0 {
					output = fmt.Sprintf("import { x , y } from 'url-targets/%s';\n", test.first)
				}
				assertNoDuplicatesRuntimeResult(t, root, code, []any{map[string]any{"considerQueryString": considerQueryString}}, test.diagnostics, output)
			})
		}
	}
}

func TestNoDuplicatesPathAliasFallback(t *testing.T) {
	root := noDuplicatesRuntimeRoot(t)
	const code = "import { value as first } from 'shared-types/first';\nimport { value as second } from 'shared-types/alias';\n"
	for _, test := range []struct {
		name, paths string
		diagnostics int
	}{
		{"missing wildcard target", `{"*": ["./missing/*"]}`, 2},
		{"missing exact targets", `{"shared-types/first": ["./missing.ts"], "shared-types/alias": ["./missing.ts"]}`, 2},
		{"missing exact targets outrank existing wildcard", `{"shared-types/first": ["./missing.ts"], "shared-types/alias": ["./missing.ts"], "*": ["./substitution/text.ts"]}`, 2},
		{"missing specific wildcard outranks existing wildcard", `{"shared-types/*": ["./missing/*"], "*": ["./substitution/text.ts"]}`, 2},
		{"successful declaration alias inside node_modules", `{"shared-types/first": ["./node_modules/shared-types/shared.d.ts"]}`, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			caseRoot := root
			caseRoot.FS = utils.NewOverlayVFS(root.FS, map[string]string{
				tspath.ResolvePath(root.Dir, "tsconfig.json"): `{"compilerOptions": {"module": "esnext", "moduleResolution": "bundler", "paths": ` + test.paths + `}, "files": ["entry.ts"]}`,
			})
			output := code
			if test.diagnostics != 0 {
				output = "import { value as first , value as second } from 'shared-types/first';\n"
			}
			assertNoDuplicatesRuntimeResult(t, caseRoot, code, nil, test.diagnostics, output)
		})
	}
}

func TestNoDuplicatesCommonJSRuntimePaths(t *testing.T) {
	root := noDuplicatesRuntimeRoot(t)
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.json"): `{"compilerOptions": {"module": "node16", "moduleResolution": "node16"}, "files": ["entry.cts"]}`,
	})
	const code = "import { x } from './commonjs/text';\nimport { y } from './commonjs/text.js';\n"
	const output = "import { x , y } from './commonjs/text';\n"
	assertNoDuplicatesRuntimeFileResult(t, root, "entry.cts", code, nil, 2, output)
}

func TestNoDuplicatesRuntimeConditions(t *testing.T) {
	root := noDuplicatesRuntimeRoot(t)
	const code = "import { value as first } from 'conditional-runtime/first';\nimport { value as second } from 'conditional-runtime/second';\n"
	for _, test := range []struct {
		name, compilerOptions string
		diagnostics           int
	}{
		{"import", `{"module": "esnext", "moduleResolution": "bundler"}`, 2},
		{"require", `{"module": "node16", "moduleResolution": "node16"}`, 0},
		{"custom condition", `{"module": "esnext", "moduleResolution": "bundler", "customConditions": ["development"]}`, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			caseRoot := root
			caseRoot.FS = utils.NewOverlayVFS(root.FS, map[string]string{
				tspath.ResolvePath(root.Dir, "tsconfig.json"): `{"compilerOptions": ` + test.compilerOptions + `, "files": ["entry.ts"]}`,
			})
			output := code
			if test.diagnostics == 2 {
				output = "import { value as first , value as second } from 'conditional-runtime/first';\n"
			}
			assertNoDuplicatesRuntimeResult(t, caseRoot, code, nil, test.diagnostics, output)
		})
	}
}

func TestNoDuplicatesSymlinkIdentities(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/runtime-paths.txtar")
	directory := archive.Materialize(t, "")
	linkPath := filepath.Join(directory, "link.js")
	if err := os.Symlink("coexisting/text.js", linkPath); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
	root := embedfs.Root{Dir: tspath.NormalizePath(directory), FS: bundled.WrapFS(osvfs.FS())}
	t.Run("target/distinct type declarations", func(t *testing.T) {
		caseRoot := noDuplicatesSymlinkOptions(t, root, false)
		const code = "import { value as first } from './link.js';\nimport { value as second } from './coexisting/text.js';\n"
		assertNoDuplicatesRuntimeResult(t, caseRoot, code, nil, 0, code)
	})
	if err := os.Remove(linkPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real/text.js", linkPath); err != nil {
		t.Fatal(err)
	}
	const code = "import { value as first } from './link.js';\nimport { value as second } from './real/text.js';\n"
	for _, preserve := range []bool{false, true} {
		t.Run(fmt.Sprintf("target/preserve=%v", preserve), func(t *testing.T) {
			caseRoot := noDuplicatesSymlinkOptions(t, root, preserve)
			count, output := 0, code
			if !preserve {
				count = 2
				output = "import { value as first , value as second } from './link.js';\n"
			}
			assertNoDuplicatesRuntimeResult(t, caseRoot, code, nil, count, output)
		})
	}

	if err := os.Remove(filepath.Join(directory, "entry.ts")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real/entry.ts", filepath.Join(directory, "entry.ts")); err != nil {
		t.Fatal(err)
	}
	for _, preserve := range []bool{false, true} {
		t.Run(fmt.Sprintf("importer/preserve=%v", preserve), func(t *testing.T) {
			caseRoot := noDuplicatesSymlinkOptions(t, root, preserve)
			code := fmt.Sprintf("import { value as first } from './text.js';\nimport { value as second } from %q;\n", tspath.ResolvePath(root.Dir, "text.js"))
			count, output := 0, code
			if preserve {
				count = 2
				output = "import { value as first , value as second } from './text.js';\n"
			}
			assertNoDuplicatesRuntimeResult(t, caseRoot, code, nil, count, output)
		})
	}
	t.Run("importer/compiler cache directory differs", func(t *testing.T) {
		caseRoot := noDuplicatesSymlinkOptions(t, root, false)
		code := fmt.Sprintf("import { value as first } from './only.js';\nimport { value as second } from %q;\n", tspath.ResolvePath(root.Dir, "only.ts"))
		assertNoDuplicatesRuntimeResult(t, caseRoot, code, nil, 0, code)
	})
}

func noDuplicatesSymlinkOptions(t *testing.T, root embedfs.Root, preserve bool) embedfs.Root {
	t.Helper()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.json"): fmt.Sprintf(`{"compilerOptions": {"module": "esnext", "moduleResolution": "bundler", "allowJs": true, "preserveSymlinks": %v}, "files": ["entry.ts"]}`, preserve),
	})
	return root
}

func noDuplicatesRuntimeRoot(t *testing.T) embedfs.Root {
	t.Helper()
	root := fixtures.GetRootDir()
	root.Dir = "/no-duplicates-runtime"
	archive := txtarfs.MustParseFile(t, "testdata/runtime-paths.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string]string, len(names))
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

func assertNoDuplicatesRuntimeResult(t *testing.T, root embedfs.Root, code string, options []any, wantDiagnostics int, wantOutput string) {
	t.Helper()
	assertNoDuplicatesRuntimeFileResult(t, root, "entry.ts", code, options, wantDiagnostics, wantOutput)
}

func assertNoDuplicatesRuntimeFileResult(t *testing.T, root embedfs.Root, name, code string, options []any, wantDiagnostics int, wantOutput string) {
	t.Helper()
	filePath := tspath.ResolvePath(root.Dir, name)
	fs := utils.NewOverlayVFS(root.FS, map[string]string{filePath: code})
	program, err := utils.CreateProgram(true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs))
	if err != nil {
		t.Fatal(err)
	}
	if sourceFile := program.GetSourceFile(pathutil.File(filePath)); sourceFile == nil || sourceFile.Text() != code {
		t.Fatal("test source file was not loaded with the expected content")
	}
	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: lintprogram.NewFromCompiler(program), File: filePath, HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name: no_duplicates.NoDuplicatesRule.Name, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_duplicates.NoDuplicatesRule.Run(ctx, options)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{Demand: rule.EditDemandAll, Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})
	if len(diagnostics) != wantDiagnostics {
		t.Errorf("got %d diagnostics, want %d: %#v", len(diagnostics), wantDiagnostics, diagnostics)
	}
	output, _, fixed := linter.ApplyRuleFixes(code, diagnostics)
	if output != wantOutput {
		t.Errorf("unexpected autofix:\ngot  %q\nwant %q", output, wantOutput)
	}
	if fixed != (code != wantOutput) {
		t.Errorf("fixed = %v, want %v", fixed, code != wantOutput)
	}
}

func TestNoDuplicatesImportAttributes(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &no_duplicates.NoDuplicatesRule,
		[]rule_tester.ValidTestCase{
			// Different import attributes describe distinct module requests.
			{Code: `import { x } from './foo' with { type: 'json' }; import { y } from './foo' with { type: 'text' }`},
			{Code: `import { x } from './foo'; import { y } from './foo' with { type: 'text' }`},
		},
		[]rule_tester.InvalidTestCase{
			// Attribute order is not part of module-request identity.
			{
				Code:   `import {x} from './foo' with { type: 'json', mode: 'strict' }; import {y} from './foo' with { mode: 'strict', type: 'json' }`,
				Output: []string{`import {x,y} from './foo' with { type: 'json', mode: 'strict' }; `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicates", Line: 1, Column: 17},
					{MessageId: "noDuplicates", Line: 1, Column: 80},
				},
			},
		},
	)
}

func TestNoDuplicatesEditDemand(t *testing.T) {
	t.Parallel()

	const code = "import { first } from './shared';\nimport { second } from './shared';\n"
	program, sourceFile := createNoDuplicatesProgram(t, "edit-demand.ts", code)

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:         lintprogram.NewFromCompiler(program),
			File:            sourceFile.FileName().AsString(),
			HasTypeInfo:     true,
			GetRulesForFile: noDuplicatesConfiguredRules,
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixes := run(rule.EditDemandAutofix)
	suggestionsOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	for name, diagnostics := range map[string][]rule.RuleDiagnostic{
		"diagnostics only": diagnosticsOnly,
		"autofixes":        autofixes,
		"suggestions only": suggestionsOnly,
		"all edits":        allEdits,
	} {
		if len(diagnostics) != 2 {
			t.Fatalf("%s: got %d diagnostics, want 2", name, len(diagnostics))
		}
		for i, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Errorf("%s: diagnostic %d unexpectedly has suggestions", name, i)
			}
		}
	}

	for i := range allEdits {
		want := diagnosticWithoutEdits(allEdits[i])
		for name, diagnostics := range map[string][]rule.RuleDiagnostic{
			"diagnostics only": diagnosticsOnly,
			"autofixes":        autofixes,
			"suggestions only": suggestionsOnly,
		} {
			if got := diagnosticWithoutEdits(diagnostics[i]); !reflect.DeepEqual(got, want) {
				t.Errorf("%s: diagnostic %d changed:\ngot  %#v\nwant %#v", name, i, got, want)
			}
		}
	}

	if got := len(autofixes[0].Fixes()); got == 0 {
		t.Fatal("autofixes: first diagnostic has no merge fix")
	}
	if !reflect.DeepEqual(autofixes[0].Fixes(), allEdits[0].Fixes()) {
		t.Fatalf("autofix and all-edits modes produced different fixes:\nautofix %#v\nall     %#v", autofixes[0].Fixes(), allEdits[0].Fixes())
	}
	for name, diagnostics := range map[string][]rule.RuleDiagnostic{
		"diagnostics only": diagnosticsOnly,
		"suggestions only": suggestionsOnly,
	} {
		for i, diagnostic := range diagnostics {
			if got := len(diagnostic.Fixes()); got != 0 {
				t.Errorf("%s: diagnostic %d has %d fixes, want 0", name, i, got)
			}
		}
	}
	for i := 1; i < len(autofixes); i++ {
		if got := len(autofixes[i].Fixes()); got != 0 {
			t.Errorf("autofixes: diagnostic %d has %d fixes, want 0", i, got)
		}
	}

	const unfixableCode = "import first from './shared';\nimport second from './shared';\n"
	unfixableProgram, unfixableSourceFile := createNoDuplicatesProgram(t, "unfixable.ts", unfixableCode)
	var unfixableDiagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:         lintprogram.NewFromCompiler(unfixableProgram),
		File:            unfixableSourceFile.FileName().AsString(),
		HasTypeInfo:     true,
		GetRulesForFile: noDuplicatesConfiguredRules,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: func(diagnostic rule.RuleDiagnostic) {
				unfixableDiagnostics = append(unfixableDiagnostics, diagnostic)
			},
		},
	})
	if len(unfixableDiagnostics) != 2 {
		t.Fatalf("unfixable imports: got %d diagnostics, want 2", len(unfixableDiagnostics))
	}
	for i, diagnostic := range unfixableDiagnostics {
		if diagnostic.FixesPtr != nil {
			t.Errorf("unfixable imports: diagnostic %d has a non-nil fix payload", i)
		}
	}
}

func TestNoDuplicatesResolutionModeOverrides(t *testing.T) {
	t.Parallel()

	rootDir := fixtures.GetRootDir()
	const fileName = "resolution-mode.ts"
	filePath := tspath.ResolvePath(rootDir.Dir, fileName)
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{
		filePath: `import type { ImportType } from "conditional-package" with { "resolution-mode": "import" };
import type { RequireType } from "conditional-package" with { "resolution-mode": "require" };`,
		tspath.ResolvePath(rootDir.Dir, "tsconfig.resolution-mode.json"): `{
  "compilerOptions": {"module": "node16", "moduleResolution": "node16", "target": "es2020"},
  "files": ["resolution-mode.ts"]
}`,
		tspath.ResolvePath(rootDir.Dir, "node_modules/conditional-package/package.json"): `{
  "name": "conditional-package",
  "exports": {".": {"import": "./import.d.ts", "require": "./require.d.ts"}}
}`,
		tspath.ResolvePath(rootDir.Dir, "node_modules/conditional-package/import.d.ts"):  `export interface ImportType { kind: "import" }`,
		tspath.ResolvePath(rootDir.Dir, "node_modules/conditional-package/require.d.ts"): `export interface RequireType { kind: "require" }`,
	})
	program, err := utils.CreateProgram(
		true,
		fs,
		rootDir.Dir,
		"tsconfig.resolution-mode.json",
		utils.CreateCompilerHost(rootDir.Dir, fs),
	)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(pathutil.File(filePath))
	if sourceFile == nil {
		t.Fatalf("source file %q not found", filePath)
		return
	}

	var resolvedPaths []string
	for _, statement := range sourceFile.Statements.Nodes {
		moduleSpecifier := statement.AsImportDeclaration().ModuleSpecifier
		resolved := program.GetResolvedModuleFromModuleSpecifier(sourceFile, moduleSpecifier)
		if resolved == nil {
			t.Fatalf("import on line %d did not resolve", len(resolvedPaths)+1)
			continue
		}
		resolvedPaths = append(resolvedPaths, resolved.ResolvedFileName.AsString())
	}
	if len(resolvedPaths) != 2 || resolvedPaths[0] == resolvedPaths[1] {
		t.Fatalf("resolution-mode overrides did not select distinct targets: %v", resolvedPaths)
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:         lintprogram.NewFromCompiler(program),
		File:            sourceFile.FileName().AsString(),
		HasTypeInfo:     true,
		GetRulesForFile: noDuplicatesConfiguredRules,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandNone,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	if len(diagnostics) != 0 {
		t.Fatalf("imports resolved with different modes were reported as duplicates: %#v", diagnostics)
	}
}

type diagnosticIdentity struct {
	Range    [2]int
	RuleName string
	Message  rule.RuleMessage
	FilePath string
	Severity rule.DiagnosticSeverity
}

func diagnosticWithoutEdits(diagnostic rule.RuleDiagnostic) diagnosticIdentity {
	return diagnosticIdentity{
		Range:    [2]int{diagnostic.Range.Pos(), diagnostic.Range.End()},
		RuleName: diagnostic.RuleName,
		Message:  diagnostic.Message,
		FilePath: diagnostic.FilePath,
		Severity: diagnostic.Severity,
	}
}

func createNoDuplicatesProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFS(rootDir.FS, map[string]string{tspath.ResolvePath(rootDir.Dir, fileName): code})
	host := utils.CreateCompilerHost(rootDir.Dir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(tspath.ToRootedFilePath(fileName, pathutil.Directory(rootDir.Dir)))
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

func noDuplicatesConfiguredRules(*ast.SourceFile) []rule.ConfiguredRule {
	return []rule.ConfiguredRule{{
		Name:     no_duplicates.NoDuplicatesRule.Name,
		Severity: rule.SeverityError,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return no_duplicates.NoDuplicatesRule.Run(ctx, nil)
		},
	}}
}
