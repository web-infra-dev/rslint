package prefer_const

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed prefer_const.schema.json
var schemaJSON []byte

type preferConstOptions struct {
	destructuring          string // "any" or "all", default "any"
	ignoreReadBeforeAssign bool   // default false
}

func parseOptions(options []any) preferConstOptions {
	result := preferConstOptions{
		destructuring:          "any",
		ignoreReadBeforeAssign: false,
	}
	if len(options) == 0 {
		return result
	}

	m, _ := options[0].(map[string]any)
	if d, ok := m["destructuring"].(string); ok {
		result.destructuring = d
	}
	if v, ok := m["ignoreReadBeforeAssign"].(bool); ok {
		result.ignoreReadBeforeAssign = v
	}

	return result
}

// candidateInfo holds information about a single binding name candidate.
type candidateRef uint32

type candidateInfo struct {
	nameNode                    *ast.Node
	declNode                    *ast.Node
	symbol                      *ast.Symbol
	reportNode                  *ast.Node // node to report on (may differ from nameNode for uninitialized vars)
	nextCandidateWithSameSymbol candidateRef
	hasInitializer              bool
	seenPostDeclarationWrite    bool
	readBeforeFirstAssign       bool
}

type destructuringGroup struct {
	candidateStart int
	candidateEnd   int
}

type declarationListInfo struct {
	node           *ast.Node
	candidateStart int
	candidateEnd   int
	allHaveInit    bool
}

type symbolInfo struct {
	firstWrite    *ast.Node
	candidateHead candidateRef // one-based candidate index; zero means no candidate
	writeCount    uint8        // saturated at 2; the rule only distinguishes 0, 1, and multiple writes
}

// preferConstState derives the reference facts needed by this rule while the
// linter performs its normal file traversal. Calling RefStore.References for
// every binding would trigger a second whole-file walk that buckets all
// reference-position identifiers, even though prefer-const only needs writes
// and, for uninitialized bindings, reads before the first assignment.
type preferConstState struct {
	ctx              *rule.RuleContext
	opts             preferConstOptions
	destructuringAll bool
	sourceText       string

	declarationLists    []declarationListInfo
	destructuringGroups []destructuringGroup
	candidates          []candidateInfo
	symbols             map[*ast.Symbol]symbolInfo
	uninitializedNames  map[string]struct{}
}

// https://eslint.org/docs/latest/rules/prefer-const
var PreferConstRule = rule.Rule{
	Name:   "prefer-const",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sourceText := ctx.SourceFile.Text()
		// A real let declaration necessarily contains this exact byte sequence.
		// False positives (for example in comments) only take the normal path;
		// false negatives are impossible, so this is safe as an early exit.
		if !strings.Contains(sourceText, "let") {
			return nil
		}

		opts := parseOptions(options)
		state := preferConstState{
			ctx:              &ctx,
			opts:             opts,
			destructuringAll: opts.destructuring == "all",
			sourceText:       sourceText,
		}

		return rule.RuleListeners{
			ast.KindVariableDeclarationList:        state.visitDeclarationList,
			ast.KindIdentifier:                     state.visitIdentifier,
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) { state.finish() },
		}
	},
}

