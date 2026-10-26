// Test-only producer for the Node conformance suites. Production uses the
// already-owned Program SourceFile and never reparses in the plugin path.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
)

func main() {
	var input struct {
		FilePath   string `json:"filePath"`
		Text       string `json:"text"`
		SourceType string `json:"sourceType"`
		TextBytes  []byte `json:"textBytes"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		panic(err)
	}
	if input.TextBytes != nil {
		input.Text = string(input.TextBytes)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := tspath.ResolvePath(tspath.NormalizePath(cwd), input.FilePath)
	fileName := tspath.RootedFilePathFromAbsolute(path)
	kind := core.GetScriptKindFromFileName(fileName)
	hasBOM := strings.HasPrefix(input.Text, "\ufeff")
	text := strings.TrimPrefix(input.Text, "\ufeff")
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: fileName, PathKey: tspath.CaseSensitive.PathKey(fileName.AsPath())}, text, kind)
	if len(file.Diagnostics()) != 0 {
		fmt.Fprintln(os.Stderr, "test source contains syntax errors")
		os.Exit(1)
	}
	var builder estree.Builder
	data, err := builder.Build(context.Background(), file, estree.Options{SourceType: input.SourceType, HasBOM: hasBOM})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(data); err != nil {
		panic(err)
	}
}
