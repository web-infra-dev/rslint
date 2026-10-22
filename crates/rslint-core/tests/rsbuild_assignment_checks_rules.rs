//! Frozen assignment-rule contracts, including authored JavaScript/JSDoc scopes.
use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, text: &str, extension: &str, fix: bool) -> Vec<Value> {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-assignment-checks","files":[{"path":format!("case{extension}"),"text":text}],"rules":[{"name":rule}],"fix":fix});
    let result: Value = serde_json::from_str(
        &lint_json(&request.to_string()).unwrap_or_else(|error| panic!("{rule}: {error}\n{text}")),
    )
    .unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let file = &result["files"][0];
    assert_eq!(file["appliedRules"], json!([rule]));
    assert!(
        file.get("output").is_none(),
        "These rules have no autofixes"
    );
    let messages = file["messages"].as_array().unwrap();
    for message in messages {
        assert_eq!(message["ruleId"], rule);
        assert_eq!(message["severity"], 2);
        assert!(message.get("fix").is_none());
        assert!(message.get("suggestions").is_none());
    }
    messages.clone()
}
fn corpus(rule: &str) {
    let cases: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-assignment-checks.json"
    ))
    .unwrap();
    let suite = cases
        .as_array()
        .unwrap()
        .iter()
        .find(|suite| suite["rule"] == rule)
        .unwrap();
    for category in ["valid", "invalid"] {
        assert!(!suite[category].as_array().unwrap().is_empty());
        for case in suite[category].as_array().unwrap() {
            let text = case["text"].as_str().unwrap();
            let messages = lint(
                rule,
                text,
                case["extension"].as_str().unwrap_or(".ts"),
                false,
            );
            let mut expected = case["expected"].as_array().cloned().unwrap_or_default();
            // The frozen RuleTester records listener emission order. The public
            // API calls StableSortDiagnosticsByFileAndStart (server/lint.go:515),
            // so retain every assertion and apply that API's stable source order.
            expected.sort_by_key(|message| {
                (
                    message["line"].as_u64().unwrap_or(1),
                    message["column"].as_u64().unwrap_or(1),
                )
            });
            assert_eq!(
                messages.len(),
                expected.len(),
                "{rule}/{category}: {}\n{text}\n{messages:?}",
                case["origin"]
            );
            for (actual, expected) in messages.iter().zip(expected) {
                for (key, value) in expected.as_object().unwrap() {
                    assert_eq!(
                        &actual[key], value,
                        "{rule}: {}\n{text}\n{actual}",
                        case["origin"]
                    );
                }
            }
        }
    }
}
#[test]
fn frozen_no_class_assign() {
    corpus("no-class-assign");
}
#[test]
fn frozen_no_const_assign() {
    corpus("no-const-assign");
}
#[test]
fn frozen_no_ex_assign() {
    corpus("no-ex-assign");
}
#[test]
fn frozen_no_func_assign() {
    corpus("no-func-assign");
}
#[test]
fn frozen_no_import_assign() {
    corpus("no-import-assign");
}
#[test]
fn frozen_no_self_assign() {
    corpus("no-self-assign");
}

fn marked(rule: &str, source: &str, id: &str, description: &str) {
    let mut text = String::new();
    let mut start = None;
    let mut ranges = Vec::new();
    for character in source.chars() {
        match character {
            '«' => assert!(start.replace(text.len()).is_none()),
            '»' => ranges.push([start.take().unwrap(), text.len()]),
            _ => text.push(character),
        }
    }
    assert!(start.is_none());
    for fix in [false, true] {
        let messages = lint(rule, &text, ".ts", fix);
        assert_eq!(messages.len(), ranges.len(), "{text}: {messages:?}");
        for (message, &[begin, end]) in messages.iter().zip(&ranges) {
            assert_eq!(message["messageId"], id);
            assert_eq!(message["message"], description);
            assert_eq!(
                message["range"],
                json!([
                    text[..begin].encode_utf16().count(),
                    text[..end].encode_utf16().count()
                ])
            );
            for (offset, line, column) in [(begin, "line", "column"), (end, "endLine", "endColumn")]
            {
                let prefix = &text[..offset];
                assert_eq!(message[line], prefix.split('\n').count());
                assert_eq!(
                    message[column],
                    prefix.rsplit('\n').next().unwrap().encode_utf16().count() + 1
                );
            }
        }
    }
}
#[test]
fn declaration_assignments_report_only_the_written_identifier() {
    marked(
        "no-class-assign",
        "const emoji='😀'; class A {}\n«A» = 1; function f(A) { A = 2; }",
        "classReassignment",
        "'A' is a class.",
    );
    marked(
        "no-const-assign",
        "const {x: value} = source;\n({x: «value» = 2} = source); function f(value) { value++; }",
        "const",
        "'value' is constant.",
    );
    marked(
        "no-func-assign",
        "function value() {}\nfor («value» of values) {}",
        "isAFunction",
        "'value' is a function.",
    );
    marked(
        "no-ex-assign",
        "try {} catch ({a,b}) { «b»++; ({a: «a»} = value); function f(a) { a++; } }",
        "unexpected",
        "Do not assign to the exception parameter.",
    );
}
#[test]
fn imports_distinguish_bindings_members_and_deeper_writes() {
    marked(
        "no-import-assign",
        "import {value as local} from './mod';\n«local» += 1; local.deep = 2;",
        "readonly",
        "'local' is read-only.",
    );
    marked(
        "no-import-assign",
        "import * as ns from './mod';\n«ns».member++; Object.assign(«ns», value); ns.member.deep = 1;",
        "readonlyMember",
        "The members of 'ns' are read-only.",
    );
}
#[test]
fn self_assignment_reports_right_side_and_ecmascript_whitespace() {
    marked(
        "no-self-assign",
        "obj.a = «obj\u{00a0}.\u{feff}a»",
        "selfAssignment",
        "'obj.a' is assigned to itself.",
    );
    marked(
        "no-self-assign",
        "[x, ...xs] = [«x», ...ys]",
        "selfAssignment",
        "'x' is assigned to itself.",
    );
    marked(
        "no-self-assign",
        "({p: value} = {p: «value»})",
        "selfAssignment",
        "'value' is assigned to itself.",
    );
    assert!(lint("no-self-assign", "this = this", ".ts", false).is_empty());
}
#[test]
fn no_import_assign_preserves_parameter_initializer_and_jsdoc_scopes() {
    let valid = "import { value } from 'mod'; function outer(){ let value; function inner(parameter=(value=1)){ function value(){} } }";
    assert!(lint("no-import-assign", valid, ".js", false).is_empty());
    let invalid =
        "import { value } from 'mod'; function f(parameter=(value=1)){ function value(){} }";
    let messages = lint("no-import-assign", invalid, ".js", false);
    assert_eq!(messages.len(), 1);
    assert_eq!(messages[0]["message"], "'value' is read-only.");
    assert!(lint("no-import-assign", "import {value} from 'mod'; function f(value){ /** @typedef {number} value */ value=1; }", ".js", false).is_empty());
}
