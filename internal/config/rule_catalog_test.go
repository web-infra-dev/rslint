package config

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules/catalog"
)

func nativeRuleCatalog() *rule.Catalog {
	return catalog.Native()
}
