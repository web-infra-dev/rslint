#!/usr/bin/env python3
# cspell:words GOPATH GOROOT Gopkg buildinf nonignored tinygo uninspected ustar
"""Freeze the external oracle catalog and audit removal of Go from this tree.

Only Python and Git are required. The catalog reads committed baseline objects,
never dirty compiler files or a generated Go executable. The audit is a source
and artifact absence check, not evidence of rule or entrypoint compatibility.
"""

import argparse
import collections
import gzip
import hashlib
import io
import json
from pathlib import Path, PurePosixPath
import re
import subprocess
import sys
import tempfile
import tarfile
import zipfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
BASELINE = "e0a29f751d4def86d27931759098f27748fa3003"
CATALOG = HERE / "rust-migration-catalog.json"
REFERENCE_REVIEW = "scripts/tsrs/rust-migration-reference-review.json"


# Bounded metadata reader from inventory.py, kept here so the final tool has
# no dependency on the old Go-oriented migration tooling.
MODULE = "github.com/web-infra-dev/rslint/"
TOKEN = re.compile(
    r'//[^\n]*|/\*[\s\S]*?\*/|"(?:\\.|[^"\\])*"|`[^`]*`|'
    r"'(?:\\.|[^'\\])*'|[A-Za-z_]\w*|\d+|[^\s]"
)


def tokens(text):
    return [
        (m.group(), m.start())
        for m in TOKEN.finditer(text)
        if not m.group().startswith(("//", "/*"))
    ]


def balanced_end(ts, start):
    pairs = {"(": ")", "{": "}", "[": "]"}
    stack = []
    for i in range(start, len(ts)):
        value = ts[i][0]
        if value in pairs:
            stack.append(pairs[value])
        elif value in pairs.values():
            assert stack and stack.pop() == value
            if not stack:
                return i + 1
    raise ValueError("unbalanced expression")


def split(ts, delimiter=","):
    result, start, depth = [], 0, 0
    for i, (value, _) in enumerate(ts):
        if value in "({[":
            depth += 1
        elif value in ")}]":
            depth -= 1
        elif value == delimiter and depth == 0:
            result.append(ts[start:i])
            start = i + 1
    if ts[start:]:
        result.append(ts[start:])
    return result


