package valid_title

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var matcherPatternSchema = rule.Union(
	rule.String(),
	rule.Array(rule.String()),
	rule.Object(map[string]rule.Schema{
		"describe": rule.Union(rule.String(), rule.Array(rule.String())).Default(nil),
		"test":     rule.Union(rule.String(), rule.Array(rule.String())).Default(nil),
		"it":       rule.Union(rule.String(), rule.Array(rule.String())).Default(nil),
	}),
).Default(nil)

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
	invalidPatterns          []invalidPattern
	mustNotMatch             matchersByFn
	mustMatch                matchersByFn
}

type invalidPattern struct {
	optionPath string
	pattern    string
	err        error
}

func compileRE2(pat string) (*regexp2.Regexp, error) {
	re, err := regexp2.Compile(pat, jsRegexOpts)
	if err != nil {
		return nil, err
	}
	return re, nil
}

func matchRE2(re *regexp2.Regexp, s string) bool {
	if re == nil {
		return false
	}
	m, err := re.FindStringMatch(s)
	return err == nil && m != nil
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
	if len(options) == 0 {
		return compiledOptions{}
	}
	m, ok := options[0].(map[string]any)
	if !ok || m == nil {
		return compiledOptions{}
	}

	co := compiledOptions{
		ignoreSpaces:             rule.Must[bool](m["ignoreSpaces"]),
		ignoreTypeOfDescribeName: rule.Must[bool](m["ignoreTypeOfDescribeName"]),
		ignoreTypeOfTestName:     rule.Must[bool](m["ignoreTypeOfTestName"]),
	}

	if dw := rule.Must[[]any](m["disallowedWords"]); len(dw) > 0 {
		co.disallowedConcat, co.invalidPatterns = compileDisallowedWords(dw, co.invalidPatterns)
	}

	if mn := m["mustNotMatch"]; mn != nil {
		var invalids []invalidPattern
		co.mustNotMatch, invalids = compileMatcherPatterns(mn, "mustNotMatch")
		co.invalidPatterns = append(co.invalidPatterns, invalids...)
	}
	if mm := m["mustMatch"]; mm != nil {
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

func trimFXPrefix(word string) string {
	if word == "" {
		return ""
	}
	if word[0] == 'f' || word[0] == 'x' {
		return word[1:]
	}
	return word
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

func jestTitleInner(sourceFile *ast.SourceFile, n *ast.Node) (string, bool) {
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
func shouldValidateEachPrintf(jestFn *jestUtils.ParsedJestFnCall, call *ast.CallExpression) bool {
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
	reDupPrefix            = regexp.MustCompile(`^([\x60'"]).+? `)
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
	return "/" + strings.ReplaceAll(src, "/", "\\/") + "/u"
}

func eachInvalidSpecifier(title string) string {
	s := strings.ReplaceAll(title, "%%", "")
	return reEachInvalidSpecifier.FindString(s)
}

func jestEmptyFunctionName(kind jestUtils.JestFnType) string {
	if kind == jestUtils.JestFnTypeDescribe {
		return "describe"
	}
	return "test"
}

// ValidTitleRule enforces ESLint jest/valid-title.
var ValidTitleRule = rule.Rule{
	Name: "jest/valid-title",
	Schema: rule.Tuple(rule.Object(map[string]rule.Schema{
		"ignoreSpaces":             rule.Bool().Default(false),
		"ignoreTypeOfDescribeName": rule.Bool().Default(false),
		"ignoreTypeOfTestName":     rule.Bool().Default(false),
		"disallowedWords":          rule.Array(rule.String()).Default([]any{}),
		"mustMatch":                matcherPatternSchema,
		"mustNotMatch":             matcherPatternSchema,
	})),
	RunWithOptions: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFn := jestUtils.ParseJestFnCall(node, ctx)
				if jestFn == nil {
					return
				}
				if jestFn.Kind != jestUtils.JestFnTypeDescribe && jestFn.Kind != jestUtils.JestFnTypeTest {
					return
				}

				call := node.AsCallExpression()
				if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				arg := call.Arguments.Nodes[0]

				title, ok := jestTitleInner(ctx.SourceFile, arg)
				if !ok {
					if binaryExprContainsStringLit(arg) {
						return
					}
					ignored := false
					if jestFn.Kind == jestUtils.JestFnTypeDescribe && co.ignoreTypeOfDescribeName {
						ignored = true
					}
					if jestFn.Kind == jestUtils.JestFnTypeTest && co.ignoreTypeOfTestName {
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

				unprefixedName := trimFXPrefix(jestFn.Name)
				firstTok := title
				if i := strings.IndexByte(title, ' '); i >= 0 {
					firstTok = title[:i]
				}
				if strings.EqualFold(firstTok, unprefixedName) {
					raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, arg, false)
					fix := duplicatePrefixReplacement(raw)
					ctx.ReportNodeWithFixes(arg, rule.RuleMessage{
						Id:          "duplicatePrefix",
						Description: "should not have duplicate prefix",
					}, rule.RuleFixReplace(ctx.SourceFile, arg, fix))
				}

				fnKey := trimFXPrefix(jestFn.Name)

				if me := matcherFor(fnKey, co.mustNotMatch); matchRE2(me.re, title) {
					buildMustNotReport(ctx, arg, unprefixedName, me)
					return
				}

				me := matcherFor(fnKey, co.mustMatch)
				if me.re != nil && !matchRE2(me.re, title) {
					buildMustMatchReport(ctx, arg, unprefixedName, me)
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
