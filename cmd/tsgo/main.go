package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/microsoft/typescript-go/shim/api/encoder"
	"github.com/microsoft/typescript-go/shim/api/protocol"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
	"gotest.tools/v3/internal/source"
)

type EncodedSourceFile = []byte
type CheckResult = struct {
	ModuleList  []string            `json:"module_list"`
	SourceFiles []EncodedSourceFile `json:"source_files"`
}

func main() {
	log.Println("tsgo")
	os.Exit(runMain())
}
func runMain() int {
	var (
		config string
		help   bool
	)
	flag.StringVar(&config, "config", "", "path to tsconfig.json")
	flag.BoolVar(&help, "help", false, "show help")
	flag.Parse()
	if help {
		flag.Usage()
		return 0
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	currentDirectory, err := os.Getwd()
	if err != nil {

		log.Printf("error getting current directory: %v", err)
		return 1
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)
	host := utils.CreateCompilerHost(currentDirectory, fs)
	program, err := utils.CreateProgram(
		true,
		fs,
		currentDirectory,
		config,
		host)
	if err != nil {
		log.Printf("error create Ts program: %v", err)
		return 1
	}
	checkResult := CheckResult{}

	for _, file := range program.GetSourceFiles() {
		checkResult.ModuleList = append(checkResult.ModuleList, string(file.Path()))
		sourceFile = file.AsSourceFile();
		encodedSourceFile := encoder.EncodeSourceFile(
			protocol.FileHandle(sourceFile)
		)
		checkResult.SourceFiles = append(checkResult.SourceFiles, encodedSourceFile)
	}
	//log.Printf("module_list len: %v", checkResult)
	//result, err := cbor.Marshal(checkResult)
	result, err := json.Marshal(checkResult)
	if err != nil {
		log.Printf("error marshaling checkResult: %v", err)
		return 1
	}
	log.Printf("vvvvv %v", result)
	return 0
}
