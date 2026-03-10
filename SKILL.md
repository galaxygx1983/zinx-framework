---
name: zinx-framework
description: Golang 轻量级并发服务器框架 Zinx 开发指南。用于构建 TCP 服务器、游戏服务器、长连接服务。TRIGGER 当用户需要：(1) 开发 Golang TCP 服务器 (2) 实现游戏服务器框架 (3) 处理长连接业务 (4) 学习 Zinx 框架使用。
---

# Zinx Framework

> Golang 轻量级并发服务器框架 - 基于 Zinx v1.2.7
> 复杂度：medium | 语言：Go | 许可证：MIT

## 概述

Zinx 是一个基于 Golang 的轻量级 TCP 并发服务器框架，适用于游戏服务器、长连接服务、消息转发等领域。框架采用渐进式迭代开发，当前最新版本为 **v1.2.7**（2025 年 6 月）。

**源码**: https://github.com/aceld/zinx  
**文档**: https://github.com/aceld/zinx/wiki | https://www.yuque.com/aceld/tsgooa  
**Go 版本要求**: 1.23+

### 核心特性

- 轻量级 TCP 服务器框架
- 支持 TCP/WebSocket/KCP 多种协议
- Worker 池模式处理并发
- 消息路由机制
- 心跳检测
- 连接属性管理
- 读写分离模型
- 支持 Protobuf 协议

---

## Workflow 1: 快速开始

**触发条件**: 当需要快速创建一个基础的 TCP 服务器时

### 前置条件

- Go 1.23+
- 安装依赖：`go get github.com/aceld/zinx`

### 步骤

#### Step 1: 创建服务器

```go
package main

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

// PingRouter 处理 MsgId=1 的路由
type PingRouter struct {
	znet.BaseRouter
}

// Handle 处理客户端消息
func (r *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))
}

func main() {
	// 1. 创建服务器
	s := znet.NewServer()

	// 2. 注册路由
	s.AddRouter(1, &PingRouter{})

	// 3. 启动服务
	s.Serve()
}
```

#### Step 2: 创建客户端

```go
package main

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"time"
)

// 客户端业务逻辑
func pingLoop(conn ziface.IConnection) {
	for {
		err := conn.SendMsg(1, []byte("Ping...Ping...Ping...[FromClient]"))
		if err != nil {
			fmt.Println(err)
			break
		}
		time.Sleep(1 * time.Second)
	}
}

// 连接创建时的钩子函数
func onClientStart(conn ziface.IConnection) {
	fmt.Println("onClientStart is Called ... ")
	go pingLoop(conn)
}

func main() {
	// 创建客户端
	client := znet.NewClient("127.0.0.1", 8999)

	// 设置连接成功后的钩子函数
	client.SetOnConnStart(onClientStart)

	// 启动客户端
	client.Start()

	// 防止进程退出
	select {}
}
```

#### Step 3: 配置文件 (可选)

创建 `conf/zinx.json`:

```json
{
  "Name": "zinx v1.2 demoApp",
  "Host": "0.0.0.0",
  "TCPPort": 8999,
  "MaxConn": 12000,
  "WorkerPoolSize": 10,
  "MaxPacketSize": 4096,
  "LogDir": "./log",
  "LogFile": "server.log",
  "LogSaveDays": 15,
  "LogCons": true,
  "LogIsolationLevel": 0,
  "HeartbeatMax": 10
}
```

**验证方法**:
```bash
# 启动服务器
go run server.go

# 启动客户端
go run client.go
```

---

## Workflow 2: 核心接口与实现

### IServer 接口

服务器核心接口定义：

```go
type IServer interface {
    // 基础方法
    Start()
    Stop()
    Serve()

    // 路由注册
    AddRouter(msgID uint32, router IRouter)
    AddRouterSlices(msgID uint32, router ...RouterHandler) IRouterSlices
    Group(start, end uint32, Handlers ...RouterHandler) IGroupRouterSlices
    Use(Handlers ...RouterHandler) IRouterSlices

    // 连接管理
    GetConnMgr() IConnManager
    SetOnConnStart(func(IConnection))
    SetOnConnStop(func(IConnection))
    GetOnConnStart() func(IConnection)
    GetOnConnStop() func(IConnection)

    // 协议与处理
    GetPacket() IDataPack
    GetMsgHandler() IMsgHandle
    SetPacket(IDataPack)

    // 心跳检测
    StartHeartBeat(time.Duration)
    StartHeartBeatWithOption(time.Duration, *HeartBeatOption)
    GetHeartBeat() IHeartbeatChecker

    // 其他
    SetWebsocketAuth(func(r *http.Request) error)
    ServerName() string
}
```

### IConnection 接口

连接接口定义：

