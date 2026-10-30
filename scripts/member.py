#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys
from env import environment, service_environment
root = Path(__file__).resolve().parents[1]
args = sys.argv[1:]
env = environment()
if args and args[0] == '--compose':
    args = args[1:]
    command = ['docker','compose','--env-file','.env','-p','goseed','-f','deploy/compose/compose.yaml','run','--rm','--no-deps','-e','MEMBER_TOKEN_PUBLIC_KEY=','-e','MEMBER_TOKEN_PRIVATE_KEY=','-e','MEMBER_REFRESH_KEY=','--entrypoint','/app/memberctl','member-rpc']
else:
    command = ['go','run','./services/member/cmd/member']
    env = service_environment(env, 'member-command')
subprocess.run(command + ['-f','services/member/rpc/etc/member.yaml'] + args, cwd=root, env=env, check=True)
