//! Frozen no-async-mock-factory inputs plus exact suggestion and binding checks.
use rslint_core::lint_json;
use serde_json::{Value, json};
use std::{
    fs,
    path::PathBuf,
    sync::atomic::{AtomicUsize, Ordering},
};

const RULE: &str = "rstest/no-async-mock-factory";
static NEXT: AtomicUsize = AtomicUsize::new(0);
struct FixtureDir(PathBuf);
impl Drop for FixtureDir {
    fn drop(&mut self) {
        fs::remove_dir_all(&self.0).expect("remove test directory");
    }
}
fn lint(entry: &Value, fix: bool) -> Value {
    let text = entry.as_str().or_else(|| entry["text"].as_str()).unwrap();
    let root = FixtureDir(std::env::temp_dir().join(format!(
        "rslint-rstest-mock-{}-{}",
        std::process::id(),
        NEXT.fetch_add(1, Ordering::Relaxed)
    )));
    fs::create_dir(&root.0).unwrap();
    if let Some(files) = entry["files"].as_object() {
        for (name, contents) in files {
            let path = root.0.join(name);
            fs::create_dir_all(path.parent().unwrap()).unwrap();
            fs::write(path, contents.as_str().unwrap()).unwrap();
        }
    }
    fs::write(root.0.join("input.ts"), text).unwrap();
    fs::write(root.0.join("tsconfig.json"), json!({
        "compilerOptions": {"strict":true, "target":"ES2022", "module":"ESNext", "moduleResolution":"Bundler"},
        "files":["input.ts"]
    }).to_string()).unwrap();
    let request = json!({
        "schemaVersion":1, "cwd":root.0.to_str().unwrap(), "project":"tsconfig.json",
        "files":[{"path":"input.ts","text":text}], "rules":[{"name":RULE}], "fix":fix
    });
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(
        result["files"].as_array().unwrap().len(),
        1,
        "imports are not extra admitted lint files"
    );
    let file = result["files"][0].clone();
    assert_eq!(file["appliedRules"], json!([RULE]));
    assert!(file["hasSyntaxErrors"].is_null(), "{text}: {file}");
    assert_eq!(
        fs::read_to_string(root.0.join("input.ts")).unwrap(),
        text,
        "lint cannot mutate input files"
    );
    file
}
fn corpus() -> Value {
    serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-rstest-mock.json"
    ))
    .unwrap()
}
#[test]
fn frozen_valid_cases_keep_build_matching_and_non_promise_factories_silent() {
    let cases = corpus();
    let valid = cases[0]["valid"].as_array().unwrap();
    assert!(valid.len() >= 58);
    for entry in valid {
        let file = lint(entry, false);
        assert!(
            file["messages"].as_array().unwrap().is_empty(),
            "{entry}: {file}"
        );
    }
}
#[test]
fn frozen_invalid_cases_resolve_syntax_local_bindings_and_native_promise_types() {
    let cases = corpus();
    let invalid = cases[0]["invalid"].as_array().unwrap();
    assert!(invalid.len() >= 68);
    for entry in invalid {
        let file = lint(entry, false);
        let messages = file["messages"].as_array().unwrap();
        assert!(!messages.is_empty(), "{entry}: {file}");
        for message in messages {
            assert_eq!(message["ruleId"], RULE);
            assert_eq!(message["messageId"], "asyncMockFactory");
            assert_eq!(message["severity"], 2);
            assert!(
                message["fix"].is_null(),
                "suggestions must not be automatic fixes"
            );
        }
    }
}
#[test]
fn async_diagnostic_and_suggestion_match_frozen_exact_payload() {
    let source = "rs.mock('./sum', async () => ({ sum: () => 0 }))";
    let file = lint(&json!(source), false);
    assert_eq!(
        file["messages"],
        json!([{
            "ruleId":RULE, "messageId":"asyncMockFactory", "severity":2,
            "message":"Mock factory must return the module object synchronously; 'mock()' throws when the factory returns a Promise. To keep part of the original module, import it with `with { rstest: 'importActual' }` and spread it in.",
            "line":1,"column":18,"endLine":1,"endColumn":48,"range":[17,47],
            "suggestions":[{
                "messageId":"suggestRemoveAsync",
                "desc":"Remove 'async' so the factory returns the module object directly.",
                "fix":{"range":[17,23],"text":""}
            }]
        }])
    );
    assert_eq!(lint(&json!(source), true)["messages"], file["messages"]);
    assert!(lint(&json!(source), true)["output"].is_null());
}
#[test]
fn suggestions_keep_comments_type_annotations_and_nested_await_semantics() {
    for source in [
        "rs.mock('./m', async (): Promise<object> => ({}));",
        "rs.mock('./m', (): Promise<object> => Promise.resolve({}));",
        "rs.mock('./m', () => Promise.resolve(/* keep */ {}));",
        "rs.mock('./m', () => Promise.resolve({} /* keep */));",
        "rs.mock('./m', async () => ({ [await ready]() {} }));",
    ] {
        let file = lint(&json!(source), false);
        assert_eq!(file["messages"].as_array().unwrap().len(), 1);
        assert!(
            file["messages"][0]["suggestions"].is_null(),
            "{source}: {file}"
        );
    }
    let source = "rs.mock('./m', async () => ({ inner: async () => await ready }));";
    let file = lint(&json!(source), false);
    assert_eq!(
        file["messages"][0]["suggestions"][0]["messageId"],
        "suggestRemoveAsync"
    );
}
#[test]
fn unwrap_resolve_retains_arrow_grammar_and_utf16_offsets() {
    let source = "/* 😀 */ rs.mock('./m', () => Promise.resolve({ value: 1 }));";
    let file = lint(&json!(source), false);
    let start = source.find("Promise.resolve").unwrap();
    let end = source.rfind(")").unwrap();
    assert_eq!(
        file["messages"][0]["suggestions"][0],
        json!({
            "messageId":"suggestUnwrapPromiseResolve",
            "desc":"Return the module object directly instead of wrapping it in 'Promise.resolve()'.",
            "fix":{"range":[source[..start].encode_utf16().count(),source[..end].encode_utf16().count()],"text":"({ value: 1 })"}
        })
    );
}
#[test]
fn shared_factory_reports_each_call_and_each_receiver_member() {
    let source = "const factory = async () => ({}); rs.mock('./a', factory); rs.doMock('./b', factory); rstest.mockRequire('./c', factory); rstest.doMockRequire('./d', factory);";
    let file = lint(&json!(source), false);
    let messages = file["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 4);
    for (message, member) in messages
        .iter()
        .zip(["mock", "doMock", "mockRequire", "doMockRequire"])
    {
        assert!(
            message["message"]
                .as_str()
                .unwrap()
                .contains(&format!("'{member}()'"))
        );
        assert!(
            message["suggestions"].is_null(),
            "the identifier argument is not rewritten"
        );
    }
}
