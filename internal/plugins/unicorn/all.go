package unicorn_plugin

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/filename_case"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		filename_case.FilenameCaseRule,
	}
}
