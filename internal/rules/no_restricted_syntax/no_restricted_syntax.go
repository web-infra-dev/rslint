package no_restricted_syntax

import (
	_ "embed"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_restricted_syntax.schema.json
var schemaJSON []byte

var estreeBuilderPool = sync.Pool{
	New: func() any {
		return new(estree.Builder)
	},
}

// NoRestrictedSyntaxRule is the rslint port of ESLint's `no-restricted-syntax`.
//
// Configuration mirrors ESLint: the rule accepts a list of selector strings or
// `{ selector, message? }` objects. Selectors are evaluated against a scoped
// read-only TSESTree projected by internal/languages/js/estree so every JS,
// JSX, and TypeScript ESTree node, field, and parent-child relation matches
// upstream ESLint + @typescript-eslint/parser semantics.
//
// https://eslint.org/docs/latest/rules/no-restricted-syntax
var NoRestrictedSyntaxRule = rule.Rule{
	Name:   "no-restricted-syntax",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		plan := cachedRulePlan(options)
		if plan.empty {
			return nil
		}
		if plan.eager {
			evaluateFile(ctx, plan)
			return nil
		}
		ran := false
		trigger := func(*ast.Node) {
			if ran {
				return
			}
			ran = true
			evaluateFile(ctx, plan)
		}
		listeners := make(rule.RuleListeners, len(plan.triggerKinds))
		for _, kind := range plan.triggerKinds {
			listeners[kind] = trigger
		}
		return listeners
	},
}

func evaluateFile(ctx rule.RuleContext, plan *rulePlan) {
	builder, ok := estreeBuilderPool.Get().(*estree.Builder)
	if !ok {
		builder = new(estree.Builder)
	}
	defer func() {
		builder.Reset()
		estreeBuilderPool.Put(builder)
	}()
	_ = builder.WithTree(ctx.SourceFile, estree.TreeOptions{
		SourceType: ctx.LanguageOptions.EffectiveSourceType(),
	}, func(tree *estree.Tree) {
		walkTree(tree, plan, &ctx)
	})
}

type walkFrame struct {
	id      estree.NodeID
	exiting bool
}

func walkTree(tree *estree.Tree, plan *rulePlan, ctx *rule.RuleContext) {
	root := tree.Root()
	if root == estree.NoNode {
		return
	}
	var memo *evalMemo
	if plan.hasMemo {
		memo = &evalMemo{}
	}
	var stackBuf [64]walkFrame
	var childBuf [32]estree.NodeID
	stack := append(stackBuf[:0], walkFrame{id: root})
	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		tid := tree.Type(frame.id)
		if frame.exiting {
			plan.exit[tid].evaluate(tree, frame.id, ctx, memo)
			continue
		}
		if bucket := plan.enter[tid]; bucket != nil {
			bucket.evaluate(tree, frame.id, ctx, memo)
		}
		if plan.exit[tid] != nil {
			stack = append(stack, walkFrame{id: frame.id, exiting: true})
		}
		children := tree.AppendChildren(frame.id, childBuf[:0])
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, walkFrame{id: children[i]})
		}
	}
}

// ruleEntry is a parsed option entry: the original selector string,
// the compiled selector tree, and the message (custom or default).
type ruleEntry struct {
	selector string
	compiled selector
	message  string
	exit     bool
}

func (e ruleEntry) formatMessage() string {
	return e.message
}

const maxCachedRulePlans = 64

type rulePlan struct {
	enter        []*ruleBucket
	exit         []*ruleBucket
	triggerKinds []ast.Kind
	eager        bool
	empty        bool
	hasMemo      bool
}

var emptyRulePlan = &rulePlan{empty: true}

var rulePlanCache = struct {
	sync.RWMutex
	entries map[string]*rulePlan
	order   []string
}{
	entries: make(map[string]*rulePlan),
	order:   make([]string, 0, maxCachedRulePlans),
}

func optionsCacheKey(options []any) (string, bool) {
	var sb strings.Builder
	for _, item := range options {
		switch x := item.(type) {
		case string:
			sb.WriteByte('S')
			sb.WriteString(strconv.Itoa(len(x)))
			sb.WriteByte(':')
			sb.WriteString(x)
		case map[string]any:
			sel, _ := x["selector"].(string)
			msg, _ := x["message"].(string)
			sb.WriteByte('O')
			sb.WriteString(strconv.Itoa(len(sel)))
			sb.WriteByte(':')
			sb.WriteString(sel)
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(len(msg)))
			sb.WriteByte(':')
			sb.WriteString(msg)
		default:
			return "", false
		}
	}
	return sb.String(), true
}

