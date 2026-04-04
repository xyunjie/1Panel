# OpenResty 安装方案设计

## 问题背景

应用商店已被移除，但网站管理功能的整个链路都依赖 `getAppInstallByKey("openresty")` 从 `apps` + `app_installs` 表中查找 OpenResty 安装记录，获取：
- 容器名称（用于 `docker exec nginx -t/-s reload`）
- 安装路径（用于读写 nginx 配置文件）
- HTTP/HTTPS 端口
- docker-compose 路径（用于容器启停）

当前前端 `app-status` 组件会检查应用安装状态，如果未安装则显示安装引导。但应用商店移除后，安装流程不再可用。

## 影响范围

依赖 `getAppInstallByKey("openresty")` 的文件（agent 端）：
- `nginx.go` / `nginx_utils.go` — 所有 nginx 配置操作
- `website.go` — 创建网站
- `website_utils.go` — 网站配置、WAF、路径计算
- `website_ssl.go` — SSL 证书
- `website_domain.go` — 域名管理
- `website_proxy.go` / `website_lb.go` — 反代和负载均衡
- `website_rewrite.go` — URL 重写
- `backup_website.go` — 网站备份
- `snapshot_*.go` — 快照
- `cronjob_*.go` — 定时任务
- `ai.go` / `mcp_server.go` — AI 和 MCP

## 方案选择

### 方案 A：在网站页面中内置 OpenResty 安装入口（推荐）

在网站管理页面中，当检测到 OpenResty 未安装时，显示一个安装引导页面，通过内置的 docker-compose 模板直接安装 OpenResty，不依赖应用商店。

**前端改动：**
1. 网站页面（`website/index.vue`）检测 OpenResty 安装状态
2. 未安装时显示安装引导 UI（选择端口、确认安装）
3. 已安装时正常显示网站列表

**后端改动：**
1. 新增 API：检查 OpenResty 安装状态
2. 新增 API：安装 OpenResty（使用内置 docker-compose 模板）
3. 安装完成后在 `apps` + `app_installs` 表中创建记录，保持与现有代码的兼容
4. 内置 OpenResty 的 docker-compose.yml 和 .env 模板

**优点：**
- 改动最小，现有的 `getAppInstallByKey` 链路完全不需要改
- 用户体验好，进入网站页面就能看到安装引导
- 安装参数（端口等）可控

**缺点：**
- 仍然依赖 `apps` + `app_installs` 表结构

### 方案 B：抽象 OpenResty 为独立的基础设施组件

将 OpenResty 从"应用"概念中剥离，作为系统基础设施组件独立管理，有自己的配置表和管理逻辑。

**优点：**
- 架构更清晰，OpenResty 不再伪装成"应用"

**缺点：**
- 改动巨大，需要重构所有 `getAppInstallByKey` 调用点
- 需要新建数据库表、迁移逻辑
- 风险高

### 方案 C：保留 app_installs 记录，但安装流程内置化

与方案 A 类似，但更轻量：
- 不新增专门的安装 API
- 在 agent 启动时或首次访问网站功能时，自动检测 Docker 中是否有 OpenResty 容器
- 如果有，自动创建/更新 app_installs 记录
- 如果没有，前端引导用户安装

## 推荐方案：方案 A

方案 A 是最务实的选择。具体实施步骤：

### Step 1：内置 OpenResty docker-compose 模板
在 agent 中内置 OpenResty 的 docker-compose.yml 和默认配置，不依赖应用商店的远程下载。

### Step 2：新增 OpenResty 管理 API
- `GET /api/v2/openresty/status` — 检查安装状态
- `POST /api/v2/openresty/install` — 安装 OpenResty（参数：HTTP 端口、HTTPS 端口）
- 安装逻辑：
  1. 渲染 docker-compose.yml 模板
  2. 创建必要的目录结构（conf、www、ssl 等）
  3. 执行 `docker compose up -d`
  4. 在 `apps` + `app_installs` 表中创建记录

### Step 3：前端网站页面增加安装检测
- 进入网站页面时调用 status API
- 未安装：显示安装引导（端口配置 + 安装按钮）
- 已安装：正常显示网站列表

### Step 4：保持现有链路不变
- `getAppInstallByKey("openresty")` 继续工作
- 所有网站/nginx 相关服务代码无需修改
