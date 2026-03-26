import requests
import time

# 测试配置
BASE_URL = "http://localhost:8888/api/open/connect/session"
TOKEN = "test_token"  # 测试用的 token


def test_connect_session():
    """测试 /api/open/connect/session 接口"""
    print("=== 测试 /api/open/connect/session 接口 ===")
    
    # 构建请求头
    headers = {
        "Authorization": f"Bearer {TOKEN}"
    }
    
    try:
        # 发送请求
        response = requests.get(BASE_URL, headers=headers)
        
        # 打印响应信息
        print(f"状态码: {response.status_code}")
        print(f"响应内容: {response.json()}")
        
        # 检查响应
        if response.status_code == 200:
            result = response.json().get("result")
            if result:
                print("\n✓ 成功获取 ticket")
                print(f"ticket: {result.get('ticket')}")
                print(f"expire_at: {result.get('expire_at')}")
                print(f"expires_in: {result.get('expires_in')}")
                print(f"redirect_url: {result.get('redirect_url')}")
            else:
                print("\n✗ 响应中没有 result 字段")
        else:
            print(f"\n✗ 请求失败: {response.status_code}")
            
    except Exception as e:
        print(f"测试失败: {e}")


if __name__ == "__main__":
    test_connect_session()