```go
type IConnection interface {
    // 生命周期
    Start()
    Stop()
    Context() context.Context

    // 连接信息
    GetName() string
    GetConnection() net.Conn
    GetWsConn() *websocket.Conn
    GetConnID() uint64
    GetConnIdStr() string
    GetMsgHandler() IMsgHandle
    GetWorkerID() uint32
    RemoteAddr() net.Addr
    LocalAddr() net.Addr

    // 发送数据
    Send(data []byte) error
    SendToQueue(data []byte, opts ...MsgSendOption) error
    SendMsg(msgID uint32, data []byte) error
    SendBuffMsg(msgID uint32, data []byte, opts ...MsgSendOption) error

    // 属性管理
    SetProperty(key string, value interface{})
    GetProperty(key string) (interface{}, error)
    RemoveProperty(key string)

    // 心跳与回调
    IsAlive() bool
    SetHeartBeat(checker IHeartbeatChecker)
    AddCloseCallback(handler, key interface{}, callback func())
    RemoveCloseCallback(handler, key interface{})
    InvokeCloseCallbacks()
}
```

### IRouter 接口

路由接口定义：

```go
type IRouter interface {
    PreHandle(request IRequest)  // 处理前钩子
    Handle(request IRequest)     // 业务处理
    PostHandle(request IRequest) // 处理后钩子
}

// 新版路由处理器
type RouterHandler func(request IRequest)

type IRouterSlices interface {
    Use(Handlers ...RouterHandler)
    AddHandler(msgId uint32, handlers ...RouterHandler)
    Group(start, end uint32, Handlers ...RouterHandler) IGroupRouterSlices
    GetHandlers(MsgId uint32) ([]RouterHandler, bool)
}
```

### IRequest 接口

请求接口定义：

```go
type IRequest interface {
    GetConnection() IConnection  // 获取连接
    GetMsgID() uint32            // 获取消息 ID
    GetData() []byte             // 获取消息数据
    GetDataLen() uint32          // 获取数据长度
}
```

---

## Workflow 3: 消息封包与拆包

### 消息结构

```go
type Message struct {
    ID      uint32
    DataLen uint32
    Data    []byte
}

type IDataPack interface {
    GetHeadLen() uint32
    Pack(msg *Message) ([]byte, error)
    Unpack([]byte) (*Message, error)
}
```

### 封包格式

```
| ID(4 字节) | DataLen(4 字节) | Data(N 字节) |
```

### 自定义封包器

```go
type DataPack struct {
    packHeadLen uint32
}

func NewDataPack() *DataPack {
    return &DataPack{packHeadLen: 8}
}

func (dp *DataPack) Pack(msg *ziface.Message) ([]byte, error) {
    dataBuff := bytes.NewBuffer([]byte{})
    
    // 写入消息 ID
    binary.Write(dataBuff, binary.LittleEndian, msg.ID)
    // 写入数据长度
    binary.Write(dataBuff, binary.LittleEndian, msg.DataLen)
    // 写入数据
    binary.Write(dataBuff, binary.LittleEndian, msg.Data)
    
    return dataBuff.Bytes(), nil
}

func (dp *DataPack) Unpack(binaryData []byte) (*ziface.Message, error) {
    dataBuff := bytes.NewReader(binaryData)
    msg := &ziface.Message{}
    
    // 读取消息 ID
    binary.Read(dataBuff, binary.LittleEndian, &msg.ID)
    // 读取数据长度
    binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen)
    // 读取数据
    msg.Data = make([]byte, msg.DataLen)
    binary.Read(dataBuff, binary.LittleEndian, &msg.Data)
    
    return msg, nil
}
```

---

## Workflow 4: Worker 池与消息处理

### Worker 池模式

```go
type MsgHandler struct {
    ApiRouter    map[uint32]ziface.IRouter
    Workers      []*Worker
    TaskQueue    []chan *ziface.Request
    WorkerPoolSize uint32
    MaxWorkerTaskLen uint32
}

type Worker struct {
    ID   uint32
    Task chan *ziface.Request
    Pool *MsgHandler
}

func (w *Worker) Start() {
    go func() {
        for {
            select {
            case req := <-w.Task:
                w.Pool.ExecuteMethod(req)
            case <-w.Pool.Quit:
                return
            }
        }
    }()
}
```

### 消息处理模式

```go
// 方式 1: 绑定 Worker 模式 (每个连接绑定固定 Worker)
s.GetMsgHandler().SetBindWorker(true)

// 方式 2: DynamicBind 模式 (类似 Bind 但不闲置 Worker)
s.GetMsgHandler().SetDynamicBind(true)

// 方式 3: 默认轮询模式
```

---

## Workflow 5: 心跳检测

### 基础心跳检测

