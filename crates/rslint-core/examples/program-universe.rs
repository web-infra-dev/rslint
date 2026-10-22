//! Read-only correctness probe of the production Program construction path.
//!
//! Compile the actual private modules rather than introducing a product API or
//! reimplementing project construction in this observer. Only owned strings
//! and JSON leave the worker. This tool is not a performance benchmark.

#[allow(dead_code)]
#[path = "../src/programs.rs"]
mod programs;
#[allow(dead_code)]
#[path = "../src/projects.rs"]
mod projects;
#[allow(dead_code)]
#[path = "../src/source.rs"]
mod source;

use std::collections::{BTreeMap, BTreeSet};
use std::io::{self, Read, Write};
use std::sync::Arc;

use serde::Deserialize;
use serde_json::{Value, json};
use tsrs_api::{Session, SessionOptions};
use tsrs_vfs::{FS, bundled, osvfs};

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
struct Request {
    schema_version: u32,
    cwd: String,
    source_only_current_directory: Option<String>,
    files: Vec<Input>,
    #[serde(default)]
    single_threaded: bool,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct Input {
    path: String,
    policy: projects::ProjectPolicy,
}

struct SessionOwner {
    session: Arc<Session>,
    fs: Arc<source::SourceFs>,
}

impl Drop for SessionOwner {
    fn drop(&mut self) {
        self.session.close();
        self.fs.clear();
    }
}

fn observe(request: Request) -> Result<Value, String> {
    if request.schema_version != 1
        || !std::path::Path::new(&request.cwd).is_absolute()
        || request.files.is_empty()
    {
        return Err("observer requires schemaVersion 1, absolute cwd and admitted files".into());
    }
    let mut seen_inputs = BTreeSet::new();
    let inputs = request
        .files
        .into_iter()
        .map(|input| {
            if !seen_inputs.insert(input.path.clone()) {
                return Err(format!("duplicate admitted input: {}", input.path));
            }
            Ok(projects::InputProject {
                path: input.path,
                policy: input.policy,
            })
        })
        .collect::<Result<Vec<_>, String>>()?;
    let base: Arc<dyn FS> = Arc::new(bundled::wrap_fs(osvfs::fs()));
    let fs = Arc::new(source::SourceFs::new(base));
    fs.set_generation(&request.cwd, &BTreeMap::new())?;
    let owner = SessionOwner {
        session: Session::new(SessionOptions::new(
            request.cwd.clone(),
            bundled::default_lib_path(),
            fs.clone(),
            false,
        )),
        fs,
    };
    let plan = projects::select(&owner.session, &inputs)?;
    let input_paths = inputs
        .iter()
        .map(|input| input.path.clone())
        .collect::<Vec<_>>();
    let programs = programs::Programs::build(
        &plan,
        &input_paths,
        request.source_only_current_directory.as_deref().unwrap_or(&request.cwd),
        owner.fs.clone(),
        request.single_threaded,
    )?;
    // Retain every distinct compiler-provided text, including variants of a
    // path in different Programs. Deduplication affects transport only; every
    // source membership/flag record remains present in its owning Program.
    let mut text_ids = BTreeMap::<String, usize>::new();
    let mut texts = Vec::new();
    let mut observed = BTreeSet::new();
    let mut universes = Vec::new();
    for (input_index, project_index) in plan.file_projects.iter().enumerate() {
        if !observed.insert(*project_index) {
            continue;
        }
        let lease = programs.for_file(input_index)?;
        let program = lease.program();
        let mut source_files = Vec::new();
        for file in program.source_files() {
            let text_id = *text_ids.entry(file.text().to_owned()).or_insert_with(|| {
                let index = texts.len();
                texts.push(file.text().to_owned());
                index
            });
            source_files.push(json!({
                "path": file.file_name(),
                "textIndex": text_id,
                "sourceBytes": file.text().len(),
                "isDefaultLibrary": program.is_source_file_default_library(file.path()),
                "isExternalLibrary": program.is_source_file_from_external_library(*file),
            }));
        }
        let options = tsrs_tsoptions::gojson::compiler_options_to_go_json(&program.options());
        let options: Value = serde_json::from_str(&tsrs_core::json::marshal(&options)?)
            .map_err(|error| error.to_string())?;
        universes.push(json!({
            "projectIndex": project_index,
            "hasTypeInformation": lease.has_type_information,
            "currentDirectory": program.get_current_directory(),
            "compilerOptions": options,
            "sourceFiles": source_files,
        }));
    }
    Ok(json!({
        "schemaVersion": 1,
        "backend": "tsrs-rust",
        "compiler": {"revision": rslint_core::COMPILER_REVISION},
        "singleThreaded": request.single_threaded,
        "projects": plan.projects,
        "fileProjects": plan.file_projects,
        "universes": universes,
        "texts": texts,
    }))
}

fn run() -> Result<(), String> {
    let mut input = String::new();
    io::stdin()
        .read_to_string(&mut input)
        .map_err(|error| error.to_string())?;
    let request: Request = serde_json::from_str(&input).map_err(|error| error.to_string())?;
    let value = std::thread::Builder::new()
        .name("rslint-program-universe".into())
        .stack_size(512 << 20)
        .spawn(move || observe(request))
        .map_err(|error| error.to_string())?
        .join()
        .map_err(|_| "Program universe observer panicked".to_owned())??;
    let output = io::stdout();
    let mut output = io::BufWriter::new(output.lock());
    serde_json::to_writer(&mut output, &value).map_err(|error| error.to_string())?;
    output.write_all(b"\n").map_err(|error| error.to_string())
}

fn main() {
    if let Err(error) = run() {
        eprintln!("program-universe: {error}");
        std::process::exit(1);
    }
}
