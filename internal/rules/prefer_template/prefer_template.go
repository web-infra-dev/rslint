package prefer_template

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/prefer-template
var PreferTemplateRule = rule.Rule{
	Name: "prefer-template",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Tracks concat chains already reported, keyed by the top binary
		// expression's position, so nested literals don't double-report.
		reported := map[int]bool{}

		checkForStringConcat := func(node *ast.Node) {
			if !utils.IsStringLiteralOrTemplate(node) {
				return
			}
			// Walk past the paren wrapper (if any) to find the logical parent.
			parent := ast.WalkUpParenthesizedExpressions(node.Parent)
			if !utils.IsPlusBinaryExpression(parent) {
				return
			}
			top := getTopConcatBinaryExpression(parent)
			if reported[top.Pos()] {
				return
			}
			reported[top.Pos()] = true

			if !hasNonStringLiteral(top) {
				return
			}

			msg := rule.RuleMessage{
				Id:          "unexpectedStringConcatenation",
				Description: "Unexpected string concatenation.",
			}

			// Octal / non-octal-decimal escape sequences can't be represented
			// in a template literal, so skip the autofix in that case.
			if hasOctalOrNonOctalDecimalEscape(ctx.SourceFile, top) {
				ctx.ReportNode(top, msg)
				return
			}

			fixed := toTemplateLiteral(ctx.SourceFile, top, "", "")
			ctx.ReportNodeWithFixes(top, msg, rule.RuleFixReplace(ctx.SourceFile, top, fixed))
		}

		return rule.RuleListeners{
			ast.KindStringLiteral:                 checkForStringConcat,
			ast.KindNoSubstitutionTemplateLiteral: checkForStringConcat,
			ast.KindTemplateExpression:            checkForStringConcat,
		}
	},
}

// getTopConcatBinaryExpression walks up through concatenations (transparently
// crossing `ParenthesizedExpression` wrappers) and returns the outermost
// concatenation.
func getTopConcatBinaryExpression(node *ast.Node) *ast.Node {
	for {
		p := ast.WalkUpParenthesizedExpressions(node.Parent)
		if !utils.IsPlusBinaryExpression(p) {
			return node
		}
		node = p
	}
}

// hasStringLiteral reports whether the concat subtree rooted at node contains
// any string literal.
func hasStringLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if utils.IsPlusBinaryExpression(node) {
		bin := node.AsBinaryExpression()
		return hasStringLiteral(bin.Right) || hasStringLiteral(bin.Left)
	}
	return utils.IsStringLiteralOrTemplate(node)
}

// hasNonStringLiteral reports whether the concat subtree rooted at node
// contains any operand that is not a string literal (i.e. something that
// actually benefits from template-literal conversion).
func hasNonStringLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if utils.IsPlusBinaryExpression(node) {
		bin := node.AsBinaryExpression()
		return hasNonStringLiteral(bin.Right) || hasNonStringLiteral(bin.Left)
	}
	return !utils.IsStringLiteralOrTemplate(node)
}

// startsWithTemplateCurly reports whether converting node to a template
// literal produces output that begins with a `${...}` interpolation. When
// true, the preceding operand can "absorb" extra text into that interpolation.
func startsWithTemplateCurly(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindBinaryExpression:
		return startsWithTemplateCurly(node.AsBinaryExpression().Left)
	case ast.KindTemplateExpression:
		// Head is empty iff the template starts directly with `${`.
		return node.AsTemplateExpression().Head.Text() == ""
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return false
	}
	// Any other expression becomes `${...}` wholesale, so it starts with curly.
	return true
}

// endsWithTemplateCurly reports whether converting node to a template literal
// produces output that ends with a `${...}` interpolation.
func endsWithTemplateCurly(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindBinaryExpression:
		return startsWithTemplateCurly(node.AsBinaryExpression().Right)
	case ast.KindTemplateExpression:
		spans := node.AsTemplateExpression().TemplateSpans.Nodes
		lastSpan := spans[len(spans)-1].AsTemplateSpan()
		// A trailing `${...}` leaves the final tail quasi empty.
		return lastSpan.Literal.Text() == ""
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return false
	}
	return true
}

