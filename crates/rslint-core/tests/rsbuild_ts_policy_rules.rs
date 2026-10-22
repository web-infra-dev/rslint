//! Rsbuild's default policy rule corpus and exact diagnostics/suggestions.
use rslint_core::lint_json;
use serde_json::{Value, json};

const BAN: &str = "@typescript-eslint/ban-ts-comment";
const REQUIRE: &str = "@typescript-eslint/no-require-imports";
const UNUSED: &str = "@typescript-eslint/no-unused-expressions";
const REFERENCE: &str = "@typescript-eslint/triple-slash-reference";

fn lint_files(rule: &str, files: &[(&str, &str)], fix: bool) -> Vec<Value> {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-policy-tests",
        "files":files.iter().map(|(path,text)|json!({"path":path,"text":text})).collect::<Vec<_>>(),
        "rules":[{"name":rule}],"fix":fix});
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    let results = result["files"].as_array().unwrap();
    assert_eq!(results.len(), files.len(), "retain admitted silent files");
    for (result, (path, _)) in results.iter().zip(files) {
        assert_eq!(
            result["filePath"],
            format!("/rslint-rsbuild-policy-tests/{path}")
        );
        assert_eq!(result["appliedRules"], json!([rule]));
    }
    results.clone()
}

fn messages(rule: &str, source: &str) -> Vec<Value> {
    lint_files(rule, &[("input.ts", source)], false)[0]["messages"]
        .as_array()
        .unwrap()
        .clone()
}

fn utf16(text: &str, index: usize) -> usize {
    text[..index].encode_utf16().count()
}

#[test]
fn policy_corpus_exercises_default_rules_and_preserves_silent_files() {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-ts-policy.json"
    ))
    .unwrap();
    for group in corpus.as_array().unwrap() {
        let rule = group["rule"].as_str().unwrap();
        for category in ["valid", "invalid"] {
            let cases = group[category].as_array().unwrap();
            assert!(!cases.is_empty());
            for case in cases {
                let (text, extension) = case.as_str().map_or_else(
                    || {
                        (
                            case["text"].as_str().unwrap(),
                            case["extension"].as_str().unwrap_or(".ts"),
                        )
                    },
                    |text| (text, ".ts"),
                );
                let results =
                    lint_files(rule, &[(format!("input{extension}").as_str(), text)], false);
                let diagnostics = results[0]["messages"].as_array().unwrap();
                assert_eq!(
                    diagnostics.is_empty(),
                    category == "valid",
                    "{rule}: {text}: {diagnostics:?}"
                );
                for diagnostic in diagnostics {
                    assert_eq!(diagnostic["ruleId"], rule);
                    assert_eq!(diagnostic["severity"], 2);
                    assert!(diagnostic["endColumn"].as_u64().is_some());
                }
            }
        }
    }
}

#[test]
fn ban_comment_reports_whole_trivia_but_suggests_only_last_directive_marker() {
    let source = "const emoji = '😀'; // @ts-ignore repeat @ts-ignore";
    let diagnostics = messages(BAN, source);
    assert_eq!(diagnostics.len(), 1);
    let report = &diagnostics[0];
    assert_eq!(report["messageId"], "tsIgnoreInsteadOfExpectError");
    assert_eq!(
        report["message"],
        "Prefer '@ts-expect-error' over '@ts-ignore' as it requires error to be present in next line."
    );
    assert_eq!(report["line"], 1);
    assert_eq!(
        report["column"],
        utf16(source, source.find("//").unwrap()) + 1
    );
    assert_eq!(report["endColumn"], source.encode_utf16().count() + 1);
    assert!(report.get("fix").is_none());
    let marker = source.rfind("@ts-ignore").unwrap();
    assert_eq!(
        report["suggestions"],
        json!([{
            "messageId":"replaceTsIgnoreWithTsExpectError",
            "desc":"Replace '@ts-ignore' with '@ts-expect-error'.",
            "fix":{"range":[utf16(source,marker), utf16(source,marker+"@ts-ignore".len())],"text":"@ts-expect-error"}
        }])
    );
    let fixed = lint_files(BAN, &[("input.ts", source)], true);
    assert!(
        fixed[0].get("output").is_none(),
        "this edit must remain a suggestion"
    );
    assert_eq!(fixed[0]["messages"], json!(diagnostics));
}