class Inventory:
    def __init__(self, repo, revision):
        self.revision = subprocess.check_output(
            ["git", "-C", str(repo), "rev-parse", revision + "^{commit}"], text=True
        ).strip()
        archive = subprocess.check_output(
            ["git", "-C", str(repo), "archive", "--format=tar", self.revision]
        )
        self.sources = {}
        with tarfile.open(fileobj=io.BytesIO(archive)) as tree:
            for item in tree:
                if item.isfile() and Path(item.name).suffix in {
                    ".go", ".ts", ".mts", ".rs", ".json", ".toml", ".yml", ".md", ".mjs"
                }:
                    self.sources[item.name] = tree.extractfile(item).read().decode(
                        "utf-8", errors="replace"
                    )
        self.lexed = {}
        self.used_sources = set()
        self.package_index = collections.defaultdict(list)
        for path in sorted(self.sources):
            if path.endswith(".go") and not path.endswith("_test.go"):
                self.package_index[str(PurePosixPath(path).parent)].append(path)

    def lex(self, path):
        self.used_sources.add(path)
        if path not in self.lexed:
            self.lexed[path] = tokens(self.sources[path])
        return self.lexed[path]

    def evidence(self, path, offset):
        self.used_sources.add(path)
        return f"{path}:{self.sources[path].count(chr(10), 0, offset) + 1}"

    def imports(self, path):
        ts = self.lex(path)
        imports = {}
        for i, (value, _) in enumerate(ts):
            if value.startswith('"' + MODULE):
                target = json.loads(value)[len(MODULE):]
                preceding = ts[i - 1][0]
                alias = preceding if re.fullmatch(r"\w+", preceding) else target.rsplit("/", 1)[-1]
                if alias == "import":
                    alias = target.rsplit("/", 1)[-1]
                imports[alias] = target
        return imports

    def package_files(self, package):
        return self.package_index[package]

    def declaration(self, package, name, kind):
        for path in self.package_files(package):
            ts = self.lex(path)
            for i in range(len(ts) - 2):
                if ts[i][0] != kind or ts[i + 1][0] != name:
                    continue
                if kind == "var":
                    eq = i + 2
                    while ts[eq][0] != "=":
                        eq += 1
                    return path, ts[eq + 1:], self.evidence(path, ts[i][1])
                params_start = i + 2
                assert ts[params_start][0] == "("
                params_end = balanced_end(ts, params_start)
                params = [part[0][0] for part in split(ts[params_start + 1:params_end - 1])]
                body_start = params_end
                while ts[body_start][0] != "{":
                    body_start += 1
                body_end = balanced_end(ts, body_start)
                body = ts[body_start + 1:body_end - 1]
                depth = 0
                for j, (value, _) in enumerate(body):
                    if value == "return" and depth == 0:
                        return path, params, body[j + 1:], self.evidence(path, ts[i][1])
                    if value in "({[":
                        depth += 1
                    elif value in ")}]":
                        depth -= 1
                raise ValueError(f"no top-level return: {path} {name}")
        raise ValueError(f"declaration missing: {package} {name} {kind}")

    def literal_fields(self, ts):
        start = next(i for i, token in enumerate(ts) if token[0] == "{")
        end = balanced_end(ts, start)
        fields = {}
        for part in split(ts[start + 1:end - 1]):
            if len(part) > 2 and part[1][0] == ":":
                fields[part[0][0]] = part[2:]
        return fields

    def scalar(self, ts, env, path):
        if len(ts) == 1:
            value = ts[0][0]
            if value.startswith('"'):
                return json.loads(value)
            if value in {"true", "false"}:
                return value == "true"
            if value in env:
                return env[value]
            package = str(PurePosixPath(path).parent)
            for candidate in self.package_files(package):
                match = re.search(
                    r"\b" + re.escape(value) + r'\s*=\s*("(?:\\.|[^"\\])*"|true|false)',
                    self.sources[candidate],
                )
                if match:
                    self.used_sources.add(candidate)
                    return json.loads(match.group(1))
        if len(ts) == 3 and ts[1][0] == "." and ts[0][0] in env:
            return env[ts[0][0]].get(ts[2][0], False if ts[2][0] == "RequiresTypeInfo" else None)
        # Configs contain callbacks and non-metadata expressions. Preserve them
        # as opaque values; asking one to supply a required name/flag fails.
        return {"unresolved": "".join(t[0] for t in ts)}

    def expression(self, path, ts, env, provenance):
        head = "".join(t[0] for t in ts[:3])
        if head == "rule.Rule":
            fields = self.literal_fields(ts)
            name = self.scalar(fields["Name"], env, path)
            type_info = self.scalar(fields["RequiresTypeInfo"], env, path) if "RequiresTypeInfo" in fields else False
            if not isinstance(name, str) or not isinstance(type_info, bool):
                raise ValueError(f"unsupported metadata: {path}: {name}, {type_info}")
            return {
                "name": name,
                "requires_type_info": type_info,
                "metadata_evidence": provenance + [self.evidence(path, ts[0][1])],
            }
        call_start = next(i for i, token in enumerate(ts) if token[0] == "(")
        callable_name = "".join(t[0] for t in ts[:call_start])
        call_end = balanced_end(ts, call_start)
        arguments = split(ts[call_start + 1:call_end - 1])
        if callable_name == "rule.CreateRule":
            wrapper_source = self.sources["internal/rule/rule.go"]
            assert 'Name:             "@typescript-eslint/" + r.Name' in wrapper_source
            result = self.expression(path, arguments[0], env, provenance)
            result["name"] = "@typescript-eslint/" + result["name"]
            result["metadata_evidence"].append(self.evidence(
                "internal/rule/rule.go", wrapper_source.index("func CreateRule")
            ))
            self.used_sources.add("internal/rule/rule.go")
            return result
        values = []
        for argument in arguments:
            # A constructor's config literal, as opposed to a function call.
            if any(t[0] == "{" for t in argument[:4]):
                fields = self.literal_fields(argument)
                values.append({key: self.scalar(value, env, path) for key, value in fields.items()})
            else:
                values.append(self.scalar(argument, env, path))
        if "." in callable_name:
            alias, name = callable_name.split(".", 1)
            package = self.imports(path)[alias]
        else:
            package, name = str(PurePosixPath(path).parent), callable_name
        factory_path, params, returned, evidence = self.declaration(package, name, "func")
        assert len(params) == len(values), (callable_name, params, values)
        return self.expression(factory_path, returned, dict(zip(params, values)), provenance + [evidence])


