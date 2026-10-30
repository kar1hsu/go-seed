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
    # Match Compose: an explicitly exported TZ overrides the .env value.
    env['TZ'] = os.environ.get('TZ', env['TZ'])
    return env

# Keep host execution facilities, not arbitrary application credentials. In
# particular, dotenv/private keys must never flow into Vite's process environment.
_HOST_KEYS = frozenset('PATH HOME USER LOGNAME SHELL TMPDIR TMP TEMP LANG LC_ALL LC_CTYPE TERM SYSTEMROOT COMSPEC USERPROFILE LOCALAPPDATA APPDATA PATHEXT GOCACHE GOMODCACHE GOPATH GOTOOLCHAIN GOPROXY GOSUMDB GOPRIVATE GONOSUMDB GONOPROXY GOFLAGS CGO_ENABLED CC CXX SDKROOT DEVELOPER_DIR SSL_CERT_FILE SSL_CERT_DIR HTTP_PROXY HTTPS_PROXY ALL_PROXY NO_PROXY http_proxy https_proxy all_proxy no_proxy PNPM_HOME COREPACK_HOME NVM_BIN'.split())
_CONFIGS = {
    'admin-api': 'apps/admin/api/etc/admin-api.yaml',
    'app-api': 'apps/app/api/etc/app-api.yaml',
    'system-rpc': 'services/system/rpc/etc/system.yaml',
    'member-rpc': 'services/member/rpc/etc/member.yaml',
    'system-worker': 'services/system/worker/etc/system-worker.yaml',
    'system-scheduler': 'services/system/scheduler/etc/system-scheduler.yaml',
}

def service_environment(source, service, extra=()):
    """Pass only variables actually referenced by this process's configuration."""
    root = Path(__file__).resolve().parents[1]
    if service == 'web':
        required = {'WEB_PORT', 'ADMIN_API_PORT', 'APP_API_PORT'}
    else:
        config_service = service.replace('-command', '-rpc')
        required = set(re.findall(r'\$\{([A-Z][A-Z0-9_]*)\}', (root / _CONFIGS[config_service]).read_text()))
        if service.endswith('-command'):
            # The shared YAML is interpolated before decoding the MySQL-only
            # command config. Keep public address fields so Host:Port remains
            # valid YAML, while withholding runtime signing/service secrets.
            password = 'MYSQL_PASSWORD' if service == 'system-command' else 'MEMBER_MYSQL_PASSWORD'
            required = {key for key in required if key == password or not any(part in key for part in ('_TOKEN_', '_REFRESH_KEY', 'SECRET', 'PASSWORD', '_KEY'))}
    allowed = _HOST_KEYS | required | {'TZ'} | set(extra)
    return {key: value for key, value in source.items() if key in allowed}
