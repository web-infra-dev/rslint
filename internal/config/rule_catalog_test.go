package config

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
)

func baseRuleCatalog() *rule.Catalog {
	return rules.All()
}
