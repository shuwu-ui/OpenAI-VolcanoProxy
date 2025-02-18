# DeepSeek 模型代理服务

基于 Gin 框架的 HTTP 代理服务，实现 OpenAI 格式 API 到火山引擎模型服务的协议转换

## 功能特性

- 🚀 模型名称映射机制（支持多模型别名配置）
- 🌊 完整流式响应支持（Server-Sent Events）
- ⚡ 自动透传额外参数到后端服务
- 🔒 请求鉴权与错误处理
- 📊 请求日志记录与监控

## 模型映射表

| 客户端请求模型名称                  | 火山引擎实际模型 ID                 |
|------------------------------------|-----------------------------------|
| `deepseek-r1`                     | `ep-20250218144306-sshj5`        |
| `deepseek-v3`                     | `ep-20250218170310-pvmdb`        |
| `deepseek-r1-distill-qwen-7b`     | `ep-20250218182157-7rdv4`        |
| `deepseek-r1-distill-qwen-32b`    | `ep-20250218182233-vxkqv`        |

## 快速开始

### 环境要求
- Go 1.20+
- Gin 框架

### 安装运行
```bash
go get github.com/gin-gonic/gin
go run main.go
```

### 服务调用示例
```bash
curl --location 'http://localhost:5000/v1/chat/completions' \
--header 'Authorization: Bearer YOUR_VOLC_ACCESS_KEY' \
--header 'Content-Type: application/json' \
--data '{
    "model": "deepseek-v3",
    "messages": [{"role": "user", "content": "你好"}],
    "temperature": 0.7,
    "stream": true
}'
```

## 请求参数说明

### 必填参数
- `model`: 使用的模型名称（参考上方映射表）
- `messages`: 消息数组，格式为 `[{role, content}]`
- `stream`: 是否启用流式响应

### 扩展参数
支持所有火山引擎原生参数：
- `temperature`
- `top_p`
- `max_tokens`
- `...` 等其他官方参数

## 注意事项

1. **API 密钥安全**
   - 使用火山引擎的 Access Key 作为鉴权凭证
   - 务必通过 HTTPS 传输敏感信息
   - 建议在服务端设置 IP 白名单

2. **模型兼容性**
   - 未配置的模型名称会自动使用 `deepseek-r1`
   - 模型列表更新需同步修改服务端的 `modelMapping`

3. **流式响应**
   - 保持连接直到收到 `[DONE]` 事件
   - 需要客户端正确处理 chunked 响应
   - 建议设置 30-60 秒的超时时间

4. **性能建议**
   - 生产环境需部署负载均衡
   - 建议设置合理的重试机制
   - 监控 5xx 错误和响应延迟

5. **计费说明**
   - 实际消耗火山引擎的 API 调用额度
   - 流式响应按完整请求 tokens 计费
   - 建议在火山控制台设置用量告警

## 响应格式

### 普通响应
```json
{
    "id": "chatcmpl-3QfHb5kAOZw9",
    "object": "chat.completion",
    "created": 1689000000,
    "choices": [{
        "index": 0,
        "message": {
            "role": "assistant",
            "content": "你好！有什么可以帮助您的吗？"
        }
    }]
}
```

### 流式响应
```
data: {"id":"chatcmpl-3QfHb5kAOZw9","object":"chat.completion.chunk"...}

data: [DONE]
```

## 错误代码
| 状态码 | 说明                      |
|--------|-------------------------|
| 401    | 无效的授权信息            |
| 400    | 请求参数格式错误          |
| 502    | 后端服务不可用            |
| 504    | 后端服务响应超时          |
