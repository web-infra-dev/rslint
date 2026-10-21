//! Rust-owned lint execution over the actual tsrs AST and checker.
//!
//! This is the new engine under construction, not a compatibility claim for the
//! complete rule catalog. Unsupported enabled rules are errors. No legacy
//! process, AST export, or fallback engine is used by this entry point.

mod comments;
mod edits;
mod rules;

use std::collections::{BTreeMap, BTreeSet};
use std::sync::Arc;

use serde::{Deserialize, Serialize};
use serde_json::{Value, json};
use tsrs_api::{Handler as _, Response, Session, SessionOptions};
use tsrs_ast::{Node, SourceFile};
use tsrs_checker::Checker;
use tsrs_core::{
    P,
    context::{CheckerLifetime, Context, with_checker_lifetime},
    tspath,
};
use tsrs_project::ID;
use tsrs_vfs::{FS, bundled, osvfs};

pub const COMPILER_REVISION: &str = "280b426b89cd4ecd6153cf48228826bd60a99c61";

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
struct Request {
    schema_version: u32,
    cwd: String,
    files: Vec<InputFile>,
    rules: Vec<RuleConfig>,
    project: Option<String>,
    compiler_options: Option<Value>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct InputFile {
    path: String,
    text: Option<String>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct RuleConfig {
    name: String,
    #[serde(default = "error_severity")]
    severity: u8,
    #[serde(default)]
    options: Vec<Value>,
}

fn error_severity() -> u8 {
    2
}
fn default_compiler_options() -> Value {
    json!({"strict": true, "allowJs": true, "target": 9, "module": 99, "moduleResolution": 100})
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct LintResult {
    schema_version: u32,
    backend: &'static str,
    compiler: Value,
    files: Vec<FileResult>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct FileResult {
    file_path: String,
    messages: Vec<Diagnostic>,
    applied_rules: Vec<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct Diagnostic {
    rule_id: String,
    message_id: &'static str,
    message: String,
    severity: u8,
    line: usize,
    column: usize,
    end_line: usize,
    end_column: usize,
    range: [usize; 2],
    #[serde(skip_serializing_if = "Option::is_none")]
    fix: Option<edits::Fix>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    suggestions: Vec<ReportedSuggestion>,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct ReportedSuggestion {
    message_id: &'static str,
    desc: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<BTreeMap<String, String>>,
    fix: edits::Fix,
}

pub(crate) struct Suggestion {
    message_id: &'static str,
    message: String,
    data: Option<BTreeMap<String, String>>,
    edits: Vec<edits::Edit>,
}

pub(crate) struct Violation {
    pub node: P<Node>,
    pub message_id: &'static str,
    pub message: String,
    pub byte_range: Option<[usize; 2]>,
    pub edits: Vec<edits::Edit>,
    pub suggestions: Vec<Suggestion>,
}

impl Violation {
    pub fn new(node: P<Node>, message_id: &'static str, message: impl Into<String>) -> Self {
        Self {
            node,
            message_id,
            message: message.into(),
            byte_range: None,
            edits: Vec::new(),
            suggestions: Vec::new(),
        }
    }

    pub fn with_byte_range(mut self, range: [usize; 2]) -> Self {
        self.byte_range = Some(range);
        self
    }

    pub fn with_fix(mut self, edits: Vec<edits::Edit>) -> Self {
        self.edits = edits;
        self
    }

    pub fn with_suggestion(
        mut self,
        message_id: &'static str,
        message: impl Into<String>,
        edits: Vec<edits::Edit>,
    ) -> Self {
        self.suggestions.push(Suggestion {
            message_id,
            message: message.into(),
            data: None,
            edits,
        });
        self
    }
}

struct SessionOwner(Arc<Session>);
impl Drop for SessionOwner {
    fn drop(&mut self) {
        self.0.close();
    }
}

/// The only public boundary uses owned JSON. Compiler pointers never leave the
/// worker, and all checker leases end before its snapshot/session is released.
pub fn lint_json(request: &str) -> Result<String, String> {
    let request: Request =
        serde_json::from_str(request).map_err(|e| format!("invalid lint request: {e}"))?;
    validate(&request)?;
    std::thread::Builder::new()
        .name("rslint-rust-core".into())
        .stack_size(512 << 20)
        .spawn(move || lint(request))
        .map_err(|e| format!("cannot start compiler worker: {e}"))?
        .join()
        .map_err(|panic| {
            let text = panic
                .downcast_ref::<String>()
                .map(String::as_str)
                .or_else(|| panic.downcast_ref::<&str>().copied())
                .unwrap_or("unknown compiler panic");
            format!("Rust lint worker failed: {text}")
        })?
        .and_then(|result| serde_json::to_string(&result).map_err(|e| e.to_string()))
}

fn validate(request: &Request) -> Result<(), String> {
    if request.schema_version != 1 {
        return Err("unsupported lint request schemaVersion".into());
    }
    if !std::path::Path::new(&request.cwd).is_absolute() {
        return Err("cwd must be absolute".into());
    }
    if request.files.is_empty() {
        return Err("files must include every admitted input".into());
    }
    if request
        .compiler_options
        .as_ref()
        .is_some_and(|options| !options.is_object())
    {
        return Err("compilerOptions must be an object".into());
    }
    if request.project.is_some() && request.compiler_options.is_some() {
        return Err("project and compilerOptions cannot be combined".into());
    }
    let mut names = BTreeSet::new();
    for rule in &request.rules {
        if !names.insert(&rule.name) {
            return Err(format!("duplicate rule: {}", rule.name));
        }
        if !matches!(rule.severity, 1 | 2) {
            return Err(format!("invalid severity for {}", rule.name));
        }
        if !rules::supported(&rule.name) {
            return Err(format!("rule has not been migrated to Rust: {}", rule.name));
        }
        if !rule.options.is_empty() {
            return Err(format!("{} does not accept options", rule.name));
        }
    }
    Ok(())
}

fn path(cwd: &str, name: &str) -> String {
    tspath::normalize_path(&tspath::combine_paths(cwd, &[name]))
}

fn lint(request: Request) -> Result<LintResult, String> {
    let mut inputs = Vec::new();
    let mut seen = BTreeSet::new();
    let mut overrides = BTreeMap::new();
    for file in request.files {
        let name = path(&request.cwd, &file.path);
        if !seen.insert(name.clone()) {
            return Err(format!("duplicate admitted file: {name}"));
        }
        if let Some(text) = file.text {
            overrides.insert(name.clone(), text);
        }
        inputs.push(name);
    }
    let fs: Arc<dyn FS> = Arc::new(bundled::wrap_fs(osvfs::fs()));
    let session = SessionOwner(Session::new(SessionOptions::new(
        request.cwd.clone(),
        bundled::default_lib_path(),
        fs,
        false,
    )));
    let mut options = json!({"fileSystem": {"kind": "layer", "files": overrides}});
    if let Some(project) = &request.project {
        options["openProjects"] = json!([path(&request.cwd, project)]);
    } else {
        options["createPrograms"] = json!([{"rootFiles": inputs, "compilerOptions": request.compiler_options.unwrap_or_else(default_compiler_options)}]);
    }
    let response = session
        .0
        .handle_request(
            "createSnapshot",
            &serde_json::to_vec(&options).map_err(|e| e.to_string())?,
        )
        .map_err(|e| format!("create compiler snapshot: {e}"))?;
    let response: Value = match response {
        Response::Json(text) => serde_json::from_str(&text).map_err(|e| e.to_string())?,
        Response::Binary(_) => return Err("unexpected binary snapshot metadata".into()),
    };
    let handle = response["snapshot"]
        .as_u64()
        .ok_or("missing compiler snapshot")?;
    let ids: Vec<String> = if request.project.is_some() {
        response["projects"]
            .as_array()
            .ok_or("missing configured projects")?
            .iter()
            .filter_map(|p| p["id"].as_str().map(str::to_owned))
            .collect()
    } else {
        response["operation"]["createdPrograms"]
            .as_array()
            .ok_or("missing created programs")?
            .iter()
            .filter_map(|p| p.as_str().map(str::to_owned))
            .collect()
    };
    if ids.is_empty() {
        return Err("compiler did not create a project".into());
    }
    let snapshot = session.0.snapshot_data(handle).map_err(|e| e.to_string())?;
    let context = with_checker_lifetime(&Context::background(), CheckerLifetime::API);
    let mut files = Vec::new();
    for name in inputs {
        let mut owner = None;
        for id in &ids {
            let program = snapshot
                .get_program(&ID(id.clone()))
                .map_err(|e| e.to_string())?;
            if let Some(file) = program.get_source_file(&name) {
                owner = Some((program, file));
                break;
            }
        }
        let (program, file) = owner
            .ok_or_else(|| format!("admitted file is missing from compiler project: {name}"))?;
        if !program
            .get_syntactic_diagnostics(&context, Some(file))
            .is_empty()
        {
            return Err(format!(
                "syntax diagnostics are not yet supported by the Rust entry point: {name}"
            ));
        }
        comments::ensure_supported(file)?;
        let mut checker = program.get_type_checker_for_file_exclusive(&context, file);
        let mut messages = Vec::new();
        visit(
            file.as_node(),
            file,
            &mut checker,
            &request.rules,
            &mut messages,
        )?;
        drop(checker);
        // Equal starts retain emission order, matching the original engine.
        messages.sort_by_key(|message| message.range[0]);
        files.push(FileResult {
            file_path: name,
            messages,
            applied_rules: request.rules.iter().map(|r| r.name.clone()).collect(),
        });
    }
    drop(snapshot);
    Ok(LintResult {
        schema_version: 1,
        backend: "tsrs-rust",
        compiler: json!({"revision": COMPILER_REVISION}),
        files,
    })
}

fn visit(
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
    rules: &[RuleConfig],
    out: &mut Vec<Diagnostic>,
) -> Result<(), String> {
    for rule in rules {
        for violation in rules::check(&rule.name, node, file, checker) {
            let text = file.text();
            let [start, end] = match violation.byte_range {
                Some(range) => range,
                None => {
                    let pos = violation
                        .node
                        .pos()
                        .source_offset()
                        .ok_or("synthetic diagnostic start")?;
                    let end = violation
                        .node
                        .end()
                        .source_offset()
                        .ok_or("synthetic diagnostic end")?;
                    [tsrs_scanner::skip_trivia(text, pos) as usize, end as usize]
                }
            };
            if start > end || !text.is_char_boundary(start) || !text.is_char_boundary(end) {
                return Err("invalid compiler diagnostic range".into());
            }
            let (offset, line, column) = position(text, start);
            let (end_offset, end_line, end_column) = position(text, end);
            let fix = edits::merge_edits(text, &violation.edits)?;
            let mut suggestions = Vec::new();
            for suggestion in violation.suggestions {
                if let Some(fix) = edits::merge_edits(text, &suggestion.edits)? {
                    suggestions.push(ReportedSuggestion {
                        message_id: suggestion.message_id,
                        desc: suggestion.message,
                        data: suggestion.data,
                        fix,
                    });
                }
            }
            out.push(Diagnostic {
                rule_id: rule.name.clone(),
                message_id: violation.message_id,
                message: violation.message,
                severity: rule.severity,
                line,
                column,
                end_line,
                end_column,
                range: [offset, end_offset],
                fix,
                suggestions,
            });
        }
    }
    let mut error = None;
    node.for_each_child(&mut |child| {
        if let Err(value) = visit(child, file, checker, rules, out) {
            error = Some(value);
            return true;
        }
        false
    });
    error.map_or(Ok(()), Err)
}

fn position(text: &str, end: usize) -> (usize, usize, usize) {
    let mut line = 1;
    let mut column = 1;
    let mut offset = 0;
    let mut previous_cr = false;
    // The public JS diagnostic contract counts offsets in BOM-free source.
    // tsrs retains the initial BOM in its byte-indexed source file.
    let prefix = &text[..end];
    for character in prefix.strip_prefix('\u{feff}').unwrap_or(prefix).chars() {
        offset += character.len_utf16();
        match character {
            '\r' => {
                line += 1;
                column = 1;
            }
            '\n' if previous_cr => {}
            '\n' | '\u{2028}' | '\u{2029}' => {
                line += 1;
                column = 1;
            }
            _ => column += character.len_utf16(),
        }
        previous_cr = character == '\r';
    }
    (offset, line, column)
}

#[cfg(test)]
mod tests;
