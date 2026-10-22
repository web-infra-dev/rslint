//! Default constituent rules, including all runnable default cases from the
//! frozen suites. The corpus records original skips/non-default exclusions.
use rslint_core::lint_json;
use serde_json::{Value, json};

const DUPLICATE: &str = "@typescript-eslint/no-duplicate-type-constituents";
const REDUNDANT: &str = "@typescript-eslint/no-redundant-type-constituents";

fn lint_files(rule: &str, files: &[(&str, &str)], fix: bool) -> Vec<Value> {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-relations-tests",
        "files":files.iter().map(|(path,text)|json!({"path":path,"text":text})).collect::<Vec<_>>(),
        "rules":[{"name":rule}],"fix":fix});
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    let results = result["files"].as_array().unwrap();
    assert_eq!(results.len(), files.len(), "retain admitted silent files");
    for (result, (path, _)) in results.iter().zip(files) {
        assert_eq!(
            result["filePath"],
            format!("/rslint-rsbuild-relations-tests/{path}")
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
fn constituent_corpus_preserves_runnable_default_baseline_cases() {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-type-relations.json"
    ))
    .unwrap();
    let groups = corpus.as_array().unwrap();
    assert_eq!(groups.len(), 2);
    assert_eq!(
        groups[0]["knownBaselineSkipped"].as_array().unwrap().len(),
        2
    );
    assert_eq!(
        groups[0]["excludedNonDefaultCases"]
            .as_array()
            .unwrap()
            .len(),
        2
    );
    for group in groups {
        let rule = group["rule"].as_str().unwrap();
        for category in ["valid", "invalid"] {
            let cases = group[category].as_array().unwrap();
            assert!(!cases.is_empty());
            for case in cases {
                let source = case.as_str().unwrap();
                let diagnostics = messages(rule, source);
                assert_eq!(
                    diagnostics.is_empty(),
                    category == "valid",
                    "{rule}: {source}: {diagnostics:?}"
                );
                for diagnostic in diagnostics {
                    assert_eq!(diagnostic["ruleId"], rule);
                    assert_eq!(diagnostic["severity"], 2);
                    assert!(diagnostic["endColumn"].as_u64().is_some());
                    assert!(diagnostic.get("suggestions").is_none());
                }
            }
        }
    }
}

#[test]
fn duplicate_reports_typed_identity_and_keeps_trivia_in_message_and_fix() {
    let source = "type T = string | string;";
    let diagnostics = messages(DUPLICATE, source);
    assert_eq!(diagnostics.len(), 1);
    let report = &diagnostics[0];
    assert_eq!(report["messageId"], "duplicate");
    assert_eq!(
        report["message"],
        "Union type constituent is duplicated with  string."
    );
    assert_eq!(report["column"], 19);
    assert_eq!(report["endColumn"], 25);
    assert_eq!(report["fix"], json!({"range":[16,24],"text":" "}));
    let fixed = lint_files(DUPLICATE, &[("input.ts", source)], true);
    assert_eq!(fixed[0]["output"], "type T = string  ;");
    assert_eq!(fixed[0]["messages"], json!([]));
    let source = "type Value = string; type T = Value | string;";
    let diagnostics = messages(DUPLICATE, source);
    assert_eq!(
        diagnostics.len(),
        1,
        "alias equality must come from the checker"
    );
    assert_eq!(
        diagnostics[0]["message"],
        "Union type constituent is duplicated with  Value."
    );
    assert!(
        messages(DUPLICATE, "type T = Missing | Missing;").is_empty(),
        "unresolved intrinsic error types must not count as equal constituents"
    );
}

#[test]
fn duplicate_optional_undefined_and_nested_parentheses_have_atomic_edits() {
    let source = "function f(value?: string | undefined) {}";
    let diagnostics = messages(DUPLICATE, source);
    assert_eq!(diagnostics.len(), 1);
    assert_eq!(diagnostics[0]["messageId"], "unnecessary");
    assert_eq!(
        diagnostics[0]["message"],
        "Explicit undefined is unnecessary on an optional parameter."
    );
    let fixed = lint_files(DUPLICATE, &[("input.ts", source)], true);
    assert_eq!(fixed[0]["output"], "function f(value?: string  ) {}");
    assert_eq!(fixed[0]["messages"], json!([]));
    let fixed = lint_files(
        DUPLICATE,
        &[("input.ts", "type T = string | (((string)));")],
        true,
    );
    assert_eq!(fixed[0]["output"], "type T = string  ;");
    assert_eq!(fixed[0]["messages"], json!([]));
}

