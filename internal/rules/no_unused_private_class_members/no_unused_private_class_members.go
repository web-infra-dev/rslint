package no_unused_private_class_members

import (
	"fmt"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type privateMember struct {
	declNode     *ast.Node
	nameNode     *ast.Node
	name         string
	hasReference bool
	isAccessor   bool
	isUsed       bool
}

type classMembers struct {
	classNode *ast.Node
	members   map[string]*privateMember
	order     []string
}

type sourceItem struct {
	start     int
	end       int
	text      string
	isComment bool
}

var NoUnusedPrivateClassMembersRule = rule.Rule{
	Name: "no-unused-private-class-members",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var stack []*classMembers
		var sourceItems []sourceItem

		getSourceItems := func() []sourceItem {
			if sourceItems == nil {
				sourceItems = collectSourceItems(ctx.SourceFile, ctx.Comments)
			}
			return sourceItems
		}

		pushClass := func(node *ast.Node) {
			stack = append(stack, collectPrivateMembers(node))
		}

		popClass := func(_ *ast.Node) {
			if len(stack) == 0 {
				return
			}
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			reportUnusedMembers(ctx, current, getSourceItems())
		}

		handlePrivateIdentifier := func(node *ast.Node) {
			if isPrivateMemberDeclarationName(node) {
				return
			}

			member := findTrackedPrivateMember(stack, node)
			if member == nil || member.isUsed {
				return
			}

			member.hasReference = true
			if member.isAccessor {
				member.isUsed = true
				return
			}

			if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
				if isWriteOnlyAccess(parent) {
					return
				}
			}

			member.isUsed = true
		}

		var scanComputedKey func(node *ast.Node)
		scanComputedKey = func(node *ast.Node) {
			if node == nil {
				return
			}
			if node.Kind == ast.KindPrivateIdentifier {
				handlePrivateIdentifier(node)
			}
			node.ForEachChild(func(child *ast.Node) bool {
				scanComputedKey(child)
				return false
			})
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration:                      pushClass,
			rule.ListenerOnExit(ast.KindClassDeclaration): popClass,
			ast.KindClassExpression:                       pushClass,
			rule.ListenerOnExit(ast.KindClassExpression):  popClass,
			ast.KindPrivateIdentifier:                     handlePrivateIdentifier,
			ast.KindPropertyAssignment: func(node *ast.Node) {
				property := node.AsPropertyAssignment()
				if property == nil {
					return
				}
				name := property.Name()
				if name == nil || name.Kind != ast.KindComputedPropertyName {
					return
				}
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindObjectLiteralExpression || !ast.IsAssignmentTarget(parent) {
					return
				}
				// The generic PrivateIdentifier listener can miss computed keys in
				// destructuring assignment patterns, so scan that key explicitly.
				scanComputedKey(name)
			},
		}
	},
}

func collectPrivateMembers(classNode *ast.Node) *classMembers {
	result := &classMembers{
		classNode: classNode,
		members:   map[string]*privateMember{},
	}

	for _, memberNode := range classNode.Members() {
		switch memberNode.Kind {
		case ast.KindPropertyDeclaration,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			nameNode := memberNode.Name()
			if nameNode == nil || nameNode.Kind != ast.KindPrivateIdentifier {
				continue
			}
			name := nameNode.AsPrivateIdentifier().Text
			if _, exists := result.members[name]; !exists {
				result.order = append(result.order, name)
			}
			result.members[name] = &privateMember{
				declNode:   memberNode,
				nameNode:   nameNode,
				name:       name,
				isAccessor: memberNode.Kind == ast.KindGetAccessor || memberNode.Kind == ast.KindSetAccessor,
			}
		}
	}

	return result
}

func isPrivateMemberDeclarationName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyDeclaration,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return parent.Name() == node
	default:
		return false
	}
}