func (s *preferConstState) visitDeclarationList(node *ast.Node) {
	declList := node.AsVariableDeclarationList()
	if declList == nil || node.Flags&ast.NodeFlagsLet == 0 || declList.Declarations == nil || isInForStatement(node) {
		return
	}

	list := declarationListInfo{
		node:           node,
		candidateStart: len(s.candidates),
		allHaveInit:    true,
	}
	isForInOrOf := isInForInOrOf(node)
	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Name() == nil {
			continue
		}

		hasInitializer := varDecl.Initializer != nil || isForInOrOf
		if !hasInitializer {
			list.allHaveInit = false
		}
		nameNode := varDecl.Name()
		candidateStart := len(s.candidates)
		if nameNode.Kind == ast.KindIdentifier {
			if nameNode.Text() != "" {
				s.addCandidate(nameNode, decl, hasInitializer)
			}
		} else {
			utils.CollectBindingNames(nameNode, func(ident *ast.Node, _ string) {
				s.addCandidate(ident, decl, hasInitializer)
			})
		}
		candidateEnd := len(s.candidates)
		if candidateEnd == candidateStart {
			continue
		}

		if s.destructuringAll && (nameNode.Kind == ast.KindObjectBindingPattern || nameNode.Kind == ast.KindArrayBindingPattern) {
			s.destructuringGroups = append(s.destructuringGroups, destructuringGroup{
				candidateStart: candidateStart,
				candidateEnd:   candidateEnd,
			})
		}
	}
	list.candidateEnd = len(s.candidates)
	if list.candidateEnd != list.candidateStart {
		s.declarationLists = append(s.declarationLists, list)
	}
}

func (s *preferConstState) addCandidate(nameNode, declNode *ast.Node, hasInitializer bool) {
	symbol := utils.BindingNameSymbol(nameNode)
	candidate := candidateInfo{
		nameNode:       nameNode,
		declNode:       declNode,
		symbol:         symbol,
		hasInitializer: hasInitializer,
	}
	if symbol != nil {
		if s.symbols == nil {
			s.symbols = make(map[*ast.Symbol]symbolInfo)
		}
		info := s.symbols[symbol]
		candidate.nextCandidateWithSameSymbol = info.candidateHead
		info.candidateHead = candidateRef(len(s.candidates) + 1)
		s.symbols[symbol] = info
		if !hasInitializer {
			if s.uninitializedNames == nil {
				s.uninitializedNames = make(map[string]struct{})
			}
			s.uninitializedNames[nameNode.Text()] = struct{}{}
		}
	}
	s.candidates = append(s.candidates, candidate)
}

func (s *preferConstState) visitIdentifier(node *ast.Node) {
	isWrite := utils.IsWriteReference(node)
	if !isWrite {
		if _, ok := s.uninitializedNames[node.Text()]; !ok {
			return
		}
	}

	symbol := s.ctx.Refs.Resolve(node)
	if symbol == nil {
		return
	}
	info := s.symbols[symbol]
	if isWrite {
		// Count writes for every symbol, including writes that precede a let
		// declaration and non-let targets that may share a destructuring write
		// group with a candidate.
		if s.symbols == nil {
			s.symbols = make(map[*ast.Symbol]symbolInfo)
		}
		if info.firstWrite == nil {
			info.firstWrite = node
		}
		if info.writeCount < 2 {
			info.writeCount++
		}
		s.symbols[symbol] = info
	}

	for head := info.candidateHead; head != 0; {
		candidate := &s.candidates[int(head)-1]
		head = candidate.nextCandidateWithSameSymbol
		if candidate.hasInitializer || node.Pos() < candidate.declNode.Pos() {
			continue
		}
		if isWrite {
			candidate.seenPostDeclarationWrite = true
		} else if !candidate.seenPostDeclarationWrite {
			candidate.readBeforeFirstAssign = true
		}
	}
}

func (s *preferConstState) finish() {
	// Declaration lists and destructuring groups are appended in the same AST
	// traversal order, so a single cursor can consume each sparse group once.
	destructuringGroupIndex := 0
	for i := range s.declarationLists {
		s.finishDeclarationList(&s.declarationLists[i], &destructuringGroupIndex)
	}
}

