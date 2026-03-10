// Zinx TCP Client 完整示例
// 功能：展示如何使用 Zinx 客户端连接服务器、发送消息、处理响应
package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

// 客户端业务逻辑 - 定时发送 Ping 消息
func pingLoop(conn ziface.IConnection) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-conn.Context().Done():
			fmt.Println("[Client] Connection closed, stopping ping loop")
			return
		case <-ticker.C:
			err := conn.SendMsg(1, []byte("Ping from client"))
			if err != nil {
				fmt.Println("[Client] Send ping error:", err)
				return
			}
			fmt.Println("[Client] Send: Ping")
		}
	}
}

// 心跳消息发送
func heartbeatLoop(conn ziface.IConnection) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-conn.Context().Done():
			fmt.Println("[Client] Connection closed, stopping heartbeat")
			return
		case <-ticker.C:
			err := conn.SendMsg(99, []byte("heartbeat"))
			if err != nil {
				fmt.Println("[Client] Send heartbeat error:", err)
				return
			}
		}
	}
}

// 连接创建时的回调
func onClientStart(conn ziface.IConnection) {
	fmt.Println("===========================================")
	fmt.Println("[Client] Connected to server successfully!")
	fmt.Printf("[Client] Local:  %s\n", conn.LocalAddr().String())
	fmt.Printf("[Client] Remote: %s\n", conn.RemoteAddr().String())
	fmt.Printf("[Client] ConnID: %d\n", conn.GetConnID())
	fmt.Println("===========================================")

	// 设置连接属性
	conn.SetProperty("loginTime", time.Now())
	conn.SetProperty("version", "1.0.0")

	// 启动业务协程
	go pingLoop(conn)
	go heartbeatLoop(conn)

	// 发送欢迎消息请求
	time.Sleep(500 * time.Millisecond)
	conn.SendMsg(2, []byte("Hello Server!"))
}

// 连接断开时的回调
func onClientStop(conn ziface.IConnection) {
	fmt.Println("===========================================")
	fmt.Println("[Client] Connection closed!")

	// 获取连接属性
	if loginTime, err := conn.GetProperty("loginTime"); err == nil {
		if t, ok := loginTime.(time.Time); ok {
			duration := time.Since(t)
			fmt.Printf("[Client] Connection duration: %v\n", duration)
		}
	}
	fmt.Println("===========================================")
}

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║              Zinx Client Starting...                      ║
╚═══════════════════════════════════════════════════════════╝
`)

	// 创建客户端
	client := znet.NewClient("127.0.0.1", 8999)

	// 设置连接钩子
	client.SetOnConnStart(onClientStart)
	client.SetOnConnStop(onClientStop)

	// 启动客户端
	client.Start()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("\n[Client] Shutting down...")
	client.Stop()
}
