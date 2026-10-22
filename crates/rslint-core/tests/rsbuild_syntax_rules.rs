//! Default configuration and diagnostic/edit contracts for the Rsbuild syntax batch.

use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, sources: &[String]) -> Value {
    let request = json!({
        "schemaVersion":1,"cwd":"/rslint-rsbuild-syntax-tests",
        "files":sources.iter().enumerate().map(|(index, text)|
            json!({"path":format!("case-{index}.ts"),"text":text})).collect::<Vec<_>>(),
        "rules":[{"name":rule}]
    });
    let result = lint_json(&request.to_string()).unwrap_or_else(|error| panic!("{rule}: {error}"));
    let result: Value = serde_json::from_str(&result).unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), sources.len());
    for file in result["files"].as_array().unwrap() {
        assert_eq!(file["appliedRules"], json!([rule]));
    }
    result
}

#[test]
fn frozen_baseline_positive_and_negative_rule_inputs() {
    let suites: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-syntax.json"
    ))
    .unwrap();
    for suite in suites.as_array().unwrap() {
        let rule = suite["rule"].as_str().unwrap();
        let valid = suite["valid"].as_array().unwrap();
        let invalid = suite["invalid"].as_array().unwrap();
        assert!(!valid.is_empty() && !invalid.is_empty());
        let sources: Vec<_> = valid
            .iter()
            .chain(invalid)
            .map(|value| value.as_str().unwrap().to_owned())
            .collect();
        let result = lint(rule, &sources);
        for (index, file) in result["files"].as_array().unwrap().iter().enumerate() {
            let diagnostics = file["messages"].as_array().unwrap();
            assert_eq!(
                diagnostics.is_empty(),
                index < valid.len(),
                "{rule}: {}",
                sources[index]
            );
        }
    }
}

fn unmark(marked: &str) -> (String, Vec<[usize; 2]>) {
    let mut text = String::new();
    let mut begin = None;
    let mut ranges = Vec::new();
    for character in marked.chars() {
        match character {
            '«' => assert!(begin.replace(text.len()).is_none()),
            '»' => ranges.push([begin.take().unwrap(), text.len()]),
            _ => text.push(character),
        }
    }
    assert!(begin.is_none());
    (text, ranges)
}

fn utf16(text: &str, byte: usize) -> usize {
    text[..byte].encode_utf16().count()
}

fn marked_case(rule: &str, marked: &str, id: &str, message: &str) -> (String, Vec<Value>) {
    let (source, ranges) = unmark(marked);
    let result = lint(rule, std::slice::from_ref(&source));
    let messages = result["files"][0]["messages"].as_array().unwrap();
    assert_eq!(messages.len(), ranges.len(), "{source}");
    for (diagnostic, [start, end]) in messages.iter().zip(ranges) {
        assert_eq!(diagnostic["ruleId"], rule);
        assert_eq!(diagnostic["messageId"], id);
        assert_eq!(diagnostic["message"], message);
        assert_eq!(diagnostic["severity"], 2);
        assert_eq!(
            diagnostic["range"],
            json!([utf16(&source, start), utf16(&source, end)])
        );
        for (prefix, offset) in [("", start), ("end", end)] {
            let line = source[..offset].split('\n').count();
            let column = source[..offset]
                .rsplit('\n')
                .next()
                .unwrap()
                .encode_utf16()
                .count()
                + 1;
            let line_key = if prefix.is_empty() { "line" } else { "endLine" };
            let column_key = if prefix.is_empty() {
                "column"
            } else {
                "endColumn"
            };
            assert_eq!(diagnostic[line_key], line);
            assert_eq!(diagnostic[column_key], column);
        }
        assert!(
            diagnostic.get("fix").is_none(),
            "suggestions are not automatic fixes"
        );
    }
    (source, messages.to_vec())
}

#[test]
fn case_declarations_suggest_the_entire_clause_for_each_lexical_statement() {
    let (source, messages) = marked_case(
        "no-case-declarations",
        "/* 😀 */ switch (a) { case 1: /* before */ {} «let x=1;» /* mid */ «class C {}» break; /* after */ }",
        "unexpected",
        "Unexpected lexical declaration in case block.",
    );
    let start = source.find("{} let").unwrap();
    let end = source.find("break;").unwrap() + "break;".len();
    let expected = json!([{
        "messageId":"addBrackets","desc":"Add {} brackets around the case block.",
        "fix":{"range":[utf16(&source,start),utf16(&source,end)],"text":format!("{{ {} }}",&source[start..end])}
    }]);
    for message in messages {
        assert_eq!(message["suggestions"], expected);
    }
}

#[test]
fn static_block_suggestion_replaces_only_the_brace_interior() {
    let (source, messages) = marked_case(
        "no-empty-static-block",
        "/* 😀 */ class C { static /* outside */ «{ \n  }» /* after */ }",
        "unexpected",
        "Unexpected empty static block.",
    );
    let start = source.find("{ \n").unwrap() + 1;
    let end = source.find("  }").unwrap() + 2;
    assert_eq!(
        messages[0]["suggestions"],
        json!([{
            "messageId":"suggestComment","desc":"Add comment inside empty static block.",
            "fix":{"range":[utf16(&source,start),utf16(&source,end)],"text":" /* empty */ "}
        }])
    );
}

#[test]
fn empty_patterns_keep_baseline_binding_node_spans() {
    marked_case(
        "no-empty-pattern",
        "function f(«{}», {x: «{}»}) {}",
        "unexpected",
        "Unexpected empty object pattern.",
    );
    marked_case(
        "no-empty-pattern",
        "const {x: «[]»} = source;",
        "unexpected",
        "Unexpected empty array pattern.",
    );
}

#[test]
fn duplicate_cases_report_the_whole_later_clause() {
    marked_case(
        "no-duplicate-case",
        "switch (a) { case a + /* comment */ b: break; «case a+b:\n /* body */ work(); break;» default: break; }",
        "unexpected",
        "Duplicate case label.",
    );
}

#[test]
fn generator_heads_exclude_decorators_and_export_but_include_property_owners() {
    for source in [
        "export default «async function* foo»() {return 0;}",
        "class C { @dec(() => '(') «public static async *[name('(')]<T extends () => void> /* parameters */ »(x:T) {return x;} }",
        "class C { @dec «public field = function* foo»() {return 0;}; }",
        "const o = { «[(() => '(')()]: function*»() {return 0;} };",
        "/* 😀 */ «function*\n  foo»() {return 0;}",
        "«function* outer»() { function* inner() {yield 1;} }",
        "function* outer() { «function* inner»() {return 1;} yield 2; }",
    ] {
        marked_case(
            "require-yield",
            source,
            "missingYield",
            "This generator function does not have 'yield'.",
        );
    }
}