FEATURES = [
    ("compiler_program", "Parsing, binding, checking, cross-file module resolution and root/file ownership",
     ["internal/program", "internal/program/loader", "internal/utils"],
     ["internal/utils/create_program.go", "internal/program/program.go", "internal/program/typecheck.go"],
     ["go test ./internal/program ./internal/program/loader ./internal/utils"]),
    ("rule_engine", "Listener traversal, type grants, globals, reference indexes, comments and disable directives",
     ["internal/linter", "internal/rule", "internal/rule_tester"],
     ["internal/rule/context.go", "internal/linter/linter.go", "internal/rule_tester/rule_tester.go"],
     ["go test ./internal/linter ./internal/rule ./internal/rule_tester"]),
    ("catalog_compatibility", "All registered rules, compiler compatibility, heritage and nullable checker grants",
     ["internal/rules", "internal/config"],
     ["internal/rules/all.go", "internal/rules/compiler_compatibility_test.go", "internal/config/all_rules_test.go"],
     ["go test ./internal/rules -run 'Test(AllContainsEveryGoRuleExactlyOnce|AllRulesCompilerCompatibility|HeritageMemberSemantics)$'",
      "go test ./internal/config -run 'Test(AllRules_|GapFile_)'"]),
    ("configuration", "JS/TS config activation, options/schema, discovery, ignores, globals, per-file settings and plugin declarations",
     ["internal/config", "packages/rslint/tests"],
     ["internal/config/plugin_declarations.go", "packages/rslint/src/config/config-loader.ts"],
     ["go test ./internal/config",
      "CI=true pnpm --dir packages/rslint exec rs test run tests/config-loader.test.ts tests/rule-metadata.test.ts"]),
    ("native_cli", "Default CLI, discovery/target workload, reports, exit codes, type-check and fix flags",
     ["cmd/rslint"],
     ["cmd/rslint/main.go", "cmd/rslint/lint_command.go", "cmd/rslint/lint_generation.go"],
     ["go test ./cmd/rslint"]),
    ("js_cli_service", "Node CLI launcher, programmatic lint and persistent service",
     ["packages/rslint/tests"],
     ["packages/rslint/src/index.ts", "packages/rslint/src/cli/cli.ts", "packages/rslint/src/service/service.ts"],
     ["CI=true pnpm --dir packages/rslint exec rs test run tests/args.test.ts tests/engine.test.ts tests/service.test.ts tests/run-cli.test.ts tests/path-identity.test.ts"]),
    ("cli_integration", "Native/Node executable behavior, config inheritance, ignore selection, syntax/type diagnostics, fix and virtual projects",
     ["packages/rslint-test-tools/tests/cli"],
     ["packages/rslint-test-tools/rstack.config.mts", "packages/rslint-test-tools/tests/cli/basic.test.ts"],
     ["CI=true pnpm --dir packages/rslint-test-tools exec rs test run tests/cli/basic.test.ts tests/cli/spawn-cli.test.ts tests/cli/file-args.test.ts tests/cli/disable-comments.test.ts",
      "Select the additional exact tests/cli paths listed in this row for config, plugin, fix and type-check behavior."]),
    ("api_protocol_ast", "Versioned API server, AST/type transport, diagnostics projection, in-memory files and JavaScript AST consumers",
     ["internal/api", "internal/api/server"],
     ["internal/api/api.go", "internal/api/server/handler.go", "internal/api/server/ast_info.go", "packages/rslint-api/src/index.ts"],
     ["go test ./internal/api ./internal/api/server",
      "pnpm --filter @rslint/api build"]),
    ("third_party_plugins", "Object-form ESLint plugins, worker reverse IPC, AST/scope bridge and cancellation",
     ["internal/linter", "internal/api/server", "internal/lsp", "packages/rslint/tests", "packages/rslint-test-tools/tests/cli/js-config/eslint-plugin-conformance"],
     ["packages/rslint/src/eslint-plugin/lint-worker.ts", "internal/linter/pipeline_plugin.go", "internal/lsp/eslint_plugin.go"],
     ["go test ./internal/linter ./internal/api/server ./internal/lsp -run 'ESLint|Plugin'",
      "CI=true pnpm --dir packages/rslint exec rs test run tests/plugin-source.test.ts"]),
    ("fixes", "Lazy fixes, suggestions, iterative fixing, conflict ordering, idempotence and fixed output",
     ["internal/linter", "internal/rule", "internal/api/server", "internal/lsp"],
     ["internal/api/server/lint_fix_test.go", "internal/lsp/service_fixall_test.go", "cmd/rslint/lint_commit.go"],
     ["go test ./internal/linter ./internal/rule ./internal/api/server ./internal/lsp -run 'Fix|Suggestion'"]),
    ("lsp", "Incremental editor overlay, project/config watches, diagnostics, code actions and fix-all",
     ["internal/lsp"],
     ["cmd/rslint/lsp.go", "internal/lsp/server.go", "internal/lsp/document_sync.go", "internal/lsp/code_action.go"],
     ["go test ./internal/lsp"]),
    ("vscode_extension", "Editor extension launch, configuration/diagnostics, multi-root and migration install",
     ["packages/vscode-extension"],
     ["packages/vscode-extension/package.json"],
     ["Inspect packages/vscode-extension/package.json and .github/workflows/ci.yml for platform-specific test launchers."]),
    ("wasm", "Browser worker and virtual filesystem API entrypoint",
     ["packages/rslint-wasm"],
     ["packages/rslint-wasm/src/index.ts", "packages/rslint-wasm/src/worker.ts", "packages/rslint-wasm/scripts/build.mjs"],
     ["pnpm --filter @rslint/wasm build"]),
    ("rust_client_native", "One-shot compiler project/config commands, semantic binary exports and Rust/N-API client",
     ["cmd/tsgo", "crates/tsgo-client/tests", "crates/rslint-native"],
     ["cmd/tsgo/main.go", "crates/tsgo-client/src/lib.rs", "crates/rslint-native/src/lib.rs"],
     ["go test ./cmd/tsgo", "cargo test -p tsgo-client"]),
]


def feature_matrix(inv):
    result = []
    for name, description, test_roots, sources, commands in FEATURES:
        for source in sources:
            if source not in inv.sources:
                raise ValueError("feature source missing: " + source)
            inv.used_sources.add(source)
        tests = sorted(
            p for p in inv.sources
            if any(p.startswith(root + "/") for root in test_roots)
            and "/testdata/" not in p and "/fixtures/" not in p
            and (p.endswith("_test.go") or p.endswith(".test.ts")
                 or (p.startswith("crates/") and p.endswith("_test.rs")))
        )
        result.append(dict(
            id=name, description=description, sources=sources, tests=tests,
            suggested_commands=commands, migration_status="pending",
            compatibility_status="not_run",
            caveat="Test paths inventory coverage locations; passing parity evidence is still required.",
        ))
    return result



def sha(data):
    return hashlib.sha256(data).hexdigest()


def write_json(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2, ensure_ascii=False) + "\n")


def git(repo, *args):
    return subprocess.check_output(["git", "-C", str(repo), *args])


