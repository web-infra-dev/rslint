//! TypeScript syntax and global-type rules enabled by the frozen Rsbuild config.
//! Global references use checker declarations and the baseline's lexical scopes.

use crate::{Violation, edits::Edit};
use tsrs_ast::{
    Kind, Node, NodeFlags, SourceFile, get_source_file_of_node, is_class_like,
    is_function_like_declaration, is_function_like_or_class_static_block_declaration,
};
use tsrs_checker::Checker;
use tsrs_core::P;
use tsrs_scanner::{get_token_pos_of_node, skip_trivia};

pub(crate) const NAMES: &[&str] = &[
    "@typescript-eslint/no-unnecessary-type-constraint",
    "@typescript-eslint/no-wrapper-object-types",
    "@typescript-eslint/prefer-as-const",
    "@typescript-eslint/prefer-namespace-keyword",
    "@typescript-eslint/no-unsafe-function-type",
];

pub(crate) fn check(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Vec<Violation> {
    let report = match name {
        "@typescript-eslint/no-unnecessary-type-constraint" => unnecessary_constraint(node, file),
        "@typescript-eslint/prefer-as-const" => prefer_as_const(node, file),
        "@typescript-eslint/prefer-namespace-keyword" => prefer_namespace(node, file),
        "@typescript-eslint/no-wrapper-object-types"
        | "@typescript-eslint/no-unsafe-function-type" => banned_type(name, node, file, checker),
        _ => None,
    };
    report.into_iter().collect()
}

fn end(node: P<Node>) -> usize {
    node.end().expect_source("parsed node has source end") as usize
}

fn trimmed_range(node: P<Node>, text: &str) -> Option<[usize; 2]> {
    let start = skip_trivia(text, node.pos().source_offset()?) as usize;
    let end = node.end().source_offset()? as usize;
    (start <= end && end <= text.len()).then_some([start, end])
}

fn unnecessary_constraint(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    if node.kind() != Kind::TypeParameter || node.flags().intersects(NodeFlags::Reparsed) {
        return None;
    }
    let parent = node.parent()?;
    if matches!(
        parent.kind(),
        Kind::InferType | Kind::MappedType | Kind::JSDocTemplateTag
    ) {
        return None;
    }
    let parameter = node.as_type_parameter_declaration();
    let constraint = parameter.constraint()?;
    let constraint_name = match constraint.kind() {
        Kind::AnyKeyword => "any",
        Kind::UnknownKeyword => "unknown",
        _ => return None,
    };
    let name = parameter.name();
    let report = Violation::new(node, "unnecessaryConstraint", format!(
        "Constraining the generic type `{}` to `{constraint_name}` does nothing and is unnecessary.", name.text()
    )).with_byte_range([get_token_pos_of_node(node, file, false) as usize, end(node)]);
    if end(name) > end(constraint) {
        return Some(report);
    }
    let disambiguate = parent.kind() == Kind::ArrowFunction
        && parameter.default_type().is_none()
        && parent.type_parameters().len() == 1
        && [".cts", ".mts", ".tsx"]
            .iter()
            .any(|extension| file.file_name().ends_with(extension));
    let next = skip_trivia(file.text(), end(node) as u32) as usize;
    let replacement = if disambiguate && file.text().as_bytes().get(next) != Some(&b',') {
        ","
    } else {
        ""
    };
    Some(report.with_suggestion(
        "removeUnnecessaryConstraint",
        format!("Remove the unnecessary `{constraint_name}` constraint."),
        vec![Edit::replace([end(name), end(constraint)], replacement)],
    ))
}

fn comparable_literal(node: P<Node>) -> bool {
    (node.kind() >= Kind::FirstLiteralToken && node.kind() <= Kind::LastLiteralToken)
        || matches!(node.kind(), Kind::TrueKeyword | Kind::FalseKeyword)
}

fn prefer_as_const(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    let (value, type_node, can_fix) = match node.kind() {
        Kind::VariableDeclaration => {
            let declaration = node.as_variable_declaration();
            (declaration.initializer()?, declaration.type_()?, false)
        }
        Kind::PropertyDeclaration => {
            let declaration = node.as_property_declaration();
            (declaration.initializer()?, declaration.type_()?, false)
        }
        Kind::AsExpression => {
            let expression = node.as_as_expression();
            (expression.expression(), expression.type_(), true)
        }
        Kind::TypeAssertionExpression => {
            let expression = node.as_type_assertion();
            (expression.expression(), expression.type_(), true)
        }
        _ => return None,
    };
    if !comparable_literal(value) || type_node.kind() != Kind::LiteralType {
        return None;
    }
    let literal = type_node.as_literal_type_node().literal();
    if !comparable_literal(literal) || literal.kind() == Kind::NoSubstitutionTemplateLiteral {
        return None;
    }
    let text = file.text();
    let value_range = trimmed_range(value, text)?;
    let literal_range = trimmed_range(literal, text)?;
    if text[value_range[0]..value_range[1]] != text[literal_range[0]..literal_range[1]] {
        return None;
    }
    if can_fix {
        return Some(
            Violation::new(
                literal,
                "preferConstAssertion",
                "Expected a `const` instead of a literal type assertion.",
            )
            .with_byte_range(literal_range)
            .with_fix(vec![Edit::replace(literal_range, "const")]),
        );
    }
    let report = Violation::new(
        literal,
        "variableConstAssertion",
        "Expected a `const` assertion instead of a literal type annotation.",
    )
    .with_byte_range(literal_range);
    let Some(colon) = type_node
        .pos()
        .source_offset()
        .and_then(|pos| pos.checked_sub(1))
    else {
        return Some(report);
    };
    if text.as_bytes().get(colon as usize) != Some(&b':') || end(type_node) > text.len() {
        return Some(report);
    }
    Some(report.with_suggestion(
        "variableSuggest",
        "You should use `as const` instead of type annotation.",
        vec![
            Edit::remove([colon as usize, end(type_node)]),
            Edit::replace([end(value), end(value)], " as const"),
        ],
    ))
}

fn prefer_namespace(node: P<Node>, file: P<SourceFile>) -> Option<Violation> {
    if node.kind() != Kind::ModuleDeclaration {
        return None;
    }
    let declaration = node.as_module_declaration();
    if declaration.name().kind() == Kind::StringLiteral
        || declaration.keyword() != Kind::ModuleKeyword
    {
        return None;
    }
    let text = file.text();
    let declaration_start = skip_trivia(text, node.pos().source_offset()?) as usize;
    let keyword_start =
        if let Some(modifiers) = node.modifiers().filter(|list| !list.nodes().is_empty()) {
            skip_trivia(text, modifiers.end().source_offset()?) as usize
        } else {
            declaration_start
        };
    let keyword_end = keyword_start.checked_add("module".len())?;
    // Dotted namespace segments inherit the keyword but own no keyword token.
    if text.get(keyword_start..keyword_end) != Some("module") {
        return None;
    }
    Some(
        Violation::new(
            node,
            "useNamespace",
            "Use 'namespace' instead of 'module' to declare custom TypeScript modules.",
        )
        .with_byte_range([declaration_start, end(node)])
        .with_fix(vec![Edit::replace(
            [keyword_start, keyword_end],
            "namespace",
        )]),
    )
}

fn banned_type(
    name: &str,
    node: P<Node>,
    file: P<SourceFile>,
    checker: &mut Checker,
) -> Option<Violation> {
    if node.kind() != Kind::TypeReference {
        return None;
    }
    let identifier = node.as_type_reference_node().type_name();
    if identifier.kind() != Kind::Identifier {
        return None;
    }
    let type_name = identifier.text();
    let unsafe_function = name == "@typescript-eslint/no-unsafe-function-type";
    let preferred = match type_name {
        "Function" if unsafe_function => "",
        "BigInt" if !unsafe_function => "bigint",
        "Boolean" if !unsafe_function => "boolean",
        "Number" if !unsafe_function => "number",
        "Object" if !unsafe_function => "object",
        "String" if !unsafe_function => "string",
        "Symbol" if !unsafe_function => "symbol",
        _ => return None,
    };
    if !is_global_reference(identifier, file, checker) {
        return None;
    }
    if unsafe_function {
        return Some(Violation::new(
            identifier,
            "bannedFunctionType",
            "The `Function` type accepts any function-like value.\nPrefer explicitly defining any function parameters and return type.",
        ));
    }
    let report = Violation::new(
        identifier,
        "bannedClassType",
        format!(
            "Prefer using the primitive `{preferred}` as a type name, rather than the upper-cased `{type_name}`."
        ),
    );
    let heritage = node.parent().is_some_and(|parent| {
        parent.kind() == Kind::HeritageClause
            && parent.parent().is_some_and(|grand| {
                (grand.kind() == Kind::InterfaceDeclaration
                    && parent.as_heritage_clause().token() == Kind::ExtendsKeyword)
                    || (is_class_like(grand)
                        && parent.as_heritage_clause().token() == Kind::ImplementsKeyword)
            })
    });
    Some(if heritage {
        report
    } else {
        report.with_fix(vec![Edit::replace(
            trimmed_range(identifier, file.text())?,
            preferred,
        )])
    })
}

fn is_global_reference(identifier: P<Node>, file: P<SourceFile>, checker: &mut Checker) -> bool {
    if checker
        .get_symbol_at_location_exported(identifier)
        .is_some_and(|symbol| {
            symbol
                .declarations()
                .iter()
                .any(|&declaration| get_source_file_of_node(declaration) == Some(file))
        })
    {
        return false;
    }
    // Checker lookup alone misses script-scope collisions with lib declarations,
    // and value-only declarations shadow types in ESLint's scope model.
    !lexically_shadowed(identifier, identifier.text())
}

fn named(node: P<Node>, name: &str) -> bool {
    node.name().is_some_and(|identifier| {
        identifier.kind() == Kind::Identifier && identifier.text() == name
    })
}

fn binding_has_name(node: P<Node>, name: &str) -> bool {
    match node.kind() {
        Kind::Identifier => node.text() == name,
        Kind::ObjectBindingPattern | Kind::ArrayBindingPattern => {
            node.elements().iter().any(|&element| {
                element.kind() == Kind::BindingElement
                    && element
                        .as_binding_element()
                        .name()
                        .is_some_and(|binding| binding_has_name(binding, name))
            })
        }
        _ => false,
    }
}

fn declarations_have_name(list: P<Node>, name: &str) -> bool {
    list.kind() == Kind::VariableDeclarationList
        && list
            .as_variable_declaration_list()
            .declarations()
            .nodes()
            .iter()
            .any(|&declaration| {
                binding_has_name(declaration.as_variable_declaration().name(), name)
            })
}

fn statement_declares(statement: P<Node>, name: &str, imports: bool, types: bool) -> bool {
    match statement.kind() {
        Kind::VariableStatement => {
            declarations_have_name(statement.as_variable_statement().declaration_list(), name)
        }
        Kind::FunctionDeclaration | Kind::ClassDeclaration | Kind::EnumDeclaration => {
            named(statement, name)
        }
        Kind::TypeAliasDeclaration | Kind::InterfaceDeclaration => types && named(statement, name),
        Kind::ModuleDeclaration => {
            named(statement, name)
                && !statement
                    .parent()
                    .is_some_and(|parent| parent.kind() == Kind::ModuleDeclaration)
                && !statement
                    .body()
                    .is_some_and(|body| body.kind() == Kind::ModuleDeclaration)
        }
        Kind::ImportEqualsDeclaration => imports && named(statement, name),
        Kind::ImportDeclaration if imports => statement
            .as_import_declaration()
            .import_clause()
            .is_some_and(|clause| {
                if named(clause, name) {
                    return true;
                }
                clause
                    .as_import_clause()
                    .named_bindings()
                    .is_some_and(|bindings| match bindings.kind() {
                        Kind::NamespaceImport => named(bindings, name),
                        Kind::NamedImports => bindings
                            .as_named_imports()
                            .elements()
                            .nodes()
                            .iter()
                            .any(|&specifier| named(specifier, name)),
                        _ => false,
                    })
            }),
        _ => false,
    }
}

fn hoisted_var(node: P<Node>, name: &str) -> bool {
    let list = match node.kind() {
        Kind::VariableStatement => Some(node.as_variable_statement().declaration_list()),
        Kind::ForStatement => node.as_for_statement().initializer(),
        Kind::ForInStatement | Kind::ForOfStatement => {
            Some(node.as_for_in_or_of_statement().initializer())
        }
        Kind::ModuleDeclaration => return false,
        _ if is_function_like_or_class_static_block_declaration(node) => return false,
        _ => None,
    };
    if list.is_some_and(|list| {
        list.kind() == Kind::VariableDeclarationList
            && !list.flags().intersects(NodeFlags::BlockScoped)
            && declarations_have_name(list, name)
    }) {
        return true;
    }
    if node.kind() == Kind::VariableStatement {
        return false;
    }
    node.for_each_child(&mut |child| hoisted_var(child, name))
}

fn hoisted_function(statement: P<Node>, name: &str) -> bool {
    let child = match statement.kind() {
        Kind::FunctionDeclaration => return named(statement, name),
        Kind::IfStatement => {
            let branch = statement.as_if_statement();
            return hoisted_function(branch.then_statement(), name)
                || branch
                    .else_statement()
                    .is_some_and(|other| hoisted_function(other, name));
        }
        Kind::LabeledStatement => statement.as_labeled_statement().statement(),
        Kind::WhileStatement => statement.as_while_statement().statement(),
        Kind::DoStatement => statement.as_do_statement().statement(),
        Kind::ForStatement => {
            let statement = statement.as_for_statement();
            if statement.initializer().is_some_and(block_scoped_list) {
                return false;
            }
            statement.statement()
        }
        Kind::ForInStatement | Kind::ForOfStatement => {
            let statement = statement.as_for_in_or_of_statement();
            if block_scoped_list(statement.initializer()) {
                return false;
            }
            statement.statement()
        }
        _ => return false,
    };
    hoisted_function(child, name)
}

fn block_scoped_list(node: P<Node>) -> bool {
    node.kind() == Kind::VariableDeclarationList && node.flags().intersects(NodeFlags::BlockScoped)
}

fn statements_shadow(container: P<Node>, name: &str, imports: bool, types: bool) -> bool {
    container.statements().iter().any(|&statement| {
        statement_declares(statement, name, imports, types) || hoisted_function(statement, name)
    })
}

pub(crate) fn lexically_shadowed(identifier: P<Node>, name: &str) -> bool {
    let mut child = identifier;
    let mut in_parameter_decorator = false;
    let mut crossed_scope = false;
    while let Some(scope) = child.parent() {
        if scope.kind() == Kind::Parameter {
            in_parameter_decorator = child.kind() == Kind::Decorator;
        }
        match scope.kind() {
            Kind::SourceFile => {
                return statements_shadow(scope, name, true, true) || hoisted_var(scope, name);
            }
            Kind::Block => {
                if statements_shadow(scope, name, false, true) {
                    return true;
                }
            }
            Kind::ModuleBlock => {
                if statements_shadow(scope, name, true, true) || hoisted_var(scope, name) {
                    return true;
                }
            }
            Kind::CaseBlock => {
                if scope
                    .as_case_block()
                    .clauses()
                    .nodes()
                    .iter()
                    .any(|&clause| statements_shadow(clause, name, false, false))
                {
                    return true;
                }
            }
            Kind::ClassStaticBlockDeclaration => {
                // Static blocks own a direct body field, not Node::body_data.
                // Their `var` declarations hoist within this block only.
                if hoisted_var(scope.as_class_static_block_declaration().body(), name) {
                    return true;
                }
            }
            Kind::CatchClause => {
                if scope
                    .as_catch_clause()
                    .variable_declaration()
                    .is_some_and(|declaration| {
                        binding_has_name(declaration.as_variable_declaration().name(), name)
                    })
                {
                    return true;
                }
            }
            Kind::ClassDeclaration | Kind::ClassExpression => {
                let escapes = scope.kind() == Kind::ClassExpression
                    && child.kind() == Kind::Decorator
                    && crossed_scope;
                if !escapes && named(scope, name) {
                    return true;
                }
            }
            Kind::EnumDeclaration => {
                if named(scope, name) {
                    return true;
                }
            }
            Kind::ForStatement => {
                if scope
                    .as_for_statement()
                    .initializer()
                    .is_some_and(|list| declarations_have_name(list, name))
                {
                    return true;
                }
            }
            Kind::ForInStatement | Kind::ForOfStatement => {
                if declarations_have_name(scope.as_for_in_or_of_statement().initializer(), name) {
                    return true;
                }
            }
            _ if is_function_like_declaration(scope) => {
                let direct_parameter = scope.parameters().contains(&child);
                let escapes = child.kind() == Kind::Decorator
                    || scope.name() == Some(child)
                    || (in_parameter_decorator && crossed_scope && direct_parameter);
                if !escapes {
                    if scope.parameters().iter().any(|parameter| {
                        parameter
                            .name()
                            .is_some_and(|binding| binding_has_name(binding, name))
                    }) {
                        return true;
                    }
                    if matches!(
                        scope.kind(),
                        Kind::FunctionDeclaration | Kind::FunctionExpression
                    ) && named(scope, name)
                    {
                        return true;
                    }
                    if !direct_parameter && scope.body().is_some_and(|body| hoisted_var(body, name))
                    {
                        return true;
                    }
                }
            }
            _ => {}
        }
        if is_function_like_declaration(scope) || is_class_like(scope) {
            crossed_scope = true;
        }
        child = scope;
    }
    false
}
