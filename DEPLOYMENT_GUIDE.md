# GameServer 阿里云部署指南

## 🚀 快速解决方案（推荐先用这个）

如果你需要立即解决SSH断开导致进程结束的问题，使用这个方案：

### 1. 上传文件到服务器
```bash
# 在本地项目目录执行
scp quick_deploy.sh your-username@your-server-ip:/home/your-username/
scp server_manager.sh your-username@your-server-ip:/home/your-username/
```

### 2. 在服务器上执行
```bash
# SSH连接到服务器
ssh your-username@your-server-ip

# 进入项目目录
cd /path/to/your/GameServer

# 给脚本执行权限
chmod +x quick_deploy.sh server_manager.sh

# 编辑数据库配置
nano quick_deploy.sh
# 修改以下行为你的实际数据库信息：
# export DB_USER=your_actual_db_user
# export DB_PASSWORD=your_actual_db_password

# 快速部署
./quick_deploy.sh
```

### 3. 验证部署
```bash
# 查看screen会话
screen -ls

# 连接到服务器查看日志
screen -r gameserver

# 分离会话（按键组合）
# Ctrl+A 然后按 D

# 现在你可以安全地关闭SSH连接！
```

## 🏗️ 生产环境部署方案

### 方案A: systemd 服务（推荐生产环境）

#### 1. 准备工作
```bash
# 在服务器上创建项目目录
sudo mkdir -p /opt/gameserver
cd /opt/gameserver

# 上传项目文件
# 方法1: 使用scp
scp -r /path/to/local/GameServer/* your-username@your-server:/opt/gameserver/

# 方法2: 使用git clone（如果代码在git仓库）
git clone https://github.com/your-username/GameServer.git .
```

#### 2. 执行部署脚本
```bash
# 给脚本执行权限
chmod +x deploy_gameserver.sh

# 执行部署（会提示输入数据库信息）
./deploy_gameserver.sh
```

#### 3. 服务管理
```bash
# 查看服务状态
sudo systemctl status gameserver

# 启动服务
sudo systemctl start gameserver

# 停止服务
sudo systemctl stop gameserver

# 重启服务
sudo systemctl restart gameserver

# 查看日志
sudo journalctl -u gameserver -f
```

### 方案B: Docker 部署

#### 1. 创建Dockerfile（已生成）
项目根目录下的 `Dockerfile` 已经准备好。

#### 2. 构建和运行
```bash
# 构建Docker镜像
docker build -t gameserver .

# 创建数据库配置文件
cat > .env << EOF
DB_HOST=your-db-host
DB_PORT=3306
DB_NAME=gameserver
DB_USER=your-db-user
DB_PASSWORD=your-db-password
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
EOF

# 运行容器
docker run -d \
  --name gameserver \
  --restart unless-stopped \
  -p 8080:8080 \
  --env-file .env \
  gameserver

# 查看日志
docker logs -f gameserver
```

## 📋 部署前检查清单

### 系统要求
- [ ] 操作系统：Ubuntu 18.04+ 或 CentOS 7+
- [ ] Go 版本：1.19+
- [ ] 内存：至少 512MB
- [ ] 磁盘：至少 1GB 可用空间
- [ ] 网络：开放 8080 端口（或你自定义的端口）

### 数据库准备
- [ ] MySQL 5.7+ 或 8.0+
- [ ] 创建数据库和用户
- [ ] 运行初始化SQL脚本

```sql
-- 创建数据库
CREATE DATABASE gameserver CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER 'gameserver'@'%' IDENTIFIED BY 'your-secure-password';
GRANT ALL PRIVILEGES ON gameserver.* TO 'gameserver'@'%';
FLUSH PRIVILEGES;

-- 运行初始化脚本
SOURCE /path/to/GameServer/internal/database/init_tables.sql;
```

### 防火墙配置
```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 8080/tcp

# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# 阿里云安全组
# 在阿里云控制台添加入方向规则：端口8080，协议TCP
```

## 🔧 管理和维护

### 使用管理脚本
```bash
# 给管理脚本执行权限
chmod +x server_manager.sh

# 查看帮助
./server_manager.sh help

# 常用操作
./server_manager.sh status    # 查看状态
./server_manager.sh start     # 启动服务
./server_manager.sh stop      # 停止服务
./server_manager.sh logs      # 查看日志
./server_manager.sh health    # 健康检查
./server_manager.sh update    # 更新服务
./server_manager.sh screen    # Screen会话管理
```

### 日志管理
```bash
# systemd 日志
sudo journalctl -u gameserver -f --since "1 hour ago"

# 文件日志（如果配置了）
tail -f /var/log/gameserver/gameserver.log

# 清理旧日志
sudo journalctl --vacuum-time=7d
```

### 监控和告警
```bash
# 检查服务状态
curl http://localhost:8080/health

# 检查指标
curl http://localhost:8080/metrics

# 简单的监控脚本
cat > monitor.sh << 'EOF'
#!/bin/bash
if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "GameServer 健康检查失败！" | mail -s "服务器告警" your-email@example.com
    systemctl restart gameserver
fi
EOF

# 添加到crontab（每5分钟检查一次）
echo "*/5 * * * * /path/to/monitor.sh" | crontab -
```

## 🚨 故障排除

### 常见问题

#### 1. 服务启动失败
```bash
# 查看详细错误信息
sudo journalctl -u gameserver -n 50

# 检查配置文件
sudo systemctl cat gameserver

# 检查文件权限
ls -la /opt/gameserver/
```

#### 2. 数据库连接失败
```bash
# 测试数据库连接
mysql -h your-db-host -u your-db-user -p your-db-name

# 检查防火墙
telnet your-db-host 3306
```

#### 3. 端口被占用
```bash
# 查看端口占用
sudo netstat -tlnp | grep :8080
sudo lsof -i :8080

# 修改配置使用其他端口
sudo systemctl edit gameserver
```

#### 4. 内存不足
```bash
# 检查内存使用
free -h
top -p $(pgrep gameserver)

# 调整系统参数
sudo sysctl vm.swappiness=10
```

### 性能优化
```bash
# 调整文件描述符限制
echo "gameserver soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "gameserver hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# 调整TCP参数
sudo sysctl net.core.somaxconn=65535
sudo sysctl net.ipv4.tcp_max_syn_backlog=65535
```

## 📞 获取帮助

如果遇到问题，请：

1. 查看日志文件获取错误信息
2. 检查系统资源使用情况
3. 验证网络连接和防火墙配置
4. 参考项目文档和源代码

## 🔄 自动化部署（CI/CD）

可以考虑使用以下工具实现自动化部署：
- GitHub Actions
- GitLab CI/CD
- Jenkins
- 阿里云云效

这样每次代码更新都可以自动部署到服务器。