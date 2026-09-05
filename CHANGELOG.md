# Changelog

本项目遵循语义化版本。

## Unreleased

- 修正 Go module 路径，支持版本化 `go install`；
- 增加 Linux amd64/arm64 的自动化 Release 构建与校验和；
- 增加校验 Release 制品的用户安装脚本；
- 增加 `make install`、`make uninstall` 和安装测试；
- 增加 `prox config path/show/init/validate`；
- 增加 Bash completion 和 `prox completion bash`；
- Release 同时生成 Linux `.deb` 和 `.rpm` 软件包；
- 修复 Shell Hook 对全局 `--config` 参数的转发。

## 0.1.0 - 2026-09-05

首个可用版本：

- Linux + Bash 4.3+；
- HTTP Proxy；
- `prox init bash`；
- `prox on` / `prox off`；
- `prox status`；
- `prox check`；
- `prox run -- <command>`；
- 环境快照、恢复和漂移检测；
- TCP 与 HTTP/HTTPS 分层健康检查；
- 严格 JSON 配置；
- 零第三方 Go Module。
