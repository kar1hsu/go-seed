#!/usr/bin/env python3
"""Regenerate in an isolated copy, then prove additive contracts preserve handwritten code."""
from pathlib import Path
import hashlib
import os
import shutil
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
areas = ['apps', 'services', 'pkg', 'cmd', 'templates', 'docs/openapi', 'web/admin/src/api/generated', 'web/admin/src/router']
def snapshot(base):
    return {str(p.relative_to(base)): hashlib.sha256(p.read_bytes()).hexdigest()
            for area in areas for p in (base/area).rglob('*') if p.is_file()}
def run(command, cwd):
    result = subprocess.run(command, cwd=cwd, env=env, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if result.returncode:
        raise SystemExit(result.stdout)

env = os.environ.copy()
env['PATH'] = str(root/'.tools/bin') + os.pathsep + env['PATH']
# Inherited selectors must not silently skip half of the generated contract set.
env.pop('APP', None)
env.pop('DOMAIN', None)
with tempfile.TemporaryDirectory(prefix='goseed-gen-') as temporary:
    target = Path(temporary)/'repo'
    shutil.copytree(root, target, ignore=shutil.ignore_patterns('.git','.ai-context','.agents','.tools','.tmp','node_modules','dist','bin','.env','.data','__pycache__'))
    (target/'web/admin/node_modules').symlink_to(root/'web/admin/node_modules', target_is_directory=True)
    before = snapshot(target)
    run(['bash','scripts/gen.sh','all'], target)
    after = snapshot(target)
    differences = sorted(k for k in before.keys() | after.keys() if before.get(k) != after.get(k))
    if differences:
        raise SystemExit('Generated files are stale (run make gen):\n'+'\n'.join(differences))
    handwritten = {k:v for k,v in after.items() if k.endswith('.go') and ('/logic/' in k or '/svc/' in k or '/config/' in k or '/handler/' in k and not k.endswith('/routes.go') or k in ['apps/admin/api/admin.go','apps/app/api/app.go','services/system/rpc/system.go'])}
    with (target/'apps/admin/api/desc/system/ping.api').open('a') as output:
        output.write('\ntype GenerationProbeRequest { Value string `json:"value"` }\n@server(prefix: /admin-api/v1\n group: generation)\nservice admin-api {\n @handler GenerationProbe\n post /generation-probe (GenerationProbeRequest) returns (PingResponse)\n}\n')
    with (target/'apps/admin/api/resources/endpoints.yaml').open('a') as output:
        output.write('- {method: post, path: /admin-api/v1/generation-probe, access: public}\n')
    proto = target/'services/system/rpc/system.proto'
    proto.write_text(proto.read_text().replace('service System {','service System {\n rpc GenerationProbe(PingRequest) returns (PingResponse);'))
    run(['bash','scripts/gen.sh','all'], target)
    current = snapshot(target)
    for name, value in handwritten.items():
        if current.get(name) != value:
            raise SystemExit(f'Generation overwrote handwritten file: {name}')
    assert (target/'apps/admin/api/internal/handler/generation/generation_probe_handler.go').is_file()
    assert (target/'services/system/rpc/internal/logic/generation_probe_logic.go').is_file()
    run(['go','build','./...'], target)
print('Repeat generation is identical; additive API/RPC contracts preserve handwritten code and compile.')
