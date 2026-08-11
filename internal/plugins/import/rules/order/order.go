// Package order ports eslint-plugin-import's `order` rule to rslint.
//
// See: https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/order.md
//
// Observable differences from upstream are documented in `order.md`. The
// option schema, messages, report locations, and fixes otherwise follow the
// upstream rule's public behavior.
package order

import (
	_ "embed"
	"fmt"
	"math"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/stringutil"
	importutil "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// allTypes lists the import-type buckets the rule recognizes, in upstream's
// canonical order. `convertGroupsToRanks` uses this for "omitted types"
// computation.
var allTypes = []string{
	"builtin", "external", "internal", "unknown",
	"parent", "sibling", "index", "object", "type",
}

// defaultGroups matches upstream's default `groups` option.
var defaultGroups = []any{"builtin", "external", "parent", "sibling", "index"}

// defaultDistinctGroup mirrors upstream's `defaultDistinctGroup = true`.
const defaultDistinctGroup = true

// ---------------------------------------------------------------------------
// Options & ranks
// ---------------------------------------------------------------------------

type alphabetizeOptions struct {
	order           string // "ignore" | "asc" | "desc"
	orderImportKind string // "ignore" | "asc" | "desc"
	caseInsensitive bool
}

type namedOptions struct {
	imports    bool
	exports    bool
	require    bool
	cjsExports bool
	types      string // "mixed" | "types-first" | "types-last"
}

type pathGroup struct {
	pattern           string
	patternOptions    importutil.MinimatchOptions
	patternOptionsSet bool
	matcher           *importutil.Minimatcher
	group             string
	positionRaw       string // "before" | "after" | ""
	position          float64
}

type ranks struct {
	groups       map[string]float64
	omittedTypes []string
	pathGroups   []pathGroup
	maxPosition  int
}

type options struct {
	groups                        []any
	pathGroups                    []pathGroup
	pathGroupsExcludedImportTypes map[string]bool
	distinctGroup                 bool
	newlinesBetween               string
	newlinesBetweenTypes          string
	alphabetize                   alphabetizeOptions
	named                         namedOptions
	sortTypesGroup                bool
	warnOnUnassigned              bool
	consolidateIslands            string

	// derived
	ranks               ranks
	rankErr             error
	isSortingTypesGroup bool
	consolidating       bool
	internalRegex       *regexp.Regexp
}

// parseOptions extracts the rule options from the weakly-typed input,
// inheriting `import/internal-regex` from settings. It never panics — a
// malformed `groups` array sets `rankErr`, which the caller surfaces as a
// single Program-level diagnostic, matching upstream's behaviour.
func parseOptions(raw any, settings map[string]any) options {
	opts := options{
		groups:                        defaultGroups,
		pathGroupsExcludedImportTypes: map[string]bool{"builtin": true, "external": true, "object": true},
		distinctGroup:                 defaultDistinctGroup,
		newlinesBetween:               "ignore",
		alphabetize:                   alphabetizeOptions{order: "ignore", orderImportKind: "ignore"},
		named:                         namedOptions{types: "mixed"},
		consolidateIslands:            "never",
	}

	m, _ := raw.(map[string]any)
	if m != nil {
		if g, ok := m["groups"].([]any); ok && len(g) > 0 {
			opts.groups = g
		}
		if pgs, ok := m["pathGroups"].([]any); ok {
			opts.pathGroups = parsePathGroups(pgs)
		}
		if pge, ok := m["pathGroupsExcludedImportTypes"].([]any); ok {
			opts.pathGroupsExcludedImportTypes = map[string]bool{}
			for _, v := range pge {
				if s, ok := v.(string); ok {
					opts.pathGroupsExcludedImportTypes[s] = true
				}
			}
		}
		if v, ok := m["distinctGroup"].(bool); ok {
			opts.distinctGroup = v
		}
		if v, ok := m["newlines-between"].(string); ok {
			opts.newlinesBetween = v
		}
		if v, ok := m["newlines-between-types"].(string); ok {
			opts.newlinesBetweenTypes = v
		}
		if v, ok := m["sortTypesGroup"].(bool); ok {
			opts.sortTypesGroup = v
		}
		if v, ok := m["warnOnUnassignedImports"].(bool); ok {
			opts.warnOnUnassigned = v
		}
		if v, ok := m["consolidateIslands"].(string); ok {
			opts.consolidateIslands = v
		}
		if a, ok := m["alphabetize"].(map[string]any); ok {
			if v, ok := a["order"].(string); ok {
				opts.alphabetize.order = v
			}
			if v, ok := a["orderImportKind"].(string); ok {
				opts.alphabetize.orderImportKind = v
			}
			if v, ok := a["caseInsensitive"].(bool); ok {
				opts.alphabetize.caseInsensitive = v
			}
		}
		if n, ok := m["named"]; ok {
			parseNamed(n, &opts.named)
		}
	}

	// `newlines-between-types` defaults to the value of `newlines-between`,
	// matching upstream `options['newlines-between-types'] || newlinesBetweenImports`.
	if opts.newlinesBetweenTypes == "" {
		opts.newlinesBetweenTypes = opts.newlinesBetween
	}

	if settings != nil {
		if s, ok := settings["import/internal-regex"].(string); ok && s != "" {
			if re, err := regexp.Compile(s); err == nil {
				opts.internalRegex = re
			}
		}
	}

	r, err := buildRanks(opts.groups, opts.pathGroups)
	if err != nil {
		opts.rankErr = err
	}
	opts.ranks = r

	isTypeGroupInGroups := !slices.Contains(opts.ranks.omittedTypes, "type")
	opts.isSortingTypesGroup = isTypeGroupInGroups && opts.sortTypesGroup

	opts.consolidating = opts.consolidateIslands == "inside-groups" &&
		(opts.newlinesBetween == "always-and-inside-groups" ||
			opts.newlinesBetweenTypes == "always-and-inside-groups")

	return opts
}

// parseNamed handles both the boolean shorthand (`named: true|false`) and the
// full object form. Upstream treats missing per-call toggles (`import`,
// `export`, etc.) as inheriting `enabled`.
func parseNamed(raw any, n *namedOptions) {
	switch v := raw.(type) {
	case bool:
		n.imports = v
		n.exports = v
		n.require = v
		n.cjsExports = v
	case map[string]any:
		enabled, _ := v["enabled"].(bool)
		assign := func(key string, target *bool) {
			if x, ok := v[key].(bool); ok {
				*target = x
			} else {
				*target = enabled
			}
		}
		assign("import", &n.imports)
		assign("export", &n.exports)
		assign("require", &n.require)
		assign("cjsExports", &n.cjsExports)
		if t, ok := v["types"].(string); ok {
			n.types = t
		}
	}
}

func parsePathGroups(raw []any) []pathGroup {
	out := make([]pathGroup, 0, len(raw))
	for _, e := range raw {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		pg := pathGroup{}
		if s, ok := m["pattern"].(string); ok {
			pg.pattern = s
		}
		if s, ok := m["group"].(string); ok {
			pg.group = s
		}
		if s, ok := m["position"].(string); ok {
			pg.positionRaw = s
		}
		if rawOptions, ok := m["patternOptions"].(map[string]any); ok {
			pg.patternOptionsSet = true
			pg.patternOptions.NoNegate, _ = rawOptions["nonegate"].(bool)
			pg.patternOptions.NoComment, _ = rawOptions["nocomment"].(bool)
			pg.patternOptions.NoCase, _ = rawOptions["nocase"].(bool)
			pg.patternOptions.MatchBase, _ = rawOptions["matchBase"].(bool)
			pg.patternOptions.NoGlobStar, _ = rawOptions["noglobstar"].(bool)
			pg.patternOptions.NoExt, _ = rawOptions["noext"].(bool)
			pg.patternOptions.NoBrace, _ = rawOptions["nobrace"].(bool)
			pg.patternOptions.Dot, _ = rawOptions["dot"].(bool)
			pg.patternOptions.Partial, _ = rawOptions["partial"].(bool)
			pg.patternOptions.FlipNegate, _ = rawOptions["flipNegate"].(bool)
			pg.patternOptions.AllowWindowsEscape, _ = rawOptions["allowWindowsEscape"].(bool)
		}
		matcherOptions := pg.patternOptions
		if !pg.patternOptionsSet {
			matcherOptions.NoComment = true
		}
		pg.matcher = importutil.CompileMinimatch(pg.pattern, matcherOptions)
		out = append(out, pg)
	}
	return out
}

// buildRanks converts the user's `groups` and `pathGroups` configuration into
// numeric ranks. Mirrors `convertGroupsToRanks` + `convertPathGroupsForRanks`
// in upstream.
func buildRanks(groupsCfg []any, pathGroupsCfg []pathGroup) (ranks, error) {
	rankObject := make(map[string]float64)
	for i, g := range groupsCfg {
		base := float64(i) * 2
		switch v := g.(type) {
		case string:
			if !slices.Contains(allTypes, v) {
				return ranks{}, fmt.Errorf("incorrect configuration of the rule: unknown type `%q`", v)
			}
			if _, dup := rankObject[v]; dup {
				return ranks{}, fmt.Errorf("incorrect configuration of the rule: `%q` is duplicated", v)
			}
			rankObject[v] = base
		case []any:
			for _, item := range v {
				s, ok := item.(string)
				if !ok || !slices.Contains(allTypes, s) {
					return ranks{}, fmt.Errorf("incorrect configuration of the rule: unknown type `%v`", item)
				}
				if _, dup := rankObject[s]; dup {
					return ranks{}, fmt.Errorf("incorrect configuration of the rule: `%q` is duplicated", s)
				}
				rankObject[s] = base
			}
		default:
			return ranks{}, fmt.Errorf("incorrect configuration of the rule: invalid group entry %v", g)
		}
	}

	var omitted []string
	for _, t := range allTypes {
		if _, ok := rankObject[t]; !ok {
			omitted = append(omitted, t)
		}
	}
	for _, t := range omitted {
		rankObject[t] = float64(len(groupsCfg)) * 2
	}

	pgs, maxPos := convertPathGroupsForRanks(pathGroupsCfg)
	return ranks{
		groups:       rankObject,
		omittedTypes: omitted,
		pathGroups:   pgs,
		maxPosition:  maxPos,
	}, nil
}

// convertPathGroupsForRanks resolves the abstract `position: "after" | "before"`
// flags into numeric offsets, and returns the maxPosition denominator used by
// computePathRank. Mirrors upstream literally.
func convertPathGroupsForRanks(in []pathGroup) ([]pathGroup, int) {
	after := map[string]int{}
	before := map[string][]int{}

	out := make([]pathGroup, len(in))
	copy(out, in)

	for i := range out {
		switch out[i].positionRaw {
		case "after":
			if after[out[i].group] == 0 {
				after[out[i].group] = 1
			}
			out[i].position = float64(after[out[i].group])
			after[out[i].group]++
		case "before":
			before[out[i].group] = append(before[out[i].group], i)
		default:
			out[i].position = 0
		}
	}

	maxPosition := 1
	for _, indexes := range before {
		groupLen := len(indexes)
		for j, idx := range indexes {
			out[idx].position = float64(-(groupLen - j))
		}
		if groupLen > maxPosition {
			maxPosition = groupLen
		}
	}
	for _, n := range after {
		if n-1 > maxPosition {
			maxPosition = n - 1
		}
	}

	if maxPosition > 10 {
		maxPosition = int(math.Pow(10, math.Ceil(math.Log10(float64(maxPosition)))))
	} else {
		maxPosition = 10
	}
	return out, maxPosition
}

// ---------------------------------------------------------------------------
// Per-import bookkeeping
// ---------------------------------------------------------------------------

type importEntry struct {
	// node is the syntax node upstream reports and uses for source locations:
	// an import/export declaration, require call, or CommonJS assignment.
	node *ast.Node
	// root is the enclosing statement moved by a fixer. It differs from node
	// for require calls and CommonJS assignments.
	root *ast.Node

	value        string
	displayName  string
	alias        string
	typ          string // "import", "require", "import:object", "export"
	importKind   string // "type", "typeof", ""
	classifyType string

	rank        float64
	isMultiline bool
}

// namedEntry records a single name in a named-imports / named-exports /
// destructured-require list. Used for intra-list sort step.
type namedEntry struct {
	node        *ast.Node
	value       string
	displayName string
	alias       string
	typ         string // "import", "export"
	kind        string // "type", "value", ""

	rank float64
}

// sourceInfo keeps source-derived data lazy. Most files need the text for
// import names, but line maps and comments are only needed by newline checks
// or when the diagnostic consumer actually requests an autofix.
type sourceInfo struct {
	file           *ast.SourceFile
	commentStore   *rule.CommentStore
	text           string
	lineStarts     []core.TextPos
	comments       []*ast.CommentRange
	commentsLoaded bool
}

func (source *sourceInfo) lines() []core.TextPos {
	if source.lineStarts == nil && source.file != nil {
		source.lineStarts = source.file.ECMALineMap()
	}
	return source.lineStarts
}

func (source *sourceInfo) allComments() []*ast.CommentRange {
	if !source.commentsLoaded {
		source.commentsLoaded = true
		if source.commentStore != nil {
			source.comments = source.commentStore.All()
		}
	}
	return source.comments
}

// ---------------------------------------------------------------------------
// Module classification (delegates to importutil.ClassifyImport)
// ---------------------------------------------------------------------------

func computeRank(entry *importEntry, opts options) float64 {
	r := opts.ranks
	isTypeGroupInGroups := !slices.Contains(r.omittedTypes, "type")
	isTypeOnly := entry.importKind == "type"

	var impType string
	if entry.typ == "import:object" {
		impType = "object"
	} else if isTypeOnly && isTypeGroupInGroups && !opts.isSortingTypesGroup {
		impType = "type"
	} else {
		impType = entry.classifyType
		// "absolute" is intentionally NOT in `allTypes` — upstream returns it
		// from typeTest but never registers a rank for it, so the entry falls
		// through to `rank == -1` and gets ignored. We replicate that here.
	}

	excluded := opts.pathGroupsExcludedImportTypes[impType]
	excludedFromPathRank := isTypeOnly && isTypeGroupInGroups && opts.pathGroupsExcludedImportTypes["type"]

	rank := math.NaN()
	if !excluded && !excludedFromPathRank {
		rank = computePathRank(r, entry.value)
	}

	if math.IsNaN(rank) {
		groupRank, ok := r.groups[impType]
		if !ok {
			return -1
		}
		rank = groupRank
	}

	if isTypeOnly && opts.isSortingTypesGroup {
		rank = r.groups["type"] + rank/10
	}

	if entry.typ != "import" && !strings.HasPrefix(entry.typ, "import:") {
		rank += 100
	}
	return rank
}

func computePathRank(r ranks, path string) float64 {
	for _, pg := range r.pathGroups {
		if matchPathGroup(path, pg) {
			return r.groups[pg.group] + pg.position/float64(r.maxPosition)
		}
	}
	return math.NaN()
}

func matchPathGroup(path string, pg pathGroup) bool {
	if pg.matcher != nil {
		return pg.matcher.Match(path)
	}
	return false
}

// ---------------------------------------------------------------------------
// Out-of-order detection
// ---------------------------------------------------------------------------

func findOutOfOrder(imports []*importEntry) []*importEntry {
	if len(imports) == 0 {
		return nil
	}
	maxSeen := imports[0]
	var out []*importEntry
	for _, imp := range imports {
		if imp.rank < maxSeen.rank {
			out = append(out, imp)
		}
		if maxSeen.rank < imp.rank {
			maxSeen = imp
		}
	}
	return out
}

// reverseRanks duplicates the slice and negates every rank, so a forward
// findOutOfOrder over the result is equivalent to scanning the original from
// the end. Mirrors upstream's `reverse(array)`.
func reverseRanks(in []*importEntry) []*importEntry {
	out := make([]*importEntry, len(in))
	for i, v := range in {
		dup := *v
		dup.rank = -v.rank
		out[len(in)-1-i] = &dup
	}
	return out
}

func makeOutOfOrderReports(ctx rule.RuleContext, imported []*importEntry, source *sourceInfo) {
	out := findOutOfOrder(imported)
	if len(out) == 0 {
		return
	}
	rev := reverseRanks(imported)
	revOut := findOutOfOrder(rev)
	if len(revOut) < len(out) {
		// Use the reversed list (with negated ranks) as the search space — the
		// `found = imp.rank > X` predicate works against negated ranks too.
		// The entries in `rev` carry the same node identity as the originals,
		// so report positions still come from the original AST.
		reportOutOfOrder(ctx, rev, revOut, "after", source)
		return
	}
	reportOutOfOrder(ctx, imported, out, "before", source)
}

func reportOutOfOrder(ctx rule.RuleContext, imported []*importEntry, outOfOrder []*importEntry, order string, source *sourceInfo) {
	ordered := append([]*importEntry(nil), outOfOrder...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ast.CompareNodePositions(ordered[i].node, ordered[j].node) < 0
	})
	for _, imp := range ordered {
		var found *importEntry
		for _, ii := range imported {
			if ii.rank > imp.rank {
				found = ii
				break
			}
		}
		if found == nil {
			continue
		}
		makeOutOfOrderReport(ctx, found, imp, order, source)
	}
}

