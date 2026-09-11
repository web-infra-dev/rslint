package server

import (
	"encoding/json"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	api "github.com/web-infra-dev/rslint/internal/api"
)

func TestHandleGetAstInfoProgramCacheInvalidation(t *testing.T) {
	cache := astInfoProgramCache
	cache.mu.Lock()
	originalFileContent := cache.fileContent
	originalSourceFileName := cache.sourceFileName
	originalCompilerOptions := cache.compilerOptions
	originalProgram := cache.program
	originalSourceFile := cache.sourceFile
	cache.fileContent = ""
	cache.sourceFileName = ""
	cache.compilerOptions = ""
	cache.program = nil
	cache.sourceFile = nil
	cache.mu.Unlock()
	t.Cleanup(func() {
		cache.mu.Lock()
		cache.fileContent = originalFileContent
		cache.sourceFileName = originalSourceFileName
		cache.compilerOptions = originalCompilerOptions
		cache.program = originalProgram
		cache.sourceFile = originalSourceFile
		cache.mu.Unlock()
	})

	handler := &Handler{}
	loadProgram := func(fileName, content string, compilerOptions map[string]any) *compiler.Program {
		t.Helper()
		response, err := handler.HandleGetAstInfo(api.GetAstInfoRequest{
			FileContent:     content,
			SourceFileName:  fileName,
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
		program, sourceFile := getCachedProgram(fileName, content, optionsJSON)
		if program == nil || sourceFile == nil {
			t.Fatal("HandleGetAstInfo() did not publish its Program cache entry")
		}
		return program
	}

	const initialContent = "const value = 1\n"
	initialProgram := loadProgram("/index.ts", initialContent, nil)
	if cachedProgram := loadProgram("/index.ts", initialContent, nil); cachedProgram != initialProgram {
		t.Fatal("identical content and compiler options did not reuse the cached Program")
	}
	jsxProgram := loadProgram("/index.tsx", initialContent, nil)
	if jsxProgram == initialProgram {
		t.Fatal("changed source filename reused the previous Program")
	}

	const changedContent = "const value = 2\n"
	if cachedProgram, _ := getCachedProgram("/index.tsx", changedContent, "{}"); cachedProgram != nil {
		t.Fatal("changed content unexpectedly matched the cached Program")
	}
	changedContentProgram := loadProgram("/index.tsx", changedContent, nil)
	if changedContentProgram == jsxProgram {
		t.Fatal("changed content reused the previous Program")
	}

	changedOptions := map[string]any{"strict": false}
	changedOptionsJSON, err := json.Marshal(changedOptions)
	if err != nil {
		t.Fatalf("marshal changed compiler options: %v", err)
	}
	if cachedProgram, _ := getCachedProgram("/index.tsx", changedContent, string(changedOptionsJSON)); cachedProgram != nil {
		t.Fatal("changed compiler options unexpectedly matched the cached Program")
	}
	if changedOptionsProgram := loadProgram("/index.tsx", changedContent, changedOptions); changedOptionsProgram == changedContentProgram {
		t.Fatal("changed compiler options reused the previous Program")
	}
}

func TestHandleGetAstInfoUsesRequestedSourceFileName(t *testing.T) {
	handler := &Handler{}
	for _, testCase := range []struct {
		name     string
		fileName string
		content  string
	}{
		{name: "default", content: "const value = 1\n"},
		{name: "JavaScript", fileName: "index.js", content: "const value = 1\n"},
		{name: "TypeScript JSX", fileName: "index.tsx", content: "const value = <div />\n"},
		{name: "JavaScript JSX", fileName: "index.jsx", content: "const value = <div />\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			response, err := handler.HandleGetAstInfo(api.GetAstInfoRequest{
				FileContent:    testCase.content,
				SourceFileName: testCase.fileName,
				Kind:           int(ast.KindSourceFile),
			})
			if err != nil {
				t.Fatalf("HandleGetAstInfo() error = %v", err)
			}
			if response.Node == nil || response.Node.Kind != int(ast.KindSourceFile) {
				t.Fatalf("HandleGetAstInfo() node = %#v, want SourceFile", response.Node)
			}

			wantFileName := testCase.fileName
			if wantFileName == "" {
				wantFileName = "index.ts"
			}
			if astInfoProgramCache.sourceFile == nil || astInfoProgramCache.sourceFile.FileName() != "/"+wantFileName {
				t.Fatalf("source file = %v, want /%s", astInfoProgramCache.sourceFile, wantFileName)
			}
		})
	}
}
