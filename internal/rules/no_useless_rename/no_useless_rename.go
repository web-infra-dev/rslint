package no_useless_rename

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-useless-rename

type options struct {
	ignoreDestructuring bool
	ignoreImport        bool
	ignoreExport        bool
}

func parseOptions(opts any) options {
	out := options{}
	m := utils.GetOptionsMap(opts)
	if m == nil {
		return out
	}
	if v, ok := m["ignoreDestructuring"].(bool); ok {
		out.ignoreDestructuring = v
	}
	if v, ok := m["ignoreImport"].(bool); ok {
		out.ignoreImport = v
	}
	if v, ok := m["ignoreExport"].(bool); ok {
		out.ignoreExport = v
	}
	return out
}

type noUselessRenameState struct {
	ctx        rule.RuleContext
	sourceText string
}

// hasCommentBytes reports whether `//` or `/*` appears in the byte range
// [start, end). Scanner-based lookups are anchored at token boundaries and can
// miss comments between these specific child nodes. The ranges here contain
// only identifiers, punctuation, parentheses, and the `as` keyword, so a raw
// scan cannot mistake string or regexp contents for comments.
func (s *noUselessRenameState) hasCommentBytes(start, end int) bool {
	if end > len(s.sourceText) {
		end = len(s.sourceText)
	}
	for i := start; i+1 < end; i++ {
		if s.sourceText[i] == '/' && (s.sourceText[i+1] == '/' || s.sourceText[i+1] == '*') {
			return true
		}
	}
	return false
}

// report emits a diagnostic and, when requested, replaces outerNode with the
// source text at keepNode. keepThroughEnd retains the suffix from keepNode to
// outerNode's end (used for a binding default such as `{foo: foo = 1}`).
func (s *noUselessRenameState) report(
	outerNode *ast.Node,
	displayName string,
	reportType string,
	keepNode *ast.Node,
	keepThroughEnd bool,
) {
	msg := rule.RuleMessage{
		Id:          "unnecessarilyRenamed",
		Description: reportType + " " + displayName + " unnecessarily renamed.",
	}
	outerRange := utils.TrimNodeTextRange(s.ctx.SourceFile, outerNode)
	if keepNode == nil {
		s.ctx.ReportRange(outerRange, msg)
		return
	}

	s.ctx.ReportRangeWithDeferredFixes(outerRange, msg, func() []rule.RuleFix {
		keepRange := utils.TrimNodeTextRange(s.ctx.SourceFile, keepNode)
		if keepThroughEnd {
			keepRange = core.NewTextRange(keepRange.Pos(), outerRange.End())
		}
		// Comments outside keepRange would be lost by the replacement, so the
		// diagnostic remains intentionally unfixable in that case.
		if s.hasCommentBytes(outerRange.Pos(), keepRange.Pos()) ||
			s.hasCommentBytes(keepRange.End(), outerRange.End()) {
			return nil
		}
		return []rule.RuleFix{
			rule.RuleFixReplaceRange(outerRange, s.sourceText[keepRange.Pos():keepRange.End()]),
		}
	})
}

// checkImport handles `import { foo as foo } from '...'`.
func (s *noUselessRenameState) checkImport(node *ast.Node) {
	spec := node.AsImportSpecifier()
	if spec.PropertyName == nil {
		return
	}
	importedName, ok := utils.GetStaticPropertyName(spec.PropertyName)
	if !ok {
		return
	}
	local := spec.Name()
	if importedName != local.AsIdentifier().Text {
		return
	}
	s.report(node, importedName, "Import", local, false)
}

// checkExport handles `export { foo as foo }`.
func (s *noUselessRenameState) checkExport(node *ast.Node) {
	spec := node.AsExportSpecifier()
	if spec.PropertyName == nil {
		return
	}
	localName, ok := utils.GetStaticPropertyName(spec.PropertyName)
	if !ok {
		return
	}
	exportedName, ok := utils.GetStaticPropertyName(spec.Name())
	if !ok || localName != exportedName {
		return
	}
	s.report(node, localName, "Export", spec.PropertyName, false)
}

// checkBindingElement handles declaration patterns such as
// `let {foo: foo} = value` and function parameters.
func (s *noUselessRenameState) checkBindingElement(node *ast.Node) {
	if node.Parent == nil || node.Parent.Kind != ast.KindObjectBindingPattern {
		return
	}
	element := node.AsBindingElement()
	if element.PropertyName == nil || element.PropertyName.Kind == ast.KindComputedPropertyName {
		return
	}
	keyName, ok := utils.GetStaticPropertyName(element.PropertyName)
	if !ok {
		return
	}
	bindingName := element.Name()
	if bindingName == nil || bindingName.Kind != ast.KindIdentifier ||
		keyName != bindingName.AsIdentifier().Text {
		return
	}
	// Retain any default initializer: `{foo: foo = 1}` becomes `{foo = 1}`.
	s.report(node, keyName, "Destructuring assignment", bindingName, true)
}

