//! Frozen default expression-rule contracts; complete diagnostic/fix oracle inputs
//! live beside the other direct Rust+JS migration fixtures.
use rslint_core::lint_json;
use serde_json::{Value, json};
fn lint(rule: &str, text: &str, extension: &str, fix: bool) -> Value {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-expression-checks","files":[{"path":format!("case{extension}"),"text":text}],"rules":[{"name":rule}],"fix":fix});
    let result: Value = serde_json::from_str(
        &lint_json(&request.to_string()).unwrap_or_else(|error| panic!("{rule}: {error}\n{text}")),
    )
    .unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    assert_eq!(result["files"][0]["appliedRules"], json!([rule]));
    for message in result["files"][0]["messages"].as_array().unwrap() {
        assert_eq!(message["ruleId"], rule);
        assert_eq!(message["severity"], 2);
        assert!(message.get("suggestions").is_none());
        if rule != "no-extra-boolean-cast" {
            assert!(message.get("fix").is_none());
        }
    }
    result["files"][0].clone()
}
fn description(id: &str) -> &str {
    match id {
        "unexpectedNegation" => "Redundant double negation.",
        "unexpectedCall" => "Redundant Boolean call.",
        "preferRestParams" => "Use the rest parameters instead of 'arguments'.",
        "preferSpread" => "Use the spread operator instead of '.apply()'.",
        "comparisonWithNaN" => "Use the isNaN function to compare with NaN.",
        "switchNaN" => {
            "'switch(NaN)' can never match a case clause. Use Number.isNaN instead of the switch."
        }
        "caseNaN" => "'case NaN' can never match. Use Number.isNaN before the switch.",
        _ => panic!("unknown message {id}"),
    }
}
fn corpus(rule: &str) {
    let cases: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-expression-checks.json"
    ))
    .unwrap();
    let group = cases
        .as_array()
        .unwrap()
        .iter()
        .find(|group| group["rule"] == rule)
        .unwrap();
    for category in ["valid", "invalid"] {
        assert!(!group[category].as_array().unwrap().is_empty());
        for entry in group[category].as_array().unwrap() {
            let text = entry["text"].as_str().unwrap();
            let extension = entry["extension"].as_str().unwrap_or(".ts");
            let file = lint(rule, text, extension, false);
            let messages = file["messages"].as_array().unwrap();
            let mut expected = entry["expected"].as_array().cloned().unwrap_or_default();
            // Public API sorts by source position; Go RuleTester records listener order.
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
                entry["origin"]
            );
            for (actual, expected) in messages.iter().zip(expected) {
                for (key, value) in expected.as_object().unwrap() {
                    assert_eq!(
                        &actual[key], value,
                        "{rule}: {}\n{text}\n{actual}",
                        entry["origin"]
                    );
                }
                assert_eq!(
                    actual["message"],
                    description(actual["messageId"].as_str().unwrap())
                );
            }
            if let Some(output) = entry.get("outputSteps") {
                let fixed = lint(rule, text, extension, true);
                if output.is_null() {
                    assert!(fixed.get("output").is_none(), "{text}: {fixed}");
                } else {
                    assert_eq!(
                        fixed["output"],
                        *output.as_array().unwrap().last().unwrap(),
                        "{text}: {fixed}"
                    );
                }
                let after = fixed["output"].as_str().unwrap_or(text);
                assert_eq!(
                    fixed["messages"],
                    lint(rule, after, extension, false)["messages"],
                    "post-fix diagnostics must be fresh"
                );
            }
        }
    }
}
#[test]
fn frozen_no_extra_boolean_cast() {
    corpus("no-extra-boolean-cast");
}
#[test]
fn frozen_prefer_rest_params() {
    corpus("prefer-rest-params");
}
#[test]
fn frozen_prefer_spread() {
    corpus("prefer-spread");
}
#[test]
fn frozen_use_isnan() {
    corpus("use-isnan");
}

