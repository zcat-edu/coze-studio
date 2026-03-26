Zcat 单点登录 Coze 接口文档

接口说明

- 用途：zcat 用户已登录后，点击 Coze 入口，自动进入 Coze。
- 行为：Coze 会根据 zcat 传入的用户标识自动登录；若本地无账号，则自动注册后登录。
- 登录入口：GET /api/auth/third_login

对应实现见 backend/api/handler/coze/passport_service.go:264。

请求地址

https://{coze-domain}/api/auth/third_login

请求方式

GET

请求参数

- ticket
    - 类型：string
    - 必填：是
    - 含义：加密后的登录票据
- timestamp
    - 类型：string 或 int64
    - 必填：是
    - 含义：当前 Unix 秒级时间戳
- sign
    - 类型：string
    - 必填：是
    - 含义：请求签名，hex 字符串
- platform
    - 类型：string
    - 必填：否
    - 固定值：zcat
    - 说明：不传时后端默认按 zcat 处理；建议始终传 zcat

已废弃参数

- app_id
    - 不再要求提供
    - 现在可以完全不传

ticket 明文格式
解密后的明文格式固定为：

uid|timestamp

示例：

10086|1743043200

字段说明：

- uid：zcat 用户唯一标识，必须稳定且全局唯一
- timestamp：与本次登录票据对应的秒级时间戳

ticket 加密规则

- 算法：AES-GCM
- 密钥：双方约定的 AES_KEY
- 密钥长度：必须是 16 / 24 / 32 字节
- nonce：12 字节随机数
- 组装方式：
    - 原始字节 = nonce + ciphertext
    - 最终传输值 = base64url 编码后的字符串

sign 签名规则

- 算法：HMAC-SHA256
- 密钥：双方约定的 SIGN_SECRET
- 签名原文：

ticket + timestamp

- 输出格式：hex 小写字符串

伪代码：

sign = hex(HMAC_SHA256(ticket + timestamp, SIGN_SECRET))

请求示例

GET https://{coze-domain}/api/auth/third_login?ticket=XXX&timestamp=1743043200&sign=YYY&platform=zcat

zcat 侧接入流程

1. 用户在 zcat 已登录。
2. zcat 后端获取当前用户 uid。
3. 生成当前时间戳 timestamp。
4. 拼接明文：uid|timestamp。
5. 用 AES_KEY 做 AES-GCM 加密，得到 ticket。
6. 用 SIGN_SECRET 计算 sign = HMAC_SHA256(ticket + timestamp)。
7. zcat 将浏览器重定向到：

   https://{coze-domain}/api/auth/third_login?ticket=...&timestamp=...&sign=...&platform=zcat

Coze 侧处理逻辑

1. 校验 platform 是否为 zcat
2. 校验外层 timestamp
3. 校验 sign
4. 解密 ticket
5. 解析出 uid|timestamp
6. 根据 uid + zcat 查找本地用户
7. 若不存在则自动创建
8. 生成 Coze 的 session_key cookie
9. 跳转到 Coze 首页 /

自动注册/登录逻辑见 backend/api/handler/coze/passport_service.go:344 和 backend/application/user/user.go:481。

- HTTP 状态码：302 Found
- 响应头：
    - Set-Cookie: session_key=...
    - Location: /

说明：浏览器收到后会进入 Coze，并带上登录态。

失败响应

- 400 unsupported platform
    - platform 非 zcat
- 401 invalid timestamp
    - 时间戳格式不合法
- 401 expired
    - 外层时间戳过期
- 401 invalid sign
    - 签名错误
- 401 invalid ticket
    - ticket 解密失败
- 401 invalid ticket format
    - 解密后不是 uid|timestamp
- 401 expired inner ts
    - ticket 内时间戳过期
- 500 AES_KEY not configured
    - Coze 侧未配置解密密钥

时效要求

- 外层 timestamp 与 Coze 服务端时间差不能超过 60 秒
- ticket 内部的时间戳同样不能超过 60 秒

zcat 对接建议

- uid 不要使用可变字段，如昵称、邮箱
- 建议使用 zcat 用户主键或稳定 UUID
- 建议 always 传 platform=zcat
- 建议由 zcat 后端生成 ticket 和 sign，不要在前端生成