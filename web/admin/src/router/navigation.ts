import {
  DataBoard,
  Monitor,
  Document,
  InfoFilled,
} from '@element-plus/icons-vue'

export const navigation = [
  { path: '/', title: '工作台', icon: DataBoard, group: '概览' },
  { path: '/monitor', title: '服务状态', icon: Monitor, group: '概览' },
  { path: '/api-docs', title: '接口文档', icon: Document, group: '开发工具' },
  { path: '/about', title: '项目信息', icon: InfoFilled, group: '开发工具' },
]
