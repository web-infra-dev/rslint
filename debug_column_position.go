package main

import (
	"fmt"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func main() {
	// Test the exact multi-line case that's failing
	code := `
a
  ['SHOUT_CASE'];
`
	
	// Create filesystem and compiler host
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	fs = utils.NewOverlayVFSForFile("test.ts", code)
	host := utils.CreateCompilerHost(".", fs)
	
	// Create program
	program, err := utils.CreateProgram(true, fs, ".", "tsconfig.json", host)
	if err != nil {
		fmt.Printf("Error creating program: %v\n", err)
		return
	}
	
	sourceFile := program.GetSourceFile("test.ts")
	if sourceFile == nil {
		fmt.Printf("Source file not found\n")
		return
	}
	
	text := string(sourceFile.Text())
	fmt.Printf("Source text:\n%s\n", text)
	fmt.Printf("Source text length: %d\n", len(text))
	
	// Find each character's position
	for i, char := range text {
		if char == '[' {
			line, column := scanner.GetLineAndCharacterOfPosition(sourceFile, i)
			fmt.Printf("'[' found at byte position %d, line %d (0-based), column %d (0-based)\n", i, line, column)
			fmt.Printf("'[' 1-based position: line %d, column %d\n", line+1, column+1)
		}
	}
	
	// Also check what the exact positions are for the line content
	lines := []string{}
	currentLine := ""
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(char)
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	fmt.Printf("\nLines breakdown:\n")
	for i, line := range lines {
		fmt.Printf("Line %d: '%s'\n", i+1, line)
		for j, char := range line {
			if char == '[' {
				fmt.Printf("  '[' at column %d (1-based)\n", j+1)
			}
		}
	}
}