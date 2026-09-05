# prox 产品需求文档

> Product Requirements Document

| 项目 | 内容 |
| --- | --- |
| 产品名称 | `prox` |
| 文档版本 | 0.1 |
| 产品版本 | V0.1 |
| 状态 | Draft |
| 日期 | 2026-09-05 |
| 首发平台 | Linux + Bash |
| 实现约束 | 语言无关 |

## 1. 产品概述

`prox` 是一个面向终端用户的代理环境管理工具，用于安全、明确、可逆地启用、关闭、检查和临时使用网络代理。

它不提供代理服务，不负责创建网络隧道，也不替代 Clash、V2Ray、企业代理或 VPN。它只管理当前终端及其子进程使用的代理环境，并提供最基本的可用性检查。

一句话定位：

> 为 Linux 终端提供一套简单、可靠、可诊断的代理开关。

V0.1 的核心操作固定为：

```bash
prox on
prox off
prox status
prox check
prox run -- <command>
```

## 2. 背景与问题

开发者通常通过以下方式为终端配置代理：

```bash
export ALL_PROXY="http://127.0.0.1:7890"
```

或者将多个代理变量写进 `.bashrc`：

```bash
export http_proxy="http://127.0.0.1:7890"
export https_proxy="http://127.0.0.1:7890"
export ALL_PROXY="http://127.0.0.1:7890"
```

这种方式存在以下问题：

1. 代理程序未启动时，所有继承代理变量的命令都可能失败。
2. 代理变量名称较多，不同程序读取的变量不完全一致。
3. 用户经常忘记清理已经设置的变量。
4. 手动关闭代理时，容易误删终端中原本存在的代理设置。
5. 当前终端是否已经启用代理并不直观。
6. 代理端口存在，不代表代理能够访问目标网络。
7. 使用 `curl -I` 手动检查虽然有效，但输出冗长，难以用于脚本判断。
8. 只想让一条命令使用代理时，设置整个终端环境显得过重。

`prox` 需要将这些零散操作收敛成稳定的命令语义。

## 3. 产品目标

### 3.1 V0.1 目标

V0.1 必须实现：

1. 用户可以通过一条命令为当前终端启用代理。
2. 启用前先检查代理端点，代理不可用时不得污染当前环境。
3. 用户可以通过一条命令关闭代理，并恢复启用前的环境。
4. 用户可以查看当前终端的代理状态，且该操作不访问外网。
5. 用户可以执行完整代理健康检查，并获得简洁、明确的结果。
6. 用户可以只为单条命令及其子进程提供代理。
7. 所有日常操作都具有稳定的退出状态，能够用于 Shell 脚本。
8. 默认配置可以覆盖最常见的本地代理场景，无需首次交互式初始化。

### 3.2 成功标准

V0.1 达到以下条件即可视为有效：

- 新用户在 5 分钟内完成安装并成功执行 `prox check`。
- 代理端口未监听时，`prox on` 在 3 秒内失败，且当前代理变量保持不变。
- `prox off` 可以准确恢复 `prox on` 之前的相关环境变量。
- `prox run` 执行结束后，当前 Shell 环境不发生变化。
- 默认完整检查在网络正常时 8 秒内完成。
- 用户不需要理解每一个代理环境变量，也能正确使用工具。

## 4. 非目标

V0.1 不包含以下能力：

- 不启动、停止或管理代理服务进程。
- 不提供 VPN、隧道、流量转发或代理服务器。
- 不修改桌面系统代理。
- 不修改 Git、npm、pip、Docker 等工具的持久配置。
- 不修改 systemd 服务或系统级环境变量。
- 不修改 `/etc/environment`、`/etc/profile` 等系统文件。
- 不自动向 `sudo` 或 root 环境传播代理变量。
- 不提供后台守护进程或持续监控。
- 不根据网络环境自动开启或关闭代理。
- 不支持代理订阅、节点选择或测速。
- 不支持 PAC。
- 不支持多代理 Profile。
- 不支持带用户名和密码的代理。
- 不提供图形界面。
- 不承诺让所有程序都遵守代理环境变量。