func cachedRulePlan(options []any) *rulePlan {
	if len(options) == 0 {
		return emptyRulePlan
	}
	key, cacheable := optionsCacheKey(options)
	if !cacheable {
		return buildRulePlan(options)
	}

	rulePlanCache.RLock()
	plan := rulePlanCache.entries[key]
	rulePlanCache.RUnlock()
	if plan != nil {
		return plan
	}

	plan = buildRulePlan(options)
	rulePlanCache.Lock()
	if cached := rulePlanCache.entries[key]; cached != nil {
		rulePlanCache.Unlock()
		return cached
	}
	if len(rulePlanCache.order) == maxCachedRulePlans {
		oldest := rulePlanCache.order[0]
		delete(rulePlanCache.entries, oldest)
		copy(rulePlanCache.order, rulePlanCache.order[1:])
		rulePlanCache.order[len(rulePlanCache.order)-1] = key
	} else {
		rulePlanCache.order = append(rulePlanCache.order, key)
	}
	rulePlanCache.entries[key] = plan
	rulePlanCache.Unlock()
	return plan
}

type typeSet struct {
	all   bool
	types map[estree.TypeID]struct{}
}

func newTypeSet() *typeSet {
	return &typeSet{types: make(map[estree.TypeID]struct{})}
}

func (s *typeSet) markAll() {
	s.all = true
	s.types = nil
}

func (s *typeSet) add(id estree.TypeID) {
	if s.all || id == estree.UnknownType {
		return
	}
	s.types[id] = struct{}{}
}

func (s *typeSet) unionWith(other *typeSet) {
	if s.all {
		return
	}
	if other.all {
		s.markAll()
		return
	}
	for id := range other.types {
		s.types[id] = struct{}{}
	}
}

func (s *typeSet) intersectWith(other *typeSet) {
	if other.all {
		return
	}
	if s.all {
		s.all = false
		s.types = make(map[estree.TypeID]struct{}, len(other.types))
		for id := range other.types {
			s.types[id] = struct{}{}
		}
		return
	}
	for id := range s.types {
		if _, ok := other.types[id]; !ok {
			delete(s.types, id)
		}
	}
}

// candidateTypes computes the exact ESLint nodeTypes dispatch set for sel.
// Selectors whose rightmost host is constrained to specific ESTree type names
// receive only those TypeIDs; unconstrained selectors (`*`, attributes, `:not`,
// `:has`, `:nth-child`, node classes) receive all=true.
func candidateTypes(sel selector) *typeSet {
	switch v := sel.(type) {
	case subjectSelector:
		return candidateTypes(v.Inner)
	case identifierSelector:
		s := newTypeSet()
		if v.Wildcard {
			s.markAll()
		} else if v.ExactTypeID != estree.UnknownType {
			s.add(v.ExactTypeID)
		}
		return s
	case compoundSelector:
		left := candidateTypes(v.Left)
		left.intersectWith(candidateTypes(v.Right))
		return left
	case classSelector:
		return candidateTypes(v.Inner)
	case attrSelector:
		return candidateTypes(v.Inner)
	case combinatorSelector:
		return candidateTypes(v.Right)
	case pseudoSelector:
		switch v.Name {
		case "is", "matches":
			s := newTypeSet()
			for _, arg := range v.Args {
				s.unionWith(candidateTypes(arg))
			}
			return s
		case "function":
			if v.ExactFunction {
				return classTypeSet(isFunctionType)
			}
			s := newTypeSet()
			s.markAll()
			return s
		default:
			// Match ESLint's analyzeSelector: only `:function` narrows
			// candidate node types; `:statement`, `:declaration`, `:expression`,
			// `:pattern`, `:not`, `:has`, and `:nth-*` must visit every node
			// type so reverse sibling subjects like `!A ~ :declaration` match
			// both sides identically to upstream ESLint.
			s := newTypeSet()
			s.markAll()
			return s
		}
	case combinedPseudo:
		left := candidateTypes(v.Inner)
		right := candidateTypes(v.Pseudo)
		left.intersectWith(right)
		return left
	case unionSelector:
		s := newTypeSet()
		for _, child := range v.Selectors {
			s.unionWith(candidateTypes(child))
		}
		return s
	}
	return newTypeSet()
}