// octalEscapePattern matches a raw source substring containing an octal
// (`\1`-`\7`, `\0[0-7]`) or non-octal-decimal (`\8`, `\9`, `\0[89]`) escape
// sequence. The `(?s)` flag makes `.` match newlines so `\\.` skips any
// escaped pair, including line continuations.
var octalEscapePattern = regexp.MustCompile(`(?s)^(?:[^\\]|\\.)*?\\(?:[1-9]|0[0-9])`)

// hasOctalOrNonOctalDecimalEscape reports whether any string literal inside
// the concat subtree contains an octal or non-octal-decimal escape — these
// cannot be represented in a template literal, so autofix must be skipped.
func hasOctalOrNonOctalDecimalEscape(sourceFile *ast.SourceFile, node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if utils.IsPlusBinaryExpression(node) {
		bin := node.AsBinaryExpression()
		return hasOctalOrNonOctalDecimalEscape(sourceFile, bin.Left) ||
			hasOctalOrNonOctalDecimalEscape(sourceFile, bin.Right)
	}
	if node.Kind == ast.KindStringLiteral {
		return octalEscapePattern.MatchString(utils.TrimmedNodeText(sourceFile, node))
	}
	return false
}

// gapText returns the characters in [from, to) with any syntactic tokens
// scrubbed out — only whitespace and comments are kept. It mirrors ESLint's
// behavior of ignoring token boundaries (including `(` / `)` that wrap an
// operand) when collecting the text between two positions.
func gapText(sourceFile *ast.SourceFile, from, to int) string {
	if from >= to {
		return ""
	}
	src := sourceFile.Text()
	var sb strings.Builder
	pos := from
	for pos < to {
		triviaEnd := scanner.SkipTrivia(src, pos)
		if triviaEnd >= to {
			sb.WriteString(src[pos:to])
			return sb.String()
		}
		sb.WriteString(src[pos:triviaEnd])
		// Skip past the token that starts at triviaEnd.
		tokRange := scanner.GetRangeOfTokenAtPosition(sourceFile, triviaEnd)
		pos = tokRange.End()
	}
	return sb.String()
}

// toTemplateLiteral recursively builds the template-literal replacement for
// a concat subtree. `textBeforeNode`/`textAfterNode` are the trivia fragments
// (whitespace + comments) that surrounded the `+` consumed by the caller;
// they get spliced into a flanking `${...}` when possible, otherwise they
// are rendered back into the output verbatim.
func toTemplateLiteral(sourceFile *ast.SourceFile, currentNode *ast.Node, textBeforeNode, textAfterNode string) string {
	node := ast.SkipParentheses(currentNode)

	if node.Kind == ast.KindStringLiteral {
		return stringLiteralToTemplate(sourceFile, node)
	}
	if node.Kind == ast.KindTemplateExpression || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		// A template literal is already valid output; emit it verbatim.
		return utils.TrimmedNodeText(sourceFile, node)
	}

	if utils.IsPlusBinaryExpression(node) && hasStringLiteral(node) {
		bin := node.AsBinaryExpression()
		src := sourceFile.Text()

		// Split the "+" neighborhood into the trivia before and after the
		// operator token, so comments are preserved in the output. Anchor on
		// the inner (paren-stripped) operands so any surrounding `(` / `)`
		// tokens are shed while the whitespace around them is kept — this
		// matches ESLint's `sourceCode.getTokensBetween`-based behavior.
		innerLeft := ast.SkipParentheses(bin.Left)
		innerRight := ast.SkipParentheses(bin.Right)
		plusStart := scanner.SkipTrivia(src, bin.OperatorToken.Pos())
		plusEnd := bin.OperatorToken.End()
		innerRightStart := scanner.SkipTrivia(src, innerRight.Pos())
		textBeforePlus := gapText(sourceFile, innerLeft.End(), plusStart)
		textAfterPlus := gapText(sourceFile, plusEnd, innerRightStart)

		if endsWithTemplateCurly(bin.Left) {
			// `...${X}` + Y --> `...${X  ...  }Y` (fuse Y into the trailing curly)
			leftPart := toTemplateLiteral(sourceFile, bin.Left, textBeforeNode, textBeforePlus+textAfterPlus)
			rightPart := toTemplateLiteral(sourceFile, bin.Right, "", textAfterNode)
			return leftPart[:len(leftPart)-1] + rightPart[1:]
		}
		if startsWithTemplateCurly(bin.Right) {
			// X + `${Y}...` --> `X${  ...  Y}...` (fuse X into the leading curly)
			leftPart := toTemplateLiteral(sourceFile, bin.Left, textBeforeNode, "")
			rightPart := toTemplateLiteral(sourceFile, bin.Right, textBeforePlus+textAfterPlus, textAfterNode)
			return leftPart[:len(leftPart)-1] + rightPart[1:]
		}

		// Neither side can host the surrounding trivia; keep the `+` and
		// emit two separate template literals.
		return toTemplateLiteral(sourceFile, bin.Left, textBeforeNode, "") +
			textBeforePlus + "+" + textAfterPlus +
			toTemplateLiteral(sourceFile, bin.Right, "", textAfterNode)
	}

	// Any other expression becomes a `${...}` interpolation with the
	// surrounding trivia placed inside the curly.
	return "`${" + textBeforeNode + utils.TrimmedNodeText(sourceFile, node) + textAfterNode + "}`"
}

