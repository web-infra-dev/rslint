// Package valid_title ports eslint-plugin-jest's valid-title to Rstest.
//
// The rule body is deliberately independent of the jest implementation rather
// than shared: almost every step of it is a "report or not" branch, and four of
// those branches take different values on Rstest (see the D1-D5 comments
// below). Only the deterministic primitives — regexp compilation, whitespace
// classification, text ranges — are reused, from internal/utils.
package valid_title

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed valid_title.schema.json
var schemaJSON []byte

type matcherEntry struct {
	re *regexp2.Regexp
	// customText non-empty ⇒ use mustMatchCustom / mustNotMatchCustom
	customText string
}

type matchersByFn struct {
	describe matcherEntry
	test     matcherEntry
	it       matcherEntry
}

type compiledOptions struct {
	ignoreSpaces             bool
	ignoreTypeOfDescribeName bool
	ignoreTypeOfTestName     bool
	disallowedConcat         *regexp2.Regexp
	invalidPatterns          []invalidPattern
	mustNotMatch             matchersByFn
	mustMatch                matchersByFn
}

type invalidPattern struct {
	optionPath string
	pattern    string
	err        error
}

func firstOptionMap(options []any) map[string]interface{} {
	if len(options) == 0 {
		return nil
	}
	m, ok := options[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return m
}

func boolFromMap(m map[string]interface{}, key string, def bool) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

func compileRE2(pat string) (*regexp2.Regexp, error) {
	re, err := utils.CompileRegexp2(pat, utils.JSUnicodeRegexOptions)
	if err != nil {
		return nil, err
	}
	return re, nil
}

func compileMatcherPatterns(raw interface{}, optionPath string) (matchersByFn, []invalidPattern) {
	out := matchersByFn{}
	var invalids []invalidPattern
	if raw == nil {
		return out, nil
	}

	setAll := func(e matcherEntry) {
		out.describe, out.test, out.it = e, e, e
	}

	switch x := raw.(type) {
	case string:
		if x == "" {
			break
		}
		if re, err := compileRE2(x); err != nil {
			invalids = append(invalids, invalidPattern{
				optionPath: optionPath,
				pattern:    x,
				err:        err,
			})
		} else if re != nil {
			setAll(matcherEntry{re: re})
		}
	case []interface{}:
		me := matcherEntry{}
		if len(x) >= 1 {
			if s, ok := x[0].(string); ok && s != "" {
				re, err := compileRE2(s)
				if err != nil {
					invalids = append(invalids, invalidPattern{
						optionPath: optionPath,
						pattern:    s,
						err:        err,
					})
				} else {
					me.re = re
				}
			}
		}
		if len(x) >= 2 {
			if s, ok := x[1].(string); ok {
				me.customText = s
			}
		}
		if me.re != nil {
			setAll(me)
		}
	case map[string]interface{}:
		for _, key := range []string{"describe", "test", "it"} {
			if v, ok := x[key]; ok {
				invalids = append(invalids, fillMatcherField(&out, key, v, optionPath+"."+key)...)
			}
		}
	}
	return out, invalids
}

func fillMatcherField(ms *matchersByFn, key string, raw interface{}, optionPath string) []invalidPattern {
	e := matcherEntry{}
	var invalids []invalidPattern

	switch x := raw.(type) {
	case string:
		if x == "" {
			break
		}
		re, err := compileRE2(x)
		if err != nil {
			invalids = append(invalids, invalidPattern{
				optionPath: optionPath,
				pattern:    x,
				err:        err,
			})
		} else {
			e.re = re
		}
	case []interface{}:
		if len(x) >= 1 {
			if s, ok := x[0].(string); ok && s != "" {
				re, err := compileRE2(s)
				if err != nil {
					invalids = append(invalids, invalidPattern{
						optionPath: optionPath,
						pattern:    s,
						err:        err,
					})
				} else {
					e.re = re
				}
			}
		}
		if len(x) >= 2 {
			if s, ok := x[1].(string); ok {
				e.customText = s
			}
		}
	}

	switch key {
	case "describe":
		ms.describe = e
	case "test":
		ms.test = e
	case "it":
		ms.it = e
	}

	return invalids
}

func parseCompiledOptions(options []any) compiledOptions {
	m := firstOptionMap(options)
	if m == nil {
		return compiledOptions{}
	}

	co := compiledOptions{
		ignoreSpaces:             boolFromMap(m, "ignoreSpaces", false),
		ignoreTypeOfDescribeName: boolFromMap(m, "ignoreTypeOfDescribeName", false),
		ignoreTypeOfTestName:     boolFromMap(m, "ignoreTypeOfTestName", false),
	}

	if dw, ok := m["disallowedWords"]; ok && dw != nil {
		co.disallowedConcat, co.invalidPatterns = compileDisallowedWords(dw, co.invalidPatterns)
	}

	if mn, ok := m["mustNotMatch"]; ok {
		var invalids []invalidPattern
		co.mustNotMatch, invalids = compileMatcherPatterns(mn, "mustNotMatch")
		co.invalidPatterns = append(co.invalidPatterns, invalids...)
	}
	if mm, ok := m["mustMatch"]; ok {
		var invalids []invalidPattern
		co.mustMatch, invalids = compileMatcherPatterns(mm, "mustMatch")
		co.invalidPatterns = append(co.invalidPatterns, invalids...)
	}

	return co
}

func compileDisallowedWords(raw interface{}, invalids []invalidPattern) (*regexp2.Regexp, []invalidPattern) {
	items, ok := raw.([]interface{})
	if !ok || len(items) == 0 {
		return nil, invalids
	}
	parts := make([]string, 0, len(items))
	for _, it := range items {
		w, ok := it.(string)
		if ok && w != "" {
			parts = append(parts, w)
		}
	}
	if len(parts) == 0 {
		return nil, invalids
	}
	pattern := "(?i)\\b(" + strings.Join(parts, "|") + ")\\b"
	re, err := compileRE2(pattern)
	if err != nil {
		invalids = append(invalids, invalidPattern{
			optionPath: "disallowedWords",
			pattern:    pattern,
			err:        err,
		})
		return nil, invalids
	}
	return re, invalids
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

// matcherFor keys the mustMatch / mustNotMatch groups off the semantic API name
// (D5), so an aliased registration is grouped by what it registers rather than
// by the local identifier. Unlike the jest implementation there is no fallback
// group: a future root API would otherwise silently inherit the `it` patterns.
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

// isArrayParameterizedRegistration reports whether this registration is the
// outer call of an array-based `.each` / `.for` factory, which is the only shape
// whose title goes through Rstest's printf formatting.
//
// D2: `.for` formats its title with the same formatName as `.each`
// (packages/core/src/runtime/runner/runtime.ts:466-576 at rstest c4b67c72), so
// both are covered — jest has no `.for` at all.
//
// D3: the jest implementation walks MemberEntries looking for `each`. That view
// is call-site only and goes empty once the factory is reached through an alias,
// so the shape of the callee is used instead: a tagged-template table
// (test.each`…`) interpolates with $var and is skipped, exactly as upstream
// skips it.
func isArrayParameterizedRegistration(parsed *rstestUtils.ParsedRstestFnCall, call *ast.CallExpression) bool {
	if parsed == nil || !parsed.IsParameterized() || call == nil {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	return callee != nil && callee.Kind == ast.KindCallExpression
}

var (
	reAccOpen  = regexp.MustCompile(`^([\x60'"]) +`)
	reAccClose = regexp.MustCompile(` +([\x60'"])$`)
	// D1: the accepted specifiers are Rstest's, not jest's. formatRegExp is
	// /%[sdjifoOc%]/ (packages/core/src/runtime/util.ts:159 at rstest c4b67c72)
	// and %# / %$ are substituted ahead of it by formatName, so `%p` — valid to
	// jest — is left verbatim by Rstest, while `%O` and `%c` — reported by jest
	// — are expanded by formatTemplate.
	// cspell:ignore sdjifo
	reEachInvalidSpecifier = regexp.MustCompile(`%[^sdjifoOc#$%]`)
)

func accidentalSpaceReplacement(rawSrc string) string {
	s := reAccOpen.ReplaceAllString(rawSrc, "$1")
	s = reAccClose.ReplaceAllString(s, "$1")
	return s
}

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
	if rawSrc[prefixEnd] != ' ' || !strings.EqualFold(rawSrc[1:prefixEnd], fnName) {
		return "", false
	}
	return rawSrc[:1] + rawSrc[prefixEnd+1:], true
}

func regexpToMessagePattern(re *regexp2.Regexp) string {
	if re == nil {
		return ""
	}
	src := re.String()
	return "/" + strings.ReplaceAll(src, "/", "\\/") + "/u"
}

func eachInvalidSpecifier(title string) string {
	s := strings.ReplaceAll(title, "%%", "")
	return reEachInvalidSpecifier.FindString(s)
}

func rstestEmptyFunctionName(kind rstestUtils.RstestFnType) string {
	if kind == rstestUtils.RstestFnTypeDescribe {
		return "describe"
	}
	return "test"
}

// ValidTitleRule enforces rstest/valid-title.
var ValidTitleRule = rule.Rule{
	Name:   "rstest/valid-title",
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
		analysis := rstestUtils.NewRstestCallAnalysis(ctx)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// import.meta.rstest.test(...) and @rstest/playwright are official
				// registration shapes, so title validation applies to them too.
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return
				}
				if parsed.Kind != rstestUtils.RstestFnTypeDescribe && parsed.Kind != rstestUtils.RstestFnTypeTest {
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
					if parsed.Kind == rstestUtils.RstestFnTypeDescribe && co.ignoreTypeOfDescribeName {
						ignored = true
					}
					if parsed.Kind == rstestUtils.RstestFnTypeTest && co.ignoreTypeOfTestName {
						ignored = true
					}
					if !ignored && arg.Kind != ast.KindTemplateExpression {
						ctx.ReportNodeWithFixes(arg, rule.RuleMessage{
							Id:          "titleMustBeString",
							Description: "Title must be a string",
						})
					}
					return
				}

				if title == "" {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "emptyTitle",
						Description: rstestEmptyFunctionName(parsed.Kind) + " should not have an empty title",
						Data: map[string]string{
							"rstestFunctionName": rstestEmptyFunctionName(parsed.Kind),
						},
					})
					return
				}

				if isArrayParameterizedRegistration(parsed, call) {
					if spec := eachInvalidSpecifier(title); spec != "" {
						ctx.ReportNode(arg, rule.RuleMessage{
							Id:          "invalidEachSpecifier",
							Description: fmt.Sprintf("%q is not a valid format specifier", spec),
							Data:        map[string]string{"specifier": spec},
						})
					}
				}

				if co.disallowedConcat != nil {
					m, err := co.disallowedConcat.FindStringMatch(title)
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
				// behaviour (valid-title.ts:304-333).
				if !co.ignoreSpaces {
					trimmed := strings.TrimFunc(title, utils.IsStrWhiteSpace)
					if len(trimmed) != len(title) {
						raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, arg, false)
						fix := accidentalSpaceReplacement(raw)
						if fix == raw {
							ctx.ReportNode(arg, rule.RuleMessage{
								Id:          "accidentalSpace",
								Description: "should not have leading or trailing spaces",
							})
						} else {
							ctx.ReportNodeWithFixes(arg, rule.RuleMessage{
								Id:          "accidentalSpace",
								Description: "should not have leading or trailing spaces",
							}, rule.RuleFixReplace(ctx.SourceFile, arg, fix))
						}
					}
				}

				// Playwright exposes describe registrations through test.describe.
				// The parser keeps Name as the root API ("test") while Kind records
				// the registration kind, so describe titles must use the describe
				// matcher group and prefix. Test registrations still use Name to
				// distinguish test from it after import or same-file alias resolution.
				fnName := parsed.Name
				if parsed.Kind == rstestUtils.RstestFnTypeDescribe {
					fnName = "describe"
				}
				firstTok := title
				if i := strings.IndexByte(title, ' '); i >= 0 {
					firstTok = title[:i]
				}
				if strings.EqualFold(firstTok, fnName) {
					raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, arg, false)
					msg := rule.RuleMessage{
						Id:          "duplicatePrefix",
						Description: "should not have duplicate prefix",
					}
					if fix, ok := duplicatePrefixReplacement(raw, fnName); ok {
						ctx.ReportNodeWithFixes(arg, msg, rule.RuleFixReplace(ctx.SourceFile, arg, fix))
					} else {
						ctx.ReportNode(arg, msg)
					}
				}

				if me := matcherFor(fnName, co.mustNotMatch); utils.Regexp2MatchString(me.re, title) {
					buildMustNotReport(ctx, arg, fnName, me)
					return
				}

				me := matcherFor(fnName, co.mustMatch)
				if me.re != nil && !utils.Regexp2MatchString(me.re, title) {
					buildMustMatchReport(ctx, arg, fnName, me)
				}
			},
		}
	},
}