def make_catalog(repo, revision):
    if repo.resolve() == ROOT.resolve() or ROOT.resolve() in repo.resolve().parents:
        raise ValueError("--baseline must point outside the candidate repository")

    class Frozen(Inventory):
        def scalar(self, ts, env, path):
            value = super().scalar(ts, env, path)
            if isinstance(value, dict) and "unresolved" in value:
                return dict(value, schema_source=path, schema_tokens=ts)
            return value

        def schema(self, ts, path, depth=0, env=None):
            if depth > 15:
                raise ValueError("recursive schema: " + path)
            env = env or {}
            expression = "".join(t[0] for t in ts)
            package = str(PurePosixPath(path).parent)
            argument = None
            if len(ts) == 3 and ts[1][0] == "." and ts[0][0] in env:
                argument = env[ts[0][0]][ts[2][0]]
            elif len(ts) == 1 and ts[0][0] in env:
                argument = env[ts[0][0]]
            if argument is not None:
                if not isinstance(argument, dict) or "schema_tokens" not in argument:
                    raise ValueError("unsupported schema argument: " + path)
                return self.schema(argument["schema_tokens"], argument["schema_source"], depth + 1)
            # Every accepted expression is resolved to its committed schema.
            if expression.startswith("rule.NewSchema(") or expression.startswith("NewSchema("):
                start = next(i for i, item in enumerate(ts) if item[0] == "(")
                end = balanced_end(ts, start)
                return self.schema(ts[start + 1:end - 1], path, depth + 1, env)
            if expression.startswith("[]byte("):
                start = next(i for i, item in enumerate(ts) if item[0] == "(")
                end = balanced_end(ts, start)
                literal = ts[start + 1:end - 1]
                if len(literal) != 1:
                    raise ValueError("unsupported schema bytes: " + path)
                raw = literal[0][0]
                text = raw[1:-1] if raw.startswith("`") else json.loads(raw)
                self.used_sources.add(path)
                return json.loads(text), [self.evidence(path, literal[0][1])]
            parts = expression.split(".")
            if len(parts) == 3 and parts[-1] == "Schema":
                package = self.imports(path)[parts[0]]
                source, initializer, evidence = self.declaration(package, parts[1], "var")
                value = self.expression(source, initializer, {}, [evidence])
                return value["options_schema"], value["options_schema_evidence"]
            if len(parts) == 2:
                package = self.imports(path)[parts[0]]
                name = parts[1]
            elif len(parts) == 1 and re.fullmatch(r"\w+", expression):
                name = expression
            else:
                raise ValueError("unsupported schema expression in " + path + ": " + expression)
            for candidate in self.package_files(package):
                source = self.sources[candidate]
                match = re.search(r"//go:embed\s+(\S+)\s+(?:var\s+)?" + re.escape(name) + r"\s+\[\]byte", source)
                if match:
                    schema_path = str(PurePosixPath(candidate).parent / match.group(1))
                    self.used_sources.update([candidate, schema_path])
                    return json.loads(self.sources[schema_path]), [self.evidence(candidate, match.start()), schema_path]
            source, initializer, evidence = self.declaration(package, name, "var")
            # declaration returns the source tail, so retain just this initializer.
            if any(item[0] == "(" for item in initializer[:5]):
                start = next(i for i, item in enumerate(initializer) if item[0] == "(")
                initializer = initializer[:balanced_end(initializer, start)]
            value, sources = self.schema(initializer, source, depth + 1)
            return value, [evidence, *sources]

        def expression(self, path, ts, env, provenance):
            result = super().expression(path, ts, env, provenance)
            if "".join(t[0] for t in ts[:3]) == "rule.Rule":
                fields = self.literal_fields(ts)
                if "Schema" not in fields:
                    raise ValueError("registered rule lacks schema: " + path)
                result["options_schema"], result["options_schema_evidence"] = self.schema(fields["Schema"], path, env=env)
            return result

        def rules(self):
            # The earlier inventory filtered variable names ending in Rule and
            # missed NoWebpackLoaderSyntax. Count every qualified registration.
            root = "internal/rules/all.go"
            imports = self.imports(root)
            aliases = re.findall(r"implementedRules = append\(implementedRules, (\w+)\.GetAllRules\(\)\.\.\.\)", self.sources[root])
            if not aliases or "coreRules()..." not in self.sources[root]:
                raise ValueError("unsupported root catalog")
            catalogs = [root] + [imports[alias] + "/all.go" for alias in aliases]
            rows = []
            groups = {"core": "eslint", "typescript": "typescript-eslint", "react_hooks": "eslint-plugin-react-hooks", "jsx_a11y": "eslint-plugin-jsx-a11y", "rstest": "rstest"}
            js_config = "packages/rslint-test-tools/rstack.config.mts"
            self.used_sources.add(js_config)
            configured = {"packages/rslint-test-tools/" + value[3:-1] for value, _ in tokens(self.sources[js_config]) if value.startswith(("'./tests/", '"./tests/')) and value.endswith((".test.ts'", '.test.ts"'))}
            for catalog in catalogs:
                catalog_imports = self.imports(catalog)
                namespace = "core" if catalog == root else catalog.split("/")[2]
                registrations = re.findall(r"^\s*(\w+)\.(\w+),\s*$", self.sources[catalog], re.M)
                if not registrations:
                    raise ValueError("empty catalog: " + catalog)
                for alias, variable in registrations:
                    package = catalog_imports[alias]
                    source, initializer, evidence = self.declaration(package, variable, "var")
                    row = self.expression(source, initializer, {}, [evidence])
                    group = groups.get(namespace, "eslint-plugin-" + namespace.replace("_", "-"))
                    slug = row["name"].split("/", 1)[-1]
                    js_prefix = f"packages/rslint-test-tools/tests/{group}/rules/{slug}"
                    js_tests = sorted(p for p in self.sources if p == js_prefix + ".test.ts" or (p.startswith(js_prefix + "/") and p.endswith(".test.ts")))
                    shared = {str(PurePosixPath(e.rsplit(":", 1)[0]).parent) for e in row["metadata_evidence"]} - {package, "internal/rule"}
                    code = "\n".join(self.sources[p] for p in self.package_files(package))
                    row.update(namespace=namespace, variable=variable, package=package, source=source, catalog=catalog,
                               registration_evidence=self.evidence(catalog, self.sources[catalog].index(f"{alias}.{variable},")),
                               go_tests=sorted(p for p in self.sources if str(PurePosixPath(p).parent) == package and p.endswith("_test.go")),
                               js_integration_tests=js_tests,
                               configured_js_integration_tests=[p for p in js_tests if p in configured],
                               constructor_tests=sorted(p for p in self.sources if p.endswith("_test.go") and str(PurePosixPath(p).parent) in shared),
                               compiler_imports=sorted(set(re.findall(r'"(github.com/microsoft/TypeScript/tsc/[^"]+)"', code))),
                               directly_mentions_type_checker=bool(re.search(r"\b(?:TypeChecker|checker\.)", code)),
                               migration_status="pending_rust_implementation", compatibility_status="not_verified", rust_sources=[], verification=[])
                    rows.append(row)
            if len({row["name"] for row in rows}) != len(rows):
                raise ValueError("duplicate catalog registration")
            return sorted(rows, key=lambda row: row["name"])

    inv = Frozen(repo, revision)
    rows = inv.rules()
    surfaces = feature_matrix(inv)
    for surface in surfaces:
        surface["oracle_commands"] = surface.pop("suggested_commands")
        surface.update(migration_status="pending_rust_js_implementation", compatibility_status="not_verified", rust_js_sources=[], verification=[])
    references = set(inv.used_sources)
    for row in rows:
        references.update(row["go_tests"] + row["js_integration_tests"] + row["constructor_tests"])
        references.update(inv.package_files(row["package"]))
    for surface in surfaces:
        references.update(surface["tests"])
    hashes = {path: sha(inv.sources[path].encode()) for path in sorted(references)}
    if inv.revision == BASELINE and (len(rows), len(surfaces)) != (813, 14):
        raise ValueError("frozen baseline catalog changed unexpectedly")
    return {
        "schema_version": 1,
        "baseline": {"repository": "https://github.com/web-infra-dev/rslint.git", "commit": inv.revision, "input": "external Git checkout; committed objects only", "tree": git(repo, "rev-parse", inv.revision + "^{tree}").decode().strip()},
        "generator": {"path": "scripts/tsrs/rust-migration-inventory.py", "sha256": sha(Path(__file__).read_bytes()), "baseline_metadata_reader": "embedded bounded source interpreter"},
        "summary": {"registered_rules": len(rows), "declared_type_required_rules": sum(row["requires_type_info"] for row in rows), "surfaces": len(surfaces), "verified_migrated_rules": 0, "verified_surfaces": 0, "namespace_counts": dict(sorted(collections.Counter(row["namespace"] for row in rows).items()))},
        "corrections": [{"previous_count": 812, "verified_count": 813, "previously_omitted": "import/no-webpack-loader-syntax", "reason": "The older inventory matched only registration variable names ending in Rule; NoWebpackLoaderSyntax also belongs to the root catalog."}],
        "scope": ["Every builtin registration and its option schema is resolved from committed baseline source without executing Go.", "Go test paths and commands refer only to the external frozen oracle. They are migration input, never a candidate runtime dependency.", "Schemas include schema-declared defaults. Defaults and behavior implemented in rule bodies remain covered by source hashes and oracle test references.", "No test, port, behavior compatibility or performance result is claimed by this inventory."],
        "rules": rows, "surfaces": surfaces, "baseline_source_sha256": hashes,
    }


