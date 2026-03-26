import requests
import time
from Crypto.Cipher import AES
import binascii

def decrypt_aes_cbc(encrypted_hex, key):
    """使用 AES-CBC 解密 ticket"""
    # 将 key 转换为字节数组
    key_bytes = key.encode('utf-8')
    
    # 确保密钥长度为 16、24 或 32 字节
    if len(key_bytes) not in [16, 24, 32]:
        raise ValueError("Invalid key length")
    
    # 使用密钥的前 16 字节作为 IV
    iv = key_bytes[:16]
    
    # 将 hex 编码的密文转换为字节数组
    ciphertext = binascii.unhexlify(encrypted_hex)
    
    # 创建 AES 解密器
    cipher = AES.new(key_bytes, AES.MODE_CBC, iv)
    
    # 解密
    plaintext = cipher.decrypt(ciphertext)
    
    # 移除 PKCS5 填充
    padding = plaintext[-1]
    plaintext = plaintext[:-padding]
    
    # 将字节数组转换为字符串
    return plaintext.decode('utf-8')

def test_decrypt_ticket():
    """测试解密 ticket"""
    url = "http://localhost:8888/api/open/connect/session"
    
    # 发送请求获取 ticket
    headers = {"Authorization": "Bearer test_token"}
    response = requests.get(url, headers=headers)
    
    if response.status_code != 200:
        print(f"请求失败: {response.status_code}")
        print(f"响应内容: {response.json()}")
        return
    
    data = response.json()
    ticket = data['result']['ticket']
    expire_at = data['result']['expire_at']
    
    print(f"获取到的 ticket: {ticket}")
    print(f"过期时间戳: {expire_at}")
    
    # 使用默认密钥解密
    key = "00112233445566778899aabbccddeeff"
    try:
        plaintext = decrypt_aes_cbc(ticket, key)
        print(f"解密后的明文: {plaintext}")
        
        # 验证解密结果
        parts = plaintext.split('|')
        if len(parts) == 3:
            uid, bid, timestamp = parts
            print(f"UID: {uid}")
            print(f"BID: {bid}")
            print(f"时间戳: {timestamp}")
            
            # 验证时间戳
            if int(timestamp) == expire_at - 600:
                print("\n✓ 时间戳验证通过")
            else:
                print("\n✗ 时间戳验证失败")
        else:
            print("\n✗ 解密结果格式不正确")
    except Exception as e:
        print(f"\n✗ 解密失败: {e}")

if __name__ == "__main__":
    print("开始测试解密 ticket...")
    test_decrypt_ticket()
    print("测试完成！")