func makeOutOfOrderReport(ctx rule.RuleContext, first, second *importEntry, order string, source *sourceInfo) {
	firstDisplay := first.displayName
	secondDisplay := second.displayName
	// Disambiguate when two specifiers share a name (alphabetize-named only).
	if firstDisplay == secondDisplay {
		if first.alias != "" {
			firstDisplay = firstDisplay + " as " + first.alias
		}
		if second.alias != "" {
			secondDisplay = secondDisplay + " as " + second.alias
		}
	}

	msg := rule.RuleMessage{
		Id: "order",
		Description: fmt.Sprintf(
			"`%s` %s should occur %s %s of `%s`",
			secondDisplay, makeImportDescription(second),
			order,
			makeImportDescription(first),
			firstDisplay,
		),
		Data: map[string]string{
			"first":      firstDisplay,
			"second":     secondDisplay,
			"firstKind":  makeImportDescription(first),
			"secondKind": makeImportDescription(second),
			"order":      order,
		},
	}

	ctx.ReportNodeWithDeferredFixes(second.node, msg, func() []rule.RuleFix {
		return buildSwapFix(first, second, order, source)
	})
}

func makeImportDescription(e *importEntry) string {
	switch e.typ {
	case "export":
		if e.importKind == "type" {
			return "type export"
		}
		return "export"
	}
	if e.importKind == "type" {
		return "type import"
	}
	if e.importKind == "typeof" {
		return "typeof import"
	}
	return "import"
}

