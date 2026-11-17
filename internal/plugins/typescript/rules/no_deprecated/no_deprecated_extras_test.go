package no_deprecated

// cspell:ignore deprecatedish

import (
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoDeprecatedRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedRule, []rule_tester.ValidTestCase{
		{Code: `const value = 1; value;`},
		{
			Code: `
/** @deprecated */
const oldValue = 1;
oldValue;
      `,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{"oldValue"},
				},
			},
		},
		{
			Code: `
/** @deprecated */
const oldValue = 1;
oldValue;
      `,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{
							"from": "file",
							"name": "oldValue",
						},
					},
				},
			},
		},
		{
			Code: `
/** @deprecated */
const oldValue = 1;
oldValue;
      `,
			Options: map[string]interface{}{
				"allow": []interface{}{
					map[string]interface{}{
						"from": "file",
						"name": "oldValue",
					},
				},
			},
		},
		{
			// Regression: referencing a catch-clause binding must not panic.
			// The binding is a VariableDeclaration whose parent is a
			// CatchClause, not a VariableDeclarationList.
			Code: `
declare function log(x: unknown): void;
try {
  log(1);
} catch (error) {
  log(error);
}
`,
		},
		{
			// A computed access is allowed by the type at its receiver.
			Code: `
interface Thing {
  /** @deprecated */
  oldProp: string;
}
declare const thing: Thing;
thing['oldProp'];
`,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{"from": "file", "name": "Thing"},
					},
				},
			},
		},
		{
			// The receiver is written as the type parameter, not its constraint.
			Code: `
interface Thing {
  /** @deprecated */
  oldProp: string;
}
function g<T extends Thing>(t: T) {
  t['oldProp'];
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{"from": "file", "name": "T"},
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			// A computed key names no value, so it never matches a specifier.
			Code: `
interface Thing {
  /** @deprecated */
  oldProp: string;
}
declare const thing: Thing;
thing['oldProp'];
`,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{"oldProp"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated"},
			},
		},
		{
			Code: `
interface Thing {
  /** @deprecated */
  oldProp: string;
}
function g<T extends Thing>(t: T) {
  t['oldProp'];
}
`,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{"from": "file", "name": "Thing"},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated"},
			},
		},
		{
			Code: `/** @deprecated */ const oldValue = 1; oldValue;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 1, Column: 40},
			},
		},
		{
			Code: `/** @deprecated */ let oldValue = undefined; oldValue;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated", Line: 1, Column: 46},
			},
		},
		{
			Code: `
/** @deprecated Use newValue instead. */
const oldValue = 1;
oldValue;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecatedWithReason"},
			},
		},
		{
			Code: `
/** @deprecated */
const oldValue = 1;
oldValue;
      `,
			Options: []interface{}{
				map[string]interface{}{
					"allow": []interface{}{
						map[string]interface{}{
							"from":    "package",
							"name":    "oldValue",
							"package": "other-pkg",
						},
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "deprecated"},
			},
		},
	})
}

// A specifier that decodes into nothing would otherwise drop the whole allow
// list, leaving the rule reporting everything it was configured to permit.
func TestNoDeprecatedSchemaRejectsUnknownFrom(t *testing.T) {
	t.Parallel()
	allow := func(entry map[string]any) []any {
		return []any{map[string]any{"allow": []any{entry}}}
	}

	if err := NoDeprecatedRule.Schema.Validate(allow(map[string]any{
		"from": "package", "name": "oldValue", "package": "demo-pkg",
	})); err != nil {
		t.Fatalf("expected a well-formed specifier to validate, got %v", err)
	}
	if err := NoDeprecatedRule.Schema.Validate(allow(map[string]any{
		"from": "module", "name": "oldValue",
	})); err == nil {
		t.Fatal("expected an unknown `from` to be rejected")
	}
	if err := NoDeprecatedRule.Schema.Validate(allow(map[string]any{
		"from": "package", "name": "oldValue",
	})); err == nil {
		t.Fatal("expected `from: package` without a package to be rejected")
	}
}

func runNoDeprecatedDiagnosticsForFiles(t *testing.T, files map[string]string, entryFile string, options any) []rule.RuleDiagnostic {
	t.Helper()
	rootDir := fixtures.GetRootDir()
	virtualFiles := make(map[string]string, len(files))
	for fileName, content := range files {
		virtualFiles[tspath.ResolvePath(rootDir.Dir, fileName)] = content
	}
	fs := utils.NewOverlayVFS(cachedvfs.From(rootDir.FS), virtualFiles)
	host := utils.CreateCompilerHost(rootDir.Dir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir.Dir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(entryFile)
	if sourceFile == nil {
		t.Fatalf("failed to resolve entry file: %s", entryFile)
	}
	diagnostics := []rule.RuleDiagnostic{}
	var diagnosticsMu sync.Mutex
	programs := []*lintprogram.Program{lintprogram.NewFromCompiler(program)}
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         programs,
		TargetsByProgram: [][]string{{sourceFile.FileName()}},
		SingleThreaded:   true,
		GetRulesForFile: func(_ *ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     "test",
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoDeprecatedRule.Run(ctx, rule_tester.ResolveTestCaseOptions(t, &NoDeprecatedRule, options))
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	_, err = linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnosticsMu.Lock()
				diagnostics = append(diagnostics, diagnostic)
				diagnosticsMu.Unlock()
			},
		},
	})
	if err != nil {
		t.Fatalf("error running linter: %v", err)
	}
	return diagnostics
}

