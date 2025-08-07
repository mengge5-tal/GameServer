# GameServer v2.0 - Clean Architecture

一个基于WebSocket的游戏服务器，采用Clean Architecture和DDD设计模式，支持用户认证、装备管理、好友系统和排行榜等功能。

## 🏗️ 架构概览

本项目采用**Clean Architecture**（洁净架构）和**DDD**（领域驱动设计）模式，确保代码的可维护性、测试性和扩展性。

### 架构层次

```
┌─────────────────────────────────────────────────────────────┐
│                    Interface Layer                          │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │   WebSocket     │ │      HTTP       │ │   Middleware    ││
│  │   Handlers      │ │   Endpoints     │ │   & Security    ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Application Layer                          │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │   Use Cases     │ │   App Services  │ │      DTOs       ││
│  │   & Workflows   │ │   & Orchestr.   │ │   & Mappers     ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │    Entities     │ │  Domain Services│ │  Value Objects  ││
│  │  & Aggregates   │ │  & Business     │ │  & Repository   ││
│  │                 │ │     Rules       │ │   Interfaces    ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │   Repository    │ │     Database    │ │     Cache &     ││
│  │ Implementations │ │   Connections   │ │   External      ││
│  │                 │ │   & Migrations  │ │   Services      ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## 📁 项目结构

```
GameServer/
├── cmd/                           # 应用程序入口
│   └── server/
│       ├── main.go               # 旧版本主程序
│       └── main_new.go           # 新架构主程序
├── internal/                     # 私有应用代码
│   ├── domain/                   # 🔷 Domain Layer (核心业务逻辑)
│   │   ├── entity/               # 业务实体
│   │   │   └── user.go          # 用户、玩家、装备等实体
│   │   ├── repository/           # 仓储接口
│   │   │   └── user_repository.go
│   │   ├── service/              # 领域服务
│   │   │   └── auth_service.go
│   │   └── valueobject/          # 值对象
│   │       └── message.go        # 消息类型定义
│   ├── application/              # 🔶 Application Layer (应用逻辑)
│   │   ├── dto/                  # 数据传输对象
│   │   │   ├── auth_dto.go
│   │   │   ├── player_dto.go
│   │   │   ├── friend_dto.go
│   │   │   └── ranking_dto.go
│   │   └── service/              # 应用服务
│   │       ├── auth_service.go
│   │       ├── player_service.go
│   │       ├── friend_service.go
│   │       └── ranking_service.go
│   ├── infrastructure/           # 🔸 Infrastructure Layer (基础设施)
│   │   ├── cache/                # 缓存实现
│   │   │   └── cache_service.go
│   │   ├── config/               # 配置管理
│   │   │   └── config.go
│   │   ├── container/            # 依赖注入容器
│   │   │   └── container.go
│   │   ├── database/             # 数据库连接
│   │   │   └── connection.go
│   │   └── repository/           # 仓储实现
│   │       ├── mysql_user_repository.go
│   │       ├── mysql_player_repository.go
│   │       ├── mysql_equipment_repository.go
│   │       ├── mysql_friend_repository.go
│   │       ├── mysql_ranking_repository.go
│   │       ├── mysql_sourcestone_repository.go
│   │       └── mysql_experience_repository.go
│   └── interfaces/               # 🔹 Interface Layer (接口层)
│       └── websocket/            # WebSocket接口
│           ├── client.go         # 客户端连接管理
│           ├── hub.go            # 连接中心
│           ├── router.go         # 消息路由
│           ├── handlers.go       # 消息处理器
│           └── service_interfaces.go
├── pkg/                          # 公共库代码
│   ├── logger/                   # 日志系统
│   └── metrics/                  # 监控指标
└── docs/                         # 项目文档
```

## 🚀 快速开始

### 环境要求
- Go 1.24+
- MySQL 5.7+

### 1. 配置环境变量

复制环境变量模板并配置：
```bash
cp .env.example .env
```

编辑 `.env` 文件，设置数据库连接信息：
```env
DB_HOST=your_database_host
DB_PORT=3306
DB_NAME=your_database_name
DB_USER=your_database_user
DB_PASSWORD=your_database_password
```

### 2. 安装依赖
```bash
go mod download
```

### 3. 初始化数据库
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE your_database_name;"

# 运行初始化脚本
mysql -u your_user -p your_database_name < internal/database/init_tables.sql
```

### 4. 启动服务器

**使用新架构启动：**
```bash
go run cmd/server/main_new.go
```

**或编译后运行：**
```bash
go build -o gameserver-v2 cmd/server/main_new.go
./gameserver-v2
```

服务器将在 `101.201.51.135:8080` 启动

## 🌐 API端点

| 端点 | 类型 | 描述 |
|------|------|------|
| `/ws` | WebSocket | 主要游戏通信接口 |
| `/health` | GET | 健康检查 |
| `/metrics` | GET | 系统监控指标 |
| `/info` | GET | 服务信息和架构详情 |

## 📱 功能特性

### 🔐 认证系统
- ✅ 用户注册/登录/登出
- ✅ 密码强度验证 (大小写字母+数字+特殊字符)
- ✅ 用户名格式验证 (3-20字符，字母数字下划线)
- ✅ bcrypt密码加密
- ✅ 会话管理和缓存

### 🎮 玩家系统
- ✅ 玩家信息管理 (等级、经验、游戏等级、血气)
- ✅ 装备系统 (获取、保存、删除装备)
- ✅ 源石管理
- ✅ 数据缓存优化