// buildSwapFix produces a single replacement that swaps `first` and `second`
// in source order. Mirrors upstream `fixOutOfOrder` (non-named branch).
//
// Skips when:
//   - `canReorder` returns false (an unrelated statement sits between them)
//   - the two entries are already in different scopes (defensive)
//   - any of the line bounds are unresolvable
func buildSwapFix(first, second *importEntry, order string, source *sourceInfo) []rule.RuleFix {
	if first == nil || second == nil || first.root == nil || second.root == nil {
		return nil
	}
	isExports := first.typ == "export" && second.typ == "export"
	if !isExports && !canReorder(first.root, second.root) {
		return nil
	}

	lineStarts := source.lines()
	comments := source.allComments()
	firstStart, firstEnd := importutil.LineRangeWithComments(source.text, first.root, lineStarts, comments)
	secondStart, secondEnd := importutil.LineRangeWithComments(source.text, second.root, lineStarts, comments)
	if firstStart < 0 || firstEnd < 0 || secondStart < 0 || secondEnd < 0 {
		return nil
	}

	// Source-order invariant per upstream `fixOutOfOrder`:
	//   - "before": firstRoot appears earlier than secondRoot in source.
	//   - "after":  secondRoot appears earlier than firstRoot in source.
	// If either direction is violated, the caller paired entries wrong.
	if order == "before" && ast.CompareNodePositions(first.root, second.root) > 0 {
		return nil
	}
	if order == "after" && ast.CompareNodePositions(second.root, first.root) > 0 {
		return nil
	}

	newCode := source.text[secondStart:secondEnd]
	if !strings.HasSuffix(newCode, "\n") {
		newCode += "\n"
	}

	if order == "before" {
		replacement := newCode + source.text[firstStart:secondStart]
		return []rule.RuleFix{{
			Range: core.NewTextRange(firstStart, secondEnd),
			Text:  replacement,
		}}
	}
	// order == "after"
	gap := source.text[secondEnd:firstEnd]
	if gap != "" && !strings.HasSuffix(gap, "\n") && !strings.HasSuffix(gap, "\r") {
		lineEnding := "\n"
		if strings.HasSuffix(newCode, "\r\n") {
			lineEnding = "\r\n"
		}
		gap += lineEnding
	}
	replacement := gap + newCode
	return []rule.RuleFix{{
		Range: core.NewTextRange(secondStart, firstEnd),
		Text:  replacement,
	}}
}

// canReorder returns true when every statement between `first` and `second`
// (inclusive) is itself an import / require / export — i.e. moving them past
// each other won't change semantics.
//
// `node` here is the statement node passed during collection — we look it up
// in the parent block's statement list. Both sides MUST share a parent block;
// callers that detect cross-block pairs should refuse the fix earlier.
func canReorder(first, second *ast.Node) bool {
	body := siblingsOf(first)
	if body == nil {
		return false
	}
	firstIdx := slices.Index(body, first)
	secondIdx := slices.Index(body, second)
	if firstIdx < 0 || secondIdx < 0 {
		return false
	}
	if firstIdx > secondIdx {
		firstIdx, secondIdx = secondIdx, firstIdx
	}
	for i := firstIdx; i <= secondIdx; i++ {
		if !canCrossWhileReorder(body[i]) {
			return false
		}
	}
	return true
}

// siblingsOf returns the statement list that directly contains node. The
// tsgo AST owns the complete set of statement containers, so this keeps the
// rule correct if another supported container is added later.
func siblingsOf(node *ast.Node) []*ast.Node {
	if node == nil || node.Parent == nil || !node.Parent.CanHaveStatements() {
		return nil
	}
	return node.Parent.Statements()
}

// canCrossWhileReorder mirrors upstream `canCrossNodeWhileReorder`: only
// imports, top-level static requires (any `var x = require('foo').…` form),
// and ImportEqualsDeclaration are safe to reorder past.
func canCrossWhileReorder(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindImportDeclaration:
		return node.AsImportDeclaration().ImportClause != nil
	case ast.KindImportEqualsDeclaration:
		return ast.IsExternalModuleImportEqualsDeclaration(node)
	case ast.KindVariableStatement:
		return isSupportedRequireVarStatement(node)
	}
	return false
}

func isSupportedRequireVarStatement(node *ast.Node) bool {
	vs := node.AsVariableStatement()
	dl := vs.DeclarationList.AsVariableDeclarationList()
	if dl == nil || len(dl.Declarations.Nodes) != 1 {
		return false
	}
	d := dl.Declarations.Nodes[0].AsVariableDeclaration()
	if d.Initializer == nil {
		return false
	}
	name := d.Name()
	if name == nil || (name.Kind != ast.KindIdentifier && name.Kind != ast.KindObjectBindingPattern) {
		return false
	}
	if staticRequireCall(d.Initializer) != nil {
		return true
	}
	initializer := ast.SkipParentheses(d.Initializer)
	if initializer == nil || initializer.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(initializer.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	var receiver *ast.Node
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = callee.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		receiver = callee.AsElementAccessExpression().Expression
	default:
		return false
	}
	return staticRequireCall(receiver) != nil
}

// findRequireCall follows only member/call receiver chains. TypeScript-only
// wrappers intentionally stop the walk because ESTree's getRequireBlock does
// not cross them.
func findRequireCall(n *ast.Node) *ast.CallExpression {
	cur := ast.SkipParentheses(n)
	for cur != nil {
		if call := staticRequireCall(cur); call != nil {
			return call
		}
		var inner *ast.Node
		switch cur.Kind {
		case ast.KindCallExpression:
			if cur.AsCallExpression().QuestionDotToken != nil {
				return nil
			}
			inner = cur.AsCallExpression().Expression
		case ast.KindPropertyAccessExpression:
			if cur.AsPropertyAccessExpression().QuestionDotToken != nil {
				return nil
			}
			inner = cur.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			if cur.AsElementAccessExpression().QuestionDotToken != nil {
				return nil
			}
			inner = cur.AsElementAccessExpression().Expression
		default:
			return nil
		}
		cur = ast.SkipParentheses(inner)
	}
	return nil
}

func staticRequireCall(node *ast.Node) *ast.CallExpression {
	call := importutil.GetRequireCall(node, false)
	if call == nil || call.QuestionDotToken != nil {
		return nil
	}
	argument := ast.SkipParentheses(call.Arguments.Nodes[0])
	if argument == nil || argument.Kind != ast.KindStringLiteral {
		return nil
	}
	return call
}

