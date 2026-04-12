package newline_after_import

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type ruleOptions struct {
	count            int
	exactCount       bool
	considerComments bool
}

func parseOptions(options any) ruleOptions {
	opts := ruleOptions{count: 1}
	optsMap := utils.GetOptionsMap(options)
	if optsMap != nil {
		if v, ok := optsMap["count"]; ok {
			if f, ok := v.(float64); ok && f >= 1 {
				opts.count = int(f)
			}
		}
		if v, ok := optsMap["exactCount"]; ok {
			if b, ok := v.(bool); ok {
				opts.exactCount = b
			}
		}
		if v, ok := optsMap["considerComments"]; ok {
			if b, ok := v.(bool); ok {
				opts.considerComments = b
			}
		}
	}
	return opts
}

func makeMessage(typ string, count int) rule.RuleMessage {
	s := ""
	if count > 1 {
		s = "s"
	}
	id := "newlineAfterImport"
	if typ == "require" {
		id = "newlineAfterRequire"
	}
	return rule.RuleMessage{
		Id:          id,
		Description: fmt.Sprintf("Expected %d empty line%s after %s statement not followed by another %s.", count, s, typ, typ),
	}
}

// NewlineAfterImportRule enforces a newline after import/require statements.
// See: https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/newline-after-import.md
var NewlineAfterImportRule = rule.Rule{
	Name: "import/newline-after-import",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// The linter visits SourceFile's children (not the SourceFile node itself),
		// so a KindSourceFile listener would never fire. Process top-level statements directly.
		sourceFile := ctx.SourceFile
		if sourceFile != nil && sourceFile.Statements != nil {
			statements := sourceFile.Statements.Nodes
			lineStarts := sourceFile.ECMALineMap()
			text := sourceFile.Text()

			checkImportDeclarations(ctx, statements, lineStarts, text, opts)
			checkRequireCalls(ctx, statements, lineStarts, text, opts)
		}

		return rule.RuleListeners{}
	},
}

// isImportStatement returns true for ImportDeclaration and non-export ImportEqualsDeclaration.
// "export import x = ..." is excluded because it re-exports rather than imports.
func isImportStatement(stmt *ast.Node) bool {
	switch stmt.Kind {
	case ast.KindImportDeclaration:
		return true
	case ast.KindImportEqualsDeclaration:
		return !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport)
	}
	return false
}

func checkImportDeclarations(ctx rule.RuleContext, statements []*ast.Node, lineStarts []core.TextPos, text string, opts ruleOptions) {
	for i, stmt := range statements {
		if !isImportStatement(stmt) {
			continue
		}

		// ESLint's checkImport checks considerComments before checking if nextNode
		// exists, so a trailing comment after the last import can still trigger a report.
		if opts.considerComments {
			if checkCommentGap(ctx, stmt, stmt.End(), lineStarts, text, opts, "import") {
				continue
			}
		}

		if i+1 >= len(statements) {
			continue
		}

		nextStmt := statements[i+1]

		if isImportStatement(nextStmt) {
			continue
		}

		checkForNewLine(ctx, stmt, nextStmt, lineStarts, text, opts, "import")
	}
}

// findLastTopLevelRequire returns the last (in document order) top-level require() call
// not nested inside a function, block, object literal, or decorator.
// Returns nil if no top-level require is found.
// This mirrors ESLint's level-tracking for require detection.
func findLastTopLevelRequire(node *ast.Node) *ast.Node {
	if ast.IsRequireCall(node, true /* requireStringLiteralLikeArgument */) {
		return node
	}
	var result *ast.Node
	node.ForEachChild(func(child *ast.Node) bool {
		switch child.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindBlock, ast.KindObjectLiteralExpression, ast.KindDecorator:
			return false // these create a new scope level in ESLint
		}
		if found := findLastTopLevelRequire(child); found != nil {
			result = found
			// Don't stop — keep scanning to find the last one in document order.
		}
		return false
	})
	return result
}

func checkRequireCalls(ctx rule.RuleContext, statements []*ast.Node, lineStarts []core.TextPos, text string, opts ruleOptions) {
	for i, stmt := range statements {
		requireCall := findLastTopLevelRequire(stmt)
		if requireCall == nil {
			continue
		}
		if i+1 >= len(statements) {
			continue
		}

		nextStmt := statements[i+1]

		// For requires, ESLint uses the require CALL's end line for the comment search
		// window, but the containing STATEMENT's end line for the gap calculation.
		if opts.considerComments {
			if checkCommentGap(ctx, stmt, requireCall.End(), lineStarts, text, opts, "require") {
				continue
			}
		}

		if findLastTopLevelRequire(nextStmt) != nil {
			continue
		}

		checkForNewLine(ctx, stmt, nextStmt, lineStarts, text, opts, "require")
	}
}

