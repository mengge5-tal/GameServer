#!/bin/bash

# GameServer 停止脚本

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}正在停止 GameServer...${NC}"

# 1. 尝试停止 systemd 服务
if systemctl is-active --quiet gameserver 2>/dev/null; then
    echo -e "${BLUE}检测到 systemd 服务，正在停止...${NC}"
    sudo systemctl stop gameserver
    if ! systemctl is-active --quiet gameserver 2>/dev/null; then
        echo -e "${GREEN}✅ systemd 服务已停止${NC}"
    else
        echo -e "${RED}❌ systemd 服务停止失败${NC}"
    fi
fi

# 2. 尝试停止 screen 会话
if screen -ls 2>/dev/null | grep -q gameserver; then
    echo -e "${BLUE}检测到 screen 会话，正在停止...${NC}"
    screen -S gameserver -X quit
    sleep 1
    if ! screen -ls 2>/dev/null | grep -q gameserver; then
        echo -e "${GREEN}✅ screen 会话已停止${NC}"
    else
        echo -e "${RED}❌ screen 会话停止失败${NC}"
    fi
fi

# 3. 检查并终止残留进程
if pgrep -f "gameserver" > /dev/null; then
    echo -e "${BLUE}检测到残留进程，正在终止...${NC}"
    
    # 优雅停止
    pkill -TERM -f gameserver
    sleep 3
    
    # 如果还有进程，强制终止
    if pgrep -f "gameserver" > /dev/null; then
        echo -e "${YELLOW}优雅停止失败，强制终止进程...${NC}"
        pkill -KILL -f gameserver
        sleep 1
    fi
    
    if ! pgrep -f "gameserver" > /dev/null; then
        echo -e "${GREEN}✅ 所有进程已终止${NC}"
    else
        echo -e "${RED}❌ 仍有进程运行${NC}"
        echo -e "${YELLOW}手动终止进程:${NC}"
        ps aux | grep gameserver | grep -v grep
    fi
else
    echo -e "${GREEN}✅ 没有检测到运行中的进程${NC}"
fi

# 4. 检查端口占用
if netstat -tlnp 2>/dev/null | grep -q :8080 || ss -tlnp 2>/dev/null | grep -q :8080; then
    echo -e "${RED}⚠️  端口 8080 仍被占用${NC}"
    echo -e "${YELLOW}端口占用详情:${NC}"
    netstat -tlnp 2>/dev/null | grep :8080 || ss -tlnp | grep :8080
else
    echo -e "${GREEN}✅ 端口 8080 已释放${NC}"
fi

echo -e "${GREEN}==================================${NC}"
echo -e "${GREEN}GameServer 停止操作完成！${NC}"
echo -e "${GREEN}==================================${NC}"

# 显示最终状态
echo -e "\n${BLUE}最终状态检查:${NC}"
echo -e "systemd 服务: $(systemctl is-active gameserver 2>/dev/null || echo '不存在')"
echo -e "screen 会话: $(screen -ls 2>/dev/null | grep gameserver || echo '无')"
echo -e "运行进程: $(pgrep -f gameserver | wc -l) 个"
echo -e "端口监听: $(netstat -tlnp 2>/dev/null | grep -c :8080 || echo '0') 个"