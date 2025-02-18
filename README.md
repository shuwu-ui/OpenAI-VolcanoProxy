# OpenAI-VolcanoProxy

基于Flask的中间层代理服务，用于将OpenAI格式的API请求转发至火山引擎的深度学习模型服务。

## 功能特性

- 🚀 无缝兼容OpenAI API格式
- 🔀 支持多模型名称映射（DeepSeek系列模型）
- 🌊 完整的流式响应支持
- 🔧 全参数透传能力
- 🔒 火山引擎API密钥认证

## 快速开始

### 环境要求
- Python 3.7+
- Flask 2.0+
- requests 2.25+

### 安装步骤
```bash
git clone https://github.com/yourusername/OpenAI-VolcanoProxy.git
cd OpenAI-VolcanoProxy
pip install -r requirements.txt
```

### 配置说明
1. 部署到可访问火山引擎服务的服务器
2. 服务默认运行在5000端口（通过`app.run()`参数可调）
3. 无需额外配置，直接使用火山引擎API密钥即可

### API使用
**基础请求示例：**
```bash
curl http://localhost:5000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_VOLC_AK" \  # 火山引擎API密钥
  -d '{
    "model": "deepseek-v3",
    "messages": [{"role": "user", "content": "你好"}],
    "temperature": 0.7
  }'
```

**流式响应示例：**
```bash
curl http://localhost:5000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_VOLC_AK" \
  -d '{
    "model": "deepseek-r1-distill-qwen-7b",
    "messages": [{"role": "user", "content": "解释量子计算"}],
    "stream": true
  }'
```

## 完整模型映射表
| 客户端模型名称                  | 火山引擎模型ID           |
|---------------------------------|--------------------------|
| `deepseek-r1`                   | `ep-20250218144306-sshj5`|
| `deepseek-v3`                   | `ep-20250218170310-pvmdb`|
| `deepseek-r1-distill-qwen-7b`   | `ep-20250218182157-7rdv4`|
| `deepseek-r1-distill-qwen-32b`  | `ep-20250218182233-vxkqv`|

## 参数支持
支持所有OpenAI兼容参数，包括但不限于：
- `temperature`：响应随机性（0-2）
- `top_p`：核心采样概率（0-1）
- `max_tokens`：最大输出token数（最高支持4000）
- `presence_penalty`：话题新鲜度（-2-2）
- `frequency_penalty`：重复惩罚（-2-2）

## 核心注意事项
1. **密钥认证**：必须使用有效的[火山引擎API Key](https://console.volcengine.com/)
2. **模型版本**：不同模型参数支持范围可能不同，建议参考官方文档
3. **超时机制**：默认30分钟超时，长文本生成建议适当调整
4. **错误处理**：HTTP 500错误表示火山引擎服务端异常
5. **流式响应**：`stream: true`时返回text/event-stream格式数据
6. **默认模型**：当使用未映射模型时自动回退到deepseek-r1

## 服务监控
建议监控以下指标：
- 请求成功率
- 平均响应延迟
- 模型调用分布
- 异常响应统计

## 授权协议
[MIT License](LICENSE)

---

> **模型选择建议**  
> 7B模型：适合通用场景和快速响应  
> 32B模型：适合复杂推理任务  
> R1系列：平衡精度与速度  
> V3系列：最新优化版本

> **技术支持**  
> 模型服务相关问题请联系火山引擎技术团队  
> 代理层问题请提交GitHub Issue