func (s *preferConstState) finishDeclarationList(list *declarationListInfo, destructuringGroupIndex *int) {
	var candidateBuffer [4]int
	constCandidates := candidateBuffer[:0]

	for candidateIndex := list.candidateStart; candidateIndex < list.candidateEnd; {
		groupStart := candidateIndex
		groupEnd := candidateIndex + 1
		requireAll := false
		if *destructuringGroupIndex < len(s.destructuringGroups) {
			group := s.destructuringGroups[*destructuringGroupIndex]
			if group.candidateStart == candidateIndex {
				groupEnd = group.candidateEnd
				requireAll = true
				*destructuringGroupIndex = *destructuringGroupIndex + 1
			}
		}

		reportableStart := len(constCandidates)
		for ; candidateIndex < groupEnd; candidateIndex++ {
			if s.shouldReport(&s.candidates[candidateIndex]) {
				constCandidates = append(constCandidates, candidateIndex)
			}
		}
		if requireAll && len(constCandidates)-reportableStart != groupEnd-groupStart {
			constCandidates = constCandidates[:reportableStart]
		}
	}

	if s.destructuringAll {
		writeIndex := 0
		for _, index := range constCandidates {
			candidate := &s.candidates[index]
			if candidate.reportNode != nil && !s.allDestructuringWriteTargetsConst(candidate.reportNode) {
				continue
			}
			constCandidates[writeIndex] = index
			writeIndex++
		}
		constCandidates = constCandidates[:writeIndex]
	}

	totalBindings := list.candidateEnd - list.candidateStart
	canFix := list.allHaveInit && totalBindings > 0 && len(constCandidates) == totalBindings
	var buildFix func() []rule.RuleFix
	if canFix {
		letStart := -1
		buildFix = func() []rule.RuleFix {
			if letStart < 0 {
				letStart = scanner.SkipTrivia(s.sourceText, list.node.Pos())
			}
			letRange := list.node.Loc.WithPos(letStart).WithEnd(letStart + len("let"))
			return []rule.RuleFix{rule.RuleFixReplaceRange(letRange, "const")}
		}
	}
	for _, index := range constCandidates {
		candidate := &s.candidates[index]
		reportOn := candidate.nameNode
		if candidate.reportNode != nil {
			reportOn = candidate.reportNode
		}
		msg := rule.RuleMessage{
			Id:          "useConst",
			Description: "'" + candidate.nameNode.Text() + "' is never reassigned. Use 'const' instead.",
		}
		reportRange := reportOn.Loc.WithPos(scanner.SkipTrivia(s.sourceText, reportOn.Pos()))
		if canFix {
			s.ctx.ReportRangeWithDeferredFixes(reportRange, msg, buildFix)
		} else {
			s.ctx.ReportRange(reportRange, msg)
		}
	}
}

func (s *preferConstState) shouldReport(candidate *candidateInfo) bool {
	if candidate.symbol == nil {
		return false
	}
	// A global listed by `/* exported */` is shared with separately loaded
	// files, any of which may reassign it, so upstream declines to judge
	// whether it could be `const`.
	if s.ctx.IsExportedGlobalBinding(candidate.declNode, candidate.nameNode.Text()) {
		return false
	}
	info := s.symbols[candidate.symbol]
	if candidate.hasInitializer {
		return info.writeCount == 0
	}
	if info.writeCount != 1 {
		return false
	}

	writeNode := findWriteInSameBlock(info.firstWrite, candidate.declNode, s.ctx)
	if writeNode == nil {
		return false
	}
	if !candidate.readBeforeFirstAssign {
		candidate.reportNode = writeNode
	}
	return !s.opts.ignoreReadBeforeAssign || !candidate.readBeforeFirstAssign
}

// isInForStatement checks if a VariableDeclarationList is the initializer of a regular for statement.
func isInForStatement(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	return node.Parent.Kind == ast.KindForStatement
}

// isInForInOrOf checks if a VariableDeclarationList is the initializer of a for-in or for-of statement.
func isInForInOrOf(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	return node.Parent.Kind == ast.KindForInStatement || node.Parent.Kind == ast.KindForOfStatement
}

// findContainingBlock finds the nearest Block, SourceFile, ModuleBlock, CaseClause, or DefaultClause ancestor.
func findContainingBlock(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock,
			ast.KindCaseClause, ast.KindDefaultClause:
			return true
		}
		return false
	})
}

