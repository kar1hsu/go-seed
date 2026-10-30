#!/usr/bin/env python3
"""Prepare local development settings without replacing existing values."""
from pathlib import Path
import json
import os
import secrets
import shlex
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]

def entries(text):
    result = {}
    for line in text.splitlines():
        line = line.strip()
        if not line or line.startswith('#'):
            continue
        key, sep, value = line.partition('=')
        if not sep:
            raise ValueError('Invalid .env entry')
        parts = shlex.split(value, comments=True)
        if len(parts) > 1:
            raise ValueError(f'Quote spaces in .env value: {key}')
        result[key] = parts[0] if parts else ''
    return result

def key_command():
    source = ROOT / 'cmd/keygen/main.go'
    if shutil.which('go'):
        return ['go', 'run', str(source)]
    # No Go installation is needed for the documented container-only workflow.
    return ['docker', 'run', '--rm', '-i', '--network', 'none',
            '-v', f'{source}:/keygen.go:ro', '-e', 'GOTOOLCHAIN=local',
            'golang:1.25.11-bookworm', 'go', 'run', '/keygen.go']

def initialize(root=ROOT):
    path = root / '.env'
    template = (root / '.env.example').read_text()
    text = path.read_text() if path.exists() else template
    lines, existing = text.splitlines(), entries(text)
    defaults = entries(template)
    settings = {**defaults,
        **{key: secrets.token_urlsafe(32) for key in ('MYSQL_PASSWORD', 'MYSQL_ROOT_PASSWORD', 'MEMBER_MYSQL_PASSWORD')},
        **{key: secrets.token_urlsafe(48) for key in ('MEMBER_ADMIN_RPC_SECRET', 'MEMBER_APP_RPC_SECRET', 'MEMBER_SYSTEM_RPC_SECRET', 'ADMIN_RPC_SECRET')},
        'ADMIN_BOOTSTRAP_PASSWORD': secrets.token_urlsafe(24),
        'ADMIN_COOKIE_SECURE': 'false',
        'MEMBER_ALLOWED_ORIGINS': ','.join(f'http://{host}:{existing.get(key, default)}' for host in ('127.0.0.1', 'localhost') for key, default in [('GATEWAY_PORT','8088'),('APP_API_PORT','8080')]),
        'ADMIN_ALLOWED_ORIGINS': ','.join(f'http://{host}:{existing.get(key, default)}' for host in ('127.0.0.1', 'localhost') for key, default in [('WEB_PORT','5173'),('GATEWAY_PORT','8088'),('ADMIN_API_PORT','8081')]),
    }
    fields = {'PublicKey': 'TOKEN_PUBLIC_KEY', 'PrivateKey': 'TOKEN_PRIVATE_KEY', 'RefreshKey': 'REFRESH_KEY'}
    domains = ('ADMIN', 'MEMBER')
    supplied = [{field: existing.get(f'{domain}_{suffix}', '') for field, suffix in fields.items()} for domain in domains]
    # Keys stay off argv and logs. A validation failure leaves .env untouched.
    result = subprocess.run(key_command(), input=json.dumps(supplied), text=True, capture_output=True)
    if result.returncode:
        raise SystemExit('Identity key initialization failed: ' + result.stderr.strip())
    generated = json.loads(result.stdout)
    for domain, keys in zip(domains, generated, strict=True):
        for field, suffix in fields.items():
            settings[f'{domain}_{suffix}'] = keys[field]
    for key, value in settings.items():
        if existing.get(key, '') or key in existing and value == '':
            continue
        lines = [line for line in lines if not line.strip().startswith(key + '=')]
        lines.append(f'{key}={value}')
    descriptor, temporary = tempfile.mkstemp(prefix='.env.', dir=root)
    try:
        with os.fdopen(descriptor, 'w') as output:
            output.write('\n'.join(lines) + '\n')
        os.chmod(temporary, 0o600)
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)

if __name__ == '__main__':
    initialize()
    print('Local .env is ready; existing settings preserved. Initial password: ADMIN_BOOTSTRAP_PASSWORD in .env.')
