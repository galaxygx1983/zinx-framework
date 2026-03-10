// Zinx 心跳检测完整示例
// 功能：展示多种心跳检测配置方式和自定义心跳处理
package main

import (
	"fmt"
	"time"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

// ==================== 业务路由器 ====================

type PingRouter struct {
	znet.BaseRouter
}

func (r *PingRouter) Handle(request ziface.IRequest) {
	conn := request.GetConnection()
	fmt.Printf("[Business] Ping from Client[%d]\n", conn.GetConnID())
	conn.SendMsg(1, []byte("Pong"))
}

// ==================== 心跳检测配置 ====================

// 方式 1: 基础心跳检测 (默认配置)
func startBasicHeartbeat(server *znet.Server) {
	// 每 30 秒检测一次，使用默认心跳消息
	server.StartHeartBeat(30 * time.Second)
	fmt.Println("[Heartbeat] Basic heartbeat started (30s interval)")
}

// 方式 2: 自定义心跳配置
func startCustomHeartbeat(server *znet.Server) {
	option := &ziface.HeartBeatOption{
		HeartBeatMax: 5, // 最大心跳次数，超过则断开连接

		// 自定义心跳消息
		MakeMsg: func(conn ziface.IConnection) []byte {
			return []byte("{\"type\":\"heartbeat\",\"ts\":" + fmt.Sprintf("%d", time.Now().Unix()) + "}")
		},

		// 远程不活跃时的处理
		OnRemoteNotAlive: func(conn ziface.IConnection) {
			fmt.Printf("[Heartbeat] Client[%d] not alive, disconnecting\n", conn.GetConnID())
			conn.Stop()
		},

		// 心跳超时回调
		OnHeartbeatTimeout: func(conn ziface.IConnection) {
			fmt.Printf("[Heartbeat] Client[%d] heartbeat timeout\n", conn.GetConnID())
		},
	}

	server.StartHeartBeatWithOption(10*time.Second, option)
	fmt.Println("[Heartbeat] Custom heartbeat started (10s interval, max 5 retries)")
}

// 方式 3: 业务心跳处理 (在业务层处理心跳)
type HeartbeatRouter struct {
	znet.BaseRouter
}

func (r *HeartbeatRouter) Handle(request ziface.IRequest) {
	conn := request.GetConnection()

	// 更新最后活跃时间
	conn.SetProperty("lastActive", time.Now())

	fmt.Printf("[Heartbeat] Received heartbeat from Client[%d]\n", conn.GetConnID())

	// 回复心跳响应
	conn.SendMsg(99, []byte("heartbeat-ack"))
}

// ==================== 连接管理 ====================

// 检查连接活跃度的定时器
func startActiveCheck(server *znet.Server) {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			connMgr := server.GetConnMgr()
			conns := connMgr.GetAllConnIDs()

			for _, connID := range conns {
				conn, err := connMgr.Get(connID)
				if err != nil {
					continue
				}

				// 检查最后活跃时间
				if lastActive, err := conn.GetProperty("lastActive"); err == nil {
					if t, ok := lastActive.(time.Time); ok {
						if time.Since(t) > 120*time.Second {
							fmt.Printf("[ActiveCheck] Client[%d] inactive for 2min, disconnecting\n", connID)
							conn.Stop()
						}
					}
				}
			}
		}
	}()
	fmt.Println("[ActiveCheck] Connection activity checker started")
}

// ==================== 主函数 ====================

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║            Zinx Heartbeat Demo Server                     ║
╠═══════════════════════════════════════════════════════════╣
║  Heartbeat Methods:                                       ║
║  1. Basic:     30s interval, default msg                  ║
║  2. Custom:    10s interval, custom JSON msg              ║
║  3. Business:  Handle in router (MsgId=99)                ║
║  4. Active:    Check last activity every 60s              ║
╚═══════════════════════════════════════════════════════════╝
`)

	server := znet.NewServer(func(s *znet.Server) {
		s.Name = "ZinxHeartbeatServer"
		s.TCPPort = 8999
		s.MaxConn = 12000
	})

	// 注册业务路由
	server.AddRouter(1, &PingRouter{})
	server.AddRouter(99, &HeartbeatRouter{}) // 心跳消息处理

	// 连接钩子
	server.SetOnConnStart(func(conn ziface.IConnection) {
		fmt.Printf("[Connect] Client[%d] connected\n", conn.GetConnID())
		// 初始化最后活跃时间
		conn.SetProperty("lastActive", time.Now())
		conn.SetProperty("connectTime", time.Now())
	})

	server.SetOnConnStop(func(conn ziface.IConnection) {
		fmt.Printf("[Disconnect] Client[%d] disconnected\n", conn.GetConnID())
	})

	// 启动心跳检测 (取消注释选择需要的方式)
	// startBasicHeartbeat(server)      // 方式 1
	startCustomHeartbeat(server) // 方式 2

	// 启动业务活跃度检查
	startActiveCheck(server) // 方式 4

	// 启动服务器
	server.Serve()
}
