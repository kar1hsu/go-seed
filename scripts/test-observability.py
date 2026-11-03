#!/usr/bin/env python3
"""Tail sampling and full-stack shutdown regressions in a disposable Compose project."""
import json
from pathlib import Path
import shlex
import subprocess
import tempfile
import time
from urllib import error, request
import uuid

ROOT = Path(__file__).resolve().parents[1]
DEPLOY = ROOT / 'deploy'
HTTP = request.build_opener(request.ProxyHandler({}))


def run(*args):
    return subprocess.check_output(args, cwd=ROOT, text=True).strip()


def fetch(url, data=None):
    req = request.Request(url, data=data, headers={'Accept': 'application/json', 'Content-Type': 'application/json'})
    try:
        response = HTTP.open(req, timeout=5)
    except error.HTTPError as exc:
        response = exc
    with response:
        return response.status, response.read()


def wait(label, probe, seconds=60):
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        try:
            if result := probe():
                return result
        except (OSError, TimeoutError):
            pass
        time.sleep(1)
    raise RuntimeError('timed out: ' + label)


# OTLP/HTTP fixtures enter through the same gRPC receiver as the Go services.
BRIDGE = """
otelcol.receiver.otlp "test_input" {
    http { endpoint = "0.0.0.0:4318" }
    output { traces = [otelcol.exporter.otlp.loopback.input] }
}
otelcol.exporter.otlp "loopback" {
    client {
        endpoint = "127.0.0.1:4317"
        tls { insecure = true }
    }
}
"""


def stored(url):
    status, body = fetch(url)
    return body if status == 200 else None


def test_sampling(otlp, tempo):
    # Each entry is one trace. A fast rejection mixed with a real fault must
    # retain the failure. A slow request still qualifies independently.
    cases = {
        'fast success': (False, [('http.status_code', 200)], 10),
        'API-only rejection': (False, [('http.status_code', 401)], 10),
        'slow success': (True, [('http.status_code', 200)], 1500),
        'server error': (True, [('http.status_code', 503)], 10),
        'mixed': (True, [('rpc.grpc.status_code', 16), ('rpc.grpc.status_code', 13)], 10),
    }
    for code in (2, 3, 4, 5, 6, 7, 8, 13, 14, 15, 16):
        cases[f'gRPC {code}'] = (code in (2, 4, 13, 14, 15), [('rpc.grpc.status_code', code)], 10)
    identifiers, batch = {}, []
    now = time.time_ns()
    for name, (_, attributes, duration) in cases.items():
        trace_id = identifiers[name] = uuid.uuid4().hex
        for key, code in attributes:
            is_rpc = key.startswith('rpc.')
            batch.append({
                'traceId': trace_id, 'spanId': uuid.uuid4().hex[:16], 'name': name,
                'kind': 2, 'startTimeUnixNano': str(now - duration * 10**6), 'endTimeUnixNano': str(now),
                'status': {'code': 2 if is_rpc or code >= 500 else 0},
                'attributes': [
                    {'key': key, 'value': {'intValue': str(code)}},
                    {'key': 'http.client_ip', 'value': {'stringValue': '203.0.113.77'}},
                    {'key': 'user_agent.original', 'value': {'stringValue': 'private-agent-marker'}},
                ],
            })
    payload = {'resourceSpans': [{'scopeSpans': [{'spans': batch}]}]}
    status, _ = fetch(otlp + '/v1/traces', json.dumps(payload).encode())
    if status != 200:
        raise RuntimeError(f'OTLP fixture rejected: HTTP {status}')
    # All traces were delivered in the same batch. Positive controls must be
    # visible before interpreting a 404 as a sampling decision.
    for name, (keep, _, _) in cases.items():
        if keep:
            url = tempo + '/api/traces/' + identifiers[name]
            raw = wait(name, lambda: stored(url))
            if any(key in raw for key in (b'http.client_ip', b'203.0.113.77', b'user_agent.original', b'private-agent-marker')):
                raise RuntimeError('private span attribute stored: ' + name)
    for name, (keep, _, _) in cases.items():
        if not keep and fetch(tempo + '/api/traces/' + identifiers[name])[0] != 404:
            raise RuntimeError('expected trace to be sampled out: ' + name)
    print('PASS 0% baseline: fast rejections dropped; 5xx, infrastructure RPC errors, mixed and slow traces retained; private attributes removed', flush=True)


def main():
    # The fixture only needs image pins and the stop timeout; no project credentials.
    services = json.loads(run('docker', 'compose', '--env-file', '/dev/null', '-f', str(DEPLOY / 'compose/compose.yaml'),
                              '--profile', 'observability', 'config', '--no-interpolate', '--no-normalize',
                              '--format', 'json'))['services']
    project = 'goseedobs' + uuid.uuid4().hex[:12]
    volume = project + '_traces'
    with tempfile.TemporaryDirectory() as directory:
        tmp = Path(directory)
        bridge = tmp / 'input.alloy'
        bridge.write_text(BRIDGE)
        compose_file = tmp / 'compose.json'
        compose_file.write_text(json.dumps({'services': {
            'tempo': {'image': services['tempo']['image'], 'ports': ['127.0.0.1::3200'],
                      'stop_grace_period': services['tempo']['stop_grace_period'],
                      'command': ['-config.file=/etc/tempo.yaml'],
                      'volumes': [f'{DEPLOY}/observability/tempo.yaml:/etc/tempo.yaml:ro', 'traces:/var/tempo']},
            'alloy': {'image': services['alloy']['image'], 'profiles': ['observability'], 'depends_on': ['tempo'],
                      'command': ['run', '/etc/alloy/test'], 'ports': ['127.0.0.1::4318'],
                      'environment': {'TRACE_BASELINE_PERCENT': '0'},
                      'volumes': [f'{DEPLOY}/observability/alloy.alloy:/etc/alloy/test/traces.alloy:ro',
                                  f'{bridge}:/etc/alloy/test/input.alloy:ro']},
        }, 'volumes': {'traces': {}}}))
        dc = ['docker', 'compose', '-p', project, '-f', str(compose_file)]
        try:
            run(*dc, '--profile', 'observability', 'up', '-d')
            tempo = 'http://' + run(*dc, 'port', 'tempo', '3200')
            otlp = 'http://' + run(*dc, '--profile', 'observability', 'port', 'alloy', '4318')
            wait('Tempo readiness', lambda: fetch(tempo + '/ready')[0] == 200, 120)
            test_sampling(otlp, tempo)
            run('make', 'compose-down', 'COMPOSE=' + shlex.join(dc))
            for kind in ('container', 'network'):
                if run('docker', kind, 'ls', *(['-a'] if kind == 'container' else []), '-q',
                       '--filter', 'label=com.docker.compose.project=' + project):
                    raise RuntimeError('compose-down left a ' + kind)
            run('docker', 'volume', 'inspect', '--format', '{{.Name}}', volume)
            print('PASS compose-down removed profile containers and network, preserved named volume', flush=True)
        finally:
            # Only the randomly named fixture project owns this volume.
            run(*dc, '--profile', '*', 'down', '--volumes')


if __name__ == '__main__':
    main()
