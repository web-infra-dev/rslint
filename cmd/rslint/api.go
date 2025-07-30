package main

import (
	"encoding/json"
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
	rslintconfig "github.com/typescript-eslint/rslint/internal/config"
	"github.com/typescript-eslint/rslint/internal/linter"
	"github.com/typescript-eslint/rslint/internal/rule"
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

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllTypeSriptEslintPluginRules()

	// Load rslint configuration and determine which tsconfig files to use
	_, tsConfigs, configDirectory := rslintconfig.LoadConfigurationWithFallback(req.Config, currentDirectory, fs)

	// Get rules from request.RuleOptions
	var rules []rule.Rule
	if len(req.RuleOptions) > 0 {
		// Only use the rules specified in the request
		for ruleName := range req.RuleOptions {
			if r, exists := rslintconfig.GlobalRuleRegistry.GetRule(ruleName); exists {
				rules = append(rules, r)
			}
		}
	} else {
		// If no specific rules requested, don't run any rules (IPC mode should be explicit)
		rules = []rule.Rule{}
	}

	// Create compiler host
	host := utils.CreateCompilerHost(configDirectory, fs)
	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          host.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
	}

	// Create programs from all tsconfig files found in rslint config
	programs := []*compiler.Program{}
	for _, configFileName := range tsConfigs {
		program, err := utils.CreateProgram(false, fs, configDirectory, configFileName, host)
		if err != nil {
			return nil, fmt.Errorf("error creating TS program for %s: %v", configFileName, err)
		}
		programs = append(programs, program)
	}

	// Find source files from all programs
	files := []*ast.SourceFile{}

	// If specific files are provided, use those
	if len(req.Files) > 0 {
		for _, filePath := range req.Files {
			absPath := tspath.ResolvePath(configDirectory, filePath)
			// Try to find the file in any of the programs
			for _, program := range programs {
				sourceFile := program.GetSourceFile(absPath)
				if sourceFile != nil {
					files = append(files, sourceFile)
					break // Found in this program, no need to check others
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

	// Run linter
	err = linter.RunLinter(
		programs,
		false, // Don't use single-threaded mode for IPC
		files,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return utils.Map(rules, func(r rule.Rule) linter.ConfiguredRule {
				// Get rule options from request if available
				var ruleOptions interface{}
				if req.RuleOptions != nil {
					if opt, ok := req.RuleOptions[r.Name]; ok {
						// Handle case where options are sent as JSON string
						if optStr, isString := opt.(string); isString {
							var parsed interface{}
							if err := json.Unmarshal([]byte(optStr), &parsed); err == nil {
								// Successfully parsed JSON string
								// If it's an array like ["error", {...}], extract the options object
								if arr, isArray := parsed.([]interface{}); isArray && len(arr) > 1 {
									ruleOptions = arr[1]
								} else {
									ruleOptions = parsed
								}
							} else {
								// If not valid JSON, use the original value
								ruleOptions = opt
							}
						} else {
							ruleOptions = opt
						}
					}
				}
				
				return linter.ConfiguredRule{
					Name: r.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.Run(ctx, ruleOptions)
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
		RuleCount:   len(rules),
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