#[test]
fn expect_error_description_counts_graphemes_and_only_last_block_line() {
    for description in [
        "ab",
        "a\u{0301}b\u{0301}",
        "👨‍👩‍👧‍👦👨‍👩‍👧‍👦",
        "🇦🇺🇨🇦",
        "\u{00a0}ab\u{feff}",
    ] {
        let source = format!("/** first line\n * @ts-expect-error {description} */");
        let diagnostics = messages(BAN, &source);
        assert_eq!(diagnostics.len(), 1, "{description}");
        assert_eq!(
            diagnostics[0]["messageId"],
            "tsDirectiveCommentRequiresDescription"
        );
        assert_eq!(
            diagnostics[0]["message"],
            "Include a description after the '@ts-expect-error' directive to explain why the '@ts-expect-error' is necessary. The description must be 3 characters or longer."
        );
        assert_eq!(diagnostics[0]["line"], 1);
        assert_eq!(diagnostics[0]["column"], 1);
        assert_eq!(diagnostics[0]["endLine"], 2);
        assert!(diagnostics[0].get("suggestions").is_none());
    }
    assert!(messages(BAN, "/* @ts-ignore\n */").is_empty());
    assert!(messages(BAN, "// @ts-expect-error 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦").is_empty());
    let diagnostics = messages(BAN, "// @ts-nocheck\nconst x = 1;\n// @ts-nocheck");
    assert_eq!(diagnostics.len(), 1);
    assert_eq!(diagnostics[0]["messageId"], "tsDirectiveComment");
    assert_eq!(
        diagnostics[0]["message"],
        "Do not use '@ts-nocheck' because it alters compilation errors."
    );
}

#[test]
fn require_import_uses_call_range_and_preserves_external_module_reference_range() {
    let source = "const emoji = '😀'; import value = require('pkg');";
    let diagnostics = messages(REQUIRE, source);
    assert_eq!(diagnostics.len(), 1);
    let report = &diagnostics[0];
    assert_eq!(report["messageId"], "noRequireImports");
    assert_eq!(
        report["message"],
        "A `require()` style import is forbidden."
    );
    assert_eq!(
        report["column"],
        utf16(source, source.find("require").unwrap()) + 1
    );
    assert_eq!(
        report["endColumn"],
        utf16(source, source.rfind(')').unwrap() + 1) + 1
    );
    assert!(report.get("fix").is_none());
    assert!(report.get("suggestions").is_none());
    let results = lint_files(
        REQUIRE,
        &[
            (
                "global.d.ts",
                "declare function require(id: string): unknown;",
            ),
            ("input.ts", "require('pkg');"),
        ],
        false,
    );
    assert_eq!(results[0]["messages"], json!([]));
    assert_eq!(
        results[1]["messages"].as_array().unwrap().len(),
        1,
        "ambient declaration in another file does not become local require"
    );
}

#[test]
fn expression_reports_include_assertions_parentheses_and_statement_semicolon() {
    let source = "const emoji = '😀'; ((foo as string)!);";
    let report = &messages(UNUSED, source)[0];
    assert_eq!(report["messageId"], "unusedExpression");
    assert_eq!(
        report["message"],
        "Expected an assignment or function call and instead saw an expression."
    );
    assert_eq!(
        report["column"],
        utf16(source, source.find("((foo").unwrap()) + 1
    );
    assert_eq!(report["endColumn"], source.encode_utf16().count() + 1);
    assert!(report.get("fix").is_none());
    assert!(report.get("suggestions").is_none());
    let source = "class C { static { 'first'; run(); 'late'; } }";
    let diagnostics = messages(UNUSED, source);
    assert_eq!(diagnostics.len(), 1);
    assert_eq!(diagnostics[0]["column"], source.find("'late'").unwrap() + 1);
    assert_eq!(
        diagnostics[0]["endColumn"],
        source.find("'late'").unwrap() + "'late';".len() + 1
    );
    assert!(messages(UNUSED, "foo satisfies string; (foo satisfies string)!;").is_empty());
}

#[test]
fn triple_slash_uses_only_leading_trivia_and_static_matching_imports() {
    let source = "\t/// <reference path=\"foo\" />\r\n/// <reference types=\"foo\" />\r\nimport 'f\\u006fo';\r\n/// <reference path=\"late\" />";
    let diagnostics = messages(REFERENCE, source);
    assert_eq!(diagnostics.len(), 2);
    for report in &diagnostics {
        assert_eq!(report["messageId"], "tripleSlashReference");
        assert_eq!(
            report["message"],
            "Do not use a triple slash reference for foo, use `import` style instead."
        );
        assert!(report.get("fix").is_none());
        assert!(report.get("suggestions").is_none());
    }
    assert_eq!(diagnostics[0]["line"], 1);
    assert_eq!(diagnostics[0]["column"], 2);
    assert_eq!(diagnostics[0]["endColumn"], 30);
    assert_eq!(diagnostics[1]["line"], 2);
    assert_eq!(diagnostics[1]["column"], 1);
    let diagnostics = messages(
        REFERENCE,
        "/// <reference types=\"foo\" />\nimport 'foo'; import type { Foo } from 'foo';",
    );
    assert_eq!(
        diagnostics.len(),
        2,
        "each matching import reports the reference, as in the frozen oracle"
    );
    assert_eq!(diagnostics[0], diagnostics[1]);
}
