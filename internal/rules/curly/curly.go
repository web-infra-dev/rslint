package curly

import (
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// curlyOptions mirrors ESLint's positional options: the first element selects
// the mode ("all" | "multi" | "multi-line" | "multi-or-nest"); the second, when
// present, is the literal "consistent".
type curlyOptions struct {
	multiOnly   bool
	multiLine   bool
	multiOrNest bool
	consistent  bool
}

func parseOptions(options any) curlyOptions {
	opts := curlyOptions{}

	var first, second string
	switch v := options.(type) {
	case string:
		first = v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				first = s
			}
		}
		if len(v) > 1 {
			if s, ok := v[1].(string); ok {
				second = s
			}
		}
	}

	switch first {
	case "multi":
		opts.multiOnly = true
	case "multi-line":
		opts.multiLine = true
	case "multi-or-nest":
		opts.multiOrNest = true
	}
	if second == "consistent" {
		opts.consistent = true
	}
	return opts
}

// expectation is a tri-state mirroring ESLint's `expected` (true / false / null).
type expectation int

const (
	expectDontCare expectation = iota
	expectBraces
	expectNoBraces
)

// preparedCheck mirrors the object returned by ESLint's prepareCheck: whether
// the body currently is a block (actual) and whether it should be (expected).
type preparedCheck struct {
	node      *ast.Node
	body      *ast.Node
	name      string
	condition bool
	actual    bool
	expected  expectation
}

type curlyChecker struct {
	ctx     rule.RuleContext
	sf      *ast.SourceFile
	text    string
	lineMap []core.TextPos
	opts    curlyOptions
}

// CurlyRule enforces consistent brace usage on control statements.
// https://eslint.org/docs/latest/rules/curly
var CurlyRule = rule.Rule{
	Name: "curly",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.UnwrapOptions(_options)
		c := &curlyChecker{
			ctx:     ctx,
			sf:      ctx.SourceFile,
			text:    ctx.SourceFile.Text(),
			lineMap: ctx.SourceFile.ECMALineMap(),
			opts:    parseOptions(options),
		}

		return rule.RuleListeners{
			ast.KindIfStatement: c.checkIfStatement,
			ast.KindWhileStatement: func(node *ast.Node) {
				c.check(c.prepareCheck(node, node.AsWhileStatement().Statement, "while", true))
			},
			ast.KindDoStatement: func(node *ast.Node) {
				c.check(c.prepareCheck(node, node.AsDoStatement().Statement, "do", false))
			},
			ast.KindForStatement: func(node *ast.Node) {
				c.check(c.prepareCheck(node, node.AsForStatement().Statement, "for", true))
			},
			ast.KindForInStatement: func(node *ast.Node) {
				c.check(c.prepareCheck(node, node.AsForInOrOfStatement().Statement, "for-in", false))
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				c.check(c.prepareCheck(node, node.AsForInOrOfStatement().Statement, "for-of", false))
			},
		}
	},
}

func (c *curlyChecker) lineOf(pos int) int {
	return scanner.ComputeLineOfPosition(c.lineMap, pos)
}

// checkIfStatement handles the whole if / else-if / else chain when visiting the
// top `if`, and skips inner `else if`s (they are checked through the top `if`).
func (c *curlyChecker) checkIfStatement(node *ast.Node) {
	parent := node.Parent
	isElseIf := parent != nil && parent.Kind == ast.KindIfStatement &&
		parent.AsIfStatement().ElseStatement == node
	if isElseIf {
		return
	}
	for _, pc := range c.prepareIfChecks(node) {
		c.check(pc)
	}
}