// isDirectChildOfBlock checks if a write node is a direct statement within the
// given block. Walks from the write node to the block, ensuring there are no
// intervening control flow statements (if, while, for, etc.) or nested blocks.
// This handles both braced (`if (c) { x=1; }`) and brace-less (`if (c) x=1;`) forms.
func isDirectChildOfBlock(writeNode *ast.Node, declBlock *ast.Node) bool {
	current := writeNode.Parent
	for current != nil {
		if current == declBlock {
			return true
		}
		switch current.Kind {
		// Any of these between the write and the block means the write is nested
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock,
			ast.KindCaseClause, ast.KindDefaultClause,
			ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement,
			ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWithStatement, ast.KindLabeledStatement, ast.KindSwitchStatement:
			return false
		}
		current = current.Parent
	}
	return false
}

// findWriteInSameBlock finds the single write reference to an uninitialized variable
// if it is at the same block nesting level as its declaration. Returns the write node,
// or nil if the write is in a nested block or not found.
// ESLint only suggests const for uninitialized variables when the write can be merged
// into the declaration (i.e., same block level). Writes inside nested blocks (if, for,
// try, function bodies, etc.) cannot be safely merged.
func findWriteInSameBlock(writeNode *ast.Node, declNode *ast.Node, ctx *rule.RuleContext) *ast.Node {
	declBlock := findContainingBlock(declNode)
	if declBlock == nil || writeNode == nil {
		return nil
	}

	// The write must be a direct statement in the declaration's containing block.
	// Writes inside nested blocks, control flow bodies (if/while/for without braces),
	// labeled statements, etc. cannot be safely merged into the declaration.
	if !isDirectChildOfBlock(writeNode, declBlock) {
		return nil
	}

	// ESLint only flags uninitialized variables when the write is a standalone
	// assignment statement (ExpressionStatement > AssignmentExpression),
	// not when embedded in conditions, chained assignments, or other expressions.
	if !isStandaloneAssignment(writeNode) {
		return nil
	}

	// ESLint doesn't report variables in destructuring assignments that contain
	// member expressions (obj.prop) or identifiers from a different block scope.
	// Same-scope var/const targets don't suppress reporting (only suppress fix).
	if hasNonReportableDestructuringTarget(writeNode, declNode, ctx) {
		return nil
	}

	return writeNode
}

// allDestructuringWriteTargetsConst checks whether all identifier targets in the
// destructuring assignment containing writeNode have at most 1 write reference.
// Used with destructuring: "all" to suppress reporting when not all variables in
// the destructuring write group can be const.
func (s *preferConstState) allDestructuringWriteTargetsConst(writeNode *ast.Node) bool {
	assignExpr := ast.FindAncestor(writeNode, func(n *ast.Node) bool {
		return ast.IsDestructuringAssignment(n)
	})
	if assignExpr == nil {
		return true // not in a destructuring, no group constraint
	}

	left := assignExpr.AsBinaryExpression().Left
	allConst := true
	utils.VisitDestructuringIdentifiers(left, func(ident *ast.Node) {
		if !allConst {
			return
		}
		// RefStore.Resolve handles shorthand property names the same as
		// plain identifiers (unlike checker.GetSymbolAtLocation, which
		// resolves a shorthand name to the property symbol rather than the
		// variable it reads/writes).
		sym := s.ctx.Refs.Resolve(ident)
		if sym == nil {
			return
		}
		if s.symbols[sym].writeCount > 1 {
			allConst = false
		}
	})
	return allConst
}

// hasNonReportableDestructuringTarget checks if a write node is inside a destructuring
// assignment that should suppress REPORTING. This is limited to:
// 1. Member expressions (obj.prop, arr[i]) — can never be const declarations
// 2. Identifiers whose declaration is in a different block scope — can't safely merge
// Same-scope identifiers (var, const, import, param, etc.) do NOT suppress reporting.
// Uses ctx.Refs symbol resolution instead of name-set collection to correctly
// handle all declaration types (imports, function/class declarations, parameters, etc.).
func hasNonReportableDestructuringTarget(writeNode *ast.Node, declNode *ast.Node, ctx *rule.RuleContext) bool {
	assignExpr := ast.FindAncestor(writeNode, func(n *ast.Node) bool {
		return ast.IsDestructuringAssignment(n)
	})
	if assignExpr == nil {
		return false
	}

	left := assignExpr.AsBinaryExpression().Left
	declBlock := findContainingBlock(declNode)
	if declBlock == nil {
		return true
	}
	return hasNonReportableTarget(left, declBlock, ctx)
}

