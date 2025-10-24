package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// RuleRegistration represents a rule that needs to be registered
type RuleRegistration struct {
	Name         string // Full rule name (e.g., "@typescript-eslint/no-explicit-any")
	PackageName  string // Go package name (e.g., "no_explicit_any")
	VariableName string // Go variable name (e.g., "NoExplicitAnyRule")
	ImportPath   string // Full import path
	Plugin       string // Plugin type
}

var (
	ruleName    = flag.String("rule", "", "Rule name to register (e.g., no-explicit-any)")
	plugin      = flag.String("plugin", "typescript-eslint", "Plugin name: typescript-eslint, import, or empty for core rules")
	configPath  = flag.String("config", "internal/config/config.go", "Path to config.go file")
	dryRun      = flag.Bool("dry-run", false, "Preview changes without modifying files")
	autoDetect  = flag.Bool("auto", false, "Auto-detect all unregistered rules and register them")
)

func main() {
	flag.Parse()

	if *autoDetect {
		fmt.Println("Auto-detecting unregistered rules...")
		if err := autoRegisterRules(); err != nil {
			log.Fatalf("Error auto-registering rules: %v", err)
		}
		return
	}

	if *ruleName == "" {
		log.Fatal("Error: -rule flag is required (or use -auto)\nUsage: go run scripts/register-rule.go -rule <rule-name> [options]")
	}

	reg := buildRegistration(*ruleName, *plugin)

	if *dryRun {
		fmt.Println("=== DRY RUN MODE ===")
		fmt.Printf("Would register rule: %s\n", reg.Name)
		fmt.Printf("Import: %s\n", reg.ImportPath)
		fmt.Printf("Registration: GlobalRuleRegistry.Register(\"%s\", %s.%s)\n",
			reg.Name, reg.PackageName, reg.VariableName)
		return
	}

	if err := registerRule(reg); err != nil {
		log.Fatalf("Error registering rule: %v", err)
	}

	fmt.Printf("✓ Successfully registered rule: %s\n", reg.Name)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'go build ./...' to verify the code compiles")
	fmt.Println("  2. Run tests to ensure the rule works correctly")
}

func buildRegistration(name, pluginName string) RuleRegistration {
	reg := RuleRegistration{
		PackageName:  toSnakeCase(name),
		VariableName: toPascalCase(name) + "Rule",
		Plugin:       pluginName,
	}

	switch pluginName {
	case "typescript-eslint":
		reg.Name = "@typescript-eslint/" + name
		reg.ImportPath = fmt.Sprintf("github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/%s", reg.PackageName)
	case "import":
		reg.Name = "import/" + name
		reg.ImportPath = fmt.Sprintf("github.com/web-infra-dev/rslint/internal/plugins/import/rules/%s", reg.PackageName)
	default:
		reg.Name = name
		reg.ImportPath = fmt.Sprintf("github.com/web-infra-dev/rslint/internal/rules/%s", reg.PackageName)
	}

	return reg
}

