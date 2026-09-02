// TestSortPropTypesExtras locks in branches and edge shapes that the upstream
// test suite does not exercise. Upstream-migrated cases live in the sibling
// sort_prop_types_upstream_test.go file.
package sort_prop_types

import (
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestSortPropTypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &SortPropTypesRule, []rule_tester.ValidTestCase{
		// Upstream keeps TypeScript assertions opaque in this path.
		{Code: `const x = { propTypes: ({ z: PropTypes.string, a: PropTypes.string } as any) };`, Tsx: true},
		// Identifier resolution does not apply to object-literal propTypes properties.
		{Code: `const props = { z: PropTypes.string, a: PropTypes.string }; const x = { propTypes: props };`, Tsx: true},
		// The shape listener likewise leaves TypeScript assertions opaque.
		{Code: `any.shape(({ z: PropTypes.string, a: PropTypes.string } as any));`, Options: map[string]any{"sortShapeProp": true}, Tsx: true},
		// Upstream only recognizes a property access named `shape`, not a bare call.
		{Code: `shape({ z: PropTypes.string, a: PropTypes.string });`, Options: map[string]any{"sortShapeProp": true}, Tsx: true},
		// A defaulted parameter is an ESTree AssignmentPattern without a direct annotation.
		{Code: `type Props = { z: string; a: string }; function Component(p: Props = {} as Props) {}`, Options: map[string]any{"checkTypes": true}, Tsx: true},
		// Upstream checks only a direct, already-seen type-literal alias.
		{Code: `type Props = { z: string; a: string }; type Alias = Props; function Component(p: Alias) {}`, Options: map[string]any{"checkTypes": true}, Tsx: true},
		{Code: `type Props = { z: string; a: string }; function Component(p: Namespace.Props) {}`, Options: map[string]any{"checkTypes": true}, Tsx: true},
		{Code: `function Component(p: Props) {} type Props = { z: string; a: string };`, Options: map[string]any{"checkTypes": true}, Tsx: true},
		// Locks in upstream checkNode()'s nil-value early return for a declaration without an initializer.
		{Code: `class Component { propTypes: unknown; }`, Tsx: true},
		// ---- Dimension 4: spread resets the ordering group ----
		{Code: `Component.propTypes = { z: PropTypes.string, ...shared, a: PropTypes.string };`, Tsx: true},
		// ---- Dimension 4: TS expression wrappers on the propTypes object ----
		{Code: `Component.propTypes = ({ a: PropTypes.string, z: PropTypes.string } as any);`, Tsx: true},
		// N/A: private keys and element access do not participate in upstream's static key sort.
		// An unnamed TypeScript member breaks the comparison chain.
		{Code: `type Props = { z: string; (): void; a: string }; function Component(props: Props) {}`, Options: map[string]any{"checkTypes": true}, Tsx: true},
		// A wrapper configured as Object.assign matches the bare callee name,
		// but not the member expression itself, as in the upstream rule.
		{Code: `Component.propTypes = Object.assign({ z: PropTypes.string, a: PropTypes.string });`, Settings: map[string]interface{}{"propWrapperFunctions": []any{map[string]interface{}{"object": "Object", "property": "assign"}}}, Tsx: true},
		// Object-literal propTypes detection uses the authored identifier key,
		// not a static string key.
		{Code: `const C = { "propTypes": { z: PropTypes.string, a: PropTypes.string } };`, Tsx: true},
		// Required detection does not cross a TypeScript assertion wrapper.
		{Code: `Component.propTypes = { a: PropTypes.string, b: (PropTypes.string.isRequired as any) };`, Options: map[string]any{"requiredFirst": true}, Tsx: true},
		// Identifier resolution likewise keeps TypeScript assertions opaque.
		{Code: `const p = ({ z: PropTypes.string, a: PropTypes.string } as const); Component.propTypes = p;`, Tsx: true},
		// Shape sorting only examines a direct object or identifier argument.
		{Code: `PropTypes.shape(forbidExtraProps({ z: PropTypes.string, a: PropTypes.string }));`, Options: map[string]any{"sortShapeProp": true}, Settings: map[string]interface{}{"propWrapperFunctions": []any{"forbidExtraProps"}}, Tsx: true},
		// JavaScript relational string ordering compares UTF-16 code units.
		{Code: `Component.propTypes = { "\u{10000}": PropTypes.string, "\uE000": PropTypes.string };`, Tsx: true},
		// Optional wrapper calls remain opaque to the upstream direct-callee lookup.
		{Code: `Component.propTypes = wrap?.({ z: PropTypes.string, a: PropTypes.string });`, Settings: map[string]interface{}{"propWrapperFunctions": []any{"wrap"}}, Tsx: true},
		// Auto-accessor fields are not included in the upstream class-property listener.
		{Code: `class C { static accessor propTypes = { z: PropTypes.string, a: PropTypes.string }; }`, Tsx: true},
		// A comma expression is not an assignment target in the upstream listener.
		{Code: `Component.propTypes, ({ z: PropTypes.string, a: PropTypes.string });`, Tsx: true},
		// An empty string key falls back to its authored source text upstream.
		{Code: `C.propTypes = { 1: PropTypes.string, "": PropTypes.string };`, Tsx: true},
		// Optional required access is opaque, while the computed identifier is visible.
		{Code: `Component.propTypes = { a: PropTypes.string, b: PropTypes?.isRequired };`, Options: map[string]any{"requiredFirst": true}, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// Locks in upstream callbackPropsLastSeen: report a misplaced callback only once.
		{Code: `Component.propTypes = { onChange: PropTypes.func, a: PropTypes.string, b: PropTypes.string };`, Options: map[string]any{"callbacksLast": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "callbackPropsLast", Line: 1, Column: 25}}},
		// Upstream falls back to the authored text of computed keys when sorting.
		{Code: `Component.propTypes = { [z]: PropTypes.string, a: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 48}}},
		// Locks in upstream checkSorted() required-first arm: a required prop after an optional prop.
		{Code: `Component.propTypes = { a: PropTypes.string, b: PropTypes.string.isRequired };`, Options: map[string]any{"requiredFirst": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredPropsFirst", Line: 1, Column: 46}}},
		// Locks in upstream CallExpression listener: shape properties are checked independently.
		{Code: `Component.propTypes = { item: PropTypes.shape({ z: PropTypes.string, a: PropTypes.string }) };`, Options: map[string]any{"sortShapeProp": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 70}}},
		// Upstream compares numeric keys as numbers instead of lexicographic strings.
		{Code: `C.propTypes = { 2: T, 10: T, 1: T };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 30}}},
		// ---- Real-user: #2702 TypeScript alias used by a function component ----
		{Code: `type Props = { z: string; a: string }; function Component(props: Props) { return null; }`, Options: map[string]any{"checkTypes": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 27}}},
		// ---- Real-user: #3578 inline TypeScript component props ----
		{Code: `const Component = (props: { z: string; a: string }) => null;`, Options: map[string]any{"checkTypes": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted", Line: 1, Column: 40}}},
		// Typed class fields named `props` are checked by the upstream rule.
		{Code: `class Component { static props: unknown = { z: PropTypes.string, a: PropTypes.string }; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		// Computed identifier keys are recognized, while quoted keys are not.
		{Code: `const C = { [propTypes]: { z: PropTypes.string, a: PropTypes.string } };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		// An object/property wrapper entry also matches the bare property name.
		{Code: `Component.propTypes = assign({ z: PropTypes.string, a: PropTypes.string });`, Settings: map[string]interface{}{"propWrapperFunctions": []any{map[string]interface{}{"object": "Object", "property": "assign"}}}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		// A dotted string setting is compared as authored by upstream and does not
		// become a bare `assign` wrapper after shared normalization.
		{Code: `Component.propTypes = assign({ z: PropTypes.string, a: PropTypes.string });`, Settings: map[string]interface{}{"propWrapperFunctions": []any{"Object.assign"}}, Tsx: true},
		// Computed class names follow the authored identifier rule: quoted strings
		// are ignored, while computed identifiers are checked.
		{Code: `class C { static ["propTypes"] = { z: PropTypes.string, a: PropTypes.string }; }`, Tsx: true},
		{Code: `class C { static [propTypes] = { z: PropTypes.string, a: PropTypes.string }; }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		// The assignment listener also recognizes a computed identifier member.
		{Code: `C[propTypes] = { z: PropTypes.string, a: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		// Shape detection does not cross TypeScript expression wrappers on the callee.
		{Code: `(PropTypes.shape as any)({ z: PropTypes.string, a: PropTypes.string });`, Options: map[string]any{"sortShapeProp": true}, Tsx: true},
		// Assignment detection does not cross TypeScript expression wrappers on the member.
		{Code: `(Component.propTypes as any) = { z: PropTypes.string, a: PropTypes.string };`, Tsx: true},
		{Code: `(Component.propTypes!) = { z: PropTypes.string, a: PropTypes.string };`, Tsx: true},
		// Private names are exposed by ESTree without the leading '#'.
		{Code: `class C { static #propTypes = { z: PropTypes.string, a: PropTypes.string }; method() { this.#propTypes = { z: PropTypes.string, a: PropTypes.string }; } }`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}, {MessageId: "propsNotSorted"}}},
		// Computed shape names use the ESTree member's identifier name.
		{Code: `Component.propTypes = { x: PropTypes[shape]({ z: PropTypes.string, a: PropTypes.string }) };`, Options: map[string]any{"sortShapeProp": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		// JavaScript relational comparison preserves non-string key values.
		{Code: `C.propTypes = { 2: PropTypes.string, [true]: PropTypes.string };`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		{Code: "C.propTypes = { 10n: PropTypes.string, 2n: PropTypes.string };", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		{Code: "C.propTypes = { a: PropTypes.string, [`z`]: PropTypes.string };", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		// Computed required access is visible to the upstream property-name lookup.
		{Code: `Component.propTypes = { a: PropTypes.string, b: PropTypes[isRequired] };`, Options: map[string]any{"requiredFirst": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredPropsFirst"}}},
		// Parenthesized type annotations are transparent in typescript-eslint.
		{Code: `function Component(props: ({ z: string; a: string })) {}`, Options: map[string]any{"checkTypes": true}, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
		// Destructured bindings resolve through their containing variable declaration.
		{Code: `const { props } = { z: PropTypes.string, a: PropTypes.string }; Component.propTypes = props;`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "propsNotSorted"}}},
	})
}

func TestSortPropTypesDoesNotResolveObjectsAcrossFiles(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	declarationFile := tspath.NormalizePath(filepath.Join(dir, "global.ts"))
	usageFile := tspath.NormalizePath(filepath.Join(dir, "cross-file.ts"))
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		declarationFile: `const sharedProps = { z: PropTypes.string, a: PropTypes.string };`,
		usageFile:       `C.propTypes = sharedProps;`,
	})
	host := utils.CreateCompilerHost(dir, fs)
	program, err := utils.CreateProgramFromOptions(true, &core.CompilerOptions{
		Target: core.ScriptTargetESNext,
	}, []string{declarationFile, usageFile}, host)
	if err != nil {
		t.Fatalf("create typed program: %v", err)
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:     lintprogram.NewFromCompiler(program),
		File:        usageFile,
		HasTypeInfo: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     SortPropTypesRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return SortPropTypesRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	})

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostic count = %d, want 0: %+v", len(diagnostics), diagnostics)
	}
}
