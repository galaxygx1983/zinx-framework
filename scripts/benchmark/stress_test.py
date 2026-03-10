#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Zinx Server 压力测试脚本
功能：模拟多客户端并发连接，测试服务器性能

依赖：pip install asyncio aiosockets

用法:
    python stress_test.py -c 100 -d 60 -s 127.0.0.1:8999
"""

import asyncio
import aiosockets
import argparse
import time
import statistics
from dataclasses import dataclass, field
from typing import List, Optional
from datetime import datetime


@dataclass
class TestConfig:
    """测试配置"""
    host: str = "127.0.0.1"
    port: int = 8999
    connections: int = 100
    duration: int = 60
    message_size: int = 100
    send_interval: float = 0.1


@dataclass
class ClientStats:
    """单个客户端统计"""
    client_id: int
    connected_at: float = 0
    disconnected_at: float = 0
    messages_sent: int = 0
    messages_received: int = 0
    errors: int = 0
    latencies: List[float] = field(default_factory=list)

    @property
    def connection_duration(self) -> float:
        return self.disconnected_at - self.connected_at if self.disconnected_at else 0

    @property
    def avg_latency(self) -> float:
        return statistics.mean(self.latencies) if self.latencies else 0

    @property
    def p99_latency(self) -> float:
        if not self.latencies:
            return 0
        sorted_latencies = sorted(self.latencies)
        idx = int(len(sorted_latencies) * 0.99)
        return sorted_latencies[min(idx, len(sorted_latencies) - 1)]


@dataclass
class TestResult:
    """测试结果汇总"""
    total_connections: int = 0
    successful_connections: int = 0
    failed_connections: int = 0
    total_messages_sent: int = 0
    total_messages_received: int = 0
    total_errors: int = 0
    test_duration: float = 0
    client_stats: List[ClientStats] = field(default_factory=list)

    @property
    def success_rate(self) -> float:
        return (self.successful_connections / self.total_connections * 100) if self.total_connections > 0 else 0

    @property
    def messages_per_second(self) -> float:
        return self.total_messages_sent / self.test_duration if self.test_duration > 0 else 0

    @property
    def avg_latency(self) -> float:
        all_latencies = [lat for stats in self.client_stats for lat in stats.latencies]
        return statistics.mean(all_latencies) if all_latencies else 0

    @property
    def p99_latency(self) -> float:
        all_latencies = [lat for stats in self.client_stats for lat in stats.latencies]
        if not all_latencies:
            return 0
        sorted_latencies = sorted(all_latencies)
        idx = int(len(sorted_latencies) * 0.99)
        return sorted_latencies[min(idx, len(sorted_latencies) - 1)]


class ZinxClient:
    """Zinx TCP 客户端"""

    def __init__(self, client_id: int, config: TestConfig, stats: ClientStats, results: TestResult):
        self.client_id = client_id
        self.config = config
        self.stats = stats
        self.results = results
        self.socket: Optional[aiosockets.Socket] = None
        self.running = False

    async def connect(self) -> bool:
        """连接到服务器"""
        try:
            self.socket = await aiosockets.connect(self.config.host, self.config.port)
            self.stats.connected_at = time.time()
            self.results.successful_connections += 1
            print(f"[Client {self.client_id}] Connected")
            return True
        except Exception as e:
            self.stats.errors += 1
            self.results.failed_connections += 1
            print(f"[Client {self.client_id}] Connection failed: {e}")
            return False

    async def send_message(self, msg_id: int, data: bytes) -> bool:
        """发送消息 (TLV 格式：MsgId(4) + Len(4) + Data)"""
        try:
            if not self.socket:
                return False

            # TLV 格式封包
            import struct
            length = len(data)
            packet = struct.pack('<II', msg_id, length) + data

            await self.socket.send(packet)
            self.stats.messages_sent += 1
            return True
        except Exception as e:
            self.stats.errors += 1
            self.results.total_errors += 1
            print(f"[Client {self.client_id}] Send error: {e}")
            return False

    async def receive_message(self) -> Optional[bytes]:
        """接收消息"""
        try:
            if not self.socket:
                return None

            import struct
            # 读取消息头 (8 字节)
            header = await self.socket.recv(8)
            if len(header) < 8:
                return None

            msg_id, length = struct.unpack('<II', header)

            # 读取消息体
            if length > 0:
                data = await self.socket.recv(length)
                return data
            return b''
        except Exception as e:
            self.stats.errors += 1
            return None

    async def run(self):
        """运行客户端测试"""
        self.running = True
        message = b'X' * self.config.message_size

        try:
            while self.running and self.socket:
                start_time = time.time()

                # 发送消息
                await self.send_message(1, message)

                # 接收响应
                response = await self.receive_message()
                if response:
                    latency = time.time() - start_time
                    self.stats.latencies.append(latency)
                    self.stats.messages_received += 1
                    self.results.total_messages_received += 1

                # 等待下次发送
                await asyncio.sleep(self.config.send_interval)

        except asyncio.CancelledError:
            pass
        except Exception as e:
            self.stats.errors += 1
            print(f"[Client {self.client_id}] Error: {e}")
        finally:
            self.stats.disconnected_at = time.time()
            self.results.total_messages_sent += self.stats.messages_sent
            if self.socket:
                await self.socket.close()
            print(f"[Client {self.client_id}] Disconnected")

    async def close(self):
        """关闭连接"""
        self.running = False
        if self.socket:
            await self.socket.close()


async def run_stress_test(config: TestConfig) -> TestResult:
    """运行压力测试"""
    print(f"""
