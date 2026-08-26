package server

import (
	"encoding/json"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	api "github.com/web-infra-dev/rslint/internal/api"
)

func TestHandleGetAstInfoProgramCacheInvalidation(t *testing.T) {
	cache := astInfoProgramCache
	cache.mu.Lock()
	originalFileContent := cache.fileContent
	originalCompilerOptions := cache.compilerOptions
	originalProgram := cache.program
	originalSourceFile := cache.sourceFile
	cache.fileContent = ""
	cache.compilerOptions = ""
	cache.program = nil
	cache.sourceFile = nil
	cache.mu.Unlock()
	t.Cleanup(func() {
		cache.mu.Lock()
		cache.fileContent = originalFileContent
		cache.compilerOptions = originalCompilerOptions
		cache.program = originalProgram
		cache.sourceFile = originalSourceFile
		cache.mu.Unlock()
	})

	handler := &Handler{}
	loadProgram := func(content string, compilerOptions map[string]any) *compiler.Program {
		t.Helper()
		response, err := handler.HandleGetAstInfo(api.GetAstInfoRequest{
			FileContent:     content,
			Kind:            int(ast.KindSourceFile),
			CompilerOptions: compilerOptions,
		})
		if err != nil {
			t.Fatalf("HandleGetAstInfo() error = %v", err)
		}
		if response.Node == nil || response.Node.Kind != int(ast.KindSourceFile) {
			t.Fatalf("HandleGetAstInfo() node = %#v, want SourceFile", response.Node)
		}

		optionsJSON := "{}"
		if compilerOptions != nil {
			encoded, err := json.Marshal(compilerOptions)
			if err != nil {
				t.Fatalf("marshal compiler options: %v", err)
			}
			optionsJSON = string(encoded)
		}
		program, sourceFile := getCachedProgram(content, optionsJSON)
		if program == nil || sourceFile == nil {
			t.Fatal("HandleGetAstInfo() did not publish its Program cache entry")
		}
		return program
	}

	const initialContent = "const value = 1\n"
	initialProgram := loadProgram(initialContent, nil)
	if cachedProgram := loadProgram(initialContent, nil); cachedProgram != initialProgram {
		t.Fatal("identical content and compiler options did not reuse the cached Program")
	}

	const changedContent = "const value = 2\n"
	if cachedProgram, _ := getCachedProgram(changedContent, "{}"); cachedProgram != nil {
		t.Fatal("changed content unexpectedly matched the cached Program")
	}
	changedContentProgram := loadProgram(changedContent, nil)
	if changedContentProgram == initialProgram {
		t.Fatal("changed content reused the previous Program")
	}

	changedOptions := map[string]any{"strict": false}
	changedOptionsJSON, err := json.Marshal(changedOptions)
	if err != nil {
		t.Fatalf("marshal changed compiler options: %v", err)
	}
	if cachedProgram, _ := getCachedProgram(changedContent, string(changedOptionsJSON)); cachedProgram != nil {
		t.Fatal("changed compiler options unexpectedly matched the cached Program")
	}
	if changedOptionsProgram := loadProgram(changedContent, changedOptions); changedOptionsProgram == changedContentProgram {
		t.Fatal("changed compiler options reused the previous Program")
	}
}
