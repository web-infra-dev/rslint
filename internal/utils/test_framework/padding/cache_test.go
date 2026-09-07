package padding

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestEngineIsSharedAcrossRules(t *testing.T) {
	const code = "setup();\nbeforeAll(connect);\ntest('works', run);"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/padding.ts",
		Path:     "/padding.ts",
	}, code, core.ScriptKindTS)
	cache := rule.NewFileCache()
	ctx := rule.RuleContext{SourceFile: sourceFile}.WithFileCache(cache)

	first := rule.CachedByFile(ctx, paddingFileCacheKey{}, func() *engine {
		return &engine{sourceFile: sourceFile}
	})
	second := rule.CachedByFile(ctx, paddingFileCacheKey{}, func() *engine {
		t.Fatal("second padding rule replaced the shared engine")
		return nil
	})

	if first != second {
		t.Fatalf("cached engine was not reused: first=%p second=%p", first, second)
	}
}
