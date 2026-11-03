#!/usr/bin/env python3
"""Exercise the local Grafana/Loki/Tempo stack using only loopback HTTP."""
import argparse
import json
import re
import subprocess
import sys
import time
import uuid
from pathlib import Path
from urllib import error, parse, request

from env import environment


ROOT = Path(__file__).resolve().parents[1]
SELECTOR = '{compose_project="goseed"}'
COMPOSE = ['docker', 'compose', '--env-file', '.env', '-p', 'goseed',
           '-f', 'deploy/compose/compose.yaml', '--profile', 'observability']


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def wait(label, probe, timeout=60):
    deadline = time.monotonic() + timeout
    while True:
        try:
            result = probe()
            if result:
                return result
        except (error.URLError, TimeoutError, ConnectionError):
            pass
        if time.monotonic() >= deadline:
            raise RuntimeError('timed out: ' + label)
        time.sleep(1)


def entries(sequence, name):
    """Decode an OTLP attribute without coercing or printing the entire span."""
    return next((entry['value'].get('stringValue') for entry in sequence
                 if entry['key'] == name), None)


def spans(data):
    for batch in data.get('batches', data.get('resourceSpans', [])):
        service = entries(batch.get('resource', {}).get('attributes', []), 'service.name')
        for scope in batch.get('scopeSpans', batch.get('instrumentationLibrarySpans', [])):
            for span in scope.get('spans', []):
                yield service, span