func findTrackedPrivateMember(stack []*classMembers, node *ast.Node) *privateMember {
	name := node.AsPrivateIdentifier().Text
	for i := len(stack) - 1; i >= 0; i-- {
		if member, ok := stack[i].members[name]; ok {
			return member
		}
	}
	return nil
}

func isWriteOnlyAccess(memberExpr *ast.Node) bool {
	target := ast.GetAssignmentTarget(memberExpr)
	if target == nil {
		return false
	}

	switch target.Kind {
	case ast.KindBinaryExpression:
		binary := target.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			return true
		}
		return isExpressionStatementParent(target.Parent)
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		return isExpressionStatementParent(target.Parent)
	case ast.KindForInStatement, ast.KindForOfStatement:
		return true
	default:
		return false
	}
}

func isExpressionStatementParent(parent *ast.Node) bool {
	parent = ast.WalkUpParenthesizedExpressions(parent)
	return parent != nil && parent.Kind == ast.KindExpressionStatement
}

func reportUnusedMembers(ctx rule.RuleContext, classInfo *classMembers, items []sourceItem) {
	for _, name := range classInfo.order {
		member := classInfo.members[name]
		if member == nil || member.isUsed {
			continue
		}

		message := unusedPrivateClassMemberMessage(member.name)
		if member.hasReference {
			ctx.ReportNode(member.nameNode, message)
			continue
		}

		fixes := []rule.RuleFix{
			rule.RuleFixReplaceRange(memberRemovalRange(ctx.SourceFile, member.declNode, items), ""),
		}
		if token, ok := semicolonInsertionToken(ctx.SourceFile, classInfo.classNode, member.declNode, items); ok {
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(token.end, token.end), ";"))
		}

		ctx.ReportNodeWithSuggestions(member.nameNode, message, rule.RuleSuggestion{
			Message:  removeUnusedPrivateClassMemberMessage(member.name),
			FixesArr: fixes,
		})
	}
}

func unusedPrivateClassMemberMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unusedPrivateClassMember",
		Description: fmt.Sprintf("'%s' is defined but never used.", name),
		Data:        map[string]string{"classMemberName": name},
	}
}

func removeUnusedPrivateClassMemberMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeUnusedPrivateClassMember",
		Description: fmt.Sprintf("Remove unused private class member '%s'.", name),
		Data:        map[string]string{"classMemberName": name},
	}
}

func memberRemovalRange(sourceFile *ast.SourceFile, memberNode *ast.Node, items []sourceItem) core.TextRange {
	// Suggestions follow ESLint's comment-sensitive member removal behavior:
	// remove attached leading/trailing comments, but preserve unrelated comments
	// and neighboring members that share a line.
	memberTextRange := utils.TrimNodeTextRange(sourceFile, memberNode)
	memberItem := sourceItem{start: memberTextRange.Pos(), end: memberNode.End()}
	leadingComments := leadingCommentsForMember(sourceFile, memberItem, items)
	trailingComments := trailingCommentsForMember(sourceFile, memberItem, items)
	shouldRemoveLeadingComments := len(leadingComments) > 0 && !sharesLineWithAnotherToken(sourceFile, memberItem, items)
	lastItemToRemove := memberItem
	if len(trailingComments) > 0 {
		lastItemToRemove = trailingComments[len(trailingComments)-1]
	}

	previousToken, hasPreviousToken := itemBefore(items, memberItem.start, false)
	nextToken, hasNextToken := itemAfter(items, lastItemToRemove.end, true)
	nextTokenStartsOnNewLine := hasNextToken && itemStartLine(sourceFile, nextToken) > itemEndLine(sourceFile, lastItemToRemove)
	shouldRemoveOwnLine := !shouldRemoveLeadingComments &&
		startsOnOwnLine(sourceFile, memberItem, items) &&
		nextTokenStartsOnNewLine

	start := memberItem.start
	end := lastItemToRemove.end

	switch {
	case shouldRemoveLeadingComments:
		if nextTokenStartsOnNewLine {
			start = lineStartIndex(sourceFile.Text(), leadingComments[0].start)
			end = lineStartIndex(sourceFile.Text(), nextToken.start)
		} else {
			start = leadingComments[0].start
			end = nextToken.start
		}
	case shouldRemoveOwnLine:
		start = lineStartIndex(sourceFile.Text(), memberItem.start)
		end = lineStartIndex(sourceFile.Text(), nextToken.start)
	case hasPreviousToken && itemEndLine(sourceFile, previousToken) == itemStartLine(sourceFile, memberItem):
		start = previousToken.end
	case hasNextToken && itemStartLine(sourceFile, nextToken) == itemEndLine(sourceFile, lastItemToRemove):
		end = nextToken.start
	}

	return core.NewTextRange(start, end)
}