func (c *curlyChecker) prepareIfChecks(node *ast.Node) []preparedCheck {
	checks := []preparedCheck{}

	for current := node; current != nil; current = current.AsIfStatement().ElseStatement {
		ifStmt := current.AsIfStatement()
		checks = append(checks, c.prepareCheck(current, ifStmt.ThenStatement, "if", true))
		alt := ifStmt.ElseStatement
		if alt != nil && alt.Kind != ast.KindIfStatement {
			checks = append(checks, c.prepareCheck(current, alt, "else", false))
			break
		}
	}

	if c.opts.consistent {
		// If any node should have, or already has, braces then they all must.
		anyExpected := false
		for _, pc := range checks {
			var v bool
			switch pc.expected {
			case expectBraces:
				v = true
			case expectNoBraces:
				v = false
			default:
				v = pc.actual
			}
			if v {
				anyExpected = true
				break
			}
		}
		target := expectNoBraces
		if anyExpected {
			target = expectBraces
		}
		for i := range checks {
			checks[i].expected = target
		}
	}

	return checks
}

func (c *curlyChecker) prepareCheck(node, body *ast.Node, name string, condition bool) preparedCheck {
	hasBlock := body.Kind == ast.KindBlock
	expected := expectDontCare

	switch {
	case hasBlock && (len(c.blockStatements(body)) != 1 || c.areBracesNecessary(body)):
		expected = expectBraces
	case c.opts.multiOnly:
		expected = expectNoBraces
	case c.opts.multiLine:
		if !c.isCollapsedOneLiner(body) {
			expected = expectBraces
		}
		// otherwise the body may or may not have braces
	case c.opts.multiOrNest:
		if hasBlock {
			statement := c.blockStatements(body)[0]
			leadingComments := c.hasLeadingComments(body, statement)
			if !c.isOneLiner(statement) || leadingComments {
				expected = expectBraces
			} else {
				expected = expectNoBraces
			}
		} else {
			if c.isOneLiner(body) {
				expected = expectNoBraces
			} else {
				expected = expectBraces
			}
		}
	default:
		// "all"
		expected = expectBraces
	}

	return preparedCheck{
		node:      node,
		body:      body,
		name:      name,
		condition: condition,
		actual:    hasBlock,
		expected:  expected,
	}
}

func (c *curlyChecker) check(pc preparedCheck) {
	if pc.expected == expectDontCare {
		return
	}
	wantBraces := pc.expected == expectBraces
	if wantBraces == pc.actual {
		return
	}

	bodyRange := utils.TrimNodeTextRange(c.sf, pc.body)

	if wantBraces {
		fixText := "{" + c.text[bodyRange.Pos():bodyRange.End()] + "}"
		c.ctx.ReportRangeWithFixes(
			bodyRange,
			c.message(true, pc.condition, pc.name),
			rule.RuleFixReplaceRange(bodyRange, fixText),
		)
		return
	}

	// Unnecessary braces: body is a block statement.
	if fix, ok := c.buildRemoveBracesFix(pc.node, pc.body, bodyRange); ok {
		c.ctx.ReportRangeWithFixes(bodyRange, c.message(false, pc.condition, pc.name), fix)
	} else {
		c.ctx.ReportRange(bodyRange, c.message(false, pc.condition, pc.name))
	}
}

func (c *curlyChecker) message(missing, condition bool, name string) rule.RuleMessage {
	var id, desc string
	switch {
	case missing && condition:
		id, desc = "missingCurlyAfterCondition", "Expected { after '"+name+"' condition."
	case missing:
		id, desc = "missingCurlyAfter", "Expected { after '"+name+"'."
	case condition:
		id, desc = "unexpectedCurlyAfterCondition", "Unnecessary { after '"+name+"' condition."
	default:
		id, desc = "unexpectedCurlyAfter", "Unnecessary { after '"+name+"'."
	}
	return rule.RuleMessage{Id: id, Description: desc, Data: map[string]string{"name": name}}
}

// buildRemoveBracesFix builds the fix that strips the braces from a block body.
// It returns ok=false when removing the braces would change semantics due to
// ASI, in which case the diagnostic is reported without a fix.
func (c *curlyChecker) buildRemoveBracesFix(node, body *ast.Node, bodyRange core.TextRange) (rule.RuleFix, bool) {
	openBraceStart := scanner.SkipTrivia(c.text, body.Pos()) // `{`
	closeBracePos := body.End() - 1                          // `}`

	if c.needsSemicolon(closeBracePos, openBraceStart+1) {
		return rule.RuleFix{}, false
	}

	resultingBodyText := c.text[openBraceStart+1 : closeBracePos]
	prefix := ""
	if c.needsPrecedingSpace(node, body, openBraceStart) {
		prefix = " "
	}
	return rule.RuleFixReplaceRange(bodyRange, prefix+resultingBodyText), true
}

