//! Frozen rslint default-option contracts for the Rsbuild core-check batch.
//! The JSON cases also run through the native-addon/baseline oracle harness.
use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, text: &str, fix: bool) -> Value {
    let request = json!({
        "schemaVersion":1, "cwd":"/rslint-rsbuild-core-checks",
        "files":[{"path":"case.ts","text":text}],
        "rules":[{"name":rule}], "fix":fix
    });
    let result: Value = serde_json::from_str(
        &lint_json(&request.to_string()).unwrap_or_else(|error| panic!("{rule}: {error}\n{text}")),
    )
    .unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    assert_eq!(result["files"][0]["appliedRules"], json!([rule]));
    assert!(
        result["files"][0].get("hasSyntaxErrors").is_none(),
        "{text}: {result}"
    );
    assert!(
        result["files"][0].get("output").is_none(),
        "suggestions are not autofixes"
    );
    result["files"][0]["messages"].clone()
}

#[test]
fn frozen_default_cases_hit_and_do_not_hit_each_rule() {
    let suites: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-core-checks.json"
    ))
    .unwrap();
    for suite in suites.as_array().unwrap() {
        let rule = suite["rule"].as_str().unwrap();
        for (group, expected_empty) in [("valid", true), ("invalid", false)] {
            let cases = suite[group].as_array().unwrap();
            assert!(!cases.is_empty());
            for case in cases {
                let text = case.as_str().unwrap();
                let messages = lint(rule, text, false);
                assert_eq!(
                    messages.as_array().unwrap().is_empty(),
                    expected_empty,
                    "{rule}/{group}: {text}\n{messages}"
                );
                for message in messages.as_array().unwrap() {
                    assert_eq!(message["ruleId"], rule);
                    assert!(message.get("fix").is_none());
                    assert_eq!(message["severity"], 2);
                }
            }
        }
    }
}

fn utf16(text: &str, byte: usize) -> usize {
    text[..byte].encode_utf16().count()
}
fn assert_location(message: &Value, text: &str, range: [usize; 2]) {
    assert_eq!(
        message["range"],
        json!([utf16(text, range[0]), utf16(text, range[1])])
    );
    for (offset, line_key, column_key) in [
        (range[0], "line", "column"),
        (range[1], "endLine", "endColumn"),
    ] {
        let prefix = &text[..offset];
        assert_eq!(message[line_key], prefix.split('\n').count());
        assert_eq!(
            message[column_key],
            prefix.rsplit('\n').next().unwrap().encode_utf16().count() + 1
        );
    }
}

#[test]
fn async_executor_reports_only_async_keyword_and_preserves_parameter_scope() {
    let text = "/* 😀 */ function f(x = new Promise(/* lead */ async () => {})) { new Promise(async () => {}); var Promise; }";
    let messages = lint("no-async-promise-executor", text, true);
    assert_eq!(messages.as_array().unwrap().len(), 1);
    assert_eq!(messages[0]["messageId"], "async");
    assert_eq!(
        messages[0]["message"],
        "Promise executor functions should not be async."
    );
    let begin = text.find("async").unwrap();
    assert_location(&messages[0], text, [begin, begin + 5]);
    assert!(messages[0].get("suggestions").is_none());
}

#[test]
fn empty_statement_comments_are_suggestions_over_brace_interiors() {
    for (kind, text) in [
        ("block", "/* 😀 */ if (x) /* outside */ { \n\t}"),
        ("switch", "/* 😀 */ switch(x) /* outside */ { \n\t}"),
    ] {
        let messages = lint("no-empty", text, true);
        assert_eq!(messages.as_array().unwrap().len(), 1);
        let message = &messages[0];
        let begin = text.find('{').unwrap();
        let end = text.rfind('}').unwrap() + 1;
        assert_location(message, text, [begin, end]);
        assert_eq!(message["messageId"], "unexpected");
        assert_eq!(message["message"], format!("Empty {kind} statement."));
        assert_eq!(
            message["suggestions"],
            json!([{
                "messageId":"suggestComment", "desc":format!("Add comment inside empty {kind} statement."), "data":{"type":kind},
                "fix":{"range":[utf16(text,begin+1),utf16(text,end-1)],"text":" /* empty */ "}
            }])
        );
    }
}

