#!/usr/bin/env python3
from pathlib import Path
import json
import re
import subprocess

root = Path(__file__).resolve().parents[1]
# Runtime CI need not fetch AI documentation. Gitlinks are checked by doctor locally.
submodule_paths = [root/'.ai-context/go-zero', root/'.agents/skills/zero-skills']
paths = [p for area in ('apps', 'services', 'pkg', 'cmd') for p in (root / area).rglob('*.go')]
formatted = subprocess.check_output(['gofmt', '-l', *map(str, paths)], text=True)
if formatted:
    raise SystemExit('Run gofmt:\n' + formatted)
for area in ('apps', 'services', 'pkg', 'cmd', 'web'):
    for p in (root / area).rglob('go.mod'):
        if 'node_modules' not in p.parts:
            raise SystemExit(f'Nested Go module: {p}')
modules = subprocess.check_output(['go', 'list', '-f', '{{.ImportPath}} {{join .Imports " "}}', './...'], cwd=root, text=True)
base = 'github.com/kar1hsu/go-seed/'
for line in modules.splitlines():
    package, *imports = line.split()
    for dependency in imports:
        if package.startswith(base+'pkg/') and dependency.startswith((base+'apps/', base+'services/')):
            raise SystemExit(f'Public package imports business implementation: {package} -> {dependency}')
        if package.startswith(base+'apps/') and dependency.startswith(base+'services/'):
            if not re.fullmatch(re.escape(base) + r'services/[^/]+/rpc/(pb|[^/]+)', dependency):
                raise SystemExit(f'API crosses domain boundary: {package} -> {dependency}')
            if '/internal' in dependency:
                raise SystemExit(f'API imports internal implementation: {dependency}')
for doc in [root/'README.md',root/'AGENTS.md',*(root/'docs').glob('*.md'),*(root/'apps').rglob('*.md'),*(root/'services').rglob('*.md'),root/'web/admin/README.md']:
    for target in re.findall(r'\[[^\]]*\]\(([^)]+)\)', doc.read_text()):
        target = target.split('#')[0]
        if not target or '://' in target or target.startswith('mailto:'):
            continue
        resolved = (doc.parent / target).resolve()
        if not resolved.exists() and any(resolved.is_relative_to(path) and not (path/'README.md').exists() for path in submodule_paths):
            continue
        if not resolved.exists():
            raise SystemExit(f'Broken relative link: {doc.relative_to(root)} -> {target}')
print('Formatting, module boundaries and documentation links passed.')
