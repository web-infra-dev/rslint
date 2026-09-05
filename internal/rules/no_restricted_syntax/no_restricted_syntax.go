package no_restricted_syntax

import (
	_ "embed"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed no_restricted_syntax.schema.json
var schemaJSON []byte

// NoRestrictedSyntaxRule is the rslint port of ESLint's `no-restricted-syntax`.
//
// Configuration mirrors ESLint: the rule accepts a list of strings or
// `{ selector, message? }` objects. Each selector follows the esquery
// grammar (the subset implemented in parser.go). Parsed selectors are indexed
// by candidate kind and common string attributes; each match produces one
// diagnostic with the user-supplied message (or a default).
//
// https://eslint.org/docs/latest/rules/no-restricted-syntax
var NoRestrictedSyntaxRule = rule.Rule{
	Name:   "no-restricted-syntax",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		plan := cachedRulePlan(options)
		if len(plan.buckets) == 0 {
			return rule.RuleListeners{}
		}

		mc := &matchContext{sf: ctx.SourceFile}

		visit := func(node *ast.Node, bucket *ruleBucket) {
			if isTransparentEstreeContainer(node) ||
				(node.Kind == ast.KindQualifiedName && !utils.IsHeritageQualifiedName(node)) {
				return
			}
			first, second, useAll := bucket.candidates(node, mc)
			visitCandidates := func(match func(index int, indexed bool)) {
				if useAll {
					for index := range bucket.entries {
						match(index, false)
					}
					return
				}

				// Both candidate lists retain ESLint specificity order. Merge them so a
				// dispatch hit cannot reorder diagnostics relative to selectors
				// that do not participate in the index.
				left, right := 0, 0
				for left < len(first) || right < len(second) {
					var index int
					indexed := false
					switch {
					case right == len(second):
						index = first[left]
						left++
					case left == len(first):
						index = second[right]
						right++
						indexed = true
					case first[left] < second[right]:
						index = first[left]
						left++
					default:
						index = second[right]
						right++
						indexed = true
					}
					match(index, indexed)
				}
			}

			// A dispatch key read from the physical node cannot prune its
			// folded ESTree child: type, parent and fields can all differ.
			target := virtualTarget(node)
			visitVirtual := func() {
				if !bucket.hasVirtual {
					return
				}
				if node.Kind == ast.KindConstructor {
					keyTarget := constructorKey(node, mc).typeName
					for index := range bucket.entries {
						bucket.matchVirtual(index, node, keyTarget, mc, &ctx)
					}
				}
				if target != "" {
					for index := range bucket.entries {
						bucket.matchVirtual(index, node, target, mc, &ctx)
					}
				}
			}
			// ChainExpression surrounds the physical node; the other facades
			// are its children. Exit events reverse that order.
			virtualBeforePhysical := bucket.entries[0].exit != (target == "ChainExpression")
			if virtualBeforePhysical {
				visitVirtual()
			}
			visitCandidates(func(index int, indexed bool) {
				bucket.matchPhysical(index, node, mc, &ctx, indexed)
			})
			if !virtualBeforePhysical {
				visitVirtual()
			}
		}

		listeners := make(rule.RuleListeners, len(plan.buckets))
		for k, bucket := range plan.buckets {
			switch k {
			case ast.KindSourceFile:
				// The engine starts at SourceFile's children. Program enter runs
				// eagerly, and EOF provides the matching Program exit event.
				visit(ctx.SourceFile.AsNode(), bucket)
			case rule.ListenerOnExit(ast.KindSourceFile):
				listeners[rule.ListenerOnExit(ast.KindEndOfFile)] = func(*ast.Node) { visit(ctx.SourceFile.AsNode(), bucket) }
			default:
				listeners[k] = func(node *ast.Node) { visit(node, bucket) }
			}
		}
		return listeners
	},
}

// ruleEntry is a parsed option entry: the original selector string,
// the compiled selector tree, and the message (custom or default).
type ruleEntry struct {
	selector string
	compiled selector
	message  string
	exit     bool
}