// ---------------------------------------------------------------------------
// Newlines-between report
// ---------------------------------------------------------------------------

func makeNewlinesBetweenReport(ctx rule.RuleContext, imported []*importEntry, opts options, source *sourceInfo) {
	sourceText := source.text
	lineStarts := source.lines()
	if len(imported) < 2 {
		return
	}

	getEmptyLines := func(prev, cur *importEntry) int {
		prevEndLine := scanner.ComputeLineOfPosition(lineStarts, prev.node.End())
		curStartPos := scanner.SkipTrivia(sourceText, cur.node.Pos())
		curStartLine := scanner.ComputeLineOfPosition(lineStarts, curStartPos)
		count := 0
		for ln := prevEndLine + 1; ln < curStartLine; ln++ {
			lineStart := int(lineStarts[ln])
			lineEnd := len(sourceText)
			if ln+1 < len(lineStarts) {
				lineEnd = int(lineStarts[ln+1])
			}
			if utils.IsECMABlankLine(sourceText[lineStart:lineEnd]) {
				count++
			}
		}
		return count
	}

	prev := imported[0]
	for _, cur := range imported[1:] {
		empty := getEmptyLines(prev, cur)
		distinctStart := cur.rank-1 >= prev.rank

		isTypeOnly := cur.importKind == "type"
		isPrevTypeOnly := prev.importKind == "type"
		isNormalNextToTypeRelevant := isTypeOnly != isPrevTypeOnly && opts.isSortingTypesGroup
		isTypeOnlyRelevant := isTypeOnly && opts.isSortingTypesGroup

		nlBetween := opts.newlinesBetween
		nlBetweenTypes := opts.newlinesBetweenTypes
		if opts.isSortingTypesGroup && opts.consolidating &&
			(prev.isMultiline || cur.isMultiline) && nlBetween == "never" {
			nlBetween = "always-and-inside-groups"
		}
		if opts.isSortingTypesGroup && opts.consolidating &&
			(isNormalNextToTypeRelevant || prev.isMultiline || cur.isMultiline) &&
			nlBetweenTypes == "never" {
			nlBetweenTypes = "always-and-inside-groups"
		}

		notIgnored := (isTypeOnlyRelevant && nlBetweenTypes != "ignore") ||
			(!isTypeOnlyRelevant && nlBetween != "ignore")
		if !notIgnored {
			prev = cur
			continue
		}

		var shouldAssertNL, shouldAssertNoNLWithin, shouldAssertNoNLBetween bool
		if isTypeOnlyRelevant || isNormalNextToTypeRelevant {
			shouldAssertNL = nlBetweenTypes == "always" || nlBetweenTypes == "always-and-inside-groups"
			shouldAssertNoNLWithin = nlBetweenTypes != "always-and-inside-groups"
		} else {
			shouldAssertNL = nlBetween == "always" || nlBetween == "always-and-inside-groups"
			shouldAssertNoNLWithin = nlBetween != "always-and-inside-groups"
		}
		shouldAssertNoNLBetween = !opts.isSortingTypesGroup ||
			!isNormalNextToTypeRelevant ||
			nlBetweenTypes == "never"

		isSameGroup := opts.distinctGroup && cur.rank == prev.rank ||
			!opts.distinctGroup && !distinctStart

		alreadyReported := false
		if shouldAssertNL {
			if cur.rank != prev.rank && empty == 0 {
				if opts.distinctGroup || distinctStart {
					alreadyReported = true
					ctx.ReportNodeWithDeferredFixes(prev.node,
						rule.RuleMessage{
							Id:          "groupNewline",
							Description: "There should be at least one empty line between import groups",
						},
						func() []rule.RuleFix {
							return []rule.RuleFix{fixInsertNewlineAfter(prev.root, source)}
						},
					)
				}
			} else if empty > 0 && shouldAssertNoNLWithin {
				if isSameGroup {
					alreadyReported = true
					reportRemoveBlankLine(ctx, prev, cur, sourceText, lineStarts, "withinGroupNewline",
						"There should be no empty line within import group")
				}
			}
		} else if empty > 0 && shouldAssertNoNLBetween {
			alreadyReported = true
			reportRemoveBlankLine(ctx, prev, cur, sourceText, lineStarts, "groupNewline",
				"There should be no empty line between import groups")
		}

		if !alreadyReported && opts.consolidating {
			if empty == 0 && cur.isMultiline {
				ctx.ReportNodeWithDeferredFixes(prev.node,
					rule.RuleMessage{
						Id:          "consolidate",
						Description: "There should be at least one empty line between this import and the multi-line import that follows it",
					},
					func() []rule.RuleFix {
						return []rule.RuleFix{fixInsertNewlineAfter(prev.root, source)}
					},
				)
			} else if empty == 0 && prev.isMultiline {
				ctx.ReportNodeWithDeferredFixes(prev.node,
					rule.RuleMessage{
						Id:          "consolidate",
						Description: "There should be at least one empty line between this multi-line import and the import that follows it",
					},
					func() []rule.RuleFix {
						return []rule.RuleFix{fixInsertNewlineAfter(prev.root, source)}
					},
				)
			} else if empty > 0 && !prev.isMultiline && !cur.isMultiline && isSameGroup {
				reportRemoveBlankLine(ctx, prev, cur, sourceText, lineStarts, "consolidate",
					"There should be no empty lines between this single-line import and the single-line import that follows it")
			}
		}

		prev = cur
	}
}

func reportRemoveBlankLine(ctx rule.RuleContext, prev, cur *importEntry, sourceText string, lineStarts []core.TextPos, msgId, msgText string) {
	msg := rule.RuleMessage{Id: msgId, Description: msgText}
	ctx.ReportNodeWithDeferredFixes(prev.node, msg, func() []rule.RuleFix {
		fix := fixRemoveBlankLineBetween(sourceText, prev.root, cur.root, lineStarts)
		if fix == nil {
			return nil
		}
		return []rule.RuleFix{*fix}
	})
}

func fixInsertNewlineAfter(node *ast.Node, source *sourceInfo) rule.RuleFix {
	end := importutil.SameLineEndWithComments(source.text, node, source.lines(), source.allComments())
	return rule.RuleFix{
		Range: core.NewTextRange(end, end),
		Text:  "\n",
	}
}

func fixRemoveBlankLineBetween(text string, prev, cur *ast.Node, lineStarts []core.TextPos) *rule.RuleFix {
	prevEndLine := scanner.ComputeLineOfPosition(lineStarts, prev.End())
	curStartPos := scanner.SkipTrivia(text, cur.Pos())
	curStartLine := scanner.ComputeLineOfPosition(lineStarts, curStartPos)
	if curStartLine <= prevEndLine+1 {
		return nil
	}
	// Range to remove = end of prev's line through start of first non-blank
	// line at or before cur. Only safe if all intermediate lines are blank.
	prevLineEnd := len(text)
	if prevEndLine+1 < len(lineStarts) {
		prevLineEnd = int(lineStarts[prevEndLine+1])
	}
	curLineStart := int(lineStarts[curStartLine])
	if !utils.IsECMABlankLine(text[prevLineEnd:curLineStart]) {
		return nil
	}
	// Remove (curStartLine - prevEndLine - 1) blank lines.
	removeFrom := prevLineEnd
	removeTo := curLineStart
	f := rule.RuleFix{Range: core.NewTextRange(removeFrom, removeTo), Text: ""}
	return &f
}

// ---------------------------------------------------------------------------
// Alphabetize
// ---------------------------------------------------------------------------