// The deprecated value's type is the literal `1`, which carries no name, so
// every one of these specifiers can only match the referenced value. Matching a
// value narrows past its name for `from: "package"` alone.
func TestAllowSpecifierMatchesImportedValue(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"deprecated.ts": `
/** @deprecated */
export const oldValue = 1;
		`,
		"main.ts": `
import { oldValue } from './deprecated';
oldValue;
		`,
	}

	allow := func(entries ...interface{}) []interface{} {
		return []interface{}{map[string]interface{}{"allow": entries}}
	}

	for _, testCase := range []struct {
		name     string
		options  []interface{}
		reported bool
	}{
		{name: "without-allow", options: nil, reported: true},
		{name: "shorthand-name", options: allow("oldValue")},
		{name: "from-file", options: allow(map[string]interface{}{"from": "file", "name": "oldValue"})},
		{
			name:    "from-file-with-unrelated-path",
			options: allow(map[string]interface{}{"from": "file", "path": "./nope.ts", "name": "oldValue"}),
		},
		{name: "from-lib", options: allow(map[string]interface{}{"from": "lib", "name": "oldValue"})},
		{
			name:     "from-package",
			options:  allow(map[string]interface{}{"from": "package", "package": "nope-pkg", "name": "oldValue"}),
			reported: true,
		},
		{
			name:     "unrelated-name",
			options:  allow(map[string]interface{}{"from": "file", "name": "other"}),
			reported: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			diagnostics := runNoDeprecatedDiagnosticsForFiles(t, files, "main.ts", testCase.options)
			expected := 0
			if testCase.reported {
				expected = 1
			}
			if len(diagnostics) != expected {
				t.Fatalf("expected %d diagnostics, got %d", expected, len(diagnostics))
			}
			if testCase.reported && diagnostics[0].Message.Id != "deprecated" {
				t.Fatalf("expected message id deprecated, got %s", diagnostics[0].Message.Id)
			}
		})
	}
}

func TestNoDeprecatedReportsLetInitializedToUndefinedInTsxFixture(t *testing.T) {
	t.Parallel()
	diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
		"react.tsx": `
        /** @deprecated */ let a = undefined;
        a;
      `,
	}, "react.tsx", nil)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}
	if diagnostics[0].Message.Id != "deprecated" {
		t.Fatalf("expected message id deprecated, got %s", diagnostics[0].Message.Id)
	}
}

func TestNoDeprecatedDeduplicatesExactRanges(t *testing.T) {
	t.Parallel()
	const code = `
interface Legacy {
  /** @deprecated Use current instead. */
  old(): void;
}
declare const legacy: Legacy;
legacy.old();
legacy.old();
`
	diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
		"main.ts": code,
	}, "main.ts", nil)
	if len(diagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics, got %#v", diagnostics)
	}
	for index, diagnostic := range diagnostics {
		if diagnostic.Message.Id != "deprecatedWithReason" ||
			diagnostic.Message.Description != "`old` is deprecated. Use current instead." {
			t.Fatalf("diagnostic %d message = %#v", index, diagnostic.Message)
		}
		if got := code[diagnostic.Range.Pos():diagnostic.Range.End()]; got != "old" {
			t.Fatalf("diagnostic %d range text = %q, want old", index, got)
		}
		if diagnostic.FixesPtr != nil || diagnostic.Suggestions != nil {
			t.Fatalf("diagnostic %d unexpectedly included edits: %#v", index, diagnostic)
		}
	}
}

