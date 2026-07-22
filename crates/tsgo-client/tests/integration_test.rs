use std::env;
use std::ffi::OsStr;
use std::path::PathBuf;
use tsgo_client::Api;
use tsgo_client::client::{Client, Options};
use tsgo_client::symbolflags::SymbolFlags;

use serde::Serialize;

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
    let possible_paths = ["target/tsgo", "bin/tsgo"];

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
    let api =
        Api::with_uninitialized_client(uninitialized_client).expect("Failed to initialize API");

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
    assert_eq!(
        project.module_exports.len(),
        project.module_list.len(),
        "Expected exports for every module"
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

    let index_module = project
        .module_list
        .iter()
        .position(|path| path.ends_with("/src/index.ts"))
        .expect("Expected index.ts module");
    let export_names = project.module_exports[index_module]
        .iter()
        .filter_map(|symbol_id| {
            project
                .semantic
                .symtab
                .iter()
                .find(|(id, _)| id == symbol_id)
                .map(|(_, data)| String::from_utf8_lossy(&data.name).into_owned())
        })
        .collect::<Vec<_>>();
    assert!(export_names.iter().any(|name| name == "greet"));
    assert!(export_names.iter().any(|name| name == "add"));
    assert!(export_names.iter().any(|name| name == "Calculator"));
    assert!(!export_names.iter().any(|name| name == "Person"));

    println!("\n✓ Integration test passed!");
    println!("  - Loaded {} source files", project.source_files.len());
    println!("  - Found {} symbols", project.semantic.symtab.len());
    println!("  - Found {} types", project.semantic.typetab.len());
}

#[test]
fn test_runtime_module_exports() {
    let tsgo_path = get_tsgo_path().expect("Could not find tsgo executable");
    let fixture_dir = get_fixtures_dir().join("module-exports");
    let config_file = fixture_dir.join("tsconfig.json");
    let options = Options {
        cwd: Some(fixture_dir),
        log_file: None,
        config_file: config_file.to_string_lossy().to_string(),
    };
    let client = Client::builder(OsStr::new(&tsgo_path), options)
        .build()
        .expect("Failed to build client");
    let api = Api::with_uninitialized_client(client).expect("Failed to initialize API");
    let mut buffer = Vec::new();
    let project = api
        .load_project(&mut buffer)
        .expect("Failed to load project");
    let index_module = project
        .module_list
        .iter()
        .position(|path| path.ends_with("/src/index.ts"))
        .expect("Expected index.ts module");
    let export_names = project.module_exports[index_module]
        .iter()
        .map(|symbol_id| {
            project
                .semantic
                .symtab
                .iter()
                .find(|(id, _)| id == symbol_id)
                .map(|(_, data)| String::from_utf8_lossy(&data.name).into_owned())
                .expect("Expected exported symbol in semantic table")
        })
        .collect::<Vec<_>>();

    assert!(export_names.iter().any(|name| name == "directValue"));
    assert!(export_names.iter().any(|name| name == "barrelValue"));
    assert!(export_names.iter().any(|name| name == "otherBarrelValue"));
    assert!(!export_names.iter().any(|name| name == "DirectType"));
    assert!(!export_names.iter().any(|name| name == "BarrelType"));
    assert!(!export_names.iter().any(|name| name == "RuntimeTypeOnly"));
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

    let shorthand_ts = src_dir.join("shorthand.ts");
    assert!(shorthand_ts.exists(), "shorthand.ts should exist");
}

#[derive(Debug, Serialize, PartialEq, Eq, PartialOrd, Ord)]
struct ShorthandSymbolMapping {
    source_node_span: String,
    target_symbol_name: String,
    target_decl_span: String,
}

#[test]
fn test_get_shorthand_assignment_value_symbol() {
    let tsgo_path = get_tsgo_path().expect(
        "Could not find tsgo executable. \
         Please build tsgo first or ensure it's in your PATH.",
    );

    let fixture_dir = get_fixtures_dir().join("simple-project");
    let config_file = fixture_dir.join("tsconfig.json");

    let options = Options {
        cwd: Some(fixture_dir.clone()),
        log_file: None,
        config_file: config_file.to_string_lossy().to_string(),
    };

    let uninitialized_client = Client::builder(OsStr::new(&tsgo_path), options)
        .build()
        .expect("Failed to build client");

    let api =
        Api::with_uninitialized_client(uninitialized_client).expect("Failed to initialize API");

    let mut buffer = Vec::new();
    let project = api
        .load_project(&mut buffer)
        .expect("Failed to load project");

    let semantic = &project.semantic;

    // Collect shorthand symbol mappings (source -> target)
    let mut shorthand_mappings = Vec::new();

    for (node_ref, _source_symbol_id) in &semantic.node2sym {
        if let Some(target_symbol_id) = semantic.get_shorthand_assignment_value_symbol(node_ref) {
            // Get the target symbol data
            if let Some((_, target_symbol_data)) = semantic
                .symtab
                .iter()
                .find(|(id, _)| *id == target_symbol_id)
            {
                let flags = SymbolFlags::from_bits_truncate(target_symbol_data.flags);

                // Verify it has VALUE or ALIAS flags
                assert!(
                    flags.intersects(SymbolFlags::VALUE | SymbolFlags::ALIAS),
                    "Shorthand value symbol should have VALUE or ALIAS flags, got: {flags:?}"
                );

                let target_symbol_name =
                    String::from_utf8_lossy(&target_symbol_data.name).to_string();

                // Only collect mappings from our test symbols
                if [
                    "name", "age", "username", "userAge", "isActive", "id", "email",
                ]
                .contains(&target_symbol_name.as_str())
                {
                    // Format source node span
                    let source_span = format!(
                        "{}:{}..{}",
                        node_ref.sourcefile_id, node_ref.start, node_ref.end
                    );

                    // Format target declaration span
                    let target_decl_span = if let Some(decl) = &target_symbol_data.decl {
                        format!("{}:{}..{}", decl.sourcefile_id, decl.start, decl.end)
                    } else {
                        "unknown".to_string()
                    };

                    shorthand_mappings.push(ShorthandSymbolMapping {
                        source_node_span: source_span,
                        target_symbol_name: target_symbol_name.clone(),
                        target_decl_span,
                    });
                }
            }
        }
    }

    // Sort by source span for consistent snapshots
    shorthand_mappings.sort();

    println!("\n✓ Shorthand assignment test passed!");
    println!(
        "  - Found {} unique shorthand symbol mappings",
        shorthand_mappings.len()
    );
    println!(
        "  - Symbols: {:?}",
        shorthand_mappings
            .iter()
            .map(|m| &m.target_symbol_name)
            .collect::<Vec<_>>()
    );

    // Verify we found the expected symbols
    assert!(
        shorthand_mappings.len() >= 7,
        "Expected at least 7 shorthand symbols, found {}",
        shorthand_mappings.len()
    );

    // Generate snapshot
    insta::assert_json_snapshot!(shorthand_mappings);
}

