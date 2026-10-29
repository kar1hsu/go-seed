#!/usr/bin/env python3
"""RBAC tests use temporary databases; HTTP checks use uniquely named fixture records."""
from pathlib import Path
import subprocess
import sys
from env import environment
root = Path(__file__).resolve().parents[1]
env = environment()
args = ['go', 'test', '-race', '-count=1', '-tags=integration']
if '--http' in sys.argv:
    env['RBAC_SMOKE_URL'] = env.get('RBAC_SMOKE_URL', f"http://127.0.0.1:{env['ADMIN_API_PORT']}")
    args += ['-run', '^TestHTTPManagement$']
subprocess.run(args + ['./services/system/internal/application/management'], cwd=root, env=env, check=True)
