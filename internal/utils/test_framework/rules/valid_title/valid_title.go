// Package valid_title is the framework-neutral body of valid-title for Jest and
// Rstest. Adapters supply registration parsing, the printf specifier set of the
// framework's parameterized titles, and legacy aliases.
package valid_title

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed valid_title.schema.json
var schemaJSON []byte

// Config configures valid-title for one test framework.
type Config struct {
	Name string
	// FunctionNameDataKey is the message data key that carries the function
	// name (`jestFunctionName`, `rstestFunctionName`).
	FunctionNameDataKey string
	// InvalidEachSpecifier matches the first printf specifier the framework
	// does not format. It is applied after `%%` has been removed.
	InvalidEachSpecifier *regexp.Regexp
	// LegacyAliases maps a registration name to the API it aliases
	// (`fit` → `it`), so matcher groups, messages and the duplicate-prefix check
	// key off what is registered rather than the local spelling.
	LegacyAliases map[string]string
	Prepare       func(rule.RuleContext) Runtime
}

// Runtime supplies framework-specific call semantics to the shared rule.
type Runtime struct {
	Parse func(*ast.Node) *testFramework.ParsedCall
	// IsParameterized reports whether the registration went through an
	// `.each` / `.for` style factory. Only array-based factories format
	// their titles with printf, which the shared rule derives from the callee
	// shape.
	IsParameterized func(node *ast.Node, parsed *testFramework.ParsedCall) bool
}

func binaryExprContainsStringLit(n *ast.Node) bool {
	if n == nil || n.Kind != ast.KindBinaryExpression {
		return false
	}
	be := n.AsBinaryExpression()
	if be == nil || be.OperatorToken == nil {
		return false
	}
	if ast.IsLogicalOrCoalescingBinaryOperator(be.OperatorToken.Kind) ||
		ast.IsAssignmentOperator(be.OperatorToken.Kind) ||
		be.OperatorToken.Kind == ast.KindCommaToken {
		return false
	}
	if ast.IsStringLiteralLike(be.Left) {
		return true
	}
	if ast.IsStringLiteralLike(be.Right) {
		return true
	}
	return binaryExprContainsStringLit(be.Left)
}

// rawTemplateLiteralText returns the contents of a template literal that has no
// substitutions. The range has to come from TrimNodeTextRange first: ast.Pos()
// includes leading trivia, so stripping the backticks by offsetting Pos()
// directly would cut into a comment or whitespace instead.
func rawTemplateLiteralText(sourceFile *ast.SourceFile, node *ast.Node) string {
	if sourceFile == nil {
		return ""
	}
	r := utils.TrimNodeTextRange(sourceFile, node)
	start := r.Pos()
	end := r.End()
	sourceText := sourceFile.Text()
	if sourceText == "" || start+1 >= end-1 {
		return ""
	}
	if end-1 > len(sourceText) || start+1 < 0 {
		return ""
	}
	return sourceText[start+1 : end-1]
}

func staticTitle(sourceFile *ast.SourceFile, n *ast.Node) (string, bool) {
	if n == nil {
		return "", false
	}
	switch n.Kind {
	case ast.KindStringLiteral:
		return n.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return rawTemplateLiteralText(sourceFile, n), true
	default:
		return "", false
	}
}

// matcherFor keys the mustMatch / mustNotMatch groups off the semantic API name.
// There is no fallback group: a future root API would otherwise silently
// inherit the `it` patterns.
func matcherFor(fnName string, ms matchersByFn) matcherEntry {
	switch fnName {
	case "describe":
		return ms.describe
	case "test":
		return ms.test
	case "it":
		return ms.it
	default:
		return matcherEntry{}
	}
}

var (
	reAccOpen  = regexp.MustCompile(`^([\x60'"]) +`)
	reAccClose = regexp.MustCompile(` +([\x60'"])$`)
)

func accidentalSpaceReplacement(rawSrc string) string {
	s := reAccOpen.ReplaceAllString(rawSrc, "$1")
	s = reAccClose.ReplaceAllString(s, "$1")
	return s
}

