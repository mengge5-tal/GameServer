#!/bin/bash

# WebSocket 连接诊断脚本
echo "================================================"
echo "WebSocket 连接诊断脚本"
echo "服务器IP: 101.201.51.135"
echo "端口: 8080"
echo "WebSocket路径: /ws"
echo "================================================"
echo

# 1. 检查服务器基本状态
echo "1. 检查服务器进程状态"
echo "--------------------------------"
ps aux | grep gameserver | grep -v grep
if [ $? -eq 0 ]; then
    echo "✓ GameServer 进程正在运行"
else
    echo "✗ GameServer 进程未运行"
fi
echo

# 2. 检查端口监听状态
echo "2. 检查端口 8080 监听状态"
echo "--------------------------------"
netstat -tlnp | grep :8080
if [ $? -eq 0 ]; then
    echo "✓ 端口 8080 正在监听"
else
    echo "✗ 端口 8080 未监听"
fi
echo

# 3. 检查防火墙状态（CentOS/RHEL）
echo "3. 检查防火墙状态"
echo "--------------------------------"
if command -v firewall-cmd &> /dev/null; then
    echo "使用 firewalld："
    systemctl status firewalld --no-pager | head -3
    echo
    echo "检查端口 8080 是否开放："
    firewall-cmd --query-port=8080/tcp
    if [ $? -eq 0 ]; then
        echo "✓ 防火墙已开放端口 8080"
    else
        echo "✗ 防火墙未开放端口 8080"
        echo "建议运行: firewall-cmd --permanent --add-port=8080/tcp && firewall-cmd --reload"
    fi
    echo
    echo "当前开放的端口："
    firewall-cmd --list-ports
elif command -v iptables &> /dev/null; then
    echo "使用 iptables："
    iptables -L INPUT -n | grep 8080
    if [ $? -eq 0 ]; then
        echo "✓ iptables 规则包含端口 8080"
    else
        echo "? iptables 未找到明确的 8080 端口规则"
    fi
else
    echo "未找到防火墙管理工具"
fi
echo

# 4. 本地网络接口测试
echo "4. 本地网络接口测试"
echo "--------------------------------"
echo "测试本地 HTTP 连接："
timeout 5 curl -s -I http://localhost:8080/health
if [ $? -eq 0 ]; then
    echo "✓ 本地 HTTP 连接成功"
else
    echo "✗ 本地 HTTP 连接失败"
fi
echo

echo "测试本地 WebSocket 连接（使用 websocat 如果可用）："
if command -v websocat &> /dev/null; then
    timeout 5 echo '{"type":"heartbeat","action":"ping"}' | websocat ws://localhost:8080/ws
    if [ $? -eq 0 ]; then
        echo "✓ 本地 WebSocket 连接成功"
    else
        echo "✗ 本地 WebSocket 连接失败"
    fi
else
    echo "websocat 工具未安装，跳过本地 WebSocket 测试"
fi
echo

# 5. 外部网络连接测试
echo "5. 外部网络连接测试"
echo "--------------------------------"
echo "从公网IP测试 HTTP 连接："
timeout 10 curl -s -I http://101.201.51.135:8080/health
if [ $? -eq 0 ]; then
    echo "✓ 外部 HTTP 连接成功"
    echo "响应内容："
    timeout 5 curl -s http://101.201.51.135:8080/health | head -5
else
    echo "✗ 外部 HTTP 连接失败"
fi
echo

# 6. 网络连通性测试
echo "6. 网络连通性测试"
echo "--------------------------------"
echo "测试端口连通性："
timeout 5 telnet 101.201.51.135 8080 << EOF
quit
EOF
if [ $? -eq 0 ]; then
    echo "✓ 端口 8080 可达"
else
    echo "✗ 端口 8080 不可达"
fi
echo

echo "使用 nc (netcat) 测试端口："
if command -v nc &> /dev/null; then
    timeout 5 nc -zv 101.201.51.135 8080
    if [ $? -eq 0 ]; then
        echo "✓ nc 测试端口 8080 可达"
    else
        echo "✗ nc 测试端口 8080 不可达"
    fi
else
    echo "nc 工具未安装"
fi
echo

# 7. 检查系统资源
echo "7. 系统资源检查"
echo "--------------------------------"
echo "内存使用情况："
free -h | head -2
echo
echo "CPU 负载："
uptime
echo
echo "磁盘使用情况："
df -h | head -2
echo

# 8. 检查服务器日志（最近的错误）
echo "8. 检查最近的系统日志"
echo "--------------------------------"
echo "最近的系统错误日志："
journalctl --since "5 minutes ago" --priority=err --no-pager | tail -10
echo

# 9. 网络配置检查
echo "9. 网络配置检查"
echo "--------------------------------"
echo "网络接口信息："
ip addr show | grep -A 5 "101.201.51.135"
echo
echo "路由信息："
ip route | head -5
echo

# 10. WebSocket 特定检查
echo "10. WebSocket 特定检查"
echo "--------------------------------"
echo "检查 WebSocket 升级响应："
timeout 10 curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" http://101.201.51.135:8080/ws
if [ $? -eq 0 ]; then
    echo "✓ WebSocket 握手响应正常"
else
    echo "✗ WebSocket 握手失败"
fi
echo

# 11. 安全组和云服务器特定检查
echo "11. 云服务器安全组检查"
echo "--------------------------------"
echo "提醒检查以下阿里云配置："
echo "1. ECS 安全组规则是否开放 8080 端口"
echo "2. 入方向规则: 协议类型=TCP, 端口范围=8080/8080, 授权对象=0.0.0.0/0"
echo "3. 确认 ECS 实例的内网IP和公网IP绑定正确"
echo

echo "================================================"
echo "诊断完成"
echo "================================================"
echo "如果外部连接失败，请检查："
echo "1. 阿里云 ECS 安全组规则"
echo "2. 服务器防火墙设置"
echo "3. 应用程序是否正确绑定到 0.0.0.0:8080"
echo "4. 网络ACL或其他安全策略"