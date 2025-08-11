#!/bin/bash

# GameServer 管理脚本
# 用于管理游戏服务器的启动、停止、重启、状态查看等

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

SERVICE_NAME="gameserver"
APP_DIR="/opt/gameserver"

show_help() {
    echo -e "${GREEN}GameServer 管理脚本${NC}"
    echo
    echo -e "${YELLOW}使用方法:${NC}"
    echo -e "  $0 [命令]"
    echo
    echo -e "${YELLOW}可用命令:${NC}"
    echo -e "  start     - 启动服务器"
    echo -e "  stop      - 停止服务器"  
    echo -e "  restart   - 重启服务器"
    echo -e "  status    - 查看服务器状态"
    echo -e "  logs      - 查看实时日志"
    echo -e "  health    - 健康检查"
    echo -e "  update    - 更新服务器"
    echo -e "  screen    - 使用screen方式管理"
    echo -e "  help      - 显示帮助信息"
}

check_systemd() {
    if systemctl list-unit-files | grep -q "^${SERVICE_NAME}.service"; then
        return 0
    else
        return 1
    fi
}

start_service() {
    if check_systemd; then
        echo -e "${BLUE}使用 systemd 启动服务...${NC}"
        sudo systemctl start $SERVICE_NAME
        if systemctl is-active --quiet $SERVICE_NAME; then
            echo -e "${GREEN}✅ 服务启动成功${NC}"
        else
            echo -e "${RED}❌ 服务启动失败${NC}"
            sudo systemctl status $SERVICE_NAME --no-pager
        fi
    else
        echo -e "${YELLOW}systemd 服务不存在，使用 screen 启动...${NC}"
        start_screen
    fi
}

stop_service() {
    if check_systemd; then
        echo -e "${BLUE}使用 systemd 停止服务...${NC}"
        sudo systemctl stop $SERVICE_NAME
        echo -e "${GREEN}✅ 服务已停止${NC}"
    else
        echo -e "${BLUE}停止 screen 会话...${NC}"
        screen -S gameserver -X quit 2>/dev/null && echo -e "${GREEN}✅ screen 会话已停止${NC}" || echo -e "${YELLOW}没有运行的 screen 会话${NC}"
    fi
}

restart_service() {
    echo -e "${BLUE}重启服务器...${NC}"
    stop_service
    sleep 2
    start_service
}

show_status() {
    if check_systemd; then
        echo -e "${BLUE}systemd 服务状态:${NC}"
        sudo systemctl status $SERVICE_NAME --no-pager
    fi
    
    echo -e "\n${BLUE}screen 会话状态:${NC}"
    screen -ls | grep gameserver || echo "没有运行的 gameserver screen 会话"
    
    echo -e "\n${BLUE}进程状态:${NC}"
    ps aux | grep gameserver | grep -v grep || echo "没有运行的 gameserver 进程"
    
    echo -e "\n${BLUE}端口监听状态:${NC}"
    netstat -tlnp 2>/dev/null | grep :8080 || ss -tlnp | grep :8080 || echo "端口 8080 未被监听"
}

show_logs() {
    if check_systemd; then
        echo -e "${BLUE}查看 systemd 服务日志 (按 Ctrl+C 退出):${NC}"
        sudo journalctl -u $SERVICE_NAME -f
    else
        echo -e "${BLUE}连接到 screen 会话查看日志:${NC}"
        screen -r gameserver
    fi
}

health_check() {
    echo -e "${BLUE}执行健康检查...${NC}"
    
    # 检查端口
    if netstat -tlnp 2>/dev/null | grep -q :8080 || ss -tlnp 2>/dev/null | grep -q :8080; then
        echo -e "✓ 端口 8080 正在监听"
    else
        echo -e "${RED}❌ 端口 8080 未监听${NC}"
        return 1
    fi
    
    # 检查HTTP健康检查端点
    if command -v curl >/dev/null 2>&1; then
        echo -e "${BLUE}检查 HTTP 端点...${NC}"
        if curl -s -f http://localhost:8080/health > /dev/null; then
            echo -e "✓ /health 端点响应正常"
        else
            echo -e "${YELLOW}⚠️  /health 端点无响应${NC}"
        fi
        
        if curl -s -f http://localhost:8080/metrics > /dev/null; then
            echo -e "✓ /metrics 端点响应正常"
        else
            echo -e "${YELLOW}⚠️  /metrics 端点无响应${NC}"
        fi
    else
        echo -e "${YELLOW}curl 未安装，跳过 HTTP 检查${NC}"
    fi
    
    echo -e "${GREEN}✅ 健康检查完成${NC}"
}