func classTypeSet(table []bool) *typeSet {
	s := newTypeSet()
	for i, ok := range table {
		if ok {
			s.add(estree.TypeID(i))
		}
	}
	return s
}

func assignMemoIDs(sel selector, next *uint32) selector {
	switch v := sel.(type) {
	case subjectSelector:
		v.Inner = assignMemoIDs(v.Inner, next)
		return v
	case compoundSelector:
		v.Left = assignMemoIDs(v.Left, next)
		v.Right = assignMemoIDs(v.Right, next)
		return v
	case classSelector:
		v.Inner = assignMemoIDs(v.Inner, next)
		return v
	case attrSelector:
		v.Inner = assignMemoIDs(v.Inner, next)
		return v
	case combinatorSelector:
		v.Left = assignMemoIDs(v.Left, next)
		v.Right = assignMemoIDs(v.Right, next)
		switch v.Kind {
		case combDescendant:
			*next++
			v.memoID = *next
		case combSibling:
			*next++
			v.memoID = *next
			*next++ // reserve memoID+1 for reverse sibling direction
		}
		return v
	case relativeSelector:
		v.Inner = assignMemoIDs(v.Inner, next)
		return v
	case combinedPseudo:
		v.Inner = assignMemoIDs(v.Inner, next)
		v.Pseudo = assignPseudoMemoIDs(v.Pseudo, next)
		return v
	case pseudoSelector:
		return assignPseudoMemoIDs(v, next)
	case unionSelector:
		selectors := make([]selector, len(v.Selectors))
		for i, child := range v.Selectors {
			selectors[i] = assignMemoIDs(child, next)
		}
		v.Selectors = selectors
		return v
	}
	return sel
}

func assignPseudoMemoIDs(v pseudoSelector, next *uint32) pseudoSelector {
	if len(v.Args) > 0 {
		args := make([]selector, len(v.Args))
		for i, arg := range v.Args {
			args[i] = assignMemoIDs(arg, next)
		}
		v.Args = args
	}
	if v.Name == "has" {
		*next++
		v.memoID = *next
	}
	return v
}

func buildRulePlan(options []any) *rulePlan {
	entries := parseRuleOptions(options)
	entries = deduplicateRuleEntries(entries)
	sortRuleEntries(entries)
	if len(entries) == 0 {
		return emptyRulePlan
	}

	var memoCounter uint32
	for i := range entries {
		entries[i].compiled = assignMemoIDs(entries[i].compiled, &memoCounter)
	}

	typeCount := estree.TypeCount()
	enterPerType := make([][]ruleEntry, typeCount)
	exitPerType := make([][]ruleEntry, typeCount)
	hasAny := false
	eager := false
	seenTrigger := make(map[ast.Kind]struct{})
	var triggerKinds []ast.Kind

	addTrigger := func(tid estree.TypeID) {
		if eager {
			return
		}
		kinds, ok := estree.TriggerKinds(tid)
		if !ok {
			eager = true
			triggerKinds = nil
			return
		}
		for _, k := range kinds {
			if _, exists := seenTrigger[k]; !exists {
				seenTrigger[k] = struct{}{}
				triggerKinds = append(triggerKinds, k)
			}
		}
	}

	for _, entry := range entries {
		targets := candidateTypes(entry.compiled)
		perType := enterPerType
		if entry.exit {
			perType = exitPerType
		}
		if targets.all {
			hasAny = true
			eager = true
			triggerKinds = nil
			for i := range typeCount {
				perType[i] = append(perType[i], entry)
			}
			continue
		}
		for tid := range targets.types {
			hasAny = true
			perType[tid] = append(perType[tid], entry)
			addTrigger(tid)
		}
	}

	if !hasAny {
		return emptyRulePlan
	}

	plan := &rulePlan{
		enter:        make([]*ruleBucket, typeCount),
		exit:         make([]*ruleBucket, typeCount),
		triggerKinds: triggerKinds,
		eager:        eager,
		hasMemo:      memoCounter > 0,
	}
	for i := range typeCount {
		if len(enterPerType[i]) > 0 {
			bucket := buildRuleBucket(enterPerType[i])
			plan.enter[i] = &bucket
		}
		if len(exitPerType[i]) > 0 {
			bucket := buildRuleBucket(exitPerType[i])
			plan.exit[i] = &bucket
		}
	}
	return plan
}

type ruleBucket struct {
	entries  []ruleEntry
	dispatch *stringAttrDispatch
}

