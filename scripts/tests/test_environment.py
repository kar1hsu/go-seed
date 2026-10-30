import base64
import importlib.util
import json
import os
from pathlib import Path
import stat
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / 'scripts'))
from env import service_environment
spec = importlib.util.spec_from_file_location('init_env', ROOT / 'scripts/init-env.py')
init_env = importlib.util.module_from_spec(spec)
spec.loader.exec_module(init_env)

class EnvironmentTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.tool_dir = tempfile.TemporaryDirectory()
        cls.keygen = str(Path(cls.tool_dir.name) / 'keygen')
        subprocess.run(['go', 'build', '-o', cls.keygen, './cmd/keygen'], cwd=ROOT, check=True)

    @classmethod
    def tearDownClass(cls):
        cls.tool_dir.cleanup()

    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        (self.root / '.env.example').write_text((ROOT / '.env.example').read_text())
        command = patch.object(init_env, 'key_command', return_value=[self.keygen])
        command.start()
        self.addCleanup(command.stop)

    def test_initialization_preserves_keys_passwords_and_custom_values(self):
        path = self.root / '.env'
        path.write_text('MYSQL_PASSWORD="keep this password"\nADMIN_BOOTSTRAP_PASSWORD=unchanged\nTZ=UTC\nCUSTOM_VALUE=keep\nADMIN_TOKEN_SECRET=legacy-secret\n')
        init_env.initialize(self.root)
        first = path.read_text()
        values = init_env.entries(first)
        self.assertEqual(values['MYSQL_PASSWORD'], 'keep this password')
        self.assertEqual(values['ADMIN_BOOTSTRAP_PASSWORD'], 'unchanged')
        self.assertEqual(values['TZ'], 'UTC')
        self.assertEqual(values['CUSTOM_VALUE'], 'keep')
        self.assertEqual(values['ADMIN_TOKEN_SECRET'], 'legacy-secret')
        self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)
        self.assertNotEqual(values['ADMIN_TOKEN_PUBLIC_KEY'], values['MEMBER_TOKEN_PUBLIC_KEY'])
        self.assertNotEqual(values['ADMIN_REFRESH_KEY'], values['MEMBER_REFRESH_KEY'])
        for domain in ('ADMIN', 'MEMBER'):
            for suffix, size in [('TOKEN_PRIVATE_KEY', 64), ('TOKEN_PUBLIC_KEY', 32), ('REFRESH_KEY', 32)]:
                value = values[f'{domain}_{suffix}']
                self.assertEqual(len(base64.urlsafe_b64decode(value + '=' * (-len(value) % 4))), size)
        init_env.initialize(self.root)
        self.assertEqual(path.read_text(), first)

    def test_corrupt_existing_pair_does_not_rewrite_env(self):
        init_env.initialize(self.root)
        path = self.root / '.env'
        values = init_env.entries(path.read_text())
        broken = path.read_text().replace(values['ADMIN_TOKEN_PUBLIC_KEY'], values['MEMBER_TOKEN_PUBLIC_KEY'])
        path.write_text(broken)
        with self.assertRaises(SystemExit):
            init_env.initialize(self.root)
        self.assertEqual(path.read_text(), broken)

    def test_private_material_only_reaches_its_issuer(self):
        init_env.initialize(self.root)
        source = {**os.environ, **init_env.entries((self.root / '.env').read_text()),
                  'ADMIN_TOKEN_SECRET': 'old', 'MEMBER_TOKEN_SECRET': 'old', 'UNRELATED_SECRET': 'hidden'}
        allowed = {
            'system-rpc': {'ADMIN_TOKEN_PUBLIC_KEY', 'ADMIN_TOKEN_PRIVATE_KEY', 'ADMIN_REFRESH_KEY'},
            'member-rpc': {'MEMBER_TOKEN_PUBLIC_KEY', 'MEMBER_TOKEN_PRIVATE_KEY', 'MEMBER_REFRESH_KEY'},
            'admin-api': {'ADMIN_TOKEN_PUBLIC_KEY'},
            'app-api': {'MEMBER_TOKEN_PUBLIC_KEY'},
            'system-worker': set(), 'system-scheduler': set(), 'web': set(),
            'system-command': set(), 'member-command': set(),
        }
        credentials = {key for key in source if '_TOKEN_' in key or key.endswith('_REFRESH_KEY')}
        for service, expected in allowed.items():
            with self.subTest(service=service):
                actual = service_environment(source, service)
                self.assertEqual(set(actual) & credentials, expected)
                self.assertNotIn('UNRELATED_SECRET', actual)
                self.assertNotIn('ADMIN_BOOTSTRAP_PASSWORD', actual)
                self.assertNotIn('MYSQL_ROOT_PASSWORD', actual)
                self.assertEqual(actual['TZ'], source['TZ'])
                self.assertIn('PATH', actual)
        for command, domain in [('system-command', 'SYSTEM'), ('member-command', 'MEMBER')]:
            scoped = service_environment(source, command)
            # The config loader interpolates the whole YAML before selecting
            # MySQL fields; missing address variables previously produced ':'.
            for field in (domain + '_RPC_HOST', domain + '_RPC_PORT', 'REDIS_HOST', 'REDIS_PORT'):
                self.assertTrue(scoped.get(field), command + ' is missing ' + field)
            self.assertFalse(any('SECRET' in key for key in scoped))
        web = service_environment(source, 'web')
        self.assertFalse(any('PASSWORD' in key or 'SECRET' in key or key.endswith('_KEY') for key in web))

    def test_cross_domain_public_key_cannot_be_used_for_refresh(self):
        init_env.initialize(self.root)
        path = self.root / '.env'
        values = init_env.entries(path.read_text())
        broken = path.read_text().replace(values['ADMIN_REFRESH_KEY'], values['MEMBER_TOKEN_PUBLIC_KEY'])
        path.write_text(broken)
        with self.assertRaises(SystemExit):
            init_env.initialize(self.root)
        self.assertEqual(path.read_text(), broken)

    def test_compose_key_distribution_matches_local_processes(self):
        # Structural check uses Compose's actual YAML merger/interpolator, without
        # starting containers and without printing secrets to the test output.
        init_env.initialize(self.root)
        env_file = str(self.root / '.env')
        clean = {key: value for key, value in os.environ.items() if key not in init_env.entries((ROOT / '.env.example').read_text())}
        output = subprocess.check_output(['docker', 'compose', '--env-file', env_file, '-f', str(ROOT / 'deploy/compose/compose.yaml'), 'config', '--format', 'json'], env=clean, text=True)
        services = json.loads(output)['services']
        source = init_env.entries((self.root / '.env').read_text())
        for service in ('system-rpc', 'member-rpc', 'admin-api', 'app-api', 'system-worker', 'system-scheduler'):
            actual = services[service]['environment']
            expected = service_environment(source, service)
            keys = lambda env: {key for key in env if '_TOKEN_' in key or key.endswith('_REFRESH_KEY')}
            self.assertEqual(keys(actual), keys(expected))
