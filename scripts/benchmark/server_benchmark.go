// Zinx Server 性能基准测试
// 功能：测试 Zinx 服务器的吞吐量和延迟
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

var (
	totalMessages  uint64
	totalBytes     uint64
	activeClients  int32
	latencies      []time.Duration
	latenciesMutex sync.Mutex
	testStartTime  time.Time
)

// EchoRouter 回声路由器 - 测试吞吐量
type EchoRouter struct {
	znet.BaseRouter
}

func (r *EchoRouter) Handle(request ziface.IRequest) {
	conn := request.GetConnection()
	data := request.GetData()

	// 记录接收时间
	receiveTime := time.Now()

	// 获取发送时间 (从消息中解析)
	if len(data) >= 8 {
		sendTime := time.Unix(0, int64(byteToUint64(data[:8])))
		latency := time.Since(sendTime)

		latenciesMutex.Lock()
		latencies = append(latencies, latency)
		latenciesMutex.Unlock()
	}

	// 回声响应
	conn.SendMsg(1, data)

	// 统计
	atomic.AddUint64(&totalMessages, 1)
	atomic.AddUint64(&totalBytes, uint64(len(data)))
}

// PingRouter 简单 ping 路由器
type PingRouter struct {
	znet.BaseRouter
}

func (r *PingRouter) Handle(request ziface.IRequest) {
	conn := request.GetConnection()
	conn.SendMsg(1, []byte("pong"))
	atomic.AddUint64(&totalMessages, 1)
}

// 辅助函数：byte 转 uint64
func byteToUint64(b []byte) uint64 {
	_ = b[7] // bounds check hint
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

// 统计报告
func printStats() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		messages := atomic.LoadUint64(&totalMessages)
		bytes := atomic.LoadUint64(&totalBytes)
		clients := atomic.LoadInt32(&activeClients)
		elapsed := time.Since(testStartTime).Seconds()

		fmt.Printf("\n[Stats] Time: %.0fs | Clients: %d | Messages: %d | Throughput: %.0f msg/s | Bandwidth: %.2f MB/s\n",
			elapsed, clients, messages,
			float64(messages)/elapsed,
			float64(bytes)/elapsed/1024/1024)
	}
}

// 延迟统计
func printLatencyStats() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		latenciesMutex.Lock()
		if len(latencies) > 0 {
			var sum time.Duration
			minLatency := latencies[0]
			maxLatency := latencies[0]

			for _, lat := range latencies {
				sum += lat
				if lat < minLatency {
					minLatency = lat
				}
				if lat > maxLatency {
					maxLatency = lat
				}
			}

			avgLatency := sum / time.Duration(len(latencies))

			// P99 延迟
			sorted := make([]time.Duration, len(latencies))
			copy(sorted, latencies)
			// 简单排序 (实际应使用 sort.Slice)
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i] > sorted[j] {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
			p99Idx := len(sorted) * 99 / 100
			p99Latency := sorted[p99Idx]

			fmt.Printf("[Latency] Samples: %d | Avg: %v | Min: %v | Max: %v | P99: %v\n",
				len(latencies), avgLatency, minLatency, maxLatency, p99Latency)

			// 重置 (避免内存增长)
			if len(latencies) > 10000 {
				latencies = latencies[len(latencies)-1000:]
			}
		}
		latenciesMutex.Unlock()
	}
}

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║           Zinx Server Benchmark                           ║
╠═══════════════════════════════════════════════════════════╣
║  This server is designed for performance testing:         ║
║  - EchoRouter: Echoes back all messages                   ║
║  - PingRouter: Simple ping/pong                           ║
║                                                           ║
║  Use with benchmark client to measure:                    ║
║  - Throughput (messages/second)                           ║
║  - Bandwidth (MB/second)                                  ║
║  - Latency (avg, min, max, p99)                           ║
╚═══════════════════════════════════════════════════════════╝
`)

	server := znet.NewServer(func(s *znet.Server) {
		s.Name = "ZinxBenchmarkServer"
		s.TCPPort = 8999
		s.MaxConn = 10000
		s.WorkerPoolSize = 50
		s.MaxWorkerTaskLen = 4096
		s.MaxPacketSize = 65536
		s.LogCons = true
	})

	// 注册路由
	server.AddRouter(1, &EchoRouter{})
	server.AddRouter(2, &PingRouter{})

	// 连接钩子
	server.SetOnConnStart(func(conn ziface.IConnection) {
		atomic.AddInt32(&activeClients, 1)
		fmt.Printf("[Connect] Client[%d] connected. Total: %d\n",
			conn.GetConnID(), atomic.LoadInt32(&activeClients))
	})

	server.SetOnConnStop(func(conn ziface.IConnection) {
		atomic.AddInt32(&activeClients, -1)
		fmt.Printf("[Disconnect] Client[%d] disconnected. Total: %d\n",
			conn.GetConnID(), atomic.LoadInt32(&activeClients))
	})

	// 启动统计
	testStartTime = time.Now()
	go printStats()
	go printLatencyStats()

	// 启动服务器
	server.Serve()
}
