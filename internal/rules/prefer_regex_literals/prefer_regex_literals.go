package prefer_regex_literals

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var safeRegexLiteralContentRe = regexp.MustCompile("^[-\\w\\\\\\[\\](){} \\t\\r\\n\\v\\f!@#$%^&*+=/~`.><?,'\"|:;]*$")

var validPrecedingTokens = map[string]bool{
	"(": true, ";": true, "[": true, ",": true, "=": true, "+": true,
	"*": true, "-": true, "?": true, "~": true, "%": true, "**": true,
	"!": true, "typeof": true, "instanceof": true, "&&": true, "||": true,
	"??": true, "return": true, "...": true, "delete": true, "void": true,
	"in": true, "<": true, ">": true, "<=": true, ">=": true, "==": true,
	"===": true, "!=": true, "!==": true, "<<": true, ">>": true,
	">>>": true, "&": true, "|": true, "^": true, ":": true, "{": true,
	"=>": true, "*=": true, "<<=": true, ">>=": true, ">>>=": true,
	"^=": true, "|=": true, "&=": true, "??=": true, "||=": true,
	"&&=": true, "**=": true, "+=": true, "-=": true, "/=": true,
	"%=": true, "/": true, "do": true, "break": true, "continue": true,
	"debugger": true, "case": true, "throw": true,
}

type options struct {
	disallowRedundantWrapping bool
}

// https://eslint.org/docs/latest/rules/prefer-regex-literals
var PreferRegexLiteralsRule = rule.Rule{
	Name: "prefer-regex-literals",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.LegacyUnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)

		check := func(node *ast.Node, callee *ast.Node, argsList *ast.NodeList) {
			if !isBuiltinRegExpCallee(ctx, utils.SkipAssertionsAndParens(callee)) {
				return
			}

			args := nodeListNodes(argsList)
			if opts.disallowRedundantWrapping && isUnnecessarilyWrappedRegexLiteral(ctx, args) {
				reportRedundantRegExp(ctx, node, args)
				return
			}

			if hasOnlyStaticStringArguments(ctx, args) {
				reportStaticStringRegExp(ctx, node, args)
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				check(node, call.Expression, call.Arguments)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				newExpr := node.AsNewExpression()
				check(node, newExpr.Expression, newExpr.Arguments)
			},
		}
	},
}

func parseOptions(rawOptions any) options {
	opts := options{}
	if optsMap := utils.GetOptionsMap(rawOptions); optsMap != nil {
		if v, ok := optsMap["disallowRedundantWrapping"].(bool); ok {
			opts.disallowRedundantWrapping = v
		}
	}
	return opts
}

func nodeListNodes(list *ast.NodeList) []*ast.Node {
	if list == nil {
		return nil
	}
	return list.Nodes
}

func hasOnlyStaticStringArguments(ctx rule.RuleContext, args []*ast.Node) bool {
	if len(args) != 1 && len(args) != 2 {
		return false
	}
	for _, arg := range args {
		if _, ok := staticStringValue(ctx, arg); !ok {
			return false
		}
	}
	return true
}

func isUnnecessarilyWrappedRegexLiteral(ctx rule.RuleContext, args []*ast.Node) bool {
	if len(args) == 1 {
		return isRegexLiteral(args[0])
	}
	if len(args) == 2 && isRegexLiteral(args[0]) {
		_, ok := staticStringValue(ctx, args[1])
		return ok
	}
	return false
}

func reportStaticStringRegExp(ctx rule.RuleContext, node *ast.Node, args []*ast.Node) {
	pattern, _ := staticStringValue(ctx, args[0])
	flags := ""
	if len(args) == 2 {
		flags, _ = staticStringValue(ctx, args[1])
	}

	msg := rule.RuleMessage{
		Id:          "unexpectedRegExp",
		Description: "Use a regular expression literal instead of the 'RegExp' constructor.",
	}

	suggestions := []rule.RuleSuggestion{}
	if literal, ok := buildLiteralReplacement(pattern, flags); ok && canFixTo(ctx, node, literal) {
		suggestions = append(suggestions, buildSuggestion(ctx, node, "replaceWithLiteral", "Replace with an equivalent regular expression literal.", literal, nil))
	}
	reportWithSuggestions(ctx, node, msg, suggestions)
}

