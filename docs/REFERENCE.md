# Zinx Framework 参考文档

## 官方资源

### 核心链接

| 资源 | 链接 | 说明 |
|------|------|------|
| GitHub | https://github.com/aceld/zinx | 官方源码仓库 |
| Wiki (EN) | https://github.com/aceld/zinx/wiki | 英文维基文档 |
| 语雀文档 | https://www.yuque.com/aceld/tsgooa | 中文官方文档 |
| 官方网站 | http://zinx.me | 框架官网 |
| 作者博客 | https://www.yuque.com/aceld | 作者 Aceld 博客 |

### 社区渠道

| 平台 | 链接/方式 | 说明 |
|------|----------|------|
| Discord | https://discord.gg/xQ8Xxfyfcz | 官方 Discord 社区 |
| 微信 | 添加 `ace_ld` 备注 `zinx` | 官方微信群 |
| QQ 群 | 扫描官网 QQ 群二维码 | 官方 QQ 交流群 |
| Gitter | https://gitter.im/zinx_go/community | Gitter 聊天室 |

---

## 教程资源

### 视频教程

| 平台 | 链接 | 说明 |
|------|------|------|
| Bilibili | https://www.bilibili.com/video/av71067087 | 官方系列教程 |
| YouTube | https://www.youtube.com/watch?v=U95iF-HMWsU | YouTube 教程列表 |
| 抖音 | 搜索 "zinx 框架" | 抖音短视频教程 |

### 图文教程

| 教程 | 链接 | 说明 |
|------|------|------|
| DEV Community | https://dev.to/aceld | 英文教程系列 |
| 语雀专栏 | https://www.yuque.com/aceld | 中文原创教程 |

---

## 学习路线

### 入门阶段

1. **基础概念**
   - [ ] 了解 TCP 服务器基本原理
   - [ ] 学习 Zinx 架构设计
   - [ ] 完成 QuickStart 示例

2. **核心组件**
   - [ ] IServer 接口与实现
   - [ ] IConnection 连接管理
   - [ ] IRouter 路由机制
   - [ ] IDataPack 消息封包

3. **实践练习**
   - [ ] 搭建 TCP Echo 服务器
   - [ ] 实现简单的客户端
   - [ ] 配置服务器参数

### 进阶阶段

1. **高级特性**
   - [ ] RouterSlices 中间件模式
   - [ ] Worker 池模式配置
   - [ ] 心跳检测机制
   - [ ] 连接属性管理

2. **协议支持**
   - [ ] WebSocket 集成
   - [ ] KCP 协议配置
   - [ ] Protobuf 序列化

3. **性能优化**
   - [ ] RequestPool 对象池
   - [ ] DynamicBind 模式
   - [ ] 缓冲区优化

### 实战阶段

1. **项目实践**
   - [ ] 聊天室服务器
   - [ ] 游戏服务器框架
   - [ ] 消息推送服务

2. **生产部署**
   - [ ] Docker 容器化
   - [ ] 监控与日志
   - [ ] 性能基准测试

---

## 示例代码索引

### 基础示例

| 示例 | 路径 | 说明 |
|------|------|------|
| TCP 服务器 | `examples/server/` | 基础 TCP 服务器 |
| TCP 客户端 | `examples/client/` | 配套客户端 |
| WebSocket | `examples/websocket/` | WebSocket 服务器 |
| 中间件 | `examples/middleware/` | RouterSlices 中间件 |
| 心跳检测 | `examples/heartbeat/` | 心跳配置示例 |

### 部署脚本

| 脚本 | 路径 | 说明 |
|------|------|------|
| Docker | `scripts/docker/` | Docker 部署配置 |
| 压测 | `scripts/benchmark/` | 性能测试工具 |

---

## API 参考

### 核心接口

