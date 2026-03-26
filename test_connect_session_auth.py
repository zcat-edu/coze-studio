import requests
import time

def test_connect_session_with_auth():
    """测试带认证的 connect session 接口"""
    url = "http://localhost:8888/api/open/connect/session"
    
    # 测试1: 不带 Authorization 头
    print("测试1: 不带 Authorization 头")
    response = requests.get(url)
    print(f"状态码: {response.status_code}")
    print(f"响应内容: {response.json()}")
    print()
    
    # 测试2: 带无效格式的 Authorization 头
    print("测试2: 带无效格式的 Authorization 头")
    headers = {"Authorization": "invalid_token"}
    response = requests.get(url, headers=headers)
    print(f"状态码: {response.status_code}")
    print(f"响应内容: {response.json()}")
    print()
    
    # 测试3: 带正确格式的 Authorization 头
    print("测试3: 带正确格式的 Authorization 头")
    headers = {"Authorization": "Bearer test_token"}
    response = requests.get(url, headers=headers)
    print(f"状态码: {response.status_code}")
    print(f"响应内容: {response.json()}")
    
    # 验证响应格式
    if response.status_code == 200:
        data = response.json()
        assert data.get("code") == 200
        assert data.get("message") == "success"
        assert "result" in data
        result = data["result"]
        assert "ticket" in result
        assert "expire_at" in result
        assert "expires_in" in result
        assert "redirect_url" in result
        assert result["expires_in"] == 600
        print("\n✓ 响应格式验证通过")
    
if __name__ == "__main__":
    print("开始测试带认证的 connect session 接口...")
    test_connect_session_with_auth()
    print("测试完成！")
