import {
  DataBoard,
  Monitor,
  Document,
  InfoFilled,
  Lock,
  User,
  Key,
} from '@element-plus/icons-vue'
import { menuResources } from './resources.generated'
const icons = { DataBoard, Monitor, Document, InfoFilled, Lock, User, Key }
export const navigation = menuResources.map((item) => ({
  ...item,
  icon: icons[item.icon],
  permission: 'permission' in item ? item.permission : undefined,
  hidden: 'hidden' in item && item.hidden,
}))
