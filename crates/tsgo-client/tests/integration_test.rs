use std::env;
use std::ffi::OsStr;
use std::path::PathBuf;
use tsgo_client::client::{Client, Options};
use tsgo_client::Api;

/// Get the path to the tsgo executable for testing.
/// Tries to build tsgo from cmd/tsgo or finds an existing binary.
fn get_tsgo_path() -> Option<PathBuf> {
    let manifest_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let repo_root = manifest_dir
        .parent()
        .and_then(|p| p.parent())
        .expect("Could not find repo root");

    // Try to build tsgo from cmd/tsgo (this is the correct version for tsgo-client)
    let tsgo_output = repo_root.join("target/tsgo");
    let cmd_tsgo_dir = repo_root.join("cmd/tsgo");

    if cmd_tsgo_dir.exists() {
        eprintln!("Building tsgo from cmd/tsgo...");
        let status = std::process::Command::new("go")
            .args(["build", "-o"])
            .arg(&tsgo_output)
            .arg("./cmd/tsgo")
            .current_dir(repo_root)
            .status();

        if status.is_ok() && tsgo_output.exists() {
            eprintln!("✓ Built tsgo successfully");
            return Some(tsgo_output);
        }
    }

    // Fall back to searching for existing binaries
    let possible_paths = [
        "target/tsgo",
        "bin/tsgo",
    ];

    for path in &possible_paths {
        let full_path = repo_root.join(path);
        if full_path.exists() {
            return Some(full_path);
        }
    }

    None
}

/// Get the path to the test fixtures directory
fn get_fixtures_dir() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures")
}

#[test]
fn test_tsgo_integration_simple_project() {
    let tsgo_path = get_tsgo_path().expect(
        "Could not find tsgo executable. \
         Please build tsgo first or ensure it's in your PATH.",
    );

    let fixture_dir = get_fixtures_dir().join("simple-project");
    let config_file = fixture_dir.join("tsconfig.json");

    assert!(
        fixture_dir.exists(),
        "Test fixture directory does not exist: {fixture_dir:?}"
    );
    assert!(
        config_file.exists(),
        "tsconfig.json does not exist: {config_file:?}"
    );

    // Set up options for the tsgo client
    let options = Options {
        cwd: Some(fixture_dir.clone()),
        log_file: None,
        config_file: config_file.to_string_lossy().to_string(),
    };

    // Build and spawn the tsgo process
    let uninitialized_client = Client::builder(OsStr::new(&tsgo_path), options)
        .build()
        .expect("Failed to build client");

    // Initialize the API
    let api = Api::with_uninitialized_client(uninitialized_client)
        .expect("Failed to initialize API");

    // Load the TypeScript project
    let mut buffer = Vec::new();
    let project = api
        .load_project(&mut buffer)
        .expect("Failed to load project");

    // Verify we got project data
    println!("Root files: {:?}", project.root_files);
    println!("Number of source files: {}", project.source_files.len());
    println!("Number of modules: {}", project.module_list.len());
    println!("Number of diagnostics: {}", project.diagnostics.len());

    // Basic assertions
    assert!(
        !project.source_files.is_empty(),
        "Expected at least one source file"
    );
    assert!(
        !project.module_list.is_empty(),
        "Expected at least one module"
    );

    // Check semantic data exists
    assert!(
        !project.semantic.symtab.is_empty(),
        "Expected symbols in symbol table"
    );
    assert!(
        !project.semantic.typetab.is_empty(),
        "Expected types in type table"
    );

    // Verify primitive types are set
    assert_ne!(project.semantic.primtypes.string, 0);
    assert_ne!(project.semantic.primtypes.number, 0);
    assert_ne!(project.semantic.primtypes.any, 0);

    println!("\n✓ Integration test passed!");
    println!("  - Loaded {} source files", project.source_files.len());
    println!("  - Found {} symbols", project.semantic.symtab.len());
    println!("  - Found {} types", project.semantic.typetab.len());
}

#[test]
fn test_tsgo_client_builder() {
    let tsgo_path = get_tsgo_path().expect("Could not find tsgo executable");
    let fixture_dir = get_fixtures_dir().join("simple-project");
    let config_file = fixture_dir.join("tsconfig.json");

    // Test builder pattern
    let options = Options {
        cwd: Some(fixture_dir.clone()),
        log_file: None,
        config_file: config_file.to_string_lossy().to_string(),
    };

    let client = Client::builder(OsStr::new(&tsgo_path), options)
        .log_file("test.log".to_string())
        .build();

    assert!(client.is_ok(), "Failed to build client with builder");
}

#[test]
fn test_fixture_structure() {
    let fixture_dir = get_fixtures_dir().join("simple-project");
    assert!(fixture_dir.exists(), "Fixture directory should exist");

    let tsconfig = fixture_dir.join("tsconfig.json");
    assert!(tsconfig.exists(), "tsconfig.json should exist");

    let src_dir = fixture_dir.join("src");
    assert!(src_dir.exists(), "src directory should exist");

    let index_ts = src_dir.join("index.ts");
    assert!(index_ts.exists(), "index.ts should exist");

    let utils_ts = src_dir.join("utils.ts");
    assert!(utils_ts.exists(), "utils.ts should exist");
}