func TestNoDeprecatedMatchesUpstreamMessageFormatting(t *testing.T) {
	t.Parallel()

	t.Run("jsdoc-links", func(t *testing.T) {
		t.Parallel()
		const code = `
/**
 * @deprecated This only works with the legacy {@link render} and {@link
 * renderSync} APIs. Use {@link Options} with {@link compile}, {@link
 * compileString}, {@link compileAsync}, and {@link compileStringAsync} instead.
 */
type LegacyOptions = string;
const legacyValue: LegacyOptions = '';
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.ts": code}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		want := "`LegacyOptions` is deprecated. This only works with the legacy {@link render } and {@link * renderSync} APIs. Use {@link Options } with {@link compile }, {@link * compileString}, {@link compileAsync }, and {@link compileStringAsync } instead."
		if got := diagnostics[0].Message.Description; got != want {
			t.Fatalf("message = %q, want %q", got, want)
		}
	})

	t.Run("import-alias-with-jsdoc-links", func(t *testing.T) {
		t.Parallel()
		files := map[string]string{
			"deprecated.ts": `
/**
 * @deprecated This only works with the legacy {@link render} and {@link
 * renderSync} APIs. Use {@link Options} instead.
 */
export type LegacyOptions = string;
`,
			"main.ts": `
import type { LegacyOptions as LegacySassOptions } from './deprecated';
declare const options: LegacySassOptions;
`,
		}
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, files, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		want := "`LegacySassOptions` is deprecated. This only works with the legacy {@link render } and {@link * renderSync} APIs. Use {@link Options } instead."
		if got := diagnostics[0].Message.Description; got != want {
			t.Fatalf("message = %q, want %q", got, want)
		}
	})

	t.Run("multiline-reason", func(t *testing.T) {
		t.Parallel()
		const code = `
interface Cache {
  /**
   * @deprecated This option has no effect. Rspack keeps only one persistent
   * cache per compiler path.
   */
  maxVersions?: number;
}
declare const cache: Cache;
cache.maxVersions;
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.ts": code}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		want := "`maxVersions` is deprecated. This option has no effect. Rspack keeps only one persistent\ncache per compiler path."
		if got := diagnostics[0].Message.Description; got != want {
			t.Fatalf("message = %q, want %q", got, want)
		}
	})

	t.Run("blank-line-in-reason", func(t *testing.T) {
		t.Parallel()
		const code = `
/**
 * @deprecated First line.
 *
 * Third line.
 */
type BlankLine = string;
const value: BlankLine = '';
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.ts": code}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		want := "`BlankLine` is deprecated. First line.\n\nThird line."
		if got := diagnostics[0].Message.Description; got != want {
			t.Fatalf("message = %q, want %q", got, want)
		}
	})

	t.Run("crlf-reason", func(t *testing.T) {
		t.Parallel()
		code := strings.ReplaceAll(`/**
 * @deprecated First line.
 * Second line.
 */
type CRLF = string;
const value: CRLF = '';
`, "\n", "\r\n")
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.ts": code}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		want := "`CRLF` is deprecated. First line.\r\nSecond line."
		if got := diagnostics[0].Message.Description; got != want {
			t.Fatalf("message = %q, want %q", got, want)
		}
	})

	t.Run("jsdoc-link-display-parts", func(t *testing.T) {
		t.Parallel()
		const code = `
class NewThing {}
/** @deprecated Use {@link NewThing}. */ class OldResolved {}
/** @deprecated Use {@link MissingThing}. */ class OldMissing {}
/** @deprecated Use {@link NewThing | replacement}. */ class OldLabel {}
/** @deprecated Use {@linkcode NewThing}. */ class OldCode {}
/** @deprecated Use {@linkplain NewThing label}. */ class OldPlain {}
/** @deprecated See {@link https://example.com | docs}. */ class OldURL {}
/**
 * @deprecated Use {@link
 * NewThing}.
 */
class OldMultiline {}
/** @deprecated A {@link NewThing<T>} B. */ class GenericResolved {}
/** @deprecated A {@link Missing<T>} B. */ class GenericMissing {}
/** @deprecated A {@link NewThing() | call label} B. */ class CallResolved {}
/** @deprecated A {@link Missing() | call label} B. */ class CallMissing {}
/**
 * @deprecated A {@link NewThing |
 * label} B.
 */
class WrappedLabel {}
new OldResolved();
new OldMissing();
new OldLabel();
new OldCode();
new OldPlain();
new OldURL();
new OldMultiline();
new GenericResolved();
new GenericMissing();
new CallResolved();
new CallMissing();
new WrappedLabel();
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.ts": code}, "main.ts", nil)
		if len(diagnostics) != 12 {
			t.Fatalf("expected 12 diagnostics, got %#v", diagnostics)
		}
		want := []string{
			"`OldResolved` is deprecated. Use {@link NewThing}.",
			"`OldMissing` is deprecated. Use {@link MissingThing }.",
			"`OldLabel` is deprecated. Use {@link NewThing" + "replacement}.",
			"`OldCode` is deprecated. Use {@linkcode NewThing}.",
			"`OldPlain` is deprecated. Use {@linkplain NewThing" + "label}.",
			"`OldURL` is deprecated. See {@link https://example.com docs}.",
			"`OldMultiline` is deprecated. Use {@link * NewThing}.",
			"`GenericResolved` is deprecated. A {@link NewThing<T>} B.",
			"`GenericMissing` is deprecated. A {@link Missing<T>} B.",
			"`CallResolved` is deprecated. A {@link NewThing() | call label} B.",
			"`CallMissing` is deprecated. A {@link Missing() | call label} B.",
			"`WrappedLabel` is deprecated. A {@link NewThing} * label} B.",
		}
		for index, diagnostic := range diagnostics {
			if got := diagnostic.Message.Description; got != want[index] {
				t.Fatalf("diagnostic %d message = %q, want %q", index, got, want[index])
			}
		}
	})
}

