import { api } from '../../../lib/http'
import { unwrap, type Query } from '../api'
import type { components } from '../../../api/generated/schema'
type Schemas = components['schemas']
export type Dictionary = Schemas['DictionaryRecord']
export type DictionaryItem = Schemas['DictionaryItem']
export type Setting = Schemas['SettingRecord']
const header = { 'X-GoSeed-CSRF': '1' as const }
export const settings = {
  dictionary: (id: string) =>
    unwrap(
      api.GET('/admin-api/v1/system/dictionaries/{id}', {
        params: { path: { id } },
      }),
    ),
  dictionaries: (query: Query) =>
    unwrap(api.GET('/admin-api/v1/system/dictionaries', { params: { query } })),
  createDictionary: (body: Schemas['CreateDictionaryRequest']) =>
    unwrap(
      api.POST('/admin-api/v1/system/dictionaries', {
        params: { header },
        body,
      }),
    ),
  updateDictionary: (id: string, body: Schemas['UpdateDictionaryRequest']) =>
    unwrap(
      api.PUT('/admin-api/v1/system/dictionaries/{id}', {
        params: { path: { id }, header },
        body,
      }),
    ),
  deleteDictionary: (id: string, revision: string) =>
    unwrap(
      api.DELETE('/admin-api/v1/system/dictionaries/{id}', {
        params: { path: { id }, query: { revision }, header },
      }),
    ),
  settings: (query: Query) =>
    unwrap(api.GET('/admin-api/v1/system/settings', { params: { query } })),
  createSetting: (body: Schemas['CreateSettingRequest']) =>
    unwrap(
      api.POST('/admin-api/v1/system/settings', { params: { header }, body }),
    ),
  updateSetting: (id: string, body: Schemas['UpdateSettingRequest']) =>
    unwrap(
      api.PUT('/admin-api/v1/system/settings/{id}', {
        params: { path: { id }, header },
        body,
      }),
    ),
  deleteSetting: (id: string, revision: string) =>
    unwrap(
      api.DELETE('/admin-api/v1/system/settings/{id}', {
        params: { path: { id }, query: { revision }, header },
      }),
    ),
}
