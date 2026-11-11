package prefer_math_min_max_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_math_min_max"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestPreferMathMinMaxExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{FileName: "input.js", Code: `value === limit ? value : limit;`},
		{FileName: "input.js", Code: `value && limit ? value : limit;`},
		{FileName: "input.ts", Code: `((value > limit) as boolean) ? value : limit;`},
		{FileName: "input.js", Code: `obj["key"] < limit ? obj.key : limit;`},
		{FileName: "input.js", Code: `amount > 0 ? \u0061mount : 0;`},
		{FileName: "input.js", Code: `(obj?.key) < limit ? obj.key : limit;`},
		{FileName: "input.ts", Code: `(source satisfies number) > 0 ? source : 0;`},
		{FileName: "input.ts", Code: `function cap(n: Number) { return (n as Number) > 0 ? n : 0; }`},
		{FileName: "input.ts", Code: `function cap(value: (string)) { return value > 'a' ? value : 'a'; }`},
		{FileName: "input.ts", Code: `const { value }: { value: number } = source; value < 0 ? value : 0;`},
		{FileName: "input.js", Code: `function pick({ value = "zero" }) { return value > 0 ? value : 0; }`},
		{FileName: "input.js", Code: `const limit = /foo/; limit > value ? limit : value;`},
		{FileName: "input.js", Code: `const v = BigInt.asUintN(64, raw); v > 0 ? v : 0;`},
		{FileName: "input.js", Code: `const v = flag ? BigInt(raw) : 0; v > 0 ? v : 0;`},
		{FileName: "input.ts", Code: `const v = source satisfies bigint; v > 0 ? v : 0;`},
		{FileName: "input.ts", Code: `function pick<T extends number | bigint>(obj: { value: T }) { return obj.value > 0 ? obj.value : 0; }`},
	}

	disabledMath := minMaxCase("input.js", `height > 50 ? 50 : height;`, "min", 1, 1, 1, 26, "")
	disabledMath.Globals = map[string]any{"Math": "off"}
	invalid := []rule_tester.InvalidTestCase{
		minMaxCase("input.js", `value >= value ? value : value;`, "min", 1, 1, 1, 31, `Math.min(value, value);`),
		minMaxCase("input.js", `(obj?.size) <= limit ? obj?.size : limit;`, "min", 1, 1, 1, 41, `Math.min(obj?.size, limit);`),
		minMaxCase("input.js", `obj?.read() <= limit ? obj?.read() : limit;`, "min", 1, 1, 1, 43, ""),
		minMaxCase("input.ts", `class Box { #size = 1; min(limit: number) { return this.#size < limit ? this.#size : limit; } }`, "min", 1, 52, 1, 91,
			`class Box { #size = 1; min(limit: number) { return Math.min(this.#size, limit); } }`),
		minMaxCase("input.tsx", `const view = <span>🦄{n >= 0 ? n : 0}</span>;`, "max", 1, 23, 1, 37, `const view = <span>🦄{Math.max(n, 0)}</span>;`),
		minMaxCase("input.js", `function cap(height) { return (/** @type {number} */ (height)) > 50 ? 50 : height; }`, "min", 1, 31, 1, 82,
			`function cap(height) { return Math.min((/** @type {number} */ (height)), 50); }`),
		minMaxCase("input.js", `const value = "text"; function clamp(value) { return value < 0 ? 0 : value; }`, "max", 1, 54, 1, 75,
			`const value = "text"; function clamp(value) { return Math.max(value, 0); }`),
		minMaxCase("input.ts", `function pick({ value }: { value: number }, limit: number) { return value > limit ? value : limit; }`, "max", 1, 69, 1, 98,
			`function pick({ value }: { value: number }, limit: number) { return Math.max(value, limit); }`),
		minMaxCase("input.ts", `const value = "one"; try {} catch (value: unknown) { value > 0 ? value : 0; }`, "max", 1, 54, 1, 75,
			`const value = "one"; try {} catch (value: unknown) { Math.max(value, 0); }`),
		minMaxCase("input.ts", `try {} catch ({value}: any) { value > limit ? value : limit; }`, "max", 1, 31, 1, 60,
			`try {} catch ({value}: any) { Math.max(value, limit); }`),
		minMaxCase("input.js", `for (const item of(+a > b ? +a : b)) use(item);`, "max", 1, 20, 1, 35,
			`for (const item of (Math.max(+a, b))) use(item);`),
		minMaxCase("input.ts", `const v = (x < y ? x : y)satisfies number;`, "min", 1, 12, 1, 25,
			`const v = (Math.min(x, y)) satisfies number;`),
		minMaxCase("input.ts", `function cap(value: ((number)), limit: number) { return value > limit ? value : limit; }`, "max", 1, 57, 1, 86,
			`function cap(value: ((number)), limit: number) { return Math.max(value, limit); }`),
		minMaxCase("input.ts", `declare const value: (number); value > 0 ? 0 : value;`, "min", 1, 32, 1, 53,
			`declare const value: (number); Math.min(value, 0);`),
		minMaxCase("input.ts", `(value as (Number)) > limit ? value : limit;`, "max", 1, 1, 1, 44,
			`Math.max(value as (Number), limit);`),
		minMaxCase("input.ts", `(<(number)>value) < limit ? value : limit;`, "min", 1, 1, 1, 42,
			`Math.min(<(number)>value, limit);`),
		minMaxCase("input.ts", `const v = (a as (number))! < b ? a : b;`, "min", 1, 11, 1, 39,
			`const v = Math.min((a as (number))!, b);`),
		minMaxCase("input.ts", `/** @type {() => bigint} */
const read: { (): number } = () => 1;
const value = read();
value > 0 ? value : 0;`, "max", 4, 1, 4, 22,
			`/** @type {() => bigint} */
const read: { (): number } = () => 1;
const value = read();
Math.max(value, 0);`),
		minMaxCase("input.js", "const v = (obj[`a${index}`]) < limit ? obj[`a${index}`] : limit;", "min", 1, 11, 1, 64, ""),

		// Keep the diagnostic when the Math binding or repeated evaluation
		// would make the replacement unsafe.
		minMaxCase("input.js", `function cap(Math, height) { return height > 50 ? 50 : height; }`, "min", 1, 37, 1, 62, ""),
		disabledMath,
		minMaxCase("input.js", `const capped = (delete cache.key, height) > 50 ? 50 : (delete cache.key, height);`, "min", 1, 16, 1, 81, ""),
		minMaxCase("input.js", `const v = (class { static { touch(); } }) < 0 ? (class { static { touch(); } }) : 0;`, "min", 1, 11, 1, 84, ""),
		minMaxCase("input.js", `(class { value = read() }) > 0 ? 0 : (class { value = read() });`, "min", 1, 1, 1, 64, ""),
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_math_min_max.PreferMathMinMaxRule, valid, invalid)
}

