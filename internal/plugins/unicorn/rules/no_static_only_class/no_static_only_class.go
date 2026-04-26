package no_static_only_class

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// modifierFlagsDisqualifying captures the TS-only modifiers that exclude a
// member from "valid static" status: the three accessibility keywords
// (`public` / `private` / `protected`), `readonly`, `declare` (ambient), and
// any class-element decorator. tsgo carries decorators on the same modifier
// list as keywords, so a single flag check covers both.
const modifierFlagsDisqualifying = ast.ModifierFlagsAccessibilityModifier |
	ast.ModifierFlagsReadonly |
	ast.ModifierFlagsAmbient |
	ast.ModifierFlagsDecorator

// isClassElementMethodLike reports whether a class element is one of the
// kinds the upstream rule treats as `MethodDefinition`: regular method,
// getter, setter, or constructor.
//
// KindConstructor is included because tsgo parses `static constructor()` as
// a real ConstructorDeclaration even though, with `static`, it is
// semantically just a method named "constructor". Upstream's ESTree treats
// `static constructor()` as a `MethodDefinition` and the `IsStatic` check
// below filters out actual (non-static) constructors.
func isClassElementMethodLike(node *ast.Node) bool {
	return ast.IsMethodDeclaration(node) ||
		ast.IsGetAccessorDeclaration(node) ||
		ast.IsSetAccessorDeclaration(node) ||
		ast.IsConstructorDeclaration(node)
}

// isStaticMember mirrors upstream's `isStaticMember` predicate: a class
// element counts as a "valid" static member only when it is a public,
// non-decorated, non-readonly, non-declare static field, method, accessor,
// or static-named-constructor with a non-private key.
func isStaticMember(member *ast.Node) bool {
	if member == nil {
		return false
	}
	if !ast.IsPropertyDeclaration(member) && !isClassElementMethodLike(member) {
		return false
	}
	if !ast.HasStaticModifier(member) {
		return false
	}
	if ast.IsPrivateIdentifierClassElementDeclaration(member) {
		return false
	}
	if ast.HasSyntacticModifier(member, modifierFlagsDisqualifying) {
		return false
	}
	return true
}

// classKeywordPos returns the start position of the `class` keyword,
// skipping over modifiers (`export`, `default`, `abstract`, decorators, …)
// and any leading trivia. tsgo carries those modifiers on the class node,
// while ESTree puts them on a wrapping `Export*Declaration`.
func classKeywordPos(node *ast.Node, sourceText string) int {
	pos := node.Pos()
	if mods := node.Modifiers(); mods != nil && len(mods.Nodes) > 0 {
		pos = mods.Nodes[len(mods.Nodes)-1].End()
	}
	return scanner.SkipTrivia(sourceText, pos)
}

// classHeadRange returns the range from the `class` keyword to the end of
// the last token before the body's `{`, matching ESLint's
// `getClassHeadLocation`. We scan tokens forward and update `end` to each
// token's end until we encounter `{`. Comments and whitespace between
// tokens are handled transparently by the scanner; this avoids the
// off-by-one problems of using `typeParameters.End()` (which excludes the
// closing `>`).
func classHeadRange(node *ast.Node, sf *ast.SourceFile) core.TextRange {
	start := classKeywordPos(node, sf.Text())
	end := start + len("class")

	s := scanner.GetScannerForSourceFile(sf, start)
	for {
		kind := s.Token()
		if kind == ast.KindOpenBraceToken || kind == ast.KindEndOfFile {
			break
		}
		end = s.TokenEnd()
		s.Scan()
	}

	return core.NewTextRange(start, end)
}

// openBracePos returns the position of the class body's opening `{`. We
// scan forward from the `class` keyword.
func openBracePos(node *ast.Node, sf *ast.SourceFile) int {
	s := scanner.GetScannerForSourceFile(sf, classKeywordPos(node, sf.Text()))
	for {
		kind := s.Token()
		if kind == ast.KindOpenBraceToken || kind == ast.KindEndOfFile {
			return s.TokenStart()
		}
		s.Scan()
	}
}

// staticKeywordToken finds the `static` modifier token on a class member.
// Returns the range of the keyword *token only* (leading trivia stripped).
// The caller must guarantee `static` is present (it is for any member that
// passed `isStaticMember`).
func staticKeywordToken(member *ast.Node, sourceText string) (core.TextRange, bool) {
	if mods := member.Modifiers(); mods != nil {
		for _, m := range mods.Nodes {
			if m != nil && m.Kind == ast.KindStaticKeyword {
				start := scanner.SkipTrivia(sourceText, m.Pos())
				return core.NewTextRange(start, m.End()), true
			}
		}
	}
	return core.TextRange{}, false
}

