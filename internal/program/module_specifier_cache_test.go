package program

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
)

func parseModuleSpecifierCacheFile(source string) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/module-cache.ts",
		Path:     "/module-cache.ts",
	}, source, core.ScriptKindTS)
}

func TestModuleSpecifierCachePublishesOneResultPerSyntaxSet(t *testing.T) {
	file := parseModuleSpecifierCacheFile("import './target'; const value = require('./target');")
	esm := cachedModuleSpecifiers(file, ESModuleReferences)
	all := cachedModuleSpecifiers(file, ESModuleReferences|CommonJSReferences)
	if len(esm) != 1 || len(all) != 2 {
		t.Fatalf("cached module specifiers = (%d, %d), want (1, 2)", len(esm), len(all))
	}
	if again := cachedModuleSpecifiers(file, ESModuleReferences); &again[0] != &esm[0] {
		t.Fatal("the same syntax set published more than one result")
	}
}

func TestModuleSpecifierCacheUsesExactSourceFileIdentity(t *testing.T) {
	original := parseModuleSpecifierCacheFile("import './target';")
	edited := parseModuleSpecifierCacheFile("import './target'; const value = require('./target');")
	kinds := ESModuleReferences | CommonJSReferences
	before := cachedModuleSpecifiers(original, kinds)
	after := cachedModuleSpecifiers(edited, kinds)
	if len(before) != 1 || len(after) != 2 {
		t.Fatalf("distinct source generations returned %d and %d references", len(before), len(after))
	}
}

func TestModuleSpecifierCachePublishesConcurrentResult(t *testing.T) {
	file := parseModuleSpecifierCacheFile("import './target'; const value = require('./target');")
	kinds := ESModuleReferences | CommonJSReferences
	const callers = 32
	start := make(chan struct{})
	results := make(chan []moduleSpecifier, callers)
	for range callers {
		go func() {
			<-start
			results <- cachedModuleSpecifiers(file, kinds)
		}()
	}
	close(start)

	first := <-results
	if len(first) != 2 {
		t.Fatalf("cached %d module specifiers, want 2", len(first))
	}
	for range callers - 1 {
		result := <-results
		if len(result) != 2 || &result[0] != &first[0] {
			t.Fatal("concurrent callers did not observe one published result")
		}
	}
}
