package reactutil

import "slices"

// ComponentMap maps a component tag name (e.g. "a", "Link") to the set of
// attribute names that identify its link target (e.g. ["href"] or ["to"]).
type ComponentMap map[string][]string

// DefaultLinkComponents returns the default link-component map: {"a": ["href"]}.
func DefaultLinkComponents() ComponentMap {
	return ComponentMap{"a": {"href"}}
}

// DefaultFormComponents returns the default form-component map: {"form": ["action"]}.
func DefaultFormComponents() ComponentMap {
	return ComponentMap{"form": {"action"}}
}

// ReadComponentsFromSettings extracts a component-name→attribute-list map
// from `settings.<key>`, matching upstream `util/linkComponents`.
//
// Upstream builds the map via `new Map(DEFAULT.concat(settings[key]).map(…))`,
// where same-key entries use last-wins (replace) semantics. This function
// mirrors that: a settings entry for an already-present component replaces
// the base entry entirely.
//
// Shapes accepted (each entry may appear standalone or as an element of an
// outer array, mirroring upstream's `DEFAULT.concat(settings[key] || [])`):
//
//   - string: "Link"                                    → {Link: [defaultAttr]}
//   - {name, <attrField>}: <attrField> string or []str  → {name: [attr…]}
//
// `attrField` is "linkAttribute" for linkComponents and "formAttribute" for
// formComponents — upstream uses distinct field names for each category
// (`value.linkAttribute` vs `value.formAttribute`), so getting this wrong
// would silently fall back to the default attribute for every custom form
// component the user configures.
func ReadComponentsFromSettings(settings map[string]interface{}, key, attrField, defaultAttr string, base ComponentMap) ComponentMap {
	out := ComponentMap{}
	for k, v := range base {
		out[k] = slices.Clone(v)
	}
	if settings == nil {
		return out
	}
	raw, ok := settings[key]
	if !ok {
		return out
	}
	// addOne mirrors upstream's per-entry mapper inside the Map constructor.
	// Each entry REPLACES any previous entry with the same name (last-wins),
	// matching `new Map([...pairs])` semantics.
	addOne := func(entry interface{}) {
		switch e := entry.(type) {
		case string:
			out[e] = []string{defaultAttr}
		case map[string]interface{}:
			name, _ := e["name"].(string)
			if name == "" {
				return
			}
			var attrs []string
			// Mirrors upstream's `[].concat(value[attrField])` coercion:
			// string → single-element list, array → as-is, missing → empty
			// (which we backfill with the default attribute).
			switch la := e[attrField].(type) {
			case string:
				attrs = []string{la}
			case []interface{}:
				for _, v := range la {
					if s, ok := v.(string); ok {
						attrs = append(attrs, s)
					}
				}
			}
			if len(attrs) == 0 {
				attrs = []string{defaultAttr}
			}
			out[name] = attrs
		}
	}
	// Upstream accepts either a single entry (string/object) or an array of
	// them at `settings[key]`. JS's `[].concat(x)` flattens both into the
	// final list; we mirror that by accepting either shape here.
	switch r := raw.(type) {
	case string:
		addOne(r)
	case map[string]interface{}:
		addOne(r)
	case []interface{}:
		for _, entry := range r {
			addOne(entry)
		}
	}
	return out
}
