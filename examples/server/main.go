// Zinx TCP Server 完整示例
// 功能：展示 Zinx 框架的基础服务器搭建、路由注册、连接管理等核心功能
package main

import (
	"fmt"
	"time"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

// ==================== 路由器定义 ====================

// PingRouter 处理 ping 消息 (MsgId=1)
type PingRouter struct {
	znet.BaseRouter
}

func (r *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("==> Recv from client : msgId=", request.GetMsgID(),
		", data=", string(request.GetData()))

	// 回复客户端
	conn := request.GetConnection()
	if err := conn.SendMsg(1, []byte("Pong from server")); err != nil {
		fmt.Println("Send pong error:", err)
	}
}

// HelloRouter 处理 hello 消息 (MsgId=2)
type HelloRouter struct {
	znet.BaseRouter
}

func (r *HelloRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("[PreHandle] Processing hello message...")
}

func (r *HelloRouter) Handle(request ziface.IRequest) {
	fmt.Println("==> HelloRouter Handle : msgId=", request.GetMsgID(),
		", data=", string(request.GetData()))

	conn := request.GetConnection()
	reply := fmt.Sprintf("Hello, I received your message: %s", string(request.GetData()))
	if err := conn.SendMsg(2, []byte(reply)); err != nil {
		fmt.Println("Send reply error:", err)
	}
}

func (r *HelloRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("[PostHandle] Hello message processed")
}

// ==================== 连接钩子函数 ====================

// OnClientConnect 连接建立时的回调
func OnClientConnect(conn ziface.IConnection) {
	fmt.Println("==> OnClientConnect - Client connected:",
		conn.RemoteAddr().String())

	// 设置连接属性
	conn.SetProperty("connectTime", time.Now())
	conn.SetProperty("clientAddr", conn.RemoteAddr().String())

	// 发送欢迎消息
	go func() {
		time.Sleep(100 * time.Millisecond)
		conn.SendMsg(99, []byte("Welcome to Zinx Server!"))
	}()
}

// OnClientDisconnect 连接断开时的回调
func OnClientDisconnect(conn ziface.IConnection) {
	fmt.Println("==> OnClientDisconnect - Client disconnected:",
		conn.RemoteAddr().String())

	// 获取连接属性
	if connectTime, err := conn.GetProperty("connectTime"); err == nil {
		if t, ok := connectTime.(time.Time); ok {
			duration := time.Since(t)
			fmt.Printf("Connection duration: %v\n", duration)
		}
	}
}

// ==================== 主函数 ====================

func main() {
	// 创建服务器
	server := znet.NewServer(func(s *znet.Server) {
		s.Name = "ZinxDemoServer"
		s.TCPPort = 8999
		s.MaxConn = 12000
		s.WorkerPoolSize = 10
	})

	// 注册路由
	server.AddRouter(1, &PingRouter{})
	server.AddRouter(2, &HelloRouter{})

	// 注册连接钩子
	server.SetOnConnStart(OnClientConnect)
	server.SetOnConnStop(OnClientDisconnect)

	// 启动心跳检测 (可选)
	server.StartHeartBeat(30 * time.Second)

	// 启动服务器
	fmt.Println(`
              ██                        
              ▀▀                        
 ████████   ████     ██▄████▄  ▀██  ██▀ 
     ▄█▀      ██     ██▀   ██    ████   
   ▄█▀        ██     ██    ██    ▄██▄   
 ▄██▄▄▄▄▄  ▄▄▄██▄▄▄  ██    ██   ▄█▀▀█▄  
 ▀▀▀▀▀▀▀▀  ▀▀▀▀▀▀▀▀  ▀▀    ▀▀  ▀▀▀  ▀▀▀ 
                                        
 Zinx Demo Server Starting...
`)

	server.Serve()
}
