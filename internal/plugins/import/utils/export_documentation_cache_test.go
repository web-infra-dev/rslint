package utils_test

import (
	"sync"
	"testing"

	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

func TestExportDocumentationAcrossSettingsAndReexports(t *testing.T) {
	contexts := contextsForFiles(t, map[string]string{
		"consumer.ts": `import {old} from './barrel';`,
		"barrel.ts":   `export {old} from './target';`,
		"target.ts":   "/** @module target\n * @deprecated module reason\n */\n// Deprecated: TomDoc reason\nexport const old=1;",
	}, "consumer.ts")
	ctx := contexts[0]
	request := modules.SourceFromSpecifier(firstImportSpecifier(t, ctx.SourceFile))
	exports, ok := import_utils.GetExportMap(ctx, request)
	if !ok || exports.Get("old") == nil {
		t.Fatal("expected re-export metadata")
	}
	if exports.Deprecation(ctx) != nil {
		t.Fatal("dependency module docs leaked onto the barrel")
	}
	// The preceding block supplies JSDoc metadata with default settings.
	if dep := exports.Get("old").Deprecation(ctx); dep == nil || dep.Description != "module reason" {
		t.Fatalf("default documentation: %#v", dep)
	}
	disabled := ctx
	disabled.Settings = map[string]any{"import/docstyle": []any{}}
	if dep := exports.Get("old").Deprecation(disabled); dep != nil {
		t.Fatalf("disabled documentation: %#v", dep)
	}
	tomdoc := ctx
	tomdoc.Settings = map[string]any{"import/docstyle": []any{"tomdoc"}}
	// TomDoc sees the first block, whose text doesn't begin with a status.
	if dep := exports.Get("old").Deprecation(tomdoc); dep != nil {
		t.Fatalf("TomDoc reused default documentation: %#v", dep)
	}
	var wait sync.WaitGroup
	for range 4 {
		wait.Go(func() {
			for range 10 {
				if dep := exports.Get("old").Deprecation(ctx); dep == nil || dep.Description != "module reason" {
					t.Errorf("concurrent documentation: %#v", dep)
				}
			}
		})
	}
	wait.Wait()
}
