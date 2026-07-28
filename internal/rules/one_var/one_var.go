package one_var

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	modeAlways      = "always"
	modeNever       = "never"
	modeConsecutive = "consecutive"
)

// typeOpts holds the per-statement-type configuration for initialized and
// uninitialized declarations. An empty string ("") means no mode is set for
// that group, in which case the rule does not check it.
type typeOpts struct {
	initialized   string
	uninitialized string
}

// oneVarOptions is the parsed rule options.
//
// `initialized` / `uninitialized` keys in the original ESLint schema are
// distributed onto every per-kind typeOpts at parse time, so the runtime only
// needs to look at the per-kind opts for each declaration.
type oneVarOptions struct {
	var_             typeOpts
	let_             typeOpts
	const_           typeOpts
	using_           typeOpts
	awaitUsing       typeOpts
	separateRequires bool
}

func (o *oneVarOptions) getType(key string) typeOpts {
	switch key {
	case "var":
		return o.var_
	case "let":
		return o.let_
	case "const":
		return o.const_
	case "using":
		return o.using_
	case "awaitUsing":
		return o.awaitUsing
	}
	return typeOpts{}
}

// funcScope tracks whether `var` declarations have been seen at the current
// function-level scope, broken down by initialization status (and require()
// status for separateRequires). Reused for block-level let/const/using/awaitUsing.
type funcScope struct {
	initialized   bool
	uninitialized bool
	required      bool
}

// blockScope holds per-kind state for the lexical (block) scopes.
type blockScope struct {
	let_       funcScope
	const_     funcScope
	using_     funcScope
	awaitUsing funcScope
}

func (b *blockScope) get(key string) *funcScope {
	switch key {
	case "let":
		return &b.let_
	case "const":
		return &b.const_
	case "using":
		return &b.using_
	case "awaitUsing":
		return &b.awaitUsing
	}
	return nil
}

// parseOptions normalizes the weakly-typed ESLint options shape — string,
// `[string]`, object, or `[object]` — into oneVarOptions. The `initialized` /
// `uninitialized` keys override the corresponding bucket of every per-kind
// typeOpts (matching upstream's distribution).
func parseOptions(opts any) oneVarOptions {
	const defaultMode = modeAlways

	var rawString string
	var rawObject map[string]interface{}

	switch v := opts.(type) {
	case string:
		rawString = v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				rawString = s
			} else if m, ok := v[0].(map[string]interface{}); ok {
				rawObject = m
			}
		}
	case map[string]interface{}:
		rawObject = v
	}

	result := oneVarOptions{}

	if rawObject == nil {
		mode := defaultMode
		if rawString != "" {
			mode = rawString
		}
		t := typeOpts{initialized: mode, uninitialized: mode}
		result.var_ = t
		result.let_ = t
		result.const_ = t
		result.using_ = t
		result.awaitUsing = t
		return result
	}

	if v, ok := rawObject["separateRequires"].(bool); ok {
		result.separateRequires = v
	}

	getStr := func(key string) string {
		if v, ok := rawObject[key].(string); ok {
			return v
		}
		return ""
	}

	for _, kv := range []struct {
		key  string
		dest *typeOpts
	}{
		{"var", &result.var_},
		{"let", &result.let_},
		{"const", &result.const_},
		{"using", &result.using_},
		{"awaitUsing", &result.awaitUsing},
	} {
		s := getStr(kv.key)
		kv.dest.initialized = s
		kv.dest.uninitialized = s
	}

	if _, ok := rawObject["uninitialized"]; ok {
		s := getStr("uninitialized")
		result.var_.uninitialized = s
		result.let_.uninitialized = s
		result.const_.uninitialized = s
		result.using_.uninitialized = s
		result.awaitUsing.uninitialized = s
	}
	if _, ok := rawObject["initialized"]; ok {
		s := getStr("initialized")
		result.var_.initialized = s
		result.let_.initialized = s
		result.const_.initialized = s
		result.using_.initialized = s
		result.awaitUsing.initialized = s
	}

	return result
}

// optsKey converts a kind keyword to its options-map key
// ("await using" → "awaitUsing", others unchanged).
func optsKey(kind string) string {
	if kind == "await using" {
		return "awaitUsing"
	}
	return kind
}

