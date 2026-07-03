package no_else_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	allowElseIf bool
}

func parseOptions(ruleOptions any) options {
	opts := options{allowElseIf: true}
	optsMap := utils.GetOptionsMap(ruleOptions)
	if optsMap == nil {
		return opts
	}
	if allowElseIf, ok := optsMap["allowElseIf"].(bool); ok {
		opts.allowElseIf = allowElseIf
	}
	return opts
}

var unexpectedMessage = rule.RuleMessage{
	Id:          "unexpected",
	Description: "Unnecessary 'else' after 'return'.",
}

// https://eslint.org/docs/latest/rules/no-else-return
var NoElseReturnRule = rule.Rule{
	Name: "no-else-return",
	Run: func(ctx rule.RuleContext, ruleOptions any) rule.RuleListeners {
		opts := parseOptions(ruleOptions)
		check := checkIfWithoutElse
		if !opts.allowElseIf {
			check = checkIfWithElse
		}

		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindIfStatement): func(node *ast.Node) {
				check(ctx, node)
			},
		}
	},
}

func checkForReturn(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindReturnStatement
}

func naiveHasReturn(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindBlock {
		block := node.AsBlock()
		if block == nil || block.Statements == nil || len(block.Statements.Nodes) == 0 {
			return false
		}
		return checkForReturn(block.Statements.Nodes[len(block.Statements.Nodes)-1])
	}
	return checkForReturn(node)
}

func hasElse(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIfStatement {
		return false
	}
	ifStmt := node.AsIfStatement()
	return ifStmt != nil && ifStmt.ThenStatement != nil && ifStmt.ElseStatement != nil
}

func checkForIf(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIfStatement || !hasElse(node) {
		return false
	}
	ifStmt := node.AsIfStatement()
	return naiveHasReturn(ifStmt.ThenStatement) && naiveHasReturn(ifStmt.ElseStatement)
}

func checkForReturnOrIf(node *ast.Node) bool {
	return checkForReturn(node) || checkForIf(node)
}

func alwaysReturns(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindBlock {
		block := node.AsBlock()
		if block == nil || block.Statements == nil {
			return false
		}
		for _, stmt := range block.Statements.Nodes {
			if checkForReturnOrIf(stmt) {
				return true
			}
		}
		return false
	}
	return checkForReturnOrIf(node)
}

func checkIfWithoutElse(ctx rule.RuleContext, node *ast.Node) {
	if !isStatementListParent(node.Parent) {
		return
	}

	consequents := []*ast.Node{}
	var alternate *ast.Node
	for current := node; current != nil && current.Kind == ast.KindIfStatement; {
		ifStmt := current.AsIfStatement()
		if ifStmt == nil || ifStmt.ElseStatement == nil {
			return
		}
		consequents = append(consequents, ifStmt.ThenStatement)
		alternate = ifStmt.ElseStatement
		current = ifStmt.ElseStatement
	}

	for _, consequent := range consequents {
		if !alwaysReturns(consequent) {
			return
		}
	}
	displayReport(ctx, alternate)
}

func checkIfWithElse(ctx rule.RuleContext, node *ast.Node) {
	if !isStatementListParent(node.Parent) {
		return
	}
	ifStmt := node.AsIfStatement()
	if ifStmt == nil || ifStmt.ElseStatement == nil {
		return
	}
	if alwaysReturns(ifStmt.ThenStatement) {
		displayReport(ctx, ifStmt.ElseStatement)
	}
}

func displayReport(ctx rule.RuleContext, elseNode *ast.Node) {
	if fixes := buildFixes(ctx, elseNode); len(fixes) > 0 {
		ctx.ReportNodeWithFixes(elseNode, unexpectedMessage, fixes...)
		return
	}
	ctx.ReportNode(elseNode, unexpectedMessage)
}

func isStatementListParent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock,
		ast.KindCaseClause, ast.KindDefaultClause, ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
}

