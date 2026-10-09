// cspell:ignore symtab typetab
use std::env;
use std::path::{Path, PathBuf};
use std::sync::OnceLock;
use tsgo_client::Api;
use tsgo_client::client::{Client, Options};
use tsgo_client::proto::{NodeReference, ProjectResponse, Semantic, SymbolData, TypeData};

/// Build the matching server once, including when integration tests run concurrently.
pub(crate) fn get_tsgo_path() -> Option<PathBuf> {
    static SERVER: OnceLock<PathBuf> = OnceLock::new();
    Some(
        SERVER
            .get_or_init(|| {
                let repo_root = Path::new(env!("CARGO_MANIFEST_DIR"))
                    .parent()
                    .unwrap()
                    .parent()
                    .unwrap();
                let tsgo_output = repo_root.join("target/tsgo");
                std::fs::create_dir_all(tsgo_output.parent().unwrap()).unwrap();
                let status = std::process::Command::new("go")
                    .args(["build", "-o"])
                    .arg(&tsgo_output)
                    .arg("./cmd/tsgo")
                    .current_dir(repo_root)
                    .status()
                    .expect("Failed to build tsgo");
                assert!(status.success(), "Building tsgo failed: {status}");
                tsgo_output
            })
            .clone(),
    )
}

pub(crate) fn load_project<'buf>(
    directory: &Path,
    buffer: &'buf mut Vec<u8>,
) -> ProjectResponse<'buf> {
    let server = get_tsgo_path().expect("Could not find tsgo executable");
    let options = Options {
        cwd: Some(directory.to_path_buf()),
        config_file: directory
            .join("tsconfig.json")
            .to_string_lossy()
            .into_owned(),
        ..Default::default()
    };
    let client = Client::builder(server.as_os_str(), options)
        .build()
        .expect("Failed to build client");
    let project = Api::with_uninitialized_client(client)
        .expect("Failed to initialize API")
        .load_project(buffer)
        .expect("Failed to load project");
    assert!(
        project.diagnostics.is_empty(),
        "Unexpected diagnostics: {:?}",
        project.diagnostics
    );
    project
}

pub(crate) fn project_from_source<'buf>(
    source: &str,
    buffer: &'buf mut Vec<u8>,
) -> (tempfile::TempDir, ProjectResponse<'buf>) {
    let directory = tempfile::tempdir().unwrap();
    std::fs::write(
        directory.path().join("tsconfig.json"),
        r#"{"include":["./index.ts"]}"#,
    )
    .unwrap();
    std::fs::write(directory.path().join("index.ts"), source).unwrap();
    let project = load_project(directory.path(), buffer);
    (directory, project)
}

pub(crate) fn source_file_id(project: &ProjectResponse<'_>, suffix: &str) -> u32 {
    project
        .module_list
        .iter()
        .position(|path| path.ends_with(suffix))
        .unwrap_or_else(|| panic!("Missing source file {suffix}")) as u32
}

pub(crate) fn symbol_data(semantic: &Semantic, id: u32) -> &SymbolData {
    &semantic
        .symtab
        .iter()
        .find(|(symbol, _)| *symbol == id)
        .unwrap_or_else(|| panic!("Missing symbol {id}"))
        .1
}

pub(crate) fn named_symbol<'a>(
    semantic: &'a Semantic,
    file: u32,
    name: &[u8],
) -> (u32, &'a SymbolData) {
    let matches = semantic
        .symtab
        .iter()
        .filter(|(_, data)| {
            data.name == name
                && data
                    .decl
                    .as_ref()
                    .is_some_and(|decl| decl.sourcefile_id == file)
        })
        .collect::<Vec<_>>();
    assert_eq!(
        matches.len(),
        1,
        "Expected one local symbol {}",
        String::from_utf8_lossy(name)
    );
    (matches[0].0, &matches[0].1)
}

pub(crate) fn type_data(semantic: &Semantic, id: u32) -> &TypeData {
    &semantic
        .typetab
        .iter()
        .find(|(ty, _)| *ty == id)
        .unwrap_or_else(|| panic!("Missing type {id}"))
        .1
}

pub(crate) fn node_type(semantic: &Semantic, file: u32, start: u32, end: u32) -> &TypeData {
    let id = semantic
        .node2type
        .iter()
        .find(|(node, _)| node.sourcefile_id == file && node.start == start && node.end == end)
        .unwrap_or_else(|| panic!("Missing node type at {start}..{end}"))
        .1;
    type_data(semantic, id)
}

pub(crate) fn external_symbol(semantic: &Semantic, namespace: &str, name: &str) -> u32 {
    let matches = semantic
        .external_symbols
        .iter()
        .filter(|symbol| symbol.namespace == namespace.as_bytes() && symbol.name == name.as_bytes())
        .collect::<Vec<_>>();
    assert_eq!(
        matches.len(),
        1,
        "Expected one external symbol {namespace}:{name}"
    );
    matches[0].symbol_id
}

pub(crate) fn same_location(left: &NodeReference, right: &NodeReference) -> bool {
    left.sourcefile_id == right.sourcefile_id && left.start == right.start && left.end == right.end
}

pub(crate) fn node_symbol(semantic: &Semantic, node: &NodeReference) -> u32 {
    semantic
        .node2sym
        .iter()
        .find(|(location, _)| same_location(location, node))
        .unwrap_or_else(|| panic!("Missing node symbol at {node:?}"))
        .1
}

/// Get the path to the test fixtures directory
pub(crate) fn get_fixtures_dir() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures")
}
