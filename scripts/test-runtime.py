#!/usr/bin/env python3
"""Build and test the shared Go runtime image; requires Go and Docker."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
IMAGE = 'goseed-runtime-test:local'


def main():
    subprocess.run(['docker', 'build', '-f', 'deploy/docker/service.Dockerfile',
                    '--build-arg', 'SERVICE_PATH=services/system/rpc', '-t', IMAGE, '.'],
                   cwd=ROOT, check=True)
    metadata = json.loads(subprocess.check_output(['docker', 'image', 'inspect', IMAGE], text=True))[0]
    build_env = {**os.environ, 'CGO_ENABLED': '0', 'GOOS': 'linux', 'GOARCH': metadata['Architecture']}
    with tempfile.TemporaryDirectory(prefix='goseed-runtime-') as directory:
        binary = str(Path(directory) / 'verify-runtime')
        subprocess.run(['go', 'build', '-trimpath', '-o', binary, './tests/runtime'],
                       cwd=ROOT, env=build_env, check=True)
        for zone in ('Asia/Shanghai', 'UTC', 'America/New_York'):
            subprocess.run(['docker', 'run', '--rm', '--network', 'none', '--read-only',
                            '--mount', f'type=bind,src={binary},dst=/verify-runtime,readonly',
                            '--mount', 'type=volume,dst=/app/data/files',
                            '-e', f'TZ={zone}', '--entrypoint', '/verify-runtime', IMAGE], check=True)


if __name__ == '__main__':
    main()