func leadingCommentsForMember(sourceFile *ast.SourceFile, member sourceItem, items []sourceItem) []sourceItem {
	previousToken, hasPreviousToken := itemBefore(items, member.start, false)
	startAfter := 0
	if hasPreviousToken {
		startAfter = previousToken.end
	}

	commentsBefore := make([]sourceItem, 0)
	for _, item := range items {
		if !item.isComment {
			continue
		}
		if item.start >= startAfter && item.end <= member.start {
			commentsBefore = append(commentsBefore, item)
		}
	}

	lastNonLeadingComment := -1
	for i, comment := range commentsBefore {
		next := member
		if i < len(commentsBefore)-1 {
			next = commentsBefore[i+1]
		}
		if !startsOnOwnLine(sourceFile, comment, items) || itemStartLine(sourceFile, next)-itemEndLine(sourceFile, comment) > 1 {
			lastNonLeadingComment = i
		}
	}

	return commentsBefore[lastNonLeadingComment+1:]
}

func trailingCommentsForMember(sourceFile *ast.SourceFile, member sourceItem, items []sourceItem) []sourceItem {
	if sharesLineWithAnotherToken(sourceFile, member, items) {
		return nil
	}

	nextToken, hasNextToken := itemAfter(items, member.end, false)
	endBefore := len(sourceFile.Text())
	if hasNextToken {
		endBefore = nextToken.start
	}

	trailingComments := make([]sourceItem, 0)
	for _, item := range items {
		if !item.isComment {
			continue
		}
		if item.start >= member.end && item.end <= endBefore && itemStartLine(sourceFile, item) == itemEndLine(sourceFile, member) {
			trailingComments = append(trailingComments, item)
		}
	}
	return trailingComments
}

func sharesLineWithAnotherToken(sourceFile *ast.SourceFile, member sourceItem, items []sourceItem) bool {
	previousToken, hasPreviousToken := itemBefore(items, member.start, false)
	nextToken, hasNextToken := itemAfter(items, member.end, false)
	return (hasPreviousToken && itemEndLine(sourceFile, previousToken) == itemStartLine(sourceFile, member)) ||
		(hasNextToken && itemStartLine(sourceFile, nextToken) == itemEndLine(sourceFile, member))
}

func startsOnOwnLine(sourceFile *ast.SourceFile, item sourceItem, items []sourceItem) bool {
	previous, ok := itemBefore(items, item.start, true)
	return !ok || itemEndLine(sourceFile, previous) != itemStartLine(sourceFile, item)
}