// isRequireDecl reports whether the declarator's initializer is a `require(...)`
// call. Mirrors ESLint's `isRequire` (which looks at `init.callee.name` directly,
// effectively ignoring parens because espree strips them). tsgo preserves
// parens as explicit nodes, so we apply SkipParentheses on both the initializer
// and the callee to match the effective ESLint behavior.
func isRequireDecl(decl *ast.Node) bool {
	vd := decl.AsVariableDeclaration()
	if vd == nil || vd.Initializer == nil {
		return false
	}
	init := ast.SkipParentheses(vd.Initializer)
	if init == nil || init.Kind != ast.KindCallExpression {
		return false
	}
	call := init.AsCallExpression()
	if call == nil || call.Expression == nil {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	return callee.AsIdentifier().Text == "require"
}

type declarationStats struct {
	initialized   int
	uninitialized int
	requires      int
	total         int
}

func collectDeclarationStats(decls []*ast.Node, inspectRequires bool) declarationStats {
	stats := declarationStats{total: len(decls)}
	for _, d := range decls {
		vd := d.AsVariableDeclaration()
		if vd == nil {
			continue
		}
		if vd.Initializer == nil {
			stats.uninitialized++
		} else {
			stats.initialized++
		}
		if inspectRequires && isRequireDecl(d) {
			stats.requires++
		}
	}
	return stats
}

func (s declarationStats) hasRequire() bool {
	return s.requires > 0
}

func (s declarationStats) allRequire() bool {
	return s.requires == s.total
}

// getStatementListSiblings returns the parent's statement list IF the parent
// is a body-array container. Mirrors ESLint's effective sibling lookup —
// `parent.body && Array.isArray(parent.body)` — which excludes SwitchCase
// (uses `consequent`, not `body`), IfStatement, LabeledStatement, etc.
//
// In tsgo, ClassStaticBlockDeclaration's body is a Block node, so var
// declarations inside a static block have parent === Block (covered by the
// KindBlock arm).
func getStatementListSiblings(parent *ast.Node) []*ast.Node {
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindBlock:
		b := parent.AsBlock()
		if b != nil && b.Statements != nil {
			return b.Statements.Nodes
		}
	case ast.KindSourceFile:
		sf := parent.AsSourceFile()
		if sf != nil && sf.Statements != nil {
			return sf.Statements.Nodes
		}
	case ast.KindModuleBlock:
		mb := parent.AsModuleBlock()
		if mb != nil && mb.Statements != nil {
			return mb.Statements.Nodes
		}
	}
	return nil
}

// findPreviousSibling returns the immediate previous sibling Statement of
// stmtNode in its parent's statement list, or nil if none.
func findPreviousSibling(stmtNode *ast.Node) *ast.Node {
	siblings := getStatementListSiblings(stmtNode.Parent)
	if siblings == nil {
		return nil
	}

	targetPos := stmtNode.Pos()
	index := sort.Search(len(siblings), func(i int) bool {
		return siblings[i].Pos() >= targetPos
	})
	for i := index; i < len(siblings) && siblings[i].Pos() == targetPos; i++ {
		if siblings[i] == stmtNode {
			if i == 0 {
				return nil
			}
			return siblings[i-1]
		}
	}

	// Parser-owned statement lists are source ordered, so the binary-search
	// path above is the normal case. Retain a linear fallback for synthetic or
	// recovery ASTs whose node positions do not preserve that invariant.
	for i, sibling := range siblings {
		if sibling == stmtNode {
			if i == 0 {
				return nil
			}
			return siblings[i-1]
		}
	}
	return nil
}

// isInStatementListPosition reports whether `parent` is a body-array container
// where splitting one declaration into multiple statements is syntactically
// valid. Excludes positions like `if (foo) var x, y;` where the result would
// either be a syntax error (for let/const/using) or change scoping semantics.
func isInStatementListPosition(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock,
		ast.KindCaseClause, ast.KindDefaultClause:
		return true
	}
	return false
}