func TestNoDeprecatedRespectsScopedLineDisables(t *testing.T) {
	t.Parallel()
	const code = `
interface Legacy {
  /** @deprecated */
  old(): void;
}
declare const legacy: Legacy;
// eslint-disable-next-line test
legacy.old();
legacy.old(); // eslint-disable-line test
legacy.old();
`
	diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
		"main.ts": code,
	}, "main.ts", nil)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
	}
	wantStart := strings.LastIndex(code, "old")
	if diagnostics[0].Range.Pos() != wantStart || diagnostics[0].Range.End() != wantStart+len("old") {
		t.Fatalf("diagnostic range = %#v, want [%d, %d)", diagnostics[0].Range, wantStart, wantStart+len("old"))
	}
}

func TestNoDeprecatedIgnoresNameOnlyFallbackForElementAccessAndJsxAny(t *testing.T) {
	t.Parallel()

	t.Run("element-access-cross-object-collision", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
const a = { /** @deprecated */ b: 'string' };
const a2 = { b: 'string' };
a2['b'];
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})

	t.Run("element-access-any-typed-object", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
interface U1 { /** @deprecated */ shared: number }
const obj: any = {};
obj['shared'];
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})

	t.Run("jsx-attribute-any-props", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.tsx": `
declare namespace JSX {
  interface Element {}
  interface IntrinsicAttributes {}
}

interface U2 { /** @deprecated */ foo: number }
declare function Comp(p: any): any;
const x = <Comp foo={1} />;
			`,
		}, "main.tsx", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})
}

func TestNoDeprecatedIgnoresNameOnlyFallbackForResolvedTargets(t *testing.T) {
	t.Parallel()

	t.Run("dynamic-import-binding-cross-symbol-collision", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"implementation.ts": `
export function glob(patterns: string[]): Promise<string[]> {
  return Promise.resolve(patterns);
}
export function globSync(patterns: string[]): string[] {
  return patterns;
}
			`,
			"current.ts": `
export { glob, globSync } from './implementation';
			`,
			"main.ts": `
interface LegacyFinder {
  /** @deprecated Provide patterns as the first argument instead. */
  glob(pattern: string): Promise<string[]>;
  /** @deprecated Provide patterns as the first argument instead. */
  globSync(pattern: string): string[];
}

async function run() {
  const { glob, globSync } = await import('./current');
  await glob(['*.ts']);
  globSync(['*.ts']);
}

run();
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})

	t.Run("resolved-method-signature-cross-type-collision", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"current.ts": `
export interface CurrentEmitter {
  on(event: string, listener: () => void): CurrentEmitter;
}
			`,
			"main.ts": `
import type { CurrentEmitter } from './current';

interface LegacyEmitter {
  /** @deprecated Use addListener instead. */
  on(event: string, listener: () => void): LegacyEmitter;
}

declare const current: CurrentEmitter;
current.on('error', () => {});
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})

	t.Run("selected-non-deprecated-overload", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
interface Emitter {
  /** @deprecated Use destroy instead. */
  on(event: 'abort', listener: () => void): Emitter;
  on(event: 'error', listener: (error: Error) => void): Emitter;
}

declare const emitter: Emitter;
emitter.on('error', () => {});
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected 0 diagnostics, got %#v", diagnostics)
		}
	})
}

func TestNoDeprecatedMatchesUpstreamResolutionBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("computed-key-literal-types", func(t *testing.T) {
		t.Parallel()
		const code = `
declare const Keys: { numeric: 1 };
declare const keySource: { property: 'old' };
declare const annotatedKey: 'old';

const target = {
  /** @deprecated Use currentNumeric instead. */
  [1]: 'numeric',
  /** @deprecated Use current instead. */
  old: 'string',
};
const bigintTarget = {
  /** @deprecated Mirrors upstream's String(PseudoBigInt) key. */
  ['[object Object]']: 'bigint',
};
declare const bigintKey: 1n;

target[Keys.numeric];
target[keySource.property];
target[annotatedKey];
function read(key: 'old') {
  return target[key];
}
target['anything' as 'old'];
bigintTarget[bigintKey];
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": code,
		}, "main.ts", nil)
		if len(diagnostics) != 6 {
			t.Fatalf("expected 6 diagnostics, got %#v", diagnostics)
		}
		gotRanges := map[string]int{}
		for _, diagnostic := range diagnostics {
			gotRanges[code[diagnostic.Range.Pos():diagnostic.Range.End()]]++
		}
		for _, want := range []string{"Keys.numeric", "keySource.property", "annotatedKey", "key", "'anything' as 'old'", "bigintKey"} {
			if gotRanges[want] != 1 {
				t.Fatalf("diagnostic ranges = %#v, want one %q", gotRanges, want)
			}
		}
	})

	t.Run("computed-key-widened-types", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
const target = {
  /** @deprecated Use current instead. */
  old: 'string',
};
let mutableKey = 'old';
const assertedKey = 'old' as string;
target[mutableKey];
target[assertedKey];
`,
		}, "main.ts", nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected no diagnostics, got %#v", diagnostics)
		}
	})

	t.Run("mixed-overload-reference-and-calls", func(t *testing.T) {
		t.Parallel()
		const code = `
interface Emitter {
  /** @deprecated Use destroy instead. */
  on(event: 'abort', listener: () => void): Emitter;
  on(event: 'error', listener: (error: Error) => void): Emitter;
}

declare const emitter: Emitter;
const methodReference = emitter.on;
emitter.on('error', () => {});
emitter.on('abort', () => {});

interface Props {
  /** @deprecated Use current instead. */
  callback: () => void;
}
declare const props: Props;
props.callback();
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": code,
		}, "main.ts", nil)
		if len(diagnostics) != 3 {
			t.Fatalf("expected 3 diagnostics, got %#v", diagnostics)
		}
		gotRanges := map[string]int{}
		for _, diagnostic := range diagnostics {
			gotRanges[code[diagnostic.Range.Pos():diagnostic.Range.End()]]++
		}
		if gotRanges["on"] != 2 || gotRanges["callback"] != 1 {
			t.Fatalf("diagnostic ranges = %#v, want two on and one callback", gotRanges)
		}
	})

	t.Run("mixed-declarations-remain-deprecated-when-destructured", func(t *testing.T) {
		t.Parallel()
		const code = `
interface Emitter {
  /** @deprecated Use destroy instead. */
  on(event: 'abort', listener: () => void): Emitter;
  on(event: 'error', listener: (error: Error) => void): Emitter;
}
declare const emitter: Emitter;
const { on } = emitter;
on('error', () => {});
on('abort', () => {});

interface Merged {
  /** @deprecated Use current instead. */
  item: string;
}
interface Merged {
  item: string;
}
declare const merged: Merged;
const { item } = merged;
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": code,
		}, "main.ts", nil)
		if len(diagnostics) != 3 {
			t.Fatalf("expected 3 diagnostics, got %#v", diagnostics)
		}
		gotRanges := map[string]int{}
		for _, diagnostic := range diagnostics {
			gotRanges[code[diagnostic.Range.Pos():diagnostic.Range.End()]]++
		}
		if gotRanges["on"] != 2 || gotRanges["item"] != 1 {
			t.Fatalf("diagnostic ranges = %#v, want two on and one item", gotRanges)
		}
	})

	t.Run("directly-exported-mixed-overload-binding", func(t *testing.T) {
		t.Parallel()
		const code = `
async function run() {
  const { glob } = await import('./dependency');
  await glob(['*.ts']);
}
void run();
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"dependency.ts": `
export function glob(patterns: string[]): Promise<string[]>;
/** @deprecated Provide patterns as the first argument instead. */
export function glob(options: { patterns: string[] }): Promise<string[]>;
export function glob(input: string[] | { patterns: string[] }): Promise<string[]> {
  return Promise.resolve(Array.isArray(input) ? input : input.patterns);
}
`,
			"main.ts": code,
		}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected one diagnostic, got %#v", diagnostics)
		}
		if got := code[diagnostics[0].Range.Pos():diagnostics[0].Range.End()]; got != "glob" {
			t.Fatalf("diagnostic range = %q, want glob", got)
		}
	})

	t.Run("deprecated-reexport-binding", func(t *testing.T) {
		t.Parallel()
		const code = `
import * as imported from './dependency';
import { normalFunction } from './dependency';
const { normalVariable } = imported;
normalFunction();
`
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"dependency.ts": `
const normalVariable = 1;
function normalFunction(): void {}
export {
  /** @deprecated */
  normalVariable,
  /** @deprecated */
  normalFunction,
};
`,
			"main.ts": code,
		}, "main.ts", nil)
		if len(diagnostics) != 2 {
			t.Fatalf("expected two diagnostics, got %#v", diagnostics)
		}
		gotRanges := map[string]int{}
		for _, diagnostic := range diagnostics {
			gotRanges[code[diagnostic.Range.Pos():diagnostic.Range.End()]]++
		}
		for _, want := range []string{"normalVariable", "normalFunction"} {
			if gotRanges[want] != 1 {
				t.Fatalf("diagnostic ranges = %#v, want one %q", gotRanges, want)
			}
		}
	})
}

