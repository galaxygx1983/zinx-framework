# Zinx Framework Skill

> Zinx 是一个基于 Golang 的轻量级 TCP 并发服务器框架

[![License: MIT](https://img.shields.io/badge/License-MIT-black.svg)](https://github.com/aceld/zinx/blob/master/LICENSE)
[![Zinx Version](https://img.shields.io/badge/zinx-v1.2.7-blue.svg)](https://github.com/aceld/zinx)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)

## 概述

本技能仓库包含 Zinx 框架的完整开发指南和示例代码，帮助开发者快速上手 Zinx 框架进行 TCP/WebSocket/KCP服务器开发。

## 目录结构

```
zinx-framework/
├── SKILL.md                    # 核心技能文档
├── README.md                   # 本文件
├── docs/                       # 参考文档
│   └── REFERENCE.md            # 完整参考文档
├── examples/                   # 示例代码
│   ├── server/                 # TCP 服务器示例
│   ├── client/                 # TCP 客户端示例
│   ├── websocket/              # WebSocket 示例
│   ├── middleware/             # 中间件示例
│   ├── heartbeat/              # 心跳检测示例
│   └── chatroom/               # 聊天室示例
├── scripts/                    # 工具脚本
│   ├── docker/                 # Docker 部署
│   └── benchmark/              # 性能测试
└── conf/                       # 配置文件模板
```

## 快速开始

### 1. 环境要求

- Go 1.23+
- Git

### 2. 安装依赖

```bash
go get github.com/aceld/zinx
```

### 3. 运行示例

```bash
# 启动服务器
cd examples/server
go mod tidy
go run main.go

# 启动客户端 (新终端)
cd examples/client
go mod tidy
go run main.go
```

## 示例说明

### TCP 服务器 (`examples/server/`)

基础的 TCP 服务器示例，包含：
- 路由注册
- 连接钩子函数
- 心跳检测
- 连接属性管理

### WebSocket 服务器 (`examples/websocket/`)

WebSocket 服务器示例，包含：
- WebSocket 认证
- 自定义路径配置
- 聊天消息处理

### 中间件 (`examples/middleware/`)

RouterSlices 中间件模式示例：
- 日志中间件
- 恢复中间件
- 认证中间件
- 限流中间件
- 路由组

### 心跳检测 (`examples/heartbeat/`)

心跳检测配置示例：
- 基础心跳
- 自定义心跳
- 业务层心跳处理

## Docker 部署

```bash
cd scripts/docker
docker-compose up -d
```

服务端口：
- TCP: 8999
- WebSocket: 9000
- Prometheus: 9090
- Grafana: 3000

## 性能测试

### Go 基准测试

```bash
cd scripts/benchmark
go mod tidy
go run server_benchmark.go
```

### Python 压力测试

```bash
pip install aiosockets
python stress_test.py -c 100 -d 60
```

## 配置项

完整配置参考 `conf/zinx.json`：

```json
{
  "Name": "zinxServer",
  "Host": "0.0.0.0",
  "TCPPort": 8999,
  "WsPort": 9000,
  "MaxConn": 12000,
  "WorkerPoolSize": 10,
  "MaxPacketSize": 4096,
  "LogDir": "./log",
  "HeartbeatMax": 10
}
```

## 核心概念

### 路由器 (Router)

```go
type PingRouter struct {
    znet.BaseRouter
}

func (r *PingRouter) Handle(request ziface.IRequest) {
    conn := request.GetConnection()
    conn.SendMsg(1, []byte("Pong"))
}
```

### 中间件 (Middleware)

```go
func LoggingMiddleware(request ziface.IRequest) {
    fmt.Printf("MsgId: %d, DataLen: %d\n",
        request.GetMsgID(), request.GetDataLen())
}

server.Use(LoggingMiddleware)
```

### 连接钩子

```go
server.SetOnConnStart(func(conn ziface.IConnection) {
    fmt.Println("Client connected:", conn.GetConnID())
})
```

## 学习资源

| 资源 | 链接 |
|------|------|
| GitHub | https://github.com/aceld/zinx |
| 官方文档 | https://www.yuque.com/aceld/tsgooa |
| Wiki | https://github.com/aceld/zinx/wiki |
| 视频教程 | https://www.bilibili.com/video/av71067087 |

## 常见问题

**Q: 如何处理粘包/拆包？**

A: Zinx 默认使用 TLV 格式，可通过 `SetPacket()` 自定义封包器。

**Q: Worker 模式如何选择？**

A: 
- Hash (默认): 通用场景
- Bind: 长连接
- DynamicBind: 平衡资源利用

**Q: 如何实现连接认证？**

A: 使用 `SetOnConnStart` 钩子函数进行认证。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 致谢

- Zinx 作者：[Aceld(刘丹冰)](https://github.com/aceld)
- 所有贡献者：https://github.com/aceld/zinx/graphs/contributors

---

最后更新：2026-03-10
