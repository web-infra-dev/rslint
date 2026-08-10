package no_restricted_syntax

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoRestrictedSyntaxRule is the rslint port of ESLint's `no-restricted-syntax`.
//
// Configuration mirrors ESLint: the rule accepts a list of strings or
// `{ selector, message? }` objects. Each selector follows the esquery
// grammar (the subset implemented in parser.go). At lint time every node
// is matched against every parsed selector; each match produces one
// diagnostic with the user-supplied message (or a default).
//
// https://eslint.org/docs/latest/rules/no-restricted-syntax
var NoRestrictedSyntaxRule = rule.Rule{
	Name: "no-restricted-syntax",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		plan := cachedRulePlan(_options)
		if len(plan.buckets) == 0 {
			return rule.RuleListeners{}
		}

		mc := &matchContext{sf: ctx.SourceFile}

		visit := func(node *ast.Node, bucket *ruleBucket) {
			first, second, useAll := bucket.candidates(node, mc)
			if useAll {
				for index := range bucket.entries {
					bucket.matchAndReport(index, node, mc, &ctx)
				}
				return
			}

			// Both candidate lists retain option order. Merge them so a
			// dispatch hit cannot reorder diagnostics relative to selectors
			// that do not participate in the index.
			left, right := 0, 0
			for left < len(first) || right < len(second) {
				var index int
				switch {
				case right == len(second):
					index = first[left]
					left++
				case left == len(first):
					index = second[right]
					right++
				case first[left] < second[right]:
					index = first[left]
					left++
				default:
					index = second[right]
					right++
				}
				bucket.matchAndReport(index, node, mc, &ctx)
			}
		}

		listeners := make(rule.RuleListeners, len(plan.buckets))
		for k, bucket := range plan.buckets {
			listeners[k] = func(node *ast.Node) {
				visit(node, bucket)
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
	entries := parseRuleOptions(rule.LegacyUnwrapOptions(options))
	if len(entries) == 0 {
		return plan
	}

	// Build per-kind buckets of (selector, message) entries. Selectors whose
	// candidate kinds are the universe land in everyKind and are evaluated on
	// every interesting visit.
	perKind := make(map[ast.Kind][]ruleEntry)
	var everyKind []ruleEntry
	for _, entry := range entries {
		kinds := candidateKinds(entry.compiled)
		if kinds.universe {
			everyKind = append(everyKind, entry)
			continue
		}
		for kind := range kinds.kinds {
			perKind[kind] = append(perKind[kind], entry)
		}
	}
	if len(everyKind) > 0 {
		for _, kind := range allInterestingKinds {
			perKind[kind] = append(perKind[kind], everyKind...)
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
	entries  []ruleEntry
	dispatch *stringAttrDispatch
}

type stringAttrDispatch struct {
	path     []string
	byValue  map[string][]int
	fallback []int
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
	}
	for index, entry := range entries {
		value, ok := dispatchValueForPath(entry.compiled, bestKey)
		if !ok {
			dispatch.fallback = append(dispatch.fallback, index)
			continue
		}
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
	switch value := sel.(type) {
	case attrSelector:
		attrs = collectDispatchAttrs(value.Inner, attrs)
		if value.Op == attrEqual {
			if expected, ok := exactStringValue(value.Value); ok {
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

func (bucket *ruleBucket) matchAndReport(index int, node *ast.Node, mc *matchContext, ctx *rule.RuleContext) {
	entry := &bucket.entries[index]
	if !matches(entry.compiled, node, mc) {
		return
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "restrictedSyntax",
		Description: entry.formatMessage(),
	})
}

func (e ruleEntry) formatMessage() string {
	if e.message != "" {
		return e.message
	}
	return fmt.Sprintf("Using '%s' is not allowed.", e.selector)
}

// parseRuleOptions normalises the loosely-typed options value handed to
// the rule into a list of ruleEntry. ESLint accepts a mix of strings and
// `{ selector, message? }` objects; rslint receives a single string, a
// single object, or an []interface{} depending on how config.go unwrapped
// the array. Selectors that fail to parse are silently dropped — ESLint
// rejects the whole config in that case, but for runtime resilience we
// prefer to drop the offending entry over panicking.
func parseRuleOptions(opts any) []ruleEntry {
	if opts == nil {
		return nil
	}
	var entries []ruleEntry
	switch v := opts.(type) {
	case string:
		entries = append(entries, buildEntryFromString(v))
	case map[string]interface{}:
		if e, ok := buildEntryFromObject(v); ok {
			entries = append(entries, e)
		}
	case []interface{}:
		for _, item := range v {
			switch x := item.(type) {
			case string:
				entries = append(entries, buildEntryFromString(x))
			case map[string]interface{}:
				if e, ok := buildEntryFromObject(x); ok {
					entries = append(entries, e)
				}
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
	compiled, err := parseSelector(sel)
	if err != nil {
		return ruleEntry{selector: sel}
	}
	return ruleEntry{selector: sel, compiled: compiled}
}

func buildEntryFromObject(m map[string]interface{}) (ruleEntry, bool) {
	rawSel, _ := m["selector"].(string)
	if rawSel == "" {
		return ruleEntry{}, false
	}
	compiled, err := parseSelector(rawSel)
	if err != nil {
		return ruleEntry{}, false
	}
	msg, _ := m["message"].(string)
	return ruleEntry{selector: rawSel, compiled: compiled, message: msg}, true
}
