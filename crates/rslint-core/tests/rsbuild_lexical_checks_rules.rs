//! Public syntax gating and strict diagnostics for the frozen rsbuild lexical
//! rules. Full original listener cases/ranges also live in the module tests.
use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, source: &str, extension: &str, fix: bool) -> Value {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-lexical-tests",
        "files":[{"path":format!("input{extension}"),"text":source}],
        "rules":[{"name":rule}],"fix":fix});
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    result["files"][0].clone()
}

#[test]
fn whitespace_keeps_utf16_ranges_and_only_exempts_strings_by_default() {
    let rule = "no-irregular-whitespace";
    for source in [
        "const text = '\u{00a0}';",
        "const text = '\\u00a0';",
        "\u{feff}const x = 1;",
        "const node = <div title='\u{00a0}'/>;",
    ] {
        let result = lint(rule, source, ".tsx", false);
        assert_eq!(result["appliedRules"], json!([rule]));
        assert_eq!(result["messages"], json!([]));
    }
    for source in [
        "// 😀\u{00a0}",
        "/* 😀\u{00a0} */",
        "const text = `😀\u{00a0}`;",
        "const regex = /😀\u{00a0}/;",
        "const node = <div>😀\u{00a0}</div>;",
    ] {
        let start = source[..source.find('\u{00a0}').unwrap()]
            .encode_utf16()
            .count();
        let result = lint(rule, source, ".tsx", true);
        assert_eq!(result["appliedRules"], json!([rule]));
        assert_eq!(
            result["messages"],
            json!([{
                "ruleId":rule,"severity":2,"messageId":"noIrregularWhitespace","message":"Irregular whitespace not allowed.",
                "line":1,"column":start+1,"endLine":1,"endColumn":start+2,"range":[start,start+1]
            }])
        );
        assert!(result.get("output").is_none());
    }
}

#[test]
fn multiline_reports_only_the_linking_token_and_preserves_optional_roots() {
    let rule = "no-unexpected-multiline";
    for (source, id, message) in [
        (
            "foo\n(bar)",
            "function",
            "Unexpected newline between function and ( of function call.",
        ),
        (
            "foo\n[bar]",
            "property",
            "Unexpected newline between object and [ of property access.",
        ),
        (
            "foo\n`bar`",
            "taggedTemplate",
            "Unexpected newline between template tag and template literal.",
        ),
        (
            "foo\n/bar/gi",
            "division",
            "Unexpected newline between numerator and division operator.",
        ),
    ] {
        let result = lint(rule, source, ".ts", true);
        assert_eq!(result["appliedRules"], json!([rule]));
        assert_eq!(
            result["messages"],
            json!([{"ruleId":rule,"severity":2,"messageId":id,"message":message,"line":2,"column":1,"endLine":2,"endColumn":2,"range":[4,5]}])
        );
        assert!(result.get("output").is_none());
    }
    for source in [
        "foo\n()",
        "foo\n?.[bar]",
        "foo\n?.(bar)",
        "foo\n/bar/ gi",
        "foo /bar/gi",
        "tag<T>`bar`",
        "foo\n/bar/g\\u0061",
    ] {
        assert_eq!(
            lint(rule, source, ".ts", false)["messages"],
            json!([]),
            "{source}"
        );
    }
}

#[test]
fn numeric_original_token_and_bigint_are_distinguished_without_fixes() {
    let rule = "no-loss-of-precision";
    let source = "/* 😀 */ const x = 9_007_199_254_740_993;";
    let start = source[..source.find("9_").unwrap()].encode_utf16().count();
    let length = "9_007_199_254_740_993".len();
    let result = lint(rule, source, ".ts", true);
    assert_eq!(
        result["messages"],
        json!([{"ruleId":rule,"severity":2,"messageId":"noLossOfPrecision","message":"This number literal will lose precision at runtime.","line":1,"column":start+1,"endLine":1,"endColumn":start+length+1,"range":[start,start+length]}])
    );
    assert!(result.get("output").is_none());
    for source in [
        "const x = 9007199254740993n;",
        "const x = '9007199254740993';",
        "const x = 9007199254740992;",
        "const x = 0x40000000000000;",
        "const x = 5e-324;",
        "const x = 1e-9999;",
    ] {
        let result = lint(rule, source, ".ts", false);
        assert_eq!(result["appliedRules"], json!([rule]));
        assert_eq!(result["messages"], json!([]), "{source}");
    }
}

#[test]
fn original_parser_rejected_numeric_cases_remain_syntax_diagnostics() {
    let rule = "no-loss-of-precision";
    for source in [
        "var x = 019.5",
        "var x = 0195",
        "var x = 00195",
        "var x = 0008",
        "var x = 0377777777777777777",
        "var x = 0400000000000000001",
    ] {
        let result = lint(rule, source, ".ts", false);
        assert_eq!(result["appliedRules"], json!([]));
        let messages = result["messages"].as_array().unwrap();
        assert!(!messages.is_empty());
        assert!(
            messages
                .iter()
                .all(|m| m["ruleId"].as_str().unwrap().starts_with("TypeScript(TS"))
        );
        assert!(
            messages
                .iter()
                .all(|m| m.get("fix").is_none() && m.get("suggestions").is_none())
        );
    }
}

#[test]
fn corpus_retains_default_cases_original_skips_and_nondefault_options() {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-lexical-checks.json"
    ))
    .unwrap();
    let groups = corpus.as_array().unwrap();
    assert_eq!(groups.len(), 3);
    assert_eq!(
        groups
            .iter()
            .map(|g| g["valid"].as_array().unwrap().len() + g["invalid"].as_array().unwrap().len())
            .sum::<usize>(),
        490
    );
    assert_eq!(
        groups[0]["excludedNonDefaultCases"]
            .as_array()
            .unwrap()
            .len(),
        131
    );
    assert_eq!(
        groups[2]["knownBaselineSkipped"].as_array().unwrap().len(),
        6
    );
    for group in groups {
        assert!(!group["valid"].as_array().unwrap().is_empty());
        assert!(!group["invalid"].as_array().unwrap().is_empty());
        for entry in group["invalid"].as_array().unwrap() {
            assert!(
                !entry["expectedRuleDiagnostics"]
                    .as_array()
                    .unwrap()
                    .is_empty()
            );
        }
    }
}