### 👥 社交系统
- ✅ 好友系统 (添加、删除、好友申请)
- ✅ 好友请求管理 (接受、拒绝申请)
- ✅ 好友排行榜
- ✅ 实时状态更新

### 🏆 排行系统
- ✅ 多类型排行榜 (等级、经验值、装备战力)
- ✅ 实时排名更新
- ✅ 个人排名查询
- ✅ 排行榜缓存

### 🔧 技术特性
- ✅ **Clean Architecture** - 分层清晰，职责分离
- ✅ **依赖注入** - 减少耦合，提高可测试性
- ✅ **Repository模式** - 数据访问抽象
- ✅ **多级缓存** - 内存缓存提升性能
- ✅ **配置管理** - 环境变量配置，无硬编码
- ✅ **实时通信** - WebSocket双向通信
- ✅ **错误处理** - 统一错误处理和响应
- ✅ **日志系统** - 结构化日志记录
- ✅ **监控指标** - 内置性能监控

## 📋 消息协议

### 请求消息格式
```json
{
  "type": "auth|player|equip|friend|rank|heartbeat",
  "action": "specific_action",
  "data": {
    "key": "value"
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 响应消息格式
```json
{
  "success": true|false,
  "code": 0,
  "message": "Success|Error message",
  "data": {
    "response_data": "value"
  },
  "requestId": "unique-request-id",
  "timestamp": 1640995200
}
```

### 支持的消息类型

| 类型 | 操作 | 描述 | 认证需求 |
|------|------|------|----------|
| `auth` | `login` | 用户登录 | ❌ |
| `auth` | `register` | 用户注册 | ❌ |
| `auth` | `logout` | 用户登出 | ✅ |
| `heartbeat` | `ping` | 心跳检测 | ❌ |
| `player` | `getPlayerInfo` | 获取玩家信息 | ✅ |
| `player` | `updatePlayer` | 更新玩家信息 | ✅ |
| `equip` | `getEquip` | 获取装备列表 | ✅ |
| `equip` | `saveEquip` | 保存装备 | ✅ |
| `equip` | `deleteEquip` | 删除装备 | ✅ |
| `friend` | `getFriends` | 获取好友列表 | ✅ |
| `friend` | `addFriend` | 发送好友申请 | ✅ |
| `friend` | `acceptFriend` | 接受好友申请 | ✅ |
| `friend` | `rejectFriend` | 拒绝好友申请 | ✅ |
| `friend` | `removeFriend` | 删除好友 | ✅ |
| `friend` | `getFriendRank` | 获取好友排行 | ✅ |
| `rank` | `getAllRank` | 获取排行榜 | ✅ |
| `rank` | `getRank` | 获取个人排名 | ✅ |

## 🏗️ 架构优势

### 与旧版本对比

| 特性 | 旧版本 | 新版本 v2.0 |
|------|--------|-------------|
| **架构模式** | 分层混乱 | Clean Architecture + DDD |
| **业务逻辑** | 分散在各处 | 集中在Domain和Application层 |
| **数据访问** | 直接SQL操作 | Repository模式抽象 |
| **依赖管理** | 硬编码依赖 | 依赖注入容器 |
| **配置管理** | 硬编码配置 | 环境变量配置 |
| **代码重复** | 多处重复逻辑 | DRY原则，统一实现 |
| **测试性** | 难以测试 | 高度可测试 |
| **扩展性** | 难以扩展 | 易于扩展和维护 |

### 核心优势

1. **分离关注点** - 每层都有明确的职责
2. **依赖反转** - 高层模块不依赖低层模块
3. **可测试性** - 通过接口抽象，易于单元测试
4. **可维护性** - 模块化设计，便于维护和修改
5. **可扩展性** - 新功能易于添加，不影响现有代码
6. **性能优化** - 多级缓存和数据库连接池优化

## 🛠️ 开发指南

### 添加新功能

1. **定义实体** - 在 `domain/entity/` 中定义业务实体
2. **创建仓储接口** - 在 `domain/repository/` 中定义数据访问接口
3. **实现仓储** - 在 `infrastructure/repository/` 中实现具体的数据访问
4. **创建应用服务** - 在 `application/service/` 中实现业务逻辑
5. **定义DTO** - 在 `application/dto/` 中定义数据传输对象
6. **添加接口处理器** - 在 `interfaces/websocket/` 中添加消息处理器
7. **注册依赖** - 在 `infrastructure/container/` 中注册新的依赖

### 测试策略

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/application/service/...

# 运行带覆盖率的测试
go test -cover ./...
```

## 📊 监控和运维

### 健康检查
```bash
curl http://101.201.51.135:8080/health
```

### 系统指标
```bash
curl http://localhost:8080/metrics
```

### 服务信息
```bash
curl http://localhost:8080/info
```

## 🔧 配置选项

详细配置请参考 `.env.example` 文件：

- **数据库配置** - 连接信息和连接池设置
- **服务器配置** - 主机和端口设置
- **WebSocket配置** - 缓冲区大小和超时设置
- **安全配置** - 密码加密强度
- **日志配置** - 日志级别和格式
- **缓存配置** - TTL和清理间隔
- **限流配置** - 请求频率限制

## 📚 相关文档

- [API文档](docs/API_DOCUMENTATION.md)
- [部署指南](docs/DEPLOYMENT.md)
- [性能优化](docs/PERFORMANCE_OPTIMIZATION.md)
- [安全改进](docs/SECURITY_IMPROVEMENTS.md)

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📄 许可证

[MIT License](LICENSE)

---

**GameServer v2.0** - 企业级Clean Architecture游戏服务器