func registerRule(reg RuleRegistration) error {
	content, err := os.ReadFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Add import
	newContent := addImport(string(content), reg)

	// Add registration
	newContent = addRegistration(newContent, reg)

	// Format the code
	formatted, err := format.Source([]byte(newContent))
	if err != nil {
		return fmt.Errorf("failed to format code: %w", err)
	}

	// Write back
	if err := os.WriteFile(*configPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func addImport(content string, reg RuleRegistration) string {
	importSection := extractImportSection(content)
	if importSection == "" {
		return content
	}

	// Check if import already exists
	importLine := fmt.Sprintf("\t\"%s\"", reg.ImportPath)
	if strings.Contains(importSection, importLine) {
		fmt.Println("Import already exists, skipping...")
		return content
	}

	// Find where to insert the new import
	// We'll insert it in alphabetical order within the appropriate section
	lines := strings.Split(importSection, "\n")
	var newLines []string
	inserted := false

	for i, line := range lines {
		if !inserted && strings.Contains(line, "github.com/web-infra-dev/rslint/internal") {
			// Find the right position alphabetically
			if shouldInsertBefore(line, reg.ImportPath) {
				newLines = append(newLines, importLine)
				inserted = true
			}
		}
		newLines = append(newLines, line)

		// If we reach the end of internal imports and haven't inserted yet
		if !inserted && i < len(lines)-1 &&
		   strings.Contains(line, "github.com/web-infra-dev/rslint/internal") &&
		   !strings.Contains(lines[i+1], "github.com/web-infra-dev/rslint/internal") {
			newLines = append(newLines, importLine)
			inserted = true
		}
	}

	if !inserted {
		// Insert before the last closing paren
		for i := len(newLines) - 1; i >= 0; i-- {
			if strings.TrimSpace(newLines[i]) == ")" {
				newLines = append(newLines[:i], append([]string{importLine}, newLines[i:]...)...)
				break
			}
		}
	}

	newImportSection := strings.Join(newLines, "\n")
	return strings.Replace(content, importSection, newImportSection, 1)
}

func addRegistration(content string, reg RuleRegistration) string {
	// Find the appropriate registration function
	var funcName string
	switch reg.Plugin {
	case "typescript-eslint":
		funcName = "registerAllTypeScriptEslintPluginRules"
	case "import":
		funcName = "registerAllEslintImportPluginRules"
	default:
		// For core rules, add to RegisterAllRules function
		funcName = "RegisterAllRules"
	}

	// Find the function
	funcRegex := regexp.MustCompile(fmt.Sprintf(`func %s\(\) \{([^}]+)\}`, funcName))
	matches := funcRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		fmt.Printf("Warning: Could not find function %s\n", funcName)
		return content
	}

	funcBody := matches[1]
	registrationLine := fmt.Sprintf("\tGlobalRuleRegistry.Register(\"%s\", %s.%s)",
		reg.Name, reg.PackageName, reg.VariableName)

	// Check if already registered
	if strings.Contains(funcBody, registrationLine) {
		fmt.Println("Rule already registered, skipping...")
		return content
	}

	// Add registration in alphabetical order
	lines := strings.Split(funcBody, "\n")
	var newLines []string
	inserted := false

	for _, line := range lines {
		if !inserted && strings.Contains(line, "GlobalRuleRegistry.Register") {
			if shouldInsertRegistrationBefore(line, reg.Name) {
				newLines = append(newLines, registrationLine)
				inserted = true
			}
		}
		newLines = append(newLines, line)
	}

	if !inserted {
		// Insert before the closing brace
		for i := len(newLines) - 1; i >= 0; i-- {
			if strings.TrimSpace(newLines[i]) == "" && i > 0 {
				newLines = append(newLines[:i], append([]string{registrationLine}, newLines[i:]...)...)
				inserted = true
				break
			}
		}
	}

	if !inserted {
		newLines = append([]string{registrationLine}, newLines...)
	}

	newFuncBody := strings.Join(newLines, "\n")
	return strings.Replace(content, funcBody, newFuncBody, 1)
}

func extractImportSection(content string) string {
	importRegex := regexp.MustCompile(`import \([^)]+\)`)
	match := importRegex.FindString(content)
	return match
}

func shouldInsertBefore(existingLine, newImportPath string) bool {
	// Extract import path from existing line
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindStringSubmatch(existingLine)
	if len(matches) < 2 {
		return false
	}
	existingPath := matches[1]

	// Extract package name for comparison
	existingPkg := filepath.Base(existingPath)
	newPkg := filepath.Base(newImportPath)

	return newPkg < existingPkg
}

func shouldInsertRegistrationBefore(existingLine, newRuleName string) bool {
	// Extract rule name from existing registration line
	re := regexp.MustCompile(`Register\("([^"]+)"`)
	matches := re.FindStringSubmatch(existingLine)
	if len(matches) < 2 {
		return false
	}
	existingName := matches[1]

	return newRuleName < existingName
}

func autoRegisterRules() error {
	// Find all rule directories
	plugins := []struct {
		name string
		path string
	}{
		{"typescript-eslint", "internal/plugins/typescript/rules"},
		{"import", "internal/plugins/import/rules"},
		{"", "internal/rules"},
	}

	var allRegistrations []RuleRegistration

	for _, p := range plugins {
		ruleDirs, err := findRuleDirectories(p.path)
		if err != nil {
			fmt.Printf("Warning: Could not scan %s: %v\n", p.path, err)
			continue
		}

		for _, ruleDir := range ruleDirs {
			ruleName := filepath.Base(ruleDir)
			// Convert snake_case back to kebab-case
			ruleName = strings.ReplaceAll(ruleName, "_", "-")

			reg := buildRegistration(ruleName, p.name)
			allRegistrations = append(allRegistrations, reg)
		}
	}

	if len(allRegistrations) == 0 {
		fmt.Println("No rules found to register")
		return nil
	}

	fmt.Printf("Found %d rules to potentially register\n", len(allRegistrations))

	// Read current config
	content, err := os.ReadFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	registered := 0
	for _, reg := range allRegistrations {
		// Check if already registered
		if strings.Contains(string(content), reg.VariableName) {
			continue
		}

		fmt.Printf("Registering: %s\n", reg.Name)
		if err := registerRule(reg); err != nil {
			fmt.Printf("  Warning: Failed to register %s: %v\n", reg.Name, err)
			continue
		}
		registered++

		// Re-read content for next iteration
		content, _ = os.ReadFile(*configPath)
	}

	fmt.Printf("\n✓ Registered %d new rules\n", registered)
	return nil
}

func findRuleDirectories(basePath string) ([]string, error) {
	var dirs []string

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(basePath, entry.Name())
			// Check if it contains a .go file (rule implementation)
			files, err := os.ReadDir(fullPath)
			if err != nil {
				continue
			}
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".go") && !strings.HasSuffix(f.Name(), "_test.go") {
					dirs = append(dirs, fullPath)
					break
				}
			}
		}
	}

	sort.Strings(dirs)
	return dirs, nil
}

func toSnakeCase(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "-")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