// mutateRanksToAlphabetize re-ranks entries so that, within each group, items
// are sorted alphabetically per `opts`. Mirrors upstream literally — including
// the per-segment path comparison and the secondary `orderImportKind` sort.
func mutateRanksToAlphabetize(imported []*importEntry, opts alphabetizeOptions) {
	groups := map[float64][]*importEntry{}
	for _, e := range imported {
		groups[e.rank] = append(groups[e.rank], e)
	}

	keys := make([]float64, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Float64s(keys)

	multiplier := 1
	if opts.order == "desc" {
		multiplier = -1
	}
	multiplierKind := 0
	switch opts.orderImportKind {
	case "asc":
		multiplierKind = 1
	case "desc":
		multiplierKind = -1
	}

	cmp := makeAlphaComparator(opts.caseInsensitive, multiplier, multiplierKind)
	for _, k := range keys {
		sort.SliceStable(groups[k], func(i, j int) bool {
			return cmp(groups[k][i].value, groups[k][i].importKind,
				groups[k][j].value, groups[k][j].importKind) < 0
		})
	}

	newRanks := map[string]float64{}
	var newRank float64
	for _, k := range keys {
		for _, e := range groups[k] {
			newRanks[alphabetizedRankKey(e)] = k + newRank
			newRank++
		}
	}
	for _, e := range imported {
		if r, ok := newRanks[alphabetizedRankKey(e)]; ok {
			e.rank = r
		}
	}
}

func alphabetizedRankKey(entry *importEntry) string {
	return entry.value + "|" + entry.importKind
}

// makeAlphaComparator returns a comparator returning negative / zero / positive
// for (aValue, aKind) vs (bValue, bKind), honouring upstream's per-segment
// path comparison rules.
func makeAlphaComparator(caseInsensitive bool, multiplier, multiplierKind int) func(av, ak, bv, bk string) int {
	return func(av, ak, bv, bk string) int {
		va, vb := av, bv
		if caseInsensitive {
			va = stringutil.ToLowerJS(va)
			vb = stringutil.ToLowerJS(vb)
		}
		var result int
		if !strings.Contains(va, "/") && !strings.Contains(vb, "/") {
			result = utils.CompareJSStrings(va, vb)
		} else {
			A := strings.Split(va, "/")
			B := strings.Split(vb, "/")
			minLen := len(A)
			if len(B) < minLen {
				minLen = len(B)
			}
			for i := range minLen {
				if i == 0 && (A[i] == "." || A[i] == "..") && (B[i] == "." || B[i] == "..") {
					if A[i] != B[i] {
						result = utils.CompareJSStrings(va, vb)
						break
					}
					continue
				}
				if c := utils.CompareJSStrings(A[i], B[i]); c != 0 {
					result = c
					break
				}
			}
			if result == 0 && len(A) != len(B) {
				if len(A) < len(B) {
					result = -1
				} else {
					result = 1
				}
			}
		}
		result *= multiplier
		if result == 0 && multiplierKind != 0 {
			ka := ak
			if ka == "" {
				ka = "value"
			}
			kb := bk
			if kb == "" {
				kb = "value"
			}
			result = multiplierKind * utils.CompareJSStrings(ka, kb)
		}
		return result
	}
}

// ---------------------------------------------------------------------------
// Named ordering (named imports, exports, destructured require, cjs exports)
// ---------------------------------------------------------------------------

// makeNamedOrderReport sorts a single specifier list (e.g. the names inside
// `import { a, b } from 'x'`) and reports any out-of-order entries. Honours
// `named.types: 'mixed' | 'types-first' | 'types-last'`.
func makeNamedOrderReport(ctx rule.RuleContext, named []*namedEntry, opts options, sourceText string) {
	if len(named) <= 1 {
		return
	}

	// Build pseudo-rank from named.types: types-first → type=0,value=1; types-last → value=0,type=1; mixed → 0.
	namedGroups := []string{}
	switch opts.named.types {
	case "types-first":
		namedGroups = []string{"type"}
	case "types-last":
		namedGroups = []string{"value"}
	}

	rankOf := func(kind string) float64 {
		k := kind
		if k == "" {
			k = "value"
		}
		for i, g := range namedGroups {
			if g == k {
				return float64(i)
			}
		}
		return float64(len(namedGroups))
	}

	for _, e := range named {
		e.rank = rankOf(e.kind)
	}

	// Re-rank by alphabetical sort within each group.
	imps := make([]*importEntry, len(named))
	for i, n := range named {
		imps[i] = &importEntry{
			node:        n.node,
			root:        n.node,
			value:       n.value + ":" + n.alias,
			displayName: n.displayName,
			alias:       n.alias,
			typ:         n.typ,
			importKind:  n.kind,
			rank:        n.rank,
		}
	}
	if opts.alphabetize.order != "ignore" {
		mutateRanksToAlphabetize(imps, opts.alphabetize)
	}

	// Out-of-order detection (named flavour). When the reverse direction
	// yields fewer reports we use the reversed (negated-rank) list as the
	// search space, so the `found.rank > imp.rank` predicate still picks the
	// right partner.
	out := findOutOfOrder(imps)
	if len(out) == 0 {
		return
	}
	rev := reverseRanks(imps)
	revOut := findOutOfOrder(rev)
	if len(revOut) < len(out) {
		reportNamedOutOfOrder(ctx, rev, revOut, "after", sourceText)
		return
	}
	reportNamedOutOfOrder(ctx, imps, out, "before", sourceText)
}

func reportNamedOutOfOrder(ctx rule.RuleContext, all []*importEntry, outOfOrder []*importEntry, order string, sourceText string) {
	ordered := append([]*importEntry(nil), outOfOrder...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].node.Pos() < ordered[j].node.Pos()
	})
	for _, imp := range ordered {
		var found *importEntry
		for _, ii := range all {
			if ii.rank > imp.rank {
				found = ii
				break
			}
		}
		if found == nil {
			continue
		}
		first := found
		second := imp
		firstDisplay := first.displayName
		secondDisplay := second.displayName
		if firstDisplay == secondDisplay {
			if first.alias != "" {
				firstDisplay = firstDisplay + " as " + first.alias
			}
			if second.alias != "" {
				secondDisplay = secondDisplay + " as " + second.alias
			}
		}
		msg := rule.RuleMessage{
			Id: "order",
			Description: fmt.Sprintf(
				"`%s` %s should occur %s %s of `%s`",
				secondDisplay, makeImportDescription(second),
				order,
				makeImportDescription(first),
				firstDisplay,
			),
		}
		ctx.ReportNodeWithDeferredFixes(second.node, msg, func() []rule.RuleFix {
			return buildNamedSwapFix(ctx.SourceFile, sourceText, first, second, order)
		})
	}
}

func buildNamedSwapFix(sourceFile *ast.SourceFile, sourceText string, first, second *importEntry, order string) []rule.RuleFix {
	if sourceFile == nil || first == nil || second == nil {
		return nil
	}
	firstStart, firstEnd, ok := namedSpecifierBounds(sourceFile, first.node)
	if !ok {
		return nil
	}
	secondStart, secondEnd, ok := namedSpecifierBounds(sourceFile, second.node)
	if !ok {
		return nil
	}
	firstRange := utils.TrimNodeTextRange(sourceFile, first.node)
	secondRange := utils.TrimNodeTextRange(sourceFile, second.node)
	if firstStart > firstRange.Pos() || firstRange.End() > firstEnd ||
		secondStart > secondRange.Pos() || secondRange.End() > secondEnd {
		return nil
	}

	firstCode := sourceText[firstStart:firstRange.End()]
	firstTrivia := sourceText[firstRange.End():firstEnd]
	secondCode := sourceText[secondStart:secondRange.End()]
	secondTrivia := sourceText[secondRange.End():secondEnd]

	switch order {
	case "before":
		if firstRange.Pos() > secondRange.Pos() || secondStart == 0 || firstEnd > secondStart-1 {
			return nil
		}
		trimmedTrivia := strings.TrimRightFunc(secondTrivia, utils.IsStrWhiteSpace)
		gapCode := sourceText[firstEnd : secondStart-1]
		whitespaces := secondTrivia[len(trimmedTrivia):]
		return []rule.RuleFix{{
			Range: core.NewTextRange(firstStart, secondEnd),
			Text:  secondCode + "," + trimmedTrivia + firstCode + firstTrivia + gapCode + whitespaces,
		}}
	case "after":
		if secondRange.Pos() > firstRange.Pos() || secondEnd >= len(sourceText) || secondEnd+1 > firstStart {
			return nil
		}
		trimmedTrivia := strings.TrimRightFunc(firstTrivia, utils.IsStrWhiteSpace)
		gapCode := sourceText[secondEnd+1 : firstStart]
		whitespaces := firstTrivia[len(trimmedTrivia):]
		return []rule.RuleFix{{
			Range: core.NewTextRange(secondStart, firstEnd),
			Text:  gapCode + firstCode + "," + trimmedTrivia + secondCode + whitespaces,
		}}
	default:
		return nil
	}
}

