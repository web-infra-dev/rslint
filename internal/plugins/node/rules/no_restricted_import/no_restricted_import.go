package no_restricted_import

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_restricted_import.schema.json
var schemaJSON []byte

type restrictionPattern struct {
	matcher           *nodeutil.GlobMatcher
	absolute, negated bool
}

type restriction struct {
	patterns []restrictionPattern
	message  string
}

func parseRestrictions(options []any) []restriction {
	if len(options) == 0 {
		return nil
	}
	definitions, _ := options[0].([]any)
	restrictions := make([]restriction, 0, len(definitions))
	for _, definition := range definitions {
		var names []string
		var message string
		switch value := definition.(type) {
		case string:
			names = []string{value}
		case map[string]any:
			if name, ok := value["name"].(string); ok {
				names = []string{name}
			} else {
				names = utils.ToStringSlice(value["name"])
			}
			message, _ = value["message"].(string)
		}
		if message != "" {
			message = " " + message
		}
		r := restriction{message: message}
		for _, name := range names {
			negated := strings.HasPrefix(name, "!") && !strings.HasPrefix(name, "!(")
			if negated {
				name = name[1:]
			}
			// Preserve the pattern: changing separators changes glob semantics.
			absolute := nodeutil.IsAbsolutePath(name)
			r.patterns = append(r.patterns, restrictionPattern{nodeutil.CompileGlob(name), absolute, negated})
		}
		restrictions = append(restrictions, r)
	}
	return restrictions
}

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-restricted-import.js
var NoRestrictedImportRule = rule.Rule{
	Name:   "node/no-restricted-import",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		restrictions := parseRestrictions(options)
		if len(restrictions) == 0 {
			return nil
		}
		var resolutionOptions [2]*nodeutil.ResolutionOptions
		return nodeutil.VisitImports(nodeutil.ImportVisitorOptions{IncludeCore: true}, func(source *ast.Node, name string, typeOnly bool) {
			var filePath string
			resolved := false
			for _, restriction := range restrictions {
				matched := false
				for _, pattern := range restriction.patterns {
					// Positive patterns add matches, negatives remove them in order.
					if pattern.negated != matched {
						continue
					}
					target := name
					if pattern.absolute {
						if !resolved && ctx.Program() != nil {
							index := 0
							if typeOnly {
								index = 1
							}
							if resolutionOptions[index] == nil {
								resolution := nodeutil.ImportResolutionOptions(ctx, typeOnly, nil)
								resolutionOptions[index] = &resolution
							}
							filePath = nodeutil.ImportFilePath(ctx.Program(), name, ctx.SourceFile.FileName(), typeOnly, *resolutionOptions[index])
							resolved = true
						}
						if filePath == "" {
							continue
						}
						target = filePath
					}
					if pattern.matcher.Match(target) {
						matched = !pattern.negated
					}
				}
				if matched {
					ctx.ReportNode(source, rule.RuleMessage{
						Id: "restricted", Description: "'" + name + "' module is restricted from being used." + restriction.message,
						Data: map[string]string{"name": name, "customMessage": restriction.message},
					})
					return
				}
			}
		})
	},
}