func reportRedundantRegExp(ctx rule.RuleContext, node *ast.Node, args []*ast.Node) {
	regexNode := utils.SkipAssertionsAndParens(args[0])
	pattern, literalFlags, _ := regexLiteralPatternAndFlags(regexNode)

	if len(args) == 2 {
		argFlags, _ := staticStringValue(ctx, args[1])
		suggestions := []rule.RuleSuggestion{}

		replacement := "/" + pattern + "/" + argFlags
		if canFixTo(ctx, node, replacement) {
			suggestions = append(suggestions, buildSuggestion(
				ctx,
				node,
				"replaceWithLiteralAndFlags",
				fmt.Sprintf("Replace with an equivalent regular expression literal with flags '%s'.", argFlags),
				replacement,
				map[string]string{"flags": argFlags},
			))
		}

		mergedFlags := mergeRegexFlags(literalFlags, argFlags)
		mergedReplacement := "/" + pattern + "/" + mergedFlags
		if !areFlagsEqual(mergedFlags, argFlags) && canFixTo(ctx, node, mergedReplacement) {
			suggestions = append(suggestions, buildSuggestion(
				ctx,
				node,
				"replaceWithIntendedLiteralAndFlags",
				fmt.Sprintf("Replace with a regular expression literal with flags '%s'.", mergedFlags),
				mergedReplacement,
				map[string]string{"flags": mergedFlags},
			))
		}

		reportWithSuggestions(ctx, node, rule.RuleMessage{
			Id:          "unexpectedRedundantRegExpWithFlags",
			Description: "Use regular expression literal with flags instead of the 'RegExp' constructor.",
		}, suggestions)
		return
	}

	suggestions := []rule.RuleSuggestion{}
	literal := utils.TrimmedNodeText(ctx.SourceFile, regexNode)
	if canFixTo(ctx, node, literal) {
		suggestions = append(suggestions, buildSuggestion(
			ctx,
			node,
			"replaceWithLiteral",
			"Replace with an equivalent regular expression literal.",
			literal,
			nil,
		))
	}

	reportWithSuggestions(ctx, node, rule.RuleMessage{
		Id:          "unexpectedRedundantRegExp",
		Description: "Regular expression literal is unnecessarily wrapped within a 'RegExp' constructor.",
	}, suggestions)
}

func reportWithSuggestions(ctx rule.RuleContext, node *ast.Node, msg rule.RuleMessage, suggestions []rule.RuleSuggestion) {
	if len(suggestions) == 0 {
		ctx.ReportNode(node, msg)
		return
	}
	ctx.ReportNodeWithSuggestions(node, msg, suggestions...)
}

func buildSuggestion(ctx rule.RuleContext, node *ast.Node, messageID string, description string, replacement string, data map[string]string) rule.RuleSuggestion {
	return rule.RuleSuggestion{
		Message: rule.RuleMessage{
			Id:          messageID,
			Description: description,
			Data:        data,
		},
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplaceRange(utils.TrimNodeTextRange(ctx.SourceFile, node), getSafeOutput(ctx.SourceFile, node, replacement)),
		},
	}
}

func buildLiteralReplacement(pattern string, flags string) (string, bool) {
	if !safeRegexLiteralContentRe.MatchString(pattern) {
		return "", false
	}
	if pattern == "" {
		pattern = "(?:)"
	}
	return "/" + escapeRegexLiteralContent(pattern) + "/" + flags, true
}

