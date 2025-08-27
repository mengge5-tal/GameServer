#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
游戏服务器聊天系统测试脚本
测试私聊、世界聊天和工会聊天功能
"""

import asyncio
import websockets
import json
import time

SERVER_URL = "ws://101.201.51.135:8080/ws"

class ChatTester:
    def __init__(self, user_id=None):
        self.ws = None
        self.user_id = user_id
        self.request_id = 1
        
    def generate_request_id(self):
        req_id = f"test_{self.request_id}"
        self.request_id += 1
        return req_id
    
    async def connect(self):
        try:
            self.ws = await websockets.connect(SERVER_URL)
            print(f"✅ 连接成功: {SERVER_URL}")
            return True
        except Exception as e:
            print(f"❌ 连接失败: {e}")
            return False
    
    async def send_message(self, msg_type, action, data):
        message = {
            "type": msg_type,
            "action": action,
            "data": data,
            "requestId": self.generate_request_id(),
            "timestamp": int(time.time())
        }
        
        try:
            await self.ws.send(json.dumps(message))
            print(f"📤 发送: {msg_type}.{action} - {json.dumps(data, ensure_ascii=False)}")
            
            # 等待响应
            response = await self.ws.recv()
            response_data = json.loads(response)
            
            if response_data.get('success'):
                print(f"✅ 成功响应: {json.dumps(response_data, ensure_ascii=False)}")
            else:
                print(f"❌ 错误响应: {response_data.get('message', 'Unknown error')}")
            
            return response_data
        except Exception as e:
            print(f"❌ 发送消息失败: {e}")
            return None
    
    async def register_and_login(self, username, password):
        print(f"\n=== 注册和登录用户: {username} ===")
        
        # 先尝试注册
        register_result = await self.send_message("auth", "register", {
            "username": username,
            "password": password
        })
        
        # 登录
        login_result = await self.send_message("auth", "login", {
            "username": username,
            "password": password
        })
        
        if login_result and login_result.get('success'):
            user_data = login_result.get('data', {})
            self.user_id = user_data.get('userid')  # 注意字段名是userid不是user_id
            print(f"🎯 登录成功，用户ID: {self.user_id}")
            return True
        
        return False
    
    async def test_private_chat(self, target_user_id, message):
        print(f"\n=== 测试私聊功能 ===")
        print(f"发送私聊给用户 {target_user_id}: {message}")
        
        result = await self.send_message("chat", "sendPrivateMessage", {
            "to_user_id": target_user_id,
            "content": message
        })
        
        # 获取私聊消息
        await asyncio.sleep(0.5)
        await self.send_message("chat", "getPrivateMessages", {
            "other_user_id": target_user_id,
            "page": 1,
            "limit": 10
        })
        
        return result
    
    async def test_world_chat(self, channel_id, message):
        print(f"\n=== 测试世界聊天功能 ===")
        
        # 加入频道
        await self.send_message("chat", "joinWorldChannel", {
            "channel_id": channel_id
        })
        
        await asyncio.sleep(0.5)
        
        # 发送世界消息
        print(f"发送世界消息到频道 {channel_id}: {message}")
        result = await self.send_message("chat", "sendWorldMessage", {
            "content": message
        })
        
        # 获取频道列表
        await asyncio.sleep(0.5)
        await self.send_message("chat", "getWorldChannels", {})
        
        return result
    
    async def test_union_chat(self, message):
        print(f"\n=== 测试工会聊天功能 ===")
        print(f"发送工会消息: {message}")
        
        # 发送工会消息
        result = await self.send_message("chat", "sendUnionMessage", {
            "content": message
        })
        
        # 获取工会消息
        await asyncio.sleep(0.5)
        await self.send_message("chat", "getRecentUnionMessages", {
            "limit": 10
        })
        
        return result
    
    async def close(self):
        if self.ws:
            await self.ws.close()
            print("🔌 连接已关闭")

async def run_chat_tests():
    """运行完整的聊天系统测试"""
    print("🚀 开始聊天系统测试...")
    
    # 创建两个测试用户
    user1 = ChatTester()
    user2 = ChatTester()
    
    try:
        # 连接到服务器
        if not await user1.connect():
            return
        if not await user2.connect():
            return
        
        # 注册和登录用户
        await user1.register_and_login("chat_test_user1", "Aaa123456!")
        await user2.register_and_login("chat_test_user2", "Aaa123456!")
        
        if not user1.user_id or not user2.user_id:
            print("❌ 用户登录失败，无法继续测试")
            return
        
        print(f"\n📋 测试用户信息:")
        print(f"   用户1 ID: {user1.user_id}")
        print(f"   用户2 ID: {user2.user_id}")
        
        # 测试私聊
#         await user1.test_private_chat(user2.user_id, "你好，这是来自用户1的私聊消息")
#         await asyncio.sleep(1)
#         await user2.test_private_chat(user1.user_id, "收到！这是用户2的回复")
#
        # 测试世界聊天
#         await user1.test_world_chat(1, "用户1在世界频道1发言")
#         await asyncio.sleep(1)
#         await user2.test_world_chat(1, "用户2也在世界频道1发言")
        
        # 测试工会聊天（需要用户在同一工会）
        await user1.test_union_chat("测试测试测试测试")
        await asyncio.sleep(1)
        await user2.test_union_chat("测测测测试试试试")
        
        print("\n🎉 聊天系统测试完成！")
        
    except Exception as e:
        print(f"❌ 测试过程中出现错误: {e}")
    finally:
        await user1.close()
        await user2.close()

async def simple_chat_test():
    """简单的聊天测试 - 单用户测试基础功能"""
    print("🔧 运行简单聊天测试...")
    
    tester = ChatTester()
    
    try:
        if not await tester.connect():
            return
            
        # 登录测试用户
        if await tester.register_and_login("simple_test_user", "Aaa123456!"):
            print(f"✅ 登录成功，用户ID: {tester.user_id}")
            
            # 测试心跳
            await tester.send_message("heartbeat", "ping", {})
            await asyncio.sleep(0.5)
            
            # 测试世界聊天
            await tester.send_message("chat", "joinWorldChannel", {"channel_id": 1})
            await asyncio.sleep(0.5)
            
            await tester.send_message("chat", "sendWorldMessage", {"content": "测试消息"})
            await asyncio.sleep(0.5)
            
            await tester.send_message("chat", "getWorldChannels", {})
            
            print("✅ 基础聊天功能测试完成")
        
    except Exception as e:
        print(f"❌ 简单测试失败: {e}")
    finally:
        await tester.close()

if __name__ == "__main__":
    print("聊天系统测试工具")
    print("================")
    print("1. 完整测试（需要两个用户）")
    print("2. 简单测试（单用户基础功能）")
    
    choice = input("请选择测试类型 (1/2): ").strip()
    
    if choice == "1":
        asyncio.run(run_chat_tests())
    elif choice == "2":
        asyncio.run(simple_chat_test())
    else:
        print("无效选择，运行简单测试...")
        asyncio.run(simple_chat_test())