## 5. 目标用户

### 5.1 主要用户

- Linux 开发者；
- 经常使用 `curl`、Git、包管理器和各类 CLI 的工程师；
- 使用 Clash、V2Ray 或本地 HTTP 代理的用户；
- 在直连和代理网络之间切换的终端用户。

### 5.2 典型使用场景

#### 场景 A：临时开启当前终端代理

```bash
prox on
git clone https://github.com/example/project.git
prox off
```

#### 场景 B：只让一条命令使用代理

```bash
prox run -- git clone https://github.com/example/project.git
```

#### 场景 C：代理应用已经打开，但不确定节点是否可用

```bash
prox check
```

#### 场景 D：命令访问异常，需要确认当前终端是否仍在使用代理

```bash
prox status
```

## 6. 产品原则

### 6.1 显式

代理只能由用户明确启用。V0.1 不在打开新终端时自动启用代理。

### 6.2 失败不污染

如果配置无效或代理端点不可达，`prox on` 必须在修改任何代理变量之前失败。

### 6.3 可逆

`prox off` 必须恢复 `prox on` 之前的状态，而不是简单地清空全部代理变量。

### 6.4 作用域优先

产品应优先推荐 `prox run`。能够限定在单条命令内的代理，不应要求用户修改整个终端会话。

### 6.5 状态与健康分离

“当前终端已经设置代理变量”和“代理当前可以正常访问网络”是两个不同概念：

- 激活状态由 `prox status` 表达；
- 网络健康由 `prox check` 表达。

### 6.6 快速且可预测

`status` 不执行网络请求；`on` 只进行快速端点检查；只有 `check` 执行完整外网检查。

### 6.7 保持透明

`prox run` 不应改变目标命令的标准输入、标准输出和标准错误行为，并应尽可能原样传递退出状态与终止信号。

### 6.8 隐私克制

V0.1 不收集遥测数据，不上传配置，不记录用户执行的目标命令。

## 7. V0.1 产品范围

### 7.1 平台范围

V0.1 支持：

- Linux；
- Bash；
- 交互式终端会话；
- HTTP 代理；
- 通过 HTTP CONNECT 访问 HTTPS 目标。

以下平台和协议推迟到后续版本：

- Zsh；
- Fish；
- macOS；
- Windows PowerShell；
- SOCKS5 / SOCKS5H；
- HTTPS Proxy；
- 代理认证。

### 7.2 命令范围

| 命令 | 作用 | 是否访问网络 |
| --- | --- | --- |
| `prox init bash` | 输出 Bash 集成代码 | 否 |
| `prox config <command>` | 查看、创建或校验用户配置 | 否 |
| `prox on` | 为当前 Shell 启用代理 | 仅检查代理端点 |
| `prox off` | 恢复启用前的代理环境 | 否 |
| `prox status` | 查看当前 Shell 的激活状态 | 否 |
| `prox check` | 检查代理端点和外网访问 | 是 |
| `prox run -- <command>` | 仅为目标命令启用代理 | 仅检查代理端点 |

## 8. 核心概念

### 8.1 Proxy Configuration

描述 `prox` 当前使用的代理地址、绕过列表、检查目标和超时策略。

V0.1 只有一份默认配置，不引入命名 Profile。

### 8.2 Activation State

描述当前 Shell 是否由 `prox` 启用了代理：

| 状态 | 含义 |
| --- | --- |
| `inactive` | 当前 Shell 未由 `prox` 启用代理 |
| `active` | 当前 Shell 正在使用 `prox` 设置的代理变量 |
| `drifted` | 已经启用，但相关变量随后被其他操作修改 |

激活状态只属于当前 Shell，不在不同终端窗口之间共享。

### 8.3 Health State

描述一次主动检查的结果：

