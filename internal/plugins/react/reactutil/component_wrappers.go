package reactutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// ComponentWrapperEntry describes one user-configured component-wrapping
// call site. Either form is recognized:
//
//   - `{property: "memo", object: "React"}` matches `<object>.<property>(fn)`
//     calls. Empty `object` is treated as `DefaultReactPragma`.
//   - `{property: "memo"}` matches bare `<property>(fn)` calls when `object`
//     is empty.
//
// Mirrors eslint-plugin-react's `settings.componentWrapperFunctions` —
// strings in the user setting expand to `{property: <s>}`, objects pass
// through.
//
// `IsUserConfigured` distinguishes entries the user explicitly added via
// `settings.componentWrapperFunctions` from entries we inject as
// hardcoded defaults (memo / forwardRef, pragma-qualified or bare).
// Upstream's `isDestructuredFromPragmaImport` adds a runtime guard to
// bare default entries — they only match when the bare callee was
// destructure-imported from the pragma module. User-configured bare
// entries match freely because they do not depend on import resolution.
type ComponentWrapperEntry struct {
	Object           string
	Property         string
	IsUserConfigured bool
}

// DefaultComponentWrappers is the always-on wrapper list every React rule
// uses regardless of `settings.componentWrapperFunctions`. Mirrors upstream:
// `{property: 'memo', object: pragma}`, `{property: 'forwardRef', object: pragma}`,
// plus the bare `memo` / `forwardRef` aliases.
func DefaultComponentWrappers(pragma string) []ComponentWrapperEntry {
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	return []ComponentWrapperEntry{
		{Object: pragma, Property: "memo"},
		{Object: pragma, Property: "forwardRef"},
		{Property: "memo"},
		{Property: "forwardRef"},
	}
}

// GetComponentWrapperFunctions reads `settings.componentWrapperFunctions`
// and merges the user's additions on top of `DefaultComponentWrappers`.
// Accepted shapes per entry:
//
//   - string: "myMemo" → {Property: "myMemo"}
//   - object: {"property": "memo", "object": "React"} →
//     {Object: "React", Property: "memo"}; "object" defaults to empty
//     (bare call) when omitted
//   - object with `"object": "<pragma>"` placeholder — upstream's
//     `getWrapperFunctions` (Components.js) substitutes the placeholder
//     with the configured pragma at read time, so users can write
//     `{property: 'memo', object: '<pragma>'}` and have it match
//     whichever pragma the file is configured with. We mirror that
//     substitution exactly.
//
// Unknown / malformed entries are silently ignored, matching upstream.
func GetComponentWrapperFunctions(settings map[string]interface{}, pragma string) []ComponentWrapperEntry {
	out := DefaultComponentWrappers(pragma)
	if settings == nil {
		return out
	}
	raw, ok := settings["componentWrapperFunctions"]
	if !ok {
		return out
	}
	resolvedPragma := pragma
	if resolvedPragma == "" {
		resolvedPragma = DefaultReactPragma
	}
	add := func(v interface{}) {
		switch e := v.(type) {
		case string:
			if e != "" {
				out = append(out, ComponentWrapperEntry{Property: e, IsUserConfigured: true})
			}
		case map[string]interface{}:
			prop, _ := e["property"].(string)
			if prop == "" {
				return
			}
			obj, _ := e["object"].(string)
			if obj == "<pragma>" {
				obj = resolvedPragma
			}
			out = append(out, ComponentWrapperEntry{Object: obj, Property: prop, IsUserConfigured: true})
		}
	}
	switch r := raw.(type) {
	case []interface{}:
		for _, v := range r {
			add(v)
		}
	default:
		add(r)
	}
	return out
}

// MatchesAnyComponentWrapper reports whether `call` matches any entry in
// `wrappers`, with `fn` as its first argument (paren / TS-wrapper transparent).
// Pass an empty pragma to default to "React"; the pragma is only consulted
// for entries whose `Object` is empty AND need to fall back to the configured
// pragma — but `DefaultComponentWrappers` already pre-fills the pragma form,
// so callers normally shouldn't need to thread pragma through twice.
//
// Optional-chain handling mirrors upstream's `isPragmaComponentWrapper`:
//
//   - Member-level and call-level optional forms (`React?.memo(arg)` and
//     `(React.memo)?.(arg)`) are recognized because the ESTree callee remains
//     a MemberExpression in both cases.
//
//   - A call-level optional bare Identifier (`memo?.(arg)`) follows the same
//     configured-wrapper or pragma-import gate as its non-optional form.
func MatchesAnyComponentWrapper(call, fn *ast.Node, wrappers []ComponentWrapperEntry) bool {
	return matchesAnyComponentWrapperCore(call, fn, wrappers, "", nil)
}

