# system-rpc

独立运行的 system 领域 RPC 进程。启动、配置和测试见[领域 README](../README.md)，方法见 [RPC.md](RPC.md)。

生成入口是 [system.proto](system.proto)，`make gen-rpc DOMAIN=system` 生成 pb、server、client；config、svc 和 logic 为保留的手写骨架。