func buildFixes(ctx rule.RuleContext, elseNode *ast.Node) []rule.RuleFix {
	if elseNode == nil || elseNode.Parent == nil || !isSafeFromNameCollisions(elseNode) {
		return nil
	}

	sf := ctx.SourceFile
	text := sf.Text()
	elseStart := findElseKeywordStart(sf, elseNode.Parent, elseNode)
	if elseStart < 0 {
		return nil
	}

	startToken, ok := utils.TokenAtOrAfter(sf, utils.TrimNodeTextRange(sf, elseNode).Pos())
	if !ok {
		return nil
	}
	firstTokenOfElseBlock := startToken
	if startToken.Text == "{" {
		var found bool
		firstTokenOfElseBlock, found = utils.TokenAtOrAfter(sf, startToken.End)
		if !found {
			return nil
		}
	}

	lastIfToken, ok := utils.PreviousTokenBefore(sf, elseNode.Parent, elseStart)
	if !ok {
		return nil
	}

	ifStmt := elseNode.Parent.AsIfStatement()
	if ifStmt == nil || ifStmt.ThenStatement == nil {
		return nil
	}

	ifBlockMaybeUnsafe := ifStmt.ThenStatement.Kind != ast.KindBlock && lastIfToken.Text != ";"
	elseBlockUnsafe := startsUnsafeForASI(firstTokenOfElseBlock.Text)
	if ifBlockMaybeUnsafe && elseBlockUnsafe {
		return nil
	}

	elseTokens := utils.TokensOfNode(sf, elseNode)
	if len(elseTokens) == 0 {
		return nil
	}
	endToken := elseTokens[len(elseTokens)-1]
	lastTokenOfElseBlock := endToken
	if len(elseTokens) > 1 {
		lastTokenOfElseBlock = elseTokens[len(elseTokens)-2]
	}
	if lastTokenOfElseBlock.Text != ";" {
		if nextToken, found := utils.TokenAtOrAfter(sf, endToken.End); found {
			nextTokenUnsafe := startsUnsafeForASI(nextToken.Text)
			nextTokenOnSameLine := utils.IsSameLine(sf, nextToken.Start, lastTokenOfElseBlock.Start)
			if nextTokenUnsafe || (nextTokenOnSameLine && nextToken.Text != "}") {
				return nil
			}
		}
	}

	fixedSource := utils.TrimmedNodeText(sf, elseNode)
	if startToken.Text == "{" {
		trimmed := utils.TrimNodeTextRange(sf, elseNode)
		if trimmed.End()-trimmed.Pos() >= 2 {
			fixedSource = text[trimmed.Pos()+1 : trimmed.End()-1]
		}
	}

	retainStart := enclosingFunctionStart(sf, elseNode)
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(retainStart, retainStart), ""),
		rule.RuleFixReplaceRange(core.NewTextRange(elseStart, elseNode.End()), fixedSource),
	}
}

func enclosingFunctionStart(sf *ast.SourceFile, node *ast.Node) int {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(current) {
			return utils.TrimNodeTextRange(sf, current).Pos()
		}
	}
	return 0
}

func startsUnsafeForASI(text string) bool {
	if text == "" {
		return false
	}
	switch text[0] {
	case '(', '[', '/', '+', '`', '-':
		return true
	}
	return false
}

func findElseKeywordStart(sf *ast.SourceFile, ifNode, elseNode *ast.Node) int {
	elseNodeStart := utils.TrimNodeTextRange(sf, elseNode).Pos()
	elseStart := -1
	for _, token := range utils.TokensOfNode(sf, ifNode) {
		if token.Kind == ast.KindElseKeyword && token.End <= elseNodeStart {
			elseStart = token.Start
		}
	}
	return elseStart
}

func isSafeFromNameCollisions(elseNode *ast.Node) bool {
	if elseNode.Kind == ast.KindFunctionDeclaration {
		return false
	}
	if elseNode.Kind != ast.KindBlock {
		return true
	}

	names := directBlockScopedNames(elseNode)
	if len(names) == 0 {
		return true
	}

	target := elseNode.Parent
	if target != nil {
		target = target.Parent
	}
	if target == nil {
		return true
	}

	for name := range names {
		if hasConflictingDeclaration(target, name) || hasUnsafeReference(target, elseNode, name) {
			return false
		}
	}
	return true
}

func directBlockScopedNames(blockNode *ast.Node) map[string]struct{} {
	names := map[string]struct{}{}
	block := blockNode.AsBlock()
	if block == nil || block.Statements == nil {
		return names
	}
	for _, stmt := range block.Statements.Nodes {
		collectDirectBlockScopedStatementNames(stmt, names)
	}
	return names
}

func collectDirectBlockScopedStatementNames(stmt *ast.Node, names map[string]struct{}) {
	if stmt == nil {
		return
	}
	switch stmt.Kind {
	case ast.KindVariableStatement:
		varStmt := stmt.AsVariableStatement()
		if varStmt == nil || varStmt.DeclarationList == nil ||
			varStmt.DeclarationList.Flags&ast.NodeFlagsBlockScoped == 0 {
			return
		}
		declList := varStmt.DeclarationList.AsVariableDeclarationList()
		if declList == nil || declList.Declarations == nil {
			return
		}
		for _, decl := range declList.Declarations.Nodes {
			if decl == nil || decl.Kind != ast.KindVariableDeclaration {
				continue
			}
			if name := decl.AsVariableDeclaration().Name(); name != nil {
				utils.CollectBindingNames(name, func(_ *ast.Node, name string) {
					names[name] = struct{}{}
				})
			}
		}
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration,
		ast.KindEnumDeclaration, ast.KindModuleDeclaration,
		ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration,
		ast.KindImportEqualsDeclaration:
		collectDeclarationName(stmt, names)
	}
}

