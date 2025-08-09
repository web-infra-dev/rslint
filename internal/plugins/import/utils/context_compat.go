package utils

import "github.com/web-infra-dev/rslint/internal/rule"

// https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/utils/contextCompat.js#L37
func GetPhysicalFilename(ctx rule.RuleContext) string {
	return ctx.SourceFile.FileName()
}
