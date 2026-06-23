package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/web-infra-dev/rslint/internal/config"
)

func main() {
	// Register all rules
	config.RegisterAllRules()

	// Get all rules from the registry
	rulesMap := config.GlobalRuleRegistry.GetAllRules()

	// Get sorted rule names
	var ruleNames []string
	for k := range rulesMap {
		ruleNames = append(ruleNames, k)
	}
	sort.Strings(ruleNames)

	var lines []string
	var typedCount, untypedCount int
	for _, name := range ruleNames {
		r := rulesMap[name]
		if r.RunWithOptions != nil {
			type0 := "never"
			if r.Schema0 != nil {
				type0 = r.Schema0.TSType()
			}

			type1 := "never"
			if r.Schema1 != nil {
				type1 = r.Schema1.TSType()
			}

			var line string
			if type1 == "never" {
				if type0 == "never" {
					line = fmt.Sprintf("    %q?: RuleEntry;", name)
				} else {
					line = fmt.Sprintf("    %q?: RuleEntry<%s>;", name, type0)
				}
			} else {
				line = fmt.Sprintf("    %q?: RuleEntry<%s, %s>;", name, type0, type1)
			}
			lines = append(lines, line)
			typedCount++
		} else {
			untypedCount++
		}
	}

	generatedTypes := strings.Join(lines, "\n")

	dtsPath := "./packages/rslint/dist/index.d.ts"
	if _, err := os.Stat(dtsPath); os.IsNotExist(err) {
		dtsPath = "./dist/index.d.ts"
	}

	contentBytes, err := os.ReadFile(dtsPath)
	if err != nil {
		log.Fatalf("failed to read %s: %v", dtsPath, err)
	}
	content := string(contentBytes)

	magicComment := "/** @__RULE_OPTIONS__ */"
	if !strings.Contains(content, magicComment) {
		log.Fatalf("magic comment %q not found in %s. Please run `pnpm --filter @rslint/core run build:js` first to compile the typescript files.", magicComment, dtsPath)
	}

	// Replace the entire line containing the magic comment to avoid extra indentation from the comment line itself.
	fileLines := strings.Split(content, "\n")
	found := false
	for i, line := range fileLines {
		if strings.Contains(line, magicComment) {
			fileLines[i] = generatedTypes
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("failed to find line containing magic comment in split lines")
	}

	newContent := strings.Join(fileLines, "\n")

	err = os.WriteFile(dtsPath, []byte(newContent), 0644)
	if err != nil {
		log.Fatalf("failed to write to %s: %v", dtsPath, err)
	}

	fmt.Printf("patched %d rule types, while %d rules are not.\n", typedCount, untypedCount)
}
