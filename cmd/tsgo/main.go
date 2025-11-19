package main

import (
	"flag"
	"log"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/microsoft/typescript-go/shim/api"
	"github.com/microsoft/typescript-go/shim/api/encoder"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type EncodedSourceFile = []byte
type CheckResult = struct {
	ModuleList  []string            `json:"module_list"`
	SourceFiles []EncodedSourceFile `json:"source_files"`
	RootFiles   []string            `json:"root_files"`
}

func main() {
	os.Exit(runMain())
}
func runMain() int {
	var (
		config   string
		help     bool
		api_mode bool
	)
	flag.StringVar(&config, "config", "", "path to tsconfig.json")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&api_mode, "api", true, "api mode")
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
	checkResult.RootFiles = program.CommandLine().FileNames()

	for _, file := range program.GetSourceFiles() {
		checkResult.ModuleList = append(checkResult.ModuleList, string(file.FileName()))
		sourceFile := file.AsSourceFile()
		encodedSourceFile, err := encoder.EncodeSourceFile(sourceFile, string(api.FileHandle(sourceFile)))
		if err != nil {
			log.Printf("error encoding source file %v: %v", file.Path(), err)
			return 1
		}

		checkResult.SourceFiles = append(checkResult.SourceFiles, encodedSourceFile)
	}
	result, err := cbor.Marshal(checkResult)
	if err != nil {
		log.Printf("error marshaling checkResult: %v", err)
		return 1
	}
	_, err = os.Stdout.Write(result)
	if err != nil {
		log.Printf("error writing result to stdout: %v", err)
		return 1
	}
	return 0
}
