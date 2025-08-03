# Sourcestone API 接口文档

本文档描述了 sourcestone 表的 CRUD 操作接口。

## 接口概览

所有 sourcestone 相关的操作都通过 WebSocket 连接进行，消息格式遵循项目的统一标准。

### 消息类型
- **type**: `"sourcestone"`
- **actions**: `createSourcestone`, `getSourcestones`, `getSourcestone`, `updateSourcestone`, `deleteSourcestone`, `deleteAllSourcestones`

## 1. 创建 Sourcestone

### 请求
```json
{
  "type": "sourcestone",
  "action": "createSourcestone",
  "data": {
    "equipid": 12345,
    "sourcetype": 1,
    "count": 10,
    "quality": 3
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 响应
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "message": "Sourcestone created successfully",
    "equipid": 12345
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

## 2. 获取用户所有 Sourcestones

### 请求（无过滤条件）
```json
{
  "type": "sourcestone",
  "action": "getSourcestones",
  "data": {},
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 请求（带过滤条件）
```json
{
  "type": "sourcestone",
  "action": "getSourcestones",
  "data": {
    "sourcetype": 1,
    "quality": 3
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 响应
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "sourcestones": [
      {
        "equipid": 12345,
        "sourcetype": 1,
        "count": 10,
        "quality": 3,
        "userid": 1
      }
    ],
    "count": 1
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

## 3. 获取特定 Sourcestone

### 请求
```json
{
  "type": "sourcestone",
  "action": "getSourcestone",
  "data": {
    "equipid": 12345
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 响应
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "equipid": 12345,
    "sourcetype": 1,
    "count": 10,
    "quality": 3,
    "userid": 1
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

## 4. 更新 Sourcestone

### 请求
```json
{
  "type": "sourcestone",
  "action": "updateSourcestone",
  "data": {
    "equipid": 12345,
    "count": 15,
    "quality": 4
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 响应
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "message": "Sourcestone updated successfully",
    "equipid": 12345
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

## 5. 删除 Sourcestone

### 请求
```json
{
  "type": "sourcestone",
  "action": "deleteSourcestone",
  "data": {
    "equipid": 12345
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 响应
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "message": "Sourcestone deleted successfully",
    "equipid": 12345
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

## 6. 删除用户所有 Sourcestones

### 请求
```json
{
  "type": "sourcestone",
  "action": "deleteAllSourcestones",
  "data": {},
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 响应
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "message": "All sourcestones deleted successfully",
    "deleted_count": 5
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

## 错误响应

### 未认证用户
```json
{
  "success": false,
  "code": 1006,
  "message": "User not authenticated",
  "data": null,
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 无效请求参数
```json
{
  "success": false,
  "code": 1007,
  "message": "Invalid request format",
  "data": null,
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 数据库错误
```json
{
  "success": false,
  "code": 1005,
  "message": "Failed to create sourcestone",
  "data": null,
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### Sourcestone 未找到
```json
{
  "success": false,
  "code": 1007,
  "message": "Sourcestone not found",
  "data": null,
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

## 数据字段说明

| 字段名 | 类型 | 描述 | 必填 |
|--------|------|------|------|
| equipid | int | 装备ID，作为主键之一 | 是 |
| sourcetype | int | 源石类型 | 创建时必填 |
| count | int | 数量 | 创建时必填 |
| quality | int | 品质等级 | 创建时必填 |
| userid | int | 用户ID，由系统自动设置 | 系统自动 |

## 注意事项

1. 所有操作都需要用户先登录认证
2. `equipid` + `userid` 组合作为唯一标识
3. 更新操作支持部分字段更新，不传的字段保持原值
4. 删除操作会检查权限，只能删除属于当前用户的数据
5. 查询操作支持可选的过滤条件