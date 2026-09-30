package valid_title

import (
	"strings"

	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

type matcherEntry struct {
	re *esregexp.RegExp
	// customText non-empty ⇒ use mustMatchCustom / mustNotMatchCustom
	customText string
}

type matchersByFn struct {
	describe matcherEntry
	test     matcherEntry
	it       matcherEntry
}

type compiledOptions struct {
	ignoreSpaces             bool
	ignoreTypeOfDescribeName bool
	ignoreTypeOfTestName     bool
	disallowedConcat         *esregexp.RegExp
	invalidPatterns          []invalidPattern
	mustNotMatch             matchersByFn
	mustMatch                matchersByFn
}

type invalidPattern struct {
	optionPath string
	pattern    string
	err        error
}

func firstOptionMap(options []any) map[string]interface{} {
	if len(options) == 0 {
		return nil
	}
	m, ok := options[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return m
}

func boolFromMap(m map[string]interface{}, key string, def bool) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

func compileRE2(pat string) (*esregexp.RegExp, error) {
	re, err := esregexp.Compile(pat, "u")
	if err != nil {
		return nil, err
	}
	return re, nil
}

func compileMatcherPatterns(raw interface{}, optionPath string) (matchersByFn, []invalidPattern) {
	out := matchersByFn{}
	var invalids []invalidPattern
	if raw == nil {
		return out, nil
	}

	setAll := func(e matcherEntry) {
		out.describe, out.test, out.it = e, e, e
	}

	switch x := raw.(type) {
	case string:
		if x == "" {
			break
		}
		if re, err := compileRE2(x); err != nil {
			invalids = append(invalids, invalidPattern{
				optionPath: optionPath,
				pattern:    x,
				err:        err,
			})
		} else if re != nil {
			setAll(matcherEntry{re: re})
		}
	case []interface{}:
		me := matcherEntry{}
		if len(x) >= 1 {
			if s, ok := x[0].(string); ok && s != "" {
				re, err := compileRE2(s)
				if err != nil {
					invalids = append(invalids, invalidPattern{
						optionPath: optionPath,
						pattern:    s,
						err:        err,
					})
				} else {
					me.re = re
				}
			}
		}
		if len(x) >= 2 {
			if s, ok := x[1].(string); ok {
				me.customText = s
			}
		}
		if me.re != nil {
			setAll(me)
		}
	case map[string]interface{}:
		for _, key := range []string{"describe", "test", "it"} {
			if v, ok := x[key]; ok {
				invalids = append(invalids, fillMatcherField(&out, key, v, optionPath+"."+key)...)
			}
		}
	}
	return out, invalids
}

func fillMatcherField(ms *matchersByFn, key string, raw interface{}, optionPath string) []invalidPattern {
	e := matcherEntry{}
	var invalids []invalidPattern

	switch x := raw.(type) {
	case string:
		if x == "" {
			break
		}
		re, err := compileRE2(x)
		if err != nil {
			invalids = append(invalids, invalidPattern{
				optionPath: optionPath,
				pattern:    x,
				err:        err,
			})
		} else {
			e.re = re
		}
	case []interface{}:
		if len(x) >= 1 {
			if s, ok := x[0].(string); ok && s != "" {
				re, err := compileRE2(s)
				if err != nil {
					invalids = append(invalids, invalidPattern{
						optionPath: optionPath,
						pattern:    s,
						err:        err,
					})
				} else {
					e.re = re
				}
			}
		}
		if len(x) >= 2 {
			if s, ok := x[1].(string); ok {
				e.customText = s
			}
		}
	}

	switch key {
	case "describe":
		ms.describe = e
	case "test":
		ms.test = e
	case "it":
		ms.it = e
	}

	return invalids
}

func parseCompiledOptions(options []any) compiledOptions {
	m := firstOptionMap(options)
	if m == nil {
		return compiledOptions{}
	}

	co := compiledOptions{
		ignoreSpaces:             boolFromMap(m, "ignoreSpaces", false),
		ignoreTypeOfDescribeName: boolFromMap(m, "ignoreTypeOfDescribeName", false),
		ignoreTypeOfTestName:     boolFromMap(m, "ignoreTypeOfTestName", false),
	}

	if dw, ok := m["disallowedWords"]; ok && dw != nil {
		co.disallowedConcat, co.invalidPatterns = compileDisallowedWords(dw, co.invalidPatterns)
	}

	if mn, ok := m["mustNotMatch"]; ok {
		var invalids []invalidPattern
		co.mustNotMatch, invalids = compileMatcherPatterns(mn, "mustNotMatch")
		co.invalidPatterns = append(co.invalidPatterns, invalids...)
	}
	if mm, ok := m["mustMatch"]; ok {
		var invalids []invalidPattern
		co.mustMatch, invalids = compileMatcherPatterns(mm, "mustMatch")
		co.invalidPatterns = append(co.invalidPatterns, invalids...)
	}

	return co
}

func compileDisallowedWords(raw interface{}, invalids []invalidPattern) (*esregexp.RegExp, []invalidPattern) {
	items, ok := raw.([]interface{})
	if !ok || len(items) == 0 {
		return nil, invalids
	}
	parts := make([]string, 0, len(items))
	for _, it := range items {
		w, ok := it.(string)
		if ok && w != "" {
			parts = append(parts, w)
		}
	}
	if len(parts) == 0 {
		return nil, invalids
	}
	// Upstream: new RegExp(`\\b(${words.join("|")})\\b`, "iu").
	pattern := "\\b(" + strings.Join(parts, "|") + ")\\b"
	re, err := esregexp.Compile(pattern, "iu")
	if err != nil {
		invalids = append(invalids, invalidPattern{
			optionPath: "disallowedWords",
			pattern:    pattern,
			err:        err,
		})
		return nil, invalids
	}
	return re, invalids
}
