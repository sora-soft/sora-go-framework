## Context

当前 `TCPListener` 的端口绑定发生在 `NewTCPListener` 构造函数中，导致构造即分配系统资源（`net.Listen`）。`Listener.Start()` 只负责启动 accept loop，无法控制绑定时机。

架构层次：
```
Listener (RPC层)          → Start() 启动 accept loop
  └─ TransportListener    → Accept() / Close() / GetMetaInfo()
       └─ TCPListener     → 构造时 net.Listen() ← 问题所在
```

## Goals / Non-Goals

**Goals:**
- 将端口绑定从构造延迟到 `Listener.Start()` 调用链中
- `TransportListener` interface 新增 `StartListen(ctx) error`
- `NewTCPListener` 变为纯构造（无副作用），`StartListen` 执行绑定
- `Listener.Start()` 在 accept loop 之前调用 `tl.StartListen(ctx)`

**Non-Goals:**
- 不重构命名（Listener vs TCPListener 命名问题另行处理）
- 不修改 `TCPListener.Accept()` 内部创建 `Connection` 的逻辑
- 不引入新的传输层实现

## Decisions

### Decision 1: TransportListener interface 新增 StartListen 方法

`StartListen(ctx context.Context) error` 负责执行底层地址绑定。与 `Accept`/`Close` 并列。

**替代方案：**
- 新建 `StartableTransportListener` 接口 — 增加类型断言复杂度，目前所有实现都需要 StartListen，没有必要拆分接口
- 使用 `io.Closer` + 自定义 `Starter` — 过度设计，一个接口三个方法足够清晰

**选择理由：** 当前仅有一个实现（TCP），未来新增传输层时都需要延迟绑定。统一放在 `TransportListener` 中最简单。

### Decision 2: TCPListener.StartListen 执行端口绑定

将 `NewTCPListener` 中 `net.Listen` 的逻辑移入 `StartListen`。`NewTCPListener` 仅校验参数、保存 opts。

**状态保护：** `Accept` 在 `StartListen` 未调用时返回错误（`net.Listener` 为 nil）。`StartListen` 内部设置 `listener` 字段后标记 `started`。

### Decision 3: Listener.Start 调用链

```
Listener.Start(ctx)
  ├─ SetState(Starting)
  ├─ tl.StartListen(ctx)      ← 新增
  ├─ go acceptLoop()
  └─ SetState(Ready)
```

`StartListen` 失败时直接进入 Error 状态并返回错误，不启动 accept loop。

## Risks / Trade-offs

- **[BREAKING]** `TransportListener` interface 变更，所有实现需新增 `StartListen` → 目前仅 TCP 一处，影响可控
- **[Risk]** `Accept` 在 `StartListen` 之前被调用 → 通过 `started` 标志保护，返回明确错误
- **[Risk]** `StartListen` 多次调用 → 幂等处理，第二次调用返回 nil 或错误
