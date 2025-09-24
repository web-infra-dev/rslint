package triple_slash_reference

import (
    "regexp"

    "github.com/microsoft/typescript-go/shim/ast"
    "github.com/microsoft/typescript-go/shim/core"
    "github.com/web-infra-dev/rslint/internal/rule"
)

type tripleSlashRef struct {
    importName string
    rng        core.TextRange
}

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

            text := sf.Text()

            // Determine the position of the first statement to restrict to file header comments
            firstStmtPos := len(text)
            if sf.Statements != nil && len(sf.Statements.Nodes) > 0 {
                firstStmtPos = sf.Statements.Nodes[0].Pos()
            }

            // Scan header for triple-slash references. In some cases the
            // first statement position can be 0; fall back to scanning the
            // whole file to ensure we catch standalone directives.
            header := text[:firstStmtPos]
            scan := header
            if firstStmtPos == 0 {
                scan = text
            }
            // (?m) multiline: match from the start of a line optional spaces then /// <reference ...>
            lineRe := regexp.MustCompile(`(?m)^[ \t]*///[ \t]*<reference[ \t]*(types|path|lib)[ \t]*=[ \t]*["']([^"']+)["']`)
            idxs := lineRe.FindAllStringSubmatchIndex(scan, -1)
            for _, m := range idxs {
                if len(m) < 6 {
                    continue
                }
                start := m[0]
                end := m[1]
                kind := scan[m[2]:m[3]]
                mod := scan[m[4]:m[5]]
                tr := core.NewTextRange(start, end)

                switch kind {
                case "types":
                    // Upstream behavior: report immediately for both
                    // "never" and "prefer-import".
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

        return rule.RuleListeners{
            // Handle file-level triple-slash directives
            ast.KindSourceFile: func(node *ast.Node) {
                parseHeaderComments()
            },
        }
    },
})
