#!/usr/bin/env python3
"""Verify the original admission capture and its frozen current input anchors.

This reads files and package metadata only; it never builds or runs lint. The
original capture used its recorded concurrency, which is not automatically a
proof of the new single-threaded Program selection. Nor do its runtime anchors
establish the historical transitive dependency closure. Those remain separate,
explicit checks for the benchmark supplement.
"""

import argparse
import json
from pathlib import Path
import shutil
import subprocess
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent))
import benchmark as bench

PROJECT_PIN = "e7d65898ce92d5f5d78afb0d93b4670b5e2a469d"
BASELINE_PIN = "e0a29f751d4def86d27931759098f27748fa3003"
COUNTS = {"plans": 33, "file_passes": 2394, "unique_files": 2394,
          "rule_passes": 221014, "type_aware_rule_passes": 31461,
          "unique_rule_count": 129, "typed_files": 1232}
EVIDENCE_HASHES = {"trace": "traceSHA256", "summary": "summarySHA256",
                   "source": "sourceRecordSHA256", "record": "commandRecordSHA256"}
require = bench.require


def inside(root, relative):
    require(isinstance(relative, str) and relative and not Path(relative).is_absolute()
            and ".." not in Path(relative).parts, "invalid manifest input path")
    return root / relative


def command(argv):
    return subprocess.check_output(argv, text=True, timeout=60).strip()


def expanded_rule(manifest, index):
    require(type(index) is int and 0 <= index < len(manifest["rules"]), "invalid rule table index")
    rule = dict(manifest["rules"][index])
    environment = rule["environment"]
    require(type(environment) is int and 0 <= environment < len(manifest["environments"]),
            "invalid rule environment index")
    rule["environment"] = manifest["environments"][environment]
    return rule


def verify_trace(manifest, trace, recorded_cwd, project):
    """Compare every original event and ordered file/rule/config to its table."""
    position, next_plan, active, plan_keys = 0, 0, None, set()
    names, rules_seen = set(), set()
    counts = {name: 0 for name in COUNTS}
    for line in trace.decode("utf-8").splitlines():
        require(line.strip(), "blank original trace record")
        row = bench.parse_json(line)
        require(row.get("schemaVersion") == 1, "unexpected original trace schema")
        key = (row.get("pid"), row.get("planSequence"))
        require(all(type(value) is int and value > 0 for value in key), "invalid original plan identity")
        event = row.get("event")
        if event == "planStart":
            require(active is None and key not in plan_keys and next_plan < len(manifest["plans"]),
                    "interleaved, duplicated or extra original trace plan")
            require(row.get("fileCount") == manifest["plans"][next_plan]["files"],
                    "original plan count differs from manifest")
            active = {"key": key, "index": next_plan, "files": 0, "count": row["fileCount"]}
            next_plan += 1
            plan_keys.add(key)
        elif event == "file":
            require(active is not None and active["key"] == key and position < len(manifest["files"]),
                    "original file outside a live plan or extra file")
            require(row.get("fileIndex") == active["files"], "original trace reordered or omitted a file")
            expected = manifest["files"][position]
            require(len(expected) == 6, "unknown manifest file table contract")
            name, hash_value, byte_count, config_index, plan_index, file_index = expected
            try:
                actual_name = Path(row["filePath"]).relative_to(recorded_cwd).as_posix()
            except ValueError as error:
                raise bench.BenchmarkError("original admitted path escapes recorded cwd") from error
            require((actual_name, plan_index, file_index) == (name, active["index"], active["files"]),
                    "original admission order/path/plan differs from manifest")
            require(row.get("sourceSHA256") == hash_value and row.get("sourceBytes") == byte_count,
                    "original trace source hash/bytes differ from manifest: " + name)
            current = inside(project, name)
            require(current.is_file() and current.stat().st_size == byte_count
                    and bench.file_digest(current) == hash_value, "frozen admitted input changed: " + name)
            require(type(config_index) is int and 0 <= config_index < len(manifest["configurations"]),
                    "invalid configuration index")
            config = manifest["configurations"][config_index]
            for field in ("hasTypeChecker", "hasPreparedRules", "pluginLanguageOptions"):
                require(row.get(field) == config[field], "original configuration metadata differs: " + name)
            require(type(row["hasTypeChecker"]) is bool and type(row["hasPreparedRules"]) is bool,
                    "nonboolean original checker capability")
            environment = config["environment"]
            require(type(environment) is int and 0 <= environment < len(manifest["environments"]),
                    "invalid configuration environment index")
            require(row.get("environment") == manifest["environments"][environment],
                    "original configuration environment differs: " + name)
            expected_groups = {group: [expanded_rule(manifest, index) for index in config["rules"][group]]
                               for group in ("all", "native", "plugins")}
            require(row.get("rules") == expected_groups, "original full rule/options/severity metadata differs: " + name)
            enabled = expected_groups["all"]
            counts["rule_passes"] += len(enabled)
            counts["type_aware_rule_passes"] += sum(rule["requiresTypeInformation"] for rule in enabled)
            counts["typed_files"] += int(row["hasTypeChecker"])
            names.add(name)
            rules_seen.update(rule["name"] for rule in enabled)
            position += 1
            active["files"] += 1
        elif event == "planEnd":
            require(active is not None and active["key"] == key
                    and row.get("fileCount") == active["files"] == active["count"],
                    "missing or incomplete original trace plan")
            active = None
        else:
            raise bench.BenchmarkError("unexpected original trace event")
    require(active is None and position == len(manifest["files"]) and next_plan == len(manifest["plans"]),
            "truncated original trace")
    counts.update(plans=next_plan, file_passes=position, unique_files=len(names), unique_rule_count=len(rules_seen))
    require(counts == manifest["counts"], "original trace totals differ from manifest")
    return counts


