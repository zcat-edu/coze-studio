import requests

# 测试配置
BASE_URL = "http://localhost:8888/api/open/connect/session"


def test_connect_session():
    """测试 /api/open/connect/session 接口"""
    print("=== 测试 /api/open/connect/session 接口 ===")
    
    try:
        # 发送请求
        response = requests.get(BASE_URL)
        
        # 打印响应信息
        print(f"状态码: {response.status_code}")
        print(f"响应内容: {response.text}")
        
    except Exception as e:
        print(f"测试失败: {e}")


if __name__ == "__main__":
    test_connect_session()
