package main

import (
	"log"
	"os"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func runBrowser() int {
	content, err := os.ReadFile("/tsconfig.json")
	if err != nil {
		log.Fatalf("Error reading tsconfig.json: %v", err)
		return 1
	}
	log.Printf("tsconfig.json: %v\n", string(content))

	rootDir := "/"
	tsconfigPath := "/tsconfig.json"
	fs := bundled.WrapFS(osvfs.FS());
	host := utils.CreateCompilerHost(rootDir, fs)

	program, err := utils.CreateProgram(true, fs, rootDir, tsconfigPath, host)

	if err != nil {
		log.Fatalf("Error creating TS program: %v", err)
		return 1
	} else {
		for _, file := range program.GetSourceFiles() {
			if(file.FileName() == "/input.ts") {
				log.Printf("file: %v\n,%v", file.Text(),file.Node)
			}
		}
		return 0
	}
}
