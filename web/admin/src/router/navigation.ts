import {
  Menu,
  DataBoard,
  Monitor,
  Document,
  InfoFilled,
  Lock,
  User,
  Key,
  Collection,
  Setting,
  FolderOpened,
  Timer,
} from '@element-plus/icons-vue'
import { menuResources } from './resources.generated'
export const icons = {
  Menu,
  DataBoard,
  Monitor,
  Document,
  InfoFilled,
  Lock,
  User,
  Key,
  Collection,
  Setting,
  FolderOpened,
  Timer,
}
export const navigation = menuResources.map((item) => ({
  ...item,
  iconName: item.icon,
  icon: icons[item.icon],
  permission: 'permission' in item ? item.permission : undefined,
  hidden: 'hidden' in item && item.hidden,
}))

export function menuIcon(name: string) {
  return Object.hasOwn(icons, name)
    ? icons[name as keyof typeof icons]
    : FolderOpened
}
