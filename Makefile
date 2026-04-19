test:
	go test ./...

build:
	go build ./...

build-admin:
	cd vben-admin/apps/admin && npm run build

build-user:
	cd vben-admin/apps/user && npm run build