// stringLiteralToTemplate rewrites a string literal as a template literal
// (wrapped in backticks), escaping `${` and backticks so their meaning is
// preserved.
func stringLiteralToTemplate(sourceFile *ast.SourceFile, node *ast.Node) string {
	raw := utils.TrimmedNodeText(sourceFile, node)
	if len(raw) < 2 {
		return "`" + raw + "`"
	}
	quote := raw[0]
	inner := raw[1 : len(raw)-1]

	// Escape template-sensitive chars, then drop the now-redundant escape in
	// front of the original quote character (e.g. `\'` -> `'` in a source
	// string that was single-quoted).
	escaped := escapeTemplateSpecials(inner)
	escaped = strings.ReplaceAll(escaped, "\\"+string(quote), string(quote))
	return "`" + escaped + "`"
}

// escapeTemplateSpecials adds backslashes in front of `${` and backticks so
// they are treated as literal text when placed inside a template literal.
// A special character is already escaped when preceded by an odd number of
// backslashes; in that case it is left untouched. This mirrors ESLint's
// regex-driven implementation.
func escapeTemplateSpecials(s string) string {
	var sb strings.Builder
	i := 0
	for i < len(s) {
		ch := s[i]
		if ch != '\\' && ch != '$' && ch != '`' {
			sb.WriteByte(ch)
			i++
			continue
		}

		// Count the run of consecutive backslashes at position i.
		j := i
		for j < len(s) && s[j] == '\\' {
			j++
		}
		backslashes := j - i

		// Case 1: `\*${` — escape `${` unless already escaped (odd count).
		if j+1 < len(s) && s[j] == '$' && s[j+1] == '{' {
			sb.WriteString(strings.Repeat("\\", backslashes))
			if backslashes%2 == 0 {
				sb.WriteString("\\${")
			} else {
				sb.WriteString("${")
			}
			i = j + 2
			continue
		}
		// Case 2: `\*\`` — escape the backtick unless already escaped.
		if j < len(s) && s[j] == '`' {
			sb.WriteString(strings.Repeat("\\", backslashes))
			if backslashes%2 == 0 {
				sb.WriteString("\\`")
			} else {
				sb.WriteString("`")
			}
			i = j + 1
			continue
		}
		// Case 3: backslash run followed by a non-special char — preserve the
		// escape pair as-is (e.g. `\n`, `\x27`).
		if backslashes > 0 {
			sb.WriteString(s[i:j])
			if j < len(s) {
				sb.WriteByte(s[j])
				i = j + 1
			} else {
				i = j
			}
			continue
		}
		// Case 4: bare `$` not followed by `{` — no escape needed.
		sb.WriteByte(ch)
		i++
	}
	return sb.String()
}
