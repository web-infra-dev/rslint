//! Exact frozen Rstest Promise diagnostics and control-flow regressions.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

const RULE: &str = "rstest/valid-expect-in-promise";

fn compiler_options() -> Value {
    json!({"jsx":"preserve","target":"ESNext","module":"CommonJS",
        "moduleResolution":"Node10","strict":true,"strictNullChecks":true,
        "esModuleInterop":true,"lib":["es2015","es2017","esnext","DOM"],
        "experimentalDecorators":true,"types":[]})
}

fn lint(rule: &str, entry: &Value, fix: bool) -> Value {
    let root = "/rslint-rstest-promise-tests";
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

#[test]
fn frozen_rstest_promise_cases_preserve_all_diagnostics_and_locations() {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-rstest-promise.json"
    ))
    .unwrap();
    let groups = groups.as_array().unwrap();
    assert_eq!(groups.len(), 1);
    let group = &groups[0];
    assert_eq!(group["rule"], RULE);
    assert_eq!(group["valid"].as_array().unwrap().len(), 57);
    assert_eq!(group["invalid"].as_array().unwrap().len(), 42);
    assert_eq!(group["originalSkip"], json!([]));
    assert_eq!(group["excludedOptions"], json!([]));
    for category in ["valid", "invalid"] {
        for (index, entry) in group[category].as_array().unwrap().iter().enumerate() {
            let result = lint(RULE, entry, false);
            let mut actual = result["messages"].as_array().unwrap().clone();
            let expected = entry["expected"].as_array().cloned().unwrap_or_default();
            assert_eq!(
                actual.len(),
                expected.len(),
                "{category}/{index}: {}: {actual:?}",
                entry["text"]
            );
            assert_eq!(actual.is_empty(), category == "valid");
            for expected in expected {
                let found = actual.iter().position(|actual| {
                    [
                        "messageId",
                        "message",
                        "line",
                        "column",
                        "endLine",
                        "endColumn",
                    ]
                    .iter()
                    .all(|field| {
                        expected
                            .get(field)
                            .is_none_or(|value| value == &actual[field])
                    })
                });
                assert!(
                    found.is_some(),
                    "{category}/{index}: {}: expected {expected:?}, remaining {actual:?}",
                    entry["text"]
                );
                let report = actual.remove(found.unwrap());
                assert_eq!(report["messageId"], "expectInFloatingPromise");
                assert_eq!(
                    report["message"],
                    "This promise should either be returned or awaited to ensure the assertions in its chain are called"
                );
                assert_eq!(report["ruleId"], RULE);
                assert_eq!(report["severity"], 2);
                assert!(report.get("fix").is_none());
                assert!(report["suggestions"].as_array().is_none_or(Vec::is_empty));
            }
        }
    }
}

#[test]
fn consumption_must_cover_every_reachable_path_and_not_an_overwritten_binding() {
    for source in [
        "test('case', async () => { const p = task.then(v => expect(v).toBe(1)); if (flag) await p; else return p; });",
        "test('case', async () => { const p = task.then(v => expect(v).toBe(1)); if (flag) throw new Error(); await p; });",
        "test('case', async () => { return; task.then(v => expect(v).toBe(1)); });",
    ] {
        assert!(messages(RULE, source).is_empty(), "{source}");
    }
    for source in [
        "test('case', async () => { const p = task.then(v => expect(v).toBe(1)); if (flag) await p; });",
        "test('case', async () => { let p = task.then(v => expect(v).toBe(1)); p = other; await p; });",
        "test('case', async () => { let p; await p; p = task.then(v => expect(v).toBe(1)); });",
    ] {
        assert_eq!(messages(RULE, source).len(), 1, "{source}");
    }
}

#[test]
fn swallowed_rejections_and_finally_returns_do_not_count_as_consumption() {
    for source in [
        "test('case', async () => { const p = task.then(v => expect(v).toBe(1)); try { await p; } catch (error) { throw error; } });",
        "test('case', () => { try { return task.then(v => expect(v).toBe(1)); } catch (error) {} });",
    ] {
        assert!(messages(RULE, source).is_empty(), "{source}");
    }
    for source in [
        "test('case', async () => { const p = task.then(v => expect(v).toBe(1)); try { await p; } catch (error) {} });",
        "test('case', async () => { const p = task.then(v => expect(v).toBe(1)); try { await p; } finally { return; } });",
    ] {
        assert_eq!(messages(RULE, source).len(), 1, "{source}");
    }
}

