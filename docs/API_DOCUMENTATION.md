# 游戏服务器 API 文档

## 服务器信息
- **服务器地址**: `ws://localhost:8080/ws`
- **协议**: WebSocket
- **数据格式**: JSON

## 消息格式

### 请求消息格式
```json
{
  "type": "消息类型",
  "action": "具体操作",
  "data": {}, 
  "requestId": "请求ID",
  "timestamp": 1640995200
}
```

### 响应消息格式
```json
{
  "success": true/false,
  "code": 0,
  "message": "Success",
  "data": {},
  "requestId": "对应的请求ID",
  "timestamp": 1640995200
}
```

## 错误码定义
- `0`: 成功
- `1001`: 无效请求
- `1002`: 用户不存在
- `1003`: 密码错误
- `1004`: 用户已存在
- `1005`: 数据库错误
- `1006`: 未授权
- `1007`: 参数错误
- `1008`: 服务器内部错误

---

## API 接口详情

### 1. 认证模块 (type: "auth")

#### 1.1 用户注册
- **Action**: `register`
- **说明**: 注册新用户账号
- **认证要求**: 无需登录

**请求示例**:
```json
{
  "type": "auth",
  "action": "register",
  "data": {
    "username": "mengge",
    "password": "Aaa123456!"
  },
  "requestId": "register-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "userid": 1,
    "username": "mengge"
  },
  "requestId": "register-request-id",
  "timestamp": 1640995200
}
```

**注意事项**:
- 用户名：3-20个字符，只能包含字母、数字和下划线
- 密码：至少8位，必须包含大写字母、小写字母、数字和特殊字符

#### 1.2 用户登录
- **Action**: `login`
- **说明**: 用户登录认证，登录成功后会设置在线状态
- **认证要求**: 无需登录

**请求示例**:
```json
{
  "type": "auth",
  "action": "login",
  "data": {
    "username": "mengge",
    "password": "Aaa123456!"
  },
  "requestId": "login-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "userid": 1,
    "username": "mengge"
  },
  "requestId": "login-request-id",
  "timestamp": 1640995200
}
```

**注意事项**:
- 登录成功后，WebSocket连接会绑定用户身份
- 同一用户重复登录会踢掉之前的连接
- 登录后用户状态自动设为在线

#### 1.3 用户登出
- **Action**: `logout`
- **说明**: 用户登出，清除登录状态并设置离线
- **认证要求**: 需要登录

**请求示例**:
```json
{
  "type": "auth",
  "action": "logout",
  "data": {},
  "requestId": "logout-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": "Logged out successfully",
  "requestId": "logout-request-id",
  "timestamp": 1640995200
}
```

---

### 2. 装备模块 (type: "equip")

#### 2.1 获取装备列表
- **Action**: `getEquip`
- **说明**: 获取当前用户的所有装备
- **认证要求**: 需要登录

**请求示例**:
```json
{
  "type": "equip",
  "action": "getEquip",
  "data": {},
  "requestId": "get-equip-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": [
    {
      "equipid": 41000001,
      "type": 4,
      "quality": 1,
      "damage": 150,
      "crit": 15,
      "critdamage": 200,
      "damagespeed": 120,
      "bloodsuck": 8,
      "hp": 300,
      "movespeed": 110,
      "equipname": "传说之剑",
      "userid": 12,
      "defense": 80,
      "goodfortune": 20
    }
  ],
  "requestId": "get-equip-request-id",
  "timestamp": 1640995200
}
```

#### 2.2 保存装备
- **Action**: `saveEquip`
- **说明**: 新增或更新装备信息
- **认证要求**: 需要登录

**装备ID生成规则**: `[type][quality][6位序号]`
- 例如：type=4, quality=1的第一件装备ID为 `41000001`
- equipid为0时表示新增，系统自动生成ID
- equipid非0时表示更新指定装备

**新增装备请求示例**:
```json
{
  "type": "equip",
  "action": "saveEquip",
  "data": {
    "equipment": {
      "equipid": 0,
      "type": 4,
      "quality": 1,
      "damage": 150,
      "crit": 15,
      "critdamage": 200,
      "damagespeed": 120,
      "bloodsuck": 8,
      "hp": 300,
      "movespeed": 110,
      "equipname": "传说之剑",
      "defense": 80,
      "goodfortune": 20
    }
  },
  "requestId": "save-equip-request-id",
  "timestamp": 1640995200
}
```

