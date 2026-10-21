//! Cases from the fixed baseline rule suites, retaining complete diagnostic
//! spans. Markers are removed before the source reaches the compiler.

use rslint_core::lint_json;
use serde_json::{Value, json};

fn fixture(rule: &str, marked: &str, id: &str, message: &str) -> (String, Vec<Value>) {
    let mut source = String::new();
    let mut start = None;
    let mut spans = Vec::new();
    for character in marked.chars() {
        match character {
            '«' => assert!(start.replace(source.len()).is_none()),
            '»' => spans.push((
                start.take().expect("unmatched closing marker"),
                source.len(),
            )),
            _ => source.push(character),
        }
    }
    assert!(start.is_none(), "unclosed marker: {marked}");
    let messages = spans
        .into_iter()
        .map(|(start, end)| {
            let before = &source[..start];
            let through = &source[..end];
            json!({
                "ruleId":rule,"messageId":id,"message":message,"severity":2,
                "line":before.split('\n').count(),
                "column":before.rsplit('\n').next().unwrap().encode_utf16().count()+1,
                "endLine":through.split('\n').count(),
                "endColumn":through.rsplit('\n').next().unwrap().encode_utf16().count()+1,
                "range":[before.encode_utf16().count(),through.encode_utf16().count()]
            })
        })
        .collect();
    (source, messages)
}

fn run_rule(rule: &str, valid: &[&str], invalid: &[(&str, &str, &str)]) {
    let cases: Vec<_> = valid
        .iter()
        .map(|source| fixture(rule, source, "", ""))
        .chain(
            invalid
                .iter()
                .map(|(source, id, text)| fixture(rule, source, id, text)),
        )
        .collect();
    assert!(!cases.is_empty());
    let request = json!({
        "schemaVersion":1,"cwd":"/rslint-structural-tests",
        "files":cases.iter().enumerate().map(|(index,(source,_))|
            json!({"path":format!("case-{index}.ts"),"text":source})).collect::<Vec<_>>(),
        "rules":[{"name":rule}]
    });
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    let files = result["files"].as_array().unwrap();
    assert_eq!(files.len(), cases.len(), "retain files without diagnostics");
    for (index, ((source, messages), file)) in cases.iter().zip(files).enumerate() {
        assert_eq!(
            file,
            &json!({"filePath":format!("/rslint-structural-tests/case-{index}.ts"),
                "messages":messages,"appliedRules":[rule]}),
            "{rule}: {source}"
        );
    }
}

#[test]
fn sparse_arrays_keep_all_baseline_elisions_and_assignment_target_exemptions() {
    let invalid = [
        "var a = [«,»];",
        "var a = [ 1,«,» 2];",
        "[\r\n\t/* comment */«,»\n// comment\n «,»];",
        "[(( [a,] )),«,»«,»];",
        "[«,»(( [a,] )),«,»];",
        "[a, b] = [1, «,» 2];",
        "[value = [«,»]] = source;",
        "const emoji = \"😀\"; const values = [1,«,» 2];",
        "const [{value = [«,»]}] = source;",
    ];
    let invalid: Vec<_> = invalid
        .iter()
        .map(|source| {
            (
                *source,
                "unexpectedSparseArray",
                "Unexpected comma in middle of array.",
            )
        })
        .collect();
    run_rule(
        "no-sparse-arrays",
        &[
            "var a = [ 1, 2, ]",
            "[, suggestion] = await all();",
            "[, , endLine, endChar] = o.range;",
            "[, ref, authorName] = match;",
            "[a, , b] = [1, 2, 3];",
            "for ([, x] of y) {}",
            "[[, a]] = b;",
            "({ value: [, a] } = b);",
            "[...[, a]] = b;",
            "const [, value] = source;",
            "const [] = source; const empty = [];",
        ],
        &invalid,
    );
}

#[test]
fn delete_var_preserves_parentheses_and_typescript_wrapper_boundaries() {
    let invalid = [
        "«delete x»",
        "var x; «delete x»;",
        "let y = 1; «delete y»;",
        "«delete (x)»;",
        "«delete ((x))»;",
        "var x; if (true) { «delete x»; }",
        "function f() { «delete x»; }",
        "const f = () => { «delete x»; };",
        "for (;;) { «delete x»; }",
        "try { «delete x»; } catch(e) {}",
        "class C { m() { «delete x»; } }",
        "«delete x»; «delete y»;",
        "«delete arguments»;",
        "«delete NaN»;",
        "if (true) { «delete (x)»; }",
        "switch(0) { case 0: «delete x»; }",
        "while (false) { «delete x»; }",
        "async function f() { «delete x»; }",
        "function* g() { «delete x»; }",
        "«delete\nx»",
        "/* 😀 */ «delete /* operand */ (x)»;",
    ];
    let invalid: Vec<_> = invalid
        .iter()
        .map(|source| (*source, "unexpected", "Variables should not be deleted."))
        .collect();
    run_rule(
        "no-delete-var",
        &[
            "delete x.prop;",
            "delete foo.bar.baz;",
            "delete obj[\"key\"];",
            "delete obj[0];",
            "delete a?.b;",
            "delete obj[`key`];",
            "delete (x.prop);",
            "delete ((obj[\"key\"]));",
            "delete (x as any);",
            "delete x!;",
            "delete (<any>x);",
            "delete (x satisfies any);",
            "delete (0, x);",
        ],
        &invalid,
    );
}

#[test]
fn duplicate_args_preserve_each_binding_occurrence_and_function_kind_scope() {
    run_rule(
        "no-dupe-args",
        &[
            "function foo(a, b, c) {}",
            "var foo = function(a, b, c) {}",
            "function foo(a, b, c, d) {}",
            "function foo(a, {b, c}) {}",
            "function foo({a}, {b}) {}",
            "function foo([a, b], c) {}",
            "function foo({key:a}, {key:b}) {}",
            "function foo(a = external, {other: b = external}) {}",
            "function foo(a) { function bar(a) {} }",
            "const arrow = (a, a) => a;",
            "class C { method(a, a) {} constructor(a, a) {} }",
            "const object = { method(a, a) {} };",
        ],
        &[
            (
                "function foo(a, b, «a») {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
            (
                "function foo(a, «a», «a») {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
            (
                "var foo = function(a, b, «b») {}",
                "unexpected",
                "Duplicate param 'b'.",
            ),
            (
                "function foo(a, {«a»}) {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
            (
                "function foo(a, [«a»]) {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
            (
                "function foo(a, {key: [«a» = 1, ...rest]}) {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
            (
                "function foo(a, ...«a») {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
            (
                "function foo(a: string, «a»?: number) {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
            (
                r"function foo(a, «\u0061») {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
            (
                "/* 😀 */ function foo(α, [«α»]) {}",
                "unexpected",
                "Duplicate param 'α'.",
            ),
            (
                "function foo({key: {a}}, {other: {«a»}}) {}",
                "unexpected",
                "Duplicate param 'a'.",
            ),
        ],
    );
}

#[test]
fn structural_rule_options_remain_optionless() {
    for rule in ["no-delete-var", "no-dupe-args", "no-sparse-arrays"] {
        let request = json!({
            "schemaVersion":1,"cwd":"/rslint-structural-tests",
            "files":[{"path":"case.ts","text":"const value = 1;"}],
            "rules":[{"name":rule,"options":[{}]}]
        });
        let error = lint_json(&request.to_string()).unwrap_err();
        assert!(error.contains("options"), "{rule}: {error}");
    }
}