| 状态 | 含义 |
| --- | --- |
| `healthy` | 配置、代理端点和测试请求均成功 |
| `config_invalid` | 配置无法解析或不受支持 |
| `endpoint_unreachable` | 无法连接代理主机或端口 |
| `proxy_rejected` | 代理拒绝、认证失败或无法建立隧道 |
| `target_failed` | 代理端点可达，但测试目标访问失败 |

健康状态是一次检查结果，不作为永久状态缓存。

### 8.4 Environment Snapshot

`prox on` 修改环境之前保存的相关变量状态，包括：

- 变量原本是否存在；
- 如果存在，其原始值是什么。

`prox off` 使用该快照恢复环境。

## 9. 默认配置

### 9.1 配置字段

| 字段 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `proxy_url` | 否 | `http://127.0.0.1:7890` | HTTP 代理地址 |
| `no_proxy` | 否 | `localhost,127.0.0.1,::1,.local` | 不经过代理的地址 |
| `check_url` | 否 | `https://www.google.com/generate_204` | 完整健康检查目标 |
| `expected_status` | 否 | `204` | 默认检查期望的 HTTP 状态码 |
| `connect_timeout` | 否 | `2s` | 建立连接的最长时间 |
| `request_timeout` | 否 | `8s` | 完整检查的最长时间 |

### 9.2 配置存储

- 配置应位于用户级配置目录，并遵循 XDG Base Directory 约定。
- 配置文件应允许用户直接阅读和编辑。
- 具体序列化格式由技术设计决定，不属于本 PRD 范围。
- 配置文件不存在时，工具使用内置默认值并正常工作。
- V0.1 不提供交互式配置向导；提供非交互式 `prox config path/show/init/validate` 命令。

### 9.3 配置优先级

从高到低：

1. 当前命令明确传入的参数；
2. 用户配置文件；
3. 内置默认值。

V0.1 不通过大量专用环境变量覆盖配置，避免再次引入隐式状态。

## 10. 环境变量行为

### 10.1 启用时设置的变量

`prox on` 和 `prox run` 必须为作用域内设置：

```text
http_proxy
https_proxy
all_proxy
HTTPS_PROXY
ALL_PROXY
no_proxy
NO_PROXY
```

其中 `http_proxy`、`https_proxy` 和 `all_proxy` 在默认配置下均使用：

```text
http://127.0.0.1:7890
```

`https_proxy` 中的 `http://` 表示客户端使用 HTTP 协议连接代理，并不表示只能访问 HTTP 网站。HTTPS 目标通过 HTTP CONNECT 建立隧道。

### 10.2 `HTTP_PROXY`

V0.1 不主动设置大写 `HTTP_PROXY`，也不主动删除用户已有的 `HTTP_PROXY`。

如果启用时发现它已经存在，`prox` 应提示兼容性警告，但不得擅自修改。

### 10.3 环境恢复

- `prox on` 必须先保存所有受管理变量的原始状态。
- `prox off` 必须恢复原始值和“是否存在”状态。
- 如果变量在启用后被手动修改，状态显示为 `drifted`。
- V0.1 中，执行 `prox off` 仍恢复启用前快照，并明确提示检测到变量漂移。
- `prox` 激活期间，受管理的变量视为由 `prox` 管理。

## 11. 命令需求

### 11.1 `prox init bash`

#### 目的

为 Bash 提供当前 Shell 集成，使 `prox on` 和 `prox off` 能够修改调用它们的 Shell 环境。

#### 使用方式

```bash
eval "$(prox init bash)"
```

#### 需求

- `INIT-01`：命令只能向标准输出写入 Bash 集成代码。
- `INIT-02`：警告和错误必须写入标准错误，避免污染可执行输出。
- `INIT-03`：重复初始化必须安全且幂等。
- `INIT-04`：生成内容不得执行网络请求。
- `INIT-05`：来自配置的值必须经过严格 Shell 转义，不能形成命令注入。
- `INIT-06`：集成层只负责父 Shell 环境修改，不复制健康检查等核心业务逻辑。

