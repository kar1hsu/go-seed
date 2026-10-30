import fs from 'node:fs/promises'
import { loadResources } from './resources.mjs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import converter from 'swagger2openapi'
import openapiTS, { astToString } from 'openapi-typescript'
const root = fileURLToPath(new URL('../../../', import.meta.url))
const resources = await loadResources(root)
for (const app of ['admin', 'app']) {
  const endpoints =
    app === 'admin' ? resources.endpointMap : resources.appEndpointMap
  const bearer = app === 'admin' ? 'AdminBearer' : 'MemberBearer'
  const refresh = app === 'admin' ? 'AdminRefresh' : 'MemberRefresh'
  const cookie = app === 'admin' ? 'goseed_refresh' : 'goseed_member_refresh'
  const prefix = app === 'admin' ? '/admin-api/v1' : '/api/v1'
  const swagger = JSON.parse(
    await fs.readFile(path.join(root, `docs/openapi/${app}.json`), 'utf8'),
  )
  for (const item of Object.values(swagger.paths)) {
    for (const parameter of item.delete?.parameters || [])
      if (parameter.in === 'formData') parameter.in = 'query'
    if (item.delete) delete item.delete.consumes
  }
  delete swagger['x-date']
  await fs.writeFile(
    path.join(root, `docs/openapi/${app}.json`),
    JSON.stringify(swagger, null, 2) + '\n',
  )
  const { openapi } = await converter.convertObj(swagger, {})
  openapi.servers = [{ url: '/' }]
  openapi.components.schemas.ErrorResponse = {
    type: 'object',
    required: ['code', 'message', 'data', 'traceId'],
    properties: {
      code: { type: 'integer' },
      message: { type: 'string' },
      data: { nullable: true, enum: [null] },
      traceId: { type: 'string' },
    },
  }
  openapi.components.securitySchemes = {
    [bearer]: { type: 'http', scheme: 'bearer', bearerFormat: 'JWT' },
    [refresh]: { type: 'apiKey', in: 'cookie', name: cookie },
  }
  for (const [route, item] of Object.entries(openapi.paths)) {
    for (const method of ['get', 'post', 'put', 'patch', 'delete']) {
      const operation = item[method]
      if (!operation) continue
      const entry = endpoints.get(method + ' ' + route)
      if (entry.permission || entry.access === 'session')
        operation.security = [{ [bearer]: [] }]
      if (entry.access === 'cookie') operation.security = [{ [refresh]: [] }]
      if (entry.permission) operation['x-permission'] = entry.permission
      if (entry.requestBody === 'binary')
        operation.requestBody = {
          required: true,
          content: {
            'application/octet-stream': {
              schema: { type: 'string', format: 'binary' },
            },
          },
          description: 'Raw file body, at most 10 MiB. Create metadata first.',
        }
      if (entry.responseBody === 'binary')
        operation.responses['200'] = {
          description: 'Private attachment; no JSON envelope.',
          content: {
            'application/octet-stream': {
              schema: { type: 'string', format: 'binary' },
            },
          },
        }
      const success = entry.status || 200
      if (success !== 200) {
        operation.responses[success] = operation.responses['200']
        delete operation.responses['200']
      }
      if (method !== 'get') {
        operation.parameters ||= []
        operation.parameters.push({
          in: 'header',
          name: 'X-GoSeed-CSRF',
          required: true,
          schema: { type: 'string', enum: ['1'] },
          description:
            'Required for same-origin write requests, including CLI clients.',
        })
      }
      if ([prefix + '/auth/login', prefix + '/auth/refresh'].includes(route)) {
        operation.responses[success].headers = {
          'Set-Cookie': {
            schema: { type: 'string' },
            description: `Host-only ${cookie} cookie; HttpOnly; SameSite=Strict; Path=${prefix}/auth; Secure over HTTPS.`,
          },
        }
      }
      for (const code of [400, 401, 403, 404, 408, 409, 429, 500, 503, 504]) {
        operation.responses[code] = {
          description:
            'GoSeed error envelope; code is stable, message is safe for clients.',
          content: {
            'application/json': {
              schema: { $ref: '#/components/schemas/ErrorResponse' },
            },
          },
        }
      }
    }
  }
  await fs.writeFile(
    path.join(root, `docs/openapi/${app}.openapi.json`),
    JSON.stringify(openapi, null, 2) + '\n',
  )
  const ast = await openapiTS(openapi)
  await fs.writeFile(
    new URL(
      `../src/api/generated/${app === 'admin' ? 'schema' : 'app-schema'}.ts`,
      import.meta.url,
    ),
    `// Code generated from the ${app} .api contract. DO NOT EDIT.\n` +
      astToString(ast),
  )
}
