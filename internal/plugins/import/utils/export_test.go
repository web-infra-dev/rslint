package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// CompileModuleSettings gives external tests the compiled form of raw
// `import/` settings, which production callers reach through SettingsFor.
var CompileModuleSettings = compileModuleSettings

// HasCachedExportMap reports whether file currently has a cached ExportMap in
// index, without computing one.
func HasCachedExportMap(index *ModuleIndex, file *ast.SourceFile) bool {
	return index.cachedExportMap(file) != nil
}

// ExportMapBuildStats records internal build and propagation counts for a
// single GetExportMap query.
type ExportMapBuildStats struct {
	BuildCount           int
	NamedSpecifierChecks int
	SCCQueueEvents       int
	OriginStepScans      int
}

// GetExportMapWithBuildCount runs GetExportMap and returns how many modules
// actually built an ExportMap during the query.
func GetExportMapWithBuildCount(ctx rule.RuleContext, source modules.Source) (*ExportMap, int, bool) {
	exports, stats, ok := GetExportMapWithStats(ctx, source)
	return exports, stats.BuildCount, ok
}

// GetExportMapWithStats runs GetExportMap and returns internal build and
// named-specifier propagation counts for the query.
func GetExportMapWithStats(ctx rule.RuleContext, source modules.Source) (*ExportMap, ExportMapBuildStats, bool) {
	if !ctx.Program().IsValid() || ctx.SourceFile == nil {
		return nil, ExportMapBuildStats{}, false
	}
	builder := newExportBuilder(IndexFor(ctx), ctx.Program())
	exports, ok := getExportMap(ctx.SourceFile, source, builder)
	return exports, ExportMapBuildStats{
		BuildCount:           builder.buildCount,
		NamedSpecifierChecks: builder.namedSpecifierChecks,
		SCCQueueEvents:       builder.sccQueueEvents,
		OriginStepScans:      builder.originStepScans,
	}, ok
}