// checkAssignmentProperty is registered only for object properties traversed
// as assignment patterns, so ordinary object literals never enter this path.
func (s *noUselessRenameState) checkAssignmentProperty(node *ast.Node) {
	property := node.AsPropertyAssignment()
	nameNode := property.Name()
	if nameNode == nil || nameNode.Kind == ast.KindComputedPropertyName {
		return
	}
	keyName, ok := utils.GetStaticPropertyName(nameNode)
	if !ok || property.Initializer == nil {
		return
	}

	// A default initializer is represented as BinaryExpression(=). Parentheses
	// around the renamed identifier are otherwise transparent for comparison.
	initializer := property.Initializer
	targetIdentifier := initializer
	hasDefault := false
	if targetIdentifier.Kind == ast.KindBinaryExpression {
		binary := targetIdentifier.AsBinaryExpression()
		if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			hasDefault = true
			targetIdentifier = binary.Left
		}
	}
	leftParenthesized := hasDefault && targetIdentifier.Kind == ast.KindParenthesizedExpression
	targetIdentifier = ast.SkipParentheses(targetIdentifier)
	if targetIdentifier == nil || targetIdentifier.Kind != ast.KindIdentifier ||
		keyName != targetIdentifier.AsIdentifier().Text {
		return
	}

	// `{foo: (foo) = value}` cannot become a shorthand property while keeping
	// the parentheses, matching ESLint's diagnostic-without-fix behavior.
	if leftParenthesized {
		s.report(node, keyName, "Destructuring assignment", nil, false)
		return
	}
	keepNode := targetIdentifier
	if hasDefault {
		keepNode = initializer
	}
	s.report(node, keyName, "Destructuring assignment", keepNode, false)
}

// checkWrappedAssignmentPattern continues the linter's allow-pattern walk
// through syntax wrappers whose children are otherwise visited as ordinary
// expressions. It is registered on the allow-pattern listener keys, so normal
// parenthesized and non-null expressions do not enter this path.
func (s *noUselessRenameState) checkWrappedAssignmentPattern(node *ast.Node) {
	for node != nil && (node.Kind == ast.KindParenthesizedExpression || node.Kind == ast.KindNonNullExpression) {
		node = node.Expression()
	}
	s.walkAssignmentPattern(node)
}

// checkIterationStatement covers assignment patterns used as for-in/of
// initializers. The linter's dedicated allow-pattern traversal is initiated by
// `=` assignments; iteration targets need this narrow rule-level entry point.
func (s *noUselessRenameState) checkIterationStatement(node *ast.Node) {
	s.walkAssignmentPattern(node.Initializer())
}

func (s *noUselessRenameState) walkAssignmentPattern(node *ast.Node) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		s.walkAssignmentPattern(node.Expression())
	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			s.walkAssignmentPattern(element)
		}
	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			s.walkAssignmentPattern(property)
		}
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		s.walkAssignmentPattern(node.Expression())
	case ast.KindPropertyAssignment:
		s.checkAssignmentProperty(node)
		initializer := node.AsPropertyAssignment().Initializer
		if initializer != nil && initializer.Kind == ast.KindBinaryExpression {
			binary := initializer.AsBinaryExpression()
			if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
				return
			}
		}
		s.walkAssignmentPattern(initializer)
	}
}

var NoUselessRenameRule = rule.Rule{
	Name: "no-useless-rename",
	Run: func(ctx rule.RuleContext, optionsAny []any) rule.RuleListeners {
		opts := parseOptions(rule.LegacyUnwrapOptions(optionsAny))
		if opts.ignoreDestructuring && opts.ignoreImport && opts.ignoreExport {
			return nil
		}

		state := &noUselessRenameState{
			ctx:        ctx,
			sourceText: ctx.SourceFile.Text(),
		}
		listeners := make(rule.RuleListeners, 8)
		if !opts.ignoreImport {
			listeners[ast.KindImportSpecifier] = state.checkImport
		}
		if !opts.ignoreExport {
			listeners[ast.KindExportSpecifier] = state.checkExport
		}
		if !opts.ignoreDestructuring {
			listeners[ast.KindBindingElement] = state.checkBindingElement
			listeners[rule.ListenerOnAllowPattern(ast.KindPropertyAssignment)] = state.checkAssignmentProperty
			checkWrappedPattern := state.checkWrappedAssignmentPattern
			listeners[rule.ListenerOnAllowPattern(ast.KindParenthesizedExpression)] = checkWrappedPattern
			listeners[rule.ListenerOnAllowPattern(ast.KindNonNullExpression)] = checkWrappedPattern
			checkIteration := state.checkIterationStatement
			listeners[ast.KindForInStatement] = checkIteration
			listeners[ast.KindForOfStatement] = checkIteration
		}
		return listeners
	},
}
