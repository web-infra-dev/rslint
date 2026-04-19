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

var NoUselessRenameRule = rule.Rule{
	Name: "no-useless-rename",
	Run: func(ctx rule.RuleContext, optionsAny any) rule.RuleListeners {
		opts := parseOptions(optionsAny)
		sf := ctx.SourceFile

		// report emits the diagnostic, attaching an autofix when possible.
		// - `outerNode`: the node whose range is replaced by the fix (the
		//   BindingElement / PropertyAssignment / ImportSpecifier / ExportSpecifier).
		// - `displayName`: the name shown in the message (corresponds to the
		//   ESLint rule's `initial.name || initial.value`).
		// - `reportType`: one of "Destructuring assignment" / "Import" / "Export".
		// - `fix`: precomputed fix or nil.
		report := func(outerNode *ast.Node, displayName, reportType string, fix *rule.RuleFix) {
			msg := rule.RuleMessage{
				Id:          "unnecessarilyRenamed",
				Description: reportType + " " + displayName + " unnecessarily renamed.",
			}
			if fix != nil {
				ctx.ReportNodeWithFixes(outerNode, msg, *fix)
				return
			}
			ctx.ReportNode(outerNode, msg)
		}

		// hasCommentBytes reports whether `//` or `/*` appears in the byte
		// range [start, end). Scanner-based lookups (`utils.HasCommentsInRange`)
		// are anchored at token boundaries and miss comments that sit between
		// specific child nodes — here we just need to know whether the
		// PropertyAssignment / BindingElement / specifier contains any
		// comment between the "removed" prefix and the "kept" tail. Since
		// these ranges only cover identifiers, colons, parentheses, and the
		// `as` keyword, a raw byte scan can't be confused by string or regex
		// literals the way it could in general code.
		sourceText := sf.Text()
		hasCommentBytes := func(start, end int) bool {
			if end > len(sourceText) {
				end = len(sourceText)
			}
			for i := start; i+1 < end; i++ {
				if sourceText[i] == '/' && (sourceText[i+1] == '/' || sourceText[i+1] == '*') {
					return true
				}
			}
			return false
		}

		// buildRangeFix replaces `outerRange` with the source text at
		// `keepRange`. Returns nil if any comment sits in the portion of
		// `outerRange` outside `keepRange` (the "removed" prefix
		// [outer.Pos, keep.Pos) or suffix [keep.End, outer.End)) — dropping
		// comments silently would lose information.
		buildRangeFix := func(outerRange, keepRange core.TextRange) *rule.RuleFix {
			if hasCommentBytes(outerRange.Pos(), keepRange.Pos()) {
				return nil
			}
			if hasCommentBytes(keepRange.End(), outerRange.End()) {
				return nil
			}
			text := sourceText[keepRange.Pos():keepRange.End()]
			fix := rule.RuleFixReplaceRange(outerRange, text)
			return &fix
		}

		// ---- ImportSpecifier: `import { foo as foo } from '...'` ----
		// In tsgo's AST, `propertyName` is the name before `as` (= ESTree's
		// `imported`), `name` is after `as` (= ESTree's `local`). ESLint only
		// fires when an explicit `as` is written (imported.range !== local.range);
		// tsgo encodes that as `propertyName != nil`.
		checkImport := func(node *ast.Node) {
			if opts.ignoreImport {
				return
			}
			spec := node.AsImportSpecifier()
			if spec == nil || spec.PropertyName == nil {
				return
			}
			// `spec.PropertyName` is a ModuleExportName (Identifier | StringLiteral);
			// `spec.Name()` is always an Identifier (the local binding).
			// `utils.GetStaticPropertyName` resolves both to their canonical text.
			importedName, ok := utils.GetStaticPropertyName(spec.PropertyName)
			if !ok {
				return
			}
			localName := spec.Name().AsIdentifier().Text
			if importedName != localName {
				return
			}
			outerRange := utils.TrimNodeTextRange(sf, node)
			replRange := utils.TrimNodeTextRange(sf, spec.Name())
			fix := buildRangeFix(outerRange, replRange)
			report(node, importedName, "Import", fix)
		}

		// ---- ExportSpecifier: `export { foo as foo }` ----
		// Both `propertyName` (before `as`) and `name` (after `as`) are
		// ModuleExportName (Identifier | StringLiteral). ESLint only fires when
		// `as` is explicit (local.range !== exported.range); in tsgo that's
		// `propertyName != nil`.
		checkExport := func(node *ast.Node) {
			if opts.ignoreExport {
				return
			}
			spec := node.AsExportSpecifier()
			if spec == nil || spec.PropertyName == nil {
				return
			}
			// Both ends are ModuleExportName — accept Identifier and StringLiteral.
			localName, ok := utils.GetStaticPropertyName(spec.PropertyName)
			if !ok {
				return
			}
			exportedName, ok := utils.GetStaticPropertyName(spec.Name())
			if !ok {
				return
			}
			if localName != exportedName {
				return
			}
			outerRange := utils.TrimNodeTextRange(sf, node)
			replRange := utils.TrimNodeTextRange(sf, spec.PropertyName)
			fix := buildRangeFix(outerRange, replRange)
			report(node, localName, "Export", fix)
		}

		// ---- Destructuring declaration: `let {foo: foo} = obj;` / `function f({foo: foo}) {}` ----
		// tsgo represents these as BindingElement children of ObjectBindingPattern.
		// - `PropertyName` is the source key (before `:`); `Name()` is the binding
		//   target (after `:`, an Identifier, another pattern, or a rest binding).
		// - Rest elements and shorthand bindings have `PropertyName == nil`.
		// - Computed keys can't be resolved statically — the rule skips them
		//   (matching ESLint's `property.computed` check).
		checkBindingElement := func(node *ast.Node) {
			if opts.ignoreDestructuring {
				return
			}
			// Parent guard: only object destructuring; array-binding elements
			// don't carry a `propertyName` so this is belt-and-braces.
			if node.Parent == nil || node.Parent.Kind != ast.KindObjectBindingPattern {
				return
			}
			be := node.AsBindingElement()
			if be == nil || be.PropertyName == nil {
				return
			}
			// Skip computed property names — ESLint: "we have no idea if a
			// rename is useless or not". `GetStaticPropertyName` can resolve
			// `['foo']` statically, but the rule intentionally treats computed
			// keys as opaque regardless of whether they're constant-foldable.
			if be.PropertyName.Kind == ast.KindComputedPropertyName {
				return
			}
			keyName, ok := utils.GetStaticPropertyName(be.PropertyName)
			if !ok {
				return
			}
			bindingName := be.Name()
			if bindingName == nil || bindingName.Kind != ast.KindIdentifier {
				return
			}
			if keyName != bindingName.AsIdentifier().Text {
				return
			}
			// Replacement text runs from the binding name through the end of
			// the element, so it includes any default (`= init`). For
			// `{foo: foo = 1}` the output becomes `{foo = 1}`.
			outerRange := utils.TrimNodeTextRange(sf, node)
			nameRange := utils.TrimNodeTextRange(sf, bindingName)
			keepRange := core.NewTextRange(nameRange.Pos(), outerRange.End())
			fix := buildRangeFix(outerRange, keepRange)
			report(node, keyName, "Destructuring assignment", fix)
		}

		// ---- Destructuring assignment pattern: `({foo: foo} = obj);` ----
		// tsgo keeps the outer node as `ObjectLiteralExpression`; only the
		// assignment context reclassifies it semantically. `ast.IsAssignmentTarget`
		// walks up through parentheses / arrays / nested object literals to
		// confirm we're in an assignment target position.
		checkAssignmentProperty := func(node *ast.Node) {
			if opts.ignoreDestructuring {
				return
			}
			parent := node.Parent
			if parent == nil || parent.Kind != ast.KindObjectLiteralExpression {
				return
			}
			if !ast.IsAssignmentTarget(parent) {
				return
			}
			pa := node.AsPropertyAssignment()
			if pa == nil {
				return
			}
			nameNode := pa.Name()
			if nameNode == nil {
				return
			}
			// Skip computed keys (see BindingElement branch for the rationale).
			if nameNode.Kind == ast.KindComputedPropertyName {
				return
			}
			keyName, ok := utils.GetStaticPropertyName(nameNode)
			if !ok {
				return
			}
			init := pa.Initializer
			if init == nil {
				return
			}

			// Classify the initializer:
			//  - ParenthesizedExpression wrapping Identifier → useless rename
			//    with no default; unwrap to reach the identifier.
			//  - BinaryExpression(=) → destructuring default; treat like
			//    ESLint's AssignmentPattern. The identifier being renamed
			//    is the (unwrapped) left side.
			//  - Anything else → not a rename.
			targetIdent := init
			hasDefault := false
			if targetIdent.Kind == ast.KindBinaryExpression {
				bin := targetIdent.AsBinaryExpression()
				if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
					hasDefault = true
					targetIdent = bin.Left
				}
			}
			// Autofix can't handle `({foo: (foo) = a} = obj)` — shorthand
			// properties don't accept parenthesised identifiers. ESLint emits
			// the diagnostic with a null fix in this case.
			leftParenthesized := hasDefault && targetIdent.Kind == ast.KindParenthesizedExpression
			targetIdent = ast.SkipParentheses(targetIdent)
			if targetIdent == nil || targetIdent.Kind != ast.KindIdentifier {
				return
			}
			if keyName != targetIdent.AsIdentifier().Text {
				return
			}

			var fix *rule.RuleFix
			if !leftParenthesized {
				outerRange := utils.TrimNodeTextRange(sf, node)
				// Replacement spans the logical "value" — either the bare
				// identifier (no default) or the full `ident = init`
				// expression. When the identifier sits inside explicit parens
				// like `(foo)`, we deliberately take only the identifier's
				// range so the emitted shorthand is `{foo}`, not `{(foo)}`.
				var replRange core.TextRange
				if hasDefault {
					// Keep the full BinaryExpression range — this includes any
					// default value. The left was not parenthesized (checked
					// above), so emitting the raw text is safe.
					replRange = utils.TrimNodeTextRange(sf, init)
				} else {
					replRange = utils.TrimNodeTextRange(sf, targetIdent)
				}
				fix = buildRangeFix(outerRange, replRange)
			}
			report(node, keyName, "Destructuring assignment", fix)
		}

		return rule.RuleListeners{
			ast.KindImportSpecifier:    checkImport,
			ast.KindExportSpecifier:    checkExport,
			ast.KindBindingElement:     checkBindingElement,
			ast.KindPropertyAssignment: checkAssignmentProperty,
		}
	},
}
