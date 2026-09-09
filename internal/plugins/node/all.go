package node_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/hashbang"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_exports_assign"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_extraneous_import"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_extraneous_require"
	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_process_exit"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		hashbang.HashbangRule,
		no_exports_assign.NoExportsAssignRule,
		no_extraneous_import.NoExtraneousImportRule,
		no_extraneous_require.NoExtraneousRequireRule,
		no_process_exit.NoProcessExitRule,
	}
}
