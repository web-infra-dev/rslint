#!/usr/bin/env python3
"""Inventory current declarations against the frozen rsbuild acceptance set.

This does not execute lint or mark any rule's behavior accepted. The production
engine does not import the workload manifest or this inventory.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[2]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, default=ROOT / 'docs/tsrs-migration/rsbuild-coverage.json')
    args = parser.parse_args()
    manifest_path = ROOT / 'docs/tsrs-migration/rsbuild-workload.json'
    manifest_bytes = manifest_path.read_bytes()
    manifest = json.loads(manifest_bytes)
    directory = ROOT / 'crates/rslint-core/src/rules'
    registry = (directory / 'mod.rs').read_text()
    match = re.search(r'pub\(crate\) fn supported\(name: &str\) -> bool \{(.*?)\n\}', registry, re.S)
    assert match, 'unknown registry format'
    names, sources = [], {'crates/rslint-core/src/rules/mod.rs': (directory / 'mod.rs').read_bytes()}
    for part in re.sub(r'\s', '', match[1]).split('||'):
        if module := re.fullmatch(r'([a-z_]+)::NAMES\.contains\(&name\)', part):
            filename = directory / f'{module[1]}.rs'
            content = filename.read_bytes()
            declaration = re.search(r'const NAMES: &\[&str\] = &\[(.*?)\];', content.decode(), re.S)
            assert declaration, filename
            names.extend(json.loads('[' + re.sub(r',\s*$', '', declaration[1]) + ']'))
            sources[str(filename.relative_to(ROOT))] = content
        else:
            constant = re.fullmatch(r'name==([A-Z_]+)', part)
            assert constant, f'unsupported registry expression: {part}'
            value = re.search(rf'const {constant[1]}: &str = "([^"]+)";', registry)
            assert value, constant[1]
            names.append(value[1])
    assert len(names) == len(set(names)), 'duplicate rule declaration'
    required = {}
    for rule in manifest['rules']:
        required.setdefault(rule['name'], []).append(rule)
    assert len(required) == 129
    rows = [{'name': name, 'configurations': configurations,
             'implementation': 'source_declared' if name in names else 'missing',
             'fullBehaviorAccepted': False}
            for name, configurations in sorted(required.items())]
    covered = len(set(names) & required.keys())
    result = {
        'schemaVersion': 1,
        'scope': 'Only rules actually enabled in frozen rsbuild configuration; the historical catalog is provenance, not the acceptance scope.',
        'manifestSHA256': hashlib.sha256(manifest_bytes).hexdigest(),
        'counts': {'requiredRules': len(required), 'sourceDeclared': covered,
                   'missing': len(required) - covered, 'retainedOutOfScope': len(names) - covered},
        'sourceSHA256': {name: hashlib.sha256(content).hexdigest() for name, content in sorted(sources.items())},
        'caveat': 'Declarations are not compatibility acceptance. Consult separately pinned build and behavior receipts.',
        'rules': rows,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result['counts']))


if __name__ == '__main__':
    main()