const maxDispatchDepth = 4

type stringAttrDispatch struct {
	path            []pathSegment
	singleValue     string
	singleHits      []int
	byValue         map[string][]int
	fallback        []int
	next            *stringAttrDispatch
	hasNonStringKey bool
}

type dispatchAttr struct {
	path  []pathSegment
	key   string
	value string
}

type dispatchStat struct {
	path   []pathSegment
	count  int
	values map[string]struct{}
}

func buildRuleBucket(entries []ruleEntry) ruleBucket {
	bucket := ruleBucket{entries: entries}
	if len(entries) < 2 {
		return bucket
	}
	indices := make([]int, len(entries))
	for i := range indices {
		indices[i] = i
	}
	bucket.dispatch = buildStringDispatch(entries, indices, 0)
	return bucket
}

func buildStringDispatch(entries []ruleEntry, indices []int, depth int) *stringAttrDispatch {
	if len(indices) < 2 || depth >= maxDispatchDepth {
		return nil
	}

	stats := make(map[string]*dispatchStat)
	for _, entryIdx := range indices {
		attrs := collectDispatchAttrs(entries[entryIdx].compiled, nil)
		for index, attr := range attrs {
			duplicate := false
			for previous := range index {
				if attrs[previous].key == attr.key {
					duplicate = true
					break
				}
			}
			if duplicate {
				continue
			}
			stat := stats[attr.key]
			if stat == nil {
				stat = &dispatchStat{
					path:   attr.path,
					values: make(map[string]struct{}),
				}
				stats[attr.key] = stat
			}
			stat.count++
			stat.values[attr.value] = struct{}{}
		}
	}

	bestKey := ""
	bestCount, bestValues := 0, 0
	for key, stat := range stats {
		valueCount := len(stat.values)
		if stat.count > bestCount ||
			(stat.count == bestCount && valueCount > bestValues) ||
			(stat.count == bestCount && valueCount == bestValues && (bestKey == "" || key < bestKey)) {
			bestKey = key
			bestCount = stat.count
			bestValues = valueCount
		}
	}
	if bestCount < 2 {
		return nil
	}

	dispatch := &stringAttrDispatch{
		path:     stats[bestKey].path,
		fallback: make([]int, 0, len(indices)-bestCount),
	}
	for value := range stats[bestKey].values {
		if couldMatchNonString(value) {
			dispatch.hasNonStringKey = true
			break
		}
	}
	if bestValues == 1 {
		for value := range stats[bestKey].values {
			dispatch.singleValue = value
		}
		dispatch.singleHits = make([]int, 0, bestCount)
	} else {
		dispatch.byValue = make(map[string][]int, bestValues)
	}

	for _, entryIdx := range indices {
		value, ok := dispatchValueForPath(entries[entryIdx].compiled, bestKey)
		if !ok {
			dispatch.fallback = append(dispatch.fallback, entryIdx)
			continue
		}
		if dispatch.byValue == nil {
			dispatch.singleHits = append(dispatch.singleHits, entryIdx)
		} else {
			dispatch.byValue[value] = append(dispatch.byValue[value], entryIdx)
		}
	}
	if len(dispatch.fallback) >= 2 {
		if next := buildStringDispatch(entries, dispatch.fallback, depth+1); next != nil {
			dispatch.next = next
			dispatch.fallback = nil
		}
	}
	return dispatch
}

// couldMatchNonString reports whether expected could match valueAsString on
// any non-string, non-BigInt ValueKind (undefined, null, bool, number, RegExp,
// node/object, or list).
func couldMatchNonString(s string) bool {
	if s == "" || s == "undefined" || s == "null" || s == "true" || s == "false" || s == "NaN" || s == "Infinity" || s == "-Infinity" {
		return true
	}
	if s[0] == '/' || s[0] == '[' || s[0] == '-' || (s[0] >= '0' && s[0] <= '9') {
		return true
	}
	return strings.IndexByte(s, ',') >= 0
}

