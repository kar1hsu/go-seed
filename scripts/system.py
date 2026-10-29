#!/usr/bin/env python3
"""Execute domain commands with the same local settings as make dev."""
from pathlib import Path
import subprocess
import sys
from env import environment

root = Path(__file__).resolve().parents[1]
env = environment()
args = sys.argv[1:]
if args and args[0] == '--compose':
    args = args[1:]
    command = ['docker','compose','--env-file','.env','-p','goseed','-f','deploy/compose/compose.yaml','run','--rm','--no-deps','-e','ADMIN_BOOTSTRAP_PASSWORD','--entrypoint','/app/systemctl','system-rpc']
else:
    command = ['go','run','./services/system/cmd/system']
subprocess.run(command + ['-f','services/system/rpc/etc/system.yaml'] + args, cwd=root, env=env, check=True)