func namedSpecifierBounds(sourceFile *ast.SourceFile, node *ast.Node) (int, int, bool) {
	if sourceFile == nil || node == nil {
		return 0, 0, false
	}
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	search := nodeRange.Pos()
	start := -1
	for search > 0 {
		token, ok := utils.TokenBeforePosition(sourceFile, search)
		if !ok {
			break
		}
		if token.Kind == ast.KindCommaToken || token.Kind == ast.KindOpenBraceToken {
			start = token.End
			break
		}
		search = token.Start
	}
	search = nodeRange.End()
	end := -1
	for search < len(sourceFile.Text()) {
		token, ok := utils.TokenAtOrAfter(sourceFile, search)
		if !ok {
			break
		}
		if token.Kind == ast.KindCommaToken || token.Kind == ast.KindCloseBraceToken {
			end = token.Start
			break
		}
		search = token.End
	}
	return start, end, start >= 0 && end >= start
}

// ---------------------------------------------------------------------------
// Statement walking
// ---------------------------------------------------------------------------

// blockState is the per-statement-list bookkeeping. Each ModuleBlock and the
// SourceFile get their own `blockState`, mirroring upstream's `node.parent`-
// keyed `importMap`.
//
// `exports` mirrors upstream's `exportMap` for alphabetizing named CommonJS
// export assignments; `namedLists` handles names within one declaration.
type blockState struct {
	imports []*importEntry
	exports []*importEntry
	// namedLists collects deferred named-order reports — populated when we
	// see an import / export / require / module.exports = {} that has
	// candidate named children.
	namedLists [][]*namedEntry
}

func (bs *blockState) addImport(e *importEntry)    { bs.imports = append(bs.imports, e) }
func (bs *blockState) addNamed(list []*namedEntry) { bs.namedLists = append(bs.namedLists, list) }
func (bs *blockState) addExport(e *importEntry)    { bs.exports = append(bs.exports, e) }

// processBlock walks the `statements` list of one block and dispatches each
// node to the appropriate handler. When it sees a `declare module {}` or
// `declare namespace {}`, it recurses into the inner ModuleBlock.
func processBlock(ctx rule.RuleContext, statements []*ast.Node, opts options, results map[*ast.Node]*blockState, container *ast.Node) {
	bs := &blockState{}
	results[container] = bs

	for _, stmt := range statements {
		switch stmt.Kind {
		case ast.KindImportDeclaration:
			handleImportDeclaration(ctx, stmt, opts, bs)
		case ast.KindImportEqualsDeclaration:
			handleImportEqualsDeclaration(ctx, stmt, opts, bs)
		case ast.KindVariableStatement:
			if container.Kind == ast.KindSourceFile {
				handleVariableStatement(ctx, stmt, opts, bs)
			}
		case ast.KindExportDeclaration:
			handleExportDeclaration(ctx, stmt, opts, bs)
		case ast.KindModuleDeclaration:
			// Recurse into `declare module 'x' { ... }` /
			// `namespace Foo { ... }`.
			recurseIntoModuleDeclaration(ctx, stmt, opts, results)
		}
	}
}

func recurseIntoModuleDeclaration(ctx rule.RuleContext, decl *ast.Node, opts options, results map[*ast.Node]*blockState) {
	body := decl.AsModuleDeclaration().Body
	if body == nil {
		return
	}
	switch body.Kind {
	case ast.KindModuleBlock:
		blk := body.AsModuleBlock()
		if blk.Statements != nil {
			processBlock(ctx, blk.Statements.Nodes, opts, results, body)
		}
	case ast.KindModuleDeclaration:
		// `declare module A.B { ... }` nests as ModuleDeclaration → ModuleDeclaration → ModuleBlock.
		recurseIntoModuleDeclaration(ctx, body, opts, results)
	}
}

func handleImportDeclaration(ctx rule.RuleContext, stmt *ast.Node, opts options, bs *blockState) {
	decl := stmt.AsImportDeclaration()
	if decl.ModuleSpecifier == nil || decl.ModuleSpecifier.Kind != ast.KindStringLiteral {
		return
	}
	hasSpecifiers := decl.ImportClause != nil
	// Upstream: `if (node.specifiers.length || options.warnOnUnassignedImports)`.
	// `specifiers` includes default + namespace + named — `ImportClause` is
	// non-nil iff there's at least one binding, so this maps cleanly.
	if !hasSpecifiers && !opts.warnOnUnassigned {
		return
	}
	value := decl.ModuleSpecifier.AsStringLiteral().Text
	kind := importKindOf(stmt)
	entry := &importEntry{
		node:         stmt,
		root:         stmt,
		value:        value,
		displayName:  value,
		typ:          "import",
		importKind:   kind,
		classifyType: importutil.ClassifyImport(ctx, value, decl.ModuleSpecifier, opts.internalRegex),
		isMultiline:  isMultiline(stmt, ctx.SourceFile),
	}
	bs.addImport(entry)

	if opts.named.imports && decl.ImportClause != nil {
		clause := decl.ImportClause.AsImportClause()
		if clause != nil && clause.NamedBindings != nil &&
			clause.NamedBindings.Kind == ast.KindNamedImports {
			ni := clause.NamedBindings.AsNamedImports()
			named := collectNamedImports(ni)
			if len(named) > 1 {
				bs.addNamed(named)
			}
		}
	}
}

func collectNamedImports(ni *ast.NamedImports) []*namedEntry {
	if ni == nil || ni.Elements == nil {
		return nil
	}
	out := make([]*namedEntry, 0, len(ni.Elements.Nodes))
	for _, spec := range ni.Elements.Nodes {
		if spec.Kind != ast.KindImportSpecifier {
			continue
		}
		s := spec.AsImportSpecifier()
		propName, localName, ok := importSpecNames(s)
		if !ok {
			// One unsupported entry makes the list non-deterministic, so match
			// the require/CommonJS paths and leave the whole list untouched.
			return nil
		}
		kind := ""
		if s.IsTypeOnly {
			kind = "type"
		}
		alias := ""
		if s.PropertyName != nil {
			alias = localName
		}
		out = append(out, &namedEntry{
			node:        spec,
			value:       propName,
			displayName: propName,
			alias:       alias,
			typ:         "import",
			kind:        kind,
		})
	}
	return out
}

func importSpecNames(s *ast.ImportSpecifier) (propName, localName string, ok bool) {
	if s == nil || s.Name() == nil {
		return "", "", false
	}
	if s.PropertyName != nil {
		propName, ok = utils.GetStaticPropertyName(s.PropertyName)
	} else {
		propName, ok = utils.GetStaticPropertyName(s.Name())
	}
	if !ok {
		return "", "", false
	}
	localName, ok = utils.GetStaticPropertyName(s.Name())
	return propName, localName, ok
}

func handleImportEqualsDeclaration(ctx rule.RuleContext, stmt *ast.Node, opts options, bs *blockState) {
	if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
		return
	}
	eq := stmt.AsImportEqualsDeclaration()
	var value, displayName, typ, classify string
	if ast.IsExternalModuleImportEqualsDeclaration(stmt) {
		expr := ast.GetExternalModuleImportEqualsDeclarationExpression(stmt)
		if expr.Kind != ast.KindStringLiteral {
			return
		}
		value = expr.AsStringLiteral().Text
		displayName = value
		typ = "import"
		classify = importutil.ClassifyImport(ctx, value, expr, opts.internalRegex)
	} else {
		value = ""
		displayName = utils.TrimmedNodeText(ctx.SourceFile, eq.ModuleReference)
		typ = "import:object"
		classify = "object"
	}
	kind := ""
	if eq.IsTypeOnly {
		kind = "type"
	}
	bs.addImport(&importEntry{
		node:         stmt,
		root:         stmt,
		value:        value,
		displayName:  displayName,
		typ:          typ,
		importKind:   kind,
		classifyType: classify,
		isMultiline:  isMultiline(stmt, ctx.SourceFile),
	})
}

func handleVariableStatement(ctx rule.RuleContext, stmt *ast.Node, opts options, bs *blockState) {
	vs := stmt.AsVariableStatement()
	dl := vs.DeclarationList.AsVariableDeclarationList()
	if dl == nil {
		return
	}
	for _, declarationNode := range dl.Declarations.Nodes {
		d := declarationNode.AsVariableDeclaration()
		if d.Initializer == nil {
			continue
		}
		ce := findRequireCall(d.Initializer)
		if ce == nil || ce.Arguments == nil || len(ce.Arguments.Nodes) != 1 {
			continue
		}
		arg := ast.SkipParentheses(ce.Arguments.Nodes[0])
		if arg == nil || arg.Kind != ast.KindStringLiteral {
			continue
		}
		value := arg.AsStringLiteral().Text
		multilineNode := ce.AsNode()
		// ESTree's registerNode promotes only a require call that is the direct
		// VariableDeclarator initializer to the whole declaration. A call inside
		// `require("x").member` stays the CallExpression for multiline checks.
		if ast.SkipParentheses(d.Initializer) == ce.AsNode() {
			multilineNode = stmt
		}
		bs.addImport(&importEntry{
			node:         ce.AsNode(),
			root:         stmt,
			value:        value,
			displayName:  value,
			typ:          "require",
			classifyType: importutil.ClassifyImport(ctx, value, arg, opts.internalRegex),
			isMultiline:  isMultiline(multilineNode, ctx.SourceFile),
		})

	}
}

