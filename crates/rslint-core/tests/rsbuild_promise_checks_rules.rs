//! Frozen Promise-rule inputs and exact diagnostic/suggestion contracts.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

const FLOATING: &str = "@typescript-eslint/no-floating-promises";
const MISUSED: &str = "@typescript-eslint/no-misused-promises";

fn compiler_options() -> Value {
    json!({"jsx":"preserve","target":"ESNext","module":"CommonJS",
        "moduleResolution":"Node10","strict":true,"strictNullChecks":true,
        "esModuleInterop":true,"lib":["es2015","es2017","esnext","DOM"],
        "experimentalDecorators":true,"types":[]})
}

fn lint(rule: &str, entry: &Value, fix: bool) -> Value {
    let root = "/rslint-rsbuild-promise-tests";
    let default_name = format!("input{}", entry["extension"].as_str().unwrap_or(".ts"));
    let name = entry["fileName"].as_str().unwrap_or(&default_name);
    let source = entry["text"].as_str().unwrap();
    let mut sources = Vec::new();
    if let Some(files) = entry["files"].as_object() {
        for (path, source) in files {
            sources.push(json!({"path":path,"text":source}));
        }
    }
    let mut roots = vec![name.to_string()];
    if let Some(files) = entry["projectFiles"].as_array() {
        roots.extend(files.iter().map(|path| path.as_str().unwrap().to_string()));
    }
    roots.sort();
    roots.dedup();
    sources.push(json!({"path":"tsconfig.json","text":json!({
        "compilerOptions":entry.get("compilerOptions").cloned().unwrap_or_else(compiler_options),
        "files":roots
    }).to_string()}));
    let request = json!({"schemaVersion":1,"cwd":root,"sources":sources,"fix":fix,
        "files":[{"path":name,"text":source,"rules":[{"name":rule}],
        "policy":{"kind":"explicit","configFiles":[format!("{root}/tsconfig.json")]}}]});
    let result: Value =
        serde_json::from_str(&lint_projects_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let result = result["files"][0].clone();
    assert_eq!(result["hasTypeInformation"], true);
    assert_eq!(result["appliedRules"], json!([rule]));
    result
}

fn messages(rule: &str, source: &str) -> Vec<Value> {
    lint(rule, &json!({"text":source}), false)["messages"]
        .as_array()
        .unwrap()
        .clone()
}

fn suggestion_output(source: &str, suggestion: &Value) -> String {
    let units = source.encode_utf16().collect::<Vec<_>>();
    let start = suggestion["fix"]["range"][0].as_u64().unwrap() as usize;
    let end = suggestion["fix"]["range"][1].as_u64().unwrap() as usize;
    String::from_utf16(&units[..start]).unwrap()
        + suggestion["fix"]["text"].as_str().unwrap()
        + &String::from_utf16(&units[end..]).unwrap()
}

fn expected_matches(source: &str, expected: &Value, actual: &Value) -> bool {
    for field in [
        "messageId",
        "message",
        "line",
        "column",
        "endLine",
        "endColumn",
    ] {
        if let Some(value) = expected.get(field)
            && value != &actual[field]
        {
            return false;
        }
    }
    let expected_suggestions = expected["suggestions"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    let actual_suggestions = actual["suggestions"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    expected_suggestions.len() == actual_suggestions.len()
        && expected_suggestions
            .iter()
            .zip(actual_suggestions.iter())
            .all(|(expected, actual)| {
                expected["messageId"] == actual["messageId"]
                    && expected["output"].as_str()
                        == Some(suggestion_output(source, actual).as_str())
            })
}

#[test]
fn frozen_default_promise_cases_preserve_positions_and_suggestion_outputs() {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-promise-checks.json"
    ))
    .unwrap();
    let groups = groups.as_array().unwrap();
    assert_eq!(groups.len(), 2);
    assert_eq!(groups[0]["valid"].as_array().unwrap().len(), 46);
    assert_eq!(groups[0]["invalid"].as_array().unwrap().len(), 65);
    assert_eq!(groups[1]["valid"].as_array().unwrap().len(), 100);
    assert_eq!(groups[1]["invalid"].as_array().unwrap().len(), 91);
    assert_eq!(groups[0]["originalSkip"].as_array().unwrap().len(), 0);
    assert_eq!(groups[1]["originalSkip"].as_array().unwrap().len(), 1);
    assert_eq!(groups[0]["excludedOptions"].as_array().unwrap().len(), 75);
    assert_eq!(groups[1]["excludedOptions"].as_array().unwrap().len(), 26);
    for group in groups {
        let rule = group["rule"].as_str().unwrap();
        for category in ["valid", "invalid"] {
            for (index, entry) in group[category].as_array().unwrap().iter().enumerate() {
                let result = lint(rule, entry, false);
                let mut actual = result["messages"].as_array().unwrap().clone();
                let source = entry["text"].as_str().unwrap();
                let expected = entry["expected"].as_array().cloned().unwrap_or_default();
                assert_eq!(
                    actual.len(),
                    expected.len(),
                    "{rule}/{category}/{index}: {source}: {actual:?}"
                );
                assert_eq!(
                    actual.is_empty(),
                    category == "valid",
                    "{rule}/{category}/{index}"
                );
                for expected in expected {
                    let found = actual
                        .iter()
                        .position(|actual| expected_matches(source, &expected, actual));
                    assert!(
                        found.is_some(),
                        "{rule}/{category}/{index}: {source}: expected {expected:?}; remaining {actual:?}"
                    );
                    let report = actual.remove(found.unwrap());
                    assert_eq!(report["ruleId"], rule);
                    assert_eq!(report["severity"], 2);
                    assert!(
                        report.get("fix").is_none(),
                        "both rules only offer suggestions"
                    );
                }
            }
        }
    }
}

#[test]
fn floating_messages_and_suggestions_use_utf16_and_preserve_parentheses() {
    let source = "const emoji = '😀'; (Promise.resolve(1));";
    let reports = messages(FLOATING, source);
    assert_eq!(reports.len(), 1);
    let report = &reports[0];
    assert_eq!(report["messageId"], "floatingVoid");
    assert_eq!(
        report["message"],
        "Promises must be awaited, end with a call to .catch, end with a call to .then with a rejection handler or be explicitly marked as ignored with the `void` operator."
    );
    assert_eq!(report["line"], 1);
    assert_eq!(
        report["column"],
        source[..source.find("(Promise").unwrap()]
            .encode_utf16()
            .count()
            + 1
    );
    assert_eq!(
        report["suggestions"][0]["desc"],
        "Add void operator to ignore."
    );
    assert_eq!(report["suggestions"][1]["desc"], "Add await operator.");
    assert_eq!(
        suggestion_output(source, &report["suggestions"][0]),
        "const emoji = '😀'; void (Promise.resolve(1));"
    );
    assert_eq!(
        suggestion_output(source, &report["suggestions"][1]),
        "const emoji = '😀'; await (Promise.resolve(1));"
    );
    assert_eq!(
        lint(FLOATING, &json!({"text":source}), true)["messages"],
        json!(reports)
    );
    assert!(
        lint(FLOATING, &json!({"text":source}), true)
            .get("output")
            .is_none()
    );
}

#[test]
fn floating_distinguishes_arrays_rejection_handlers_and_library_promises() {
    let reports = messages(
        FLOATING,
        "[Promise.resolve()]; Promise.resolve().catch(1); void Promise.resolve();",
    );
    assert_eq!(reports.len(), 2);
    assert_eq!(reports[0]["messageId"], "floatingPromiseArrayVoid");
    assert!(
        reports[0]["suggestions"]
            .as_array()
            .is_none_or(Vec::is_empty)
    );
    assert_eq!(
        reports[1]["messageId"],
        "floatingUselessRejectionHandlerVoid"
    );
    assert!(
        reports[1]["message"]
            .as_str()
            .unwrap()
            .ends_with("A rejection handler that is not a function will be ignored.")
    );
    let source = "export {}; class Promise { then(callback: () => void): this { return this; } } new Promise();";
    assert!(
        messages(FLOATING, source).is_empty(),
        "a local Promise class is not the library Promise"
    );
    assert!(
        messages(FLOATING, "class Derived extends Promise<void> {} Derived;").is_empty(),
        "Promise constructors are not Promise instances"
    );
}

#[test]
fn cross_file_callback_overloads_and_promise_declarations_use_the_selected_program() {
    let entry = json!({"text":"import { sink, legal, value } from './dependency'; sink(async () => {}); legal(async () => {}); if (value) {}",
        "files":{"dependency.ts":"export declare function sink(cb: () => void): void; export declare function legal(cb: () => void): void; export declare function legal(cb: () => Promise<void>): void; export declare const value: Promise<void>;"}});
    let result = lint(MISUSED, &entry, false);
    let reports = result["messages"].as_array().unwrap();
    assert_eq!(
        reports
            .iter()
            .map(|report| report["messageId"].as_str().unwrap())
            .collect::<Vec<_>>(),
        vec!["voidReturnArgument", "conditional"]
    );
    assert_eq!(
        reports[0]["message"],
        "Promise returned in function argument where a void return was expected."
    );
    assert_eq!(
        reports[1]["message"],
        "Expected non-Promise value in a boolean conditional."
    );
    let handled = json!({"text":"import { task } from './dependency'; task().catch(() => {}); void task();",
        "files":{"dependency.ts":"export async function task(): Promise<void> {}"}});
    assert!(
        lint(FLOATING, &handled, false)["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
    let floating = json!({"text":"import { task } from './dependency'; task();",
        "files":{"dependency.ts":"export async function task(): Promise<void> {}"}});
    assert_eq!(
        lint(FLOATING, &floating, false)["messages"][0]["messageId"],
        "floatingVoid"
    );
}

#[test]
fn misused_avoids_duplicate_predicates_and_checks_exact_context_nodes() {
    let source = "const pred = { test: async () => true }; [1].filter(pred.test);";
    let reports = messages(MISUSED, source);
    assert_eq!(reports.len(), 1);
    assert_eq!(reports[0]["messageId"], "predicate");
    assert_eq!(
        reports[0]["range"],
        json!([
            source.find("pred.test").unwrap(),
            source.find("pred.test").unwrap() + 9
        ])
    );
    assert!(messages(MISUSED, "Promise.resolve().finally(async () => {});").is_empty());
    assert!(
        messages(
            MISUSED,
            "declare const optional: Promise<void> | undefined; if (optional) {}"
        )
        .is_empty()
    );
    let jsx = lint(
        MISUSED,
        &json!({"extension":".tsx","text":"declare function Component(props: { callback: () => void }): null; <Component callback={async () => {}} />;"}),
        false,
    );
    let source = "declare function Component(props: { callback: () => void }): null; <Component callback={async () => {}} />;";
    let report = &jsx["messages"][0];
    assert_eq!(report["messageId"], "voidReturnAttribute");
    assert_eq!(
        report["range"],
        json!([
            source.find("{async").unwrap(),
            source.rfind('}').unwrap() + 1
        ])
    );
}
