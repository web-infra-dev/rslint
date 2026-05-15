package valid_title

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	jestutils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const jsRegexOpts = regexp2.ECMAScript | regexp2.Unicode

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
	mustNotMatch             matchersByFn
	mustMatch                matchersByFn
}

func firstOptionMap(options any) map[string]interface{} {
	if options == nil {
		return nil
	}
	arr, ok := options.([]interface{})
	if !ok || len(arr) == 0 {
		return nil
	}
	m, ok := arr[0].(map[string]interface{})
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

func compileRE2(pat string) *regexp2.Regexp {
	if pat == "" {
		return nil
	}
	re, err := regexp2.Compile(pat, jsRegexOpts)
	if err != nil {
		return nil
	}
	return re
}

func matchRE2(re *regexp2.Regexp, s string) bool {
	if re == nil {
		return false
	}
	m, err := re.FindStringMatch(s)
	return err == nil && m != nil
}

func compileMatcherPatterns(raw interface{}) matchersByFn {
	out := matchersByFn{}
	if raw == nil {
		return out
	}

	setAll := func(e matcherEntry) {
		out.describe, out.test, out.it = e, e, e
	}

	switch x := raw.(type) {
	case string:
		if re := compileRE2(x); re != nil {
			setAll(matcherEntry{re: re})
		}
	case []interface{}:
		me := matcherEntry{}
		if len(x) >= 1 {
			if s, ok := x[0].(string); ok {
				me.re = compileRE2(s)
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
				fillMatcherField(&out, key, v)
			}
		}
	}
	return out
}

func fillMatcherField(ms *matchersByFn, key string, raw interface{}) {
	e := matcherEntry{}

	switch x := raw.(type) {
	case string:
		e.re = compileRE2(x)
	case []interface{}:
		if len(x) >= 1 {
			if s, ok := x[0].(string); ok {
				e.re = compileRE2(s)
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
}

func parseCompiledOptions(options any) compiledOptions {
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
		co.disallowedConcat = compileDisallowedWords(dw)
	}

	if mn, ok := m["mustNotMatch"]; ok {
		co.mustNotMatch = compileMatcherPatterns(mn)
	}
	if mm, ok := m["mustMatch"]; ok {
		co.mustMatch = compileMatcherPatterns(mm)
	}

	return co
}

func compileDisallowedWords(raw interface{}) *regexp2.Regexp {
	items, ok := raw.([]interface{})
	if !ok || len(items) == 0 {
		return nil
	}
	parts := make([]string, 0, len(items))
	for _, it := range items {
		w, ok := it.(string)
		if ok && w != "" {
			parts = append(parts, regexp.QuoteMeta(w))
		}
	}
	if len(parts) == 0 {
		return nil
	}
	pattern := "(?i)\\b(" + strings.Join(parts, "|") + ")\\b"
	return compileRE2(pattern)
}

func trimFXPrefix(word string) string {
	if word == "" {
		return ""
	}
	if word[0] == 'f' || word[0] == 'x' {
		return word[1:]
	}
	return word
}

func binaryPlusContainsStringLit(n *ast.Node) bool {
	if n == nil || n.Kind != ast.KindBinaryExpression {
		return false
	}
	be := n.AsBinaryExpression()
	if be == nil || be.OperatorToken == nil || be.OperatorToken.Kind != ast.KindPlusToken {
		return false
	}
	if utils.IsStringLiteralOrTemplate(be.Left) {
		return true
	}
	if utils.IsStringLiteralOrTemplate(be.Right) {
		return true
	}
	return binaryPlusContainsStringLit(be.Left) || binaryPlusContainsStringLit(be.Right)
}

func jestTitleInner(n *ast.Node) (string, bool) {
	if n == nil {
		return "", false
	}
	switch n.Kind {
	case ast.KindStringLiteral:
		return n.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return n.AsNoSubstitutionTemplateLiteral().Text, true
	default:
		return "", false
	}
}

func matcherFor(fnKey string, ms matchersByFn) matcherEntry {
	switch fnKey {
	case "describe":
		return ms.describe
	case "test":
		return ms.test
	default:
		return ms.it
	}
}

// Mirrors eslint-plugin-jest: only run printf checks for the outer each title call
// (when .each(...) is invoked as (...)(title, cb)), not tagged-template factories.
func shouldValidateEachPrintf(jestFn *jestutils.ParsedJestFnCall, call *ast.CallExpression) bool {
	if jestFn == nil || call == nil {
		return false
	}
	for _, entry := range jestFn.MemberEntries {
		if entry.Name != "each" || entry.Node == nil {
			continue
		}
		parent := entry.Node.Parent
		if parent == nil {
			continue
		}
		gp := parent.Parent
		return gp != nil && gp.Kind == ast.KindCallExpression
	}
	return false
}

var (
	reDupPrefix            = regexp.MustCompile(`^([\x60'"]).+?\s+`)
	reAccOpen              = regexp.MustCompile(`^([\x60'"]) +`)
	reAccClose             = regexp.MustCompile(` +([\x60'"])$`)
	reEachInvalidSpecifier = regexp.MustCompile(`%[^psdifjo#$%]`)
)

func accidentalSpaceReplacement(rawSrc string) string {
	s := reAccOpen.ReplaceAllString(rawSrc, "$1")
	s = reAccClose.ReplaceAllString(s, "$1")
	return s
}

func duplicatePrefixReplacement(rawSrc string) string {
	return reDupPrefix.ReplaceAllString(rawSrc, "$1")
}

func regexpToMessagePattern(re *regexp2.Regexp) string {
	if re == nil {
		return ""
	}
	src := re.String()
	return "/" + strings.ReplaceAll(src, "/", "\\/") + "/"
}

func eachInvalidSpecifier(title string) string {
	s := strings.ReplaceAll(title, "%%", "")
	return reEachInvalidSpecifier.FindString(s)
}

func jestEmptyFunctionName(kind jestutils.JestFnType) string {
	if kind == jestutils.JestFnTypeDescribe {
		return "describe"
	}
	return "test"
}

// ValidTitleRule enforces ESLint jest/valid-title.
var ValidTitleRule = rule.Rule{
	Name: "jest/valid-title",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		co := parseCompiledOptions(options)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFn := jestutils.ParseJestFnCall(node, ctx)
				if jestFn == nil {
					return
				}
				if jestFn.Kind != jestutils.JestFnTypeDescribe && jestFn.Kind != jestutils.JestFnTypeTest {
					return
				}

				call := node.AsCallExpression()
				if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				arg := call.Arguments.Nodes[0]

				title, ok := jestTitleInner(arg)
				if !ok {
					if binaryPlusContainsStringLit(arg) {
						return
					}
					ignored := false
					if jestFn.Kind == jestutils.JestFnTypeDescribe && co.ignoreTypeOfDescribeName {
						ignored = true
					}
					if jestFn.Kind == jestutils.JestFnTypeTest && co.ignoreTypeOfTestName {
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
					ctx.ReportNode(arg, rule.RuleMessage{
						Id:          "emptyTitle",
						Description: jestEmptyFunctionName(jestFn.Kind) + " should not have an empty title",
						Data: map[string]string{
							"jestFunctionName": jestEmptyFunctionName(jestFn.Kind),
						},
					})
					return
				}

				if shouldValidateEachPrintf(jestFn, call) {
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

				if !co.ignoreSpaces {
					trimmed := strings.TrimFunc(title, utils.IsStrWhiteSpace)
					if len(trimmed) != len(title) {
						raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, arg, false)
						fix := accidentalSpaceReplacement(raw)
						ctx.ReportNodeWithFixes(arg, rule.RuleMessage{
							Id:          "accidentalSpace",
							Description: "should not have leading or trailing spaces",
						}, rule.RuleFixReplace(ctx.SourceFile, arg, fix))
					}
				}

				unpref := trimFXPrefix(jestFn.Name)
				firstTok := title
				if i := strings.IndexByte(title, ' '); i >= 0 {
					firstTok = title[:i]
				}
				if strings.EqualFold(firstTok, unpref) {
					raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, arg, false)
					fix := duplicatePrefixReplacement(raw)
					ctx.ReportNodeWithFixes(arg, rule.RuleMessage{
						Id:          "duplicatePrefix",
						Description: "should not have duplicate prefix",
					}, rule.RuleFixReplace(ctx.SourceFile, arg, fix))
				}

				fnKey := trimFXPrefix(jestFn.Name)

				if me := matcherFor(fnKey, co.mustNotMatch); matchRE2(me.re, title) {
					buildMustNotReport(ctx, arg, unpref, me)
					return
				}

				me := matcherFor(fnKey, co.mustMatch)
				if me.re != nil && !matchRE2(me.re, title) {
					buildMustMatchReport(ctx, arg, unpref, me)
				}
			},
		}
	},
}

func buildMustNotReport(ctx rule.RuleContext, arg *ast.Node, jestFnName string, me matcherEntry) {
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
		Description: fmt.Sprintf("%s should not match %s", jestFnName, patStr),
		Data: map[string]string{
			"jestFunctionName": jestFnName,
			"pattern":          patStr,
		},
	})
}

func buildMustMatchReport(ctx rule.RuleContext, arg *ast.Node, jestFnName string, me matcherEntry) {
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
		Description: fmt.Sprintf("%s should match %s", jestFnName, patStr),
		Data: map[string]string{
			"jestFunctionName": jestFnName,
			"pattern":          patStr,
		},
	})
}
