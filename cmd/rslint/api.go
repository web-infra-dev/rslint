package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	ipc "github.com/typescript-eslint/rslint/internal/api"
	"github.com/typescript-eslint/rslint/internal/linter"
	"github.com/typescript-eslint/rslint/internal/rule"
	rslintconfig "github.com/typescript-eslint/rslint/internal/config"

	"github.com/typescript-eslint/rslint/internal/utils"
)

// IPCHandler implements the ipc.Handler interface
type IPCHandler struct{}

// HandleLint handles lint requests in IPC mode
func (h *IPCHandler) HandleLint(req ipc.LintRequest) (*ipc.LintResponse, error) {
	// Format is not used for IPC mode as we return structured data
	_ = req.Format

	// Set working directory if provided
	if req.WorkingDirectory != "" {
		if err := os.Chdir(req.WorkingDirectory); err != nil {
			return nil, fmt.Errorf("failed to change directory: %w", err)
		}
	}

	// Get current directory
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)

	// Create filesystem
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	// Apply file contents if provided
	if len(req.FileContents) > 0 {
		fs = utils.NewOverlayVFS(fs, req.FileContents)
	}

	// Load configuration using same logic as CMD mode
	loader := rslintconfig.NewConfigLoader(fs, currentDirectory)
	
	var rslintConfig rslintconfig.RslintConfig
	var tsConfigs []string
	var configDirectory string
	
	if req.Config != "" {
		rslintConfig, configDirectory, err = loader.LoadRslintConfig(req.Config)
		if err != nil {
			return nil, fmt.Errorf("error loading rslint config: %v", err)
		}
	} else {
		rslintConfig, configDirectory, err = loader.LoadDefaultRslintConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading default rslint config: %v", err)
		}
	}

	tsConfigs, err = loader.LoadTsConfigsFromRslintConfig(rslintConfig, configDirectory)
	if err != nil {
		return nil, fmt.Errorf("error loading tsconfig from rslint config: %v", err)
	}

	// Update current directory to config directory
	currentDirectory = configDirectory

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllTypeSriptEslintPluginRules()

	// Create compiler host
	host := utils.CreateCompilerHost(currentDirectory, fs)
	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}

	// Create programs from all tsconfig files
	programs := []*compiler.Program{}
	for _, configFileName := range tsConfigs {
		program, err := utils.CreateProgram(false, fs, currentDirectory, configFileName, host)
		if err != nil {
			return nil, fmt.Errorf("error creating TS program for %s: %v", configFileName, err)
		}
		programs = append(programs, program)
	}

	// Find source files
	files := []*ast.SourceFile{}

	// If specific files are provided, use those
	if len(req.Files) > 0 {
		for _, filePath := range req.Files {
			absPath := tspath.ResolvePath(currentDirectory, filePath)
			for _, program := range programs {
				sourceFile := program.GetSourceFile(absPath)
				if sourceFile != nil {
					files = append(files, sourceFile)
					break // Found the file in one of the programs
				}
			}
		}
	} else {
		// Otherwise use all source files from all programs
		for _, program := range programs {
			for _, file := range program.SourceFiles() {
				p := string(file.Path())
				if strings.Contains(p, "/node_modules/") {
					continue
				}
				// skip bundled files
				if strings.Contains(p, "bundled:") {
					continue
				}
				files = append(files, file)
			}
		}
	}
	slices.SortFunc(files, func(a *ast.SourceFile, b *ast.SourceFile) int {
		return len(b.Text()) - len(a.Text())
	})

	// Collect diagnostics
	var diagnostics []ipc.Diagnostic
	var diagnosticsLock sync.Mutex
	errorsCount := 0

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()

		diagnosticStart := d.Range.Pos()
		diagnosticEnd := d.Range.End()

		startLine, startColumn := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticStart)
		endLine, endColumn := scanner.GetLineAndCharacterOfPosition(d.SourceFile, diagnosticEnd)

		diagnostic := ipc.Diagnostic{
			RuleName: d.RuleName,
			Message:  d.Message.Description,
			FilePath: tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions),
			Range: ipc.Range{
				Start: ipc.Position{
					Line:   startLine + 1, // Convert to 1-based indexing
					Column: startColumn + 1,
				},
				End: ipc.Position{
					Line:   endLine + 1,
					Column: endColumn + 1,
				},
			},
		}

		diagnostics = append(diagnostics, diagnostic)
		errorsCount++
	}

	// Run linter using the same logic as CMD mode
	err = linter.RunLinter(
		programs,
		false, // Don't use single-threaded mode for IPC
		files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			activeRules := rslintconfig.GlobalRuleRegistry.GetEnabledRules(rslintConfig, sourceFile.FileName())
			
			// Filter rules based on request.RuleOptions if provided
			if len(req.RuleOptions) > 0 {
				filteredRules := []rule.Rule{}
				for _, r := range activeRules {
					if _, ok := req.RuleOptions[r.Name]; ok {
						filteredRules = append(filteredRules, r)
					}
				}
				activeRules = filteredRules
			}
			
			return utils.Map(activeRules, func(r rule.Rule) linter.ConfiguredRule {
				return linter.ConfiguredRule{
					Name: r.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, nil)
					},
				}
			})
		},
		diagnosticCollector,
	)
	if err != nil {
		return nil, fmt.Errorf("error running linter: %v", err)
	}
	if diagnostics == nil {
		diagnostics = []ipc.Diagnostic{}
	}
	// Create response
	return &ipc.LintResponse{
		Diagnostics: diagnostics,
		ErrorCount:  errorsCount,
		FileCount:   len(files),
		RuleCount:   len(rslintconfig.GlobalRuleRegistry.GetEnabledRules(rslintConfig, "")),
	}, nil
}

// runAPI runs the linter in IPC mode
func runAPI() int {
	handler := &IPCHandler{}
	service := ipc.NewService(os.Stdin, os.Stdout, handler)

	if err := service.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error in IPC mode: %v\n", err)
		return 1
	}
	return 0
}
