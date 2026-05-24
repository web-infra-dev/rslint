package generator_star_spacing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/stylisticutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	msgMissingBefore    = "missingBefore"
	msgMissingAfter     = "missingAfter"
	msgUnexpectedBefore = "unexpectedBefore"
	msgUnexpectedAfter  = "unexpectedAfter"

	descMissingBefore    = "Missing space before *."
	descMissingAfter     = "Missing space after *."
	descUnexpectedBefore = "Unexpected space before *."
	descUnexpectedAfter  = "Unexpected space after *."
)

type sides struct {
	before, after bool
}

type kindModes struct {
	named, anonymous, method, shorthand sides
}

// stringPresets mirrors upstream's `optionDefinitions`. Unknown strings fall
// back to the supplied defaults (matches upstream where `optionDefinitions[s]`
// would be `undefined` and `Object.assign({}, defaults, undefined)` keeps the
// defaults).
var stringPresets = map[string]sides{
	"before":  {before: true, after: false},
	"after":   {before: false, after: true},
	"both":    {before: true, after: true},
	"neither": {before: false, after: false},
}

// optionToDefinition mirrors upstream's `optionToDefinition`:
//   - nil → defaults
//   - string → preset lookup (unknown string → defaults)
//   - map → Object.assign({}, defaults, {before, after})
func optionToDefinition(option any, defaults sides) sides {
	if option == nil {
		return defaults
	}
	if s, ok := option.(string); ok {
		if preset, present := stringPresets[s]; present {
			return preset
		}
		return defaults
	}
	if m, ok := option.(map[string]any); ok {
		out := defaults
		if v, ok := m["before"].(bool); ok {
			out.before = v
		}
		if v, ok := m["after"].(bool); ok {
			out.after = v
		}
		return out
	}
	return defaults
}

// parseOptions resolves the per-kind modes. Mirrors upstream's create()
// option-resolution block exactly, including the `shorthand ?? method`
// fallback so an explicit `method` override propagates to shorthand methods
// when shorthand is not separately specified.
func parseOptions(raw any) kindModes {
	var top any
	switch v := raw.(type) {
	case nil:
		top = nil
	case []any:
		if len(v) > 0 {
			top = v[0]
		}
	default:
		top = raw
	}

	rootDefault := stringPresets["before"]
	defaults := optionToDefinition(top, rootDefault)

	var namedOpt, anonOpt, methodOpt, shorthandOpt any
	if m, ok := top.(map[string]any); ok {
		namedOpt = m["named"]
		anonOpt = m["anonymous"]
		methodOpt = m["method"]
		shorthandOpt = m["shorthand"]
	}
	if shorthandOpt == nil {
		shorthandOpt = methodOpt
	}

	return kindModes{
		named:     optionToDefinition(namedOpt, defaults),
		anonymous: optionToDefinition(anonOpt, defaults),
		method:    optionToDefinition(methodOpt, defaults),
		shorthand: optionToDefinition(shorthandOpt, defaults),
	}
}

// asteriskNode returns the generator `*` token field for a function-like node,
// or nil if the node is not a generator (or not one of the function-like kinds
// we handle).
func asteriskNode(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken
	}
	return nil
}

// functionName returns the identifier-name node of a FunctionDeclaration /
// FunctionExpression, or nil when the function is anonymous.
func functionName(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Name()
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Name()
	}
	return nil
}

// hasAnyModifier reports whether the node carries at least one modifier
// keyword (static, async, public, override, …). For MethodDeclaration this is
// equivalent to "the `*` is NOT the first real token of the method node" —
// which is the condition upstream uses to gate the before-side check via
// `starToken === sourceCode.getFirstToken(node.parent)`.
func hasAnyModifier(node *ast.Node) bool {
	mods := node.Modifiers()
	return mods != nil && len(mods.Nodes) > 0
}

// starTokenRange returns the precise [pos, end) byte range of the generator
// `*` token. tsgo's raw `Pos()` on the AsteriskToken includes any leading
// trivia; GetRangeOfTokenAtPosition lands on the actual punctuator.
func starTokenRange(sf *ast.SourceFile, asterisk *ast.Node) (start, end int, ok bool) {
	if asterisk == nil {
		return 0, 0, false
	}
	rng := scanner.GetRangeOfTokenAtPosition(sf, asterisk.Pos())
	if rng.End() <= rng.Pos() {
		return 0, 0, false
	}
	return rng.Pos(), rng.End(), true
}

// nextRealTokenStart returns the start position of the next non-trivia token
// at or after `from`. Returns (-1, false) on end-of-file.
func nextRealTokenStart(sf *ast.SourceFile, from int) (int, bool) {
	s := scanner.GetScannerForSourceFile(sf, from)
	if s.Token() == ast.KindEndOfFile {
		return -1, false
	}
	return s.TokenStart(), true
}

