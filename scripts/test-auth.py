#!/usr/bin/env python3
"""Run auth integration checks against the project's development database."""
from pathlib import Path
import subprocess
import sys
from env import environment

root = Path(__file__).resolve().parents[1]
env = environment()
args = ['go','test','-race','-count=1','-tags=integration']
if '--http' in sys.argv:
    env['AUTH_SMOKE_URL'] = env.get('AUTH_SMOKE_URL', f"http://127.0.0.1:{env['ADMIN_API_PORT']}")
    args += ['-run','^TestHTTPAuthLifecycle$']
subprocess.run(args + ['./services/system/internal/application/auth'], cwd=root, env=env, check=True)
