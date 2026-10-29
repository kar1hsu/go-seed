import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import converter from 'swagger2openapi'
import openapiTS, { astToString } from 'openapi-typescript'

const root = fileURLToPath(new URL('../../../', import.meta.url))
for (const app of ['admin', 'app']) {
  const swagger = JSON.parse(await fs.readFile(path.join(root, `docs/openapi/${app}.json`), 'utf8'))
  // goctl embeds wall-clock generation time; omit it from reproducible artifacts.
  delete swagger['x-date']
  await fs.writeFile(path.join(root, `docs/openapi/${app}.json`), JSON.stringify(swagger, null, 2) + '\n')
  // goctl emits Swagger 2. Normalize it to OpenAPI 3, with the common error contract.
  const { openapi } = await converter.convertObj(swagger, {})
  openapi.servers = [{ url: '/' }]
  openapi.components.schemas.ErrorResponse = {
    type: 'object', required: ['code', 'message', 'data', 'traceId'],
    properties: {
      code: { type: 'integer' }, message: { type: 'string' },
      data: { nullable: true, enum: [null] }, traceId: { type: 'string' },
    },
  }
  for (const item of Object.values(openapi.paths)) {
    for (const method of ['get', 'post', 'put', 'patch', 'delete']) {
      const operation = item[method]
      if (!operation) continue
      for (const code of [400, 401, 403, 404, 408, 409, 429, 500, 503, 504]) {
        operation.responses[code] = {
          description: 'GoSeed error envelope; code is stable, message is safe for clients.',
          content: { 'application/json': { schema: { $ref: '#/components/schemas/ErrorResponse' } } },
        }
      }
    }
  }
  await fs.writeFile(path.join(root, `docs/openapi/${app}.openapi.json`), JSON.stringify(openapi, null, 2) + '\n')
  if (app === 'admin') {
    const ast = await openapiTS(openapi)
    await fs.writeFile(new URL('../src/api/generated/schema.ts', import.meta.url), '// Code generated from the admin .api contract. DO NOT EDIT.\n' + astToString(ast))
  }
}
