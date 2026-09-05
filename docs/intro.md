# 代理关了，终端还在瞎连？用 prox 管住 Linux 终端里的 http_proxy 代理设置

> 面向使用者的产品介绍。当前文档对应 V0.2.0。

经常上网的朋友都知道，在 Linux 终端里走 HTTP 代理，几乎总绕不开几个东西：本机 Clash / V2Ray、公司网关、实验室或云上的转发端口，最后都会变成类似的系统配置指令：

```bash
export http_proxy=http://127.0.0.1:7890
export https_proxy=http://127.0.0.1:7890
# 或：http://proxy.company.example:8080
```

如果直接写进 `~/.bashrc`，代理可用时确实很省事。可一旦代理软件退出、公司网关变更或端口换了，Git、curl、包管理器却还在连已经失效的地址 —— 超时、报错、莫名其妙的失败，都可能出现。

**`prox`** 开源项目的诞生，就是为了把这件事做对：**它不是替代你的代理软件，只负责安全、可逆、可诊断地管理终端里的代理环境变量。**

GitHub 仓库地址：[https://github.com/luhuadong/prox](https://github.com/luhuadong/prox)


## 是什么

**`prox`** 是一个面向 Linux 终端的开源代理环境管理工具。

如果只能用一句话介绍，那就是 —— **为 Linux 终端提供一套简单、可靠、可诊断的代理开关。**

它做的事很克制：

- 设置（或恢复）当前终端及其子进程使用的代理环境变量；
- 在启用前检查代理端点是否可用；
- 为单条命令提供临时代理环境；
- 用分层健康检查告诉你「端口通不通」和「能不能真正出网」。

它刻意不做的事：

- 不提供代理服务本身；
- 不管理 Clash / V2Ray / 企业代理 / VPN；
- 不改 `git`、`npm`、Docker、桌面系统代理等持久配置；
- 不收集遥测数据；
- 不依赖 curl、nc、jq 或后台守护进程。

当前版本支持：**Linux + Bash 4.3+ + HTTP Proxy**。核心用 Go 实现，Bash Hook 负责修改当前 Shell 环境。

日常命令可以概括为：

```bash
prox on
prox off
prox status
prox check
prox run -- <command>
```

其中 **`prox run` 只影响指定命令**，是最推荐的日常用法（后面会介绍）。


## 为什么需要它

### 常见痛点

1. **永久写入 `.bashrc`**  
   代理软件退出后，环境变量还在。工具继续往死端口发包，失败原因往往不直观。

2. **手动 `export` / `unset`**  
   容易漏变量（大小写混用、`no_proxy` 忘设）、关不干净，或把公司原有代理一并清掉。

3. **「感觉开了代理」和「代理真的可用」混为一谈**  
   Shell 里有变量，不代表代理端口在听，更不代表 HTTPS 能通。

### prox 给出的答案

| 问题 | prox 的做法 |
| --- | --- |
| 代理挂了还继续污染环境 | `prox on` 前先做端点检查，失败则**不改环境** |
| 关闭时代理变量乱清 | `prox off` **恢复启用前的快照**，不是无条件 `unset` |
| 启用后又被手改乱了 | 检测 **drifted**，关闭时仍尽量恢复快照 |
| 只想给一条命令走代理 | `prox run -- ...`，不碰当前 Shell |
| 想确认代理是否健康 | `prox check`：配置 → TCP → 真实 HTTP/HTTPS |

它把「开关」和「诊断」放在同一套小工具里，让日常代理使用变得可预期。


## 怎么装

任选一种即可。安装后确保对应目录在 `PATH` 中（用户级安装通常是 `~/.local/bin`）。

### 安装脚本（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/luhuadong/prox/main/scripts/install.sh | sh
```

脚本会按架构下载 Release 预编译包并校验 SHA-256。它不会改 Shell 配置，也不会自动创建用户配置文件。

每个 Release 还提供 `.deb` / `.rpm`，可用系统包管理器安装，并附带 Bash completion。

例如：

```bash
sudo dpkg -i prox_0.2.0_linux_amd64.deb
```

### 其他方式

```bash
# Go
go install github.com/luhuadong/prox/cmd/prox@latest

# 源码
git clone https://github.com/luhuadong/prox.git
cd prox
make check
make install PREFIX="$HOME/.local"
```

安装后自检：

```bash
prox version
prox check --local
```

**注意：** 若你以前用过 `make install` 装到 `~/.local/bin`，后来又装了 deb/rpm，`PATH` 里用户目录通常优先于 `/usr/bin`。此时 `prox version` 可能仍显示旧版。删掉旧二进制或调整 `PATH` 即可：

```bash
rm -f ~/.local/bin/prox
hash -r
prox version
```


## 怎么用

### 1. 最快路径：不改当前 Shell

多数场景到这一步就够了 —— **不需要** `prox init`：

```bash
# 本地代理软件先开着，或代理网关可达（默认检查 127.0.0.1:7890）
prox check --local          # 只测端口，不出外网
prox check                  # 再做完整健康检查（会访问 check_url）

prox run -- curl -I https://www.google.com
prox run -- git clone https://github.com/example/project.git
```

`prox run` 在端点检查通过后，把代理变量交给目标命令及其子进程；当前终端环境保持不变。

### 2. 需要整段会话都走代理：先加载 Hook

只有 `prox on` / `prox off`，以及准确的「当前 Shell 是否已由 prox 激活」的 `prox status`，才需要 Bash 集成：

```bash
# 当前终端临时加载（不会自动开启代理，也不访问网络）
eval "$(prox init bash)"

# 确认无误后，可写入 ~/.bashrc，每个新 Bash 自动加载
```

然后：

```bash
prox on          # 启用前检查端点；失败则环境不变
prox status      # 只看 Shell 状态，不上网
prox off         # 恢复启用前的变量快照
```

可以把它理解成：

- `init`：装好「开关能力」；
- `on` / `off`：真正拧开关；
- `run`：一次性的临时插座，不必拧墙上的总开关。

### 3. 配置代理地址

默认使用内置值：`http://127.0.0.1:7890`。没有配置文件也不会报错。

用户配置路径：

```bash
~/.config/prox/config.json
```

（若设置了 `XDG_CONFIG_HOME`，则为 `$XDG_CONFIG_HOME/prox/config.json`。）

常用配置命令：

```bash
prox config path                 # 打印默认配置文件路径
prox config show                 # 查看合并内置默认后的完整配置
prox config init                 # 写入默认 7890 的最小配置（不覆盖已有文件）
prox config init --proxy http://127.0.0.1:7897
prox config validate
```

最小配置示例：

```json
{
  "proxy_url": "http://127.0.0.1:7890"
}
```

省略的字段继续继承内置默认（`no_proxy`、`check_url`、超时等）。也可单次指定文件：

```bash
prox --config /path/to/config.json check
```

### 4. 读懂状态与健康

`prox status` 回答的是「**当前 Shell 是否由 prox 启用**」，不上网：

```bash
State: inactive
State: active
State: drifted    # 启用后又被手动改乱了
```

`prox check` 回答的是「**配置的代理能不能用**」：

1. 配置是否合法；
2. 代理主机端口能否 TCP 连通；
3. 能否经该代理访问健康检查 URL，并得到预期状态码（默认 Google `generate_204` → HTTP 204）。

两者分开，是为了避免把「环境变量开了」误当成「网络通了」。

## 设计理念

1. **失败不改环境**  
   端点不可达时，执行 `prox on` / `prox run` 直接失败，避免半残环境。

2. **恢复快照，而不是清空**  
   若启用前已有公司代理等变量，执行 `prox off` 会还原，而不是一律删除。

3. **Go 核心 + Bash Hook**  
   网络、URL、超时、TLS、测试等核心代码使用 Go 语言编写；修改父 Shell 环境等操作依赖 Bash Hook。职责清晰，依赖少（零第三方 Go Module）。

4. **边界刻意收窄**  
   当前最新的 V0.2 版本暂不支持 SOCKS5、代理认证、PAC、多 Profile、Zsh/Fish、macOS/Windows 等。目标是把「终端代理开关 + 检查」做到简单可预测，而不是做成又一个全能代理套件。



## 典型工作流速查

```bash
# 安装后
prox version
prox config show

# 代理软件已启动
prox check --local
prox check

# 推荐：单命令走代理
prox run -- curl https://example.com

# 可选：整段会话
eval "$(prox init bash)"   # 或已写入 ~/.bashrc
prox on
# ... 在本终端工作 ...
prox off
```


## 小结

`prox` 不是又一个代理客户端，而是终端侧的**代理环境开关与诊断器**：

- **是什么**：管理 HTTP 代理相关环境变量，并检查代理是否可用；
- **为什么**：告别「写进 bashrc 就忘关」、失败原因难查、开关不可逆等问题；
- **怎么用**：优先 `prox check` + `prox run`；需要会话级开关时再 `init` + `on` / `off`。

如果你已经在用 HTTP 代理（本机、公司网关或其它可达端点），又希望 Linux 终端里的开关更干净、更可诊断，欢迎试用，也欢迎在 GitHub 上反馈问题与想法。
