// Zinx WebSocket 服务器示例
// 功能：展示如何使用 Zinx 框架搭建 WebSocket 服务器
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

// ChatRouter 处理聊天消息
type ChatRouter struct {
	znet.BaseRouter
}

func (r *ChatRouter) Handle(request ziface.IRequest) {
	conn := request.GetConnection()
	msg := string(request.GetData())

	fmt.Printf("[WebSocket Chat] Client %d: %s\n", conn.GetConnID(), msg)

	// 广播消息给所有连接 (简化示例，实际应使用连接管理器)
	reply := fmt.Sprintf("Server received: %s", msg)
	if err := conn.SendMsg(request.GetMsgID(), []byte(reply)); err != nil {
		fmt.Println("Send error:", err)
	}
}

// AuthRouter 处理认证消息
type AuthRouter struct {
	znet.BaseRouter
}

func (r *AuthRouter) Handle(request ziface.IRequest) {
	conn := request.GetConnection()
	token := string(request.GetData())

	fmt.Printf("[Auth] Client %d token: %s\n", conn.GetConnID(), token)

	// 简单 token 验证
	if token == "valid-token" {
		conn.SetProperty("authenticated", true)
		conn.SetProperty("userId", 12345)
		conn.SendMsg(100, []byte("Auth success"))
	} else {
		conn.SendMsg(100, []byte("Auth failed"))
		time.Sleep(100 * time.Millisecond)
		conn.Stop()
	}
}

// WebSocket 认证函数
func websocketAuth(r *http.Request) error {
	// 从 URL 参数获取 token
	token := r.URL.Query().Get("token")
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// 从 Header 获取 token
	if token == "" {
		token = r.Header.Get("Authorization")
	}

	if token != "valid-token" && token != "Bearer valid-token" {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func main() {
	// 创建 WebSocket 服务器
	server := znet.NewServer(func(s *znet.Server) {
		s.Name = "ZinxWebSocketServer"
		s.TCPPort = 8999     // TCP 端口
		s.WsPort = 9000      // WebSocket 端口
		s.WsPath = "/ws"     // WebSocket 路径
		s.Mode = "websocket" // 启用 WebSocket 模式

		// 设置 WebSocket 认证
		s.SetWebsocketAuth(websocketAuth)
	})

	// 注册路由
	server.AddRouter(1, &ChatRouter{})  // 聊天消息
	server.AddRouter(2, &AuthRouter{})  // 认证消息
	server.AddRouter(99, &ChatRouter{}) // 欢迎消息响应

	// 连接钩子
	server.SetOnConnStart(func(conn ziface.IConnection) {
		fmt.Println("[WS] Client connected:", conn.RemoteAddr())
		// 发送欢迎消息
		conn.SendMsg(99, []byte("Welcome to Zinx WebSocket Server!"))
	})

	server.SetOnConnStop(func(conn ziface.IConnection) {
		fmt.Println("[WS] Client disconnected:", conn.RemoteAddr())
	})

	// 心跳检测
	server.StartHeartBeat(30 * time.Second)

	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║           Zinx WebSocket Server Starting...               ║
╠═══════════════════════════════════════════════════════════╣
║  TCP Port:  8999                                          ║
║  WS Port:   9000                                          ║
║  WS Path:   /ws                                           ║
║  Auth:      Token required                                ║
╚═══════════════════════════════════════════════════════════╝
`)

	server.Serve()
}
