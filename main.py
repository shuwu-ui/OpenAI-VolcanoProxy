from flask import Flask, request, jsonify, Response
import requests
import json

app = Flask(__name__)

# 模型名称映射表
MODEL_MAPPING = {
    'deepseek-r1': 'ep-20250218144306-sshj5',
    'deepseek-v3': 'ep-20250218170310-pvmdb',
    'deepseek-r1-distill-qwen-7b': 'ep-20250218182157-7rdv4',
    'deepseek-r1-distill-qwen-32b': 'ep-20250218182233-vxkqv'
}


def fangzou(api_key, model, messages, stream=True, **kwargs):
    """支持透传所有OpenAI兼容参数的后端调用函数"""
    # 模型名称转换
    model = MODEL_MAPPING.get(model, 'ep-20250218144306-sshj5')

    base_url = "https://ark.cn-beijing.volces.com/api/v3"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {api_key}",
    }

    # 构造包含所有参数的请求体
    payload = {
        "model": model,
        "messages": messages,
        "stream": stream,
        **kwargs  # 包含所有额外参数
    }

    try:
        response = requests.post(
            f"{base_url}/chat/completions",
            headers=headers,
            json=payload,
            timeout=1800,
            stream=stream
        )
        response.raise_for_status()

        if stream:
            # 流式响应处理
            def generate():
                for line in response.iter_lines():
                    if line:
                        yield line.decode('utf-8') + '\n'

            return generate()
        else:
            # 非流式响应处理
            return [json.dumps(response.json())]

    except Exception as e:
        error_msg = json.dumps({"error": str(e)})
        return [error_msg]


@app.route('/v1/chat/completions', methods=['POST'])
def handle_request():
    """处理OpenAI格式请求并保持参数兼容性"""
    # 鉴权处理
    auth_header = request.headers.get('Authorization')
    if not auth_header or not auth_header.startswith('Bearer '):
        return jsonify({"error": "Invalid Authorization header"}), 401
    api_key = auth_header.split(' ', 1)[1].strip()

    # 参数校验
    try:
        data = request.get_json()
        if 'model' not in data or 'messages' not in data:
            return jsonify({"error": "Missing required parameters"}), 400
    except:
        return jsonify({"error": "Invalid JSON format"}), 400

    # 提取基础参数
    stream = data.get('stream', False)

    # 分离需要特殊处理的参数
    base_params = {
        'model': data['model'],
        'messages': data['messages'],
        'stream': stream
    }

    # 收集其他所有参数
    extra_params = {k: v for k, v in data.items() if k not in base_params}

    # 生成响应
    return Response(
        fangzou(
            api_key=api_key,
            **base_params,
            **extra_params
        ),
        mimetype='text/event-stream' if stream else 'application/json'
    )


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
