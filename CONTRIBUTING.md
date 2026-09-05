# Contributing to prox

感谢参与 `prox`。

## 开始之前

当前版本有意保持克制。提交新功能前，请先确认它是否属于当前产品范围。多 Profile、SOCKS、其他 Shell 和后台监控应先通过 Issue 讨论。

## 本地检查

提交前运行：

```bash
make check
```

它会执行格式检查、Go 单元测试、`go vet`、Bash 语法检查和 Bash Hook 集成测试。

## 发布

版本号以 Git tag 为唯一来源，不需要修改 Go 源码常量。准备发布时：

1. 将 `CHANGELOG.md` 中的 `Unreleased` 改为版本号和发布日期；
2. 运行 `make check`、`make release-check` 和 `make snapshot`；
3. 提交发布准备改动；
4. 创建并推送形如 `v0.2.0` 的 tag。

tag 推送后，GitHub Actions 会构建并发布对应版本的二进制、deb、rpm 和校验和。

## 提交原则

- 一个变更解决一个明确问题；
- 用户可见行为需要测试；
- 错误信息应说明失败阶段和可执行的下一步；
- 不得让健康检查隐式继承当前代理环境；
- 不得在日志或错误信息中暴露代理凭据；
- Shell 动态值必须正确转义。

## License

除非明确另行说明，向本项目提交的贡献将依据 Apache License 2.0 授权。
