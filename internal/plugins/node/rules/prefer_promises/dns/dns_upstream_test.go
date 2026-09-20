// cspell:ignore Naptr
package dns_test

import (
	"maps"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/prefer_promises/dns"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func dnsError(name string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	messageID, call := "preferPromises", "dns.promises."+name+"()"
	if name == "Resolver" {
		messageID, call = "preferPromisesNew", "new "+call
	}
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID, Message: "Use '" + call + "' instead.",
		Line: line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func runDNSTests(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	t.Helper()
	nodeGlobals := func(overrides map[string]any) map[string]any {
		globals := map[string]any{"require": "readonly", "process": "readonly", "global": "readonly"}
		maps.Copy(globals, overrides)
		return globals
	}
	for i := range valid {
		if valid[i].FileName == "" {
			valid[i].FileName = "input.js"
		}
		valid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
		valid[i].Globals = nodeGlobals(valid[i].Globals)
	}
	for i := range invalid {
		if invalid[i].FileName == "" {
			invalid[i].FileName = "input.js"
		}
		invalid[i].LanguageOptions = rule.LanguageOptions{SourceType: "module"}
		invalid[i].Globals = nodeGlobals(invalid[i].Globals)
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allowJs.json", t, &dns.PreferPromisesDNSRule, valid, invalid)
}

// Every test from eslint-plugin-n v18.3.0, with exact upstream diagnostics.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/prefer-promises/dns.js
func TestPreferPromisesDNSUpstream(t *testing.T) {
	runDNSTests(t,
		[]rule_tester.ValidTestCase{
			{Code: "const dns = require('dns'); dns.lookupSync()"},
			{Code: "const dns = require('dns'); dns.promises.lookup()"},
			{Code: "const dns = require('node:dns'); dns.promises.lookup()"},
			{Code: "const {promises} = require('dns'); promises.lookup()"},
			{Code: "const {promises: dns} = require('dns'); dns.lookup()"},
			{Code: "const {promises: {lookup}} = require('dns'); lookup()"},
			{Code: "import dns from 'dns'; dns.promises.lookup()"},
			{Code: "import dns from 'node:dns'; dns.promises.lookup()"},
			{Code: "import * as dns from 'dns'; dns.promises.lookup()"},
			{Code: "import {promises} from 'dns'; promises.lookup()"},
			{Code: "import {promises as dns} from 'dns'; dns.lookup()"},
			{Code: "const dns = process.getBuiltinModule('dns'); dns.promises.lookup()"},
			{Code: "const dns = process.getBuiltinModule('node:dns'); dns.promises.lookup()"},
			{Code: "const {promises} = process.getBuiltinModule('dns'); promises.lookup()"},
			{Code: "const {promises: dns} = process.getBuiltinModule('dns'); dns.lookup()"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const dns = require('dns'); dns.lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 29, 1, 41),
				},
			},
			{Code: "const dns = require('node:dns'); dns.lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 34, 1, 46),
				},
			},
			{Code: "const {lookup} = require('dns'); lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 34, 1, 42),
				},
			},
			{Code: "import dns from 'dns'; dns.lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 24, 1, 36),
				},
			},
			{Code: "import dns from 'node:dns'; dns.lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 29, 1, 41),
				},
			},
			{Code: "import * as dns from 'dns'; dns.lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 29, 1, 41),
				},
			},
			{Code: "import {lookup} from 'dns'; lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 29, 1, 37),
				},
			},
			{Code: "const dns = process.getBuiltinModule('dns'); dns.lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 46, 1, 58),
				},
			},
			{Code: "const dns = process.getBuiltinModule('node:dns'); dns.lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 51, 1, 63),
				},
			},
			{Code: "const {lookup} = process.getBuiltinModule('dns'); lookup()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 51, 1, 59),
				},
			},
			// Upstream: other DNS members.
			{Code: "const dns = require('dns'); dns.lookupService()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookupService", 1, 29, 1, 48),
				},
			},
			{Code: "const dns = require('dns'); new dns.Resolver()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("Resolver", 1, 29, 1, 47),
				},
			},
			{Code: "const dns = require('dns'); dns.getServers()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("getServers", 1, 29, 1, 45),
				},
			},
			{Code: "const dns = require('dns'); dns.resolve()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolve", 1, 29, 1, 42),
				},
			},
			{Code: "const dns = require('dns'); dns.resolve4()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolve4", 1, 29, 1, 43),
				},
			},
			{Code: "const dns = require('dns'); dns.resolve6()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolve6", 1, 29, 1, 43),
				},
			},
			{Code: "const dns = require('dns'); dns.resolveAny()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolveAny", 1, 29, 1, 45),
				},
			},
			{Code: "const dns = require('dns'); dns.resolveCname()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolveCname", 1, 29, 1, 47),
				},
			},
			{Code: "const dns = require('dns'); dns.resolveMx()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolveMx", 1, 29, 1, 44),
				},
			},
			{Code: "const dns = require('dns'); dns.resolveNaptr()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolveNaptr", 1, 29, 1, 47),
				},
			},
			{Code: "const dns = require('dns'); dns.resolveNs()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolveNs", 1, 29, 1, 44),
				},
			},
			{Code: "const dns = require('dns'); dns.resolvePtr()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolvePtr", 1, 29, 1, 45),
				},
			},
			{Code: "const dns = require('dns'); dns.resolveSoa()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolveSoa", 1, 29, 1, 45),
				},
			},
			{Code: "const dns = require('dns'); dns.resolveSrv()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolveSrv", 1, 29, 1, 45),
				},
			},
			{Code: "const dns = require('dns'); dns.resolveTxt()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("resolveTxt", 1, 29, 1, 45),
				},
			},
			{Code: "const dns = require('dns'); dns.reverse()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("reverse", 1, 29, 1, 42),
				},
			},
			{Code: "const dns = require('dns'); dns.setServers()",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("setServers", 1, 29, 1, 45),
				},
			},
		},
	)
}

// All four executable documentation examples; rule-enabling comments are omitted.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/prefer-promises/dns.md
func TestPreferPromisesDNSDocumentation(t *testing.T) {
	runDNSTests(t,
		[]rule_tester.ValidTestCase{
			// Documentation example 3.
			{Code: "const { promises: dns } = require(\"dns\")\n\nasync function lookup(hostname) {\n    const { address, family } = await dns.lookup(hostname)\n    //...\n}"},
			// Documentation example 4.
			{Code: "import { promises as dns } from \"dns\"\n\nasync function lookup(hostname) {\n    const { address, family } = await dns.lookup(hostname)\n    //...\n}"},
		},
		[]rule_tester.InvalidTestCase{
			// Documentation example 1.
			{Code: "const dns = require(\"dns\")\n\nfunction lookup(hostname) {\n    dns.lookup(hostname, (error, address, family) => {\n        //...\n    })\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 4, 5, 6, 7),
				},
			},
			// Documentation example 2.
			{Code: "import dns from \"dns\"\n\nfunction lookup(hostname) {\n    dns.lookup(hostname, (error, address, family) => {\n        //...\n    })\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 4, 5, 6, 7),
				},
			},
		},
	)
}