update_service() {
    echo -e "${BLUE}更新服务器...${NC}"
    
    # 构建新版本
    echo "构建新版本..."
    go build -o gameserver-new ./cmd/server
    
    if [[ ! -f gameserver-new ]]; then
        echo -e "${RED}❌ 构建失败${NC}"
        return 1
    fi
    
    # 停止服务
    stop_service
    
    # 备份旧版本
    if [[ -f gameserver ]]; then
        mv gameserver gameserver.backup.$(date +%Y%m%d_%H%M%S)
    fi
    
    # 替换新版本
    mv gameserver-new gameserver
    chmod +x gameserver
    
    # 如果是systemd服务，复制到应用目录
    if check_systemd && [[ -d $APP_DIR ]]; then
        sudo cp gameserver $APP_DIR/
        sudo chown gameserver:gameserver $APP_DIR/gameserver
    fi
    
    # 启动服务
    start_service
    
    echo -e "${GREEN}✅ 服务器更新完成${NC}"
}

start_screen() {
    # 检查是否已有screen会话
    if screen -ls | grep -q gameserver; then
        echo -e "${YELLOW}gameserver screen 会话已存在${NC}"
        echo -e "连接现有会话: screen -r gameserver"
        return 0
    fi
    
    # 检查可执行文件
    if [[ ! -f gameserver ]]; then
        echo -e "${YELLOW}gameserver 可执行文件不存在，正在构建...${NC}"
        go build -o gameserver ./cmd/server
    fi
    
    # 创建启动脚本
    cat > .start_gameserver.sh << 'EOF'
#!/bin/bash
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-3306}
export DB_NAME=${DB_NAME:-gameserver}
export DB_USER=${DB_USER:-gameserver}
export DB_PASSWORD=${DB_PASSWORD:-password}
export SERVER_HOST=${SERVER_HOST:-0.0.0.0}
export SERVER_PORT=${SERVER_PORT:-8080}

echo "启动 GameServer..."
echo "数据库: $DB_HOST:$DB_PORT/$DB_NAME"
echo "服务器: $SERVER_HOST:$SERVER_PORT"
echo "按 Ctrl+A 然后 D 分离会话"
echo "=========================="

./gameserver
EOF
    
    chmod +x .start_gameserver.sh
    
    # 创建screen会话
    screen -dmS gameserver bash .start_gameserver.sh
    sleep 1
    
    if screen -ls | grep -q gameserver; then
        echo -e "${GREEN}✅ gameserver screen 会话已启动${NC}"
        echo -e "连接会话: screen -r gameserver"
        echo -e "分离会话: Ctrl+A 然后按 D"
        echo -e "停止会话: screen -S gameserver -X quit"
    else
        echo -e "${RED}❌ screen 会话启动失败${NC}"
    fi
}

manage_screen() {
    echo -e "${GREEN}Screen 会话管理${NC}"
    echo
    echo -e "${YELLOW}选择操作:${NC}"
    echo "1) 启动新的 screen 会话"
    echo "2) 连接现有 screen 会话" 
    echo "3) 查看所有 screen 会话"
    echo "4) 停止 gameserver screen 会话"
    echo "5) 返回主菜单"
    echo
    read -p "请选择 (1-5): " choice
    
    case $choice in
        1) start_screen ;;
        2) 
            if screen -ls | grep -q gameserver; then
                screen -r gameserver
            else
                echo -e "${RED}没有运行的 gameserver 会话${NC}"
            fi
            ;;
        3) screen -ls ;;
        4) 
            screen -S gameserver -X quit 2>/dev/null && echo -e "${GREEN}✅ 已停止${NC}" || echo -e "${YELLOW}没有运行的会话${NC}"
            ;;
        5) return ;;
        *) echo -e "${RED}无效选择${NC}" ;;
    esac
}

# 主逻辑
case "${1:-help}" in
    start)
        start_service
        ;;
    stop)
        stop_service
        ;;
    restart)
        restart_service
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    health)
        health_check
        ;;
    update)
        update_service
        ;;
    screen)
        manage_screen
        ;;
    help)
        show_help
        ;;
    *)
        echo -e "${RED}未知命令: $1${NC}"
        show_help
        exit 1
        ;;
esac