// eslintVarDeclView is the single source of truth for translating tsgo's
// VariableStatement / VariableDeclarationList shape to ESLint's
// VariableDeclaration view. This is where every "tsgo has it as a modifier
// but ESLint has it as a wrapping node (or vice versa)" divergence lives —
// keep it concentrated here so the rest of the rule code can reason in terms
// of the ESLint model without sprinkling AST-shape patches everywhere.
//
// Fields:
//   - hasExportWrapper:  in ESLint, the VariableDeclaration is wrapped by
//     ExportNamedDeclaration. Affects: report column,
//     autofix join, sibling lookup for `consecutive`.
//     In tsgo, this is encoded as `export` on the
//     VariableStatement.Modifiers list.
//   - hasOtherModifier:  any non-export modifier (`declare` etc.). ESLint
//     keeps these inside VariableDeclaration; we keep them
//     in the report range and disable autofix to avoid
//     producing syntactically dubious output.
//   - reportStart:       first byte of the (would-be ESLint) VariableDeclaration.
//     For `export const x` this is the `const` keyword; for
//     `declare const x` this is the `declare` keyword; for
//     `export declare const x` this is the `declare` keyword.
//     For for-init forms, the VariableDeclarationList itself.
//   - reportEnd:         last byte of the VariableDeclaration. Includes the
//     trailing `;` for top-level forms; matches the
//     VariableDeclarationList's end for for-init.
type eslintVarDeclView struct {
	hasExportWrapper bool
	hasOtherModifier bool
	reportStart      int
	reportEnd        int
}

// makeView builds an eslintVarDeclView for either a top-level VariableStatement
// (parent of declList) or a for-init VariableDeclarationList (when there is no
// wrapping VariableStatement).
func makeView(declList *ast.Node, sf *ast.SourceFile) eslintVarDeclView {
	view := eslintVarDeclView{}
	parent := declList.Parent
	if parent != nil && parent.Kind == ast.KindVariableStatement {
		stmtNode := parent
		view.reportEnd = stmtNode.End()
		mods := stmtNode.Modifiers()
		if mods != nil {
			for _, m := range mods.Nodes {
				if m.Kind == ast.KindExportKeyword {
					view.hasExportWrapper = true
				} else {
					view.hasOtherModifier = true
				}
			}
		}
		// reportStart logic mirrors ESLint's VariableDeclaration node bounds:
		//   * with `export`: skip past it (ExportNamedDeclaration absorbs it
		//     in ESLint), so start at the next non-export modifier (e.g. `declare`)
		//     or at the kind keyword.
		//   * without `export`: start at the VariableStatement's first token
		//     (which is either a `declare`-style modifier or the kind keyword).
		if view.hasExportWrapper {
			view.reportStart = utils.TrimNodeTextRange(sf, declList).Pos()
			for _, m := range mods.Nodes {
				if m.Kind != ast.KindExportKeyword {
					view.reportStart = utils.TrimNodeTextRange(sf, m).Pos()
					break
				}
			}
		} else {
			view.reportStart = utils.TrimNodeTextRange(sf, stmtNode).Pos()
		}
	} else {
		// for-init: no wrapping VariableStatement. The VariableDeclarationList
		// itself is the closest analog to ESLint's VariableDeclaration.
		view.reportStart = utils.TrimNodeTextRange(sf, declList).Pos()
		view.reportEnd = declList.End()
	}
	return view
}

type eslintVarDeclViewCache struct {
	declList *ast.Node
	sf       *ast.SourceFile
	view     eslintVarDeclView
	ready    bool
}

func (c *eslintVarDeclViewCache) get() eslintVarDeclView {
	if !c.ready {
		c.view = makeView(c.declList, c.sf)
		c.ready = true
	}
	return c.view
}

// asKeywordHost picks the right node to feed into utils.GetVarKeywordRange:
// VariableStatement (so its modifier list is skipped) when present, else the
// VariableDeclarationList itself.
func asKeywordHost(declList *ast.Node) *ast.Node {
	if declList.Parent != nil && declList.Parent.Kind == ast.KindVariableStatement {
		return declList.Parent
	}
	return declList
}

// reportRange returns the (start, end) range to attach the diagnostic to.
func (v eslintVarDeclView) reportRange() core.TextRange {
	return core.NewTextRange(v.reportStart, v.reportEnd)
}

// canFix reports whether autofix can run. ESLint's joinDeclarations / split
// fix paths require the VariableDeclaration to NOT be wrapped in
// ExportNamedDeclaration (because their body-array lookup stops there) and
// to NOT carry a `declare`-style modifier (which would produce invalid syntax
// when merged or split).
func (v eslintVarDeclView) canFix() bool {
	return !v.hasExportWrapper && !v.hasOtherModifier
}

