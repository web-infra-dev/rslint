// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/prefer-optional-catch-binding.js
package prefer_optional_catch_binding_test

import (
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_optional_catch_binding"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestPreferOptionalCatchBindingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_optional_catch_binding.PreferOptionalCatchBindingRule, []rule_tester.ValidTestCase{
		{Code: "try {} catch {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch {\n\terror\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(used) {\n\tconsole.error(used);\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch(usedInADeeperScope) {\n\tfunction foo() {\n\t\tfunction bar() {\n\t\t\tconsole.error(usedInADeeperScope);\n\t\t}\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch ({message}) {alert(message)}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch ({cause: {message}}) {alert(message)}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "try {} catch({nonExistsProperty = thisWillExecute()}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\ntry {\n\t// do something\n} catch {\n\t// ignore error\n}\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\ntry {\n\tawait fetch(url);\n} catch {\n\t// ignore fetch errors\n}\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
		{Code: "// ✅\ntry {\n\tdoSomething();\n} catch (error) {\n\t// error is actually used\n\tconsole.log(error.message);\n}\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "try {} catch (_) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `_`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch (foo) {\n\tfunction bar(foo) {}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {\n\tfunction bar(foo) {}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `foo`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch (outer) {\n\ttry {} catch (inner) {\n\t}\n}\ntry {\n\ttry {} catch (inTry) {\n\t}\n} catch (another) {\n\ttry {} catch (inCatch) {\n\t}\n} finally {\n\ttry {} catch (inFinally) {\n\t}\n}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {\n\ttry {} catch {\n\t}\n}\ntry {\n\ttry {} catch {\n\t}\n} catch {\n\ttry {} catch {\n\t}\n} finally {\n\ttry {} catch {\n\t}\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `outer`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "with-name", Message: "Remove unused catch binding `inner`.", Line: 2, Column: 16, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "with-name", Message: "Remove unused catch binding `inTry`.", Line: 6, Column: 16, EndLine: 6, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "with-name", Message: "Remove unused catch binding `another`.", Line: 8, Column: 10, EndLine: 8, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "with-name", Message: "Remove unused catch binding `inCatch`.", Line: 9, Column: 16, EndLine: 9, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			{MessageId: "with-name", Message: "Remove unused catch binding `inFinally`.", Line: 12, Column: 16, EndLine: 12, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch (theRealErrorName) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `theRealErrorName`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "/* comment */\ntry {\n\t/* comment */\n\t// comment\n} catch (\n\t/* comment */\n\t// comment\n\tunused\n\t/* comment */\n\t// comment\n) {\n\t/* comment */\n\t// comment\n}\n/* comment */", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"/* comment */\ntry {\n\t/* comment */\n\t// comment\n} catch \n\t/* comment */\n\t// comment\n\t\n\t/* comment */\n\t// comment\n{\n\t/* comment */\n\t// comment\n}\n/* comment */"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `unused`.", Line: 8, Column: 2, EndLine: 8, EndColumn: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try    {    } catch    (e)  \n  \t  {    }", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try    {    } catch    {    }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch(e) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch{}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch (e){}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `e`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch ({}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 15, EndLine: 1, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch ({/* inner comment */ message}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 15, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch ({message}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch ({message: notUsedMessage}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 15, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "try {} catch ({cause: {message}}) {}", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"try {} catch {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "without-name", Message: "Remove unused catch binding.", Line: 1, Column: 15, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\ntry {\n\t// do something\n} catch (notUsedError) {\n\t// ignore error\n}\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\ntry {\n\t// do something\n} catch {\n\t// ignore error\n}\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `notUsedError`.", Line: 4, Column: 10, EndLine: 4, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
		{Code: "// ❌\ntry {\n\tawait fetch(url);\n} catch (error) {\n\t// error is not used, just continue\n}\n\n", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"window": "readonly", "global": "readonly", "self": "readonly"}, Output: []string{"// ❌\ntry {\n\tawait fetch(url);\n} catch {\n\t// error is not used, just continue\n}\n\n"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "with-name", Message: "Remove unused catch binding `error`.", Line: 4, Column: 10, EndLine: 4, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
		}},
	})
}