func collectNamedFromBindingPattern(pat *ast.Node) []*namedEntry {
	bp := pat.AsBindingPattern()
	if bp == nil || bp.Elements == nil {
		return nil
	}
	out := make([]*namedEntry, 0, len(bp.Elements.Nodes))
	for _, el := range bp.Elements.Nodes {
		if el.Kind != ast.KindBindingElement {
			continue
		}
		be := el.AsBindingElement()
		if be.DotDotDotToken != nil || be.Initializer != nil {
			// ESTree represents these as RestElement / AssignmentPattern rather
			// than Property nodes with identifier values, so upstream abandons
			// named sorting for the whole destructuring list.
			return nil
		}
		if be.PropertyName != nil && be.PropertyName.Kind != ast.KindIdentifier {
			// Computed / numeric / string-literal key — bail out (entire list).
			return nil
		}
		if be.Name() == nil || be.Name().Kind != ast.KindIdentifier {
			return nil
		}
		var prop, local string
		if be.PropertyName != nil {
			prop = be.PropertyName.AsIdentifier().Text
		} else {
			prop = be.Name().AsIdentifier().Text
		}
		local = be.Name().AsIdentifier().Text
		alias := ""
		if be.PropertyName != nil {
			alias = local
		}
		out = append(out, &namedEntry{
			node:        el,
			value:       prop,
			displayName: prop,
			alias:       alias,
			typ:         "require",
		})
	}
	return out
}

func handleExportDeclaration(ctx rule.RuleContext, stmt *ast.Node, opts options, bs *blockState) {
	if !opts.named.exports {
		return
	}
	ed := stmt.AsExportDeclaration()
	if ed.ExportClause == nil || ed.ExportClause.Kind != ast.KindNamedExports {
		return
	}
	ne := ed.ExportClause.AsNamedExports()
	if ne.Elements == nil {
		return
	}
	out := make([]*namedEntry, 0, len(ne.Elements.Nodes))
	for _, spec := range ne.Elements.Nodes {
		if spec.Kind != ast.KindExportSpecifier {
			continue
		}
		s := spec.AsExportSpecifier()
		if s.Name() == nil {
			continue
		}
		var prop, exp string
		var ok bool
		if s.PropertyName != nil {
			prop, ok = utils.GetStaticPropertyName(s.PropertyName)
		} else {
			prop, ok = utils.GetStaticPropertyName(s.Name())
		}
		if !ok {
			return
		}
		exp, ok = utils.GetStaticPropertyName(s.Name())
		if !ok {
			return
		}
		alias := ""
		if s.PropertyName != nil {
			alias = exp
		}
		kind := ""
		if s.IsTypeOnly {
			kind = "type"
		}
		out = append(out, &namedEntry{
			node:        spec,
			value:       prop,
			displayName: prop,
			alias:       alias,
			typ:         "export",
			kind:        kind,
		})
	}
	if len(out) > 1 {
		bs.addNamed(out)
	}
}

// handleExpressionStatement deals with `module.exports = { ... }` (cjsExports).
// Upstream restricts the form to identifier-keyed AND identifier-valued
// shorthand or longhand assignments — `{ a, b }` or `{ a: aVal, b: bVal }` —
// because anything else (`{ a: 1 }`, `{ ['x']: y }`, computed names) can't be
// reliably re-emitted as a deterministic name list. We follow the same shape
// check.
//
// Distinguishing the global `module.exports` from a local binding requires
// scope information. The rule's reference map resolves each receiver in its
// lexical scope; without that map, the structural global form is accepted.
func handleExpressionStatement(ctx rule.RuleContext, stmt *ast.Node, opts options, bs *blockState) {
	if !opts.named.cjsExports {
		return
	}
	expr := ast.SkipParentheses(stmt.AsExpressionStatement().Expression)
	if expr == nil || expr.Kind != ast.KindBinaryExpression {
		return
	}
	be := expr.AsBinaryExpression()
	if be.OperatorToken == nil || be.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}
	if !isCJSExportsTarget(ctx, be.Left) {
		nameParts := getNamedCJSExportParts(ctx, be.Left)
		if len(nameParts) > 0 {
			name := strings.Join(nameParts, ".")
			bs.addExport(&importEntry{
				node:        expr,
				root:        stmt,
				value:       name,
				displayName: name,
				typ:         "export",
				rank:        0,
			})
		}
		return
	}
	right := ast.SkipParentheses(be.Right)
	if right == nil || right.Kind != ast.KindObjectLiteralExpression {
		return
	}
	ole := right.AsObjectLiteralExpression()
	if ole.Properties == nil {
		return
	}
	out := make([]*namedEntry, 0, len(ole.Properties.Nodes))
	for _, p := range ole.Properties.Nodes {
		if p.Kind != ast.KindPropertyAssignment && p.Kind != ast.KindShorthandPropertyAssignment {
			return
		}
		var keyText, valText string
		switch p.Kind {
		case ast.KindPropertyAssignment:
			pa := p.AsPropertyAssignment()
			var ok bool
			keyText, ok = cjsIdentifierPropertyName(pa.Name())
			if !ok {
				return
			}
			initializer := ast.SkipParentheses(pa.Initializer)
			if initializer == nil || initializer.Kind != ast.KindIdentifier {
				return
			}
			valText = initializer.AsIdentifier().Text
		case ast.KindShorthandPropertyAssignment:
			spa := p.AsShorthandPropertyAssignment()
			if spa.ObjectAssignmentInitializer != nil || spa.PostfixToken != nil || spa.Type != nil {
				return
			}
			if spa.Name() == nil || spa.Name().Kind != ast.KindIdentifier {
				return
			}
			keyText = spa.Name().AsIdentifier().Text
			valText = keyText
		}
		alias := ""
		if p.Kind == ast.KindPropertyAssignment {
			alias = valText
		}
		out = append(out, &namedEntry{
			node:        p,
			value:       keyText,
			displayName: keyText,
			alias:       alias,
			typ:         "export",
		})
	}
	if len(out) > 1 {
		bs.addNamed(out)
	}
}

// cjsIdentifierPropertyName implements the deliberately narrow ESTree shape
// accepted upstream: the property key itself must be an Identifier. In tsgo a
// computed identifier such as [name] is wrapped in ComputedPropertyName, so we
// peel only that wrapper; string/numeric/static computed keys still bail out.
func cjsIdentifierPropertyName(name *ast.Node) (string, bool) {
	if name == nil {
		return "", false
	}
	if name.Kind == ast.KindComputedPropertyName {
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	if name == nil || name.Kind != ast.KindIdentifier {
		return "", false
	}
	return name.AsIdentifier().Text, true
}

// handleNamedRequireDeclaration mirrors upstream's VariableDeclarator
// listener. It intentionally accepts direct object-destructuring requires in
// every lexical scope, while module ordering itself remains Program-only.
func handleNamedRequireDeclaration(node *ast.Node, results map[*ast.Node]*blockState) {
	declaration := node.AsVariableDeclaration()
	if declaration.Initializer == nil || staticRequireCall(declaration.Initializer) == nil ||
		declaration.Name() == nil || declaration.Name().Kind != ast.KindObjectBindingPattern {
		return
	}
	named := collectNamedFromBindingPattern(declaration.Name())
	if len(named) > 1 {
		container := nearestStatementListContainer(node)
		if container == nil {
			return
		}
		stateForContainer(results, container).addNamed(named)
	}
}

// nearestStatementListContainer delegates the set of valid statement-list
// owners to tsgo instead of duplicating its SourceFile/Block/CaseBlock kinds.
func nearestStatementListContainer(node *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.CanHaveStatements() {
			return parent
		}
	}
	return nil
}

