package dynamic_import_chunkname

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed dynamic_import_chunkname.schema.json
var schemaJSON []byte

const (
	defaultChunknameFormat = `([0-9a-zA-Z-_/.]|\[(request|index)\])+`

	// Differs from upstream: every magic comment key accepts the Rspack
	// `rspack` prefix as well as the webpack one, and `FetchPriority` is
	// recognized for both prefixes.
	prefix = `(?:webpack|rspack)`

	noLeadingCommentMessage = "dynamic imports require a leading comment with the webpack chunkname"
	nonBlockCommentMessage  = "dynamic imports require a /* foo */ style comment, not a // foo comment"
	noPaddingMessage        = "dynamic imports require a block comment padded with spaces - /* foo */"
	invalidSyntaxMessage    = `dynamic imports require a "webpack" comment with valid syntax`
	eagerModeMessage        = "dynamic imports using eager mode do not need a webpackChunkName"
)

var (
	paddedCommentRegex = esregexp.MustCompile(`^ (\S[\s\S]+\S) $`, "")

	commentStyleRegex = esregexp.MustCompile(
		`^( (`+
			`(`+prefix+`ChunkName: .+)`+
			`|((`+prefix+`(?:Prefetch|Preload)): (true|false|-?[0-9]+))`+
			`|(`+prefix+`Ignore: (true|false))`+
			`|((`+prefix+`(?:Include|Exclude)): \/.*\/)`+
			`|(`+prefix+`Mode: ["'](lazy|lazy-once|eager|weak)["'])`+
			`|(`+prefix+`Exports: (['"]\w+['"]|\[(['"]\w+['"], *)+(['"]\w+['"]*)\]))`+
			`|(`+prefix+`FetchPriority: ["'](low|high|auto)["'])`+
			`),?)+ $`, "")

	eagerModeRegex = esregexp.MustCompile(prefix+`Mode: ["']eager["'],? `, "")
	trailingComma  = esregexp.MustCompile(`,$`, "")
)

type options struct {
	importFunctions []string
	allowEmpty      bool
	chunknameFormat string
}

func parseOptions(raw []any) options {
	opts := options{chunknameFormat: defaultChunknameFormat}
	if len(raw) == 0 {
		return opts
	}
	config, ok := raw[0].(map[string]any)
	if !ok {
		return opts
	}
	if functions, ok := config["importFunctions"].([]any); ok {
		for _, function := range functions {
			if name, ok := function.(string); ok {
				opts.importFunctions = append(opts.importFunctions, name)
			}
		}
	}
	opts.allowEmpty, _ = config["allowEmpty"].(bool)
	if format, ok := config["webpackChunknameFormat"].(string); ok {
		opts.chunknameFormat = format
	}
	return opts
}

// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/dynamic-import-chunkname.js
var DynamicImportChunknameRule = rule.Rule{
	Name:   "import/dynamic-import-chunkname",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		// Upstream builds this regexp while creating the rule, so an invalid
		// format fails the whole rule there. Without a usable pattern nothing
		// can be matched here.
		chunkSubstrRegex, err := esregexp.Compile(prefix+`ChunkName: ["']`+opts.chunknameFormat+`["'],? `, "")
		if err != nil {
			return rule.RuleListeners{}
		}
		chunkSubstrFormat := `webpackChunkName: ["']` + opts.chunknameFormat + `["'],? `

		run := func(node *ast.Node, arg *ast.Node) {
			if arg == nil {
				return
			}
			comments := leadingComments(ctx, arg)

			if len(comments) == 0 && !opts.allowEmpty {
				ctx.ReportNode(node, rule.RuleMessage{Description: noLeadingCommentMessage})
				return
			}

			text := ctx.SourceFile.Text()
			isChunknamePresent := false
			isEagerModePresent := false

			for _, comment := range comments {
				if comment.Kind != ast.KindMultiLineCommentTrivia {
					ctx.ReportNode(node, rule.RuleMessage{Description: nonBlockCommentMessage})
					return
				}
				value := utils.CommentValue(text, comment)

				if !paddedCommentRegex.Test(value) {
					ctx.ReportNode(node, rule.RuleMessage{Description: noPaddingMessage})
					return
				}
				// Webpack evaluates the comment as an object literal body.
				if !isValidWebpackCommentBody(value) || !commentStyleRegex.Test(value) {
					ctx.ReportNode(node, rule.RuleMessage{Description: invalidSyntaxMessage})
					return
				}
				if eagerModeRegex.Test(value) {
					isEagerModePresent = true
				}
				if chunkSubstrRegex.Test(value) {
					isChunknamePresent = true
				}
			}

			if isChunknamePresent && isEagerModePresent {
				ctx.ReportNodeWithDeferredSuggestions(node, rule.RuleMessage{Description: eagerModeMessage}, func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{
						{
							Message:  rule.RuleMessage{Description: "Remove webpackChunkName"},
							FixesArr: removeMatchingFix(text, comments, chunkSubstrRegex),
						},
						{
							Message:  rule.RuleMessage{Description: "Remove webpackMode"},
							FixesArr: removeMatchingFix(text, comments, eagerModeRegex),
						},
					}
				})
			}

			if !isChunknamePresent && !opts.allowEmpty && !isEagerModePresent {
				ctx.ReportNode(node, rule.RuleMessage{
					Description: "dynamic imports require a leading comment in the form /*" + chunkSubstrFormat + "*/",
				})
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := utils.ESTreeCallCallee(call.Expression)
				if callee == nil {
					return
				}
				switch {
				case callee.Kind == ast.KindImportKeyword:
				case callee.Kind == ast.KindIdentifier && slices.Contains(opts.importFunctions, callee.Text()):
				default:
					return
				}
				var arg *ast.Node
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					arg = call.Arguments.Nodes[0]
				}
				run(node, arg)
			},
		}
	},
}

// leadingComments returns the comments between the token before arg and arg
// itself (ESLint's sourceCode.getCommentsBefore(arg)).
func leadingComments(ctx rule.RuleContext, arg *ast.Node) []*ast.CommentRange {
	expression := utils.ESTreeRuntimeExpression(arg)
	if expression == nil {
		expression = arg
	}
	start := utils.TrimNodeTextRange(ctx.SourceFile, expression).Pos()
	var comments []*ast.CommentRange
	for comment := range utils.GetCommentsInRange(ctx.SourceFile, core.NewTextRange(expression.Pos(), start)) {
		comments = append(comments, &comment)
	}
	return comments
}

// removeMatchingFix removes the first matching part from the first comment
// whose value matches pattern, deleting the comment when nothing remains.
func removeMatchingFix(text string, comments []*ast.CommentRange, pattern *esregexp.RegExp) []rule.RuleFix {
	for _, comment := range comments {
		value := utils.CommentValue(text, comment)
		if !pattern.Test(value) {
			continue
		}
		stripped, err := pattern.ReplaceFirst(value, "")
		if err != nil {
			return nil
		}
		replacement := ecmascript.StringTrim(stripped)
		if trimmed, err := trailingComma.ReplaceFirst(replacement, ""); err == nil {
			replacement = trimmed
		}
		textRange := core.NewTextRange(comment.Pos(), comment.End())
		if replacement == "" {
			return []rule.RuleFix{rule.RuleFixRemoveRange(textRange)}
		}
		return []rule.RuleFix{rule.RuleFixReplaceRange(textRange, "/* "+replacement+" */")}
	}
	return nil
}
