// TestRequireAwaitExtras locks in branches and edge shapes that the upstream suites don't exercise.
// The migrated upstream cases live in require_await_upstream_test.go.
package require_await

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func requireAwaitSuggestion(source string, occurrence int) []rule_tester.InvalidTestCaseSuggestion {
	start := 0
	for current := 0; ; current++ {
		index := strings.Index(source[start:], "async")
		if index < 0 {
			panic(fmt.Sprintf("async occurrence %d not found", occurrence))
		}
		index += start
		end := index + len("async")
		if current == occurrence {
			for end < len(source) {
				switch source[end] {
				case ' ', '\t', '\r', '\n':
					end++
				default:
					return []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeAsync",
						Output:    source[:index] + source[end:],
					}}
				}
			}
			return []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "removeAsync",
				Output:    source[:index],
			}}
		}
		start = end
	}
}

func addRequireAwaitSuggestions(cases []rule_tester.InvalidTestCase) {
	for caseIndex := range cases {
		for errorIndex := range cases[caseIndex].Errors {
			cases[caseIndex].Errors[errorIndex].Suggestions = requireAwaitSuggestion(
				cases[caseIndex].Code,
				errorIndex,
			)
		}
	}
}

func withRequireAwaitSuggestions(cases []rule_tester.InvalidTestCase) []rule_tester.InvalidTestCase {
	addRequireAwaitSuggestions(cases)
	return cases
}

func TestRequireAwaitExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		nil,
		[]rule_tester.InvalidTestCase{{
			Code: "async function parenthesized(): (Promise<number>) { return 1; }",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "missingAwait",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeAsync",
					Output:    "function parenthesized(): (number) { return 1; }",
				}},
			}},
		}},
	)
}

