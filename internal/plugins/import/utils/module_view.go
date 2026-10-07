package utils

import "github.com/web-infra-dev/rslint/internal/utils/modules"

type moduleViewKind uint8

const (
	moduleViewAuthored moduleViewKind = iota
	moduleViewDefaultOnly
	moduleViewUnknown
)

// moduleViewFor keeps host-defined attribute semantics out of physical module
// resolution. Known text and JSON loaders expose a default value; an unknown
// loader has no statically safe export shape.
func moduleViewFor(attributes modules.ImportAttributes) moduleViewKind {
	switch attributes.State {
	case modules.AttributesNone:
		return moduleViewAuthored
	case modules.AttributesStatic:
		entries := attributes.Entries()
		if len(entries) == 0 {
			return moduleViewAuthored
		}
		moduleType, hasType := attributes.Value("type")
		if hasType && (moduleType == "text" || moduleType == "json") {
			for _, entry := range entries {
				if entry.Name != "type" && entry.Name != "resolution-mode" {
					return moduleViewUnknown
				}
			}
			return moduleViewDefaultOnly
		}
		// resolution-mode changes lookup conditions, not the target's export
		// shape. Every other host-defined attribute remains conservative.
		if !hasType {
			for _, entry := range entries {
				if entry.Name != "resolution-mode" {
					return moduleViewUnknown
				}
			}
			return moduleViewAuthored
		}
		return moduleViewUnknown
	default:
		return moduleViewUnknown
	}
}

// HasAuthoredModuleView reports whether a request executes the target module's
// authored dependency and export graph. Attribute-selected data/text views and
// unknown host loaders do not expose that graph.
func HasAuthoredModuleView(source modules.Source) bool {
	return moduleViewFor(source.Attributes) == moduleViewAuthored
}

// HasDefaultOnlyModuleView reports the built-in data loaders whose request is
// known not to execute the target as a JavaScript module.
func HasDefaultOnlyModuleView(source modules.Source) bool {
	return moduleViewFor(source.Attributes) == moduleViewDefaultOnly
}

func defaultOnlyExportMap() *ExportMap {
	exports := newExportMap()
	exports.set(defaultExportName, nil)
	return exports
}
