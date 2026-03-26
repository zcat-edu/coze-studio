import requests
import json

# 测试平台A登录端点
url = "http://localhost:8888/api/passport/web/platform-a/login/"

# 准备请求数据
payload = {
    "encrypted_session_key": "test_encrypted_session_key"
}

headers = {
    "Content-Type": "application/json"
}

# 发送请求
response = requests.post(url, data=json.dumps(payload), headers=headers)

# 打印响应
print(f"Status Code: {response.status_code}")
print("Response Content:")
print(json.dumps(response.json(), indent=2, ensure_ascii=False))