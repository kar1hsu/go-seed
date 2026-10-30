#!/usr/bin/env python3
"""Provision only the GoSeed development member database on our Compose MySQL."""
import re
import subprocess
from pathlib import Path
from env import environment
root = Path(__file__).resolve().parents[1]
env = environment()
database, user, password = (env[key] for key in ('MEMBER_MYSQL_DATABASE','MEMBER_MYSQL_USER','MEMBER_MYSQL_PASSWORD'))
if not re.fullmatch(r'goseed_[a-z0-9_]+', database) or not re.fullmatch(r'goseed_[a-z0-9_]+', user):
    raise SystemExit('Development database and user must use goseed_ identifiers')
if database == env['MYSQL_DATABASE'] or user == env['MYSQL_USER'] or len(password) < 20:
    raise SystemExit('Member requires its own database, user and password; run make init')
def quote(value):
    return "'" + value.replace('\\', '\\\\').replace("'", "''") + "'"
sql = f"CREATE DATABASE IF NOT EXISTS `{database}` CHARACTER SET utf8mb4;\nCREATE USER IF NOT EXISTS {quote(user)}@'%' IDENTIFIED BY {quote(password)};\nGRANT ALL PRIVILEGES ON `{database}`.* TO {quote(user)}@'%';\n"
env['MYSQL_PWD'] = env['MYSQL_ROOT_PASSWORD']
command = ['docker','compose','--env-file','.env','-p','goseed','-f','deploy/compose/compose.yaml','exec','-T','-e','MYSQL_PWD','mysql','mysql','-uroot']
result = subprocess.run(command, input=sql, text=True, cwd=root, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
if result.returncode:
    raise SystemExit('Member database provisioning failed; check the GoSeed MySQL instance and credentials')
print('Member development database and dedicated account are ready; existing passwords preserved.')
