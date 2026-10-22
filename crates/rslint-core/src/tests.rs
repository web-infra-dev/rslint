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
    let syntax: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(syntax["files"][0]["hasSyntaxErrors"], true);
    assert_eq!(syntax["files"][0]["appliedRules"], json!([]));
    assert_eq!(
        syntax["files"][0]["messages"],
        json!([{
            "ruleId":"TypeScript(TS8010)", "severity":2, "message":"Type annotations can only be used in TypeScript files.",
            "line":1,"column":14,"endLine":1,"endColumn":20,"range":[13,19]
        }])
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

#[test]
fn successive_snapshots_recheck_unchanged_importers_and_release_handles() {
    // Exercise the exact observation path used by automatic fixing, without
    // adding a fake production rule that changes exported declaration types.
    std::thread::Builder::new()
        .stack_size(512 << 20)
        .spawn(|| {
            let request: Request = serde_json::from_value(request(
                json!([]),
                json!([{"name":"@typescript-eslint/no-unsafe-unary-minus"}]),
            ))
            .unwrap();
            let inputs: Vec<String> = vec![
                "/rslint-rust-test/types.ts".into(),
                "/rslint-rust-test/main.ts".into(),
            ];
            let session = SessionOwner::new(&request.cwd);
            let mut overlays = BTreeMap::from([
                (
                    inputs[0].clone(),
                    "export declare const value: number;".into(),
                ),
                (
                    inputs[1].clone(),
                    "import { value } from './types'; -value;".into(),
                ),
            ]);
            for (declaration, expected) in [("number", 0), ("string", 1), ("bigint", 0)] {
                overlays.insert(
                    inputs[0].clone(),
                    format!("export declare const value: {declaration};"),
                );
                let observed = observe(&session, &request, &inputs, &overlays).unwrap();
                assert_eq!(observed.len(), 2);
                assert_eq!(observed[1].result.messages.len(), expected);
                assert_eq!(observed[1].source, overlays[&inputs[1]]);
            }
        })
        .unwrap()
        .join()
        .unwrap();
}

#[test]
fn automatic_fixes_keep_disk_marks_project_overlays_and_final_positions() {
    use std::sync::atomic::{AtomicUsize, Ordering};
    static NEXT: AtomicUsize = AtomicUsize::new(0);
    let temp = std::env::temp_dir().join(format!(
        "rslint-rust-fix-{}-{}",
        std::process::id(),
        NEXT.fetch_add(1, Ordering::Relaxed)
    ));
    std::fs::create_dir_all(&temp).unwrap();
    struct Cleanup(std::path::PathBuf);
    impl Drop for Cleanup {
        fn drop(&mut self) {
            let _ = std::fs::remove_dir_all(&self.0);
        }
    }
    let _cleanup = Cleanup(temp.clone());
    let source = "\u{feff}'😀'; value!!!!; debugger; object?.prop!;\r\n";
    let expected = "\u{feff}'😀'; value!; debugger; object?.prop!;\r\n";
    let mut disk_files = vec![("utf8.ts", source.as_bytes().to_vec())];
    for (name, big_endian) in [("utf16le.ts", false), ("utf16be.ts", true)] {
        let bytes = source
            .encode_utf16()
            .flat_map(|unit| {
                if big_endian {
                    unit.to_be_bytes()
                } else {
                    unit.to_le_bytes()
                }
            })
            .collect();
        disk_files.push((name, bytes));
    }
    for (name, bytes) in &disk_files {
        std::fs::write(temp.join(name), bytes).unwrap();
    }
    std::fs::write(
        temp.join("tsconfig.json"),
        r#"{"compilerOptions":{"strict":true},"include":["*.ts"]}"#,
    )
    .unwrap();
    let extra = "@typescript-eslint/no-extra-non-null-assertion";
    let optional = "@typescript-eslint/no-non-null-asserted-optional-chain";
    let input = json!({"schemaVersion":1,"cwd":temp.to_str().unwrap(),"project":"tsconfig.json","fix":true,
        "files":[{"path":"utf8.ts"},{"path":"utf16le.ts"},{"path":"utf16be.ts"},{"path":"overlay.ts","text":"export const clean = 1;"}],
        "rules":[{"name":extra},{"name":"no-debugger"},{"name":optional}]});
    let value: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(value["files"].as_array().unwrap().len(), 4);
    for (index, (name, bytes)) in disk_files.iter().enumerate() {
        assert_eq!(value["files"][index]["output"], expected);
        let messages = value["files"][index]["messages"].as_array().unwrap();
        assert_eq!(messages.len(), 2);
        assert_eq!(messages[0]["ruleId"], "no-debugger");
        assert_eq!(messages[0]["column"], 15);
        assert_eq!(messages[1]["ruleId"], optional);
        assert!(messages[1].get("fix").is_none());
        assert_eq!(std::fs::read(temp.join(name)).unwrap(), *bytes);
    }
    assert!(value["files"][3].get("output").is_none());
    assert_eq!(value["files"][3]["messages"], json!([]));
}

#[cfg(unix)]
#[test]
fn fix_overlays_preserve_symlink_module_identity() {
    let temp = std::env::temp_dir().join(format!("rslint-rust-symlink-{}", std::process::id()));
    std::fs::create_dir_all(&temp).unwrap();
    struct Cleanup(std::path::PathBuf);
    impl Drop for Cleanup {
        fn drop(&mut self) {
            let _ = std::fs::remove_dir_all(&self.0);
        }
    }
    let _cleanup = Cleanup(temp.clone());
    std::fs::write(temp.join("real.ts"), "export enum E { A }").unwrap();
    std::os::unix::fs::symlink(temp.join("real.ts"), temp.join("link.ts")).unwrap();
    let consumer =
        "import { E as A } from './real'; import { E as B } from './link'; A.A === B.A; value!!;";
    std::fs::write(temp.join("main.ts"), consumer).unwrap();
    let rule = "@typescript-eslint/no-unsafe-enum-comparison";
    let mut input = json!({"schemaVersion":1,"cwd":temp.to_str().unwrap(),
        "files":[{"path":"link.ts"},{"path":"main.ts"}],"rules":[{"name":rule}]});
    let initial: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    input["fix"] = json!(true);
    let fixed: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(initial, fixed);
    input["rules"]
        .as_array_mut()
        .unwrap()
        .push(json!({"name":"@typescript-eslint/no-extra-non-null-assertion"}));
    let fixed: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(
        fixed["files"][1]["output"],
        consumer.replace("value!!", "value!")
    );
    assert_eq!(
        fixed["files"][1]["messages"],
        initial["files"][1]["messages"]
    );
    assert!(fixed["files"][0].get("output").is_none());

    // Import through a second alias that is NOT part of the explicit inputs.
    // Its reads must see the edited canonical source, even when the compiler
    // keeps symlink module names distinct under preserveSymlinks.
    std::os::unix::fs::symlink(temp.join("real.ts"), temp.join("other.ts")).unwrap();
    std::fs::write(temp.join("real.ts"), "export const value = 1;").unwrap();
    for preserve in [false, true] {
        let input = json!({"schemaVersion":1,"cwd":temp.to_str().unwrap(),"fix":true,
            "compilerOptions":{"strict":true,"preserveSymlinks":preserve,"moduleResolution":100,"module":99},
            "files":[{"path":"link.ts","text":"export const value = 'text';"},
                {"path":"main.ts","text":"import { value } from './other'; -value; value!!;"}],
            "rules":[{"name":"@typescript-eslint/no-unsafe-unary-minus"},{"name":"@typescript-eslint/no-extra-non-null-assertion"}]});
        let result: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
        assert_eq!(result["files"][1]["messages"].as_array().unwrap().len(), 1);
        assert_eq!(
            result["files"][1]["messages"][0]["message"],
            "Argument of unary negation should be assignable to number | bigint but is \"text\" instead."
        );
        assert_eq!(
            result["files"][1]["output"],
            "import { value } from './other'; -value; value!;"
        );
    }
}
