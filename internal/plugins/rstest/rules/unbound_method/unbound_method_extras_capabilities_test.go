package unbound_method

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestUnboundMethodProgramCapabilities(t *testing.T) {
	root := txtarfs.MustParseFile(t, "testdata/capabilities.txtar").Materialize(t, "")
	root = tspath.NormalizePath(root)
	fs := bundled.WrapFS(osvfs.FS())
	raw, err := utils.CreateProgram(true, fs, root, "tsconfig.json", utils.CreateCompilerHost(root, fs))
	if err != nil {
		t.Fatal(err)
	}
	file := raw.GetSourceFile(tspath.ResolvePath(root, "input.ts"))
	if file == nil {
		t.Fatal("missing capability input")
	}
	sourceOnly, err := lintprogram.NewFromBoundSources(raw, []*ast.SourceFile{file})
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []bool{true, false} {
		t.Run(map[bool]string{true: "source-only", false: "tsconfig"}[source], func(t *testing.T) {
			program := lintprogram.NewFromCompiler(raw)
			if source {
				program = sourceOnly
			}
			var ran bool
			var reports []string
			plan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
				Programs: []*lintprogram.Program{program}, TargetsByProgram: [][]string{{file.FileName()}}, SingleThreaded: true,
				GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
					return []rule.ConfiguredRule{{
						Name: UnboundMethodRule.Name, RequiresTypeInfo: UnboundMethodRule.RequiresTypeInfo, Severity: rule.SeverityError,
						Run: func(ctx rule.RuleContext) rule.RuleListeners {
							ran = true
							if !raw.Options().Strict.IsTrue() {
								t.Error("tsconfig strict option was lost")
							}
							probeCapabilities(t, ctx)
							return UnboundMethodRule.Run(ctx, nil)
						},
					}}
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			result, err := linter.RunLinter(linter.RunLinterOptions{
				LintPlan: plan, SingleThreaded: true,
				Consumer: rule.DiagnosticConsumer{Report: func(d rule.RuleDiagnostic) {
					if d.Message.Id != "unboundWithoutThisAnnotation" {
						t.Errorf("unexpected message %s", d.Message.Id)
					}
					reports = append(reports, file.Text()[d.Range.Pos():d.Range.End()])
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			_, executed := result.ExecutedRules[UnboundMethodRule.Name]
			if source {
				if ran || executed || len(reports) != 0 {
					t.Fatal("source-only Program did not filter the actual Rstest rule")
				}
				return
			}
			if !ran || !executed {
				t.Fatal("tsconfig Program did not execute the Rstest rule")
			}
			want := []string{"service.method", "ambient.method", "Promise.all", "value.method", "value.method", "service.method", "service['method']", "Service.method"}
			if !reflect.DeepEqual(reports, want) {
				t.Fatalf("reports = %q, want %q", reports, want)
			}
			if diagnostics := raw.GetSemanticDiagnostics(context.Background(), file); len(diagnostics) != 0 {
				t.Fatalf("invalid capability fixture: %v", diagnostics)
			}
		})
	}
}

func probeCapabilities(t *testing.T, ctx rule.RuleContext) {
	t.Helper()
	seen := map[string]bool{}
	utils.VisitDescendants(ctx.SourceFile.AsNode(), func(node *ast.Node) bool {
		if ast.IsPropertyAccessExpression(node) && node.Name().Text() == "method" {
			symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
			if symbol == nil || symbol.ValueDeclaration == nil {
				t.Fatalf("method lost its declaration: %s", node.Expression().Text())
			}
			file := ast.GetSourceFileOfNode(symbol.ValueDeclaration).FileName()
			if node.Expression().Text() == "ambient" {
				if !strings.HasSuffix(file, "/ambient.d.ts") {
					t.Fatalf("ambient symbol source: %s", file)
				}
				seen["ambient"] = true
			} else if node.Expression().Text() == "service" || node.Expression().Text() == "value" {
				if !strings.HasSuffix(file, "/service.ts") {
					t.Fatalf("cross-file symbol source: %s", file)
				}
				seen[node.Expression().Text()] = true
				if node.Expression().Text() == "value" {
					typ := ctx.TypeChecker.GetTypeAtLocation(node.Expression())
					if typ.Flags()&checker.TypeFlagsUnion != 0 {
						t.Error("control-flow narrowing retained null")
					}
					if typ.Flags()&checker.TypeFlagsTypeParameter != 0 {
						seen["constraint"] = true
					} else {
						seen["narrowing"] = true
					}
				}
			}
		}
		if ast.IsVariableDeclaration(node) && ast.IsIdentifier(node.Name()) {
			switch node.Name().Text() {
			case "native":
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node.Initializer())
				if symbol == nil || symbol.ValueDeclaration == nil || !strings.Contains(ast.GetSourceFileOfNode(symbol.ValueDeclaration).FileName(), "lib.es5.d.ts") {
					t.Error("native symbol is not from default lib")
				}
				seen["native"] = true
			case "promise", "thenable":
				typ := ctx.TypeChecker.GetTypeAtLocation(node.Name())
				wantPromise := node.Name().Text() == "promise"
				if utils.IsPromiseLike(ctx.Program(), ctx.TypeChecker, typ) != wantPromise || !utils.IsThenableType(ctx.TypeChecker, node.Name(), typ) {
					t.Error("Promise/thenable type capability mismatch")
				}
				seen[node.Name().Text()] = true
			case "result":
				typ := ctx.TypeChecker.GetTypeAtLocation(node.Name())
				if typ.Flags()&checker.TypeFlagsObject == 0 || checker.Checker_getPropertyOfType(ctx.TypeChecker, typ, "method") == nil {
					t.Error("awaited type lost Service members")
				}
				seen["awaited"] = true
			}
		}
		return true
	})
	for _, name := range []string{"ambient", "service", "value", "constraint", "narrowing", "native", "promise", "thenable", "awaited"} {
		if !seen[name] {
			t.Errorf("missing capability evidence: %s", name)
		}
	}
}
