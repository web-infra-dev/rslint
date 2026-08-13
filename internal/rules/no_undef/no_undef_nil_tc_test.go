package no_undef

import (
	"fmt"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestNoUndefRuleTypeCheckerInvariant is the regression barrier against the
// old hybrid semantics: acquiring a TypeChecker must not expose host/runtime,
// ambient, cross-file, or non-default type names to an ESLint scope rule. The
// parser's default TypeScript type globals remain available in either mode.
func TestNoUndefRuleTypeCheckerInvariant(t *testing.T) {
	t.Parallel()
	root := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(root.Dir, "no-undef-checker-invariant.ts")
	crossFilePath := tspath.ResolvePath(root.Dir, "no-undef-cross-file-global.ts")
	ambientPath := tspath.ResolvePath(root.Dir, "no-undef-external-global.d.ts")
	code := `var declaredLocal = 1; declaredLocal;
console.log(1);
window;
crossFileGlobal123;
externalAmbientGlobal123;
type DefaultTypeGlobal123 = AsyncIterator<string>;
type CrossFileTypeUse123 = crossFileType123;
type ExternalAmbientTypeUse123 = externalAmbientType123;
Promise.resolve();
var buf = new Float16Array(8);
Temporal;
AsyncIterator;
myConfiguredGlobal;
myOffGlobal;
var myOffLocal = 1; myOffLocal;
undeclaredName123;
`
	fs := utils.NewOverlayVFS(root.FS, map[string]string{
		filePath:      code,
		crossFilePath: "var crossFileGlobal123 = 1; interface crossFileType123 {}",
		ambientPath:   "declare var externalAmbientGlobal123: string; interface externalAmbientType123 {}",
	})
	program, err := utils.CreateProgram(
		true, fs, root.Dir, "tsconfig.json", utils.CreateCompilerHost(root.Dir, fs),
	)
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatalf("source file not found for %s", filePath)
		return
	}
	if program.GetSourceFile(crossFilePath) == nil || program.GetSourceFile(ambientPath) == nil {
		t.Fatal("checker-invariance fixture did not include cross-file declarations")
	}

	checker, done := program.GetTypeChecker(t.Context())
	t.Cleanup(done)

	want := []string{
		"'console' is not defined.",
		"'window' is not defined.",
		"'crossFileGlobal123' is not defined.",
		"'externalAmbientGlobal123' is not defined.",
		"'crossFileType123' is not defined.",
		"'externalAmbientType123' is not defined.",
		"'AsyncIterator' is not defined.",
		"'myOffGlobal' is not defined.",
		"'undeclaredName123' is not defined.",
	}

	for name, typeChecker := range map[string]bool{
		"without checker": false,
		"with checker":    true,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var tc = checker
			if !typeChecker {
				tc = nil
			}
			var reported []string
			ctx := (rule.RuleContext{
				SourceFile:  sourceFile,
				Program:     program,
				TypeChecker: tc,
				Refs:        rule.NewRefStore(sourceFile, program.Options(), tc, rule.RefStoreInit{}),
				Globals: rule.NewGlobals(rule.LanguageOptions{}, rule.GlobalsInit{}, map[string]utils.GlobalAccess{
					"myConfiguredGlobal": utils.GlobalAccessReadonly,
					"myOffGlobal":        utils.GlobalAccessOff,
					"myOffLocal":         utils.GlobalAccessOff,
				}, nil, nil),
			}).WithReporter("test/no-undef", rule.SeverityError, func(d rule.RuleDiagnostic) {
				reported = append(reported, d.Message.Description)
			})

			listeners := NoUndefRule.Run(ctx, nil)
			var walk func(n *ast.Node) bool
			walk = func(n *ast.Node) bool {
				if cb, ok := listeners[n.Kind]; ok {
					cb(n)
				}
				n.ForEachChild(walk)
				return false
			}
			walk(sourceFile.AsNode())

			if fmt.Sprint(reported) != fmt.Sprint(want) {
				t.Fatalf("reported = %v, want %v", reported, want)
			}
		})
	}
}
