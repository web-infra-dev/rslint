// Package order ports eslint-plugin-import's `order` rule to rslint.
//
// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/order.md
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
	"github.com/microsoft/typescript-go/shim/tspath"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// allTypes lists every public import group. Its order is used only while
// assigning the fallback rank for groups omitted from configuration.
var allTypes = []string{
	"builtin", "external", "internal", "unknown",
	"parent", "sibling", "index", "object", "type",
}

// defaultGroups is the rule's documented default group order.
var defaultGroups = []any{"builtin", "external", "parent", "sibling", "index"}

// These defaults are immutable after initialization. Reusing them avoids
// rebuilding identical maps and ranks for every file using the common
// no-options configuration.
var defaultPathGroupsExcludedImportTypes = map[string]bool{
	"builtin":  true,
	"external": true,
	"object":   true,
}

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
	patternOptions    minimatch3.Options
	patternOptionsSet bool
	matcher           *minimatch3.Matcher
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

var defaultRanks = func() ranks {
	r, err := buildRanks(defaultGroups, nil)
	if err != nil {
		panic(err)
	}
	return r
}()

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
	typeGroupInGroups   bool
	isSortingTypesGroup bool
	consolidating       bool
}

// parseOptions consumes the normalized ESLint-style options tuple. A malformed
// groups array is recorded so Run can report one file-level configuration
// error without risking a panic in a syntax listener.
func parseOptions(rawOptions []any) options {
	opts := options{
		groups:                        defaultGroups,
		pathGroupsExcludedImportTypes: defaultPathGroupsExcludedImportTypes,
		distinctGroup:                 defaultDistinctGroup,
		newlinesBetween:               "ignore",
		alphabetize:                   alphabetizeOptions{order: "ignore", orderImportKind: "ignore"},
		named:                         namedOptions{types: "mixed"},
		consolidateIslands:            "never",
		ranks:                         defaultRanks,
	}

	var raw any
	if len(rawOptions) > 0 {
		raw = rawOptions[0]
	}
	m, _ := raw.(map[string]any)
	customRanks := false
	if m != nil {
		if g, ok := m["groups"].([]any); ok {
			opts.groups = g
			customRanks = true
		}
		if pgs, ok := m["pathGroups"].([]any); ok {
			opts.pathGroups = parsePathGroups(pgs)
			customRanks = true
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

	// The public option uses JavaScript's `||` fallback, so both an omitted value
	// and an empty string inherit `newlines-between`.
	if opts.newlinesBetweenTypes == "" {
		opts.newlinesBetweenTypes = opts.newlinesBetween
	}

	if customRanks {
		r, err := buildRanks(opts.groups, opts.pathGroups)
		if err != nil {
			opts.rankErr = err
		}
		opts.ranks = r
	}

	opts.typeGroupInGroups = !slices.Contains(opts.ranks.omittedTypes, "type")
	opts.isSortingTypesGroup = opts.typeGroupInGroups && opts.sortTypesGroup

	opts.consolidating = opts.consolidateIslands == "inside-groups" &&
		(opts.newlinesBetween == "always-and-inside-groups" ||
			opts.newlinesBetweenTypes == "always-and-inside-groups")

	return opts
}

// parseNamed handles both the boolean shorthand (`named: true|false`) and the
// full object form. Missing per-syntax toggles (`import`, `export`, etc.)
// inherit `enabled`.
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
	// cspell:ignore nonegate nocomment noext nobrace nonull
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
			pg.patternOptions.NoNull, _ = rawOptions["nonull"].(bool)
		}
		matcherOptions := pg.patternOptions
		if !pg.patternOptionsSet {
			matcherOptions.NoComment = true
		}
		pg.matcher = minimatch3.New(pg.pattern, matcherOptions)
		out = append(out, pg)
	}
	return out
}

