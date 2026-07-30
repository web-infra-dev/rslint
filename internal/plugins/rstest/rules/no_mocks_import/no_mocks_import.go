package no_mocks_import

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	sharedNoMocksImport "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_mocks_import"
)

// NewRule creates a no-mocks-import rule for a test framework.
func NewRule(name string, mockFunction string) rule.Rule {
	return sharedNoMocksImport.NewRule(name, mockFunction)
}

var NoMocksImportRule = NewRule("rstest/no-mocks-import", "rs.mock")
