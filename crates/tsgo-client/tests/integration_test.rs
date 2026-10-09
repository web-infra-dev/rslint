// cspell:ignore symbolflags symtab typetab primtypes
#[path = "common/helper.rs"]
mod helper;

use helper::{
    external_symbol, get_fixtures_dir, get_tsgo_path, load_project, named_symbol, node_symbol,
    node_type, project_from_source, same_location, source_file_id, symbol_data, type_data,
};
use serde::Serialize;
use std::ffi::OsStr;
use tsgo_client::Api;
use tsgo_client::client::{Client, Options};
use tsgo_client::proto::NodeReference;
use tsgo_client::symbolflags::SymbolFlags;

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

    // Shared DOM members retain every observed qualified name through transport.
    let document_listener = project
        .semantic
        .external_symbols
        .iter()
        .find(|symbol| symbol.namespace == b"global" && symbol.name == b"document.addEventListener")
        .expect("Expected document.addEventListener metadata");
    assert!(
        project.semantic.external_symbols.iter().any(|symbol| {
            symbol.symbol_id == document_listener.symbol_id
                && symbol.namespace == b"global"
                && symbol.name != document_listener.name
        }),
        "Expected multiple qualified names for the shared listener symbol"
    );

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
fn test_type_literal_symbol_has_symbol_data() {
    let tsgo_path = get_tsgo_path().expect("Could not find tsgo executable");
    let fixture_dir = get_fixtures_dir().join("simple-project");
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
    let semantic = &project.semantic;

    let index_sourcefile_id = project
        .module_list
        .iter()
        .position(|path| path.ends_with("/src/index.ts"))
        .expect("Expected index.ts module") as u32;
    let parameter_symbol_id = semantic
        .symtab
        .iter()
        .find(|(_, data)| {
            data.name == b"o"
                && data
                    .decl
                    .as_ref()
                    .is_some_and(|decl| decl.sourcefile_id == index_sourcefile_id)
        })
        .map(|(id, _)| *id)
        .expect("Expected symbol data for parameter o");
    let type_id = semantic
        .sym2type
        .iter()
        .find(|(symbol_id, _)| *symbol_id == parameter_symbol_id)
        .map(|(_, type_id)| *type_id)
        .expect("Expected type for parameter o");
    let type_symbol_id = semantic
        .typetab
        .iter()
        .find(|(id, _)| *id == type_id)
        .and_then(|(_, data)| data.symbol)
        .expect("Expected the type literal to have a symbol");
    let type_symbol_data = semantic
        .symtab
        .iter()
        .find(|(id, _)| *id == type_symbol_id)
        .map(|(_, data)| data)
        .expect("Expected type literal symbol data in symtab");

    assert!(
        SymbolFlags::from_bits_truncate(type_symbol_data.flags).contains(SymbolFlags::TYPE_LITERAL),
        "Expected a type literal symbol, got flags {}",
        type_symbol_data.flags
    );
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
struct ShorthandBindingSymbolMapping {
    local_decl_span: String,
    local_symbol_name: String,
    property_symbol_name: String,
    property_decl_span: String,
}

#[test]
fn test_get_shorthand_binding_property_symbol() {
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
    let client = Client::builder(OsStr::new(&tsgo_path), options)
        .build()
        .expect("Failed to build client");
    let api = Api::with_uninitialized_client(client).expect("Failed to initialize API");
    let mut buffer = Vec::new();
    let project = api
        .load_project(&mut buffer)
        .expect("Failed to load project");
    let semantic = &project.semantic;
    let mut mappings = Vec::new();

    for (local_symbol_id, property_symbol_id) in &semantic.shorthand_binding_symbols {
        let Some(property_symbol_id_from_lookup) =
            semantic.get_shorthand_binding_property_symbol(*local_symbol_id)
        else {
            panic!("shorthand binding property lookup should find the local symbol");
        };
        assert_eq!(*property_symbol_id, property_symbol_id_from_lookup);

        let Some((_, property_symbol_data)) = semantic
            .symtab
            .iter()
            .find(|(id, _)| id == property_symbol_id)
        else {
            panic!("property symbol should be present in symtab");
        };
        let property_symbol_name = String::from_utf8_lossy(&property_symbol_data.name).to_string();
        if !["destructured", "defaulted"].contains(&property_symbol_name.as_str()) {
            continue;
        }

        assert_ne!(
            local_symbol_id, property_symbol_id,
            "property symbol should differ from the local binding symbol"
        );

        let Some((_, local_symbol_data)) =
            semantic.symtab.iter().find(|(id, _)| id == local_symbol_id)
        else {
            panic!("local binding symbol should be present in symtab");
        };
        let local_symbol_name = String::from_utf8_lossy(&local_symbol_data.name).to_string();
        assert_eq!(local_symbol_name, property_symbol_name);

        let property_flags = SymbolFlags::from_bits_truncate(property_symbol_data.flags);
        assert!(
            property_flags.contains(SymbolFlags::PROPERTY),
            "shorthand binding target should be a property, got: {property_flags:?}"
        );

        let local_decl_span = if let Some(decl) = &local_symbol_data.decl {
            format!("{}:{}..{}", decl.sourcefile_id, decl.start, decl.end)
        } else {
            "unknown".to_string()
        };
        let property_decl_span = if let Some(decl) = &property_symbol_data.decl {
            format!("{}:{}..{}", decl.sourcefile_id, decl.start, decl.end)
        } else {
            "unknown".to_string()
        };
        mappings.push(ShorthandBindingSymbolMapping {
            local_decl_span,
            local_symbol_name,
            property_symbol_name,
            property_decl_span,
        });
    }

    mappings.sort();
    assert_eq!(
        mappings.len(),
        2,
        "expected mappings for the plain and defaulted shorthand bindings"
    );
    insta::assert_json_snapshot!(mappings);
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

#[test]
fn test_element_access_resolves_property_and_value_type() {
    let source = r#"const obj = { a: 1 }; const value = obj["a"];"#;
    let mut buffer = Vec::new();
    let (_directory, project) = project_from_source(source, &mut buffer);
    let file = source_file_id(&project, "/index.ts");
    let semantic = &project.semantic;
    let offset = source.find(r#""a""#).unwrap();
    let property = node_symbol(
        semantic,
        &NodeReference {
            sourcefile_id: file,
            start: source[..offset].encode_utf16().count() as u32,
            end: source[..offset + 3].encode_utf16().count() as u32,
        },
    );
    let property = symbol_data(semantic, property);
    assert_eq!(property.name, b"a");
    assert!(SymbolFlags::from_bits_truncate(property.flags).contains(SymbolFlags::PROPERTY));

    let (value, _) = named_symbol(semantic, file, b"value");
    let value_type = semantic
        .sym2type
        .iter()
        .find(|(id, _)| *id == value)
        .expect("Missing element access result type")
        .1;
    assert_eq!(value_type, semantic.primtypes.number);
}

#[test]
fn test_type_declaration_and_object_literal_symbols() {
    let source = "class ClassType {}\ninterface InterfaceType {}\nconst classValue = new ClassType();\nlet interfaceValue: InterfaceType;\nconst anonymousValue = {};\nconst primitiveValue = 1;";
    let mut buffer = Vec::new();
    let (_directory, project) = project_from_source(source, &mut buffer);
    let file = source_file_id(&project, "/index.ts");
    let semantic = &project.semantic;
    for (name, expected_name, expected_flags) in [
        ("classValue", "ClassType", SymbolFlags::CLASS),
        ("interfaceValue", "InterfaceType", SymbolFlags::INTERFACE),
    ] {
        let (symbol, _) = named_symbol(semantic, file, name.as_bytes());
        let ty = semantic
            .sym2type
            .iter()
            .find(|(id, _)| *id == symbol)
            .expect("Missing symbol type")
            .1;
        let declaration = type_data(semantic, ty)
            .symbol
            .expect("Missing type declaration symbol");
        let declaration = symbol_data(semantic, declaration);
        assert_eq!(declaration.name, expected_name.as_bytes());
        assert!(SymbolFlags::from_bits_truncate(declaration.flags).contains(expected_flags));
    }
    let (anonymous, _) = named_symbol(semantic, file, b"anonymousValue");
    let anonymous_type = semantic
        .sym2type
        .iter()
        .find(|(id, _)| *id == anonymous)
        .expect("Missing object literal type")
        .1;
    let object_symbol = type_data(semantic, anonymous_type)
        .symbol
        .expect("Missing object literal symbol");
    assert!(
        SymbolFlags::from_bits_truncate(symbol_data(semantic, object_symbol).flags)
            .contains(SymbolFlags::OBJECT_LITERAL)
    );
    let (primitive, _) = named_symbol(semantic, file, b"primitiveValue");
    let primitive_type = semantic
        .sym2type
        .iter()
        .find(|(id, _)| *id == primitive)
        .unwrap()
        .1;
    assert!(type_data(semantic, primitive_type).symbol.is_none());
}

#[test]
fn test_import_aliases_preserve_merged_local_values() {
    let mut buffer = Vec::new();
    let project = load_project(&get_fixtures_dir().join("semantic-project"), &mut buffer);
    let semantic = &project.semantic;
    let file = source_file_id(&project, "/index.ts");
    let source = include_str!("fixtures/semantic-project/index.ts");
    let offset = source.find("alias }").unwrap();
    // Query the imported identifier in the data returned to Rust.
    let alias = node_symbol(
        semantic,
        &NodeReference {
            sourcefile_id: file,
            start: source[..offset].trim_end().encode_utf16().count() as u32,
            end: source[..offset + "alias".len()].encode_utf16().count() as u32,
        },
    );
    let data = symbol_data(semantic, alias);
    assert_eq!(data.name, b"alias");
    let target = semantic
        .alias_symbols
        .iter()
        .find(|(id, _)| *id == alias)
        .expect("Missing import alias target")
        .1;
    assert_eq!(symbol_data(semantic, target).name, b"originalValue");
    assert_eq!(
        symbol_data(semantic, target)
            .decl
            .as_ref()
            .unwrap()
            .sourcefile_id,
        source_file_id(&project, "/module.ts")
    );

    let merged_file = source_file_id(&project, "/merged.ts");
    let (local, data) = named_symbol(semantic, merged_file, b"SymbolLinks");
    assert!(SymbolFlags::from_bits_truncate(data.flags).intersects(SymbolFlags::VALUE));
    assert!(!semantic.alias_symbols.iter().any(|(id, _)| *id == local));
    assert!(
        semantic
            .node2sym
            .iter()
            .any(|(node, id)| node.sourcefile_id == merged_file
                && *id == local
                && node.start > data.decl.as_ref().unwrap().end),
        "Missing local value reference"
    );
}

#[test]
fn test_side_effect_imports_resolve_to_target_modules() {
    let mut buffer = Vec::new();
    let project = load_project(
        &get_fixtures_dir().join("semantic-project/side-effect"),
        &mut buffer,
    );
    let source = include_str!("fixtures/semantic-project/index.ts");
    let file = source_file_id(&project, "/index.ts");
    for (specifier, target) in [
        ("'./script.js'", "/script.ts"),
        ("'./module.js'", "/module.ts"),
    ] {
        let start = source.find(specifier).unwrap();
        let node = NodeReference {
            sourcefile_id: file,
            start: source[..start].trim_end().encode_utf16().count() as u32,
            end: source[..start + specifier.len()].encode_utf16().count() as u32,
        };
        let target_module = project
            .semantic
            .node2sym
            .iter()
            .find(|(location, _)| same_location(location, &node))
            .map(|(_, symbol)| {
                symbol_data(&project.semantic, *symbol)
                    .decl
                    .as_ref()
                    .expect("Missing module declaration")
                    .sourcefile_id
            })
            .or_else(|| {
                project
                    .semantic
                    .node2module
                    .iter()
                    .find(|(location, _)| same_location(location, &node))
                    .map(|(_, module)| *module)
            });
        assert_eq!(
            target_module,
            Some(source_file_id(&project, target)),
            "Specifier {specifier}"
        );
    }
}

#[test]
fn test_utf16_positions_and_utf8_symbol_names() {
    let source = "let a = `💀`;\nlet b = 1;\nconst f = () => 1;";
    let mut buffer = Vec::new();
    let (_directory, project) = project_from_source(source, &mut buffer);
    let file = source_file_id(&project, "/index.ts");
    let byte_offset = source.find("b =").unwrap();
    let utf16_offset = source[..byte_offset].encode_utf16().count() as u32;
    assert_ne!(byte_offset as u32, utf16_offset);
    let node = NodeReference {
        sourcefile_id: file,
        start: utf16_offset - 1,
        end: utf16_offset + 1,
    };
    let symbol = node_symbol(&project.semantic, &node);
    let data = symbol_data(&project.semantic, symbol);
    assert_eq!(data.name, b"b");
    let declaration = data.decl.as_ref().unwrap();
    assert_eq!(declaration.sourcefile_id, file);
    assert_eq!(declaration.start, node.start);
    let declaration_end = source.find(";\nconst f").unwrap();
    assert_eq!(
        declaration.end,
        source[..declaration_end].encode_utf16().count() as u32
    );
    assert_eq!(
        node_type(&project.semantic, file, node.start, node.end).id,
        project.semantic.primtypes.number
    );
    for (_, symbol) in &project.semantic.symtab {
        std::str::from_utf8(&symbol.name).expect("Symbol names must have valid UTF-8");
    }
}

#[test]
fn test_external_symbol_names_and_aliases() {
    let mut buffer = Vec::new();
    let project = load_project(&get_fixtures_dir().join("external-symbols"), &mut buffer);
    let semantic = &project.semantic;
    for name in [
        "globalThis",
        "Math",
        "Math.abs",
        "Object.prototype.hasOwnProperty",
        "console.log",
    ] {
        external_symbol(semantic, "global", name);
    }
    assert_eq!(
        external_symbol(semantic, "global", "A.abs"),
        external_symbol(semantic, "global", "Math.abs")
    );
    let paths: [(&str, &[&str]); 5] = [
        ("example-dependency", &["api", "other"]),
        ("example-dependency/index.js", &["api", "other"]),
        ("example-reexport", &["renamed"]),
        ("@scope/pkg", &["api"]),
        ("@scope/pkg/subpath", &["api"]),
    ];
    for suffix in ["", ".run", ".nested", ".nested.value", ".self"] {
        let target = external_symbol(semantic, "example-dependency", &format!("api{suffix}"));
        for (namespace, names) in paths {
            for name in names {
                assert_eq!(
                    external_symbol(semantic, namespace, &format!("{name}{suffix}")),
                    target
                );
            }
        }
    }
    external_symbol(semantic, "example-dependency", "value");
    external_symbol(semantic, "node:assert", "ok");
    external_symbol(semantic, "react/jsx-runtime", "jsx");
    assert!(
        !semantic
            .external_symbols
            .iter()
            .any(|symbol| symbol.namespace == b"./local"
                || symbol.name == b"hidden"
                || symbol.name == b"OnlyType")
    );
    let mut unique = std::collections::HashSet::new();
    for symbol in &semantic.external_symbols {
        assert!(
            unique.insert((symbol.symbol_id, &symbol.namespace, &symbol.name)),
            "Duplicate external symbol {symbol:?}"
        );
    }
}

#[test]
fn test_external_jsx_intrinsic_element_symbols() {
    for (fixture, namespace, name) in [
        ("external-symbols", "react", "JSX.IntrinsicElements.div"),
        (
            "external-symbols/global-jsx",
            "global",
            "React.JSX.IntrinsicElements.div",
        ),
    ] {
        let mut buffer = Vec::new();
        let project = load_project(&get_fixtures_dir().join(fixture), &mut buffer);
        external_symbol(&project.semantic, namespace, name);
    }
}