// needsPrecedingSpace mirrors the `do{...}` special case: removing the braces
// from a `do` body that is glued to the `do` keyword can fuse the keyword with
// the first inner token (e.g. `do{foo()}` would lose the token boundary between
// `do` and the call), so a space is required.
func (c *curlyChecker) needsPrecedingSpace(node, body *ast.Node, openBraceStart int) bool {
	if node.Kind != ast.KindDoStatement {
		return false
	}
	// `do` is adjacent to `{` only when there is no trivia before the block.
	if openBraceStart != body.Pos() {
		return false
	}
	firstContent := scanner.SkipTrivia(c.text, openBraceStart+1)
	if firstContent >= len(c.text) {
		return false
	}
	r, _ := utf8.DecodeRuneInString(c.text[firstContent:])
	return scanner.IsIdentifierPart(r)
}

// needsSemicolon reports whether removing the braces would require a semicolon
// to be inserted to preserve semantics (ASI hazard). Mirrors ESLint's
// needsSemicolon; the curly fixer treats a true result as "don't fix".
func (c *curlyChecker) needsSemicolon(closeBracePos, contentStart int) bool {
	tbStart, tbEnd, tbKind := c.lastTokenBefore(contentStart, closeBracePos)

	taStart := scanner.SkipTrivia(c.text, closeBracePos+1)
	hasAfter := taStart < len(c.text) &&
		scanner.ScanTokenAtPosition(c.sf, taStart) != ast.KindEndOfFile

	if tbKind == ast.KindSemicolonToken {
		// The last statement already has a semicolon.
		return false
	}
	if !hasAfter {
		// No statements after this block.
		return false
	}

	// If the last node surrounded by braces is itself a BlockStatement (other
	// than a function/arrow body), removing the braces can't break ASI.
	lastBlockNode := ast.GetNodeAtPosition(c.sf, tbStart, false)
	if lastBlockNode != nil && lastBlockNode.Kind == ast.KindBlock {
		p := lastBlockNode.Parent
		if p != nil && p.Kind != ast.KindFunctionExpression && p.Kind != ast.KindArrowFunction {
			return false
		}
	}

	if c.lineOf(tbEnd) == c.lineOf(taStart) {
		// The next token is on the same line.
		return true
	}
	// The next token starts with a character that would disrupt ASI.
	switch c.text[taStart] {
	case '(', '[', '/', '`', '+', '-':
		return true
	}
	// The last token is ++ or --.
	if tbKind == ast.KindPlusPlusToken || tbKind == ast.KindMinusMinusToken {
		return true
	}
	return false
}

// lastTokenBefore scans tokens in [fromPos, beforePos) and returns the last one.
func (c *curlyChecker) lastTokenBefore(fromPos, beforePos int) (start, end int, kind ast.Kind) {
	s := scanner.GetScannerForSourceFile(c.sf, fromPos)
	for s.TokenStart() < beforePos && s.Token() != ast.KindEndOfFile {
		start, end, kind = s.TokenStart(), s.TokenEnd(), s.Token()
		s.Scan()
	}
	return
}

// areBracesNecessary mirrors astUtils.areBracesNecessary: a single-statement
// block still needs its braces when the statement is a lexical declaration, or
// when it ends with an `if` that would capture a trailing `else`.
func (c *curlyChecker) areBracesNecessary(block *ast.Node) bool {
	statement := c.blockStatements(block)[0]
	return isLexicalDeclaration(statement) ||
		(hasUnsafeIf(statement) && c.isFollowedByElseKeyword(block))
}

