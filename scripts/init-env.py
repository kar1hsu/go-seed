#!/usr/bin/env python3
"""Prepare local development settings without replacing existing values."""
from pathlib import Path
import secrets

root = Path(__file__).resolve().parents[1]
path = root / '.env'
if not path.exists():
    path.write_text((root / '.env.example').read_text())
    path.chmod(0o600)
lines = path.read_text().splitlines()
existing = dict(line.split('=', 1) for line in lines if line and not line.startswith('#') and '=' in line)
defaults = dict(line.split('=', 1) for line in (root / '.env.example').read_text().splitlines() if line and not line.startswith('#') and '=' in line)
settings = {**defaults,
    'ADMIN_TOKEN_SECRET': secrets.token_urlsafe(48),
    'ADMIN_RPC_SECRET': secrets.token_urlsafe(48),
    'ADMIN_BOOTSTRAP_PASSWORD': secrets.token_urlsafe(24),
    'ADMIN_COOKIE_SECURE': 'false',
    'ADMIN_ALLOWED_ORIGINS': ','.join(f'http://{host}:{existing.get(key, default)}' for host in ('127.0.0.1', 'localhost') for key, default in [('WEB_PORT','5173'),('GATEWAY_PORT','8088'),('ADMIN_API_PORT','8081')]),
}
for key, value in settings.items():
    if existing.get(key, '').strip().strip('"\''):
        continue
    lines = [line for line in lines if not line.startswith(key + '=')]
    lines.append(f'{key}={value}')
path.write_text('\n'.join(lines) + '\n')
path.chmod(0o600)
print('Local .env is ready; existing settings preserved. Initial password: ADMIN_BOOTSTRAP_PASSWORD in .env.')
