## 变更类型

<!-- 勾选一项或多项 -->

- [ ] feat：新功能
- [ ] fix：缺陷修复
- [ ] refactor：重构（不改变外部行为）
- [ ] perf：性能优化
- [ ] docs：文档
- [ ] test：测试
- [ ] chore：构建 / 依赖 / 配置

## 变更内容

<!-- 说明改了什么、为什么改。如果是修复，写清根因，不只写现象。 -->

## 关联 Issue

<!-- 例：Closes #123。没有可留空。 -->

## 验证方式

<!-- 必填：贴出实际执行的命令与输出，不要只写「本地测试通过」。 -->

```
# 例
go build ./... && go vet ./... && go test ./...
```

## 自查清单

- [ ] 已跑通 `go build ./...` / `go vet ./...` / `go test ./...`
- [ ] 新增或修改的逻辑有对应测试；依赖数据库的测试已用 `TEST_DB` 门控
- [ ] 未提交任何真实凭据（`configs/*.yaml`、`.env`、证书、AccessKey）
- [ ] 涉及接口变更时，已同步更新 `docs/api/openapi.json`
- [ ] 涉及表结构变更时，已同步更新 `docs/database-design.md`
- [ ] 提交信息符合 Conventional Commits，且作者为本人账号
