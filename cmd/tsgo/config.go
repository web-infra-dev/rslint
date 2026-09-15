package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// FinalCompilerOptions contains resolved values for consumers checking compiler
// assumptions, including defaults and implied options.
type FinalCompilerOptions struct {
	Target                       string `json:"target"`
	Strict                       bool   `json:"strict"`
	NoImplicitAny                bool   `json:"noImplicitAny"`
	NoImplicitThis               bool   `json:"noImplicitThis"`
	StrictNullChecks             bool   `json:"strictNullChecks"`
	StrictFunctionTypes          bool   `json:"strictFunctionTypes"`
	StrictBindCallApply          bool   `json:"strictBindCallApply"`
	StrictPropertyInitialization bool   `json:"strictPropertyInitialization"`
	StrictBuiltinIteratorReturn  bool   `json:"strictBuiltinIteratorReturn"`
	UseUnknownInCatchVariables   bool   `json:"useUnknownInCatchVariables"`
	VerbatimModuleSyntax         bool   `json:"verbatimModuleSyntax"`
	IsolatedModules              bool   `json:"isolatedModules"`
	UseDefineForClassFields      bool   `json:"useDefineForClassFields"`
	NoUncheckedIndexedAccess     bool   `json:"noUncheckedIndexedAccess"`
}

func finalCompilerOptions(options *core.CompilerOptions) FinalCompilerOptions {
	return FinalCompilerOptions{
		Target:                       strings.ToLower(options.GetEmitScriptTarget().String()),
		Strict:                       options.GetStrictOptionValue(core.TSUnknown),
		NoImplicitAny:                options.GetStrictOptionValue(options.NoImplicitAny),
		NoImplicitThis:               options.GetStrictOptionValue(options.NoImplicitThis),
		StrictNullChecks:             options.GetStrictOptionValue(options.StrictNullChecks),
		StrictFunctionTypes:          options.GetStrictOptionValue(options.StrictFunctionTypes),
		StrictBindCallApply:          options.GetStrictOptionValue(options.StrictBindCallApply),
		StrictPropertyInitialization: options.GetStrictOptionValue(options.StrictPropertyInitialization),
		StrictBuiltinIteratorReturn:  options.GetStrictOptionValue(options.StrictBuiltinIteratorReturn),
		UseUnknownInCatchVariables:   options.GetStrictOptionValue(options.UseUnknownInCatchVariables),
		VerbatimModuleSyntax:         options.VerbatimModuleSyntax.IsTrue(),
		IsolatedModules:              options.GetIsolatedModules(),
		UseDefineForClassFields:      options.GetUseDefineForClassFields(),
		NoUncheckedIndexedAccess:     options.NoUncheckedIndexedAccess.IsTrue(),
	}
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("config", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var configPath string
	flags.StringVar(&configPath, "config", "tsconfig.json", "path to tsconfig.json, relative to the working directory")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: tsgo config [options]")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "unexpected positional arguments:", strings.Join(flags.Args(), " "))
		return 2
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	cwd = tspath.NormalizePath(cwd)
	configPath = tspath.ResolvePath(cwd, configPath)
	host := utils.CreateCompilerHost(cwd, cachedvfs.From(osvfs.FS()))
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(configPath, &core.CompilerOptions{}, nil, host, nil)
	if config != nil {
		diagnostics = append(diagnostics, config.GetConfigFileParsingDiagnostics()...)
	}
	if len(diagnostics) > 0 {
		for _, diagnostic := range diagnostics {
			path := configPath
			if diagnostic.File() != nil {
				path = diagnostic.File().FileName()
			}
			fmt.Fprintf(stderr, "%s: TS%d: %s\n", path, diagnostic.Code(), diagnostic.String())
		}
		return 1
	}
	encoder := json.NewEncoder(stdout)
	if err := encoder.Encode(finalCompilerOptions(config.CompilerOptions())); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