// MatchesAnyComponentWrapperWithChecker is the import-aware variant.
// When `tc` is non-nil and the callee is a bare Identifier matching a
// hardcoded bare default entry (`{Property: "memo"}` /
// `{Property: "forwardRef"}` from `DefaultComponentWrappers`), the
// callee's binding must additionally be destructured from / imported
// from / required from the pragma module (per
// `IsDestructuredFromPragmaImport`). This precisely mirrors upstream's
//
//	(!wrapperFunction.object ||
//	 (wrapperFunction.object === pragma &&
//	  this.isDestructuredFromPragmaImport(node, node.callee.name)))
//
// gate. Without this, `memo(arrow)` would silently classify when `memo`
// is a user-defined function unrelated to React — leading to over-reports
// where upstream skips. Use this variant whenever a TypeChecker is
// available; otherwise the same helper falls back to its syntax-only import
// scan.
func MatchesAnyComponentWrapperWithChecker(call, fn *ast.Node, wrappers []ComponentWrapperEntry, pragma string, tc *checker.Checker) bool {
	return matchesAnyComponentWrapperCore(call, fn, wrappers, pragma, tc)
}

func matchesAnyComponentWrapperCore(call, fn *ast.Node, wrappers []ComponentWrapperEntry, pragma string, tc *checker.Checker) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	c := call.AsCallExpression()
	if c.Arguments == nil || len(c.Arguments.Nodes) == 0 {
		return false
	}
	if SkipExpressionWrappers(c.Arguments.Nodes[0]) != fn {
		return false
	}
	callee := SkipExpressionWrappers(c.Expression)
	for _, w := range wrappers {
		if w.Property == "" {
			continue
		}
		switch callee.Kind {
		case ast.KindIdentifier:
			if callee.AsIdentifier().Text != w.Property {
				continue
			}
			if w.Object == "" {
				if w.IsUserConfigured {
					// User-configured bare entry: accept any callee shape
					// (call-level optional included). User entries don't
					// need the pragma-import gate since they're explicit
					// opt-in.
					return true
				}
				// Hardcoded bare default (memo / forwardRef without
				// object): upstream gates with
				// `isDestructuredFromPragmaImport`. We always run that
				// gate — when a TypeChecker is available it resolves
				// the binding precisely, and when not it falls back to
				// a syntax-only SourceFile scan that handles the
				// canonical top-level pragma-import shapes.
				if !IsDestructuredFromPragmaImport(callee, pragma, tc) {
					continue
				}
				return true
			}
			// Entry HAS an Object — upstream's bare-callee arm:
			//
			//   wrapperFunction.property === node.callee.name && (
			//     !wrapperFunction.object
			//     || (wrapperFunction.object === pragma &&
			//         this.isDestructuredFromPragmaImport(node, node.callee.name))
			//   )
			//
			// translates to: when the entry's Object equals the active
			// pragma AND the bare identifier callee is destructured /
			// imported / required from the pragma module, the entry
			// still matches even though `node.callee` is not a
			// MemberExpression. This covers e.g.
			// `componentWrapperFunctions: [{property: 'observer', object: '<pragma>'}]`
			// + `import { observer } from 'react'` + `observer(arrow)`.
			effectivePragma := pragma
			if effectivePragma == "" {
				effectivePragma = DefaultReactPragma
			}
			if w.Object != effectivePragma {
				continue
			}
			if !IsDestructuredFromPragmaImport(callee, pragma, tc) {
				continue
			}
			return true
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if w.Object == "" {
				continue
			}
			var object, property *ast.Node
			if callee.Kind == ast.KindPropertyAccessExpression {
				member := callee.AsPropertyAccessExpression()
				object, property = member.Expression, member.Name()
			} else {
				member := callee.AsElementAccessExpression()
				object, property = member.Expression, ast.SkipParentheses(member.ArgumentExpression)
			}
			obj := SkipExpressionWrappers(object)
			if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != w.Object {
				continue
			}
			if EsTreeName(property) == w.Property {
				return true
			}
		}
	}
	return false
}
