"""Load local KEY=value settings without executing shell code."""
from pathlib import Path
import os
import re
import shlex

def environment():
    path = Path(__file__).resolve().parents[1] / '.env'
    if not path.is_file():
        raise SystemExit('Missing .env: cp .env.example .env')
    env = os.environ.copy()
    for line in ((path.parent / '.env.example').read_text() + '\n' + path.read_text()).splitlines():
        line = line.strip()
        if not line or line.startswith('#'):
            continue
        key, sep, value = line.partition('=')
        if not sep or not re.fullmatch(r'[A-Z][A-Z0-9_]*', key):
            raise SystemExit('Invalid .env entry (expected KEY=value)')
        values = shlex.split(value, comments=True)
        if len(values) > 1:
            raise SystemExit(f'Quote spaces in .env value: {key}')
        env[key] = values[0] if values else ''
    return env