### 11.2 `prox on`

#### 目的

安全地为当前 Shell 及其后续子进程启用代理。

#### 预期流程

1. 读取配置；
2. 验证代理 URL；
3. 对代理主机和端口执行快速连接检查；
4. 保存当前环境快照；
5. 一次性设置受管理的环境变量；
6. 标记当前 Shell 为 `active`；
7. 输出简短结果。

#### 需求

- `ON-01`：配置验证失败时，不得修改任何环境变量。
- `ON-02`：代理端点不可达时，不得修改任何环境变量。
- `ON-03`：环境修改必须具有事务性，不得留下部分变量已设置的中间状态。
- `ON-04`：首次启用前必须保存环境快照。
- `ON-05`：已经使用相同配置启用时，再次执行应成功且不重复覆盖快照。
- `ON-06`：当前状态为 `drifted` 时，不得静默覆盖，应提示用户先执行 `prox off`。
- `ON-07`：成功信息必须包含经过脱敏的代理地址。
- `ON-08`：`prox on` 不执行完整外网检查，以保证速度和可预测性。

#### 成功示例

```text
$ prox on
Proxy enabled: http://127.0.0.1:7890
```

#### 失败示例

```text
$ prox on
Proxy endpoint is unavailable: 127.0.0.1:7890
Environment unchanged.
```

### 11.3 `prox off`

#### 目的

退出 `prox` 管理状态并恢复启用前的环境。

#### 需求

- `OFF-01`：必须按照环境快照恢复原始状态。
- `OFF-02`：原本不存在的变量应被删除，原本存在的变量应恢复原值。
- `OFF-03`：当前未启用时执行必须是安全的幂等操作。
- `OFF-04`：检测到变量漂移时应警告，但仍恢复快照。
- `OFF-05`：完成后必须删除当前 Shell 中由 `prox` 创建的内部状态。
- `OFF-06`：不得访问网络。

#### 示例

```text
$ prox off
Proxy disabled. Previous environment restored.
```

### 11.4 `prox status`

#### 目的

查看当前 Shell 是否由 `prox` 启用代理。

#### 需求

- `STATUS-01`：不得发起网络请求。
- `STATUS-02`：必须区分 `inactive`、`active` 和 `drifted`。
- `STATUS-03`：激活时显示代理地址，但必须隐藏 URL 中可能存在的敏感信息。
- `STATUS-04`：应显示配置来源是内置默认值还是用户配置文件。
- `STATUS-05`：不得把“已激活”描述为“代理健康”。

#### 示例

```text
$ prox status
State: active
Proxy: http://127.0.0.1:7890
Health: not checked
```

未启用时：

```text
$ prox status
State: inactive
```

### 11.5 `prox check`

#### 目的

验证配置的代理是否能够完成真实的外网请求。

#### 命令形式

```bash
prox check
prox check --local
prox check --url <URL>
prox check --quiet
```

#### 默认检查流程

1. 验证配置；
2. 建立到代理端点的 TCP 连接；
3. 显式通过该代理访问检查 URL；
4. 禁用 `NO_PROXY` 对此次检查的绕过作用；
5. 对 HTTPS 目标验证隧道、TLS 和 HTTP 响应；
6. 验证返回状态码；
7. 输出阶段结果和总耗时。

#### 需求

- `CHECK-01`：检查必须显式使用配置中的代理，不得依赖当前 Shell 是否已启用代理。
- `CHECK-02`：即使当前状态为 `inactive`，也必须可以执行检查。
- `CHECK-03`：`--local` 只执行配置和端点检查，不访问外网。
- `CHECK-04`：默认使用 GET 请求访问 `check_url`，不使用 HEAD 请求。
- `CHECK-05`：默认检查 URL 必须返回极小或空响应体。
- `CHECK-06`：不得将响应正文输出到终端。
- `CHECK-07`：必须实施连接超时和总请求超时。
- `CHECK-08`：默认不重试，以便结果快速、明确。
- `CHECK-09`：使用 `--url` 时，HTTP 200～399 视为目标可达；未指定 `--url` 时按配置的 `expected_status` 判断。
- `CHECK-10`：`--quiet` 不输出成功信息，只通过退出状态表达结果；失败原因仍可写入标准错误。
- `CHECK-11`：目标请求失败时，应报告“测试目标访问失败”，不得仅凭单一目标就断言整个代理永久失效。
- `CHECK-12`：应尽可能区分超时、连接拒绝、代理认证、DNS、TLS 和 HTTP 状态异常。