// duplicatePrefixReplacement drops the leading `<name> ` from a title literal.
// The fix is only safe when the raw source spells both the name and the
// separating space literally: an escaped spelling (`'te\u0073t foo'`,
// `'test\u0020foo'`) is still reported, but without an edit.
func duplicatePrefixReplacement(rawSrc string, fnName string) (string, bool) {
	if fnName == "" || len(rawSrc) < len(fnName)+3 {
		return "", false
	}
	prefixEnd := 1 + len(fnName)
	if rawSrc[len(rawSrc)-1] != rawSrc[0] {
		return "", false
	}
	switch rawSrc[0] {
	case '`', '\'', '"':
	default:
		return "", false
	}
	if rawSrc[prefixEnd] != ' ' || !ecmascript.EqualsWhenLowercased(rawSrc[1:prefixEnd], fnName) {
		return "", false
	}
	return rawSrc[:1] + rawSrc[prefixEnd+1:], true
}

func regexpToMessagePattern(re *esregexp.RegExp) string {
	if re == nil {
		return ""
	}
	// Source rather than String: the pattern reaches regexp2 rewritten, and
	// the message has to show what its author wrote.
	src := re.Source()
	return "/" + strings.ReplaceAll(src, "/", "\\/") + "/u"
}

func emptyFunctionName(kind testFramework.FnKind) string {
	if kind == testFramework.FnKindDescribe {
		return "describe"
	}
	return "test"
}

// semanticName is the API a registration belongs to. A describe registration is
// always `describe`: Playwright exposes it as `test.describe`, where the parser
// keeps the root API name.
func (c Config) semanticName(parsed *testFramework.ParsedCall) string {
	if parsed.Kind == testFramework.FnKindDescribe {
		return "describe"
	}
	if alias, ok := c.LegacyAliases[parsed.Name]; ok {
		return alias
	}
	return parsed.Name
}

