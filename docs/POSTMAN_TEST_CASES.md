# Postman WebSocket 测试用例

## 连接设置
**WebSocket URL**: `ws://localhost:8080/ws`

## 测试用例

### 1. 用户注册
```json
{
  "type": "auth",
  "action": "register",
  "data": {
    "username": "mengge",
    "password": "123456"
  },
  "requestId": "req_register_001",
  "timestamp": 1640995200
}
```
**预期响应**:
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "userid": 1,
    "username": "mengge"
  },
  "requestId": "req_register_001",
  "timestamp": 1753625500
}
```

### 2. 用户登录
```json
{
  "type": "auth",
  "action": "login",
  "data": {
    "username": "mengge",
    "password": "123456"
  },
  "requestId": "req_login_001",
  "timestamp": 1640995200
}
```

### 3. 获取玩家信息
```json
{
  "type": "player",
  "action": "getPlayerInfo",
  "data": {},
  "requestId": "req_player_001",
  "timestamp": 1640995200
}
```

### 4. 更新玩家信息
```json
{
  "type": "player",
  "action": "updatePlayerInfo",
  "data": {
    "level": 5,
    "experience": 1500,
    "gamelevel": 3,
    "blood_energy": 120
  },
  "requestId": "req_update_player_001",
  "timestamp": 1640995200
}
```

### 5. 保存装备（新增）- 重点测试
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
  "requestId": "req_save_equip_001",
  "timestamp": 1640995200
}
```
**预期响应**: equipid 应该是 41000001

### 6. 保存第二件相同类型品质的装备
```json
{
  "type": "equip",
  "action": "saveEquip",
  "data": {
    "equipment": {
      "equipid": 0,
      "type": 4,
      "quality": 1,
      "damage": 180,
      "crit": 20,
      "critdamage": 250,
      "damagespeed": 130,
      "bloodsuck": 10,
      "hp": 350,
      "movespeed": 115,
      "equipname": "传说之剑+1",
      "defense": 90,
      "goodfortune": 25
    }
  },
  "requestId": "req_save_equip_002",
  "timestamp": 1640995200
}
```
**预期响应**: equipid 应该是 41000002

### 7. 保存不同类型的装备
```json
{
  "type": "equip",
  "action": "saveEquip",
  "data": {
    "equipment": {
      "equipid": 0,
      "type": 2,
      "quality": 3,
      "damage": 200,
      "crit": 25,
      "critdamage": 300,
      "damagespeed": 100,
      "bloodsuck": 15,
      "hp": 400,
      "movespeed": 105,
      "equipname": "神话护甲",
      "defense": 150,
      "goodfortune": 30
    }
  },
  "requestId": "req_save_equip_003",
  "timestamp": 1640995200
}
```
**预期响应**: equipid 应该是 23000001

### 8. 获取装备列表
```json
{
  "type": "equip",
  "action": "getEquip",
  "data": {},
  "requestId": "req_get_equip_001",
  "timestamp": 1640995200
}
```

### 9. 更新装备
```json
{
  "type": "equip",
  "action": "saveEquip",
  "data": {
    "equipment": {
      "equipid": 41000001,
      "type": 4,
      "quality": 1,
      "damage": 200,
      "crit": 25,
      "critdamage": 300,
      "damagespeed": 140,
      "bloodsuck": 12,
      "hp": 400,
      "movespeed": 120,
      "equipname": "传说之剑·强化",
      "defense": 100,
      "goodfortune": 35
    }
  },
  "requestId": "req_update_equip_001",
  "timestamp": 1640995200
}
```

### 10. 删除装备
```json
{
  "type": "equip",
  "action": "delEquip",
  "data": {
    "equipid": 41000002
  },
  "requestId": "req_del_equip_001",
  "timestamp": 1640995200
}
```

### 11. 批量删除装备
```json
{
  "type": "equip",
  "action": "batchDelEquip",
  "data": {
    "quality": 1
  },
  "requestId": "req_batch_del_001",
  "timestamp": 1640995200
}
```

### 12. 心跳测试
```json
{
  "type": "heartbeat",
  "action": "ping",
  "data": {},
  "requestId": "req_ping_001",
  "timestamp": 1640995200
}
```

### 13. 获取排行榜
```json
{
  "type": "rank",
  "action": "getAllRank",
  "data": {
    "rank_type": "level",
    "limit": 10
  },
  "requestId": "req_rank_001",
  "timestamp": 1640995200
}
```

### 14. 获取个人排名
```json
{
  "type": "rank",
  "action": "getSelfRank",
  "data": {
    "rank_type": "level"
  },
  "requestId": "req_self_rank_001",
  "timestamp": 1640995200
}
```

## 测试流程建议

### 顺序测试：
1. **连接WebSocket** - 建立连接
2. **注册用户** - 用例1
3. **登录用户** - 用例2
4. **获取玩家信息** - 用例3
5. **保存装备测试** - 用例5、6、7（重点测试equipid生成）
6. **获取装备列表** - 用例8（验证装备是否正确保存）
7. **更新装备** - 用例9
8. **删除装备** - 用例10
9. **其他功能测试** - 用例11-14

### 装备ID测试验证：
- 第一件 type=4, quality=1 的装备 → equipid = 41000001
- 第二件 type=4, quality=1 的装备 → equipid = 41000002  
- 第一件 type=2, quality=3 的装备 → equipid = 23000001

### 错误测试：
- 未登录状态下访问需要认证的接口
- 保存装备时缺少type或quality字段
- 删除不存在的装备
- 无效的用户名密码登录

## 注意事项
1. 所有需要认证的操作都必须先登录
2. `timestamp` 可以使用固定值或当前时间戳
3. `requestId` 建议每个请求使用不同的值
4. 重点关注装备ID的生成规律是否符合预期