func collectDispatchAttrs(sel selector, attrs []dispatchAttr) []dispatchAttr {
	switch value := sel.(type) {
	case subjectSelector:
		return collectDispatchAttrs(value.Inner, attrs)
	case compoundSelector:
		attrs = collectDispatchAttrs(value.Left, attrs)
		return collectDispatchAttrs(value.Right, attrs)
	case attrSelector:
		attrs = collectDispatchAttrs(value.Inner, attrs)
		if value.Op == attrEqual && canDispatchPath(value.Path) {
			if expected, ok := exactStringValue(value.Value); ok {
				attrs = append(attrs, dispatchAttr{
					path:  value.Path,
					key:   pathKey(value.Path),
					value: expected,
				})
			}
		}
	case classSelector:
		return collectDispatchAttrs(value.Inner, attrs)
	case combinatorSelector:
		if hasReverseSiblingSubject(value) {
			return attrs
		}
		return collectDispatchAttrs(value.Right, attrs)
	case combinedPseudo:
		return collectDispatchAttrs(value.Inner, attrs)
	}
	return attrs
}

func canDispatchPath(path []pathSegment) bool {
	if len(path) == 0 {
		return false
	}
	for _, seg := range path {
		if seg.special == pathLength || seg.index >= 0 {
			return false
		}
	}
	last := path[len(path)-1]
	return last.raw == "name" || last.raw == "operator" || last.raw == "kind" || last.raw == "value" || last.special == pathType
}

func pathKey(path []pathSegment) string {
	if len(path) == 1 {
		return path[0].raw
	}
	var sb strings.Builder
	for i, seg := range path {
		if i > 0 {
			sb.WriteByte(0)
		}
		sb.WriteString(seg.raw)
	}
	return sb.String()
}

func exactStringValue(value attrValue) (string, bool) {
	switch value.Kind {
	case attrValueString:
		return value.Str, true
	case attrValueIdent:
		return value.Ident, true
	default:
		return "", false
	}
}

func dispatchValueForPath(sel selector, targetKey string) (string, bool) {
	for _, attr := range collectDispatchAttrs(sel, nil) {
		if attr.key == targetKey {
			return attr.value, true
		}
	}
	return "", false
}

func (d *stringAttrDispatch) lookup(tree *estree.Tree, id estree.NodeID) []int {
	val := lookupAttrPath(tree, id, d.path)
	var key string
	switch val.Kind {
	case estree.ValueString, estree.ValueBigInt:
		key = val.Str
	default:
		if !d.hasNonStringKey {
			return nil
		}
		key = valueAsString(tree, val)
	}
	if d.byValue == nil {
		if key == d.singleValue {
			return d.singleHits
		}
		return nil
	}
	return d.byValue[key]
}

func (bucket *ruleBucket) evaluate(tree *estree.Tree, id estree.NodeID, ctx *rule.RuleContext, memo *evalMemo) {
	d := bucket.dispatch
	if d == nil {
		for index := range bucket.entries {
			bucket.matchEntry(index, tree, id, ctx, memo)
		}
		return
	}

	if d.next == nil {
		hits := d.lookup(tree, id)
		if len(d.fallback) == 0 {
			for _, index := range hits {
				bucket.matchEntry(index, tree, id, ctx, memo)
			}
			return
		}
		if len(hits) == 0 {
			for _, index := range d.fallback {
				bucket.matchEntry(index, tree, id, ctx, memo)
			}
			return
		}
		// Merge fallback and hits in ESLint specificity order.
		left, right := 0, 0
		for left < len(d.fallback) || right < len(hits) {
			var index int
			switch {
			case right == len(hits):
				index = d.fallback[left]
				left++
			case left == len(d.fallback):
				index = hits[right]
				right++
			case d.fallback[left] < hits[right]:
				index = d.fallback[left]
				left++
			default:
				index = hits[right]
				right++
			}
			bucket.matchEntry(index, tree, id, ctx, memo)
		}
		return
	}

	var matchedBuf [16]int
	matched := matchedBuf[:0]
	for cur := d; cur != nil; cur = cur.next {
		matched = append(matched, cur.lookup(tree, id)...)
		matched = append(matched, cur.fallback...)
	}
	if len(matched) > 1 {
		slices.Sort(matched)
	}
	for _, index := range matched {
		bucket.matchEntry(index, tree, id, ctx, memo)
	}
}

func (bucket *ruleBucket) matchEntry(index int, tree *estree.Tree, id estree.NodeID, ctx *rule.RuleContext, memo *evalMemo) {
	entry := &bucket.entries[index]
	if matchesSelector(entry.compiled, tree, id, estree.NoNode, memo) {
		ctx.ReportRange(tree.Range(id), rule.RuleMessage{
			Id:          "restrictedSyntax",
			Description: entry.message,
		})
	}
}