func TestRequireAwaitDiagnosticPayloads(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		nil,
		withRequireAwaitSuggestions([]rule_tester.InvalidTestCase{
			{
				Code: `class DampingSwipe {
  protected async isSwipeHorizontalDisAllow(left: number) {
    return left < 0;
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'isSwipeHorizontalDisAllow' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 44,
				}},
			},
			{
				Code: `const ClipboardHelper = {
  async loadClipboardSecuritySDK() {
    return;
  },
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'loadClipboardSecuritySDK' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 33,
				}},
			},
			{
				Code: `const mockStore = {
  async getAllKeys() {
    return [];
  },
  async getItem() {
    return {};
  },
  async getItems() {
    return [];
  },
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async method 'getAllKeys' has no 'await' expression.",
						Line:      2,
						Column:    3,
						EndLine:   2,
						EndColumn: 19,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method 'getItem' has no 'await' expression.",
						Line:      5,
						Column:    3,
						EndLine:   5,
						EndColumn: 16,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method 'getItems' has no 'await' expression.",
						Line:      8,
						Column:    3,
						EndLine:   8,
						EndColumn: 17,
					},
				},
			},
			{
				Code: `const manager = {
  async warmup() {
    return {};
  },
  async preload() {
    return;
  },
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async method 'warmup' has no 'await' expression.",
						Line:      2,
						Column:    3,
						EndLine:   2,
						EndColumn: 15,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method 'preload' has no 'await' expression.",
						Line:      5,
						Column:    3,
						EndLine:   5,
						EndColumn: 16,
					},
				},
			},
			{
				Code: `function render() {
  const { data: meta, loading: isSearchingMetaData } = useRequest(async () => {
    return 1;
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function has no 'await' expression.",
					Line:      2,
					Column:    76,
					EndLine:   2,
					EndColumn: 78,
				}},
			},
			{
				Code: `const element = (
  <LazyReadDocChatInput
          onBeforeSendMessage={async () =>
            true
          }
  />
);`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function has no 'await' expression.",
					Line:      3,
					Column:    41,
					EndLine:   3,
					EndColumn: 43,
				}},
			},
			{
				Code: `const rawSubmit = useSubmit({
    onBeforeSendMessage: async () => onBeforeSendMessage?.() ?? true,
});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'onBeforeSendMessage' has no 'await' expression.",
					Line:      2,
					Column:    5,
					EndLine:   2,
					EndColumn: 32,
				}},
			},
			{
				Code: `const getGlobalInfoService = async () => {
  return getGlobalContainer();
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function 'getGlobalInfoService' has no 'await' expression.",
					Line:      1,
					Column:    39,
					EndLine:   1,
					EndColumn: 41,
				}},
			},
			{
				Code: `class ActionManager {
  runResetTaskFn = async () => {
    return;
  };
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'runResetTaskFn' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 26,
				}},
			},
			{
				Code: `const commentSDK = {
  commentSDKManager: async () => {
    return {};
  },
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'commentSDKManager' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 28,
				}},
			},
			{
				Code: `const wrapped = {
  field: ((async () => 1)),
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'field' has no 'await' expression.",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 18,
				}},
			},
			{
				Code: `class AutoAccessor {
  accessor field = async () => 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function has no 'await' expression.",
					Line:      2,
					Column:    29,
					EndLine:   2,
					EndColumn: 31,
				}},
			},
			{
				Code: `class PrivateMembers {
  static async #method() {
    return 1;
  }
  static #field = async () => 1;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Static private async method #method has no 'await' expression.",
					},
					{
						MessageId: "missingAwait",
						Message:   "Static private async method #field has no 'await' expression.",
					},
				},
			},
			{
				Code: `const computed = {
  ['field']: async () => 1,
};`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async method 'field' has no 'await' expression.",
				}},
			},
			{
				Code: `class EmptyNames {
  async ''() {
    return 1;
  }
  '' = async () => 2;
}
const object = {
  '': async () => 3,
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
					},
				},
			},
			{
				Code: `declare function consume(value: unknown): void;
const dynamic = 1;
const object = {
  [dynamic]: async function ownName() { consume(1); },
  [1 + 1]: async () => consume(2),
  [true ? 4 : dynamic]: async () => consume(3),
  [String(5)]: async () => consume(4),
  [({ key: 6 }).key]: async () => consume(5),
};
class Fields {
  [dynamic] = async function ownField() { consume(6); };
  [1 + 2] = async () => consume(7);
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
						Line:      4,
						Column:    3,
						EndLine:   4,
						EndColumn: 36,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method '2' has no 'await' expression.",
						Line:      5,
						Column:    3,
						EndLine:   5,
						EndColumn: 18,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method '4' has no 'await' expression.",
						Line:      6,
						Column:    3,
						EndLine:   6,
						EndColumn: 31,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
						Line:      7,
						Column:    3,
						EndLine:   7,
						EndColumn: 22,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method '6' has no 'await' expression.",
						Line:      8,
						Column:    3,
						EndLine:   8,
						EndColumn: 29,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method has no 'await' expression.",
						Line:      11,
						Column:    3,
						EndLine:   11,
						EndColumn: 38,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async method '3' has no 'await' expression.",
						Line:      12,
						Column:    3,
						EndLine:   12,
						EndColumn: 19,
					},
				},
			},
			{
				Code: `let assigned: () => Promise<void>;
assigned = async () => consume(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function 'assigned' has no 'await' expression.",
				}},
			},
			{
				Code: `function defaults(callback = async () => consume(1)) {
  return callback;
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function 'callback' has no 'await' expression.",
				}},
			},
			{
				Code: `export default async () => consume(1);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "missingAwait",
					Message:   "Async arrow function 'default' has no 'await' expression.",
				}},
			},
			{
				Code: `declare function consume(value: unknown): void;

function bindingElement({ callback = async () => consume(1) } = {}) {
  return callback;
}

let shorthand: () => Promise<void>;
({ shorthand = async () => consume(2) } = {});

let compound: (() => Promise<void>) | undefined;
compound ??= async () => consume(3);

export default async function () {
  consume(4);
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingAwait",
						Message:   "Async arrow function 'callback' has no 'await' expression.",
						Line:      3,
						Column:    47,
						EndLine:   3,
						EndColumn: 49,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async arrow function 'shorthand' has no 'await' expression.",
						Line:      8,
						Column:    25,
						EndLine:   8,
						EndColumn: 27,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async arrow function 'compound' has no 'await' expression.",
						Line:      11,
						Column:    23,
						EndLine:   11,
						EndColumn: 25,
					},
					{
						MessageId: "missingAwait",
						Message:   "Async function 'default' has no 'await' expression.",
						Line:      13,
						Column:    16,
						EndLine:   13,
						EndColumn: 31,
					},
				},
			},
		}),
	)
}