// fixCombine generates fixes to merge the current declaration list into the
// previous sibling VariableStatement of the same kind. Returns nil when the
// view is not fixable, or there is no matching previous sibling.
func fixCombine(view eslintVarDeclView, currentDeclList *ast.Node, kind string, sf *ast.SourceFile) []rule.RuleFix {
	if !view.canFix() {
		return nil
	}
	parent := currentDeclList.Parent
	if parent == nil || parent.Kind != ast.KindVariableStatement {
		return nil
	}
	stmtNode := parent

	prev := findPreviousSibling(stmtNode)
	if prev == nil || prev.Kind != ast.KindVariableStatement {
		return nil
	}
	prevDeclList := prev.AsVariableStatement().DeclarationList
	if prevDeclList == nil || utils.GetVarDeclListKind(prevDeclList) != kind {
		return nil
	}
	prevView := makeView(prevDeclList, sf)
	if !prevView.canFix() {
		return nil
	}

	// Keyword scanning is needed only when autofixes are requested. All
	// diagnostics use deferred fix builders, so keep this work off the normal
	// diagnostics-only path.
	keywordRange := utils.GetVarKeywordRange(asKeywordHost(currentDeclList), sf)

	var fixes []rule.RuleFix

	prevText := utils.TrimmedNodeText(sf, prev)
	if strings.HasSuffix(prevText, ";") {
		semiPos := prev.End() - 1
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(semiPos, prev.End()), ","))
	} else {
		fixes = append(fixes, rule.RuleFixInsertAfter(prev, ","))
	}

	if kind == "await using" {
		// `keywordRange` covers `await`; the next non-trivia token is `using`.
		usingRange := scanner.GetRangeOfTokenAtPosition(sf, keywordRange.End())
		fixes = append(fixes, rule.RuleFixRemoveRange(usingRange))
	}
	fixes = append(fixes, rule.RuleFixRemoveRange(keywordRange))

	return fixes
}

// fixSplit generates fixes to split a multi-declarator declaration into
// separate statements. Returns nil when:
//   - the node is in a non-statement-list position (e.g. `if (foo) var x, y;`);
//   - the wrapping VariableStatement carries any non-export modifier (e.g.
//     `declare var x, y;`).
//
// `export` is preserved by prepending "export " before each split keyword.
// `declare` etc. suppress the fix — upstream ESLint would drop them and
// produce non-compilable TS, so rslint deliberately keeps the diagnostic but
// withholds the fix to avoid breaking user code. (Documented divergence.)
func fixSplit(view eslintVarDeclView, declList *ast.Node, kind string, sf *ast.SourceFile) []rule.RuleFix {
	parent := declList.Parent
	if parent == nil || parent.Kind != ast.KindVariableStatement {
		return nil
	}
	stmtNode := parent
	if view.hasOtherModifier {
		return nil
	}
	exportPrefix := ""
	if view.hasExportWrapper {
		exportPrefix = "export "
	}

	if !isInStatementListPosition(stmtNode.Parent) {
		return nil
	}

	declarations := declList.AsVariableDeclarationList().Declarations
	if declarations == nil {
		return nil
	}
	decls := declarations.Nodes

	var fixes []rule.RuleFix

	for i := range len(decls) - 1 {
		decl := decls[i]
		s := scanner.GetScannerForSourceFile(sf, decl.End())
		if s.Token() != ast.KindCommaToken {
			continue
		}
		commaStart := s.TokenStart()
		commaEnd := s.TokenEnd()

		nextDeclStart := utils.TrimNodeTextRange(sf, decls[i+1]).Pos()
		between := sf.Text()[commaEnd:nextDeclStart]

		switch {
		case len(between) == 0:
			// `var x,y` — no whitespace; insert the kind keyword with a
			// trailing space so it doesn't fuse with the next identifier.
			fixes = append(fixes, rule.RuleFixReplaceRange(
				core.NewTextRange(commaStart, commaEnd),
				"; "+exportPrefix+kind+" ",
			))
		case utils.ContainsLineTerminator(between, 0, len(between)) ||
			strings.Contains(between, "//") ||
			strings.Contains(between, "/*"):
			// Line break or comment between `,` and the next declarator —
			// preserve the original whitespace/comments and place the new
			// kind keyword right before the next declarator.
			replacement := ";" + between + exportPrefix + kind + " "
			fixes = append(fixes, rule.RuleFixReplaceRange(
				core.NewTextRange(commaStart, nextDeclStart),
				replacement,
			))
		default:
			// Single-or-multi space — keep the existing whitespace; only
			// replace `,` with `; <kind>`.
			fixes = append(fixes, rule.RuleFixReplaceRange(
				core.NewTextRange(commaStart, commaEnd),
				"; "+exportPrefix+kind,
			))
		}
	}

	return fixes
}

