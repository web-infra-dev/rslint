#!/usr/bin/env python3
"""Pinned Rsbuild benchmark: original tsgo CLI versus the direct Rust/JS CLI.

smoke keeps every failure and makes no performance claim. proof additionally
observes the ordinary CLI's native calls and the complete public API. measure
requires that proof plus explicit, hash-bound correctness/build review evidence.
No Go service, trace binary, partial rule selection or input rewrite is used.
The sampler and sleep/clock guards are shared with benchmark.py.
"""

import argparse
import copy
import importlib.util
import os
from pathlib import Path
import platform
import re
import shutil
import subprocess
import sys
import time

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import benchmark as bench

SIDES = ("baseline", "candidate")
CASES = ("full", "engine")
PROJECT_PIN = "e7d65898ce92d5f5d78afb0d93b4670b5e2a469d"
BASELINE_PIN = "e0a29f751d4def86d27931759098f27748fa3003"
require, read, save = bench.require, bench.read_json, bench.save_json
sha, canonical, digest = bench.file_digest, bench.canonical, bench.digest_bytes


def git(root, *argv):
    return subprocess.check_output(["git", "-C", str(root), *argv])


def source_files(root):
    names = git(root, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
    return {name.decode(): sha(root / name.decode()) for name in names.split(b"\0")
            if name and (root / name.decode()).is_file()}


def tree_files(root):
    result, seen = {}, {}
    for directory, dirs, files in os.walk(root, followlinks=True):
        logical = Path(directory).relative_to(root).as_posix()
        real = str(Path(directory).resolve())
        if real in seen:
            result[logical + "/@alias"] = seen[real]
            dirs[:] = []
            continue
        seen[real] = logical
        dirs[:] = sorted(name for name in dirs if name not in (".git", ".cache", ".pnpm-workspace-state-v1"))
        for name in sorted(files):
            file = Path(directory) / name
            if file.is_file():
                result[file.relative_to(root).as_posix()] = sha(file)
    return result


def snapshot(args):
    inputs = source_files(args.project)
    # Ignored workspace outputs are part of the checker/import input too.
    for name in ("packages", "scripts"):
        for directory, dirs, files in os.walk(args.project / name):
            dirs[:] = sorted(item for item in dirs if item not in ("node_modules", ".git", ".cache"))
            for name in sorted(files):
                file = Path(directory) / name
                if file.is_file():
                    inputs[file.relative_to(args.project).as_posix()] = sha(file)
    dependencies = tree_files(args.project / "node_modules")
    require(inputs and dependencies, "project sources/dependencies must not be empty")
    result = {"schema_version": 1, "project_commit": git(args.project, "rev-parse", "HEAD").decode().strip(),
              "project_inputs": inputs, "project_dependencies": dependencies,
              "workload_manifest_sha256": sha(args.manifest), "node_sha256": sha(args.node),
              "support": {name: sha(HERE / name) for name in (
                  "benchmark.py", "rsbuild-backend.mjs", "rust-rsbuild-benchmark.py", "rust-rsbuild-observe.mjs", "rust-rsbuild-manifest-check.py")},
              "conditions": conditions(), "backends": {}}
    if args.frozen_evidence:
        result["frozen_evidence_sha256"] = sha(args.frozen_evidence)
    if args.project_selection:
        result["project_selection_sha256"] = sha(args.project_selection)
    for side in SIDES:
        root = getattr(args, side)
        artifacts = {}
        for relative in ("packages/rslint/dist", "packages/rslint/bin"):
            require((root / relative).is_dir(), "unbuilt core: " + str(root / relative))
            artifacts.update({relative + "/" + key: value for key, value in tree_files(root / relative).items()})
        native = args.native[side]
        artifacts[native.relative_to(root).as_posix()] = sha(native)
        result["backends"][side] = {"commit": git(root, "rev-parse", "HEAD").decode().strip(),
                                     "source": source_files(root), "artifacts": artifacts,
                                     "core_dependencies": tree_files(root / "packages/rslint/node_modules")}
    return result


def conditions():
    return {"CI": "true", "NODE_DISABLE_COMPILE_CACHE": "1", "GOMAXPROCS": "1",
            "RAYON_NUM_THREADS": "1", "GOMEMLIMIT": "4GiB"}


def normalize(value, args):
    if isinstance(value, str):
        for root, label in sorted(((args.project, "<project>"), (args.baseline, "<backend>"),
                                   (args.candidate, "<backend>")), key=lambda pair: -len(str(pair[0]))):
            value = value.replace(str(root), label)
        return value
    if isinstance(value, list):
        return [normalize(item, args) for item in value]
    if isinstance(value, dict):
        result = {}
        for key, item in value.items():
            key = normalize(key, args)
            require(key not in result, "path normalization collided")
            result[key] = normalize(item, args)
        return result
    return value


def cli_output(prefix, record, args):
    rows = [bench.parse_json(line) for line in Path(str(prefix) + ".stdout").read_text().splitlines() if line.strip()]
    for row in rows:
        require(isinstance(row, dict) and set(row) == {"ruleName", "message", "filePath", "range", "severity"},
                "unexpected CLI JSONL; no output may be discarded")
    result = {"diagnostics": sorted(normalize(rows, args), key=canonical), "exit_code": record["exit_code"],
              "stderr": normalize(Path(str(prefix) + ".stderr").read_text(), args)}
    save(str(prefix) + ".normalized.json", result)
    return result


def command(args, side, case, observation=None, api=None):
    root = getattr(args, side)
    argv = [str(args.node), "--import", str(HERE / "rsbuild-backend.mjs")]
    env = {**conditions(), "RSLINT_BENCH_CORE": str(root / "packages/rslint")}
    if side == "candidate":
        # Guard is present in measured runs too. It writes nothing unless proof
        # is requested and refuses child-process/compiler fallbacks.
        argv += ["--import", str(HERE / "rust-rsbuild-observe.mjs")]
        env["RSLINT_RUST_GUARD"] = str(args.native[side])
        if observation:
            env["RSLINT_RUST_OBSERVATION"] = str(observation)
    if api:
        argv += [str(HERE / "rust-rsbuild-observe.mjs"), "--api", str(args.config), str(api)]
    elif case == "full":
        argv += [str(args.project / "node_modules/rstack/bin/rs.js"), "lint", "--singleThreaded", "--format", "jsonline"]
    else:
        argv += [str(root / "packages/rslint/bin/rslint.js"), "--singleThreaded", "--config", str(args.config), "--format", "jsonline"]
    return {"argv": argv, "cwd": os.path.relpath(args.project, args.output),
            "inherit_env": ["PATH", "HOME", "TMPDIR", "LANG", "LC_ALL", "SYSTEMROOT"], "env": env,
            "expected_exit_codes": list(range(126)), "timeout_seconds": args.timeout,
            "process_expectations": {"required_executables": ([args.native[side].name] if side == "baseline" else [])}}


def run(args, side, case, label, lock_fd, observation=None, api=None):
    prefix = args.output / label
    print(label, flush=True)
    record = bench.run_once(command(args, side, case, observation, api), args.output, {}, prefix,
                            args.interval, lock_fd)
    record.update(backend=side, case=case)
    save(str(prefix) + ".record.json", record)
    return prefix, record


def smoke(args, identity, lock_fd):
    rows = []
    for case in args.cases:
        for side in args.sides:
            row = {"case": case, "backend": side}
            try:
                prefix, record = run(args, side, case, case + "-" + side, lock_fd)
                row["output"] = cli_output(prefix, record, args)
                row["status"] = "completed" if record["exit_code"] in (0, 1) and not row["output"]["stderr"] else "failed"
            except Exception as error:
                row.update(status="failed", error=str(error))
            rows.append(row)
            save(args.output / "smoke.json", rows)
    require(snapshot(args) == identity, "source/artifact/input changed during smoke")
    require(all(row["status"] == "completed" for row in rows), "smoke retained one or more failures; no performance conclusion")


def disk_text(filename):
    raw = filename.read_bytes()
    # Keep U+FEFF, exactly as the actual candidate's source snapshot does.
    return raw.decode("utf-16-le" if raw.startswith(b"\xff\xfe") else
                      "utf-16-be" if raw.startswith(b"\xfe\xff") else "utf-8")


def validate_disk_overlays(request):
    """The real benchmark cannot replace dependencies/configs with phantom data."""
    seen = set()
    for source in [*request.get("sources", []), *[row for row in request.get("files", []) if "text" in row]]:
        require(isinstance(source, dict) and isinstance(source.get("path"), str) and isinstance(source.get("text"), str),
                "malformed native source overlay")
        filename = Path(source["path"])
        require(filename.is_absolute() and filename.is_file(), "native source overlay is not an existing disk file")
        physical = filename.resolve()
        require(physical not in seen, "native source overlays alias the same physical input")
        seen.add(physical)
        require(source["text"] == disk_text(filename), "native overlay changes frozen disk content: " + str(filename))


def portable_project(value, args):
    if isinstance(value, str):
        return value.replace(str(args.project), "<project>")
    if isinstance(value, list):
        return [portable_project(item, args) for item in value]
    if isinstance(value, dict):
        return {key: portable_project(item, args) for key, item in value.items()}
    return value


def candidate_compiler_revision(args):
    cargo = (args.candidate / "crates/rslint-core/Cargo.toml").read_text()
    pins = dict(re.findall(r'^\s*(tsrs_\w+)\s*=\s*\{\s*git\s*=\s*"https://github\.com/oxc-project/tsrs\.git",\s*rev\s*=\s*"([a-f0-9]{40})"', cargo, re.MULTILINE))
    require({"tsrs_compiler", "tsrs_checker", "tsrs_parser", "tsrs_binder", "tsrs_ast"} <= set(pins)
            and len(set(pins.values())) == 1, "candidate must declare one reproducible tsrs compiler/checker/parser/binder/AST pin")
    revision = next(iter(pins.values()))
    declared = re.search(r'pub const COMPILER_REVISION:\s*&str\s*=\s*"([a-f0-9]{40})"',
                         (args.candidate / "crates/rslint-core/src/lib.rs").read_text())
    require(declared is not None and declared.group(1) == revision, "native compiler metadata declaration differs from Cargo pin")
    return revision


def validate_project_plan(request, actual, args):
    """Bind every actual lint request file to the original selected project.

    planTsrsSync uses the same projects::select operation as lint execution.
    Program construction's own root/options checks remain in production core.
    This comparison establishes selected metadata, not full compiler source-
    universe equivalence. The baseline trace has no reference fields, so those
    candidate fields are reported without claiming an unavailable comparison.
    """
    require(args.project_selection and args.project_selection.is_file(), "proof requires --project-selection")
    oracle = read(args.project_selection)
    require(oracle.get("manifestSHA256") == sha(args.manifest)
            and oracle.get("baselineCommit") == BASELINE_PIN and oracle.get("projectCommit") == PROJECT_PIN,
            "baseline project-selection pins differ")
    require(actual.get("schemaVersion") == 1 and actual.get("backend") == "tsrs-rust", "not a native project plan")
    require(actual.get("compiler") == {"revision": candidate_compiler_revision(args)}, "native planner compiler revision differs from pinned candidate Cargo/source")
    manifest = read(args.manifest)
    require(oracle.get("compilerCommit") == manifest["baseline"]["compilerCommit"], "baseline compiler pin differs")
    by_file, plan_indices = {}, set()
    for plan in oracle.get("plans", []):
        index = plan.get("manifestPlanIndex")
        require(type(index) is int and index not in plan_indices, "duplicate/missing baseline plan index")
        plan_indices.add(index)
        for filename in plan.get("admittedFiles", []):
            require(filename not in by_file, "duplicate baseline selected file")
            by_file[filename] = plan
    expected_files = {row[0]: row for row in manifest["files"]}
    require(set(by_file) == set(expected_files) and plan_indices == set(range(len(manifest["plans"]))),
            "baseline project-selection does not cover the frozen workload")
    files, file_projects, projects = request.get("files", []), actual.get("fileProjects", []), actual.get("projects", [])
    require(len(files) == len(file_projects) == len(expected_files), "actual project plan omits admitted files")
    require(isinstance(projects, list) and projects, "actual project plan has no typed projects")
    checked, source_only, selected_indices, project_pairs = set(), [], set(), set()
    project_records = []
    for source, project_index in zip(files, file_projects):
        relative = os.path.relpath(source["path"], args.project).replace(os.sep, "/")
        require(relative in expected_files and relative not in checked, "actual project plan duplicate/unknown file")
        checked.add(relative)
        row = expected_files[relative]
        expected_plan = by_file[relative]
        require(row[4] == expected_plan["manifestPlanIndex"], "baseline file-to-plan mapping differs")
        expected = expected_plan["project"]
        typed = manifest["configurations"][row[3]]["hasTypeChecker"]
        require((project_index is not None) is typed, "actual selected checker differs: " + relative)
        policy = source.get("policy")
        require(isinstance(policy, dict), "native file omitted project policy")
        kind = policy.get("kind")
        if kind == "service":
            # projectService's configured search boundary can live in an
            # external config module. The production selector deliberately
            # searches to filesystem/node_modules boundaries when that root is
            # not an ancestor. A service lookup can also produce no project.
            require(set(policy) == {"kind", "rootDirectory"}
                    and policy.get("rootDirectory") == str(args.config.parent), "projectService configured root differs from frozen explicit config origin")
        else:
            require(policy == {"kind": "sourceOnly"}, "frozen source/project policy changed")
        if not typed:
            require(project_index is None, "source-only file gained a selected project")
            source_only.append((source, expected_plan))
            continue
        require(type(project_index) is int and 0 <= project_index < len(projects), "invalid selected native project index")
        selected_indices.add(project_index)
        selected = portable_project(projects[project_index], args)
        require(selected.get("configFileName") == expected.get("configFileName"), "selected config differs: " + relative)
        require(selected.get("currentDirectory") == expected.get("currentDirectory"), "selected project cwd differs: " + relative)
        require(sorted(selected.get("rootFiles", [])) == sorted(expected.get("rootFiles", [])), "selected project roots differ: " + relative)
        require(selected.get("compilerOptions") == expected.get("compilerOptions"), "selected compiler options differ: " + relative)
        require(policy.get("kind") == "service" and selected.get("sourceReferences") is True,
                "frozen projectService source-reference behavior changed")
        pair = (project_index, expected_plan["manifestPlanIndex"])
        if pair not in project_pairs:
            project_pairs.add(pair)
            project_records.append({"project_index": project_index, "manifest_plan_index": pair[1],
                                    "metadata_sha256": digest(canonical(selected)),
                                    "config": selected["configFileName"], "roots": len(selected["rootFiles"]),
                                    "project_references": selected.get("projectReferences"),
                                    "source_references": selected.get("sourceReferences"),
                                    "config_parsing_diagnostics": selected.get("configFileParsingDiagnostics")})
    require(checked == set(expected_files) and selected_indices == set(range(len(projects))), "unused/omitted native project or source")
    source_only_cwd = request.get("sourceOnlyCurrentDirectory", request["cwd"])
    require(isinstance(source_only_cwd, str) and Path(source_only_cwd).is_absolute(), "source-only cwd missing")
    source_only_by_plan = {}
    for source, expected_plan in source_only:
        expected = expected_plan["project"]
        require(expected.get("configFileName") == "", "baseline source-only config changed")
        require(portable_project(source_only_cwd, args) == expected.get("currentDirectory"), "source-only host cwd differs")
        require(actual.get("sourceOnlyCompilerOptions") == expected.get("compilerOptions"), "source-only compiler options differ")
        source_only_by_plan.setdefault(expected_plan["manifestPlanIndex"], {"files": [], "expected": expected})["files"].append(source["path"])
    for group in source_only_by_plan.values():
        require(sorted(portable_project(group["files"], args)) == sorted(group["expected"]["rootFiles"]), "source-only roots differ")
    return {"baseline_project_selection_sha256": sha(args.project_selection),
            "actual_plan_sha256": digest(canonical(actual)), "files": len(checked),
            "typed_files": len(checked) - len(source_only), "source_only_files": len(source_only),
            "typed_projects": len(projects), "selected_project_plan_pairs": len(project_pairs),
            "source_only_current_directory": portable_project(source_only_cwd, args),
            "projects": project_records, "matched_common_fields": ["configFileName", "currentDirectory", "rootFiles", "compilerOptions"],
            "reference_comparison": "candidate values recorded; baseline trace contains no projectReferences/sourceReferences fields",
            "program_source_universe_equivalence": "not established by project planning"}


def project_selection_negative_probes(request, actual, args):
    """Prove the verifier rejects plausible wrong projects, options and overlays."""
    first_typed = next(index for index, project in enumerate(actual["fileProjects"]) if project is not None)
    project_index = actual["fileProjects"][first_typed]
    mutated = []
    changed = copy.deepcopy(actual)
    changed["projects"][project_index]["configFileName"] += ".alternate.json"
    mutated.append(("alternate_config_same_roots", request, changed, False))
    changed = copy.deepcopy(actual)
    options = changed["projects"][project_index]["compilerOptions"]
    options["strictNullChecks"] = not options.get("strictNullChecks", False)
    mutated.append(("strict_null_checks_changed", request, changed, False))
    changed = copy.deepcopy(actual)
    changed["compiler"]["revision"] = "0" * 40
    mutated.append(("planner_compiler_revision_changed", request, changed, False))
    changed_request = copy.deepcopy(request)
    changed_request["files"][first_typed]["policy"] = {"kind": "sourceOnly"}
    mutated.append(("typed_file_policy_changed", changed_request, actual, False))
    changed_request = copy.deepcopy(request)
    changed_request["sourceOnlyCurrentDirectory"] = str(args.project / "wrong-source-only-cwd")
    mutated.append(("source_only_cwd_changed", changed_request, actual, False))
    changed_request = copy.deepcopy(request)
    changed_request["sources"] = [*changed_request.get("sources", []), {"path": str(args.output / "phantom-not-on-disk.ts"), "text": "export {}"}]
    mutated.append(("phantom_dependency_overlay", changed_request, actual, True))
    result = []
    for name, bad_request, bad_plan, overlay in mutated:
        try:
            validate_disk_overlays(bad_request) if overlay else validate_project_plan(bad_request, bad_plan, args)
        except bench.BenchmarkError as error:
            result.append({"name": name, "status": "rejected", "error": str(error)})
        else:
            raise bench.BenchmarkError("project-selection verifier accepted negative probe: " + name)
    return result


def native_workload(path, args, expected_fix=False, expected_calls=1):
    """Validate actual normal-entry native calls, never a supplied expected result."""
    rows = [bench.parse_json(line) for line in path.read_text().splitlines() if line.strip()]
    require(rows and all(row.get("event") in ("nativeAddon", "nativeRequest", "nativeProjectPlan", "nativeResponse") for row in rows),
            "native observation missing, errored or attempted a child-process fallback")
    addons = [row for row in rows if row["event"] == "nativeAddon"]
    require(addons and all(Path(row["filename"]) == args.native["candidate"] for row in addons), "unpinned native addon observed")
    calls = {}
    for row in rows:
        if row["event"] == "nativeAddon":
            continue
        key, event = row.get("call"), row["event"]
        require(type(key) is int and key > 0 and event not in calls.setdefault(key, {}), "duplicate/missing native call identity")
        calls[key][event] = row if event == "nativeProjectPlan" else row[event[len("native"):].lower()]
    require(len(calls) == expected_calls, "unexpected native lint call count")
    manifest = read(args.manifest)
    by_file = {row[0]: row for row in manifest["files"]}
    typed = {row["name"]: row["requiresTypeInformation"] for row in manifest["rules"]}
    summaries = []
    for index, pair in sorted(calls.items()):
        require(set(pair) == {"nativeRequest", "nativeProjectPlan", "nativeResponse"}, "incomplete native call")
        request, response = pair["nativeRequest"], pair["nativeResponse"]
        require(request.get("schemaVersion") == response.get("schemaVersion") == 1 and response.get("backend") == "tsrs-rust",
                "not a normal tsrs configured-lint generation")
        fix = expected_fix[index - 1] if isinstance(expected_fix, list) else expected_fix
        require(request.get("fix", False) is fix, "unexpected fix mode")
        require(response.get("summary", {}).get("threads") == 1, "candidate has unexpected rule-worker concurrency")
        if expected_calls == 1:
            require(request.get("singleThreaded") is True, "normal CLI did not forward --singleThreaded")
        require(Path(request.get("cwd", "")) == args.project, "native project cwd differs")
        validate_disk_overlays(request)
        project_plan = pair["nativeProjectPlan"]
        plan_request = {"schemaVersion": request["schemaVersion"], "cwd": request["cwd"],
                        "files": [{"path": item["path"], "policy": item["policy"]} for item in request["files"]],
                        "sources": request.get("sources", []) + [{"path": item["path"], "text": item["text"]}
                            for item in request["files"] if "text" in item]}
        require(project_plan.get("request") == plan_request, "native planner did not receive the actual lint policies/overlays")
        project_selection = validate_project_plan(request, project_plan["plan"], args)
        require(response.get("compiler") == project_plan["plan"].get("compiler")
                == {"revision": candidate_compiler_revision(args)}, "native lint/planner compiler revision differs from pinned Cargo/source")
        inputs, outputs = request.get("files", []), response.get("files", [])
        require(len(inputs) == len(outputs) == len(by_file), "native did not retain every admitted file")
        seen, rules_count, typed_count, typed_files = set(), 0, 0, 0
        for source, output in zip(inputs, outputs):
            source_path = Path(source.get("path", ""))
            require(source_path.is_absolute() and source_path == Path(output.get("filePath", "")), "native input/output file identity differs")
            relative = os.path.relpath(source_path, args.project).replace(os.sep, "/")
            require(relative in by_file and relative not in seen, "extra/duplicate native file: " + relative)
            seen.add(relative)
            row = by_file[relative]
            require(source_path.is_file() and sha(source_path) == row[1] and source_path.stat().st_size == row[2], "admitted bytes differ: " + relative)
            if "text" in source:
                # CLI freezes BOM-decoded source. Preserve exact source meaning;
                # raw disk hashes and byte sizes above still cover the original.
                raw = source_path.read_bytes()
                if raw.startswith(b"\xff\xfe"):
                    text = raw.decode("utf-16-le")
                elif raw.startswith(b"\xfe\xff"):
                    text = raw.decode("utf-16-be")
                else:
                    text = raw.decode("utf-8")
                require(source["text"] == text, "native source overlay differs: " + relative)
            config = manifest["configurations"][row[3]]
            has_types = config["hasTypeChecker"]
            require(output.get("hasTypeInformation") is has_types, "actual type capability differs: " + relative)
            require(not output.get("hasSyntaxErrors", False), "syntax failure skipped lint rules: " + relative)
            expected = sorted([{"name": manifest["rules"][i]["name"],
                                "severity": 2 if manifest["rules"][i]["severityName"] == "error" else 1,
                                "options": manifest["rules"][i]["options"] or []}
                               for i in config["rules"]["all"]], key=lambda item: item["name"])
            configured = source.get("rules", [])
            require(all(item.get("name") in typed for item in configured), "native configured unknown rule")
            active = [item for item in configured if has_types or not typed[item["name"]]]
            require(sorted(active, key=lambda item: item["name"]) == expected, "actual native rule/options/severity differs: " + relative)
            applied = output.get("appliedRules")
            require(isinstance(applied, list) and len(set(applied)) == len(applied) and sorted(applied) == sorted(item["name"] for item in expected),
                    "native skipped or added actual rules: " + relative)
            rules_count += len(applied)
            typed_count += sum(typed[name] for name in applied)
            typed_files += int(has_types)
        require(seen == set(by_file), "native omitted silent admitted files")
        counts = {"files": len(seen), "rule_file_pairs": rules_count, "typed_rule_file_pairs": typed_count, "typed_files": typed_files}
        require(counts == {"files": manifest["counts"]["unique_files"], "rule_file_pairs": manifest["counts"]["rule_passes"],
                           "typed_rule_file_pairs": manifest["counts"]["type_aware_rule_passes"], "typed_files": manifest["counts"]["typed_files"]},
                "native workload totals differ")
        summaries.append({"call": index, "fix": fix, "counts": counts,
                          "request_sha256": digest(canonical(request)), "response_sha256": digest(canonical(response)),
                          "compiler": response.get("compiler"), "summary": response["summary"],
                          "project_selection": project_selection})
    negative_probes = project_selection_negative_probes(calls[min(calls)]["nativeRequest"],
                                                        calls[min(calls)]["nativeProjectPlan"]["plan"], args)
    return {"observation_sha256": sha(path), "calls": summaries, "no_child_processes": True,
            "project_selection_negative_probes": negative_probes}


def api_projection(value):
    rows = []
    for file in value["lint"]:
        for message in file["messages"]:
            rows.append({"filePath": file["filePath"], "ruleName": message["ruleId"], "message": message["message"],
                         "severity": {1: "warning", 2: "error"}[message["severity"]],
                         "range": {"start": {"line": message["line"], "column": message["column"]},
                                   "end": {"line": message["endLine"], "column": message["endColumn"]}}})
    return sorted(rows, key=canonical)


def proof(args, identity, lock_fd):
    require(set(args.sides) == set(SIDES) and set(args.cases) == set(CASES), "proof requires both sides and command scopes")
    answer = {"schema_version": 1, "status": "incomplete", "identity_sha256": digest(canonical(identity)),
              "cases": {}, "checks": {}, "scope": "complete frozen rsbuild workload; supplemental review still mandatory"}
    save(args.output / "proof.json", answer)
    for case in CASES:
        results = {}
        for side in SIDES:
            trace = args.output / (case + "-" + side + ".native.jsonl") if side == "candidate" else None
            prefix, record = run(args, side, case, case + "-" + side, lock_fd, observation=trace)
            results[side] = cli_output(prefix, record, args)
            require(record["exit_code"] in (0, 1), "ordinary CLI failed before normal lint completion")
            if trace:
                answer["cases"].setdefault(case, {})["native_workload"] = native_workload(trace, args)
        require(results["baseline"] == results["candidate"], "complete CLI diagnostics/positions/messages/exit/stderr differ: " + case)
        # Make observation overhead demonstrably output-neutral.
        prefix, record = run(args, "candidate", case, case + "-candidate-unobserved", lock_fd)
        require(cli_output(prefix, record, args) == results["candidate"], "native observation changed candidate behavior")
        answer["cases"][case]["expected"] = results["baseline"]
        save(args.output / "proof.json", answer)
    require(answer["cases"]["full"]["expected"] == answer["cases"]["engine"]["expected"], "full/engine CLI outputs differ")
    observed = {}
    for side in SIDES:
        destination = args.output / ("api-" + side + ".json")
        trace = args.output / "api-candidate.native.jsonl" if side == "candidate" else None
        prefix, record = run(args, side, "engine", "api-" + side, lock_fd, observation=trace, api=destination)
        require(record["exit_code"] == 0 and destination.is_file(), "public API observation failed")
        observed[side] = {"result": normalize(read(destination), args), "exit_code": record["exit_code"],
                          "stdout": normalize(Path(str(prefix) + ".stdout").read_text(), args),
                          "stderr": normalize(Path(str(prefix) + ".stderr").read_text(), args)}
        save(args.output / ("api-" + side + ".normalized.json"), observed[side])
        if trace:
            answer["api_native_workload"] = native_workload(trace, args, expected_fix=[False, True], expected_calls=2)
    require(observed["baseline"] == observed["candidate"], "complete public API lint/fix results differ")
    manifest_files = {row[0] for row in read(args.manifest)["files"]}
    for mode in ("lint", "fixed"):
        files = observed["baseline"]["result"][mode]
        require(len(files) == len(manifest_files) and {row["filePath"] for row in files} == manifest_files,
                "public API omitted/repeated/added admitted files, including silent files")
    require(api_projection(observed["baseline"]["result"]) == answer["cases"]["engine"]["expected"]["diagnostics"], "public API/CLI projection differs")
    require(snapshot(args) == identity, "source/artifact/input changed during proof")
    answer.update(status="pending_supplement", api=observed["baseline"],
                  checks={name: True for name in ("workload", "native_applied_rules", "typed_capabilities", "diagnostics",
                      "positions", "messages", "api_fixes", "api_cli_projection", "observation_equivalence", "candidate_no_child_processes",
                      "selected_projects", "project_selection_negative_probes")})
    save(args.output / "proof.json", answer)


def measure(args, identity, lock_fd):
    require(args.proof and args.supplement, "measure requires --proof and --supplement")
    require(set(args.sides) == set(SIDES) and set(args.cases) == set(CASES), "measure requires both sides and command scopes")
    fingerprint = digest(canonical(identity))
    accepted, supplement = read(args.proof), read(args.supplement)
    require(accepted.get("schema_version") == 1 and accepted.get("status") == "pending_supplement"
            and accepted.get("identity_sha256") == fingerprint, "proof is not accepted for this exact snapshot")
    for name in ("workload", "native_applied_rules", "typed_capabilities", "diagnostics", "positions", "messages", "api_fixes",
                 "api_cli_projection", "observation_equivalence", "candidate_no_child_processes", "selected_projects", "project_selection_negative_probes"):
        require(accepted.get("checks", {}).get(name) is True, "missing proof check: " + name)
    require(supplement.get("schema_version") == 1 and supplement.get("identity_sha256") == fingerprint
            and supplement.get("proof_sha256") == sha(args.proof), "supplement must bind exact proof and input/artifact identity")
    for name in ("cli_fix_disk", "type_aware_fixture", "rule_hit_no_hit", "baseline_workload", "project_type_selection",
                 "compiler_library_differences_reviewed", "no_tsgo_fallback", "source_build_provenance", "optimized_builds", "single_threaded_baseline"):
        require(supplement.get("checks", {}).get(name) is True, "missing supplemental review check: " + name)
    for name in bench.FIXTURE_CHECKS:
        require(supplement.get("fixture_coverage", {}).get(name) is True, "missing fixture coverage: " + name)
    require(isinstance(supplement.get("evidence"), list) and supplement["evidence"], "no supplemental evidence")
    for row in supplement["evidence"]:
        file = args.supplement.parent / row["path"]
        require(not Path(row["path"]).is_absolute() and file.is_file() and sha(file) == row["sha256"], "supplement evidence missing/changed")
    proof_hash, supplement_hash = sha(args.proof), sha(args.supplement)
    records = []
    for case_index, case in enumerate(CASES):
        for phase, count in (("warmup", args.warmups), ("measurement", args.repeats)):
            for repetition in range(count):
                for side in (SIDES if (case_index + repetition) % 2 == 0 else SIDES[::-1]):
                    require(sha(args.proof) == proof_hash and sha(args.supplement) == supplement_hash, "proof/supplement changed")
                    require(snapshot(args) == identity, "snapshot changed before sample")
                    prefix, record = run(args, side, case, f"{case}-{phase}-{repetition + 1:02d}-{side}", lock_fd)
                    record.update(phase=phase, repetition=repetition + 1)
                    records.append(record)
                    save(args.output / "records.json", records)
                    require(cli_output(prefix, record, args) == accepted["cases"][case]["expected"], "sample output/exit/stderr differs from proof")
                    require(snapshot(args) == identity, "snapshot changed during sample")
    save(args.output / "summary.json", {"status": "complete", "identity_sha256": fingerprint,
         "proof_sha256": proof_hash, "supplement_sha256": supplement_hash,
         "memory_method": bench.MEMORY_METHOD, "cases": bench.summarize(records)})


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--stage", choices=("smoke", "proof", "measure"), required=True)
    for name, default in (("project", "../rsbuild-benchmark"), ("baseline", "../rslint-tsgo-baseline"),
                          ("candidate", "."),
                          ("output", "../../artifacts/rust-js-migration/rust-rsbuild-run"),
                          ("manifest", "docs/tsrs-migration/rsbuild-workload.json")):
        parser.add_argument("--" + name, type=Path, default=Path(default))
    parser.add_argument("--lock", type=Path, required=True)
    parser.add_argument("--lock-timeout", type=float, default=120)
    parser.add_argument("--timeout", type=float, default=900)
    parser.add_argument("--interval", type=float, default=0.002)
    parser.add_argument("--warmups", type=int, default=2)
    parser.add_argument("--repeats", type=int, default=7)
    parser.add_argument("--sides", choices=SIDES, nargs="+", default=list(SIDES))
    parser.add_argument("--cases", choices=CASES, nargs="+", default=list(CASES))
    parser.add_argument("--proof", type=Path)
    parser.add_argument("--supplement", type=Path)
    parser.add_argument("--frozen-evidence", type=Path, help="JSON paths to the manifest-pinned original trace/summary/source/record/stdout/stderr")
    parser.add_argument("--project-selection", type=Path, help="Current single-threaded baseline project-selection.json; exact SHA is bound into identity")
    args = parser.parse_args()
    for name in ("project", "baseline", "candidate", "output", "manifest", "lock", "proof", "supplement", "frozen_evidence", "project_selection"):
        if getattr(args, name):
            setattr(args, name, getattr(args, name).resolve())
    require(args.warmups >= 2 and args.repeats >= 7, "at least two warmups and seven interleaved repeats required")
    require(0 < args.interval <= 0.1 and 0 < args.timeout <= 86400 and args.lock_timeout > 0, "invalid timeouts/interval")
    require(not args.output.exists() or not any(args.output.iterdir()), "use a fresh output directory")
    for root in (args.project, args.baseline, args.candidate):
        require(os.path.commonpath((root, args.output)) != str(root), "evidence must be outside project/backend trees")
    args.output.mkdir(parents=True, exist_ok=True)
    require(git(args.project, "rev-parse", "HEAD").decode().strip() == PROJECT_PIN, "project commit differs")
    require(git(args.baseline, "rev-parse", "HEAD").decode().strip() == BASELINE_PIN, "baseline commit differs")
    require(not git(args.baseline, "diff", "--name-only", "HEAD").strip(), "baseline tracked sources are dirty")
    manifest = read(args.manifest)
    require(manifest["project"]["commit"] == PROJECT_PIN and len(manifest["files"]) == 2394 and len(manifest["rules"]) == 129,
            "not the frozen 2394-file/129-rule workload")
    args.node = Path(shutil.which("node")).resolve()
    arch = {"arm64": "arm64", "aarch64": "arm64", "x86_64": "x64"}[platform.machine()]
    tuple_name = ("darwin-" + arch if sys.platform == "darwin" else "linux-" + arch + "-gnu")
    args.native = {}
    for side in SIDES:
        package = "@rslint/native-" + tuple_name + ("/bin" if side == "baseline" else "")
        resolved = subprocess.check_output([str(args.node), "--input-type=module", "-e",
            "import {createRequire} from 'node:module'; import fs from 'node:fs'; "
            "console.log(fs.realpathSync(createRequire(process.argv[1]).resolve(process.argv[2])))",
            str(getattr(args, side) / "packages/rslint/package.json"), package]).decode().strip()
        args.native[side] = Path(resolved)
        require(os.path.commonpath((getattr(args, side), args.native[side])) == str(getattr(args, side)), "native resolves outside backend tree")
    args.config = (args.project / "node_modules/rstack").resolve() / "dist/rslintConfig.js"
    require(args.config.is_file(), "installed Rstack lint config missing")
    spec = importlib.util.spec_from_file_location("machine_lock", HERE / "with-machine-lock.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    waiter = None
    try:
        with args.lock.open("a+") as handle:
            waited, waiter = module.acquire(handle, args.lock_timeout)
            save(args.output / "environment.json", {"platform": platform.uname()._asdict(), "started_unix": time.time(),
                "lock_wait_seconds": waited, "load_before": list(os.getloadavg()), "conditions": conditions(),
                "node_version": subprocess.check_output([str(args.node), "--version"]).decode().strip(),
                "concurrency": "both timed CLIs --singleThreaded; API has no public singleThreaded option and is correctness-only; GOMAXPROCS=RAYON_NUM_THREADS=1",
                "pnpm_launcher": "excluded; exact Node-expanded rs lint workflow included",
                "cache": "Node compile cache disabled; warm OS caches; new process each invocation",
                "memory_method": bench.MEMORY_METHOD})
            (args.output / "process-load-before.txt").write_bytes(subprocess.check_output(["ps", "-Ao", "pid,ppid,pgid,pcpu,pmem,etime,comm"]))
            if sys.platform == "darwin":
                thermal = subprocess.run(["pmset", "-g", "therm"], capture_output=True)
                (args.output / "thermal-before.txt").write_bytes(thermal.stdout + thermal.stderr)
            if args.stage != "smoke":
                require(args.frozen_evidence and args.frozen_evidence.is_file(), "proof/measure require --frozen-evidence")
                require(args.project_selection and args.project_selection.is_file(), "proof/measure require --project-selection")
                evidence_paths = read(args.frozen_evidence)
                require(set(evidence_paths) == {"trace", "summary", "source", "record", "stdout", "stderr"}, "incomplete original evidence map")
                evidence_paths = {key: (args.frozen_evidence.parent / value).resolve() for key, value in evidence_paths.items()}
                verifier_spec = importlib.util.spec_from_file_location("rust_frozen_manifest", HERE / "rust-rsbuild-manifest-check.py")
                verifier = importlib.util.module_from_spec(verifier_spec)
                verifier_spec.loader.exec_module(verifier)
                verified = verifier.verify(args.manifest, args.project, args.baseline, evidence_paths, node=args.node)
                save(args.output / "frozen-manifest-verification.json", verified)
            identity = snapshot(args)
            save(args.output / "identity.json", identity)
            if args.stage == "smoke":
                smoke(args, identity, handle.fileno())
            elif args.stage == "proof":
                proof(args, identity, handle.fileno())
            else:
                measure(args, identity, handle.fileno())
    except BaseException as error:
        save(args.output / "failure.json", {"status": "invalid_or_incomplete", "error": str(error), "no_valid_measurement_summary": True})
        raise
    finally:
        if sys.platform == "darwin":
            thermal = subprocess.run(["pmset", "-g", "therm"], capture_output=True)
            (args.output / "thermal-after.txt").write_bytes(thermal.stdout + thermal.stderr)
        if waiter:
            waiter.stdin.close()
            waiter.wait(timeout=10)
            waiter.stdout.close()


if __name__ == "__main__":
    main()
