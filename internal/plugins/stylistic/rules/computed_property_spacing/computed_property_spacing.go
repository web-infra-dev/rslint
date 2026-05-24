package computed_property_spacing

import (
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	spacingNever  = "never"
	spacingAlways = "always"
)

type options struct {
	spaced                 bool
	enforceForClassMembers bool
}

// parseOptions accepts upstream's two-argument shape:
//
//	['computed-property-spacing']                                → defaults
//	['computed-property-spacing', 'always' | 'never']            → mode only
//	['computed-property-spacing', mode, { enforceForClassMembers }]
//
// rslint's config loader unwraps single-element option arrays into a bare
// object, so the second-position object may also arrive as a top-level
// map[string]interface{}. Mode in that case defaults to 'never' and we honor
// the supplied enforceForClassMembers value.
//
// Upstream default: ['never', { enforceForClassMembers: true }] — the
// enforceForClassMembers default stays true regardless of whether the user
// supplies the object at all.
func parseOptions(raw any) options {
	opts := options{spaced: false, enforceForClassMembers: true}

	var arr []interface{}
	switch v := raw.(type) {
	case []interface{}:
		arr = v
	case string:
		arr = []interface{}{v}
	case map[string]interface{}:
		arr = []interface{}{spacingNever, v}
	}

	if len(arr) > 0 {
		if s, ok := arr[0].(string); ok && s == spacingAlways {
			opts.spaced = true
		}
	}
	if len(arr) > 1 {
		if m, ok := arr[1].(map[string]interface{}); ok {
			if b, ok := m["enforceForClassMembers"].(bool); ok {
				opts.enforceForClassMembers = b
			}
		}
	}
	return opts
}

// findOpenBracketFrom returns the byte position of the next `[` token starting
// at or after `searchStart`. Returns -1 when no `[` precedes EOF.
//
// Used by ElementAccessExpression / IndexedAccessType where `[` does not sit
// at the node's left edge (the receiver / object type comes first, and may be
// followed by trivia and — for optional chaining — a `?.` token). Using the
// scanner instead of a byte walk is what lets us skip across `?.[…]` without
// special-casing the optional-chain glyph.
func findOpenBracketFrom(sf *ast.SourceFile, searchStart int) int {
	s := scanner.GetScannerForSourceFile(sf, searchStart)
	for s.Token() != ast.KindEndOfFile {
		if s.Token() == ast.KindOpenBracketToken {
			return s.TokenStart()
		}
		s.Scan()
	}
	return -1
}

// findCloseBracketFrom returns the byte position of the next `]` token
// starting at or after `searchStart`. Returns -1 if none.
//
// Used instead of `node.End()-1` because some tsgo node kinds (notably
// ComputedPropertyName accessed through `parent.Name()`) carry End() values
// whose relationship to the literal `]` token is parser-version-dependent;
// the scanner is the authoritative source of token positions.
func findCloseBracketFrom(sf *ast.SourceFile, searchStart int) int {
	s := scanner.GetScannerForSourceFile(sf, searchStart)
	for s.Token() != ast.KindEndOfFile {
		if s.Token() == ast.KindCloseBracketToken {
			return s.TokenStart()
		}
		s.Scan()
	}
	return -1
}

// resolved is the per-node working set produced by the resolvers below.
// ok=false signals the listener should skip this node (parser recovery,
// missing inner expression, or — for computed-name parents — the name is
// not actually a ComputedPropertyName).
type resolved struct {
	openPos, closePos int
	ok                bool
}

func validRange(text string, r resolved) bool {
	if !r.ok || r.openPos < 0 || r.closePos < 0 || r.openPos+1 > r.closePos {
		return false
	}
	if r.openPos >= len(text) || r.closePos >= len(text) {
		return false
	}
	return text[r.openPos] == '[' && text[r.closePos] == ']'
}

// pendingReport is one buffered diagnostic. We collect these as the listener
// walk progresses, then sort by sortKey (source position) and emit at the
// outermost exit. See ComputedPropertySpacingRule.Run for why buffering is
// necessary — the rslint test framework does not sort diagnostics, and our
// AST-traversal order doesn't always match source order (chained ElementAccess
// like `obj[a][b]` visits outer first but the outer `[` token lives AFTER the
// inner `]` in source).
type pendingReport struct {
	sortKey int
	rng     core.TextRange
	msg     rule.RuleMessage
	fix     rule.RuleFix
}

