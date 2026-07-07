package utils_test

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
)

func TestResolveSourceFileFromSourceFile(t *testing.T) {
	t.Parallel()

	ctx, specifier := contextForImport(t, "./bar")

	resolvedPath, target, ok := import_utils.ResolveSourceFileFromSourceFile(ctx, ctx.SourceFile, specifier)
	if !ok {
		t.Fatal("ResolveSourceFileFromSourceFile() did not resolve ./bar")
	}
	if target == nil {
		t.Fatal("ResolveSourceFileFromSourceFile() returned nil target")
	}
	if got := tspath.NormalizeSlashes(resolvedPath); !strings.HasSuffix(got, "/bar.ts") {
		t.Fatalf("resolvedPath = %q, want suffix /bar.ts", resolvedPath)
	}
	if target.FileName() != resolvedPath {
		t.Fatalf("target.FileName() = %q, want %q", target.FileName(), resolvedPath)
	}
}

func TestResolveModuleReferenceFromSourceFileInvalidInput(t *testing.T) {
	t.Parallel()

	ctx, _ := contextForImport(t, "./bar")

	if resolvedPath, target, ok := import_utils.ResolveModuleReferenceFromSourceFile(ctx, ctx.SourceFile, nil); ok || resolvedPath != "" || target != nil {
		t.Fatalf("ResolveModuleReferenceFromSourceFile(nil) = (%q, %#v, %v), want empty result", resolvedPath, target, ok)
	}
}