```go
// 启动默认心跳检测
s.StartHeartBeat(10 * time.Second)
```

### 自定义心跳检测

```go
option := &ziface.HeartBeatOption{
    HeartBeatMax: 10,  // 最大心跳次数
    HeartBeatFunc: func(conn ziface.IConnection) {
        // 自定义心跳处理逻辑
        err := conn.SendMsg(9999, []byte("heartbeat"))
        if err != nil {
            conn.Stop()
        }
    },
    OnRemoteNotAlive: func(conn ziface.IConnection) {
        // 远程不活跃时的处理
        conn.Stop()
    },
}

s.StartHeartBeatWithOption(5*time.Second, option)
```

---

## Workflow 6: WebSocket 支持

### 启用 WebSocket

```go
// 创建支持 WebSocket 的服务器
s := znet.NewServer(func(s *znet.Server) {
    s.SetWebsocketAuth(func(r *http.Request) error {
        // 自定义 WebSocket 认证逻辑
        token := r.URL.Query().Get("token")
        if token != "valid-token" {
            return errors.New("invalid token")
        }
        return nil
    })
})
```

### WebSocket 路径配置

```go
// 自定义 WebSocket 路径
s.SetWebsocketPath("/ws")
```

---

## Workflow 7: 连接属性管理

### 设置和获取属性

```go
func (r *PingRouter) Handle(request ziface.IRequest) {
    conn := request.GetConnection()
    
    // 设置属性
    conn.SetProperty("userId", 12345)
    conn.SetProperty("username", "player1")
    
    // 获取属性
    userId, _ := conn.GetProperty("userId")
    username, _ := conn.GetProperty("username")
    
    // 删除属性
    conn.RemoveProperty("tempKey")
}
```

### 连接回调

```go
// 添加关闭回调
conn.AddCloseCallback("handler1", "key1", func() {
    fmt.Println("Connection closed, cleanup...")
})

// 移除回调
conn.RemoveCloseCallback("handler1", "key1")
```

---

## Workflow 8: 路由组与中间件

### 路由组

```go
// 创建路由组
group := s.Group(100, 200)

// 添加组内处理器
group.AddHandler(101, func(request ziface.IRequest) {
    // 处理逻辑
})

group.AddHandler(102, func(request ziface.IRequest) {
    // 处理逻辑
})
```

### 中间件

```go
// 全局中间件
s.Use(func(request ziface.IRequest) {
    fmt.Println("Global middleware")
})

// 组中间件
group.Use(func(request ziface.IRequest) {
    fmt.Println("Group middleware")
})
```

---

## Workflow 9: KCP 协议支持

### 启用 KCP

```go
// 配置文件添加 KCP 相关配置
{
  "KcpPort": 9090,
  "KcpConfig": {
    "Conv": 123456,
    "Mtu": 1350,
    "Sndwnd": 1024,
    "Rcvwnd": 1024,
    "Mode": 0
  }
}
```

---

## 配置项参考

### 完整配置示例

