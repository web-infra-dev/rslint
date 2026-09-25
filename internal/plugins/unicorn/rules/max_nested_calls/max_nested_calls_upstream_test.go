package max_nested_calls_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/max_nested_calls"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMaxNestedCallsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&max_nested_calls.MaxNestedCallsRule,
		[]rule_tester.ValidTestCase{
			{Code: "foo();"},
			{Code: "foo(bar(baz()));"},
			{Code: "foo(bar(), baz(), qux());"},
			{Code: "query().filter().map().toArray();"},
			{Code: "foo()[bar()]().baz();"},
			{Code: "foo(() => bar(baz(qux())));"},
			{Code: "foo(bar(class {field = baz(qux());}));"},
			{Code: "foo(bar(class {static {baz(qux());}}));"},
			{Code: "new Foo(new Bar(new Baz()));"},
			{Code: "await foo(await bar(await baz()));"},
			{Code: "foo?.(bar?.(baz?.()));"},
			{Code: "foo(...bar(baz()));"},
			{Code: "foo(condition ? bar(baz()) : qux());"},
			{Code: "nativeDiffButtons.parentElement.after(\n\ttooltipped(\n\t\t{},\n\t\t<a className={cx('btn', isHidingWhitespace() && 'color-fg-subtle')} />,\n\t),\n);", Tsx: true},
			{Code: "nativeDiffButtons.parentElement.after(\n\ttooltipped(\n\t\t{},\n\t\t<>{cx('btn', isHidingWhitespace() && 'color-fg-subtle')}</>,\n\t),\n);", Tsx: true},
			{Code: "foo(bar(baz(qux())));", Options: []any{map[string]any{"max": 4}}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "foo(bar(baz(qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 1, Column: 13, EndLine: 1, EndColumn: 18}}},
			{Code: "foo(bar(baz()));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 2.", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}, Options: []any{map[string]any{"max": 2}}},
			{Code: "new Foo(new Bar(new Baz(new Qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 1, Column: 25, EndLine: 1, EndColumn: 34}}},
			{Code: "await foo(await bar(await baz(await qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 1, Column: 37, EndLine: 1, EndColumn: 42}}},
			{Code: "foo?.(bar?.(baz?.(qux?.())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}}},
			{Code: "foo(...bar(baz(qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21}}},
			{Code: "foo(condition ? bar(baz(qux())) : zed());", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 1, Column: 25, EndLine: 1, EndColumn: 30}}},
			{Code: "foo(class {field = bar(baz(qux(zed())));});", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 1, Column: 32, EndLine: 1, EndColumn: 37}}},
			{Code: "<Component value={foo(bar(baz(qux())))} />;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 1, Column: 31, EndLine: 1, EndColumn: 36}}, Tsx: true},
			{Code: "mergeReports(await pMap(\n\tawait mergeWithFileConfigs(uniq(paths), inputOptions, configFiles),\n\tasync ({files, options, prettierOptions}) => runEslint(files, buildConfig(options, prettierOptions), {isQuiet: options.quiet}),\n));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls", Message: "Call is nested too deeply. Maximum allowed is 3.", Line: 2, Column: 29, EndLine: 2, EndColumn: 40}}},
		},
	)
}
