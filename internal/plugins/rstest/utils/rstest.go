package utils

import (
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
)

// RstestImportModule is the module Rstest test globals are imported from.
const RstestImportModule = "@rstest/core"

// Run-mode / concurrency modifiers shared by the Rstest `test` and `describe`
// APIs, mirroring SHARED_RUN_MODIFIERS in
// rstest/packages/core/src/runtime/runner/runtime.ts.
var rstestSharedModifiers = map[string]bool{
	"only":       true,
	"todo":       true,
	"skip":       true,
	"concurrent": true,
	"sequential": true,
}

// Conditional and parameterized members available on both `test` and
// `describe`. Each of these is installed as a factory that returns a new test
// API which is then invoked again (test.runIf(cond)(...), test.each(cases)(...)),
// so they all form a "factory call + actual call" shape.
var rstestParameterizedMembers = map[string]bool{
	"runIf":  true,
	"skipIf": true,
	"each":   true,
	"for":    true,
}

// isValidRstestMember reports whether a chained member is legal for the given
// Rstest root. `fails` is only available on `test`/`it`, not `describe`.
func isValidRstestMember(root string, member string) bool {
	if rstestSharedModifiers[member] || rstestParameterizedMembers[member] {
		return true
	}
	if member == "fails" && root != "describe" {
		return true
	}
	return false
}

// isValidRstestCall validates a resolved Rstest test/describe call chain. Rstest
// installs its run-mode modifiers as chainable getters, so any ordering of the
// allowed members is legal (e.g. `test.concurrent.only`, `test.only.for`).
func isValidRstestCall(name string, members []string) bool {
	switch name {
	case "test", "it", "describe":
	default:
		// hooks (beforeAll, ...) take no members.
		return len(members) == 0
	}

	for _, member := range members {
		if !isValidRstestMember(name, member) {
			return false
		}
	}
	return true
}

// RstestFnCallParseConfig returns the shared function-call parser config for
// Rstest.
func RstestFnCallParseConfig() jestUtils.FnCallParseConfig {
	return jestUtils.FnCallParseConfig{
		ImportModule:           RstestImportModule,
		IsValidChain:           isValidRstestCall,
		ParameterizedModifiers: rstestParameterizedMembers,
	}
}
