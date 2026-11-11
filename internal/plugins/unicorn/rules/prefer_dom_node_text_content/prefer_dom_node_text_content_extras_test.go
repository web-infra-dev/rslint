package prefer_dom_node_text_content_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_dom_node_text_content"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
)

func TestPreferDomNodeTextContentExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: `class View { #innerText = ''; read() { return this.#innerText; } }`},
		{Code: `const {'innerText': text, ...innerText} = source;`},
		{Code: `let text; ({'innerText': text, ...innerText} = source);`},
		// Assignment targets can contain ordinary object literals in receivers or default values.
		{Code: `([({innerText}).text] = source); ({x: text = {innerText}} = source); ({[call({innerText})]: text} = source);`},
		{Tsx: true, Code: `const view = <ui.innerText><ui.innerText /></ui.innerText>;`},
		{Code: `type Query = typeof ns.innerText; type Reference = ns.innerText;`},
		{Code: `interface Base {innerText: string} interface Shape extends Base {} type Alias = Shape; declare const value: Alias; const alias = value; (alias).innerText;`},
		{Code: `class Base {innerText = ''} const Alias = Base; class Shape extends Alias {} declare const value: Shape; value.innerText;`},
		{Code: `declare const item: {innerText: string}; (item as any).innerText; (item satisfies {innerText: string}).innerText; item!.innerText;`},
		{Code: `interface A {innerText: string} interface B {innerText: string; extra: true} declare const union: A | B; declare const intersection: A & B; union.innerText; intersection.innerText;`},
		{Code: `declare function value(): {innerText: string}; value().innerText;`},
		{Code: `import {Node as Model} from './library'; declare const x: Model; x.innerText;`},
		{FileName: "case.js", Code: `/** @type {{innerText: string}} */ const value = {innerText: ''}; value.innerText;`},
		{FileName: "case.js", Code: `/** @param {{innerText: string}} value */ function read(value) { return value.innerText; }`},
		{FileName: "case.js", Code: `/** @returns {{innerText: string}} */ function make() { return {innerText: ''}; } make().innerText;`},
		{FileName: "case.js", Code: `/** @type {function(): string} */ const f = () => ''; f.innerText;`},
		{FileName: "case.js", Code: `/** @type {function()} */ const f = () => ''; f.innerText;`},
		{FileName: "case.js", Code: `/** @returns {function(): string} */ function make() { return () => ''; } make().innerText;`},
		{FileName: "case.js", Code: `/** @returns {HTMLElement} */ const make = /** @returns {{innerText: string}} */ () => value; make().innerText;`},
		{FileName: "case.js", Code: `/** @param {HTMLElement} value */ const read = /** @param {{innerText: string}} value */ value => value.innerText;`},
		{FileName: "case.js", Code: `/** @type {function(): {innerText: string}} */ const make = () => ({innerText: ''}); make().innerText;`},
		{FileName: "case.js", Code: `/** @param {function(): {innerText: string}} make */ function read(make) { make.innerText; return make().innerText; }`},
		{FileName: "case.js", Code: `/** @param make {function(): {innerText: string}} */ function read(make) { return make().innerText; }`},
		{FileName: "case.js", Code: `
/** @param {function():
 *   {innerText: '{'}} make */
function read(other, make = () => ({innerText: '{'})) { return make().innerText; }
`},
		{FileName: "case.js", Code: `/** @type {() => {innerText: string}} */ const make = /** @returns {HTMLElement} */ () => node; make().innerText;`},
		{FileName: "case.js", Code: `/** @param {() => {innerText: string}} make */ function read(make) { return make().innerText; }`},
		{FileName: "case.js", Code: `/** @param {() => {innerText: string}} make */ function read(make = () => ({innerText: ''})) { return make().innerText; }`},
		{FileName: "case.js", Code: `
/** @type {(() => {innerText: string}) | (() => {innerText: number})} */ const union = value;
/** @type {(() => {innerText: string}) & {property: true}} */ const intersection = value;
/** @type {{(): {innerText: string}}} */ const signature = value;
union().innerText; intersection().innerText; signature().innerText;
`},
	}
	invalid := []rule_tester.InvalidTestCase{
		{FileName: "case.js", Code: "node.\n  inn\\u0065rText;", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(2, 3, 17, "node.\n  textContent;"),
		}},
		{FileName: "case.js", Code: `("🦄", (node))?. /* keep */ innerText;`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 29, 38, `("🦄", (node))?. /* keep */ textContent;`),
		}},
		{FileName: "case.js", Code: `let innerText; for ({innerText} in source) {}`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 22, 31, `let innerText; for ({textContent: innerText} in source) {}`),
		}},
		{FileName: "case.js", Code: `let innerText; ([{innerText}] = source);`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 19, 28, `let innerText; ([{textContent: innerText}] = source);`),
		}},
		{FileName: "case.js", Code: `const {inn\u0065rText /* keep */ = 'fallback'} = node;`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 8, 22, `const {textContent: innerText /* keep */ = 'fallback'} = node;`),
		}},
		{FileName: "case.js", Code: "/** @type {other.Factory} */ const make = value;\nmake().innerText;", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(2, 8, 17, "/** @type {other.Factory} */ const make = value;\nmake().textContent;"),
		}},
		// Nested assignment defaults leave the object pattern's source unknown.
		{Code: `declare const fallback: {innerText: string}; let innerText; ({entry: {innerText} = fallback} = source);`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 71, 80, `declare const fallback: {innerText: string}; let innerText; ({entry: {textContent: innerText} = fallback} = source);`),
		}},
		{Code: `function read<T extends HTMLElement>({innerText}: T) { return innerText; }`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 39, 48, `function read<T extends HTMLElement>({textContent: innerText}: T) { return innerText; }`),
		}},
		{Code: `declare const elem: HTMLElement | {innerText: string}; elem.innerText;`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 61, 70, `declare const elem: HTMLElement | {innerText: string}; elem.textContent;`),
		}},
		{Code: `type T = typeof ns.innerText; interface Shape extends ns.innerText {} class View implements ns.innerText {}`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 58, 67, `type T = typeof ns.innerText; interface Shape extends ns.textContent {} class View implements ns.innerText {}`),
			textContentError(1, 96, 105, `type T = typeof ns.innerText; interface Shape extends ns.innerText {} class View implements ns.textContent {}`),
		}},
		{Code: "declare const model: {innerText: string};\n(model || model).innerText;\n(model ?? node).innerText;\n(test ? model : model).innerText;", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(3, 17, 26, "declare const model: {innerText: string};\n(model || model).innerText;\n(model ?? node).textContent;\n(test ? model : model).innerText;"),
		}},
		{Code: "type Cycle = Cycle;\ndeclare const unknown: Cycle;\ndeclare const model: {innerText: string};\nunknown.innerText;\nmodel.innerText; model.innerText;", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(4, 9, 18, "type Cycle = Cycle;\ndeclare const unknown: Cycle;\ndeclare const model: {innerText: string};\nunknown.textContent;\nmodel.innerText; model.innerText;"),
		}},
		// JSDoc on the destructuring declaration does not type the pattern for upstream's source-only check.
		{FileName: "case.js", Code: `/** @type {{innerText: string}} */ const {innerText} = node;`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 43, 52, `/** @type {{innerText: string}} */ const {textContent: innerText} = node;`),
		}},
		{FileName: "case.js", Code: `/** @param {{innerText: string}} value */ function read({innerText}) { return innerText; }`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 58, 67, `/** @param {{innerText: string}} value */ function read({textContent: innerText}) { return innerText; }`),
		}},
		{Code: `/** @type {{innerText: string}} */ const value = (node as HTMLElement)!; value.innerText;`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 80, 89, `/** @type {{innerText: string}} */ const value = (node as HTMLElement)!; value.textContent;`),
		}},
		{FileName: "case.js", Code: `/** @type {{innerText: string}} */ const value = node, extra = node; value.innerText; extra.innerText;`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 76, 85, `/** @type {{innerText: string}} */ const value = node, extra = node; value.textContent; extra.innerText;`),
			textContentError(1, 93, 102, `/** @type {{innerText: string}} */ const value = node, extra = node; value.innerText; extra.textContent;`),
		}},
		{FileName: "case.js", Code: "/** @type {{innerText: string}}\n * @type {HTMLElement}\n */ const value = node; value.innerText;", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(3, 31, 40, "/** @type {{innerText: string}}\n * @type {HTMLElement}\n */ const value = node; value.textContent;"),
		}},
		{FileName: "case.js", Code: `/** @param {() => {innerText: string}} make */ function read(make = async () => model) { return make().innerText; }`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 104, 113, `/** @param {() => {innerText: string}} make */ function read(make = async () => model) { return make().textContent; }`),
		}},
		{FileName: "case.js", Code: `/** @param {() => {innerText: string}} make */ function read(make = function* () { yield model; }) { return make().innerText; }`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 116, 125, `/** @param {() => {innerText: string}} make */ function read(make = function* () { yield model; }) { return make().textContent; }`),
		}},
		{Code: `/** @param {() => {innerText: string}} make */ function read(make = (() => node) as () => HTMLElement) { return make().innerText; }`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 120, 129, `/** @param {() => {innerText: string}} make */ function read(make = (() => node) as () => HTMLElement) { return make().textContent; }`),
		}},
		{Code: `/** @param {() => {innerText: string}} make */ function read(make = (): HTMLElement => node) { return make().innerText; }`, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(1, 110, 119, `/** @param {() => {innerText: string}} make */ function read(make = (): HTMLElement => node) { return make().textContent; }`),
		}},
		{FileName: "case.js", Code: "/** @type {function} */ const f = () => '';\nf.innerText;", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(2, 3, 12, "/** @type {function} */ const f = () => '';\nf.textContent;"),
		}},
		{FileName: "case.js", Code: "/** @param {function(): {innerText: string}} make */ function read(other, make) {\n  other().innerText; make().innerText;\n}", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(2, 11, 20, "/** @param {function(): {innerText: string}} make */ function read(other, make) {\n  other().textContent; make().innerText;\n}"),
		}},
		{FileName: "case.js", Code: "/** @param {function(): {innerText: string}} [make=identity] */ function read(make) {\n  make.innerText;\n  make().innerText;\n}", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(3, 10, 19, "/** @param {function(): {innerText: string}} [make=identity] */ function read(make) {\n  make.innerText;\n  make().textContent;\n}"),
		}},
		{FileName: "case.js", Code: "/** @param {function(): {innerText: string}} make */ function read(make) {\n  make = replacement;\n  make().innerText;\n}", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(3, 10, 19, "/** @param {function(): {innerText: string}} make */ function read(make) {\n  make = replacement;\n  make().textContent;\n}"),
		}},
		{FileName: "case.js", Code: "/** @type {{}} */ const value = document.createElement('div');\nvalue.innerText;", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(2, 7, 16, "/** @type {{}} */ const value = document.createElement('div');\nvalue.textContent;"),
		}},
		{FileName: "case.js", Code: "/** @param {{}} value */ function read(value) {\n  value.innerText;\n}", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(2, 9, 18, "/** @param {{}} value */ function read(value) {\n  value.textContent;\n}"),
		}},
		{FileName: "case.js", Code: "/** @returns {{}} */ function make() { return document.createElement('div'); }\nmake().innerText;", Errors: []rule_tester.InvalidTestCaseError{
			textContentError(2, 8, 17, "/** @returns {{}} */ function make() { return document.createElement('div'); }\nmake().textContent;"),
		}},
	}
	for _, types := range [][2]string{{"{innerText: string}", "HTMLElement"}, {"HTMLElement", "{innerText: string}"}} {
		parameter := fmt.Sprintf("/** @param {%s} value\n * @param {%s} value\n * @param {{innerText: string}} other\n */ function read(value, other) {\n  value.innerText; other.innerText;\n}", types[0], types[1])
		returns := fmt.Sprintf("/** @returns {%s}\n * @returns {%s}\n */ function make() { return node; }\nmake().innerText;", types[0], types[1])
		invalid = append(invalid,
			rule_tester.InvalidTestCase{FileName: "case.js", Code: parameter, Errors: []rule_tester.InvalidTestCaseError{
				textContentError(5, 9, 18, strings.Replace(parameter, ".innerText", ".textContent", 1)),
			}},
			rule_tester.InvalidTestCase{FileName: "case.js", Code: returns, Errors: []rule_tester.InvalidTestCaseError{
				textContentError(4, 8, 17, strings.Replace(returns, ".innerText", ".textContent", 1)),
			}},
		)
	}
	testedRule := preferDomNodeTextContentWithoutTypes()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, valid, invalid)
}

