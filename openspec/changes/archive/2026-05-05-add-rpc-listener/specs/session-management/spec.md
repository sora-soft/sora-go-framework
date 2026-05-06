## ADDED Requirements

### Requirement: acceptLoop 接受连接并注册 session
Listener SHALL 运行 acceptLoop goroutine，持续调用 `TransportListener.Accept(ctx)` 获取新 Connection，为每个 Connection 分配唯一 sessionId，注册到 sessions map。

#### Scenario: 接受新连接
- **WHEN** acceptLoop 收到来自 TransportListener.Accept 的 Connection
- **THEN** 分配 uuid 作为 sessionId，调用 `newConnector(sessionId, conn)` 注册

#### Scenario: Accept 返回错误
- **WHEN** TransportListener.Accept 返回错误
- **THEN** acceptLoop 继续运行（不退出），等待下次 Accept

#### Scenario: context 取消
- **WHEN** Listener 的内部 context 被取消（Stop 调用）
- **THEN** acceptLoop 退出

### Requirement: Connection.Serve 在 acceptLoop 中调用
acceptLoop 注册 session 后 SHALL 调用 `conn.Serve()` 完成 server-side 初始化。Serve 负责等待客户端 codec 协商并启动 readLoop。

#### Scenario: Serve 成功
- **WHEN** conn.Serve() 成功完成 codec 协商
- **THEN** Connection 进入 Ready 状态，触发 OnSessionOpen callback

#### Scenario: Serve 失败
- **WHEN** conn.Serve() 失败（codec 不支持、handshake 超时等）
- **THEN** 不注册该 session，触发 OnSessionClose callback

### Requirement: Session 生命周期回调
Listener SHALL 支持以下回调：`OnSessionOpen(conn)`、`OnSessionClose(conn, error)`。

#### Scenario: session 建立
- **WHEN** Connection 通过 Serve 成功进入 Ready
- **THEN** 调用 OnSessionOpen(conn)

#### Scenario: session 断开
- **WHEN** Connection 进入 Error 或 Stopped 状态
- **THEN** 从 sessions map 移除，调用 OnSessionClose(conn, err)

### Requirement: Request/Notify 路由通过 callback
Listener SHALL 通过 callback 将收到的 Request 和 Notify 路由给上层：`OnRequest(conn, *ReqPacketData)`、`OnNotify(conn, *NotifyPacketData)`。

#### Scenario: 收到 Request
- **WHEN** 某个 session 的 readLoop 收到 Request packet
- **THEN** 调用 OnRequest(conn, req)

#### Scenario: 收到 Notify
- **WHEN** 某个 session 的 readLoop 收到 Notify packet
- **THEN** 调用 OnNotify(conn, notify)

### Requirement: newConnector 注册 session
Listener SHALL 提供 `newConnector(sessionId string, conn *Connection)` 方法，将 Connection 注册到 sessions map。

#### Scenario: 注册新 session
- **WHEN** 调用 `newConnector("abc", conn)`
- **THEN** sessions map 中 "abc" → conn

### Requirement: GetSession 查询 session
Listener SHALL 提供 `GetSession(sessionId string) (*Connection, bool)` 方法查询指定 session。

#### Scenario: session 存在
- **WHEN** 调用 `GetSession("abc")` 且 session 存在
- **THEN** 返回 (conn, true)

#### Scenario: session 不存在
- **WHEN** 调用 `GetSession("xyz")` 且 session 不存在
- **THEN** 返回 (nil, false)