func escapeRegexLiteralContent(pattern string) string {
	var out strings.Builder
	for i := 0; i < len(pattern); {
		r, w := utf8.DecodeRuneInString(pattern[i:])
		if r == '\\' && i+w < len(pattern) {
			next, nextW := utf8.DecodeRuneInString(pattern[i+w:])
			if escaped := escapedLineContinuation(next); escaped != "" {
				out.WriteString(escaped)
				i += w + nextW
				continue
			}
		}
		switch r {
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		case '\v':
			out.WriteString(`\v`)
		case '\f':
			out.WriteString(`\f`)
		case '/':
			out.WriteString(`\/`)
		default:
			out.WriteString(pattern[i : i+w])
		}
		i += w
	}
	return out.String()
}

func escapedLineContinuation(r rune) string {
	switch r {
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	case '\v':
		return `\v`
	case '\f':
		return `\f`
	default:
		return ""
	}
}

// staticStringValue intentionally stays syntactic: ESLint only treats literal
// strings, static templates, and String.raw static templates as reportable here.
func staticStringValue(ctx rule.RuleContext, node *ast.Node) (string, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text, true
	case ast.KindTaggedTemplateExpression:
		return stringRawStaticTemplateValue(ctx, node)
	default:
		return "", false
	}
}

func stringRawStaticTemplateValue(ctx rule.RuleContext, node *ast.Node) (string, bool) {
	tagged := node.AsTaggedTemplateExpression()
	if tagged == nil || tagged.Tag == nil || tagged.Template == nil {
		return "", false
	}
	if !isStringRawTag(ctx, tagged.Tag) {
		return "", false
	}
	template := ast.SkipParentheses(tagged.Template)
	if template == nil || template.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return "", false
	}
	text := utils.TrimmedNodeText(ctx.SourceFile, template)
	if len(text) >= 2 && text[0] == '`' && text[len(text)-1] == '`' {
		return text[1 : len(text)-1], true
	}
	return template.AsNoSubstitutionTemplateLiteral().Text, true
}

func isStringRawTag(ctx rule.RuleContext, tag *ast.Node) bool {
	tag = utils.SkipAssertionsAndParens(tag)
	if tag == nil {
		return false
	}

	var object *ast.Node
	var property string
	switch tag.Kind {
	case ast.KindPropertyAccessExpression:
		access := tag.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return false
		}
		object = access.Expression
		property = access.Name().AsIdentifier().Text
	case ast.KindElementAccessExpression:
		access := tag.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return false
		}
		object = access.Expression
		value, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(access.ArgumentExpression))
		if !ok {
			return false
		}
		property = value
	default:
		return false
	}

	if property != "raw" {
		return false
	}
	object = utils.SkipAssertionsAndParens(object)
	return object != nil &&
		object.Kind == ast.KindIdentifier &&
		object.AsIdentifier().Text == "String" &&
		!utils.IsShadowed(object, "String")
}

func isRegexLiteral(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	return node != nil && node.Kind == ast.KindRegularExpressionLiteral
}

func regexLiteralPatternAndFlags(node *ast.Node) (string, string, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindRegularExpressionLiteral {
		return "", "", false
	}
	pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
	return pattern, flags, true
}

func isBuiltinRegExpCallee(ctx rule.RuleContext, callee *ast.Node) bool {
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		if callee.AsIdentifier().Text != "RegExp" {
			return false
		}
		return !utils.IsShadowed(callee, "RegExp")
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return false
		}
		if access.Name().AsIdentifier().Text != "RegExp" {
			return false
		}
		return isKnownGlobalObject(access.Expression)
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return false
		}
		value, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(access.ArgumentExpression))
		if !ok || value != "RegExp" {
			return false
		}
		return isKnownGlobalObject(access.Expression)
	}

	return false
}

func isKnownGlobalObject(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	name := node.AsIdentifier().Text
	switch name {
	case "globalThis", "window", "self", "global":
		return !utils.IsShadowed(node, name)
	default:
		return false
	}
}