func TestRequireAwaitDeferredTypeChecks(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		[]rule_tester.ValidTestCase{
			{Code: `
declare function consume(value: number): number;
async function returnsAwait() {
  return consume(await Promise.resolve(1));
}
      `},
			{Code: `
declare function consume(value: number): number;
const arrow = async () => consume(await Promise.resolve(1));
      `},
			{Code: `
declare function consume(value: number): number;
async function* yieldsAwait() {
  yield consume(await Promise.resolve(1));
}
      `},
			{Code: `
async function* awaitsBeforeYield() {
  await Promise.resolve();
  yield Promise.resolve(1);
}
      `},
			{Code: `
interface StructuralThenable {
  then(onfulfilled: (value: number) => unknown): unknown;
}
declare const value: StructuralThenable;
async function returnsThenable() {
  return value;
}
      `},
		},
		withRequireAwaitSuggestions([]rule_tester.InvalidTestCase{
			{
				Code: `
interface FakeThenable {
  then(value: number): unknown;
}
declare const value: FakeThenable;
async function returnsFakeThenable() {
  return value;
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
interface FakeThenable {
  then(value: number): unknown;
}
declare const value: FakeThenable;
const returnsFakeThenable = async () => value;
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
interface FakeThenable {
  then(value: number): unknown;
}
declare const value: FakeThenable;
async function* yieldsFakeThenable() {
  yield value;
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
async function outer() {
  function* syncGenerator() {
    yield Promise.resolve(1);
  }
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
async function outer() {
  return async () => await Promise.resolve(1);
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
			{
				Code: `
async function* outer() {
  yield async () => await Promise.resolve(1);
}
        `,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingAwait"}},
			},
		}),
	)
}

func TestRequireAwaitDeepScopeStack(t *testing.T) {
	const depth = 40
	var source strings.Builder
	for i := range depth {
		source.WriteString("async function f")
		source.WriteString(strconv.Itoa(i))
		source.WriteString("() {\n")
	}
	source.WriteString("await Promise.resolve();\n")
	for range depth {
		source.WriteString("}\n")
	}

	errors := make([]rule_tester.InvalidTestCaseError, depth-1)
	for i := range errors {
		errors[i].MessageId = "missingAwait"
		errors[i].Suggestions = requireAwaitSuggestion(source.String(), depth-2-i)
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireAwaitRule,
		nil,
		[]rule_tester.InvalidTestCase{{
			Code:   source.String(),
			Errors: errors,
		}},
	)
}

func TestRequireAwaitEditDemand(t *testing.T) {
	const source = "async function value(): Promise<number> { return 1; }"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	compilerProgram, sourceFile, err := helper.CreateTestProgram(source, "require-await-edit-demand.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(compilerProgram),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:             RequireAwaitRule.Name,
					Severity:         rule.SeverityError,
					RequiresTypeInfo: RequireAwaitRule.RequiresTypeInfo,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return RequireAwaitRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone:       run(rule.EditDemandNone),
		rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
		rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
		rule.EditDemandAll:        run(rule.EditDemandAll),
	}
	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	wantIdentity := withoutEdits(diagnostics[rule.EditDemandAll])
	for demand, diagnostic := range diagnostics {
		if got := withoutEdits(diagnostic); !reflect.DeepEqual(got, wantIdentity) {
			t.Errorf("demand %d changed diagnostic identity:\ngot:  %#v\nwant: %#v", demand, got, wantIdentity)
		}
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d unexpectedly materialized autofixes", demand)
		}
	}
	if diagnostics[rule.EditDemandNone].Suggestions != nil || diagnostics[rule.EditDemandAutofix].Suggestions != nil {
		t.Fatal("suggestions were materialized without suggestion demand")
	}
	suggestionOnly := diagnostics[rule.EditDemandSuggestion].Suggestions
	allSuggestions := diagnostics[rule.EditDemandAll].Suggestions
	if suggestionOnly == nil || allSuggestions == nil || !reflect.DeepEqual(*suggestionOnly, *allSuggestions) {
		t.Fatalf("suggestion artifacts differ between suggestion-only and all demand")
	}
	if len(*suggestionOnly) != 1 || len((*suggestionOnly)[0].FixesArr) != 3 {
		t.Fatalf("suggestions = %#v, want one suggestion with three edits", *suggestionOnly)
	}
}

func TestRequireAwaitThenableFastPathParity(t *testing.T) {
	const source = `
interface GoodThenable {
  then(onfulfilled: (value: number) => unknown): unknown;
}
interface RestThenable {
  then(...callbacks: Array<(value: number) => unknown>): unknown;
}
interface UnionCallbackThenable {
  then(onfulfilled: ((value: number) => unknown) | { tag: string }): unknown;
}
interface BadThenable {
  then(value: number): unknown;
}
interface NoArgThenable {
  then(): unknown;
}
declare const good: GoodThenable;
declare const rest: RestThenable;
declare const unionCallback: UnionCallbackThenable;
declare const bad: BadThenable;
declare const noArg: NoArgThenable;

const caseNumber = 1;
const casePromise = Promise.resolve(1);
const caseGood = good;
const caseRest = rest;
const caseUnionCallback = unionCallback;
const caseBad = bad;
const caseNoArg = noArg;
const caseGoodUnion = null as number | GoodThenable;
const caseBadUnion = null as number | BadThenable;
const caseIntersection = null as GoodThenable & { tag: string };
`
	expected := map[string]bool{
		"caseNumber":        false,
		"casePromise":       true,
		"caseGood":          true,
		"caseRest":          true,
		"caseUnionCallback": true,
		"caseBad":           false,
		"caseNoArg":         false,
		"caseGoodUnion":     true,
		"caseBadUnion":      false,
		"caseIntersection":  true,
	}

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(source, "require-await-thenable-parity.ts", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()

	seen := make(map[string]bool, len(expected))
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsExpressionNode(node) {
			typ := typeChecker.GetTypeAtLocation(node)
			got := isThenableType(typeChecker, node, typ)
			shared := utils.IsThenableType(typeChecker, node, typ)
			if got != shared {
				t.Errorf(
					"expression kind %v at [%d,%d): fast path = %t, shared helper = %t",
					node.Kind,
					node.Pos(),
					node.End(),
					got,
					shared,
				)
			}
		}
		if node.Kind == ast.KindVariableDeclaration {
			name := node.Name()
			initializer := node.Initializer()
			if name != nil && initializer != nil {
				if want, ok := expected[name.Text()]; ok {
					typ := typeChecker.GetTypeAtLocation(initializer)
					got := isThenableType(typeChecker, initializer, typ)
					shared := utils.IsThenableType(typeChecker, initializer, typ)
					if got != shared {
						t.Errorf("%s: fast path = %t, shared helper = %t", name.Text(), got, shared)
					}
					if got != want {
						t.Errorf("%s: isThenableType = %t, want %t", name.Text(), got, want)
					}
					seen[name.Text()] = true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.AsNode().ForEachChild(visit)

	for name := range expected {
		if !seen[name] {
			t.Errorf("test case %s was not visited", name)
		}
	}
}

func TestRequireAwaitMissingProjectReferenceOutput(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/missing_project_reference_output.txtar")
	root := tspath.NormalizePath(archive.Materialize(t, "missing-project-reference-output"))
	appRoot := tspath.ResolvePath(root, "app")
	entryPath := tspath.ResolvePath(appRoot, "src/main.ts")
	dependencyPath := tspath.ResolvePath(root, "fixture-lib/src/index.ts")
	fs := bundled.WrapFS(osvfs.FS())
	compilerProgram, err := utils.CreateProgram(
		true,
		fs,
		appRoot,
		"tsconfig.json",
		utils.CreateCompilerHost(appRoot, fs),
	)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	if compilerProgram.GetSourceFile(dependencyPath) != nil {
		t.Fatalf("fixture precondition failed: project-reference source %q was loaded despite its missing declaration output", dependencyPath)
	}

	sourceProgram := lintprogram.NewFromCompiler(compilerProgram)
	var diagnostics []rule.RuleDiagnostic
	testutil.LintProgram(t, testutil.LintProgramOptions{
		Program:                sourceProgram,
		Files:                  []string{entryPath},
		ExcludedPathSubstrings: []string{},
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:             RequireAwaitRule.Name,
				Environment:      &rule.RuleEnvironment{},
				Severity:         rule.SeverityError,
				RequiresTypeInfo: true,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return RequireAwaitRule.Run(ctx, nil)
				},
			}}
		},
		OnDiagnostic: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	})

	wantLines := []int{25, 26, 27, 28, 29, 30, 31, 33, 35}
	if len(diagnostics) != len(wantLines) {
		t.Fatalf("diagnostics = %d (%v), want %d at lines %v", len(diagnostics), diagnostics, len(wantLines), wantLines)
	}
	for index, diagnostic := range diagnostics {
		line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(diagnostic.SourceFile, diagnostic.Range.Pos())
		if line+1 != wantLines[index] {
			t.Errorf("diagnostic %d line = %d, want %d", index+1, line+1, wantLines[index])
		}
	}
}

func TestRequireAwaitRecoveryAST(t *testing.T) {
	for index, source := range []string{
		`async function value() {`,
		`const value = async () => {`,
		`async function* value() {`,
		`async function outer() { function inner() {`,
		`class Value { async method() {`,
		`async function outer() { const inner = async () => {`,
	} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			fileName := "/require-await-recovery-" + strconv.Itoa(index) + ".ts"
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: fileName,
				Path:     tspath.Path(fileName),
			}, source, core.ScriptKindTS)
			ctx := rule.RuleContext{
				SourceFile:     sourceFile,
				DisableManager: rule.NewDisableManager(sourceFile, rule.NewCommentStore(sourceFile)),
			}.WithReporter(
				RequireAwaitRule.Name,
				rule.SeverityError,
				func(diagnostic rule.RuleDiagnostic) {
					if diagnostic.Range.Pos() < 0 ||
						diagnostic.Range.End() < diagnostic.Range.Pos() ||
						diagnostic.Range.End() > len(source) {
						t.Errorf(
							"out-of-bounds diagnostic range [%d,%d) for source length %d",
							diagnostic.Range.Pos(),
							diagnostic.Range.End(),
							len(source),
						)
					}
				},
			)
			listeners := RequireAwaitRule.Run(ctx, nil)

			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if listener := listeners[node.Kind]; listener != nil {
					listener(node)
				}
				node.ForEachChild(visit)
				if listener := listeners[rule.ListenerOnExit(node.Kind)]; listener != nil {
					listener(node)
				}
				return false
			}
			sourceFile.AsNode().ForEachChild(visit)
		})
	}
}
