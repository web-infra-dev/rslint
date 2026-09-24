package import_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/consistent_type_specifier_style"
	default_rule "github.com/web-infra-dev/rslint/internal/plugins/import/rules/default"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/export"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/first"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/max_dependencies"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/namespace"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/newline_after_import"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_amd"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_anonymous_default_export"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_commonjs"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_default_export"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_duplicates"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_dynamic_require"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_mutable_exports"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_as_default"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_namespace"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_restricted_paths"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_self_import"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_unresolved"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_webpack_loader_syntax"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/order"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		consistent_type_specifier_style.ConsistentTypeSpecifierStyleRule,
		default_rule.DefaultRule,
		export.ExportRule,
		first.FirstRule,
		max_dependencies.MaxDependenciesRule,
		namespace.NamespaceRule,
		newline_after_import.NewlineAfterImportRule,
		no_amd.NoAmdRule,
		no_anonymous_default_export.NoAnonymousDefaultExportRule,
		no_commonjs.NoCommonjsRule,
		no_cycle.NoCycleRule,
		no_default_export.NoDefaultExportRule,
		no_duplicates.NoDuplicatesRule,
		no_dynamic_require.NoDynamicRequireRule,
		no_mutable_exports.NoMutableExportsRule,
		no_named_as_default.NoNamedAsDefaultRule,
		no_namespace.NoNamespaceRule,
		no_restricted_paths.NoRestrictedPathsRule,
		no_self_import.NoSelfImportRule,
		no_unresolved.NoUnresolvedRule,
		no_webpack_loader_syntax.NoWebpackLoaderSyntax,
		order.OrderRule,
	}
}