func canFixTo(ctx rule.RuleContext, node *ast.Node, literal string) bool {
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	if utils.HasCommentInsideNode(ctx.SourceFile, node) {
		return false
	}
	if before, ok := tokenBefore(ctx.SourceFile, nodeRange.Pos()); ok {
		if !validPrecedingTokens[before.text] {
			return false
		}
	}
	return utils.IsValidRegexLiteral(literal)
}

func areFlagsEqual(flagsA string, flagsB string) bool {
	a := []rune(flagsA)
	b := []rune(flagsB)
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
	return slices.Equal(a, b)
}

func mergeRegexFlags(flagsA string, flagsB string) string {
	seen := map[rune]bool{}
	var out strings.Builder
	for _, flags := range []string{flagsA, flagsB} {
		for _, flag := range flags {
			if !seen[flag] {
				seen[flag] = true
				out.WriteRune(flag)
			}
		}
	}
	return out.String()
}

func getSafeOutput(sourceFile *ast.SourceFile, node *ast.Node, replacement string) string {
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	output := replacement

	if before, ok := tokenBefore(sourceFile, nodeRange.Pos()); ok &&
		before.end == nodeRange.Pos() &&
		!canTokensBeAdjacent(before.text, output) {
		output = " " + output
	}

	if after, ok := tokenAfter(sourceFile, nodeRange.End()); ok &&
		nodeRange.End() == after.start &&
		!canTokensBeAdjacent(output, after.text) {
		output += " "
	}

	return output
}

type tokenInfo struct {
	kind  ast.Kind
	text  string
	start int
	end   int
}

func tokenBefore(sourceFile *ast.SourceFile, pos int) (tokenInfo, bool) {
	s := scanner.GetScannerForSourceFile(sourceFile, 0)
	var last tokenInfo
	found := false
	for s.Token() != ast.KindEndOfFile && s.TokenStart() < pos {
		if s.TokenEnd() <= pos {
			last = tokenInfo{
				kind:  s.Token(),
				text:  tokenText(sourceFile, s.Token(), s.TokenStart(), s.TokenEnd()),
				start: s.TokenStart(),
				end:   s.TokenEnd(),
			}
			found = true
		}
		s.Scan()
	}
	return last, found
}

func tokenAfter(sourceFile *ast.SourceFile, pos int) (tokenInfo, bool) {
	start := scanner.SkipTrivia(sourceFile.Text(), pos)
	s := scanner.GetScannerForSourceFile(sourceFile, start)
	if s.Token() == ast.KindEndOfFile {
		return tokenInfo{}, false
	}
	return tokenInfo{
		kind:  s.Token(),
		text:  tokenText(sourceFile, s.Token(), s.TokenStart(), s.TokenEnd()),
		start: s.TokenStart(),
		end:   s.TokenEnd(),
	}, true
}

func tokenText(sourceFile *ast.SourceFile, kind ast.Kind, start int, end int) string {
	if start >= 0 && end <= len(sourceFile.Text()) && start < end {
		return sourceFile.Text()[start:end]
	}
	if text := scanner.TokenToString(kind); text != "" {
		return text
	}
	return kind.String()
}

func canTokensBeAdjacent(left string, right string) bool {
	if left == "" || right == "" {
		return true
	}
	if strings.HasSuffix(left, "/") && strings.HasPrefix(right, "/") {
		return false
	}
	if strings.HasSuffix(left, "/") && startsIdentifierLike(right) {
		return false
	}
	if endsIdentifierLike(left) && startsIdentifierLike(right) {
		return false
	}
	return true
}

func startsIdentifierLike(text string) bool {
	r, _ := utf8.DecodeRuneInString(text)
	return r != utf8.RuneError && (scanner.IsIdentifierStart(r) || unicode.IsLetter(r))
}

func endsIdentifierLike(text string) bool {
	r, _ := utf8.DecodeLastRuneInString(text)
	return r != utf8.RuneError && scanner.IsIdentifierPart(r)
}
