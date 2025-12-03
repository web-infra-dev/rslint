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
	ModuleList      []string            `json:"module_list"`
	SourceFiles     []EncodedSourceFile `json:"source_files"`
	RootFiles       []string            `json:"root_files"`
	Semantic        Semantic            `json:"semantic"`
	Diagnostics     []Diagnostics       `json:"diagnostics"`
	SourceFileExtra []SourceFileExtra   `json:"source_file_extra"`
}
type SourceFileExtra = struct {
	HasExternalModuleIndicator bool `json:"has_external_module_indicator"`
	HasCommonJSModuleIndicator bool `json:"has_common_js_module_indicator"`
}
type Location = struct {
	Start int32 `json:"start"`
	End   int32 `json:"end"`
}
type Diagnostics = struct {
	Message  string   `json:"message"`
	Category int32    `json:"category"`
	File     int32    `json:"file"`
	Loc      Location `json:"loc"`
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

func getDiagnostics(program *compiler.Program, fileMap *map[string]int32) []Diagnostics {
	diagnostics := program.GetSemanticDiagnostics(context.Background(), nil)
	diagnostics = append(diagnostics, program.GetGlobalDiagnostics(context.Background())...)
	diagnostics = append(diagnostics, program.GetOptionsDiagnostics(context.Background())...)
	diags := []Diagnostics{}
	for _, diag := range diagnostics {
		diags = append(diags, Diagnostics{
			Message:  diag.Message(),
			Category: int32(diag.Category()),
			File:     (*fileMap)[diag.File().FileName()],
			Loc: Location{
				Start: int32(diag.Pos()),
				End:   int32(diag.End()),
			},
		})
	}
	return diags
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
	checkResult.RootFiles = program.CommandLine().FileNames()
	checkResult.Semantic = NewSemantic()
	checkResult.SourceFileExtra = []SourceFileExtra{}

	initPrimitiveTypes(tc, &checkResult.Semantic)
	fileMap := make(map[string]int32)
	for sourcefileId, file := range program.GetSourceFiles() {
		fileMap[string(file.FileName())] = int32(sourcefileId)
		checkResult.ModuleList = append(checkResult.ModuleList, string(file.FileName()))
		sourceFile := file.AsSourceFile()

		encodedSourceFile, err := encoder.EncodeSourceFile(sourceFile, string(api.FileHandle(sourceFile)))
		if err != nil {
			log.Printf("error encoding source file %v: %v", file.Path(), err)
			return 1
		}
		checkResult.SourceFileExtra = append(checkResult.SourceFileExtra, SourceFileExtra{
			HasExternalModuleIndicator: sourceFile.ExternalModuleIndicator != nil,
			HasCommonJSModuleIndicator: sourceFile.CommonJSModuleIndicator != nil,
		})
		checkResult.SourceFiles = append(checkResult.SourceFiles, encodedSourceFile)

		CollectSemanticInFile(tc, file, &checkResult.Semantic, sourcefileId)
	}
	checkResult.Diagnostics = getDiagnostics(program, &fileMap)

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