// NewRule creates a valid-title rule for a test framework.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			co := parseCompiledOptions(options)
			if len(co.invalidPatterns) > 0 {
				for _, bad := range co.invalidPatterns {
					ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{
						Id: "invalidPattern",
						Description: fmt.Sprintf(
							"Invalid regular expression in `%s` option: `%s`: %s",
							bad.optionPath, bad.pattern, bad.err.Error(),
						),
					})
				}
				return rule.RuleListeners{}
			}
			runtime := config.Prepare(ctx)

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if parsed == nil {
						return
					}
					if parsed.Kind != testFramework.FnKindDescribe && parsed.Kind != testFramework.FnKindTest {
						return
					}

					call := node.AsCallExpression()
					if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
						return
					}
					arg := call.Arguments.Nodes[0]

					title, ok := staticTitle(ctx.SourceFile, arg)
					if !ok {
						if binaryExprContainsStringLit(arg) {
							return
						}
						ignored := false
						if parsed.Kind == testFramework.FnKindDescribe && co.ignoreTypeOfDescribeName {
							ignored = true
						}
						if parsed.Kind == testFramework.FnKindTest && co.ignoreTypeOfTestName {
							ignored = true
						}
						if !ignored && arg.Kind != ast.KindTemplateExpression {
							ctx.ReportNode(arg, rule.RuleMessage{
								Id:          "titleMustBeString",
								Description: "Title must be a string",
							})
						}
						return
					}

					if title == "" {
						name := emptyFunctionName(parsed.Kind)
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "emptyTitle",
							Description: name + " should not have an empty title",
							Data:        map[string]string{config.FunctionNameDataKey: name},
						})
						return
					}

					if runtime.IsParameterized != nil && runtime.IsParameterized(node, parsed) {
						// Only the outer call of an array-based factory
						// (`test.each(rows)(title, fn)`) formats its title with
						// printf; a tagged-template table interpolates `$var`.
						if callee := ast.SkipParentheses(call.Expression); callee != nil && callee.Kind == ast.KindCallExpression {
							s := strings.ReplaceAll(title, "%%", "")
							if spec := config.InvalidEachSpecifier.FindString(s); spec != "" {
								ctx.ReportNode(arg, rule.RuleMessage{
									Id:          "invalidEachSpecifier",
									Description: fmt.Sprintf("%q is not a valid format specifier", spec),
									Data:        map[string]string{"specifier": spec},
								})
							}
						}
					}

					if co.disallowedConcat != nil {
						m, err := co.disallowedConcat.Unwrap().FindStringMatch(title)
						if err == nil && m != nil {
							g := m.GroupByNumber(1)
							if g != nil && g.String() != "" {
								word := g.String()
								ctx.ReportNode(arg, rule.RuleMessage{
									Id:          "disallowedWord",
									Description: fmt.Sprintf("%q is not allowed in test titles", word),
									Data:        map[string]string{"word": word},
								})
								return
							}
						}
					}

					// accidentalSpace and duplicatePrefix both fall through, so a
					// title like ' describe foo' reports twice. That is upstream
					// behaviour.
					if !co.ignoreSpaces {
						trimmed := ecmascript.StringTrim(title)
						if len(trimmed) != len(title) {
							ctx.ReportNodeWithDeferredFixes(arg, rule.RuleMessage{
								Id:          "accidentalSpace",
								Description: "should not have leading or trailing spaces",
							}, func() []rule.RuleFix {
								raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, arg, false)
								fix := accidentalSpaceReplacement(raw)
								if fix == raw {
									return nil
								}
								return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, arg, fix)}
							})
						}
					}

					fnName := config.semanticName(parsed)
					firstTok := title
					if i := strings.IndexByte(title, ' '); i >= 0 {
						firstTok = title[:i]
					}
					if ecmascript.EqualsWhenLowercased(firstTok, fnName) {
						ctx.ReportNodeWithDeferredFixes(arg, rule.RuleMessage{
							Id:          "duplicatePrefix",
							Description: "should not have duplicate prefix",
						}, func() []rule.RuleFix {
							raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, arg, false)
							if fix, ok := duplicatePrefixReplacement(raw, fnName); ok {
								return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, arg, fix)}
							}
							return nil
						})
					}

					if me := matcherFor(fnName, co.mustNotMatch); me.re.Test(title) {
						config.reportMustNot(ctx, arg, fnName, me)
						return
					}

					me := matcherFor(fnName, co.mustMatch)
					if me.re != nil && !me.re.TestOrTimeout(title) {
						config.reportMustMatch(ctx, arg, fnName, me)
					}
				},
			}
		},
	}
}

func (c Config) reportMustNot(ctx rule.RuleContext, arg *ast.Node, fnName string, me matcherEntry) {
	if me.customText != "" {
		ctx.ReportNode(arg, rule.RuleMessage{
			Id:          "mustNotMatchCustom",
			Description: me.customText,
			Data:        map[string]string{"message": me.customText},
		})
		return
	}
	patStr := regexpToMessagePattern(me.re)
	ctx.ReportNode(arg, rule.RuleMessage{
		Id:          "mustNotMatch",
		Description: fmt.Sprintf("%s should not match %s", fnName, patStr),
		Data: map[string]string{
			c.FunctionNameDataKey: fnName,
			"pattern":             patStr,
		},
	})
}

func (c Config) reportMustMatch(ctx rule.RuleContext, arg *ast.Node, fnName string, me matcherEntry) {
	if me.customText != "" {
		ctx.ReportNode(arg, rule.RuleMessage{
			Id:          "mustMatchCustom",
			Description: me.customText,
			Data:        map[string]string{"message": me.customText},
		})
		return
	}
	patStr := regexpToMessagePattern(me.re)
	ctx.ReportNode(arg, rule.RuleMessage{
		Id:          "mustMatch",
		Description: fmt.Sprintf("%s should match %s", fnName, patStr),
		Data: map[string]string{
			c.FunctionNameDataKey: fnName,
			"pattern":             patStr,
		},
	})
}
