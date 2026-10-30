import type { FormRules } from 'element-plus'
export const resourceRules: FormRules = {
  code: [
    {
      required: true,
      max: 64,
      pattern: /^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$/,
      message: '最长 64 位小写标识，支持字母、数字、下划线及点分段',
      trigger: 'blur',
    },
  ],
  name: [
    {
      required: true,
      whitespace: true,
      max: 64,
      message: '请输入 1–64 字的名称',
      trigger: 'blur',
    },
  ],
  description: [{ max: 256, message: '说明最多 256 字', trigger: 'blur' }],
}
export function validateSettingValue(type: string, value: string): string {
  if (new TextEncoder().encode(value).length > 4096)
    return '配置值不能超过 4096 字节'
  if (type === 'integer') {
    try {
      const number = BigInt(value)
      if (
        String(number) !== value ||
        number < -(2n ** 63n) ||
        number > 2n ** 63n - 1n
      )
        return '请输入 64 位有符号整数'
    } catch {
      return '请输入 64 位有符号整数'
    }
  }
  if (type === 'boolean' && value !== 'true' && value !== 'false')
    return '请选择布尔值'
  if (type === 'json') {
    try {
      JSON.parse(value)
    } catch {
      return '请输入有效的 JSON'
    }
  }
  return ''
}
