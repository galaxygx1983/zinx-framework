// Zinx RouterSlices 中间件示例
// 功能：展示新版 RouterSlices 模式，支持中间件链式调用
package main

import (
	"fmt"
	"time"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

// ==================== 中间件定义 ====================

// LoggingMiddleware 日志中间件
func LoggingMiddleware(request ziface.IRequest) {
	conn := request.GetConnection()
	startTime := time.Now()

	fmt.Printf("[Middleware] [%s] Client[%d] MsgId[%d] DataLen[%d]\n",
		startTime.Format("15:04:05"),
		conn.GetConnID(),
		request.GetMsgID(),
		request.GetDataLen())

	// 将开始时间存入 request 上下文
	request.Set("startTime", startTime)

	// 继续执行下一个处理器
}

// RecoveryMiddleware 恢复中间件 - 捕获 panic
func RecoveryMiddleware(request ziface.IRequest) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("[Middleware] [PANIC RECOVERED] %v\n", err)
			conn := request.GetConnection()
			conn.SendMsg(500, []byte("Internal Server Error"))
		}
	}()
}

// AuthMiddleware 认证中间件
func AuthMiddleware(request ziface.IRequest) {
	conn := request.GetConnection()

	// 检查是否已认证
	if authenticated, err := conn.GetProperty("authenticated"); err == nil {
		if auth, ok := authenticated.(bool); ok && auth {
			return // 已认证，继续
		}
	}

	// 某些消息不需要认证
	if request.GetMsgID() == 1 { // Ping 消息
		return
	}

	// 未认证，拒绝请求
	fmt.Printf("[Middleware] Client[%d] not authenticated, rejecting MsgId[%d]\n",
		conn.GetConnID(), request.GetMsgID())
	conn.SendMsg(401, []byte("Unauthorized"))
	request.Abort() // 中止后续处理器执行
}

// RateLimitMiddleware 简单的限流中间件
func RateLimitMiddleware(request ziface.IRequest) {
	conn := request.GetConnection()

	// 获取请求计数
	var count int
	if c, err := conn.GetProperty("requestCount"); err == nil {
		count = c.(int)
	}

	count++
	conn.SetProperty("requestCount", count)

	// 限制每秒 10 个请求
	if count > 10 {
		fmt.Printf("[Middleware] Client[%d] rate limited\n", conn.GetConnID())
		conn.SendMsg(429, []byte("Too Many Requests"))
		request.Abort()
		return
	}

	// 重置计数 (简化示例，实际应使用时间窗口)
	if count%100 == 0 {
		conn.SetProperty("requestCount", 0)
	}
}

// ==================== 业务处理器 ====================

// PingHandler 处理 Ping 消息
func PingHandler(request ziface.IRequest) {
	conn := request.GetConnection()
	fmt.Printf("[Handler] Ping from Client[%d]\n", conn.GetConnID())
	conn.SendMsg(1, []byte("Pong"))
}

// HelloHandler 处理 Hello 消息
func HelloHandler(request ziface.IRequest) {
	conn := request.GetConnection()
	data := string(request.GetData())
	fmt.Printf("[Handler] Hello from Client[%d]: %s\n", conn.GetConnID(), data)

	reply := fmt.Sprintf("Hello, %s!", data)
	conn.SendMsg(2, []byte(reply))
}

// ChatHandler 处理聊天消息
func ChatHandler(request ziface.IRequest) {
	conn := request.GetConnection()
	data := string(request.GetData())
	fmt.Printf("[Handler] Chat from Client[%d]: %s\n", conn.GetConnID(), data)

	// 广播聊天消息 (简化示例)
	reply := fmt.Sprintf("[Broadcast] Client[%d]: %s", conn.GetConnID(), data)
	conn.SendMsg(3, []byte(reply))
}

// AuthHandler 处理认证请求
func AuthHandler(request ziface.IRequest) {
	conn := request.GetConnection()
	token := string(request.GetData())

	fmt.Printf("[Handler] Auth request from Client[%d], token: %s\n",
		conn.GetConnID(), token)

	if token == "valid-token" {
		conn.SetProperty("authenticated", true)
		conn.SetProperty("userId", conn.GetConnID())
		conn.SetProperty("authTime", time.Now())
		conn.SendMsg(100, []byte("Authentication successful"))
		fmt.Printf("[Handler] Client[%d] authenticated\n", conn.GetConnID())
	} else {
		conn.SendMsg(100, []byte("Authentication failed"))
	}
}

// ==================== 路由组 ====================

// 创建聊天相关的路由组
func setupChatGroup(server *znet.Server) {
	// 创建路由组 (MsgId 10-19)
	chatGroup := server.Group(10, 19, LoggingMiddleware, AuthMiddleware)

	// 添加组内路由
	chatGroup.AddHandler(10, func(request ziface.IRequest) {
		conn := request.GetConnection()
		conn.SendMsg(10, []byte("Chat room info"))
	})

	chatGroup.AddHandler(11, func(request ziface.IRequest) {
		conn := request.GetConnection()
		conn.SendMsg(11, []byte("User list"))
	})

	chatGroup.AddHandler(12, ChatHandler)
}

// ==================== 主函数 ====================

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║        Zinx RouterSlices Middleware Server                ║
╚═══════════════════════════════════════════════════════════╝
`)

	// 创建启用 RouterSlices 模式的服务器
	server := znet.NewDefaultRouterSlicesServer()

	// 配置服务器
	server.Name = "ZinxMiddlewareServer"
	server.TCPPort = 8999

	// 注册全局中间件 (对所有路由生效)
	server.Use(LoggingMiddleware, RecoveryMiddleware)

	// 注册路由和处理器
	server.AddRouterSlices(1, PingHandler)
	server.AddRouterSlices(2, AuthMiddleware, HelloHandler)
	server.AddRouterSlices(3, AuthMiddleware, RateLimitMiddleware, ChatHandler)
	server.AddRouterSlices(100, AuthHandler)

	// 设置路由组
	setupChatGroup(server)

	// 连接钩子
	server.SetOnConnStart(func(conn ziface.IConnection) {
		fmt.Println("[Server] Client connected:", conn.GetConnID())
		conn.SetProperty("requestCount", 0)
	})

	server.SetOnConnStop(func(conn ziface.IConnection) {
		fmt.Println("[Server] Client disconnected:", conn.GetConnID())
	})

	// 启动服务器
	server.Serve()
}

// ==================== 上下文使用示例 ====================

// 在处理器中使用上下文
func handlerWithContext(request ziface.IRequest) {
	// 从中间件获取数据
	if startTime, ok := request.Get("startTime"); ok {
		if t, ok := startTime.(time.Time); ok {
			latency := time.Since(t)
			fmt.Printf("Request latency: %v\n", latency)
		}
	}

	// 获取连接上下文用于优雅退出
	ctx := request.GetConnection().Context()

	// 在协程中使用
	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("Connection closed, cleaning up...")
			return
		case <-time.After(5 * time.Second):
			fmt.Println("Task completed")
		}
	}()
}
