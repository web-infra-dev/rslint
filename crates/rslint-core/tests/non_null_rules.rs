use rslint_core::lint_json;
use serde_json::{Value, json};

const EXTRA: &str = "@typescript-eslint/no-extra-non-null-assertion";
const OPTIONAL: &str = "@typescript-eslint/no-non-null-asserted-optional-chain";

fn lint(rule: &str, sources: &[&str]) -> Vec<Value> {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-non-null-tests",
        "files": sources.iter().enumerate().map(|(i,text)| json!({"path":format!("case-{i}.ts"),"text":text})).collect::<Vec<_>>(),
        "rules":[{"name":rule}]});
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    let files = result["files"].as_array().unwrap();
    assert_eq!(files.len(), sources.len());
    files
        .iter()
        .enumerate()
        .map(|(i, file)| {
            assert_eq!(
                file["filePath"],
                format!("/rslint-non-null-tests/case-{i}.ts")
            );
            assert_eq!(file["appliedRules"], json!([rule]));
            file["messages"].clone()
        })
        .collect()
}

// Apply the public payload in its actual JS UTF-16 coordinate space. This is
// not the engine's future iterative autofix implementation.
fn apply(source: &str, edits: &[&Value]) -> String {
    let (bom, source) = source
        .strip_prefix('\u{feff}')
        .map_or(("", source), |s| ("\u{feff}", s));
    let mut units: Vec<u16> = source.encode_utf16().collect();
    let mut edits = edits.to_vec();
    edits.sort_by_key(|edit| edit["range"][0].as_u64().unwrap());
    for edit in edits.into_iter().rev() {
        let start = edit["range"][0].as_u64().unwrap() as usize;
        let end = edit["range"][1].as_u64().unwrap() as usize;
        units.splice(start..end, edit["text"].as_str().unwrap().encode_utf16());
    }
    format!("{bom}{}", String::from_utf16(&units).unwrap())
}

#[test]
fn redundant_assertions_keep_comments_parentheses_and_utf16_fix_ranges() {
    let cases = [
        ("foo!!.bar;", "foo!.bar;"),
        ("bar!?.n;", "bar?.n;"),
        ("bar!?.();", "bar?.();"),
        ("(foo!)!.bar;", "(foo)!.bar;"),
        ("(bar!)?.n;", "(bar)?.n;"),
        ("(bar)!?.n;", "(bar)?.n;"),
        ("(bar!)?.();", "(bar)?.();"),
        (
            "const result = /* leading */ value!!;",
            "const result = /* leading */ value!;",
        ),
        (
            "const result = value /* first */ ! /* second */ !;",
            "const result = value /* first */  /* second */ !;",
        ),
        ("const result = 数据!!;", "const result = 数据!;"),
        ("const result = value!!!;", "const result = value!;"),
        ("bar!?.[0];", "bar?.[0];"),
        ("\u{feff}'😀'; 数据!!;", "\u{feff}'😀'; 数据!;"),
    ];
    let sources = cases.iter().map(|(s, _)| *s).collect::<Vec<_>>();
    for ((source, expected), messages) in cases.iter().zip(lint(EXTRA, &sources)) {
        let messages = messages.as_array().unwrap();
        assert_eq!(
            messages.len(),
            if source.contains("!!!") { 2 } else { 1 },
            "{source}"
        );
        for message in messages {
            assert_eq!(message["messageId"], "noExtraNonNullAssertion");
            assert_eq!(message["message"], "Forbidden extra non-null assertion.");
            assert!(message.get("suggestions").is_none());
        }
        assert_eq!(
            apply(
                source,
                &messages.iter().map(|m| &m["fix"]).collect::<Vec<_>>()
            ),
            *expected
        );
    }
    assert_eq!(
        lint(EXTRA, &["/* leading */ foo!!;"])[0],
        json!([{
            "ruleId":EXTRA,"messageId":"noExtraNonNullAssertion","message":"Forbidden extra non-null assertion.","severity":2,
            "line":1,"column":15,"endLine":1,"endColumn":19,"range":[14,18],"fix":{"range":[17,18],"text":""}
        }])
    );
}

#[test]
fn optional_chain_reports_suggestions_instead_of_automatic_fixes() {
    let cases = [
        ("foo?.bar!;", "foo?.bar;"),
        ("foo?.['bar']!;", "foo?.['bar'];"),
        ("foo?.bar()!;", "foo?.bar();"),
        ("foo.bar?.()!;", "foo.bar?.();"),
        ("(foo?.bar)!.baz", "(foo?.bar).baz"),
        ("(foo?.bar)!().baz", "(foo?.bar)().baz"),
        ("(foo?.bar)!", "(foo?.bar)"),
        ("(foo?.bar)!()", "(foo?.bar)()"),
        ("(foo?.bar!)", "(foo?.bar)"),
        ("(foo?.bar!)()", "(foo?.bar)()"),
        (
            "\u{feff}'😀'; 数据?.值 /* keep */ !;",
            "\u{feff}'😀'; 数据?.值 /* keep */ ;",
        ),
    ];
    let sources = cases.iter().map(|(s, _)| *s).collect::<Vec<_>>();
    for ((source, expected), messages) in cases.iter().zip(lint(OPTIONAL, &sources)) {
        let messages = messages.as_array().unwrap();
        assert_eq!(messages.len(), 1, "{source}");
        let message = &messages[0];
        assert_eq!(message["messageId"], "noNonNullOptionalChain");
        assert_eq!(
            message["message"],
            "Optional chain expressions can return undefined by design - using a non-null assertion is unsafe and wrong."
        );
        assert!(message.get("fix").is_none());
        let suggestions = message["suggestions"].as_array().unwrap();
        assert_eq!(suggestions.len(), 1);
        assert_eq!(suggestions[0]["messageId"], "suggestRemovingNonNull");
        assert_eq!(
            suggestions[0]["desc"],
            "You should remove the non-null assertion."
        );
        assert!(suggestions[0].get("data").is_none());
        assert_eq!(apply(source, &[&suggestions[0]["fix"]]), *expected);
    }
}

#[test]
fn valid_assertions_and_chain_boundaries_stay_silent() {
    for (rule, sources) in [
        (
            EXTRA,
            vec![
                "const foo: { bar: number } | null = null; const bar = foo!.bar;",
                "function foo(bar: number | undefined) { const bar: number = bar!; }",
                "function foo(bar?: { n: number }) { return bar?.n; }",
                "checksCounter?.textContent!.trim();",
                "function foo(key: string | null) { const obj = {}; return obj?.[key!]; }",
            ],
        ),
        (
            OPTIONAL,
            vec![
                "foo.bar!;",
                "foo.bar!.baz;",
                "foo.bar!.baz();",
                "foo.bar()!;",
                "foo.bar()!();",
                "foo.bar()!.baz;",
                "foo?.bar;",
                "foo?.bar();",
                "(foo?.bar).baz!;",
                "(foo?.bar()).baz!;",
                "foo?.bar!.baz;",
                "foo?.bar!();",
                "foo?.['bar']!.baz;",
            ],
        ),
    ] {
        for (source, result) in sources.iter().zip(lint(rule, &sources)) {
            assert_eq!(result, json!([]), "{rule}: {source}");
        }
    }
}
