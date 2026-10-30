<script setup lang="ts">
import {
  ElDescriptions,
  ElDescriptionsItem,
  ElTag,
  ElTable,
  ElTableColumn,
} from 'element-plus'
import { project } from '../../../app/project'
const stack = Object.entries(project.dependencies).map(([name, version]) => ({
  name,
  version,
}))
</script>
<template>
  <div class="page-heading">
    <div>
      <h1>项目信息</h1>
      <p>项目版本、技术依赖与开发说明。</p>
    </div>
    <ElTag type="info" effect="plain">v{{ project.version }}</ElTag>
  </div>
  <section class="panel">
    <div class="panel-heading"><h2>基本信息</h2></div>
    <div class="panel-body">
      <ElDescriptions :column="1" border
        ><ElDescriptionsItem label="项目名称">GoSeed</ElDescriptionsItem
        ><ElDescriptionsItem label="管理端版本">{{
          project.version
        }}</ElDescriptionsItem
        ><ElDescriptionsItem label="项目仓库"
          ><a
            class="text-link"
            :href="project.repository"
            target="_blank"
            rel="noopener noreferrer"
            >{{ project.repository }}</a
          ></ElDescriptionsItem
        ><ElDescriptionsItem label="后端架构"
          >go-zero · API / RPC / Worker / Scheduler · GORM</ElDescriptionsItem
        ><ElDescriptionsItem label="当前阶段"
          >管理员认证、角色权限、操作日志、字典配置、私有文件与 Asynq 任务、C
          端会员业务样板已接入。</ElDescriptionsItem
        ></ElDescriptions
      >
    </div>
  </section>
  <div class="docs-grid">
    <section class="panel">
      <div class="panel-heading">
        <h2>前端依赖</h2>
        <span class="panel-caption">来自 package.json</span>
      </div>
      <div class="panel-table">
        <ElTable :data="stack"
          ><ElTableColumn
            prop="name"
            label="依赖"
            min-width="210" /><ElTableColumn
            prop="version"
            label="版本"
            min-width="100"
        /></ElTable>
      </div>
    </section>
    <section class="panel">
      <div class="panel-heading"><h2>常用命令</h2></div>
      <div class="panel-body">
        <pre class="code-block">
# 启动本地开发环境
make infra-up
make dev

# 检查、测试和构建
make check
make test
make build

# 生成接口与客户端类型
make gen</pre
        >
        <p class="body-description">
          运行前先按仓库 README 安装固定版本工具并配置 .env。
        </p>
      </div>
    </section>
  </div>
</template>
