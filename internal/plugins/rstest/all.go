package rstest

import (
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_identical_title"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_mocks_import"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func GetAllRules() []rule.Rule {
	return []rule.Rule{
		no_identical_title.NoIdenticalTitleRule,
		no_mocks_import.NoMocksImportRule,
	}
}
