package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IPCHandler implements the ipc.Handler interface
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
		return nil, fmt.Errorf("error getting current directory: %w", err)
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)

	// Create filesystem
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	allowedFiles := []string{}
	// Apply file contents if provided
	if len(req.FileContents) > 0 {
		fileContents := make(map[string]string, len(req.FileContents))
		for k, v := range req.FileContents {
			normalizePath := tspath.NormalizePath(k)
			fileContents[normalizePath] = v
			allowedFiles = append(allowedFiles, normalizePath)
		}
		fs = utils.NewOverlayVFS(fs, fileContents)

	}

	// Initialize rule registry with all available rules
	rslintconfig.RegisterAllRules()

	// Load rslint configuration and determine which tsconfig files to use
	_, tsConfigs, configDirectory := rslintconfig.LoadConfigurationWithFallback(req.Config, currentDirectory, fs)

	type RuleWithOption struct {
		rule   rule.Rule
		option interface{}
	}
	rulesWithOptions := []RuleWithOption{}
	// filter rule based on request.RuleOptions
	if len(req.RuleOptions) > 0 {
		for _, r := range rslintconfig.GlobalRuleRegistry.GetAllRules() {
			if option, ok := req.RuleOptions[r.Name]; ok {
				rulesWithOptions = append(rulesWithOptions, RuleWithOption{
					rule:   r,
					option: option,
				})
			}
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
			return nil, fmt.Errorf("error creating TS program for %s: %w", configFileName, err)
		}
		programs = append(programs, program)
	}

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
			MessageId: d.Message.Id,
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

		// Add fixes if available
		if d.FixesPtr != nil && len(*d.FixesPtr) > 0 {
			var fixes []api.Fix
			for _, fix := range *d.FixesPtr {
				// Convert TextRange to character positions
				startPos := fix.Range.Pos()
				endPos := fix.Range.End()

				fixes = append(fixes, api.Fix{
					Text:     fix.Text,
					StartPos: startPos,
					EndPos:   endPos,
				})
			}
			diagnostic.Fixes = fixes
		}

		diagnostics = append(diagnostics, diagnostic)
		errorsCount++
	}

	// Run linter
	lintedFilesCount, err := linter.RunLinter(
		programs,
		false, // Don't use single-threaded mode for IPC
		allowedFiles,
		utils.ExcludePaths,
		func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			return utils.Map(rulesWithOptions, func(r RuleWithOption) linter.ConfiguredRule {

				return linter.ConfiguredRule{
					Name: r.rule.Name,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return r.rule.Run(ctx, r.option)
					},
				}
			})
		},
		diagnosticCollector,
	)
	if err != nil {
		return nil, fmt.Errorf("error running linter: %w", err)
	}

	if diagnostics == nil {
		diagnostics = []api.Diagnostic{}
	}
	// Create response
	return &api.LintResponse{
		Diagnostics: diagnostics,
		ErrorCount:  errorsCount,
		FileCount:   int(lintedFilesCount),
		RuleCount:   len(rulesWithOptions),
	}, nil
}

// HandleApplyFixes handles apply fixes requests in IPC mode
func (h *IPCHandler) HandleApplyFixes(req api.ApplyFixesRequest) (*api.ApplyFixesResponse, error) {
	// Convert API diagnostics to rule diagnostics for use with linter.ApplyRuleFixes
	var ruleDiagnostics []rule.RuleDiagnostic

	for _, clientDiag := range req.Diagnostics {
		if len(clientDiag.Fixes) == 0 {
			continue
		}

		// Convert API fixes to rule fixes
		var ruleFixes []rule.RuleFix
		for _, clientFix := range clientDiag.Fixes {
			// Create TextRange from start and end positions
			textRange := core.NewTextRange(clientFix.StartPos, clientFix.EndPos)

			ruleFix := rule.RuleFix{
				Text:  clientFix.Text,
				Range: textRange,
			}
			ruleFixes = append(ruleFixes, ruleFix)
		}

		// Create rule diagnostic
		ruleDiag := rule.RuleDiagnostic{
			Range:    core.NewTextRange(0, 0), // Not used by ApplyRuleFixes
			RuleName: clientDiag.RuleName,
			Message: rule.RuleMessage{
				Id:          clientDiag.MessageId,
				Description: clientDiag.Message,
			},
			FixesPtr: &ruleFixes,
		}

		ruleDiagnostics = append(ruleDiagnostics, ruleDiag)
	}

	// Use linter.ApplyRuleFixes to apply the fixes
	code := req.FileContent
	outputs := []string{}
	wasFixed := false
	
	// Apply fixes iteratively to handle overlapping fixes
	for {
		fixedContent,unapplied, fixed := linter.ApplyRuleFixes(code, ruleDiagnostics)
		if !fixed {
			break
		}
		
		outputs = append(outputs, fixedContent)
		code = fixedContent
		wasFixed = true
		
		// Update diagnostics to only include unapplied ones for next iteration
		ruleDiagnostics = unapplied
		if len(ruleDiagnostics) == 0 {
			break
		}
	}

	// Count applied and unapplied fixes
	appliedCount := len(req.Diagnostics) - len(ruleDiagnostics)
	unappliedCount := len(ruleDiagnostics)

	return &api.ApplyFixesResponse{
		FixedContent:   outputs,
		WasFixed:       wasFixed,
		AppliedCount:   appliedCount,
		UnappliedCount: unappliedCount,
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