#[test]
fn unsafe_negation_preserves_comments_and_both_distinct_suggestions() {
    for operator in ["in", "instanceof"] {
        let text = format!("/* 😀 */ ! /* operand */ obj.key {operator} /* rhs */ source;");
        let messages = lint("no-unsafe-negation", &text, true);
        assert_eq!(messages.as_array().unwrap().len(), 1);
        let begin = text.find('!').unwrap();
        let left_end = text.find("obj.key").unwrap() + 7;
        let expression_end = text.len() - 1;
        assert_location(&messages[0], &text, [begin, left_end]);
        assert_eq!(messages[0]["messageId"], "unexpected");
        assert_eq!(
            messages[0]["message"],
            format!("Unexpected negating the left operand of '{operator}' operator.")
        );
        assert_eq!(
            messages[0]["suggestions"],
            json!([
                {"messageId":"suggestNegatedExpression", "desc":format!("Negate '{operator}' expression instead of its left operand. This changes the current behavior."),"fix":{"range":[utf16(&text,begin+1),utf16(&text,expression_end)],"text":format!("({})",&text[begin+1..expression_end])}},
                {"messageId":"suggestParenthesisedNegation","desc":"Wrap negation in '()' to make the intention explicit. This preserves the current behavior.","fix":{"range":[utf16(&text,begin),utf16(&text,left_end)],"text":format!("({})",&text[begin..left_end])}}
            ])
        );
    }
}

#[test]
fn typeof_undefined_suggestion_and_invalid_literal_span_exclude_parentheses() {
    let text = "/* 😀 */ (typeof source) === ((undefined));";
    let messages = lint("valid-typeof", text, true);
    assert_eq!(messages.as_array().unwrap().len(), 1);
    let begin = text.find("undefined").unwrap();
    assert_location(&messages[0], text, [begin, begin + 9]);
    assert_eq!(messages[0]["messageId"], "invalidValue");
    assert_eq!(messages[0]["message"], "Invalid typeof comparison value.");
    assert_eq!(
        messages[0]["suggestions"],
        json!([{"messageId":"suggestString","desc":"Use \"undefined\" instead of undefined.","fix":{"range":[utf16(text,begin),utf16(text,begin+9)],"text":"\"undefined\""}}])
    );
    let text = "typeof source === ((`invalid`))";
    let messages = lint("valid-typeof", text, false);
    assert_eq!(messages.as_array().unwrap().len(), 1);
    let begin = text.find('`').unwrap();
    assert_location(&messages[0], text, [begin, begin + 9]);
    assert!(messages[0].get("suggestions").is_none());
}

#[test]
fn loop_reports_header_through_closing_parenthesis() {
    let text = "/* 😀 */ for(let i=0; i<10; (other++, i--) /* closing */ ) { work(); }";
    let messages = lint("for-direction", text, true);
    assert_eq!(messages.as_array().unwrap().len(), 1);
    let begin = text.find("for").unwrap();
    let end = text.find(") { work").unwrap() + 1;
    assert_location(&messages[0], text, [begin, end]);
    assert_eq!(messages[0]["messageId"], "incorrectDirection");
    assert_eq!(
        messages[0]["message"],
        "The update clause in this loop moves the variable in the wrong direction."
    );
    assert!(messages[0].get("suggestions").is_none());
}

#[test]
fn out_of_scope_nondefault_options_are_rejected_not_ignored() {
    for (rule, options) in [
        ("no-empty", json!([{"allowEmptyCatch":true}])),
        (
            "no-unsafe-negation",
            json!([{"enforceForOrderingRelations":true}]),
        ),
        ("valid-typeof", json!([{"requireStringLiterals":true}])),
    ] {
        let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-core-checks","files":[{"path":"case.ts","text":""}],"rules":[{"name":rule,"options":options}]});
        assert!(
            lint_json(&request.to_string())
                .unwrap_err()
                .contains("does not accept options")
        );
    }
}