const maxCachedRulePlans = 64

type rulePlan struct {
	// options keeps the backing array alive while its address identifies this
	// normalized config in rulePlanCache. Rule options are immutable after
	// config resolution.
	options []any
	buckets map[ast.Kind]*ruleBucket
}

type rulePlanCacheKey struct {
	options uintptr
	length  int
}

var emptyRulePlan = &rulePlan{}

var rulePlanCache = struct {
	sync.RWMutex
	entries map[rulePlanCacheKey]*rulePlan
	order   []rulePlanCacheKey
}{
	entries: make(map[rulePlanCacheKey]*rulePlan),
	order:   make([]rulePlanCacheKey, 0, maxCachedRulePlans),
}

func cachedRulePlan(options []any) *rulePlan {
	if len(options) == 0 {
		return emptyRulePlan
	}
	key := rulePlanCacheKey{
		options: reflect.ValueOf(options).Pointer(),
		length:  len(options),
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

func buildRulePlan(options []any) *rulePlan {
	plan := &rulePlan{options: options}
	entries := parseRuleOptions(options)
	entries = deduplicateRuleEntries(entries)
	sortRuleEntries(entries)
	if len(entries) == 0 {
		return plan
	}

	// Build per-kind buckets in ESLint's specificity order. This order is
	// observable when multiple selectors report the same node, including when
	// a universe selector is interleaved with kind-specific selectors.
	perKind := make(map[ast.Kind][]ruleEntry)
	for _, entry := range entries {
		kinds := candidateKinds(entry.compiled)
		listenerKind := func(kind ast.Kind) ast.Kind {
			if entry.exit {
				return rule.ListenerOnExit(kind)
			}
			return kind
		}
		if kinds.universe {
			for _, kind := range allInterestingKinds {
				key := listenerKind(kind)
				perKind[key] = append(perKind[key], entry)
			}
			continue
		}
		for kind := range kinds.kinds {
			key := listenerKind(kind)
			perKind[key] = append(perKind[key], entry)
		}
	}
	plan.buckets = make(map[ast.Kind]*ruleBucket, len(perKind))
	for kind, entries := range perKind {
		bucket := buildRuleBucket(entries)
		plan.buckets[kind] = &bucket
	}
	return plan
}

type ruleBucket struct {
	entries    []ruleEntry
	dispatch   *stringAttrDispatch
	hasVirtual bool
}

type stringAttrDispatch struct {
	path     []string
	byValue  map[string][]int
	fallback []int
	matched  []selector
}

type dispatchAttr struct {
	path  []string
	key   string
	value string
}

type dispatchStat struct {
	path   []string
	count  int
	values map[string]struct{}
}

func buildRuleBucket(entries []ruleEntry) ruleBucket {
	bucket := ruleBucket{entries: entries}

	stats := make(map[string]*dispatchStat)
	for _, entry := range entries {
		bucket.hasVirtual = bucket.hasVirtual || !isBareWildcardSelector(entry.compiled)
		attrs := collectDispatchAttrs(entry.compiled, nil)
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
	if bestCount == 0 || (bestCount == 1 && !supportsSingleDispatch(stats[bestKey].path)) {
		return bucket
	}

	dispatch := &stringAttrDispatch{
		path:     stats[bestKey].path,
		byValue:  make(map[string][]int, bestValues),
		fallback: make([]int, 0, len(entries)-bestCount),
		matched:  make([]selector, len(entries)),
	}
	for index, entry := range entries {
		value, ok := dispatchValueForPath(entry.compiled, bestKey)
		if !ok {
			dispatch.fallback = append(dispatch.fallback, index)
			continue
		}
		stripped, ok := stripDispatchAttr(entry.compiled, bestKey, value)
		if !ok {
			dispatch.fallback = append(dispatch.fallback, index)
			continue
		}
		dispatch.matched[index] = stripped
		dispatch.byValue[value] = append(dispatch.byValue[value], index)
	}
	bucket.dispatch = dispatch
	return bucket
}

func supportsSingleDispatch(path []string) bool {
	if len(path) == 0 {
		return false
	}
	switch path[len(path)-1] {
	case "name", "operator", "type":
		return true
	default:
		return false
	}
}

func collectDispatchAttrs(sel selector, attrs []dispatchAttr) []dispatchAttr {
	if selectorTargetsClassBody(sel) || selectorTargetsJSXEmptyExpression(sel) ||
		selectorTargetsEstreeType(sel, "TSEnumBody") || selectorTargetsEstreeType(sel, "FunctionExpression") ||
		selectorTargetsEstreeType(sel, "JSXElement") || selectorTargetsEstreeType(sel, "JSXOpeningElement") {
		return attrs
	}
	switch value := sel.(type) {
	case subjectSelector:
		attrs = collectDispatchAttrs(value.Inner, attrs)
	case attrSelector:
		attrs = collectDispatchAttrs(value.Inner, attrs)
		if value.Op == attrEqual {
			if expected, ok := exactStringValue(value.Value); ok && expected != "undefined" {
				attrs = append(attrs, dispatchAttr{
					path:  value.Path,
					key:   strings.Join(value.Path, "\x00"),
					value: expected,
				})
			}
		}
	case classSelector:
		attrs = collectDispatchAttrs(value.Inner, attrs)
	case combinatorSelector:
		attrs = collectDispatchAttrs(value.Right, attrs)
	case combinedPseudo:
		attrs = collectDispatchAttrs(value.Inner, attrs)
	}
	return attrs
}

func stripDispatchAttr(sel selector, pathKey, expected string) (selector, bool) {
	switch value := sel.(type) {
	case subjectSelector:
		if inner, ok := stripDispatchAttr(value.Inner, pathKey, expected); ok {
			value.Inner = inner
			return value, true
		}
	case attrSelector:
		if inner, ok := stripDispatchAttr(value.Inner, pathKey, expected); ok {
			value.Inner = inner
			return value, true
		}
		actual, exact := exactStringValue(value.Value)
		if value.Op == attrEqual && exact && actual == expected && strings.Join(value.Path, "\x00") == pathKey {
			return value.Inner, true
		}
	case classSelector:
		if inner, ok := stripDispatchAttr(value.Inner, pathKey, expected); ok {
			value.Inner = inner
			return value, true
		}
	case combinatorSelector:
		if right, ok := stripDispatchAttr(value.Right, pathKey, expected); ok {
			value.Right = right
			return value, true
		}
	case combinedPseudo:
		if inner, ok := stripDispatchAttr(value.Inner, pathKey, expected); ok {
			value.Inner = inner
			return value, true
		}
	}
	return sel, false
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

func dispatchValueForPath(sel selector, pathKey string) (string, bool) {
	for _, attr := range collectDispatchAttrs(sel, nil) {
		if attr.key == pathKey {
			return attr.value, true
		}
	}
	return "", false
}

func (bucket *ruleBucket) candidates(node *ast.Node, mc *matchContext) (first, second []int, useAll bool) {
	if bucket.dispatch == nil {
		return nil, nil, true
	}
	value, ok, handled := lookupStringAttrPath(node, bucket.dispatch.path, mc)
	if !handled {
		return nil, nil, true
	}
	if !ok {
		return bucket.dispatch.fallback, nil, false
	}
	return bucket.dispatch.fallback, bucket.dispatch.byValue[value], false
}

func (bucket *ruleBucket) matchPhysical(index int, node *ast.Node, mc *matchContext, ctx *rule.RuleContext, indexed bool) {
	entry := &bucket.entries[index]
	compiled := entry.compiled
	if indexed && bucket.dispatch != nil && bucket.dispatch.matched[index] != nil {
		compiled = bucket.dispatch.matched[index]
	}
	if node.Kind == ast.KindSourceFile && isBareWildcardSelector(entry.compiled) {
		return
	}
	if isPlainParameter(node) && !isBareWildcardSelector(entry.compiled) {
		return
	}
	if matchesInScopeTarget(compiled, node, mc, nil, "physical") {
		ctx.ReportRange(utils.GetESTreeBindingIdentifierRange(mc.sf, node), rule.RuleMessage{
			Id:          "restrictedSyntax",
			Description: entry.formatMessage(),
		})
	}
}

func (bucket *ruleBucket) matchVirtual(index int, node *ast.Node, target string, mc *matchContext, ctx *rule.RuleContext) {
	entry := &bucket.entries[index]
	// Preserve the documented bare-wildcard traversal. Other broad selectors
	// must see the folded child, including regex and negative selectors.
	if isBareWildcardSelector(entry.compiled) || !matchesInScopeTarget(entry.compiled, node, mc, nil, target) {
		return
	}
	message := rule.RuleMessage{Id: "restrictedSyntax", Description: entry.formatMessage()}
	switch target {
	case "Identifier", "Literal":
		if token, ok := utils.TokenBeforePosition(mc.sf, node.ParameterList().Pos()-1); ok {
			ctx.ReportRange(core.NewTextRange(token.Start, token.End), message)
		}
	case "ClassBody":
		ctx.ReportRange(classBodyTextRange(mc.sf, node), message)
	case "TSEnumBody":
		ctx.ReportRange(enumBodyTextRange(mc.sf, node), message)
	case "JSXEmptyExpression":
		nodeRange := utils.TrimNodeTextRange(mc.sf, node)
		ctx.ReportRange(core.NewTextRange(nodeRange.Pos()+1, max(nodeRange.Pos()+1, nodeRange.End()-1)), message)
	case "FunctionExpression", "TSEmptyBodyFunctionExpression":
		ctx.ReportRange(functionValueTextRange(mc.sf, node), message)
	default:
		ctx.ReportNode(node, message)
	}
}

func (e ruleEntry) formatMessage() string {
	return e.message
}

func functionValueTextRange(sf *ast.SourceFile, node *ast.Node) core.TextRange {
	// tsgo records the position immediately after '(' (or '<' for generic
	// methods). Reuse that parser boundary, including for bodyless overloads.
	if parameters := node.TypeParameterList(); parameters != nil && len(parameters.Nodes) > 0 {
		return core.NewTextRange(parameters.Pos()-1, node.End())
	}
	if parameters := node.ParameterList(); parameters != nil {
		return core.NewTextRange(parameters.Pos()-1, node.End())
	}
	return utils.TrimNodeTextRange(sf, node)
}

func classBodyTextRange(sf *ast.SourceFile, node *ast.Node) core.TextRange {
	var members *ast.NodeList
	switch node.Kind {
	case ast.KindClassDeclaration:
		members = node.AsClassDeclaration().Members
	case ast.KindClassExpression:
		members = node.AsClassExpression().Members
	}
	if sf == nil || members == nil {
		return core.NewTextRange(node.Pos(), node.End())
	}

	return listBodyTextRange(sf, node, members)
}

func enumBodyTextRange(sf *ast.SourceFile, node *ast.Node) core.TextRange {
	if node == nil || node.Kind != ast.KindEnumDeclaration {
		return core.TextRange{}
	}
	return listBodyTextRange(sf, node, node.AsEnumDeclaration().Members)
}

func listBodyTextRange(sf *ast.SourceFile, node *ast.Node, members *ast.NodeList) core.TextRange {
	if members != nil {
		return core.NewTextRange(members.Pos()-1, node.End())
	}
	return utils.TrimNodeTextRange(sf, node)
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
			if value.Name != "*" {
				result.identifiers++
			}
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

// ESLint handles the event-phase suffix outside esquery. Keep the phase on the
// entry so the linter can register the selector on the matching enter or exit
// listener while retaining the original suffix in the diagnostic text.
func parseRuleSelector(sel string) (selector, bool, error) {
	exit := strings.HasSuffix(sel, ":exit")
	compiled, err := parseSelector(strings.TrimSuffix(sel, ":exit"))
	return compiled, exit, err
}

func defaultRuleMessage(selector string) string {
	return "Using '" + selector + "' is not allowed."
}
