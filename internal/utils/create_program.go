package utils

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

func CreateCompilerHost(cwd string, fs vfs.FS) compiler.CompilerHost {
	defaultLibraryPath := bundled.LibPath()
	compilerOptions := core.CompilerOptions{}
	return compiler.NewCompilerHost(&compilerOptions, cwd, fs, defaultLibraryPath)
}

func CreateProgram(singleThreaded bool, fs vfs.FS, cwd string, tsconfigPath string, host compiler.CompilerHost) (*compiler.Program, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !fs.FileExists(resolvedConfigPath) {
		return nil, fmt.Errorf("couldn't read tsconfig at %v", resolvedConfigPath)
	}

	program := compiler.NewProgram(compiler.ProgramOptions{
		ConfigFileName: resolvedConfigPath,
		SingleThreaded: singleThreaded,
		Host:           host,
	})
	if program == nil {
		return nil, fmt.Errorf("couldn't create program")
	}

	diagnostics := program.GetSyntacticDiagnostics(nil)
	if len(diagnostics) != 0 {
		return nil, fmt.Errorf("found %v syntactic errors. Try running \"tsgo --noEmit\" first\n", len(diagnostics))
	}

	program.BindSourceFiles()

	// program.CreateCheckers()

	return program, nil
}