// reportBefore emits the missing/unexpected diagnostic for the before side.
// Anchors the report range on the `*` token itself (matches upstream where
// `node = rightToken` and `rightToken === starToken` for the before side).
func reportBefore(
	ctx rule.RuleContext,
	spaced bool,
	required bool,
	prevEnd int,
	starStart, starEnd int,
) {
	if spaced == required {
		return
	}
	reportRange := core.NewTextRange(starStart, starEnd)
	if required {
		ctx.ReportRangeWithFixes(
			reportRange,
			rule.RuleMessage{Id: msgMissingBefore, Description: descMissingBefore},
			rule.RuleFix{Text: " ", Range: core.NewTextRange(starStart, starStart)},
		)
	} else {
		ctx.ReportRangeWithFixes(
			reportRange,
			rule.RuleMessage{Id: msgUnexpectedBefore, Description: descUnexpectedBefore},
			rule.RuleFix{Text: "", Range: core.NewTextRange(prevEnd, starStart)},
		)
	}
}

// reportAfter emits the missing/unexpected diagnostic for the after side.
// Anchored on the `*` (matches upstream where `node = leftToken` and
// `leftToken === starToken` for the after side).
func reportAfter(
	ctx rule.RuleContext,
	spaced bool,
	required bool,
	nextStart int,
	starStart, starEnd int,
) {
	if spaced == required {
		return
	}
	reportRange := core.NewTextRange(starStart, starEnd)
	if required {
		ctx.ReportRangeWithFixes(
			reportRange,
			rule.RuleMessage{Id: msgMissingAfter, Description: descMissingAfter},
			rule.RuleFix{Text: " ", Range: core.NewTextRange(starEnd, starEnd)},
		)
	} else {
		ctx.ReportRangeWithFixes(
			reportRange,
			rule.RuleMessage{Id: msgUnexpectedAfter, Description: descUnexpectedAfter},
			rule.RuleFix{Text: "", Range: core.NewTextRange(starEnd, nextStart)},
		)
	}
}

// checkGenerator runs the full before/after pair for one function-like node.
// `kind` selects the mode (named / anonymous / method / shorthand); the
// before-side check is skipped for method/shorthand when there is no
// preceding modifier, matching upstream's
// `(method||shorthand) && star === firstToken(parent)` short-circuit.
func checkGenerator(ctx rule.RuleContext, node *ast.Node, kind string, modes kindModes) {
	asterisk := asteriskNode(node)
	if asterisk == nil {
		return
	}
	starStart, starEnd, ok := starTokenRange(ctx.SourceFile, asterisk)
	if !ok {
		return
	}

	var mode sides
	switch kind {
	case "named":
		mode = modes.named
	case "anonymous":
		mode = modes.anonymous
	case "method":
		mode = modes.method
	case "shorthand":
		mode = modes.shorthand
	}

	// Skip the before-check when the `*` is the first real token of a
	// method/shorthand (upstream's `(method||shorthand) && star ===
	// firstToken(parent)` short-circuit). Equivalent here: only skip when
	// kind is method/shorthand AND the node has no leading modifier (incl.
	// decorators, which tsgo also stores in Modifiers()).
	isMethodLike := kind == "method" || kind == "shorthand"
	if !isMethodLike || hasAnyModifier(node) {
		prevEnd, ok := stylisticutil.FindPrevTokenEnd(ctx.SourceFile, node.Pos(), starStart)
		if ok {
			spaced := starStart-prevEnd > 0
			reportBefore(ctx, spaced, mode.before, prevEnd, starStart, starEnd)
		}
	}

	nextStart, ok := nextRealTokenStart(ctx.SourceFile, starEnd)
	if ok {
		spaced := nextStart-starEnd > 0
		reportAfter(ctx, spaced, mode.after, nextStart, starStart, starEnd)
	}
}

// methodKind classifies a MethodDeclaration node by its container:
//   - ObjectLiteralExpression parent → "shorthand"
//   - ClassDeclaration / ClassExpression parent → "method"
//
// Returns "" for any other container; the caller should skip those.
func methodKind(node *ast.Node) string {
	parent := node.Parent
	if parent == nil {
		return ""
	}
	switch parent.Kind {
	case ast.KindObjectLiteralExpression:
		return "shorthand"
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return "method"
	}
	return ""
}

// GeneratorStarSpacingRule enforces consistent spacing around the `*` in
// generator function declarations, function expressions, and method
// shorthand. Ported from @stylistic/eslint-plugin's generator-star-spacing.
var GeneratorStarSpacingRule = rule.Rule{
	Name: "@stylistic/generator-star-spacing",
	Run: func(ctx rule.RuleContext, raw any) rule.RuleListeners {
		modes := parseOptions(raw)

		checkFunc := func(node *ast.Node) {
			kind := "named"
			if functionName(node) == nil {
				kind = "anonymous"
			}
			checkGenerator(ctx, node, kind, modes)
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkFunc,
			ast.KindFunctionExpression:  checkFunc,
			ast.KindMethodDeclaration: func(node *ast.Node) {
				kind := methodKind(node)
				if kind == "" {
					return
				}
				checkGenerator(ctx, node, kind, modes)
			},
		}
	},
}