func TestNoDeprecatedMultilineImportsAndMultiDeclarators(t *testing.T) {
	t.Parallel()

	t.Run("shorthand-property-assignment-still-reports-deprecated-variable", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
/** @deprecated */
declare const test: string;
const bar = { test };
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		if diagnostics[0].Message.Id != "deprecated" {
			t.Fatalf("expected deprecated message id, got %s", diagnostics[0].Message.Id)
		}
		if !strings.Contains(diagnostics[0].Message.Description, "`test`") {
			t.Fatalf("expected test diagnostic, got %#v", diagnostics[0].Message)
		}
	})

	t.Run("multiline-import-only-reports-usage", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"dep.ts": `
/** @deprecated */
export const oldValue = 1;
			`,
			"main.ts": `
import {
  oldValue,
} from './dep';

oldValue;
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		if diagnostics[0].Message.Id != "deprecated" {
			t.Fatalf("expected deprecated message id, got %s", diagnostics[0].Message.Id)
		}
		if !strings.Contains(diagnostics[0].Message.Description, "oldValue") {
			t.Fatalf("expected oldValue diagnostic, got %#v", diagnostics[0].Message)
		}
	})

	t.Run("statement-jsdoc-only-applies-to-first-declarator", func(t *testing.T) {
		t.Parallel()
		diagnostics := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{
			"main.ts": `
/** @deprecated */ const a = 1, b = 2;
a;
b;
			`,
		}, "main.ts", nil)
		if len(diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
		}
		if !strings.Contains(diagnostics[0].Message.Description, "`a`") {
			t.Fatalf("expected only a to be reported, got %#v", diagnostics[0].Message)
		}
	})
}

// Independently checked against typescript-eslint v8.72.2 and TypeScript 6.0.2.
func TestNoDeprecatedAdditionalParity(t *testing.T) {
	type expected struct {
		id, message string
		start, end  int
	}
	for _, tc := range []struct {
		name, code string
		options    any
		node16     bool
		want       []expected
	}{
		{name: "nearby-reason", code: `/** @deprecated unrelated reason */ const first = 1;
/** @deprecated */ const second = 2;
second;`, want: []expected{{"deprecated", "`second` is deprecated.", 90, 96}}},
		{name: "nested-reason", code: `/** @deprecated outer */ class Old {
/** @deprecated inner */ method() {}
}
new Old();`, want: []expected{{"deprecatedWithReason", "`Old` is deprecated. outer", 80, 83}}},
		{name: "tag-boundary", code: `/** @deprecatedish no */ const x=1;
x;`, want: []expected{}},
		{name: "tag-in-prose", code: `/** This mentions @deprecated in prose. */ const x=1;
x;`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. in prose.", 54, 55}}},
		{name: "tag-followed-by-return", code: `/** @deprecated old
 * @returns current
 */ function old() {}
old();`, want: []expected{{"deprecatedWithReason", "`old` is deprecated. old", 62, 65}}},
		{name: "multiple-declarations", code: `/** @deprecated reason */ const first=1, second=2;
first; second;`, want: []expected{{"deprecatedWithReason", "`first` is deprecated. reason", 51, 56}}},
		{name: "array-binding", code: `const [/** @deprecated local */ x] = [1];
x;`, want: []expected{}},
		{name: "shadowing", code: `/** @deprecated old */ const x=1;
function f(x:number){return {x};}
x;`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. old", 68, 69}}},
		{name: "shorthand", code: `/** @deprecated reason */ const x=1;
const o={x};`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. reason", 46, 47}}},
		{name: "selected-overload", code: `/** @deprecated old */ function f(x:0):void;
function f(x:1):void;
function f(x:number){}
f(1);f(0);`, want: []expected{{"deprecatedWithReason", "`f` is deprecated. old", 95, 96}}},
		{name: "paren-callee", code: `/** @deprecated old */ function f(x:0):void;
function f(x:1):void;
function f(x:number){}
(f)(1);(f)(0);`, want: []expected{{"deprecatedWithReason", "`f` is deprecated. old", 98, 99}}},
		{name: "empty-reason-overload", code: `/** @deprecated unrelated */ function f(x:0):void;
/** @deprecated */ function f(x:1):void;
function f(x:number){}
f(1);`, want: []expected{{"deprecated", "`f` is deprecated.", 115, 116}}},
		{name: "computed-overload", code: `interface A {
/** @deprecated old */ m(x:0):void;
m(x:1):void;
}
declare const a:A; a['m'](1);`, want: []expected{{"deprecatedWithReason", "`m` is deprecated. old", 86, 89}}},
		{name: "computed-key-reference", code: `/** @deprecated key */ const k='x';
const a={x:1};a[k];`, want: []expected{}},
		{name: "computed-empty-key", code: `const a={
/** @deprecated empty */ '':1
};a[''];`, want: []expected{{"deprecatedWithReason", "`` is deprecated. empty", 44, 46}}},
		{name: "private-field", code: `class C {
/** @deprecated secret */ #x=1;
f(){return this.#x;}
}`, want: []expected{{"deprecatedWithReason", "`#x` is deprecated. secret", 58, 60}}},
		{name: "inheritance", code: `class A {
/** @deprecated parent */ m(){}
}
class B extends A {m(){}}
new B().m();`, want: []expected{{"deprecatedWithReason", "`m` is deprecated. parent", 78, 79}}},
		{name: "contextual-method", code: `interface A {
/** @deprecated old */ m():void;
}
const a:A={m(){}};`, want: []expected{{"deprecatedWithReason", "`m` is deprecated. old", 60, 61}}},
		{name: "contextual-getter", code: `interface A {
/** @deprecated old */ x:number;
}
const a:A={get x(){return 1;}};`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. old", 64, 65}}},
		{name: "contextual-shorthand-double", code: `interface A {
/** @deprecated property */ x:number;
}
/** @deprecated variable */ const x=1;
const a:A={x};`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. property", 104, 105}, {"deprecatedWithReason", "`x` is deprecated. variable", 104, 105}}},
		{name: "contextual-computed-expression", code: `interface A {
/** @deprecated property */ x:number;
}
const key='x';const a:A={[key]:1};`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. property", 80, 83}}},
		{name: "binding-rename-same", code: `const a={
/** @deprecated property */ x:1
};const {x:x}=a;`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. property", 51, 52}}},
		{name: "binding-rest", code: `const a={
/** @deprecated property */ x:1
};const {...x}=a;`, want: []expected{}},
		{name: "binding-default", code: `const a={
/** @deprecated property */ x:1
};const {x=1}=a;`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. property", 51, 52}}},
		{name: "qualified-type", code: `namespace N {
/** @deprecated old */ export type T=string;
}
type X=N.T;`, want: []expected{{"deprecatedWithReason", "`T` is deprecated. old", 70, 71}}},
		{name: "jsx-closing-member", code: `namespace N {
/** @deprecated old */ export function C(){return null;}
}
const x=<N.C></N.C>;`, want: []expected{{"deprecatedWithReason", "`C` is deprecated. old", 84, 85}, {"deprecatedWithReason", "`C` is deprecated. old", 90, 91}}},
		{name: "jsdoc-multiple", code: `/** @deprecated first */
/** @deprecated second */
const x=1;x;`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. second", 61, 62}}},
		{name: "long-reason", code: `/** @deprecated First line
 *
 * Last line
 */const x=1;x;`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. First line\n\nLast line", 56, 57}}},
		{name: "inheritance-local-tag-blocks", code: `class A {
/** @deprecated parent */ m(){}
}
class B extends A {
/** @returns nothing */ m(){}
}
new B().m();`, want: []expected{}},
		{name: "inheritance-local-reason", code: `class A {
/** @deprecated parent */ m(){}
}
class B extends A {
/** @deprecated child */ m(){}
}
new B().m();`, want: []expected{{"deprecatedWithReason", "`m` is deprecated. child", 105, 106}}},
		{name: "inheritance-explicit-tag", code: `class A {
/** @deprecated parent */ m(){}
}
class B extends A {
/** @inheritDoc
 * @deprecated child */ m(){}
}
new B().m();`, want: []expected{{"deprecatedWithReason", "`m` is deprecated. parent", 120, 121}}},
		{name: "inheritance-interface", code: `interface A {
/** @deprecated interface */ m():void;
}
class B implements A {m(){}}
new B().m();`, want: []expected{{"deprecatedWithReason", "`m` is deprecated. interface", 92, 93}}},
		{name: "inheritance-static", code: `class A {
/** @deprecated static */ static m(){}
}
class B extends A {static m(){}}
B.m();`, want: []expected{{"deprecatedWithReason", "`m` is deprecated. static", 86, 87}}},
		{name: "inheritance-getter", code: `class A {
/** @deprecated getter */ get x(){return 1;}
set x(v:number){}
}
class B extends A {get x(){return 1;}}
new B().x;`, want: []expected{{"deprecatedWithReason", "`x` is deprecated. getter", 122, 123}}},
		{name: "inheritance-overloaded-base", code: `class A {
/** @deprecated overload */ m(x:0):void;
m(x:1):void;
m(x:number){}
}
class B extends A {m(x:number){}}
new B().m(0);`, want: []expected{}},
		{name: "inheritance-recursive", code: `class A implements A {m(){}}
new A().m();`, want: []expected{}},
		{name: "paren-property-call", code: `class A {
/** @deprecated zero */ m(x:0):void;
m(x:1):void;
m(x:number){}
}
const a=new A();((a.m))(1);((a.m))(0);`, want: []expected{{"deprecatedWithReason", "`m` is deprecated. zero", 107, 108}}},
		{name: "initializer-jsdoc", code: `const old = /** @deprecated initializer */ ()=>{};
old();`, want: []expected{{"deprecatedWithReason", "`old` is deprecated. initializer", 51, 54}}},
		{name: "unicode-range", code: `const text="😀"; /** @deprecated */ const 旧=1;
旧;`, want: []expected{}},
		{name: "unicode-reference-range", code: `/** @deprecated use new */ const 旧=1;
const text="😀"; 旧;`, want: []expected{{"deprecatedWithReason", "`旧` is deprecated. use new", 59, 62}}},
		{name: "previous-valid-44", code: `
        async function fn() {
          const d = await import('./deprecated.js');
          d.default['foo'];
        }
      `, node16: true, want: []expected{}},
		{name: "previous-valid-62", code: `
        interface AllowedType {
          /** @deprecated */
          prop: string;
        }

        const obj: AllowedType = {
          prop: 'test',
        };

        const value = obj['prop'];
      `, options: []any{map[string]any{"allow": []any{map[string]any{"from": "file", "name": "AllowedType"}}}}, want: []expected{}},
		{name: "previous-valid-72", code: `
      const a = { /** @deprecated */ b: 'string' };
      const a2 = { b: 'string' };

      a2['b'];
    `, want: []expected{}},
		{name: "previous-valid-73", code: `
      interface U1 {
        /** @deprecated */
        shared: number;
      }

      const obj: any = {};
      obj['shared'];
    `, want: []expected{}},
		{name: "previous-valid-74", code: `
        interface U2 {
          /** @deprecated */
          foo: number;
        }

        declare function Comp(p: any): any;

        const x = <Comp foo={1} />;
      `, want: []expected{}},
		{name: "previous-valid-75", code: `
declare function log(x: unknown): void;
try {
  log(1);
} catch (error) {
  log(error);
}
    `, want: []expected{}},
		{name: "previous-invalid-10", code: `
        /** @deprecated */ const a = { b: 1 };
        console.log(a);
      `, want: []expected{{"deprecated", "`a` is deprecated.", 68, 69}}},
		{name: "previous-invalid-13", code: `
        /** @deprecated */ const a = { b: 1 };
        console.log(a.b);
      `, want: []expected{{"deprecated", "`a` is deprecated.", 68, 69}}},
		{name: "previous-invalid-14", code: `
        /** @deprecated */ const a = { b: 1 };
        console.log(a?.b);
      `, want: []expected{{"deprecated", "`a` is deprecated.", 68, 69}}},
		{name: "previous-invalid-22", code: `
        /** @deprecated */
        declare const test: string;
        const myObj = {
          prop: test,
          deep: {
            prop: test,
          },
        };
      `, want: []expected{{"deprecated", "`test` is deprecated.", 104, 108}, {"deprecated", "`test` is deprecated.", 146, 150}}},
		{name: "previous-invalid-79", code: `
        import {
          deprecatedVariable,
        } from './deprecated';

        const foo = deprecatedVariable;
      `, want: []expected{{"deprecated", "`deprecatedVariable` is deprecated.", 100, 118}}},
		{name: "previous-invalid-139", code: `
        /** @deprecated */
        const a = 1, b = 2;

        a;
        b;
      `, want: []expected{{"deprecated", "`a` is deprecated.", 65, 66}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"main.tsx": tc.code,
				"deprecated.ts": `/** @deprecated */
export const deprecatedVariable = 1;
/** @deprecated */
export default {foo: 1};`,
			}
			if tc.node16 {
				files["tsconfig.json"] = `{"compilerOptions":{"jsx":"preserve","module":"node16","moduleResolution":"node16","strict":true,"esModuleInterop":true,"types":[]}}`
			}
			got := runNoDeprecatedDiagnosticsForFiles(t, files, "main.tsx", tc.options)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d diagnostics, want %d: %#v", len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				d := got[i]
				if d.Message.Id != w.id || d.Message.Description != w.message || d.Range.Pos() != w.start || d.Range.End() != w.end {
					t.Errorf("diagnostic %d = %s %q [%d,%d), want %s %q [%d,%d)", i, d.Message.Id, d.Message.Description, d.Range.Pos(), d.Range.End(), w.id, w.message, w.start, w.end)
				}
				if d.FixesPtr != nil || d.Suggestions != nil {
					t.Error("no-deprecated must not provide fixes or suggestions")
				}
			}
		})
	}
}

func TestNoDeprecatedLongAliasChain(t *testing.T) {
	const code = "import { value } from './alias0';\nvalue;"
	files := map[string]string{"main.ts": code}
	for i := range 72 {
		name := "alias" + strconv.Itoa(i) + ".ts"
		if i == 71 {
			files[name] = "export const value=1;"
			continue
		}
		tag := ""
		if i == 70 {
			tag = "/** @deprecated deep alias */ "
		}
		files[name] = "export { " + tag + "value } from './alias" + strconv.Itoa(i+1) + "';"
	}
	got := runNoDeprecatedDiagnosticsForFiles(t, files, "main.ts", nil)
	if len(got) != 1 {
		t.Fatalf("got %d diagnostics, want 1", len(got))
	}
	d := got[0]
	if d.Message.Id != "deprecatedWithReason" || d.Message.Description != "`value` is deprecated. deep alias" || d.Range.Pos() != len(code)-6 || d.Range.End() != len(code)-1 {
		t.Fatalf("unexpected diagnostic: %#v", d)
	}
}

func TestNoDeprecatedClassExpressionInheritance(t *testing.T) {
	for _, tc := range []struct{ code, message string }{
		{`const C=class { method(){} }; new C().method();`, ""},
		{`class Base {
/** @deprecated inherited */ method(){}
}
const C=class extends Base {method(){}};
new C().method();`, "`method` is deprecated. inherited"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			got := runNoDeprecatedDiagnosticsForFiles(t, map[string]string{"main.ts": tc.code}, "main.ts", nil)
			if tc.message == "" {
				if len(got) != 0 {
					t.Fatalf("unexpected diagnostics: %#v", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("got %d diagnostics, want 1", len(got))
			}
			d := got[0]
			start := strings.LastIndex(tc.code, "method")
			if d.Message.Id != "deprecatedWithReason" || d.Message.Description != tc.message || d.Range.Pos() != start || d.Range.End() != start+len("method") {
				t.Fatalf("unexpected diagnostic: %#v", d)
			}
		})
	}
}