func semicolonInsertionToken(sourceFile *ast.SourceFile, classNode *ast.Node, memberNode *ast.Node, items []sourceItem) (sourceItem, bool) {
	// If removing this member lets the previous field initializer continue into
	// the next member (`[`, `*`, `in`, or `instanceof`), insert a semicolon.
	nextToken, hasNextToken := itemAfter(items, memberNode.End(), false)
	if !hasNextToken || !canContinueExpressionInClassBody(nextToken) {
		return sourceItem{}, false
	}

	memberStart := utils.TrimNodeTextRange(sourceFile, memberNode).Pos()
	previousToken, hasPreviousToken := itemBefore(items, memberStart, false)
	if !hasPreviousToken || isSemicolonSafePreviousToken(previousToken) {
		return sourceItem{}, false
	}

	members := classNode.Members()
	memberIndex := -1
	for i, member := range members {
		if member == memberNode {
			memberIndex = i
			break
		}
	}
	if memberIndex <= 0 {
		return sourceItem{}, false
	}

	previousMember := members[memberIndex-1]
	if !ast.IsPropertyDeclaration(previousMember) {
		return sourceItem{}, false
	}
	previousProperty := previousMember.AsPropertyDeclaration()
	if previousProperty == nil || previousProperty.Initializer == nil {
		return sourceItem{}, false
	}

	text := sourceFile.Text()
	i := previousMember.End() - 1
	for i > previousMember.Pos() && isASCIIWhitespace(text[i]) {
		i--
	}
	if i >= 0 && text[i] == ';' {
		return sourceItem{}, false
	}

	initializer := previousProperty.Initializer
	if initializer.Kind == ast.KindPostfixUnaryExpression {
		return sourceItem{}, false
	}
	if initializer.Kind == ast.KindArrowFunction {
		if arrow := initializer.AsArrowFunction(); arrow != nil && arrow.Body != nil && arrow.Body.Kind == ast.KindBlock {
			return sourceItem{}, false
		}
	}

	return previousToken, true
}

func canContinueExpressionInClassBody(token sourceItem) bool {
	return token.text == "[" || token.text == "*" || token.text == "in" || token.text == "instanceof"
}

func isSemicolonSafePreviousToken(token sourceItem) bool {
	switch token.text {
	case ":", ";", "{", "=>", "++", "--":
		return true
	default:
		return false
	}
}

func collectSourceItems(sourceFile *ast.SourceFile, sourceComments []*ast.CommentRange) []sourceItem {
	text := sourceFile.Text()
	items := make([]sourceItem, 0)

	scan := scanner.GetScannerForSourceFile(sourceFile, 0)
	for scan.Token() != ast.KindEndOfFile {
		start := scan.TokenStart()
		end := scan.TokenEnd()
		if start < end {
			items = append(items, sourceItem{
				start: start,
				end:   end,
				text:  text[start:end],
			})
		}
		scan.Scan()
	}

	seenComments := map[[2]int]bool{}
	for _, comment := range sourceComments {
		key := [2]int{comment.Pos(), comment.End()}
		if seenComments[key] {
			continue
		}
		seenComments[key] = true
		items = append(items, sourceItem{
			start:     comment.Pos(),
			end:       comment.End(),
			text:      text[comment.Pos():comment.End()],
			isComment: true,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].start != items[j].start {
			return items[i].start < items[j].start
		}
		if items[i].end != items[j].end {
			return items[i].end < items[j].end
		}
		return !items[i].isComment && items[j].isComment
	})

	return items
}

func itemBefore(items []sourceItem, pos int, includeComments bool) (sourceItem, bool) {
	var previous sourceItem
	found := false
	for _, item := range items {
		if item.start >= pos {
			break
		}
		if item.end <= pos && (includeComments || !item.isComment) {
			previous = item
			found = true
		}
	}
	return previous, found
}

func itemAfter(items []sourceItem, pos int, includeComments bool) (sourceItem, bool) {
	for _, item := range items {
		if item.start >= pos && (includeComments || !item.isComment) {
			return item, true
		}
	}
	return sourceItem{}, false
}

func itemStartLine(sourceFile *ast.SourceFile, item sourceItem) int {
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, item.start)
	return line
}

func itemEndLine(sourceFile *ast.SourceFile, item sourceItem) int {
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sourceFile, item.end)
	return line
}

func lineStartIndex(text string, pos int) int {
	if pos > len(text) {
		pos = len(text)
	}
	for pos > 0 && text[pos-1] != '\n' && text[pos-1] != '\r' {
		pos--
	}
	return pos
}

func isASCIIWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