fn marked(rule: &str, source: &str, id: &str) -> (String, Value) {
    let mut text = String::new();
    let mut begin = None;
    let mut ranges = Vec::new();
    for ch in source.chars() {
        match ch {
            '«' => assert!(begin.replace(text.len()).is_none()),
            '»' => ranges.push([begin.take().unwrap(), text.len()]),
            _ => text.push(ch),
        }
    }
    assert!(begin.is_none());
    let file = lint(rule, &text, ".ts", false);
    let messages = file["messages"].as_array().unwrap();
    assert_eq!(messages.len(), ranges.len(), "{text}: {file}");
    for (message, [begin, end]) in messages.iter().zip(ranges) {
        assert_eq!(message["messageId"], id);
        assert_eq!(message["message"], description(id));
        assert_eq!(
            message["range"],
            json!([
                text[..begin].encode_utf16().count(),
                text[..end].encode_utf16().count()
            ])
        );
        for (offset, line, column) in [(begin, "line", "column"), (end, "endLine", "endColumn")] {
            let prefix = &text[..offset];
            assert_eq!(message[line], prefix.split('\n').count());
            assert_eq!(
                message[column],
                prefix.rsplit('\n').next().unwrap().encode_utf16().count() + 1
            );
        }
    }
    (text, file)
}
#[test]
fn boolean_cast_fixes_keep_precedence_spacing_and_comment_safety() {
    let (text, file) = marked(
        "no-extra-boolean-cast",
        "function f(){return!«Boolean()»}",
        "unexpectedCall",
    );
    let begin = text.find('!').unwrap();
    let end = text.find('}').unwrap();
    assert_eq!(
        file["messages"][0]["fix"],
        json!({"range":[begin,end],"text":" true"})
    );
    assert_eq!(
        lint("no-extra-boolean-cast", &text, ".ts", true)["output"],
        "function f(){return true}"
    );
    for (source, output) in [
        ("!Boolean(a || b)", "!(a || b)"),
        ("!!(a ? b : c) ? yes : no", "(a ? b : c) ? yes : no"),
        (r"if (!!(/\/\//.test(x))) {}", r"if (/\/\//.test(x)) {}"),
        ("if (Boolean('/* string */')) {}", "if ('/* string */') {}"),
        ("if (Boolean(`// template`)) {}", "if (`// template`) {}"),
        (
            "if (Boolean(`${x /* comment */}`)) {}",
            "if (Boolean(`${x /* comment */}`)) {}",
        ),
        ("! /* keep */ Boolean()", "! /* keep */ Boolean()"),
    ] {
        let fixed = lint("no-extra-boolean-cast", source, ".ts", true);
        assert_eq!(
            fixed["output"].as_str().unwrap_or(source),
            output,
            "{source}: {fixed}"
        );
    }
}
#[test]
fn boolean_cast_follows_frozen_syntax_and_default_option_boundaries() {
    for source in [
        "if ((Boolean)(x)) {}",
        "if (!( !x)) {}",
        "if (x || !!y) {}",
        "if (x ?? Boolean(y)) {}",
        "if (new Boolean(x)) {}",
    ] {
        assert!(
            lint("no-extra-boolean-cast", source, ".ts", false)["messages"]
                .as_array()
                .unwrap()
                .is_empty(),
            "{source}"
        );
    }
    let source = "function f(Boolean:any){if(Boolean(x)){}}";
    assert_eq!(
        lint("no-extra-boolean-cast", source, ".ts", false)["messages"]
            .as_array()
            .unwrap()
            .len(),
        1,
        "frozen rule intentionally matches Boolean syntax even when shadowed"
    );
}
#[test]
fn rest_params_reports_references_and_retains_frozen_scope_boundaries() {
    marked(
        "prefer-rest-params",
        "function f(){ const emoji='😀'; return «arguments»[0] + (() => «arguments»[1])(); }",
        "preferRestParams",
    );
    assert!(lint("prefer-rest-params","function f(){return arguments.length;} function g(){let arguments;return arguments[0]}",".ts",false)["messages"].as_array().unwrap().is_empty());
    marked(
        "prefer-rest-params",
        "function f(){ return { «arguments» }; }",
        "preferRestParams",
    );
}
#[test]
fn spread_compares_tokens_instead_of_values_or_trimmed_source() {
    marked(
        "prefer-spread",
        "const emoji='😀'; «obj\n.foo.apply(obj, args)»;",
        "preferSpread",
    );
    for source in [
        "[a,].m.apply([a], args)",
        "(x).y.m.apply(x.y,args)",
        "['a'].m.apply([\"a\"],args)",
        "[1].m.apply([1.0],args)",
        "obj.m.apply(obj,([1,2]))",
    ] {
        assert!(
            lint("prefer-spread", source, ".ts", false)["messages"]
                .as_array()
                .unwrap()
                .is_empty(),
            "{source}"
        );
    }
    for (extension, source, count) in [
        (".js", r"obj.m.apply(\u006fbj,args)", 1),
        (".ts", r"obj.m.apply(\u006fbj,args)", 0),
        (".jsx", "(<p>&amp;</p>).m.apply(<p>&#38;</p>,args)", 1),
        (".tsx", "(<p>&amp;</p>).m.apply(<p>&#38;</p>,args)", 0),
        (".jsx", "(<p>&#x10000;</p>).m.apply(<p>&#0;</p>,args)", 1),
    ] {
        assert_eq!(
            lint("prefer-spread", source, extension, false)["messages"]
                .as_array()
                .unwrap()
                .len(),
            count,
            "{extension}: {source}"
        );
    }
}
#[test]
fn nan_respects_global_values_and_exact_switch_clause_ranges() {
    marked(
        "use-isnan",
        "type Number = {NaN:0}; const emoji='😀'; «x === Number['NaN']»;",
        "comparisonWithNaN",
    );
    marked(
        "use-isnan",
        "switch(x){ «case (0,NaN): break;» }",
        "caseNaN",
    );
    for source in [
        "import {NaN} from 'mod'; x===NaN",
        "function f({Number}){return Number.NaN===x}",
        "const C=class Number{method(){return x===Number.NaN}}",
        "x === (a,(b,NaN))",
        "x === (NaN as number)",
    ] {
        assert!(
            lint("use-isnan", source, ".ts", false)["messages"]
                .as_array()
                .unwrap()
                .is_empty(),
            "{source}"
        );
    }
}
