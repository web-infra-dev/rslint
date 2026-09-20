package tsconfig

import (
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/program"
)

type nearestConfigKey string
type compilerOptionsKey string

// Read parses an explicit config through this generation's FS.
// Returned options are cached and must be treated as immutable.
func Read(p *program.Program, fileName string) *core.CompilerOptions {
	if p.FS() == nil {
		return nil
	}
	fileName = tspath.ResolvePath(p.CurrentDirectory(), fileName)
	return program.Cached(p, compilerOptionsKey(fileName), func() *core.CompilerOptions {
		host := compiler.NewCompilerHost(p.CurrentDirectory(), p.FS(), p.DefaultLibraryPath(), nil, nil, nil)
		parsed, _ := tsoptions.GetParsedCommandLineOfConfigFile(fileName, &core.CompilerOptions{}, nil, host, nil)
		if parsed != nil {
			return parsed.CompilerOptions()
		}
		return nil
	})
}

// FindNearest reads the nearest tsconfig using the same immutable
// FS and tsgo config parser, including extends. It does not create a Program.
func FindNearest(p *program.Program, fileName string) *core.CompilerOptions {
	if p.FS() == nil {
		return nil
	}
	directory := tspath.GetDirectoryPath(tspath.ResolvePath(p.CurrentDirectory(), fileName))
	return program.Cached(p, nearestConfigKey(directory), func() *core.CompilerOptions {
		config, found := tspath.ForEachAncestorDirectory(directory, func(directory string) (string, bool) {
			config := tspath.ResolvePath(directory, "tsconfig.json")
			return config, p.FS().FileExists(config)
		})
		if found {
			return Read(p, config)
		}
		return nil
	})
}