#### 成功示例

```text
$ prox check
Proxy: http://127.0.0.1:7890

PASS  Config       valid
PASS  Endpoint     127.0.0.1:7890 reachable    1 ms
PASS  Internet     google.com returned 204     316 ms

Proxy is healthy.
```

#### 失败示例

```text
$ prox check
Proxy: http://127.0.0.1:7890

PASS  Config       valid
PASS  Endpoint     127.0.0.1:7890 reachable    1 ms
FAIL  Internet     request timed out           8.0 s

The proxy endpoint is reachable, but the test request failed.
```

### 11.6 `prox run -- <command>`

#### 目的

只为一条命令及其子进程提供代理，不修改当前 Shell。

#### 使用方式

```bash
prox run -- curl https://www.google.com
prox run -- git clone https://github.com/example/project.git
```

#### 需求

- `RUN-01`：运行前必须验证配置和代理端点。
- `RUN-02`：代理端点不可达时，不得启动目标命令。
- `RUN-03`：只修改目标进程及其子进程的环境。
- `RUN-04`：目标命令结束后，当前 Shell 环境必须保持不变。
- `RUN-05`：目标命令启动后，必须原样连接标准输入、标准输出和标准错误。
- `RUN-06`：必须尽可能透明地传递终止信号。
- `RUN-07`：目标命令正常启动后，`prox run` 必须返回目标命令的退出状态。
- `RUN-08`：默认只检查代理端点，不执行额外外网请求。
- `RUN-09`：`--` 作为 `prox` 参数和目标命令之间的明确分隔符。

## 12. 状态与生命周期

### 12.1 状态转换

```text
inactive
   │
   │ prox on，预检查成功
   ▼
active
   │
   ├── 受管理变量被外部修改 ──> drifted
   │
   └── prox off ─────────────> inactive

drifted
   │
   └── prox off，恢复快照 ───> inactive
```

### 12.2 生命周期约束

- 状态仅存在于当前 Shell 会话。
- 关闭终端后不需要保留激活状态或环境快照。
- 新打开的终端默认处于 `inactive`。
- 从已启用代理的 Shell 启动的子 Shell，可以继承代理变量，但不应自动伪装成由 `prox` 管理的完整激活状态。
- V0.1 不支持嵌套多层 `prox on`。
- 代理服务在启用后退出时，`prox` 不在后台自动改变环境；用户通过 `prox check` 获取最新健康状态。

## 13. 输出与交互规范

### 13.1 输出原则

- 输出首先说明结果，再给原因。
- 成功信息应简短，错误信息应可操作。
- 标准输出用于正常结果。
- 标准错误用于警告和失败原因。
- 非交互环境下不得依赖颜色表达含义。
- 输出到非终端目标时，不应包含无法关闭的 ANSI 控制字符。
- V0.1 的人类可读文本不作为稳定机器接口；脚本应依据退出状态。

### 13.2 地址脱敏

如果输入中意外包含用户信息，显示时必须脱敏：

```text
http://user:****@proxy.example.com:8080
```

即使 V0.1 不支持代理认证，也不得在错误输出中完整泄露密码。

## 14. 退出状态

### 14.1 管理命令

`on`、`off`、`status`、`check` 和 `init` 使用：

| 状态码 | 含义 |
| ---: | --- |
| `0` | 操作成功；对于 `check` 表示检查通过 |
| `1` | 操作执行失败或健康检查未通过 |
| `2` | 命令使用错误或配置无效 |