func collectDeclarationName(decl *ast.Node, names map[string]struct{}) {
	if name := decl.Name(); name != nil && name.Kind == ast.KindIdentifier {
		names[name.Text()] = struct{}{}
	}
}

func hasConflictingDeclaration(target *ast.Node, name string) bool {
	if hasCatchBindingForTarget(target, name) {
		return true
	}
	if fn := functionOwningBody(target); fn != nil {
		if hasFunctionNameOrParameter(fn, name) {
			return true
		}
	}
	if utils.HasHoistedVarDeclaration(target, name) {
		return true
	}
	return hasDirectLexicalDeclaration(target, name)
}

func hasCatchBindingForTarget(target *ast.Node, name string) bool {
	if target == nil || target.Kind != ast.KindBlock || target.Parent == nil || target.Parent.Kind != ast.KindCatchClause {
		return false
	}
	catchClause := target.Parent.AsCatchClause()
	if catchClause == nil || catchClause.Block != target || catchClause.VariableDeclaration == nil {
		return false
	}
	varDecl := catchClause.VariableDeclaration.AsVariableDeclaration()
	return varDecl != nil && varDecl.Name() != nil && utils.HasNameInBindingPattern(varDecl.Name(), name)
}

func functionOwningBody(target *ast.Node) *ast.Node {
	if target == nil || target.Parent == nil {
		return nil
	}
	parent := target.Parent
	if ast.IsFunctionLikeDeclaration(parent) && parent.Body() == target {
		return parent
	}
	return nil
}

func hasFunctionNameOrParameter(fn *ast.Node, name string) bool {
	if fn.Kind == ast.KindFunctionDeclaration || fn.Kind == ast.KindFunctionExpression {
		if n := fn.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
			return true
		}
	}
	return utils.HasShadowingParameter(fn, name)
}

func hasDirectLexicalDeclaration(target *ast.Node, name string) bool {
	for _, stmt := range directStatements(target) {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindVariableStatement:
			varStmt := stmt.AsVariableStatement()
			if varStmt == nil || varStmt.DeclarationList == nil ||
				varStmt.DeclarationList.Flags&ast.NodeFlagsBlockScoped == 0 {
				continue
			}
			if utils.HasVarDeclListWithName(varStmt.DeclarationList, name) {
				return true
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration,
			ast.KindEnumDeclaration, ast.KindModuleDeclaration,
			ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration,
			ast.KindImportEqualsDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		case ast.KindImportDeclaration:
			for _, binding := range utils.GetImportBindingNodes(stmt) {
				if binding != nil && binding.Text() == name {
					return true
				}
			}
		}
	}
	return false
}

func directStatements(target *ast.Node) []*ast.Node {
	if target == nil {
		return nil
	}
	switch target.Kind {
	case ast.KindSourceFile:
		if sf := target.AsSourceFile(); sf != nil && sf.Statements != nil {
			return sf.Statements.Nodes
		}
	case ast.KindBlock:
		if block := target.AsBlock(); block != nil && block.Statements != nil {
			return block.Statements.Nodes
		}
	case ast.KindModuleBlock:
		if block := target.AsModuleBlock(); block != nil && block.Statements != nil {
			return block.Statements.Nodes
		}
	case ast.KindCaseClause, ast.KindDefaultClause:
		if clause := target.AsCaseOrDefaultClause(); clause != nil && clause.Statements != nil {
			return clause.Statements.Nodes
		}
	case ast.KindClassStaticBlockDeclaration:
		staticBlock := target.AsClassStaticBlockDeclaration()
		if staticBlock != nil && staticBlock.Body != nil {
			body := staticBlock.Body.AsBlock()
			if body != nil && body.Statements != nil {
				return body.Statements.Nodes
			}
		}
	}
	return nil
}

func hasUnsafeReference(target, elseNode *ast.Node, name string) bool {
	found := false
	var walk func(*ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil || node == elseNode {
			return false
		}
		if node.Kind == ast.KindIdentifier && node.Text() == name && !utils.IsNonReferenceIdentifier(node) {
			found = true
			return true
		}
		node.ForEachChild(func(child *ast.Node) bool {
			return walk(child)
		})
		return found
	}
	walk(target)
	return found
}
