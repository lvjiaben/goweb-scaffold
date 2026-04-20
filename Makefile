CONFIG ?= configs/config.yaml
FORMAT ?= text
MODULE ?= demo_article
TABLE ?= demo_article
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

test:
	go test ./...

build:
	go build ./...

build-admin:
	cd vben-admin/apps/admin && npm run build

build-user:
	cd vben-admin/apps/user && npm run build

codegen-tables:
	go run ./cmd/codegen tables -config $(CONFIG) -format $(FORMAT)

codegen-modules:
	go run ./cmd/codegen modules -config $(CONFIG) -format $(FORMAT)

codegen-preview:
	go run ./cmd/codegen preview -config $(CONFIG) -module $(MODULE) -table $(TABLE) $(if $(PAYLOAD),-payload $(PAYLOAD),) $(if $(FROM),-from $(FROM),) -format $(FORMAT)

codegen-diff:
	go run ./cmd/codegen diff -config $(CONFIG) -module $(MODULE) -table $(TABLE) $(if $(PAYLOAD),-payload $(PAYLOAD),) $(if $(FROM),-from $(FROM),) -overwrite=$(OVERWRITE) -register-module=$(REGISTER_MODULE) -upsert-menu=$(UPSERT_MENU) -format $(FORMAT)

codegen-generate:
	go run ./cmd/codegen generate -config $(CONFIG) -module $(MODULE) -table $(TABLE) $(if $(PAYLOAD),-payload $(PAYLOAD),) $(if $(FROM),-from $(FROM),) -overwrite=$(OVERWRITE) -register-module=$(REGISTER_MODULE) -upsert-menu=$(UPSERT_MENU) -format $(FORMAT)

codegen-regenerate:
	go run ./cmd/codegen regenerate -config $(CONFIG) -module $(MODULE) -history-id=$(HISTORY_ID) -overwrite=$(OVERWRITE) -register-module=$(REGISTER_MODULE) -upsert-menu=$(UPSERT_MENU) -format $(FORMAT)

codegen-remove:
	go run ./cmd/codegen remove -config $(CONFIG) -module $(MODULE) -remove-files=$(REMOVE_FILES) -unregister-module=$(UNREGISTER_MODULE) -remove-menu=$(REMOVE_MENU) -remove-history=$(REMOVE_HISTORY) -remove-lock=$(REMOVE_LOCK) -format $(FORMAT)
