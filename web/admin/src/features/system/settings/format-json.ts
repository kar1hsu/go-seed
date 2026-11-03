// The API validates JSON. Change only structural whitespace so numbers, string
// escapes and key order survive exactly as stored.
export function formatSettingJson(value: string): string {
  let depth = 0
  const newline = () => '\n' + '  '.repeat(depth)
  return value.replace(
    /"(?:[^"\\]|\\.)*"|\{\s*\}|\[\s*\]|\s+|[{}[\],:]/g,
    (token) => {
      switch (token[0]) {
        case '"':
          return token
        case '{':
        case '[':
          if (token.length > 1) return token[0] === '{' ? '{}' : '[]'
          depth++
          return token + newline()
        case '}':
        case ']':
          depth--
          return newline() + token
        case ',':
          return ',' + newline()
        case ':':
          return ': '
        default:
          return ''
      }
    },
  )
}
