package estree_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
)

func TestBuildWithOriginsAndProjectOriginsParity(t *testing.T) {
	sourceText := `
		export function f() {
			void import.meta;
			const t = new.target;
			const chain = (a as any)?.b?.c;
			const seq = (1, 2, 3);
			return [t, chain, seq];
		}
	`
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test/origins.ts",
		Path:     tspath.Path("/test/origins.ts"),
	}, sourceText, core.ScriptKindTS)
	table := encoder.GetNodeIndexTable(file)
	opts := estree.Options{SourceType: "module"}

	var builder estree.Builder
	frame, withOrigins, err := builder.BuildWithOrigins(t.Context(), file, opts, table)
	if err != nil {
		t.Fatal(err)
	}
	if len(frame) == 0 {
		t.Fatal("expected non-empty ESTree frame")
	}
	builder.Reset()

	projectedOnly, err := builder.ProjectOrigins(t.Context(), file, opts, table)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(withOrigins, projectedOnly) {
		t.Fatalf("ProjectOrigins diverged from BuildWithOrigins:\nwithOrigins=%+v\nprojectedOnly=%+v", withOrigins, projectedOnly)
	}
	if len(withOrigins.FirstTokens) < 2 {
		t.Fatalf("expected MetaProperty firstTokens for import.meta and new.target, got %+v", withOrigins.FirstTokens)
	}
	if len(withOrigins.Aliases) == 0 {
		t.Fatalf("expected SequenceExpression aliases for (1, 2, 3)")
	}
	if len(withOrigins.ChainAliases) == 0 {
		t.Fatalf("expected ChainExpression aliases for optional chain")
	}
}
