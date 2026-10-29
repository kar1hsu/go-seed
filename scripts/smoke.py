#!/usr/bin/env python3
"""Check real running services, including the final HTTP/OpenAPI envelope."""
import json
import datetime
import re
from pathlib import Path
import time
import urllib.request
from env import environment

env = environment()
root = Path(__file__).resolve().parents[1]
def get_json(url):
    # Independent RPC clients can reconnect at different times after startup/outage.
    deadline = time.monotonic() + 30
    while True:
        try:
            with urllib.request.urlopen(url, timeout=4) as response:
                return json.load(response)
        except OSError:
            if time.monotonic() >= deadline:
                raise
            time.sleep(.5)

for app, key, prefix in [('admin','ADMIN_API_PORT','admin-api'),('app','APP_API_PORT','api')]:
    base = f"http://127.0.0.1:{env[key]}"
    for probe in ('healthz', 'readyz'):
        assert get_json(f'{base}/{probe}') == {'status':'ok'}
    body = get_json(f'{base}/{prefix}/v1/system/ping')
    assert body['code'] == 0 and body['message'] == 'ok', body
    assert body['data']['service'] == 'system-rpc' and body['data']['version'] == '0.1.0', body
    assert re.fullmatch(r'[0-9a-f]{32}', body['traceId']) and body['traceId'] != '0'*32, body
    assert datetime.datetime.fromisoformat(body['data']['time'].replace('Z', '+00:00')).utcoffset() == datetime.timedelta(0), body
    schema = json.loads((root/f'docs/openapi/{app}.openapi.json').read_text())['components']['schemas']['PingResponse']
    assert set(body) == set(schema['properties']), body
    print(f'{app}: health + readiness + API -> RPC envelope + trace ID passed')
print('Smoke checks passed.')
