package no_unused_modules

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_unused_modules.schema.json
var schemaJSON []byte

// nativeExtensions is the default analyzed file set when import/extensions is
// not configured: every JavaScript and TypeScript extension rslint lints.
// Upstream analyzes only `.js`, `.mjs` and `.cjs` by default.
var nativeExtensions = []string{".js", ".mjs", ".cjs", ".jsx", ".ts", ".tsx", ".mts", ".cts"}

type options struct {
	src                     []string
	ignoreExports           []string
	missingExports          bool
	unusedExports           bool
	ignoreUnusedTypeExports bool
}

func parseOptions(raw []any) options {
	var opts options
	if len(raw) == 0 {
		return opts
	}
	values, _ := raw[0].(map[string]any)
	opts.src = utils.ToStringSlice(values["src"])
	opts.ignoreExports = utils.ToStringSlice(values["ignoreExports"])
	opts.missingExports = values["missingExports"] == true
	opts.unusedExports = values["unusedExports"] == true
	opts.ignoreUnusedTypeExports = values["ignoreUnusedTypeExports"] == true
	return opts
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-unused-modules.js
var NoUnusedModulesRule = rule.Rule{
	Name:   "import/no-unused-modules",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, raw []any) rule.RuleListeners {
		opts := parseOptions(raw)
		if !opts.missingExports && !opts.unusedExports {
			return nil
		}
		exports := collectModuleExports(ctx.SourceFile)
		if opts.unusedExports {
			checkUnusedExports(ctx, opts, exports)
		}
		if opts.missingExports && exports.count == 0 {
			reportMissingExports(ctx)
		}
		return nil
	},
}

func reportMissingExports(ctx rule.RuleContext) {
	message := rule.RuleMessage{Id: "noExports", Description: "No exports found"}
	if statements := ctx.SourceFile.Statements.Nodes; len(statements) > 0 {
		ctx.ReportNode(statements[0], message)
		return
	}
	// A file of only comments reports its empty Program, which ESTree places
	// after the leading comments.
	end := ctx.SourceFile.End()
	start := scanner.SkipTrivia(ctx.SourceFile.Text(), ctx.SourceFile.EndOfFileToken.Pos())
	ctx.ReportRange(core.NewTextRange(min(start, end), end), message)
}

func checkUnusedExports(ctx rule.RuleContext, opts options, exports moduleExports) {
	if len(exports.checked) == 0 {
		return
	}
	fileName := ctx.SourceFile.FileName()
	cwd := ctx.ProcessCurrentDirectory()
	if cwd == "" {
		cwd = ctx.Program().CurrentDirectory()
	}
	src := opts.src
	if len(src) == 0 {
		// Upstream's default is process.cwd().
		src = []string{cwd}
	}
	opts.src = src
	extensions := nativeExtensions
	if ctx.Settings["import/extensions"] != nil {
		extensions = import_utils.FileExtensions(ctx.Settings)
	}
	// Without a pass-wide scope (editor passes), project usage is unknown;
	// reporting from the current Program alone would mislead.
	index, ok := rule.CachedByGeneration(ctx, newIndexKey(cwd, opts, extensions, ctx.Settings), func(scope *rule.GenerationScope) *projectIndex {
		return buildProjectIndex(scope, indexInput{
			fs:         ctx.Program().FS(),
			cwd:        cwd,
			opts:       opts,
			extensions: extensions,
			settings:   ctx.Settings,
			compiled:   import_utils.SettingsFor(ctx),
			fallback:   ctx.Program(),
		})
	})
	if !ok || !index.analyzes(fileName) || index.ignored(fileName) || isPackageEntry(ctx.Program(), fileName) {
		return
	}
	used := index.usageOf(fileName)
	for _, export := range exports.checked {
		if export.isType && opts.ignoreUnusedTypeExports || used.uses(export.name) {
			continue
		}
		reportUnused(ctx, export)
	}
}

func reportUnused(ctx rule.RuleContext, export localExport) {
	ctx.ReportNode(export.node, rule.RuleMessage{
		Id:          "unused",
		Description: fmt.Sprintf("exported declaration '%s' not used within other modules", export.name),
	})
}