**更新装备请求示例**:
```json
{
  "type": "equip",
  "action": "saveEquip",
  "data": {
    "equipment": {
      "equipid": 41000001,
      "type": 4,
      "quality": 1,
      "damage": 180,
      "crit": 20,
      "critdamage": 250,
      "damagespeed": 130,
      "bloodsuck": 10,
      "hp": 350,
      "movespeed": 115,
      "equipname": "强化传说之剑",
      "defense": 90,
      "goodfortune": 25
    }
  },
  "requestId": "update-equip-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "equipid": 41000001,
    "type": 4,
    "quality": 1,
    "damage": 150,
    "crit": 15,
    "critdamage": 200,
    "damagespeed": 120,
    "bloodsuck": 8,
    "hp": 300,
    "movespeed": 110,
    "equipname": "传说之剑",
    "userid": 12,
    "defense": 80,
    "goodfortune": 20
  },
  "requestId": "save-equip-request-id",
  "timestamp": 1640995200
}
```

**必需字段**:
- `type`: 装备类型（必须大于0）
- `quality`: 装备品质（必须大于0）

#### 2.3 删除装备
- **Action**: `delEquip`
- **说明**: 删除指定装备
- **认证要求**: 需要登录

**请求示例**:
```json
{
  "type": "equip",
  "action": "delEquip",
  "data": {
    "equipid": 41000001
  },
  "requestId": "delete-equip-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "equipid": 41000001,
    "deleted": true
  },
  "requestId": "delete-equip-request-id",
  "timestamp": 1640995200
}
```

#### 2.4 批量删除装备
- **Action**: `batchDelEquip`
- **说明**: 删除指定品质的所有装备
- **认证要求**: 需要登录

**请求示例**:
```json
{
  "type": "equip",
  "action": "batchDelEquip",
  "data": {
    "quality": 1
  },
  "requestId": "batch-delete-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "quality": 1,
    "deleted_count": 5
  },
  "requestId": "batch-delete-request-id",
  "timestamp": 1640995200
}
```

---

### 3. 玩家信息模块 (type: "player")

#### 3.1 获取玩家信息
- **Action**: `getPlayerInfo`
- **说明**: 获取当前玩家的详细信息
- **认证要求**: 需要登录

**请求示例**:
```json
{
  "type": "player",
  "action": "getPlayerInfo",
  "data": {},
  "requestId": "get-player-info-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "userid": 12,
    "level": 5,
    "experience": 1500,
    "gamelevel": 3,
    "bloodenergy": 100
  },
  "requestId": "get-player-info-request-id",
  "timestamp": 1640995200
}
```

**注意事项**:
- 如果玩家信息不存在，系统会自动创建默认记录
- 默认值：level=1, experience=0, gamelevel=1, bloodenergy=100

#### 3.2 更新玩家信息
- **Action**: `updatePlayerInfo`
- **说明**: 更新玩家信息
- **认证要求**: 需要登录

**请求示例**:
```json
{
  "type": "player",
  "action": "updatePlayerInfo",
  "data": {
    "level": 6,
    "experience": 2000,
    "gamelevel": 4,
    "bloodenergy": 120
  },
  "requestId": "update-player-info-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "userid": 12,
    "level": 6,
    "experience": 2000,
    "gamelevel": 4,
    "bloodenergy": 120
  },
  "requestId": "update-player-info-request-id",
  "timestamp": 1640995200
}
```

---

### 4. 心跳模块 (type: "heartbeat")

#### 4.1 心跳检测
- **Action**: `ping`
- **说明**: 检测连接是否正常
- **认证要求**: 无需登录

**请求示例**:
```json
{
  "type": "heartbeat",
  "action": "ping",
  "data": {},
  "requestId": "ping-request-id",
  "timestamp": 1640995200
}
```

