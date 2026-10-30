#!/usr/bin/env python3
"""Run built services and Vite, terminating only our own process groups."""
import os
from pathlib import Path
import signal
import subprocess
import sys
import time
from env import environment

root = Path(__file__).resolve().parents[1]
os.chdir(root)
env = environment()
children = []

def interrupted(signum, _):
    raise KeyboardInterrupt

signal.signal(signal.SIGTERM, interrupted)
signal.signal(signal.SIGINT, interrupted)
try:
    commands = [
        ['bin/system-worker', '-f', 'services/system/worker/etc/system-worker.yaml'],
        ['bin/system-scheduler', '-f', 'services/system/scheduler/etc/system-scheduler.yaml'],
        ['bin/system-rpc', '-f', 'services/system/rpc/etc/system.yaml'],
        ['bin/member-rpc', '-f', 'services/member/rpc/etc/member.yaml'],
        ['bin/admin-api', '-f', 'apps/admin/api/etc/admin-api.yaml'],
        ['bin/app-api', '-f', 'apps/app/api/etc/app-api.yaml'],
        ['pnpm', '--dir', 'web/admin', 'dev'],
    ]
    for command in commands:
        children.append(subprocess.Popen(command, env=env, start_new_session=True))
    print(f"GoSeed admin: http://127.0.0.1:{env['WEB_PORT']} (Ctrl+C stops all child processes)", flush=True)
    while True:
        for child in children:
            if child.poll() is not None:
                raise RuntimeError(f'Process exited with status {child.returncode}')
        time.sleep(.25)
except KeyboardInterrupt:
    pass
except Exception as error:
    print(error, file=sys.stderr)
    sys.exit(1)
finally:
    for child in children:
        if child.poll() is None:
            os.killpg(child.pid, signal.SIGTERM)
    for child in children:
        try:
            child.wait(timeout=12)
        except subprocess.TimeoutExpired:
            os.killpg(child.pid, signal.SIGKILL)
            child.wait()