╔═══════════════════════════════════════════════════════════╗
║              Zinx Server Stress Test                      ║
╠═══════════════════════════════════════════════════════════╣
║  Target:     {config.host}:{config.port}
║  Connections: {config.connections}
║  Duration:   {config.duration} seconds
║  Msg Size:   {config.message_size} bytes
║  Interval:   {config.send_interval} seconds
╚═══════════════════════════════════════════════════════════╝
""")

    results = TestResult()
    results.total_connections = config.connections
    clients: List[ZinxClient] = []
    tasks: List[asyncio.Task] = []

    start_time = time.time()

    # 创建客户端
    for i in range(config.connections):
        stats = ClientStats(client_id=i)
        results.client_stats.append(stats)
        client = ZinxClient(i, config, stats, results)
        clients.append(client)

    # 并发连接
    print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Connecting {config.connections} clients...")
    connection_tasks = [client.connect() for client in clients]
    await asyncio.gather(*connection_tasks, return_exceptions=True)

    print(f"[{datetime.now().strftime('%H:%M:%S')}] All clients connected, starting traffic...")

    # 启动客户端
    for client in clients:
        task = asyncio.create_task(client.run())
        tasks.append(task)

    # 等待测试完成
    try:
        await asyncio.sleep(config.duration)
    except KeyboardInterrupt:
        print("\nTest interrupted by user")

    # 停止所有客户端
    print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Stopping test...")
    for client in clients:
        await client.close()

    for task in tasks:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass

    end_time = time.time()
    results.test_duration = end_time - start_time

    return results


def print_report(results: TestResult):
    """打印测试报告"""
    print(f"""

╔═══════════════════════════════════════════════════════════╗
║                    Test Report                            ║
╠═══════════════════════════════════════════════════════════╣
║  CONNECTIONS                                              ║
║  ├─ Total:      {results.total_connections:>6}                                    ║
║  ├─ Success:    {results.successful_connections:>6} ({results.success_rate:>5.1f%)                           ║
║  └─ Failed:     {results.failed_connections:>6}                                    ║
║                                                           ║
║  MESSAGES                                                 ║
║  ├─ Sent:       {results.total_messages_sent:>8}                                  ║
║  ├─ Received:   {results.total_messages_received:>8}                                  ║
║  └─ Throughput: {results.messages_per_second:>8.2f} msg/s                           ║
║                                                           ║
║  LATENCY                                                  ║
║  ├─ Average:    {results.avg_latency*1000:>8.2f} ms                                 ║
║  └─ P99:        {results.p99_latency*1000:>8.2f} ms                                 ║
║                                                           ║
║  ERRORS                                                   ║
║  └─ Total:      {results.total_errors:>6}                                    ║
║                                                           ║
║  DURATION:        {results.test_duration:>6.2f} seconds                             ║
╚═══════════════════════════════════════════════════════════╝
""")


def main():
    parser = argparse.ArgumentParser(description='Zinx Server Stress Test')
    parser.add_argument('-s', '--server', default='127.0.0.1:8999', help='Server address')
    parser.add_argument('-c', '--connections', type=int, default=100, help='Number of connections')
    parser.add_argument('-d', '--duration', type=int, default=60, help='Test duration in seconds')
    parser.add_argument('-m', '--message-size', type=int, default=100, help='Message size in bytes')
    parser.add_argument('-i', '--interval', type=float, default=0.1, help='Send interval in seconds')

    args = parser.parse_args()

    host, port = args.server.split(':')
    config = TestConfig(
        host=host,
        port=int(port),
        connections=args.connections,
        duration=args.duration,
        message_size=args.message_size,
        send_interval=args.interval
    )

    # 运行测试
    results = asyncio.run(run_stress_test(config))

    # 打印报告
    print_report(results)


if __name__ == '__main__':
    main()