MANIFEST = re.compile(r"^(?:go(?:\.[\w-]+)?\.(?:mod|sum|work)(?:\.sum)?|Gopkg\.(?:toml|lock))$")
BUILD_REF = re.compile(r'''actions/setup-go|(?:^|[^\w])(?:GOOS|GOARCH|GOROOT|GOPATH|GOWORK|GOFLAGS|CGO_ENABLED)\b|golangci|tinygo|wasm_exec\.(?:js|ts)|\bgo\s+(?:build|run|test|install|generate|mod|work|env|version|tool)\b|\bgo(?:\.[\w-]+)?\.(?:mod|sum|work)\b|github\.com/(?:microsoft/(?:TypeScript|typescript-go)|golang)/|(?:(?:spawn|exec|Command(?:::new)?)\([^\n]*["'](?:go|tsgo)["'])''')
CODE_EXTENSIONS = {".py", ".sh", ".bash", ".zsh", ".rs", ".js", ".mjs", ".cjs", ".ts", ".mts", ".cts", ".tsx", ".jsx", ".toml", ".json", ".yml", ".yaml", ".make", ".mk"}
GO_MAGIC = b"\xff Go buildinf:"


def inspect_data(path, data, origin, depth=0):
    findings = []
    name = PurePosixPath(path).name
    suffix = PurePosixPath(path).suffix.lower()
    def add(kind, **extra):
        findings.append(dict(kind=kind, path=path, origin=origin, **extra))
    if depth > 4:
        add("uninspected_nested_archive")
        return findings
    if suffix == ".go":
        add("go_source")
    if MANIFEST.fullmatch(name):
        add("go_manifest")
    if GO_MAGIC in data or (b"\0" in data and b"runtime.morestack" in data and b"runtime.main" in data):
        add("precompiled_go_artifact", sha256=sha(data))
    if data.startswith(b"\x1f\x8b"):
        try:
            # Scan compressed release products too; do not trust their suffix.
            with gzip.GzipFile(fileobj=io.BytesIO(data)) as stream:
                expanded = stream.read(512 * 1024 * 1024 + 1)
            if len(expanded) > 512 * 1024 * 1024:
                add("uninspected_compressed_artifact", reason="expanded artifact exceeds 512 MiB scan limit")
            elif GO_MAGIC in expanded or (b"\0" in expanded and b"runtime.morestack" in expanded and b"runtime.main" in expanded):
                add("compressed_go_artifact", sha256=sha(data))
            else:
                findings.extend(inspect_data(path + "!gzip", expanded, origin, depth + 1))
        except (OSError, EOFError):
            add("invalid_compressed_artifact")
    if data.startswith(b"PK\x03\x04") or data[257:262] == b"ustar":
        try:
            if data.startswith(b"PK\x03\x04"):
                archive = zipfile.ZipFile(io.BytesIO(data))
                entries = [(item.filename, item.file_size, not item.is_dir(), lambda item=item: archive.read(item)) for item in archive.infolist()]
            else:
                archive = tarfile.open(fileobj=io.BytesIO(data), mode="r:")
                entries = [(item.name, item.size, item.isfile(), lambda item=item: archive.extractfile(item).read()) for item in archive.getmembers()]
            with archive:
                if sum(size for _, size, _, _ in entries) > 512 * 1024 * 1024:
                    add("uninspected_compressed_artifact", reason="expanded archive exceeds 512 MiB scan limit")
                else:
                    for name, _, regular, read in entries:
                        if regular:
                            findings.extend(inspect_data(path + "!/" + name, read(), origin, depth + 1))
        except (OSError, EOFError, ValueError, RuntimeError, zipfile.BadZipFile, tarfile.TarError):
            add("invalid_compressed_artifact")
    if b"\0" in data:
        return findings
    try:
        text = data.decode("utf-8")
    except UnicodeDecodeError:
        return findings
    if suffix not in {".go", ".md", ".mdx", ".json"} and re.search(r"(?m)^package\s+[A-Za-z_]\w*\s*$", text) and re.search(r"(?m)^(?:func\s+(?:\(|[A-Za-z_])|import\s*\()", text):
        add("go_source_disguised_extension")
    if suffix in CODE_EXTENSIONS or name in {"Makefile", "Dockerfile"}:
        for number, line in enumerate(text.splitlines(), 1):
            for match in BUILD_REF.finditer(line):
                add("go_build_or_runtime_reference", line=number,
                    line_sha256=sha(line.encode("utf-8")), token=match.group())
    if suffix in {".patch", ".diff"} and re.search(r"(?m)^\+\+\+ (?:b/)?[^\n]+\.go(?:\s|$)", text):
        add("go_source_patch")
    return findings


