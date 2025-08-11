#!/bin/bash

# GameServer 快速部署脚本（使用screen方案）
# 适用于快速解决当前问题

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}GameServer 快速部署 (使用 screen)${NC}"

# 1. 安装 screen
echo -e "${BLUE}安装 screen...${NC}"
if command -v apt >/dev/null 2>&1; then
    sudo apt update && sudo apt install -y screen
elif command -v yum >/dev/null 2>&1; then
    sudo yum install -y screen
else
    echo -e "${YELLOW}请手动安装 screen${NC}"
fi

# 2. 构建应用
echo -e "${BLUE}构建应用...${NC}"
go build -o gameserver ./cmd/server

# 3. 创建启动脚本
cat > start_server.sh << 'EOF'
#!/bin/bash
echo "启动 GameServer..."
export DB_HOST=rm-2zevr95ez9rrid70uho.mysql.rds.aliyuncs.com
export DB_PORT=3306
export DB_NAME=Vampire
export DB_USER=wwk18255113901
export DB_PASSWORD=BaiChen123456+
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

./gameserver
EOF

chmod +x start_server.sh

# 4. 创建 screen 会话并启动服务器
echo -e "${BLUE}创建 screen 会话并启动服务器...${NC}"
screen -dmS gameserver bash -c './start_server.sh'

echo -e "${GREEN}✅ GameServer 已在 screen 会话中启动！${NC}"
echo
echo -e "${YELLOW}常用命令:${NC}"
echo -e "  • 查看会话: screen -ls"
echo -e "  • 连接会话: screen -r gameserver"
echo -e "  • 分离会话: Ctrl+A 然后按 D"
echo -e "  • 停止服务器: screen -S gameserver -X quit"
echo
echo -e "${BLUE}现在你可以安全地关闭SSH连接，服务器将继续运行！${NC}"