package rule

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

type programCacheTestKey struct{}

func TestRuleContextBindsOneProgramGeneration(t *testing.T) {
	raw := programCacheTestProgram(t)
	sourceFile := raw.GetSourceFile("/program-cache-fixture/file.ts")
	if sourceFile == nil {
		t.Fatal("fixture source file was not parsed")
	}
	if !sourceFile.IsBound() {
		binder.BindSourceFile(sourceFile)
	}
	sourceOnly, err := lintprogram.NewFromBoundSources(raw, []*ast.SourceFile{sourceFile})
	if err != nil {
		t.Fatalf("NewFromBoundSources: %v", err)
	}

	compilerBacked := lintprogram.NewFromCompiler(raw)
	ctx := (RuleContext{SourceFile: sourceFile}).WithProgram(compilerBacked)
	if ctx.Program() != compilerBacked {
		t.Fatal("context lost its Program generation")
	}
	if sourceOnly.CanProvideTypeChecker(sourceFile) {
		t.Fatal("source-only Program unexpectedly provided a checker")
	}

	defer func() {
		if recover() == nil {
			t.Fatal("rebinding a context to another Program did not panic")
		}
	}()
	_ = ctx.WithProgram(sourceOnly)
}

func TestRuleContextRejectsForeignSourceGeneration(t *testing.T) {
	owner := lintprogram.NewFromCompiler(programCacheTestProgram(t))
	foreignRaw := programCacheTestProgram(t)
	foreign := foreignRaw.GetSourceFile("/program-cache-fixture/file.ts")
	if foreign == nil {
		t.Fatal("foreign fixture source file was not parsed")
	}

	defer func() {
		if recover() == nil {
			t.Fatal("binding a foreign AST generation did not panic")
		}
	}()
	_ = (RuleContext{SourceFile: foreign}).WithProgram(owner)
}

func TestCachedByProgramSharesCompilerGenerationAcrossFacades(t *testing.T) {
	raw := programCacheTestProgram(t)
	firstContext := (RuleContext{}).WithProgram(lintprogram.NewFromCompiler(raw))
	secondContext := (RuleContext{}).WithProgram(lintprogram.NewFromCompiler(raw))

	builds := 0
	build := func() *int {
		builds++
		value := builds
		return &value
	}
	first := CachedByProgram(firstContext, programCacheTestKey{}, build)
	second := CachedByProgram(secondContext, programCacheTestKey{}, build)
	if builds != 1 || first != second {
		t.Fatalf("shared generation cache: builds=%d same=%v", builds, first == second)
	}
}

func programCacheTestProgram(t *testing.T) *compiler.Program {
	t.Helper()
	files := map[string]string{
		"/program-cache-fixture/file.ts":       `import "./dependency"; export const value = 1;`,
		"/program-cache-fixture/dependency.ts": "export const dependency = 1;\n",
	}
	fs := rslint_utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), files)
	host := rslint_utils.CreateCompilerHost("/", fs)
	raw, err := rslint_utils.CreateProgramFromOptions(true, &core.CompilerOptions{}, []string{"/program-cache-fixture/file.ts"}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptions: %v", err)
	}
	return raw
}
