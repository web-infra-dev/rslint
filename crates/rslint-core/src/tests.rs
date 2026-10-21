use super::*;

fn request(files: Value, rules: Value) -> Value {
    json!({"schemaVersion":1,"cwd":"/rslint-rust-test","files":files,"rules":rules})
}

#[test]
fn cross_file_types_and_silent_files_use_the_native_checker() {
    let input = request(
        json!([
            {"path":"types.ts","text":"export declare const unsafe: string; export declare const safe: number;"},
            {"path":"main.ts","text":"import { unsafe, safe } from './types';\n-unsafe; -safe; debugger;"},
            {"path":"clean.ts","text":"export const value = -1;"}
        ]),
        json!([{"name":"no-debugger"},{"name":"@typescript-eslint/no-unsafe-unary-minus"}]),
    );
    let result: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    assert_eq!(result["files"].as_array().unwrap().len(), 3);
    assert_eq!(result["files"][0]["messages"], json!([]));
    assert_eq!(result["files"][2]["messages"], json!([]));
    let messages = result["files"][1]["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 2);
    assert_eq!(
        messages[0]["message"],
        "Argument of unary negation should be assignable to number | bigint but is string instead."
    );
    assert_eq!(messages[0]["line"], 2);
    assert_eq!(messages[0]["column"], 1);
    assert_eq!(messages[0]["endColumn"], 8);
    assert_eq!(messages[1]["ruleId"], "no-debugger");
}

#[test]
fn positions_use_javascript_utf16_and_line_terminators() {
    assert_eq!(
        position("😀\r\n\u{2028}x", "😀\r\n\u{2028}".len()),
        (5, 3, 1)
    );
    let input = request(
        json!([{"path":"a.ts","text":"'😀'; debugger;\r\ndebugger;"}]),
        json!([{"name":"no-debugger"}]),
    );
    let value: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(value["files"][0]["messages"][0]["column"], 7);
    assert_eq!(value["files"][0]["messages"][0]["range"], json!([6, 15]));
    assert_eq!(value["files"][0]["messages"][1]["line"], 2);
    let input = request(
        json!([{"path":"bom.ts","text":"\u{feff}debugger;\r\n'😀'; debugger;"}]),
        json!([{"name":"no-debugger"}]),
    );
    let value: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(value["files"][0]["messages"][0]["column"], 1);
    assert_eq!(value["files"][0]["messages"][0]["range"], json!([0, 9]));
    assert_eq!(value["files"][0]["messages"][1]["column"], 7);
    assert_eq!(value["files"][0]["messages"][1]["range"], json!([17, 26]));
}

#[test]
fn unsupported_rules_and_missing_admissions_fail_instead_of_skipping() {
    let mut input = request(
        json!([{"path":"a.ts","text":"debugger;"}]),
        json!([{"name":"not-migrated"}]),
    );
    assert!(
        lint_json(&input.to_string())
            .unwrap_err()
            .contains("not been migrated")
    );
    input["rules"] = json!([{"name":"no-debugger"}]);
    input["files"][0] = json!({"path":"missing.ts"});
    assert!(
        lint_json(&input.to_string())
            .unwrap_err()
            .contains("missing from compiler project")
    );
    input["files"] = json!([{"path":"a.ts","text":""},{"path":"./a.ts","text":""}]);
    assert!(
        lint_json(&input.to_string())
            .unwrap_err()
            .contains("duplicate admitted file")
    );
    input["files"] = json!([{"path":"bad.js","text":"const value: number = 1; debugger;"}]);
    assert!(
        lint_json(&input.to_string())
            .unwrap_err()
            .contains("syntax diagnostics")
    );
    input["files"] = json!([{"path":"good.ts","text":"debugger;"}]);
    input["project"] = json!("tsconfig.json");
    input["compilerOptions"] = json!({"strict":false});
    assert!(
        lint_json(&input.to_string())
            .unwrap_err()
            .contains("cannot be combined")
    );
}
