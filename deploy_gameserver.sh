#!/bin/bash

# GameServer 阿里云部署脚本
# 使用方法：./deploy_gameserver.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量 - 请根据你的实际情况修改
APP_NAME="gameserver"
APP_DIR="/opt/gameserver"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"
LOG_DIR="/var/log/gameserver"
USER_NAME="gameserver"  # 建议创建专用用户
BUILD_NAME="gameserver"

echo -e "${GREEN}==================================================================="
echo -e "GameServer 阿里云部署脚本"
echo -e "===================================================================${NC}"

# 1. 检查运行权限
if [[ $EUID -eq 0 ]]; then
   echo -e "${RED}请不要使用root用户运行此脚本${NC}"
   echo -e "${YELLOW}建议使用有sudo权限的普通用户${NC}"
   exit 1
fi

echo -e "${BLUE}步骤 1: 检查系统环境...${NC}"

# 检查系统
if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    echo -e "✓ 操作系统: $NAME $VERSION"
else
    echo -e "${RED}❌ 无法识别操作系统${NC}"
    exit 1
fi

# 检查Go环境
if command -v go >/dev/null 2>&1; then
    echo -e "✓ Go 版本: $(go version)"
else
    echo -e "${RED}❌ Go 未安装，请先安装 Go 1.19+${NC}"
    exit 1
fi

echo -e "${BLUE}步骤 2: 安装必要的系统工具...${NC}"

# 安装必要工具
sudo apt update -y
sudo apt install -y systemctl curl wget htop

echo -e "${BLUE}步骤 3: 创建应用用户和目录...${NC}"

# 创建专用用户（如果不存在）
if ! id "$USER_NAME" &>/dev/null; then
    echo "创建用户: $USER_NAME"
    sudo useradd -r -s /bin/false $USER_NAME
else
    echo -e "✓ 用户 $USER_NAME 已存在"
fi

# 创建必要目录
sudo mkdir -p $APP_DIR
sudo mkdir -p $LOG_DIR
sudo chown -R $USER_NAME:$USER_NAME $APP_DIR
sudo chown -R $USER_NAME:$USER_NAME $LOG_DIR

echo -e "${BLUE}步骤 4: 构建应用程序...${NC}"

# 构建应用
echo "正在构建 GameServer..."
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

go build -a -installsuffix cgo -ldflags="-w -s" -o $BUILD_NAME ./cmd/server

if [[ ! -f $BUILD_NAME ]]; then
    echo -e "${RED}❌ 构建失败${NC}"
    exit 1
fi

echo -e "✓ 构建成功: $(ls -lh $BUILD_NAME)"

echo -e "${BLUE}步骤 5: 部署应用文件...${NC}"

# 停止现有服务
if systemctl is-active --quiet $APP_NAME 2>/dev/null; then
    echo "停止现有服务..."
    sudo systemctl stop $APP_NAME
fi

# 复制应用文件
sudo cp $BUILD_NAME $APP_DIR/
sudo cp -r internal $APP_DIR/ 2>/dev/null || true
sudo cp -r web $APP_DIR/ 2>/dev/null || true
sudo chown -R $USER_NAME:$USER_NAME $APP_DIR
sudo chmod +x $APP_DIR/$BUILD_NAME

echo -e "✓ 应用文件已部署到 $APP_DIR"

echo -e "${BLUE}步骤 6: 配置环境变量...${NC}"

# 提示用户输入数据库配置
echo -e "${YELLOW}请输入数据库配置信息:${NC}"
read -p "数据库主机 (默认: localhost): " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "数据库端口 (默认: 3306): " DB_PORT
DB_PORT=${DB_PORT:-3306}

read -p "数据库名称 (默认: gameserver): " DB_NAME
DB_NAME=${DB_NAME:-gameserver}

read -p "数据库用户名: " DB_USER
if [[ -z "$DB_USER" ]]; then
    echo -e "${RED}❌ 数据库用户名不能为空${NC}"
    exit 1
fi

read -s -p "数据库密码: " DB_PASSWORD
echo
if [[ -z "$DB_PASSWORD" ]]; then
    echo -e "${RED}❌ 数据库密码不能为空${NC}"
    exit 1
fi

read -p "服务器端口 (默认: 8080): " SERVER_PORT
SERVER_PORT=${SERVER_PORT:-8080}

