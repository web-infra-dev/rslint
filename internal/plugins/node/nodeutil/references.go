package nodeutil

import (
	"maps"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/referencetracker"
)

// ReferenceTrace describes the reads, calls and constructors a Node rule needs.
// The traversal and event contract are shared with other rule plugins.
type ReferenceTrace = referencetracker.Trace

// ReferenceTracker adds Node module entry points to the shared API tracker.
// Module names, builtin aliases and legacy ESM behavior stay in this adapter.
type ReferenceTracker struct {
	*referencetracker.Tracker
	ctx             rule.RuleContext
	moduleEvaluator *utils.StaticStringEvaluator
}

func NewReferenceTracker(ctx rule.RuleContext) *ReferenceTracker {
	return &ReferenceTracker{Tracker: referencetracker.New(ctx), ctx: ctx}
}

func (tracker *ReferenceTracker) constantString(node *ast.Node) (string, bool) {
	if tracker.moduleEvaluator == nil {
		tracker.moduleEvaluator = utils.NewStaticStringEvaluatorWithoutScope()
	}
	return tracker.moduleEvaluator.EvalToString(node)
}

func readReference(trace *ReferenceTrace, node *ast.Node) {
	if trace.Read != nil {
		trace.Read(node)
	}
}

// TrackModules follows CommonJS, process.getBuiltinModule and legacy ESM imports.
// Builtin node: aliases use the same trace, retaining canonical callback names.
func (tracker *ReferenceTracker) TrackModules(modules map[string]*ReferenceTrace) {
	lookup := func(name string) *ReferenceTrace {
		if strings.HasPrefix(name, "node:") {
			if !isNodeBuiltin(name) {
				return nil
			}
			name = strings.TrimPrefix(name, "node:")
		}
		return modules[name]
	}
	load := func(node *ast.Node) {
		args := node.AsCallExpression().Arguments
		if args == nil || len(args.Nodes) == 0 {
			return
		}
		name, ok := tracker.constantString(args.Nodes[0])
		if trace := lookup(name); ok && trace != nil {
			readReference(trace, node)
			tracker.TrackExpression(node, trace)
		}
	}
	tracker.TrackGlobals(map[string]*ReferenceTrace{"require": {Call: load}})
	tracker.TrackGlobals(map[string]*ReferenceTrace{"process": {Properties: map[string]*ReferenceTrace{
		"getBuiltinModule": {Call: load},
	}}})
	for _, node := range tracker.ctx.SourceFile.Statements.Nodes {
		if node.Kind != ast.KindImportDeclaration && node.Kind != ast.KindExportDeclaration {
			continue
		}
		source := ast.GetExternalModuleName(node)
		if source == nil || source.Kind != ast.KindStringLiteral {
			continue
		}
		trace := lookup(source.Text())
		if trace == nil {
			continue
		}
		readReference(trace, node)
		// In legacy mode a namespace's default property aliases the CommonJS
		// export. Reading that alias must not report the module a second time.
		moduleValue := &ReferenceTrace{Properties: trace.Properties, Call: trace.Call, Construct: trace.Construct}
		properties := maps.Clone(trace.Properties)
		if properties == nil {
			properties = map[string]*ReferenceTrace{}
		}
		properties["default"] = moduleValue
		if node.Kind == ast.KindImportDeclaration {
			for _, binding := range utils.GetImportBindingNodes(node) {
				value := moduleValue
				switch binding.Parent.Kind {
				case ast.KindNamespaceImport:
					value = &ReferenceTrace{Properties: properties}
				case ast.KindImportSpecifier:
					imported := binding.Parent.PropertyName()
					if imported == nil {
						imported = binding
					}
					value = properties[imported.Text()]
					if value == nil {
						continue
					}
					readReference(value, binding.Parent)
				}
				tracker.TrackBinding(binding, value)
			}
		} else {
			clause := node.AsExportDeclaration().ExportClause
			if clause == nil || clause.Kind == ast.KindNamespaceExport {
				for _, name := range slices.Sorted(maps.Keys(trace.Properties)) {
					readReference(trace.Properties[name], node)
				}
			} else {
				for _, specifier := range clause.AsNamedExports().Elements.Nodes {
					local := specifier.AsExportSpecifier().PropertyName
					if local == nil {
						local = specifier.Name()
					}
					if next := properties[local.Text()]; next != nil {
						readReference(next, specifier)
					}
				}
			}
		}
	}
}