### 14.2 `prox run`

- 在目标命令启动之前发生的 `prox` 错误返回 `125`。
- 目标命令成功启动后，返回目标命令本身的退出状态。
- 命令不存在、不可执行和信号终止应尽量保持所在平台的常规语义。

## 15. 错误处理

错误信息至少回答两个问题：

1. 哪个阶段失败；
2. 用户下一步可以做什么。

示例：

```text
Proxy endpoint is unavailable: 127.0.0.1:7890
Start your proxy application or update proxy_url, then run `prox check --local`.
```

```text
The proxy endpoint is reachable, but https://www.google.com/generate_204 timed out.
Check the selected proxy node or retry with `prox check --url <URL>`.
```

不得只输出模糊错误：

```text
Proxy failed.
```

## 16. 安全与隐私要求

- `SEC-01`：不得将配置值未经转义直接拼接为 Shell 命令。
- `SEC-02`：不得通过检查 URL 发送用户命令、目录、主机名或其他无关信息。
- `SEC-03`：不得记录用户通过 `prox run` 执行的命令。
- `SEC-04`：不得默认收集遥测或崩溃报告。
- `SEC-05`：不得关闭 TLS 证书验证来让健康检查通过。
- `SEC-06`：不得在终端输出中暴露代理密码。
- `SEC-07`：不得自动将代理变量传播到提权后的环境。
- `SEC-08`：配置错误时必须安全失败，不得回退到未声明的代理地址。

## 17. 兼容性说明

代理环境变量属于广泛使用的事实约定，但并非所有程序都以相同方式实现。

V0.1 应在文档中明确：

- `prox` 只保证正确设置其声明支持的环境变量；
- 目标程序是否读取这些变量由目标程序决定；
- 已经运行的进程不会因当前 Shell 环境变化而自动更新；
- systemd 服务、Docker daemon、桌面应用和 `sudo` 不属于当前 Shell 的普通子进程场景；
- 如果某个工具忽略代理环境变量，用户需要查阅该工具自身的代理配置方式。

## 18. 验收标准

### 18.1 基本流程

- `AC-01`：无配置文件时，工具使用 `http://127.0.0.1:7890`。
- `AC-02`：代理端口监听时，`prox on` 成功设置所有受管理变量。
- `AC-03`：代理端口未监听时，`prox on` 返回失败，环境与执行前完全一致。
- `AC-04`：`prox on` 后执行 `prox off`，所有受管理变量恢复原始存在状态和原值。
- `AC-05`：连续执行两次 `prox on` 不覆盖第一次保存的环境快照。
- `AC-06`：未启用时执行 `prox off` 不报破坏性错误。

### 18.2 状态

- `AC-07`：`prox status` 不产生任何网络连接。
- `AC-08`：手动修改已启用的代理变量后，状态显示为 `drifted`。
- `AC-09`：`status` 不把 `active` 表述为 `healthy`。

### 18.3 健康检查

- `AC-10`：`prox check` 在当前 Shell 未启用代理时仍可工作。
- `AC-11`：完整检查显式使用配置代理，不受当前 `NO_PROXY` 影响。
- `AC-12`：默认目标返回 204 时，检查退出状态为 0。
- `AC-13`：代理端点连接失败时，结果为 `endpoint_unreachable`。
- `AC-14`：代理端点可达而测试请求超时时，结果为 `target_failed`。
- `AC-15`：达到总超时后检查必须结束，不得无限等待。
- `AC-16`：`prox check --quiet` 成功时不输出普通信息。

### 18.4 单命令代理

- `AC-17`：`prox run` 的子进程能够读取代理变量。
- `AC-18`：`prox run` 结束后，父 Shell 的代理变量完全不变。
- `AC-19`：目标命令返回非零状态时，`prox run` 返回相同状态。
- `AC-20`：代理端点不可达时，目标命令不会被启动。

### 18.5 安全

