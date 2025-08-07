# GameServer v2.0 - Clean Architecture

一个基于WebSocket的游戏服务器，采用Clean Architecture和DDD设计模式，支持用户认证、装备管理、好友系统和排行榜等功能。

> **🚀 Version 2.0 现已发布！** 项目已完全重构为企业级Clean Architecture，具备更高的可维护性、可测试性和扩展性。

## ⚡ 快速开始

### 1. 环境配置
```bash
# 复制环境变量模板
cp .env.example .env
# 编辑配置文件，设置数据库连接信息
vim .env
```

### 2. 启动服务器
```bash
# 编译并启动
go build -o gameserver ./cmd/server
./gameserver
```

服务器将在 `101.201.51.135:8080` 启动

## 📚 完整文档
- **[完整架构文档](README_v2.md)** - 详细的架构设计和使用指南
- **[迁移指南](MIGRATION_GUIDE.md)** - 从旧版本迁移的详细步骤

## 项目结构

```
GameServer/
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器主程序
│       └── main.go        # 程序入口点
├── internal/              # 私有应用代码
│   ├── config/           # 配置管理
│   ├── database/         # 数据库连接和SQL文件
│   ├── handlers/         # 业务逻辑处理器
│   │   ├── auth/        # 认证相关
│   │   ├── player/      # 玩家管理
│   │   ├── equipment/   # 装备系统
│   │   ├── friend/      # 好友系统
│   │   ├── rank/        # 排行榜
│   │   └── heartbeat/   # 心跳检测
│   ├── models/          # 数据模型
│   └── server/          # 核心服务器代码
│       ├── router.go    # 路由系统
│       ├── middleware.go # 中间件
│       ├── websocket.go  # WebSocket处理
│       └── message.go    # 消息定义
├── pkg/                  # 可重用的库代码
│   ├── logger/          # 日志系统
│   └── metrics/         # 监控指标
├── configs/             # 配置文件
├── docs/               # 项目文档
├── scripts/            # 构建和部署脚本
└── web/               # 静态资源
```

## 功能特性

### 🔐 认证系统
- 用户注册/登录/登出
- 密码强度验证
- 用户名格式验证
- 基于bcrypt的密码加密

### 🎮 游戏功能
- 玩家信息管理
- 装备系统（获取、保存、删除装备）
- 好友系统（添加、删除、好友申请）
- 排行榜（等级、经验值排行）

### 🔧 技术架构
- **路由系统**: 标准的消息路由分发
- **中间件支持**: 认证、日志、限流、验证中间件
- **WebSocket**: 实时双向通信
- **数据库**: MySQL支持
- **日志系统**: 结构化日志记录
- **监控**: 内置指标收集

## 快速开始

### 环境要求
- Go 1.19+
- MySQL 5.7+

### 安装依赖
```bash
go mod download
```

### 配置数据库
1. 创建MySQL数据库
2. 运行 `internal/database/init_tables.sql` 初始化表结构
3. 设置环境变量或修改配置文件

### 启动服务器
```bash
go build -o gameserver ./cmd/server
./gameserver
```

服务器默认运行在 `101.201.51.135:8080`

### API端点
- **WebSocket**: `ws://101.201.51.135:8080/ws`
- **健康检查**: `GET /health`
- **监控指标**: `GET /metrics`
- **路由信息**: `GET /routes`

## 消息格式

### 请求消息
```json
{
  "type": "auth",
  "action": "login",
  "data": {
    "username": "user123",
    "password": "password123"
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 响应消息
```json
{
  "success": true,
  "code": 0,
  "message": "Success",
  "data": {
    "userid": 1,
    "username": "user123"
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

## 支持的消息类型

| 类型 | 操作 | 描述 |
|------|------|------|
| `auth` | `login` | 用户登录 |
| `auth` | `register` | 用户注册 |
| `auth` | `logout` | 用户登出 |
| `heartbeat` | `ping` | 心跳检测 |
| `equip` | `getEquip` | 获取装备 |
| `equip` | `saveEquip` | 保存装备 |
| `player` | `getPlayerInfo` | 获取玩家信息 |
| `friend` | `getFriends` | 获取好友列表 |
| `rank` | `getAllRank` | 获取排行榜 |

## 开发说明

### 添加新的处理器
1. 在 `internal/handlers/` 下创建新的包
2. 实现处理逻辑
3. 在 `internal/server/handlers_adapter.go` 中注册路由

### 中间件
项目包含以下中间件：
- **AuthMiddleware**: 认证验证
- **LoggingMiddleware**: 请求日志记录
- **RateLimitMiddleware**: 请求限流
- **ValidationMiddleware**: 消息格式验证

### 数据库迁移
数据库脚本位于 `internal/database/` 目录下。

### 数据库表结构
## user
userid,int,NO,PRI,,auto_increment
username,varchar(45),NO,"",,""
passward,varchar(45),NO,"",,""
## sourcestone
equipid,int,NO,PRI,,""
sourcetype,int,YES,"",,""
count,int,YES,"",,""
quality,int,YES,"",,""
userid,int,YES,"",,""
## ranking
id,int,NO,PRI,,auto_increment
userid,int,NO,MUL,,""
rank_type,"enum('level','experience','equipment_power')",YES,"",level,""
rank_value,int,NO,"",0,""
rank_position,int,NO,"",0,""
updated_at,timestamp,NO,"",CURRENT_TIMESTAMP,DEFAULT_GENERATED on update CURRENT_TIMESTAMP
## playerinfo
userid,int,NO,PRI,,""
level,int,YES,"",,""
experience,int,YES,"",,""
gamelevel,int,YES,"",,""
bloodenergy,int,YES,"",,""
## friend_request
id,int,NO,PRI,,auto_increment
fromuserid,int,NO,MUL,,""
touserid,int,NO,"",,""
message,varchar(255),YES,"","",""
status,"enum('pending','accepted','rejected')",YES,"",pending,""
created_at,timestamp,NO,"",CURRENT_TIMESTAMP,DEFAULT_GENERATED
updated_at,timestamp,NO,"",CURRENT_TIMESTAMP,DEFAULT_GENERATED on update CURRENT_TIMESTAMP
## friend
id,int,NO,PRI,,auto_increment
fromuserid,int,NO,MUL,,""
touserid,int,NO,"",,""
status,"enum('pending','accepted','blocked')",YES,"",pending,""
created_at,timestamp,NO,"",CURRENT_TIMESTAMP,DEFAULT_GENERATED
updated_at,timestamp,NO,"",CURRENT_TIMESTAMP,DEFAULT_GENERATED on update CURRENT_TIMESTAMP
## experience
level,int,NO,PRI,,""
value,int,NO,"",,""
## equip
equipid,int,NO,PRI,,""
quality,int,NO,"",,""
damage,int,YES,"",,""
crit,int,YES,"",,""
critdamage,int,YES,"",,""
damagespeed,int,YES,"",,""
bloodsuck,int,YES,"",,""
hp,int,YES,"",,""
movespeed,int,YES,"",,""
equipname,varchar(45),YES,"",,""
userid,int,NO,"",,""
defense,int,YES,"",,""
goodfortune,int,YES,"",,""
type,int,YES,"",1,""



## 许可证

[MIT License](LICENSE)