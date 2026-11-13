package consistent_json_file_read_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_json_file_read"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentJSONFileReadExtras(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		valid := []string{
			`JSON.parse(...fs.readFileSync(file));`,
			`JSON.parse(fs.readFileSync());`,
			`JSON['parse'](fs.readFileSync(file));`,
			`JSON.parse(new fs.readFileSync(file));`,
			`JSON.parse(await fs?.promises.readFile(file));`,
			`JSON.parse((fs?.readFileSync)(file));`,
			`class Reader { #readFileSync() {} parse(file) { return JSON.parse(this.#readFileSync(file)); } }`,
			`function parse(JSON) { return JSON.parse(fs.readFileSync(file)); }`,
			`function parse(undefined) { return JSON.parse(fs.readFileSync(file, undefined)); }`,
			`export const contents = fs.readFileSync(file); JSON.parse(contents);`,
			`const contents = fs.readFileSync(file); export {contents}; JSON.parse(contents);`,
			`let contents = fs.readFileSync(file); contents = other; JSON.parse(contents);`,
			`const a = b, b = a; JSON.parse(a);`,
			`JSON.parse(fs.readFileSync(file, void setup()));`,
			`JSON.parse(fs.readFileSync(file, { get encoding() { return null; } }));`,
			`JSON.parse(fs.readFileSync(file, {__proto__: {flag: 'r'}, encoding: null}));`,
		}
		defaultOption := readCase(`JSON.parse(fs.readFileSync(file));`, []string{`JSON.parse(fs.readFileSync(file, 'utf8'));`}, false, 1, 12, 1, 33)
		defaultOption.Options = []any{"string"}
		runCases(t, valid, []rule_tester.InvalidTestCase{
			defaultOption,
			readCase(`const file = "📄.json"; JSON.parse(fs.readFileSync(file));`, []string{`const file = "📄.json"; JSON.parse(fs.readFileSync(file, 'utf8'));`}, false, 1, 36, 1, 57),
			readCase(`(JSON).parse((fs.readFileSync)((file)));`, []string{`(JSON).parse((fs.readFileSync)((file), 'utf8'));`}, false, 1, 14, 1, 39),
			readCase(`const method = 'read' + 'FileSync'; JSON.parse(fs[method](file));`, []string{`const method = 'read' + 'FileSync'; JSON.parse(fs[method](file, 'utf8'));`}, false, 1, 48, 1, 64),
			readCase(`JSON.parse(fs.readFileSync(file, {encoding: 'utf8', ...{encoding: null}}));`, []string{`JSON.parse(fs.readFileSync(file, 'utf8'));`}, false, 1, 34, 1, 73),
			readCase(`JSON.parse(fs.readFileSync(file, (/* keep */ null)));`, []string{`JSON.parse(fs.readFileSync(file, (/* keep */ 'utf8')));`}, false, 1, 46, 1, 50),
			readCase("JSON.parse(fs.readFileSync(file, {\n  encoding: null\n}));", []string{`JSON.parse(fs.readFileSync(file, 'utf8'));`}, false, 1, 34, 3, 2),
		}, false)
	})

	t.Run("buffer", func(t *testing.T) {
		valid := []string{
			`const options = {encoding: 'utf8'}; const alias = options; alias.flag = 'r'; JSON.parse(fs.readFileSync(file, options));`,
			`JSON.parse(fs.readFileSync(file, {[(log(), 'encoding')]: 'utf8'}));`,
			`JSON.parse(fs.readFileSync(file, {encoding: 'utf8', [Symbol.iterator]: 123}));`,
			`JSON.parse(fs.readFileSync(file, {__proto__: null, encoding: 'utf8'}));`,
			`let charset = 'utf8'; charset = 'ascii'; JSON.parse(fs.readFileSync(file, charset));`,
		}
		runCases(t, valid, []rule_tester.InvalidTestCase{
			readCase(`JSON.parse(await fs.readFile(file, (/* keep */ "utf8"),));`, []string{`JSON.parse(await fs.readFile(file /* keep */,));`}, true, 1, 48, 1, 54),
			readCase("JSON.parse(await fs.readFile(file, // keep\r\n  \"utf8\",));", []string{"JSON.parse(await fs.readFile(file // keep\r\n,));"}, true, 2, 3, 2, 9),
			readCase(`let charset = "utf8"; JSON.parse(fs.readFileSync(file, charset));`, []string{`let charset = "utf8"; JSON.parse(fs.readFileSync(file));`}, true, 1, 56, 1, 63),
		}, true)
	})

	t.Run("TypeScript", func(t *testing.T) {
		const code = `import {readFileSync as read} from 'fs'; const view = <pre>{JSON.parse(read(file as string))}</pre>;`
		const output = `import {readFileSync as read} from 'fs'; const view = <pre>{JSON.parse(read(file as string, 'utf8'))}</pre>;`
		call := readCase(code, []string{output}, false, 1, 72, 1, 92)
		call.FileName = "file.tsx"
		const optionsCode = `JSON.parse(fs.readFileSync(file, {encoding: null} satisfies Options));`
		const optionsOutput = `JSON.parse(fs.readFileSync(file, 'utf8'));`
		options := readCase(optionsCode, []string{optionsOutput}, false, 1, 34, 1, 68)
		options.FileName = "file.ts"
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &consistent_json_file_read.ConsistentJSONFileReadRule,
			[]rule_tester.ValidTestCase{
				{Code: `import type {readFileSync as read} from 'fs'; JSON.parse(read(file));`, FileName: "file.ts"},
				{Code: `using contents = fs.readFileSync(file); JSON.parse(contents);`, FileName: "file.ts"},
			}, []rule_tester.InvalidTestCase{call, options})
	})
}

func TestConsistentJSONFileReadEditDemand(t *testing.T) {
	const fileName = "file.mjs"
	const code = `JSON.parse(fs.readFileSync(file));
JSON.parse(fs.readFileSync(file, {/* keep */ encoding: null}));
JSON.parse(fs.readFileSync(file, null));`
	const wantOutput = `JSON.parse(fs.readFileSync(file, 'utf8'));
JSON.parse(fs.readFileSync(file, {/* keep */ encoding: null}));
JSON.parse(fs.readFileSync(file, 'utf8'));`
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	compilerProgram, sourceFile, err := helper.CreateTestProgram(code, fileName, "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	p := program.NewFromCompiler(compilerProgram)
	r := consistent_json_file_read.ConsistentJSONFileReadRule
	var full []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: p,
			File:    sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: r.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners { return r.Run(ctx, nil) },
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) { got = append(got, diagnostic) },
			},
		})
		if len(got) != 3 {
			t.Fatalf("demand %d: got %d diagnostics, want 3", demand, len(got))
		}
		if demand == rule.EditDemandAll {
			full = got
			if full[0].FixesPtr == nil || full[1].FixesPtr != nil || full[2].FixesPtr == nil {
				t.Fatal("expected fixes for the first and third diagnostics")
			}
			output, _, fixed := linter.ApplyRuleFixes(code, full)
			if !fixed || output != wantOutput {
				t.Fatalf("fixed output = %q, want %q", output, wantOutput)
			}
		}
		for i, diagnostic := range got {
			want := full[i]
			if demand&rule.EditDemandAutofix == 0 {
				want.FixesPtr = nil
			}
			if diagnostic.Suggestions != nil || !reflect.DeepEqual(want, diagnostic) {
				t.Errorf("demand %d diagnostic %d: got %+v, want %+v", demand, i, diagnostic, want)
			}
		}
	}
}