// isObjectLiteralGrandparent reports whether the member's enclosing container
// is an ObjectLiteralExpression (`{ [k](){} }`) versus a class body
// (`class A { [k](){} }`). Method / GetAccessor / SetAccessor share a kind
// across object literals and classes in tsgo, so grandparent dispatch is the
// only way to tell them apart for the enforceForClassMembers gate.
func isObjectLiteralGrandparent(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	return node.Parent.Kind == ast.KindObjectLiteralExpression
}

// isAbstractMember reports whether `node` carries the `abstract` modifier.
// Upstream's stylistic rule listens on `MethodDefinition` / `PropertyDefinition`
// only — typescript-eslint represents abstract members as the DISTINCT AST
// kinds `TSAbstractMethodDefinition` / `TSAbstractPropertyDefinition`, which
// upstream does NOT listen on, so abstract members are silently skipped. tsgo
// keeps the same MethodDeclaration / PropertyDeclaration kind for both
// abstract and concrete members and distinguishes them via a modifier flag —
// so to stay aligned we must skip abstract members explicitly.
func isAbstractMember(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract)
}

// inJSDoc reports whether `node` was parsed inside a JSDoc `@type {...}` (or
// similar) annotation. tsgo parses JSDoc type expressions into the regular
// type AST (`IndexedAccessTypeNode`, etc.) with the `NodeFlagsJSDoc` flag
// set; typescript-eslint's parser treats JSDoc comments as plain trivia and
// never builds AST for their interior. Upstream stylistic therefore never
// sees `[K]` inside `/** @type {Foo[K]} */`, so we must skip JSDoc-flagged
// nodes too to keep the rule's user-visible output aligned.
func inJSDoc(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Flags&ast.NodeFlagsJSDoc != 0
}

