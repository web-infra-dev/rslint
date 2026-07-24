package triple_slash_reference

import (
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type TripleSlashReferenceOptions struct {
	Lib   string `json:"lib"`   // "always" | "never"
	Path  string `json:"path"`  // "always" | "never"
	Types string `json:"types"` // "always" | "never" | "prefer-import"
}

type pendingReference struct {
	textRange core.TextRange
	module    string
}

// ECMAScript's \s includes WhiteSpace and LineTerminator code points. Keep the
// set explicit: Go regexp's \s is ASCII-only and would not match the upstream
// rule for comments containing non-ASCII whitespace.
const ecmascriptWhitespace = `[\t\n\v\f\r \x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}]`

// This is the upstream regexp applied to ESLint's Comment.value. Comment.value
// omits the leading "//", so a real triple-slash comment starts with "/".
// Preserve its deliberately permissive details: no required whitespace before
// the reference kind, "|" is accepted as a quote, and the module capture is
// greedy and does not require a closing tag.
var tripleSlashReferencePattern = regexp.MustCompile(
	`^/` + ecmascriptWhitespace + `*<reference` +
		ecmascriptWhitespace + `*(types|path|lib)` +
		ecmascriptWhitespace + `*=` +
		ecmascriptWhitespace + `*["|'](.*)["|']`,
)

// GetLeadingCommentRanges only calls NodeFactory.NewCommentRange, which is a
// pure value constructor. Reuse one hook-free factory instead of allocating
// the comparatively large NodeFactory once per linted file.
var tripleSlashReferenceCommentFactory = ast.NewNodeFactory(ast.NodeFactoryHooks{})

// TripleSlashReferenceRule implements the triple-slash-reference rule.
// Disallow certain triple slash directives in favor of import declarations.
var TripleSlashReferenceRule = rule.CreateRule(rule.Rule{
	Name: "triple-slash-reference",
	Run:  run,
})

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := parseOptions(options)
	if opts.Lib == "always" && opts.Path == "always" && opts.Types == "always" {
		return rule.RuleListeners{}
	}

	text := ctx.SourceFile.Text()
	references := make([]pendingReference, 0)

	// sourceCode.getCommentsBefore(Program) only exposes the comments in the
	// file's leading trivia. Scanning that prefix also avoids materializing the
	// linter's full-file token/comment cache for this rule.
	for comment := range scanner.GetLeadingCommentRanges(tripleSlashReferenceCommentFactory, text, 0) {
		if comment.Kind != ast.KindSingleLineCommentTrivia ||
			comment.Pos() < 0 ||
			comment.Pos()+2 > comment.End() ||
			comment.End() > len(text) {
			continue
		}

		commentValue := text[comment.Pos()+2 : comment.End()]
		match := tripleSlashReferencePattern.FindStringSubmatchIndex(commentValue)
		if match == nil {
			continue
		}

		referenceKind := commentValue[match[2]:match[3]]
		module := commentValue[match[4]:match[5]]
		textRange := core.NewTextRange(comment.Pos(), comment.End())

		switch referenceKind {
		case "lib":
			if opts.Lib == "never" {
				reportReference(&ctx, textRange, module)
			}
		case "path":
			if opts.Path == "never" {
				reportReference(&ctx, textRange, module)
			}
		case "types":
			switch opts.Types {
			case "never":
				reportReference(&ctx, textRange, module)
			case "prefer-import":
				references = append(references, pendingReference{
					textRange: textRange,
					module:    module,
				})
			}
		}
	}

	if len(references) == 0 {
		return rule.RuleListeners{}
	}

	reportMatchingReferences := func(module string) {
		for _, reference := range references {
			if reference.module == module {
				reportReference(&ctx, reference.textRange, reference.module)
			}
		}
	}

	return rule.RuleListeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			declaration := node.AsImportDeclaration()
			module, ok := utils.GetStaticStringLiteralValue(declaration.ModuleSpecifier)
			if ok {
				reportMatchingReferences(module)
			}
		},
		ast.KindImportEqualsDeclaration: func(node *ast.Node) {
			declaration := node.AsImportEqualsDeclaration()
			if declaration.ModuleReference == nil ||
				declaration.ModuleReference.Kind != ast.KindExternalModuleReference {
				return
			}

			externalReference := declaration.ModuleReference.AsExternalModuleReference()
			module, ok := utils.GetStaticStringLiteralValue(externalReference.Expression)
			if ok {
				reportMatchingReferences(module)
			}
		},
	}
}

func parseOptions(options []any) TripleSlashReferenceOptions {
	opts := TripleSlashReferenceOptions{
		Lib:   "always",
		Path:  "never",
		Types: "prefer-import",
	}

	if optionsMap := utils.GetOptionsMap(options); optionsMap != nil {
		if lib, ok := optionsMap["lib"].(string); ok {
			opts.Lib = lib
		}
		if path, ok := optionsMap["path"].(string); ok {
			opts.Path = path
		}
		if types, ok := optionsMap["types"].(string); ok {
			opts.Types = types
		}
	}

	return opts
}

func reportReference(ctx *rule.RuleContext, textRange core.TextRange, module string) {
	ctx.ReportRange(
		textRange,
		rule.RuleMessage{
			Id:          "tripleSlashReference",
			Description: "Do not use a triple slash reference for " + module + ", use `import` style instead.",
			Data: map[string]string{
				"module": module,
			},
		},
	)
}
