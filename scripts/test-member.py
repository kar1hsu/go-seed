#!/usr/bin/env python3
"""Use isolated temporary databases; HTTP checks create and remove named fixtures."""
from pathlib import Path
import subprocess
import sys
from env import environment
root = Path(__file__).resolve().parents[1]
env = environment()
args = ['go','test','-race','-count=1','-tags=integration']
if '--http' in sys.argv:
    env['MEMBER_SMOKE'] = '1'
    args += ['-run','^TestHTTPMember$']
subprocess.run(args + ['./services/member/internal/application/account'], cwd=root, env=env, check=True)
