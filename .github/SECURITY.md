# 安全策略

## 报告漏洞

**请不要通过公开 Issue 报告安全问题。**

优先使用 GitHub 的 [Private vulnerability reporting](https://docs.github.com/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability)
功能：进入仓库的 **Security** 标签页 → **Report a vulnerability**，该通道只有维护者可见。

若无法使用上述通道，请发邮件至：`758773857@qq.com`

## 请在报告中包含

- 受影响的版本或 commit、部署方式（Docker / 裸机）
- 复现步骤或 PoC，以及预期行为与实际行为
- 影响评估（可读取哪些数据、是否需要已登录、是否能提权）
- 你希望如何署名（默认匿名）

## 处理流程

1. 维护者在收到报告后确认收到，并初步判断影响范围
2. 修复在**非公开**分支进行，验证通过后合入并发布
3. 按你的意愿在致谢中署名；如需 CVE 编号可在披露前协助申请

## 范围

**在范围内**：本仓库代码本身的安全缺陷（认证/鉴权绕过、越权访问、注入、敏感信息泄露、上传校验绕过等）。

**不在范围内**：

- 依赖库自身的已知漏洞（请直接向对应上游项目报告）
- 需要物理接触或已取得服务器 root 权限的攻击场景
- 缺少 HTTPS、弱口令、端口暴露等**部署配置**问题——这些属于部署方责任，
  相关上线前检查项见仓库 `docs/DEPLOY.md` 的「上线前必做」小节
- 拒绝服务类的压测结果（除非能证明是代码缺陷而非容量不足）

## 已知的部署前提

本项目部分安全能力依赖部署方的正确配置，请勿把以下情况当作漏洞报告：

- 生产环境未启用 HTTPS
- `jwt.secret` 或数据库密码仍使用示例值
- MySQL / Redis 端口暴露到公网
- 未配置短信验证码接口限流

以上均为已知的上线前必做项，见 `docs/DEPLOY.md`。
