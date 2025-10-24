package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// RuleMetadata holds information about a rule to be generated
type RuleMetadata struct {
	Name              string            // Rule name in kebab-case (e.g., "no-explicit-any")
	PackageName       string            // Go package name in snake_case (e.g., "no_explicit_any")
	RuleName          string            // Go variable name (e.g., "NoExplicitAnyRule")
	StructName        string            // Options struct name (e.g., "NoExplicitAnyOptions")
	Plugin            string            // Plugin name (e.g., "typescript-eslint", "import", "")
	PluginPrefix      string            // Plugin prefix for registration (e.g., "@typescript-eslint/", "import/", "")
	FullRuleName      string            // Full rule name with plugin prefix
	Description       string            // Brief description of the rule
	Category          string            // Rule category (e.g., "Best Practices", "Possible Errors")
	HasOptions        bool              // Whether the rule has configuration options
	Options           map[string]string // Option names and types
	RequiresTypeInfo  bool              // Whether the rule requires TypeScript type information
	HasAutofix        bool              // Whether the rule provides automatic fixes
	TargetASTNodes    []string          // AST node kinds this rule listens to
	MessageIDs        []string          // Error message IDs
}

// ESLintRuleSchema represents the structure of an ESLint rule's metadata
type ESLintRuleSchema struct {
	Meta struct {
		Type     string            `json:"type"`
		Docs     struct {
			Description      string `json:"description"`
			Category         string `json:"category"`
			RequiresTypeInfo bool   `json:"requiresTypeChecking"`
		} `json:"docs"`
		Schema   json.RawMessage   `json:"schema"`
		Messages map[string]string `json:"messages"`
		Fixable  string            `json:"fixable,omitempty"`
	} `json:"meta"`
}

var (
	ruleName      = flag.String("rule", "", "Rule name in kebab-case (e.g., no-explicit-any)")
	plugin        = flag.String("plugin", "typescript-eslint", "Plugin name: typescript-eslint, import, or empty for core rules")
	description   = flag.String("description", "", "Brief description of the rule")
	astNodes      = flag.String("ast-nodes", "", "Comma-separated list of AST node types (e.g., InterfaceDeclaration,ClassDeclaration)")
	requiresTypes = flag.Bool("requires-types", false, "Whether the rule requires type information")
	hasOptions    = flag.Bool("has-options", false, "Whether the rule has configuration options")
	hasAutofix    = flag.Bool("has-autofix", false, "Whether the rule provides automatic fixes")
	batchFile     = flag.String("batch", "", "Path to a file containing rule names (one per line) for batch generation")
	dryRun        = flag.Bool("dry-run", false, "Preview what would be generated without creating files")
	fetchMetadata = flag.Bool("fetch", false, "Attempt to fetch rule metadata from ESLint/TypeScript-ESLint repositories")
	outputDir     = flag.String("output", "", "Output directory (defaults to appropriate plugin directory)")
)

