.PHONY: build build-backend build-frontend build-datamanagementd test test-backend test-frontend test-frontend-critical test-datamanagementd fmt-go-changed check-fmt-go-changed

PNPM ?= python3 tools/verify.py frontend pnpm
export GOTOOLCHAIN := go$(shell sed -n 's/^go //p' backend/go.mod)
export PYTHONDONTWRITEBYTECODE := 1

FRONTEND_CRITICAL_VITEST := \
	src/api/__tests__/client.spec.ts \
	src/api/__tests__/tokenRefresh.spec.ts \
	src/views/admin/orders/__tests__/AdminOrdersView.spec.ts \
	src/views/auth/__tests__/LinuxDoCallbackView.spec.ts \
	src/views/auth/__tests__/WechatCallbackView.spec.ts \
	src/views/user/__tests__/PaymentView.spec.ts \
	src/views/user/__tests__/PaymentResultView.spec.ts \
	src/components/user/profile/__tests__/ProfileEditForm.spec.ts \
	src/components/user/profile/__tests__/ProfileInfoCard.spec.ts \
	src/views/admin/__tests__/SettingsView.spec.ts \
	src/components/admin/usage/__tests__/UsageStatsCards.spec.ts \
	src/composables/__tests__/useQoderOAuth.spec.ts \
	src/components/provider/__tests__/CreateProviderModal.qoder.spec.ts \
	src/views/admin/__tests__/ProvidersView.qoderCreate.spec.ts

# 一键编译前后端
build: build-backend build-frontend

# 编译后端（复用 backend/Makefile）
build-backend:
	@$(MAKE) -C backend build

# 编译前端（需要已安装依赖）
build-frontend:
	@$(PNPM) --dir frontend run build

# 编译 datamanagementd（宿主机数据管理进程）
build-datamanagementd:
	@cd datamanagement && go build -o datamanagementd ./cmd/datamanagementd

# 运行测试（后端 + 前端）
test: test-backend test-frontend

test-backend:
	@$(MAKE) -C backend test

test-frontend:
	@$(PNPM) --dir frontend run lint:check
	@$(PNPM) --dir frontend run typecheck
	@$(MAKE) test-frontend-critical

test-frontend-critical:
	@$(PNPM) --dir frontend exec vitest run $(FRONTEND_CRITICAL_VITEST)

test-datamanagementd:
	@cd datamanagement && go test ./...

# 提交前只格式化改动的手写 Go 文件，复用 backend/.golangci.yml。
fmt-go-changed:
	@python3 tools/format_go.py

check-fmt-go-changed:
	@python3 tools/format_go.py --check $(if $(FMT_BASE),--base "$(FMT_BASE)")

# 完整验证顺序由这里维护，CI 分组调用同名目标。
.PHONY: verify verify-environment verify-diff verify-static verify-backend verify-frontend verify-embed verify-scripts verify-installer verify-tools verify-security verify-security-backend verify-security-frontend install-hooks
verify:
	@python3 tools/verify.py full

verify-environment:
	@python3 tools/verify.py preflight all

verify-diff:
	@python3 tools/verify.py diff

verify-static:
	@python3 tools/verify.py preflight go
	@$(MAKE) -C backend test-architecture
	@cd backend && bash ../tools/golangci-lint.sh config verify
	@cd backend && bash ../tools/golangci-lint.sh run --timeout=30m ./...
	@cd backend && bash ../tools/golangci-lint.sh run --timeout=30m --build-tags=unit ./...
	@cd backend && bash ../tools/golangci-lint.sh run --timeout=30m --build-tags=integration ./...

# 单组入口供 CI 矩阵和本地调度共用。
.PHONY: verify-backend-ordinary verify-backend-unit verify-backend-integration verify-frontend-build verify-frontend-checks verify-frontend-ci-checks
verify-backend:
	@$(MAKE) verify-backend-ordinary
	@$(MAKE) verify-backend-unit
	@$(MAKE) verify-backend-integration

verify-backend-ordinary verify-backend-unit verify-backend-integration:
	@python3 tools/verify.py preflight backend
	@python3 tools/verify.py go-test $(patsubst verify-backend-%,%,$@)

verify-frontend:
	@$(MAKE) verify-frontend-build
	@$(MAKE) verify-frontend-checks

verify-frontend-build:
	@python3 tools/verify.py preflight frontend
	@$(PNPM) --dir frontend install --frozen-lockfile
	@$(PNPM) --dir frontend run typecheck
	@$(PNPM) --dir frontend run build:assets

# 本地在构建后复用依赖，CI 的独立检出先执行安装入口。
verify-frontend-ci-checks:
	@python3 tools/verify.py preflight frontend
	@$(PNPM) --dir frontend install --frozen-lockfile
	@$(MAKE) verify-frontend-checks

verify-frontend-checks:
	@$(PNPM) --dir frontend run lint:check
	@$(PNPM) --dir frontend run test:run

# embed 测试要求已经生成前端资源，.keep 占位文件不能代替生产构建。
verify-embed:
	@test -s backend/internal/web/dist/index.html || (echo '请先运行 make verify-frontend'; exit 1)
	@python3 tools/verify.py preflight go
	@cd backend && bash ../tools/golangci-lint.sh run --timeout=30m --build-tags=embed ./...
	@python3 tools/verify.py go-test embed
	@cd backend && CGO_ENABLED=0 go build -tags=embed -trimpath -o bin/server ./cmd/server

verify-tools:
	@python3 -m unittest discover -s tools -p 'test_*.py'

verify-scripts:
	@/bin/bash -n deploy/apple-container.sh deploy/install.sh
	@/bin/bash deploy/tests/apple-container-test.sh
	@/bin/sh deploy/tests/docker-compose-security-test.sh
	@/bin/sh deploy/tests/docker-compose-postgres-test.sh
	@/bin/sh deploy/tests/docker-compose-variants-test.sh
	@/bin/sh deploy/tests/docker-compose-gateway-env-test.sh
	@/bin/sh deploy/tests/docker-runtime-resources-test.sh
	@/bin/sh deploy/test-caddyfile-cache.sh
	@/bin/sh tools/goreleaser_prebuilt_test.sh
	@python3 tools/test_notify_discord_release.py

verify-installer:
	@if [ "$$(uname -s)" = Linux ]; then bash deploy/tests/install-brand-test.sh; else docker run --rm -v "$(CURDIR):/source:ro" -w /source ubuntu:24.04 bash deploy/tests/install-brand-test.sh; fi

verify-security:
	@python3 tools/verify.py run verify-security-backend verify-security-frontend

verify-security-backend:
	@python3 tools/verify.py preflight go
	@cd backend && go run golang.org/x/vuln/cmd/govulncheck@$$(cat ../.govulncheck-version) ./...

verify-security-frontend:
	@python3 tools/verify.py preflight frontend
	@$(PNPM) --dir frontend install --frozen-lockfile
	@python3 tools/check_pnpm_audit_exceptions.py --run --exceptions .github/audit-exceptions.yml

install-hooks:
	@python3 tools/pre_push.py --install

.PHONY: verify-go-version
verify-go-version:
	@python3 tools/verify.py preflight go
