export type CloseAction = 'close' | 'others' | 'right' | 'all'

// The workbench stays pinned. Closing a background tab never changes the active page.
export function closeTabs(
  paths: string[],
  active: string,
  target: string,
  action: CloseAction,
) {
  const index = paths.indexOf(target)
  if (index < 0) return { paths, active }
  const next = paths.filter(
    (path, i) =>
      path === '/' ||
      (action === 'close'
        ? path !== target
        : action === 'others'
          ? path === target
          : action === 'right'
            ? i <= index
            : false),
  )
  if (!next.includes('/')) next.unshift('/')
  const activeIndex = paths.indexOf(active)
  const adjacent = paths
    .slice(0, activeIndex)
    .reverse()
    .find((path) => next.includes(path))
  return {
    paths: next,
    active: next.includes(active)
      ? active
      : action === 'others'
        ? target
        : adjacent || next[0]!,
  }
}
