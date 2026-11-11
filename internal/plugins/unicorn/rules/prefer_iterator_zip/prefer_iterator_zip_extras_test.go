package prefer_iterator_zip_test

// cspell:words índice nombre nombres valore valores

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_iterator_zip"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferIteratorZipExtras(t *testing.T) {
	const cached = `for (let index = 0; index < length; index++)`
	const minimum = `Math.min(names.length, scores.length)`
	const suggestion = `for (const [name, score] of Iterator.zip([names, scores]))`
	const argumentHeader = `for (let index = 0; index < Math.min(names.length, arguments.length); index++)`
	const annotatedArrays = "/**\n * @param {string[]} names\n * @param {Uint8Array} scores\n */\nfunction pair(names, scores) {\n  "
	valid := []rule_tester.ValidTestCase{
		// Only a plain let index and direct length comparisons qualify.
		{FileName: "input.js", Code: `for (let index = 0; index < ` + minimum + `; )` + loopBody},
		{FileName: "input.js", Code: `for (let [index] = [0]; index < ` + minimum + `; index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0, length = ` + minimum + `, other = 1; index < length; index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math["min"](names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min?.(names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < (Math?.min)(names.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names["length"], scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names?.length, scores.length); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(names.length, scores.length, ...bounds); index++)` + loopBody},
		{FileName: "input.js", Code: `for (let index = 0; index < Math.min(index.length, names.length); index++) { save(index[index], names[index]); }`},
		// The cached declaration must be in the same statement list.
		{FileName: "input.js", Code: `var length = ` + minimum + `; ` + cached + loopBody},
		{FileName: "input.js", Code: `export const length = ` + minimum + `; ` + cached + loopBody},
		{FileName: "input.js", Code: `const length = ` + minimum + `; label: ` + cached + loopBody},
		{FileName: "input.js", Code: `const other = ` + minimum + `; ` + cached + loopBody},
		{FileName: "input.js", Code: `const length = Math.max(names.length, scores.length); ` + cached + loopBody},
		// JSDoc can identify inputs whose indexing differs from iteration.
		{FileName: "input.js", Code: "/** @param {string} names */\nfunction pair(names, scores) { " + loopHeader + loopBody + " }"},
		{FileName: "input.js", Code: "/** @param {{ length: number, [index: number]: string }} names */\nfunction pair(names, scores) { " + loopHeader + loopBody + " }"},
		{FileName: "input.js", Code: "/** @param {?String} [input] */\nfunction pair(input, scores) { const names = (input); " + loopHeader + loopBody + " }"},
		{FileName: "input.js", Code: "/** @type {string} */\nlet names = input;\n" + loopHeader + loopBody},
		// Reference roles within the body.
		{FileName: "input.js", Code: loopHeader + ` { [...names[index]] = scores[index]; }`},
		{FileName: "input.ts", Code: loopHeader + ` { type Index = typeof index; save(names[index], scores[index]); }`},
		{FileName: "input.ts", Code: loopHeader + ` { save(names[index as number], scores[index]); }`},
		{FileName: "input.ts", Code: loopHeader + ` { ((names[index] as any)!)?.(); save(scores[index]); }`},
		{FileName: "input.ts", Code: loopHeader + ` { delete ((names[index] as any)!); save(scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { save({read() { return names[index]; }}, scores[index]); }`},
		{FileName: "input.js", Code: loopHeader + ` { class Pair { [names[index]]() {} } save(scores[index]); }`},
		{FileName: "input.js", Code: `function run(names) { ` + argumentHeader + ` { save(names[index], arguments[index]); const read = () => arguments; } }`},
	}
	invalid := []rule_tester.InvalidTestCase{
		// Parentheses are transparent and positions use UTF-16 columns.
		expectZip("input.js", `const marker = "🐣"; for (let índice = 0x0;
  ((índice) < (Math).min((nombres).length, valores.length));
  (índice) += 0b1) { save(nombres[índice], valores[índice]); }`,
			`const marker = "🐣"; for (const [nombre, valore] of Iterator.zip([nombres, valores])) { save(nombre, valore); }`, 1, 22, 3, 19),
		expectZip("input.js", "switch (kind) { case 0:\n  const length = "+minimum+";\n  "+cached+loopBody+"\n}",
			"switch (kind) { case 0:\n  "+suggestion+" { save(name, score); }\n}", 3, 3, 3, 47),
		expectZip("input.js", `let element; for (let index = 0; index < Math.min(species.length, status.length); index++) { save(species[index], status[index]); }`,
			`let element; for (const [element_, element__] of Iterator.zip([species, status])) { save(element_, element__); }`, 1, 14, 1, 91),
		expectZip("input.js", `for (let index = 0; index < Math.min(classes.length, sets.length); index++) { save(classes[index], sets[index]); }`,
			`for (const [class_, set_] of Iterator.zip([classes, sets])) { save(class_, set_); }`, 1, 1, 1, 76),
		expectZip("input.js", `for (let index = 0; index < Math.min(people.length, persons.length); index++) { save(people[index], persons[index]); }`,
			`for (const [person, person_] of Iterator.zip([people, persons])) { save(person, person_); }`, 1, 1, 1, 78),
		// The Go inflector uses cookie where upstream pluralize uses cooky.
		expectZip("input.js", `for (let index = 0; index < Math.min(cookies.length, scores.length); index++) { save(cookies[index], scores[index]); }`,
			`for (const [cookie, score] of Iterator.zip([cookies, scores])) { save(cookie, score); }`, 1, 1, 1, 78),
		expectZip("input.js", `class Counter { #index = 0; scan(names, scores) { `+loopHeader+` { save(names[index]?.value, scores[index], this.#index, {names: 1}); } } }`,
			`class Counter { #index = 0; scan(names, scores) { `+suggestion+` { save(name?.value, score, this.#index, {names: 1}); } } }`, 1, 51, 1, 126),
		expectZip("input.js", loopHeader+` { consume({[names[index]]() { return 1; }}, scores[index]); }`,
			suggestion+` { consume({[name]() { return 1; }}, score); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(/** @type {string} */ (names[index]), scores[index]); }`,
			suggestion+` { save(/** @type {string} */ (name), score); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(new names[index](scores[index]), "/* literal */"); }`,
			suggestion+` { save(new name(score), "/* literal */"); }`, 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(names[index], scores[index]); { let name_; function callback() { return name; } } }`,
			`for (const [name__, score] of Iterator.zip([names, scores])) { save(name__, score); { let name_; function callback() { return name; } } }`, 1, 1, 1, 76),
		// Computed keys and default values are reads inside a write pattern.
		expectZip("input.js", loopHeader+` { ({[names[index]]: output} = object); [output = scores[index]] = values; }`,
			suggestion+` { ({[name]: output} = object); [output = score] = values; }`, 1, 1, 1, 76),
		// Ordinary functions own their arguments. An arrow captures its outer arguments.
		expectZip("input.js", `function run(names) { `+argumentHeader+` { save(names[index], arguments[index]); function nested() { use(arguments); } } }`,
			`function run(names) { for (const [name, argument] of Iterator.zip([names, arguments])) { save(name, argument); function nested() { use(arguments); } } }`, 1, 23, 1, 101),
		// Only bindings visible to the rewrite can shadow a built-in.
		expectZip("input.js", `function run(Iterator, names, scores) { `+loopHeader+loopBody+` }`, "", 1, 41, 1, 116),
		expectZip("input.js", `function run(Math, names, scores) { `+loopHeader+loopBody+` }`, "", 1, 37, 1, 112),
		expectZip("input.js", "const Math = custom; const length = "+minimum+";\n"+cached+loopBody, "", 2, 1, 2, 45),
		expectZip("input.ts", `type Iterator = string; `+loopHeader+loopBody,
			`type Iterator = string; `+suggestion+` { save(name, score); }`, 1, 25, 1, 100),
		expectZip("input.js", `for (let Iterator = 0; Iterator < Math.min(names.length, scores.length); Iterator++) { save(names[Iterator], scores[Iterator]); }`,
			suggestion+` { save(name, score); }`, 1, 1, 1, 85),
		expectZip("input.js", loopHeader+` { const Iterator = 0; save(Iterator, names[index], scores[index]); }`,
			suggestion+` { const Iterator = 0; save(Iterator, name, score); }`, 1, 1, 1, 76),
		// JSDoc-annotated arrays and typed arrays still qualify.
		expectZip("input.js", annotatedArrays+loopHeader+loopBody+"\n}",
			annotatedArrays+suggestion+" { save(name, score); }\n}", 6, 3, 6, 78),
		// Direct eval can read bindings even without a static reference.
		expectZip("input.js", loopHeader+` { save(eval("index"), names[index], scores[index]); }`, "", 1, 1, 1, 76),
		expectZip("input.js", loopHeader+` { save(() => (eval)("index"), names[index], scores[index]); }`, "", 1, 1, 1, 76),
		expectZip("input.js", `for (let index = 0, length = `+minimum+`; index < length; index++) { save(eval("length"), names[index], scores[index]); }`, "", 1, 1, 1, 93),
		expectZip("input.js", "const length = "+minimum+";\n"+cached+loopBody+` save(eval("length"));`, "", 2, 1, 2, 45),
		expectZip("input.js", "switch (kind) { case 0:\n  const length = "+minimum+";\n  "+cached+loopBody+"\ncase 1: save(eval('length'));\n}", "", 3, 3, 3, 47),
		expectZip("input.js", loopHeader+` { save(eval?.("index"), (0, eval)("index"), names[index], scores[index]); }`,
			suggestion+` { save(eval?.("index"), (0, eval)("index"), name, score); }`, 1, 1, 1, 76),
	}
	disabledIterator := expectZip("input.js", loopHeader+loopBody, "", 1, 1, 1, 76)
	disabledIterator.Globals = map[string]any{"Iterator": "off"}
	withLoop := expectZip("input.js", `with (env) { `+loopHeader+loopBody+` }`, "", 1, 14, 1, 89)
	withRead := expectZip("input.js", loopHeader+` { with (env) { save(names[index], scores[index]); } }`, "", 1, 1, 1, 76)
	withLoop.LanguageOptions = rule.LanguageOptions{SourceType: "script"}
	withRead.LanguageOptions = withLoop.LanguageOptions
	invalid = append(invalid, disabledIterator, withLoop, withRead)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_iterator_zip.PreferIteratorZipRule, valid, invalid)
}

func TestPreferIteratorZipEditDemand(t *testing.T) {
	const commentedHeader = `for (let index = 0; /* keep */ index < Math.min(names.length, scores.length); index++)`
	const code = loopHeader + loopBody + "\n" + commentedHeader + loopBody
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "edit-demand.js")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	sourceProgram, err := program.NewFromRoots(program.RootOptions{
		RootFileNames: []string{fileName}, Host: utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: program.SourceOnlyCompilerOptions(), SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	target := prefer_iterator_zip.PreferIteratorZipRule
	lint := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram, File: fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: target.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return target.Run(ctx, nil)
				}}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, d)
			}},
		})
		return diagnostics
	}
	all := lint(rule.EditDemandAll)
	if len(all) != 2 {
		t.Fatalf("got %d diagnostics, want 2", len(all))
	}
	for i, header := range []string{loopHeader, commentedHeader} {
		got := all[i]
		if got.FixesPtr != nil || code[got.Range.Pos():got.Range.End()] != header {
			t.Fatalf("unexpected diagnostic for %q: %+v", header, got)
		}
	}
	if all[0].Suggestions == nil || len(*all[0].Suggestions) != 1 || all[1].Suggestions != nil {
		t.Fatal("expected one suggestion for the header without a comment")
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		want := append([]rule.RuleDiagnostic(nil), all...)
		if demand != rule.EditDemandSuggestion {
			want[0].Suggestions = nil
		}
		if got := lint(demand); !reflect.DeepEqual(got, want) {
			t.Errorf("edit demand %d changed diagnostics or attached unexpected edits", demand)
		}
	}
	suggestion := (*all[0].Suggestions)[0]
	if suggestion.Message.Id != "prefer-iterator-zip/suggestion" || suggestion.Message.Description != "Use `Iterator.zip(…)`." {
		t.Errorf("unexpected suggestion message: %+v", suggestion.Message)
	}
	output, _, _ := linter.ApplyRuleFixes(code, []rule.RuleSuggestion{suggestion})
	if want := `for (const [name, score] of Iterator.zip([names, scores])) { save(name, score); }` + "\n" + commentedHeader + loopBody; output != want {
		t.Errorf("got suggestion output %q, want %q", output, want)
	}
}
