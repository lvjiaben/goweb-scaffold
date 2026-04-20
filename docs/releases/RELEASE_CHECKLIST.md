# Release Checklist

## 版本文件

- [ ] `VERSION` 已更新
- [ ] `CHANGELOG.md` 已更新
- [ ] 当前 release notes 已写入 `docs/releases/v0.9.0-rc.1.md`

## 文档

- [ ] `docs/versioning.md`
- [ ] `docs/compatibility.md`
- [ ] `docs/releases/RELEASE_POLICY.md`
- [ ] `docs/releases/RELEASE_CHECKLIST.md`
- [ ] `docs/upgrade/UPGRADE_v0.9.0-rc.1.md`
- [ ] README 已给出版本与文档入口

## 校验

- [ ] `go test ./...`
- [ ] `go build ./...`
- [ ] `npm run build` for `apps/admin`
- [ ] `npm run build` for `apps/user`
- [ ] `make release-check`
- [ ] `go run ./cmd/codegen version -format json`

## 发布说明

- [ ] 明确写明这是 `rc`
- [ ] 明确当前 template version = `v7`
- [ ] 明确当前兼容 `goweb-core v0.9.0-rc.1`