// hasNonReportableTarget checks if a destructuring pattern contains targets that
// should suppress reporting: member expressions, or identifiers declared in a
// different block scope. Uses ctx.Refs to resolve each identifier's declaration
// rather than pre-collecting names, so it correctly handles imports, parameters,
// function declarations, class declarations, etc.
func hasNonReportableTarget(node *ast.Node, declBlock *ast.Node, ctx *rule.RuleContext) bool {
	if node.Kind == ast.KindIdentifier {
		sym := ctx.Refs.Resolve(node)
		if sym == nil || len(sym.Declarations) == 0 {
			// Can't resolve — treat as non-reportable (conservative)
			return true
		}
		targetDeclBlock := findContainingBlock(sym.Declarations[0])
		return targetDeclBlock != declBlock
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if found {
			return true
		}
		switch child.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			found = true
			return true
		case ast.KindIdentifier:
			sym := ctx.Refs.Resolve(child)
			if sym == nil || len(sym.Declarations) == 0 {
				found = true
				return true
			}
			if findContainingBlock(sym.Declarations[0]) != declBlock {
				found = true
				return true
			}
		case ast.KindShorthandPropertyAssignment:
			shorthand := child.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil {
				valSym := ctx.Refs.Resolve(shorthand.Name())
				if valSym == nil || len(valSym.Declarations) == 0 {
					found = true
					return true
				}
				if findContainingBlock(valSym.Declarations[0]) != declBlock {
					found = true
					return true
				}
			}
		case ast.KindPropertyAssignment:
			pa := child.AsPropertyAssignment()
			if pa != nil && pa.Initializer != nil {
				if hasNonReportableTarget(pa.Initializer, declBlock, ctx) {
					found = true
					return true
				}
			}
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindSpreadElement, ast.KindSpreadAssignment:
			if hasNonReportableTarget(child, declBlock, ctx) {
				found = true
				return true
			}
		case ast.KindBinaryExpression:
			be := child.AsBinaryExpression()
			if be != nil && be.Left != nil {
				if hasNonReportableTarget(be.Left, declBlock, ctx) {
					found = true
					return true
				}
			}
		}
		return false
	})
	return found
}

// isStandaloneAssignment checks if a write reference identifier is part of an
// assignment expression that is directly inside an ExpressionStatement.
// Uses ast.GetAssignmentTarget from the TypeScript shim for robust pattern walking
// through destructuring, shorthand properties, parentheses, etc.
// Returns false for ++/--, for-in/of targets, conditions, chained assignments,
// and other non-statement expressions.
func isStandaloneAssignment(identNode *ast.Node) bool {
	target := ast.GetAssignmentTarget(identNode)
	if target == nil {
		return false
	}

	// ++/-- (PrefixUnary/PostfixUnary) can't be converted to const for uninitialized variables.
	// for-in/of targets may execute multiple times.
	switch target.Kind {
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindForInStatement, ast.KindForOfStatement:
		return false
	}

	// GetAssignmentTarget may return a default value's BinaryExpression
	// (e.g. [x = 5] returns x=5, not [x=5]=[1]). If so, find the outer destructuring.
	if utils.IsDefaultValueInDestructuringAssignment(target) {
		target = ast.FindAncestor(target.Parent, func(n *ast.Node) bool {
			return ast.IsDestructuringAssignment(n)
		})
		if target == nil {
			return false
		}
	}

	// The assignment must be directly inside an ExpressionStatement (possibly through parens).
	parent := target.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindExpressionStatement
}
