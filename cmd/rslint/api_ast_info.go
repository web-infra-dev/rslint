package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/inspector"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// programCache holds a cached Program instance for AST info requests
type programCache struct {
	mu              sync.RWMutex
	fileContent     string
	compilerOptions string // JSON serialized for comparison
	program         *compiler.Program
	sourceFile      *ast.SourceFile
}

// Global program cache for AST info requests
var astInfoProgramCache = &programCache{}

// HandleGetAstInfo handles get AST info requests in IPC mode
func (h *IPCHandler) HandleGetAstInfo(req api.GetAstInfoRequest) (*api.GetAstInfoResponse, error) {
	// Fixed user file name for program creation
	const userFileName = "/index.ts"

	// Serialize compiler options for comparison
	compilerOptionsJSON := "{}"
	if req.CompilerOptions != nil {
		jsonBytes, err := json.Marshal(req.CompilerOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal compiler options: %w", err)
		}
		compilerOptionsJSON = string(jsonBytes)
	}

	// Check if we can use cached program
	program, userSourceFile := getCachedProgram(req.FileContent, compilerOptionsJSON)
	if program == nil || userSourceFile == nil {
		// Cache miss - create new program
		var err error
		program, userSourceFile, err = createAndCacheProgram(userFileName, req.FileContent, compilerOptionsJSON, req.CompilerOptions)
		if err != nil {
			return nil, err
		}
	}

	// Get type checker
	typeChecker, done := program.GetTypeChecker(context.Background())
	defer done()

	// Determine which source file to query
	// If FileName is set to an external file, query that file (e.g., lib.d.ts)
	// Otherwise, query the user's source file
	var targetSourceFile *ast.SourceFile
	if req.FileName != "" && req.FileName != userFileName {
		targetSourceFile = program.GetSourceFile(req.FileName)
		if targetSourceFile == nil {
			return &api.GetAstInfoResponse{}, nil
		}
	} else {
		targetSourceFile = userSourceFile
	}

	isExternalFile := targetSourceFile != userSourceFile

	// Build the response
	// Use userSourceFile as the "current" file for the builder
	// This determines which files are considered "external" (fileName will be set for nodes not in userSourceFile)
	builder := api.NewAstInfoBuilder(typeChecker, userSourceFile)
	response := &api.GetAstInfoResponse{}

	// Special case: if requesting SourceFile by kind, build it directly without Node conversion
	if req.Kind > 0 && ast.Kind(req.Kind) == ast.KindSourceFile {
		response.Node = builder.BuildSourceFileNodeInfo(targetSourceFile)
		// SourceFile doesn't have type/symbol/signature/flow, so return early
		return response, nil
	}

	// Find the node at the specified position (with optional end for exact matching)
	node := inspector.FindNodeAtPosition(targetSourceFile, req.Position, req.End, req.Kind)
	if node == nil {
		return &api.GetAstInfoResponse{}, nil
	}

	// Build node info
	response.Node = builder.BuildNodeInfo(node)

	// Build type info
	t := inspector.GetTypeAtNode(typeChecker, node)
	if t != nil {
		response.Type = builder.BuildTypeInfo(t)
	}

	// Build symbol info
	// First try to get symbol directly from node
	symbol := typeChecker.GetSymbolAtLocation(node)
	// If no symbol at node, try to get it from the type
	if symbol == nil && t != nil {
		symbol = t.Symbol()
	}
	if symbol != nil {
		response.Symbol = builder.BuildSymbolInfo(symbol)
	}

	// Build signature info
	sig := inspector.GetSignatureOfNode(typeChecker, node)
	if sig != nil {
		response.Signature = builder.BuildSignatureInfo(sig)
	}

	// Build flow info (only for nodes in user's source file)
	if !isExternalFile {
		flowNode := inspector.GetFlowNodeOfNode(node)
		if flowNode != nil {
			response.Flow = builder.BuildFlowInfo(flowNode)
		}
	}

	return response, nil
}

// getCachedProgram returns the cached program if it matches the current request
func getCachedProgram(fileContent, compilerOptionsJSON string) (*compiler.Program, *ast.SourceFile) {
	astInfoProgramCache.mu.RLock()
	defer astInfoProgramCache.mu.RUnlock()

	if astInfoProgramCache.program == nil {
		return nil, nil
	}

	// Check if cache is valid (only fileContent and compilerOptions matter)
	if astInfoProgramCache.fileContent == fileContent &&
		astInfoProgramCache.compilerOptions == compilerOptionsJSON {
		return astInfoProgramCache.program, astInfoProgramCache.sourceFile
	}

	return nil, nil
}

// createAndCacheProgram creates a new program and caches it
func createAndCacheProgram(fileName, fileContent, compilerOptionsJSON string, compilerOptions map[string]any) (*compiler.Program, *ast.SourceFile, error) {
	// Create a virtual filesystem with the provided file content
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	fileContents := map[string]string{
		fileName: fileContent,
	}
	fs = utils.NewOverlayVFS(fs, fileContents)

	// Build tsconfig from request options or use defaults
	tsconfigContent := buildTsConfigContent(fileName, compilerOptions)
	tsconfigPath := "/tsconfig.json"
	fs = utils.NewOverlayVFS(fs, map[string]string{
		tsconfigPath: tsconfigContent,
	})

	// Create compiler host and program
	host := utils.CreateCompilerHost("/", fs)
	program, err := utils.CreateProgram(false, fs, "/", tsconfigPath, host)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create program: %w", err)
	}

	// Get the source file
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		return nil, nil, errors.New("failed to get source file")
	}

	// Update cache
	astInfoProgramCache.mu.Lock()
	astInfoProgramCache.fileContent = fileContent
	astInfoProgramCache.compilerOptions = compilerOptionsJSON
	astInfoProgramCache.program = program
	astInfoProgramCache.sourceFile = sourceFile
	astInfoProgramCache.mu.Unlock()

	return program, sourceFile, nil
}

// buildTsConfigContent creates a tsconfig.json content string from compiler options
func buildTsConfigContent(fileName string, compilerOptions map[string]any) string {
	// Default compiler options
	opts := map[string]any{
		"target":           "ESNext",
		"module":           "ESNext",
		"strict":           true,
		"strictNullChecks": true,
	}

	// Merge with provided options (provided options override defaults)
	for k, v := range compilerOptions {
		opts[k] = v
	}

	// Serialize compiler options to JSON
	optsJSON, err := json.Marshal(opts)
	if err != nil {
		// Fallback to minimal config on error
		return fmt.Sprintf(`{"compilerOptions":{"target":"ESNext","module":"ESNext","strict":true},"files":["%s"]}`, fileName)
	}

	return fmt.Sprintf(`{"compilerOptions":%s,"files":["%s"]}`, string(optsJSON), fileName)
}