// skipWhitespaceForward returns the position of the first non-whitespace
// character (per JS `\s*`) at or after `pos`.
func skipWhitespaceForward(text string, pos int) int {
	for pos < len(text) {
		c := text[pos]
		switch c {
		case ' ', '\t', '\n', '\r', '\v', '\f':
			pos++
		default:
			// U+00A0 (NBSP) and other Unicode whitespace are rare in source
			// code; matching `\s` here is overkill — ASCII coverage is enough
			// for fix correctness on real-world code.
			return pos
		}
	}
	return pos
}

// findTrailingSemicolonInRange uses the scanner to locate a `;` token
// starting at `startPos`, bounded by `endPos`. The scanner skips both
// whitespace and comments, so this correctly handles cases like
// `static a /* c */;`.
func findTrailingSemicolonInRange(sf *ast.SourceFile, startPos, endPos int) (int, bool) {
	if startPos >= endPos {
		return 0, false
	}
	s := scanner.GetScannerForSourceFile(sf, startPos)
	if s.TokenStart() >= endPos {
		return 0, false
	}
	if s.Token() == ast.KindSemicolonToken {
		return s.TokenStart(), true
	}
	return 0, false
}

// findEqualsToken locates the `=` token of a PropertyDeclaration's
// initializer assignment. Scans tokens from after the property name. The
// computed-key form `[((c))] = ((2))` is handled correctly because we scan
// at token granularity, skipping over balanced `]`/`)` etc.
func findEqualsToken(member *ast.Node, sf *ast.SourceFile) (core.TextRange, bool) {
	startPos := member.Pos()
	if name := member.Name(); name != nil {
		startPos = name.End()
	}
	s := scanner.GetScannerForSourceFile(sf, startPos)
	for s.TokenStart() < member.End() {
		if s.Token() == ast.KindEqualsToken {
			return core.NewTextRange(s.TokenStart(), s.TokenEnd()), true
		}
		s.Scan()
	}
	return core.TextRange{}, false
}

// propertyBlocksClassFix reports whether a PropertyDeclaration blocks the
// class-level autofix.
//
// Upstream cases (mirrored as-is):
//   - TS type annotation (`static a: number = 1`) — not representable on an
//     object literal property.
//   - Raw initializer text includes `this` — upstream uses
//     `sourceCode.getText(value)` (no leading trivia), so we strip trivia
//     via `scanner.SkipTrivia` to align: `static a = /* this */ 1` is fine.
//
// rslint additions (suppress when fix output would be invalid TypeScript):
//   - PostfixToken `?` (optional). Object members can't be declared optional
//     (TS1162: "An object member cannot be declared optional.").
//   - PostfixToken `!` (definite assignment assertion). TS1255: "A definite
//     assignment assertion '!' is not permitted in this context."
//
// Upstream's typescript-eslint preserves these tokens in the fix output and
// produces invalid TS; rslint refuses the fix instead.
func propertyBlocksClassFix(member *ast.Node, text string) bool {
	if !ast.IsPropertyDeclaration(member) {
		return false
	}
	pd := member.AsPropertyDeclaration()
	if pd == nil {
		return false
	}
	if pd.Type != nil {
		return true
	}
	if pd.PostfixToken != nil {
		return true
	}
	if pd.Initializer != nil {
		init := pd.Initializer
		start := scanner.SkipTrivia(text, init.Pos())
		if strings.Contains(text[start:init.End()], "this") {
			return true
		}
	}
	return false
}

// memberContentEnd returns the position after the last "real" token of a
// member (excluding any trailing `;` for methods/accessors/constructors,
// which is a separate SemicolonClassElement in tsgo).
func memberContentEnd(member *ast.Node) int {
	switch {
	case ast.IsPropertyDeclaration(member):
		pd := member.AsPropertyDeclaration()
		if pd.Initializer != nil {
			return pd.Initializer.End()
		}
		if pd.Type != nil {
			return pd.Type.End()
		}
		if pd.PostfixToken != nil {
			return pd.PostfixToken.End()
		}
		if name := member.Name(); name != nil {
			return name.End()
		}
	}
	if body := member.Body(); body != nil {
		return body.End()
	}
	return member.End()
}

// findFollowingSemicolonElement returns the SemicolonClassElement that
// immediately follows `members[idx]`, if any.
func findFollowingSemicolonElement(members []*ast.Node, idx int) *ast.Node {
	if idx+1 < len(members) && members[idx+1] != nil && ast.IsSemicolonClassElement(members[idx+1]) {
		return members[idx+1]
	}
	return nil
}

