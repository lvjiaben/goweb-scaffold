CONFIG ?= configs/config.yaml
FORMAT ?= text
MODULE ?= app_user
TABLE ?= app_user
PAYLOAD ?=
FROM ?=
HISTORY_ID ?= 0
OVERWRITE ?= true
REGISTER_MODULE ?= true
UPSERT_MENU ?= true
REMOVE_FILES ?= true
UNREGISTER_MODULE ?= true
REMOVE_MENU ?= true
REMOVE_HISTORY ?= false
REMOVE_LOCK ?= true
PLAN ?=
OUTPUT ?=

.PHONY: help install run dev build build-backend build-user frontend-install fmt vet mod clean \
	codegen-tables codegen-modules codegen-preview codegen-generate codegen-regenerate codegen-remove codegen-batch-generate migrate-info db-reset

help: ## 显示命令帮助
	@awk 'BEGIN {FS = ":.*##"; printf "goweb-scaffold 可用命令：\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-24s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## 安装 Go 依赖
	@echo "📦 安装 Go 依赖"
	@go mod download

frontend-install: ## 安装前端依赖
	@echo "📦 安装 vben-admin 依赖"
	@cd vben-admin && pnpm install

run: ## 运行后端服务
	@echo "🚀 启动后端服务"
	@go run ./cmd/server

dev: run ## 开发模式运行后端

build: ## 构建 Go 后端
	@echo "🔨 构建 Go 后端"
	@go build ./...

build-backend: ## 构建 vben-admin backend 端
	@echo "🔨 构建 vben-admin backend"
	@cd vben-admin/apps/backend && npm run build

build-user: ## 构建 vben-admin user 端
	@echo "🔨 构建 vben-admin user"
	@cd vben-admin/apps/user && npm run build

fmt: ## 格式化 Go 代码
	@echo "🧹 gofmt"
	@gofmt -w $$(find . -path './vben-admin' -prune -o -name '*.go' -print)

vet: ## 运行 go vet
	@echo "🔎 go vet"
	@go vet ./...

mod: ## 整理 Go module
	@echo "🧩 go mod tidy"
	@go mod tidy

clean: ## 清理构建产物
	@echo "🧽 清理构建产物"
	@rm -rf bin dist tmp vben-admin/apps/backend/dist vben-admin/apps/user/dist

migrate-info: ## 显示当前迁移文件
	@echo "🗃️ 迁移文件"
	@ls -1 migrations

db-reset: ## 提示数据库重建方式
	@echo "⚠️ 请手工重建 PostgreSQL 数据库后执行 migrations/0001_init.sql"

codegen-tables: ## 列出业务表
	@echo "🧬 codegen tables"
	@go run ./cmd/codegen tables -config $(CONFIG) -format $(FORMAT)

codegen-modules: ## 列出已生成模块
	@echo "🧬 codegen modules"
	@go run ./cmd/codegen modules -config $(CONFIG) -format $(FORMAT)

codegen-preview: ## 预览生成方案
	@echo "🧬 codegen preview MODULE=$(MODULE) TABLE=$(TABLE)"
	@go run ./cmd/codegen preview -config $(CONFIG) -module $(MODULE) -table $(TABLE) $(if $(PAYLOAD),-payload $(PAYLOAD),) $(if $(FROM),-from $(FROM),) -format $(FORMAT)

codegen-generate: ## 生成 backend CRUD 文件
	@echo "🧬 codegen generate MODULE=$(MODULE) TABLE=$(TABLE)"
	@go run ./cmd/codegen generate -config $(CONFIG) -module $(MODULE) -table $(TABLE) $(if $(PAYLOAD),-payload $(PAYLOAD),) $(if $(FROM),-from $(FROM),) -overwrite=$(OVERWRITE) -register-module=$(REGISTER_MODULE) -upsert-menu=$(UPSERT_MENU) -format $(FORMAT)

codegen-regenerate: ## 重新生成模块
	@echo "🧬 codegen regenerate MODULE=$(MODULE)"
	@go run ./cmd/codegen regenerate -config $(CONFIG) -module $(MODULE) -history-id=$(HISTORY_ID) -overwrite=$(OVERWRITE) -register-module=$(REGISTER_MODULE) -upsert-menu=$(UPSERT_MENU) -format $(FORMAT)

codegen-remove: ## 卸载生成模块
	@echo "🧬 codegen remove MODULE=$(MODULE)"
	@go run ./cmd/codegen remove -config $(CONFIG) -module $(MODULE) -remove-files=$(REMOVE_FILES) -unregister-module=$(UNREGISTER_MODULE) -remove-menu=$(REMOVE_MENU) -remove-history=$(REMOVE_HISTORY) -remove-lock=$(REMOVE_LOCK) -format $(FORMAT)

codegen-batch-generate: ## 批量生成模块，需 PLAN=...
	@echo "🧬 codegen batch generate PLAN=$(PLAN)"
	@go run ./cmd/codegen batch -config $(CONFIG) -plan $(PLAN) -mode generate -format $(FORMAT)