// getReportPos returns the position used for diagnostic reporting, following ESLint's convention:
//   - single-line node → trimmed start position (preserves original column)
//   - multi-line node  → column 0 of the node's end line
func getReportPos(node *ast.Node, text string, lineStarts []core.TextPos) int {
	trimmedPos := scanner.SkipTrivia(text, node.Pos())
	startLine := scanner.ComputeLineOfPosition(lineStarts, trimmedPos)
	endLine := scanner.ComputeLineOfPosition(lineStarts, node.End())

	if startLine == endLine {
		return trimmedPos
	}
	return int(lineStarts[endLine])
}

// checkForNewLine reports when the gap between node and nextNode doesn't satisfy the count/exactCount options.
func checkForNewLine(ctx rule.RuleContext, node, nextNode *ast.Node, lineStarts []core.TextPos, text string, opts ruleOptions, typ string) {
	nodeEndLine := scanner.ComputeLineOfPosition(lineStarts, node.End())
	nextNodeStart := scanner.SkipTrivia(text, nextNode.Pos())
	nextNodeStartLine := scanner.ComputeLineOfPosition(lineStarts, nextNodeStart)

	lineDiff := nextNodeStartLine - nodeEndLine
	expectedDiff := opts.count + 1

	if lineDiff < expectedDiff || (opts.exactCount && lineDiff != expectedDiff) {
		pos := getReportPos(node, text, lineStarts)
		reportRange := core.NewTextRange(pos, pos)
		msg := makeMessage(typ, opts.count)

		// When exactCount is set and there are too many blank lines, report without fix.
		if opts.exactCount && lineDiff > expectedDiff {
			ctx.ReportRange(reportRange, msg)
		} else {
			fix := rule.RuleFixInsertAfter(node, strings.Repeat("\n", expectedDiff-lineDiff))
			ctx.ReportRangeWithFixes(reportRange, msg, fix)
		}
	}
}

// checkCommentGap looks for a comment within count+1 lines after windowAnchorEnd.
// If found, reports when the gap between the statement (node) and the comment is too small.
// (following ESLint's commentAfterImport semantics: only "too few lines" is reported,
// not "too many" — even with exactCount).
// Returns true if a relevant comment was found (regardless of whether an error was reported).
//
// windowAnchorEnd controls the search window:
//   - For imports: stmt.End() (the import node IS the statement)
//   - For requires: requireCall.End() (ESLint uses the require call's end line for the window,
//     but the containing statement's end line for the gap calculation)
//
// It checks trailing comments first (same line as the node), then leading comments
// (on subsequent lines), because GetLeadingCommentRanges skips same-line comments
// (collecting starts false until a newline is encountered).
func checkCommentGap(ctx rule.RuleContext, node *ast.Node, windowAnchorEnd int, lineStarts []core.TextPos, text string, opts ruleOptions, typ string) bool {
	windowEndLine := scanner.ComputeLineOfPosition(lineStarts, windowAnchorEnd)
	stmtEndLine := scanner.ComputeLineOfPosition(lineStarts, node.End())

	nodeFactory := &ast.NodeFactory{}

	// Check trailing comments (same line as node end).
	for commentRange := range scanner.GetTrailingCommentRanges(nodeFactory, text, node.End()) {
		commentLine := scanner.ComputeLineOfPosition(lineStarts, commentRange.Pos())
		if commentLine >= windowEndLine && commentLine <= windowEndLine+opts.count+1 {
			return handleCommentInGap(ctx, node, lineStarts, text, opts, typ, commentLine, stmtEndLine)
		}
	}

	// Check leading comments (after the next newline).
	for commentRange := range scanner.GetLeadingCommentRanges(nodeFactory, text, node.End()) {
		commentLine := scanner.ComputeLineOfPosition(lineStarts, commentRange.Pos())
		if commentLine > windowEndLine+opts.count+1 {
			break // past the search window
		}
		if commentLine >= windowEndLine {
			return handleCommentInGap(ctx, node, lineStarts, text, opts, typ, commentLine, stmtEndLine)
		}
	}

	return false
}

// handleCommentInGap reports when the gap between the statement end and the comment is too small.
// Always returns true to indicate a comment was found.
func handleCommentInGap(ctx rule.RuleContext, node *ast.Node, lineStarts []core.TextPos, text string, opts ruleOptions, typ string, commentLine, stmtEndLine int) bool {
	lineDiff := commentLine - stmtEndLine
	expectedDiff := opts.count + 1

	if lineDiff < expectedDiff {
		pos := getReportPos(node, text, lineStarts)
		reportRange := core.NewTextRange(pos, pos)
		msg := makeMessage(typ, opts.count)
		fix := rule.RuleFixInsertAfter(node, strings.Repeat("\n", expectedDiff-lineDiff))
		ctx.ReportRangeWithFixes(reportRange, msg, fix)
	}
	return true
}