# Keep the same Node resolver and localeCompare traversal as the original capture
# so a hash mismatch represents changed input, rather than Python sorting rules.
CURRENT_METADATA = r"""
import fs from 'node:fs';
import path from 'node:path';
import {createRequire} from 'node:module';
import {createHash} from 'node:crypto';
const project = process.argv[1], manifest = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'));
const hash = data => createHash('sha256').update(data).digest('hex');
const canonical = value => Array.isArray(value) ? `[${value.map(canonical).join(',')}]` :
  value !== null && typeof value === 'object' ? `{${Object.keys(value).sort().map(key => `${JSON.stringify(key)}:${canonical(value[key])}`).join(',')}}` : JSON.stringify(value);
const require = createRequire(path.join(project, 'package.json'));
const anchors = manifest.preparation.configurationRuntimeAnchors.map(anchor => {
  const metadata = require.resolve(`${anchor.package}/package.json`), root = path.dirname(metadata);
  return {package: anchor.package, version: JSON.parse(fs.readFileSync(metadata, 'utf8')).version,
    files: anchor.files.map(([name]) => {
      if (path.isAbsolute(name) || name.split('/').includes('..')) throw new Error('invalid anchor path');
      return [name, hash(fs.readFileSync(path.join(root, name)))];
    })};
});
const roots = [];
for (const parent of ['packages', 'scripts']) {
  const directory = path.join(project, parent);
  if (!fs.existsSync(directory)) continue;
  for (const entry of fs.readdirSync(directory).sort()) {
    for (const suffix of ['dist', 'compiled']) {
      const name = `${parent}/${entry}/${suffix}`;
      if (fs.existsSync(path.join(project, name))) roots.push(name);
    }
  }
}
const generated = roots.map(name => {
  const root = path.join(project, name), rows = [];
  function visit(directory) {
    for (const entry of fs.readdirSync(directory, {withFileTypes:true}).sort((a,b) => a.name.localeCompare(b.name))) {
      const full = path.join(directory, entry.name);
      if (entry.isSymbolicLink()) throw new Error('generated source symlink needs explicit provenance');
      if (entry.isDirectory()) visit(full);
      else if (entry.isFile()) rows.push([path.relative(root, full).split(path.sep).join('/'), fs.statSync(full).size, hash(fs.readFileSync(full))]);
      else throw new Error('unsupported generated input kind');
    }
  }
  visit(root);
  return {path:name, files:rows.length, bytes:rows.reduce((sum,row) => sum + row[1], 0), sha256:hash(canonical(rows))};
});
console.log(JSON.stringify({anchors,generated,node:process.version,packageManager:JSON.parse(fs.readFileSync(path.join(project,'package.json'),'utf8')).packageManager}));
"""


