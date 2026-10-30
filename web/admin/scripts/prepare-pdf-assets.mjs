import { cpSync, mkdirSync, rmSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { createRequire } from 'node:module'

const require = createRequire(import.meta.url)
const source = dirname(require.resolve('pdfjs-dist/package.json'))
const destination = fileURLToPath(new URL('../public/pdfjs/', import.meta.url))
// This directory contains generated dependency assets only; never user files.
rmSync(destination, { recursive: true, force: true })
mkdirSync(destination, { recursive: true })
for (const folder of ['cmaps', 'standard_fonts', 'wasm', 'iccs'])
  cpSync(join(source, folder), join(destination, folder), { recursive: true })
cpSync(join(source, 'LICENSE'), join(destination, 'LICENSE'))
