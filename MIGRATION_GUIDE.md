# 架构迁移指南

本文档指导如何从旧架构迁移到新的Clean Architecture。

## 🎯 迁移概述

本次重构将项目从混乱的分层架构迁移到标准的Clean Architecture + DDD模式，主要改进：

- ✅ **消除业务逻辑重复** - 统一认证、装备、好友等业务逻辑
- ✅ **实施依赖注入** - 移除全局变量和硬编码依赖
- ✅ **统一配置管理** - 移除硬编码配置，使用环境变量
- ✅ **Repository模式** - 抽象数据访问，提高可测试性
- ✅ **分层架构** - 清晰的职责分离

## 🔄 迁移步骤

### 1. 环境准备

**备份现有数据库：**
```bash
mysqldump -u username -p database_name > backup.sql
```

**设置环境变量：**
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件，移除硬编码的数据库信息
vim .env
```

### 2. 启动新架构

**停止旧服务：**
```bash
# 如果旧服务在运行，先停止
pkill -f gameserver
```

**启动新架构服务：**
```bash
# 编译新版本
go build -o gameserver-v2 cmd/server/main_new.go

# 启动新服务
./gameserver-v2
```

### 3. 验证功能

使用WebSocket客户端测试所有功能：

**认证测试：**
```json
// 注册用户
{
  "type": "auth",
  "action": "register", 
  "data": {"username": "testuser", "password": "Test123!@#"},
  "requestId": "req-1",
  "timestamp": 1640995200
}

// 用户登录
{
  "type": "auth",
  "action": "login",
  "data": {"username": "testuser", "password": "Test123!@#"},
  "requestId": "req-2", 
  "timestamp": 1640995200
}
```

**玩家功能测试：**
```json
// 获取玩家信息
{
  "type": "player",
  "action": "getPlayerInfo",
  "data": {},
  "requestId": "req-3",
  "timestamp": 1640995200
}

// 获取装备
{
  "type": "equip", 
  "action": "getEquip",
  "data": {},
  "requestId": "req-4",
  "timestamp": 1640995200
}
```

## 📊 架构对比

### 旧架构问题

```
# 旧架构的问题代码示例

// 1. 业务逻辑重复
internal/server/handlers_temp.go:handleLogin()
internal/server/handlers_auth_optimized.go:handleLoginOptimized()

// 2. 硬编码配置
config.go:62 - Host: "rm-2zevr95ez9rrid70uho.mysql.rds.aliyuncs.com"

// 3. 直接数据库操作
handlers_temp.go:45 - db.QueryRow("SELECT userid, username...")

// 4. 全局变量滥用
server/cache.go - var globalCache = make(map[string]interface{})
```

### 新架构优势

```
# 新架构的改进

// 1. 统一业务逻辑
internal/application/service/auth_service.go:Login()

// 2. 环境变量配置
internal/infrastructure/config/config.go:getEnvRequired("DB_HOST")

// 3. Repository模式
internal/infrastructure/repository/mysql_user_repository.go:GetByUsername()

// 4. 依赖注入
internal/infrastructure/container/container.go:NewContainer()
```

## 🗂️ 文件映射关系

| 旧文件路径 | 新文件路径 | 说明 |
|-----------|-----------|------|
| `internal/server/handlers_temp.go` | `internal/application/service/auth_service.go` | 认证逻辑重构 |
| `internal/server/handlers_equip.go` | `internal/application/service/player_service.go` | 装备逻辑重构 |
| `internal/models/*.go` | `internal/domain/entity/user.go` | 实体统一定义 |
| `internal/server/websocket.go` | `internal/interfaces/websocket/*.go` | WebSocket重构 |
| `internal/config/config.go` | `internal/infrastructure/config/config.go` | 配置管理改进 |

## 🔧 配置迁移

### 数据库配置

**旧配置（硬编码）：**
```go
Database: DatabaseConfig{
    Host: "rm-2zevr95ez9rrid70uho.mysql.rds.aliyuncs.com",
    User: "wwk18255113901", 
    Password: "BaiChen123456+",
}
```

**新配置（环境变量）：**
```bash
# .env 文件
DB_HOST=your_database_host
DB_USER=your_database_user  
DB_PASSWORD=your_database_password
```

### 缓存配置

**旧配置（分散代码）：**
```go
// 分散在各个文件中
cache.SetTTL(15 * time.Minute)
```

**新配置（统一管理）：**
```bash
# .env 文件
CACHE_DEFAULT_TTL=15m
CACHE_CLEANUP_INTERVAL=5m
```

## 🧪 功能验证清单

### ✅ 认证系统
- [ ] 用户注册功能正常
- [ ] 用户登录功能正常  
- [ ] 密码验证规则生效
- [ ] 用户名验证规则生效
- [ ] 登出功能正常

### ✅ 玩家系统
- [ ] 获取玩家信息正常
- [ ] 更新玩家信息正常
- [ ] 获取装备列表正常
- [ ] 保存装备功能正常
- [ ] 删除装备功能正常

### ✅ 社交系统
- [ ] 获取好友列表正常
- [ ] 发送好友申请正常
- [ ] 接受好友申请正常
- [ ] 拒绝好友申请正常
- [ ] 删除好友功能正常
- [ ] 好友排行榜正常

### ✅ 排行系统
- [ ] 获取排行榜正常
- [ ] 获取个人排名正常
- [ ] 排名更新功能正常

### ✅ 系统功能
- [ ] 心跳检测正常
- [ ] 健康检查正常
- [ ] 监控指标正常
- [ ] 缓存功能正常

## 🚨 注意事项

### 数据库兼容性
- ✅ 数据库表结构完全兼容
- ✅ 无需数据迁移
- ✅ 支持平滑切换

### API兼容性  
- ✅ WebSocket消息格式完全兼容
- ✅ 所有现有功能保持不变
- ✅ 客户端无需修改

### 性能影响
- ✅ 新增缓存层，性能提升
- ✅ 连接池优化，并发能力增强  
- ✅ 依赖注入可能带来轻微性能开销（可忽略）

## 🔙 回滚计划

如果迁移出现问题，可以快速回滚：

**停止新服务：**
```bash
pkill -f gameserver-v2
```

**启动旧服务：**
```bash
go build -o gameserver-old cmd/server/main.go
./gameserver-old
```

**恢复数据库（如有必要）：**
```bash
mysql -u username -p database_name < backup.sql
```

## 📈 后续优化

迁移完成后，可以进行以下优化：

1. **添加单元测试** - 利用新架构的可测试性
2. **监控集成** - 集成Prometheus、Grafana等监控工具
3. **性能调优** - 基于监控数据进行性能优化
4. **文档完善** - 完善API文档和开发指南
5. **CI/CD流程** - 建立自动化测试和部署流程

## 💡 故障排查

### 常见问题

**1. 数据库连接失败**
```
Error: database host is required (set DB_HOST environment variable)
```
解决：检查 `.env` 文件中的数据库配置

**2. 端口冲突**
```
Error: bind: address already in use
```
解决：停止旧服务或修改端口配置

**3. 依赖注入失败**
```
Error: Failed to initialize container
```
解决：检查所有依赖配置是否正确

### 日志查看
```bash
# 查看详细启动日志
./gameserver-v2 2>&1 | tee startup.log

# 查看错误日志  
grep -i error startup.log
```

## 📞 技术支持

如果在迁移过程中遇到问题：

1. 检查本迁移指南
2. 查看系统日志
3. 验证环境变量配置
4. 确认数据库连接
5. 对比新旧架构差异

---

**迁移完成后，您将拥有一个企业级的Clean Architecture游戏服务器！** 🎉