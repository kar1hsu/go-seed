#!/usr/bin/env python3
"""Always run storage protocol tests; optionally run against a dedicated S3 test endpoint."""
from pathlib import Path
import os
import subprocess
root = Path(__file__).resolve().parents[1]
subprocess.run(['go', 'test', '-race', '-count=1', './pkg/storage'], cwd=root, check=True)
keys = ['GOSEED_S3_TEST_ENDPOINT', 'GOSEED_S3_TEST_ACCESS', 'GOSEED_S3_TEST_SECRET']
if any(os.getenv(key) for key in keys):
    if not all(os.getenv(key) for key in keys):
        raise SystemExit('Set all GOSEED_S3_TEST_ENDPOINT/ACCESS/SECRET for a dedicated S3 test service.')
    subprocess.run(['go', 'test', '-race', '-tags=integration', '-run', '^TestS3RoundTrip$', '-count=1', './pkg/storage'], cwd=root, check=True)
else:
    print('S3 protocol tests passed. Live S3 roundtrip skipped: no dedicated test endpoint configured.')