#[test]
fn only_element_wise_consumption_reaches_promises_stored_in_an_array() {
    for source in [
        "test('case', async () => { const pending = [task.then(v => expect(v).toBe(1))]; await Promise.all(pending); });",
        "test('case', async () => { const pending = [task.then(v => expect(v).toBe(1))]; await pending[0]; });",
        "test('case', () => { const { job: [pending] } = { job: [task.then(v => expect(v).toBe(1))] }; return pending; });",
    ] {
        assert!(messages(RULE, source).is_empty(), "{source}");
    }
    for source in [
        "test('case', async () => { const pending = [task.then(v => expect(v).toBe(1))]; await pending; });",
        "test('case', async () => { const pending = [task.then(v => expect(v).toBe(1))]; await Promise.allSettled(pending); });",
    ] {
        assert_eq!(messages(RULE, source).len(), 1, "{source}");
    }
}

#[test]
fn destructuring_assignments_track_lexical_bindings_instead_of_property_symbols() {
    for source in [
        "test('case', () => { let pending; ({ pending } = { pending: promise.then(value => assert.equal(value, 1)) }); return pending; });",
        "test('case', async () => { let pending; ({ pending = other } = { pending: promise.then(value => assert.equal(value, 1)) }); await pending; });",
        "test('case', () => { let pending; ({ job: { pending } } = { job: { pending: promise.then(value => assert.equal(value, 1)) } }); return pending; });",
        "test('case', () => { let result; ({ pending: result } = { pending: promise.then(value => assert.equal(value, 1)) }); return result; });",
    ] {
        assert!(messages(RULE, source).is_empty(), "{source}");
    }
    for source in [
        "test('case', () => { let pending; ({ pending } = { pending: promise.then(value => assert.equal(value, 1)) }); });",
        "test('case', () => { let pending; ({ pending } = { pending: promise.then(value => assert.equal(value, 1)) }); pending = other; return pending; });",
        "test('case', () => { let pending; ({ pending } = { pending: promise.then(value => assert.equal(value, 1)) }); { const pending = other; return pending; } });",
    ] {
        assert_eq!(messages(RULE, source).len(), 1, "{source}");
    }
}

#[test]
fn recursive_handlers_and_shared_report_nodes_keep_distinct_chains() {
    let source = "function handler(v) { expect(v).toBe(1); return task.then(handler); } test('case', () => task.then(handler));";
    assert!(messages(RULE, source).is_empty());
    let source = "function a(v) { return task.then(b); } function b(v) { expect(v).toBe(1); return task.then(a); } test('case', () => { task.then(a); });";
    assert_eq!(messages(RULE, source).len(), 1);
    let source = "test('case', () => { list.push(first.then(v => expect(v).toBe(1)), second.then(v => expect(v).toBe(2))); });";
    let reports = messages(RULE, source);
    assert_eq!(reports.len(), 2);
    assert_eq!(reports[0]["range"], reports[1]["range"]);
}

#[test]
fn promise_chain_gate_recognizes_cooked_literal_keys_and_enforces_arity() {
    for chain in [
        r#"task["\x74hen"](v => expect(v).toBe(1))"#,
        r#"task[ /* comment */ ("then") ](v => expect(v).toBe(1))"#,
        "task[`finally`](() => expect(value).toBe(1))",
        "task[// comment\u{2028}(\"catch\")](v => expect(v).toBe(1))",
    ] {
        let source = format!("const emoji = '😀'; test('case', () => {{ {chain}; }});");
        let reports = messages(RULE, &source);
        assert_eq!(reports.len(), 1, "{source}");
        assert_eq!(
            reports[0]["range"][0],
            source[..source.find("task[").unwrap()]
                .encode_utf16()
                .count()
        );
    }
    for source in [
        "import { test } from 'rstack/test'; const helper = () => 1; class Holder {} function plain() {} test('case', () => { task.then(helper); });",
        "test('case', () => { task.then(); });",
        "test('case', () => { task.catch(() => expect(value).toBe(1), extra); });",
        "test('case', () => { task.then(() => expect(value).toBe(1), extra, another); });",
        "run('case', () => { task.then(() => expect(value).toBe(1)); });",
    ] {
        assert!(messages(RULE, source).is_empty(), "{source}");
    }
}

#[test]
fn fix_mode_reports_without_modifying_promise_control_flow() {
    let entry = json!({"text":"import { test, expect } from 'rstack/test'; test('case', () => { task.then(v => expect(v).toBe(1)); });"});
    let before = lint(RULE, &entry, false);
    let after = lint(RULE, &entry, true);
    assert_eq!(before["messages"].as_array().unwrap().len(), 1);
    assert_eq!(before["messages"], after["messages"]);
    assert!(after.get("output").is_none());
}