#### IServer
```go
type IServer interface {
    Start()
    Stop()
    Serve()
    AddRouter(msgID uint32, router IRouter)
    AddRouterSlices(msgID uint32, router ...RouterHandler) IRouterSlices
    Group(start, end uint32, Handlers ...RouterHandler) IGroupRouterSlices
    Use(Handlers ...RouterHandler) IRouterSlices
    GetConnMgr() IConnManager
    SetOnConnStart(func(IConnection))
    SetOnConnStop(func(IConnection))
    GetPacket() IDataPack
    SetPacket(IDataPack)
    GetMsgHandler() IMsgHandle
    StartHeartBeat(time.Duration)
    StartHeartBeatWithOption(time.Duration, *HeartBeatOption)
    SetWebsocketAuth(func(r *http.Request) error)
}
```

#### IConnection
```go
type IConnection interface {
    Start()
    Stop()
    Context() context.Context
    GetConnection() net.Conn
    GetConnID() uint64
    RemoteAddr() net.Addr
    Send(data []byte) error
    SendMsg(msgID uint32, data []byte) error
    SendBuffMsg(msgID uint32, data []byte, opts ...MsgSendOption) error
    SetProperty(key string, value interface{})
    GetProperty(key string) (interface{}, error)
    IsAlive() bool
    AddCloseCallback(handler, key interface{}, callback func())
}
```

### 配置项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| Name | string | "ZinxServerApp" | 服务器名称 |
| Host | string | "0.0.0.0" | 监听 IP |
| TCPPort | int | 8999 | TCP 端口 |
| WsPort | int | 9000 | WebSocket 端口 |
| MaxConn | int | 12000 | 最大连接数 |
| WorkerPoolSize | uint32 | 10 | Worker 池大小 |
| MaxPacketSize | uint32 | 4096 | 最大包大小 |

---

## 常见问题 (FAQ)

### Q1: 如何处理粘包/拆包？

Zinx 默认使用 TLV (Type-Length-Value) 格式：
- 前 4 字节：消息 ID
- 中 4 字节：数据长度
- 后 N 字节：消息数据

可通过 `SetPacket()` 自定义封包器。

### Q2: Worker 模式如何选择？

- **Hash** (默认): 轮询分配，适合通用场景
- **Bind**: 每连接绑定固定 Worker，适合长连接
- **DynamicBind**: 动态创建 Worker，平衡资源利用

### Q3: 如何实现连接认证？

使用 `SetOnConnStart` 钩子：
```go
server.SetOnConnStart(func(conn ziface.IConnection) {
    token := getConnectionToken(conn)
    if !validateToken(token) {
        conn.Stop()
    }
})
```

### Q4: WebSocket 如何鉴权？

```go
server.SetWebsocketAuth(func(r *http.Request) error {
    token := r.URL.Query().Get("token")
    if token != "valid-token" {
        return fmt.Errorf("invalid token")
    }
    return nil
})
```

### Q5: 如何优雅关闭服务器？

```go
signalChan := make(chan os.Signal, 1)
signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-signalChan
    server.Stop()
    os.Exit(0)
}()

server.Serve()
```

---

## 相关项目

### 官方衍生

| 项目 | 链接 | 说明 |
|------|------|------|
| moke-kit | https://github.com/GStones/moke-kit | 微服务框架 |
| tcptest | https://github.com/xxl6097/tcptest | TCP 调试工具 |

### 社区移植

| 项目 | 语言 | 链接 |
|------|------|------|
| zinx(C++) | C++ | https://github.com/marklion/zinx |
| zinx-lua | Lua | https://github.com/huqitt/zinx-lua |
| ginx | Java | https://github.com/ModuleCode/ginx |

---

## 版本历史

| 版本 | 日期 | 主要更新 |
|------|------|----------|
| v1.2.7 | 2025.06 | AsyncOpResult 修复、WebSocket 增强 |
| v1.2.5 | 2024.05 | 许可证改为 MIT、RequestPool |
| v1.2.0 | 2023.x | DynamicBind、RouterSlices |
| v1.0.0 | 2022.x | WebSocket 支持、完整版 |
| v0.10 | 2020.x | 连接属性设置 |

---

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 许可证

Zinx 使用 MIT 许可证，详见 [LICENSE](https://github.com/aceld/zinx/blob/master/LICENSE)

---

最后更新：2026-03-10
