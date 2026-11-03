import assert from 'node:assert/strict'
import { test } from 'node:test'
import { formatSettingJson } from '../src/features/system/settings/format-json.ts'

test('JSON preview indents nested values without changing their source literals', () => {
  const input = String.raw`{"id":9007199254740993123,"values":[ null, {"n":1.00e+30}, {}, [ ] ],"text":"a,  b: {\"x\"}\\"}`
  assert.equal(
    formatSettingJson(input),
    String.raw`{
  "id": 9007199254740993123,
  "values": [
    null,
    {
      "n": 1.00e+30
    },
    {},
    []
  ],
  "text": "a,  b: {\"x\"}\\"
}`,
  )
})

test('JSON preview also supports scalar settings and empty containers', () => {
  for (const value of ['null', 'false', '""', '"  布尔  "', '-0', '{}', '[]']) {
    assert.equal(formatSettingJson(value), value)
  }
})
