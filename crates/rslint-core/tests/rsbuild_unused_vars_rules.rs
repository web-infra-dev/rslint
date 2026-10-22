//! Frozen no-unused-vars contracts: symbol identity, read/write classification,
//! type-space references, ambient scopes, exact ignore options and suggestions.
use rslint_core::{lint_json, lint_projects_json};
use serde_json::{Value, json};
const RULE: &str = "@typescript-eslint/no-unused-vars";

fn lint(text: &str, extension: &str, options: &Value, fix: bool) -> Value {
    let cwd = "/rslint-rsbuild-unused-vars";
    let name = format!("case{extension}");
    let config = format!("{cwd}/tsconfig.json");
    let request = json!({"schemaVersion":1,"cwd":cwd,
        "files":[{"path":name,"text":text,"rules":[{"name":RULE,"options":options}],"policy":{"kind":"explicit","configFiles":[config]}}],
        "sources":[{"path":config,"text":json!({"files":[name],"compilerOptions":{"strict":true,"allowJs":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler","jsx":"preserve"}}).to_string()}],"fix":fix});
    let result: Value = serde_json::from_str(
        &lint_projects_json(&request.to_string()).unwrap_or_else(|error| panic!("{error}\n{text}")),
    )
    .unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let file = result["files"][0].clone();
    assert_eq!(file["appliedRules"], json!([RULE]));
    assert!(
        file.get("output").is_none(),
        "default no-unused-vars offers suggestions, never an automatic edit"
    );
    for message in file["messages"].as_array().unwrap() {
        assert_eq!(message["ruleId"], RULE);
        assert_eq!(message["severity"], 2);
        assert!(message.get("fix").is_none());
    }
    file
}
fn corpus(options: Value) {
    let suites: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-unused-vars.json"
    ))
    .unwrap();
    let suite = suites
        .as_array()
        .unwrap()
        .iter()
        .find(|suite| suite["options"] == options)
        .unwrap();
    let mut failures = Vec::new();
    for category in ["valid", "invalid"] {
        assert!(!suite[category].as_array().unwrap().is_empty());
        for case in suite[category].as_array().unwrap() {
            let text = case["text"].as_str().unwrap();
            let file = lint(
                text,
                case["extension"].as_str().unwrap_or(".ts"),
                &options,
                false,
            );
            let messages = file["messages"].as_array().unwrap();
            let mut expected = case["expected"].as_array().cloned().unwrap_or_default();
            expected.sort_by_key(|message| {
                (
                    message["line"].as_u64().unwrap_or(1),
                    message["column"].as_u64().unwrap_or(1),
                )
            });
            if messages.len() != expected.len() {
                failures.push(format!(
                    "{category}: {}\n{text}\nexpected: {expected:?}\nactual: {messages:?}",
                    case["origin"]
                ));
            }
            for (actual, expected) in messages.iter().zip(expected) {
                for (key, value) in expected.as_object().unwrap() {
                    if &actual[key] != value {
                        failures.push(format!(
                            "{}: {key}: expected {value}, actual {}\n{text}\n{actual}",
                            case["origin"], actual[key]
                        ));
                    }
                }
            }
        }
    }
    assert!(
        failures.is_empty(),
        "{} preserved assertion failures:\n{}",
        failures.len(),
        failures.join("\n\n")
    );
}
#[test]
fn frozen_default_symbol_and_scope_cases() {
    corpus(json!([]));
}
#[test]
fn frozen_vars_ignore_pattern_cases() {
    corpus(json!([{"varsIgnorePattern":"^_"}]));
}
#[test]
fn frozen_args_ignore_pattern_cases() {
    corpus(json!([{"argsIgnorePattern":"^_"}]));
}
#[test]
fn frozen_rsbuild_combined_options() {
    corpus(json!([{"varsIgnorePattern":"^_","argsIgnorePattern":"^_"}]));
}
#[test]
fn actual_options_keep_catch_bindings_and_append_exact_messages() {
    let options = json!([{"varsIgnorePattern":"^_","argsIgnorePattern":"^_"}]);
    let file = lint(
        "const unused = 1; export function run(arg: unknown) {} try {} catch (_error) {}",
        ".ts",
        &options,
        true,
    );
    let messages = file["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 3);
    assert_eq!(
        messages[0]["message"],
        "'unused' is assigned a value but never used. Allowed unused vars must match /^_/u."
    );
    assert_eq!(
        messages[1]["message"],
        "'arg' is defined but never used. Allowed unused args must match /^_/u."
    );
    assert_eq!(
        messages[2]["message"],
        "'_error' is defined but never used."
    );
}
#[test]
fn unused_imports_preserve_incremental_suggestion_ranges_without_autofixing() {
    let text = "import { first, second } from 'missing';\n";
    let file = lint(text, ".ts", &json!([]), true);
    let messages = file["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 2);
    assert_eq!(
        messages[0]["suggestions"][0]["messageId"],
        "removeUnusedVar"
    );
    assert_eq!(
        messages[0]["suggestions"][0]["desc"],
        "Remove unused variable \"first\"."
    );
    assert_eq!(
        messages[0]["suggestions"][0]["fix"]["range"],
        json!([9, 16])
    );
    assert_eq!(
        messages[1]["suggestions"][0]["messageId"],
        "removeUnusedImportDeclaration"
    );
    assert_eq!(
        messages[1]["suggestions"][0]["fix"]["range"],
        json!([0, text.len()])
    );
}
#[test]
fn rejects_unimplemented_option_semantics_instead_of_silently_ignoring_them() {
    for options in [
        json!([{"varsIgnorePattern":"^ignore"}]),
        json!([{"args":"none"}]),
        json!([{"reportUsedIgnorePattern":true}]),
        json!([{"enableAutofixRemoval":{"imports":true}}]),
    ] {
        let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-unused-vars","files":[{"path":"case.ts","text":"const unused = 1;"}],"rules":[{"name":RULE,"options":options}]});
        let error = lint_json(&request.to_string()).unwrap_err();
        assert!(error.contains("no-unused-vars"), "{error}");
        assert!(error.contains("not supported"), "{error}");
    }
}

#[test]
fn parameter_property_reads_set_the_after_used_boundary_without_counting_names() {
    let valid = "export class Example { constructor(value: string, private name: string) { console.log(this.name); } }";
    assert!(
        lint(valid, ".ts", &json!([]), false)["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
    let unused = "export class Example { constructor(value: string, private name: string) {} }";
    let file = lint(unused, ".ts", &json!([]), false);
    assert_eq!(file["messages"].as_array().unwrap().len(), 1);
    assert_eq!(
        file["messages"][0]["message"],
        "'value' is defined but never used."
    );
    let unrelated = "export class Example { constructor(value: string, private name: string) { console.log({name: 'unrelated'}.name); } }";
    let file = lint(unrelated, ".ts", &json!([]), false);
    assert_eq!(file["messages"].as_array().unwrap().len(), 1);
    assert_eq!(
        file["messages"][0]["message"],
        "'value' is defined but never used."
    );
}

#[test]
fn namespace_export_aliases_preserve_type_and_value_binding_selection() {
    let source =
        "type Token = {}; namespace Box { const Token = 1; export type { Token }; } consume(Box);";
    let file = lint(source, ".ts", &json!([]), false);
    let messages = file["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 1);
    assert_eq!(
        messages[0]["message"],
        "'Token' is assigned a value but never used."
    );
    assert_eq!(messages[0]["column"], 40);
    for source in [
        "const value = 1; export namespace N { export { value }; } consume(N);",
        "namespace Original { export const value = 1; } export namespace Public { export import Alias = Original; }",
    ] {
        assert!(
            lint(source, ".ts", &json!([]), false)["messages"]
                .as_array()
                .unwrap()
                .is_empty()
        );
    }
}
