use rslint_core::lint_projects_json;
use serde_json::{Value, json};

const CWD: &str = "/rslint-configured-rules";
fn run(files: Value, sources: Value, fix: bool) -> Value {
    let request = json!({"schemaVersion":1,"cwd":CWD,"files":files,"sources":sources,"fix":fix});
    serde_json::from_str(&lint_projects_json(&request.to_string()).unwrap()).unwrap()
}
fn input(name: &str, text: &str, rule: &str, typed: bool) -> Value {
    json!({"path":name,"text":text,"rules":[{"name":rule}],"policy":if typed {json!({"kind":"service","rootDirectory":CWD})} else {json!({"kind":"sourceOnly"})}})
}
fn config(roots: &[&str]) -> Value {
    json!({"path":"tsconfig.json","text":json!({"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler"},"files":roots}).to_string()})
}

#[test]
fn independent_configurations_keep_order_severities_and_quiet_files() {
    let mut warning = input("second.ts", "debugger;", "no-debugger", false);
    warning["rules"][0]["severity"] = json!(1);
    let result = run(
        json!([
            input("first.ts", "debugger;", "no-empty", false),
            warning,
            input(
                "quiet.ts",
                "export const answer = 42;",
                "no-debugger",
                false
            )
        ]),
        json!([]),
        false,
    );
    let files = result["files"].as_array().unwrap();
    assert_eq!(files.len(), 3);
    assert_eq!(files[0]["appliedRules"], json!(["no-empty"]));
    assert_eq!(files[0]["messages"], json!([]));
    assert_eq!(files[1]["filePath"], format!("{CWD}/second.ts"));
    assert_eq!(files[1]["messages"][0]["severity"], 1);
    assert_eq!(files[2]["messages"], json!([]));
}

#[test]
fn imported_source_is_not_promoted_to_a_typed_lint_root() {
    let rule = "@typescript-eslint/no-unsafe-unary-minus";
    let text = "export const value = 'text'; -value;";
    let result = run(
        json!([
            input(
                "main.ts",
                "import {value} from './values'; -value;",
                rule,
                true
            ),
            input("values.ts", text, rule, true),
            input("silent.ts", "export {};", "no-debugger", false)
        ]),
        json!([config(&["main.ts"])]),
        false,
    );
    assert_eq!(result["files"][0]["appliedRules"], json!([rule]));
    assert_eq!(result["files"][0]["messages"][0]["messageId"], "unaryMinus");
    assert_eq!(result["files"][1]["appliedRules"], json!([]));
    assert_eq!(result["files"][1]["messages"], json!([]));
    assert_eq!(result["files"].as_array().unwrap().len(), 3);
}

#[test]
fn fixes_rebuild_cross_file_types_and_keep_suggestion_only_input_unchanged() {
    let mut fixed = input(
        "values.ts",
        "export const value = (1 as number | null)!!;",
        "@typescript-eslint/no-extra-non-null-assertion",
        true,
    );
    fixed["rules"] = json!([{"name":"@typescript-eslint/no-extra-non-null-assertion"}]);
    let result = run(
        json!([
            fixed,
            input(
                "main.ts",
                "import { value } from './values'; -value;",
                "@typescript-eslint/no-unsafe-unary-minus",
                true
            ),
            input(
                "suggestion.ts",
                "type Empty = {};",
                "@typescript-eslint/no-empty-object-type",
                false
            )
        ]),
        json!([config(&["values.ts", "main.ts"])]),
        true,
    );
    assert_eq!(
        result["files"][0]["output"],
        "export const value = (1 as number | null)!;"
    );
    assert_eq!(result["files"][0]["messages"], json!([]));
    assert_eq!(result["files"][1]["messages"], json!([]));
    assert_eq!(
        result["files"][2]["messages"][0]["suggestions"]
            .as_array()
            .unwrap()
            .len(),
        2
    );
    assert!(result["files"][2].get("output").is_none());
}

#[test]
fn syntax_errors_do_not_skip_other_files_and_source_only_keeps_binding() {
    let result = run(
        json!([
            input("broken.ts", "let = ;", "no-debugger", false),
            input(
                "bound.ts",
                "const Promise = custom; new Promise(async () => {});",
                "no-async-promise-executor",
                false
            ),
            input(
                "hit.ts",
                "new Promise(async () => {});",
                "no-async-promise-executor",
                false
            )
        ]),
        json!([]),
        false,
    );
    assert_eq!(result["files"][0]["hasSyntaxErrors"], true);
    assert_eq!(result["files"][0]["appliedRules"], json!([]));
    assert_eq!(result["files"][1]["messages"], json!([]));
    assert_eq!(result["files"][2]["messages"].as_array().unwrap().len(), 1);
}

#[test]
fn unknown_rules_and_options_fail_even_without_type_information() {
    for rules in [
        json!([{"name":"unknown/plugin-rule"}]),
        json!([{"name":"no-debugger","options":[true]}]),
    ] {
        let mut file = input("input.ts", "export {};", "no-debugger", false);
        file["rules"] = rules;
        assert!(
            lint_projects_json(&json!({"schemaVersion":1,"cwd":CWD,"files":[file]}).to_string())
                .is_err()
        );
    }
}
