package no_deprecated_api

import (
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func collectNoTypecheckReferences(code string, iterate func(*referenceTracker)) []trackedReference {
	sf := parser.ParseSourceFile(
		ast.SourceFileParseOptions{FileName: "/file.js", Path: "/file.js"},
		code,
		core.ScriptKindJS,
	)
	tracker := &referenceTracker{sourceFile: sf}
	collectIdentifiers(sf.AsNode(), &tracker.allIdentifiers)
	return tracker.capture(func() {
		iterate(tracker)
	})
}

func assertSingleReadPath(t *testing.T, refs []trackedReference, want string) {
	t.Helper()
	if len(refs) != 1 {
		t.Fatalf("expected one reference, got %d: %#v", len(refs), refs)
	}
	if refs[0].typ != refRead {
		t.Fatalf("expected read reference, got %v", refs[0].typ)
	}
	if got := strings.Join(refs[0].path, "."); got != want {
		t.Fatalf("expected path %q, got %q", want, got)
	}
}

func TestProcessGetBuiltinModuleNoTypecheckFallback(t *testing.T) {
	cases := []string{
		`process.getBuiltinModule('fs').exists`,
		`process['getBuiltinModule']('fs').exists`,
		`const { getBuiltinModule } = process; getBuiltinModule('fs').exists`,
		`const { getBuiltinModule: gbm } = process; gbm('fs').exists`,
		`const getBuiltinModule = process.getBuiltinModule; getBuiltinModule('fs').exists`,
		`const getBuiltinModule = process['getBuiltinModule']; getBuiltinModule('fs').exists`,
	}
	for _, code := range cases {
		t.Run(code, func(t *testing.T) {
			refs := collectNoTypecheckReferences(code, func(tracker *referenceTracker) {
				tracker.iterateProcessGetBuiltinModuleReferences(modules)
			})
			assertSingleReadPath(t, refs, "fs.exists")
		})
	}
}

func TestLhsAssignmentNoTypecheckFallback(t *testing.T) {
	t.Run("unresolved identifier assignment does not propagate", func(t *testing.T) {
		refs := collectNoTypecheckReferences(`b = require('fs'); b.exists`, func(tracker *referenceTracker) {
			tracker.iterateCjsReferences(modules)
		})
		if len(refs) != 0 {
			t.Fatalf("expected no references, got %#v", refs)
		}
	})

	t.Run("resolved identifier assignment propagates", func(t *testing.T) {
		refs := collectNoTypecheckReferences(`let b; b = require('fs'); b.exists`, func(tracker *referenceTracker) {
			tracker.iterateCjsReferences(modules)
		})
		assertSingleReadPath(t, refs, "fs.exists")
	})

	t.Run("object assignment destructuring still reports the member read", func(t *testing.T) {
		refs := collectNoTypecheckReferences(`({ exists } = require('fs')); exists`, func(tracker *referenceTracker) {
			tracker.iterateCjsReferences(modules)
		})
		assertSingleReadPath(t, refs, "fs.exists")
	})
}