#[derive(Debug, Serialize, PartialEq, Eq, PartialOrd, Ord)]
struct ParameterPropertySymbolMapping {
    source_node_span: String,
    primary_symbol_name: String,
    extra_symbol_name: String,
    primary_decl_span: String,
    extra_decl_span: String,
}

#[test]
fn test_get_parameter_property_symbols() {
    let tsgo_path = get_tsgo_path().expect(
        "Could not find tsgo executable. \
         Please build tsgo first or ensure it's in your PATH.",
    );

    let fixture_dir = get_fixtures_dir().join("simple-project");
    let config_file = fixture_dir.join("tsconfig.json");

    let options = Options {
        cwd: Some(fixture_dir.clone()),
        log_file: None,
        config_file: config_file.to_string_lossy().to_string(),
    };

    let uninitialized_client = Client::builder(OsStr::new(&tsgo_path), options)
        .build()
        .expect("Failed to build client");

    let api =
        Api::with_uninitialized_client(uninitialized_client).expect("Failed to initialize API");

    let mut buffer = Vec::new();
    let project = api
        .load_project(&mut buffer)
        .expect("Failed to load project");

    let semantic = &project.semantic;
    let mut mappings = Vec::new();

    for (node_ref, extra_symbol_id) in &semantic.parameter_property_symbols {
        let Some(extra_symbol_id_from_lookup) = semantic.get_parameter_property_symbol(node_ref)
        else {
            panic!("parameter property lookup should find the recorded location");
        };

        assert_eq!(*extra_symbol_id, extra_symbol_id_from_lookup);

        let Some((_, primary_symbol_id)) = semantic.node2sym.iter().find(|(location, _)| {
            location.sourcefile_id == node_ref.sourcefile_id
                && location.start == node_ref.start
                && location.end == node_ref.end
        }) else {
            panic!("primary parameter property symbol should be present in node2sym");
        };
        let primary_symbol_id = *primary_symbol_id;
        let extra_symbol_id = *extra_symbol_id;

        assert_ne!(
            primary_symbol_id, extra_symbol_id,
            "extra parameter property symbol should differ from node2sym"
        );

        let Some((_, primary_symbol_data)) = semantic
            .symtab
            .iter()
            .find(|(id, _)| *id == primary_symbol_id)
        else {
            panic!("primary symbol should be present in symtab");
        };

        let Some((_, extra_symbol_data)) = semantic
            .symtab
            .iter()
            .find(|(id, _)| *id == extra_symbol_id)
        else {
            panic!("extra symbol should be present in symtab");
        };

        let primary_flags = SymbolFlags::from_bits_truncate(primary_symbol_data.flags);
        let extra_flags = SymbolFlags::from_bits_truncate(extra_symbol_data.flags);
        assert!(
            primary_flags.intersects(SymbolFlags::PROPERTY | SymbolFlags::CLASS_MEMBER),
            "node2sym for parameter property should be the property symbol, got: {primary_flags:?}"
        );
        assert!(
            extra_flags.contains(SymbolFlags::FUNCTION_SCOPED_VARIABLE),
            "extra parameter property symbol should be the parameter symbol, got: {extra_flags:?}"
        );

        let primary_symbol_name = String::from_utf8_lossy(&primary_symbol_data.name).to_string();
        let extra_symbol_name = String::from_utf8_lossy(&extra_symbol_data.name).to_string();

        if ["testType", "count", "enabled"].contains(&primary_symbol_name.as_str()) {
            assert_eq!(primary_symbol_name, extra_symbol_name);

            let source_node_span = format!(
                "{}:{}..{}",
                node_ref.sourcefile_id, node_ref.start, node_ref.end
            );
            let primary_decl_span = if let Some(decl) = &primary_symbol_data.decl {
                format!("{}:{}..{}", decl.sourcefile_id, decl.start, decl.end)
            } else {
                "unknown".to_string()
            };
            let extra_decl_span = if let Some(decl) = &extra_symbol_data.decl {
                format!("{}:{}..{}", decl.sourcefile_id, decl.start, decl.end)
            } else {
                "unknown".to_string()
            };

            mappings.push(ParameterPropertySymbolMapping {
                source_node_span,
                primary_symbol_name,
                extra_symbol_name,
                primary_decl_span,
                extra_decl_span,
            });
        }
    }

    mappings.sort();

    assert_eq!(
        mappings.len(),
        3,
        "Expected 3 parameter property mappings, found {}",
        mappings.len()
    );

    insta::assert_json_snapshot!(mappings);
}