func TestPreferMathMinMaxSourceOnlyJSDoc(t *testing.T) {
	const code = `/** @type {function(): bigint} */
const readBig = source;
const first = readBig();
first > 0 ? first : 0;

/** @param {function(): bigint} read */
function cap(read) {
  const value = read();
  return value > 0 ? value : 0;
}

const zero = 0; /** @returns {bigint} */ function readInline() { return 1n; }
const inline = readInline();
inline > zero ? inline : zero;

/** @type {function(): bigint} */
let readMutable = source;
readMutable = () => 1;
const last = readMutable();
last > 0 ? last : 0;`
	sourceProgram, fileName := sourceOnlyMinMaxProgram(t, "jsdoc.js", code)
	target := prefer_math_min_max.PreferMathMinMaxRule
	var diagnostics []rule.RuleDiagnostic
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program: sourceProgram, Files: []string{fileName},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name: target.Name, Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners { return target.Run(ctx, nil) },
			}}
		},
		OnDiagnostic: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) },
	})
	if len(diagnostics) != 1 {
		t.Fatalf("got %d diagnostics, want one for the reassigned callback", len(diagnostics))
	}
	diagnostic := diagnostics[0]
	if text := code[diagnostic.Range.Pos():diagnostic.Range.End()]; text != "last > 0 ? last : 0" {
		t.Fatalf("reported %q, want the comparison using the reassigned callback", text)
	}
}