func TestPreferDomNodeTextContentTypeInformation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_dom_node_text_content.PreferDomNodeTextContentRule,
		[]rule_tester.ValidTestCase{{
			Code: `const model = {innerText: 'data'}; model.innerText; const {innerText} = model;`,
		}},
		[]rule_tester.InvalidTestCase{{
			Code: "const div = document.createElement('div');\ndiv.innerText;",
			Errors: []rule_tester.InvalidTestCaseError{
				textContentError(2, 5, 14, "const div = document.createElement('div');\ndiv.textContent;"),
			},
		}},
	)
}

func TestPreferDomNodeTextContentClosureAnnotations(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	for _, annotation := range []string{
		"(function(): string)",
		"((function(): {innerText: string}))",
		"!function(): {innerText: string}",
		"(function(): {innerText: string}) | (function(): {innerText: number})",
		"(function(): {innerText: string}) & !{property: true}",
		"function(): function(): string",
		"function(function(): string): {innerText: string}",
		"function(this: HTMLElement, string=, ...Array<Array<function(): string>>): {innerText: string}",
		"function(): {function(): string, note: 'function(): {'}",
		"function(): {value: string\n function(): string}",
		"function(): {value: function(): string\n function(): string}",
		"function(): {value:\n function(): string}",
		"function(): {value: function()\n new(): string}",
	} {
		valid = append(valid,
			rule_tester.ValidTestCase{FileName: "case.js", Code: fmt.Sprintf("/** @type {%s} */ const make = source;\nmake.innerText;\nmake().innerText;", annotation)},
			rule_tester.ValidTestCase{FileName: "case.js", Code: fmt.Sprintf("/** @param {%s} make */ function read(other, make) {\n  make.innerText;\n  make().innerText;\n}", annotation)},
		)
	}
	valid = append(valid, rule_tester.ValidTestCase{
		FileName: "case.js",
		Code:     "/** @returns {(function(): function(): string)} */ function make() { return source; }\nmake().innerText;",
	})

	var invalid []rule_tester.InvalidTestCase
	for _, annotation := range []string{
		"?function(): {innerText: string}",
		"(function(): {innerText: string})=",
		"(function(): {innerText: string}) | (function(): HTMLElement)",
		"function(function(): string): HTMLElement",
		"function(): (function(): string) | HTMLElement",
		"function(): {}",
	} {
		binding := fmt.Sprintf("/** @type {%s} */ const make = source;\nmake().innerText;", annotation)
		parameter := fmt.Sprintf("/** @param {%s} make */ function read(make) {\n  make().innerText;\n}", annotation)
		invalid = append(invalid,
			rule_tester.InvalidTestCase{FileName: "case.js", Code: binding, Errors: []rule_tester.InvalidTestCaseError{
				textContentError(2, 8, 17, strings.Replace(binding, ".innerText", ".textContent", 1)),
			}},
			rule_tester.InvalidTestCase{FileName: "case.js", Code: parameter, Errors: []rule_tester.InvalidTestCaseError{
				textContentError(2, 10, 19, strings.Replace(parameter, ".innerText", ".textContent", 1)),
			}},
		)
	}
	for _, annotation := range []string{
		"(function)",
		"function(,): string",
		"function(this): string",
		"function(keyof HTMLElement): string",
		"function(): keyof HTMLElement",
		"(function(): {innerText: string})[]",
	} {
		code := fmt.Sprintf("/** @type {%s} */ const make = source;\nmake.innerText;", annotation)
		invalid = append(invalid, rule_tester.InvalidTestCase{FileName: "case.js", Code: code, Errors: []rule_tester.InvalidTestCaseError{
			textContentError(2, 6, 15, strings.Replace(code, ".innerText", ".textContent", 1)),
		}})
	}
	optional := "/** @param {(function(): {innerText: string})} [make=identity] */ function read(make) {\n  make.innerText;\n  make().innerText;\n}"
	invalid = append(invalid, rule_tester.InvalidTestCase{FileName: "case.js", Code: optional, Errors: []rule_tester.InvalidTestCaseError{
		textContentError(3, 10, 19, strings.Replace(optional, "make().innerText", "make().textContent", 1)),
	}})

	testedRule := preferDomNodeTextContentWithoutTypes()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &testedRule, valid, invalid)
}

