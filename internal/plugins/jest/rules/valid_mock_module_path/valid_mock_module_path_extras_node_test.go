// cspell:ignore Dstar eode
package valid_mock_module_path

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// nodeResolution lists package and imports requests with whether Node's
// require.resolve (v26.5.1) finds a file for them in the fixture tree, run from
// the directory of testFile. A request Node rejects with an error other than a
// missing module, noted beside it, is reported as well.
var nodeResolution = []struct {
	request  string
	resolves bool
}{
	// Pattern matches are substituted into the target before it is read as a URL:
	// `?` and `#` end the path, `\` separates segments, escapes are decoded.
	{`"m-star/entry.js"`, true},
	{`"m-star/entry.js?raw"`, true},
	{`"m-star/entry.js#hash"`, true},
	{`"m-star/entry.js?"`, true},
	{`"m-star/entry.js#"`, true},
	{`"m-star/entry.js%23hash"`, true},
	{`"m-star/entry.js?a#b"`, true},
	{`"m-star/entry.js#a?b"`, true},
	{`"m-star/a b.js"`, true},
	{`"m-star/a%20b.js"`, true},
	{`"m-star/sub\\x.js"`, true},
	{`"m-star/sub/x.js"`, true},
	{`"m-star/sub%2Fx.js"`, false},        // ERR_INVALID_MODULE_SPECIFIER
	{`"m-star/entry.js?%2F"`, false},      // ERR_INVALID_MODULE_SPECIFIER
	{`"m-star/entry.js#%5C"`, false},      // ERR_INVALID_MODULE_SPECIFIER
	{`"m-star/./entry.js"`, false},        // ERR_INVALID_MODULE_SPECIFIER
	{`"m-star/sub/../entry.js"`, false},   // ERR_INVALID_MODULE_SPECIFIER
	{`"m-star/%2e/entry.js"`, false},      // ERR_INVALID_MODULE_SPECIFIER
	{`"m-star/node_modules/x.js"`, false}, // ERR_INVALID_MODULE_SPECIFIER
	{`"m-star//entry.js"`, true},
	{`"m-star-js/entry"`, true},
	{`"m-star-js/entry?raw"`, false},
	{`"m-star-js/entry#hash"`, false},
	{`"m-star-js/entry%23hash"`, true},
	{`"m-star-js/sub\\x"`, true},
	{`"m-multi-star/x"`, true},
	{`"m-target-q-star/entry"`, true},
	{`"s-star/entry%3Fx.js"`, true},
	{`"s-star/sub/x.js/"`, false},
	{`"s-star/%2E/entry.js"`, false},        // ERR_INVALID_MODULE_SPECIFIER
	{`"s-star/%2E%2E/lib/entry.js"`, false}, // ERR_INVALID_MODULE_SPECIFIER
	{`"s-star/%6eode_modules/x"`, false},    // ERR_INVALID_MODULE_SPECIFIER
	{`"s-star/x"`, true},
	{`"s-star/x?raw"`, false},
	{`"s-star/entry.js?/../x"`, false}, // ERR_INVALID_MODULE_SPECIFIER
	{`"s-star/entry.js?%2e%2e/"`, true},
	{`"s-star/a%20b.js?q"`, true},
	{`"s-star/sub%5Cx.js"`, false}, // ERR_INVALID_MODULE_SPECIFIER
	{`"s-multi/x"`, true},
	{`"s-multi/x?y"`, false},
	{`"@s/pkg/entry.js"`, true},
	{`"@s/pkg/entry.js?raw"`, true},
	{`"@s/pkg/entry.js#x"`, true},
	{`"@s\\pkg/entry.js"`, false},
	{`"@s/pkg\\entry.js"`, false},
	{`"cond/entry.js?raw"`, true},
	{`"cond/entry.js#x"`, true},
	{`"fixture-root/entry.js"`, true},
	{`"fixture-root/entry.js?raw"`, true},
	{`"fixture-root/sub\\x.js"`, true},
	{`"m-star\\entry.js"`, false},
	{`"m%2Dstar/entry.js"`, false},
	// Targets are validated as Node validates them; invalid entries are skipped
	// in fallback arrays.
	{`"m-target-back"`, true},
	{`"m-target-double-slash"`, true},
	{`"m-target-parent-segment"`, false}, // ERR_INVALID_PACKAGE_TARGET
	{`"m-target-encoded-parent"`, false}, // ERR_INVALID_PACKAGE_TARGET
	{`"m-target-nm"`, false},             // ERR_INVALID_PACKAGE_TARGET
	{`"m-arr-parent-segment"`, true},
	{`"m-arr-encoded-parent"`, true},
	{`"m-arr-nm"`, true},
	{`"m-arr-parent"`, true},
	{`"t-query-separator"`, false},     // ERR_INVALID_MODULE_SPECIFIER
	{`"t-fragment-separator"`, false},  // ERR_INVALID_MODULE_SPECIFIER
	{`"t-arr-query-separator"`, false}, // ERR_INVALID_MODULE_SPECIFIER
	{`"t-upper-dot"`, false},           // ERR_INVALID_PACKAGE_TARGET
	{`"t-enc-nm"`, false},              // ERR_INVALID_PACKAGE_TARGET
	{`"t-bad-utf8"`, false},            // URIError
	{`"t-trunc-utf8"`, false},          // URIError
	{`"t-cond-bad"`, true},
	// A request ending in `/` names no file.
	{`"s-trailing/"`, false},  // ERR_PACKAGE_PATH_NOT_EXPORTED
	{`"s-trailing/."`, false}, // ERR_PACKAGE_PATH_NOT_EXPORTED
	{`"s-trailing?x"`, false},
	// Without an exports map, the request is a plain path: no URL decoding.
	{`"u-legacy/lib/a b.js"`, true},
	{`"u-legacy/lib/a%20b.js"`, false},
	{`"u-legacy/lib/c%20d.js"`, true},
	{`"u-legacy/lib/q.js?x"`, false},
	{`"u-legacy/lib/q.js#x"`, false},
	{`"u-legacy//lib/q.js"`, true},
	{`"u-legacy/lib\\q.js"`, false},
	{`"u-legacy\\lib\\q.js"`, false},
	// Imports maps, including patterns and bare package targets.
	{`"#p/entry.js?x"`, true},
	{`"#p/entry.js#x"`, true},
	{`"#p/a%20b.js"`, true},
	{`"#p/sub\\x.js"`, true},
	{`"#b/entry.js?x"`, true},
	{`"#b/entry.js"`, true},
	{`"#b/x?y"`, false},
	{`"#bad"`, false}, // ERR_INVALID_PACKAGE_TARGET
	{`"#arr"`, true},
	{`"#p/sub/../entry.js"`, false}, // ERR_INVALID_MODULE_SPECIFIER
	{`"#p/%2e/entry.js"`, false},    // ERR_INVALID_MODULE_SPECIFIER
	{`"#p//entry.js"`, true},
	{`"#p/entry.js?%2F"`, false}, // ERR_INVALID_MODULE_SPECIFIER
	{`"#dep2"`, true},
	{`"#dep3"`, true},
	{`"#b/sub\\x.js"`, true},
	{`"#b/entry.js#y"`, true},
}

func TestValidMockModulePathNodeResolution(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, entry := range nodeResolution {
		code := "jest.mock(" + entry.request + ")"
		if entry.resolves {
			valid = append(valid, rule_tester.ValidTestCase{Code: code, FileName: testFile})
		} else {
			invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, FileName: testFile, Errors: invalidAt(1)})
		}
	}
	rule_tester.RunRuleTester(mockModuleRoot(t), "tsconfig.json", t, &ValidMockModulePathRule, valid, invalid)
}
