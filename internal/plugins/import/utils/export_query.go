package utils

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// Only module-local facts belong here. Following re-exports for each requested
// name repeats graph work on wide barrels and on queries from different roots.
type exportPossibilities struct {
	names map[string]bool // Presence declares a name; true permits namespace metadata.
	star  bool
}

// MayHaveExport only rules out a definitely absent name. Possible exports and
// unknown edges still require GetExportMap for ordering, metadata and validity.
func MayHaveExport(ctx rule.RuleContext, source modules.Source, name string) bool {
	return mayHaveExport(ctx, source, name, false)
}

// MayHaveNamespaceExport is the same conservative query for namespace metadata.
func MayHaveNamespaceExport(ctx rule.RuleContext, source modules.Source, name string) bool {
	return mayHaveExport(ctx, source, name, true)
}

func mayHaveExport(ctx rule.RuleContext, source modules.Source, name string, namespaceOnly bool) bool {
	index := IndexFor(ctx)
	link := resolveExportLink(ctx.Program(), ctx.SourceFile, index.settings, source)
	if !link.Resolved {
		return true
	}
	if link.View == moduleViewDefaultOnly {
		return !namespaceOnly && name == defaultExportName
	}
	if link.Target == nil {
		return true
	}
	if index.cachedExportMap(link.Target) != nil {
		return true
	}
	possible := index.possibilities.Get(link.Target, func() exportPossibilities {
		return localExportPossibilities(index.localExportsOf(ctx.Program(), link.Target))
	})
	namespace, exists := possible.names[name]
	return possible.star && name != defaultExportName || exists && (!namespaceOnly || namespace)
}

func localExportPossibilities(local *localExports) exportPossibilities {
	possible := exportPossibilities{names: make(map[string]bool)}
	namespaces := make(map[string]bool)
	for _, binding := range local.Imports {
		if binding.NamespaceName != "" && binding.Link.Resolved {
			namespaces[binding.NamespaceName] = true
		}
	}
	add := func(name string, namespace bool) {
		// Later exports may overwrite earlier metadata in the full map. Keep
		// their union here: this filter only proves absence, never precedence.
		possible.names[name] = possible.names[name] || namespace
	}
	for _, step := range local.Steps {
		switch step.Kind {
		case exportStepNames:
			for _, name := range step.Names {
				add(name, false)
			}
		case exportStepLocalDefault:
			add(defaultExportName, namespaces[step.Local])
		case exportStepNamed:
			for _, spec := range step.Specs {
				namespace := namespaces[spec.LocalIdent]
				if step.FromModule {
					// The target decides whether this alias carries namespace
					// metadata. Leave that graph-dependent answer to ExportMap.
					namespace = !step.Link.Resolved || !spec.LocalOK || step.Link.View != moduleViewDefaultOnly
				}
				add(spec.Exported, namespace)
			}
		case exportStepStar:
			possible.star = possible.star || !step.Link.Resolved || step.Link.View != moduleViewDefaultOnly
		}
	}
	return possible
}
