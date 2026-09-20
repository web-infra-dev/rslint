package no_unpublished_bin

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

//go:embed no_unpublished_bin.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-bin.js
var NoUnpublishedBinRule = rule.Rule{
	Name:   "node/no-unpublished-bin",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		program := ctx.Program()
		if program == nil || ctx.SourceFile.FileName() == "<input>" {
			return nil
		}
		fileName := ctx.SourceFile.FileName()
		pkg := packagejson.FindNearestValid(program, fileName)
		if pkg == nil {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		relative := tspath.GetRelativePathFromDirectory(pkg.Directory(), fileName, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: program.FS().UseCaseSensitiveFileNames()})
		converted, ok := nodeutil.ConvertPath(relative, opts, ctx.Settings)
		if !ok {
			return nil
		}
		// The Program already supplies a normalized file path. Resolve again
		// only when conversion changed it; nodeutil owns publication policy.
		absolute := fileName
		if converted != relative {
			absolute = tspath.ResolvePath(pkg.Directory(), converted)
		}
		if !nodeutil.IsBinFile(program, pkg, absolute) || !nodeutil.IsUnpublished(program, pkg, absolute) {
			return nil
		}

		// Report the complete Program, including comments and empty files.
		ctx.ReportRange(ctx.SourceFile.AsNode().Loc, rule.RuleMessage{
			Id:          "invalidIgnored",
			Description: "npm ignores '" + converted + "'. Check 'files' field of 'package.json' or '.npmignore'.",
			Data:        map[string]string{"name": converted},
		})
		return nil
	},
}
