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
			{Code: "foo(bar(baz(qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}},
			{Code: "foo(bar(baz()));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}, Options: []any{map[string]any{"max": 2}}},
			{Code: "new Foo(new Bar(new Baz(new Qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}},
			{Code: "await foo(await bar(await baz(await qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}},
			{Code: "foo?.(bar?.(baz?.(qux?.())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}},
			{Code: "foo(...bar(baz(qux())));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}},
			{Code: "foo(condition ? bar(baz(qux())) : zed());", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}},
			{Code: "foo(class {field = bar(baz(qux(zed())));});", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}},
			{Code: "<Component value={foo(bar(baz(qux())))} />;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}, Tsx: true},
			{Code: "mergeReports(await pMap(\n\tawait mergeWithFileConfigs(uniq(paths), inputOptions, configFiles),\n\tasync ({files, options, prettierOptions}) => runEslint(files, buildConfig(options, prettierOptions), {isQuiet: options.quiet}),\n));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "max-nested-calls"}}},
		},
	)
}