```json
{
  "Name": "zinxServer",
  "Host": "0.0.0.0",
  "TCPPort": 8999,
  "KcpPort": 9090,
  "MaxConn": 12000,
  "WorkerPoolSize": 10,
  "MaxWorkerTaskLen": 1024,
  "MaxMsgChanLen": 1024,
  "MaxPacketSize": 4096,
  "LogDir": "./log",
  "LogFile": "server.log",
  "LogSaveDays": 15,
  "LogCons": true,
  "LogIsolationLevel": 0,
  "HeartbeatMax": 10,
  "WsPort": 8080,
  "WsPath": "/ws"
}
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Name | 服务器名称 | ZinxServerApp |
| Host | 服务器 IP | 0.0.0.0 |
| TCPPort | TCP 监听端口 | 8999 |
| KcpPort | KCP 监听端口 | - |
| MaxConn | 最大连接数 | 12000 |
| WorkerPoolSize | Worker 池大小 | 10 |
| MaxWorkerTaskLen | Worker 任务队列长度 | 1024 |
| MaxMsgChanLen | 消息队列长度 | 1024 |
| MaxPacketSize | 最大包大小 | 4096 |
| LogDir | 日志目录 | ./log |
| LogFile | 日志文件名 | - |
| LogSaveDays | 日志保留天数 | 15 |
| LogCons | 是否输出到控制台 | true |
| LogIsolationLevel | 日志隔离级别 (0-3) | 0 |
| HeartbeatMax | 心跳最大次数 | 10 |
| WsPort | WebSocket 端口 | 8080 |
| WsPath | WebSocket 路径 | /ws |

---

## 模块结构

```
zinx/
├── ziface/              # 接口抽象层
│   ├── iserver.go       # 服务器接口
│   ├── iconnection.go   # 连接接口
│   ├── irouter.go       # 路由接口
│   ├── imessage.go      # 消息接口
│   ├── idatapack.go     # 封包接口
│   ├── imsghandle.go    # 消息处理接口
│   └── iheartbeat.go    # 心跳接口
├── znet/                # 网络实现层
│   ├── server.go        # 服务器实现
│   ├── client.go        # 客户端实现
│   ├── connection.go    # 连接实现
│   ├── router.go        # 路由实现
│   └── msghandler.go    # 消息处理实现
├── zpack/               # 封包拆包模块
├── zconf/               # 配置管理模块
├── zlog/                # 日志模块
├── ztimer/              # 定时器模块
├── zdecoder/            # 解码器模块
├── zinterceptor/        # 拦截器模块
├── znotify/             # 通知模块
├── zasync_op/           # 异步操作模块
├── zutils/              # 工具模块
└── examples/            # 示例代码
```

---

## v1.2.x 新特性

### v1.2.7 更新 (2025.06)

- 修复并发情况下 AsyncOpResult 可能导致 callback 调用丢失的问题
- 新增 SendBuffMsg API 支持超时
- 支持 WebSocket 自定义 Header
- 支持 WebSocket 路径配置
- 修复 Client 中共享变量 conn 的可见性问题
- 优化缓冲发送逻辑，减少系统调用

### v1.2.5 更新

- **许可证变更**: 从 GPL 3.0 改为 MIT（企业可闭源使用）
- 新增 RequestPool 模块
- SendBuffMsg 更名为 SendToQueue
- 新增 KCP 连接配置选项
- 新增 addclosecallback() 支持多个回调
- 更新 HeartBeat 路由注册方式
- Request 对象改为 sync pool 管理

### v1.2.x 核心改进

1. **DynamicBind 模式**: 类似 Bind 模式但不闲置 Worker
2. **Request 池化**: 使用 sync.Pool 管理 Request 对象
3. **WebSocket 增强**: 支持鉴权、自定义路径和 Header
4. **性能优化**: 优化缓冲发送逻辑
5. **并发安全**: 多处并发问题修复

---

## MMO 游戏应用案例

### AOI (Area of Interest) 算法

用于处理玩家视野范围内的其他玩家同步：

```
┌────┬────┬────┐
│ 0,0│ 1,0│ 2,0│
├────┼────┼────┤
│ 0,1│ 1,1│ 2,1│
├────┼────┼────┤
│ 0,2│ 1,2│ 2,2│
└────┴────┴────┘
```

九宫格算法：获取周围 9 个格子的所有玩家

### Protobuf 协议

```protobuf
syntax = "proto3";

message Player {
    int32 pid = 1;
    float x = 2;
    float y = 3;
    float z = 4;
}

message Talk {
    string content = 1;
}
```

---

## 常见问题

### Q1: 如何处理粘包/拆包问题？

Zinx 默认使用 TLV 格式 (Type-Length-Value)，通过自定义 `IDataPack` 实现：

```go
// 设置自定义封包器
s.SetPacket(zpack.NewDataPack())
```

### Q2: 如何实现连接认证？

使用 `SetOnConnStart` 钩子函数：

```go
s.SetOnConnStart(func(conn ziface.IConnection) {
    // 认证逻辑
    token := getConnectionToken(conn)
    if !validateToken(token) {
        conn.Stop()
    }
})
```

### Q3: 如何优雅关闭服务器？

```go
// 监听系统信号
signalChan := make(chan os.Signal, 1)
signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-signalChan
    s.Stop()
    os.Exit(0)
}()

s.Serve()
```

---

## 参考资料

- **源码**: https://github.com/aceld/zinx
- **Wiki**: https://github.com/aceld/zinx/wiki
- **中文文档**: https://www.yuque.com/aceld/tsgooa
- **教程视频**: Bilibili/抖音/YouTube 搜索 "zinx 框架"
- **社区**: Discord | 微信 (ace_ld) | QQ 群
- **作者**: Aceld(刘丹冰) - danbing.at@gmail.com

---

## 版本历史

| 版本 | 功能特性 |
|------|----------|
| V0.1 | 基础 Server |
| V0.2 | 连接封装与业务绑定 |
| V0.3 | 路由功能 |
| V0.4 | 全局配置 |
| V0.5 | 消息封装 |
| V0.6 | 多路由模式 |
| V0.7 | 读写分离模型 |
| V0.8 | 消息队列及多任务 |
| V0.9 | 连接管理 |
| V0.10 | 连接属性设置 |
| V1.0 | 完整版 (WebSocket 支持) |
| V1.2.x | MIT 许可、Request 池、DynamicBind、KCP |

---

Updated based on Zinx v1.2.7 (June 2025)
