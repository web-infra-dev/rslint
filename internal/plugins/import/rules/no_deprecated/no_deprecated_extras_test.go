package no_deprecated_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_deprecated"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestNoDeprecatedExtras(t *testing.T) {
	rule_tester.RunRuleTester(deprecatedRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &no_deprecated.NoDeprecatedRule, []rule_tester.ValidTestCase{ // later comments override statement and earlier blocks
		{Code: "import {plain, clean} from './values';", FileName: "consumer.tsx", Tsx: true},
		// disabled export docs
		{Code: "import {old, dash} from './values';", FileName: "consumer.tsx", Tsx: true, Settings: map[string]any{"import/docstyle": []any{}}},
		// local export list does not copy declaration docs
		{Code: "import {old} from './local'; old;", FileName: "consumer.tsx", Tsx: true},
		// computed properties stop namespace traversal
		{Code: "import * as ns from './values'; ns['old']; ns[key]; const {old}=ns;", FileName: "consumer.tsx", Tsx: true},
		// parameters, catch, blocks, hoisted vars and loop shadowing
		{Code: "import * as ns from './values'; function f(ns){ns.old} try{}catch(ns){ns.old} { ns.old; let ns; } function g(){ns.old; var ns;} for(let ns of []){ns.old}", FileName: "consumer.tsx", Tsx: true},
		// no module tag means no module report
		{Code: "import './not-module';", FileName: "consumer.tsx", Tsx: true},
		// ignored and unresolved modules
		{Code: "import {old} from './values'; import {missing} from './absent';", FileName: "consumer.tsx", Tsx: true, Settings: map[string]any{"import/ignore": []any{"values"}}},
		// extension setting excludes export metadata
		{Code: "import {old} from './values';", FileName: "consumer.tsx", Tsx: true, Settings: map[string]any{"import/extensions": []any{".ts"}}},
		// bare or empty namespace
		{Code: "import * as ns from './values'; ns; import * as empty from './empty'; empty.old;", FileName: "consumer.tsx", Tsx: true},
		// ESTree literal import names have no .name; pinned TS parser cannot parse
		// them, so this native representation case is retained without a reference run.
		{Code: "import {'old' as alias} from './values'; alias;"},
	}, []rule_tester.InvalidTestCase{ // empty description, import alias and use
		{Code: "import {old as alias} from './values'; alias;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 21}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 40, EndLine: 1, EndColumn: 45}}},
		// dash, multiline and Unicode descriptions
		{Code: "import {dash, multiline, unicode} from './values';", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: use newValue", Line: 1, Column: 9, EndLine: 1, EndColumn: 13}, {MessageId: "deprecated", Message: "Deprecated: first line\n  second line", Line: 1, Column: 15, EndLine: 1, EndColumn: 24}, {MessageId: "deprecated", Message: "Deprecated: unicode space", Line: 1, Column: 26, EndLine: 1, EndColumn: 33}}},
		// last block and ordinary block
		{Code: "import {last, ordinary, terminated, repeated} from './values';", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: last", Line: 1, Column: 9, EndLine: 1, EndColumn: 13}, {MessageId: "deprecated", Message: "Deprecated: ordinary block", Line: 1, Column: 15, EndLine: 1, EndColumn: 23}, {MessageId: "deprecated", Message: "Deprecated: note", Line: 1, Column: 25, EndLine: 1, EndColumn: 35}, {MessageId: "deprecated", Message: "Deprecated: repeat first", Line: 1, Column: 37, EndLine: 1, EndColumn: 45}}},
		// destructured exports and per-variable comments
		{Code: "import {a, b, later, blocked} from './values';", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: tuple", Line: 1, Column: 9, EndLine: 1, EndColumn: 10}, {MessageId: "deprecated", Message: "Deprecated: tuple", Line: 1, Column: 12, EndLine: 1, EndColumn: 13}, {MessageId: "deprecated", Message: "Deprecated: later variable", Line: 1, Column: 15, EndLine: 1, EndColumn: 20}, {MessageId: "deprecated", Message: "Deprecated: statement", Line: 1, Column: 22, EndLine: 1, EndColumn: 29}}},
		// TomDoc paragraph, ordered style override
		{Code: "import {tom} from './values';", FileName: "consumer.tsx", Tsx: true, Settings: map[string]any{"import/docstyle": []any{"jsdoc", "tomdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: wrapped over two lines", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		// last matching style wins
		{Code: "import {old} from './public';", FileName: "consumer.tsx", Tsx: true, Settings: map[string]any{"import/docstyle": []any{"jsdoc", "tomdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: TomDoc", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		// reverse order restores JSDoc
		{Code: "import {old} from './public';", FileName: "consumer.tsx", Tsx: true, Settings: map[string]any{"import/docstyle": []any{"tomdoc", "jsdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: JSDoc", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		// duplicate style uses first insertion order
		{Code: "import {old} from './public';", FileName: "consumer.tsx", Tsx: true, Settings: map[string]any{"import/docstyle": []any{"jsdoc", "tomdoc", "jsdoc"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: TomDoc", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		// named and star reexports
		{Code: "import {renamed, ordinary} from './barrel'; renamed;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}, {MessageId: "deprecated", Message: "Deprecated: ordinary block", Line: 1, Column: 18, EndLine: 1, EndColumn: 26}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 45, EndLine: 1, EndColumn: 52}}},
		// TypeScript assigned default through named reexport
		{Code: "import {Old} from './assigned-barrel.ts'; Old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: assigned class", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}, {MessageId: "deprecated", Message: "Deprecated: assigned class", Line: 1, Column: 43, EndLine: 1, EndColumn: 46}}},
		// TypeScript assigned class docs
		{Code: "import Old from './assigned-class.ts'; Old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: assigned class", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}, {MessageId: "deprecated", Message: "Deprecated: assigned class", Line: 1, Column: 40, EndLine: 1, EndColumn: 43}}},
		// TypeScript runtime function uses assignment docs
		{Code: "import old from './assigned-function.ts'; old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: assignment reason", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}, {MessageId: "deprecated", Message: "Deprecated: assignment reason", Line: 1, Column: 43, EndLine: 1, EndColumn: 46}}},
		// TypeScript assigned value docs
		{Code: "import old from './assigned-value.ts'; old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: assigned value", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}, {MessageId: "deprecated", Message: "Deprecated: assigned value", Line: 1, Column: 40, EndLine: 1, EndColumn: 43}}},
		// TypeScript assigned namespace member docs
		{Code: "import * as ns from './assigned-namespace.ts'; ns.old(); ns.value;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: member reason", Line: 1, Column: 51, EndLine: 1, EndColumn: 54}, {MessageId: "deprecated", Message: "Deprecated: variable reason", Line: 1, Column: 61, EndLine: 1, EndColumn: 66}}},
		// default expression docs
		{Code: "import Default from './values'; Default;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: default expression", Line: 1, Column: 8, EndLine: 1, EndColumn: 15}, {MessageId: "deprecated", Message: "Deprecated: default expression", Line: 1, Column: 33, EndLine: 1, EndColumn: 40}}},
		// deprecated namespace receiver and member report in source order
		{Code: "import ns from './ns-default'; ns.old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: namespace reason", Line: 1, Column: 8, EndLine: 1, EndColumn: 10}, {MessageId: "deprecated", Message: "Deprecated: namespace reason", Line: 1, Column: 32, EndLine: 1, EndColumn: 34}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 35, EndLine: 1, EndColumn: 38}}},
		// deep namespaces and parentheses
		{Code: "import * as ns from './deep'; console.log(((ns).nested.inner).old);", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 63, EndLine: 1, EndColumn: 66}}},
		// optional member access
		{Code: "import * as ns from './values'; ns?.old; (ns)?.ordinary;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 37, EndLine: 1, EndColumn: 40}, {MessageId: "deprecated", Message: "Deprecated: ordinary block", Line: 1, Column: 48, EndLine: 1, EndColumn: 56}}},
		// computed identifiers are ignored even through parentheses
		{Code: "import {old} from './values'; other[(old)];", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		// private member
		{Code: "import * as ns from './values'; class C { #old; read(){ return ns.#old; } }", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 67, EndLine: 1, EndColumn: 71}}},
		// module references from arrows, class fields and static blocks
		{Code: "import {old} from './values'; const f=()=>old; class C { x=old; static{old;} }", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 43, EndLine: 1, EndColumn: 46}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 60, EndLine: 1, EndColumn: 63}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 72, EndLine: 1, EndColumn: 75}}},
		// default parameter uses outer import
		{Code: "import {old} from './values'; function f(x=old){ const old=1; }", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 44, EndLine: 1, EndColumn: 47}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 56, EndLine: 1, EndColumn: 59}}},
		// property key follows first scope reference
		{Code: "import {old} from './values'; console.log({old:1},old);", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 44, EndLine: 1, EndColumn: 47}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 51, EndLine: 1, EndColumn: 54}}},
		// property key without a same-scope reference
		{Code: "import {old} from './values'; console.log({old:1});", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		// JSX tags are JSXIdentifier, runtime expression is Identifier
		{Code: "import {old} from './values'; const x=<old attr={old}/>;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 50, EndLine: 1, EndColumn: 53}}},
		// type-only import and reference
		{Code: "import type {old} from './values'; type T=typeof old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 14, EndLine: 1, EndColumn: 17}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 50, EndLine: 1, EndColumn: 53}}},
		// TS wrapper does not erase a receiver boundary
		{Code: "import * as ns from './values'; (ns as any).old; (ns!).old; ns.old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 64, EndLine: 1, EndColumn: 67}}},
		// JSDoc cast erases its receiver wrapper
		{Code: "import * as ns from './values'; (/** @type {any} */ (ns)).old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 59, EndLine: 1, EndColumn: 62}}},
		// module side effect import
		{Code: "import './module';", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: whole module", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}}},
		// module report survives disabled export docs
		{Code: "import {old} from './module'; old;", FileName: "consumer.tsx", Tsx: true, Settings: map[string]any{"import/docstyle": []any{}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: whole module", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}},
		// module and individual export both deprecated
		{Code: "import {old} from './module'; old;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: whole module", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}, {MessageId: "deprecated", Message: "Deprecated: one export", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}, {MessageId: "deprecated", Message: "Deprecated: one export", Line: 1, Column: 31, EndLine: 1, EndColumn: 34}}},
		// first module doc stops scanning
		{Code: "import {fine} from './first-module';", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: ignored second module", Line: 1, Column: 9, EndLine: 1, EndColumn: 13}}},
		// UTF-16 and multiline locations
		{Code: "import {old as 古い} from './values';\nconst emoji='😀'; 古い;", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "deprecated", Message: "Deprecated.", Line: 2, Column: 19, EndLine: 2, EndColumn: 21}}},
		// export specifier identifiers follow upstream scope references
		{Code: "import {old} from './values'; export {old as alias};", FileName: "consumer.tsx", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}, {MessageId: "deprecated", Message: "Deprecated.", Line: 1, Column: 39, EndLine: 1, EndColumn: 42}}},
		// Doctrine stops at the malformed @param before @deprecated. The metadata
		// reader tolerates unrelated malformed types; see Differences from upstream.
		{Code: "import {old} from './malformed-doc';", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: tolerated malformed type", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
	})
}

// Namespace variables share their first documentation candidate but must cache
// the complete candidate list independently. Reproduces real TypeScript 4.5 APIs.
func TestNoDeprecatedNamespaceDocsCache(t *testing.T) {
	rule_tester.RunRuleTester(deprecatedRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &no_deprecated.NoDeprecatedRule, nil, []rule_tester.InvalidTestCase{{Code: "import {value, other} from './assigned-namespace.ts'; value; other; import * as ns from './assigned-namespace.ts'; ns.value; ns.other;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "deprecated", Message: "Deprecated: variable reason", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}, {MessageId: "deprecated", Message: "Deprecated: other reason", Line: 1, Column: 16, EndLine: 1, EndColumn: 21}, {MessageId: "deprecated", Message: "Deprecated: variable reason", Line: 1, Column: 55, EndLine: 1, EndColumn: 60}, {MessageId: "deprecated", Message: "Deprecated: other reason", Line: 1, Column: 62, EndLine: 1, EndColumn: 67}, {MessageId: "deprecated", Message: "Deprecated: variable reason", Line: 1, Column: 119, EndLine: 1, EndColumn: 124}, {MessageId: "deprecated", Message: "Deprecated: other reason", Line: 1, Column: 129, EndLine: 1, EndColumn: 134}}}})
}
