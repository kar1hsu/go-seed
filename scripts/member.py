#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys
from env import environment
root = Path(__file__).resolve().parents[1]
args = sys.argv[1:]
if args and args[0] == '--compose':
    args = args[1:]
    command = ['docker','compose','--env-file','.env','-p','goseed','-f','deploy/compose/compose.yaml','run','--rm','--no-deps','--entrypoint','/app/memberctl','member-rpc']
else:
    command = ['go','run','./services/member/cmd/member']
subprocess.run(command + ['-f','services/member/rpc/etc/member.yaml'] + args, cwd=root, env=environment(), check=True)