def verify(manifest_path, project, baseline, evidence, node=None):
    manifest_path, project, baseline = (Path(value).resolve() for value in (manifest_path, project, baseline))
    node = str(Path(node or shutil.which("node")).resolve())
    manifest = bench.read_json(manifest_path)
    require(manifest["schemaVersion"] == 1 and manifest["project"]["commit"] == PROJECT_PIN
            and manifest["baseline"]["commit"] == BASELINE_PIN and manifest["counts"] == COUNTS,
            "not the frozen rsbuild workload")
    for root, pin in ((project, PROJECT_PIN), (baseline, BASELINE_PIN)):
        require(command(["git", "-C", str(root), "rev-parse", "HEAD"]) == pin, "checkout commit differs")
        require(not command(["git", "-C", str(root), "diff", "--name-only", "HEAD"]), "tracked frozen checkout inputs are dirty")
    compiler = command(["git", "-C", str(baseline), "ls-tree", BASELINE_PIN, "typescript-go"]).split()
    require(len(compiler) >= 3 and compiler[2] == manifest["baseline"]["compilerCommit"], "baseline compiler pin differs")
    raw, hashes = {}, {}
    for key, expected_key in EVIDENCE_HASHES.items():
        raw[key] = Path(evidence[key]).read_bytes()
        hashes[key] = bench.digest_bytes(raw[key])
        require(hashes[key] == manifest["baseline"][expected_key], "original baseline " + key + " hash differs")
    for key, length_key in (("stdout", "stdoutBytes"), ("stderr", "stderrBytes")):
        raw[key] = Path(evidence[key]).read_bytes()
        hashes[key] = bench.digest_bytes(raw[key])
        require(len(raw[key]) == manifest["baseline"][length_key] == 0, "original baseline output was not silent")
    source, summary, record = (bench.parse_json(raw[key].decode("utf-8")) for key in ("source", "summary", "record"))
    require(source["sha"] == PROJECT_PIN and source["node"] == manifest["project"]["node"]
            and list(source["files"].items()) == [tuple(row) for row in manifest["project"]["inputs"]],
            "original source record differs from manifest")
    for name, expected in manifest["project"]["inputs"]:
        require(bench.file_digest(inside(project, name)) == expected, "frozen project configuration input changed: " + name)
    require(record["exit_code"] == summary["exit_code"] == manifest["baseline"]["exitCode"] == 0
            and summary["correctness_only"] is True and summary["trace_sha256"] == hashes["trace"],
            "original baseline failed or has a different trace")
    require(record["argv"][-3:] == ["lint", "--format", "jsonline"] and record["env"]["CI"] == "true"
            and Path(record["cwd"]).is_absolute(), "unexpected original baseline command")
    counts = verify_trace(manifest, raw["trace"], Path(record["cwd"]), project)
    require(all(summary[name] == value for name, value in counts.items()), "original summary counts differ")
    current = bench.parse_json(command([node, "--input-type=module", "-e", CURRENT_METADATA, str(project), str(manifest_path)]))
    require(current["node"] == manifest["project"]["node"] and current["packageManager"] == manifest["project"]["packageManager"],
            "current Node/package manager declaration differs from frozen capture")
    require(current["anchors"] == manifest["preparation"]["configurationRuntimeAnchors"], "current configuration runtime anchors differ")
    require(current["generated"] == manifest["preparation"]["generatedOutputs"], "current generated checker inputs differ")
    return {"schema_version": 1, "status": "passed", "manifest_sha256": bench.file_digest(manifest_path),
            "verifier_sha256": bench.file_digest(__file__), "evidence_sha256": hashes, "counts": counts,
            "node": current["node"], "node_sha256": bench.file_digest(node),
            "original_concurrency": {name: record["env"].get(name) for name in ("GOMAXPROCS", "RAYON_NUM_THREADS", "GOMEMLIMIT")},
            "checks": {"original_trace_matches_manifest": True, "current_frozen_sources_and_configs": True,
                       "current_runtime_anchors": True, "current_generated_outputs": True},
            "single_threaded_project_selection_validated": False,
            "historical_transitive_dependency_closure_validated": False,
            "limitations": ["Original trace concurrency is preserved; a new same-concurrency project/type selection check is still required.",
                            "Current runtime anchors and generated outputs were added after the original capture and do not prove its complete historical dependency closure.",
                            "Admission evidence does not prove rule diagnostic, fix or suggestion equivalence."]}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("manifest", "project", "baseline", *["baseline-" + key for key in (*EVIDENCE_HASHES, "stdout", "stderr")], "output"):
        parser.add_argument("--" + name, type=Path, required=True)
    parser.add_argument("--node", type=Path)
    args = parser.parse_args()
    require(not args.output.exists(), "use a fresh manifest verification output")
    evidence = {key: getattr(args, "baseline_" + key) for key in (*EVIDENCE_HASHES, "stdout", "stderr")}
    result = verify(args.manifest, args.project, args.baseline, evidence, args.node)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    bench.save_json(args.output, result)
    print(json.dumps({"status": "passed", "manifest_sha256": result["manifest_sha256"], "counts": result["counts"]}))


if __name__ == "__main__":
    main()