// isExportDefault reports whether the class node is `export default class
// ...`. In tsgo the `export default` modifiers are on the class itself.
func isExportDefault(node *ast.Node) bool {
	flags := node.ModifierFlags()
	return flags&ast.ModifierFlagsExport != 0 && flags&ast.ModifierFlagsDefault != 0
}

// linePosition returns the 1-based line number of `pos` in `text`.
func linePosition(text string, pos int) int {
	line := 1
	for i := 0; i < pos && i < len(text); i++ {
		if text[i] == '\n' {
			line++
		}
	}
	return line
}

// trimmedTextBetween reports whether the source between `[start, end)` has
// any non-whitespace content (used to detect comments between `class` and
// `{` in the multi-line `return class /* c */\n{` upstream edge case).
func trimmedTextBetween(text string, start, end int) string {
	if start >= end || end > len(text) {
		return ""
	}
	return strings.TrimSpace(text[start:end])
}

// buildFix walks the class members and produces the list of fixes. Returns
// nil when the rule's autofix preconditions aren't met (declare/abstract/
// implements, named class expression, named export-default class, property
// with type annotation, or property with `this` in its initializer).
func buildFix(node *ast.Node, sf *ast.SourceFile) []rule.RuleFix {
	text := sf.Text()
	flags := node.ModifierFlags()

	// Upstream's `switchClassToObject` early-out:
	//   - declare class
	//   - abstract class
	//   - class with `implements`
	if flags&(ast.ModifierFlagsAmbient|ast.ModifierFlagsAbstract) != 0 {
		return nil
	}
	// `implements` heritage check.
	if hc := getHeritageClauses(node); hc != nil {
		for _, clause := range hc.Nodes {
			if clause != nil && clause.AsHeritageClause().Token == ast.KindImplementsKeyword {
				return nil
			}
		}
	}

	isClassExpr := node.Kind == ast.KindClassExpression
	hasName := node.Name() != nil
	exportDefault := isExportDefault(node)

	// Named class expression — upstream skips fix.
	if isClassExpr && hasName {
		return nil
	}
	// Named `export default class A` — upstream skips fix.
	if exportDefault && hasName {
		return nil
	}

	// Type parameters: rslint suppresses the autofix when the class declares
	// `<T, ...>`. Upstream's `switchClassToObject` blindly replaces `class`
	// with `const` and keeps the name+type-parameters region intact, which
	// produces `const A<T> = { ... }` — syntactically invalid TypeScript
	// (`const` declarations don't accept type parameters). The diagnostic is
	// still reported. Documented under "Differences from ESLint".
	if hasTypeParameters(node) {
		return nil
	}

	// Per-member autofix preconditions.
	for _, m := range node.Members() {
		if m == nil {
			continue
		}
		if propertyBlocksClassFix(m, text) {
			return nil
		}
		// `accessor` (TC39 auto-accessor) on a property — `{ accessor a:
		// ..., }` is not valid object literal syntax. Diagnostic stays.
		if ast.IsPropertyDeclaration(m) && ast.HasSyntacticModifier(m, ast.ModifierFlagsAccessor) {
			return nil
		}
		// Method-like member without a body (overload signature). Object
		// literal methods must have bodies; emitting `a(): void;,` would
		// be invalid TS. Diagnostic stays.
		if isClassElementMethodLike(m) && m.Body() == nil {
			return nil
		}
	}

	fixes := make([]rule.RuleFix, 0, 8)
	addFix := func(start, end int, replacement string) {
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, end), replacement))
	}

	classPos := classKeywordPos(node, text)
	classEnd := classPos + len("class")
	bracePos := openBracePos(node, sf)

	// Special-case: ClassExpression directly inside `return class /* c */\n{`
	// (or with line break) where there's a comment between `class` and `{`.
	// Matches upstream's `parent.type === 'ReturnStatement' && body line !=
	// parent line && trimmed text between class end and brace > 0` arm.
	if isClassExpr {
		parent := node.Parent
		bracePosLine := linePosition(text, bracePos)
		var parentLine int
		if parent != nil {
			parentLine = linePosition(text, parent.Pos())
		}
		isReturnSpecial := parent != nil &&
			parent.Kind == ast.KindReturnStatement &&
			bracePosLine != parentLine &&
			trimmedTextBetween(text, classEnd, bracePos) != ""

		if isReturnSpecial {
			// Replace `class` with `{`, then remove the original `{`.
			addFix(classPos, classEnd, "{")
			addFix(bracePos, bracePos+1, "")
		} else {
			// Anonymous class expression: drop `class` and any whitespace
			// after it (matches `removeSpacesAfter`).
			endRm := skipWhitespaceForward(text, classEnd)
			addFix(classPos, endRm, "")
		}
	} else if exportDefault {
		// `export default class { ... }` (anonymous) → drop `class` plus
		// trailing whitespace.
		endRm := skipWhitespaceForward(text, classEnd)
		addFix(classPos, endRm, "")
	} else {
		// `class A { ... }` (declaration) or `export class A { ... }` →
		// `const A = { ... };`. We keep any `export` modifier untouched,
		// only the `class` keyword is replaced.
		addFix(classPos, classEnd, "const")
		addFix(bracePos, bracePos, "= ")
		addFix(node.End(), node.End(), ";")
	}

	// Per-member transformations.
	members := node.Members()
	for i, m := range members {
		if m == nil || ast.IsSemicolonClassElement(m) {
			continue
		}
		// Remove `static` keyword + trailing whitespace. Use the keyword's
		// token range (leading trivia stripped) so we don't accidentally
		// eat the indentation BEFORE `static`.
		if loc, ok := staticKeywordToken(m, text); ok {
			endRm := skipWhitespaceForward(text, loc.End())
			addFix(loc.Pos(), endRm, "")
		}

		// PropertyDeclaration: replace `=` with `:`, OR insert `: undefined`.
		if ast.IsPropertyDeclaration(m) {
			pd := m.AsPropertyDeclaration()
			contentEnd := memberContentEnd(m)
			semiPos, hasSemi := findTrailingSemicolonInRange(sf, contentEnd, m.End())
			if pd.Initializer != nil {
				if eq, ok := findEqualsToken(m, sf); ok {
					addFix(eq.Pos(), eq.End(), ":")
				}
			} else if hasSemi {
				addFix(semiPos, semiPos, ": undefined")
			} else {
				addFix(contentEnd, contentEnd, ": undefined")
			}
			if hasSemi {
				addFix(semiPos, semiPos+1, ",")
			} else {
				addFix(contentEnd, contentEnd, ",")
			}
			continue
		}
		// Method / accessor / constructor: a separate SemicolonClassElement
		// after the method, if any, holds the `;`. Replace just the `;`
		// token (skip leading trivia so any comments between `}` and `;`
		// are preserved, matching ESLint's `replaceText(token, ',')`).
		if semi := findFollowingSemicolonElement(members, i); semi != nil {
			semiStart := scanner.SkipTrivia(text, semi.Pos())
			addFix(semiStart, semi.End(), ",")
			continue
		}
		// No `;` follows — append `,` after the method body.
		appendAt := m.End()
		if body := m.Body(); body != nil {
			appendAt = body.End()
		}
		addFix(appendAt, appendAt, ",")
	}

	return fixes
}