func main() {
	flag.Parse()

	if *batchFile != "" {
		processBatch(*batchFile)
		return
	}

	if *ruleName == "" {
		log.Fatal("Error: -rule flag is required\nUsage: go run scripts/generate-rule.go -rule <rule-name> [options]")
	}

	metadata := buildMetadata(*ruleName, *plugin, *description, *astNodes, *requiresTypes, *hasOptions, *hasAutofix, *fetchMetadata)

	if *dryRun {
		fmt.Println("=== DRY RUN MODE ===")
		fmt.Printf("Would generate rule: %s\n", metadata.FullRuleName)
		fmt.Printf("Package: %s\n", metadata.PackageName)
		fmt.Printf("Plugin: %s\n", metadata.Plugin)
		fmt.Printf("Output directory: %s\n", getRuleDirectory(metadata))
		fmt.Println("\nPreview of generated files:")
		previewGeneration(metadata)
		return
	}

	if err := generateRule(metadata); err != nil {
		log.Fatalf("Error generating rule: %v", err)
	}

	fmt.Printf("✓ Successfully generated rule: %s\n", metadata.FullRuleName)
	fmt.Printf("  Rule file: %s/%s.go\n", getRuleDirectory(metadata), metadata.PackageName)
	fmt.Printf("  Test file: %s/%s_test.go\n", getRuleDirectory(metadata), metadata.PackageName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Implement the rule logic in the Run function")
	fmt.Println("  2. Add test cases to the test file")
	fmt.Println("  3. Register the rule in internal/config/config.go")
	fmt.Printf("     - Add import: \"%s/%s\"\n", getImportPath(metadata), metadata.PackageName)
	fmt.Printf("     - Add registration: GlobalRuleRegistry.Register(\"%s\", %s.%s)\n",
		metadata.FullRuleName, metadata.PackageName, metadata.RuleName)
}

func buildMetadata(name, pluginName, desc, astNodesStr string, reqTypes, opts, autofix, fetch bool) RuleMetadata {
	metadata := RuleMetadata{
		Name:             name,
		PackageName:      toSnakeCase(name),
		RuleName:         toPascalCase(name) + "Rule",
		StructName:       toPascalCase(name) + "Options",
		Plugin:           pluginName,
		Description:      desc,
		RequiresTypeInfo: reqTypes,
		HasOptions:       opts,
		HasAutofix:       autofix,
		Options:          make(map[string]string),
		MessageIDs:       []string{"default"},
	}

	// Set plugin prefix
	switch pluginName {
	case "typescript-eslint":
		metadata.PluginPrefix = "@typescript-eslint/"
	case "import":
		metadata.PluginPrefix = "import/"
	default:
		metadata.PluginPrefix = ""
	}
	metadata.FullRuleName = metadata.PluginPrefix + name

	// Parse AST nodes
	if astNodesStr != "" {
		metadata.TargetASTNodes = strings.Split(astNodesStr, ",")
		for i := range metadata.TargetASTNodes {
			metadata.TargetASTNodes[i] = strings.TrimSpace(metadata.TargetASTNodes[i])
		}
	} else {
		// Default to a placeholder
		metadata.TargetASTNodes = []string{"FunctionDeclaration"}
	}

	// Try to fetch metadata if requested
	if fetch {
		fetchedMetadata := fetchRuleMetadata(name, pluginName)
		if fetchedMetadata != nil {
			if metadata.Description == "" {
				metadata.Description = fetchedMetadata.Description
			}
			metadata.Category = fetchedMetadata.Category
			if len(fetchedMetadata.MessageIDs) > 0 {
				metadata.MessageIDs = fetchedMetadata.MessageIDs
			}
		}
	}

	// Set default description if still empty
	if metadata.Description == "" {
		metadata.Description = fmt.Sprintf("TODO: Add description for %s rule", name)
	}

	return metadata
}

func fetchRuleMetadata(name, pluginName string) *RuleMetadata {
	var url string

	switch pluginName {
	case "typescript-eslint":
		// Try to fetch from TypeScript-ESLint repository
		url = fmt.Sprintf("https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main/packages/eslint-plugin/src/rules/%s.ts", name)
	case "import":
		url = fmt.Sprintf("https://raw.githubusercontent.com/import-js/eslint-plugin-import/main/src/rules/%s.js", name)
	default:
		// Try ESLint core rules
		url = fmt.Sprintf("https://raw.githubusercontent.com/eslint/eslint/main/lib/rules/%s.js", name)
	}

	fmt.Printf("Fetching metadata from: %s\n", url)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Warning: Could not fetch metadata: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Warning: Rule not found at %s (status: %d)\n", url, resp.StatusCode)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Warning: Could not read response: %v\n", err)
		return nil
	}

	metadata := &RuleMetadata{}

	// Extract description from source code using regex
	descRegex := regexp.MustCompile(`description:\s*['"](.+?)['"]`)
	if matches := descRegex.FindSubmatch(body); len(matches) > 1 {
		metadata.Description = string(matches[1])
	}

	// Extract message IDs
	messagesRegex := regexp.MustCompile(`messages:\s*\{([^}]+)\}`)
	if matches := messagesRegex.FindSubmatch(body); len(matches) > 1 {
		msgIDRegex := regexp.MustCompile(`(\w+):`)
		msgIDs := msgIDRegex.FindAllSubmatch(matches[1], -1)
		for _, msgID := range msgIDs {
			if len(msgID) > 1 {
				metadata.MessageIDs = append(metadata.MessageIDs, string(msgID[1]))
			}
		}
	}

	return metadata
}