func deduplicateRuleEntries(entries []ruleEntry) []ruleEntry {
	if len(entries) < 2 {
		return entries
	}
	positions := make(map[string]int, len(entries))
	result := make([]ruleEntry, 0, len(entries))
	for _, entry := range entries {
		if index, ok := positions[entry.selector]; ok {
			result[index] = entry
			continue
		}
		positions[entry.selector] = len(result)
		result = append(result, entry)
	}
	return result
}

type selectorSpecificity struct {
	attributes  int
	identifiers int
}

func sortRuleEntries(entries []ruleEntry) {
	if len(entries) < 2 {
		return
	}
	specificities := make(map[string]selectorSpecificity, len(entries))
	for _, entry := range entries {
		specificities[entry.selector] = analyzeSelectorSpecificity(entry.compiled)
	}
	sort.Slice(entries, func(left, right int) bool {
		a := specificities[entries[left].selector]
		b := specificities[entries[right].selector]
		if a.attributes != b.attributes {
			return a.attributes < b.attributes
		}
		if a.identifiers != b.identifiers {
			return a.identifiers < b.identifiers
		}
		return ecmascript.CompareStrings(entries[left].selector, entries[right].selector) < 0
	})
}

func analyzeSelectorSpecificity(sel selector) selectorSpecificity {
	var result selectorSpecificity
	var visit func(selector)
	visit = func(current selector) {
		switch value := current.(type) {
		case subjectSelector:
			visit(value.Inner)
		case identifierSelector:
			if !value.Wildcard {
				result.identifiers++
			}
		case compoundSelector:
			visit(value.Left)
			visit(value.Right)
		case attrSelector:
			visit(value.Inner)
			result.attributes++
		case classSelector:
			visit(value.Inner)
			result.attributes++
		case combinatorSelector:
			visit(value.Left)
			visit(value.Right)
		case combinedPseudo:
			visit(value.Inner)
			visitPseudoSpecificity(value.Pseudo, &result, visit)
		case pseudoSelector:
			visitPseudoSpecificity(value, &result, visit)
		case unionSelector:
			for _, child := range value.Selectors {
				visit(child)
			}
		}
	}
	visit(sel)
	return result
}

func visitPseudoSpecificity(value pseudoSelector, result *selectorSpecificity, visit func(selector)) {
	switch value.Name {
	case "is", "matches", "not":
		for _, child := range value.Args {
			visit(child)
		}
	case "nth-child", "nth-last-child":
		result.attributes++
		// ESLint deliberately does not include selectors nested inside :has().
	}
}

// parseRuleOptions turns the rule's options — a variadic list of selectors,
// each either a string or a `{ selector, message? }` object — into ruleEntry
// values. Selectors that fail to parse are silently dropped. This is a
// deliberate compatibility divergence from ESLint, whose config validation
// rejects the whole rule configuration for malformed selectors.
func parseRuleOptions(options []any) []ruleEntry {
	var entries []ruleEntry
	for _, item := range options {
		switch x := item.(type) {
		case string:
			entries = append(entries, buildEntryFromString(x))
		case map[string]interface{}:
			if e, ok := buildEntryFromObject(x); ok {
				entries = append(entries, e)
			}
		}
	}
	out := make([]ruleEntry, 0, len(entries))
	for _, e := range entries {
		if e.compiled != nil {
			out = append(out, e)
		}
	}
	return out
}

func buildEntryFromString(sel string) ruleEntry {
	compiled, exit, err := parseRuleSelector(sel)
	if err != nil {
		return ruleEntry{selector: sel}
	}
	return ruleEntry{selector: sel, compiled: compiled, message: defaultRuleMessage(sel), exit: exit}
}

func buildEntryFromObject(m map[string]interface{}) (ruleEntry, bool) {
	rawSel, _ := m["selector"].(string)
	if rawSel == "" {
		return ruleEntry{}, false
	}
	compiled, exit, err := parseRuleSelector(rawSel)
	if err != nil {
		return ruleEntry{}, false
	}
	msg, _ := m["message"].(string)
	if msg == "" {
		msg = defaultRuleMessage(rawSel)
	}
	return ruleEntry{selector: rawSel, compiled: compiled, message: msg, exit: exit}, true
}

func parseRuleSelector(sel string) (selector, bool, error) {
	exit := strings.HasSuffix(sel, ":exit")
	compiled, err := parseSelector(strings.TrimSuffix(sel, ":exit"))
	return compiled, exit, err
}

func defaultRuleMessage(selector string) string {
	return "Using '" + selector + "' is not allowed."
}