class Smoke:
    def __init__(self):
        env = environment()
        self.base = {name: 'http://127.0.0.1:' + env[key] for name, key in (
            ('gateway', 'GATEWAY_PORT'), ('admin', 'ADMIN_API_PORT'), ('app', 'APP_API_PORT'),
            ('tempo', 'TEMPO_PORT'), ('loki', 'LOKI_PORT'), ('grafana', 'GRAFANA_PORT'))}
        self.http = request.build_opener(request.ProxyHandler({}))
        self.start = int(time.time()) - 1

    def fetch(self, service, path, *, query=None, body=None, headers=None):
        url = self.base[service] + path
        if query:
            url += '?' + parse.urlencode(query)
        hdr = {'Accept': 'application/json', **(headers or {})}
        if body is not None:
            hdr.update({'Content-Type': 'application/json', 'X-GoSeed-CSRF': '1'})
        req = request.Request(url, headers=hdr, data=None if body is None else json.dumps(body).encode())
        try:
            res = self.http.open(req, timeout=5)
        except error.HTTPError as exc:
            res = exc
        with res:
            return res.status, res.read()

    def call(self, path, expected, **kwargs):
        status, body = self.fetch('gateway', path, **kwargs)
        require(status in expected, f'{path} returned HTTP {status}, expected {expected}')
        trace_id = json.loads(body).get('traceId', '')
        require(re.fullmatch('[a-f0-9]{32}', trace_id), path + ' has no valid trace ID')
        return trace_id

    def trace(self, trace_id, predicate):
        def read():
            status, raw = self.fetch('tempo', '/api/traces/' + trace_id)
            if status != 200:
                return None
            data = json.loads(raw)
            return data if predicate(list(spans(data))) else None
        return wait('Tempo /api/traces/' + trace_id + ' expected spans', read)

    def wait_for_delivery(self):
        # Exporters reconnect asynchronously. A request sent before reconnection
        # can be lost; verify a fresh ping rather than waiting on a lost span.
        pending = []
        next_ping = 0.0

        def delivered():
            nonlocal next_ping
            if time.monotonic() >= next_ping:
                pending.append(self.call('/admin-api/v1/system/ping', {200}))
                next_ping = time.monotonic() + 5
            for trace_id in reversed(pending):
                status, raw = self.fetch('tempo', '/api/traces/' + trace_id)
                if status == 200 and {'admin-api', 'system-rpc'} <= {
                    service for service, _ in spans(json.loads(raw))
                }:
                    return True
            return False

        wait('OTLP delivery: fresh ping with admin-api and system-rpc spans', delivered)

    def logs(self, text):
        query = SELECTOR + ' |= ' + json.dumps(text)
        status, raw = self.fetch('loki', '/loki/api/v1/query_range', query={
            'query': query, 'start': self.start * 10**9, 'end': time.time_ns(), 'limit': 500})
        require(status == 200, f'Loki query_range HTTP {status}, selector {SELECTOR}')
        data = json.loads(raw)
        require(data.get('status') == 'success', 'Loki query_range failed: ' + SELECTOR)
        return [(stream['stream'].get('compose_service'), line)
                for stream in data['data']['result'] for _, line in stream['values']]

    def access(self, trace_id, path, status):
        def find():
            api = gateway = False
            for service, line in self.logs(trace_id):
                if service not in ('admin-api', 'gateway'):
                    continue
                row = json.loads(line)
                api |= (service == 'admin-api' and row.get('content') == 'http access'
                        and row.get('trace') == trace_id and row.get('route') == path
                        and row.get('status') == status and (status < 500 or row.get('level') == 'error'))
                gateway |= (service == 'gateway' and row.get('uri') == path and row.get('status') == status
                            and trace_id in row.get('traceparent', ''))
            return api and gateway
        wait(f'Loki {SELECTOR} |= "{trace_id}": {path} status {status}, admin-api and gateway', find)

    def run(self):
        checks = [('loki', '/ready'), ('tempo', '/ready'), ('grafana', '/api/health')]
        wait('Loki, Tempo and Grafana readiness', lambda: all(
            self.fetch(service, path)[0] == 200 for service, path in checks), 120)
        self.wait_for_delivery()
        password = 'OBS_PASSWORD_' + uuid.uuid4().hex
        cookie = 'OBS_COOKIE_' + uuid.uuid4().hex
        token = 'OBS_TOKEN_' + uuid.uuid4().hex
        login = '/admin-api/v1/auth/login'
        login_id = self.call(login, {401}, body={'username': 'obs_' + uuid.uuid4().hex[:16], 'password': password},
                             headers={'Cookie': 'fixture=' + cookie, 'Authorization': 'Bearer ' + token})
        data = self.trace(login_id, lambda rows: {'admin-api', 'system-rpc'} <= {svc for svc, _ in rows})
        self.access(login_id, login, 401)
        print('PASS failed login: API and RPC spans, matching API and gateway logs. Trace:', login_id, flush=True)

        keyword = 'OBS_QUERY_' + uuid.uuid4().hex
        forged = uuid.uuid4().hex
        baggage = 'OBS_BAGGAGE_' + uuid.uuid4().hex
        path = '/admin-api/v1/system/settings'
        actual = self.call(path, {401}, query={'keyword': keyword, 'page': 1}, headers={
            'traceparent': f'00-{forged}-bbbbbbbbbbbbbbbb-01', 'baggage': 'tenant=' + baggage})
        require(actual != forged, 'external trace ID accepted')
        query_trace = self.trace(actual, lambda rows: any(svc == 'admin-api' for svc, _ in rows))
        self.access(actual, path, 401)
        targets = [entries(s.get('attributes', []), 'http.target') for _, s in spans(query_trace)]
        require(path in targets and not any('?' in target for target in targets if target), 'unexpected http.target')
        for marker in [password, cookie, token, keyword, baggage]:
            require(marker not in json.dumps([data, query_trace]) and not self.logs(marker),
                    'sensitive test marker found in Tempo or Loki')
        require(self.fetch('tempo', '/api/traces/' + forged)[0] == 404, 'forged trace ID is stored')
        print('PASS query, password, token, cookie, baggage and forged trace checks', flush=True)

        for svc in ('admin', 'app'):
            require(self.fetch(svc, '/readyz')[0] == 200, svc + ' readiness failed')
        # Allow the SDK batch exporter and the sampling decision to complete.
        time.sleep(13)
        expr = '{ span:name =~ ".*(CheckReady|grpc[.]health[.]v1[.]Health/Check)" }'
        status, raw = self.fetch('tempo', '/api/search', query={
            'q': expr, 'start': self.start, 'end': int(time.time())})
        require(status == 200 and json.loads(raw).get('traces', []) == [], 'probe filter: ' + expr)
        print('PASS probe spans filtered', flush=True)

    def fault(self):
        def compose(action, service):
            subprocess.run(COMPOSE + [action, service], cwd=ROOT, check=True)

        path = '/admin-api/v1/system/ping'
        try:
            compose('stop', 'system-rpc')
            status, raw = self.fetch('gateway', path)
            require(500 <= status < 600, f'{path} outage returned HTTP {status}')
            trace_id = json.loads(raw)['traceId']
            self.trace(trace_id, lambda rows: any(s.get('status', {}).get('code') in
                       (2, 'STATUS_CODE_ERROR') for _, s in rows))
            self.access(trace_id, path, status)
        finally:
            compose('start', 'system-rpc')
            wait('system-rpc recovery', lambda: self.fetch('admin', '/readyz')[0] == 200)
        try:
            compose('stop', 'alloy')
            self.call(path, {200})
        finally:
            compose('start', 'alloy')
            self.wait_for_delivery()
        print('PASS system-rpc outage, error trace and access log; Alloy outage and recovery', flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--fault', action='store_true')
    args = parser.parse_args()
    try:
        smoke = Smoke()
        smoke.run()
        if args.fault:
            smoke.fault()
    except (RuntimeError, OSError, ValueError, KeyError, subprocess.CalledProcessError) as exc:
        sys.exit(f'Observability smoke failed: {exc}')