func msg(id, kind string) rule.RuleMessage {
	if id == "splitRequires" {
		return rule.RuleMessage{
			Id:          id,
			Description: "Split requires to be separated into a single block.",
		}
	}

	var description string
	switch id {
	case "combineUninitialized":
		description = "Combine this with the previous '" + kind + "' statement with uninitialized variables."
	case "combineInitialized":
		description = "Combine this with the previous '" + kind + "' statement with initialized variables."
	case "splitUninitialized":
		description = "Split uninitialized '" + kind + "' declarations into multiple statements."
	case "splitInitialized":
		description = "Split initialized '" + kind + "' declarations into multiple statements."
	case "combine":
		description = "Combine this with the previous '" + kind + "' statement."
	case "split":
		description = "Split '" + kind + "' declarations into multiple statements."
	}
	return rule.RuleMessage{
		Id:          id,
		Description: description,
		Data:        map[string]string{"type": kind},
	}
}

// recordTypes flips scope flags for declarators that match the option's
// MODE_ALWAYS, including the `required` flag for separateRequires-tracked
// requires. Mirrors ESLint's `recordTypes`.
func recordTypes(scope *funcScope, opts typeOpts, stats declarationStats, separateRequires bool) {
	if opts.uninitialized == modeAlways && stats.uninitialized > 0 {
		scope.uninitialized = true
	}
	if opts.initialized == modeAlways && stats.initialized > 0 {
		if separateRequires {
			if stats.requires > 0 {
				scope.required = true
			}
			if stats.initialized > stats.requires {
				scope.initialized = true
			}
		} else {
			scope.initialized = true
		}
	}
}

// hasOnlyOneStatement returns true if the current declaration would be the
// first of its kind in scope (under "always" semantics) AND records its types
// in the scope. Returns false (without recording) on conflict — matches
// ESLint's `hasOnlyOneStatement` precisely, including the `hasRequires` /
// `currentScope.required` carve-outs for separateRequires.
func hasOnlyOneStatement(scope *funcScope, opts typeOpts, stats declarationStats, separateRequires bool) bool {
	hasReq := stats.hasRequire()

	if opts.uninitialized == modeAlways && opts.initialized == modeAlways {
		if scope.uninitialized || scope.initialized {
			if !hasReq {
				return false
			}
		}
	}
	if stats.uninitialized > 0 {
		if opts.uninitialized == modeAlways && scope.uninitialized {
			return false
		}
	}
	if stats.initialized > 0 {
		if opts.initialized == modeAlways && scope.initialized {
			if !hasReq {
				return false
			}
		}
	}
	if scope.required && hasReq {
		return false
	}
	recordTypes(scope, opts, stats, separateRequires)
	return true
}

func hasMode(opts typeOpts, mode string) bool {
	return opts.initialized == mode || opts.uninitialized == mode
}

func currentScope(key string, funcStack []funcScope, blockStack []blockScope) *funcScope {
	if key == "var" {
		if len(funcStack) == 0 {
			return nil
		}
		return &funcStack[len(funcStack)-1]
	}
	if len(blockStack) == 0 {
		return nil
	}
	return blockStack[len(blockStack)-1].get(key)
}

