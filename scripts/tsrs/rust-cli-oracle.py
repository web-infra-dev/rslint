#!/usr/bin/env python3
"""Compare real CLI lint and on-disk fixes against an external frozen baseline.

Run with with-machine-lock.py. This is a correctness fixture, never a performance
sample; the real Rsbuild checkout is not modified.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile


def sha(data):
    return hashlib.sha256(data).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--baseline", type=Path, required=True)
    parser.add_argument("--candidate", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    roots = {side: getattr(args, side).resolve() for side in ("baseline", "candidate")}
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    scripts = Path(__file__).resolve().parent
    node = shutil.which("node")
    assert node
    assert subprocess.check_output(["git", "-C", str(roots["baseline"]), "rev-parse", "HEAD"]).decode().strip() == "e0a29f751d4def86d27931759098f27748fa3003"
    rules = ["no-debugger", "no-var", "prefer-const", "@typescript-eslint/no-floating-promises", "@typescript-eslint/no-unsafe-call"]
    manifest = json.loads((scripts.parent.parent / "docs/tsrs-migration/rsbuild-workload.json").read_text())
    assert set(rules) <= {rule["name"] for rule in manifest["rules"]}
    sources = {
        "hit.ts": "import { external, pending } from './types';\nexport function run() { debugger; pending(); external(); var count = 1; return count; }\n",
        "clean.ts": "import { pending } from './types';\nexport async function run() { await pending(); const count = 1; return count; }\n",
        "types.ts": "export declare function pending(): Promise<string>;\nexport declare const external: any;\n",
        "tsconfig.json": json.dumps({"compilerOptions": {"strict": True, "target": "ESNext", "module": "NodeNext", "moduleResolution": "NodeNext"}, "include": ["*.ts"]}),
        "rslint.config.mjs": "export default " + json.dumps([{"files": ["**/*.ts"], "plugins": ["@typescript-eslint"], "languageOptions": {"parserOptions": {"projectService": True}}, "rules": {rule: "error" for rule in rules}}]) + ";\n",
    }
    report = {"status": "incomplete", "scope": "CLI fixture only; full Rsbuild proof is separate", "rules": rules, "source_sha256": {name: sha(text.encode()) for name, text in sources.items()}, "backends": {}}
    try:
        for side, backend in roots.items():
            with tempfile.TemporaryDirectory(prefix="rslint-cli-oracle-") as temporary:
                fixture = Path(temporary).resolve()
                for name, text in sources.items():
                    (fixture / name).write_text(text)
                env = {**os.environ, "CI": "true", "NODE_DISABLE_COMPILE_CACHE": "1", "RAYON_NUM_THREADS": "1", "GOMAXPROCS": "1"}
                argv = [node]
                if side == "candidate":
                    addon = backend / "npm/rslint/darwin-arm64/rslint.darwin-arm64.node"
                    assert addon.is_file(), "This witness currently targets the measured Darwin ARM64 artifact"
                    report["candidate_addon_sha256"] = sha(addon.read_bytes())
                    env["RSLINT_RUST_GUARD"] = str(addon)
                    argv += ["--import", str(scripts / "rust-rsbuild-observe.mjs")]
                argv += [str(backend / "packages/rslint/bin/rslint.js"), "--config", str(fixture / "rslint.config.mjs"), "--singleThreaded", "--format", "jsonline", "."]
                rounds = []
                for stage, fix in (("lint", False), ("fix", True), ("after_fix", False)):
                    command = argv + (["--fix"] if fix else [])
                    result = subprocess.run(command, cwd=fixture, env=env, capture_output=True, timeout=120)
                    stdout, stderr = result.stdout.decode().replace(str(fixture), "<fixture>"), result.stderr.decode().replace(str(fixture), "<fixture>")
                    (output / (side + "-" + stage + ".stdout")).write_text(stdout)
                    (output / (side + "-" + stage + ".stderr")).write_text(stderr)
                    diagnostics = sorted([json.loads(line) for line in stdout.splitlines() if line.strip()], key=lambda item: json.dumps(item, sort_keys=True))
                    rounds.append({"stage": stage, "exit_code": result.returncode, "diagnostics": diagnostics, "stderr": stderr, "disk": {name: (fixture / name).read_text() for name in sources}})
                    assert result.returncode == 1 and stderr == "", (side, stage, rounds[-1])
                report["backends"][side] = rounds
        assert report["backends"]["baseline"] == report["backends"]["candidate"], "CLI diagnostics or on-disk fixes differ"
        first, fixed, final = report["backends"]["baseline"]
        names = {message["ruleName"] for message in first["diagnostics"]}
        assert {"no-debugger", "no-var", "@typescript-eslint/no-floating-promises", "@typescript-eslint/no-unsafe-call"} <= names
        assert all(message["filePath"] == "hit.ts" for message in first["diagnostics"]), "clean input or support declarations unexpectedly reported"
        assert first["disk"] == sources, "ordinary lint changed input"
        assert fixed["disk"]["hit.ts"] != sources["hit.ts"] and "const count = 1" in fixed["disk"]["hit.ts"], "CLI did not write its fix"
        assert fixed["disk"] == final["disk"], "final lint changed fixed sources"
        report.update(status="passed", checks={"hit": True, "no_hit": True, "type_aware_hit": True, "type_aware_no_hit": True, "cli_fix_disk": True, "no_tsgo_fallback": True})
    finally:
        (output / "result.json").write_text(json.dumps(report, indent=2) + "\n")
    print("passed: public CLI typed hit/no-hit, diagnostics and on-disk fixes")


if __name__ == "__main__":
    main()