// buildRanks converts the user's `groups` and `pathGroups` configuration into
// numeric ranks. Group entries occupy even ranks; path-group positions use the
// gaps between them.
func buildRanks(groupsCfg []any, pathGroupsCfg []pathGroup) (ranks, error) {
	rankObject := make(map[string]float64)
	for i, g := range groupsCfg {
		base := float64(i) * 2
		switch v := g.(type) {
		case string:
			if !slices.Contains(allTypes, v) {
				return ranks{}, fmt.Errorf("incorrect configuration of the rule: unknown type `%q`", v)
			}
			// JSON Schema's uniqueItems applies only within this array. A type
			// repeated across a scalar and a nested group remains schema-valid;
			// JavaScript object assignment makes the later rank win.
			rankObject[v] = base
		case []any:
			for _, item := range v {
				s, ok := item.(string)
				if !ok || !slices.Contains(allTypes, s) {
					return ranks{}, fmt.Errorf("incorrect configuration of the rule: unknown type `%v`", item)
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

// convertPathGroupsForRanks turns before/after positions into fractional rank
// offsets. maxPosition keeps every offset inside its configured parent group.
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
	// node is the syntax node reported to the user: an import/export
	// declaration, require call, or CommonJS assignment.
	node *ast.Node
	// statement is the tsgo statement moved by whole-item fixes. It can be
	// wider than node when an expression belongs to unbraced control flow.
	statement *ast.Node
	// specifier is retained for lazy module resolution. Files with fewer than
	// two orderable entries never pay for classification.
	specifier *ast.Node

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

// pendingReport keeps detection separate from emission. import/order gathers
// several independent statement lists (the source file, nested namespaces,
// blocks containing CommonJS exports, and named-specifier lists). Their
// containers can overlap, so emitting one container at a time can put a later
// outer diagnostic before an earlier nested one. We queue only the diagnostic
// identity and the deferred fixer, then emit by the actual report node.
type pendingReport struct {
	node  *ast.Node
	msg   rule.RuleMessage
	fixes func() []rule.RuleFix
}

type reportQueue struct {
	reports []pendingReport
}

func (queue *reportQueue) add(node *ast.Node, msg rule.RuleMessage, fixes func() []rule.RuleFix) {
	queue.reports = append(queue.reports, pendingReport{node: node, msg: msg, fixes: fixes})
}

func (queue *reportQueue) flush(ctx rule.RuleContext) {
	if len(queue.reports) > 1 {
		slices.SortStableFunc(queue.reports, func(left, right pendingReport) int {
			return ast.CompareNodePositions(left.node, right.node)
		})
	}
	for _, report := range queue.reports {
		ctx.ReportNodeWithDeferredFixes(report.node, report.msg, report.fixes)
	}
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

// Whole-statement fixes keep adjacent comments on the statement's ending
// line. This is deliberately rule-local: it describes what import/order moves,
// not a general comment-attachment policy for other rules.
const maxAdjacentComments = 100

func statementLineRange(text string, node *ast.Node, lineStarts []core.TextPos, comments []*ast.CommentRange) (int, int) {
	return startOfLineWithComments(text, node, lineStarts, comments),
		endOfLineWithComments(text, node, lineStarts, comments)
}

func startOfLineWithComments(text string, node *ast.Node, lineStarts []core.TextPos, comments []*ast.CommentRange) int {
	if node == nil {
		return -1
	}
	start := scanner.SkipTrivia(text, node.Pos())
	endingLine := scanner.ComputeLineOfPosition(lineStarts, node.End())
	index := sort.Search(len(comments), func(i int) bool { return comments[i].End() > start }) - 1
	for seen := 0; index >= 0 && seen < maxAdjacentComments; index, seen = index-1, seen+1 {
		comment := comments[index]
		if scanner.ComputeLineOfPosition(lineStarts, comment.Pos()) != endingLine ||
			scanner.ComputeLineOfPosition(lineStarts, comment.End()) != endingLine ||
			!onlyHorizontalWhitespace(text, comment.End(), start) {
			break
		}
		start = comment.Pos()
	}
	for index := start - 1; index > 0; index-- {
		if text[index] != ' ' && text[index] != '\t' {
			break
		}
		start = index
	}
	return start
}

func endOfLineWithComments(text string, node *ast.Node, lineStarts []core.TextPos, comments []*ast.CommentRange) int {
	if node == nil {
		return -1
	}
	end := sameLineEndWithComments(text, node, lineStarts, comments)
	for end < len(text) {
		switch text[end] {
		case ' ', '\t', '\r':
			end++
		case '\n':
			return end + 1
		default:
			return end
		}
	}
	return end
}

func sameLineEndWithComments(text string, node *ast.Node, lineStarts []core.TextPos, comments []*ast.CommentRange) int {
	if node == nil {
		return -1
	}
	endingLine := scanner.ComputeLineOfPosition(lineStarts, node.End())
	end := node.End()
	index := sort.Search(len(comments), func(i int) bool { return comments[i].Pos() >= end })
	for seen := 0; index < len(comments) && seen < maxAdjacentComments; index, seen = index+1, seen+1 {
		comment := comments[index]
		if scanner.ComputeLineOfPosition(lineStarts, comment.Pos()) != endingLine ||
			scanner.ComputeLineOfPosition(lineStarts, comment.End()) != endingLine ||
			!onlyHorizontalWhitespace(text, end, comment.Pos()) {
			break
		}
		end = comment.End()
	}
	return end
}

func onlyHorizontalWhitespace(text string, start, end int) bool {
	if start < 0 || start > end || end > len(text) {
		return false
	}
	for _, char := range text[start:end] {
		if char != ' ' && char != '\t' && char != '\r' {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Module classification
// ---------------------------------------------------------------------------

var (
	scopedModulePattern = regexp.MustCompile(`^@[^/]+/?[^/]+`)
	bareModulePattern   = regexp.MustCompile(`^\w`)
)

type importClassifier struct {
	ctx               rule.RuleContext
	settings          *import_utils.ModuleSettings
	packagePath       string
	packagePathLoaded bool
	caseSensitive     bool
}

func newImportClassifier(ctx rule.RuleContext) *importClassifier {
	classifier := &importClassifier{
		ctx:           ctx,
		settings:      import_utils.SettingsFor(ctx),
		caseSensitive: true,
	}
	if sourceProgram := ctx.Program(); sourceProgram != nil && sourceProgram.FS() != nil {
		classifier.caseSensitive = sourceProgram.FS().UseCaseSensitiveFileNames()
	}
	return classifier
}

// classify maps one written specifier to import/order's public group names.
// Resolution comes from the shared Program; the remaining tests are cheap
// lexical checks over the written module name.
func (classifier *importClassifier) classify(name string, specifier *ast.Node) string {
	if classifier.settings.IsInternalSpecifier(name) {
		return "internal"
	}
	if tspath.IsRootedDiskPath(name) {
		return "absolute"
	}

	// Relative spellings have a fixed public group and normally need no module
	// resolution. A deliberately configured core-module name is the exception:
	// a successful resolution suppresses builtin classification.
	builtinCandidate := isBuiltinModuleName(name, classifier.ctx.Settings)
	if !builtinCandidate {
		if relativeType, ok := relativeImportType(name); ok {
			return relativeType
		}
	}

	resolvedPath := ""
	if sourceProgram := classifier.ctx.Program(); sourceProgram != nil && specifier != nil {
		resolvedPath, _, _ = sourceProgram.ResolveModule(classifier.ctx.SourceFile, specifier)
	}
	if builtinCandidate && resolvedPath == "" {
		return "builtin"
	}
	if relativeType, ok := relativeImportType(name); ok {
		return relativeType
	}
	packagePath := classifier.contextPackagePath()
	if classifier.settings.IsExternalPathFromPackage(packagePath, resolvedPath, classifier.caseSensitive) {
		return "external"
	}
	if resolvedPath != "" {
		return "internal"
	}
	if isExternalLookingName(name) {
		return "external"
	}
	return "unknown"
}

func relativeImportType(name string) (string, bool) {
	if name == ".." || strings.HasPrefix(name, "../") || strings.HasPrefix(name, `..\`) {
		return "parent", true
	}
	switch name {
	case ".", "./", "./index", "./index.js":
		return "index", true
	}
	if strings.HasPrefix(name, "./") || strings.HasPrefix(name, `.\`) {
		return "sibling", true
	}
	return "", false
}

func (classifier *importClassifier) contextPackagePath() string {
	if classifier.packagePathLoaded {
		return classifier.packagePath
	}
	classifier.packagePathLoaded = true
	start := ""
	if classifier.ctx.SourceFile != nil {
		start = tspath.GetDirectoryPath(classifier.ctx.SourceFile.FileName())
	}
	if sourceProgram := classifier.ctx.Program(); sourceProgram != nil && start != "" {
		if packageDir := sourceProgram.NearestPackageJSONDirectory(start); packageDir != "" {
			classifier.packagePath = tspath.NormalizePath(packageDir)
			return classifier.packagePath
		}
	}
	if cwd := classifier.ctx.ProcessCurrentDirectory(); cwd != "" {
		classifier.packagePath = tspath.NormalizePath(cwd)
		return classifier.packagePath
	}
	if sourceProgram := classifier.ctx.Program(); sourceProgram != nil && sourceProgram.CurrentDirectory() != "" {
		classifier.packagePath = tspath.NormalizePath(sourceProgram.CurrentDirectory())
		return classifier.packagePath
	}
	classifier.packagePath = tspath.NormalizePath(start)
	return classifier.packagePath
}

func isBuiltinModuleName(name string, settings map[string]interface{}) bool {
	if name == "" {
		return false
	}
	base := baseModule(name)
	if core.NodeCoreModules()[base] {
		return true
	}
	if settings == nil {
		return false
	}
	switch extras := settings["import/core-modules"].(type) {
	case []string:
		return slices.Contains(extras, base)
	case []any:
		for _, extra := range extras {
			if value, ok := extra.(string); ok && value == base {
				return true
			}
		}
	}
	return false
}

func baseModule(name string) string {
	if scopedModulePattern.MatchString(name) {
		parts := strings.Split(name, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
		return parts[0] + "/undefined"
	}
	if slash := strings.IndexByte(name, '/'); slash >= 0 {
		return name[:slash]
	}
	return name
}

func isExternalLookingName(name string) bool {
	return name != "" && (bareModulePattern.MatchString(name) || scopedModulePattern.MatchString(name))
}

func computeRank(entry *importEntry, opts options) float64 {
	r := opts.ranks
	isTypeOnly := entry.importKind == "type"

	var impType string
	if entry.typ == "import:object" {
		impType = "object"
	} else if isTypeOnly && opts.typeGroupInGroups && !opts.isSortingTypesGroup {
		impType = "type"
	} else {
		impType = entry.classifyType
		// "absolute" is intentionally not configurable as a group. An absolute
		// specifier therefore has no rank and is ignored by whole-item ordering.
	}

	excluded := opts.pathGroupsExcludedImportTypes[impType]
	excludedFromPathRank := isTypeOnly && opts.typeGroupInGroups && opts.pathGroupsExcludedImportTypes["type"]

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

// reverseRanks duplicates the slice and negates every rank, making the normal
// forward scan equivalent to scanning the original sequence from the end.
func reverseRanks(in []*importEntry) []*importEntry {
	out := make([]*importEntry, len(in))
	for i, v := range in {
		dup := *v
		dup.rank = -v.rank
		out[len(in)-1-i] = &dup
	}
	return out
}

// countReverseOutOfOrder counts what findOutOfOrder(reverseRanks(imports))
// would return without allocating the reversed entries. The actual reversed
// slice is needed only when that direction produces fewer diagnostics.
func countReverseOutOfOrder(imports []*importEntry) int {
	if len(imports) == 0 {
		return 0
	}
	minSeen := imports[len(imports)-1].rank
	count := 0
	for i := len(imports) - 1; i >= 0; i-- {
		rank := imports[i].rank
		if rank > minSeen {
			count++
		}
		if rank < minSeen {
			minSeen = rank
		}
	}
	return count
}

func makeOutOfOrderReports(reports *reportQueue, imported []*importEntry, source *sourceInfo) {
	out := findOutOfOrder(imported)
	if len(out) == 0 {
		return
	}
	if countReverseOutOfOrder(imported) < len(out) {
		rev := reverseRanks(imported)
		revOut := findOutOfOrder(rev)
		// Use the reversed list (with negated ranks) as the search space — the
		// `found = imp.rank > X` predicate works against negated ranks too.
		// The entries in `rev` carry the same node identity as the originals,
		// so report positions still come from the original AST.
		reportOutOfOrder(reports, rev, revOut, "after", source, true)
		return
	}
	reportOutOfOrder(reports, imported, out, "before", source, false)
}

func reportOutOfOrder(reports *reportQueue, imported []*importEntry, outOfOrder []*importEntry, order string, source *sourceInfo, restoreSourceOrder bool) {
	ordered := outOfOrder
	if restoreSourceOrder && len(outOfOrder) > 1 {
		ordered = append([]*importEntry(nil), outOfOrder...)
		sort.SliceStable(ordered, func(i, j int) bool {
			return ast.CompareNodePositions(ordered[i].node, ordered[j].node) < 0
		})
	}
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
		makeOutOfOrderReport(reports, found, imp, order, source)
	}
}

func makeOutOfOrderReport(reports *reportQueue, first, second *importEntry, order string, source *sourceInfo) {
	disambiguateDisplayNames(first, second)
	firstDisplay := first.displayName
	secondDisplay := second.displayName

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

	reports.add(second.node, msg, func() []rule.RuleFix {
		return buildSwapFix(first, second, order, source)
	})
}

// disambiguateDisplayNames intentionally mutates the entries. When several
// diagnostics share one entry, an alias appended for the first report remains
// visible to later reports.
func disambiguateDisplayNames(first, second *importEntry) {
	if first == nil || second == nil || first.displayName != second.displayName {
		return
	}
	if first.alias != "" {
		first.displayName += " as " + first.alias
	}
	if second.alias != "" {
		second.displayName += " as " + second.alias
	}
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

// buildSwapFix produces a single replacement that swaps first and second in
// source order while retaining their attached same-line comments.
//
// Skips when:
//   - `canReorder` returns false (an unrelated statement sits between them)
//   - the two entries are already in different scopes (defensive)
//   - any of the line bounds are unresolvable
func buildSwapFix(first, second *importEntry, order string, source *sourceInfo) []rule.RuleFix {
	if first == nil || second == nil {
		return nil
	}
	firstRoot := first.statement
	secondRoot := second.statement
	if firstRoot == nil || secondRoot == nil {
		return nil
	}
	isExports := first.typ == "export" && second.typ == "export"
	if !isExports && !canReorder(firstRoot, secondRoot) {
		return nil
	}

	lineStarts := source.lines()
	comments := source.allComments()
	firstStart, firstEnd := statementLineRange(source.text, firstRoot, lineStarts, comments)
	secondStart, secondEnd := statementLineRange(source.text, secondRoot, lineStarts, comments)
	if firstStart < 0 || firstEnd < 0 || secondStart < 0 || secondEnd < 0 {
		return nil
	}

	// Source-order invariant:
	//   - "before": firstRoot appears earlier than secondRoot in source.
	//   - "after":  secondRoot appears earlier than firstRoot in source.
	// If either direction is violated, the caller paired entries wrong.
	if order == "before" && ast.CompareNodePositions(firstRoot, secondRoot) > 0 {
		return nil
	}
	if order == "after" && ast.CompareNodePositions(secondRoot, firstRoot) > 0 {
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
	replacement := source.text[secondEnd:firstEnd] + newCode
	return []rule.RuleFix{{
		Range: core.NewTextRange(secondStart, firstEnd),
		Text:  replacement,
	}}
}

// canReorder returns true when every statement between first and second is a
// side-effect-free module binding the fixer may safely cross. CommonJS export
// sorting has its own source-order contract and bypasses this check.
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

// canCrossWhileReorder accepts non-side-effect imports, external import-equals
// declarations, and the narrow require declarations whose evaluation can move.
func canCrossWhileReorder(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindImportDeclaration:
		return hasImportBindings(node.AsImportDeclaration())
	case ast.KindImportEqualsDeclaration:
		return ast.IsExternalModuleImportEqualsDeclaration(node)
	case ast.KindVariableStatement:
		return isSupportedRequireVarStatement(node)
	}
	return false
}

func isSupportedRequireVarStatement(node *ast.Node) bool {
	// Exported require declarations are not ordering candidates and therefore
	// cannot become invisible barriers to a fix.
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
		return false
	}
	vs := node.AsVariableStatement()
	dl := vs.DeclarationList.AsVariableDeclarationList()
	if dl == nil || dl.Declarations == nil || len(dl.Declarations.Nodes) != 1 {
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
	if requireCallWithLiteralArgument(skipRequireTransparentExpressions(d.Initializer)) != nil {
		return true
	}
	initializer := skipRequireTransparentExpressions(d.Initializer)
	if initializer == nil || initializer.Kind != ast.KindCallExpression || ast.IsOptionalChain(initializer) {
		return false
	}
	callee := skipRequireTransparentExpressions(initializer.AsCallExpression().Expression)
	if callee == nil || ast.IsOptionalChain(callee) {
		return false
	}
	if !ast.IsAccessExpression(callee) {
		return false
	}
	return requireCallWithLiteralArgument(skipRequireTransparentExpressions(utils.AccessExpressionObject(callee))) != nil
}

// skipRequireTransparentExpressions removes parentheses and the JSDoc type
// assertion wrapper tsgo retains in JavaScript files. ESTree represents the
// same JSDoc-authored expression without an assertion node. Explicit
// TypeScript `as`, `satisfies`, and non-null wrappers remain opaque, matching
// eslint-plugin-import's traversal.
func skipRequireTransparentExpressions(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		isJSDocAssertion := ast.IsJSDocTypeAssertion(node)
		parenthesized := node.AsParenthesizedExpression()
		if parenthesized == nil {
			return node
		}
		node = parenthesized.Expression
		if isJSDocAssertion && node != nil && node.Kind == ast.KindAsExpression {
			node = node.AsAsExpression().Expression
		}
	}
	return node
}

// requireCall starts with tsgo's own predicate and only adds the two AST-shape
// adjustments this rule needs: parentheses are transparent and optional calls
// are not sortable because moving them can change short-circuiting behavior.
func requireCall(node *ast.Node) *ast.CallExpression {
	node = skipRequireTransparentExpressions(node)
	if node == nil || !ast.IsCallExpression(node) || ast.IsOptionalChain(node) {
		return nil
	}
	if ast.IsRequireCall(node, false) {
		return node.AsCallExpression()
	}
	call := node.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return nil
	}
	callee := skipRequireTransparentExpressions(call.Expression)
	if !ast.IsIdentifier(callee) || callee.Text() != "require" || ast.IsOptionalChain(callee) {
		return nil
	}
	return call
}

func requireCallWithStringLiteralArgument(node *ast.Node) *ast.CallExpression {
	call := requireCall(node)
	if call == nil {
		return nil
	}
	argument := skipRequireTransparentExpressions(call.Arguments.Nodes[0])
	if argument == nil || argument.Kind != ast.KindStringLiteral {
		return nil
	}
	return call
}

func requireCallWithLiteralArgument(node *ast.Node) *ast.CallExpression {
	call := requireCall(node)
	if call == nil {
		return nil
	}
	argument := skipRequireTransparentExpressions(call.Arguments.Nodes[0])
	if argument == nil {
		return nil
	}
	switch argument.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return call
	default:
		return nil
	}
}

func findStaticRequireCall(node *ast.Node) *ast.CallExpression {
	for current := skipRequireTransparentExpressions(node); current != nil && !ast.IsOptionalChain(current); {
		if call := requireCallWithStringLiteralArgument(current); call != nil {
			return call
		}
		switch current.Kind {
		case ast.KindCallExpression:
			current = skipRequireTransparentExpressions(current.AsCallExpression().Expression)
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			current = skipRequireTransparentExpressions(utils.AccessExpressionObject(current))
		default:
			return nil
		}
	}
	return nil
}

// identifierAccessProperty returns the property expression only when it is an
// identifier. This distinction matters for CommonJS compatibility:
// `module[exports]` participates, while `module["exports"]` does not.
func identifierAccessProperty(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	var property *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property = node.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		property = node.AsElementAccessExpression().ArgumentExpression
	default:
		return nil
	}
	property = ast.SkipParentheses(property)
	if !ast.IsIdentifier(property) {
		return nil
	}
	return property
}

// identifierPropertyName unwraps tsgo's ComputedPropertyName node but keeps
// literal and expression keys out of named sorting.
func identifierPropertyName(name *ast.Node) *ast.Node {
	if name != nil && name.Kind == ast.KindComputedPropertyName {
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	if !ast.IsIdentifier(name) {
		return nil
	}
	return name
}

// movableStatement returns the smallest statement whose parent is an actual
// reorderable statement list. Case/default clauses are excluded: moving one
// clause statement independently can cross case labels, so the enclosing
// switch statement is the safe unit. Unbraced if/loop bodies similarly rise to
// their enclosing statement, while nodes in ordinary blocks stay local.
func movableStatement(statement *ast.Node) *ast.Node {
	for current := statement; current != nil; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			return current
		}
		if parent.CanHaveStatements() && parent.Kind != ast.KindCaseClause && parent.Kind != ast.KindDefaultClause {
			return current
		}
	}
	return statement
}

// ---------------------------------------------------------------------------
// Newlines-between report
// ---------------------------------------------------------------------------

func makeNewlinesBetweenReport(reports *reportQueue, imported []*importEntry, opts options, source *sourceInfo) {
	if len(imported) < 2 {
		return
	}
	sourceText := source.text
	lineStarts := source.lines()

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
			if ecmascript.IsBlank(sourceText[lineStart:lineEnd]) {
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
					previous := prev
					reports.add(previous.node,
						rule.RuleMessage{
							Id:          "groupNewline",
							Description: "There should be at least one empty line between import groups",
						},
						func() []rule.RuleFix {
							return []rule.RuleFix{fixInsertNewlineAfter(previous.statement, source)}
						},
					)
				}
			} else if empty > 0 && shouldAssertNoNLWithin {
				if isSameGroup {
					alreadyReported = true
					reportRemoveBlankLine(reports, prev, cur, source, "withinGroupNewline",
						"There should be no empty line within import group")
				}
			}
		} else if empty > 0 && shouldAssertNoNLBetween {
			alreadyReported = true
			reportRemoveBlankLine(reports, prev, cur, source, "groupNewline",
				"There should be no empty line between import groups")
		}

		if !alreadyReported && opts.consolidating {
			if empty == 0 && cur.isMultiline {
				previous := prev
				reports.add(previous.node,
					rule.RuleMessage{
						Id:          "consolidate",
						Description: "There should be at least one empty line between this import and the multi-line import that follows it",
					},
					func() []rule.RuleFix {
						return []rule.RuleFix{fixInsertNewlineAfter(previous.statement, source)}
					},
				)
			} else if empty == 0 && prev.isMultiline {
				previous := prev
				reports.add(previous.node,
					rule.RuleMessage{
						Id:          "consolidate",
						Description: "There should be at least one empty line between this multi-line import and the import that follows it",
					},
					func() []rule.RuleFix {
						return []rule.RuleFix{fixInsertNewlineAfter(previous.statement, source)}
					},
				)
			} else if empty > 0 && !prev.isMultiline && !cur.isMultiline && isSameGroup {
				reportRemoveBlankLine(reports, prev, cur, source, "consolidate",
					"There should be no empty lines between this single-line import and the single-line import that follows it")
			}
		}

		prev = cur
	}
}

func reportRemoveBlankLine(reports *reportQueue, prev, cur *importEntry, source *sourceInfo, msgId, msgText string) {
	msg := rule.RuleMessage{Id: msgId, Description: msgText}
	reports.add(prev.node, msg, func() []rule.RuleFix {
		fix := fixRemoveBlankLineBetween(source, prev.statement, cur.statement)
		if fix == nil {
			return nil
		}
		return []rule.RuleFix{*fix}
	})
}

func fixInsertNewlineAfter(node *ast.Node, source *sourceInfo) rule.RuleFix {
	end := sameLineEndWithComments(source.text, node, source.lines(), source.allComments())
	return rule.RuleFix{
		Range: core.NewTextRange(end, end),
		Text:  "\n",
	}
}

func fixRemoveBlankLineBetween(source *sourceInfo, prev, cur *ast.Node) *rule.RuleFix {
	if source == nil || prev == nil || cur == nil {
		return nil
	}
	lineStarts := source.lines()
	comments := source.allComments()
	removeFrom := endOfLineWithComments(source.text, prev, lineStarts, comments)
	removeTo := startOfLineWithComments(source.text, cur, lineStarts, comments)
	if removeFrom < 0 || removeTo < removeFrom || removeTo > len(source.text) ||
		!ecmascript.IsBlank(source.text[removeFrom:removeTo]) {
		return nil
	}
	f := rule.RuleFix{Range: core.NewTextRange(removeFrom, removeTo), Text: ""}
	return &f
}

// ---------------------------------------------------------------------------
// Alphabetize
// ---------------------------------------------------------------------------

// mutateRanksToAlphabetize assigns stable ranks within each configured group.
// Paths compare segment by segment, followed by the optional import-kind key.
func mutateRanksToAlphabetize(imported []*importEntry, opts alphabetizeOptions) {
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

	groupSizes := make(map[float64]int)
	for _, entry := range imported {
		groupSizes[entry.rank]++
	}
	groups := make(map[float64][]alphabetizeEntry, len(groupSizes))
	for rank, size := range groupSizes {
		groups[rank] = make([]alphabetizeEntry, 0, size)
	}
	for _, entry := range imported {
		value := entry.value
		if opts.caseInsensitive {
			value = ecmascript.StringToLowerCase(value)
		}
		kind := entry.importKind
		if kind == "" {
			kind = "value"
		}
		groups[entry.rank] = append(groups[entry.rank], alphabetizeEntry{
			entry:        entry,
			value:        value,
			segmentCount: strings.Count(value, "/") + 1,
			kind:         kind,
		})
	}

	keys := make([]float64, 0, len(groups))
	for rank := range groups {
		keys = append(keys, rank)
	}
	sort.Float64s(keys)

	cmp := makeAlphaComparator(multiplier, multiplierKind)
	for _, k := range keys {
		v8StableSortAlphabetized(groups[k], cmp)
	}

	newRanks := make(map[alphabetizedRankKey]float64, len(imported))
	var newRank float64
	for _, k := range keys {
		for _, prepared := range groups[k] {
			newRanks[makeAlphabetizedRankKey(prepared.entry)] = k + newRank
			newRank++
		}
	}
	for _, entry := range imported {
		if rank, ok := newRanks[makeAlphabetizedRankKey(entry)]; ok {
			entry.rank = rank
		}
	}
}

type alphabetizedRankKey struct {
	value      string
	importKind string
}

func makeAlphabetizedRankKey(entry *importEntry) alphabetizedRankKey {
	return alphabetizedRankKey{value: entry.value, importKind: entry.importKind}
}

type alphabetizeEntry struct {
	entry        *importEntry
	value        string
	segmentCount int
	kind         string
}

// makeAlphaComparator compares the precomputed module names segment by segment
// and then import kinds, returning a negative, zero, or positive result.
func makeAlphaComparator(multiplier, multiplierKind int) func(a, b alphabetizeEntry) int {
	return func(a, b alphabetizeEntry) int {
		result := compareAlphaValues(a, b)
		result *= multiplier
		if result == 0 && multiplierKind != 0 {
			result = multiplierKind * ecmascript.CompareStrings(a.kind, b.kind)
		}
		return result
	}
}

func compareAlphaValues(a, b alphabetizeEntry) int {
	if a.segmentCount == 1 && b.segmentCount == 1 {
		return ecmascript.CompareStrings(a.value, b.value)
	}

	aStart, bStart := 0, 0
	for segment := 0; ; segment++ {
		aEnd, bEnd := len(a.value), len(b.value)
		if offset := strings.IndexByte(a.value[aStart:], '/'); offset >= 0 {
			aEnd = aStart + offset
		}
		if offset := strings.IndexByte(b.value[bStart:], '/'); offset >= 0 {
			bEnd = bStart + offset
		}
		aPart, bPart := a.value[aStart:aEnd], b.value[bStart:bEnd]

		if segment == 0 && isRelativeRoot(aPart) && isRelativeRoot(bPart) && aPart != bPart {
			// The upstream comparator stops at different relative roots, then
			// still uses segment count as its fallback.
			return compareInts(a.segmentCount, b.segmentCount)
		}
		if result := ecmascript.CompareStrings(aPart, bPart); result != 0 {
			return result
		}

		aDone, bDone := aEnd == len(a.value), bEnd == len(b.value)
		if aDone || bDone {
			switch {
			case aDone && bDone:
				return 0
			case aDone:
				return -1
			default:
				return 1
			}
		}
		aStart, bStart = aEnd+1, bEnd+1
	}
}

func isRelativeRoot(part string) bool {
	return part == "." || part == ".."
}

func compareInts(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// ---------------------------------------------------------------------------
// Named ordering (named imports, exports, destructured require, cjs exports)
// ---------------------------------------------------------------------------

// makeNamedOrderReport sorts a single specifier list (e.g. the names inside
// `import { a, b } from 'x'`) and reports any out-of-order entries. Honours
// `named.types: 'mixed' | 'types-first' | 'types-last'`.
func makeNamedOrderReport(reports *reportQueue, named []*namedEntry, opts options, source *sourceInfo) {
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
	if countReverseOutOfOrder(imps) < len(out) {
		rev := reverseRanks(imps)
		revOut := findOutOfOrder(rev)
		reportNamedOutOfOrder(reports, rev, revOut, "after", source, true)
		return
	}
	reportNamedOutOfOrder(reports, imps, out, "before", source, false)
}

func reportNamedOutOfOrder(reports *reportQueue, all []*importEntry, outOfOrder []*importEntry, order string, source *sourceInfo, restoreSourceOrder bool) {
	ordered := outOfOrder
	if restoreSourceOrder && len(outOfOrder) > 1 {
		ordered = append([]*importEntry(nil), outOfOrder...)
		sort.SliceStable(ordered, func(i, j int) bool {
			return ordered[i].node.Pos() < ordered[j].node.Pos()
		})
	}
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
		disambiguateDisplayNames(first, second)
		firstDisplay := first.displayName
		secondDisplay := second.displayName
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
		reports.add(second.node, msg, func() []rule.RuleFix {
			return buildNamedSwapFix(source.file, source.text, first, second, order)
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
		trimmedTrivia := strings.TrimRightFunc(secondTrivia, ecmascript.IsWhiteSpaceOrLineTerminator)
		gapCode := sourceText[firstEnd : secondStart-1]
		whitespaces := secondTrivia[len(trimmedTrivia):]
		return []rule.RuleFix{{
			Range: core.NewTextRange(firstStart, secondEnd),
			Text:  secondCode + "," + trimmedTrivia + firstCode + firstTrivia + gapCode + whitespaces,
		}}
	case "after":
		if secondRange.Pos() > firstRange.Pos() || secondEnd+1 > firstStart {
			return nil
		}
		trimmedTrivia := strings.TrimRightFunc(firstTrivia, ecmascript.IsWhiteSpaceOrLineTerminator)
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
	// Prefer tokens from the parsed list that owns this member. A standalone
	// scanner starting at the beginning of a TSX file can lose JSX context and
	// mistake an earlier attribute's object braces for this list's delimiters.
	// Parser-backed parent tokens retain the correct context for named imports,
	// exports, binding patterns, and object literals.
	if node.Parent != nil {
		start, end := -1, -1
		for _, token := range utils.TokensOfNode(sourceFile, node.Parent) {
			isDelimiter := token.Kind == ast.KindCommaToken ||
				token.Kind == ast.KindOpenBraceToken ||
				token.Kind == ast.KindCloseBraceToken
			if !isDelimiter {
				continue
			}
			if token.End <= nodeRange.Pos() {
				start = token.End
				continue
			}
			if token.Start >= nodeRange.End() {
				end = token.Start
				break
			}
		}
		if start >= 0 && end >= start {
			return start, end, true
		}
	}

	// Separate CommonJS assignment statements are not members of a parsed
	// comma-delimited parent. Retain the source-scanner fallback for those.
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
		if token.Start >= search {
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
		if token.End <= search {
			break
		}
		search = token.End
	}
	return start, end, start >= 0 && end >= start
}

// ---------------------------------------------------------------------------
// Statement walking
// ---------------------------------------------------------------------------

// blockState holds the candidates belonging to one tsgo statement container.
// imports and exports are ordered against siblings; namedLists are independent
// specifier or object-binding lists collected while visiting that container.
type blockState struct {
	imports []*importEntry
	exports []*importEntry
	// namedLists collects deferred named-order reports — populated when we
	// see an import / export / require / module.exports = {} that has
	// candidate named children.
	namedLists [][]*namedEntry
}

// cjsScopeIndex maps identifier references to their lexical scope. The public
// cjsExports behavior checks only that scope's own declarations when deciding
// whether `module` or `exports` is shadowed; it does not walk parent scopes.
type cjsScopeIndex map[*ast.Node]*scope.Scope

func buildCJSScopeIndex(sourceFile *ast.SourceFile) cjsScopeIndex {
	manager := scope.Build(sourceFile, scope.Options{CollectReferences: true})
	index := make(cjsScopeIndex, len(manager.References))
	for _, reference := range manager.References {
		index[reference.Identifier] = reference.From
	}
	return index
}

func (bs *blockState) addImport(e *importEntry)    { bs.imports = append(bs.imports, e) }
func (bs *blockState) addNamed(list []*namedEntry) { bs.namedLists = append(bs.namedLists, list) }
func (bs *blockState) addExport(e *importEntry)    { bs.exports = append(bs.exports, e) }

type blockCollector struct {
	byContainer map[*ast.Node]*blockState
	ordered     []*blockState
}

func newBlockCollector() *blockCollector {
	return &blockCollector{byContainer: map[*ast.Node]*blockState{}}
}

func (collector *blockCollector) state(container *ast.Node) *blockState {
	if container == nil {
		return nil
	}
	if state := collector.byContainer[container]; state != nil {
		return state
	}
	state := &blockState{}
	collector.byContainer[container] = state
	collector.ordered = append(collector.ordered, state)
	return state
}

func (collector *blockCollector) statementState(node *ast.Node) *blockState {
	container := ast.FindAncestor(node.Parent, func(parent *ast.Node) bool {
		return parent.CanHaveStatements()
	})
	return collector.state(container)
}

func handleImportDeclaration(ctx rule.RuleContext, stmt *ast.Node, opts options, bs *blockState) {
	decl := stmt.AsImportDeclaration()
	if decl.ModuleSpecifier == nil || decl.ModuleSpecifier.Kind != ast.KindStringLiteral {
		return
	}
	// An empty `import {} from "x"` has a tsgo ImportClause but no bindings, so
	// it follows the same opt-in path as a side-effect import.
	if !hasImportBindings(decl) && !opts.warnOnUnassigned {
		return
	}
	value := decl.ModuleSpecifier.AsStringLiteral().Text
	kind := importKindOf(stmt)
	entry := &importEntry{
		node:        stmt,
		statement:   movableStatement(stmt),
		specifier:   decl.ModuleSpecifier,
		value:       value,
		displayName: value,
		typ:         "import",
		importKind:  kind,
		isMultiline: opts.consolidating && isMultiline(stmt, ctx.SourceFile),
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

func hasImportBindings(declaration *ast.ImportDeclaration) bool {
	if declaration == nil || declaration.ImportClause == nil {
		return false
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil {
		return false
	}
	if clause.Name() != nil {
		return true
	}
	if clause.NamedBindings == nil {
		return false
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		return true
	case ast.KindNamedImports:
		named := clause.NamedBindings.AsNamedImports()
		return named != nil && named.Elements != nil && len(named.Elements.Nodes) > 0
	default:
		return false
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
		propName = sortableSpecifierName(s.PropertyName)
	} else {
		propName = sortableSpecifierName(s.Name())
	}
	localName = sortableSpecifierName(s.Name())
	return propName, localName, true
}

// sortableSpecifierName returns the identifier value used by named-order
// diagnostics. String-named specifiers have no identifier name and are
// rendered as JavaScript's "undefined".
func sortableSpecifierName(node *ast.Node) string {
	if node != nil && node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text
	}
	return "undefined"
}

func handleImportEqualsDeclaration(ctx rule.RuleContext, stmt *ast.Node, opts options, bs *blockState) {
	if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
		return
	}
	eq := stmt.AsImportEqualsDeclaration()
	var value, displayName, typ, classify string
	var specifier *ast.Node
	if ast.IsExternalModuleImportEqualsDeclaration(stmt) {
		expr := ast.GetExternalModuleImportEqualsDeclarationExpression(stmt)
		if expr == nil || expr.Kind != ast.KindStringLiteral {
			return
		}
		value = expr.AsStringLiteral().Text
		displayName = value
		typ = "import"
		specifier = expr
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
		statement:    movableStatement(stmt),
		specifier:    specifier,
		value:        value,
		displayName:  displayName,
		typ:          typ,
		importKind:   kind,
		classifyType: classify,
		isMultiline:  opts.consolidating && isMultiline(stmt, ctx.SourceFile),
	})
}

func handleVariableStatement(ctx rule.RuleContext, stmt *ast.Node, opts options, bs *blockState) {
	if ast.HasSyntacticModifier(stmt, ast.ModifierFlagsExport) {
		return
	}
	vs := stmt.AsVariableStatement()
	dl := vs.DeclarationList.AsVariableDeclarationList()
	if dl == nil || dl.Declarations == nil {
		return
	}
	for _, declarationNode := range dl.Declarations.Nodes {
		d := declarationNode.AsVariableDeclaration()
		if d.Initializer == nil {
			continue
		}
		ce := findStaticRequireCall(d.Initializer)
		if ce == nil || ce.Arguments == nil || len(ce.Arguments.Nodes) != 1 {
			continue
		}
		arg := skipRequireTransparentExpressions(ce.Arguments.Nodes[0])
		if arg == nil || arg.Kind != ast.KindStringLiteral {
			continue
		}
		value := arg.AsStringLiteral().Text
		multilineNode := ce.AsNode()
		// A direct initializer spans the declaration for multiline grouping. In
		// a member/call chain, only the require expression itself is measured.
		if ast.SkipParentheses(d.Initializer) == ce.AsNode() {
			multilineNode = stmt
		}
		bs.addImport(&importEntry{
			node:        ce.AsNode(),
			statement:   movableStatement(stmt),
			specifier:   arg,
			value:       value,
			displayName: value,
			typ:         "require",
			isMultiline: opts.consolidating && isMultiline(multilineNode, ctx.SourceFile),
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
			// Rest and defaulted bindings do not provide a plain key/value pair;
			// leave the complete destructuring list untouched.
			return nil
		}
		if be.Name() == nil || be.Name().Kind != ast.KindIdentifier {
			return nil
		}
		var prop, local string
		if be.PropertyName != nil {
			property := identifierPropertyName(be.PropertyName)
			if property == nil {
				// Only identifier keys participate. Computed identifiers are
				// unwrapped; literal and expression keys are not sorted.
				return nil
			}
			prop = property.AsIdentifier().Text
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

func handleExportDeclaration(stmt *ast.Node, opts options, bs *blockState) {
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
		var prop string
		if s.PropertyName != nil {
			prop = sortableSpecifierName(s.PropertyName)
		} else {
			prop = sortableSpecifierName(s.Name())
		}
		alias := ""
		if s.PropertyName != nil && s.Name().Kind == ast.KindIdentifier {
			alias = s.Name().AsIdentifier().Text
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
// Only identifier-keyed and identifier-valued shorthand/longhand properties
// participate: `{ a, b }`, `{ a: aValue }`, and computed identifier keys.
// Literal keys, methods, spreads, defaults, or expression values leave the
// complete list untouched.
//
// Distinguishing the global `module.exports` from a local binding requires
// scope information. The rule's reference map resolves each receiver in its
// lexical scope; without that map, the structural global form is accepted.
func handleExpressionStatement(scopes cjsScopeIndex, stmt *ast.Node, opts options, collector *blockCollector) {
	if !opts.named.cjsExports {
		return
	}
	rawExpression := stmt.AsExpressionStatement().Expression
	if rawExpression == nil {
		return
	}
	expr := ast.SkipParentheses(rawExpression)
	// tsgo represents simple, compound, and logical assignments as binary
	// expressions. The rule accepts the whole assignment family, not only `=`.
	if expr == nil || !ast.IsAssignmentExpression(expr, false /* excludeCompoundAssignment */) {
		return
	}
	be := expr.AsBinaryExpression()
	if !isCJSExportsTarget(scopes, be.Left) {
		nameParts := getNamedCJSExportParts(scopes, be.Left)
		if len(nameParts) > 0 {
			state := collector.state(stmt.Parent)
			if state == nil {
				return
			}
			name := strings.Join(nameParts, ".")
			state.addExport(&importEntry{
				node:        expr,
				statement:   movableStatement(stmt),
				value:       name,
				displayName: name,
				typ:         "export",
				rank:        0,
			})
		}
		return
	}
	if be.Right == nil {
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
			property := identifierPropertyName(pa.Name())
			if property == nil || pa.Initializer == nil {
				return
			}
			keyText = property.AsIdentifier().Text
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
		if state := collector.state(stmt.Parent); state != nil {
			state.addNamed(out)
		}
	}
}

// handleNamedRequireDeclaration accepts direct object-destructuring requires
// in every lexical scope; whole-module require ordering remains Program-only.
func handleNamedRequireDeclaration(node *ast.Node, collector *blockCollector) {
	declaration := node.AsVariableDeclaration()
	if declaration.Initializer == nil || requireCallWithLiteralArgument(declaration.Initializer) == nil ||
		declaration.Name() == nil || declaration.Name().Kind != ast.KindObjectBindingPattern {
		return
	}
	named := collectNamedFromBindingPattern(declaration.Name())
	if len(named) > 1 {
		if state := collector.statementState(node); state != nil {
			state.addNamed(named)
		}
	}
}

// isCJSExportsTarget returns true when the LHS of an assignment refers to the
// global `module.exports` or `exports` identifier without local shadowing.
//
// Dot access and identifier element access (`module[exports]`) are accepted;
// string-literal element access is intentionally distinct. Shadowing is
// checked in the reference's current lexical scope.
func isCJSExportsTarget(scopes cjsScopeIndex, node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	// `exports` (bare identifier).
	if ast.IsExportsIdentifier(node) {
		return !isDeclaredInCurrentScope(scopes, node)
	}
	// `module.exports` / `module[exports]`.
	if !ast.IsAccessExpression(node) {
		return false
	}
	property := identifierAccessProperty(node)
	if property == nil || property.Text() != "exports" {
		return false
	}
	moduleIdent := ast.SkipParentheses(utils.AccessExpressionObject(node))
	if !ast.IsModuleIdentifier(moduleIdent) {
		return false
	}
	return !isDeclaredInCurrentScope(scopes, moduleIdent)
}

func getNamedCJSExportParts(scopes cjsScopeIndex, node *ast.Node) []string {
	if node == nil {
		return nil
	}
	root := ast.SkipParentheses(node)
	var parent *ast.Node
	var result []string
	for root != nil {
		if !ast.IsAccessExpression(root) {
			goto done
		}
		property := identifierAccessProperty(root)
		if property == nil {
			return nil
		}
		result = append([]string{property.AsIdentifier().Text}, result...)
		parent = root
		root = ast.SkipParentheses(utils.AccessExpressionObject(root))
	}

done:
	if isCJSExportsTarget(scopes, root) {
		return result
	}
	if isCJSExportsTarget(scopes, parent) && len(result) > 1 {
		return result[1:]
	}
	return nil
}

func isDeclaredInCurrentScope(scopes cjsScopeIndex, ident *ast.Node) bool {
	current := scopes[ident]
	if current == nil || ident == nil {
		return false
	}
	return len(current.Declarations(ident.Text())) > 0
}

// importKindOf returns "type" only for a whole-declaration type import.
// `import { type X } from 'mod'` remains a value import for whole-item ordering;
// its per-specifier kind is considered only by named sorting.
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
// `node.Pos()` includes leading trivia and can begin on the preceding line, so
// the comparison starts at the node's first non-trivia token.
//
// The shared IsSameLine helper uses the source file's ECMAScript line map, so
// LF, CRLF, U+2028, and U+2029 all follow the parser's own line semantics.
func isMultiline(node *ast.Node, sf *ast.SourceFile) bool {
	if sf == nil {
		return false
	}
	tokenStart := scanner.SkipTrivia(sf.Text(), node.Pos())
	return !utils.IsSameLine(sf, tokenStart, node.End())
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
		opts := parseOptions(rawOptions)
		sf := ctx.SourceFile
		if sf == nil {
			return rule.RuleListeners{}
		}

		// Surface a single file-level diagnostic for malformed `groups`; syntax
		// listeners do not run when the configuration cannot produce ranks.
		if opts.rankErr != nil {
			ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{
				Id:          "configError",
				Description: opts.rankErr.Error(),
			})
			return rule.RuleListeners{}
		}

		collector := newBlockCollector()
		source := &sourceInfo{file: sf, commentStore: ctx.Comments, text: sf.Text()}
		reports := &reportQueue{}

		// The linter already walks every node once. Collect directly from that
		// traversal so SourceFile, namespace ModuleBlock, and future tsgo
		// statement containers all follow the same path without a private AST
		// recursion.
		listeners := rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				if state := collector.statementState(node); state != nil {
					handleImportDeclaration(ctx, node, opts, state)
				}
			},
			ast.KindImportEqualsDeclaration: func(node *ast.Node) {
				if state := collector.statementState(node); state != nil {
					handleImportEqualsDeclaration(ctx, node, opts, state)
				}
			},
			ast.KindVariableStatement: func(node *ast.Node) {
				// Whole-module require ordering is Program-only. Named destructuring
				// is collected separately from VariableDeclaration in every scope.
				if node.Parent != nil && node.Parent.Kind == ast.KindSourceFile {
					if state := collector.state(node.Parent); state != nil {
						handleVariableStatement(ctx, node, opts, state)
					}
				}
			},
		}
		if opts.named.exports {
			listeners[ast.KindExportDeclaration] = func(node *ast.Node) {
				if state := collector.statementState(node); state != nil {
					handleExportDeclaration(node, opts, state)
				}
			}
		}
		if opts.named.require {
			listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
				handleNamedRequireDeclaration(node, collector)
			}
		}
		if opts.named.cjsExports {
			cjsScopes := buildCJSScopeIndex(sf)
			listeners[ast.KindExpressionStatement] = func(node *ast.Node) {
				handleExpressionStatement(cjsScopes, node, opts, collector)
			}
		}
		listeners[rule.ListenerOnExit(ast.KindEndOfFile)] = func(*ast.Node) {
			var classifier *importClassifier
			for _, state := range collector.ordered {
				if classifier == nil && len(state.imports) > 1 {
					classifier = newImportClassifier(ctx)
				}
				finalizeBlock(reports, state, opts, source, classifier)
			}
			reports.flush(ctx)
		}
		return listeners
	},
}

// finalizeBlock runs newline, alphabetize, whole-item, and named checks for one
// statement container in the rule's public diagnostic order.
func finalizeBlock(reports *reportQueue, bs *blockState, opts options, source *sourceInfo, classifier *importClassifier) {
	// A single module entry cannot participate in any whole-item diagnostic.
	// Delay both resolution and rank construction until a block has a pair.
	var imported []*importEntry
	if len(bs.imports) > 1 {
		imported = bs.imports[:0]
		for _, e := range bs.imports {
			if e.specifier != nil {
				e.classifyType = classifier.classify(e.value, e.specifier)
			}
			r := computeRank(e, opts)
			if r == -1 {
				continue
			}
			e.rank = r
			imported = append(imported, e)
		}
	}

	if opts.newlinesBetween != "ignore" || opts.newlinesBetweenTypes != "ignore" {
		makeNewlinesBetweenReport(reports, imported, opts, source)
	}

	if opts.alphabetize.order != "ignore" {
		mutateRanksToAlphabetize(imported, opts.alphabetize)
	}

	makeOutOfOrderReports(reports, imported, source)

	// ES lists are gathered from statement lists and CommonJS lists from the
	// linter's traversal. Restore source order before reporting mixed syntax.
	sort.SliceStable(bs.namedLists, func(i, j int) bool {
		return ast.CompareNodePositions(bs.namedLists[i][0].node, bs.namedLists[j][0].node) < 0
	})
	for _, list := range bs.namedLists {
		makeNamedOrderReport(reports, list, opts, source)
	}

	if opts.alphabetize.order != "ignore" && len(bs.exports) > 1 {
		mutateRanksToAlphabetize(bs.exports, opts.alphabetize)
		makeOutOfOrderReports(reports, bs.exports, source)
	}
}