var OneVarRule = rule.Rule{
	Name: "one-var",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)
		needsVarScope := hasMode(opts.var_, modeAlways)
		needsBlockScope := hasMode(opts.let_, modeAlways) ||
			hasMode(opts.const_, modeAlways) ||
			hasMode(opts.using_, modeAlways) ||
			hasMode(opts.awaitUsing, modeAlways)

		var funcStack []funcScope
		if needsVarScope {
			funcStack = []funcScope{{}}
		}
		var blockStack []blockScope
		if needsBlockScope {
			blockStack = []blockScope{{}}
		}

		checkDeclList := func(node *ast.Node) {
			declList := node.AsVariableDeclarationList()
			if declList == nil || declList.Declarations == nil {
				return
			}
			parent := node.Parent
			if parent == nil {
				return
			}

			var stmtNode *ast.Node
			isForInit := false
			isForInOfInit := false
			switch parent.Kind {
			case ast.KindVariableStatement:
				stmtNode = parent
			case ast.KindForStatement:
				stmtNode = parent
				isForInit = true
			case ast.KindForInStatement, ast.KindForOfStatement:
				stmtNode = parent
				isForInOfInit = true
			default:
				return
			}

			kind := utils.GetVarDeclListKind(node)
			if kind == "" {
				return
			}
			key := optsKey(kind)
			typeOpts := opts.getType(key)
			if typeOpts.initialized == "" && typeOpts.uninitialized == "" {
				return
			}

			usesAlways := hasMode(typeOpts, modeAlways)
			usesConsecutive := hasMode(typeOpts, modeConsecutive)
			usesNever := hasMode(typeOpts, modeNever)
			declarations := declList.Declarations.Nodes
			stats := collectDeclarationStats(declarations, usesAlways || usesConsecutive)
			viewCache := eslintVarDeclViewCache{
				declList: node,
				sf:       ctx.SourceFile,
			}

			// 1. splitRequires: separateRequires + mixed requires/non-requires.
			if typeOpts.initialized == modeAlways && opts.separateRequires {
				if stats.hasRequire() && !stats.allRequire() {
					view := viewCache.get()
					ctx.ReportRange(view.reportRange(), msg("splitRequires", kind))
				}
			}

			// 2. consecutive: previous sibling is a same-kind VariableStatement.
			//
			// ESLint walks `node.parent.body` to find the previous sibling.
			// `export`-wrapped declarations are inside ExportNamedDeclaration
			// (no `.body` array), so consecutive never fires on either side
			// of an export wrapper. Honor both guards via the view abstraction.
			if usesConsecutive && !isForInit && !isForInOfInit && stmtNode.Kind == ast.KindVariableStatement {
				view := viewCache.get()
				if !view.hasExportWrapper {
					if prev := findPreviousSibling(stmtNode); prev != nil && prev.Kind == ast.KindVariableStatement {
						prevDeclList := prev.AsVariableStatement().DeclarationList
						if prevDeclList != nil && utils.GetVarDeclListKind(prevDeclList) == kind {
							prevView := makeView(prevDeclList, ctx.SourceFile)
							if !prevView.hasExportWrapper {
								prevDecls := prevDeclList.AsVariableDeclarationList().Declarations.Nodes
								prevStats := collectDeclarationStats(prevDecls, true)
								combinedAllRequire := stats.allRequire() && prevStats.allRequire()
								combinedAnyRequire := stats.hasRequire() || prevStats.hasRequire()
								mixedRequires := combinedAnyRequire && !combinedAllRequire

								if !mixedRequires {
									reportRange := view.reportRange()
									switch {
									case typeOpts.initialized == modeConsecutive && typeOpts.uninitialized == modeConsecutive:
										ctx.ReportRangeWithDeferredFixes(reportRange, msg("combine", kind), func() []rule.RuleFix {
											return fixCombine(view, node, kind, ctx.SourceFile)
										})
									case typeOpts.initialized == modeConsecutive && stats.initialized > 0 && prevStats.initialized > 0:
										ctx.ReportRangeWithDeferredFixes(reportRange, msg("combineInitialized", kind), func() []rule.RuleFix {
											return fixCombine(view, node, kind, ctx.SourceFile)
										})
									case typeOpts.uninitialized == modeConsecutive && stats.uninitialized > 0 && prevStats.uninitialized > 0:
										ctx.ReportRangeWithDeferredFixes(reportRange, msg("combineUninitialized", kind), func() []rule.RuleFix {
											return fixCombine(view, node, kind, ctx.SourceFile)
										})
									}
								}
							}
						}
					}
				}
			}

			// 3. always: scope already saw a same-kind statement.
			if usesAlways {
				scope := currentScope(key, funcStack, blockStack)
				if scope != nil && !hasOnlyOneStatement(scope, typeOpts, stats, opts.separateRequires) {
					view := viewCache.get()
					reportRange := view.reportRange()
					if typeOpts.initialized == modeAlways && typeOpts.uninitialized == modeAlways {
						ctx.ReportRangeWithDeferredFixes(reportRange, msg("combine", kind), func() []rule.RuleFix {
							return fixCombine(view, node, kind, ctx.SourceFile)
						})
					} else {
						if typeOpts.initialized == modeAlways && stats.initialized > 0 {
							ctx.ReportRangeWithDeferredFixes(reportRange, msg("combineInitialized", kind), func() []rule.RuleFix {
								return fixCombine(view, node, kind, ctx.SourceFile)
							})
						}
						if typeOpts.uninitialized == modeAlways && stats.uninitialized > 0 {
							// for-in/for-of left side: skip combineUninitialized
							// (mirrors ESLint's `parent.left === node` early return).
							if !isForInOfInit {
								ctx.ReportRangeWithDeferredFixes(reportRange, msg("combineUninitialized", kind), func() []rule.RuleFix {
									return fixCombine(view, node, kind, ctx.SourceFile)
								})
							}
						}
					}
				}
			}

			// 4. never: multiple declarators in a single statement (skip ForStatement init).
			if usesNever && !isForInit {
				totalDecls := stats.initialized + stats.uninitialized
				if totalDecls > 1 {
					view := viewCache.get()
					reportRange := view.reportRange()
					if typeOpts.initialized == modeNever && typeOpts.uninitialized == modeNever {
						ctx.ReportRangeWithDeferredFixes(reportRange, msg("split", kind), func() []rule.RuleFix {
							return fixSplit(view, node, kind, ctx.SourceFile)
						})
					} else if typeOpts.initialized == modeNever && stats.initialized > 0 {
						ctx.ReportRangeWithDeferredFixes(reportRange, msg("splitInitialized", kind), func() []rule.RuleFix {
							return fixSplit(view, node, kind, ctx.SourceFile)
						})
					} else if typeOpts.uninitialized == modeNever && stats.uninitialized > 0 {
						ctx.ReportRangeWithDeferredFixes(reportRange, msg("splitUninitialized", kind), func() []rule.RuleFix {
							return fixSplit(view, node, kind, ctx.SourceFile)
						})
					}
				}
			}
		}

		// "never" and "consecutive" are statement-local and do not need any
		// function or block boundary bookkeeping.
		if !needsVarScope && !needsBlockScope {
			return rule.RuleListeners{
				ast.KindVariableDeclarationList: checkDeclList,
			}
		}

		startBlock := func(*ast.Node) {
			if needsBlockScope {
				blockStack = append(blockStack, blockScope{})
			}
		}
		endBlock := func(*ast.Node) {
			if needsBlockScope && len(blockStack) > 0 {
				blockStack = blockStack[:len(blockStack)-1]
			}
		}
		startFunction := func(*ast.Node) {
			if needsVarScope {
				funcStack = append(funcStack, funcScope{})
			}
			startBlock(nil)
		}
		endFunction := func(*ast.Node) {
			if needsVarScope && len(funcStack) > 0 {
				funcStack = funcStack[:len(funcStack)-1]
			}
			endBlock(nil)
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                              startFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration):         endFunction,
			ast.KindFunctionExpression:                               startFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):          endFunction,
			ast.KindArrowFunction:                                    startFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):               endFunction,
			ast.KindMethodDeclaration:                                startFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):           endFunction,
			ast.KindGetAccessor:                                      startFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):                 endFunction,
			ast.KindSetAccessor:                                      startFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):                 endFunction,
			ast.KindConstructor:                                      startFunction,
			rule.ListenerOnExit(ast.KindConstructor):                 endFunction,
			ast.KindClassStaticBlockDeclaration:                      startFunction,
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): endFunction,

			ast.KindBlock:                                startBlock,
			rule.ListenerOnExit(ast.KindBlock):           endBlock,
			ast.KindForStatement:                         startBlock,
			rule.ListenerOnExit(ast.KindForStatement):    endBlock,
			ast.KindForInStatement:                       startBlock,
			rule.ListenerOnExit(ast.KindForInStatement):  endBlock,
			ast.KindForOfStatement:                       startBlock,
			rule.ListenerOnExit(ast.KindForOfStatement):  endBlock,
			ast.KindSwitchStatement:                      startBlock,
			rule.ListenerOnExit(ast.KindSwitchStatement): endBlock,
			// NOTE: TypeScript `namespace N { ... }` and `declare module 'x' { ... }`
			// have a TSModuleBlock body in the AST, but ESLint's one-var rule
			// does NOT register any listener for it — declarations inside live
			// in the enclosing function scope. Mirror upstream by NOT pushing
			// a new scope on ModuleBlock entry. (Verified against ESLint v10.2.1
			// on `namespace A { var a } namespace B { var a }` reports combine.)

			ast.KindVariableDeclarationList: checkDeclList,
		}
	},
}