def reference_review(repo):
    """Load reviewed *lines*, never executable-path or URL exclusions.

    Both the complete UTF-8 line (excluding its line terminator) and the whole
    file are pinned. A moved/edited line or any file edit invalidates its review.
    Metadata is itself scanned as candidate source. No review applies to release
    artifacts, source/manifest signatures, archives, or submodules.
    """
    path = repo / REFERENCE_REVIEW
    if not path.exists():
        return {}, None
    data = path.read_bytes()
    document = json.loads(data)
    if document.get("schema_version") != 1 or not isinstance(document.get("files"), list):
        raise ValueError("invalid reference review schema")
    result = {}
    categories = {"source_attribution", "fixture_text", "negative_assertion",
                  "external_oracle_metadata", "detector_test_or_pattern"}
    for entry in document["files"]:
        name = entry.get("path")
        pure = PurePosixPath(name) if isinstance(name, str) else None
        if (not pure or pure.is_absolute() or ".." in pure.parts
                or str(pure) != name or name in result):
            raise ValueError("invalid or duplicate reviewed path")
        if not re.fullmatch(r"[0-9a-f]{64}", entry.get("sha256", "")):
            raise ValueError("invalid reviewed file digest: " + name)
        lines = {}
        for row in entry.get("lines", []):
            numbers = row.get("line_numbers", [])
            if (not isinstance(row.get("reason"), str) or not row["reason"].strip()
                    or row.get("category") not in categories
                    or not re.fullmatch(r"[0-9a-f]{64}", row.get("sha256", ""))
                    or not numbers):
                raise ValueError("invalid reviewed line: " + name)
            for number in numbers:
                if type(number) is not int or number <= 0 or number in lines:
                    raise ValueError("invalid or duplicate reviewed line number: " + name)
                lines[number] = row
        if not lines:
            raise ValueError("empty reviewed path: " + name)
        result[name] = dict(sha256=entry["sha256"], lines=lines)
    return result, sha(data)


