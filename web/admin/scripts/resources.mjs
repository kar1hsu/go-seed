import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import YAML from 'yaml'
import prettier from 'prettier'

export async function loadResources(root) {
  const read = async (name) =>
    YAML.parse(await fs.readFile(path.join(root, name), 'utf8'))
  const permissions = await read('services/system/resources/permissions.yaml')
  const menus = await read('web/admin/src/router/menu.yaml')
  const endpoints = await read('apps/admin/api/resources/endpoints.yaml')
  const known = new Set()
  for (const p of permissions) {
    if (known.has(p.code) || !/^\w+\.\w+\.\w+$/.test(p.code))
      throw Error('Duplicate or invalid permission: ' + p.code)
    known.add(p.code)
  }
  for (const p of permissions)
    for (const dependency of p.requires || [])
      if (!known.has(dependency) || dependency === p.code)
        throw Error('Invalid permission dependency')
  const components = new Set([
    'dashboard',
    'monitor',
    'administrators',
    'roles',
    'apiDocs',
    'about',
    'security',
  ])
  const paths = new Set()
  for (const menu of menus) {
    if (
      paths.has(menu.path) ||
      !components.has(menu.component) ||
      (menu.permission && !known.has(menu.permission))
    )
      throw Error('Invalid menu: ' + menu.path)
    paths.add(menu.path)
  }
  const endpointMap = new Map()
  for (const entry of endpoints) {
    const key = entry.method + ' ' + entry.path
    if (
      endpointMap.has(key) ||
      (entry.permission && !known.has(entry.permission)) ||
      (!entry.permission &&
        !['public', 'cookie', 'session'].includes(entry.access))
    )
      throw Error('Invalid endpoint: ' + key)
    endpointMap.set(key, entry)
  }
  const swagger = JSON.parse(
    await fs.readFile(path.join(root, 'docs/openapi/admin.json'), 'utf8'),
  )
  const implemented = new Set()
  for (const [route, item] of Object.entries(swagger.paths))
    for (const method of ['get', 'post', 'put', 'patch', 'delete'])
      if (item[method]) {
        const key = method + ' ' + route
        if (!endpointMap.has(key))
          throw Error('Missing endpoint authorization declaration: ' + key)
        implemented.add(key)
      }
  for (const key of endpointMap.keys())
    if (!implemented.has(key))
      throw Error('Orphan endpoint declaration: ' + key)
  return { permissions, menus, endpointMap }
}
if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const root = fileURLToPath(new URL('../../../', import.meta.url))
  const { menus } = await loadResources(root)
  const target = path.join(root, 'web/admin/src/router/resources.generated.ts')
  const source = await prettier.format(
    '// Code generated from menu.yaml. DO NOT EDIT.\nexport const menuResources = ' +
      JSON.stringify(menus, null, 2) +
      ' as const\n',
    { ...(await prettier.resolveConfig(target)), parser: 'typescript' },
  )
  if (process.argv.includes('--write')) await fs.writeFile(target, source)
  else if ((await fs.readFile(target, 'utf8')) !== source)
    throw Error('Menu declarations are stale; run make gen-web')
  console.log(
    'Permission references, menu component allowlist and HTTP coverage passed.',
  )
}