// hasTypeParameters reports whether the class declares `<T, ...>`.
func hasTypeParameters(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindClassDeclaration:
		tp := node.AsClassDeclaration().TypeParameters
		return tp != nil && len(tp.Nodes) > 0
	case ast.KindClassExpression:
		tp := node.AsClassExpression().TypeParameters
		return tp != nil && len(tp.Nodes) > 0
	}
	return false
}

// getHeritageClauses returns the heritage clause list for a class node.
// Inlined here to keep the rule self-contained without dragging the
// `internal/utils` import for one helper call.
func getHeritageClauses(node *ast.Node) *ast.NodeList {
	switch node.Kind {
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		return node.AsClassExpression().HeritageClauses
	}
	return nil
}

var NoStaticOnlyClassRule = rule.Rule{
	Name: "unicorn/no-static-only-class",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		check := func(node *ast.Node) {
			// Skip classes that extend another class — the static members
			// might be supplementing inherited instance members.
			if ast.GetExtendsHeritageClauseElement(node) != nil {
				return
			}

			// Skip class-level decorators.
			if ast.HasDecorators(node) {
				return
			}

			// Iterate members. SemicolonClassElement entries (a stray `;`
			// between members) have no ESTree analog — treat them as
			// whitespace. Any non-static-member disqualifies the class.
			hasMember := false
			for _, m := range node.Members() {
				if m == nil || ast.IsSemicolonClassElement(m) {
					continue
				}
				hasMember = true
				if !isStaticMember(m) {
					return
				}
			}
			if !hasMember {
				return
			}

			msg := rule.RuleMessage{
				Id:          "noStaticOnlyClass",
				Description: "Use an object instead of a class with only static members.",
			}
			headRange := classHeadRange(node, ctx.SourceFile)

			fixes := buildFix(node, ctx.SourceFile)
			if len(fixes) == 0 {
				ctx.ReportRange(headRange, msg)
				return
			}
			ctx.ReportRangeWithFixes(headRange, msg, fixes...)
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: check,
			ast.KindClassExpression:  check,
		}
	},
}