func (c *curlyChecker) isFollowedByElseKeyword(block *ast.Node) bool {
	nextStart := scanner.SkipTrivia(c.text, block.End())
	if nextStart >= len(c.text) {
		return false
	}
	return scanner.ScanTokenAtPosition(c.sf, nextStart) == ast.KindElseKeyword
}

// isCollapsedOneLiner reports whether the body sits on the same line as the
// token that precedes it (its closing `)` / `do` / `else`).
func (c *curlyChecker) isCollapsedOneLiner(node *ast.Node) bool {
	// node.Pos() lands right after the preceding token, so its line is that
	// token's line.
	return c.lineOf(node.Pos()) == c.lineOf(c.lastNonSemiTokenEnd(node))
}

// isOneLiner reports whether the node spans a single line, ignoring a trailing
// semicolon.
func (c *curlyChecker) isOneLiner(node *ast.Node) bool {
	if node.Kind == ast.KindEmptyStatement {
		return true
	}
	start := scanner.SkipTrivia(c.text, node.Pos())
	return c.lineOf(start) == c.lineOf(c.lastNonSemiTokenEnd(node))
}

// lastNonSemiTokenEnd returns the end position of the node's last token,
// excluding a trailing semicolon.
func (c *curlyChecker) lastNonSemiTokenEnd(node *ast.Node) int {
	start := scanner.SkipTrivia(c.text, node.Pos())
	trimmedEnd := utils.TrimNodeTextRange(c.sf, node).End()
	// Fast path: when the whole node is on one line, a trailing semicolon
	// can't move the end onto another line.
	if c.lineOf(start) == c.lineOf(trimmedEnd) {
		return trimmedEnd
	}
	s := scanner.GetScannerForSourceFile(c.sf, start)
	end := start
	for s.TokenStart() < node.End() && s.Token() != ast.KindEndOfFile {
		if s.Token() != ast.KindSemicolonToken {
			end = s.TokenEnd()
		}
		s.Scan()
	}
	return end
}

// hasLeadingComments reports whether there are comments between the block's
// opening `{` and its first statement (ESLint's getCommentsBefore(statement)).
func (c *curlyChecker) hasLeadingComments(block, statement *ast.Node) bool {
	openBraceStart := scanner.SkipTrivia(c.text, block.Pos())
	stmtStart := scanner.SkipTrivia(c.text, statement.Pos())
	return utils.HasCommentsInRange(c.sf, core.NewTextRange(openBraceStart+1, stmtStart))
}

func (c *curlyChecker) blockStatements(block *ast.Node) []*ast.Node {
	return block.AsBlock().Statements.Nodes
}

// isLexicalDeclaration mirrors astUtils.isLexicalDeclaration: let/const/using/
// await using variable declarations and function/class declarations.
//
// NOTE: Unlike ESLint (which only sees JavaScript), tsgo also parses TypeScript
// declarations that are equally illegal as an unbraced control-statement body
// (`if (a) enum E {}` is a syntax error). They are treated the same as lexical
// declarations so the autofix never strips a required block.
func isLexicalDeclaration(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindVariableStatement:
		declList := node.AsVariableStatement().DeclarationList
		return declList.Flags&ast.NodeFlagsBlockScoped != 0
	case ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindImportEqualsDeclaration:
		return true
	}
	return false
}

// hasUnsafeIf reports whether the code contains an `if` that would become
// associated with an `else` appended directly after it.
func hasUnsafeIf(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt.ElseStatement == nil {
			return true
		}
		return hasUnsafeIf(ifStmt.ElseStatement)
	case ast.KindForStatement:
		return hasUnsafeIf(node.AsForStatement().Statement)
	case ast.KindForInStatement, ast.KindForOfStatement:
		return hasUnsafeIf(node.AsForInOrOfStatement().Statement)
	case ast.KindLabeledStatement:
		return hasUnsafeIf(node.AsLabeledStatement().Statement)
	case ast.KindWithStatement:
		return hasUnsafeIf(node.AsWithStatement().Statement)
	case ast.KindWhileStatement:
		return hasUnsafeIf(node.AsWhileStatement().Statement)
	default:
		return false
	}
}