- `AC-21`：恶意或畸形配置值不能通过 Shell 集成执行额外命令。
- `AC-22`：包含用户信息的代理 URL 在所有输出中均被脱敏。
- `AC-23`：健康检查保持 TLS 证书验证开启。

## 19. 发布要求

V0.1 发布物至少包括：

- `prox` 可执行入口；
- Bash 集成方式；
- 默认配置说明；
- 安装与卸载说明；
- 五个日常命令的快速开始；
- 支持范围与非目标说明；
- License；
- 最小自动化测试；
- 版本号输出，例如 `prox version` 或 `prox --version`。

README 的首屏示例应保持简短：

```bash
prox on
prox check
prox run -- git clone https://github.com/example/project.git
prox off
```

## 20. 后续版本候选

以下能力可以进入后续规划，但不应阻塞 V0.1：

1. 命名 Profile，例如 `home`、`office` 和 `clash`；
2. SOCKS5 / SOCKS5H；
3. Zsh、Fish 和 PowerShell；
4. JSON 格式检查结果；
5. 多检查目标与失败回退；
6. 出口 IP 检查；
7. 代理认证；
8. `prox doctor` 诊断不同 CLI 的代理支持情况；
9. Shell 补全；
10. Homebrew、deb、rpm 等安装渠道；
11. 自动检测本地常见代理端口；
12. 项目目录级代理配置；
13. 临时启动一个带代理环境的新 Shell。

## 21. 技术实现边界

本 PRD 不指定实现语言、框架或第三方依赖。

具体实现可以是 Shell、Go、Rust、Python 或其他形式，但必须满足以下产品约束：

1. `prox on` 和 `prox off` 能改变当前 Shell，而不仅是工具自身的子进程。
2. 核心检查逻辑不能在多个 Shell 适配层中重复实现。
3. Shell 集成必须安全转义所有动态值。
4. `prox run` 必须保持良好的进程透明性。
5. 健康检查不得依赖当前已经设置的代理环境变量。
6. 产品行为和退出状态必须符合本 PRD，而不因实现语言不同发生变化。

实现语言选择、内部模块划分、配置序列化格式和第三方库选择应在后续技术设计文档中确定。

## 22. V0.1 完成定义

只有同时满足以下条件，V0.1 才可以发布：

- 六个范围内命令全部实现；
- 所有 V0.1 验收标准通过；
- 代理不可用时不会污染用户环境；
- `off` 已验证可以恢复既有环境；
- `run` 已验证不影响父 Shell；
- 健康检查能够区分端点失败和目标请求失败；
- README 与实际命令行为一致；
- 已明确标注仅支持 Linux + Bash + HTTP Proxy；
- 没有将后续版本功能提前混入 V0.1。

---

## 附录 A：推荐的首版体验

安装集成后：

```bash
$ prox status
State: inactive

$ prox check
Proxy: http://127.0.0.1:7890

PASS  Config       valid
PASS  Endpoint     127.0.0.1:7890 reachable    1 ms
PASS  Internet     google.com returned 204     316 ms

Proxy is healthy.

$ prox on
Proxy enabled: http://127.0.0.1:7890

$ prox status
State: active
Proxy: http://127.0.0.1:7890
Health: not checked

$ prox off
Proxy disabled. Previous environment restored.
```

对于大多数一次性任务，推荐：

```bash
prox run -- <command>
```

而不是长期保持整个终端处于代理状态。

## 附录 B：术语

| 术语 | 含义 |
| --- | --- |
| 代理端点 | 代理服务监听的主机和端口 |
| 激活 | 将代理变量应用到当前 Shell |
| 环境快照 | 激活前受管理环境变量的原始状态 |
| 漂移 | 激活后受管理变量被其他操作修改 |
| 本地检查 | 只验证配置和代理端点连接 |
| 完整检查 | 通过指定代理完成一次真实 HTTPS 请求 |
| 单命令代理 | 只将代理环境提供给一个命令及其子进程 |
