package program

import (
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// CompilerOptionsSupportFileName reports whether a TypeScript project can
// admit fileName through normal source-file resolution. Direct config roots do
// not need this guard because ParsedCommandLine.FileNames is authoritative.
func CompilerOptionsSupportFileName(options *core.CompilerOptions, fileName string) bool {
	if options == nil {
		return false
	}
	extensions := tspath.SupportedTSExtensions
	if options.GetAllowJS() {
		extensions = tspath.AllSupportedExtensions
	}
	extensions = tsoptions.GetSupportedExtensionsWithJsonIfResolveJsonModule(
		options,
		extensions,
	)
	for _, group := range extensions {
		if tspath.FileExtensionIsOneOf(fileName, group) {
			return true
		}
	}
	return false
}
