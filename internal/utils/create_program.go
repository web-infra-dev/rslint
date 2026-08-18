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

func configNotFoundError(resolvedConfigPath string) error {
	return fmt.Errorf("couldn't read tsconfig at %v", resolvedConfigPath)
}

func CreateProgram(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost) (*compiler.Program, error) {
	return createProgram(singleThreaded, fs, cwd, tsconfigPath, host, false)
}

// CreateProgramUsingProjectReferenceSources creates a Program that redirects
// project-reference declaration outputs back to their source files. This is
// the mode used by the TypeScript language service for live source analysis.
func CreateProgramUsingProjectReferenceSources(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost) (*compiler.Program, error) {
	return createProgram(singleThreaded, fs, cwd, tsconfigPath, host, true)
}

func createProgram(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost, useSourceOfProjectReference bool) (*compiler.Program, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !fs.FileExists(resolvedConfigPath) {
		return nil, configNotFoundError(resolvedConfigPath)
	}

	configParseResult, _ := tsoptions.GetParsedCommandLineOfConfigFile(tsconfigPath, &core.CompilerOptions{}, nil, host, nil)

	return createProgramFromConfig(singleThreaded, configParseResult, host, useSourceOfProjectReference)
}

// CreateProgramLenient creates a tsconfig-backed Program but tolerates
// syntactic errors. Callers that own diagnostic admission use this so their
// final target set decides which syntax diagnostics are user-visible.
func CreateProgramLenient(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost) (*compiler.Program, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !fs.FileExists(resolvedConfigPath) {
		return nil, configNotFoundError(resolvedConfigPath)
	}

	configParseResult, _ := tsoptions.GetParsedCommandLineOfConfigFile(tsconfigPath, &core.CompilerOptions{}, nil, host, nil)

	return CreateProgramFromParsedConfigLenient(singleThreaded, configParseResult, host)
}

// CreateProgramFromOptions creates a program from in-memory compiler options and root file names,
// without requiring a tsconfig file on disk.
func CreateProgramFromOptions(singleThreaded bool, compilerOptions *core.CompilerOptions, rootFileNames []string, host compiler.CompilerHost) (*compiler.Program, error) {
	configParseResult := tsoptions.NewParsedCommandLine(compilerOptions, rootFileNames, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
		CurrentDirectory:          host.GetCurrentDirectory(),
	})

	return createProgramFromConfig(singleThreaded, configParseResult, host, false)
}

// CreateProgramFromOptionsLenient creates a Program like
// CreateProgramFromOptions but leaves syntax-diagnostic admission to its
// caller.
func CreateProgramFromOptionsLenient(singleThreaded bool, compilerOptions *core.CompilerOptions, rootFileNames []string, host compiler.CompilerHost) (*compiler.Program, error) {
	configParseResult := tsoptions.NewParsedCommandLine(compilerOptions, rootFileNames, tspath.ComparePathsOptions{
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
		CurrentDirectory:          host.GetCurrentDirectory(),
	})

	return CreateProgramFromParsedConfigLenient(singleThreaded, configParseResult, host)
}

// CreateProgramFromParsedConfigLenient creates and binds a Program from an
// already parsed command line while tolerating source syntax errors. Program
// loaders that own config parsing use this to preserve their cache boundary
// without duplicating compiler construction.
func CreateProgramFromParsedConfigLenient(singleThreaded bool, config *tsoptions.ParsedCommandLine, host compiler.CompilerHost) (*compiler.Program, error) {
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

	program.BindSourceFiles()
	return program, nil
}

func createProgramFromConfig(singleThreaded bool, config *tsoptions.ParsedCommandLine, host compiler.CompilerHost, useSourceOfProjectReference bool) (*compiler.Program, error) {
	opts := compiler.ProgramOptions{
		Config:                      config,
		SingleThreaded:              core.TSTrue,
		Host:                        host,
		UseSourceOfProjectReference: useSourceOfProjectReference,
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