**成功响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "action": "pong",
    "time": "ok"
  },
  "requestId": "ping-request-id",
  "timestamp": 1640995200
}
```

---

## 系统特性

### 1. 在线状态管理
- **自动设置**: 用户登录时自动设为在线状态
- **自动清理**: 用户登出或断开连接时自动设为离线状态
- **服务器重启**: 服务器启动时所有用户状态重置为离线

### 2. 用户身份验证
- **连接绑定**: 登录成功后用户ID绑定到WebSocket连接
- **自动识别**: 后续请求无需传递用户ID，服务器自动识别
- **数据隔离**: 每个用户只能访问自己的数据

### 3. 数据安全
- **密码加密**: 使用bcrypt算法加密存储密码
- **参数验证**: 严格验证用户名和密码格式
- **权限控制**: 基于连接的用户权限验证

---

## 使用流程

### 1. 基本流程
1. 建立WebSocket连接到 `ws://localhost:8080/ws`
2. 发送注册请求（首次使用）
3. 发送登录请求
4. 进行游戏相关操作（装备、玩家信息等）
5. 发送登出请求（可选）
6. 断开连接

### 2. JavaScript 客户端示例
```javascript
// 建立连接
const ws = new WebSocket('ws://localhost:8080/ws');

// 发送消息的通用函数
function sendMessage(type, action, data) {
    const message = {
        type: type,
        action: action,
        data: data,
        requestId: 'req_' + Date.now(),
        timestamp: Math.floor(Date.now() / 1000)
    };
    ws.send(JSON.stringify(message));
}

// 处理响应
ws.onmessage = function(event) {
    const response = JSON.parse(event.data);
    console.log('收到响应:', response);
    
    if (!response.success) {
        console.error('请求失败:', response.message);
    }
};

// 注册示例
sendMessage('auth', 'register', {
    username: 'testuser',
    password: 'Abc123456!'
});

// 登录示例
sendMessage('auth', 'login', {
    username: 'testuser',
    password: 'Abc123456!'
});

// 获取装备示例
sendMessage('equip', 'getEquip', {});

// 保存装备示例
sendMessage('equip', 'saveEquip', {
    equipment: {
        equipid: 0,
        type: 4,
        quality: 1,
        damage: 150,
        crit: 15,
        critdamage: 200,
        damagespeed: 120,
        bloodsuck: 8,
        hp: 300,
        movespeed: 110,
        equipname: "传说之剑",
        defense: 80,
        goodfortune: 20
    }
});

// 获取玩家信息示例
sendMessage('player', 'getPlayerInfo', {});

// 心跳检测示例
sendMessage('heartbeat', 'ping', {});
```

---

## 错误处理

### 常见错误及解决方案

#### 1. 认证错误
```json
{
  "success": false,
  "code": 1002,
  "message": "Invalid username or password"
}
```
**解决方案**: 检查用户名和密码是否正确

#### 2. 参数错误
```json
{
  "success": false,
  "code": 1007,
  "message": "Type and quality are required and must be positive"
}
```
**解决方案**: 检查必需参数是否提供且格式正确

#### 3. 未授权访问
```json
{
  "success": false,
  "code": 1006,
  "message": "Authentication required"
}
```
**解决方案**: 先进行登录操作

#### 4. 数据库错误
```json
{
  "success": false,
  "code": 1005,
  "message": "Database error"
}
```
**解决方案**: 检查服务器日志，可能是数据库连接问题

---

## 注意事项

1. **JSON格式**: 请确保发送的JSON格式正确，注意不要有多余的逗号
2. **时间戳**: timestamp字段为Unix时间戳（秒）
3. **请求ID**: requestId用于匹配请求和响应，建议使用唯一值
4. **连接管理**: WebSocket连接断开会自动设置用户离线状态
5. **并发登录**: 同一用户多处登录时，新连接会踢掉旧连接
6. **密码安全**: 密码必须符合强度要求（8位以上，包含大小写字母、数字、特殊字符）
7. **装备ID**: 新增装备时equipid设为0，更新时使用实际ID
8. **数据隔离**: 用户只能操作自己的数据，无法访问其他用户信息

---

## 测试建议

1. 使用项目提供的 `web/test_client.html` 进行功能测试
2. 按照认证→装备→玩家信息的顺序进行测试
3. 测试异常情况（错误密码、无效参数等）
4. 验证用户数据隔离功能
5. 测试连接断开后的状态处理

---

## 技术支持

如有问题请参考：
- 服务器日志：检查具体错误信息
- 数据库状态：确认数据库连接正常
- 网络连接：确认WebSocket连接建立成功