<div align="center">
  <img src="assets/logo.svg" alt="TokenRouter" width="112" />

  <h1>TokenRouter</h1>

  <p>自托管的 AI API 网关，带用户、计费和运营后台</p>

  <p>
    <a href="https://github.com/TokenFlux/TokenRouter/actions/workflows/backend-ci.yml"><img src="https://github.com/TokenFlux/TokenRouter/actions/workflows/backend-ci.yml/badge.svg" alt="CI" /></a>
    <a href="https://github.com/TokenFlux/TokenRouter/releases"><img src="https://img.shields.io/github/v/release/TokenFlux/TokenRouter?display_name=tag" alt="Release" /></a>
    <a href="https://github.com/TokenFlux/TokenRouter/pkgs/container/tokenrouter"><img src="https://img.shields.io/badge/container-ghcr.io%2Ftokenflux%2Ftokenrouter-2496ED?logo=docker&logoColor=white" alt="Container" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-LGPL--3.0--or--later-4c1.svg" alt="License: LGPL-3.0-or-later" /></a>
  </p>

  <p><strong>简体中文</strong> | <a href="README_EN.md">English</a></p>
</div>

## 简介

TokenRouter 接入多家上游的模型服务，提供统一入口。管理员录入上游提供商（OAuth 账号、API Key、云厂商凭据），按分组配置可用模型、协议、倍率和限额；用户通过平台签发的 API Key 调用 Anthropic、OpenAI 或 Gemini 格式的接口。TokenRouter 负责认证、选提供商、转换协议、失败时切换提供商，并按用量扣费。

服务端是 Go 程序，数据存 PostgreSQL，缓存、限流和并发计数用 Redis。Vue 3 编写的用户控制台和管理后台在发布构建时嵌进同一个二进制文件。

TokenRouter 从 [Sub2API](https://github.com/Wei-Shaw/sub2api) 分叉后持续开发，感谢上游项目和所有贡献者。

## 功能

网关：

- 客户端入口有 Anthropic Messages、OpenAI Responses（含 WebSocket）、Chat Completions、Embeddings、Images、Gemini v1beta 和 Grok 视频接口。分组允许的协议之间可以互相转换，一个分组里可以混用不同平台的提供商。
- 调度器按模型、协议能力、限流状态和粘性会话挑选提供商。上游返回可重试错误时切换下一个提供商，开始向客户端输出内容后不再切换。
- 每个 API Key 可以单独设置额度、并发、RPM 和模型重定向。复合 Key 绑定多个分组，用模型名前缀选组。
- 内容审核支持关键词和哈希规则、同步拦截和自动封禁。

用户和计费：

- 登录方式有邮箱密码、Passkey、TOTP 双因素认证，以及 GitHub、Google、LinuxDo、OIDC、微信和钉钉。
- 团队可以共享 Key 和额度，成员有各自的角色。
- 计费按价格配置结算，用户可以用余额、订阅套餐、额度包和兑换码，邀请返利也算在内。
- 内置支付，支持易支付、支付宝官方、微信支付官方、Stripe 和 Airwallex。
- 创作台可以直接在网页上生成和编辑图片。

运营：

- 管理后台汇总每次请求的使用记录（模型链、Token 用量、费用和诊断信息），支持按用户、分组、提供商和模型查看。
- Ops 面板采集请求错误、上游错误、提供商可用性和主机指标，支持配置告警规则和邮件日报、周报。
- 管理后台支持在线检查新版本、升级和回滚。

## 支持的上游

| 平台 | 接入方式 |
| --- | --- |
| Anthropic | OAuth、Setup Token、API Key、AWS Bedrock、Vertex AI |
| OpenAI | OAuth、API Key |
| Gemini | OAuth、API Key、Vertex AI |
| Antigravity | OAuth |
| Grok / xAI | OAuth、API Key |
| Qoder | Qoder COSY |
| Kimi | API Key（按量付费、Coding 套餐） |
| 智谱 | API Key（按量付费、Coding 套餐） |
| DeepSeek | API Key（按量付费） |

各平台支持的入口及仅限导入的组合，见[上游提供商能力矩阵](docs/interfaces/upstream_provider_matrix.md)。

## 快速开始

用 Docker Compose 部署时，需要 Docker 20.10 和 Docker Compose v2 或更高版本：

```bash
mkdir -p tokenrouter-deploy && cd tokenrouter-deploy

# 下载 Compose 文件和 .env，并生成数据库密码和密钥
curl -sSL https://raw.githubusercontent.com/TokenFlux/TokenRouter/main/deploy/docker-deploy.sh | bash

docker compose up -d
```

启动后打开 `http://localhost:8080`。没有在 `.env` 里设置 `ADMIN_PASSWORD` 时，管理员密码是随机生成的，可以从日志里找到：

```bash
docker compose logs tokenrouter | grep "admin password"
```

Compose 默认只监听本机地址。对外提供服务时，在前面加反向代理，或者修改 `.env` 里的 `BIND_HOST`。

其他部署方式：

- [安装脚本](docs/guides/deployment/index.md#脚本安装)：在 Linux 上安装二进制并注册 systemd 服务，需要自备 PostgreSQL 15+ 和 Redis 7+。
- [单独的应用容器](deploy/DOCKER.md#独立容器)：连接已有的 PostgreSQL 和 Redis。
- [Apple container](docs/guides/deployment/apple_container.md)：在 macOS 上本地运行。

从 Sub2API 升级时，先看[部署指南](docs/guides/deployment/index.md#从旧名称部署升级)里的兼容说明，直接套用新的 Compose 模板会建出空数据卷。

## 本地开发

后端需要 Go 1.27，前端需要 Node.js 20 和 pnpm 9。

```bash
# 后端
cd backend
go run ./cmd/server

# 前端，开发服务器把 API 请求代理到后端
cd frontend
pnpm install --frozen-lockfile
pnpm run dev

# 或者用源码构建的完整环境
docker compose -f deploy/docker-compose.dev.yml up --build
```

`make build` 编译前后端，`make test` 运行后端测试、前端 lint、类型检查和关键测试。测试分层、生成代码和提交前的格式化要求见[开发、验证与上游同步](docs/operations/development_workflow.md)。

## 文档

- [使用与运维指南](docs/guides/index.md)：部署、支付配置和外部系统集成。
- [工程文档](docs/index.md)：架构、业务规则、接口和运维约束，修改代码前从这里查。
- [升级说明](docs/operations/upgrade_notes.md)：各版本需要注意的变化。

## 许可证

本项目依据 [GNU Lesser General Public License v3.0 或更高版本](LICENSE) 发布。

Copyright (c) 2026 Wesley Liddick & TokenFlux
