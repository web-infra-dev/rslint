package triple_slash_reference

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// maskBlockComments replaces the contents of block comments (/* ... */)
// with spaces, preserving newlines and overall string length so that
// byte offsets remain aligned with the original source.
func maskBlockComments(s string) string {
	b := []byte(s)
	n := len(b)
	for i := 0; i+1 < n; i++ {
		if b[i] == '/' && b[i+1] == '*' {
			// inside block comment
			j := i + 2
			for j+1 < n {
				// preserve newlines for correct line/column mapping
				if b[j] != '\n' && b[j] != '\r' {
					b[j] = ' '
				}
				if b[j] == '*' && b[j+1] == '/' {
					// mask the opening and closing too
					b[i] = ' '
					b[i+1] = ' '
					b[j] = ' '
					b[j+1] = ' '
					i = j + 1
					break
				}
				j++
			}
		}
	}
	return string(b)
}

// Note: no per-file state needed beyond immediate reports

func buildMessage(module string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "tripleSlashReference",
		Description: "Do not use a triple slash reference for " + module + ", use `import` style instead.",
	}
}

// Options: { lib?: 'always' | 'never'; path?: 'always' | 'never'; types?: 'always' | 'never' | 'prefer-import' }
type TripleSlashReferenceOptions struct {
	Lib   string
	Path  string
	Types string
}

// normalizeOptions parses options from either array/object forms and applies defaults
func normalizeOptions(options any) TripleSlashReferenceOptions {
	// Defaults from upstream rule
	opts := TripleSlashReferenceOptions{
		Lib:   "always",
		Path:  "never",
		Types: "prefer-import",
	}

	if options == nil {
		return opts
	}

	// Support array format: [ { ... } ] and object format: { ... }
	var m map[string]interface{}
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			if mm, ok := arr[0].(map[string]interface{}); ok {
				m = mm
			}
		}
	} else if mm, ok := options.(map[string]interface{}); ok {
		m = mm
	}

	if m == nil {
		return opts
	}

	if v, ok := m["lib"].(string); ok && v != "" {
		opts.Lib = v
	}
	if v, ok := m["path"].(string); ok && v != "" {
		opts.Path = v
	}
	if v, ok := m["types"].(string); ok && v != "" {
		opts.Types = v
	}
	return opts
}

var TripleSlashReferenceRule = rule.CreateRule(rule.Rule{
	Name: "triple-slash-reference",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := normalizeOptions(options)

		// Parse header comments once per file
		parsed := false

		parseHeaderComments := func() {
			if parsed {
				return
			}
			parsed = true

			// If everything is allowed, do nothing
			if opts.Lib == "always" && opts.Path == "always" && opts.Types == "always" {
				return
			}

			sf := ctx.SourceFile
			if sf == nil {
				return
			}

			fullText := sf.Text()
			// Look only at leading comments before the first token (header area)
			start := len(scanner.GetShebang(fullText))
			// Match triple-slash reference directives within individual leading comments
			lineRe := regexp.MustCompile(`(?m)^[ \t]*///[ \t]*<reference[ \t]*(types|path|lib)[ \t]*=[ \t]*["']([^"']+)["']`)
			for comment := range scanner.GetLeadingCommentRanges(&ast.NodeFactory{}, fullText, start) {
				// slice the comment text and mask any nested block comments (safety)
				commentText := maskBlockComments(fullText[comment.Pos():comment.End()])
				if loc := lineRe.FindStringSubmatchIndex(commentText); len(loc) >= 6 {
					kind := commentText[loc[2]:loc[3]]
					mod := commentText[loc[4]:loc[5]]
					// Convert to absolute range in file text by offsetting with comment.Pos()
					tr := core.NewTextRange(comment.Pos()+loc[0], comment.Pos()+loc[1])
					switch kind {
					case "types":
						if opts.Types == "never" || opts.Types == "prefer-import" {
							ctx.ReportRange(tr, buildMessage(mod))
						}
					case "path":
						if opts.Path == "never" {
							ctx.ReportRange(tr, buildMessage(mod))
						}
					case "lib":
						if opts.Lib == "never" {
							ctx.ReportRange(tr, buildMessage(mod))
						}
					}
				}
			}
		}

		return rule.RuleListeners{
			// Handle file-level triple-slash directives
			ast.KindSourceFile: func(node *ast.Node) {
				parseHeaderComments()
			},
		}
	},
})