echo -e "${BLUE}步骤 7: 创建 systemd 服务...${NC}"

# 创建systemd服务文件
sudo tee $SERVICE_FILE > /dev/null <<EOF
[Unit]
Description=GameServer - WebSocket Game Server
Documentation=https://github.com/your-username/GameServer
After=network.target mysql.service
Wants=network.target
StartLimitBurst=3
StartLimitIntervalSec=10

[Service]
Type=simple
User=$USER_NAME
Group=$USER_NAME
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$BUILD_NAME
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
Restart=always
RestartSec=5
RestartPreventExitStatus=23

# 标准输出和错误输出
StandardOutput=append:$LOG_DIR/gameserver.log
StandardError=append:$LOG_DIR/gameserver-error.log

# 环境变量
Environment=DB_HOST=$DB_HOST
Environment=DB_PORT=$DB_PORT
Environment=DB_NAME=$DB_NAME
Environment=DB_USER=$DB_USER
Environment=DB_PASSWORD=$DB_PASSWORD
Environment=SERVER_HOST=0.0.0.0
Environment=SERVER_PORT=$SERVER_PORT
Environment=LOG_LEVEL=info
Environment=LOG_FORMAT=json

# 安全设置
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ReadWritePaths=$APP_DIR $LOG_DIR

# 资源限制
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

echo -e "✓ systemd 服务文件已创建"

echo -e "${BLUE}步骤 8: 启用并启动服务...${NC}"

# 重载systemd配置
sudo systemctl daemon-reload

# 启用开机自启
sudo systemctl enable $APP_NAME

# 启动服务
sudo systemctl start $APP_NAME

# 等待服务启动
sleep 3

echo -e "${BLUE}步骤 9: 检查服务状态...${NC}"

# 检查服务状态
if systemctl is-active --quiet $APP_NAME; then
    echo -e "${GREEN}✅ GameServer 服务启动成功！${NC}"
    echo -e "服务状态: $(systemctl is-active $APP_NAME)"
    echo -e "服务地址: http://$(curl -s ifconfig.me):$SERVER_PORT"
    echo -e "WebSocket: ws://$(curl -s ifconfig.me):$SERVER_PORT/ws"
else
    echo -e "${RED}❌ GameServer 服务启动失败${NC}"
    echo -e "${YELLOW}查看错误日志:${NC}"
    sudo journalctl -u $APP_NAME --no-pager -n 20
    exit 1
fi

echo -e "${BLUE}步骤 10: 配置防火墙...${NC}"

# 配置防火墙（如果有ufw）
if command -v ufw >/dev/null 2>&1; then
    echo "配置防火墙规则..."
    sudo ufw allow $SERVER_PORT/tcp
    echo -e "✓ 防火墙规则已添加: 端口 $SERVER_PORT"
fi

echo -e "${GREEN}==================================================================="
echo -e "🎉 GameServer 部署完成！"
echo -e "==================================================================="
echo -e "服务信息:"
echo -e "  • 服务名称: $APP_NAME"
echo -e "  • 安装目录: $APP_DIR"
echo -e "  • 日志目录: $LOG_DIR"
echo -e "  • 访问地址: http://$(curl -s ifconfig.me):$SERVER_PORT"
echo -e "  • WebSocket: ws://$(curl -s ifconfig.me):$SERVER_PORT/ws"
echo -e ""
echo -e "常用命令:"
echo -e "  • 查看状态: sudo systemctl status $APP_NAME"
echo -e "  • 启动服务: sudo systemctl start $APP_NAME"
echo -e "  • 停止服务: sudo systemctl stop $APP_NAME"
echo -e "  • 重启服务: sudo systemctl restart $APP_NAME"
echo -e "  • 查看日志: sudo journalctl -u $APP_NAME -f"
echo -e "  • 查看文件日志: tail -f $LOG_DIR/gameserver.log"
echo -e ""
echo -e "健康检查:"
echo -e "  • curl http://localhost:$SERVER_PORT/health"
echo -e "  • curl http://localhost:$SERVER_PORT/metrics"
echo -e "${NC}"

# 清理构建文件
rm -f $BUILD_NAME

echo -e "${GREEN}部署脚本执行完成！${NC}"