def audit(repo, artifact_roots):
    findings, checked, errors, reviewed = [], [], [], []
    review, review_sha256 = reference_review(repo)
    seen_reviews = set()
    entries = git(repo, "ls-files", "--cached", "--others", "--exclude-standard", "-z").decode().split("\0")
    stages = git(repo, "ls-files", "--stage", "-z").decode().split("\0")
    submodules = [item.split("\t", 1)[1] for item in stages if item.startswith("160000 ")]
    for path in submodules:
        findings.append(dict(kind="unresolved_submodule", path=path, origin="candidate", reason="Submodule source is outside the candidate file list; remove it or explicitly inspect and replace this gate before acceptance."))
    for path in sorted(set(filter(None, entries))):
        full = repo / path
        if path in submodules:
            continue
        if full.is_symlink():
            findings.append(dict(kind="uninspected_symlink", path=path, origin="candidate"))
            continue
        if not full.exists():
            errors.append(dict(path=path, reason="listed file is missing; stage intended deletions before the final gate"))
            continue
        if not full.is_file():
            errors.append(dict(path=path, reason="not a regular file"))
            continue
        try:
            data = full.read_bytes()
        except OSError:
            errors.append(dict(path=path, reason="file is unreadable"))
            continue
        checked.append(dict(path=path, sha256=sha(data)))
        approved = review.get(path)
        if approved:
            seen_reviews.add(path)
            actual_lines = data.decode("utf-8").splitlines()
            valid = approved["sha256"] == sha(data)
            for number, row in approved["lines"].items():
                valid = (valid and number <= len(actual_lines)
                         and sha(actual_lines[number - 1].encode("utf-8")) == row["sha256"]
                         and bool(BUILD_REF.search(actual_lines[number - 1])))
            if not valid:
                errors.append(dict(path=path, reason="reviewed reference file or line changed; explicit re-review required"))
                approved = None
        for finding in inspect_data(path, data, "candidate"):
            line = approved["lines"].get(finding.get("line")) if approved else None
            if (finding["kind"] == "go_build_or_runtime_reference" and line
                    and line["sha256"] == finding["line_sha256"]):
                reviewed.append(dict(**finding, category=line["category"], reason=line["reason"],
                                     file_sha256=approved["sha256"]))
            else:
                findings.append(finding)
    for path in sorted(set(review) - seen_reviews):
        errors.append(dict(path=path, reason="reviewed reference file is absent from inspected source"))
    products = []
    for index, root in enumerate(artifact_roots):
        label = "artifact-" + str(index + 1)
        paths = sorted(root.rglob("*")) if root.is_dir() else [root]
        if not root.exists():
            errors.append(dict(path=label, reason="artifact root is missing"))
        for path in paths:
            relative = label + ("/" + str(path.relative_to(root)) if root.is_dir() else "/" + root.name)
            if path.is_symlink():
                errors.append(dict(path=relative, reason="artifact symlink requires explicit target scan"))
            elif path.is_file():
                try:
                    data = path.read_bytes()
                except OSError:
                    errors.append(dict(path=relative, reason="artifact is unreadable"))
                    continue
                products.append(dict(path=relative, sha256=sha(data), size=len(data)))
                findings.extend(inspect_data(relative, data, "release_artifact"))
    if artifact_roots and not products:
        errors.append(dict(path="release_artifacts", reason="empty product selection"))
    counts = dict(sorted(collections.Counter(row["kind"] for row in findings).items()))
    return dict(schema_version=2, scope="tracked and nonignored untracked working-tree files plus explicitly selected release artifacts; Git history and ignored build caches are not candidate source", source_files=len(checked), source_manifest_sha256=sha(json.dumps(checked, sort_keys=True).encode()), artifact_files=products, counts=counts, findings=findings, errors=errors, reviewed_references=reviewed, reference_review_sha256=review_sha256, zero_go_gate_passed=not findings and not errors and bool(products), limitations=["This gate cannot establish semantic compatibility or absence of dynamically constructed downloads and commands. Runtime process evidence and clean Rust/Node-only builds remain separate required gates.", "No product absence verdict is granted without explicit nonempty --artifact-root selection.", "Reviewed non-executable reference lines are reported separately with exact path, line number, line and file digests, category and reason. No whole-path, test-directory or URL exclusion applies."])


