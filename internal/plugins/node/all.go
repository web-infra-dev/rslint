package node_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/callback_return"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/exports_style"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/global_require"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/handle_callback_err"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/hashbang"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_exports_assign"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_extraneous_import"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_extraneous_require"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_missing_import"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_mixed_requires"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_new_require"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_process_exit"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		callback_return.CallbackReturnRule,
		exports_style.ExportsStyleRule,
		global_require.GlobalRequireRule,
		handle_callback_err.HandleCallbackErrRule,
		hashbang.HashbangRule,
		no_exports_assign.NoExportsAssignRule,
		no_extraneous_import.NoExtraneousImportRule,
		no_extraneous_require.NoExtraneousRequireRule,
		no_missing_import.NoMissingImportRule,
		no_mixed_requires.NoMixedRequiresRule,
		no_new_require.NoNewRequireRule,
		no_process_exit.NoProcessExitRule,
	}
}