func TestPreferMathMinMaxEditDemand(t *testing.T) {
	// Exercise annotation and scope behavior with no TypeScript project.
	const code = `type Wide = number | bigint;
declare function readWide(): Wide;
const wide = readWide();
wide > 0 ? wide : 0;
type Branded = bigint & { tag: true };
declare function readBranded(): Branded;
const branded = readBranded();
branded > 0 ? branded : 0;
function outer<T extends bigint, U extends T>(input: U) {
  function read(): T { return input; }
  const constrained = read();
  constrained > 0 ? constrained : 0;
  type Value = U;
  function readAlias(): Value { return input; }
  const aliased = readAlias();
  aliased > 0 ? aliased : 0;
}
function readIntersection(): bigint & unknown { return 1n; }
const intersection = readIntersection();
intersection > 0 ? intersection : 0;
type Any = any;
function readAny(): bigint & Any { return 1; }
const numeric = readAny();
numeric > 0 ? numeric : 0;
type NestedAny = Any & unknown;
function readNestedAny(): bigint & NestedAny { return 1; }
const nestedNumeric = readNestedAny();
nestedNumeric > 0 ? nestedNumeric : 0;
type ReadNumber = () => number;
/** @type {() => bigint} */
const readAlias: ReadNumber = () => 1;
const fromAlias = readAlias();
fromAlias > 0 ? fromAlias : 0;
/** @type {() => bigint} */
const readDynamic: any = () => 1;
const fromDynamic = readDynamic();
fromDynamic > 0 ? fromDynamic : 0;
declare const readBig: () => bigint;
const big = readBig();
big > 0 ? big : 0;
function cap(read: (() => bigint)) {
  const wrapped = read();
  return wrapped > 0 ? wrapped : 0;
}
declare function readSigned(): -1n;
const signed = readSigned();
signed > 0 ? signed : 0;
let stable = (): bigint => 1n;
const stableValue = stable();
stableValue > 0 ? stableValue : 0;
/** @returns {bigint} */ function read() { return 1n; }
for (read of callbacks) {
  const value = read();
  value > 0 ? value : 0;
}
read() > 0 ? read() : 0;
type Cycle = Cycle;
declare function readCycle(): Cycle;
const uncertain = readCycle();
uncertain > 0 ? uncertain : 0;
` + "String.raw`value` > 0 ? String.raw`value` : 0;"
	comparisons := []string{
		"numeric > 0 ? numeric : 0",
		"nestedNumeric > 0 ? nestedNumeric : 0",
		"fromAlias > 0 ? fromAlias : 0",
		"fromDynamic > 0 ? fromDynamic : 0",
		"value > 0 ? value : 0",
		"read() > 0 ? read() : 0",
		"uncertain > 0 ? uncertain : 0",
		"String.raw`value` > 0 ? String.raw`value` : 0",
	}
	replacements := []string{
		"Math.max(numeric, 0)",
		"Math.max(nestedNumeric, 0)",
		"Math.max(fromAlias, 0)",
		"Math.max(fromDynamic, 0)",
		"Math.max(value, 0)",
		"",
		"Math.max(uncertain, 0)",
		"",
	}
	sourceProgram, fileName := sourceOnlyMinMaxProgram(t, "edit-demand.ts", code)
	target := prefer_math_min_max.PreferMathMinMaxRule
	results := make(map[rule.EditDemand][]rule.RuleDiagnostic)
	for _, demand := range []rule.EditDemand{rule.EditDemandAll, rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: sourceProgram, File: fileName,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: target.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						if ctx.TypeChecker != nil {
							t.Fatal("expected a project-free file")
						}
						return target.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) {
				results[demand] = append(results[demand], d)
			}},
		})
		if len(results[demand]) != len(comparisons) {
			t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(results[demand]), len(comparisons))
		}
		for i, got := range results[demand] {
			if comparison := code[got.Range.Pos():got.Range.End()]; comparison != comparisons[i] {
				t.Errorf("demand %d: got comparison %q, want %q", demand, comparison, comparisons[i])
			}
			want := results[rule.EditDemandAll][i]
			if demand&rule.EditDemandSuggestion == 0 {
				want.Suggestions = nil
			}
			if got.FixesPtr != nil || !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d: unexpected diagnostic or edits at index %d", demand, i)
			}
		}
	}
	all := results[rule.EditDemandAll]
	for i, diagnostic := range all {
		replacement := replacements[i]
		if replacement == "" {
			if diagnostic.Suggestions != nil {
				t.Errorf("unexpected suggestion for %q", comparisons[i])
			}
			continue
		}
		if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
			t.Fatalf("expected one suggestion for %q", comparisons[i])
		}
		output, _, _ := linter.ApplyRuleFixes(code, *diagnostic.Suggestions)
		if want := strings.Replace(code, comparisons[i], replacement, 1); output != want {
			t.Errorf("suggestion for %q produced %q, want %q", comparisons[i], output, want)
		}
	}
}

func sourceOnlyMinMaxProgram(t *testing.T, name, code string) (*program.Program, string) {
	t.Helper()
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, name)
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	sourceProgram, err := program.NewFromRoots(program.RootOptions{
		RootFileNames: []string{fileName}, Host: utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: program.SourceOnlyCompilerOptions(), SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return sourceProgram, fileName
}
