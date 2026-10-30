#!/usr/bin/env python3
"""Execute domain commands with the same local settings as make dev."""
from pathlib import Path
import subprocess
import sys
from env import environment, service_environment

root = Path(__file__).resolve().parents[1]
env = environment()
args = sys.argv[1:]
if args and args[0] == '--compose':
    args = args[1:]
    command = ['docker','compose','--env-file','.env','-p','goseed','-f','deploy/compose/compose.yaml','run','--rm','--no-deps','-e','ADMIN_TOKEN_PUBLIC_KEY=','-e','ADMIN_TOKEN_PRIVATE_KEY=','-e','ADMIN_REFRESH_KEY=','--entrypoint','/app/systemctl','system-rpc']
    if 'bootstrap' in args:
        command[command.index('--entrypoint'):command.index('--entrypoint')] = ['-e', 'ADMIN_BOOTSTRAP_PASSWORD']
else:
    command = ['go','run','./services/system/cmd/system']
    env = service_environment(env, 'system-command', ['ADMIN_BOOTSTRAP_PASSWORD'] if 'bootstrap' in args else [])
subprocess.run(command + ['-f','services/system/rpc/etc/system.yaml'] + args, cwd=root, env=env, check=True)
