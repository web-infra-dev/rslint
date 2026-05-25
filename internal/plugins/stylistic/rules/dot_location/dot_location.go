package dot_location

import (
	"regexp"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/stylisticutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const locationProperty = "property"

type options struct {
	onObject bool
}

// parseOptions accepts upstream's single-argument shape:
//
//	['dot-location']                       → defaults (onObject = true)
//	['dot-location', 'object' | 'property']
//
// rslint's config loader may unwrap a single-element option array into a bare
// string, so we accept the bare form too.
func parseOptions(raw any) options {
	opts := options{onObject: true}
	var arr []interface{}
	switch v := raw.(type) {
	case []interface{}:
		arr = v
	case string:
		arr = []interface{}{v}
	}
	if len(arr) > 0 {
		if s, ok := arr[0].(string); ok && s == locationProperty {
			opts.onObject = false
		}
	}
	return opts
}

// decimalIntegerPattern mirrors ESLint's isDecimalIntegerNumericToken regex
// (used by the autofix to decide whether `5\n.x` needs a space when collapsed
// to `5 .x` — without the space, `5.` would re-tokenize as a decimal float).
//
//	`0`                  — single zero
//	`0[0-7]*[89][0-9]*`  — starts with `0` but contains `8` or `9` (so it
//	                       can't be parsed as legacy octal — pure decimal)
//	`[1-9](_?[0-9])*`    — starts with 1-9, optional `_` numeric separators
//
// Hex (`0x`), binary (`0b`), octal-explicit (`0o`), legacy-octal (`0[0-7]+`),
// exponent (`1e3`), and float-with-dot (`5.0`) all fail this pattern — none
// of them would token-fuse with a trailing `.` to produce a different parse,
// so no space is needed in the fix.
var decimalIntegerPattern = regexp.MustCompile(`^(?:0|0[0-7]*[89][0-9]*|[1-9](?:_?[0-9])*)$`)

// dotInfo holds the resolved dot location and the immediately-preceding
// non-trivia token end position for one node visit. `ok=false` signals the
// listener should skip this node (parser recovery, missing child, or the
// optional `Qualifier` of an ImportType being absent).
type dotInfo struct {
	dotStart     int
	dotEnd       int
	dotText      string // "." or "?."
	prevTokenEnd int    // end of the receiver/qualifier token preceding the dot
	isDecInt     bool   // prev token is a decimal-integer NumericLiteral
	propPos      int    // real (post-trivia) start of the property identifier
	ok           bool
}

// findDotInRange scans tokens forward from `scanFrom` and returns the first
// `.` or `?.` token whose start lies strictly before `propPos`.
//
// **Pre-condition (critical)**: `scanFrom` MUST sit OUTSIDE any receiver
// expression that could contain a template literal. tsgo's raw scanner does
// not track template-substitution brace depth — if started inside a template
// substitution (or just after one without re-entering the outer template's
// scan state), it can mis-tokenize a closing backtick as the start of a new
// template literal and silently eat the surrounding source. Concretely:
// starting at `Expression.End()` / `Left.End()` / `Attributes.End()` is
// safe; starting at `Expression.Pos()` is NOT (would re-scan the receiver
// from scratch and get confused on `\`${x}\``-containing arguments). This
// pre-condition is encoded in each per-kind resolver below.
func findDotInRange(sf *ast.SourceFile, scanFrom, propPos int) (dotStart, dotEnd int, dotText string, ok bool) {
	if scanFrom >= propPos {
		return 0, 0, "", false
	}
	s := scanner.GetScannerForSourceFile(sf, scanFrom)
	for s.Token() != ast.KindEndOfFile {
		start := s.TokenStart()
		if start >= propPos {
			return 0, 0, "", false
		}
		switch s.Token() {
		case ast.KindDotToken:
			return start, s.TokenEnd(), ".", true
		case ast.KindQuestionDotToken:
			return start, s.TokenEnd(), "?.", true
		}
		s.Scan()
	}
	return 0, 0, "", false
}

// DotLocationRule enforces consistent newline placement around `.` in member
// expressions, type qualifiers, JSX tag names, meta-properties, and import
// types. Ported from @stylistic/eslint-plugin's dot-location.
//
// Listener coverage:
//   - PropertyAccessExpression — covers `obj.prop`, `obj?.prop`, `this.#a`,
//     and JSX `<Form.Input />` (tsgo represents the JSX tag name as a
//     PropertyAccessExpression — see `JsxTagNameExpression` in the tsgo AST).
//   - MetaProperty            — covers `import.meta`, `new.target`.
//   - QualifiedName           — TypeScript type-position `A.B`.
//   - ImportType              — `import('foo').Qualifier`.
//
// ElementAccessExpression (`obj['prop']`) is intentionally NOT listened to —
// upstream skips MemberExpression when `node.computed === true`, and the
// computed form lives in a distinct tsgo node kind so we simply don't
// register a listener for it.
var DotLocationRule = rule.Rule{
	Name: "@stylistic/dot-location",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		sf := ctx.SourceFile

		// propTokenStart skips leading trivia from a child node's full-start
		// (`node.Pos()`, which includes the trivia between its previous
		// sibling/token and its first real token). The dot/property line
		// comparison must use the token start, not the trivia start —
		// otherwise `obj.\nprop` (where `prop.Pos()` lands ON the `\n`)
		// compares dot.End to a position on the same line as the dot and
		// silently passes.
		propTokenStart := func(name *ast.Node) int {
			return scanner.SkipTrivia(sf.Text(), name.Pos())
		}

		resolvePropAccess := func(node *ast.Node) dotInfo {
			pae := node.AsPropertyAccessExpression()
			if pae == nil || pae.Expression == nil {
				return dotInfo{}
			}
			name := node.Name()
			if name == nil {
				return dotInfo{}
			}
			propPos := propTokenStart(name)
			exprEnd := pae.Expression.End()
			ds, de, dt, ok := findDotInRange(sf, exprEnd, propPos)
			if !ok {
				return dotInfo{}
			}
			info := dotInfo{
				dotStart:     ds,
				dotEnd:       de,
				dotText:      dt,
				prevTokenEnd: exprEnd,
				propPos:      propPos,
				ok:           true,
			}
			// Decimal-integer autofix gate: only when the receiver is a bare
			// NumericLiteral whose source text is a pure decimal integer.
			// `5\n.x` (NumericLiteral receiver) → space needed: `5 .x`.
			// `(5)\n.x` (ParenthesizedExpression receiver) → no space:
			// `(5).x` re-parses cleanly. The Expression.Kind check makes
			// this distinction directly from the AST, avoiding the need
			// to backward-scan tokens.
			if pae.Expression.Kind == ast.KindNumericLiteral {
				realStart := scanner.SkipTrivia(sf.Text(), pae.Expression.Pos())
				if realStart < exprEnd {
					info.isDecInt = decimalIntegerPattern.MatchString(sf.Text()[realStart:exprEnd])
				}
			}
			return info
		}

		resolveQualifiedName := func(node *ast.Node) dotInfo {
			qn := node.AsQualifiedName()
			if qn == nil || qn.Left == nil || qn.Right == nil {
				return dotInfo{}
			}
			propPos := propTokenStart(qn.Right)
			leftEnd := qn.Left.End()
			ds, de, dt, ok := findDotInRange(sf, leftEnd, propPos)
			if !ok {
				return dotInfo{}
			}
			return dotInfo{
				dotStart:     ds,
				dotEnd:       de,
				dotText:      dt,
				prevTokenEnd: leftEnd,
				propPos:      propPos,
				ok:           true,
			}
		}

		resolveImportType := func(node *ast.Node) dotInfo {
			it := node.AsImportTypeNode()
			if it == nil || it.Qualifier == nil || it.Argument == nil {
				return dotInfo{}
			}
			propPos := propTokenStart(it.Qualifier)
			// Start scanning just past the argument (or attributes, when
			// present). The interval between this point and the qualifier
			// contains only `)`, the dot, optional commas, and trivia — no
			// template literal context — so the raw scanner re-enters
			// safely.
			afterArg := it.Argument.End()
			if it.Attributes != nil {
				afterArg = it.Attributes.End()
			}
			// Walk forward to locate the closing `)` (whose line is the
			// "object-side" line for this dot) and then the dot itself.
			info := dotInfo{propPos: propPos}
			s := scanner.GetScannerForSourceFile(sf, afterArg)
			closeParenEnd := -1
			for s.Token() != ast.KindEndOfFile {
				start := s.TokenStart()
				if start >= propPos {
					return info
				}
				tok := s.Token()
				if tok == ast.KindCloseParenToken {
					closeParenEnd = s.TokenEnd()
				}
				if tok == ast.KindDotToken {
					info.dotStart = start
					info.dotEnd = s.TokenEnd()
					info.dotText = "."
					info.prevTokenEnd = closeParenEnd
					info.ok = closeParenEnd >= 0
					return info
				}
				s.Scan()
			}
			return info
		}

		resolveMetaProperty := func(node *ast.Node) dotInfo {
			mp := node.AsMetaProperty()
			if mp == nil {
				return dotInfo{}
			}
			name := node.Name()
			if name == nil {
				return dotInfo{}
			}
			propPos := propTokenStart(name)
			// `import.meta` / `new.target`: scanning from node.Pos() is
			// safe — only a keyword (`import`/`new`), a `.`, and the name
			// identifier live in this range; no template-literal context
			// to confuse the scanner.
			info := dotInfo{propPos: propPos}
			s := scanner.GetScannerForSourceFile(sf, node.Pos())
			prevEnd := -1
			for s.Token() != ast.KindEndOfFile {
				start := s.TokenStart()
				if start >= propPos {
					return info
				}
				if s.Token() == ast.KindDotToken {
					info.dotStart = start
					info.dotEnd = s.TokenEnd()
					info.dotText = "."
					info.prevTokenEnd = prevEnd
					info.ok = prevEnd >= 0
					return info
				}
				prevEnd = s.TokenEnd()
				s.Scan()
			}
			return info
		}

		// pendingReport buffers one diagnostic so we can emit them in source
		// position order at the outermost exit. See depth/flush logic below
		// for why buffering is necessary.
		type pendingReport struct {
			rng  core.TextRange
			msg  rule.RuleMessage
			fix  rule.RuleFix
			sort int
		}
		var depth int
		var pending []pendingReport

		addReport := func(info dotInfo, msg rule.RuleMessage, fix rule.RuleFix) {
			pending = append(pending, pendingReport{
				rng:  core.NewTextRange(info.dotStart, info.dotEnd),
				msg:  msg,
				fix:  fix,
				sort: info.dotStart,
			})
		}

		flush := func() {
			if len(pending) == 0 {
				return
			}
			sort.SliceStable(pending, func(i, j int) bool {
				return pending[i].sort < pending[j].sort
			})
			for _, p := range pending {
				ctx.ReportRangeWithFixes(p.rng, p.msg, p.fix)
			}
			pending = pending[:0]
		}

		check := func(info dotInfo) {
			if !info.ok {
				return
			}
			if opts.onObject {
				if stylisticutil.SameLineByPos(sf, info.prevTokenEnd, info.dotStart) {
					return
				}
				// Fix: move dot up to sit right after the previous token,
				// keeping all surrounding trivia intact.
				//
				// ESLint's two-op fix (insertTextAfter(prev, dotText) +
				// remove(dot)) reduces to a single replace from prevEnd to
				// dotEnd:
				//   original: <trivia_a><dot>
				//   replaced: <insertText><trivia_a>
				// where insertText prepends a space when the prev token is a
				// decimal integer literal and dotText starts with `.` —
				// otherwise `5\n.x` → `5.x` would re-lex `5.` as a float.
				insertText := info.dotText
				if info.dotText[0] == '.' && info.isDecInt {
					insertText = " " + insertText
				}
				triviaA := sf.Text()[info.prevTokenEnd:info.dotStart]
				addReport(info,
					rule.RuleMessage{
						Id:          "expectedDotAfterObject",
						Description: "Expected dot to be on same line as object.",
					},
					rule.RuleFix{
						Text:  insertText + triviaA,
						Range: core.NewTextRange(info.prevTokenEnd, info.dotEnd),
					},
				)
				return
			}
			if stylisticutil.SameLineByPos(sf, info.dotEnd, info.propPos) {
				return
			}
			// Fix: move dot down to sit right before the property.
			//   original: <dot><trivia_b>
			//   replaced: <trivia_b><dotText>
			triviaB := sf.Text()[info.dotEnd:info.propPos]
			addReport(info,
				rule.RuleMessage{
					Id:          "expectedDotBeforeProperty",
					Description: "Expected dot to be on same line as property.",
				},
				rule.RuleFix{
					Text:  triviaB + info.dotText,
					Range: core.NewTextRange(info.dotStart, info.propPos),
				},
			)
		}

		// Enter: bump depth, resolve and buffer any report for this node.
		// Exit:  decrement depth; flush sorted when we return to depth 0
		//        (i.e. we've finished the outermost listened subtree).
		//
		// Buffering is needed because pre-order tsgo traversal does NOT
		// match source order for our listener mix:
		//   - Chained PAE: `a.b.c` → outer's `.c` visits first but its dot
		//     sits AFTER the inner `a.b` dot in source.
		//   - ImportType + QualifiedName: `import('m').A.B` → outer
		//     ImportType visits first AND its dot sits before the inner
		//     QualifiedName dot — pre-order happens to be correct.
		// Neither pure pre-order nor pure post-order works for both shapes,
		// so we collect everything and sort by dot position at the end.
		enter := func(resolver func(*ast.Node) dotInfo) func(*ast.Node) {
			return func(node *ast.Node) {
				depth++
				check(resolver(node))
			}
		}
		exit := func(_ *ast.Node) {
			depth--
			if depth == 0 {
				flush()
			}
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression:                      enter(resolvePropAccess),
			rule.ListenerOnExit(ast.KindPropertyAccessExpression): exit,
			ast.KindMetaProperty:                                  enter(resolveMetaProperty),
			rule.ListenerOnExit(ast.KindMetaProperty):             exit,
			ast.KindQualifiedName:                                 enter(resolveQualifiedName),
			rule.ListenerOnExit(ast.KindQualifiedName):            exit,
			ast.KindImportType:                                    enter(resolveImportType),
			rule.ListenerOnExit(ast.KindImportType):               exit,
		}
	},
}