func generateRule(metadata RuleMetadata) error {
	ruleDir := getRuleDirectory(metadata)

	// Create rule directory
	if err := os.MkdirAll(ruleDir, 0755); err != nil {
		return fmt.Errorf("failed to create rule directory: %w", err)
	}

	// Generate rule implementation file
	ruleFile := filepath.Join(ruleDir, metadata.PackageName+".go")
	if err := generateFile(ruleFile, ruleTemplate, metadata); err != nil {
		return fmt.Errorf("failed to generate rule file: %w", err)
	}

	// Generate test file
	testFile := filepath.Join(ruleDir, metadata.PackageName+"_test.go")
	if err := generateFile(testFile, testTemplate, metadata); err != nil {
		return fmt.Errorf("failed to generate test file: %w", err)
	}

	return nil
}

func generateFile(filePath string, tmpl string, metadata RuleMetadata) error {
	t, err := template.New("file").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, metadata); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated Go code
	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		// If formatting fails, write the unformatted code and return a warning
		fmt.Printf("Warning: Could not format %s: %v\n", filePath, err)
		formatted = []byte(buf.String())
	}

	if err := os.WriteFile(filePath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func previewGeneration(metadata RuleMetadata) {
	fmt.Println("\n--- Rule Implementation (preview) ---")
	t, _ := template.New("preview").Parse(ruleTemplate)
	t.Execute(os.Stdout, metadata)

	fmt.Println("\n--- Test File (preview) ---")
	t2, _ := template.New("preview").Parse(testTemplate)
	t2.Execute(os.Stdout, metadata)
}

func processBatch(batchFilePath string) {
	content, err := os.ReadFile(batchFilePath)
	if err != nil {
		log.Fatalf("Error reading batch file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	successful := 0
	failed := 0

	fmt.Printf("Processing batch file: %s\n", batchFilePath)
	fmt.Printf("Found %d rule names\n\n", len(lines))

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fmt.Printf("[%d/%d] Generating rule: %s\n", i+1, len(lines), line)

		metadata := buildMetadata(line, *plugin, "", *astNodes, *requiresTypes, *hasOptions, *hasAutofix, *fetchMetadata)

		if *dryRun {
			fmt.Printf("  Would create: %s\n", getRuleDirectory(metadata))
			successful++
			continue
		}

		if err := generateRule(metadata); err != nil {
			fmt.Printf("  ✗ Error: %v\n", err)
			failed++
		} else {
			fmt.Printf("  ✓ Success\n")
			successful++
		}
	}

	fmt.Printf("\nBatch processing complete:\n")
	fmt.Printf("  Successful: %d\n", successful)
	fmt.Printf("  Failed: %d\n", failed)
}

func getRuleDirectory(metadata RuleMetadata) string {
	if *outputDir != "" {
		return filepath.Join(*outputDir, metadata.PackageName)
	}

	base := "internal"
	switch metadata.Plugin {
	case "typescript-eslint":
		return filepath.Join(base, "plugins", "typescript", "rules", metadata.PackageName)
	case "import":
		return filepath.Join(base, "plugins", "import", "rules", metadata.PackageName)
	default:
		return filepath.Join(base, "rules", metadata.PackageName)
	}
}

func getImportPath(metadata RuleMetadata) string {
	switch metadata.Plugin {
	case "typescript-eslint":
		return "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules"
	case "import":
		return "github.com/web-infra-dev/rslint/internal/plugins/import/rules"
	default:
		return "github.com/web-infra-dev/rslint/internal/rules"
	}
}

func toSnakeCase(s string) string {
	// Convert kebab-case to snake_case
	return strings.ReplaceAll(s, "-", "_")
}

func toPascalCase(s string) string {
	// Convert kebab-case to PascalCase
	parts := strings.Split(s, "-")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// Templates for code generation
const ruleTemplate = `package {{.PackageName}}

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
{{- if .HasAutofix}}
	"github.com/web-infra-dev/rslint/internal/utils"
{{- end}}
)

{{if .HasOptions -}}
// {{.StructName}} defines the configuration options for this rule
type {{.StructName}} struct {
	// TODO: Add option fields here
	// Example: AllowSomePattern bool ` + "`json:\"allowSomePattern\"`" + `
}

// parseOptions parses and validates the rule options
func parseOptions(options any) {{.StructName}} {
	opts := {{.StructName}}{
		// Set default values here
	}

	if options == nil {
		return opts
	}

	// Handle both array format [{ option: value }] and object format { option: value }
	var optsMap map[string]interface{}
	if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
		optsMap, _ = optArray[0].(map[string]interface{})
	} else {
		optsMap, _ = options.(map[string]interface{})
	}

	if optsMap != nil {
		// TODO: Parse option values from optsMap
		// Example:
		// if v, ok := optsMap["allowSomePattern"].(bool); ok {
		//     opts.AllowSomePattern = v
		// }
	}

	return opts
}
{{end}}

{{- if eq .Plugin "typescript-eslint"}}
// {{.RuleName}} implements the {{.Name}} rule
// {{.Description}}
var {{.RuleName}} = rule.CreateRule(rule.Rule{
	Name: "{{.Name}}",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
{{- else}}
// {{.RuleName}} implements the {{.Name}} rule
// {{.Description}}
var {{.RuleName}} = rule.Rule{
	Name: "{{.FullRuleName}}",
	Run:  run,
}

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
{{- end}}
{{- if .HasOptions}}
	opts := parseOptions(options)
	_ = opts // Use opts in your rule logic
{{- end}}

	return rule.RuleListeners{
{{- range .TargetASTNodes}}
		ast.Kind{{.}}: func(node *ast.Node) {
			// TODO: Implement rule logic for {{.}}
			{{- if $.RequiresTypeInfo}}
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}
			{{- end}}

			// Example: Check some condition and report
			// if violatesRule(node) {
			//     ctx.ReportNode(node, rule.RuleMessage{
			//         Id:          "{{index $.MessageIDs 0}}",
			//         Description: "TODO: Add error message",
			//     })
			// }

			{{- if $.HasAutofix}}
			// Example with autofix:
			// ctx.ReportNodeWithFixes(node, rule.RuleMessage{
			//     Id:          "{{index $.MessageIDs 0}}",
			//     Description: "TODO: Add error message",
			// }, rule.RuleFixReplace(ctx.SourceFile, node, "fixed code"))
			{{- end}}
		},
{{- end}}
	}
}
`

const testTemplate = `package {{.PackageName}}

import (
	"testing"

{{- if eq .Plugin "typescript-eslint"}}
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
{{- else if eq .Plugin "import"}}
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/fixtures"
{{- else}}
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
{{- end}}
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func Test{{.RuleName}}(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&{{.RuleName}},
		[]rule_tester.ValidTestCase{
			// TODO: Add valid test cases
			{Code: ` + "`" + `
// Add valid code example here
const x = 1;
` + "`" + `},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases
			{
				Code: ` + "`" + `
// Add invalid code example here
var x = 1;
` + "`" + `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "{{index .MessageIDs 0}}",
						Line:      2, // TODO: Update line number
						Column:    1, // TODO: Update column number
					},
				},
{{- if .HasAutofix}}
				Output: []string{` + "`" + `
// Add expected output after autofix
const x = 1;
` + "`" + `},
{{- end}}
			},
		},
	)
}

{{- if .HasOptions}}

func Test{{.RuleName}}WithOptions(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&{{.RuleName}},
		[]rule_tester.ValidTestCase{
			{
				Code: ` + "`" + `
// Add code that is valid with specific options
` + "`" + `,
				Options: map[string]interface{}{
					// TODO: Add option values
					// "optionName": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
{{- end}}
`
