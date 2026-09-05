package utils

// RstestGlobals is the API surface injected by Rstest when `globals` is
// enabled. The source of truth is globalApiList in
// packages/core/src/utils/constants.ts, whose compile-time exhaustiveness
// check keeps it aligned with the public Rstest API.
var RstestGlobals = map[string]struct{}{
	"test":           {},
	"describe":       {},
	"it":             {},
	"expect":         {},
	"afterAll":       {},
	"afterEach":      {},
	"beforeAll":      {},
	"beforeEach":     {},
	"rstest":         {},
	"rs":             {},
	"assert":         {},
	"onTestFinished": {},
	"onTestFailed":   {},
}

func IsRstestGlobal(name string) bool {
	_, ok := RstestGlobals[name]
	return ok
}