def self_test():
    def kinds(path, data):
        return {row["kind"] for row in inspect_data(path, data, "self_test")}
    assert "go_source" in kinds("src/old.go", b"package old\n")
    assert "go_source_disguised_extension" in kinds("old.go.txt", b"package old\nfunc run() {}\n")
    assert "go_manifest" in kinds("go.oracle.work.sum", b"")
    assert "go_build_or_runtime_reference" in kinds("package.json", b'{"build":"go build ./cmd"}')
    assert "go_build_or_runtime_reference" in kinds("build.rs", b'Command::new("go");')
    assert "precompiled_go_artifact" in kinds("bin/rslint", b"\0" + GO_MAGIC + b"test")
    assert "compressed_go_artifact" in kinds("browser.wasm.gz", gzip.compress(b"\0" + GO_MAGIC))
    zipped = io.BytesIO()
    with zipfile.ZipFile(zipped, "w", compression=zipfile.ZIP_DEFLATED) as archive:
        archive.writestr("bin/hidden", b"\0" + GO_MAGIC)
    assert "precompiled_go_artifact" in kinds("extension.vsix", zipped.getvalue())
    assert "precompiled_go_artifact" not in kinds("audit.py", b'patterns = ["runtime.main", "runtime.morestack"]')
    assert not kinds("src/main.rs", b"fn main() {}\n")
    with tempfile.TemporaryDirectory(prefix="rslint-rust-audit-") as temp:
        root = Path(temp)
        git(root, "init", "--quiet")
        (root / "main.rs").write_text("fn main() {}\n")
        product = root / "product.node"
        product.write_bytes(b"\0compiled-rust-fixture")
        assert audit(root, [product])["zero_go_gate_passed"]
        assert not audit(root, [])["zero_go_gate_passed"]
        (root / "untracked.go").write_text("package old\n")
        result = audit(root, [product])
        assert not result["zero_go_gate_passed"]
        assert result["counts"]["go_source"] == 1
        git(root, "add", "untracked.go")
        assert audit(root, [product])["counts"]["go_source"] == 1
        (root / "untracked.go").unlink()
        assert not audit(root, [product])["zero_go_gate_passed"]
        git(root, "rm", "--cached", "--quiet", "untracked.go")
        assert audit(root, [product])["zero_go_gate_passed"]
        # Exact reviewed prose is admitted, but not a changed line, an appended
        # invocation in the same file, a copied path, or matching product bytes.
        fixture = root / "fixture.test.ts"
        source = b'const text = "Now go build something great.";\n'
        fixture.write_bytes(source)
        review_path = root / REFERENCE_REVIEW
        review_path.parent.mkdir(parents=True)
        write_json(review_path, dict(schema_version=1, files=[dict(
            path="fixture.test.ts", sha256=sha(source), lines=[dict(
                line_numbers=[1], sha256=sha(source.rstrip(b"\n")),
                category="fixture_text", reason="A literal fixture sentence.")])]))
        result = audit(root, [product])
        assert result["zero_go_gate_passed"]
        assert len(result["reviewed_references"]) == 1
        assert result["reviewed_references"][0]["line"] == 1
        fixture.write_bytes(source.replace(b"great", b"else"))
        result = audit(root, [product])
        assert not result["zero_go_gate_passed"] and result["errors"]
        assert result["counts"]["go_build_or_runtime_reference"] == 1
        fixture.write_bytes(source + b'child_process.execSync("go build ./cmd");\n')
        result = audit(root, [product])
        assert not result["zero_go_gate_passed"]
        assert result["counts"]["go_build_or_runtime_reference"] == 2
        # Updating only the file digest does not approve any newly added line.
        document = json.loads(review_path.read_text())
        document["files"][0]["sha256"] = sha(fixture.read_bytes())
        write_json(review_path, document)
        result = audit(root, [product])
        assert not result["zero_go_gate_passed"] and not result["errors"]
        assert result["counts"]["go_build_or_runtime_reference"] == 1
        document["files"][0]["sha256"] = sha(source)
        write_json(review_path, document)
        fixture.write_bytes(source)
        copy = root / "other.test.ts"
        copy.write_bytes(source)
        assert audit(root, [product])["counts"]["go_build_or_runtime_reference"] == 1
        copy.unlink()
        assert audit(root, [fixture])["counts"]["go_build_or_runtime_reference"] == 1
        # Source signatures cannot be approved even with matching review hashes.
        fixture.write_bytes(b"package hidden\nfunc run() {}\n" + source)
        assert "go_source_disguised_extension" in audit(root, [product])["counts"]
        fixture.write_bytes(source)
        (root / "broken-link").symlink_to("missing-target")
        assert not audit(root, [product])["zero_go_gate_passed"]
    print("Rust migration inventory self-test passed")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--baseline", type=Path, help="external frozen oracle Git checkout; required for catalog generation")
    parser.add_argument("--revision", default=BASELINE)
    parser.add_argument("--catalog", type=Path, default=CATALOG)
    parser.add_argument("--check-catalog", action="store_true", help="compare deterministic regeneration without writing")
    parser.add_argument("--repo", type=Path, default=ROOT)
    parser.add_argument("--output", type=Path, default=ROOT / "target/rust-js-migration/audit.json")
    parser.add_argument("--artifact-root", type=Path, action="append", default=[])
    parser.add_argument("--gate", action="store_true", help="exit nonzero unless source and explicit release artifact scans contain zero Go blockers")
    parser.add_argument("--self-test", action="store_true", help="run bounded detector and fail-closed gate checks without a build")
    args = parser.parse_args()
    if args.self_test:
        self_test()
        return 0
    if args.baseline:
        catalog = make_catalog(args.baseline.resolve(), args.revision)
        if args.check_catalog:
            if json.loads(args.catalog.read_text()) != catalog:
                raise ValueError("catalog differs from committed external oracle metadata")
        else:
            write_json(args.catalog, catalog)
        print(json.dumps(catalog["summary"]))
    elif args.check_catalog:
        parser.error("--check-catalog requires --baseline")
    result = audit(args.repo.resolve(), [path.resolve() for path in args.artifact_root])
    write_json(args.output, result)
    print(json.dumps({"counts": result["counts"], "source_files": result["source_files"], "artifact_files": len(result["artifact_files"]), "zero_go_gate_passed": result["zero_go_gate_passed"]}))
    return 0 if not args.gate or result["zero_go_gate_passed"] else 1


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (ValueError, KeyError, AssertionError, OSError, subprocess.CalledProcessError) as error:
        print("Rust migration inventory failed: " + str(error), file=sys.stderr)
        sys.exit(2)
