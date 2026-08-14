// TestCatchErrorNameUpstream migrates the behavioral contract from upstream
// v73.0.0 test/catch-error-name.js. rslint-specific lock-in cases live in
// catch_error_name_extras_test.go.
package catch_error_name_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/catch_error_name"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const messageID = "catch-error-name"

func expected(original, fixed string) string {
	return "The catch parameter `" + original + "` should be named `" + fixed + "`."
}

func invalid(code, output, original, fixed string, options ...any) rule_tester.InvalidTestCase {
	column := strings.Index(code, original) + 1
	result := rule_tester.InvalidTestCase{
		Code:   code,
		Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID,
			Message:   expected(original, fixed),
			Line:      1,
			Column:    column,
		}},
	}
	if len(options) > 0 {
		result.Options = options
	}
	return result
}

func TestCatchErrorNameUpstream(t *testing.T) {
	customErr := map[string]any{"name": "err"}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&catch_error_name.CatchErrorNameRule,
		[]rule_tester.ValidTestCase{
			{Code: "try {} catch (error) {}"},
			{Code: "try {} catch (descriptiveError) {}"},
			{Code: "try {} catch (descriptive_error__) {}"},
			{Code: "try {} catch (_) {}"},
			{Code: "try {} catch ({message}) {}"},
			{Code: "try {} catch {}"},
			{Code: "obj.catch(error => {})"},
			{Code: "obj.then(undefined, error => {})"},
			{Code: "obj.catch?.(bad => {})"},
			{Code: "obj.then?.(undefined, bad => {})"},
			{Code: "obj.catch(error => {}, extra)"},
			{Code: "obj.then(undefined, error => {}, extra)"},
			{Code: "foo(function (error) {})"},
			{Code: "foo().catch(function (error) {})"},
			{Code: "obj.catch(({message}) => {})"},
			{Code: "try {} catch (err) {}", Options: customErr},
			{Code: "try {} catch (skipThisNameCheck) {}", Options: map[string]any{"ignore": []any{"^skip"}}},
			{Code: "promise.catch(unicorn => {})", Options: map[string]any{"ignore": []any{"unicorn"}}},
		},
		[]rule_tester.InvalidTestCase{
			invalid("try {} catch (err) { console.log(err) }", "try {} catch (error) { console.log(error) }", "err", "error"),
			invalid("try {} catch (error) { console.log(error) }", "try {} catch (err) { console.log(err) }", "error", "err", customErr),
			invalid("obj.catch(err => err)", "obj.catch(error => error)", "err", "error"),
			invalid("obj.then(null, err => err)", "obj.then(null, error => error)", "err", "error"),
			invalid("obj?.catch(err => err)", "obj?.catch(error => error)", "err", "error"),
			invalid("try {} catch (_) { console.log(_) }", "try {} catch (error) { console.log(error) }", "_", "error"),
			invalid("try {} catch (notMatching) {}", "try {} catch (error) {}", "notMatching", "error", map[string]any{"ignore": []any{"unicorn"}}),
		},
	)
}
