# 贡献指南

## 分支模型

| 分支 | 用途 |
|---|---|
| `develop` | 集成主分支，所有改动通过 MR 合到这里 |
| `feature/<模块>` | 新功能开发 |
| `fix/<模块>` | 缺陷修复 |

合并一律走 GitLab Merge Request，合并前确认 CI 通过。

## 提交规范

采用 Conventional Commits：

```
类型(范围): 一句话说清改了什么

- 正文用列表分点（可选）
```

| 类型 | 含义 |
|---|---|
| `feat` | 新功能 |
| `fix` | 修 bug |
| `refactor` | 重构（行为不变） |
| `perf` | 性能优化 |
| `test` | 测试 |
| `docs` | 文档 |
| `style` | 格式（不影响逻辑） |
| `chore` | 构建、配置、仓库维护 |

范围写模块名，如 `feat(comment):`、`fix(admin):`。

### 署名规则（硬性要求）

- **提交作者只能是你本人的 git 账号**；
- **提交信息中不得出现 AI 署名**（`Co-Authored-By`、`Generated with` 等），也不得出现他人姓名。

## 测试

- 新增 / 修改业务逻辑应附带测试；
- **测试代码必须提交** —— 它检查的是代码行为，行为是全队共有的资产；
- 依赖 MySQL 的用例用 `TEST_DB` 环境变量门控，未设置时自动 Skip。
  **只能使用一次性专用测试库 `tuyou_test`，绝不能指向生产库或共享开发库**：部分测试会执行 `AutoMigrate` 和全表删除 `follow` 数据；
  此类用例在连接数据库前检查 DSN 的数据库名。不需要真库测试时不要设置 `TEST_DB`。

  ```bash
  TEST_DB="root:密码@tcp(127.0.0.1:3306)/tuyou_test?charset=utf8mb4&parseTime=True&loc=Local" go test ./...
  ```

## 合并前自查

```bash
go build ./...          # 编译通过
go vet ./...            # 静态检查
go test ./...           # 测试全绿（无 TEST_DB 时自动跳过数据库用例）
golangci-lint run       # lint（初期为非阻断，存量问题清理后转强制）
```

## 配置与安全红线

- 任何真实凭据（AccessKey、密码、JWT secret）**不得写进任何被提交的文件**；
  配置一律走 `configs/config.example.yaml` 样例 + 环境变量注入；
- 新增配置文件遵循 `configs/config.<环境>.yaml` 命名（已被 `.gitignore` 排除）；
- 禁止提交：`.idea/`、`logs/`、`bin/`、AI 工作目录（`.claude/`、`.superpowers/`）。
