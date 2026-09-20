// cspell:ignore fokup
package dns_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional scope and AST cases checked against eslint-plugin-n v18.3.0.
func TestPreferPromisesDNSExtras(t *testing.T) {
	runDNSTests(t,
		[]rule_tester.ValidTestCase{
			// Promise-only modules and promise Resolver are allowed.
			{Code: "const dns = require('dns/promises'); dns.lookup(); new dns.Resolver(); import other from 'node:dns/promises'; other.resolve4();"},
			// Reads, exports, method construction and Resolver calls are not selected.
			{Code: "import dns, {lookup, Resolver} from 'dns'; dns.lookup; dns.Resolver; new lookup(); Resolver(); export {lookup} from 'dns'; export * from 'dns'; dns.lookup.call(null);"},
			// Other modules and unrelated local objects are ignored.
			{Code: "const dns = require('./dns'); dns.lookup(); new dns.Resolver(); const other = {lookup() {}}; other.lookup();"},
			// Shadowed loader roots are ignored.
			{Code: "function f(require, process) { require('dns').lookup(); process.getBuiltinModule('dns').resolve(); }"},
			// Writes disable global roots throughout the file.
			{Code: "require('dns').lookup(); require = other; process.getBuiltinModule('dns').lookup(); process = other;"},
			// Disabled globals do not seed tracking.
			{Code: "require('dns').lookup(); process.getBuiltinModule('dns').lookup();",
				Globals: map[string]any{"require": "off", "process": "off"},
			},
			// Dynamic module and property names are not resolved through variables.
			{Code: "const name = 'dns'; require(name).lookup(); const dns = require('dns'); const key = 'lookup'; dns[key](); require().lookup(); process.getBuiltinModule().lookup();"},
			// Private keys are not DNS APIs.
			{Code: "import dns from 'dns'; class C { #lookup; #Resolver; f() { dns.#lookup(); new dns.#Resolver(); } }"},
			// Dynamic imports, rest patterns and member writes do not create aliases.
			{Code: "const dns = await import('dns'); dns.lookup(); const {...rest} = require('dns'); rest.lookup(); const holder = {}; holder.dns = require('dns'); holder.dns.lookup();"},
			// JSX tags and type queries are not calls.
			{Code: "import dns from 'dns'; type Lookup = typeof dns.lookup; const el = <dns.lookup />;",
				FileName: "input.tsx",
			},
			// TypeScript import-equals does not create an ESM reference.
			{Code: "import dns = require('dns'); dns.lookup();",
				FileName: "input.ts",
			},
			// A scoped process import is not a global loader.
			{Code: "import process from 'node:process'; process.getBuiltinModule('dns').lookup();"},
			// TypeScript assertion assignment targets are not aliases.
			{Code: "let dns; (dns as any) = require('dns'); dns.lookup();",
				FileName: "input.ts",
			},
			// Ambient and namespace declarations shadow global loaders.
			{Code: "declare const require: any; require('dns').lookup(); namespace process {} process.getBuiltinModule('dns').lookup();",
				FileName: "input.ts",
			},
			// Type-only imports do not make executable calls into type syntax.
			{Code: "import type {Resolver} from 'dns'; type R = InstanceType<typeof Resolver>;",
				FileName: "input.ts",
			},
			// Dynamic template expressions and spread module arguments do not match.
			{Code: "const name = 'dns'; require(`${name}`).lookup(); require(...['dns']).lookup(); process.getBuiltinModule(...['dns']).lookup();"},
		},
		[]rule_tester.InvalidTestCase{
			// Inline module access and parenthesized calls.
			{Code: "(require('dns').lookup)('example.com', cb); (require('node:dns')).resolve4('example.com', cb);",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 1, 1, 43),
					dnsError("resolve4", 1, 45, 1, 94),
				},
			},
			// Aliased require and process loaders.
			{Code: "const load = require; load('dns').lookup(); const {getBuiltinModule: builtin} = process; builtin('node:dns').reverse();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 23, 1, 43),
					dnsError("reverse", 1, 90, 1, 119),
				},
			},
			// Destructured ESM and namespace default aliases.
			{Code: "import {lookup as lookupHost} from 'node:dns'; lookupHost(); import * as dns from 'dns'; const {default: api} = dns; api.lookupService();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 48, 1, 60),
					dnsError("lookupService", 1, 118, 1, 137),
				},
			},
			// Constructor aliases, parentheses and omitted arguments.
			{Code: "import {Resolver as R} from 'node:dns'; new (R); const {Resolver} = require('dns'); new Resolver();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("Resolver", 1, 41, 1, 48),
					dnsError("Resolver", 1, 85, 1, 99),
				},
			},
			// Static computed strings, templates and destructuring.
			{Code: "const dns = require('d' + 'ns'); dns['look' + 'up'](); dns[`resolve4`](); const {[`reverse`]: run} = dns; run();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 34, 1, 54),
					dnsError("resolve4", 1, 56, 1, 73),
					dnsError("reverse", 1, 107, 1, 112),
				},
			},
			// Optional loaders, members and calls.
			{Code: "require?.('dns')?.lookup?.(); process?.getBuiltinModule?.('node:dns')?.resolve?.(); (require('dns')?.reverse)();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 1, 1, 29),
					dnsError("resolve", 1, 31, 1, 83),
					dnsError("reverse", 1, 85, 1, 112),
				},
			},
			// Assignment and parameter defaults retain aliases.
			{Code: "let lookup; ({lookup} = require('dns')); lookup(); function f({resolve: run} = require('dns')) { run(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 42, 1, 50),
					dnsError("resolve", 1, 98, 1, 103),
				},
			},
			// Flow-insensitive aliases survive later writes.
			{Code: "let dns = require('dns'); const lookup = dns.lookup; dns = other; lookup(); dns.resolve();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 67, 1, 75),
					dnsError("resolve", 1, 77, 1, 90),
				},
			},
			// Global-object module loaders.
			{Code: "global.require('dns').lookup(); globalThis.process.getBuiltinModule('dns').lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 1, 1, 31),
					dnsError("lookup", 1, 33, 1, 84),
				},
			},
			// Conditional, logical and sequence values preserve module references.
			{Code: "(flag ? require('dns') : other).lookup(); (require('dns') || other).reverse(); (0, require('dns')).resolve();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 1, 1, 41),
					dnsError("reverse", 1, 43, 1, 78),
					dnsError("resolve", 1, 80, 1, 109),
				},
			},
			// Independent paths to one binding preserve duplicate diagnostics.
			{Code: "const dns = flag ? require('dns') : require('node:dns'); dns.lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 58, 1, 70),
					dnsError("lookup", 1, 58, 1, 70),
				},
			},
			// Nested calls are sorted by source range.
			{Code: "const dns = require('dns'); dns.lookup(dns.resolve());",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 29, 1, 54),
					dnsError("resolve", 1, 40, 1, 53),
				},
			},
			// TypeScript assertions, non-null receivers and generic calls.
			{Code: "import dns from 'dns'; (dns as any).lookup(); dns!.resolve(); (dns satisfies any).reverse(); dns.lookup<string>();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 24, 1, 45),
					dnsError("resolve", 1, 47, 1, 61),
					dnsError("reverse", 1, 63, 1, 92),
					dnsError("lookup", 1, 94, 1, 114),
				},
			},
			// JSX expression containers remain runtime calls.
			{Code: "import dns from 'dns'; const el = <dns.lookup value={dns.lookup()}>{new dns.Resolver()}</dns.lookup>;",
				FileName: "input.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 54, 1, 66),
					dnsError("Resolver", 1, 69, 1, 87),
				},
			},
			// JSDoc wrappers preserve lookup calls.
			{Code: "const dns = require('dns'); (/** @type {any} */ (dns)).lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 29, 1, 64),
				},
			},
			// UTF-16 columns and multiline call ranges.
			{Code: "const dns = require('dns');\r\n'😀'; dns\r\n  .lookup('example.com',\r\n    callback);\r\nnew dns.Resolver;",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 2, 7, 4, 14),
					dnsError("Resolver", 5, 1, 5, 17),
				},
			},
			// A local parameter shadows only its own references.
			{Code: "import dns from 'dns'; function f(dns) { dns.lookup(); } dns.lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 58, 1, 70),
				},
			},
			// Aliased object defaults and constructor through getBuiltinModule.
			{Code: "function f(dns = process.getBuiltinModule('node:dns')) { new dns.Resolver(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("Resolver", 1, 58, 1, 76),
				},
			},
			// Merged type and value declarations retain the value alias.
			{Code: "type lookup = string; let lookup; lookup = require('dns').lookup; lookup();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 67, 1, 75),
				},
			},
			// Quoted ESM names and namespace default access.
			{Code: "import {'lookup' as run} from 'dns'; run(); import * as dns from 'node:dns'; dns.default.lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 38, 1, 43),
					dnsError("lookup", 1, 78, 1, 98),
				},
			},
			// Logical assignment and nested default aliases.
			{Code: "let run; run ||= require('dns').lookup; run(); const {lookup = fallback} = require('dns'); lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 41, 1, 46),
					dnsError("lookup", 1, 92, 1, 100),
				},
			},
			// Default-value expressions in destructuring patterns are tracked.
			{Code: "const {run = require('dns').lookup} = other; run();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 46, 1, 51),
				},
			},
			// Cyclic aliases terminate without losing reachable calls.
			{Code: "let dns = require('dns'); let other = dns; dns = other; other.lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 57, 1, 71),
				},
			},
			// Repeated method alias sources retain distinct messages at one range.
			{Code: "import {lookup, reverse} from 'dns'; const run = flag ? reverse : lookup; run();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 75, 1, 80),
					dnsError("reverse", 1, 75, 1, 80),
				},
			},
			// Optional TypeScript receivers and call expressions.
			{Code: "import dns from 'dns'; dns?.lookup!(); (dns?.lookup)!(); dns['lookup' as const]();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 24, 1, 38),
					dnsError("lookup", 1, 40, 1, 56),
					dnsError("lookup", 1, 58, 1, 82),
				},
			},
			// Using declarations keep module aliases.
			{Code: "using dns = require('dns'); dns.lookup();",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 29, 1, 41),
				},
			},
			// Class static blocks and decorators evaluate DNS calls.
			{Code: "import dns from 'dns'; class C { static { dns.lookup(); } @dns.lookup() accessor value; }",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 43, 1, 55),
					dnsError("lookup", 1, 60, 1, 72),
				},
			},
			// Import attributes still track DNS imports.
			{Code: "import dns from 'dns' with {type: 'json'}; dns.lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 44, 1, 56),
				},
			},
			// Escaped identifiers and string literals retain canonical messages.
			{Code: "const dn\\u0073 = require('d\\u006es'); dns.l\\u006fokup(); dns['l\\u006fokup']();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 39, 1, 56),
					dnsError("lookup", 1, 58, 1, 78),
				},
			},
			// For-of body calls preserve bindings.
			{Code: "import dns from 'dns'; for (const item of data) { dns.lookup(item); }",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 51, 1, 67),
				},
			},
			// JSDoc type declarations do not shadow actual globals.
			{Code: "/** @typedef {object} require */ const dns = require('dns'); dns.lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 1, 62, 1, 74),
				},
			},
			// Parenthesized optional chains with leading comments preserve ranges.
			{Code: "const dns = require('dns');\n(/* leading */ dns?.lookup)();\nnew /* constructor */ (dns.Resolver);",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 2, 1, 2, 30),
					dnsError("Resolver", 3, 1, 3, 37),
				},
			},
			// Unicode line separators use ECMAScript locations.
			{Code: "const dns = require('dns');\u2028dns.lookup();\u2029new dns.Resolver();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 2, 1, 2, 13),
					dnsError("Resolver", 3, 1, 3, 19),
				},
			},
			// Enable and disable comments preserve the middle report.
			// The Go RuleTester registers the rule under the name "test".
			{Code: "import dns from 'dns';\n// eslint-disable-next-line test\ndns.lookup();\ndns.resolve(); // eslint-disable-line test\n/* eslint-disable test */\nnew dns.Resolver();\n/* eslint-enable test */\ndns.lookup();",
				Errors: []rule_tester.InvalidTestCaseError{
					dnsError("lookup", 8, 1, 8, 13),
				},
			},
		},
	)
}
