package triple_slash_reference

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type TripleSlashReferenceOptions struct {
	Lib   string `json:"lib"`   // "always" | "never"
	Path  string `json:"path"`  // "always" | "never" | "prefer-import"
	Types string `json:"types"` // "always" | "never" | "prefer-import"
}

var tripleSlashRegex = regexp.MustCompile(`^///\s*<reference\s+(path|types|lib)\s*=`)

// TripleSlashReferenceRule implements the triple-slash-reference rule
// Disallow certain triple slash directives
var TripleSlashReferenceRule = rule.CreateRule(rule.Rule{
	Name: "triple-slash-reference",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := TripleSlashReferenceOptions{
		Lib:   "always",
		Path:  "always",
		Types: "prefer-import",
	}

	// Parse options
	if options != nil {
		var optsMap map[string]interface{}
		var ok bool

		// Handle array format: [{ option: value }]
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			optsMap, ok = optArray[0].(map[string]interface{})
		} else {
			// Handle direct object format: { option: value }
			optsMap, ok = options.(map[string]interface{})
		}

		if ok {
			if lib, ok := optsMap["lib"].(string); ok {
				opts.Lib = lib
			}
			if path, ok := optsMap["path"].(string); ok {
				opts.Path = path
			}
			if types, ok := optsMap["types"].(string); ok {
				opts.Types = types
			}
		}
	}

	// Get the full text of the source file
	text := ctx.SourceFile.Text()

	// Split into lines to check for triple-slash references
	lines := strings.Split(text, "\n")

	// Check if file has imports
	hasImport := hasImportStatements(ctx.SourceFile)

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is a triple-slash reference
		if !tripleSlashRegex.MatchString(trimmed) {
			continue
		}

		// Determine the type of reference
		var refType string
		if strings.Contains(trimmed, `path=`) {
			refType = "path"
		} else if strings.Contains(trimmed, `types=`) {
			refType = "types"
		} else if strings.Contains(trimmed, `lib=`) {
			refType = "lib"
		}

		// Check if this reference should be reported
		shouldReport := false
		switch refType {
		case "path":
			shouldReport = opts.Path == "never"
		case "types":
			shouldReport = opts.Types == "never" || (opts.Types == "prefer-import" && hasImport)
		case "lib":
			shouldReport = opts.Lib == "never"
		}

		if shouldReport {
			// Calculate position (approximate - this is simplified)
			pos := 0
			for i := range lineNum {
				pos += len(lines[i]) + 1 // +1 for newline
			}

			ctx.ReportRange(
				core.NewTextRange(pos, pos+len(line)),
				rule.RuleMessage{
					Id:          "tripleSlashReference",
					Description: "Do not use a triple slash reference for " + refType + ", use `import` style instead.",
				},
			)
		}
	}

	return rule.RuleListeners{}
}

// hasImportStatements checks if the source file contains any import statements
func hasImportStatements(sourceFile *ast.SourceFile) bool {
	if sourceFile.Statements == nil {
		return false
	}

	for _, stmt := range sourceFile.Statements.Nodes {
		switch stmt.Kind {
		case ast.KindImportDeclaration, ast.KindImportEqualsDeclaration:
			return true
		}
	}
	return false
}
