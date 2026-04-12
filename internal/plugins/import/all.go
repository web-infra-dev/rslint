package import_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/first"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/newline_after_import"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_duplicates"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_mutable_exports"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_self_import"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_webpack_loader_syntax"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		first.FirstRule,
		newline_after_import.NewlineAfterImportRule,
		no_duplicates.NoDuplicatesRule,
		no_mutable_exports.NoMutableExportsRule,
		no_self_import.NoSelfImportRule,
		no_webpack_loader_syntax.NoWebpackLoaderSyntax,
	}
}
