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
	api "github.com/typescript-eslint/rslint/internal/api"
	rslintconfig "github.com/typescript-eslint/rslint/internal/config"
	"github.com/typescript-eslint/rslint/internal/linter"
	"github.com/typescript-eslint/rslint/internal/rule"
	"github.com/typescript-eslint/rslint/internal/utils"
)

// IPCHandler implements the api.Handler interface
type IPCHandler struct{}

// HandleLint handles lint requests in IPC mode
func (h *IPCHandler) HandleLint(req api.LintRequest) (*api.LintResponse, error) {
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

	// Create rules with their configurations from request.RuleOptions
	var rulesWithConfig []rslintconfig.EnabledRuleWithConfig
	if len(req.RuleOptions) > 0 {
		// Process each rule from the request
		for ruleName, ruleConfig := range req.RuleOptions {
			// Try to find the rule by name, handling both namespaced and non-namespaced variants
			var ruleImpl rule.Rule
			var exists bool
			
			// First try exact match
			if ruleImpl, exists = rslintconfig.GlobalRuleRegistry.GetRule(ruleName); !exists {
				// If not found, try common aliases
				aliases := []string{}
				if strings.HasPrefix(ruleName, "@typescript-eslint/") {
					// Try without the namespace prefix
					aliases = append(aliases, strings.TrimPrefix(ruleName, "@typescript-eslint/"))
				} else {
					// Try with the namespace prefix
					aliases = append(aliases, "@typescript-eslint/"+ruleName)
				}
				
				for _, alias := range aliases {
					if ruleImpl, exists = rslintconfig.GlobalRuleRegistry.GetRule(alias); exists {
						break
					}
				}
			}
			
			if exists {
				// Parse the rule config - can be just a string or an array with options
				var options map[string]interface{}
				var level string

				switch v := ruleConfig.(type) {
				case string:
					level = v
				case []interface{}:
					if len(v) > 0 {
						if levelStr, ok := v[0].(string); ok {
							level = levelStr
							if len(v) > 1 {
								if opts, ok := v[1].(map[string]interface{}); ok {
									options = opts
								} else {
									// For rules that expect a simple option value, pass it directly
									options = map[string]interface{}{"value": v[1]}
								}
							}
						}
					}
				default:
					// Handle JSON objects that might have been passed as interface{}
					if jsonBytes, err := json.Marshal(v); err == nil {
						var parsed map[string]interface{}
						if json.Unmarshal(jsonBytes, &parsed) == nil {
							if levelStr, ok := parsed["level"].(string); ok {
								level = levelStr
							}
							if opts, ok := parsed["options"].(map[string]interface{}); ok {
								options = opts
							}
						}
					}
				}

				if level != "off" && level != "" {
					rulesWithConfig = append(rulesWithConfig, rslintconfig.EnabledRuleWithConfig{
						Rule: ruleImpl,
						Config: &rslintconfig.RuleConfig{
							Level:   level,
							Options: options,
						},
					})
				}
			}
		}
	} else {
		// If no specific rules requested, use all available rules with default configuration
		allRules := rslintconfig.GlobalRuleRegistry.GetAllRules()
		for _, ruleImpl := range allRules {
			rulesWithConfig = append(rulesWithConfig, rslintconfig.EnabledRuleWithConfig{
				Rule: ruleImpl,
				Config: &rslintconfig.RuleConfig{
					Level: "error",
				},
			})
		}
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
	var diagnostics []api.Diagnostic
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

		diagnostic := api.Diagnostic{
			RuleName:  d.RuleName,
			MessageID: d.Message.Id,
			Message:   d.Message.Description,
			FilePath:  tspath.ConvertToRelativePath(d.SourceFile.FileName(), comparePathOptions),
			Range: api.Range{
				Start: api.Position{
					Line:   startLine + 1, // Convert to 1-based indexing
					Column: startColumn + 1,
				},
				End: api.Position{
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
			return utils.Map(rulesWithConfig, func(ruleWithConfig rslintconfig.EnabledRuleWithConfig) linter.ConfiguredRule {
				return linter.ConfiguredRule{
					Name:     ruleWithConfig.Rule.Name,
					Severity: ruleWithConfig.Config.GetSeverity(),
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						// Pass options directly as received - if it's wrapped in {"value": X}, unwrap it
						options := ruleWithConfig.Config.Options
						var finalOptions interface{} = options
						if len(options) == 1 {
							if val, hasValue := options["value"]; hasValue {
								// This was a simple option that got wrapped, unwrap it
								finalOptions = val
							}
						}
						return ruleWithConfig.Rule.Run(ctx, finalOptions)
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
		diagnostics = []api.Diagnostic{}
	}
	// Create response
	return &api.LintResponse{
		Diagnostics: diagnostics,
		ErrorCount:  errorsCount,
		FileCount:   len(files),
		RuleCount:   len(rulesWithConfig),
	}, nil
}

// runAPI runs the linter in IPC mode
func runAPI() int {
	handler := &IPCHandler{}
	service := api.NewService(os.Stdin, os.Stdout, handler)

	if err := service.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error in IPC mode: %v\n", err)
		return 1
	}
	return 0
}