func buildMustNotReport(ctx rule.RuleContext, arg *ast.Node, rstestFnName string, me matcherEntry) {
	if me.customText != "" {
		ctx.ReportNode(arg, rule.RuleMessage{
			Id:          "mustNotMatchCustom",
			Description: me.customText,
			Data: map[string]string{
				"message": me.customText,
			},
		})
		return
	}
	patStr := regexpToMessagePattern(me.re)
	ctx.ReportNode(arg, rule.RuleMessage{
		Id:          "mustNotMatch",
		Description: fmt.Sprintf("%s should not match %s", rstestFnName, patStr),
		Data: map[string]string{
			"rstestFunctionName": rstestFnName,
			"pattern":            patStr,
		},
	})
}

func buildMustMatchReport(ctx rule.RuleContext, arg *ast.Node, rstestFnName string, me matcherEntry) {
	if me.customText != "" {
		ctx.ReportNode(arg, rule.RuleMessage{
			Id:          "mustMatchCustom",
			Description: me.customText,
			Data: map[string]string{
				"message": me.customText,
			},
		})
		return
	}
	patStr := regexpToMessagePattern(me.re)
	ctx.ReportNode(arg, rule.RuleMessage{
		Id:          "mustMatch",
		Description: fmt.Sprintf("%s should match %s", rstestFnName, patStr),
		Data: map[string]string{
			"rstestFunctionName": rstestFnName,
			"pattern":            patStr,
		},
	})
}
