<div align="center">
  <img src="assets/logo.svg" alt="TokenRouter" width="112" />

  <h1>TokenRouter</h1>

  <p>A self-hosted AI API gateway with user management, billing, and an operations console</p>

  <p>
    <a href="https://github.com/TokenFlux/TokenRouter/actions/workflows/backend-ci.yml"><img src="https://github.com/TokenFlux/TokenRouter/actions/workflows/backend-ci.yml/badge.svg" alt="CI" /></a>
    <a href="https://github.com/TokenFlux/TokenRouter/releases"><img src="https://img.shields.io/github/v/release/TokenFlux/TokenRouter?display_name=tag" alt="Release" /></a>
    <a href="https://github.com/TokenFlux/TokenRouter/pkgs/container/tokenrouter"><img src="https://img.shields.io/badge/container-ghcr.io%2Ftokenflux%2Ftokenrouter-2496ED?logo=docker&logoColor=white" alt="Container" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-LGPL--3.0--or--later-4c1.svg" alt="License: LGPL-3.0-or-later" /></a>
  </p>

  <p><a href="README.md">简体中文</a> | <strong>English</strong></p>
</div>

## Overview

TokenRouter connects model services from many upstream vendors to a single endpoint. Administrators add upstream providers (OAuth accounts, API keys, cloud credentials) and configure groups that set the available models, protocols, price multipliers, and limits. Users call Anthropic, OpenAI, or Gemini style APIs with keys issued by TokenRouter. TokenRouter authenticates each request, picks a provider, converts the protocol when needed, switches to another provider when one fails, and charges for usage.

The server is a Go program. PostgreSQL stores the data. Redis holds caches, rate limits, and concurrency counters. Release builds embed the Vue 3 user console and admin dashboard in the same binary.

TokenRouter is a fork of [Sub2API](https://github.com/Wei-Shaw/sub2api) under continued development. Thanks to the upstream project and all contributors.

## Features

Gateway:

- Client endpoints cover Anthropic Messages, OpenAI Responses (including WebSocket), Chat Completions, Embeddings, Images, Gemini v1beta, and Grok video. Requests convert between the protocols a group allows, and one group mixes providers from different platforms.
- The scheduler picks a provider by model, protocol support, rate limit state, and sticky sessions. When an upstream returns a retryable error, TokenRouter switches to the next provider; once output has reached the client, it stays on the current one.
- Each API key can have its own quota, concurrency limit, RPM limit, and model redirects. A composite key is bound to several groups and selects one by a model name prefix.
- Content moderation supports keyword and hash rules, inline blocking, and automatic bans.

Users and billing:

- Sign-in options include email and password, passkeys, TOTP two-factor authentication, GitHub, Google, LinuxDo, OIDC, WeChat, and DingTalk.
- Teams share keys and quotas, and each member has a role.
- Charges follow pricing configs. Users pay from balance, subscription plans, quota packs, or redeem codes, and referral rebates are supported.
- Built-in payments support EasyPay, Alipay, WeChat Pay, Stripe, and Airwallex.
- The creative studio generates and edits images in the browser.

Operations:

- Every request creates a usage log with the model chain, token usage, cost, and diagnostics. The admin dashboard aggregates usage by user, group, provider, and model.
- The Ops dashboard collects request errors, upstream errors, provider availability, and host metrics, with alert rules and daily or weekly email reports.
- Administrators can check for new versions, upgrade, and roll back from the dashboard.

## Supported Upstreams

| Platform | Access methods |
| --- | --- |
| Anthropic | OAuth, Setup Token, API key, AWS Bedrock, Vertex AI |
| OpenAI | OAuth, API key |
| Gemini | OAuth, API key, Vertex AI |
| Antigravity | OAuth |
| Grok / xAI | OAuth, API key |
| Qoder | Qoder COSY |
| Kimi | API key (pay as you go, Coding plan) |
| Zhipu | API key (pay as you go, Coding plan) |
| DeepSeek | API key (pay as you go) |

The [upstream provider capability matrix (Chinese)](docs/interfaces/upstream_provider_matrix.md) lists the endpoints each platform supports and the combinations that import but not forward.

## Quick Start

Docker Compose deployment needs Docker 20.10 or later and Docker Compose v2 or later:

```bash
mkdir -p tokenrouter-deploy && cd tokenrouter-deploy

# Download the Compose file and .env, and generate the database password and secrets
curl -sSL https://raw.githubusercontent.com/TokenFlux/TokenRouter/main/deploy/docker-deploy.sh | bash

docker compose up -d
```

Then open `http://localhost:8080`. If `ADMIN_PASSWORD` is not set in `.env`, TokenRouter generates a random admin password and prints it to the log:

```bash
docker compose logs tokenrouter | grep "admin password"
```

Compose listens on the loopback address by default. To serve other machines, put a reverse proxy in front of it or change `BIND_HOST` in `.env`.

Other deployment options:

- [Install script (Chinese)](docs/guides/deployment/index.md#脚本安装) installs the binary on Linux as a systemd service. You provide PostgreSQL 15+ and Redis 7+.
- [Standalone container (Chinese)](deploy/DOCKER.md#独立容器) connects to an existing PostgreSQL and Redis.
- [Apple container (Chinese)](docs/guides/deployment/apple_container.md) runs TokenRouter locally on macOS.

If you are upgrading from Sub2API, read the compatibility notes in the [deployment guide (Chinese)](docs/guides/deployment/index.md#从旧名称部署升级) first. Starting the new Compose template over an old deployment creates empty data volumes.

## Development

The backend needs Go 1.27. The frontend needs Node.js 20 and pnpm 9.

```bash
# Backend
cd backend
go run ./cmd/server

# Frontend; the dev server proxies API requests to the backend
cd frontend
pnpm install --frozen-lockfile
pnpm run dev

# Or run the full stack built from source
docker compose -f deploy/docker-compose.dev.yml up --build
```

`make build` builds the backend and frontend. `make test` runs the backend tests, frontend lint, type checks, and the critical frontend tests. Test tiers, code generation, and pre-commit formatting are described in [development workflow (Chinese)](docs/operations/development_workflow.md).

## Documentation

- [Usage and operations guides (Chinese)](docs/guides/index.md): deployment, payment setup, and external integrations.
- [Engineering documentation (Chinese)](docs/index.md): architecture, business rules, interfaces, and operational constraints. Start here before changing code.
- [Upgrade notes (Chinese)](docs/operations/upgrade_notes.md): changes to watch for in each release.

## License

This project is distributed under the [GNU Lesser General Public License v3.0 or later](LICENSE).

Copyright (c) 2026 Wesley Liddick & TokenFlux