#[test]
fn constituent_queries_resolve_imported_aliases_across_source_files() {
    let files = [
        ("types.ts", "export type Name = string;"),
        (
            "input.ts",
            "import type { Name } from './types'; type T = Name | string;",
        ),
    ];
    let results = lint_files(DUPLICATE, &files, false);
    assert_eq!(results[0]["messages"], json!([]));
    assert_eq!(results[1]["messages"].as_array().unwrap().len(), 1);
    assert_eq!(results[1]["messages"][0]["messageId"], "duplicate");
    let files = [
        ("types.ts", "export type Name = 'hello';"),
        (
            "input.ts",
            "import type { Name } from './types'; type T = Name | string;",
        ),
    ];
    let results = lint_files(REDUNDANT, &files, false);
    assert_eq!(results[0]["messages"], json!([]));
    assert_eq!(results[1]["messages"].as_array().unwrap().len(), 1);
    assert_eq!(
        results[1]["messages"][0]["message"],
        "\"hello\" is overridden by string in this union type."
    );
}

#[test]
fn redundant_literals_preserve_messages_ranges_and_have_no_automatic_fix() {
    let source = "const emoji = '😀'; type T = string | 'hello';";
    let diagnostics = messages(REDUNDANT, source);
    assert_eq!(diagnostics.len(), 1);
    let report = &diagnostics[0];
    assert_eq!(report["messageId"], "literalOverridden");
    assert_eq!(
        report["message"],
        "\"hello\" is overridden by string in this union type."
    );
    let start = source.find("'hello'").unwrap();
    assert_eq!(report["column"], utf16(source, start) + 1);
    assert_eq!(
        report["endColumn"],
        utf16(source, start + "'hello'".len()) + 1
    );
    assert!(report.get("fix").is_none());
    let fixed = lint_files(REDUNDANT, &[("input.ts", source)], true);
    assert!(fixed[0].get("output").is_none());
    assert_eq!(fixed[0]["messages"], json!(diagnostics));
    assert_eq!(
        messages(REDUNDANT, "type T = boolean | true;")[0]["message"],
        "true is overridden by boolean in this union type."
    );
    assert_eq!(
        messages(REDUNDANT, "type B = boolean; type T = B | false;")[0]["message"],
        "false is overridden by boolean in this union type."
    );
    // The frozen intersection branch uses exact flags, unlike the union mask.
    assert!(messages(REDUNDANT, "type B = boolean; type T = B & false;").is_empty());
    assert!(messages(REDUNDANT, "type B = true | false; type T = B | false;").is_empty());
    assert_eq!(
        messages(REDUNDANT, "type T = number & 1;")[0]["message"],
        "number is overridden by the 1 in this intersection type."
    );
    assert_eq!(
        messages(REDUNDANT, "type T = (1 | 2) & number;")[0]["message"],
        "1 | 2 is overridden by the number in this intersection type."
    );
}

#[test]
fn redundant_top_bottom_and_return_type_exceptions_match_the_oracle() {
    for (source, id, message) in [
        (
            "type T = any | string;",
            "overrides",
            "'any' overrides all other types in this union type.",
        ),
        (
            "type T = unknown | string;",
            "overrides",
            "'unknown' overrides all other types in this union type.",
        ),
        (
            "type T = never | string;",
            "overridden",
            "'never' is overridden by other types in this union type.",
        ),
        (
            "type T = never & string;",
            "overrides",
            "'never' overrides all other types in this intersection type.",
        ),
        (
            "type T = unknown & string;",
            "overridden",
            "'unknown' is overridden by other types in this intersection type.",
        ),
    ] {
        let diagnostics = messages(REDUNDANT, source);
        assert_eq!(diagnostics.len(), 1, "{source}");
        assert_eq!(diagnostics[0]["messageId"], id);
        assert_eq!(diagnostics[0]["message"], message);
    }
    assert!(messages(REDUNDANT, "type F = () => string | never;").is_empty());
    assert!(messages(REDUNDANT, "interface F { new (): string | never; }").is_empty());
    assert_eq!(
        messages(REDUNDANT, "type F = () => (string | never);").len(),
        1,
        "the frozen exception checks only the immediate function-like parent"
    );
}
