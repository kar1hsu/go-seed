import type { components } from '../api/generated/schema'
import { menuIcon, navigation } from './navigation'

type MenuTree = components['schemas']['MenuTree']
const defaultGroups = [...new Set(navigation.map((item) => item.group))].map(
  (title) => {
    const pages = navigation
      .filter((item) => item.group === title)
      .map((item) => ({
        path: item.path,
        title: item.title,
        icon: item.iconName,
        hidden: item.hidden,
      }))
    return {
      id: 'builtin-' + title,
      title,
      icon: pages[0]!.icon,
      hidden: false,
      pages,
    }
  },
)

// Bind display settings only to routes shipped with this build. Router guards
// and button permissions continue to use the compiled permission declarations.
export function presentationGroups(
  tree: MenuTree | null,
  has: (code?: string) => boolean,
) {
  return (tree?.groups ?? defaultGroups).map((group) => ({
    id: group.id,
    title: group.title,
    hidden: group.hidden,
    icon: menuIcon(group.icon),
    items: group.pages.flatMap((page) => {
      const route = navigation.find((item) => item.path === page.path)
      if (!route || !has(route.permission)) return []
      return [
        {
          ...route,
          title: page.title,
          group: group.title,
          icon: menuIcon(page.icon),
          hidden: page.hidden,
        },
      ]
    }),
  }))
}
export type PresentationGroup = ReturnType<typeof presentationGroups>[number]