func stateForContainer(results map[*ast.Node]*blockState, container *ast.Node) *blockState {
	if state := results[container]; state != nil {
		return state
	}
	state := &blockState{}
	results[container] = state
	return state
}

// isCJSExportsTarget returns true when the LHS of an assignment refers to the
// global `module.exports` or `exports` identifier without local shadowing.
//
// ESTree treats `module.exports` and `module[exports]` alike here because both
// expose an Identifier property; string-literal element access remains a
// Literal and is intentionally rejected. The reference store then determines
// whether `module` / `exports` resolves to a value declared in this file.
func isCJSExportsTarget(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	// `exports` (bare identifier).
	if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == "exports" {
		return !isShadowedGlobal(ctx, node)
	}
	// `module.exports` / `module[exports]`.
	var receiver, property *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		receiver, property = access.Expression, access.Name()
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		receiver, property = access.Expression, ast.SkipParentheses(access.ArgumentExpression)
	default:
		return false
	}
	if property == nil || property.Kind != ast.KindIdentifier || property.Text() != "exports" {
		return false
	}
	moduleIdent := ast.SkipParentheses(receiver)
	if moduleIdent == nil || moduleIdent.Kind != ast.KindIdentifier || moduleIdent.Text() != "module" {
		return false
	}
	return !isShadowedGlobal(ctx, moduleIdent)
}

func getNamedCJSExportParts(ctx rule.RuleContext, node *ast.Node) []string {
	if node == nil {
		return nil
	}
	root := ast.SkipParentheses(node)
	var parent *ast.Node
	var result []string
	for root != nil {
		var receiver *ast.Node
		var property *ast.Node
		switch root.Kind {
		case ast.KindPropertyAccessExpression:
			access := root.AsPropertyAccessExpression()
			receiver = access.Expression
			property = access.Name()
		case ast.KindElementAccessExpression:
			access := root.AsElementAccessExpression()
			receiver = access.Expression
			property = access.ArgumentExpression
		default:
			goto done
		}
		property = ast.SkipParentheses(property)
		if property == nil || property.Kind != ast.KindIdentifier {
			return nil
		}
		result = append([]string{property.AsIdentifier().Text}, result...)
		parent = root
		root = ast.SkipParentheses(receiver)
	}

done:
	if isCJSExportsTarget(ctx, root) {
		return result
	}
	if isCJSExportsTarget(ctx, parent) && len(result) > 1 {
		return result[1:]
	}
	return nil
}

// isShadowedGlobal returns true when the identifier resolves to a user-
// declared symbol (i.e. NOT the ambient/global `module` or `exports`).
//
// A value declaration authored in this source file shadows the CommonJS
// global; library declarations and ambient augmentations outside it do not.
func isShadowedGlobal(ctx rule.RuleContext, ident *ast.Node) bool {
	if ctx.Refs == nil || ctx.SourceFile == nil {
		return false
	}
	return utils.IsValueSymbolDeclaredInFile(ctx.Refs.Resolve(ident), ctx.SourceFile)
}

// importKindOf returns "type" for `import type X from 'mod'` (whole-declaration
// type-only). Note: `import { type X } from 'mod'` is per-specifier and
// classified as a value import for ordering purposes — upstream's
// `node.importKind` only reflects the leading `type` keyword.
func importKindOf(node *ast.Node) string {
	if node.Kind != ast.KindImportDeclaration {
		return ""
	}
	clause := node.AsImportDeclaration().ImportClause
	if clause == nil {
		return ""
	}
	if ast.IsTypeOnlyImportDeclaration(clause) {
		return "type"
	}
	return ""
}

// isMultiline reports whether the node spans more than one logical line.
//
// `node.Pos()` includes the node's leading trivia, which for any non-first
// statement begins on the previous line — so we step over trivia with
// `SkipTrivia` to get the real first line of the node, mirroring ESLint's
// `loc.start.line` which points at the first non-trivia token.
//
// Line numbers come from `GetECMALineAndUTF16CharacterOfPosition`, which
// recognizes every line terminator the host scanner does (LF / CRLF /
// U+2028 / U+2029).
func isMultiline(node *ast.Node, sf *ast.SourceFile) bool {
	if sf == nil {
		return false
	}
	tokenStart := scanner.SkipTrivia(sf.Text(), node.Pos())
	startLine, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, tokenStart)
	endLine, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, node.End())
	return startLine != endLine
}

// ---------------------------------------------------------------------------
// Rule entry
// ---------------------------------------------------------------------------

// OrderRule is the exported `import/order` rule.
//
//go:embed order.schema.json
var schemaJSON []byte

var OrderRule = rule.Rule{
	Name:   "import/order",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		var raw any
		if len(rawOptions) > 0 {
			raw = rawOptions[0]
		}
		opts := parseOptions(raw, ctx.Settings)
		sf := ctx.SourceFile
		if sf == nil {
			return rule.RuleListeners{}
		}

		// Surface a single Program-level diagnostic for malformed `groups`,
		// matching upstream's `try { ... } catch (error) { Program(node) {} }`.
		if opts.rankErr != nil {
			ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{
				Id:          "configError",
				Description: opts.rankErr.Error(),
			})
			return rule.RuleListeners{}
		}

		if sf.Statements == nil {
			return rule.RuleListeners{}
		}

		source := &sourceInfo{file: sf, commentStore: ctx.Comments, text: sf.Text()}

		// Collect module ordering from statement lists. Named require/CommonJS
		// forms use the linter's shared traversal below so nested scopes are
		// visited without a second recursive walk of the file.
		results := map[*ast.Node]*blockState{}
		processBlock(ctx, sf.Statements.Nodes, opts, results, sf.AsNode())

		listeners := rule.RuleListeners{}
		if opts.named.require {
			listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
				handleNamedRequireDeclaration(node, results)
			}
		}
		if opts.named.cjsExports {
			listeners[ast.KindExpressionStatement] = func(node *ast.Node) {
				// Upstream keys exportMap by AssignmentExpression.parent.parent,
				// which is the ExpressionStatement's direct parent. This matters
				// for unbraced if/loop bodies: distinct control-flow statements must
				// never be merged into the SourceFile export sequence.
				if node.Parent != nil {
					handleExpressionStatement(ctx, node, opts, stateForContainer(results, node.Parent))
				}
			}
		}
		listeners[rule.ListenerOnExit(ast.KindEndOfFile)] = func(*ast.Node) {
			// Apply per-block reports in document order to keep diagnostics
			// stable (matters for test snapshots).
			for _, container := range blockKeysInDocOrder(results) {
				finalizeBlock(ctx, results[container], opts, source)
			}
		}
		return listeners
	},
}

// finalizeBlock runs newlines-between, alphabetize, out-of-order, and named
// reports for one block, in upstream's order.
func finalizeBlock(ctx rule.RuleContext, bs *blockState, opts options, source *sourceInfo) {
	// Compute ranks; drop entries that fail to classify (rank == -1).
	imported := bs.imports[:0]
	for _, e := range bs.imports {
		r := computeRank(e, opts)
		if r == -1 {
			continue
		}
		e.rank = r
		imported = append(imported, e)
	}

	if opts.newlinesBetween != "ignore" || opts.newlinesBetweenTypes != "ignore" {
		makeNewlinesBetweenReport(ctx, imported, opts, source)
	}

	if opts.alphabetize.order != "ignore" {
		mutateRanksToAlphabetize(imported, opts.alphabetize)
	}

	makeOutOfOrderReports(ctx, imported, source)

	// ES lists are gathered from statement lists and CommonJS lists from the
	// linter's traversal. Restore source order before reporting mixed syntax.
	sort.SliceStable(bs.namedLists, func(i, j int) bool {
		return ast.CompareNodePositions(bs.namedLists[i][0].node, bs.namedLists[j][0].node) < 0
	})
	for _, list := range bs.namedLists {
		makeNamedOrderReport(ctx, list, opts, source.text)
	}

	if opts.alphabetize.order != "ignore" && len(bs.exports) > 1 {
		mutateRanksToAlphabetize(bs.exports, opts.alphabetize)
		makeOutOfOrderReports(ctx, bs.exports, source)
	}
}

// blockKeysInDocOrder returns the block-container nodes sorted by their start
// position. Iterating in this order keeps reports aligned with the source.
func blockKeysInDocOrder(results map[*ast.Node]*blockState) []*ast.Node {
	keys := make([]*ast.Node, 0, len(results))
	for k := range results {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return ast.CompareNodePositions(keys[i], keys[j]) < 0
	})
	return keys
}
