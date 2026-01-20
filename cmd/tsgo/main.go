package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/microsoft/typescript-go/shim/api"
	"github.com/microsoft/typescript-go/shim/api/encoder"
	"github.com/microsoft/typescript-go/shim/ast"
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

func getDiagnostics(diagnostics []*ast.Diagnostic, fileMap *map[string]int32) []Diagnostics {
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

func printDiagnostics(diagnostics []Diagnostics, fileMap map[string]int32) {
	if len(diagnostics) == 0 {
		log.Println("✓ No diagnostics found.")
		return
	}

	// Count by category
	var warnings, errors, suggestions, messages int
	for _, diag := range diagnostics {
		switch diag.Category {
		case 0:
			warnings++
		case 1:
			errors++
		case 2:
			suggestions++
		case 3:
			messages++
		}
	}

	log.Printf("\n Found %d diagnostic(s): %d error(s), %d warning(s), %d suggestion(s), %d message(s)\n",
		len(diagnostics), errors, warnings, suggestions, messages)
	log.Println(strings.Repeat("─", 80))

	for i, diag := range diagnostics {
		fileName := ""
		for file, id := range fileMap {
			if id == diag.File {
				fileName = file
				break
			}
		}

		var category, icon string
		switch diag.Category {
		case 0:
			category = "Warning"
			icon = "⚠"
		case 1:
			category = "Error"
			icon = "✗"
		case 2:
			category = "Suggestion"
			icon = "💡"
		case 3:
			category = "Message"
			icon = "ℹ"
		default:
			category = "Unknown"
			icon = "•"
		}

		log.Printf("\n %s %s\n", icon, category)
		log.Printf("   %s\n", diag.Message)
		log.Printf("   → %s:%d-%d\n", fileName, diag.Loc.Start, diag.Loc.End)

		if i < len(diagnostics)-1 {
			log.Println(strings.Repeat("─", 80))
		}
	}
	log.Println()
}
func runMain() int {
	var (
		config          string
		help            bool
		api_mode        bool
		transpileServer bool
		serverCwd       string
		transpile       bool
		transpileFile   string
		transpileModule string
		transpileTarget string
		transpileJsx    string
		inlineSourceMap bool
		sourceMap       bool
		typecheck       bool
	)
	flag.StringVar(&config, "config", "", "path to tsconfig.json")
	flag.BoolVar(&help, "help", false, "show help")
	flag.BoolVar(&api_mode, "api", false, "api mode")
	flag.BoolVar(&transpileServer, "transpileServer", false, "run transpile server mode")
	flag.StringVar(&serverCwd, "cwd", "", "current working directory for transpile server")
	flag.BoolVar(&transpile, "transpile", false, "transpile a single file to JS")
	flag.StringVar(&transpileFile, "file", "", "path to input file for --transpile")
	flag.StringVar(&transpileModule, "module", "", "module kind for --transpile (commonjs, esnext, node16, nodenext, etc)")
	flag.StringVar(&transpileTarget, "target", "", "script target for --transpile (es2019, es2020, esnext, etc)")
	flag.StringVar(&transpileJsx, "jsx", "", "jsx emit for --transpile (react, react-jsx, preserve, etc)")
	flag.BoolVar(&inlineSourceMap, "inlineSourceMap", false, "emit inline source map for --transpile")
	flag.BoolVar(&sourceMap, "sourceMap", false, "emit source map for --transpile")
	flag.BoolVar(&typecheck, "typecheck", false, "enable typechecking for --transpile")
	flag.Parse()
	if help {
		flag.Usage()
		return 0
	}
	if transpileServer {
		cwd := serverCwd
		if cwd == "" {
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				log.Printf("error resolving cwd: %v", err)
				return 1
			}
		}
		return runTranspileServer(cwd)
	}
	if transpile {
		return runTranspile(transpileOptions{
			Config:          config,
			File:            transpileFile,
			Module:          transpileModule,
			Target:          transpileTarget,
			Jsx:             transpileJsx,
			InlineSourceMap: inlineSourceMap,
			SourceMap:       sourceMap,
			TypeCheck:       typecheck,
		})
	}
	program, err := CreateProgram(config)
	if err != nil {
		log.Printf("error creating program: %v", err)
		return 1
	}
	diagnostics := compiler.GetDiagnosticsOfAnyProgram(context.Background(), program, nil, false, func(ctx context.Context, file *ast.SourceFile) []*ast.Diagnostic {

		diags := program.GetBindDiagnostics(ctx, file)
		return diags
	},
		func(ctx context.Context, file *ast.SourceFile) []*ast.Diagnostic {
			diags := program.GetSemanticDiagnostics(ctx, file)
			return diags
		})
	tc, done := program.GetTypeChecker(context.Background())

	defer done()

	checkResult := CheckResult{}
	checkResult.RootFiles = program.CommandLine().FileNames()
	checkResult.Semantic = NewSemantic()
	checkResult.SourceFileExtra = []SourceFileExtra{}

	initPrimitiveTypes(tc, &checkResult.Semantic)
	fileMap := make(map[string]int32)
	for sourcefileId, file := range program.GetSourceFiles() {
		fileMap[file.FileName()] = int32(sourcefileId)
		checkResult.ModuleList = append(checkResult.ModuleList, file.FileName())
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
	checkResult.Diagnostics = getDiagnostics(diagnostics, &fileMap)

	if !api_mode {
		// Print diagnostics in human-readable format
		printDiagnostics(checkResult.Diagnostics, fileMap)
		if len(checkResult.Diagnostics) > 0 {
			return 1
		}
		return 0
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
