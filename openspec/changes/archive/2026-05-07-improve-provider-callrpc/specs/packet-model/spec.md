## ADDED Requirements

### 需求:Response 类型
系统必须在 `pkg/rpc/packet` 包中定义 `Response[T any]` struct，包含 `Error *PayloadError` 和 `Result T` 两个导出字段。该类型描述 RPC 响应的线路格式 `{error, result}`。

#### 场景:反序列化成功响应
- **当** 从 Packet 解码 `Response[EchoResp]`，payload 为 `{"error": null, "result": {"message": "hi"}}`
- **那么** `resp.Error` 为 nil，`resp.Result.Message` 为 `"hi"`

#### 场景:反序列化错误响应
- **当** 从 Packet 解码 `Response[any]`，payload 为 `{"error": {"code": "ERR_METHOD_NOT_FOUND", ...}, "result": null}`
- **那么** `resp.Error.Code` 为 `"ERR_METHOD_NOT_FOUND"`，`resp.Result` 为零值

### 需求:PayloadError 类型
系统必须在 `pkg/rpc/packet` 包中定义 `PayloadError` struct，包含 `Code string`、`Message string`、`Level int`、`Name string`、`Args any` 五个导出字段（含 `json` 标签）。

#### 场景:PayloadError 序列化
- **当** 序列化 `PayloadError{Code: "ERR_INTERNAL", Message: "fail", Level: 1, Name: "InternalError"}`
- **那么** JSON 输出包含 `"code": "ERR_INTERNAL"`、`"message": "fail"`、`"level": 1`、`"name": "InternalError"`
