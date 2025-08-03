package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/collections"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

func CreateCompilerHost(cwd string, fs vfs.FS) compiler.CompilerHost {
	defaultLibraryPath := bundled.LibPath()
	var extendedConfigCache collections.SyncMap[tspath.Path, *tsoptions.ExtendedConfigCacheEntry]
	return compiler.NewCompilerHost(cwd, fs, defaultLibraryPath, &extendedConfigCache)
}

func CreateProgram(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost) (*compiler.Program, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !fs.FileExists(resolvedConfigPath) {
		return nil, fmt.Errorf("couldn't read tsconfig at %v", resolvedConfigPath)
	}

	configParseResult, _ := tsoptions.GetParsedCommandLineOfConfigFile(tsconfigPath, &core.CompilerOptions{}, host, nil)

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

	diagnostics := program.GetSyntacticDiagnostics(context.Background(), nil)
	if len(diagnostics) != 0 {
		return nil, fmt.Errorf("found %v syntactic errors. Try running \"tsgo --noEmit\" first\n", len(diagnostics))
	}

	program.BindSourceFiles()

	// program.CreateCheckers()

	return program, nil
}
