package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// SyntacticError carries structured diagnostics for syntax errors.
// Callers can type-assert to access the raw diagnostics for rich rendering.
type SyntacticError struct {
	Diagnostics []*ast.Diagnostic
	msg         string
}

func (e *SyntacticError) Error() string {
	return e.msg
}

func CreateCompilerHost(cwd string, fs vfs.FS) compiler.CompilerHost {
	defaultLibraryPath := bundled.LibPath()
	return compiler.NewCompilerHost(cwd, fs, defaultLibraryPath, nil, nil)
}

func CreateProgram(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost) (*compiler.Program, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !fs.FileExists(resolvedConfigPath) {
		return nil, fmt.Errorf("couldn't read tsconfig at %v", resolvedConfigPath)
	}

	configParseResult, _ := tsoptions.GetParsedCommandLineOfConfigFile(tsconfigPath, &core.CompilerOptions{}, nil, host, nil)

	return createProgramFromConfig(singleThreaded, configParseResult, host)
}

// CreateProgramFromOptions creates a program from in-memory compiler options and root file names,
// without requiring a tsconfig file on disk.
func CreateProgramFromOptions(singleThreaded bool, compilerOptions *core.CompilerOptions, rootFileNames []string, host compiler.CompilerHost) (*compiler.Program, error) {
	configParseResult := tsoptions.NewParsedCommandLine(compilerOptions, rootFileNames, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
		CurrentDirectory:          host.GetCurrentDirectory(),
	})

	return createProgramFromConfig(singleThreaded, configParseResult, host)
}

// CreateProgramFromOptionsLenient creates a program like CreateProgramFromOptions but
// tolerates syntactic errors. This is used for fallback programs where the user's source
// code may contain syntax errors (that's why they're running a linter).
func CreateProgramFromOptionsLenient(singleThreaded bool, compilerOptions *core.CompilerOptions, rootFileNames []string, host compiler.CompilerHost) (*compiler.Program, error) {
	configParseResult := tsoptions.NewParsedCommandLine(compilerOptions, rootFileNames, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
		CurrentDirectory:          host.GetCurrentDirectory(),
	})

	opts := compiler.ProgramOptions{
		Config:         configParseResult,
		SingleThreaded: core.TSTrue,
		Host:           host,
	}
	if !singleThreaded {
		opts.SingleThreaded = core.TSFalse
	}
	program := compiler.NewProgram(opts)
	if program == nil {
		return nil, errors.New("couldn't create program")
	}

	program.BindSourceFiles()
	return program, nil
}

func createProgramFromConfig(singleThreaded bool, config *tsoptions.ParsedCommandLine, host compiler.CompilerHost) (*compiler.Program, error) {
	opts := compiler.ProgramOptions{
		Config:         config,
		SingleThreaded: core.TSTrue,
		Host:           host,
	}
	if !singleThreaded {
		opts.SingleThreaded = core.TSFalse
	}
	program := compiler.NewProgram(opts)
	if program == nil {
		return nil, errors.New("couldn't create program")
	}

	syntacticDiags := program.GetSyntacticDiagnostics(context.Background(), nil)
	if len(syntacticDiags) != 0 {
		var msgs []string
		for _, d := range syntacticDiags {
			if d.File() != nil {
				line, col := scanner.GetECMALineAndUTF16CharacterOfPosition(d.File(), d.Pos())
				msgs = append(msgs, fmt.Sprintf("  %s(%d,%d): error TS%d: %s",
					d.File().FileName(), line+1, col+1, d.Code(), d.String()))
			} else {
				msgs = append(msgs, fmt.Sprintf("  error TS%d: %s", d.Code(), d.String()))
			}
		}
		return nil, &SyntacticError{
			Diagnostics: syntacticDiags,
			msg:         fmt.Sprintf("found %d syntactic error(s):\n%s", len(syntacticDiags), strings.Join(msgs, "\n")),
		}
	}

	program.BindSourceFiles()

	// program.CreateCheckers()

	return program, nil
}

// CollectProgramFiles collects all source file paths from the given programs
// into a set for fast lookup. Also stores symlink-resolved paths to handle
// platform differences (e.g. macOS /tmp → /private/tmp).
func CollectProgramFiles(programs []*compiler.Program, fs vfs.FS) map[string]struct{} {
	fileSet := make(map[string]struct{})
	for _, prog := range programs {
		for _, sf := range prog.GetSourceFiles() {
			name := sf.FileName()
			if _, ok := fileSet[name]; ok {
				continue
			}
			fileSet[name] = struct{}{}
			if resolved := fs.Realpath(name); resolved != name {
				fileSet[resolved] = struct{}{}
			}
		}
	}
	return fileSet
}

