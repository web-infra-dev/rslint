package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/microsoft/typescript-go/shim/api"
	"github.com/microsoft/typescript-go/shim/api/encoder"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
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
	SymbolTable SymbolTable         `json:"symbol_table"`
	TypeTable   TypeTable           `json:"type_table"`
}

func CreateProgram(config string) (*compiler.Program, error) {
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return program, nil
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
	program, err := CreateProgram(config)
	tc, done := program.GetTypeChecker(context.Background())
	defer done()
	if err != nil {
		log.Printf("error creating program: %v", err)
		return 1
	}
	checkResult := CheckResult{}
	checkResult.SymbolTable = SymbolTable{}
	checkResult.TypeTable = TypeTable{}
	checkResult.RootFiles = program.CommandLine().FileNames()

	for sourcefileId, file := range program.GetSourceFiles() {
		checkResult.ModuleList = append(checkResult.ModuleList, string(file.FileName()))
		sourceFile := file.AsSourceFile()

		encodedSourceFile, err := encoder.EncodeSourceFile(sourceFile, string(api.FileHandle(sourceFile)))
		if err != nil {
			log.Printf("error encoding source file %v: %v", file.Path(), err)
			return 1
		}

		checkResult.SourceFiles = append(checkResult.SourceFiles, encodedSourceFile)
		CollectSemanticInFile(tc, file, &checkResult.SymbolTable, &checkResult.TypeTable, sourcefileId)
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
