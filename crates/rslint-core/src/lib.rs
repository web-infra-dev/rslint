//! Rust-owned lint execution over the actual tsrs AST and checker.
//!
//! This is the new engine under construction, not a compatibility claim for the
//! complete rule catalog. Unsupported enabled rules are errors. No legacy
//! process, AST export, or fallback engine is used by this entry point.

mod autofix;
mod comments;
mod configured;
mod directives;
mod edits;
mod programs;
mod projects;
mod rules;
mod source;

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

pub use configured::lint_projects_json;

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
    #[serde(default)]
    fix: bool,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct InputFile {
    path: String,
    text: Option<String>,
}

#[derive(Clone, Deserialize)]
#[serde(deny_unknown_fields)]
pub(crate) struct RuleConfig {
    name: String,
    #[serde(default = "error_severity")]
    severity: u8,
    #[serde(default)]
    options: Vec<Value>,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
struct PlanRequest {
    schema_version: u32,
    cwd: String,
    files: Vec<PlanInput>,
    #[serde(default)]
    sources: Vec<InputFile>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct PlanInput {
    path: String,
    policy: projects::ProjectPolicy,
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
    summary: LintSummary,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct LintSummary {
    fixed_issues: usize,
    /// Actual rule-execution workers, excluding compiler-internal parallelism.
    threads: usize,
}

#[derive(Clone, Serialize)]
#[serde(rename_all = "camelCase")]
struct FileResult {
    file_path: String,
    messages: Vec<Diagnostic>,
    applied_rules: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    has_type_information: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    output: Option<String>,
    #[serde(skip_serializing_if = "std::ops::Not::not")]
    has_syntax_errors: bool,
}

#[derive(Clone, Serialize)]
#[serde(rename_all = "camelCase")]
struct Diagnostic {
    rule_id: String,
    #[serde(skip_serializing_if = "String::is_empty")]
    message_id: String,
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
    // Owned original byte edits are used for atomic automatic fix selection.
    // Public merged UTF-16 payloads must never be used as compiler offsets.
    #[serde(skip)]
    edits: Vec<edits::Edit>,
}

#[derive(Clone, Serialize)]
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

struct SessionOwner {
    session: Arc<Session>,
    fs: Arc<source::SourceFs>,
}

impl SessionOwner {
    fn new(cwd: &str) -> Self {
        let base: Arc<dyn FS> = Arc::new(bundled::wrap_fs(osvfs::fs()));
        let fs = Arc::new(source::SourceFs::new(base));
        let session = Session::new(SessionOptions::new(
            cwd.into(),
            bundled::default_lib_path(),
            fs.clone(),
            false,
        ));
        Self { session, fs }
    }
}
impl Drop for SessionOwner {
    fn drop(&mut self) {
        self.session.close();
        // At this compiler pin the parse-config host retains its base FS Arc.
        // Clear generation-owned sources even on errors and panics.
        self.fs.clear();
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

/// Resolve per-file project policy without guessing ownership from imports.
/// This planning boundary returns owned metadata; it does not execute lint.
pub fn plan_json(request: &str) -> Result<String, String> {
    let request: PlanRequest = serde_json::from_str(request)
        .map_err(|error| format!("invalid project plan request: {error}"))?;
    if request.schema_version != 1 || !std::path::Path::new(&request.cwd).is_absolute() {
        return Err("project plan requires schemaVersion 1 and absolute cwd".into());
    }
    if request.files.is_empty() {
        return Err("project plan files must include every admitted input".into());
    }
    std::thread::Builder::new()
        .name("rslint-rust-projects".into())
        .stack_size(512 << 20)
        .spawn(move || {
            let owner = SessionOwner::new(&request.cwd);
            let mut sources = BTreeMap::new();
            for source in request.sources {
                let name = path(&request.cwd, &source.path);
                let text = source.text.ok_or("planning source overlay requires text")?;
                if sources.insert(name, text).is_some() {
                    return Err("duplicate project plan overlay".into());
                }
            }
            owner.fs.set_generation(&request.cwd, &sources)?;
            let mut seen = BTreeSet::new();
            let mut inputs = Vec::new();
            for input in request.files {
                let name = path(&request.cwd, &input.path);
                if !seen.insert(name.clone()) {
                    return Err(format!("duplicate admitted project input: {name}"));
                }
                inputs.push(projects::InputProject {
                    path: name,
                    policy: input.policy,
                });
            }
            let plan = projects::select(&owner.session, &inputs)?;
            serde_json::to_string(&json!({
                "schemaVersion": 1, "backend": "tsrs-rust",
                "compiler": {"revision": COMPILER_REVISION},
                "projects": plan.projects, "fileProjects": plan.file_projects,
                "sourceOnlyCompilerOptions": projects::source_only_compiler_options(),
            }))
            .map_err(|error| error.to_string())
        })
        .map_err(|error| format!("cannot start project planning worker: {error}"))?
        .join()
        .map_err(|_| "Rust project planning worker failed".to_owned())?
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
    validate_rules(&request.rules)
}

fn validate_rules(rules: &[RuleConfig]) -> Result<(), String> {
    let mut names = BTreeSet::new();
    for rule in rules {
        if !names.insert(&rule.name) {
            return Err(format!("duplicate rule: {}", rule.name));
        }
        if !matches!(rule.severity, 1 | 2) {
            return Err(format!("invalid severity for {}", rule.name));
        }
        if !rules::supported(&rule.name) {
            return Err(format!("rule has not been migrated to Rust: {}", rule.name));
        }
        rules::validate_options(&rule.name, &rule.options)?;
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
    for file in &request.files {
        let name = path(&request.cwd, &file.path);
        if !seen.insert(name.clone()) {
            return Err(format!("duplicate admitted file: {name}"));
        }
        if let Some(text) = &file.text {
            overrides.insert(name.clone(), text.clone());
        } else if request.fix {
            // The compiler disk VFS strips BOMs. Fixing needs the original
            // source mark; freeze admitted disk sources as explicit overlays.
            if let Some(text) = source::read(&name)? {
                overrides.insert(name.clone(), text);
            }
        }
        inputs.push(name);
    }
    let session = SessionOwner::new(&request.cwd);
    let initial = observe(&session, &request, &inputs, &overrides)?;
    let outcome = if request.fix {
        autofix::run(initial, |sources| {
            let mut overlays = overrides.clone();
            overlays.extend(sources.clone());
            observe(&session, &request, &inputs, &overlays)
        })?
    } else {
        autofix::Outcome {
            files: initial.into_iter().map(|file| file.result).collect(),
            fixed_issues: 0,
        }
    };
    Ok(LintResult {
        schema_version: 1,
        backend: "tsrs-rust",
        compiler: json!({"revision": COMPILER_REVISION}),
        files: outcome.files,
        summary: LintSummary {
            fixed_issues: outcome.fixed_issues,
            threads: 1,
        },
    })
}

/// A generation contains only owned values. No compiler pointer outlives the
/// observation, so the next generation can rebuild all cross-file type state.
struct FileObservation {
    result: FileResult,
    source: String,
}

fn observe(
    owner: &SessionOwner,
    request: &Request,
    inputs: &[String],
    overrides: &BTreeMap<String, String>,
) -> Result<Vec<FileObservation>, String> {
    owner.fs.set_generation(&request.cwd, overrides)?;
    let session = &owner.session;
    let mut options = json!({});
    if let Some(project) = &request.project {
        options["openProjects"] = json!([path(&request.cwd, project)]);
    } else {
        options["createPrograms"] = json!([{"rootFiles": inputs, "compilerOptions": request.compiler_options.clone().unwrap_or_else(default_compiler_options)}]);
    }
    let response = session
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
    let snapshot = session.snapshot_data(handle).map_err(|e| e.to_string())?;
    let context = with_checker_lifetime(&Context::background(), CheckerLifetime::API);
    let mut files = Vec::new();
    for name in inputs {
        let mut owner = None;
        for id in &ids {
            let program = snapshot
                .get_program(&ID(id.clone()))
                .map_err(|e| e.to_string())?;
            if let Some(file) = program.get_source_file(name) {
                owner = Some((program, file));
                break;
            }
        }
        let (program, file) = owner
            .ok_or_else(|| format!("admitted file is missing from compiler project: {name}"))?;
        files.push(observe_file(
            name,
            file,
            program,
            &context,
            &request.rules,
            true,
        )?);
    }
    drop(snapshot);
    // Session owns a registry reference as well as the Arc above. Release each
    // successful generation before building the next; SessionOwner closes all
    // remaining handles on an error or panic.
    session
        .handle_request(
            "release",
            &serde_json::to_vec(&json!({"snapshot": handle})).map_err(|e| e.to_string())?,
        )
        .map_err(|e| format!("release compiler snapshot: {e}"))?;
    Ok(files)
}

fn observe_file(
    name: &str,
    file: P<SourceFile>,
    program: &tsrs_compiler::Program,
    context: &Context,
    file_rules: &[RuleConfig],
    has_type_information: bool,
) -> Result<FileObservation, String> {
    let syntactic = program.get_syntactic_diagnostics(context, Some(file));
    let has_syntax_errors = !syntactic.is_empty();
    let mut messages = Vec::new();
    if has_syntax_errors {
        for diagnostic in syntactic {
            let start = diagnostic
                .pos()
                .ok_or("synthetic syntax diagnostic start")? as usize;
            let end = diagnostic.end().ok_or("synthetic syntax diagnostic end")? as usize;
            let (offset, line, column) = position(file.text(), start);
            let (end_offset, end_line, end_column) = position(file.text(), end);
            messages.push(Diagnostic {
                rule_id: format!("TypeScript(TS{})", diagnostic.code()),
                message_id: String::new(),
                message: diagnostic.localize(),
                severity: 2,
                line,
                column,
                end_line,
                end_column,
                range: [offset, end_offset],
                fix: None,
                suggestions: vec![],
                edits: vec![],
            });
        }
    } else {
        let directives = directives::Directives::parse(file)?;
        directives
            .ensure_supported()
            .map_err(|error| format!("{error} in {}", file.file_name()))?;
        let _globals_scope = directives.enter_globals(file);
        let prepared_rules = rules::prepare(file_rules, file.text())?;
        let mut checker = program.get_type_checker_for_file_exclusive(context, file);
        visit(
            file.as_node(),
            &mut checker,
            &prepared_rules,
            &RuleContext {
                file,
                program,
                default_library: &|source| program.is_source_file_default_library(source.path()),
                directives: &directives,
                has_type_information,
            },
            &mut messages,
        )?;
        drop(checker);
    }
    // Equal starts retain emission order, matching the original engine.
    messages.sort_by_key(|message| message.range[0]);
    Ok(FileObservation {
        result: FileResult {
            file_path: name.to_owned(),
            messages,
            applied_rules: if has_syntax_errors {
                vec![]
            } else {
                file_rules.iter().map(|r| r.name.clone()).collect()
            },
            output: None,
            has_type_information: None,
            has_syntax_errors,
        },
        source: file.text().to_owned(),
    })
}

struct RuleContext<'a> {
    file: P<SourceFile>,
    program: &'a tsrs_compiler::Program,
    default_library: &'a dyn Fn(P<SourceFile>) -> bool,
    directives: &'a directives::Directives,
    has_type_information: bool,
}

fn visit(
    node: P<Node>,
    checker: &mut Checker,
    rules: &rules::PreparedRules<'_>,
    context: &RuleContext<'_>,
    out: &mut Vec<Diagnostic>,
) -> Result<(), String> {
    let file = context.file;
    for prepared in rules.for_kind(node.kind()) {
        let rule = prepared.config;
        for violation in prepared.check(
            node,
            file,
            checker,
            context.default_library,
            context.has_type_information,
            context.program,
        ) {
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
            if context
                .directives
                .is_rule_disabled(text, &rule.name, start)?
            {
                continue;
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
                message_id: violation.message_id.into(),
                message: violation.message,
                severity: rule.severity,
                line,
                column,
                end_line,
                end_column,
                range: [offset, end_offset],
                fix,
                suggestions,
                edits: violation.edits,
            });
        }
    }
    let mut error = None;
    node.for_each_child(&mut |child| {
        if let Err(value) = visit(child, checker, rules, context, out) {
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

/// Immutable scheduling metadata without creating compiler state.
pub fn rule_capabilities_json(request: &str) -> Result<String, String> {
    let names: Vec<String> = serde_json::from_str(request)
        .map_err(|error| format!("invalid rule capability request: {error}"))?;
    let mut result = BTreeMap::new();
    for name in names {
        if !rules::supported(&name) {
            return Err(format!("rule has not been migrated to Rust: {name}"));
        }
        let typed = rules::requires_type_information(&name);
        result.insert(name, typed);
    }
    serde_json::to_string(&result).map_err(|error| error.to_string())
}