// ComputedPropertySpacingRule enforces consistent spacing inside computed
// property brackets — object-literal keys, destructuring patterns, class
// member names, indexed access types, and computed member access (including
// optional chaining `obj?.[x]`). Ported from @stylistic/eslint-plugin's
// computed-property-spacing.
//
// Listener dispatch mirrors no_useless_computed_key: we listen on the parent
// kinds (PropertyAssignment / BindingElement / MethodDeclaration /
// GetAccessor / SetAccessor / PropertyDeclaration) rather than on
// KindComputedPropertyName directly. ElementAccessExpression and
// IndexedAccessType get their own listeners.
//
// Reports are buffered and emitted in source-position order at the
// outermost exit (when our listener-kind nesting depth drops back to 0).
// Without buffering, chained ElementAccess (`obj[a][b]`) would emit outer's
// brackets before inner's, because tsgo visits the outer ElementAccess node
// first even though its `[` token sits AFTER the inner `]` in source. The
// rslint test framework does not re-sort diagnostics, so the rule itself
// must guarantee source order.
var ComputedPropertySpacingRule = rule.Rule{
	Name: "@stylistic/computed-property-spacing",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		sf := ctx.SourceFile
		text := sf.Text()

		// --- buffered emission state ---
		// `depth` counts how many of our LISTENED kinds we're currently
		// inside (across all kinds, not per-kind). When depth returns to
		// zero we've finished the topmost subtree of listened kinds and
		// can flush — every report inside belongs to that subtree, so
		// sorting and emitting is order-preserving for the whole tree.
		var depth int
		var pending []pendingReport

		addReport := func(sortKey int, rng core.TextRange, msg rule.RuleMessage, fix rule.RuleFix) {
			pending = append(pending, pendingReport{sortKey: sortKey, rng: rng, msg: msg, fix: fix})
		}

		flush := func() {
			if len(pending) == 0 {
				return
			}
			sort.SliceStable(pending, func(i, j int) bool {
				return pending[i].sortKey < pending[j].sortKey
			})
			for _, p := range pending {
				ctx.ReportRangeWithFixes(p.rng, p.msg, p.fix)
			}
			pending = pending[:0]
		}

		// --- opening / closing analyzers (buffer instead of emitting) ---

		collectOpening := func(r resolved) {
			innerLow := r.openPos + 1
			if innerLow > r.closePos {
				return
			}
			firstStart := utils.SkipLeadingWhitespace(text, innerLow, r.closePos)
			// `isTokenOnSameLine(before, first)` short-circuit — skip the
			// opening check when a newline separates `[` from its first
			// inner token/comment.
			if utils.ContainsLineTerminator(text, innerLow, firstStart) {
				return
			}
			hasSpace := firstStart > innerLow
			if opts.spaced {
				if !hasSpace {
					addReport(
						r.openPos,
						core.NewTextRange(r.openPos, r.openPos+1),
						rule.RuleMessage{
							Id:          "missingSpaceAfter",
							Description: "A space is required after '['.",
							Data:        map[string]string{"tokenValue": "["},
						},
						rule.RuleFix{
							Text:  " ",
							Range: core.NewTextRange(r.openPos+1, r.openPos+1),
						},
					)
				}
				return
			}
			if hasSpace {
				addReport(
					innerLow,
					core.NewTextRange(innerLow, firstStart),
					rule.RuleMessage{
						Id:          "unexpectedSpaceAfter",
						Description: "There should be no space after '['.",
						Data:        map[string]string{"tokenValue": "["},
					},
					rule.RuleFix{
						Text:  "",
						Range: core.NewTextRange(innerLow, firstStart),
					},
				)
			}
		}

		collectClosing := func(r resolved) {
			innerLow := r.openPos + 1
			if innerLow > r.closePos {
				return
			}
			lastEnd := utils.SkipTrailingWhitespace(text, innerLow, r.closePos)
			if utils.ContainsLineTerminator(text, lastEnd, r.closePos) {
				return
			}
			hasSpace := lastEnd < r.closePos
			if opts.spaced {
				if !hasSpace {
					addReport(
						r.closePos,
						core.NewTextRange(r.closePos, r.closePos+1),
						rule.RuleMessage{
							Id:          "missingSpaceBefore",
							Description: "A space is required before ']'.",
							Data:        map[string]string{"tokenValue": "]"},
						},
						rule.RuleFix{
							Text:  " ",
							Range: core.NewTextRange(r.closePos, r.closePos),
						},
					)
				}
				return
			}
			if hasSpace {
				addReport(
					lastEnd,
					core.NewTextRange(lastEnd, r.closePos),
					rule.RuleMessage{
						Id:          "unexpectedSpaceBefore",
						Description: "There should be no space before ']'.",
						Data:        map[string]string{"tokenValue": "]"},
					},
					rule.RuleFix{
						Text:  "",
						Range: core.NewTextRange(lastEnd, r.closePos),
					},
				)
			}
		}

		// --- resolvers: turn each listener kind into (openPos, closePos) ---

		resolveElementAccess := func(node *ast.Node) resolved {
			eae := node.AsElementAccessExpression()
			if eae == nil || eae.Expression == nil || eae.ArgumentExpression == nil {
				return resolved{}
			}
			openPos := findOpenBracketFrom(sf, eae.Expression.End())
			closePos := findCloseBracketFrom(sf, eae.ArgumentExpression.End())
			return resolved{openPos: openPos, closePos: closePos, ok: true}
		}

		resolveIndexedAccessType := func(node *ast.Node) resolved {
			iat := node.AsIndexedAccessTypeNode()
			if iat == nil || iat.ObjectType == nil || iat.IndexType == nil {
				return resolved{}
			}
			openPos := findOpenBracketFrom(sf, iat.ObjectType.End())
			closePos := findCloseBracketFrom(sf, iat.IndexType.End())
			return resolved{openPos: openPos, closePos: closePos, ok: true}
		}

		resolveComputedNameFromNode := func(name *ast.Node) resolved {
			if name == nil || name.Kind != ast.KindComputedPropertyName {
				return resolved{}
			}
			cpn := name.AsComputedPropertyName()
			if cpn == nil || cpn.Expression == nil {
				return resolved{}
			}
			openPos := findOpenBracketFrom(sf, name.Pos())
			closePos := findCloseBracketFrom(sf, cpn.Expression.End())
			return resolved{openPos: openPos, closePos: closePos, ok: true}
		}

		resolveFromName := func(node *ast.Node) resolved {
			return resolveComputedNameFromNode(node.Name())
		}

		resolveBindingElement := func(node *ast.Node) resolved {
			be := node.AsBindingElement()
			if be == nil {
				return resolved{}
			}
			return resolveComputedNameFromNode(be.PropertyName)
		}

		// --- gate helpers: filter on enforceForClassMembers ---

		alwaysCheck := func(resolver func(*ast.Node) resolved) func(*ast.Node) resolved {
			return resolver
		}

		classMemberGated := func(resolver func(*ast.Node) resolved) func(*ast.Node) resolved {
			return func(node *ast.Node) resolved {
				if !opts.enforceForClassMembers {
					return resolved{}
				}
				if isAbstractMember(node) {
					return resolved{} // see isAbstractMember docstring
				}
				return resolver(node)
			}
		}

		// containerGated handles MethodDeclaration / GetAccessor / SetAccessor —
		// always check on object literals, gate on enforceForClassMembers when
		// the parent is a class body. Returns empty for any other parent kind
		// (e.g. InterfaceDeclaration / TypeLiteral — defensive; tsgo uses
		// MethodSignature there, a distinct kind).
		containerGated := func(resolver func(*ast.Node) resolved) func(*ast.Node) resolved {
			return func(node *ast.Node) resolved {
				if isObjectLiteralGrandparent(node) {
					return resolver(node)
				}
				parent := node.Parent
				if parent == nil {
					return resolved{}
				}
				if parent.Kind != ast.KindClassDeclaration && parent.Kind != ast.KindClassExpression {
					return resolved{}
				}
				if !opts.enforceForClassMembers {
					return resolved{}
				}
				if isAbstractMember(node) {
					return resolved{} // see isAbstractMember docstring
				}
				return resolver(node)
			}
		}

		// --- listener factories ---
		// enter: bump depth, then collect reports if this node is in scope.
		// exit:  decrement depth; if back to 0, flush sorted.

		enter := func(resolver func(*ast.Node) resolved) func(*ast.Node) {
			return func(node *ast.Node) {
				depth++
				if inJSDoc(node) {
					return // see inJSDoc docstring — upstream parser drops JSDoc internals
				}
				r := resolver(node)
				if validRange(text, r) {
					collectOpening(r)
					collectClosing(r)
				}
			}
		}
		exit := func(node *ast.Node) {
			depth--
			if depth == 0 {
				flush()
			}
		}

		// --- per-kind gated resolvers ---

		eaR := alwaysCheck(resolveElementAccess)
		iatR := alwaysCheck(resolveIndexedAccessType)
		paR := alwaysCheck(resolveFromName)     // object-literal PropertyAssignment
		beR := alwaysCheck(resolveBindingElement)
		mdR := containerGated(resolveFromName)  // method (object literal OR class)
		gaR := containerGated(resolveFromName)  // get accessor
		saR := containerGated(resolveFromName)  // set accessor
		pdR := classMemberGated(resolveFromName) // PropertyDeclaration (always class)

		return rule.RuleListeners{
			ast.KindElementAccessExpression:                      enter(eaR),
			rule.ListenerOnExit(ast.KindElementAccessExpression): exit,
			ast.KindIndexedAccessType:                            enter(iatR),
			rule.ListenerOnExit(ast.KindIndexedAccessType):       exit,
			ast.KindPropertyAssignment:                           enter(paR),
			rule.ListenerOnExit(ast.KindPropertyAssignment):      exit,
			ast.KindBindingElement:                               enter(beR),
			rule.ListenerOnExit(ast.KindBindingElement):          exit,
			ast.KindMethodDeclaration:                            enter(mdR),
			rule.ListenerOnExit(ast.KindMethodDeclaration):       exit,
			ast.KindGetAccessor:                                  enter(gaR),
			rule.ListenerOnExit(ast.KindGetAccessor):             exit,
			ast.KindSetAccessor:                                  enter(saR),
			rule.ListenerOnExit(ast.KindSetAccessor):             exit,
			ast.KindPropertyDeclaration:                          enter(pdR),
			rule.ListenerOnExit(ast.KindPropertyDeclaration):     exit,
		}
	},
}
