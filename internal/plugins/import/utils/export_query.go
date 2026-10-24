package utils

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

type exportQuery struct {
	file          *ast.SourceFile
	name          string
	namespaceOnly bool
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
	key := exportQuery{file: link.Target, name: name, namespaceOnly: namespaceOnly}
	return index.possibilities.Get(key, func() bool {
		return index.mayHaveExport(ctx.Program(), link.Target, name, namespaceOnly, make(map[exportKey]bool))
	})
}

func (index *ModuleIndex) mayHaveExport(sourceProgram *program.Program, file *ast.SourceFile, name string, namespaceOnly bool, seen map[exportKey]bool) bool {
	if file == nil {
		return true
	}
	key := exportKey{file: file, name: name}
	if seen[key] {
		return false
	}
	seen[key] = true
	local := index.localExportsOf(sourceProgram, file)
	namespaceBinding := func(localName string) bool {
		if localName != "" {
			for _, binding := range local.Imports {
				if binding.NamespaceName == localName && binding.Link.Resolved {
					return true
				}
			}
		}
		return false
	}
	for _, step := range local.Steps {
		switch step.Kind {
		case exportStepNames:
			if !namespaceOnly && slices.Contains(step.Names, name) {
				return true
			}
		case exportStepLocalDefault:
			if name == defaultExportName && (!namespaceOnly || namespaceBinding(step.Local)) {
				return true
			}
		case exportStepNamed:
			for _, spec := range step.Specs {
				if spec.Exported != name {
					continue
				}
				if !namespaceOnly {
					return true
				}
				if !step.FromModule {
					if namespaceBinding(spec.LocalIdent) {
						return true
					}
				} else if !step.Link.Resolved || !spec.LocalOK {
					return true
				} else if step.Link.View != moduleViewDefaultOnly && index.mayHaveExport(sourceProgram, step.Link.Target, spec.Local, namespaceOnly, seen) {
					return true
				}
			}
		case exportStepStar:
			if name == defaultExportName {
				continue
			}
			if !step.Link.Resolved {
				return true
			}
			if step.Link.View != moduleViewDefaultOnly && index.mayHaveExport(sourceProgram, step.Link.Target, name, namespaceOnly, seen) {
				return true
			}
		}
	}
	return false
}