func TestPreferDomNodeTextContentImportedTypes(t *testing.T) {
	directory := txtarfs.MustParseFile(t, "testdata/imported_types.txtar").Materialize(t, "typed")
	root := rule_tester.Root{Dir: tspath.NormalizePath(directory), FS: bundled.WrapFS(osvfs.FS())}
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &prefer_dom_node_text_content.PreferDomNodeTextContentRule,
		[]rule_tester.ValidTestCase{{
			FileName: "caller.ts",
			Code:     "import {Node as Model} from './library';\ndeclare const x: Model;\nx.innerText;\nnew Model().innerText;",
		}},
		[]rule_tester.InvalidTestCase{{
			FileName: "caller.ts",
			Code:     "import {Element as View} from './library';\ndeclare const view: View;\nview.innerText;",
			Errors: []rule_tester.InvalidTestCaseError{
				textContentError(3, 6, 15, "import {Element as View} from './library';\ndeclare const view: View;\nview.textContent;"),
			},
		}},
	)
}

func TestPreferDomNodeTextContentEditDemand(t *testing.T) {
	t.Parallel()
	const source = `interface Value {innerText: string}
declare const value: Value;
value.innerText;
declare function make(): Value;
make?.().innerText;
(make?.()).innerText;
node?.innerText;
const {innerText} = node;
`
	wantStarts := []int{
		strings.Index(source, "(make?.()).innerText") + len("(make?.())."),
		strings.Index(source, "node?.innerText") + len("node?."),
		strings.Index(source, "const {innerText}") + len("const {"),
	}
	compiler, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(source, "edit-demand.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	if file == nil {
		t.Fatal("missing edit-demand.ts")
	}
	sourceProgram, err := program.NewFromBoundSources(compiler, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	native := prefer_dom_node_text_content.PreferDomNodeTextContentRule
	var baseline []rule.RuleDiagnostic
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram, File: file.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: native.Name, Severity: rule.SeverityError, Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return native.Run(ctx, nil)
				}}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		})
		if len(diagnostics) != len(wantStarts) {
			t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(diagnostics), len(wantStarts))
		}
		if demand == rule.EditDemandAll {
			baseline = diagnostics
		}
		for i, diagnostic := range diagnostics {
			want := baseline[i]
			if diagnostic.Range.Pos() != wantStarts[i] {
				t.Errorf("demand %d: diagnostic %d starts at %d, want %d", demand, i, diagnostic.Range.Pos(), wantStarts[i])
			}
			if diagnostic.RuleName != want.RuleName || diagnostic.Range != want.Range || diagnostic.Severity != want.Severity || !reflect.DeepEqual(diagnostic.Message, want.Message) {
				t.Errorf("demand %d changed diagnostic %d", demand, i)
			}
			if diagnostic.FixesPtr != nil {
				t.Errorf("demand %d attached an autofix", demand)
			}
			if demand&rule.EditDemandSuggestion == 0 {
				if diagnostic.Suggestions != nil {
					t.Errorf("demand %d unexpectedly attached a suggestion", demand)
				}
				continue
			}
			if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 ||
				(*diagnostic.Suggestions)[0].Message.Description != "Switch to `.textContent`." ||
				!reflect.DeepEqual(diagnostic.Suggestions, want.Suggestions) {
				t.Errorf("demand %d: invalid suggestion on diagnostic %d", demand, i)
			}
		}
	}
}
